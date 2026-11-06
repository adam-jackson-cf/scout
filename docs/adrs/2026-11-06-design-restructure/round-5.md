# Round 5: ADR-009, one design document with requirements and ADR appendices

Date: 2026-11-06. This round follows design rules 1–3 in `AGENTS.md`/`CLAUDE.md`. The innovation council lead ran the review with two independent peers:
- `innovation-challenger` (GPT) reviewed normative content and the ADR column.
- `experiment-peer` (Opus) reviewed provenance, ADR records, rules and citations.

The review was read-only. It made no model calls outside the council and did not touch `.codegraph`.

**I could not recompute the digests.** Neither the lead nor either peer had a shell, so:
- no file was hashed;
- `git diff --cached --stat` was not run;
- `test -L CLAUDE.md` was not run.

The proposal says `CLAUDE.md` is a symlink. The peer confirmed only that its text reads the same as `AGENTS.md`. The digests below are the ones the main agent submitted.

| File | SHA-256 (submitted; not recomputed) |
|---|---|
| `docs/scout-design.html` | `a432ba776f5ed7d69309c06eb440870f084c247f677e6af90250cb6729867996` |
| `docs/adrs/2026-11-06-design-restructure/proposal.md` | `3236b644d51f890cb35edd65e1a321342abfffc198441fb639c7964d1a157fc4` |
| `…/requirements-map.md` | `8a46cea4f0281ad3b27c11e37e167ba658d03e98286cd0aac267035f3b48b3dd` |
| `…/design-section-map.json` | `3cc0225e4bb98d6e39fbb6e71d870937adb84093cd9d301a85c026fb204da974` |
| `…/ADR.md` (ADR-009) | `e0d2d7f12d9185527e5c48138cdf54aa738c5235aae1e11ccf1ceeb1a17e1bfc` |
| `AGENTS.md` (`CLAUDE.md` symlink, as stated) | `aff517d34b71ac2ed189dcc9f4d3b03380e9b91e9b754a2bd98aceeaa4718511` |
| `docs/adrs/README.md` | `7c90063cd3050dd05984990dade0553e47e4fc7d43051b9e9c92f241c2e307e0` |
| `docs/jev-adapter-audit.md` | `30894dbf17bf9f815ceaadce2c3c9ae75d2cf0e31d4a91753985dbfe6818630d` |
| ADR-000 `2026-10-19-v1-decisions/ADR.md` | `4d631ce92da11fc7c42467d2b18cc12902f4e193d92170a7e5e33a9a8547b36f` |
| ADR-001 `2026-10-21-e1b-v2/ADR.md` | `e5db1e7659fb562bd3c89cb36483d789d0e669d3208c8f2a9e0617346d2be8fa` |
| ADR-002 `2026-10-22-rescope/ADR.md` | `4adb8f70b90d2ca4deb941b56d161ccce424549f34c323b53cd14c4c80a9ce2b` |
| ADR-003 `2026-10-23-decisions/ADR.md` | `f42a2ea88e1b02d9a41b558841f32c7393a02cce58f7ccdd0c401caa679fdfe7` |
| ADR-004 `2026-10-23-clawpatch-patch-v3/ADR.md` | `cbaf321eed0957fa29c2ac7bca02b981f3acd8c8efc895e10bc7a1bf685e577a` |
| ADR-005 `2026-10-23-codex-luna-mapping/ADR.md` | `ff8a06e034a20dc0242732a71743b7f99c1cc82d1a5293eab6d29a272e34e00c` |
| ADR-006 `2026-10-24-area-refresh-fallback/ADR.md` | `371a870bd4302d5bf3a6f224679b6ac2ad321a0c3b1dd4aeb6e70ce4941e9a82` |
| ADR-007 `2026-10-24-cutover-cleanup/ADR.md` | `de9fffdf63fc0ec61a98bdde7327617682d7b336b1199711537d48b26ee04721` |
| ADR-008 `2026-10-31-jev-merge-association/ADR.md` | `76dbd3aec2503d447270a91bdfb9a9c44e35ca730c45199add44a34847c45ab0` |

**The changed digests match the round-4 fixes.** Four digests differ from round 4:
- html: R4-01 and R4-04;
- `proposal.md`: R4-04 and the "Round 4 fixes" section;
- ADR-003: R4-03;
- the audit: R4-02.

All other digests are unchanged, as round 4 required.

## Scope

- **Under review:** worktree `/Users/adamjackson/Projects/work/scout-req`, branch `agent/requirements-path`, all changes staged. This covers:
  - the "Round 4 fixes" (`proposal.md:84-92`) at every location they touched, including each ADR-column addition;
  - a check that the fixes introduced no new inconsistency;
  - a fresh broad pass over proposal questions 1–5.
- **Base:** `main` at `9a3d19f5` in `/Users/adamjackson/Projects/work/scout/`, including the deleted originals and the frozen `docs/design-reviews/` records.
- **Checks made:**
  - **ADR column:** every Changes bullet of ADR-001…ADR-008 against the ADR column, cross-checked against each folder's `proposed.diff` and the `requirements-map.md` rows for §12 R1–R18 and §14 A1–A14.
  - **Legacy citations:**
    - the R4-02 edits against the REQ text and `design-section-map.json`;
    - a fresh grep of the live tree for `design R\d`, `R14`, `table[- ]1`, `AL `, `req §`, `design §` and `Appendix \d`;
    - deleted-path mentions.
  - **ADR records:**
    - R4-03 against base `STATUS.md` §1;
    - R4-04 against README `:29`;
    - supersession and amendment in both directions for all ten ADRs, and README statuses.
  - **Rules and ledgers:**
    - experiment rules 1–7, word for word against base `CLAUDE.md`;
    - that the ledgers are append-only.
  - **Proposal counts:** 388/323/43/22 map rows and 312 JSON keys.
  - **Normative content:** a semantic trace of 68 rows (REQ-43–69, 86–102, 132–138 and 139–155), plus targeted checks of base design `:72`, `:133`, `:176-180`, `:293`, `:322`, `:378`, `:386`, `:394` and `:488`.
- **Not reopened:** decisions settled in ADR-000…ADR-008.

## Verification of round-4 fixes

| # | Evidence | Result |
|---|---|---|
| R4-01 | **Every addition listed in `proposal.md:88` is present.** The column was checked at `html:527`, `545`, `547`, `549`, `553`, `555`, `567`, `600`, `637`, `644`, `677`, `679`, `687`, `688`, `700`, `704`, `713`, `714`, `718` and `719`.<br>**Each added ADR introduced or changed a clause that the REQ carries:**<br>• REQ-60, mapper and association fallback as codex-only exceptions: ADR-005 and ADR-008.<br>• REQ-167, Phase 4 onboarding/manual map counts: ADR-006 (`ADR.md` "Replace Phase 4's debounced-burst acceptance").<br>• REQ-170, X2b and the onboarding/manual-only branch: ADR-005 and ADR-006.<br>• REQ-171, "merge-time files instead follow X11": ADR-008.<br>• REQ-2, primary family: ADR-003.<br>• REQ-16, `requires ∈ {…model-cli…}`: ADR-002.<br>• REQ-18, model-CLI skips in the matrix: ADR-002.<br>• REQ-52, codex on `PATH`: ADR-005.<br>• REQ-54, E1b clustering rule not used for scopes: ADR-001.<br>• REQ-56, `defaultMapTimeoutMS`: ADR-004.<br>• REQ-78, recomputed bases count: ADR-003.<br>• REQ-180, the `ruff` ignore: ADR-003, already cited.<br>• The remaining rows were confirmed by the challenger's trace.<br>**No unsupported addition was found, and requirement text is unchanged in the sampled rows.**<br>**The fix is incomplete.** The claim that "the ADR column now cites every ADR whose change a requirement carries" is not yet true (R5-01). | partly fixed → R5-01 |
| R4-02 | **Audit.** `jev-adapter-audit.md:19` reads "(REQ-87)" and `:79` (row 38, Design-level) reads "(REQ-87; kept open by REQ-99)".<br>**Code comments:**<br>• `internal/decisions/client.go:20` reads "REQ-87, REQ-92"; REQ-92 (`html:574`) carries "failed units".<br>• `internal/policy/policy.go:330` reads "REQ-70–REQ-76", which are the six eligible rows.<br>• `internal/guardrails/focused_tests.go:33` reads "REQ-76", which carries "Go/Python have no corresponding construct" (`html:598`).<br>**Map keys:** `"§12 R14"` is at `json:929` and `"AL table 1"` at `:2078`.<br>**Residue:** the live-tree grep found no remaining legacy citations other than intentional `@9a3d19f5` pins and README `:29`'s notation key | fixed |
| R4-03 | ADR-003 `ADR.md:54` cites "`docs/STATUS.md`@9a3d19f5 §1, revision 2.3 history entry". Base `STATUS.md:25` is in §1, which starts at `:5`, and records adoption with `75ef1e66…`, which equals ADR-003's signed digest | fixed |
| R4-04 | **Both statements now agree with README `:29`:**<br>• `html:966` (ADR-009 card): "`docs/adrs/README.md` maps them and says which revision a frozen citation of a deleted file means".<br>• `proposal.md:43` says the same.<br>**The old rule is gone.** No remaining file applies "deleted files are read at `9a3d19f5`" to frozen records. The checked files were README, STATUS, CONTRIBUTING, `AGENTS.md`, all `ADR.md` files and the html | fixed |
| `CLAUDE.md` | The proposal (`:92`) states that it is a symlink. The peer confirmed only that its text is identical; nobody could run `test -L` | condition (see below) |

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R5-01 | minor | `docs/scout-design.html:622` (REQ-35), `:682` (REQ-41), `:706` (REQ-160), `:616` (REQ-49), `:548` (REQ-55), `:549` (REQ-56); `proposal.md:88` | **Six requirements still leave out an ADR that introduced or changed a clause they carry.** This is the same class of defect as R4-01, and it makes the claim at `proposal.md:88` untrue.<br>• **REQ-35 lacks ADR-002.** ADR-002 `ADR.md` Changes says it "Synchronises … REQ-29, REQ-31, REQ-35, REQ-36, REQ-38, REQ-39, REQ-41". The wording "controlled-continuation or within-scope event-study design signed off the same way" is ADR-002's (base `2026-10-22-rescope/proposed.diff:169-170`), and `requirements-map.md:72` maps Appendix 11 → REQ-35 + ADR-002.<br>• **REQ-41 lacks ADR-002.** The same Changes bullet names REQ-41. ADR-002 replaced "Wilson 95 % interval" with Wilson intervals for unweighted binomial lines and design-based intervals for weighted or paired estimators (`proposed.diff:198-199`). REQ-41 carries that clause and cites only ADR-000.<br>• **REQ-160 lacks ADR-002.** "Optional logged-in model lanes" is ADR-002's change from "opt-in LLM lanes" to "opt-in model-CLI lanes where a CLI is logged in" (`proposed.diff:133-134`).<br>• **REQ-49 lacks ADR-003.** The comparability definition, "head tool/policy digests", "merge-base revision" and "`uninspected: [{family, reason}]` … an empty list means every family was inspected", was introduced by ADR-003 (`2026-10-23-decisions/proposed.diff:78-79`; the old line had only "digest mismatch → paired base/head scan"). ADR-007 later made the field mandatory.<br>• **REQ-55 lacks ADR-003.** "Excludes Clawpatch records with `status: \"skipped\"` from every ownership lookup" is ADR-003's addition (`proposed.diff:43-44`); revision 2.2 had only "reads only the latest refresh's merged output".<br>• **REQ-56 lacks ADR-005.** ADR-005 added the installed-but-failing CLI case ("keeps the heuristic map and reports exploration incomplete") and R15's "never `refreshed`" (`2026-10-23-codex-luna-mapping/proposed.diff:36,75`; ADR-005 `ADR.md` Decision bullet 6 and Changes "adds risk R15"). `requirements-map.md:144` maps R15 → REQ-58, REQ-56. REQ-58 cites ADR-005; REQ-56 does not.<br>• **How the pass missed them:** the R4-01 pass read the `ADR.md` Changes bullets. Four of these six clauses appear there only as section numbers (§4, §7.2, §7.1), and two (REQ-35, REQ-41) are named explicitly but were still missed. | Add ADR-002 to REQ-35, REQ-41 and REQ-160; ADR-003 to REQ-49 and REQ-55; ADR-005 to REQ-56. Before resubmitting, run a clause-level pass: for each ADR-001…008 `proposed.diff` (ADR-008: its `proposal.md` and `requirements.proposed.diff`), list each changed sentence, take its REQs from `requirements-map.md`, and confirm the column cites that ADR. Correct the R4-01 bullet (`proposal.md:88`) and record this pass under "Round 5 fixes". |
| R5-02 | minor | `docs/scout-design.html:554`, `:556`, `:627` (REQ-17, REQ-62, REQ-107); `:683` (REQ-136); `:574` (REQ-92); `:622` (REQ-35) | **Four base statements were dropped or narrowed in the requirements.** Each is mapped to the cited REQ, but that REQ no longer carries it.<br>• **(a) JUnit ingestion.** Base design `:72` and `:133` (§5.1 lane 5) say: "`artifacts` — LCOV/Cobertura/JUnit + PR metadata, revision/producer matched … `unmeasured` when absent". Map `:201` sends lane 5 to REQ-17 and REQ-62–66. No html line mentions JUnit; only LCOV and Cobertura survive, in REQ-107. The code implements JUnit (`internal/artifacts/artifacts.go:1-2`, `:197-203`; `cmd/scout/main.go:228`), and its comment cites REQ-43 and REQ-107, neither of which carries JUnit.<br>• **(b) Actionable yield granularity.** Base `:322` reads "**Actionable yield** per lane/rule". REQ-136 keeps the formula but drops "per lane/rule"; the only "per-lane" left qualifies false assurance. Map `:104` sends §9.1 to REQ-41 and REQ-136.<br>• **(c) Jev failure disclosure.** Base `:178` says network errors, timeouts, HTTP 4xx/5xx, schema-invalid answers, exhausted 429/529, `max_tokens_exceeded`/422, a missing key, a fork and egress off "each leave the unit with no annotation, and each affected finding is disclosed `ranking: deterministic (reason)`". REQ-92 attaches "leaves no annotation and discloses `ranking: deterministic (reason)`" only to "Missing key, fork or egress denial". Network, HTTP and schema failures are only "failed units", and REQ-93's generic "per-finding disclosures" does not name the disclosure. Map `:211`: "§6 Failures are failed units" → REQ-92.<br>• **(d) Who may sign off the qualifying study.** Base `:293` says "The qualifying study may be signed off under Scout's experiment rules or under inertia-bench's rules", and Appendix 11 (`:488`) says the rule "is satisfiable by a signed-off bench study". REQ-35 says only "preregistered council-signed pooled test". Read together, REQ-36 ("Inertia-bench owns pooled studies") and REQ-42 imply that such a study can come from inertia-bench. But the explicit permission to sign off under inertia-bench's rules is gone, and `AGENTS.md` design rule 3 governs design-change evidence, not claim upgrades. Map `:72` and `:228` send these to REQ-35. | Restore each statement in its REQ, keeping the IDs:<br>(a) the artifacts-lane formats LCOV, Cobertura and JUnit plus PR metadata, revision/producer matched, `unmeasured` when absent (for example in REQ-62 or REQ-107);<br>(b) "per lane/rule" for actionable yield in REQ-136;<br>(c) in REQ-92, no annotation and `ranking: deterministic (reason)` for every listed failure cause, with `ranking: partial (n/N)` kept for capped units;<br>(d) in REQ-35, that the qualifying study may be signed off under Scout's experiment rules or inertia-bench's rules.<br>List these text changes as content restorations under "Round 5 fixes". Optionally repoint `artifacts.go:2` to the REQ that now carries JUnit (comment-only). |
| R5-03 | nit | `proposal.md:27` (change-set item 8) | **The proposal calls STATUS §4 an "old decision list".** Item 8 says "the old decision lists (§4, §6) are removed". Base `docs/STATUS.md:151` §4 is "What moved to inertia-bench, and the subject interface Scout must provide", a contract note, not a decision list. Only §6 (`:170`) is a decision list. | Reword, for example: "§4 (inertia-bench move and subject-interface notes; now ADR-002, REQ-37, REQ-135) and §6 (user decisions) are removed". |

## Answers to the proposal's questions

1. **Is normative content lost, weakened or changed?** Four details were lost (R5-02). Otherwise no change of meaning was found:
   - the 68-row trace (REQ-43–69, 86–102, 132–138 and 139–155) found nothing else;
   - formulas, thresholds and phase scenarios match, as in round 4.
2. **Do the ADR summaries and cards match their records?** Yes for the records themselves:
   - supersession and amendment agree in both directions for all ten ADRs (card headers `html:736…949` against each `ADR.md` status line);
   - README `:15-24` statuses match;
   - the R4-03 pointer is now correct.

   The ADR provenance column in Appendix A is still incomplete (R5-01).
3. **Does the Overview agree with Appendix A?** Yes. No new inconsistency was found.
4. **Are rule, tooling and citation changes complete, with no live reference to a deleted file?** Yes:
   - R4-02 closed the remaining legacy citations;
   - deleted-path mentions remain only where intended (`html:630,953,956`, `remaining-work.md:101`, audit line 7);
   - `.omp/gate.json`, `.prettierignore` and `infrastructure.py` name the new paths.
5. **Experiment rules, ledgers, frozen records:** no violation.
   - Experiment rules 1–7 match base word for word.
   - The research ledger only gains R-062 (69 → 70 lines), and the hypothesis ledger only gains one History line.
   - Each frozen folder has the base file set plus `ADR.md`, and no frozen record contains new-path strings.

## Considered and not raised

- **REQ-53 + ADR-003** (challenger). ADR-003's Phase 4 acceptance spelled out "one-to-one match … unmatched active areas kept or retired per §4" (`decisions/proposed.diff:155`). It did so as a description of the existing §4 rule ("reconciliation applies the §4 rule"; ADR-003 `ADR.md` Decision: "under the existing ≥ 0.5 matching rule"). That restates the rule rather than changing it. Adding ADR-003 would be harmless, because extra citations are never an issue.
- **ADR-000 as the origin of REQ-13, REQ-53, REQ-55 and REQ-56.** `html:519` requires a changed requirement to cite the ADR that changed it. Round 4 accepted REQ-10 (ADR-007 and ADR-008, without ADR-000) as following the convention, and the peer agrees. Only missing changers are raised.
- **ADR-009 card wording** (peer P5-04). The card's frozen-record list leaves out "signed review rounds", which README `:27`, ADR-009 `ADR.md` and `proposal.md:43` include. It is a shorter list, not a contradiction.
- **ADR-005 `ADR.md:35` "amends the rev 2.4 baseline from ADR-004"**, while ADR-004 is shown as plain "accepted". The phrase describes the base revision, not a relation between ADRs.
- **ADR-007's "Amend" wording** for edges whose status lines say "partly superseded". Rounds 1–2 accepted this grouping.
- **Round-4 nits:** none was found to have become an inconsistency.

## Peers

- **`innovation-challenger`:**
  - Confirmed every R4-01 addition.
  - Found all six R5-01 rows plus REQ-53, which was not adopted.
  - Found R5-02 (a), (b) and (c).
  - Traced 68 rows.
- **`experiment-peer`:**
  - Verified R4-02, R4-03 and R4-04, all supersession edges, the rules, the ledgers, the frozen folders and the counts.
  - Found the REQ-35 and REQ-41 rows of R5-01 independently.
  - Found R5-02 (d) (rated nit; raised here as minor because the base states it explicitly) and R5-03.
- **Lead:**
  - Found the REQ-56/ADR-005 gap independently through `requirements-map.md:144` (R15 → REQ-56) and `codex-luna-mapping/proposed.diff:36,75`.
  - Re-checked every peer claim against the base: rescope diff `:133-134`, `:169-170`, `:198-199`; decisions diff `:43-44`, `:78-79`, `:155`; base design `:72`, `:133`, `:178`, `:293`, `:322`, `:488`; base STATUS `:151`, `:170`; JUnit in `internal/artifacts`.
  - Rejected REQ-53 + ADR-003 and the ADR-000 origin rows.
- **Effect of the peers:** without them, the lead would have raised only REQ-56. Every R5-02 item and R5-03 come from a peer.

## Verdict

**3 issues (R5-01 minor, R5-02 minor, R5-03 nit): NOT signed off.**

No sign-off applies to the digests above, and none of them was recomputed in this round.

### Required with the next submission

- **Changed digests:** expected only for:
  - `docs/scout-design.html` (R5-01, R5-02);
  - `proposal.md` (corrected R4-01 bullet, R5-03, and a "Round 5 fixes" section).

  All other digests should stay as they are. If `artifacts.go:2` is repointed, list it as comment-only.
- **Recomputed digests:** the main agent must recompute every digest with `shasum -a 256` on the staged worktree. It must also run:
  - `test -L CLAUDE.md && readlink CLAUDE.md`, expecting `AGENTS.md`;
  - `git -C /Users/adamjackson/Projects/work/scout-req diff --cached --stat`, to confirm that no other governed file is staged.

  Record the outputs in the submission.

---

## Council memo

1. **Recommendation:** fix R5-01, R5-02 and R5-03 in one pass and submit round 6.
   - For R5-01, use a pass driven by the diffs, not the Changes bullets. Each ADR's `proposed.diff` sentence → `requirements-map.md` row → ADR column. Rounds 3, 4 and 5 each found more rows because earlier passes worked from summaries.
   - R5-02 restores base text only; no new decision is needed.
2. **Alternatives:**
   - (a) **Sign off with conditions.** Rejected: rule 1 requires zero issues.
   - (b) **Make the ADR column informative only** (header note: "cites the principal ADRs"). This would end the class of defect, but it is a new design decision that weakens the provenance rule R3-02 and R4-01 enforced. It is acceptable only if the user chooses it, and the council would review it.
   - (c) **Generate the ADR column from a clause-level source map** (the challenger's suggestion). This is more robust but adds an artifact. The diff-driven pass in (1) gets the same result for this change.
   - (d) **Treat R5-02 (d) as implied** by REQ-36 and REQ-42. Rejected: the base states the permission explicitly, and the fix costs one clause.
3. **Evidence:** each issue row cites the exact worktree and base `file:line`, and the verification table cites each R4 fix location.
4. **Risks:**
   - **Unverified digests:** they are still not checked, so round 6 must review the actual staged bytes.
   - **R5-01:** a later ADR author may miss an interacting decision, for example changing baseline comparability without seeing ADR-003.
   - **R5-02:**
     - (a) leaves shipped JUnit ingestion without any design anchor;
     - (b) allows pooled yield to hide weak rules;
     - (c) allows failed Jev calls to go undisclosed per finding;
     - (d) allows a qualifying inertia-bench study to be wrongly rejected.
   - **Reversibility:** all fixes are text edits and fully reversible.
5. **Execution outline:**
   1. Add the six R5-01 citations, then run the diff-driven ADR-column pass and record any further rows.
   2. Restore R5-02 (a)–(d) in REQ-62 or REQ-107, REQ-136, REQ-92 and REQ-35, keeping the IDs. Optionally repoint `artifacts.go:2`.
   3. Fix `proposal.md:27` (R5-03) and `:88`, and add "Round 5 fixes".
   4. Recompute every digest with `shasum -a 256`, run `test -L CLAUDE.md` and `git diff --cached --stat`, and confirm that only the html, the proposal and any comment-only code changed.
   5. Submit round 6 with the digest list and the command outputs.
