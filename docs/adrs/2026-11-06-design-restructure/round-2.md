# Design review round 2: ADR-009, one design document with requirements and ADR appendices

Date: 2026-11-06. Rules: `AGENTS.md`/`CLAUDE.md` design rules 1–3. The innovation council (lead) ran this review with two independent peers: `innovation-challenger` (GPT, adversarial) and `experiment-peer` (Opus, evidence). The review was read-only. It used no shell and computed no hashes. It made no model calls outside the council and did not touch `.codegraph`.

Reviewed files (SHA-256). The main agent computed these digests; the council computed none.

| File | SHA-256 |
|---|---|
| `docs/scout-design.html` | `01173d1c43e88bd3996eb26d6e9355fb887ed033cdca0ab1586d1b85ed431afd` |
| `docs/adrs/2026-11-06-design-restructure/proposal.md` | `f3902ae7a584e8bd4a6820cca539689480edda2df143dbebb6fc54e972317e9a` |
| `docs/adrs/2026-11-06-design-restructure/requirements-map.md` | `3dc2b80523062a4d80ee69e9d1dd18e807ae17cffb48ae33d0f225596b46af16` |
| `docs/adrs/2026-11-06-design-restructure/design-section-map.json` | `3cc0225e4bb98d6e39fbb6e71d870937adb84093cd9d301a85c026fb204da974` (unchanged since round 1) |
| `docs/adrs/2026-11-06-design-restructure/ADR.md` | `005bfee16df88cac9ba0e6401ca874552109032fc96be09de83d5f9c64bd0c1f` |
| `AGENTS.md` | `7abc91dc004d90ffc29be6679b5775d01866eaad0b1158de25b7d73753516e5c` |
| `docs/adrs/README.md` | `e296160a142929b86113d404b1ef5ae31cbb2513dbf96a6201606d8ab12dd8cd` |

## Scope

- **Under review:**
  - worktree `/Users/adamjackson/Projects/work/scout-req`, branch `agent/requirements-path`;
  - the fixes listed in `proposal.md` "Round 1 fixes";
  - everything those fixes touched: Appendix A rows, Overview, ADR cards, every `docs/adrs/*/ADR.md`, `AGENTS.md`, `docs/adrs/README.md`, `docs/STATUS.md`, `CONTRIBUTING.md`, `README.md` and `docs/plans/2026-11-06-remaining-work.md`.
- **Base:** `main` at `9a3d19f5` in `/Users/adamjackson/Projects/work/scout/`, including deleted files and the `docs/design-reviews/` originals.
- **Checks made:**
  - each R1 fix against round 1's required fix and the base wording;
  - supersession in both directions across all ten cards and summaries;
  - Overview restatements against Appendix A;
  - digest provenance in the new `ADR.md` files against the frozen round records;
  - where ADR-009's own signed digest is said to be recorded;
  - `AGENTS.md` experiment rules, word for word against base `CLAUDE.md`.
- **Not reopened:** decisions settled in ADR-000…ADR-008.

## Verification

| # | Round-1 issue | Evidence | Result |
|---|---|---|---|
| R1-01 | Report profile, exit 2/3 as neutral check | REQ-21 (`html:588`): "under `report`, exits 2 and 3 are surfaced as a neutral check, not a failure". Base design :205 | fixed |
| R1-02 | Causal-claim designs | REQ-35 (`html:622`): "Causal claims need a controlled-continuation or within-scope event-study design signed off the same way; gate use is a further separate decision", plus "repeated dashboard significance never upgrades a claim". Base :293 | fixed |
| R1-03 | Rare-security fire-rate proxy | REQ-22 (`html:589`): "first-baseline `pre-existing` counts in non-generated paths are the fire-rate proxy". Base :208; AL §3 | fixed |
| R1-04 | One gate reason per family | REQ-27 (`html:612`): "with one gate reason per uninspected family". The `guardrails.*` exemption and unchanged `pre-existing`/`regressed` attribution are kept. Base :204 | fixed |
| R1-05 | Over-cap PRs get `linkage: incomplete` | REQ-127 (`html:647`): "a PR that exceeds either cap (50 files, 100 commits)…". Base :247 | fixed |
| R1-06 | Jev backlog and its timing | REQ-99 (`html:581`): B-1–B-4 pointer, and "B-3 runs the row-30 flip-rate harness on the first Choice that gates anything". `requirements-map.md:407` records the restoration; "Unplaced statements: none" is gone | fixed |
| R1-07 | ADR status and provenance | ADR-009 reads `accepted` in the card (`html:949`), README `:20`, STATUS `:7` and `ADR.md:3`. ADR-000: "no council round" at `html:729,736` and README `:3`. The footer phrase "v1 council decision memo" matches the base memo header (`docs/archive/scout-design-v1.md:3`) | fixed (STATUS's digest pointer: R2-01) |
| R1-08 | ADR-009 `ADR.md` | The file exists and the plan link (`remaining-work.md:14`) resolves. The Review section (`ADR.md:36-38`) has no outcome and no digests, which new rule 2 requires (`AGENTS.md:41-43`) | partly → R2-01 |
| R1-09 | Governed documents | `AGENTS.md:35-40` names `docs/adrs/*/ADR.md`, the Met/Fixed status and Design-level rows of `docs/jev-adapter-audit.md`, and `docs/evidence/study-export.md`. `.omp/gate.json` adds the audit file to the sensitive list | fixed |
| R1-10 | Undisclosed changes | Decisions are now listed in `proposal.md` "Decisions this ADR makes", `ADR.md:9-16,25-26` and `html:952-960`. ADR-000 item 6 and ADR-002 Appendix 12 are marked superseded by ADR-009. But "keep their IDs and meaning" (`proposal.md` item 3) and `html:963` remain | partly → R2-04, R2-05 |
| R1-11 | Supersession records | Every edge is now reciprocal (table below). The status-line exception is at `AGENTS.md:44-46`, `html:732` and README `:10`. One relation is still missing (ADR-008 → ADR-002) | fixed; residual in R2-05 |
| R1-12 | Overview vs Appendix A | All six named items are aligned:<br>• popover `html:450` separates the whole-run and per-family cases and keeps the `guardrails.*` exemption<br>• exit 3 at `:442` matches `cmd/scout` readiness (not ready → `exitUnavailable`)<br>• nightly onboarding at `:324`<br>• LLM column at `:354`<br>• `unknown` ≤ 10 % at `:449`<br>• "Three record types plus association labels" at `:454`<br>The `:449` rewording changes the `unknown` denominator, and the ledger diagram misstates when a finding blocks | partly → R2-02 |
| R1-13 | Precedence | `html:185`: "Appendix A is the normative requirement set… Where they differ, Appendix A prevails". `ADR.md:9` agrees | fixed |
| R1-14 | Experiment list | X1–X6, X9–X11 at `AGENTS.md:5`, `:59` (ledger rule 2) and REQ-42 (`html:529`). `requirements-map.md:403` records the choice of §13's list over §1's | fixed |
| R1-15 | STATUS §6.9 citation | `AGENTS.md:73-74`: "(user decision 2026-10-23, R-039, ADR-003)". R-039 (`research-ledger.md:47`) is that decision | fixed |
| R1-16 | Dangling references | REQ-110 (`html:630`) cites `docs/inertia-research.md` §5 at `9a3d19f5`. REQ-111 (`:631`) cites REQ-125. The stale plan lines are gone. `ADR.md` citations are pinned and in the past tense. But the blanket `@9a3d19f5` pin was also applied to historical digest inventories and one section citation, where it is false | partly → R2-03 |
| R1-17 | README path map | `docs/adrs/README.md:25-29` covers moved folders, every deleted file, STATUS §4/§6 at `9a3d19f5`, and `design-section-map.json` for section citations (keys `AL §3`, `Appendix 6`, `Appendix 12` and `req §8` sampled) | fixed |
| — | `AGENTS.md` experiment rules | Rules 1–7 are word for word with base `CLAUDE.md`. Only the scope line changed (X9–X11, an ADR-009 decision). Worktree `CLAUDE.md` matches `AGENTS.md` | ok |
| — | `requirements-map.md` counts | The table spans lines 11–399 (389 rows). 22 rows are `dropped`. The peer counted 43 rows marked `ADR` | match |

**Supersession check (R1-11).** Each older ADR's status line and card names the successors that the newer cards claim:

| Older ADR | Superseded or amended by | Card evidence |
|---|---|---|
| 000 | superseded in part by 004–009; amended by 001–003 | `html:736`, and the newer headers at `:759`, `:783`, `:807`, `:831`, `:854`, `:877`, `:901`, `:925`, `:949` |
| 001 | superseded in part by 002; amended by 003 | `:759`, `:783`, `:807` |
| 002 | superseded in part by 003, 005, 007, 009 (ADR-008 missing: R2-05) | `:783` |
| 003 | superseded in part by 007 | `:807`, `:901` |
| 005 | superseded in part by 006, 007 | `:854`, `:877`, `:901` |
| 006 | superseded in part by 008 | `:877`, `:925` |
| 007 | superseded in part by 008 | `:901`, `:925` |

## Issues

### R2-01 (minor): ADR-009's signed digest is said to be where it is not, and ADR-009's `ADR.md` breaks the new rule 2

- **Location:**
  - `docs/STATUS.md:7`
  - `CONTRIBUTING.md:29`
  - `AGENTS.md:41-43` (rule 2)
  - `docs/adrs/2026-11-06-design-restructure/ADR.md:36-38`
- **Evidence:**
  - STATUS :7 says "signed digest in `docs/adrs/2026-11-06-design-restructure/ADR.md`".
  - CONTRIBUTING :29 says "The digest must match the SHA-256 recorded in `docs/adrs/2026-11-06-design-restructure/ADR.md`". At base, CONTRIBUTING :29 pointed to STATUS §1, which held the digest (base STATUS :7, `d1fb1274…`).
  - ADR-009's `ADR.md` Review section says only "The final round names the signed digests". `proposal.md` confirms this is deliberate: "so the summary does not hash itself".
  - New rule 2 requires each `ADR.md` to contain "review outcome and signed digests". ADR-009's summary has neither.
  - Result: the CONTRIBUTING fresh-machine check cannot be followed, and ADR-009 breaks the rule it introduces.
  - The pointer approach is sound; a summary that is itself signed cannot hold its final outcome or its own digest. The rule and the pointers just have to say so.
- **Required fix:**
  - Amend rule 2: an `ADR.md` that is itself among the signed files names the folder's final round for the review outcome and digests instead of containing them. Keep `docs/adrs/README.md:3` consistent.
  - Make ADR-009's Review section name the final round file.
  - Point STATUS :7 and CONTRIBUTING :29 to that final round, `docs/adrs/2026-11-06-design-restructure/round-<N>.md`.

### R2-02 (minor): two Overview restatements of the gate diverge from Appendix A

- **Location:**
  - `html:449` (activation callout, rewritten for R1-12);
  - `html:499` (ledger diagram connector label).
- **Evidence:**
  - **`:449`:** "at least 5 comparable PR runs, with zero unwarranted would-blocks and at most 10% of **those runs** with `unknown` attribution". "Those runs" refers to the comparable runs, which by definition exclude the `comparison: none` runs that must be counted.
    - Base :204: the bound counts runs "that are `comparison: none` or carry at least one base-gap `unknown` finding, over **all PR runs in the window**".
    - REQ-22 (`:589`): "`unknown` in at most 10 % of PR runs".
    - REQ-78 (`:600`): "the `unknown` bound counts each PR run once if `comparison: none` or any base-gap unknown exists".
  - **`:499`:** the arrow to "Blocks the PR / exit 1" is labelled "if new in the PR and the check is activated".
    - REQ-21 (`:588`): under `report`, findings never fail CI, so a check can be activated without blocking.
    - REQ-20 (`:587`): `regressed` findings also block.
    - The gate diagram above (`:410-433`) correctly includes "Guardian profile?".
  - Appendix A prevails (`:185`), so neither changes the normative set. But R1-12's required fix was to align Overview statements with their REQs.
- **Required fix:**
  - `:449`: "…and `unknown` attribution in at most 10 % of all PR runs in the window (a run counts once if it is `comparison: none` or has a base-gap `unknown` finding)". Keep the separate minimum of five comparable-base runs.
  - `:499`: "if introduced or regressed, the check is activated and the profile is `guardian`", or "only when every gate condition above holds".

### R2-03 (minor): the R1-16 pinning attached `@9a3d19f5` to historical artifacts it does not identify

- **Location:**
  - `docs/adrs/2026-10-31-jev-merge-association/ADR.md:51`
  - `docs/adrs/2026-10-24-area-refresh-fallback/ADR.md:57`
  - `docs/adrs/2026-10-24-cutover-cleanup/ADR.md:54`
  - `docs/adrs/2026-10-21-e1b-v2/ADR.md:44`
- **Evidence:**
  - **`ADR-008/ADR.md:51`:** "`docs/scout-design.md`@9a3d19f5: `7b102ebb…0551`". That is the round-8 draft digest (base `…/round-8.md:5`). The very next line, `:52`, gives the same path at the same commit as `d1fb1274…`, which is correct (`proposal.md:3`). One file at one commit cannot have two digests.
  - **`ADR-006/ADR.md:57`:** "`SCOUT_REBUILD_REQUIREMENTS.md`@9a3d19f5: `9efcea1b…`". That is the rev-2.5-era requirements base (base `…/area-refresh-fallback/round-5.md:65`).
  - **`ADR-007/ADR.md:54`:** "`SCOUT_REBUILD_REQUIREMENTS.md`@9a3d19f5 base: `bf6b8d29…`". That is the rev-2.6 requirements (`…/area-refresh-fallback/round-5-response.md:22`). At `9a3d19f5` the requirements are `dab8fa33…` (`ADR-008/ADR.md:52`).
  - **`ADR-001/ADR.md:44`:** cites "`docs/scout-design.md`@9a3d19f5 §15 item 5". At `9a3d19f5`, §15 holds only item 1 (base design :467-471). Item 5 exists in the signed snapshot `docs/adrs/2026-10-21-e1b-v2/scout-design.proposed.md:403`.
  - Round 1's instruction ("pin historical citations… to `@9a3d19f5`") was too blunt for digest inventories. The council owns that part of the cause.
- **Required fix:**
  - In digest inventories, drop `@9a3d19f5` and name each artifact as its round named it. Examples: "`docs/scout-design.md` draft as reviewed in round 8"; "requirements base as reviewed in round 5".
  - Point `ADR-001/ADR.md:44` to `docs/adrs/2026-10-21-e1b-v2/scout-design.proposed.md` §15 item 5.
  - Re-check every other `@9a3d19f5` pin that describes a state at an earlier revision.

### R2-04 (minor): leftover "no change in meaning" wording, and the ADR-009 card omits a decision

- **Location:**
  - `proposal.md`, change-set items 3 and 4;
  - `html:963`;
  - `html:951-957` (card decisions).
- **Evidence:**
  - **`proposal.md` item 3:** "`REQ-1`…`REQ-41` keep their IDs and meaning". This contradicts the proposal's own "Decisions this ADR makes" (REQ-37 changes) and `requirements-map.md:403` ("40 preserve meaning; REQ-37 authority-corrected"). R1-10 named item 3 explicitly.
  - **`proposal.md` item 4:** "Each folder gains an `ADR.md` summary, ADR-001 to ADR-008". It omits the ADR-000 summary (named in the next bullet) and ADR-009's own `ADR.md`.
  - **`html:963`:** "Apart from REQ-37 and REQ-42 above, the restructure moves and condenses text". The same card lists the ended Phase-6 cutover (`:954`) and the deletion of `docs/enaible-findings.md` (`:956`) as decisions.
  - **Card decisions:** the status-line exception, a decision in `ADR.md:16` and the proposal, is not listed. README :3 says the cards summarise the `ADR.md` files.
- **Required fix:**
  - Item 3: "keep their IDs; meaning is unchanged except REQ-37 (see Decisions this ADR makes)".
  - Item 4: name ADR-000 and ADR-009 as well.
  - `html:963`: "Apart from the decisions above, the restructure moves and condenses text".
  - Add the status-line exception to the card's decision list.

### R2-05 (minor): REQ-37 and REQ-42 provenance, and ADR-002's experiment-list relation

- **Location:**
  - REQ-37 (`html:676`) and REQ-42 (`html:529`);
  - the Appendix A rule (`html:519`);
  - the ADR-002 card header (`html:783`) and `ADR-002/ADR.md:3`;
  - the ADR-009 card header (`html:949`).
- **Evidence:**
  - **Appendix A's rule (`:519`):** "A changed requirement keeps its ID and cites the ADR that changed it". REQ-37 cites only ADR-002, but ADR-009 records that it changed REQ-37 (`ADR.md:14`, `html:960`).
  - **REQ-42 (`:529`):** cites only ADR-002, whose decision lists "X1–X6, X9 and X10" (`html:786`, `ADR-002/ADR.md:9`). X11 came from ADR-008 ("Adds planned X11", `html:936`), and ADR-009 chose the X9–X11 list.
  - **ADR-002's status** ("partly superseded by ADR-003, ADR-005, ADR-007, ADR-009") does not name ADR-008.
  - **ADR-009's header** relates to ADR-002 only through "Appendix 12".
- **Required fix:**
  - Add ADR-009 to REQ-37's ADR column, and ADR-008 and ADR-009 to REQ-42's.
  - Add "amended by ADR-008" (X11) to ADR-002's status line and card. Note the matching relation in ADR-008's Changes.
  - Widen ADR-009's ADR-002 relation to include the experiment list.

## Answers to the proposal's questions

1. **Normative content lost, weakened or changed?**
   - R1-01…R1-06 are restored with the base meaning.
   - The disclosed changes are REQ-37 and REQ-42.
   - No new loss was found in the sampled areas. The leftover undisclosed-change wording is R2-04.
2. **Do the ADR summaries and cards match their records?**
   - Decisions and supersession: yes, except R2-05.
   - Digest provenance: no (R2-03).
3. **Does the Overview agree with Appendix A?** The six round-1 items now agree; two restatements do not (R2-02).
4. **Are the rule, tooling and citation changes complete?**
   - No live reference to a deleted file remains unmapped.
   - Rule 2 and the digest pointers conflict (R2-01).
5. **Experiment rules, ledgers, frozen records:**
   - No violation. Experiment rules 1–7 are word for word.
   - Ledgers are append-only (R-062 added, earlier rows unchanged).
   - Frozen rounds are untouched.

## Considered and not raised

- **REQ-99 cites the audit "Backlog recorded" section, which rule 1 does not govern** (peer N6). REQ-99 itself states the one binding trigger (B-3 on the first gating Choice). B-2's tuning and default replacement are governed by REQ-102. The audit labels its backlog "not defects". Editing the backlog therefore cannot change what REQ-99 requires.
- **STATUS stale pointers:**
  - "`AGENTS.md:52-61`" for Secrets (peer N7);
  - "Latest research entry R-047" against "R-001…R-062".
  
  Both were stale at base (base STATUS :133-134, :149), and STATUS is not a signed file. The main agent may fix them along with R2-01 without affecting review.
- **Exit-3 Overview wording** ("a required tool missing"): it matches `cmd/scout/main.go` readiness (not ready → `exitUnavailable`). Base :211's "required blocking lane that could not run → exit 2" covers a different case.
- **ADR-000 "v1 council decision memo"** (`html:755`, `ADR.md:54`): the base memo header says "council decision memo (2026-10-17)". The decisions are recorded as user decisions with no round files. This is consistent.
- **"Supersedes" used where an ADR's own record says "amends" or "narrows" ADR-000:** wording only; the direction of every relation is correct.
- **Push-to-default trigger (`html:323`) omits the opt-in association lanes:** an omission, not a contradiction. REQ-13 and the LLM column (`:354`) cover them.
- **Challenger's severity for the activation denominator (major):** downgraded to minor and merged into R2-02. Appendix A (REQ-22, REQ-78) is correct and now explicitly prevails (`:185`), so the defect is confined to explanatory text.
- **"Adopted 2026-11-06" in signed bytes before sign-off:** follows the precedent R1-07 accepted. If adoption slips past that date, the dates must be updated, and that would be a change needing review.
- **`prereg §3.1` / `§3.3` in REQ-111/REQ-117:** wording carried from base, pointing at existing `evaluation/change-cost-v2` files.
- **Limits:** no shell was used. Nobody hashed files or ran `go build`/`go vet`. Map rows and section keys were sampled, not counted one by one, except the totals the peer counted.

## Peers

- **`innovation-challenger`:**
  - Confirmed R1-01…R1-06, R1-09, R1-11, R1-13–R1-15 and R1-17.
  - Added the activation-denominator and ledger-diagram findings (R2-02), the ADR-001 §15 pin (part of R2-03) and the REQ-37/REQ-42 provenance gap (R2-05).
  - It rated the denominator issue major; the lead downgraded it (see above).
- **`experiment-peer`:**
  - Confirmed the experiment rules word for word and the map counts.
  - Supplied the digest-pin evidence for R2-03 (ADR-006, ADR-007, ADR-008).
  - Added CONTRIBUTING :29 and the rule-2 conflict (R2-01), the proposal item 3/4 wording and the card omission (R2-04), and the ADR-008 → ADR-002 relation (R2-05).
  - Its N6 and N7 were not adopted (see above).
- **Lead:**
  - Found the STATUS :7 and CONTRIBUTING :29 pointers independently.
  - Verified every peer claim against base and frozen records, including `round-8.md:5`, the `area-refresh-fallback` rounds, base §15 and the e1b-v2 snapshot.
- **Effect of the peers:** without them the lead would have reported 2 issues. The peers added R2-02, R2-03 and R2-05. None of the issues is major.

## Verdict

**5 issues (0 major, 5 minor: R2-01 to R2-05): NOT signed off.**

No sign-off applies to the digests above. Every round-1 major issue is resolved; what remains is bookkeeping, provenance and two explanatory restatements. Fix every issue, then submit a round 3 naming new digests for:

- `docs/scout-design.html`
- `proposal.md`
- `requirements-map.md`
- `design-section-map.json`
- ADR-009 `ADR.md`
- `AGENTS.md`
- `docs/adrs/README.md`
- the edited `ADR.md` files of ADR-001, ADR-002, ADR-006, ADR-007 and ADR-008

Round 3 re-checks every fix and anything it touches.
