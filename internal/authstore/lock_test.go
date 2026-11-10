package authstore

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofrs/flock"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"
)

func setTestDirs(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
}

func holdLock(t *testing.T) *flock.Flock {
	t.Helper()
	lockPath, err := LockFilePath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(lockPath), 0o700))
	holder := flock.New(lockPath)
	ok, err := holder.TryLock()
	require.NoError(t, err)
	require.True(t, ok)
	t.Cleanup(func() { _ = holder.Close() })
	return holder
}

func TestWithLockReportsLockHeld(t *testing.T) {
	setTestDirs(t)

	held, err := WithLock(context.Background(), func(lockHeld bool) (bool, error) { return lockHeld, nil })
	require.NoError(t, err)
	require.True(t, held)
}

// A stuck holder must not wedge other processes: after the lock timeout the
// caller proceeds without the lock and is told so.
func TestWithLockProceedsWithoutLockAfterTimeout(t *testing.T) {
	setTestDirs(t)
	t.Setenv(testLockTimeoutEnv, "200ms")
	holdLock(t)

	start := time.Now()
	held, err := WithLock(context.Background(), func(lockHeld bool) (bool, error) { return lockHeld, nil })
	require.NoError(t, err)
	require.False(t, held)
	require.Less(t, time.Since(start), 5*time.Second)
}

func TestWithLockHonorsCallerCancellation(t *testing.T) {
	setTestDirs(t)
	holdLock(t)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	called := false
	_, err := WithLock(ctx, func(bool) (struct{}, error) { called = true; return struct{}{}, nil })
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.False(t, called)
}

// blockingKeyring simulates a Secret Service unlock prompt nobody answers.
func blockingKeyring(t *testing.T) {
	t.Helper()
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	block := func() { <-release }
	SetKeyringBackend(mockKeyringBackend{
		get:    func(string, string) (string, error) { block(); return "", nil },
		set:    func(string, string, string) error { block(); return nil },
		delete: func(string, string) error { block(); return nil },
	})
	t.Cleanup(ResetKeyring)
}

// probeOKThenBlock passes the availability probe, then hangs every real call.
func probeOKThenBlock(t *testing.T) {
	t.Helper()
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	SetKeyringBackend(mockKeyringBackend{
		get: func(service, _ string) (string, error) {
			if service == keyringService+"-probe" {
				return "", keyring.ErrNotFound
			}
			<-release
			return "", nil
		},
		set:    func(string, string, string) error { <-release; return nil },
		delete: func(string, string) error { <-release; return nil },
	})
	t.Cleanup(ResetKeyring)
}

// At login, a hung keyring falls back to the file (GitHub CLI's behavior)
// after one timeout, without waiting again to delete the stale entry.
func TestLoginSaveFallsBackToFileWhenKeyringHangs(t *testing.T) {
	setTestDirs(t)
	t.Setenv(testKeyringTimeoutEnv, "150ms")
	probeOKThenBlock(t)

	start := time.Now()
	backend, err := SaveRefreshToken("refresh-token")
	require.NoError(t, err)
	require.Equal(t, "file", backend)
	require.Less(t, time.Since(start), 290*time.Millisecond, "one keyring timeout, not two")

	token, loaded, err := LoadRefreshTokenWithPreferredBackend("file")
	require.NoError(t, err)
	require.Equal(t, "refresh-token", token)
	require.Equal(t, "file", loaded)
}

func TestHungProbeIsBounded(t *testing.T) {
	setTestDirs(t)
	t.Setenv(testKeyringTimeoutEnv, "100ms")
	blockingKeyring(t)

	start := time.Now()
	_, _, err := LoadRefreshTokenWithPreferredBackend("keyring")
	require.ErrorIs(t, err, ErrKeyringUnavailable)
	require.Less(t, time.Since(start), 2*time.Second)
}

// A keyring failure is reported as such, never as "no token", and (like
// GitHub CLI) the keyring is retried on the next call rather than written off.
func TestKeyringTimeoutIsReportedAndRetried(t *testing.T) {
	setTestDirs(t)
	t.Setenv(testKeyringTimeoutEnv, "100ms")
	probeOKThenBlock(t)

	_, _, err := LoadRefreshTokenWithPreferredBackend("keyring")
	require.ErrorIs(t, err, ErrKeyringTimeout)
	_, _, err = LoadRefreshTokenWithPreferredBackend("keyring")
	require.ErrorIs(t, err, ErrKeyringTimeout)
}

// A rotated token goes back to the backend it came from; a keyring failure is
// an error rather than a silent move to plain text.
func TestPersistRefreshTokenKeepsBackend(t *testing.T) {
	setTestDirs(t)
	t.Setenv(testKeyringTimeoutEnv, "100ms")
	probeOKThenBlock(t)

	_, err := PersistRefreshToken("rotated", "keyring")
	require.ErrorIs(t, err, ErrKeyringTimeout)
	path, err := CredentialsFilePath()
	require.NoError(t, err)
	_, statErr := os.Stat(path)
	require.True(t, errors.Is(statErr, os.ErrNotExist), "must not fall back to plain text")

	backend, err := PersistRefreshToken("rotated", "file")
	require.NoError(t, err)
	require.Equal(t, "file", backend)
}

func TestSaveRefreshTokenToFileSkipsKeyring(t *testing.T) {
	setTestDirs(t)
	SetKeyringBackend(mockKeyringBackend{
		get: func(string, string) (string, error) { t.Fatal("keyring must not be called"); return "", nil },
		set: func(string, string, string) error { t.Fatal("keyring must not be called"); return nil },
	})
	t.Cleanup(ResetKeyring)

	require.NoError(t, SaveRefreshTokenToFile("refresh-token"))
	token, backend, err := LoadRefreshTokenWithPreferredBackend("file")
	require.NoError(t, err)
	require.Equal(t, "refresh-token", token)
	require.Equal(t, "file", backend)
}
