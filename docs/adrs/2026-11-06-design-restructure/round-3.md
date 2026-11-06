# Design review round 3: ADR-009, one design document with requirements and ADR appendices

Date: 2026-11-06. Rules: `AGENTS.md`/`CLAUDE.md` design rules 1–3. The innovation council (lead) ran this review with two independent peers: `innovation-challenger` (GPT, adversarial) and `experiment-peer` (Opus, evidence and provenance). The review was read-only. It made no model calls outside the council and did not touch `.codegraph`.

**Digest recomputation was not possible.** You asked for digests to be recomputed first. Neither the lead nor either peer had a shell, so no file was hashed and `git diff --cached --stat` was not run. A mismatch cannot be detected or ruled out. The table below lists the digests the main agent submitted, unverified.

| File | SHA-256 (submitted; not recomputed) |
|---|---|
| `docs/scout-design.html` | `dc2e3da50c151e8c7872d9b42014c8e5c45c336cb58f616e75ad299ddd290097` |
| `docs/adrs/2026-11-06-design-restructure/proposal.md` | `61db6c643b1c95d9261879380fc0712c382d48938284ee1632b3c39ebb2f1069` |
| `…/requirements-map.md` | `3dc2b80523062a4d80ee69e9d1dd18e807ae17cffb48ae33d0f225596b46af16` |
| `…/design-section-map.json` | `3cc0225e4bb98d6e39fbb6e71d870937adb84093cd9d301a85c026fb204da974` |
| `…/ADR.md` (ADR-009) | `af4332fa2a17c7aca98d0b5fc0171cd46755811ae52b75f86f5e17f645dd591e` |
| `AGENTS.md` | `aff517d34b71ac2ed189dcc9f4d3b03380e9b91e9b754a2bd98aceeaa4718511` |
| `docs/adrs/README.md` | `13d0c004c74d7525a14fa650c5341b321481b92e706a305ddd53503b9dcc887b` |
| ADR-000 `2026-10-19-v1-decisions/ADR.md` | `4d631ce92da11fc7c42467d2b18cc12902f4e193d92170a7e5e33a9a8547b36f` |
| ADR-001 `2026-10-21-e1b-v2/ADR.md` | `e5db1e7659fb562bd3c89cb36483d789d0e669d3208c8f2a9e0617346d2be8fa` |
| ADR-002 `2026-10-22-rescope/ADR.md` | `6cf77ff92d6c3dacaa2cf082342eba5a400f74297ac274b5cbd3f592350443fa` |
| ADR-003 `2026-10-23-decisions/ADR.md` | `e8d5afa224b927fd992ca79df6b6bdcb8b9e2bf7d284d44e3ede6c701f6f026e` |
| ADR-004 `2026-10-23-clawpatch-patch-v3/ADR.md` | `e6e91bdf68aac6a17662697c99b9d3bd5a224bbdb653bb2ceed59ff28780b013` |
| ADR-005 `2026-10-23-codex-luna-mapping/ADR.md` | `3fca3e58a6f21709a1995c88da988443d148df99a7948988e1a4327a2dd93c82` |
| ADR-006 `2026-10-24-area-refresh-fallback/ADR.md` | `371a870bd4302d5bf3a6f224679b6ac2ad321a0c3b1dd4aeb6e70ce4941e9a82` |
| ADR-007 `2026-10-24-cutover-cleanup/ADR.md` | `de9fffdf63fc0ec61a98bdde7327617682d7b336b1199711537d48b26ee04721` |
| ADR-008 `2026-10-31-jev-merge-association/ADR.md` | `76dbd3aec2503d447270a91bdfb9a9c44e35ca730c45199add44a34847c45ab0` |

## Scope

- **Under review:** the worktree `/Users/adamjackson/Projects/work/scout-req`, branch `agent/requirements-path`, which contains:
  - the "Round 2 fixes" in `proposal.md`;
  - everything those fixes touched;
  - a fresh broad pass over questions 1–5.
- **Base:** `main` at `9a3d19f5` in `/Users/adamjackson/Projects/work/scout/`, including the deleted originals and the frozen `docs/design-reviews/` rounds.
- **Checks made:**
  - each R2 required fix, at every location it named;
  - supersession and amendment in both directions for all ten ADRs (cards against `ADR.md` status lines);
  - every digest in each `ADR.md` against its frozen round;
  - every `@9a3d19f5` pin against the base content at that commit;
  - experiment rules 1–7, word for word against base `CLAUDE.md`;
  - ledger append-only status;
  - live references to deleted files outside frozen records;
  - the claimed counts;
  - 19 or more sampled `requirements-map.md` rows, traced from base source text to the target REQ.
- **Not reopened:** decisions settled in ADR-000…ADR-008.

## Verification of round-2 fixes

| # | Evidence | Result |
|---|---|---|
| R2-01 | Rule 2 exception at `AGENTS.md:44-46`, with matching wording in worktree `CLAUDE.md`. README `:3` agrees. ADR-009 `ADR.md:38-39`, `docs/STATUS.md:7` and `CONTRIBUTING.md:29` point to the final round. But all of them hard-code `round-3.md` as final, and that is now false (R3-07) | fixed in substance; residual R3-07 |
| R2-02 | `html:449`: "`unknown` attribution in at most 10% of all PR runs in the window (a run counts once if it is `comparison: none` or has a base-gap `unknown` finding)". The separate five-comparable-run minimum is kept. `html:499`: "only when every gate condition above holds (guardian profile)" | fixed |
| R2-03 | Digest inventories now name artifacts as their rounds did:<br>• ADR-008 `ADR.md:51` matches base `round-8.md:5`<br>• ADR-006 `:57` matches `round-5.md`<br>• ADR-007 `:54` matches `round-7.md`<br>ADR-001 `:44` now cites `e1b-v2/scout-design.proposed.md` §15 item 5 (present at `:403`). Residual: ADR-004 `ADR.md:13`, and the README resolution rule | partly → R3-05 |
| R2-04 | `proposal.md:13` ("meaning is unchanged except REQ-37"), `:18-20` (ADR-000 and ADR-009 named), `html:957` (status-line exception on the card), `html:964` ("Apart from the decisions above") | fixed |
| R2-05 | REQ-37 cites "ADR-002, ADR-009" (`html:676`). REQ-42 cites "ADR-002, ADR-008, ADR-009" (`html:529`). ADR-002's status line and card say "amended by ADR-008". ADR-008 `ADR.md:35` and `html:925` reciprocate. The ADR-009 card (`html:949`) names "Appendix 12, experiment list". But ADR-009 `ADR.md:26` and ADR-002 `ADR.md:39-42` do not, so the `proposal.md:70` claim is false | partly → R3-04 |

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R3-01 | **major** | `docs/scout-design.html:672` (REQ-158) | **REQ-158 changes the refresh guard's missingness rule.**<br>• REQ-158 says: "**Guard** is `unmeasured` with fewer than `min_clusters` residual clusters in eligible merges or no fallback-answered cluster".<br>• Base design `:116` applies this condition to the **pending-new rate only**. It says "The fallback-rate numerator counts clusters Jev left `unresolved` or `none`, **whether or not the fallback ran**". It also says skipped clusters (`no model CLI`, `no consent`) "count in the fallback rate only".<br>• The signed interpretation agrees: base `docs/design-reviews/2026-10-31-jev-merge-association/round-8.md` (round-7 verification table, row "What 'It' refers to") reads "the pending-new rate only; the fallback rate stays measured". Its "`min_clusters` floors" row says the two floors cover different sets.<br>• As written, a hosted runner without `codex` would never be able to fire the fallback-rate warning. REQ-158 also contradicts its own first sentence. Appendix A prevails (`html:185`), so this is an undisclosed change of meaning. | Scope the "no fallback-answered cluster / eligible merges" condition to the pending-new rate. State that the fallback rate is evaluated over the whole window once it holds `min_clusters` residual clusters, whether or not the fallback ran. Splitting REQ-158 into one clause per rate is recommended. |
| R3-02 | minor | `html:693-694` (REQ-176, REQ-177) | **X11 loses a signed constraint and cites the wrong ADRs.**<br>• REQ-176 says "one **proxy-selected** before labelling". Base `:444` says "chosen by a **model-free** proxy fixed before labelling".<br>• Both rows cite "ADR-000, ADR-002, ADR-003". X11 was added by ADR-008 (`ADR-008/ADR.md:35`, `html:936`). The Appendix A rule at `:519` requires a requirement to cite the ADR that introduced or changed it. | Restore "a model-free proxy fixed before labelling". Add ADR-008 to the ADR column of REQ-176 and REQ-177. |
| R3-03 | minor | `requirements-map.md:254-260` and the count statements (`:404`; `proposal.md:14`) | **The allow-list mapping is shifted and includes a row that does not exist.**<br>• Base `docs/scout-allowlist.md:19-24` has **six** eligible rows. The map lists rows 1–7 → REQ-70…REQ-76.<br>• Rows 4–6 (`strictness-lowered`, `SUPPRESS.added.unscoped`, `TESTQ.focused-test-added`) are actually in REQ-74, REQ-75 and REQ-76 (`html:596-598`). The map sends them to REQ-73, REQ-74 and REQ-75.<br>• "Row 7" does not exist.<br>• No content is lost: REQ-70…76 carry every row, and the section-level rows cover them. But the disposition record is wrong, and the "389 rows (324 REQ…)" count includes the extra row. | Map:<br>• row 3 → REQ-72, REQ-73<br>• row 4 → REQ-74, REQ-73<br>• row 5 → REQ-75<br>• row 6 → REQ-76<br>Delete "row 7". Update the counts wherever they are stated (`requirements-map.md` Counts, `proposal.md` item 3). |
| R3-04 | minor | ADR-009 `ADR.md:26`; ADR-002 `ADR.md:39-42`; `proposal.md:70` | **The R2-05 fix landed only on the card.**<br>• The card (`html:949`) says ADR-009 supersedes ADR-002's "Appendix 12, experiment list".<br>• ADR-009 `ADR.md:26` names Appendix 12 only.<br>• ADR-002's Changes list records the later effects of ADR-003, 005, 007 and 009. It records neither ADR-008's X11 amendment (which is in its status line) nor ADR-009's experiment-list change.<br>• `proposal.md:70` says the relation "names the experiment list". | In ADR-009 `ADR.md:26`, add the experiment list: base Appendix 11 / §1 "X1–X6, X9, X10" → REQ-42 X1–X6, X9–X11. Add matching ADR-008 and ADR-009 bullets to ADR-002's Changes. The proposal claim then becomes true. |
| R3-05 | minor | ADR-004 `ADR.md:13` (and `:33`); ADR-005 `ADR.md:56`; `docs/adrs/README.md:29` | **Some `@9a3d19f5` pins and digest inventories are still wrong (left over from R2-03).**<br>• **ADR-004 `:13`:** "Correct `docs/scout-design.md`@9a3d19f5 §4's timeout symbol to `defaultMapTimeoutMS…`". At `9a3d19f5`, §4 already holds the corrected symbol (base `:109`). The text that was corrected is the revision-2.3 base, as the same file says at `:31`.<br>• **ADR-005 `:56`:** "The final round also names…" lists base design, `PROPOSAL.md` and `DRAFT-NOTES.md`. It omits `SCOUT_REBUILD_REQUIREMENTS.md` `cd84592f…`, which the round-6 sign-off names as a base (base `round-6.md`, Verdict).<br>• **README `:29`:** says any citation to a deleted design file "reads the file at commit `9a3d19f5`". Frozen rounds that cite `docs/scout-design.md` by line mean an earlier revision. | • ADR-004 `:13`: cite "`scout-design.base.md` (revision 2.3) §4".<br>• ADR-004 `:33` may stay as it is; it is true at `9a3d19f5`. Citing `round-5.md`'s "no companion requirement" is preferred.<br>• ADR-005 `:56`: add the requirements base digest `cd84592f…` as round 6 names it.<br>• README `:29`: say that `9a3d19f5` holds each file's last version, and that a frozen record citing an earlier revision means that revision. Point readers to the folder's `scout-design.base.md`/`proposed.md` snapshots or to git history. |
| R3-06 | minor | `docs/jev-adapter-audit.md:69,71` (rows 28 and 30, both **Fixed**) | **A governed document changed but is not in the reviewed set.**<br>• `AGENTS.md:35-40` governs "the Met/Fixed status and Design-level rows" of this file.<br>• Base `:69` "(design §6 `:162`)" is now "(REQ-96)". Base `:71` "design §6" is now "REQ-86, REQ-99".<br>• The rewrite is disclosed (`proposal.md` item 5), but it is not in the digest list, so no sign-off covers it.<br>• Separately, line 7 now says `docs/scout-design.html` "was not edited". That file did not exist when the audit was written. | Add `docs/jev-adapter-audit.md` and its SHA-256 to the reviewed files, and state that the edit is citation-only. Recommended: fix the anachronism on line 7 in the same edit, for example "the design document (then `docs/scout-design.md`)". |
| R3-07 | minor | ADR-009 `ADR.md:39`; `docs/STATUS.md:7`; `CONTRIBUTING.md:29` | **The final-round pointers name `round-3.md`, which is not the final round.** This round finds issues. As a result, ADR-009's signed summary, the STATUS digest pointer and the CONTRIBUTING fresh-machine check all point to a round that does not sign off. | Point all three to the round that actually signs off with zero issues (expected `round-4.md`), or use number-free wording such as "the zero-issue final round in this folder". Resubmit the new ADR-009 `ADR.md` digest. |

## Answers to the proposal's questions

1. **Is normative content lost, weakened or changed?** Yes, in two places:
   - R3-01: the refresh-guard missingness rule, a change of meaning.
   - R3-02: X11's model-free selection proxy.

   The disposition record has a mapping error (R3-03). The other sampled rows carry the base content intact:
   - REQ-6, 13, 20–24, 52, 54, 70–77, 84, 94, 96, 105, 107, 120, 127, 135 and 137;
   - REQ-22's demotion rule and fire-rate proxy;
   - REQ-84's gate-disable tracking.
2. **Do the ADR summaries and cards match their records?**
   - Supersession endpoints agree in both directions for all ten ADRs.
   - Signed digests match their frozen rounds.
   - Exceptions: relation scope (R3-04) and residual provenance (R3-05).
3. **Does the Overview agree with Appendix A?** Yes. The R2-02 restatements, the triggers, the LLM column, the exit codes and the ledger record types all agree with their REQs.
4. **Are the rule, tooling and citation changes complete, with no live reference to a deleted file?**
   - No live reference to a deleted file remains outside frozen or historical records. The only mentions are `html:630,953,956` and `remaining-work.md:101`, all intentional.
   - No stale `design §`/`req §`/`AL §` citation remains in live files.
   - The governed audit file is not in the reviewed set (R3-06).
   - The pointers to the final round are wrong (R3-07).
5. **Experiment rules, ledgers, frozen records:** no violation found.
   - Experiment rules 1–7 are word for word; only the scope line, ledger rule 2 (X9–X11) and the Jev citation changed.
   - The research ledger only gains R-062; the hypothesis ledger only gains one History line.
   - Frozen rounds contain no new-path strings.

## Considered and not raised

- **ADR-003 `ADR.md:53` omits the `PROPOSAL.md`/`DRAFT-NOTES.md` digests** (peer F4). The round-7 sign-off statement names exactly the four digests the summary lists. Round 7 itself declined an omission of this kind as a nit.
- **Other `@9a3d19f5` pins** (ADR-006 `:34-35`, ADR-008 `:31-33`): at that commit, the base Appendix items 1, 20 and 22(a) still contain the superseded clause with its supersession marker (base `:478-500`), so the pins point to the right content.
- **REQ-3 ("CI/nightly runs are differential") against REQ-160 ("Nightly is full")**: this tension is carried word for word from base REQ-3 (`SCOUT_REBUILD_REQUIREMENTS.md:11`). It is not introduced by ADR-009 and is not reopened here.
- **The ADR-008 card uses the `prov` colour** (the legend says "proposed") while its text reads "accepted · provisional". This is cosmetic. The text, the status line and README `:20` are correct.
- **The ADR-009 card footer gives no round outcome.** The card summarises `ADR.md`, which defers the outcome to the final round by rule.
- **"Supersedes" against "amends" wording between cards and `ADR.md` Changes**: wording only, as round 2 accepted.

## Peers

- **`innovation-challenger`:**
  - Found R3-01, R3-02 and R3-03, and independently found R3-04.
  - Sampled 19 map rows from base source text to target REQ.
  - Rated R3-01 major. The lead agrees: following round-1 precedent (R1-01, R1-02), an undisclosed change in the meaning of an Appendix A requirement is major.
- **`experiment-peer`:**
  - Found R3-05 (the ADR-004 pin and the ADR-005 inventory), R3-06 and R3-07, and independently found R3-04.
  - Confirmed the digest inventories against every frozen round.
  - Confirmed the counts: 389 table rows, 22 dropped, 43 ADR, 312 JSON keys, and REQ-1…REQ-183 each present exactly once.
  - Its F4 for ADR-003 was not adopted (see above).
- **Lead:**
  - Found R3-04 and the R3-07 pointer problem independently, and added the README `:29` part of R3-05.
  - Re-read every peer finding against base and frozen records: base design `:116`, `:444`, `:476-500`; round-8 verification table; allow-list `:19-24`; audit `:69,71`; ADR-005 round-6 Verdict.
- **Effect of the peers:** without them the lead would have reported 2 issues (R3-04 and R3-07, with part of R3-05). The challenger supplied the only major issue.

## Verdict

**7 issues (1 major: R3-01; 6 minor: R3-02 to R3-07): NOT signed off.**

No sign-off applies to the digests above. In addition, no digest was recomputed in this round.

---

## Council memo

1. **Recommendation:**
   - Fix R3-01 to R3-07 and submit round 4.
   - Name new digests for:
     - `docs/scout-design.html`, `proposal.md`, `requirements-map.md`, ADR-009 `ADR.md`, `docs/adrs/README.md`;
     - the ADR-002, ADR-004 and ADR-005 `ADR.md` files;
     - `docs/jev-adapter-audit.md`;
     - plus the unchanged files.
   - Recompute every digest with `shasum -a 256` before submitting.
2. **Alternatives:**
   - (a) Sign off with R3-01 recorded as a condition. Rejected: rule 1 requires zero issues, and Appendix A prevails, so the changed guard semantics would become normative.
   - (b) Treat R3-03 as cosmetic. Rejected: the map is ADR-009's evidence that nothing was lost, and its counts are quoted in signed text.
   - (c) Use number-free final-round wording (R3-07). Preferred: it stops each new round from invalidating the pointers.
3. **Evidence:** each issue row cites exact worktree and base `file:line`. The verification table cites the R2 fix locations.
4. **Risks:**
   - If R3-01 stays, an implementation following REQ-158 may suppress the fallback-rate warning when no `codex` or no consent is available. The lane is opt-in and off by default, which limits the exposure.
   - Digests are still unverified. If the staged bytes differ from the submitted list, round 4 must review the actual bytes.
   - Every fix is a text edit, so all are reversible.
5. **Execution outline:**
   1. Rewrite REQ-158 per R3-01, and check REQ-157/158 against base `:116` sentence by sentence.
   2. Restore REQ-176's "model-free proxy" and add ADR-008 to REQ-176/177.
   3. Correct `requirements-map.md:254-260` and the counts in the map and the proposal.
   4. Update ADR-009 `ADR.md:26` and ADR-002's Changes (R3-04), and make `proposal.md:70` true.
   5. Apply the R3-05 edits (ADR-004 `:13`, ADR-005 `:56`, README `:29`).
   6. Add `docs/jev-adapter-audit.md` to the reviewed set (optionally fixing line 7).
   7. Repoint `ADR.md:39`, `STATUS.md:7` and `CONTRIBUTING.md:29` (R3-07).
   8. Run `shasum -a 256` on every reviewed file and `git -C …/scout-req diff --cached --stat`, and confirm no other governed file is staged.
   9. Submit round 4 with the "Round 3 fixes" section and the digest list.
