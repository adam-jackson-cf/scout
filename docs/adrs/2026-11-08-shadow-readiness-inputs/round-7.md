# Round 7: ADR-012 (shadow-window readiness inputs)

Date: 2026-11-08. This round follows design rules 1–3 and the ledger rules in `CLAUDE.md`, and the ADR rules in `docs/adrs/README.md`. ADR-012 is an agent decision made under the owner's standing delegation (R-064). Where a fix involves a choice, the council names the option it recommends under that delegation.

The round reviews the bytes revised after round 6 (`round-6-response.md`). It has three parts:

1. Check every ADR012-R6-01..10 fix against its required fix.
2. Re-confirm that every earlier closure (R1–R5) still holds in the current bytes.
3. Look for new issues, numbered ADR012-R7-xx.

The council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review. It checked R6 closure at every named location and tried to build an implementation that meets the literal text but breaks REQ-22, REQ-27 or REQ-78, concentrating on how the new R6 rules interact. It compared the signed files with base by reading the raw text.
- **`experiment-peer` (Opus):** methodological review. It treated readiness as a measurement and looked for confounds, inflation and deflation paths, gaming, results that cannot be reproduced, and terms the R6 fixes left undefined. It also checked fixture coverage and the ledger.

The review was read-only. Nobody edited files, executed product code, ran model calls outside the council, or touched `.codegraph`. The lead read the required codegraph skill references. Codegraph was not used, because the material under review is documents and the council had no shell. Searches used `read` and `grep`. The experiment-peer ran one web search and the lead ran one, both to check GitHub's skip-instruction behaviour (ADR012-R7-04).

**Main computed and verified the digests below with `shasum`. The council did not: the lead and both peers had read-only tools and no shell.** This round is not a sign-off, so the digests only identify the bytes reviewed. All paths are under `docs/`.

| File | SHA-256 (Main) |
|---|---|
| `adrs/2026-11-08-shadow-readiness-inputs/ADR.md` | `8938352129b57354c3a5eca04b83e84cf3f1f88c41e33e6935a60e5e8d52697b` |
| `scout-design.html` | `32c728ab62761a4b79a817253576f1bee2d33692a60b5d0ab2e43b5ee993d2ca` |
| `adrs/README.md` | `cde415ff10b659a073d010f8a48c281d6f75a5a1cba466962d589e937b94587f` |
| `STATUS.md` | `cf667e7251454a6557e1030639e0b30547dbb6472ba73f36537be936f2dc0453` |
| `plans/2026-11-06-remaining-work.md` | `25039a435ff9d6f0a8d84cd4512364763f16e327dc751c259890abcf04160ea7` |
| `references/policy-file.md` | `2c6afa56099b48ea578da025622b4b5cbe13fe33e1eb0d6722674533a5cf476a` |
| `references/ci-integration.md` | `be8eedce37e1d82cb6563dbdfff5ec8dff481bfe0b6775983daed9967226002b` |
| `references/cli-commands.md` | `40868054552cc2b3d210ec17a61a1e564d41c53e797aaccd0c5c2202359777e8` |
| `research-ledger.md` | `26da601acd0b8bac968f4461ab2465cfe40ed910c3208418ab5f2a31cfa638fd` |
| `adrs/2026-10-19-v1-decisions/ADR.md` (ADR-000) | `1cd521fd96439d87275ba353f7b05451c3e0c4d556ee665529aab85970ed81cb` |
| `adrs/2026-10-23-decisions/ADR.md` (ADR-003) | `973d24c4cfdfdd98acb4070a7bd72ca0c0e1ddf303bd4ccec8f92fe1cf117cc5` |
| `adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) | `ee593b6241814c24035671d666566bc2b7c399135867226c26e9626f3936abc8` |
| `adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) | `80e3375f81656a8856cf662cbec99a36d8e00bc451ff3c3592925d2293f44d95` |

The following record is not signed. Main computed its SHA-256:

| File | SHA-256 (Main) |
|---|---|
| `round-6-response.md` | `dac32c7ce23c5fe9b7cc47b78cfe17c3ee988275dbbcff1f141f97616b9fad0f` |

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11`, branch `agent/policy-shadow-adrs`, based on origin/main `7bb73e52`. Every change is staged and uncommitted.
- Base bytes come from `/Users/adamjackson/Projects/work/scout-adr11b`.

**Files read.** Line numbers below refer to the raw worktree files.
- ADR-012 `ADR.md:1–68`, all lines:
  - `:3–6` status, revision and sources; `:10–12` decision and delegated narrowing;
  - `:16–22` items 1–7, where `:19` is the consolidated "Clocks, frontier and late facts" paragraph;
  - `:26` triage authorization; `:28` triage rows; `:30` ShadowReady;
  - `:34–40` persistence, manifest/merge, cutover and writer boundary;
  - `:44` CLI; `:48–52` amendments, sources and route;
  - `:56–62` implementation acceptance invariants; `:66–68` folder records.
- `round-6.md` (all 413 lines) and `round-6-response.md` (all lines).
- `round-2.md:263–276,419–428,498`, checking the R2-05 cancellation settlement.
- `scout-design.html`, raw:
  - the Activation callout `:449`;
  - REQ-22 `:589`, REQ-78 `:600`, REQ-26 `:611`, REQ-27 `:612`, REQ-48 `:615`, REQ-51 `:618` and REQ-161 `:707`, plus grep context for REQ-10/39/47/160/162/163/166/168;
  - the cards for ADR-000 `:736`, ADR-003 `:807`, ADR-007 `:901`, ADR-008 `:925` and ADR-012 `:1007–1011`.
- Base `scout-design.html:449,618`.
- `cli-commands.md:37–40,120–134`, `policy-file.md:23,76–84` and `ci-integration.md:97–106`.
- Plan `:14–15,35,61–62,89–92,100`; STATUS `:7,9,121,129–131`; README `:12–30`; `research-ledger.md:72–76`.
- Status line 3 of ADR-000, ADR-003, ADR-007 and ADR-008.
- Code: `.github/workflows/scout-guardian.yml:20–55`, covering the trigger, the concurrency group and `cancel-in-progress: true`.
- The challenger compared the four historical ADR bodies and the REQ-160/162/163/168 rows with base.

## Issue summary

**9 distinct issues are open: 3 major and 6 minor.**

| ID | Severity | Summary |
|---|---|---|
| ADR012-R7-01 | major | A refused late fact can never be "published", so it blocks C from advancing for good. The late record's time coordinate, the source of its stamp and the pending disclosure are also undefined |
| ADR012-R7-02 | major | Turning a late fact into a gap discards its would_block events, so they escape the cohort-wide zero-not_warranted and zero-Undecided vetoes. Merged or imported triage rows at or before C are undefined |
| ADR012-R7-03 | major | Nothing binds a PR event to the attempt that accounts for it, so an attempt on another head can "account" a failing head. "Latency creates a gap" is ambiguous |
| ADR012-R7-04 | minor | A "provider-recorded cancellation/skip" cannot be checked against the list's own rule that reasons never depend on the author or a maintainer. Manual cancels and skip markers are not distinguished |
| ADR012-R7-05 | minor | The "exactly" field list for the canonical dossier leaves out the resolved `as_of`, repository, selector and population, and has no versioned encoding |
| ADR012-R7-06 | minor | How attempt IDs compare in tie-breaks is undefined |
| ADR012-R7-07 | minor | The disclosed pin tuple leaves out catalogue, which contradicts the reason R6-08 gave for excluding catalogue changes from the selector |
| ADR012-R7-08 | minor | Mirrors and terms drift: REQ-27 and `ci:101` lack required R6-01/R6-08 elements; `policy-file:82` uses the removed term `collectedThrough`; C is undefined; plan `:90` is stale; one fixture is missing from B4-S and `ci:106`; REQ-51 lacks the wiring caveat |
| ADR012-R7-09 | minor | The overview Activation callout leaves out inspected-only exposure and the cross-interval zero-not_warranted and zero-Undecided vetoes |

## Closure of round-6 issues

"Fixed" means the required document fix is present in the signed bytes at every named location. It does not mean the implementation is validated. "Fixed as prescribed" means the council's own prescription was applied but left a gap, which is carried as a new R7 issue. "Partly" means a named location lacks a required element. Where the peers split, the lead adjudicated as shown.

| Issue | Status | Evidence / remainder |
|---|---|---|
| ADR012-R6-01 | **Partly** | `event_at`, the closed list, collector failures as gaps, fork assessment regardless of approval, and the two-part condition for advancing C are in `ADR.md:16,17,19,57–58`, REQ-78 `:600`, `cli-commands.md:126` and `ci-integration.md:101`. CI makes the frontier condition binding by reference: "Event/attempt completeness … follow ADR-012/REQ-78". REQ-27 `:612` has the population, the closed list and the gap rules. It lacks required item 1 (membership keyed on `started_at`) and item 6 (the two-part condition for C), and gives no reference to them. That remainder is carried in ADR012-R7-08. The prescription itself left PR events unbound to attempts: ADR012-R7-03. The challenger rated this Partly; the peer rated it Fixed as prescribed. |
| ADR012-R6-02 | **Fixed as prescribed** | The enumeration-time bound on t, late-coverage gaps stamped by the publishing CAS, invariance for earlier `as_of`, the late-gap count, and merge/import obeying sealing are in `ADR.md:19,36,58`, REQ-48 `:615`, REQ-78 and `cli-commands.md:126`. The fixtures are in plan `:61` and `ci:106`. The gaps between this rule, attempt completeness and the cohort-wide vetoes are ADR012-R7-01 and ADR012-R7-02. |
| ADR012-R6-03 | **Fixed** | `recorded_at = max(publishing-run provider start, parent C+1ms, parent maximum recorded_at+1ms)` is recomputed on every retry, and publication is refused when the provider start is unavailable. This is in `ADR.md:19,28,60`, REQ-78, `cli-commands.md:130` and `policy-file.md:82`. The fast-clock, correction, quiet-reconcile fixture asserts the default result (`ADR.md:60`, plan `:61`, `ci:106`). `policy-file.md:82` writes the C term as "parent collectedThrough": ADR012-R7-08. |
| ADR012-R6-04 | **Fixed** | Cohort-wide zero-Undecided is required before every ready result, including the provisional one, and the prior-interval undecided count is reported. This is in `ADR.md:20,30,59`, REQ-22 `:589`, REQ-78, `policy-file.md:80` and `cli-commands.md:128`, with fixtures in plan `:61` and `ci:106`. Late facts bypass it: ADR012-R7-02. |
| ADR012-R6-05 | **Fixed** | The representative at `as_of` supplies comparability, inspection status, PRRuns membership and the origin, as `ADR.md:21` states. The text is mirrored in REQ-78, `cli-commands.md:128` and `policy-file.md:78`. Card `:1010` states the inspected-comparable origin. The day-10 origin, not_inspected/not_applicable and later-incomplete-representative fixtures are at `ADR.md:62`, plan `:61` and `ci:106`. |
| ADR012-R6-06 | **Fixed** | Frontier records are immutable, keyed by C and created only on a strict increase (`ADR.md:19`). There is one triage row per event per publication (`:28`). The default version is the maximum (`started_at`, attempt ID) (`:44`). These are mirrored in REQ-78 and `cli-commands.md:126,130`, with fixtures at `ADR.md:61`. How attempt IDs compare is still undefined: ADR012-R7-06. |
| ADR012-R6-07 | **Fixed as prescribed** | The canonical dossier, the live envelope and the exit rule (`as_of` ≤ C evaluates; exit 3 only when `as_of` > max(C, reader now)) are in `ADR.md:28,44`, REQ-78 and `cli-commands.md:130,132`, with fixtures at `ADR.md:61`. The council's prescribed field list left out `as_of`: ADR012-R7-05. |
| ADR012-R6-08 | **Partly** | Eligibility, the selector and provider enumeration are defined in `ADR.md:16`, REQ-27, REQ-78, `ci:101` and `cli:126`. "In-scope attempt" is defined in `ADR.md:44`, REQ-78 and `cli:126`, but is missing from REQ-27 and `ci:101`, which round 6 named. That remainder is in ADR012-R7-08. The reason given for excluding catalogue changes from the selector does not hold: ADR012-R7-07. The challenger rated this Partly; the peer rated it Fixed. |
| ADR012-R6-09 | **Fixed** | `ADR.md:30` now says "add the declared cohort-wide zero-not_warranted and zero-Undecided conditions across gaps". `ADR.md:48` and card Changes `:1010` declare the deliberate amendment to REQ-22's during-the-window clause, including provisional readiness. The overview callout is ADR012-R7-09. |
| ADR012-R6-10 | **Fixed** | Card `:1009` and STATUS `:9` end with "requirement-changing wiring needs its own adopted zero-issue wiring ADR (REQ-161)". REQ-51 `:618` has the same gap but was not named in R6-10: ADR012-R7-08. |

## Re-confirmation of earlier closures (R1–R5)

| Closure | Holds? | Evidence |
|---|---|---|
| R1-01, R3-06, R4-05: version-scoped event identity, many-to-many links, union across attempts | Yes, except for late facts | `ADR.md:20,22,34,36`; REQ-27. Late attempts' links are lost: ADR012-R7-02 |
| R1-03/06/07, R2-13/14: exact cohorts, observation dedup, inspected-only exposure, distinctPRs | Yes | `ADR.md:20–21`; REQ-22/78; `policy-file.md:78–80` |
| R1-04: membership keyed on `started_at`; elapsed-days formula | Yes | `ADR.md:19,21` |
| R1-05: a collection boundary can be recovered | **Reopened through an interaction** | Refusing to publish late attempts conflicts with the unconditional "published" condition (`ADR.md:19`): ADR012-R7-01 |
| R2-05: concurrency supersession and cancellations | Yes in substance | `ADR.md:17`. Checkability: ADR012-R7-04 |
| R1-08/09/11/14, R2-10, R3-01, R4-07/10: CI/local/legacy separation, migration, v2 cutover, query-only v13 | Yes | `ADR.md:34–38,44`; REQ-10 `:610`; REQ-48 `:615`; `ci:104` |
| R3-02/07, R4-02/08/11: protected-source authorization, current maintainer authority, replay idempotence, writer isolation | Yes | `ADR.md:26,40`; REQ-26 `:611`, REQ-51 `:618`, REQ-161 `:707`; `ci:102–103`; `cli:134` |
| R3-03, R4-06: guardian comparison and empty override inputs | Yes | `ADR.md:17–18`; REQ-27; `ci:101` |
| R2-11, R3-05, R4-03/04, R5-03/05/07: interval selection, bounds, triage cutoff, frontier-pinned minima | Yes | `ADR.md:19–20,28,44`; REQ-78; `cli:126–132`. `policy-file:82` term: ADR012-R7-08 |
| R5-01: wiring-ADR gate | Yes | `ADR.md:12,56`; REQ-161; `ci:99`; plan `:61,90`; STATUS `:9,121`; `cli:134`; `policy-file:84`. REQ-51 omission: ADR012-R7-08 |
| R5-02: population; R5-04: cross-gap not_warranted; R5-06: liveness on quiet repositories | Yes, except for late facts | `ADR.md:16–20`. Late facts: ADR012-R7-01 and ADR012-R7-02 |
| R5-08: Days and declared counting changes | Yes | `ADR.md:21,48`; card `:1010` |
| R5-09: scanner-output failure is comparison-none | Yes | `ADR.md:17,62`; REQ-27; `ci:101` |
| R5-10: exact cohort, default version, pin mixtures, exits | Yes in `ADR.md` | `ADR.md:44,62`. The "disclose pin mixtures" fixture is missing from B4-S and `ci:106`: ADR012-R7-08 |
| R5-11: declared change set; "selects no …" | Yes | `ADR.md:4,48`; card `:1010` |
| R5-12: restored round records | Yes | `round-4.md:390–392` and `round-5.md:336–338,380–392` carry their verdicts. Exact-prefix identity rests on Main's attestation, because the council cannot hash |
| Historical status lines and index | Yes | ADR-000/003/007/008 `:3` match README `:18,21,25,26` and card relations `:736,807,901,925,1007` |

## Are the invariants testable and enough for REQ-22, REQ-27 and REQ-78?

| Requirement need | Testable and sufficient? | Where / gap |
|---|---|---|
| REQ-22: counting unit, deduplication, exact cohort | Yes | `ADR.md:20–22` |
| REQ-22: unknown ≤ 10 %, counted once | Yes | `ADR.md:21` |
| REQ-22: 14 days, inspected origin, `as_of` endpoint | Yes | `ADR.md:21,62` (R6-05) |
| REQ-22: zero `not_warranted` per exact version, across gaps | **Partly** | Late facts escape it (ADR012-R7-02) |
| REQ-22: maintainers triage every persisted event; zero-Undecided | **Partly** | Late facts' events are never persisted (ADR012-R7-02) |
| REQ-22: maintainer-only, append-only, idempotent triage | Yes | `ADR.md:26,28`. Merged rows at or before C: ADR012-R7-02 |
| REQ-27: guardian-equivalent comparison and parity | Yes | `ADR.md:17–18` |
| REQ-27: population independent of content, author and approval | **Partly** | An attempt on another head can account an event (ADR012-R7-03); manual cancels and skip markers (ADR012-R7-04) |
| REQ-78: complete-interval accounting | **No** | Event-to-attempt binding is missing (ADR012-R7-03); a late fact blocks C (ADR012-R7-01) |
| REQ-78: historical reproduction and `dossier_digest` | **Partly** | `as_of` and encoding (ADR012-R7-05); attempt-ID order (ADR012-R7-06) |
| REQ-78: triage liveness | **Partly** | C can freeze after a late fact (ADR012-R7-01) |
| Governance of deferred wiring | Yes, with one mirror leftover | ADR012-R7-08 (REQ-51) |

## Issues

### ADR012-R7-01: A refused late fact blocks the frontier for good; its coordinate, stamp source and pending disclosure are undefined (major)

- **Location:** `ADR.md:19` ("every attempt started≤t is terminal and published in that state"; "Refuse late additions at/before the destination's sealed C, but never drop them: append an immutable late-coverage gap with publishing-CAS recorded_at. For as_of≥that stamp, the late fact's point is a gap"); `ADR.md:36` ("Import follows the same rule: pre-C additions become publication-stamped late gaps"); `ADR.md:44` (live envelope); `ADR.md:58`; REQ-48 `:615`; REQ-78 `:600`; `cli-commands.md:126,130`.
- **Problem:**
  - **Frontier deadlock.** The attempt condition has no recorded-gap alternative, unlike the event condition. A late attempt with `started_at` ≤ C is refused, so it is never "published in that state".
  - **Counterexample** [INFERENCE: constructed from the text, not executed]:
    1. C is day 30 and the default result is ready.
    2. On day 31, reconcile finds an authenticated terminal attempt that started on day 20. The earlier attempt listing missed it.
    3. The writer appends a late gap stamped day 31.
    4. Every candidate t > day 30 still covers an attempt that started on day 20 and was never published. C cannot advance.
    5. The default `as_of = C` never shows the late gap, and never shows any later triage correction. Ready stays true indefinitely.

    This is the R6-03 harm class: disclosure of rows after the cutoff does not stop ready from being true. It also reopens R1-05's requirement that a collection boundary can be recovered. The `ADR.md:58` fixture checks only a chosen cutoff, not that the default cutoff can reach the late gap.
  - **Why late attempts are expected.** The enumeration bound covers PR events (`ADR.md:16,19`). Attempt completeness needs a listing of attempts, with its own lag. Late attempts are therefore an ordinary outcome of listing lag, not only an artifact of import.
  - **Coordinate undefined.** "The late fact's point" names no clock. An event on day 1 whose attempt started on day 10 can be gapped at day 1 or at day 10. That changes the containing interval and the Days origin (challenger).
  - **Stamp source undefined outside publication.** `scout import` loads into a local ledger (`cli-commands.md:38`) without a publishing CAS. A local import or merge of a fact before C has no defined `recorded_at`.
  - **No pending disclosure.** A late record stamped C+1 ms while an in-flight attempt holds C stays invisible. Nothing discloses it, and a maintainer can open an activation PR meanwhile (experiment-peer).
- **Required fix (recommended option):** state the following once in the `ADR.md:19` paragraph, and mirror it in REQ-78, REQ-48 and `cli-commands.md:126,130`.
  1. An appended late record discharges the completeness obligation of the attempt or event it covers. The condition for advancing C then reads: every attempt started ≤ t is terminal and published in that state, or covered by a published late record. C may then advance within the enumeration bound.
  2. A late record's effective time is fixed by class:
     - an unaccounted event uses `event_at`;
     - a late attempt uses its `started_at`, and also its accounted event's `event_at` where that differs.

     Each effective time is a point gap for interval-scoped counts.
  3. Only a trusted publishing CAS assigns `recorded_at` to a late record. A local store that holds late facts before publication keeps them unstamped and pending. No readiness query counts them at any `as_of`, and the live envelope discloses them.
  4. Add "late records stamped after `as_of`, and pending late facts" to the labelled live envelope (`ADR.md:44`).
  5. Fixtures:
     - after a day-20 attempt is discovered once C = day 30, reconcile advances C, the default result reflects the late record, and the `as_of` = day 30 dossier is identical;
     - an event and attempt with different times place the gap at the defined coordinates.

### ADR012-R7-02: Late facts escape the cohort-wide vetoes, and merged triage rows at or before C are undefined (major)

- **Location:** `ADR.md:19` (late facts are refused and replaced by a gap); `ADR.md:20` (the vetoes cover "every persisted CI event … with a linked assessed cohort attempt started≤as_of, across all intervals"); `ADR.md:30`; `ADR.md:36` ("Merge unions immutable facts/links/history only subject to the sealed-frontier/late-gap rule"); REQ-22 `:589`; REQ-48 `:615`; REQ-78; `cli:126,128`. The council's own prescription is at `round-6.md:170–172`.
- **Problem:**
  - **Counterexample** [INFERENCE: constructed from the text, not executed]:
    1. Attempt X starts at 09:58 and finds new would_block event E.
    2. C = 10:00 is sealed without X.
    3. X arrives later and is refused. Only a gap at 09:58 is appended, so E is never persisted.
    4. If 09:58 lies in an earlier part of the interval, the interval containing `as_of` still spans 14 days and 5 inspected observations after the point gap.
    5. Readiness certifies, although the trusted collector observed E and nobody triaged it.

    The same works when an authenticated import brings X with an authorized `not_warranted` row (challenger).
  - This is the harm round 6 rated major under R6-04, where a collection artifact cleared an event the cohort had not triaged. REQ-22's "maintainers triage every persisted would_block event" is defeated because E never becomes persisted.
  - **Triage rows.** The late-gap rule speaks of a fact's "point" and of interval-scoped counts. It does not cover triage or history rows. A merged `not_warranted` row whose `recorded_at` is at or before the destination's C either loses the veto, if refused, or rewrites dossiers for `as_of` in [r, C], if inserted at r (experiment-peer).
  - "Across all intervals" can be read as leaving out attempts that started inside a gap.
- **Required fix (recommended option):** state the following in `ADR.md:19,20,36` and mirror it in REQ-22/48/78 and `cli:126,128`.
  1. Persist a late fact append-only, with its attempt, links and events, flagged late and carrying its publishing-CAS stamp. The point stays a gap for interval-scoped counts. Its events join the cohort-wide zero-not_warranted and zero-Undecided sets for every `as_of` ≥ the stamp, as Undecided unless authorized triage exists. This replaces "refuse … append a gap".
  2. The cohort-wide sets cover linked assessed cohort attempts whatever their interval, gap or late status.
  3. A triage or history row absent from the destination, with `recorded_at` ≤ the destination's C, is appended under a fresh publishing-CAS `recorded_at`. Its original action ID, actor, policy revision and `recorded_at` are kept as provenance, and the append is idempotent by action ID. A row whose original authorization cannot be verified from the record is refused and logged. Rows are never inserted at their original stamp.
  4. Fixtures:
     - a late attempt carrying a would_block event keeps the cohort not ready at `as_of` ≥ the stamp until the event is triaged warranted;
     - a late import carrying a `not_warranted` row blocks readiness after its stamp;
     - in both cases `as_of` = C₁ is unchanged.
  - **Acceptable alternative (challenger):** keep the affected exact cohort unmeasured until late evidence is reconciled. Not recommended, because a single lag would suspend readiness availability, while item 1 reuses the reviewed veto mechanism.

### ADR012-R7-03: PR events are not bound to the attempts that account for them (major)

- **Location:** `ADR.md:17` ("An eligible event is accounted only by a terminal assessed/comparison-none attempt, non_default_base, or provider-attested not_assessed…"); the attempt record fields in `ADR.md:17` (PR key, H, B, merge_base, `started_at`, outcome, "event links", where those links are would_block links per `:22`); `ADR.md:19`; REQ-27 `:612`; REQ-78; `ci:101`; `cli:126`.
- **Problem:**
  - Nothing says which attempt accounts which PR event.
  - **Counterexample** [INFERENCE: constructed from the text, not executed]:
    1. PR P receives `synchronize` E1 at head H1. The guardian finishes H1, because `cancel-in-progress` (`scout-guardian.yml:41–43`) supersedes only runs still in progress.
    2. H1 fires a would_block event, and the author pushes fix H2 as event E2.
    3. The collector assesses only H2 and treats both E1 and E2 as accounted for PR P.
    4. No gap is recorded, C advances, and H1's would_block event, possibly a false positive, is never persisted.
  - The collector, not the provider, then chooses which heads enter the population, which defeats R5-02 and R6-01. The R5-02 fixture ("omitted failing eligible PR") and the R6-01 fixture ("collector-skipped unstarted event") both pass.
  - Because the binding is not stored, an advance of C cannot be justified from exported state.
  - "Collector … latency … creates a gap containing event_at" can be read as gapping every event assessed with any delay. Under that reading a nightly collector would never be ready. Under the other reading, latency is a gap only when the event is left unaccounted. Two conforming implementations would compute different intervals.
- **Required fix (recommended option):** state the following in `ADR.md:17,19` and mirror it in REQ-27/78, `ci:101` and `cli:126`.
  1. Each attempt records the provider PR event IDs and `event_at` values it accounts for.
  2. An event is accounted only by a terminal assessed or comparison-none attempt that meets all three conditions:
     - same PR key;
     - H commit equal to the event's provider-recorded head SHA;
     - `started_at` ≥ `event_at`.

     Records for `non_default_base` and for closed-list `not_assessed` reasons name the event they account for.
  3. An event waiting for its attempt holds C back. A gap containing `event_at` is recorded only when the collector will not account that event, for example because of non-start, deferral past its implementation-chosen deadline, closure or merge before collection, or approval waiting. Reword "latency" in this sense.
  4. Fixture: H1 is assessed by the guardian, H2 is pushed later, and the collector assesses only H2. Then E1 is gapped or blocks C, and is never accounted.

### ADR012-R7-04: Cancellations and skips on the closed list cannot be checked against the list's own independence rule (minor)

- **Location:** `ADR.md:17` ("provider-recorded cancellation/skip before assessment, including its own concurrency supersession. These reasons cannot depend on outcome/content/labels/author/maintainer approval"); REQ-27/78; `ci:101`; `cli:126`.
- **Problem:**
  - GitHub does not run `pull_request` workflows when the head commit message contains `[skip ci]` or a related instruction ([GitHub docs, "Skipping workflow runs"](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/skip-workflow-runs); the lead re-checked this). That non-run is controlled by the author's content.
  - An author's or maintainer's manual cancel before assessment yields a provider "cancelled" conclusion, like a concurrency cancel. [INFERENCE: the run conclusion alone does not carry the cancellation cause.]
  - A loose implementation therefore excludes events the author chose. A strict one cannot use the concurrency exclusion that R2-05 settled (`round-2.md:267,272–273`) to stop the unknown rate tracking push frequency.
- **Required fix (recommended option):**
  - Exclude a cancellation only when it is verifiable as supersession or as cancellation by the provider: a later run of the same pinned lane and concurrency group for the same PR exists, or the provider records a system cancellation.
  - Exclude a skip only when the provider records it from the pinned lane's own configuration.
  - Manual cancellations and skip-instruction non-runs are assessed or gapped.
  - Mirror this in REQ-27/78, `ci:101` and `cli:126`.
  - Add a fixture: a `[skip ci]` head and a manual pre-assessment cancel are each assessed or gapped.
  - If the fix cites the GitHub skip documentation, add it to R-070 under the ledger rule.

### ADR012-R7-05: The canonical dossier field list leaves out `as_of`, identity and encoding (minor)

- **Location:** `ADR.md:44` ("Canonical dossier/dossier_digest covers exactly state fixed at/before as_of: cohort, interval, frontier record, pins, protected policy revision/effective minima, counts/exclusions, triage at as_of and readiness/provisional reason…"); REQ-78 `:600`, whose evidence is `{would_block, warranted, window_days, pr_runs, dossier_digest}`; `cli:130`; `policy-file.md:23`.
- **Problem:**
  - Two cutoffs between the same frontier records can give identical counts, integer Days and verdicts. A dossier that serializes exactly the listed fields then cannot say which cutoff it certifies (challenger).
  - The repository ID, selector version and path-filter population limit are disclosures fixed at `as_of` (`ADR.md:16,57`), but they are not in the "exactly" list.
  - No versioned canonical encoding is required, so a Scout upgrade can change the digest of identical state, and the stability fixture tests only one version (experiment-peer).
  - Activation evidence does not record `as_of`, so a reviewer cannot reproduce the digest without searching the C history.
- **Required fix (recommended option):**
  - Add the resolved `as_of`, the repository ID, the selector version and the path-filter population to the canonical dossier.
  - Require a canonical encoding that carries a format version.
  - Add `as_of` to the REQ-78 `evidence` record and to the `policy-file.md:23` evidence row.
  - Mirror this in `ADR.md:44`, REQ-78 and `cli:130`.
  - Add a fixture: identical counts at two cutoffs give different digests.

### ADR012-R7-06: How attempt IDs compare in tie-breaks is undefined (minor)

- **Location:** `ADR.md:20,21` ("started_at then attempt ID"); `ADR.md:44` ("maximum (started_at, attempt ID)"); REQ-78; `cli:126,128`; `policy-file.md:78,80`.
- **Problem:**
  - [INFERENCE] Provider start times are likely recorded to the second, so ties are plausible.
  - Comparing composite IDs as text or as numbers gives different representatives and default versions. Round 6 described this hazard for action IDs ("999 vs 1000"). Two Scout versions could therefore produce different dossiers from the same state.
  - The R6-06 fixture tests determinism within one implementation only.
- **Required fix (recommended option):**
  - Define the attempt ID's canonical form as the tuple (provider repository ID, run ID, run attempt, job ID), compared numerically field by field, wherever "attempt ID" breaks a tie.
  - Mirror this in REQ-78 and the references.
  - Add a fixture: with tied `started_at`, ID 1000 orders after ID 999.

### ADR012-R7-07: The disclosed pin tuple leaves out catalogue (minor)

- **Location:** `ADR.md:16` ("Scout/tool/catalogue/policy pin changes do not [start a new interval]"); `ADR.md:44`, REQ-78 and `cli:128` ("each distinct Scout/tool/policy pin tuple"); the attempt records a catalogue version (`ADR.md:17`); the council's premise is at `round-6.md:282`.
- **Problem:** round 6 excluded pin changes from the selector because per-tuple disclosure covers them. The disclosed tuple omits catalogue, so a catalogue change in mid-interval is neither a selector change nor disclosed.
- **Required fix:** disclose each distinct Scout/tool/catalogue/policy tuple in `ADR.md:44`, REQ-78 and `cli:128`.

### ADR012-R7-08: Mirrors and terms drift (minor)

- **Location and problem:**
  1. **REQ-27 `:612`** lacks R6-01 item 1 (membership keyed on `started_at`) and item 6 (the two-part condition for C), and gives no reference to REQ-78 for them.
  2. **"In-scope attempt"** (R6-08) is defined in neither REQ-27 nor `ci:101`.
  3. **`policy-file.md:82`** says "nondecreasing collectedThrough … default triage cutoff is collectedThrough" and "parent collectedThrough+1ms". That term appears nowhere else in the signed set. The same line gives only "terminal completeness" as the condition for advancing C, which contradicts `ADR.md:19`.
  4. **C** is used throughout `ADR.md:19,28,44` but is never defined in words.
  5. **Plan `:90`** summarizes "immutable terminal-complete frontiers, cross-gap exact-cohort false-positive veto". It omits event accounting, the provider bound and the Undecided veto.
  6. **B4-S (plan `:61`) and `ci:106`** omit the R5-10 fixture "disclose pin mixtures" (`ADR.md:62`).
  7. **REQ-51 `:618`**, a line ADR-012 added (base `:618` lacks it), reads "Workflow/sandbox/transport mechanics remain implementation choices bounded by these invariants". It has no caveat that requirement-changing wiring needs an ADR, which is the same class as R6-10.
- **Required fix:**
  - Add to REQ-27 a normative reference to REQ-78's frontier rule and its in-scope-attempt definition, and add the same in-scope reference to `ci:101`.
  - Define C once in `ADR.md:19` and REQ-78 as "C, the sealed provider-clock collection frontier".
  - In `policy-file.md:82`, replace `collectedThrough` with C and state both conditions for advancing C.
  - Update plan `:90` to "event- and attempt-complete, provider-bounded frontiers; cross-gap zero-not_warranted and zero-Undecided vetoes".
  - Append "pin-mixture disclosure" to B4-S and `ci:106`.
  - Append to REQ-51 "; requirement-changing wiring needs its own adopted zero-issue wiring ADR (REQ-161)".

### ADR012-R7-09: The overview Activation callout leaves out the new activation predicates (minor)

- **Location:** `scout-design.html:449`, edited by ADR-012 and listed in the change set at `ADR.md:4`, compared with REQ-22 `:589`, REQ-78 and `ADR.md:20–21,30,48`.
- **Problem:**
  - The callout reads "at least 5 comparable-base PR observations … with zero unwarranted would-blocks and unknown … in the window".
  - It leaves out:
    - the inspected-only rule for exposure;
    - the cohort-wide zero-Undecided precondition, which now also applies to provisional readiness;
    - the cross-interval scope of zero-not_warranted.
  - Five comparable observations that were not inspected, or a clean interval after a gap that follows an undecided event, satisfy the callout but fail the requirements. This is the same class as R6-10.
- **Required fix:** summarize the rule as at least 5 inspected comparable-base observations, with zero unwarranted and zero undecided would-blocks for the exact version across all intervals, and the unknown share at most 10 % in the window. Link to REQ-22/78 for the counting rules.

## Checks that found no issue

- **Accepted ADRs.** ADR-000, ADR-003, ADR-007 and ADR-008 differ from base only on status line 3 (challenger, raw-text comparison of all four bodies). Their digests equal those reviewed in round 6. README rows `:18,21,25,26` and card relations agree.
- **Unchanged requirements.** REQ-160, REQ-162, REQ-163 and REQ-168 have identical raw text to base (challenger). No workflow, key, transport, sandbox or command is selected (`ADR.md:12,40`; card `:1010`; `policy-file.md:84`).
- **Declared change set.** The changed rows REQ-10/22/26/27/39/47/48/51/78/161/166, the Activation callout and Appendix B match `ADR.md:4` and card `:1010` (challenger).
- **Triage order.** `recorded_at` ≥ parent maximum + 1 ms, with one row per event per publication, gives a total order per event within one lineage (`ADR.md:19,28`). The action-ID tie-break is reachable only through cross-lineage merge, which ADR012-R7-02 settles.
- **Provider-clock stamping.** `recorded_at` > parent C keeps dossiers at or before C invariant. With both clocks on provider time, the R6-03 poisoning path is closed.
- **Zero-Undecided ordering.** `ADR.md:30,59` require it before the rare-security provisional result. That corrects the order in `policy.go:416–424`, which is an obligation on later implementation; the ADR does not claim it is implemented.
- **Exit rule.** Identical in `ADR.md:44`, REQ-78 and `cli-commands.md:132`; unmeasured maps to exit 2.
- **Fixtures.** All 12 fixtures from round-6 step 11 appear in `ADR.md:57–62`, plan B4-S `:61` and `ci:106`. The R5-10 mapping is present apart from pin mixtures (ADR012-R7-08).
- **Ledger.** R-077 (`research-ledger.md:76`) links `round-6.md`. R-064, R-069 and R-070 are unchanged. STATUS `:131` lists R-001…R-064, R-069, R-070 and R-077, with R-065–R-068 and R-071–R-076 pre-allocated and next free R-078, which matches the ledger. No R-ID or ADR ID is reused.

## Considered and not raised

- **"predicates additionally includes"** (`ADR.md:20`, REQ-78, `cli:128`, `policy-file:80`). A grammar error that does not change the meaning. Correct it to "include" while making the R7 edits.
- **REQ-51 byte-identical import versus late records.** "Export → import → export round-trips" (`cli-commands.md:38`; REQ-48 `:615`) concerns a destination with no sealed C, where no late rule applies. ADR012-R7-01 item 3 settles imports into stores that have a C.
- **Days can shrink after a reassessment** (experiment-peer). Because the origin comes from the representative, Days can fall but never rise. Not raised.
- **A gap clears the unknown rate.** Unknown attribution is a rate measured afresh in each interval. Unlike Undecided, no untriaged item is lost. Not raised.
- **Whether a provider can attest complete enumeration.** [INFERENCE] GitHub's list endpoints give no explicit completeness guarantee, so the bound rests on the consistency lag the implementation must justify (`ADR.md:19`). This is already a stated obligation, with refusal as the conservative outcome. Recorded as a risk.
- **Byte identity of the signed set.** The peers compared raw text; neither could diff or hash. Main's `shasum` verification is relied on.

## Peers and where they changed the outcome

- **`innovation-challenger`** raised 2 major and 4 minor findings. It rated R6-01 and R6-08 Partly and the rest Fixed. Adopted:
  - the frontier deadlock after a refused late attempt, which became ADR012-R7-01 at major. It was new to the lead, and it **changed the lead's late-fact recommendation**: persisting late facts (R7-02) alone would not have made them visible to the default query;
  - late facts escaping the cohort-wide veto, merged into ADR012-R7-02. The lead found this independently; the challenger added the import variant with `not_warranted`. Its alternative (unmeasured until reconciled) is recorded as acceptable but not recommended;
  - the undefined late-gap coordinate, folded into ADR012-R7-01;
  - the incomplete REQ-27 and `ci:101` mirrors, folded into ADR012-R7-08. This changed the lead's closure rating of R6-01 and R6-08 from Fixed to Partly;
  - `as_of` missing from the dossier, which became ADR012-R7-05;
  - the Activation callout, which became ADR012-R7-09 (also found by the lead and the peer).
- **`experiment-peer`** raised 2 major and 7 minor findings. It rated R6-01/02 Fixed as prescribed and the rest Fixed. Adopted:
  - late-fact event loss and undefined triage-row merges, merged into ADR012-R7-02. The triage-row half and the gap-period wording were new to the lead;
  - event-to-attempt binding, which became ADR012-R7-03 at major. The lead found it independently; the peer's condition `started_at` ≥ `event_at`, and recording the event IDs, were adopted;
  - the checkability of cancellations and skips, which became ADR012-R7-04. It was new to the lead, and the lead re-verified the GitHub skip behaviour;
  - dossier identity and encoding, merged into ADR012-R7-05;
  - the stale `collectedThrough` and undefined C, merged into ADR012-R7-08 (the stale term was also found by the lead);
  - late records stamped after C not being disclosed, folded into ADR012-R7-01;
  - attempt-ID order, which became ADR012-R7-06 (also found by the lead);
  - catalogue missing from the pin tuple, which became ADR012-R7-07. It was new to the lead and corrects the council's own `round-6.md:282` premise;
  - the callout and the missing pin-mixture fixture, folded into ADR012-R7-09 and ADR012-R7-08.

  Not adopted: none.
- **Lead-only:**
  - the REQ-51 wiring caveat;
  - stamping of local imports and merges (ADR012-R7-01 item 3);
  - the "latency" ambiguity (ADR012-R7-03 item 3);
  - plan `:90` drift.

## Verdict

**ADR-012: 9 issues open (3 major, 6 minor). Not signed off.** Open: ADR012-R7-01, -02, -03, -04, -05, -06, -07, -08 and -09.

Round-6 closure:
- R6-03, -04, -05, -06, -09 and -10 are fixed.
- R6-02 and R6-07 are fixed as prescribed. Their remaining gaps are carried as ADR012-R7-01/-02 and ADR012-R7-05.
- R6-01 and R6-08 are partly fixed. Their missing REQ-27 and `ci:101` mirror elements are carried as ADR012-R7-08.

Earlier closures hold except R1-05, which ADR012-R7-01 reopens through an interaction.

No digests are signed in this round. The 13 digests above identify the bytes reviewed.

## Council memo

1. **Recommendation.** Do not sign off. Keep the narrowed, invariants-only ADR and apply the 9 fixes using the recommended options. Then submit round 8 on the revised signed set, with Main supplying fresh digests for all 13 files and the response.
   - Every R6 fix was applied faithfully at its main locations.
   - All three open major issues sit in one layer: how evidence becomes accounted. That covers which attempt accounts which event (R7-03), and how a fact that arrives after sealing is persisted, discharges completeness, is stamped and still feeds the cohort-wide vetoes (R7-01, R7-02).
   - To end the iteration, the writer should state the late-fact lifecycle once in the `ADR.md:19` paragraph and in REQ-78, as one rule: late facts are appended with a stamp, cover their points as gaps, discharge completeness and feed the vetoes. Other mirrors should reference that rule instead of restating it.

2. **Alternatives.**
   - **(A) Persist late facts with a stamp, as gaps for exposure and as evidence for the vetoes (recommended).** This keeps history invariant, allows default-query liveness, and makes the veto survive late arrival.
   - **(B) Keep the affected cohort unmeasured until late evidence is reconciled (challenger).** Simpler, but one listing lag suspends readiness availability, and "reconciled" needs its own definition. Acceptable, not recommended.
   - **(C) Bound C additionally by a provider-attested attempt-listing completeness time.** This reduces how often late attempts occur but cannot remove them, because imports and lag misestimates remain. It complements (A) and does not replace it.
   - **(D) Bind events to attempts at PR level only.** Simpler, but it lets the collector choose heads. Rejected.
   - **(E) Sign off and carry the items as implementation obligations.** This breaks design rule 1, and R7-01 to R7-03 change what counts as activation evidence. Rejected.

3. **Evidence.** The issue sections above cite exact lines. The key ones:
   - `ADR.md:16,17,19,20,21,28,30,36,44,48,56–62`;
   - `scout-design.html:449,589,600,612,615,618,707,1009,1010`, and base `:449,618`;
   - `cli-commands.md:38,126,128,130,132`; `policy-file.md:23,78,80,82`; `ci-integration.md:101,106`; plan `:61,90`; STATUS `:9,131`;
   - `round-2.md:267,272–273`; `round-6.md:170–172,282`;
   - `scout-guardian.yml:41–43`; `policy.go:416–424`;
   - [GitHub "Skipping workflow runs"](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/skip-workflow-runs).

4. **Risks.**
   - **Provider guarantees.** If neither the event listing nor the attempt listing can be shown complete within a justified lag, C advances slowly or not at all. This is conservative.
   - **Late-evidence vetoes** (R7-02) can hold a cohort until old late events are triaged. This is intended and visible.
   - **Binding to the head SHA** (R7-03) needs the provider's head SHA per event. If it is unavailable, the event is a gap, which is conservative.
   - **Manual cancels and skip markers** (R7-04) become gaps, so authors can force interval resets. That deflates readiness and never inflates it. It is disclosed.
   - **Adding `as_of` to the policy evidence** (R7-05) touches the documented activation record. It sits within ADR-012's declared REQ-78 scope, and the implementation is pending.
   - **Reversibility.** All changes are document-level and nothing is adopted. A superseding ADR can replace this one.

5. **Execution outline.**
   1. R7-01: in `ADR.md:19`, state that a late record discharges completeness; fix the effective time per class (`event_at` or `started_at`); make the trusted publishing CAS the only stamp source, with local pending late facts uncounted; add "late records after as_of, and pending late facts" to the live envelope at `:44`. Mirror this in REQ-48/78 and `cli:126,130`.
   2. R7-02: persist late facts with their links and events, flagged late; their events join the cohort-wide vetoes for `as_of` ≥ the stamp; the vetoes cover attempts whatever their gap or late status; merged or imported triage rows at or before C are re-stamped by a fresh CAS with original provenance, idempotent by action ID, and refused if unverifiable. Edit `ADR.md:19,20,36`, REQ-22/48/78 and `cli:126,128`.
   3. R7-03: add event IDs to the attempt record; account an event only by the same PR and head SHA with `started_at` ≥ `event_at`; name the event in non_default_base and not_assessed records; reword "latency" so it covers only events that will not be accounted. Edit `ADR.md:17,19`, REQ-27/78, `ci:101` and `cli:126`.
   4. R7-04: limit cancellations to verifiable supersession or system cancellation, and skips to those recorded from configuration; manual cancels and skip markers are assessed or gapped. Mirror the text, and add the GitHub skip documentation to R-070 if it is cited.
   5. R7-05: add `as_of`, the repository ID, the selector version, the path-filter population and a canonical encoding version to the dossier; add `as_of` to the REQ-78 evidence and to `policy-file.md:23`. Edit `ADR.md:44` and `cli:130`.
   6. R7-06: define the attempt-ID tuple and its numeric order everywhere "attempt ID" breaks a tie.
   7. R7-07: disclose the Scout/tool/catalogue/policy tuple in `ADR.md:44`, REQ-78 and `cli:128`.
   8. R7-08:
      - add to REQ-27 a reference to REQ-78's frontier rule and in-scope definition, and the in-scope reference to `ci:101`;
      - define C in `ADR.md:19` and REQ-78;
      - fix `policy-file.md:82` (C, both conditions);
      - update plan `:90`;
      - add pin-mixture disclosure to B4-S and `ci:106`;
      - add the wiring caveat to REQ-51.
   9. R7-09: update the Activation callout at `scout-design.html:449`.
   10. Add fixtures to `ADR.md:56–62`, plan B4-S `:61` and `ci:106`:
       - the default query recovers after a day-20 late attempt;
       - unequal coordinates for a late event and its attempt;
       - a late would_block event, and a late `not_warranted` import, block readiness after the stamp while `as_of` = C₁ is unchanged;
       - assessing only H2 leaves E1 unaccounted;
       - `[skip ci]` and manual cancel;
       - identical counts at different cutoffs give different digests;
       - attempt IDs 999 and 1000 order correctly.
   11. Fix the "predicates includes" grammar. Extend R-077's round list with `round-7.md`. Add R-070 sources only if they are cited.
   12. Write `round-7-response.md` mapping each issue. Main supplies the SHA-256 of all 13 signed files and of the response, then round 8 follows.
