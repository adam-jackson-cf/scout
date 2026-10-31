#!/usr/bin/env bash
# The ONE axum run (FROZEN-AXUM.md). Stages, in order (each refuses to redo finished output; a failed stage is resumed, never re-tuned):
#   onboard-collect   [model: claude/codex via Clawpatch] 4 scopes x 3 draws + 1 contrast draw at C0
#   onboard           [model: Jev via Infisical] consensus + Jev for contested files -> data catalogue (+ sha256)
#   window            [model: labellers] per round FIRST..LAST: versioning -> build_units -> label -> gold
#   baselines         model-free B0/B1/B2 per round
#   states            model-free Jev state dump for the window units (read by the Terra step)
#   jev               [model: Jev via Infisical] V-sem on the window units, budget X2D_AXUM_JEV_BUDGET (default 0.5 USD) across the rounds
#   escalate          [model: codex gpt-6.1-terra high] Terra for every cluster Jev leaves none/unresolved, then the report JSON
#   report            model-free tables -> axum/report.md
# Usage (scout checkout):  evaluation/x2d-postrun-tuning/axum/run_axum.sh STAGE [ARGS]
# The Jev key is injected by Infisical only:  infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- run_axum.sh jev
set -euo pipefail
A=$(cd "$(dirname "$0")" && pwd)
SCOUT=$(cd "$A/../../.." && pwd)
PILOT=$A/pilot
WT=${X2D_WORKTREE:-/Users/adamjackson/Projects/inflight/scout-x2d}
export X2D_DATA=${X2D_AXUM_DATA:-$WT/evaluation/x2d-merge-association-pilot/data/postrun-axum}
SRC=${X2D_SRC:-/tmp/scout-x2c-smoke/_src/axum}
PY=$WT/.tools/venv/bin/python
C0=978ae6335862b264b4368e01a52177d98c42b2d9
FIRST=${X2D_AXUM_FIRST:-1}
LAST=${X2D_AXUM_LAST:-5}
R=$X2D_DATA/rounds
# shellcheck disable=SC1091
source "$WT/.tools/activate"
cd "$SCOUT"
GOPILOT=./evaluation/x2d-postrun-tuning/axum/pilot

stage=${1:-}
case "$stage" in
  onboard-collect)
    "$PY" "$PILOT/collect_onboard.py" --source "$SRC" --data "$X2D_DATA" --scopes "$PILOT/prepass/scopes.json" --work /tmp/scout-x2d-axum ;;
  onboard)
    mkdir -p "$X2D_DATA/onboard-jev"
    go run "$GOPILOT/onboard" -draws "$X2D_DATA/onboard/draws" -scopes "$PILOT/prepass/scopes.json" -repo "$SRC" -rev "$C0" -out "$X2D_DATA/catalogue" -contrast "$X2D_DATA/onboard/contrast" \
      -jev-out "$X2D_DATA/onboard-jev/scores.jsonl" -questions-src "$PILOT/questions" -thresholds "$PILOT/thresholds.json" -workers 8 -max-cost-usd "${X2D_AXUM_ONBOARD_BUDGET:-0.25}" ;;
  window)
    for r in $(seq "$FIRST" "$LAST"); do
      [ -s "$R/$r/version.json" ] || "$PY" "$PILOT/versioning.py" --round "$r"
      [ -s "$R/$r/units.jsonl" ] || "$PY" "$PILOT/build_units.py" --repo "$SRC" --round "$r"
      if [ ! -s "$R/$r/labels.jsonl" ]; then "$PY" "$PILOT/label_x2d.py" --round "$r" --workers 6; fi
      [ -s "$R/$r/gold.jsonl" ] || "$PY" "$PILOT/gold.py" --round "$r"
      echo "round $r done"
    done ;;
  baselines)
    for r in $(seq "$FIRST" "$LAST"); do "$PY" "$PILOT/baselines.py" --round "$r" --repo "$SRC"; done ;;
  states)
    for r in $(seq "$FIRST" "$LAST"); do
      [ -s "$R/$r/units.window.jsonl" ] || "$PY" -c "
import json
rows = [json.loads(l) for l in open('$R/$r/units.jsonl')]
with open('$R/$r/units.window.jsonl', 'w') as f:
    for u in rows:
        if u.get('set') == 'window' and not u.get('control'):
            f.write(json.dumps(u, sort_keys=True) + '\n')
"
      [ -s "$R/$r/states.window.jsonl" ] || go run "$GOPILOT/jev" -dump-state -in "$R/$r/units.window.jsonl" -catalogue "$X2D_DATA/catalogue/catalogue.json" -arms sem -out "$R/$r/states.window.jsonl" \
        -questions-src "$R/$r/inputs/questions" -settings "$R/$r/settings.json"
    done ;;
  jev)
    BUDGET=${X2D_AXUM_JEV_BUDGET:-0.5}
    for r in $(seq "$FIRST" "$LAST"); do
      left=$("$PY" -c "
import json
spent = 0.0
for rr in range($FIRST, $LAST + 1):
    try:
        for l in open('$R/%d/jev/scores.jsonl' % rr):
            d = json.loads(l)
            if d.get('cost_known') and d.get('cost_usd'): spent += d['cost_usd']
    except FileNotFoundError:
        pass
print(round($BUDGET - spent, 6))
")
      mkdir -p "$R/$r/jev"
      go run "$GOPILOT/jev" -in "$R/$r/units.window.jsonl" -catalogue "$X2D_DATA/catalogue/catalogue.json" -out "$R/$r/jev/scores.jsonl" -arms sem \
        -questions-dir "$R/$r/jev/questions" -questions-src "$R/$r/inputs/questions" -settings "$R/$r/settings.json" -workers 8 -max-cost-usd "$left"
      echo "round $r jev done (budget left before it: $left)"
    done ;;
  escalate)
    "$PY" "$A/escalate.py" "$PILOT" --prompt "$A/../terra_prompt_v1.txt" --tag "axum-r$FIRST-$LAST" --rounds "$FIRST-$LAST" --workers 6 ;;
  report)
    "$PY" "$A/report_axum.py" "$PILOT" "$FIRST-$LAST" ;;
  *) sed -n '2,13p' "$0"; exit 2 ;;
esac
