I found **11 open issues: 1 High, 6 Medium and 4 Low.** None of them is a demand for deployed proof; each is a design or interface gap a faithful implementation would hit. ADR.md and wiring-contract.json changed while I was reviewing, and round-2.md appeared, so this applies to the latest bytes I read: ADR.md at 19.1 KB, the 9.0 KB contract with `operation_grammar` and `start.receipt_schema`, and round-2.md. I did not repeat round-2 issues R2-01 to R2-08 except where noted.

## Issues

**F1 (High). A v1 collector job's attempts collide with ADR-012's attempt identity.**
- W5 has one trusted job reassess many observed PR heads ("Process oldest observable heads first").
- ADR-012 item 6 defines an attempt ID as (repository ID, run ID, run attempt, job ID), and item 4 makes `started_at` the job's start time.
- The schema-14 table is keyed `PRIMARY KEY(scope, repository_id, attempt_id, check_id)` (`internal/state/shadow_migration.go:14`). The readiness code stores attempts by attempt ID, `cohort[f.ID] = f` (`internal/shadow/observation.go:18-24`), and links join on `AttemptID`.
- So a second head assessed in the same job overwrites the first or violates the key. Tie-break ordering also becomes impossible, and ADR-012 says unverifiable ordering refuses readiness.
- ADR-012 item 1 also says "Collector reassessments only add attempts to existing observations", but v1 persists no PR-lane facts. W5 never says where v1 observations come from.
- Reassessment is currently disabled until a clock amendment, but W5 still approves this model, and the amendment is described as supplying only the clock.
- **Fix:** either take the reassessment model out of W5 and give it to the later amendment, or define a per-head attempt identity and how observations arise. The second option is an ADR-012 identity change and needs explicit amendment wording. It contradicts the ADR's "does not narrow ADR-012" claim as written.

**F2 (Medium). The shadow step can starve the existing assessment path.**
- The `shadow` step sits before artifact upload and enforcement, inside a job with `timeout-minutes: 60` (`scout-guardian-trusted.yml`).
- No step timeout and no reserved time at the end of the job are specified. If the job times out it is cancelled, and enforcement is then skipped because its condition is `!cancelled() && …` (contract `assessment_enforcement_if`). The job conclusion stops reflecting the assessment exit 0/1, which undoes W2's independence promise.
- Reconcile also pages runs, attempts and jobs with `GITHUB_TOKEN`. That token is limited to 1,000 requests per hour per repository ([GitHub REST rate limits](https://docs.github.com/en/rest/using-the-rest-api/rate-limits-for-the-rest-api)), the same budget the existing `assess --fetch-pr-metadata` uses on every push and nightly run.
- This applies even with zero reassessments.
- **Fix:** add a step `timeout-minutes` to the contract, a time reserve that keeps upload and enforcement alive, and an API request budget that leaves headroom for the assessment.

**F3 (Medium). What v1 writes when reassessment is unavailable is undefined.**
- W5 says "API facts already individually authenticated may still be recorded", but also says not to record "fictional historic PR-lane results".
- A `PRRunFact` needs head tree, merge base, check ID, rule version, comparison, inspection and pins (`internal/domain/shadow.go`). A PR-lane attempt seen only through REST has none of these.
- So it is undefined whether v1 persists REST-observed PR-lane attempts at all. That decides whether the 2-hour attempt deadline in W4 governs anything; otherwise it is dead configuration, the same objection round 1 raised against the lag number.
- **Fix:** list exactly which record types v1 writes (for example only the start marker, triage rows, refusals and diagnostics), and either say which attempts the deadline applies to or mark it inert.

**F4 (Medium). There is no collection horizon, and backlog vetoes are permanent.**
- Oldest-first processing has no lower bound tied to the start marker.
- Under ADR-012 item 5, every linked would_block event vetoes its cohort across all intervals. Reassessing pre-marker or abandoned PR heads therefore creates permanent vetoes from history the ADR itself says cannot certify.
- Oldest-first also leaves current PRs waiting behind the backlog.
- **Fix:** specify a horizon (for example runs created at or after the marker, or an explicit recorded horizon) and the processing order, in both the prose and the contract settings.

**F5 (Medium). Triage cannot keep up with v1's vetoes.**
- Each dispatch triages one event. The trusted concurrency group holds one pending run, and a newer run replaces it (R-103). Runs can take up to 60 minutes, and pushes and the nightly run also displace pending triage.
- Clearing N vetoes therefore needs N strictly serial dispatches.
- "Ship collection and triage together" does not make triage keep pace.
- **Fix:** allow batch triage (several event IDs, still one row per event per publication as ADR-012 requires), or cap collection to triage capacity, and disclose it.

**F6 (Medium). "Genesis" clashes with ADR-012's lineage term.**
- W5 says "no guessed genesis/frontier", and W6 says a later source must "establish its own attested interval/genesis".
- In ADR-012, the genesis publication ID defines lineage. A new genesis would make v1 history cross-lineage: fresh stamps, original authors re-authorized, and each import reopening events as Undecided. That would undo v1 triage.
- **Fix:** say that the later source continues the same lineage and sets a new interval origin or first frontier, not a lineage genesis.

**F7 (Medium). The base commit B for v1 reassessment has no defined source.**
- W5 lists "repository/PR/head/start/recipe bindings" but not B.
- ADR-012 item 2 requires persisting B, and new event IDs are built from merge_base(B,H) (item 7).
- If v1 takes B from a different source than a later complete collector, the same PR head can produce non-deduplicated vetoes. That source could be the current PR base or the REST run's `pull_requests[].base` [INFERENCE: whether REST reports event-time base SHA is unverified].
- **Fix:** specify B's authenticated source, its reachability check and the refusal when it is missing.

**F8 (Medium). Turning off persisted credentials exposes a latent restore bug, and the credential mechanism is not stated.**
- The restore step treats any non-zero `git ls-remote --exit-code` as "No scout/state branch yet" and still succeeds with `available=false`.
- With `persist-credentials: false`, a credential failure on a private repository becomes a silent full re-baseline attempt. `steps.baseline.outcome == 'success'` still lets `shadow` run, so W2's "unsafe restore suppresses" claim is false here.
- W3 also does not say whether the credential helper/GIT_ASKPASS lives in the workflow or in the binary.
- If it lives in the binary, an older binary also breaks existing restore and publish, not just the shadow commands. W6's mismatch statement covers only `shadow`.
- **Fix:** restore must fail closed on any exit other than 0 or 2. Name the mechanism and the steps it binds in the contract; this ties into R2-01(d). Then correct W6's old-binary statement.

**F9 (Low). Prose and contract disagree on the receipt.**
- W6 ¶2 says the receipt records "last intact v2 commit and migration result".
- The contract's `start.required_bindings` has `migration_state_commit` and neither of those fields.
- The ADR's own rule makes any disagreement a gate failure.

**F10 (Low). The pre-triage step whitelist is incomplete.**
- W1 omits the existing `actions/cache@v4` "Restore the scanner toolchain cache" step and the new normalization step.
- Both precede triage/start, so P25 ordering checks would flag the real workflow.

**F11 (Low). Some companion documents are stale or cite incompletely.**
- `ci-integration.md:79,87` still say "Fail the job on exit 2 or 3" without scope; the R2-05 fix landed only at :107.
- `cli-commands.md:136` still says "The implementation PR chooses collection and triage commands… No … shadow writer/scan command is adopted here", directly above ADR-015's reserved commands.
- The design's Appendix B card for ADR-015 cites R-100/R-101 only, while the ADR cites R-100–R-105 (grep found no R-102–R-105 in `scout-design.html`).

## Round-2 items not yet reflected in the current contract
- R2-01(c)(d)(e)(f)(h)(j) are still absent: the normalization step object, the per-step map of `GITHUB_TOKEN` and other secrets, the exit map as data, the event-ID regex, the artifact glob grammar, and the pinned `timeout-minutes` and trusted-job `if`.
- Severity stays as round 2 set it.

## Checked and found correct
- **Baselines:** the trigger, cron, concurrency, permission and `v0.1.6` baselines match the workflows.
- **Release artifacts:** the artifact list matches `release.yml`.
- **Policy keys:** `min_days`/`min_pr_runs` 14/5 match `internal/policy/policy.go:17,92-93`.
- **Old-binary exit:** an older binary does give exit 3 for `scout shadow …`, because unknown commands return `exitUnavailable` (`cmd/scout/main_test.go:47-48`).
- **Pre-start exit:** the pre-start reconcile code 2 (`not_started`) now resolves R2-05's caller-breakage concern.
- **Ledger rows:** R-104 and R-105 now exist.
