# X2c post-run critical review — innovation-challenger (different model family)

Date 2026-10-29. Independent, blind to the other reviewer; read-only. Requested by the user after the run; not a council round.

## 1. Verdict

X2c supports **no adoption of its tested pipelines**, but not rejection of semantic association at merge (`evaluation/x2c-mapper-comparison/RESULTS.md:10-18`). The user’s objection is substantively correct: X2c compared independently regenerated maps using file-gated candidate pairs, not persistent semantic groups updated from merged PRs. Retain the existing product policy pending a properly targeted experiment, and correct the report’s overbroad headline and demonstrably incorrect leave-one-transition-out summary.

## 2. The user’s challenge, answered

**Yes: the experiment tested a narrower, materially different idea.** It did include semantic judgment, but only **after file overlap had decided which relationships could exist**.

For brevity, references below use:

- `X2c/` = `evaluation/x2c-mapper-comparison/`
- `X2/` = `evaluation/x2-area-stability/`

| Intended idea | What X2c actually tested |
|---|---|
| Associate code by feature responsibility, without requiring common files | Every candidate required at least one shared owned-file path; zero-overlap semantic continuations were excluded before Jev (`X2c/build_candidates.py:31-40`). |
| Consider association when a PR merges | Four unchanged-tree transitions and two synthetic mutations: adding a utility package and renaming a directory; no historical merged-PR sequence (`X2/PREREGISTRATION.md:54-81`). |
| Use the change and existing semantic context | Jev saw titles, summaries, kind, entrypoint paths, truncated owned-file lists and overlap counts—not PR intent, diff, source bodies, dependency evidence or persistent association history (`X2c/PREREGISTRATION.md:66`; `X2c/jev/main.go:68-101`). |
| Maintain durable semantic groups | One-to-one matching between regenerated features; Jev’s multi-refresh lifecycle was unmeasured (`X2c/jev_match.py:40-50`; `X2c/RESULTS.md:539-559`). |

This is **semantic reranking of a file-derived candidate graph**, not file-independent semantic association.

The distinction matters numerically: the recorded fastapi candidate ceilings were **0.6336, 0.7787 and 0.7622** for haiku, luna-xhigh and terra-low. Those three arms could not meet 0.80 even with perfect judgments over their permitted candidates (`X2c/RESULTS.md:37-46`). Terra-none had more candidate coverage, but remained subject to one-to-one capacity and unstable map granularity.

**What this says about semantic-at-merge:** it supplies useful warnings about full-remap instability and candidate recall. It neither demonstrates nor disproves an incremental semantic-at-merge approach.

## 3. Issues

| ID | Severity | Finding | Evidence | Consequence |
|---|---|---|---|---|
| C1 | **Critical** | **The plain-language conclusion exceeds the tested construct.** “Matching areas by meaning … instead of by file overlap” is not the implemented pipeline, and refresh-on-merge was not exercised. | `X2c/RESULTS.md:12`; `X2c/PREREGISTRATION.md:55,66,71`; `X2/PREREGISTRATION.md:62-81`. | [INFERENCE] Readers could reject the user’s proposal based on failure of a different architecture. |
| M1 | **Major** | **The file gate makes three arms incapable of passing.** Zero-overlap continuations receive neither semantic scores nor oracle labels. | `X2c/build_candidates.py:31-40`; `X2c/RESULTS.md:37-46`. | Failure cannot isolate semantic scorer quality. Removing the gate might help, but the study contains no labels establishing how many excluded pairs truly continue a concern. |
| M2 | **Major** | **The reported candidate ceiling is only an upper bound, not attainable matching capacity.** One-to-one collisions and feature-count contraction impose additional limits. | `X2c/analysis_x2c.py:155-162`; `X2c/jev_match.py:40-50`; `X2c/RESULTS.md:653`. | Terra-none fastapi T4 has **68 prior areas and 38 next features**: even unrestricted one-to-one matching cannot exceed **38/68**, which Jev already achieves (`RESULTS.md:443`). [INFERENCE] That transition’s loss is not fixable by a better scorer; the 0.171 “gap to ceiling” cannot all be attributed to matching quality. |
| M3 | **Major** | **“Survival” does not measure durable semantic correctness.** It counts matches, including overlapping concerns; unmatched areas can remain active. Jev’s chained lifecycle is absent. | `X2/PREREGISTRATION.md:17-20,92-103`; `X2c/RESULTS.md:366,539-559`. | The best fastapi Jev cell has **73 overlapping labels among 173 matches**, strict rate **0.422**. Zero `different_concern` labels does not establish correct identity preservation or absence of semantic drift. |
| M4 | **Major** | **The oracle validates descriptions, not code-grounded responsibility.** The guide forbids judging feature correctness; scorer and labellers receive the same mapper-authored evidence. | `X2/GUIDE.md:3-15,27-37`; `X2c/PREREGISTRATION.md:55,66`; `X2c/RESULTS.md:665-675`. | [INFERENCE] Consistently misleading summaries can yield agreement without correct associations. Blinding is useful, but it cannot supply missing source evidence. |
| M5 | **Major** | **The evidence does not resolve the 0.80 boundary statistically or operationally.** Six dependent transitions share maps, there are two repositories and one sequence per arm, and pooled Wilson intervals are not intervals for the primary mean. | `X2/PREREGISTRATION.md:71-81,93-94`; `X2c/RESULTS.md:35,79,859`. | Keep the preregistered failure; do not reinterpret 0.7887 as a pass. Equally, do not treat this as proof that the configuration’s repeat-run performance is below 0.80. |
| M6 | **Major** | **RESULTS’ leave-one-transition-out statement is false.** It says every held-out case has no selectable threshold. | `X2c/RESULTS.md:418-420` versus `X2c/data/analysis_output.txt:1303-1305` and `data/results.json → arms["terra-none+jev"].loto`. | Holding out T4 or T5 selects **0.90**; holding out T6 selects **0.80**. Their held-out fastapi survivals still fail: **0.5588235294117647**, **0.6052631578947368**, **0.7241379310344828**. The full-data verdict stands, but its sensitivity narrative needs correction. |
| M7 | **Major** | **Mapper quality and granularity confound the arm comparison.** The coverage floor was chosen after smoke counts; owning files is not evidence of coherent concerns. | `X2c/PREREGISTRATION.md:82`; `X2c/DRAFT-NOTES.md:39-45`; `X2c/RESULTS.md:634-641`. | Terra-none has 0.4516 times luna’s area count on axum but 1.5443 times on fastapi. [INFERENCE] Arm comparisons mix model choice, coverage and grouping granularity, rather than measuring a common semantic task. The guard did not cause the observed primary failures. |
| m1 | **Minor** | **Controls establish limited label hygiene, not the required semantic error rate.** Swaps create inconsistent records; identical pairs are attention checks; controls never reach Jev. | `X2/PREREGISTRATION.md:153-165`; `X2c/PREREGISTRATION.md:55`; `X2c/RESULTS.md:670-675`. | Passing 159 swapped and 80 identical controls does not establish recall for zero-overlap continuations, correctness of natural hard negatives, or Jev calibration. Synthetic scorer validation tests wiring, not real Jev judgment. |
| m2 | **Minor** | **The interrupted label run materially narrowed reported agreement coverage.** The blanket claim that deviations change no number is too broad. | `X2c/RESULTS.md:663-667,828-835`. | κ = 0.8857231501111743 covers **260**, not all **1,986**, real pairs. Resume is permitted and retained labels were not replaced, but this is not full-corpus reliability evidence. |
| m3 | **Minor** | **“Cheap and fast” is only partly measured.** Mapper cost is absent; Jev refresh time is summed service time divided by eight, not observed end-to-end merge latency. | `X2c/RESULTS.md:598,766,864`; `X2c/analysis_x2c.py:101-122`. | Cost/latency feasibility is encouraging for this candidate set, not demonstrated for unrestricted candidates or an operational merge updater. |

### Assessment of execution

The recorded conduct is substantially stronger than the construct validity:

- Round 5 explicitly reissued sign-off after the frozen validation-file change; RESULTS records digest equality and pre-run synthetic validation (`X2c/reviews/round-5.md:57-75`; `RESULTS.md:770-824`). **I inspected these records, not independently recomputed hashes or reran validation.**
- The label resume follows the inherited append-only recovery mechanism (`X2/label.py:151-192`). The report records zero labelled pairs without call rows.
- Jev’s twelve schema failures were resumed before analysis, as authorized. Stored answers end at **08:15:23Z**; RESULTS places analysis at **09:02:39Z** (`data/jev/scores.jsonl:3984-3985`; `RESULTS.md:823,832`).
- Concurrent scoring and labelling do not expose labels to Jev: its input contains pair states only. [INFERENCE] Concurrency is not evidence of label leakage.
- The `+jev` probe inconsistency is a presentation defect: the integration checks the base arm’s probes (`analysis_x2c.py:286-296`). It does not independently invalidate the primary result.
- The operator’s abandoned protocol-change consideration is insufficiently described to reconstruct the decision context. Record what was considered and what had been inspected; **there is no evidence here that the protocol actually changed**.

## 4. What the result does and does not license

### It licenses

- **No adoption under the frozen rule:** every tested cell fails C1 on fastapi.
- Keeping the existing onboarding/manual policy while alternatives remain unvalidated.
- Reporting that Jev increased recorded best-threshold survival in all four arms, without changing adoption.
- Treating full-remap variability as a serious engineering risk: feature counts change sharply even on an unchanged tree.

The **both-repositories rule is reasonable as a conservative product acceptance condition**: a good axum result should not conceal failure on fastapi. It is not evidence of broad repository generalization, and the cross-repository confirmation gate never determined this outcome because no cell initially adopted.

The **0.80 rule is defensible as a frozen screening assumption**, not an established user-utility threshold. [INFERENCE] For frequent merges, permitting substantial mismatch on each refresh may be too permissive for persistent groups; for real concern deletion or deliberate splits, raw survival can instead punish correct behavior. The next study must distinguish those events.

### It does not license

- “Semantic matching does not work.”
- “Semantic association at merge cannot work.”
- “The failures are entirely mapper failures” or “entirely matcher failures.”
- “Spurious semantic identity changes were zero.”
- “Terra-none is generally the best mapper,” or that any mapper is cheaper in measured token cost.
- Treating HD5c as refuted.

**HD5c’s `not-supported` state is defensible:** its actual ledger wording explicitly limits the claim to these arms, methods and **file-sharing candidates** (`docs/hypothesis-ledger.md:31-35`). The overclaim is principally the report’s broad paraphrase, not that scoped state. The ledger outcome also omits the per-cell intervals requested by the preregistration (`X2c/PREREGISTRATION.md:89`).

## 5. Recommended next experiment

### Recommendation: update persistent semantic groups from merged changes

**Do not merely repeat X2c with another mapper.** Test whether a merged change belongs to existing responsibilities, changes their boundaries, creates a new responsibility, or genuinely splits/merges/retires one.

This is a proposal, not an adopted design change.

1. **Unit and moment**
   - Unit: a merged PR’s concern-level change and its association with the area state immediately before that merge.
   - Process actual parent-to-merged-tree changes in chronological order, once per merged PR.
   - Carry each arm’s resulting state forward; never reset areas from a newly generated map between primary observations.
   - Score unaffected-area stability separately so copying everything unchanged cannot masquerade as success.

2. **Minimal comparison**
   - Use one frozen mapper configuration initially, rather than another four-model search.
   - Compare **incremental semantic association** against the existing frozen-map reuse baseline and a full-remap/Jaccard reference.
   - Keep initial areas identical across arms. Repeat the chronological sequence with independent model sessions to separate sampling variability from PR effects.
   - Use two repository histories for a bounded feasibility study, with a development prefix and untouched test suffix. Freeze window selection, sample size and precision requirements before scoring; an undersized outcome remains inconclusive.

3. **Candidates without a file gate**
   - For the small study, consider **every existing area** for each changed concern; zero-overlap candidates must remain eligible.
   - Use semantic descriptions, APIs/symbols and dependency relationships to order candidates—not shared paths as an exclusion rule.
   - Include “new concern,” “no relevant semantic change” and “insufficient evidence.”
   - Only introduce top-K retrieval after measuring its oracle candidate recall, including an audit of excluded candidates.

4. **State**
   - Supply the prior area’s responsibility and boundaries, relevant before/after code, diff, entrypoints, dependency evidence, and PR intent available at merge time.
   - Treat PR text as untrusted evidence, not authoritative truth.
   - Preserve association history and distinguish membership updates from identity changes.

5. **Oracle and labels**
   - Freeze a source-grounded guide before running.
   - Use the required two independent blind sessions—one Claude, one Codex—with adjudication, retaining `agent_reviewed` provenance.
   - Reviewers receive code/diff/contracts/tests and the complete candidate inventory, but no method predictions or scores.
   - Label true continuation, new concern, retirement, split, merge, unrelated association and uncertainty. Allow multiple valid associations where responsibilities genuinely overlap.
   - Add controls for zero-overlap continuity, shared generic files with different responsibilities, cross-cutting PRs, and genuine concern replacement. Report complete-run agreement, not just the first execution segment.

6. **Primary outcomes**
   - Correct association precision and recall, especially **zero-file-overlap recall**.
   - False identity merges and unnecessary identity creation.
   - Multi-merge responsibility drift and correct handling of splits/retirements.
   - Candidate recall separately from scorer accuracy.
   - Actual merge-processing latency, failures, cost and unmeasured counts.
   - Intervals respecting PR/history dependence, rather than treating all pair decisions as independent.

7. **Decision rule**
   - Fix acceptable false-association cost, semantic granularity and treatment of splits/merges before collection.
   - Tune once on development data; evaluate frozen rules on held-out histories.
   - Do not transplant X2c’s raw matched-area survival bar as semantic recall.
   - Obtain council zero-issue sign-off on protocol and implementation, including planted-effect/null recovery, before the run.

### Viable alternative

Retain full remapping, but match **all semantic candidates** and represent explicit split/merge lineage rather than forcing one-to-one identity replacement.

[INFERENCE] This alternative may be appropriate if periodic rediscovery is itself valuable. Choose it over incremental updating only if source-grounded evaluation shows that full maps recover important new responsibilities which the incremental approach misses, at acceptable drift and cost. Simply dropping the file gate will not fix the observed 68-to-38 contraction.

## 6. Points where I disagree with the study’s own RESULTS

1. **Headline, line 12:** replace the general rejection of “matching areas by meaning” with failure of **file-gated, one-to-one matching of regenerated maps under the tested schedule**.
2. **Candidate ceiling, lines 37–46:** call it a **zero-overlap upper bound**, not attainable maximum survival or a purely mapper-intrinsic property. It depends on candidate policy and ignores matching capacity.
3. **Leave-one-transition-out, line 420:** correct the three non-null terra-none+jev cases. Existing JSON and text output already contain the evidence; no rerun is needed.
4. **Deviation preamble, line 828:** “none changes a number” is not justified; interruption changed which agreement statistics were reported.
5. **C2 “passes vacuously,” line 861:** zero observed negatives among labelled matches is an empirical result, not a vacuous truth. The correct caution is that the labels and accepted-overlap definition do not establish safe semantic identity preservation.
6. **Mapper-versus-matcher explanation:** map variability is demonstrated, but its semantic cause is not. [INFERENCE] The data require separating omitted concerns, renamed paths, granularity contraction, candidate exclusion and assignment conflicts before attributing failure.

**Bottom line:** preserve the narrow negative result; do not let it answer the untested semantic-at-merge question.
