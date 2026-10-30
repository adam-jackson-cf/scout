# Round 11 response: cap change (rev 12, user decision R-052)

Not a council review round. This records the one change made after the round-11 sign-off (`reviews/round-11.md`): the Jev cost caps rise from $1 per round / $10 total to **$2 per round (round 0 included) / $20 total** (user: "$2 per round, $20 total"; `docs/research-ledger.md` R-052). The change voids the round-11 sign-off (`signoff.json` still pins the round-11 bytes) and needs council round 12 with zero issues. No model call, nothing under `internal/`, `cmd/`, `scripts/`, `toolchain/`, `third_party/`, `go.*`, no write to `../scout-x2d`, no commit, `signoff.json` unchanged (SHA-256 `e7059f2a9407e38c4d5e8aff0480913fab36a32a69ccbb9721c9633934fdbc2d` before and after).

## 1. Edits

| # | File:line | Change |
|---|---|---|
| 1 | `x2d.py:45` | `CAPS`: `usd_per_round` 1.0 -> 2.0, `usd_total` 10.0 -> 20.0 |
| 1b | `x2d.py:521` | docstring "$10 total" -> "$20 total" (text only) |
| 2 | `thresholds.json:181-182` | `caps.usd_per_round` 2.0, `caps.usd_total` 20.0 |
| 2b | `thresholds.json:183` | caps note "$10 total" -> "$20 total" (text only) |
| 3a | `PREREGISTRATION.md:3` | status: "Revision 12 (after the round-0 cost stop; user decision R-052), DRAFT for council round 12" |
| 3b | `PREREGISTRATION.md:14` | change log rev 11 -> rev 12 |
| 3c | `PREREGISTRATION.md:16` | ledger-sources line: R-052 appended |
| 3d | `PREREGISTRATION.md:156` | §10 re-score "$10 total" -> "$20 total" (R3-17) |
| 3e | `PREREGISTRATION.md:192` | §12: "< $0.30 per round" struck through and marked superseded, measured prices (about $0.000092 fan-out, about $0.000589 other; 7,755 of 7,872 requests, $1.0001), projection about $9.74 (R-052, marked a projection), caps $2 / $20, re-score against the $20 total and a round's $2 |
| 3f | `PREREGISTRATION.md:242` | §14 item 19: the round-0 cost stop and the decision, verbatim "$2 per round, $20 total", R-052 |
| 4 | `README.md:69` | rev-4 note: "$10 total" -> "total cap only ($10 in rev 4, $20 from rev 12)" |
| 5a | `validation/validate.py:768` | `jev_budget_test`: round 4's spend `9.00` -> `x2d.CAPS["usd_total"] - 1.0` (the total binds as before: expected 0.75 / 0.35 / 0.40 unchanged); comments at 771-773 |
| 5b | `validation/validate.py:781,784-787` | `jev_budget_v7_test`: round 3 spend `1.0` -> `x2d.CAPS["usd_per_round"]`; expected values `8.2` / `0.7` replaced by `CAPS total - cap - 0.3 - 0.5` and `CAPS per round - 0.3`; one added conjunct `other_run < total left` |
| 5c | `validation/validate.py:678,778` | R3-17 check label and docstring name `x2d.CAPS` instead of $10 / $1 / 8.2; check count unchanged, none split |
| 6a | `DRAFT-NOTES.md:121` | section 14, rev-12 entry |
| 6b | `docs/hypothesis-ledger.md:37` | HD6 History line; HD6 stays `preregistered` |

Line 218 of `validate.py` already compares `thresholds.json` caps with `x2d.CAPS`, so it needed no edit.

Left unchanged and noted: `run.sh:248` (a comment still says "$10 total"; run.sh is frozen, the comment does not affect behaviour); the Go drivers' `-max-cost-usd` flag defaults of 1 in `jev/main.go:736` and `onboard/main.go:1081` (run.sh always passes the budget from `x2d.jev_budget`, so the defaults never apply in the pipeline). Either can be changed in council round 12 if wanted.

## 2. Validation (round-11 Addendum condition 3)

| Run | Condition | Result | Log |
|---|---|---|---|
| (a) | `signoff.json` moved aside, restored; SHA-256 before = after = `e7059f2a9407e38c4d5e8aff0480913fab36a32a69ccbb9721c9633934fdbc2d` | **651 checks, 0 failed** | `/tmp/x2d-val-r12-noRecord.log` |
| (b) | `signoff.json` in place | **not as expected: not exactly 1 failed.** The run aborted at the first pin-checking check: `FAIL L wide labelling on stub sessions…` (86 PASS before it, then a `StopIteration` traceback at `validate.py:854`). Its detail: `validation/validate.py` and `x2d.py` differ from the signed-off digests. | `/tmp/x2d-val-r12-withRecord.log` |

(b) cannot give the round-11 result until a sign-off record exists for these bytes: `signoff.json` still pins the round-11 digests of the five changed files (table below), so every stage that calls `x2d.verify_pins(require_signoff=True)` refuses, and the `L` check's stub labeller session fails before the later checks run. This is by construction (the record is written only after council sign-off, round-11 Addendum), not a defect in the edits. Re-run (b) after council round 12 writes the new `signoff.json`; the expected single FAIL is then `C without the council's sign-off record…`.

## 3. SHA-256 of every frozen file (44; computed after both validation runs, no frozen file edited since)

Five files differ from the round-11 record; each is marked with its round-11 digest.

```
GUIDE.md 3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980
PREREGISTRATION.md 668110fffc52414c490becd6543d9712f87e84a46db4968c81a74445e1e32c31   <- CHANGED (round-11 digest b5029a6fede44f3c88438f04075b86f3ff24f5ab9afdb19cfcf61865ecfdec22)
README.md fc51642bf03d55847a2086590932ee5bf44e64e8847b2cdc8eaaadd36e5d6124   <- CHANGED (round-11 digest 0b9fc29c8a7d18ae39df21e7340cd04d4ade10bdfbe73955ee6b9041b8528959)
analysis_x2d.py b0a66eedb739c2952b14f8468ae56da65d9628f6fa345f8dc51380520b397ade
b2.py 5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f
baselines.py a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0
build_controls.py a71c4e634b5a0d0b42eda15f8e58220fef159c077fd3bf22d36041dd3e4857ab
build_units.py c052246c828024a25fc6a143d513e111ecbf63b42b8b5912dd37e960a1c3e243
collect_onboard.py c93fa59c8fe305459336e08e52c60f365d8df0a70d7b6ca395c58e8f88da5fff
compose.py 296001923384652d7d31cdb33ca937ed409783023de2b6ee53c62e4ea8ab5bd1
frontend.py dd4a0fdc1e3595eeefb82a8eafab600255b3b0fe95f67e2fb8a07eb5972ea0d8
gold.py 6b0afaad2a3ae6652a02207813c5a3492ba6954bd0a745b58a3f98c960bb321e
jev/main.go 63fbec0534a2ddf3d55d07c074c6a882e199ec2b90cf771a984c9fa9ee586ca0
label_x2d.py fa5d60ef2bbf2147102cf986048f7711cd2ad90d8cd7a0ddbe89aa5162729dcb
measurement/lever-result.json 298902e717d0db8826c35d219695d386b19644f99cfc205d6b2d891100b0352c
measurement/lever_check.py 8f8008d3a28d72d5e768bc292e4917449fd8cb31ca0a642ff6057ab51b705a42
onboard/main.go 7e2ed1edc64b694851e7729e17af4f2ba63dafa7c9cdcf19cd3a5a255c08d46b
power_sim.py a955f7177280bfd6d20a2514d21a04819c426610d594ee4e7dbfe0cdfa9c5279
prepass.py 1e5863d32974f1d7fecfd4df5c30b6e02aa137b9db491780eab11cc00b500696
prepass/commits.jsonl 1581986e0efbf26bb51aea894c3ffcdd23f54ef6e1d6d6ab0e4e99248f35bd1c
prepass/k0.json f89e8d0b264eb33eb8bcb5aa4a4910bd462c3e1befea480b721ee514947a98ab
prepass/reserve.json a73558ed54efdd767c5c0cf9a5accdcf41f511acd044d630208a8fb3e129f12a
prepass/rounds.json 0aad85b0199e747630d0e863b7a5e83b2624684c50c1150ee4e8775fde1d7516
prepass/scopes.json a24bfc8d19aab9fd5ab0817f317cee560df65c63b1d63b5dbd10b5529d16a512
prepass/summary.json e1cfced0da45bcd86cc516eef39fa70c10b3fa5d630820d415cdea47045cc1c1
quality.py 6d65293d64c76611368fa8eaf93d4f48432cf203bb65aed9ed63130b8527cade
questions/merge.area_covers@v1.json cce59e612b67d76b96d927613c1a780ea64915c38210b7be8fceb5cfaa7e5af0
questions/merge.covered@v1.json f37eea94bb6e2c720cbbb904066dd06c368ecc7581c90063b9f17e5a751058b9
questions/merge.which_area@v1.json cbb7ba7e92432a1e0ffc9bf7d8718a9001ba4227137a6ab60a1e359393eff928
questions/onboard.covered@v1.json ec4d62be051c016ab945738c155401a17f091654e1efff9cfd60ec18314eaa70
questions/onboard.same_responsibility@v1.json 31f3b25d3bbf535b026bd6f0a022d39f01e7c519cc22649195d341b75a49a7b1
questions/onboard.which_file_area@v1.json 425edf93d08dba42d7f1da755601d5dec27cded8b67593401b7656b24cf12f57
reference_arm.py d139f0179c06ca5c6aff5784ceab2feb4113f888c51cd46f760f7a3d943bb739
run.sh 31029ba557d453e620b214f242ed6454622617e34f0850fcd7a0667ac82806c5
scopes.py 5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072
thresholds.json 7142d7ac15abbcd27b75ac8005a7f1bffed0c6a70395c65e8f631bfcd32995eb   <- CHANGED (round-11 digest 5c211eb939b4880ae21e79fdfcfc25060722f73a96b99d6f27aaceb2153094a8)
tune.py 2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c
validation/r4_checks.py f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52
validation/stub_clawpatch_x2d.js d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f
validation/stub_jev.py f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d
validation/stub_labeller.py d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5
validation/validate.py d9f77f0db1ef8733f3d4902c81ca4ac9762026d94c0836572b2280facdf08128   <- CHANGED (round-11 digest b9fc5de8d47a5a45077953bf72568ebdc30451e91eaed3a458c35e2910e08396)
versioning.py 331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168
x2d.py 62e8cc1de402e221d6a57d04f9ebfc50c9e7bf4875ac5f3d65709830a60665ca   <- CHANGED (round-11 digest 32c093e972cff31e86dae51412e948c70f07348465d299e18036813f82c8d464)
```

## 4. External pin

`~/.omp/agent/skills/jev-actions/scripts/tune_threshold.py` SHA-256 `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c` = `signoff.json` `tune_script_sha256` (equal; the file is unchanged).

## Main-agent addendum (2026-10-30, supersedes the digest table and validation above where they differ)

- `run.sh:248` comment: "charged to the $10 total" → "charged to the total cap (`x2d.CAPS['usd_total']`, $20 since rev 12)". Comment only; no behaviour change.
- Validation (a) on the final bytes, `signoff.json` moved aside (SHA-256 `e7059f2a9407e38c4d5e8aff0480913fab36a32a69ccbb9721c9633934fdbc2d` before and after, equal): `651 checks, 0 failed` / `all validation checks passed` (`/tmp/x2d-validate-r12a.log`). Run (b) requires the new record and is run after round 12, per the round-11 Addendum.
- Not changed (unused defaults; `run.sh` always passes `-max-cost-usd` from `x2d.jev_budget`): `jev/main.go:736` and `onboard/main.go` `-max-cost-usd` defaults of 1.
- Council composition from round 12: lead Opus 5.5 high (was Fable 5.1 high in the live config), challenger GPT-6 Astra high, experiment peer Opus 5.5 high (owner decision 2026-10-30).

### Final digests (44 frozen files, after the run above)

| file | sha256 |
|---|---|
| `GUIDE.md` | `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980` |
| `PREREGISTRATION.md` | `668110fffc52414c490becd6543d9712f87e84a46db4968c81a74445e1e32c31` |
| `README.md` | `fc51642bf03d55847a2086590932ee5bf44e64e8847b2cdc8eaaadd36e5d6124` |
| `analysis_x2d.py` | `b0a66eedb739c2952b14f8468ae56da65d9628f6fa345f8dc51380520b397ade` |
| `b2.py` | `5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f` |
| `baselines.py` | `a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0` |
| `build_controls.py` | `a71c4e634b5a0d0b42eda15f8e58220fef159c077fd3bf22d36041dd3e4857ab` |
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
| `prepass/commits.jsonl` | `1581986e0efbf26bb51aea894c3ffcdd23f54ef6e1d6d6ab0e4e99248f35bd1c` |
| `prepass/k0.json` | `f89e8d0b264eb33eb8bcb5aa4a4910bd462c3e1befea480b721ee514947a98ab` |
| `prepass/reserve.json` | `a73558ed54efdd767c5c0cf9a5accdcf41f511acd044d630208a8fb3e129f12a` |
| `prepass/rounds.json` | `0aad85b0199e747630d0e863b7a5e83b2624684c50c1150ee4e8775fde1d7516` |
| `prepass/scopes.json` | `a24bfc8d19aab9fd5ab0817f317cee560df65c63b1d63b5dbd10b5529d16a512` |
| `prepass/summary.json` | `e1cfced0da45bcd86cc516eef39fa70c10b3fa5d630820d415cdea47045cc1c1` |
| `prepass.py` | `1e5863d32974f1d7fecfd4df5c30b6e02aa137b9db491780eab11cc00b500696` |
| `quality.py` | `6d65293d64c76611368fa8eaf93d4f48432cf203bb65aed9ed63130b8527cade` |
| `questions/merge.area_covers@v1.json` | `cce59e612b67d76b96d927613c1a780ea64915c38210b7be8fceb5cfaa7e5af0` |
| `questions/merge.covered@v1.json` | `f37eea94bb6e2c720cbbb904066dd06c368ecc7581c90063b9f17e5a751058b9` |
| `questions/merge.which_area@v1.json` | `cbb7ba7e92432a1e0ffc9bf7d8718a9001ba4227137a6ab60a1e359393eff928` |
| `questions/onboard.covered@v1.json` | `ec4d62be051c016ab945738c155401a17f091654e1efff9cfd60ec18314eaa70` |
| `questions/onboard.same_responsibility@v1.json` | `31f3b25d3bbf535b026bd6f0a022d39f01e7c519cc22649195d341b75a49a7b1` |
| `questions/onboard.which_file_area@v1.json` | `425edf93d08dba42d7f1da755601d5dec27cded8b67593401b7656b24cf12f57` |
| `reference_arm.py` | `d139f0179c06ca5c6aff5784ceab2feb4113f888c51cd46f760f7a3d943bb739` |
| `run.sh` | `f48cbf01a9e69c67f58c98e3c78880db814f7e28452587d0981176385d937595` |
| `scopes.py` | `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072` |
| `thresholds.json` | `7142d7ac15abbcd27b75ac8005a7f1bffed0c6a70395c65e8f631bfcd32995eb` |
| `tune.py` | `2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c` |
| `validation/r4_checks.py` | `f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52` |
| `validation/stub_clawpatch_x2d.js` | `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f` |
| `validation/stub_jev.py` | `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d` |
| `validation/stub_labeller.py` | `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5` |
| `validation/validate.py` | `d9f77f0db1ef8733f3d4902c81ca4ac9762026d94c0836572b2280facdf08128` |
| `versioning.py` | `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168` |
| `x2d.py` | `62e8cc1de402e221d6a57d04f9ebfc50c9e7bf4875ac5f3d65709830a60665ca` |

Tune pin: `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c`.
