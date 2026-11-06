# Round 12: ADR-009, one design document with requirements and ADR appendices

Date: 2026-11-06. This round follows design rules 1–3 in `AGENTS.md`/`CLAUDE.md`. The innovation council lead ran it with two independent peers. Neither peer was told the lead's view.
- **`innovation-challenger` (GPT):**
  - checked the R11-01 split against base `D:468-472,476,497-498` and the ADR-000, ADR-002, ADR-003 and ADR-006 records;
  - tested the new multi-source clause and the exception in map `:3` adversarially, including whether each base heading contains its child rows;
  - compared 21 metric REQs (REQ-103…REQ-123) with `D:233-241,251-262`. Rounds 9–11 had not sampled these;
  - checked the Overview against Appendix A.
- **`experiment-peer` (Opus):**
  - recounted the map, the JSON and the `req REQ-n` rows;
  - recomputed every parent union and every JSON key that stands for several rows, by hand;
  - checked the proposal's "Round 10 fixes" and "Round 11 fixes";
  - answered questions 2, 4 and 5.

The review was read-only. It made no model calls outside the council and did not touch `.codegraph`.

**The main agent recomputed the digests; the council did not.** The lead and both peers had only read-only tools and no shell, so nobody on the council hashed a file. Earlier rounds accepted a main-agent recomputation, and this round does too. The main agent reported these outputs on the staged worktree:
- `shasum -a 256` on each file below;
- `test -L CLAUDE.md && readlink CLAUDE.md` → `AGENTS.md`;
- `git diff --cached --stat` → `251 files changed, 7874 insertions(+), 1781 deletions(-)`. That is round 11's 250 files plus the new `round-11.md`.

No Go file has changed since round 6.

| File | SHA-256 (main agent, `shasum -a 256`) | vs round 11 |
|---|---|---|
| `docs/scout-design.html` | `f0aa226a88c3c8eda782a7327be85b2bccb5a879df177673d12021d50ffdc463` | unchanged |
| `docs/adrs/2026-11-06-design-restructure/proposal.md` | `a719b3698433c9db08f05090c165ef8e8893f4a09615b7b616691297ccddd267` | changed (R11-02, "Round 11 fixes") |
| `…/requirements-map.md` | `95d6b1a150ba35f3b7b3cacf00e61e801f5d664e48d4ccba0a20744ca2ddea1f` | changed (R11-01, R11-03) |
| `…/design-section-map.json` | `f4d3526944414a677a92a402dce4d0acc8c3df6a1b7eff766d0a230590e2df35` | changed (R11-01) |
| `…/ADR.md` (ADR-009) | `1393b6eb7488688c24ef25106ec6861616ae61e75675570651e5b97ab1fae83a` | unchanged |
| `…/completeness-audit.md` | `0fa44fa63e084d2ea9b0690fc5bd159037cb526980a736e99da621ea9505c31b` | unchanged |
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

**The digest changes match round 11's "Required with the next submission" list exactly.** Only `proposal.md`, `requirements-map.md` and `design-section-map.json` changed, plus the new `round-11.md`. The html, the audit, `ADR.md` and the 12 other files are byte-identical to `round-11.md`'s table. The peer's file timestamps agree **[INFERENCE]**.

## Scope

- **Under review:** worktree `/Users/adamjackson/Projects/work/scout-req`, branch `agent/requirements-path`, with all changes staged. The review covered:
  - the "Round 11 fixes" (`proposal.md:170-178`);
  - the R11-02 edit to the round-10 check line (`proposal.md:167`);
  - the new map `:3` multi-source clause and its exception;
  - the split rows `map:367-368`;
  - map `:47,71` and JSON `"Appendix 20"` (`:499-506`) and `"§15"` (`:1047-1053`);
  - a final broad pass over proposal questions 1–5.
- **Base:** `main` at `9a3d19f5` in `/Users/adamjackson/Projects/work/scout/`. `D:n` is line n of the base `docs/scout-design.md`, and `AL:n` is line n of the base `docs/scout-allowlist.md`.
- **Checks made:**
  - **Base design:** `D:7-21,468-472,476-500`.
  - **Allow-list structure:** `AL:5,11,13,15,26,28`.
  - **Requirements sections:** `SCOUT_REBUILD_REQUIREMENTS.md:7-81`.
  - **ADR records:**
    - `2026-10-19-v1-decisions/ADR.md:5,20-21,36-37`;
    - `2026-10-23-decisions/ADR.md:30`;
    - `2026-10-24-area-refresh-fallback/ADR.md:35`;
    - `2026-10-23-codex-luna-mapping/scout-design.proposed.md:444-448` and `proposed.diff:102-103`;
    - `2026-10-24-area-refresh-fallback/proposed.diff:145-149`.
  - **Map and JSON:** map `:1-399`; JSON `:1-40,495-506,1040-1056,1585-1625`; html REQ rows `:546,551,623-651,687,697`.
- **Not reopened:** decisions settled in ADR-000…ADR-008, and earlier "Considered and not raised" items unless there is new evidence.

## Verification of round-11 fixes

| # | Evidence (worktree vs base) | Result |
|---|---|---|
| R11-01 | **The Appendix 20 row and key are fixed.**<br>• `map:71` = REQ-31, REQ-58, ADR-005, REQ-53, REQ-170, ADR-006.<br>• JSON `"Appendix 20"` (`:499-506`) holds the same six IDs.<br>• REQ-53 (`html:546`) carries the retained 0.5 default [ASSUMPTION], and REQ-170 (`:687`) carries X2b. ADR-006 closed the pending clause (`area-refresh-fallback/ADR.md:35`).<br>**The two-source row is split:**<br>• `:367` "Design §15 former numbered decisions" → ADR-002, ADR-003;<br>• `:368` "Design Appendix 20 pending study clause" → ADR-005, ADR-006, contained in `:71` and in the key.<br>**The split row `:367` misses one decision record:** R12-01. | fixed (new row: R12-01) |
| R11-02 | **The pair count is corrected.** `proposal.md:167,175,178` report 41.<br>• The map has exactly 41 `req REQ-n` rows (`map:298-338`), each in one base section (`SCOUT_REBUILD_REQUIREMENTS.md:7-75`).<br>• The explanation of the old 43 adds up: 41 plus `map:340` (dropped) and `:341` (ADR).<br>• The edited round-10 line says plainly that the round-10 submission reported 43, so it does not rewrite history. | fixed |
| R11-03 | **Map `:3` now carries the exception.** It exempts the document-title rows, "Design Appendix heading" (`:48`) and "req Appendix — Removed / superseded heading" (`:342`), and it removes the generic range provenance of `:343-359` from the union target. That resolves the two rows R11-03 named.<br>**The exception's wording is inaccurate for three of the rows it covers:** R12-03.<br>**A non-exempt subsection parent still breaks the rule:** R12-02. | fixed for the named rows (R12-02, R12-03) |
| Multi-source clause (`map:3`) | **Both named rows propagate:**<br>• `:364` (ADR-009) reaches §10.3 (`:42`), §10 (`:39`), Appendix 6 (`:78`) and Appendix 12 (`:80`), and the JSON keys `"§10.3"`, `"Appendix 6"` and `"Appendix 12"`;<br>• `:384` (ADR-002) reaches A8 (`:155`), A10 (`:157`) and §14 (`:46`), and the JSON keys `"§14 A8"` (`:1015`), `"§14 A10"` (`:1021`) and `"§14"`.<br>The removed-item rows that name several old sections (for example `:343`, "(§§3, 10, 12.1, 13)") carry only generic range provenance, which the exception removes. "§4 observed E3/CP07/X2/X2b …" (`:362`) names one design section; X2 and X2b in it are study names. | satisfied |
| Parent unions (lead and peer, by hand) | **Every named-section parent holds all its children's IDs, dropped targets included.** The lead and the peer each recomputed:<br>• Design §1–§15 and their subsections (§3.x, §5.x, §7.x, §8.x, §10.x);<br>• Appendix 20, 22 and 23;<br>• AL §1–§5;<br>• `req §1`–`§10`;<br>• Historical `req §11` and `§12`.<br>**Exception:** the allow-list subsection row, R12-02. | pass except R12-02 |
| Counts | **The peer's recount matches the main agent's report:**<br>• 389 map rows: 323 REQ, 44 ADR (listed `map:11,13,48-58,78-80,128-129,155,157,269,276,282,339,341-342,360-364,366-371,377,382-384,387-389`), 22 dropped (`:281,286,287,340,343-359,365`);<br>• 312 JSON keys.<br>One ADR row more than round 11, from the R11-01 split. | match |
| Conditions from round 11 | **Provided by the main agent:** digests, `readlink`, `git diff --cached --stat`, the key and row counts, and the union check output including multi-source rows. | met (main-agent outputs accepted) |

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R12-01 | nit | `requirements-map.md:367` ("Design §15 former numbered decisions" → ADR-002, ADR-003); `:47` (§15 parent); `design-section-map.json:1047-1053` (`"§15"`) | **The new §15 history row omits ADR-000.**<br>• **Base text:** the passage `:367` maps (`D:472`) names three decided sets:<br>  – "Former items 1 (study export), 2 (PR metadata) and 3 (inertia framing) were decided on 2026-10-23 and are Appendix 16, 17 and 18" (ADR-003, `2026-10-23-decisions/ADR.md:30`);<br>  – "**All other open items from v1 §15 are decided (Appendix)** and applied above" (ADR-000);<br>  – "revision 2.1's items 2 (X8) and 3 … left the list in revision 2.2 (Appendix 11, 14)" (ADR-002).<br>• **ADR-000 is that record:** its sources are "`docs/archive/scout-design-v1.md`@9a3d19f5 §15 and `docs/scout-design.md`@9a3d19f5 Appendix items 1–7" (`2026-10-19-v1-decisions/ADR.md:5,20-21`). `D:476` dates Appendix items 1–7 to "v1 §15, 2026-10-19".<br>• **Why now:** the merged row of round 11 also lacked ADR-000, but R11-01's split isolates this passage as its own ADR-disposition row for the first time. An ADR row's only job is to name the decision record (map `:5`).<br>• **Related:** after the split, the §15 parent and its key still hold ADR-005 and ADR-006. No §15 child row names them, and base §15 never carried an Appendix-20 clause (`D:468-472`; codex-luna `scout-design.proposed.md:444-448`). Extra IDs are not defects (`round-5.md:109-110`), so this part is not raised. | Add ADR-000 to `map:367`, to the §15 parent `:47` and to JSON `"§15"`. Optionally remove ADR-005 and ADR-006 from `:47` and `"§15"` in the same edit, and disclose the change in `proposal.md`. |
| R12-02 | nit | `requirements-map.md:271` ("AL §1 Eligible blocking checks — …" → REQ-70…REQ-78); `:280` ("AL §1 demoted seed checks" → REQ-21, REQ-79…REQ-83); `design-section-map.json:1594-1604`; rule `map:3` | **The allow-list subsection row is not a union of its children.**<br>• **Base structure:** the subsection `### Eligible blocking checks` starts at `AL:15`. Its table (`AL:17-24`) and the paragraph "Demoted from the seed **[CH]**: `guardrails.GATE.coverage-pragma-added` → report-only (table 2); `generic-api-key` → report-only; wildcard `CONFIG.*` → exact ids" (`AL:26`) both come before the next heading, `## 2.` (`AL:28`). So that paragraph is a child of the subsection.<br>• **The gap:** the paragraph's row `:280` holds REQ-21 and REQ-79…REQ-83, but the subsection row `:271` and its JSON key hold only REQ-70…REQ-78. This row is not among the exceptions.<br>• **New evidence:** the base heading boundary. Round 9 (R9-07, `round-9.md:115`) repaired only the AL §1 parent (`:270`), which is correct. A check that pairs rows by name cannot see this child, because `:280` is named "AL §1 …", not "AL §1 Eligible …".<br>• **Why the check missed it:** `proposal.md:178` reports 0 parent gaps. | Add REQ-21 and REQ-79…REQ-83 to `map:271` and to JSON `"AL §1 Eligible blocking checks"`. Extend the union check so it uses heading containment in the three base documents as well as row names, re-run it, and report the output. |
| R12-03 | nit | `requirements-map.md:3` (exception sentence); `:11`, `:269`, `:339`, `:342` | **The exception says the exempt rows "map only their own heading text". Three of them also map prose:**<br>• `:11`: "Design title / status / adoption and implementation chronology";<br>• `:339`: "req title / revision/status authority paragraphs";<br>• `:342`: "… heading and provenance note".<br>**"Document-title rows" is also never tied to specific rows.** **[INFERENCE]** It means `:11`, `:269` (the AL H1, `AL:1`) and `:339`.<br>**Cause:** the wording follows round 11's own suggested fix (`round-11.md:82`, "map only their own heading text"). The council owns that part of the cause.<br>**Effect:** no ID changes, but the sentence misdescribes rows in the same file. | Reword map `:3` to say that these rows map only their own text and do not take their children's IDs. Name the rows, for example: "the document-title rows (`Design title …`, `AL Scout PR blocking allow-list …`, `req title …`), 'Design Appendix heading' and 'req Appendix — Removed / superseded heading and provenance note' map only their own text (heading and the prose named in their Source), not their children's IDs". |

## Answers to the proposal's questions

1. **Is normative content lost, weakened or changed?** **No.**
   - The html is byte-identical to round 11, whose content was signed off in substance.
   - The challenger compared 21 metric REQs, which rounds 9–11 had not sampled:
     - REQ-103…REQ-111 (`html:623-631`) against `D:233-241`;
     - REQ-112…REQ-123 (`:632-643`) against `D:251-262`.

     Formulas, denominators, thresholds, tool settings, availability phases and missingness qualifiers are preserved. REQ-125…REQ-131 (`:645-651`) carry the population and comparison context (`D:229,243,247,267-271`).
   - This round's findings are index traceability defects, not normative loss.
2. **Do the ADR summaries and cards match their records?** **Yes.**
   - **Relation lines:** each card's relation line (`html:736,759,783,807,831,854,877,901,925,949`) matches the status line of its `ADR.md`, and supersede/amend links agree in both directions.
   - **Round counts:** ADR-001: 5; ADR-002: 6; ADR-003: 7; ADR-004: 5; ADR-005: 6; ADR-006: 5; ADR-007: 7; ADR-008: 11.
   - **Sign-off digests:** the peer spot-checked them against each final round, for example `d1fb1274…` and `dab8fa33…` in jev-merge `round-11.md:5`.
3. **Does the Overview agree with Appendix A?** **Yes.** The challenger found no new contradiction:
   - triggers and publishing (`html:198,308,323-330,366`) agree with REQ-26, REQ-27 and REQ-160;
   - decision ownership (`:340-368`) agrees with REQ-28, REQ-139, REQ-140, REQ-153 and REQ-154;
   - gate, activation and attribution (`:389-450`) agree with REQ-20…24 and REQ-78;
   - ledger records (`:455-509`) agree with REQ-5…9.
4. **Are rule, tooling and citation changes complete, with no live reference to a deleted file?**
   - **Deleted files:** no live reference. The peer grepped `internal/`, `cmd/`, `.github/`, `scripts/`, `toolchain/`, `.omp/`, `.agents/`, `.scout/`, `AGENTS.md`, `README.md`, `CONTRIBUTING.md` and `.prettierignore`. The only remaining hits are pinned or historical:
     - `jev-adapter-audit.md:7`;
     - REQ-110 at `9a3d19f5`;
     - the ADR-009 card;
     - the plan's deletion list;
     - `adrs/README.md:29`.
   - **Citation index:** two gaps remain, R12-01 and R12-02.
5. **Experiment rules, ledgers, frozen records:** **no violation.**
   - `AGENTS.md` experiment rules 1–7 match the base word for word, apart from the settled preamble (R1-14).
   - The research ledger is identical to the base through R-061 and appends R-062.
   - The hypothesis ledger adds one History line, newest first.
   - The frozen digests are unchanged.

## Considered and not raised

- **§15 parent and its key keep ADR-005 and ADR-006 after the split.** No base text supports them, but extra IDs are not defects (`round-5.md:109-110`). Trimming them is optional under R12-01.
- **The proposal says "the two remaining such rows" (`proposal.md:174`), but the removed-item rows `:343,344,347,350,354,355,358` also name several old sections.** Their only IDs are the generic range, which the `:3` exception removes, so nothing is missed (peer).
- **`proposal.md:167` states the cause of the old 43 as fact, where round 11 marked it [INFERENCE].** The arithmetic is consistent: 41 + `:340` + `:341` = 43.
- **`proposal.md:167` keeps round 10's "0 JSON keys missing".** That was the true output of round 10's narrower check, and R11-01 is disclosed at `:174`.
- **`"§13 X1"`, `"§12 R1"` and `"§14 A1"` do not include X10/X11, R10–R18 or A10–A14.** These are different items, not children.
- **"Design Appendix heading" (`:48`) lacks ADR-001.** It is exempt under the `:3` exception.

## Peers

- **`innovation-challenger`:**
  - Confirmed R11-01's Appendix 20 additions, R11-02, the multi-source propagation, questions 1 and 3, and the 21 metric rows.
  - Found R12-02 and confirmed R12-01 independently.
- **`experiment-peer`:**
  - Confirmed R11-01…R11-03, every count (389 rows: 323/44/22; 312 keys; 41 pairs), every parent union it named, and JSON keys equal to their map unions; answered questions 2, 4 and 5.
  - Found R12-03.
  - Read `:367` as complete. The lead and challenger disagree on the evidence of `D:472` ("All other open items from v1 §15 are decided (Appendix)") and ADR-000 `ADR.md:5,20-21`.
- **Lead:**
  - Separately recomputed every design, Appendix, AL, `req` and Historical union and the multi-source rows.
  - Found R12-01, and the surplus ADR-005 and ADR-006 on §15, before the peers reported.
  - Confirmed R12-02 against `AL:15,26,28` and R12-03 against map `:11,339,342`.
- **Where the peers changed the outcome:** working alone, the lead would have reported one issue (R12-01). The challenger's base-heading containment test added R12-02. The peer's literal reading of the new exception sentence added R12-03. On R12-01 the peer's dissent did not change the outcome, because base text resolved it.

## Verdict

**3 issues (R12-01, R12-02 and R12-03, all nits): NOT signed off.**

No sign-off applies to the digests above. The html, the audit, `ADR.md` and every other file not named below are signed off in substance and should stay byte-stable.

### Required with the next submission

- **Changed digests:** expected only for:
  - `requirements-map.md` (R12-01, R12-02, R12-03);
  - `design-section-map.json` (R12-01, R12-02);
  - `proposal.md` ("Round 12 fixes").

  Any other changed digest must be disclosed in `proposal.md`.
- **Recomputed digests and checks:** the main agent must recompute every digest with `shasum -a 256` on the staged worktree. It must also re-run:
  - `test -L CLAUDE.md && readlink CLAUDE.md`, expecting `AGENTS.md`;
  - `git -C /Users/adamjackson/Projects/work/scout-req diff --cached --stat`.

  It must report:
  - the JSON key count and the map row counts, with any change from 312 and 389 (323/44/22) explained;
  - the union check output, now also covering children by base heading containment in `docs/scout-design.md`, `SCOUT_REBUILD_REQUIREMENTS.md` and `docs/scout-allowlist.md`.

---

## Council memo

1. **Recommendation:** fix R12-01…R12-03 in one pass, which means map, JSON and proposal edits only, and submit round 13.
   - Add ADR-000 to `map:367`, `:47` and `"§15"`.
   - Add REQ-21 and REQ-79…REQ-83 to `map:271` and `"AL §1 Eligible blocking checks"`.
   - Reword the exception in map `:3`.
   - Do not touch the html or the audit.
2. **Alternatives:**
   - (a) **Sign off with conditions.** Rejected: design rule 1 requires zero issues.
   - (b) **Re-merge the §15 and Appendix 20 rows into one row holding ADR-000, 002, 003, 005 and 006, and propagate the full union to both.** Rejected: it adds unrelated ADRs to Appendix 20's resolver (challenger).
   - (c) **Exempt subsection rows from the union rule.** Rejected: frozen citations resolve through the JSON (`adrs/README.md:29`), and a narrower rule would hide real child relations.
3. **Evidence:** `D:472,476`; `2026-10-19-v1-decisions/ADR.md:5,20-21`; `2026-10-23-decisions/ADR.md:30`; `AL:15,26,28`; map `:3,11,47,71,269,271,280,339,342,367-368`; JSON `:499-506,1047-1053,1594-1604`; `proposal.md:167-178`.
4. **Risks:**
   - **R12-01:** citations of §15 miss the v1 decision record.
   - **R12-02:** a citation of the eligible-checks subsection misses the demoted checks (REQ-79…83) and the no-wildcard rule (REQ-21).
   - **R12-03:** a signed record misdescribes its own exempt rows.
   - **Reversibility:** every fix is a text or JSON edit and is fully reversible.
   - **Unknowns:** the council saw no per-file staged diff and relies on the main agent's digests and its statement that only three files changed. Peer timestamps agree **[INFERENCE]**.
   - **Residual:** checks that only compare row names can miss parent/child relations. The heading-containment check is meant to close that gap.
5. **Execution outline:**
   1. Edit map `:367`, `:47` and JSON `"§15"`: add ADR-000, and optionally drop ADR-005 and ADR-006.
   2. Edit map `:271` and JSON `"AL §1 Eligible blocking checks"`: add REQ-21 and REQ-79…REQ-83.
   3. Reword map `:3` as in R12-03.
   4. Extend the union script to use base heading containment, then re-run the union, cross-reference and count checks.
   5. Add "Round 12 fixes" to `proposal.md` with the script outputs.
   6. Recompute every digest with `shasum -a 256`, run `readlink` and `git diff --cached --stat`, and submit round 13.
