# E1b-v2 — Response to council review round 6

All four round-6 issues fixed; no rebuttals. No real Jev or annotator calls, no model fitted to real data, no `validation/` directory.

| # | Issue | Fix (file:line) |
|---|---|---|
| 1 (minor) | `validate.kappa` counted identical `unparsable` answers as agreement; such items never reached C | Code now follows prereg §5. A single rule, `validate.disagree(a, b)`, is true when the labels differ or either is `unparsable` (`validate.py:163-165`), and every use goes through it. In `kappa`, an unparsable answer never agrees and `unparsable` is excluded from `pe` (`:168-175`). `consensus` returns A only when A and B agree (parsable), otherwise C, and an unparsable or missing C ⇒ `None` (`:183-187`). `annotate.py:73` sends every `disagree` item to C, including both-unparsable items (docstring `:8-9`; import `:22`). Prereg `:149-150` states this explicitly. Tests `tests.py:290-302`: a both-unparsable item resolves via C when C labels it and stays unresolved without C; `disagree` truth table; κ(["x","unparsable"], same) = 1/3 exactly; all-unparsable κ = 0. |
| 2 (nit) | `DESIGN-NOTES.md:31,74` coverage "0.92–0.97" | Now "0.93–0.98" (stored min 0.93 `null.T2`, max 0.98 `permutation.T2`). |
| 3 (nit) | `DESIGN-NOTES.md:77` said T3 rejects under the null | Now states the harness values (`null_area_confounded`: T1 0.96, T3 0.035, from `harness-results.json`). T1/T3 remain associational because both are level tests exposed to area-level confounding of Q. |
| 4 (nit) | `DESIGN-NOTES.md:102` incomplete list of T5 verdicts | Now lists every verdict `decide()` can give an inadequate T5: `supported`, `detected_below_SESOI`, `construct_invalid`, `inconclusive`, `not_decided`; never `equivalent` (`analysis.py:299-323`). |

## Verification (main agent, 2026-10-21)

- `tests.py`: **64 checks pass**. The consensus check was strengthened, and the `disagree` and `kappa` checks were added.
- Harness and power not re-run. `analysis.py`, `power.py` and `harness.py` are byte-identical to round 6 (unchanged digests), and none imports `validate`, `kappa`, `consensus` or `annotate`. `validate.py` changes only G4/G5/G7 computation, which runs after sign-off.
- `DESIGN-NOTES.md` §10: 48 digests recomputed. Changed since round 6: `PREREGISTRATION.md`, `annotate.py`, `tests.py`, `validate.py`. Copy: `/tmp/e1bv2-digests-round7.txt`.
