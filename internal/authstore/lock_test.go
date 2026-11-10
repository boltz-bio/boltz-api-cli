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

func TestWithLockTimesOutAndNamesLockPath(t *testing.T) {
	setTestDirs(t)
	t.Setenv(LockTimeoutEnv, "200ms")

	lockPath, err := LockFilePath()
	require.NoError(t, err)

	// Take the lock through WithLock so the holder is recorded, then hold it
	// from a separate flock handle while a second WithLock waits.
	_, err = WithLock(context.Background(), func() (struct{}, error) { return struct{}{}, nil })
	require.NoError(t, err)
	holder := flock.New(lockPath)
	ok, err := holder.TryLock()
	require.NoError(t, err)
	require.True(t, ok)
	defer holder.Close()

	start := time.Now()
	called := false
	_, err = WithLock(context.Background(), func() (struct{}, error) {
		called = true
		return struct{}{}, nil
	})
	require.False(t, called)
	require.Less(t, time.Since(start), 5*time.Second)

	var lockErr *LockTimeoutError
	require.True(t, errors.As(err, &lockErr), "got %v", err)
	require.Equal(t, lockPath, lockErr.Path)
	require.Contains(t, err.Error(), lockPath)
	require.Contains(t, lockErr.Holder, "pid ")
}

func TestWithLockHonorsCallerCancellation(t *testing.T) {
	setTestDirs(t)

	lockPath, err := LockFilePath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(lockPath), 0o700))
	holder := flock.New(lockPath)
	ok, err := holder.TryLock()
	require.NoError(t, err)
	require.True(t, ok)
	defer holder.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err = WithLock(ctx, func() (struct{}, error) { return struct{}{}, nil })
	require.ErrorIs(t, err, context.DeadlineExceeded)
	var lockErr *LockTimeoutError
	require.False(t, errors.As(err, &lockErr))
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

func TestHungKeyringIsBoundedAndFallsBackToFile(t *testing.T) {
	setTestDirs(t)
	t.Setenv(KeyringTimeoutEnv, "100ms")
	blockingKeyring(t)

	start := time.Now()
	backend, err := SaveRefreshToken("refresh-token")
	require.NoError(t, err)
	require.Equal(t, "file", backend)

	token, loaded, err := LoadRefreshTokenWithPreferredBackend("keyring")
	require.NoError(t, err)
	require.Equal(t, "refresh-token", token)
	require.Equal(t, "file", loaded)
	// One probe timeout; later calls skip the keyring entirely.
	require.Less(t, time.Since(start), 2*time.Second)
}

func TestKeyringFailureIsReportedNotTreatedAsMissingToken(t *testing.T) {
	setTestDirs(t)
	t.Setenv(KeyringTimeoutEnv, "100ms")

	// Probe succeeds, then the real read hangs.
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
	})
	t.Cleanup(ResetKeyring)

	_, _, err := LoadRefreshTokenWithPreferredBackend("keyring")
	require.ErrorIs(t, err, ErrKeyringTimeout)

	// The process now treats the keyring as unavailable instead of hanging again.
	_, _, err = LoadRefreshTokenWithPreferredBackend("keyring")
	require.ErrorIs(t, err, ErrKeyringUnavailable)
}

func TestDisableKeyringEnv(t *testing.T) {
	setTestDirs(t)
	t.Setenv(DisableKeyringEnv, "1")
	SetKeyringBackend(mockKeyringBackend{
		get: func(string, string) (string, error) { t.Fatal("keyring must not be called"); return "", nil },
		set: func(string, string, string) error { t.Fatal("keyring must not be called"); return nil },
	})
	t.Cleanup(ResetKeyring)

	backend, err := SaveRefreshToken("refresh-token")
	require.NoError(t, err)
	require.Equal(t, "file", backend)
}
