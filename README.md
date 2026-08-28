# Boltz CLI

The official command-line tool for the [Boltz API](https://api.boltz.bio/docs).
Use it to authenticate, submit molecular modeling jobs, and download results
from a terminal, script, or coding-agent workflow.

<!-- x-release-please-start-version -->

## Installation

### Install or update

The recommended installer downloads the latest release for your platform from
Boltz's install CDN. Rerun the same command to update an existing installation.

macOS and Linux:

```sh
curl -fsSL https://install.boltz.bio/boltz-api/install.sh | sh
```

Windows PowerShell:

```powershell
irm https://install.boltz.bio/boltz-api/install.ps1 | iex
```

By default, the installer updates an existing `boltz-api` on `PATH`. If no
existing binary is found, it installs to `$HOME/.local/bin` on macOS/Linux and
`%LOCALAPPDATA%\Programs\Boltz\bin` on Windows. Set `BOLTZ_API_INSTALL_DIR` to
choose a different install directory.

After installation, update an installer-managed binary directly from the CLI:

```sh
boltz-api update
boltz-api update --check
```

The update command downloads the platform archive from the install CDN,
verifies its published SHA-256 checksum, and replaces the current executable
atomically. If the binary is managed by another package manager, use that
package manager's update command instead.

The installer uses `https://install.boltz.bio/boltz-api` for release metadata
and binary downloads. Set `BOLTZ_API_INSTALL_BASE_URL` to use a mirror. Set
`BOLTZ_API_RELEASE_RETRIES` or `BOLTZ_API_RELEASE_RETRY_DELAY` to override the
default retry count and delay.

For reproducible installs, pin a version:

```sh
curl -fsSL https://install.boltz.bio/boltz-api/install.sh | BOLTZ_API_VERSION=0.42.0 sh
```

```powershell
$env:BOLTZ_API_VERSION = "0.42.0"; irm https://install.boltz.bio/boltz-api/install.ps1 | iex
```

### Build from source with Go

To build from source, you need [Go](https://go.dev/doc/install) version 1.25 or later installed.

```sh
go install 'github.com/boltz-bio/boltz-api-cli/cmd/boltz-api@latest'
```

Once you have run `go install`, the binary is placed in your Go bin directory:

- **Default location**: `$HOME/go/bin` (or `$GOPATH/bin` if GOPATH is set)
- **Check your path**: Run `go env GOPATH` to see the base directory

If commands aren't found after installation, add the Go bin directory to your PATH:

```sh
# Add to your shell profile (.zshrc, .bashrc, etc.)
export PATH="$PATH:$(go env GOPATH)/bin"
```

<!-- x-release-please-end -->

## Usage

The CLI follows a resource-based command structure:

```sh
boltz-api [resource] <command> [flags...]
```

For example, start a structure-and-binding prediction from a YAML input file:

```sh
boltz-api predictions:structure-and-binding start \
  --input @yaml://./prediction-input.yaml \
  --model boltz-2.1
```

For commands that should wait for completion and download results, use the
resource's `run` command:

```sh
boltz-api predictions:structure-and-binding run \
  --input @yaml://./prediction-input.yaml \
  --model boltz-2.1 \
  --name aspirin-check
```

Use `--help` on any command to see its available flags:

```sh
boltz-api predictions:structure-and-binding start --help
```

### Environment variables

| Environment variable | Required | Default value |
| -------------------- | -------- | ------------- |
| `BOLTZ_API_KEY`      | no       | `null`        |
| `BOLTZ_API_NO_UPDATE_CHECK` | no | `0` |

When run interactively, the CLI checks once per day whether a newer release is
available and prints an upgrade suggestion to stderr. The check is advisory and
does not run in CI, when output is redirected, or when a custom API base URL is
configured. Set `BOLTZ_API_NO_UPDATE_CHECK=1` to disable it explicitly.

Set `BOLTZ_API_KEY` for API-key authentication. OAuth authentication can also be
configured with:

- `BOLTZ_API_AUTH_ISSUER_URL`
- `BOLTZ_API_AUTH_CLIENT_ID`
- `BOLTZ_API_AUTH_SCOPE` (comma-separated)
- `BOLTZ_API_AUTH_AUDIENCE`
- `BOLTZ_API_AUTH_AUTHORIZATION_URL`
- `BOLTZ_API_AUTH_TOKEN_URL`
- `BOLTZ_API_AUTH_USERINFO_URL`
- `BOLTZ_API_AUTH_REVOCATION_URL`
- `BOLTZ_API_ORG`
- `BOLTZ_API_LISTEN_PORT`

### Global flags

- `--api-key` (can also be set with `BOLTZ_API_KEY` env var)
- `--help` - Show command line usage
- `--debug` - Enable debug logging (includes HTTP request/response details)
- `--version`, `-v` - Show the CLI version
- `--base-url` - Use a custom API backend URL
- `--format` - Change the output format (`auto`, `explore`, `json`, `jsonl`, `pretty`, `raw`, `yaml`)
- `--format-error` - Change the output format for errors (`auto`, `explore`, `json`, `jsonl`, `pretty`, `raw`, `yaml`)
- `--transform` - Transform the data output using [GJSON syntax](https://github.com/tidwall/gjson/blob/master/SYNTAX.md). On paginated or streamed list commands, the transform runs on each item unless you use `--format raw`.
- `--transform-error` - Transform the error output using [GJSON syntax](https://github.com/tidwall/gjson/blob/master/SYNTAX.md)
- `--auth-issuer-url` - OIDC issuer URL used for OAuth login and bearer-token refresh
- `--auth-client-id` - OAuth client ID for public-client login
- `--auth-scope` - OAuth scope to request (repeatable)
- `--auth-audience` - OAuth audience/resource to request during login
- `--auth-authorization-url` - Override the discovered authorization endpoint
- `--auth-token-url` - Override the discovered token endpoint
- `--auth-userinfo-url` - Override the discovered userinfo endpoint
- `--auth-revocation-url` - Override the discovered revocation endpoint
- `--org` - Persist or override the local selected organization
- `--listen-port` - Bind the OAuth loopback listener to a specific port

### OAuth authentication

The CLI supports API-key mode and OAuth bearer-token mode. When `--api-key` or
`BOLTZ_API_KEY` is present, API-key mode wins. Otherwise the CLI uses a
stored OAuth session if one matches the configured issuer, client ID, audience,
and scopes.

Use API keys for CI and automation. Use OAuth for interactive local sessions.

Start a login flow with:

```sh
boltz-api auth login
```

For agent or MCP subprocess usage where a localhost callback is not practical,
use OAuth device authorization:

```sh
boltz-api auth login --device-code
```

Machine callers can request newline-delimited JSON events and surface the
`auth_url` event to the user:

```sh
boltz-api auth login --device-code --json-events
```

By default, OAuth login uses Boltz's first-party OAuth configuration:

- issuer: `https://lab.boltz.bio`
- client ID: `boltz-cli`
- scopes: `openid offline_access profile email compute:run`
- audience/resource: `boltz-compute-api`
- loopback callback: `http://127.0.0.1:8421/oauth/callback`

Available auth commands:

- `boltz-api auth login`
- `boltz-api auth logout`
- `boltz-api auth whoami`
- `boltz-api auth status`
- `boltz-api auth validate`
- `boltz-api auth orgs`
- `boltz-api auth wait`
- `boltz-api auth switch-org <org>`
- `boltz-api config show`
- `boltz-api config reset`

Command roles:

- `auth whoami` - concise local identity and current mode
- `auth status` - stable machine-readable auth diagnostics without refreshing
- `auth validate` - local auth check that may refresh an expired OAuth access token
- `auth orgs` - list organization IDs available to the current OAuth session or API key
- `auth switch-org` - store the OAuth organization ID to send with compute API requests
- `auth wait` - wait for usable local auth to appear, returning structured `success` or `waiting` status
- `config show` - show the path and contents of the local non-secret config file
- `config reset` - remove the local non-secret config file

`auth status`, `auth validate`, `auth orgs`, and `auth wait` return structured output.
They exit with code `1` when no usable auth mode is available. `auth status`
remains read-only; `auth validate` may refresh an expired OAuth access token
using the stored refresh token; `auth wait` stays read-only and polls local auth
state until usable auth appears or the timeout expires. In API-key mode,
`auth validate` confirms that an API key is configured locally; it does not
make a server round-trip.

For machine callers that need to wait for a browser-based login to finish:

```sh
boltz-api --format json auth wait --timeout 60s --poll-interval 2s
```

`auth status` includes the local config path and warns when the resolved OAuth
issuer or client ID looks like a placeholder. The CLI stores non-secret auth
configuration in the OS user config directory:

- macOS: `~/Library/Application Support/boltz-api/config.yaml`
- Linux: `${XDG_CONFIG_HOME:-~/.config}/boltz-api/config.yaml`
- Windows: `%APPDATA%\boltz-api\config.yaml`

The local OAuth session cache is stored in the OS user cache directory:

- macOS: `~/Library/Caches/boltz-api/session.json`
- Linux: `${XDG_CACHE_HOME:-~/.cache}/boltz-api/session.json`
- Windows: `%LocalAppData%\boltz-api\session.json`

Refresh tokens are stored in the OS keychain when available, with a fallback to
`credentials.json` in the same config directory.

### Run and download results

Resource `run` commands start a remote job, wait for completion, and write the local run directory path to stdout.
They accept the same request flags as `start`, plus local output flags such as `--name`, `--run-dir`, `--root-dir`,
`--poll-interval-seconds`, and `--progress-format`. Pipeline resources also accept `--download-mode` and `--workers`.

`download-results` creates or resumes a local run directory under `boltz-experiments/` and checkpoints progress in `.boltz-run.json`.
When `--name` and `--run-dir` are omitted, the run ID maps to a deterministic readable name such as
`boltz-experiments/lucid-atom-checks-2f7148`, so repeated downloads for the same remote run resume in the same directory.
It also writes a sanitized `run.json` for the remote run. Pipeline downloads always include a
`results/index.jsonl` manifest with one result per line. Small-molecule
pipelines (`small-molecule:design`, `small-molecule:library-screen`)
additionally get a sibling `results/summary.csv` — a scientist-friendly flat
projection of the same rows with `smiles` and `id` first, the per-row
`paths` object and `created_at` dropped (nested fields are flattened with
dotted keys; slices are encoded as JSON strings). By default, pipeline downloads use
`--download-mode everything`, which writes `results/<result-id>/metadata.json`, downloads each archive,
extracts it, and adds local artifact paths to the manifest. Use `--download-mode metadata_only` to
write only the manifest metadata.

Examples:

```sh
boltz-api predictions:structure-and-binding run --input @yaml://./prediction-input.yaml --name example-prediction
boltz-api predictions:adme run --input @json://./adme-input.json --name adme-run
boltz-api protein:design run --input @json://./protein-design-input.json --name protein-run
boltz-api download-results --id sab_pred_123
boltz-api download-results --id sab_pred_123 --name example-run
boltz-api download-results --name example-run
boltz-api download-results --id prot_des_123 --name batch-run
boltz-api download-results --id prot_des_123 --name batch-run-light --download-mode metadata_only
boltz-api download-results --id sab_pred_123 --name human-run --progress-format text --verbose
```

Use `download-status` to read the local checkpoint without making API calls:

```sh
boltz-api --format json download-status --name example-run
```

By default, `download-results` emits machine-readable JSON Lines progress events on stderr while stdout still prints the final run directory. Use `--progress-format text --verbose` for human-readable progress logs instead.

### Passing files as arguments

For API fields that accept file contents, prefix a local path with `@` and the
CLI reads the file before sending the request. The same syntax works inside JSON
or YAML values, for example `{"file_field": "@myfile.ext"}`.

To parse a file as structured JSON or YAML and inject the parsed object or
array, use `@json://...` or `@yaml://...`:

```sh
boltz-api predictions:structure-and-binding start \
  --input @yaml://./prediction-input.yaml \
  --model boltz-2.1

boltz-api predictions:structure-and-binding start <<'YAML'
input:
  entities: "@yaml://./entities.yaml"
model: boltz-2.1
YAML
```

If you need to pass a string literal that begins with an `@` sign, you can
escape the `@` sign to avoid accidentally passing a file.

```sh
boltz-api admin:workspaces create --name '\@example'
```

#### Explicit encoding

For JSON endpoints, the CLI tool does filetype sniffing to determine whether the
file contents should be sent as a string literal (for plain text files) or as a
base64-encoded string literal (for binary files). If you need to explicitly send
the file as either plain text or base64-encoded data, you can use
`@file://myfile.txt` (for string encoding) or `@data://myfile.dat` (for
base64-encoding). Use `@json://...` or `@yaml://...` only when you want the CLI
to parse the referenced file and inject structured data. Note that absolute
paths will begin with `@file://`, `@data://`, `@json://`, or `@yaml://`,
followed by a third `/` (for example, `@file:///tmp/file.txt`).

### Structured input for design and screen commands

For small-molecule/protein design and library-screen create or estimate
commands, prefer a single top-level `--input` value. The CLI merges that object
into the request body, so `idempotency_key` and `workspace_id` can still stay as
their own top-level flags:

```sh
boltz-api small-molecule:library-screen start \
  --input @json:///tmp/input.json \
  --idempotency-key req_123 \
  --workspace-id ws_123

boltz-api protein:design start \
  --input @json:///tmp/input.json \
  --idempotency-key req_123
```

Field-specific flags can override fields from `--input` when you want to tweak
part of a payload:

```sh
boltz-api small-molecule:library-screen start \
  --input @json:///tmp/input.json \
  --target @json:///tmp/target-override.json

boltz-api small-molecule:library-screen start \
  --molecule '{smiles: CCO}' \
  --molecule '{smiles: CCN}' \
  --target @json:///tmp/target.json

boltz-api protein:library-screen start \
  --protein @json:///tmp/protein-a.json \
  --protein @json:///tmp/protein-b.json \
  --target @json:///tmp/target.json
```

When piping JSON or YAML on stdin, the CLI merges that data onto the HTTP
request body, so use API body field names. For example, use `molecules` or
`proteins` in stdin payloads rather than the repeatable flag names `--molecule`
or `--protein`:

```sh
boltz-api small-molecule:library-screen start <<'YAML'
molecules:
  - smiles: CCO
  - smiles: CCN
target: {}
YAML

boltz-api protein:library-screen start <<'YAML'
proteins:
  - {}
  - {}
target: {}
YAML
```

Use `--help` on a specific command to see the repeatable flag names it accepts.

### Transform behavior

`--transform` applies to the whole response for single-object commands. On
paginated or streamed list commands, it applies to each emitted item unless you
use `--format raw`, in which case it runs on the full response page.

Examples:

```sh
# Per-item extraction on list output
boltz-api small-molecule:library-screen list-results \
  --id sm_scr_123 \
  --transform 'input_molecule.id'

# Whole-list reshaping or aggregation is better handled with jq
boltz-api small-molecule:library-screen list-results \
  --id sm_scr_123 \
  --format raw | jq '.data[] | {id, binding_confidence: .metrics.binding_confidence}'
```

Array-root expressions such as `#.{...}` are not the right tool in streamed
per-item mode.
