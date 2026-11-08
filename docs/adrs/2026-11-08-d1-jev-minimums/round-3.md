# Round 3: ADR-010, record owner-agreed Jev triage minimums

Date: 2026-11-08. This round follows design rules 1–3 and the ledger rules in `CLAUDE.md`, and the ADR rules in `docs/adrs/README.md`. The innovation council lead ran it with two independent peers, and neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review. It checked each R2 fix against round 2's required fix and reconfirmed all thirteen R1 items against base. It also checked the adopted-form route, the ADR-003 card and summary against base, and the test assertions, and looked for new contradictions and HTML defects.
- **`experiment-peer` (Opus):** evidence and inference review. It focused on R2-03, R2-04 and R2-05. It recomputed `scores.csv`, `tuning.json` (fold thresholds, `tuned_heldout`, `strata.axum`) and the Wilson intervals. It also spot-checked three-call means against the raw E2 score files, and compared the quoted ask record against REQ-98 and the rest of the signed set.

The review was read-only. It made no model calls outside the council, ran no label comparison or threshold computation, and did not touch `.codegraph`.

**The main agent computed the digests; the council did not.** The lead and both peers had read-only tools and no shell. The main agent ran `shasum -a 256` at head `44f37ff0416fdf1d69a168b08cc216165bd4e755`. The same digests appear in `history://D1Impl` (final turn) next to `git diff --stat 04d5e632..HEAD`. The exploratory README's own digests for `scores.csv` and `tuning.json` match the table.

| File | SHA-256 (main agent) |
|---|---|
| `docs/scout-design.html` | `26b55ea82237023769d7317281e48ac32fd9106284176df10d369c941ca9d7a2` |
| `docs/adrs/2026-11-08-d1-jev-minimums/ADR.md` | `3b3eaba00d63d0dc777024aac0361b6ae50b50323d8f25cdb36bb506e21fa063` |
| `docs/adrs/README.md` | `77f8b0b4b04cf2c392996c6e06d6acc6455b4c076c9c7852ee0b289cb6a27197` |
| `docs/STATUS.md` | `d2c80854f486a6060f65e63d8aa5f078978f333b9a81a2848b50d43c5aab8427` |
| `docs/jev-adapter-audit.md` | `92aa0487580ba94987431aabfa44106e5a0ed1d4918498722a66c33930f7d118` |
| `docs/adrs/2026-10-23-decisions/ADR.md` | `6b92c168364db09db4095885858e9a86db450ec9dc7853a7f9cefb785206e2bb` |
| `docs/research-ledger.md` | `ffc274b82b1a45ab7d7c6fdf1a8b2431b0f496ec9bae3ddb48c15a1923351ad1` |
| `docs/plans/2026-11-06-remaining-work.md` | `9137529ca827b4147d7ffc43bd06219b2cac82a9782b27c7f26d7a202031013f` |
| `internal/decisions/thresholds/defaults.json` | `5dba9d910331bd5c35c1af1143e35617d4c5a5b6dc99323d3f0ed7e6180c8651` |
| `internal/decisions/thresholds/thresholds_test.go` | `fba178e3bd6469e3d932af98c65ae239c85bd5e1cd539dfb39e2473712336b91` |
| `.omp/test-map.json` | `96de43c74f2769c69c83e84e68a6701d3c83a991a818d47bedb73811f1f703db` |
| `evaluation/d1-triage-minimums-exploratory/README.md` | `3759df7e2e01ff334ba0711f4259a9df43f8ed4ba943548e4da5deb1f91eb1c9` |
| `evaluation/d1-triage-minimums-exploratory/scores.csv` | `c7679a2da9a30b3b905e4f8be3c81f4eed6ad459e85426db719a356f3818c2d2` |
| `evaluation/d1-triage-minimums-exploratory/tuning.json` | `03402fc8a1d183338ccf58ce7c723584a44f5fd615944b12509f519e0f587f9b` |

Five digests are the same as in round 2: `docs/adrs/README.md`, `docs/jev-adapter-audit.md`, `docs/adrs/2026-10-23-decisions/ADR.md`, `thresholds_test.go` and `.omp/test-map.json`, plus `scores.csv` and `tuning.json`. The other seven changed in the round-2 fixes, and this round reread them in full.

## Scope

- **Under review:** worktree `/Users/adamjackson/Projects/work/scout-d1`, branch `agent/d1-minimums`, head `44f37ff0416fdf1d69a168b08cc216165bd4e755`. Base is `main` at `04d5e632`; the main worktree `/Users/adamjackson/Projects/work/scout` is at that commit and was used for every base comparison.
- **Prior records:** `round-1.md` (13 issues), `round-1-response.md`, `round-2.md` (6 minor issues) and `round-2-response.md`.
- **Submission items round 2 asked for:**
  1. **Per-issue response:** `round-2-response.md`.
  2. **Fourteen digests:** in the table above.
  3. **Diff stat:** `git diff --stat 04d5e632..HEAD` (`history://D1Impl`, final turn) shows **18 files changed, +835 −64**. These are the fourteen signed files plus four review records (`round-1.md`, `round-1-response.md`, `round-2.md`, `round-2-response.md`). Every changed file outside the review records is in the signed set.
  4. **ADR-003 card:** the main agent confirms it is byte-identical to base apart from its `rel` span (`round-2-response.md`), and the challenger confirmed this independently.
- **Owner ask record (R2-05):** R-063 quotes it. The main agent confirms the quote is verbatim from its `ask` call of 2026-11-08, and that the owner selected "(a) Triage only" and "Confirm: misses cost more". The compacted `history://Main` shows only the start of that call: `ask({"questions":[{"id":"scope","header":"D1 scope","question":"Correction: earlier I said the X3, X5 and X10 preregistrati…`. That prefix matches R-063's Q1 header and opening words. Q2, the options and the selections cannot be seen in any readable transcript. **The council accepts the verbatim record on the main agent's attestation, as it accepts the digests.**
- **Files read:**
  - ADR-010 `ADR.md`, in full.
  - `scout-design.html`:
    - REQ-98 at `:580`, base and branch, compared in full;
    - REQ-88 at `:570`, REQ-100 at `:582`, REQ-101 at `:583`;
    - REQ-169–REQ-175 at `:686-692`;
    - the ADR-003 card at `:806-829`;
    - the ADR-009 and ADR-010 cards at `:948-989`.
  - The ADR-003 `ADR.md` status line.
  - `docs/adrs/README.md`.
  - The ADR-009 `ADR.md` Review section at `:36-39` and ADR-009 `round-13.md` step 5 at `:212`.
  - R-063 at `research-ledger.md:71`, in full.
  - `docs/STATUS.md` at `:1-20,123,127,129,144`.
  - The plan at `:14-36,76-97`, base and branch.
  - `jev-adapter-audit.md` table at `:40-82`, every row, base and branch.
  - `defaults.json`, in full.
  - `thresholds_test.go` assertions at `:19-37`.
  - The exploratory `README.md` (in full), `scores.csv` and `tuning.json`.
  - The jev-actions `building-eval-data.md` guidance at `:85-102`.
  - `history://D1Impl`, including the main agent's brief and final turn; `history://Main`, compacted.
- **Not reopened:**
  - decisions settled in ADR-000…ADR-009;
  - the round-1 and round-2 dispositions, including round 2's acceptance of ADR.md:27 as the correction to the frozen test comment, unless new evidence appeared (see "Considered and not raised", R3-P1).

## Round-2 issues: verification

| Issue | Status | Evidence |
|---|---|---|
| ADR010-R2-01 | **Met** | <ul><li>`STATUS.md:1` and `:5` both say 2026-11-08.</li><li>`:7` reads "decisions ADR-000…ADR-010 in `docs/adrs/`". It names `docs/adrs/2026-11-08-d1-jev-minimums/` as the latest amendment's review, and keeps "Restructure provenance: ADR-009, adopted 2026-11-06" with its own final round.</li><li>This agrees with `:127` ("ADR-000…ADR-010").</li><li>The other 2026-11-07 dates (`:16` and others) are dates of historical observations, not of the document's state.</li><li>No pointer in STATUS still treats ADR-009 as the latest.</li></ul> |
| ADR010-R2-02 | **Met** | <ul><li>The ADR-010 card at `html:971-989` is `s-acc`/accepted and has no "Before adoption" section. Its D1-R conditions sit under a separately labelled `<details>` ("D1-R entry and examined-data conditions", `:983-987`), and its footer points to the review folder (`:988`).</li><li>ADR.md Conditions (`:34-38`) no longer contains the pending-adoption sentence.</li><li>ADR.md Review (`:40-43`) delegates outcome and digests to "the explicit zero-issue final `round-<N>.md` in this folder (highest N)", and records adoption outside the signed set. It cites route (a) and ADR-009 `round-13.md` step 5, which exists at `:212`: "Mark ADR-009 adopted only through files that are not in the signed set".</li><li>This matches the ADR-009 precedent (ADR-009 `ADR.md:38-39`; its card at `html:948-969` has no adoption gate) and `docs/adrs/README.md:3`.</li></ul> |
| ADR010-R2-03 | **Met** | <ul><li>ADR.md:24 now says REQ-98 "removes the old `triage.real_defect` recall `[ASSUMPTION]` parenthesis", which the lead confirmed against base `html:580`. The inaccurate "inserts only" is gone.</li><li>ADR.md:26 records the deletion and restoration of the blindness sentence.</li><li>`defaults.json:21` reads "Agent labels come from two independent sessions that never see Jev's answer (REQ-100)." That is base's sentence word for word, plus the citation.</li><li>REQ-100 (`html:582`) supports it: "never showing answer/rank/score/stratum/prior verdict".</li></ul> |
| ADR010-R2-04 | **Met** | <ul><li>(a) The exploratory `README.md:35`, ADR.md:19 and R-063 all say the figures are "pooled across per-fold thresholds (0.385 in 13, 0.26 in 9, 0.24333349999999998 in 3 of 25 fold tests), not measured at one fixed threshold". The council counted `tuning.json` `fold_thresholds` and got the same: 0.385 at positions 2, 4, 5, 7, 8, 10, 12, 14, 15, 16, 18, 21 and 25; 0.26 at 1, 6, 11, 13, 17, 19, 22, 23 and 24; 0.24333349999999998 at 3, 9 and 20.</li><li>(b) All three disclose: axum in-sample precision 0.5 (TP 1, FP 1, n = 8, one positive), which matches `strata.axum`; three-call means against one production request (REQ-88, `html:570`); and pooled out-of-fold precision 0.7950310559006211 below 0.80, with joint attainability not shown and a D1-R re-check required.</li></ul> |
| ADR010-R2-05 | **Met** (verification limit recorded under Scope) | <ul><li>R-063 quotes Q1 ("D1 scope") and Q2 ("Triage costlier side"), every option label and description, and both selections.</li><li>It says the options were "assistant-drafted" and that the owner "selected the named options, not authored their wording". It notes that no additional free text was supplied.</li><li>It quotes the initial direction verbatim ("…you can create labelled datat or use runs we have already actioned"), matching round 2's `:1411` quote.</li><li>It attributes Q2's rationale sentence to the assistant.</li><li>ADR.md:9-10 say the owner "selected the assistant-drafted option" for both answers.</li><li>Q2 opens "For triage", so the miss-cost confirmation is limited to triage.</li><li>Q1's stem describes REQ-98 correctly ("each decision its own default of 0.8/0.8, with 0.9 on that decision's own costlier side"). Option (a)'s description matches the branch REQ-98 ("X3/X5/X10 retain their own per-decision defaults until the user sets them or a preregistration replaces them").</li><li>R1-01's condition ("with REQ-98 described correctly") is therefore shown on the record.</li><li>No document presents option text as the owner's words: checked STATUS:17, :123 and :127, plan:80-82, audit row 22, the card at `html:975-980`, REQ-98 and README:45.</li></ul> |
| ADR010-R2-06 | **Met** | <ul><li>Plan:33 reads `X1P[X1 or dedicated triage preregistration signed] --> D1R[D1-R triage review and optimise]`.</li><li>ADR.md:30 states "a signed X1/dedicated triage preregistration as prerequisite to D1-R (preregistration→D1-R)".</li><li>Plan:82 keeps "does not block drafting Track C preregistrations", and D1→C stays at plan:32.</li><li>No node or text implies that D1-R blocks the X1 preregistration.</li></ul> |

## Round-1 issues: reconfirmation

| Issue | Status | Evidence |
|---|---|---|
| ADR010-R1-01 | **Met** | Triage-only scope appears in ADR.md:9-12, REQ-98 (`html:580`), the card at `:975-976`, STATUS:17 and :127, and plan:80-82. D1→C is at plan:32. REQ-169–REQ-175 (`html:686-692`) are unchanged against base, including "pending the user's minimums". |
| ADR010-R1-02 | **Met** (with round 2's test-comment disposition) | REQ-98 keeps "status `provisional`", "until user costs/minimums are set", the point-minimums-only clause, "confirms/replaces carried rev-2.1 values" and "names the costlier side (X10 recall n/a)". ADR.md:25 quotes the ADR-003 Conditions correctly. ADR.md:27 still carries the comment correction that round 2 accepted. |
| ADR010-R1-03 | **Met** | The deviation is named as limited to the analysis, with no precedent and no design-rule-3 exception: ADR.md:17 and :36, card `:980`, R-063, STATUS:123, README:3, plan:82. |
| ADR010-R1-04 | **Met** | All 52 cases and the E2 `v2_noul` score files are quarantined, and legacy `v1_choice` labels are excluded from the entry floor: ADR.md:12 and :37, card `:984-985`, plan:82, R-063, `defaults.json:21`, README:43. |
| ADR010-R1-05 | **Met** | "part of X1 or its own" and "does not change X1's ordinal decision rule": ADR.md:12, REQ-98, card `:977`, plan:82, STATUS:127. |
| ADR010-R1-06 | **Met** | The CSV, JSON and recipe are committed and pinned (README:7-31). The corrected semantics, interval method, denominators and limits are at README:35-39 and R-063. |
| ADR010-R1-07 | **Met** | The ADR-003 card at `html:806-829` matches base apart from its `rel` span at `:807` (challenger, raw comparison; main agent's confirmation). The ADR-003 `ADR.md` changes only its status line at `:3`. The index is at `docs/adrs/README.md:17`. |
| ADR010-R1-08 | **Met** (R2-01 and R2-02 closed) | ADR.md:3 "accepted". The card is `s-acc` with its footer. Index at `README.md:24` "accepted". STATUS:7 and :127. Pending state appears only outside the signed set (`round-2-response.md`). |
| ADR010-R1-09 | **Met** | Audit row 22 (`jev-adapter-audit.md:63`) records `agreed` (owner, 2026-11-08, R-063), the separate miss-cost confirmation, the missing current-question labels and the D1-R tuning still to come. The lead checked rows 1–38 against base: **only row 22's text changed**. The other 80-line churn in the diff stat is prettier padding the "Fix / note" column, and no Met/Fixed status, Design-level row or evidence cell changed. |
| ADR010-R1-10 | **Met** | `defaults.json:18-21`: owner-set minimums against the untuned threshold, legacy E2 labels against the absent current-version label set, and non-citable exploratory F1. `label_set.summary` stays `"none"`. |
| ADR010-R1-11 | **Met** (superseded by R2-02) | The D1-R conditions are separately labelled at `html:983-987`, and ADR.md:38 calls them conditions for later work. |
| ADR010-R1-12 | **Met** | STATUS:17, :129 and :144 (latest entry R-063); plan:97 "Appendix B indexes ADR-000…ADR-010"; the D1, D1-R and X1P nodes have edges. |
| ADR010-R1-13 | **Met** | ADR Changes (`:24-32`) lists every changed signed file and both deletions. The records section is at `:45-49`. `.omp/test-map.json:4` maps the renamed proof identity (unchanged since round 2). |

## Recomputation (checked clean)

- **`scores.csv`:** 51 data rows, 28 labelled 1 and 23 labelled 0, all `dev`, model `typesafe/jev-1.14-20261014`, question version `e2-v2_noul`.
  - The peer confirmed that the three scores per case are separate calls (`rep` 0/1/2 with distinct generation IDs in `jev_v2_noul_1.14.jsonl`).
  - Three means spot-checked against the CSV: `architecture-routing-utils-cycle` 0.276667, `rust-performance-websocket-demo` 0.346667, `rust-quality-documented-query-overlap` 0.743333.
- **`tuned_heldout`:** TP 128, FP 33, FN 12.
  - P = 128/161 = 0.7950310559006211; R = 128/140 = 0.9142857142857143; F1 = 256/301 = 0.8504983388704318.
  - Totals check: 140 = 28 × 5 and 255 = 51 × 5.
- **Fixed 0.5:** TP 95, FP 10, FN 45, so P = 95/105 = 0.9047619047619048 and R = 95/140 = 0.6785714285714286.
- **Fold thresholds:** 13 / 9 / 3 of 25, as listed above. The spread is 0.1417, above 0.05, which matches the file's warning.
- **`strata.axum`:** P 0.5, TP 1, FP 1, FN 0, n = 8. The CSV has 8 axum rows with one positive.
- **Wilson intervals (z = 1.96), recomputed by the lead and the peer:**
  - Precision 25.6/32.2: centre 0.763584, interval [0.628092, 0.899076].
  - Recall 25.6/28: centre 0.864303, interval [0.754973, 0.973633].
  - Both equal README:37 and R-063.
- **Production makes one request per question.** REQ-88 (`html:570`) says "one request per unit", and the peer found no code in `internal/decisions` that repeats a call.
- **Values unchanged:** `defaults.json` has threshold 0.3, `min_precision` 0.8, `min_recall` 0.9, `metric` fbeta with β = 2, `gating` false, `score_space` raw, `status` provisional, `label_set` none and `minimums_provenance` agreed. `thresholds_test.go:19-37` asserts each of these.

## Issues

None. **Count: 0 issues (0 critical, 0 major, 0 minor).**

## Checks that found no issue

- **New edits are consistent.** The 13/9/3, axum, mean-of-3 and 0.795 figures agree across the README, ADR.md and R-063. A search of STATUS, the plan, `defaults.json`, the card and audit row 22 for `0.385`, "held-out", `0.63` and `0.76` found no stale figure.
- **HTML.** The edited spans at `html:580` and `:971-989` are properly nested. The challenger found no truncated or garbled sentences in the edited text.
- **Miss-cost scope.** No document applies the miss-cost confirmation or the triage pair to X3/X5/X10. Plan:81 keeps "0.9 on each decision's own costlier side".
- **Q1's X3 claim.** Q1 says precision is X3's stricter side. That matches X3's carried precision ≥ 0.9 (`html:688`), and no document adopts it beyond the existing requirements.
- **Diff stat.** All 18 changed files are accounted for: 14 signed and 4 review records.
- **Frozen proof.** The `thresholds_test.go` digest is the same as in round 2, and its assertions match `defaults.json`.
- **Hypothesis ledger.** No product hypothesis was proposed, tested or changed.

## Considered and not raised

- **R3-P1, peer F1: ADR.md:27 and R-063 contradict Q2's stem ("This is why recall is 0.90").** Declined, for four reasons:
  1. **Order of events.** The owner's acceptance ("yes accept, but include a review and optimise task at some stage") came *before* Q2, which was asked after round 1. The agreement cannot have been caused by a confirmation given later.
  2. **The owner's own direction.** The owner said "use the data not costs".
  3. **Attribution.** R-063 already marks Q2's rationale sentence as assistant-drafted. That sentence describes the rationale of the pre-existing default (base REQ-98's "recall [ASSUMPTION]"; base row 22 "recall raised because a miss costs more"), not the owner's reason.
  4. **No denial.** ADR.md:27 says only that the pair was not *set because* of cost, and records the separate confirmation. It does not deny that the confirmation agrees with REQ-98's costlier-side rule.

  The quoted text is new evidence, but it does not change round 2's disposition.
- **R3-P2, peer F2: what the owner was shown before accepting is not recorded.** The main agent's brief (`history://D1Impl`) describes the K1 figures as "Held-out at best-F1 threshold 0.385: precision 0.80 (Wilson 95% 0.63–0.90), recall 0.91 (0.76–0.97)… Fold thresholds 0.24–0.39 (unstable)", with the limits. Declined as an issue, for four reasons:
  1. **Not material to the minimums.** The minimums are targets, and the misdescription concerns the threshold, which was not adopted. 0.795 rounds to 0.80, and the interval shown (0.63–0.90) already conveyed that 0.80 was not established.
  2. **Already decided in round 2.** Round 2 knew K1's "The data points to 0.80 precision" and the 0.385 misdescription. It required disclosure in the record (done, R2-04), not a new presentation to the owner.
  3. **Low stakes.** The minimums gate nothing while ranking is ordinal (REQ-98).
  4. **The owner reviews again.** D1-R requires reviewing the minimums with the owner on properly split data.

  K1's exact text cannot now be read in the compacted `history://Main`. That K1 showed the interval is **[INFERENCE]** from the main agent's brief.
- **Nits declined.** Each would change signed bytes and require another round, for no gain in accuracy:
  - **"original" (peer F3).** ADR.md:17 calls "use the data not costs" the owner's "original" direction, and README:3 cites it as the analysis direction, while R-063 orders the jev-actions direction first. "Original" fairly contrasts the direction with the later acceptance and clarifications, and R-063 holds the exact order.
  - **Intervals (F4).** ADR.md:19 and R-063 do not repeat "threshold-selection uncertainty"; README:37 states it. They say nothing false.
  - **Wording (F5).** "potentially preventing transfer" is weaker phrasing than the README's "may not transfer", but has the same meaning.
  - **Paraphrase (F6).** Q2's "'use the data, not costs'" is the assistant's paraphrase inside verbatim assistant text, so quoting it as written is correct.
  - **Review wording (challenger).** The challenger suggested ADR-009's neutral Review wording instead of "Adopted on…". Optional: within route (a) the wording is accurate once this round is recorded and adoption is committed.
- **Ask-record verifiability (peer F7).** Recorded under Scope as an attestation limit, as with the digests. It is not a defect in the bytes.
- **Audit row 37 (`jev-adapter-audit.md:78`), lead.** Row 37 stays Met, citing row 22, while ADR.md:31 says row 22 no longer claims "a nonexistent visible TODO". Not an issue:
  - The guidance (`building-eval-data.md`, "leave a visible TODO naming the missing cases") is met by `defaults.json` assumptions[3], which names the missing cases (≥ 30 human/`agent_reviewed` labels per label, B-2). Row 22 cites those assumptions.
  - Base had no literal TODO either: no match in `internal/decisions/thresholds` at base. The basis for row 37 is therefore unchanged.
- **Plan:14, lead.** The "Design" row reads "ADR-009 adopted 2026-11-06". It is unchanged from base and still true, and plan:97 indexes ADR-000…ADR-010. This is not the internal contradiction R2-01 fixed.
- **Round-2 nits** (README:28 absolute path, row 22's stale `thresholds_test.go:13-91`, B-2's stale pointer, index column alignment, "agent-reviewed" on the card): already declined in round 2 as optional; not reopened.

## Peers

- **`innovation-challenger`** raised **0 issues**. It found R2-01…R2-06 Met and reconfirmed all thirteen R1 items, including raw comparisons of the ADR-003 card and summary against base, REQ-169–REQ-175 unchanged, and the test assertions. It noted the attestation limit on the ask record and suggested an optional Review wording; the council declined the wording change.
- **`experiment-peer`** raised **2 minor items (F1, F2) and 5 nits (F3–F7)**, all declined with reasons above. It found R2-03, R2-04 and R2-05 Met, recomputed every figure and found it correct, and confirmed against the raw data that the scores are separate calls.
- **Lead:**
  - recomputed the fold-threshold counts and Wilson intervals;
  - compared audit rows 1–38 against base, finding padding only outside row 22;
  - checked the diff stat against the signed set, REQ-98 against base, and the ADR-009 `round-13.md` step 5 citation;
  - considered audit row 37 and plan:14 and raised neither.
- **Where the peers changed the outcome:**
  - The peer's F1 led the lead to recheck round 2's test-comment disposition against the newly visible Q2 text. The order of events kept the disposition unchanged.
  - The peer's F2 and F7 led the council to state the attestation limit and the **[INFERENCE]** about K1 explicitly in this record.
  - The challenger's adopted-form caveat confirmed that ADR.md:43 becomes accurate only through this round plus an adoption commit outside the signed set.

## Verdict

**Zero issues. Signed off.**

The innovation council signs off ADR-010 ("Record owner-agreed Jev triage minimums") at branch `agent/d1-minimums`, head `44f37ff0416fdf1d69a168b08cc216165bd4e755`, base `main` `04d5e632`, with zero open issues. This round is the folder's final round. The sign-off covers exactly these fourteen files at these SHA-256 digests, computed by the main agent:

- `docs/scout-design.html`: `26b55ea82237023769d7317281e48ac32fd9106284176df10d369c941ca9d7a2`
- `docs/adrs/2026-11-08-d1-jev-minimums/ADR.md`: `3b3eaba00d63d0dc777024aac0361b6ae50b50323d8f25cdb36bb506e21fa063`
- `docs/adrs/README.md`: `77f8b0b4b04cf2c392996c6e06d6acc6455b4c076c9c7852ee0b289cb6a27197`
- `docs/STATUS.md`: `d2c80854f486a6060f65e63d8aa5f078978f333b9a81a2848b50d43c5aab8427`
- `docs/jev-adapter-audit.md`: `92aa0487580ba94987431aabfa44106e5a0ed1d4918498722a66c33930f7d118`
- `docs/adrs/2026-10-23-decisions/ADR.md`: `6b92c168364db09db4095885858e9a86db450ec9dc7853a7f9cefb785206e2bb`
- `docs/research-ledger.md`: `ffc274b82b1a45ab7d7c6fdf1a8b2431b0f496ec9bae3ddb48c15a1923351ad1`
- `docs/plans/2026-11-06-remaining-work.md`: `9137529ca827b4147d7ffc43bd06219b2cac82a9782b27c7f26d7a202031013f`
- `internal/decisions/thresholds/defaults.json`: `5dba9d910331bd5c35c1af1143e35617d4c5a5b6dc99323d3f0ed7e6180c8651`
- `internal/decisions/thresholds/thresholds_test.go`: `fba178e3bd6469e3d932af98c65ae239c85bd5e1cd539dfb39e2473712336b91`
- `.omp/test-map.json`: `96de43c74f2769c69c83e84e68a6701d3c83a991a818d47bedb73811f1f703db`
- `evaluation/d1-triage-minimums-exploratory/README.md`: `3759df7e2e01ff334ba0711f4259a9df43f8ed4ba943548e4da5deb1f91eb1c9`
- `evaluation/d1-triage-minimums-exploratory/scores.csv`: `c7679a2da9a30b3b905e4f8be3c81f4eed6ad459e85426db719a356f3818c2d2`
- `evaluation/d1-triage-minimums-exploratory/tuning.json`: `03402fc8a1d183338ccf58ce7c723584a44f5fd615944b12509f519e0f587f9b`

**Conditions of this sign-off:**

1. **The bytes are frozen.** Any change to these files, including the declined nits, voids this sign-off and needs a new zero-issue round.
2. **Adoption happens outside the signed set** (route (a); ADR-009 `round-13.md` step 5). The signed files are not edited to mark adoption.
3. **The ask record rests on the main agent's attestation**, as do the digests.
4. **D1-R's entry floor, preregistration and examined-data quarantine bind later work.** This sign-off authorises no further Jev-versus-label comparison or threshold tuning outside the experiment rules.

---

## Council memo

1. **Recommendation.** Adopt ADR-010 on this round.
   - Save this record as `docs/adrs/2026-11-08-d1-jev-minimums/round-3.md`.
   - Verify the fourteen digests against the files.
   - Commit with adoption recorded only outside the signed set.

   Make no further edits to signed bytes.
2. **Alternatives.**
   - **(A) Sign off and adopt now (recommended).** Every R1 and R2 issue is closed; the remaining points are declined nits or attestation limits.
   - **(B) Apply the nits (F3–F6, Review wording) first.** This gives marginally tidier prose, but it changes four or five signed files and forces round 4 for no change in accuracy or decision.
   - **(C) Treat F2 as blocking and show the owner the corrected figures before adoption.** This gives the strongest informed-consent record, but it reopens round 2's disposition over a threshold misdescription that does not bear on the minimums. D1-R already requires reviewing the minimums with the owner, and the minimums gate nothing (REQ-98).
   - **(D) Reword ADR.md:27 per F1.** Not supported: the order of events and "use the data not costs" make the current text accurate.
3. **Evidence.**
   - **ADR-010:** `ADR.md:3,9-12,17-20,24-31,34-43`.
   - **Design:** `scout-design.html:570,580,582,686-692,806-829,948-989`, with base `:580`.
   - **Ledger:** `research-ledger.md:71`.
   - **STATUS:** `:1,5,7,17,123,127,129,144`.
   - **Plan:** `:14,32-33,80-82,97`, base and branch.
   - **Audit table:** `jev-adapter-audit.md:40-82`, base and branch (only row 22 differs beyond padding).
   - **Thresholds:** `defaults.json:5-21`; `thresholds_test.go:19-37`.
   - **Exploratory record:** `tuning.json` (`fold_thresholds`, `tuned_heldout`, `default_0.5_heldout`, `strata.axum`); `scores.csv`; exploratory `README.md:7-45`.
   - **Precedent:** ADR-009 `ADR.md:38-39` and `round-13.md:212`.
   - **Guidance:** `building-eval-data.md:85-102`.
   - **Transcripts:** `history://D1Impl` (brief, diff stat, digests); `history://Main` (compacted ask prefix).
4. **Risks.**
   - **Unverified transcript.** The council relies on the main agent for the ask record and the digests. If the attested ask text proves inaccurate, the scope question must go back to the owner.
   - **The owner decided on pre-correction figures** ("at 0.385", precision "0.80"). This is low impact while nothing is gated. D1-R must present corrected, properly split figures when it reviews the minimums.
   - **Joint attainability of 0.80/0.90 is unproven**, and the per-stratum weakness (axum, nest, webpack) may make the pair infeasible. D1-R handles this.
   - **Reversibility.** Everything is text or JSON under version control and can be reverted. A later ADR can amend it in the usual way.
5. **Execution outline.**
   1. Save this record verbatim as `docs/adrs/2026-11-08-d1-jev-minimums/round-3.md`.
   2. Run `shasum -a 256` on the fourteen signed files, and expect exactly the digests in the Verdict.
   3. Run `git diff --stat 04d5e632..HEAD` after staging. Expect the 18 files reported plus `round-3.md` as the only addition: 19 files, with no signed file changed.
   4. Commit and merge. Record ADR-010's adoption only outside the signed set (commit message or PR, and later STATUS updates under a future change), without editing the fourteen signed files.
   5. Leave D1-R for an X1 or dedicated council-signed preregistration that meets the label floor and quarantine above.
