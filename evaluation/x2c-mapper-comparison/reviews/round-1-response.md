# X2c — response to council round 1

Date: 2026-10-28. Responds to `reviews/round-1.md` (16 issues: 7 major, 9 minor). Every fix below was applied to the files named and is exercised by `run.sh validate`. No real-data run, no model call, no Jev call, no predictor–outcome or Jev-versus-label computation was made; only synthetic validation (stub Clawpatch, fake CLIs, a stub Jev endpoint). `.codegraph`, frozen X2/X2b files, the design, the ledgers and STATUS were not touched; nothing was committed.

## Decisions taken (main-agent decisions and council defaults)
| Issue | Choice | Reason |
|---|---|---|
| 1(c) coverage floor | **Option B: `COVERAGE_FLOOR = 0.5` on owned files; area ratio reported, not gated** (main-agent decision) | The user asked for haiku and the lowest suitable terra effort to be compared; a 0.8 floor would exclude them by construction. Recorded in PREREGISTRATION §7. |
| 3 score definition | **Council recommendation:** `s = mean over the two orders of P(same_concern) + P(overlapping_concern)`; greedy order (−s, −P(same_concern), prior id, next id) | Not resolvable from the council text alone; the council's recommended option was taken (listed for the main agent). PREREGISTRATION §6 item 5 gives the reason. |
| 15 per-cell false-pass probability (n used) | n = realised labelled count at t\* for an `adopt` cell, else at 0.30 | The council text says "the realised labelled count" without naming the threshold; this was the drafter's choice (stated in §8). |
| 16 open-question defaults | The seven DRAFT-NOTES defaults recorded as assumptions in §3 | As the council asked; frozen at sign-off. |

## Per-issue fixes
| # | Fix | Where | Check |
|---|---|---|---|
| 1 | (a) Disclosure bracket with the smoke/X2b counts, adapted to floor 0.5 (at 0.5 haiku default on fastapi, 109/243 ≈ 0.45, is expected just below, every other requested arm above; the council's "admits every requested arm" is not exact for that one cell). (b) The haiku sentence of §3 replaced: `haiku-low` excluded to save collection time only, haiku default kept because the user asked for it. (c) Option B recorded with its reason in §7; `x2c.COVERAGE_FLOOR = 0.5`. Mirrored in DRAFT-NOTES judgement 6, 8, 9, open question 1 and the HD5c row (≥ 50 %). | PREREGISTRATION header (b), §3, §7; `x2c.py`; DRAFT-NOTES | validation S5 "coverage floor is the preregistered 0.5" |
| 2 | `coverage_check` computes mean areas per map and `area_ratio_to_reference` per repository; `select()` appends "areas per map: … of the reference" to the proposal (and a coarser-granularity note when a ratio is < 0.8); §1 parenthesis, §7 Consequence, §11 and §12 updated. | `analysis_x2c.py`; PREREGISTRATION | S5 "area ratio is reported, not gated": haiku ratio in (0.30, 0.40), proposal carries `areas per map` |
| 3 | `read_scores` stores (P(same)+P(overlapping), P(same)) from `probabilities`; `pair_score` returns the two-order means; `greedy` sorts by (−s, −p_same, prior id, next id) and thresholds on `s`; `jaccard` in Jev rows carries `s`; stub puts the residual mass on `different_concern`; `main.go` unchanged for this. Validation S1–S3 needed no re-pin (planted `s` equals the planted P(same)). | `jev_match.py`; `stub_jev.py`; PREREGISTRATION §6 item 5, §7 proposal text | S6 score unit (0.3 + 0.6 → 0.9, P(same) 0.3 kept; tie-break by P(same)); S1–S3 unchanged counts |
| 4 | §6 states the ÷ 8 as an assumption (product `decisions.Run` is sequential, run.go:186); §7 Consequence: a Jev `adopt` needs a concurrent runner; `jev_refresh_cost` adds `seconds_sequential`; Jev cells carry it; `jev_requests` reports the share of answers with `attempts` > 1; all rendered; §12. | `analysis_x2c.py`; PREREGISTRATION §6, §7, §12 | S1 "sequential-runner … items are computed and rendered" |
| 5 | `jev_check` puts unscored breaches in `arm_reasons[arm]`; `analyze` uses `jr = reasons + arm_reasons[arm]` per cell; `arm_reasons` rendered; §6 and §7 Citable updated. | `analysis_x2c.py`; PREREGISTRATION | S4 two-arm pipeline: failed requests on luna features leave `terra-none+jev` eligible, `luna-xhigh+jev` inconclusive, outcome `adopt`; S4 single-arm check reads `arm_reasons` |
| 6 | Added: failed confirmation through `analyze()` (patched `confirm`) → ineligible, `no-adoption`; missing map seconds (re-plant with `seconds=None`) → Jev cell ineligible "no map wall time recorded", `no-adoption`; the l.266 check renamed "per-refresh Jev cost is computed (> 0)". Deviation from the council text: the Jaccard luna cell is `fallback` (`not-adopted`) so only the Jev cell can be `ineligible` for a missing time. Checks for issues 2, 5, 8, 9, 12 added. | `validation/validate.py` | S5 two new checks; S4 rename |
| 7 | `go_code_digests` digests every non-test file of `internal/decisions`, `consent.go`, `areas.go`, `records.go`, `go.mod`, `go.sum` (docstring corrected); `run.sh jev` appends a line to `<data>/jev/code-digests.jsonl` before `go run`; `analyze` stores the lines under `res["x2c"]["jev_code_digests"]`, `render` prints them; §6, §12, README list the files. | `x2c.py`; `run.sh`; `analysis_x2c.py`; PREREGISTRATION; README | C "run.sh jev records the Jev code digests" |
| 8 | Driver: after the reply, `choiceComplete` requires one probability per option and a sum within 1e-6 of 1, else a failure row of kind `schema` (billed cost and model kept) and the unit is unscored; tolerance stated in §6. | `jev/main.go`; `stub_jev.py` (`drop_option`) | J "schema": three of four options → exit 3, kind `schema`, no answer |
| 9 | `existing` returns `spent` (Σ `cost_usd` over answer and failure rows with `cost_known`); `cost := spent`; refuse to start when `spent` exceeds the cap ("budget already exhausted"); §6 says the stop is cumulative. The optional `cancel()` on the first unknown-cost reply was **not** implemented. | `jev/main.go`; PREREGISTRATION §6 | J "cost stop is cumulative": resume after the cap exit refuses (exit 2) with no request |
| 10 | `billed_cost_usd` = Σ over answer and failure rows with `cost_known`; failure rows reported by `kind`; foreign-model failure rows counted (reported, not gated); `jev_refresh_cost` starts from the measured transitions (status ok at the first threshold) with 0 for transitions without candidates; per-refresh figure answers-only (§6); §12. | `analysis_x2c.py`; PREREGISTRATION | S4 "per-refresh Jev cost" (6 transitions per repository); S1 `failures_by_kind` |
| 11 | `collect)` passes `"$D"` and the collector gets `--data`; the usage comment says `X2C_DATA` covers `collect`. Validation by refusal rather than by stub Clawpatch writes: `run.sh collect` cannot be given the stub binary or repository spec, so a record planted under `$X2C_DATA` makes the first collection refuse as already recorded, naming that file, with no data created in the default directory. | `run.sh` | C "run.sh collect passes X2C_DATA to the collector" |
| 12 | `render` prints the placeholder NOTE after `A.render(res)`; §12 carries the sentence. | `analysis_x2c.py`; PREREGISTRATION §12 | S1 "report states that `<arm>+jev` … schema placeholders" |
| 13 | §6 Matching reworded as the council wrote it, with the (−s, −P(same), prior id, next id) order; docstring of `jev_match.py` aligned. | PREREGISTRATION §6; `jev_match.py` | — (prose) |
| 14 | R-008 and R-038 added to the sources line and to HD5c's sources (DRAFT-NOTES draft rows; ledgers untouched); effect column "—"; ledger outcome sentence appended to §7 "Ledger state". | PREREGISTRATION header, §7; DRAFT-NOTES | — (prose) |
| 15 | (a)/(b) §8 text added; (c) §7 text added and `select()` appends "X2b (same configuration): fallback" for a luna-xhigh Jaccard `adopt` (`X2B_LUNA_JACCARD_VERDICT`), `analyze` reports `x2b_comparison`; (d) §12 RESULTS list extended; computed and rendered: map-time dispersion (n, mean, min, max, per map id), the candidate ceiling, the family block (cells examined and eligible, per-cell false-pass probability at 0.08 and 0.10, their sum). | PREREGISTRATION §7, §8, §12; `analysis_x2c.py` | S1 reporting-items check (family bound, ceiling, `x2b_comparison.agree`, rendered strings) |
| 16 | §5 label-use sentence replaced (named exploratory Jev–label statistics, X2 §6 superseded for X2c only); §1/disclosure "does not require Jaccard ≥ 0.30 but still requires one shared owned file"; §6 minimums sentence; §7 unmeasured-reference sentence; the seven open-question defaults recorded in §3. | PREREGISTRATION | — (prose) |

## Validation
`source .tools/activate && evaluation/x2c-mapper-comparison/run.sh validate` → exit 0, `all validation checks passed`: **70 X2c checks** (counted by the script; one of them is X2's own validation, 233 PASS, 0 FAIL, re-run unchanged) **+ 233 = 303 PASS, 0 FAIL**; 12 more X2c checks than the 58 of the draft. Log of the run: `/tmp/x2c-validate-r1.log` (not kept in the repository).

## SHA-256 of every study file (recomputed after the final validation run)
| File | SHA-256 |
|---|---|
| `PREREGISTRATION.md` | `ab99dcd7a387612645e1009a0f0637bf53d09574c39473714cc636718edb1801` |
| `README.md` | `367dc6ea83cbcfad532b0cc8c02db78b054062613d026ca678de86506f559bd8` |
| `x2c.py` | `cbef182f37682630340ab0055c9c8e93c04dbf38be9036800fe4bfbb848c631e` |
| `collect_x2c.py` | `7365e02bbfedfe4b66c678771803936302446ed07f5ac58e161f070ab9e9dcff` |
| `build_candidates.py` | `3597e005e1e1388de04e82a47b8e59164fcf65c60b13c9dbad6c3c58d7c75029` |
| `jev_match.py` | `536b775fe4e29ffeeda1214540395c7d1b346a7d55c8369704774b3dbf73123b` |
| `analysis_x2c.py` | `f34501100bea1351badad7f3b80cc70924d17b77c48a7f0b37c6221d261d8e55` |
| `jev/main.go` | `6a6001a5710d1e9a661171874ab7c9a562bb1df4f7cd1b7c6d39512922a1af1f` |
| `run.sh` | `a5f03458064ca61370f95927970f13c4322e3674ab42506efcdf005e55491514` |
| `validation/validate.py` | `93b9a9d7d427aa2322a31dd66a306af7317718dfce60beee3ce7a4455f54293e` |
| `validation/stub_clawpatch_x2c.js` | `480340b2ca7be7cba69d14eac3017336d9d207e472f5e8b15692e271dd150b89` |
| `validation/stub_jev.py` | `972e5caae4301d3f24722813cfb3cbae57c4facc492cbf0127b749c4ec913f0d` |
| `validation/collect-x2c.diff` | `b75cf8e937b9925bc0c9fc723d45a31e097f7da83bc4fe953960098c32d4e128` |
| `DRAFT-NOTES.md` (not a signed document) | `cac4eca0cae84536195b12cd83aee2c8c368b0ac7e7f5705de6e4665b39bf1f5` |
Code under test (non-test files; blob digests for the sign-off round):
| File | SHA-256 |
|---|---|
| `internal/decisions/audit.go` | `09ff2df8aece406181052c760a6798adaa0c4e37068aa4d9b724454bcb919cb2` |
| `internal/decisions/client.go` | `e56aadef3f245151e489f4a28de11ab953730a139ab4809085f5a95c0e6806ae` |
| `internal/decisions/question.go` | `832c87601e3ec22d0101bcb4a46160546dfdb56d4c10d7f3038f74eaaf3eee0b` |
| `internal/decisions/rank.go` | `605a39bb268b39b132aa2e8b851aa0ab17dd2faf3a10c4a4aced4575f61c36ab` |
| `internal/decisions/run.go` | `e696b6fa95b11b108763681462ef8f0002ae3dbcaa6d2c83af5928e15a3284ae` |
| `internal/decisions/triage.go` | `8d22dc18b5c0f74a9df3b44f0e81191b7e2d3e9e9009593875b224a70d4b161c` |
| `go.mod` | `7009be395a9f776345c3810af869a955a8ec569f0e9c969e9dd81b61ef710f24` |
| `go.sum` | `24d99eff5500559b16cff2acba29f1ceaf3d73110046cbc6906f61cc2a927255` |
| `internal/policy/consent.go` | `c8e55ff285e2ba3e5a38e66e09704f2d0ef607f552a773954b1a02c2ce7c3e4d` |
| `internal/areas/areas.go` | `317564d0eba850d8375663c271070b38f85acd7317a83f3e2df2efe5509323da` |
| `internal/clawpatch/records.go` | `942d4efc31fd113ba831d8cd0d5759dc41d79c14f369f2d0632377706cecfe77` |
| `scripts/clawpatch-safe.patch` | `dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3` |
| `scripts/build-clawpatch.sh` | `1b3eda9d166759be4dc04761757e93658cb8cb5e1088b1b6832a77639bc6b9b5` |

## Not done / open
- Issue 9's optional `cancel()` on the first unknown-cost reply was not implemented (the council marked it optional).
- Issue 11's validation uses a refusal under `X2C_DATA` rather than a full stub-Clawpatch collection through `run.sh collect` (the stub cannot be injected through `run.sh`).
- Issue 6: only the Jev cell becomes `ineligible` for a missing map time (the luna Jaccard cell is `fallback`), not "both luna cells".
- At floor 0.5 haiku default on fastapi is expected just below the floor on the smoke counts (109/243 ≈ 0.45, one draw); the study will report it either way.
- Round 2 (council re-review, zero issues) is required before any experiment is run (rule 1); the digest table above is the sign-off candidate.
