#!/usr/bin/env bash
# X2b offline stages. Run from anywhere; needs the pinned toolchain (`source .tools/activate`). No stage here calls a model except `label`.
# The replay, pair-building and labelling code is the FROZEN X2 code, run unchanged on X2b's data directory (every stage first verifies the X2 digests).
#   run.sh validate      X2's own validation (reused code), then X2b's checks (codex-only pipeline, collector, drift guards); no data
#   run.sh replay        write transitions (codex arm only), replay areas.Reconcile for every grid threshold over data/maps
#   run.sh pairs         build blinded labelling units from the replay
#   run.sh label         agent labelling (claude -p + codex exec + adjudicator); appends to data/pairs/labels.jsonl
#   run.sh analysis      join, decide, X2b served/flags check and decision combination (data/results.json, data/analysis_output.txt)
# `replay` and `pairs` refuse once data/pairs/labels.jsonl has rows. X2B_DATA overrides the data directory (validation).
set -euo pipefail
cd "$(dirname "$0")/../.."
X2=evaluation/x2-area-stability
X2B=evaluation/x2b-area-stability-luna
PY=.tools/venv/bin/python
D=${X2B_DATA:-$X2B/data}
verify() { "$PY" -c "import sys; sys.path.insert(0, '$X2B'); import x2b; bad = x2b.verify_pins(); sys.exit('frozen X2 files differ from their pins: ' + '; '.join(bad)) if bad else 0"; }
refuse_if_labelled() {
  if [ -s "$D/pairs/labels.jsonl" ]; then
    echo "refusing: $D/pairs/labels.jsonl has rows; replay and pairs outputs are bound to the labels (X2 PREREGISTRATION section 6)" >&2
    exit 1
  fi
}
case "${1:-}" in
  validate) exec "$PY" "$X2B/validation/validate.py" ;;
  replay)
    verify; refuse_if_labelled
    GRID=$("$PY" -c "import sys; sys.path.insert(0, '$X2B'); import x2b, common; print(','.join(map(str, common.GRID_PCT)))")
    mkdir -p "$D"
    "$PY" -c "import sys; sys.path.insert(0, '$X2B'); import x2b, common; common.write_transitions('$D/transitions.json', arms=list(x2b.ARMS))"
    go run "./$X2/replay" -maps "$D/maps" -transitions "$D/transitions.json" -grid "$GRID" -out "$D/replay.jsonl" \
      -chain-out "$D/chain.jsonl" -meta-out "$D/replay-meta.json" -root . ;;
  pairs) verify; refuse_if_labelled; "$PY" "$X2/build_pairs.py" --data "$D" --replay "$D/replay.jsonl" --out "$D/pairs" ;;
  label) verify; "$PY" "$X2/label.py" --pairs "$D/pairs" "${@:2}" ;;
  analysis) "$PY" "$X2B/analysis_x2b.py" --data "$D" --replay "$D/replay.jsonl" --pairs "$D/pairs" --out "$D/results.json" --text "$D/analysis_output.txt" ;;
  *) sed -n '2,11p' "$0"; exit 2 ;;
esac
