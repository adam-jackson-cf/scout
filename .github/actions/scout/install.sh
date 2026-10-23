#!/usr/bin/env bash
# Download the pinned Scout binary and Clawpatch bundle from a release by
# version and verify every sha256 before use. Never builds from the checkout.
# Env: SCOUT_VERSION (v1.2.3), RELEASE_REPO (owner/name), DEST (directory).
# Asset names match .github/workflows/release.yml.
set -euo pipefail

[[ "$SCOUT_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.]+)?$ ]] || { echo '::error::scout-version must be a release tag such as v0.1.0.'; exit 3; }
[[ "$RELEASE_REPO" =~ ^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$ ]] || { echo '::error::release-repo must be owner/name.'; exit 3; }

case "$(uname -s)-$(uname -m)" in
  Linux-x86_64) platform=linux-amd64 ;;
  Linux-aarch64 | Linux-arm64) platform=linux-arm64 ;;
  Darwin-x86_64) platform=darwin-amd64 ;;
  Darwin-arm64) platform=darwin-arm64 ;;
  *) echo '::error::Unsupported runner platform.'; exit 3 ;;
esac

if command -v sha256sum > /dev/null; then
  digest() { sha256sum "$1" | cut -d ' ' -f1; }
else
  digest() { shasum -a 256 "$1" | cut -d ' ' -f1; }
fi

base="https://github.com/${RELEASE_REPO}/releases/download/${SCOUT_VERSION}"
mkdir -p "$DEST/clawpatch"
cd "$DEST"
fetch() { curl --fail --silent --show-error --location --retry 3 --output "$1" "$base/$1"; }
verify() {
  local want got
  want="$(cut -d ' ' -f1 "$2")"
  got="$(digest "$1")"
  [[ -n "$want" && "$want" == "$got" ]] || { echo "::error::sha256 mismatch for $1."; exit 3; }
}

fetch "scout-$platform"
fetch "scout-$platform.sha256"
verify "scout-$platform" "scout-$platform.sha256"
chmod +x "scout-$platform"

fetch clawpatch-manifest.json
fetch clawpatch-manifest.json.sha256
verify clawpatch-manifest.json clawpatch-manifest.json.sha256
asset="$(jq -er .asset clawpatch-manifest.json)"
entrypoint="$(jq -er .entrypoint clawpatch-manifest.json)"
manifest_sha="$(jq -er .sha256 clawpatch-manifest.json)"
[[ "$asset" =~ ^clawpatch-[A-Za-z0-9._-]+\.tar\.gz$ ]] || { echo '::error::Unexpected Clawpatch asset name.'; exit 3; }
[[ "$entrypoint" != /* && "$entrypoint" != *..* ]] || { echo '::error::Unexpected Clawpatch entrypoint.'; exit 3; }
fetch "$asset"
fetch "$asset.sha256"
verify "$asset" "$asset.sha256"
[[ "$(digest "$asset")" == "$manifest_sha" ]] || { echo '::error::Clawpatch bundle does not match its manifest.'; exit 3; }
tar -xzf "$asset" -C clawpatch

printf 'SCOUT_BIN=%s\n' "$DEST/scout-$platform" >> "$GITHUB_ENV"
printf 'CLAWPATCH_BIN=%s\n' "$DEST/clawpatch/$entrypoint" >> "$GITHUB_ENV"
