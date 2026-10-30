# X2d — council experiment review, round 18 (rev 13 after the round-17 fixes: the one permitted re-gate of P1 and P2 on fresh items, R-053)

Date: 2026-10-30. Reviewer: innovation council. Opus 5.5 wrote the lead synthesis. Two peers worked read-only. Each received the inputs without the lead's leaning:
- `innovation-challenger` = AdversarialRound18 (GPT Astra);
- `experiment-peer` = ConfoundRound18 (Opus 5.5).

Governance: `CLAUDE.md` experiment rules 1–7; user decision R-053 (option (c) of `reviews/pilot-round-0-gate-stop-memo.md`).

Scope:
- `PREREGISTRATION.md` revision 13 as amended for round 17, and the changed implementation: `build_controls.py`, `analysis_x2d.py`, `run.sh`, `x2d.py`, `validation/regate_checks.py` and `README.md`. `gold.py` and `validation/validate.py` were re-read where the fixtures call them.
- The inputs: `reviews/round-17.md`, `reviews/round-17-response.md`, `reviews/regate-supply.{md,json}`, `/tmp/validate-r17.log` and `signoff.json` (round 16).
- The brief's questions: is each fix present as worded, and does it work on the code paths the fixtures exercise? Do the changed lines add defects? Do the fresh items avoid every exposed path and target directory? Can the pre-Jev check be bypassed, and does it read any Jev output? And the two departures.

Reviewed files (SHA-256): `DRAFT-NOTES.md` `fd5778c67ad481df16ccbf464f40b82ecac13223deed2b2f29b858382229e65d`; `GUIDE.md` `3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980`; `PREREGISTRATION.md` `b9110486014bca580be3f14629c0beda54f35109a2c89ca5ca04ad95085140a8`; `README.md` `bb7a01425e48071f3dba34a831c5de644c04c7cab65c28fcd2aed181d6292ce3`; `analysis_x2d.py` `305848bff23a184c06f3d9ace8323014a7c6fc0155500858db318da38addaf19`; `b2.py` `5bb9c12d8cec5ea5cf02959f2ee51f9275d65c746ce51e314552623cd28db14f`; `baselines.py` `a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0`; `build_controls.py` `34c466c34fe77db559ad0044724cebef2ee78adc617fab4f5f69880f26b3a128`; `build_units.py` `c052246c828024a25fc6a143d513e111ecbf63b42b8b5912dd37e960a1c3e243`; `collect_onboard.py` `c93fa59c8fe305459336e08e52c60f365d8df0a70d7b6ca395c58e8f88da5fff`; `compose.py` `296001923384652d7d31cdb33ca937ed409783023de2b6ee53c62e4ea8ab5bd1`; `frontend.py` `dd4a0fdc1e3595eeefb82a8eafab600255b3b0fe95f67e2fb8a07eb5972ea0d8`; `gold.py` `a7831d65d40c0fa9ee15555a8488532dc646990725ac47a670f1fb6fec6bb8e5`; `jev/main.go` `63fbec0534a2ddf3d55d07c074c6a882e199ec2b90cf771a984c9fa9ee586ca0`; `label_x2d.py` `fa5d60ef2bbf2147102cf986048f7711cd2ad90d8cd7a0ddbe89aa5162729dcb`; `measurement/lever-result.json` `298902e717d0db8826c35d219695d386b19644f99cfc205d6b2d891100b0352c`; `measurement/lever_check.py` `8f8008d3a28d72d5e768bc292e4917449fd8cb31ca0a642ff6057ab51b705a42`; `onboard/main.go` `7e2ed1edc64b694851e7729e17af4f2ba63dafa7c9cdcf19cd3a5a255c08d46b`; `power_sim.py` `a955f7177280bfd6d20a2514d21a04819c426610d594ee4e7dbfe0cdfa9c5279`; `prepass/commits.jsonl` `1581986e0efbf26bb51aea894c3ffcdd23f54ef6e1d6d6ab0e4e99248f35bd1c`; `prepass/k0.json` `f89e8d0b264eb33eb8bcb5aa4a4910bd462c3e1befea480b721ee514947a98ab`; `prepass/reserve.json` `a73558ed54efdd767c5c0cf9a5accdcf41f511acd044d630208a8fb3e129f12a`; `prepass/rounds.json` `0aad85b0199e747630d0e863b7a5e83b2624684c50c1150ee4e8775fde1d7516`; `prepass/scopes.json` `a24bfc8d19aab9fd5ab0817f317cee560df65c63b1d63b5dbd10b5529d16a512`; `prepass/summary.json` `e1cfced0da45bcd86cc516eef39fa70c10b3fa5d630820d415cdea47045cc1c1`; `prepass.py` `1e5863d32974f1d7fecfd4df5c30b6e02aa137b9db491780eab11cc00b500696`; `quality.py` `6d65293d64c76611368fa8eaf93d4f48432cf203bb65aed9ed63130b8527cade`; `questions/merge.area_covers@v1.json` `cce59e612b67d76b96d927613c1a780ea64915c38210b7be8fceb5cfaa7e5af0`; `questions/merge.covered@v1.json` `f37eea94bb6e2c720cbbb904066dd06c368ecc7581c90063b9f17e5a751058b9`; `questions/merge.which_area@v1.json` `cbb7ba7e92432a1e0ffc9bf7d8718a9001ba4227137a6ab60a1e359393eff928`; `questions/onboard.covered@v1.json` `ec4d62be051c016ab945738c155401a17f091654e1efff9cfd60ec18314eaa70`; `questions/onboard.same_responsibility@v1.json` `31f3b25d3bbf535b026bd6f0a022d39f01e7c519cc22649195d341b75a49a7b1`; `questions/onboard.which_file_area@v1.json` `425edf93d08dba42d7f1da755601d5dec27cded8b67593401b7656b24cf12f57`; `reference_arm.py` `d139f0179c06ca5c6aff5784ceab2feb4113f888c51cd46f760f7a3d943bb739`; `run.sh` `46bcf50055b5492a81dc526355eb7f678f55262d1d624c5fdd35f2cd5809260e`; `scopes.py` `5a43296396365fabeaab96879673597864e3dc7c83faa225558ed932dadbe072`; `signoff.json` `578a2462a6f25953c2e744a1d3dcc546f58af868f2e0e343e204367a89552e61`; `thresholds.json` `7142d7ac15abbcd27b75ac8005a7f1bffed0c6a70395c65e8f631bfcd32995eb`; `tune.py` `2628c5fa59bb3135f5618495ff193ebf7c3b03116724b1f3f06e44e87c6c134c`; `validation/r4_checks.py` `f7f14077d104558bf51cf66e2b5a6ea1fb34347d4dbd7496d6e9bb680aa39d52`; `validation/regate_checks.py` `07c575370bb9f17b27396e55752c7d5e49372324b674f7f0c083cbc546ce5943`; `validation/stub_clawpatch_x2d.js` `d701211edbc90fba134f97a00106e1a47dc6bc39a7b49e5a34f14a1c73d4f91f`; `validation/stub_jev.py` `f2ec432f778468e28d5ea6eaab30232a9f5491b7d1fd52232149c46dabaa0f9d`; `validation/stub_labeller.py` `d38631fdd39c582feb6b9feda8c28d4abeec4043b865ccab617e57a22fc595e5`; `validation/validate.py` `8fab5ee017b8eecd8b0dc13dacb45edf9d7383e8586334ec7f994bf764ee16c9`; `versioning.py` `331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168`; `x2d.py` `cf55b50d44fddc87eb36c889b9bba72e750746c5ef118b5247ffe7b8d47ab11d`

Rule 3 check: no predictor–outcome association has been computed. The council read the following, all read-only.
- In `../scout-x2d/…/data/rounds/0/`:
  - `control-key.jsonl`: rows matched on `target_dir`. The grep printed whole rows. Their other fields are construction facts only: `area_id`, the construction gold `{A}`, ablated areas and the decoy area. No label and no Jev answer appear in them.
  - `units.jsonl`: only `case_id`, `control_kind` and `members` of the `control: true` rows, read through a jq projection.
- `reviews/regate-supply.json`, which holds membership only.
- The PASS lines and the summary of `/tmp/validate-r17.log`.

No Jev score, arm or answer file was opened, and no Jev answer was compared with any label or gold.

## 1. Verification

Status: **verified** = as claimed; **issue** = see §2; **condition** = left to the sign-off conditions (§5).

| Item | Status | Evidence | Issue |
|---|---|---|---|
| R17-01 (1)–(2) exclusion code | verified | `exposed()` adds `presented` (the members of every round-0 `control: true` row) and `target_dirs` (the `target_dir` of the `kind == "p1"` key rows) (`build_controls.py:461-466`). It reads round 0 of `x2d.main_data(data)`, never the re-gate root (`:449`). `_candidate_dirs` skips `forbid_dirs` before `taken`, for every stratum (`:229-232`). `_p1_target` skips a directory when any new path is in `forbid_paths` (`:259-262`). `_p1_placements` seeds `taken = set(forbid_dirs)` (`:564`). `regate_plan` passes both sets (`:589-590`, `:598`, `:603`), and so does `make_regate` (`:673-674`) | — |
| R17-01 (3) counts | verified; wording departure accepted | `supply["exposed"]` carries `presented_paths` 83, `rev12_p1_target_dirs` 17, `p1_rejected_for_presented_path` 0 and `p1_directory_candidates_rejected_as_rev12_target` 662 (`:636`; `regate-supply.json` `exposed`). The 662 counts directory visits summed over the 40 placements of the final trial, not distinct directories, and `regate-supply.md` says so. The counts were not copied into `funnel`, and the response says this openly. That changes no selection and hides no count | — |
| R17-01 (4) text | verified | `PREREGISTRATION.md:111`: the Excluded list carries the council's clause word for word, after "the PRs named (`#<digits>`) in the guide;" | — |
| R17-01 (5) fixtures | verified | Toy world (`regate_checks.py:304-310`): the only stratum-1 host `tests/bdir` is chosen without the forbidden sets and never with either set, and each rejection is counted once (log `:427`). Real clone (`:595-614`): no planned item sits at a presented path or in a rev-12 target directory. In a data copy, a rev-12 P1 control planted at a stratum-1 placement (`bbd77ef2` → `docs_src/response_cookies` on the stub catalogue) displaces that placement, and the supply records +1 directory and the rejections (log `:479`). Reverting `forbid_dirs` would let the deterministic plan reuse the planted directory, so the check would fail [INFERENCE: static] | — |
| R17-01 (6) freshness of the regenerated supply | verified, recomputed independently by the lead and AdversarialRound18 | 40 of 40 planned P1 presented paths (`target_dir/basename`) avoid all 83 presented paths. 0 of 40 P1 target directories is among the 17 rev-12 target directories (`control-key.jsonl:1,2,3,6,8,12,13,20,21,23,24,25,26,27,31,33,34,36,39,46`). The 37 stratum-1 directories are distinct. The 41 P2 member paths avoid the presented paths and the 51 rev-12 P1 original paths, and so do the 40 P1 source paths. No member path, PR or cluster id is repeated across the 80 sources. Strata 37/0/3; P2 has 40 clones in 16 areas (4,4,4,3,3,3, 2×9, 1). By construction `forbid_paths` can bite only on presented paths outside the C0 tree, because `_p1_target` already refuses C0 paths (`:263`) and the rev-12 P1 paths lie in forbidden directories; hence 0 rejections. Three stratum-1 targets are rev-12 *source* directories, not presented paths or targets (§3.6) | — |
| R17-02 (1) precheck function | verified | `_regate_denominators` (`analysis_x2d.py:1427-1481`) opens `units.jsonl`, `control-key.jsonl`, `gold.jsonl`, `arms/B1.jsonl` and `arms/B2.jsonl`. It checks only the presence of `labels.jsonl`. It never opens `arms-p1/`, `arms-p2/` or `jev-p/`. `regate_label_problems` (`:1484-1506`) applies `< 20` eligible stratum-1 items, `< 8` positives after twins, rules-resolved `> 0`, and B1/B2 `> 0.20` or `None`. It refuses outside the re-gate root (`:1489-1490`). The CLI prints the counts and exits 1 on any problem (`:1749`, `:1781-1786`). `regate_gates` uses the same function (`:1524`), so the precheck's denominators equal the gate's: `gold.jsonl` is create-once in the re-gate (`run.sh:364`) | — |
| R17-02 (2) wiring; bypass | verified | `run.sh:369-371`: gold presence check → `--regate-precheck` → only then `runs_of > .runs` and the `jev_run` loop. The refusal message is word for word. If `regate-analyse` runs after a refused `regate-jev`, it cannot write `gates.json`: `regate_gates` first `_need`s `arms-p1/V-sem.jsonl`, `jev-p/scores.p1.jsonl`, `arms-p2/V-sem.jsonl` and `jev-p/scores.p2.jsonl` (`:1525-1527`), and the CLI writes only after a successful return (`:1791-1793`). No environment variable skips the step. The only way round it is to run the Go driver by hand outside `run.sh`, which is a protocol breach, not a code path | — |
| R17-02 (3)–(6) text | verified | `PREREGISTRATION.md:110`: stop-rule text word for word (pre-Jev conjuncts, "not a gate result", HD6 `preregistered` meanwhile, "A re-gate that passes this check and is then not PASS for BOTH P1 and P2 stops the pilot"). `:116` mirrors it. The UNMEASURED clauses at `:112-113` read "a defensive status, the pre-Jev check makes it unreachable". The HD6 map `:253` carries the shortfall clause and the status-specific reason. The clustered sizing sentence is in `:113` and in the `x2d.py:52` comment | — |
| R17-02 (7) supply wording | verified in substance | `regate-supply.md:47` carries the replacement ("construction counts are met … yield is unmeasured until labels exist"). The Result line `:55` still reads "Both minimums are met", but its parenthesis now names only construction counts (40 P1, 37 stratum 1, 40 P2 clones), so the sentence is true. The response's account of that line is inexact. Not raised: the file is not frozen, the build comparison uses the JSON, and the positive yield comes from the precheck record (§3.4) | — |
| R17-02 (8) | verified | `positives_by_ablated_area` (`analysis_x2d.py:1575`); asserted in the P2 planted-effect check | — |
| R17-02 (9) fixtures | verified | Unit fixtures (`regate_checks.py:449-482`, log `:400-402`): with `arms-p1/`, `arms-p2/` and `jev-p/` deleted, 20 / 8 / 8-of-40 / 0 gives no problem; 19 eligible, 7 positives, B1 at 9 of 40 and one rules-resolved unit each give exactly one problem; outside the root it refuses. End to end (`:761-779`, log `:490`): `run.sh regate-jev` refuses with "no Jev request was made" and "direct-label P2 positives 0 < 8", writes no score file under `jev-p/` and no `gates.json`, and leaves the whole-directory manifest unchanged. The same labelled state with planted truth passes the precheck through to Jev (log `:492-493`) and to the planted null (log `:497`) | departure 2: §3.2 |
| R17-03 | verified | `PREREGISTRATION.md:111`: the legacy-population sentence and the corrected replay wording. The supply items carry `date` (`build_controls.py:640,643`). `gates.json` `p1_items` and `p2_items` carry `date` (from `units.jsonl`; `make_unit` writes it, `build_units.py:195`) and `origin` (from the key) (`analysis_x2d.py:1455-1457,1475-1477`). The fixture `:484` checks that the keys exist, not their values. The build check (log `:485`) asserts non-empty dates in `supply.json`. The figures reproduce: P1 2019-03-03 to 2022-09-01 (22/9/5/4 by year); 36 P2 clones `in_no_unit` with PR numbers below #10000; 4 clones from 2023–2024 (D 2, G 2) | — |
| R17-04 | verified | `PREREGISTRATION.md:112` (no dominance claim; A-only recovery and `sibling_in_acceptable` reported), `:113` (easing paragraph), `:15` (change log). Code: `_p1_stats` `:1416-1417`; `none_rate_all_positives` `:1576`. Fixtures: `p1-bacc` asserts 0.0 and 24 (`:134-138`); the P2 bounds assert 0.85 and 0.80 (`:200-209`) | — |
| R17-05 | verified | `PREREGISTRATION.md:250` carries the "[superseded in rev 13 for every stage after `power`: item 20]" marker. `:251` carries "**Fourth run commit.**" word for word, with `730674e`, `0b56594`, `b6a5bf3` and the rev-13 commit (`b6a5bf3` matches `pilot-round-0.md:52`). `README.md:15` matches | — |
| R17-06 | verified | `gates_closed` (`run.sh:293-295`, `-s` test, message word for word) is called right after `verify` in `stage_gates_build` (`:298`), `stage_gates_label` (`:308`) and `stage_gates_jev` (`:324`), before any writer. The aggregate `gates` stage starts with `stage_gates_build` and so refuses too. Fixture `:780-785` (log `:491`): all three refuse with "never written again", and the `rounds/0` manifest (excluding `regate/`) is unchanged | — |
| R17-07 | verified (procedural, as specified) | `PREREGISTRATION.md:110` carries the record procedure word for word. The "runs alone first" sentence appears twice, which is harmless | departure 1: §3.1 |
| New defects in changed lines | none at the bar | Below the bar: the `build_controls.py:36` docstring keeps the old "first-parent commit that added their files" wording (the frozen preregistration text is the corrected one); `PREREGISTRATION.md:110` says the precheck "reads only" four kinds of file, omitting the presence check on `labels.jsonl`; `:196` "each once" does not mention a resumed `regate-jev`; the check name at `regate_checks.py:775` says "in the labels" where the plant is in `gold.jsonl` (§3.2) | — |
| Stale figures | none | No "34/2/4" in `PREREGISTRATION.md`, `README.md` or `x2d.py`; it appears only as history in `regate-supply.md:3`. "24 / 8 / 8 of 40" (`:112`, `x2d.py:47`) describes the target pattern and is still correct. No live mention of the removed combined stage (`run.sh:403` dispatches the four stages only) | — |
| Digest table | verified as text | 45 rows (`round-17-response.md` §5). 36 rows equal `signoff.json` (round 16) row by row, including every Jev-side file and the state builders. `gold.py` (`a7831d65…e6b5`) and `validate.py` (`8fab5ee0…16c9`) equal the rev-13 draft table. 6 rows changed since round 17 (PREREG, README, `analysis_x2d.py`, `build_controls.py`, `run.sh`, `x2d.py`) plus `regate_checks.py`. `signoff.json` lists 44 files; the new record must list 45. Byte identity cannot be proven read-only | condition |
| Validation (a) | verified from the log | `/tmp/validate-r17.log:779-781`: "715 checks, 0 failed", "all validation checks passed", record SHA-256 `578a2462…2e61` before = after. No line starts with `FAIL` | condition |
| Status line and ledger | verified | `PREREGISTRATION.md:3` reads "DRAFT for council round 18" (as in the precedent of rounds 15–16). `docs/hypothesis-ledger.md:39` ends "PREREGISTRATION revision 13 and council review pending"; the main agent updates it at sign-off (§5) | — |

## 2. Issues

None. Every round-17 fix is present and works on the code paths its fixtures exercise. The regenerated supply avoids every presented path and every rev-12 target directory. The pre-Jev check cannot be bypassed through `run.sh` and reads no Jev output. The three candidate issues raised by ConfoundRound18 do not reach the bar (§3.3–§3.5).

## 3. Rulings

1. **Departure 1, the combined `run.sh regate` stage removed: accepted.** A combined stage would have run label, Jev and analyse straight after the build, skipping the R17-07 comparison and record boundary. The dispatch (`run.sh:403`), the header (`:8-9`), `README.md:36,67` and `PREREGISTRATION.md:110` agree. Each prerequisite is still enforced in code (`regate_ready`, `regate_open`, the gold check, the precheck, `_need` on the score files), and the end-to-end check runs each stage separately.
2. **Departure 2, low yield planted by rewriting gold rows in a labelled copy: accepted.**
   - Both the precheck and `regate_gates` read only `gold.jsonl` (`covered` and `acceptable`). Nothing compares gold with labels, so a gold–label mismatch in the copy can neither hide a refusal nor cause one.
   - The rewritten rows have the shape `gold.py` gives a covered direct label (`covered: true`, `acceptable [r2]`, `primary r2`, `none_kind null`, `derived_gold` kept).
   - The labels → gold step is tested separately: at pipeline level (log `:488-489`), in `gold_checks` (`regate_checks.py:231-271`), and by `gold.py` refusing incomplete gold.
   - This record states that the plant is in `gold.jsonl`. The check text ("in the labels", `:775`) and `round-17-response.md` R17-02 (9) ("own labels rewritten") describe it loosely. Neither is a gate figure.
3. **Candidate R18-01 (the `regate-supply.md` Result line and the stratum table): below the bar.**
   - On the Result line, see §1 R17-02 (7).
   - The "can host" table (`:35-41`) is headed "before the removal sets and the leak check". It counts directory kinds on the full catalogue. The achieved strata come from `BU.stratify` after the trial's removal set. So "stratum 3: 0" beside 3 achieved items is consistent, not contradictory.
   - That all 3 stratum-3 items share `tests/test_tutorial/test_response_change_status_code` is a report-only fact (§3.6).
4. **Candidate R18-03 (no digest for the R17-07 comparison reference): below the bar; procedure added.** `reviews/regate-supply.json` is fixed by the rev-13 sign-off commit, so the comparison is auditable through git. Run step 1 nevertheless records its SHA-256 at that commit (§5). If it differs from the committed bytes, that is a mismatch stop.
5. **Candidate R18-02 (the HD6 map and stops before `regate/gates.json` exists): below the bar; reading recorded.**
   - These stops route to a decision by the council and the user and do not end the pilot:
     - the build refusal (`:111`);
     - the build-record mismatch (`:110`, "a stop for a new council round");
     - the frozen-copy refusal (`regate_frozen_problems`);
     - the gold refusal;
     - a `regate-jev` that does not complete (the driver exits non-zero, `set -e`, and `regate-analyse` then refuses on missing files);
     - the pre-Jev shortfall.
   - They are pauses like the "cap breach" (§8, §12), not "an integrity stop before `verdict`" in the sense of `:253`. HD6 therefore stays `preregistered` until that decision. This matches the item-19 precedent (the rev-12 cost stop left HD6 `preregistered`).
   - Only a written `regate/gates.json` that is not PASS for both P1 and P2 maps to `inconclusive`, with the reason at `:253`. This is a reading of the frozen text, not a change to it.
6. **Report-only facts RESULTS must carry beside the re-gate figures (no fix; every needed field is in `gates.json`):**
   - **Area concentration.** Area `e287e9` is A for 5 stratum-1 items and `9f08e6` for 3. Wilson intervals treat items as independent, so give per-area accuracy and a one-item-per-area sensitivity check. `e287e9`, `9f08e6`, `3de70a` and `617fde` are also P2 ablated areas, but on the separate `catalogue-p2`.
   - **Rev-12 source directories as targets.** Three stratum-1 targets are rev-12 *source* directories: `tests/test_tutorial/test_separate_openapi_schemas` (ctl-1), `scripts/playwright/separate_openapi_schemas` (ctl-6, ctl-25) and `tests/test_validate_response_recursive` (ctl-24/ctl-9). Their sibling cue points at a rev-12 item's area. The basenames are new and C0 paths are refused, so this is not a path exposure. RESULTS lists them.
   - **Topic-named directories.** 24 of the 37 stratum-1 targets are `tests/test_tutorial/test_<topic>` and 9 are `docs_src/<topic>`. Report path-following by `target_dir` prefix.
   - **Stratum 3.** The stratum-3 figure rests on one directory and is not a stratum-level result.
   - **Carried over.** The round-17 §4 competing explanations still apply.
7. **Items already ruled in round 17** (bars, P2 binding, strata 2–3 report-only, the abstention knife edge, legacy population validity, budget overshoot by requests in flight, D-origin sources as permitted by memo §1): not re-raised.

## 4. Synthesis: where each reviewer changed the outcome

- **AdversarialRound18** found no issue at the bar. It independently recomputed the freshness figures (83 / 17 / 0 intersections) and extended the check to the 51 rev-12 P1 original paths and to path, PR and cluster uniqueness across the 80 sources. It traced that `regate-analyse` after a refused `regate-jev` cannot write `gates.json`, and that a hand-run Go driver is the only route round the precheck (a protocol breach). It accepted both departures. This confirmed the lead's sign-off leaning and widened the freshness evidence.
- **ConfoundRound18** raised three minor candidates (R18-01 to R18-03). The lead verified each against the files and ruled them below the bar (§3.3–§3.5), because none changes a gate, the verdict or a frozen statement RESULTS would cite. Its points nevertheless changed the record in three ways:
  - the SHA-256 of `reviews/regate-supply.json` is recorded at run step 1;
  - the HD6 reading for pauses before `gates.json` is ruled explicitly;
  - the RESULTS facts in §3.6 are its competing explanations (area concentration, rev-12 source directories, topic cues, one stratum-3 directory).

  Its condition on departure 2 (state that the plant is in `gold.jsonl`) is met by this record.
- **The lead** contributed four points:
  - the equivalence argument: precheck and gate share `_regate_denominators`, and gold is create-once, so "UNMEASURED is unreachable" holds in code and not only in procedure;
  - the observation that `forbid_paths` can bite only on presented paths outside the C0 tree, which explains the 0 rejections;
  - the row-by-row digest comparison against `signoff.json`;
  - the run order below.

## 5. Verdict

**ZERO ISSUES — signed off.**

### Sign-off statement

The innovation council signs off `PREREGISTRATION.md` revision 13 and its implementation for the one permitted re-gate (R-053) and the rest of the pilot. Reviewed preregistration and script digests (SHA-256, from `reviews/round-17-response.md` §5, 45 files):

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

Tune pin (`tune_script_sha256`, `~/.omp/agent/skills/jev-actions/scripts/tune_threshold.py`): `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c`.

**The sign-off holds only if all of the following are true before run step 1.** If any one fails, the sign-off lapses and a new council round is needed.
1. The 45 digests above are recomputed from the committed bytes and each one is equal. The tune pin is equal.
2. A new `signoff.json` has `round` `reviews/round-18.md`. Its `files` equal the 45-row table above, its `tune_script_sha256` equals the pin, and its `code_under_test` equals round 11's 17 rows exactly (same keys, same digests).
3. Validation (a), with `signoff.json` moved aside, gives 715 checks and 0 failed. Record the log path, and the record's SHA-256 before and after the move.
4. Validation (b), with the new record in place, shows exactly 1 failure, the expected one: `C without the council's sign-off record the collector refuses before any clone or model call`. Record the log path.
5. Everything goes in one commit: the rev-13 files, `signoff.json`, `reviews/regate-supply.{md,json}`, `reviews/round-17.md`, `reviews/round-17-response.md` and this round. Record its full hash: it is the **rev-13 sign-off commit**, the fourth run commit.
6. The run worktree `../scout-x2d` moves to that commit by checkout only: no `git clean`, no `stash -u`; `data/` untracked and untouched. Record `git rev-parse HEAD` of the worktree.

### Run order (authorised once the conditions hold)

Stages run from the run worktree. Jev stages run under `infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- …`. Labelling runs through `claude -p` and `codex exec` only. Every record below is appended to `reviews/pilot-round-0.md` (never inside its existing gate record) and committed.

1. **`run.sh regate-build` alone, then the plan comparison (R17-07).**
   - Compare the P1 and P2 item lists of `data/rounds/0/regate/supply.json` with `reviews/regate-supply.json` as committed at the rev-13 sign-off commit: cluster ids, members, target directories, strata and ablated areas.
   - Record:
     - the comparison;
     - the SHA-256 of `reviews/regate-supply.json` (§3.4);
     - the SHA-256 of `regate/regate.json` (with its `utc`), `regate/supply.json` and `regate/rounds/0/control-key.jsonl`.
   - Any mismatch is a stop for a new council round, never a re-build. HD6 stays `preregistered` (§3.5).
2. **`run.sh regate-label`, then the pre-Jev check.**
   - The check may be run alone first. It is read-only and prints label-side counts only: `X2D_DATA=<data>/rounds/0/regate .tools/venv/bin/python evaluation/x2d-merge-association-pilot/analysis_x2d.py --regate-precheck`. Record its `regate precheck:` line and any `PROBLEM:` lines.
   - On any problem: stop, no Jev request; the drafter reports, the council and the user decide; HD6 stays `preregistered`.
3. **`run.sh regate-jev`, then `run.sh regate-analyse`.**
   - `regate-jev` repeats the check before any request; record its precheck line from the log.
   - Record the SHA-256 of `regate/gates.json`, the P1 and P2 statuses, and the re-gate Jev spend.
   - If `regate-jev` does not complete, that is a pause (§3.5), not a gate result.
4. **On PASS for both P1 and P2:** write the round-0 critique (`reviews/pilot-round-0.md`, with every line `x2d.critique_problems(0)` requires). **Otherwise (FAIL or UNMEASURED):** stop. Write `RESULTS.md` (step 7). HD6 → `inconclusive` with the reason "round-0 gates P1 and P2 failed and the one permitted re-gate did not pass (re-gate P1 <status>, P2 <status>); HD6 never under test".
5. **Rounds 1–10.**
   - Each round needs the previous critique; copy each critique into the run worktree's `reviews/` before the stage that reads it, and record the copy's SHA-256.
   - `rescore-v7` runs after the round-7 critique.
   - The worktree is not checked out again before `verdict`.
6. **`run.sh verdict`.**
7. **`RESULTS.md`.**
   - It reports every preregistered re-gate figure with its interval: the decisive figures, `recovery_all_items_A_only`, `sibling_in_acceptable`, `none_rate_all_positives`, `positives_by_ablated_area`, `derived_vs_direct`, per-stratum figures and path-following, by year and origin. It also carries the legacy-population statement and the §3.6 facts.
   - `Deviations` names the four run commits (`730674e`, `0b56594`, `b6a5bf3`, the rev-13 sign-off commit) and the stages run at each.

Ledgers (main agent, at the sign-off commit): add a `docs/hypothesis-ledger.md` HD6 History line saying that PREREGISTRATION revision 13 was signed off in council round 18 (rounds 17–18: 7 → 0 issues) and that HD6 stays `preregistered`. The §14 map applies from here, with the §3.5 reading. Cite R-053.
