# E1b-v2 — Response to council review round 7

All three round-7 issues are fixed as text-only changes; no rebuttals. Issue 2 uses the council's recommended option A, amending the preregistration to the implemented definition. No real Jev or annotator calls were made, no model was fitted to real data, and no `validation/` directory exists.

| # | Issue | Fix (file:line) |
|---|---|---|
| 1 (nit) | Prereg G7 sentence and `annotate.py` docstring said `unparsable` ⇒ consensus `None` | `PREREGISTRATION.md:168`: "an `unparsable` annotator answer is a disagreement: C adjudicates the pair, and if C gives no parsable label the pair is unresolved". `annotate.py:12` uses the same wording. Both now agree with prereg `:149-150` and `validate.consensus`. |
| 2 (minor) | Prereg §3.1 implied rank → top 10 → filter absent; `collect.py:210-211` filters absent first | Option A. `PREREGISTRATION.md:65-66`: "Among the qualifying partners **absent from the PR**, the top 10 per file by confidence `n(f,g)/n(f)` (then `n(f,g)`, then path) are kept; a kept partner existing at `S` counts as *expected-but-absent*". This is `collect.py:210-211` (filter `g not in in_pr`, sort key `(-conf, -n, path)`, `[:TOP_PARTNERS]`) followed by the existence check at `:223-226`. `collect.py` and all data are unchanged. |
| 3 (nit) | `DESIGN-NOTES.md:35` still called the proxy unvalidated (open question 3) | Now: "Proxy validated by gate G7 before T5 is decided (G7 thresholds accepted in council round 2, open question 3 closed)". The Where column cites `collect.py, validate.py`. |

The optional >10-partner fixture was not added. The amended text now describes the code exactly, so the fixture would only pin the existing sort key.

## Verification (main agent, 2026-10-21)

- `tests.py`: **64 checks pass** (log `/tmp/e1bv2-tests-r8.log`).
- Harness and power not re-run. `analysis.py`, `power.py`, `harness.py` and `collect.py` are byte-identical to round 7.
- `DESIGN-NOTES.md` §10: 48 digests recomputed. Changed since round 7: `PREREGISTRATION.md`, `annotate.py`. `DESIGN-NOTES.md` is not in the digest list. Copy: `/tmp/e1bv2-digests-round8.txt`.
