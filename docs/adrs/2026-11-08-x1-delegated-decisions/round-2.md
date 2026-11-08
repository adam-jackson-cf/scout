# X1: council experiment review, round 4 (revision 4 with ADR-014 round 2)

Date: 2026-11-08. Reviewer: innovation council. The lead wrote the synthesis. Two read-only peers ran in parallel:
- `innovation-challenger` (X1Council4.ChallengerR4);
- `experiment-peer` (X1Council4.PeerR4).

This record is X1 round 4. It is also ADR-014's second council review, and it must be saved byte for byte as `docs/adrs/2026-11-08-x1-delegated-decisions/round-2.md`.

**Governance.**
- `CLAUDE.md` experiment rules 1–7.
- Design rules 1–3, for ADR-014.
- The ledger rules, and the General sections "Jev labelling", "Model use" and "Secrets".
- Standing owner delegation R-064: where an owner choice arises, this record states the council's recommended option as the fix.

**Scope.** Revision 4, as listed in `reviews/rev4-digests.json` (40 files):
- `PREREGISTRATION.md`, `oc-manifest.json`, `README.md`, `reviews/README.md` and `reviews/round-3-response.md`;
- the validation scripts and outputs: `ranking.py`, `self_test.py`, `interval_diagnostic.{py,go}`, `compute_benchmark.py`, `outer_simulation.py`, the diagnostic spec, result and archived sources, the compute baseline, benchmark and three projections, the pilot summary and the synthetic results;
- `inputs/x11-rev4-digests.json`;
- ADR-014's `ADR.md` and `round-1.md`;
- the integrated shared files: `docs/adrs/README.md`, `docs/scout-design.html`, `docs/research-ledger.md`, `docs/hypothesis-ledger.md` and `docs/STATUS.md`.

Each round-3 item was checked against the fix that `reviews/round-3.md` required.

**What was and was not run.**
- No shell, model call, self-test, simulation, benchmark or digest recomputation was run.
- Evidence comes from reading files, jq queries over the JSON outputs, and arithmetic.
- `codegraph` was not used. Its reference documents were read first, but this read-only session had no way to run the executable.
- Paths are relative to `evaluation/x1-triage-ranking/` unless stated otherwise. "P l.N" means a line of revision 4's `PREREGISTRATION.md`.

**Reviewed files (SHA-256).** Main checked all 40 entries of `reviews/rev4-digests.json` with shasum. Main reports the manifest's own SHA-256 as `01c790837f961cd208fed98b3bd653b514da36e6d8bcee3b2fa4270f88e1876c`. This round did not recompute any digest.

| File | SHA-256 |
|---|---|
| `docs/STATUS.md` | `cc7fdc81918bfd80aa6108bdd5e71422760da386461093ac1bd60cfc2fec9275` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/ADR.md` | `54c368e69a856335879a2553459e995fc39b0657a910916fe5906ac829f214b4` |
| `docs/adrs/2026-11-08-x1-delegated-decisions/round-1.md` | `7bc229fd392215971646bfc547883050e82512c21777fbd311cf779e3e394710` |
| `docs/adrs/README.md` | `1b181875b596c5f07af16a68f3251aafb01c41d0f34354cfa9b9b0285e26ab09` |
| `docs/hypothesis-ledger.md` | `83b0d4743d960f571b5a79a2d1c43dc35e31f235eec5f48de0cdb33d943b3ae7` |
| `docs/research-ledger.md` | `c93be813ee6ef32ea37500622c6681a59f60ff18438fe570216b81c1a714de4d` |
| `docs/scout-design.html` | `f7180b872e7380b8c2a644b8c7fe33b90c9ccf3e3cc0a4c8c3662227dbb3d8a9` |
| `PREREGISTRATION.md` | `fcd9a9562cd34110617756e910c81569bf7ddf9c1d56cd1a9f703a32baccfc83` |
| `README.md` | `75e0c267dd53bf3af8f4468f07a3ba4fa611344bda9bc40e944703d5475200e9` |
| `inputs/x11-rev4-digests.json` | `e9edba67c885623df6900a08aee1cde463e5fa78f828bbb04ca3d8efd7a0324c` |
| `oc-manifest.json` | `5d0d4708df6ec33328e53e949d133cdacd403dc029c38d595087d2d1e4fc0dc0` |
| `reviews/README.md` | `8a868486badd356d759cb733f147c15f6480a5504b655b7d723ec43b7be256ea` |
| `reviews/rev3-digests.json` | `894dc63b0522a23589187c6de9fa4e672878fc1c07b85bfbbdcd224e5ab2f448` |
| `reviews/round-1-response.md` | `66340fd83af645782d5a8772fa7c74e4261d34a7e9d9ff01b8383fc5d5443ffe` |
| `reviews/round-1.md` | `89eab4a5d0f63487da200ab34db28dc0528ed9fcc73745c9d1075272339e172d` |
| `reviews/round-2-response.md` | `4c926bfa0111e458cb23b006101fa605ef57005cb2065a1f6e7260144bc14986` |
| `reviews/round-2.md` | `4021d173533876e775bf9200bc3a1aa1164f04feffbe9f3beb56dfb3b400d1ef` |
| `reviews/round-3-response.md` | `0117db6c3af240beddbef0ebc2a77a9ea886bb00f7aec4a4a984e89184e5b6ad` |
| `reviews/round-3.md` | `7bc229fd392215971646bfc547883050e82512c21777fbd311cf779e3e394710` |
| `validation/compute-baseline-rev4.json` | `44a3b3e8ca82cbb98c3b9a9b8ef04e82053b0184eacf4d4d35f237ba6ffb3497` |
| `validation/compute-benchmark-rev4.json` | `e5f068d7badf2f43afcfbd2ee0c0d0a270efe605329a43a865ae114a7aab1ca8` |
| `validation/compute-projection-rev4-final.json` | `b422c2a6fc7de6e548fa94cc4f36beec403ea9b169aff0429fb01386c44947dd` |
| `validation/compute-projection-rev4-total.json` | `65752e13309281fc8d1019034428103e5f2a445248bc1c142a958b402713a3c1` |
| `validation/compute-projection-rev4.json` | `4d2ca24912b6c94689138ff626869bd7473747a2dfde86497a59925f6e10dc0b` |
| `validation/compute_benchmark.py` | `dabebd882ea064da987c7460271edde243fb1dccd08468c932bc376a023456c4` |
| `validation/interval-diagnostic-rev4.json` | `50d41cc86c0342e05110d322b5d452ffcad7570c612efbbaee705ce11cc18a54` |
| `validation/interval-diagnostic-sources-rev4.json` | `d12d80627cecd6e7ee0b291dee939638cf7752c24ae62164bb495bf0ab693722` |
| `validation/interval-diagnostic-spec-rev4.json` | `e5a93a49c14a4345cba2fd6f2d4ddc55a998eaa199d371299d5965443fe54bb6` |
| `validation/interval_diagnostic.go` | `184fa317c972bd06bae02a46ba4a214a2bae5a1ebe8176c4e0cacbe8377fbbf3` |
| `validation/interval_diagnostic.py` | `a4b233374bdfec466bfc5862166497a4bbac48b13668aee2697fcb90e2b7c345` |
| `validation/outer-results.json` | `b851c2c1b6ffacee1957ade0071aa9fd38ede0dce5ebd2e3577a2967923f6352` |
| `validation/outer_simulation.py` | `91d07760b61551810282622c2bda53c34fcfcb72a9ffd1798b1a5ffb8ece23f6` |
| `validation/pilot-summary-rev4.json` | `05da900b22b738a61c36d4f02efc76c722109aa2dcffff40fba04f828c14475d` |
| `validation/ranking.py` | `405f2a4a466cd61fd46c8fc9e6aa25d98f76c3612d9c010eec614a79afdfcbad` |
| `validation/self_test.py` | `3c47f232dcb5fe1f5d52b8e4c156105c8e606d22260dd9bb778d86a57ae685ff` |
| `validation/synthetic-results-rev1.json` | `fbc4cc0de2c948574a76216cffd12cb33b488f0fedd3c3519de4ddf0b5a82277` |
| `validation/synthetic-results-rev2.json` | `24549f3ed2fcaecb357bf747ec31058ae115592f8bf63a1319763d2a84faa77d` |
| `validation/synthetic-results-rev4-percentile.json` | `b206c43600f80dab426f303a915695176418680296105144545e332dca04c56d` |
| `validation/synthetic-results-rev4.json` | `9f3ec381a945028c5a246ea723f03ee771815ff3bfdf609c917505e12c66c2f2` |
| `validation/synthetic-results.json` | `e65ec36c7d57a0d15f32bc3d2d7bd5f1f76f4148189ec2676653ef996901a350` |

**Byte-preservation checks.** These digests carry over unchanged from round 3:
- `round-1.md`, `round-2.md`, `synthetic-results-rev1.json`, `synthetic-results-rev2.json`, `synthetic-results.json` and `outer-results.json`.
- `round-3.md` and ADR-014's `round-1.md` are identical (`7bc229fd…`).

These have changed since round 3:
- `round-1-response.md` and `round-2-response.md`. X1-R3-12 instructed the R-064 wording change; both files are labelled as historical responses.
- `ranking.py`, `self_test.py` and `outer_simulation.py`. The revision-3 bytes that produced `outer-results.json` and `synthetic-results.json` are not archived (X1-R4-09).

**Verdicts.**
- **X1: 12 issues open (0 blocking, 2 major, 10 minor). No sign-off.**
- **ADR-014: 1 issue open (X1-R4-13, minor). No sign-off.** ADR-014 can also be adopted only on the joint zero-issue X1 round, because its decision table must carry the revision-5 choices that X1-R4-01 to X1-R4-12 change.
- **Joint signed set: 13 issues open.**
- No preparation, inventory, Jev, labelling or analysis stage is authorised (§12).

**Severity scale.** Unchanged from rounds 1–3:
- **blocking:** can produce a wrong or unreachable verdict, or breaks experiment rules 1–7 or REQ-169.
- **major:** misleads the result, leaves a preregistered quantity undefined, or makes the study unrepeatable or likely not testable.
- **minor:** text, consistency or provenance. Minor issues must still be fixed before sign-off (rule 1).

---

## 1. Recommendation

**Revise to revision 5 and submit it, together with the concrete preparation artefacts, for X1 round 5 / ADR-014 round 3. Do not sign off.**

Revision 4 is a substantial and faithful response: 11 of the 14 round-3 issues are fixed and 3 are partly fixed. Keep all of the following:
- the predeclared diagnostic and its mechanically applied studentised branch, with no fallback;
- the sealed OC manifest: cells, five populations, eligible-trial denominators, realised power and null, row timing and the fail-closed final evaluator;
- the 252-cell missing-label grid at 10,000 eligible trials;
- the rank-structure factor and the refined M grid;
- the compiled kernel, with Go/Python parity and deterministic workers;
- the delivery mapping, the Maurer–Pontil bound and the named product-backed acceptance frames;
- the adopted-form ADR-014, its folder records, index row and Appendix B card;
- the single landed R-064.

Two defects must be settled before preparation, because the 45-day clock starts at the first preparation action (P l.208):

1. **The compute fit rests on a preparation cost that was measured on 5- or 20-run templates, not on L-run populations** (X1-R4-01). Scaled to a genuine L = 400 population, the projection plausibly exceeds the 2,880 core-hour cap [INFERENCE].
2. **The product-backed recovery acceptance can fail by chance for a correct pipeline** (X1-R4-02). Failure blocks full authority.

The ten minor issues are cheap synthetic or text work:
- a calibration side rule;
- the size-stress level;
- the zero-variance pivot;
- M̂ below the grid and the attempt cap;
- safety-grid rates and scheduling;
- two fixture gaps;
- reporting and provenance;
- threats with a throughput screen;
- leftover text;
- ledger rows.

## 2. Alternatives considered

| Option | Description | Tradeoff | Council position |
|---|---|---|---|
| A. Sign off revision 4 now and carry the fixes into the implementation round | Freezes the protocol immediately | Rule 1 requires zero issues. X1-R4-01 undermines the pre-preparation compute gate that round 3 made binding. | Rejected |
| B. **Revision 5 closes all 13 issues; the preparation artefacts are reviewed with it; then a zero-issue round** | Viability and resource fit are settled before the clock | One more round | **Recommended** |
| C. Authorise a scoped model-free preparation now | Starts pinning work early | Starts the clock with the resource fit unproven (X1-R4-01). The preparation artefacts do not exist yet (response, "Remaining authority boundaries"). | Rejected |
| D. Drop the descriptive canonical-safety grid to free compute | Saves about 1,010 core-hours | Round 3 (X1-R3-02) requires a canonical-label safety OC. Rescheduling removes the clock cost without losing it (X1-R4-07). | Rejected |
| E. Broaden the interval diagnostic before accepting studentised | More configurations | The branch has already triggered, and the locked-inventory gate remains binding. A broader diagnostic could only re-select the same branch. | Rejected. The representative-cell scope is accepted as method-selection evidence (§4). |
| F. Replace percentile-t with a basic or BCa interval | Other second-order or skew-aware options | Round 3 fixed a binary branch before outcomes. Switching now would be an outcome-informed method choice. | Rejected |
| G. Move all model-free implementation (OC generator, final evaluator, product harness, isolated GUIDE) before preparation | Shortens the clocked path | More work before the clock starts, but none of it needs real data | **Recommended practice**, not a sign-off condition (§11) |

## 3. Rule checks

- **Rules 1–2.** This round is recorded with digests. Issues are open, so iteration is required.
- **Rule 3.** No predictor–outcome association was computed on real data:
  - the diagnostic is synthetic (`synthetic_only: true`), and its spec is saved create-once before any outcome (`interval_diagnostic.py:117–118`);
  - the branch rule is applied mechanically (`:119–120`), and a source or binary change aborts the run (`:123`);
  - `pilot-summary-rev4.json` derives only from the synthetic `outer-results.json`.
- **Rule 4.**
  - Kernel level: the planted effect gives support, the census null is kept, and the binding-cap unequal-π null is not supported with interval [−.0324, .0327] and real weighting mutants (`self_test.py:54–81`).
  - The product-backed acceptance criteria are defined, but their recovery clause is a chance gate (X1-R4-02).
  - The fixture coverage of the gradient generator and of the verdict boundaries is incomplete (X1-R4-08).
- **Rule 5.** Nothing is signed yet. Revision-1 and revision-2 evidence and the revision-3 outputs are preserved byte for byte. The revision-3 executed sources are not archived (X1-R4-09).
- **Rule 6.** The full pilot blocks are now reported. The diagnostic summary and the near-threshold OC wording are incomplete, and the projections lack reproduce commands (X1-R4-09).
- **Model use and secrets.** Unchanged and compliant (P l.93, l.206).
- **Design rules.**
  - ADR-014's format now follows rule 2 (X1-R3-06 is fixed).
  - Rule 3 does not bar it: the cited diagnostic and benchmark are implementation evidence inside this jointly signed set, not product-effect results.
  - The tracker and claim residuals are in X1-R4-13.
- **Ledgers.**
  - R-064 appears once, in its landed form (`docs/research-ledger.md:73`), and R-078 is present (`:81`).
  - Internal-study rows are missing for the revision-4 diagnostic and benchmark (X1-R4-12).
  - `docs/STATUS.md` is stale on R-078 and on the ADR list (X1-R4-13).

## 4. Verification of round-3 fixes

| Issue | Status | Evidence |
|---|---|---|
| X1-R3-01 (major) | **Fixed** | **(a)** README reports all 12 M × SD × failure blocks, with correlated-seed caveats, P(Bin(100, .025) ≥ 8) = .0037, 4.26 expected exceedances, and capped realised and support ranges.<br>**(b)** Nine cells, L ∈ {40, 100, 400}, 2,000 trials, B = 1,999, independent seeds, create-once spec (`interval-diagnostic-spec-rev4.json`).<br>**(c)** The rule fired on M850-L400 (percentile lower-bound misses 118/2,000, Wilson lower .049) and on no-cap-L400 (147/2,000). Studentised is fully specified (P l.111–113), with hand fixtures (`self_test.py:122–127`) and a fresh-seed rerun.<br>**(d)** A size-dependence factor is in the grid (P l.159); its level is reopened in X1-R4-04.<br>**Recorded interpretation:** one representative cell per block, at the adverse configuration (SD 0, overlap 0), is accepted as method-selection evidence only. It is not block-wide coverage certification. |
| X1-R3-02 (major) | **Fixed** | `oc-manifest.json` defines cells, five populations, eligibility, coverage gate, calibrated realised power and null, row timing and canonical safety (P l.153–175). Residual definitional gaps are in X1-R4-03. |
| X1-R3-03 (major) | **Fixed** | 252 cells × 10,000 eligible trials, sealed mask and shortfall rules, no grid reduction (manifest `missing_label_grid`). |
| X1-R3-04 (major) | **Partly fixed** | Exact linear-density gradient (P l.157; `ranking.py:388–404`); grid {2,000, 1,500, 1,000, 850, 700, 550, 400}; capped reachability disclosed; R-052 projection M̂ ≈ 848.9 → row 700 (P l.200). **Residual:** the inventory cap-tail list still reads "M = 2,000/850/400" (P l.41), so it reports only three rows (X1-R4-11). |
| X1-R3-05 (major) | **Partly fixed** | Per-trial complete and full-512-vertex costs at B = 1,999 on 60,000-unit frames, L ∈ {40, 400}; one- and two-worker byte equality; caps of 8 workers, 2,880 core-hours and 15 days (P l.177–179). The arithmetic reproduces exactly (§6). **Residual:** the population-preparation input is not a full-population measurement (X1-R4-01). |
| X1-R3-06 (major) | **Fixed** | Adopted-form status (ADR l.3); no pending prose; Review and Records sections; `round-1.md` equals `round-3.md`; index row (`docs/adrs/README.md:28`); card (`docs/scout-design.html:1006–1018`, "no relation"). Tracker residuals are in X1-R4-13. |
| X1-R3-07 (minor) | **Fixed** | `outer_simulation.py:115–118` emits only numerical planning flags. The final-evaluator contract is at P l.175 and in the manifest `final_evaluator_contract`. |
| X1-R3-08 (minor) | **Fixed** | Heterogeneous binding-cap null with π < 1, both discordance directions, non-zero width, and ignored-π and nominal-π mutants (`self_test.py:58–81`). |
| X1-R3-09 (minor) | **Fixed** | Mapping table (P §4). `STUDY_REASONS` and `FAILURE_REASONS` (`ranking.py:322–326`) and fixtures, including the exact 10 % boundary and the non-dilution case (`self_test.py:187–195`). |
| X1-R3-10 (minor) | **Fixed** | Theorem 4 stated exactly, with ln 40, Bessel variance and b = 900 (P l.204); R-078. |
| X1-R3-11 (minor) | **Fixed as specified** | Named frames, ≤ 1e−12 parity, seeds 169/170/171 and a projection formula (P l.183). The recovery clause that round 3 itself specified is a chance gate (X1-R4-02). |
| X1-R3-12 (minor) | **Fixed** | X1's duplicate R-064 is removed. The landed row at `docs/research-ledger.md:73` matches `scout-adr11b/docs/research-ledger.md:73`. The fast-forward 0350ef0e → 7bb73e52 is in the worktree reflog (`scout/.git/worktrees/scout-x1/logs/HEAD:2`). R-067 resolves to the local copy; R-078 is added. |
| X1-R3-13 (minor) | **Partly fixed** | The editing instruction is replaced with "No discretionary coverage-review rule exists" (P l.165). **Residual:** "support1" (P l.135) and "and20–2000-issue" (P l.224) remain (X1-R4-11). |
| X1-R3-14 (minor) | **Fixed** | The README links to ADR-014 and has no options table. |

## 5. New issues

### X1-R4-01: major. The compute fit uses a population-preparation cost measured on 5- or 20-run templates, not on L-run populations

- **Location:**
  - `validation/compute_benchmark.py` `trial_frame` (l.46–56 docstring: "Five independent 1,000-unit run templates are replicated to L eligible runs") and `projection()` (`setup_cpu`, `prep_hours`);
  - `compute-projection-rev4-final.json` and `compute-projection-rev4-total.json` (`max_preparation_cpu_seconds` .482249);
  - manifest `resources.calibration_projection` and `populations.sizes`;
  - P l.161, l.167, l.169, l.177–179; ADR-014 row "What compute execution/caps?".
- **Problem:**
  - **What the contract requires.** It generates a finite population on every eligible run's locked size (P l.151; manifest `populations.sizes`), with up to 22 preparations per calibrated population and one per canonical-safety outcome population.
  - **What was measured.** The benchmark's "preparation" generates five 1,000-unit runs (uniform fixture) or twenty runs (mixed fixture; 60,000 units across three repositories), then replicates groups to L. Its maximum setup CPU of .482 s therefore measures about 60,000 units, roughly 8 µs per unit.
  - **Scaling** [INFERENCE: linear per-unit scaling]. A genuine L = 400 population of 1,000-unit runs has 1.2 M units, about 9.6 s per preparation (about 20×).
  - **Effect on the total** [INFERENCE: arithmetic].
    - The projection contains about 355,880 preparations: 62,972 base, 215,628 calibration, 36,960 safety calibration and 40,320 canonical.
    - They are currently costed at about 95 core-hours.
    - At 9.6 s with 2× overhead they cost about 1,906 core-hours, which brings the total to about 3,200 core-hours, above the 2,880 cap.
  - **The gate's conclusion is unsupported.** `fits_caps: true` is correct arithmetic on an input that does not measure the registered work. P l.179 makes a failed fit a stop before preparation.
- **Required fix** (recommended options under R-064):
  - (a) **Register a fixed-key calibration path in the manifest.** Calibration and canonical-safety evaluations keep the population keys fixed and recompute only the latent-gain-dependent Jev allocation, the truth and the per-case states. The population is not regenerated per bisection step.
    - Add an equivalence fixture: the fixed-key path must equal full regeneration exactly on small populations.
  - (b) **Benchmark genuine populations.** Measure independent L-run population construction, and the registered calibration path, at L ∈ {40, 400}, with mean run size 1,000 and maximum 2,000.
  - (c) **Recompute the authoritative total.**
    - Use a committed script with repository-relative paths, covering X1-R4-04's added cells.
    - Record the per-unit preparation cost, so locked sizes up to the 2,000-run bound can be extrapolated at inventory.
  - (d) If the total exceeds the cap, revise before preparation. Shrinking the grid is permitted only through a reviewed revision.
  - (e) Extend the locked-size re-measurement before full authority to cover preparation and calibration as well as per-trial cost.
  - Update ADR-014's compute row.

### X1-R4-02: major. The product-backed recovery acceptance fails by chance even for a correct pipeline

- **Location:** P l.183; `reviews/README.md` ("planted support and nonzero-width noisy unequal-π null recover on fixed seeds"); ADR-014 row "What product parity?".
- **Problem:**
  - **Effect clause.** At seeds 169/170/171, "planted realised gain ≥ .20 as support" is required on every seed. The frame (row, cap and prevalence) is unspecified. At the power standard of .80–.90, P(3 of 3 supports) is about .51–.73. In a capped row, realised .20 may be infeasible altogether.
  - **Null clause.** "Not supported with nonzero width containing 0" per seed fails with probability about 1 − .95³ ≈ .14 from the coverage clause alone. "Not supported" also requires best-case UB < .10 (P l.130).
  - **Consequence** [INFERENCE: binomial arithmetic]. A correct pipeline may be refused full authority on fixed seeds, which makes the study not testable by construction. This is round 3's own X1-R3-11 specification, refined here.
- **Required fix:**
  - Keep exact parity (stored π, draws, point, interval and verdict ≤ 1e−12) as the binding implementation check.
  - **Effect frames.** Use locked sizes with M = 2,000 (no binding cap), failure 0, no unannotated units, prevalence .20 and realised gain ≥ .30, verified before use by an OC check of ≥ 2,000 trials with support Wilson lower bound ≥ .99. Require support on all three seeds.
  - **Null frames.** Per seed, require verdict ≠ support and non-zero width. Judge coverage of 0 through the OC coverage gate, not per seed.
  - Update P l.183, `reviews/README.md` and ADR-014.

### X1-R4-03: minor. The manifest leaves the realised-.10 calibration side and the canonical-safety identities undefined

- **Location:** manifest `calibrated_power_and_null.calibration` (sides defined only for ≤ 0 and ≥ .20), `missing_label_grid.realised_gain`, `final_safety_OC.dimensions.realised_gain` and `.reachability`, `populations.cell_identity_kind` and `.phases`; P l.165, l.167, l.169.
- **Problem:**
  - **Realised .10.** The missing and safety grids both target realised .10, but no attainable-side rule exists for it.
  - **Safety identities.** `cell_identity_kind` omits safety profiles and interval groups. The exact canonical identity that lets masks and bootstrap draws be shared across critical shares, while criticality draws stay separate, is undefined.
  - **Consequence.** Two compliant implementations can generate different populations and seeds. The coverage effect is negligible, because truth is computed from the population itself, but the manifest stops being an exact contract.
- **Required fix:**
  - **.10 rule.** The calibrated target is the smallest attainable realised gain ≥ .10 (mirroring the power side). Report target, realised value and excess; an unreachable target is unmeasured.
  - **Identity kinds.** Add `safety-profile` and `safety-interval-group`. Interval-group identity excludes the critical share; criticality uses a separate phase `canonical-criticality` keyed by the critical share.
  - Add deterministic fixtures for both mappings.

### X1-R4-04: minor. The binding coverage cells omit the size-dependence stress that drove the method change

- **Location:** P l.159; manifest `coverage_dimensions.size_dependence` and `size_dependence_choice`; ADR-014 row "What size dependence?".
- **Problem:**
  - The stated reason for a_size = .05 (an absolute .20 decrement makes realised .20 power unreachable at prevalence .55) applies only to calibrated power references.
  - Coverage cells measure against whatever truth arises, so that reason does not bind them. Yet the binding coverage gate uses a quarter of the .20 stress under which the diagnostic separated the methods (percentile 147/2,000 against studentised 14/31 at no-cap L = 400).
  - "Fixed before diagnostic outcomes" is the author's statement; no artefact verifies it.
- **Required fix:**
  - Set requested-latent coverage cells to size dependence {0, .05, .20} (16,128 cells; the generator already accepts .20, `ranking.py:358`). Keep {0, .05} for calibrated references and the safety grid.
  - Include the extra cells in X1-R4-01's re-projection.
  - Reword the timing claim as author-recorded in P, the manifest and ADR-014.

### X1-R4-05: minor. Zero-variance replicates whose point differs are dropped, which pulls the pivot quantiles inward

- **Location:** P l.111; `ranking.py:186–191`; the Go helper; `self_test.py:126`.
- **Problem:**
  - A replicate with V* = 0 and D* ≠ D̂ has pivot ±∞. Omitting it (up to the 5 % inadequacy limit) shrinks the corresponding tail quantile, which is anti-conservative on that side.
  - The sign is well defined, unlike a zero-denominator replicate.
- **Required fix:**
  - Retain such replicates as signed infinities in the pivot ordering, and do not count them as omitted. An infinite selected quantile yields a [−1, 1]-clipped endpoint.
  - Undefined variance and zero denominators remain omitted.
  - Update P l.111, the Python and Go kernels, the hand fixture and the Go differential fixture.

### X1-R4-06: minor. The row-selection rule is silent when M̂ falls below the grid, and the fixed attempt cap ignores the sampling-gate pass probability

- **Location:** P l.161, l.173, l.200; manifest `row_timing`, `eligibility`, `populations.maximum_attempts_per_population`.
- **Problem:**
  - **M̂ below 400.** No registered row exists, so the outcome label is undefined.
  - **Attempt cap.** The cap is 4,000 attempts per 2,000 eligible. If keyed SRS redraws on the locked frame pass N_r ≥ 1,000 and ΣN ≤ 60,000 less than half the time, every cell exhausts its attempts and fails, whatever the coverage. That probability is computable from locked sizes alone.
- **Required fix:**
  - **M̂ rule.** "No registered row ≤ M̂ → not testable; no test calls."
  - **Gate-pass probability.** At inventory, estimate the sampling-gate pass probability p̂_gate from 100,000 model-free keyed SRS redraws of the locked eligible frame. Set the per-population attempt cap to ceil(2 × 2,000 / p̂_gate). Adequacy failures still count against the cap.
  - Include the cap in the locked-size compute re-measurement. A failed fit means not testable before dev.

### X1-R4-07: minor. The safety boundary grid does not bracket the transition at locked sizes, and it sits on the clocked critical path although it gates nothing

- **Location:** manifest `final_safety_OC.unresolved_rate` {0, .0001, .0005, .001, .005, .02} and `critical_positive_share` {0, .001, .01, .05}; `row_timing`; P l.169, l.173, l.216; `compute-projection-rev4-total.json` (`additional_core_hours.total` 1,010.36).
- **Problem:**
  - **Rates.** P(no unresolved) = (1 − r)^M moves from near 1 to near 0 over r ≈ 0.1/M to 3/M. With M = Σ floor(N/2) up to 30,000, the smallest non-zero rate already gives about e^−3, and 5e−4 gives about e^−15. The "boundary" grid can therefore miss the boundary entirely at large locked sizes, while it does bracket it at small ones.
  - **Schedule.** The grid is descriptive, but as written it runs in the full-OC stage before dev. It accounts for 1,010 of the 1,409 projected core-hours (about 5.3 of 7.3 days at 8 cores) inside the 45-day clock, which reduces the labelling envelope (P l.206).
- **Required fix:**
  - **Locked-size rates.** At inventory, set unresolved rates r = κ / M̄ for κ ∈ {0, .1, .5, 1, 3, 10}, where M̄ is the expected sampled below-floor count from the locked eligible sizes.
  - **Critical shares.** Set c = min(1, j / TP̄_below) for j ∈ {0, .1, 1, 3}, with TP̄_below the expected true positives below the floor in each synthetic population, computed before outer trials. The grid stays 6 × 4, so compute is unchanged.
  - **Scheduling.** The grid runs after every binding row is complete. It gates neither dev nor test, and may run concurrently with dev and test labelling within the same 8-worker, 2,880 core-hour and 15-day caps. It must finish before gold freeze; otherwise it is reported unmeasured, with no effect on the verdict.
  - Update P, the manifest and ADR-014.

### X1-R4-08: minor. Fixtures do not test the gradient's defining property or the verdict boundaries

- **Location:** `self_test.py:202–204` ("refined cap grid and gradient" checks only cap membership); `self_test.py:182–186`; `ranking.py:202–213`.
- **Problem:**
  - **Gradient.** The linear-gradient generator governs half of all coverage cells, but no fixture checks the exact positive count below k or the declining density.
  - **Verdict.** The function is correct on reading, but these boundaries are untested:
    - unresolved safety with otherwise supporting retrieval → inconclusive;
    - point = .10 exactly → support;
    - LB = 0 exactly → not support;
    - UB = .10 exactly → not "not supported";
    - violating case with inadequate retrieval → not supported.
- **Required fix:**
  - **Gradient fixtures.** For several H, P and seeds, assert exactly P positive slots. Assert that for every prefix j, |#positives in the first j slots − Σ_{i<j} p_i| < 1, which follows from systematic allocation. Assert that the first half holds more positives than the second when 0 < p < 1.
  - **Verdict fixtures.** Add the five boundary fixtures above.

### X1-R4-09: minor. Reporting and provenance gaps

- **Location:** README "Predeclared interval diagnostic" and the response ("Largest fresh L = 400 upper Wilson endpoint is .0334"); P l.135; `pilot-summary-rev4.json`; `compute-projection-rev4-final.json` and `-total.json` (no script; absolute worktree paths); README "Reproduce" (absolute `/Users/.../scout-x1/...` paths); `outer-results.json` embedded source digests.
- **Problem:**
  - **(a) Diagnostic summary.** The summary reports only L = 400. Fresh studentised M2000-failure02-L100 has lower-tail misses of 55/2,000, Wilson [.0212, .0356]: above the .035 bar at pilot precision (`interval-diagnostic-rev4.json` `fresh_studentized[4]`). The studentised asymmetry also reverses: upper-bound misses exceed lower-bound misses in eight of nine cells.
  - **(b) Near-threshold OC.** P l.135 promises "near-threshold OC at .05/.08/.10/.12", but requested latent gains are descriptive only (manifest `requested_latent_rows`), and in capped rows they land far below realised .10.
  - **(c) Projections.** The pilot summary and the authoritative projection have no committed generator, and there are no repository-relative reproduce commands.
  - **(d) Revision-3 sources.** The executed revision-3 sources (ranking `6b2910…`, self-test `0f5f98…`, outer `9760d0…`) are overwritten and not archived. This is not disclosed.
- **Required fix:**
  - (a) Tabulate all nine initial and all nine fresh cells (both tails, Wilson intervals, bias, SD and bootstrap SD) in the README, and record the L = 100 exceedance and the asymmetry reversal.
  - (b) Restate P l.135: near-threshold OC is reported descriptively, binned by realised population gain at .05/.08/.10/.12 from the existing cells, with mean point bias.
  - (c) Add a committed script that reproduces both the pilot summary and the projections, and reproduce commands relative to the repository root.
  - (d) Archive the revision-3 bytes if they can be recovered; otherwise disclose that they cannot.

### X1-R4-10: minor. The development-case source, four threats and a pre-OC throughput screen are missing

- **Location:** P l.103, l.200–206 ("dev cases" is undefined), l.214–216, l.222–228; manifest `resources.remaining_reservation`.
- **Problem:**
  - **Undefined source.** P never says which units are the ≥ 200 "development cases" behind r̂, durations, retry and adjudication rates, and c_U. The only defined development collection is the D1-R fastapi/axum census (P l.188).
  - **Unlisted threats.** None of the following is a listed threat:
    - transferring r̂, durations and cost from the dev repositories to the test repositories;
    - OC validity only inside the generator's ranges;
    - the positive point bias in size-dependent cells (+.0018 to +.0027, `bias_mean`);
    - subscription quota for up to about 150,000 sessions.
  - **Late detection.** Labelling infeasibility is detected only after preparation, inventory, full OC and dev, although N is known at inventory. With n_max = 300 dev cases, the Maurer–Pontil term alone gives U95 ≥ 7·900·ln 40 / (3·299) ≈ 25.9 s, so the projection is at least 2 × 2.4 × N × 25.9 / 4 ≈ 31.1·N seconds [INFERENCE: P l.204–206 with the .40 adjudication floor and zero retries].
- **Required fix:**
  - **Dev source.** State that the dev cases are the D1-R fastapi/axum development census (P §9), with n_max fixed in the preparation artefacts.
  - **Threats.** Add the four threats.
  - **Inventory screen.** After inventory, if 31.1·N_reserved seconds (recomputed from n_max) exceeds 45 days − elapsed − 7-day Jev reservation − projected OC wall days at locked sizes, stop as not testable before the full OC.

### X1-R4-11: minor. Text residuals

- **Location:** P l.41 ("M = 2,000/850/400"); P l.135 ("support1"); P l.224 ("and20–2000-issue"); `docs/hypothesis-ledger.md:42` ("Council round1 found27 issues"; uncommitted, so it can be corrected in place).
- **Required fix:**
  - P l.41: "each registered cap-order M in the manifest grid".
  - Respace the three tokens: "support 1", "and 20–2,000-issue", "round 1 found 27".

### X1-R4-12: minor. The research ledger lacks the revision-4 internal-study sources

- **Location:** `docs/research-ledger.md`; P l.11; ADR-014 "Informed by" and the rows on primary interval and compute caps; HD9 sources.
- **Problem:**
  - The revision-4 diagnostic justifies ADR-014's primary-interval row, and the benchmark and projection justify its resource caps.
  - Neither has a ledger row (ledger rule 1). Precedent: R-066, and X11a's R-075/R-079 in `scout-x11`.
- **Required fix:**
  - Add internal-study rows for:
    - the revision-4 interval diagnostic (spec and result digests, with X1-R3-01 as the recommending source);
    - the revision-5 compute benchmark and projection.
  - Cite them in P §6 and §8, in ADR-014 and in HD9.
  - Allocate IDs that collide with nothing: R-077 is used in `scout-adr11` and R-079 in `scout-x11`.

### X1-R4-13 (ADR-014 and joint trackers): minor. Unverifiable claims in signed ADR text, a stale STATUS, and an index note

- **Location:**
  - ADR-014 rows "What size dependence?" (".05 fixed before diagnostic outcomes, not selected by them") and "Which final canonical-safety grid and cost?" ("acknowledged by Main before sealing");
  - `docs/STATUS.md:7, 126` ("decisions ADR-000…ADR-011") and `:128` ("next free R-078");
  - `docs/adrs/README.md`: the route-(a) paragraph names only ADR-011, and the index jumps from ADR-011 to ADR-014.
- **Problem:**
  - **Claims.** Once signed, ADR text becomes permanent. These two claims cannot be checked from any artefact, and they record process narrative.
  - **STATUS.** STATUS is in the signed set but says R-078 is free, while X1 uses it, and it omits ADR-014.
  - **Index.** The index does not say that ADR-014 takes effect only on its final zero-issue round, or that ADR-012 and ADR-013 are allocated to concurrent work.
- **Required fix:**
  - **ADR text.** Reword both claims as author-recorded, or remove them. Carry every revision-5 decision (X1-R4-01 to X1-R4-12) into the single table.
  - **STATUS.**
    - Add ADR-014 to l.7 and l.126 ("effective on its final zero-issue council round").
    - At l.128, record R-078 as X1's (plus the X1-R4-12 IDs) and the next free ID under the current allocation.
  - **Index.** Extend the route-(a) sentence to ADR-014, and add: "IDs were allocated across concurrent work, so a gap does not imply a missing record."

## 6. Assessment of the targeted questions

- **Studentised diagnostic choice: sound, and correctly selected.**
  - The rule was fixed before outcomes and executed mechanically (`interval_diagnostic.py:117–123`).
  - The triggers were extreme: M850-L400 Wilson lower .049, and no-cap-L400 147/2,000.
  - The implementation matches P l.111–113:
    - replicate variance uses the resample's own m = 19 (`ranking.py:167–177`);
    - pivots are reversed and unclipped (`:195–199`);
    - the hand variance of .16 and the [−.4, .4] pivot fixture pass;
    - Go/Python parity covers both methods and the full 512-combination envelope.
  - The final Python kernel is tied to the diagnostic's unchanged Go helper (`184fa317…`) through those parity fixtures.
  - Fresh studentised L = 400 tails (lower-bound / upper-bound misses): 18/33, 45/51, 14/31, all within .035 at pilot precision. One L = 100 cell exceeds it (X1-R4-09).
  - The power cost of the switch is unmeasured until the full OC.
  - Remaining defects are the zero-variance rule (X1-R4-05) and the size-stress level (X1-R4-04).
- **Boundary safety grid.**
  - The arithmetic is coherent: 336 profiles, 8,064 outcomes, and 336 complete plus 1,680 full-vertex groups.
  - Interval reuse across critical shares is exact, given separate keys for criticality.
  - The rates are absolute and miss the transition at large locked M, and the grid sits on the critical path (X1-R4-07). Its identities are undefined (X1-R4-03).
  - The grid is descriptive, so no coverage or missing cell is lost; that is correct.
- **1,409 core-hours / 7.3 days.**
  - The arithmetic reproduces exactly:
    - 176.75 + 147.74 + 16.87 = 341.36;
    - + 57.77 = 399.13;
    - + 4.73 + 984.93 + 9.90 + 10.80 = 1,409.49;
    - ÷ 192 = 7.341 days.
  - The worst observed per-trial costs (.0012681 s complete, .0527639 s full-vertex) are used, with 2× attempts and 2× overhead.
  - The preparation input is invalid (X1-R4-01). Eight-worker scaling is projected, not observed. L above 400 is deferred to locked-size re-measurement, which is acceptable once X1-R4-01(e) extends it to preparation.
  - Among the three projection files, `-total` is authoritative and is labelled so in P l.179 and the README. `-rev4` and `-final` are superseded components; X1-R4-09(c) adds the missing script.
- **ADR-014, index row and card.**
  - The format satisfies design rule 2: adopted-form status (matching ADR-011's ADR.md l.3), no stale outcome text, Review naming the explicit zero-issue final round in its own folder, and Records listing `round-1.md` and later rounds.
  - The card reads "accepted / no relation", with the adoption condition in its third bullet.
  - The decision table is complete and agrees with P on every row checked.
  - The ID is allocated soundly (round 3 §6), with no collision found.
  - The residuals are in X1-R4-13.
- **Rebase integrity.**
  - The reflog shows a fast-forward 0350ef0e → 7bb73e52.
  - A single R-064 (l.73) is identical to `scout-adr11b/docs/research-ledger.md:73`; R-069 is retained (l.74).
  - The ADR-011 index row, card (`scout-design.html:991–1004`), REQ-22/78 text (l.589, l.600) and STATUS content are present.
  - X1's R-065 to R-068, R-071, R-072, R-078 and HD9 rows are present.
  - R-078 does not collide with R-077 (`scout-adr11`) or R-079 (`scout-x11`).
  - This was a section comparison, not a bytewise whole-tree diff against 7bb73e52.

## 7. Checked and sound

- **Diagnostic provenance.** The spec, binary and sources are embedded. The archived pre-extension `ranking.py` (`7202f906…`) is preserved in `interval-diagnostic-sources-rev4.json`.
- **Manifest counts.** 7 × 4 × 2 × 6 × 2 × 2 × 2 × 2 × 2 = 10,752 coverage cells. The 896 profiles give 1,792 references. 7 × 2 × 3 × 2 × 3 = 252 missing cells. The safety grid has 336 profiles. Calibration needs 224,840 preparations, minus the 9,212 already counted.
- **Gradient algebra.** The slot densities sum exactly to P, and systematic allocation yields exactly P positives (`ranking.py:388–404`). Only the fixture is missing (X1-R4-08).
- **Size dependence.** The fixture shows a gain of .20 at size 250 and .15 at size 1,000 under a_size = .05 (`self_test.py:205–207`).
- **Delivery integrity.** The run-level key-absence and sink-failure reasons precede per-unit classification. Unknown-cost units do not dilute the failure rate.
- **Worker determinism.** The scientific JSON digest `4471645409…` is identical at one and two workers.
- **Ledgers.** R-067 resolves locally. The R-078 text matches Theorem 4. HD9 is still proposed, with an accurate history line.

## 8. Evidence index

- **Diagnostic, initial percentile (lower-bound / upper-bound misses per 2,000):**
  - M850 L40/100/400: 29/0, 85/11, 118/11;
  - M2000-failure .02: 52/61, 59/52, 61/63;
  - no-cap: 24/1, 101/4, 147/8.
- **Diagnostic, fresh studentised:**
  - M850: 0/8, 10/38, 18/33;
  - M2000-failure .02: 45/47, 55/48 (Wilson upper .03562), 45/51;
  - no-cap: 2/5, 8/26, 14/31.
- **Benchmark** (`compute-benchmark-rev4.json`):
  - setup CPU about .235 s (uniform fixture, 15,000 template units) to .482 s (mixed fixture, 60,000 units);
  - 60,000 selected units every trial;
  - `frame_design.population_representation`: "Uniform: L grouped 1000-unit runs" built from five templates.
- **Preparation re-scaling** [INFERENCE]: 8 µs per unit × 1.2 M units ≈ 9.6 s per preparation; 355,880 preparations × 19.3 s ≈ 1,906 core-hours, against about 95 costed.
- **Recovery-gate arithmetic** [INFERENCE]: .8³ = .512; .9³ = .729; 1 − .95³ = .143.
- **Throughput bound** [INFERENCE]: 7 × 900 × ln 40 / (3 × 299) = 25.91 s; 2 × 2.4 × 25.91 / 4 = 31.09 s per reserved case.
- **Ledger allocations:** R-077 at `scout-adr11/docs/research-ledger.md:76`; R-079 at `scout-x11/docs/research-ledger.md:76`; STATUS l.128 says "next free R-078".

## 9. Risks

- **Gate multiplicity.** About 3,600 one-sided tail tests per row at Wilson upper ≤ .035 with 10,000 trials. P(row passes) is about .9 if every tail sits at .025, and about .1 if around 100 tails sit at .028 [INFERENCE: peer arithmetic].
- **Cost-cap reachability.** The R-052 projection selects row 700. In frames dominated by runs above 700 units, realised .20 is likely infeasible, so the study would stop after dev as not testable [INFERENCE]. The action is the same as for inconclusive (switch to the deterministic default). This is a legitimate product finding, accepted in round 3.
- **Clock.** After X1-R4-01, X1-R4-07 and X1-R4-10 the clocked path is more predictable. It still includes the implementation review rounds unless option G is followed.
- **Shared-file staleness.**
  - The joint signed set holds whole-file digests of shared documents (`STATUS`, both ledgers, ADR index and design document).
  - Any sibling landing (ADR-012, ADR-013, X11a) before X1 merges changes those bytes and forces a new round.
  - Recommendation for the final sign-off: also name the X1-owned rows and card as extracts.
- **Reversibility.** Every fix is pre-data and synthetic. No product, default, accepted ADR or requirement changes.

## 10. Where the peers changed the council's recommendation

- **Challenger:**
  - Found that the preparation-cost input measures templates, not populations. **Adopted as major X1-R4-01**; the lead quantified it (about 20× at L = 400, exceeding the cap). The lead had verified only the projection arithmetic.
  - Raised the realised-.10 side and the safety identity. **Adopted**, downgraded from major to **minor** X1-R4-03, because the numerical effect is negligible and the issue is determinism.
  - Raised the diagnostic-scope amendment. **Adopted as a recorded interpretation** (§4), not an issue.
  - Found the stale cap list. **Adopted** (X1-R4-11).
- **Experiment peer:**
  - Found the recovery chance gate. **Adopted as major X1-R4-02.**
  - Raised the throughput screen and implementation before the clock. **Adopted** as minor X1-R4-10 (screen, threats and dev source) plus option G as recommended practice. Downgraded from major because the protocol already stops before test; the screen saves effort but does not change validity.
  - Raised:
    - the zero-variance pivot (X1-R4-05);
    - the coverage size stress (X1-R4-04);
    - M̂ below 400 and the attempt cap (X1-R4-06);
    - verdict-boundary fixtures (part of X1-R4-08);
    - locked-size safety rates (part of X1-R4-07);
    - provenance (X1-R4-09);
    - ledger rows (X1-R4-12);
    - the ADR claims (part of X1-R4-13).
    All **adopted**.
  - Its near-threshold fix (+100 core-hours of new references) is **refined** to descriptive realised-gain binning of existing cells (X1-R4-09(b)).
  - Its estimate that preparation cost still fits even at 10× is **overridden** by the per-unit scaling in X1-R4-01.
- **Found by the lead only:**
  - the gradient-property fixture gap;
  - the safety grid's position on the critical path;
  - the stale STATUS (R-078 and ADR list) and the index route-(a) and gap note;
  - the fixed-key calibration path as the recommended remedy;
  - the shared-file staleness risk.

## 11. Execution outline for revision 5 (one integration owner)

1. **Resource fix** (X1-R4-01, X1-R4-04, X1-R4-06):
   - fixed-key calibration path and equivalence fixture;
   - genuine L ∈ {40, 400} population and calibration benchmark;
   - coverage size level .20;
   - inventory-time p̂_gate attempt cap;
   - committed projection script.
   *Test:* the new total fits 2,880 core-hours and 15 days, or the revision documents a reviewed redesign.
2. **Acceptance and kernel** (X1-R4-02, X1-R4-05):
   - recovery frames with OC-verified support ≥ .99, and per-seed null "not support, non-zero width";
   - signed-infinity pivots in Python and Go.
   *Test:* updated hand and differential fixtures pass.
3. **Manifest** (X1-R4-03, X1-R4-06, X1-R4-07):
   - .10 side rule and safety identities;
   - M̂ below the grid → not testable;
   - locked-size κ/M̄ and j/TP̄ rates;
   - safety-grid scheduling.
   *Test:* identity and mapping fixtures pass.
4. **Fixtures** (X1-R4-08): gradient property and five verdict boundaries.
5. **Text and reporting** (X1-R4-09 to X1-R4-11):
   - full diagnostic table and the realised-gain near-threshold report;
   - reproduce commands;
   - revision-3 source disclosure;
   - dev source, threats and inventory screen;
   - text residuals.
6. **Ledgers and trackers (Main)** (X1-R4-12, X1-R4-13): internal-study rows, STATUS, index note, and the ADR-014 table carrying every change.
7. **Preparation artefacts:**
   - concrete source, tool, scanner, submodule, policy and dev pins (including n_max);
   - exposure-only manifest and recipe;
   - GUIDE-author isolation record.
   **Recommended:** also draft and submit the OC generator, final evaluator, product harness and isolated GUIDE.
8. **Run and submit.** From the pinned toolchain, run the self-test, the benchmark and the projection script once, with create-once outputs. Record all digests and submit X1 round 5, saved also as ADR-014 `round-3.md`.

## 12. Stage authorisation

| Stage | Status now | Earliest path |
|---|---|---|
| Drafting and synthetic validation, including X1-R4-01's benchmark and the model-free implementation drafts (option G) | **Permitted** (unchanged) | — |
| Preparation | **Not authorised**: 13 open issues (rule 1); resource fit unproven (X1-R4-01); preparation artefacts absent | A zero-issue round on revision 5, reviewed together with the preparation artefacts |
| Inventory and SRS draw | **Not authorised** | After signed preparation outputs and a scoped review |
| Full OC and implementation | **Not authorised** | After inventory locks sizes. Binding gates, the X1-R4-02 parity and acceptance, and locked-size compute (including preparation) must pass before any dev model call. |
| Dev Jev and labels; test scoring and labels | **Not authorised** | Under full authority only |
| Gold unblinding and analysis | **Not authorised** | — |
| ADR-014 adoption | **Not authorised** | The joint zero-issue final round saved in its folder, with the revision-5 table, index and card |
| Default, threshold or product-sampler change | **Not authorised** | A separate signed ADR |

**Next stage authorised:** none beyond drafting and synthetic validation.

**Sign-off statement:** none. Round 4 closes with:
- **X1: 12 issues open**:
  - **0 blocking**;
  - **2 major:** X1-R4-01 and X1-R4-02;
  - **10 minor:** X1-R4-03 to X1-R4-12.
- **ADR-014: 1 issue open** (X1-R4-13, minor). Adoption also waits for the joint zero-issue X1 round.

The reviewed files are those in `reviews/rev4-digests.json` (manifest SHA-256 `01c790837f961cd208fed98b3bd653b514da36e6d8bcee3b2fa4270f88e1876c`, as reported by Main), with the per-file digests in the table above. No preparation, inventory, Jev, labelling or analysis stage is authorised, and ADR-014 is not adopted.
