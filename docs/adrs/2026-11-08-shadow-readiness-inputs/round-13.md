# Round 13: ADR-012 (shadow-window readiness inputs)

Date: 2026-11-08. This round applies design rules 1–3 and the ledger rules in `CLAUDE.md`, plus the ADR rules in `docs/adrs/README.md`. ADR-012 is an agent decision made under the owner's standing delegation (R-064). Where a fix involves a choice, the council names the option it recommends under that delegation.

The round reviews the bytes revised after round 12 (`round-12-response.md`). It has three parts:

1. Check the ADR012-R12-01 fix against the required fix in `round-12.md`.
2. Confirm that every earlier closure (R1–R11) still holds in the current bytes.
3. Look for new issues, numbered ADR012-R13-xx.

The council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review. It looked for false exclusions, readiness that a compliant repository could never reach, mirror drift, any remaining "created before cancellation" wording, ledger ID reuse and changes to historical ADR bodies. It also examined the new "started" lower bound against the ADR's clock definitions, runs cancelled while still queued, and rerun attempts.
- **`experiment-peer` (Opus):** methodological review. It built nine timelines under the two-bound condition. It then looked for faulty implementations that pass every fixture, tested whether each new fixture fails its target implementation, and checked the Amendments provenance against `round-12-response.md` and R-077.

The review was read-only. Nobody edited files, ran product code, made model calls outside the council, or touched `.codegraph`.

- The lead read the required codegraph skill references (`readiness-and-bootstrap.md`, `query-selection.md`). Codegraph was not used, because the material under review is documents and the council had no shell. Searches used `read`, `grep` and `glob`.
- The lead fetched GitHub's [re-running workflows and jobs](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/re-run-workflows-and-jobs) page this round. It says: "The workflow will also use the same `GITHUB_SHA` (commit SHA) and `GITHUB_REF` (git ref) of the original event that triggered the workflow run." R-070 already links this page.
- The challenger cited the [workflow-runs REST page](https://docs.github.com/en/rest/actions/workflow-runs#get-a-workflow-run-attempt), which documents the `created_at`, `run_started_at` and `run_attempt` fields and the attempt endpoint.

**Main computed and verified the digests below with `shasum`. The council did not, because the lead and both peers had read-only tools and no shell.** This round is not a sign-off, so the digests only identify the bytes reviewed. All paths are under `docs/`.

| File | SHA-256 (Main) |
|---|---|
| `adrs/2026-11-08-shadow-readiness-inputs/ADR.md` | `02fa2afb17279802ec82337a645b55a75851dee65c2e36ad55f187c4ccd2758d` |
| `scout-design.html` | `8c7c564990b5894404ab39553e455f8b7869be14b127cca0dd7d53afa7d6cb2b` |
| `adrs/README.md` | `cde415ff10b659a073d010f8a48c281d6f75a5a1cba466962d589e937b94587f` |
| `STATUS.md` | `17e6e4bf7f131111897600dcdbd17e5d51c13eace4f38b65cf7622357d3d563c` |
| `plans/2026-11-06-remaining-work.md` | `3c7d6b2c7e56a5f9fa4f7cae5e8bc07ab5f205da2e870121c3c0ec2b326ab711` |
| `references/policy-file.md` | `7b747b02563fb3fdfa41914b99b9a7fa25a12546ad23fe98e922b99cccc80727` |
| `references/ci-integration.md` | `6e8d7ea1bea1e23ce294c74a63c97bfdd8ec60801fc4002e17ea89acce249dd5` |
| `references/cli-commands.md` | `722bbc7be5228c9e96b68e40d2715de68629e1b70e3920ab399d69847e3d2926` |
| `research-ledger.md` | `fddbe08fcf124b53a938a6875c7798a73ef4be08af4fa5fda821c3ea05a54a3d` |
| `adrs/2026-10-19-v1-decisions/ADR.md` (ADR-000) | `1cd521fd96439d87275ba353f7b05451c3e0c4d556ee665529aab85970ed81cb` |
| `adrs/2026-10-23-decisions/ADR.md` (ADR-003) | `973d24c4cfdfdd98acb4070a7bd72ca0c0e1ddf303bd4ccec8f92fe1cf117cc5` |
| `adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) | `ee593b6241814c24035671d666566bc2b7c399135867226c26e9626f3936abc8` |
| `adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) | `80e3375f81656a8856cf662cbec99a36d8e00bc451ff3c3592925d2293f44d95` |

Main also computed the SHA-256 of one unsigned record:

| File | SHA-256 (Main) |
|---|---|
| `round-12-response.md` | `b0b2fa5130152aed1fce1db9cec621c7c1cc1d8f60a97e9f508687ecaf95e3f1` |

README, STATUS, the policy-file reference and the four historical ADR files have the same digests as in round 12. The changed files are ADR.md, the design, the plan, the ci-integration and cli-commands references, and the ledger.

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11`, branch `agent/policy-shadow-adrs`, based on origin/main `7bb73e52`. Every change is staged and uncommitted.
- Base bytes come from `/Users/adamjackson/Projects/work/scout-adr11b`.

**Files read.** Line numbers below refer to the raw worktree files.
- ADR-012 `ADR.md`, all lines. The main ones are:
  - `:17`, population and supersession;
  - `:19`, clocks: "started_at is provider job start, not run creation";
  - `:21`, attempt ID (provider repository ID, run ID, run attempt, job ID);
  - `:48`, Amendments;
  - `:56–62`, acceptance invariants.
- `round-12.md` and `round-12-response.md`, all lines.
- `scout-design.html`, raw: REQ-78 `:600`, REQ-27 `:612`, ADR-012 card `:1007–1010`.
- `ci-integration.md:95–108` and `cli-commands.md:124–136`, raw.
- Plan `:61`, raw; `research-ledger.md:75–76`, raw; STATUS `:131`.
- `.github/workflows/scout-guardian.yml:1–43`. It is a reusable `workflow_call` workflow. Its workflow-level concurrency is at `:41–43`: `group: scout-guardian-${{ github.repository }}-${{ github.ref }}`, `cancel-in-progress: true`.
- A docs-wide grep, excluding the historical round files, for "created before cancellation" and "created after the cancelled run".
- A grep of every round and response in this folder for any earlier discussion of a rerun as a canceller or superseder. The only rerun match is round 2's "partial rerun" in the terminal matrix (`round-2-response.md:31`).
- The challenger compared the four historical ADRs with base, raw.

## Issue summary

**2 issues are open, both minor. None is major.**

| ID | Severity | Summary |
|---|---|---|
| ADR012-R13-01 | minor | The round-12 clause in `ADR.md:48` calls the adopted lower bound ("after cancelled-run start") "the recommended" window. Round 12 recommended "created after the cancelled run", a creation bound. The stricter bound is acceptable; the signed record of the delegated decision is not accurate |
| ADR012-R13-02 | minor | A rerun of an older guardian run reuses the original `GITHUB_REF` and so joins the same ref-keyed concurrency group, where it can cancel the final head's in-progress run. The rule's run-level "created" gives the right answer. No fixture fails an implementation that dates the superseding run by its rerun attempt, and such an implementation excludes the final head |

## Closure of round-12 issue

"Fixed" means the required document fix is present in the reviewed bytes at every named location. It does not mean the implementation is validated.

| Issue | Status | Evidence |
|---|---|---|
| ADR012-R12-01 | **Fixed in substance, with a deviation; the provenance is reopened as ADR012-R13-01** | **Rule.** All five locations (`ADR.md:17`, REQ-78 `:600`, REQ-27 `:612`, `ci:101`, `cli:126`) read: "created after the cancelled run started and before its cancellation". Each adds: "If no qualifying run can be identified within both bounds, assess or gap the head, never exclude it; missing proof cannot certify readiness." "Created before cancellation" no longer appears anywhere in the docs outside the round files.<br>The required text was "created after the cancelled run". The adopted lower bound, the cancelled run's start, is no earlier than its creation, so the window is never wider than the required one. A cancelled run that never started has no start time, and the fallback then assesses or gaps its head. Both peers and the lead found no false-exclusion path from this change [INFERENCE]. Readiness stays reachable through collector reassessment of the cancelled head and later complete intervals (`ADR.md:16–21`), so under the calibration rules the stricter bound is not a defect.<br>**Fixtures.** The same paragraph appears at `ADR.md:57`, plan `:61` and `ci:106`: "Supersession-window fixtures (all other exclusion conditions satisfied): (1) … (2) … never accept older R1 as its superseder. (3) … keep every existing fixture." All earlier fixtures are kept. An implementation without the lower bound fails fixtures 2 and 3, one per evidence path. One without the upper bound fails fixture 1.<br>**Records.** R-077 (`research-ledger.md:76`) is titled "council rounds 2–12 findings" and links `round-12.md`. R-070 is unchanged, and no new ledger ID was added. `ADR.md:48` records the round-12 decision, but misattributes the lower bound to round 12's recommendation (ADR012-R13-01). |

## Re-confirmation of earlier closures (R1–R11)

| Closure | Holds? | Evidence |
|---|---|---|
| R1-04: started_at membership; elapsed-days formula | Yes | `ADR.md:19–21` |
| R1-05, R7-01, R8-07, R10-03: recoverable frontier, deadline point gaps, CAS-parent anchor, (parent C, new C] range | Yes | `ADR.md:19`; REQ-78 `:600`; `ci:101`; `cli:126`; `policy-file:82` (digest unchanged) |
| R2-05, R6-01/08, R7-04, R8-05, R9-04, R10-04, R11-01, R12-01: population, exclusions, causal supersession, window | Yes, apart from ADR012-R13-02 | `ADR.md:16–17`; REQ-27/78; `ci:101`; `cli:126` |
| R11-02: round-9 provenance clause | Yes | `ADR.md:48`: "conditioned on the cancelled run’s pinned cancel-in-progress (replaced in round 10)" |
| R5-09: scanner-output failure is comparison-none | Yes | `ADR.md:17` |
| R1-08/09/11/14, R2-10, R3-01, R4-07/10: CI/local/legacy separation, migration, v2 cutover, query-only v13 | Yes | `ADR.md:34–38,44` |
| R3-02/07, R4-02/08/11, R8-02, R9-02/03: authorization, current-tip authority, replay, writer isolation, lineage, import refusal | Yes | `ADR.md:26,28,36,40`; `cli:130` |
| R3-03, R4-06: guardian comparison, empty override inputs | Yes | `ADR.md:17–18`; REQ-27 |
| R2-11, R3-05, R4-03/04, R5-03/05/07, R6-02/07, R7-01, R8-03, R9-01: interval selection, cutoffs, frontier minima, dossier invariance, single predicate | Yes | `ADR.md:19–20,28,44`; REQ-78; `cli:126–132` |
| R8-04, R10-02: exclusive point boundaries; strict and equal boundary fixtures | Yes | `ADR.md:19,58`; plan `:61`; `ci:106` |
| R10-01: independent day-30 rerun assertions; both halves of deadline visibility | Yes | `ADR.md:57–58`; plan `:61`; `ci:106` |
| R5-06, R6-03, R8-01: liveness, skewed-clock triage, scoped refusal | Yes | `ADR.md:19,44,60` |
| R5-01, R6-10: wiring-ADR gate | Yes | `ADR.md:12,56`; `ci:99`; plan `:61`; `cli:134` |
| R5-04, R6-04, R7-02, R8-06, R9-05: cohort-wide zeroes including provisional; fixtures cannot pass by refusal alone | Yes | `ADR.md:20,30,57–60` |
| R5-08, R6-05: Days, representative origin | Yes | `ADR.md:21,48,62`; card `:1009` |
| R5-10, R7-05/06/07: exact cohort, default version, dossier encoding, attempt-ID order, pin tuples | Yes | `ADR.md:21,44,61–62`; REQ-78 |
| R7-03: event-to-attempt binding | Yes | `ADR.md:17`; REQ-27/78 |
| R7-08/09: mirror drift; Activation callout | Yes | The five rule locations carry identical window and fallback text. The three fixture locations carry the identical paragraph. Card `:1009` has no window wording to drift |
| Historical ADRs and untouched requirements | Yes | ADR-000/003/007/008, README and STATUS digests are unchanged since round 12. The challenger's raw comparison with base found only the `:3` status lines differ |
| Ledger | Yes | R-077 is extended with round 12. STATUS `:131` still names next free R-078. No R-ID or ADR ID is reused |

## Are the invariants testable and enough for REQ-22, REQ-27 and REQ-78?

| Requirement need | Testable and sufficient? | Where / gap |
|---|---|---|
| REQ-22: counting unit, deduplication, exact cohort, unknown ≤ 10 % | Yes | `ADR.md:20–22` |
| REQ-22: 14 days, inspected origin, `as_of` endpoint | Yes | `ADR.md:19,21,58,62` |
| REQ-22: zero `not_warranted` and zero Undecided across gaps, late and imported evidence | Yes | `ADR.md:20,57–60` |
| REQ-22: maintainer-only, append-only, idempotent triage | Yes | `ADR.md:26,28,60` |
| REQ-27: population independent of content, author and approval | **Partly** | The rule text is sound; the fixtures miss the rerun-canceller case (ADR012-R13-02) |
| REQ-78: complete-interval accounting, recoverable frontier | Yes | `ADR.md:19,58` |
| REQ-78: historical reproduction, `dossier_digest` | Yes | `ADR.md:19,44,58,61` |
| REQ-78: current certification and liveness | Yes | `ADR.md:19,44,58,60` |
| Governance of deferred wiring | Yes, apart from the provenance error (ADR012-R13-01) | `ADR.md:12,48,56` |

## Issues

### ADR012-R13-01: The Amendments record misattributes the adopted lower bound to round 12's recommendation (minor)

- **Location:** `ADR.md:48`, which reads: "Round-12 delegated decision: adopt the recommended two-bound superseding-run window, explicitly after cancelled-run start and before cancellation; absent qualifying proof assesses/gaps the head rather than excluding it."
- **Problem:**
  - Round 12's required fix (recommended option) was "created after the cancelled run and before its cancellation", a creation lower bound. The bytes adopt the cancelled run's *start*, which is stricter. As a result, a cancelled run that never started, such as one cancelled while queued during rapid pushes, can never be superseded, and its head is assessed or gapped.
  - That substitution is acceptable on its merits (see the round-12 closure). But the signed record presents it as the recommended option, and nothing else in the signed set records that it departs from the recommendation. `round-12-response.md:3,7` describes it the same way, but that record is historical and stays as it is.
  - R-064 makes the owner's later review rely on accurate records of delegated decisions. This is the same class of defect as ADR012-R11-02.
- **Severity:** minor. It does not change any outcome, but it is a factual error in a signed file.
- **Required fix (recommended option: keep the stricter bound and record it accurately):** in `ADR.md:48`, replace the round-12 clause with:

  > "Round-12 delegated decision: adopt a two-bound superseding-run window ending at the cancelled attempt's cancellation; its lower bound is the cancelled attempt's job start, deliberately stricter than round 12's recommended run-creation bound, so a cancelled run that never started is assessed/gapped; absent qualifying proof assesses/gaps the head rather than excluding it."

  - If ADR012-R13-02's recommended rule wording is adopted, this clause matches it exactly.
  - Alternative, not recommended: change the five rule locations back to round 12's creation bound. That restores exclusion for queued-run supersession. It also reopens a choice that fails closed and that the calibration rules accept.

### ADR012-R13-02: A rerun of an older run can be accepted as the superseding run, and no fixture fails that implementation (minor)

- **Location:**
  - The window clause at `ADR.md:17`, REQ-78 `:600`, REQ-27 `:612`, `ci:101` and `cli:126`.
  - The fixtures at `ADR.md:57`, plan `:61` and `ci:106`.
- **New evidence:**
  - GitHub's re-run page: reruns "use the same `GITHUB_SHA` (commit SHA) and `GITHUB_REF` (git ref) of the original event" (fetched by the lead this round).
  - The guardian concurrency group is keyed by `github.ref`, with `cancel-in-progress: true` (`scout-guardian.yml:41–43`). A rerun of R1 therefore enters the same group as the newer R2 and cancels R2 if R2 is in progress. That follows from the documented rule that a run entering a group cancels the in-progress run [INFERENCE: the [control-workflow-concurrency](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency) semantics applied to reruns].
  - The rerun executes R1's original, unedited configuration, so it passes every configuration check.
  - The ADR requires per-attempt records keyed by (repository, run ID, run attempt, job ID) (`ADR.md:21`). The REST API exposes per-attempt `run_started_at` and an attempt endpoint (challenger). Dating a "run" by its attempt record is therefore a natural implementation [INFERENCE].
  - No earlier round considered a rerun as a canceller or superseder; the folder grep found only round 2's "partial rerun".
- **Problem** (experiment-peer, verified by the lead) [INFERENCE: constructed timeline, not executed]:
  1. Pinned `true`, unedited configuration.
  2. R1 on H1 is terminal: it failed, or was legitimately cancelled by R2.
  3. R2 on H2 starts.
  4. A maintainer reruns R1. Attempt 2 joins the group and cancels R2 before assessment. A provider concurrency cause is recorded, and no actor-initiated cancel is recorded. There is no H3.

  - Read correctly, the rule dates the run by its creation. R1 was created before R2 started, so it falls outside the window and H2 is assessed or gapped.
  - An implementation that dates the candidate by its rerun attempt's creation or start finds a qualifying "run" inside the window and excludes the final head H2. That is the false exclusion targeted by R11-01 and R12-01, reached by a new route.
  - All three window fixtures have no rerun, so this implementation passes every fixture.
  - The trigger is ordinary maintainer behaviour: re-running an old failed or cancelled guardian run while a newer run is in progress.
- **Severity:** minor, the same as R12-01. The rule's wording already gives the right answer; the defect is that the binding fixtures cannot tell a correct implementation from a naive one. The fix only fails closed, and every legitimate supersession comes from a newly created run, so a compliant repository's readiness is unaffected.
- **Required fix (recommended option):**
  - **Rule.** At the five rule locations, replace "created after the cancelled run started and before its cancellation" with:

    > "created (its first attempt; a rerun attempt of an earlier-created run never qualifies) after the cancelled attempt's job started and before that attempt's cancellation"

    Keep the fallback sentence and every other condition unchanged. This states the stricter start bound that was actually adopted, fixes both coordinates in the ADR's own attempt and clock vocabulary (`ADR.md:19,21`), and excludes reruns.
  - **Fixture.** Add fixture (4) at `ADR.md:57`, plan `:61` and `ci:106`, keeping all existing fixtures:

    > "(4) Pinned=true, unedited configuration; R1 on H1 is terminal; R2 on H2 has started; a maintainer reruns R1 and the rerun attempt cancels R2 before assessment; no H3. Once with provider concurrency cause and once with no actor recorded/no provider cause: assess/gap H2, never accept R1's rerun attempt as its superseder."
  - **Records:**
    - add the round-13 delegated decision to `ADR.md:48`;
    - add the re-run fact to R-070's summary ("reruns reuse the original `GITHUB_SHA`/`GITHUB_REF`, so a rerun of an older run joins the same ref-keyed group (round 13)"). The source is already linked in R-070, so no new ID is needed;
    - extend R-077 with `round-13.md`.

## Checks that found no issue

- **The stricter "started" lower bound** (challenger, experiment-peer, lead). It is never earlier than the creation bound, so the window only narrows [INFERENCE]. A never-started cancelled run falls to the fallback. Reading "run start" as GitHub's `run_started_at` gives a window no wider than the required one. Neither peer nor the lead found a false exclusion apart from the superseding-run dating in ADR012-R13-02.
- **Timelines under the current wording** (experiment-peer) [INFERENCE]:
  - fixtures 1–3 and a manual cancel followed by a later push give correct outcomes;
  - rapid-push supersession of a queued R1 is assessed or gapped, which fails closed;
  - an older run cancelling a newer one out of order is assessed or gapped, which fails closed;
  - a PR-added workflow cancelling the final head's run, queued or started, is assessed or gapped;
  - a rerun attempt of R1 cancelled by a newer R3 excludes H1, which is legitimate because H3 supersedes it.
- **A rerun of the cancelled run itself** (challenger, lead). Under GitHub's newest-entrant semantics, a run created before a rerun attempt cannot cancel it. A run-level reading of "started" could still accept that earlier run, but the head it would exclude is either superseded or assessed by another run, so the outcome is harmless [INFERENCE].
- **Fixture discrimination.**
  - Removing only the lower bound fails fixtures 2 and 3, one per evidence path.
  - Removing only the upper bound fails fixture 1.
  - Removing both fails all three.
- **Mirrors.**
  - The five rule locations carry identical window and fallback text.
  - The three fixture locations carry the identical paragraph.
  - There is no stale window wording in the card or elsewhere outside the round files.
  - `policy-file.md` (digest unchanged) has no supersession text.
- **Ledger.** R-077 links round 12, R-070 is unchanged as the response states, no ID is reused, and STATUS next-free R-078 is consistent.
- **Historical bodies.** The digests are unchanged since round 12, and the challenger's raw comparison with base shows only the `:3` status lines differ.

## Considered and not raised

- **"Fixture 1 can pass with a never-started R1"** (experiment-peer, N1 part 1). The fixture heading requires "all other exclusion conditions satisfied", and fixture 1 isolates the upper bound. The lower bound must therefore hold, which means R1 started. A test that leaves R1 unstarted does not meet the fixture as written, and ADR012-R13-02's wording makes the coordinate explicit.
- **The stricter bound's cost in exposure** (experiment-peer T4, challenger). Queued-run supersessions now gap unless the collector reassesses the head. This fails closed, readiness stays reachable, and round 10 already weighed this loss of exclusions. Not a defect under the calibration rules.
- **"Collector reassessments only add attempts to existing observations" (`:16`) might forbid a first assessment** (experiment-peer caveat). The sentence limits the population to existing PR events and does not forbid accounting an existing event. Even read strictly, later gap-free intervals reach readiness. Not raised.
- **The fallback says "the head" while exclusion applies per attempt or event.** This matches the existing "is assessed/gapped" phrasing. Wording only.
- **The provider cannot name the canceller.** Settled in rounds 10–12. Not reopened.

## Peers and where they changed the outcome

- **`innovation-challenger`** reported zero issues and recommended closing ADR012-R12-01. It verified every mirror and found no regression, no remaining old wording, no ID reuse and no change to historical bodies.
  - It concluded that the "started" bound only narrows exclusions. This confirmed the lead's view that the deviation is not a defect on its merits.
  - Its point that the cancelled attempt's job start must govern for reruns shaped the coordinate wording in ADR012-R13-02.
- **`experiment-peer`** raised two minor findings.
  - Its N2, a rerun as superseder, became ADR012-R13-02. **This was new to the lead and changed the verdict to issues open.** The lead confirmed the re-run `GITHUB_REF` fact from GitHub's documentation and found no earlier treatment of the case.
  - Its N1 provenance point became ADR012-R13-01. The lead had independently noticed the textual deviation, and the peer's framing as a provenance error, by the R11-02 precedent, decided it.
  - Not adopted: N1 part 1, fixture-1 weakness (see Considered and not raised). Also not adopted: the peer's preferred option A of reverting to the creation bound; the lead recommends keeping the stricter bound and recording it accurately.
- **Lead-only:**
  - the severity calls;
  - the combined coordinate wording for the window clause;
  - the check that a rerun of the cancelled run itself is harmless;
  - the decision not to require a "R1 started" clause in fixture 1.

## Verdict

**ADR-012: 2 issues open (0 major, 2 minor). Not signed off.** Open: ADR012-R13-01 and ADR012-R13-02.

Round-12 closure: ADR012-R12-01 is fixed in substance at every rule and fixture location. Its provenance record is reopened as ADR012-R13-01.

All earlier closures (R1–R11) hold. The population and causal-supersession group is qualified by ADR012-R13-02.

No digests are signed in this round. The 13 digests above identify the bytes reviewed.

## Council memo

1. **Recommendation.** Do not sign off. Apply both fixes using the recommended options, then submit round 14 with fresh digests from Main for all 13 signed files and the response.
   - The R12 two-bound window is in place and consistent across every mirror.
   - Two gaps remain. The signed record misstates which lower bound round 12 recommended. And the superseding run's creation coordinate is not pinned against reruns, which a documented GitHub behaviour makes reachable and no fixture tests.
2. **Alternatives.**
   - **Recommended:** keep the stricter start bound, state both coordinates in attempt and creation terms with an explicit rerun exclusion, add fixture (4) for both evidence paths, and correct `:48`.
   - **B: revert to round 12's creation bound** at the five locations, plus the rerun exclusion and fixture (4). This restores exclusion for queued-run supersession, but reopens a choice that fails closed for no safety gain. Acceptable, not preferred.
   - **C: fixtures only for R13-02,** relying on "run … created" to exclude reruns. This gives the same discrimination with less churn, but leaves implementers to derive the coordinate. Acceptable, not preferred.
   - **D: drop supersession exclusion entirely.** Round 10 rejected this, and there is no new evidence. Not recommended.
   - **Signing off with known minor issues** is not permitted under design rule 1.
3. **Evidence.**
   - `ADR.md:17,19,21,48,57`; REQ-78 `:600`; REQ-27 `:612`; `ci:101,106`; `cli:126`; plan `:61`; `research-ledger.md:75–76`; STATUS `:131`.
   - `.github/workflows/scout-guardian.yml:41–43`.
   - `round-12.md` (the ADR012-R12-01 required fix); `round-12-response.md:3,7`; `round-2-response.md:31`.
   - GitHub: [re-running workflows and jobs](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/re-run-workflows-and-jobs), fetched by the lead; [control-workflow-concurrency](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency); [workflow-runs REST](https://docs.github.com/en/rest/actions/workflow-runs#get-a-workflow-run-attempt), cited by the challenger.
4. **Risks.**
   - That a rerun's concurrency evaluation cancels an in-progress newer run is [INFERENCE] from documented semantics. The implementation must confirm provider behaviour.
   - All timelines were constructed, not executed.
   - The council could not compute digests.
   - Both fixes are document changes, fully reversible, and only fail closed.
   - Residual risks accepted in earlier rounds (author-forced interval resets, scanner self-falsification, provider limits, loss of supersession exclusions) are unchanged.
5. **Execution outline.**
   1. At `ADR.md:17`, REQ-78 `:600`, REQ-27 `:612`, `ci:101` and `cli:126`, replace "created after the cancelled run started and before its cancellation" with "created (its first attempt; a rerun attempt of an earlier-created run never qualifies) after the cancelled attempt's job started and before that attempt's cancellation". Check that the five locations stay identical.
   2. Add fixture (4) from ADR012-R13-02 to `ADR.md:57`, plan `:61` and `ci:106`, keeping all existing fixtures.
   3. In `ADR.md:48`, replace the round-12 clause with the ADR012-R13-01 text, and add a round-13 delegated-decision clause covering the rerun exclusion.
   4. Add the re-run `GITHUB_SHA`/`GITHUB_REF` fact to R-070's summary, and extend R-077's title and links with `round-13.md`. No new ledger ID is needed.
   5. Write `round-13-response.md` mapping both issues. Main supplies the SHA-256 of all 13 signed files and of the response, then round 14 follows.
