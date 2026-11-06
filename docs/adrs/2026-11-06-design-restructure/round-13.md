# Round 13: ADR-009, one design document with requirements and ADR appendices

Date: 2026-11-06. This round follows design rules 1–3 in `AGENTS.md`/`CLAUDE.md`. The innovation council lead ran it with two independent peers, and neither peer was told the lead's view.

- **`innovation-challenger` (GPT):**
  - tested the reworded union rule and exception in map `:3` adversarially;
  - checked heading containment against every heading in the three base documents;
  - verified R12-01 against `D:468-476` and ADR-000;
  - verified R12-02 against `AL:7-28`;
  - fact-checked the counts in "Round 12 fixes";
  - sampled 15 REQs that rounds 9–12 had not sampled;
  - re-checked the Overview against Appendix A.
- **`experiment-peer` (Opus):**
  - recounted the map rows and JSON keys;
  - recomputed by hand every union that round 12 touched;
  - compared line positions against round 12's citations;
  - fact-checked "Round 12 fixes";
  - answered questions 2, 4 and 5.

The review was read-only. It made no model calls outside the council and did not touch `.codegraph`.

**The main agent computed the digests; the council did not.** The lead and both peers had only read-only tools and no shell, so no council member hashed a file. Earlier rounds accepted the main agent's recomputation, and this round does too. The main agent reported these outputs on the staged worktree:
- `shasum -a 256` on each file below;
- `test -L CLAUDE.md && readlink CLAUDE.md` → `AGENTS.md`;
- `git diff --cached --stat` → `252 files changed, 8087 insertions(+), 1781 deletions(-)`. That is round 12's 251 files plus the new `round-12.md`.

No Go file has changed since round 6.

| File | SHA-256 (main agent, `shasum -a 256`) | vs round 12 |
|---|---|---|
| `docs/scout-design.html` | `f0aa226a88c3c8eda782a7327be85b2bccb5a879df177673d12021d50ffdc463` | unchanged |
| `docs/adrs/2026-11-06-design-restructure/proposal.md` | `7135fe5d3153285265a7d61cf1da274929276206053950eefa8bccb57b00912f` | changed ("Round 12 fixes") |
| `…/requirements-map.md` | `9a8db8aa2df66475cd0665f7c03e1af64ee83216f64b804491f0eab4522ded53` | changed (R12-01, R12-02, R12-03) |
| `…/design-section-map.json` | `ad9dde4b0e6e99937b744a3bffcb40c68dfc84d166b6bb74960f268dfd853bb9` | changed (R12-01, R12-02) |
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

**The digest changes match round 12's "Required with the next submission" list exactly.** Only `proposal.md`, `requirements-map.md` and `design-section-map.json` changed, plus the new `round-12.md`. The html, the audit, `ADR.md` and the 12 other files are byte-identical to `round-12.md`'s table. The peer's file timestamps agree **[INFERENCE]**.

## Scope

- **Under review:** worktree `/Users/adamjackson/Projects/work/scout-req`, branch `agent/requirements-path`, with all changes staged. The review covered:
  - "Round 12 fixes" (`proposal.md:180-188`);
  - the reworded map `:3`;
  - map `:11,47,48,269-287,339-342,360-389`;
  - JSON `"AL §1"` (`:433-451`), `"Appendix 20"` (`:499-506`), `"§15"` (`:1047-1054`) and `"AL §1 Eligible blocking checks"` (`:1595-1611`);
  - a final broad pass over proposal questions 1–5.
- **Base:** `main` at `9a3d19f5` in `/Users/adamjackson/Projects/work/scout/`. `D:n` is line n of the base `docs/scout-design.md`; `AL:n` is line n of the base `docs/scout-allowlist.md`.
- **Checks made:**
  - **Every heading in the three base documents:**
    - design `D:1,25,41,61,63,84,94,100,106,123,125,141,145,158,192,194,198,202,213,219,221,227,245,265,273,295,303,318,333,335,345,368,374,406,431,450,468,476`;
    - allow-list `AL:1,5,15,28,42,49,59`;
    - requirements `SCOUT_REBUILD_REQUIREMENTS.md:1,7,14,24,31,38,49,57,64,68,75,79`.
  - **Base text:**
    - `D:5-21` (revision paragraphs and evidence notation, all under the H1);
    - `D:261-263` (the last §8.3 metrics, including `fix_touches`);
    - `AL:7-27`.
  - **Wider repository:** a grep of the html, `ADR.md`, the audit, `adrs/README.md` and `AGENTS.md`, for any restatement of the union rule or the map counts. There were no hits.
- **Not reopened:** decisions settled in ADR-000…ADR-008, and earlier "Considered and not raised" items unless there is new evidence.

## Verification of round-12 fixes

| # | Evidence (worktree vs base) | Result |
|---|---|---|
| R12-01 | **ADR-000 is now on all three targets:**<br>• the row `map:367`, "Design §15 former numbered decisions" → ADR-002, ADR-003, ADR-000;<br>• the §15 parent `:47` → REQ-98, ADR-002, ADR-003, ADR-005, ADR-006, ADR-000;<br>• JSON `"§15"` (`:1047-1054`), which holds the same six IDs.<br>**The base supports it:** `D:472` says "All other open items from v1 §15 are decided (Appendix)", and `D:476` dates Appendix items 1–7 to "v1 §15, 2026-10-19". ADR-000 records exactly those items (`2026-10-19-v1-decisions/ADR.md:5,20-21`).<br>**The hand union of §15's children holds:** `:162` (REQ-98) plus `:367` gives REQ-98, ADR-000, ADR-002 and ADR-003, and the parent and key contain all four.<br>**Disclosed extras:** ADR-005 and ADR-006 stay on the parent (`proposal.md:184`). Extra IDs are not defects (`round-5.md:109-110`). | fixed |
| R12-02 | **The row and key now hold the child's IDs:** `map:271` and JSON `"AL §1 Eligible blocking checks"` (`:1595-1611`) both hold REQ-70…REQ-78, REQ-21 and REQ-79…REQ-83.<br>**Base structure:** the heading is at `AL:15`, the demoted paragraph at `AL:26`, and the next heading at `AL:28`. Only `AL:26` among the paragraph children lies under `AL:15`. The disagreement, prerequisites and escape-hatch paragraphs (`AL:9-13`) come before it.<br>**The union equals its children:**<br>• table rows `:254-259`;<br>• "AL table 1" `:385`;<br>• `:280`.<br>**The AL §1 parent is unchanged and complete:** `:270` and `"AL §1"` = REQ-21, REQ-23, REQ-45 and REQ-70…REQ-83, the union of `:271` and `:277-280`. | fixed |
| R12-03 | **`map:3` names all five exempt rows, and each name matches its Source cell exactly:**<br>• `:11`, "Design title / status / adoption and implementation chronology";<br>• `:269`, "AL Scout PR blocking allow-list at launch — council decision memo";<br>• `:339`, "req title / revision/status authority paragraphs";<br>• `:48`, "Design Appendix heading";<br>• `:342`, "req Appendix — Removed / superseded heading and provenance note".<br>**The wording is now accurate:** "map only their own text (the heading and the prose named in their Source)" covers the chronology, the authority paragraphs and the provenance note that R12-03 named.<br>**The union rule now states heading containment:** "rows whose source lies under the parent's heading in the base document".<br>**Nothing else restates the rule:** no other file restates the rule or the exempt set (grep, no hits). | fixed |
| Heading containment (lead, both peers) | **Every non-exempt heading parent and its JSON key holds the IDs of every row under its base heading.** Checked:<br>• design §1–§15 and all 21 subsections, with their non-goal, lane, risk, experiment, assumption, Phase, metric, history and rejected rows;<br>• Appendix items and sub-items 22(a)–(d) and 23(a)–(e);<br>• AL §1–§5, including the dropped children `:281,286,287`;<br>• `req §1`–`§10`, including `:340` (dropped) and `:341` (ADR);<br>• the Historical `req §11`–`§13` entries.<br>**Rows under the design H1 are covered by the title-row exemption:** the revision paragraphs and both evidence-notation rows (`D:5-21`).<br>**Metric rows sit under the right subsections:** REQ-103…REQ-111 fall under §8.2 and REQ-112…REQ-124 under §8.3. `fix_touches` is at `D:263`, before `D:265`. | pass |
| Multi-source rows | **Both rows propagate:**<br>• `:364` (ADR-009) reaches §10.3, §10, Appendix 6 and Appendix 12;<br>• `:384` (ADR-002) reaches §14 A8, §14 A10 and §14.<br>The rows and their JSON keys are unchanged since round 12. | pass |
| "Round 12 fixes" claims (`proposal.md:180-188`) | **Heading counts are correct:**<br>• 38 design = H1 + 15 sections + 21 subsections + Appendix;<br>• 7 allow-list;<br>• 12 requirements.<br>**The child-row breakdown adds up:** 84 = 22 metric rows (`:232-253`) + 15 AL table rows (`:254-268`) + 41 `req REQ-n` rows + 6 AL paragraph rows.<br>**The six paragraph rows are not identified in the proposal.** Every AL parent union checks by hand, whichever six they are (peer). | consistent |
| Counts | **Both peers independently recounted:**<br>• 389 map rows = 323 REQ, 44 ADR (the same row list as round 12) and 22 dropped (`:281,286,287,340,343-359,365`);<br>• 312 JSON keys, by parser and by raw grep, so there is no hidden duplicate key.<br>**Line shifts:** JSON lines move only after `"§15"` (+1) and inside `"AL §1 Eligible…"` (+6).<br>**History is intact:** `proposal.md:167` and `:170-178` are unchanged. | match |
| Conditions from round 12 | **Provided by the main agent:**<br>• digests;<br>• `readlink`;<br>• `git diff --cached --stat`;<br>• key and row counts;<br>• the union check output, now also using base heading containment. | met (main-agent outputs accepted) |

## Issues

None.

## Answers to the proposal's questions

1. **Is normative content lost, weakened or changed?** **No.**
   - The html is byte-identical to rounds 11–12, whose content was signed off in substance.
   - The challenger compared 15 more REQs that earlier rounds had not sampled with their base sources:
     - REQ-62…REQ-66 and REQ-69 (`html:556-560,563`) against `D:128-153`;
     - REQ-70…REQ-77 and REQ-79 (`html:592-599,601`) against `AL:13,19-24,32-37`.

     Triggers, capabilities, fixtures, provisional cost, exact check mechanisms, exemptions and overrides are preserved.
2. **Do the ADR summaries and cards match their records?** **Yes.** The html and every `ADR.md` are byte-identical to round 12, which verified the relation lines, round counts and sign-off digests.
3. **Does the Overview agree with Appendix A?** **Yes.** The challenger found no contradiction:
   - publishing, triggers and the mapping schedule (`html:198-330`) agree with REQ-13, REQ-26, REQ-27 and REQ-160;
   - decision ownership (`:340-368`) agrees with REQ-28 and REQ-139…REQ-154;
   - gate, activation and attribution (`:389-450`) agree with REQ-20…REQ-24 and REQ-78;
   - ledger records (`:455-509`) agree with REQ-5…REQ-9.
4. **Are rule, tooling and citation changes complete, with no live reference to a deleted file?** **Yes.**
   - No file outside the three changed files has changed since round 12, whose grep found only pinned or historical references.
   - The changed files name the base documents only as pinned provenance (`map:3`, "Authority: … revision 2.8").
   - Round 12's two citation-index gaps (R12-01, R12-02) are closed.
5. **Experiment rules, ledgers, frozen records:** **no violation.**
   - The experiment rules 1–7 in `AGENTS.md` match the base word for word, apart from the settled preamble (R1-14).
   - The ledgers and frozen digests are unchanged since round 12.
   - `round-12.md` is present and complete.

## Considered and not raised

- **ADR-005 and ADR-006 on the §15 parent and `"§15"`.** Extra IDs are not defects (`round-5.md:109-110`); round 12 made trimming optional, and `proposal.md:184` discloses it.
- **`proposal.md:184` names only the §15 parent for the kept extras, not the JSON key.** The key equals the row exactly.
- **`proposal.md:182` ("every other file are unchanged") does not mention the new `round-12.md`.** Round 12 accepted the same wording, and the 252-file `--stat` accounts for it.
- **"6 allow-list paragraph rows" (`proposal.md:188`) does not list the six.** It reports script output, and every AL union checks by hand whichever six they are.
- **JSON `"§14 A10"` holds REQ-38, which row `:157` does not.** It is an extra ID, it predates round 12, and it is not a defect.
- **Exempt rows lack their descendants' IDs.** This is now intentional and described accurately in `:3`. All non-exempt section and item mappings carry the IDs.

## Peers

- **`innovation-challenger`:**
  - confirmed R12-01…R12-03;
  - found no gap in any non-exempt heading parent or JSON key;
  - confirmed the 38/7/12 heading counts and the 84-row arithmetic;
  - sampled 15 new REQs for question 1;
  - confirmed question 3;
  - raised no issue.
- **`experiment-peer`:**
  - confirmed R12-01…R12-03 by hand union;
  - recounted 389 rows (323/44/22) and 312 keys, the keys by both parser and grep;
  - showed that every line shift comes from the two edited keys;
  - confirmed that the earlier proposal sections are untouched;
  - answered questions 2, 4 and 5;
  - raised no issue, with four observations, all recorded above.
- **Lead:**
  - separately checked every base heading boundary and the containment of all design, AL and `req` parents;
  - checked the five exempt names against their rows;
  - checked `D:5-21` and `D:263`;
  - grepped the unchanged files for any restatement of the rule or counts;
  - found no issue.
- **Where the peers changed the outcome:** they did not. All three analyses independently reached zero issues. The peer's raw-grep key count and the challenger's 15-REQ sample widened the evidence beyond what the lead checked alone.

## Verdict

**0 issues: SIGNED OFF.**

### Conditions

The main agent runs these before commit. Any mismatch voids this sign-off and needs a new round.

1. `shasum -a 256` on each of the 18 files in the table above, on the staged worktree. Each output must equal the digest listed.
2. `test -L CLAUDE.md && readlink CLAUDE.md` must print `AGENTS.md`.
3. `git -C /Users/adamjackson/Projects/work/scout-req diff --cached --stat` must show 252 files changed plus this `round-13.md`, with no other file added or changed. That is 253 files once this record is staged.
4. The council did not hash any file. This sign-off rests on the main agent's recomputation, as in earlier rounds.

### Sign-off

The innovation council signs off ADR-009, the design restructure, with zero issues. This covers the following files at these SHA-256 digests in `/Users/adamjackson/Projects/work/scout-req` (branch `agent/requirements-path`, staged), against base `main` `9a3d19f5`:

- `docs/scout-design.html` `f0aa226a88c3c8eda782a7327be85b2bccb5a879df177673d12021d50ffdc463`
- `docs/adrs/2026-11-06-design-restructure/proposal.md` `7135fe5d3153285265a7d61cf1da274929276206053950eefa8bccb57b00912f`
- `docs/adrs/2026-11-06-design-restructure/requirements-map.md` `9a8db8aa2df66475cd0665f7c03e1af64ee83216f64b804491f0eab4522ded53`
- `docs/adrs/2026-11-06-design-restructure/design-section-map.json` `ad9dde4b0e6e99937b744a3bffcb40c68dfc84d166b6bb74960f268dfd853bb9`
- `docs/adrs/2026-11-06-design-restructure/ADR.md` (ADR-009) `1393b6eb7488688c24ef25106ec6861616ae61e75675570651e5b97ab1fae83a`
- `docs/adrs/2026-11-06-design-restructure/completeness-audit.md` `0fa44fa63e084d2ea9b0690fc5bd159037cb526980a736e99da621ea9505c31b`
- `AGENTS.md` (`CLAUDE.md` → `AGENTS.md` symlink) `aff517d34b71ac2ed189dcc9f4d3b03380e9b91e9b754a2bd98aceeaa4718511`
- `docs/adrs/README.md` `7c90063cd3050dd05984990dade0553e47e4fc7d43051b9e9c92f241c2e307e0`
- `docs/jev-adapter-audit.md` `30894dbf17bf9f815ceaadce2c3c9ae75d2cf0e31d4a91753985dbfe6818630d`
- `docs/adrs/2026-10-19-v1-decisions/ADR.md` `4d631ce92da11fc7c42467d2b18cc12902f4e193d92170a7e5e33a9a8547b36f`
- `docs/adrs/2026-10-21-e1b-v2/ADR.md` `e5db1e7659fb562bd3c89cb36483d789d0e669d3208c8f2a9e0617346d2be8fa`
- `docs/adrs/2026-10-22-rescope/ADR.md` `4adb8f70b90d2ca4deb941b56d161ccce424549f34c323b53cd14c4c80a9ce2b`
- `docs/adrs/2026-10-23-decisions/ADR.md` `f42a2ea88e1b02d9a41b558841f32c7393a02cce58f7ccdd0c401caa679fdfe7`
- `docs/adrs/2026-10-23-clawpatch-patch-v3/ADR.md` `cbaf321eed0957fa29c2ac7bca02b981f3acd8c8efc895e10bc7a1bf685e577a`
- `docs/adrs/2026-10-23-codex-luna-mapping/ADR.md` `ff8a06e034a20dc0242732a71743b7f99c1cc82d1a5293eab6d29a272e34e00c`
- `docs/adrs/2026-10-24-area-refresh-fallback/ADR.md` `371a870bd4302d5bf3a6f224679b6ac2ad321a0c3b1dd4aeb6e70ce4941e9a82`
- `docs/adrs/2026-10-24-cutover-cleanup/ADR.md` `de9fffdf63fc0ec61a98bdde7327617682d7b336b1199711537d48b26ee04721`
- `docs/adrs/2026-10-31-jev-merge-association/ADR.md` `76dbd3aec2503d447270a91bdfb9a9c44e35ca730c45199add44a34847c45ab0`

These files are frozen at these digests. Any later change, however small, needs a new council round with zero issues before adoption.

---

## Council memo

1. **Recommendation:** adopt ADR-009 at the digests above.
   - Stage this record as `round-13.md`.
   - Run Conditions 1–3 and commit.
   - Make no further edits to the signed files.
2. **Alternatives:**
   - (a) **Trim ADR-005 and ADR-006 from §15 and `"§15"` first.** This is more precise, but it changes a signed file and would need round 14. It is not needed for completeness.
   - (b) **Ask the proposal to list the six AL paragraph rows the script located.** It would be more transparent, but the hand unions already make this output non-load-bearing. Also not needed.
   - (c) **Another sampling round on the html.** Not justified: the html has been byte-stable since round 11, and rounds 9–13 have sampled well over 50 REQs against the base.
3. **Evidence:**
   - Base: `D:1-21,261-265,468-476`; `AL:1,5,7-28,42,49,59`; `SCOUT_REBUILD_REQUIREMENTS.md` headings `:1-79`; `2026-10-19-v1-decisions/ADR.md:5,20-21`.
   - Map: `:3,11,43,47,48,162,232-287,339-342,360-389`.
   - JSON: `:433-451,499-506,1047-1054,1595-1611`.
   - Proposal: `:167,170-188`.
   - Precedents: `round-12.md` (issue table and required items); `round-5.md:109-110`.
4. **Risks:**
   - **Residual:** the council relies on the main agent's digests and diff stat. Conditions 1–3 limit this.
   - **Unknown:** the union-check script itself was not reviewed. The hand unions by the lead and both peers independently cover every parent touched since round 11.
   - **Reversibility:** every artefact is text or JSON under version control, and the restructure is fully reversible by revert.
5. **Execution outline:**
   1. Write this record to `docs/adrs/2026-11-06-design-restructure/round-13.md` and stage it.
   2. Run `shasum -a 256` on the 18 files and compare each digest with the table.
   3. Run `test -L CLAUDE.md && readlink CLAUDE.md` and expect `AGENTS.md`.
   4. Run `git diff --cached --stat` and expect 253 files, with `round-13.md` the only addition since the reported 252.
   5. Commit. Mark ADR-009 adopted only through files that are not in the signed set, because the signed files are frozen.
