# Round 8: ADR-009, one design document with requirements and ADR appendices

Date: 2026-11-06. This round follows design rules 1–3 in `AGENTS.md`/`CLAUDE.md`. The innovation council lead ran the review with two independent peers:
- `innovation-challenger` (GPT) checked whether each round-7 fix and each sweep restoration says what the base says, and searched for qualifier losses the sweep missed.
- `experiment-peer` (Opus) checked whether the Overview matches Appendix A, the map and its JSON twin, the accuracy of the proposal, provenance, and the broad questions 2, 4 and 5.

The review was read-only. It made no model calls outside the council and did not touch `.codegraph`.

**The main agent recomputed the digests; the council did not.** The lead and both peers had read-only tools and no shell, so none of them hashed a file. Earlier rounds accepted a main-agent recomputation, so this round does too. The main agent reported these outputs on the staged worktree:
- `shasum -a 256` on each file below;
- `test -L CLAUDE.md && readlink CLAUDE.md` → `AGENTS.md`;
- `git diff --cached --stat` → `246 files changed, 6116 insertions(+), 1781 deletions(-)`. That is round 7's 245 files plus the new `round-7.md`.

The main agent also states that no Go file has changed since round 6, when `go build` and `go vet` passed.

| File | SHA-256 (main agent, `shasum -a 256`) | vs round 7 |
|---|---|---|
| `docs/scout-design.html` | `3201ea248b2a7fb592cb6358a5c36ee995034245c5f707c533612ba53d5962b7` | changed (R7-01, R7-02, sweep) |
| `docs/adrs/2026-11-06-design-restructure/proposal.md` | `1b9b448ae4a35d95c796bb50065bb6facb1402f32a77c6ecf988d25e11cf6ad5` | changed ("Round 7 fixes") |
| `…/requirements-map.md` | `57f247168a90278f131696c257f470e7b7028082e65d523d62430b844a40e086` | changed (REQ-16, REQ-33 rows) |
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

**The digest changes match the claimed edits.** Round 7 predicted that the html and the proposal would change, and that `requirements-map.md` would change if the map was updated. It was, and the change is disclosed at `proposal.md:125`. The other 14 digests match the `round-7.md` table byte for byte.

## Scope

- **Under review:** worktree `/Users/adamjackson/Projects/work/scout-req`, branch `agent/requirements-path`, with all changes staged. The review covered:
  - the "Round 7 fixes" (`proposal.md:110-125`);
  - the eight sweep restorations at each location they touched;
  - the map changes;
  - a check for any new inconsistency between the Overview and REQ-9/REQ-16;
  - a final broad pass over proposal questions 1–5.
- **Base:** `main` at `9a3d19f5` in `/Users/adamjackson/Projects/work/scout/`, including the deleted originals and the frozen `docs/design-reviews/` records.
- **Checks made:**
  - **REQ-9 and REQ-16:** worktree `html:536,553` against base design `:90,102,211`.
  - **REQ-33:** `html:620` against base `:225`. The sweep notes cite `:224`, but the sentence is at `:225`. The sweep notes are not a reviewed file.
  - **REQ-87:** `html:569` against base `:174`.
  - **REQ-96:** `html:578` against base `:182`.
  - **REQ-169 and REQ-170:** `html:686-687` against base `:435-437`. E3's protocol was checked against `evaluation/map-stability-v1/RESULTS.md:97,103` and `evaluation/x2-area-stability/preregistration.md:11-13,79-82`.
  - **REQ-174:** `html:691` against base `:441-442`.
  - **REQ-176 and REQ-177:** `html:693-694` against base `:444`.
  - **REQ-178:** `html:715` against base `:377-378` and REQ-6 (`html:533`, base `:98`).
  - **Map:** rows `:14,16,25,28,30,31,373` and the counts (388 = 323 REQ + 43 ADR + 22 dropped).
  - **JSON twin:** `design-section-map.json:126-159,528-530,1385-1432,2022-2025`, against map rows, `ADR.md:28` and README `:29`.
  - **Overview:** `html:440-442,455,494-505,189,263,374`.
  - **Provenance:** traced through the base `docs/design-reviews/*/proposed.diff` files and the e1b-v2 base snapshot.
  - **Live-tree greps:** for deleted paths and for code comments citing the changed sections; base `internal/decisions/run.go:20` is now `REQ-93`.
  - **Records:** ledgers, `AGENTS.md` experiment rules, and the ADR-000…008 digests.
- **Not reopened:** decisions settled in ADR-000…ADR-008, and rounds 1–7 "Considered and not raised" items unless there was new evidence.

## Verification of round-7 fixes and sweep restorations

| # | Evidence (worktree vs base) | Result |
|---|---|---|
| R7-01 | REQ-174 (`html:691`) now says "replays the last 200 merged PRs from 3 adopter repositories/three languages". Base `:441` says "Replay last 200 merged PRs of 3 adopter repos, three languages". | fixed |
| R7-02 | **REQ-16 (`html:553`)** adds "a required blocking lane that could not run on head gives exit 2; optional lanes never turn a required lane red". That is base `:211` word for word in meaning, and it keeps the base `:102` exit 2/3 and exit 1 split.<br>**REQ-9 (`:536`)** adds "Coverage gates only as 'required lane incomplete' (exit 2)". That is base `:90`.<br>Both ADR columns are unchanged and correct: the clauses date from v1, which is ADR-000. | fixed |
| Sweep REQ-33 | "these measures never gate, set severity, forecast or compute a composite" matches base `:225`.<br>Base `:225` also says Scout never "upgrades `claim` on its own". REQ-35 already carries that ("Scout never upgrades claims itself"). | faithful |
| Sweep REQ-87 | "discarding and counting mismatches as unit `incomplete`" matches base `:174` ("discarded (unit `incomplete`, counted)"). The clause comes from ADR-002 (rescope `proposed.diff:109`), which is already in the ADR column. | faithful |
| Sweep REQ-96 | "After X1's preregistration is signed off" matches base `:182` ("Once X1's preregistration is signed off"). The clause comes from ADR-002 (`proposed.diff:117`).<br>The requirement still has a separate loss: R8-03. | restoration faithful |
| Sweep REQ-169 | "compares triage retrieval at a fixed review budget" matches base `:435`. The ≥ 10-point threshold, the interval lower bound > 0 and the audit-floor condition are exact.<br>The requirement still has a separate loss: R8-04. | restoration faithful |
| Sweep REQ-170 | **The numbers are exact:** fastapi/axum, 3 fresh, 2 same-state and 2 mutations; X2b uses the same repositories, revisions, schedule, grid and guide, codex arm only. That matches base `:436-437`, and ADR-005 is in the column (`codex-luna-mapping/proposed.diff:85`).<br>**The phrase "E3's reconciliation protocol" credits E3 with a reconciliation step it did not have:** R8-06. | defect (nit) |
| Sweep REQ-174 (X9) | "hand-checks 200 moved/pasted/new labels and 200 stale-reference hits on the corpus" matches base `:442`, and the Wilson LB ≥ 0.9 threshold is exact. | faithful |
| Sweep REQ-176 | "council-signed preregistration before any X11 run" and "Freeze thresholds/prompt/guard before data" match base `:444`. ADR-008 is in the column. | faithful |
| Sweep REQ-178 | "runs twice, the second after inserting 20 lines…" matches base `:377`.<br>The same sentence still drops "in one file": R8-02. | restoration faithful |
| Map changes | **The disclosed rows are in place:** REQ-16 is in §3.2 (`:16`), §7 (`:25`), §7.3 (`:28`) and §3.2 Coverage (`:373`); REQ-33 is in §8.1 (`:31`).<br>**Parent rows** §3, §7 and §8 are still unions of their children.<br>**Counts are unchanged:** 388/323/43/22.<br>**The JSON twin was not updated:** R8-01. | partly |
| "Round 7 fixes" section | Each REQ description matches the html, and each "(base §N)" falls under the right base heading: `:225` is §8.1, `:174`/`:182` are §6, `:435-444` are §13, and `:377` is §11. "IDs and ADR columns unchanged" holds.<br>`:121` repeats the R8-06 wording, and `:125` does not mention the JSON. | accurate apart from R8-01/R8-06 |
| Overview vs REQ-9/16 | **No contradiction:**<br>• The exit legend (`html:441`) says "2 incomplete: a required lane did not finish".<br>• The Coverage arrow (`:494-495,505`) says "Marks the run incomplete / exit 2".<br>• "Only an issue can block" (`:455`) is consistent with REQ-8 and REQ-33.<br>• Non-goal `:189` is consistent with REQ-4 and REQ-33.<br>The Overview does not mention X1/X2/X6/X9/X11, buckets, model mismatch or the Phase 1 scenario. | consistent |
| Conditions from round 7 | The main agent recorded the digests, `readlink` and `git diff --cached --stat`. | met (main-agent outputs accepted) |

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R8-01 | minor | `docs/adrs/2026-11-06-design-restructure/design-section-map.json:126-151` (`"§7.3"`), `:152-159` (`"§3.2"`), `:528-530` (`"§8.1"`), `:1385-1432` (`"§7"`), `:2022-2025` (`"§3.2 Coverage"`) | **The citation index no longer matches the map on the five rows that round 8 changed.**<br>• Map `:16,25,28,373` now include REQ-16, and `:31` includes REQ-33. The JSON keys still lack them; for example, `"§8.1"` is `["REQ-35"]` and `"§3.2 Coverage"` is `["REQ-9","REQ-11"]`.<br>• Before this round those keys matched the map rows exactly. The JSON digest has been `3cc0225e…` since round 1.<br>• README `:29` and ADR-009 `ADR.md:28` name the JSON as the resolver: for frozen `design §n` citations it "gives the current REQ and ADR IDs".<br>• So a frozen citation of base §7.3 `:211` (the R7-02 clause) does not reach REQ-16. Frozen `2026-10-21-e1b-v2/round-4.md:21` cites §8.1 for "population and missingness stated", which is REQ-33 content, and the JSON sends it to REQ-35 only (peer).<br>• `proposal.md:125` does not mention the JSON. | Add REQ-16 to JSON keys `§3.2`, `§3.2 Coverage`, `§7` and `§7.3`, and add REQ-33 to `§8.1`, so they match the map rows. Keep 312 keys. Disclose the change in the proposal and report the new JSON digest. |
| R8-02 | minor | `docs/scout-design.html:715` (REQ-178) | **The Phase 1 scenario has lost its same-file qualifier.**<br>• Base `:377` reads "duplicating a same-rule finding **in one file**". The acceptance criterion at `:378` is "two IDs for two occurrences", and base `:98` makes "Distinct occurrences of one rule in one file" distinct issues.<br>• REQ-178 says only "duplicating a same-rule occurrence". Map `:163-165` sends the scenario here.<br>• **[INFERENCE]** A duplicate placed in another file gets separate IDs trivially, because the path differs. The scenario would then never test the same-file case, which is the hard one. This is the same class as R7-01: a loosened test frame.<br>• The sweep edited this sentence and left this qualifier out. | Change it to "duplicating a same-rule finding in one file". ADR column unchanged (ADR-000; e1b-v2 base `:304`). |
| R8-03 | minor | `docs/scout-design.html:578` (REQ-96) | **The rank buckets have lost their population.** Base `:182` reads "a rank bucket (high / normal / low) cut by rank position **within the run**". REQ-96 says only "rank-position buckets". Map `:215` sends §6 Calibration only to REQ-96, and no other REQ restores the clause. **[INFERENCE]** Buckets cut across runs or across history would satisfy the condensed text but not the base. | Restore "cut by rank position within the run". ADR column unchanged (ADR-002, rescope `proposed.diff:117`). |
| R8-04 | minor | `docs/scout-design.html:686` (REQ-169) | **X1 has lost its required label frame.**<br>• Base `:435` reads "Labels: §6 Labels protocol on **probability-sampled audit units**". REQ-169 leaves "population/sampling" to the preregistration. Map `:118` sends X1 only to REQ-169.<br>• REQ-94 governs nightly audits and REQ-138 governs scorecard lines. Neither requires X1's evaluation labels to come from those units, and REQ-102 only says the preregistration "fixes … labels".<br>• **[INFERENCE]** An X1 run labelled on a convenience sample would satisfy the requirement as now written. | Add that X1's labels follow REQ-100/101 on probability-sampled audit units. ADR column unchanged (ADR-002, rescope `proposed.diff:295`). |
| R8-05 | minor | `docs/scout-design.html:694` (REQ-177) | **The X11 reporting units have lost their denominators.** Base `:444` requires "duplicate pending areas **per real new concern**" and "the **share** of wrong assignments on clusters Jev assigned without escalation". REQ-177 says only "duplicate pending areas, wrong un-escalated assignments". Map `:127` sends X11 only to REQ-176/177. **[INFERENCE]** Raw counts would satisfy the condensed text; the base requires rates normalised by concern and by un-escalated assignments. | Restore both: "duplicate pending areas per real new concern" and "the share of wrong assignments among clusters Jev assigned without escalation". ADR column unchanged (ADR-008). |
| R8-06 | nit | `docs/scout-design.html:687` (REQ-170); `proposal.md:121` | **"E3's reconciliation protocol" credits E3 with reconciliation it did not have.**<br>• Base `:436` reads "E3 protocol **with** reconciliation": E3's map schedule plus Scout's area reconciliation.<br>• E3 had none: its RESULTS `:97` say "nothing reconciles them against prior features", and `:103` says reconciliation "would need a Scout-side reconciliation layer". X2's preregistration (`:11-13`, `:82`) adds Scout's `areas.Reconcile` on top of E3's protocol.<br>• This wording was introduced by this round's sweep. | Change both places to "the E3 protocol with Scout's area reconciliation" (or the base's "E3 protocol with reconciliation"). ADR column unchanged. |

## Answers to the proposal's questions

1. **Is normative content lost, weakened or changed?**
   - **Lost or weakened:** four qualifier losses remain from the original migration (R8-02 to R8-05), and one restoration changes the meaning of an attribution (R8-06). The sweep missed all four losses. Three of them sit in requirements the sweep itself edited (REQ-96, REQ-169, REQ-178).
   - **Preserved:** every other restoration matches the base, and every number in the changed REQs matches: 200/400, 3/2/2, ≥ 10 points, LB > 0, ≥ 0.8 / ≤ 5 %, Wilson LB ≥ 0.9, ≥ 3 repositories, ≥ 2 mapping commits, 20 lines.
2. **Do the ADR summaries and cards match their records?** Yes. The ADR-000…008 digests are unchanged. The ADR-008 card's "council sign-off on X11's preregistration before any run" (`html:948`) agrees with REQ-176. The provenance of every restored clause is already in its REQ's ADR column (peer table: v1/ADR-000; rescope `:109,117,295`/ADR-002; codex-luna-mapping `:85`/ADR-005; revision 2.8/ADR-008).
3. **Does the Overview agree with Appendix A?** Yes. No new divergence was found; see the verification table.
4. **Are rule, tooling and citation changes complete, with no live reference to a deleted file?**
   - **Deleted files:** no live reference. Grepping `internal/`, `cmd/`, `.github/`, `scripts/`, `toolchain/`, `.omp/`, `.agents/` and `.prettierignore` finds none, and the `.omp` files name the new paths.
   - **Citations:** incomplete. The JSON resolver is stale on five keys (R8-01). No live code comment depends on those keys. Base `decisions/run.go:20` "(design §7.3)" now cites REQ-93, which holds the optional decisions-lane rule.
5. **Experiment rules, ledgers, frozen records:** no violation.
   - Experiment rules 1–7 match base word for word.
   - The research ledger is unchanged up to `:69`, and R-062 is appended at `:70`.
   - The hypothesis ledger is unchanged since round 7.
   - The frozen records are untouched; their digests are as in round 7.

## Considered and not raised

- **REQ-87 omits "at config load".** REQ-99 incorporates the Met/Fixed adapter practices, and `jev-adapter-audit.md` row 13 requires rejection at config load (challenger).
- **REQ-138 omits "signed off".** REQ-94 keeps the Jev-quality lines `insufficient` until X1 signs off (challenger).
- **The proposal calls X9's 200 labels plus 200 hits "its 400 labels" (`proposal.md:122`).** It is a paraphrase in a change log; each hand-check yields a label, and the html is exact (peer).
- **REQ-6's "distinct occurrences stay distinct" is broader than base `:98` "of one rule in one file".** Broadening a protection loses nothing. The loss is in REQ-178's scenario frame (R8-02).
- **Exit 3 "required tool missing" vs REQ-16's "could not run → exit 2".** Settled in round 2 (`round-2.md:187`); no new evidence.
- **The Overview's "Rank: Jev P(real defect)" (`html:263`) vs REQ-96's "never the probability number".** The label names the ranking input, not a displayed value.
- **Older differences between the JSON and the map (`"§14 A10"` → REQ-38; `"req REQ-5"`).** They predate round 8 and look deliberate for a citation index (peer).
- **Earlier map changes (R3-03, allow-list rows 3–6).** The JSON has no per-row allow-list keys, so they did not cause drift.

## Peers

- **`innovation-challenger`:**
  - Verified REQ-9, REQ-16, REQ-33, REQ-87, REQ-174 (X6 and X9) and REQ-176 as faithful, with exact numbers and correct provenance.
  - Found R8-03 (REQ-96 "within the run"), R8-04 (REQ-169 probability-sampled audit units) and R8-05 (REQ-177 denominators).
  - Found R8-02 independently of the lead.
  - Judged REQ-170 faithful. The lead disagrees on the E3 attribution (R8-06), on line-level evidence from E3 RESULTS `:97,103`.
- **`experiment-peer`:**
  - Verified the Overview against REQ-9 and REQ-16, the map rows, parent unions and counts, the accuracy of the proposal, and the provenance of every restored clause.
  - Ran the live-tree, `.omp`, ledger and `AGENTS.md` checks.
  - Found R8-01 independently of the lead, and added the concrete frozen-citation case (e1b-v2 `round-4.md:21`).
- **Lead:**
  - Verified every R7 fix and sweep restoration against the base lines.
  - Found R8-01: the JSON vs map comparison, after reading README `:29` and `ADR.md:28` on the JSON's role.
  - Found R8-02: base `:377` vs `html:715`.
  - Found R8-06: base `:436` vs E3 RESULTS and the X2 preregistration.
  - Confirmed R8-03, R8-04 and R8-05 against base `:182,435,444` and checked that no other REQ restores them (grep for "within the run", "probability-sampled", "per real new concern" and "share of wrong" in the html).
- **Where the peers changed the outcome:** without the challenger, the lead would have raised R8-01, R8-02 and R8-06 only. R8-03, R8-04 and R8-05 come from the challenger. The peer confirmed R8-01 and showed that the drift starts in this round, so it does not reach back to earlier rounds.

## Verdict

**6 issues (R8-01 to R8-05 minor, R8-06 nit): NOT signed off.**

No sign-off applies to the digests above.

### Required with the next submission

- **Changed digests:** expected only for:
  - `docs/scout-design.html` (R8-02 to R8-06);
  - `proposal.md` (a "Round 8 fixes" section and the `:121` wording);
  - `design-section-map.json` (R8-01).

  `requirements-map.md` should stay at `57f24716…` unless the main agent changes it, in which case the submission must disclose it. Every other digest should stay as in the table above.
- **Recomputed digests and checks:** the main agent must recompute every digest with `shasum -a 256` on the staged worktree. It must also re-run:
  - `test -L CLAUDE.md && readlink CLAUDE.md`, expecting `AGENTS.md`;
  - `git -C /Users/adamjackson/Projects/work/scout-req diff --cached --stat`.

  Record the outputs. It should also confirm that the JSON still has 312 keys. `go build`/`go vet` need not be re-run unless Go files change.

---

## Council memo

1. **Recommendation:** fix R8-01…R8-06 in one pass and submit round 9.
   - **Restore the qualifiers word for word from the base**, with no new IDs and no ADR-column changes.
   - **Regenerate `design-section-map.json` from `requirements-map.md`** rather than editing it by hand, so the two cannot drift again.
   - **Replace the token-based sweep with a phrase-level comparison** for the base areas where all of this round's losses sat: §6 (Calibration, Pinned model, Failures), §11 (every Phase Scenario and Acceptance line) and §13 (every column of every X row, including the Design/Labels column). For each base clause, name the REQ phrase that carries it or the REQ that supplies it. Any clause without one is a restoration. A token sweep cannot catch these losses, because the lost words ("within the run", "in one file", "per real new concern", "probability-sampled") are ordinary words, not binding-qualifier tokens.
2. **Alternatives:**
   - (a) **Sign off with conditions.** Rejected: design rule 1 requires zero issues.
   - (b) **Treat R8-03, R8-04 and R8-05 as preregistration matters** (REQ-102 says preregistrations fix labels and sampling). Rejected for the same reason as R7-01: the base design fixes these frames and units, and the restructure must not hand them to future preregistrations without saying so.
   - (c) **Add cross-references instead of restoring text** (for example, REQ-169 "labels per REQ-94/100"). Acceptable only if the referenced REQ states the exact frame. For R8-04 it does not fully do so, so restoring the text is simpler.
   - (d) **Leave the JSON stale and document that the map prevails.** Rejected: README `:29` names the JSON as the resolver.
3. **Evidence:** each issue row cites the exact worktree and base `file:line`, and the verification table cites every restoration location.
4. **Risks:**
   - **R8-02:** the Phase 1 identity test could pass without testing same-file duplicates.
   - **R8-03:** buckets could be cut across runs.
   - **R8-04:** X1 could be evaluated on a non-probability sample.
   - **R8-05:** X11 could report raw counts that cannot be compared across repositories.
   - **R8-01:** frozen §7.3 and §8.1 citations resolve to the wrong requirements.
   - **R8-06:** a reader could credit E3 with a reconciliation result.
   - **Reversibility:** all fixes are text or JSON edits and are fully reversible.
   - **Unknown:** the council saw no per-file staged diff and again relies on the main agent's statement of what changed. The challenger's search was adversarial but not exhaustive, so round 9 may find more losses of this class unless the phrase-level comparison in step 3 of the outline is done.
5. **Execution outline:**
   1. REQ-178 (`html:715`): change "duplicating a same-rule occurrence" to "duplicating a same-rule finding in one file".
   2. REQ-96 (`:578`): change "rank-position buckets" to "buckets cut by rank position within the run".
   3. REQ-169 (`:686`): add "on probability-sampled audit units labelled per REQ-100/101".
   4. REQ-177 (`:694`): change to "duplicate pending areas per real new concern, the share of wrong assignments among clusters Jev assigned without escalation".
   5. REQ-170 (`:687`) and `proposal.md:121`: change to "the E3 protocol with Scout's area reconciliation".
   6. Regenerate `design-section-map.json` from the map. Confirm that `§3.2`, `§3.2 Coverage`, `§7` and `§7.3` include REQ-16, that `§8.1` includes REQ-33, and that there are still 312 keys.
   7. Run the phrase-level comparison of §6, §11 and §13 described in point 1, and record any further restorations.
   8. Add "Round 8 fixes" to `proposal.md`, listing the JSON change.
   9. Recompute every digest with `shasum -a 256`, run `readlink` and `git diff --cached --stat`, and submit round 9 with the outputs.
