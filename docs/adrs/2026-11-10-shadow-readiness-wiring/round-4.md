# ADR-015 council review — round 4 (no sign-off)

## Scope, provenance and limits
- **Read only.** No edits, builds, tests, formatters, live provider calls or hashing. This is a design review. No P23/S6, P24 or P25 implementation, provider receipt or cutover proof was performed, and none is claimed.
- **Worktree:** `/Users/adamjackson/Projects/work/scout-adr015`. Folder: `docs/adrs/2026-11-10-shadow-readiness-wiring`.
- **Reviewed:**
  - The full `ADR.md` and `wiring-contract.json`.
  - `round-3.md` and `round-3-response.md`. Earlier rounds are historical only.
  - The named companion paragraphs: `STATUS.md:13,15`; `adrs/README.md:18,37`; plan `:94,96`; `research-ledger.md:96-105` (R-100 to R-109); `scout-design.html:696` (REQ-25), `:707` (REQ-161) and the ADR-015 card `:1049-1061`; `ci-integration.md:79,87,99-110`; `cli-commands.md:136-142`; `policy-file.md:86`.
  - Targeted facts: the actual `scout-nightly.yml`, `scout-guardian-trusted.yml:39-67,74-245` and the `scout-guardian.yml` concurrency/permissions, plus `internal/baseline/baseline.go:201,329`.
  - A sweep of the signed set for left-over route-A text (`SHADOW_OPERATION`, `prepare-auth`, `shadow-inputs`, `triage/start`, `frontier-only`, `bounded collection`, `action_id_encoding`).
- **Digests:** the council computed no digests. The values below are Main's, from `local://adr015-round4-digests.json` (`verdict: pending`, `open_issues: null`). Main computed them with both hashlib and `shasum -a 256`; the two agreed on all 10 files. They are listed only to identify the bytes reviewed. Main reports that the pre-commit hook passed at commit `9799b1c5` without changing these bytes, and that no frozen file changed during round 4. The council saw no drift between its reads.

| Path | SHA-256 (Main; not recomputed by the council) |
|---|---|
| docs/STATUS.md | 21a06fa4265a106d85b645a22efebce2155feac3610ef4da69594cff5a69ba65 |
| docs/adrs/2026-11-10-shadow-readiness-wiring/ADR.md | 833a249df011aa492d58f2825bec312bae668aa39f287a78f19e19bffc6f654d |
| docs/adrs/2026-11-10-shadow-readiness-wiring/wiring-contract.json | 6a75a5eaaa720c3c1220a948c2dedea60b1b6a1381e50c5ccbe51c7c368119e0 |
| docs/adrs/README.md | 4a7affe5f8a711dca78cef75e90b384f9cd25adb758c55cd43683f45c6e172ff |
| docs/plans/2026-11-06-remaining-work.md | 8281afc3b908b1a9c4c16aca84a5ad5cb7bafad845e16d5bfdce4253a6ffe69d |
| docs/references/ci-integration.md | e636e4da88a845478c7ffe7eb7feefd5c4ef6db3ecd6105b7eb4cf4ab319bfc4 |
| docs/references/cli-commands.md | 27e6f033508be7eb377893bf063b5edcd53d61fc275f1fe8fb2ec2483dde79f5 |
| docs/references/policy-file.md | 964c813eaca641b46c69b3a54049240046274e6427395dbb2545a8598bea2222 |
| docs/research-ledger.md | a6fe13d920e2589ced6e32834f86cd4142af1c852caf635b8ea08f528aee2904 |
| docs/scout-design.html | 9a877aaa37b8083892195af8d217743340e79aba6de4c210f75015069e0eb8ce |

- **Peers.** Both were real, independently spawned and finished before this verdict:
  - `innovation-challenger`, `ADR015Wiring.ADR015R4Adversarial`: **1 Low**, the stale triage/start clauses at `ci-integration.md:79,87`. It found all other route-B fixes for R3-01 to R3-05 coherent. It recomputed no digests and claimed no live proofs.
  - `experiment-peer`, `ADR015Wiring.ADR015R4Integration`: **2 Low**. One is the same `ci-integration.md:79,87` finding. The other is the W6 ordering, which lets the permission grants merge before the S6 proof. It found everything else in parity. It recomputed no digests and could not hash requirement rows against `baseline_commit`.
- **Lead (independent):** found the `ci-integration.md:79,87` clauses during its own residue sweep, before reading either peer's output.

## Verdict
**2 open issues (Low 2). Not signed off. No `scout-signoff/v1` block is issued.**

## Round-3 issues
| ID | Status | Evidence |
|---|---|---|
| ADR015-R3-01 (N1) | **Resolved** | `ADR.md` W4: "Durable writes = none". The contract has `durable_write_types: []`, and its 12 `forbidden_writes` match W4's list exactly. No start, triage or validate command, no dispatch input and no marker. Reassessment is false, and all deadlines and lag are `null`. W5 states what is open, why, what unblocks it, and that ADR-012 readiness stays incomplete. Route A is "known viable option, not adopted" in W5, the card, STATUS, the plan, `ci:101`, `cli:142` and R-109. REQ-48's hash is under `unchanged`. |
| ADR015-R3-02 (N2) | **Resolved** | W3 and contract `settings`/`ordinary_reconcile`/`operation_exits` agree on: 90-second deadline, `timeout-minutes: 2`, 100 requests including retries, floor 200, and stopping without waiting or retrying on 403/429/`retry-after`/missing or malformed headers. `continue-on-error: ${{ !inputs.fail-on-incomplete }}`. `fail-on-incomplete` is an actual boolean `workflow_call` input (`scout-guardian-trusted.yml:39-42`). The `step_timeout`, `job_timeout` and `old_binary` exits are mapped. The R-109 row exists. |
| ADR015-R3-03 (N3) | **Resolved in W2 and the contract; one ordering defect in W6 (ADR015-R4-02)** | The askpass helper uses fresh `mktemp -d "$RUNNER_TEMP/scout-auth.XXXXXX"` per Git step with an EXIT trap and holds no token. The Git steps are exactly `baseline` and `Publish the state baseline`. No shared prepare-auth step. `shadow` gets `GITHUB_TOKEN` only. `ls_remote_exit` 0/2/other fails closed. The S6 targets are identical in W2 and `s6_integrity_targets`. `existing_step_guards` is unchanged. Git inherits `os.Environ()` (`baseline.go:329`), so `GIT_ASKPASS` reaches restore and publish. |
| ADR015-R3-04 (N4) | **Resolved** | No `logical_action`/`action_id_encoding` remains. Route B removes it. |
| ADR015-R3-05 (N5) | **Resolved in ADR, contract, design, index and ledger; one companion left over (ADR015-R4-01)** | The title, card `:1051`, README `:37` and STATUS use "diagnostic inventory". REQ-160 has no ADR-015 sentence and its hash is under `unchanged`. Only REQ-25 and REQ-161 are under `approved`. "What changes and review" gives roles for R-100 to R-109. The remaining `frontier-only` hits (plan `:61`, `ci:114`) are ADR-012 fixture text. |

**Checked and in parity with the actual workflows:**
- Triggers and cron `23 3 * * *`.
- The three dispatch inputs.
- Caller and reusable permission ceilings (`contents: write`, `pull-requests: read`).
- `trusted_job_if` and `timeout-minutes: 60`.
- Trusted concurrency `cancel-in-progress: false`; the PR workflow uses `true` and `contents: read`.
- Step names used in `step_order_constraints`, and restore id `baseline`.
- The enforce step has no `if:`, and the upload step uses `always()`. A last-step `!cancelled()` shadow guard therefore cannot suppress either one.

## Issues
| ID | Sev | Source | Finding | Evidence | Minimal fix |
|---|---|---|---|---|---|
| ADR015-R4-01 | Low | lead, adversarial and integration, each independently | **The signed CI reference still states triage/start exit semantics as "reserved ADR-015 wiring".** This contradicts W1 ("No validate/triage/start command … is adopted here"), contract `commands` and `operation_exits` (reconcile only), and the same file's `:101` deferral. It is the R3-05 synchronization left incomplete. | `ci-integration.md:79` "…; triage/start 2/3 always fail."; `:87` "…; triage/start require publication/idempotent success." | `:79`: end the cell at "…shadow 3 consults this flag." `:87`: end at "…3 warns/fails with that flag." Delete both triage/start clauses and keep the table alignment. |
| ADR015-R4-02 | Low | integration; lead confirmed | **W6 orders the permission grants before the S6 proof.** The `actions:read` grants are edits to `scout-nightly.yml` and `scout-guardian-trusted.yml`. "First land … grants" followed by "prove … S6 … before workflow merge" lets the token-widening workflow change merge before S6. That contradicts W2 ("S6 must establish scanner/process isolation **before implementation merges**"), REQ-25 `:696`, contract `implementation_proofs[0]` and the decision contract. | `ADR.md` W6 ¶1, first sentence | Replace the first sentence of W6 ¶1 with: "For diagnostic implementation: first ship a compatible released binary, retaining existing release assets/install scripts; prove restore, per-step credential behavior and Linux/macOS S6 process/filesystem integrity; only then merge the caller/reusable `actions:read` grants and the workflow change together in one paired workflow merge." No contract change is needed: its "before workflow merge" already covers the grant merge. |

## Decision memo
1. **Recommendation:** keep route B unchanged. Apply only the two text fixes above, to `ci-integration.md:79,87` and the first sentence of W6 ¶1. Do not change `wiring-contract.json` or any other signed file. Then hold round 5 with both peers on a fresh 10-file manifest. Sign off only at zero issues.
2. **Alternatives:**
   - **(a) Accept R4-02 as harmless because W2 governs.** Rejected. The ADR's own prose would contradict itself on the one ordering the decision contract makes non-negotiable.
   - **(b) Also reword plan `:94` ("compatible release/caller grants and S6 isolation before workflow merge").** Optional. Read as written, it already puts S6 before any workflow merge. Leaving it avoids extra signed-byte churn.
   - **(c) Route A publication-envelope controls.** Still a known viable future option. Not adopted, and not reopened without new evidence.
3. **Evidence:**
   - **Signed set:** `ci-integration.md:79,87,101`; `ADR.md` W1, W2 ¶2–4 and W6 ¶1.
   - **Contract:** `commands`, `operation_exits`, `implementation_proofs`, `s6_integrity_targets` and `ordinary_reconcile`.
   - **Design:** REQ-25 `:696`, REQ-161 `:707`, card `:1049-1061`.
   - **Workflows and code:** `scout-guardian-trusted.yml:39-67,111-245`; `scout-nightly.yml:6-55`; `scout-guardian.yml:38-43`; `baseline.go:329`.
   - **Peer records:** `agent://ADR015Wiring.ADR015R4Adversarial` and `agent://ADR015Wiring.ADR015R4Integration`.
4. **Risks:**
   - Both fixes change signed bytes (`ADR.md`, `ci-integration.md`), so the manifest must be recomputed.
   - Requirement-row hashes have not been re-verified against `baseline_commit` by the council or by either peer. They rest on Main's computation and P25 implementation checks.
   - Everything is design-only and reversible by a superseding ADR.
5. **Execution outline:**
   1. Edit `ci-integration.md:79,87` and W6 ¶1 exactly as above.
   2. Write `round-4-response.md` mapping R4-01 and R4-02 to the fixes, with no sign-off block.
   3. Freeze the bytes and recompute all 10 digests with hashlib and `shasum -a 256`.
   4. Hold round 5 with both peers on those bytes. Confirm the two fixes and that nothing else changed. Issue a `scout-signoff/v1` block only at zero issues.

## Peer influence
- **Adversarial peer:** confirmed R4-01 independently, matching the lead's own finding. Its zero findings elsewhere support closing R3-01 to R3-04.
- **Integration peer:** added R4-02, which the lead had missed. The lead read W2 and the contract and treated S6-before-merge as settled, without testing W6's sequencing words. After checking W6 against W2, REQ-25 and the contract, the lead accepted R4-02 as a grounded Low issue. This is why the verdict is 2 issues rather than 1.
