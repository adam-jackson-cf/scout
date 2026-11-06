#!/usr/bin/env bash
set -euo pipefail
case "${1:-ci}" in
  commit|ci|path) ;;
  *) echo "unknown mode: $1" >&2; exit 2 ;;
esac
root="$(git rev-parse --show-toplevel)"
cd "$root"
source .tools/activate
exec .tools/venv/bin/python scripts/guardrails/gate.py "$@"
