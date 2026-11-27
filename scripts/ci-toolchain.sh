#!/usr/bin/env bash
# Provision the pinned scanner toolchain from the Scout release tag being used, never from the
# assessed checkout. Run by the reusable workflows from a clone of the release tag (this script's
# own clone; it does not clone again): installs bubblewrap on Linux runners, then runs this clone's
# scripts/bootstrap-tools.sh into TOOLCHAIN_DIR/tools.
# Env in: TOOLCHAIN_DIR (directory outside the workspace; TOOLCHAIN_DIR/tools is the cacheable
#         install), GITHUB_ENV, GITHUB_PATH (files to append to).
# Appends SCOUT_TOOLS_DIR to $GITHUB_ENV and TOOLCHAIN_DIR/tools/bin to $GITHUB_PATH.
set -euo pipefail

src="$(cd "$(dirname "$0")/.." && pwd)"
bash "$src/scripts/ci-sandbox.sh"

toolchain_dir="${TOOLCHAIN_DIR:?TOOLCHAIN_DIR must name a directory outside the workspace}"
github_env="${GITHUB_ENV:?GITHUB_ENV must name the workflow environment file}"
github_path="${GITHUB_PATH:?GITHUB_PATH must name the workflow path file}"
SCOUT_TOOLS_DIR="$toolchain_dir/tools" bash "$src/scripts/bootstrap-tools.sh"
printf 'SCOUT_TOOLS_DIR=%s\n' "$toolchain_dir/tools" >>"$github_env"
printf '%s\n' "$toolchain_dir/tools/bin" >>"$github_path"
