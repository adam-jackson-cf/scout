# ADR-015 council review — round 3 (no sign-off)

## Scope, sources and limits
- **Read only.** No edits, builds, tests, formatters, live provider calls or hashing. This is a design review. P23/S6, P24, P25 code, provider capability receipts and all-writer cutover were not performed and are not claimed.
- **Worktree:** `/Users/adamjackson/Projects/work/scout-adr015`.
- **Reviewed:**
  - The full `ADR.md` and `wiring-contract.json`.
  - `round-1.md`, `round-1-response.md`, `round-2.md`, `round-2-integration.md` and `round-2-response.md`.
  - The actual `scout-nightly.yml` and `scout-guardian-trusted.yml`.
  - The changed companion paragraphs: `STATUS.md:13,15`, `adrs/README.md:18,37`, `plans/2026-11-06-remaining-work.md:94,96`, `research-ledger.md:96-104` (R-100–R-108), `scout-design.html:615` (REQ-48), `:696` (REQ-25), `:706` (REQ-160), `:707` (REQ-161) and the ADR-015 card at `:1050-1061`, `ci-integration.md:79,87,99-110`, `cli-commands.md:136-142` and `policy-file.md:86`.
  - Targeted code facts listed under Evidence, and ADR-012 `ADR.md:17-44`.
- **Byte freeze:** Main reports that all signed files stayed unchanged during this round and that the formatter ran before hashing. The council saw no drift between its reads. It computed **no** digests: the values below are Main's hashlib values from `local://adr015-round3-digests.json` (`verdict: pending`, `open_issues: null`), listed only to identify the bytes reviewed.

| Path | SHA-256 (Main, hashlib; not recomputed by the council) |
|---|---|
| docs/STATUS.md | eb21df80512e7a51c1958209dab139321a55296f7ddbac4ae433a2ae952c3a61 |
| docs/adrs/2026-11-10-shadow-readiness-wiring/ADR.md | daa0b2c52b2906e7f2fd00b334195e4a720f84df2ee02cae95f9320f09582b49 |
| docs/adrs/2026-11-10-shadow-readiness-wiring/wiring-contract.json | c1b9a27a1d9c3d50523bc5a6801d361e884b205037c608fd2bc03db9c9873fd0 |
| docs/adrs/README.md | 3ad288147935642b378af923f7edb99b79365cca470f671445d726d7b79f8e76 |
| docs/plans/2026-11-06-remaining-work.md | 864766198829ba60b8a433bf97f519e9a02effd3cbdc18d340807c85b577f48a |
| docs/references/ci-integration.md | ca8d51af807ac6b719fde14c7e03b7ff18c01019f55b20e77020d8847cdbb893 |
| docs/references/cli-commands.md | dfa5c26c5e42c4475d9cb9a23fb1c2051fa9d6671713b6164cbcbabc5304911c |
| docs/references/policy-file.md | 65533f92f41d1819b7d442493a1348f6b03cf1af353b8ed5b1b193c36e1d845d |
| docs/research-ledger.md | cb50082b56fb5dbcb6c9f7fc22479753786423c8e68182346bd4a2f231de7c0b |
| docs/scout-design.html | 442abbc0a21d38540e37e72867e77b31bb5d42b5d63ff90d53132e73be4336c4 |

- **Peers (real and independently spawned; both finished before this record):**
  - `innovation-challenger`, `ADR015Wiring.ADR015R3Adversarial`: **zero remaining design issues.** It confirmed R2-01 to R2-04 resolved, and the credential, P25 and write-boundary revisions coherent. Its alternative is triage/start only, without inventory. It verified no digests and claimed no live proofs.
  - `experiment-peer`, `ADR015Wiring.ADR015R3Integration`: **5 open issues (2 Medium, 3 Low), N1–N5.**
    - The no-reassessment boundary removes the identity, B, clock and backlog defects without pre-authorizing them.
    - F1, F3–F7, F9 and F10 are resolved. F2 is resolved except the timeout (N2), F8 except helper integrity (N3), and F11 except remaining drift (N5).
    - R2-01 to R2-08 are resolved in the contract.
    - It verified no digests.
- **Lead (independent, before the integration result):** found N1 and N2 independently. Lead evidence for them is merged below.

## Verdict
**5 open issues (Medium 2, Low 3). Not signed off. No `scout-signoff/v1` block is issued.**

The peers disagree (0 issues versus 5). Repository evidence settles it. The adversarial pass did not test the storage or decoding code, or the runner's step-timeout semantics. N1 and N2 are grounded in code and in the runner's source, and the lead reproduced both independently. N3–N5 are concrete and checkable.

**Confirmed resolved, no remaining issue:**
- **Round 2:** R2-01 to R2-08, and round-2 integration F1, F3–F7, F9 and F10.
- **Guard chain:** the literal guards match the actual workflow. If cache restore or provisioning fails, `baseline` is skipped, so the action and `shadow` steps skip too.
- **Event ID:** `^wb-[0-9a-f]{16}$` matches `engine/gate.go:383-384`.
- **Askpass with the existing binary:** git calls inherit `os.Environ()` (`baseline/baseline.go:329`), so the token-free askpass helper works with the existing binary's restore and publish.
- **Final-round selection:** the folder has no `scout-signoff/v1` fenced block. `round-2-integration.md` is not an exact `round-N.md`.
- **Precedent:** the ADR-011/ADR-013 "highest N" final-round wording matches ADR-015 W7.
- **Ledger:** rows R-100–R-108 exist with links. R-108's 1,000 requests/hour figure matches GitHub's rate-limit page (line 58).

## Issues

| ID | Sev | Source | Finding | Evidence | Required fix |
|---|---|---|---|---|---|
| ADR015-R3-01 (N1) | Medium | integration N1; lead independently | **Several v1 durable writes have no storage, decoding or history semantics, and the obvious reuse is unsafe.** The contract allows `start_marker`, `immutable_action_refusal` and `action_idempotence`. Only triage rows have storage. None of the other three exists in schema 14, in `ShadowPublication` or in the manifest-v2 record kinds. Unknown kinds and fields fail decoding. The only existing refusal type is ADR-012's import refusal. It requires `SourceLineage` and `Original.EventID`, so a `start` refusal cannot use it. It also carries per-event Undecided reopening and a current-refusal trigger, so reusing it for direct triage would let a non-maintainer dispatch reopen a resolved veto. Also unspecified: export/import and fast-forward-prefix treatment, cross-lineage handling of a marker, the exit for triage on an unknown, legacy or non-CI event, and the exit for a repeated `start`. | `shadow_migration.go:14-26`; `domain/shadow.go:120-132`; `state/shadow_export.go:17-21,111-131,144,189-194`; `domain/shadow_inputs.go:28-30` and `state/shadow_inputs.go:64-66` (import-refusal shape); ADR-012 `ADR.md:28,44`; REQ-48 (`scout-design.html:615`, an exhaustive file list); contract `settings.durable_write_types`; `ADR.md:50,62`. Schema 14 and manifest v2 are **unreleased** (`CHANGELOG.md:3,6-7`). | **Recommended (route B, see memo):** remove `start` and `triage` from v1, and with them their durable records, so v1 writes nothing. **If start/triage are retained (route A):** specify storage exactly as in "Route A exact storage" below. Do not create schema 15, a manifest v3 or a new export file. Supplement REQ-48 with one sentence. |
| ADR015-R3-02 (N2) | Medium | integration N2; lead independently | **The 2-minute step timeout contradicts the reconcile exit map.** The runner marks a timed-out step `Failed` and merges that into the job result. The contract sets no `continue-on-error`. A slow inventory therefore fails ordinary push and nightly jobs even when `fail-on-incomplete` is false. That contradicts "3: warning; fail iff fail-on-incomplete" and W4's "Cancellation of this last step means unavailable inventory". GitHub says to wait at least one minute after a secondary rate limit, or to honour `retry-after`. A compliant client can therefore exceed 2 minutes on its own, and nothing tells the binary to stop instead of waiting. Inventory can also push a job past its 60-minute timeout; a cancelled job conclusion then masks a passing enforcement step. | `ADR.md:44`; contract `ordinary_reconcile`, `operation_exits.reconcile`; actions/runner `src/Runner.Worker/StepsRunner.cs:327-329` ("has timed out", `TaskResult.Failed`) and `:274-277` (job result merge); GitHub workflow-syntax `steps[*].timeout-minutes` ("killing the process"); REST rate-limit docs line 113 (secondary limits); trusted job `timeout-minutes: 60`. | Contract `settings`: `"inventory_internal_deadline_seconds": 90` covering REST and any git. On a secondary limit (403/429), a `retry-after`, `x-ratelimit-remaining` 0 or primary remaining below 200, stop **without waiting**, and on the deadline exit 3 `inventory_partial`. `ordinary_reconcile` gains `"continue-on-error": "${{ !inputs.fail-on-incomplete }}"`, keeping `timeout-minutes: 2` as a backstop. `operation_exits.reconcile` gains `"step_timeout": "backstop kill = step failure; ignored unless fail-on-incomplete"` and `"job_timeout": "job cancelled; disclosed, never collection"`. W4 replaces "Cancellation of this last step" with "A deadline or timeout kill". Add ledger row **R-109**: runner `StepsRunner.cs` timeout result and the GitHub secondary-rate-limit guidance, cited for this decision. |
| ADR015-R3-03 (N3) | Low | integration N3; lead extension | **Trusted bytes and guard inputs used after scanners have run are not integrity-bound.** `prepare-auth` writes `$RUNNER_TEMP/scout-auth/askpass` once, before restore. The token-bearing publish and `shadow` steps run it after the assessment step. Runner command files (`GITHUB_ENV`, `GITHUB_PATH`, `BASH_ENV`) written during assessment shape later steps. The lead adds that, in route A, enforcement, upload, publish and `shadow` are newly guarded by `env.SHADOW_OPERATION`, a `GITHUB_ENV` value that any unsandboxed process in an earlier step can rewrite. That would suppress enforcement, a dependency the baseline enforcement step does not have. [INFERENCE] Exploitation needs an unsandboxed writer: the bwrap profile binds `/` read-only and allows writes only under `state`, and macOS denies writes outside `state`. Whether model-CLI or Clawpatch processes are sandboxed in that step is not shown. | `ADR.md:68`; contract `runtime_contract.git_credential_helper` (`prepare_step`, `git_steps`), `existing_step_if`, `assessment_enforcement_if`; `sandbox/sandbox.go:53-57` and darwin profile. | (a) Create the helper inside each git step with `mktemp -d "$RUNNER_TEMP/scout-auth.XXXXXX"` and remove it with `trap`. Remove the `prepare-auth` step and its order constraints, or else have every token step check the helper's fixed SHA-256 (taken from `script_lines`) before setting `GIT_ASKPASS`. (b) Route A only: replace every `env.SHADOW_OPERATION` guard with `steps.shadow-inputs.outputs.operation`; an earlier step's output cannot be rewritten later. (c) List these explicitly as S6/P23 write-integrity targets: `GITHUB_ENV`, `GITHUB_PATH`, `GITHUB_OUTPUT`, `BASH_ENV`, the helper, the installed Scout/Clawpatch binaries and the toolchain directory. |
| ADR015-R3-04 (N4) | Low | integration N4 | **The serialization of the logical action ID is not fixed.** Readiness returns unmeasured unless `ShadowTriage.ActionID` is decimal. The contract gives `logical_action: [repository_id, run_id]` with no encoding, so a composite string would silently make readiness unmeasured. | `shadow/query.go:85-88`; contract `auth.logical_action`; ADR-012 `ADR.md:26,28`. | Route A only: `"action_id_encoding": "ActionID = canonical base-10 run_id (no sign or leading zeros); RepositoryID = canonical base-10 numeric repository ID; applies to triage rows, start marker and refusals"`. Route B removes the field. |
| ADR015-R3-05 (N5) | Low | integration N5; lead addition | **Wording still says v1 collects, and REQ-160 omits a reserved command.** W2 and W3 say "frontier-only reconciliation", but frontier writes are forbidden. W3 says "run bounded collection". The title, card and README row say "collection". REQ-160 reserves `reconcile\|triage\|start` but omits `validate`, a governed command (W2 ¶3, contract `commands.validate`). Lead addition: "What changes and review" (`ADR.md:80`) gives roles for R-100–R-103 only, while the header cites R-100–R-108. | `ADR.md:1,24,34,76,80`; `adrs/README.md:37`; `scout-design.html:706,1051`; contract `commands`. | Use "read-only readiness status; C unchanged" and "diagnostic inventory" throughout (title, W2, W3, W7 ¶3, card, README row). Route A: add `validate` to REQ-160 and update `approved_requirement_sha256.REQ-160`. Route B: REQ-160 needs no ADR-015 change, so revert its ADR-015 sentence and move REQ-160 to `unchanged.requirement_sha256`. Add one sentence giving the roles of R-104–R-109. |

### Route A exact storage (only if start/triage stay in v1)
- **No DDL, no new file, no new record kind, no version change.** Schema 14 and manifest v2 are unreleased.
- **Envelope fields.** Add two optional fields to the existing `ShadowPublication` envelope. It is already exported as `publication` records in `pr-runs.jsonl` (counted and checksummed), stored in `shadow_publications.payload`, CAS-stamped (`RecordedAt`), lineage-chained (parent/genesis) and part of the whole-history prefix:
  - `collection_start`: action ID, numeric actor ID and login, policy revision, receipt reference and `sha256`, and lineage genesis. At most one per lineage. On empty history, the start publication **is** the genesis.
  - `action_refusals[]`: action ID, operation, optional event ID, refusal code, numeric actor ID and login, policy revision and `run_attempt`. Idempotent by (repository ID, action ID).
- **Accepted triage** stays the existing `triage` row with the same action ID. Idempotence is the existence of a triage row, a marker or a refusal with that action ID. No separate idempotence record is needed.
- **Semantics:** action refusals and markers are **not** triage verdicts or import refusals. They never change event triage state, Undecided status or ADR-012 current-refusal triggers.
- **Imports:**
  - A strict same-lineage fast-forward preserves them byte-identically.
  - Cross-lineage and pending imports keep them only as source provenance inside pending material. They never become destination authority, so the destination needs its own `start`.
  - Old manifest-v2 readers fail on the new fields. This is accepted because v2 is unreleased and W6 requires every writer to upgrade.
- **Exits:**
  - Triage on an unknown, legacy, local or non-CI event: exit 1, refusal recorded (`event_not_ci`).
  - Repeated `start` with the same receipt digest: exit 0 `already_published`, no write.
  - Repeated `start` with a different digest: exit 1, refusal recorded (`marker_exists`). The marker is never mutated.
- **REQ-48 supplement** (one sentence; record its approved hash in the contract): "ADR-015 adds optional collection-start and direct-action-refusal fields to the v2 publication envelope; they are lineage-bound control records that never alter triage, Undecided state or refusal triggers, and cross-lineage imports never adopt them as destination authority."
- **Why not Main's `metadata` cell `shadow_control_v1` plus `state/v1/shadow-control.jsonl`:**
  - It holds an append-only log in a mutable single-row `TEXT` cell, so immutability rests on code alone.
  - It duplicates the stamp and lineage that the publication chain already provides.
  - It adds a file, which changes REQ-48's file list.
  - It needs new manifest count/checksum, merge and prefix logic.
  - It is acceptable as a fallback, but it is larger. The council did not verify the plan-S2 "versioned additional files" allowance Main cites, because the plan lives in a session-local file.

## Decision memo

1. **Recommendation: route B, reduce v1 to read-only diagnostic inventory plus credential hardening.** Remove `start`, `triage`, the four dispatch inputs and their normalization step, the rollout receipt, the authorization machinery and every durable write. Defer them to the later source/reassessment ADR, which must create the first authenticated CI events anyway.
   - **Why:** v1 creates no events (W5). Schema 14 and manifest v2 are unreleased (`CHANGELOG.md:6-7`). Only trusted CI writers may append events (ADR-012 `ADR.md:26`). No certifying source exists in any lineage. So production can contain no authenticated CI event to triage. [INFERENCE from these facts.] In v1, `start` gates only that empty set and a diagnostic horizon. Both are dead weight in the sense of this repository's engineering rules, and they carry the whole N1/N4 surface and most of N3.
   - **What remains:**
     - The `shadow` step goes last. It runs `scout shadow reconcile` with the N2 deadline/`continue-on-error` contract, takes `GITHUB_TOKEN` only, and needs no git token.
     - Existing step guards stay **unchanged**. Shadow runs after enforcement, so it cannot suppress it, and no `existing_step_if` deltas remain.
     - `actions:read` remains, as a breaking grant, disclosed and gated on S6. Assessment's `GITHUB_TOKEN` widens, so S6 must pass before merge.
     - Also kept: `persist-credentials: false`, per-step askpass (N3a/c), and restore that fails closed on any `ls-remote` exit other than 0 or 2.
     - An old binary now produces reconcile exit 3, a warning unless fail-on-incomplete. It no longer hard-fails before assessment.
     - Inventory order becomes newest first by original `created_at`, then numeric run/attempt/job IDs, within the caps. There is no marker horizon, and the result is never coverage.
   - **Not settled by route B:** existing/imported vetoes and same-lineage obligations still pass to the later ADR, together with the W1 authorization text from rounds 1–3, which that ADR can reuse.

2. **Alternatives.**
   - **(A) Keep start/triage and apply R3-01 to R3-05 using the route A storage above.** This is viable and moderate in size, and the authority path ships early. But triage cannot be exercised live until a source exists, and the signed surface stays large (receipt, inputs, exit maps).
   - **(A′) Route A with Main's metadata-cell plus control-file storage.** Viable but larger, for the reasons above.
   - **(C) Adversarial peer's triage/start only, without inventory.** Rejected on the evidence: triage acts on an empty set, and the start marker would then gate nothing. It still needs `actions:read` for W1's REST corroboration.
   - **(D) Withdraw shadow wiring entirely and ship only the credential hardening.** Smallest of all. It also drops the diagnostic inventory that measures real population limits (R-101: fork bindings, retention, saturation) as input to the source ADR. It is acceptable if `actions:read` breakage is judged not worth diagnostics, but it is not recommended, because the source ADR needs that data.
   - **Schema 15 or manifest v3:** rejected. Nothing is released, so no cutover is needed.

3. **Evidence.**
   - **Storage and decoding:** `internal/state/shadow_migration.go:14-26`; `internal/domain/shadow.go:120-132`; `internal/state/shadow_export.go:17-21,111-131,144,189-194`; `internal/domain/shadow_inputs.go:28-30`; `internal/state/shadow_inputs.go:64-66`.
   - **Readiness and identity code:** `internal/shadow/query.go:85-88`; `internal/engine/gate.go:383-384`.
   - **Git and sandbox:** `internal/baseline/baseline.go:142,201,227,329`; `internal/sandbox/sandbox.go:53-57`.
   - **Releases:** `CHANGELOG.md:3,6-7`.
   - **Requirements and ADR-012:** REQ-48 `scout-design.html:615`; REQ-160 `:706`; ADR-012 `ADR.md:26,28,44`.
   - **Workflow facts:** the trusted job has `timeout-minutes: 60`; the enforce step has no `if:`; checkout does not set `persist-credentials`; restore treats any non-zero `ls-remote` as "branch absent".
   - **External:** actions/runner `StepsRunner.cs:274-277,327-329`; GitHub workflow-syntax `steps[*].timeout-minutes`; REST rate limits, lines 58 and 113.
   - **Peer records:** `agent://ADR015Wiring.ADR015R3Adversarial` and `agent://ADR015Wiring.ADR015R3Integration`.

4. **Risks.**
   - **Route B:** a later breaking change re-adds dispatch inputs. That is additive with defaults and non-breaking for callers. [INFERENCE] The source ADR also carries more scope.
   - **Route A:** a larger review surface, plus untestable live triage.
   - **Both routes:**
     - `actions:read` breaks external callers that do not grant it. The grant widens the token held in the assessment step until S6 passes.
     - Inventory consumes the shared 1,000 requests/hour budget, bounded at 100 per run with a floor of 200.
     - [INFERENCE] Whether model-CLI or Clawpatch processes run sandboxed in the assessment step is unproven, so N3(c) is an S6 gate, not an assumption.
   - **Reversibility:** everything is design-only and reversible by a superseding ADR. Signed history is immutable.

5. **Execution outline.**
   1. Main picks route B (recommended) or A, and edits `ADR.md` and `wiring-contract.json`:
      - Both routes: R3-02, R3-03(a)(c) and R3-05.
      - Route A also: R3-01 (route A storage), R3-03(b) and R3-04.
      - Route B also: remove W1 inputs and authorization, W2 actions and normalization, W6 start/receipt, and the related contract blocks (`operation_grammar`, `caller_dispatch_inputs`, `trusted_workflow_call_inputs`, `forwarding`, `action_ids`, `action_if`, `existing_step_if`, the `assessment_enforcement_if` delta, `runtime_contract.normalization`, `auth`, `start`, `commands.validate/triage/start`), and set `durable_write_types: []`.
   2. Update the companions to match:
      - **Design:** REQ-25/161, REQ-160 (revert or add `validate` per route), REQ-48 (route A only), and the card title.
      - **Trackers and references:** README row, STATUS, plan, `ci-integration`, `cli-commands`, `policy-file`.
      - **Ledger:** add R-109.
      - **Hashes:** recompute the contract's approved and unchanged REQ hashes.
   3. Write `round-3-response.md` mapping R3-01 to R3-05 to fixes. It must contain no signoff block.
   4. Freeze the bytes, then recompute the full signed-set manifest with both `shasum -a 256` and `hashlib`.
   5. Hold round 4 with both peers on exactly those bytes. Check the R3 fixes, contract literalness and companion parity. Issue a `scout-signoff/v1` block only at zero issues.
   6. Implementation proofs P23–P25 and the S6 and provider receipts remain separate gates.

## Peer influence on this recommendation
- **Integration peer:** added N3–N5, which the lead had not raised. The lead had dismissed "frontier-only" as harmless wording; the peer's REQ-160 `validate` drift is concrete and was accepted. It confirmed N1 and N2, including the import-refusal reuse hazard the lead had also found.
- **Adversarial peer:** its zero-issue confirmation of R2-01 to R2-04 is accepted. Its minimal alternative (C) was rejected on the evidence. It prompted the lead to test the inverse minimal route, which became the recommendation (B).
- **Main's metadata-cell proposal:** evaluated as A′. The envelope storage is preferred if route A is chosen.
