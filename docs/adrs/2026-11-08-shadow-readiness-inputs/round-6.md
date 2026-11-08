# Round 6: ADR-012 (shadow-window readiness inputs)

Date: 2026-11-08. This round follows design rules 1–3 and the ledger rules in `CLAUDE.md`, and the ADR rules in `docs/adrs/README.md`. ADR-012 is an agent decision under the owner's standing delegation (R-064). Where a fix involves a choice, the council names the option it recommends under that delegation.

The round reviews the bytes revised after round 5 (`round-5-response.md`). It has three parts:

1. Check every ADR012-R5-01..12 fix against its required fix, including the R5-12 record restoration.
2. Test whether each invariant can be tested and whether together they are enough for REQ-22, REQ-27 and REQ-78.
3. Look for new issues.

The council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review. It checked R5 closure at every named location and tried to build an implementation that meets the literal text but breaks REQ-22/27/78. It also probed interactions between the R5 fixes and checked the declared change set against the diff with base.
- **`experiment-peer` (Opus):** methodological review. It treated readiness as a measurement and looked for confounds, inflation, deflation, gaming, results that cannot be reproduced, and terms that the round-5 fixes left undefined. It also checked closure and fixture coverage.

The review was read-only. Nobody edited files, executed product code, ran model calls outside the council, or touched `.codegraph`. Codegraph was not used, because these are documents and the council had no shell. Searches used `read` and `grep`.

**Main computed and verified the digests below with `shasum`. The council did not: the lead and both peers had read-only tools and no shell.** This round is not a sign-off, so the digests only identify the bytes reviewed. All paths are under `docs/`.

| File | SHA-256 (Main) |
|---|---|
| `adrs/2026-11-08-shadow-readiness-inputs/ADR.md` | `198fe542207d88b705ebe9c18c7b2036f8d877288f0e2205f7d83760bd435432` |
| `scout-design.html` | `d530d3081410c1f5a7503a2ea7bf018d02876502c7db7cd43aa56ba435446120` |
| `adrs/README.md` | `cde415ff10b659a073d010f8a48c281d6f75a5a1cba466962d589e937b94587f` |
| `STATUS.md` | `7200ced2009d535e8f8d4dbb03c491f4e3fa05b87082813bdf585484de4785ee` |
| `plans/2026-11-06-remaining-work.md` | `005cdc2af563a0a110fdaa6c2a53b495135400e3a7b1b06a66e3ee651074ad86` |
| `references/policy-file.md` | `6e26c0c5ce03e5b2c431e0fbf569c2d6a0e51c9c03bdb0c9b940ec915b5af824` |
| `references/ci-integration.md` | `4bd5b83547fe92da26403d38f4bbf9b0174b291a27827a60b326b4ce3f1a70db` |
| `references/cli-commands.md` | `3f660b2ae1fc074a6a1f26773c23fd19a38b50974b4a83787e4b1cb65a496e4b` |
| `research-ledger.md` | `97243481f03f2a40c74cf965fb917fda56540c2c22fd06325210ab68faa3c3e8` |
| `adrs/2026-10-19-v1-decisions/ADR.md` (ADR-000) | `1cd521fd96439d87275ba353f7b05451c3e0c4d556ee665529aab85970ed81cb` |
| `adrs/2026-10-23-decisions/ADR.md` (ADR-003) | `973d24c4cfdfdd98acb4070a7bd72ca0c0e1ddf303bd4ccec8f92fe1cf117cc5` |
| `adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) | `ee593b6241814c24035671d666566bc2b7c399135867226c26e9626f3936abc8` |
| `adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) | `80e3375f81656a8856cf662cbec99a36d8e00bc451ff3c3592925d2293f44d95` |

The following records are not signed. Main computed their SHA-256:

| File | SHA-256 (Main) |
|---|---|
| `round-5-response.md` | `8c9c8e3171dc61f686428144dc3b0fc3c483419b5f64386b23d771d7acbb9d0d` |
| `round-4.md` (restored) | `0b132628f177fe1c0c648891bac2f00f9641300e0969f6069034c3721a5b464c` |
| `round-5.md` (restored) | `874bfa21d2e0f6ea83ac4b7fb421d1bab05b9bc478336ffee9ec5bab420f9f25` |

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11`, branch `agent/policy-shadow-adrs`, based on origin/main `7bb73e52`. Every change is staged and uncommitted.
- Base bytes come from `/Users/adamjackson/Projects/work/scout-adr11b`.

**Files read.** Line numbers below refer to the raw worktree files.
- ADR-012 `ADR.md:1–68`, all lines:
  - `:3–6` status, revision and sources; `:10–12` decision and delegated narrowing;
  - `:16–22` items 1–7 on which observations count;
  - `:26` triage authorization; `:28` triage rows and cutoffs; `:30` ShadowReady;
  - `:34–40` persistence, manifest, cutover and writer boundary;
  - `:44` CLI; `:48–52` amendments, sources and route;
  - `:56–62` implementation acceptance invariants; `:66–68` folder records.
- `round-5.md` (all 392 lines) and `round-5-response.md` (all lines).
- `round-4.md:290–392`, the restored tail.
- `round-1.md:190–198`, `round-1-response.md:20–23`, `round-2.md:263–276`, `round-2-response.md:30–34`, `round-3.md:340–343` and `round-4.md:333`, located by grep for `unstarted|cancel`.
- `scout-design.html`, raw:
  - the Activation callout `:449`;
  - REQ-22 `:589`, REQ-78 `:600`, REQ-26 `:611`, REQ-27 `:612`, REQ-48 `:615`, REQ-51 `:618` and REQ-161 `:707`, plus grep context for REQ-10/39/47/160/162/163/166/168;
  - the cards for ADR-000 `:736`, ADR-003 `:807`, ADR-007 `:901`, ADR-008 `:925`, ADR-011 `:992–1003` and ADR-012 `:1007–1011`.
- `cli-commands.md:118–134`, `policy-file.md:74–84`, `ci-integration.md:95–106`.
- Plan `:14–15,35,61–62,90,92,100`; STATUS `:7,9,121,129–131`; README `:12–30`; `research-ledger.md:73–76`.
- Code:
  - `internal/policy/policy.go:384–432`: `ShadowStats` and `ShadowReady`, including the order of the provisional return and the Undecided check;
  - `.github/workflows/*.yml`: trigger blocks only.
- The challenger compared the four historical ADR bodies and the REQ-160/162/163/168 rows with base.

## Issue summary

**10 distinct issues are open: 4 major and 6 minor.**

| ID | Severity | Summary |
|---|---|---|
| ADR012-R6-01 | major | Eligible events without a started attempt escape completeness. A collector can choose "unstarted" or "cancelled", events have no time, and the provider attestation from rounds 1–2 was lost |
| ADR012-R6-02 | major | The sealed frontier has no upper bound, and late eligible facts that are refused leave no trace. Merge and recompute can also alter sealed state |
| ADR012-R6-03 | major | The writer-clock term in `recorded_at`, made sticky by the R5-05 term, can hide later triage corrections from `current_window` indefinitely while ready is true |
| ADR012-R6-04 | major | A collection gap still clears undecided events, and the rare-security provisional path returns before the Undecided check |
| ADR012-R6-05 | minor | Comparability, inspection status, PRRuns membership and the Days origin are not tied to the representative attempt |
| ADR012-R6-06 | minor | Ordering ties: frontier records at the same C, several triage rows for one event in one publication, and default-version selection at equal `started_at` |
| ADR012-R6-07 | minor | The reproducible dossier and its `dossier_digest` are mixed with live diagnostics, and reader-clock exits can reject the default query |
| ADR012-R6-08 | minor | The round-5 fixes left "would run", "selector", "provider enumeration" and "in-scope attempt" undefined, and guardian path filters contradict "never depends on PR content" |
| ADR012-R6-09 | minor | The cross-interval veto changes REQ-22's "during the window" clause but is not declared, and `ADR.md:30` calls it retained "verbatim" |
| ADR012-R6-10 | minor | The ADR-012 card and STATUS `:9` still say wiring "belongs to"/is "left to" the implementation PR without the R5-01 wiring-ADR caveat |

## Closure of round-5 issues

"Fixed" means the required document fix is present in the signed bytes at every named location. It does not mean the implementation is validated. "Fixed as prescribed" means the council's own prescription was applied but left a gap, carried as a new R6 issue. The lead and both peers agree on presence at every row. Where the peers split between Fixed and Partly, the lead adjudicated as shown.

| Issue | Status | Evidence / remainder |
|---|---|---|
| ADR012-R5-01 | **Fixed** | The wiring-ADR sentence is in `ADR.md:12,56`, REQ-161 `:707`, `ci-integration.md:99`, plan B4-S `:61` (with the dependency "and, where wiring changes requirements, an adopted wiring ADR"), STATUS `:121`, `cli-commands.md:134` and `policy-file.md:84`. Plan `:90` carries it too. Summary leftovers in the card and STATUS `:9`: ADR012-R6-10. |
| ADR012-R5-02 | **Fixed as prescribed** | The population, forks and bots, the selector ban, `non_default_base` and enumeration are in `ADR.md:16`, REQ-27 `:612`, REQ-78 `:600` and `ci-integration.md:101`, with fixtures at `ADR.md:57` and B4-S. The "unstarted" exclusion carried over from the removed topology, and the undefined "its period", let an eligible event escape completeness: ADR012-R6-01. Challenger: Fixed. Peer: Partly. |
| ADR012-R5-03 | **Fixed as prescribed** | The provider clock, terminal-complete advancement, preservation of pre-C facts and refusal of late facts are in `ADR.md:19,58`, REQ-78 and `cli-commands.md:126`. No upper bound on t, and silent refusal: ADR012-R6-02. Events without a start: ADR012-R6-01. |
| ADR012-R5-04 | **Fixed** | The cohort-wide zero-`not_warranted` veto and the prior-interval count are in `ADR.md:20,30,59`, REQ-22 `:589`, REQ-78, `policy-file.md:80` and `cli-commands.md:128`. Its Undecided counterpart (ADR012-R6-04) and its declaration (ADR012-R6-09) are new. |
| ADR012-R5-05 | **Fixed** | The three-term formula, reassignment on retry and superseded-verdict disclosure are in `ADR.md:28,60`, REQ-78, `cli-commands.md:130` and `policy-file.md:82`. Its interaction with cutoff visibility: ADR012-R6-03. |
| ADR012-R5-06 | **Fixed** | Advancement on triage publication and periodic reconcile, quiet repositories, frontier age, rows after the cutoff and the name "triage at as_of" are in `ADR.md:28,44,61`, REQ-78 and `cli-commands.md:130`. ADR012-R6-03 shows that advancing C need not make a fast-stamped row visible. The challenger rated this Partly. The lead carries the defect as ADR012-R6-03, not as R5-06 being open. |
| ADR012-R5-07 | **Fixed** | Policy revision and minima per frontier, and selection by the latest frontier at or before `as_of`, are in `ADR.md:19,34,44,61`, REQ-78, `cli-commands.md:128,130` and `policy-file.md:80`. Frontier records at the same C value: ADR012-R6-06. |
| ADR012-R5-08 | **Fixed** | The Days formula and the declared inspected-only origin and PRRuns exclusions are in `ADR.md:21,48`, REQ-22 `:589`, REQ-78, card `:1010`, `policy-file.md:78` and `cli-commands.md:128`. Origin taken from the representative, and the missing origin and exclusion fixture: ADR012-R6-05. |
| ADR012-R5-09 | **Fixed** | Scanner output failure is comparison-none, and only trusted enumeration, attestation or publication failure is a gap. This is in `ADR.md:17,19`, REQ-27, REQ-78 and `ci-integration.md:101`, with the scanner-deletion fixture at `ADR.md:62` and B4-S. |
| ADR012-R5-10 | **Fixed** | Exact-cohort attempts and links, the default version (most recently pinned), the exit boundary and non-additive pin tuples are in `ADR.md:20,22,44,62`, REQ-78 and `cli-commands.md:126,128,132`. Tie-break and "in-scope": ADR012-R6-06 and ADR012-R6-08. |
| ADR012-R5-11 | **Fixed** | Card `:1010` reads "Selects no workflow/key/commands/transport/retention/install mechanics; REQ-162/163/168 unchanged". `ADR.md:4` and the card list the overview Activation callout. `ADR.md:48` adds "trusted readiness publication (REQ-26/51/161) for ADR-000/003/007". These agree with README `:18,21,25,26`, card relations `:736,807,901,925,1007` and the ADR-000/003/007/008 status lines, which the challenger found to differ from base only on line 3. |
| ADR012-R5-12 | **Fixed** | `round-4.md:299` now has its body (the `pull_request_target` recheck), followed by fidelity, checks, considered items, peers (`:337`), memo (`:356–388`) and verdict (`:390–392`): "ADR-012: 16 issues open (2 major, 14 minor). Not signed off." `round-5.md` ends with peers, verdict (`:336–338`) and a memo through `:392`. Both are outside the signed set, as required. Main attests that the old 299-line bytes are an exact prefix. The council cannot verify this: it has no shell and no copy of the old file. Line 299 matches the heading round 5 cited. |

## Are the invariants testable and enough for REQ-22, REQ-27 and REQ-78?

Each bullet at `ADR.md:56–62` names an observable fixture. The R5-01 bullet is a review-process gate, checkable at PR review rather than as a runtime fixture, which is acceptable. The gaps are invariants whose oracle is undefined, or that a conforming implementation can satisfy while defeating the requirement.

| Requirement need | Testable and sufficient? | Where / gap |
|---|---|---|
| REQ-22: counting unit, deduplication, exact cohort | Yes | `ADR.md:20–22`; REQ-22/78 |
| REQ-22: unknown ≤ 10 %, counted once | Yes | `ADR.md:21`; it matches `policy.go:415` (equality passes) |
| REQ-22: 14 days, inspected origin, `as_of` endpoint | **Partly** | Which attempt supplies the origin and the status (ADR012-R6-05) |
| REQ-22: zero `not_warranted` per exact version | **Partly** | The cross-gap veto holds. Fast-stamped corrections can be hidden (ADR012-R6-03), and the change is undeclared (ADR012-R6-09) |
| REQ-22: maintainers triage every persisted event | **No** | A gap clears undecided events (ADR012-R6-04) |
| REQ-22: maintainer-only, append-only, idempotent triage | Yes | `ADR.md:26`; REQ-22/161 |
| REQ-27: guardian-equivalent comparison, inputs, parity proof | Yes | `ADR.md:17–18`; REQ-27 |
| REQ-27/51: protected base, scanner failure is comparison-none not a gap | Yes | `ADR.md:16–17,19`; REQ-27/51 |
| REQ-78: complete-interval accounting | **No** | Events without attempts (ADR012-R6-01); no bound on the frontier, and silent refusals (ADR012-R6-02); undefined frame terms (ADR012-R6-08) |
| REQ-78: historical reproduction and `dossier_digest` | **Partly** | Live fields in the dossier and reader-clock exits (ADR012-R6-07); frontier and triage ties (ADR012-R6-06); merge after sealing (ADR012-R6-02) |
| REQ-78: triage liveness on quiet repositories | **Partly** | Defeated by fast writer clocks (ADR012-R6-03) |
| Governance of deferred wiring | Yes, with a summary leftover | ADR012-R6-10 |

## Issues

### ADR012-R6-01: Eligible events without a started attempt escape completeness (major)

- **Location:**
  - `ADR.md:17`: "conflicted/unstarted/cancelled assessments are not_assessed, excluded and disclosed".
  - `ADR.md:16`: "Judge completeness against provider enumeration: an unaccounted eligible event makes its period a gap".
  - `ADR.md:19`: "Advance frontier C to t only when every eligible attempt with started_at≤t is terminal (assessed, comparison-none or not_assessed)…".
  - `ADR.md:58`: "Queued/in-flight/unpublished eligible attempts prevent advancing past their start unless their period is a recorded gap."
  - REQ-27 `:612`: "conflict/unstarted/cancelled attempts are excluded/disclosed".
  - REQ-78 `:600`: "all eligible attempts started≤t"; `cli-commands.md:126`; `ci-integration.md:101`.
- **Problem:**
  - Rounds 1–2 allowed `not_assessed` only for **provider-confirmed** non-starts of a fixed producer job. See `round-1-response.md:21` ("Provider-confirmed pre-assessment cancellations/skips"), `round-1.md:195` and `round-2.md:272,276`. Under that producer, "Runs whose job has not started do not hold the cutoff" was safe, because the provider recorded the job state.
  - The round-4 scope cut removed that producer. "Unstarted" now also covers a trusted collector that simply never starts an assessment.
  - **Counterexample:** a nightly collector enumerates failing PR event E, which was merged within the day, and records it as `not_assessed`/unstarted. E is "accounted", so its period is not a gap. E has no `started_at`, so nothing holds C back. The interval certifies, and E appears only in an exclusion count.
  - The R5-02 fixture ("omit a failing eligible PR event") passes trivially, because E was not omitted. The R5-03 fixture has no oracle for queued attempts, which have no start.
  - The rule keyed on events (`:16`) and the rule keyed on attempts (`:19`) use unconnected clocks. "Its period" is undefined.
  - Approval-pending fork runs are a related case. Maintainers grant approval after seeing the PR, so treating approval-pending as a provider-attested exclusion lets a content-informed maintainer choice select the population. That contradicts `ADR.md:16` ("never depends on PR content, labels, author choice").
- **Required fix (recommended option):** state the following in `ADR.md:16,17,19,58`, REQ-27, REQ-78, `ci-integration.md:101` and `cli-commands.md:126`.
  1. Each provider-enumerated eligible event carries its provider-recorded event time `event_at`. Observation membership in an interval stays keyed on the attempt's `started_at`, the R1-04 choice.
  2. An eligible event is accounted only by:
     - a terminal attempt (assessed or comparison-none);
     - `non_default_base`;
     - `not_assessed` for a provider-attested reason from a closed list, in which the pinned guardian lane itself did not assess the event: a merge conflict with no merge result, or a provider-recorded cancellation or skip of that lane's run before assessment, including its own concurrency supersession, as settled in R2-05.
  3. Reasons on that list never depend on outcome, content, labels, author or maintainer approval.
  4. Anything attributable to the collector is a recorded gap containing that event's `event_at`. That covers not starting, deferring, latency, a PR closed or merged before the collector reached it, and waiting on an approval the collector does not need.
  5. The trusted collector assesses eligible fork events whatever the workflow-approval state. If the chosen wiring cannot, those events are gaps. Treating approval-pending as `not_assessed` is the rejected alternative.
  6. C may advance to t only when both hold:
     - every eligible event with `event_at` ≤ t is accounted or lies in a recorded gap;
     - every attempt with `started_at` ≤ t is terminal and published in that state.
  7. Replace the "queued" fixture with two fixtures:
     - an event the collector skipped as "unstarted" prevents its period from certifying;
     - an approval-pending fork event becomes an assessed attempt or a gap, never an exclusion.

### ADR012-R6-02: The sealed frontier has no upper bound, and refused late facts leave no trace (major)

- **Location:** `ADR.md:19` ("Advance frontier C to t only when…"; "refuse/log late facts in that range"); REQ-78 `:600` ("refuse/log late additions started≤C"); `cli-commands.md:126`; `ADR.md:58`; `ADR.md:36` ("Merge unions immutable facts/links/history … and recomputes coverage"); REQ-48 `:615` ("union immutable rows/conflict-refuse and recompute coverage").
- **Problem:**
  - **Future seal.** Nothing requires t to be at or before the time the certifying enumeration was taken, so the condition is vacuously true for a future t. A writer that estimates "provider now" with a fast clock, or that seals up to a future `recorded_at` to satisfy R5-06, permanently refuses every eligible attempt that later starts at or before C. Those PRs are neither counted nor gapped, and the interval certifies. The default query then has `as_of` = C later than the reader's clock and exits 3 (`ADR.md:44`).
  - **Listing lag.** Failing attempt X starts at 09:58 and is missing from a lagged provider listing when C = 10:00 is sealed. X arrives at 10:30 and is refused and logged. No later query, including `current_window`, ever reflects it. This reaches the R5-02 harm through the R5-03 rule.
  - **Merge after sealing.** An export carrying attempt X (started at or before C₁, absent from the destination) is unioned by merge or import. Merge then changes `as_of = C₁`, and "recompute coverage" can change sealed intervals or frontier records. `:19` arguably controls, but `:36` and REQ-48 do not say so.
- **Required fix (recommended option):** state the following in `ADR.md:19,36,58`, REQ-48, REQ-78 and `cli-commands.md:126`.
  1. t never exceeds the provider-attested time through which the certifying enumeration is complete. That is the provider listing or delivery time, less any provider consistency lag that the implementation must justify. It is never the writer clock.
  2. An eligible fact refused because it is at or before C is not dropped. Append an immutable late-coverage gap record stamped by the publishing-CAS `recorded_at`, which is therefore greater than C.
  3. Queries with `as_of` ≥ that `recorded_at` treat the point of the late fact as a gap for interval-scoped counts. Queries with `as_of` < that `recorded_at` are unchanged, so historical reproduction holds. Disclose the late-gap count.
  4. Merge and import obey the same rule. They never add facts at or before the destination's C, and never alter sealed intervals, C history or frontier records. A late fact becomes a late-gap record.
  5. Fixtures:
     - sealing past the enumeration time is refused;
     - a lagged-listing attempt yields a late gap that is visible after its `recorded_at`, while `as_of = C₁` is unchanged;
     - merging an export that holds a fact from before C leaves `as_of = C₁` unchanged.

  Making `ready` null for the whole affected interval (peer option) is an acceptable alternative. It is not recommended, because a late gap reuses the reviewed interval and gap mechanism.

### ADR012-R6-03: The writer-clock term in `recorded_at` can hide later corrections indefinitely (major)

- **Location:** `ADR.md:28` ("max(writer UTC now, parent collectedThrough + 1 ms, parent maximum recorded_at + 1 ms)"; "current_window uses as_of=C"; "Advance C whenever terminal completeness is certified…"); REQ-78 `:600`; `cli-commands.md:130`; `policy-file.md:82`; fixtures `ADR.md:60–61`.
- **Problem:**
  - C is on the provider clock and never passes provider time; ADR012-R6-02 makes that explicit. `recorded_at` takes the writer clock, and the R5-05 term makes any maximum sticky for every later row.
  - **Counterexample:** event E is `warranted` and visible at C. A writer whose clock is one day ahead publishes any row, stamped tomorrow. A correctly clocked maintainer later corrects E to `not_warranted`. The parent-maximum term stamps the correction after tomorrow as well.
  - Quiet reconciles advance C throughout today, but the default query keeps `warranted`, so ready can be true. A runner clock a year ahead hides every later triage row for a year. Append-only storage cannot repair this.
  - The R5-05 fixture passes, because the later row wins in unfiltered history. The R5-06 fixture passes, because C advances. The correction still never reaches `current_window`. Disclosing rows after the cutoff does not stop ready from being true.
  - Before round 5 a fast clock affected only its own row. The R5-05 term spreads it to every later row. This is the same harm class as R5-05, which is major.
- **Required fix (recommended option):**
  - Replace "writer UTC now" with the provider-recorded start time of the publishing run, on the same provider clock as C: recorded_at = max(publishing-run provider start, parent C + 1 ms, parent maximum recorded_at + 1 ms), recomputed on every CAS retry.
  - If that provider time is unavailable or unreliable, publication is refused, mirroring "unreliable start time prevents certification".
  - Mirror this in REQ-78, `cli-commands.md:130` and `policy-file.md:82`.
  - Add a fixture: a fast-clock writer publishes, a later writer publishes a correction, and after a quiet reconcile the **default readiness result** reflects the correction.
  - **Acceptable alternative:** keep the writer clock but refuse any publication whose writer clock exceeds provider time by more than a declared tolerance.
  - **Not chosen:** the challenger's durable sequence plus a fail-closed pending-correction rule. A sequence alone does not make the fast-stamped row itself visible. Once both clocks are provider time, the remaining delay is bounded by the longest in-flight attempt, which R5-06's disclosure already covers.

### ADR012-R6-04: A collection gap still clears undecided events (major)

- **Location:**
  - `ADR.md:20`: "Days/PRRuns/TotalRuns/UnknownAttribution/WouldBlock/Warranted/Undecided remain interval-scoped".
  - REQ-78 `:600` and `cli-commands.md:128`: "other window counts unchanged".
  - REQ-22 `:589`, base text kept: "maintainers triage every persisted `would_block` event".
  - `policy.go:416–424`: the rare-security provisional return comes **before** the `Undecided > 0` check.
  - The council's own R5-04 fix scoped Undecided to the interval (`round-5.md:193`).
- **Problem:**
  - Interval 1 holds 3 undecided events of the exact version. An outage creates a gap. Interval 2 is clean for 14 days with 5 observations, and readiness returns "shadow window satisfied". For a rare-security check, interval-scoped WouldBlock = 0 takes the provisional path in the same way.
  - Round 5's own rationale applies equally here: "Version bumps are the designed reset (REQ-22). A collection artifact should not be one" (`round-5.md:190`). Untriaged events may be false positives nobody examined.
  - This corrects the council's R5-04 text, which did not consider Undecided.
- **Required fix (recommended option):**
  - Zero-Undecided is evaluated over the same cohort-wide event set as zero-`not_warranted`: every persisted CI event of the repository and exact cohort linked to an assessed cohort attempt started at or before `as_of`, across all intervals, using triage at `as_of`.
  - Every ready result requires it, including the rare-security provisional result.
  - Report a prior-interval undecided count.
  - The other counts stay interval-scoped.
  - Update `ADR.md:20,30`, REQ-22, REQ-78, `policy-file.md:80` and `cli-commands.md:128`.
  - Fixtures: interval-1 undecided events, then a gap, then a clean interval 2 is not ready; the rare-security variant is not provisional-ready.
  - Disclosure alone is the rejected alternative, as it was in R5-04.

### ADR012-R6-05: The representative does not fix comparability, inspection status or the Days origin (minor)

- **Location:** `ADR.md:21` ("PRRuns counts only selected-check fully inspected comparable-base observations"; "origin is earliest started_at among in-interval represented comparable observations whose selected check is inspected"); REQ-78 `:600`; `cli-commands.md:128`; `policy-file.md:78`; card `:1010`; fixture `ADR.md:62` ("Days ends at as_of").
- **Problem:**
  - Observations have no single `started_at` or status; attempts do.
  - **Counterexample:** observation O has attempt a1 on day 0 (comparison-none) and a reassessment a2 on day 10 (comparable and inspected).
    - Taking O's earliest start gives Days = 10 the moment a2 lands, although nothing comparable was inspected before day 10.
    - Taking an inspected earlier attempt also lets PRRuns count O when its latest attempt is not inspected.
  - Two conforming implementations differ at the 14-day boundary.
  - The acceptance bullet tests only the endpoint, not the declared inspected-only origin or the PRRuns exclusions.
- **Required fix:**
  - Comparability, inspection status, PRRuns membership and the origin `started_at` all come from the observation's representative at `as_of`: the latest terminal assessed cohort attempt with interval.start ≤ `started_at` ≤ `as_of`, ordered by (`started_at`, attempt ID).
  - The origin is the earliest representative `started_at` among represented observations whose representative is comparable and inspected.
  - Mirror this in REQ-78, `cli-commands.md:128` and `policy-file.md:78`.
  - Add fixtures to `ADR.md:62` and B4-S:
    - the a1/a2 case gives origin day 10;
    - an earlier `not_inspected` or `not_applicable` observation followed by an inspected one neither backdates Days nor inflates PRRuns.

### ADR012-R6-06: Ordering ties in frontier records, triage rows and the default version (minor)

- **Location:** `ADR.md:19` ("Each frontier retains the current protected-tip policy revision/effective min_days/min_pr_runs"); `ADR.md:44` ("latest published frontier at/before as_of"; "latest-started in-scope attempt"); `ADR.md:28` ("latest authorized row by recorded_at then action ID"); REQ-78; `cli-commands.md:126,130`.
- **Problem:**
  - **Frontier records.** C = T is published under 14/5, and the minima then change to 30/10. A publication that cannot advance C writes another frontier record at T under the new revision. `--as-of T` now evaluates under 30/10, which breaks the R5-07 fixture.
  - **Triage rows.** One publication carries one `recorded_at`. Two rows for the same event in that publication tie on `recorded_at`, and a text comparison of "repository ID plus dispatch run ID" sorts run 999 after run 1000.
  - **Default version.** Two attempts with equal `started_at` that pin different versions, as can happen during an upgrade or rollback, both qualify as "latest-started". Storage or import order then chooses which cohort the default command evaluates.
- **Required fix:**
  - A frontier record is created only when C strictly increases. It is immutable, keyed by its C value, and carries the policy revision at creation.
  - A publication holds at most one triage row per event, or an explicit within-publication sequence.
  - The default version is ordered by (`started_at`, attempt ID).
  - Mirror this in REQ-78 and `cli-commands.md:126,130`.
  - Fixtures: a minima change without a C advance leaves `--as-of T` unchanged; tied default-version attempts give the same cohort after export and import.

### ADR012-R6-07: Live diagnostics are mixed into the reproducible dossier, and reader-clock exits can reject the default query (minor)

- **Location:**
  - `ADR.md:44`: the report includes "frontier age, rows-after-cutoff"; "C<as_of≤reader now exits 2; as_of>reader now exits 3".
  - `ADR.md:28`: "Publication after C cannot change a dossier at/before C".
  - `ADR.md:58`: "reproduce the identical as_of=C1 facts, links and result".
  - REQ-78 `:600`: activation evidence records `dossier_digest`, and the exits are mirrored there.
  - `cli-commands.md:130,132`.
  - Round 5 relied on the digest binding the output (`round-5.md:316`).
- **Problem:**
  - "Rows after cutoff" for `as_of = C₁` grows with every later publication, and frontier age depends on the reader's clock. A digest over the full output is therefore not reproducible.
  - A reader whose clock lags the provider clock gets `as_of` = C later than its own now, so the **default** query exits 3. An explicit `--as-of` in (reader now, C] also exits 3, although it is covered.
- **Required fix:**
  - Define the canonical dossier as the fields fixed by state at or before `as_of`: cohort, interval, frontier record, pins, policy revision, counts and triage at `as_of`. `dossier_digest` covers exactly those fields.
  - Frontier age (stated as reader clock), authorized rows after `as_of`, installed-version mismatch and reader time form a labelled live envelope outside the digest.
  - Exits:
    - `as_of` ≤ C always evaluates;
    - C < `as_of` ≤ reader now exits 2;
    - `as_of` > max(C, reader now) exits 3.
  - Mirror this in REQ-78 and `cli-commands.md:130,132`.
  - Fixtures: the digest is identical before and after a later publication; the default query still evaluates when the reader's clock is behind C.

### ADR012-R6-08: Terms introduced by the round-5 fixes are undefined (minor)

- **Location:** `ADR.md:16` ("runs or would run", "Eligibility never depends on PR content", "a selector change", "provider enumeration"); `ADR.md:44` and REQ-78 ("latest-started in-scope attempt"); REQ-27 `:612`; `ci-integration.md:101`; `cli-commands.md:126`.
- **Problem:**
  - The guardian is a reusable workflow (`scout-guardian.yml:20–21`), so adopter callers may add `paths:` filters. "Would run" then depends on PR paths, which contradicts the content clause read literally.
  - Whether a Scout, tool or catalogue pin upgrade is a "selector change" is undetermined. Two conforming implementations cut intervals differently and compute different Days and PRRuns.
  - "Provider enumeration" could mean runs or events.
  - "In-scope" could include `non_default_base` or `not_assessed` attempts.
- **Required fix (recommended definitions):**
  - Eligibility is the pinned guardian PR lane's protected default-branch trigger configuration (event types, branch and path filters), with no further filter. The content clause reads "beyond that configuration". Path filters are disclosed as a limit on the population.
  - "Selector" means that configuration plus the collector's eligibility-logic version. Changing either starts a new interval. Scout, tool, catalogue and policy pin changes are not selector changes: the per-tuple pin disclosure and the cohort split already cover them.
  - "Provider enumeration" is the provider's listing of PR events, with `event_at`, on PRs based on the protected default branch. It is not a listing of runs.
  - An "in-scope attempt" is this repository's CI-scope eligible attempt on the protected default base (excluding `non_default_base`) with a recorded pin, of any terminal outcome.
  - Mirror these in REQ-27, REQ-78, `ci-integration.md:101` and `cli-commands.md:126`.

### ADR012-R6-09: The cross-interval veto is undeclared and described as "verbatim" (minor)

- **Location:** `ADR.md:30` ("Retain the 14-day/5-observation defaults, cohort-wide zero-not-warranted condition across gaps, maintainer/window/per-repository clauses … verbatim"); `ADR.md:48`; card Changes `:1010`.
- **Problem:**
  - Extending zero-`not_warranted`, and zero-Undecided under ADR012-R6-04, past the window changes REQ-22's "during the window" clause. Under the R5-08 standard, such counting changes are declared.
  - `ADR.md:48` and the card declare only the PRRuns and Days changes. `:30` presents the new condition as retained verbatim.
- **Required fix:**
  - Declare the cross-interval zero-`not_warranted` and zero-Undecided predicates as deliberate changes to REQ-22's "during the window" clause, in `ADR.md:48` and the card Changes.
  - Reword `ADR.md:30` to: "Retain … verbatim, and add the declared cohort-wide zero-not_warranted and zero-undecided conditions across gaps."

### ADR012-R6-10: Summary statements omit the R5-01 wiring caveat (minor)

- **Location:** ADR-012 card decision bullet `:1009` ("workflow/transport/sandbox wiring belongs to the implementation PR"); STATUS `:9` ("Workflow/transport/sandbox wiring is deliberately left to implementation under those invariants").
- **Problem:**
  - The card is governed design. Read alone, it still hands wiring to the implementation PR. It omits the path R5-01 added: a wiring change that alters requirements needs its own adopted ADR.
  - REQ-161, STATUS `:121`, plan `:61,90` and the references are correct.
- **Required fix:** append "; requirement-changing wiring needs its own adopted zero-issue wiring ADR (REQ-161)" to card `:1009` and to STATUS `:9`.

## Checks that found no issue

- **Accepted ADRs.** ADR-000, ADR-003, ADR-007 and ADR-008 differ from base only on line 3 (challenger, all four bodies). README rows `:18,21,25,26` and the card relations agree. ADR-011's card and README row match base.
- **Unchanged requirements.** REQ-160, REQ-162, REQ-163 and REQ-168 are byte-identical to base (challenger). No workflow, key, transport or command is selected. `policy-file.md:84` mentions the superseded key only to forbid it.
- **Declared change set.** REQ-10/22/26/27/39/47/48/51/78/161/166, the Activation callout and Appendix B match `ADR.md:4` and card `:1010` (challenger, diff against base).
- **Cross-file consistency** (peer). The exit codes, Days formula, three-term formula, default-version rule, cutoff (`recorded_at` ≤ `as_of`, default C) and unknown bound agree across `ADR.md`, REQ-22/27/78, `policy-file.md`, `cli-commands.md` and `ci-integration.md`. The exceptions are ADR012-R6-06 and ADR012-R6-07.
- **Fixtures.** All 8 fixtures round 5 required appear in `ADR.md:56–62`, plan B4-S `:61` and `ci-integration.md:106`.
- **Ledger.** R-077 cites `round-5.md`, and its caveat column leaves the design choices to the ADR. R-064, R-069 and R-070 are unchanged in substance. STATUS `:131` gives the next free ID as R-078. No reused ADR or R-ID was found.
- **Cross-gap re-triage** (challenger). Triage at `as_of` allows an explicit, append-only maintainer correction of a prior-interval `not_warranted`. A gap alone cannot clear it.
- **Guardian parity.** The comparison, protected-base and empty-override obligations stay consistent. No new REQ-27 parity defect was found.

## Considered and not raised

- **Disclosing author and bot concentration (experiment-peer, minor).** Bots and forks belong to the guardian's own population, and the guardian gates them equally. Per-tree counting with `distinctPRs` was settled (`round-3.md:338,340`). The implementation may add author-count or bot-share diagnostics. They are not needed for REQ-22.
- **A durable publication sequence with fail-closed pending corrections (challenger's R6-03 option).** Not chosen; see ADR012-R6-03. Once both clocks are provider time, the provider-clock term closes the poisoning, and the bounded delay is disclosed under R5-06.
- **Merge against sealing as a separate issue (experiment-peer finding 9).** Merged into ADR012-R6-02, because it is the same rule about late facts.
- **A separate acceptance-mapping issue (challenger finding 5).** The missing origin and exclusion fixture is folded into ADR012-R6-05. Mapping the R5-10 fixtures into B4-S is folded into execution step 11.
- **GitHub's `run_started_at` being nullable** (challenger cites the [workflow-run REST schema](https://docs.github.com/en/rest/actions/workflow-runs#list-workflow-runs-for-a-repository)). The council does not rely on it, and the lead did not re-verify it [INFERENCE]. The existing "unreliable start time prevents certification" and the ADR012-R6-03 refusal already cover a missing provider time. If the writer cites this schema to justify a fix, it must be added to R-070 under the ledger rule.
- **Exact-prefix restoration of rounds 4 and 5.** Main's attestation is accepted, because the council cannot hash. These records are outside the signed set.

## Peers and where they changed the outcome

- **`innovation-challenger`** raised 1 major and 4 minor findings. It rated R5-03, R5-06, R5-08 and R5-10 Partly and every other R5 issue Fixed. Adopted:
  - the writer-clock poisoning, which became ADR012-R6-03 at major. The lead found it independently. The challenger's combined fixture (fast clock, then correction, then quiet reconcile, asserting the default result) was adopted as the acceptance test;
  - the conflict between live diagnostics and the immutable dossier, which became ADR012-R6-07. This was new to the lead and corrects the council's own `round-5.md:316` assumption;
  - the no-start events, merged into ADR012-R6-01;
  - the default-version tie-break, merged into ADR012-R6-06;
  - acceptance coverage, folded into ADR012-R6-05 and the execution outline.

  Not adopted: its recommended sequence-plus-fail-closed option for ADR012-R6-03.
- **`experiment-peer`** raised 3 major and 8 minor findings. It rated R5-02 Partly and the rest Fixed, most with new residuals. Adopted:
  - collector-skipped events, which became ADR012-R6-01 at major. The lead had found the missing time binding; the peer's citation of `round-2.md:272` and `round-2-response.md:31` showed the provider attestation was lost in the scope cut. The peer's open question on approval-pending forks is answered there: assess or gap, never exclude;
  - the frontier upper bound and listing lag, which became ADR012-R6-02 at major. The lead had found the future-seal half; the peer added listing lag and refusal traces. The lead chose a late-gap record over the peer's `ready = null`;
  - undecided events cleared by a gap, which became ADR012-R6-04 at major. This was new to the lead, which added the provisional-ordering point from `policy.go:416–424`;
  - the clock running ahead, merged into ADR012-R6-03 with severity **raised** from minor to major, because it defeats the R5-06 fixture while ready is true;
  - the representative tie-in, which became ADR012-R6-05 and matches the lead;
  - frontier and triage ties, which became ADR012-R6-06;
  - the dossier and reader exits, merged into ADR012-R6-07;
  - undefined terms, which became ADR012-R6-08 (the path-filter point matches the lead);
  - merge bypass, merged into ADR012-R6-02;
  - the undeclared veto, which became ADR012-R6-09.

  Not adopted: finding 11 (concentration disclosure).
- **Lead-only:** ADR012-R6-10 (summary statements) and the reader-clock exit boundary within ADR012-R6-07, which the peer also noted.

## Verdict

**ADR-012: 10 issues open (4 major, 6 minor). Not signed off.** Open: ADR012-R6-01, -02, -03, -04, -05, -06, -07, -08, -09 and -10. All 12 round-5 issues are closed:

- R5-01, -04 to -12 are fixed;
- R5-02 and R5-03 are fixed as prescribed, and their remaining gaps are carried as ADR012-R6-01 and ADR012-R6-02.

No digests are signed in this round. The 13 digests above identify the bytes reviewed.

## Council memo

1. **Recommendation.** Do not sign off this round. Keep the narrowed, invariants-only ADR and apply the 10 fixes with the recommended options above, then submit round 7 on the revised signed set, with Main supplying fresh digests for all 13 files.
   - The round-5 fixes were applied faithfully and consistently across every mirror.
   - The open defects sit in one layer: binding time and completeness. That covers event time against attempt time, the upper bound of the frontier, the writer clock against the provider clock, gap scope for triage, and what the reproducible dossier contains.
   - To shrink the surface for round 7, the writer should consolidate these rules into one "clocks, frontier and late facts" paragraph in `ADR.md` and mirror it once in REQ-78, instead of scattering clauses.
   - All fixes are text-level and select no mechanism.

2. **Alternatives.**
   - **(A) Fix within the narrowed ADR (recommended).** Event-keyed completeness, a provider-bounded frontier, late-gap records, provider-clock `recorded_at`, a cohort-wide triage-completeness veto and a canonical dossier boundary. This keeps intervals, gaps and cutoffs as reviewed.
   - **(B) Key observation membership on event time instead of attempt start.** This removes the two-clock problem in R6-01, but it reopens the R1-04 `started_at` choice and changes Days. Not recommended.
   - **(C) A monotonic publication sequence for ordering, plus a provider-clock cutoff stamp.** This is equivalent to the R6-03 fix and acceptable if the writer prefers it.
   - **(D) Classify approval-pending forks as `not_assessed`.** Simpler wiring, but a content-informed maintainer choice selects the population, contrary to `ADR.md:16`. Rejected.
   - **(E) Sign off and carry these items as implementation obligations.** This breaks design rule 1, and R6-01 to R6-04 change what counts as activation evidence. Rejected.

3. **Evidence.** The issue sections above cite exact lines. The key ones:
   - `ADR.md:16,17,19,20,21,28,30,36,44,48,56–62`;
   - `scout-design.html:589,600,612,615,707,1009,1010`;
   - `cli-commands.md:126,128,130,132`; `policy-file.md:78,80,82`; `ci-integration.md:101`; STATUS `:9`;
   - `round-1-response.md:21`; `round-1.md:195`; `round-2.md:272,276`; `round-2-response.md:31`; `round-5.md:190,193,316`;
   - `policy.go:403–424`; `scout-guardian.yml:20–21`.

4. **Risks.**
   - **Provider timestamps** (R6-01/02/03). If the provider exposes no reliable event, listing or run-start time, the implementation must gap or refuse rather than certify. This is conservative and may delay readiness.
   - **Fork assessment without approval** (R6-01) may need wiring that changes requirements, and therefore a wiring ADR under R5-01. That route already exists.
   - **The stricter Undecided rule** (R6-04) blocks the exact version until old events are triaged. This is intended; triage is append-only and visible.
   - **Overspecification drift.** Each fix states only an observable invariant plus a fixture. Trigger, cadence and transport stay unselected.
   - **Reversibility.** All changes are document-level and nothing is adopted. Schema 13 still ships, and a superseding ADR can replace this one.

5. **Execution outline.**
   1. R6-01: add `event_at`, the closed list of provider-attested `not_assessed` reasons, collector failure as a gap, fork assessment regardless of approval, and the event-keyed frontier condition. Edit `ADR.md:16,17,19,58`, REQ-27, REQ-78, `ci-integration.md:101` and `cli-commands.md:126`.
   2. R6-02: add the enumeration-time bound on t, late-gap records with publishing-CAS `recorded_at`, and merge/import obeying sealing. Edit `ADR.md:19,36,58`, REQ-48, REQ-78 and `cli-commands.md:126`.
   3. R6-03: replace "writer UTC now" with the provider start time of the publishing run, and refuse publication when it is unavailable. Edit `ADR.md:28`, REQ-78, `cli-commands.md:130` and `policy-file.md:82`.
   4. R6-04: add the cohort-wide zero-Undecided predicate, required for provisional results too, and the prior-interval undecided count. Edit `ADR.md:20,30`, REQ-22, REQ-78, `policy-file.md:80` and `cli-commands.md:128`.
   5. R6-05: take status, PRRuns membership and the origin from the representative at `as_of`. Edit `ADR.md:21`, REQ-78, `cli-commands.md:128` and `policy-file.md:78`.
   6. R6-06: make frontier records strictly increasing and immutable; allow one triage row per event per publication or add a sequence; order the default version by (`started_at`, attempt ID). Edit `ADR.md:19,28,44`, REQ-78 and `cli-commands.md:126,130`.
   7. R6-07: define the canonical dossier and digest, the live envelope, and the exit rule (`as_of` ≤ C evaluates; exit 3 only above max(C, reader now)). Edit `ADR.md:44`, REQ-78 and `cli-commands.md:130,132`.
   8. R6-08: define eligibility and the selector (guardian trigger configuration plus eligibility-logic version; pins excluded), provider enumeration (events) and in-scope attempt. Edit `ADR.md:16,44`, REQ-27, REQ-78, `ci-integration.md:101` and `cli-commands.md:126`.
   9. R6-09: declare the cross-interval zero-`not_warranted` and zero-Undecided predicates in `ADR.md:48` and the card Changes, and reword `ADR.md:30`.
   10. R6-10: add the wiring-ADR caveat to card `:1009` and STATUS `:9`.
   11. Extend `ADR.md:56–62`, plan B4-S `:61` and `ci-integration.md:106` with the new fixtures. Map the existing R5-10 fixtures into B4-S explicitly. The new fixtures:
       - a collector-skipped "unstarted" event;
       - an approval-pending fork;
       - sealing past enumeration time;
       - a lagged-listing late gap;
       - merge of a fact from before C;
       - fast clock, then correction, then quiet reconcile (asserting the default result);
       - undecided events across a gap, plus the rare-security variant;
       - the representative origin and the not_inspected/not_applicable cases;
       - a minima change without a C advance;
       - tied default-version attempts across export and import;
       - the digest stable after a later publication;
       - a reader clock behind C.
   12. If a provider-API source is cited to justify a fix, add it to R-070. Extend R-077's round list with `round-6.md`.
   13. Write `round-6-response.md` mapping each issue. Main supplies the SHA-256 of all 13 signed files and of the response, then round 7 follows.
