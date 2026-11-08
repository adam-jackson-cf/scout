# Round 8: ADR-012 (shadow-window readiness inputs)

Date: 2026-11-08. This round follows design rules 1–3 and the ledger rules in `CLAUDE.md`, and the ADR rules in `docs/adrs/README.md`. ADR-012 is an agent decision made under the owner's standing delegation (R-064). Where a fix involves a choice, the council names the option it recommends under that delegation.

The round reviews the bytes revised after round 7 (`round-7-response.md`). It has three parts:

1. Check every ADR012-R7-01..09 fix against its required fix.
2. Re-confirm that every earlier closure (R1–R6) still holds in the current bytes.
3. Look for new issues, numbered ADR012-R8-xx.

The council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review. It tried to find deadlocks and unreachable gates in the single late-evidence lifecycle. It also probed for loopholes in import re-stamping, attempt-ID order and exclusion causes, and checked how the signed files mirror ADR.md and the R7 closures.
- **`experiment-peer` (Opus):** methodological review. It built timelines with concrete provider event times, attempt starts, publications, C advances, imports and triage. It checked whether the rules give one answer, and whether the acceptance fixtures can actually expose the failures they target.

The review was read-only. Nobody edited files, executed product code, ran model calls outside the council, or touched `.codegraph`. The lead read the required codegraph skill references. Codegraph was not used, because the material under review is documents and the council had no shell. Searches used `read` and `grep`. The lead ran one web search, and the experiment-peer cited the same page: both confirmed GitHub's 30-day expiry of fork runs awaiting approval (ADR012-R8-07).

**Main computed and verified the digests below with `shasum`. The council did not: the lead and both peers had read-only tools and no shell.** This round is not a sign-off, so the digests only identify the bytes reviewed. All paths are under `docs/`.

| File | SHA-256 (Main) |
|---|---|
| `adrs/2026-11-08-shadow-readiness-inputs/ADR.md` | `079fff315519d17ee6695c6b05cca2495696a87e0a91016e2ae9a305c1b7fd29` |
| `scout-design.html` | `259473736d7ab402f0da0bbd6b3de1652526691fbae8eb4376fa3ccd8949a14d` |
| `adrs/README.md` | `cde415ff10b659a073d010f8a48c281d6f75a5a1cba466962d589e937b94587f` |
| `STATUS.md` | `cf667e7251454a6557e1030639e0b30547dbb6472ba73f36537be936f2dc0453` |
| `plans/2026-11-06-remaining-work.md` | `ad898091d0a61a46195e72cf28f1519aa6c3ac1a6de0198d24732517fa4ce8b9` |
| `references/policy-file.md` | `508e92e0fde1ba2277943f4b338e61394ee027ca80f2c9d1cf5fd10332cd535f` |
| `references/ci-integration.md` | `8674d70b818d34490f7c94dca71683c5b06906ac19a2de089de9bb26f895b5f1` |
| `references/cli-commands.md` | `4f0b70fb55c2e156b1e6521e68ca99f2d49370ab447e8e6bc50328889b728cc6` |
| `research-ledger.md` | `1904a6ff989c38b6e5fe52ea62eaf96afc7051b0e4496eea86961a2d653a4409` |
| `adrs/2026-10-19-v1-decisions/ADR.md` (ADR-000) | `1cd521fd96439d87275ba353f7b05451c3e0c4d556ee665529aab85970ed81cb` |
| `adrs/2026-10-23-decisions/ADR.md` (ADR-003) | `973d24c4cfdfdd98acb4070a7bd72ca0c0e1ddf303bd4ccec8f92fe1cf117cc5` |
| `adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) | `ee593b6241814c24035671d666566bc2b7c399135867226c26e9626f3936abc8` |
| `adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) | `80e3375f81656a8856cf662cbec99a36d8e00bc451ff3c3592925d2293f44d95` |

The following record is not signed. Main computed its SHA-256:

| File | SHA-256 (Main) |
|---|---|
| `round-7-response.md` | `6b6bb51eff69778e076ae781abcc939de46c8a19fb1b1fde880ac86ff271cec5` |

README, STATUS and the four historical ADR files have the same digests as in round 7.

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11`, branch `agent/policy-shadow-adrs`, based on origin/main `7bb73e52`. Every change is staged and uncommitted.
- Base bytes come from `/Users/adamjackson/Projects/work/scout-adr11b`.

**Files read.** Line numbers below refer to the raw worktree files.
- ADR-012 `ADR.md:1–68`, all lines:
  - `:3–6` status, revision and sources; `:10–12` decision and delegated narrowing;
  - `:16–22` items 1–7, where `:19` is the "Clocks, frontier and late evidence" paragraph;
  - `:26` triage authorization; `:28` triage rows; `:30` ShadowReady;
  - `:34–40` persistence, manifest/merge, cutover and writer boundary;
  - `:44` CLI; `:48–52` amendments, sources and route;
  - `:56–62` implementation acceptance invariants; `:66–68` folder records.
- `round-7.md` (all 416 lines) and `round-7-response.md` (all lines).
- `scout-design.html`, raw:
  - the Activation callout `:449`;
  - REQ-22 `:589`, REQ-78 `:600`, REQ-10 `:610`, REQ-26 `:611`, REQ-27 `:612`, REQ-48 `:615`, REQ-51 `:618`;
  - grep context for REQ-39 `:704`, REQ-161 `:707`, REQ-166 `:712`, and REQ-160/162/163/168 `:706–714`;
  - the ADR-012 card `:1009–1011`.
- `cli-commands.md:120–134`, `policy-file.md:76–84` and `ci-integration.md:97–106`.
- Plan `:61,90,92`; STATUS `:9,121`; README `:28–30`; `research-ledger.md:73–76`.
- The challenger compared the four historical ADR bodies with base by reading the raw text.

## Issue summary

**7 distinct issues are open: 2 major and 5 minor.**

| ID | Severity | Summary |
|---|---|---|
| ADR012-R8-01 | major | The current-window refusal has no defined scope. On the literal reading, current certification is never reachable. The refusal is also not recorded anywhere the activation evidence can show, `--as-of C` gets round it, and refusals have no assigned exit codes |
| ADR012-R8-02 | major | Fresh stamps on imported history can bring back an older verdict that a later one had superseded. A multi-row import breaks "one row per event per publication". The import authorization rule is ambiguous, and a same-lineage restore gets re-stamped |
| ADR012-R8-03 | minor | "Late" is undefined, and the publication-visibility predicate is stated only for the vetoes. Default-version and representative selection can leak late facts into historical cutoffs |
| ADR012-R8-04 | minor | A late "point gap" can be read two ways: it ends the interval, or it only punctures it. The puncture reading lowers the unknown share |
| ADR012-R8-05 | minor | A manual cancellation followed by a later push still meets the supersession exclusion test |
| ADR012-R8-06 | minor | The R7-02 and R6-03 fixtures can pass because current-window refuses, without ever exercising the veto |
| ADR012-R8-07 | minor | An attempt has no gap path when it is "not yet terminal", so a fork run awaiting approval can hold C back for up to 30 days |

## Closure of round-7 issues

"Fixed" means the required document fix is present in the signed bytes at every named location. It does not mean the implementation is validated. "Fixed as prescribed" means the council's own prescription was applied but left a gap, which is carried as a new R8 issue. "Partly" means part of the required behaviour is still missing.

| Issue | Status | Evidence / remainder |
|---|---|---|
| ADR012-R7-01 | **Fixed** | The discharge clause ("every attempt started≤t terminal/published or covered by a published late record") is in `ADR.md:19`, REQ-78 `:600` and `policy-file.md:82`. Each class has its coordinates: `event_at` for an event, and `started_at` plus the accounted `event_at` for an attempt (`ADR.md:19`). Only the trusted CAS stamps; local late facts stay pending and uncounted (`ADR.md:19`, REQ-48 `:615`, `cli:126`). The live envelope covers late records stamped after `as_of` and pending late facts (`ADR.md:44`, `cli:130`). Fixtures: `ADR.md:58`, plan `:61`, `ci:106`. Both peers rate it closed. The refusal this fix introduced is ADR012-R8-01. |
| ADR012-R7-02 | **Fixed as prescribed** | Late attempts, links and events are persisted, flagged and stamped. Their events join both cohort-wide vetoes "whatever its interval, gap or late status" (`ADR.md:19–20`, REQ-22 `:589`, REQ-78, `cli:128`, `policy-file:80`). Imported history is freshly stamped with its provenance, deduplicated by action ID, and its authorization is verified (`ADR.md:19,36`, REQ-48). The prescribed re-stamping lets an older verdict come back into force: ADR012-R8-02. The fixtures do not isolate the veto: ADR012-R8-06. |
| ADR012-R7-03 | **Fixed** | Attempts durably name the event IDs and times they account for. Binding requires the same PR key, the event's head SHA and `started_at` ≥ `event_at`. Exclusion records name their event, and "A delay alone is not a gap" (`ADR.md:17`, REQ-27 `:612`, REQ-78, `ci:101`, `cli:126`). The H2-only fixture is at `ADR.md:57`, plan `:61` and `ci:106`. |
| ADR012-R7-04 | **Partly** | The list of provider-proven causes, and assessing or gapping manual cancels, skip instructions and forks awaiting approval, are in `ADR.md:17`, REQ-27/78, `ci:101` and `cli:126`. The fixture is at `ADR.md:57`. But a manual cancel followed by a later push meets the "later run of the same … concurrency group" test: ADR012-R8-05. The challenger rated this Closed; the experiment-peer rated it Partly. The lead adjudicates Partly on the peer's counterexample. No skip document is cited, so R-070 needs no addition. |
| ADR012-R7-05 | **Fixed** | Repository ID, resolved `as_of`, selector version, path-filter population and a retained, format-versioned canonical encoding are all present (`ADR.md:44`, REQ-78, `cli:130`). `as_of` is in REQ-78 `evidence` and `policy-file.md:23`. The distinct-cutoff fixture is at `ADR.md:61`. Where the refusal status belongs is ADR012-R8-01. |
| ADR012-R7-06 | **Fixed** | The numeric tuple (provider repository ID, run ID, run attempt, job ID) is compared field by field, with refusal and no text fallback (`ADR.md:21`, REQ-78, `cli:126`, `policy-file:78`). The 999/1000 fixture is at `ADR.md:61`. The repository filter applies before cohorts are formed, so no cross-repository collision exists (challenger). |
| ADR012-R7-07 | **Fixed** | The Scout/tool/catalogue/policy tuple is in `ADR.md:44`, REQ-78 and `cli:128,130`. The fixture is at `ADR.md:62`. |
| ADR012-R7-08 | **Fixed** | All seven items are present: the REQ-27 normative reference (`:612`, end of the row); the `ci:101` in-scope/started_at reference; `policy-file:82` with C and both conditions; C defined in `ADR.md:19` and REQ-78; plan `:90`; the pin-mixture fixture in plan `:61` and `ci:106`; and the REQ-51 caveat (`:618`). |
| ADR012-R7-09 | **Fixed** | The callout (`:449`) states inspected comparable observations, zero unwarranted and zero undecided would-blocks for the exact version across all intervals including provisional readiness, and an unknown share of at most 10 % in the window. It links `#req-22`/`#req-78` for counting and conservative certification. |

## Re-confirmation of earlier closures (R1–R6)

| Closure | Holds? | Evidence |
|---|---|---|
| R1-01, R3-06, R4-05: version-scoped event identity, many-to-many links, union across attempts | Yes | `ADR.md:20,22,34`; REQ-27. Late attempts now keep their links (`ADR.md:19`) |
| R1-03/06/07, R2-13/14: exact cohorts, observation dedup, inspected-only exposure, distinctPRs | Yes | `ADR.md:20–21`; REQ-22/78; `policy-file.md:78–80` |
| R1-04: membership keyed on `started_at`; elapsed-days formula | Yes | `ADR.md:19,21` |
| R1-05: a collection boundary can be recovered | **Restored by R7-01, apart from non-terminal attempts** | Discharge clause in `ADR.md:19`. Attempts that never become terminal: ADR012-R8-07 |
| R2-05, R6-01/08: population, supersession, exclusions | Yes, apart from manual cancel then push | `ADR.md:16–17`. ADR012-R8-05 |
| R1-08/09/11/14, R2-10, R3-01, R4-07/10: CI/local/legacy separation, migration, v2 cutover, query-only v13 | Yes | `ADR.md:34–38,44`; REQ-10 `:610`; REQ-48 `:615`; `ci:104` |
| R3-02/07, R4-02/08/11: protected-source authorization, current maintainer authority, replay idempotence, writer isolation | Yes within one lineage | `ADR.md:26,40`; REQ-26 `:611`, REQ-51 `:618`, REQ-161 `:707`; `ci:102–103`. Imports from another lineage: ADR012-R8-02 |
| R3-03, R4-06: guardian comparison and empty override inputs | Yes | `ADR.md:17–18`; REQ-27; `ci:101` |
| R2-11, R3-05, R4-03/04, R5-03/05/07, R6-07: interval selection, bounds, triage cutoff, frontier-pinned minima, dossier invariance | Yes, apart from the default-version path | `ADR.md:19–20,28,44`; REQ-78; `cli:126–132`. Late visibility in version and representative selection: ADR012-R8-03 |
| R5-06, R6-03: quiet-repository liveness and skewed-clock triage | **Depends on how R8-01 is read** | Holds if only triage and late records trigger the refusal. Does not hold on the broad reading: ADR012-R8-01 |
| R5-01, R6-10: wiring-ADR gate | Yes | `ADR.md:12,56`; REQ-51 `:618`; REQ-161; `ci:99`; plan `:61,90`; STATUS `:9,121`; `cli:134`; `policy-file:84` |
| R5-04, R6-04: zero-not_warranted and zero-Undecided across gaps, including provisional | Yes, apart from verdicts resurrected by import | `ADR.md:20,30,59`; REQ-22 `:589`. ADR012-R8-02 |
| R5-08, R6-05: Days, representative origin and declared counting changes | Yes | `ADR.md:21,48,62`; card `:1010` |
| R5-09: scanner-output failure is comparison-none | Yes | `ADR.md:17,62`; REQ-27; `ci:101` |
| R5-10, R7-07: exact cohort, default version, pin mixtures, exits | Yes | `ADR.md:44,62`; plan `:61`; `ci:106` |
| R5-11/12: declared change set; restored round records | Yes | `ADR.md:4,48`; card `:1010`. Historical rounds were not edited, per Main's digest attestation |
| Historical status lines and index | Yes | ADR-000/003/007/008 have unchanged digests and match README `:18,21,25,26`. The challenger's raw-text comparison found only status line 3 differs from base |

## Are the invariants testable and enough for REQ-22, REQ-27 and REQ-78?

| Requirement need | Testable and sufficient? | Where / gap |
|---|---|---|
| REQ-22: counting unit, deduplication, exact cohort, unknown ≤ 10 % | Yes | `ADR.md:20–22`. Late-point unknown share: ADR012-R8-04 |
| REQ-22: 14 days, inspected origin, `as_of` endpoint | Yes | `ADR.md:21,62` |
| REQ-22: zero `not_warranted` and zero Undecided across gaps, including late evidence | **Partly** | An imported older verdict can supersede a later one (ADR012-R8-02). The fixtures can pass without testing the veto (ADR012-R8-06) |
| REQ-22: maintainer-only, append-only, idempotent triage | **Partly** | Which authorization an imported row needs (ADR012-R8-02) |
| REQ-27: population independent of content, author and approval | **Partly** | Manual cancel then push (ADR012-R8-05) |
| REQ-78: complete-interval accounting and recoverable frontier | **Partly** | Non-terminal attempts (ADR012-R8-07); point-gap semantics (ADR012-R8-04) |
| REQ-78: historical reproduction and `dossier_digest` | **Partly** | Late visibility in default-version and representative selection (ADR012-R8-03); where the refusal status goes (ADR012-R8-01) |
| REQ-78: current certification liveness | **Partly** | Undefined refusal scope (ADR012-R8-01) |
| Governance of deferred wiring | Yes | `ADR.md:12,56`; REQ-51/161 |

## Issues

### ADR012-R8-01: The current-window refusal has no defined scope or status, and refusals have no exit codes (major)

- **Location:**
  - `ADR.md:44`: "current_window refuses certification (unmeasured/exit2) while any trusted readiness input is pending or stamped after C; explicit historical reproduction remains frozen and cannot certify current readiness"
  - `ADR.md:19`: "Publishing CAS assigns recorded_at=max(publishing-run provider start, parent C+1ms, …)"; "Advance C whenever these conditions permit at triage publication and periodic trusted reconcile"; "No pending input or ambiguous evidence may certify current readiness"; "Unproven completeness/time means no certification"; cross-lineage "refuse readiness"
  - `ADR.md:21`: "unavailable/unverifiable ordering refuses readiness"
  - `ADR.md:28`: "Publication after C cannot change the canonical dossier at/before C"
  - Mirrors: REQ-78 `:600`; `cli-commands.md:130,132`; `policy-file.md:23,82`
- **Problem:**
  1. **Undefined scope.** "Any trusted readiness input" could mean only cohort triage and late records. It could equally cover every published fact, frontier record or publication.
     - **Counterexample** [INFERENCE: constructed from the text, not executed]. The consistency lag is 10 minutes, and an hourly reconcile lists its runs 20 seconds after it starts:
       1. Reconcile R1 starts at 10:00:00 and seals C₁ = 09:50:20. Anything it stamps is stamped 10:00:00, which is after C₁.
       2. Reconcile R2 starts at 11:00:00 and seals C₂ = 10:50:20. Its own stamps are again after C₂.
     - On the broad reading, the current window **never certifies**, even in a quiet repository. In an active repository, normal attempts with `started_at` > C are published continually.
     - On the narrow reading, the refusal is temporary and fails closed. A triage that clears the last Undecided event at 10:30 stays refused until R2 lands.
     - The two readings give different exit codes for the same state, and one of them makes a compliant repository's gate unreachable (experiment-peer).
  2. **Unstated liveness consequence.** Every publication is stamped after the parent C. A ready result after triage is therefore reachable only at a later C at or beyond that stamp. Nothing says so.
  3. **Where the status goes.** `ADR.md:28` forbids a later publication from changing the dossier at or before C, so the refusal cannot be part of the digest. Yet nothing else records it. `--as-of <C>` then returns exit 0 with the same `dossier_digest`, and REQ-78 activation evidence `{as_of, dossier_digest}` cannot tell a refused current-window run from that bypass. The "deliberately stricter" refusal (`round-7-response.md`, consolidation 2) becomes advisory.
  4. **No exit codes for refusals.** `ADR.md:44` maps unmeasured to 2 and unsupported schema to 3. Several refusals are mapped to neither: "no certification", "refuses readiness" for ordering and cross-lineage conflicts, and "refuses reproduction" for an unsupported encoding.
- **Required fix (recommended option):** state the following in `ADR.md:44` and `:19`, and mirror it in REQ-78, `cli:130,132` and `policy-file:23,82`.
  1. **Inputs that trigger the refusal:**
     - late records of the repository that are stamped after C;
     - authorized triage or history rows stamped after C for events in the queried exact cohort's veto set;
     - pending local late facts and pending imported history in the reader's store.

     Normal facts with `started_at` > C, frontier records and publication metadata do not trigger it.
  2. **Effect.** A refused current-window run is unmeasured with exit 2, whatever result was computed. This keeps the current wording and states it explicitly.
  3. **Liveness.** A stamped input stops causing refusal once C is at or beyond its stamp. The publication that stamps it reaches that C only if its attested bound minus lag is at or beyond the stamp; otherwise a later reconcile does. Reconcile cadence must let a quiet repository's C pass every stamp.
  4. **Status.** Add a certification status, `current | refused | historical`, to the labelled live envelope, outside the digest. The dossier's reason stays the result computed at `as_of`. REQ-78 and `policy-file.md:23`: activation evidence must come from a current-window result whose status is `current`.
  5. **Exit codes:**
     - evidence refusals (unproven completeness or time, unverifiable attempt ordering, a cross-lineage verdict conflict, pending or post-C inputs) are unmeasured, exit 2;
     - an unsupported canonical encoding format is exit 3, like an unsupported schema.
  6. **Fixtures** in `ADR.md:56–62`, plan B4-S and `ci:106`:
     - the triage that clears the last Undecided event exits 2 until C reaches its stamp, then exits 0 ready;
     - a quiet reconcile whose only record is the frontier gives a computed result;
     - triage of another check after C does not refuse this cohort;
     - a late record stamped after C, while an in-flight attempt holds C back, exits 2 with status `refused`, and `--as-of C` gives the same digest with status `historical`.

### ADR012-R8-02: Re-stamped imported history can bring back a superseded verdict, and the import rules conflict (major)

- **Location:**
  - `ADR.md:19`: "Imported triage/history newly entering a sealed ledger also receives a fresh CAS stamp … deduplicate by action ID and refuse/log unverifiable original authorization. This deliberately applies to every new imported history row … If cross-lineage verdict order cannot be proved, refuse readiness"
  - `ADR.md:26`: the author must be "still … authorized … when publishing"
  - `ADR.md:28`: "A publication holds at most one triage row per event. Latest authorized row by recorded_at then action ID is current"; a superseded `not_warranted` row is only *reported*
  - `ADR.md:36`; REQ-48 `:615`; REQ-78; `cli:130`; `policy-file:82`
- **Problem:**
  1. **False-ready path.** The rule covers only the case where order *cannot* be proved.
     - **Counterexample** [INFERENCE: constructed from the text, not executed; found independently by the lead and both peers]:
       1. In lineage A, maintainer M1 marks event E `warranted` on day 25.
       2. In destination B, M2 marks E `not_warranted` on day 32.
       3. A is imported into B on day 35. The original stamps prove that A's row is older, so no refusal applies.
       4. The fresh stamp is day 35, so the latest-by-`recorded_at` rule makes `warranted` current. Zero-not_warranted passes once C reaches day 35.
     - The R6-03 rule that "replay cannot overtake a newer verdict" is bypassed through a different action ID.
  2. **Contradiction.** Importing two or more new rows for one event under one CAS breaks "at most one triage row per event per publication". Round 7's total-order check rested on that rule (experiment-peer).
  3. **Authorization is ambiguous.** `:19` checks only the original authorization. `:26` requires authorization at the current tip when publishing. A revoked author's imported row is accepted under one reading and rejected under the other, which changes the superseded-history counts and the digest (experiment-peer).
  4. **Same-lineage restore** (lead). "Every new imported history row" also covers restoring a later state of the *same* trusted lineage into a persistent sealed store (REQ-3 local runs). Rows the trusted CAS already stamped would be re-stamped. A local store has no trusted CAS, so they stay pending. That store then refuses current readiness and computes historical dossiers that differ from a fresh restore of the same ledger.
- **Required fix (recommended option):** state the following in `ADR.md:19,28,36` and mirror it in REQ-48/78, `cli:130` and `policy-file:82`.
  1. **Same-lineage fast-forward.** An import is a fast-forward when the destination's frontier history, and its rows stamped at or before its C, are an identical prefix of the import. A fast-forward adopts rows at their original trusted stamps, so the result equals a fresh restore. Fresh stamping applies only to rows absent from the destination's lineage.
  2. **Cross-lineage rows:**
     - deduplicate them by action ID;
     - publish an event's rows in successive publications, ordered by original (`recorded_at`, numeric action ID), keeping one row per event per publication.
  3. **Conflicts.** After an import, an event whose authorized history holds rows from more than one lineage is **Undecided**, never `warranted` or `not_warranted`. It stays Undecided until a new authorized maintainer verdict is published in the destination after the import. This replaces "refuse readiness if order cannot be proved".
  4. **Authorization.** An imported row needs verifiable original authorization, and its original author must also pass the `:26` current-tip check at the import publication. Otherwise the row is refused and logged, and the event is Undecided until a new verdict.
  5. **Fixtures:**
     - the day-25/day-32 case in both verdict orders stays not ready until a new verdict;
     - a multi-row import for one event keeps one row per event per publication;
     - a repeated import is a no-op;
     - a fast-forward restore into a sealed local store gives the same dossiers as a fresh restore.
- **Acceptable alternative (challenger):** separate publication visibility from supersession. A freshly stamped row could not supersede a verdict proved to be later. Not recommended: it needs ordering evidence across lineages that the design has not shown can be authenticated, while the Undecided rule reuses the reviewed zero-Undecided veto.

### ADR012-R8-03: "Late" is undefined, and late visibility is stated only for the vetoes (minor)

- **Location:**
  - `ADR.md:19`: "late facts" is not defined, and "earlier dossiers are unchanged"
  - `ADR.md:20`: "late facts additionally require publication stamp≤as_of", stated only in the veto sentence
  - `ADR.md:21`: representative selection
  - `ADR.md:44`: "Default version is check pin of maximum (started_at, numeric attempt-ID tuple) in-scope attempt≤as_of"
  - REQ-78; `cli:126`
- **Problem:**
  - The round-6 definition "at/before the destination's sealed C" was removed.
  - The explicit default-version formula, the in-scope definition and the representative rule use `started_at` ≤ `as_of` alone.
  - **Counterexample** [INFERENCE] (challenger):
    1. At C = day 30, the latest known attempt started on day 10 with pin v1, so `--as-of day30` without `--version` selects v1.
    2. A late attempt started on day 20 with pin v2 is published on day 31.
    3. The same historical command now selects v2, contradicting "earlier dossiers are unchanged".
  - A late rerun of an existing observation would similarly replace its representative at historical cutoffs (experiment-peer).
  - The `ADR.md:58` day-20 fixture would not catch either leak.
- **Severity:** the challenger rated this major. The lead adjudicates **minor**. A late record stamped after C makes current-window refuse (ADR012-R8-01), and historical results cannot certify current readiness, so no false ready can result. The harm is to reproduction, the same class as R7-05 and R7-06.
- **Required fix:**
  - **Define "late".** A fact is late when its `started_at` (attempt) or `event_at` (event) is at or before the destination's C at publication.
  - **One visibility predicate for every query:** default version, in-scope attempts, representative, links, vetoes, gaps and Days origin.
    - A normal fact is visible when `started_at` ≤ `as_of`.
    - A late fact additionally needs its stamp ≤ `as_of`.
    - A pending fact is never visible.
  - Mirror this in REQ-78 and `cli:126`.
  - **Fixture.** Run the same `--as-of day30` command without `--version` three times: before a pending import, after it, and after publication of a later-started late attempt with a different pin. That attempt is a newer attempt of an existing observation with a different inspection status. The cohort and digest stay identical.

### ADR012-R8-04: A late point gap can be read two ways (minor)

- **Location:** `ADR.md:19` ("Effective point gaps … those points exclude interval exposure"); `ADR.md:20` ("late points are gaps, never exposure"); REQ-78.
- **Problem:**
  - A point gap either (a) ends the containing interval, or (b) only removes the late attempt from exposure while the interval continues.
  - Under (b), a late comparison-none rerun is left out of TotalRuns and UnknownAttribution, and an earlier inspected attempt stays representative. The unknown share falls and PRRuns rises. That is a mild inflation path [INFERENCE] (lead).
  - Also undefined:
    - which interval an attempt whose `started_at` falls on a gap point belongs to;
    - whether a late attempt accounting several events creates a point gap at each `event_at` (experiment-peer).
  - Round 7's R7-02 counterexample assumed reading (a).
- **Required fix:**
  - A point gap ends the interval. The previous interval ends strictly before the point and the next starts strictly after it. Attempts at the point belong to no interval.
  - Every accounted `event_at` gets its own point gap.
  - Mirror this in REQ-78.
  - **Fixture:** with a late point on day 20, the interval containing `as_of` = day 35 starts after day 20, and so does the Days origin.

### ADR012-R8-05: A manual cancel followed by a later push still meets the supersession test (minor)

- **Location:** `ADR.md:17` ("pre-assessment cancellation verifiable as a later run of the same pinned lane/concurrency group for the same PR … manual cancellations … are assessed or gapped"); REQ-27 `:612`; REQ-78; `ci:101`; `cli:126`. Code: `scout-guardian.yml:41–43` (`cancel-in-progress`, cited in round 7).
- **Problem:**
  - An author manually cancels H1 before it is assessed, then pushes H2 later.
  - A later run in the same group exists, so the cancel meets the exclusion test and the manual-cancel rule at the same time.
  - The author can therefore drop a head from the population: the R7-04 harm (experiment-peer).
- **Required fix:**
  - Supersession must be causal. Either the provider records the concurrency cause, or the superseding run was created before the cancellation and no actor-initiated cancel is recorded.
  - Otherwise the cancel is manual, and the event is assessed or gapped.
  - Mirror this in REQ-27/78, `ci:101` and `cli:126`.
  - **Fixture:** a manual cancel, then a later push, is assessed or gapped.

### ADR012-R8-06: The veto fixtures can pass through refusal alone (minor)

- **Location:** `ADR.md:58` ("Late would_block remains Undecided/not-ready…", "a late not_warranted import blocks after its fresh stamp"); `ADR.md:60` ("default readiness result reflects the correction"); plan `:61`; `ci:106`.
- **Problem:**
  - While C is below the stamp, current-window refuses with exit 2.
  - That outcome already satisfies "not-ready/blocks" without exercising the cohort veto.
  - The fixtures therefore do not expose the R7-02 failure, and the R6-03 correction fixture can pass the same way (experiment-peer).
- **Required fix:** each of these fixtures must:
  - advance C to at least the stamp, with no other input that triggers refusal;
  - assert a computed exit 0 not-ready result with the named veto reason;
  - be paired with a control (the same data without the late event, the import or the correction) that gives ready.

  Mirror this in plan B4-S and `ci:106`.

### ADR012-R8-07: Attempts that are not yet terminal have no gap path (minor)

- **Location:** `ADR.md:19` ("every attempt started≤t terminal/published or covered by a published late record"); compare the event rule in `ADR.md:17` ("gap … only when it will not be accounted … approval waiting"); REQ-78; `policy-file:82`.
- **Problem:**
  - An event has a "will not be accounted" gap. An attempt has no "will not become terminal" counterpart, and "provider started_at" is not tied to job start.
  - A pending fork approval expires only after 30 days, when the run is marked failed ([GitHub Docs: Approving workflow runs from forks](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/approve-runs-from-forks)).
  - An implementation that records such a run as a started, non-terminal attempt holds C back for up to 30 days per fork. Combined with ADR012-R8-01, a steady stream of fork PRs would keep current certification refused (experiment-peer). [INFERENCE: how the API reports such runs was not verified.]
- **Required fix:**
  - An attempt starts at the provider's job start. A run awaiting approval has no started attempt; its event follows the event rule.
  - An attempt still not terminal at the collector's chosen deadline is gapped at its points, which discharges its completeness. Its later terminal facts follow the late-evidence lifecycle.
  - Mirror this in REQ-78 and `policy-file:82`.
  - **Fixture:** a fork run awaiting approval does not hold C back.
  - If the ADR text cites the GitHub page, add it to R-070.

## Checks that found no issue

- **Late lifecycle reachability.** With the discharge clause, a published late record lets C advance (`ADR.md:19`). A temporary refusal that clears at the next reconcile fails closed and is not a deadlock (challenger). On the broad reading this does not hold: ADR012-R8-01.
- **Exit rule for cutoffs.** `as_of` ≤ C, C < `as_of` ≤ reader now, and `as_of` > max(C, reader now) cover every case, including a reader clock behind C (`ADR.md:44`, `cli:132`).
- **Default-version rollback and Days origin** are determinate (`ADR.md:21,44,62`). Days can shrink after a rerun, which only makes the gate harder to pass.
- **Same-lineage triage order** stays total (`ADR.md:19,28`).
- **Mirrors.** REQ-22/27/48, `ci:101`, `cli:126–130` and `policy-file:80–82` reference REQ-78's single lifecycle instead of restating variants. The "predicates include" grammar is fixed (`ADR.md:20`).
- **Unchanged requirements and accepted ADRs.** REQ-160/162/163/168 are unchanged. The four historical ADR digests equal round 7's, and only status line 3 differs from base (challenger, raw-text comparison).
- **Ledger.** R-077 (`research-ledger.md:76`) links `round-7.md`. R-064, R-069 and R-070 are unchanged. No R-ID or ADR ID is reused.

## Considered and not raised

- **Refusing only ready results** (experiment-peer, ADR012-R8-01 item 2). Turning a computed not-ready into exit 2 fails closed, and with item 2 of the fix it is determinate. Not adopted, so the current wording is kept.
- **The `policy-file.md:52–57` example lacks `as_of`/`dossier_digest`.** Base has the same gap, and the example documents current code (experiment-peer).
- **Whether the provider can attest complete enumeration.** This is a stated implementation obligation, and refusal is the conservative outcome (`ADR.md:19`). It is recorded as a risk.
- **Two "Required fixtures" sentences at `ci:106`.** Redundant, not contradictory.
- **Byte identity of the signed set.** The council cannot hash, so Main's `shasum` verification is relied on.

## Peers and where they changed the outcome

- **`innovation-challenger`** raised 2 major findings. It rated R7-01 and R7-03..09 closed and R7-02 partly closed.
  - Imported fresh stamps bringing back a superseded verdict became ADR012-R8-02. The lead and the experiment-peer found it independently.
  - Default-version selection bypassing late visibility was new to the lead. It was merged into ADR012-R8-03 and downgraded to minor, because current-window refusal stops a false ready.
  - The challenger found no lifecycle deadlock. The lead accepts that conclusion on the narrow reading only, and rejects it as a closure because the refusal scope is undefined (ADR012-R8-01).
- **`experiment-peer`** raised 2 major and 7 minor findings. It rated R7-04 partly closed and R7-02 closed as prescribed. Adopted:
  - the broad-reading timeline for the refusal, which became ADR012-R8-01 and **changed the lead's view**: the lead had initially read the refusal narrowly, as an acceptable fail-closed delay. The peer's refusal-status and refusal-exit findings were folded into it;
  - the multi-row import contradiction and the current-tip authorization ambiguity, folded into ADR012-R8-02. Both were new to the lead, and the peer's "Undecided until a new verdict" fix was adopted;
  - the undefined term "late", merged into ADR012-R8-03;
  - point-gap semantics, which became ADR012-R8-04. **This changed the lead's view**: the lead had taken "gaps bound intervals" as decisive, then found the puncture reading's unknown-share inflation;
  - manual cancel then push, which became ADR012-R8-05 and changed the R7-04 rating to Partly;
  - fixture strength, which became ADR012-R8-06;
  - non-terminal attempts, which became ADR012-R8-07. The lead re-checked the GitHub 30-day expiry.

  Not adopted: limiting the refusal to ready results (see above).
- **Lead-only:**
  - same-lineage fast-forward re-stamping (ADR012-R8-02 item 4);
  - the unknown-share inflation argument under the puncture reading (ADR012-R8-04);
  - the severity call on ADR012-R8-03.

## Verdict

**ADR-012: 7 issues open (2 major, 5 minor). Not signed off.** Open: ADR012-R8-01, -02, -03, -04, -05, -06 and -07.

Round-7 closure:
- R7-01, -03, -05, -06, -07, -08 and -09 are fixed.
- R7-02 is fixed as prescribed. Its remaining gaps are carried as ADR012-R8-02 and ADR012-R8-06.
- R7-04 is partly fixed. The remainder is carried as ADR012-R8-05.

Earlier closures hold, with these qualifications:
- R1-05 is restored apart from ADR012-R8-07.
- R5-06 and R6-03 depend on how ADR012-R8-01 is read.
- R6-07 holds apart from ADR012-R8-03.

No digests are signed in this round. The 13 digests above identify the bytes reviewed.

## Council memo

1. **Recommendation.** Do not sign off. Apply the 7 fixes using the recommended options, then submit round 9 with fresh digests from Main for all 13 signed files and the response.
   - The R7 consolidation worked: completeness discharge, event binding, dossier identity, attempt order and the mirrors are now sound.
   - What remains sits in two places:
     - how the new fail-closed refusal is scoped and recorded (R8-01);
     - how imported history is ordered (R8-02).
   - The rest tightens definitions: visibility, point gaps, supersession cause, attempt start, and fixtures that actually exercise the vetoes.
   - State the visibility predicate (R8-03) and the refusal scope (R8-01) once, in `ADR.md:19/44` and REQ-78. Have the mirrors reference them.

2. **Alternatives.**
   - **(A) Scoped refusal with a live status, Undecided on cross-lineage conflicts, fast-forward restores (recommended).** Fails closed, determinate, reachable, and reuses the reviewed zero-Undecided veto.
   - **(B) Separate visibility from supersession by authenticated cross-lineage order (challenger).** Keeps more verdicts available after imports, but needs ordering evidence that has not been demonstrated. Acceptable, not recommended.
   - **(C) Drop the current-window refusal and rely on disclosure.** Simpler, but it reintroduces the round-7 case where a stuck frontier leaves an old ready result usable. Rejected.
   - **(D) Forbid cross-lineage import into sealed ledgers.** Simplest, but it removes the REQ-48 merge capability that ADR-007/008 rely on. Rejected.
   - **(E) Sign off and carry the items as implementation obligations.** This breaks design rule 1, and R8-01/02 change what counts as activation evidence. Rejected.

3. **Evidence.**
   - `ADR.md:17,19,20,21,26,28,36,44,56–62`.
   - `scout-design.html:449,589,600,610–612,615,618,704,707,1009–1011`.
   - `cli-commands.md:126–132`; `policy-file.md:23,78–84`; `ci-integration.md:99–106`.
   - Plan `:61,90`; STATUS `:9,121`; README `:18–30`; `research-ledger.md:76`.
   - `round-7.md` (issue sections, and "Checks that found no issue" on triage order); `round-7-response.md` (consolidations 1–4).
   - `scout-guardian.yml:41–43`, cited in round 7.
   - [GitHub Docs: Approving workflow runs from forks](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/approve-runs-from-forks).

4. **Risks.**
   - **Refusal windows** (R8-01) delay current certification until the reconcile after the last triage. This is intended and now stated. Reconcile cadence becomes a liveness parameter of the implementation.
   - **Undecided after imports** (R8-02) can hold a cohort until a maintainer re-triages. This is conservative and visible.
   - **Point gaps end intervals** (R8-04), so late evidence resets Days. This deflates readiness and never inflates it.
   - **Supersession and job start** (R8-05, R8-07) need provider cause and job-start data [INFERENCE: availability not verified]. Without them the event or attempt is gapped, which is conservative.
   - **Provider completeness.** The attestation remains unproven, and refusal is the fallback.
   - **Reversibility.** Every change is document-level and nothing is adopted. A superseding ADR can replace this one.

5. **Execution outline.**
   1. **R8-01.** Edit `ADR.md:44,19` and mirror the edits in REQ-78, `cli:130,132` and `policy-file:23,82`:
      - define the inputs that trigger refusal;
      - keep refusal as unmeasured exit 2;
      - state the liveness rule;
      - add `certification: current|refused|historical` to the live envelope, and require activation evidence from status `current`;
      - assign exits: evidence refusals 2, unsupported encoding 3.
   2. **R8-02.** Edit `ADR.md:19,28,36`, REQ-48/78, `cli:130` and `policy-file:82`:
      - same-lineage fast-forward adopts original stamps;
      - cross-lineage rows are deduplicated, published one row per event per publication in original order, and need both original and current-tip authorization;
      - an event with conflicting lineages is Undecided until a new verdict.
   3. **R8-03.** Define "late", and state one visibility predicate for every query in `ADR.md:19–21,44`, REQ-78 and `cli:126`.
   4. **R8-04.** Point gaps end intervals, with exclusive boundaries and one gap per accounted `event_at` (`ADR.md:19–20`, REQ-78).
   5. **R8-05.** Supersession must be causal (`ADR.md:17`, REQ-27/78, `ci:101`, `cli:126`).
   6. **R8-07.** Attempts start at job start; an attempt still not terminal at the deadline is gapped and later handled as late (`ADR.md:19`, REQ-78, `policy-file:82`). Add the GitHub page to R-070 only if the text cites it.
   7. **R8-06 and new fixtures.** Add them to `ADR.md:56–62`, plan B4-S `:61` and `ci:106`:
      - veto fixtures with C past the stamp, a computed not-ready result and a ready control;
      - the refusal fixtures from R8-01 item 6;
      - the import fixtures from R8-02 item 5 (both orders, multi-row, repeat, fast-forward);
      - the historical default-version fixture;
      - the day-20 point-gap interval fixture;
      - manual cancel then push;
      - a fork run awaiting approval.
   8. Extend R-077's list of rounds with `round-8.md`.
   9. Write `round-8-response.md` mapping each issue. Main supplies the SHA-256 of all 13 signed files and of the response, then round 9 follows.
