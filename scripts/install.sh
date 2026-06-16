#!/usr/bin/env sh

set -eu

version="${BOLTZ_API_VERSION:-latest}"
install_dir="${BOLTZ_API_INSTALL_DIR:-}"
install_base_url="${BOLTZ_API_INSTALL_BASE_URL:-https://install.boltz.bio/boltz-api}"
release_retries="${BOLTZ_API_RELEASE_RETRIES:-12}"
release_retry_delay="${BOLTZ_API_RELEASE_RETRY_DELAY:-10}"

need() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "boltz-api installer requires $1" >&2
    exit 1
  fi
}

need curl
need find
need grep
need install
need sed

case "$release_retries" in
  '' | *[!0-9]*)
    echo "BOLTZ_API_RELEASE_RETRIES must be a non-negative integer" >&2
    exit 1
    ;;
esac

case "$release_retry_delay" in
  '' | *[!0-9]*)
    echo "BOLTZ_API_RELEASE_RETRY_DELAY must be a non-negative integer" >&2
    exit 1
    ;;
esac

case "$(uname -s)" in
  Darwin) os="macos" ;;
  Linux) os="linux" ;;
  *)
    echo "Unsupported operating system: $(uname -s)" >&2
    exit 1
    ;;
esac

case "$os" in
  macos) config_file="${HOME}/Library/Application Support/boltz-api/config.yaml" ;;
  linux) config_file="${XDG_CONFIG_HOME:-${HOME}/.config}/boltz-api/config.yaml" ;;
esac

warn_existing_config() {
  if [ ! -f "$config_file" ]; then
    return
  fi

  config_issuer="$(sed -n 's/^[[:space:]]*issuer_url:[[:space:]]*["'\'']*\([^"'\'']*\)["'\'']*[[:space:]]*$/\1/p' "$config_file" | head -n 1)"
  config_client="$(sed -n 's/^[[:space:]]*client_id:[[:space:]]*["'\'']*\([^"'\'']*\)["'\'']*[[:space:]]*$/\1/p' "$config_file" | head -n 1)"

  if [ -n "$config_issuer" ] && [ "$config_issuer" != "https://lab.boltz.bio" ]; then
    echo "Warning: existing boltz-api config at ${config_file} sets auth issuer to ${config_issuer}." >&2
    echo "Run 'boltz-api config show' to inspect it or 'boltz-api config reset' to remove non-secret local config." >&2
  fi
  if [ -n "$config_client" ] && [ "$config_client" != "boltz-cli" ]; then
    echo "Warning: existing boltz-api config at ${config_file} sets auth client ID to ${config_client}." >&2
    echo "Run 'boltz-api config show' to inspect it or 'boltz-api config reset' to remove non-secret local config." >&2
  fi
}

warn_existing_config

case "$(uname -m)" in
  x86_64 | amd64) arch="amd64" ;;
  arm64 | aarch64) arch="arm64" ;;
  i386 | i686) arch="386" ;;
  armv*) arch="arm" ;;
  *)
    echo "Unsupported CPU architecture: $(uname -m)" >&2
    exit 1
    ;;
esac

if [ "$version" = "latest" ]; then
  tag=""
  install_base_url="${install_base_url%/}"
  release_url="${install_base_url}/latest.json"
  allow_release_fallback=1
else
  case "$version" in
    v*) tag="$version" ;;
    *) tag="v$version" ;;
  esac
  install_base_url="${install_base_url%/}"
  release_url="${install_base_url}/releases/${tag}/release.json"
  allow_release_fallback=0
fi

case "$os" in
  macos) ext="zip" ;;
  linux) ext="tar.gz" ;;
esac

retry=0
while :; do
  if ! release_json="$(curl -fsSL "$release_url")"; then
    echo "Could not fetch boltz-api release metadata from ${release_url}" >&2
    exit 1
  fi

  tag="$(printf '%s\n' "$release_json" | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' | head -n 1)"
  latest_tag="$tag"

  if [ -z "$tag" ]; then
    echo "Could not determine the boltz-api release tag" >&2
    exit 1
  fi

  asset_url="$(printf '%s\n' "$release_json" \
    | sed -n 's/.*"browser_download_url"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
    | grep "boltz-api_.*_${os}_${arch}.*\\.${ext}$" \
    | head -n 1)"

  if [ -n "$asset_url" ]; then
    asset_tag="$(printf '%s\n' "$asset_url" | sed -n 's#.*/releases/download/\([^/]*\)/.*#\1#p' | head -n 1)"
    if [ -z "$asset_tag" ]; then
      asset_tag="$(printf '%s\n' "$asset_url" | sed -n 's#.*/releases/\([^/]*\)/.*#\1#p' | head -n 1)"
    fi
    if [ -n "$asset_tag" ]; then
      tag="$asset_tag"
    fi
    if [ "$allow_release_fallback" -eq 1 ] && [ "$tag" != "$latest_tag" ]; then
      echo "Latest boltz-api release ${latest_tag} has no ${os}/${arch} asset yet; installing ${tag} instead." >&2
    fi
    break
  fi

  if [ "$retry" -ge "$release_retries" ]; then
    echo "No boltz-api release asset found for ${os}/${arch} in ${tag} after ${release_retries} retries" >&2
    exit 1
  fi

  retry=$((retry + 1))
  echo "No boltz-api release asset found for ${os}/${arch} in ${tag}; retrying in ${release_retry_delay}s (${retry}/${release_retries})" >&2
  sleep "$release_retry_delay"
done

version_number="$(printf '%s' "$tag" | sed 's/^v//')"

if [ -z "$install_dir" ]; then
  if command -v boltz-api >/dev/null 2>&1; then
    existing_binary="$(command -v boltz-api)"
    install_dir="$(dirname "$existing_binary")"
  else
    existing_binary=""
    install_dir="${HOME}/.local/bin"
  fi
else
  existing_binary="${install_dir}/boltz-api"
fi

if [ -x "$existing_binary" ]; then
  current_version="$("$existing_binary" --version 2>/dev/null | sed -n 's/.* \([0-9][0-9]*\.[0-9][^ ]*\).*/\1/p' | head -n 1 || true)"
  if [ "$current_version" = "$version_number" ]; then
    echo "boltz-api ${tag} is already installed at ${existing_binary}"
    exit 0
  fi
fi

tmpdir="$(mktemp -d 2>/dev/null || mktemp -d -t boltz-api)"
cleanup() {
  rm -rf "$tmpdir"
}
trap cleanup EXIT INT TERM

archive="${tmpdir}/boltz-api.${ext}"
curl -fL "$asset_url" -o "$archive"

case "$ext" in
  zip)
    need unzip
    unzip -q "$archive" -d "$tmpdir"
    ;;
  tar.gz)
    need tar
    tar -xzf "$archive" -C "$tmpdir"
    ;;
esac

binary="$(find "$tmpdir" -type f -name boltz-api | head -n 1)"
if [ -z "$binary" ]; then
  echo "Downloaded archive did not contain boltz-api" >&2
  exit 1
fi

mkdir -p "$install_dir"
install -m 0755 "$binary" "${install_dir}/boltz-api"

echo "Installed boltz-api ${tag} to ${install_dir}/boltz-api"

case ":$PATH:" in
  *":$install_dir:"*) ;;
  *)
    echo "Add ${install_dir} to PATH to run boltz-api without the full path." >&2
    ;;
esac
