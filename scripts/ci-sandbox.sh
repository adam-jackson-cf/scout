#!/usr/bin/env bash
# Provision Linux scanner isolation. Guardian callers retain warning-only probe failure;
# repository verification uses --strict so an unusable sandbox fails before its tests.
set -euo pipefail

strict=false
case "${1:-}" in
  "") ;;
  --strict) strict=true ;;
  *) printf 'usage: scripts/ci-sandbox.sh [--strict]\n' >&2; exit 2 ;;
esac
[[ $# -le 1 ]] || { printf 'usage: scripts/ci-sandbox.sh [--strict]\n' >&2; exit 2; }

if [[ "$(uname -s)" == Linux ]]; then
  sudo apt-get update -qq
  sudo apt-get install -y bubblewrap
  # Ubuntu 24.04 blocks unprivileged user namespaces, which bubblewrap needs.
  # The key is absent on kernels without the AppArmor restriction; the probe decides.
  sudo sysctl -w kernel.apparmor_restrict_unprivileged_userns=0 > /dev/null 2>&1 || true
  if ! bwrap --unshare-user --ro-bind / / true; then
    if [[ "$strict" == true ]]; then
      echo "::error::bubblewrap cannot create user namespaces on this runner; repository gate requires the scanner sandbox" >&2
      exit 1
    fi
    echo "::warning::bubblewrap cannot create user namespaces on this runner; sandboxed scanner lanes will report blocked"
  fi
fi
