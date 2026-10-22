# E1b-v2 — Response to council review round 5

All four round-5 nits are fixed; no rebuttals. Every change is text (docstrings, comments, notes); no executable line changed. No real Jev or annotator calls, no model fitted to real data, no `validation/` directory.

| # | Issue | Fix (file:line) |
|---|---|---|
| 1 | DESIGN-NOTES null coverage "0.935–0.980"; actual minimum 0.930 | `DESIGN-NOTES.md:87` now "0.930–0.980". Verified from `harness-results.json`: min 0.930 (`null.T2`), max 0.980 (`permutation.T2`) over the 20 null-scenario test cells. |
| 2 | `power.py` docstring stated the superseded adequacy rule | `power.py:9-16`: analytic MDE80 (normal approximation, not an 80 %-power effect), verification at the analytic MDE80, empirical MDE80 grid {1, 1.25, 1.5} × 200 sims, and adequacy = upper MC bound of the empirical MDE80 ≤ SESOI (prereg §7). |
| 3 | `analysis.py` T5 estimand omitted the absent-partners control; "bootstrap replicates" | `analysis.py:12` adds `log1p(absent partners)_std`, matching `estimate()` (`:160-163`); `:25` and `:98` say "jackknife replicates"; `:98` notes that non-convergence raises. |
| 4 | Prereg "DRAFT … NOT RUN" becomes false after sign-off | `PREREGISTRATION.md:3`: status now points to `reviews/` (sign-off and digests) and `RESULTS.md` (run), and states that the document does not change after sign-off. |

## Verification (main agent, 2026-10-21)

- `tests.py`: **62 checks pass**.
- Harness and power not re-run: the `analysis.py` and `power.py` changes are docstring/comment-only (lines 12, 25, 98 and 9-16), so executed code is identical to round 5, where both reproduced their stored outputs.
- `DESIGN-NOTES.md` §10: 48 digests recomputed; changed since round 5: `PREREGISTRATION.md`, `analysis.py`, `power.py` only. Copy: `/tmp/e1bv2-digests-round6.txt`.
