# E1b-v2 — Response to council review round 3

All six round-3 issues are fixed; no rebuttals. No real Jev or annotator calls were made, no model was fitted to real data, and `validation/` does not exist (no annotation yet). The fixes were made by the author agent before it stopped on a provider rate limit; the main agent verified them in the files, completed issue 3's prereg wording and issue 6's notes, re-ran `tests.py`, checked the panel and recomputed every digest.

| # | Issue | Fix (file:line) |
|---|---|---|
| 1 (must-fix) | G7 items revealed the proxy decision through their pair count | `validate.py:139-148`: items are now pair-level. Explained partners emit their `via_ref` pair; hidden partners emit one item per seeded via (≤ 3). Opaque `D…` ids are assigned after a global shuffle; `partner_key`, `lang` and `proxy` are stripped from the blinded file (`:149-156`). `gates()` aggregates pairs to partners (`:217-229`: any `yes` → yes; else any unresolved/`cannot_tell` → unresolved; else no) and reports partner-level cell sizes (`:239`). Test: `tests.py:213-224` (4 partners / 8 pairs, aggregation, cell sizes, unresolved share, precisions). |
| 2 | G7 skipped a sampled language with no decided hidden partner | `validate.py:235-238` iterates over every sampled hidden language; a language with `n = 0` reports `hidden_precision: None`, and `:242-244` fails the gate in that case. Tests: `tests.py:225-232` (fails with zero decided; passes once decided). |
| 3 | G1 exclusion was reported but not enforced | `build_panel.py:22,27-30,159,163`: `snapshot_coverage()` (same formula as `gates.py`) is recorded in `repo-gate.json` and required ≥ 0.85 for `pass`. Prereg l.35 now lists G1 in the repository gate. Test: `tests.py:235-240`. Re-run outcome: all 16 repositories have coverage ≥ 0.948 (axum lowest); the panel is unchanged (7,717 rows, 12 repositories). `panel*.csv.gz` digests changed because the files were rewritten, not their content. |
| 4 | Stale usage lines | `analysis.py:5` now includes `--gates gates.json`; `harness.py:6` is `[sims] [out.json]`. |
| 5 | SESOI rationale mischaracterised R-017 and R-022 | Prereg l.131: "R-015 reports defect associations, R-017 coordination/resolution-time effects, R-022 is conceptual; none reports an effect size for incomplete changes". SESOI values unchanged (decision c). |
| 6 | DESIGN-NOTES §9 listed decided questions as open | `DESIGN-NOTES.md` §9: items 1–4 marked closed with the round-2 decisions; items 5–8 remain open. §8 updated to 59 checks and the new G7/G1 tests. |

## Verification (main agent, 2026-10-20)

- `uv run --with-requirements requirements.txt python tests.py`: **59 checks pass**.
- Harness: `analysis.py`'s digest changed after round 3 (issue 4's usage line), so the harness was re-run (200 sims per scenario) against the current file: output identical to the stored `harness-results.json` apart from timing fields, and every scenario's checks are true. `power.py` and `power-results.json` are unchanged from round 3.
