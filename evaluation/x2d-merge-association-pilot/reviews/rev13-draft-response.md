# Rev 13 draft response (R-053, option (c): one re-gate of P1 and P2 on fresh items)

Drafter's response for council round 17. No model call was made, nothing was written under `../scout-x2d`, no commit, `signoff.json` untouched (SHA-256 `578a2462a6f25953c2e744a1d3dcc546f58af868f2e0e343e204367a89552e61` before and after every run; it was moved aside only for dev runs and validation, then restored). `signoff.json` is stale against the edited frozen files until the council signs off.

## 1. Fresh supply (membership only; `reviews/regate-supply.md`, `.json`)

| origin | clusters | pool |
|---|---|---|
| D (remaining eligible D units + PR episodes) | 20 | 3 |
| G (non-guide clusters) | 59 | 7 |
| P | 30 | 0 |
| in_no_unit (C0-owned, in no unit) | 553 | 119 |
| **total** | | **129** (129 PRs, 69 areas) |

P1 candidates that can host stratum 1 / 2 / 3: 126 / 60 / 89. Planned P1: 40 items, strata 34 / 2 / 4 (minimum 40, stratum 1 >= 20: met). Planned P2: 40 clones in 16 areas (minimum 8 positives; P(Binomial(40, 1/3) >= 8) = 0.98, 0.82 at 1/4). Both minimums met; no stop. Command: in `regate-supply.md`.

## 2. Changed files (`file:line`, one line each)

`x2d.py`
- 44-57: `REGATE_*` constants (root name, gate-set tag, P1 N/strata/stratum-1 min/bars, baseline bar 0.20, P2 clones/areas/positives min/none bar/abstention cap), all from existing bars.
- 245-259: `regate_root`, `is_regate_root`, `main_data` (a regate root maps to the study data directory).
- 262: `gate_problems` (round 1 needs regate PASS for BOTH P1 and P2 when either failed in round 0; any other FAIL stops).
- 295: `regate_frozen_problems` (re-gate root's Jev side equals round 0, thresholds 0.70).
- 609: `jev_budget` maps the regate root to the study data, so the re-gate spends from the round-0 $2 cap and the $20 total.

`build_controls.py`
- 213-262: `_owned_dirs`, `_candidate_dirs`, `_p1_target(..., leak_check, taken, own_dirs)` (ctl-23 leak fix), `_p1_row` (shared P1 row builder).
- 425-536: `_guide_prs`, `exposed`, `_add_commits`, `fresh_pool` (hash-order, membership-only fresh pool excluding used sources, path twins, PR episodes, guide PRs).
- 537-683: `_p1_placements`, `regate_plan`, `make_regate` (stratum mix, P2 sizing, refusal on a short supply).
- 684-744: `regate_precondition`, `write_regate` (one shot, writes only `rounds/0/regate/`), `_print_supply`.
- 33-35, 754-755: `--regate-supply [--out]`, `--regate` flags; the legacy build refuses inside a regate root; root-commit parent bug fixed.

`gold.py`
- 45: `build_gold` keeps a clone's own agreed label when `gold_rule: "direct_label"` and attaches `derived_gold`.
- 319: `required_gold_ids` covers set F and the direct-label clones.

`analysis_x2d.py`
- 1407: `_p1_stats` (accuracy on answered, abstention, all-item recovery, path-following rates).
- 1421: `regate_gates` and `--regate-gates` (gate set `x2d-regate-v1`, writes `regate/gates.json` once; `--gates` refuses inside a regate root).

`run.sh`
- 67: `require_gates` uses `x2d.gate_problems`.
- 339-376: `in_regate`, `regate_ready`, `regate_open`, `stage_regate_{build,label,jev,analyse}` (one shot, frozen-side check at each stage).
- 396-397: `regate` and `regate-*` stage dispatch.

`validation/validate.py`: 16-17 docstring G13, 492 import, 3805 `regate_checks.unit_checks`, 3819 `regate_checks.e2e_regate`.
`validation/regate_checks.py` (new): `p1_checks` 116, `p2_checks` 179, `gold_checks` 229, `leak_checks` 273, `gate_problem_checks` 317, `run_refusal_checks` 352, `budget_checks` 419, `unit_checks` 440, `pool_checks` 474, `build_checks` 550, `e2e_regate` 644.

`PREREGISTRATION.md`: status line 3; change log rev 12 -> 13 at 15; ledger line 17 (R-053); §5 bullets 108-114 and stop rule 116; §11 line 195; §12 line 200; §14 item 20 at 251 and HD6 map clause at 253. `README.md`, `DRAFT-NOTES.md`: matched (drafted by a subagent; I spot-checked bars, budget, HD6 clause and fixture list against the code, not every sentence).

## 3. New fixtures (G13, 56 checks)

Unit (36): P1 planted effect PASS / path-following planted null FAIL / sibling B acceptable is not path-following / accuracy bound 0.80 / abstention bound 0.15 / baseline bar 0.20 / exclusions and UNMEASURED below 20 / missing V-sem arm row refuses; P2 planted effect, planted null, bounds, UNMEASURED below 8, path-twin exclusion, retained-covered report-only, rev-12 gates.json untouched; gold direct vs derived (direct label wins, rev-12 rule unchanged, `required_gold_ids`, no label no gold); leak fix (4); `gate_problems` (3); budget (3); `is_regate_root`/`main_data`; run.sh refusals (round 1 without PASS, build twice, label/jev/analyse after gates.json, frozen Jev side, missing regate.json, build without a P1/P2 FAIL, build after another gate FAIL).
End to end (20): fresh pool on the real clone (3), short-supply refusal, build and its content, leak fix on the built items, frozen Jev side copy, supply recorded, existing bytes unchanged after the build, label with its own cap, P2 gold direct vs derived, P1 gold, Jev stage, planted effect PASS end to end, round 1 after PASS, one-shot (manifest SHA-256 unchanged after four refusals), study data differs only by new files under `regate/`, planted null FAIL end to end.

## 4. Validation

`validate.py` with `signoff.json` moved aside: **707 checks, 0 failed**, "all validation checks passed" (651 before). Log: `/tmp/validate-rev13.log`. `signoff.json` SHA-256 before = after = `578a2462a6f25953c2e744a1d3dcc546f58af868f2e0e343e204367a89552e61`.

## 5. Frozen digests (45 files; `x2d.frozen_files()` also returns 45), after validation

| file | sha256 |
|---|---|
| `GUIDE.md` | `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980` |
| `PREREGISTRATION.md` | `d4cdad866ccbe2f15ae5de7fda3aa885130389ddf9e22cf362bab89c0f8d6751` |
| `README.md` | `31f9dded37431713ef92780a646ec71ac05e6879d568da4539588c4c61f8f456` |
| `analysis_x2d.py` | `37fe3712abfda315dd665db03d67509971ae251a36582f964fa3617c48d1e8a4` |
| `b2.py` | `5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f` |
| `baselines.py` | `a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0` |
| `build_controls.py` | `efbcbf0636e2b157ee60af21d48c3b1feb845cfd84cb0914c0eda390f4e67990` |
| `build_units.py` | `c052246c828024a25fc6a143d513e111ecbf63b42b8b5912dd37e960a1c3e243` |
| `collect_onboard.py` | `c93fa59c8fe305459336e08e52c60f365d8df0a70d7b6ca395c58e8f88da5fff` |
| `compose.py` | `296001923384652d7d31cdb33ca937ed409783023de2b6ee53c62e4ea8ab5bd1` |
| `frontend.py` | `dd4a0fdc1e3595eeefb82a8eafab600255b3b0fe95f67e2fb8a07eb5972ea0d8` |
| `gold.py` | `a7831d65d40c0fa9ee15555a8488532dc646990725ac47a670f1fb6fec6bb8e5` |
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
| `run.sh` | `99eb7422022241e1836fa3a9c6dbc84a8976d5f8e00fefaddcb78e1f1079c754` |
| `scopes.py` | `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072` |
| `thresholds.json` | `7142d7ac15abbcd27b75ac8005a7f1bffed0c6a70395c65e8f631bfcd32995eb` |
| `tune.py` | `2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c` |
| `validation/r4_checks.py` | `f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52` |
| `validation/regate_checks.py` | `e87f99f677dc149d2f798a5eac43f953bf69fc6d5acbc57fed36a7524c737083` |
| `validation/stub_clawpatch_x2d.js` | `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f` |
| `validation/stub_jev.py` | `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d` |
| `validation/stub_labeller.py` | `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5` |
| `validation/validate.py` | `8fab5ee017b8eecd8b0dc13dacb45edf9d7383e8586334ec7f994bf764ee16c9` |
| `versioning.py` | `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168` |
| `x2d.py` | `0ae78958c2d39e424cbc417c8c4250db0643934658f11a82146f663fac394b00` |

## 6. Open questions for the council

1. P1 bars: stratum-1 accuracy on answered items >= 0.80 (R_assign) and abstention <= 0.15 (g), B1/B2 <= 0.20 over all eligible items. Together they imply at least 0.68 all-item recovery, never easier than rev 12's 0.60. Rule on them.
2. P2 stays binding (memo §5)? Bars: none rate on answered direct-label positives >= 0.80 (`NEW_AREA_GATE_MIN`), abstention <= 0.15; UNMEASURED below 8 positives does not pass.
3. Path-following is report-only (no bar). Acceptable?
4. Strata 2-3 are small on the real supply (planned 34 / 2 / 4): reported, not decisive. Acceptable?
5. g = 0.15 with the frozen provisional 0.70 thresholds is a knife-edge for P1 abstention; a fail is a one-shot stop.
6. The pool is mostly `in_no_unit` clusters (119 of 129), some pre-2023; the reserve sets give only 10.
7. Fresh P2 sources are labelled on the full catalogue (set F) to report the derived gold; this adds 40 labelled units (W = 120, cap 330).
8. P2 yield assumption 1/3 from 4 positives in 12 clones; the real yield on fresh clones is unknown [INFERENCE].
