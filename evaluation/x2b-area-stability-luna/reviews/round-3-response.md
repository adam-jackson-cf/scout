# X2b round-3 response (preregistration revision 4)

Answers `reviews/round-3.md` (1 issue, nit, parts a/b). Council replacement texts used verbatim. No model call; no change to `x2b.py`, `collect_x2b.py`, `analysis_x2b.py`, `run.sh`, `validation/`, ledgers, design or Go.

| # | Fix | Where |
|---|---|---|
| 1a | §2.6 last sentence replaced: the versions file records `x2b.client_go_sha256`, `x2b.provider_go_sha256`, `areas_go_sha256`, `records_go_sha256`, `x2b.collect_x2b_sha256`, `x2b.x2_pins`, `clawpatch_cli_sha256`, `scout_head`; RESULTS shows the four Go digests = `2a2a346` blobs, `collect_x2b_sha256` = sign-off digest, `x2_pins` = `X2_PINS`, `clawpatch_cli_sha256` with `clawpatch_build_safety` (no git blob). README frozen-files sentence replaced likewise | `PREREGISTRATION.md:45`; `README.md:41` |
| 1b | Run-commit wording: `:3` (status "Revision 4"; final round names the run-commit rule, run commit recorded in `RESULTS.md` when created after sign-off); `:103` "blob digests in the sign-off round; the run commit per §2.6"; `README.md:41` as above | `PREREGISTRATION.md:3,103`; `README.md:41` |
| rev | `README.md:3` "revision 4"; `DRAFT-NOTES.md` header "Revision 4 draft", round-3 response reference, "rounds 1–3 … round 4 pending", digest heading "round-3 response", two digest rows | `README.md:3`; `DRAFT-NOTES.md:3,5,8-9` |

## Judgement calls
1. `DRAFT-NOTES.md` header also names `reviews/round-3-response.md` and "rounds 1–3 / round 4 pending" (the council asked only for "Revision 4 draft" and the two digest rows; consistent with the file's purpose). The "Not done" bullet still reads "Round 3 review of the round-2 response" (not named by the council; DRAFT-NOTES is not a signed document).

## Verification
- `source .tools/activate && evaluation/x2b-area-stability-luna/run.sh validate` → all passed: **53 X2b checks + X2's 233 = 286 PASS, 0 FAIL** (unchanged).

## Digests (SHA-256, recomputed)
| File | SHA-256 |
|---|---|
| `PREREGISTRATION.md` | `7efb06334d4a7f6086e536286decd4443076e988d724443ff8e626a197ab207d` |
| `README.md` | `0b9f6de2f3250bfb389ff03368007bc38a31cf536f01e8064bbfce3b1d9bdf3a` |
| `x2b.py` | `2f63debd7363715d0f4d17019d3713de3a6ce8f01e31ed7aa46a28d3ab0f23ad` (unchanged) |
| `collect_x2b.py` | `c1aa630764de2d17dabc2fc296b7f3e7542f84d5c9343da584459102ae369fec` (unchanged) |
| `analysis_x2b.py` | `f1302f6cb95e59fba8a46586d5779be58956654d2937bd5a5fc1ecfeff3b0d09` (unchanged) |
| `run.sh` | `50f09fd6bd2f2bc6ee19497ba71d04f0445dfeecd2afbed94999a90162bcfc92` (unchanged) |
| `validation/validate.py` | `b1e26ffc3d3e505a60344bd94af975fe9a2f95ea5f9379b329fa3fd9d27c4fd5` (unchanged) |
| `validation/stub_clawpatch_x2b.js` | `3ac67b968c77995c23ba1965d28ea5b0037d170eda0a6a04039a167c730a135a` (unchanged) |
| `validation/collect-x2b.diff` | `efef0f82b3d2493bbcbc8851f40f2e2db19c8627197e8eada0e7ca83d8615a24` (unchanged) |

Code under test, `git show 2a2a346:<path> | shasum -a 256` (each equals the working-tree file; `git diff --quiet 2a2a346 HEAD -- internal/ scripts/ third_party/clawpatch` reports no difference at HEAD `42bf9cc`):
`internal/clawpatch/client.go` `ff05791c317eef52061755f2d9124c118c53c09004ec81ea6a2102d6c683fd88`; `internal/clawpatch/provider.go` `3b3a97346b87c676263ce80337d9bb911493a82c31f8662f3f1d12a022226e7f`; `internal/areas/areas.go` `8ed2e70815fdcc605d60402fb603cfb905d58457a5869b3795d59507a93cb484`; `internal/clawpatch/records.go` `57344e025d5e2d0b7a10abf543b20b55f899602c1dd95d2b6a2b62d9b823dfbe`; `scripts/clawpatch-safe.patch` `dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3`; `scripts/build-clawpatch.sh` `99e3f6b3c670cd8c297b4ed498eafc735b7b5045aea8b97c7b55dc6343c21653`.
