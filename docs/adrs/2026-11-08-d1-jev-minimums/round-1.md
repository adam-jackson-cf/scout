# Round 1: ADR-010, record owner-agreed Jev triage minimums

Date: 2026-11-08. This round follows design rules 1–3 and the ledger rules in `CLAUDE.md`, and the ADR rules in `docs/adrs/README.md`. The innovation council lead ran it with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review of the owner-decision scope, the design-rule-3 framing, word-by-word changes to REQ-98, consistency across documents, ADR format, and whether accepted ADRs were edited. It also read the main session transcript (`history://Main`) to find what the owner actually decided.
- **`experiment-peer` (Opus):** evidence and inference review of R-063. It covered what the exploratory analysis supports, competing readings of the owner direction, data already examined and its reuse, and statements made false by the diff. It also read `tune_threshold.py`, the E2 question file and the production triage question.

The review was read-only. It made no model calls outside the council and did not touch `.codegraph`.

**The main agent computed the digests; the council did not.** The lead and both peers had read-only tools and no shell.

| File | SHA-256 (main agent) |
|---|---|
| `docs/scout-design.html` | `15a7614d4e5d6f3d98c585b8ffca5b24218efdc6b0d557375180330ecddca4d0` |
| `docs/adrs/2026-11-08-d1-jev-minimums/ADR.md` | `39f6e38e385d60aa6ae5bc8e3d8b2bb92c97cc4745155c5a277ae021a2f7d9ab` |
| `docs/adrs/README.md` | `31fcd9454995479271c05462dc1d82e6c990f65c7afae5715f8da055bda6055a` |
| `docs/STATUS.md` | `dcaa59944b83bbb946dd099425328b59c5041f501de61b2c6ed31e42debb61e6` |

## Scope

- **Under review:** worktree `/Users/adamjackson/Projects/work/scout-d1`, branch `agent/d1-minimums`, head `95b892e8cee28f6c002cff0f5aa6be7bacbc0fdb`. Base is `main` at `04d5e632`; the main worktree `/Users/adamjackson/Projects/work/scout` is at that commit and was used for every base comparison.
- **Commits in range:**
  - `b8d81025`: `defaults.json`, `thresholds_test.go`, research ledger R-063, plan D1 and D1-R.
  - `95b892e8`: design html, ADR-010, ADR index, STATUS.
- **Files read:**
  - ADR-010 `ADR.md`.
  - `scout-design.html`: REQ-98 at `:580`; the ADR-003 card at `:807-826`; the ADR-010 card at `:971-985`; REQ-169 and REQ-171–REQ-175 at `:686-692`; Appendix B rules at `:729-733`.
  - `docs/adrs/README.md`.
  - `docs/STATUS.md` at `:7,17,123,127,129,144`.
  - `docs/research-ledger.md` at `:63,67,71`.
  - The plan, `docs/plans/2026-11-06-remaining-work.md`, at `:20-35,66-80,94`.
  - `internal/decisions/thresholds/{defaults.json,thresholds.go,thresholds_test.go}`.
  - `internal/decisions/triage.go:13-35`.
  - `evaluation/jev-triage-v1/questions.py:9-12`.
  - `docs/jev-adapter-audit.md:63,84`.
  - The precedent records ADR-003, ADR-008 and ADR-009 `ADR.md`, and `2026-11-06-design-restructure/round-13.md`.
  - `~/.omp/agent/skills/jev-actions/scripts/tune_threshold.py:89-172,328-353`.
  - `history://Main:1377-1513`, the owner exchange.
- **Not reopened:** decisions settled in ADR-000…ADR-009.

## Issues

| ID | Severity | Location | Problem | Required fix |
|---|---|---|---|---|
| ADR010-R1-01 | major | ADR.md:11 (Decision bullet 3); research-ledger.md:71 (R-063 "used for" column); plan:79 (D1 "Done", Blocks "—") and plan:32 (D1→C edge removed). Compare scout-design.html:580 (REQ-98), :820 (ADR-003 card: "Other per-decision tuning minimums remain 0.8/0.8, raised to 0.9 on the costlier side"), STATUS.md:127 ("Other per-decision minimums retain their defaults"), ADR.md:23 ("Other per-decision defaults … remain unchanged") and REQ-169/171/172/173/175 ("pending the user's minimums"). | **The documents give two incompatible scopes for the owner decision.**<br>• The ADR, R-063 and the plan say the triage pair P 0.80 / R 0.90 is "the starting point for X3/X5/X10".<br>• REQ-98, the ADR-003 card, STATUS and ADR Changes say the other decisions keep their defaults.<br>**What the owner said:** they accepted K3, "The X3, X5 and X10 preregistrations then start from 0.80/0.90", with "yes accept, but include a review and optimise task at some stage" (`history://Main:1509-1513`).<br>**But the acceptance rested on a misdescription.** The earlier explanation (`:1401`, `:1407`) said the 0.8/0.9 pair applies to the X3/X5/X10 point minimums. Base REQ-98 gives each decision 0.8/0.8, with 0.9 on *that decision's* costlier side (`completeness-audit.md` D:184.15: "uses the user's minimums if set and otherwise these defaults, and names its costlier side").<br>For X3, an assignment-precision decision (rev 2.1 precision ≥ 0.9, REQ-171), starting at recall 0.90 reverses the default. A pair estimated on triage scores also says nothing about file-to-feature assignment, `is_fix` or pair detection.<br>**This is an essential decision that the repository cannot settle. Escalated to the owner.** | Ask the owner, with REQ-98 described correctly, to choose:<br>(a) **triage only.** X3/X5/X10 keep the REQ-98 per-decision defaults (0.8/0.8, with 0.9 on each decision's own costlier side) until the owner sets them or a preregistration replaces them; or<br>(b) **0.80/0.90 for all three.** The X3/X5/X10 preregistrations start from P 0.80 / R 0.90 (X10 recall n/a), and each names its costlier side and confirms or replaces them.<br>Quote the answer verbatim in R-063. Then make one rule appear in REQ-98, the ADR Decision and Changes, the ADR-010 card, STATUS:127, R-063 and the plan:<br>• the D1 row, its Blocks column and the D1→C edge stay as "Done / —" only under (b); under (a), split D1 and keep the remainder open;<br>• the "pending the user's minimums" clauses in REQ-169/171/172/173/175 must match the answer. |
| ADR010-R1-02 | major | scout-design.html:580 (REQ-98); :820 (ADR-003 card, bullet 2); ADR.md:9 and :22; thresholds_test.go:28 | **REQ-98 changes beyond triage provenance, without saying so.**<br>Base: "…(`triage.real_defect`: recall [ASSUMPTION]), status `provisional`, until user costs/minimums are set … each preregistration confirms/replaces carried rev-2.1 values".<br>Branch: "…until the user sets them … confirms/replaces its minimums".<br>Three unscoped changes:<br>(i) "status `provisional`" is dropped from every other decision's defaults, although REQ-183 relies on that marking;<br>(ii) "costs" is dropped as a pending user input for every decision;<br>(iii) "carried rev-2.1 values" is replaced.<br>ADR.md:23 says the other defaults "remain unchanged". ADR.md:22 misquotes ADR-003 as "until the user sets minimums"; ADR-003 `ADR.md` Conditions says "until the user sets costs and minimums".<br>"With misses the costlier side" is presented as part of the owner's agreement (ADR.md:9, REQ-98), and the test comment says "The owner agreed … because a miss costs more". But the owner said "use the data not costs" (`history://Main:1454`), and K3 put no costlier side to them. The miss-costlier rule is the retained `cost_rule` assumption. | • Restore "status `provisional`" for the per-decision defaults and "carried rev-2.1 values", inserting the triage clause without removing them.<br>• Costs: either restore "costs/minimums", or state in REQ-98 and ADR Changes, citing R-063, that the owner set minimums from data rather than costs, scoped as R1-01 decides.<br>• Mark "misses the costlier side" as a retained assumption unless the owner confirms it.<br>• Correct the ADR.md:22 quote and the `thresholds_test.go:28` comment.<br>• List every REQ-98 wording change in ADR Changes. |
| ADR010-R1-03 | major | ADR.md:17-18; ADR-010 card :980; R-063 caveat; STATUS.md:123 | **The rules the comparison set aside are not named.**<br>The analysis was a Jev-versus-label comparison on real data outside any signed preregistration. REQ-102 ("No Jev-versus-label comparison … occurs outside a council-signed preregistration") and CLAUDE.md General "Jev labelling" ("no Jev-versus-label comparison runs outside a signed-off preregistration") prohibit that. The ADR discusses only design rule 3 and says the owner "waived preregistration".<br>**What the transcript shows:** the owner directed the analysis ("use the data", "use the data not costs", `history://Main:1446,1454`) after the assistant's question about the rule was interrupted (`:1450`). The owner accepted the result after being told it ran without a preregistration and cannot be cited (`:1507`, `:1513`). "Waived" is the author's word.<br>**Precedent:** R-055 records "an explicit user exception to the rule that Jev-versus-label comparisons run only under a signed-off preregistration". | In the ADR (Informed by and Conditions), R-063, STATUS:123 and the card:<br>• name REQ-102 and CLAUDE.md General "Jev labelling" as the rules set aside;<br>• say exactly what happened: the owner directed the analysis and accepted it after being told it was unpreregistered and not citable;<br>• limit this to the R-063 analysis, say it is no precedent, and say it grants no exception to design rule 3. |
| ADR010-R1-04 | major | ADR.md:12 and Conditions :26-29; plan:80 (D1-R); card :983 | **Nothing limits reuse of data that has now been examined.**<br>The 51 labels in `evaluation/jev-triage-labels-v1/labels.jsonl` and the saved scores `evaluation/jev-triage-v1/raw/{jev,unlabelled}_v2_noul_1.14.jsonl` have been compared and tuned on. Nothing bars them from a held-out test set.<br>With 28/23 already labelled, D1-R's entry condition ("at least 30 … per label") could be met with 2 + 7 new labels, and a random repository split would then put examined cases into test. REQ-169 makes X1's preregistration fix "test-set reuse".<br>These labels were also made under E2's `v1_choice` guide, not a frozen guide for `triage.real_defect@v1` (REQ-100). | Add a Condition to the ADR, the card, D1-R and R-063:<br>• the 52 `jev-triage-labels-v1` cases and the E2 `v2_noul` score files are **examined data**;<br>• any X1 or D1-R preregistration must disclose them, may use them only as development data, and must draw confirmatory test data from unexamined cases;<br>• D1-R states whether `v1_choice`-guide labels count toward its entry threshold. |
| ADR010-R1-05 | major | ADR.md:12; plan:80 | **D1-R is assigned to unrelated experiments.** Both say its preregistration may be "part of X3/X10 or its own". X3 (REQ-171) assigns unowned files to features. X10 (REQ-175) detects competing implementations and measures no recall. Neither has triage findings, labels or estimand. Triage retrieval is X1 (REQ-169). As written, an X3 or X10 sign-off could be read as authorising triage tuning. | Change it to "part of X1 or its own". State that a D1-R threshold does not change X1's ordinal decision rule. |
| ADR010-R1-06 | major | research-ledger.md:71 (R-063); ADR.md:17; defaults.json:18 | **R-063 misdescribes its figures and cannot be reproduced.**<br>(a) "Held-out at best-F1 threshold 0.385: precision 0.80 …, recall 0.91". In `tune_threshold.py` `tune()` (`:146-172`), the "tuned held-out" figures pool out-of-fold counts over 25 fold-tests, each at that fold's own training threshold (0.24–0.39). The shipped 0.385 is `choose(rows)` on all 51 cases, so no held-out figure was measured at 0.385.<br>(b) The script prints no intervals (`fmt`, `:328-331`). The Wilson method and n are not recorded; the pooled counts contain each case five times.<br>(c) The outputs stayed in `$TMPDIR/d1-data/` (`history://Main:1492`) and were never committed. The ledger rule requires a link; R-055 committed its outputs (`evaluation/x2d-postrun-tuning/`).<br>(d) The disclosed mismatches are incomplete:<br>  • `v2_noul` (`questions.py:9-12`) is not `triage.real_defect@v1` (`triage.go:13-35`: data-not-instructions clause, `state.facts`, an 8-line/600-byte excerpt bound);<br>  • the analysis optimised F1 with minimums 0.01, while the entry uses F-beta β = 2.<br>(e) The owner's answers are not quoted, unlike R-059 and R-061. | • Reword the figures as "pooled out-of-fold (5-fold × 5 repeats, per-fold thresholds 0.24–0.39); in-sample best-F1 threshold 0.385". State the interval method and n, or remove the intervals.<br>• Either commit the exact `scores.csv`, command and output (no new model calls) and link them, or remove the figures from R-063 and the ADR.<br>• Add the question/state mismatch and the F1-versus-F2 note to the R-063 caveats, the ADR Informed by and `defaults.json` assumptions[0].<br>• Quote K3 and the owner's answers verbatim. |
| ADR010-R1-07 | major | scout-design.html:820 (ADR-003 card, Open conditions); ADR.md:23 | **An accepted ADR's card is edited.**<br>Appendix B says "Append-only" (`:729`) and "Accepted ADRs are never edited, except their status line" (`:732`). The README says the cards summarise the ADR files.<br>The base card bullets match ADR-003 `ADR.md` Conditions: "Tuning minimums remain 0.8/0.8 … until the user sets costs and minimums" and "The 0.9 triage-recall judgement is an assumption…". The branch replaces them with 2026-11-08 content, including "ADR-010 proposes…", under ADR-003's accepted header. | • Restore the ADR-003 card body to base, byte for byte.<br>• Put the amendment only in REQ-98 and the ADR-010 card.<br>• In the adopting commit, change only ADR-003's card `rel` span, its `ADR.md` status line and its README status, to name ADR-010 (amends).<br>• Correct ADR.md:23 to match. |
| ADR010-R1-08 | major | ADR.md:3 and :33-36; html:580 ("ADR-010 (proposed)"), :972, :983-984; README.md:24; STATUS.md:127 | **The files meant for signing are written in proposed form.**<br>Signed files are frozen. Adoption would require editing "proposed", "council review pending", "Pending; no council round or sign-off exists yet" and "Review pending:", which would void the digests. The status-line exception covers an *earlier* accepted ADR when a later one is adopted. It does not cover this ADR's own body, the REQ table, the card body or footer, the README or STATUS.<br>**Precedents:** ADR-009 signed its accepted-form bytes (`round-13.md` step 5: "Mark ADR-009 adopted only through files that are not in the signed set"). ADR-008 used a content sign-off followed by separate adoption rounds. | State one route:<br>**(a) recommended:** submit the signed set in adopted form:<br>  • ADR-010 status accepted, with the Review section in ADR-009's form;<br>  • REQ-98 citing "ADR-010";<br>  • the card as `s-acc`/accepted, footer "Review: `docs/adrs/2026-11-08-d1-jev-minimums/`";<br>  • README "accepted"; STATUS "ADR-000…ADR-010";<br>  • the ADR-003 status-line, `rel` and README updates from R1-07.<br>**(b)** an ADR-008-style content sign-off, then an adoption round on the adopted bytes. |
| ADR010-R1-09 | major | docs/jev-adapter-audit.md:63 (row 22) | **A governed document contradicts the new default.** Row 22 still reads "minimums precision 0.8 / recall 0.9 (`assumed`: skill default 0.8/0.8, recall raised because a miss costs more…)", while `defaults.json` now says `"agreed"`. CLAUDE.md design rule 1 governs the Met/Fixed rows. ADR Changes and the signed set both leave it out. | • Update row 22: `agreed` (owner, 2026-11-08, R-063), with miss-costlier as a retained assumption (R1-02).<br>• Make its "visible TODO" wording match R1-10.<br>• List it in ADR Changes and add the file to the signed set. |
| ADR010-R1-10 | minor | internal/decisions/thresholds/defaults.json:18-20 | **The entry contradicts itself.** assumptions[2], kept from base, says "No labelled cases exist … the status stays provisional until labels set per-decision minimums". 52 human-labelled triage cases exist (STATUS:123, REQ-101), R-063 used 51 of them, and the minimums are now owner-set. assumptions[0] says "from an exploratory data run" without "not citable" or "threshold not tuned on it". | • Reword assumptions[2]: human labels exist for E2's `v1_choice` question; no label set exists for this question version; the threshold is untuned; the minimums were owner-set (R-063); status stays provisional until a preregistered tuning run (D1-R) reports ready.<br>• Add "not citable experiment evidence" to assumptions[0].<br>• Keep `label_set.summary: "none"`. |
| ADR010-R1-11 | minor | scout-design.html:983 (ADR-010 card, "Before adoption") | **Later-work conditions appear as conditions for adoption.** The card lists D1-R's entry conditions (≥ 30 labels per label, a preregistration) under "Before adoption", which makes them look like prerequisites for this provenance-only amendment. The plan says D1-R "does not block starting track C preregistrations". | Keep only the council sign-off under "Before adoption". Move the D1-R entry conditions, with the R1-04 condition, to the rationale or to a separately labelled D1-R section. |
| ADR010-R1-12 | minor | STATUS.md:17 and :144; plan:94 and :32 | **Stale pointers.**<br>• STATUS:17: "Track C preregistrations and D1 remain open", against plan:79 "Done".<br>• STATUS:144: "Latest research entry **R-062**", against STATUS:129 "R-001…R-063".<br>• plan:94: "Appendix B indexes ADR-000…ADR-009", but the ADR-010 card exists.<br>• The mermaid D1 node has no edges, and D1-R is missing from the graph. | Make these agree with the R1-01 outcome: D1 status, D1-R outstanding, R-063, the ADR-010 index range, and a D1-R node with its relation to X1. |
| ADR010-R1-13 | minor | ADR.md:12, :22-24; ADR.md overall | **ADR Changes is incomplete and partly inaccurate.**<br>• It calls D1-R "already-recorded" and says the "embedded-default provenance and frozen proof were committed in `b8d81025`". Both are inside the reviewed range, and "frozen proof" is undefined.<br>• Changes does not list `defaults.json`, `thresholds_test.go`, R-063, the plan rows or audit row 22.<br>• There is no "Records in this folder" section (ADR-003, ADR-008 and ADR-009 all have one).<br>• The allowed label sources differ: ADR "human/`agent_reviewed`"; `defaults.json` "human, outcome or agent_reviewed"; audit B-2 "(`agent_reviewed`)". | • List every changed file and what changed in it.<br>• Name `thresholds_test.go`'s assertions instead of "frozen proof".<br>• Add "Records in this folder" (ADR.md and the round files).<br>• Use one label-source list everywhere, or say why they differ. |

**Count: 13 issues (9 major, 4 minor).**

## Checks that found no issue

- **Owner reservation.** REQ-98 reserves minimums to the user (base: "until user costs/minimums are set"). The owner's acceptance of P 0.80 / R 0.90 for `triage.real_defect` is real (`history://Main:1509-1513`), and the provenance change from `assumed` to `agreed` was disclosed before acceptance (`:1509`).
- **Design rule 3 framing.** The values are unchanged and the minimums are the user's to set, so recording the owner's choice needs no design-rule-3 exception. No changed document presents R-063 as signed evidence: html:580 "not citable experiment evidence"; ADR.md:18; STATUS:123; plan:79; the R-063 caveat. R1-03 concerns REQ-102 and the Jev-labelling rule, not rule 3.
- **Values unchanged, base against branch `defaults.json`:** threshold 0.3, `min_precision` 0.8, `min_recall` 0.9, `fbeta` β = 2, `gating: false`, `score_space: raw`, `status: provisional`, `label_set: none`. `thresholds_test.go:13-38` asserts all of them.
- **REQ-98 clauses kept:** defaults substitute only for precision/recall point minimums (X3/X10/X5); not X1 gain or X4/X5 lower bounds; each preregistration confirms or replaces and names its costlier side; X10 recall n/a. REQ-100, REQ-101, REQ-102 and REQ-169–REQ-175 are byte-identical to base.
- **ID and folder:** ADR-010 is the next sequential ID, and `2026-11-08-d1-jev-minimums/` matches the date. R-063 follows R-062.
- **Accepted ADR files:** ADR-003 `ADR.md` is unchanged from base. Holding its status-line update until adoption is correct. The card is the defect (R1-07).
- **Frozen experiments:** `evaluation/x2d-merge-association-pilot/signoff.json` pins `defaults.json` (`5fa8aebe…`). X2d's rounds finished before ADR-008 was adopted, so no running signed-off experiment is affected **[INFERENCE]**.
- **Hypothesis ledger:** no product hypothesis was preregistered or tested, so the ledger rule is not triggered. HD2 needs no state change.

## Considered and not raised

- **"Critical: R-063 needs an explicit exception to design rule 3" (peer C-1).** Not raised as framed. The owner may set minimums under REQ-98, and recording that choice changes no value. The real gap is the unnamed REQ-102 / Jev-labelling deviation, which is R1-03.
- **"≥ 30 per label is underpowered for per-stratum checks" (peer m-1).** It mirrors B-2 and `defaults.json` as a floor. REQ-102 has the preregistration fix volumes. Advisory for D1-R.
- **Rule-family concentration and an unrepresentative positive rate (peer M-6).** It is plausible that many positives are mutable-action findings and that 28/51 positives is a convenience-sample rate, but the council did not verify this fully **[INFERENCE]**. Advisory: D1-R should stratify by rule family and weight precision to the production population (REQ-138, X1 inclusion probabilities).
- **A REQ-102 carve-out saying that recording user-set minimums is not a "default change".** It would clarify the rules but is not needed for this change. Optional.
- **`docs/adrs/README.md:24` column padding.** Cosmetic. No markdown gate exists in `scripts/gate.sh`.

## Peers

- **`innovation-challenger`** raised 8 items: REQ-98 weakening (R1-02); the ADR-003 card, with the Appendix B "Append-only" wording (R1-07); examined-data reuse (R1-04); X3/X10 (R1-05); the test-comment rationale (R1-02); `defaults.json` (R1-10); STATUS pointers (R1-12); the card's "Before adoption" (R1-11). It found the transcript evidence for the owner's K3 acceptance.
- **`experiment-peer`** raised 1 critical, 10 major, 9 minor and 3 nit. Merged into R1-01–R1-13 above; the declined items are listed under "Considered and not raised". Its unique findings:
  - audit row 22 (R1-09);
  - `tune_threshold.py` pooled out-of-fold semantics and the question/state mismatch (R1-06), both re-verified by the lead (`tune_threshold.py:146-172`, `triage.go:13-35` against `questions.py:9-12`);
  - plan:94 (R1-12);
  - the "frozen proof" wording (R1-13).
- **Lead:** the signed-form problem (R1-08) and the misdescription behind the owner's acceptance (R1-01), from `history://Main:1401,1407` against base REQ-98.
- **Where the peers changed the outcome:**
  - The challenger's transcript evidence changed R1-01 from "scope creep beyond the owner decision" to "owner-accepted, but on a misdescription; escalate". It also showed the owner declined costs ("use the data not costs"), so R1-02 now allows restoring or declaring costs rather than requiring restoration.
  - The challenger's Appendix B "Append-only" wording raised R1-07 from minor to major.
  - The peer added R1-09 and the technical substance of R1-06.
  - The peer's C-1 was downgraded and reframed as R1-03.

## Verdict

**13 issues open.** Not signed off. ADR-010 must not be adopted.

**Escalation (R1-01):** the owner must choose (a) or (b) before the next round. The council cannot derive the scope of the owner decision.

### Required with the next submission

1. The owner's verbatim answer to R1-01, recorded in R-063.
2. Fixes for R1-01–R1-13, and a per-issue response (`round-1-response.md`).
3. The route chosen under R1-08. If it is (a), the full signed set in adopted form: `docs/scout-design.html`, ADR-010 `ADR.md`, `docs/adrs/README.md`, `docs/STATUS.md`, `docs/jev-adapter-audit.md`, ADR-003 `ADR.md` (status line only), `docs/research-ledger.md`, the plan, `internal/decisions/thresholds/defaults.json` and `thresholds_test.go`. Give the main agent's SHA-256 for each, and the `git diff --stat` against `04d5e632`.
4. Confirmation that the ADR-003 card body is byte-identical to base apart from its `rel` span.

---

## Council memo

1. **Recommendation.** Do not adopt ADR-010 as drafted. Put the R1-01 scope question to the owner, with REQ-98 described correctly. Then cut ADR-010 down to a provenance-only amendment:
   - record the owner's triage minimums, and the X3/X5/X10 starting point only if the owner confirms (b);
   - restore the REQ-98 clauses and the ADR-003 card;
   - name the REQ-102 / Jev-labelling deviation and quarantine the examined data;
   - re-home D1-R under X1;
   - make R-063 accurate and reproducible;
   - fix audit row 22;
   - resubmit in adopted form.
2. **Alternatives.**
   - **(A) Triage-only amendment.** Smallest blast radius and consistent with base REQ-98. The owner's literal K3 wording would then need re-confirmation.
   - **(B) Adopt the owner's literal X3/X5/X10 starting point.** Faithful to the words accepted, but built on a misdescription. It reverses X3's default costlier side and requires edits to REQ-169–REQ-175.
   - **(C) Record an explicit design-rule-3 exception in the R-059 style** and keep "from data" as the basis. More process for no benefit, since REQ-98 already reserves minimums to the user.
   - **(D) Defer ADR-010 until D1-R reports.** Avoids every framing risk, but leaves `minimums_provenance` misdescribing the owner's actual decision.
3. **Evidence.**
   - Base against branch REQ-98 (`scout-design.html:580`).
   - ADR-003 card base `:820` against branch; Appendix B `:729,732`.
   - ADR-003 `ADR.md` Conditions.
   - `history://Main:1401,1407,1446,1450,1454,1492,1507,1509,1513`.
   - `tune_threshold.py:146-172,328-331`; `triage.go:13-35`; `questions.py:9-12`.
   - `jev-adapter-audit.md:63`; `defaults.json:10,18-20`.
   - R-055 (`research-ledger.md:63`); ADR-009 `round-13.md` (signed accepted-form bytes, step 5); ADR-008 `ADR.md` Review (content and adoption rounds).
   - `completeness-audit.md` D:184.15.
4. **Risks.**
   - The owner may confirm (b). The design then carries a triage-derived starting point into unrelated decisions; each preregistration's confirm/replace limits the harm.
   - If R1-04 is not fixed, examined cases could enter an X1 or D1-R test split and contaminate a later signed result.
   - The council relies on the main agent's digests.
   - Everything is text under version control and reversible by revert. Nothing gates on these minimums while ranking is ordinal (REQ-98).
5. **Execution outline.**
   1. Put the R1-01 (a)/(b) question to the owner and quote the answer in R-063.
   2. Restore the ADR-003 card and the REQ-98 base clauses; add the scoped triage clause (R1-02, R1-07).
   3. Add the REQ-102 / Jev-labelling deviation and the examined-data condition to the ADR, the card, R-063, STATUS and D1-R (R1-03, R1-04).
   4. Re-home D1-R to "part of X1 or its own" (R1-05).
   5. Correct R-063's figures, caveats and links, committing the outputs or removing the figures (R1-06).
   6. Update audit row 22 and `defaults.json` assumptions[0] and [2] (R1-09, R1-10).
   7. Split the card's "Before adoption"; fix STATUS:17,144, plan:32,94 and ADR Changes (R1-11–R1-13).
   8. Write the signed set in adopted form, including the ADR-003 status-line, `rel` and README updates (R1-08).
   9. The main agent recomputes the digests and diff stat, writes `round-1-response.md`, and requests council round 2.
