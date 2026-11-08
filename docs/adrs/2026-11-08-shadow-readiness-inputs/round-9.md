# Round 9: ADR-012 (shadow-window readiness inputs)

Date: 2026-11-08. This round follows design rules 1–3 and the ledger rules in `CLAUDE.md`, and the ADR rules in `docs/adrs/README.md`. ADR-012 is an agent decision made under the owner's standing delegation (R-064). Where a fix involves a choice, the council names the option it recommends under that delegation.

The round reviews the bytes revised after round 8 (`round-8-response.md`). It has three parts:

1. Check every ADR012-R8-01..07 fix against its required fix.
2. Re-confirm that every earlier closure (R1–R7) still holds in the current bytes.
3. Look for new issues, numbered ADR012-R9-xx.

The council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review. It looked for false-ready paths, unreachable certification and loopholes in the scoped refusal, the history-import rules, the single visibility predicate, interval-ending point gaps, causal supersession and the job-start/deadline rule. It also checked the mirrors and the untouched requirements and accepted ADRs.
- **`experiment-peer` (Opus):** methodological review. It built concrete timelines to test whether each rule gives exactly one answer, and checked whether each acceptance fixture can expose the failure it targets.

The review was read-only. Nobody edited files, executed product code, ran model calls outside the council, or touched `.codegraph`. The lead read the required codegraph skill references (`readiness-and-bootstrap.md`, `query-selection.md`). Codegraph was not used, because the material under review is documents and the council had no shell. Searches used `read` and `grep`. The experiment-peer ran two web searches against GitHub documentation (repository Actions settings; REST events). Those searches informed only items in "Considered and not raised" and "Risks". The lead ran none.

**Main computed and verified the digests below with `shasum`. The council did not: the lead and both peers had read-only tools and no shell.** This round is not a sign-off, so the digests only identify the bytes reviewed. All paths are under `docs/`.

| File | SHA-256 (Main) |
|---|---|
| `adrs/2026-11-08-shadow-readiness-inputs/ADR.md` | `6473b3c7d1dbeee2ee7c604db06776e1da3229f9b08322c145b1a94eb1d3ce84` |
| `scout-design.html` | `7e7f0645154f4a0fd31b63ddb1933ef43ec428b9b65d93bc1a7edcfb54ce676c` |
| `adrs/README.md` | `cde415ff10b659a073d010f8a48c281d6f75a5a1cba466962d589e937b94587f` |
| `STATUS.md` | `02f351feaf593c3eb2359ac8f466b8363449388ac1b9e07f3d59652ed04bbf26` |
| `plans/2026-11-06-remaining-work.md` | `a7db7dd21fea9e78f2713399039ad9d690f7543f6a89caaa46c89536fe03d969` |
| `references/policy-file.md` | `94b722d09da00bbd5b5cbfed1cb72c9c562a326e2cf48b38c72dcb1ad7024e65` |
| `references/ci-integration.md` | `ef6cb765f5618eed6c5c38c6e2a100e0257cac03006076324900f7f1cd07ecb2` |
| `references/cli-commands.md` | `071d068c42a4a8a0c4435bb9acfcf6795cdebd604dc83bc15eecf0c6ef2e4a4a` |
| `research-ledger.md` | `7d726c054ebd3385734750c9772de9bc39e320275d3c3eaef309ce181a19b941` |
| `adrs/2026-10-19-v1-decisions/ADR.md` (ADR-000) | `1cd521fd96439d87275ba353f7b05451c3e0c4d556ee665529aab85970ed81cb` |
| `adrs/2026-10-23-decisions/ADR.md` (ADR-003) | `973d24c4cfdfdd98acb4070a7bd72ca0c0e1ddf303bd4ccec8f92fe1cf117cc5` |
| `adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) | `ee593b6241814c24035671d666566bc2b7c399135867226c26e9626f3936abc8` |
| `adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) | `80e3375f81656a8856cf662cbec99a36d8e00bc451ff3c3592925d2293f44d95` |

The following record is not signed. Main computed its SHA-256:

| File | SHA-256 (Main) |
|---|---|
| `round-8-response.md` | `a69d9d12c030da2851efb3cbb0f44f13a7818c2abfe248cfa30b5769898e57d5` |

README and the four historical ADR files have the same digests as in round 8.

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11`, branch `agent/policy-shadow-adrs`, based on origin/main `7bb73e52`. Every change is staged and uncommitted.
- Base bytes come from `/Users/adamjackson/Projects/work/scout-adr11b`.

**Files read.** Line numbers below refer to the raw worktree files.
- ADR-012 `ADR.md`, all lines:
  - `:3–6` status, revision and sources; `:10–12` decision and delegated narrowing;
  - `:16–22` items 1–7, where `:17` is the population/exclusion item and `:19` is "Clocks, frontier and late evidence";
  - `:26` triage authorization; `:28` triage rows and history import; `:30` ShadowReady;
  - `:34–40` persistence, manifest/merge, cutover and writer boundary;
  - `:44` CLI, live envelope and refusal scope; `:48–52` amendments, sources and route;
  - `:56–62` implementation acceptance invariants.
- `round-8.md` (header, issues `:134–290`, checks, peers, verdict and memo) and `round-8-response.md` (all lines).
- `round-7.md:376–392` (risks), `round-6.md:143–160,371,383`, `round-5.md:259` and `round-2-response.md:31–32`, all via grep context.
- `scout-design.html`, raw: Activation callout `:449`; REQ-22 `:589`; REQ-78 `:600`; REQ-27 `:612`; REQ-48 `:615`; REQ-51 `:618` (and base `:618`); grep context for REQ-10/26/39/160–168 `:610,611,704–714`; ADR-012 card `:1007–1010`.
- `cli-commands.md:126–132`, `policy-file.md:23,78–82` and `ci-integration.md:99–106`, all raw.
- Plan `:61,90`; STATUS `:7,9,121,131`; README `:12–30`; `research-ledger.md:73–76`.
- `.github/workflows/scout-guardian.yml:41–50`.
- The challenger compared REQ-160/162/163/168 and the four historical ADR bodies with base by reading the raw text.

## Issue summary

**5 distinct issues are open: 2 major and 3 minor.**

| ID | Severity | Summary |
|---|---|---|
| ADR012-R9-01 | major | Point-gap records have no lateness rule, and "At stamp≤as_of these gaps … take effect" conflicts with the normal-coordinate predicate. One reading hides a deadline gap at the frontier C it made possible, so a false ready can result. The other changes earlier historical dossiers |
| ADR012-R9-02 | major | The fast-forward test ignores destination rows published after C, and "lineage" is undefined. A diverged or fork-restored chain can pass as a fast-forward, which skips the import authorization check and the Undecided rule. Which events a later import sends back to Undecided is also unclear |
| ADR012-R9-03 | minor | The Undecided marker set when an imported row is refused has no record, stamp, visibility rule or export. It also does not trigger refusal, although an authorized imported row does |
| ADR012-R9-04 | minor | The second supersession test ("superseding run created before cancellation, no actor cancel recorded") is causal only when the pinned lane cancels in-progress runs |
| ADR012-R9-05 | minor | Four round-8 fixtures cannot detect their target failure: the default-version/representative split, resolution of a cross-lineage Undecided, the timing of deadline-late facts, and two refusal triggers with no fixture |

## Closure of round-8 issues

"Fixed" means the required document fix is present in the signed bytes at every named location. It does not mean the implementation is validated. "Fixed as prescribed" means the council's own prescription was applied but left a gap, which is carried as a new R9 issue.

| Issue | Status | Evidence |
|---|---|---|
| ADR012-R8-01 | **Fixed as prescribed** | The trigger list is exact: "repository late records stamped>C; authorized triage/history stamped>C for this exact cohort's veto events; and pending local late facts or imported history". Normal post-C facts, frontiers and metadata do not trigger it. Effect: "certification=refused, unmeasured/exit 2 regardless of the computed result". Live `current\|refused\|historical` sits outside the digest, and "Only current_window with certification=current can supply activation evidence" (`ADR.md:44`). Liveness and cadence are at `ADR.md:19`. Exit codes: evidence refusals are 2 and unsupported encoding is 3 (`:44`). Mirrors: REQ-78 `:600` (including "Activation evidence must come from current_window with certification=current"), `cli:130,132`, `policy-file:23,82`. Fixtures: `ADR.md:60`, plan `:61`, `ci:106`. The refusal-trigger gap for refused imports is ADR012-R9-03. |
| ADR012-R8-02 | **Fixed as prescribed** | `ADR.md:28` has the identical-prefix original-stamp fast-forward, action-ID deduplication, successive one-row-per-event publications in original (`recorded_at`, numeric action ID) order, original plus current-tip authorization, multi-lineage Undecided until a new destination verdict, and replay as a no-op. `:36` applies it to merge/import. Mirrors: REQ-48 `:615`, REQ-78, `cli:130`, `policy-file:82`. Fixtures: `ADR.md:60`, plan, ci. The prefix and lineage gap is ADR012-R9-02; the refused-row marker is ADR012-R9-03; the missing resolution assertion is ADR012-R9-05. |
| ADR012-R8-03 | **Fixed as prescribed** | "Late" is defined at `ADR.md:19` ("≤ destination C at publication"). One predicate names default version, in-scope attempts, representative, links, vetoes, gaps and Days origin (`:19`). `:20` says "throughout", and `:44` selects the latest "visible" attempt. Mirrors: REQ-78, `cli:126`, `policy-file:80`. Fixture: `ADR.md:58`, plan, ci. Lateness of point-gap records is undefined: ADR012-R9-01. The fixture weakness is ADR012-R9-05(a). |
| ADR012-R8-04 | **Fixed** | `ADR.md:19`: "Each point ends the containing interval: previous interval ends strictly before it, next starts strictly after it, and attempts at it belong to no interval". Points include "every accounted event_at". `:20`: "never puncture exposure". Mirror: REQ-78. Fixture (day 20 / day 35 / Days origin): `ADR.md:58`, plan, ci. |
| ADR012-R8-05 | **Fixed as prescribed** | `ADR.md:17` gives both causal branches. Mirrors: REQ-27 `:612`, REQ-78, `ci:101` (explicit), and `cli:126` by reference ("causal-supersession"). Fixture: `ADR.md:57`, plan, ci. The second branch is not causal without cancel-in-progress: ADR012-R9-04. |
| ADR012-R8-06 | **Fixed** | `ADR.md:58`: "Each test advances C≥all stamps with no refusal trigger and has an otherwise identical ready control". `:60` gives the same for the correction fixture. Mirrors: plan `:61`, `ci:106`. A refusal alone can no longer pass these fixtures. |
| ADR012-R8-07 | **Fixed as prescribed** | `ADR.md:19`: "started_at is provider job start, not run creation (approval waiting has no started attempt)" and "At the collector's chosen deadline a non-terminal attempt is point-gapped; later terminal facts follow the late lifecycle". Mirrors: REQ-78, `policy-file:82`, plus `ci:101` and `cli:126`. Fixture: `ADR.md:57`, plan, ci. The GitHub page is not cited, so R-070 correctly stays unchanged. When the new deadline gap becomes visible is ADR012-R9-01; the deadline fixture is ADR012-R9-05(c). |

## Re-confirmation of earlier closures (R1–R7)

| Closure | Holds? | Evidence |
|---|---|---|
| R1-04: membership keyed on `started_at`; elapsed-days formula | Yes | `ADR.md:19–21` |
| R1-05: a collection boundary can be recovered | Yes, restored by R8-07, apart from gap visibility | `ADR.md:19`. When the discharging gap becomes visible: ADR012-R9-01 |
| R2-05, R6-01/08, R7-04: population, supersession, exclusions | Yes, apart from the cancel-in-progress condition | `ADR.md:16–17`; REQ-27/78; `ci:101`. ADR012-R9-04 |
| R5-09: scanner-output failure is comparison-none | Yes | `ADR.md:17`: "Authenticated assessment input/output failure is comparison-none". The experiment-peer rated it broken; see "Considered and not raised" |
| R1-08/09/11/14, R2-10, R3-01, R4-07/10: CI/local/legacy separation, migration, v2 cutover, query-only v13 | Yes | `ADR.md:34–38,44`; REQ-10 `:610`; REQ-48 `:615`; `ci:104` |
| R3-02/07, R4-02/08/11: protected-source authorization, current maintainer authority, replay idempotence, writer isolation | Yes within one lineage | `ADR.md:26,28,40`; REQ-26 `:611`, REQ-51 `:618`, REQ-161 `:707`; `ci:102–103`. Imports: ADR012-R9-02, ADR012-R9-03 |
| R3-03, R4-06: guardian comparison and empty override inputs | Yes | `ADR.md:17–18`; REQ-27; `ci:101` |
| R2-11, R3-05, R4-03/04, R5-03/05/07, R6-02/07, R7-01: interval selection, bounds, triage cutoff, frontier-pinned minima, dossier invariance | Yes, apart from point-gap lateness | `ADR.md:19–20,28,44`; REQ-78; `cli:126–132`. ADR012-R9-01 |
| R5-06, R6-03: quiet-repository liveness and skewed-clock triage | **Yes** (was conditional in round 8) | The narrow refusal scope and cadence clause (`ADR.md:19,44`) resolve the R8-01 condition |
| R5-01, R6-10: wiring-ADR gate | Yes | `ADR.md:12,56`; REQ-51 `:618`; REQ-161; `ci:99`; plan `:61`; STATUS `:9,121`; `cli:134`; `policy-file:84` |
| R5-04, R6-04, R7-02: zero-not_warranted and zero-Undecided across gaps and late evidence, including provisional | Yes, apart from import edges | `ADR.md:20,28,30,58–59`; REQ-22 `:589`. ADR012-R9-02 |
| R5-08, R6-05: Days, representative origin and declared counting changes | Yes | `ADR.md:21,48,62`; card `:1009` |
| R5-10, R7-05/06/07: exact cohort, default version, dossier fields and encoding, attempt-ID order, pin tuples | Yes | `ADR.md:21,44,61–62`; REQ-78; `cli:126,130` |
| R7-03: event-to-attempt binding | Yes | `ADR.md:17,57`; REQ-27/78; `ci:101` |
| R7-08/09: mirror drift and Activation callout | Yes | REQ-27 `:612` normative reference; `policy-file:82` uses C; REQ-51 caveat; callout `:449` names inspected exposure and both cross-interval vetoes |

## Are the invariants testable and enough for REQ-22, REQ-27 and REQ-78?

| Requirement need | Testable and sufficient? | Where / gap |
|---|---|---|
| REQ-22: counting unit, deduplication, exact cohort, unknown ≤ 10 % | Yes | `ADR.md:20–22` |
| REQ-22: 14 days, inspected origin, `as_of` endpoint | Yes | `ADR.md:19,21,58,62` |
| REQ-22: zero `not_warranted` and zero Undecided across gaps, including late and imported evidence | **Partly** | Diverged-chain fast-forward (ADR012-R9-02); refused-row marker (ADR012-R9-03); no resolution assertion (ADR012-R9-05) |
| REQ-22: maintainer-only, append-only, idempotent triage | **Partly** | Fast-forward skips import authorization on a diverged chain (ADR012-R9-02) |
| REQ-27: population independent of content, author and approval | **Partly** | Supersession test without cancel-in-progress (ADR012-R9-04) |
| REQ-78: complete-interval accounting and recoverable frontier | **Partly** | Visibility of normal deadline/event gaps at the C they enable (ADR012-R9-01) |
| REQ-78: historical reproduction and `dossier_digest` | **Partly** | Point lateness (ADR012-R9-01); fast-forward ≠ fresh restore (ADR012-R9-02) |
| REQ-78: current certification and liveness | **Partly** | Refused-import marker missing from the trigger list (ADR012-R9-03) |
| Governance of deferred wiring | Yes | `ADR.md:12,56`; REQ-51/161 |

## Issues

### ADR012-R9-01: Point-gap records have no lateness rule, and their visibility conflicts with the single predicate (major)

- **Location:**
  - `ADR.md:19`: "A fact is late when its started_at (attempt) or event_at (event) is ≤ destination C at publication"
  - `ADR.md:19`: "normal facts require their provider coordinate≤as_of, late facts additionally require stamp≤as_of"
  - `ADR.md:19`: "points are started_at and every accounted event_at for an attempt, or event_at for an unaccounted event"
  - `ADR.md:19`: "At the collector's chosen deadline a non-terminal attempt is point-gapped"
  - `ADR.md:19`: "At stamp≤as_of these gaps and late events take effect; earlier dossiers stay identical"
  - Mirrors: REQ-78 `:600`; `cli:126`; `policy-file:82`; `ci:101`
- **Problem:** a point-gap record is neither an attempt nor an event, so the lateness definition does not cover it. "These gaps" grammatically covers every point gap, including ordinary deadline gaps and unaccounted-event gaps, and it conflicts with the coordinate rule for normal facts. The two readings give different gate results and different historical dossiers.
  - **Reading S: gaps take effect at their stamp** (challenger) [INFERENCE: constructed from the text, not executed].
    1. Parent C = day 30, and the containing interval otherwise qualifies.
    2. A job starts on day 31 and is still running at the collector's deadline.
    3. A publication whose run starts on day 32 records its point gap. Under the CAS formula its stamp is ≈ day 32. Its attested bound minus lag lets C advance to day 31.9.
    4. At `as_of` = C = day 31.9 the gap is invisible (stamp > `as_of`), yet it has discharged completeness. The non-terminal attempt is unpublished. The interval therefore runs through day 31 with that attempt silently missing.
    5. No refusal trigger applies: the gap is not a late record (31 > parent C). The current window is `current` and can be ready.
    - Normal gaps are always stamped after the C they enable (C ≤ run start − lag < stamp), so this is systematic, not an edge case. It applies equally to unaccounted-event gaps.
  - **Reading P: gaps take effect at their coordinate** (experiment-peer) [INFERENCE].
    1. Event E1 at 08:00 is accounted by attempt A1, and C reaches 12:50.
    2. Rerun A2 of the same head starts at 13:00 and also accounts E1.
    3. A2 hangs and is point-gapped at its deadline with points {13:00, 08:00}.
    4. The record is "normal" if its lateness follows A2 (13:00 > C). The 08:00 point is then visible by coordinate, and `--as-of 10:00` gains an interval boundary it did not have before. That violates "earlier dossiers stay identical" (R5-03, R6-07).
- **Severity:** the challenger rated this major and the experiment-peer rated its half minor. The lead adjudicates **major**: reading S is a false-ready path at the current window, and reading P breaks historical reproduction.
- **Required fix:**
  - **Lateness per point.** Each point of a gap record is late when it is ≤ destination C at the record's publication; otherwise it is normal.
  - **Visibility.** A normal point takes effect when point ≤ `as_of`. A late point takes effect only once its stamp ≤ `as_of` as well. Reword "At stamp≤as_of these gaps and late events take effect" to cover late points and late facts only.
  - It follows that C may pass a point only through a published gap in which that point is normal, so the gap is visible at every cutoff the advance permits. State this.
  - Mirror in REQ-78, `cli:126`, `policy-file:82` and `ci:101`.
  - **Fixtures:**
    - with parent C < point ≤ new C < stamp, the query at `as_of` = new C has its interval ending at the point;
    - a rerun's deadline gap whose accounted `event_at` is ≤ parent C leaves dossiers at cutoffs before its stamp identical, and the point takes effect once C ≥ the stamp.

### ADR012-R9-02: The fast-forward test ignores destination rows published after C, and "lineage" is undefined (major)

- **Location:**
  - `ADR.md:28`: "Same-lineage fast-forward means destination frontier history and rows stamped≤its C are an identical prefix of the import; adopt original trusted stamps, including in a sealed local store, exactly as fresh restore. Only rows absent from the destination lineage receive fresh CAS stamps."
  - `ADR.md:28`: "each later new cross-lineage import requires resolution again"
  - `ADR.md:36`: "Merge/import preserve immutable evidence"; `ADR.md:60`: "sealed local fast-forward equals fresh-restore dossiers"
  - "lineage" appears only as a row field (`ADR.md:28`) and is defined nowhere in `docs/` (experiment-peer grep).
  - Mirrors: REQ-48 `:615`; REQ-78 `:600`; `cli:130`; `policy-file:82`
- **Problem:**
  1. **A diverged chain passes the prefix test.** [INFERENCE] (lead, with both peers independently)
     1. Destination B seals C = day 30.
     2. On day 30 at 12:00, maintainer M2 marks event E `not_warranted` in B (row b1, stamped after C).
     3. Chain A was restored from B at C = day 30 and then continued on its own. On day 30 at 18:00, A records E `warranted` (row a1, A's stamp).
     4. A is imported into B. B's frontier history and its rows stamped ≤ day 30 are a prefix of A, so the literal text calls this a fast-forward.
     5. a1 keeps its 18:00 stamp, so latest-wins makes `warranted` current and zero-not_warranted passes once C reaches it.
     - B also keeps b1, so the result differs from a fresh restore of A, contradicting the rule's own claim. b1 also becomes visible under A's adopted later frontier, rewriting what A had sealed.
     - Read through the `lineage` field instead, the same import is cross-lineage and E is Undecided. Same state, different gate results.
  2. **Fast-forward skips import safeguards.** A fork that restored upstream state and then ran its own trusted writer produces exactly this shape (lead). Fast-forward adoption needs neither the original author's current-tip authorization nor the multi-lineage Undecided rule. Rows authorized only under the fork's maintainers can therefore become current in the destination. The same shape follows the "republish the last intact v2 commit and reapply subsequent state" recovery (`ADR.md:38`) if a stale copy continued.
  3. **A non-empty destination without C** passes vacuously: there is no sealed prefix to compare (challenger).
  4. **Which events a new import re-opens is unclear** (experiment-peer, lead). After E is resolved, an import that brings rows only for E′ re-opens E under a global reading but not under a per-event reading. And if a destination verdict lands between two successive publications of one import, "after that import" does not say whether it resolves E. If it does, the second publication's freshly stamped row becomes current, which is the round-8 resurrection path again.
- **Severity:** the experiment-peer rated this major and the challenger minor. The lead adjudicates **major**: whether the authorization and Undecided safeguards apply depends on this test, and a chain that passes the structural test reaches a false ready through rows that were never authorized in the destination.
- **Required fix (recommended option):**
  1. **Define lineage** as the trusted publication chain, identified at least by provider repository ID, state ref and genesis publication ID, and recorded on every published fact and row.
  2. **A fast-forward** requires the same lineage, and the destination's **entire** published history (every frontier record and every stamped fact and row, including those stamped after its C) must be an identical prefix of the import. A non-empty destination never qualifies vacuously. Everything else follows the cross-lineage rule.
  3. **Re-opening is per event.** An event becomes Undecided only when an import publication adds at least one new row for it, accepted or refused. It is resolved only by an authorized destination verdict stamped after the last such publication.
  4. Mirror in REQ-48/78, `cli:130` and `policy-file:82`.
  5. **Fixtures:**
     - a diverged-prefix import with a destination-only post-C correction is cross-lineage and Undecided;
     - a fork-restored chain is cross-lineage and needs current-tip authorization;
     - an import adding rows only for E′ leaves a resolved E resolved;
     - a destination verdict published between two publications of one import does not resolve E.
- **Acceptable alternative (challenger):** treat every import into a non-empty destination as cross-lineage, keeping fast-forward only for empty restores. Not recommended: it re-stamps genuine same-lineage restores into persistent local stores, the case round 8 fixed.

### ADR012-R9-03: The Undecided marker for a refused imported row has no defined record and does not trigger refusal (minor)

- **Location:**
  - `ADR.md:28`: "otherwise refuse/log the row and mark its event Undecided until a new authorized destination verdict"
  - `ADR.md:28`: "latest authorized visible row … is current"; "Both query modes filter stamps≤as_of"; "Post-C publications never alter the canonical dossier≤C"
  - `ADR.md:44`: the trigger list names "**authorized** triage/history stamped>C"
  - Mirrors: REQ-78; `cli:130`; `policy-file:82`; REQ-48 `:615`
- **Problem:**
  - A refused row is not authorized history, yet its refusal must change the event's state. The ADR defines no record carrying that change. Applying it immediately would alter dossiers at or before C. Discarding it bypasses the required reset. Only a stamped record is consistent with the other rules, and its visibility, export and idempotence are not stated (both peers).
  - **Inversion** [INFERENCE] (experiment-peer):
    1. C = 10:50, and E is `warranted`.
    2. At 10:55 an import brings an E row whose author was revoked. The row is refused, and E is marked Undecided.
    3. Nothing triggers refusal, so the current window is `current` and exit 0 ready.
    4. Had the author still been authorized, the multi-lineage row would have triggered refusal.
- **Severity:** the challenger rated this major and the experiment-peer minor. The lead adjudicates **minor**: the dossier certified at C reflects every authorized verdict, and no unauthorized input makes anything ready. The harm is an undefined record, which hurts reproduction, and a trigger list inconsistent with the R8-01 principle.
- **Required fix (recommended option):**
  - Persist an immutable import-refusal record: action ID, event ID, reason, lineage, CAS `recorded_at`. It is idempotent by action ID and exported/imported under the same rules.
  - Its Undecided effect applies only once its stamp ≤ `as_of`, and it is resolved only by a later authorized destination verdict.
  - Trigger-list wording: "triage/history records stamped>C that change the triage state of this exact cohort's veto events, whether authorized verdicts or import-refusal records".
  - Mirror in REQ-48/78, `cli:130` and `policy-file:82`.
  - **Fixture:** a revoked-author import on a cohort veto event leaves the C dossier identical. Current certification is refused until C ≥ the stamp, then computes exit 0 not-ready Undecided. A new authorized verdict restores ready, and a replay is a no-op.
- **Acceptable alternative (challenger):** quarantine the whole import under a durable certification refusal until an authorized resolution. Not recommended: it is coarser than an event-level record.

### ADR012-R9-04: The second supersession test is causal only when the lane cancels in-progress runs (minor)

- **Location:** `ADR.md:17` ("superseding run created before cancellation with no actor-initiated cancel recorded"); REQ-27 `:612`; REQ-78; `ci:101`. Code: `scout-guardian.yml:41–43` (`cancel-in-progress: true`).
- **Problem** [INFERENCE] (experiment-peer):
  1. The lane's concurrency setting is changed to `cancel-in-progress: false`. That is not a selector change (`ADR.md:16`), so no new interval starts.
  2. An author pushes H2 while H1 is running, then manually cancels H1 before it is assessed.
  3. If the provider records no canceller, the branch is satisfied and H1 is excluded on an author's choice.
  - Today's `cancel-in-progress: true` makes the branch causal. Without it the test fails open, which is the R8-05 harm.
- **Required fix:** the second branch applies only when the cancelled run's pinned lane configuration declares `cancel-in-progress` for that concurrency group; otherwise the cancellation is assessed or gapped. Mirror in REQ-27/78 and `ci:101`. **Fixture:** with `cancel-in-progress: false`, push H2 then manually cancel H1 with no canceller recorded; H1 is assessed or gapped.

### ADR012-R9-05: Four round-8 fixtures cannot detect their target failure (minor)

- **Location:** `ADR.md:57,58,60`; plan `:61`; `ci:106`.
- **Problem** (experiment-peer, checked by the lead):
  - **a. Default version versus representative.** `:58` uses one late rerun "with a different pin/inspection status". A different-pin rerun is outside the v1 cohort whatever its visibility. An implementation that leaks late facts only into representative selection therefore passes.
  - **b. Cross-lineage resolution.** `:60` asserts Undecided "until a new destination verdict" but never that the new verdict resolves it. An implementation stuck on Undecided forever passes.
  - **c. Deadline lateness.** `:57` "its later terminal facts are late" holds only if C passed `started_at` before they were published. As written, the fixture either fails a correct implementation or tests nothing.
  - **d. Refusal triggers.** No fixture covers pending local late facts or pending imported history refusing and then clearing after publication and a C advance. No fixture covers a late record of an unrelated check refusing this cohort; only the opposite triage case is tested.
- **Required fix:** in `ADR.md:57–60`, plan B4-S and `ci:106`:
  - **a.** Use two late reruns: one with the same pin and a different inspection status, which must not change the representative at the historical cutoff, and one with a different pin, which must not change the default version.
  - **b.** Assert that a new authorized destination verdict returns the cohort to ready.
  - **c.** Advance C past `started_at` before publishing the terminal facts, and assert they are late and stamp-gated.
  - **d.** Add both refusal-trigger fixtures.

## Checks that found no issue

- **Refusal liveness.** A triage that clears the last Undecided event is refused until C ≥ its stamp, then is current and ready. A frontier-only reconcile computes normally. A late record of an unrelated check refuses every cohort, which is justified because late points end every cohort's intervals, and it clears at the next C advance past its stamp (both peers).
- **Exit codes.** `ADR.md:44` and `cli:130–132` agree: computed results exit 0; scoped refusal overrides computed 0 to exit 2; evidence refusals and unmeasured exit 2; invalid inputs, unsupported schema and unsupported encoding exit 3. A visible cross-lineage Undecided computes not-ready with exit 0 (challenger).
- **Historical bypass.** `--as-of C` is `historical` and never activation evidence (`ADR.md:44`; REQ-78; `cli:130,132`; `policy-file:23,82`).
- **Late attempt at day 20 found after C = day 30.** The day-30 digest stays the same. Once `as_of` ≥ the stamp, the interval and Days origin start after day 20 (experiment-peer).
- **Day-25/day-32 cross-lineage verdicts in either order.** E is Undecided from the import stamp until a new verdict, and earlier cutoffs are unchanged, apart from the R9-02 edges.
- **Mirrors.** REQ-22 `:589` keeps the cohort-wide veto under REQ-78 visibility. `cli:128` and `policy-file:80` keep both vetoes, including for provisional readiness. STATUS `:121`, plan `:90` and card `:1009` summarize the round-8 consolidations accurately.
- **Unchanged requirements and accepted ADRs.** REQ-160/162/163/168 are unchanged from base. The four historical ADR digests equal round 8's, and only their status lines differ from base (challenger, raw-text comparison).
- **Ledger.** R-077 (`research-ledger.md:76`) links `round-8.md`. R-064, R-069 and R-070 are unchanged. No R-ID or ADR ID is reused, and STATUS `:131` names next free R-078.

## Considered and not raised

- **"R5-09 is broken" and author- or outsider-caused interval resets** (experiment-peer finding 1). R5-09's sentence "Gaps arise only when the trusted collector cannot enumerate …" (`round-5.md:259`) was deliberately broadened by R6-01 (`round-6.md:152–153`): anything attributable to the collector, including an approval wait the collector does not need, is a gap, and the collector should assess forks whatever their approval state. Round 7 explicitly accepted that authors can force interval resets ("deflates readiness and never inflates it", `round-7.md:387`). The new deadline gap applies only to attempts still non-terminal at the deadline. Terminal timeouts stay comparison-none (`ADR.md:17`), so "actual terminal unknown never resets a window" (`round-2-response.md:32`) still holds. This is a fail-closed default. An implementation that assesses forks and sets its deadline at or above the lane's job timeout keeps readiness reachable for a compliant repository. It is recorded under Risks.
- **Approval waiting "cannot hold C" (`ADR.md:57`) versus "cannot indefinitely hold C" (`ADR.md:48`)** (lead). Gapping at once versus holding until the deadline are both fail-closed, and neither makes readiness unreachable beyond the previous item.
- **Certification cannot be audited from the activation record** (experiment-peer finding 5). Round 8 chose option A, a live status plus an activation rule, deliberately. Activation evidence is maintainer-attested under REQ-22/78. Published trigger records stamped in (C, activation] can be audited from `scout/state`. Adding a state-commit field changes the policy evidence schema without a defect that requires it.
- **REQ-51 "byte-identically"** (challenger R9-04). The clause is unchanged from base (`:618`) and concerns importing an export into a fresh store. `ADR.md:36` keeps that ("A destination without C retains genuine byte-identical v2 round trips"). A cross-lineage import into an existing ledger is a merge, not a reproduction.
- **Provider timestamps at 1-second resolution** may put same-second attempts on a point outside every interval. This fails closed and is rare.
- **Pending local facts the trusted ledger never publishes** keep that one local reader refused. A fresh restore clears it, and CI readers are unaffected.
- **`policy-file:80` "predicate above"** points at nothing above it, but the same line resolves it. This is wording.

## Peers and where they changed the outcome

- **`innovation-challenger`** raised 2 major and 2 minor findings. It rated R8-04/05/06 fixed and R8-01/02/03/07 fixed as prescribed.
  - The visibility conflict for normal gaps became ADR012-R9-01 and **changed the lead's view**: the lead had read "these gaps" as late gaps only.
  - The refused-import reset lifecycle was merged into ADR012-R9-03 and down-rated to minor.
  - The fast-forward prefix gap and the vacuous no-C case were merged into ADR012-R9-02. The lead had found the prefix gap independently.
  - Not adopted: REQ-51 byte identity (see above).
- **`experiment-peer`** raised 2 major and 5 minor findings. It rated R8-04/06 fixed and the rest fixed as prescribed. Adopted:
  - per-point lateness, merged into ADR012-R9-01. This was new to the lead;
  - the undefined lineage and the re-opening scope, merged into ADR012-R9-02 at major;
  - the refused-import inversion, merged into ADR012-R9-03;
  - supersession without cancel-in-progress, which became ADR012-R9-04 and was new to the lead;
  - fixture strength, which became ADR012-R9-05 and was new to the lead.

  Not adopted: R5-09 broken, and certification auditability (see above).
- **Lead-only:**
  - the fork-restored chain as an authorization bypass (ADR012-R9-02 item 2);
  - a destination verdict published between two publications of one import (ADR012-R9-02 item 4);
  - the approval-wait wording check, not raised;
  - the severity calls on ADR012-R9-01, -02 and -03.

## Verdict

**ADR-012: 5 issues open (2 major, 3 minor). Not signed off.** Open: ADR012-R9-01, -02, -03, -04 and -05.

Round-8 closure:
- R8-04 and R8-06 are fixed.
- R8-01, -02, -03, -05 and -07 are fixed as prescribed. Their remaining gaps are carried as ADR012-R9-01 to -05.

Earlier closures hold, with these qualifications:
- R1-05 and the dossier-invariance group hold apart from ADR012-R9-01.
- The authorization and veto groups hold apart from ADR012-R9-02/03.
- The population group holds apart from ADR012-R9-04.
- R5-06 and R6-03 now hold without condition.

No digests are signed in this round. The 13 digests above identify the bytes reviewed.

## Council memo

1. **Recommendation.** Do not sign off. Apply the 5 fixes using the recommended options, then submit round 10 with fresh digests from Main for all 13 signed files and the response.
   - The round-8 consolidations worked. Scoped refusal, live certification, the single predicate and interval-ending gaps are sound. What remains is at two seams: when gap records become visible, and import identity.

2. **Alternatives.**
   - **(A, recommended)** Per-point lateness for gap records; lineage defined as the trusted chain, with a whole-history prefix and per-event re-opening; a stamped import-refusal record that triggers refusal; a cancel-in-progress condition on supersession; strengthened fixtures.
   - **(B)** For R9-01, broaden refusal to every normal gap publication. Rejected: it recreates the R8-01 liveness problem.
   - **(C)** For R9-02, treat every import into a non-empty destination as cross-lineage. Acceptable but coarser: it re-stamps genuine same-lineage local restores.
   - **(D)** For R9-03, quarantine the whole import. Acceptable but coarser.
   - **(E)** Sign off and carry these as implementation obligations. Rejected: design rule 1, and R9-01/02 change what counts as activation evidence.

3. **Evidence.**
   - `ADR.md:16,17,19,20,28,36,38,44,48,57–60`;
   - `scout-design.html:589,600,612,615,618` (and base `:618`);
   - `cli-commands.md:126,130,132`; `policy-file.md:23,80,82`; `ci-integration.md:101,106`; plan `:61`;
   - `round-5.md:259`; `round-6.md:152–153`; `round-7.md:387`; `round-2-response.md:32`;
   - `scout-guardian.yml:41–43,50`.

4. **Risks.**
   - **Author- and outsider-caused gaps.** If the chosen wiring cannot assess forks before approval, or sets a deadline below the lane's job timeout, first-time-contributor PRs or slow scans can keep intervals short. This fails closed and was accepted in rounds 6–7; the implementation should avoid it.
   - **Provider capability** [INFERENCE]: attested complete event enumeration, job start times and the cancellation cause or canceller may need per-PR timeline APIs and an hours-scale lag. If they are unavailable, C stalls or events are gapped, which is conservative.
   - **Lineage identity** depends on the trusted writer recording a chain identity at genesis. A ledger without one must be treated as cross-lineage.
   - **Reversibility.** All changes are document-level and nothing is adopted. Schema 13 still ships, and a superseding ADR can replace this one.

5. **Execution outline.**
   1. R9-01: in `ADR.md:19`, define per-point lateness, coordinate visibility for normal points and stamp visibility for late points, and that C advances only through normal points. Reword "these gaps". Mirror in REQ-78, `cli:126`, `policy-file:82` and `ci:101`. Add both fixtures.
   2. R9-02: in `ADR.md:28,36`, define lineage, require a whole-published-history prefix and no vacuous non-empty destination, and make re-opening per event, resolved after the last publication. Mirror in REQ-48/78, `cli:130` and `policy-file:82`. Add the four fixtures.
   3. R9-03: in `ADR.md:28,44`, add the import-refusal record and extend the trigger list. Mirror in REQ-48/78, `cli:130` and `policy-file:82`. Add the revoked-author fixture.
   4. R9-04: in `ADR.md:17`, condition the second branch on cancel-in-progress. Mirror in REQ-27/78 and `ci:101`. Add the fixture.
   5. R9-05: revise the fixtures in `ADR.md:57–60`, plan B4-S and `ci:106` (items a–d).
   6. Update the card `:1009`, STATUS `:121` and plan `:90` only if their summaries become inaccurate. Extend R-077's list of rounds with `round-9.md`.
   7. Write `round-9-response.md` mapping each issue. Main supplies the SHA-256 of all 13 signed files and of the response, then round 10 follows.
