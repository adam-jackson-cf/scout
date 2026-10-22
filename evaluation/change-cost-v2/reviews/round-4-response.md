# E1b-v2 — Response to council review round 4

All four round-4 issues are fixed; no rebuttals. No real Jev or annotator calls, no model fitted to real data, no `validation/` directory. Fixes made and verified by the main agent.

| # | Issue | Fix (file:line) |
|---|---|---|
| 1 (minor) | S1/S2/P1 jackknife decided cluster presence from `CONTROLS+["Q"]`, not each statistic's own columns | `analysis.py:227-239`: `EXPLORATORY_COLS` (S1 `y_review`, S2 `y_rework`, P1 `Q`+`D`+`y_active`) and `stat_cols()`, which maps `P1_*` to the P1 set and raises on an unknown statistic. `analyse()` uses `stat_cols(nme)` for presence, G and df (`:268`). `estimate()` (`:167`) and `predictive_gain()` (`:179`) select rows with the same sets, so the jackknife and the estimators cannot diverge. Tests: `tests.py:243-255` (one repository's `y_active` removed: P1 df = its own clusters − 1 and is below T5's; `stat_cols` mapping and the unknown-name error). |
| 2 (minor) | `logit` returned after exhausting IRLS iterations without a convergence check | `analysis.py:104-114`: returns only when max \|step\| < 1e-8, otherwise raises `RuntimeError`. Test: `tests.py:256-263` (`iters=1` raises; the default budget converges). |
| 3 (nit) | DESIGN-NOTES l.12 said 55 tests | Now 62; §8 lists the new tests. |
| 4 (nit) | `validate.py` docstring said `RUBRIC.txt` and described G7 at pair level | `validate.py:5` `RUBRIC.json`; `:15-18` states pair labelling, partner aggregation, partner-level metrics, pair-level kappa and the zero-decided-language failure, matching `gates()` (`:219-246`). |

## Verification (main agent, 2026-10-20/21)

- `tests.py`: **62 checks pass** (59 + 3 new; the council expected 61 — the extra check covers `stat_cols`).
- Harness re-run (200 sims/scenario) against the current `analysis.py`: output identical to `harness-results.json` apart from timing fields; every scenario's checks true. The convergence check never triggered.
- Power re-run (`power.py panel-final.csv.gz 60`, ~107 min): output identical to `power-results.json` apart from timing fields. The convergence check never triggered on the real design.
- `DESIGN-NOTES.md` §10: 48 digests recomputed; changed since round 4: `analysis.py`, `tests.py`, `validate.py` only. Copy for the reviewer: `/tmp/e1bv2-digests-round5.txt`.
