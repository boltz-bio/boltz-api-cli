package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/boltz-bio/boltz-api-cli/internal/authconfig"
	"github.com/boltz-bio/boltz-api-cli/internal/authstore"
	"github.com/stretchr/testify/require"
)

func TestAuthStatusSubprocessReturnsJSONAndExitOneWhenUnauthenticated(t *testing.T) {
	binary := buildCLIBinary(t)
	env := authProcessEnv(t)

	result := runCLI(t, binary, env, "--format", "json", "auth", "status")
	require.Equal(t, 1, result.ExitCode)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(result.Stdout), &response))
	require.Equal(t, false, response["authenticated"])
	require.Equal(t, "none", response["effective_mode"])
}

func TestAuthLoginSubprocessPersistsDiscoveredBaseURLAndRoutesLaterRequest(t *testing.T) {
	binary := buildCLIBinary(t)
	t.Setenv(authconfig.EnvBaseURL, "")
	t.Setenv(authconfig.EnvAuthIssuerURL, "")
	t.Setenv("BOLTZ_API_KEY", "")
	env := withoutEnvironment(authProcessEnv(t), authconfig.EnvBaseURL, authconfig.EnvAuthIssuerURL, "BOLTZ_API_KEY")
	env = installFakeBrowserOpener(t, env)

	type observedRequest struct {
		path          string
		authorization string
	}
	var requestMu sync.Mutex
	var requests []observedRequest
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMu.Lock()
		requests = append(requests, observedRequest{
			path:          r.URL.Path,
			authorization: r.Header.Get("Authorization"),
		})
		requestMu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"latest":"0.39.0","minimum_supported":"0.8.0","update_available":false}`))
	}))
	defer apiServer.Close()
	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{
		BaseURL:   "https://old-tenant-api.example.com",
		IssuerURL: "https://old-tenant-issuer.example.com",
		ClientID:  "old-client",
		Scopes:    []string{"openid"},
	}))

	provider := newSubprocessOIDCProvider(t)
	defer provider.Close()
	provider.SetComputeAPIBaseURL(apiServer.URL)

	login := runAuthLoginSubprocess(t, binary, env,
		"--auth-issuer-url", provider.IssuerURL(),
		"--auth-client-id", "client-123",
		"--auth-scope", "openid",
		"--auth-scope", "email",
		"--listen-port", "0",
		"auth", "login",
	)
	require.NoError(t, login.Err, login.Stderr)
	require.Contains(t, login.Stdout, "Authentication successful.")
	config, err := authconfig.Load()
	require.NoError(t, err)
	require.Equal(t, apiServer.URL, config.BaseURL)
	require.Equal(t, provider.IssuerURL(), config.IssuerURL)
	status := runCLI(t, binary, env, "--format", "json", "auth", "status")
	require.Equal(t, 0, status.ExitCode, "stdout:\n%s\nstderr:\n%s", status.Stdout, status.Stderr)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(status.Stdout), &response))
	require.Equal(t, true, response["authenticated"])
	require.Equal(t, "oauth", response["effective_mode"])
	require.Equal(t, []any{"openid", "email"}, response["granted_scopes"])

	requestMu.Lock()
	requests = nil
	requestMu.Unlock()

	version := runCLI(t, binary, env, "--format", "json", "cli", "version")
	require.Equal(t, 0, version.ExitCode, version.Stderr)

	requestMu.Lock()
	require.Len(t, requests, 1)
	require.Equal(t, "/compute/v1/cli/version", requests[0].path)
	require.Equal(t, "Bearer login-access", requests[0].authorization)
	requestMu.Unlock()
}

func TestAuthLoginSubprocessOAuthFailurePreservesConfigBytes(t *testing.T) {
	binary := buildCLIBinary(t)
	t.Setenv(authconfig.EnvBaseURL, "")
	t.Setenv(authconfig.EnvAuthIssuerURL, "")
	env := installFakeBrowserOpener(t, withoutEnvironment(authProcessEnv(t), authconfig.EnvBaseURL, authconfig.EnvAuthIssuerURL))

	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{
		BaseURL:   "https://old-api.example.com",
		IssuerURL: "https://old-issuer.example.com",
		ClientID:  "old-client",
		Scopes:    []string{"openid"},
	}))
	configPath, err := authstore.ConfigFilePath()
	require.NoError(t, err)
	before, err := os.ReadFile(configPath)
	require.NoError(t, err)
	previousSession := authstore.Session{
		IssuerURL:   "https://old-issuer.example.com",
		ClientID:    "old-client",
		Scopes:      []string{"openid"},
		AccessToken: "old-access",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(time.Hour),
		TokenURL:    "https://old-issuer.example.com/token",
	}
	require.NoError(t, authstore.SaveSession(previousSession))
	_, err = authstore.SaveRefreshToken("old-refresh")
	require.NoError(t, err)

	provider := newSubprocessOIDCProviderWithTokenStatus(t, http.StatusBadRequest)
	defer provider.Close()

	login := runAuthLoginSubprocess(t, binary, env,
		"--base-url", "https://new-api.example.com",
		"--auth-issuer-url", provider.IssuerURL(),
		"--auth-client-id", "client-123",
		"--auth-scope", "openid",
		"--listen-port", "0",
		"auth", "login",
	)
	require.Error(t, login.Err)

	after, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.Equal(t, before, after)
	currentSession, err := authstore.LoadSession()
	require.NoError(t, err)
	require.NotNil(t, currentSession)
	require.Equal(t, previousSession.AccessToken, currentSession.AccessToken)
	currentToken, _, err := authstore.LoadRefreshToken()
	require.NoError(t, err)
	require.Equal(t, "old-refresh", currentToken)
}

func TestAuthLoginSubprocessWithoutBaseOverridePreservesStoredBaseURL(t *testing.T) {
	binary := buildCLIBinary(t)
	t.Setenv(authconfig.EnvBaseURL, "")
	t.Setenv(authconfig.EnvAuthIssuerURL, "")
	env := installFakeBrowserOpener(t, withoutEnvironment(authProcessEnv(t), authconfig.EnvBaseURL, authconfig.EnvAuthIssuerURL))

	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{
		BaseURL:   "https://stored-api.example.com",
		IssuerURL: "https://old-issuer.example.com",
		ClientID:  "old-client",
		Scopes:    []string{"openid"},
	}))
	provider := newSubprocessOIDCProvider(t)
	defer provider.Close()

	login := runAuthLoginSubprocess(t, binary, env,
		"--auth-issuer-url", provider.IssuerURL(),
		"--auth-client-id", "client-123",
		"--auth-scope", "openid",
		"--listen-port", "0",
		"auth", "login",
	)
	require.NoError(t, login.Err, login.Stderr)

	config, err := authconfig.Load()
	require.NoError(t, err)
	require.Equal(t, "https://stored-api.example.com", config.BaseURL)
	require.Equal(t, provider.IssuerURL(), config.IssuerURL)
}

func TestAuthLoginSubprocessInvalidBaseURLFailsBeforeOAuth(t *testing.T) {
	binary := buildCLIBinary(t)
	env := authProcessEnv(t)
	provider := newSubprocessOIDCProvider(t)
	defer provider.Close()

	result := runCLI(t, binary, env,
		"--base-url", "api.customer.example.com",
		"--auth-issuer-url", provider.IssuerURL(),
		"auth", "login",
	)
	require.NotEqual(t, 0, result.ExitCode)
	require.Contains(t, result.Stderr, "missing a scheme")
	require.Equal(t, 0, provider.DiscoveryRequests())
}

func TestAuthValidateSubprocessClearsInvalidGrantState(t *testing.T) {
	binary := buildCLIBinary(t)
	env := authProcessEnv(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"expired refresh token"}`))
	}))
	defer server.Close()

	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{
		IssuerURL: "https://issuer.example.com",
		ClientID:  "client-123",
		Audience:  authconfig.DefaultAudience,
		Scopes:    []string{"openid", "email"},
	}))
	require.NoError(t, authstore.SaveSession(authstore.Session{
		IssuerURL:   "https://issuer.example.com",
		ClientID:    "client-123",
		Audience:    authconfig.DefaultAudience,
		Scopes:      []string{"openid", "email"},
		AccessToken: "expired-access",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(-1 * time.Hour),
		TokenURL:    server.URL,
		JWKSURL:     "https://issuer.example.com/jwks",
		Algorithms:  []string{"RS256"},
	}))
	_, err := authstore.SaveRefreshToken("refresh-token")
	require.NoError(t, err)

	result := runCLI(t, binary, env, "--format", "json", "auth", "validate")
	require.Equal(t, 1, result.ExitCode)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(result.Stdout), &response))
	require.Equal(t, false, response["valid"])

	session, err := authstore.LoadSession()
	require.NoError(t, err)
	require.Nil(t, session)

	refreshToken, _, err := authstore.LoadRefreshToken()
	require.NoError(t, err)
	require.Empty(t, refreshToken)
}

type cliResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

type authLoginResult struct {
	Stdout string
	Stderr string
	Err    error
}

func runAuthLoginSubprocess(t *testing.T, binary string, env []string, args ...string) authLoginResult {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary, args...)
	cmd.Env = env
	cmd.Dir = repoRoot(t)

	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	require.NoError(t, cmd.Start())

	linesCh := make(chan []string, 1)
	urlCh := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		lines := make([]string, 0, 8)
		expectURL := false
		for scanner.Scan() {
			line := scanner.Text()
			lines = append(lines, line)
			if expectURL && strings.HasPrefix(line, "http") {
				urlCh <- line
				expectURL = false
				continue
			}
			if strings.Contains(line, "Open this URL to authenticate:") {
				expectURL = true
			}
		}
		linesCh <- lines
	}()

	var authURL string
	select {
	case authURL = <-urlCh:
	case <-ctx.Done():
		t.Fatal("timed out waiting for auth URL")
	}

	response, err := http.Get(authURL)
	require.NoError(t, err)
	_ = response.Body.Close()

	waitErr := cmd.Wait()
	lines := <-linesCh
	return authLoginResult{
		Stdout: strings.Join(lines, "\n"),
		Stderr: stderr.String(),
		Err:    waitErr,
	}
}

func runCLI(t *testing.T, binary string, env []string, args ...string) cliResult {
	t.Helper()

	cmd := exec.Command(binary, args...)
	cmd.Env = env
	cmd.Dir = repoRoot(t)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return cliResult{
			Stdout:   stdout.String(),
			Stderr:   stderr.String(),
			ExitCode: 0,
		}
	}

	var exitErr *exec.ExitError
	require.True(t, errors.As(err, &exitErr), "unexpected command error: %v\nstderr:\n%s", err, stderr.String())
	return cliResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitErr.ExitCode(),
	}
}

func buildCLIBinary(t *testing.T) string {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "boltz-api")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", binary, "./cmd/boltz-api")
	cmd.Dir = repoRoot(t)
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, string(output))
	return binary
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func authProcessEnv(t *testing.T) []string {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
	t.Setenv("BOLTZ_API_TEST_DISABLE_KEYRING", "1")
	authstore.SetKeyringBackend(subprocessTestKeyringBackend{
		get: func(service, key string) (string, error) {
			return "", errors.New("keyring unavailable")
		},
		set: func(service, key, value string) error {
			return errors.New("keyring unavailable")
		},
		delete: func(service, key string) error {
			return errors.New("keyring unavailable")
		},
	})
	t.Cleanup(authstore.ResetKeyring)

	env := append([]string{}, os.Environ()...)
	env = append(env,
		"HOME="+home,
		"XDG_CONFIG_HOME="+home,
		"XDG_CACHE_HOME="+home,
		"BOLTZ_API_TEST_DISABLE_KEYRING=1",
	)
	return env
}

func withoutEnvironment(env []string, names ...string) []string {
	blocked := make(map[string]struct{}, len(names))
	for _, name := range names {
		blocked[name] = struct{}{}
	}
	result := make([]string, 0, len(env))
	for _, entry := range env {
		name, _, found := strings.Cut(entry, "=")
		if found {
			if _, remove := blocked[name]; remove {
				continue
			}
		}
		result = append(result, entry)
	}
	return result
}

func installFakeBrowserOpener(t *testing.T, env []string) []string {
	t.Helper()

	dir := t.TempDir()
	name := "xdg-open"
	body := "#!/bin/sh\nexit 0\n"
	if runtime.GOOS == "darwin" {
		name = "open"
	}
	if runtime.GOOS == "windows" {
		name = "rundll32.bat"
		body = "@echo off\r\nexit /b 0\r\n"
	}

	require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o755))
	path := dir + string(os.PathListSeparator) + os.Getenv("PATH")
	result := make([]string, 0, len(env)+1)
	replaced := false
	for _, entry := range env {
		if strings.HasPrefix(entry, "PATH=") {
			result = append(result, "PATH="+path)
			replaced = true
			continue
		}
		result = append(result, entry)
	}
	if !replaced {
		result = append(result, "PATH="+path)
	}
	return result
}

type subprocessTestKeyringBackend struct {
	get    func(service, key string) (string, error)
	set    func(service, key, value string) error
	delete func(service, key string) error
}

func (m subprocessTestKeyringBackend) Get(service, key string) (string, error) {
	if m.get != nil {
		return m.get(service, key)
	}
	return "", nil
}

func (m subprocessTestKeyringBackend) Set(service, key, value string) error {
	if m.set != nil {
		return m.set(service, key, value)
	}
	return nil
}

func (m subprocessTestKeyringBackend) Delete(service, key string) error {
	if m.delete != nil {
		return m.delete(service, key)
	}
	return nil
}

type subprocessOIDCProvider struct {
	server *httptest.Server

	mu                sync.Mutex
	discoveryRequests int
	tokenStatus       int
	computeAPIBaseURL string
}

func newSubprocessOIDCProvider(t *testing.T) *subprocessOIDCProvider {
	return newSubprocessOIDCProviderWithTokenStatus(t, http.StatusOK)
}

func newSubprocessOIDCProviderWithTokenStatus(t *testing.T, tokenStatus int) *subprocessOIDCProvider {
	t.Helper()

	provider := &subprocessOIDCProvider{tokenStatus: tokenStatus}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		provider.mu.Lock()
		provider.discoveryRequests++
		computeAPIBaseURL := provider.computeAPIBaseURL
		provider.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		document := map[string]any{
			"issuer":                 provider.server.URL,
			"authorization_endpoint": provider.server.URL + "/authorize",
			"token_endpoint":         provider.server.URL + "/token",
			"userinfo_endpoint":      provider.server.URL + "/userinfo",
		}
		if computeAPIBaseURL != "" {
			document["boltz_compute_api_base_url"] = computeAPIBaseURL
		}
		_ = json.NewEncoder(w).Encode(document)
	})
	mux.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
		callback, err := url.Parse(r.URL.Query().Get("redirect_uri"))
		require.NoError(t, err)
		query := callback.Query()
		query.Set("code", "login-code")
		query.Set("state", r.URL.Query().Get("state"))
		callback.RawQuery = query.Encode()
		http.Redirect(w, r, callback.String(), http.StatusFound)
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		require.Equal(t, "authorization_code", r.PostForm.Get("grant_type"))
		require.Equal(t, "client-123", r.PostForm.Get("client_id"))
		require.Equal(t, "login-code", r.PostForm.Get("code"))
		require.NotEmpty(t, r.PostForm.Get("redirect_uri"))
		require.NotEmpty(t, r.PostForm.Get("code_verifier"))
		w.Header().Set("Content-Type", "application/json")
		if provider.tokenStatus != http.StatusOK {
			w.WriteHeader(provider.tokenStatus)
			_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"rejected test login"}`))
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"login-access","refresh_token":"login-refresh","token_type":"Bearer","expires_in":3600}`))
	})
	mux.HandleFunc("/userinfo", func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "Bearer login-access", r.Header.Get("Authorization"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sub":"user-123","email":"user@example.com","name":"User Example","preferred_username":"user-example"}`))
	})

	provider.server = httptest.NewServer(mux)
	return provider
}

func (p *subprocessOIDCProvider) Close() {
	p.server.Close()
}

func (p *subprocessOIDCProvider) IssuerURL() string {
	return p.server.URL
}

func (p *subprocessOIDCProvider) SetComputeAPIBaseURL(baseURL string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.computeAPIBaseURL = baseURL
}

func (p *subprocessOIDCProvider) DiscoveryRequests() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.discoveryRequests
}
