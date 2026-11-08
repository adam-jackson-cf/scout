# Round 14: ADR-012 (shadow-window readiness inputs)

Date: 2026-11-08. This round applies design rules 1–3 and the ledger rules in `CLAUDE.md`, plus the ADR rules in `docs/adrs/README.md`. ADR-012 is an agent decision made under the owner's standing delegation (R-064). Where a fix involves a choice, the council names the option it recommends under that delegation.

The round reviews the bytes revised after round 13 (`round-13-response.md`). It has three parts:

1. Check the ADR012-R13-01 and ADR012-R13-02 fixes against the required fixes in `round-13.md`.
2. Confirm that every earlier closure (R1–R12) still holds in the current bytes.
3. Look for new issues, numbered ADR012-R14-xx.

The council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review. It checked that the new window clause matches word for word at the five rule locations and that fixture (4) matches at the three fixture locations. It also checked the added fallback text, the accuracy of the Amendments provenance, the ledger edits, ID reuse, historical ADR bodies against base and any stale wording. On GitHub semantics it tested full reruns, failed-job-only reruns and single-job reruns against the new "first attempt" and "cancelled attempt's job started" coordinates.
- **`experiment-peer` (Opus):** methodological review. It built timelines under the new wording and a discrimination table of seven implementations against fixtures (1)–(4). It looked for any faulty implementation that passes all four fixtures and still excludes a final head, checked whether any fixture contradicts the rule, and checked the Amendments provenance against `round-12.md`, `round-13.md` and `round-13-response.md`.

The review was read-only. Nobody edited files, ran product code, made model calls outside the council, or touched `.codegraph`.

- The lead read the required codegraph skill references (`readiness-and-bootstrap.md`, `query-selection.md`). The council did not use codegraph, because the material under review is documents and the council had no shell. Searches used `read`, `grep` and `glob`.
- The lead ran a web search on GitHub's rerun timestamp semantics. The search summary, which cites the [workflow-runs REST page](https://docs.github.com/en/rest/actions/workflow-runs), says a rerun keeps the run's `id`, increments `run_attempt`, leaves `created_at` at the original run's creation, and changes `run_started_at` for each attempt. Nobody read the page directly, so this is [INFERENCE]. If it is wrong, the new fallback clause fails closed.
- The challenger cited GitHub's [contexts](https://docs.github.com/en/actions/reference/workflows-and-actions/contexts) page (stable run IDs, incrementing attempts), the [re-running workflows and jobs](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/re-run-workflows-and-jobs) page (full, failed-job-only and single-job reruns, with retained outputs for jobs not rerun) and the [workflow-jobs REST](https://docs.github.com/en/rest/actions/workflow-jobs) page (job identity, `run_attempt`, `started_at`).

**Main computed and verified the digests below with `shasum`. The council did not, because the lead and both peers had read-only tools and no shell.** All paths are under `docs/`.

| File | SHA-256 (Main) |
|---|---|
| `adrs/2026-11-08-shadow-readiness-inputs/ADR.md` | `a2869a076e0820cc13552ac9055cee22d0a57effb7d770c4501c9fb1c20343d8` |
| `scout-design.html` | `54d769cbcaec081a6764ef35f2898a91a3f839537ab7ea79e978763de5d7f587` |
| `adrs/README.md` | `cde415ff10b659a073d010f8a48c281d6f75a5a1cba466962d589e937b94587f` |
| `STATUS.md` | `17e6e4bf7f131111897600dcdbd17e5d51c13eace4f38b65cf7622357d3d563c` |
| `plans/2026-11-06-remaining-work.md` | `33afb05b4bd721e381475368e8354bc8e71aa24656e996a56819190f8154fe35` |
| `references/policy-file.md` | `7b747b02563fb3fdfa41914b99b9a7fa25a12546ad23fe98e922b99cccc80727` |
| `references/ci-integration.md` | `057594820d494584c004ba8d02a600fe1384196dcafa1f669500127aa1cf4ec4` |
| `references/cli-commands.md` | `a5ce50bab10d6c3348053e2321ce9e88499830780dd117ea0279e303b19d21e8` |
| `research-ledger.md` | `1ff9e7158b5619b41d1bb8a060d2b8ab9d58080a9c9620ba0524092f65313f77` |
| `adrs/2026-10-19-v1-decisions/ADR.md` (ADR-000) | `1cd521fd96439d87275ba353f7b05451c3e0c4d556ee665529aab85970ed81cb` |
| `adrs/2026-10-23-decisions/ADR.md` (ADR-003) | `973d24c4cfdfdd98acb4070a7bd72ca0c0e1ddf303bd4ccec8f92fe1cf117cc5` |
| `adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) | `ee593b6241814c24035671d666566bc2b7c399135867226c26e9626f3936abc8` |
| `adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) | `80e3375f81656a8856cf662cbec99a36d8e00bc451ff3c3592925d2293f44d95` |

Main also computed the SHA-256 of one unsigned record:

| File | SHA-256 (Main) |
|---|---|
| `round-13-response.md` | `c6f6fdd1601adb152a72f95ad01dd4f7c235d211db0cf236eb37cfa3f869a703` |

**Changes since round 13:**
- Unchanged: README, STATUS, the policy-file reference and the four historical ADR files keep their round-13 digests.
- Changed: ADR.md, the design, the plan, the ci-integration and cli-commands references, and the ledger. These are the files the round-13 fixes name.

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11`, branch `agent/policy-shadow-adrs`, based on origin/main `7bb73e52`. Every change is staged and uncommitted.
- Base bytes come from `/Users/adamjackson/Projects/work/scout-adr11b`.

**Files read.** Line numbers refer to the raw worktree files.
- ADR-012 `ADR.md`, all lines. The main ones are:
  - `:17`: population, exclusions, the supersession window and the fallback;
  - `:19`: clocks ("started_at is provider job start, not run creation");
  - `:21`: attempt ID;
  - `:48`: Amendments, including the round-12 and round-13 clauses;
  - `:57`: the supersession-window fixtures (1)–(4).
- `round-12.md`, `round-13.md` and `round-13-response.md`, all lines.
- `scout-design.html`, raw: REQ-78 `:600`, REQ-27 `:612`, REQ-22 `:589`, REQ-48 `:615`, Appendix B status headers `:736–925`, and the ADR-012 card `:1007–1012`.
- `ci-integration.md:101,106` and `cli-commands.md:126`.
- Plan `:61`; `research-ledger.md:75–76`, raw; STATUS `:131`.

**Searches:**
- A full-sentence grep of the window clause together with its fallback sentence matched all five rule locations. A second full-paragraph grep of fixtures (1)–(4) matched all three fixture locations.
- A docs-wide grep, excluding the round files, for "rerun-attempt time", "rerun attempt", "created after the cancelled", "created before cancellation" and "cancelled run started".
- A design-wide grep for "supersed" and "rerun".
- The challenger compared the four historical ADRs with base, raw.

## Issue summary

**No issues are open.** No ADR012-R14-xx issue was raised.

## Closure of round-13 issues

"Fixed" means the required document fix is present in the reviewed bytes at every named location. It does not mean the implementation is validated.

| Issue | Status | Evidence |
|---|---|---|
| ADR012-R13-01 | **Fixed** | `ADR.md:48` contains round 13's required text verbatim, apart from curly apostrophes: "Round-12 delegated decision: adopt a two-bound superseding-run window ending at the cancelled attempt’s cancellation; its lower bound is the cancelled attempt’s job start, deliberately stricter than round 12’s recommended run-creation bound, so a cancelled run that never started is assessed/gapped; absent qualifying proof assesses/gaps the head rather than excluding it."<br>The experiment-peer checked this against round 12's recommended "created after the cancelled run and before its cancellation", which is a creation bound, so the record is now accurate.<br>The following round-13 clause, "retain that stricter fail-closed lower bound and date a candidate superseding run by its original run created_at, never rerun-attempt time; missing distinction falls to the existing assess/gap fallback", matches round 13's recommended option. It credits the council with nothing it did not recommend. |
| ADR012-R13-02 | **Fixed, with one recorded deviation that fails closed** | **Rule.** `ADR.md:17`, REQ-78 `:600`, REQ-27 `:612`, `ci:101` and `cli:126` carry the same full sentence, matched by one grep: "created (run created_at at its first attempt, never rerun-attempt time; a rerun attempt of an earlier-created run never qualifies) after the cancelled attempt’s job started and before that attempt’s cancellation, whose executed configuration verifiably equals …". Round 13's required coordinates are all present. The added "run created_at … never rerun-attempt time" names the same coordinate using the provider field. The configuration, concurrency-group, `cancel-in-progress` and cause conditions are unchanged.<br>**Fallback (deviation).** Round 13 said to keep the fallback sentence unchanged. All five locations now insert "including when original run creation cannot be distinguished from rerun-attempt time". This is a special case of "no qualifying run" and can only fail closed. It is recorded at `:48` and in the response. Accepted, not a defect.<br>**Fixture.** Fixture (4) is at `ADR.md:57`, plan `:61` and `ci:106`, matched by one paragraph grep, verbatim as required, and fixtures (1)–(3) are kept.<br>**Records.** `:48` adds the round-13 decision. R-070 (`research-ledger.md:75`) adds: "Reruns reuse the original GITHUB_SHA/GITHUB_REF, so a rerun of an older run joins the same ref-keyed group (round 13; cancellation follows concurrency semantics [INFERENCE]); the candidate window must use original run created_at, never rerun-attempt time". R-077 (`:76`) is retitled "rounds 2–13" and links `round-13.md`. No new ID was created. |

## Re-confirmation of earlier closures (R1–R12)

| Closure | Holds? | Evidence |
|---|---|---|
| R1-04: started_at membership; elapsed-days formula | Yes | `ADR.md:19–21` |
| R1-05, R7-01, R8-07, R10-03: recoverable frontier, deadline point gaps, CAS-parent anchor, (parent C, new C] range | Yes | `ADR.md:19`; REQ-78 `:600`; `ci:101`; `cli:126`; `policy-file` (digest unchanged) |
| R2-05, R6-01/08, R7-04, R8-05, R9-04, R10-04, R11-01, R12-01, R13-02: population, exclusions, causal supersession, two-bound window, rerun exclusion | Yes | `ADR.md:16–17`; REQ-27/78; `ci:101`; `cli:126`. The five locations are identical |
| R11-02, R13-01: Amendments provenance for rounds 9 and 12 | Yes | `ADR.md:48`: "conditioned on the cancelled run’s pinned cancel-in-progress (replaced in round 10)"; the round-12 clause is quoted above |
| R5-09: scanner-output failure is comparison-none | Yes | `ADR.md:17` |
| R1-08/09/11/14, R2-10, R3-01, R4-07/10: CI/local/legacy separation, migration, v2 cutover, query-only v13 | Yes | `ADR.md:34–38,44`; REQ-48 `:615` |
| R3-02/07, R4-02/08/11, R8-02, R9-02/03: authorization, current-tip authority, replay, writer isolation, lineage, import refusal | Yes | `ADR.md:26,28,36,40`; `ci:102` |
| R3-03, R4-06: guardian comparison, empty override inputs | Yes | `ADR.md:17–18`; REQ-27 |
| R2-11, R3-05, R4-03/04, R5-03/05/07, R6-02/07, R7-01, R8-03, R9-01: interval selection, cutoffs, frontier minima, dossier invariance, single predicate | Yes | `ADR.md:19–20,28,44`; REQ-78; `cli:126–128` |
| R8-04, R10-02: exclusive point boundaries; strict and equal boundary fixtures | Yes | `ADR.md:19,58` |
| R10-01: independent day-30 rerun assertions; both halves of deadline visibility | Yes | `ADR.md:57–58`; plan `:61`; `ci:106` |
| R5-06, R6-03, R8-01: liveness, skewed-clock triage, scoped refusal | Yes | `ADR.md:19,44,60` |
| R5-01, R6-10: wiring-ADR gate | Yes | `ADR.md:12,56` |
| R5-04, R6-04, R7-02, R8-06, R9-05: cohort-wide zeroes including provisional; fixtures cannot pass by refusal alone | Yes | `ADR.md:20,30,57–60` |
| R5-08, R6-05: Days, representative origin | Yes | `ADR.md:21,62`; card `:1009` |
| R5-10, R7-05/06/07: exact cohort, default version, dossier encoding, attempt-ID order, pin tuples | Yes | `ADR.md:21,44,61–62`; REQ-78 |
| R7-03: event-to-attempt binding | Yes | `ADR.md:17`; REQ-27/78 |
| R7-08/09: mirror drift; Activation callout | Yes | The five rule locations carry an identical window and fallback, and the three fixture locations carry an identical paragraph. No stale window wording appears outside the round files. Card `:1009` has no window wording |
| Historical ADRs and untouched requirements | Yes | ADR-000/003/007/008, README, STATUS and policy-file digests are unchanged since round 13. The challenger's raw comparison with base found that only the `:3` status lines differ. The design headers `:736,807,901,925` name the ADR-012 amendments |
| Ledger | Yes | R-070 and R-077 are extended without a new ID. STATUS `:131` still names next free R-078, with R-065–068 and R-071–076 pre-allocated. No R-ID or ADR ID is reused |

## Are the invariants testable and enough for REQ-22, REQ-27 and REQ-78?

| Requirement need | Testable and sufficient? | Where |
|---|---|---|
| REQ-22: counting unit, deduplication, exact cohort, unknown ≤ 10 % | Yes | `ADR.md:20–22` |
| REQ-22: 14 days, inspected origin, `as_of` endpoint | Yes | `ADR.md:19,21,58,62` |
| REQ-22: zero `not_warranted` and zero Undecided across gaps, late and imported evidence | Yes | `ADR.md:20,57–60` |
| REQ-22: maintainer-only, append-only, idempotent triage | Yes | `ADR.md:26,28,60` |
| REQ-27: population independent of content, author and approval | Yes | `ADR.md:17,57`. Fixture (4) closes the rerun-canceller gap |
| REQ-78: complete-interval accounting, recoverable frontier | Yes | `ADR.md:19,58` |
| REQ-78: historical reproduction, `dossier_digest` | Yes | `ADR.md:19,44,58,61` |
| REQ-78: current certification and liveness | Yes | `ADR.md:19,44,58,60` |
| Governance of deferred wiring and delegated provenance | Yes | `ADR.md:12,48,56` |

## Checks that found no issue

- **Fixture discrimination** (experiment-peer, verified by the lead) [INFERENCE: constructed timelines, not executed]:
  - removing only the upper bound fails (1);
  - removing only the lower bound fails (2), (3) and (4);
  - dating the superseder by its rerun attempt's creation or `run_started_at` fails (4) on both evidence paths, because the rerun time falls inside (R2 job start, R2 cancellation) and H2 would be excluded;
  - the correct implementation passes all four.
- **No faulty implementation passes all four fixtures and falsely excludes a final head** (experiment-peer, lead) [INFERENCE].
  - Three variants pass all four fixtures: a lower bound taken from the cancelled run's creation, a lower bound taken from another attempt of the cancelled run, and dating the superseder by its first job start.
  - None of them excludes a final head. Any same-PR pinned-lane run created after the cancelled run began is either on a newer head, so the excluded head is not final, or on the same head, so it accounts for or gaps that head itself. The variants therefore only add exclusions of heads that are already superseded, or fail closed.
- **No fixture contradicts the rule.**
  - In fixture (4), R1 was created before R2's job started, so its rerun is "a rerun attempt of an earlier-created run".
  - In fixture (2), "legitimately supersedes" implies R1 had started before R2 was created, which meets the stricter bound.
- **Rerun modes** (challenger) [INFERENCE]:
  - A full or partial rerun cannot move an earlier-created candidate's creation into the window.
  - When the cancelled guardian job is itself rerun, that job attempt's start governs.
  - A partial rerun that leaves the guardian job untouched supplies no new job-start coordinate under `ADR.md:19,21`.
  - Queued cancellations, equal timestamps, delayed admission, or unavailable proof of original creation can lose exclusions. All of them fall to assess/gap, not to a false exclusion.
- **A rerun cannot be the cancelling event for a qualifying new run** (lead) [INFERENCE]. Suppose a new run R2 was created inside the window. Its first attempt joined the group while the cancelled attempt was in progress, so that entry is what cancelled it. A later R2 rerun could only happen after R2's first attempt was terminal. The "first attempt" coordinate therefore never misdates a legitimate superseder.
- **Readiness stays reachable.** No rule or fixture requires a supersession exclusion. Lost exclusions fall to collector reassessment or gaps, and a later complete interval reaches readiness. Rounds 10 and 13 already accepted the loss of supersession exclusions by name.
- **Clock vocabulary.** "started_at is provider job start, not run creation" (`:19`) defines attempt coordinates. "run created_at" in the window names a separate provider field used only for the superseding run. There is no conflict.
- **Mirrors and stale wording.** No "created before cancellation", "created after the cancelled run" or "cancelled run started" text remains in the docs outside the round files. The design mentions supersession only at REQ-27/78, and its "rerun" mentions elsewhere (REQ-13, REQ-22, REQ-48, REQ-165) are unrelated or unchanged.

## Considered and not raised

- **Fixture (4) test-data strictness** (experiment-peer). With second-granularity REST timestamps, test data that sets the rerun's creation equal to R2's cancellation could let a naive implementation that uses a strict "<" check pass by accident [INFERENCE]. This concerns fixture construction in the implementation PR, not the ADR. Recommended practice for that PR: place both the rerun attempt's creation and its `run_started_at` strictly inside the window.
- **"Cancellation" as the cancel request or `completed_at`** (experiment-peer). The choice can widen the window by the cancel latency, and only for heads that are already superseded. The ambiguity dates from round 12, the fail-closed fallback covers it, and it is an implementation detail.
- **No fixture for a cancelled run that never started** (experiment-peer). An implementation using the cancelled run's creation as the lower bound passes all four fixtures. It only excludes queued heads that a newer head supersedes, which round 12 had recommended as legitimate, so REQ-27 is unaffected. Round 13 considered fixture weakness of this kind and did not raise it. There is no new evidence.
- **Whether GitHub keeps `created_at` across reruns.** This is search-summary evidence [INFERENCE]. If it is wrong, the new fallback clause ("cannot be distinguished from rerun-attempt time") fails closed, and the implementation must show the behaviour.
- **R-070 states a design consequence** ("the candidate window must use original run created_at"). Earlier rounds put similar sentences in that row, and R-077's caveat keeps architecture choices in ADR-012. Wording only.
- **The fallback addition departs from "keep the fallback sentence unchanged".** It fails closed and is recorded at `:48`. It is not a defect under the calibration rules.

## Peers and where they changed the outcome

- **`innovation-challenger`** reported zero issues and recommended closing both round-13 issues.
  - It confirmed the five identical rule mirrors, the three identical fixture mirrors, the ledger edits, STATUS next-free R-078, and that historical bodies differ from base only at `:3`.
  - Its rerun-mode analysis (full, failed-job-only and single-job) went beyond the lead's own check and confirmed that the coordinates hold for all three modes.
  - It did not change the verdict.
- **`experiment-peer`** reported zero issues.
  - Its discrimination table confirmed that fixture (4) fails the naive rerun-dating implementation on both evidence paths, and that no faulty implementation passes all four fixtures while excluding a final head.
  - Its test-data strictness note is recorded above as guidance for the implementation PR, not as an issue.
  - It did not change the verdict.
- **Lead-only:**
  - the check that a rerun cannot be the cancelling event for a qualifying new run;
  - the clock-vocabulary consistency check;
  - classifying the fallback addition as an accepted deviation;
  - the severity calls;
  - the GitHub `created_at` search.

## Verdict

**ADR-012: zero issues. Signed off.**

ADR012-R13-01 and ADR012-R13-02 are fixed at every named location. All earlier closures (R1–R12) hold, and no new issue was found.

**Sign-off statement.** The innovation council signs off ADR-012 (shadow-window readiness inputs) with zero open issues. The sign-off covers exactly these 13 signed files, at the SHA-256 digests that Main computed and verified with `shasum` (the council could not compute digests itself). All paths are under `docs/`.

- `adrs/2026-11-08-shadow-readiness-inputs/ADR.md` `a2869a076e0820cc13552ac9055cee22d0a57effb7d770c4501c9fb1c20343d8`
- `scout-design.html` `54d769cbcaec081a6764ef35f2898a91a3f839537ab7ea79e978763de5d7f587`
- `adrs/README.md` `cde415ff10b659a073d010f8a48c281d6f75a5a1cba466962d589e937b94587f`
- `STATUS.md` `17e6e4bf7f131111897600dcdbd17e5d51c13eace4f38b65cf7622357d3d563c`
- `plans/2026-11-06-remaining-work.md` `33afb05b4bd721e381475368e8354bc8e71aa24656e996a56819190f8154fe35`
- `references/policy-file.md` `7b747b02563fb3fdfa41914b99b9a7fa25a12546ad23fe98e922b99cccc80727`
- `references/ci-integration.md` `057594820d494584c004ba8d02a600fe1384196dcafa1f669500127aa1cf4ec4`
- `references/cli-commands.md` `a5ce50bab10d6c3348053e2321ce9e88499830780dd117ea0279e303b19d21e8`
- `research-ledger.md` `1ff9e7158b5619b41d1bb8a060d2b8ab9d58080a9c9620ba0524092f65313f77`
- `adrs/2026-10-19-v1-decisions/ADR.md` (ADR-000) `1cd521fd96439d87275ba353f7b05451c3e0c4d556ee665529aab85970ed81cb`
- `adrs/2026-10-23-decisions/ADR.md` (ADR-003) `973d24c4cfdfdd98acb4070a7bd72ca0c0e1ddf303bd4ccec8f92fe1cf117cc5`
- `adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) `ee593b6241814c24035671d666566bc2b7c399135867226c26e9626f3936abc8`
- `adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) `80e3375f81656a8856cf662cbec99a36d8e00bc451ff3c3592925d2293f44d95`

The unsigned `round-13-response.md` (`c6f6fdd1601adb152a72f95ad01dd4f7c235d211db0cf236eb37cfa3f869a703`) was reviewed as context.

**Limits of the sign-off:**
- It adopts the design only. It does not validate the implementation.
- Any change to these bytes needs a new zero-issue round.
- Adoption is recorded outside the signed files, as `ADR.md` (adopted-form route (a)) requires.
- Owner review under R-064 remains pending.

## Council memo

1. **Recommendation.** Sign off ADR-012 at the 13 digests above and record adoption outside the signed files.
   - Both round-13 fixes are present and identical at every mirror.
   - The only deviation, the added fallback clause, fails closed and is recorded.
   - Both peers independently found no new issue.
2. **Alternatives.**
   - **Recommended: sign off now.**
   - **B: require the fallback sentence to revert to round 13's exact wording.** This adds churn for no behavioural gain, because the added clause is a special case of "no qualifying run". Not recommended.
   - **C: add a fixture for a cancelled run that never started, or require strict-inside timestamps in fixture (4).** This would tighten implementation tests, but these are implementation-PR fixture details that the fail-closed fallback already covers. Not an ADR requirement.
   - **D: revert to round 12's creation lower bound.** This restores exclusion for queued supersessions but reopens an accepted fail-closed choice. Not recommended.
3. **Evidence.**
   - Worktree files: `ADR.md:17,19,21,48,57`; REQ-78 `:600`; REQ-27 `:612`; card `:1007–1012`; `ci:101,106`; `cli:126`; plan `:61`; `research-ledger.md:75–76`; STATUS `:131`.
   - Records: `round-12.md` (the round-12 recommendation); `round-13.md` (the R13-01 and R13-02 required fixes); `round-13-response.md`.
   - Searches: the full-sentence and full-paragraph mirror greps and the stale-wording greps described above.
   - GitHub: [workflow-runs REST](https://docs.github.com/en/rest/actions/workflow-runs) (search summary, [INFERENCE]); [re-running workflows and jobs](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/re-run-workflows-and-jobs); [contexts](https://docs.github.com/en/actions/reference/workflows-and-actions/contexts); [workflow-jobs REST](https://docs.github.com/en/rest/actions/workflow-jobs).
4. **Risks.**
   - Three provider behaviours are [INFERENCE]: that `created_at` is stable across reruns, how rerun concurrency cancels runs, and the exact cancellation timestamp. The implementation must demonstrate them, and the fallback assesses or gaps when proof is missing.
   - All timelines were constructed, not executed.
   - The council could not compute digests and relies on Main's `shasum`.
   - Residual risks accepted in earlier rounds are unchanged: author-forced interval resets, scanner self-falsification, provider limits, and loss of supersession exclusions.
   - The decision is document-only and reversible by a superseding ADR, and owner review under R-064 remains open.
5. **Execution outline.**
   1. Save this record as `round-14.md` in `docs/adrs/2026-11-08-shadow-readiness-inputs/`.
   2. Confirm with `shasum -a 256` that the 13 signed files still match the digests above before committing, and do not modify them.
   3. Record adoption outside the signed files. ADR-012 is effective on this round.
   4. In the implementation PR, satisfy every acceptance invariant at `ADR.md:56–62`. Build fixture (4) with the rerun attempt's creation and `run_started_at` strictly inside (R2 job start, R2 cancellation). Show provider rerun and timestamp behaviour, or show that the fallback assesses or gaps.
   5. Any later change to the signed bytes needs a new zero-issue council round.
