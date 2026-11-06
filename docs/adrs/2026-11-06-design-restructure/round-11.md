# Round 11: ADR-009, one design document with requirements and ADR appendices

Date: 2026-11-06. This round follows design rules 1–3 in `AGENTS.md`/`CLAUDE.md`. The innovation council lead ran it with two independent peers. Neither peer was told the lead's view.
- **`innovation-challenger` (GPT):** checked the content of R10-01…R10-06 against the base. It checked the html changes the proposal claims and spot-checked 20 rows of the new ADR-column check against the frozen `proposed.diff` and `scout-design.base.md` lines. It also checked 28 internal REQ→REQ references, the remaining `§` symbols in Appendix A, and proposal questions 1 and 3.
- **`experiment-peer` (Opus):** checked R10-07…R10-11. It recomputed every `req §n` union from base `SCOUT_REBUILD_REQUIREMENTS.md`, and checked all 22 `dropped` rows and every JSON key that stands for more than one map row. It swept every parent/child group, recounted the map, the JSON, the inventory and the audit totals, checked the proposal's "Round 10 fixes", and answered questions 2, 4 and 5.

The review was read-only. It made no model calls outside the council and did not touch `.codegraph`.

**The main agent recomputed the digests; the council did not.** The lead and both peers had only read-only tools and no shell, so nobody on the council hashed a file. Earlier rounds accepted a main-agent recomputation, and this round does too. The main agent reported these outputs on the staged worktree:
- `shasum -a 256` on each file below;
- `test -L CLAUDE.md && readlink CLAUDE.md` → `AGENTS.md`;
- `git diff --cached --stat` → `250 files changed, 7675 insertions(+), 1781 deletions(-)`. That is round 10's 249 files plus the new `round-10.md`.

No Go file has changed since round 6.

| File | SHA-256 (main agent, `shasum -a 256`) | vs round 10 |
|---|---|---|
| `docs/scout-design.html` | `f0aa226a88c3c8eda782a7327be85b2bccb5a879df177673d12021d50ffdc463` | changed (R10-01…R10-07, ADR columns) |
| `docs/adrs/2026-11-06-design-restructure/proposal.md` | `1a62be0b05cc63e64622cff44c44a6d503dd132923013f26d30ec47795ecac34` | changed ("Round 10 fixes") |
| `…/requirements-map.md` | `c24a6b6c40a3775c79cfd16d55371745ef646f536e085273289da661352a1bc7` | changed (R10-07…R10-10) |
| `…/design-section-map.json` | `7efd0bf693ae2396c27d580dd3947ecbcaa16fad8bc596bb6bf4fcc65d106851` | changed (R10-07…R10-10) |
| `…/ADR.md` (ADR-009) | `1393b6eb7488688c24ef25106ec6861616ae61e75675570651e5b97ab1fae83a` | unchanged |
| `…/completeness-audit.md` | `0fa44fa63e084d2ea9b0690fc5bd159037cb526980a736e99da621ea9505c31b` | changed (R10-02…R10-06, R10-11, ADR-column check) |
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

**The digest changes match round 10's "Required with the next submission" list exactly.** The html, proposal, map, JSON and audit changed, and so did the new `round-10.md`. `ADR.md` and the 12 other files are byte-identical to `round-10.md`'s table.

## Scope

- **Under review:** worktree `/Users/adamjackson/Projects/work/scout-req`, branch `agent/requirements-path`, with all changes staged. The review covered:
  - the "Round 10 fixes" (`proposal.md:150-168`);
  - the restored REQ-88, REQ-91 and REQ-167 qualifiers;
  - the REQ-176 and REQ-178 references;
  - the 14 added ADR citations and their trace (`completeness-audit.md:521-598`);
  - the extended union rule (`requirements-map.md:3`) and the map and JSON changes;
  - the re-checked inventory (`completeness-audit.md:178-519`);
  - a final broad pass over proposal questions 1–5.
- **Base:** `main` at `9a3d19f5` in `/Users/adamjackson/Projects/work/scout/`. `D:n` is line n of the base `docs/scout-design.md`.
- **Checks made:**
  - **Rows against the base:** `html:542,548-549,569-573,612,616,622,666,673,693,696,701,703,706,711,713-715` against `D:170,173-177,393-394,444,497`.
  - **Frozen sources:** rescope `proposed.diff:110,112,140-145,214-218`; codex-luna-mapping `proposed.diff:30-40`; e1b-v2 `scout-design.base.md:89-95,144,158-175,238`; area-refresh-fallback `ADR.md:34-35`.
  - **Map and JSON:** map `:1-75,253-388` against JSON `:172,452,499,700,706,1012,1363-1381,1583,1647,1749,2044`.
  - **Base requirements:** section membership in `SCOUT_REBUILD_REQUIREMENTS.md:7-79`.
- **Not reopened:** decisions settled in ADR-000…ADR-008, and earlier "Considered and not raised" items unless there is new evidence.

## Verification of round-10 fixes

| # | Evidence (worktree vs base) | Result |
|---|---|---|
| R10-01 | REQ-178's ADR column now reads "ADR-000, ADR-002, ADR-003, ADR-007, ADR-008" (`html:715`). ADR-007 introduced the genuine-v1 DDL fixture, the read-only refusal and the ordered schema check (cutover-cleanup `proposed.diff:83`). | fixed |
| R10-02 | **Both references are replaced.** REQ-178 says "the REQ-165 benchmark and the REQ-38 pinned fastapi assess" and "Acceptance: REQ-165 and REQ-38 criteria" (`html:715`). REQ-165 (`:711`) holds the 5,000-job / 5 MB benchmark, and REQ-38 (`:703`) holds the < 10 min fastapi bar. Audit `:310,315` is corrected.<br>**No bare design `§` remains in Appendix A** (`html:526-720`, grep). There are exactly five `§` symbols, and each names another document:<br>• `:630`: `docs/inertia-research.md` §5 at `9a3d19f5`;<br>• `:631`: E1b preregistration §3.1;<br>• `:637` (twice): E1b preregistration §3.3;<br>• `:713`: `docs/STATUS.md` §1.<br>This matches `proposal.md:168`. | fixed |
| R10-03 | REQ-176 says "the carried REQ-146 policy thresholds and REQ-159 guard defaults" (`html:693`). REQ-159 (`:673`) holds 20, 20, 0.40, 0.20 in the named parameter order, and REQ-146 holds 0.535/0.60. Together they carry `D:444` "R-057 values and the §4 guard defaults". Audit `:489` is corrected. | fixed |
| R10-04 | REQ-88 says "compute path role/test/fixture/generated status in code and supply them as facts in state" (`html:570`), which equals `D:175`. Audit `:225` is corrected. | fixed |
| R10-05 | REQ-91 says "then the other question families (rubric functions; commit and unowned-file units once X5/X3 admit them) in stable unit-key order" (`html:573`), which equals `D:177`. Audit `:242-246` is corrected. | fixed |
| R10-06 | REQ-167 says "(measured by the scenario harness and recorded in `docs/STATUS.md` §1; `scout map` does not print it)" (`html:713`), which equals `D:394`. Audit `:380-381` is corrected. | fixed |
| R10-07 | **ADR-009 now appears in every place the round asked for:**<br>• REQ-168's ADR column (`html:714`);<br>• map `:43` (`§11`) and `:181-183` (Phase 6, Scenario, Acceptance);<br>• JSON `"§11"` (`:172`) and the three `"§11 Phase 6…"` keys (`:1363-1381`). | fixed |
| R10-08 | **Every `req §n` row is now the union of its children.** The lead and the peer each recomputed it, using the base section membership (`SCOUT_REBUILD_REQUIREMENTS.md:7,14,24,31,38,49,57,64,68,75`) and the `req REQ-n` rows `map:298-338`:<br>• `req §2` (`map:289`) = REQ-5…11, 44, 46, 47;<br>• `req §3` (`:290`) adds REQ-52…57;<br>• `req §5` (`:292`) = REQ-20…27, 48…51, 70…84, 161…163;<br>• `req §6`, `req §7` and `req §9` (`:293`, `:294`, `:296`) are complete.<br>The JSON `"req §n"` keys equal the map rows. `"req §3"` now resolves the `D:110` citation to REQ-53. The count stated in the proposal is wrong: R11-02. | fixed (count: R11-02) |
| R10-09 | **The named parents now hold their dropped children's targets:**<br>• AL §3 (`map:273`, JSON `:452`) adds REQ-78, REQ-48 and ADR-000;<br>• AL §5 (`map:275`, JSON `:1583`) adds REQ-23, REQ-81 and REQ-183;<br>• `req §8` (`map:295`) adds ADR-002, matching JSON `:1749`.<br>Map `:3` now states the dropped-children rule.<br>**Two heading rows do not follow the rule as now stated:** R11-03. | fixed (wording: R11-03) |
| R10-10 | `"Evidence notation"` (JSON `:2044`) = ADR-009, REQ-43, ADR-001, ADR-002, ADR-008, the union of `map:58-59`.<br>The peer swept the other keys that stand for several map rows: `§2`, `§6 Rejected`, `§8.5 E1b`, `§10.1`…`§10.3`, `§4`, `§12`, `§14`, `§15`, `§14 A8`/`A10` (`map:383`) and `Appendix 6`/`12` (`map:364`). All pass, except **`"Appendix 20"`**: R11-01. | fixed (except R11-01) |
| R10-11 | **The named carriers are corrected:**<br>• `D:173.10` → REQ-155 (`audit:217`);<br>• `D:176.3` → REQ-154 (`:235`);<br>• `D:184.9` → REQ-97 (`:282`);<br>• `D:378.10` is marked as an omitted evidence pointer (`:324`);<br>• `D:394.3` → REQ-167 with STATUS §1 (`:380`).<br>**Sample of other rows:** the peer grepped the html for the quoted phrases of more than 25 other rows (§6, §11 and every X row of §13). Every quote is present with its key terms.<br>**Reclassification:** `D:170.5` is now a tooling observation (`audit:204`). Its normative part, "no logged-in CLI → `skipped (no model CLI)`", stays with REQ-61 (`html:701`). | fixed |
| ADR-column check (`audit:521-598`) | **Counts:** 13 rows changed and 14 citations added (11 × ADR-000, 2 × ADR-002, 1 × ADR-005). 13 + 39 = 52. All 14 citations appear in the html (`:542,569,570,612,616,622,666,696,701,713,715`).<br>**ADR-000 sources** (lead and challenger), lines of e1b-v2 `scout-design.base.md`:<br>• `:90`: both mappers at onboarding, never on PRs, 900000 timeout;<br>• `:91`: matched areas update file set and provenance;<br>• `:93`: latest-refresh ownership;<br>• `:95`: zero-feature failure semantics;<br>• `:144`: resolved-model pinning and the contextual sanitizer;<br>• `:160`: trusted publish, paired recompute, AWS/OIDC dropped;<br>• `:164`: identical digests;<br>• `:175`: pinned release, base-branch policy, adoption sequence;<br>• `:238`: the 5 % MSE / CI > 0 rule and no significance upgrade.<br>**Other sources:**<br>• ADR-002: rescope `:143` (trusted Jev-only secret) and `:217` (decided Mac < 10 min, which REQ-178 now requires through REQ-38);<br>• ADR-005: codex-luna `:36` (codex-only mapper; REQ-167 requires `codex` on `PATH`).<br>**Sampled "no change" rows:** the challenger traced 20 of them, including REQ-89…92, 96, 101 and 171…182, to their `+` lines, and all pass. The lead checked that REQ-160's `refresh-map` dispatch is pre-rescope (rescope `proposed.diff:133`, ADR-000). | correct |
| Overview vs Appendix A | The challenger and lead found no Overview text that conflicts with the REQs changed in round 10. Publishing, mapping, activation and attribution agree (`html:198,265-267,308,323-324,353,366,449-450`). | consistent |
| Conditions from round 10 | **Provided by the main agent:** digests, `readlink`, `git diff --cached --stat`, 312 keys, and the union and cross-reference check outputs.<br>**Recounted by the peer:** 312 JSON keys; 388 map rows (323 REQ, 43 ADR, 22 dropped); 331 inventory rows; 72 audit candidates (64 distinct). | met (main-agent outputs accepted) |

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R11-01 | nit | `requirements-map.md:71` ("Design Appendix 20" → REQ-31, REQ-58, ADR-005); `:367`; `design-section-map.json:499-503` (`"Appendix 20"`) | **A citation of "Appendix 20" does not reach ADR-006, which closed its pending clause, or the REQ that carries its retained default.**<br>• **Base text:** Appendix 20 (`D:497`) holds three clauses:<br>  – the codex-only mapper pin;<br>  – "The area-match default Jaccard 0.5 is kept for now (A1)" plus the X2b-pending clause;<br>  – "**Pending clause closed 2026-10-24 by item 21**".<br>  ADR-006 made the closure (`2026-10-24-area-refresh-fallback/ADR.md:35`: "Close ADR-005's pending-X2b clause in … Appendix decision 20").<br>• **Two-source row:** `map:367`, "Design §15 former numbered decisions / Appendix 20 pending study clause", holds ADR-002, ADR-003, ADR-005 and ADR-006. Neither the `Appendix 20` row nor its JSON key includes them.<br>• **Rules broken:**<br>  – map `:3`: a parent holds its sub-items' IDs;<br>  – `proposal.md:167`: a key holds the IDs of every map row it represents.<br>  The same pattern is handled correctly for `§14 A8 and A10` and `§10.3 / Appendix 6,12`.<br>• **Carriers miss ADR-006:** REQ-31 cites ADR-002/005/007/008 (`html:697`) and REQ-58 cites ADR-004/005 (`html:551`).<br>• **No carrier for the retained 0.5 default:** that clause is carried by REQ-53 (`html:546`, "`mapping.areas.match_jaccard` 0.5 [ASSUMPTION…]"), which is not on the row.<br>• **Effect:** frozen citations of "Appendix 20" do not lead to the closing decision or the carrying REQ. These are the ledger rows R-040, R-042 and R-043, ADR-000 `ADR.md:36` and ADR-005 `ADR.md:40`. This is the R10-07 failure mode. | Make `map:71` and JSON `"Appendix 20"` hold ADR-006 and REQ-53, and optionally REQ-170.<br>Then do one of these:<br>• add all of `map:367`'s IDs as well; or<br>• split `:367` into a §15 row and an "Appendix 20 pending study clause" row (ADR-005, ADR-006), and update the row counts.<br>Extend the union script to rows that name two sources, and disclose the change in `proposal.md`. |
| R11-02 | nit | `proposal.md:167` | **The proposal reports "`req §n` ← `req REQ-n` pairs: 43", but the map has 41.** There are exactly 41 `req REQ-n` rows (`map:298-338`), and base `SCOUT_REBUILD_REQUIREMENTS.md:7-75` places each in exactly one section. **[INFERENCE]** The extra two are the non-`req REQ-n` children `map:340` (dropped) and `:341` (ADR). This figure is reported as script output in a signed record. | Report 41, or say what the 43 counts. |
| R11-03 | nit | `requirements-map.md:3` (rule); `:48` ("Design Appendix heading" → ADR-000…ADR-009); `:342` ("req Appendix — Removed / superseded heading" → ADR-009) | **Two heading rows do not follow the union rule as map `:3` now states it.**<br>• **The rule:** a parent holds the IDs of its "sub-items" and the "targets named by `dropped` children".<br>• **`:48`:** maps the base heading `## Appendix — Decisions taken` (`D:476`), whose items 1–23 are its sub-items. Yet `:48` holds none of their REQ IDs, unlike the §12 and §13 parents, which do hold their items' IDs.<br>• **`:342`:** its 17 dropped children `:343-359` name "REQ-1–REQ-41 and their supplemental REQs … ADR-000–ADR-008", but the row holds only ADR-009.<br>• **Impact:** neither row has a JSON key, and no bare "design Appendix" or "req Appendix" citation exists in the base or worktree (peer grep), so no citation resolves wrongly. But the stated rule and the "0 map parents missing" claim (`proposal.md:167`) hold only because the script does not treat these rows as parents. | Choose one:<br>• make `:48` and `:342` unions; or<br>• state in map `:3` that document-title and appendix-heading rows (`:48`, `:342`, and the title rows) map only their own heading text, and that the generic range provenance in `:343-359` is not a union target.<br>Then correct the proposal's check description to match. |

## Answers to the proposal's questions

1. **Is normative content lost, weakened or changed?** **No, apart from the R11-01 index gap.**
   - R10-04, R10-05 and R10-06 restore their qualifiers word for word (`D:175,177,394`).
   - R10-02 and R10-03 now point to REQs that hold the referenced criteria and values.
   - The ADR-column additions change no row text.
   - The inventory re-check found no further loss, and the peer's sample of more than 25 quotes confirms it.
2. **Do the ADR summaries and cards match their records?** **Yes.** The peer checked every card's status, relation line, round count and issue count against its `ADR.md`:
   - ADR-001: 5 rounds, 12 → 0 issues;
   - ADR-002: 6, 25 → 0;
   - ADR-003: 7, 16 → 0;
   - ADR-004: 5, 10 → 0;
   - ADR-005: 6, 17 → 0;
   - ADR-006: 5, 12 → 0;
   - ADR-007: 7, 16 → 0;
   - ADR-008: 11, 14 → 0.

   The ADR-009 card (`html:949-972`) matches `ADR.md` (lead).
3. **Does the Overview agree with Appendix A?** **Yes:** the challenger and lead found no conflict in the round-10 areas (see the verification table).
4. **Are rule, tooling and citation changes complete, with no live reference to a deleted file?**
   - **Deleted files:** no live reference. The peer grepped `internal/`, `cmd/`, `.github/`, `scripts/`, `toolchain/`, `.omp/`, `.agents/`, `AGENTS.md`, `README.md`, `CONTRIBUTING.md` and `.prettierignore`. The only `docs/` hits are:
     - append-only ledger rows;
     - `jev-adapter-audit.md:7`;
     - REQ-110's pinned source;
     - ADR-009's own records;
     - the README path map;
     - the plan's deletion list.
   - **Citation index:** incomplete in one place: R11-01.
5. **Experiment rules, ledgers, frozen records:** **no violation.**
   - `AGENTS.md:10-31` matches base rules 1–7 word for word, apart from the settled preamble.
   - The research ledger only appends R-062.
   - The hypothesis-ledger History line follows the ledger's newest-first order.
   - The frozen digests are unchanged.

## Considered and not raised

- **Extra ADR citations, such as ADR-005 on REQ-167 and ADR-002 on REQ-178.** Each is supported by the cited `+` line. Extra citations are not defects (`round-5.md:110`).
- **REQ-91 omits "How X1 treats capped and failed units is fixed by its preregistration".** REQ-169 carries it (`html:686`, "treatment of capped and failed units"), and audit `:248` names REQ-169.
- **REQ-167 drops "(`docs/hypothesis-ledger.md`)" after HD5.** This is an evidence pointer, not an obligation.
- **`D:170.5` "the 'model CLI logged in' check covers both paths".** REQ-60 and REQ-61 carry the readiness and skip behaviour for every model-CLI lane. Only the bootstrap-script version report is omitted, and it is not a runtime obligation.
- **Appendix 1, 6, 11, 21 and 22(a) rows lack some amending ADRs.** Map `:3` says REQ targets "may" name the governing ADR. Unlike Appendix 20, none of these rows has a sub-item row that holds the missing ADR, and each amending ADR is reachable through a carrier REQ (peer).
- **Design, AL and req title rows are not unions.** They map document preambles. R11-03 asks only that the rule say so.
- **JSON `"§14 A10"` is a superset of its row.** Settled in `round-8.md:118`.
- **D:176.3's carrier, REQ-154, does not say "not annotations".** REQ-28 restricts annotations to triage output, so the meaning is preserved (peer).
- **"Accepted" status on ADR-009 before sign-off.** This is the adoption-commit convention, accepted since R1-07.

## Peers

- **`innovation-challenger`:**
  - Confirmed R10-01…R10-06, the proposal's html claims, 20 frozen-source traces, 28 internal-reference pairs, the five legitimate `§` symbols, and questions 1 and 3.
  - Raised no issues.
- **`experiment-peer`:**
  - Confirmed R10-07…R10-11, the counts and the audit totals, and answered questions 2, 4 and 5.
  - Found R11-01, R11-02 and R11-03.
- **Lead:**
  - Separately verified R10-01…R10-07 against `D:173-177,393-394,444` and the 14 added citations against their frozen lines.
  - Hand-recomputed every `req §n` and AL union.
  - Confirmed each peer finding against `D:476,497`, map `:3,48,71,298-342,367`, JSON `:499-503` and area-refresh-fallback `ADR.md:35`.
  - Extended R11-01 with the missing REQ-53 carrier for Appendix 20's retained 0.5 default.
- **Where the peers changed the outcome:** working alone, the lead found no issue and would have signed off. The experiment-peer's check of rows that name two sources (R11-01), its recount of the pairs (R11-02) and its literal reading of the new map `:3` rule (R11-03) showed that the extended union check still skips rows it does not classify as parents.

## Verdict

**3 issues (R11-01, R11-02 and R11-03, all nits): NOT signed off.**

No sign-off applies to the digests above.

### Required with the next submission

- **Changed digests:** expected only for:
  - `requirements-map.md` and `design-section-map.json` (R11-01, R11-03);
  - `proposal.md` (R11-02, plus a "Round 11 fixes" section).

  Any other changed digest must be disclosed in `proposal.md`.
- **Recomputed digests and checks:** the main agent must recompute every digest with `shasum -a 256` on the staged worktree. It must also re-run:
  - `test -L CLAUDE.md && readlink CLAUDE.md`, expecting `AGENTS.md`;
  - `git -C /Users/adamjackson/Projects/work/scout-req diff --cached --stat`.

  It must report:
  - the JSON key count and the map row counts, with any change from 312 and 388 (323/43/22) explained;
  - the union check output, now also covering every map row whose Source names two sections (for example `:367`, `:364`, `:383`) against each section's row and key.

---

## Council memo

1. **Recommendation:** fix R11-01…R11-03 in one pass, which means map, JSON and proposal edits only, and submit round 12. Do not change the html or the audit, so that the content signed off in substance this round stays byte-stable.
2. **Alternatives:**
   - (a) **Sign off with conditions.** Rejected: design rule 1 requires zero issues.
   - (b) **Narrow the map `:3` claim to design sections and drop the multi-source rule.** Rejected: frozen ledger rows and ADR summaries cite "Appendix 20", and `adrs/README.md:29` makes the JSON their resolver.
   - (c) **Leave R11-02 and R11-03 as wording.** This was not adopted. The fixes are one-line edits, and the proposal's check outputs are part of the signed record.
3. **Evidence:** each issue row cites its exact worktree and base `file:line`, and the verification table cites every frozen source line used.
4. **Risks:**
   - **R11-01:** citations of "Appendix 20" miss the X2b closure (ADR-006) and the 0.5 carrier (REQ-53).
   - **R11-02 and R11-03:** a signed record misstates what its own check covered.
   - **Reversibility:** every fix is a text or JSON edit and is fully reversible.
   - **Unknowns:** the council saw no per-file staged diff. It relies on the main agent's digests and statements for "only these files changed" and for the "64 carriers corrected" and "11 parents / 15 keys extended" counts. The peer's inference agrees with 11 and 15.
5. **Execution outline:**
   1. **`map:71` and JSON `"Appendix 20"`:** add ADR-006 and REQ-53, plus optionally REQ-170. Either add ADR-002 and ADR-003 from `:367`, or split `:367` into a §15 row and an Appendix 20 row.
   2. **Map `:3`:** state how title and appendix-heading rows (`:48`, `:342`) and generic range provenance (`:343-359`) are treated, or make `:48` and `:342` unions.
   3. **`proposal.md:167`:** correct "43" to 41, or define it. Add "Round 11 fixes".
   4. **Checks:** extend the union script to rows that name two sources and re-run it. Recompute every digest with `shasum -a 256`, run `readlink` and `git diff --cached --stat`, report the key and row counts, and submit round 12 with the outputs.
