package authstore

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
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

var (
	// KeyringTimeout bounds each keyring call. The keyring libraries can block
	// forever (Secret Service waits on an unlock prompt nobody can answer on a
	// headless host; macOS runs /usr/bin/security unbounded). 60s matches
	// GitHub CLI, long enough to answer a real prompt.
	KeyringTimeout = 60 * time.Second
	// LockTimeout bounds the wait for another process's refresh. After it, the
	// caller proceeds without the lock (as GitHub CLI does), so one stuck
	// process can't block every invocation.
	LockTimeout = 60 * time.Second

	// ErrKeyringTimeout is returned when a keyring call exceeds KeyringTimeout.
	ErrKeyringTimeout = errors.New("system keyring did not respond in time")

	keyringBackend KeyringBackend = defaultKeyringBackend{}

	keyringAvailability   *bool
	keyringAvailabilityMu sync.Mutex
)

// withKeyringTimeout runs a keyring call, giving up after KeyringTimeout. The
// call can't be cancelled, so on timeout it is left running in its goroutine.
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
	select {
	case r := <-done:
		return r.value, r.err
	case <-time.After(KeyringTimeout):
		var zero T
		return zero, ErrKeyringTimeout
	}
}

func keyringGet(service, key string) (string, error) {
	b := keyringBackend
	return withKeyringTimeout(func() (string, error) { return b.Get(service, key) })
}

func keyringSet(service, key, value string) error {
	b := keyringBackend
	_, err := withKeyringTimeout(func() (struct{}, error) { return struct{}{}, b.Set(service, key, value) })
	return err
}

func keyringDelete(service, key string) error {
	b := keyringBackend
	_, err := withKeyringTimeout(func() (struct{}, error) { return struct{}{}, b.Delete(service, key) })
	return err
}

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

func SaveRefreshToken(token string) (string, error) {
	if strings.TrimSpace(token) == "" {
		return "", ClearRefreshToken()
	}

	if KeyringAvailable() && keyringSet(keyringService, refreshTokenAccount, token) == nil {
		if path, pathErr := CredentialsFilePath(); pathErr == nil {
			_ = os.Remove(path)
		}
		return "keyring", nil
	}

	if KeyringAvailable() {
		_ = keyringDelete(keyringService, refreshTokenAccount)
	}

	path, err := CredentialsFilePath()
	if err != nil {
		return "", err
	}
	body, err := json.MarshalIndent(refreshTokenFile{RefreshToken: token}, "", "  ")
	if err != nil {
		return "", err
	}
	if err := WriteFileAtomically(path, append(body, '\n'), 0o600); err != nil {
		return "", err
	}
	return "file", nil
}

func LoadRefreshToken() (string, string, error) {
	return LoadRefreshTokenWithPreferredBackend("")
}

func LoadRefreshTokenWithPreferredBackend(preferred string) (string, string, error) {
	switch strings.TrimSpace(strings.ToLower(preferred)) {
	case "keyring":
		return loadRefreshTokenFromKeyring()
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
	if KeyringAvailable() {
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
	return "", "", nil
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

	_, err := keyringGet(keyringService+"-probe", "availability-check")
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

// WithLock runs fn while holding the cross-process session lock, which stops
// concurrent processes from each spending the single-use rotating refresh
// token. The lock is best effort: if it isn't acquired within LockTimeout, fn
// runs anyway with lockHeld=false and must then skip destructive cleanup, since
// a peer may be mid-refresh. Caller cancellation is returned as an error.
func WithLock[T any](ctx context.Context, fn func(lockHeld bool) (T, error)) (T, error) {
	var zero T

	lockPath, err := LockFilePath()
	if err != nil {
		return zero, err
	}
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o700); err != nil {
		return zero, err
	}

	lockCtx, cancel := context.WithTimeout(ctx, LockTimeout)
	defer cancel()
	lock := flock.New(lockPath)
	ok, err := lock.TryLockContext(lockCtx, lockRetryDelay)
	if ctx.Err() != nil {
		return zero, ctx.Err()
	}
	if ok && err == nil {
		defer lock.Close()
	}
	return fn(ok && err == nil)
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
