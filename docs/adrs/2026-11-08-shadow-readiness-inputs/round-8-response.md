# ADR-012 round 8 response

Date: 2026-11-08. Responds to `round-8.md`; no sign-off is claimed. Every recommended fix is adopted under standing owner delegation R-064, subject to round 9. No historical round or earlier response is changed.

## Delegated consolidations

1. Adopt council option A: a single scoped refusal and live certification status, not a refusal on every publication. Activation requires `current_window`/`current`; historical output cannot bypass it. Publication/encoding mechanisms remain implementation choices.
2. Replace cross-lineage ordering inference with the existing Undecided veto until a new destination verdict. Identical-prefix same-lineage restores retain trusted stamps. Keep only authorization, ordering, idempotence and historical-visibility invariants; import scheduling belongs to implementation.
3. Use one visibility predicate for every query and make every point gap end an interval. No per-query exceptions or puncture semantics. Unknown provider causes fail closed as assessed/gapped; approval waiting has no started job and deadline-expired jobs are gapped.
4. Replace repeated evolving fixture inventories in the plan/CI reference with the canonical ADR acceptance contract plus the explicit round-8 cases. Retain every prior acceptance invariant. Lag, deadlines, reconcile cadence and encoding mechanics remain implementation obligations, not selected mechanisms.

These decisions are also recorded in ADR.md’s Amendments section and mirrored in the design card. This is a design-only revision; no implementation or experiments were run.

## Issue mapping

| Issue | Fix and locations | Required acceptance obligation |
|---|---|---|
| ADR012-R8-01 | ADR clocks/interface and REQ-78 scope refusal to repository late records after C, queried-cohort authorized history after C, and pending local late facts/imported history. Ordinary later attempts/frontier/metadata do not refuse. Live certification is current/refused/historical outside the digest; only current-window/current can activate. Refusal is unmeasured/2 regardless of computed result; unsupported encoding is 3. Stamps cease refusing at C≥stamp and quiet reconcile must eventually pass them. CLI/policy, STATUS and plan mirror this. | Last-Undecided triage refuses until stamp then becomes ready; frontier-only reconcile computes; unrelated-check triage does not refuse; held-C late record refuses while explicit C shares digest with historical status. |
| ADR012-R8-02 | ADR history/persistence, REQ-48/78 and CLI/policy adopt original-stamp identical-prefix fast-forward (including sealed local restore). Other-lineage rows deduplicate by action ID and publish successively, one row/event/publication in original stamp/numeric-action order, retaining provenance and requiring original plus current-tip author authorization. Invalid authorization or multi-lineage history leaves event Undecided until a new authorized destination verdict after import. | Both day25/day32 verdict orders remain not ready until new verdict; multi-row import respects publication invariant; repeat import is no-op; fast-forward dossiers equal fresh restore. |
| ADR012-R8-03 | ADR clocks defines late against destination C at publication. One normal-coordinate/late-stamp/pending-invisible predicate applies to default version, in-scope attempts, representatives, links, vetoes, gaps and Days origin; REQ-78 and CLI mirror it. | Same historical day30 command without version has identical cohort/digest before/after pending import and after publishing later-started late rerun with changed pin/inspection. |
| ADR012-R8-04 | ADR clocks/counting and REQ-78 make each point an exclusive interval boundary; attempts at it belong to no interval. Every accounted event_at and started_at receives its point. | Day20 point makes day35 interval and Days origin start after day20. |
| ADR012-R8-05 | ADR population, REQ-27/78 and CI/CLI require causal supersession: provider concurrency cause or superseding same-lane/group/PR run created before cancellation with no actor cancel recorded. Otherwise assess/gap. | Manual cancellation followed by later push remains assessed/gapped. |
| ADR012-R8-06 | ADR acceptance, B4-S and CI explicitly require C≥all relevant stamps, no other refusal trigger, computed exit 0 not-ready with named veto, and paired ready control for late-event/import/correction tests. Cross-lineage import uses Undecided, not a resurrected imported verdict. | A refusal alone cannot pass any veto fixture. |
| ADR012-R8-07 | ADR clocks, REQ-78 and policy define attempt start as provider job start; approval waiting has no started attempt and follows event accounting/gaps. Non-terminal job at chosen deadline receives point gaps discharging completeness; later terminal facts follow late lifecycle. | Approval-pending fork does not hold C; deadline-expired job can discharge completeness without waiting for terminal provider status. |

## Signed-set consistency and evidence

- R-077 now links round 8. R-070 is unchanged: this revision does not cite or rely on the GitHub expiry duration/page; it binds provider job-start and conservative gap invariants instead. No ledger ID is added.
- The ADR index and four historical ADR status amendments remain applicable without further changes; accepted historical bodies remain untouched.
- Implementation remains pending everywhere. No checks/builds/tests/formatters were run; the main agent should submit the staged signed files and this response for round 9 review. Requested hashes and staged diff statistics are supplied with the handoff, not signed here.
