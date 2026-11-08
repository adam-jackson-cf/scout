# Round 3 response: ADR-012

Date: 2026-11-08. Writer response to ADR012-R3-01–13. Every recommended option is selected by the agent under R-064; owner review remains pending in editable trackers. This is a fix record, not council closure or sign-off. No code, commit, tests, builds, experiments or rendering checks.

## Issue mapping

| Issue | Delegated choice and change |
|---|---|
| ADR012-R3-01 (major) | Choose the simple shared release floor: N is the first schema-14 release; all scout/state-sharing workflow pins must be N+, no pre-N run in flight/rerun. Trusted guardian performs the first v2 publish; shadow begin/finish on v1 publish nothing and report the named unmeasured gap. Only genuine v1 without v2-only files normalizes; downgrade/newer-schema publication refuses exit 3, rechecked every CAS retry, with last-v2-commit remedy. ADR, REQ-48/161/168, policy/CI/CLI references and delivery proofs agree. No independent pin-distribution mechanism. |
| ADR012-R3-02 (major) | Choose protected workflow_dispatch triage through the existing writer, not a new broad issue-triage system. Event/verdict/reason enter via env. Provider triggering_actor must match trusted base-policy maintainers/CODEOWNERS; reject/log non-maintainers. Schema-14 append-only would_block_triage records actor/reason/time/authorization/run provenance; current verdict derives from latest authorized row. Export history deterministically, union/conflict-refuse on merge, CAS-publish. Threaded through REQ-22/27/48/161, ADR and CLI/CI/policy references; authorization/history proofs required. |
| ADR012-R3-03 (major) | Choose sandbox-built merge-tree of registered B/H, built-in git merge-tree --write-tree only, assessed against B; retain original tree/merge-base dedup identity. Record merge-tree OID. Conflicts are not_assessed (merge_conflict). Non-conflicting equivalence to GitHub merge is [ASSUMPTION]; same-head event/comparison parity with guardian is a delivery proof. REQ-27 scopes this to shadow and leaves guardian unchanged. |
| ADR012-R3-04 (major) | Restore a narrowly scoped protected caller gate: exactly one pinned reusable job with literal release inputs and no secrets, otherwise caller_not_dedicated/not_assessed/configuration unmeasured. Finish/reconcile reject foreign attempt jobs. Trusted scan preamble checks current run/attempt intent before PR fetch; missing intent is unregistered_attempt/not_assessed and disclosed. Document rerun-all; finalize-only reruns finish the original attempt. Caller/partial-rerun proofs required. |
| ADR012-R3-05 (minor) | Representatives require started_at≤as_of; triage uses append-only authorized rows at/before cutoff. ADR and REQ-78 specify both. CLI labels explicit-as-of historical_reproduction versus latest-cutoff current_window and keeps any descriptive current-triage re-evaluation separate. |
| ADR012-R3-06 (minor) | Event evidence unions eligible links across every assessed attempt of a represented observation at the cutoff; comparability/unknown still use its latest assessed representative. ADR, REQ-22/78 and references explicitly prevent an empty later rerun hiding earlier not-warranted evidence. |
| ADR012-R3-07 (minor) | Intent/provider, not payload, controls PR/head/tree/base identity; independently re-derive merge_base through compare API. Mismatch yields comparison-none/identity_mismatch. Artifact remains bounded schema-validated untrusted data. Accept the explicit residual that a compromised sandboxed scanner can falsify its one attempt's events/statuses, not intent/other attempts/writer state. |
| ADR012-R3-08 (minor) | Add all eight rollout/hardening conditions: public event-policy allow, SHA data-only fetch (explicit unsafe-checkout opt-in only if used), env-only validated numeric/SHA PR identities/no PR-field run interpolation, no caller secrets, hosted or isolated ephemeral runners, no cache restore/save/write-capable cache-mode, no sandbox command files or output interpreted as workflow commands, delivered-event-only [ASSUMPTION] with blocked trigger unmeasured. ADR, REQ-161/168, policy and CI checklists plus proofs cover them. |
| ADR012-R3-09 (minor) | REQ-26/51 explicitly classify protected no-PR-execution register/finalize/triage as trusted default-branch jobs, writing shadow data only, never baseline issues/dispositions/export-base. REQ-162/163 extend verified tag install to shadow before PR fetch, without caches. ADR revision/card include all added REQs. All four historical status-line notes, README rows and card relations consistently name amended counting/eligibility, workflow/publish/install, migration and export/merge areas; bodies untouched. |
| ADR012-R3-10 (minor) | Declare no-eligible-unit/not_applicable exclusion from the selected check's five-observation minimum as deliberate owner-decision amendment under R-064, in ADR and card. REQ-78 uses comparable-base observations and the exact PR/head-tree/merge-base key. REQ-78 and planned policy section define pr_runs/min_pr_runs as PRRuns observations, not attempts. |
| ADR012-R3-11 (minor) | Name planned scout diff --shadow --base B --head merge-tree-oid --baseline-dir export --out file. Specify B/H objects, sandbox merge tree, intent-recorded state-export commit, protected base policy and guardian-equivalent authorized scout-override/scout:allow metadata as offline inputs. No baseline-remote access inside sandbox. ADR and CLI contract contain the command. |
| ADR012-R3-12 (minor) | Expand unmerged R-070 with secure pull_request_target, reused-job reruns, artifact REST schema/upload v4 and conflict-trigger documentation; retain source facts rather than freezing choices into the ledger. Cite them in ADR/card. Replace unsupported Main attribution with agent selection under R-064. No new research ID needed. |
| ADR012-R3-13 (minor) | Add Track B item B4-S for ADR-012 implementation and all proofs; B5 now depends on B4 and B4-S plus completion of Scout's own shadow window. STATUS decisions-in-force range includes ADR-012 conditionally on final post-rebase zero-issue review. |

## Exact canonical signed set (13 files)

Paths are repository-relative:

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

CI/CLI references are newly included because round 3 requires their explicit rollout/interface contracts. Council rounds and responses, including this response, are separately reviewed records, not signed canonical files. Prior rounds/responses remain byte-identical. No edits were made to scout-adr11b for this revision.

## Submission/adoption conditions

SHA-256 values and git diff --stat are supplied externally to the main agent, without self-hashing this response. Submit these canonical bytes for round 4. R-064 is absent until prerequisite ADR-011 merges; rebase onto it, resolve intervening changes, then renew council review of the actual final signed bytes. A pre-rebase zero-issue round cannot authorize final adoption. Record adoption outside the signed set only after explicit zero-issue sign-off. Implementation remains the later B4-S PR; shipped schema is still 13.
