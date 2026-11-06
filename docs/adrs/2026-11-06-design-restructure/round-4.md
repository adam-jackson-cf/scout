# Round 4: ADR-009, one design document with requirements and ADR appendices

Date: 2026-11-06. Rules: `AGENTS.md`/`CLAUDE.md` design rules 1–3. The innovation council (lead) ran this review with two independent peers: `innovation-challenger` (GPT, adversarial review of normative content) and `experiment-peer` (Opus, provenance, ADR records and rules). The review was read-only. It made no model calls outside the council and did not touch `.codegraph`.

**I could not recompute the digests.** Neither the lead nor either peer had a shell. No file was hashed, `git diff --cached --stat` was not run, and nobody checked whether `CLAUDE.md` is a symlink. The digests below are the ones the main agent submitted. A mismatch cannot be detected or ruled out.

| File | SHA-256 (submitted; not recomputed) |
|---|---|
| `docs/scout-design.html` | `99f089517e67be2ec26f000f50414ba0f6c3792ab977dfa1dc8b6357cbfbd63f` |
| `docs/adrs/2026-11-06-design-restructure/proposal.md` | `68fef4a59697f0e594bff5ae0e1dbf708582c1d258af15af62a299fbe8f4d523` |
| `…/requirements-map.md` | `8a46cea4f0281ad3b27c11e37e167ba658d03e98286cd0aac267035f3b48b3dd` |
| `…/design-section-map.json` | `3cc0225e4bb98d6e39fbb6e71d870937adb84093cd9d301a85c026fb204da974` |
| `…/ADR.md` (ADR-009) | `e0d2d7f12d9185527e5c48138cdf54aa738c5235aae1e11ccf1ceeb1a17e1bfc` |
| `AGENTS.md` | `aff517d34b71ac2ed189dcc9f4d3b03380e9b91e9b754a2bd98aceeaa4718511` |
| `docs/adrs/README.md` | `7c90063cd3050dd05984990dade0553e47e4fc7d43051b9e9c92f241c2e307e0` |
| `docs/jev-adapter-audit.md` | `b1f4ec433c787e46700623e24e436fd2fb2b73759c5a3f3c2524e3d7de3973d8` |
| ADR-000 `2026-10-19-v1-decisions/ADR.md` | `4d631ce92da11fc7c42467d2b18cc12902f4e193d92170a7e5e33a9a8547b36f` |
| ADR-001 `2026-10-21-e1b-v2/ADR.md` | `e5db1e7659fb562bd3c89cb36483d789d0e669d3208c8f2a9e0617346d2be8fa` |
| ADR-002 `2026-10-22-rescope/ADR.md` | `4adb8f70b90d2ca4deb941b56d161ccce424549f34c323b53cd14c4c80a9ce2b` |
| ADR-003 `2026-10-23-decisions/ADR.md` | `e8d5afa224b927fd992ca79df6b6bdcb8b9e2bf7d284d44e3ede6c701f6f026e` |
| ADR-004 `2026-10-23-clawpatch-patch-v3/ADR.md` | `cbaf321eed0957fa29c2ac7bca02b981f3acd8c8efc895e10bc7a1bf685e577a` |
| ADR-005 `2026-10-23-codex-luna-mapping/ADR.md` | `ff8a06e034a20dc0242732a71743b7f99c1cc82d1a5293eab6d29a272e34e00c` |
| ADR-006 `2026-10-24-area-refresh-fallback/ADR.md` | `371a870bd4302d5bf3a6f224679b6ac2ad321a0c3b1dd4aeb6e70ce4941e9a82` |
| ADR-007 `2026-10-24-cutover-cleanup/ADR.md` | `de9fffdf63fc0ec61a98bdde7327617682d7b336b1199711537d48b26ee04721` |
| ADR-008 `2026-10-31-jev-merge-association/ADR.md` | `76dbd3aec2503d447270a91bdfb9a9c44e35ca730c45199add44a34847c45ab0` |

**The digest list matches the R3 fixes.** Eight files are unchanged since round 3, as expected, because no R3 fix touched them:
- `design-section-map.json` and `AGENTS.md`;
- ADR-000, ADR-001, ADR-003, ADR-006, ADR-007 and ADR-008.

The changed files match the fixes:
- html: R3-01 and R3-02;
- `proposal.md` and `requirements-map.md`: R3-03 and R3-04;
- ADR-009 `ADR.md`: R3-04 and R3-07;
- ADR-002: R3-04;
- ADR-004, ADR-005 and the README: R3-05;
- the audit, which is newly listed: R3-06.

## Scope

- **Under review:** worktree `/Users/adamjackson/Projects/work/scout-req`, branch `agent/requirements-path`, all changes staged. This covers:
  - the "Round 3 fixes" in `proposal.md` and every location they touched;
  - a new inconsistency check of the fixes: REQ-157/158 against base design `:116`, the frozen round-8 interpretation and the Overview; map counts against the proposal; ADR-002/008/009 relations on the cards against the `ADR.md` files;
  - a fresh broad pass over proposal questions 1–5.
- **Base:** `main` at `9a3d19f5` in `/Users/adamjackson/Projects/work/scout/`, including the deleted originals and the frozen `docs/design-reviews/` rounds.
- **Checks made:**
  - every R3 required fix, at each location it named;
  - supersession and amendment, in both directions, for all ten ADRs (card headers `html:736–949` against each `ADR.md` status line and Changes section);
  - digests against frozen rounds for ADR-002, ADR-004, ADR-005 and ADR-008;
  - every `@9a3d19f5` pin in all `ADR.md` files;
  - experiment rules 1–7, word for word against base `CLAUDE.md`;
  - that the ledgers are append-only;
  - live references to deleted files and stale citation forms in `internal/`, `cmd/`, `.github/`, `scripts/`, `toolchain/`, `docs/plans`, `docs/references`, STATUS, README, CONTRIBUTING and the audit;
  - the map counts (recounted);
  - 30 further legacy-REQ rows, traced from the base `SCOUT_REBUILD_REQUIREMENTS.md` text to the target REQ (challenger);
  - the measurement formulas (base design `:247-271`), experiment thresholds (`:435-444`), phase scenarios (`:372-404`) and allow-list rows (`AL :19-24`);
  - the ADR column of REQ-10, 13, 28, 31, 52–60, 167, 169–177.
- **Not reopened:** decisions settled in ADR-000…ADR-008.

## Verification of round-3 fixes

| # | Evidence | Result |
|---|---|---|
| R3-01 | **Window and fallback rate.** `html:671` (REQ-157) defines the fallback rate as "clusters Jev left unresolved/none ÷ residual clusters, counted whether or not the fallback ran". Pending-new uses only merges where the fallback was not skipped and answered every escalated cluster.<br>**Floors and unmeasured.** `html:672` (REQ-158): the fallback rate "is evaluated over the whole window once it holds at least `min_clusters` residual clusters". The pending-new rate "is `unmeasured` while its eligible merges hold fewer than `min_clusters` residual clusters or no fallback-answered cluster". Jev-failure, Jev-cap and PR-cap clusters enter neither rate. Fallback-cap, Terra-failure, no-CLI and no-consent clusters count only in the fallback rate.<br>**Base match.** This matches base design `:116` sentence by sentence and frozen `2026-10-31-jev-merge-association/round-8.md:22-28` ("pending-new rate only"; different floor populations).<br>**No new inconsistency:**<br>• REQ-156 (reset and generation) and REQ-159 (20/20/0.40/0.20 [ASSUMPTION]) are unchanged.<br>• The ADR-008 card (`html:942-943`) and ADR-008 `ADR.md` Conditions agree.<br>• The Overview does not restate the guard rates (`html:281-284, 354-355`). | fixed |
| R3-02 | `html:693` "one chosen by a model-free proxy fixed before labelling" = base `:444`. REQ-176 and REQ-177 (`html:693-694`) cite "ADR-000, ADR-002, ADR-003, ADR-008" | fixed |
| R3-03 | **Mapping.** `requirements-map.md:254-259`: rows 1–6 map to REQ-70, REQ-71, REQ-72/73, REQ-74/73, REQ-75 and REQ-76. This matches base `docs/scout-allowlist.md:19-24` and `html:592-598`. No row 7 remains.<br>**Counts.** The table runs over lines 11–398, which is 388 rows. Grep counts 43 `ADR` rows (lines 11, 13, 48–58, 78–80, 128–129, 155, 157, 269, 276, 282, 339, 341–342, 360–364, 366–370, 376, 381–383, 386–388) and 22 `dropped` rows (281, 286–287, 340, 343–359, 365). No other disposition exists, which leaves 323 `REQ` rows.<br>**Agreement.** `:403` and the proposal (item 3 and R3-03) agree. The challenger independently got 312 JSON keys and REQ-1…183 each defined once | fixed |
| R3-04 | ADR-009 `ADR.md:26` names "Appendix 12 … and its experiment list. Base design §1 and Appendix 11 said 'X1–X6, X9, X10'; REQ-42 adopts X1–X6, X9–X11". ADR-002 `ADR.md` Changes contains the ADR-008 bullet ("adds X11") and the ADR-009 bullet ("replaces the experiment list"). ADR-002's status line and card (`html:783`) say "amended by ADR-008". The ADR-008 card (`html:925`) says "amends ADR-002 (adds X11)". The proposal claims are now true | fixed |
| R3-05 | **ADR-004.** `ADR.md:13` cites "`scout-design.base.md` (revision 2.3) §4". The rev-2.3 base `:97` holds the old `defaultMapCodexTimeoutMS` (`clawpatch-patch-v3/PROPOSAL.md:21`, P6).<br>**ADR-005.** `ADR.md:56` lists `5c16a807…`, `cd84592f…`, `1a59951c…` and `6c5ec06f…`. These equal base `codex-luna-mapping/round-6.md:3,50`.<br>**README.** `README.md:29` now distinguishes the last version (`9a3d19f5`) from the revision current when a frozen record was written.<br>**Residue.** The old rule survives in two other places (R4-04) | fixed at named locations; residue → R4-04 |
| R3-06 | **In the reviewed set.** The audit is now listed (`b1f4ec43…`).<br>**Edits against base.** Line 6 changes the path (`docs/design-reviews/` → `docs/adrs/`). Line 7 now reads "(then `docs/scout-design.md`, now `docs/scout-design.html`) … was not edited by this audit". Row 28 (`:69`) changes "design §6 `:162`" to REQ-96. Row 30 (`:71`) changes "design §6" to "REQ-86, REQ-99". REQ-96 carries "never the probability number" and REQ-86/REQ-99 carry realised order and the orderflip harness.<br>**Residue.** The audit still cites the deleted design's risk R14 (R4-02) | fixed at named rows; residue → R4-02 |
| R3-07 | **Pointers.** These now use number-free wording:<br>• ADR-009 `ADR.md:38-39`: "the zero-issue final round in this folder (`round-<N>.md` with the highest N)";<br>• `docs/STATUS.md:7`: "signed digest in the zero-issue final round of `docs/adrs/2026-11-06-design-restructure/`";<br>• `CONTRIBUTING.md:29`: "the zero-issue final round of the latest ADR that changed the design".<br>**Agreement.** Rule 2 (`AGENTS.md:44-46`, `CLAUDE.md:44-46`) and README `:3` agree. No live file names `round-3.md` as final | fixed |

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R4-01 | minor | `docs/scout-design.html:700` (REQ-60), `:713` (REQ-167), `:687` (REQ-170), `:688` (REQ-171) | **Four requirements leave out the ADR that introduced their content.** This is the same class of defect as R3-02. The convention at `html:519` and in R3-02's fix, which legacy rows already follow, is that a requirement cites the ADR that introduced or changed it: REQ-10 cites ADR-007 and ADR-008; REQ-13 cites ADR-006 and ADR-008; REQ-31 cites ADR-002, ADR-005, ADR-007 and ADR-008.<br>• **REQ-167** cites ADR-000, ADR-003 and ADR-007. Its Phase-4 acceptance (one onboarding map, none on merges, PR or nightly, one manual map; base `:393-394`) is ADR-006's change. The ADR-006 card says it "replaces Phase 4's debounced-burst acceptance" (`html:888`; ADR-006 `ADR.md` Changes).<br>• **REQ-170** cites ADR-000, ADR-002 and ADR-003. X2b and "X2b governs the pinned luna/xhigh mapper" come from ADR-005 (`ADR-005/ADR.md` Changes: "adds the planned X2b experiment"). "Otherwise agent maps are onboarding/manual only" comes from ADR-006 (base `:436-437`: "Applied by revision 2.6").<br>• **REQ-171** cites ADR-000, ADR-002 and ADR-003. "Merge-time files instead follow X11" is a revision-2.8 carve-out (base `:438`), from ADR-008.<br>• **REQ-60** cites ADR-002, ADR-004 and ADR-007. "Mapper and association fallback are codex-only exceptions" comes from ADR-005 (mapper) and ADR-008 (fallback). | Add ADR-006 to REQ-167. Add ADR-005 and ADR-006 to REQ-170. Add ADR-008 to REQ-171. Add ADR-005 and ADR-008 to REQ-60.<br>Then make one systematic pass. For each ADR-004…ADR-008, take the sections and Appendix items its `ADR.md` Changes list names, and use `requirements-map.md` to find the REQs those map to. Any REQ whose text carries the changed clause must cite that ADR. List any further rows corrected in the round-4 fixes section. |
| R4-02 | minor | `docs/jev-adapter-audit.md:19` and `:79` (row 38, **Design-level**); `internal/decisions/client.go:20`; `internal/policy/policy.go:330`; `internal/guardrails/focused_tests.go:33` | **Five live files still cite the deleted design and allow-list in forms the rewrite missed.**<br>• **The citations:** "(design R14)" at audit `:19`, at audit `:79` in a governed Design-level row, and at `client.go:20`; "AL table-1 list" at `policy.go:330`; "(allow-list table 1)" at `focused_tests.go:33`. All five are unchanged from base.<br>• **Why it matters:** `scout-design.html` has no R14. `requirements-map.md:365` drops the §12 risk statements and maps only their mitigations. `design-section-map.json` has keys for both forms: `"§12 R14"` at `:929` → REQ-87, REQ-29, REQ-30, and `"AL table 1"` at `:2078` → REQ-70…78.<br>• **The rewrite was meant to cover these:** base `internal/guardrails/suppress.go:1-2` "(design §5.1/§5.2, allow-list table 1)" is now "(REQ-62, REQ-72–REQ-76)".<br>• **Claims this makes false:** proposal item 5 ("every … citation"), the ADR-009 card `html:965`, the audit's "citation-only" scope, and the round-3 answer to question 4. | Repoint every citation through `design-section-map.json`:<br>• audit `:19` and `:79` → REQ-87, plus REQ-99 for row 38, which REQ-99 keeps open;<br>• `client.go:20` → REQ-87 and REQ-92;<br>• `policy.go:330` → REQ-70–REQ-76;<br>• `focused_tests.go:33` → REQ-76.<br>Grep the live tree again for `design R\d`, `table[- ]1`, `AL `, `req §`, `design §` and `Appendix \d`. Resubmit the audit digest (still citation-only). |
| R4-03 | minor | ADR-003 `docs/adrs/2026-10-23-decisions/ADR.md:54` | **ADR-003's adoption record points to the wrong STATUS sections.** It says "`docs/STATUS.md`@9a3d19f5 §2 and §6.9 record adoption".<br>• At `9a3d19f5`, §2 (`STATUS.md:129-135`) lists the decisions in force and §6.9 (`:186-195`) records the user's decisions. Neither records the adoption of revision 2.3.<br>• The adoption record is §1 `:25`: "Revision 2.3 (history), adopted 2026-10-23 … equals the digest signed off … in `round-7.md`".<br>• ADR-004 `:51`, ADR-005 `:58` and ADR-007 `:56` already cite §1 history entries. | Cite "`docs/STATUS.md`@9a3d19f5 §1, revision 2.3 history entry". Resubmit the ADR-003 digest. |
| R4-04 | minor | `docs/scout-design.html:966` (ADR-009 card); `proposal.md:43` ("Not changed") | **The resolution rule that R3-05 corrected is still stated in two other files.** Both say "`docs/adrs/README.md` maps them, and deleted files are read at `9a3d19f5`", and both apply this to frozen records.<br>• README `:29` now says a frozen record citing a deleted file "means the revision current when the record was written".<br>• The README is right. For example, frozen ADR-004 rounds cite the rev-2.3 design by line, and that text is at `scout-design.base.md:97`, not at `9a3d19f5`. | Align both statements with README `:29`, for example "`docs/adrs/README.md` maps them and says which revision a frozen citation means". Resubmit the html and proposal digests. |

## Answers to the proposal's questions

1. **Is normative content lost, weakened or changed?** No change of meaning was found.
   - R3-01 and R3-02 are restored exactly.
   - The challenger traced 30 more legacy-REQ rows (REQ-1/2/4, 5/7–11, 12/14/15, 16–19, 25/26, 28–32, 33/34/36, 38–41). It also checked the formulas (base `:247-271`), the experiment thresholds (`:435-444`), the phase scenarios (`:372-404`) and allow-list rows 1–6. The only finding was REQ-2's "opt-in", which is not adopted (see "Considered and not raised").
   - The ADR provenance column is incomplete in four rows (R4-01).
2. **Do the ADR summaries and cards match their records?**
   - Supersession and amendment agree in both directions for all ten ADRs, and README `:15-24` statuses match.
   - The signed digests match their frozen rounds.
   - Exceptions: ADR-003's adoption pointer (R4-03) and the card's resolution rule (R4-04).
3. **Does the Overview agree with Appendix A?** Yes. The gate diagram, the activation callout (`html:449`), the comparable-base popover, exit precedence, the onboarding/manual map schedule, the push-only opt-in association and the ledger records (`:454-476`) all agree with their REQs. The Overview states no guard formula that could compete with REQ-157/158.
4. **Are rule, tooling and citation changes complete, with no live reference to a deleted file?** Not quite: five live citations to the deleted risk R14 and to allow-list "table 1" remain (R4-02). Otherwise:
   - deleted-path mentions remain only where intended (`html:630,953,956`, `remaining-work.md:101`, earlier ledger rows, audit line 7);
   - `CP07 §5` in `internal/areas` is an evaluation citation, not a design citation.
5. **Experiment rules, ledgers, frozen records:** no violation.
   - `AGENTS.md` and `CLAUDE.md` experiment rules 1–7 match base `CLAUDE.md` word for word. Only the scope line, design rules 1–2, ledger rule 2 (X9–X11) and the Jev-labelling citation changed, and all are disclosed.
   - The research ledger only gains R-062, and the hypothesis ledger only gains one History line.
   - Each frozen folder has the base file set plus `ADR.md`, and no frozen round contains new-path strings.

## Considered and not raised

- **REQ-2 drops "Language support is opt-in and minimal"** (challenger F1, proposed as major). Not adopted, for three reasons:
  - Round 1 already judged this wording only (`round-1.md:258`).
  - In base REQ-2 (`SCOUT_REBUILD_REQUIREMENTS.md:10`), the phrase is followed by a colon and a list: tier 1, CC/NLOC/duplication on seven languages, "tools provisioned only for the detected stack". The worktree REQ-2 (`html:527`) keeps all three items.
  - The design, which the base requirements file says prevails (`:3`), has no language opt-in mechanism. The base design capability matrix (`:145-154`), the v1 design and `internal/policy` contain no opt-in for languages; the only opt-in there is for semantic and model lanes, which REQ-17 and REQ-62–65 carry.

  The challenger disagreed. If the user intends a per-language opt-in, that would be a new design decision, not a migration defect.
- **`CLAUDE.md` is not in the digest list** (peer F4). Its text reads the same as `AGENTS.md`, rounds 1–3 accepted `AGENTS.md` alone, and nobody could tell whether it is a symlink. I made this a condition instead of an issue (see "Required with the next submission").
- **ADR-002's Changes records later ADRs, although rule 2 allows only status-line edits.** Every `ADR.md` is first created in this change set, so nothing accepted is being edited. R3-04 required these bullets.
- **Nits:**
  - ADR-002 `ADR.md:38` and `:43` both mention that ADR-009 ends the cutover.
  - Titles differ between cards, `ADR.md` files and the README (ADR-000, 002, 003, 008, 009).
  - Audit row 28 could also cite REQ-30.
  - ADR-009 `ADR.md` Changes does not mention the audit edit; proposal items 5 and 6 do.
  - The proposal's R3-06 line omits the line-6 path change.
  - The proposal's "Round 2 fixes" R2-01 bullet still names `round-3.md`. It is a historical record that the R3-07 bullet supersedes.
- **STATUS, CONTRIBUTING, README, ledgers and `.omp/*` are not in the digest list.** Design rule 1 does not govern them. Precedent (`2026-10-22-rescope/round-6.md:48`) names such companion edits "as reviewed".
- **REQ-58 and REQ-60 also cite ADR-004.** Citing an extra ADR does not hide provenance; only missing ADRs are raised (R4-01).

## Peers

- **`innovation-challenger`:**
  - Recounted the map, the JSON keys and the REQ IDs.
  - Verified R3-01 sentence by sentence against base `:116` and round-8.
  - Traced 30 further legacy rows.
  - Found three of the four R4-01 rows (REQ-167, REQ-170, REQ-171); the lead found REQ-60 while checking them.
  - Proposed REQ-2 "opt-in" as major; not adopted (see above).
- **`experiment-peer`:**
  - Verified the R3 fixes and all supersession edges, pins and frozen-round digests.
  - Found R4-02, R4-03 and R4-04, and raised the `CLAUDE.md` coverage question.
- **Lead:**
  - Re-checked every peer finding against base and frozen records: base design `:116`, `:393-394`, `:423`, `:436-438`; base STATUS `:25`, `:129`, `:186`; `design-section-map.json:929,2078`; base and worktree `suppress.go:1-2`; ADR-006 card `:888`; README `:29`.
  - Grepped the live tree for citation forms on its own.
  - Had noticed the `proposal.md:43` wording, but had judged it consistent until the peer showed that `html:966` states it as a rule for frozen records.
- **Effect of the peers:** without them, the lead would have reported no issues this round. Every raised issue originates with, or is confirmed by, a peer.

## Verdict

**4 issues (all minor: R4-01 to R4-04): NOT signed off.**

No sign-off applies to the digests above, and none of them was recomputed in this round.

### Required with the next submission

- The new digests change for:
  - `docs/scout-design.html` (R4-01, R4-04);
  - `proposal.md` (R4-04, plus a "Round 4 fixes" section);
  - ADR-003 `ADR.md` (R4-03);
  - `docs/jev-adapter-audit.md` (R4-02).

  The other listed digests should stay as they are.
- `CLAUDE.md`: either its SHA-256 (expected to equal `AGENTS.md`'s `aff517d3…` if the bytes are identical), or a statement that it is a symlink to `AGENTS.md` (`test -L CLAUDE.md`).
- R4-02's code-comment edits are comment-only. List them in the fixes section; they are not hashed.

---

## Council memo

1. **Recommendation:** fix R4-01 to R4-04 and submit round 5.
   - Use the systematic ADR-column pass in R4-01 instead of a four-row patch, so that round 5 does not find more rows of the same kind.
   - Recompute every digest with `shasum -a 256`, run `git -C …/scout-req diff --cached --stat`, and settle the `CLAUDE.md` question before submitting.
2. **Alternatives:**
   - (a) Sign off with R4-01…04 as conditions. Rejected: rule 1 requires zero issues.
   - (b) Treat the ADR column as informative only and drop R4-01. Rejected: `html:519` makes it a rule, and R3-02 applied it.
   - (c) Adopt the challenger's REQ-2 "opt-in" finding. Rejected because of round-1 precedent and because no design backing exists. It stays open to the user as a new decision.
   - (d) For R4-02, delete the "(design R14)" citations instead of repointing them. Acceptable for the code comments, but audit row 38 needs an anchor in the design; REQ-99 serves.
3. **Evidence:** each issue row cites exact worktree and base `file:line`. The verification table cites each R3 fix location and the recount.
4. **Risks:**
   - The digests are still unverified. If the staged bytes differ from the list, round 5 must review the actual bytes.
   - R4-01 is a provenance gap, not a change of meaning. Its risk is that a future ADR author misses an interacting decision.
   - R4-02 leaves a governed Design-level row anchored to a risk that no longer exists.
   - All fixes are text edits and fully reversible.
5. **Execution outline:**
   1. Add ADR-006 to REQ-167, ADR-005 and ADR-006 to REQ-170, ADR-008 to REQ-171, and ADR-005 and ADR-008 to REQ-60. Run the ADR-004…008 Changes-to-REQ pass and record any further rows.
   2. Repoint audit `:19` and `:79`, `client.go:20`, `policy.go:330` and `focused_tests.go:33` through `design-section-map.json`. Grep the live tree again for `design R\d`, `table[- ]1`, `AL `, `req §`, `design §` and `Appendix \d`.
   3. In ADR-003 `ADR.md:54`, cite "`docs/STATUS.md`@9a3d19f5 §1, revision 2.3 history entry".
   4. Align `html:966` and `proposal.md:43` with README `:29`.
   5. Add a "Round 4 fixes" section to `proposal.md`.
   6. Run `shasum -a 256` on every reviewed file plus `CLAUDE.md` (or `test -L CLAUDE.md`), and run `git diff --cached --stat`. Confirm that no other governed file is staged.
   7. Optionally run `go build ./... && go vet ./...`; the code edits are comment-only.
   8. Submit round 5 with the digest list.
