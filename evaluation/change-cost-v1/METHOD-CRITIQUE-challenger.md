# E1/E1b: independent methodological challenge

**Recommendation:** retain descriptive-only Scout signals. Do not interpret either study as refuting H1/H2/H4, and do not make “collect more windows” the first remedy. Repair measurement and separate the prediction question from the causal question first.

References below use **E1** = `evaluation/inertia-v1/`, **E1b** = `evaluation/change-cost-v1/`.

## Ranked findings

### 1. Construct mismatch — **invalidates broad hypothesis rejection**
E1 predicts regex-labelled fix-touch counts. E1b predicts churn, files touched, repeat touches, PR activity/elapsed time, and dispersion ratios (`PREREGISTRATION.md:30–55`; `features.py:90–117`). None measures active human effort. Repeat touches include ordinary continued development. Coupling explicitly includes dependencies explainable by imports.

**[INFERENCE]** H1 is only indirectly operationalized; H2's unconditional dispersion confounds unpredictability with task mix; H4's hidden-coupling/incomplete-change mechanism is not identified. A reliable predictor of these proxies would not establish debt-caused cost; a null predictor cannot refute that mechanism. Log-MSE reduction is not percentage effort saved.

### 2. Collection defects — **invalidates affected measurement definitions; estimate impact unquantified**
- E1b `collect.py:247–249` drops file-deletion touches; `:265–266` assigns area deletions zero churn, contradicting added-plus-removed lines.
- `features.py:108–117` accepts any later-sequence touch satisfying `t-ts <= 7*DAY`, including negative differences. Sequence follows topological order (`collect.py:235–236`), not guaranteed timestamp order.
- Follow-up streams are confined to each window (`features.py:137–145`), without complete-observation eligibility at the right boundary.
- PR linkage unions merge and constituent OIDs, but `linked_prs` never checks the PR merge date against its requested window (`features.py:147–178`).

**[INFERENCE]** A PR spanning T0 can therefore contaminate lagged controls with post-T0 information. Actual leakage, timestamp reversals, and deletion effects on estimates were not quantified. Require raw-event audits and corrected sensitivity analyses—not an assumed direction of bias.

### 3. Predictor centering and uncertainty target — **weakens prediction conclusions substantially**
E1b `model.py:36–84` standardizes predictors separately within every repository, including the held-out repository, but leaves outcomes uncentered and fits one intercept. The PR models contain a singleton JUnit repository (`model-results.json`, `.area.o4c.counts`).

**[INFERENCE, algebraic]** Each held-out repository's average prediction is the training outcome grand mean; singleton predictors all become zero. Between-repository predictor information is removed while between-repository outcome shifts remain in MSE. This is a restrictive, transductive prediction target, not ordinary deployment calibration. It is not direct outcome leakage merely because held-out predictors are used.

The gain bootstrap resamples units within fixed repositories using fixed predictions; it does not refit or retune (`model.py:86–112`). **[INFERENCE]** It omits training/new-repository uncertainty and treats cochanged areas as exchangeable independent units. Coefficient bootstraps refit but fix tuning and scaling. E1 uses the same broad modeling approach (`model.py`).

### 4. Selection, populations, and adjustment — **weakens generalizability and causal interpretation**
E1b requires future activity and historical area existence; PR models additionally require measured PR outcomes in both windows (`features.py:193–219`; `model.py:41`). Structural deltas compare changing file populations (`features.py:181–214`). E1 excludes files deleted before the endpoint (`RESULTS.md:112`).

**[INFERENCE]** Avoided, inactive, deleted, or reorganized code can disappear from the target population. Growth/moves/removals can masquerade as degradation. Lagged outcomes are legitimate controls for incremental forecasting, but may absorb persistent quality-related costs. Removing them does not establish causality either. A temporal causal diagram and an explicit target population are needed; neither “always control history” nor “history is necessarily a mediator” is defensible.

### 5. Decision rules and measurement error — **weakens interpretation**
E1's quality-set gain is **1.18%, CI [−0.62%, 3.24%]** (`RESULTS.md:15`), despite positive nominal individual ablations. Its classifier precision is **0.524 overall**, **0.636 in the source-touching subset** (`classifier-metrics.json`); validation used a single pipeline-author labeller and approximately equal repository sampling (`RESULTS.md:45–49,143`).

**[INFERENCE]** These precision estimates do not establish the error rate among file-touch-weighted outcomes. Nor does noisy outcome measurement automatically attenuate regression slopes; differential errors can bias either way.

E1b has **60 areas across five repositories**, **35 for PR outcomes**. O1 level gain is **−0.44%, CI [−2.72%, 1.42%]**; O4c level gain **9.05%, CI [−2.89%, 18.04%]** (`model-results.json`, respective `.area` records). **[INFERENCE]** “Everything is underpowered for 5%” is too sweeping: O1's reported conditional interval excludes that gain, while O4c remains compatible with substantial gains. Neither interval establishes equivalence of actual effort. Exploratory file-level clone/dispersion success is not confirmatory replication (`.file.o5a.ablations.clone_pct`).

## Better alternatives and evidence needed

**Preferred observational redesign:** task/PR-level prospective panel with workflow-normalized changes, pre-change scope/task category, stable-population structural metrics, and sampled active-work time separated from waiting. Measure H2 through forecast errors or conditional dispersion; adjudicate corrective follow-ups and distinguish statically explained coupling. Validate future periods within repositories separately from unseen repositories. Use hierarchical effects and dependence-aware resampling that repeats the complete fit/tuning pipeline.

**Viable causal alternative:** randomized matched tasks on behavior-equivalent healthy/degraded snapshots, fresh sessions, blinded outcome assessment, and repeated seeds. Measure active time, completion, corrective work, and conditional variance. Include human/agent interaction contrasts before claiming agent magnification. **[INFERENCE]** This improves identification but risks artificial degradation and limited external validity; real-repository panel evidence remains complementary.

**[INFERENCE: illustrative planning, not measured power]** For independent paired tasks, two-sided 5% significance, 80% power, and standardized paired effect 0.3, the normal approximation gives `(1.96+0.84)²/0.3² ≈ 88` pairs before heterogeneity/multiplicity. Repeated seeds do not replace independent tasks. Panel sample requirements need simulation using observed repository/task dependence and the actual predictive decision rule; variance and rare-correction outcomes need separate calculations.

## Scout §8 implications

Keep descriptive/no-gating safeguards and raw-event retention. In `docs/scout-design.md:234–242`, replace “effort measured directly (PR hours, review rounds)” with coordination/latency proxies; do not equate “interest” with identified excess debt cost. Replace the **300 scope-windows / 15 repositories** heuristic with cluster-aware power simulation. Do not automatically promote event studies to causal evidence, or dismiss sign-stable coefficients as numerical artifacts solely because ridge shrinks them.

**Permissible conclusion:** the implemented pipelines did not meet their preregistered support criteria on the selected proxy outcomes. Prediction usefulness, causal debt effects, and product gating remain separate unresolved decisions.

Read-only review; no model reruns or repository modifications.