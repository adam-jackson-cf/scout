# Round 10: ADR-009, one design document with requirements and ADR appendices

Date: 2026-11-06. This round follows design rules 1–3 in `AGENTS.md`/`CLAUDE.md`. The innovation council lead ran the review with two independent peers:
- **`innovation-challenger` (GPT):** checked the round-9 rewrites clause by clause against the base. That covered REQ-89, REQ-90, REQ-91, REQ-98, REQ-42 and REQ-167…REQ-182, set against `D:173-177,184,377-402,435-444`. It checked the opposite direction too, for run-result narrative turned into requirements. It traced the ADR columns through the frozen `proposed.diff` files and checked the new §6/§11/§13 clause inventory in `completeness-audit.md`.
- **`experiment-peer` (Opus):** checked R9-06…R9-09. It recounted the audit totals, swept every parent/child group of the map and JSON, and checked the Overview against the REQs changed in round 9. It also checked the proposal's "Round 9 fixes" and answered questions 2, 4 and 5.

The review was read-only. It made no model calls outside the council and did not touch `.codegraph`.

**The main agent recomputed the digests; the council did not.** The lead and both peers had read-only tools and no shell, so none of them hashed a file. Earlier rounds accepted a main-agent recomputation, so this round does too. The main agent reported these outputs on the staged worktree:
- `shasum -a 256` on each file below;
- `test -L CLAUDE.md && readlink CLAUDE.md` → `AGENTS.md`;
- `git diff --cached --stat` → `249 files changed, 7245 insertions(+), 1781 deletions(-)`. That is round 9's 248 files plus the new `round-9.md`.

It also reported:
- 183 distinct REQ rows and 10 ADR cards in a headless render;
- a script check with 0 gaps on two rules: each parent map row contains the union of its children, and each JSON key contains its map row's IDs and its child keys' IDs;
- 312 JSON keys and 388 map rows (323 REQ, 43 ADR, 22 dropped);
- no Go file changed since round 6.

| File | SHA-256 (main agent, `shasum -a 256`) | vs round 9 |
|---|---|---|
| `docs/scout-design.html` | `ea73bd3bf6c9264ddeec6265856958e9c2fddf6e06f13509a613321eaac4e2c9` | changed (R9-01…R9-06, R9-08) |
| `docs/adrs/2026-11-06-design-restructure/proposal.md` | `1d2455c8ecd4a92a5c51c1b84a2ffb10aa10c83913abd1dff18c7696cf7385d2` | changed ("Round 9 fixes") |
| `…/requirements-map.md` | `ef6358967e45b2eefaab2b6e5c3842295b21d227f0386b92a8d14a2686bba36d` | changed (R9-07) |
| `…/design-section-map.json` | `d665128f0311579922db65fe325ac74e4f9da8d0be163d9bd6d1ecbbe18927d0` | changed (R9-07) |
| `…/ADR.md` (ADR-009) | `1393b6eb7488688c24ef25106ec6861616ae61e75675570651e5b97ab1fae83a` | unchanged |
| `…/completeness-audit.md` | `419ea41c4f84f9f1e33e22281d1af774db32eb516756e6681d56665de45af22c` | changed (R9-06, R9-09, inventory) |
| `AGENTS.md` (`CLAUDE.md` symlink) | `aff517d34b71ac2ed189dcc9f4d3b03380e9b91e9b754a2bd98aceeaa4718511` | unchanged |
| `docs/adrs/README.md` | `7c90063cd3050dd05984990dade0553e47e4fc7d43051b9e9c92f241c2e307e0` | unchanged |
| `docs/jev-adapter-audit.md` | `30894dbf17bf9f815ceaadce2c3c9ae75d2cf0e31d4a91753985dbfe6818630d` | unchanged |
| ADR-000 `2026-10-19-v1-decisions/ADR.md` | `4d631ce92da11fc7c42467d2b18cc12902f4e193d92170a7e5e33a9a8547b36f` | unchanged |
| ADR-001 `2026-10-21-e1b-v2/ADR.md` | `e5db1e7659fb562bd3c89cb36483d789d0e669d3208c8f2a9e0617346d2be8fa` | unchanged |
| ADR-002 `2026-10-22-rescope/ADR.md` | `4adb8f70b90d2ca4deb941b56d161ccce424549f34c323b53cd14c4c80a9ce2b` | unchanged |
| ADR-003 `2026-10-23-decisions/ADR.md` | `f42a2ea88e1b02d9a41b558841f32c7393a02cce58f7ccdd0c401caa679fdfe7` | unchanged |
| ADR-004 `2026-10-23-clawpatch-patch-v3/ADR.md` | `cbaf321eed0957fa29c2ac7bca02b981f3acd8c8efc895e10bc7a1bf685e577a` | unchanged |
| ADR-005 `2026-10-23-codex-luna-mapping/ADR.md` | `ff8a06e034a20dc0242732a71743b7f99c1cc82d1a5293eab6d29a272e34e00c` | unchanged |
| ADR-006 `2026-10-24-area-refresh-fallback/ADR.md` | `371a870bd4302d5bf3a6f224679b6ac2ad321a0c3b1dd4aeb6e70ce4941e9a82` | unchanged |
| ADR-007 `2026-10-24-cutover-cleanup/ADR.md` | `de9fffdf63fc0ec61a98bdde7327617682d7b336b1199711537d48b26ee04721` | unchanged |
| ADR-008 `2026-10-31-jev-merge-association/ADR.md` | `76dbd3aec2503d447270a91bdfb9a9c44e35ca730c45199add44a34847c45ab0` | unchanged |

**The digest changes match round 9's "Required with the next submission" list exactly.** The html, proposal, map, JSON and audit changed. `ADR.md` and the 12 other files are byte-identical to `round-9.md`'s table.

## Scope

- **Under review:** worktree `/Users/adamjackson/Projects/work/scout-req`, branch `agent/requirements-path`, with all changes staged. The review covered:
  - the "Round 9 fixes" (`proposal.md:140-148`);
  - the near-verbatim §11/§13 carry in REQ-167…REQ-182, and REQ-89/REQ-91;
  - the Overview wording on who publishes;
  - the map and JSON union changes;
  - the corrected audit and its new §6/§11/§13 inventory (`completeness-audit.md:176+`);
  - a final broad pass over proposal questions 1–5.
- **Base:** `main` at `9a3d19f5` in `/Users/adamjackson/Projects/work/scout/`. `D:n` is line n of the base `docs/scout-design.md`.
- **Checks made:**
  - Every rewritten REQ, clause by clause, against `D:173-177,184,377-402,435-446`.
  - The ADR columns against the frozen diffs: rescope `proposed.diff:110,112,254-256,265-266`; decisions `proposed.diff:142,146,155,163`; area-refresh-fallback `proposed.diff:100-101`; cutover-cleanup `proposed.diff:83`; e1b-v2 `proposed.diff:142`.
  - Map rows `:1-60,163-184,268-340` against JSON keys.
  - Overview `html:170-505`.
  - Live-tree greps for deleted paths, the ledgers, and `AGENTS.md` experiment rules 1–7.
- **Not reopened:** decisions settled in ADR-000…ADR-008, and earlier "Considered and not raised" items unless there is new evidence.

## Verification of round-9 fixes

| # | Evidence (worktree vs base) | Result |
|---|---|---|
| R9-01 | **X1's preregistration list matches the base word for word.** REQ-169 (`html:686`) says "…treatment of capped and failed units, versioning of verdicts and test-set reuse, and sample size…", which equals `D:435`.<br>**The keying clause is restored.** REQ-89 (`:571`) says "how Jev-quality statistics are keyed to them is fixed by X1's preregistration", which equals `D:175`. It was introduced by rescope `proposed.diff:110`, and ADR-002 is cited. | fixed |
| R9-02 | REQ-171 (`:688`) restores:<br>• "per file (path proximity, import edges)";<br>• "split, … interval";<br>• "(files assigned the correct feature ÷ files assigned any feature)".<br>All three equal `D:438`. | fixed |
| R9-03 | REQ-176 (`:693`) says "labels follow the agent protocol (REQ-100/REQ-101)", which carries `D:444`. A different cross-reference error remains in the same row: R10-03. | fixed |
| R9-04 | REQ-171, REQ-173 and REQ-175 say "(rev 2.1: 200 labels)". Each has a confirm-or-replace sentence (`D:438,440,443`). | fixed |
| R9-05 | **REQ-179** (`:716`) equals `D:381-382`. It includes consent off, the rule-digest bump, the simulated 14 days/5 PR runs, paired merge-base attribution with whole-run and per-family `unknown`, and "the same path applies before a baseline is published".<br>**REQ-180** (`:717`) equals `D:385-386`, including "one run and not the next".<br>**REQ-167** (`:713`) equals `D:393-394`, including "in quick succession (new unowned package, new manifest, content-only)", apart from R10-06. | fixed |
| R9-06 | REQ-26 (`:611`) has no 90-day duty. Audit candidate 29 (`audit:24`) is `OVERREACH` and cites `round-1.md:255` and `scout-guardian.yml:118`. | fixed |
| R9-07 | **The four named parents are fixed in both the map and the JSON:** §3.2 (`map:16`), §8.5 (`:35`), Appendix 22 (`:60`) and AL §1 (`:270`) (peer, with JSON lines).<br>**The union rule still fails in three places:**<br>• the `req §n` ← `req REQ-n` pairs: R10-08;<br>• parents whose children are `dropped` rows that name REQs: R10-09;<br>• the JSON `"Evidence notation"` key: R10-10. | partly fixed |
| R9-08 | **"Default-branch runs (push, nightly or dispatch)" now appears in all three places:** `html:198`, `:308` ("trusted default-branch CAS") and `:366`.<br>**This matches Appendix A and the workflow:** REQ-25 (`:696`), REQ-26 (`:611`) and `scout-guardian-trusted.yml:60-62,193-197`. | fixed |
| R9-09 | **(a)** Candidate 25 now credits AL §4 / ADR-000 (`audit:20`).<br>**(b)** Candidate 28 no longer says "contradictory" (`:23`).<br>**(c)** The note on candidate numbers is at `:5`.<br>**(d)** "Seven audit slices" appears in the audit and the proposal.<br>**(e)** `proposal.md:138` records the `ADR.md` edit.<br>**The totals add up.** The peer recounted from the rows: 72 = 21 lost + 19 loosened + 1 overreach + 27 already carried + 1 not normative + 3 ADR or dropped. That is 64 distinct after 8 duplicates. | fixed |
| Reverse fidelity (§11/§13) | **No new obligations were added.** No run result, date or evidence pointer became a requirement in REQ-167…182:<br>• X2/X2b run statistics, X5's G4 prior (still in REQ-124 `:644` with ADR-001), the R-040/R-047 pointers, "(reference: `build_panel.active_hours`)" and "(IR §5)" were dropped correctly.<br>• REQ-170's "A new mapper default requires a new council-signed study on X2b's protocol" is base text from R15 (`D:424`), not new.<br>• REQ-176's "carried policy thresholds" resolve to REQ-146 (`:660`; `t_choice` 0.535, covered 0.60 = `D:463` A14). | faithful |
| ADR columns of rewritten rows | **Correct for most rows:**<br>• REQ-169, REQ-171, REQ-172, REQ-173 and REQ-175 cite ADR-002 (rescope `:295`ff).<br>• REQ-179 cites ADR-003 (decisions `:142`).<br>• REQ-180 cites ADR-001 and ADR-003 (e1b-v2 `:142`, decisions `:146`).<br>• REQ-181 and REQ-182 cite ADR-002 (rescope `:254-256,265-266`).<br>• REQ-167 cites ADR-003 and ADR-006 (decisions `:155`, area-refresh-fallback `:100-101`).<br>• REQ-89 and REQ-91 cite ADR-002 (rescope `:110,112`).<br>**Two rows are wrong:** REQ-178 lacks ADR-007 (R10-01), and REQ-168 lacks ADR-009 (R10-07). | correct except R10-01, R10-07 |
| Clause inventory (§6, §11, §13) | **It covers every base clause, but some carrier claims are false.** It lists the claim rows without checking that the quoted carrier contains the clause. Two losses (R10-04, R10-05) and five false carrier claims (R10-11) came from this. | inaccurate in places |
| Overview vs round-9 REQs | **Consistent:**<br>• artifacts (`:179`) with REQ-26;<br>• Rank (`:265-267`) with REQ-89, REQ-91 and REQ-169;<br>• onboarding and manual refresh (`:323-324,353`) with REQ-167;<br>• the activation callout (`:449`: 14 days, 5 runs, 10 %) with REQ-179;<br>• pending areas (`:353`) with REQ-176 and REQ-177. | consistent |
| Conditions from round 9 | The main agent provided the digests, `readlink`, `git diff --cached --stat`, the 312-key count and the parent check output. | met (main-agent outputs accepted) |

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R10-01 | minor | `docs/scout-design.html:715` (REQ-178, ADR column) | **REQ-178 lacks ADR-007, which introduced the migration clauses it now carries.** Round 9 carried `D:378` near-verbatim into REQ-178:<br>• the genuine-v1 DDL fixture (`schemaV1`, `inventory_snapshots`, `schemaV3` at `41ea1eb`), "not a current store rewound with `PRAGMA user_version`";<br>• the refusal of versions 1, 2 and 4–11 on a read-only connection before the writable open, with no `writer.lock`, `artifacts/` or journal-mode change;<br>• the ordered schema-equality check.<br>ADR-007 introduced all of these (cutover-cleanup `proposed.diff:82→83`; ADR-007 card `html:904-905,917`). REQ-178's column lists only ADR-000, ADR-003 and ADR-008, and the proposal says "no other ADR column changes" (`proposal.md:144`). Under the round-5 convention (`round-5.md:110`), a missing changing ADR is raised. | Add ADR-007 to REQ-178's ADR column, and record the change in `proposal.md`. |
| R10-02 | minor | `html:715` (REQ-178); `completeness-audit.md:308,313` | **REQ-178 has two cross-references to a section that no longer exists.** It says "Scenario: **§10.1** benchmark and the pinned fastapi assess" and "Acceptance: **§10.1** criteria". The new document has no §10.1, and no other Appendix A row uses a bare `§n` for the design (grep `§\d+\.\d` on the html). These references define which benchmark is run and what passes. Their content now lives in REQ-165 (`:711`: the 5,000-job / 5 MB benchmark) and REQ-38 (the pinned fastapi assess < 10 min). Round 9's method changed other section references to REQ references (audit `:180`), but missed these two. | Replace both with REQ-165 and REQ-38, for example "Scenario: the REQ-165 benchmark and the REQ-38 pinned fastapi assess…; Acceptance: REQ-165 and REQ-38 criteria". Update audit `:308,313`. |
| R10-03 | minor | `html:693` (REQ-176); `completeness-audit.md:487` | **REQ-176 points to the wrong REQ for the guard defaults.** It says "the carried policy thresholds and **REQ-155** guard defaults as starting points". REQ-155 (`:669`) covers staleness rules. The four guard defaults (20, 20, 0.40, 0.20) are in **REQ-159** (`:673`). `D:444` says "the §4 guard defaults". A preregistration that follows this reference would find no defaults. | Change it to "REQ-159 guard defaults" (and optionally "REQ-146 thresholds"), and correct audit `:487`. ADR column unchanged. |
| R10-04 | minor | `html:570` (REQ-88); `completeness-audit.md:223` | **REQ-88 says the state facts are computed, but not that they are sent.** `D:175` says "Path role, test, fixture and generated status are computed in code **and supplied as facts in state**". REQ-88 says only "compute path role/test/fixture/generated status in code". No Appendix A row says the facts are sent in state (grep "as facts", "facts in state", "supplied as": 0 hits). **[INFERENCE]** Without it, the code could compute these facts and still let Jev infer them. The audit row `:223` claims the clause is carried. Introduced by rescope `proposed.diff:110` (ADR-002, already cited). | Restore "and supply them as facts in state" in REQ-88, and correct audit `:223`. |
| R10-05 | minor | `html:573` (REQ-91); `completeness-audit.md:241-242` | **REQ-91 has lost the condition that admits commit and unowned-file units to the cap order.** `D:177` says "then the other question families (rubric functions; commit and unowned-file units **once X5/X3 admit them**) in a stable unit-key order". REQ-91 says only "then other families in stable unit-key order", and no row carries the admission condition (grep "X5/X3", "admit": 0 hits). REQ-124 gates the `fix_touches` measure and REQ-171 gates its use as a reconciliation aid. Neither says those units may not take call budget before X5/X3. Round 9 said this row restored `D:177` (`proposal.md:144`). Rescope `proposed.diff:112` (ADR-002, cited). | Restore "(rubric functions; commit and unowned-file units once X5/X3 admit them)" in REQ-91, and correct audit `:241-242`. |
| R10-06 | nit | `html:713` (REQ-167); `completeness-audit.md:378` | **REQ-167 no longer says where the Phase 4 report is recorded.** `D:394` says "measured by the scenario harness **and recorded in `docs/STATUS.md` §1**". REQ-167 drops the destination. This is not an evidence pointer: ADR-003 chose it as round-2 option (a), because `scout map` discards the reconciliation report (decisions `PROPOSAL.md:14` D6; `proposed.diff:155`). It also still applies: `docs/STATUS.md:70-84` holds the pre-2.6 run, and the 2.6 re-run is pending. Audit `:378` quotes the base clause in full, including STATUS §1, and claims REQ-167 carries it. | Restore "and recorded in `docs/STATUS.md` §1" in REQ-167. ADR column unchanged (ADR-003). |
| R10-07 | nit | `html:714` (REQ-168, ADR column); `requirements-map.md:181-183` and `:43`; JSON `"§11 Phase 6"`, `"§11 Phase 6 Scenario"`, `"§11 Phase 6 Acceptance"` (`design-section-map.json:1359-1381`) and `"§11"` | **The Phase 6 rows do not record ADR-009, which removed two of their clauses.** ADR-009 removed "docs replace the archived spec" and "`SCOUT_REBUILD_REQUIREMENTS.md` archived" from `D:401-402` (`ADR.md:11`; audit `:416` assigns both to ADR-009). The removal itself is authorised. But:<br>• REQ-168's column (ADR-000, 002, 003, 007) does not cite ADR-009, although REQ-37 and REQ-42 cite ADR-009 for their changes.<br>• The three Phase 6 map rows and their JSON keys list only REQs, although map `:3` says each source row records its disposition.<br>• The `§11` parent is not a union with ADR-009 either (no match for `Design §11 —.*ADR-009`).<br>So a frozen citation of "§11 Phase 6" does not lead to the decision that changed it. | Add ADR-009 to REQ-168's ADR column, to map rows `:181-183` and `:43`, and to the four JSON keys. Disclose the change in `proposal.md`. Key count unchanged. |
| R10-08 | minor | `requirements-map.md:289-296`; JSON `"req §2"`…`"req §9"` (`design-section-map.json:1598-1651`); map `:3`; `proposal.md:146` | **The old requirements sections are not unions of their own requirements.** In `SCOUT_REBUILD_REQUIREMENTS.md`, old REQ-5…11 sit under §2, REQ-12…15 under §3, and so on. The audit groups them the same way (`audit:119-121`). Yet each `req §n` row lists only the old IDs and leaves out the split-off REQs of its `req REQ-n` children (`map:298-338`):<br>• §2 lacks REQ-44, 46 and 47;<br>• §3 lacks REQ-52…57;<br>• §4 lacks REQ-62…69 and REQ-84;<br>• §5 lacks REQ-48…51, REQ-70…84 and REQ-161…163;<br>• §6 lacks REQ-58…61, REQ-86…102, REQ-139/140 and REQ-147…150;<br>• §7 lacks REQ-103…129 and REQ-132…134;<br>• §8 lacks REQ-135;<br>• §9 lacks REQ-57, REQ-136…138, REQ-160, REQ-164 and REQ-165.<br>**A real citation breaks.** Base `D:110` cites "(req §3)" for "never a second mapper". The JSON returns REQ-12…15, but the clause is in REQ-53 (`html:546`). `adrs/README.md:29` makes the JSON the resolver for `req §n` citations.<br>The proposal (`:146`) and map `:3` both say every parent is a union. **[INFERENCE]** The script paired keys only by name prefix, so `req §n` / `req REQ-n` was never compared. | Add each `req REQ-n` row's IDs to its `req §n` parent in the map, regenerate those JSON keys, extend the check to these pairs, and correct `proposal.md:146`. Key count stays 312. |
| R10-09 | minor | `requirements-map.md:273,275,281,286,295,340`; JSON `"AL §3"` (`:451-456`), `"AL §5"` (`:1576-1580`) | **Parents leave out the REQs that their `dropped` children name, including REQs that carry the parent's own text.**<br>• **AL §3** (REQ-21, 22, 23, ADR-003) lacks REQ-78, REQ-48 and ADR-000, which child `:281` names. AL §3's own text has the activation record `blocking: [{id, version, activated_at, evidence…}]` and `signals/would_block.jsonl` (base `scout-allowlist.md:44-47`). REQ-78 (`html:600`) and REQ-48 carry these, and audit A4 (`:93`) says so.<br>• **AL §5** (REQ-22, 70, ADR-000) lacks REQ-23, which child `:286` names ("decided: maintainer-only REQ-23"). It also lacks REQ-81 and REQ-183, which audit `:118` says carry its inline assumptions.<br>• **req §8**: the JSON includes ADR-002, from dropped child `:340`, but map row `:295` does not, so the map and the JSON follow different rules. | **AL §3:** add REQ-78, REQ-48 and ADR-000.<br>**AL §5:** add REQ-23, REQ-81 and REQ-183.<br>**req §8:** add ADR-002 to map row `:295`.<br>Make the same changes in the JSON. State in map `:3` that the REQ and ADR targets of `dropped` children enter the union. |
| R10-10 | nit | JSON `"Evidence notation"` (`design-section-map.json:1929-1932`); map `:58-59` | **The JSON key has fewer IDs than its map rows.** The key is [ADR-009, REQ-43]. Map `:58` ("aliases and study admissibility") lists ADR-009, ADR-001, ADR-002 and ADR-008. So citations of "evidence notation" do not reach the E1b and X2d admissibility decisions. The research ledger cites "evidence notation" in R-042, R-044 and R-047 (peer). The proposal (`:146`) says every key contains its map rows' IDs. | Set the key to ADR-009, ADR-001, ADR-002, ADR-008 and REQ-43, and include multi-row keys in the check. |
| R10-11 | nit | `completeness-audit.md:215,233,280,322,378`; `proposal.md:146` | **Some inventory rows name the wrong carrier.** The inventory claims to quote "its Appendix A carrier" (`:180`), but these quotes do not contain the clause:<br>• `D:173.10` (the catalogue-only version change is not stale): the clause is in REQ-155 (`html:669`), which is not listed.<br>• `D:176.3` (association decisions are `associations` rows): the row quotes REQ-28, but the clause is in REQ-154 (`:668`).<br>• `D:184.9` (launch `provisional` until X1 reports): the row quotes REQ-93's lane-health text, but the clause is in REQ-97 (`:579`).<br>• `D:378.10` ("R-047)"): an evidence pointer, quoted against an unrelated REQ-178 phrase.<br>• `D:394.3`: see R10-06.<br>Also, `:223` and `:241-242` claimed clauses that are lost (R10-04 and R10-05), and the proposal's R9-07 claim is false (R10-08 to R10-10). | Correct these rows to the REQs that actually carry them. Mark `D:378.10` as an omitted evidence pointer. After the R10-02…R10-06 fixes, re-check every inventory quote against the clause it claims. Correct `proposal.md:146`. |

## Answers to the proposal's questions

1. **Is normative content lost, weakened or changed?**
   - **Yes:**
     - three qualifier losses: R10-04 (§6 state facts), R10-05 (§6 cap admission) and R10-06 (§11 Phase 4 record);
     - two wrong references: R10-02 (dangling §10.1) and R10-03 (REQ-155 instead of REQ-159).
   - **Preserved:** every §13 decision rule and every number (≥ 10 points, LB > 0, 0.9 with its denominator, LB ≥ 0.6, 0.85/0.8 with 0.75/0.7, last 200 PRs × 3 repositories, 200 + 200, Wilson LB ≥ 0.9, 0.8 and no prevalence claim, ≥ 0.80 with lower bounds), and every Phase 1–6 frame. No run result became a requirement.
2. **Do the ADR summaries and cards match their records?**
   - **Yes:** the ADR-000…008 digests are unchanged. The peer checked the ADR-008 and ADR-009 cards against their `ADR.md` files: 0.60/0.535, 0.922 [0.840, 0.964] 71/77, 0.622 [0.514, 0.719] 51/82, and "final round 11, 14 → 0".
   - **Gaps:** REQ ADR columns, not the cards: R10-01 and R10-07.
3. **Does the Overview agree with Appendix A?** Yes. R9-08 is fixed, and nothing contradicts the REQs changed in round 9.
4. **Are rule, tooling and citation changes complete, with no live reference to a deleted file?**
   - **Deleted files:** no live reference. The peer grepped `internal/`, `cmd/`, `.github/`, `scripts/`, `toolchain/`, `.omp/`, `.agents/`, `AGENTS.md`, `README.md`, `CONTRIBUTING.md` and the non-frozen `docs/`. The only hits are ledger rows, labelled history, the README path maps, the plan's deletion list and REQ-110's citation pinned at `9a3d19f5`.
   - **Citation index:** incomplete: R10-07, R10-08, R10-09 and R10-10.
5. **Experiment rules, ledgers, frozen records:** no violation.
   - `AGENTS.md:10-31` equals base rules 1–7, apart from the settled X1–X6, X9–X11 preamble.
   - The research ledger only appends R-062.
   - The hypothesis ledger only appends one History line.
   - The frozen digests are unchanged.

## Considered and not raised

- **The Phase 6 cutover clauses missing from REQ-168.** ADR-009 authorises the removal (`ADR.md:11`; audit `:416`), and the challenger confirmed it. Only the missing ADR-009 citation is raised (R10-07).
- **REQ-170 "E3 protocol with Scout's area reconciliation" against `D:436` "with reconciliation".** Settled by R8-06.
- **REQ-176 "(carried policy thresholds)" instead of "R-057 values".** The values are in REQ-146 and A14. Only the wrong guard-default reference is raised (R10-03).
- **REQ-91 leaves out the cap defaults ($0.50, 2,000 calls).** REQ-29 (`html:566`) carries them with their `[ASSUMPTION]` markers.
- **REQ-90 leaves out the catalogue-only exception.** REQ-155 carries it. This is an inventory error only (R10-11).
- **Run-flow SVG (`:217-221`) and trigger list (`:323-324`) say "push … publishes".** They mirror `D:196` and REQ-160 and do not deny nightly publishing. `:198` now says nightly and dispatch runs publish (peer).
- **`§7.1`, `§7.4`, `§10.1` and `§10.2` map to the whole REQ-160…168 block.** This over-includes, which makes citations resolve to more IDs, not fewer. It predates round 9.
- **JSON `"§14 A10"` → REQ-38.** Settled (`round-8.md:118`).
- **Extra ADRs in a REQ's column (for example, ADR-007 on REQ-167 and REQ-168).** Only missing changing ADRs are raised (`round-5.md:110`).
- **"Accepted" status on ADR-009 before sign-off.** This is the adoption-commit convention, accepted since R1-07.

## Peers

- **`innovation-challenger`:**
  - Verified R9-01…R9-05 and the reverse-fidelity check across REQ-42, REQ-89…91, REQ-98 and REQ-167…182.
  - Found R10-02, R10-03, R10-04, R10-05 and most of R10-11, and found R10-01 independently.
  - Judged the Phase 6 removal authorised. The lead agrees, and raises only the citation gap (R10-07).
- **`experiment-peer`:**
  - Verified R9-06, R9-08 and R9-09, the audit totals, the cards, the ledgers, the rules and the absence of deleted-path references.
  - Swept every parent/child group and found R10-08, R10-09 and R10-10.
- **Lead:**
  - Found R10-01 (from cutover-cleanup `proposed.diff:83`), R10-06 (from decisions `PROPOSAL.md:14` D6, which makes STATUS §1 a decision ADR-003 adopted) and R10-07.
  - Re-checked every peer finding against `D:175,177,377-378,394,444`, `html:546,570,573,600,669,673,693,713-715`, map `:43,58,181-183,268-340`, JSON `:1607,1929` and base `scout-allowlist.md:40-62`.
- **Where the peers changed the outcome:** working alone, the lead would have raised only R10-01, R10-06 and R10-07, and would have judged R9-07 fully fixed. The challenger added the §6 qualifier losses and the cross-reference errors. The peer showed that the union check never covered the `req §n` pairs, `dropped` targets or keys that span several map rows.

## Verdict

**11 issues (R10-01, R10-02, R10-03, R10-04, R10-05, R10-08 and R10-09 minor; R10-06, R10-07, R10-10 and R10-11 nits): NOT signed off.**

No sign-off applies to the digests above.

### Required with the next submission

- **Changed digests:** expected only for:
  - `docs/scout-design.html` (R10-01 to R10-07);
  - `requirements-map.md` and `design-section-map.json` (R10-07 to R10-10);
  - `completeness-audit.md` (R10-02 to R10-06 and R10-11);
  - `proposal.md` (a "Round 10 fixes" section and the corrected R9-07 claim).

  Any other changed digest must be disclosed in `proposal.md`.
- **Recomputed digests and checks:** the main agent must recompute every digest with `shasum -a 256` on the staged worktree. It must also re-run:
  - `test -L CLAUDE.md && readlink CLAUDE.md`, expecting `AGENTS.md`;
  - `git -C /Users/adamjackson/Projects/work/scout-req diff --cached --stat`.

  It must confirm that the JSON still has 312 keys, and report the output of the extended union check. That check must cover:
  - (a) every design, Appendix and AL parent/child pair;
  - (b) every `req §n` ← `req REQ-n` pair;
  - (c) the REQ and ADR targets named by `dropped` children;
  - (d) each JSON key against the union of every map row it represents.

  It should also report a grep showing no bare `§n.n` design reference in Appendix A. `go build`/`go vet` need not be re-run unless Go files change.

---

## Council memo

1. **Recommendation:** fix R10-01…R10-11 in one pass and submit round 11. Make every mechanical check a script whose output is submitted, so it is not left to review:
   - the extended union check above;
   - a cross-reference check: every `REQ-n` mentioned inside a row exists and is on topic, and no design `§` reference remains;
   - an ADR-column check for the rewritten rows, done by finding each carried clause's first `+` line in the frozen `proposed.diff` files;
   - an inventory check: each quoted carrier contains the clause's key terms.

   Rounds 7–10 each found defects in exactly the places a prose review and a partial script left unchecked.
2. **Alternatives:**
   - (a) **Sign off with conditions.** Rejected: design rule 1 requires zero issues.
   - (b) **Drop the `req §n` and AL union claims, and narrow map `:3` to design sections.** This is cheaper. But frozen records cite `req §n` (for example `D:110`, `D:200`), and README `:29` makes the JSON their resolver, so narrowing the claim would keep those citations resolving incompletely. Rejected.
   - (c) **Treat inventory errors as non-blocking.** Rejected: the inventory is a signed record of carriage, and two of its false claims hid real losses (R10-04, R10-05).
3. **Evidence:** each issue row cites the exact worktree and base `file:line`, and the verification table cites every frozen diff line used.
4. **Risks:**
   - **R10-01 and R10-07:** ADR traceability gaps for migration rules and Phase 6 changes.
   - **R10-02 and R10-03:** a phase or preregistration that follows the references finds nothing.
   - **R10-04 and R10-05:** Jev might infer path role itself, or spend budget on units that are not yet admitted.
   - **R10-06:** the Phase 4 re-run report has no designated record.
   - **R10-08 to R10-10:** frozen `req §n`, AL §3/§5 and "evidence notation" citations resolve incompletely.
   - **Reversibility:** every fix is a text or JSON edit and is fully reversible.
   - **Unknowns:** the council saw no per-file staged diff and relies on the main agent's digests and statements. The "19 parents extended" claim was not verified.
5. **Execution outline:**
   1. **REQ-178:** add ADR-007, and replace "§10.1" with REQ-165/REQ-38.
   2. **REQ-176:** REQ-155 → REQ-159.
      **REQ-88:** add "and supply them as facts in state".
      **REQ-91:** restore "(rubric functions; commit and unowned-file units once X5/X3 admit them)".
      **REQ-167:** restore "and recorded in `docs/STATUS.md` §1".
   3. **REQ-168:** add ADR-009. Add ADR-009 to map `:43,181-183` and the four JSON keys.
   4. **Map and JSON:**
      - make the `req §2`…`§9` parents unions of their `req REQ-n` children;
      - AL §3: add REQ-78, REQ-48 and ADR-000;
      - AL §5: add REQ-23, REQ-81 and REQ-183;
      - req §8: add ADR-002 to the map row;
      - `"Evidence notation"`: add ADR-001, ADR-002 and ADR-008;
      - state the rule for `dropped` children in map `:3`.
   5. **Audit:** correct inventory rows `:215,223,233,241-242,280,308,313,322,378,487`, and re-run the quote check.
   6. **Proposal:** add "Round 10 fixes", including the ADR-column changes, and correct `:146`.
   7. **Checks:** run the extended union, cross-reference, ADR-trace and inventory scripts. Recompute every digest with `shasum -a 256`, run `readlink` and `git diff --cached --stat`, confirm 312 keys, and submit round 11 with the outputs.
