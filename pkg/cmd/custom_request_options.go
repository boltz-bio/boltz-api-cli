// Custom CLI extension code. Not generated.
package cmd

import (
	"net/http"
	"os"
	"strings"

	"github.com/boltz-bio/boltz-api-cli/internal/authconfig"
	"github.com/boltz-bio/boltz-api-cli/internal/authmode"
	"github.com/boltz-bio/boltz-api-go/option"
	"github.com/urfave/cli/v3"
)

func additionalRequestOptions(cmd *cli.Command) []option.RequestOption {
	root := cmd
	if root != nil {
		if resolvedRoot := root.Root(); resolvedRoot != nil {
			root = resolvedRoot
		}
	}
	if root == nil {
		return nil
	}

	return []option.RequestOption{
		option.WithMiddleware(func(r *http.Request, mn option.MiddlewareNext) (*http.Response, error) {
			// Identify the originating integration for server-side usage analytics.
			// Advisory only (never auth-bearing); the server validates against an allowlist.
			r.Header.Set("X-Boltz-Client", detectClient())
			r.Header.Set("X-Boltz-Client-Version", Version)

			resolved, err := authconfig.Resolve(root)
			if err != nil {
				return nil, authmode.WrapConfigError(err)
			}

			auth, err := authmode.Resolve(r.Context(), resolved)
			if err != nil {
				return nil, err
			}

			if auth.Mode == authmode.ModeOAuth {
				r.Header.Del("x-api-key")
				r.Header.Set("Authorization", "Bearer "+auth.AccessToken)
				if selectedOrg := strings.TrimSpace(resolved.SelectedOrg); selectedOrg != "" {
					r.Header.Set("X-Boltz-Organization-Id", selectedOrg)
				}
			}

			return mn(r)
		}),
	}
}

// detectClient resolves which integration is driving the CLI, for usage analytics.
func detectClient() string {
	return detectClientFromEnv(os.Getenv, os.Environ())
}

// detectClientFromEnv is the pure core of detectClient (env injected for testability).
// Precedence: an explicit BOLTZ_API_CLIENT (set by wrappers such as the MCP plugin)
// wins; otherwise the host agent is inferred from its environment; otherwise the
// caller is a bare CLI invocation.
func detectClientFromEnv(getenv func(string) string, environ []string) string {
	// A malformed explicit value is ignored (not sent) so it can never produce a
	// header value that net/http rejects, which would otherwise break every request.
	if explicit := strings.TrimSpace(getenv("BOLTZ_API_CLIENT")); explicit != "" && isCleanClientToken(explicit) {
		return explicit
	}
	if getenv("CLAUDECODE") != "" || getenv("CLAUDE_CODE_ENTRYPOINT") != "" {
		return "claude-code"
	}
	if hasEnvWithPrefix(environ, "CODEX_") {
		return "codex"
	}
	if hasEnvWithPrefix(environ, "GEMINI_") {
		return "gemini"
	}
	return "boltz-cli"
}

func hasEnvWithPrefix(environ []string, prefix string) bool {
	for _, kv := range environ {
		if strings.HasPrefix(kv, prefix) {
			return true
		}
	}
	return false
}

// isCleanClientToken reports whether s is a safe client id to send as an HTTP
// header value: a run of [A-Za-z0-9._-]. Guards against a malformed
// BOLTZ_API_CLIENT producing a header value that net/http would reject.
func isCleanClientToken(s string) bool {
	for _, r := range s {
		isLetter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		isDigit := r >= '0' && r <= '9'
		if !isLetter && !isDigit && r != '.' && r != '_' && r != '-' {
			return false
		}
	}
	return true
}
