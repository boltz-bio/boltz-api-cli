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

func TestAuthLoginSubprocessCreatesSession(t *testing.T) {
	binary := buildCLIBinary(t)
	env := authProcessEnv(t)
	env = installFakeBrowserOpener(t, env)
	require.NoError(t, authconfig.SaveBaseURL("https://api.customer.example.com"))
	provider := newSubprocessOIDCProvider(t)
	defer provider.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, binary,
		"auth", "login",
		"--auth-issuer-url", provider.IssuerURL(),
		"--auth-client-id", "client-123",
		"--auth-scope", "openid",
		"--auth-scope", "email",
		"--listen-port", "0",
	)
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

	resp, err := http.Get(authURL)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	require.NoError(t, cmd.Wait(), stderr.String())
	lines := <-linesCh
	require.Contains(t, strings.Join(lines, "\n"), "Authentication successful.")

	status := runCLI(t, binary, env, "--format", "json", "auth", "status")
	require.Equal(t, 0, status.ExitCode, status.Stderr)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(status.Stdout), &response))
	require.Equal(t, true, response["authenticated"])
	require.Equal(t, "oauth", response["effective_mode"])
	require.Equal(t, []any{"openid", "email"}, response["granted_scopes"])

	config, err := authconfig.Load()
	require.NoError(t, err)
	require.Equal(t, "https://api.customer.example.com", config.BaseURL)
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

	mu sync.Mutex
}

func newSubprocessOIDCProvider(t *testing.T) *subprocessOIDCProvider {
	t.Helper()

	provider := &subprocessOIDCProvider{}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                 provider.server.URL,
			"authorization_endpoint": provider.server.URL + "/authorize",
			"token_endpoint":         provider.server.URL + "/token",
			"userinfo_endpoint":      provider.server.URL + "/userinfo",
		})
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
