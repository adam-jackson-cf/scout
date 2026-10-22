#!/usr/bin/env bash
# Packages or restores Scout's state without touching the assessed source tree.
set -euo pipefail

usage() {
  cat >&2 <<'USAGE'
Usage:
  scripts/bundle-state.sh create --state-dir REPOSITORY/.scout --output OUTPUT.tar.gz
  scripts/bundle-state.sh extract --state-dir REPOSITORY/.scout --input INPUT.tar.gz

The create command writes OUTPUT.tar.gz and OUTPUT.tar.gz.sha256. The extract
command verifies the sidecar checksum before restoring a bundle. State paths
must be named .scout; no target source path is accepted.
USAGE
}

die() {
  printf 'bundle-state: %s\n' "$*" >&2
  exit 1
}

require_state_directory() {
  local state_dir="$1"
  [[ "$(basename "$state_dir")" == ".scout" ]] || die "state directory must be named .scout: $state_dir"
}

require_external_output() {
  local state_dir="$1"
  local output="$2"
  local state_root
  local output_parent
  state_root="$(cd "$(dirname "$state_dir")" && pwd -P)" || die "cannot resolve state root: $state_dir"
  output_parent="$(cd "$(dirname "$output")" && pwd -P)" || die "output parent is unavailable: $(dirname "$output")"
  case "$output_parent/" in
    "$state_root/"*) die "bundle output must be outside the assessed repository: $output" ;;
  esac
}

checksum_file() {
  printf '%s  %s\n' "$(shasum -a 256 "$1" | cut -d ' ' -f1)" "$(basename "$1")" > "$1.sha256"
}

verify_checksum() {
  local input="$1"
  local checksum="$input.sha256"
  [[ -f "$checksum" ]] || die "missing checksum sidecar: $checksum"

  local expected _ actual
  read -r expected _ < "$checksum" || die "cannot read checksum sidecar: $checksum"
  [[ "$expected" =~ ^[a-fA-F0-9]{64}$ ]] || die "invalid SHA-256 in checksum sidecar: $checksum"
  actual="$(shasum -a 256 "$input" | cut -d ' ' -f1)" || die "cannot checksum bundle: $input"
  [[ "$actual" == "$expected" ]] || die "bundle checksum mismatch: $input"
}

validate_archive_paths() {
  local archive="$1"
  local entry
  while IFS= read -r entry; do
    case "$entry" in
      .scout|.scout/*|./.scout|./.scout/*) ;;
      *) die "bundle contains a path outside .scout: $entry" ;;
    esac
    case "$entry" in
      ../*|*/../*|*/..) die "bundle contains an unsafe path: $entry" ;;
    esac
  done < <(tar -tzf "$archive")
  local listing
  while IFS= read -r listing; do
    case "${listing:0:1}" in
      -|d) ;;
      *) die "bundle contains a link or special file" ;;
    esac
  done < <(tar -tvzf "$archive")
}

for prerequisite in tar shasum find cut; do
  command -v "$prerequisite" >/dev/null 2>&1 || die "required command is unavailable: $prerequisite"
done

mode="${1:-}"
shift || true
state_dir=""
input=""
output=""

while (($#)); do
  case "$1" in
    --state-dir)
      (($# >= 2)) || die "--state-dir requires a path"
      state_dir="$2"
      shift 2
      ;;
    --input)
      (($# >= 2)) || die "--input requires a path"
      input="$2"
      shift 2
      ;;
    --output)
      (($# >= 2)) || die "--output requires a path"
      output="$2"
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

case "$mode" in
  create)
    [[ -n "$state_dir" && -n "$output" && -z "$input" ]] || die "create requires --state-dir and --output"
    require_state_directory "$state_dir"
    [[ -d "$state_dir" ]] || die "state directory is unavailable: $state_dir"
    [[ ! -e "$output" ]] || die "refusing to replace existing bundle: $output"
    [[ ! -e "$output.sha256" ]] || die "refusing to replace existing checksum: $output.sha256"
    [[ -z "$(find "$state_dir" -type l -print -quit)" ]] || die "state directory contains symbolic links"
    require_external_output "$state_dir" "$output"
    tar -czf "$output" -C "$(dirname "$state_dir")" ".scout"
    checksum_file "$output"
    ;;
  extract)
    [[ -n "$state_dir" && -n "$input" && -z "$output" ]] || die "extract requires --state-dir and --input"
    require_state_directory "$state_dir"
    [[ -f "$input" ]] || die "bundle is unavailable: $input"
    if [[ -e "$state_dir" ]] && [[ -n "$(find "$state_dir" -mindepth 1 -print -quit)" ]]; then
      die "refusing to merge bundle into non-empty state directory: $state_dir"
    fi
    verify_checksum "$input"
    validate_archive_paths "$input"
    mkdir -p "$(dirname "$state_dir")"
    tar -xzf "$input" -C "$(dirname "$state_dir")"
    ;;
  *)
    usage
    die "expected create or extract"
    ;;
esac
