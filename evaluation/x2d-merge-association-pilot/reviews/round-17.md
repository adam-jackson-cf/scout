# X2d — council experiment review, round 17 (rev 13: the one permitted re-gate of P1 and P2 on fresh items, R-053)

Date: 2026-10-30. Reviewer: innovation council. Opus 5.5 wrote the lead synthesis. Two peers worked read-only and received the inputs without the lead's leaning:
- `innovation-challenger` = AdversarialRegate (GPT Astra);
- `experiment-peer` = ConfoundRegate (Opus 5.5).

Governance: `CLAUDE.md` experiment rules 1–7; user decision R-053 (option (c) of `reviews/pilot-round-0-gate-stop-memo.md`).

Scope:
- `PREREGISTRATION.md` revision 13 and its implementation: `x2d.py`, `build_controls.py`, `gold.py`, `analysis_x2d.py`, `run.sh`, `validation/validate.py`, the new `validation/regate_checks.py`, and `README.md`;
- the inputs: memo §1, §5 and §7; `reviews/pilot-round-0.md:105-210`; `reviews/regate-supply.{md,json}`; `reviews/rev13-draft-response.md`; `/tmp/validate-rev13.log`;
- the review questions of the brief: repairs vs memo §1, freshness, supply and population, bars, P2 binding, fixtures (rule 4), one shot and the round-1 refusal, round-0 bytes, budget, the HD6 clause, and nothing else moved.

Reviewed files (SHA-256): `DRAFT-NOTES.md` `f3c45bd5a25fe87a9ec0bd8f272172549f343243dc21ecf167ad8231bc57d46f`; `GUIDE.md` `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980`; `PREREGISTRATION.md` `d4cdad866ccbe2f15ae5de7fda3aa885130389ddf9e22cf362bab89c0f8d6751`; `README.md` `31f9dded37431713ef92780a646ec71ac05e6879d568da4539588c4c61f8f456`; `analysis_x2d.py` `37fe3712abfda315dd665db03d67509971ae251a36582f964fa3617c48d1e8a4`; `b2.py` `5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f`; `baselines.py` `a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0`; `build_controls.py` `efbcbf0636e2b157ee60af21d48c3b1feb845cfd84cb0914c0eda390f4e67990`; `build_units.py` `c052246c828024a25fc6a143d513e111ecbf63b42b8b5912dd37e960a1c3e243`; `collect_onboard.py` `c93fa59c8fe305459336e08e52c60f365d8df0a70d7b6ca395c58e8f88da5fff`; `compose.py` `296001923384652d7d31cdb33ca937ed409783023de2b6ee53c62e4ea8ab5bd1`; `frontend.py` `dd4a0fdc1e3595eeefb82a8eafab600255b3b0fe95f67e2fb8a07eb5972ea0d8`; `gold.py` `a7831d65d40c0fa9ee15555a8488532dc646990725ac47a670f1fb6fec6bb8e5`; `jev/main.go` `63fbec0534a2ddf3d55d07c074c6a882e199ec2b90cf771a984c9fa9ee586ca0`; `label_x2d.py` `fa5d60ef2bbf2147102cf986048f7711cd2ad90d8cd7a0ddbe89aa5162729dcb`; `measurement/lever-result.json` `298902e717d0db8826c35d219695d386b19644f99cfc205d6b2d891100b0352c`; `measurement/lever_check.py` `8f8008d3a28d72d5e768bc292e4917449fd8cb31ca0a642ff6057ab51b705a42`; `onboard/main.go` `7e2ed1edc64b694851e7729e17af4f2ba63dafa7c9cdcf19cd3a5a255c08d46b`; `power_sim.py` `a955f7177280bfd6d20a2514d21a04819c426610d594ee4e7dbfe0cdfa9c5279`; `prepass/commits.jsonl` `1581986e0efbf26bb51aea894c3ffcdd23f54ef6e1d6d6ab0e4e99248f35bd1c`; `prepass/k0.json` `f89e8d0b264eb33eb8bcb5aa4a4910bd462c3e1befea480b721ee514947a98ab`; `prepass/reserve.json` `a73558ed54efdd767c5c0cf9a5accdcf41f511acd044d630208a8fb3e129f12a`; `prepass/rounds.json` `0aad85b0199e747630d0e863b7a5e83b2624684c50c1150ee4e8775fde1d7516`; `prepass/scopes.json` `a24bfc8d19aab9fd5ab0817f317cee560df65c63b1d63b5dbd10b5529d16a512`; `prepass/summary.json` `e1cfced0da45bcd86cc516eef39fa70c10b3fa5d630820d415cdea47045cc1c1`; `prepass.py` `1e5863d32974f1d7fecfd4df5c30b6e02aa137b9db491780eab11cc00b500696`; `quality.py` `6d65293d64c76611368fa8eaf93d4f48432cf203bb65aed9ed63130b8527cade`; `questions/merge.area_covers@v1.json` `cce59e612b67d76b96d927613c1a780ea64915c38210b7be8fceb5cfaa7e5af0`; `questions/merge.covered@v1.json` `f37eea94bb6e2c720cbbb904066dd06c368ecc7581c90063b9f17e5a751058b9`; `questions/merge.which_area@v1.json` `cbb7ba7e92432a1e0ffc9bf7d8718a9001ba4227137a6ab60a1e359393eff928`; `questions/onboard.covered@v1.json` `ec4d62be051c016ab945738c155401a17f091654e1efff9cfd60ec18314eaa70`; `questions/onboard.same_responsibility@v1.json` `31f3b25d3bbf535b026bd6f0a022d39f01e7c519cc22649195d341b75a49a7b1`; `questions/onboard.which_file_area@v1.json` `425edf93d08dba42d7f1da755601d5dec27cded8b67593401b7656b24cf12f57`; `reference_arm.py` `d139f0179c06ca5c6aff5784ceab2feb4113f888c51cd46f760f7a3d943bb739`; `run.sh` `99eb7422022241e1836fa3a9c6dbc84a8976d5f8e00fefaddcb78e1f1079c754`; `scopes.py` `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072`; `signoff.json` `578a2462a6f25953c2e744a1d3dcc546f58af868f2e0e343e204367a89552e61`; `thresholds.json` `7142d7ac15abbcd27b75ac8005a7f1bffed0c6a70395c65e8f631bfcd32995eb`; `tune.py` `2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c`; `validation/r4_checks.py` `f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52`; `validation/regate_checks.py` `e87f99f677dc149d2f798a5eac43f953bf69fc6d5acbc57fed36a7524c737083`; `validation/stub_clawpatch_x2d.js` `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f`; `validation/stub_jev.py` `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d`; `validation/stub_labeller.py` `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5`; `validation/validate.py` `8fab5ee017b8eecd8b0dc13dacb45edf9d7383e8586334ec7f994bf764ee16c9`; `versioning.py` `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168`; `x2d.py` `0ae78958c2d39e424cbc417c8c4250db0643934658f11a82146f663fac394b00`

Rule 3 check: no predictor–outcome association has been computed. The council read the following, all read-only:
- in `../scout-x2d/…/data/rounds/0/`: `control-key.jsonl` (the `kind`, `original_paths` and `target_dir` fields only) and the `units.jsonl` member lists of controls. These are membership, not answers;
- `reviews/regate-supply.json`, which is membership only;
- `/tmp/validate-rev13.log`.

The round-0 per-item Jev answers are already disclosed in `reviews/pilot-round-0.md`. They were not re-examined or extended. The council compared no Jev answer with any label or gold.

## 1. Verification

Status: **verified** = as claimed; **issue** = see §2; **not verifiable read-only** = left to the conditions.

| Item | Status | Evidence | Issue |
|---|---|---|---|
| Repairs match memo §1 (1)–(3); justified by facts that do not involve Jev | verified, with disclosure gaps | P2 direct-label gold, with derived gold reported beside it: `gold.py:45-79`. "Any `none` counts" and twins excluded: `analysis_x2d.py:1491-1529`. P1 acceptable-set gold that must contain A, with exclusions: `:1441-1462`. Leak fix: `build_controls.py:228-255` (`leak_check`). Accuracy on answered items plus a separate abstention cap: `:1478-1480`, `:1529`. Every repair rests on a label, ownership or construction fact (memo §2). The bar rationale in `PREREGISTRATION.md:112-113` overstates P1 and is silent on P2 | R17-04 |
| Fresh means fresh: no exposed path, twin or PR episode | **issue** | `exposed()` (`build_controls.py:435-448`) covers the rev-12 source units, their members, the P1 `original_paths`, their PRs and the guide PRs. Neither it nor `_p1_target` checks the paths or directories at which the round-0 controls were *presented*. Three planned items are presented at the exact paths of rev-12 P1 items. Six or more reuse rev-12 target directories, among them three where round 0 showed a path pull or a near-duplicate split (§2) | R17-01 |
| Twins and PR episodes among the sources | verified | Used source → twin → PR episode → guide PR, then one cluster per PR and no shared path (`build_controls.py:500-528`). The funnel in `regate-supply.md` agrees | — |
| Supply and sizes | partly verified | The membership counts reproduce from `regate-supply.json`: pool 129; P1 40 (strata 34/2/4); P2 40 clones in 16 areas. But all 40 P1 items and 36 of the 40 P2 clones are `in_no_unit` clusters of PRs < #10000, which is 2019–2022. Only four clones carry 5-digit PRs (`regate-supply.json:642,660,768,777`). The response calls this "some pre-2023". The direct-label positive yield is unmeasured, but `regate-supply.md:43,47` reports "met" | R17-02, R17-03 |
| Bars derived only from existing bars | verified | `x2d.py:51-59`: `REGATE_P1_ACC_MIN = MINIMUMS["R_assign"]` (0.80); `REGATE_*_ABSTAIN_MAX = UNRESOLVED_RATE_MAX` (0.15); baseline 0.20; `REGATE_P2_NONE_MIN = NEW_AREA_GATE_MIN` (0.80); floors 20 and 8. None comes from round-0 data. Rulings in §3 | (text: R17-04) |
| Fixtures exercise the real code (rule 4) | verified | Unit fixtures call `analysis_x2d.regate_gates` and `gold.build_gold` on synthetic worlds (`regate_checks.py:116-271`). The end-to-end check runs `run.sh regate-build/-label/-jev/-analyse` through `make_regate`, `baselines.py`, `label_x2d.py`, `gold.py`, the Go driver against the stub and `compose.py` (`:644-738`). Both gates have a planted effect (PASS) and a planted null (FAIL); the null also refuses round 1 (`:731-734`) | — |
| One shot; round-1 refusal | code verified; procedure missing | `regate-build` refuses on an existing `regate/` (`run.sh:349-352`). Label, Jev and analyse refuse once `gates.json` exists (`:343-346`). `--regate-gates` is create-once (`analysis_x2d.py:1719-1730`). `--gates` refuses inside the root (`:1731-1733`). `gate_problems` needs the tag and PASS on both (`x2d.py:262-291`). Deleting `regate/` erases the guard and the spend record together, and nothing records the build outside it | R17-07 |
| No existing `rounds/0` byte rewritten | re-gate stages verified; legacy stages open | `write_regate` writes only under the new root (`build_controls.py:702-735`). Every later stage runs under `in_regate`, and the e2e manifest check covers this. The legacy `gates-label` and `gates-jev` stages are still callable after the stop and overwrite `rounds/0/sessions.json` (`label_x2d.py:470-478,612`), `labels-check.json` (`analysis_x2d.py:1700-1707`) and `jev-in/.runs` (`run.sh:317-318`) | R17-06 |
| Budget charge | verified | `jev_budget` maps the root to the study data (`x2d.py:617`). `score_files` globs `rounds/*/**/scores*.jsonl` recursively, so re-gate spend counts against round 0's $2 and the $20 total. Spent so far: $1.0985 (`pilot-round-0.md:80`). Rev-12 `p1` cost $0.025 for 20 units and `p2` $0.0149 for 12, so about $0.10 for 80 units [INFERENCE] | — (challenger point rejected, §3) |
| HD6 map clause | verified, wording to extend | `PREREGISTRATION.md:253`: a stop → `inconclusive`, never `not-supported`; a re-gate PASS leaves `preregistered`. It needs the status-specific reason and the pre-Jev shortfall case | R17-02 |
| Run commits | **issue** | `PREREGISTRATION.md:250` (item 19) still assigns "rounds 1–10 and `verdict`" to the rev-12 commit, and `README.md:15` does the same. Item 20 (`:251`) does not supersede this | R17-05 |
| Nothing else moved | verified | The 45-row table (`rev13-draft-response.md` §5) against `signoff.json` (round 16): 36 rows are equal. That includes every Jev-side file (`GUIDE.md`, the six `questions/*`, `thresholds.json` `7142d7ac…95eb`, `compose.py`, `jev/main.go`, `label_x2d.py`, `baselines.py`, `b2.py`) and the state builders (`build_units.py`, `frontend.py`, `b2.py`), so round 0's `version.json` does not drift. 8 rows changed: PREREG, README, `analysis_x2d.py`, `build_controls.py`, `gold.py`, `run.sh`, `validate.py`, `x2d.py`. 1 row is new: `validation/regate_checks.py` | — |
| Validation (a) | verified from the log | `/tmp/validate-rev13.log` ends "707 checks, 0 failed" and "all validation checks passed", with no line starting `FAIL`. Record digest `578a2462…2e61` before = after. That the run used exactly these bytes is not verifiable read-only | condition |

## 2. Issues

### R17-01 (major) — Fresh P1 items reuse the presented paths and target directories of round-0 controls

- **Evidence.** `exposed()` (`build_controls.py:435-448`) has no entry for presented paths or target directories. `_p1_target` (`:228-255`) checks only the C0 tree, sibling keys, rules (a)–(f) and the leak. `_p1_placements` seeds `taken` empty (`:549`). The supply record places:
  - cluster `1cb7d6c0` (stratum 1, decisive) at `docs_src/settings/app01/tutorial001.py` (`regate-supply.json:374-385`), the presented path of ctl-20 (`control-key.jsonl:20`, target `docs_src/settings/app01`, original `docs_src/request_form_models/tutorial001.py`);
  - `7d84155c` at `tests/test_tutorial005.py` (`:220-231`), the path of ctl-31 (`:31`, target `tests`);
  - `2439717f` at `tests/test_modules_same_name_body/test_tutorial001.py` (`:120-131`), the path of ctl-1 (target `tests/test_modules_same_name_body`, also ctl-23's leak directory).
- **Directory reuse.** Further items reuse `tests/test_tutorial/test_header_params` (`:428`, ctl-24), `.github/actions/people` (`:516`, ctl-33) and `tests/test_modules_same_name_body` again (`:406`). The council has seen Jev's round-0 behaviour on exactly these directory cues.
- **Why it matters.** P1 is a path-cue test. The memo's freshness rule (§1, §7 step 4: "no exposed paths") is broken for a decisive stratum-1 item and for several cues with known outcomes. Selection was mechanical, so this is exposure, not steering. It still makes the re-gate partly a re-test of seen conditions.
- **Fix.**
  1. `build_controls.exposed()`: add `"presented"` = the members of every `units.jsonl` row of round 0 with `control: true`, and `"target_dirs"` = `{k["target_dir"] for k in control-key rows with kind "p1"}`.
  2. `regate_plan` and `make_regate`: pass both sets to `_p1_placements(…, forbid_paths, forbid_dirs)`. That function starts `taken = set(forbid_dirs)` (which `_candidate_dirs` then skips for every stratum) and passes `forbid_paths` to `_p1_target`. There a directory is skipped when any new path is in `forbid_paths`.
  3. Record both counts and the placements rejected for them in `supply["exposed"]` and the funnel.
  4. `PREREGISTRATION.md:111`, in the **Excluded** list after "the PRs named (`#<digits>`) in the guide;", insert: "a P1 placement whose presented path equals that of any round-0 control, or whose target directory was the target directory of any rev-12 P1 item (the rev-12 target directories seed the never-reuse set);".
  5. Fixture (G13): in a data copy, plant a rev-12 P1 control whose target directory and presented path are the only stratum-1 host of a pool candidate. The plan must not place the candidate there; it falls back or excludes it. The check is specified independently of the code.
  6. Regenerate `reviews/regate-supply.{md,json}` (membership only). If the minimums (40 P1, ≥ 20 stratum 1, 40 clones) are no longer met, the build refuses and the drafter escalates. Sizes are never shrunk.

### R17-02 (major) — A label-only shortfall stops the pilot only after Jev has run; the P2 sizing claim is overstated

- **Evidence.**
  - The P1 denominator (eligible stratum-1 items ≥ 20) and the P2 denominator (direct-label positives ≥ 8) are fixed by the gold and the membership alone (`analysis_x2d.py:1441-1462`, `:1491-1508`). The same holds for `rules_resolved_units` and the B1/B2 rates over eligible items (`:1468`, `:1477-1479`).
  - Yet `regate_inner_jev` checks only that `gold.jsonl` exists (`run.sh:363-366`). It then spends Jev, and `regate-analyse` records UNMEASURED or FAIL. That stops the pilot for good (`PREREGISTRATION.md:110,116`).
  - Memo §5 requires the opposite for supply: "If supply falls short of the rev-13 sizes, escalate to the user rather than shrink the sizes."
  - The P2 sizing (`PREREGISTRATION.md:113`, `x2d.py:56`, `regate-supply.md:43,47`, "met") treats the 40 clones as independent. They share 16 ablations (4, 4, 4, 3, 3, 3, nine of 2, 1). With yield fully correlated within an area, P(X ≥ 8) ≈ 0.88 at a yield of 1/3 and ≈ 0.71 at 1/4 [INFERENCE: normal approximation, Σn² = 112].
  - The round-0 yield came mostly from areas owning one file. The re-gate ablates areas owning at least 2 candidates. Area 9a4a0e, whose round-0 clone was labelled covered by a near-duplicate (a label fact), contributes 4 clones.
- **Why it matters.** A shortfall in labels (supply, not Jev) would end the pilot as "did not pass", with HD6 `inconclusive`. That is exactly the "mis-built gate decides" outcome R-049 and R-053 set out to avoid. The supply record also claims a positive count that membership cannot establish.
- **Fix.**
  1. `analysis_x2d.py`: factor the label-only part of `regate_gates` into `_regate_denominators(d)`. It reads only `units.jsonl`, `control-key.jsonl`, `gold.jsonl`, `arms/B1.jsonl` and `arms/B2.jsonl`, and `regate_gates` uses it unchanged. Add `regate_label_problems(data=None)` and the CLI flag `--regate-precheck` (inside the regate root only). It lists:
     - eligible stratum-1 P1 items < `REGATE_P1_STRATUM1_MIN`;
     - P2 direct-label positives after twin exclusion < `REGATE_P2_POSITIVES_MIN`;
     - `rules_resolved_units` > 0;
     - B1 or B2 over eligible items > `REGATE_BASELINE_MAX`.

     It prints the counts and exits 1 on any problem. It never opens `arms-p1/`, `arms-p2/` or `jev-p/`.
  2. `run.sh` `regate_inner_jev`: after the gold check, add `py "$S/analysis_x2d.py" --regate-precheck || { echo "refusing: the re-gate's label-only denominators or construction checks fail (no Jev request was made): stop, report; the council and the user decide (PREREGISTRATION 5)" >&2; exit 1; }`.
  3. `PREREGISTRATION.md:110`: replace "a re-gate that is not PASS for BOTH P1 and P2 (FAIL, or UNMEASURED, which does not pass) stops the pilot" with: "after gold and before any re-gate Jev request, `regate-jev` checks the conjuncts that need no Jev answer: ≥ 20 eligible stratum-1 P1 items, ≥ 8 direct-label P2 positives, `rules_resolved_units` = 0, and B1 and B2 ≤ 0.20 over the eligible P1 items. A miss is a supply or construction finding like a build refusal, not a gate result: no Jev request is made, the drafter reports, the council and the user decide, and HD6 stays `preregistered` meanwhile. A re-gate that passes this check and is then not PASS for BOTH P1 and P2 stops the pilot".
  4. Mirror the change at `:116`. Make the UNMEASURED clauses of `:112-113` read "(a defensive status; the pre-Jev check makes it unreachable)".
  5. HD6 map (`:253`): add "a pre-Jev re-gate shortfall leaves HD6 `preregistered` pending the user's decision". Change the stop reason to "round-0 gates P1 and P2 failed and the one permitted re-gate did not pass (re-gate P1 <status>, P2 <status>); HD6 never under test".
  6. Replace the binomial sentence at `:113` and in the `x2d.py:56` comment with: "P(X ≥ 8) = 0.98 (0.82 at 1/4) for independent clones; the clones share 16 ablations, and with yield fully correlated within an area P(X ≥ 8) ≈ 0.88 at 1/3 and ≈ 0.71 at 1/4 [INFERENCE]; a shortfall is caught before any Jev request (above)".
  7. `regate-supply.md:43,47`: replace "minimum 8 positives with margin: met" and "Both minimums are met" with "construction counts met; the direct-label positive yield is unmeasured until labels (sizing above is an [INFERENCE])".
  8. Add `"positives_by_ablated_area"` to the P2 value (`analysis_x2d.py:1515-1522`).
  9. Fixtures:
     - unit: worlds with 19 eligible stratum-1 items, 7 positives, B1 at 9 of 40 and one rules-resolved item each give a problem, and the bounds pass. The check runs with no `arms-p1/`, `arms-p2/` or `jev-p/` present;
     - end to end: a stub labeller with a planted low yield makes `run.sh regate-jev` refuse, with no score file under `regate/rounds/0/jev-p/`, no `regate/gates.json`, and the study manifest unchanged.

### R17-03 (moderate) — The fresh population is entirely 2019–2022 legacy files, and the preregistration does not say so

- **Evidence.**
  - All 40 planned P1 items and 36 of the 40 P2 clones are `in_no_unit` with PRs < #10000. The reserve already holds every 2023→C0 residual cluster whose files exist at C0 (`PREREGISTRATION.md:100`).
  - Item text and PR text come from the adding commit (`build_controls.py:266`, `:645`), 2–6 years before C0.
  - `:111` reports origins only. `rev13-draft-response.md` §6 Q6 says "some pre-2023".
  - Also, `:111` says the pool replays "every commit up to C0". The code replays only the most recent first-parent `--diff-filter=A` addition of each eligible C0 file, plus the reserve commits (`build_controls.py:450-476`).
- **Why it matters.**
  - Older text against a C0 catalogue can move the labellers' gold away from A, which means exclusions. It can also lower Jev's confidence, which means abstention.
  - Old, central files may instead be easier, because the catalogue was written around them.
  - RESULTS could present the re-gate as a re-test of the rev-12 population or of window-era files.
- **Fix.**
  1. `PREREGISTRATION.md:111`, after "(4 per area at most).", insert: "All planned P1 items and all but 4 P2 clones are `in_no_unit` clusters of 2019–2022 PRs (every 2023→C0 cluster is already in the reserve), shown with their text and PR at the adding commit, 2–6 years older than C0 and than the rev-12 P items. The re-gate tests the repaired instrument on legacy files; `RESULTS.md` states this beside every re-gate figure and never presents it as a re-test of the rev-12 population or as evidence about window-era files."
  2. Replace "replays the front end (§3) on the first-parent commit that added each file and keeps the residual clusters of every commit up to C0" with "replays the front end (§3) on every first-parent commit up to C0 that is the most recent addition (`git log --first-parent --no-renames --diff-filter=A`) of a C0 file owned by exactly one catalogue area, and on the reserve's commits, and keeps the residual clusters".
  3. Add `"date": r["date"]` to the P1 and P2 item dicts of the supply record (`build_controls.py:615,618`).
  4. Carry `date` and `origin` into `p1_items` and `p2_items` of `gates.json`, so RESULTS can tabulate figures by year and origin.

### R17-04 (moderate) — The bar rationale claims P1 is "never easier" (false) and is silent that P2 is easier on abstention

- **Evidence.**
  - `PREREGISTRATION.md:112` ends: "so the repaired gate is never easier than the rev-12 gate". But the gold widened from {A} to the agreed set that contains A, items whose gold excludes A are removed, and the 0.68 floor holds on eligible stratum-1 items only. Memo §2's sensitivity shows that widening the gold changes the score.
  - P2: rev 12 required `none` on ≥ 0.80 of all gap positives, with abstention counted as a miss (`analysis_x2d.py:1373-1394`). Rev 13 needs ≥ 0.80 of the *answered* positives plus abstention ≤ 0.15, which implies only ≥ 0.68 of all positives. With 8 positives, 6 `none` + 1 assigned + 1 abstained passes rev 13 and failed rev 12. Neither `:113` nor the change log `:15` says so.
- **Why it matters.** RESULTS would inherit a false claim about difficulty, and an undisclosed easing of the vetoing gate.
- **Fix.**
  1. `:112`: replace "so the repaired gate is never easier than the rev-12 gate." with "so on eligible stratum-1 items the bars imply an acceptable-set recovery of at least 0.68, above the rev-12 bar of 0.60; this is not dominance over the rev-12 gate: the gold is wider than rev 12's single A (memo §1 (2)) and the population and strata differ, so `RESULTS.md` reports the A-only recovery beside it."
  2. `:113`, after "(`REGATE_P2_POSITIVES_MIN`, the rev-12 floor).", insert: "This is easier than rev 12 on abstention: at abstention 0.15 the bars imply `none` on ≥ 0.68 of all positives against rev 12's 0.80 with abstention a miss. Accepted because an abstention is flagged, not the silent wrong assignment P2 guards against (§9), and it follows R-051 (memo §1 (3)); `RESULTS.md` reports the `none` rate over all positives beside the decisive figures."
  3. `_p1_stats` (`analysis_x2d.py:1411-1419`): add `"recovery_all_items_A_only"`, meaning assigned to the item's `area_id` over all items, and `"sibling_in_acceptable"`, meaning the count of items whose `sibling_area_ids` meet `acceptable`.
  4. P2 value (`:1515`): add `"none_rate_all_positives": _rate(len(none_pos), len(pos))`.
  5. Fixtures: extend `p1-bacc` to assert both new P1 fields. Extend the P2 bounds check to assert `none_rate_all_positives` (for example 17/20 `none` with 3 abstained → 0.85).
  6. Change log `:15`: add "P2's decisive rate moves to answered positives with an abstention cap (easier on abstention; disclosed)".

### R17-05 (moderate) — The run-commit statement now contradicts rev 13

- **Evidence.**
  - `PREREGISTRATION.md:250` (item 19): "the re-freeze and every stage after it (the rest of `gates-jev`, `gates-analyse`, `power`, rounds 1–10 and `verdict`) run at the rev-12 sign-off commit … names the three run commits".
  - `README.md:15` says the same ("every later stage at the rev-12 sign-off commit").
  - The rev-12 commit has no re-gate stage and no rev-13 `gate_problems`, so it cannot run the re-gate or round 1. Item 20 (`:251`) is silent on this.
- **Why it matters.** It is a frozen statement made false, and RESULTS `Deviations` would name the wrong commits.
- **Fix.**
  1. In item 19, after "rounds 1–10 and `verdict`", insert "[superseded in rev 13 for every stage after `power`: item 20]".
  2. Append to item 20: "**Fourth run commit.** Item 19's rev-12 commit ran the round-0 re-freeze, `gates-jev`, `gates-analyse` and `power`. The re-gate stages, the round-0 critique, rounds 1–10, `rescore-v7`, `verdict` and `RESULTS.md` run at the rev-13 sign-off commit, to which the run worktree moves by checkout only (no `git clean`, no `stash -u`; `data/` untracked). `RESULTS.md` `Deviations` names the four run commits (`730674e`, `0b56594`, `b6a5bf3`, the rev-13 commit) and the stages run at each."
  3. `README.md:15`: replace "the round-0 re-freeze and every later stage at the rev-12 sign-off commit" with "the round-0 re-freeze, `gates-jev`, `gates-analyse` and `power` at the rev-12 sign-off commit, and the re-gate and every later stage at the rev-13 sign-off commit".

### R17-06 (minor) — The legacy round-0 stages can still rewrite round-0 bytes after the gate stop

- **Evidence.** `stage_gates_label` (`run.sh:302-313`) and `stage_gates_jev` (`:317-330`) do not check for `rounds/0/gates.json`. `label_x2d.write_sessions` replaces `rounds/0/sessions.json` (`label_x2d.py:470-478`, called at `:612`). `--labels-check 0` replaces `labels-check.json` (`analysis_x2d.py:1700-1707`). `runs_of > "$R0/jev-in/.runs"` truncates and rewrites that file. `gates-analyse` already refuses (`:1731-1733`).
- **Why it matters.** RESULTS reports round 0's κ and its session count. Running a legacy stage after the stop could silently change those records, and the claim that no existing `rounds/0` byte can be rewritten would not hold.
- **Fix.**
  1. At the top of `stage_gates_build`, `stage_gates_label` and `stage_gates_jev`, after `verify`, add `[ ! -s "$D/rounds/0/gates.json" ] || { echo "refusing: $D/rounds/0/gates.json exists: round 0's gates were analysed and its files are never written again (rev 13; the re-gate lives in rounds/0/regate)" >&2; exit 1; }`.
  2. Fixture (G13): on a finished copy, each of the three stages refuses, and the SHA-256 manifest of `rounds/0` (excluding `regate/`) is unchanged.

### R17-07 (minor) — Nothing outside `regate/` records the build or its result

- **Evidence.** The one-shot guard and the re-gate's spend both live inside `regate/` (`run.sh:349`, `x2d.py:612-619`). `gate_problems` trusts the stored statuses (`x2d.py:283-291`). Nothing compares the built items with the reviewed plan. The R10-03 procedure (`PREREGISTRATION.md:249`) did this for the rev-11 build.
- **Why it matters.** Deleting `regate/` would erase the attempt and its cost without trace. A build that differed from the reviewed plan would also go unnoticed.
- **Fix.** `PREREGISTRATION.md:110`, after the "**One shot:**" sentence, insert: "`run.sh regate-build` runs alone first. The main agent then compares the P1 and P2 item lists of `regate/supply.json` (cluster ids, members, target directories, strata, ablated areas) with `reviews/regate-supply.json` and records, and commits, in `reviews/pilot-round-0.md` the comparison and the SHA-256 of `regate/regate.json` (with its `utc`), `regate/supply.json` and `regate/rounds/0/control-key.jsonl`. Any mismatch is a stop for a new council round, never a re-build. The pre-Jev check output and the SHA-256 of `regate/gates.json` are recorded the same way at their stages. A deleted or rebuilt `regate/` is a protocol breach reported in `Deviations`."

## 3. Rulings

1. **P1 bars: accepted.**
   - 0.80 accuracy on the answered stratum-1 items (`R_assign`), 0.15 abstention on them (g), and B1/B2 ≤ 0.20 over the eligible items.
   - Every item is gold-covered, so answered accuracy is `R_assign_res` on the planted population. Both values are the strictest existing bars, and none is set from round-0 data.
   - The rationale text is corrected by R17-04. Under R17-02 the baseline bar and the floor of 20 are checked before Jev, as construction checks.
2. **P2 stays binding: yes.**
   - It is a round-0 instrument check against the silent wrong assignment (§9), not one of R-051's verdict-level new-area figures. The 0.80 bound already survives as a round-0 gate bound (§14 item 16(i)), and the user adopted (c) "keeping the gates binding" (memo §1).
   - Its bars (≥ 0.80 `none` on the answered positives, abstention ≤ 0.15, floor 8) are accepted, with the easing disclosed (R17-04).
3. **Path-following is report-only: accepted.** No bar exists for it and memo §1 (3) forbids new bars. A wrong composed answer already counts against accuracy.
4. **Strata 2–3 small (34/2/4 at drafting) and report-only: accepted.** Stratum 1 decides (memo §1 (2)).
5. **Abstention knife edge at the frozen provisional 0.70: accepted** as the stated price of gating before tuning (memo §5). The user chose (c) over (d′). RESULTS names which conjunct failed and gives the per-reason abstention split (`reason` in `p1_items`).
6. **Population (119 of 129 pool clusters `in_no_unit`; in fact every P1 item, all pre-2023): valid for these gates, with disclosure.** It is the same construct: a displaced C0-owned file → its area; an ablated area → `none`. The gold comes from blind labels on the same items. It is a different population, so it is not evidence about window-era files (R17-03).
7. **Labelling the 40 P2 source units on the full catalogue (W = 120, cap 330): accepted.** Memo §1 (1) requires the derived gold to be reported.
8. **Yield assumption of 1/3: admissible** (a label fact), but stated as clustered and uncertain (R17-02). With the pre-Jev check, 40 clones need no larger margin.
9. **Budget: accepted.** The re-gate counts against the round-0 $2 cap and the $20 total; about $0.10 of $0.90 headroom [INFERENCE].
   - AdversarialRegate noted that the Go driver refuses only when `spent > maxCost` (`jev/main.go:877`), so a fresh file with zero allowance still sends up to 8 in-flight requests.
   - **Not raised.** This is the preregistered cumulative stop, which always overshoots by in-flight requests (round 0 stopped at $1.0001). `jev/main.go` is frozen and unchanged. The re-gate has about 9× headroom.
10. **HD6 map clause: accepted**, with the R17-02 additions (the pre-Jev shortfall case and the status-specific reason).
11. **Implementer departures:** none beyond the above. The `root-commit parent` fix in the legacy build does not affect any recorded round-0 byte (the legacy build refuses on an existing round 0).

## 4. Synthesis: where each reviewer changed the outcome

- **AdversarialRegate** found R17-01 (presented-path collisions), R17-05 (run commit) and R17-06 (legacy writes). It also raised the P1 "never easier" text, now part of R17-04, and the "met" wording, now part of R17-02. The lead had missed R17-01 and R17-05, and both were verified against `control-key.jsonl:1,20,23,24,31,33`, `regate-supply.json` and `PREREGISTRATION.md:250`. Its budget point was rejected (§3.9).
- **ConfoundRegate** found that the population is *entirely* pre-2023 (R17-03; the lead had "some", taken from the drafter), the P2 easing (R17-04), the clustered sizing figures (R17-02), the procedural record (R17-07) and the status-specific HD6 reason. It also flagged the reuse of round-0 target directories, which the lead made binding inside R17-01.
- **The lead** contributed the pre-Jev label-only check (R17-02): neither peer proposed moving the label-only and construction conjuncts before Jev. It also contributed the digest comparison (36 equal, 8 changed, 1 new) and the rulings in §3.
- **Competing explanations RESULTS must carry** (no fix needed beyond the fields added above):
  - for a PASS: wider gold (A-only recovery), an acceptable sibling (`sibling_in_acceptable`), old central files, cues in the PR text, exclusions removing ambiguous items, and labellers seeing more context than Jev;
  - for a FAIL: the untuned threshold (which conjunct failed, abstention reasons), cross-scope near-duplicate areas (R-054/HD7), legacy content, clustered P2 yield, strong topic cues in test-directory paths, and run-to-run noise near the thresholds.

## 5. Verdict

**7 ISSUES** — R17-01 to R17-07. Not signed off. The re-gate must not be built.

For round 18, the drafter supplies a response mapping each fix to `file:line`, the regenerated `reviews/regate-supply.{md,json}` after R17-01 and R17-03, the new G13 fixtures, a validation log with `signoff.json` moved aside, and the recomputed digests. A round-18 sign-off would then require:
- the digests recomputed equal and the tune pin `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c` equal;
- a new `signoff.json` with the round-18 `round`, whose `code_under_test` equals round 11's 17 rows;
- validation (a) with 0 failed;
- validation (b) with exactly the 1 expected failure;
- one commit;
- the worktree moved by checkout only.

The run order would be: `regate-build` alone → the R17-07 comparison and records → `regate-label` → `regate-jev` (pre-Jev check first) → `regate-analyse` → (PASS on both) the round-0 critique → rounds 1–10 → `verdict` → RESULTS naming the four run commits.
