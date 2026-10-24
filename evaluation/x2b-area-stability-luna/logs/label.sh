#!/usr/bin/env bash
cd /Users/adamjackson/Projects/inflight/scout-x2b || exit 1
unset SCOUT_TOOLS_ACTIVE
source .tools/activate
echo "x2b-label started $(date -u +%FT%TZ)"
nohup evaluation/x2b-area-stability-luna/run.sh label > /tmp/x2b-run/label.log 2>&1
rc=$?
echo "x2b-label finished rc=$rc $(date -u +%FT%TZ)"
exit $rc
