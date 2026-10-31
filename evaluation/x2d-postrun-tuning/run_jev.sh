#!/usr/bin/env bash
# V-sem Jev scoring (merge.covered + merge.which_area, pinned model, v1 questions) of the window units of rounds 1-7, one scores file per round,
# under a TOTAL budget across all rounds (default 1.0 USD). Run from anywhere; needs the key injected:
#   infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- evaluation/x2d-postrun-tuning/run_jev.sh
# Resumable and append-only (the driver skips answered requests). Inputs/settings are each round's own frozen copies in the postrun data root.
set -euo pipefail
cd "${X2D_WORKTREE:-/Users/adamjackson/Projects/inflight/scout-x2d}"
S=evaluation/x2d-merge-association-pilot
export X2D_DATA=$PWD/$S/data/postrun
PY=.tools/venv/bin/python
BUDGET=${X2D_POSTRUN_BUDGET:-1.0}
R=$X2D_DATA/rounds
for r in 1 2 3 4 5 6 7; do
  [ -s "$R/$r/units.window.jsonl" ] || $PY -c "
import json, sys
rows = [json.loads(l) for l in open('$R/$r/units.jsonl')]
with open('$R/$r/units.window.jsonl', 'w') as f:
    for u in rows:
        if u.get('set') == 'window' and not u.get('control'):
            f.write(json.dumps(u, sort_keys=True) + '\n')
"
  left=$($PY -c "
import glob, json
spent = 0.0
for p in glob.glob('$R/*/jev/scores.jsonl'):
    for l in open(p):
        d = json.loads(l)
        if d.get('cost_known') and d.get('cost_usd'): spent += d['cost_usd']
print(round($BUDGET - spent, 6))
")
  mkdir -p "$R/$r/jev"
  go run "./$S/jev" -in "$R/$r/units.window.jsonl" -catalogue "$X2D_DATA/catalogue/catalogue.json" -out "$R/$r/jev/scores.jsonl" -arms sem \
    -questions-dir "$R/$r/jev/questions" -questions-src "$R/$r/inputs/questions" -settings "$R/$r/settings.json" -workers 8 -max-cost-usd "$left"
  echo "round $r jev done (budget left before it: $left)"
done
