# Round 4: ADR-012 (shadow-window readiness inputs)

Date: 2026-11-08. This round follows design rules 1–3 and the ledger rules in `CLAUDE.md`, and the ADR rules in `docs/adrs/README.md`. ADR-012 is an agent decision taken under the owner's standing delegation (R-064). This is the first round on the integrated bytes. The branch `agent/policy-shadow-adrs` has been fast-forwarded to origin/main `7bb73e52`, where ADR-011 is adopted (PR #17). Where a fix involves a choice, the council names the option it recommends under that delegation.

The council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review of `pull_request_target` and dispatch safety, internal consistency, mechanism soundness and rebase hygiene.
- **`experiment-peer` (Opus):** closure of ADR012-R3-01..13 element by element, fidelity to REQ-22, REQ-27 and REQ-78, counting confounds, ledgers, trackers and status lines.

The review was read-only. Nobody edited files, executed product code, ran model calls outside the council, or touched `.codegraph`.

**Main computed and verified the digests below with `shasum`. The council did not, because the lead and both peers had read-only tools and no shell.** This round is not a sign-off, so the digests only identify the bytes that were reviewed. All paths are under `docs/`.

| File | SHA-256 (Main) |
|---|---|
| `adrs/2026-11-08-shadow-readiness-inputs/ADR.md` | `5768cc9ba5d9c58a73f985f182b63f1d41bd796fbf5fe51e643345b9cc9eca04` |
| `scout-design.html` | `ec07b45c3ace38f74823d533286e298ca806595e5e342f197267ccd6743d0128` |
| `adrs/README.md` | `64778def497063c7909e68f9b06f8085e598bd76fe344bd4e80f35dce0f4e65b` |
| `STATUS.md` | `5f5e9d578b1c1ae85d25ca204eb9ba06c564bdcb59689d7391bad4987bd1668e` |
| `plans/2026-11-06-remaining-work.md` | `91a24902029eb06e3635c38baf2ce5b402e0db219acd6c99b18e210b1022a493` |
| `references/policy-file.md` | `f5122b885e83fbcaf69f8834dc888a97c2e9656951f96151c9993979eceaef7a` |
| `references/ci-integration.md` | `2b03080791670581194fedb433f32813d0566ca7233d6547157048deebc6f322` |
| `references/cli-commands.md` | `f101ebefe1bad8d1cd33c5f691b97a306f40ed821fff90dfe159260b5fc614cb` |
| `research-ledger.md` | `b156e8e4405a24782f5cf91cd03d44467ff5b349b06125e21eefa83c1a356b56` |
| `adrs/2026-10-19-v1-decisions/ADR.md` (ADR-000) | `368c9c04ecb8fd3a1eeaf659ec321fd4c15b7b07292b86a15245b0424c4b5c4e` |
| `adrs/2026-10-23-decisions/ADR.md` (ADR-003) | `22f284b7ebec6271fb7c6398da2b4b809ab3c73af441f18dc65351046d338707` |
| `adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) | `144d42400d90158942c612fc4940979bd66deeaeffe32fab03ffad6195277ba3` |
| `adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) | `80e3375f81656a8856cf662cbec99a36d8e00bc451ff3c3592925d2293f44d95` |

The round-3 response record is not signed. Its SHA-256 (Main) is `6f3ffda7ce7411a0db33e13d3d6545dcf6ac17f672c614de2e8b664e09a663eb`.

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11`, branch `agent/policy-shadow-adrs` at origin/main `7bb73e52`. The tracked changes are staged and uncommitted, and the ADR-012 folder is untracked.
- Base bytes from `/Users/adamjackson/Projects/work/scout-adr11b` (`2b95ac74`, whose docs match main).
- Diff stat (Main): 12 files changed, +103/−36.

**Files read:**
- ADR-012 `ADR.md`, all lines. Line numbers below refer to the raw file: `:16` workflow, `:20–22` phases, `:24` transport, `:26` authority, `:28–30` cutover, `:32` hardening, `:36` identity, `:50–56` events and triage, `:60–69` counts, `:73–75` export and CLI, `:79` amendments, `:83–87` sources and delivery.
- `round-3.md` and `round-3-response.md`.
- `scout-design.html`, worktree against base:
  - REQ-10/22/26/27/39/46/47/48/51/78/161/162/163/166/168 (`:589,600,610–618,704,707–709,712,714`);
  - the Activation callout `:449`;
  - the cards for ADR-000 `:736`, ADR-003 `:807`, ADR-007 `:901`, ADR-008 `:925`, ADR-011 `:991–1004` and ADR-012 `:1006–1012`.
- `adrs/README.md`, all lines, worktree against base.
- Line 3 of ADR-000, ADR-003, ADR-007 and ADR-008, worktree against base. Their bodies were not re-read this round.
- ADR-011 `ADR.md:1–8`.
- `STATUS.md:7,9,120–121,129,131`; plan `:15,20–61,88–99`; `ci-integration.md:10–14,69–114`; `cli-commands.md:24–28,120–133`; `policy-file.md:17–27,76–86`; `research-ledger.md:3–6,72–76` (both trees).
- Code:
  - `.github/workflows/scout-guardian.yml:95–110`;
  - `scout-guardian-trusted.yml:3–65,192–203`;
  - `cmd/scout/diff.go:15–40`;
  - `internal/engine/gate.go:309–343,378–392`;
  - `internal/policy/policy.go:395–458`.
- External:
  - GitHub [events that trigger workflows (`workflow_dispatch`)](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#workflow_dispatch);
  - [contexts reference](https://docs.github.com/en/actions/reference/workflows-and-actions/contexts) (`actor` against `triggering_actor` on re-runs);
  - [re-running workflows](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/re-run-workflows-and-jobs);
  - [Securely using pull_request_target](https://docs.github.com/en/actions/reference/security/securely-using-pull_request_target);
  - [`git merge-tree`](https://git-scm.com/docs/git-merge-tree), read by the challenger.

## Issue summary

**16 distinct issues are open: 2 major and 14 minor.**

| ID | Severity | Summary |
|---|---|---|
| ADR012-R4-01 | major | REQ-22's owner-decided per-repository auto-demotion clause regressed from "that repo's check" to an ambiguous "it". This reverts main's ADR011-R2-03 wording and contradicts `ADR.md:79` |
| ADR012-R4-02 | major | Dispatch triage and manual reconcile are not bound to the default branch, to current protected policy, or to a single dispatch. A re-run can let a non-maintainer author a verdict or can silently revert a newer verdict |
| ADR012-R4-03 | minor | The triage cutoff is not publication-safe: rows stamped before `as_of` can arrive later, and the `current_window` triage cutoff is unstated |
| ADR012-R4-04 | minor | Representatives, the event union and Days have no lower bound at the start of the interval, and REQ-78 limits `started_at≤as_of` to "historical" queries |
| ADR012-R4-05 | minor | Shadow event IDs change when B advances inside one observation, so the union counts duplicates |
| ADR012-R4-06 | minor | The override inputs contradict the guardian lane: the guardian passes none, but shadow "freezes provider provenance". Event parity fails |
| ADR012-R4-07 | minor | The schema-refusal remedy treats a downgrade and a newer schema the same way and prescribes rollback for both (council's own R3-01 wording) |
| ADR012-R4-08 | minor | The sandbox's writable scope, and the integrity of trusted steps that run after it, are unspecified, so the stated residual-risk boundary is not guaranteed |
| ADR012-R4-09 | minor | Per-check `not_inspected`/`not_applicable` observations are not disclosed, so a provisional zero-event result can rest on uninspected observations |
| ADR012-R4-10 | minor | REQ-10 omits the `would_block_triage` history table and how v13 triage migrates |
| ADR012-R4-11 | minor | REQ-51's opening clause is weaker than base, which said "PR/fork state access is read-only" |
| ADR012-R4-12 | minor | Amended-area names are incomplete, and the README's ADR-000 row drops ADR-001–003 |
| ADR012-R4-13 | minor | The design overview's Activation callout `:449` still counts PR runs |
| ADR012-R4-14 | minor | REQ-168's release list omits `scout-shadow.yml` |
| ADR012-R4-15 | minor | Trackers: the plan graph lacks B4-S→B5, plan `:15` omits B4-S, and STATUS `:7` "Latest amendment" stays ADR-011 |
| ADR012-R4-16 | minor | Ledger gaps: dispatch, contexts and `merge-tree` sources, plus the repository facts behind the round-3 decisions |

## Closure of round-3 issues

| Issue | Status | Evidence / remainder |
|---|---|---|
| ADR012-R3-01 cutover | **Closed** | Release floor N, quiescence, guardian-first v2 publish, the shadow gap on v1, hybrid refusal and a per-retry recheck with exit 3 (`ADR.md:28–30`; REQ-48 `:615`; REQ-161 `:707`; REQ-168 `:714`; `ci-integration.md:103–104`; `policy-file.md:80`). The mixed-version proof is at `:87`. The remedy wording copied from the council's R3-01 text is wrong for newer schemas: ADR012-R4-07. |
| ADR012-R3-02 triage path | **Closed as specified** | Dispatch, `env` inputs, `triggering_actor` against maintainers/CODEOWNERS, append-only history, export, union and CAS (`ADR.md:54–56`; REQ-22/27/48/161; `cli-commands.md:130`; proofs `:87`). The authorization binding gaps are new: ADR012-R4-02. |
| ADR012-R3-03 merge tree | **Closed** | Sandbox `git merge-tree --write-tree`, built-in merge only, assessed against B, OID recorded, `merge_conflict` is `not_assessed`, [ASSUMPTION] plus parity proof (`ADR.md:21,36`). REQ-27 `:612` says "Shadow alone … guardian behavior is unchanged". Remaining threats to parity are ADR012-R4-05 and -06. |
| ADR012-R3-04 run integrity | **Closed** | `caller_not_dedicated` (`:20`), the foreign-job check (`:22`), the current-attempt preamble with `unregistered_attempt` (`:21`), finalize-only re-runs (`:22`), re-run-all (`ci-integration.md:112`) and proofs (`:87`). |
| ADR012-R3-05 `as_of` | **Partly closed** | `started_at≤as_of` and the triage cutoff are in `ADR.md:56,60`, and the CLI separates `historical_reproduction` from `current_window` (`:75`). REQ-78 qualifies the rule as "Historical" (ADR012-R4-04). Triage cutoff publication: ADR012-R4-03. |
| ADR012-R3-06 event union | **Closed** | `ADR.md:60,67`; REQ-22; REQ-78; `policy-file.md:82`. Duplicate IDs inside the union: ADR012-R4-05. |
| ADR012-R3-07 identity | **Closed** | `ADR.md:22,24`; REQ-27; REQ-51. |
| ADR012-R3-08 hardening | **Closed** | All 8 items are in `ADR.md:32`, REQ-161, REQ-168, `ci-integration.md:103–110` and `policy-file.md:86`, with proofs at `:87`. New remainders: ADR012-R4-02 (dispatch ref) and ADR012-R4-08 (sandbox write scope). |
| ADR012-R3-09 REQ reconciliation | **Partly closed** | REQ-26/51/162/163 are amended, and `ADR.md:4` and the card list 14 REQs. REQ-51's base clause was weakened (ADR012-R4-11), and some areas are unnamed (ADR012-R4-12). |
| ADR012-R3-10 terminology | **Closed at the named locations** | `ADR.md:79` and the card list the eligibility exclusion. REQ-78 uses observation terms and defines `pr_runs`, and so does `policy-file.md:82`. The same stale term remains in the overview `:449`: ADR012-R4-13. |
| ADR012-R3-11 scan command | **Closed as worded** | The command and its inputs are at `ADR.md:21` and `cli-commands.md:127`. The override clause written to the council's own wording contradicts the guardian invocation: ADR012-R4-06. |
| ADR012-R3-12 ledger | **Closed for the required elements** | R-070 has the secure-guidance, re-run, artifact REST, upload v4 and conflict-trigger links (`research-ledger.md:75`), and `ADR.md:83` and the card cite it. No "Main" credit remains in the signed files (grep, no match). New gaps: ADR012-R4-16. |
| ADR012-R3-13 trackers | **Partly closed** | The B4-S row is at plan `:60`, B5 depends on B4 and B4-S at `:61`, and STATUS `:129` is conditional. The plan dependency graph `:31–35` still shows only `B4 --> B5`: ADR012-R4-15. |
| ADR011-R2-04 (shared, deferred to the post-rebase review) | **Closed** | ADR-012 makes no content-binding claim. ADR-011's reset-gap sentence in REQ-22 and its REQ-78 sentences are byte-identical to base (`:589,600`). |
| ADR011-R2-03 (adopted wording on main) | **Regressed** | The "that repo's check" wording required by ADR011-R2-03 and present on main was replaced: ADR012-R4-01. |

## Rebase onto adopted ADR-011

- **R-064 and R-069** (`research-ledger.md:73–74`) are byte-identical to base (lead, raw read of both trees; challenger and peer agree). R-070 and R-077 are new rows at `:75–76`. Ledger header line 4 changes only its R-064 note ("lands with ADR-011 first" → "was adopted with ADR-011 (PR #17)"). That is a header, not a row.
- **The ADR-011 card** (`scout-design.html:991–1004`) is byte-identical to base `:991–1004`. The ADR-012 card is appended after it.
- **The README ADR-011 row** (`adrs/README.md:29`, base `:27`) is byte-identical. The ADR-011 note paragraph `:12` is updated to say ADR-011 was adopted. That is not a row change.
- **Duplicates:** each of R-064, R-069, R-070 and R-077, the ADR-011 and ADR-012 cards, and the ADR-011 and ADR-012 README rows occurs once (grep).
- **Main's bytes not kept:** one sentence of REQ-22 that carries ADR-011 attribution (ADR012-R4-01).

## Issues

### ADR012-R4-01: REQ-22's per-repository auto-demotion clause regressed (major)

- **Location:** `scout-design.html:589` against base `:589`; `ADR.md:79` ("Restore the existing maintainer/window/per-repository/window-end clauses verbatim"); `round-2.md:147–148` (ADR011-R2-03).
- **Problem:**
  - Base and main read "At least 3 confirmed-unwarranted blocks in the trailing 20 blocks auto-demote **that repo's check** to report for that repo and record it".
  - The worktree reads "auto-demote **it** to report for that repo". The preceding sentence ends "the `@N` constant, an unenforced convention", so "it" has no correct antecedent.
  - ADR011-R2-03 required exactly this wording to remove the same ambiguity, and main adopted it. ADR012-R2-07 (major) made restoring the owner-decided REQ-22 clauses a requirement, and round 3 confirmed them byte-identical. The regression entered after round 3.
  - It changes owner-decided text, contradicts `ADR.md:79`, and fails to keep main's bytes in a sentence attributed to ADR-011.
- **Required fix:** restore the base bytes verbatim: "…auto-demote that repo's check to report for that repo and record it [ASSUMPTION]…". Make no other change to that sentence.

### ADR012-R4-02: Dispatch triage is not bound to the default branch, to current protected policy, or to a single dispatch (major)

- **Location:** `ADR.md:16,26,54,56`; REQ-22 `:589`, REQ-26 `:611`, REQ-27 `:612`, REQ-161 `:707`; `ci-integration.md:105,114`; `cli-commands.md:130`; precedent at `scout-guardian-trusted.yml:62`.
- **Problem:**
  1. **No default-branch gate.**
     - GitHub lets `workflow_dispatch` run against any branch or tag. The run uses that ref's workflow file and checkout ([events documentation](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#workflow_dispatch)).
     - The source and caller validation in `begin` covers only PR registration. The triage and manual-reconcile dispatch modes have no equivalent, yet REQ-26 classifies them as trusted default-branch jobs.
     - The existing trusted lane enforces `github.ref == refs/heads/<default>` (`scout-guardian-trusted.yml:62`). ADR-012 does not reuse that guard.
  2. **Unpinned policy source.** "Trusted base-policy maintainers/CODEOWNERS" does not say which revision.
     - A dispatch against a branch whose policy lists the dispatcher would authorize them [INFERENCE: if the policy is read from the checkout].
     - Re-runs keep the original SHA, so a maintainer who has since been removed can still pass.
  3. **Cross-actor re-runs.** On a re-run, `github.triggering_actor` is the person re-running and the original inputs are reused ([contexts reference](https://docs.github.com/en/actions/reference/workflows-and-actions/contexts)). A maintainer who re-runs a non-maintainer's rejected dispatch records that non-maintainer's verdict and reason as an authorized row.
  4. **Stale re-runs revert verdicts.** Each re-run attempt appends a row with a new writer time. That row becomes "latest" and silently reverts a newer verdict, so a non-malicious re-run can flip the zero-not-warranted condition.
- **Required fix (recommended option: reuse the trusted-lane guard and make dispatches idempotent):**
  1. All `workflow_dispatch` modes (triage and manual reconcile) run writer steps only when `github.ref == refs/heads/<default branch>`, the same guard as `scout-guardian-trusted.yml:62`, and the workflow source SHA is reachable from the protected default-branch tip. Otherwise they reject and log before any writer step.
  2. Load maintainers/CODEOWNERS from the **current** protected default-branch tip at execution time, never from dispatch inputs, the workspace checkout or the run's original SHA. Record that revision as the authorization policy revision.
  3. Require `github.actor == github.triggering_actor`, with that actor authorized. Otherwise reject and log without writing a row.
  4. The triage ID is (repository ID, dispatch run ID). A re-run of a dispatch whose ID already exists publishes nothing.
  5. State the residual risk: collaborators with write access can run arbitrary workflows on other branches with `contents: write`. Integrity of `scout/state` against them rests on the `scout/state` ruleset (REQ-161) and is unchanged by ADR-012.
  6. Update REQ-22/26/27/161, the CI and CLI references, and `:87`. Add proofs that a branch or tag dispatch is rejected, a re-run by a revoked maintainer is rejected, a cross-actor re-run is rejected, and a re-run of an old dispatch is idempotent.

### ADR012-R4-03: The triage cutoff is not publication-safe (minor)

- **Location:** `ADR.md:56` ("writer UTC time"), `:60,67,73,75`; REQ-78 `:600`; `cli-commands.md:133`.
- **Problem:**
  - A row stamped t ≤ `as_of` can reach `scout/state` after a dossier was computed, for example during CAS retries or while a dispatch is queued. A later `--as-of` reproduction then includes it and differs from the dossier.
  - Runner clocks are compared with provider times.
  - Nothing states the triage cutoff used by the default `current_window`.
- **Required fix (recommended option: monotonic publication stamp):**
  - Each triage row carries `recorded_at`. The writer assigns it on the CAS attempt that publishes the row, as max(writer UTC now, the parent commit's collectedThrough + 1 ms), and re-assigns it on every retry. collectedThrough never decreases along `scout/state`.
  - Any row absent from a state whose collectedThrough is C then has `recorded_at` > C. So `recorded_at ≤ as_of ≤ C` selects the same rows from that state and from every later one.
  - Historical triage filters on `recorded_at`. Dispatch time stays provenance.
  - `current_window` uses triage cutoff = collectedThrough.
  - State this in `ADR.md:56,75`, REQ-78 and the CLI reference. Add a proof that a triage row published after a reconcile does not change an earlier `as_of` result.

### ADR012-R4-04: The cutoff bounds are incomplete (minor)

- **Location:** `ADR.md:46,60,66,67`; REQ-78 `:600` ("**Historical** representatives require started_at≤as_of"); `cli-commands.md:133`.
- **Problem:**
  - The window is "the reconciled interval containing as_of". But representatives, the event union and Days are bounded only by `started_at≤as_of`.
  - Attempts from before a gap boundary can therefore supply representatives, events or the first started time for Days. That spans an unreconciled gap and weakens "a genuine unreconcilable gap ends an interval" and "only completely reconciled computed inputs".
  - REQ-78's "Historical" qualifier can be read as exempting the default query.
- **Required fix:**
  - Require `interval.start ≤ started_at ≤ as_of` for representatives, union links and the Days origin, in every query, the default collectedThrough cutoff included.
  - An observation whose only assessed attempts precede the interval is not represented.
  - Reword REQ-78: "Representatives and unioned event links require interval start ≤ started_at ≤ as_of for every query, including the default collectedThrough cutoff."

### ADR012-R4-05: Shadow event IDs multiply inside one observation when B advances (minor)

- **Location:** `ADR.md:36` (B is not in the dedup key), `:50` (ID from `(base, head, check, issue, rule_version)`), `:21` (`--base <B> --head <merge-tree-oid>`); `internal/engine/gate.go:385–388` (`wouldBlockID(base, head, check, issue)`).
- **Problem:**
  - A re-run of all jobs, or a reopen, after B advances keeps the same (PR key, H tree, merge_base) but produces a new B and a new merge tree. That gives new event IDs for the same finding.
  - The union then counts duplicates, which inflates WouldBlock and Undecided and doubles the triage work. "Counts deduplicate events" does not hold.
- **Required fix:**
  - For shadow events, base = `merge_base(B,H)` and head = H tree OID: the observation identity. B and the merge-tree OID stay attempt provenance.
  - Add a proof that a re-run after B advances links the existing event.

### ADR012-R4-06: The override inputs contradict the guardian lane (minor)

- **Location:** `ADR.md:21` ("override inputs … have exactly guardian semantics and are frozen with provider provenance in the intent"), `:36` (parity); `cli-commands.md:126–127`; `scout-guardian.yml:102–105`; `cmd/scout/diff.go:24–25`; `internal/engine/gate.go:322–337`.
- **Problem:**
  - The guardian PR lane passes neither `--override-actor` nor `--override-text`, so its override inputs are empty.
  - An accepted override clears `WouldBlock` (`gate.go:334`).
  - If shadow ingests provider labels or body text, a PR carrying a maintainer override yields zero shadow events while the unchanged guardian reports the finding. Parity fails.
  - The clause follows the council's own R3-11 wording, which did not check the guardian invocation.
- **Required fix (recommended option):**
  - Shadow passes exactly the override inputs that the pinned guardian PR lane passes. Today that is none: `begin` freezes empty override inputs, and `ADR.md:21` and the CLI reference drop "provider provenance" and "override provenance".
  - The parity proof includes a PR whose body or labels carry `scout:allow`/`scout-override`.
  - Wiring overrides into both lanes would need its own ADR.

### ADR012-R4-07: The schema-refusal remedy conflates a downgrade with a newer schema (minor)

- **Location:** `ADR.md:30`; REQ-48 `:615`; `ci-integration.md:104`; `policy-file.md:80`; `cli-commands.md` (planned contract).
- **Problem:** "Refuse a newer one, exit 3; … republish from the last v2 scout/state commit" applies the downgrade remedy to a genuinely newer manifest. Following it would discard state written by a newer release. The council's R3-01 fix introduced this wording.
- **Required fix:**
  - **Downgraded hybrid** (a v1 manifest listing v2-only files): refuse with exit 3. The remedy is to republish from the last intact v2 commit, disclosing that state written after it must be reapplied.
  - **Newer schema** (manifest newer than v2): refuse with exit 3. The remedy is to upgrade the writer to a release that supports that schema, and never to republish an older snapshot over it.
  - Keep the recheck on every CAS retry. Update all five locations.

### ADR012-R4-08: The sandbox's writable scope and the integrity of later trusted steps are unspecified (minor)

- **Location:** `ADR.md:21,24`; REQ-51 `:618`; REQ-163 `:709`; `ci-integration.md:107,109`.
- **Problem:**
  - The ADR forbids tokens, network and runner command files inside the sandbox, and it uploads after the sandbox exits. It does not forbid the sandbox from writing files that later trusted steps in the same job execute: installed tools under `$RUNNER_TEMP`, runner action directories, `$HOME` and git configuration.
  - A compromised scanner that tampers with these gains `ACTIONS_RUNTIME_TOKEN` in the upload step [INFERENCE]. It could then upload another attempt's artifact name ahead of time, which turns attempt 2 into comparison-none.
  - That breaks the accepted residual risk at `:24`, which says a compromised scanner "cannot change … other attempts".
- **Required fix:**
  - The sandbox can write only a fresh per-attempt output directory and a scratch git object directory. Tools, runner and action directories, `$HOME` and git configuration are read-only or not mounted.
  - Post-sandbox trusted steps execute only files outside the sandbox-writable paths and read only the bounded output file.
  - Add this to `ADR.md:21`, REQ-51/163 and CI checklist item 7. Add a proof that a sandboxed process cannot modify any file that a later step executes.

### ADR012-R4-09: Per-check inspection status is not disclosed (minor)

- **Location:** `ADR.md:62,64,65,69,75`; `cli-commands.md:133`; `policy.go:416–418` (the provisional return comes before the PRRuns check).
- **Problem:**
  - `not_inspected` observations enter TotalRuns with zero events and are not counted as unknown ("other meanings of unknown do not count").
  - A rare-security provisional result can therefore rest on one inspected observation (for Days) plus any number of uninspected ones, and the output does not show this.
- **Required fix (disclosure only, no rule change):**
  - Readiness reports per-check counts of represented observations that were `inspected`, `not_inspected` and `not_applicable`.
  - The provisional reason states the inspected count.
  - Update `ADR.md:75`, REQ-78 and the CLI reference.

### ADR012-R4-10: REQ-10 omits the triage history table (minor)

- **Location:** REQ-10 `:610` against `ADR.md:52,56`.
- **Problem:**
  - REQ-10 lists facts, links, event metadata and intervals, but not `would_block_triage`.
  - It does not say how v13 triage values become legacy-provenance history.
- **Required fix (recommended option):**
  - Add the append-only `would_block_triage` table to REQ-10.
  - State that the 13→14 migration copies each v13 triage value into one history row with legacy provenance and its original `triaged_at`, and that the derived latest-authorized view replaces direct reads of the old column.

### ADR012-R4-11: REQ-51's base clause was weakened (minor)

- **Location:** REQ-51 `:618`. Base: "PR/fork state access is read-only". Worktree: "PR/fork jobs never publish durable **baseline** state; temporary local stores are permitted".
- **Problem:** The qualifier allows PR/fork jobs to publish non-baseline (shadow) data to `scout/state`. That conflicts with ADR-012's own trust model, where only protected register, finalize and triage jobs write. This change to an ADR-007/008 clause is not declared.
- **Required fix:** "PR/fork jobs have read-only access to `scout/state` and never publish to it; temporary local stores are permitted."

### ADR012-R4-12: Amended-area names are incomplete (minor; remainder of R3-09)

- **Location:** line 3 of ADR-000, ADR-003 and ADR-007; README `:18,21,25`; cards `:736,807,901,1007`.
- **Problem:**
  - ADR-007 owns the state export and manifest cutover (REQ-48 attribution), but its note omits the manifest-v2 export and merge change.
  - The ADR-000/003 notes omit the maintainer triage history added to REQ-22.
  - README `:18` now reads "partly superseded; amended by ADR-012 (…)", which implies ADR-012 is the only amender. The status line names ADR-001, ADR-002 and ADR-003.
- **Required fix:**
  - ADR-007: "(13→14 migration, export manifest v2, shadow publish/install)".
  - ADR-000/003: "(readiness observations/eligibility, maintainer triage history, shadow workflow/publish/install)".
  - Apply these to the status lines, README rows, card relations and ADR-012 card relation.
  - README `:18`: "partly superseded; amended by ADR-001, ADR-002, ADR-003, ADR-012 (…)".
  - ADR-008 needs no change: ADR-012's REQ-51 change touches the publish clause, not ADR-008's association-row clause.

### ADR012-R4-13: The design overview still counts PR runs (minor; remainder of R3-10)

- **Location:** `scout-design.html:449`, the Activation callout, unchanged from base. It says "at least 5 PR runs with a comparable base … at most 10% of all PR runs in the window (a run counts once…)".
- **Problem:** It contradicts REQ-22/78, and the overview is part of the governed design.
- **Required fix:** "at least 5 comparable-base PR observations (distinct PR, head tree and merge-base) … at most 10% of observations in the window (an observation counts once…)". Change nothing else.

### ADR012-R4-14: REQ-168's release list omits the shadow workflow (minor)

- **Location:** REQ-168 `:714`, which ships reusable workflows `scout-guardian.yml` and `scout-guardian-trusted.yml` only. The same REQ now requires shadow adoption and dogfooding through Scout's own shadow window (plan `:61`).
- **Required fix:** "reusable workflows (`scout-guardian.yml`, `scout-guardian-trusted.yml`, `scout-shadow.yml`)".

### ADR012-R4-15: Tracker inconsistencies (minor; remainder of R3-13)

- **Location:** plan `:31–35` (only `B4 --> B5`), plan `:15` ("B3–B5 remain open"), STATUS `:7` ("Latest amendment: … ADR-011").
- **Problem:**
  - The plan's dependency graph and its summary row disagree with the table at `:60–61`.
  - ADR-011's canonical STATUS named itself the latest amendment, with its effectiveness condition. ADR-012's signed bytes will be stale on adoption.
- **Required fix:**
  - Add `B4S[B4-S ADR-012 implementation] --> B5` to the graph.
  - Change plan `:15` to "B3–B5 and B4-S remain open".
  - In STATUS `:7`, name ADR-012 as the latest amendment, "effective on the explicit final zero-issue round of `docs/adrs/2026-11-08-shadow-readiness-inputs/`", and ADR-011 (adopted, PR #17) as the prior amendment.

### ADR012-R4-16: Ledger gaps (minor)

- **Location:** `research-ledger.md:75–76`; `ADR.md:83`; card `:1010`.
- **Problem:** The round-3 and round-4 decisions rest on sources that are not in the ledger:
  - GitHub `workflow_dispatch` ref semantics, and the contexts reference (`actor` against `triggering_actor` on re-runs);
  - the `git merge-tree --write-tree` documentation;
  - repository facts:
    - `scout-guardian.yml:55,58,102–105`: merge-commit checkout, diff against `base.sha`, no override flags;
    - `scout-guardian-trusted.yml:62`: default-branch guard;
    - `scout-guardian-trusted.yml:111–118`: restore before publish;
    - `would_block.go:79` and `cmd/scout/main.go:537–597`: no writer for event triage;
    - `policy.go:435–458`: `AcceptOverride` matches exact names;
    - `engine/gate.go:385–388`: event-ID inputs.
  - R-077's "informed" column omits REQ-26/162/163, and the row does not cite `round-3.md`.
- **Required fix:**
  - Add the three documentation links to R-070.
  - Add the repository facts, `round-3.md` and REQ-26/162/163 to R-077. Both rows are ADR-012's own unadopted rows.
  - Cite them in `ADR.md:83` and the card.

## `pull_request_target` safety recheck

- **Clean:**
  - No job runs PR content with a token. Register and finalize never check out PR trees.
  - Scan installs tools before fetching, fetches by validated SHA, removes credentials, and runs network-denied without caches or command files (`ADR.md:20–22,32`; CI checklist `:105–109`).
  - Job-level permissions replace caller grants, so `contents: write` does not reach scan. `contents: read` is enough for the preamble's read of `scout/state` ([workflow syntax](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax#jobsjob_idpermissions), challenger).
  - Identity comes from the intent and the compare API. The artifact is bounded, untrusted data.
  - Under `pull_request_target` the default-branch cache is read-only unless a write-capable `cache-mode` is declared, and ADR-012 forbids one.
- **Open:** dispatch-route binding (ADR012-R4-02) and sandbox write scope (ADR012-R4-08).

## Fidelity to REQ-22, REQ-27 and REQ-78

- **REQ-22:**
  - The maintainer-PR, during-window triage, "return that repo's check to shadow", "for that repo" and "zero events at window end … `provisional: true` on Track A alone" text is byte-identical to base, except ADR012-R4-01.
  - The changes from runs to observations, and from "10% of PR runs" to "10% of observations", are declared at `ADR.md:79`.
  - ADR-011's sentences are unchanged.
- **REQ-27:** the entire base text is a byte-identical prefix. The new sentences are scoped to shadow ("Shadow alone … guardian behavior is unchanged").
- **REQ-78:** the base sentences and ADR-011's sentences are byte-identical. The exits match `ADR.md:75`. Open: the "Historical" qualifier (ADR012-R4-04) and the triage cutoff (ADR012-R4-03).

## Checks that found no issue

- **Status-line convention.** ADR-012's `Status: accepted` before sign-off matches ADR-011's route (a) (ADR-011 `ADR.md:3`) and the README rule `:3`. `ADR.md:85` leaves the digests to the final round.
- **Status lines.** Line 3 of ADR-000, ADR-003, ADR-007 and ADR-008 is the only line compared. Each matches its README row and card relation (completeness: ADR012-R4-12).
- **REQ attribution.** Exactly 14 REQs carry ADR-012, matching `ADR.md:4` and the card (peer). REQ-46 is unchanged.
- **STATUS.** `:9` and `:129` are conditional. The ID ranges at `:131` are consistent with the ledger.
- **Ledger rows R-070 and R-077.** They hold sources and facts, and their caveats leave the design choices to the ADR. R-070's claims about GitHub match its secure-guidance page.
- **Code consistency.** The `AcceptOverride` semantics (exact names, no team resolution) match `policy.go:435–458`. The ShadowReady gates match `policy.go:403–425`.

## Considered and not raised

- **A later re-run erases an earlier unknown (peer N8).** Taking comparability from the latest representative was the council's own R3-06 rule. It matches the guardian gate, which also uses its latest run after a re-run, and re-runs need write access. Not a defect.
- **`ADR.md:85` final-round wording (peer N16).** "Final integrated canonical bytes require renewed review" is satisfied by any zero-issue round on the integrated bytes. The README rule and this folder's rounds decide finality. The writer may clarify this, but it does not block.
- **The `current_window` conditional wording at STATUS `:121`, plan `:89` and README `:30` (part of peer N12).** STATUS `:9`, plan `:91` and the README note `:14` already carry the effectiveness condition.
- **ADR-002 and ADR-009 status lines (REQ-39/168 attribution).** ADR-012's additions to those REQs are CLI and CI-adoption content that traces to ADR-000, ADR-003 and ADR-007, which are already amended.
- **Shadow `cancel-in-progress` (peer).** Cancellations are disclosed as `not_assessed` (the R2-05 rule). The template should use `cancel-in-progress: false`. That is an implementation note.
- **`.gitattributes` effects on `merge-tree`.** Covered by the [ASSUMPTION] and the parity proof.
- **The guardian PR lane passes no override inputs, against REQ-23.** This gap predates ADR-012 and lies outside it. It is recorded here for the trackers [INFERENCE: no other wiring was found in `.github/`].

## Peers and where they changed the outcome

- **`innovation-challenger`** raised 2 major and 1 minor finding. Adopted:
  - the dispatch default-branch gate and policy revision, merged into ADR012-R4-02, which is major. The lead had drafted this as minor.
  - the override-input mismatch, adopted as ADR012-R4-06 with severity **lowered** from major to minor. It affects only PRs that carry override text, and the fix is a wording change.
  - the remedy conflation (ADR012-R4-07). This was new to the lead and corrects the council's own R3-01 text.

  Its checks of permission inheritance and of `triggering_actor` semantics on re-runs were confirmed.
- **`experiment-peer`** raised 1 major and 15 minor findings. Adopted:
  - N1, the REQ-22 regression. The lead had drafted it as minor. The peer's major is upheld after the lead found that main's wording was the required ADR011-R2-03 fix (ADR012-R4-01).
  - N3, merged into ADR012-R4-02. Its stale-re-run reversion **raised** that issue to major.
  - N4, merged with the lead's draft as ADR012-R4-03. The lead's monotonic-stamp fix was chosen over the peer's per-commit evaluation, because it needs no state-commit argument on the CLI.
  - N5 and N14, merged as ADR012-R4-04.
  - N6 (ADR012-R4-05), N7 (ADR012-R4-09), N9 (ADR012-R4-10) and N10 (ADR012-R4-11). These were new to the lead.
  - N11 in part (ADR012-R4-12; ADR-008's REQ-51 area not required), N2 (ADR012-R4-13, also the lead's), N12 in part (ADR012-R4-15), N13 (ADR012-R4-14, also the lead's) and N15 (ADR012-R4-16, with the lead's additions).

  Not adopted: N8, N16, and the conditional-wording part of N12.
- **Lead-only:** ADR012-R4-08 (sandbox write scope) and the residual-risk statement in ADR012-R4-02.

## Council memo

1. **Recommendation.** Do not sign off. Keep the protected register → sandboxed merge-tree scan → trusted finalize design, which every round-3 fix strengthened. Apply the 16 fixes on the integrated bytes and submit round 5 with fresh Main digests for the same 13 signed files.

2. **Alternatives.**
   - **(A, recommended)** The current design plus the fixes. Dispatch triage reuses the trusted lane's default-branch guard, current-tip policy and idempotent run-ID rows.
   - **(B)** Record triage verdicts through maintainer PRs to a protected triage file, read by reconcile. Branch protection and CODEOWNERS then supply authorization, and git history supplies publication order. That is strong but adds high friction for each verdict and reverses the R3-02 choice.
   - **(C)** Wire maintainer overrides into both the guardian and shadow lanes (the alternative for ADR012-R4-06). That is more faithful to REQ-23 but changes the guardian, so it needs its own ADR.
   - **(D)** A scheduled snapshot collector (from round 3). It avoids `pull_request_target`, but it misses intermediate heads and changes the sampling unit.

3. **Evidence.**
   - REQ-22 regression: `scout-design.html:589` against base, and `round-2.md:147–148`.
   - Dispatch: GitHub's [dispatch](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#workflow_dispatch) and [contexts](https://docs.github.com/en/actions/reference/workflows-and-actions/contexts) documentation, and `scout-guardian-trusted.yml:62`.
   - Event IDs: `gate.go:385–388`.
   - Overrides: `scout-guardian.yml:102–105`, `diff.go:24–25`, `gate.go:322–337`.
   - REQ-51: base `:618`. Overview: `:449`. Release list: `:714`. Plan graph: `:31–35`.
   - Rebase hygiene: raw comparisons of the ledger at `:73–74`, the card at `:991–1004` and README `:29`.

4. **Risks.**
   - The sandbox and the dispatch guards are unimplemented delivery prerequisites.
   - GitHub's event-policy enforcement starts 2026-12-02.
   - `merge-tree` parity remains an [ASSUMPTION].
   - Write-access collaborators remain outside ADR-012's boundary, and the `scout/state` ruleset is the control against them.
   - Reversibility is high: the change is design-only, schema 13 still ships, and a superseding ADR can replace it.

5. **Execution outline.**
   1. Restore REQ-22's demotion sentence verbatim (R4-01). Restore REQ-51's read-only clause (R4-11).
   2. Specify the dispatch default-branch gate, current-tip policy, the `actor == triggering_actor` check, run-ID idempotency, the residual-risk statement and proofs (R4-02). Add the monotonic `recorded_at` and the `current_window` triage cutoff (R4-03).
   3. Add the interval lower bound and REQ-78 wording (R4-04), shadow event-ID inputs (R4-05), and inspection-status disclosure (R4-09).
   4. Freeze empty override inputs and extend the parity proof (R4-06). Split the schema remedies (R4-07). Specify the sandbox write scope and post-sandbox integrity (R4-08).
   5. Make the REQ-10 triage table and migration explicit (R4-10). Fix area names and the README row (R4-12), the overview callout (R4-13), the REQ-168 release list (R4-14), and the plan graph, plan `:15` and STATUS `:7` (R4-15).
   6. Extend R-070 and R-077 and cite them (R4-16).
   7. Main recomputes the 13 digests and the diff stat against main, then submits round 5.

## Verdict

**ADR-012: 16 issues open (2 major, 14 minor). Not signed off.** Open: ADR012-R4-01 to ADR012-R4-16. The digests in the table above identify the bytes that were reviewed and carry no sign-off.
