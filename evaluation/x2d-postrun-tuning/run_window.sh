#!/usr/bin/env bash
# Post-run window pipeline for refinement rounds 1-7 (units -> wide labels -> gold), run from the X2d run worktree.
# Calls the pilot's modules directly with the pilot's own arguments; never edits frozen files; never touches rounds 8-10.
# Resumable: a stage whose output exists is skipped. Usage: run_window.sh [FIRST [LAST]]   (default 1 7)
set -euo pipefail
cd "${X2D_WORKTREE:-/Users/adamjackson/Projects/inflight/scout-x2d}"
S=evaluation/x2d-merge-association-pilot
export X2D_DATA=$PWD/$S/data/postrun
PY=.tools/venv/bin/python
SRC=${X2D_SRC:-/tmp/scout-x2c-smoke/_src/fastapi}
R=$X2D_DATA/rounds
for r in $(seq "${1:-1}" "${2:-7}"); do
  [ -s "$R/$r/version.json" ] || $PY $S/versioning.py --round "$r"
  [ -s "$R/$r/units.jsonl" ] || $PY $S/build_units.py --repo "$SRC" --round "$r"
  if [ ! -s "$R/$r/labels.jsonl" ]; then $PY $S/label_x2d.py --round "$r" --workers 6; fi
  [ -s "$R/$r/gold.jsonl" ] || $PY $S/gold.py --round "$r"
  echo "round $r done"
done
