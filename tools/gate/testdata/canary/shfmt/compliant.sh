#!/usr/bin/env bash
set -euo pipefail
if [[ -n "${1:-}" ]]; then
  printf '%s\n' "$1"
fi
