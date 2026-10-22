# E1b-v2 — Response to council review round 8

The single round-8 nit is fixed; no rebuttals. The change is text-only in `DESIGN-NOTES.md`. No script, preregistration, data or result file changed, so the round-8 digests stand.

| # | Issue | Fix (file:line) |
|---|---|---|
| 1 (nit) | `DESIGN-NOTES.md:66` labelled all-panel raw active-hour quantiles as `Y_active` next to n = 7,030 | `DESIGN-NOTES.md:66` now separates the two row bases. **All 7,717 panel rows, raw hours** (`constructs.json` `active_hours_q`, `single_commit_share` and `neg_duration_share`, computed over all rows in `constructs.py:55-56`): median 1.0 h, p75 2.3 h, p95 8.4 h; single-commit 42 %; negative durations 0. **`Y_active`, n = 7,030 non-missing** (`y_active_q`, `y_active_n`, `y_active_floor_share`, `constructs.py:53-54`; variance block `:74`): median 0 (1.0 h), p75 0.933 (≈ 2.5 h), p95 2.187 (≈ 8.9 h); floor share 38.2 %; within-repo SD 0.94 log; area ICC 0.006. |

## Verification (main agent, 2026-10-21)

- Values read from `constructs.json:1228-1243`: `y_active_q` 0.5 = 0.0, 0.75 = 0.9327, 0.95 = 2.1866; `active_hours_q` 0.5 = 1.0, 0.75 = 2.2761, 0.95 = 8.4461.
- `tests.py`: 64 checks pass. No code changed since round 8 (log `/tmp/e1bv2-tests-r8.log`).
- Digests: unchanged from `/tmp/e1bv2-digests-round8.txt`. `DESIGN-NOTES.md` is not in the digest list, and §10 still matches every file.
