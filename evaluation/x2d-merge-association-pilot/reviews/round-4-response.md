# X2d — response to council round 4 (and its Addendum, R4-13)

Date: 2026-10-30. Applies every fix of `reviews/round-4.md`: R4-01, R4-02, R4-04 … R4-12 as issued and the Addendum's R4-13, which supersedes R4-03's fix items. PREREGISTRATION is now **rev 5**. Not signed off: this is the drafter's response and awaits council round 5 (`reviews/round-5.md`). No real model call, Clawpatch map, label or Jev score was made (stubs only); no commit; nothing under `internal/`, `cmd/`, `scripts/`, `toolchain/`, `third_party/` or `go.*` was touched; no `.codegraph` was touched.

## User decision record (sign-off condition (i))

- **Date:** 2026-10-30.
- **Context:** the main agent put the round-4 disclosure (R4-03: the rev-4 GO rule's six inside/abstention cells, labelled drafter-reported simulations, not the real pilot's power) and two options to the user: **(1)** proceed as preregistered; **(2)** revise the GO rule (a new preregistration revision and council round).
- **The user chose (2), revise the GO rule, in these words (verbatim):** *"Judge accuracy and the gain over the baseline on the clusters Jev actually answers, with a separate cap on how often it abstains (the council sets the exact cap and re-checks power). New-area detection is reported with intervals, but ~10 cases can't veto."*
- **Research-ledger ID: R-051** (`docs/research-ledger.md`, appended by the main agent; not edited here), cited beside R-049 in PREREGISTRATION §14 (the USER DECISION RECORD paragraph) and in the HD6 row of `docs/hypothesis-ledger.md`.
- **Council specification:** `reviews/round-4.md` Addendum, R4-13. The rule was set before any predictor–outcome association exists (rule 3).
- **HD6** (`docs/hypothesis-ledger.md`): the sentence gains "on the clusters it resolves, abstaining on ≤ 15 %", a 2026-10-30 History line cites R-051, the state stays `proposed`. The HD6 clause "detects genuinely new concerns with precision and recall ≥ 0.80" was left byte-identical as instructed; it now reads against the no-veto rule (see "For the council / main agent" below).

## Validation

`source .tools/activate && evaluation/x2d-merge-association-pilot/run.sh validate`: **631 checks, 0 failed** (exit 0; final log `/tmp/x2d-val-r5d.log`, run after the last edit of every frozen study file; nothing skipped). Round 3 had 586 checks (+45: abstention / resolved-pairing / go-rule unit checks, the R4-04 P2 fixture, power-grid checks, the four e2e abstention worlds, and the R4-01, R4-02, R4-06, R4-07, R4-08, R4-11 fixtures of `validation/r4_checks.py`). Earlier runs of this response: 630 checks / 3 failed (stale `real_none_proxy` bound assertions after R4-13 made it INFO without a bound; `run.sh power` exited 1 on a missed bar) → fixed → 630 / 0 (log `/tmp/x2d-val-r5c.log`) → the R4-04 fixture added (it was missing) → 631 / 0.

## Power (drafter-reported simulations; model-free; `power_sim.run`, schema `x2d-power-v4`, seed 20261029, 1,000 pilots, bootstrap B = 1,000)

Model: the empirical 91 clusters / 40 PRs structure, **10 gold-`none` positives inside the population**, perception / N1 / informativeness taken as passing, arm correlation 0.5 (0 for N3), `fn_share` 0.5, abstention drawn per cluster independent of correctness in the honest cells; the verdict is `analysis_x2d.go_verdict` itself (conditions a–l). Each cell shows GO share, Wilson 95 % and (GO / INCONCLUSIVE / NO-GO counts). These are simulations on assumed accuracies, not the real pilot's power.

| cell | setting | GO share [Wilson 95 %] (GO/INC/NO-GO) | bar | status |
|---|---|---|---|---|
| P3-u0 | p_J 0.95, p_B2 0.80, ρ 0.5, u 0 | 0.859 [0.836, 0.879] (859/0/141) | ≥ 0.80 | met |
| P3-u05 | same, u 0.05 | 0.886 [0.865, 0.904] (886/0/114) | ≥ 0.75 | met |
| P3-u10 | same, u 0.10 | 0.724 [0.695, 0.751] (724/56/220) | ≥ 0.60 | met |
| N3 | p 0.50, u 0 / 0.10 | 0.000 / 0.000 | ≤ 0.10 | met |
| N3 | p 0.60, u 0 / 0.10 | 0.000 / 0.000 | ≤ 0.10 | met |
| N3 | p 0.70, u 0 / 0.10 | 0.003 [0.001, 0.009] / 0.002 [0.001, 0.007] | ≤ 0.10 | met |
| N3 | p 0.80, u 0 / 0.10 | 0.034 [0.024, 0.047] / 0.028 [0.019, 0.040] | ≤ 0.10 | met |
| N3 | p 0.85, u 0 / 0.10 | 0.020 [0.013, 0.031] / 0.022 [0.015, 0.033] | ≤ 0.10 | met |
| N3 | p 0.90, u 0 / 0.10 | 0.010 [0.005, 0.018] / 0.013 [0.008, 0.022] | ≤ 0.10 | met |
| N3 | p 0.95, u 0 / 0.10 | 0.001 [0.000, 0.006] / 0.001 [0.000, 0.006] | ≤ 0.10 | met |
| N3-PR | p 0.50, u 0 / 0.10 | 0.000 / 0.000 | ≤ 0.10 | met |
| N3-PR | p 0.60, u 0 / 0.10 | 0.001 [0.000, 0.006] / 0.000 | ≤ 0.10 | met |
| N3-PR | p 0.70, u 0 / 0.10 | 0.006 [0.003, 0.013] / 0.004 [0.002, 0.010] | ≤ 0.10 | met |
| N3-PR | p 0.80, u 0 / 0.10 | 0.041 [0.030, 0.055] / 0.041 [0.030, 0.055] | ≤ 0.10 | met |
| N3-PR | p 0.85, u 0 / 0.10 | 0.046 [0.035, 0.061] / 0.045 [0.034, 0.060] | ≤ 0.10 | met |
| N3-PR | p 0.90, u 0 / 0.10 | 0.019 [0.012, 0.029] / 0.020 [0.013, 0.031] | ≤ 0.10 | met |
| N3-PR | p 0.95, u 0 / 0.10 | 0.003 [0.001, 0.009] / 0.003 [0.001, 0.009] | ≤ 0.10 | met |
| N4a (abstain on B2's errors, ≤ 13 clusters) | p 0.80 / 0.85 / 0.90 / 0.95 | 0.000 [0.000, 0.004] in all four (0/0/1000) | ≤ 0.10 | met |
| **N4b (oracle self-abstention)** | **p 0.85** | **0.255 [0.229, 0.283] (255/42/703)** | ≤ 0.10 | **MISSED** |
| **N4b** | **p 0.88** | **0.228 [0.203, 0.255] (228/39/733)** | ≤ 0.10 | **MISSED** |
| **N4b** | **p 0.90** | **0.141 [0.121, 0.164] (141/20/839)** | ≤ 0.10 | **MISSED** |
| N4b | p 0.95 | 0.006 [0.003, 0.013] (6/0/994) | ≤ 0.10 | met |
| N4c (p_J 0.90, p_B2 0.80, own-error abstention) | report-only | 0.871 [0.849, 0.890] (871/2/127) | — | report-only |
| P3-PR (PR latent, u 0.05) | report-only | 0.827 [0.802, 0.849] (827/6/167) | — | report-only |
| P3-ρ0 (ρ 0, u 0.05) | report-only | 0.856 [0.833, 0.876] (856/3/141) | — | report-only |
| δ = 0.15 curve (u 0.05) p_B2 0.70 / 0.75 / 0.80 / 0.85 | report-only | 0.470 [0.439, 0.501] / 0.762 [0.735, 0.787] / 0.886 [0.865, 0.904] / 0.914 [0.895, 0.930] | — | report-only |

`checks` = {P3_u0 true, P3_u05 true, P3_u10 true, N3 true, N3_PR true, N4a true, **N4b false**}; `all_pass` false; `p3_real_structure.share` 0.859; `p3_abstention` {0.05: 0.886, 0.10: 0.724}; `n4_max` 0.255.

**A binding bar is missed: N4b, in three of its four cells (p = 0.85, 0.88, 0.90).** I did not move the bar. Which rule applies (the instruction's alternative): **the Addendum says to report it and route it to the round-0 mechanism** (§(6): "a binding cell below its bar is reported in `reviews/round-4-response.md` and goes to the round-0 mechanism of (7)"; §(7): `user_decision: proceed` is required when any binding cell on the live structure is below its bar). So the check is **not left failing**: validation asserts the bars from `power.json`, records the miss (`checks.N4b` false, `all_pass` false, a printed "BAR MISS" line), and the round-0 critique then requires `user_decision: proceed` (`x2d.critique_problems(0)`, fixture-tested). `power_sim.py`'s exit code is 0 on a missed bar (it was 1 for the old P3 bar) so the round-0 stage chain is not cut; the miss is carried by `power.json`.

Why N4b misses [INFERENCE from the simulation; the measured shares are above]: when Jev abstains exactly on its own errors every resolved Jev answer is correct (P_assign = R_assign_res = 1), `gain_all` = (Jev-correct − B2-correct)/n is zero in expectation but positive (strictly > 0) in about half of the pilots by sampling noise, and `gain_res` = 1 − P(B2 right | Jev right) is ≈ +0.10 ± noise, so the point (d) and the guard (h) are both passed together in ≈ 0.14–0.26 of pilots. The council's arithmetic (Addendum (1), "`gain_all` = 0.88 − 0.88 = 0 → fails `h`") holds for the expectation, not for a pilot: `h` as specified (`gain_all` point > 0) is not a guard against noise. The e2e gaming world (below) is built with exactly equal correct counts so that `gain_all` = 0 and `h` fails deterministically; the simulation draws the counts independently, which is what a real pilot would do.

For the live structure the stub world's `rounds/0/k23.json` (91 clusters / 40 PRs, `real_none_proxy` 39 → n_pos capped at 20) gave: P3-u0 0.857, `p3_abstention` 0.05 = 0.857 and 0.10 = 0.744, `n4_max` 0.231, N4b false (the stub world's values, not the real catalogue's).

## End-to-end stub world (all ten rounds + `verdict`, real session cap)

Planted effect (stub V-sem 0.95 vs the real B2) → **GO**, all 12 conditions pass (gain_res 0.857, n_res 91, gain_all 0.857, u 0, P_assign 0.914, R_assign_res 0.928, 22 real positives, perception 0.917 vs labeller 1.0). New R4-13 (8) worlds, each a copy of the finished run whose primary-arm composed frozen-round rows are rewritten, then `analysis --round 8/9/10` and `--verdict` re-run:
- 5 % hash-selected abstentions (5 of 91, u = 0.055) → **GO**, every condition true (incl. g, h, i).
- 20 % (18 of 91, u = 0.198) → **NO-GO**, the cap the only failing point condition.
- Gaming world (both arms right on exactly 80 of 91 clusters; Jev abstains on the 11 it would get wrong, u = 0.121): **INCONCLUSIVE**, the guard `gain_all` (= 0.0 exactly) the only failing condition, `gain_res` = 0.1375 reported.
- R4-05 concentrated-gain null (B2 wrong on all 18 clusters of the two largest PRs 16403 and 16584, Jev right everywhere; 40 PRs): **INCONCLUSIVE**; gain_res point 0.198 and P_assign pass; spread fails with `n_jev_only_prs` = 2; `lb90` = 0.0 equals the independently recomputed `pr_cluster_bootstrap` bound (P(resample misses both PRs) = 0.129).
- Retained and re-asserted under the new rule: redrawn-primary null (B2 accuracy 0.055 → NO-GO, gain_res −0.033, P_assign 0.011), common-accuracy null (both arms 0.95 → NO-GO; P_assign 0.904 passes; failing: gain_res point, its bootstrap bound, the guard), copy-B2 null (gain_res = gain_all = 0 → not GO).

## Choices where the council offered one (recommendation taken, listed in PREREGISTRATION §14 item 16)

- **R4-03 → R4-13:** the user chose to revise the GO rule; R4-03's fix items (1)–(4) are superseded (the disclosure lives in §9 / §11 as the rule's operating characteristics; the `user_decision` trigger is R4-13 (7)).
- **Cap and guard kinds (Addendum (1)):** the unresolved-rate cap 0.15 is a point condition (breach → NO-GO); the guard (`gain_all` > 0) is INCONCLUSIVE.
- **Futility (Addendum (5)):** only the gain term changes (`gain_res`); no new disjunct on `u`; `u` and `gain_all` reported per window.
- **Power bars (Addendum (6)):** the council's bars; the drafter reports measured cells and moves none.
- **`real_none_proxy`:** INFO only; no longer a `user_decision` trigger.
- **R4-02 cap breach:** an integrity pause, not counted towards futility.
- **R4-09:** the council's HD6 state map.
- **R4-06 departure (a):** accepted as conditional; the consumer now checks `input_sha256`.
- **R4-07:** `required_gold_ids(0)` = D ∪ P ∪ every non-P2 control (the council's rule replaces the rev-4 "labelled or attempted" rule).

## Implementer departures (for council round 5)

1. **R4-02 file writer:** `rounds/R/clarify-refused.json` is written by `label_x2d.clarify_precheck` (where cap, used and planned are in scope), not by `stage_clarify`; `run.sh clarify` still exits non-zero (`set -e`). The refusal does not block a later successful `clarify` (the existing refused-then-clarified e2e flow keeps working); a round can then hold both the file and a revision, which is harmless.
2. **R4-13 (8) worlds:** the abstention worlds rewrite the primary arm's *composed* rows (`below_t_choice`) in copies of the finished run; the stub Jev server (`validation/stub_jev.py`) was not given an abstention share. Everything downstream (results, informativeness, verdict, the bootstrap) is the real pipeline; the thresholds / compose path for abstentions is covered by the existing composition fixtures.
3. **R4-13 (3):** `paired()` has the `resolved_only` mode; `pooling_rows` carries both the resolved and the all-cluster figures in one row set (`n_m, a, ac, gc, gc_res, abst, n_all, d_all, n_res, d_res, b2_res, abst_b2_n, abst_b2_correct`) instead of a second mode. The old `gain_stat` over `paired()` rows is kept for the descriptive tables; the GO rule uses `gain_res_stat` / `gain_all_stat`.
4. **`power_sim` exit code** is 0 when a binding cell is missed (above).
5. **R4-04 FAIL rule** also fails when a `none` is not `none_correct_gap` (`gap_ok != none_n`), as the council text says.
6. **Verdict planted positives:** `pooled.planted_new` in `verdict.json` is built from the rounds' trap-control `arm_none` rates (the frozen-round scored units contain no controls); `unresolved_by_round` is reported beside the pooled `unresolved` block.

## Per-issue

| id | status | what was done (file) |
|---|---|---|
| R4-01 | fixed | `run.sh stage_gold` (l.182-185) and `stage_gates_label` (l.296-297) call `gold.py` per file only when the file is absent or empty, echoing "exists (append-only): skipped"; `gold.py` still refuses. Fixtures (`validation/r4_checks.py`): round 1 with main gold and `rounds/2/membership.json` present and a contrast-window unit unlabelled refuses, then after `fill` + `labels-check` writes the contrast gold only (main gold and membership byte-identical); round 0 likewise via `gates-label --new-round fill`. README reproduce note. |
| R4-02 | fixed | `label_x2d.clarify_precheck` writes create-once `rounds/R/clarify-refused.json` {utc, reason, cap, used, planned} and exits non-zero; `run.sh require_labels_check` and `analysis_x2d --labels-check` treat it like a recorded revision (round proceeds scored and non-informative); cap breach is an integrity pause (PREREGISTRATION §8, §10 (v), §12; `thresholds.json` `integrity_pause_note`; `run.sh` comments; README). Fixtures: refused pre-check → file, `labels-check` exits 0, `gold` proceeds, `results.json` `informative.labels_check_ok` false, the file cannot be written twice. |
| R4-03 | superseded by R4-13 (user decision) | see the user decision record; the disclosure is §9 / §11 / README "Operating characteristics of the GO rule"; the measured cells above. |
| R4-04 | fixed | `analysis_x2d.gates()`: below 8 label-derived gap positives P2 is **FAIL** when `len(gap_pos) − none_n ≥ 2` or `gap_ok != none_n`, else UNMEASURED (counts either way); §5; RESULTS wording "R_new on catalogue gaps has no planted test". Fixtures: 0 of 7 → FAIL; 6 of 7 → UNMEASURED; 7 of 7 → UNMEASURED (`validate.py` n1/gates block). |
| R4-05 | fixed | e2e concentrated-gain null (`e2e_abstention_worlds`): INCONCLUSIVE, gain point and P_assign pass, `n_jev_only_prs` 2, `lb90` = independently recomputed bootstrap bound (0.0); §11 names it. |
| R4-06 | fixed | `tune._check_v7_scores` compares the header `input_sha256` with the SHA-256 of `jev-v7/units.jsonl` ("the v_7 units file is not the one the re-score read"); fixture: one byte of `jev-v7/units.jsonl` changed after scoring → `units 8` / `tune 8` refused, nothing written; §14 item 15 (a). |
| R4-07 | fixed | `gold.required_gold_ids(0)` = D ∪ P ∪ every non-P2 control of `units.jsonl`; fixture: one never-started control → `gold 0` refuses naming it, fill, success; §14 item 15 (f), §5. |
| R4-08 | fixed | `x2d.critique_problems(0)` requires `frozen_go_population: 8=<n>/<prs> 9=… 10=…` (from `k23.json`, core-gate excluded, equal to the recorded values) and `user_decision: proceed` when any frozen round has < 20 clusters or < 6 PRs (or a power check is false); §5; fixtures in `r4_checks.critique_checks`. |
| R4-09 | fixed | PREREGISTRATION §14 HD6 state map (run commit → `preregistered`; round 1 → `under-test`; GO → `supported`, NO-GO incl. futility → `not-supported`, INCONCLUSIVE → `inconclusive`; an integrity stop → `inconclusive` with the reason), effect restated as "`gain_res` with its 90 % bound, `u` with its interval". |
| R4-10 | fixed | every `summarise` numerator is read with `.get` (`analysis_x2d.summarise`); fixture: an arm with no assignment over covered gold → `r_assign_res.point == 0.0`, `verdict()` completes; an arm with zero resolved clusters → `n_res` 0, INCONCLUSIVE. |
| R4-11 | fixed | `tune._check_round_scores` (r ≥ 1) refuses before writing when `compose.missing_required(round scores, window units, catalogue area ids)` is non-empty ("re-run jev R"); fixture: header-only `jev/scores.jsonl` → `tune` refused, nothing written. |
| R4-12 | fixed | `analysis_x2d.py` docstring (labels-check exit rule), `README.md` l.41 / the cap formula with the `C` and `I` terms, `thresholds.json` `kappa_basis` (GO population, all-unit κ reported) and `pre_run_stop` note (P3 / power shortfall moved out of `no_go`); comments in `run.sh` / `x2d.py` / `tune.py`. No numbers changed. |
| R4-13 | fixed | `analysis_x2d`: `go_verdict` conditions a–l and precedence; `summarise` keys `p_assign, r_assign_res, r_assign, gain_res, gain_all, n_res, unresolved, abstained, real_new, planted_new`; `paired(resolved_only=)`, `pooling_rows`, `gain_res_stat`, `gain_all_stat`, `unresolved_block`, `abstention_table`; `futility` on `gain_res` with `u` and `gain_all` reported; `round_results` / `report.md` / pack abstention section; `verdict()` per-round `u`, planted figures. `x2d.MINIMUMS` {P_assign, R_assign}, `UNRESOLVED_RATE_MAX` 0.15, `RESOLVED_MIN` 50, `NEW_AREA_GATE_MIN` 0.80. `tune.composed_dev` enforces P_assign, R_assign_res and the dev `u` ≤ 0.15 (P_new / R_new recorded). `power_sim.py` rev 5 (schema v4, cells above). `thresholds.json` (`minimums`, `go`, `new_area`, `futility`, `power`), PREREGISTRATION §9, §10 (iv), §11, §13, §14, README, HD6. Validation: GO-rule unit tests (cap breach → NO-GO even when b/e/f pass; guard alone → INCONCLUSIVE; n_res 49 → INCONCLUSIVE; P_new 0.3 on 12 positives → no effect), the hand-made 8-cluster world, spy checks on `simulate`, the four e2e worlds. |

## For the council / main agent

- **N4b is missed** (above). Routed, not moved. The guard `h` (`gain_all` > 0) does not suppress noise-driven GO under oracle self-abstention; a stricter guard (e.g. a bootstrap lower bound > 0 on `gain_all`) would change the specified rule and is for the council.
- **HD6 wording:** the ledger sentence still says "detects genuinely new concerns with precision and recall ≥ 0.80"; under the user's decision new-area detection is descriptive. Left as instructed; reword if the council agrees.
- **README** "Rev 3 additions" paragraph still mentions `power_sim.simulate_inside` and `p3_inside_min` as history of rev 3; the rev-5 paragraph below it supersedes both.

## Digests (SHA-256, recomputed after the last edit of every frozen file; the 44 files of `x2d.frozen_files()`)

| File | SHA-256 |
|---|---|
| `GUIDE.md` | `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980` |
| `PREREGISTRATION.md` | `8310a44bc9edb42645b34977ad332baa0e5240571f541873a03a2c7328fcbcf7` |
| `README.md` | `845d80e27d86b5f12055c5216df9c762509a2dd53e804a00cfe554afaa872462` |
| `analysis_x2d.py` | `5c4e7d84633ed97a76d55e0949c51d3a0c7e9b05e5faca16b738f8f6c42ae645` |
| `b2.py` | `5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f` |
| `baselines.py` | `a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0` |
| `build_controls.py` | `7ded8c9177d8e21131ed91ee0aa423a4dd8c64ac35b215f82e93b1bfbfe6c81b` |
| `build_units.py` | `c052246c828024a25fc6a143d513e111ecbf63b42b8b5912dd37e960a1c3e243` |
| `collect_onboard.py` | `c93fa59c8fe305459336e08e52c60f365d8df0a70d7b6ca395c58e8f88da5fff` |
| `compose.py` | `3b47ed4d495d6c080a83b988c67357fc57d9846b01a8350769de843a6661b603` |
| `frontend.py` | `dd4a0fdc1e3595eeefb82a8eafab600255b3b0fe95f67e2fb8a07eb5972ea0d8` |
| `gold.py` | `6b0afaad2a3ae6652a02207813c5a3492ba6954bd0a745b58a3f98c960bb321e` |
| `jev/main.go` | `63fbec0534a2ddf3d55d07c074c6a882e199ec2b90cf771a984c9fa9ee586ca0` |
| `label_x2d.py` | `82fc06bede3b93ee7348d557290d4ed761d77c4d953e80cb981f00662f5250c4` |
| `measurement/lever-result.json` | `298902e717d0db8826c35d219695d386b19644f99cfc205d6b2d891100b0352c` |
| `measurement/lever_check.py` | `8f8008d3a28d72d5e768bc292e4917449fd8cb31ca0a642ff6057ab51b705a42` |
| `onboard/main.go` | `7e2ed1edc64b694851e7729e17af4f2ba63dafa7c9cdcf19cd3a5a255c08d46b` |
| `power_sim.py` | `6ebc772c7b517cb4f616d1ad2ef3a95c623f255c9fbab1506fbe3a1cfb46ae75` |
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
| `run.sh` | `e85cd7312a329dab60dbdf322949184cae0ac3782f2bacc38e816ecb5b67549b` |
| `scopes.py` | `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072` |
| `thresholds.json` | `3fc6971ad4833c8f6c4c29159781abc68ae4821eb042f2877ce4d6fe156448fb` |
| `tune.py` | `0e24eeaa8de539950f739867c1adc7124eb104b027baacf5648a002d6d45d88a` |
| `validation/r4_checks.py` | `02d5900dec17fd3200aca022dff6f6052037d2c629e1c25cdab9c0b5479c1a3b` |
| `validation/stub_clawpatch_x2d.js` | `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f` |
| `validation/stub_jev.py` | `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d` |
| `validation/stub_labeller.py` | `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5` |
| `validation/validate.py` | `bb34a6939bdc2af116211fc25a778970d46aca0c7c77b72047939ad12c34791f` |
| `versioning.py` | `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168` |
| `x2d.py` | `352b9b2400c5dd11f9011cc1aac589ded11f8fa0c31f04b7b1b96d1f609bdf52` |

Unchanged since round 3: `GUIDE.md`, `b2.py`, `baselines.py`, `build_controls.py`, `build_units.py`, `collect_onboard.py`, `compose.py`, `frontend.py`, `jev/main.go`, `measurement/*`, `onboard/main.go`, `prepass.py`, `prepass/*`, `quality.py`, `questions/*`, `reference_arm.py`, `scopes.py`, `validation/stub_clawpatch_x2d.js`, `validation/stub_jev.py`, `validation/stub_labeller.py`, `versioning.py`. New frozen file: `validation/r4_checks.py`. External pin: `tune_threshold.py` (`~/.omp/agent/skills/jev-actions/scripts/tune_threshold.py`) = `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c` (recomputed 2026-10-30; pinned in `signoff.json` at sign-off).

Reproduce: `source .tools/activate && evaluation/x2d-merge-association-pilot/run.sh validate`; digests: the command at the end of `DRAFT-NOTES.md` §5; power: `X2D_SKIP_SIGNOFF=1 python3 -c "import sys; sys.path.insert(0, 'evaluation/x2d-merge-association-pilot'); import power_sim; print(power_sim.run(pilots=1000, boot_B=1000)['checks'])"` (deterministic; the table above).
