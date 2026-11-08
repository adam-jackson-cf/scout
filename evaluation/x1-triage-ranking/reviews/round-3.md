# X1: council experiment review, round 3 (revision 3 with ADR-014)

Date: 2026-11-08. Reviewer: innovation council. The lead wrote the synthesis. Two read-only peers ran in parallel: `innovation-challenger` (X1Council3.ChallengerX1R3) and `experiment-peer` (X1Council3.PeerX1R3).

This record is X1 round 3. It is also the first council review of ADR-014, and X1-R3-06 asks for it to be saved a second time as that ADR's `round-1.md`.

**Governance.**
- `CLAUDE.md` experiment rules 1–7.
- Design rules 1–3, for ADR-014.
- Ledger rules, and the General sections "Jev labelling", "Model use" and "Secrets".
- Standing owner delegation R-064: where an owner choice arises, the council's recommended option is stated as the fix.

**Scope.** Revision 3 of:
- `PREREGISTRATION.md`, `README.md` and `reviews/README.md`;
- `reviews/round-2-response.md`;
- `validation/{ranking.py,self_test.py,outer_simulation.py,synthetic-results.json,outer-results.json}`;
- the preserved `validation/synthetic-results-rev1.json` and `synthetic-results-rev2.json`;
- `docs/adrs/2026-11-08-x1-delegated-decisions/ADR.md` (ADR-014);
- the uncommitted rows in `docs/hypothesis-ledger.md` (HD9) and `docs/research-ledger.md` (R-064 to R-068, R-071, R-072).

Each round-2 item was checked against its required fix in `reviews/round-2.md`. Product code was read where the protocol depends on it: `internal/decisions/{audit.go,run.go,client.go}`, `internal/engine/ranking.go` and `internal/policy/consent.go`.

**What was and was not run.**
- No shell, no model call, no self-test or simulation run, and no digest recomputation.
- Evidence comes from reading files, jq queries over `validation/outer-results.json`, and arithmetic.
- One web search confirmed the Maurer–Pontil bound (X1-R3-10).
- `codegraph` was not used. Its reference documents were read first, but no executable was available in this read-only session.
- Paths are relative to `evaluation/x1-triage-ranking/` unless stated. "P l.N" means a line of revision 3's `PREREGISTRATION.md`; "ADR l.N" means a line of ADR-014.

**Reviewed files (SHA-256).** Main checked all 17 entries of `reviews/rev3-digests.json` with shasum. Main reports the manifest's own SHA-256 as `894dc63b0522a23589187c6de9fa4e672878fc1c07b85bfbbdcd224e5ab2f448`. This round did not recompute any digest.

| File | SHA-256 |
|---|---|
| `docs/research-ledger.md` | `d474ba10c3f256d6963c2854a9a6e7043bec35225fa6118712124ff19b5566b7` |
| `docs/hypothesis-ledger.md` | `cf3168997ff1cb68c46c5b4eda3aac8dd70ea79dd3f8e7438da9a7c0c70c1eb0` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/ADR.md` | `21113bcfc3affd34fdee28a3ddf3e05ffe55b3d74d7ca10b16df339bcfbf4a95` |
| `PREREGISTRATION.md` | `ce3cbc33fa6a31965ca4d13cd0eb17c90f0ccaa5242aaf1ec21bf70a6032b1da` |
| `README.md` | `0edc2b270c0234a7e04205f23261968cdd9aa133ab51c2a033816196a0c64723` |
| `reviews/README.md` | `1394ad8e0473c6b4bc92ebbab7716ef5b2873ca3806264bbc6fb69c29a34948f` |
| `reviews/round-1-response.md` | `ad8ebd2430b198422ef51dcd648a0be84d0d948cd9fe87664b3f57eb94d47d77` |
| `reviews/round-2-response.md` | `c630dddff2381f62e21fcab809d0ba201573a45f90cd0aec15a6ac199eb31233` |
| `reviews/round-1.md` | `89eab4a5d0f63487da200ab34db28dc0528ed9fcc73745c9d1075272339e172d` |
| `reviews/round-2.md` | `4021d173533876e775bf9200bc3a1aa1164f04feffbe9f3beb56dfb3b400d1ef` |
| `validation/ranking.py` | `6b2910f7b1ee409d5ead14ff8ecd8cc12626870224610b798c899ce1ef4133d2` |
| `validation/self_test.py` | `0f5f98013495ef9ad968d860de78da65b8f781db4726001c746fe4a13a7fbefb` |
| `validation/outer_simulation.py` | `9760fa0eb7f3ef017eeccb148ad0d5d06dd2066e21db9bea09d34232b521a067` |
| `validation/synthetic-results-rev1.json` | `fbc4cc0de2c948574a76216cffd12cb33b488f0fedd3c3519de4ddf0b5a82277` |
| `validation/synthetic-results-rev2.json` | `24549f3ed2fcaecb357bf747ec31058ae115592f8bf63a1319763d2a84faa77d` |
| `validation/synthetic-results.json` | `e65ec36c7d57a0d15f32bc3d2d7bd5f1f76f4148189ec2676653ef996901a350` |
| `validation/outer-results.json` | `b851c2c1b6ffacee1957ade0071aa9fd38ede0dce5ebd2e3577a2967923f6352` |

Three digests carry over unchanged:
- `synthetic-results-rev1.json` equals the revision-1 evidence digest recorded in round 1.
- `synthetic-results-rev2.json` equals round 2's `validation/synthetic-results.json`.
- `round-1.md` is unchanged.

By Main's computation, the reviewed historical evidence and council record are therefore preserved byte for byte.

**Verdict: 14 issues open (0 blocking, 6 major, 8 minor). No sign-off.** No preparation, inventory, Jev, labelling or analysis stage is authorised, and ADR-014 is not adopted (§12).

**Severity scale.** Unchanged from rounds 1 and 2:
- **blocking:** can produce a wrong or unreachable verdict, or breaks experiment rules 1–7 or REQ-169.
- **major:** misleads the result, leaves a preregistered quantity undefined, or makes the study unrepeatable or likely not testable.
- **minor:** text, consistency or provenance. Minor issues must still be fixed before sign-off (rule 1).

The scale does not cover breaches of the design rules. This round classifies a design-rule-2 breach in a co-signed ADR as **major**.

---

## 1. Recommendation

**Revise to revision 4 and resubmit for X1 round 4 (ADR-014 round 2). Do not sign off.**

Revision 3 implements round 2 faithfully at protocol level: 10 of 15 issues are fixed and 5 are partly fixed. Keep all of the following:
- the keyed run SRS and finite eligible-envelope estimand;
- the stored-π audit through product `DrawAudit`;
- the McCarthy–Snowden structure, unless X1-R3-01 replaces it;
- the one-sided miss-rate gate with no waiver;
- the delivery classification and integrity precedence;
- the Bernoulli safety OC and the dev futility stop;
- the cap-order generator;
- the throughput and clock definitions;
- the ADR-014 decision table as content.

The round-2 blocking risk is now **contained**: no model call can happen until a binding outer-OC gate passes. The remaining defects concern whether that gate is executable and reachable, and they must be settled **before preparation**, because P l.171 starts the 45-day clock at the first preparation action:

1. **The pilot predicts the primary interval will fail in capped rows** (X1-R3-01). The README reports a single cell selected as the maximum and leaves out the systematic support-side asymmetry.
2. **The binding OC is not executable as written** (X1-R3-02, X1-R3-03). The cell set, populations, denominators and row timing are undefined, power and null are not defined as realised or requested gain, and the 2,000-trial missing-label gate fails 31 % of the time even for an exact interval.
3. **Reachability and compute are decided by things nobody has measured** (X1-R3-04, X1-R3-05):
   - the generator's choice of how positives are ranked below k, combined with R-052's projected M̂ ≈ 849, makes every capped row fail power;
   - missing-label compute and parallel execution are unplanned, although both can be benchmarked now.
4. **ADR-014 does not follow design rule 2** (X1-R3-06). Its rounds would sit outside its folder, its signed status text would go stale, and its index row and Appendix B card are deferred to unsigned edits.

All of the fixes are synthetic or documentary work, which is already permitted.

## 2. Alternatives considered

| Option | Description | Tradeoff | Council position |
|---|---|---|---|
| A. Sign off a scoped preparation/inventory stage now and leave the OC issues to the full-OC round (challenger's lean; X11 staged precedent) | Real model-free work starts now | Starts the 45-day clock with six major issues open and a pilot that predicts gate failure. Rule 1 requires zero issues on the scope being authorised. | Rejected |
| B. **Close all 14 issues in revision 4, including a preregistered interval diagnostic, a worst-case compute benchmark and a sealed OC manifest. Then a zero-issue round together with the concrete preparation artefacts authorises preparation.** | Viability is established before the clock starts | Revision 4 may replace the primary interval (X1-R3-01c), which needs fresh validation | **Recommended** |
| C. Replace the interval with a studentised one now, without a diagnostic | Pre-empts the expected failure | The change would not rest on evidence, and the diagnostic is cheap. A preregistered decision rule removes researcher choice. | Rejected; the studentised interval is the preregistered branch of B |
| D. Start the 45-day clock at the first dev model call instead of at preparation | Separates OC compute from labelling time | Hides the elapsed-effort and staleness cost that the clock exists to bound, and does not fix the method risk | Not adopted |
| E. Override the product cap for the study, e.g. 2,000 calls with no cost cap | Removes cap truncation | The estimand stops being the deployed default treatment, which reverses the X1-R1-08 decision | Rejected |

## 3. Rule checks

- **Rules 1–2.** This round is recorded with digests. Issues are open, so iteration is required.
- **Rule 3.** No predictor–outcome association was computed. The scripts accept only synthetic frames, and the response reports no real-data runs.
- **Rule 4.**
  - The kernel recovers a planted effect: self-test support, and 100/100 support in the 16 uncapped, no-failure requested-.20 cells.
  - It reports a planted null as not supported, with non-zero width: [−.0228, .0281].
  - The noisy null is a census, so the null with unequal π is still missing (X1-R3-08).
  - Product-backed validation is correctly deferred, but its acceptance criteria are undefined (X1-R3-11).
- **Rule 5.** Nothing is signed yet. Revision-1 and revision-2 evidence is preserved, and the failed first noisy-null attempt is disclosed in the manifest and README.
- **Rule 6.**
  - Pilot outcomes carry Wilson intervals.
  - The selective reporting of the pilot maximum is corrected by X1-R3-01(a).
  - The empirical-Bernstein bound is not a valid cited bound (X1-R3-10).
- **Model use and secrets.**
  - Labellers use `claude -p` and `codex exec`; Jev uses the Infisical-injected key only (P l.169).
  - The product reports a missing key per unit as a network failure. Classification is handled in X1-R3-09.
- **Design rules.**
  - ADR-014 breaches rule 2's record location and freezes stale outcome text (X1-R3-06).
  - Its ID allocation is sound (§6).
  - Rule 3 does not apply, because ADR-014 cites no experiment results.
- **Ledgers.**
  - HD9, R-071 and R-072 are present.
  - R-064 diverges from the row landing with ADR-011 and would duplicate it. R-067's link does not resolve on this branch (X1-R3-12).

## 4. Verification of round-2 fixes

| Issue | Status | Evidence |
|---|---|---|
| X1-R2-01 (blocking) | **Partly fixed**; the blocking risk is contained | **Fixed:**<br>• (a) keyed SRS (P l.39; `ranking.py` `select_runs`);<br>• (b) m = G − 1, no FPC (P l.101; `outer_simulation.py:43`);<br>• (e) sequencing (P l.179; `reviews/README.md`);<br>• (g) deterministic census cells relabelled (P §7; README);<br>• (h) R-071/R-072 cited at P l.101.<br>**Residuals:**<br>• (c) one synthetic population per cell (`outer_simulation.py:49`);<br>• (d) the gate's cell set, timing and denominators are undefined, and the 2,000-trial gate is mis-calibrated;<br>• (f) missing-label compute is unmeasured.<br>These are carried as X1-R3-01 to X1-R3-05. No model call is possible without the binding gate. |
| X1-R2-02 | **Partly fixed** | Retrieval keys are shared with probability ρ, otherwise independent (`ranking.py:326–329`). Null interval [−.0228, .0281]. Arm-swap, unpaired, unit-level and π mutations are detected. The noisy null uses 50-unit uncapped runs, so every π = 1: it is a census null, and no unequal-π null exists (X1-R3-08). |
| X1-R2-03 | Fixed at protocol level | Bernoulli grids (P l.139; `outer_simulation.py` safety loop). Severity-critical counts reported at inventory (P l.41). Futility stop (P l.95). The joint canonical-label safety OC is still unspecified (X1-R3-02). |
| X1-R2-04 | Fixed | Classes, denominator and precedence (P l.62–68, l.117). Per-run reservation accounting with no global latch (P l.161). `delivery_integrity` with fixtures. Product failure kinds are not mapped individually (X1-R3-09). |
| X1-R2-05 | Fixed as specified | Baseline head M minus failures (`ranking.py:338–347`), matching product `run.go:196–212`. M̂ rule (P l.163). Inventory cap-tail shares (P l.41). Reachability is carried as X1-R3-04. |
| X1-R2-06 | Fixed | Attempt reservation, ≥ 200 dev cases, projection, quota pause and clocks (P l.165–171). The bound's constants are wrong (X1-R3-10). |
| X1-R2-07 | Fixed | The selected-fraction mutation is 5/7, distinct from 1 and from π. Differing-run bootstrap oracle. 64 all-d, unequal-π, heterogeneous completions with excess 0. Zero-denominator and > 5 % inadequacy tests. `zero_tail` guard. The stale comment is removed. |
| X1-R2-08 | Fixed | Wilson intervals for support, veto and unmeasured, and miss rates in every scenario. Script SHA-256s, Python 3.14.10 and elapsed time embedded in both JSON files. |
| X1-R2-09 | Fixed | Siding denominators, the n ≥ 20 trigger and the two permitted outcomes (P l.91). Zero escalations are infeasible (P l.153, l.157). |
| X1-R2-10 | Fixed | `cannot_tell` is unknown Y (P l.93). Percentiles use the retained B′ (P l.105; `ranking.py:143–146`). |
| X1-R2-11 | Fixed at protocol and kernel level | Pending-only queue (P l.58; `pending_uniforms`), matching `audit.go`'s skip of stored draws. The full product crash-resume remains gated (X1-R3-11). |
| X1-R2-12 | Fixed | Scope limited to k = 30 % (P l.7, l.51). Descriptive competing-explanation lines (P l.109). |
| X1-R2-13 | **Partly fixed** | HD9 history updated. R-071 and R-072 present. R-067's digests check out against X11 (`scout-x11/.../reviews/rev4-digests.json:8,27`). Two residuals: R-064 diverges from the ADR-011 landing row, and R-067's path does not resolve (X1-R3-12). |
| X1-R2-14 | **Partly fixed** | ADR-014 captures every choice, with options, rationale and source round, including the round-2 choices. Its governance format is non-compliant (X1-R3-06), and the README duplicates its table divergently (X1-R3-14). |
| X1-R2-15 | **Partly fixed** | The Sonnet 5.5 June 2026 cutoff is recorded (P l.23; R-068). Fused tokens and an editing instruction remain (X1-R3-13). |

## 5. New issues

### X1-R3-01: major. The pilot shows a systematic support-side miss excess that the README understates. Diagnose the primary interval, and confirm or replace it, before the 45-day clock starts

- **Location:** P l.101 (primary interval), l.133 (outer design), l.171 (clock start); README l.34; the response's "Exercised synthetic evidence"; `outer_simulation.py:41–43, 49–55, 66–86`; `outer-results.json`.
- **Problem:**
  - **The selected maximum alone is not conclusive.**
    - The README and response present one cell: 8/100, Wilson [.041, .150].
    - It is the maximum over 1,152 tail tests (576 retrieval cells with scanner-critical controls false, × 2 tails).
    - P(Bin(100, .025) ≥ 8) ≈ .0035, so about 4 such tails are expected under exact nominal coverage [INFERENCE: arithmetic].
  - **The real signal is structural, and it is not reported.** jq over the 576 scanner-false retrieval cells:

    | Rows | LB > truth (support side) | UB < truth |
    |---|---|---|
    | All | 892 of 57,600 (.0155) | 296 (.0051) |
    | M = 850 | 470/19,200 (.0245) | 27/19,200 (.0014) |
    | M = 850, gain SD 0 | 302/9,600 (.0315) | 8/9,600 |
    | M = 2,000 | 221/19,200 | 206/19,200 |

    The peer reports .0282 / .0247 for the M = 2,000, failure .02, SD 0 block across both scanner populations.
  - **These rates appear despite a built-in cushion.**
    - The pilot frame has L = 40 runs per repository and samples 20 (q = .5).
    - The m = G − 1 bootstrap without FPC then overstates the between-run variance by about 1/(1 − q) = 2 [INFERENCE: standard SRS-without-replacement variance].
    - With that cushion, a centred, approximately normal interval would miss each tail only about .003 of the time.
    - The locked frames' q is unknown and may be much smaller, which removes the cushion.
  - **Leading explanation** (peer H1; the lead concurs) [INFERENCE]:
    - Under the cap, per-run gain depends on run size, and the equal-repository ratio is dominated by the largest runs.
    - A sample with few capped large runs gives a high D̂ **and** a narrow bootstrap distribution, so LB lands above truth.
    - Percentile noise at B = 199 cannot produce a one-sided pattern.
    - The vertex envelope is absent from the pilot.
    - The truth definition matches the estimator (`outer_simulation.py:55`; `ranking.py:359–375`).
  - **Seeds are shared across cells, and this is not disclosed.** The trial RNG seed is common to every cell (`outer_simulation.py:67`). The first repository's SRS is therefore identical in every cell for a given trial (`:71`), so pooled rates overstate precision.
  - **Consequence** [INFERENCE: binomial arithmetic, peer]:
    - At a true one-sided rate of .03, one cell passes the ≤ .035 gate with probability ≈ .79 at 10,000 trials.
    - With about 190 cells near .03, the M = 850 row almost surely fails.
    - Proceeding to preparation and inventory with this method spends the clock on a predicted "not testable".
- **Required fix** (synthetic work, already permitted):
  - (a) **Reporting.** Replace the "worst cell" paragraph in the README and response with:
    - pooled one-sided miss rates per M row × gain SD × failure rate, with Wilson intervals labelled as correlated by shared seeds;
    - the expected maximum under nominal coverage;
    - capped-row realised gains and power (X1-R3-04).
  - (b) **Diagnostic run.**
    - Create-once output with embedded digests; independent per-cell seeds; B = 1,999; ≥ 2,000 trials per cell.
    - Cover the M = 850 block, the M = 2,000/failure .02 block, and a no-cap scenario in which run gain depends on run size.
    - Use frame sizes L ∈ {40, 100, 400} runs per repository.
    - Emit per cell: mean and median of D̂ − truth; SD(D̂); mean bootstrap SD; bootstrap skewness; capped runs sampled per trial; both one-sided miss rates.
  - (c) **Decision rule, fixed now.**
    - **If** in the L = 400 diagnostic any block's pooled one-sided miss rate has Wilson MC95 lower bound > .025: revision 4 replaces the primary with the council-recommended alternative.
      - The alternative is a studentised (percentile-t) paired run-ratio bootstrap: run-linearised ratio variance per repository (residuals (J_i − B_i) − D̂_r T_i), no FPC, m = G − 1 resampling, the equal-repository combination, retained B′, the same vertex envelope and common draws.
      - Specify it fully, with hand fixtures, then rerun the diagnostic on fresh seeds.
    - **Otherwise** retain McCarthy–Snowden and record the diagnostic as the reason.
    - Either way there is no automatic fallback, and the post-inventory binding gate applies unchanged.
  - (d) Add a run-size × gain dependence factor to the full OC grid.

### X1-R3-02: major. The binding OC contract does not define its cells, populations, denominators, row timing, or the meaning of "power" and "null"

- **Location:** P l.41 ("every registered synthetic reference cell"), l.133, l.135, l.139, l.141, l.143; `outer_simulation.py:49–50, 82–86, 87–100, 115–116`.
- **Problem:**
  - **Cell set.**
    - "Every required safety-clear reference cell" is undefined.
    - Coverage depends only on retrieval: miss counts are identical across the 24 safety settings, and every histogram bin is a multiple of 24 (peer). "Safety-clear" therefore selects nothing for coverage.
    - Scanner-critical-true cells are simply a second population draw.
  - **Row timing.** The gate applies "before any model call" (l.141), but the cap row is selected after dev calls (l.143). The protocol does not say whether every row must pass before dev, or only the selected row before test.
  - **Populations.** There is one synthetic population per cell (`seed = 900`, l.49).
  - **Denominators.**
    - No rule covers outer draws that fail the sampling gates (N_r ≥ 1,000 and ΣN ≤ 60,000, neither implemented in the pilot) or that fail adequacy.
    - The pilot counts trials with no point estimate as non-misses in the `trials` denominator (l.83–86).
    - Its flag requires `inadequate == 0` (l.115). At locked sizes that could make the gate unreachable whatever the coverage.
  - **Missing-label cells.**
    - Only a trial count is given. There is no unknown-share grid, mechanism, or dependence on d, run or rank, and no mapping from invalid sessions and canonical criticality to unknown Y.
    - The pilot draws unresolved status over **all** below-floor units, instead of generating canonical labels and applying `safety()` (l.87–100).
  - **Power and null semantics.** ".20-gain" and "null" do not say whether they mean requested latent gain or realised population gain. For requested .20, realised gain is .003–.045 at M = 400 and .039–.108 at M = 850 (jq).
- **Required fix:** add a sealed, machine-readable `oc-manifest.json` to the signed set, defining the following (recommended choices):
  - **Coverage cells:** every retrieval cell in every M row, i.e. prevalence × gain SD × requested gain × ρ × failure × scanner-control population × the rank-structure level of X1-R3-04 × the size-dependence factor of X1-R3-01(d). Safety settings are excluded from coverage.
  - **Populations:** five independent synthetic populations per cell at locked sizes, with 2,000 eligible trials each (10,000 per cell). Gate on the pooled cell and report per-population rates.
  - **Eligible trials:**
    - Draws that fail a sampling gate or adequacy are counted and excluded from miss denominators.
    - Trials continue along the predeclared seed sequence until the required eligible count is reached, capped at twice that count.
    - A cell that reaches the cap is unmeasured, which means not testable.
    - Report P(gate failure) and P(inadequate) with Wilson intervals. In a power cell, a Wilson upper bound of P(inadequate | gates pass) above .05 fails that cell.
  - **Row timing:**
    - Every row is computed before any dev model call. If no row passes both coverage and power, stop before dev (not testable).
    - After dev, the row selected by M̂ must already have passed, otherwise no test calls.
  - **Power:** P(support) Wilson lower bound ≥ .80 at **realised** population gain ≥ .20, with the latent location calibrated without outcomes on each synthetic population. A row where realised .20 cannot be reached is power-infeasible and fails.
  - **Null:** realised population gain ≤ 0, with false-support Wilson upper bound ≤ .05. Requested-latent cells stay as descriptive reachability rows.
  - **Missing-label cells:** as in X1-R3-03.
  - **Final safety OC:** generate canonical truth and criticality states, including invalid sessions and `cannot_tell`, and apply `safety()`.

### X1-R3-03: major. The 2,000-trial missing-label gate fails 31 % of the time even for an exact interval

- **Location:** P l.141 ("at least 2,000 trials, same one-sided gate"); the ADR-014 coverage row (ADR l.30).
- **Problem:**
  - A Wilson upper bound ≤ .035 at 2,000 trials needs about ≤ 53 misses.
  - At an exactly nominal .025 per tail, P(pass) ≈ .69 for one tail [INFERENCE: binomial arithmetic, peer].
  - Across several cells and two tails, failure is near-certain even for a perfect interval, so the gate cannot be reached by design.
- **Required fix:**
  - Missing-label gate cells use ≥ 10,000 eligible trials with the same ≤ .035 rule.
  - Seal the missing-label grid in the manifest:
    - weighted unknown share {.01, .05};
    - mechanism {MCAR; concentrated in d = +1; concentrated in d = −1};
    - applied to each M row × prevalence {.10, .55} × realised gain {0, .10, .20}, at ρ = .5 and gain SD .10;
    - truth is the full-label population gain.
  - Revision 4 may propose a smaller named subset, justified by the X1-R3-05 benchmark, for review before sealing. If the sealed grid cannot fit, the outcome is not testable; the grid is never shrunk silently.

### X1-R3-04: major. Capped-row reachability is undisclosed and rests on generator assumptions that are not registered

- **Location:** P l.135, l.137, l.143, l.163; README l.34; `ranking.py:326–347`; R-052.
- **Problem:**
  - **Power fails in every capped row.** For requested .20: support is 0/100 in every M = 400 cell and at most 87/100 at M = 850 (jq). The README reports only the 16 uncapped cells.
  - **The projected M̂ selects the worst row.** R-052's ≈ $0.000589 per request gives M̂ ≈ 849 < 850, so the rule selects the M = 400 row. A one-sided upper bound c_U lowers M̂ further [INFERENCE: the triage price is unmeasured].
  - **Generator assumptions drive the capped-row results.**
    - The baseline order below rank k is uniformly random (`base_rest` shuffled).
    - Jev's preferred positives are drawn over **all** positives, independent of the baseline head except through ρ (`:327–329, :344`).
    - Together these maximise the damage the cap does, and neither is registered as an assumption with an alternative.
  - **The grid is coarse.** {2,000, 850, 400} rounds M̂ down by up to 53 %.
- **Required fix:**
  - Register a rank-structure factor with two levels:
    - (i) the current uniform order below k;
    - (ii) a sealed gradient in which positive density below k declines linearly with baseline rank.
  - The coverage gate applies to both levels. Power is judged at realised gain (X1-R3-02).
  - Refine the M grid to {2,000, 1,500, 1,000, 850, 700, 550, 400}. The inventory reports units beyond each grid M.
  - Disclose capped-row power, realised gains and the R-052-based M̂ projection in the README and response.

### X1-R3-05: major. Compute feasibility and parallel execution are unplanned, although both can be measured now

- **Location:** P l.145, l.169–171, l.179; README l.36.
- **Problem:**
  - **The complete-label projection is already large.** It is 8.96 days serial, or 17.9 days with the 2× overhead, for all 1,152 cells [README; INFERENCE]. X1-R3-02 to X1-R3-04 enlarge the grid.
  - **Missing-label cost is unmeasured.** Each trial evaluates up to 8 vertex bootstraps per repository and up to 512 combination distributions of B′ values (`ranking.py` `analyse`). That is plausibly tens of times a complete-label trial [INFERENCE: unmeasured].
  - **No parallel-execution rule exists.** The OC nevertheless runs inside the 45-day clock, which must also cover dev and test labelling.
  - **Feasibility is checked too late.** It is checked only "before full authority", after preparation and inventory. Worst-case sizes are bounded in advance (20 runs × ≤ 2,000 units per repository; ΣN ≤ 60,000), so it can be measured now.
- **Required fix:**
  - Permit process-level parallel runs of the identical reviewed kernel. Use deterministic per-cell seeds, and add a fixture showing byte-identical output whatever the worker count.
  - In revision 4, benchmark on worst-case synthetic frames, at L ∈ {40, 400}, the per-trial cost of complete-label and missing-label cells at B = 1,999.
  - Project core-hours for the sealed manifest, and declare the cores used and a binding core-hour and wall cap (precedent: X11's R-075). The cap must fit inside the 45-day clock while leaving the throughput reservation.
  - If the projection does not fit, revise before preparation. Re-measure at locked sizes before full authority, as now.

### X1-R3-06: major. ADR-014 does not follow design rule 2 and would freeze stale outcome text

- **Location:**
  - ADR l.3–6 (status and "Council outcome: pending"), l.10 and l.53 (R-064 claim), l.57 (review record named in the experiment folder);
  - README l.44, l.46 ("No ADR index/design-card rows are added before the PR lands").
  - Compare: ADR-010's "Review" and "Records in this folder" sections (`docs/adrs/2026-11-08-d1-jev-minimums/ADR.md`); the ADR-011/012 status lines; `docs/adrs/README.md`; the Appendix B cards in `docs/scout-design.html` (ADR-010 card at l.971–975).
- **Problem:**
  - **Record location.**
    - Design rule 2 places the council rounds in the ADR's own folder, and a signed `ADR.md` names that folder's final round.
    - ADR-014's folder holds only `ADR.md`, and it names `evaluation/x1-triage-ranking/reviews/round-N.md`.
  - **Stale status text.**
    - "Accepted — agent decision …; owner review pending" and "Council outcome: pending" are in the signed set, so once signed they become permanently stale.
    - "Accepted" before a zero-issue round contradicts `docs/adrs/README.md` ("adopted only after an innovation-council round with zero issues"), unless it is written in ADR-010's or ADR-011's adopted form.
  - **Deferred index row and card.** Every ADR has a `docs/adrs/README.md` row and an Appendix B card. Deferring both until "the PR lands" means unsigned later edits to the design document.
  - **False R-064 claim.** "Byte-identical research-ledger row originating in ADR-011/012" is no longer true (X1-R3-12).
  - **Scope (not an issue).** ADR-014 is not a substantive product-design change: REQ-169 delegates these choices, and ADR l.10–12 excludes product adoption. The council accepts an ADR as the R-064 capture mechanism, but the design-rule format then applies.
- **Required fix** (recommended: route (a), as used for ADR-010):
  - Put the status line in adopted form: "accepted; decided by the agent under owner delegation (R-064)", as ADR-011/012 do.
  - Delete "Council outcome: pending" and all other pending prose.
  - Write the Review section as: "This summary is among the signed files … Adopted on the explicit zero-issue final `round-<N>.md` in this folder." Add a "Records in this folder" section. Owner review stays outside the signed narrative; R-064 already records it.
  - Save each joint council record also as `docs/adrs/2026-11-08-x1-delegated-decisions/round-<N>.md`. **This record is ADR-014 `round-1.md`.**
  - Add the `docs/adrs/README.md` index row and an Appendix B card for ADR-014 (accepted; protocol-capture scope; no relation) to the signed set.
  - Correct the R-064 sentence as in X1-R3-12.

### X1-R3-07: minor. The per-cell `binding_gate_established` flag can claim a gate it does not check

- **Location:** `outer_simulation.py:115–116`; manifest `execution.binding_gate_established`.
- **Problem:**
  - The flag checks only trial count, `inadequate == 0` and the Wilson upper bounds.
  - It ignores B = 1,999, locked-inventory provenance and manifest membership.
  - It is emitted on safety-adverse rows, and its inadequacy rule differs from P.
  - A planning run with `--trials 10000` could print `true`.
  - All 27,648 flags are currently false, consistent with the README and the manifest, so nothing has been falsely approved.
- **Required fix:**
  - Planning outputs emit only a numerical `tail_criterion_met_at_pilot_precision` and a `fails_at_pilot_precision` (Wilson lower bound > .035).
  - Only a separate, reviewed final evaluator may emit `binding_gate_established`. It must check exact B, manifest completeness and membership, the eligible-trial counts of X1-R3-02, and input provenance, and it must fail closed.

### X1-R3-08: minor. The noisy null does not exercise unequal audit probabilities

- **Location:** `self_test.py:42–48`; `ranking.py:299–301` (defaults `sizes=(50,)`, `cap=2000`).
- **Problem:**
  - With 50-unit uncapped runs, π = min(1, 50/(2·25)) = 1 for every unit, so the noisy null is a census.
  - The challenger confirmed in `planted_null` that every unit is selected and the conditional interval is degenerate.
  - Round 2 required the null to vary across runs **with unequal π**.
- **Required fix:**
  - Add a noisy null fixture with heterogeneous sizes, a binding cap and π < 1, with known population truth.
  - Assert both discordance directions and a non-zero-width interval containing 0.
  - Evaluate the ignored-π and nominal-π mutations on it.
  - Keep the census null as a separate fixture.

### X1-R3-09: minor. Delivery classification does not map every product failure kind

- **Location:** P l.64–66; `ranking.py:275–294`; product `internal/decisions/client.go:29–36, 143–144` and `run.go:199–212, 246–249, 272–273`.
- **Problem:**
  - The product emits:
    - failure kinds `network`, `timeout`, `http`, `non-retryable`, `rate-limited`, `schema-invalid`, `model-mismatch` and `unsafe-to-persist`;
    - capped reasons "call cap reached", "billed cost unknown" and "cost cap reached";
    - the sink failure "annotations not persisted".
  - P says only "transport/mismatch".
  - A missing key appears per unit as `FailNetwork` "no key" (`client.go:143–144`). A harness would count it as product transport unless it detects run-level key absence first.
- **Required fix:** add a mapping table, and update the kernel reasons and fixtures to match:
  - the .10 numerator is {network, timeout, http, non-retryable, rate-limited, schema-invalid, model-mismatch, unsafe-to-persist};
  - "annotations not persisted" and run-level key absence are harness integrity failures;
  - "billed cost unknown" capped units are product treatment and are excluded from the attempted denominator.

### X1-R3-10: minor. The empirical-Bernstein U95 is not a valid cited bound

- **Location:** P l.167.
- **Problem:**
  - The stated bound, mean + sqrt(2s² ln 20/n) + 3·900·ln 20/n, matches no published bound.
  - Maurer and Pontil (2009), Theorem 4, gives mean + sqrt(2V_n ln(2/δ)/n) + 7b ln(2/δ)/(3(n − 1)), where V_n is the Bessel-corrected sample variance. At δ = .05 that is ln 40 ([COLT 2009 paper](https://www.learningtheory.org/colt2009/papers/012.pdf); [arXiv:0907.3740](https://arxiv.org/abs/0907.3740)).
  - No R-ID is given.
- **Required fix:**
  - Adopt Maurer–Pontil Theorem 4, one-sided, with δ = .05 and b = 900, stated exactly.
  - Add a research-ledger row with a new R-ID and cite it at P l.167.

### X1-R3-11: minor. Product-backed rule-4 validation has no acceptance criteria

- **Location:** P l.147, l.179; `reviews/README.md`.
- **Required fix:**
  - **Named synthetic frames:** capped; whole-run fallback with explicit `DrawAudit`; mid-run crash-resume; unequal π; unknown-cost stop; one frame per delivery reason.
  - **Parity:** stored π, draws, point, interval and verdict equal the kernel's exactly (≤ 1e-12 for floats).
  - **Recovery:** at locked sizes on listed seeds, the planted effect gives support, and the noisy unequal-π null gives not supported with non-zero width.
  - Give a compute projection. Failure means no full authority.

### X1-R3-12: minor. Ledger integrity: R-064 diverges from its landing row, and R-067's link does not resolve

- **Location:** `docs/research-ledger.md:72, 75`; P l.11; ADR l.10, l.53; README l.14, l.44, ledger section.
- **Problem:**
  - **R-064 differs from the row that will land.**
    - X1's R-064 matches `scout-x11/docs/research-ledger.md:71`, but not the row landing with ADR-011, `scout-adr11b/docs/research-ledger.md:73`.
    - That row differs in two cells. Informed: "ADR-011 Gitleaks version contract; delegated design decisions". Caveat: "the owner may review later, recorded as a new `user` row".
    - adr11b's ledger header (l.4) says: "R-064 lands with ADR-011 first; other work reuses that direction rather than appending a duplicate."
  - **The sibling ledger moved during this round.** At the start of this round, `scout-adr11/docs/research-ledger.md:72` held an R-064 identical to X1's. Later in the round that file no longer held R-064; l.72 is now R-070.
  - **R-067's link does not resolve.**
    - It cites `evaluation/x11-merge-association/reviews/rev4-digests.json`, which is absent from this worktree. The research-ledger skill requires links to resolve.
    - The digests themselves are verified in `scout-x11/.../reviews/rev4-digests.json:8,27` and X11 `round-4.md:121,132`.
  - **The in-place edits are accepted.** R-067 and R-068 were corrected in place. This overrides round 2's append-only instruction and the peer's proposal:
    - the rows were never committed to any branch, so there is no ledger history to preserve;
    - restoring known-wrong bytes only to supersede them would degrade the permanent ledger;
    - append-only binds from first commit.
- **Required fix:**
  - **R-064.** Remove X1's R-064 row from X1's diff, and rebase onto the commit that lands ADR-011's R-064 **before the signing round**. Take the signed ledger digests after the rebase. Replace the "copied identically / byte-identical … ADR-011/012" wording (P l.11; ADR l.10, l.53; README l.14, l.44; response) with "R-064 as landed with ADR-011".
  - **R-067** (recommended). Copy X11's `reviews/rev4-digests.json` byte for byte to `inputs/x11-rev4-digests.json`, add it to the signed set with its SHA-256, and point R-067 at it.
  - Add the Maurer–Pontil row (X1-R3-10).

### X1-R3-13: minor. Legibility is still not restored, and the frozen text contains an editing instruction

- **Location:**
  - **P:** l.141 "for10k", "for2k", "Delete all discretionary … rules"; l.151 "threshold0.30"; l.153 "grid0"; l.161 "USD31"; l.163 "one-sided95%", "seed169300"; l.167 "timeout900s", "bound900", "at900", "At least200"; l.171 "former48h"; l.179 "rule4"; l.180 "at least200".
  - **README table:** "first20", "First20", "SRS20", "k20%", "k30%", "fraction1", "M2000/850/400", "failures0/.02".
  - **ADR-014:** "k20%", "k30%", "Fraction1", "First20", "SRS20", "Product2000/USD.50", "USD31", "At least200", "continues;45 days", "inventory48h", "k.20/.40".
- **Required fix:**
  - Do a full spacing pass on all three files.
  - Restate P l.141 normatively: "No discretionary coverage-review rule exists."
  - Leave the fixed council rounds untouched.

### X1-R3-14: minor. Two divergent decision tables

- **Location:** README l.48–69 versus ADR l.16–49. Example: README l.51 gives the run-selection options as "First20; SRS; redraw until gates pass", while ADR-014 gives "First20 consecutive; SRS; superpopulation-only model".
- **Required fix:** make ADR-014 the single source. The README links to it and keeps no table.

## 6. Assessment of the targeted questions

- **Adverse pilot coverage.**
  - The worst cell (8/100, MC95 [.041, .150]) is reported correctly, but on its own it is selection-biased and noise-consistent (about 4 expected under nominal coverage).
  - The evidence that counts is the support-side asymmetry concentrated in capped rows: .0315 versus .0008 at M = 850 with gain SD 0. It arises despite a twofold variance cushion from q = .5.
  - It is adverse. It is not yet proof that the final interval fails at locked sizes, and X1-R3-01 decides it before preparation.
- **Binding-gate flags.**
  - All 27,648 are false, consistent with the manifest and README.
  - The flag's semantics would let a planning run claim the binding gate (X1-R3-07).
  - The gate itself has no defined cell set or denominators (X1-R3-02), and its missing-label variant is mis-calibrated (X1-R3-03).
- **Locked-inventory OC.**
  - The sequence (inventory → full OC → dev → test) is correct, and the outer design targets population truth with a real run SRS.
  - The contract is not executable (X1-R3-02, X1-R3-03), reachability is hostage to the generator (X1-R3-04), and compute is unplanned (X1-R3-05).
- **Product-backed implementation.** Correctly scoped and sequenced as a prerequisite for full authority, but it has no acceptance criteria (X1-R3-11) and no failure-kind mapping (X1-R3-09).
- **ADR-014 content.**
  - The table is complete and consistent with P. Spot checks agreed on: verdict order; delivery classes and the .10 gate; per-run reservation with no global latch; Bernoulli safety and the futility rule; throughput and clocks; stage order.
  - The README correctly separates protocol capture from later product ADRs.
  - **The ID allocation is sound**, overriding the challenger's "unverified":
    - ADR-011 is in `scout-adr11b`, ADR-012 in `scout-adr11`;
    - ADR-013 is cited by X11a (`scout-x11/evaluation/x11-merge-association/PREREGISTRATION.md:10`; research-ledger R-073 to R-076);
    - no competing ADR-014 exists in any `scout*` worktree.
  - The format defects are in X1-R3-06.

## 7. Checked and sound

- **`select_runs`** matches P l.39: key `169-x1-runs`, message repository NUL SHA, ascending digest with SHA tie-break, ≥ 20 runs, duplicates refused. A fixture checks order independence.
- **Outer pilot structure** (`outer_simulation.py:43, 50, 55, 71`):
  - a fresh SRS of 20 per trial, drawn from a finite population;
  - inner binomial audit draws within equal-π groups, equivalent to per-unit Bernoulli draws;
  - m = 19 bootstrap streams;
  - truth is the population equal-repository ratio at π = 1.
- **Cap order** matches product `run.go:196–212`: baseline-ordered units, `index >= cap`, and a per-run stop on unknown cost (`blind`).
- **Verdict precedence** (`verdict()`) matches P l.117–123, with fixtures for integrity over veto and for retrieval negative over safety unknown.
- **Futility M = Σ floor(N_run/2)** equals the below-floor count, since `below` means jev > (n + 1) // 2.
- **The pending-only HMAC queue** matches the stored-draw skip in `audit.go`.
- **Manifest figures match JSON:** 27,648 scenarios, 100 trials, B = 199, 770.455… s, every flag false, maximum miss 8. The README's worst-cell parameters, its realised gain of −.039269… and its 16 cells at 100/100 all match jq output.
- **Ledgers and references:**
  - R-067's digests match X11's preserved revision-4 manifest.
  - R-071 and R-072 are present and cited; HD9 is proposed, with its history line.
  - The Sonnet 5.5 cutoff in P l.23 agrees with R-068.

## 8. Evidence index

- **Pilot jq, 576 scanner-false retrieval cells:**
  - lower-tail miss histogram {0: 183, 1: 146, 2: 112, 3: 59, 4: 47, 5: 20, 6: 7, 7: 1, 8: 1};
  - upper-tail miss histogram {0: 404, 1: 109, 2: 28, 3: 18, 4: 11, 5: 3, 6: 3};
  - lower/upper misses by row: M = 400 201/63, M = 850 470/27, M = 2,000 221/206; M = 850 with SD 0 302/8 (per 9,600 trials);
  - realised gain for requested .20: M = 400 .003–.045, M = 850 .039–.108, M = 2,000 .168–.215;
  - support at requested .20: M = 400 always 0; M = 850 at most 87/100; M = 2,000 always 100/100.
- **Peer jq, both scanner populations:** M = 850 with SD 0 lower 587/19,200 (.0306); M = 2,000/failure .02/SD 0 .0282 / .0247. Miss counts are identical across the 24 safety settings.
- **Gate arithmetic (peer):** a pass needs about ≤ 313/10,000 or ≤ 53/2,000 misses. P(pass) at 10,000 trials: .785 at .030, .36 at .032. P(pass) at 2,000 trials and nominal .025: .69.
- **Product:**
  - failure kinds `internal/decisions/client.go:29–36`; missing key reported as a network failure at `:143–144`;
  - cap and blind reasons `run.go:199–212`; mismatch and sink failures `run.go:246–249, 272–273`.
- **Appendix B cards:** `docs/scout-design.html:971–975` (ADR-010). ADR index: `docs/adrs/README.md`.
- **R-064 variants:**
  - X1 `docs/research-ledger.md:72`;
  - `scout-x11/docs/research-ledger.md:71`;
  - `scout-adr11b/docs/research-ledger.md:73` (canonical landing; header at l.4);
  - `scout-adr11/docs/research-ledger.md:72` held R-064 at the start of this round and held R-070 later in the round.
- **ADR-013 allocation:** `scout-x11/evaluation/x11-merge-association/PREREGISTRATION.md:10`; `scout-x11/docs/research-ledger.md:72–75`.
- **Maurer and Pontil (2009), Theorem 4:** https://www.learningtheory.org/colt2009/papers/012.pdf; https://arxiv.org/abs/0907.3740.

## 9. Risks

- **The interval gate may fail even after amendment.** The studentised alternative is not guaranteed to pass, and X1 would then be not testable before models. This is reversible only through a new protocol and is decided pre-data.
- **The product cap may make X1 untestable at product defaults.** If M̂ is small, little reorderable mass exists in large runs. That is a legitimate product finding: it points to the same recommendation as inconclusive (switch to the deterministic default). The refined grid and rank-structure factor keep the generator from deciding it alone.
- **Compute and clock.** A full OC with missing-label cells may not fit alongside near-census labelling within 45 days. X1-R3-05 decides this before the clock starts.
- **Sibling instability.** The ADR-011/012 branches and ledgers changed during this review, so the signed ledger digests must be taken after rebasing onto landed ADR-011.
- **Shared-seed pilot evidence.** Pooled rates overstate precision. The diagnostic must use independent per-cell seeds.
- **Reversibility.** Every fix is pre-data and synthetic. No product, default, design document or accepted ADR changes.

## 10. Where the peers changed the council's recommendation

- **Challenger:**
  - Raised the binding-flag semantics. **Adopted** as X1-R3-07, downgraded from major to minor because no flag is currently true and the final evaluator needs its own review.
  - Raised the census noisy null. **Adopted** (X1-R3-08).
  - Raised the ADR-014 record location and stale pending text. **Adopted** (X1-R3-06); the lead added the missing index row and Appendix B card.
  - Proposed the studentised alternative. **Adopted** as the preregistered branch of X1-R3-01(c).
  - Its "ADR allocation unverified" finding is **overridden** by the ADR-013 evidence (§6).
  - Its lean towards a scoped preparation/inventory approval before the OC repairs is **overridden**: the clock starts at preparation (Alternatives, option A).
- **Experiment peer:**
  - Found the missing-label gate mis-calibration (X1-R3-03), which the lead missed.
  - Found the product failure-kind mapping (X1-R3-09), the Maurer–Pontil constants (X1-R3-10), the missing acceptance criteria (X1-R3-11) and the divergent tables (X1-R3-14).
  - Found the adr11b R-064 divergence. The lead had checked only `scout-adr11`, which then changed during the round.
  - Contributed the hypothesis table (H1–H7), the R-052 projection M̂ ≈ 849 → M = 400 row, the refined M grid, the five-population design, pre-dev row gating and parallel execution.
  - Its "restore and append R-067/R-068" is **overridden**: never-committed rows may be corrected in place.
  - Its "miss rates over adequate trials" is **refined** into the eligible-trial rule with a twice-count cap (X1-R3-02).
- **Found by the lead only:**
  - the FPC-cushion argument showing that the pilot's support-side asymmetry is real signal despite q = .5;
  - shared-seed correlation via the first repository's identical SRS;
  - defining power and null on realised gain, with outcome-free calibration;
  - the per-unit "no key" network failure;
  - the frozen copy for R-067;
  - measuring compute feasibility now, before preparation;
  - the major classification of the design-rule-2 breach.

## 11. Execution outline for revision 4 (one integration owner)

1. **Interval diagnostic** (X1-R3-01). Implement the (b) run, apply the (c) rule, and update reporting as in (a).
   *Test:* create-once JSON with per-cell bias, SD, bootstrap SD, skewness and capped-run counts. If the studentised branch triggers, add hand fixtures and a fresh-seed rerun.
2. **OC manifest and generator** (X1-R3-02 to X1-R3-04, X1-R3-01(d)):
   - `oc-manifest.json`;
   - rank-structure and size-dependence factors;
   - the refined M grid;
   - realised-gain calibration;
   - five populations per cell;
   - the eligible-trial rule;
   - the missing-label grid at ≥ 10,000 trials;
   - canonical-label safety generation;
   - pre-dev and post-dev row gating.

   *Test:* the manifest enumerates every gate cell, and the generator refuses unregistered factor levels.
3. **Compute benchmark** (X1-R3-05): worst-case complete-label and missing-label per-trial cost at B = 1,999; core-hour projection; parallel-determinism fixture; declared cap.
4. **Kernel and fixtures** (X1-R3-07 to X1-R3-09): planning-flag rename; unequal-π noisy null; delivery mapping table with matching reasons.
5. **P text** (X1-R3-10, X1-R3-11, X1-R3-13): Maurer–Pontil bound; product-backed acceptance criteria; spacing pass and normative l.141.
6. **ADR-014** (X1-R3-06, X1-R3-14): adopted-form status and Review section; folder records; index row and Appendix B card; README links to the ADR and drops its table.
7. **Ledgers (Main)** (X1-R3-12):
   - rebase onto landed ADR-011 and remove X1's own R-064 row;
   - add the frozen X11 digest copy and point R-067 at it;
   - add the Maurer–Pontil R-ID;
   - fix the "byte-identical" wording.
8. **Preparation artefacts:** concrete source, tool, scanner and dev pins; the exposure-only manifest and extraction recipe; the GUIDE-author isolation record.
9. **Digests and submission.** Run `self_test.py`, the diagnostic and the benchmark once from the pinned toolchain with create-once outputs. Record the digests of every reviewed file. Submit X1 round 4, also saved as ADR-014 round 2.

## 12. Stage authorisation

| Stage | Status now | Earliest path |
|---|---|---|
| Drafting and synthetic validation, including the X1-R3-01 diagnostic and the X1-R3-05 benchmark | **Permitted** (unchanged) | — |
| Preparation (model-free pins, exposure manifest, GUIDE-author record) | **Not authorised**: 14 open issues (rule 1); its concrete artefacts do not exist (P §11.1); viability (X1-R3-01 to X1-R3-05) must be settled before the clock starts | A zero-issue round on revision 4, reviewed together with the preparation artefacts |
| Inventory and SRS draw (deterministic scans, eligible frame, keyed SRS, locked sizes) | **Not authorised** | After signed preparation outputs and a scoped review |
| Full OC and implementation round (sealed manifest run, GUIDE, product-backed pipeline) | **Not authorised** | After inventory locks sizes. All binding coverage, power and compute gates, and the X1-R3-11 parity, must pass before any dev model call. |
| Dev Jev and labels; test scoring and labels | **Not authorised** | Under full authority only; test calls only after the dev gates pass (cost/M̂ row, throughput, safety futility) |
| Gold unblinding and analysis | **Not authorised** | — |
| ADR-014 adoption | **Not authorised** | A zero-issue final round saved in its own folder, with the index row and card in the signed set |
| Default, threshold or product-sampler change | **Not authorised** | Separate signed ADR |

**Sign-off statement:** none. Round 3 closes with **14 issues open**:
- **0 blocking**;
- **6 major:** X1-R3-01 to X1-R3-06;
- **8 minor:** X1-R3-07 to X1-R3-14.

The reviewed files are those in `reviews/rev3-digests.json` (manifest SHA-256 `894dc63b0522a23589187c6de9fa4e672878fc1c07b85bfbbdcd224e5ab2f448`, as reported by Main), with the per-file digests in the table above. No preparation, inventory, Jev, labelling or analysis stage is authorised, and ADR-014 is not adopted.
