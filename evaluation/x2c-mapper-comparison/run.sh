#!/usr/bin/env bash
# X2c stages. Run from anywhere; needs the pinned toolchain (`source .tools/activate`). Only `collect`, `label` and `jev` call models.
# The replay and labelling code is the FROZEN X2 code, run unchanged on X2c's data directory (every stage first verifies the X2 digests).
#   run.sh validate                       X2's own validation (reused code), then X2c's checks (multi-arm collector, both methods, decision rule); no data, no model call
#   run.sh collect SRC_FASTAPI SRC_AXUM   the 8 (arm, repo) collections in x2c.COLLECTION_ORDER, sequentially (SRC_*: a clone containing the pinned revision)
#   run.sh replay                         transitions for the four arms, replay areas.Reconcile at every grid threshold over data/maps (Jaccard method)
#   run.sh pairs                          build the candidate pairs (every prior-area x next-feature pair sharing >= 1 owned file) + controls + jev-input.jsonl
#   run.sh label [--workers N]            agent labelling of the candidate set (claude -p + codex exec + adjudicator); appends to data/pairs/labels.jsonl
#   run.sh jev                            score the candidate pairs with Jev (needs the Jev key: run as `infisical run --projectId ... --env dev --silent -- run.sh jev`)
# `replay` and `pairs` refuse once data/pairs/labels.jsonl has rows. `jev` never reads labels; `analysis` needs both. X2C_DATA overrides the data directory (every stage, `collect`
# included, reads and writes it) and X2C_JEV_ENDPOINT the Jev route (validation points it at a local stub; the real run leaves it unset).
set -euo pipefail
cd "$(dirname "$0")/../.."
X2=evaluation/x2-area-stability
X2C=evaluation/x2c-mapper-comparison
PY=.tools/venv/bin/python
D=${X2C_DATA:-$X2C/data}
verify() { "$PY" -c "import sys; sys.path.insert(0, '$X2C'); import x2c; bad = x2c.verify_pins() + x2c.check_jev_pin(); sys.exit('pin check failed: ' + '; '.join(bad)) if bad else 0"; }
refuse_if_labelled() {
  if [ -s "$D/pairs/labels.jsonl" ]; then
    echo "refusing: $D/pairs/labels.jsonl has rows; replay and pairs outputs are bound to the labels (X2 PREREGISTRATION section 6)" >&2
    exit 1
  fi
}
case "${1:-}" in
  validate) exec "$PY" "$X2C/validation/validate.py" ;;
  collect)
    verify
    [ $# -eq 3 ] || { echo "usage: run.sh collect SRC_FASTAPI SRC_AXUM" >&2; exit 2; }
    "$PY" -c "
import sys, subprocess
sys.path.insert(0, '$X2C'); import x2c
src = {'fastapi': sys.argv[1], 'axum': sys.argv[2]}
for arm, repo in x2c.COLLECTION_ORDER:
    r = subprocess.run(['$PY', '$X2C/collect_x2c.py', '--arm', arm, '--repo', repo, '--source', src[repo], '--data', sys.argv[3]])
    if r.returncode != 0:
        sys.exit('collection %s/%s failed (rc %d); stopping: later collections are not started' % (arm, repo, r.returncode))
" "$2" "$3" "$D" ;;
  replay)
    verify; refuse_if_labelled
    GRID=$("$PY" -c "import sys; sys.path.insert(0, '$X2C'); import x2c, common; print(','.join(map(str, common.GRID_PCT)))")
    mkdir -p "$D"
    "$PY" -c "import sys; sys.path.insert(0, '$X2C'); import x2c, common; common.write_transitions('$D/transitions.json', arms=list(x2c.ARMS))"
    go run "./$X2/replay" -maps "$D/maps" -transitions "$D/transitions.json" -grid "$GRID" -out "$D/replay.jsonl" \
      -chain-out "$D/chain.jsonl" -meta-out "$D/replay-meta.json" -root . ;;
  pairs) verify; refuse_if_labelled; "$PY" "$X2C/build_candidates.py" --data "$D" --replay "$D/replay.jsonl" --out "$D/pairs" ;;
  label) verify; "$PY" "$X2/label.py" --pairs "$D/pairs" "${@:2}" ;;
  jev)
    verify
    mkdir -p "$D/jev"
    "$PY" -c "import sys, json, time; sys.path.insert(0, '$X2C'); import x2c; open('$D/jev/code-digests.jsonl', 'a').write(json.dumps({'utc': time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime()), 'digests': x2c.go_code_digests()}) + '\n')"
    go run "./$X2C/jev" -in "$D/pairs/jev-input.jsonl" -out "$D/jev/scores.jsonl" -questions-dir "$D/jev/questions" -workers 8 ${X2C_JEV_ENDPOINT:+-endpoint "$X2C_JEV_ENDPOINT"} "${@:2}" ;;
  analysis)
    "$PY" "$X2C/analysis_x2c.py" --data "$D" --replay "$D/replay.jsonl" --pairs "$D/pairs" --jev "$D/jev/scores.jsonl" --out "$D/results.json" --text "$D/analysis_output.txt" ;;
  *) sed -n '2,11p' "$X2C/run.sh"; exit 2 ;;
esac
