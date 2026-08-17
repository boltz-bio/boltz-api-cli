package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/boltz-bio/boltz-api-cli/internal/authconfig"
	"github.com/stretchr/testify/require"
)

func TestConfigSetPersistsBaseURLForOrdinaryCommands(t *testing.T) {
	binary := buildCLIBinary(t)
	// Empty-but-set is ignored by base-URL resolution; setting it before
	// authProcessEnv snapshots the environment neutralizes any BOLTZ_BASE_URL
	// inherited from the caller's shell.
	t.Setenv(authconfig.EnvBaseURL, "")
	env := authProcessEnv(t)

	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"latest":"0.39.0","minimum_supported":"0.8.0","update_available":false}`))
	}))
	defer server.Close()

	set := runCLI(t, binary, env, "--format", "json", "config", "set", "--base-url", server.URL)
	require.Equal(t, 0, set.ExitCode, set.Stderr)

	var setResponse map[string]any
	require.NoError(t, json.Unmarshal([]byte(set.Stdout), &setResponse))
	require.Equal(t, server.URL, setResponse["base_url"])

	result := runCLI(t, binary, env, "--api-key", "test-key", "--format", "json", "cli", "version")
	require.Equal(t, 0, result.ExitCode, result.Stderr)
	require.Equal(t, "/compute/v1/cli/version", requestedPath)
}
