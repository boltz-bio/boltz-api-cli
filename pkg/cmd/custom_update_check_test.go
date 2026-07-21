// Custom CLI extension code. Not generated.
package cmd

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	boltzapi "github.com/boltz-bio/boltz-api-go"
	"github.com/stretchr/testify/require"
)

func TestShouldRunUpdateCheck(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		ci          string
		optOut      string
		interactive bool
		expected    bool
	}{
		{name: "interactive command", interactive: true, expected: true},
		{name: "non-interactive command", interactive: false, expected: false},
		{name: "ci", ci: "true", interactive: true, expected: false},
		{name: "explicit opt out", optOut: "1", interactive: true, expected: false},
		{name: "help", args: []string{"--help"}, interactive: true, expected: false},
		{name: "version", args: []string{"--version"}, interactive: true, expected: false},
		{name: "completion", args: []string{"__complete"}, interactive: true, expected: false},
		{name: "manual version endpoint", args: []string{"cli", "version"}, interactive: true, expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := shouldRunUpdateCheckWith(test.args, test.interactive, test.ci, test.optOut)
			require.Equal(t, test.expected, got)
		})
	}
}

func TestRunUpdateCheckCachesSuccessfulCheck(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("HOME", cacheHome)
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	originalNow := updateCheckNow
	originalFetch := updateCheckFetch
	originalWriter := updateCheckWriter
	t.Cleanup(func() {
		updateCheckNow = originalNow
		updateCheckFetch = originalFetch
		updateCheckWriter = originalWriter
	})

	now := time.Date(2026, time.July, 21, 12, 0, 0, 0, time.UTC)
	updateCheckNow = func() time.Time { return now }
	var fetchCalls int
	updateCheckFetch = func(context.Context) (*boltzapi.CliVersionResponse, error) {
		fetchCalls++
		return &boltzapi.CliVersionResponse{
			Latest:           "0.39.0",
			MinimumSupported: "0.8.0",
			UpdateAvailable:  true,
			Install: boltzapi.CliVersionResponseInstall{
				MacosLinux: "install command",
			},
			Message: "A newer boltz-api CLI is available.",
		}, nil
	}
	var output bytes.Buffer
	updateCheckWriter = &output

	require.NoError(t, runUpdateCheck(context.Background()))
	require.Contains(t, output.String(), "A newer boltz-api CLI is available.")
	require.Contains(t, output.String(), "Run `boltz-api update` to update")
	require.Contains(t, output.String(), "Update with: install command")
	require.Equal(t, 1, fetchCalls)

	output.Reset()
	require.NoError(t, runUpdateCheck(context.Background()))
	require.Empty(t, output.String())
	require.Equal(t, 1, fetchCalls)

	cachePath, err := updateCheckCachePath()
	require.NoError(t, err)
	cache, err := readUpdateCheckCache(cachePath)
	require.NoError(t, err)
	require.Equal(t, now, cache.LastCheckedAt)
}

func TestRunUpdateCheckDoesNotFailWhenEndpointIsUnavailable(t *testing.T) {
	cacheHome := t.TempDir()
	t.Setenv("HOME", cacheHome)
	t.Setenv("XDG_CACHE_HOME", cacheHome)

	originalFetch := updateCheckFetch
	t.Cleanup(func() { updateCheckFetch = originalFetch })
	updateCheckFetch = func(context.Context) (*boltzapi.CliVersionResponse, error) {
		return nil, os.ErrNotExist
	}

	require.NoError(t, runUpdateCheck(context.Background()))
	cachePath, err := updateCheckCachePath()
	require.NoError(t, err)
	_, err = os.Stat(cachePath)
	require.NoError(t, err)
}
