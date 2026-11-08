# X1: council experiment review, round 1 (preregistration and synthetic validation)

Date: 2026-11-08. Reviewer: innovation council. The lead wrote the synthesis. Two read-only peers ran in parallel: `innovation-challenger` (X1Challenger) and `experiment-peer` (X1Peer). Governance: `CLAUDE.md` experiment rules 1–7, design rules, ledger rules and General "Jev labelling". Scope: X1 PREREGISTRATION revision 1 (DRAFT), `README.md`, `reviews/README.md` and `validation/{ranking.py,self_test.py,synthetic-results.json}`. These were checked against `docs/scout-design.html` (REQ-41, REQ-97, REQ-98, REQ-100–102, REQ-136–138, REQ-169, REQ-182, REQ-183), ADR-010 (`docs/adrs/2026-11-08-d1-jev-minimums/ADR.md`, Conditions, D1-R), the product code (`internal/engine/ranking.go`, `internal/decisions/{rank,audit,triage,run}.go`, `internal/policy/consent.go`) and the X2d precedent.

Nothing was run: no shell, no model call, no validation, no tests. Every finding comes from reading files. `codegraph` was not used because it is not installed (draft README); the codegraph reference documents were read first. Paths are relative to `evaluation/x1-triage-ranking/` unless stated. Preregistration line numbers (P l.N) refer to revision 1 as reviewed.

Reviewed files (SHA-256). Main computed these digests; this read-only round did not recompute them.

| File | SHA-256 |
|---|---|
| `PREREGISTRATION.md` | `5056d51ca5755e85692acac7ea180b6459252d4aec0c75434029a3bb26f61a1f` |
| `README.md` | `7ab969226d18c856a9c0f2b8187eb230b0e5c38673e5d8e6c8349ee6e673a022` |
| `reviews/README.md` | `7ef2703cd0bbcf2268d2d5c812c1f3d8004b8752220b0031baea8d87cd70d1a2` |
| `validation/ranking.py` | `ec2c8460e3cef3b0bfce52140485cd15a8c3493cc16177408b4f90dfc8ab1793` |
| `validation/self_test.py` | `0893d29f4c103aff6bea19c6784574b3e0be8fdba9097d0cd6c734c3dc8d1d22` |
| `validation/synthetic-results.json` | `fbc4cc0de2c948574a76216cffd12cb33b488f0fedd3c3519de4ddf0b5a82277` |

**Verdict: 27 issues open (1 blocking, 13 major, 13 minor). No sign-off.** No stage beyond drafting and synthetic tests is authorised.

Severity scale:
- **blocking**: can produce a wrong or unreachable verdict, or breaks rules 1–7 or REQ-169.
- **major**: misleads the result, leaves a preregistered quantity undefined, or makes the study unrepeatable or likely not testable.
- **minor**: text, consistency or provenance. Minor issues must still be fixed before sign-off (rule 1).

---

## 1. Recommendation

**Revise to revision 2 and resubmit for round 2. Do not sign off.** The core of the protocol is sound and should be kept:
- product `drawAudit` sampling with stored π;
- the repository split;
- the REQ-169 rule: point gain ≥ .10, lower bound > 0, critical veto;
- staged authority;
- dev-only D1-R.

Three classes of defect must be fixed first:
1. **Decidability.** As written, the safety condition makes "support" practically unreachable (X1-R1-01), and the positive-count gate conflicts with the preregistered operating-characteristics (OC) grid at the minimum supply and at k = 20 % (X1-R1-03, X1-R1-04).
2. **Inference target.** The primary interval conditions on 60 fixed commits, so a few bulk runs can drive a default-on decision (X1-R1-05). The kernel's tests also cannot detect a broken 1/π weighting or a broken missing-label interval (X1-R1-09, X1-R1-10).
3. **Fidelity to the deployed treatment and to ADR-010.** These cover the cost caps, fallback runs without audit rows, the incomplete examined-data list and the underspecified D1-R (X1-R1-08, X1-R1-11, X1-R1-13, X1-R1-14).

The required fixes below are concrete. Where an owner choice would otherwise be needed, the council's recommended option is stated as the fix, per the standing owner direction.

## 2. Alternatives considered

| Option | Description | Tradeoff | Council position |
|---|---|---|---|
| A. Revision 1 as drafted | AuditFraction .50, conditional design-based multiplier interval, k = 20 % | Lowest label cost. Support is nearly unreachable (X1-R1-01). The ≥ 60-positive gate fails at the minimum N and in the prevalence .10 grid cells (X1-R1-03). The .20-gain grid cells are infeasible at prevalence ≥ .40 (X1-R1-04). The interval ignores run-level variation (X1-R1-05). | Rejected |
| B. **AuditFraction 1.0 through product `drawAudit`, run-cluster primary interval, k = 30 %, decidable safety** | Keeps REQ-169's "probability-sampled audit units" (π saturates at 1 when strata are balanced). The bottom half is already censused for safety, so the extra labels fall only in the top half: about +25–33 % over revision 1's union, about 0.83–1.0 N in total. | Higher label throughput. The interval reflects heterogeneity across commits. Positives are about doubled, so the gate is reachable. | **Recommended** |
| C. Full-frame labelling outside the product sampler, census estimator (challenger alternative) | Exact finite-frame recall, no HT weighting | Similar cost to B. It departs from REQ-169's probability-sampled audit units and REQ-182's exercised product sampler. With a conditional interval, LB > 0 would become degenerate (zero width). | Not adopted. B achieves the same label coverage inside the product sampler. |
| D. Keep .50 and lower the positive gate to ~30 | Cheapest fix for the gate conflict | Coverage of a ratio estimator at ~30 positives per repository is weak. It does not address X1-R1-04 or X1-R1-05. | Rejected |
| E. Precision@k instead of recall@k | Avoids the prevalence ceiling | REQ-169 names recall@k gain, so this would need a design change (ADR) | Out of scope. The ceiling is handled by choosing k (X1-R1-04). |

## 3. Rule checks

- **Rule 1 / 2:** This is the first round. It is recorded here with digests, and iteration is required.
- **Rule 3:** No predictor–outcome association was computed. `self_test.py` builds only synthetic frames (`ranking.py` `synthetic()`: "All labels are invented"). Neither file has a real-data or model interface. The drafting did not run any examined-data comparison (P §2, README "Do not run legacy `compare_e2.py`").
- **Rule 4:** The kernel recovers a planted .20 effect (support, interval [0.173, 0.280]) and reports a planted null as not supported (`synthetic-results.json` `planted_effect` and `planted_null`). This is a necessary first step but not sufficient: the planted frames have uniform π = .5 (X1-R1-09). Product-backed full-pipeline rule-4 validation is correctly deferred to the full-protocol authority (P §8, §11.3; `synthetic-results.json` `"production_rule4_complete": false`). The council accepts that staging.
- **Model use and secrets:** The draft makes no model call. Labeller CLIs are `claude -p` and `codex exec`. Jev uses the Infisical-injected key only (P §5, §10).
- **Design rules:** The draft makes no product, default or ADR change (README). Ledger compliance is incomplete (X1-R1-27).

## 4. Issues

### X1-R1-01: blocking. The safety condition makes "support" practically unreachable, and an incomplete census can mask a clear null

- **Location:** P §4 "Safety census" (l.54–56); §7 rules 2–3 (l.88–89); `ranking.py` `analyse()` (`safety_unknown`, ~l.153); README planning-table caveat.
- **Problem:**
  - Safety counts as unmeasured if *any* below-floor issue has no resolved label, or if any true issue has indeterminate criticality.
  - The census covers about N/2 units per repository: at least 1,500 cases at N ≥ 1,000 per repository (P l.35).
  - Both label questions allow `cannot_tell` (P l.60).
  - Even a 2 % unresolved rate gives P(no unresolved case among ~1,500) ≈ 0.98^1500 ≈ 10⁻¹³ [INFERENCE: arithmetic].
  - Rule 3 (incomplete safety census → inconclusive) runs before rule 4 (best-case upper95 < .10 → not supported), so a clear null can be reported as "inconclusive".
  - The planning simulation assumes no label failures and a complete census (README).
- **Evidence:**
  - Legacy canonical unresolved: 1 of 52 (`evaluation/jev-triage-labels-v1/labels.jsonl:154`, A and B agree `cannot_tell`, human-confirmed at `:210`).
  - Labeller B said `cannot_tell` on two further cases (`:8`, `:76`).
  - P l.56: "A true issue with indeterminate criticality, or an unlabelled below-floor issue, makes safety unmeasured."
- **Required fix:**
  - (a) **The GUIDE asks two independent questions for every labelled case.** One is truth ∈ {true, false, cannot_tell}. The other is the critical mechanism ∈ {yes, no, cannot_tell}. The definition already requires a *visible* reachable mechanism, so "no critical mechanism visible in the bounded evidence" is `no`. `cannot_tell` is reserved for a visible mechanism whose reach or guard is ambiguous.
  - (b) **A below-floor case is classified as follows:**
    - *violating* if canonical truth = true and (product severity critical or canonical critical = yes);
    - *cleared* if canonical truth = false, or if canonical critical = no and product severity ≠ critical;
    - *unresolved* otherwise.
    Safety is unmeasured only if some case is unresolved.
  - (c) **Reorder §7:**
    1. integrity failure → inconclusive;
    2. any violating case → not supported;
    3. retrieval supply, quality, missingness, denominator or interval inadequacy → inconclusive;
    4. best-case upper95 < .10 → not supported;
    5. safety unmeasured → inconclusive;
    6. worst-case point ≥ .10 and lower95 > 0 → support;
    7. otherwise inconclusive.
  - (d) Implement (a)–(c) in the kernel.
  - (e) Add to the planning and implementation OC grids:
    - below-floor unresolved rates {0, .005, .02, .05};
    - critical-true share of positives {0, .01, .05}, with criticals ranked by the same synthetic score model as other positives.
    Report P(support), P(veto) and P(safety unmeasured).

### X1-R1-02: major. Omitted criticality silently means "not critical" in the kernel

- **Location:** `ranking.py` `Unit.critical: bool | None = False` (~l.23), `validate()`, `analyse()` (~l.152–153); P l.56.
- **Problem:** A true below-floor unit without a criticality result is treated as definitively non-critical. `validate()` does not check criticality's type or provenance.
- **Required fix:**
  - Default `critical` to `None` and represent truth and criticality as three-state values.
  - `validate()` refuses malformed values.
  - Synthetic safe cases must supply explicit `False`/`no`.
  - Add fixtures for: missing truth; missing criticality; malformed criticality; a cleared case (truth unresolved, critical = no); a violation present *only* among census-only cases, not among audit draws.

### X1-R1-03: major. The ≥ 60-positive gate conflicts with the minimum supply and the preregistered OC grid

- **Location:** P §5 l.68; §3 l.35; §8 l.98–104; `ranking.py` `synthetic()`; `self_test.py` planning block.
- **Problem:**
  - The planning frames are all annotated and none is critical, so π = .5·50/2/25 = .5 everywhere.
  - Product `Assign` adds a third stratum whenever any unit is unannotated (`internal/decisions/audit.go` `Assign`; π = fraction·n/strata/size). That lowers top and bottom π to about 1/3.
  - At N = 1,000 and prevalence .10, expected audit positives are about 50 at π = .5 and about 33 with three strata, both below the gate of 60.
  - P l.104 nevertheless requires .20-gain support MC lower ≥ .80 "in every scenario", including prevalence .10, so X1 would be "not testable" at the supply floor.
- **Required fix (recommended option B):**
  - Set **`AuditFraction = 1.0`** through product `drawAudit`. π then saturates at 1 for balanced strata and stays at least ~.67 with an unannotated stratum. The bottom half is already censused, so the extra labels fall only in the top half.
  - Keep the ≥ 60 gate.
  - Rebuild the kernel planning simulation with:
    - unannotated share {0, .20};
    - severity-critical units at 2 % of the frame;
    - prevalence {.10, .20, .40, .55} (.55 approximates the examined set, 28/51, R-063; not a probability sample).
  - Add an analytical check: expected audit positives ≥ 60 at the locked inventory N.
  - State that N ≥ 1,000 is a supply floor, not a powered size.
  - Update the README "Audit fraction" delegated decision and the label-budget figures.

### X1-R1-04: major. At k = 20 %, the 10-point recall gain is bounded by prevalence and the OC grid contains infeasible cells

- **Location:** P §4 l.46 (k), §8 l.104 (grid: baseline recall .30, gains up to .20, prevalence up to .40); README "Review budget".
- **Problem:**
  - Recall@k cannot exceed k/T: .2/p at k = 20 %.
  - At prevalence .40, baseline .30 + gain .20 = .50 requires a perfect ranker.
  - At the examined prevalence (~.55), recall@k ≤ .36, so a ≥ .10 gain leaves almost no room.
  - The choice of k therefore decides whether HD9 is testable, and the draft gives no feasibility rationale.
- **Required fix:**
  - Set **k = max(1, ceil(0.30·N_run))**. The recall ceiling is then ≥ .545 at prevalence .55, and every grid cell (baseline .30 + gain ≤ .20, prevalence ≤ .55) is feasible.
  - The OC generator must refuse infeasible cells rather than silently capping recall.
  - Add the prevalence ceiling to §12.
  - Record k as an amended delegated decision. k stays distinct from the audit floor ceil(N/2).

### X1-R1-05: major. The primary interval conditions on 60 fixed commits, and a few bulk runs can determine a default-on decision

- **Location:** P §6 l.74–79 ("conditional on fixed rankings and finite frame"); §3 l.35–37; §12.
- **Problem:**
  - D_r pools HT totals over runs of 20 to 2,000 issues, so one bulk run can dominate.
  - The multiplier interval reflects only which audit units were drawn, not commit-to-commit variation.
  - REQ-169's decision ("held-out recall@k gain") concerns future runs. REQ-41 explicitly allows a "clustered bootstrap".
  - With X1-R1-03, π ≈ 1 makes the conditional interval nearly zero-width, so LB > 0 would become trivial.
  - Competing hypothesis (peer): the gain comes from a few bulk runs or one rule family.
- **Required fix:**
  - **Primary interval:** an ultimate-cluster bootstrap of runs within each repository. For each replicate, resample that repository's runs with replacement, recompute D_r from the resampled runs' stored-π HT totals, and average D_r with equal repository weights. Use B = 1,999, seed 169, canonical order, the stated percentile endpoints, and clip to [−1, 1].
  - The current conditional multiplier interval is reported as a secondary, descriptive line.
  - Add descriptive sensitivity lines outside the verdict:
    - equal-run-weighted gain per repository;
    - per-run D_run against N_run;
    - leave-largest-run-out gain;
    - gain by tool and rule family;
    - share of baseline top-k slots decided by the issue-ID tie-break.
  - The implementation OC adds a between-run heterogeneity scenario (run-level gain SD {0, .10}) alongside run-concentrated placement.

### X1-R1-06: major. Run eligibility makes X1 fragile and selects bulk-change commits

- **Location:** P §3 l.35 ("Any selected run above 2,000 … makes X1 not testable"); §12.
- **Problem:**
  - A single qualifying run above 2,000 issues stops the study. TypeScript baseline-regeneration commits plausibly exceed that [INFERENCE].
  - "First 20 runs with ≥ 20 retained issues" selects atypical bulk commits (vendoring, regenerated baselines, mass refactors).
  - Prometheus's first-parent history consists of merge commits, unlike squash-merge histories [INFERENCE].
  - None of this is in §12.
- **Required fix:**
  - Preregister a model-free eligibility rule fixed now: a candidate run is eligible iff **20 ≤ N_run ≤ 2,000** after exposure exclusion and first-occurrence removal. Runs above 2,000 are skipped as ineligible, counted and disclosed. Selection is size-based and occurs before any Jev call or label, so it is outcome-blind.
  - Add the bulk-commit selection effect and merge-commit topology to §12.
  - Report the size, path-role and rule mix of selected runs against all candidate runs at inventory.

### X1-R1-07: major. Population and session ceilings disagree arithmetically, and label throughput is unchecked

- **Location:** P §10 l.118–120.
- **Problem:**
  - "60,000 test units (3×20×2,000)": the product is 120,000.
  - The 180,900-session reservation is derived from the wrong figure. Covering 120,000 would need 360,900 sessions before retries.
  - 180,900 sessions at concurrency ≤ 4 in the ~36-day label envelope allow a mean of about 69–86 s per session, against a 900 s timeout.
- **Required fix:**
  - Keep the 60,000-unit and 180,900-session ceilings. Make an aggregate locked frame ≤ 60,000 an explicit **pre-model inventory adequacy gate**, with no truncation or replacement, and correct the arithmetic text.
  - Run development labelling (≤ 300 units) first within the full-protocol stage. Its measured p90 session time projects test labelling time, and the projection must fit the remaining envelope before any test-repository Jev call. Otherwise X1 is not testable before models.

### X1-R1-08: major. The study cost cap is ten times the product default, so it changes the treatment being evaluated

- **Location:** P §10 l.118 ("production cap 2,000 Jev calls / USD5"); README "Handling failed/capped Jev" ("Estimates deployed ranking"); `internal/policy/consent.go:25,30` (`DefaultMaxPerRun = 2000`, `DefaultMaxCostUSD = 0.50`); `internal/decisions/run.go:29–30,185–190`.
- **Problem:**
  - Capped units fall to the unannotated tier (`rank.go` `Rank`).
  - At the measured ≈ $0.000589 per request (R-052; triage price assumed similar [INFERENCE]), 2,000 calls cost about $1.18, so the product default binds at roughly 850 calls. The study's $5 cap evaluates a configuration the default does not ship.
  - Product caps are pre-call checks; a run can overshoot by the call in flight (`run.go:29–30`). P l.120 calls them "hard ceilings".
  - The study-wide "unknown-cost latch" has no mapping to a verdict.
- **Required fix:**
  - Use the **product default policy caps** (2,000 calls / $0.50 per run, pinned by policy digest), because the decision concerns the default.
  - Report cap counts and positions per run.
  - Describe per-run and study ceilings as pre-call stopping thresholds with disclosed in-flight overshoot.
  - Keep a persistent study-wide latch across dev and test calls. A run stopped by the study latch keeps the product fallback tier for its unasked units if its frame is complete; otherwise its repository is inadequate (inconclusive).
  - Rescale the study USD ceiling accordingly (60 × $0.50 plus dev).

### X1-R1-09: major. The synthetic tests cannot detect wrong inverse-probability weighting

- **Location:** `self_test.py` (planted, census and six-unit fixtures); `ranking.py` `synthetic()` (`True, False, int(issue < 10)`).
- **Problem:** Every planted fixture has uniform π (.5 or 1), so 1/π cancels in the recall ratios. The unequal-π fixture checks assignment only, so an unweighted estimator passes every check.
- **Required fix:**
  - Add independently hand-calculated unequal-π fixtures covering:
    - HT positive and retrieved totals, paired gain, interval and missing-label bounds;
    - odd-sized runs, saturation, mixed annotated and unannotated tiers, and certainty units.
  - Add a planted effect and a planted null whose positives concentrate in a low-π stratum, so that an unweighted estimator is biased.
  - Add mutation checks that must fail when 1/π is omitted, replaced by nominal .50, or recomputed from selected rows.
  - The product-backed harness additionally exercises first-draw persistence and resume.

### X1-R1-10: major. The missing-label interval is not shown to bound all compatible completions

- **Location:** P §6 l.81; `ranking.py` `extreme()` (selects by `gain` only), `analyse()` (intervals at those fits); `self_test.py` ("missing labels widen gain range").
- **Problem:**
  - The eight d-group vertices give sharp *point* extrema; the ratio is linear-fractional (sound).
  - The interval is computed only at the point-minimising and point-maximising completions. Another completion can have a less extreme point but a wider interval [INFERENCE].
  - The only test checks that the point range widens.
- **Required fix:**
  - Compute the primary interval for all eight vertex completions.
  - LB = minimum of the lower endpoints; UB = maximum of the upper endpoints. The support point remains the minimising vertex's point.
  - Call it a "vertex envelope", not a sharp confidence region.
  - Validate by exhaustive enumeration of completions on small unequal-π synthetic frames, and record any excess as a limitation.

### X1-R1-11: major. Runs where `RankingOf` falls back entirely get no audit rows

- **Location:** P §4 l.48–50; `internal/engine/ranking.go` (`!ok` → `baselineRanking`; audit only under `if ok && in.Trigger == domain.TriggerNightly`).
- **Problem:** Eligibility, config-validation or threshold-resolution failure gives a complete deterministic ranking but no `drawAudit` call. HT is then impossible for that run, and the draft does not say how such a run is handled.
- **Required fix:**
  - For such a run, call product `decisions.DrawAudit` on the frozen fallback ranking with the same sampler, fraction and isolated store, and record the reason.
  - Never invent π or omit the run.
  - Add disabled, key-missing, threshold-invalid and config-invalid fixtures alongside per-unit transport and cap failures.

### X1-R1-12: major. Common-mode labelling threatens construct validity

- **Location:** P §1 l.7 (HD9 wording), §5 l.60–62, §12.
- **Problem:**
  - Labellers answer Jev's own question on Jev's own bounded input (`internal/decisions/triage.go` question and `BuildState` bounds).
  - The baseline (severity × path role) never attempts that question.
  - Competing hypothesis: the gain reflects LLM heuristics shared with the labellers on the 8-line/600-byte excerpt, not better retrieval of real defects. §12 only names the threat.
- **Required fix:**
  - Narrow HD9 to "true defects as judged by blind agent labels of the production bounded-state question".
  - Add one preregistered descriptive validity line outside the verdict, from model-free scan continuity already collected at inventory: share of each arm's top-k issues absent from the product scan within the next 200 first-parent commits (or by envelope end), compared by arm with intervals. State that churn makes this proxy noisy.
  - Per-case human review stays excluded (CLAUDE.md, R-039).

### X1-R1-13: major. The examined-data list is incomplete (ADR-010 disclosure condition)

- **Location:** P §2 l.17; README "Repositories/split".
- **Problem:** The draft names only fastapi, axum, nest and webpack as examined.
- **Evidence:** The examined strata also include junit5, eShopOnWeb and caddy (`evaluation/d1-triage-minimums-exploratory/scores.csv:3,6,19`; `tuning.json` strata). ADR-010 Conditions: "Any X1 or D1-R preregistration must disclose them, may use them only for development".
- **Required fix:**
  - List all seven repositories and the E2 source pool as examined and dev-only.
  - Name which of them form the D1-R development pool.
  - Confirm the test repositories (django, TypeScript, prometheus) are disjoint from them; the council found them disjoint.

### X1-R1-14: major. D1-R is underspecified

- **Location:** P §9 l.106–114; README "Threshold/minimum adoption" and "Cost assumptions".
- **Problem:**
  - (a) "Weighted precision/recall" on dev has no defined sampling design or weights.
  - (b) "In every development repository with ≥ 30 labels of each class" is ambiguous at ≤ 300 units over up to seven repositories.
  - (c) The loss units (miss 2, false action 1, escalation .25) enter no formula. REQ-102 requires costs to be fixed *and used*.
  - (d) The transfer check is undefined for unannotated test rows and for "no feasible candidate", and it omits the provisional .30.
  - (e) The draft does not state that ADR-010's "set or replace the threshold" step stays open.
  - (f) It does not state how REQ-101's human-over-agent precedence applies to relabelled examined cases whose human labels answer a different question version.
- **Required fix:**
  - (a) The dev collection recipe is a model-free chronological census per dev repository up to the cap, and estimates are unweighted with Wilson intervals; or define stored π if sampling is used.
  - (b) Feasibility requires pooled point P ≥ .80 and R ≥ .90, and the same in each dev repository that reaches 30 labels per class. Repositories below the floor are reported as insufficient but stay in the pooled estimate. If fewer than two repositories reach the floor, no candidate is locked.
  - (c) Report expected loss per 100 units at each grid threshold using the stated units, as a sensitivity table that never selects the candidate.
  - (d) The transfer check classifies unannotated rows as not-escalated (no score). It reports both the locked candidate (if any) and .30, with HT weights and missing bounds.
  - (e) State that the set-or-replace decision goes to a later owner-reviewed ADR.
  - (f) State that legacy human labels (guide `triage-v1_choice/guide-1`) do not override current-guide `agent_reviewed` labels, because the question versions differ.

### X1-R1-15: minor. The labeller payload shows the baseline's main input

- **Location:** P §5 l.62; `internal/decisions/triage.go` `BuildState` (`Severity: clean(e.Severity, 32)`); `internal/engine/ranking.go` `BaselineVersion()` (severity table).
- **Problem:** "Exactly the bounded production evidence" includes scanner severity. The real-defect criteria do not use it, and the critical veto takes severity from product data.
- **Required fix:**
  - Remove `severity` from the labeller payload and add a fixture asserting it is absent.
  - Record path role and tool signal as unavoidable shared inputs in §12.

### X1-R1-16: minor. Audit draws are not reproducible, the store is unscoped, and the operator is not blinded

- **Location:** P §4 l.48–50; `internal/engine/ranking.go` (`Rand` nil → `crypto/rand`); `audit_draws` key lacks a repository.
- **Required fix:**
  - Inject a preregistered deterministic `RankingInput.Rand`, e.g. a uniform value derived from HMAC-SHA-256(seed `169-x1-audit`, repository‖run_id‖issue_id).
  - Use one isolated store per repository and refuse issue-ID collisions.
  - Give the label driver an opaque case list only; Jev and baseline exports stay sealed until gold freeze.

### X1-R1-17: minor. Jev drift and nondeterminism are unaddressed

- **Location:** P §4, §12.
- **Required fix:**
  - Interleave Jev scoring across repositories (seeded round-robin by run).
  - Pin and record labeller CLI versions.
  - State that seed 169100 orders cases globally across repositories.
  - Record "one Jev realisation" as a stated condition of the verdict. An optional seeded 5 % re-score is descriptive only.

### X1-R1-18: minor. The adjudicator reading and vendor siding are unstated
- **Location:** P §5 l.62 (C = `claude -p` sonnet; "C may see A/B reasons"); REQ-100 ("never showing … prior verdict").
- **Required fix:**
  - State that A and B's reasons are adjudication inputs, not prior verdicts (consistent with the legacy protocol, `jev-triage-labels-v1/labels.jsonl:9`).
  - Report adjudicator siding by vendor. A siding rate with A above .80 triggers council review before unblinding.

### X1-R1-19: minor. GUIDE freeze and provenance are underspecified

- **Location:** P §5 l.60; §11.3.
- **Required fix:**
  - Put the GUIDE SHA-256 in every session and label row (precedent: `guide_version` in `jev-triage-labels-v1/labels.jsonl`).
  - Keep the GUIDE to one page (REQ-100).
  - Bar the GUIDE author from E2/D1 scores and Jev-versus-label disagreements, and record the author's exposure.

### X1-R1-20: minor. The κ gate's case set is undefined

- **Location:** P §5 l.68.
- **Required fix:**
  - Compute κ on the union of labelled cases per repository, unweighted, over A/B valid sessions.
  - Treat `cannot_tell` as a third category in the confusion table.
  - The true/false κ uses cases where both labellers answered true or false.
  - Also report κ on audit draws only.

### X1-R1-21: minor. Bootstrap ordering contradicts the preregistration

- **Location:** P l.79 ("order repository/run/issue"); `ranking.py` `analyse()` (rows keep input order within a repository).
- **Required fix:**
  - Canonically sort by (repository, run, issue) with byte-wise string comparison before fitting and drawing multipliers.
  - Add an input-permutation invariance fixture.
  - Regenerate `synthetic-results.json`.

### X1-R1-22: minor. Training-data exposure and the meaning of "fresh" are unstated

- **Location:** P §2 l.17, §3 l.31, §12.
- **Required fix:**
  - Record the published training cutoffs of Jev 1.14 and both labeller models against the 2024 window.
  - State in §1 and §12 that "fresh" means unexposed to Scout and inertia-bench data, not to models or researchers (X11 planning shares these envelopes).

### X1-R1-23: minor. The safety veto has no comparator

- **Location:** P §7 rule 2.
- **Required fix:**
  - Report descriptively the number of violating cases (per X1-R1-01) placed below the *baseline's* floor. This needs labels only for census and audit cases, so report it with the overlap disclosed.
  - REQ-169's absolute veto is unchanged.

### X1-R1-24: minor. REQ-138 line definition is ambiguous

- **Location:** P §6 l.83; REQ-138 ("audit-view blind dispositions only and inverse-inclusion weights").
- **Required fix:**
  - State that X1's AUROC and routing-precision lines are study lines on `agent_reviewed` labels.
  - State what the product scorecard line uses after X1: nightly audit-view blind dispositions with annotation visibility false, at stored π, for each adopting repository's own audit period; otherwise `insufficient`.
  - Remove "audit-view annotation visibility" as a condition on agent labels.

### X1-R1-25: minor. The action wording misstates the current state

- **Location:** P §7 l.94.
- **Problem:** No `--rank` flag exists (no match in `internal/` or `cmd/`). REQ-97 launches Jev "unvalidated, X1 pending" when enabled, and REQ-169 says "Jev stays default-on iff…".
- **Required fix:** For both not-supported and inconclusive: "recommend switching to the deterministic default, with a `--rank jev` opt-in to be implemented under a separate signed ADR; route-only is not reopened".

### X1-R1-26: minor. Product sampler parameters after X1 are unstated

- **Location:** `internal/decisions/audit.go:11–17` ("strata, fraction and floor are [ASSUMPTION] until X1's preregistration is signed off"); P §4.
- **Required fix:**
  - State that X1's sign-off fixes the sampler (strata, floor, AuditFraction 1.0) for this study only.
  - Product `ProvisionalFraction = 0.05` and the interim parameters stay [ASSUMPTION] until a signed ADR (REQ-183).

### X1-R1-27: minor. Ledger compliance is incomplete

- **Location:** README ledger handoff; `docs/hypothesis-ledger.md` (no HD9 row; HD2 at l.26 is the prior triage-routing hypothesis); `docs/research-ledger.md` (last ID R-063).
- **Problem:** CLAUDE.md ledger rules 2–3 require recording a hypothesis when it is proposed, and preregistrations to cite ledger IDs.
- **Required fix:**
  - Main appends HD9 as `proposed`, citing HD2.
  - Main assigns R-IDs (from R-064) to the owner-direction, product-contract, X2d and X11a sources and to R-052 (cost), and cites them in revision 2 before the sign-off round.

---

## 5. Delegated decisions (README table)

| Decision | Council finding |
|---|---|
| Include D1-R as dev-only secondary | **Accepted**, consistent with REQ-98 and ADR-010 ("does not change X1's ordinal decision rule"). Amended by X1-R1-13 and X1-R1-14. |
| Retain ≥ .10 point gain with LB > 0 | **Accepted**. REQ-98: defaults "substitute only for precision/recall point minimums … not X1 gain". REQ-169 allows confirmation. |
| Repository holdout (django/TypeScript/prometheus) | **Accepted**. Amended by X1-R1-13 and X1-R1-22. |
| k = ceil(.20 N), floor = ceil(N/2) | Floor **accepted**; it matches `audit.go` `floor := (n + 1) / 2`. k **amended** to ceil(.30 N) (X1-R1-04). |
| AuditFraction .50, conditional interval, equal repository weights | Equal weights **accepted**. Fraction **amended** to 1.0 (X1-R1-03). Interval **amended** to a run-cluster primary (X1-R1-05) and a vertex envelope (X1-R1-10). |
| Blind census of every below-floor issue | **Accepted** in principle. Amended for decidability (X1-R1-01, X1-R1-02), with a comparator added (X1-R1-23). |
| First 2,000 commits, first 20 qualifying runs, no replacement | **Amended**: eligibility 20 ≤ N_run ≤ 2,000 and aggregate ≤ 60,000 gate (X1-R1-06, X1-R1-07). |
| Failed/capped units → product unannotated tier | **Accepted**. Amended for whole-run fallback (X1-R1-11) and product default caps (X1-R1-08). |
| Sharp missing-label bounds, ≤ 5 % gate, complete safety | Bounds and gate **accepted**. Amended by X1-R1-10 and X1-R1-01. |
| Staged authorities | **Accepted**. |
| Dev grid/F2, one frozen transfer report, later ADR | **Accepted**. Amended by X1-R1-14. |
| Loss units miss 2 / false action 1 / escalation .25 | **Amended**: they must enter a stated sensitivity formula (X1-R1-14c). |

## 6. Checked and sound

- **Sampler mirror:** `ranking.py` `assign()` matches `audit.go` `Assign`: floor (n+1)/2; certainty = severity critical and position > floor; stratum precedence; π = min(1, fraction·n/H/N_h) with H counting non-empty non-certainty strata. `DrawAudit` makes an independent Bernoulli draw per unit, stores both drawn and undrawn rows, and never redraws (`audit.go` `DrawAudit`).
- **HT and linearisation:** T_r, C_Ar and the influence √(1−π)·Y(d−D)/(πT) are correct for Poisson sampling. Certainty rows have zero variance (`ranking.py` `fit()`; census fixture exact .20 with a zero-width interval).
- **Ranking specification:** P §4 matches `rank.go` `Rank`: P(true) descending, ties by baseline then issue ID, and a separate unannotated tier. `ranking.py` consumes rank permutations rather than reimplementing ranking, so product parity is deferred to the full-protocol authority, as the draft says.
- **Missing-label point vertices:** the bounds are sharp, because the ratio is linear-fractional and each unit's direction depends only on d (with sign preserved for d = 0).
- **Labelling protocol:** blind `claude -p` and `codex exec` passes, a reason-recording adjudicator, no tools, `agent`/`agent_reviewed` provenance, append-only, frozen before unblinding, no per-case user review (P §5). This matches CLAUDE.md Jev labelling, REQ-100 and REQ-101, subject to X1-R1-15 and X1-R1-18–20.
- **Quarantine:** the 52 cases, E2 `v2_noul` and D1 exploratory data are dev-only, and legacy labels do not count toward the D1-R floor (P §2), as ADR-010 requires. The test repositories are disjoint from all seven examined repositories.
- **Verdict mapping:** support uses point ≥ .10 and LB > 0 (not LB ≥ .10), and the boundary power (~.47 at gain .10) is disclosed (`synthetic-results.json` `0.1`: 189/400). D1-R and descriptive lines cannot rescue the verdict. One confirmatory use. No signoff.json is fabricated.
- **Synthetic artifact:** the 17 checks, seeds and scenario structure in `synthetic-results.json` are consistent with `self_test.py`. Exact numerical reproduction was not exercised.

## 7. Evidence index

- REQ-169: "Jev stays default-on iff held-out recall@k gain ≥ 10 points with the interval's lower bound > 0, and no critical true issue lies below the audit floor; … the preregistration confirms or replaces it" (`docs/scout-design.html:686`).
- REQ-98: defaults "not X1 gain"; D1-R "without changing X1's ordinal decision rule" (`:580`).
- REQ-41: "design-based 95 % intervals (stratified or clustered bootstrap)" (`:682`).
- REQ-138: "audit-view blind dispositions only and inverse-inclusion weights" (`:685`).
- ADR-010 Conditions: examined data and the dev-only rule. D1-R: "re-check 0.80/0.90 feasibility per stratum, set or replace the threshold" (`docs/adrs/2026-11-08-d1-jev-minimums/ADR.md`).
- Product sampler: `internal/decisions/audit.go` (`SamplerVersion`, `ProvisionalFraction = 0.05`, `Assign`, `DrawAudit`). Audit call gated on `ok && TriggerNightly`; `crypto/rand` default (`internal/engine/ranking.go`).
- Caps: `internal/policy/consent.go:25,30`. Pre-call cost check: `internal/decisions/run.go:29–30`. Measured price ≈ $0.000589 per request: R-052 (`docs/research-ledger.md:60`).
- Examined strata: `evaluation/d1-triage-minimums-exploratory/scores.csv`. Legacy unresolved label: `evaluation/jev-triage-labels-v1/labels.jsonl:154,210`.

## 8. Risks

- **Throughput:** about 0.83–1.0 N labelled cases at up to 60,000 units may exceed CLI subscription throughput. The X1-R1-07 go/no-go stops before any test Jev call, so the risk costs time, not validity. Reversible.
- **Supply:** ≥ 20 runs with 20 ≤ N ≤ 2,000 and N ≥ 1,000 per repository may not exist in the envelopes [INFERENCE]. The inventory stage finds out without models. There is no replacement repository, so X1 would be not testable, by design.
- **Safety veto:** even when decidable, REQ-169's absolute veto may fire for a good ranker, because P(true) does not target criticality. The X1-R1-01e OC makes this probability visible before the run. The council does not alter REQ-169.
- **Construct:** agent labels on bounded state cannot establish real-world defect truth. X1-R1-12 narrows the claim and adds a model-free proxy, and the residual risk is disclosed.
- **Run-cluster coverage** with 20 unequal clusters may undercover. The implementation OC coverage check (target .95; MC lower < .90 triggers review) covers this.
- **Reversibility:** every fix is pre-data. No product, default or ADR is changed by this round.

## 9. Where the peers changed the council's recommendation

- **Challenger:**
  - Added X1-R1-11 (fallback runs without audit rows), which the lead missed.
  - Turned the lead's interval concern into the concrete vertex-endpoint fix (X1-R1-10).
  - Supplied the kernel `critical=False` default (X1-R1-02) and the in-flight overshoot detail (X1-R1-08).
  - Independently confirmed the 60,000-versus-120,000 arithmetic (X1-R1-07). The lead adopted its "keep 60,000 as a pre-model gate" fix over correcting the ceiling upward.
  - Its full-census alternative informed option B, which reaches the same coverage inside the product sampler.
  - It judged the conditional interval sound for the stated estimand. **The lead overrides this** (X1-R1-05) because of REQ-169's "held-out" decision, REQ-41's clustered option, the run-dominance hypothesis, and the near-degeneracy introduced by X1-R1-03.
- **Experiment peer:**
  - Added the §7 reordering, so an incomplete census blocks only support (X1-R1-01c).
  - Added severity leakage (X1-R1-15), the common-mode competing hypothesis and its proxy (X1-R1-12), adjudicator siding (X1-R1-18), GUIDE provenance (X1-R1-19), drift (X1-R1-17), the `--rank` wording (X1-R1-25) and the HD2 link (X1-R1-27).
  - Its R-052 price evidence moved the cost fix from "disclose the override" to "use the product defaults" (X1-R1-08).
  - Its prevalence .55 observation led the lead to the recall-ceiling finding (X1-R1-04).
  - It rated run-size skipping minor; the lead raised it to major and merged it with the bulk-commit selection threat (X1-R1-06).
- **Lead-only findings:** X1-R1-04 (k ceiling), X1-R1-05 (run-cluster primary), X1-R1-14a (undefined dev weights), X1-R1-24 (REQ-138; also raised by the peer), X1-R1-26 (product sampler parameters).

## 10. Execution outline (for revision 2; one integration owner)

1. **§4/§7 and the kernel:** three-state truth and criticality; cleared/violating/unresolved classification; reordered precedence; `critical` defaults to `None` with validation (X1-R1-01, X1-R1-02). *Test:* new fixtures, including a census-only violation and a cleared-unresolved case.
2. **§3/§4/§10:** eligibility 20 ≤ N_run ≤ 2,000; aggregate ≤ 60,000 gate; corrected arithmetic; dev-first throughput go/no-go; product default caps; pre-call ceiling wording; latch mapping (X1-R1-06, X1-R1-07, X1-R1-08).
3. **§4/§8:** AuditFraction 1.0; k = ceil(.30 N); fallback-run `DrawAudit`; seeded HMAC `Rand`; per-repository store (X1-R1-03, X1-R1-04, X1-R1-11, X1-R1-16).
4. **§6 and the kernel:** run-cluster primary interval; conditional interval secondary; vertex-envelope endpoints; canonical sort; run-sensitivity and follow-through proxy descriptive lines (X1-R1-05, X1-R1-10, X1-R1-12, X1-R1-21).
5. **Kernel validation:** unequal-π fixtures, mutation checks and exhaustive-completion checks. Rebuild planning with unannotated share {0, .20}, 2 % severity-critical units, prevalence {.10, .20, .40, .55}, feasible cells only, unresolved rates {0, .005, .02, .05} and critical-true share {0, .01, .05}. *Test:* .20-gain support MC lower ≥ .80 and null false-support MC upper ≤ .05 in every feasible cell under the run-cluster interval. Regenerate `synthetic-results.json` with create-once output (X1-R1-09).
6. **§2/§9/§5/§12:** seven examined repositories and the D1-R pool; D1-R fixes (a)–(f); labeller payload without severity; adjudicator reading and siding trigger; GUIDE provenance; κ scope; training cutoffs and the meaning of "fresh"; drift; bulk-commit and prevalence-ceiling threats (X1-R1-13–15, X1-R1-17–20, X1-R1-22).
7. **§6/§7/§4 text:** REQ-138 statement, action wording, sampler-parameter scope, safety comparator (X1-R1-23–26).
8. **README:** update the delegated-decisions table per §5 above. Main appends HD9 (`proposed`, cite HD2) and assigns R-IDs from R-064, cited in revision 2 (X1-R1-27).
9. Run the self-test once from the pinned toolchain, record the new digests of all six files, and submit council round 2 (`reviews/round-2.md`). Rule 5 applies to every byte after a zero-issue round.

**Sign-off statement:** none. Round 1 closes with **27 issues open**. No preparation, inventory, Jev, labelling or analysis stage is authorised.
