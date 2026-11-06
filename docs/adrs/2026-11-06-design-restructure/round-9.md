# Round 9: ADR-009, one design document with requirements and ADR appendices

Date: 2026-11-06. This round follows design rules 1–3 in `AGENTS.md`/`CLAUDE.md`. The innovation council lead ran the review with two independent peers:
- **`innovation-challenger` (GPT):**
  - checked the R8-02…R8-06 fixes against their base lines;
  - checked every `CONFIRMED-LOST`, `CONFIRMED-LOOSENED` and `S1` row of `completeness-audit.md` for fidelity, over-reach and the ADR column;
  - searched §6, §7, §11 and §13 for qualifier losses the audit missed.
- **`experiment-peer` (Opus):**
  - checked R8-01 (the JSON against the map, and parent unions);
  - spot-checked the `ALREADY-CARRIED` decisions;
  - compared the Overview with the REQs changed in this round;
  - checked the proposal, the audit and the `ADR.md` records;
  - answered questions 2, 4 and 5.

The review was read-only. It made no model calls outside the council and did not touch `.codegraph`.

**The main agent recomputed the digests; the council did not.** The lead and both peers had read-only tools and no shell, so none of them hashed a file. Earlier rounds accepted a main-agent recomputation, so this round does too. The main agent reported these outputs on the staged worktree:
- `shasum -a 256` on each file below;
- `test -L CLAUDE.md && readlink CLAUDE.md` → `AGENTS.md`;
- `git diff --cached --stat` → `248 files changed, 6500 insertions(+), 1781 deletions(-)`. That is round 8's 246 files plus the new `round-8.md` and `completeness-audit.md`.

The main agent also says no Go file has changed since round 6, when `go build` and `go vet` passed.

| File | SHA-256 (main agent, `shasum -a 256`) | vs round 8 |
|---|---|---|
| `docs/scout-design.html` | `e20c13cfb582c076b59be7392382879fea08845385cc63eedad3b99bf20a8ac2` | changed (R8-02…R8-06, audit restorations) |
| `docs/adrs/2026-11-06-design-restructure/proposal.md` | `af22259b9964d50f71d52b8233327ef8adb90fc3dda4cad18265497c9f015a75` | changed ("Round 8 fixes") |
| `…/requirements-map.md` | `57f247168a90278f131696c257f470e7b7028082e65d523d62430b844a40e086` | unchanged |
| `…/design-section-map.json` | `4a4e3d6e15d505b6b16e523011d8a0e000134e7e12aa59a208fd6fcf2a66da5b` | changed (R8-01) |
| `…/ADR.md` (ADR-009) | `1393b6eb7488688c24ef25106ec6861616ae61e75675570651e5b97ab1fae83a` | changed (records list; see R9-09) |
| `…/completeness-audit.md` | `4699e782188f63ec2b31d0d2255a86d7498daf04a287b8a955240164593f4ad0` | new |
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

**The digest changes match the claimed edits, with one exception.** Round 8 expected only the html, the proposal and the JSON to change. `requirements-map.md` and the 12 unchanged digests match `round-8.md`'s table byte for byte. `ADR.md` also changed: it now lists `completeness-audit.md` (`ADR.md:47`). The main agent disclosed that change in the submission, but `proposal.md` does not record it (R9-09).

## Scope

- **Under review:** worktree `/Users/adamjackson/Projects/work/scout-req`, branch `agent/requirements-path`, with all changes staged. The review covered:
  - the "Round 8 fixes" (`proposal.md:127-137`);
  - every restoration recorded in `completeness-audit.md`;
  - a spot check of its `ALREADY-CARRIED` decisions;
  - the Overview against the REQs changed in this round;
  - a final broad pass over proposal questions 1–5.
- **Base:** `main` at `9a3d19f5` in `/Users/adamjackson/Projects/work/scout/`, with the deleted originals and the frozen `docs/design-reviews/` records.
- **Checks made:**
  - R8-02…R8-06 against base `D:182,377,435,436,444`. `D:n` is line n of the base `docs/scout-design.md`.
  - All 41 `CONFIRMED-*` rows and S1 against the base lines they cite, and their provenance against rescope `proposed.diff:104,110,111,199,238-239`, decisions `proposed.diff:70,79,97,155`, ADR-006 `ADR.md:48` and ADR-008 `proposal.md:13,14,20`.
  - The `ALREADY-CARRIED` decisions 12–15, 25, 28, 33, 35, 38, 64, 85, 92, A4, G1 and G4.
  - The base design §11 Phase 1–5 Scenario and Acceptance lines (`D:377-398`) and §13 X1–X11 (`D:435-444`), clause by clause, against REQ-167–REQ-182 and REQ-89.
  - The JSON against the map on every `Design §…`, Appendix, AL and req row, and the parent unions.
  - Overview `html:170-505`.
  - Live-tree greps for deleted paths, the ledgers, and `AGENTS.md` experiment rules 1–7.
  - The trusted and PR workflows' publish and retention steps (`.github/workflows/scout-guardian-trusted.yml:60-62,193-205`, `scout-guardian.yml:118`).
- **Not reopened:** decisions settled in ADR-000…ADR-008, and rounds 1–8 "Considered and not raised" items unless there is new evidence.

## Verification of round-8 fixes and audit restorations

| # | Evidence (worktree vs base) | Result |
|---|---|---|
| R8-01 | **The five keys are fixed.** JSON `"§3.2"`, `"§3.2 Coverage"`, `"§7"` and `"§7.3"` include REQ-16, and `"§8.1"` is `[REQ-35, REQ-33]`.<br>**Every row matches its key, and there are still 312 keys.** Every `Design §…`, Appendix, AL and req row with a REQ disposition equals its JSON key as a set (peer, full row sweep).<br>**[INFERENCE]** REQ-16 and REQ-33 were appended at the end of their lists, which suggests the JSON was patched by hand. The sets are correct.<br>**Separately, the parent unions are incomplete:** R9-07. | fixed |
| R8-02 | REQ-178 (`html:715`) says "duplicating a same-rule finding in one file", which equals `D:377`. ADR-000 is cited. | fixed |
| R8-03 | REQ-96 (`html:578`) says "a high/normal/low rank bucket cut by rank position within the run", which equals `D:182`. ADR-002 is cited (rescope `proposed.diff:117`). | fixed |
| R8-04 | REQ-169 (`html:686`) says "labels from the REQ-100/REQ-101 protocol on probability-sampled audit units", which equals `D:435`. The same row still has another loss: R9-01. | fixed |
| R8-05 | REQ-177 (`html:694`) says "duplicate pending areas per real new concern, the share of wrong assignments among clusters Jev assigned without escalation", which equals `D:444`. | fixed |
| R8-06 | REQ-170 (`html:687`) and `proposal.md:121,136` say "the E3 protocol with Scout's area reconciliation", which matches `D:436`. | fixed |
| Audit 7, 8 (REQ-8) | The `delta` and `policy` definitions equal `D:92`. ADR-000 is cited. | faithful |
| Audit 9 (REQ-13) | Removing `mapping.refresh.debounce` and rejecting policies that still set it matches `D:109` and ADR-006 `ADR.md:48`. | faithful |
| Audit 10, 31, 11 (REQ-139, REQ-152) | "Enabling is acceptance of `provisional`" matches `D:114,173`. "Reconcile matches it like any area, updating file set, title, Clawpatch ids and origin" matches `D:114`. ADR-008 is cited. | faithful |
| Audit 23, 71 (REQ-101) | "No per-case user review; session rows are `agent` inputs, not labels" matches `D:186`. Introduced by decisions `proposed.diff:70`; ADR-003 is cited. | faithful |
| Audit 24 (REQ-160) | "push (full or incremental; publishes baseline)" equals `D:196`. | faithful |
| Audit 26, 27, 65, 37, S1 (REQ-27) | **Tip fetch, base-scan lane health and both banners:**<br>• "PR runs fetch the tip and resolve merge-base coverage" matches `D:200`.<br>• "the recompute takes its lane health from the base scan" matches `D:200` (decisions `:79`, ADR-003).<br>• "`no baseline available (reason); nothing is classified introduced`" equals `D:204`.<br>• "`no baseline published` before a baseline is published" matches `D:215` and old REQ-27 (`SCOUT_REBUILD_REQUIREMENTS.md:47`).<br>**ADR-000 is absent from REQ-27's column.** Round 5's convention covers this (`round-5.md:110`). | faithful |
| Audit 29 (REQ-26) | "retain workflow artifacts for 90 days" makes `D:200`'s evidence parenthetical into a requirement: **R9-06**. | **over-reach** |
| Audit 30, 67 (REQ-26) | "JSONL history is squash-safe [ASSUMPTION]" equals `D:200,454` and keeps its marker. | faithful |
| Audit 32 (REQ-90) | "gate label routing, not ranking or the gate" equals `D:173`. | faithful |
| Audit 34, 70 (REQ-28) | "an attacker can at most lower a rank" equals `D:176` (v1 `:192`). The association bound equals `D:176`. | faithful |
| Audit 36 (REQ-25) | "repository-level; used by Scout only to publish `scout/state`" equals `D:215` (decisions `:97`, ADR-003). | faithful |
| Audit 48, 101 (REQ-134) | "Subject measurements never change a metric claim" equals `D:314`. | faithful |
| Audit 49, 89 (REQ-41) | "(stratified or clustered bootstrap)" matches `D:320` (rescope `:199`, ADR-002). | faithful |
| Audit 50 (REQ-45) | "exclude line positions and quoted text" matches `D:352`. | faithful |
| Audit 51, 68 (REQ-61) | "remove AWS/OIDC" matches `D:366` and v1 `:274`. It does not prevent the deferred S3 adapter (REQ-26). | faithful |
| Audit 52 (REQ-165) | "The Phase 1 persistence benchmark in `go test ./internal/state`" matches `D:343`. | faithful |
| Audit 66 (REQ-167) | The area-ID retention definition equals `D:394` (decisions `:155`). The same row still has another loss: R9-05. | faithful |
| Audit 69 (REQ-21) | "Replace the `--fail-on` gate with profiles and attribution" matches `D:365`. | faithful |
| Audit 83, 84 (REQ-182) | The 429/timeout/cap disclosures and the critical-severity injection fixture equal `D:398`. | faithful |
| Audit 91 (REQ-138) | "(which defines the line's inputs, repositories and period)" equals `D:328`. | faithful |
| Audit 102 (REQ-180) | "all four guardrail mechanisms … with base/head evidence in all tier-1 languages" matches `D:386`. The same row still has another loss: R9-05. | faithful |
| G2, G7, G8 | **REQ-30** matches `D:481`.<br>**REQ-33**'s "report levels, deltas and trend direction" matches `D:32,225`.<br>**REQ-44**'s scope kinds match `D:88`. | faithful |
| A1–A3 (REQ-22) | Triage of every persisted event, "zero events at window end", "that repo's check to shadow", and "auto-demote it to report for that repo and record it [ASSUMPTION]" equal AL `:46-47,56` and `D:207-208`. | faithful |
| `ALREADY-CARRIED` spot checks | **The decisions are correct:**<br>• 12–15 and 92: REQ-67/68/58/147/19.<br>• 33: REQ-87, rescope `:286`.<br>• 35: REQ-98/169.<br>• 38: REQ-22, `round-1.md:80-83`.<br>• 64: `round-8.md:112`.<br>• 85: REQ-182/86/90.<br>• A4: REQ-78, decisions `:87-88`.<br>• G1, G4: dropped or ADR.<br>**Two recorded reasons are wrong:** R9-09. | decisions correct |
| Audit completeness for §11 and §13 | **The audit missed clauses in §11 and §13.**<br>• It has candidates for `D:343,386,390,394,397,398` and `D:184`. It has none for `D:381-382`, `D:385`, `D:393`, the X1 versioning clause at `D:435`, or `D:438`, `D:440` and `D:443-444`. Six losses remain there: R9-01 to R9-05.<br>• Unlike its old-REQ and allow-list inventories, the audit has no clause-by-clause inventory for §6, §11 or §13. | incomplete |
| Overview vs changed REQs | **These parts are consistent:**<br>• The comparable-base popover (`html:450`): "whole run is `unknown` and nothing is classified as introduced", plus the `guardrails.*` exemption.<br>• The activation callout (`:449`) against REQ-22 and REQ-78.<br>• The ledger diagram (`:455-503`) against REQ-8.<br>• The decision matrix (`:340-368`) against REQ-28, REQ-90, REQ-139 and REQ-152.<br>• "PR runs restore it and never write it" against REQ-27.<br>**The "only trusted pushes" wording is not consistent:** R9-08. | consistent except R9-08 |
| Conditions from round 8 | The main agent recorded the digests, `readlink`, `git diff --cached --stat` and the 312-key count. | met (main-agent outputs accepted) |

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R9-01 | minor | `docs/scout-design.html:686` (REQ-169), `:571` (REQ-89) | **X1 has lost its versioning controls.**<br>• `D:435` says the preregistration fixes "split unit, population, sampling strata and draws, … **versioning of verdicts and test-set reuse**". REQ-169 reduces this to "split/population/sampling/… /**verdict reuse**". "Verdict reuse" is not "versioning of verdicts", and test-set reuse has gone.<br>• `D:175` says "**how Jev-quality statistics are keyed to [the rank-formula and baseline versions] is fixed by X1's preregistration**". REQ-89 stores the versions but leaves out this obligation. REQ-90 covers question-version staleness only.<br>• **[INFERENCE]** Quality statistics could be pooled across rank-formula changes, and test-set reuse could be left unspecified.<br>• Both clauses come from ADR-002 (rescope `proposed.diff:110,295`), which both rows already cite. | Restore the `D:435` list word for word in REQ-169: "estimator, split unit, population, sampling strata and draws, audit floor, `k`, interval, treatment of capped and failed units, versioning of verdicts and test-set reuse, and sample size".<br>Add to REQ-89: "how Jev-quality statistics are keyed to these versions is fixed by X1's preregistration".<br>ADR columns unchanged. |
| R9-02 | minor | `html:688` (REQ-171) | **X3 has lost its precision denominator, its retrieval inputs and two preregistered items.** `D:438` reads:<br>• "precision ≥ 0.9 (**files assigned the correct feature ÷ files assigned any feature**)";<br>• "top-K candidate features **per file (path proximity, import edges)**";<br>• "Estimator, **split**, threshold tuning, candidate-coverage and recall reporting, **interval** and sample size are fixed by its … preregistration".<br>REQ-171 keeps "precision ≥ 0.9" with no denominator, "code-retrieved top-K candidates", and only "preregistered candidate coverage/recall and tuning". REQ-102 supplies the estimator and volumes, but not the split or the interval.<br>**[INFERENCE]** Counting correct `none` answers as successes, or choosing the split after seeing data, would satisfy the condensed text. The denominator is the same class as R8-05.<br>Source: ADR-002, rescope `proposed.diff:300`. | Restore all three clauses as worded in `D:438`. ADR column unchanged. |
| R9-03 | minor | `html:693-694` (REQ-176, REQ-177) | **X11 has lost its label protocol.**<br>• `D:444` requires "**agent-protocol labels (§6 Labels)**".<br>• REQ-100 (`html:582`) covers only "Jev labels", and base `D:186` lists X1/X3/X4/X5/X10 but not X11. So nothing in Appendix A now requires X11's assignment, new-area and guard labels to use the frozen guide, the blind independent `claude -p`/`codex exec` sessions and the adjudicator.<br>• REQ-177 keeps only the labeller-model safeguards. | Add to REQ-176: "labels follow the agent protocol (REQ-100/REQ-101)". ADR column unchanged (ADR-008). |
| R9-04 | nit | `html:688,690,692` (REQ-171, REQ-173, REQ-175) | **The rev-2.1 label volumes have gone.**<br>• `D:438` (X3: "unowned files (rev 2.1: 200)"), `D:440` (X5: "(rev 2.1: 200)") and `D:443` (X10: "Labels: 200 (rev 2.1)") carry planning volumes, like the rev-2.1 thresholds that the same REQs keep.<br>• REQ-98 says each preregistration "confirms/replaces carried rev-2.1 values". For volumes, no REQ now states the value that a preregistration is supposed to confirm or replace.<br>Source: ADR-002 (rescope `:300,302`). | Restore "(rev 2.1: 200 labels)" in each of the three rows as a carried value that the preregistration confirms or replaces. ADR columns unchanged. |
| R9-05 | minor | `html:716` (REQ-179), `:717` (REQ-180), `:713` (REQ-167) | **Three phase scenarios have lost their test frames, as REQ-178 did in R8-02.**<br>• **Phase 2** (`D:381-382`): "`scout init` records **consent off**", "a **rule-digest bump is simulated**", "the **shadow window is simulated (14 days, 5 PR runs)** and a maintainer activates…". Its acceptance also checks digest mismatch → `comparison: recomputed`, with paired merge-base attribution including whole-run and per-family `unknown`, "the same path applies before a baseline is published". REQ-179 has none of these. REQ-27 states the runtime behaviour but does not require the scenario to exercise it.<br>• **Phase 3** (`D:385`): "an LCOV artifact is attached to **one run and not the next**". REQ-168's generic "absent/present artifacts" does not keep the consecutive sequence, which is what tests that coverage is not carried over.<br>• **Phase 4** (`D:393`): "three default-branch merges **in quick succession (new unowned package, new manifest, content-only)**; a **content-only PR**". REQ-167 says only "three following merges/PR". ADR-006 kept "the same three merges" so that the removed structural triggers are tested.<br>**[INFERENCE]** Without these frames, three content-only merges, or a scenario with activation preconfigured, would pass.<br>Sources: v1/e1b-v2 base (ADR-000), Phase 2 acceptance revised by ADR-003 (decisions `proposed.diff:142`), and ADR-006 (`area-refresh-fallback/proposed.diff:100`). All are already in the columns. | Restore all three frames from `D:381-382,385,393`. Cross-references to REQ-27 for the expected attribution are acceptable. ADR columns unchanged. |
| R9-06 | minor | `html:611` (REQ-26); `completeness-audit.md:24` (candidate 29) | **A restoration turned evidence into a requirement.**<br>• `D:200` says "cache/artifacts are never the sole baseline (`retention-days: 90`; req §10)". The parenthetical is evidence for that rule: v1 `:212` says "eviction and 90-day retention … violate req §10", and v1 `:326` lists it under "Evidence".<br>• Round 1 ruled `retention-days: 90` "an operational value, not a requirement" (`round-1.md:255`).<br>• REQ-26 now says "**retain workflow artifacts for 90 days**". The PR workflow keeps its artifacts for 30 days (`.github/workflows/scout-guardian.yml:118`; noted in rounds 4 and 6), so the new requirement contradicts the live implementation that earlier rounds accepted.<br>• The audit records this as `CONFIRMED-LOST` and reports "0 overreach". The challenger judged it faithful; the lead disagrees on the round-1 ruling and the workflow evidence. | Delete "retain workflow artifacts for 90 days" from REQ-26 and keep "never … rely solely on cache/artifacts". Reclassify candidate 29 (`NOT-NORMATIVE` or `OVERREACH`) and correct the totals in the audit (`:175`) and `proposal.md:137`. |
| R9-07 | minor | `requirements-map.md:16,35,60,270` and the matching JSON keys (`design-section-map.json` `"§3.2"` `:153`, `"§8.5"` `:534`, `"Appendix 22"` `:276`, `"AL §1"` `:415`; JSON lines as reported by the peer) | **Four parent rows are not unions of their children, although map `:3` says they are.** The missing IDs:<br>• §3.2 lacks REQ-5 and REQ-7, which are in child "§3.2 Issue" (`:372`).<br>• §8.5 (REQ-35 only) lacks REQ-36, REQ-42 and REQ-129 (`:227,229`), and ADR-001 and ADR-003 (`:381-382`).<br>• Appendix 22 lacks REQ-162 and REQ-163, which are in "Appendix 22(d)" (`:87`).<br>• AL §1 lacks REQ-79…REQ-83, which are in "AL §1 demoted seed checks" (`:280`).<br>**Frozen citations now resolve incompletely,** because README `:29` makes the JSON the resolver. For example, frozen research-ledger rows R-044 and R-046 cite "Appendix 22" for the CI-install owner decision; that resolves without REQ-162/163.<br>This is the R8-01 class. Round 8 checked unions only for §3, §7 and §8. | Add the missing IDs to the four parents, regenerate the JSON from the map, and run a mechanical parent-contains-child check over every parent/child pair. The key count stays 312. |
| R9-08 | minor | `html:198` ("Only trusted pushes write the baseline"), `:308` ("trusted push publishes (CAS)"), `:366` ("from trusted pushes only") | **The Overview says only pushes publish, but Appendix A and the workflow allow every trusted default-branch run to publish.**<br>• REQ-26 (`:611`) says "from trusted default-branch runs only", and REQ-25 (`:696`) says "trusted default-branch jobs alone publish state". Both follow `D:200`.<br>• The trusted workflow publishes on push, schedule and dispatch whenever the exit code is 0 or 1 (`scout-guardian-trusted.yml:3-5,60-62,193-197`).<br>• "Only trusted pushes" therefore tells a reader that nightly runs never publish. This is the R2-02 class: an Overview restatement narrower than its REQ. | Say "trusted default-branch runs" (push, nightly or dispatch) at all three places. Optionally add `workflow_dispatch` to the trigger list (`:321-325`). |
| R9-09 | nit | `completeness-audit.md:3,5,20,23`; `proposal.md:127-137` | **The audit and proposal records are inaccurate in five places:**<br>(a) **Candidate 25's reason is wrong.** It credits ≥ 3/20 auto-demotion to "ADR-003 proposed.diff:91". That line is an unchanged context line. Auto-demotion comes from AL §4 (`scout-allowlist.md:56`, ADR-000), which REQ-22 already cites.<br>(b) **Candidate 28's reason is wrong.** It calls `D:200`'s "none of its comparison identifiers" a "contradictory shorthand". That clause is about manifests without `uninspected` and agrees with `D:362` and REQ-50.<br>(c) **The candidate keys cannot be traced.** They are "findings-file line numbers" (`:5`), but no findings file is kept.<br>(d) **The slice counts and clause counts disagree.** The audit says "Five audit slices" (`:3`) and the proposal says "Seven slices" (`:137`). The proposal's "22 lost and 19 loosened clauses" counts duplicates (31, 65, 67, 68, 71, 89, 101); there are 17 distinct clauses of each kind.<br>(e) **The `ADR.md` edit is not recorded.** "Round 8 fixes" does not mention the change to the records list. | (a, b) Correct both reasons.<br>(c) State that candidate numbers are internal labels, or keep the findings file.<br>(d) Make the slice counts agree, and say "candidates (including duplicates; 17 distinct lost, 17 distinct loosened)".<br>(e) Record the `ADR.md` records-list change. |

## Answers to the proposal's questions

1. **Is normative content lost, weakened or changed?**
   - **Yes.** Five losses remain in §11 and §13: R9-01, R9-02, R9-03 and R9-05 (minor) and R9-04 (nit). One restoration over-reaches (R9-06).
   - **Preserved:** every R8 fix and every other audit restoration matches its base line, and every number in them is exact: 14 days, 5 runs, 10 %, 3/20, 20 lines, 5,000 jobs / 5 MB, ≥ 10 points, LB > 0, 90 % excepted. Each restored clause's introducing ADR is in its column, or it is a v1 clause under the round-5 convention.
2. **Do the ADR summaries and cards match their records?** Yes.
   - The ADR-000…008 digests are unchanged.
   - The final-round numbers on the cards match the base records (peer).
   - ADR-009 `ADR.md` now lists `completeness-audit.md`, which is correct; the proposal does not record that edit (R9-09e).
3. **Does the Overview agree with Appendix A?** Except for R9-08, yes. Banners, activation, the ledger records, the decision matrix and the association wording are consistent with the changed REQs.
4. **Are rule, tooling and citation changes complete, with no live reference to a deleted file?**
   - **Deleted files:** no live reference. The peer grepped `internal/`, `cmd/`, `.github/`, `scripts/`, `toolchain/`, `.omp/`, `.agents/` and the non-frozen `docs/`. The only hits are frozen ledger rows, labelled history, the README path maps and the plan's deletion list.
   - **Citations:** incomplete for four parent sections (R9-07).
5. **Experiment rules, ledgers, frozen records:** no violation.
   - `AGENTS.md:10-31` equals the base word for word, apart from the settled X1–X6, X9–X11 preamble (R1-14).
   - The research ledger only appends R-062.
   - The hypothesis ledger only appends one History line.
   - The frozen records' digests are unchanged.

## Considered and not raised

- **The hero (`html:177`) and the matrix (`:361`) say "introduced" without "regressed".** The peer raised this as a nit. Both are plain-language summaries: a `regressed` issue is one the PR re-introduces. The gate diagram (`:394-395`) and REQ-20 are exact. R2-02 corrected "new in the PR", which did exclude regressed issues; this wording does not.
- **REQ-27 and REQ-61 have v1 clauses without ADR-000 in their columns.** Settled convention (`round-5.md:110`): only missing changers are raised.
- **REQ-26 "JSONL history is squash-safe [ASSUMPTION]".** It is marked as an assumption, so it adds no obligation, and REQ-183 keeps it provisional. That is unlike R9-06, which states an unmarked retention duty.
- **REQ-61 "remove AWS/OIDC" vs REQ-26's deferred S3 adapter.** This removes current prerequisites and does not prohibit the deferred adapter (challenger).
- **X1 "(route-only is not reopened)" (`D:435`).** REQ-169 fully specifies the else-branch (deterministic default, `--rank jev` opt-in), so no other outcome is permitted.
- **The JSON key order suggests hand-patching.** The sets are equal; R9-07's fix regenerates the JSON anyway.
- **The narrow-Noul primitive for X1.** REQ-88 ("narrow questions") and REQ-99 (row 5, B-1) carry it.
- **The Overview omits auto-demotion and rule-bump re-shadowing.** It under-states REQ-22 without contradicting it.

## Peers

- **`innovation-challenger`:**
  - Verified R8-02…R8-06 and all 41 `CONFIRMED-*` and S1 rows, including provenance.
  - Found R9-01, R9-02, R9-03 and R9-05 (Phase 2, 3 and 4 frames).
  - Judged candidate 29 faithful. The lead overruled this as R9-06, on the round-1 ruling and the PR workflow's 30-day retention.
- **`experiment-peer`:**
  - Verified R8-01 across every keyed row, the 312-key count, the audit totals, the 25-row restored list, the `ALREADY-CARRIED` spot checks, the Overview, the cards, the ledgers and the rules.
  - Found R9-07, R9-08 and R9-09.
  - Raised F3 ("introduced" only), which the lead did not adopt.
- **Lead:**
  - Re-verified each peer finding against the base lines and the html: `D:175,381-382,385,393,435,438,440,443-444`, map `:3,16,35,60,87,227,229,270,280,372`, decisions `proposed.diff:91`, the workflow files and `round-1.md:255`.
  - Found R9-04 and R9-06 independently.
  - Checked the Overview and the restored rows REQ-8/13/22/25/26/27/28/61/90/139/152/160/167/180/182 directly.
- **Where the peers changed the outcome:** without the challenger, the lead would have raised R9-04 and R9-06 only, and no §11 or §13 losses. R9-07, R9-08 and R9-09 come from the peer. The peers did not change R9-06; the lead raised it against the challenger's verdict.

## Verdict

**9 issues (R9-01, R9-02, R9-03, R9-05, R9-06, R9-07 and R9-08 minor; R9-04 and R9-09 nits): NOT signed off.**

No sign-off applies to the digests above.

### Required with the next submission

- **Changed digests:** expected only for:
  - `docs/scout-design.html` (R9-01 to R9-06 and R9-08);
  - `requirements-map.md` and `design-section-map.json` (R9-07);
  - `completeness-audit.md` (R9-06 and R9-09);
  - `proposal.md` (a "Round 9 fixes" section and the R9-09 corrections).

  Any other changed digest must be disclosed in `proposal.md`.
- **Recomputed digests and checks:** the main agent must recompute every digest with `shasum -a 256` on the staged worktree. It must also re-run:
  - `test -L CLAUDE.md && readlink CLAUDE.md`, expecting `AGENTS.md`;
  - `git -C /Users/adamjackson/Projects/work/scout-req diff --cached --stat`.

  It must also confirm that the JSON still has 312 keys, and report the output of the parent-contains-child check. `go build`/`go vet` need not be re-run unless Go files change.

---

## Council memo

1. **Recommendation:** fix R9-01…R9-09 in one pass and submit round 10. Change the method for §11 and §13: don't condense those cells again. Carry each base §11 Scenario/Acceptance line and each §13 Design/Decision-rule cell into its REQ near-verbatim, leaving out only run-result narrative and evidence pointers. Then add to `completeness-audit.md` a clause-by-clause inventory for §6, §11 and §13, like the existing old-REQ table, mapping each base clause → the REQ phrase that carries it. Rounds 7, 8 and 9 each found losses in exactly these sections, and candidate-based sweeps keep missing them.
2. **Alternatives:**
   - (a) **Sign off with conditions.** Rejected: design rule 1 requires zero issues.
   - (b) **Leave scenario and experiment detail to future preregistrations.** Rejected for the same reason as round 8 (b). The base fixes these frames, so handing them off is a design change that needs the user's decision and a council review.
   - (c) **Incorporate base §11/§13 at `9a3d19f5` by reference from REQ-167–182.** This would end the class at once. But it makes a deleted document normative again, against R-062's single-document direction. Escalate to the user only if (1) proves impractical.
   - (d) **Keep paraphrasing, with another candidate sweep.** Rejected: three rounds of evidence show it misses ordinary-word qualifiers.
3. **Evidence:** each issue row cites the exact worktree and base `file:line`, and the verification table cites every restoration.
4. **Risks:**
   - **R9-01 to R9-05:** experiments and phase scenarios could pass under weaker frames than the base sets.
   - **R9-06:** REQ-26 conflicts with the PR workflow's 30-day retention.
   - **R9-07:** frozen §3.2, §8.5, Appendix 22 and AL §1 citations resolve incompletely.
   - **R9-08:** a reader could believe nightly runs never publish.
   - **Reversibility:** every fix is a text or JSON edit and is fully reversible.
   - **Unknown:** the council saw no per-file staged diff and relies on the main agent's statement of what changed. Without the inventory in (1), round 10 may find more §11/§13 losses.
5. **Execution outline:**
   1. **REQ-169:** replace its preregistration list with `D:435`'s, word for word.
      - **REQ-89:** add the X1 keying clause from `D:175`.
   2. **REQ-171:** restore the precision denominator, "per file (path proximity, import edges)" and the preregistered split and interval from `D:438`.
      - **REQ-171/173/175:** add "(rev 2.1: 200 labels)".
   3. **REQ-176:** add "labels follow the agent protocol (REQ-100/REQ-101)".
   4. **REQ-179, REQ-180 and REQ-167:** restore the Phase 2, 3 and 4 frames from `D:381-382,385,393`.
   5. **REQ-26:** delete "retain workflow artifacts for 90 days".
      - **Audit:** reclassify candidate 29 and correct the totals in the audit and the proposal.
   6. **Map and JSON:**
      - add the missing IDs to the §3.2, §8.5, Appendix 22 and AL §1 parents;
      - regenerate the JSON from the map;
      - run the parent-contains-child check;
      - confirm 312 keys.
   7. **Overview (`html:198,308,366`):** change "pushes" to "default-branch runs".
   8. **Audit and proposal (R9-09 a–e):** correct candidates 25 and 28, the key note, the slice and clause counts, and the `ADR.md` disclosure.
   9. **Clause inventory:** build the §6/§11/§13 inventory described in point 1, and restore anything it finds without a carrier.
   10. **Proposal:** add "Round 9 fixes".
   11. **Checks:** recompute every digest with `shasum -a 256`, run `readlink` and `git diff --cached --stat`, and submit round 10 with the outputs.
