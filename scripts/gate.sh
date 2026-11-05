#!/usr/bin/env bash
set -euo pipefail
root="$(git rev-parse --show-toplevel)"
cd "$root"
source .tools/activate
exec .tools/venv/bin/python scripts/guardrails/gate.py "$@"
