package authstore

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gofrs/flock"
	"github.com/stretchr/testify/require"
)

func setTestDirs(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
}

func shortTimeouts(t *testing.T) {
	t.Helper()
	keyring, lock := KeyringTimeout, LockTimeout
	KeyringTimeout, LockTimeout = 100*time.Millisecond, 200*time.Millisecond
	t.Cleanup(func() { KeyringTimeout, LockTimeout = keyring, lock })
}

// A keyring call that never returns (an unanswerable unlock prompt) is cut off.
func TestHungKeyringCallTimesOut(t *testing.T) {
	setTestDirs(t)
	shortTimeouts(t)
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	SetKeyringBackend(mockKeyringBackend{
		get: func(string, string) (string, error) { <-release; return "", nil },
		set: func(string, string, string) error { <-release; return nil },
	})
	t.Cleanup(ResetKeyring)

	start := time.Now()
	backend, err := SaveRefreshToken("refresh-token") // probe times out -> file
	require.NoError(t, err)
	require.Equal(t, "file", backend)
	require.Less(t, time.Since(start), 2*time.Second)
}

func TestWithLockProceedsWithoutLockAfterTimeout(t *testing.T) {
	setTestDirs(t)
	shortTimeouts(t)

	held, err := WithLock(context.Background(), func(lockHeld bool) (bool, error) { return lockHeld, nil })
	require.NoError(t, err)
	require.True(t, held)

	lockPath, err := LockFilePath()
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(lockPath), 0o700))
	holder := flock.New(lockPath)
	ok, err := holder.TryLock()
	require.NoError(t, err)
	require.True(t, ok)
	defer holder.Close()

	held, err = WithLock(context.Background(), func(lockHeld bool) (bool, error) { return lockHeld, nil })
	require.NoError(t, err)
	require.False(t, held)
}
