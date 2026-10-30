# X2d — drafter's response to council round 7

Date: 2026-10-30. Response to `reviews/round-7.md` (3 minor issues, 0 blocking, 0 major). Every fix applies the council's text exactly; no code change. PREREGISTRATION is revision 8 (the council's instruction: status line "Revision 8 (after innovation-council round 7, `reviews/round-7.md`), DRAFT for council round 8").

## 1. Per-issue table

| id | fix applied | where |
|---|---|---|
| R7-01 | HD6 clause replaced by the council's wording ("and, over all clusters, its answers — a resolved cluster by its composed answer, an abstained cluster by its own unthresholded guess — beat B2 (one-sided 95 % lower bound of `gain_guess` > 0), so the gain is not wholly an artefact of where it abstains"); §14 "(the edge net of selection)" → "(the all-cluster edge with abstentions answered by Jev's own guess)"; ledger HD6 cell identical to §1 wording; new History line added above the round-6 line (round-6 line and HD6b unchanged; HD6 state `proposed`); change log rev 7 → rev 8 cites round 7 | `PREREGISTRATION.md` lines 3, 10, 27, 237; `docs/hypothesis-ledger.md` HD6 row (l. 32) and History l. 37 |
| R7-02 | `intervals.pooled` replaced by the council's string (bootstrap one-sided 90 % for e, 95 % for the guard h; condition b the one-sided 90 % Wilson lower bound of the pooled `P_assign`; `p_assign_bootstrap` reported, decides nothing) | `thresholds.json` l. 135 |
| R7-03 (a) | README header: "revision 8, after council round 7 `reviews/round-7.md`, draft for council round 8" | `README.md` l. 3 |
| R7-03 (b) | all three pointers → "`reviews/round-6-response.md` §4 (the full measured table, R6-07)" (README l. 44; PREREGISTRATION §11: "reports the measured cells in …" and the N4b pointer, the latter inside the (c) replacement) | `README.md` l. 44; `PREREGISTRATION.md` l. 171 |
| R7-03 (c) | clause replaced by the council's text ("so `gain_guess` is centred on 0 and the GO share is ≤ 0.10 in every cell (N4b, N4b-cap, N4b-ρ0; `reviews/round-6-response.md` §4); the guard's marginal size is not measured: … 61 GO of the 461 pilots not NO-GO") | `PREREGISTRATION.md` l. 171 |
| R7-03 (d) | "if `p_true < t_cov` → **none**" → "if `1 − p_true ≥ s` (`p_true ≤ t_cov`) → **none**" | `PREREGISTRATION.md` l. 119 |

Unchanged as instructed: `GUIDE.md`, `questions/*.json`, the Go drivers, `analysis_x2d.py`, `power_sim.py`, `compose.py`, `tune.py`, `x2d.py`, `validation/*`, `docs/research-ledger.md`. No product file touched, no commit, no model call.

## 2. Validation

`source .tools/activate && evaluation/x2d-merge-association-pilot/run.sh validate`, run after the last edit to the frozen files.
- Run 1: **647 checks, 1 failed** — `J the cumulative cost stop: a run whose billed cost exceeds -max-cost-usd stops early with exit 4 … and a re-run refuses to start (exit 2, no request)` (`validation/validate.py:669`). The detail tuple was not captured (output was tailed). All other 646 checks, including every power bar and null check, passed.
- Run 2 (no file changed in between): **647 checks, 0 failed**; the same check printed PASS.
- The failure is not caused by the edits (text-only changes; the check exercises the Go driver and the stub server) [INFERENCE: flaky; the cause was not diagnosed]. Two runs were made, not one, because of it. Round 8 should treat the check as intermittent or have it diagnosed; I did not change `validate.py` (frozen, no code change this round).
- Power INFO lines were not compared cell by cell with `reviews/round-6-response.md` §4 (the run was filtered to the summary lines in run 2 and tailed in run 1; `power_sim.py` and its inputs are unchanged by byte digest).

## 3. Digests (SHA-256, recomputed after validation; `x2d.frozen_files()`, 44 files; interpreter `.tools/venv/bin/python` 3.11.11)

Changed from round 6's response: `PREREGISTRATION.md`, `README.md`, `thresholds.json` (3 files). The other 41 digests are byte-identical to the round-6-response table where I compared them (`GUIDE.md`, `analysis_x2d.py`, `b2.py`, `baselines.py`, `build_controls.py`, `build_units.py`, the `DRAFT-NOTES.md` §5 digest list for every other row), i.e. equal to the rev 7 list in `DRAFT-NOTES.md` §5 before this round.

| File | SHA-256 |
|---|---|
| `GUIDE.md` | `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980` |
| `PREREGISTRATION.md` | `b32adc9e107feffb8107a3ba3dbdb47bcb5cb53a9d3d20557d52ec4138c003c6` |
| `README.md` | `00399498d02ce3036e3a205292c7aece66e2a15bfeb3d2125bc45a42734d977d` |
| `analysis_x2d.py` | `b0a66eedb739c2952b14f8468ae56da65d9628f6fa345f8dc51380520b397ade` |
| `b2.py` | `5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f` |
| `baselines.py` | `a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0` |
| `build_controls.py` | `7ded8c9177d8e21131ed91ee0aa423a4dd8c64ac35b215f82e93b1bfbfe6c81b` |
| `build_units.py` | `c052246c828024a25fc6a143d513e111ecbf63b42b8b5912dd37e960a1c3e243` |
| `collect_onboard.py` | `c93fa59c8fe305459336e08e52c60f365d8df0a70d7b6ca395c58e8f88da5fff` |
| `compose.py` | `296001923384652d7d31cdb33ca937ed409783023de2b6ee53c62e4ea8ab5bd1` |
| `frontend.py` | `dd4a0fdc1e3595eeefb82a8eafab600255b3b0fe95f67e2fb8a07eb5972ea0d8` |
| `gold.py` | `6b0afaad2a3ae6652a02207813c5a3492ba6954bd0a745b58a3f98c960bb321e` |
| `jev/main.go` | `63fbec0534a2ddf3d55d07c074c6a882e199ec2b90cf771a984c9fa9ee586ca0` |
| `label_x2d.py` | `fa5d60ef2bbf2147102cf986048f7711cd2ad90d8cd7a0ddbe89aa5162729dcb` |
| `measurement/lever-result.json` | `298902e717d0db8826c35d219695d386b19644f99cfc205d6b2d891100b0352c` |
| `measurement/lever_check.py` | `8f8008d3a28d72d5e768bc292e4917449fd8cb31ca0a642ff6057ab51b705a42` |
| `onboard/main.go` | `7e2ed1edc64b694851e7729e17af4f2ba63dafa7c9cdcf19cd3a5a255c08d46b` |
| `power_sim.py` | `a955f7177280bfd6d20a2514d21a04819c426610d594ee4e7dbfe0cdfa9c5279` |
| `prepass.py` | `1e5863d32974f1d7fecfd4df5c30b6e02aa137b9db491780eab11cc00b500696` |
| `prepass/commits.jsonl` | `1581986e0efbf26bb51aea894c3ffcdd23f54ef6e1d6d6ab0e4e99248f35bd1c` |
| `prepass/k0.json` | `f89e8d0b264eb33eb8bcb5aa4a4910bd462c3e1befea480b721ee514947a98ab` |
| `prepass/reserve.json` | `a73558ed54efdd767c5c0cf9a5accdcf41f511acd044d630208a8fb3e129f12a` |
| `prepass/rounds.json` | `0aad85b0199e747630d0e863b7a5e83b2624684c50c1150ee4e8775fde1d7516` |
| `prepass/scopes.json` | `a24bfc8d19aab9fd5ab0817f317cee560df65c63b1d63b5dbd10b5529d16a512` |
| `prepass/summary.json` | `e1cfced0da45bcd86cc516eef39fa70c10b3fa5d630820d415cdea47045cc1c1` |
| `quality.py` | `6d65293d64c76611368fa8eaf93d4f48432cf203bb65aed9ed63130b8527cade` |
| `questions/merge.area_covers@v1.json` | `cce59e612b67d76b96d927613c1a780ea64915c38210b7be8fceb5cfaa7e5af0` |
| `questions/merge.covered@v1.json` | `f37eea94bb6e2c720cbbb904066dd06c368ecc7581c90063b9f17e5a751058b9` |
| `questions/merge.which_area@v1.json` | `cbb7ba7e92432a1e0ffc9bf7d8718a9001ba4227137a6ab60a1e359393eff928` |
| `questions/onboard.covered@v1.json` | `ec4d62be051c016ab945738c155401a17f091654e1efff9cfd60ec18314eaa70` |
| `questions/onboard.same_responsibility@v1.json` | `31f3b25d3bbf535b026bd6f0a022d39f01e7c519cc22649195d341b75a49a7b1` |
| `questions/onboard.which_file_area@v1.json` | `425edf93d08dba42d7f1da755601d5dec27cded8b67593401b7656b24cf12f57` |
| `reference_arm.py` | `d139f0179c06ca5c6aff5784ceab2feb4113f888c51cd46f760f7a3d943bb739` |
| `run.sh` | `31029ba557d453e620b214f242ed6454622617e34f0850fcd7a0667ac82806c5` |
| `scopes.py` | `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072` |
| `thresholds.json` | `5c211eb939b4880ae21e79fdfcfc25060722f73a96b99d6f27aaceb2153094a8` |
| `tune.py` | `2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c` |
| `validation/r4_checks.py` | `f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52` |
| `validation/stub_clawpatch_x2d.js` | `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f` |
| `validation/stub_jev.py` | `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d` |
| `validation/stub_labeller.py` | `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5` |
| `validation/validate.py` | `131a45ccc5571062430f260385a9bafb1d7b1e2610a7b3248f4e9633bb9c381c` |
| `versioning.py` | `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168` |
| `x2d.py` | `32c093e972cff31e86dae51412e948c70f07348465d299e18036813f82c8d464` |

Not in the frozen table: `DRAFT-NOTES.md` `3be7a0d277a36009e41da543d315cf0b47dcacd0dffb26c98ca87b9c531209c0` (updated after validation; not frozen).

`tune_threshold.py` (external, `~/.omp/agent/skills/jev-actions/scripts/tune_threshold.py`): `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c` (recomputed 2026-10-30 after validation; equals the round-7 pin).

## 4. Open for round 8

Verify the three text fixes and the ledger edit; decide how to treat the intermittent `J the cumulative cost stop` check (one failure in two runs, passes in the second; no code changed in between).

## Main-agent addendum (after the drafter's response; for council round 8)

- The `J the cumulative cost stop` failure seen once in the drafter's first post-edit run did not reproduce: 8/8 passes running `jev_checks` alone and 18/18 running six copies in parallel under load (main agent, 2026-10-30). Its detail tuple was not captured.
- Hypothesis [INFERENCE]: a request the driver cancels at its cost stop can still reach the stub (`stub_jev.py` counts on arrival in a `ThreadingHTTPServer` thread, l.88–93) after the driver exits, so `before = stub.requests` races the late arrival and `stub.requests == before` fails on the re-run. No driver defect is implied: the re-run's refusal happens before any request (`jev/main.go` l.877–879).
- Change (validation only): `validation/validate.py` gains `_stub_quiescent(stub)` (wait until the stub's count has not changed for 0.5 s, at most 10 s) and the check reads `before` through it. No assertion changed.
- Full `run.sh validate` after this last edit: rc 0, `647 checks, 0 failed` / `all validation checks passed` (log `/tmp/x2d-validate-r8pre.log`).
- `validation/validate.py` SHA-256 is now `585082947ccb63b3a03bef01ff199cb234e3dfb33f091469e5ec0f05ad364816`; every other study file is as listed above.
