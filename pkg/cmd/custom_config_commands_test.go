// Custom CLI extension code. Not generated.
package cmd

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/boltz-bio/boltz-api-cli/internal/authconfig"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestConfigShowAndReset(t *testing.T) {
	setConfigCommandUserDirs(t)

	output, err := runConfigCommand(t, "--format", "json", "config", "set", "--base-url", "https://api.customer.example.com")
	require.NoError(t, err)

	var set map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &set))
	require.Equal(t, "https://api.customer.example.com", set["base_url"])
	require.NotEmpty(t, set["path"])
	info, statErr := os.Stat(set["path"].(string))
	require.NoError(t, statErr)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{
		IssuerURL: "https://issuer.example.com",
		ClientID:  "client-123",
		Scopes:    []string{"openid", "email"},
		Audience:  authconfig.DefaultAudience,
	}))

	output, err = runConfigCommand(t, "--format", "json", "config", "show")
	require.NoError(t, err)

	var show map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &show))
	require.Equal(t, true, show["present"])
	require.NotEmpty(t, show["path"])
	config := show["config"].(map[string]any)
	require.Equal(t, "https://issuer.example.com", config["issuer_url"])
	require.Equal(t, "client-123", config["client_id"])
	require.Equal(t, "https://api.customer.example.com", config["base_url"])

	output, err = runConfigCommand(t, "--format", "json", "config", "reset")
	require.NoError(t, err)

	var reset map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &reset))
	require.Equal(t, true, reset["removed"])

	output, err = runConfigCommand(t, "--format", "json", "config", "show")
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal([]byte(output), &show))
	require.Equal(t, false, show["present"])
}

func TestConfigSetRejectsInvalidBaseURL(t *testing.T) {
	setConfigCommandUserDirs(t)

	_, err := runConfigCommand(t, "config", "set", "--base-url", "api.customer.example.com")
	require.ErrorContains(t, err, "missing a scheme")

	config, loadErr := authconfig.Load()
	require.NoError(t, loadErr)
	require.Empty(t, config.BaseURL)
}

func runConfigCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	defer r.Close()

	root := &cli.Command{
		Name:           "boltz-api",
		Writer:         w,
		ErrWriter:      w,
		ExitErrHandler: func(context.Context, *cli.Command, error) {},
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "format", Value: "auto"},
			&cli.BoolFlag{Name: "raw-output"},
			&cli.StringFlag{Name: "transform"},
		},
		Commands: []*cli.Command{configCommand},
	}
	runErr := root.Run(context.Background(), append([]string{"boltz-api"}, args...))
	require.NoError(t, w.Close())

	output, readErr := io.ReadAll(r)
	require.NoError(t, readErr)
	return strings.TrimSpace(string(output)), runErr
}

func setConfigCommandUserDirs(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
}
