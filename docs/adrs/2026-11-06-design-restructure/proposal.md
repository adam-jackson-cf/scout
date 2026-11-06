# ADR-009 proposal: one design document with requirements and ADR appendices

Date: 2026-11-06. Driver: user direction (R-062). Base: `main` at `9a3d19f538d26085179f385ef21dcb39942c3bb5`, design revision 2.8 (`docs/scout-design.md` SHA-256 `d1fb127446be290f75d4de5416ebf40b214552f18eedc1d1a335819f314f9b75`, signed in `docs/adrs/2026-10-31-jev-merge-association/round-11.md`). Worktree under review: `/Users/adamjackson/Projects/work/scout-req` (branch `agent/requirements-path`). The base files are readable unchanged in the main checkout `/Users/adamjackson/Projects/work/scout/` (at the base commit).

## Change set

1. **New governed design:** `docs/scout-design.html`. It has three tabs:
   - Overview: run flow, who decides what, gate decision, ledger records.
   - Appendix A: requirements `REQ-1`…`REQ-183`.
   - Appendix B: ADR cards `ADR-000`…`ADR-009`.
2. **Deleted** (git keeps them): `docs/scout-design.md`, `SCOUT_REBUILD_REQUIREMENTS.md`, `docs/scout-allowlist.md`, `docs/archive/scout-design-v1.md`, `docs/scout-design-brief.md`, `docs/inertia-research.md`, `docs/enaible-findings.md`.
3. **Requirements:**
   - `REQ-1`…`REQ-41` keep their IDs. Their meaning is unchanged except REQ-37 (see "Decisions this ADR makes"). Where the design was more current, the text follows the design.
   - `REQ-42`…`REQ-183` are new and carry the design's and the allow-list's remaining normative content. None is withdrawn.
   - `requirements-map.md` maps every design heading and subsection, every Appendix item, every allow-list section and every REQ to a REQ, an ADR or a listed drop with its reason: 389 rows, of which 323 go to a REQ, 44 to an ADR and 22 are dropped.
4. **ADRs:**
   - `docs/design-reviews/` is renamed `docs/adrs/`. Folder and file names inside it do not change.
   - Each folder gains an `ADR.md` summary: ADR-001 to ADR-008, one per signed revision 2.1–2.8.
   - ADR-000 (`2026-10-19-v1-decisions/ADR.md`) records the v1 user decisions, Appendix items 1–7, with no council round.
   - ADR-009 (`2026-11-06-design-restructure/ADR.md`) records this change.
   - `docs/adrs/README.md` indexes the ADRs and maps old paths to new ones.
5. **Citations:** `design-section-map.json` (312 keys) drove the rewrite of every `design §n`, `Appendix N`, `AL §n` and `req §n` citation. The rewrite covers code comments (`internal/**`, `cmd/scout/**`), the two guardian workflows, `toolchain/versions.env`, `docs/references/*`, `docs/plans/*`, `docs/jev-adapter-audit.md`, `README.md`, `CONTRIBUTING.md`, `docs/STATUS.md` and `AGENTS.md`. Only comments, documentation strings and the `--matrix` help citation change. `go build` and `go vet` pass.
6. **Rules and tooling:**
   - `AGENTS.md` design rule 1 names `docs/scout-design.html` and the documents it governs: `docs/adrs/*/ADR.md`, the Met/Fixed status and Design-level rows of `docs/jev-adapter-audit.md`, and the future `docs/evidence/study-export.md`. Rule 2 defines ADR folders with an `ADR.md` summary, sequential IDs, and supersession recorded on the older ADR's status line.
   - `.omp/gate.json` (sensitive list, now also `jev-adapter-audit.md`), `.omp/ledger.d/c.json` and the regenerated `.omp/coverage-ledger.json`, `.prettierignore`, `scripts/guardrails/infrastructure.py` and the gate-contract test now name the new paths. `docs/archive/` exclusions are removed.
7. **Ledgers (append-only):** research ledger R-062; a hypothesis-ledger History line recording that no hypothesis changes state and that earlier `docs/design-reviews/<folder>/…` paths resolve to `docs/adrs/<folder>/…`.
8. **STATUS:** the revision-history paragraphs, §4 (inertia-bench move and subject-interface notes; now ADR-002, REQ-37 and REQ-135) and §6 (user decisions of 2026-10-22 and 2026-10-23; now ADR-002 and ADR-003) are removed.

## Decisions this ADR makes

These are changes, not moves:

- It ends the Phase 6 spec cutover of ADR-000 item 6 and ADR-002 Appendix 12. No separate product-contract or adapter-contracts documents are written. The requirements file is retired now.
- It deletes `docs/enaible-findings.md` instead of moving it to `docs/evidence/`.
- REQ-37 follows design §8.7's principles-only subject interface. The old requirement's exact schema, cap and sort order are not adopted.
- REQ-42 adopts design §13's experiment list (X1–X6, X9–X11).
- An accepted ADR's status line may be updated when a later ADR supersedes or amends it.

Every other requirement keeps its meaning. ADR-009's `ADR.md` points to this folder's final round for the signed digests, so the summary does not hash itself.

## Not changed

- Frozen evaluation records, signed review rounds, proposed-design snapshots and earlier ledger rows keep their old `docs/design-reviews/…` paths. `docs/adrs/README.md` maps them and says which revision a frozen citation of a deleted file means.
- No product code behaviour changes.

## Round 1 fixes

`round-1.md` raised 17 issues. Each is addressed:

- R1-01 to R1-06: the requirements named there are restored in Appendix A.
- R1-07: ADR-009 reads `accepted` in the card, the README index and STATUS, and the whole change lands only after sign-off. Appendix B and the README say that ADR-000 has no council round.
- R1-08: ADR-009's `ADR.md` is added.
- R1-09: design rule 1 is restored with an explicit list of governed documents.
- R1-10: this section; the ADR-009 card and `ADR.md` list the decisions.
- R1-11: supersession is recorded in both directions on the cards and the `ADR.md` status lines, and the status-line exception is written into rule 2.
- R1-12: the Overview's attribution popover, exit-3 wording, nightly trigger, LLM column, activation callout and record count are aligned with Appendix A.
- R1-13: the header states that Appendix A prevails.
- R1-14 and R1-15: `AGENTS.md` uses X1–X6, X9–X11 and cites R-039/ADR-003.
- R1-16: REQ-110 and REQ-111 are repointed; `ADR.md` citations are pinned to `9a3d19f5`; the stale plan lines are removed.
- R1-17: the README path map covers deleted files and section citations.

## Round 2 fixes

`round-2.md` raised 5 minor issues. Each is addressed:

- R2-01: rule 2 says an `ADR.md` that is among the signed files names the final round for its outcome and digests. ADR-009's `ADR.md`, STATUS and CONTRIBUTING point to `round-3.md`.
- R2-02: the activation callout counts `unknown` over all PR runs in the window. The ledger diagram's block arrow reads "only when every gate condition above holds (guardian profile)".
- R2-03: digest inventories in the `ADR.md` files name each artifact as its round named it. `@9a3d19f5` pins remain only where the content is the state at that commit. ADR-001 cites the e1b-v2 proposed snapshot for §15 item 5.
- R2-04: proposal items 3 and 4 are corrected. The ADR-009 card says "apart from the decisions above" and lists the status-line exception.
- R2-05: REQ-37 cites ADR-009, and REQ-42 cites ADR-008 and ADR-009. ADR-008 amends ADR-002 by adding X11; both `ADR.md` files and both cards record it. ADR-009's relation to ADR-002 names the experiment list.

## Round 3 fixes

`round-3.md` raised 7 issues (1 major). Each is addressed:

- R3-01: REQ-157 counts the fallback rate whether or not the fallback ran. REQ-158 evaluates the fallback rate over the whole window once it holds `min_clusters` residual clusters, and applies the eligible-merge and fallback-answered conditions to the pending-new rate only.
- R3-02: REQ-176 restores "chosen by a model-free proxy fixed before labelling". REQ-176 and REQ-177 cite ADR-008.
- R3-03: allow-list rows 3–6 map to REQ-72/73, REQ-74/73, REQ-75 and REQ-76; the non-existent row 7 is removed. The map now has 388 rows (323 REQ, 43 ADR, 22 dropped).
- R3-04: ADR-009's `ADR.md` names the experiment list in its ADR-002 relation. ADR-002's Changes record ADR-008's X11 amendment and ADR-009's experiment-list change.
- R3-05: ADR-004 cites `scout-design.base.md` (revision 2.3) §4 for the corrected timeout symbol. ADR-005 lists the requirements base digest `cd84592f…` that round 6 names. `docs/adrs/README.md` says a frozen record's citation means the revision current when it was written.
- R3-06: `docs/jev-adapter-audit.md` joins the reviewed files. Its edits are citation-only (rows 28 and 30: `design §6` → REQ-96 and REQ-86, REQ-99), plus line 7 now names the design file as it was then.
- R3-07: ADR-009's `ADR.md`, STATUS and CONTRIBUTING point to the zero-issue final round without a round number.

## Round 4 fixes

`round-4.md` raised 4 minor issues. Each is addressed:

- R4-01: REQ-60 adds ADR-005 and ADR-008; REQ-167 adds ADR-006; REQ-170 adds ADR-005 and ADR-006; REQ-171 adds ADR-008. A pass over the Changes of ADR-001…ADR-008 also corrected REQ-2 (+ADR-003), REQ-16, REQ-18, REQ-30, REQ-39 (+ADR-002), REQ-52 (+ADR-005), REQ-54 (+ADR-001), REQ-56 (+ADR-004), REQ-78 (+ADR-003), REQ-117 (+ADR-003), REQ-124 (+ADR-002), REQ-132 and REQ-134 (+ADR-001), and REQ-168, REQ-181 and REQ-182 (+ADR-002). That pass worked from summaries and missed rows; round 5's diff-driven pass completes it. Requirement text is unchanged.
- R4-02: the remaining legacy citations are repointed. `docs/jev-adapter-audit.md` lines 19 and 79 cite REQ-87 (row 38 also REQ-99). Comment-only edits: `internal/decisions/client.go:20` → REQ-87, REQ-92; `internal/policy/policy.go:330` → REQ-70–REQ-76; `internal/guardrails/focused_tests.go:33` → REQ-76.
- R4-03: ADR-003 cites `docs/STATUS.md`@9a3d19f5 §1, revision 2.3 history entry, as its adoption record.
- R4-04: the ADR-009 card and "Not changed" above say that `docs/adrs/README.md` gives which revision a frozen citation of a deleted file means.
- `CLAUDE.md` is a symlink to `AGENTS.md`.

## Round 5 fixes

`round-5.md` raised 3 issues (2 minor, 1 nit). Each is addressed:

- R5-01: REQ-35, REQ-41 and REQ-160 add ADR-002; REQ-49 and REQ-55 add ADR-003; REQ-56 adds ADR-005. A clause-level pass then read every `proposed.diff` and `requirements.proposed.diff` of ADR-001…ADR-007, and ADR-008's `requirements.proposed.diff` plus its design changes against the ADR-007 snapshot. It traced each changed normative sentence to its REQs and found no further missing citation.
- R5-02 content restorations (IDs unchanged): REQ-62 names the artifacts lane's LCOV, Cobertura and JUnit inputs plus PR metadata, revision/producer matched, `unmeasured` when absent. REQ-136 measures actionable yield per lane/rule. REQ-92 gives every listed failure cause no annotation and `ranking: deterministic (reason)`, and keeps `ranking: partial (n/N)` for caps. REQ-35 says the qualifying study may be signed off under Scout's experiment rules or inertia-bench's rules. Comment-only: `internal/artifacts/artifacts.go:2` adds REQ-62.
- R5-03: change-set item 8 names STATUS §4 and §6 by content.

## Round 6 fixes

`round-6.md` raised 3 issues (1 minor, 2 nits). Each is addressed:

- R6-01 content restoration: REQ-92 adds "any other non-success response is likewise a failed unit" (base design §6). The ADR column is unchanged.
- R6-02: change-set item 8 sends STATUS §6 to ADR-002 and ADR-003.
- R6-03: change-set item 7 describes the hypothesis-ledger History line without quoting it.

## Round 7 fixes

`round-7.md` raised 2 minor issues. Both are content restorations; IDs and ADR columns are unchanged:

- R7-01: REQ-174 replays "the last 200 merged PRs".
- R7-02: REQ-16 adds that a required blocking lane that could not run on head gives exit 2 and that optional lanes never turn a required lane red. REQ-9 adds that Coverage gates only as "required lane incomplete" (exit 2).
- Sweep for the same class: every base design, allow-list and requirements sentence with a binding qualifier (exit codes, never, only, must, at least/at most, last, fresh, blind, before, until, numeric thresholds, sampling frames) was traced to its mapped REQ. It restored eight further qualifiers:
  - REQ-33: the measures never gate, set severity, forecast or compute a composite (base §8.1).
  - REQ-87: response-model mismatches are discarded and counted (base §6).
  - REQ-96: buckets wait until X1's preregistration is signed off (base §6).
  - REQ-169: X1 compares triage retrieval at a fixed review budget (base §13).
  - REQ-170: X2/X2b use the E3 protocol with Scout's area reconciliation, 3 fresh maps, 2 same-state maps and 2 mutations; X2b repeats the same repositories, revisions, schedule, grid and guide, codex arm only (base §13).
  - REQ-174: X9 hand-checks its 400 labels on the corpus (base §13).
  - REQ-176: preregistration comes before any X11 run, and thresholds, prompt and guard are frozen before data (base §13).
  - REQ-178: the Phase 1 scenario's second run follows the 20-line insertion (base §11).
- `requirements-map.md` adds REQ-16 to the §3.2, §3.2 Coverage, §7 and §7.3 rows and REQ-33 to §8.1. Row counts are unchanged.

## Round 8 fixes

`round-8.md` raised 6 issues (5 minor, 1 nit). Each is addressed; IDs and ADR columns are unchanged:

- R8-01: `design-section-map.json` adds REQ-16 to `§3.2`, `§3.2 Coverage`, `§7` and `§7.3`, and REQ-33 to `§8.1`. It still has 312 keys, and every `Design §…` row of `requirements-map.md` with a REQ disposition now equals its JSON key.
- R8-02: REQ-178 duplicates "a same-rule finding in one file".
- R8-03: REQ-96 cuts the bucket "by rank position within the run".
- R8-04: REQ-169 takes X1's labels from the REQ-100/REQ-101 protocol on probability-sampled audit units.
- R8-05: REQ-177 reports duplicate pending areas per real new concern and the share of wrong assignments among clusters Jev assigned without escalation.
- R8-06: REQ-170 and this proposal say "the E3 protocol with Scout's area reconciliation".
- Phrase-level audit, as the round-8 memo recommended. Seven audit slices (five initial and two gap slices) covered the whole base design, the allow-list and old REQ-1…REQ-41. Each candidate was verified against its base line and all of Appendix A. `completeness-audit.md` in this folder records every decision; its totals are corrected in round 9 (see below). Restored rows: REQ-8, REQ-13, REQ-21, REQ-22, REQ-25, REQ-26, REQ-27, REQ-28, REQ-30, REQ-33, REQ-41, REQ-44, REQ-45, REQ-61, REQ-90, REQ-101, REQ-134, REQ-138, REQ-139, REQ-152, REQ-160, REQ-165, REQ-167, REQ-180 and REQ-182.
- ADR-009's `ADR.md` lists `completeness-audit.md` under "Records in this folder".

## Round 9 fixes

`round-9.md` raised 9 issues (7 minor, 2 nits). Each is addressed:

- R9-01 to R9-05, and the memo's method change: every base §11 Scenario/Acceptance line and every §13 Question/Design/Decision-rule cell is now carried near-verbatim in its REQ, leaving out only past-run results, dates and evidence pointers. Rows rewritten: REQ-167, REQ-168, REQ-169 (the D:435 preregistration list word for word), REQ-170, REQ-171 (precision denominator, retrieval inputs, split and interval, rev-2.1 volume of 200), REQ-172, REQ-173 (volume 200), REQ-174, REQ-175 (volume 200), REQ-176 and REQ-177 (agent-protocol labels), and REQ-178 to REQ-182. REQ-89 adds that X1's preregistration fixes how Jev-quality statistics are keyed to the rank and baseline versions. REQ-91 restores the baseline-formula ordering rule (D:177). REQ-180 adds ADR-001 (active-hours acceptance, e1b-v2 `proposed.diff:142`); no other ADR column changes.
- R9-06: REQ-26 no longer requires 90-day artifact retention. `completeness-audit.md` reclassifies that candidate as overreach.
- R9-07: every parent row of `requirements-map.md` now contains the IDs of all its child rows, ADR IDs included (19 parents extended, among them §3.2, §8.5, Appendix 22 and AL §1). `design-section-map.json` follows. The round-9 script paired keys only by name prefix, so it missed the `req §n` sections, the `dropped` children and multi-row JSON keys; round 10 found those gaps (R10-08 to R10-10), and "Round 10 fixes" records the extended check.
- R9-08: the Overview says trusted default-branch runs (push, nightly or dispatch) publish the baseline, in all three places.
- R9-09: `completeness-audit.md` corrects the reasons for candidates 25 and 28, explains that candidate numbers are internal labels of the audit run, uses "seven audit slices", and reports totals with and without duplicates: 72 candidates (64 distinct): 21 lost (16 distinct), 19 loosened (17), 1 overreach, 27 already carried (26), 1 not normative, 3 ADR or dropped. It appends a clause-by-clause inventory for base §6, §11 and §13.

## Round 10 fixes

`round-10.md` raised 11 issues (7 minor, 4 nits). Each is addressed:

- R10-01: REQ-178 cites ADR-007.
- R10-02: REQ-178 cites "the REQ-165 benchmark and the REQ-38 pinned fastapi assess" and "REQ-165 and REQ-38 criteria" instead of §10.1.
- R10-03: REQ-176 names "the carried REQ-146 policy thresholds and REQ-159 guard defaults".
- R10-04: REQ-88 adds "and supply them as facts in state".
- R10-05: REQ-91 adds "(rubric functions; commit and unowned-file units once X5/X3 admit them)".
- R10-06: REQ-167 adds "and recorded in `docs/STATUS.md` §1".
- R10-07: REQ-168 cites ADR-009. The map rows `§11 Phase 6`, its Scenario and Acceptance rows, and the `§11` parent add ADR-009, and so do their JSON keys.
- R10-08 to R10-10: the union rule now covers three more relations: each `req §n` row contains the IDs of the `req REQ-n` rows of the requirements it held; a parent contains the IDs named by its `dropped` children; and a JSON key contains the IDs of every map row it represents, matched without case (so "Evidence notation" now holds ADR-009, REQ-43, ADR-001, ADR-002 and ADR-008). Map line 3 states the rule. AL §5 also adds REQ-81 and REQ-183. Eleven map parents and 15 JSON keys were extended.
- R10-11: `completeness-audit.md` corrects the carriers the round named. A re-check of all 331 inventory rows against the current HTML corrected 64 carriers or quotes. It found no further normative loss; two items are marked non-normative (a tooling observation at D:170.5 and the evidence pointer at D:378.10).
- ADR column: all 52 rows rewritten or restored in rounds 5–10 were traced clause by clause to the pre-2.1 base or to the frozen `proposed.diff` and `requirements.proposed.diff` that introduced each clause. 13 rows gain 14 citations: ADR-000 on REQ-13, REQ-25, REQ-27, REQ-35, REQ-49, REQ-55, REQ-56, REQ-61, REQ-87, REQ-88 and REQ-152; ADR-002 on REQ-25 and REQ-178; ADR-005 on REQ-167. No citation is removed. `completeness-audit.md` ("ADR-column check, round 10") lists each added ADR with the clause and diff line.

Mechanical checks run on the submitted files (script output):

- Union check: 0 map parents missing a child ID (`req §n` ← `req REQ-n` pairs: 41; the round-10 submission reported 43 because the script also counted two non-`req REQ-n` children; `dropped` targets included); 0 JSON keys missing an ID of their map rows or child keys; 312 JSON keys; 388 map rows (323 REQ, 43 ADR, 22 dropped).
- Cross-reference check: 183 REQ rows; every `REQ-n` and `REQ-a–REQ-b` reference inside a row names an existing REQ; every ADR ID is ADR-000 to ADR-009. No design `§` reference remains. Five `§` references remain and name other documents: REQ-110 (`docs/inertia-research.md` §5 at `9a3d19f5`), REQ-111 and REQ-117 (E1b preregistration §3.1, §3.3) and REQ-167 (`docs/STATUS.md` §1).

## Round 11 fixes

`round-11.md` raised 3 nits. Only `requirements-map.md`, `design-section-map.json` and this proposal change; the HTML and the audit are unchanged.

- R11-01: the two-source row "Design §15 former numbered decisions / Appendix 20 pending study clause" is split into "Design §15 former numbered decisions" (ADR-002, ADR-003) and "Design Appendix 20 pending study clause" (ADR-005, ADR-006). The "Design Appendix 20" row and JSON key now also hold REQ-53 (the retained 0.5 default), REQ-170 (X2b) and ADR-006 (which closed the pending clause). Map line 3 states that a row whose source names several sections adds its IDs to each of them; the two remaining such rows ("§10.3 / Appendix 6,12 …" and "§14 A8 and A10") already satisfy it.
- R11-02: the round-10 union-check line now reports 41 `req §n` ← `req REQ-n` pairs.
- R11-03: map line 3 states that the document-title rows, "Design Appendix heading" and "req Appendix — Removed / superseded heading" map only their own heading text, and that the generic range provenance in the removed-item rows is not a union target.

Mechanical checks on the submitted files (script output): 0 map parents missing a child ID; 41 `req §n` ← `req REQ-n` pairs; 0 gaps for multi-source rows against each named section's row and JSON key; 0 JSON keys missing an ID of their map rows or child keys; 312 JSON keys; 389 map rows (323 REQ, 44 ADR, 22 dropped), one more ADR row than before because of the R11-01 split.

## Round 12 fixes

`round-12.md` raised 3 nits. Only `requirements-map.md`, `design-section-map.json` and this proposal change; the HTML, the audit and every other file are unchanged.

- R12-01: "Design §15 former numbered decisions", the §15 parent and JSON key `§15` add ADR-000 (the v1 §15 decisions). ADR-005 and ADR-006 stay on the §15 parent; extra IDs are not defects.
- R12-02: "AL §1 Eligible blocking checks" and its JSON key add REQ-21 and REQ-79 to REQ-83 from its child "AL §1 demoted seed checks" (`AL:26`, under the `AL:15` heading).
- R12-03: map line 3 names the exempt rows (three document-title rows, "Design Appendix heading" and "req Appendix — Removed / superseded heading and provenance note"), says they map only their own text, and adds heading containment to the union rule.

Mechanical checks on the submitted files (script output): the union check now also pairs rows by heading containment in the three base documents. It parsed 38 design, 7 allow-list and 12 requirements headings, located 84 child rows by base line (22 metric rows, 15 allow-list rows, 41 `req REQ-n` rows and 6 allow-list paragraph rows), and found one gap before the R12-02 fix and 0 after. Name-based pairs: 0 gaps. Multi-source rows: 0 gaps. JSON: 0 keys missing an ID of their map rows or child keys; 312 keys. Map: 389 rows (323 REQ, 44 ADR, 22 dropped), unchanged from round 11.

## Questions for the council

1. Is any normative statement of the base design, requirements or allow-list lost, weakened or changed in meaning? Check against `requirements-map.md` and the base files.
2. Do the ADR summaries and cards match their review records: decisions, numbers, digests and supersession?
3. Do the Overview diagrams and lists agree with Appendix A?
4. Are the rule, tooling and citation changes complete, with no live reference to a deleted file?
5. Does anything violate the experiment rules, the append-only ledgers or the frozen-record rule?
