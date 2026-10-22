# E1b-v2 — Response to council review round 9

I fixed all three round-9 nits and have no rebuttals. No real Jev or annotator calls were made, no model was fitted to real data, and there is no `validation/` directory.

| # | Issue | Fix (file:line) |
|---|---|---|
| 1 | Dead conditional in `constructs.py:58` | Now `Q_within_sd=within_sd(g, "Q")`. Re-ran `constructs.py panel-final.csv.gz`; the output is **byte-identical** to `constructs.json` (`cmp`; output in `/tmp/e1bv2-constructs-r10.json`). This step computes distributions only. |
| 2 | Comment at `validate.py:89` said "panel PRs only" | Now reads "dev classes from panel PRs; `nondev` from all typed PRs, since nondev PRs never enter the panel". That matches `:92` and the `nondev` exclusion in `build_panel.py:106-109` and prereg `:45`. |
| 3 | Comment at `tests.py:331` said "python 9/10 no" | Now reads "python 8/10 no (D000, D002 yes), then D004 unparsable twice -> 7 no / 2 yes / 1 unresolved; go 7/10 no; explained: 8/10 yes". This matches the fixture at `:335,339`. |

## Verification (main agent, 2026-10-21)

- `tests.py`: **64 checks pass** (log in `/tmp/e1bv2-tests-r10.log`).
- `constructs.json` reproduces byte-for-byte, so its digest is unchanged.
- Harness and power were not re-run. `analysis.py`, `power.py` and `harness.py` are byte-identical to round 8.
- `DESIGN-NOTES.md` §10: recomputed all 48 digests. Since round 8/9 only `constructs.py`, `tests.py` and `validate.py` changed. Copy: `/tmp/e1bv2-digests-round10.txt`.
