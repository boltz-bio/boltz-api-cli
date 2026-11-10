#!/usr/bin/env bash
# Reproduces the "session.lock held for hours / session.json deleted" report on
# Linux with a fake Secret Service whose unlock prompt never completes (headless
# host) or is dismissed. Builds two CLI binaries and runs the same scenarios
# against each.
#
#   scripts/repro/keyring-hang/run.sh [OLD_REF]   # OLD_REF defaults to origin/main
#
# Requires Docker. The "new" binary is built from the current working tree.
set -euo pipefail
here="$(cd "$(dirname "$0")" && pwd)"
root="$(cd "$here/../../.." && pwd)"
old_ref="${1:-origin/main}"
work="$(mktemp -d)"
trap 'git -C "$root" worktree remove --force "$work/oldsrc" >/dev/null 2>&1 || true; rm -rf "$work"' EXIT

case "$(docker info --format '{{.Architecture}}')" in
  aarch64|arm64) arch=arm64 ;;
  *) arch=amd64 ;;
esac
mkdir -p "$work/bin"
cp "$here/repro.sh" "$here/fake_secret_service.py" "$work/"

echo "building old binary from $old_ref and new binary from the working tree (linux/$arch)..."
git -C "$root" worktree add -q --detach "$work/oldsrc" "$old_ref"
(cd "$work/oldsrc" && GOOS=linux GOARCH=$arch CGO_ENABLED=0 go build -o "$work/bin/boltz-api-old" ./cmd/boltz-api)
(cd "$root" && GOOS=linux GOARCH=$arch CGO_ENABLED=0 go build -o "$work/bin/boltz-api-new" ./cmd/boltz-api)

docker build -q -t boltz-keyring-repro "$here" >/dev/null
for v in old new; do
  docker run --rm -v "$work:/repro" boltz-keyring-repro bash /repro/repro.sh "$v"
done
