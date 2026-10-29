# X2c round-2 response

Date: 2026-10-28. Response to `reviews/round-2.md` (7 issues: 1 major, 6 minor). Every fix is the council's exact fix; no choice was left open (issue 1 and 2 follow the council's recommendation as stated in its decision memo). Not touched: `.codegraph`, frozen X2/X2b, design, ledgers, STATUS. No commits, no real-data runs, no model calls.

## Per-issue

| # | Fix applied | Where |
|---|---|---|
| 1 | l.73 prose replaced as specified (header and answer-row pin/question-version/billed-cost fail every Jev cell; `model-mismatch` and cost-unknown failure rows leave the unit unscored, are reported, not gated); parenthesis replaced. Code: `jev_requests["cost_unknown_rows"]` over answers + failures. Validation S1 asserts `cost_unknown_rows == 0`. | `PREREGISTRATION.md` l.73; `analysis_x2c.py` `analyze` |
| 2 | `fam` block replaced: per cell, max over the 13 grid thresholds and both repositories of `false_pass_probability(n, p)`, with `at` (`pct`, `repo`); `family["bound"]` renamed `sum_of_maxima`, `None` when any cell's value is `None`; note replaced as specified. §8 prose replaced from "and the per-cell probability" to "family-wise bound"; §12 list says "false-pass figures and their sum of maxima". Validation S1 keeps `0 < sum_of_maxima["0.1"] <= sum_of_maxima["0.08"]` and adds a per-cell check (value ≥ the one-repository figure at the lowest grid threshold for each repository; `at` names a grid threshold and a repository). | `PREREGISTRATION.md` §8, l.127; `analysis_x2c.py` `analyze`; `validate.py` S1 |
| 3 | `map_seconds(inventory, arms, repos)`; `sum_over_repos` is `None` unless every repository has a map time; `analyze` passes `repos`. New validation case `S5-halfseconds` exactly as specified. | `analysis_x2c.py`; `validate.py` after the S4-noseconds check |
| 4 | l.127 clause replaced as specified; `render` adds the "Jev per (arm, repo) refresh …" line; S1 asserts it is in the text. | `PREREGISTRATION.md` l.127; `analysis_x2c.py` `render`; `validate.py` S1 |
| 5 | l.68 clause replaced with the exact-ties wording. `jev_match.py` docstring unchanged (the council said l.5–7 reads correctly). | `PREREGISTRATION.md` l.68 |
| 6 | `stub_jev.Stub(scale=None)`, probabilities multiplied by `scale`, docstring fault list updated; new check `J schema: four options summing to 0.9 …` exactly as specified. | `validation/stub_jev.py`; `validate.py` after the J schema check |
| 7 | DRAFT-NOTES: R-048 row, HD5c row (score wording, "rev 2"), History ("rev 2", 306 PASS), open question 3 replaced as specified; `PREREGISTRATION.md` l.89 appended with the area-ratio scoping sentence. | `DRAFT-NOTES.md`; `PREREGISTRATION.md` l.89 |

Non-blocking notes N1–N6: no change; N1 (`run.sh jev` forwards extra flags): the real run passes none, any override is a `Deviations` entry in RESULTS.md. N2–N6 accepted as written.

## Validation

`source .tools/activate && evaluation/x2c-mapper-comparison/run.sh validate` → exit 0, output "X2c checks: 73 (plus X2's 233)", "all validation checks passed". The convention of DRAFT-NOTES is 73 + 233 = **306 PASS, 0 FAIL** (round 1: 70 + 233 = 303; three checks added: S1 family-wise per cell, S5 half-seconds, J unnormalised reply). About 260 s wall.

## Digests (SHA-256, recomputed 2026-10-28 after the validation run)

| File | SHA-256 | Since round 1 |
|---|---|---|
| `PREREGISTRATION.md` | `8fa87d05d5370f844f79f8cc5bac555172b266fa527212934536a935724a1ac7` | changed |
| `README.md` | `367dc6ea83cbcfad532b0cc8c02db78b054062613d026ca678de86506f559bd8` | unchanged |
| `x2c.py` | `cbef182f37682630340ab0055c9c8e93c04dbf38be9036800fe4bfbb848c631e` | unchanged |
| `collect_x2c.py` | `7365e02bbfedfe4b66c678771803936302446ed07f5ac58e161f070ab9e9dcff` | unchanged |
| `build_candidates.py` | `3597e005e1e1388de04e82a47b8e59164fcf65c60b13c9dbad6c3c58d7c75029` | unchanged |
| `jev_match.py` | `536b775fe4e29ffeeda1214540395c7d1b346a7d55c8369704774b3dbf73123b` | unchanged |
| `analysis_x2c.py` | `00f41631d9e27b72a0f7da5b6642f69e4db0181d8e4a2c63a546fa01aeed8648` | changed |
| `jev/main.go` | `6a6001a5710d1e9a661171874ab7c9a562bb1df4f7cd1b7c6d39512922a1af1f` | unchanged |
| `run.sh` | `a5f03458064ca61370f95927970f13c4322e3674ab42506efcdf005e55491514` | unchanged |
| `validation/validate.py` | `da66dad367f223a10c586e94fa1fc9081f682d2bfbff3a8e76e9e8ff75cf2ef1` | changed |
| `validation/stub_jev.py` | `86e01ed2fb0349fdd2cbaff7facc3d60849125e8aa48eb5c36ff35097cbef1c6` | changed |
| `validation/stub_clawpatch_x2c.js` | `480340b2ca7be7cba69d14eac3017336d9d207e472f5e8b15692e271dd150b89` | unchanged |
| `validation/collect-x2c.diff` | `b75cf8e937b9925bc0c9fc723d45a31e097f7da83bc4fe953960098c32d4e128` | unchanged |

Code under test (all unchanged since round 1; recomputed from the repository root):
`internal/decisions/audit.go` `09ff2df8aece406181052c760a6798adaa0c4e37068aa4d9b724454bcb919cb2`; `client.go` `e56aadef3f245151e489f4a28de11ab953730a139ab4809085f5a95c0e6806ae`; `question.go` `832c87601e3ec22d0101bcb4a46160546dfdb56d4c10d7f3038f74eaaf3eee0b`; `rank.go` `605a39bb268b39b132aa2e8b851aa0ab17dd2faf3a10c4a4aced4575f61c36ab`; `run.go` `e696b6fa95b11b108763681462ef8f0002ae3dbcaa6d2c83af5928e15a3284ae`; `triage.go` `8d22dc18b5c0f74a9df3b44f0e81191b7e2d3e9e9009593875b224a70d4b161c`; `go.mod` `7009be395a9f776345c3810af869a955a8ec569f0e9c969e9dd81b61ef710f24`; `go.sum` `24d99eff5500559b16cff2acba29f1ceaf3d73110046cbc6906f61cc2a927255`; `internal/policy/consent.go` `c8e55ff285e2ba3e5a38e66e09704f2d0ef607f552a773954b1a02c2ce7c3e4d`; `internal/areas/areas.go` `317564d0eba850d8375663c271070b38f85acd7317a83f3e2df2efe5509323da`; `internal/clawpatch/records.go` `942d4efc31fd113ba831d8cd0d5759dc41d79c14f369f2d0632377706cecfe77`; `scripts/clawpatch-safe.patch` `dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3`; `scripts/build-clawpatch.sh` `1b3eda9d166759be4dc04761757e93658cb8cb5e1088b1b6832a77639bc6b9b5`.

`DRAFT-NOTES.md` is not a signed document and is not digested here.
