#!/usr/bin/env bash
# X2 offline stages. Run from anywhere; needs the pinned toolchain (`source .tools/activate`). No stage here calls a model except `label`.
#   run.sh validate      synthetic planted-effect / planted-null checks (no data)
#   run.sh replay        write transitions, replay areas.Reconcile for every grid threshold over data/maps (also data/chain.jsonl, data/replay-meta.json)
#   run.sh pairs         build blinded labelling units (real and control pairs) from the replay
#   run.sh label         agent labelling (claude -p + codex exec + adjudicator); appends to data/pairs/labels.jsonl
#   run.sh analysis      join, decide, write RESULTS inputs (data/results.json, data/analysis_output.txt)
# `replay` and `pairs` refuse once data/pairs/labels.jsonl has rows (labels are bound to the pair ids). X2_DATA overrides the data directory (validation).
set -euo pipefail
cd "$(dirname "$0")/../.."
X2=evaluation/x2-area-stability
PY=.tools/venv/bin/python
D=${X2_DATA:-$X2/data}
refuse_if_labelled() {
  if [ -s "$D/pairs/labels.jsonl" ]; then
    echo "refusing: $D/pairs/labels.jsonl has rows; replay and pairs outputs are bound to the labels (PREREGISTRATION section 6)" >&2
    exit 1
  fi
}
case "${1:-}" in
  validate) exec "$PY" "$X2/validation/validate.py" ;;
  replay)
    refuse_if_labelled
    GRID=$("$PY" -c "import sys; sys.path.insert(0, '$X2'); import common; print(','.join(map(str, common.GRID_PCT)))")
    "$PY" -c "import sys; sys.path.insert(0, '$X2'); import common; common.write_transitions('$D/transitions.json')"
    go run "./$X2/replay" -maps "$D/maps" -transitions "$D/transitions.json" -grid "$GRID" -out "$D/replay.jsonl" \
      -chain-out "$D/chain.jsonl" -meta-out "$D/replay-meta.json" -root . ;;
  pairs) refuse_if_labelled; "$PY" "$X2/build_pairs.py" --data "$D" --replay "$D/replay.jsonl" --out "$D/pairs" ;;
  label) "$PY" "$X2/label.py" --pairs "$D/pairs" "${@:2}" ;;
  analysis) "$PY" "$X2/analysis.py" --data "$D" --replay "$D/replay.jsonl" --pairs "$D/pairs" --out "$D/results.json" --text "$D/analysis_output.txt" ;;
  *) sed -n '2,9p' "$0"; exit 2 ;;
esac
