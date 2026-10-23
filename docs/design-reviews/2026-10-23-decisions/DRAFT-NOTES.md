# Draft notes — revision 2.3 proposal

Round-7 draft (round-6 issue addressed, `round-6-response.md`; round 5 in `round-5-response.md`; round 4 in `round-4-response.md`; round 3 in `round-3-response.md`; round 2 in `round-2-response.md`; round 1 in `round-1-response.md`). Not yet signed off. Line numbers refer to `scout-design.proposed.md`. `scout-design.proposed.md`, `proposed.diff` and `requirements.proposed.diff` are unchanged since the round-6 draft (round 6 edited only the labels of this file and `PROPOSAL.md`).

| File | Lines | SHA-256 |
|---|---|---|
| `scout-design.base.md` (byte copy of `docs/scout-design.md`, rev 2.2) | 462 | `f81b26014e6d2e38a741249819d23fcb993541e2194f7b168c4970c3b415ded4` |
| `scout-design.proposed.md` (rev 2.3, round-6 and round-7 drafts) | 466 | `75ef1e66521b1f129e872f651bd0f8bb9989e38867fb5335ec5ad43158fb3061` |
| `proposed.diff` (`diff -u base proposed`) | 236 | 42 insertions, 38 deletions; `73da0f865fb72158ee5ee3e367b6b5dfcb7f8e0945c2e10144f121145c1eb6c1` |
| `requirements.proposed.diff` (`diff -u`, title/status header, REQ-25, 27, 29, 35, 36 and the open-decisions line) | 49 | `4e7adaccaa3ffe0e627fdd06c068a8469afbefc5a6dd09d985ca73803ab927c3` |
| `PROPOSAL.md` | — | `ae081bf7208b3cadecdd2fc241b699b5978a4c18cc2ad0c70fc75cf6032724b3` |

`docs/scout-design.md` and `SCOUT_REBUILD_REQUIREMENTS.md` are untouched. The base digest equals the round-6 sign-off digest of revision 2.2.

## Where each change is (proposed file; line numbers unchanged since the round-3 draft)

- Header: line 3 status/phase sentence; Revision 2.3 paragraph line 7. (D8, all)
- D1: §2 note line 45; §8.5 Consequences line 271; §8.6 line 281; §15 lines 438–440 and 442 (former items 1–3 now Appendix 16–18); Appendix heading line 446 and items 16–18 (lines 463–465).
- D2: §6 version-key bullet 155, Pinned model bullet 156, Cap bullet 159 (cost cap); Launch default paragraph 166; R14 line 401; Appendix 19 (line 466).
- D3: §6 Labels line 168; X1/X3/X4/X5/X10 rows (§13; "§6 Labels protocol"); Appendix 15 (line 462).
- D4: §7.2 line 182 (comparability conditions, manifest `uninspected`); §7.3 Attribution bullet line 186 and shadow bullet line 189; §7.4 tail line 197 (also the CI-toolchain sentence); Phase 2 acceptance (line 360); A2 line 425 (§14).
- D5: Phase 3 scenario line 363.
- D6: Phase 4 acceptance (line 372); the §11 `[ASSUMPTION, Xn]` sentence is removed.
- D7: §7.4 line 197; §10.2 workflows row; Phase 2 heading; Phase 6 scenario line 379.
- D8 other: §3.3 first sentence; §4 Areas and Active snapshot membership bullets; §10.2 clawpatch row (see PROPOSAL.md table with evidence).

## Round-4 judgement calls (round-3 issues; see `round-3-response.md`)

Issue 1 qualifier (otherwise-`introduced`, non-`guardrails.*`); issue 2 option (a); issue 4 release ancestry confirmed (`480949a`, `404f69b` in `v0.1.2`, not in `v0.1.1`); issue 5 qualified, not re-run (header now names v0.1.2 and run 39966674494).

## Round-3 judgement calls (choices the council left open)

- Issue 1: the design now describes the attribution that shipped in `480949a` (coverage-aware `unknown`, manifest `uninspected`) and the CI toolchain provisioning that shipped in `404f69b`; the round-2 "known gap, outside this revision" text is deleted. Counting unit of the ≤ 10 % rule: PR runs, one per run, that are `comparison: none` or carry at least one base-gap `unknown` finding, over all PR runs in the window (`ShadowStats` has no producer yet; stated in §7.3). Recomputed runs, including gapped ones, remain comparable-base runs for the ≥ 5 PR-run minimum.
- Issue 4: option (a): the Phase 4 scenario reports area-ID retention and Jaccard (text-only; `scout map` discards the report at `cmd/scout/map.go:102`); one retention ratio (named distinctly from X2's matched-area survival, HD5, round-4 issue 1) and one distribution of the retained areas' Jaccard, matched and kept pooled; the matched/kept split is X2's protocol, not a Phase 4 criterion (round-3 issue 2 option (a)).
- Issue 8: the `[ASSUMPTION, Xn]` sentence and the item (4) clause are dropped.
- Issue 5: one scope list: defaults substitute only for precision/recall point minimums (X3, X10, point values in X5); X1's gain threshold and the lower-bound forms (X4, X5) have no default substitute.

## Round-5 judgement calls (round-4 issues; see `round-4-response.md`)

- Issue 1: the Phase 4 ratio is named "area-ID retention" (counts kept areas) in line 3, line 7 and line 372; "survival" is left to X2/A1 and HD5 (matched areas only). Line 372 states the difference and cites HD5. X2's rule, A1 and HD5 are unchanged.
- Issue 2: "PR runs" at lines 7, 189 and 425, matching line 186.
- Issue 3: line 186 says `comparison` is not persisted in Scout's state, each run writes it to `scout-assessment.json` and, on PR runs, the Markdown summary; the reusable workflows upload both as run artifacts.
- Issue 4: "(Linux runners)" at line 197; PROPOSAL D8 mirrors it.
- Issue 5: PROPOSAL D6 "does not separate"; the item above names one retention ratio and one pooled Jaccard distribution.

## Judgement calls for the council (round 1; see round-1.md §5 for rulings)

1. §15 is kept with one remaining item (Jev costs and minimums, still unset by the user) rather than declared empty; the X rows' "pending the user's minimums" wording stays consistent with it.
2. "Raised to 0.9 on the costlier side" is applied to triage recall, following `docs/jev-adapter-audit.md` row 22 (precision 0.8, recall 0.9); other questions get the same rule when they are added.
3. Appendix 15 is rewritten in place (it recorded the human-review step that STATUS §6.9 item 6 supersedes), not left beside a new item. Appendix 16–18 name the old items as "rev 2.2 §15 item N" because "§15.3" already means the performance bar in Appendix 14.
4. The 52 E2 labels the user already confirmed keep provenance `human` (STATUS §1 "Jev labels").
5. Phase 4's new criterion cites X2's own metric; the observed 0.95/0.85 is not asserted as an X2 result.
6. Labelling text references "§6 Labels protocol" instead of STATUS §6.8 so the design no longer points at the superseded protocol.

## Checks for the main agent

`shasum -a 256 docs/scout-design.md docs/design-reviews/2026-10-23-decisions/scout-design.base.md` must match `f81b2601…`; `diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md` must reproduce `proposed.diff` byte for byte (no timestamps); the council's final round must name the proposed file's SHA-256 (any later edit changes it). On sign-off with a matching digest: copy proposed to `docs/scout-design.md`, apply `requirements.proposed.diff`, update STATUS §2/§5 row 4b and the ledgers (R-039 already recorded). STATUS §1 was updated in this round for the attribution fix and the CI toolchain.
