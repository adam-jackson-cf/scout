#!/usr/bin/env bash
# The ONE run on rounds 8-10 (FROZEN-8-10.md). Order: units -> labels -> gold -> B2 baselines -> one Jev pass (<= 0.5 USD, Infisical) -> Terra escalation -> report.
# Run from the scout checkout. Nothing here is changed after the freeze; a failed step is resumed (append-only), never re-tuned.
set -euo pipefail
HERE=$(cd "$(dirname "$0")" && pwd)
WT=${X2D_WORKTREE:-/Users/adamjackson/Projects/inflight/scout-x2d}
S=$WT/evaluation/x2d-merge-association-pilot
export X2D_DATA=$S/data/postrun
PY=$WT/.tools/venv/bin/python
"$HERE/run_window.sh" 8 10
for r in 8 9 10; do "$PY" "$S/baselines.py" --round "$r" --repo "${X2D_SRC:-/tmp/scout-x2c-smoke/_src/fastapi}"; done
X2D_POSTRUN_BUDGET=0.5 infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- "$HERE/run_jev.sh" 8 10
"$PY" "$HERE/escalate.py" "$S" --prompt "$HERE/terra_prompt_v1.txt" --tag frozen-r8-10 --rounds 8-10 --workers 6
