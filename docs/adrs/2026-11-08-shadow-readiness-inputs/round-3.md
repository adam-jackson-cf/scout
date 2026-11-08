# Round 3: ADR-012 (shadow-window readiness inputs)

Date: 2026-11-08. This round follows design rules 1–3 and the ledger rules in `CLAUDE.md`, and the ADR rules in `docs/adrs/README.md`. ADR-012 is an agent decision made under the owner's standing delegation (R-064). After round 2, ADR-011 and ADR-012 were split into separate PRs. This round reviews ADR-012 alone. Where a fix involves a choice, the council names the option it recommends under that delegation.

The council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review of `pull_request_target` safety, trust between jobs, the claim that nothing from the PR is ingested, fidelity to the requirements, the migration, and the status lines.
- **`experiment-peer` (Opus):** closure of every earlier issue, fidelity to REQ-22, REQ-27 and REQ-78, competing interpretations and confounds, ledgers, trackers and status lines.

The review was read-only. Nobody edited files, ran model calls outside the council, executed product code, or touched `.codegraph`.

**Main computed and verified the digests below with `shasum`. The council did not, because the lead and both peers had read-only tools and no shell.** This round is not a sign-off, so the digests only identify the bytes that were reviewed. All paths are under `docs/`.

| File | SHA-256 (Main) |
|---|---|
| `scout-design.html` | `d2af86ac21ad420f38c7a7d53d9d65a39882e3d5a3b48d1304d290a8477b7b04` |
| `adrs/README.md` | `592395734223a95a191d8911f9980fab9eab8f0fe88b2ab9c61d3e67811456aa` |
| `STATUS.md` | `7519354604f1c4329acad9cc5817e8978ac52cb6c12a563ee58f8ea4ff186119` |
| `plans/2026-11-06-remaining-work.md` | `3f3df820e7c623d6c2d89d3d2abcafcff10be39023eca97edba00e353c224a12` |
| `references/policy-file.md` | `853e39baa3fb347ab8fe2332ba4799d5d0b09ec5ca54a8ff70c0b0cd7e244094` |
| `research-ledger.md` | `270fa9bef2afe4d765ce8b8d3c566748cee2c1e538edcafbfe9e4d7f5ad7d78d` |
| `adrs/2026-10-19-v1-decisions/ADR.md` (ADR-000) | `41e8a9ecab73500e1c8fb7b9e7ea9c5d88d744258778a1d1724507953e553aa8` |
| `adrs/2026-10-23-decisions/ADR.md` (ADR-003) | `c0744a4cc64ceeadfe894ebf70a22ce8fc1d7fa1448e416e50d9644c354bb48f` |
| `adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) | `ba819ac228dbe38465a313f186baad33b14a63ae177ebbc2c33da8ffcb621e6e` |
| `adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) | `9224f15862f360a88cc9936469b8ad8ac38023c0fd96b343a0311a87e5c50278` |
| `adrs/2026-11-08-shadow-readiness-inputs/ADR.md` | `44ac6887ff5ce02ca34779dc7756407e83ebed91fb1936277436d0763282dabf` |
| `adrs/2026-11-08-shadow-readiness-inputs/round-2-response.md` | `0fb44ac74a471bb88e7ea28ecc4cc8ab3acd8f379b2483faa131fb5d7c9b9d0b` |

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11`, branch `agent/policy-shadow-adrs`, with uncommitted changes on top of main `ef6f50e8`.
- Base bytes from `/tmp/adr11-patchcheck`.
- ADR-012 will rebase onto ADR-011 once ADR-011 merges.

**Design under review:** a protected default-branch workflow registers intent. A credential-free scanner job then runs on the PR code, and a trusted finalization step completes the record. Results from PR-controlled workflows are not ingested.

**Files read:**
- ADR-012 `ADR.md` (all lines), `round-1.md` (issue summary), `round-2.md` (lines 1–61 and 180–520) and `round-2-response.md`.
- `scout-design.html`, worktree and base:
  - REQ-10/22/26/27/48/51/78/160/161/162/168 (`:589,600,610–612,615,618,706–708,714`);
  - the cards for ADR-000 `:736`, ADR-003 `:807`, ADR-007 `:901`, ADR-008 `:925` and ADR-012 `:991–995`.
- Line 3 of ADR-000, ADR-003, ADR-007 and ADR-008, worktree against base. The peers compared the bodies (see below).
- `README.md:12–27`, `STATUS.md:7–9,107–131`, plan `:15,31–36,59–60,84–101`, `policy-file.md:17–24,68–74`, `research-ledger.md:72–73`.
- Code:
  - `.github/workflows/scout-guardian.yml`, `scout-guardian-trusted.yml` and `scout-nightly.yml` (permissions, cache, restore, secrets);
  - `scripts/bootstrap-tools.sh:55–120`;
  - `internal/state/export.go:680–700`, `internal/state/would_block.go:1–103`;
  - `internal/baseline/merge.go:100–140`, `internal/baseline/baseline.go:27–28,84–97`.
- External: GitHub's "[Securely using pull_request_target](https://docs.github.com/en/actions/reference/security/securely-using-pull_request_target)" (read by the lead), plus the re-run and trigger documentation cited by the peers (links under the issues).

## Issue summary

**13 distinct issues are open: 4 major and 9 minor.**

| ID | Severity | Summary |
|---|---|---|
| ADR012-R3-01 | major | The shadow caller has its own release pin, so it can publish schema 14 and manifest v2 to `scout/state` while other workflows on older pins still read and write it: old readers fail, old writers silently downgrade |
| ADR012-R3-02 | major | No path exists for maintainers to triage CI would-block events into `scout/state`, and "authorized triage" is undefined |
| ADR012-R3-03 | major | Shadow assesses the PR head against `merge-base(B,H)`, but the gate it activates assesses the merge commit against B; PRs with merge conflicts are counted; REQ-27's new sentence has no scope |
| ADR012-R3-04 | major | Run integrity is asserted but not enforced: other jobs in the caller run can upload the transport, and a partial re-run can scan with no intent registered for that attempt |
| ADR012-R3-05 | minor | Historical `as_of` queries are not reproducible: triage is the current triage, and there is no `as_of` cutoff on representative attempts |
| ADR012-R3-06 | minor | Event counting after deduplication is ambiguous: a later failed re-run can hide an earlier attempt's not-warranted event |
| ADR012-R3-07 | minor | Finalize does not say that identity comes from the intent and not from the payload; the residual risk of a compromised scanner is not stated |
| ADR012-R3-08 | minor | `pull_request_target` platform prerequisites and hardening are missing from the rollout and delivery contract |
| ADR012-R3-09 | minor | REQ-26/51/162/163 are not reconciled with the new workflow; the status-line notes are incomplete |
| ADR012-R3-10 | minor | One owner-decided change is not listed, and the counting terminology is stale |
| ADR012-R3-11 | minor | The scan command and its data inputs (baseline source, override handling) are not specified |
| ADR012-R3-12 | minor | Ledger and provenance gaps: four GitHub facts have no source, and one decision is credited to "Main" |
| ADR012-R3-13 | minor | Trackers: the plan has no ADR-012 implementation item, B5 does not depend on one, and `STATUS.md:129` gives a stale range |

## Closure of earlier issues

| Issue | Status | Evidence / remainder |
|---|---|---|
| ADR011-R2-01 (shared; patch adoption order) | **Closed** | The patch route was removed. Separate ordinary PRs, ADR-011 first (`ADR.md:75`, README `:12`, STATUS `:9`, plan `:90`). |
| ADR011-R2-02 (shared; frozen ledger rows) | **Closed** | R-064 and R-069 are not on this branch. R-070 and R-077 (`research-ledger.md:72–73`) hold source facts only. R-077's caveat leaves the design choices to the ADR. |
| ADR011-R2-04 (shared; reset-gap list) | **Deferred, not open on these bytes** | ADR-012 makes no content-binding claim (REQ-78 `:600` defers to ADR-011). Re-check in the review after the rebase. |
| ADR011-R1-01/02, R2-03, R2-05 | Not applicable | ADR-011's own issues. They are in its separate PR. |
| ADR012-R1-01 event identity | **Closed** | `ADR.md:44`; REQ-27 `:612`. |
| ADR012-R1-02 forged PR artifacts | **Closed by redesign** | No PR-workflow import (`ADR.md:10,24`; REQ-51 `:618`). The forgery remainder inside the run is ADR012-R3-04. |
| ADR012-R1-03 epochs | **Closed** | `ADR.md:50`. |
| ADR012-R1-04 timestamps and `as_of` | **Closed** | `ADR.md:32,56,65`. |
| ADR012-R1-05 terminal states and gaps | **Closed** | `ADR.md:34–40`. |
| ADR012-R1-06 re-run inflation | **Closed** | `ADR.md:50`. |
| ADR012-R1-07 inspection predicate | **Closed** | `ADR.md:52`; REQ-78. |
| ADR012-R1-08 CI vs local scope | **Closed** | `ADR.md:44`; REQ-48. |
| ADR012-R1-09 older binaries drop facts | **Partly closed** | Manifest v2 makes old *imports* refuse (`export.go:693–694`). Old *publishers* still merge over v2 and rewrite a v1 manifest (`merge.go:128–136`): ADR012-R3-01. |
| ADR012-R1-10 lifecycle and merge | **Closed** | `ADR.md:32,63`. |
| ADR012-R1-11 ADR-007/008 amendments | **Closed** | Only line 3 differs from base. Refuse-13 is rejected (`ADR.md:46`). |
| ADR012-R1-12 REQ-166/47/39/161 | **Closed** | `:712,614,704,707`. |
| ADR012-R1-13 collector interface | **Partly closed** | `begin`, `finish` and `reconcile`, the transport and its caps, and the permissions are specified. The scan command and its inputs are not: ADR012-R3-11. |
| ADR012-R1-14 query-only v13 and exits | **Closed** | `ADR.md:46,65`; REQ-10 `:610`. |
| ADR012-R1-15 frozen pending status | **Closed** | `ADR.md:3` and card `:991` carry no pending state. Pending state appears only in the trackers. |
| ADR012-R1-16 STATUS wording | **Closed** | `STATUS.md:9` is conditional. |
| ADR012-R2-01 producer not authenticated | **Closed by redesign** | Protected source, literal pin, default base only, unprivileged sandboxed scan, privileged jobs never check out PR code (`ADR.md:16,20–22`). The caller-file gate that the R2 fix relied on was cut: ADR012-R3-04. |
| ADR012-R2-02 run-scoped artifacts | **Partly closed** | Per-attempt name, binding to the scan-job time window, and the artifact ID and digest are present (`ADR.md:24`). Uploads from other jobs and partial re-runs remain: ADR012-R3-04. |
| ADR012-R2-03 workflow-directory refusal | **Closed** | `ADR.md:16`; delivery proof at `:77`. |
| ADR012-R2-04 head / merge-base identity | **Closed for identity** | `ADR.md:30`. The change to scanning H alters what is measured: ADR012-R3-03. |
| ADR012-R2-05 terminal matrix, `started_at` | **Closed** | `ADR.md:32,36–37`. A scan in a re-run without registration is not classified: ADR012-R3-04. |
| ADR012-R2-06 retention deadline | **Closed** | `ADR.md:38`. |
| ADR012-R2-07 REQ-22 clauses, ADR-000/003 | **Closed** | All four clauses are restored verbatim (REQ-22 `:589` against base). ADR-000 and ADR-003 differ only on line 3. The amendment list is at `:69`. One amended item is missing: ADR012-R3-10. |
| ADR012-R2-08 caller permissions, policy key | **Closed** | `ADR.md:16,26`; REQ-161 `:707`; `policy-file.md:68–74`. Rollout remainders: ADR012-R3-01 and -08. |
| ADR012-R2-09 ADR-011 rollback, STATUS range | **Partly closed** | No ADR-011 content remains, and `STATUS.md:7` is conditional. `STATUS.md:129` still says "ADR-000…ADR-010 … govern the product": ADR012-R3-13. |
| ADR012-R2-10 local events | **Closed** | `ADR.md:44`. |
| ADR012-R2-11 interval reproducibility | **Closed for intervals** | `ADR.md:40,50`. Triage and cutoff reproducibility: ADR012-R3-05. |
| ADR012-R2-12 CLI version | **Closed** | `ADR.md:65`. |
| ADR012-R2-13 inspection edge cases | **Closed** | `ADR.md:52`. Owner-decided listing: ADR012-R3-10. |
| ADR012-R2-14 no-op pushes, concentration | **Closed** | Tree-OID key and `distinctPRs` (`ADR.md:50`; REQ-22). |

## Issues

### ADR012-R3-01: The separate shadow pin can publish schema 14 and manifest v2 while older Scout workflows still read and write `scout/state` (major)

- **Location:**
  - `ADR.md:16` (literal release pin of the shadow caller), `:26` (rollout covers only the policy key), `:63` ("older binaries refuse"), `:77` ("v2/downgrade refusal" proof);
  - REQ-168 `:714` ("without changing existing guardian callers");
  - `scout-guardian-trusted.yml:111–118` (`set -euo pipefail`; `state restore` before every publish);
  - `export.go:693–694` (old import refuses a non-v1 schema);
  - `merge.go:128–136` (old `MergeState` re-renders the manifest from *local* metadata and never checks the schema of the current tree).
- **Problem:**
  - The shadow caller carries its own pin, separate from the guardian pins. The first shadow publish converts `scout/state` to manifest v2.
  - A guardian trusted lane on an older pin then fails at `state restore`, which stops baseline publishing. An older PR lane can no longer restore the baseline.
  - Worse, an older publisher that restored before the conversion and then retries its compare-and-swap against the v2 tip regenerates a **v1** manifest. That manifest still lists the copied `pr-runs.jsonl` but drops the v2 metadata (intervals, gaps, legacy marker, pinned catalogue).
  - Under `ADR.md:63`, a v2 binary then "normalizes" that v1 import to empty facts. The facts are lost without any error [INFERENCE from the cited code paths].
  - REQ-168's "without changing existing guardian callers" is false for pins: every caller must move to the schema-14 release.
- **Required fix (recommended option: coordinated cutover with refusal, never silent normalization):**
  1. Release notes, the adopter checklist and the CI reference (REQ-161/168) require every workflow that shares `scout/state` (guardian PR, guardian trusted, shadow) to pin at least release N, the schema-14 release, before the shadow caller is enabled. Pre-N runs must not be in flight or re-run after the cutover.
  2. The shadow writer never performs the first v1→v2 publication. If the current manifest is v1, `begin` and `finish` publish nothing and record a collection gap with reason `guardian release predates schema 14`, until a trusted guardian run at N or later has migrated the state.
  3. Binaries at N or later refuse, with exit 3, a v1 manifest that lists any v2-only file such as `pr-runs.jsonl`. Such a manifest is a downgrade and is never normalized. On every compare-and-swap retry they re-check the current manifest schema and refuse to publish over a newer one. The remedy, named in the error, is to republish from the last v2 commit on `scout/state`.
  4. Reword REQ-168 to "without changing existing callers' permissions or flags; their release pins must be N or later".
  5. Add the mixed-version publish race to the delivery proofs at `:77`.

### ADR012-R3-02: No triage path exists for CI would-block events, and "authorized triage" is undefined (major)

- **Location:**
  - `ADR.md:57` ("current authorized triage"), `:22` and `:63` ("retains authorized triage"), `:77` ("version-specific event triage" proof);
  - REQ-22 `:589` ("maintainers triage every persisted `would_block` event");
  - `would_block.go:79` (`TriageWouldBlock` has no non-test caller, per the peer);
  - `scout triage` acts on issues, not on would-block events (`cmd/scout/main.go:537–597`, per the peer);
  - REQ-26 `:611` (only trusted default-branch runs publish);
  - `ShadowReady` refuses while `Undecided > 0` (`policy.go:422–424`, per the peer).
- **Problem:**
  - Under ADR-012, events are created only by the shadow writer, on `scout/state`. No maintainer can durably triage them, because no command writes event triage and no trusted publish path carries it.
  - Every check that has events therefore stays not-ready for good. Only the rare-security provisional route remains.
  - Nothing defines who is "authorized".
  - This is a readiness input of the gap ADR-012 claims to close (STATUS `:121`).
- **Required fix (recommended option):**
  - **Path:** add a `workflow_dispatch` triage entry to the protected shadow caller, with inputs event ID, verdict and reason. Pass them to steps through `env`, never interpolated.
  - **Authorization:** the provider-authenticated `github.triggering_actor` is checked against base-policy `maintainers`/CODEOWNERS, as `AcceptOverride` and REQ-23 do. Anyone else is rejected and logged.
  - **Storage:** the writer job appends a row to a schema-14 `would_block_triage` history table (event ID, verdict, actor, reason, time). Current triage is the latest authorized row. Export the rows append-only in v2 and union them on merge. Publish through the existing compare-and-swap.
  - **Records:** update REQ-22/27/48/161, the CLI reference and `:77`. Add a delivery proof that a non-maintainer is rejected.

### ADR012-R3-03: Shadow measures a different comparison, and a different set of PRs, from the gate it will activate (major)

- **Location:**
  - `ADR.md:30` (scan H; `merge_base = merge-base(B,H)`);
  - REQ-27 `:612` ("Assess actual fetched PR head, not synthetic merge commit", without scope);
  - `scout-guardian.yml:55,58,102–103`: the guardian lane checks out the merge commit M (no `ref`) and diffs against `base.sha`, so its merge-base is the base tip;
  - GitHub: `pull_request` workflows do not run when a PR has a merge conflict, and `pull_request_target` workflows do ([events documentation](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#pull_request)).
- **Problem:**
  1. The readiness evidence (H against an older merge-base) is not the procedure that gates after activation (M against B).
     - Shadow PRs that have not been rebased mostly fall back to paired recompute at an old revision, while the guardian lane can use the baseline at the base tip [INFERENCE].
     - Base lane health, and so base-gap unknowns, then differ. The 10% bound is estimated for the wrong procedure.
  2. PRs with merge conflicts enter the shadow window, but the guardian `pull_request` lane never runs on them.
  3. Read literally, REQ-27's unscoped sentence also redefines the guardian PR lane.
  - Round 2's fix (ADR012-R2-04) kept M as the assessed revision. The rewrite to H was not reviewed as a change of what is measured.
- **Required fix (recommended option):**
  - Inside the sandbox, compute the merge of (B, H) as data only, with `git merge-tree --write-tree` and built-in merge only. Assess the result against B, which is the guardian lane's comparison.
  - Record the merge tree OID as the assessed revision (provenance). The deduplication key stays as it is.
  - A merge conflict is `not_assessed`, reason `merge_conflict`, disclosed.
  - State the equivalence with GitHub's test merge for non-conflicting PRs as [ASSUMPTION]. Add a delivery proof that shadow and guardian agree on events and comparison for the same PR head.
  - Scope REQ-27's sentence to the shadow scan, and state that the guardian lane is unchanged.
  - *Rejected alternative:* keep H with an [ASSUMPTION] label and exclude conflicted heads. That is cheaper but keeps the base-revision mismatch.

### ADR012-R3-04: Run integrity is asserted but not enforced: other jobs in the caller run, and partial re-runs with no registration (major)

- **Location:**
  - `ADR.md:24` ("Only the protected scan job runs in that producer run with upload authority");
  - `ADR.md:16` ("record and validate … job IDs");
  - `round-2-response.md` (cuts "caller-file gates"; "Fixed protected template removes sibling-job forgery");
  - `ADR.md:20` ("Only after this publish succeeds may scan start"), `:36–37`;
  - [Re-running workflows and jobs](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/re-run-workflows-and-jobs): re-running failed jobs reuses successful jobs and their outputs.
- **Problem:**
  1. **Other jobs in the run.** Nothing requires the adopter's caller to contain only the job that calls the pinned reusable workflow, and nothing rejects runs that contain other jobs.
     - Any other job in that caller (a third-party action, or a misguided PR checkout) runs in the same run. It can upload `scout-shadow-<attempt>` inside the scan time window.
     - The round-2 fix depended on the single-job caller gate, and that gate has been cut.
  2. **Partial re-runs.** "Re-run failed jobs" after a scan failure executes scan and finalize for attempt 2 but reuses attempt 1's register job.
     - No intent exists for attempt 2, which contradicts `:20`.
     - Neither "accept" nor "drop" is defined. Dropping it silently, while attempt 1's failure stays as an assessed unknown, biases the unknown ratio upward.
- **Required fix (recommended option):**
  - `begin` parses the caller at the protected source SHA without executing it. It requires exactly one job, which calls the pinned reusable workflow with literal `scout-version` and `release-repo`.
    - Otherwise the attempt is registered `not_assessed`, reason `caller_not_dedicated`, and scan does not run.
    - Readiness reports the configuration as unmeasured with that reason.
  - `finish` and `reconcile` list the attempt's jobs (`actions: read`). Any job outside the template's register, scan and finalize jobs makes the attempt `untrusted_producer`.
  - The scan job's trusted preamble, before any PR fetch or sandbox, reads `scout/state` (read-only) and verifies that an intent exists for the current `(run_id, run_attempt)`.
    - If none exists, scan exits before assessment. The attempt is `not_assessed`, reason `unregistered_attempt`, disclosed.
    - Documentation tells users to re-run all jobs.
    - A re-run of finalize only finalizes the originally registered attempt.
  - Add delivery proofs for both cases.

### ADR012-R3-05: Historical `as_of` queries are not reproducible (minor)

- **Location:** `ADR.md:40,63,77` (claims of dossier reproduction); `:57` ("current authorized triage"); `:50` (latest terminal attempt, with no `as_of` restriction); `would_block.go:16` (single triage value and `triaged_at`); REQ-78 `:600` (`dossier_digest`).
- **Problem:**
  - Re-triaging an event after activation changes a past `--as-of` result.
  - A later attempt can replace the representative in an earlier query.
- **Required fix:**
  - Representatives are chosen only from attempts with `started_at ≤ as_of`.
  - Triage is evaluated as of `as_of`, from the append-only triage history added by ADR012-R3-02.
  - State both rules in REQ-78. Separate "historical reproduction" from "re-evaluation with current triage" in the CLI output.

### ADR012-R3-06: Event counting after deduplication is ambiguous (minor)

- **Location:** `ADR.md:44` (links from attempt to event), `:50` (one representative attempt), `:57` ("linked to at least one represented observation"); REQ-22 (zero not-warranted).
- **Problem:** An implementation could join only the representative attempt. A later failed re-run with no links would then hide an earlier attempt's not-warranted event in the same observation.
- **Required fix:**
  - Count events from the union of eligible links across *all* assessed attempts of each represented observation.
  - Comparability and unknown status still come from the representative.
  - State this in `:57`, REQ-22 and REQ-78.

### ADR012-R3-07: Finalize identity and the residual risk of a compromised scanner are not stated (minor)

- **Location:** `ADR.md:22` (accepts output "only from this registered run's fixed scan job/attempt"), `:24`, `:30` (identity fields), `:10` and REQ-51 ("Do not import results from PR-controlled workflows").
- **Problem:**
  - The ADR does not say that the PR key, head SHA, head tree, base SHA and merge-base must come from the registered intent or the provider API, rather than from the payload.
  - The scan output is computed over attacker-controlled data. A scanner exploited inside the sandbox can falsify its own attempt's payload. The sandbox limits escalation, not output integrity.
- **Required fix:**
  - Fields in the deduplication key come from the intent. `merge_base` is re-derived from the provider compare API.
  - Payload values must match, or the attempt is comparison-none with reason `identity_mismatch`.
  - State that the artifact is untrusted, bounded, schema-validated data. A compromised scanner can at most falsify that one attempt's events and statuses, and this residual risk is accepted.

### ADR012-R3-08: `pull_request_target` platform prerequisites and hardening are missing (minor)

- **Location:** `ADR.md:16,21,26,77`; REQ-161 `:707`; REQ-168 `:714`; `policy-file.md:70–72`; [Securely using pull_request_target](https://docs.github.com/en/actions/reference/security/securely-using-pull_request_target).
- **Problem:** GitHub's guidance adds constraints that the contract does not cover:
  - A default event policy blocks `pull_request_target` in public repositories. It is in evaluate mode now and enforced from 2026-12-02 for affected repositories.
  - `actions/checkout` refuses fork PR head refs unless `allow-unsafe-pr-checkout: true` is set.
  - Self-hosted runners must be isolated and ephemeral.
  - Command injection through untrusted event fields is a named risk.

  Separately, the contract does not forbid:
  - interpolating PR-controlled event fields;
  - `secrets: inherit` in the caller;
  - write-capable `cache-mode`;
  - the scanner writing to runner command files (`GITHUB_ENV`, `GITHUB_OUTPUT`, `GITHUB_PATH`, `GITHUB_STEP_SUMMARY`), or its stdout being read as workflow commands.

  When the event is blocked or undelivered, there are simply no attempts, and that is not disclosed.
- **Required fix:** add each of the following to the rollout text (REQ-161/168, policy reference, CI reference) and to the proofs at `:77`:
  1. In public repositories, an Actions event policy must explicitly allow this workflow's `pull_request_target`.
  2. PR objects are fetched with `git fetch` by SHA into a path that is never executed. If `actions/checkout` is used instead, `allow-unsafe-pr-checkout: true` is set, justified as data-only.
  3. Only numeric IDs and validated 40-hex SHAs reach steps, through `env`. Head ref, title, body, labels and fork name never appear in `${{ }}` inside `run:`.
  4. The caller passes no secrets.
  5. Jobs run on GitHub-hosted runners, or on self-hosted runners only if they are ephemeral and isolated.
  6. No job in `scout-shadow.yml` restores or saves a cache, and none declares a write-capable `cache-mode`.
  7. The sandbox has no access to runner command files. Scanner stdout and stderr are captured to files and never read as workflow commands.
  8. Readiness output states that observations exist only for delivered events [ASSUMPTION], and a blocked trigger yields unmeasured, never an empty window that looks complete.

### ADR012-R3-09: Requirements outside the edited set are not reconciled; status-line notes are incomplete (minor)

- **Location:**
  - REQ-26 `:611` ("from trusted default-branch runs only");
  - REQ-51 `:618` ("PR/fork jobs never publish durable baseline state");
  - REQ-162 `:708` ("both workflows … tag-keyed cache"), and REQ-163;
  - ADR-000 and ADR-007 line 3 ("(readiness counting unit)", "(13→14 migration)");
  - `ADR.md:4` and card `:994`.
- **Problem:**
  - The register and finalize jobs are triggered by PR events, fork PRs included, and they publish to `scout/state`. REQ-26 and REQ-51 do not say whether that counts as a trusted default-branch run or as a PR job.
  - REQ-162/163 do not cover how `scout-shadow.yml` installs tools, and REQ-162 still allows a cache.
  - ADR-012 also amends REQ-161 (ADR-000, ADR-007) and REQ-51 (ADR-007). The status-line notes name only one area each.
- **Required fix:**
  - Amend REQ-26 and REQ-51: register and finalize jobs triggered by `pull_request_target` execute no PR content, count as trusted default-branch runs, and write only shadow facts, events and triage, never baseline content (issues, dispositions, export base).
  - Extend REQ-162/163 to `scout-shadow.yml`: the same verified tag install, tools installed before any PR objects are fetched, no cache.
  - Add these REQs to `ADR.md:4` and the card.
  - Make every amended ADR's status note, README row and card relation name each amended area. For example, ADR-000: "(readiness counting unit, shadow workflow)"; ADR-007: "(13→14 migration, shadow publish and install)".

### ADR012-R3-10: One owner-decided change is not listed, and the counting terminology is stale (minor)

- **Location:**
  - `ADR.md:52`: a check with no eligible units is `not_applicable` and is left out of PRRuns, so docs-only PRs no longer count toward that check's 5;
  - `ADR.md:69` (amendment list), which omits this;
  - REQ-78 `:600`: "Recomputed paired-base runs count as comparable-base PR runs" and "each distinct revision observation", while REQ-22 keys observations on the head tree, not the revision;
  - `evidence.pr_runs`, `min_pr_runs`; `policy-file.md:19,23`.
- **Required fix:**
  - Add the eligibility exclusion to `:69` and the card, as a deliberate amendment of ADR-000's "5 comparable-base PR runs".
  - In REQ-78, write "comparable-base observations" and "each distinct (PR key, head tree OID, merge-base) observation".
  - Define `pr_runs`/`min_pr_runs` as PRRuns observations under ADR-012 in REQ-78 and in the planned section of `policy-file.md`.

### ADR012-R3-11: The scan command and its data inputs are not specified (minor; remainder of R1-13)

- **Location:** `ADR.md:21` (no command named); `scout-guardian.yml:102–103` (`--baseline-remote origin` restores over the network, but the sandbox has no network); override parsing in `diff.go` (per the peer).
- **Required fix:**
  - Name the command, for example `scout diff --shadow --base <B> --head <assessed revision> --baseline-dir <export> --out <file>`.
  - List its data inputs:
    - the fetched B and H objects;
    - the merge tree (ADR012-R3-03);
    - the `scout/state` export at a commit recorded in the intent, supplied as data.
  - State that maintainer overrides (`scout-override`, `scout:allow`) are handled exactly as in the guardian lane.

### ADR012-R3-12: Ledger and provenance gaps (minor)

- **Location:** `ADR.md:12` ("under Main's delegated simplification direction"); `ADR.md:73`; R-070 and R-077 (`research-ledger.md:72–73`).
- **Problem:**
  - The design depends on GitHub facts that have no ledger source:
    - the default `pull_request_target` event policy, read-only cache, checkout protection and runner guidance;
    - re-run semantics (reused jobs);
    - artifacts that carry no attempt or job, and the `upload-artifact` v4 same-name conflict (round 2 recorded these only in its own text);
    - `pull_request_target` running on PRs with merge conflicts.
  - "Main" is not a repository concept and has no ledger source.
- **Required fix:**
  - Add these documentation links to R-070, which is unmerged and can still be revised, and cite them in `:73` and the card.
  - Change `:12` to "the agent selected this simplification under R-064".

### ADR012-R3-13: Tracker gaps (minor)

- **Location:** plan `:59–60` (B5 "dogfood in `guardian`" depends only on B4; REQ-168 requires dogfooding through Scout's own shadow window); plan has no implementation row for ADR-012; `STATUS.md:129` ("decisions ADR-000…ADR-010 … govern the product").
- **Required fix:**
  - Add a Track B row for implementing ADR-012 (schema 14, shadow workflow, triage path, readiness command, proofs) and make B5 depend on it.
  - Make `STATUS.md:129` conditional in the same way as `:7`.

## Checks that found no issue

- **Status lines of accepted ADRs.** ADR-000, ADR-003, ADR-007 and ADR-008 differ from base only on line 3 (lead: line 3; experiment-peer: ADR-003 body read in full and identical; challenger: all four bodies unchanged). The notes match the README rows `:16,19,23,24` and the card relations `:736,807,901,925`. The ADR011-R2-03-style ambiguity is absent. (Completeness of the notes: ADR012-R3-09.)
- **REQ-22 fidelity.** The maintainer, during-window, per-repository ("that repo's"), "for that repo" and "zero events at window end" clauses are byte-identical to base. The change of counting unit is declared at `ADR.md:69`.
- **REQ-27 fidelity.** The original attribution clauses are unchanged. (Scope: ADR012-R3-03.)
- **REQ-78 fidelity.** Paired-base comparability and the rule of counting unknown once per observation are kept. The CLI exits match the ADR. (Terminology: ADR012-R3-10.)
- **The simplification kept what the requirements need.** It keeps zero-event facts, durable registration before scan, reconciliation of failed writers, all intervals and gaps, the separation of legacy and local data, pinned cohorts and maintainer-only activation. Each cut (PR-caller authentication, PR-cache inspection, cross-workflow artifacts, fork run→PR inference, retention-derived waits) removes a mechanism that the protected producer makes unnecessary.
- **"No results from PR-controlled workflows are ingested."** This holds consistently across `ADR.md:10,24`, REQ-51 `:618`, REQ-161 `:707`, STATUS `:121`, plan `:88` and `policy-file.md:72`. (The residual risk to the payload: ADR012-R3-07.)
- **Token permissions.** Register and finalize have `contents: write`, `pull-requests: read` and `actions: read`. Scan reduces to `contents: read`, with credentials removed before scanning. No new grant reaches existing guardian callers (`scout-nightly.yml:36–54` unchanged).
- **Migration.** A narrow transactional 13→14 migration with an explicit pre-v14 marker. Query-only v13 exits 3. Versions 1, 2 and 4–12 are still refused under REQ-46. Refuse-13 is rejected with a reason (`ADR.md:46`; REQ-10 `:610`). (Cutover between writers: ADR012-R3-01.)
- **Ledger rows.** R-070's claims match GitHub's documentation. R-077's code citations hold (experiment-peer). IDs do not collide: R-071 to R-076 belong to other units.
- **Numbering and README.** ADR-012 is used with ADR-011 reserved, and no ID is reused. The README note sits outside the rules list.
- **Dependabot.** Dependabot PRs into the default branch get the normal `pull_request_target` context, so the claim that action-ref input can be measured holds (experiment-peer).

## Considered and not raised

- **Cache poisoning in the default-branch scope (lead's draft major).** The lead feared that the scan job, running under `pull_request_target`, could write default-branch cache entries that `scout-guardian-trusted.yml:98–102` restores and then executes (`bootstrap-tools.sh:94`). GitHub now gives `pull_request_target` workflows read-only access to the default-branch cache unless a write-capable `cache-mode` is declared ([guidance](https://docs.github.com/en/actions/reference/security/securely-using-pull_request_target)). The remaining hardening is in ADR012-R3-08, items 6–7.
- **Counting distinct PRs instead of observations (experiment-peer N3).** Any author can push trivially different trees to reach the minimum or dilute the unknown ratio. But the observation unit, with [ASSUMPTION] and `distinctPRs` disclosure, was this council's own fix in ADR012-R2-14. The same dilution was possible under round 2's `pull_request` producer, and round 1's fix for R1-06 accepted it. Activation stays with maintainers, who see `distinctPRs`. Changing the owner-decided unit again is a design preference, not a defect.
- **Contention on `scout/state` from writes that outsiders trigger.** Each attempt publishes twice. But compare-and-swap re-reads the tip on each of 8 attempts (`baseline.go:28,84–97`). A failed shadow publish leaves a pending intent for reconciliation, and a failed baseline publish recovers on the next trusted run.
- **Fork PRs from outsiders count toward the minimum.** Unchanged from before ADR-012 (REQ-179 PR C is a fork).
- **Selection bias from excluding cancellations.** This is the classification the council chose in R2-05, and it is disclosed. The peer's suggestion to break cancellations down by PR size is noted for the implementation.
- **`round-2-response.md` in the signed set (experiment-peer).** Naming the digest of an extra reviewed record does no harm. Route (a) at `ADR.md:75` sets the minimum set.
- **Retargeted PRs.** A PR whose base is changed to the default branch fires `edited`, so it is not assessed until its next push. This is a sampling detail covered by ADR012-R3-08 item 8.
- **PR-controlled non-executable configuration (`.semgrepignore`, etc.).** The guardian lane is equally exposed, so this is not specific to ADR-012.
- **R-064 is absent from this branch's ledger.** The dependency is stated (`ADR.md:73,75`; STATUS `:131`). It affects only whether a sign-off is final (see Verdict).

## Peers and where they changed the outcome

- **`innovation-challenger`** raised 3 major and 3 minor findings and confirmed that the core prior issues are resolved at the level of the design contract. Adopted:
  - the `MergeState` downgrade path (`merge.go:128–136`). This turned the lead's restore-failure observation into ADR012-R3-01 and made R1-09 partly open again;
  - partial re-runs that skip registration (into ADR012-R3-04);
  - non-reproducible dossiers (ADR012-R3-05, severity lowered from major to minor, consistent with R2-11; the fix shares R3-02's history table);
  - the union of event links (ADR012-R3-06);
  - the event-policy prerequisite (into ADR012-R3-08);
  - `STATUS.md:129` (ADR012-R3-13).

  Its pointer to GitHub's `pull_request_target` guidance led the lead to read it, and the read-only cache rule there **removed the lead's draft major** on base-scope cache poisoning.

  Not adopted: its alternative design, a scheduled snapshot collector (see Alternatives).
- **`experiment-peer`** raised 4 major and 5 minor findings. Adopted:
  - the missing triage path (ADR012-R3-02, new to the lead);
  - the H-versus-M measurement gap, with the merge-conflict evidence and the `git merge-tree` fix. This **raised the lead's minor** REQ-27 scope note to major (ADR012-R3-03);
  - the caller-dedication gap (ADR012-R3-04);
  - REQ-26/51/162/163 (ADR012-R3-09);
  - the `not_applicable` listing and stale terms (ADR012-R3-10);
  - the scan command (ADR012-R3-11);
  - ledger and "Main" (ADR012-R3-12);
  - the B5 dependency (ADR012-R3-13).

  Not adopted: N3 (distinct-PR gating) and removing the response from the signed set.
- **Lead-only findings:** ADR012-R3-07 (identity from the intent, residual scanner risk) and the injection, secrets, runner, cache-mode and command-file hardening in ADR012-R3-08.

## Council memo

1. **Recommendation.** Do not sign off this round. Keep the simplified, protected, event-driven design: it removes every round-2 blocker. Fix the 13 issues, then submit round 4 on these canonical files. After the rebase onto ADR-011, submit a final round on the resolved bytes.

2. **Alternatives.**
   - **(A, recommended)** Protected `pull_request_target` register → sandboxed scan of the merge tree → trusted finalize, plus the fixes above. It observes every head and matches the gate's procedure.
   - **(B)** A scheduled snapshot collector that freezes current PR heads from metadata (challenger). It needs no `pull_request_target` or event-policy opt-in. But it misses intermediate heads and measures late, which changes the sampling unit, and it still needs a sandbox.
   - **(C)** Return to PR-produced artifacts with authentication (the round-2 recipe). Rejected: it has a larger trust surface and needs cache removal and caller parsing.
   - **(D)** Keep scanning H with an [ASSUMPTION] label (fallback for R3-03). Cheaper, but the readiness evidence measures a different base revision from the gate.
   - **(E)** Count distinct PRs (peer N3). Stricter, but it amends the owner-decided unit again and reverses R2-14.

3. **Evidence.**
   - Mixed-version cutover: `scout-guardian-trusted.yml:111–118`, `export.go:693–694`, `merge.go:128–136`.
   - Triage path: `would_block.go:1–103`; REQ-22 `:589`.
   - Measurement procedure: `scout-guardian.yml:55–58,102–103` against `ADR.md:30`.
   - GitHub sources:
     - [pull_request_target guidance](https://docs.github.com/en/actions/reference/security/securely-using-pull_request_target): event policy, read-only cache, checkout protection, runners, injection;
     - [re-run semantics](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/re-run-workflows-and-jobs);
     - [merge-conflict trigger behaviour](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#pull_request).
   - Status lines: line 3 of each of the four ADRs, worktree against base.

4. **Risks.**
   - The sandbox remains an unimplemented delivery prerequisite. Its failure lets a scanner exploit falsify payloads, which ADR012-R3-07 bounds.
   - GitHub platform policy is moving: event-policy enforcement starts 2026-12-02.
   - `git merge-tree` could differ from GitHub's test merge in edge cases (stated as [ASSUMPTION] and proven by delivery test).
   - The coordinated cutover needs operator discipline. The refusals in ADR012-R3-01 make mistakes fail loudly instead of losing data.
   - The dispatch-based triage interface is new.
   - Reversibility is high: the change is design-only, shipped schema stays 13, and a superseding ADR can replace it.

5. **Execution outline.**
   1. Specify the cutover: pins at release N or later, no first v2 publish by the shadow writer, refusal of downgraded v1 manifests on import and on every compare-and-swap retry, the REQ-168 wording, and the proof (R3-01).
   2. Add dispatch triage with `triggering_actor` authorization and the append-only history in schema 14. Thread it through REQ-22/27/48/161 and the CLI reference (R3-02). Derive the `as_of` triage and representative cutoff from it (R3-05).
   3. Switch the assessed revision to the merge tree computed in the sandbox, with `merge_conflict` as `not_assessed`. Scope REQ-27 (R3-03).
   4. Add caller dedication in `begin`, the foreign-job check in `finish`/`reconcile`, and the scan preamble's check that an intent exists for the current attempt (R3-04).
   5. Add the union of event links (R3-06), identity taken from the intent (R3-07), and the eight hardening and prerequisite items (R3-08).
   6. Reconcile REQ-26/51/162/163, the status notes, the README and the cards (R3-09). Fix the amendment list and terms (R3-10).
   7. Name the scan command and its inputs (R3-11). Extend R-070 and fix the "Main" credit (R3-12). Update the plan and STATUS (R3-13).
   8. Submit round 4 with new digests from Main. After ADR-011 merges, rebase and submit the final round on the resolved bytes.

## Verdict

**ADR-012: 13 issues open (4 major, 9 minor). Not signed off.** Open: ADR012-R3-01, -02, -03, -04, -05, -06, -07, -08, -09, -10, -11, -12 and -13.

The digests in the table above identify the bytes that were reviewed. They carry no sign-off. Any later zero-issue round on pre-rebase bytes cannot be the final adoption round, because ADR-012 cites R-064, which is absent from this branch's ledger, and `ADR.md:75` requires review of the final bytes after the rebase onto ADR-011.
