# Round 2: ADR-010, record owner-agreed Jev triage minimums

Date: 2026-11-08. This round follows design rules 1–3 and the ledger rules in `CLAUDE.md`, and the ADR rules in `docs/adrs/README.md`. The innovation council lead ran it with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review. It checked whether each round-1 fix appears in the branch bytes, whether the owner answers are applied consistently, the adopted-form route, the claims about `thresholds_test.go` and `.omp/test-map.json`, and whether the fixes introduced new contradictions.
- **`experiment-peer` (Opus):** evidence and inference review. It covered R-063 and `evaluation/d1-triage-minimums-exploratory/`, recomputed the CSV, `tuning.json` and the Wilson intervals by hand, and checked the quarantine, the deviation statement, how far the miss-cost answer reaches, and how `defaults.json` relates to the label rules.

The review was read-only. It made no model calls outside the council and did not touch `.codegraph`.

**The main agent computed the digests; the council did not.** The lead and both peers had read-only tools and no shell. The README's own digests for `scores.csv` and `tuning.json` match the main agent's.

| File | SHA-256 (main agent) |
|---|---|
| `docs/scout-design.html` | `2529fa3f23df417a8699199aa0bd8e8ee895089191d2f96ad39b3b36eb52616e` |
| `docs/adrs/2026-11-08-d1-jev-minimums/ADR.md` | `a669a04332456d43755de3c6fdfcca515082677270ba54f98881bca46e0926b6` |
| `docs/adrs/README.md` | `77f8b0b4b04cf2c392996c6e06d6acc6455b4c076c9c7852ee0b289cb6a27197` |
| `docs/STATUS.md` | `d6427407cffc2a6db4a61de472f4f8f1fcaa0b23d57140cd39e729a1bd1d3c3a` |
| `docs/jev-adapter-audit.md` | `92aa0487580ba94987431aabfa44106e5a0ed1d4918498722a66c33930f7d118` |
| `docs/adrs/2026-10-23-decisions/ADR.md` | `6b92c168364db09db4095885858e9a86db450ec9dc7853a7f9cefb785206e2bb` |
| `docs/research-ledger.md` | `dd013197031455a8031501757fe7b03a4394efdcae4e0472c34b0ffee0ffaa92` |
| `docs/plans/2026-11-06-remaining-work.md` | `a46697860e69bed0580bb3c1ee58b06db8954b75fcb08baab75fdfa1923785fa` |
| `internal/decisions/thresholds/defaults.json` | `8c68471f7ab22bbd786619ed055dad881e323b4f39ac274e4aec2f9ec1fe27b8` |
| `internal/decisions/thresholds/thresholds_test.go` | `fba178e3bd6469e3d932af98c65ae239c85bd5e1cd539dfb39e2473712336b91` |
| `.omp/test-map.json` | `96de43c74f2769c69c83e84e68a6701d3c83a991a818d47bedb73811f1f703db` |
| `evaluation/d1-triage-minimums-exploratory/README.md` | `37f15abec8331f50ecbf8c864d5ad368402dafc7219d10f5b157057b1c7fe3d0` |
| `evaluation/d1-triage-minimums-exploratory/scores.csv` | `c7679a2da9a30b3b905e4f8be3c81f4eed6ad459e85426db719a356f3818c2d2` |
| `evaluation/d1-triage-minimums-exploratory/tuning.json` | `03402fc8a1d183338ccf58ce7c723584a44f5fd615944b12509f519e0f587f9b` |

## Scope

- **Under review:** worktree `/Users/adamjackson/Projects/work/scout-d1`, branch `agent/d1-minimums`, head `3ca453a78d98fec792c33e54a2a4c6ba8d78e944`. Base is `main` at `04d5e632`; the main worktree `/Users/adamjackson/Projects/work/scout` is at that commit and was used for every base comparison.
- **Round-1 record and response:** `round-1.md` (13 issues) and `round-1-response.md`.
- **Files read:**
  - ADR-010 `ADR.md`.
  - `scout-design.html`: REQ-98 at `:580` (base and branch, compared in full); REQ-102 at `:584`; REQ-183 at `:720`; Appendix B rules at `:729-733`; the ADR-003 card at `:806-829` (base and branch); the ADR-009 and ADR-010 cards at `:948-990`.
  - The ADR-003 `ADR.md` status line and Conditions at `:42-47` (base and branch).
  - `docs/adrs/README.md`; the ADR-009 `ADR.md` as the route (a) precedent.
  - R-063 at `research-ledger.md:71`, read in full.
  - `docs/STATUS.md` at `:1-10,17,123,127,129,144`.
  - The plan at `:17-36,80-82,97`.
  - `jev-adapter-audit.md` row 22 at `:63` (base and branch) and B-2 at `:84`.
  - `defaults.json` (base and branch).
  - `thresholds_test.go` (base and branch; function line ranges).
  - `.omp/test-map.json` (base and branch) and `.omp/guardrails.json`.
  - The exploratory `README.md`, `scores.csv` (51 rows) and `tuning.json` (fold thresholds, `tuned_heldout`, `default_0.5_heldout`, strata, warnings).
  - `evaluation/jev-triage-v1/raw/`.
  - `history://Main:1395-1515` and `:1770`; `history://D1Impl`.
- **Not reopened:** decisions settled in ADR-000…ADR-009, and the round-1 dispositions under "Considered and not raised" unless new evidence appeared.

## Round-1 issues: verification

| Issue | Status | Evidence / remainder |
|---|---|---|
| ADR010-R1-01 | **Met** (provenance caveat: R2-05) | Triage-only scope is stated in REQ-98 (`html:580`), ADR.md:10, the card at `:975-977`, STATUS:17 and :127, R-063, plan:80-82 and audit row 22. D1→C is restored (plan:32). REQ-169–REQ-175 are unchanged against base, with "pending the user's minimums" still at `:686` and `:688-692`. |
| ADR010-R1-02 | **Met** after the council's disposition of the test comment (below); one unlisted deletion is R2-03 | REQ-98 again has "status `provisional`", "until user costs/minimums are set" and "confirms/replaces carried rev-2.1 values". The triage clause gives the data basis, not explicit costs, and the separate miss-cost confirmation. ADR.md:25 correctly quotes "until the user sets costs and minimums". |
| ADR010-R1-03 | **Met** | ADR.md:17 and :37, the card at `:980`, R-063, STATUS:123, plan:82 and README:3 all name REQ-102 and CLAUDE.md General "Jev labelling". They describe the owner's direction and informed acceptance, and limit the deviation to the R-063 analysis: no precedent and no design-rule-3 exception. This matches `history://Main:1411,1446-1454,1507-1513`. |
| ADR010-R1-04 | **Met** | All 52 cases and the E2 `v2_noul` score files are named as examined data in ADR.md:38, card:986, plan:82, R-063, STATUS:123, `defaults.json` assumptions[3] and the README. The `v1_choice` labels are excluded from D1-R's entry floor. |
| ADR010-R1-05 | **Met** | "Part of X1 or its own" and "does not change X1's ordinal decision rule" appear in REQ-98, ADR.md:12, card:977, plan:82 and STATUS:127. |
| ADR010-R1-06 | **Met in substance; residual wording and limits: R2-04** | The CSV, JSON and recipe are committed and linked. The figures are recomputed below. The interval method and denominators are stated. The question/state and F1/F2 mismatches are disclosed. The owner's answers are quoted (provenance: R2-05). |
| ADR010-R1-07 | **Met** | The ADR-003 card at `html:806-829` matches base apart from its `rel` span. ADR-003 `ADR.md` changes only its status line at `:3`, and Conditions `:42-47` are identical. The README index names the amendment. |
| ADR010-R1-08 | **Partly met: R2-01, R2-02** | Route (a) was chosen. ADR-010 is accepted, the card is `s-acc`, the footer points to the review folder, the README says "accepted", and STATUS:127 says ADR-000…ADR-010. Still in pending form: STATUS:7, card:983 and ADR.md:36. |
| ADR010-R1-09 | **Met** | Row 22 records `agreed` (owner, 2026-11-08, R-063) and the separate miss-cost confirmation. It names the missing current-question labels and the D1-R tuning still to come; the false "visible TODO" is gone. The row is listed in ADR Changes and in the signed set. |
| ADR010-R1-10 | **Met** | assumptions[0] and [2] are reworded as required; `label_set.summary` is still `"none"`. |
| ADR010-R1-11 | **Met as written** | This fix now conflicts with route (a); see R2-02. |
| ADR010-R1-12 | **Met for the named pointers; residual: R2-01, R2-06** | STATUS:17 and :144 and plan:97 are updated, and a D1-R node exists. |
| ADR010-R1-13 | **Met; residual "inserts only" wording: R2-03** | Changes lists every file, the Records section exists, the label sources are aligned, and B-2's narrower request is explained. |

### Disposition of the frozen `thresholds_test.go` comment (R1-02)

The comment at `thresholds_test.go:28` still reads "The owner agreed precision 0.8 and recall 0.9 because a miss costs more." The council **accepts ADR.md:27 as the correction**, and closes the comment part of R1-02 without a byte change to the test. The reasons:

- Both facts in the comment are now true: the owner agreed the pair, and the owner confirmed misses cost more (R-063). Only the causal "because" is wrong.
- ADR.md:27 is governed and signed. It names the comment and states the correct rationale: the owner set the pair from data and confirmed the costlier side separately.
- Audit row 22, also governed, cites both the tests and ADR-010.
- No assertion depends on the comment. The assertions match ADR.md:27: `"agreed"` at `:25-26`, and 0.8/0.9, F-beta β = 2, 0.3, ungated, no labels, raw score space and provisional status at `:19-37`.
- A code comment is not among the documents design rule 1 governs.

The "frozen" status comes from the main agent's proof workflow ("Frozen proof (DO NOT EDIT)", `history://D1Impl`), not from a repository rule. A later prover may correct the comment without a design round.

## Recomputation (checked clean)

- **`scores.csv`:** 51 data rows, 28 labelled 1 and 23 labelled 0. Every row is `dev`, model `typesafe/jev-1.14-20261014`, question version `e2-v2_noul`. The row order matches the non-`cannot_tell` human rows of `labels.jsonl`. The peer spot-checked two means of three scores.
- **`tuning.json`:** `tuned_heldout` is TP 128, FP 33, FN 12, so P = 128/161 = 0.7950310559006211 and R = 128/140 = 0.9142857142857143. `default_0.5_heldout` is TP 95, FP 10, FN 45. There are 25 of 25 folds, and the fold thresholds range from 0.24333349999999998 to 0.385.
- **Wilson intervals (z = 1.96), recomputed by the lead and the peer:**
  - Precision 25.6/32.2: [0.628092, 0.899076].
  - Recall 25.6/28: [0.754973, 0.973633].
  - Both equal the README's figures.
- **Production mismatch:** confirmed. `triage.go` has the data-not-instructions clause, `state.facts` and the 8-line/600-byte excerpt bounds; `questions.py:9-12` has none of them.

## Issues

| ID | Severity | Location | Problem | Required fix |
|---|---|---|---|---|
| ADR010-R2-01 | minor | `docs/STATUS.md:7`; also `:1` and `:5` | **The adopted-form STATUS contradicts itself.**<br>• Line 7, the main current-design pointer, still says "decisions ADR-000…ADR-009" and cites only ADR-009's signed digest.<br>• `:127` says "ADR-000…ADR-010".<br>• The header dates say "2026-11-07", but the section holds 2026-11-08 state.<br>This is left over from R1-08 and R1-12. | • Change line 7 to "ADR-000…ADR-010".<br>• Name `docs/adrs/2026-11-08-d1-jev-minimums/` as the latest amendment's review, keeping ADR-009 as the restructure provenance.<br>• Update the header dates. |
| ADR010-R2-02 | minor | `scout-design.html:983`; ADR.md:36 | **The accepted-form bytes still say adoption is pending.**<br>• The ADR-010 card is `s-acc`/accepted but keeps `<h4>Before adoption</h4>` with "Innovation council must sign off…".<br>• ADR.md:36 says "Adoption requires an innovation-council round…".<br>These files are signed and frozen, so the accepted record would show an unmet gate permanently. No other accepted card has a "Before adoption" section, and the route (a) precedent, ADR-009, has none. R1-11's instruction was written for proposed form and conflicts with R1-08 route (a); this fix supersedes it. | • Remove the card's "Before adoption" section; the footer already points to the review. Or replace it with "Open conditions" holding the D1-R conditions.<br>• Move ADR.md:36 into the Review section in past-tense, ADR-009 form: adopted on the zero-issue final round named there; adoption recorded outside the signed set. |
| ADR010-R2-03 | minor | ADR.md:24, :26; `html:580`; `defaults.json` assumptions[3] | **ADR Changes leaves out two deletions.**<br>• ADR.md:24 says REQ-98 "inserts only" the triage clause. In fact base "(`triage.real_defect`: recall [ASSUMPTION])" is deleted. The deletion is justified by the owner's confirmation, but it is not listed.<br>• Base assumptions[3] said "Agent labels come from two independent sessions that never see Jev's answer". The branch drops it, leaving only "blind two-labeller protocol". ADR.md:26 does not mention this.<br>This is left over from R1-02 ("list every REQ-98 wording change") and R1-13. | • List both deletions in ADR Changes, and replace "inserts only" with an accurate description.<br>• In `defaults.json`, either restore "never see Jev's answer" or cite REQ-100 ("never showing answer/rank/score/stratum/prior verdict"). |
| ADR010-R2-04 | minor | Exploratory `README.md:35,39`; ADR.md:19-20; R-063 (`research-ledger.md:71`) | **The evidence record overstates one point and leaves out limits.**<br>(a) **0.385:** the README says the out-of-fold figures were "**not measured at 0.385**", and ADR.md:19 and R-063 say 0.385 "was not the threshold at which those out-of-fold figures were measured". `tuning.json` `fold_thresholds` shows 0.385 in 13 of 25 fold tests, 0.26 in 9 and 0.24333349999999998 in 3.<br>(b) **Limits omitted:**<br>  • the axum stratum is as weak as nest: in-sample precision 0.5 (TP 1, FP 1; n = 8 with one positive), yet only nest and webpack are named;<br>  • scores are the mean of three Jev calls per case, while production asks one request per unit (REQ-88). Averaging reduces score noise, so a threshold tuned on means may not transfer;<br>  • pooled out-of-fold precision, 0.7950310559006211, is below the agreed 0.80 minimum. The pair was not shown to be attainable together even on these data, although K1 told the owner "The data points to 0.80 precision" (`history://Main`). | • Reword (a) in all three places: "pooled across per-fold thresholds (0.385 in 13, 0.26 in 9, 0.24333349999999998 in 3 of 25 fold tests); not measured at one fixed threshold".<br>• Add to the limits in the README, ADR.md:20 and R-063:<br>  – the weak axum stratum;<br>  – mean of 3 calls versus one production call;<br>  – out-of-fold precision 0.795 below the agreed 0.80, so joint attainability is not shown and D1-R re-checks it. |
| ADR010-R2-05 | minor | R-063 (`research-ledger.md:71`); ADR.md:9 | **The owner's answers cannot be verified and appear to be option text.**<br>• The scope and miss-cost answers came through an `ask` call (`history://Main:1770`, "Correction: earlier I said the X3, X5 and X10 preregistrati…"). Its question, options and answer are not shown in any readable transcript.<br>• "(a) Triage only: P 0.80 / R 0.90 applies to triage.real_defect only…" and "Confirm: misses cost more" read as assistant-written option labels and descriptions **[INFERENCE]**. R-063 presents them as the owner's words, and ADR.md:9 says the owner "confirmed 'Confirm: misses cost more'".<br>• As a result, the repository cannot show that R1-01's condition was met (the owner asked "with REQ-98 described correctly"). It also cannot show whether the miss-cost question was limited to triage.<br>• The owner's first direction, "you should use jev-actions to help define its starting points for precision and recall, you can create labelled datat or use runs we have already actioned" (`history://Main:1411`), is not quoted.<br>• Precedent: R-061 records the options shown and the one chosen. | In R-063:<br>• quote the scope question and the miss-cost question verbatim, with every option offered;<br>• mark which option the owner selected, and quote any free text the owner added;<br>• quote the `:1411` direction.<br>Reword ADR.md:9 to match, for example: "selected 'Confirm: misses cost more' for `triage.real_defect`". |
| ADR010-R2-06 | minor | plan:33; ADR.md:30 | **The D1-R edge in the plan graph points the wrong way.**<br>In this graph `A --> B` means A precedes or blocks B (`D1 --> C`, `C --> B4`). `D1R[...] --> X1[X1 preregistration or dedicated triage preregistration]` therefore says D1-R blocks the X1 preregistration. In fact D1-R needs a signed preregistration (plan:82: "Requires a council-signed preregistration (part of X1 or its own)") and "does not block drafting Track C preregistrations". | • Reverse the edge, for example `X1P[X1 or dedicated triage preregistration signed] --> D1R`, or label it as containment, for example `-. part of .->`.<br>• Make ADR.md:30 describe the same relation. |

**Count: 6 issues (0 critical, 0 major, 6 minor).**

## Checks that found no issue

- **Values unchanged.** Base and branch `defaults.json` both have threshold 0.3, `min_precision` 0.8, `min_recall` 0.9, F-beta β = 2, `gating: false`, `score_space: raw`, `status: provisional` and `label_set: none`. Only `minimums_provenance` and the assumption text differ.
- **REQ-98 kept clauses:** "status `provisional`", "until user costs/minimums are set", point minimums only, "confirms/replaces carried rev-2.1 values", "names the costlier side (X10 recall n/a)".
- **REQ-183.** Removing the triage [ASSUMPTION] marker is covered by REQ-98's own replacement condition ("until user … minimums are set"), recorded by this ADR.
- **Miss-cost scope.** No document applies the miss-cost confirmation or the triage pair to X3/X5/X10. Plan:81 keeps "0.9 on each decision's own costlier side". (Provenance of the answer: R2-05.)
- **ADR-003.** Card body and `ADR.md` body are identical to base apart from the `rel` span and the status line. The README status names "amended by ADR-010 (triage minimums)".
- **`.omp/test-map.json`.** Base has no mapping for either test name. The branch adds one for `…TheAgreedMinimums`, the renamed test, so "maps the renamed proof identity" is accurate. No guard is weakened.
- **`thresholds.go`.** It requires only a non-empty `minimums_provenance`, and `assumptions` is free text. Removing `outcome` cannot break the schema or the tests.
- **Placement under `evaluation/`.** Consistent with the unsigned development record `evaluation/x2d-postrun-tuning/` (R-055). The README states at the top that the record is not signed off.
- **External script pin.** `tune_threshold.py` is pinned by SHA-256 `2f850933…2e4c`, the same external-script pin X2d used in a signed-off experiment.
- **Hypothesis ledger.** No product hypothesis was proposed, tested or changed.

## Considered and not raised

- **Is the `defaults.json` edit a REQ-102 "default change"?** (peer 9). Round 1 already decided this ("not needed for this change. Optional"). No value changes, and REQ-98 reserves minimums to the user. The assistant's earlier session remark (K2.4) is not a rule. An explicit carve-out sentence stays optional.
- **Naming CLAUDE.md experiment rules 1 and 3 as also set aside** (peer 10). The record says it is not evidence, and design rule 3 bars citing it. The named prohibition (REQ-102 and "Jev labelling") matches the R-055 precedent. Adding them is optional and can go with the R2-04 edits.
- **Commit a copy of `tune_threshold.py`** (peer 12). The SHA-256 pin follows the X2d precedent.
- **Nits, optional while editing:**
  - README:28 hard-codes the `scout-d1` worktree path (README:21 gives the relative form).
  - Audit row 22's "`thresholds_test.go:13-91`" is a stale range that predates this branch (the file is 123 lines; it was 120 at base).
  - B-2's "`defaults.json` TODO" pointer is stale; B-2 is backlog, not governed, and the problem predates this branch.
  - The ADR index rows for ADR-003 and ADR-010 are not column-aligned.
  - Card:985 writes "agent-reviewed" where the code value is `agent_reviewed`.
  - The quarantine names score files without paths. The quarantine of all 52 cases already covers any score on them.
- **Threshold 0.385, recommended in K1, was not adopted.** Setting the threshold is not the owner's choice under REQ-102. The ADR keeps 0.3 and does not attribute it to the owner.

## Peers

- **`innovation-challenger`** raised 2 minor items:
  - STATUS:7 (became R2-01);
  - the R1-02 test-comment exception, which it recommended the council accept explicitly (adopted above).
  It confirmed R1-03, R1-04, R1-05, R1-07, R1-09, R1-10 and R1-11. It verified the test assertions and the test-map wording, and found that the later owner answers cannot be seen in `history://Main`.
- **`experiment-peer`** raised 12 minor items and 8 nits. Its findings became:
  - R2-04: (a) the 13-of-25 fold count; (b) axum, mean-of-3 and 0.795 below 0.80;
  - R2-05, option text;
  - R2-03, the `defaults.json` deletion;
  - R2-06, the edge direction;
  - R2-02, raised independently of the lead.
  Its items 9, 10 and 12 and the nits were declined, as listed above. It recomputed all the figures and found them correct.
- **Lead:** R2-02 and the REQ-98 part of R2-03, both found independently. The lead recomputed the Wilson intervals, compared the ADR-003 card and body against base, and settled the test-comment disposition.
- **Where the peers changed the outcome:**
  - The challenger added R2-01.
  - The peer raised the lead's suspicion that the answers were option text to an issue (R2-05), with the `:1411` omission.
  - The peer added R2-04's measured details and R2-06.
  - The peer argued that R1-02 was "not met" because the comment was unchanged. The lead, with the challenger, decided instead to accept the ADR's documented correction.

## Verdict

**6 issues open (6 minor).** Not signed off. ADR-010 must not be adopted.

### Required with the next submission

1. Fixes for R2-01–R2-06, and a per-issue `round-2-response.md`. For R2-05, quote the `ask` question and options exactly from the main session.
2. The main agent's SHA-256 for every signed-set file, in adopted form.
3. `git diff --stat 04d5e632..HEAD`, which was not supplied this round, with confirmation that every changed file other than the review records is in the signed set.
4. Re-confirmation that the ADR-003 card body is byte-identical to base apart from its `rel` span.

---

## Council memo

1. **Recommendation.** Do not adopt yet. Make the six minor text fixes, all in files already in the signed set:
   - STATUS pointer and dates (R2-01);
   - remove pending-adoption wording from the accepted bytes (R2-02);
   - list the two deletions (R2-03);
   - correct the 0.385 wording and complete the limits (R2-04);
   - record the `ask` question and options verbatim in R-063 (R2-05);
   - fix the D1-R edge (R2-06).
   Then resubmit for round 3. Close R1-02's test-comment item through ADR.md:27 without editing the frozen proof.
2. **Alternatives.**
   - **(A) Fix R2-01–R2-06 and resubmit (recommended).** The text changes are small and close every residual.
   - **(B) Also edit the `thresholds_test.go` comment.** Removes the last stale sentence, but means editing the frozen proof and renewing its digest, which adds churn for no change in behaviour.
   - **(C) Sign off now and fix in a follow-up ADR.** Not allowed: design rule 1 needs zero issues, and these files would be frozen with known errors.
   - **(D) Drop the exploratory record from the ADR and cite only the owner direction.** This removes R2-04, but loses R1-06's reproducibility, and the ledger rule still needs the source.
3. **Evidence.**
   - `history://Main:1411,1446-1454,1507-1513,1770`.
   - `tuning.json` `fold_thresholds` (13 at 0.385, 9 at 0.26, 3 at 0.24333349999999998), `strata.axum`, `tuned_heldout`.
   - `scores.csv` (51 rows, 28/23).
   - Base and branch `html:580`; `html:983`, `:806-829`, `:948-990`.
   - ADR.md:9, :19-20, :24, :26-28, :30, :36.
   - STATUS:1, :5, :7, :127; plan:32-33, :82.
   - Base and branch `defaults.json`; `thresholds_test.go:13-38`.
   - Base and branch `.omp/test-map.json:4`.
   - R-061 (precedent for recording options); ADR-009 `ADR.md` (accepted-form precedent).
4. **Risks.**
   - **R2-05 may not be resolvable:** the `ask` record may show the miss-cost question was not limited to triage, or REQ-98 described wrongly. In that case the scope must go back to the owner.
   - **The council relies on the main agent's digests** and on the claim that the CSV rebuilt byte for byte; it could only spot-check them.
   - **R2-04's limits are disclosures, not blockers.** Nothing is gated on these minimums while ranking is ordinal (REQ-98).
   - **Everything is reversible** by revert.
5. **Execution outline.**
   1. Retrieve the `ask` question, options and selections from the main session, and rewrite R-063's quotes and ADR.md:9 (R2-05).
   2. Fix STATUS:1, :5 and :7 (R2-01).
   3. Remove or rename card:983, and recast ADR.md:36 into Review form (R2-02).
   4. Update ADR Changes for the REQ-98 [ASSUMPTION] deletion and the `defaults.json` "never see Jev's answer" deletion, restoring the phrase or citing REQ-100 (R2-03).
   5. Reword the 0.385 statement and add the axum, mean-of-3 and 0.795-below-0.80 limits in the README, ADR.md:19-20 and R-063 (R2-04).
   6. Reverse or relabel plan:33 and align ADR.md:30 (R2-06).
   7. The main agent recomputes all signed-set digests and the diff stat, writes `round-2-response.md`, and requests council round 3.
