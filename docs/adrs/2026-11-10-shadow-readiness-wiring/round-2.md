# ADR-015 council review — round 2 (no sign-off)

## Scope, sources and limits
- **Read only.** No edits, builds, tests, formatters or live provider calls. This is a design review. Implementation and live proofs (P23/S6, P24, P25 code, provider capability receipts, all-writer cutover) were not performed and are not claimed.
- **Reviewed (worktree `scout-adr015`):**
  - `docs/adrs/2026-11-10-shadow-readiness-wiring/{ADR.md, wiring-contract.json, round-1.md, round-1-response.md}`
  - `.github/workflows/scout-nightly.yml` and `.github/workflows/scout-guardian-trusted.yml` (all steps)
  - `internal/sandbox/sandbox.go:41-65`
  - ADR-012 `ADR.md` items 1–7 and its R5–R8 fixture lines
  - Companion statements: `STATUS.md:13`, `adrs/README.md:18,37`, `plans/2026-11-06-remaining-work.md:94`, `research-ledger.md:96-99`, `scout-design.html:696,706,707,1050-1061`, `ci-integration.md:79,87,101-108`, `cli-commands.md:138-140`, `policy-file.md:86`
  - Manifest `local://adr015-round2-digests.json`: 10 files, `verdict: pending`, `open_issues: null`
- **Byte drift during the round (blocks sign-off on its own).**
  - The first read of `ADR.md:18` said "GitHub's documented rerun reuse of original inputs (R-102)".
  - A later read in this round says "integrity rests on trusted first-attempt context and ordering. R-102's input-reuse claim is corrected by R-104".
  - `ADR.md:26` likewise gained an enforcement condition mid-round.
  - The council therefore cannot attest that the manifest digests match the bytes it reviewed. The council computed no digests.
  - The current `ADR.md` cites R-104, but `docs/research-ledger.md` has no R-104 row (grep: no match), so the citation is dangling.
- **Peers.**
  - `innovation-challenger` (ADR015PeerAdversarial) finished with 4 issues. All are incorporated below as R2-01 to R2-04.
  - `experiment-peer` (ADR015PeerIntegration) was **still running** when this record was written. Its result is not incorporated and nothing is attributed to it. Round 3 must include it.

## Verdict
**8 open issues (High 1, Medium 4, Low 3). Not signed off. No `scout-signoff/v1` block is issued.**

Round-1 fixes confirmed sound (no remaining issue):
- **R1-01, inputs not verifiable by REST:** W1 now says this correctly; see R2-04 for the rerun premise.
- **R1-02, none sentinel and string forwarding:** the contract `forwarding` matches the actual `inputs.* || default` caller pattern.
- **R1-03, breaking `actions: read` upgrade:** disclosed in W3 and `ci-integration.md:103`.
- **R1-04, `persist-credentials: false` and the S6 proof:** see R2-07 for the remaining enumeration gap.
- **R1-07, start marker:** see R2-03 for the receipt schema.
- **R1-08/09/13, no event population, `Complete=false`, no lag or event deadline:**
  - "Run-only facts without event population" is consistent with ADR-012.
  - ADR-012 item 5 vetoes count would_block events linked to *attempts* (item 7). Those are distinct from provider PR-event accounting (item 2).
  - So "no attempt accounts an event" and "linked would_block events stay cohort-wide Undecided vetoes" do not contradict.
- **R1-11/12, concurrency and version-skew disclosure:** confirmed, but R2-04 changes the rerun remedy.
- **R1-14/15, signed set and ledger rows:** confirmed. The non-narrowing headline and the "certifying source open, no date" text are consistent across STATUS, README, the plan, design, ci-integration and policy-file.

Trigger, cron, concurrency and permission baselines in `wiring-contract.json` match the actual workflows:
- Trusted concurrency group and `cancel-in-progress: false`.
- Caller job grants `contents: write` / `pull-requests: read`.
- Checkout without `persist-credentials` (the baseline `true`).
- `scout-version: v0.1.6`.

## Issues

| ID | Sev | Source | Finding | Evidence | Required fix |
|---|---|---|---|---|---|
| ADR015-R2-01 | High | challenger #1 + lead | **The contract cannot support exact JSON surface equality.** W7 requires that every unknown change refuses, and calls the contract the normative machine copy. Yet several approved surfaces are prose or placeholders, and surfaces that W1–W3 need are missing. A correct implementation would therefore fail P25, or the implementer would have to interpret prose, which the ADR forbids. | Prose/placeholder fields: `action_if` uses the literal placeholder `'<triage\|start>'` (contract `:162`). `existing_steps_guard`, `action_placement` and `minimum_release` are prose. `release_artifacts` uses brace and `<version>` globs with no defined match grammar. REQ-25/160/161 have no approved bytes or extraction rule; only REQ-162/163/168 hashes exist (`:69-73`). **Absent:** the W2 input-normalization step (id, position, `if`, how it sets `SHADOW_OPERATION`); the step `env:` bindings for `GITHUB_TOKEN` (restore, publish, assessment, shadow); the per-operation exit mapping and summary; handling of invalid input combinations; the unchanged trusted-job `if:` and `timeout-minutes: 60` (the W4 2-hour rationale depends on 60); the caller `scout-version` delta rule; the reassessment clock source. | Make every approved surface an exact, parseable value. **(a)** Use two literal `if` strings, one each for `shadow-triage` and `shadow-start`. **(b)** Give a literal `if` per existing step: map, assessment, publish (`(steps.assessment.outputs.exit_code == '0' \|\| … == '1') && env.SHADOW_OPERATION == 'reconcile'`), upload and enforce (see R2-02). **(c)** Add a normalization step object: id, insertion index (after `baseline`, before `map`), `if`, the env it writes, and invalid handling. Recommend: invalid combination ⇒ `SHADOW_OPERATION=invalid`, all governed steps skipped, explicit summary, exit 3, no publication. **(d)** Add a step→env/secret map pinning the `GITHUB_TOKEN` and `OPENROUTER_API_KEY` scope per step. **(e)** Give the exit mapping per operation as data. **(f)** Give the W1 grammar as data: the event-ID regex, reason byte bounds, the forbidden control-character class and the operation × field matrix. **(g)** Add a `scout_version` rule: a release-tag regex; P25 classifies tag bumps as approved, while command presence is checked by the implementation's release test, not by the contract. **(h)** Define glob semantics for artifacts, or list them literally. **(i)** For REQ-25/160/161, define a canonical extraction (the `<tr>` row text with whitespace normalised) and record SHA-256 at adoption. **(j)** Pin `timeout-minutes` and the trusted-job `if` under `unchanged`. **(k)** Negative fixtures mutate each field independently. |
| ADR015-R2-02 | Medium | challenger #2 | **Assessment enforcement is not independent of shadow failure.** The enforce step has no status function, so GitHub inserts an implicit `success()`. A failing `shadow` step (for example exit 3 with fail-on-incomplete) then suppresses enforcement, which hides a failing assessment. W2 promises independence. | `scout-guardian-trusted.yml`, step "Enforce assessment result" (no `if:`); contract `existing_steps_guard`; GitHub status-check functions documentation. Main's mid-round `ADR.md:26` edit is not yet reviewed as final bytes. | Select exactly: enforce `if: ${{ !cancelled() && env.SHADOW_OPERATION == 'reconcile' && steps.assessment.outcome != 'skipped' }}`. Encode it in the contract per R2-01(b). Positive fixture: shadow fails and the assessment exit is 1, so the job reports "Scout gate failed". |
| ADR015-R2-03 | Medium | challenger #3 + lead | **The start receipt is unspecified.** W6 requires the binary to validate an "immutable digest/reference against the deployed writer inventory and no-old-writer evidence". The only interface is a free-text `--reason`, and the contract `start` (`:225-229`) has only booleans. Reference syntax, storage and trust root, lineage binding, and which facts are API-checkable versus attested are all undefined. REST cannot enumerate local or external publishers. | `ADR.md` W6 ¶2–3; contract `start` | **Select a receipt.** <ul><li>Schema `scout-rollout-receipt/v1`, stored as a file committed on the protected default branch.</li><li>Reference grammar: `<40-hex commit>:<repo-relative path>#sha256=<64-hex>`. The binary checks: the commit is an ancestor of the current protected tip; the blob hash matches; the path is under a fixed directory (for example `docs/rollout/`).</li><li>Fields: `repository_id`; `state_lineage`; `last_intact_v2_commit` (verified as a `scout/state` ancestor); `migration_result`; `writer_inventory[]` (workflow path, ref and release pin); `cancelled_runs[]` (verified completed/cancelled through REST with `actions: read`); `s6_evidence` (reference and digest); and `attested[]` for unobservable claims (local/external publishers, operational rerun prohibition, ruleset). Attested claims are authorised only by the start actor's current-tip maintainer authorization and are recorded as such.</li><li>Refuse on unknown schema, a binding mismatch, a non-terminal listed run, a missing S6 reference or a non-ancestor commit.</li><li>Put the schema and the grammar in the contract.</li></ul> This is an implementable contract only; no deployed receipt is required for docs adoption. |
| ADR015-R2-04 | Medium | challenger #4 | **The rerun-input-preservation premise is unsupported.** The cited rerun documentation promises actor privileges, SHA and ref, not inputs, and REST cannot compare inputs. W1 ¶3 nevertheless lets the same actor rerun an action that published nothing. W2's "same-actor rerun" remedy, W3's concurrency remedy and the ci-integration text depend on that premise. Main's mid-round edit cites R-104, which does not exist in the ledger. | `ADR.md:18,20,26,34`; `research-ledger.md:98` (R-102); `ci-integration.md:105`; contract `auth` | **Choose fail-closed.** <ul><li>Triage/start with `run_attempt > 1` and no published record for the logical action ⇒ `not_published`, exit 2, remedy "fresh dispatch".</li><li>Keep deduplication of published actions.</li><li>Rerun authorization may be enabled only by a later adopted ADR, after a capability receipt covers full and single-job reruns.</li><li>Update W1 ¶3, the W2 exits, the W3 concurrency remedy, `ci-integration.md:105` and the contract `auth` (add `unpublished_rerun: refuse`).</li><li>Add the R-104 ledger row correcting R-102. Do not edit R-102's historical row beyond ledger rules.</li></ul> |
| ADR015-R2-05 | Medium | lead | **`fail-on-incomplete` semantics conflict, and the pre-start state breaks existing callers.** The input is documented as "Fail the job on exit 2 or 3". W2 now makes shadow reconcile exit 2 neutral regardless, an undeclared change of meaning. Before a start marker exists, every reconcile exits 3 (W2). That state is the default for every caller that upgrades and never runs `start`. So every caller with `fail-on-incomplete: true` fails on every push or night until start, a second, undisclosed breaking upgrade. | `ci-integration.md:79,87`; trusted workflow input `fail-on-incomplete` description; `ADR.md` W2 ¶2 "Before a start marker, reconcile/triage exit 3" | **Select one.** Recommended: with no start marker, reconcile returns exit 2 with status `not_started` (unmeasured, neutral, no write). Exit 3 is reserved for real unavailability (old binary, missing state, S6/isolation refusal). Shadow reconcile exit 2 never consults `fail-on-incomplete`; exit 3 fails only with it. Update the input description, `ci-integration.md:79,87` (scoped "assessment exit 2/3; shadow reconcile: 2 always neutral"), W2 and the contract exit map. Triage before a start marker stays exit 3 (job failure). |
| ADR015-R2-06 | Low | lead; challenger capability note | **The reassessment clock is undefined, so v1 collection may be empty without saying so.** W5 requires "independently authenticated provider elapsed time", else no reassessment. The job API exposes start and completion timestamps, not an advancing clock (challenger). The contract fixes `reassessment_start_budget_ms` but names no clock and lists no proof. Implementers may reach for HTTP Date or wall time, or ship a collector that can never reassess, without disclosure. | `ADR.md` W5 ¶2; contract `settings`, `implementation_proofs`; GitHub job API | Keep the refusal (the challenger agrees; do not substitute HTTP Date or writer wall time). Add `"reassessment_clock": "unselected; requires capability receipt"` and the proof item "authenticated elapsed-time source" to the contract. State that until the clock is proven, reconcile performs zero reassessments and reports `reassessment_unavailable` (never zero findings). STATUS and the plan should say that v1 may collect no would_block events until then. |
| ADR015-R2-07 | Low | lead | **The S6 credential-confidentiality proof list omits process-argument and sibling-process channels.** "Per-command Git credentials" invites `git -c http.extraheader=…`, which puts the token in argv. The list "process environment" is ambiguous between a scanner's own environment and other processes' `/proc/<pid>/{environ,cmdline}` (or `ps` on macOS). Linux has `--unshare-pid`; the macOS profile restricts writes only. | `sandbox.go:54,62`; round-1 R1-04 (macOS write-only profile); `ADR.md` W3 ¶3 | Forbid credentials in argv: use a credential helper, `GIT_ASKPASS` or stdin. Extend S6/P23 to prove that scanners cannot read other processes' argv or environment, on Linux and macOS. Put both in the contract `implementation_proofs`. |
| ADR015-R2-08 | Low | lead | **The final round is not uniquely identifiable for P25.** W7 says "the explicit final round contains one block" but does not say how the gate locates it, or what happens if an unapproved round, a response file or a draft contains a block-like JSON. | `ADR.md` W7 ¶2 | Rule: exactly one fenced `scout-signoff/v1` block across the folder's `round-*.md` files. Its `round` must equal the file name's N, and N must be the highest round number. Any other count, or a block in a non-round file, fails closed. Add negative fixtures. |

**Procedural blocker (not a design issue):** the signed bytes changed during the round, and R-104 is dangling. Round 3 must freeze the bytes, recompute the manifest with both `shasum` and `hashlib`, and the council must review exactly those bytes.

## Decision memo

1. **Recommendation.** Keep ADR-015 as a non-certifying collection/triage wiring ADR; the design direction stays sound and ADR-012 invariants and REQ-162/163/168 are preserved. Apply R2-01 to R2-08, freeze the bytes, then hold round 3 with both peers, including the outstanding integration-peer result. Sign only at zero issues.

2. **Alternatives.**
   - **(A) Recommended:** fail-closed reruns plus a fully literal contract.
   - **(B) Allow unpublished reruns now, on the documentation premise.** Rejected: the premise is unsupported (R2-04).
   - **(C) Keep contract prose and let P25 match it heuristically.** Rejected: it gives the implementer the discretion the ADR forbids.
   - **(D) Drop collection and ship triage/start only until a reassessment clock is proven.** Smaller, but triage would have nothing to act on. (A) plus R2-06 disclosure achieves the same safety.
   - **(E) Pre-start exit 3 plus a documented breaking change for fail-on-incomplete callers.** Viable but noisier for adopters. R2-05 prefers `not_started`/2.

3. **Evidence.** The citations in the issues table. Actual workflow facts:
   - Enforce step has no `if:`.
   - Publish guard is `exit_code == '0' || '1'`.
   - Upload uses `always()`.
   - Trusted job runs only on the default ref, `timeout-minutes: 60`.
   - Checkout has no `persist-credentials` set.
   - Restore step outputs `available=false` but succeeds when `scout/state` is absent.

4. **Risks.**
   - The rerun capability remains unproven, so maintainers must re-dispatch.
   - The reassessment clock may never be provable, so v1 collection may stay empty (disclosed).
   - External callers face a breaking `actions: read` grant plus a version pairing.
   - An unobservable local publisher depends on the attested receipt fields.
   - Everything is reversible through a superseding ADR. The signed history is immutable.
   - [INFERENCE] GitHub UI choice preselection and fork `pull_requests` emptiness remain untested.

5. **Execution outline.**
   1. Main edits ADR W1–W7 and `wiring-contract.json` per R2-01 to R2-08. Update `ci-integration.md:79,87,103,105`, STATUS, the plan and the design rows to match. Add ledger row R-104.
   2. Write `round-2-response.md` mapping each issue to its fix.
   3. Freeze the bytes, then recompute the 10-file manifest.
   4. Round 3: the integration peer's result, a fresh challenger pass, and lead verification of contract literalness. Check each issue's fix, and confirm that companion statements match.
   5. At zero issues only, issue a `scout-signoff/v1` block with path-sorted files and SHA-256s. Implementation proofs P23–P25, the capability receipts and the rollout receipt remain separate gates.
