#!/usr/bin/env bash
# Download the pinned Scout binary and Clawpatch bundle from a release by version and verify every
# sha256 before use. Never builds from the checkout. Run by the reusable workflows from a clone of
# the release tag, after the workflow has validated its inputs (so none are re-checked here).
# Env in: SCOUT_VERSION (v1.2.3), RELEASE_REPO (owner/name), RUNNER_TEMP (install root),
#         GITHUB_ENV (file to append to).
# Installs into $RUNNER_TEMP/scout-release; appends SCOUT_BIN and CLAWPATCH_BIN to $GITHUB_ENV.
# Exit 3 on any download, verification or unsupported-platform failure.
# Asset names match .github/workflows/release.yml.
set -euo pipefail

die() { echo "::error::$*"; exit 3; }

# shellcheck source=scripts/sha256.sh
. "$(dirname "$0")/sha256.sh"

case "$(uname -s)-$(uname -m)" in
  Linux-x86_64) platform=linux-amd64 ;;
  Linux-aarch64 | Linux-arm64) platform=linux-arm64 ;;
  Darwin-x86_64) platform=darwin-amd64 ;;
  Darwin-arm64) platform=darwin-arm64 ;;
  *) die 'Unsupported runner platform.' ;;
esac

base="https://github.com/${RELEASE_REPO}/releases/download/${SCOUT_VERSION}"
dest="$RUNNER_TEMP/scout-release"
mkdir -p "$dest/clawpatch"
cd "$dest"
fetch() { curl --fail --silent --show-error --location --retry 3 --output "$1" "$base/$1" || die "download failed for $1."; }
verify() {
  local want got
  want="$(cut -d ' ' -f1 "$2")"
  got="$(sha256 "$1")"
  [[ -n "$want" && "$want" == "$got" ]] || die "sha256 mismatch for $1."
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
[[ "$asset" =~ ^clawpatch-[A-Za-z0-9._-]+\.tar\.gz$ ]] || die 'Unexpected Clawpatch asset name.'
[[ "$entrypoint" != /* && "$entrypoint" != *..* ]] || die 'Unexpected Clawpatch entrypoint.'
fetch "$asset"
fetch "$asset.sha256"
verify "$asset" "$asset.sha256"
[[ "$(sha256 "$asset")" == "$manifest_sha" ]] || die 'Clawpatch bundle does not match its manifest.'
tar -xzf "$asset" -C clawpatch

printf 'SCOUT_BIN=%s\n' "$dest/scout-$platform" >> "$GITHUB_ENV"
printf 'CLAWPATCH_BIN=%s\n' "$dest/clawpatch/$entrypoint" >> "$GITHUB_ENV"
