# ADR-015: Wire read-only diagnostic shadow inventory

- **Status:** accepted-form; effective only on this folder's explicit final zero-issue council round
- **Deciders:** agent-decided under owner delegation (R-064)
- **Decided:** 2026-11-10; **sources:** R-064, R-077, R-100–R-109
- **Delivery:** design only; implementation, S6 isolation, provider capability and operational proofs remain pending

## Decision and limits

Adopt the council's round-3 primary recommendation: v1 is **read-only diagnostic inventory plus credential hardening**, not production event collection or triage/start transport. No certifying event source exists. ADR-012 readiness stays incomplete/unmeasured until the later source/reassessment ADR and its implementation/proofs close the open decisions below. This is an explicit council-directed deferral, not a claim to satisfy ADR-012's implementation acceptance. Missing completeness/time is never ready.

The normative machine copy of W1–W6 is [wiring-contract.json](wiring-contract.json). Both prose and contract are signed; disagreement fails the gate. No Go, workflow, secret, tool pin or release artifact changes ship in this adoption.

### W1 — Read-only command, placement and outcomes

Reserve only `scout shadow reconcile --root <root> --json`, for bounded provider inventory and optional read-only readiness status. It performs no scanning, models, store initialization, migration, import or publication; no Git token or Git network operation is needed. No validate/triage/start command, dispatch input, normalization step, action authorization interface or start marker is adopted here. Existing workflow inputs and every existing map/assessment/publish/artifact/enforcement guard remain unchanged.

Use the existing trusted job, not another job/workflow. Add install step ID `install`; retain restore ID `baseline`. New step ID `shadow` goes **last, after assessment-artifact upload and assessment enforcement**, guarded by `!cancelled() && steps.install.outcome == 'success' && steps.baseline.outcome == 'success'`. An assessment failure therefore does not suppress inventory; unsafe installation/restore does. Shadow cannot consume time reserved for already completed artifacts/enforcement. Existing assessment exit-2/3 handling stays unchanged.

Inventory 0/2 is neutral regardless of fail-on-incomplete; 3 warns and fails only with fail-on-incomplete. Every result labels readiness unmeasured, Complete=false and C unchanged; diagnostic success is not activation evidence. Unknown/missing old-binary command is unavailable/3, without suppressing the preceding assessment. Summaries distinguish inventory_available, inventory_partial and inventory_unavailable; none implies publication. An absent baseline/store is not initialized and cannot prevent provider-only diagnostics.

### W2 — Triggers, grants and credential integrity

Retain pull_request, main push, cron `23 3 * * *` and the existing workflow_dispatch inputs. No workflow_run or pull_request_target. The unchanged trusted-job default-ref guard selects push/schedule/dispatch; the PR lane is unchanged. Keep trusted concurrency cancel=false and PR cancel=true; pending-run replacement or missed/delayed schedules never means coverage.

Add actions:read at every trusted caller/reusable-workflow/job permission ceiling, retaining contents:write and pull-requests:read. **Every external caller upgrading to the reusable revision must grant actions:read or startup fails.** No actions:write, issues:write, PR grant change or new secret. This widens the assessment token too: S6 must establish scanner/process isolation **before implementation merges**, not merely before deployment.

Set trusted checkout persist-credentials:false. Within each existing credentialed Git step (restore and baseline publish), create a fresh token-free askpass helper using `mktemp -d "$RUNNER_TEMP/scout-auth.XXXXXX"`; install the exact contract script, set step-local GIT_ASKPASS and GIT_TERMINAL_PROMPT=0, and delete the directory on exit with a trap. It returns fixed username x-access-token or reads step-local SCOUT_GIT_TOKEN for a password prompt. Never put credentials in argv/config/files. There is no shared prepare-auth step or cross-step helper. Inventory receives GITHUB_TOKEN only; it receives no Git token/helper. Existing assessment secret bindings are unchanged, but no scanner may inherit credentials or read credential-bearing sibling processes.

Restore treats `git ls-remote --exit-code` 0 as present/restore, 2 as absent, and every other status as authentication/network failure: fail restore, never rebaseline. S6/P23 protects own/sibling-process argv/environment and these explicit integrity targets on Linux/macOS: runner GITHUB_ENV/GITHUB_PATH/GITHUB_OUTPUT/GITHUB_STATE/GITHUB_STEP_SUMMARY/BASH_ENV, askpass, installed Scout/Clawpatch binaries and the toolchain directory. Environment scrubbing or read-only mounts alone are not proof. Existing enforcement guards are not newly made dependent on mutable environment values.

### W3 — Bounded inventory, no completeness attestation

Inventory is diagnostic GitHub REST run/attempt/job enumeration, newest first by original run created_at, then descending numeric run/attempt/job IDs. There is no marker horizon. Authenticate repository identity and record endpoint/query/pagination/retention/saturation/permission limitations; do not turn a successful or exhausted listing into a PR-event ledger. GitHub's filtered run listing can cap at 1,000 and supplies neither complete PR-event populations nor authenticated assessment B/recipe/inspection facts. API timestamps alone do not authenticate a complete-through boundary. Retention, deleted/skipped/cancelled runs, fork bindings, ambiguous ordering, saturation, missing pages/time or permission failures all remain explicitly unavailable/partial and readiness unmeasured.

Constants are versioned in this signed contract and the implementing released binary: 90-second internal deadline across all diagnostic work; two-minute workflow timeout backstop; at most 100 HTTP requests including retries; stop below 200 authenticated rate-remaining. On 403/429, retry-after, remaining=0, or remaining below 200, stop **without waiting or retrying** and report inventory_partial/unavailable/3. Missing/malformed rate headers also stop; request/deadline budgets include every call. No population exclusion or silently truncated success is allowed.

Set shadow continue-on-error to `${{ !inputs.fail-on-incomplete }}`. The binary deadline returns 3/inventory_partial; a timeout backstop kill is a step failure ignored only when fail-on-incomplete=false. A whole-job timeout/user cancellation may still cancel the 60-minute trusted job after assessment reporting; disclose unavailable inventory rather than promise assessment-equivalent job conclusion. Limits are operational diagnostic caps, not evidence deadlines, lag guarantees, cross-run quota reservations or completeness attestations. Changing them requires new matching adopted wiring authority.

### W4 — Exact write boundary

**Durable writes = none.** No start marker, action refusal/idempotence record, triage, PRRunFact, provider event, event link/obligation, gap, interval, frontier or S3 publication is authorized by reconcile. No reassessment. Provider attempt deadline, PR-event deadline and finite collection lag remain unset; a clock alone cannot enable certifying collection. Complete stays false and C does not advance. ADR-012 library APIs remain implementation primitives, not proof of a selected source.

Existing baseline publication remains the original assessment path; this ADR neither removes nor enlarges its authority. Diagnostic output is not persisted into readiness evidence. Existing/imported vetoes, triage and lineage history are preserved unchanged.

### W5 — Explicit deferred decisions and known viable option

**Open:** complete event/started-attempt source, authenticated observation/base H/B/merge-result and recipe/inspection provenance, per-observation provider identity, trustworthy complete-through time, evidence deadlines/lag/horizon/capacity; and triage source/transport with authenticated original/executing actor, stable logical action and current-tip authorization, rollout receipt/start and durable control encoding.

**Why open:** v1 GitHub run listing cannot prove those populations or observation facts. V1 creates no authenticated CI events; schema14/manifest-v2 are unreleased. The council therefore found no production CI event set on which an early triage/start authority could act, and chose not to adopt unused mutable control machinery.

**Unblock:** a later adopted source/reassessment ADR must select and prove the complete source, identity/B/parity/time/capacity contracts; select triage/start transport and durable semantics; pass provider refusal/capability, sandbox P23/P24, gate P25 and operational S6 proofs. It must preserve existing/imported vetoes and triage **in the same lineage**; a newly complete source establishes a new interval origin/first frontier, never replaces genesis or discards history. ADR-012 readiness and its S6 rollout remain incomplete until then. This ADR does not substitute refusal-only tests for those proofs.

**Known viable option, not adopted:** round-3 route A adds optional collection-start/direct-action-refusal fields to the existing ShadowPublication envelope, stored in shadow_publications.payload and exported as publication records in existing pr-runs.jsonl. Accepted triage remains existing triage rows; marker/triage/refusal existence supplies logical-action idempotence. Markers/refusals never alter event triage, Undecided state or current-refusal triggers. Strict same-lineage prefix preserves them; cross-lineage/pending material is provenance only and never destination authority. This would need the council's one-sentence REQ-48 supplement, exact fields/IDs/exits, compatible writers and fresh review in the later ADR. No schema15, manifest-v3, new export file, optional field, key, command or control-write authority is adopted here. See [round 3](round-3.md) for the full unadopted option.

### W6 — Operational rollout

For diagnostic implementation: pair a compatible released binary with the caller/reusable wiring revision, retaining existing release assets/install scripts. Merge caller/reusable permission grants and diagnostic workflow wiring together **only after** restore, per-step credential and Linux/macOS S6 process/filesystem proofs pass; no permission-widening workflow merge may precede them. Upgrade every trusted caller in that gated merge sequence. An old binary reports unavailable rather than manufacturing coverage. Enable only last-step diagnostics, inspect summaries/caps and retain ordinary assessment. No state-format migration, all-writer start receipt or marker is required for a read-only command; existing state-format safeguards remain unchanged.

Certifying rollout/cutover is explicitly deferred with ADR-012 plan §3 S6: migrate compatible state; upgrade and quiesce every shared-state reader/writer and old in-flight/rerun publisher; validate current protected-default operational evidence; establish destination start/interval rules under the later ADR. Do not treat diagnostic enablement as that cutover. No old in-flight/rerun writer may downgrade state. Compatibility, all-writer inventory and local/external operational attestation must not be represented as REST completeness proofs.

### W7 — P25 adoption gate

Gate the actual repository workflow/command/permission/policy/release and REQ-160–163/168 surfaces against baseline plus the exact approved deltas in wiring-contract.json. All unknown additions, changes **and deletions** fail; existing-requirement wiring needs no redundant ADR. Baseline includes literal guards, dispatch input types/defaults, permission ceilings, trigger/concurrency, release asset regexes and requirement-row hashes. This decision adds only read-only reconcile, trusted actions:read, install ID, last-step bounds and credential/restore hardening; it does not authorize the deferred control surfaces.

The explicit final round contains one fenced scout-signoff/v1 JSON block: adr, folder, round, open_issues=0, verdict=approved, and path-sorted files with SHA-256. Select highest N among exactly named round-N.md files; require exactly one block there matching N/folder/ADR. A higher pending/unsigned round fails. Lower signatures remain immutable history; reject signoff blocks in response/non-round files. Locate the unique HEAD-ancestry commit adding that final round (all parents); hash signed blobs at that immutable adoption commit, not current mutable trackers. Current final-round/contract bytes must equal adoption blobs. Missing/shallow/ambiguous history fails closed with fetch remedy. Status-line amendments follow existing ADR rules; later amendments use a new folder/contract and verified supersession links. Real signed authority, never an ADR number/mock, gates additive/destructive surface equality.

## What changes and review

REQ-25/161, the ADR index/card, trackers and references record this reserved design, not shipped behavior. REQ-48/160/162/163/168 and ADR-012's signed bytes remain unchanged. R-100 supplies actual workflows/policy/releases and the implementation plan; R-101 provider limits; R-102 initial dispatch/context analysis; R-103 concurrency; R-104 rerun correction; R-105 status guards; R-106 future event grammar; R-107 integration identity/parity findings; R-108 primary quota; R-109 runner timeout/secondary limits. Historical source/response choices are not current authority where this decision defers them.

Adopted-form route (a): signed set = this ADR; wiring-contract.json; docs/scout-design.html; docs/adrs/README.md; docs/STATUS.md; docs/plans/2026-11-06-remaining-work.md; docs/research-ledger.md; docs/references/{ci-integration,policy-file,cli-commands}.md. This folder's explicit final zero-issue round records outcome/digests; adoption is recorded outside signed files. Owner review remains pending; confirmation is a new user ledger row, rejection a superseding ADR. No implementation, experiment, deployment or readiness proof is claimed.
