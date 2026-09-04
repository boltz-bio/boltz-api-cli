// Custom CLI extension code. Not generated.
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/boltz-bio/boltz-api-cli/internal/authconfig"
	"github.com/boltz-bio/boltz-api-cli/internal/authstore"
	boltzapi "github.com/boltz-bio/boltz-api-go"
	"github.com/boltz-bio/boltz-api-go/option"
	"github.com/boltz-bio/boltz-api-go/packages/param"
	"github.com/urfave/cli/v3"
)

const (
	updateCheckCacheFile = "update-check.json"
	updateCheckInterval  = 24 * time.Hour
	updateCheckTimeout   = 750 * time.Millisecond
	updateCheckOptOutEnv = "BOLTZ_API_NO_UPDATE_CHECK"
)

type updateCheckCache struct {
	LastCheckedAt time.Time `json:"last_checked_at"`
}

var (
	updateCheckNow              = func() time.Time { return time.Now().UTC() }
	updateCheckFetch            = fetchCLIUpdateStatus
	updateCheckWriter io.Writer = os.Stderr
)

func init() {
	// Keep the update check outside generated files so it survives SDK regeneration.
	Command.Before = updateCheckBefore
	Command.Commands = append(Command.Commands, &updateCommand)
}

func updateCheckBefore(ctx context.Context, cmd *cli.Command) (context.Context, error) {
	if !shouldRunUpdateCheck(os.Args[1:]) || hasCustomBaseURL(cmd) {
		return ctx, nil
	}

	// The check is advisory. It must never prevent a command from running because
	// the cache directory, network, or version endpoint is unavailable.
	_ = runUpdateCheck(ctx)
	return ctx, nil
}

func shouldRunUpdateCheck(args []string) bool {
	return shouldRunUpdateCheckWith(
		args,
		isInteractiveUpdateCheck(),
		os.Getenv("CI"),
		os.Getenv(updateCheckOptOutEnv),
	)
}

func shouldRunUpdateCheckWith(args []string, interactive bool, ci, optOut string) bool {
	if !interactive || isTruthyEnv(ci) || isTruthyEnv(optOut) {
		return false
	}

	for _, arg := range args {
		switch arg {
		case "--help", "-h", "--version", "-v", "-V", "__complete", "cli", "update":
			return false
		}
	}

	return true
}

func isInteractiveUpdateCheck() bool {
	return isTerminal(os.Stdout)
}

func isTruthyEnv(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

func hasCustomBaseURL(cmd *cli.Command) bool {
	_, source, err := authconfig.ResolveBaseURL(cmd)
	if err != nil {
		return true
	}
	return source != authconfig.SourceDefault
}

func runUpdateCheck(ctx context.Context) error {
	cachePath, err := updateCheckCachePath()
	if err != nil {
		return err
	}

	now := updateCheckNow()
	cache, cacheErr := readUpdateCheckCache(cachePath)
	if cacheErr == nil && isFreshUpdateCheck(cache.LastCheckedAt, now) {
		return nil
	}

	// Record failed attempts too. A network outage should not add a request to
	// every CLI invocation until the outage is resolved.
	cache = updateCheckCache{LastCheckedAt: now}
	defer func() { _ = writeUpdateCheckCache(cachePath, cache) }()

	status, err := updateCheckFetch(ctx)
	if err != nil || status == nil || !status.UpdateAvailable {
		return nil
	}

	writeUpdateNotice(updateCheckWriter, status)
	return nil
}

func isFreshUpdateCheck(lastCheckedAt, now time.Time) bool {
	if lastCheckedAt.IsZero() || now.Before(lastCheckedAt) {
		return false
	}
	return now.Sub(lastCheckedAt) < updateCheckInterval
}

func updateCheckCachePath() (string, error) {
	dir, err := authstore.CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, updateCheckCacheFile), nil
}

func readUpdateCheckCache(path string) (updateCheckCache, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return updateCheckCache{}, err
	}

	var cache updateCheckCache
	if err := json.Unmarshal(body, &cache); err != nil {
		return updateCheckCache{}, err
	}
	return cache, nil
}

func writeUpdateCheckCache(path string, cache updateCheckCache) error {
	body, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	return authstore.WriteFileAtomically(path, append(body, '\n'), 0o600)
}

func fetchCLIUpdateStatus(parent context.Context) (*boltzapi.CliVersionResponse, error) {
	ctx, cancel := context.WithTimeout(parent, updateCheckTimeout)
	defer cancel()

	client := boltzapi.NewCliService(
		option.WithEnvironmentProduction(),
		option.WithHTTPClient(&http.Client{Timeout: updateCheckTimeout}),
		option.WithRequestTimeout(updateCheckTimeout),
		option.WithMaxRetries(0),
		option.WithHeader("User-Agent", fmt.Sprintf("Boltz/CLI %s", Version)),
		option.WithHeader("X-Stainless-Lang", "cli"),
		option.WithHeader("X-Stainless-Package-Version", Version),
		option.WithHeader("X-Stainless-Runtime", "go"),
	)

	return client.Version(ctx, boltzapi.CliVersionParams{
		Current:  param.NewOpt(Version),
		Platform: param.NewOpt(runtime.GOOS + "-" + runtime.GOARCH),
	})
}

func writeUpdateNotice(w io.Writer, status *boltzapi.CliVersionResponse) {
	if w == nil || status == nil || !status.UpdateAvailable {
		return
	}

	message := strings.TrimSpace(status.Message)
	if message == "" {
		message = "A newer boltz-api CLI is available."
	}

	_, _ = fmt.Fprintf(w, "\n%s (current %s, latest %s)\n", message, Version, status.Latest)
	_, _ = fmt.Fprintln(w, "Run `boltz-api update` to update, or rerun the installer:")
	if runtime.GOOS == "windows" {
		if command := strings.TrimSpace(status.Install.Windows); command != "" {
			_, _ = fmt.Fprintf(w, "Update with: %s\n", command)
		}
	} else if command := strings.TrimSpace(status.Install.MacosLinux); command != "" {
		_, _ = fmt.Fprintf(w, "Update with: %s\n", command)
	}
	_, _ = fmt.Fprintf(w, "Set %s=1 to disable this check.\n", updateCheckOptOutEnv)
}
