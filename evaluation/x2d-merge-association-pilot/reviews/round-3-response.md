# X2d — response to council round 3

Date: 2026-10-29. Applies every fix of `reviews/round-3.md` (18 issues, R3-01 … R3-18: 2 blocking, 7 major, 9 minor). PREREGISTRATION is now rev 4. Not signed off: this is the drafter's response and awaits council round 4. No real model call, Clawpatch map, label or Jev score was made (stubs only); no commit; nothing under `internal/`, `cmd/`, `scripts/`, `toolchain/`, `third_party/` or `go.*` was touched.

## Validation

`source .tools/activate && evaluation/x2d-merge-association-pilot/run.sh validate`: **586 checks, 0 failed** (exit 0; final log `/tmp/x2d-val-r3f.log`; run after the last edit of every study file; nothing skipped). Round 2 had 541 checks.

**P3 (binding, conditional on perception, N1, informativeness and ≥ 10 real positives):** p_B2 0.80, p_J 0.95, ρ 0.5, 1,000 pilots on the real 91 clusters / 40 PRs structure: GO 840 / 1000 = **0.840**, Wilson 95 % [0.816, 0.861]. **N3** maximum **0.014** (at p = 0.85); **N3-PR** maximum **0.034** (at p = 0.85); both ≤ 0.10. The live-structure run (`run.sh power` on the stub world's `rounds/0/k23.json`, 91 clusters / 40 PRs) printed P3 0.840 (Wilson [0.816, 0.861]), N3 max 0.014, N3-PR max 0.034, min inside-positive share 0.126, `real_none_proxy` 39 (the stub world's value, not the real catalogue's).

**Inside-population and abstention cells (R3-09, report-only; `power_sim.simulate_inside` with the analysis's estimators; 91 clusters, 81 covered, 10 gold-`none` positives inside; (p_J, p_B2) = (0.95, 0.80); 1,000 pilots; Wilson 95 %; computed once with `power_sim.empirical_structure()`):**

| Jev abstention | PR latent off | PR latent on |
|---|---|---|
| 0 | 0.577 [0.546, 0.607] | 0.556 [0.525, 0.587] |
| 0.05 | 0.326 [0.298, 0.356] | 0.337 [0.308, 0.367] |
| 0.10 | 0.126 [0.107, 0.148] | 0.150 [0.129, 0.173] |

**End-to-end stub world (all ten rounds + `verdict`, real session cap):** planted effect (stub V-sem 0.95 vs the real B2) → **GO** (gain 0.857, P_assign 0.914, R_assign 0.928, 22 real positives; asserted). Redrawn-primary null (primary arm's frozen-round correctness redrawn per cluster at B2's realised accuracy 0.055, 91 clusters, seed 20261029) → NO-GO (gain −0.033, P_assign 0.011). Common-accuracy null (R3-08: both arms redrawn at 0.95, 91 + 91 clusters) → NO-GO: P_assign 0.904 passes its point condition, gain −0.055; the failing conditions are the gain lower bound and the gain point. Copy-B2 null (zero paired gain) → not GO (sanity).

First-run failures and fixes (study-code and fixture bugs, no check weakened): (1) the `compose_checks` fixtures called `missing_required` without `area_ids` and hit a missing catalogue file (the fixtures now pass per-area ids); (2) `labels_check_checks` called `A.labels_digest(good, 1)` with the arguments swapped (the signature is `(r, data)`); (3) `tune._check_v7_scores` compared the header's `versions` against the plain question names, but the Go header's keys are `<question>/V-sem` (the check and its fixtures now use the three `<question>/V-sem` keys; the first full run showed this as three failures: the e2e `rescore-v7`, the header check and `run.sh units 8`); (4) `analysis_x2d.summarise` raised `KeyError: 'none_correct'` when an arm gave no `none` answer on a round with gold-`none` units (now `.get(…, 0)`; found by the redrawn-primary null, where only 2 of 91 clusters were correct).

## Choices where the council offered one, and implementer departures

Where the council offered a choice the recommendation was taken:
- **R3-05:** the re-render route (rebuild the `jev-v7` units under round 7's `excerpt_lines`) was taken, not the alternative of removing `excerpt_lines` from the per-round menu; `excerpt_lines` stays in the menu.
- **R3-09 / R3-16:** `p3_inside_min` is a report line the critique must quote; it has no threshold.
- **R3-01:** at most one guide clarification per round (`C ≤ 1`); a second is refused.

Departures from the literal text, each recorded in PREREGISTRATION §14 item 15:
- **R3-05 header:** the council wanted the `jev-v7` header to record `excerpt_lines`. The Go header has no such field and no Go change was made; each re-scored unit row carries `excerpt_lines` (`tune.py` requires every row to equal round 7's) and the header's `input_sha256` digests those rows.
- **R3-16:** the critique also enforces `p2_unmeasured: accepted` when the P2 gate is UNMEASURED (the council: "the critique must rule and record it").
- **R3-03:** "every control with a labelled gold" in round 0 is implemented as the non-P2 controls that have label rows or call rows (P2 clones take derived gold).
- **R3-12:** the pre-freeze envelope check runs for every round r ≥ 1 (not only after a catalogue-byte change).

## Per-issue

| id | status | what was done (file) |
|---|---|---|
| R3-01 | fixed | `label_x2d.cap_for` = `ceil(1.25 × (2.2·W·(1 + C) + 4.4·J + R + 2.2·I))` ≤ 500 with `C` = rows of `guide-revisions.jsonl`; `--precheck-clarify` (`clarify_precheck`) refuses a clarification that would exceed the cap with C + 1, run by `run.sh stage_clarify` before `versioning.py --clarify` records it; `versioning.clarify` refuses a second revision; `reference_arm.py` writes `unresolved: session_cap` rows instead of exiting; `thresholds.json` note and PREREGISTRATION §12 (illustrative 232 / 328). Validation runs rounds under the real cap (stub override removed except the capped fixtures), cap-refused clarify, `C == 1` in round 3. |
| R3-02 | fixed | `analysis_x2d.labels_check` decides `ok` by `kappa_answer ≥ 0.6` alone and reports `rule_top_share_over_0_40`; `--labels-check` exits 1 only for a refinement round (1–7) with no recorded revision; `run.sh require_labels_check` refuses a failed check only under that condition; PREREGISTRATION §8. e2e: round 5 refuses `gold`, round 3 proceeds and is non-informative. |
| R3-03 | fixed | `gold.required_gold_ids` (window units for r ≥ 1; round 0: D ∪ P plus labelled or attempted non-P2 controls; contrast units for the contrast tag); `gold.py` refuses before writing anything, with or without `--next`. e2e `e2e_round0_gold_gap`. |
| R3-04 | fixed | `build_controls`: `P2_MIN_GAP_POSITIVES = 8`, `P2_MAX_CLONES = 12`, `need = min(12, eligible − 20)`, refusal text "planted-gap gate cannot be built; stop, report, the council and the user decide"; `gates()` marks P2 UNMEASURED below 8 label-derived gap positives (with `no_agreed_gold`) and FAIL only for a none rate < 0.8 over ≥ 8; PREREGISTRATION §5, §11. |
| R3-05 | fixed (departure: header) | `build_units.rerender_units`; `run.sh stage_rescore_v7` writes `jev-v7/units.jsonl` rendered under round 7's `excerpt_lines`; `tune._check_v7_scores` checks it. e2e: round 2 built at 30 lines, re-score units at 60, 18 units grew; unit fixture on `rerender_units`. |
| R3-06 | fixed | `compose.missing_required(…, area_ids)` requires every catalogue area of a fan-out; `_compose_fan` gives `unresolved:failure` for a missing area row; `compose.py --catalogue`, passed by `run.sh stage_gates_jev`. Single-area partial fan-out fixture. |
| R3-07 | fixed | `analysis_x2d.labels_digest(r, data)` over the GO population's label rows, written as `labels_sha256`; `run.sh require_labels_check` refuses a changed digest ("re-run labels-check"). |
| R3-08 | fixed | e2e common-accuracy null (both arms redrawn at 0.95, seeded): P_assign point passes, verdict NO-GO, a gain condition among the failing; planted effect GO and redrawn-primary null retained. |
| R3-09 | fixed | `power_sim.simulate_inside` rewritten: copula latents for every cluster, a wrong non-abstaining answer on a positive counts in `assigned`, B2 drawn on the positives, gain and spread over all clusters; spy check on `A.go_verdict`; six cells recorded above and in §11. |
| R3-10 | fixed | `labels_check` κ is over the GO population (core-gate units excluded); `kappa_answer_all_units` and `n_units_all` reported. |
| R3-11 | fixed | `build_units.gates` adds a unit to `frozen_pr_counts` only when `stratum_reason != "core_gate"`; `frozen_pr_counts_all_units` keeps every scored cluster; e2e sums checked. |
| R3-12 | fixed | `run.sh stage_units` runs `versioning.py --round r --check` for every r ≥ 1 before the freeze; e2e catalogue-byte change refused at `units 1`. |
| R3-13 | fixed | `check_label_args "$@"` in `stage_gates_label` and `stage_quality_label`; e2e refusals. |
| R3-14 | fixed | `stage_tune` requires `gold.jsonl` and `jev/scores.jsonl`; `stage_analysis` requires `arms/B2.jsonl`, `arms/reference.jsonl`, `jev/scores.jsonl` and, in rounds 1 and 10, `arms/contrast-B2.jsonl` and `jev/scores.contrast.jsonl`; e2e refusals. |
| R3-15 | fixed | `verdict()` writes `leave_one_round_out` for rounds 8, 9, 10 (left-out round, remaining rounds, pooled figures; no verdict); e2e check; PREREGISTRATION §11, §14. |
| R3-16 | fixed | `x2d.critique_problems(0)` requires `data/power.json`, `rounds/0/gates.json`, the `real_none_proxy`, `p3_real_structure` (4 decimals) and `p3_inside_min` (4 decimals) lines equal to the recorded values, `user_decision: proceed` when the proxy < 10 or P3 < 0.80, `p2_unmeasured: accepted` when P2 is UNMEASURED; fixture check `r3_checks`. |
| R3-17 | fixed | `x2d.jev_budget` charges files under `jev-v7/` to the $10 total only and excludes them from the per-round term; `jev_budget_v7_test`. |
| R3-18 | fixed | `tune._check_v7_scores` requires exactly the three versions `merge.covered/V-sem`, `merge.which_area/V-sem`, `merge.area_covers/V-sem`, each equal to round 7's, at least one `area_covers` row, and an empty `compose.missing_required` over the dev units (round 0: the manifest's main catalogue); mutation fixtures (covered-only header, extra version, dropped fan-out row, dropped permutation). |

## Digests (SHA-256, 43 frozen files; recompute at sign-off)

External pin (not a study file, recorded in `signoff.json` at sign-off): `tune_threshold.py` = `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c` (recomputed 2026-10-29; unchanged since round 2).

| File | SHA-256 |
|---|---|
| `GUIDE.md` | `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980` |
| `PREREGISTRATION.md` | `595d7c20be0dafc04b01c0e5e97279b23c41a3fa925816e7f2d9b697e28c333a` |
| `README.md` | `536ae3f45b29477f4883c40e97ab0b5d8d7359c19fa9615daeea90bca383dfb8` |
| `analysis_x2d.py` | `a7b0de86f205e37317903bc3973a02d3899e5ed96be446d4fcecbc55baed9cfb` |
| `b2.py` | `5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f` |
| `baselines.py` | `a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0` |
| `build_controls.py` | `7ded8c9177d8e21131ed91ee0aa423a4dd8c64ac35b215f82e93b1bfbfe6c81b` |
| `build_units.py` | `c052246c828024a25fc6a143d513e111ecbf63b42b8b5912dd37e960a1c3e243` |
| `collect_onboard.py` | `c93fa59c8fe305459336e08e52c60f365d8df0a70d7b6ca395c58e8f88da5fff` |
| `compose.py` | `3b47ed4d495d6c080a83b988c67357fc57d9846b01a8350769de843a6661b603` |
| `frontend.py` | `dd4a0fdc1e3595eeefb82a8eafab600255b3b0fe95f67e2fb8a07eb5972ea0d8` |
| `gold.py` | `f18c5d9e05c9cddd11a7c75ce5f36801880ea2e4e2e9ba9733eba35fc1ff738b` |
| `jev/main.go` | `63fbec0534a2ddf3d55d07c074c6a882e199ec2b90cf771a984c9fa9ee586ca0` |
| `label_x2d.py` | `51474ad9af28ca8c9363b1ee40dc90ebfb13b23d744d1388f95c5e6f4a08781d` |
| `measurement/lever-result.json` | `298902e717d0db8826c35d219695d386b19644f99cfc205d6b2d891100b0352c` |
| `measurement/lever_check.py` | `8f8008d3a28d72d5e768bc292e4917449fd8cb31ca0a642ff6057ab51b705a42` |
| `onboard/main.go` | `7e2ed1edc64b694851e7729e17af4f2ba63dafa7c9cdcf19cd3a5a255c08d46b` |
| `power_sim.py` | `0e949ac323ec1f1b5a1922679a40b33ef4882fe25d1a9872a211515d68c50aea` |
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
| `run.sh` | `832d70136d8d3bd437e4ea1fb0e62f5f4955de97d15bdd7cb908cc3cfd7293f7` |
| `scopes.py` | `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072` |
| `thresholds.json` | `ce9e9ed9d75443173799db7f8f4f1a46562d3be4fac1219a7b6c70dcd828ab30` |
| `tune.py` | `d583438ac87da3b2dfaabc57bb6a1b0f37657e9784eb8033b565636045698e7c` |
| `validation/stub_clawpatch_x2d.js` | `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f` |
| `validation/stub_jev.py` | `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d` |
| `validation/stub_labeller.py` | `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5` |
| `validation/validate.py` | `3f58d118ee405f30368b6d81b2982dd6e904d4a702bc119d236641e3816dcaa9` |
| `versioning.py` | `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168` |
| `x2d.py` | `119f65f6ef1b881161c574f2d6cb27109eb2616ddac76ca1e013795ddcaccbe0` |

Reproduce: `source .tools/activate && evaluation/x2d-merge-association-pilot/run.sh validate`; digests: the command at the end of `DRAFT-NOTES.md` §5.
