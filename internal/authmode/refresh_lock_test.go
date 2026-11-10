package authmode

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/boltz-bio/boltz-api-cli/internal/authconfig"
	"github.com/boltz-bio/boltz-api-cli/internal/autherror"
	"github.com/boltz-bio/boltz-api-cli/internal/authstore"
	"github.com/gofrs/flock"
	"github.com/stretchr/testify/require"
)

var testResolved = authconfig.Resolved{
	IssuerURL: "https://issuer.example.com",
	ClientID:  "client-123",
	Scopes:    []string{"openid"},
}

func freezeNow(t *testing.T) time.Time {
	t.Helper()
	fixed := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	original := now
	now = func() time.Time { return fixed }
	t.Cleanup(func() { now = original })
	return fixed
}

func expiredSession(fixedNow time.Time, tokenURL, backend string) authstore.Session {
	return authstore.Session{
		IssuerURL:      "https://issuer.example.com",
		ClientID:       "client-123",
		Scopes:         []string{"openid"},
		AccessToken:    "old-access",
		TokenType:      "Bearer",
		Expiry:         fixedNow.Add(-1 * time.Minute),
		TokenURL:       tokenURL,
		JWKSURL:        "https://issuer.example.com/jwks",
		Algorithms:     []string{"RS256"},
		StorageBackend: backend,
	}
}

func authErrCode(t *testing.T, err error) string {
	t.Helper()
	var authErr *autherror.Error
	require.True(t, errors.As(err, &authErr), "expected auth error, got %v", err)
	return authErr.Envelope().Code
}

func holdSessionLock(t *testing.T) *flock.Flock {
	t.Helper()
	lockPath, err := authstore.LockFilePath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(lockPath), 0o700))
	lock := flock.New(lockPath)
	ok, err := lock.TryLock()
	require.NoError(t, err)
	require.True(t, ok)
	t.Cleanup(func() { _ = lock.Close() })
	return lock
}

// The GSK report: the keyring probe fails in one process, the refresh token
// looks empty, and the session was deleted for every other process.
func TestRefreshKeepsSessionWhenKeyringUnavailable(t *testing.T) {
	setUserDirs(t)
	useFileOnlyKeyring(t)
	fixedNow := freezeNow(t)

	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	require.NoError(t, authstore.SaveSession(expiredSession(fixedNow, server.URL, "keyring")))

	_, err := Resolve(context.Background(), testResolved)
	require.Error(t, err)
	require.Equal(t, "keyring_unavailable", authErrCode(t, err))
	require.Zero(t, calls.Load(), "no token request without a refresh token")

	session, err := authstore.LoadSession()
	require.NoError(t, err)
	require.NotNil(t, session, "session.json must survive")
}

func TestRefreshWithNoRefreshTokenKeepsSession(t *testing.T) {
	setUserDirs(t)
	useFileOnlyKeyring(t)
	fixedNow := freezeNow(t)

	require.NoError(t, authstore.SaveSession(expiredSession(fixedNow, "http://127.0.0.1:1", "file")))

	_, err := Resolve(context.Background(), testResolved)
	require.Equal(t, "reauth_required", authErrCode(t, err))

	session, err := authstore.LoadSession()
	require.NoError(t, err)
	require.NotNil(t, session)
}

// A process rotated the refresh token but session.json was not rewritten (it
// crashed in between). invalid_grant on the stale token re-reads the store and
// retries once with the rotated token.
func TestRefreshRetriesWithRotatedTokenOnInvalidGrant(t *testing.T) {
	setUserDirs(t)
	useFileOnlyKeyring(t)
	fixedNow := freezeNow(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		w.Header().Set("Content-Type", "application/json")
		switch r.PostForm.Get("refresh_token") {
		case "refresh-1":
			// Simulate the other process's rotation landing in the store.
			writeRefreshTokenFile(t, "refresh-2")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
		case "refresh-2":
			_, _ = w.Write([]byte(`{"access_token":"new-access","refresh_token":"refresh-3","token_type":"Bearer","expires_in":600}`))
		default:
			t.Errorf("unexpected refresh token %q", r.PostForm.Get("refresh_token"))
		}
	}))
	defer server.Close()

	require.NoError(t, authstore.SaveSession(expiredSession(fixedNow, server.URL, "file")))
	writeRefreshTokenFile(t, "refresh-1")

	result, err := Resolve(context.Background(), testResolved)
	require.NoError(t, err)
	require.Equal(t, "new-access", result.AccessToken)

	token, _, err := authstore.LoadRefreshTokenWithPreferredBackend("file")
	require.NoError(t, err)
	require.Equal(t, "refresh-3", token)
}

// A process waiting on the lock uses the session the holder refreshed
// instead of refreshing again with an already-rotated token.
func TestRefreshUsesSessionRefreshedByLockHolder(t *testing.T) {
	setUserDirs(t)
	useFileOnlyKeyring(t)
	fixedNow := freezeNow(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("waiter must not call the token endpoint")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	require.NoError(t, authstore.SaveSession(expiredSession(fixedNow, server.URL, "file")))
	writeRefreshTokenFile(t, "refresh-1")

	lock := holdSessionLock(t)
	go func() {
		time.Sleep(200 * time.Millisecond)
		refreshed := expiredSession(fixedNow, server.URL, "file")
		refreshed.AccessToken = "holder-access"
		refreshed.Expiry = fixedNow.Add(10 * time.Minute)
		_ = authstore.SaveSession(refreshed)
		_ = lock.Close()
	}()

	result, err := Resolve(context.Background(), testResolved)
	require.NoError(t, err)
	require.Equal(t, "holder-access", result.AccessToken)
}

// A stuck lock holder must not block everyone: after the lock timeout the
// refresh proceeds without the lock.
func TestRefreshProceedsWithoutLockAfterTimeout(t *testing.T) {
	setUserDirs(t)
	useFileOnlyKeyring(t)
	fixedNow := freezeNow(t)
	t.Setenv("BOLTZ_API_TEST_LOCK_TIMEOUT", "200ms")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-access","refresh_token":"refresh-2","token_type":"Bearer","expires_in":600}`))
	}))
	defer server.Close()

	require.NoError(t, authstore.SaveSession(expiredSession(fixedNow, server.URL, "file")))
	writeRefreshTokenFile(t, "refresh-1")
	holdSessionLock(t)

	result, err := Resolve(context.Background(), testResolved)
	require.NoError(t, err)
	require.Equal(t, "new-access", result.AccessToken)
}

// Without the lock, invalid_grant may mean a peer just spent the same token
// and is about to write its replacement, so nothing is deleted.
func TestInvalidGrantWithoutLockKeepsSession(t *testing.T) {
	setUserDirs(t)
	useFileOnlyKeyring(t)
	fixedNow := freezeNow(t)
	t.Setenv("BOLTZ_API_TEST_LOCK_TIMEOUT", "200ms")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()

	require.NoError(t, authstore.SaveSession(expiredSession(fixedNow, server.URL, "file")))
	writeRefreshTokenFile(t, "refresh-1")
	holdSessionLock(t)

	_, err := Resolve(context.Background(), testResolved)
	require.Equal(t, "reauth_required", authErrCode(t, err))

	session, err := authstore.LoadSession()
	require.NoError(t, err)
	require.NotNil(t, session)
	_, statErr := os.Stat(credentialsPath(t))
	require.NoError(t, statErr)
}

// A rotated token that can't be written back to the keyring is reported (log in
// again), not silently moved to plain text.
func TestRefreshDoesNotDowngradeKeyringToFile(t *testing.T) {
	setUserDirs(t)
	fixedNow := freezeNow(t)
	authstore.SetKeyringBackend(mockKeyringBackend{
		get: func(service, key string) (string, error) {
			if service == "boltz-api-cli-probe" {
				return "", nil
			}
			return "refresh-1", nil
		},
		set: func(service, key, value string) error { return errors.New("keyring write failed") },
	})
	t.Cleanup(authstore.ResetKeyring)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"new-access","refresh_token":"refresh-2","token_type":"Bearer","expires_in":600}`))
	}))
	defer server.Close()
	require.NoError(t, authstore.SaveSession(expiredSession(fixedNow, server.URL, "keyring")))

	_, err := Resolve(context.Background(), testResolved)
	require.Equal(t, "reauth_required", authErrCode(t, err))
	require.Contains(t, err.Error(), "keyring write failed")

	_, statErr := os.Stat(credentialsPath(t))
	require.True(t, os.IsNotExist(statErr), "refresh token must not be written to plain text")
}

// Unexpired sessions never touch the lock.
func TestResolveWithFreshTokenIgnoresHeldLock(t *testing.T) {
	setUserDirs(t)
	useFileOnlyKeyring(t)
	fixedNow := freezeNow(t)

	fresh := expiredSession(fixedNow, "http://127.0.0.1:1", "file")
	fresh.Expiry = fixedNow.Add(10 * time.Minute)
	require.NoError(t, authstore.SaveSession(fresh))

	holdSessionLock(t)

	result, err := Resolve(context.Background(), testResolved)
	require.NoError(t, err)
	require.Equal(t, "old-access", result.AccessToken)
}
