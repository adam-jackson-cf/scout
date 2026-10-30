# X2d — response to council round 8 (main agent, 2026-10-30)

| id | status | change |
|---|---|---|
| R8-01 | fixed (council text verbatim) | `PREREGISTRATION.md` §11: both replacements applied; status line → Revision 9; change-log line rev 8 → rev 9 (R8-01; R8-02 as validation-only). `README.md:3` → "(revision 9, after council round 8 `reviews/round-8.md`, draft for council round 9)". |
| R8-02 | fixed (council text verbatim) | `validation/validate.py`: second stub `stub2` (seed `j6b`) started beside `stub`; the re-run targets `stub2`; condition `stub2.requests == 0`; detail tuple `(r.returncode, r2.returncode, n_after, total, spent1, stub2.requests, r2.stderr[-200:])`; `finally` stops both; `_stub_quiescent` deleted; dead `n * 0 +` term removed. Check name and count unchanged. |

## Validation (one run after the last edit)

- Command: `source .tools/activate && evaluation/x2d-merge-association-pilot/run.sh validate`; rc 0.
- Log, untrimmed: `/tmp/x2d-validate-r9.log` (712 lines). Last lines: `647 checks, 0 failed` / `all validation checks passed`.
- Environment: `run.sh` runs `.tools/venv/bin/python` (Python 3.11.11, `run.sh:24`); `go version go1.25.6 darwin/arm64`. (The shell's own `python3` is not used by `run.sh`.)
- The J check: `PASS J the cumulative cost stop: a run whose billed cost exceeds -max-cost-usd stops early with exit 4 (fewer answers than requests) and a re-run refuses to sta…`

## Digests (SHA-256, computed after the validation run)

| file | sha256 |
|---|---|
| `DRAFT-NOTES.md` | `3be7a0d277a36009e41da543d315cf0b47dcacd0dffb26c98ca87b9c531209c0` |
| `GUIDE.md` | `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980` |
| `PREREGISTRATION.md` | `4b14dae0c3875eddd98667bf6da3935785f4817b209dd5d78145bbc0dca38429` |
| `README.md` | `e69c713c7ef486d5023460a0ed2dbf4d1881982c5cc11fbf39962fc38019f1c5` |
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
| `run.sh` | `31029ba557d453e620b214f242ed6454622617e34f0850fcd7a0667ac82806c5` |
| `scopes.py` | `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072` |
| `thresholds.json` | `5c211eb939b4880ae21e79fdfcfc25060722f73a96b99d6f27aaceb2153094a8` |
| `tune.py` | `2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c` |
| `validation/r4_checks.py` | `f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52` |
| `validation/stub_clawpatch_x2d.js` | `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f` |
| `validation/stub_jev.py` | `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d` |
| `validation/stub_labeller.py` | `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5` |
| `validation/validate.py` | `b63ca4e7e606e7284ad9f565c02655554cfe81873950ae8227ef44946ca4c42b` |
| `versioning.py` | `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168` |
| `x2d.py` | `32c093e972cff31e86dae51412e948c70f07348465d299e18036813f82c8d464` |

`tune_threshold.py` pin: `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c`.
