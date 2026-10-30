# Round 17 response (rev 13 re-gate, R-053): the drafter's answer to `reviews/round-17.md`

All seven issues applied as worded in `round-17.md` §2; the §3 rulings are followed. No model call was made, nothing was written under `../scout-x2d`, no commit, `signoff.json` untouched (SHA-256 `578a2462a6f25953c2e744a1d3dcc546f58af868f2e0e343e204367a89552e61` before and after every run; it was moved aside only for dev runs, validation and the supply run, then restored). `signoff.json` is stale against the edited frozen files until the council signs off.

## 1. Per-issue

| Issue | file:line | What was done (exact text where the council gave it) |
|---|---|---|
| R17-01 (1) | `build_controls.py:461-466` | `exposed()` adds `"presented"` = the members of every round-0 `units.jsonl` row with `control: true` and `"target_dirs"` = `{k["target_dir"] for k in control-key rows with kind "p1"}`. |
| R17-01 (2) | `build_controls.py:223-244,259-262,555-569,589-590,674` | `_candidate_dirs(..., forbid_dirs, rej)` skips a forbidden directory for every stratum; `_p1_target(..., forbid_paths, forbid_dirs, rej)` skips a directory when any new path is in `forbid_paths`; `_p1_placements(..., forbid_paths, forbid_dirs, rej)` starts `taken = set(forbid_dirs)`; `regate_plan` and `make_regate` pass `exposed()["presented"]` and `["target_dirs"]`. |
| R17-01 (3) | `build_controls.py` `regate_plan` supply dict (`"exposed": dict(info["exposed"], presented_paths, rev12_p1_target_dirs, p1_rejected_for_presented_path, p1_directory_candidates_rejected_as_rev12_target)`) | Counts recorded in `supply["exposed"]` (the final trial is re-run once with a counter). The funnel is unchanged. |
| R17-01 (4) | `PREREGISTRATION.md:111` | After "the PRs named (`#<digits>`) in the guide;" inserted: "a P1 placement whose presented path equals that of any round-0 control, or whose target directory was the target directory of any rev-12 P1 item (the rev-12 target directories seed the never-reuse set);". |
| R17-01 (5) | `validation/regate_checks.py:308` (`_p1_target` fixture), `:610` (plan fixture on the real clone) | `_p1_target` is given the only stratum-1 host as a forbidden dir and as a forbidden path; the plan fixture plants a rev-12 control at a stratum-1 placement in a data copy and the plan places no item there. |
| R17-01 (6) | `reviews/regate-supply.{md,json}` | Regenerated (section 2). Minimums met, no stop. |
| R17-02 (1) | `analysis_x2d.py:1427,1484,1749,1781-1786` | `_regate_denominators(d)` (reads only units, control-key, gold, labels presence check, `arms/B1`, `arms/B2`), `regate_label_problems` and the CLI flag `--regate-precheck` (inside the regate root only): problems are eligible stratum-1 < 20, direct-label positives (after twins) < 8, `rules_resolved_units` > 0, B1 or B2 over eligible items > 0.20; prints the counts, exit 1 on any problem; never opens `arms-p1/`, `arms-p2/`, `jev-p/`. `regate_gates` (`:1524`) uses `_regate_denominators` unchanged. |
| R17-02 (2) | `run.sh:370` | After the gold check: `py "$S/analysis_x2d.py" --regate-precheck || { echo "refusing: the re-gate's label-only denominators or construction checks fail (no Jev request was made): stop, report; the council and the user decide (PREREGISTRATION 5)" >&2; exit 1; }`. |
| R17-02 (3) | `PREREGISTRATION.md:110` | The stop-rule clause replaced with the council's text (pre-Jev check, "not a gate result", HD6 `preregistered` meanwhile, "A re-gate that passes this check and is then not PASS for BOTH P1 and P2 stops the pilot"). |
| R17-02 (4) | `PREREGISTRATION.md:112,113,116` | `:116` mirrored; the UNMEASURED clauses of `:112-113` read "(a defensive status, the pre-Jev check makes it unreachable)". |
| R17-02 (5) | `PREREGISTRATION.md:253` | Added "a pre-Jev re-gate shortfall leaves HD6 `preregistered` pending the user's decision"; the reason is now "round-0 gates P1 and P2 failed and the one permitted re-gate did not pass (re-gate P1 <status>, P2 <status>); HD6 never under test". |
| R17-02 (6) | `PREREGISTRATION.md:113`, `x2d.py:52` | "P(X ≥ 8) = 0.98 (0.82 at 1/4) for independent clones; the clones share 16 ablations, and with yield fully correlated within an area P(X ≥ 8) ≈ 0.88 at 1/3 and ≈ 0.71 at 1/4 [INFERENCE]; a shortfall is caught before any Jev request (above)." (the comment in `x2d.py` carries the same sentence). |
| R17-02 (7) | `reviews/regate-supply.md` | "minimum 8 positives with margin: met" and "Both minimums are met" replaced by "construction counts met; the direct-label positive yield is unmeasured until labels exist ...". The Result line now says the construction minimums are met. |
| R17-02 (8) | `analysis_x2d.py:1575` | `"positives_by_ablated_area"` added to the P2 value. |
| R17-02 (9) | `validation/regate_checks.py:466,473` (unit), `:775` (end to end) | Unit: 19 eligible stratum-1 items, 7 positives, B1 at 9 of 40 and one rules-resolved unit give a problem each; exactly 20 / 8 / 8 of 40 / 0 pass; no `arms-p1/`, `arms-p2/`, `jev-p/` present. End to end: a copy with the P2 clones' own labels rewritten to a kept area makes `run.sh regate-jev` refuse ("no Jev request was made"), no score file under `jev-p/`, no `regate/gates.json`, the SHA-256 manifest of the whole data directory unchanged. (The low yield is planted by rewriting the labelled copy's gold rows, not by a second stub-labeller run.) |
| R17-03 (1) | `PREREGISTRATION.md:111` | After "(4 per area at most)." inserted the council's sentence ("All planned P1 items and all but 4 P2 clones are `in_no_unit` clusters of 2019–2022 PRs ...") plus "`gates.json` carries each item's `date` and `origin`". Checked against the new supply: 40 of 40 P1 items and 36 of 40 P2 clones are `in_no_unit` with PRs below #10000, dated 2019-03-03 to 2022-09-01 (P1) and 2019-02-16 to 2022 (36 P2 clones); C0 is dated 2024-12-30. |
| R17-03 (2) | `PREREGISTRATION.md:111` | "replays the front end (§3) on every first-parent commit up to C0 that is the most recent addition (`git log --first-parent --no-renames --diff-filter=A`) of a C0 file owned by exactly one catalogue area, and on the reserve's commits, and keeps the residual clusters ...". |
| R17-03 (3) | `build_controls.py` `regate_plan` supply items | `"date": r["date"]` added to the P1 and P2 item dicts. |
| R17-03 (4) | `analysis_x2d.py:1455-1457,1475-1477` | `date` and `origin` carried into `p1_items` and `p2_items` of `gates.json` (fixture `:484`). |
| R17-04 (1) | `PREREGISTRATION.md:112` | "so the repaired gate is never easier than the rev-12 gate." replaced by the council's text ("... this is not dominance over the rev-12 gate ..."). |
| R17-04 (2) | `PREREGISTRATION.md:113` | After "(`REGATE_P2_POSITIVES_MIN`, the rev-12 floor)." the council's easing paragraph. |
| R17-04 (3) | `analysis_x2d.py:1416-1417` | `_p1_stats`: `recovery_all_items_A_only` and `sibling_in_acceptable`. |
| R17-04 (4) | `analysis_x2d.py:1576` | P2 value: `none_rate_all_positives`. |
| R17-04 (5) | `validation/regate_checks.py:137` (`p1-bacc`), `:192`, `:208` | `p1-bacc` asserts both new P1 fields (0.0 and 24); the P2 effect check asserts `positives_by_ablated_area`; the P2 bounds check asserts `none_rate_all_positives` (17 of 20 with 3 abstained: 0.85; 16 of 20: 0.80). |
| R17-04 (6) | `PREREGISTRATION.md:15` | Change log: "P2's decisive rate moves to answered positives with an abstention cap (easier on abstention; disclosed)". |
| R17-05 (1)-(2) | `PREREGISTRATION.md:250`, `:251` | Item 19: "[superseded in rev 13 for every stage after `power`: item 20]"; item 20: "**Fourth run commit.** ..." as worded. |
| R17-05 (3) | `README.md:15` | "the round-0 re-freeze, `gates-jev`, `gates-analyse` and `power` at the rev-12 sign-off commit, and the re-gate and every later stage at the rev-13 sign-off commit". |
| R17-06 (1) | `run.sh:293-295,298,308,324` | `gates_closed` (refuses when `rounds/0/gates.json` is non-empty, message "round 0's gates were analysed and its files are never written again ...") called after `verify` in `stage_gates_build`, `stage_gates_label`, `stage_gates_jev`. |
| R17-06 (2) | `validation/regate_checks.py:784` | On a finished copy each of the three stages refuses and the SHA-256 manifest of `rounds/0` (excluding `regate/`) is unchanged. |
| R17-07 | `PREREGISTRATION.md:110` | After the "**One shot:**" sentence the council's text ("`run.sh regate-build` runs alone first. The main agent then compares ... A deleted or rebuilt `regate/` is a protocol breach reported in `Deviations`."). |

Other consequences, not asked by a numbered fix: the combined `run.sh regate` stage is removed (it would have run the build with the other stages, against R17-07's "runs alone first"; `run.sh` header and dispatch, `README.md:3,36,67,74`, `PREREGISTRATION.md:15,110,196` updated); `DRAFT-NOTES.md` carries a round-17 paragraph (not a frozen file). `PREREGISTRATION.md:3` status line now reads "DRAFT for council round 18".

## 2. New supply figures (`reviews/regate-supply.md`, `.json`; membership only)

| item | figure |
|---|---|
| fresh pool | 129 clusters from 129 PRs in 69 areas; origin D 3, G 7, in_no_unit 119 (unchanged) |
| presented paths / rev-12 P1 target directories | 83 / 17 |
| P1 | 40 of 40 items; strata 1/2/3 = **37 / 0 / 3** (was 34/2/4); minimum 40 with >= 20 in stratum 1: met |
| P2 | 40 of 40 clones in 16 ablated areas (per area 4,4,4,3,3,3,2 x 9,1); origin D 2, G 2, in_no_unit 36 |
| population | P1: 40 in_no_unit, PRs < #10000, 2019-03-03 to 2022-09-01 (2019 22, 2020 9, 2021 5, 2022 4); P2: 36 such, 4 dated 2023-2024 |
| can host stratum 1 / 2 / 3 | 126 / 10 / 0 (excluding rev-12 target directories) |
| problems | none (no stop-and-report) |

The same 40 + 40 clusters as in the first record are chosen; the P1 placements changed.

## 3. New fixtures (G13; 8 added, 715 - 707)

`regate_checks.py` line numbers: `G13 R17-01 _p1_target` `:308`; `G13 R17-02 precheck` x2 `:466`, `:473`; `G13 R17-02 --regate-precheck` refuses outside the root `:482`; `G13 R17-03 ... date and origin` `:484`; `G13 R17-01 fresh P1 placements` `:610`; `G13 R17-02 end to end` `:775`; `G13 R17-06` `:784`. Extended: `p1-bacc` `:137`, the P2 planted effect `:192`, the P2 bounds `:208`.

## 4. Validation

`validate.py` with `signoff.json` moved aside: **715 checks, 0 failed**, "all validation checks passed" (651 before round 13, 707 at the first rev-13 draft). Log: `/tmp/validate-r17.log`. `signoff.json` SHA-256 before = after = `578a2462a6f25953c2e744a1d3dcc546f58af868f2e0e343e204367a89552e61`.

## 5. Frozen digests (45 files, after validation) and the tune pin

Tune pin (`~/.omp/agent/skills/jev-actions/scripts/tune_threshold.py`): `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c`, equal to the pin named in `round-17.md` §5.

Files changed since the round-17 review: `PREREGISTRATION.md`, `README.md`, `analysis_x2d.py`, `build_controls.py`, `run.sh`, `validation/regate_checks.py`, `x2d.py`. Unchanged: `gold.py`, `validation/validate.py`, `compose.py`, `baselines.py`, `b2.py`, `label_x2d.py`, `jev/main.go`, `GUIDE.md`, the six `questions/*`, `thresholds.json`, `build_units.py`, `frontend.py`, and every other row (all compared against the digest table of `rev13-draft-response.md`).

| file | sha256 |
|---|---|
| `GUIDE.md` | `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980` |
| `PREREGISTRATION.md` | `b9110486014bca580be3f14629c0beda54f35109a2c89ca5ca04ad95085140a8` |
| `README.md` | `bb7a01425e48071f3dba34a831c5de644c04c7cab65c28fcd2aed181d6292ce3` |
| `analysis_x2d.py` | `305848bff23a184c06f3d9ace8323014a7c6fc0155500858db318da38addaf19` |
| `b2.py` | `5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f` |
| `baselines.py` | `a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0` |
| `build_controls.py` | `34c466c34fe77db559ad0044724cebef2ee78adc617fab4f5f69880f26b3a128` |
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
| `run.sh` | `46bcf50055b5492a81dc526355eb7f678f55262d1d624c5fdd35f2cd5809260e` |
| `scopes.py` | `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072` |
| `thresholds.json` | `7142d7ac15abbcd27b75ac8005a7f1bffed0c6a70395c65e8f631bfcd32995eb` |
| `tune.py` | `2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c` |
| `validation/r4_checks.py` | `f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52` |
| `validation/regate_checks.py` | `07c575370bb9f17b27396e55752c7d5e49372324b674f7f0c083cbc546ce5943` |
| `validation/stub_clawpatch_x2d.js` | `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f` |
| `validation/stub_jev.py` | `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d` |
| `validation/stub_labeller.py` | `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5` |
| `validation/validate.py` | `8fab5ee017b8eecd8b0dc13dacb45edf9d7383e8586334ec7f994bf764ee16c9` |
| `versioning.py` | `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168` |
| `x2d.py` | `cf55b50d44fddc87eb36c889b9bba72e750746c5ef118b5247ffe7b8d47ab11d` |

## 6. For the council

- The low P2 yield in the end-to-end fixture is planted by rewriting gold rows in a labelled copy, not by a stub labeller; the pre-Jev check reads only gold, so the effect is the same.
- Directory candidates counted as "rejected as rev-12 targets" (662) are summed over the 40 placements of the final trial, not distinct directories (17 distinct).
- 37 / 0 / 3: strata 2-3 are report-only and now almost empty on the real supply.
