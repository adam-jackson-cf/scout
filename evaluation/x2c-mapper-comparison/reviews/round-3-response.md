# X2c round-3 response

Date: 2026-10-28. Response to `reviews/round-3.md` (4 issues, all minor). Every fix is the council's exact fix; no choice was left open (issue 2: the council's preferred prose fix taken, not the alternative code duplication). Not touched: `.codegraph`, frozen X2/X2b, design, ledgers, STATUS. No commits, no real-data runs, no model calls.

## Per-issue

| # | Fix applied | Where |
|---|---|---|
| 1 | Prose moved to the code's rule, using the council's replacement text: l.68 "A pair is *unscored* when either order has no answer row in `scores.jsonl` at analysis time (… the last answer per (pair, order) is its score, so a failure row counts only while that order stays unanswered): it is never matched and never counted as a pass."; l.73 failure rows "are retained and reported (`failures_by_kind` and `cost_unknown_rows` in `jev_requests`, `foreign_model_failure_rows` in `jev_check`), not gated; their unit is unscored, and counts in that arm's unscored share, only while that order has no answer row (a resumed run may answer it);"; `jev_match.py` docstring l.4–5 replaced as specified. No code logic changed. | `PREREGISTRATION.md` l.68, l.73; `jev_match.py` l.3–6 (docstring only) |
| 2 | Council's preferred fix: the §12 list now reads "… failure rows by kind, `cost_unknown_rows`, total billed USD and the share of answers with attempts > 1 (`jev_requests`) and `foreign_model_failure_rows` (`jev_check`, rendered on the `Jev check:` line)". The alternative (duplicating the counter into `jev_requests`) was not taken; `analysis_x2c.py` unchanged. | `PREREGISTRATION.md` l.127 |
| 3 | README l.3 score wording replaced by "a Jev score s = P(same_concern) + P(overlapping_concern), mean of both presentation orders, on every file-sharing candidate pair"; l.19 now "`round-<N>.md` … `round-<N>-response.md`". One addition beyond the council's text, for consistency with PREREGISTRATION l.3: README l.3 "revision 2, draft for council review" now reads "revision 3, submitted for council round 4" (a stale revision number in the same sentence). | `README.md` l.3, l.19 |
| 4 | (a) `validate.py`: after the S4-model check, new check "S4 a model-mismatch failure row is reported, not gated …" (`foreign_model_failure_rows == len(pm_f) > 0`, no global reason, luna-xhigh unscored > 0). (b) After the S5 coverage-floor check: four planted global-gate cases (header another model, header another question version, answer row another model, answer row another question version), each patched into a copy of `p4`'s scores and run through `AX.analyze`; asserts a matching global reason, every `+jev` cell inconclusive, Jaccard `terra-none` eligible and selected. Code exactly as the council gave it (the council's line references were from before its own listing; the insertion points are the same checks). (c) §10 S bullet: "wrong model, no cost," → "a header or answer row naming another model or question version (fails every Jev cell), a `model-mismatch` failure row (reported, not gated), no cost,". | `validation/validate.py` (after S4-model; after S5 coverage floor); `PREREGISTRATION.md` l.105 |

Council memo item: PREREGISTRATION l.3 set to "Revision 3, submitted for council round 4" (also names rounds 2 and 3 and `reviews/round-<N>-response.md`).

Non-blocking notes N1–N7: decisions. N1 (label "seconds / 8 workers"): not changed (the label is in `analysis_x2c.py`, which is unchanged; the clarification is optional and `validate.py` accepts either). N2 (S1 check title says eight cells, planted run has 2): title only, accepted as written. N3 (`sum_of_maxima` None if a cell has no labelled match): accepted; RESULTS says so if it happens. N4 (`cost_unknown_rows` also counts no-reply failures): accepted, read beside `failures_by_kind`. N5 (1e-6 sum tolerance run risk): accepted; the main agent inspects `failures_by_kind` after the first few hundred `run.sh jev` rows and on a `schema` surge stops the scorer and records the deviation. N6 (`a is None` unreachable): accepted. N7 (reason for "approximates, not bounds"): accepted as correct.

## Validation

`source .tools/activate && evaluation/x2c-mapper-comparison/run.sh validate` → exit 0, output "X2c checks: 78 (plus X2's 233)", "all validation checks passed", 78 PASS lines in the X2c output. DRAFT-NOTES convention: 78 + 233 = **311 PASS, 0 FAIL** (round 2: 73 + 233 = 306; five checks added: one S4, four S5 global gates, matching the council's expected 311). About 277 s wall.

## Digests (SHA-256, recomputed 2026-10-28 after the validation run)

| File | SHA-256 | Since round 2 |
|---|---|---|
| `PREREGISTRATION.md` | `0a05fbca28bef210a59c8746895718917d7334cc03dcee3119cd71e9c0610657` | changed |
| `README.md` | `83479745585c65d5f8f2db8e2b1777466c154d06c59e5a8d379e33b9a370f1b0` | changed |
| `x2c.py` | `cbef182f37682630340ab0055c9c8e93c04dbf38be9036800fe4bfbb848c631e` | unchanged |
| `collect_x2c.py` | `7365e02bbfedfe4b66c678771803936302446ed07f5ac58e161f070ab9e9dcff` | unchanged |
| `build_candidates.py` | `3597e005e1e1388de04e82a47b8e59164fcf65c60b13c9dbad6c3c58d7c75029` | unchanged |
| `jev_match.py` | `1234d1c22b7417b6ab5b7c02d0d913866ffeca9cc42e8711b2a3923a60e10fa0` | changed (docstring) |
| `analysis_x2c.py` | `00f41631d9e27b72a0f7da5b6642f69e4db0181d8e4a2c63a546fa01aeed8648` | unchanged |
| `jev/main.go` | `6a6001a5710d1e9a661171874ab7c9a562bb1df4f7cd1b7c6d39512922a1af1f` | unchanged |
| `run.sh` | `a5f03458064ca61370f95927970f13c4322e3674ab42506efcdf005e55491514` | unchanged |
| `validation/validate.py` | `22483acf31fd1ad5aef1e0a2e248d872e1f62d8bb63c61150c5074d66f3dd6ff` | changed |
| `validation/stub_clawpatch_x2c.js` | `480340b2ca7be7cba69d14eac3017336d9d207e472f5e8b15692e271dd150b89` | unchanged |
| `validation/stub_jev.py` | `86e01ed2fb0349fdd2cbaff7facc3d60849125e8aa48eb5c36ff35097cbef1c6` | unchanged |
| `validation/collect-x2c.diff` | `b75cf8e937b9925bc0c9fc723d45a31e097f7da83bc4fe953960098c32d4e128` | unchanged |

Code under test (all unchanged; recomputed from the repository root): `internal/decisions/audit.go` `09ff2df8aece406181052c760a6798adaa0c4e37068aa4d9b724454bcb919cb2`; `client.go` `e56aadef3f245151e489f4a28de11ab953730a139ab4809085f5a95c0e6806ae`; `question.go` `832c87601e3ec22d0101bcb4a46160546dfdb56d4c10d7f3038f74eaaf3eee0b`; `rank.go` `605a39bb268b39b132aa2e8b851aa0ab17dd2faf3a10c4a4aced4575f61c36ab`; `run.go` `e696b6fa95b11b108763681462ef8f0002ae3dbcaa6d2c83af5928e15a3284ae`; `triage.go` `8d22dc18b5c0f74a9df3b44f0e81191b7e2d3e9e9009593875b224a70d4b161c`; `go.mod` `7009be395a9f776345c3810af869a955a8ec569f0e9c969e9dd81b61ef710f24`; `go.sum` `24d99eff5500559b16cff2acba29f1ceaf3d73110046cbc6906f61cc2a927255`; `internal/policy/consent.go` `c8e55ff285e2ba3e5a38e66e09704f2d0ef607f552a773954b1a02c2ce7c3e4d`; `internal/areas/areas.go` `317564d0eba850d8375663c271070b38f85acd7317a83f3e2df2efe5509323da`; `internal/clawpatch/records.go` `942d4efc31fd113ba831d8cd0d5759dc41d79c14f369f2d0632377706cecfe77`; `scripts/clawpatch-safe.patch` `dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3`; `scripts/build-clawpatch.sh` `1b3eda9d166759be4dc04761757e93658cb8cb5e1088b1b6832a77639bc6b9b5`.

`DRAFT-NOTES.md` is not a signed document and is not digested here (updated: revision 3, round 3 recorded, digest table, 311 PASS in the validation and HD5c lines).
