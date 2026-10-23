#!/usr/bin/env bash
# Provision the pinned scanner toolchain from the Scout release tag being used,
# never from the assessed checkout. The release tag is cloned outside the
# workspace and its scripts/bootstrap-tools.sh installs checksum-verified tools
# into TOOLCHAIN_DIR/tools; PATH and SCOUT_TOOLS_DIR are exported for later steps.
# Env: SCOUT_VERSION (v1.2.3), RELEASE_REPO (owner/name), TOOLCHAIN_DIR (directory
# outside the workspace; TOOLCHAIN_DIR/tools is the cacheable install).
set -euo pipefail

[[ "$SCOUT_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.]+)?$ ]] || { echo '::error::scout-version must be a release tag such as v0.1.0.'; exit 3; }
[[ "$RELEASE_REPO" =~ ^[A-Za-z0-9._-]+/[A-Za-z0-9._-]+$ ]] || { echo '::error::release-repo must be owner/name.'; exit 3; }

if [[ "$(uname -s)" == Linux ]]; then
  sudo apt-get update -qq
  sudo apt-get install -y bubblewrap
  # Ubuntu 24.04 blocks unprivileged user namespaces, which bubblewrap needs.
  sudo sysctl -w kernel.apparmor_restrict_unprivileged_userns=0 > /dev/null 2>&1 || true
fi

src="$TOOLCHAIN_DIR/source"
rm -rf "$src"
git clone --quiet --depth 1 --branch "$SCOUT_VERSION" --recurse-submodules --shallow-submodules \
  "https://github.com/${RELEASE_REPO}.git" "$src"
SCOUT_TOOLS_DIR="$TOOLCHAIN_DIR/tools" bash "$src/scripts/bootstrap-tools.sh"
printf 'SCOUT_TOOLS_DIR=%s\n' "$TOOLCHAIN_DIR/tools" >> "$GITHUB_ENV"
printf '%s\n' "$TOOLCHAIN_DIR/tools/bin" >> "$GITHUB_PATH"
