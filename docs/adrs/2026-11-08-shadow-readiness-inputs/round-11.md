# Round 11: ADR-012 (shadow-window readiness inputs)

Date: 2026-11-08. This round applies design rules 1–3 and the ledger rules in `CLAUDE.md`, plus the ADR rules in `docs/adrs/README.md`. ADR-012 is an agent decision made under the owner's standing delegation (R-064). Where a fix involves a choice, the council names the option it recommends under that delegation.

The round reviews the bytes revised after round 10 (`round-10-response.md`). It has three parts:

1. Check every ADR012-R10-01..04 fix against its required fix.
2. Confirm that every earlier closure (R1–R9) still holds in the current bytes.
3. Look for new issues, numbered ADR012-R11-xx.

The council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review. It looked for false-ready paths and for readiness that a compliant repository could never reach. It checked whether the R10 fixes introduced contradictions in six areas: the CAS-parent anchor, the (parent C, new C] range, same-commit advances, the equality fixtures, the superseding-run condition and the deadline-visibility fixture. It also checked mirror drift, the ledger, the historical ADR bodies and README.
- **`experiment-peer` (Opus):** methodological review. It built concrete timelines for each revised rule and fixture: parent and new C, starts, points, stamps, cutoffs, retry races and GitHub concurrency semantics.

The review was read-only. Nobody edited files, ran product code, made model calls outside the council, or touched `.codegraph`.

- The lead read the required codegraph skill references (`readiness-and-bootstrap.md`, `query-selection.md`). Codegraph was not used, because the material under review is documents and the council had no shell. Searches used `read`, `grep` and `glob`.
- The lead read one GitHub page: [control-workflow-concurrency](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency).
- The experiment-peer cited that page and the [concurrency-groups REST page](https://docs.github.com/en/rest/actions/concurrency-groups).
- The challenger cited no external source.

**Main computed and verified the digests below with `shasum`. The council did not, because the lead and both peers had read-only tools and no shell.** This round is not a sign-off, so the digests only identify the bytes reviewed. All paths are under `docs/`.

| File | SHA-256 (Main) |
|---|---|
| `adrs/2026-11-08-shadow-readiness-inputs/ADR.md` | `1a8c8be6d2e9599d095dd03b86fd28adee6f42e9da8f1d8daf9c29f183609d44` |
| `scout-design.html` | `fbb6687b0b84bb9884f262b2f678d9189a903bb052a27d6b51341848e46ea42d` |
| `adrs/README.md` | `cde415ff10b659a073d010f8a48c281d6f75a5a1cba466962d589e937b94587f` |
| `STATUS.md` | `17e6e4bf7f131111897600dcdbd17e5d51c13eace4f38b65cf7622357d3d563c` |
| `plans/2026-11-06-remaining-work.md` | `ef106d3472888eec64270b2430ecc97ce5a293ebfb404bd07edaf9fa2152e49b` |
| `references/policy-file.md` | `7b747b02563fb3fdfa41914b99b9a7fa25a12546ad23fe98e922b99cccc80727` |
| `references/ci-integration.md` | `c0697ed05cbed0a51eb77aa7032a4bb53f4d2f81f6a4d0949867b6e3f7c1452b` |
| `references/cli-commands.md` | `6319b00ceffac1ce74968413349ec28c75e17bf588e0115b63f49c191dd26071` |
| `research-ledger.md` | `442ce78f1660cd86381cba031f25d375851f38d5407468896106803e7b54f199` |
| `adrs/2026-10-19-v1-decisions/ADR.md` (ADR-000) | `1cd521fd96439d87275ba353f7b05451c3e0c4d556ee665529aab85970ed81cb` |
| `adrs/2026-10-23-decisions/ADR.md` (ADR-003) | `973d24c4cfdfdd98acb4070a7bd72ca0c0e1ddf303bd4ccec8f92fe1cf117cc5` |
| `adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) | `ee593b6241814c24035671d666566bc2b7c399135867226c26e9626f3936abc8` |
| `adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) | `80e3375f81656a8856cf662cbec99a36d8e00bc451ff3c3592925d2293f44d95` |

Main also computed the SHA-256 of one unsigned record:

| File | SHA-256 (Main) |
|---|---|
| `round-10-response.md` | `7df0bcaa8291b5f212dd604d59877fe99819fdd0f825fd56535e14d8c779bd5b` |

README, STATUS and the four historical ADR files have the same digests as in round 10. The changed files are ADR.md, the design, the plan, the policy-file, ci-integration and cli-commands references, and the ledger.

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11`, branch `agent/policy-shadow-adrs`, based on origin/main `7bb73e52`. Every change is staged and uncommitted.
- Base bytes come from `/Users/adamjackson/Projects/work/scout-adr11b`.

**Files read.** Line numbers below refer to the raw worktree files.
- ADR-012 `ADR.md`, all lines:
  - `:16–22` population and the clocks/frontier rule;
  - `:26–30` triage, history import and ShadowReady;
  - `:34–40` persistence and writer boundary;
  - `:44` CLI and refusal scope;
  - `:48–52` amendments, sources and route;
  - `:56–62` acceptance invariants.
- `round-10.md` and `round-10-response.md` (all); `round-9-response.md:10,22` (grep context).
- `scout-design.html` raw: REQ-78 `:600`, REQ-27 `:612`, REQ-48 `:615` and the ADR-012 card `:1009`; grep context for REQ-22 `:589`.
- `ci-integration.md:95–108`, `cli-commands.md:124–136` and `policy-file.md:76–86`, all raw.
- Plan `:59–62`; STATUS `:7,9,121,131`; `research-ledger.md:75–76`.
- `.github/workflows/scout-guardian.yml:1–60`, whose concurrency block is at `:41–43`.
- The challenger compared the bodies of the four historical ADRs and REQ-160/162/163/168 (`:706,708,709,714`) with base.

## Issue summary

**2 distinct issues are open, both minor. None is major.**

| ID | Severity | Summary |
|---|---|---|
| ADR012-R11-01 | minor | The superseding-run and pinned-configuration conditions apply only to the inferred branch. The provider-recorded-cause branch can still exclude a head whose cancellation came from a PR-added workflow or a PR-edited guardian. No fixture fails an implementation that skips the pinned-equality check |
| ADR012-R11-02 | minor | The ADR's Amendments section credits round 9 with the round-10 superseding-run decision; round 9 actually decided the opposite |

## Closure of round-10 issues

"Fixed" means the required document fix is present in the reviewed bytes at every named location. It does not mean the implementation is validated. "Fixed as prescribed" means the council's own prescription was applied but left a gap, which is carried as a new R11 issue.

| Issue | Status | Evidence |
|---|---|---|
| ADR012-R10-01 | **Fixed** | **(a) Combined fixture replaced.** Plan `:61` and `ci:106`: "At historical day30, independently assert that a same-pin/different-inspection late rerun preserves the representative, and a different-pin late rerun preserves the default version". This matches `ADR.md:58`. Grep finds no "changed pin/inspection" or "stamp-gated" anywhere in `docs/` outside the historical round files.<br>**(c) Both halves of the deadline fixture.** `ADR.md:57`, plan `:61` and `ci:106` all keep "late and invisible before their stamp" and add: "With C≥terminal stamp and no refusal trigger, assert the late terminal attempt and its linked would_block event are visible, yielding computed exit 0 not-ready/Undecided, paired with an otherwise identical ready control without the late terminal facts."<br>**Peer timeline:** the attempt starts at 10 and is deadline-gapped; terminal facts are published from parent C 20 with stamp 25. Below the stamp, the attempt and its event W are invisible. At C ≥ 25, W vetoes cohort-wide, because the veto covers events "whatever its interval, gap or late status" (`ADR.md:20`). An implementation that drops the facts behaves like the control and fails. |
| ADR012-R10-02 | **Fixed** | `ADR.md:58`, plan `:61` and `ci:106`:<br>• "parent C<point<new C<stamp … selects the interval starting strictly after the point; the previous interval ends strictly before it";<br>• "parent C<point=new C<stamp … no complete containing interval (unmeasured/exit 2); a later cutoff after the point selects the next interval".<br>Both cases agree with the exclusive boundaries at `:19` and the exit mapping at `:44`. The equality case fails an inclusive-boundary implementation; the strict case fails one that anchors lateness on new C (peer). |
| ADR012-R10-03 | **Fixed** | **Definition.** `ADR.md:19` and REQ-78 `:600`, verbatim: "Destination C at publication means the frontier in that publication's CAS parent, the same parent C used by the stamp formula, recomputed on retry; an advance in that commit never makes its own facts or points late."<br>**Range.** "For points in (parent C, new C], advancement requires a published normal gap … points≤parent C (including equality) are late and never block advancement."<br>**Mirrors:** `cli:126`, `policy-file:82` and `ci:101`, which adds "late iff point≤that parent C".<br>**Fixtures** (same-commit and equal-to-parent-C): `ADR.md:58`, plan `:61`, `ci:106`.<br>**Peer re-run of the R10 timeline:** parent C 30.0, new C 31.95, stamp 32.0. There is now one reading: A is normal and visible, and certification is `current`. In the retry race, every published gap point above the current parent C is necessarily normal, and lateness flags are written once in the final CAS commit, so `dossier_digest` is unique. |
| ADR012-R10-04 | **Fixed as prescribed** | **Rule.** `ADR.md:17`, REQ-27 `:612` and REQ-78 `:600`: "its executed lane configuration verifiably equals the pinned protected-default configuration and evaluates cancel-in-progress=true for the same evaluated concurrency group as the cancelled run … The cancelled run's own setting is irrelevant to this inference; unknown or PR-edited configuration is assessed/gapped." Mirrors at `ci:101` and `cli:126`.<br>**Fixtures:** `ADR.md:57`, plan `:61` and `ci:106` (R1 true / R2 false, unverifiable R2, all false).<br>**Ledger:** R-070 (`research-ledger.md:75`) links workflow-syntax#concurrency and the concurrency concepts page, and records the reading as [INFERENCE] needing provider confirmation.<br>**Carried:** the new conditions are confined to "this inference", leaving the provider-cause branch open, and the prescribed fixtures cannot detect a missing pinned-equality check (ADR012-R11-01). |

## Re-confirmation of earlier closures (R1–R9)

| Closure | Holds? | Evidence |
|---|---|---|
| R1-04: started_at membership; elapsed-days formula | Yes | `ADR.md:19–21` |
| R1-05, R7-01, R8-07: recoverable frontier, deadline point gaps | Yes, and strengthened | `ADR.md:19`. The explicit (parent C, new C] range removes the equality-deadlock reading |
| R2-05, R6-01/08, R7-04, R8-05, R9-04: population, exclusions, causal supersession | Yes, apart from ADR012-R11-01 | `ADR.md:16–17`; REQ-27/78; `ci:101`; `cli:126` |
| R5-09: scanner-output failure is comparison-none | Yes | `ADR.md:17` |
| R1-08/09/11/14, R2-10, R3-01, R4-07/10: CI/local/legacy separation, migration, v2 cutover, query-only v13 | Yes | `ADR.md:34–38,44`; REQ-48 `:615` |
| R3-02/07, R4-02/08/11, R8-02, R9-02/03: authorization, current-tip authority, replay, writer isolation, lineage, import refusal | Yes | `ADR.md:26,28,36,40`; REQ-48 `:615`; `cli:130`; `policy-file:82` |
| R3-03, R4-06: guardian comparison, empty override inputs | Yes | `ADR.md:17–18`; REQ-27 |
| R2-11, R3-05, R4-03/04, R5-03/05/07, R6-02/07, R7-01, R8-03, R9-01: interval selection, cutoffs, frontier minima, dossier invariance, single predicate | Yes | `ADR.md:19–20,28,44`; REQ-78; `cli:126–132`. R10-03 removed the remaining ambiguity |
| R8-04: points are exclusive interval boundaries | Yes | `ADR.md:19`. The fixture now agrees with it (R10-02) |
| R5-06, R6-03, R8-01: liveness, skewed-clock triage, scoped refusal | Yes | `ADR.md:19,44`. A commit whose bound minus lag reaches its own stamp clears its own late records (peer) |
| R5-01, R6-10: wiring-ADR gate | Yes | `ADR.md:12,56`; `ci:99`; plan `:61`; `policy-file:84`; `cli:134` |
| R5-04, R6-04, R7-02, R8-06, R9-05: cohort-wide zeroes including provisional; fixtures cannot pass by refusal alone | Yes | `ADR.md:20,30,57–60`; REQ-22 `:589` |
| R5-08, R6-05: Days, representative origin | Yes | `ADR.md:21,48,62`; card `:1009` |
| R5-10, R7-05/06/07: exact cohort, default version, dossier encoding, attempt-ID order, pin tuples | Yes | `ADR.md:21,44,61–62`; REQ-78 |
| R7-03: event-to-attempt binding | Yes | `ADR.md:17`; REQ-27/78 |
| R7-08/09: mirror drift; Activation callout | Yes | Plan `:61` and `ci:106` now carry every ADR fixture. Card `:1009`, STATUS `:9,121` and REQ-48 `:615` agree with the ADR |
| Historical ADRs and untouched requirements | Yes | ADR-000/003/007/008, README and STATUS digests are unchanged from round 10. The challenger's raw comparison found only the `:3` status lines differ from base. REQ-160/162/163/168 match base |
| Ledger | Yes | R-077 (`research-ledger.md:76`) links `round-10.md`. R-070 adds the concurrency links. STATUS `:131` names next free R-078. No R-ID or ADR ID is reused |

## Are the invariants testable and enough for REQ-22, REQ-27 and REQ-78?

| Requirement need | Testable and sufficient? | Where / gap |
|---|---|---|
| REQ-22: counting unit, deduplication, exact cohort, unknown ≤ 10 % | Yes | `ADR.md:20–22` |
| REQ-22: 14 days, inspected origin, `as_of` endpoint | Yes | `ADR.md:19,21,58,62` |
| REQ-22: zero `not_warranted` and zero Undecided across gaps, late and imported evidence | Yes | `ADR.md:20,57–60` |
| REQ-22: maintainer-only, append-only, idempotent triage | Yes | `ADR.md:26,28,60` |
| REQ-27: population independent of content, author and approval | **Partly** | ADR012-R11-01 |
| REQ-78: complete-interval accounting, recoverable frontier | Yes | `ADR.md:19,58` |
| REQ-78: historical reproduction, `dossier_digest` | Yes | `ADR.md:19,44,58,61` |
| REQ-78: current certification and liveness | Yes | `ADR.md:19,44,58,60` |
| Governance of deferred wiring | Yes | `ADR.md:12,56` |

## Issues

### ADR012-R11-01: The provider-recorded-cause branch skips the superseding-run checks, and no fixture tests pinned equality (minor)

- **Location:**
  - `ADR.md:17`: "pre-assessment cancellation caused by supersession in the same pinned lane/concurrency group for the same PR (provider-recorded concurrency cause, or the superseding run was created before cancellation with no actor-initiated cancel recorded, and its executed lane configuration verifiably equals … ) … The cancelled run's own setting is irrelevant to this inference; unknown or PR-edited configuration is assessed/gapped."
  - The same clause in REQ-27 `:612` and REQ-78 `:600`.
  - `ci:101`: "Supersession needs provider concurrency cause or a superseding run created before cancel …".
  - `cli:126`: "Inferred supersession additionally requires …".
  - Fixtures: `ADR.md:57`, plan `:61`, `ci:106`.
- **New evidence:**
  - GitHub, [control-workflow-concurrency](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency): "If you have multiple workflows in the same repository, concurrency group names must be unique across workflows to avoid canceling in-progress jobs or runs from other workflows. Otherwise, any previously in-progress or pending job will be canceled, regardless of the workflow." The same page says job-level `jobs.<job_id>.concurrency` uses the same groups.
  - The guardian group is `scout-guardian-${{ github.repository }}-${{ github.ref }}` (`scout-guardian.yml:41–43`). It contains no workflow identity.
  - `pull_request` runs execute the PR's merge context (R-070), so a PR can add workflows and edit the guardian.
  - The concurrency-groups REST API lists only live members (experiment-peer). A recorded concurrency cause therefore does not by itself identify the superseding run [INFERENCE].
  - The R10-04 fix scoped the new conditions to "this inference". That makes the open reading of the first branch explicit; before round 10 it was only implicit.
- **Problem** (experiment-peer; the lead confirmed the documentation) [INFERENCE: constructed timelines, not executed]:
  - **Timeline A (today's pinned `true`):**
    1. Push H1 adds `.github/workflows/x.yml` on `pull_request`. Its job waits about 60 s, then joins the group with `cancel-in-progress: true`.
    2. The job cancels guardian run R1 before assessment, and GitHub records a concurrency cause.
    3. "Same … concurrency group for the same PR" with a "provider-recorded concurrency cause" is satisfied. The branch never checks which run cancelled R1.
    4. The final head H1 is excluded, with no H2, no gap and no manual action.
    - A PR that hard-codes another PR's `refs/pull/M/merge` group string could do the same to that PR's guardian run.
    - Excluding the head hides whatever its assessment would have produced, so an excluded head can never veto. That is the R8-05/R10-04 harm class.
  - **Timeline B (a future pinned `false`):**
    1. R1 runs with the pinned `false`.
    2. H2's PR edit makes R2 evaluate `true`.
    3. R2 cancels R1, and GitHub records the cause.
    - Because "unknown or PR-edited configuration is assessed/gapped" follows "irrelevant to this inference", it can be read as scoped to the inferred branch. Under that reading the provider-cause branch excludes H1, so the rule gives two answers.
    - "Only provider-proven guardian non-assessment" and "Causes must be independent of outcome/content" point to the fail-closed reading. Even so, the text does not give a single answer.
  - **Fixture gap:** an implementation that tests only "R2 evaluates true" and skips "equals the pinned configuration" passes all three fixtures (R1 true / R2 false, unverifiable R2, all false). It then wrongly excludes H1 when the pinned setting is `false` and the PR edits R2 to `true` with no canceller recorded. That is exactly the gaming case the pinned-equality clause exists for.
- **Severity:** the lead adjudicates minor, the same as R10-04.
  - The defect is an open reading next to fail-closed text, not a mandated false-ready path.
  - The cancellation shows as a cancelled guardian check on the PR.
  - It requires a deliberately adversarial workflow.
- **Required fix (recommended option):**
  - **Rule:** in `ADR.md:17`, REQ-27 `:612`, REQ-78 `:600`, `ci:101` and `cli:126`, state supersession as one condition. Exclusion requires an identified superseding run of the pinned guardian lane for the same PR key, created before the cancellation. Its executed configuration must verifiably equal the pinned protected-default configuration and evaluate `cancel-in-progress=true` for the same evaluated concurrency group. In addition there must be either a provider-recorded concurrency cause or no actor-initiated cancel recorded. The provider cause stands in only for the "no actor-initiated cancel" condition. A cancellation by any other workflow or job, by an unidentified canceller, or by a PR-edited superseding run is assessed or gapped.
  - **Fixtures** (`ADR.md:57`, plan `:61`, `ci:106`):
    1. A PR-added workflow job joins the guardian group with `cancel-in-progress: true` and cancels H1's guardian run before assessment. A provider cause is recorded and there is no H2. H1 is assessed or gapped.
    2. The pinned setting is `false`, and H2's PR edit makes R2's executed configuration `true` in the same group. R2 cancels R1 before assessment with no canceller recorded. H1 is assessed or gapped.
    3. Fixture 2 again, with a provider-recorded concurrency cause. H1 is assessed or gapped.
    - Keep the existing three fixtures.
  - **Ledger:** add the control-workflow-concurrency link and the quoted cross-workflow sentence to R-070, which is ADR-012's own unadopted row.
  - Record the decision under round-11 delegated decisions in `ADR.md:48`.

### ADR012-R11-02: The Amendments section credits round 9 with the round-10 decision (minor)

- **Location:** `ADR.md:48`: "Round-9 delegated decisions: … supersession's inferred-cause branch requires the superseding run's verified executed/pinned configuration to evaluate cancel-in-progress=true."
- **Problem** (the lead and the experiment-peer found it independently):
  - Round 9 decided the opposite. `round-9-response.md:10` says "Condition the inferred supersession branch on the cancelled run's pinned cancel-in-progress setting", and `:22` says the same.
  - The superseding-run rule is round 10's decision, and the same paragraph restates it as such.
  - The operative rule is unaffected. But the ADR summary must record what was decided and what informed it (design rule 2), and an accepted ADR can never be edited again. A false provenance statement adopted now would therefore be permanent.
- **Required fix:** in the round-9 list in `ADR.md:48`, replace that clause with "supersession's inferred-cause branch was conditioned on the cancelled run's pinned cancel-in-progress (replaced in round 10)". Leave the round-10 clause as the statement of the current rule. No mirror changes are needed.

## Checks that found no issue

- **Same-commit advance and retry** (both peers). Lateness against the CAS parent gives one answer. On retry, a fact that becomes late is either already published, and so deduplicated, or deadline-gapped, and so correctly late. A duplicate gap point above the earlier frontier is late and gated by its stamp, so dossiers do not change.
- **The (parent C, new C] range.** A published gap whose point lies above the current parent C cannot be late, because C only increases. So "normal" in the corollary is never in doubt. The equality fixture fails both an inclusive-range implementation, which deadlocks, and one that shows the point immediately, which changes the dossier at parent C.
- **Feasibility of the deadline veto fixture** (both peers). Linked-event vetoes span gaps (`ADR.md:20`), so a qualifying post-gap control makes the paired fixture feasible.
- **The day-30 rerun fixtures.** Because the reruns arrive through a pending import, the leak can be seen; an earlier deadline gap would have hidden it (peer).
- **Mirrors.** REQ-27/78, `ci:101`, `cli:126`, `policy-file:82`, plan `:61`, `ci:106`, card `:1009` and REQ-48 `:615` agree with the ADR on every R10 change.

## Considered and not raised

- **Which "pinned protected-default configuration" applies when the base changes the guardian between runs** (lead, peer). Under every reading the result either matches GitHub's behaviour or gaps the run, which fails closed. This is not a defect.
- **"points≤parent C … are late" read as relabelling an earlier normal gap at that point as late** (lead). The clause's only effect is "never block advancement". Visibility stays keyed to each record's own publication, so outcomes do not change. This is wording.
- **"visible at the enabled frontier" in the mirrors versus "every cutoff the advance permits" in the ADR** (lead, peer). A normal point is visible at every cutoff at or after it, so the two are equivalent. This is wording.
- **The "Round-9 fixtures:" label in front of the round-10 split fixtures in plan `:61` and `ci:106`** (peer). It is a label only; the fixtures bind either way.
- **The same-commit fixture does not fix the order of started_at and the point** (peer). This is a test-design choice, and visibility can still be observed.
- **An unverifiable superseding configuration deflating exposure** (challenger). It fails closed. Round 10 already weighed this deflation risk and rejected the more fail-closed Alternative C; no new evidence reopens it.
- **GitHub's `queue` property** (lead, peer). `queue: max` cannot be combined with `cancel-in-progress: true`, and the default `single` mode replaces pending runs, which the rule already gaps when there is no cause. No effect.
- **README not byte-identical to base** (challenger). This is expected. Its digest is unchanged since round 10, which reviewed its ADR-012 route and index amendments.
- **"terminal/published" read as "terminal or published" in the retry race** (peer). This predates R10, and "discharged by a published point-gap record" supports the stricter reading.
- **Requiring separate collection and frontier commits** (challenger). Not needed under the explicit parent-C anchor.

## Peers and where they changed the outcome

- **`innovation-challenger`** reported zero issues and rated R10-01..04 fixed. Its evidence matched the lead's for every closure. It found no new issue.
- **`experiment-peer`** raised 3 minor findings. It rated R10-01/03 fixed and R10-02/04 fixed as prescribed.
  - N1, the provider-cause branch gaming, became ADR012-R11-01. **This was new to the lead and changed the verdict from zero issues to issues open.** The lead verified the cross-workflow documentation sentence and the guardian group expression.
  - N2, the fixture that cannot detect a missing pinned-equality check, was folded into ADR012-R11-01 because it is the same clause, the same mirrors and the same harm.
  - N3, the misattribution, became ADR012-R11-02. The lead had found it independently.
  - Not adopted: R10-02 as "fixed as prescribed". No gap is carried, so it is rated fixed.
- **Lead-only:**
  - the cross-PR group variant in ADR012-R11-01;
  - the R-070 obligation for the new documentation page;
  - the permanence argument for ADR012-R11-02;
  - all severity calls.

## Verdict

**ADR-012: 2 issues open (0 major, 2 minor). Not signed off.** Open: ADR012-R11-01 and ADR012-R11-02.

Round-10 closure:
- R10-01, R10-02 and R10-03 are fixed.
- R10-04 is fixed as prescribed. Its remaining gap is carried as ADR012-R11-01.

All earlier closures hold. The population and causal-supersession group is qualified by ADR012-R11-01.

No digests are signed in this round. The 13 digests above identify the bytes reviewed.

## Council memo

1. **Recommendation.** Do not sign off. Apply the 2 fixes using the recommended options, then submit round 12 with fresh digests from Main for all 13 signed files and the response.
   - The R10 fixes closed the frontier, boundary and fixture defects cleanly. The only open problem is in the supersession clause: the R10 scoping left the provider-cause branch and the gaming fixture unaddressed. The other open issue is one provenance sentence.
2. **Alternatives.**
   - **R11-01:**
     - Recommended: one condition that requires an identified, pinned-configuration superseding run of the same PR, with either the provider cause or no recorded actor cancel.
     - Alternative B: drop the provider-cause branch and keep only the inferred branch. This is equivalent in effect but loses the explicit cause evidence. Acceptable, not preferred.
     - Alternative C: exclude only when the provider identifies the cancelling run. This is the most fail-closed option, but GitHub's documented APIs do not appear to expose that [INFERENCE], so in practice it gaps all supersession. Not recommended, for the deflation reason given in round 10.
   - **R11-02:**
     - Recommended: correct the attribution.
     - Alternative: delete the clause from the round-9 list. Acceptable.
   - **Signing off with known minor issues** is not permitted under design rule 1.
3. **Evidence.**
   - `ADR.md:17,19,20,48,57–58`; REQ-27 `:612`; REQ-78 `:600`; `ci:101,106`; `cli:126`; `policy-file:82`; plan `:61`; `research-ledger.md:75–76`; `round-9-response.md:10,22`.
   - `.github/workflows/scout-guardian.yml:41–43`.
   - GitHub [control-workflow-concurrency](https://docs.github.com/en/actions/how-tos/write-workflows/choose-when-workflows-run/control-workflow-concurrency) (cross-workflow cancellation; job-level groups) and the [concurrency-groups REST API](https://docs.github.com/en/rest/actions/concurrency-groups) (peer).
4. **Risks.**
   - Whether GitHub records or identifies the cancelling run is [INFERENCE]; the implementation must confirm provider behaviour.
   - All timelines were constructed, not executed.
   - The council could not compute digests.
   - Every fix is a document change and fully reversible. Implementation is still pending.
   - Residual risks accepted in earlier rounds (author-forced interval resets, scanner self-falsification, provider limits) are unchanged.
5. **Execution outline.**
   1. **R11-01:**
      - In `ADR.md:17`, REQ-27 `:612`, REQ-78 `:600`, `ci:101` and `cli:126`, replace the two-branch supersession text with the single condition above. State that cancellation by another workflow or job, by an unidentified canceller, or by a PR-edited superseding run is assessed or gapped.
      - Add the three fixtures to `ADR.md:57`, plan `:61` and `ci:106`.
      - Add the control-workflow-concurrency link and fact to R-070.
      - Record the round-11 delegated decision in `ADR.md:48`.
   2. **R11-02:** correct the round-9 clause in `ADR.md:48`.
   3. Extend R-077's list of rounds with `round-11.md`. No new ledger ID is needed.
   4. Write `round-11-response.md` mapping each issue. Main supplies the SHA-256 of all 13 signed files and of the response, then round 12 follows.
