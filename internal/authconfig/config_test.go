package authconfig

import (
	"context"
	"testing"

	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestResolvePrecedenceAndScopeReplacement(t *testing.T) {
	setUserDirs(t)
	require.NoError(t, SaveProfile(Resolved{
		BaseURL:     "https://config-api.example.com",
		IssuerURL:   "https://config.example.com",
		ClientID:    "config-client",
		Scopes:      []string{"config.read", "config.write"},
		Audience:    "config-audience",
		SelectedOrg: "config-org",
	}))

	t.Setenv(EnvAuthIssuerURL, "https://env.example.com")
	t.Setenv(EnvAuthClientID, "env-client")
	t.Setenv(EnvAuthScope, "env.read,env.write")
	t.Setenv(EnvAuthAudience, "env-audience")
	t.Setenv(EnvOrg, "env-org")
	t.Setenv(EnvBaseURL, "https://env-api.example.com")

	resolved := resolveForArgs(t,
		"--auth-client-id", "flag-client",
		"--auth-scope", "flag.read",
		"--org", "flag-org",
	)

	require.Equal(t, "https://env.example.com", resolved.IssuerURL)
	require.Equal(t, "https://env-api.example.com", resolved.BaseURL)
	require.Equal(t, "flag-client", resolved.ClientID)
	require.Equal(t, []string{"flag.read"}, resolved.Scopes)
	require.Equal(t, "env-audience", resolved.Audience)
	require.Equal(t, "flag-org", resolved.SelectedOrg)
	require.Equal(t, SourceRuntime, resolved.Sources.IssuerURL)
	require.Equal(t, SourceRuntime, resolved.Sources.BaseURL)
	require.Equal(t, SourceRuntime, resolved.Sources.ClientID)
	require.Equal(t, SourceRuntime, resolved.Sources.Scopes)
	require.Equal(t, SourceRuntime, resolved.Sources.Audience)
	require.Equal(t, SourceRuntime, resolved.Sources.SelectedOrg)
}

func TestResolveBaseURLPrecedence(t *testing.T) {
	setUserDirs(t)
	require.NoError(t, SaveProfile(Resolved{BaseURL: "https://config-api.example.com"}))

	resolved := resolveForArgs(t)
	require.Equal(t, "https://config-api.example.com", resolved.BaseURL)
	require.Equal(t, SourceConfig, resolved.Sources.BaseURL)

	t.Setenv(EnvBaseURL, "https://env-api.example.com")
	resolved = resolveForArgs(t)
	require.Equal(t, "https://env-api.example.com", resolved.BaseURL)
	require.Equal(t, SourceRuntime, resolved.Sources.BaseURL)

	resolved = resolveForArgs(t, "--base-url", "https://flag-api.example.com")
	require.Equal(t, "https://flag-api.example.com", resolved.BaseURL)
	require.Equal(t, SourceRuntime, resolved.Sources.BaseURL)
}

func TestResolveFallsBackToBoltzOAuthDefaults(t *testing.T) {
	setUserDirs(t)

	resolved := resolveForArgs(t)
	require.Equal(t, DefaultIssuerURL, resolved.IssuerURL)
	require.Equal(t, DefaultClientID, resolved.ClientID)
	require.Equal(t, []string{"openid", "offline_access", "profile", "email", "compute:run"}, resolved.Scopes)
	require.Equal(t, DefaultAudience, resolved.Audience)
	require.Equal(t, DefaultListenPort, resolved.ListenPort)
	require.Empty(t, resolved.BaseURL)
	require.Equal(t, SourceDefault, resolved.Sources.BaseURL)
	require.Equal(t, SourceDefault, resolved.Sources.IssuerURL)
	require.Equal(t, SourceDefault, resolved.Sources.ClientID)
	require.Equal(t, SourceDefault, resolved.Sources.Scopes)
	require.Equal(t, SourceDefault, resolved.Sources.Audience)
	require.Equal(t, SourceDefault, resolved.Sources.ListenPort)
}

func TestSaveProfilePersistsBaseURL(t *testing.T) {
	setUserDirs(t)
	require.NoError(t, SaveProfile(Resolved{
		BaseURL:   "https://api.customer.example.com",
		IssuerURL: "https://issuer.example.com",
		ClientID:  "client-123",
		Scopes:    []string{"openid"},
	}))

	config, err := Load()
	require.NoError(t, err)
	require.Equal(t, "https://api.customer.example.com", config.BaseURL)
}

func TestResolveRejectsInvalidStoredBaseURL(t *testing.T) {
	setUserDirs(t)
	require.NoError(t, SaveProfile(Resolved{BaseURL: "api.customer.example.com"}))

	_, err := resolveErrForArgs(t)
	require.ErrorContains(t, err, "config base_url")
	require.ErrorContains(t, err, "missing a scheme")
}

func TestSaveProfileWritesConfigVersion(t *testing.T) {
	setUserDirs(t)

	require.NoError(t, SaveProfile(Resolved{
		IssuerURL: "https://issuer.example.com",
		ClientID:  "client-123",
		Scopes:    []string{"openid"},
	}))

	config, err := Load()
	require.NoError(t, err)
	require.Equal(t, ConfigVersion, config.Version)
}

func resolveForArgs(t *testing.T, args ...string) Resolved {
	t.Helper()

	resolved, err := resolveErrForArgs(t, args...)
	require.NoError(t, err)
	return resolved
}

func resolveErrForArgs(t *testing.T, args ...string) (Resolved, error) {
	t.Helper()

	var captured *cli.Command
	command := &cli.Command{
		Name: "boltz-api",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "base-url"},
			&requestflag.Flag[string]{
				Name:    "api-key",
				Sources: cli.EnvVars("BOLTZ_API_KEY"),
			},
			&cli.StringFlag{Name: "auth-issuer-url", Sources: cli.EnvVars(EnvAuthIssuerURL)},
			&cli.StringFlag{Name: "auth-client-id", Sources: cli.EnvVars(EnvAuthClientID)},
			&cli.StringSliceFlag{Name: "auth-scope", Sources: cli.EnvVars(EnvAuthScope)},
			&cli.StringFlag{Name: "auth-audience", Sources: cli.EnvVars(EnvAuthAudience)},
			&cli.StringFlag{Name: "auth-authorization-url", Sources: cli.EnvVars(EnvAuthAuthorizationURL)},
			&cli.StringFlag{Name: "auth-token-url", Sources: cli.EnvVars(EnvAuthTokenURL)},
			&cli.StringFlag{Name: "auth-userinfo-url", Sources: cli.EnvVars(EnvAuthUserInfoURL)},
			&cli.StringFlag{Name: "auth-revocation-url", Sources: cli.EnvVars(EnvAuthRevocationURL)},
			&cli.StringFlag{Name: "org", Sources: cli.EnvVars(EnvOrg)},
			&cli.IntFlag{Name: "listen-port", Sources: cli.EnvVars(EnvListenPort)},
		},
		Commands: []*cli.Command{
			{
				Name: "inspect",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					_ = ctx
					captured = cmd
					return nil
				},
			},
		},
	}

	runArgs := append([]string{"boltz-api"}, args...)
	runArgs = append(runArgs, "inspect")
	require.NoError(t, command.Run(context.Background(), runArgs))
	require.NotNil(t, captured)

	return Resolve(captured)
}

func setUserDirs(t *testing.T) {
	t.Helper()

	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
}
