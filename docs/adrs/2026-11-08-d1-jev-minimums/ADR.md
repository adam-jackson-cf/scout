# ADR-010: Record owner-agreed Jev triage minimums

- **Status:** accepted
- **Design revision:** `docs/scout-design.html`, REQ-98 and Appendix B
- **Decided:** owner direction 2026-11-08; **sources:** R-063

## Decision

- Record `triage.real_defect` minimum precision 0.80 and minimum recall 0.90 as owner-agreed. The owner set these minimums from exploratory data, not explicit costs, and separately confirmed “Confirm: misses cost more”. Numerical values are unchanged; `minimums_provenance` changes from `assumed` to `agreed`.
- Scope is **triage only**, per R-063's verbatim owner clarification. X3/X5/X10 retain REQ-98's own per-decision defaults (0.8/0.8, with 0.9 on each decision's costlier side, status `provisional`) until the owner sets them or a preregistration replaces them. D1's triage part is complete; the non-triage remainder stays open.
- Keep threshold 0.3, status `provisional`, `label_set.summary: none`, and `gating: false`. Agreement on minimums does not establish a tuned or ready threshold, validate ranking, or change REQ-102. Each preregistration still confirms/replaces carried rev-2.1 values and names its costlier side (X10 recall n/a). Defaults substitute only for precision/recall point minimums, not X1 gain or X4/X5 lower bounds.
- Retain the owner-requested **D1-R: review and optimise the Jev triage minimums and threshold**, as part of X1 or its own council-signed preregistration, not X3/X10. D1-R does not change X1's ordinal decision rule. It requires at least 30 human/`agent_reviewed` labels per label under the frozen current-question guide and blind independent two-labeller protocol with adjudication. Legacy `v1_choice`-guide labels do not count toward that floor. Use a repository-split dev/test set, re-check 0.80/0.90 feasibility per stratum, set or replace the threshold and review the minimums with the owner.

## Informed by

- R-063 records the owner's original “use the data not costs” direction, acceptance with a later review/optimise task, triage-only clarification and separate confirmation that misses cost more. It corrects the earlier misdescription of triage minimums as an X3/X5/X10 starting point.
- The owner directed the exploratory Jev-versus-label comparison and accepted it after being told it was unpreregistered and not citable. This was an **owner-directed deviation from REQ-102 and CLAUDE.md General “Jev labelling”** for the R-063 analysis only. It is not precedent, grants no exception to design rule 3, and is not citable as experiment evidence for design changes. This ADR records owner direction under REQ-98, not empirical validation.
- `evaluation/d1-triage-minimums-exploratory/` preserves the exact CSV, regenerated JSON output and reproduction recipe using jev-actions `scripts/tune_threshold.py` (script digest in its README), without new model calls. F1 tuning with minimums 0.01, 5-fold cross-validation × 5 repeats, seed 7, used 51 human-confirmed labels (28 real_defect / 23 false_positive; one cannot_tell excluded) and mean saved Jev 1.14 scores across three repetitions.
- The pooled out-of-fold counts are TP 128, FP 33, FN 12 across 25 tests, each using its own training-selected threshold (0.24333349999999998–0.385). The in-sample best-F1 threshold 0.385 was not the threshold at which those out-of-fold figures were measured and was not shipped. R-063 and the exploratory README give exact metrics and the separately calculated descriptive Wilson intervals on per-repeat averaged counts (precision n≈32.2; recall n=28), with repeated-CV dependence explicitly unaccounted for.
- Limits: fewer than 30 cases per label; no test split; weak nest/webpack strata; unstable fold thresholds (spread >0.05); E2 `v1_choice`-guide labels versus E2 `v2_noul` scores, neither matching production `triage.real_defect@v1` question/state (data-not-instructions clause, `state.facts`, bounded excerpts); exploratory F1 rather than the configured F-beta β=2 policy. These results cannot establish production threshold readiness.

## Changes

- `docs/scout-design.html`: REQ-98 retains provisional status, pending costs/minimums and carried rev-2.1 wording for other decisions; inserts only the owner-set triage pair, data-not-explicit-costs basis, separate miss-cost confirmation, unchanged threshold and X1-scoped D1-R. REQ-169–REQ-175 stay unchanged, including their pending-user-minimums clauses. Appendix B adds ADR-010; ADR-003's historical card body is restored unchanged, with only its relation span naming this amendment.
- `docs/adrs/2026-10-23-decisions/ADR.md`: status line only, naming ADR-010 as amending the triage conditions. The historical Conditions still say “until the user sets costs and minimums”; their body is not rewritten. `docs/adrs/README.md` indexes ADR-010 and names the ADR-003 amendment.
- `internal/decisions/thresholds/defaults.json`: provenance becomes agreed; assumptions distinguish owner-set minimums from untuned threshold, legacy E2 labels from the absent current-version label set, exploratory F1 from F2 and examined development data from future tests. Removes `outcome` from this entry's replacement-label list to align D1-R with REQ-100/101: human or `agent_reviewed` labels under the blind protocol. Audit backlog B-2 specifically requests additional `agent_reviewed` collection; that is a subset, not an exclusion of eligible human labels.
- `internal/decisions/thresholds/thresholds_test.go`: asserts agreed provenance while preserving 0.8/0.9 minimums, 0.3 threshold, F-beta β=2, no labels, raw score space, provisional status and ungated behaviour. The supplied proof file is frozen and unchanged by this revision. Its existing comment's “because a miss costs more” is not the rationale for setting the pair: the owner set it from data and separately confirmed the costlier side, as recorded above.
- `.omp/test-map.json`: maps the renamed proof identity to the changed embedded-default configuration and behaviour; no guard is weakened.
- `docs/research-ledger.md`: R-063 records the verbatim scope/cost answers, correct pooled out-of-fold semantics, interval method/denominators, limitations, deviation and examined-data quarantine. `evaluation/d1-triage-minimums-exploratory/{README.md,scores.csv,tuning.json}` preserves the reproduction record. These exploratory artifacts are not signed-off experiment evidence.
- `docs/plans/2026-11-06-remaining-work.md`: splits completed D1-T from open D1, restores D1→Track C, adds the D1-R→X1/dedicated-preregistration relation, and records entry, quarantine and later work. `docs/STATUS.md` aligns decisions, label-analysis disclosure, outstanding work and ledger pointers.
- `docs/jev-adapter-audit.md`: governed Fixed row 22 records agreed provenance and current-version label/tuning gaps, not a nonexistent visible TODO; audit status and backlog B-2 are unchanged.
- This folder's `ADR.md` and review/response records capture the decision and council iterations.

## Conditions

- Adoption requires an innovation-council round with zero issues and explicit sign-off naming the reviewed file digests. The signed set is prepared in adopted form (route (a), ADR-009 round-13 step 5 precedent); the actual adoption record is outside it. No signed bytes change to mark adoption.
- The R-063 comparison was the owner-directed deviation from REQ-102 / CLAUDE.md General “Jev labelling” described above, for that exploratory analysis only, not precedent or an exception to design rule 3. This ADR authorises no further comparison or tuning outside the existing experiment rules.
- All 52 `evaluation/jev-triage-labels-v1` cases and the E2 `v2_noul` score files are **examined data**. Any X1 or D1-R preregistration must disclose them, may use them only for development, and must draw confirmatory test data from unexamined cases. Legacy `v1_choice`-guide labels do not satisfy D1-R's current-question entry floor.
- D1-R's label floor and preregistration are conditions for later work, not prerequisites for this provenance-only amendment.

## Review

- This summary is among the signed files, so it does not contain its own outcome or digests.
- Review outcome and signed SHA-256 digests: the explicit zero-issue final `round-<N>.md` in this folder (highest N). Earlier rounds and responses preserve issues and fixes. Adoption is recorded outside the signed set after that sign-off, without editing these signed files.

## Records in this folder

- `ADR.md`: this summary.
- `round-<N>.md`: council review rounds, issues and final sign-off with digests.
- `round-<N>-response.md`: per-issue fixes and review submission conditions.
