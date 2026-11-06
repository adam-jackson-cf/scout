# Round 6: ADR-009, one design document with requirements and ADR appendices

Date: 2026-11-06. This round follows design rules 1–3 in `AGENTS.md`/`CLAUDE.md`. The innovation council lead ran the review with two independent peers:
- `innovation-challenger` (GPT) reviewed normative content and the ADR column.
- `experiment-peer` (Opus) reviewed provenance, ADR records, rules, ledgers and consistency across documents.

The review was read-only. It made no model calls outside the council and did not touch `.codegraph`.

**The digests were recomputed by the main agent, not by the council.** The lead and both peers had read-only tools and no shell, so none of them hashed a file. Earlier rounds accepted a main-agent recomputation, so this round does too. The main agent ran these commands on the staged worktree and recorded the results:
- `shasum -a 256` on each file below;
- `test -L CLAUDE.md && readlink CLAUDE.md` → `AGENTS.md`;
- `git diff --cached --stat` → `244 files changed, 5736 insertions(+), 1781 deletions(-)`.

The main agent also states that only the files listed below changed since round 5, plus the comment-only edit at `internal/artifacts/artifacts.go:2`.

| File | SHA-256 (main agent, `shasum -a 256`) | vs round 5 |
|---|---|---|
| `docs/scout-design.html` | `411da20b2d2fd1d5befd342091a928ab1496140e4a55e51b3e05a8ae8edb56a2` | changed (R5-01, R5-02) |
| `docs/adrs/2026-11-06-design-restructure/proposal.md` | `7cf8c60905556dea1c25467c1666f240762c19f1867f633a559382422d257b9b` | changed (R5-03, "Round 5 fixes") |
| `…/requirements-map.md` | `8a46cea4f0281ad3b27c11e37e167ba658d03e98286cd0aac267035f3b48b3dd` | unchanged |
| `…/design-section-map.json` | `3cc0225e4bb98d6e39fbb6e71d870937adb84093cd9d301a85c026fb204da974` | unchanged |
| `…/ADR.md` (ADR-009) | `e0d2d7f12d9185527e5c48138cdf54aa738c5235aae1e11ccf1ceeb1a17e1bfc` | unchanged |
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

**The digest changes match round 5's requirement.** Only the html and `proposal.md` changed. The other 15 digests are byte-for-byte the values in the `round-5.md` table.

## Scope

- **Under review:** worktree `/Users/adamjackson/Projects/work/scout-req`, branch `agent/requirements-path`, all changes staged. This covers:
  - the "Round 5 fixes" (`proposal.md:96-100`) at every location they touched;
  - a check that the fixes introduced no new inconsistency, including the Overview against the restored REQs;
  - a fresh broad pass over proposal questions 1–5.
- **Base:** `main` at `9a3d19f5` in `/Users/adamjackson/Projects/work/scout/`, including the deleted originals and the frozen `docs/design-reviews/` records.
- **Checks made:**
  - **R5-01:**
    - the six ADR-column additions in the html;
    - each one traced to the base `proposed.diff` sentence that introduced the clause;
    - spot checks of the clause-level pass. The challenger checked ADR-006/007/008 diffs. The peer checked every changed `**REQ-n**` line in the ADR-003 and ADR-005…008 `requirements.proposed.diff`.
  - **R5-02:**
    - REQ-62, REQ-92, REQ-136 and REQ-35 against base design `:72`, `:133`, `:178`, `:293`, `:322` and `:488`;
    - the ADR that introduced each restored clause (grep of base `docs/design-reviews/*/proposed.diff`);
    - `internal/artifacts/artifacts.go:1-2` against base.
  - **R5-03:** `proposal.md:27` against base `docs/STATUS.md` §4 (`:151`), §6 (`:170`) and §6.9, and against the destination `ADR.md` source lines.
  - **New inconsistency:**
    - Overview (`html:170-515`: lane diagram, matrix, popovers, record diagram) against REQ-62/92/136/35;
    - REQ-17, 36, 41, 42, 43, 91, 93, 97 and 107;
    - `requirements-map.md` rows `:104`, `:201`, `:211`, `:228`.
  - **Broad pass:**
    - a fresh trace of 44 subclauses (base design `:215`, `:341`, `:343`, `:349-360`, `:412`, `:418` → REQ-40, 162–166, 183);
    - all nine allow-list report-only rows → REQ-79–83;
    - supersession and amendment in both directions for all ten ADRs, and README statuses;
    - a live-tree grep for deleted paths and legacy citation forms;
    - experiment rules 1–7 against base `CLAUDE.md`;
    - that the ledgers are append-only;
    - the 312 JSON keys and the 388/323/43/22 map counts.
- **Not reopened:** decisions settled in ADR-000…ADR-008.

## Verification of round-5 fixes

| # | Evidence | Result |
|---|---|---|
| R5-01 | **All six additions are present.** Each added ADR introduced a clause that the REQ carries:<br>• REQ-35 `html:622` (ADR-001, **ADR-002**, ADR-003): rescope `proposed.diff:169-170`.<br>• REQ-41 `:682` (ADR-000, **ADR-002**): rescope `:198-199`.<br>• REQ-160 `:706` (ADR-000, **ADR-002**, ADR-003, ADR-007): rescope `:133-134`.<br>• REQ-49 `:616` (**ADR-003**, ADR-007, ADR-008): decisions `proposed.diff:78-79`.<br>• REQ-55 `:548` (**ADR-003**, ADR-006, ADR-008): decisions `:43-44`.<br>• REQ-56 `:549` (ADR-004, **ADR-005**, ADR-006, ADR-008): codex-luna `proposed.diff:36,75`.<br>**Spot checks found no further missing citation.** The challenger checked ADR-006 (`proposed.diff:45-46,99-100`; `requirements.proposed.diff:17`), ADR-007 (`proposed.diff:28,37,46,60,68-71`; `requirements.proposed.diff:17,26,29,37`) and ADR-008 (`requirements.proposed.diff:19,26,35,39`). The peer checked every `-/+ **REQ-n**` line of the ADR-003 and ADR-005…008 requirements diffs (REQ-10, 13, 25, 27, 28, 29, 31, 35, 36). | fixed |
| R5-02 (a) | **REQ-62 (`html:556`) now reads:** "artifacts ingest LCOV/Cobertura/JUnit plus PR metadata, revision/producer matched, with `unmeasured` when absent".<br>• This matches base `:72` and `:133`. The triggers "artifacts … on PR/push" match the lane-5 `pr, push`.<br>• The clause has been context-only since the e1b-v2 snapshot (`2026-10-21-e1b-v2/proposed.diff:58`). ADR-000 is therefore the right origin, and it is cited.<br>• `artifacts.go:2` now reads "(REQ-43, REQ-62, REQ-107)"; base read "(design §8.2, §8.7)". The change is comment-only. | fixed |
| R5-02 (b) | **REQ-136 (`html:683`) now reads:** "actionable yield per lane/rule (`accepted` or resolved-by-change ÷ dispositioned, open separate)".<br>• This matches base `:322`.<br>• It is v1 content (context line in rescope `proposed.diff:201`), and ADR-000 is cited. | fixed |
| R5-02 (c) | **The fix follows round 5's required-fix wording.** REQ-92 (`html:574`) now gives every listed cause "a failed unit with no annotation and disclose `ranking: deterministic (reason)` per affected finding", and keeps "caps disclose `ranking: partial (n/N)`".<br>• It agrees with REQ-93, REQ-97, REQ-179 and the Overview (`html:266-267`, "else deterministic").<br>• **One sentence of the same base paragraph is still missing:** "the adapter treats any non-success as a failed unit" (R6-01). | fixed as specified; residue → R6-01 |
| R5-02 (d) | **REQ-35 (`html:622`) now reads:** "the qualifying study may be signed off under Scout's experiment rules or inertia-bench's rules".<br>• This matches base `:293` and `:488`.<br>• The clause was introduced by ADR-002 (rescope `proposed.diff:170`), which is now cited.<br>• It is consistent with REQ-36 ("Inertia-bench owns pooled studies"), REQ-42 and `AGENTS.md` design rule 3. | fixed |
| R5-03 | **The fix names STATUS §4 and §6 by content at `proposal.md:27`.**<br>• The §4 destinations are correct: ADR-002, REQ-37 (`html:676`) and REQ-135 (`:680`); the exact contract schema is a listed drop (`requirements-map.md:340`).<br>• **The §6 destination is wrong** (R6-02). | partly fixed → R6-02 |
| Conditions from round 5 | **All three outputs were recorded by the main agent:**<br>• `test -L CLAUDE.md && readlink CLAUDE.md` → `AGENTS.md`;<br>• `git diff --cached --stat`;<br>• recomputed digests. | met (main-agent outputs accepted) |

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R6-01 | minor | `docs/scout-design.html:574` (REQ-92) | **REQ-92 leaves out the base design's catch-all failure rule.**<br>• Base `docs/scout-design.md:178` ends the "Failures are failed units" paragraph: "Whether the OpenRouter route returns the same codes and headers as TypeSafe's documented API is undocumented **[INFERENCE]**; **the adapter treats any non-success as a failed unit**."<br>• ADR-002 introduced the paragraph (`2026-10-22-rescope/proposed.diff:113`), and `requirements-map.md:211` maps it only to REQ-92.<br>• REQ-92 now lists specific causes, which reads as exhaustive. No html line carries "any non-success" (grep: no match). REQ-93 covers lane health, not which responses count as failed units.<br>• The base states this fail-closed default explicitly, so its absence is a weakening (question 1), as with R5-02 (d).<br>• Round 5's required fix for R5-02 (c) asked only for "every listed failure cause", so this residue is the council's own omission from round 5. | Add to REQ-92: "any other non-success response is likewise a failed unit". Keep the listed causes and the disclosure and cap semantics. ADR-002 is already cited, so the ADR column needs no change. List it under "Round 6 fixes" as a content restoration. |
| R6-02 | nit | `proposal.md:27` (change-set item 8) | **Item 8 sends STATUS §6 to ADR-000, but §6 is in ADR-002 and ADR-003.** Item 8 reads "§6 (user decisions; now ADR-000 and ADR-003)".<br>• Base `docs/STATUS.md:170` heads §6 "Decisions on the former open questions (user, 2026-10-22; R-036)". §6.9 holds the 2026-10-23 decisions (R-039).<br>• ADR-002 `ADR.md:5` names "user direction in `docs/STATUS.md`@9a3d19f5 §6" as its source, and the map sends Appendix 9–14 (= §6.1–6.7) to ADR-002 (`requirements-map.md:72,75,76,79`).<br>• ADR-003 `ADR.md:5` cites §6.9 and carries Appendix 13 and 15 (`requirements-map.md:81-82`).<br>• ADR-000 `ADR.md:5` cites only the v1 memo and Appendix items 1–7 (2026-10-19). It carries none of §6. | Change to "§6 (user decisions of 2026-10-22 and 2026-10-23; now ADR-002 and ADR-003)". Naming R-036 and R-039 is optional. |
| R6-03 | nit | `proposal.md:26` (change-set item 7) | **Item 7 puts quotation marks around a paraphrase.** It reads: a hypothesis-ledger History line recording "no state change, paths moved". The actual line, `docs/hypothesis-ledger.md:39`, reads "All IDs: no state change. Paths `docs/design-reviews/<folder>/…` in earlier lines now resolve to `docs/adrs/<folder>/…`; folder names are unchanged (R-062)." | Drop the quotation marks, or quote the real text. |

## Answers to the proposal's questions

1. **Is normative content lost, weakened or changed?** One sentence is still lost (R6-01). Otherwise no change of meaning was found:
   - the four R5-02 restorations match the base;
   - the challenger's fresh trace found no loss: 44 subclauses (pipeline jobs and cache, 5,000-job scale bounds, retained boundaries, release-clone install, Phase 1 closing statements, Clawpatch upgrade review) and the nine allow-list report-only rows.
2. **Do the ADR summaries and cards match their records?** Yes:
   - card relations (`html:736…949`) match every `ADR.md` status line and README `:15-25`, in both directions;
   - the ADR column now cites every ADR found to have introduced a clause the REQ carries, in R5-01's six rows and the spot-checked diffs.
3. **Does the Overview agree with Appendix A?** Yes. The Overview has no lane-input list, metric text or claim-upgrade text that could disagree with the restored REQs:
   - `html:179` ("reads CI artifacts whose provenance matches") agrees with REQ-43 and REQ-62;
   - the Rank stage says "else deterministic" (`:266-267`).
4. **Are rule, tooling and citation changes complete, with no live reference to a deleted file?** Yes:
   - the live-tree grep finds only intentional historical mentions (`jev-adapter-audit.md:7`, `plans/2026-11-06-remaining-work.md:101`, `html:842,953`);
   - `artifacts.go:2` now points to REQ-62, which carries JUnit.
5. **Experiment rules, ledgers, frozen records:** no violation.
   - Experiment rules 1–7 match base `CLAUDE.md` word for word.
   - The research ledger keeps rows 1–69 and appends R-062 at `:70`.
   - The hypothesis ledger adds one newest-first History line at `:39` and leaves earlier lines unchanged.
   - Frozen records were not edited.

## Considered and not raised

- **REQ-35 leaves out "; no such study is planned in Scout (see above)"** (base `:293`; peer F2, rated nit). ADR-002 round 1, issue 21, added it to reconcile the Scout-rules branch with §8.5. REQ-36 ("Inertia-bench owns pooled studies", `html:675`) and REQ-42 already carry that point. It is a planning statement, not a requirement.
- **STATUS renumbering** (peer F3). The worktree STATUS now has §4 "Remaining work" and §5 "How to resume"; base had §5 and §7. README `:29` lists "`docs/STATUS.md` §4 and §6" among deleted items and says "Commit `9a3d19f5` holds the last version of each", which fixes the base numbering. Frozen records cite STATUS as it was when they were written, and many such citations were already stale at base (for example "§5 row 8"). STATUS is a living status file and is not among the governed documents.
- **REQ-92's "exhausted 429/529 exponential backoff honoring `retry-after`"** is condensed wording with the same meaning as base "429/529 after exponential backoff that honours `retry-after`".
- **The "[INFERENCE]" clause that OpenRouter's codes are undocumented** is an evidential note, not a requirement. Only the catch-all that follows it is normative (R6-01).
- **Round-5 "considered" items** (REQ-53 + ADR-003, ADR-000 origin rows, ADR-009 card wording, ADR-005 "amends the rev 2.4 baseline", ADR-007 "Amend" wording) are unchanged and still not issues.

## Peers

- **`innovation-challenger`:**
  - Verified all six R5-01 rows against the diffs and all four R5-02 restorations.
  - Found R6-01.
  - Traced 44 new subclauses and the nine allow-list report-only rows.
  - Spot-checked the ADR-006/007/008 diffs.
- **`experiment-peer`:**
  - Verified R5-03's §4 destinations, the six additions, the restorations, `artifacts.go:2`, the counts (312 keys) and the map rows.
  - Verified all supersession edges, rules, ledgers and the live-tree grep.
  - Found R6-02 (rated minor; recorded here as nit, consistent with R5-03) and R6-03.
  - Also reported F2 and F3, which are not raised (see above).
- **Lead:**
  - Found R6-02 independently, by checking base STATUS §6/§6.9 against the `ADR.md` source lines and map `:72-82`.
  - Noticed R6-01 independently. At first the lead held it as covered by "HTTP 4xx/5xx + schema-invalid". The challenger's argument changed that: the base states the rule explicitly, and round 5 set the precedent with R5-02 (d). It is now raised.
  - Re-checked each peer claim against the base: rescope `proposed.diff:113,169-170`; base design `:72,133,178,293,322,488`; base STATUS `:150-195`; `hypothesis-ledger.md:39`; README `:29`; ADR-000/002/003 `ADR.md:5`.
- **Effect of the peers:** R6-03 came from the peer, and R6-01 was raised because of the challenger. Without them, the lead would have raised only R6-02.

## Verdict

**3 issues (R6-01 minor, R6-02 nit, R6-03 nit): NOT signed off.**

No sign-off applies to the digests above.

### Required with the next submission

- **Changed digests:** expected only for:
  - `docs/scout-design.html` (R6-01);
  - `proposal.md` (R6-02, R6-03, and a "Round 6 fixes" section).

  All other digests in the table above should stay as they are.
- **Recomputed digests and checks:** the main agent must recompute every digest with `shasum -a 256` on the staged worktree. It must also re-run:
  - `test -L CLAUDE.md && readlink CLAUDE.md`, expecting `AGENTS.md`;
  - `git -C /Users/adamjackson/Projects/work/scout-req diff --cached --stat`.

  Record the outputs in the submission. Since `artifacts.go` changed (comment only), also record the `go build ./... && go vet ./...` result that `proposal.md:22` claims.

---

## Council memo

1. **Recommendation:** fix R6-01, R6-02 and R6-03 in one pass and submit round 7.
   - R6-01 adds one clause to REQ-92; no new decision is needed.
   - R6-02 and R6-03 are wording changes in the proposal only.
2. **Alternatives:**
   - (a) **Sign off with conditions.** Rejected: design rule 1 requires zero issues.
   - (b) **Treat R6-01 as implied** by "HTTP 4xx/5xx" and "schema-invalid answers". Rejected: the base states the fail-closed default explicitly, and round 5 held the same standard for R5-02 (d). It also protects against OpenRouter outcomes outside the list, which the base itself flags as undocumented.
   - (c) **Restore base `:178` word for word in REQ-92** (the challenger's alternative). This is acceptable but longer; the one-clause addition is enough.
3. **Evidence:** each issue row cites the exact worktree and base `file:line`, and the verification table cites each R5 fix location and its originating diff line.
4. **Risks:**
   - **R6-01:** an adapter change could treat an unlisted outcome (for example a non-standard OpenRouter response) as something other than a failed unit. That could make a ranking look like Jev's without disclosure.
   - **R6-02:** a reader looking for the 2026-10-22 decisions is sent to the wrong ADR.
   - **R6-03:** a misquotation, with no effect on behaviour.
   - **Reversibility:** all fixes are text edits and fully reversible.
   - **Unknown:** the council saw no staged-file list or per-file stat, only the main agent's summary. Round 7 relies on the main agent's statement that only the html and the proposal change.
5. **Execution outline:**
   1. In REQ-92 (`html:574`), add "any other non-success response is likewise a failed unit". Leave the ADR column unchanged.
   2. Change `proposal.md:27` §6 to "now ADR-002 and ADR-003". Remove the quotation marks at `:26`, or quote `hypothesis-ledger.md:39` word for word.
   3. Add "Round 6 fixes" to `proposal.md`.
   4. Recompute every digest with `shasum -a 256`. Run `test -L CLAUDE.md`, `git diff --cached --stat`, and `go build ./... && go vet ./...`. Confirm that only the html and the proposal changed.
   5. Submit round 7 with the digest list and the command outputs.
