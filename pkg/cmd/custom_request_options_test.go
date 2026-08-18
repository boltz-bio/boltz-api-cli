// Custom CLI extension code. Not generated.
package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/boltz-bio/boltz-api-cli/internal/authconfig"
	"github.com/boltz-bio/boltz-api-cli/internal/authstore"
	"github.com/boltz-bio/boltz-api-cli/internal/requestflag"
	githubcomboltzbioboltzcomputeapigo "github.com/boltz-bio/boltz-api-go"
	"github.com/boltz-bio/boltz-api-go/option"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestRequestOptionsKeepAPIKeyMode(t *testing.T) {
	setAuthCommandUserDirs(t)
	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{SelectedOrg: "org-config"}))

	var gotAPIKey string
	var gotAuthorization string
	var gotOrganization string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("x-api-key")
		gotAuthorization = r.Header.Get("Authorization")
		gotOrganization = r.Header.Get("X-Boltz-Organization-Id")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cmd := parsedTestCommand(t,
		"--base-url", server.URL,
		"--api-key", "api-key-123",
		"call",
	)

	client := githubcomboltzbioboltzcomputeapigo.NewClient(getDefaultRequestOptions(cmd)...)
	var result map[string]any
	require.NoError(t, client.Get(context.Background(), "/check", nil, &result))
	require.Equal(t, "api-key-123", gotAPIKey)
	require.Empty(t, gotAuthorization)
	require.Empty(t, gotOrganization)
}

func TestRequestOptionsUseStoredBaseURL(t *testing.T) {
	setAuthCommandUserDirs(t)
	t.Setenv(authconfig.EnvBaseURL, "")

	var requested bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requested = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()
	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{BaseURL: server.URL}))

	cmd := parsedTestCommand(t, "--api-key", "api-key-123", "call")
	client := githubcomboltzbioboltzcomputeapigo.NewClient(getDefaultRequestOptions(cmd)...)
	var result map[string]any
	require.NoError(t, client.Get(context.Background(), "/check", nil, &result))
	require.True(t, requested)
}

func TestRequestOptionsBaseURLPrecedence(t *testing.T) {
	setAuthCommandUserDirs(t)

	stored := httptest.NewServer(http.NotFoundHandler())
	defer stored.Close()
	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{BaseURL: stored.URL}))

	var envRequested bool
	envServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		envRequested = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer envServer.Close()
	t.Setenv(authconfig.EnvBaseURL, envServer.URL)

	cmd := parsedTestCommand(t, "--api-key", "api-key-123", "call")
	client := githubcomboltzbioboltzcomputeapigo.NewClient(getDefaultRequestOptions(cmd)...)
	var result map[string]any
	require.NoError(t, client.Get(context.Background(), "/check", nil, &result))
	require.True(t, envRequested)

	var flagRequested bool
	flagServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flagRequested = true
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer flagServer.Close()

	cmd = parsedTestCommand(t, "--base-url", flagServer.URL, "--api-key", "api-key-123", "call")
	client = githubcomboltzbioboltzcomputeapigo.NewClient(getDefaultRequestOptions(cmd)...)
	require.NoError(t, client.Get(context.Background(), "/check", nil, &result))
	require.True(t, flagRequested)
}

func TestRequestOptionsInjectBearerTokenAndRemoveInheritedAPIKey(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)

	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{
		IssuerURL:   "https://issuer.example.com",
		ClientID:    "client-123",
		Audience:    authconfig.DefaultAudience,
		Scopes:      []string{"openid", "profile"},
		SelectedOrg: "org-oauth-selected",
	}))
	require.NoError(t, authstore.SaveSession(authstore.Session{
		IssuerURL:   "https://issuer.example.com",
		ClientID:    "client-123",
		Audience:    authconfig.DefaultAudience,
		Scopes:      []string{"openid", "profile"},
		AccessToken: "oauth-access",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(10 * time.Minute),
	}))

	var gotAPIKey string
	var gotAuthorization string
	var gotOrganization string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("x-api-key")
		gotAuthorization = r.Header.Get("Authorization")
		gotOrganization = r.Header.Get("X-Boltz-Organization-Id")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cmd := parsedTestCommand(t,
		"--base-url", server.URL,
		"call",
	)

	options := append(getDefaultRequestOptions(cmd), option.WithAPIKey("sdk-default-key"))
	client := githubcomboltzbioboltzcomputeapigo.NewClient(options...)
	var result map[string]any
	require.NoError(t, client.Get(context.Background(), "/check", nil, &result))
	require.Empty(t, gotAPIKey)
	require.Equal(t, "Bearer oauth-access", gotAuthorization)
	require.Equal(t, "org-oauth-selected", gotOrganization)
}

func parsedTestCommand(t *testing.T, args ...string) *cli.Command {
	t.Helper()

	var captured *cli.Command
	root := &cli.Command{
		Name: "boltz-api",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "base-url"},
			&requestflag.Flag[string]{
				Name:    "api-key",
				Sources: cli.EnvVars("BOLTZ_API_KEY"),
			},
		},
		Commands: []*cli.Command{
			{
				Name: "call",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					_ = ctx
					captured = cmd
					return nil
				},
			},
		},
	}
	root.Flags = append(root.Flags, authFlags()...)

	runArgs := append([]string{"boltz-api"}, args...)
	require.NoError(t, root.Run(context.Background(), runArgs))
	require.NotNil(t, captured)
	return captured
}

func TestRequestOptionsSetClientHeader(t *testing.T) {
	setAuthCommandUserDirs(t)
	require.NoError(t, authconfig.SaveProfile(authconfig.Resolved{SelectedOrg: "org-config"}))
	t.Setenv("BOLTZ_API_CLIENT", "claude-desktop-mcp")

	var gotClient string
	var gotClientVersion string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClient = r.Header.Get("X-Boltz-Client")
		gotClientVersion = r.Header.Get("X-Boltz-Client-Version")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	cmd := parsedTestCommand(t,
		"--base-url", server.URL,
		"--api-key", "api-key-123",
		"call",
	)

	client := githubcomboltzbioboltzcomputeapigo.NewClient(getDefaultRequestOptions(cmd)...)
	var result map[string]any
	require.NoError(t, client.Get(context.Background(), "/check", nil, &result))
	require.Equal(t, "claude-desktop-mcp", gotClient)
	require.Equal(t, Version, gotClientVersion)
}

func TestDetectClientFromEnv(t *testing.T) {
	cases := []struct {
		name    string
		env     map[string]string
		environ []string
		want    string
	}{
		{
			name: "explicit BOLTZ_API_CLIENT wins over host detection",
			env:  map[string]string{"BOLTZ_API_CLIENT": "claude-desktop-mcp", "CLAUDECODE": "1"},
			want: "claude-desktop-mcp",
		},
		{
			name: "trims whitespace around explicit value",
			env:  map[string]string{"BOLTZ_API_CLIENT": "  codex  "},
			want: "codex",
		},
		{
			name: "explicit value with allowed punctuation is accepted",
			env:  map[string]string{"BOLTZ_API_CLIENT": "claude-desktop-mcp"},
			want: "claude-desktop-mcp",
		},
		{
			name: "malformed explicit value (control char) is ignored and falls through",
			env:  map[string]string{"BOLTZ_API_CLIENT": "bad\nvalue"},
			want: "boltz-cli",
		},
		{
			name: "CLAUDECODE marks claude-code",
			env:  map[string]string{"CLAUDECODE": "1"},
			want: "claude-code",
		},
		{
			name: "CLAUDE_CODE_ENTRYPOINT marks claude-code",
			env:  map[string]string{"CLAUDE_CODE_ENTRYPOINT": "cli"},
			want: "claude-code",
		},
		{
			name:    "a CODEX_* var marks codex",
			environ: []string{"CODEX_SANDBOX=seatbelt"},
			want:    "codex",
		},
		{
			name:    "a GEMINI_* var marks gemini",
			environ: []string{"GEMINI_CLI=1"},
			want:    "gemini",
		},
		{
			name: "no signals falls back to boltz-cli",
			want: "boltz-cli",
		},
		{
			name:    "blank BOLTZ_API_CLIENT is ignored and falls through to host detection",
			env:     map[string]string{"BOLTZ_API_CLIENT": "   "},
			environ: []string{"CODEX_HOME=/home/.codex"},
			want:    "codex",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(key string) string { return tc.env[key] }
			if got := detectClientFromEnv(getenv, tc.environ); got != tc.want {
				t.Fatalf("detectClientFromEnv() = %q, want %q", got, tc.want)
			}
		})
	}
}
