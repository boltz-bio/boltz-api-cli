// Custom CLI extension code. Not generated.
package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/boltz-bio/boltz-api-cli/internal/authconfig"
	"github.com/boltz-bio/boltz-api-cli/internal/autherror"
	"github.com/boltz-bio/boltz-api-cli/internal/authstore"
	"github.com/boltz-bio/boltz-api-cli/internal/oauthclient"
	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
	"github.com/zalando/go-keyring"
)

func TestAuthWhoAmIReportsOperatorFriendlyIdentityAndSources(t *testing.T) {
	setAuthCommandUserDirs(t)
	useFileOnlyKeyringForAuthCommandTests(t)

	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{
		IssuerURL:   "https://issuer.example.com",
		ClientID:    "client-123",
		Scopes:      []string{"openid", "profile"},
		Audience:    "aud-123",
		SelectedOrg: "org-config",
	}))

	backend, err := authstore.SaveRefreshToken("refresh-token")
	require.NoError(t, err)
	require.Equal(t, "file", backend)

	require.NoError(t, authstore.SaveSession(authstore.Session{
		IssuerURL:      "https://issuer.example.com",
		ClientID:       "client-123",
		Audience:       "aud-123",
		Scopes:         []string{"openid", "profile"},
		GrantedScopes:  []string{"openid", "profile"},
		AccessToken:    "access-token",
		TokenType:      "Bearer",
		Expiry:         time.Date(2026, 4, 21, 13, 0, 0, 0, time.UTC),
		StorageBackend: backend,
		Identity: authstore.Identity{
			Subject:           "user-123",
			Email:             "user@example.com",
			Name:              "Test User",
			PreferredUsername: "test-user",
			Claims:            map[string]any{"role": "admin"},
		},
	}))

	output, err := runAuthCommand(t, "--format", "json", "auth", "whoami")
	require.NoError(t, err)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &response))

	identity, ok := response["identity"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "user@example.com", identity["email"])
	require.Equal(t, "Test User", identity["name"])
	require.NotContains(t, identity, "claims")

	sources, ok := response["sources"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, string(authconfig.SourceConfig), sources["issuer_url"])
	require.Equal(t, string(authconfig.SourceConfig), sources["client_id"])
	require.Equal(t, string(authconfig.SourceConfig), sources["selected_org"])
	require.Equal(t, string(authconfig.SourceSessionCache), sources["session"])
	require.Equal(t, string(authconfig.SourceFile), sources["refresh_token"])
}

func TestAuthStatusReportsAPIKeyOverride(t *testing.T) {
	setAuthCommandUserDirs(t)
	useFileOnlyKeyringForAuthCommandTests(t)
	t.Setenv("BOLTZ_API_KEY", "api-key-123")

	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{
		IssuerURL:   "https://issuer.example.com",
		ClientID:    "client-123",
		Scopes:      []string{"openid", "profile"},
		SelectedOrg: "org-config",
	}))
	require.NoError(t, authstore.SaveSession(authstore.Session{
		IssuerURL:     "https://issuer.example.com",
		ClientID:      "client-123",
		Scopes:        []string{"openid", "profile"},
		GrantedScopes: []string{"openid"},
		AccessToken:   "oauth-access",
		TokenType:     "Bearer",
		Expiry:        time.Now().Add(10 * time.Minute),
		Identity: authstore.Identity{
			Email: "user@example.com",
		},
	}))
	_, err := authstore.SaveRefreshToken("refresh-token")
	require.NoError(t, err)

	output, err := runAuthCommand(t, "--format", "json", "auth", "status")
	require.NoError(t, err)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &response))
	require.Equal(t, true, response["authenticated"])
	require.Equal(t, "api_key", response["effective_mode"])
	require.Equal(t, "api_key", response["active_source"])
	require.Equal(t, true, response["api_key_overrides_oauth"])
	_, hasMissingScopes := response["missing_scopes"]
	require.False(t, hasMissingScopes)
	_, hasIdentity := response["identity"]
	require.False(t, hasIdentity)

	storedOAuthSession, ok := response["stored_oauth_session"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, []any{"profile"}, storedOAuthSession["missing_scopes"])

	warnings, ok := response["warnings"].([]any)
	require.True(t, ok)
	require.Contains(t, warnings, "API-key mode is overriding the stored OAuth session.")
}

func TestAuthStatusSessionMismatchKeepsActiveSourceUnset(t *testing.T) {
	setAuthCommandUserDirs(t)
	useFileOnlyKeyringForAuthCommandTests(t)

	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{
		IssuerURL: "https://issuer.example.com",
		ClientID:  "client-456",
		Scopes:    []string{"openid", "profile"},
	}))
	require.NoError(t, authstore.SaveSession(authstore.Session{
		IssuerURL:     "https://issuer.example.com",
		ClientID:      "client-123",
		Scopes:        []string{"openid", "profile"},
		GrantedScopes: []string{"openid", "profile"},
		AccessToken:   "oauth-access",
		TokenType:     "Bearer",
		Expiry:        time.Now().Add(10 * time.Minute),
	}))

	root := newAuthTestRoot(os.Stdout)
	snapshot, err := loadAuthSnapshot(root)
	require.NoError(t, err)

	response, authenticated := buildAuthStatusResponse(snapshot)
	require.False(t, authenticated)
	require.Equal(t, "none", response.EffectiveMode)
	require.Equal(t, "none", response.ActiveSource)
	require.NotNil(t, response.StoredOAuthSession)
	require.Equal(t, []string{"openid", "profile"}, response.StoredOAuthSession.GrantedScopes)
}

func TestAuthStatusReportsConfigPathAndPlaceholderWarnings(t *testing.T) {
	setAuthCommandUserDirs(t)
	useFileOnlyKeyringForAuthCommandTests(t)

	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{
		IssuerURL: "https://issuer.example.com",
		ClientID:  "client-123",
		Scopes:    []string{"openid"},
	}))

	output, err := runAuthCommand(t, "--format", "json", "auth", "status")
	require.Error(t, err)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &response))
	config, ok := response["config"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, config["present"])
	require.Contains(t, config["path"], "config.yaml")

	warnings, ok := response["warnings"].([]any)
	require.True(t, ok)
	require.Contains(t, warnings, `OAuth issuer URL looks like a placeholder: "https://issuer.example.com".`)
	require.Contains(t, warnings, `OAuth client ID looks like a placeholder: "client-123".`)

	actions, ok := response["actions"].([]any)
	require.True(t, ok)
	require.Contains(t, actions[0], "Inspect ")
}

func TestAuthValidateRefreshesExpiredSession(t *testing.T) {
	setAuthCommandUserDirs(t)
	useFileOnlyKeyringForAuthCommandTests(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		require.Equal(t, "refresh_token", r.PostForm.Get("grant_type"))
		require.Equal(t, "refresh-token", r.PostForm.Get("refresh_token"))
		require.Equal(t, authconfig.DefaultAudience, r.PostForm.Get("resource"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-access","refresh_token":"refresh-token-2","token_type":"Bearer","expires_in":3600,"scope":"openid profile"}`))
	}))
	defer server.Close()

	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{
		IssuerURL: "https://issuer.example.com",
		ClientID:  "client-123",
		Audience:  authconfig.DefaultAudience,
		Scopes:    []string{"openid", "profile"},
	}))
	require.NoError(t, authstore.SaveSession(authstore.Session{
		IssuerURL:     "https://issuer.example.com",
		ClientID:      "client-123",
		Audience:      authconfig.DefaultAudience,
		Scopes:        []string{"openid", "profile"},
		GrantedScopes: []string{"openid"},
		AccessToken:   "old-access",
		TokenType:     "Bearer",
		Expiry:        time.Now().Add(-1 * time.Minute),
		TokenURL:      server.URL,
		JWKSURL:       "https://issuer.example.com/jwks",
		Algorithms:    []string{"RS256"},
	}))
	_, err := authstore.SaveRefreshToken("refresh-token")
	require.NoError(t, err)

	output, err := runAuthCommand(t, "--format", "json", "auth", "validate")
	require.NoError(t, err)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &response))
	require.Equal(t, true, response["valid"])
	require.Equal(t, true, response["refreshed"])
	require.Equal(t, "oauth", response["effective_mode"])

	session, err := authstore.LoadSession()
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, "new-access", session.AccessToken)
	require.Equal(t, []string{"openid", "profile"}, session.GrantedScopes)
}

func TestAuthValidateAPIKeyExplainsLocalOnlyCheck(t *testing.T) {
	setAuthCommandUserDirs(t)
	useFileOnlyKeyringForAuthCommandTests(t)
	t.Setenv("BOLTZ_API_KEY", "api-key-123")

	output, err := runAuthCommand(t, "--format", "json", "auth", "validate")
	require.NoError(t, err)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &response))
	require.Equal(t, true, response["valid"])
	require.Equal(t, "api_key", response["effective_mode"])

	warnings, ok := response["warnings"].([]any)
	require.True(t, ok)
	require.Contains(t, warnings, "API-key validation is local only; this command confirms that an API key is configured.")
}

func TestAuthOrgsListsOAuthOrganizations(t *testing.T) {
	setAuthCommandUserDirs(t)
	useFileOnlyKeyringForAuthCommandTests(t)

	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{
		IssuerURL:   "https://issuer.example.com",
		ClientID:    "client-123",
		Audience:    authconfig.DefaultAudience,
		Scopes:      []string{"openid", "profile"},
		SelectedOrg: "org-second",
	}))
	require.NoError(t, authstore.SaveSession(authstore.Session{
		IssuerURL:   "https://issuer.example.com",
		ClientID:    "client-123",
		Audience:    authconfig.DefaultAudience,
		Scopes:      []string{"openid", "profile"},
		AccessToken: "oauth-access",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(10 * time.Minute),
	}))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/compute/v1/auth/me", r.URL.Path)
		require.Equal(t, "Bearer oauth-access", r.Header.Get("Authorization"))
		require.Equal(t, "org-second", r.Header.Get("X-Boltz-Organization-Id"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"principal_type": "user",
			"user_id": "user-123",
			"selected_organization_id": "org-second",
			"active_organization_id": null,
			"organization_memberships": [
				{"organization_id": "org-first", "organization_name": "Acme Bio", "role": "Admin"},
				{"organization_id": "org-second", "organization_name": null, "role": "Scientist"}
			]
		}`))
	}))
	defer server.Close()

	output, err := runAuthCommand(t, "--format", "json", "--base-url", server.URL, "auth", "orgs")
	require.NoError(t, err)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &response))
	require.Equal(t, "user", response["principal_type"])
	require.Equal(t, "org-second", response["selected_org"])

	organizations, ok := response["organizations"].([]any)
	require.True(t, ok)
	require.Len(t, organizations, 2)
	require.Equal(t, map[string]any{
		"organization_id":   "org-first",
		"organization_name": "Acme Bio",
		"role":              "Admin",
		"selected":          false,
		"switchable":        true,
	}, organizations[0])
	require.Equal(t, map[string]any{
		"organization_id": "org-second",
		"role":            "Scientist",
		"selected":        true,
		"switchable":      true,
	}, organizations[1])
}

func TestAuthOrgsShowsAPIKeyScope(t *testing.T) {
	setAuthCommandUserDirs(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/compute/v1/auth/me", r.URL.Path)
		require.Equal(t, "api-key-123", r.Header.Get("x-api-key"))
		require.Empty(t, r.Header.Get("X-Boltz-Organization-Id"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"principal_type": "api_key",
			"api_key_id": "key-123",
			"key_type": "workspace",
			"mode": "live",
			"organization_id": "org-api",
			"organization_name": "Acme Bio",
			"selected_organization_id": "org-api",
			"workspace_id": "ws-api"
		}`))
	}))
	defer server.Close()

	output, err := runAuthCommand(t, "--format", "json", "--base-url", server.URL, "--api-key", "api-key-123", "auth", "orgs")
	require.NoError(t, err)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &response))
	require.Equal(t, "api_key", response["principal_type"])
	require.Equal(t, "org-api", response["selected_org"])

	organizations, ok := response["organizations"].([]any)
	require.True(t, ok)
	require.Len(t, organizations, 1)
	require.Equal(t, map[string]any{
		"organization_id":   "org-api",
		"organization_name": "Acme Bio",
		"selected":          true,
		"switchable":        false,
	}, organizations[0])

	apiKey, ok := response["api_key"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "key-123", apiKey["api_key_id"])
	require.Equal(t, "workspace", apiKey["key_type"])
	require.Equal(t, "live", apiKey["mode"])
	require.Equal(t, "ws-api", apiKey["workspace_id"])
}

func TestAuthOrgsOmitsOrganizationNameWhenNullForAPIKey(t *testing.T) {
	setAuthCommandUserDirs(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/compute/v1/auth/me", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"principal_type": "api_key",
			"api_key_id": "key-456",
			"key_type": "workspace",
			"mode": "live",
			"organization_id": "org-unnamed",
			"organization_name": null,
			"selected_organization_id": "org-unnamed",
			"workspace_id": "ws-unnamed"
		}`))
	}))
	defer server.Close()

	output, err := runAuthCommand(t, "--format", "json", "--base-url", server.URL, "--api-key", "api-key-456", "auth", "orgs")
	require.NoError(t, err)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &response))

	organizations, ok := response["organizations"].([]any)
	require.True(t, ok)
	require.Len(t, organizations, 1)
	require.Equal(t, map[string]any{
		"organization_id": "org-unnamed",
		"selected":        true,
		"switchable":      false,
	}, organizations[0])
}

func TestAuthLoginDoesNotPersistProfileOnFailure(t *testing.T) {
	setAuthCommandUserDirs(t)
	useFileOnlyKeyringForAuthCommandTests(t)

	_, err := runAuthCommand(t,
		"auth", "login",
		"--auth-issuer-url", "http://127.0.0.1:1",
		"--auth-client-id", "client-123",
	)
	require.Error(t, err)

	config, loadErr := authconfig.Load()
	require.NoError(t, loadErr)
	require.Empty(t, config.IssuerURL)
	require.Empty(t, config.ClientID)
}

func TestAuthLoginConfigPersistenceFailureRollsBackSessionAndToken(t *testing.T) {
	setAuthCommandUserDirs(t)
	useFileOnlyKeyringForAuthCommandTests(t)

	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{
		BaseURL:   "https://old-api.example.com",
		IssuerURL: "https://old-issuer.example.com",
		ClientID:  "old-client",
		Scopes:    []string{"openid"},
	}))
	configPath, err := authstore.ConfigFilePath()
	require.NoError(t, err)
	configBefore, err := os.ReadFile(configPath)
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

	previousBrowserLogin := authBrowserLogin
	previousSaveProfile := authSaveProfile
	t.Cleanup(func() {
		authBrowserLogin = previousBrowserLogin
		authSaveProfile = previousSaveProfile
	})
	authBrowserLogin = func(context.Context, oauthclient.Config) (*oauthclient.LoginResult, error) {
		return &oauthclient.LoginResult{
			Provider: oauthclient.ProviderMetadata{
				IssuerURL: "https://new-issuer.example.com",
				TokenURL:  "https://new-issuer.example.com/token",
			},
			Tokens: oauthclient.TokenSet{
				AccessToken:   "new-access",
				RefreshToken:  "new-refresh",
				TokenType:     "Bearer",
				Expiry:        time.Now().Add(2 * time.Hour),
				GrantedScopes: []string{"openid"},
			},
		}, nil
	}
	var attemptedProfile authconfig.Resolved
	authSaveProfile = func(resolved authconfig.Resolved) error {
		attemptedProfile = resolved
		return errors.New("forced config persistence failure")
	}

	_, err = runAuthCommand(t,
		"--base-url", "https://new-api.example.com",
		"--auth-issuer-url", "https://new-issuer.example.com",
		"--auth-client-id", "new-client",
		"--auth-scope", "openid",
		"auth", "login",
	)
	var authErr *autherror.Error
	require.ErrorAs(t, err, &authErr)
	require.Equal(t, "config_save_failed", authErr.Envelope().Code)
	require.Equal(t, "https://new-api.example.com", attemptedProfile.BaseURL)

	configAfter, err := os.ReadFile(configPath)
	require.NoError(t, err)
	require.Equal(t, configBefore, configAfter)
	restoredSession, err := authstore.LoadSession()
	require.NoError(t, err)
	require.NotNil(t, restoredSession)
	require.Equal(t, previousSession.AccessToken, restoredSession.AccessToken)
	require.Equal(t, previousSession.IssuerURL, restoredSession.IssuerURL)
	require.Equal(t, previousSession.ClientID, restoredSession.ClientID)
	restoredToken, _, err := authstore.LoadRefreshToken()
	require.NoError(t, err)
	require.Equal(t, "old-refresh", restoredToken)
}

func TestAuthLoginJSONEventsRequiresDeviceCode(t *testing.T) {
	setAuthCommandUserDirs(t)
	useFileOnlyKeyringForAuthCommandTests(t)

	_, err := runAuthCommand(t, "auth", "login", "--json-events")
	require.Error(t, err)
	require.Contains(t, err.Error(), "--json-events")
	require.Contains(t, err.Error(), "--device-code")
}

func TestAuthWaitSucceedsOnceSessionAppears(t *testing.T) {
	setAuthCommandUserDirs(t)
	useFileOnlyKeyringForAuthCommandTests(t)

	scopes := append([]string(nil), authconfig.DefaultScopes...)
	scopes = append(scopes, authconfig.DefaultScope)

	sessionSaved := make(chan error, 1)
	go func() {
		time.Sleep(50 * time.Millisecond)
		sessionSaved <- authstore.SaveSession(authstore.Session{
			IssuerURL:     authconfig.DefaultIssuerURL,
			ClientID:      authconfig.DefaultClientID,
			Audience:      authconfig.DefaultAudience,
			Scopes:        scopes,
			GrantedScopes: scopes,
			AccessToken:   "access-token",
			TokenType:     "Bearer",
			Expiry:        time.Now().Add(10 * time.Minute),
		})
	}()

	output, err := runAuthCommand(t, "--format", "json", "auth", "wait", "--timeout", "500ms", "--poll-interval", "25ms")
	require.NoError(t, err)
	require.NoError(t, <-sessionSaved)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &response))
	require.Equal(t, "success", response["status"])
	require.Equal(t, true, response["authenticated"])
	require.Equal(t, "oauth", response["effective_mode"])
}

func TestAuthWaitReturnsWaitingStatusOnTimeout(t *testing.T) {
	setAuthCommandUserDirs(t)
	useFileOnlyKeyringForAuthCommandTests(t)

	output, err := runAuthCommand(t, "--format", "json", "auth", "wait", "--timeout", "75ms", "--poll-interval", "10ms")
	require.Error(t, err)

	var response map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &response))
	require.Equal(t, "waiting", response["status"])
	require.Equal(t, true, response["timed_out"])
	require.Equal(t, false, response["authenticated"])
}

func runAuthCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = r.Close()
	})

	root := newAuthTestRoot(w)
	runErr := root.Run(context.Background(), append([]string{"boltz-api"}, args...))
	require.NoError(t, w.Close())

	output, readErr := io.ReadAll(r)
	require.NoError(t, readErr)
	return string(output), runErr
}

func newAuthTestRoot(writer *os.File) *cli.Command {
	root := &cli.Command{
		Name:           "boltz-api",
		Writer:         writer,
		ErrWriter:      writer,
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "base-url"},
			&cli.StringFlag{Name: "format", Value: "auto"},
			&cli.BoolFlag{Name: "raw-output"},
			&cli.StringFlag{Name: "transform"},
			&requestflag.Flag[string]{
				Name:    "api-key",
				Sources: cli.EnvVars("BOLTZ_API_KEY"),
			},
		},
		Commands: []*cli.Command{authCommand},
	}
	root.Flags = append(root.Flags, authFlags()...)
	return root
}

func setAuthCommandUserDirs(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
}

func useFileOnlyKeyringForAuthCommandTests(t *testing.T) {
	t.Helper()

	authstore.SetKeyringBackend(authCommandTestKeyringBackend{
		get: func(service, key string) (string, error) {
			return "", errors.New("keyring unavailable")
		},
		set: func(service, key, value string) error {
			return errors.New("keyring unavailable")
		},
		delete: func(service, key string) error {
			return keyring.ErrNotFound
		},
	})
	t.Cleanup(authstore.ResetKeyring)
}

type authCommandTestKeyringBackend struct {
	get    func(service, key string) (string, error)
	set    func(service, key, value string) error
	delete func(service, key string) error
}

func (m authCommandTestKeyringBackend) Get(service, key string) (string, error) {
	if m.get != nil {
		return m.get(service, key)
	}
	return "", nil
}

func (m authCommandTestKeyringBackend) Set(service, key, value string) error {
	if m.set != nil {
		return m.set(service, key, value)
	}
	return nil
}

func (m authCommandTestKeyringBackend) Delete(service, key string) error {
	if m.delete != nil {
		return m.delete(service, key)
	}
	return nil
}
