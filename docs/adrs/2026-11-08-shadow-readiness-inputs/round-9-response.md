# ADR-012 round 9 response

Date: 2026-11-08. Responds to `round-9.md`; not a sign-off. Adopt the council’s recommended option A under R-064, subject to round 10. Historical reviews/responses are unchanged.

## Delegated decisions

- Keep one coordinate/late-stamp visibility predicate, classifying every gap point independently. Delete the blanket gap-stamp rule, rather than adding another refusal mechanism. A normal published gap must be visible at the frontier it enables; late points cannot rewrite sealed dossiers.
- Define lineage as the authenticated trusted publication chain. Require proof of the entire published-history prefix, not only sealed history, before preserving stamps. Default everything else to cross-lineage authorization and per-event Undecided; do not infer trust from a claimed lineage ID.
- Use the recommended event-level stamped import-refusal record, not whole-import quarantine. New accepted/refused rows reopen only their event; resolution must follow its last import publication.
- Condition the inferred supersession branch on the cancelled run’s pinned cancel-in-progress setting. Missing causal evidence remains assessed/gapped.
- Replace ineffective combined-pin and unconditional deadline-lateness fixtures with independent, observable assertions; retain no fixture whose target is satisfied merely by refusal or permanent Undecided. Implementation scheduling/mechanisms remain outside the ADR.

These decisions are recorded in ADR.md’s Amendments section. No product code, experiments, checks, builds, tests or formatters were run.

## Issue mapping

| Issue | Fix | Fail-capable acceptance obligation |
|---|---|---|
| ADR012-R9-01 | ADR clocks and REQ-78 define per-point lateness against destination C. The sole predicate is coordinate≤as_of plus stamp≤as_of only when late; pending facts are invisible. Delete blanket stamp-gating of gaps. C passes a point only via a published normal gap point, visible at every enabled cutoff. CLI/policy/CI mirror the rule. | Parent C<point≤new C<stamp must expose the normal boundary in the new-C query: previous interval ends before it, selected interval starts after it. A deadline rerun’s accounted event_at≤parent C leaves pre-stamp dossiers identical and activates its late point once C reaches its stamp. |
| ADR012-R9-02 | ADR history/persistence, REQ-48/78 and CLI/policy define lineage (provider repository ID, state ref, genesis publication ID) on all published facts/rows with authenticated ancestry. Strict fast-forward compares every published frontier/fact/row, including post-C history; non-empty/no-C never qualifies vacuously. All other adopted rows require original/current-tip authorization and fresh destination stamps. Reopening is per event and resolution follows its last new-row import publication. | Diverged-prefix destination post-C correction is cross-lineage/Undecided; fork-restored chain needs destination authorization; E′-only import leaves resolved E resolved; destination verdict between two import publications cannot resolve E after the second. Genuine fast-forward still equals fresh restore. |
| ADR012-R9-03 | Persist immutable action-ID-idempotent import-refusal records with action/event IDs, reason, lineage and CAS stamp; export/import them under the same rules. Their visible effect is Undecided until a later authorized destination verdict. Current refusal now covers post-C triage-state-changing records, whether authorized verdicts or import-refusal records. ADR interface/history, REQ-48/78 and CLI/policy mirror this. | Revoked-author import leaves C dossier identical, refuses current certification until its refusal stamp≤C, then computes exit 0 not-ready/Undecided. New authorized destination verdict restores ready; replay changes nothing. |
| ADR012-R9-04 | ADR population, REQ-27/78 and CI condition superseding-run-before-cancellation inference on the cancelled run’s pinned cancel-in-progress for that group; otherwise assess/gap. CLI references the same condition. | Pinned cancel-in-progress=false, H2 before H1 manual cancel, no recorded canceller: H1 still assessed/gapped. |
| ADR012-R9-05 | Replace the combined changed-pin/inspection fixture with independent same-pin inspection and different-pin default-version assertions. Require post-import destination verdict to restore ready. Advance C past started_at before terminal publication when asserting late/stamp-gated terminal facts. Add pending-local, pending-import and unrelated-check-late-record refusal fixtures. ADR acceptance, B4-S and CI carry the strengthened assertions. | Same-pin late rerun cannot change historical representative; different-pin late rerun cannot change historical default version. A permanently Undecided implementation fails resolution. Terminal facts published after C passed started_at must be invisible before stamp. Each pending-input class refuses then clears after publication/C advance; unrelated late evidence refuses then clears, unlike unrelated triage. |

## Mirrors and handoff

- REQ text/card, CLI/policy/CI, STATUS and plan use the corrected invariants. R-077 links round 9. R-070 remains unchanged: no new external source was used. No new ledger ID.
- README and historical ADR status amendments remain accurate without changes; accepted historical bodies and all historical review/response records are untouched.
- Implementation remains pending. Requested signed-file/response SHA-256 values and staged diff statistics are supplied in the handoff; round 10 must review the new bytes before adoption.
