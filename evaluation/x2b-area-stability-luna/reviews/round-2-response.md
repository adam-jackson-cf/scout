# X2b round-2 response (preregistration revision 3)

Answers `reviews/round-2.md` (7 issues: 2 major, 2 minor, 3 nits). Council replacement texts used verbatim except where noted under "Judgement calls". No model call; no collector, `x2b.py`, `run.sh`, stub or diff change; ledgers, design and Go untouched.

| # | Fix | Where |
|---|---|---|
| 1 | §2.6 replaced with the council text (code under test `2a2a346` and six blobs; run commit = descendant of `2a2a346` containing the signed-off files, checked by `git diff --quiet 2a2a346 <run commit> -- internal/ scripts/ third_party/clawpatch`, empty `git status --porcelain`, `scout_head` = run commit, recorded `*_sha256` = `2a2a346` blobs, `shasum -a 256` of every frozen file in the run worktree = sign-off digests); §10 last bullet, §11 (two places), README real-run paragraph and frozen-files paragraph as specified | `PREREGISTRATION.md:45,100,103-104`; `README.md:27,41` |
| 2 | Status line and design sentence as specified (rev 2.5 adopted 2026-10-24, digest `032285b3…`, `round-6.md`, `round-6-response.md`, `2a2a346`); the old §2.6 opening sentence ("Collection starts only after …") deleted; §12 "the rev 2.5 proposal named above" → "revision 2.5 (adopted; `round-6.md`)"; README `:3`, `:14`; DRAFT-NOTES digest rows, judgement 7, ledger heading and text, "Not done" bullets | `PREREGISTRATION.md:3-4,107`; `README.md:3,14`; `DRAFT-NOTES.md` |
| 3 | `:27` and DRAFT-NOTES judgement 9 as specified | `PREREGISTRATION.md:27`; `DRAFT-NOTES.md:38` |
| 4 | §4b outcome column replaced as specified | `PREREGISTRATION.md:71` |
| 5 | (a) 226–250, "exactly 10"; (b) 45/45/45/41/54/32; (c) C3 clause rewritten in §4(iv) and in the render string. Confirmed by execution before freezing: `math.ceil(0.04*250) = 10`, `ceil(0.04*251) = 11`, `ceil(0.08*125) = 10`, `ceil(0.08*126) = 11`; `C.wilson(6,126)[1] = 0.1000017 > 0.10`, `C.wilson(6,127)[1] = 0.09924 <= 0.10`, `C.wilson(6,125)[1] = 0.1008`, `C.wilson(6,120)[1] = 0.1048` | `PREREGISTRATION.md:57`; `analysis_x2b.py:155` |
| 6 | Fixed in code (first option): the missing-versions check now also asserts `"(X2b section 4): NO"` and `"no tool-versions file"` on the rendered citability line (check text updated; count unchanged); §8 text therefore stays true | `validation/validate.py:202-205` |
| 7 | Not-citable and inconclusive strings as specified (`[ASSUMPTION; X2 did not confirm 0.5, design A1]`; not-citable names the unexpected arm and "a new preregistration is needed"); §4a third bullet tag changed; `validate.py` `startswith("none: not citable")` and `"match_jaccard stays 0.5"` still hold | `analysis_x2b.py:100,109`; `PREREGISTRATION.md:65` |

## Judgement calls
1. **Governance sentence kept.** The council's `:3` replacement omits "Governed by `AGENTS.md` experiment rules …"; kept it (one sentence) at the end of the status line rather than delete a rule statement.
2. **`:11` wording.** "preceded this draft" → "preceded the first draft" (the document is no longer a draft; not in the council text; wording only).
3. **§2.6 numbering.** The council's single paragraph replaces the two former lines `:45-46`; the line now carries the whole procedure.

## Verification
- `source .tools/activate && evaluation/x2b-area-stability-luna/run.sh validate` → all passed: **53 X2b checks + X2's 233 = 286 PASS, 0 FAIL** (same count as round 1's response; one check strengthened, none added).
- Every validation claim edited in this round was checked against `validate.py` by name: §8 "each also asserted on the rendered citability line" (`validate.py:195-205,220-221`); S6x/`startswith` strings (`:197,:253-254`); the 225/226 boundary text (`:303-330`) unchanged.

## Digests (SHA-256)
| File | SHA-256 |
|---|---|
| `PREREGISTRATION.md` | `ce4e19caed4c0a0d4e9a5d9dd5c1f81598e9a0d6764511d59786ddf0a93b4337` |
| `README.md` | `d8ca96dcc507521e0e7912cae826615242053f40d8c1aaaae4567a7aabdc0ae3` |
| `x2b.py` | `2f63debd7363715d0f4d17019d3713de3a6ce8f01e31ed7aa46a28d3ab0f23ad` (unchanged) |
| `collect_x2b.py` | `c1aa630764de2d17dabc2fc296b7f3e7542f84d5c9343da584459102ae369fec` (unchanged) |
| `analysis_x2b.py` | `f1302f6cb95e59fba8a46586d5779be58956654d2937bd5a5fc1ecfeff3b0d09` |
| `run.sh` | `50f09fd6bd2f2bc6ee19497ba71d04f0445dfeecd2afbed94999a90162bcfc92` (unchanged) |
| `validation/validate.py` | `b1e26ffc3d3e505a60344bd94af975fe9a2f95ea5f9379b329fa3fd9d27c4fd5` |
| `validation/stub_clawpatch_x2b.js` | `3ac67b968c77995c23ba1965d28ea5b0037d170eda0a6a04039a167c730a135a` (unchanged) |
| `validation/collect-x2b.diff` | `efef0f82b3d2493bbcbc8851f40f2e2db19c8627197e8eada0e7ca83d8615a24` (unchanged) |

Code under test, `git show 2a2a346:<path> | shasum -a 256` (each equals the working-tree file; `git diff --quiet 2a2a346 HEAD -- internal/ scripts/ third_party/clawpatch` reports no difference at HEAD `42bf9cc`; `git status --porcelain internal scripts` empty):
`internal/clawpatch/client.go` `ff05791c317eef52061755f2d9124c118c53c09004ec81ea6a2102d6c683fd88`; `internal/clawpatch/provider.go` `3b3a97346b87c676263ce80337d9bb911493a82c31f8662f3f1d12a022226e7f`; `internal/areas/areas.go` `8ed2e70815fdcc605d60402fb603cfb905d58457a5869b3795d59507a93cb484`; `internal/clawpatch/records.go` `57344e025d5e2d0b7a10abf543b20b55f899602c1dd95d2b6a2b62d9b823dfbe`; `scripts/clawpatch-safe.patch` `dc867503f80cdb336681d72c221ec1dea24006f121b4cae56bf0fb284dae6ce3`; `scripts/build-clawpatch.sh` `99e3f6b3c670cd8c297b4ed498eafc735b7b5045aea8b97c7b55dc6343c21653`.
`git ls-tree 2a2a346 evaluation/x2b-area-stability-luna` is empty (the study is untracked at `2a2a346`, as expected).

## For the main agent
- Commit the signed-off X2b files in one commit after sign-off (run commit, a descendant of `2a2a346`); name it in the final round or in `RESULTS.md`; then, in a clean worktree there, run `git diff --quiet 2a2a346 <run commit> -- internal/ scripts/ third_party/clawpatch`, `git status --porcelain internal/ scripts/ evaluation/x2b-area-stability-luna/` and `shasum -a 256` of the frozen files.
- Ledgers: no edit needed now (HD5b row `hypothesis-ledger.md:30` still says "draft preregistration rev 2; council round 2 pending", updated at sign-off with the History line naming round, digests, `2a2a346`, run commit and 286). Optional "wording only, no state change" HD5b History line for rev 3 (round-2 observation (a)).
