package authstore

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofrs/flock"
	"github.com/zalando/go-keyring"
)

const (
	productDir          = "boltz-api"
	keyringService      = "boltz-api-cli"
	refreshTokenAccount = "oauth-refresh-token"
	lockRetryDelay      = 50 * time.Millisecond
	testDisableKeyring  = "BOLTZ_API_TEST_DISABLE_KEYRING"

	// Test-only overrides (like testDisableKeyring) so the repro and subprocess
	// tests don't have to wait out the production timeouts.
	testKeyringTimeoutEnv = "BOLTZ_API_TEST_KEYRING_TIMEOUT"
	testLockTimeoutEnv    = "BOLTZ_API_TEST_LOCK_TIMEOUT"

	// defaultKeyringTimeout bounds each keyring call. It matches GitHub CLI,
	// which raised its timeout from 3s to 60s because shorter limits cut off
	// users answering an interactive unlock prompt (cli/cli#13787).
	defaultKeyringTimeout = 60 * time.Second
	// defaultLockTimeout bounds the wait for another process's refresh, after
	// which the caller proceeds without the lock (as GitHub CLI does), so a
	// stuck holder can never wedge every invocation.
	defaultLockTimeout = 60 * time.Second
)

var (
	// ErrKeyringUnavailable means the system keyring could not be reached (or
	// was disabled), so a keyring-backed refresh token cannot be read right now.
	// It does not mean the token is gone.
	ErrKeyringUnavailable = errors.New("system keyring is unavailable")
	// ErrKeyringTimeout means a keyring call did not return in time. This is
	// typically a Secret Service unlock prompt nobody can answer (headless
	// host) or a hung Keychain call.
	ErrKeyringTimeout = errors.New("system keyring did not respond in time")
)

type KeyringBackend interface {
	Get(service, key string) (string, error)
	Set(service, key, value string) error
	Delete(service, key string) error
}

type defaultKeyringBackend struct{}

func (d defaultKeyringBackend) Get(service, key string) (string, error) {
	return keyring.Get(service, key)
}

func (d defaultKeyringBackend) Set(service, key, value string) error {
	return keyring.Set(service, key, value)
}

func (d defaultKeyringBackend) Delete(service, key string) error {
	return keyring.Delete(service, key)
}

// durationFromEnv parses a Go duration ("30s") or a number of seconds ("30").
func durationFromEnv(name string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	if d, err := time.ParseDuration(raw); err == nil && d > 0 {
		return d
	}
	if secs, err := strconv.ParseFloat(raw, 64); err == nil && secs > 0 {
		return time.Duration(secs * float64(time.Second))
	}
	return fallback
}

func keyringTimeout() time.Duration {
	return durationFromEnv(testKeyringTimeoutEnv, defaultKeyringTimeout)
}

func lockTimeout() time.Duration {
	return durationFromEnv(testLockTimeoutEnv, defaultLockTimeout)
}

// withKeyringTimeout runs a keyring call with a deadline. The keyring libraries
// have no cancellation (Secret Service waits on an unlock prompt forever; macOS
// runs /usr/bin/security with no timeout), so on timeout the call is abandoned
// in its goroutine and ErrKeyringTimeout is returned.
func withKeyringTimeout[T any](fn func() (T, error)) (T, error) {
	type result struct {
		value T
		err   error
	}
	done := make(chan result, 1)
	go func() {
		value, err := fn()
		done <- result{value, err}
	}()

	timer := time.NewTimer(keyringTimeout())
	defer timer.Stop()
	select {
	case r := <-done:
		return r.value, r.err
	case <-timer.C:
		var zero T
		return zero, ErrKeyringTimeout
	}
}

func keyringGet(service, key string) (string, error) {
	return withKeyringTimeout(func() (string, error) { return currentKeyringBackend().Get(service, key) })
}

func keyringSet(service, key, value string) error {
	_, err := withKeyringTimeout(func() (struct{}, error) {
		return struct{}{}, currentKeyringBackend().Set(service, key, value)
	})
	return err
}

func keyringDelete(service, key string) error {
	_, err := withKeyringTimeout(func() (struct{}, error) {
		return struct{}{}, currentKeyringBackend().Delete(service, key)
	})
	return err
}

func currentKeyringBackend() KeyringBackend {
	keyringAvailabilityMu.Lock()
	defer keyringAvailabilityMu.Unlock()
	return keyringBackend
}

var (
	keyringBackend KeyringBackend = defaultKeyringBackend{}

	keyringAvailability   *bool
	keyringAvailabilityMu sync.Mutex
)

type Identity struct {
	Subject           string         `json:"subject,omitempty"`
	Email             string         `json:"email,omitempty"`
	Name              string         `json:"name,omitempty"`
	PreferredUsername string         `json:"preferred_username,omitempty"`
	Claims            map[string]any `json:"claims,omitempty"`
}

type Session struct {
	IssuerURL        string    `json:"issuer_url"`
	ClientID         string    `json:"client_id"`
	Audience         string    `json:"audience,omitempty"`
	Scopes           []string  `json:"scopes,omitempty"`
	GrantedScopes    []string  `json:"granted_scopes,omitempty"`
	AccessToken      string    `json:"access_token"`
	TokenType        string    `json:"token_type,omitempty"`
	Expiry           time.Time `json:"expiry,omitempty"`
	IDToken          string    `json:"id_token,omitempty"`
	AuthorizationURL string    `json:"authorization_url,omitempty"`
	TokenURL         string    `json:"token_url,omitempty"`
	UserInfoURL      string    `json:"userinfo_url,omitempty"`
	RevocationURL    string    `json:"revocation_url,omitempty"`
	JWKSURL          string    `json:"jwks_url,omitempty"`
	Algorithms       []string  `json:"algorithms,omitempty"`
	StorageBackend   string    `json:"storage_backend,omitempty"`
	Identity         Identity  `json:"identity,omitempty"`
	CreatedAt        time.Time `json:"created_at,omitempty"`
	UpdatedAt        time.Time `json:"updated_at,omitempty"`
}

type refreshTokenFile struct {
	RefreshToken string `json:"refresh_token"`
}

func ConfigDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, productDir), nil
}

func CacheDir() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, productDir), nil
}

func ConfigFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func CredentialsFilePath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

func SessionFilePath() (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.json"), nil
}

func LockFilePath() (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.lock"), nil
}

func LoadSession() (*Session, error) {
	path, err := SessionFilePath()
	if err != nil {
		return nil, err
	}
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var session Session
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func SaveSession(session Session) error {
	now := time.Now().UTC()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	session.UpdatedAt = now

	path, err := SessionFilePath()
	if err != nil {
		return err
	}
	body, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomically(path, append(body, '\n'), 0o600)
}

func ClearSession() error {
	path, err := SessionFilePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// SaveRefreshToken stores a refresh token at login: in the system keyring
// when it works, otherwise in the 0600 credentials file (GitHub CLI's
// behavior). It returns the backend used; callers should tell the user when
// that is "file".
func SaveRefreshToken(token string) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", ClearRefreshToken()
	}

	if KeyringAvailable() {
		setErr := keyringSet(keyringService, refreshTokenAccount, token)
		if setErr == nil {
			if path, pathErr := CredentialsFilePath(); pathErr == nil {
				_ = os.Remove(path)
			}
			return "keyring", nil
		}
		// Drop a stale keyring entry so it can't shadow the file, unless the
		// keyring just hung: don't wait on it a second time.
		if !errors.Is(setErr, ErrKeyringTimeout) {
			_ = keyringDelete(keyringService, refreshTokenAccount)
		}
	}
	return "file", saveRefreshTokenFile(token)
}

// SaveRefreshTokenToFile stores a refresh token in the 0600 credentials file
// without touching the keyring (`auth login --insecure-storage`).
func SaveRefreshTokenToFile(token string) error {
	return saveRefreshTokenFile(token)
}

// PersistRefreshToken writes a rotated refresh token back to the backend the
// session already uses. Like GitHub CLI it never silently moves a keyring
// credential to plain text: if the keyring write fails the error is returned.
// Sessions with no recorded backend get SaveRefreshToken's behavior.
func PersistRefreshToken(token, backend string) (string, error) {
	switch strings.TrimSpace(strings.ToLower(backend)) {
	case "keyring":
		if !KeyringAvailable() {
			return "", ErrKeyringUnavailable
		}
		if err := keyringSet(keyringService, refreshTokenAccount, token); err != nil {
			return "", err
		}
		return "keyring", nil
	case "file":
		return "file", saveRefreshTokenFile(token)
	default:
		return SaveRefreshToken(token)
	}
}

func saveRefreshTokenFile(token string) error {
	path, err := CredentialsFilePath()
	if err != nil {
		return err
	}
	body, err := json.MarshalIndent(refreshTokenFile{RefreshToken: token}, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomically(path, append(body, '\n'), 0o600)
}

func LoadRefreshToken() (string, string, error) {
	return LoadRefreshTokenWithPreferredBackend("")
}

func LoadRefreshTokenWithPreferredBackend(preferred string) (string, string, error) {
	switch strings.TrimSpace(strings.ToLower(preferred)) {
	case "keyring":
		token, backend, err := loadRefreshTokenFromKeyring()
		if err == nil {
			return token, backend, nil
		}
		// The keyring is unreachable or failing. A token in the credentials file
		// is still usable; otherwise report the keyring error rather than "no
		// token", so callers don't mistake a transient keyring failure for a
		// logged-out session.
		if fileToken, fileBackend, fileErr := loadRefreshTokenFromFile(); fileErr == nil && strings.TrimSpace(fileToken) != "" {
			return fileToken, fileBackend, nil
		}
		return "", "", err
	case "file":
		return loadRefreshTokenFromFile()
	}

	token, backend, err := loadRefreshTokenFromKeyring()
	switch {
	case err == nil && strings.TrimSpace(token) != "":
		return token, backend, nil
	case err == nil:
		// Fall through to file-backed credentials.
	case errors.Is(err, keyring.ErrNotFound):
		// Fall through to file-backed credentials.
	default:
		// Fall back to file-backed credentials on any keyring backend error.
	}

	return loadRefreshTokenFromFile()
}

func loadRefreshTokenFromKeyring() (string, string, error) {
	if !KeyringAvailable() {
		return "", "", ErrKeyringUnavailable
	}
	token, err := keyringGet(keyringService, refreshTokenAccount)
	switch {
	case err == nil && strings.TrimSpace(token) != "":
		return token, "keyring", nil
	case err == nil:
		return "", "", nil
	case errors.Is(err, keyring.ErrNotFound):
		return "", "", nil
	default:
		return "", "", err
	}
}

func loadRefreshTokenFromFile() (string, string, error) {
	path, err := CredentialsFilePath()
	if err != nil {
		return "", "", err
	}
	body, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return "", "", nil
	}
	if err != nil {
		return "", "", err
	}
	var credentials refreshTokenFile
	if err := json.Unmarshal(body, &credentials); err != nil {
		return "", "", err
	}
	if strings.TrimSpace(credentials.RefreshToken) == "" {
		return "", "", nil
	}
	return credentials.RefreshToken, "file", nil
}

func ClearRefreshToken() error {
	var keyringErr error
	if KeyringAvailable() {
		if err := keyringDelete(keyringService, refreshTokenAccount); err != nil && !errors.Is(err, keyring.ErrNotFound) {
			keyringErr = err
		}
	}

	path, err := CredentialsFilePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return keyringErr
}

func KeyringAvailable() bool {
	keyringAvailabilityMu.Lock()
	defer keyringAvailabilityMu.Unlock()

	if os.Getenv(testDisableKeyring) != "" {
		disabled := false
		keyringAvailability = &disabled
		return false
	}

	if keyringAvailability != nil {
		return *keyringAvailability
	}

	backend := keyringBackend
	_, err := withKeyringTimeout(func() (string, error) {
		return backend.Get(keyringService+"-probe", "availability-check")
	})
	available := err == nil || errors.Is(err, keyring.ErrNotFound)
	keyringAvailability = &available
	return available
}

func SetKeyringBackend(backend KeyringBackend) {
	keyringAvailabilityMu.Lock()
	defer keyringAvailabilityMu.Unlock()

	if backend == nil {
		backend = defaultKeyringBackend{}
	}
	keyringBackend = backend
	keyringAvailability = nil
}

func ResetKeyring() {
	SetKeyringBackend(defaultKeyringBackend{})
}

func ResetKeyringAvailability() {
	keyringAvailabilityMu.Lock()
	defer keyringAvailabilityMu.Unlock()
	keyringAvailability = nil
}

func ClearRefreshTokenFile() error {
	path, err := CredentialsFilePath()
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func ClearRefreshTokenKeyring() error {
	if !KeyringAvailable() {
		return nil
	}
	if err := keyringDelete(keyringService, refreshTokenAccount); err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return err
	}
	return nil
}

func ClearAll() error {
	if err := ClearSession(); err != nil {
		return err
	}
	return ClearRefreshToken()
}

// refreshMu serializes WithLock callers inside one process; the file lock
// serializes separate processes.
var refreshMu sync.Mutex

// WithLock runs fn while holding the cross-process session lock, which exists
// to stop concurrent processes from each spending the single-use rotating
// refresh token. Following GitHub CLI, the lock is best effort: if it cannot be
// acquired within the lock timeout (a stuck holder, or a filesystem without
// advisory locks) fn still runs, with lockHeld=false, so a stuck process can
// never wedge every invocation. fn must skip destructive cleanup when
// lockHeld is false, since a peer may be mid-refresh. Caller cancellation is
// still returned as an error.
func WithLock[T any](ctx context.Context, fn func(lockHeld bool) (T, error)) (T, error) {
	refreshMu.Lock()
	defer refreshMu.Unlock()

	lockHeld := false
	if lockPath, err := LockFilePath(); err == nil && os.MkdirAll(filepath.Dir(lockPath), 0o700) == nil {
		lockCtx, cancel := context.WithTimeout(ctx, lockTimeout())
		lock := flock.New(lockPath)
		ok, lockErr := lock.TryLockContext(lockCtx, lockRetryDelay)
		cancel()
		if ctx.Err() != nil {
			var zero T
			return zero, ctx.Err()
		}
		if ok && lockErr == nil {
			lockHeld = true
			defer lock.Close()
		}
	}

	return fn(lockHeld)
}

func WriteFileAtomically(path string, data []byte, perm fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
