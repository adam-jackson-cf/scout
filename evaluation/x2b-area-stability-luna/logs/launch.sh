#!/usr/bin/env bash
# X2b collection launcher (README "Real run"): codex/fastapi then codex/axum, sequential, nohup-protected.
cd /Users/adamjackson/Projects/inflight/scout-x2b || exit 1
unset SCOUT_TOOLS_ACTIVE
source .tools/activate
SOURCE_FASTAPI=/tmp/scout-p1/fastapi
SOURCE_AXUM=/tmp/scout-p4/axum-src
PY=.tools/venv/bin/python; C=evaluation/x2b-area-stability-luna/collect_x2b.py
S=/tmp/x2b-run/status-before-second.txt
echo "x2b-collect started $(date -u +%FT%TZ)"
nohup sh -c "$PY $C --arm codex --repo fastapi --source $SOURCE_FASTAPI &&
             { git status --porcelain internal/ scripts/ evaluation/x2b-area-stability-luna/ > $S; } &&
             $PY $C --arm codex --repo axum    --source $SOURCE_AXUM" > x2b-collect.log 2>&1
rc=$?
echo "x2b-collect finished rc=$rc $(date -u +%FT%TZ)"
exit $rc
