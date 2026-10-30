# X2d — drafter's response to council round 6

Date: 2026-10-30. Review: `reviews/round-6.md` (7 minor issues, 0 blocking, 0 major). Every fix is the council's exact fix. No product file, `GUIDE.md`, `questions/*.json`, `power_sim.py`, `compose.py`, `tune.py`, `x2d.py`, `r4_checks.py` or the research ledger changed. PREREGISTRATION is now revision 7 (status line, change log rev 6 → rev 7).

## 1. Per-issue table

| id | issue | fix applied |
|---|---|---|
| R6-01 | HD6 / §1 overclaim what h shows | `PREREGISTRATION.md` §1 clause and §14 state map reworded as prescribed; `docs/hypothesis-ledger.md` HD6 clause reworded and the History line appended (state stays `proposed`; HD6b unchanged) |
| R6-02 | §11 misdescribes N4a | `PREREGISTRATION.md` §11: N4a abstains on B2's errors; the `gain_res` inflation is the own-error family N4b, N4b-cap, N4b-ρ0 |
| R6-03 | §7 raw fields and Noul boundary | `PREREGISTRATION.md` §7: raw fields read by the perception check and the guard h; `none` iff `1 − p_true ≥ s` (equality is `none`) |
| R6-04 | Recovery recipe names wrong bytes | recipe replaced in `README.md` (stage 5b-i), `PREREGISTRATION.md` §10 (v), `run.sh` comment; no code change |
| R6-05 | Planted P_new / R_new count unmeasured traps | `analysis_x2d.py` `controls_section`: `nc[1]` and `nc[2]` count only traps with gold and a measured arm row; `_planted_new_from_rounds` docstring; fixture (`validate.py`) now `none_correct {1, 2}`, `none_n` 2, pooled R_new (2, 4), P_new (2, 4), `unmeasured` 2, t3 only in `unmeasured` (same single check, count unchanged); §9 and `thresholds.json` `new_area` text |
| R6-06 | Stale or inexact text (a)–(m) | (a) `thresholds.json` intervals; (b), (c) `validate.py` labels and the e2e assertion `pooled["gain_guess"]["lb95"] > 0`; (d) `gain_all_stat` docstring; (e) `verdict()` message; (f) §10 (iv); (g) §11 `_require_complete`; (h) §11 honest-10 % world; (i) §11 N4a / N4b rows; (j) §11 "about 7 %" and the guard sentence; (k) README header, Modules row, run-commit wording; (l) `label_x2d.py` help; (m) DRAFT-NOTES run-log name (`-r6c`, as in round-5-response) and §10 (4) restated |
| R6-07 | Full measured power table | table below, read from `power_sim.run(pilots=1000, seed=x2d.SEED, boot_B=1000)`; no frozen file changed for this item |

## 2. Validation

`source .tools/activate && evaluation/x2d-merge-association-pilot/run.sh validate` once after the last study-file edit: exit 0, **647 checks, 0 failed** (log `/tmp/x2d-validate-r7a.log`; same count as round 5, no check added or removed). Every binding power bar holds (P3_u0 0.830, P3_u05 0.876, P3_u10 0.806; N3 max 0.023; N3_PR max 0.044; N4a 0.000; N4b 0.047; N4b_cap 0.061; N4b_rho0 0.055). The R5-07 fixture asserts (2, 4) / (2, 4).

## 3. Digests (SHA-256, computed after the validation run; `x2d.frozen_files()`, 44 files)

Changed from round 5: `PREREGISTRATION.md`, `README.md`, `analysis_x2d.py`, `label_x2d.py`, `run.sh`, `thresholds.json`, `validation/validate.py`. All others equal round 5.

| File | SHA-256 |
|---|---|
| `GUIDE.md` | `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980` |
| `PREREGISTRATION.md` | `64504b052389bfc20961b19703decb0d134e4af91c7428dd7dd6e9732ba7f6b8` |
| `README.md` | `cacb70e07e03dda10a8fdd00fbcb924321060032c0cd5c50f27aa8c1425873bd` |
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
| `thresholds.json` | `7ca8e255402ece302dbb60f074592276c096852922b72d257818a5f569506a0b` |
| `tune.py` | `2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c` |
| `validation/r4_checks.py` | `f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52` |
| `validation/stub_clawpatch_x2d.js` | `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f` |
| `validation/stub_jev.py` | `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d` |
| `validation/stub_labeller.py` | `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5` |
| `validation/validate.py` | `131a45ccc5571062430f260385a9bafb1d7b1e2610a7b3248f4e9633bb9c381c` |
| `versioning.py` | `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168` |
| `x2d.py` | `32c093e972cff31e86dae51412e948c70f07348465d299e18036813f82c8d464` |

`tune_threshold.py` (`~/.omp/agent/skills/jev-actions/scripts/tune_threshold.py`): `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c` (recomputed after the validation run; equals round 6's)

`DRAFT-NOTES.md` (not a frozen file): `cdacfbe551057320a8b9f81360be17e0f554a33059323c68ca10c655321b23ae`.

## 4. Full measured power table (R6-07)

Every cell of `power_sim.run(pilots=1000, seed=x2d.SEED, boot_B=1000)` (empirical 91 clusters / 40 PRs structure, 10 positives inside, deterministic), read from the returned dict (`all_pass` true; saved at `/tmp/x2d-power-r7.json`). GO count / 1000, Wilson 95 %, and the three verdict counts (GO = the GO column).

| cell | p_J | p_B2 | ρ | PR w | u | mode | GO | share | Wilson 95 % | INCONCLUSIVE | NO-GO |
|---|---|---|---|---|---|---|---|---|---|---|---|
| P3_u0 | 0.95 | 0.8 | 0.5 | 0.0 | 0.0 | random | 830 / 1000 | 0.830 | [0.805, 0.852] | 0 | 170 |
| P3_u05 | 0.95 | 0.8 | 0.5 | 0.0 | 0.05 | random | 876 / 1000 | 0.876 | [0.854, 0.895] | 0 | 124 |
| P3_u10 | 0.95 | 0.8 | 0.5 | 0.0 | 0.1 | random | 806 / 1000 | 0.806 | [0.780, 0.829] | 1 | 193 |
| N3 | 0.5 | 0.5 | 0.0 | 0.0 | 0.0 | random | 0 / 1000 | 0.000 | [0.000, 0.004] | 0 | 1000 |
| N3 | 0.5 | 0.5 | 0.0 | 0.0 | 0.1 | random | 0 / 1000 | 0.000 | [0.000, 0.004] | 0 | 1000 |
| N3 | 0.6 | 0.6 | 0.0 | 0.0 | 0.0 | random | 0 / 1000 | 0.000 | [0.000, 0.004] | 0 | 1000 |
| N3 | 0.6 | 0.6 | 0.0 | 0.0 | 0.1 | random | 0 / 1000 | 0.000 | [0.000, 0.004] | 0 | 1000 |
| N3 | 0.7 | 0.7 | 0.0 | 0.0 | 0.0 | random | 5 / 1000 | 0.005 | [0.002, 0.012] | 1 | 994 |
| N3 | 0.7 | 0.7 | 0.0 | 0.0 | 0.1 | random | 0 / 1000 | 0.000 | [0.000, 0.004] | 1 | 999 |
| N3 | 0.8 | 0.8 | 0.0 | 0.0 | 0.0 | random | 21 / 1000 | 0.021 | [0.014, 0.032] | 2 | 977 |
| N3 | 0.8 | 0.8 | 0.0 | 0.0 | 0.1 | random | 19 / 1000 | 0.019 | [0.012, 0.029] | 7 | 974 |
| N3 | 0.85 | 0.85 | 0.0 | 0.0 | 0.0 | random | 13 / 1000 | 0.013 | [0.008, 0.022] | 0 | 987 |
| N3 | 0.85 | 0.85 | 0.0 | 0.0 | 0.1 | random | 23 / 1000 | 0.023 | [0.015, 0.034] | 8 | 969 |
| N3 | 0.9 | 0.9 | 0.0 | 0.0 | 0.0 | random | 11 / 1000 | 0.011 | [0.006, 0.020] | 0 | 989 |
| N3 | 0.9 | 0.9 | 0.0 | 0.0 | 0.1 | random | 11 / 1000 | 0.011 | [0.006, 0.020] | 2 | 987 |
| N3 | 0.95 | 0.95 | 0.0 | 0.0 | 0.0 | random | 1 / 1000 | 0.001 | [0.000, 0.006] | 0 | 999 |
| N3 | 0.95 | 0.95 | 0.0 | 0.0 | 0.1 | random | 2 / 1000 | 0.002 | [0.001, 0.007] | 0 | 998 |
| N3_PR | 0.5 | 0.5 | 0.0 | 0.3 | 0.0 | random | 0 / 1000 | 0.000 | [0.000, 0.004] | 0 | 1000 |
| N3_PR | 0.5 | 0.5 | 0.0 | 0.3 | 0.1 | random | 0 / 1000 | 0.000 | [0.000, 0.004] | 0 | 1000 |
| N3_PR | 0.6 | 0.6 | 0.0 | 0.3 | 0.0 | random | 0 / 1000 | 0.000 | [0.000, 0.004] | 0 | 1000 |
| N3_PR | 0.6 | 0.6 | 0.0 | 0.3 | 0.1 | random | 0 / 1000 | 0.000 | [0.000, 0.004] | 0 | 1000 |
| N3_PR | 0.7 | 0.7 | 0.0 | 0.3 | 0.0 | random | 6 / 1000 | 0.006 | [0.003, 0.013] | 2 | 992 |
| N3_PR | 0.7 | 0.7 | 0.0 | 0.3 | 0.1 | random | 3 / 1000 | 0.003 | [0.001, 0.009] | 4 | 993 |
| N3_PR | 0.8 | 0.8 | 0.0 | 0.3 | 0.0 | random | 34 / 1000 | 0.034 | [0.024, 0.047] | 9 | 957 |
| N3_PR | 0.8 | 0.8 | 0.0 | 0.3 | 0.1 | random | 31 / 1000 | 0.031 | [0.022, 0.044] | 21 | 948 |
| N3_PR | 0.85 | 0.85 | 0.0 | 0.3 | 0.0 | random | 44 / 1000 | 0.044 | [0.033, 0.059] | 17 | 939 |
| N3_PR | 0.85 | 0.85 | 0.0 | 0.3 | 0.1 | random | 39 / 1000 | 0.039 | [0.029, 0.053] | 20 | 941 |
| N3_PR | 0.9 | 0.9 | 0.0 | 0.3 | 0.0 | random | 23 / 1000 | 0.023 | [0.015, 0.034] | 5 | 972 |
| N3_PR | 0.9 | 0.9 | 0.0 | 0.3 | 0.1 | random | 17 / 1000 | 0.017 | [0.011, 0.027] | 12 | 971 |
| N3_PR | 0.95 | 0.95 | 0.0 | 0.3 | 0.0 | random | 3 / 1000 | 0.003 | [0.001, 0.009] | 2 | 995 |
| N3_PR | 0.95 | 0.95 | 0.0 | 0.3 | 0.1 | random | 7 / 1000 | 0.007 | [0.003, 0.014] | 0 | 993 |
| N4a | 0.8 | 0.8 | 0.5 | 0.0 | 0.0 | b2_errors | 0 / 1000 | 0.000 | [0.000, 0.004] | 0 | 1000 |
| N4a | 0.85 | 0.85 | 0.5 | 0.0 | 0.0 | b2_errors | 0 / 1000 | 0.000 | [0.000, 0.004] | 0 | 1000 |
| N4a | 0.9 | 0.9 | 0.5 | 0.0 | 0.0 | b2_errors | 0 / 1000 | 0.000 | [0.000, 0.004] | 0 | 1000 |
| N4a | 0.95 | 0.95 | 0.5 | 0.0 | 0.0 | b2_errors | 0 / 1000 | 0.000 | [0.000, 0.004] | 0 | 1000 |
| N4b | 0.85 | 0.85 | 0.5 | 0.0 | 0.0 | own_errors | 47 / 1000 | 0.047 | [0.036, 0.062] | 254 | 699 |
| N4b | 0.88 | 0.88 | 0.5 | 0.0 | 0.0 | own_errors | 43 / 1000 | 0.043 | [0.032, 0.057] | 232 | 725 |
| N4b | 0.9 | 0.9 | 0.5 | 0.0 | 0.0 | own_errors | 26 / 1000 | 0.026 | [0.018, 0.038] | 136 | 838 |
| N4b | 0.95 | 0.95 | 0.5 | 0.0 | 0.0 | own_errors | 4 / 1000 | 0.004 | [0.002, 0.010] | 5 | 991 |
| N4b_cap | 0.8 | 0.8 | 0.5 | 0.0 | 0.0 | own_errors_cap | 61 / 1000 | 0.061 | [0.048, 0.078] | 400 | 539 |
| N4b_cap | 0.85 | 0.85 | 0.5 | 0.0 | 0.0 | own_errors_cap | 51 / 1000 | 0.051 | [0.039, 0.066] | 397 | 552 |
| N4b_rho0 | 0.85 | 0.85 | 0.0 | 0.0 | 0.0 | own_errors | 55 / 1000 | 0.055 | [0.042, 0.071] | 407 | 538 |
| N4b_rho0 | 0.88 | 0.88 | 0.0 | 0.0 | 0.0 | own_errors | 54 / 1000 | 0.054 | [0.042, 0.070] | 515 | 431 |
| N4b_rho0 | 0.9 | 0.9 | 0.0 | 0.0 | 0.0 | own_errors | 44 / 1000 | 0.044 | [0.033, 0.059] | 409 | 547 |
| N4c | 0.9 | 0.8 | 0.5 | 0.0 | 0.0 | own_errors | 661 / 1000 | 0.661 | [0.631, 0.690] | 220 | 119 |
| P3_PR | 0.95 | 0.8 | 0.5 | 0.3 | 0.05 | random | 812 / 1000 | 0.812 | [0.787, 0.835] | 12 | 176 |
| P3_rho0 | 0.95 | 0.8 | 0.0 | 0.0 | 0.05 | random | 853 / 1000 | 0.853 | [0.830, 0.874] | 3 | 144 |
| P3_curve_delta_0.15 | 0.85 | 0.7 | 0.5 | 0.0 | 0.05 | random | 479 / 1000 | 0.479 | [0.448, 0.510] | 2 | 519 |
| P3_curve_delta_0.15 | 0.9 | 0.75 | 0.5 | 0.0 | 0.05 | random | 787 / 1000 | 0.787 | [0.761, 0.811] | 5 | 208 |
| P3_curve_delta_0.15 | 0.95 | 0.8 | 0.5 | 0.0 | 0.05 | random | 876 / 1000 | 0.876 | [0.854, 0.895] | 0 | 124 |
| P3_curve_delta_0.15 | 1.0 | 0.85 | 0.5 | 0.0 | 0.05 | random | 910 / 1000 | 0.910 | [0.891, 0.926] | 0 | 90 |
