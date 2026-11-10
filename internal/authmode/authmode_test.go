package authmode

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/boltz-bio/boltz-api-cli/internal/authconfig"
	"github.com/boltz-bio/boltz-api-cli/internal/autherror"
	"github.com/boltz-bio/boltz-api-cli/internal/authstore"
	"github.com/gofrs/flock"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func TestResolveRefreshesExpiringSession(t *testing.T) {
	setUserDirs(t)
	useFileOnlyKeyring(t)

	fixedNow := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	originalNow := now
	now = func() time.Time { return fixedNow }
	t.Cleanup(func() { now = originalNow })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		require.Equal(t, "refresh_token", r.PostForm.Get("grant_type"))
		require.Equal(t, "refresh-1", r.PostForm.Get("refresh_token"))
		require.Equal(t, "client-123", r.PostForm.Get("client_id"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-access","refresh_token":"refresh-2","token_type":"Bearer","expires_in":3600,"scope":"openid email"}`))
	}))
	defer server.Close()

	require.NoError(t, authstore.SaveSession(authstore.Session{
		IssuerURL:   "https://issuer.example.com",
		ClientID:    "client-123",
		Scopes:      []string{"openid", "profile"},
		AccessToken: "old-access",
		TokenType:   "Bearer",
		Expiry:      fixedNow.Add(30 * time.Second),
		TokenURL:    server.URL,
	}))
	writeRefreshTokenFile(t, "refresh-1")

	result, err := Resolve(context.Background(), authconfig.Resolved{
		IssuerURL: "https://issuer.example.com",
		ClientID:  "client-123",
		Scopes:    []string{"profile", "openid"},
	})
	require.NoError(t, err)
	require.Equal(t, ModeOAuth, result.Mode)
	require.Equal(t, "new-access", result.AccessToken)

	session, err := authstore.LoadSession()
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, "new-access", session.AccessToken)
	require.Equal(t, []string{"openid", "email"}, session.GrantedScopes)

	body, err := os.ReadFile(credentialsPath(t))
	require.NoError(t, err)
	require.JSONEq(t, `{"refresh_token":"refresh-2"}`, string(body))
}

func TestResolveClearsSessionOnInvalidGrant(t *testing.T) {
	setUserDirs(t)
	useFileOnlyKeyring(t)

	fixedNow := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	originalNow := now
	now = func() time.Time { return fixedNow }
	t.Cleanup(func() { now = originalNow })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant","error_description":"expired refresh token"}`))
	}))
	defer server.Close()

	require.NoError(t, authstore.SaveSession(authstore.Session{
		IssuerURL:   "https://issuer.example.com",
		ClientID:    "client-123",
		Scopes:      []string{"openid"},
		AccessToken: "old-access",
		TokenType:   "Bearer",
		Expiry:      fixedNow.Add(-1 * time.Minute),
		TokenURL:    server.URL,
	}))
	writeRefreshTokenFile(t, "refresh-1")

	_, err := Resolve(context.Background(), authconfig.Resolved{
		IssuerURL: "https://issuer.example.com",
		ClientID:  "client-123",
		Scopes:    []string{"openid"},
	})
	require.Error(t, err)

	var authErr *autherror.Error
	require.True(t, errors.As(err, &authErr))
	require.Equal(t, "reauth_required", authErr.Envelope().Code)

	session, loadErr := authstore.LoadSession()
	require.NoError(t, loadErr)
	require.Nil(t, session)

	_, statErr := os.Stat(credentialsPath(t))
	require.Error(t, statErr)
	require.True(t, os.IsNotExist(statErr))
}

// The reported session wipe: the refresh token can't be read (keyring
// unreachable in this process), which must not delete the shared session.
func TestResolveKeepsSessionWhenRefreshTokenMissing(t *testing.T) {
	setUserDirs(t)
	useFileOnlyKeyring(t)

	fixedNow := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	originalNow := now
	now = func() time.Time { return fixedNow }
	t.Cleanup(func() { now = originalNow })

	require.NoError(t, authstore.SaveSession(authstore.Session{
		IssuerURL:      "https://issuer.example.com",
		ClientID:       "client-123",
		Scopes:         []string{"openid"},
		AccessToken:    "old-access",
		Expiry:         fixedNow.Add(-1 * time.Minute),
		TokenURL:       "http://127.0.0.1:1/token",
		JWKSURL:        "https://issuer.example.com/jwks",
		Algorithms:     []string{"RS256"},
		StorageBackend: "keyring",
	}))

	_, err := Resolve(context.Background(), authconfig.Resolved{
		IssuerURL: "https://issuer.example.com",
		ClientID:  "client-123",
		Scopes:    []string{"openid"},
	})
	var authErr *autherror.Error
	require.True(t, errors.As(err, &authErr))
	require.Equal(t, "reauth_required", authErr.Envelope().Code)

	session, loadErr := authstore.LoadSession()
	require.NoError(t, loadErr)
	require.NotNil(t, session)
}

// Without the lock, invalid_grant may be a lost race with a peer that is about
// to save the rotated token, so nothing is deleted.
func TestResolveKeepsSessionOnInvalidGrantWithoutLock(t *testing.T) {
	setUserDirs(t)
	useFileOnlyKeyring(t)
	originalLockTimeout := authstore.LockTimeout
	authstore.LockTimeout = 100 * time.Millisecond
	t.Cleanup(func() { authstore.LockTimeout = originalLockTimeout })

	fixedNow := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	originalNow := now
	now = func() time.Time { return fixedNow }
	t.Cleanup(func() { now = originalNow })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()

	require.NoError(t, authstore.SaveSession(authstore.Session{
		IssuerURL:   "https://issuer.example.com",
		ClientID:    "client-123",
		Scopes:      []string{"openid"},
		AccessToken: "old-access",
		Expiry:      fixedNow.Add(-1 * time.Minute),
		TokenURL:    server.URL,
		JWKSURL:     "https://issuer.example.com/jwks",
		Algorithms:  []string{"RS256"},
	}))
	writeRefreshTokenFile(t, "refresh-1")

	lockPath, err := authstore.LockFilePath()
	require.NoError(t, err)
	holder := flock.New(lockPath)
	ok, err := holder.TryLock()
	require.NoError(t, err)
	require.True(t, ok)
	defer holder.Close()

	_, err = Resolve(context.Background(), authconfig.Resolved{
		IssuerURL: "https://issuer.example.com",
		ClientID:  "client-123",
		Scopes:    []string{"openid"},
	})
	require.Error(t, err)

	session, loadErr := authstore.LoadSession()
	require.NoError(t, loadErr)
	require.NotNil(t, session)
	_, statErr := os.Stat(credentialsPath(t))
	require.NoError(t, statErr)
}

func TestResolvePersistsVerificationMetadataForLegacySessions(t *testing.T) {
	setUserDirs(t)
	useFileOnlyKeyring(t)

	fixedNow := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	originalNow := now
	now = func() time.Time { return fixedNow }
	t.Cleanup(func() { now = originalNow })

	var issuerURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"issuer":"` + issuerURL + `","token_endpoint":"` + issuerURL + `/token","jwks_uri":"` + issuerURL + `/jwks","id_token_signing_alg_values_supported":["RS256"]}`))
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"new-access","refresh_token":"refresh-2","token_type":"Bearer","expires_in":3600}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	issuerURL = server.URL

	require.NoError(t, authstore.SaveSession(authstore.Session{
		IssuerURL:   issuerURL,
		ClientID:    "client-123",
		Scopes:      []string{"openid"},
		AccessToken: "old-access",
		TokenType:   "Bearer",
		Expiry:      fixedNow.Add(30 * time.Second),
		TokenURL:    issuerURL + "/token",
	}))
	writeRefreshTokenFile(t, "refresh-1")

	result, err := Resolve(context.Background(), authconfig.Resolved{
		IssuerURL: issuerURL,
		ClientID:  "client-123",
		Scopes:    []string{"openid"},
	})
	require.NoError(t, err)
	require.Equal(t, "new-access", result.AccessToken)

	session, err := authstore.LoadSession()
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, issuerURL+"/jwks", session.JWKSURL)
	require.Equal(t, []string{"RS256"}, session.Algorithms)
}

func writeRefreshTokenFile(t *testing.T, token string) {
	t.Helper()

	path := credentialsPath(t)
	body, err := json.Marshal(map[string]string{"refresh_token": token})
	require.NoError(t, err)
	require.NoError(t, authstore.WriteFileAtomically(path, body, 0o600))
}

func credentialsPath(t *testing.T) string {
	t.Helper()

	path, err := authstore.CredentialsFilePath()
	require.NoError(t, err)
	return path
}

func setUserDirs(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
}

func useFileOnlyKeyring(t *testing.T) {
	t.Helper()

	authstore.SetKeyringBackend(mockKeyringBackend{
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

type mockKeyringBackend struct {
	get    func(service, key string) (string, error)
	set    func(service, key, value string) error
	delete func(service, key string) error
}

func (m mockKeyringBackend) Get(service, key string) (string, error) {
	if m.get != nil {
		return m.get(service, key)
	}
	return "", nil
}

func (m mockKeyringBackend) Set(service, key, value string) error {
	if m.set != nil {
		return m.set(service, key, value)
	}
	return nil
}

func (m mockKeyringBackend) Delete(service, key string) error {
	if m.delete != nil {
		return m.delete(service, key)
	}
	return nil
}
