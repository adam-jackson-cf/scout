# Round 12: ADR-012 (shadow-window readiness inputs)

Date: 2026-11-08. This round applies design rules 1–3 and the ledger rules in `CLAUDE.md`, plus the ADR rules in `docs/adrs/README.md`. ADR-012 is an agent decision made under the owner's standing delegation (R-064). Where a fix involves a choice, the council names the option it recommends under that delegation.

The round reviews the bytes revised after round 11 (`round-11-response.md`). It has three parts:

1. Check the ADR012-R11-01 and ADR012-R11-02 fixes against the required fixes in `round-11.md`.
2. Confirm that every earlier closure (R1–R10) still holds in the current bytes.
3. Look for new issues, numbered ADR012-R12-xx.

The council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review. It looked for false-ready paths, readiness that a compliant repository could never reach, mirror drift, leftover two-branch wording, ledger ID reuse and changes to historical ADR bodies. It also asked whether the provider can actually identify the superseding run.
- **`experiment-peer` (Opus):** methodological review. It built nine concrete timelines under the unified supersession condition. It then looked for faulty implementations that pass every fixture, and checked the Amendments provenance against responses 9–11.

The review was read-only. Nobody edited files, ran product code, made model calls outside the council, or touched `.codegraph`.

- The lead read the required codegraph skill references (`readiness-and-bootstrap.md`, `query-selection.md`). Codegraph was not used, because the material under review is documents and the council had no shell. Searches used `read`, `grep` and `glob`.
- The lead fetched no external page this round. Sources cited by the peers:
  - the experiment-peer checked the R-070 quotation word for word against [control-workflow-concurrency](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency);
  - the challenger cited the [workflow-runs REST page](https://docs.github.com/en/rest/actions/workflow-runs) and the [concurrency-groups REST page](https://docs.github.com/en/rest/actions/concurrency-groups).

**Main computed and verified the digests below with `shasum`. The council did not, because the lead and both peers had read-only tools and no shell.** This round is not a sign-off, so the digests only identify the bytes reviewed. All paths are under `docs/`.

| File | SHA-256 (Main) |
|---|---|
| `adrs/2026-11-08-shadow-readiness-inputs/ADR.md` | `7b15d01e3a2ea2c6776f650138cb731fe120d84f58cf75f7963cc0426a7d49e2` |
| `scout-design.html` | `d1cb827a12f528a69e930b7ce3534a3bd9dc4f6ea96c2a087ea1f59143d8a95f` |
| `adrs/README.md` | `cde415ff10b659a073d010f8a48c281d6f75a5a1cba466962d589e937b94587f` |
| `STATUS.md` | `17e6e4bf7f131111897600dcdbd17e5d51c13eace4f38b65cf7622357d3d563c` |
| `plans/2026-11-06-remaining-work.md` | `b980a7e7025e651fdb7ffab55efc269d4cefba7b0152386d7abc5d1293c15f8b` |
| `references/policy-file.md` | `7b747b02563fb3fdfa41914b99b9a7fa25a12546ad23fe98e922b99cccc80727` |
| `references/ci-integration.md` | `639b929a3d33f7c149c2d7293d023ca37908940303259ffb281f1d5a0e9dcecb` |
| `references/cli-commands.md` | `cbc7f686377703c471ae64a17a9e02713f988cb1b7bb90bb9776ec25b152ce04` |
| `research-ledger.md` | `4ad157ad250cb03af1fd7489a0e30578f9221d40f10eee662bbc219453af85af` |
| `adrs/2026-10-19-v1-decisions/ADR.md` (ADR-000) | `1cd521fd96439d87275ba353f7b05451c3e0c4d556ee665529aab85970ed81cb` |
| `adrs/2026-10-23-decisions/ADR.md` (ADR-003) | `973d24c4cfdfdd98acb4070a7bd72ca0c0e1ddf303bd4ccec8f92fe1cf117cc5` |
| `adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) | `ee593b6241814c24035671d666566bc2b7c399135867226c26e9626f3936abc8` |
| `adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) | `80e3375f81656a8856cf662cbec99a36d8e00bc451ff3c3592925d2293f44d95` |

Main also computed the SHA-256 of one unsigned record:

| File | SHA-256 (Main) |
|---|---|
| `round-11-response.md` | `8b5a884f077ed8ea635d673d6c29f8ebd1edc500431204f91b78abc16122ede6` |

README, STATUS, the policy-file reference and the four historical ADR files have the same digests as in round 11. The changed files are ADR.md, the design, the plan, the ci-integration and cli-commands references, and the ledger.

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11`, branch `agent/policy-shadow-adrs`, based on origin/main `7bb73e52`. Every change is staged and uncommitted.
- Base bytes come from `/Users/adamjackson/Projects/work/scout-adr11b`.

**Files read.** Line numbers below refer to the raw worktree files.
- ADR-012 `ADR.md`, all lines. The main ones are `:17` (population and supersession), `:19` (clocks and frontier), `:48` (amendments) and `:56–62` (acceptance invariants).
- `round-11.md` and `round-11-response.md`, all lines.
- `scout-design.html`, raw:
  - REQ-78 `:600`;
  - REQ-27 `:612`;
  - the ADR-012 card `:1009`;
  - grep context for the ADR-000/003/007/008 header lines `:736,807,901,925`.
- `ci-integration.md:95–108`, `cli-commands.md:124–136` and `policy-file.md:78–84`, all raw.
- Plan `:61`, raw; `research-ledger.md:75–76`, raw; STATUS `:131`.
- `.github/workflows/scout-guardian.yml:1–60`. The concurrency block is at `:41–43`, at workflow level.
- A docs-wide grep, with the historical round files excluded, for "inferred", "provider cause", "concurrency cause", "cancel-in-progress" and "supersed".
- The challenger compared the four historical ADRs and REQ-160/162/163/168 with base. The experiment-peer also read `scout-guardian-trusted.yml:53–55` and `round-9/10-response.md`.

## Issue summary

**1 issue is open, and it is minor. None is major.**

| ID | Severity | Summary |
|---|---|---|
| ADR012-R12-01 | minor | The rule states only the upper bound of the time window that identifies the superseding run ("created before cancellation"). No fixture tests either bound on its own. A naive implementation that accepts an earlier run as the superseding run passes every fixture, yet excludes the final head after a PR-added job, or a cancel with no recorded actor, cancels its guardian run |

## Closure of round-11 issues

"Fixed" means the required document fix is present in the reviewed bytes at every named location. It does not mean the implementation is validated.

| Issue | Status | Evidence |
|---|---|---|
| ADR012-R11-01 | **Fixed** | **Rule.** The same three sentences appear at all five locations: `ADR.md:17`, REQ-78 `:600`, REQ-27 `:612`, `ci:101` and `cli:126`. They read:<br>"Supersession exclusion requires an identified superseding run of the pinned guardian lane for the same PR key, created before cancellation, whose executed configuration verifiably equals the pinned protected-default configuration and evaluates cancel-in-progress=true for the same evaluated concurrency group as the cancelled run. Additionally require either a provider-recorded concurrency cause or no actor-initiated cancel recorded; the provider cause replaces only that last condition, never the run/configuration checks. Cancellation by another workflow or job, an unidentified canceller, or a PR-edited superseding run is assessed/gapped."<br>This is the single condition round 11 prescribed.<br>**Old wording gone.** The docs-wide grep (round files excluded) finds no "Inferred supersession", "irrelevant to this inference" or "Supersession needs" anywhere. The only remaining "inferred-cause" is the historical provenance clause at `ADR.md:48`.<br>**Fixtures.** All three prescribed fixtures are present at `ADR.md:57`, plan `:61` and `ci:106`:<br>1. PR-added job, provider cause, no H2;<br>2. pinned `false` with H2 PR-editing R2 to `true` and no recorded canceller;<br>3. the same with a provider cause.<br>The all-false, R1-true/R2-false and unverifiable-R2 fixtures are kept.<br>**Ledger.** R-070 (`research-ledger.md:75`) adds the control-workflow-concurrency link and the quoted cross-workflow warning, and the peer verified the quote word for word. R-077 (`:76`) is retitled "rounds 2–11" and links `round-11.md`.<br>**Delegated decision.** The round-11 decision is recorded at `ADR.md:48`. |
| ADR012-R11-02 | **Fixed** | `ADR.md:48` now reads: "supersession's inferred-cause branch was conditioned on the cancelled run's pinned cancel-in-progress (replaced in round 10)". This is the prescribed text verbatim. It matches `round-9-response.md:10,22`. The round-10 clause ("take cancellation inference only from the superseding run's verified executed configuration") matches `round-10-response.md`, and the round-11 clause matches `round-11-response.md:7` (peer). No mirror needed changing. |

## Re-confirmation of earlier closures (R1–R10)

| Closure | Holds? | Evidence |
|---|---|---|
| R1-04: started_at membership; elapsed-days formula | Yes | `ADR.md:19–21` |
| R1-05, R7-01, R8-07, R10-03: recoverable frontier, deadline point gaps, CAS-parent anchor, (parent C, new C] range | Yes | `ADR.md:19`; REQ-78 `:600`; `ci:101`; `cli:126`; `policy-file:82` |
| R2-05, R6-01/08, R7-04, R8-05, R9-04, R10-04, R11-01: population, exclusions, causal supersession | Yes, apart from ADR012-R12-01 | `ADR.md:16–17`; REQ-27/78; `ci:101`; `cli:126` |
| R5-09: scanner-output failure is comparison-none | Yes | `ADR.md:17` |
| R1-08/09/11/14, R2-10, R3-01, R4-07/10: CI/local/legacy separation, migration, v2 cutover, query-only v13 | Yes | `ADR.md:34–38,44` |
| R3-02/07, R4-02/08/11, R8-02, R9-02/03: authorization, current-tip authority, replay, writer isolation, lineage, import refusal | Yes | `ADR.md:26,28,36,40`; `cli:130`; `policy-file:82` |
| R3-03, R4-06: guardian comparison, empty override inputs | Yes | `ADR.md:17–18`; REQ-27 |
| R2-11, R3-05, R4-03/04, R5-03/05/07, R6-02/07, R7-01, R8-03, R9-01: interval selection, cutoffs, frontier minima, dossier invariance, single predicate | Yes | `ADR.md:19–20,28,44`; REQ-78; `cli:126–132` |
| R8-04, R10-02: exclusive point boundaries; strict and equal boundary fixtures | Yes | `ADR.md:19,58`; plan `:61`; `ci:106` |
| R10-01: independent day-30 rerun assertions; both halves of deadline visibility | Yes | `ADR.md:57–58`; plan `:61`; `ci:106` |
| R5-06, R6-03, R8-01: liveness, skewed-clock triage, scoped refusal | Yes | `ADR.md:19,44,60` |
| R5-01, R6-10: wiring-ADR gate | Yes | `ADR.md:12,56`; `ci:99`; plan `:61`; `policy-file:84`; `cli:134` |
| R5-04, R6-04, R7-02, R8-06, R9-05: cohort-wide zeroes including provisional; fixtures cannot pass by refusal alone | Yes | `ADR.md:20,30,57–60` |
| R5-08, R6-05: Days, representative origin | Yes | `ADR.md:21,48,62`; card `:1009` |
| R5-10, R7-05/06/07: exact cohort, default version, dossier encoding, attempt-ID order, pin tuples | Yes | `ADR.md:21,44,61–62`; REQ-78 |
| R7-03: event-to-attempt binding | Yes | `ADR.md:17`; REQ-27/78 |
| R7-08/09: mirror drift; Activation callout | Yes | Plan `:61` and `ci:106` carry every ADR supersession fixture. Card `:1009` and the REQ mirrors agree with the ADR |
| Historical ADRs and untouched requirements | Yes | ADR-000/003/007/008, README and STATUS digests are unchanged since round 11. The challenger's raw comparison found only the `:3` status lines differ from base. REQ-160/162/163/168 match base |
| Ledger | Yes | R-070 and R-077 are updated as above. STATUS `:131` names next free R-078. No R-ID or ADR ID is reused |

## Are the invariants testable and enough for REQ-22, REQ-27 and REQ-78?

| Requirement need | Testable and sufficient? | Where / gap |
|---|---|---|
| REQ-22: counting unit, deduplication, exact cohort, unknown ≤ 10 % | Yes | `ADR.md:20–22` |
| REQ-22: 14 days, inspected origin, `as_of` endpoint | Yes | `ADR.md:19,21,58,62` |
| REQ-22: zero `not_warranted` and zero Undecided across gaps, late and imported evidence | Yes | `ADR.md:20,57–60` |
| REQ-22: maintainer-only, append-only, idempotent triage | Yes | `ADR.md:26,28,60` |
| REQ-27: population independent of content, author and approval | **Partly** | The rule text is sound; the fixtures miss two cases (ADR012-R12-01) |
| REQ-78: complete-interval accounting, recoverable frontier | Yes | `ADR.md:19,58` |
| REQ-78: historical reproduction, `dossier_digest` | Yes | `ADR.md:19,44,58,61` |
| REQ-78: current certification and liveness | Yes | `ADR.md:19,44,58,60` |
| Governance of deferred wiring | Yes | `ADR.md:12,56` |

## Issues

### ADR012-R12-01: The superseding-run time window has only an explicit upper bound, and no fixture tests either bound on its own (minor)

- **Location:**
  - The rule, "created before cancellation", in five places: `ADR.md:17`, REQ-78 `:600`, REQ-27 `:612`, `ci:101` and `cli:126`.
  - The fixtures:
    - `ADR.md:57`: "Manual pre-assessment cancellation, including cancellation followed by a later push … assessed or gapped", plus the R11 fixtures;
    - plan `:61` and `ci:106`: "Manual cancel then later push is assessed/gapped", plus the R11 fixtures.
- **New evidence:**
  - GitHub's documented run and concurrency APIs expose run identity, workflow, PR association, timestamps and live group membership. They expose no field naming the run that cancelled another (challenger, citing the [workflow-runs](https://docs.github.com/en/rest/actions/workflow-runs) and [concurrency-groups](https://docs.github.com/en/rest/actions/concurrency-groups) REST pages) [INFERENCE].
  - An implementation therefore has to identify "the superseding run" with a time-window query over same-PR pinned-lane runs. The rule states only that window's upper end.
  - Round 11's fixture 1 specifies "no H2", so it never offers an earlier same-PR run as a candidate.
- **Problem** (experiment-peer; adjudicated by the lead) [INFERENCE: constructed timelines, not executed]:
  - **(a) The lower bound is missing; an earlier run is accepted.**
    1. H1 adds `.github/workflows/x.yml`, and R1 runs on H1.
    2. H2 is pushed, and unedited R2 legitimately cancels R1, so H1 is excluded, correctly.
    3. X2's job, now carrying H2's content, joins the guardian group with `cancel-in-progress: true` and cancels R2 before assessment. A provider cause is recorded, and there is no H3.
    - A naive implementation looks for "a same-PR pinned-lane run, created before R2's cancellation, with verified pinned configuration evaluating `true` in the same group". It finds R1, which meets every one of those stated conditions, and excludes the final head H2 with no gap.
    - The same happens when R2 is cancelled manually with no actor recorded.
    - The word "superseding" rules R1 out under any ordinary reading, so the rule text has no open reading. But no fixture fails this implementation: F5 has one guardian run, and the pinned-`false` fixtures fail on `true` anyway.
    - The result is the R11-01 harm reached by another route. The final merged head cannot veto.
  - **(b) The upper bound is never tested on its own.**
    - The manual-cancel-then-later-push fixture does not specify pinned `true`, an unedited configuration, no recorded actor and no provider cause.
    - Under any other settings the head gaps for another reason. The fixture can therefore pass even when the implementation skips "created before cancellation".
    - With no actor recorded, which the rule explicitly allows for, that check alone separates exclusion from a gap (the R8-05 harm).
- **Severity:** the lead adjudicates minor, the same as R10-04 and R11-01.
  - The rule's wording already gives the right answer; the defect is that the binding acceptance fixtures cannot tell a correct implementation from a naive one.
  - Path (a) needs either an adversarial workflow or a manual cancel with no actor recorded.
  - The cancellation shows as a cancelled guardian check.
  - The fix only fails closed. Every legitimate supersession has a newer superseding run, so a compliant repository's readiness is unaffected.
- **Required fix (recommended option):**
  - **Rule:** in `ADR.md:17`, REQ-78 `:600`, REQ-27 `:612`, `ci:101` and `cli:126`, replace "created before cancellation" with "created after the cancelled run and before its cancellation". This states both ends of the only window an implementation can actually query. When GitHub's non-guaranteed ordering lets an older run cancel a newer one ([control-workflow-concurrency](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency)), the newer head now gaps, which fails closed.
  - **Fixtures** (`ADR.md:57`, plan `:61`, `ci:106`), keeping all existing ones:
    1. Pinned `true`, unedited configuration, no actor recorded and no provider cause. R1 is cancelled before assessment, and H2's R2 is created after the cancellation. H1 is assessed or gapped.
    2. Pinned `true`. R2 legitimately supersedes R1. A PR-added workflow job then cancels R2 before assessment with a provider cause, and there is no H3. H2 is assessed or gapped.
    3. Fixture 2 again, but R2 is cancelled manually with no actor recorded and no provider cause. H2 is assessed or gapped.
  - **Records:** record the round-12 delegated decision in `ADR.md:48`, and extend R-077 with `round-12.md`. No new ledger ID is needed.

## Checks that found no issue

- **The unified condition across nine timelines** (experiment-peer). Each gives the correct outcome, or fails closed:
  - pinned `true` with unedited R2: excluded;
  - pinned `false`, then a manual cancel: gapped;
  - a PR-added workflow with no H2: gapped;
  - pinned `false` with R2 PR-edited to `true`, with or without a provider cause: gapped;
  - the base changes `true`→`false` between runs: gapped;
  - the base changes `false`→`true` between runs: excluded or gapped. Both fail closed, and round 11 accepted this;
  - a cancel with the actor recorded: gapped;
  - a cancel followed by a later push: gapped;
  - a superseding run that cannot be verified: gapped.
- **Fixture discrimination for the R11 targets.** An implementation that checks only "R2 evaluates `true`" fails fixtures 2 and 3. One that trusts the provider cause without identifying the run fails fixture 1. One that checks only pinned equality fails the round-9 pinned-`false` fixture.
- **A qualifying R2 exists but X was the real canceller** (lead, peer). With workflow-level `cancel-in-progress: true` (`scout-guardian.yml:41–43`), R1 is cancelled as soon as R2 enters the group, so the outcome would have been the same.
- **The "same PR key" check.** It already follows from an equal evaluated group plus the pinned configuration, because the group is keyed by `github.ref`. Using another PR's group needs a non-pinned configuration, which gaps the head.
- **Mirrors.** The five rule locations carry identical text. The three fixture locations carry identical R11 fixtures. `policy-file.md` has no supersession text to drift. Card `:1009` is consistent.
- **Readiness for a compliant repository.** Even if every unverifiable supersession gaps, a later gap-free interval, or collector reassessment of the cancelled head, reaches readiness (challenger). Round 10 already weighed this deflation risk.

## Considered and not raised

- **Whether the provider can establish causality at all** (challenger). This was settled in rounds 10 and 11, which rejected requiring the provider to identify the canceller. Pinned-`true` configuration plus group membership plus window timing implies causation under GitHub's documented semantics [INFERENCE]. The challenger's evidence instead motivates stating both window bounds in ADR012-R12-01.
- **The configuration compared with the cancelled run instead of the pinned one** (peer, point 3). The rule explicitly says "equals the pinned protected-default configuration". The case only arises under a hypothetical future pinned `false` with both heads PR-edited, and the later head is still assessed. Not raised.
- **No fixture asserts that a legitimate supersession is excluded** (peer). Exclusion is permitted, not required, and round 7 accepted the reduced-readiness cost.
- **The older run cancelling the newer one under out-of-order processing** (peer). This depends on provider timing, not content. After ADR012-R12-01's lower bound it fails closed.
- **The trusted lane's shared group expression with `cancel-in-progress: false`** (peer, `scout-guardian-trusted.yml:53–55`). Its ref differs, and it is not the pinned lane, so it would gap.
- **A byte-copy of the guardian workflow at another path** (lead). The rule requires "a run of the pinned guardian lane". Natural enumeration keys on the workflow identity, and fixture 1 already requires a non-guardian workflow to be rejected.
- **Reusable-workflow caller edits** (lead). These are covered by "executed configuration" and "PR-edited superseding run". Nothing new.
- **Wording:** the sentence order in R-070 ("(round 11). Supersession inference therefore requires …"), the repeated "Retain the existing …" sentence in `ADR.md:57`, and the "Round-9 fixtures:" label (already noted in round 11).

## Peers and where they changed the outcome

- **`innovation-challenger`** reported zero issues and rated R11-01 and R11-02 fixed. Its evidence matched the lead's for every closure. It found no regression in R1–R10, no leftover two-branch wording, no change to historical bodies and no ID reuse.
  - Its point that the provider has no documented canceller field was not raised as an issue. The lead used it as the evidence that the window bounds must be explicit in ADR012-R12-01.
- **`experiment-peer`** rated both round-11 issues fixed and raised one minor finding with three parts.
  - Parts (b) and (a) became ADR012-R12-01. **This was new to the lead and changed the verdict from zero issues to issues open.**
  - The lead made the peer's *optional* lower-bound wording part of the *required* fix, because of the challenger's evidence.
  - Not adopted: part 3, the cancelled-run configuration comparison (see Considered and not raised).
- **Lead-only:**
  - folding both window bounds into one issue;
  - the severity call;
  - the copy-workflow and reusable-workflow checks.

## Verdict

**ADR-012: 1 issue open (0 major, 1 minor). Not signed off.** Open: ADR012-R12-01.

Round-11 closure: ADR012-R11-01 and ADR012-R11-02 are fixed.

All earlier closures hold. The population and causal-supersession group is qualified by ADR012-R12-01.

No digests are signed in this round. The 13 digests above identify the bytes reviewed.

## Council memo

1. **Recommendation.** Do not sign off. Apply the ADR012-R12-01 fix using the recommended option, then submit round 13 with fresh digests from Main for all 13 signed files and the response.
   - The R11 fixes are clean and consistent across every mirror.
   - The only gap left is how the superseding run is identified. Its time window has an unstated lower bound and no discriminating fixtures, and the provider gives no other way to identify the run.
2. **Alternatives.**
   - **Recommended:** state the window as "created after the cancelled run and before its cancellation" at the five rule locations, and add the three fixtures.
   - **Alternative B: fixtures only.** Rely on the word "superseding" for the lower bound, but add fixtures 1–3. This gives the same discrimination with less churn, but leaves implementers to derive the lower bound. Acceptable, not preferred.
   - **Alternative C: drop supersession exclusion and always reassess cancelled heads** (the challenger's preferred design direction). This is the most robust option, but it changes an R10/R11 decision, and round 10 rejected it for collection cost and deflation. Not recommended without new evidence.
   - **Signing off with a known minor issue** is not permitted under design rule 1.
3. **Evidence.**
   - `ADR.md:17,48,57`; REQ-78 `:600`; REQ-27 `:612`; `ci:101,106`; `cli:126`; plan `:61`; `research-ledger.md:75–76`; STATUS `:131`.
   - `.github/workflows/scout-guardian.yml:41–43`.
   - `round-9-response.md:10,22`; `round-11-response.md:7`.
   - GitHub [control-workflow-concurrency](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency) (cross-workflow cancellation; "ordering is not guaranteed"), the [workflow-runs REST API](https://docs.github.com/en/rest/actions/workflow-runs) and the [concurrency-groups REST API](https://docs.github.com/en/rest/actions/concurrency-groups). These were cited by the peers; the lead did not fetch them this round.
4. **Risks.**
   - Whether a run identified by the time window is in fact the canceller is [INFERENCE] from GitHub semantics. The implementation must confirm provider behaviour.
   - All timelines were constructed, not executed.
   - The council could not compute digests.
   - The fix is a document change, fully reversible, and only fails closed.
   - Residual risks accepted in earlier rounds (author-forced interval resets, scanner self-falsification, provider limits, supersession deflation) are unchanged.
5. **Execution outline.**
   1. In `ADR.md:17`, REQ-78 `:600`, REQ-27 `:612`, `ci:101` and `cli:126`, replace "created before cancellation" with "created after the cancelled run and before its cancellation".
   2. Add fixtures 1–3 from ADR012-R12-01 to `ADR.md:57`, plan `:61` and `ci:106`, keeping all existing fixtures.
   3. Record the round-12 delegated decision in `ADR.md:48`. Extend R-077 (title and links) with `round-12.md`. No new ledger ID is needed.
   4. Write `round-12-response.md` mapping the issue. Main supplies the SHA-256 of all 13 signed files and of the response, then round 13 follows.
