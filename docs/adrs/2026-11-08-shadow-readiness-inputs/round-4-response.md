# Round 4 response: ADR-012

Date: 2026-11-08. Writer response to ADR012-R4-01–16. No sign-off or council closure is claimed. No commit, product code, tests, builds, formatters, rendering checks or experiment runs.

## Delegated scope cut

Selected under R-064: adopt only the evidence, counting and authorization contract needed by REQ-22/27/78, not a CI implementation design. The repeated review failures came largely from specifying a workflow before implementing it. The current ADR and references therefore remove the selected pull_request_target/dispatch topology, named scout-shadow.yml, caller/job parsing, artifact format/name/caps, retention/deadlines, fixed registration/finalizer commands, policy workflow key, sandbox recipe and installation changes. They do not replace those with a second collector design.

The implementation PR chooses that wiring and must demonstrate the retained invariants: authentic immutable CI identity, complete accounting including zero events/failed publication, guardian-equivalent assessment/inputs, stable observations/events, conservative missingness, current maintainer-only original-author triage, replay idempotence, publication-safe cutoffs, durable history and trusted-writer integrity. Schema/export compatibility and the read-only readiness query remain because persisted evidence and historical reproduction require them. REQ-162/163/168 are restored to adopted main rather than freezing release/install mechanics for a workflow no longer selected here. Historical reviews/responses are unchanged records of superseded proposals.

## Issue mapping

| Issue | Fix / narrowed disposition |
|---|---|
| ADR012-R4-01 (major) | Reconstructed REQ-22 from adopted main, retaining its complete auto-demotion sentence verbatim, including “auto-demote that repo's check to report for that repo and record it [ASSUMPTION]”. Only observation counting/triage additions change; ADR-011's version/reset and per-repository wording remains. |
| ADR012-R4-02 (major) | Removed the selected dispatch interface, but retained its necessary authorization guarantees for any implementation: protected default-branch execution/reachable source; policy/CODEOWNERS from current protected tip at publication; authenticated original author equals executing author and is still authorized; reject/log branch/tag, revoked or cross-actor execution; one stable repository-qualified action ID is replay-idempotent. Conditional GitHub mapping is repository ID plus dispatch run ID. Existing ruleset remains the trust boundary against write-capable collaborators. ADR, REQ-22/26/27/51/161 and references carry the invariant, not new workflow wiring. |
| ADR012-R4-03 | Use publication recorded_at, assigned on publishing CAS as max(writer UTC now, parent collectedThrough+1ms), reassigned on retry; collectedThrough never decreases. Historical triage uses recorded_at≤as_of, provenance dispatch time does not count, and current_window uses cutoff=collectedThrough. ADR/REQ-78/CLI state the rule and require late-publication historical-reproduction proof. |
| ADR012-R4-04 | Every query, including default cutoff, restricts representatives, unioned links and Days origin to interval.start≤started_at≤as_of. An observation with only earlier attempts is absent. ADR/REQ-78/policy/CLI now say this without a historical-only qualifier. |
| ADR012-R4-05 | Shadow event identity uses base=merge_base(B,H), head=H tree OID, then check/issue/rule_version. B and assessed merge result are attempt provenance. ADR/REQ-27 require stable linkage when B advances; event union no longer multiplies equivalent findings. |
| ADR012-R4-06 | Require exactly the pinned guardian invocation's override inputs: currently none. Remove inferred provider/body/label overrides and the extra scan command. Guardian parity proof explicitly includes override-marker PR text; wiring overrides into both lanes remains a separate design change. |
| ADR012-R4-07 | Separate remedies in ADR/REQ-48/CI: downgraded hybrid v1 refuses exit 3, recover last intact v2 and reapply later state; genuinely newer schema refuses exit 3, upgrade writer and never overwrite it with an older snapshot. Recheck every CAS retry. Removed obsolete workflow-specific copies from policy/CLI references. |
| ADR012-R4-08 | Removed the selected sandbox/post-upload sequence and REQ-163 installation amendment. Retained the integrity invariant: untrusted scans cannot alter trusted executable/configuration files, writer authorization/recipe or other attempts. The implementation must prove protection of every file later executed by trusted steps; it chooses read-only/absent mounts and writable output/scratch layout. Accepted own-attempt measurement falsification is distinguished from escalation. ADR/REQ-51/CI align. |
| ADR012-R4-09 | Readiness discloses represented selected-check inspected/not_inspected/not_applicable counts. Provisional reason includes inspected count, with no eligibility/unknown/provisional rule change beyond the declared observation contract. ADR/REQ-78/CLI align. |
| ADR012-R4-10 | REQ-10 and ADR explicitly add append-only would_block_triage and copy each v13 triage value once with legacy provenance/original triaged_at. Derived authorized-history reads replace the old column; legacy does not become new authorized CI evidence. |
| ADR012-R4-11 | Restore and clarify the base restriction: PR/fork jobs have read-only scout/state access and never publish to it; temporary local stores remain permitted. No baseline-only loophole or special PR-triggered writer exception remains. |
| ADR012-R4-12 | Historical status lines, README and card relations now name readiness observations/eligibility, maintainer triage history, trusted readiness publication, targeted migration and export manifest v2. Installation is no longer amended after the scope cut. README ADR-000 retains ADR-001/002/003 as amenders. ADR-008 relation is unchanged. Only historical status lines were edited. |
| ADR012-R4-13 | Activation overview now says five comparable-base PR observations (distinct PR/head tree/merge-base), 10% of observations, and an observation counts once. No other part of that callout changes. |
| ADR012-R4-14 | Superseded by the authorized scope cut: ADR-012 no longer selects scout-shadow.yml, so it cannot require that named workflow in the release list. Restore all of REQ-168 to adopted main. B4-S must choose and deliver whatever wiring its implementation needs before B5; no unshipped workflow is promised in this ADR. |
| ADR012-R4-15 | Add B4S→B5 to the plan graph, include B4-S in its summary and narrow its implementation row. STATUS names ADR-012 as latest amendment effective on its own explicit final zero-issue round, with adopted ADR-011 (PR #17) prior. Gap summaries no longer promise a selected workflow. |
| ADR012-R4-16 | Expand only unadopted R-070/R-077: dispatch/context/merge-tree documentation links; guardian merge/base/no-override facts, trusted default guard/restore order, exact-name authorization, event-ID inputs and reviewed lack of CLI event triage; round-3/4 links and REQ-26 plus historical REQ-162/163/168 proposal attribution. Cite these sources in ADR/card. Accepted R-064/R-069 remain untouched. |

## Exact signed set (13 files)

1. `docs/adrs/2026-11-08-shadow-readiness-inputs/ADR.md`
2. `docs/scout-design.html`
3. `docs/adrs/README.md`
4. `docs/STATUS.md`
5. `docs/plans/2026-11-06-remaining-work.md`
6. `docs/references/policy-file.md`
7. `docs/references/ci-integration.md`
8. `docs/references/cli-commands.md`
9. `docs/research-ledger.md`
10. `docs/adrs/2026-10-19-v1-decisions/ADR.md`
11. `docs/adrs/2026-10-23-decisions/ADR.md`
12. `docs/adrs/2026-10-24-cutover-cleanup/ADR.md`
13. `docs/adrs/2026-10-31-jev-merge-association/ADR.md`

Council rounds/responses are reviewed records, not signed canonical files. Prior records and ADR-011's accepted summary/card/README row remain unchanged. Main receives SHA-256 values, this response's separate digest, staged diff stat against origin/main and unstaged-change stat. These are staged uncommitted drafts for renewed council review, not adoption. Adoption still requires explicit zero-issue review of these actual integrated bytes, recorded outside the signed set.
