# data-rev8-unpatched — excluded, preserved untouched

The first X2 collection (preregistration rev 8, started 2026-10-23T17:24Z) ran on the unpatched local Clawpatch build
(`clawpatch_cli_sha256 4eab263b…`, no review-safety patch; PREREGISTRATION.md §11 Amendment 1). It holds the 7 codex/fastapi maps (all rc 0)
and the codex/fastapi `tool-versions` file, plus `x2-collect.log` (the log of that run, including the claude/fastapi refusal at the model probe).

**These maps are excluded from every analysis.** No stage reads this directory (`run.sh` stages use `data/`). Do not edit or delete its files.
No threshold–survival/spurious or other predictor–outcome quantity was computed from them.
