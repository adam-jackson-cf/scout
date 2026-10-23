#!/usr/bin/env bash
# Builds the reviewed Clawpatch CLI without modifying its source checkout.
# The revision is the third_party/clawpatch submodule gitlink; versions come from toolchain/versions.env.
set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd -P)"
repo_root="$(cd "$script_dir/.." && pwd -P)"
safety_patch="$script_dir/clawpatch-safe.patch"

manifest_value() {
  awk -F '=' -v key="$1" '$1 == key { print $2 }' "$repo_root/toolchain/versions.env"
}

CLAWPATCH_REVISION="$(git -C "$repo_root" ls-files -s third_party/clawpatch | awk '$1 == "160000" { print $2 }')"
CLAWPATCH_VERSION="$(manifest_value CLAWPATCH_VERSION)"
PNPM_VERSION="$(manifest_value PNPM_VERSION)"
readonly CLAWPATCH_REVISION CLAWPATCH_VERSION PNPM_VERSION
readonly SCOUT_REVIEW_SAFETY_PATCH_VERSION="scout-review-sanitize-v3"
readonly SCOUT_REVIEW_SAFETY_PATCH_SHA256="dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3"

[[ "$CLAWPATCH_REVISION" =~ ^[0-9a-f]{40}$ ]] || { printf 'build-clawpatch: third_party/clawpatch is not a registered submodule\n' >&2; exit 1; }
[[ -n "$CLAWPATCH_VERSION" && -n "$PNPM_VERSION" ]] || { printf 'build-clawpatch: toolchain/versions.env lacks CLAWPATCH_VERSION or PNPM_VERSION\n' >&2; exit 1; }

source_input="${CLAWPATCH_SOURCE:-$repo_root/third_party/clawpatch}"
output_input="${CLAWPATCH_OUTPUT_DIR:-${SCOUT_TOOLS_DIR:-$repo_root/.tools}/clawpatch}"

usage() {
  cat >&2 <<'USAGE'
Usage: scripts/build-clawpatch.sh [--source CHECKOUT] [--output DIRECTORY]

Builds the pinned Clawpatch CLI from a local Git checkout. The resulting absolute
CLI path is written to stdout. CLAWPATCH_SOURCE and CLAWPATCH_OUTPUT_DIR provide
the same configuration without command-line arguments.
USAGE
}

die() {
  printf 'build-clawpatch: %s\n' "$*" >&2
  exit 1
}

absolute_directory() {
  (cd "$1" && pwd -P)
}

while (($#)); do
  case "$1" in
    --source)
      (($# >= 2)) || die "--source requires a checkout path"
      source_input="$2"
      shift 2
      ;;
    --output)
      (($# >= 2)) || die "--output requires a directory path"
      output_input="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      usage
      die "unknown argument: $1"
      ;;
  esac
done

for prerequisite in git node pnpm tar shasum cut; do
  command -v "$prerequisite" >/dev/null 2>&1 || die "required command is unavailable: $prerequisite"
done

source_dir="$(absolute_directory "$source_input")" || die "source checkout is unavailable: $source_input"
mkdir -p "$output_input" || die "cannot create output directory: $output_input"
output_root="$(absolute_directory "$output_input")" || die "output directory is unavailable: $output_input"

case "$output_root/" in
  "$source_dir/"*) die "output directory must be outside the Clawpatch source checkout" ;;
esac

source_head="$(git -C "$source_dir" rev-parse --verify HEAD 2>/dev/null)" || die "source is not a Git checkout"
[[ "$source_head" == "$CLAWPATCH_REVISION" ]] || die "source HEAD is $source_head, expected $CLAWPATCH_REVISION"
git -C "$source_dir" cat-file -e "$CLAWPATCH_REVISION^{commit}" 2>/dev/null || die "pinned source revision is unavailable"

node_version="$(node -p 'process.versions.node')" || die "cannot determine Node version"
node_major="${node_version%%.*}"
[[ "$node_major" =~ ^[0-9]+$ ]] && ((node_major >= 22)) || die "Node >=22 is required; found $node_version"

installed_pnpm_version="$(pnpm --version)" || die "cannot determine pnpm version"
[[ "$installed_pnpm_version" == "$PNPM_VERSION" ]] || die "pnpm $PNPM_VERSION is required; found $installed_pnpm_version"

[[ -f "$safety_patch" ]] || die "review safety patch is unavailable: $safety_patch"
actual_patch_sha256="$(shasum -a 256 "$safety_patch" | cut -d ' ' -f1)" || die "cannot checksum review safety patch"
[[ "$actual_patch_sha256" == "$SCOUT_REVIEW_SAFETY_PATCH_SHA256" ]] || die "review safety patch checksum mismatch"

stage_dir="$(mktemp -d "$output_root/.clawpatch-build.XXXXXX")" || die "cannot create build staging directory"
cleanup() {
  rm -rf "$stage_dir"
}
trap cleanup EXIT HUP INT TERM

build_dir="$stage_dir/clawpatch"
git -C "$source_dir" archive --format=tar --prefix=clawpatch/ "$CLAWPATCH_REVISION" | tar -xf - -C "$stage_dir"
[[ -f "$build_dir/package.json" && -f "$build_dir/pnpm-lock.yaml" ]] || die "pinned source archive is missing package metadata"
# The staging directory sits inside Scout's worktree; without its own repository
# git apply resolves paths against the enclosing repository and silently applies nothing.
git init -q "$build_dir" || die "cannot initialise build repository"
git -C "$build_dir" apply --check "$safety_patch" || die "review safety patch does not apply to pinned source"
git -C "$build_dir" apply "$safety_patch" || die "cannot apply review safety patch"
git -C "$build_dir" apply --check --reverse "$safety_patch" || die "review safety patch was not applied to the build source"
rm -rf "$build_dir/.git"

node - "$build_dir/package.json" "$CLAWPATCH_VERSION" "$PNPM_VERSION" <<'NODE'
const fs = require("node:fs");
const [manifestPath, expectedVersion, expectedPnpm] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
if (
  manifest.name !== "clawpatch" ||
  manifest.version !== expectedVersion ||
  manifest.bin?.clawpatch !== "dist/cli.js" ||
  (manifest.packageManager !== `pnpm@${expectedPnpm}` && !String(manifest.packageManager).startsWith(`pnpm@${expectedPnpm}+`)) ||
  manifest.engines?.node !== ">=22"
) {
  process.exitCode = 1;
}
NODE
[[ $? -eq 0 ]] || die "pinned source package identity does not match Clawpatch $CLAWPATCH_VERSION"

pnpm --dir "$build_dir" install --frozen-lockfile
pnpm --dir "$build_dir" run build

built_cli="$build_dir/dist/cli.js"
[[ -f "$built_cli" ]] || die "build did not produce dist/cli.js"
chmod 0755 "$built_cli"
built_version="$("$built_cli" --version)" || die "built CLI failed its identity check"
[[ "$built_version" == "$CLAWPATCH_VERSION" ]] || die "built CLI reports $built_version, expected $CLAWPATCH_VERSION"
printf '%s %s\n' "$SCOUT_REVIEW_SAFETY_PATCH_VERSION" "$SCOUT_REVIEW_SAFETY_PATCH_SHA256" > "$build_dir/.scout-review-safety"

final_dir="$output_root/$CLAWPATCH_REVISION"
if [[ -e "$final_dir" || -L "$final_dir" ]]; then
  rm -rf "$final_dir"
fi
mv "$build_dir" "$final_dir"
final_cli="$final_dir/dist/cli.js"
[[ -x "$final_cli" ]] || die "built CLI is not executable after publication"

trap - EXIT HUP INT TERM
rm -rf "$stage_dir"
printf '%s\n' "$final_cli"
