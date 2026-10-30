# X2d — response to council round 2

Date: 2026-10-29. Applies every fix of `reviews/round-2.md` (18 issues, R2-01 … R2-18). PREREGISTRATION is now rev 3. Not signed off: this is the drafter's response and awaits council round 3. No real model call, Clawpatch map, label or Jev score was made (stubs only); no commit; `git status` shows changes only under the study directory and the two ledger files already modified before this round (`docs/hypothesis-ledger.md`, `docs/research-ledger.md`).

## Validation

`source .tools/activate && evaluation/x2d-merge-association-pilot/run.sh validate`: **541 checks, 0 failed** (exit 0; log `/tmp/x2d-val.log`; nothing skipped).

**P3 (binding, conditional on perception, N1, informativeness and ≥ 10 real positives):** p_B2 0.80, p_J 0.95, ρ 0.5, 1,000 pilots on the real 91 clusters / 40 PRs structure: GO 840 / 1000 = **0.840**, Wilson 95 % [0.816, 0.861]. **N3** maximum **0.014** (at p = 0.85); **N3-PR** maximum **0.034** (at p = 0.85); both ≤ 0.10. The live-structure run (`run.sh power` on the stub world's `rounds/0/k23.json`, 91 clusters / 40 PRs) printed P3 0.840 (Wilson [0.816, 0.861]), N3 max 0.014, N3-PR max 0.034, min inside-positive share 0.107, `real_none_proxy` 39 (the stub world's value, not the real catalogue's).

**End-to-end stub world (all ten rounds + `verdict`):** planted effect (stub V-sem 0.95 vs the real B2) → **GO** (gain 0.857, P_assign 0.914, R_assign 0.928, 22 real positives; asserted, R2-12). Planted null (primary arm's frozen-round correctness redrawn per cluster at B2's realised accuracy 0.055, seed 20261029) → NO-GO (gain −0.033, P_assign 0.011). Copy-B2 null (zero paired gain) → not GO (sanity).

First-run failures and fixes (study-code bugs, no check weakened): `run.sh stage_gates_jev` composed over all of `units.jsonl` including the P1/P2 controls and tripped `missing_required` (now `--units "$units"`); `power_sim.run()` passed `pw` positionally as `thresholds` to `simulate_inside` (now `pr_w=pw`); `x2d.critique_problems` compared `next_inputs_sha256` with the live digest even after the next round froze, so a mid-round clarify invalidated the previous critique (after freezing it compares round n+1's frozen `inputs_sha256`; later edits are policed by `versioning --drift`); `tune._check_v7_scores` demanded equal `versions` although `rescore-v7` runs sem + fan only (a non-empty subset of round 7's versions is accepted; model and catalogue digest stay strict). Two validate.py fixes that were not study bugs: the R2-18 check looked for the dead `cli_sessions_*` keys by prefix and matched the legitimate `cli_sessions_note` / `cli_sessions_ceiling` (now exact names); the R2-04 capped-case check no longer expects a call row for a case the cap stopped before any session. `validate.py` keeps two development-only environment knobs, `X2D_VALIDATE_SKIP` and `X2D_VALIDATE_KEEP` (the reviewed run uses neither).

Two things the council should look at:
1. **Stub world:** round 0 needs `features=3`, not 6. With 6 features only 7 P2 gap positives exist (27 eligible P units) and `build_controls` correctly refuses below 8; with 3 features there are 28 eligible units. This changes only the stub world.
2. **Clarify relabel vs session cap:** a mid-round `clarify` relabels the wide cases at about 2.2 sessions each; that cost is not in the `cap_for` formula. In the stub, round 3 used 215 sessions against a formula cap of 207; the validation allows the formula plus `ceil(2.2 W)` for the clarified round. With the real 1.25 headroom a relabel could reach the cap mid-way (the label stage then stops and `--new-round` fills). The council may want to add the clarify cost to the formula or the §12 note.

## Choices where the council offered one, and implementer departures

Where the council offered a choice the recommendation was taken. Departures from the literal text, each recorded in PREREGISTRATION §14 item 14:
- `rescore-v7` resumes with per-round `rounds/<rr>/jev-v7/.done` markers (a killed round resumes) instead of refusing when any scores file exists; `tune` and `run.sh units 8` re-check headers and completeness (R2-02).
- `JEV_SETTINGS` / `JEV_QUESTIONS` are shell variables of `run.sh`, not environment variables (R2-01/R2-02).
- `gold` is its own stage after `labels-check` (so a clarification relabel can precede the append-only gold) (R2-01/R2-04).
- `label_x2d.py` gains `--cases` and `--guide-round` (clarify relabel, R2-01).
- The real-`none` proxy is an INFO row in `gates.json`, not a stop; the round-0 critique file must carry `real_none_proxy:` and `p3_real_structure:` lines (R2-11).
- In rounds 8–10 a failed labels-check does not block; the round is scored and marked non-informative (INCONCLUSIVE) (R2-03).

## Per-issue

| id | status | what was done (file) |
|---|---|---|
| R2-01 | fixed | Per-round `rounds/<r>/inputs/{GUIDE.md,questions/}` (`x2d.inputs_dir/guide_path/questions_dir/inputs_digest/proposed_inputs_dir`, `versioning.ensure_inputs/clarify`), version.json carries `inputs_sha256`; `guide_version` = prefix + hash of the round's guide; `jev`/`label_x2d`/`run.sh` read the round's inputs; the critique pins the next round's inputs (`next_inputs_sha256`); `versioning --clarify` / `run.sh clarify` (rounds 1–7, before Jev scores) with `guide-revisions.jsonl`; gold uses the current guide version, last final row wins. PREREGISTRATION §7, §10. |
| R2-02 | fixed | `run.sh rescore-v7` (critique 7 required; per-round `.done` markers); `tune._check_v7_scores` checks model, catalogue digest, versions and completeness of every dev unit; `run.sh units 8` refuses without a valid re-score; the fake copy is gone from validation. |
| R2-03 | fixed | `prereq` (verify, gates, previous critique, `versioning --drift`) in every per-round stage; `labels-check.json` (guide version, `ok`, figures) required by `jev`/`gold`/`tune`/`analysis` (rounds 8–10: must exist); round-0 `labels-check` before gold and Jev; `versioning --drift` refuses settings/inputs edits after the freeze. |
| R2-04 | fixed | `gold.py` builds rows and `next_membership` in memory and writes both or neither; `_gold_outputs` raises `ValueError` (caught by `_divergence` / `replay_memberships`, which report `unmeasured` with the reason); baselines refuse cleanly without gold. |
| R2-05 | fixed | `derive_ablation(gold, ablated_set)` removes the set the catalogue removes; `control-key.jsonl` stores `ablated_area_ids` and `gap_positive`; the P2 gate is over gap positives only (derived gold `covered == False`), minimum 8 (`P2_MIN_GAP_POSITIVES`, `build_controls` refuses below), retained-covered clones reported. |
| R2-06 | fixed | D units (20) join `units.contrast.jsonl` in rounds 1 and 10 (set `D`); `contrast_section` reports `labels_D` / D scores separately. |
| R2-07 | fixed | `build_units` writes a lookups row for every commit with a non-zero count and `rounds/<r>/counts.json`; `_end_to_end` asserts equality (ValueError). |
| R2-08 | fixed | `rule_session_counts` parses `R\d+` ids once per A/B session; raw strings kept (`rule_counts`). |
| R2-09 | fixed | `compose` adds `raw_covered` / `raw_choice` to V-sem rows; perception compares them; thresholded agreement, unresolved rate and primary-arm agreement are reported beside it. PREREGISTRATION §8. |
| R2-10 | fixed | `controls_section` in `round_results`, `report.md` and the pack: dup agreement and output equality, trap gold-not-covered and none rate, prose equality; report-only. |
| R2-11 | fixed | `k23.json` has `real_none_proxy` and `frozen_pr_counts`; `power_sim --structure-k23` (schema `x2d-power-v3`, `P3_inside`, `p3_real_structure`); `run.sh power`; round-0 critique lines required before round 1; `gates.json` INFO row; PREREGISTRATION §5, §11. |
| R2-12 | fixed | e2e planted effect asserts GO; redrawn-primary null asserts not GO; copy-B2 kept as a zero-gain sanity test. |
| R2-13 | fixed | `compose.required_perms` / `missing_required` / `header_flip_all`; missing required row → `unresolved:failure`; `compose.py` and `run.sh analysis` refuse; tune dev rows are window units (+ D/P in round 0), no audit units, Choice rows only when every required permutation answered. |
| R2-14 | fixed | `informative` and futility use the GO-population κ (`labels_go_population`); all-unit figure kept in `labels`. |
| R2-15 | fixed | `cap_for(..., stub_flag=...)`: each caller honours the override only under its own flag (`reference_arm` passes `X2D_REFERENCE_STUB`). |
| R2-16 | fixed | `x2d.critique_problems` (exactly one `Verdict:` line, one `version_sha256`, `next_inputs_sha256`); `verify_pins` requires the frozen file set to equal the record; `x2d.require_signoff` / `reviews_dir` honour the skip flags only with a loopback `X2D_JEV_ENDPOINT` and `X2D_DATA` outside `S/data`; `run.sh label` forwards only `--new-round/--workers/--cases`. |
| R2-17 | fixed | `dump-state` also writes `noul_instructions`, `noul_criteria`, `choice_instructions` (base + delta); labeller and reference arm use them; perception always compares V-sem; §8 states it. |
| R2-18 | fixed | `cli_sessions_per_round/_round_0` removed from `thresholds.json` and `x2d.CAPS` (and the validation equality); `baselines.py` docstring order; `reviews/round-<N>.md`; K4 "κ of the joint answer"; §3 cap rationale reworded (cap 10 fixed before any outcome; it neither minimises the largest share (cap 8: 0.40) nor maximises frozen clusters (cap 6: 70)). |

## Digests (SHA-256, 43 frozen files; recompute at sign-off)

External pin (not a study file, recorded in `signoff.json` at sign-off): `tune_threshold.py` = `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c`.

| File | SHA-256 |
|---|---|
| `GUIDE.md` | `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980` |
| `PREREGISTRATION.md` | `8767fc4846edc52b79c8d3e9e1764b2d238e5455b41aae31ce14e8c07db45f6b` |
| `README.md` | `32ee7ad29c454c07c3e42441c91416e4110e39247f0f14056fb871eab816a1fe` |
| `analysis_x2d.py` | `0603e666d2847e2a3733f239a060eefb563f44cc7981e8c97b97b8140e4d72e6` |
| `b2.py` | `5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f` |
| `baselines.py` | `a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0` |
| `build_controls.py` | `8a633b0d639518921256edad9082fe9e7e43bd60a15cca46e43caaa471f52f47` |
| `build_units.py` | `4d40a47dc498e171d553c05783e395fb645689f9cf8df91940f0852cbe808eb7` |
| `collect_onboard.py` | `c93fa59c8fe305459336e08e52c60f365d8df0a70d7b6ca395c58e8f88da5fff` |
| `compose.py` | `589ed5da311ab3a21a9f986e5016557c5c5ff6ff9b86430a21e40e29635224ea` |
| `frontend.py` | `dd4a0fdc1e3595eeefb82a8eafab600255b3b0fe95f67e2fb8a07eb5972ea0d8` |
| `gold.py` | `4df56487c714be814ba18763de720e68c8b3f750ba77ca269d10b593fadce589` |
| `jev/main.go` | `63fbec0534a2ddf3d55d07c074c6a882e199ec2b90cf771a984c9fa9ee586ca0` |
| `label_x2d.py` | `d3d9a7d6c41345de996378dfefe4ff55e4b1d764b261dbc0ff49e7143d61c1aa` |
| `measurement/lever-result.json` | `298902e717d0db8826c35d219695d386b19644f99cfc205d6b2d891100b0352c` |
| `measurement/lever_check.py` | `8f8008d3a28d72d5e768bc292e4917449fd8cb31ca0a642ff6057ab51b705a42` |
| `onboard/main.go` | `7e2ed1edc64b694851e7729e17af4f2ba63dafa7c9cdcf19cd3a5a255c08d46b` |
| `power_sim.py` | `84ff0e048a51f9929e2ed0705ecf329f8303b4cdb2a424e030807c68c4bad3ff` |
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
| `reference_arm.py` | `5012083bcd22fd4ead85bbb31f44e2aa3fadac610ba66f663f0574954addc3f9` |
| `run.sh` | `b3822c6a7a5b108d748d03c07454c7468b5bce3f0a7166c02abd282c155ad692` |
| `scopes.py` | `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072` |
| `thresholds.json` | `caab0d92223eb9661267912091579c480315ef799522c829c6f2b0dd35175a36` |
| `tune.py` | `3cc5924ae8a916361d28fb1c0743b273918c2861e8f2875910a11eea0f5f09f0` |
| `validation/stub_clawpatch_x2d.js` | `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f` |
| `validation/stub_jev.py` | `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d` |
| `validation/stub_labeller.py` | `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5` |
| `validation/validate.py` | `95aac78efaba458daf9b0f2f92a0121c7dfa928dab633567541ed8c4b7ed9e82` |
| `versioning.py` | `afd1c2db9e04290ea0c3e008b187fc70706c4be535022ede651f9f5c7b3cca7a` |
| `x2d.py` | `8a12c1977a455c94a314ee62465cfa35d8383421decb78278d883216e3983140` |

Reproduce: `source .tools/activate && evaluation/x2d-merge-association-pilot/run.sh validate`; digests: the command at the end of `DRAFT-NOTES.md` §5.
