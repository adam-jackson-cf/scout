# Decision memo: X2d round 0, gates P1 and P2 failed

Paths: `data/` means `../scout-x2d/evaluation/x2d-merge-association-pilot/data/`. Other paths are relative to `evaluation/x2d-merge-association-pilot/` unless they start with `docs/`. Short area ids are the first 6 characters of the ids in `data/catalogue/catalogue.json`.

## 0. What happened and what it means

Round 0 ran to the end on the signed-off rev-12 bytes (sign-off commit `b6a5bf3`). Most checks passed: K2–K6, N1, the planted new concern, and every power cell. Two gates failed (`data/rounds/0/gates.json`; `reviews/pilot-round-0.md`). The pilot therefore stopped before round 1, as `PREREGISTRATION.md:110` requires. Round-0 Jev spend was about $1.10 of the $2 round cap.

**P1, the "move a file" test.** We moved 20 files into folders that belong to other areas and asked Jev which area each file belongs to. To pass, Jev had to name the original area on at least 12. It named it on 9 (0.45; Wilson 95 % [0.26, 0.66]). The directory baselines scored 0.05, as intended.

Jev never confidently named the folder's area. Its 11 misses break down as follows:
- **5 threshold misses.** Jev's top pick was the right area, but its confidence fell below the untuned 0.70 cut-off.
- **2 contested misses.** Jev confidently named a different area, and our own blind labellers accept that area too.
- **1 near-duplicate miss.** Jev could not settle between two almost identical areas.
- **3 real path-pull misses.** Jev leaned towards the folder's area but was not confident enough to commit.

The real weakness is in the hardest stratum (one other area owns the folder), where Jev recovered the original area on only 2 of 11.

**P2, the "remove an area" test.** We removed 10 areas from the catalogue and asked about 12 files those areas owned. After the preregistered derivation from the source labels, 7 files should have no fitting area, and Jev should answer `none` on them. It did so on 4. Three misses means FAIL.

Two of the three misses are the same file counted twice (a path twin). For all three, our blind labellers looked at the same thinned catalogue and said an area does still fit, and in each case it is the area Jev picked. The gate did not use those direct labels. It derived its gold mechanically from the source's labels on the full catalogue and discarded the clones' own labels (`gold.py:69-78`).

**What it means.** P2's failure is mostly a construction defect. P1's failure is mixed: about half comes from construction and scoring, and part is a real weakness on path conflicts. Even so, Jev hesitated rather than committed: no assignment went to the folder's area.

**Neither result can be rescued by re-scoring.** Two sensitivity checks show this:
- Scored against the labellers' broader answers, P1 is still 0.55, which fails.
- The untuned 0.70 threshold that hurt P1 is what produced 3 of P2's 4 correct `none` answers. No single threshold change rescues both gates.

HD6 has never been under test, and the window outcomes for rounds 1–10 are untouched.

## 1. Recommendation

**Option (c): repair both gates and re-run them once, on fresh held-out items, before round 1. Keep the Jev side (model, questions, guide, thresholds) frozen. Keep the round-0 FAIL on record as the finding.**

- **Report the stop in full now.** This covers gate values with intervals, the per-item P1 disclosure, the R10-02 P2 split and `Deviations`. The rev-12 FAILs are never re-scored.
- **Make PREREGISTRATION rev 13 change only what is shown to be mis-built.** Each repair is justified by a fact that does not involve Jev's answers:
  1. **P2 gold.** A clone's gold becomes its own blind label on `catalogue-p2`, made the same way as every other unit's gold. The derived gold is reported beside it. Any `none` counts the same way, whether gap or new. Path twins are excluded. The build is sized to give at least 8 direct-label positives with margin.
  2. **P1 gold.** P1 gold becomes the labellers' agreed acceptable set, which must contain A. The target choice no longer lets the baselines fall back to A's own parent directory. At least half of the items are in stratum 1, and stratum 1 decides the gate. A path-following rate is added (answers that name sibling area B when B is not acceptable).
  3. **Scoring.** P1 and P2 are brought into line with the R-051 rule: accuracy on answered items, plus a separate abstention bound. The council sets the bars in rev 13, from the bars that already exist (0.60, 0.20, 0.80, and the GO cap g), and never from these data.
- **Fresh means fresh.** The sources must not include the 20 P1 items, the 12 P2 clones or their path twins. They are chosen by a membership-only rule that is fixed before any label on them.
- **One shot.** If the repaired gates fail on the fresh items, the pilot stops. HD6 becomes `inconclusive` with the reason. There is no third attempt.

Why (c):
- It follows the study's own repair pattern: K4 says "fix … and re-run on a fresh draw from G before any round" (`PREREGISTRATION.md:104`), and the rev-10 precedent did the same.
- It respects R-049: no dismissal over two gate failures that are largely mis-built.
- It keeps the gates binding.
- It costs about $0.1 in Jev calls [INFERENCE].

Both peers reached the same core position (§6).

## 2. Why each gate failed

### P1: per item (the preregistered descriptive disclosure)

How to read the columns:
- **A** is the gold area, the file's owner before the move (`control-key.jsonl` line N = ctl-N).
- **B** is the area the directory baselines assign (`data/rounds/0/arms/B2.jsonl`; B1 is the same on every item). Stratum 3 means no owned sibling.
- **Answer** is V-sem's composed output (`arms-p1/V-sem.jsonl:1-20`), and **raw** is its unthresholded permutation-1 choice.
- **Labelled set** is the agreed acceptable set in `gold.jsonl`.

| Item | Stratum | File moved | A | Answer (conf.) | Raw | B | Labelled set | Category |
|---|---|---|---|---|---|---|---|---|
| ctl-1 | 3 | separate-schemas tests | ad8836 | **A** 1.00 | A | (none) | A, e287e9 | hit |
| ctl-2 | 1 | `tests/test_forms_single_model.py` | 3de70a Parameter declarations | 158cb3 Request form model tutorials 0.91 | 158cb3 | 3e0c34 Advanced middleware | **158cb3**, A, 7d2013 | gold narrower than labels |
| ctl-3 | 3 | dependency 008c tests | c3534a Dependency injection (tutorial) | unresolved 0.44 | A | (none) | 9a4a0e, A | threshold; probability split with the core area also titled "Dependency injection" |
| ctl-6 | 2 | playwright `image02.py` → `.github/workflows` | ef2b9b | `none` (p_true .67) | A | tie | A | covered threshold |
| ctl-8 | 1 | `tests/test_ws_dependencies.py` | 36295a | **A** .91 | A | ade7d5 | A, 9a4a0e | hit |
| ctl-12 | 2 | `fastapi/cli.py` → `fastapi/security` | 1c4aa7 | **A** 1.00 | A | 637a9b | A | hit |
| ctl-13 | 1 | form-model test | 158cb3 | **A** .86 | A | 749bd1 | A, 3de70a | hit |
| ctl-20 | 1 | `docs_src/request_form_models/…` → `docs_src/settings/app01` | 158cb3 | unresolved (the two option orders disagree) | 7d2013 Request forms | 9ce9c5 | A, 3de70a, **7d2013** | near-duplicate areas |
| ctl-21 | 2 | dependency 008d tests | c3534a | **A** .83 | A | e287e9 | 9a4a0e, A | hit |
| ctl-23 | 3 | `tests/test_additional_properties_bool.py` | e287e9 | **A** .91 | A | e287e9 (baselines fall back to `tests/`, which A owns) | A | hit; baseline leak |
| ctl-24 | 1 | recursive response tests → `test_header_params` | 617fde | unresolved (orders disagree) | **B** | ea1d02 Header parameters | A | path pull |
| ctl-25 | 1 | playwright `image01.py` → `notify-translations` | ef2b9b | `none` (p_true .66) | A | d1459f | A | covered threshold |
| ctl-26 | 1 | `docs_src/dependencies/tutorial008d*` → `docs_src/templates` | c3534a | unresolved 0.52 | A | 29a30e | 9a4a0e, A | threshold; duplicate "Dependency injection" split |
| ctl-27 | 2 | `scripts/mkdocs_hooks.py` → `.github` | 30399e | **A** 1.00 | A | tie | A | hit |
| ctl-31 | 2 | schema-example tests → `tests` | 95b28e | **A** .98 | A | e287e9 | A, e287e9 | hit |
| ctl-33 | 1 | playwright cookie-models `image01.py` → `.github/actions/people` | f0abaf | unresolved 0.49 | **B** | 342c6b People GitHub Action | A | path pull (real) |
| ctl-34 | 1 | `docs_src/sql_databases/tutorial002*` | 8a632c | unresolved 0.69 | A | 3d2f5f | A | threshold (0.01 short) |
| ctl-36 | 3 | header-models test | b58662 | **A** .96 | A | (none) | 3de70a, A | hit |
| ctl-39 | 1 | `docs_src/header_param_models/tutorial002*` → `bigger_applications/…/internal` | b58662 | unresolved 0.68 | **B** | 2874a6 Bigger applications | 3de70a, A | path pull (real) |
| ctl-46 | 1 | `tests/test_response_model_data_filter_no_inheritance.py` | 617fde Response model validation | 10c16a Response models .86 | 10c16a | 233f9b | **10c16a**, A | gold narrower than labels |

**Counts.**
- 9 hits.
- 0 composed answers equal B. Raw choice equals B on 3 of 20 (ctl-24, 33, 39; Wilson [0.05, 0.36]).
- Raw choice equals A on 14 of 20, and lies in the labelled set on 17 of 20.
- By stratum: stratum 1 recovered 2 of 11 ([0.05, 0.48]); strata 2 and 3 recovered 7 of 9 (`units.jsonl` `stratum`).
- The 20 items come from only 13 PRs, so the effective n is below 20.

**What is construction or measurement:**
- **Gold narrower than the labels.** Gold is the single owner A (`build_controls.py:316`; scored by `analysis_x2d.py:1355-1357`). The blind labellers' agreed sets are wider on 12 of 20 items and contain A on all 20. Jev's two confident "errors" (ctl-2, ctl-46) are in those sets.
- **Untuned threshold, abstention counted as a miss.** Five raw-A answers fell under the provisional 0.70 threshold (`data/rounds/0/thresholds.json` `provisional-0.7`; `compose.py:68-90`), and abstention counts as a miss. P1 dates from rev 1 and was never realigned with R-051, which judges accuracy on answered clusters and caps abstention separately (`PREREGISTRATION.md:134-136`).
- **Near-duplicate areas across scope shards.** There are two areas titled "Dependency injection" (c3534a and 9a4a0e), and "Request forms" sits next to "Request form model tutorials". This is the known scope-shard threat (`:211`); it also affects the main rounds.
- **Baseline leak on ctl-23.** This did not decide the gate.

**What looks like a real Jev weakness:**
- ctl-33 and ctl-39: the content plainly names A's topic, yet the raw choice is B.
- ctl-24: the same pull, but more diffuse.
- Jev's low confidence in stratum 1.

**Sensitivity (descriptive only, not a re-scoring):**
- Against the labelled sets, V-sem scores 11 of 20 = 0.55 ([0.34, 0.74]), which still fails. Baselines score 3 of 20 = 0.15, still within the 0.20 bar.
- Lowering both thresholds to 0.5 would give exactly 12 of 20 on P1 (the experiment-peer's calculation), but it would also send ctl-39 to B. It would also cancel P2's `none` answers on ctl-11, 22 and 44, whose p_true values are .68, .54 and .59 (`data/rounds/0/jev-p/scores.p2.jsonl:9,15,23`). The threshold pulls the two gates in opposite directions, so it must not be moved on the strength of these data.

### P2: the R10-02 split by source set

Sources are joined `control-key.jsonl` → `units.jsonl` `set` → `gold.jsonl` → `arms-p2/V-sem.jsonl`. Donor categories come from `reviews/round-10-response.md:39-52`. In this table, "derived gold" is what the gate used, and "direct label" is the clone's own blind label on `catalogue-p2` (`data/rounds/0/labels.jsonl`), which the gate discarded.

| Clone | Set | File | Ablated area | Derived gold | Direct label | V-sem | Gate |
|---|---|---|---|---|---|---|---|
| ctl-11 | P unplaceable | `pdm_build.py` | c4f3d5 Package build | gap | gap, agreed (:207) | `none` (p_true .68) | hit |
| ctl-18 | P unplaceable | `scripts/playwright/request_form_models/image01.py` | 26ac01 Playwright example: request forms | gap | **covered** f0abaf, 158cb3, adjudicated; labeller B said gap (:80-83) | unresolved .58, raw 158cb3 | miss |
| ctl-40 | P unplaceable | the same file (path twin) | 26ac01 | gap | **covered** f0abaf, 158cb3, adjudicated (:22-25) | assigned 158cb3 (.70) | miss |
| ctl-45 | P unplaceable | `CITATION.cff` | 73232e | gap | none (new), adjudicated (:138) | `none` (p_true .10) | hit |
| ctl-22 | D | `requirements-tests.txt` | c207aa Test suite and CI | gap | gap, agreed (:210) | `none` (p_true .54) | hit |
| ctl-35 | D | `tests/test_generic_parameterless_depends.py` | 9a4a0e Dependency injection (core) | gap | **covered** c3534a "Dependency injection" (tutorial), agreed (:286-288) | unresolved .68, raw c3534a | miss |
| ctl-44 | D | `tests/test_compat.py` | ea7526 Public API and compatibility | gap | none (new), agreed (:122) | `none` (p_true .59) | hit |

| Set | Clones | Gap positives | `none` | Misses | Retained covered (assigned correct) | No agreed gold |
|---|---|---|---|---|---|---|
| P (4 unplaceable, 2 leftover) | 6 | 4 | 2 | 2 (one twin pair) | 2: ctl-5, ctl-7 (0) | 0 |
| D | 6 | 3 | 2 | 1 | 3: ctl-10, ctl-16, ctl-30 (1, ctl-10) | 0 |
| Total | 12 | 7 | 4 | 3 | 5 (1) | 0 |

Ten areas were ablated, leaving `catalogue-p2` with 104 of 114. Two ablated areas are shared: 26ac01 (ctl-18 and ctl-40, a P/P path twin) and 73232e (ctl-7 and ctl-45).

**What is construction or measurement:**
- **Derived gold contradicts direct labels.** `derive_ablation` assumes the source's acceptable set (made on the full catalogue) lists every area that would fit once the best one is removed (`gold.py:92-104`). The study's own blind labels contradict this on all 3 misses, and in each case they accept the area Jev picked. This is a label-versus-label fact; Jev is not involved.
- **Twin double count.** The path twin counts one item twice. Round 11 kept the per-clone count as conservative (`reviews/round-11.md:54`).
- **A thin denominator by design.** Five of 12 clones keep a second acceptable area (`gold.jsonl:15,25,36,86,89`). The build checks ownership, not labels.

**Inherited D label errors: not observed.** The one D miss (ctl-35) had a correct source label on the full catalogue (`gold.jsonl:27`). Its problem is the derivation, not inheritance.

**Real signals:**
- On the retained-covered clones, Jev made one confident wrong assignment: ctl-5 → b58662 where the gold is 3de70a (`arms-p2/V-sem.jsonl:1`; `labels.jsonl:180`).
- Jev's `covered` reading is high (p_true 0.54–0.68) on 3 of the 4 true gaps. Its `none` answers depend on the untuned threshold.

**Sensitivity (descriptive only):**
- Collapsing the twins gives 6 positives and 2 misses: still FAIL.
- Using direct-label gold gives 4 positives and 0 misses: UNMEASURED, not PASS.

### Classification

| Gate | Construction or measurement | Real Jev weakness | Confidence |
|---|---|---|---|
| P2 | Derived gold contradicts direct labels on 3 of 3 misses; twin double count; thin denominator built in | Overconfident `covered` reading on gaps; one confident wrong answer on a retained item | moderate to high |
| P1 | Narrow single-A gold (2–3 items); untuned threshold with abstention scored as a miss (5); near-duplicate areas; baseline leak | Path pull in stratum 1 (ctl-24, 33, 39); low confidence on displaced files | moderate |

Two more confounds:
- The labellers see more than Jev: 300 lines and 20 files per area, against Jev's 60 lines and 5 sample files (`PREREGISTRATION.md:129`; `questions/merge.covered@v1.json`).
- Wording is unlikely to be the cause. The question already says "never from their directory" (`questions/merge.which_area@v1.json:7`).

## 3. Options (alternatives)

| Option | What happens | Rule implications | Cost | Main trade-off | HD6 |
|---|---|---|---|---|---|
| **(a) Stop** | Write RESULTS with the FAIL as the finding; end X2d | None beyond reporting (rule 6) | $0 Jev | Final and simple, but ends the pilot on gates shown to be partly mis-built; goes against R-049 | `inconclusive`, reason "round-0 gates P1 and P2 failed; HD6 never under test" (council to confirm the §14 map gap) |
| **(b) Guide or question fix; re-run round-0 Jev on the same items** | New question versions; re-ask the P1 and P2 items | rev 13 plus council. New question versions invalidate every round-0 Jev result (K5, N1, planted_new); a guide change also re-opens the labels | About $1.10 of Jev for a full round-0 re-run, more than the $0.90 left in the round cap, so a cap decision is needed; about 317 sessions if relabelled [INFERENCE] | Aims at the wrong cause (the evidence points to gold and thresholds, not wording). On the same items a pass is biased upwards: development, not evidence | Stays `preregistered`; a same-item pass cannot move it |
| **(c) Repair and re-gate on fresh items (recommended)** | §1 | rev 13 plus council with zero issues; planted effect and planted null for the rebuilt gates (rule 4); Jev side frozen, so K2–K6, N1 and planted_new stand | About $0.1 Jev; about 180 labelling sessions; 1–3 council rounds [INFERENCE] | Slower; enough fresh items must exist; may fail again (then stop) | Stays `preregistered`; `under-test` at round 1 if the gates pass; `inconclusive` if they fail |
| (c0) Re-score the existing items under the repaired constructs | Apply the new gold or scoring to these 20 + 12 items | Outcome-informed re-scoring (rule 3) | $0 | Inadmissible, and it does not pass anyway (P1 0.55; P2 UNMEASURED) | none |
| **(d) Gates descriptive; proceed with rounds 1–10** | Rewrite §5 so that P1 and P2 are reported only | rev 13 plus council | About $8.6 of Jev left of the $9.74 projection (R-052), plus labelling every round | **Not defensible.** It drops the only path-cue test right after it failed, while the stratum-1 weakness is real. A GO would rest on an instrument that was never validated | `under-test` at round 1; a GO → `supported` would be weak |
| (d′) Rounds 1–7 as refinement; binding fresh P1 and P2 at v7, before the frozen rounds 8–10 | Gate the tuned instrument that actually produces the GO evidence | rev 13 plus council; replaces the "stop before round 1" intent of `:110` | About $8.6, plus the fresh build | Tests the instrument at tuned thresholds, but spends about 70 % of the pilot before the check | `under-test` at round 1; `inconclusive` if the v7 gate fails |

**Is re-testing the same 20 P1 items valid?** No, not as gate evidence. The items, the cause of each failure and Jev's answers are now known to the drafter and to this council. Three items sit within 0.01–0.04 of a threshold (ctl-34, ctl-6, ctl-25), so run-to-run noise alone could flip the gate. A re-run on these items is only a development or regression check, and must be reported as such.

**Is (d) defensible?** Not as written. R-049 protects refinement from dismissal over one or two directional failures. It does not license turning a failed gate into a descriptive one after the fact. (d′) is the defensible form if the user puts refinement before gating.

## 4. Evidence

- **Gate results:** `data/rounds/0/gates.json` (P1 `{"V-sem":0.45,"B1":0.05,"B2":0.05}`; P2 `gap_positives 7`, `none_correct_gap 4`, `retained_covered {n:5, assigned_correct:1}`). Log and spend in `reviews/pilot-round-0.md`.
- **Rules:**
  - The P1 and P2 definitions, the R10-02 disclosure and the FAIL rule for fewer than 8 positives: `PREREGISTRATION.md:107-108`.
  - The stop rule: `:110`.
  - K4's fresh-draw repair: `:104`.
  - The costlier error is a silent wrong-area assignment: `:134`.
  - Integrity pauses are not evidence against HD6: `:135`.
  - The HD6 map: `:244`. It covers GO, NO-GO, INCONCLUSIVE and "integrity stop before verdict → inconclusive"; it has no clause for a round-0 gate FAIL.
  - Frozen files: `:245`.
- **Code:**
  - P1 gold is `[a]`: `build_controls.py:316-317`.
  - P1 scoring uses the key gold: `analysis_x2d.py:1349-1360`.
  - P2 gate: `:1372-1394`.
  - Clone labels replaced by derived gold: `gold.py:69-78`; `derive_ablation` at `:92-104`.
  - Composition and thresholds: `compose.py:68-90`; `data/rounds/0/thresholds.json` (all three decisions `provisional-0.7`).
- **Item data:**
  - `data/rounds/0/arms-p1/V-sem.jsonl:1-20`; `arms-p2/V-sem.jsonl:1-12`.
  - `control-key.jsonl` (line N = ctl-N).
  - `gold.jsonl:37-82` (control gold) and `:6,27,32` (P2 source gold).
  - `labels.jsonl` at the lines cited in §2.
  - `arms/B2.jsonl:29,32,87,93`.
  - Area titles: `data/catalogue/catalogue.json`, `data/rounds/0/catalogue-p1.json`.
  - Strata and PRs: `units.jsonl`.
  - Choice and covered probabilities: `jev-p/scores.p1.jsonl`, `jev-p/scores.p2.jsonl`.
- **Ledgers:**
  - R-049: `docs/research-ledger.md:57`. R-051: `:59`. R-052: `:60`.
  - HD6 row: `docs/hypothesis-ledger.md:32`. The current state is `preregistered`.

## 5. Risks

- **The redesign is outcome-informed.** This council has seen per-item answers, choice probabilities, the labelled sets on P1 and the direct labels on the clones.
  - Mitigations: each repair rests on a fact that does not involve Jev; the evidence is fresh items only; bars are not moved towards these data; one shot only.
  - Residual risk: which constructs got repaired was still guided by where Jev failed. RESULTS must disclose this.
- **There may not be enough fresh items.**
  - The P pool is used up (26 eligible = 20 P1 + 6 P2 donors).
  - About 13 eligible D units remain [INFERENCE: 19 eligible minus 6 used].
  - Other sources: G minus the PRs used for guide examples, or relocations of C0-owned files that are in no unit.
  - If supply falls short of the rev-13 sizes, escalate to the user rather than shrink the sizes.
- **Direct-label positives are rarer.** Here only 4 of 12 clones were positives, so at least 24 clones are probably needed [INFERENCE]. Labeller splits (as on ctl-18) add adjudication noise.
- **The threshold knife edge is still there.** With the thresholds frozen at provisional 0.70, the fresh gates may fail again for threshold reasons and stop the pilot, even if the tuned behaviour would have been sound. That is the price of gating before tuning; (d′) is the alternative.
- **The stratum-1 weakness is real** and may be what fails. That would be a clean finding, not a defect.
- **P2's role is in tension with the user's direction.** P2 can veto with fewer than 10 cases, while R-051 made new-area detection non-vetoing. I recommend keeping P2 binding, because it protects against silent wrong assignment (`:134`). The rev-13 council should still rule on this explicitly.
- **Other confounds:** clustering (13 PRs for 20 items; use PR-clustered intervals); labellers seeing more than Jev; relocation keeps relative imports (ctl-24); the `catalogue-p1` ablation thins sample files.
- **Reversibility.** (c) can fall back to (a) at any point, and round-0 data stay append-only and untouched. The spend in (d) and (d′) cannot be undone.

## 6. Where the peers changed this memo

- **Innovation-challenger.**
  - Supplied "repair the measurement, not the model": the Jev side stays frozen so that the effect of the repair can be attributed. Adopted.
  - Supplied "no repeated testing until success", which became the one-shot rule. Adopted.
  - Pointed out that direct labels call ctl-44 and ctl-45 *new*, not *gap*, so the replacement gold must treat gap and new uniformly rather than pick whichever label suits. Adopted.
  - Proposed closing X2d as `inconclusive` and running a successor stage. **Not adopted.** In-study repair has precedent (rev 10, the K4 rule), and the HD6 window outcomes are untouched. The difference is mostly one of ledger labels.
- **Experiment-peer.**
  - Supplied the stratum analysis (2 of 11 in stratum 1), which moved P1 from "mostly mis-built" to "mixed, with a real stratum-1 weakness". This made stratum 1 decide the repaired gate.
  - Supplied the Wilson intervals, the 13-PR clustering, the labeller-versus-Jev visibility confound and the yield arithmetic.
  - Supplied the 0.5-threshold sensitivity. I added the P2 side: the same threshold that hurts P1 produced P2's `none` answers.
  - Proposed a test–retest re-ask. Kept only as an optional, gold-free step (§7).
- **My own contribution:** the clones' direct labels discarded by `gold.py:69-78`, which the challenger found independently, and the symmetric threshold coupling.

## 7. Execution outline (option c)

1. **Report now (rule 6).**
   - Append to `reviews/pilot-round-0.md`: the §2 tables, the intervals and the council-seen disclosure (rule 3).
   - Draft the RESULTS entries: the round-0 FAIL, the R10-02 split, the P1 disclosure and `Deviations`.
   - Add an HD6 History line: still `preregistered`; gate FAIL; this memo.
   - Record the user's decision in the research ledger as R-053.
2. **Optional, no gold, about $0.04.** Re-ask the unchanged p1 and p2 requests and compare probabilities only, to see whether the knife-edge items flip from noise alone. This needs the council's prior approval as a measurement check and is recorded in the review round.
3. **Count fresh supply from membership only.** Do this for the G non-guide clusters, the remaining eligible D units and the unused C0-owned files, before any label. Record the counts. Escalate if they fall short.
4. **Write PREREGISTRATION rev 13.**
   - The repairs in §1 (1)–(3).
   - Fresh-source exclusions (no exposed paths, twins or PR episodes); sizes (P1 ≥ 40, at least half in stratum 1; P2 sized for ≥ 8 direct-label positives with margin [INFERENCE]).
   - The Jev side unchanged; the one-shot rule.
   - A §14 map clause: "round-0 gate FAIL → `inconclusive` with the reason, never `not-supported`".
   - A ruling on whether P2 stays binding.
   - The rev-12 FAILs reported as they are.
5. **Change the code:**
   - `build_controls.py`: fresh pool, twin exclusion, stratum mix, the fallback-owner leak.
   - `gold.py`: P2 gold from the clone's own label, with the derived gold reported.
   - `analysis_x2d.py`: the P1 and P2 scoring.
   - `validation/validate.py`: planted-effect and planted-null fixtures for each rebuilt gate, plus fixtures for twins, the leak and direct-label gold.
   - `run.sh`: a stage for the fresh gate set that writes under a new tag and leaves `rounds/0` untouched.
   - Update README and DRAFT-NOTES.
6. **Council rounds 17 onwards**, until zero issues. Write the new `signoff.json` with the digests.
7. **Run:**
   - Build and record the fresh controls (a membership table before any label).
   - Label them, run Jev (p1 and p2 only, v1 questions) and analyse.
   - If both gates pass, write the round-0 critique and start round 1.
   - If either fails, stop and write RESULTS. HD6 → `inconclusive`.
   - Checks for the main agent: `run.sh validate`, `x2d.verify_pins()`, then the frozen-file digests against the sign-off.

**Decision needed from the user:** (c), recommended; (a), stop; or (d′), refine first and gate at v7. (b) and (d) are not recommended.
