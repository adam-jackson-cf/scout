# X2c post-run critical review — experiment-peer

Date 2026-10-29. Independent, blind to the other reviewer; read-only. Requested by the user after the run; not a council round.

## 1. Verdict

X2c is a clean, well-governed run of a different experiment from the one the user describes. It tested whole-tree agent re-maps reconciled one-to-one, with the candidates limited to pairs that share an owned file. That is not semantic association decided at each merge. Its `no-adoption` verdict is sound for "remap the whole tree and reconcile". The data mostly point to the mapper's run-to-run variance plus the one-to-one rule as the cause, not the matcher, and the result says nothing about the semantic-at-merge idea.

## 2. The user's challenge, answered

**No, X2c did not test the semantic-at-merge idea.** It differs from it in five ways.

- **What the study recorded.** The preregistration states the user's ask as "a Jev (TypeSafe) extension that matches areas by meaning rather than by file overlap" (`PREREGISTRATION.md:13`). R-048 narrows it to "a cheap, fast Jev extension of the agent map for area matching" (`docs/research-ledger.md:56`). I searched `docs/`, the X2c preregistration, draft notes, reviews and smoke notes for "semantic grouping/association", "at merge" and "per merged PR", and none of these phrases appears. Wherever the at-merge part was discussed, it never reached the ledger, the preregistration or the council.
- **File-sharing gate.** The candidates are "every pair … that shares at least one owned file" (`PREREGISTRATION.md:55`; `build_candidates.py` `candidates()`). The council's round 1 caught this: "'a matching method that does not need file overlap' is false (candidates require one shared owned file)" (`reviews/round-1.md:35`). The fix only changed the wording to "still requires one shared owned file" (`reviews/round-1-response.md:31`, `PREREGISTRATION.md` disclosure l.8). The gate removes exactly the cases the guide's R4 was written for ("Different files, same purpose … the mapper redrew the boundary", `x2-area-stability/GUIDE.md`). The draft notes rejected "all pairs" as "n² labels, infeasible" (`DRAFT-NOTES.md` judgement call 2). For Jev alone that is not true: a fastapi transition has about 20–68 × 11–68 pairs, which costs cents at the measured ≈$0.000055 per request (`data/jev/scores.jsonl`).
- **Candidate ceiling.** With the gate, the reported ceiling on fastapi is 0.634 (haiku), 0.779 (luna-xhigh) and 0.762 (terra-low), all below 0.80 (RESULTS §1). For three of the four arms the bar was out of reach before Jev ran.
- **Remap, not merge.** Every map (F1–F3, R1, R2, M1, M2) is a full agent re-map. Clawpatch 0.7.3 builds the agent prompt "from the project and the file inventory only; the prior features enter only … ID bookkeeping" (`x2-area-stability/PREREGISTRATION.md:79-80`). So the "same-state" re-maps R1/R2 are fresh draws. M1 and M2 are synthetic structural changes (add a package, rename a directory) followed by a full re-map (`:59-65`), not merged PRs. No decision is ever made at the moment a PR merges.
- **What Jev saw.** Jev's state was per-side `title`, `summary`, `kind`, ≤ 8 `entrypoints`, ≤ 40 `owned_files`, plus `shared_files` and `shared_files_total` (`jev/main.go:69-71`). That is a pairwise feature-versus-feature judgment. It never saw a PR, a diff, the catalogue of current areas, or the history of an area. The judgment Jev made was "do these two map features describe the same concern?", not "which existing area does this merged change belong to?".

**What the result says about the semantic-at-merge approach: almost nothing directly.** One indirect lesson: re-running the agent mapper is the dominant source of instability. On the unchanged fastapi tree:
- area counts swing between 29 and 68 for terra-none, and 15 and 35 for luna;
- owned files swing from 529 to 189 for luna-xhigh in T2 (`analysis_output.txt:424`, RESULTS §6).

A design that never re-maps, and only associates changes at merge time, avoids that source of noise entirely. The X2c failure is therefore weak evidence *for* trying the merge approach, not against it [INFERENCE].

## 3. Issues

| id | severity | finding | evidence | consequence |
|---|---|---|---|---|
| I1 | critical | The construct does not match the user's idea: whole-tree re-maps plus pairwise reconcile, not per-merge semantic association. | X2 prereg `:56-60, 79-82`; X2c §2 reuses X2's seven maps; RESULTS §1 says the result concerns "refresh on merge". | The result cannot be cited against semantic-at-merge. |
| I2 | critical | The file-sharing candidate gate contradicts "by meaning rather than file overlap". The council flagged it in round 1; only the wording changed, and the user was not asked again. | `PREREGISTRATION.md:13` vs `:55`; `reviews/round-1.md:35`; response `:31`; GUIDE R4. | The Jev arm could not match a redrawn area that shares no owned file. Three of four arms had a ceiling below 0.80 on fastapi by construction. |
| I3 | major | The ceiling ignores the one-to-one constraint. The bound min(non-zero-overlap priors, next features) per transition, from RESULTS §6, gives fastapi ≤ 0.625 (haiku), ≤ 0.686 (luna), ≤ 0.746 (terra-low), ≤ 0.882 (terra-none). RESULTS reports 0.634, 0.779, 0.762, 0.959. Jev already gets 86–89 % of the bound. | My arithmetic on RESULTS §6 plus §3.4. Several transitions sit exactly at the bound: terra-none+jev T4 38/68 with 38 next features; haiku+jev T4 9/11 and T5 3/12, equal to the zero-overlap limit; terra-low+jev T1 11/25 and T3 16/25. | The "gap to ceiling" column (0.171–0.190) overstates how much a better matcher could gain. The failure lies in the maps and the one-to-one rule. |
| I4 | major | The mapper's run-to-run variance is the binding cause, but RESULTS leaves it as an `[INFERENCE]` and the headline credits the matching. | On the unchanged tree (T3/T4), haiku fastapi Jaccard T3 is 2/20. terra-low T1 has 14 of 25 prior areas with zero overlap (345 → 145 owned files). luna T2 owned files 529 → 189. X2b and X2c luna draws differ: 0.572 vs 0.473. | The study says something about Clawpatch re-mapping, not about matching by meaning. |
| I5 | major | The Jev score saturates, so the threshold grid does nothing. s = P(same) + P(overlapping) sits near 1 for any non-`different` pair. terra-low+jev is flat across all 13 thresholds, terra-none+jev across 0.30–0.80. The thresholds are X2's Jaccard grid reused, not tuned. | RESULTS §1 plateau column; `jev_match.py` `read_scores`; sample rows in `scores.jsonl:2-7` (s 0.92–1.0 for `overlapping` answers). | In practice Jev was a binary "not different" filter plus ordering by P(same). The "Jev threshold" dimension was never really tested. |
| I6 | major | C2 passes vacuously, and the labels are weakest where Jev added matches. `overlapping` counts as non-spurious. Swapped-control misses cluster in the lowest Jaccard band: 6 of 57 (≈ 10.5 %). | RESULTS §7.2. The lowest band floors at 0.30 and so includes J < 0.30 (`x2-area-stability/analysis.py:331`). Strict rate haiku axum 0.596 at 0.30 (RESULTS §14). | "0 spurious" for Jev's low-overlap matches rests on the least-validated label regime. The decision collapsed to C1 alone. |
| I7 | major | Jev's multi-refresh behaviour was not measured. The `+jev` chain, created and kept fields are placeholders. | RESULTS §3.6 ("empty … the `+jev` cells carry no chain"). In the Jaccard chain, 0/35 luna fastapi F1 areas survive every step. | For any repeated refresh, accumulated survival is the relevant quantity, and the Jev arm never reports it. |
| I8 | major | Survival does not measure the product harm. In Reconcile an unmatched prior area persists as `kept`, so "not matched" is not "identity lost". The real harms are near-duplicate areas, stale file sets and moved issue anchors, and none is measured. | X2 prereg §3.6 (`:113`: "unmatched areas persist as kept"); design §4 Areas (`docs/scout-design.md:106`). | C1 penalises a re-map's noise in a way that does not map directly onto user-visible damage. |
| I9 | major | The 0.80 per-refresh bar is inherited from X2 and has no product derivation. Under refresh-on-merge it compounds (about 0.8^k after k refreshes, if independent), so it is too lenient for frequent refreshes. For re-maps it is unreachable. "Both repos" is reasonable but n = 2, and the decision rests on fastapi alone. | PREREGISTRATION §7; chain data in RESULTS §3.6. | Neither the bar nor its failure is calibrated to what Scout needs. |
| I10 | major | Jev's accuracy per pair was never measured, although all 1,986 candidates have `agent_reviewed` labels. The preregistration bars "precision/recall curve, AUC, calibration, `tune_threshold.py`". | `PREREGISTRATION.md:55`. | Q2 ("did Jev change the verdict") is answered without knowing whether Jev classifies pairs well. |
| I11 | minor | Two cheap controls are missing. (a) A deterministic Jaccard > 0 greedy matcher, to tell "semantics" from "just a lower overlap floor". (b) An oracle maximum matching over the label-accepted pairs, as the true upper bound. | No such rows in `results.json` (keys `x2c.*`, `arms.*`). | Jev's +0.03 to +0.14 cannot be attributed to meaning. |
| I12 | minor | Jev and the labellers saw the same file-heavy state, and the labels are not blind to overlap. | `jev/main.go:69-71`; X2 prereg `:151`. | "By meaning" was partly by overlap, and any agreement is internal consistency. |
| I13 | minor | Labelling was interrupted "to consider a protocol change, then withdrew that change", with no description of what prompted it. κ covers only 260 of 1,986 real pairs (13 %). | RESULTS §12 note 2; §7.1. | It cannot be shown that the pause was not driven by the data. Agreement over the full candidate set (rule 6 completeness) is unknown. |
| I14 | minor | Collection took 7.7 h, but the per-map seconds sum to 7 × 917.9 ≈ 6,425 s (≈ 1.8 h). The gap is unexplained. | RESULTS §14 (7.7 h) vs §8 table (my sum). | The "refresh time" may leave out real per-refresh overhead. It decided nothing here. |
| I15 | minor | Arms differ in granularity, so survival is not comparable across arms: area ratio 0.25–1.54, and haiku fails the coverage guard. | RESULTS §6. | Comparisons between arms confound granularity with stability. |
| I16 | minor | The analysis output prints `INCONSISTENT … not citable` for the `+jev` pseudo-arms while the cell lines say citable. | RESULTS §12 note 4. | A cosmetic defect in a frozen script. Fix it in the next round. |

Execution otherwise matches the preregistration as recorded:
- all 26 digests equal;
- 56/56 maps ran on their first attempt;
- 0 unscored Jev units, 12 schema failures retried;
- running Jev concurrently with labelling is permitted.

## 4. What the result does and does not license

**It licenses:**
- For Clawpatch 0.7.3 whole-tree agent re-maps on fastapi and axum, with one draw per arm, one-to-one matching restricted to file-sharing pairs, and either Jaccard or the Jev concern Choice: no cell keeps ≥ 0.80 of prior areas on fastapi. So refresh-by-re-map plus Reconcile is not supported. The rev 2.6/2.7 fallback stands for *that* mechanism.
- For three of the four arms this is a property of the maps, under any matcher limited to file-sharing, one-to-one pairs.
- Jev costs about $0.002–0.012 per refresh and raised survival in every arm, with no labelled `different` merges among its matches.

**It does not license:**
- Anything about associating changes with areas at merge time.
- Anything about matching without the file gate, or many-to-one area lineage (splits and merges of areas).
- Jev's accuracy per pair, or whether calibrating or tuning Jev would help.
- Whether a cheaper mapper is "as good" (only a crude coverage proxy was used).
- Generalisation beyond two repositories and one draw each.

## 5. Recommended next experiment

**Competing explanations to settle first.** All of these use existing X2c data and no model calls. They need a short signed-off exploratory preregistration, because two of them compare Jev with labels (CLAUDE.md "Jev labelling").

| hypothesis | prediction | cheapest refuting observation |
|---|---|---|
| H-map: re-map variance dominates | An oracle maximum one-to-one matching over label-accepted candidate pairs stays below 0.80 on fastapi for haiku, luna and terra-low. | Oracle ≥ 0.80 in any of these arms. |
| H-floor: Jev's gain comes from the lower floor, not semantics | Greedy Jaccard > 0 matches Jev's survival within ~0.02 but picks up labelled `different` merges. | Jaccard > 0 survival well below Jev's, or zero `different` merges. |
| H-1:1: the one-to-one rule binds | Many-to-one association (lineage) on `overlapping` sub-part pairs lifts terra-none fastapi above 0.80. | No lift. |
| H-Jev: Jev classifies pairs well | On dev = axum, test = fastapi, the tuned "not different" classifier reaches precision and recall ≥ 0.8. | Fails on test. |

**Main study: merge-time semantic association** (call it X2d; product hypothesis HD-new; also record the user's direction as a new R-row).

- **Unit.** One merged PR, decided against the current area catalogue. Within a PR, files that are already owned and unmoved are resolved by lookup in code (the skill: "Keep known rules … exact lookups … in code"). The decision unit is each new, moved or renamed file, or each new directory, plus a PR-level "which areas does this touch".
- **Start state.** One onboarding map per repository from the production mapper, frozen at commit C0, N merged PRs before the pinned revision. A second onboarding draw serves only as a sensitivity check on the start map.
- **Candidate generation with no file gate.** The whole catalogue (17–89 areas) plus an explicit `none → new area` option. Either one Noul per (unit, area), which is cheap at these sizes, or a code pre-rank to top-k with *measured* candidate coverage. The skill's question-design page: "check candidate coverage; the model cannot choose an omitted value."
- **When it is evaluated.** Sequential replay, PR by PR. The areas are updated after each decision, so drift accumulates exactly as it would in the product.
- **Arms.**
  1. A deterministic baseline: owner of the nearest owned sibling or parent directory.
  2. Jev.
  3. An agent judgment through `claude -p` or `codex exec` (the CLAUDE.md model rule).
  4. Optionally a cascade: Jev confident, agent inside the band.
- **Oracle and labels.** A frozen guide; blind `claude -p` and `codex exec` sessions plus an adjudicator. The labellers see the PR title and description, changed paths, a diff summary and the catalogue (titles, summaries, sample files), never Jev's answer. Provenance `agent_reviewed`.
- **Strata.**
  - new file in an existing package;
  - new package;
  - rename or move;
  - cross-area refactor;
  - docs-only or test-only changes.
- **Outcomes.**
  - Precision and recall of "assign to existing area" and of "new area".
  - Wrong-area rate ≤ 0.05 with a Wilson bound.
  - The method must beat the baseline by a preregistered margin.
  - End-of-replay coherence of the touched areas: a labelled check that the file set still matches the title and summary.
  - Untouched areas keep their identity by construction; report that as 1.0, not as evidence.
- **Split.** Dev = older PRs or axum; test = newer PRs or fastapi, scored once.
- **Validation.**
  - Planted PRs with a known area: moves, new modules.
  - Planted null: a shuffled catalogue must give accuracy at chance.
  - Planted "new concern" units must get `none`.
- **Minimal size.** 2 repositories × about 150 merged PRs, giving ≥ 30, and ideally 50+, labelled units per label for each decision (sizing per `building-eval-data.md`). A third repository in a different language is preferable, given that n = 2 repositories made fastapi decisive in X2c.

## 6. Points where I disagree with RESULTS

1. **The §1 plain-terms sentence**, "nor matching areas by meaning with Jev instead of by file overlap makes areas stable enough … to refresh on merge". Jev was gated by file overlap and saw file lists, and the mechanism tested was re-mapping, not merge. The HD5c ledger row is correctly scoped ("on file-sharing candidate pairs"); the RESULTS headline is not.
2. **The candidate ceiling as "the most survival any matcher can reach".** It ignores the one-to-one constraint, so the gap column overstates headroom about twofold for luna (0.190 vs ≤ 0.097) and terra-none (0.171 vs ≤ 0.093).
3. **Underclaim.** The data identify map variance plus the one-to-one rule as the binding cause:
   - same-state re-maps are unstable;
   - Jev hits the bound in several transitions;
   - the X2b and X2c luna draws differ by 0.10.

   RESULTS marks this only `[INFERENCE]`. It is the most decision-relevant finding, because it says no matcher will rescue re-mapping.
4. **"Spurious merges were 0" presented as a positive.** C2 was vacuous (RESULTS §14 admits it). Strict rates are large, and the controls are weakest in the low-overlap band Jev exploits.
5. **The "terra-none+jev could land either side" note** (§1, §14). Its T4 is at the cardinality bound (38/68). Passing needs a different map, not a luckier match.
6. **Deviation 2.** It should name the protocol change that was considered and state that no labels or outcomes were inspected.

## 7. Improving the Jev usage

The sources are the jev-actions skill (cited by reference file and bullet) and Hamel Husain and Antaripa Saha, ["How To Use Decision Models for Evals"](https://hamel.dev/blog/posts/decision-model-evals/) (2026-10-20).

- **Pattern and candidate coverage.** `design-patterns.md` "Select instead of generate: find candidate values … in code, use a judgment to select the intended one". `question-design.md`: "For source-value selection, check candidate coverage; the model cannot choose an omitted value." X2c computed coverage (the ceiling) only after the run. It should have been a design gate, and no-gate candidates or a measured top-k should replace the file gate.
- **Question design.**
  - `question-design.md`: "Prefer binary criteria … for eval judgments" and "Ask one narrow, coherent judgment per question."
  - The decision actually needed is binary: "should B inherit A's identity (would issues about A belong to B)?" X2c used a 4-way Choice and then summed same and overlapping, which saturated s (I5). That sum also folds in the guide's R6 tie-break toward `overlapping`.
  - Use a Noul, with explicit exclusions, for that question.
  - Keep `overlapping` granularity as a separate Noul if lineage (splits and merges of areas) matters.
  - For merge-time selection, "Include a no-match outcome".
- **State.**
  - `question-design.md`: "Supply the reference evidence each judgment needs." For merge-time association that means the PR evidence and the catalogue, not two map features.
  - To test "by meaning", add a state variant without `owned_files` and `shared_files` (I12).
  - Long state (40 files a side) is a known weak-spot class. The 1.14 jaggedness page returned 404 and the 1.13 page was not applied (`PREREGISTRATION.md` §6 item 1). Record this as a limitation and test long-state cases.
- **Eval data and framing.**
  - `building-eval-data.md`: "Replay existing … commits now; a few hundred real units today beat a week of shadow traffic." Merged-PR history is exactly that source.
  - "Stratify … by failure mode … Give every stratum both labels, ideally 10 or more of each."
  - "Discover failure modes before writing judges. Read 30 to 50 real inputs." X2c read 30 X2b pairs, which is good. Repeat this on PR units.
  - The framing STOP (action, baseline, costs) was answered by assumption. That was legitimate, but the user's ask itself was not re-confirmed after the council changed it (I2).
- **Threshold tuning.**
  - `threshold-tuning.md`: "Never ship 0.5 or a cookbook number as a tuned threshold" and "Score each question once … tune on the saved scores … stratified five-fold cross-validation." X2c's 1,986 labelled pairs were ready-made `scores.csv` material, but the preregistration forbade `tune_threshold.py` (I10).
  - Hamel: "Tune each model's decision threshold before comparing performance", and choose the metric deliberately (F-beta with β < 1, because a false merge costs more).
  - The skill: "Compare Jev with an alternative … on tuned held-out performance … never on default-threshold results." X2c's Jev-vs-Jaccard Q2 compared the two methods on a shared grid with no tuning.
- **Dev/test split.**
  - `building-eval-data.md`: "Split by an independent unit, such as … repository … score `test` once." X2c applied the split only as a confirmation for `adopt` cells.
  - Next time: dev = axum or older PRs; test = fastapi or newer PRs, frozen before any wording iteration.
  - The skill: "Once test results have been seen, build a fresh test set." X2c's labels are now dev data at best.
- **Error analysis on real inputs.**
  - `error-analysis.md`: list the false positives and false negatives on dev and give each error one cause (label, policy gap, missing state, wording, model limit).
  - Run a perception check: "two blind labelers answer Jev's exact question … If Jev agrees with their consensus about as often as they agree with each other, attribute the remaining errors to policy, state, composition or labels." X2c had the data for this (κ 0.886 on 260 pairs) but the preregistration forbade the comparison.
  - The guide's tie-break should be audited: "When many labels cite one rule, check whether a labeler read that rule too broadly." R5/R6 citations dominate the label notes (`data/pairs/labels.jsonl`).
- **Calibration and escalation.**
  - Hamel: "If you want to use scores as confidence estimates, check whether the probabilities match observed outcomes"; Jev 1.13 was well calibrated on their dataset (Brier 0.187), but that must be measured per task.
  - `error-analysis.md`: add an escalation band to an agent (`claude -p` or `codex exec`) "only after measuring both stages on the same labeled cases", preferring "a fallback with different error patterns".
  - Hamel's model-selection advice: "start with a larger model that has acceptable performance … then try smaller models." For X2d, measure the agent arm first and treat Jev as the cheaper candidate on the cost–performance frontier.
- **Labels.** `labeling-cases.md`: "LLM judges tend to repeat Jev's confident errors, so their agreement with Jev is weak evidence of correctness". Its `agent_reviewed` definition asks for independent sessions, "ideally from different model families". X2c used different-family labellers (claude-opus / gpt-6.1-sol), which satisfies that. They are, however, the mappers' own families (RESULTS §7.1), so in X2d keep the mapper and labeller families apart where possible, and keep "Keep cases that labelers cannot agree on out of tuning".
