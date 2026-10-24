#!/usr/bin/env bash
# Installs Scout's pinned toolchain into a project-local directory (default: .tools/), venv-style.
#
#   scripts/bootstrap-tools.sh            install or repair; idempotent
#   scripts/bootstrap-tools.sh --relock   regenerate toolchain/python-requirements.lock and
#                                         toolchain/node/package-lock.json, then install
#   source .tools/activate                use the toolchain in this shell; `scout_deactivate` undoes it
#
# Pins: toolchain/versions.env; download checksums: toolchain/checksums.sha256; Clawpatch: the
# third_party/clawpatch submodule gitlink. Needs only bash, curl, tar, git, shasum|sha256sum and,
# for building Scout itself, a Go toolchain (GOTOOLCHAIN then fetches the pinned Go).
# System prerequisites that cannot live in .tools/ are checked and reported, never installed:
# the scanner sandbox (sandbox-exec on macOS, bubblewrap on Linux) and the account-bound CLIs.
set -euo pipefail

die() { printf 'bootstrap-tools: error: %s\n' "$*" >&2; exit 1; }
log() { printf 'bootstrap-tools: %s\n' "$*" >&2; }

repo="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
toolchain="$repo/toolchain"
tools="${SCOUT_TOOLS_DIR:-$repo/.tools}"

relock=false
case "${1:-}" in
  "") ;;
  --relock) relock=true ;;
  *) die "usage: scripts/bootstrap-tools.sh [--relock]" ;;
esac

for prerequisite in curl tar git; do
  command -v "$prerequisite" >/dev/null 2>&1 || die "required command is unavailable: $prerequisite"
done
source "$repo/scripts/sha256.sh"

while IFS='=' read -r key value; do
  [[ -z "$key" || "$key" == \#* ]] && continue
  [[ "$key" =~ ^[A-Z][A-Z0-9_]*$ && "$value" =~ ^[A-Za-z0-9._:/@+-]+$ ]] || die "malformed line in toolchain/versions.env: $key"
  printf -v "$key" '%s' "$value"
done < "$toolchain/versions.env"
for key in GO_VERSION NODE_VERSION PNPM_VERSION PYTHON_VERSION UV_VERSION GITLEAKS_VERSION SEMGREP_VERSION RUFF_VERSION LIZARD_VERSION JSCPD_VERSION CLAWPATCH_VERSION; do
  [[ -n "${!key:-}" ]] || die "toolchain/versions.env does not pin $key"
done

case "$(uname -s)/$(uname -m)" in
  Darwin/arm64) node_platform=darwin-arm64 gitleaks_platform=darwin_arm64 uv_platform=aarch64-apple-darwin ;;
  Darwin/x86_64) node_platform=darwin-x64 gitleaks_platform=darwin_x64 uv_platform=x86_64-apple-darwin ;;
  Linux/x86_64) node_platform=linux-x64 gitleaks_platform=linux_x64 uv_platform=x86_64-unknown-linux-gnu ;;
  Linux/aarch64 | Linux/arm64) node_platform=linux-arm64 gitleaks_platform=linux_arm64 uv_platform=aarch64-unknown-linux-gnu ;;
  *) die "unsupported platform: $(uname -s)/$(uname -m)" ;;
esac

mkdir -p "$tools/downloads" "$tools/bin"

# fetch URL NAME: download once into .tools/downloads and verify against toolchain/checksums.sha256.
fetch() {
  local url="$1" name="$2" dest="$tools/downloads/$2" expected
  expected="$(awk -v n="$name" '$2 == n { print $1 }' "$toolchain/checksums.sha256")"
  [[ "$expected" =~ ^[0-9a-f]{64}$ ]] || die "no pinned checksum for $name in toolchain/checksums.sha256"
  if [[ ! -f "$dest" || "$(sha256 "$dest")" != "$expected" ]]; then
    log "downloading $name"
    curl -fsSL --retry 3 -o "$dest.part" "$url" || die "download failed: $url"
    mv "$dest.part" "$dest"
  fi
  [[ "$(sha256 "$dest")" == "$expected" ]] || { rm -f "$dest"; die "checksum mismatch for $name"; }
  printf '%s\n' "$dest"
}

# stamp_matches DIR DIGEST: true when DIR was installed from inputs with DIGEST.
stamp_matches() { [[ -f "$1/.scout-stamp" && "$(< "$1/.scout-stamp")" == "$2" ]]; }

install_uv() {
  if [[ -x "$tools/uv/uv" && "$("$tools/uv/uv" --version | awk '{ print $2 }')" == "$UV_VERSION" ]]; then return; fi
  local name="uv-$uv_platform.tar.gz" archive
  archive="$(fetch "https://github.com/astral-sh/uv/releases/download/$UV_VERSION/$name" "$name")"
  rm -rf "$tools/uv" && mkdir -p "$tools/uv"
  tar -xzf "$archive" -C "$tools/uv" --strip-components 1
  log "installed uv $UV_VERSION"
}

install_node() {
  if [[ -x "$tools/node/bin/node" && "$("$tools/node/bin/node" --version)" == "v$NODE_VERSION" ]]; then return; fi
  local name="node-v$NODE_VERSION-$node_platform.tar.gz" archive
  archive="$(fetch "https://nodejs.org/dist/v$NODE_VERSION/$name" "$name")"
  rm -rf "$tools/node" && mkdir -p "$tools/node"
  tar -xzf "$archive" -C "$tools/node" --strip-components 1
  log "installed node $NODE_VERSION"
}

install_gitleaks() {
  if [[ -x "$tools/gitleaks/gitleaks" && "$("$tools/gitleaks/gitleaks" version)" == "$GITLEAKS_VERSION" ]]; then return; fi
  local name="gitleaks_${GITLEAKS_VERSION}_$gitleaks_platform.tar.gz" archive
  archive="$(fetch "https://github.com/gitleaks/gitleaks/releases/download/v$GITLEAKS_VERSION/$name" "$name")"
  rm -rf "$tools/gitleaks" && mkdir -p "$tools/gitleaks"
  tar -xzf "$archive" -C "$tools/gitleaks" gitleaks
  log "installed gitleaks $GITLEAKS_VERSION"
}

install_node_tools() {
  local lock="$toolchain/node/package-lock.json" digest
  [[ -f "$lock" ]] || die "toolchain/node/package-lock.json is missing (run with --relock)"
  digest="$(sha256 "$lock")-$NODE_VERSION"
  if stamp_matches "$tools/node-tools" "$digest"; then return; fi
  rm -rf "$tools/node-tools" && mkdir -p "$tools/node-tools"
  cp "$toolchain/node/package.json" "$lock" "$tools/node-tools/"
  (cd "$tools/node-tools" && PATH="$tools/node/bin:$PATH" npm ci --ignore-scripts --no-audit --no-fund --loglevel=error)
  printf '%s' "$digest" > "$tools/node-tools/.scout-stamp"
  log "installed node tools (jscpd $JSCPD_VERSION, pnpm $PNPM_VERSION)"
}

install_python_tools() {
  local lock="$toolchain/python-requirements.lock" digest
  [[ -f "$lock" ]] || die "toolchain/python-requirements.lock is missing (run with --relock)"
  digest="$(sha256 "$lock")-$PYTHON_VERSION"
  if stamp_matches "$tools/venv" "$digest"; then return; fi
  rm -rf "$tools/venv"
  uv_env venv --quiet --python "$PYTHON_VERSION" "$tools/venv"
  uv_env pip sync --quiet --require-hashes --python "$tools/venv/bin/python" "$lock"
  printf '%s' "$digest" > "$tools/venv/.scout-stamp"
  log "installed python $PYTHON_VERSION venv (semgrep $SEMGREP_VERSION, ruff $RUFF_VERSION, lizard $LIZARD_VERSION)"
}

# uv with every cache, interpreter and config inside .tools/ and no system interpreter.
uv_env() {
  UV_PYTHON_INSTALL_DIR="$tools/python" UV_CACHE_DIR="$tools/cache/uv" UV_PYTHON_PREFERENCE=only-managed UV_NO_CONFIG=1 \
    "$tools/uv/uv" "$@"
}

clawpatch_revision() {
  local revision
  revision="$(git -C "$repo" ls-files -s third_party/clawpatch | awk '$1 == "160000" { print $2 }')"
  [[ "$revision" =~ ^[0-9a-f]{40}$ ]] || die "third_party/clawpatch is not a registered submodule"
  printf '%s\n' "$revision"
}

install_clawpatch() {
  local revision head cli expected_marker
  revision="$(clawpatch_revision)"
  if [[ ! -e "$repo/third_party/clawpatch/.git" ]]; then
    git -C "$repo" submodule update --init third_party/clawpatch
  fi
  head="$(git -C "$repo/third_party/clawpatch" rev-parse HEAD)"
  [[ "$head" == "$revision" ]] || die "third_party/clawpatch is checked out at $head but the recorded pin is $revision (run: git submodule update third_party/clawpatch; an upgrade needs a compatibility review)"
  cli="$tools/clawpatch/$revision/dist/cli.js"
  expected_marker="$(awk -F '"' '/^readonly SCOUT_REVIEW_SAFETY_PATCH_(VERSION|SHA256)=/ { v = v (v ? " " : "") $2 } END { print v }' "$repo/scripts/build-clawpatch.sh")"
  if [[ -x "$cli" && -f "$tools/clawpatch/$revision/.scout-review-safety" && "$(<"$tools/clawpatch/$revision/.scout-review-safety")" == "$expected_marker" && "$(PATH="$tools/bin:$PATH" "$cli" --version)" == "$CLAWPATCH_VERSION" ]]; then return; fi
  PATH="$tools/bin:$PATH" CLAWPATCH_SOURCE="$repo/third_party/clawpatch" CLAWPATCH_OUTPUT_DIR="$tools/clawpatch" \
    bash "$repo/scripts/build-clawpatch.sh" > /dev/null
  log "built clawpatch $CLAWPATCH_VERSION ($revision)"
}

link_tools() {
  local target name
  for target in ../uv/uv ../node/bin/node ../node/bin/npm ../node/bin/npx ../gitleaks/gitleaks \
    ../node-tools/node_modules/.bin/pnpm ../node-tools/node_modules/.bin/jscpd \
    ../venv/bin/semgrep ../venv/bin/pysemgrep ../venv/bin/ruff ../venv/bin/lizard; do
    name="${target##*/}"
    [[ -e "$tools/bin/$target" ]] || die "expected tool is missing after install: $target"
    ln -sfn "$target" "$tools/bin/$name"
  done
}

write_activate() {
  local revision cli
  revision="$(clawpatch_revision)"
  cli="$tools/clawpatch/$revision/dist/cli.js"
  cat > "$tools/activate" <<EOF
# Generated by scripts/bootstrap-tools.sh. Use: source .tools/activate; undo: scout_deactivate
if [ -n "\${SCOUT_TOOLS_ACTIVE:-}" ]; then scout_deactivate; fi
SCOUT_TOOLS_OLD_PATH="\$PATH"
SCOUT_TOOLS_OLD_CLAWPATCH="\${SCOUT_CLAWPATCH_BINARY-}"
SCOUT_TOOLS_OLD_GOTOOLCHAIN="\${GOTOOLCHAIN-}"
export SCOUT_TOOLS_ACTIVE=1
export SCOUT_TOOLS_DIR="$tools"
export PATH="$tools/bin:\$PATH"
export SCOUT_CLAWPATCH_BINARY="$cli"
export GOTOOLCHAIN="go$GO_VERSION"
scout_deactivate() {
  PATH="\$SCOUT_TOOLS_OLD_PATH"
  if [ -n "\$SCOUT_TOOLS_OLD_CLAWPATCH" ]; then export SCOUT_CLAWPATCH_BINARY="\$SCOUT_TOOLS_OLD_CLAWPATCH"; else unset SCOUT_CLAWPATCH_BINARY; fi
  if [ -n "\$SCOUT_TOOLS_OLD_GOTOOLCHAIN" ]; then export GOTOOLCHAIN="\$SCOUT_TOOLS_OLD_GOTOOLCHAIN"; else unset GOTOOLCHAIN; fi
  unset SCOUT_TOOLS_ACTIVE SCOUT_TOOLS_DIR SCOUT_TOOLS_OLD_PATH SCOUT_TOOLS_OLD_CLAWPATCH SCOUT_TOOLS_OLD_GOTOOLCHAIN
  unset -f scout_deactivate
}
EOF
}

relock() {
  install_uv
  install_node
  uv_env pip compile --quiet --universal --generate-hashes --python-version "${PYTHON_VERSION%.*}" \
    --custom-compile-command "scripts/bootstrap-tools.sh --relock" \
    "$toolchain/python-requirements.in" -o "$toolchain/python-requirements.lock"
  (cd "$toolchain/node" && PATH="$tools/node/bin:$PATH" npm install --package-lock-only --ignore-scripts --no-audit --no-fund --loglevel=error)
  log "relocked toolchain/python-requirements.lock and toolchain/node/package-lock.json"
}

report_system() {
  local sandbox
  case "$(uname -s)" in
    Darwin) sandbox="sandbox-exec" ;;
    *) sandbox="bwrap" ;;
  esac
  if command -v "$sandbox" >/dev/null 2>&1; then log "system: scanner sandbox $sandbox present"
  else log "WARNING: scanner sandbox $sandbox is missing; every scanner lane will be blocked (Linux: install bubblewrap)"; fi
  local cli
  for cli in claude codex infisical; do
    if command -v "$cli" >/dev/null 2>&1; then log "optional: $cli present ($("$cli" --version 2>/dev/null | head -n 1))"
    else log "optional: $cli not installed (model-CLI lanes / secret injection unavailable)"; fi
  done
}

if [[ "$relock" == true ]]; then relock; fi
install_uv
install_node
install_gitleaks
install_node_tools
install_python_tools
link_tools
install_clawpatch
write_activate
report_system
log "ready: source $tools/activate"
