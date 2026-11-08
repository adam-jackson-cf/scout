# X1: council experiment review, round 2 (revision 2: preregistration, synthetic validation and ledger rows)

Date: 2026-11-08. Reviewer: innovation council. The lead wrote the synthesis. Two read-only peers ran in parallel: `innovation-challenger` (X1Council2.ChallengerX1R2) and `experiment-peer` (X1Council2.PeerX1R2).

**Governance.** `CLAUDE.md` experiment rules 1–7, the ledger rules, General "Jev labelling" and "Model use", and the standing owner delegation R-064. Under R-064, where an owner choice arises, the council's recommended option is stated as the fix.

**Scope.** Revision 2 of:
- `PREREGISTRATION.md`, `README.md` and `reviews/README.md`;
- `reviews/round-1-response.md`;
- `validation/{ranking.py,self_test.py,synthetic-results.json,synthetic-results-rev1.json}`;
- the uncommitted ledger rows: HD9 in `docs/hypothesis-ledger.md` and R-065 to R-068 in `docs/research-ledger.md`.

Each revised item was checked against its round-1 required fix (`reviews/round-1.md`). Product code was read where the protocol depends on it: `internal/decisions/{audit.go,run.go}`, `internal/engine/ranking.go` and `internal/policy/consent.go`.

**What was and was not run.**
- No shell, no model call, no self-test run and no digest recomputation.
- Evidence comes from reading the files, working the hand fixtures by arithmetic, and jq queries over `synthetic-results.json`.
- `codegraph` was not used. Its reference documents were read first, but no executable was available in this read-only session.
- Paths are relative to `evaluation/x1-triage-ranking/` unless stated. "P l.N" means a line of revision 2's `PREREGISTRATION.md`.

**Reviewed files (SHA-256).** Main computed these digests; this round did not recompute them.

| File | SHA-256 |
|---|---|
| `PREREGISTRATION.md` | `68b1a433d758686ac90e947e160294d6e4412d818fd19805ceb98433b78370f1` |
| `README.md` | `d616612a0b36ed3cd643b1c863b872799f7b4bdae7c4feb31455ff13817cd8d5` |
| `reviews/README.md` | `941d2ecd368864413e6b015e7433011e80c12f64ea6043496aedcd19800c805b` |
| `reviews/round-1.md` | `89eab4a5d0f63487da200ab34db28dc0528ed9fcc73745c9d1075272339e172d` |
| `reviews/round-1-response.md` | `d1a4a6f5a13c9715df16161ea6b134a798fa311ecd4617739757fda25f9e242a` |
| `validation/ranking.py` | `909e31e4c2015252b60c64d2d1e193db59fd7fd63b71926ff12c3c2bda0ad3ac` |
| `validation/self_test.py` | `f6e0c524655e4b57d0b115bd6c2daf6349a856f65a63ffc7630e752482c40740` |
| `validation/synthetic-results-rev1.json` | `fbc4cc0de2c948574a76216cffd12cb33b488f0fedd3c3519de4ddf0b5a82277` |
| `validation/synthetic-results.json` | `24549f3ed2fcaecb357bf747ec31058ae115592f8bf63a1319763d2a84faa77d` |
| `docs/hypothesis-ledger.md` | `56949d108723826f4b50f89286895037ab28aac9bd180591656b6dc4f291d581` |
| `docs/research-ledger.md` | `cc935439acf167d5ed367847ce30754263c9506b0208831641d957a027457e49` |

The digest of `synthetic-results-rev1.json` equals the digest round 1 recorded for `validation/synthetic-results.json` (`fbc4cc0d…`). By Main's computation, therefore, the reviewed revision-1 evidence is preserved byte for byte.

**Verdict: 15 issues open (1 blocking, 5 major, 9 minor). No sign-off.** No preparation, inventory, Jev, labelling or analysis stage is authorised (§12).

Severity scale (unchanged from round 1):
- **blocking**: can produce a wrong or unreachable verdict, or breaks rules 1–7 or REQ-169.
- **major**: misleads the result, leaves a preregistered quantity undefined, or makes the study unrepeatable or likely not testable.
- **minor**: text, consistency or provenance. Minor issues must still be fixed before sign-off (rule 1).

---

## 1. Recommendation

**Revise to revision 3 and resubmit for round 3. Do not sign off.** Revision 2 implements round 1 faithfully: 23 of 27 issues are fixed and 4 are partly fixed. Keep the following:
- the decidable safety predicate and the reordered §7;
- product default caps;
- AuditFraction 1.0 through product `DrawAudit` with stored π;
- k = 30 %;
- the vertex envelope;
- the narrowed HD9;
- the safety-clear qualification of the power gate.

Three defects must be repaired before any stage can be authorised.

1. **The primary interval is not validated against what it estimates** (X1-R2-01, blocking).
   - The planning OC holds one population fixed and varies only the audit draws.
   - In half the cells every π = 1, so all 100 trials are the same deterministic outcome.
   - The coverage rule ("MC lower < .90 triggers review") cannot be executed and is too lenient.
   - Recommended fix:
     - take a seeded simple random sample of runs from the inventoried envelope;
     - use a run bootstrap corrected for few clusters (McCarthy–Snowden, m = G − 1);
     - validate it with an outer simulation that redraws runs, against a binding one-sided miss-rate gate.
2. **The validation and OC generators are too favourable to be evidence** (X1-R2-02, X1-R2-03, X1-R2-05):
   - the planted null is structurally zero;
   - the two arms never disagree on a positive in Jev's disfavour;
   - safety failures are fixed counts rather than per-case events;
   - the product's cap order (Jev reorders only the baseline's top M) is not modelled.
3. **Stopping and treatment-delivery rules can turn a study artefact into a decisive "not supported"** (X1-R2-04), and the throughput and clock gates are not executable (X1-R2-06).

## 2. Alternatives considered

| Option | Description | Tradeoff | Council position |
|---|---|---|---|
| A. Keep "first 20 eligible runs"; add an outer superpopulation simulation (peer) | Simulate fresh exchangeable runs per trial; score coverage against the superpopulation gain | Keeps the selection rule. Real chronological runs are consecutive commits and plausibly dependent [INFERENCE], so the exchangeability the bootstrap assumes cannot be checked on real data. | Not adopted alone; its simulation design is reused in B |
| B. **Seeded simple random sample (SRS) of 20 eligible runs per repository from the full 2,000-commit envelope; McCarthy–Snowden run bootstrap (m = G − 1); outer simulation over the locked eligible-run frame** (challenger alternative, refined) | P l.39 already scans every candidate commit, so a probability sample of runs costs nothing extra. The run-stage uncertainty then comes from a real design. Omitting the finite-population correction makes the interval conservative. | The estimand becomes the eligible runs of the three pinned envelopes. Transfer to future commits becomes a stated limitation (it was never established anyway: P l.99, "cannot make purposive histories a random sample"). | **Recommended** |
| C. Keep the first-20 rule and the naive bootstrap; raise trials to 10,000 | Smallest change | More trials narrow the Monte Carlo error but do not fix the target or the few-cluster bias. Planning already shows .90–.93 coverage on audit sampling alone (JSON). | Rejected |
| D. Studentised cluster bootstrap, or cluster-t with 19 df, as an automatic fallback (peer) | A second method used if the primary fails the OC | It adds a choice made after seeing the OC. Rule 5 already allows a council-reviewed amendment before any model call. | Not adopted. Gate failure means not testable pending a new round. |
| E. Report the safety projection only (peer) versus a conservative futility stop | Before test Jev calls, project P(zero unresolved) from dev labels | Inconclusive and not-supported lead to the same action (P l.125). A near-certain inconclusive therefore spends the 45-day envelope for nothing. Reporting alone does not prevent that. | **Futility stop adopted**, using the Wilson lower bound so that it errs towards running (X1-R2-03) |

## 3. Rule checks

- **Rules 1–2.** This round is recorded here with digests. There are open issues, so iteration is required.
- **Rule 3.** No predictor–outcome association was computed. `ranking.py` and `self_test.py` build only synthetic frames and have no real-data or model interface. The response and README report no real-data runs.
- **Rule 4.**
  - The kernel recovers a planted effect (README: gain .1968, interval [.1728, .2195], support).
  - It reports a planted null as not supported. However, that null is structurally degenerate (X1-R2-02).
  - The hand-specified unequal-π null does discriminate the weighting.
  - Product-backed rule-4 validation is correctly deferred (P l.139; JSON `production_rule4_complete: false`).
- **Rule 5.** Nothing is signed yet. Revision 1's evidence is preserved, and the replaced revision-2 evidence was regenerated before review, with disclosure (README).
- **Rule 6.** Reporting duties are preregistered. Veto and safety-unmeasured probabilities lack their required intervals (X1-R2-08).
- **Model use and secrets.** Labellers use `claude -p` and `codex exec`. Jev uses the Infisical-injected key only (P l.85, l.159).
- **Ledgers.** HD9 and R-065 to R-068 are present. R-064 is absent from this branch, and the README offers a conflicting R-064 row (X1-R2-13).

## 4. Verification of round-1 fixes

| Issue | Status | Evidence |
|---|---|---|
| X1-R1-01 (blocking) | **Fixed** | Independent truth/criticality questions; violating/cleared/unresolved predicates (P l.68–71); §7 order (P l.115–121). Kernel `safety()` and `verdict()` match exactly. Fixtures cover a cleared unknown-truth case, missing criticality, a census-only veto and precedence. The OC grid has 4 unresolved rates × 3 critical shares. Residuals: the OC is unrealistic and reachability is not projected (X1-R2-03); veto/unmeasured intervals are missing (X1-R2-08). |
| X1-R1-02 | Fixed | `Unit.critical: bool \| None = None`; `validate()` refuses `bool` truth, string criticality and π ∉ (0, 1]. Safe synthetic rows pass explicit `False`. |
| X1-R1-03 | Fixed | AuditFraction 1.0 (P l.52; `assign(fraction=1.)`); ≥ 60 gate (P l.91); grid with unannotated {0, .20}, 2 % severity-critical and prevalence {.10, .20, .40, .55}; ΣπY ≥ 60 checked in all 48 populations. Supply floor wording at P l.39. Product `RankingInput.AuditFraction` accepts 1.0 (`internal/engine/ranking.go:69,371–374`). |
| X1-R1-04 | Fixed | k = max(1, ceil(.30N)) (P l.50; `Unit.k`). The double 0.3 lies just below 0.3, so `ceil` cannot round up past an exact integer. The generator refuses infeasible cells, and a fixture covers the refusal. The ceiling is stated in §12. |
| X1-R1-05 | **Partly fixed** | The run-cluster primary, conditional secondary and sensitivity lines are present (P l.99, l.105; `cluster_distribution`). The heterogeneity scenario is a fixed population scored against the finite-frame gain, so the primary interval's target is never tested (X1-R2-01). |
| X1-R1-06 | Fixed | 20 ≤ N_run ≤ 2,000; oversized runs skipped and disclosed (P l.39); inventory comparison (P l.41); bulk-commit and merge-topology threats (§12). |
| X1-R1-07 | Fixed | The 120,000 arithmetic maximum and a separate ≤ 60,000 gate (P l.39); dev-first projection gate (P l.159). The projection is not executable as written (X1-R2-06). |
| X1-R1-08 | Fixed as specified | Product caps of 2,000 calls / $0.50 (`consent.go:25,30`); pre-call wording; $31 latch (P l.157). The latch-to-verdict mapping that round 1 mandated is itself flawed (X1-R2-04). |
| X1-R1-09 | **Partly fixed** | Hand totals T = 7, retrieved J = 6 and B = 3, gain 3/7, confirmed by arithmetic. Low-π effect and null discriminate. Gaps: the selected-fraction mutation equals the omitted-weight mutation (π = 7/7 = 1), and the hand intervals come from identical runs (X1-R2-07). The stochastic null is degenerate (X1-R2-02). |
| X1-R1-10 | Fixed | Every unique vertex and every cross-repository combination is computed; min LB and max UB; "vertex envelope" naming (P l.101). The exhaustive fixture covers only one d-group and one π value (X1-R2-07). |
| X1-R1-11 | Fixed at protocol level | Explicit `DrawAudit` on whole-run fallback (P l.54). Product-backed fixtures are deferred (P l.139), as round 1 accepted. Their verdict consequence is revised by X1-R2-04. |
| X1-R1-12 | Fixed | HD9 narrowed (P l.7; `hypothesis-ledger.md:36`); continuity proxy with its limitations (P l.107). |
| X1-R1-13 | Fixed | Seven examined repositories plus the E2 pool; D1-R uses fastapi/axum; the test repositories are disjoint (P l.17–19). |
| X1-R1-14 | Fixed | (a)–(f) at P l.145–153, including the loss formula at P l.151. The zero-escalation case is undefined (X1-R2-09). |
| X1-R1-15 | Fixed | Severity removed (P l.83); `label_payload` with its fixture. |
| X1-R1-16 | Fixed | HMAC specification (P l.58), matched by `audit_uniform`; per-repository store; collision refusal in `validate()`. Resume misalignment remains (X1-R2-11). |
| X1-R1-17 | Fixed | Seed 169200 round-robin; global seed 169100; CLI pins; one Jev realisation (P l.60, l.85). |
| X1-R1-18 | Fixed | Reasons are adjudication inputs; siding is reported (P l.85, l.89). Its denominator and consequence are undefined (X1-R2-09). |
| X1-R1-19 | Fixed | GUIDE digest in every row; author isolation; the draft author is declared ineligible (P l.81). |
| X1-R1-20 | Fixed | Union / both-binary κ; three-category confusion; audit-only κ; undefined κ fails (P l.89). |
| X1-R1-21 | Fixed | Canonical sort (Python code-point order equals UTF-8 byte order); reversal fixture; regenerated artifact. |
| X1-R1-22 | Fixed | Cutoffs for A and B with R-068 links (P l.23); meaning of "fresh" (P l.9). The adjudicator's cutoff is missing (X1-R2-15). |
| X1-R1-23 | Fixed | Baseline-bottom comparator with overlap and routes disclosed (P l.75). |
| X1-R1-24 | Fixed | Study lines versus product scorecard distinguished (§6, REQ-138 paragraph). |
| X1-R1-25 | Fixed | P l.125 wording matches the required text; the flag is acknowledged as absent. |
| X1-R1-26 | Fixed | Study-only sampler scope; product .05 unchanged (§4 Scope). |
| X1-R1-27 | **Partly fixed** | HD9 (l.36, history l.40) and R-065 to R-068 (l.72–75) were appended. **R-064 is missing on this branch**, yet P l.11 and HD9 cite it, and the README proposes a conflicting R-064 row (X1-R2-13). |

## 5. New issues

### X1-R2-01: blocking. The primary interval is not validated against its estimand, and the coverage rule cannot be executed

- **Location:** P l.39 ("first20 eligible runs chronologically"), l.99, l.123, l.133 ("Fixed synthetic rankings across trial audit draws isolate audit/cluster sampling OC"), l.135, l.137 ("any MC lower<.90 triggers review" alongside "coverage/power issue→not testable"). `self_test.py` planning loop: one `synthetic(...)` population per cell built outside the trial loop; `truth` is that frame's census gain; trials vary only `draw()` and the bootstrap seed. README "Exercised rev2" and "Coverage review flag".
- **Problem:**
  - **The target is wrong.** The run-cluster bootstrap was adopted to reflect run-to-run variation (X1-R1-05), but no simulation redraws runs. Coverage is measured against the fixed frame's realised gain under audit sampling only.
  - **Half the cells are degenerate.** In the 24 populations with no unannotated share, every π = 1, because π_h = min(1, 50/(2·25)). All 100 trials therefore give the same point estimate. JSON for those cells:
    - coverage 1.0;
    - P(support) exactly 1.0 at gain .10 and .20 and 0.0 at gain 0.
  - **Planning figures are misreported.** The README's "support Wilson-MC lower ≥ 0.963" is the Wilson bound for 100 copies of one deterministic outcome.
  - **P l.123 is contradicted.** It says boundary-.10 support is "near one half", but the recorded value is 1.0 in every census cell (it is .47–.62 only where unannotated = .20).
  - **Coverage falls short even on audit sampling alone.** The audit-only cells with variation show coverage .90–.93: p .40/u .20/SD 0/g .20 gives 90/100, Wilson [.8256, .9448]. That is consistent with the known few-cluster downward bias of the naive cluster bootstrap's variance, a factor of (G − 1)/G at G = 20 [INFERENCE]; the 99-replicate percentiles add to it.
  - **The cells with heterogeneity over-cover (1.0).** The bootstrap sees run differences that the simulation holds fixed.
  - **The gate is neither executable nor adequate:**
    - "Triggers review" names no action, and P l.137 also says "coverage issue → not testable".
    - A two-sided MC lower bound of .90 at 10,000 trials passes a true coverage of about .905. One-sided miss rates matter more: the LB miss rate drives support, the UB miss rate drives rule 4.
    - README l.26 requires the final OC "before any real stage", yet the OC needs locked run sizes from the model-free inventory.
  - **Temporal dependence.** "First 20 eligible runs chronologically" selects consecutive commits, which are plausibly dependent (one feature's work) [INFERENCE]. That breaks the independent-cluster assumption the run bootstrap relies on, and no simulation can check it on real data.
  - **Compute is not planned.** P l.137 prescribes 10,000 trials × 1,999 replicates over the full grid. Each trial computes up to 8³ vertex combinations plus a multiplier interval per combination (`analyse`). No feasible compute plan is stated [INFERENCE: pure-Python cost], which invites a later unreviewed reduction.
- **Required fix (recommended option B):**
  - (a) **Run selection.** Replace "first 20 eligible runs chronologically" with a seeded simple random sample without replacement:
    - order every eligible run in the envelope by HMAC-SHA-256 (key UTF-8 `169-x1-runs`, message repository NUL commit SHA), ascending, and take the first 20;
    - the pool must hold ≥ 20 eligible runs, otherwise not testable;
    - the N_r ≥ 1,000 and ΣN ≤ 60,000 gates apply to the sample unchanged, with no redraw.
    - State the estimand: the equal-repository recall@k gain over all eligible runs of the three pinned envelopes. Transfer to future commits is a §12 limitation.
  - (b) **Primary interval.** Keep the run-cluster percentile bootstrap (B = 1,999, seed 169, common draws, vertex envelope), but resample **m_r = G_r − 1 = 19 runs** with replacement per repository (McCarthy & Snowden 1985). This removes the (G − 1)/G variance bias. Omitting the finite-population correction is conservative; state that.
  - (c) **OC design.** For each cell and trial, run an outer loop:
    - draw the SRS from a synthetic finite eligible-run population whose sizes come from the locked inventory's eligible-run frame;
    - per-run gain ~ N(gain, SD²), with SD ∈ {0, .10};
    - include d = −1 discordance (X1-R2-02) and the cap-order tier (X1-R2-05);
    - then the inner audit draw;
    - truth = the population's equal-repository gain.
    - Keep the fixed-frame audit-only simulation as a labelled diagnostic.
    - Add true gains {.05, .08, .10, .12} to show P(support) near the threshold.
  - (d) **Binding gate.** In every safety-clear reference cell, at ≥ 10,000 trials, each of P(LB > truth) and P(UB < truth) has Wilson MC95 upper ≤ .035.
    - Failure or an unmeasured cell means X1 is not testable before models.
    - There is no fallback method and no waiver. An amendment needs a new zero-issue round.
    - Delete "MC lower < .90 triggers review".
  - (e) **Sequencing.** The full OC runs after the model-free inventory locks the eligible-run frame and before any model call. Correct README l.26 and `reviews/README.md` accordingly.
  - (f) **Compute plan.** Exclude the descriptive conditional interval from the OC. State the trial counts per cell type: 10,000 for complete-label cells and at least 2,000 for missing-label cells, each with its stated MC precision. The implementation must be shown to meet these within the envelope before full authority.
  - (g) **Reporting.** Correct P l.123 and the README so they match the recorded results. Label census-cell MC bounds as deterministic.
  - (h) Add McCarthy & Snowden (1985) and Rao & Wu (1988) to the research ledger with new R-IDs, and cite them in P §6.

### X1-R2-02: major. The planted null is structurally zero, and the generator never produces disfavourable discordance

- **Location:** `ranking.py` `synthetic()`:
  - `base_top = positive[:base_count]+…`;
  - `jev_top = available_p[:jev_count]+…`;
  - `available_p` preserves the order of `positive` after removing the tail reserved for the unannotated tier.

  `self_test.py`: planted null `synthetic(0., unannotated=.20)`, and `hand_frame(null=True)`. JSON `planted_null.interval` = [0, 0]; `unequal_pi_hand.interval` = [3/7, 3/7].
- **Problem:**
  - With gain 0 and SD 0, `base_count == jev_count`, and both arms take the same prefix of `positive`. The retrieved positive sets are identical, so every positive has d = 0, D = 0 and the interval has zero width.
  - With gain > 0, the baseline's retrieved positives are a subset of Jev's, so no positive has d = −1. Paired variance is understated, and power and coverage are optimistic.
  - The gain-0 false-support check is satisfied automatically by the point ≥ .10 rule.
  - Each of the following would still pass every null check: an arm swap, unpaired arm resampling, unit-level instead of run-level resampling, or ignoring stored π.
  - The hand null discriminates the weighting, but its 20 runs are identical, so its interval is degenerate.
- **Required fix:**
  - Draw each arm's retrieved positives independently given its target recall, with a preregistered overlap parameter ρ ∈ {0, .5}. Then d = ±1 occurs in equal HT measure under the null and varies across runs with unequal π.
  - The planted null asserts non-zero interval width and that 0 is covered.
  - Add estimator-level mutations, each of which must fail at least one check: arm swap, unpaired arm resampling, unit-level resampling, ignoring stored π, nominal π.
  - Regenerate the evidence.

### X1-R2-03: major. The safety OC is deterministic and leaves out the likeliest veto route, so whether support can be reached is never projected

- **Location:** `self_test.py`:
  - `unresolved_indices = … round(unresolved_rate*bottom_count)`;
  - `critical_indices = positives[:round(critical_share*len(positives))]`.

  `ranking.py` `severity_control = negative[0]`, truth 0. P l.87 (session invalidity becomes unresolved), l.131, l.135.
- **Problem:**
  - Unresolved and critical cases are fixed counts, not per-case events. At rate .005, round(.005 × ~1,500) = 8 unresolved cases, so P(unmeasured) is exactly 1.
  - JSON: P(support) = 0 in every cell with a non-zero unresolved rate, and ≤ .25 at critical share .01.
  - Support therefore needs zero unresolved cases among M = Σ floor(N_run/2) census cases, at least ~1,500 and up to ~30,000. In practice the per-case residual rate (criticality `cannot_tell` + truth `cannot_tell` on severity-critical units + invalid sessions) must be ≲ 1/M. No step measures this before test Jev spend.
  - The "safety-clear" reference forces every scanner-critical unit to be false. Jev ignores severity (P l.47), so a true scanner-critical finding in Jev's bottom half vetoes the default. That likely route is absent from the reference.
  - Both inconclusive and not-supported map to the same recommendation (P l.125). Running into a near-certain inconclusive outcome therefore spends the 45-day envelope without informing the decision.
- **Required fix (futility stop recommended):**
  - Model unresolved and critical status as per-case Bernoulli events, with unresolved rates {0, 1e-4, 5e-4, 1e-3, .005, .02} and critical shares {0, .001, .01, .05}.
  - Add cells where scanner-critical units are true at the cell's prevalence.
  - **Inventory** (model-free): report the count of severity-critical units per repository.
  - **Before any test Jev call:**
    - compute the dev unresolved rate r̂: the unresolved predicate applied without the floor condition to dev canonical labels, with invalid sessions counted;
    - compute its Wilson 95 % lower bound r_L;
    - compute M from the locked frame.
  - If (1 − r_L)^M < .05, stop. The outcome is **inconclusive (safety undecidable, pre-model futility)**, the §7 recommendation applies, and no test Jev call is made.
  - Also report the projected P(severity veto) under the OC rank model, descriptive only.
  - REQ-169's absolute veto is unchanged.

### X1-R2-04: major. Undelivered treatment can produce a decisive "not supported"

- **Location:** P l.54, l.157 ("unknown-cost receipt latches before later calls"; "Product fallback tier retains unasked rows"), l.161 (the 7-day Jev clock).
- **Product behaviour:** `internal/decisions/run.go:185–206`. A reply with no billed cost stops only the current run (`blind`).
- **Problem:**
  - The study's latch stops **all later** Jev calls, and leaves their units in the baseline-ordered fallback tier.
  - Units unscored at the Jev-clock ceiling are treated the same way. So are harness-caused whole-run fallbacks (disabled, missing key, invalid config or threshold).
  - None of these is the deployed default-on treatment. They push D towards 0 and can trigger rule 4 ("not supported") instead of "inconclusive", and a veto computed on an undelivered ranking is equally invalid.
  - This mapping was round 1's own X1-R1-08 fix, faithfully implemented. The council corrects it here.
- **Required fix:**
  - Classify every non-annotation reason:
    - **Product treatment:** per-run call and cost caps; the per-run unknown-cost stop; per-unit transport or mismatch failures.
    - **Study or harness:** the study latch; the Jev-clock stop; disabled; missing key; invalid config or threshold.
  - Any study or harness reason affecting a test run is an **integrity failure (rule 1 → inconclusive)**. Its `DrawAudit` rows are still recorded (X1-R1-11).
  - **Delivery gate:** a repository whose per-unit transport or mismatch failures exceed .10 of attempted units is an integrity failure.
  - Remove the study-wide unknown-cost latch. For the $31 accounting, charge any run with an unknown receipt its full $0.50 reservation. Product behaviour is then unchanged and the ceiling still holds.
  - Report annotation coverage by reason, per run.

### X1-R2-05: major. The OC does not model the product's cap order, so power at locked run sizes is overstated

- **Location:** P l.131–133, l.137; `ranking.py` `synthetic()` places the unannotated tier at random, sharing {0, .20}. Product: `internal/decisions/run.go` `Order` (triage units ordered by baseline score, descending) and `Run` (caps consume units in that order).
- **Problem:**
  - At R-052's historical ≈ $0.000589 per request, the $0.50 cap binds near M ≈ 850 calls [INFERENCE; the triage price is unmeasured].
  - In runs larger than M, the unannotated tier is exactly the baseline's tail, and Jev can only reorder the baseline's top M.
  - The OC's random 0 % / 20 % tier represents neither this structure nor the large unannotated shares it implies for runs up to 2,000.
- **Required fix:**
  - The full OC models the annotated set as the baseline's top M in cap order, with M ∈ {2,000, 850, 400}, plus a random per-unit failure rate of {0, .02}.
  - After dev calls, set M̂ = min(2,000, floor(0.50 / c_U)), where c_U is the 95 % upper bound of the dev mean per-call cost.
  - If the OC row for the largest grid M ≤ M̂ fails the X1-R2-01 gate, or the gain-.20 power gate, at the locked sizes, X1 is not testable before test models. No new review is needed to apply this.
  - Inventory reports the share of locked units beyond each grid M.
  - Results report per-run gain against N_run, split at M̂ (descriptive).

### X1-R2-06: major. The throughput and clock gates cannot be executed and are not conservative

- **Location:** P l.159 ("2×(remaining reserved sessions×p90/4)"), l.161 (45-day / 48 h / 7-day assumptions); README "Throughput".
- **Problem:**
  - "Remaining reserved sessions" is undefined: it could mean attempts, logical sessions or cases, and the retry reservation could be counted separately or twice.
  - A multiple of p90 does not bound the mean. With 91 % one-second and 9 % 900-second sessions, p90 = 1 s but the mean is 81.9 s (challenger) [INFERENCE: arithmetic].
  - No minimum dev timing sample is set, and nothing happens if no valid sample exists.
  - The clocks have no defined start, pause, resume or stop.
  - The 48 h inventory window has no consequence.
  - Subscription quota or rate-limit responses would count as transport failures. After one retry they become unresolved labels, which blocks support (P l.87).
  - At 3 sessions per case with the 2× margin, a p90 of 60–90 s limits the frame to roughly 23,000–35,000 units within ~36 days (peer) [INFERENCE]. The 60,000 gate may therefore never bind.
- **Required fix:**
  - **Reserved attempts** = (2 + max(dev adjudication rate, .40)) × (1 + dev retry rate) × labelled cases in the locked frame (census ∪ audit).
  - **Projection** = 2 × reserved attempts × U95(mean attempt time) / 4, plus observed scheduling overhead. U95 is a one-sided 95 % upper bound from at least 200 completed dev cases. With fewer cases, X1 is not testable before test models.
  - **Remaining envelope** = 45 days − elapsed − 7-day Jev reservation.
  - **Quota and rate-limit responses** pause the queue with backoff. They consume no attempt, and the clock runs. Pauses are counted and reported.
  - **Clocks.** The 45 days start at the first action under preparation authority and stop at gold freeze. The 7-day Jev clock covers cumulative call wall time. Its breach maps to X1-R2-04.
  - The 48 h figure becomes a planning estimate with no gate.

### X1-R2-07: minor. Mutation and fixture gaps

- **Location:** `self_test.py`:
  - `('recomputed selected fraction', len(sample_run)/7)` equals 1, the same as the omitted-weight mutation;
  - hand intervals are degenerate because all 20 runs are identical;
  - the exhaustive fixture puts all six unknowns in issue `/3` (d = −1, π = .5).

  `ranking.py` `synthetic()`: `positive[-min(...):]` returns every positive when the minimum is 0, and the comment "positive fraction is about .20" is wrong (the share equals the prevalence).
- **Required fix:**
  - Use a fixture with unselected rows, so that the selected fraction differs from both 1 and π.
  - Test `cluster_distribution` with hand-specified draws over runs that differ.
  - Extend the exhaustive completions across all three d-groups, with unequal π and heterogeneous runs.
  - Explicitly test the zero-denominator omission and the > 5 % inadequacy rule.
  - Guard the m = 0 slice and correct the comment.

### X1-R2-08: minor. Missing intervals and missing provenance in the evidence

- **Location:** P l.135 ("intervals/counts in every cell"); `self_test.py` scenario records (only `support_MC95` and `coverage_MC95`); `synthetic-results.json`.
- **Required fix:**
  - Emit Wilson MC95 intervals for P(veto) and P(safety unmeasured) in every cell.
  - Embed the SHA-256 of `ranking.py` and `self_test.py` and the Python version in the JSON.
  - Regenerate create-once.

### X1-R2-09: minor. Siding and zero-escalation outcomes are undefined

- **Location:** P l.89 ("A siding>.80 triggers council review"), l.149–153.
- **Problem:**
  - The siding denominator is undefined. So is its handling across the two question dimensions, invalid responses, and the zero-adjudication case. Review has no defined outcomes.
  - A grid threshold with zero escalations makes P = TP/(TP+FP) undefined. All-scoreless held-out cells are reachable through fallback.
- **Required fix:**
  - **Siding:** compute per dimension, per repository and pooled, on cases where A and B gave different valid answers in that dimension and C answered validly.
    - The trigger is pooled A-siding > .80 with a denominator ≥ 20; below that, report low-n with no trigger.
    - Review outcomes are limited to (a) proceed, or (b) labels inadequate, which makes the result inconclusive. No relabelling and no GUIDE change.
  - **Zero escalations:** precision is unmeasured and that threshold is infeasible. Held-out zero-positive or zero-escalation cells are insufficient.

### X1-R2-10: minor. Two estimator details are unstated

- **Location:** P l.91, l.99; `ranking.py` `percentile()` uses `len(distribution)`.
- **Required fix:**
  - State that canonical truth `cannot_tell` is unknown Y: it enters the missing share and the vertex envelope.
  - State that percentile indices use the retained replicate count B′ after zero-denominator omissions.

### X1-R2-11: minor. The HMAC lookup misaligns on resume

- **Location:** P l.58 ("Precompute lookup in canonical ranked-issue call order"); `internal/decisions/audit.go` `DrawAudit` (calls `rand()` only for units without a stored draw).
- **Problem:** On resume, units already drawn consume no value, so a full-order lookup shifts. The "assert consumption aligns" check then refuses every resume.
- **Required fix:** At each call, build the sequence over units that still lack a stored draw, in `Assign` order. The product-backed harness tests a crash-resume.

### X1-R2-12: minor. Competing explanations lack descriptive lines, and the scope of k is unstated

- **Location:** P l.50, l.105, §12.
- **Required fix:** State that the verdict applies at k = 30 %, which is justified by feasibility, not by observed reviewer capacity. Add the following, outside the verdict and with no rescue:
  - recall gain at k ∈ {.20, .40} (labels are near-census, so this costs nothing extra);
  - per-repository headroom min(1, k/T) − R_B, and gain as a fraction of headroom;
  - the share of Jev's top-k boundary decided by ties in P(true);
  - gain by path role;
  - for each violating case, whether it rests on A/B agreement or on adjudication, plus positive agreement on criticality.

### X1-R2-13: minor. Ledger integrity: R-064 is absent, the README's R-064 conflicts, and status text is stale

- **Location:**
  - `docs/research-ledger.md` l.71–72 (R-063 is followed by R-065);
  - P l.11 and HD9 (`hypothesis-ledger.md:36`) both cite R-064;
  - the README's "Exact ledger append text" R-064 row;
  - the "not represented as already landed" wording in the README, `reviews/README.md`, P l.11 and the response row for X1-R1-27;
  - R-067's link: an absolute path into the mutable sibling worktree `scout-x11`.
- **Problem:**
  - R-064 is defined in the ADR-011/012 work (`scout-adr11/docs/research-ledger.md:72`: "User direction: delegate recommended design decisions…", full verbatim quote).
  - The README's row gives R-064 a different title, source and truncated quote. Appending it, as both peers suggested, would create a second, conflicting definition of one ID.
  - HD9 cites an ID that does not exist on this branch.
- **Required fix:**
  - Delete the README's R-064 append row, and cite R-064 as defined in the ADR-011/012 change.
  - Round 3 requires that row to be present, identically, on X1's base.
  - Update every "not landed" statement to reflect that HD9 and R-065 to R-068 are present.
  - Append-only: add a provenance-correction row binding R-067 to a commit SHA or file digest of the X11 preregistration revision 4 that was read.
  - Add the R-IDs from X1-R2-01(h).

### X1-R2-14: minor. The delegated decisions do not say how the owner's "capture as ADRs" direction is met, and the new choices are missing

- **Location:** README "Delegated decisions" ("for later owner-review ADR … No invented ADR number").
- **Required fix (recommended option):**
  - State that REQ-169 (`docs/scout-design.html:686`) places k, floor, sampling, interval and the treatment of capped and failed units inside the council-signed preregistration. These are therefore experiment-protocol choices, captured by this table and the signed round records, and need no design ADR.
  - Design-affecting outcomes each need their own ADR: the default switch, `--rank jev`, product sampler parameters, and the D1-R set-or-replace decision.
  - Add table rows for each new choice:
    - run SRS;
    - McCarthy–Snowden interval;
    - one-sided coverage gate;
    - safety futility stop;
    - treatment-delivery classification and the .10 gate;
    - cap-order M grid;
    - throughput formula and quota pause.

### X1-R2-15: minor. Legibility of the frozen text, and the adjudicator's cutoff

- **Location:** throughout P, README, `reviews/README.md` and the response. Examples: P l.1 "Round1 found27"; l.39 "first20 eligible runs"; l.85 "Claude2.1.311 / Codex0.166.0"; l.157 "USD31"; "plusUSD1". Also P l.23 and R-068 give no cutoff for C (`claude-sonnet-5-5`).
- **Problem:** The signed text is frozen permanently (rule 5), and fused numbers and words invite misreading. The adjudicator decides contested labels, so its training exposure matters as much as A's and B's.
- **Required fix:**
  - Restore spacing throughout before freeze.
  - Record the published knowledge cutoff for Sonnet 5.5, or "unreported", with a link, through an append-only R-068 correction or a new R-ID.

## 6. Assessment of the targeted questions

- **Run-cluster coverage (MC lower .826).** It is not adequate, and not because the true coverage is .826: .826 is a lower confidence bound, not an estimate. The coverage was measured against the wrong target, the gate is lax and cannot be executed, and audit-only cells already show .90–.93. See X1-R2-01. Keep point ≥ .10, LB > 0 and the veto. The LB rule is usable only behind an established adequacy gate. Peer note: once point ≥ .10, LB > 0 rarely binds, so miscoverage mainly affects rule 4 and the envelope.
- **Decidability of the safety rule.** The predicate is exhaustive and decidable given frame and labels:
  - ranks are permutations and the floor is an integer;
  - missing truth or criticality maps explicitly to cleared or unresolved;
  - the kernel matches P l.68–71.

  Reachability is the remaining problem: support requires zero unresolved cases among ~1,500–30,000 census cases (X1-R2-03).
- **AuditFraction 1.0.** Verified.
  - Product `Assign` caps π at 1, and `RankingInput.AuditFraction` accepts 1.0.
  - Census-only labels stay out of the HT estimators (P l.64).
  - "Not a promise all π = 1" is correct: π ≥ 1/H.
  - Its costs are near-census labelling (handled by the X1-R2-06 gate) and a degenerate conditional interval (correctly secondary).
- **k = 30 %.** Consistent across P l.50, `Unit.k`, the README and the generator. Feasible in every grid cell. Its scope is unstated (X1-R2-12).
- **Cost and wall ceilings.**
  - $31 = 60 × $0.50 + $1 is arithmetically right, and in-flight overshoot is disclosed.
  - The study-wide unknown-cost latch and the Jev-clock stop are mis-mapped (X1-R2-04).
  - The throughput and clock definitions are incomplete (X1-R2-06).
- **Delegated decisions.** All 15 README rows are consistent with round 1's §5 positions. The safety-clear power qualification is **accepted**: requiring 80 % support where REQ-169 mandates a veto would be contradictory, and all adverse cells stay reported. The capture mechanism and the new choices need to be added (X1-R2-14).
- **Ledger rows.**
  - HD9 matches the narrowed endpoint, is `proposed`, relates to HD2 and has a history line.
  - R-065 to R-068 match their roles in P l.11. R-065's product claims agree with `consent.go:25,30` and `audit.go`. R-068's links support the stated cutoffs (challenger checked them).
  - R-064 needs fixing (X1-R2-13).

## 7. Checked and sound

- The kernel's sampler mirror matches product `Assign` and `DrawAudit` (`audit.go`: floor (n+1)/2; certainty; strata; π = min(1, f·n/H/N_h); persisted first draw).
- The verdict order in `verdict()` matches P l.115–121, and the fixtures test each precedence.
- The bootstrap carries runs with no sampled rows as zero-total clusters (`cluster_distribution` uses a `(0, 0, 0)` default).
- The same draws are shared across vertices and arms, and repository draws are independent.
- The HMAC uniform is in [0, 1) and repository-scoped, and its encoding matches P l.58.
- The 88 listed checks equal `checks_passed`. There are 576 cells. Per-cell counts sum to 100 and the probabilities match their counts. The README's extremes match the JSON. Neither peer found a structural mismatch between code and output (numerical reproduction was not exercised).
- The examined-data quarantine, test-repository disjointness, labeller protocol and provenance (`agent_reviewed`, append-only, no per-case owner review) comply with `CLAUDE.md` and R-039.

## 8. Evidence index

- REQ-169: k, floor, interval and the treatment of capped units are fixed by the preregistration (`docs/scout-design.html:686`). REQ-41: "clustered bootstrap" (`:682`).
- Cap order and the unknown-cost stop: `internal/decisions/run.go` `Order` ("triage by baseline score descending") and `Run` ("A reply that reports no billed cost stops the run too").
- Sampler: `internal/decisions/audit.go:48–80`, `:95–115` (`rand()` is called only for units without a stored draw). Fraction plumbing: `internal/engine/ranking.go:69,371–374`. Caps: `internal/policy/consent.go:25,30`.
- Planning outcomes (jq over `synthetic-results.json`):
  - coverage < .95 occurs only in u = .20/SD 0 cells (.90, .92, .93, .93);
  - every u = 0 cell has coverage 1.0, with support 1.0 at g ≥ .10 and 0.0 at g = 0;
  - support is 0 in every cell with a non-zero unresolved rate;
  - at critical share .01, support is .25 / .09 / .01 / 0 for prevalence .10 / .20 / .40 / .55 (u = 0, SD 0, g = .20).
- R-064's existing definition: `/Users/adamjackson/Projects/work/scout-adr11/docs/research-ledger.md:72`; `scout-adr11/docs/adrs/2026-11-08-gitleaks-policy-version/ADR.md:26`.
- Few-cluster bootstrap correction: Rao & Wu (1988), JASA 83(401):231–241, https://www.tandfonline.com/doi/abs/10.1080/01621459.1988.10478591; McCarthy & Snowden (1985), Vital Health Stat 2(95), https://pubmed.ncbi.nlm.nih.gov/3992938/.

## 9. Risks

- **Safety likely decides the verdict.** At critical share ≥ .01, P(veto) is .75–1.0 in the reference grid. REQ-169's absolute veto may end Jev's default whatever its retrieval quality. The council does not alter REQ-169. X1-R2-03 makes this visible before any test spend.
- **The futility stop may end X1 before test Jev calls.** Using the Wilson lower bound means a stop needs about ≥ 3 unresolved in 300 dev cases at M ≈ 1,500 [INFERENCE: arithmetic]. The outcome maps to the same recommendation as inconclusive. Reversible only by a new protocol.
- **Changed estimand.** SRS restricts the claim to the pinned envelopes' eligible runs. The conservative McCarthy–Snowden interval costs some power, and the OC will show how much.
- **Throughput and quota.** Near-census labelling of up to ~60,000 cases may exceed subscription capacity. The X1-R2-06 gate stops before test calls, so the cost is time, not validity.
- **Cap order.** If dev cost implies a small M, large runs carry little reorderable mass. X1-R2-05 decides testability before test calls.
- **Reversibility.** Every fix is pre-data. No product, default, design or ADR is changed.

## 10. Where the peers changed the council's recommendation

- **Challenger:**
  - Proposed sampling runs from the inventoried envelope. **Adopted** as the core of the X1-R2-01 fix: the lead had planned only a superpopulation simulation plus a variance correction.
  - Supplied the counterexample showing 2×p90 is not conservative (X1-R2-06).
  - Found the missing veto and unmeasured intervals (X1-R2-08) and the undefined siding and zero-escalation outcomes (X1-R2-09).
  - Found the README's "before any real stage" sequencing contradiction.
  - Rated coverage as blocking; the lead concurs. Its two-sided "MC lower ≥ .90" threshold is **overridden** by the peer's one-sided ≤ .035 miss-rate gate.
- **Experiment peer:**
  - Found the cap-order omission (X1-R2-05), which the lead missed.
  - Found the HMAC resume misalignment (X1-R2-11), the P l.123 contradiction and the nested-arms mechanism behind the degenerate null.
  - Supplied the Bernoulli safety rates, the scanner-critical-true cell, the .10 delivery gate, the estimator details, the JSON provenance and the competing-explanation lines.
  - Three of its proposals are **overridden**:
    - Appending the README's R-064 row would collide with ADR-011/012's R-064.
    - Its report-only safety projection becomes a conservative futility stop.
    - Its "48 h overrun → not testable" becomes a non-binding planning figure.
- **Found by the lead only:**
  - the R-064 collision with `scout-adr11`;
  - removing the study-wide unknown-cost latch through per-run reservation accounting;
  - the McCarthy–Snowden m = G − 1 correction;
  - the temporal-dependence argument against first-20 selection;
  - quota pause semantics;
  - the full-OC compute plan;
  - the legibility of the frozen text and the adjudicator's cutoff.

## 11. Execution outline for revision 3 (one integration owner)

1. **P §3/§6 and kernel:** SRS run selection; McCarthy–Snowden m = G − 1 interval; estimand statement; B′ percentile wording; `cannot_tell` treated as unknown (X1-R2-01a–b, X1-R2-10). *Test:* a fixture showing the m = 19 variance on hand-specified differing runs, and selection reproducibility from the HMAC.
2. **Generator and OC:**
   - outer SRS loop over a synthetic eligible-run frame;
   - random per-run gain;
   - overlap ρ ∈ {0, .5};
   - cap-order M ∈ {2,000, 850, 400} with failures {0, .02};
   - Bernoulli safety rates and scanner-critical-true cells;
   - gains {0, .05, .08, .10, .12, .20};
   - one-sided miss-rate reporting;
   - veto and unmeasured intervals;
   - script digests embedded in the JSON.

   Refs: X1-R2-01c–d, X1-R2-02, X1-R2-03, X1-R2-05, X1-R2-08. *Test:* the noisy planted null has non-zero width and covers 0; every estimator mutation fails.
3. **Fixtures:** the X1-R2-07 items, including the m = 0 guard.
4. **P §8/§10/§11:** the binding coverage gate; sequencing (inventory → full OC → dev → test); compute plan; dev-phase gates (throughput formula, the futility stop using r_L, M̂); clock and quota semantics; treatment-delivery classification and integrity mapping; removal of the study-wide unknown-cost latch (X1-R2-01d–f, X1-R2-03, X1-R2-04, X1-R2-05, X1-R2-06).
5. **P §5/§9/§6/§12:** siding and zero-escalation rules; descriptive lines and k scope; corrected P l.123; HMAC resume rule (X1-R2-09, X1-R2-11, X1-R2-12, X1-R2-01g).
6. **README and response:** delegated-decision rows and capture statement; corrected planning claims; legibility pass; adjudicator cutoff (X1-R2-14, X1-R2-15).
7. **Ledgers (Main):**
   - R-064 present on base, identical to ADR-011/012's definition, and the README's row deleted;
   - new R-IDs for Rao & Wu and McCarthy & Snowden;
   - the R-067 provenance-correction row;
   - status text updated (X1-R2-13).
8. Run `self_test.py` once from the pinned toolchain with create-once output. Record digests of every reviewed file, then submit round 3.

## 12. Stage authorisation

| Stage | Status now | Earliest path |
|---|---|---|
| Drafting and synthetic validation | **Permitted** (unchanged) | — |
| Preparation (model-free pins, exposure manifest, GUIDE-author record) | **Not authorised**: 15 open issues (rule 1), and its concrete pins and recipes do not exist yet (P §11.1) | First candidate after a zero-issue round on revision 3, reviewed together with its submitted artefacts |
| Inventory (deterministic scans, SRS frame, locked sizes) | **Not authorised** | After signed preparation outputs and a scoped review |
| Full protocol and implementation (GUIDE, product-backed harness, full OC) | **Not authorised** | After inventory locks sizes; the full OC must pass X1-R2-01d and X1-R2-05 |
| Dev Jev and labels; test scoring and labels | **Not authorised** | Under full authority only; test calls only after the dev gates pass (throughput, futility, M̂) |
| Gold unblinding and analysis | **Not authorised** | — |
| Default, threshold or product-sampler change | **Not authorised** | Separate signed ADR |

**Sign-off statement:** none. Round 2 closes with **15 issues open** (X1-R2-01 blocking; X1-R2-02 to X1-R2-06 major; X1-R2-07 to X1-R2-15 minor). The reviewed digests are those in the table above. No preparation, inventory, Jev, labelling or analysis stage is authorised.
