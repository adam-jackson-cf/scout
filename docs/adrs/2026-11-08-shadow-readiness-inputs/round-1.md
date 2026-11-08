# Round 1: ADR-011 (Gitleaks policy `version` contract) and ADR-012 (shadow-window readiness inputs)

Date: 2026-11-08. This round follows design rules 1–3 and the ledger rules in `CLAUDE.md`, and the ADR rules in `docs/adrs/README.md`. Both ADRs are agent decisions taken under the owner's standing delegation (R-064). Where a fix involves a choice, the council names the option it recommends, under that delegation.

The innovation council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review. It checked every code citation, the counting definitions against `ShadowStats`/`ShadowReady`, the REQ edits against unchanged REQs, and the Gitleaks string and equality claims.
- **`experiment-peer` (Opus):** review of competing interpretations and confounds. It looked at whether the readiness counts can be measured validly, at attempt identity, epochs and completeness, at migration and export implications, at ADR format and route, at ledger rows, and at how the delegation is recorded.

The review was read-only. Nobody edited files, made model calls outside the council, or touched `.codegraph`.

**The writer computed the digests below; the council did not.** The lead and both peers had read-only tools and no shell. This round is not a sign-off, so the digests only identify the bytes that were reviewed.

| File | SHA-256 (writer) |
|---|---|
| `docs/adrs/2026-11-08-gitleaks-policy-version/ADR.md` | `70a3d8504536636808c1ac5f4c68ba344e062aecb319f68fbc6e83ce5b715d65` |
| `docs/adrs/2026-11-08-shadow-readiness-inputs/ADR.md` | `f3419387f3b3a95a22a7e0d7f76f367a5df4dfe85062d1e52f8f3651a2288ec7` |
| `docs/adrs/README.md` | `bc5a193ee0b3187d40a25291597ed3458e6df8a90b17caefd56bae15b1d4d20c` |
| `docs/scout-design.html` | `3b849f8e0b18c70ab847c5cf318676e9852655f4fd97464db4408c56627596cf` |
| `docs/references/policy-file.md` | `c7efe224db2c7a3e2aa41adb28a741f5e4ed7d97f7bcc5d64289f6cfbed910b6` |
| `docs/STATUS.md` | `35857d061aacac0dbf5454060b00322e9cb8099a802bc5133bb8ccf67473fb0c` |
| `docs/plans/2026-11-06-remaining-work.md` | `b34746a4ea4c63d483302caedc18a485afe9ecddde44780a83a2593d3c5ffb1a` |
| `docs/research-ledger.md` | `1e3071c363138505a9ff406166300b7842d1808f42f522438fe0de2b6583a9f4` |

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11`, branch `agent/policy-shadow-adrs`.
- The changes are uncommitted, on top of base `main` `ef6f50e8`.
- Base text was compared using `/Users/adamjackson/Projects/work/scout-gap2` for the docs, and the main worktree for `policy-file.md`.

**Files read:**
- **ADRs:** both `ADR.md` files, in full.
- **`scout-design.html`:**
  - REQ-10 `:610`, REQ-22 `:589`, REQ-27 `:612`, REQ-46/47/48 `:613-615`, REQ-51 `:618` and REQ-78 `:600`, each compared with base;
  - REQ-39 `:704`, REQ-160/161 `:706-707` and REQ-166 `:712`;
  - the Appendix B cards for ADR-010, ADR-011 and ADR-012 at `:971-1009`.
- **Other docs:**
  - `docs/adrs/README.md`;
  - `STATUS.md:9,117-122,132`;
  - plan `:84-98`;
  - `policy-file.md:21,56-68`, compared with base;
  - R-064 and R-069 (`research-ledger.md:72-73`) and the ledger header;
  - ADR-007 `ADR.md:3,9-13`, ADR-008 `ADR.md:3,15,32` and ADR-003 `ADR.md:3,14-15,34`;
  - ADR-010 `ADR.md` and its rounds 1–3, as the route (a) precedent.
- **Code:**
  - `catalog_security.go:31,44,90,217,262,281` and `spec.go:156,196-228,267`;
  - `policy.go:59-123,286-300,362-426` and `runner.go:279,612`;
  - `engine/gate.go:50-75,289-303,356-385`;
  - `engine/diff.go:55-110,203-228,265-380` and `domain/gate.go:60-130`;
  - `would_block.go` (including `:14-45`) and `store.go:20-190,240-285`;
  - `export.go:20-160,200-300,676-720`;
  - `baseline/merge.go:19-143` and `engine/readiness.go:78-110`;
  - `cmd/scout/main.go:280-335`, `state_cmds.go:70-110` and `toolchain_test.go:63`.
- **Workflows:** `.github/workflows/scout-guardian*.yml`.
- **X1 worktree:** `scout-x1/evaluation/x1-triage-ranking/README.md:36,58,75-79`, for the R-064 to R-068 allocation.

## Issue summary

**18 issues are open:** 2 blocking, 11 major and 5 minor. ADR-011 has 2 issues of its own (1 major, 1 minor). ADR-012 has 16 (2 blocking, 10 major, 4 minor).

ADR012-R1-16 concerns a shared file. ADR-011's own Review section (`:38`) requires both sign-offs to name the same final shared-file bytes, so ADR-011 cannot be signed until the shared files are final.

| ID | Severity | Short title |
|---|---|---|
| ADR011-R1-01 | major | REQ-22 says exact rule-pack equality covers parser bumps; the code does not do this |
| ADR011-R1-02 | minor | "Owner review pending" is fixed in a frozen status line, with no way to record the review |
| ADR012-R1-01 | blocking | One nullable `would_block.pr_run_id` column cannot link events to attempts |
| ADR012-R1-02 | blocking | PR-controlled workflows can forge the artifacts the collector imports |
| ADR012-R1-03 | major | Shadow epochs depend on arrival order, and versions run interleaved |
| ADR012-R1-04 | major | Window timestamp, default `as_of` and `Days` are ambiguous; the default query is always `unmeasured` |
| ADR012-R1-05 | major | Abandoned attempts and unreconcilable gaps leave readiness `unmeasured` forever |
| ADR012-R1-06 | major | Per-attempt counting lets reruns inflate PRRuns and dilute the unknown ratio |
| ADR012-R1-07 | major | Per-check "inspected" is undefined, and the current wording biases counts toward zero events |
| ADR012-R1-08 | major | CI and local collection scopes cannot be told apart in facts or exports |
| ADR012-R1-09 | major | Keeping `scout.state/v1` lets older binaries silently drop run facts and links |
| ADR012-R1-10 | major | Fact lifecycle and publish-merge rules are incomplete |
| ADR012-R1-11 | major | ADR-007/ADR-008 amendments are not recorded, and the refuse-13 alternative is missing |
| ADR012-R1-12 | minor | REQ-166, REQ-47, REQ-39 and REQ-161 still contradict the new contract |
| ADR012-R1-13 | major | The collector interface is not specified well enough to implement |
| ADR012-R1-14 | minor | CLI and migration effects on existing query-only commands are misstated |
| ADR012-R1-15 | minor | "Owner review pending" is fixed in a frozen status line (ADR-012 counterpart of ADR011-R1-02) |
| ADR012-R1-16 | minor | `STATUS.md:9` uses present-tense pending wording inside the signed bytes |

## Issues: ADR-011

### ADR011-R1-01: REQ-22 says exact rule-pack equality covers parser bumps; the code does not do this (major)

- **Location:** `scout-design.html:589` (REQ-22); ADR-011 `ADR.md:15,21`.
- **Problem:**
  - REQ-22 now reads: "Parser/rule version bumps return that repo's check to shadow: activation uses exact rule/rule-pack version equality, not merely the tool version (ADR-011)". ADR-011 calls this "Already implemented" with "No behaviour … changes".
  - Every finding's `RuleVersion` is the hand-written `spec.Rules.Version` (`runner.go:612`).
  - Validation only requires that this string contain the bundled-file digest (`spec.go:224`). Nothing ties it to `Binary.Version` (`catalog_security.go:262`), and `toolchain_test.go:63` compares only `Binary.Version` with the pins.
  - So a Gitleaks binary bump that leaves the bundled rules unchanged, with nobody editing the `8.30.1` label at `:281`, keeps activations active.
  - Semgrep-backed `security.ci-mutable-action-ref` (`:217`, `scout-semgrep-actions/v1@sha256:…`) carries no binary version at all.
  - The parser-bump clause existed at base, but ADR-011 now names a mechanism that does not cover it and calls it implemented. ADR-012's "version change starts a new cohort" inherits the same gap.
- **Required fix (recommended option):**
  1. Keep the long string unchanged today, so there is still no migration.
  2. Restate REQ-22 precisely. A rule-content change resets through exact rule-pack equality, which is implemented. For Gitleaks, the pack version must also carry the pinned binary version.
  3. Add an implementation delivery condition, pending in a later PR alongside the diagnostic. Spec validation (with a test) requires the Gitleaks pack label's version segment to equal `Binary.Version`, so a binary bump forces a new activation version.
  4. Record the remaining gap for Semgrep-backed `security.ci-mutable-action-ref` as an open known gap in STATUS and the plan, to be decided in a later ADR, not claimed as covered.
  5. Correct ADR.md:15/21 ("Already implemented", "No behaviour … changes") to cover only rule-content resets. Mirror this in the ADR-011 card, `policy-file.md:67`, STATUS:122 and plan:87.

### ADR011-R1-02: "Owner review pending" is fixed in a frozen status line, with no way to record the review (minor)

- **Location:** ADR-011 `ADR.md:3`; ADR-011 card `rel` span (`scout-design.html`, ADR-011 article header); README index row.
- **Problem:**
  - Design rule 2 allows an accepted ADR's status line to change only when a later ADR supersedes or amends it, and ADR.md is in the signed set.
  - So an owner confirmation can never be written into `ADR.md:3` or the card, and "owner review pending" stays there permanently after the owner has reviewed.
  - Neither the ADR nor R-064 says where the owner's review will be recorded.
- **Required fix (recommended option):**
  - Status line: "accepted; decided by the agent under owner delegation (R-064)". Make the same change to the card `rel` span.
  - Keep the "owner review pending" state only in editable trackers: the README index status column, STATUS and the plan.
  - Add one sentence to Review: an owner confirmation is recorded as a new `user` row in the research ledger and in those trackers; an owner rejection is handled by a superseding ADR, which updates this status line.

## Issues: ADR-012

### ADR012-R1-01: One nullable `would_block.pr_run_id` column cannot link events to attempts (blocking)

- **Location:** ADR-012 `ADR.md:14,18,20,34`; REQ-27 (`:612`); REQ-48 (`:615`).
- **Problem:**
  - Event IDs are `wouldBlockID(base, head, check, issue)` (`engine/gate.go:357,382-384`), and `would_block.id` is the primary key (`would_block.go:15`).
  - Two genuine attempts on the same merge-base and head, which `:14` says get distinct IDs, produce the same event ID. One nullable column cannot link that event to both attempts.
  - The upsert also overwrites `rule_version` on conflict (`would_block.go:38-40`). A version change at the same revisions moves an event, with its old triage, into the new version's cohort. That contradicts "old cohorts and triage remain visible" (`:28`) and "Final facts are immutable" (`:18`).
  - The publish merge keeps one row per ID (`baseline/merge.go:85-101`).
  - `:20` asks for both "preserving event identity" and "multiple attempts may share [revision pairs]". With one column, both cannot hold.
  - Separately, `(repository_id, pr_run_id)` cannot be a composite foreign key, because `would_block` has no `repository_id` column and the store holds one repository (`store.go`).
- **Required fix (recommended option):**
  - Keep legacy event IDs.
  - Derive new event IDs from `(base, head, check, issue, rule_version)`, and never overwrite `rule_version`.
  - Keep triage on the event.
  - Replace the nullable column with a many-to-many attempt↔event relation. Export it as a sorted `eventIds` list inside each `pr-runs.jsonl` record, so `would_block.jsonl` rows keep their bytes and no `prRunId: null` normalization is needed.
  - Key facts by `pr_run_id`, with `repository_id` as a checked column.
  - Define `WouldBlock`, `Warranted`, `NotWarranted` and `Undecided` as distinct events linked to at least one cohort observation (ADR012-R1-06).
  - Legacy events are events with no links.
  - Add `engine/gate.go:357,382-384` and `would_block.go:38-40` to R-069 and to ADR Informed by.
  - If a per-row field is kept instead, the legacy-normalization exception (`:50`, REQ-51) must name the change to existing `would_block.jsonl` rows.

### ADR012-R1-02: PR-controlled workflows can forge the artifacts the collector imports (blocking)

- **Location:** ADR-012 `ADR.md:24`; REQ-51 (`:618`): "validates sanitized PR-run artifacts against provider repository/run/attempt/revision/digest provenance … Fork results use the same validation and no extra authority".
- **Problem:**
  - Provider metadata can confirm the repository, run, attempt and SHAs. It cannot confirm the payload: comparison, `base_gap_unknown`, inspection status, events, and the self-reported tool, rule and policy digests.
  - `pull_request` runs use the workflow definition from the PR's merge ref [external: GitHub Actions, "Events that trigger workflows", `pull_request`; to be added to the ledger]. A PR, including one from a fork, can edit the caller workflow or `scout-version` and upload fabricated "baseline, inspected, zero-event" facts. That inflates PRRuns and deflates unknowns, so fork results do carry authority over the denominator.
  - Renaming the job also hides attempts from provider enumeration.
- **Required fix (recommended option):**
  - The trusted collector accepts a payload only when all of the following hold:
    - from git, with no artifact input: the PR changes nothing under `.github/workflows/**` or `.github/actions/**` between merge-base and head;
    - from provider metadata: the run's workflow path and job match the PR workflow configured on the default branch;
    - the payload's Scout version and tool digest equal the trusted pin at merge-base (REQ-163 already SHA-256-verifies the installed binary).
  - Every other observed attempt is finalized `comparison: none` (a measured unknown), with an `untrusted_producer` reason. This fails safe and is terminal.
  - Change REQ-51 wording from "digest provenance" to "payload checked against trusted pins".
  - Add a research-ledger row for the GitHub `pull_request` workflow-source documentation.
  - Recomputing facts inside the trusted job is rejected: it would scan untrusted content in a privileged job.

### ADR012-R1-03: Shadow epochs depend on arrival order, and versions run interleaved (major)

- **Location:** ADR-012 `ADR.md:28,34,36,48`; REQ-78 (`:600`), "Version changes start fresh epochs".
- **Problem:**
  - "A return to an earlier version starts a new shadow epoch" and "Persist the epoch/start boundary … in run facts" (`:28`). The ADR does not say who assigns the epoch or which ordering defines a transition.
  - PR runs use the merge-ref workflow, so a PR that has not rebased onto a `scout-version` bump runs the old binary while others run the new one. Interleaved versions would reset the window over and over.
  - Read-only PR jobs cannot know the epoch. Concurrent collectors that union facts would assign it in nondeterministic order, while facts are immutable (`:18,50`).
- **Required fix (recommended option):**
  - Key each cohort by the exact `(check, rule_version)` and drop the epoch concept: the rollback rule, epoch storage and epoch export.
  - With ADR011-R1-01's fix, an exact version identifies rule content and parser, so evidence for the same version stays valid. Interleaved attempts of other versions do not reset it, and other versions' cohorts stay visible.
  - Cohort start = max(first terminal attempt of that version, end of the latest unreconcilable collection gap; see ADR012-R1-05).
  - Update REQ-78, `:28,34,36,48,56,73`, the card and the plan.
  - The challenger's and experiment-peer's alternative was to derive epochs from the version in effect on the default branch at merge-base. It also works, but needs version-history plumbing and drops PRs that bump the version. The council prefers exact-version cohorts as simpler and order-free.

### ADR012-R1-04: Window timestamp, default `as_of` and `Days` are ambiguous; the default query is always `unmeasured` (major)

- **Location:** ADR-012 `ADR.md:16,28,33,36,54`.
- **Problem:**
  - Records carry started, finished and recorded timestamps (`:16`), but neither window membership "through `as_of` (inclusive)" (`:28`) nor the `Days` anchor (`:33`) names which one.
  - The default `as_of` is "UTC now" (`:54`). In-progress PR attempts are always pending, and coverage always ends before now, so by `:36` the default query returns `unmeasured` (exit 2) almost every time.
  - A future `as_of` is not rejected.
  - "Whole UTC 24-hour periods" can be read as elapsed durations or as UTC calendar days. For example, 23:00Z to 01:00Z thirteen days later gives 13 under one reading and 14 under the other.
- **Required fix (recommended option):**
  - Membership and the `Days` anchor use the attempt's `started_at`: the provider-observed start, or the local mint time. Pending attempts have a start time, so completeness can be defined from it.
  - Completeness at T means every attempt with `started_at` ≤ T is terminal and reconciled.
  - The default `as_of` is the selected scope's `collectedThrough`, and it is printed.
  - An `as_of` later than now exits 3 (invalid argument). An `as_of` later than `collectedThrough` exits 2 (unmeasured).
  - `Days = floor((as_of − first_inspected.started_at) / 24h)`.
  - Finished and recorded times stay as provenance only.

### ADR012-R1-05: Abandoned attempts and unreconcilable gaps leave readiness `unmeasured` forever (major)

- **Location:** ADR-012 `ADR.md:18,24,36,48`; REQ-51.
- **Problem:**
  - "Unverifiable/missing artifacts leave the attempt unmeasured" (`:24`), and pending attempts make the cohort unmeasured "until reconciled" (`:18,36`). No terminal state is defined.
  - Concurrency-cancelled runs, attempts cancelled before the assessment step, and expired artifacts therefore poison the window permanently.
  - A collection gap the collector cannot reconcile does the same, because a fresh epoch starts only "after upgrade/legacy import" (`:36`).
  - A single `collectedFrom`/`collectedThrough` interval (`:48`) cannot represent coverage with gaps in it.
- **Required fix (recommended option):** define terminal outcomes by provider conclusion.
  - An attempt cancelled or skipped before the assessment step is `not_assessed`. It is excluded from TotalRuns, counted and disclosed.
  - An attempt that started but has no valid, verifiable artifact by the reconciliation deadline (artifact retention) is finalized `comparison: none`.
  - An unreconcilable gap opens a new collection boundary, and cohorts start after it.
  - The manifest's `collectedFrom` is the start of the latest contiguous, fully reconciled interval ending at `collectedThrough`, so one interval is enough.

### ADR012-R1-06: Per-attempt counting lets reruns inflate PRRuns and dilute the unknown ratio (major)

- **Location:** ADR-012 `ADR.md:14,30-32`; REQ-22 (`:589`), "at least 5 comparable-base PR runs".
- **Problem:**
  - Every count is per attempt, and "a genuine rerun gets a new ID" (`:14`).
  - Rerunning one PR five times satisfies `min_pr_runs` with no new code exposure. Rerunning clean PRs pushes the ≤10% unknown ratio down.
  - The ADR picks this reading of REQ-22's "PR runs" without saying so.
- **Required fix (recommended option):**
  - Count TotalRuns, UnknownAttribution and PRRuns over distinct `(PR key, head, merge-base)` observations. Each is represented by its latest terminal attempt, by `started_at`, with ties broken by attempt ID.
  - Keep attempts for completeness and audit only.
  - State this counting unit in REQ-22/78 and `:30-33`. This is not the rejected "derive runs from `would_block`" option: observations come from facts, so zero-event runs remain.

### ADR012-R1-07: Per-check "inspected" is undefined, and the current wording biases counts toward zero events (major)

- **Location:** ADR-012 `ADR.md:16,32,33`; REQ-78 (`:600`), "Baseline and recomputed runs with the check inspected count as comparable".
- **Problem:**
  - PRRuns and the `Days` anchor depend on per-check inspection status, but the ADR never says how it follows from lane health and coverage. Open cases: partial inspection (input limits in `catalog_security.go`), no eligible changed files (for example a docs-only PR), and whether the base side must have inspected the check's family.
  - If base-family gaps count only when they produce an unknown finding, runs with a head finding drop out of PRRuns while runs without one stay in. That biases the cohort toward zero events, which favours the rare-security provisional route.
  - REQ-78's "count as comparable" also redefines ADR-003's comparability instead of stating a PRRuns condition.
- **Required fix (recommended option):** define each check's status per observation.
  - `inspected`: the head adapter for the check completed over every eligible changed unit (no blocked, failed or limited coverage), **and** the base side inspected the check's family (no base gap for it).
  - `not_applicable`: no eligible changed units.
  - Otherwise `not_inspected`.
  - Only `inspected` counts toward PRRuns and anchors `Days`. All three statuses count in TotalRuns.
  - Reword REQ-78 to "count toward PRRuns", which leaves ADR-003's comparable-base rule unchanged.

### ADR012-R1-08: CI and local collection scopes cannot be told apart in facts or exports (major)

- **Location:** ADR-012 `ADR.md:16,24,36,48,56`.
- **Problem:**
  - Reports must show "collection scope" (`:56`), and local reports "cannot claim a complete CI window" (`:36`). But neither the record fields (`:16`) nor the manifest object (`:48`) carry a scope.
  - A store restored from CI that then runs local `scout diff` mixes the two scopes. Local facts can reach `scout/state` through the self-asserted `state publish --trusted` (`state_cmds.go:86,103-105`), bypassing the collector's checks.
  - `:24` says "PR/fork jobs never write SQLite baseline state". But CI PR jobs already write would-block rows to their temporary, writable store (`diff.go:223-228`), so it is unclear which path a CI PR job uses.
- **Required fix (recommended option):**
  - Add `scope: ci|local`, plus provider identity, to facts and coverage.
  - Make `shadowReadiness` CI-scope only.
  - Exclude local-scope facts from `Export` and publish.
  - Local reports show descriptive counts with status `unmeasured` for the CI window.
  - Restate `:24` as "never publish to `scout/state`": a CI PR job may write its temporary store, but the artifact is its only durable channel.

### ADR012-R1-09: Keeping `scout.state/v1` lets older binaries silently drop run facts and links (major)

- **Location:** ADR-012 `ADR.md:48`; REQ-48 (`:615`).
- **Problem:**
  - `verifyManifest` checks only `schema` and the listed files (`export.go:678-720`), and decoding is lenient.
  - An older binary therefore imports a v14 export without error, ignoring `pr-runs.jsonl`, links and `shadowReadiness`.
  - On publish, `MergeState` copies unknown files local-wins and rebuilds the manifest from the local metadata only (`merge.go:28-33,128-136`). So a downgraded publisher silently drops `shadowReadiness`, and its would-block rows replace linked rows.
- **Required fix (recommended option):**
  - Bump the manifest schema to `scout.state/v2`, so older binaries refuse it through the existing check at `export.go:693`.
  - Keep the `state/v1/` directory as a layout name and say so explicitly.
  - Treat `scout.state/v1` imports as the legacy-normalization path.
  - Update REQ-48/51 and `:48,50`.

### ADR012-R1-10: Fact lifecycle and publish-merge rules are incomplete (major)

- **Location:** ADR-012 `ADR.md:18,50`.
- **Problem:**
  - "Conflicting payloads are errors" does not separate a legitimate pending→final transition from a real conflict.
  - The immutable payload includes a "recorded" timestamp, so two collectors importing the same attempt at different times conflict and block publishing.
  - The required union is not named in `baseline/merge.go`, where every new file is currently local-wins.
- **Required fix:**
  - Define pending < final. A final record replaces a pending one, and two finals must be identical apart from collector provenance.
  - Keep recorded and collected times out of the immutable identity (deterministic: earliest wins).
  - Name the `MergeState` changes: union facts by ID with the pending<final rule, union the links, merge coverage per scope (a gap is never marked collected), and recompute `shadowReadiness` rather than copying it from the local side.

### ADR012-R1-11: ADR-007/ADR-008 amendments are not recorded, and the refuse-13 alternative is missing (major)

- **Location:** ADR-007 `ADR.md:3,10`; ADR-008 `ADR.md:3,15`; `docs/adrs/README.md` ADR-007/008 rows; ADR-007/008/012 card `rel` spans; ADR-012 `ADR.md:58-62,71`.
- **Problem:**
  - ADR-007 decided "Keep only the v1 re-anchor migration", and ADR-008 set the schema to 13, refusing the old version rather than migrating it. ADR-012 adds a 13→14 migration and says it "amends REQ-10's former v3-only migration contract" (`:44`).
  - Design rule 2 requires the adopting commit to update the amended ADRs' status lines. The ADR-008 precedent made ADR-007 "partly superseded by ADR-008".
  - Neither status line, nor the index, nor the cards name ADR-012. The ADR-012 card `rel` says only "amends state/readiness requirements".
  - The options list omits the precedent alternative: refuse version 13 and re-baseline.
- **Required fix:**
  - ADR-007 status: "partly superseded by ADR-008; amended by ADR-012 (13→14 migration)".
  - ADR-008 status: "accepted (provisional); amended by ADR-012 (schema 14)".
  - Update the README rows and the card `rel` spans to match.
  - Add both `ADR.md` files to the signed set and list them in ADR-012's Changes.
  - Add the refuse-13 option and reject it, for example because it would lose unpublished local triage and would-block history.

### ADR012-R1-12: REQ-166, REQ-47, REQ-39 and REQ-161 still contradict the new contract (minor)

- **Location:** `scout-design.html:712` (REQ-166), `:614` (REQ-47), `:704` (REQ-39), `:707` (REQ-161).
- **Problem:**
  - REQ-166 still says to remove "the migration ladder/backfill in favor of current-schema plus v1 re-anchor".
  - REQ-47's shape parity covers only new and v1-reanchored stores, while REQ-10 adds v13-upgraded stores.
  - REQ-39's CLI inventory lists only `readiness --matrix`.
  - REQ-161 names the trusted job's token use as `--fetch-pr-metadata` only. The collector needs `actions: read` to list runs and download artifacts.
- **Required fix:**
  - Qualify REQ-166: "plus v1 re-anchor and the targeted 13→14 migration (ADR-012)".
  - Add v13-upgraded stores to REQ-47.
  - Add `readiness --shadow` to REQ-39.
  - Add the collector's `actions: read` permission and trigger to REQ-161.
  - Add ADR-012 to each row's ADR column.

### ADR012-R1-13: The collector interface is not specified well enough to implement (major)

- **Location:** ADR-012 `ADR.md:24,73`; REQ-51.
- **Problem:** The ADR calls itself a "concrete future implementation contract" (`:81`), but the following are missing:
  - the command or mode that collects;
  - the artifact name, schema ID and version, and the "bounded" size and record limits;
  - how provider attempts are enumerated, including the deadline for reconciling them;
  - which trusted job runs the collector, its trigger and its permissions.

  An implementation would need further design decisions, and so another ADR.
- **Required fix (recommended defaults):**
  - A `scout state collect --remote <name>` mode in the trusted default-branch job, run on push and schedule.
  - Artifact `scout-pr-run` carrying `scout.pr-run/v1` JSON, with stated size and event caps.
  - Enumerate `pull_request` runs of the default-branch-configured PR workflow path through the Actions API with `actions: read`.
  - The reconciliation deadline equals artifact retention (ADR012-R1-05).
  - Fold these into REQ-51/161.

### ADR012-R1-14: CLI and migration effects on existing query-only commands are misstated (minor)

- **Location:** ADR-012 `ADR.md:42,54,56`.
- **Problem:**
  - Once the schema moves to 14, every `OpenExistingReadOnly` caller fails on a local v13 store until a writable command upgrades it. That covers `engine/readiness.go:83`, which reports "existing Scout state is unavailable", and `signals.go:29`, `state_cmds.go:26,109` and `main.go:632`. So "Existing readiness … behaviour stay unchanged" (`:56`) is inaccurate.
  - It is also unclear whether a v13 store under `--shadow` gives exit 2 or exit 3.
  - Existing `readiness` exits 3 when not ready (`main.go:331-332`), while `--shadow` exits 0 when computed but not ready. The mode-specific exit table appears only in the ADR, not in REQ-78.
- **Required fix:**
  - State that query-only commands on a v13 store report "writable upgrade required" (exit 3), naming the remedy: run any writable command, or `scout state restore`.
  - Qualify "unchanged" as referring to flags and output.
  - A missing store under `--shadow` exits 2; a v13 store exits 3.
  - Put the `--shadow` exit table in REQ-78.

### ADR012-R1-15: "Owner review pending" is fixed in a frozen status line (minor)

- **Location:** ADR-012 `ADR.md:3,66`; ADR-012 card `rel` and the "Informed by" item "owner review pending"; README index row.
- **Problem:** Same as ADR011-R1-02. The owner's review can never be recorded in the signed status line or card.
- **Required fix:** Same as ADR011-R1-02.

### ADR012-R1-16: `STATUS.md:9` uses present-tense pending wording inside the signed bytes (minor; shared, blocks both ADRs)

- **Location:** `docs/STATUS.md:9`.
- **Problem:**
  - STATUS:9 says "ADR-011/012 are adopted-form review candidates … not adopted until each folder has an explicit zero-issue final council round".
  - The adopted-form ADR.md files, index and cards say "accepted". STATUS is in the signed set and cannot be changed to mark adoption, so this line would describe the ADRs as candidates permanently.
  - The ADR-010 precedent rejected exactly this (ADR010-R1-08 and R2-02: "the accepted record would show an unmet gate permanently").
  - The conditional wording at STATUS:122, plan:90/98 and the README rule line stays true after adoption and is not raised.
- **Required fix:** Use conditional adopted form, for example: "ADR-011/012 (agent-decided under R-064) take effect on each folder's final zero-issue council round; adoption is recorded outside the signed set."

## Checks that found no issue

- **Code citations:** all of them match. In ADR-011: `catalog_security.go:262,281`, `spec.go:199-226`, `policy.go:84,293,367-375` and `PROPOSAL.md:43` ("Gitleaks version strings): implementation notes, not design statements"). In ADR-012: `policy.go:384-424`, `diff.go:71-73,275,341-371`, `domain/gate.go:70-77`, `store.go:25,71-84,131-185,248-278`, `export.go:23-48,97-156,218-292` and `main.go:288-335`.
- **Gitleaks string:** the full string in ADR-011, REQ-78 and `policy-file.md:64` equals `catalog_security.go:281`.
- **Activation logic:** `CheckStatus` uses exact `a.Version != ruleVersion` (`policy.go:375`), after the eligibility and guardian checks. `generic-api-key` stays excluded (`policy.go`, `ValidateEligibleID`). `CheckIDOf` maps Gitleaks findings to `security.secrets.<rule>` (`engine/gate.go:61-62`).
- **Readiness arithmetic:**
  - Defaults are 14 days and 5 runs (`policy.go:68-69`).
  - The unknown bound rejects only `>` 10% (`policy.go:409`), so "equality passes" is correct.
  - The rare-security branch runs after the unknown and day checks and before the PR minimum (`:415-420`), so ADR-012's "still requires … the unknown bound and elapsed days" matches, and the exception is not broadened.
- **Missing producer:** `ShadowStats` has no producer outside tests. `comparison` is reported (`diff.go:275`) and not persisted (no `internal/state` match). `would_block` has no run link. R-069's static claims hold.
- **Unchanged REQs:** REQ-46's refusal safeguards and REQ-49/50's baseline-health semantics are preserved. A v13 read-only refusal follows the existing v3 pattern (`store.go:171-173`).
- **Route and format:** route (a) wording in both Review sections matches the ADR-010 precedent. Index rows, cards and REQ ADR-ID columns exist and are consistent. STATUS and the plan say implementation is pending and schema 13 is what ships.
- **Delegation record:** R-064 quotes the direction identically in R-064, ADR-011:26 and ADR-012:66. It says this is not owner review and does not waive council review. Whether the quote is verbatim rests on the writer's attestation: no transcript was readable.
- **Gitleaks pinning in REQs:** pinning the pack string in REQ-78 follows the existing binary pin in REQ-63. A pack bump already needs a design change.

## Considered and not raised

- **R-064 collides with the X1 worktree (peer: blocking).**
  - `scout-x1/evaluation/x1-triage-ranking/README.md:58,75` drafts its own R-064 ("Owner delegation of recommended drafting decisions") and allocates R-065 to R-068. That explains why this unit jumps to R-069.
  - Nothing in this unit's bytes is wrong: R-064 here is the fuller, verbatim row for the same direction, and the gap is reserved, not accidental.
  - **Condition, not an issue:** before either branch lands, Main keeps this R-064 as the only R-064 row. X1 drops its draft row and cites R-064, and X1's R-065 to R-068 land before or alongside R-069. If that ordering fails, this unit renumbers R-069 to the next free ID, which needs re-review.
- **ADR-011's self-citation of `policy-file.md:21` (peer).** Base line 21 already reads "The rule or rule-pack version that was activated", so the claim is true of both versions. Nit.
- **Per-rule digest versioning as a missing ADR-011 option (peer).** ADR-011:15 states the consequence ("Every eligible rule in this pack uses the same pack version"). Per-rule versioning is a separate redesign and is not needed to resolve known gap 3. Nit.
- **ADR-003 amendment for the inspection condition.** Resolved inside ADR012-R1-07: rewording to "count toward PRRuns" leaves ADR-003's comparability unchanged, so no ADR-003 relation is needed.

## Peers and where they changed the outcome

- **`innovation-challenger`** raised 5 issues against ADR-012 and none against ADR-012's sibling, ADR-011:
  - event identity (became R1-01);
  - epoch ordering (R1-03);
  - collection scope (R1-08);
  - timestamp ambiguity (R1-04);
  - ADR-007/008 status lines (R1-11).

  It found ADR-011 defect-free and the route wording acceptable.
- **`experiment-peer`** raised 23 items. Adopted:
  - parser bump (became ADR011-R1-01) and owner-review recording (ADR011-R1-02 and ADR012-R1-15);
  - collector trust (R1-02);
  - default `as_of` and `Days` (R1-04);
  - terminal outcomes (R1-05);
  - rerun inflation (R1-06);
  - inspection (R1-07);
  - local scope and `--trusted` (R1-08);
  - schema-v1 silent drop (R1-09);
  - lifecycle and merge (R1-10);
  - REQ-166/47/39/161 (R1-12);
  - collector interface (R1-13);
  - exit codes (R1-14).

  Merged into other issues: the foreign-key shape and the `prRunId` null rows (into R1-01), and the PR-job SQLite ambiguity (into R1-08). The R-064 collision became a condition, and two nits were declined.
- **Lead:** independently found event identity and the `rule_version` upsert (R1-01), producer integrity (R1-02), the selection bias in inspection (R1-07), the existing query-only command effects (R1-14) and `STATUS:9` (R1-16).
- **Changes to the lead's position:**
  - The experiment-peer's ADR011-R1-01 moved ADR-011 from "no own issues" to one major issue. The challenger had found ADR-011 clean; the council resolved this on the code (`spec.go:224`, `toolchain_test.go:63`, `catalog_security.go:217`).
  - The challenger's epoch finding, together with the peer's interleaved-version example, led the council to recommend dropping epochs in favour of exact-version cohorts instead of either peer's epoch-assignment fix.
  - The peer's terminal-outcome and default-`as_of` findings showed the lead's draft fixes would still leave readiness permanently `unmeasured`. R1-04 and R1-05 now close that path.
  - Severity: the peer's blocking rating of ADR011-R1-01 was reduced to major, because today's label equals the binary pin and there is no runtime defect yet. The peer's blocking R-064 item became a condition.

## Verdict

**18 issues are open (2 blocking, 11 major, 5 minor). Not signed off.**

- ADR-011 has 2 issues of its own (ADR011-R1-01 major, ADR011-R1-02 minor).
- ADR-012 has 16 issues.
- Because of ADR-011's own shared-bytes condition (`ADR.md:38`), neither ADR can be signed until its own issues are closed and the shared files (`scout-design.html`, `docs/adrs/README.md`, `STATUS.md`, the plan, the research ledger and `policy-file.md`) are final.
- The digests in the table above identify the reviewed bytes only. No sign-off is given for any of them.

Submission conditions for round 2:
1. A `round-1-response.md` in each folder, with one row per issue: the fix made, or the recommended option chosen.
2. Fresh writer-computed SHA-256 digests for every signed file, including ADR-007 and ADR-008 `ADR.md` (R1-11).
3. A diff stat against `ef6f50e8` showing that every changed file is either in the signed set or a review record.
4. The R-064 coordination condition above, resolved or scheduled.
5. If ADR-011 is split from ADR-012 (see the memo), a statement of which shared-file bytes each round signs.

---

## Council memo

1. **Recommendation.** Sign off neither ADR this round. Fix ADR-011's two issues and decouple it: move ADR-012's shared-file edits into a follow-up revision, so that ADR-011 can reach a zero-issue round 2 on its own shared bytes. Revise ADR-012 using the recommended options above:
   - an attempt↔event relation with version-scoped new event IDs;
   - workflow-integrity-gated collection, with untrusted attempts counted as `comparison: none`;
   - exact-version cohorts with no epochs;
   - `started_at` windows with default `as_of = collectedThrough`;
   - terminal provider outcomes;
   - counting over distinct revision observations;
   - an explicit inspection predicate;
   - CI-only scope;
   - a `scout.state/v2` manifest schema;
   - pending<final merge rules;
   - ADR-007/008 status amendments;
   - the REQ alignment;
   - a specified collector interface.

   Then resubmit both.
2. **Alternatives.**
   - **(A) Decouple, then fix (recommended).** ADR-011 needs small text-only changes and can be adopted quickly. ADR-012 needs substantive redesign, which would otherwise hold back ADR-011's shared bytes.
   - **(B) Fix both together, as one unit.** Simpler bookkeeping, but ADR-011 waits for ADR-012's rounds, and any later ADR-012 edit renews ADR-011's review.
   - **(C) Narrow ADR-012** to local, read-only run facts and the producer only, with no CI collector. Fewer moving parts, but it cannot meet REQ-22 readiness for CI windows, the actual gap, and would need another ADR soon.
   - **(D) Collector recomputes facts in the trusted job.** This removes payload trust, but scans untrusted PR content inside a privileged job. Rejected under REQ-51/161.
   - **(E) For ADR-011, drop "parser bumps" from REQ-22** instead of adding the binary-segment invariant. This would contradict ADR-000's intent. Rejected.
3. **Evidence.**
   - **Identity and versions:** `engine/gate.go:357,382-384`; `would_block.go:14-45`; `baseline/merge.go:28-33,85-101,128-136`; `export.go:678-720`.
   - **Scanner versions:** `runner.go:279,612`; `spec.go:199-226,224`; `toolchain_test.go:63`; `catalog_security.go:31,44,217,262,281`.
   - **Readiness logic:** `policy.go:68-69,367-424`.
   - **Collection and state:** `diff.go:71-73,203-228,275,341-371`; `store.go:25,71-86,131-185,248-279`; `engine/readiness.go:83-89`; `state_cmds.go:86,103-105`; `main.go:288-335`.
   - **Design text:** REQ-10/22/27/39/47/48/51/78/161/166 (`scout-design.html:589,600,610,612,614,615,618,704,707,712`).
   - **ADRs:** ADR-007 `ADR.md:3,10`; ADR-008 `ADR.md:3,15`; ADR-010 round-1/2/3 (route (a) precedent).
   - **Ledgers:** `research-ledger.md:72-73`; `scout-x1/.../README.md:58,75-79`.
   - **External:** GitHub Actions `pull_request` event semantics (workflow taken from the merge ref).
4. **Risks.**
   - **Trust:** even with workflow gating, a compromised release or caller pin on the default branch defeats the collector. REQ-163's SHA-256 install and branch protection remain the base of trust.
   - **Metric choices:** counting distinct observations and the base-family inspection requirement make readiness slower to reach. That is intended, but it may lengthen windows for low-traffic repositories.
   - **Fail-safe outcomes:** counting untrusted and missing-artifact attempts as `comparison: none` lets a hostile PR author raise the unknown ratio. This fails safe (delays activation), never toward activation.
   - **Schema bump:** `scout.state/v2` makes downgraded binaries refuse restore (exit 3). That is preferred to silent data loss, but needs release notes.
   - **Ledger:** the R-064 collision is a coordination risk if X1 lands first.
   - **Reversibility:** everything is documentation until implementation, and design-only ADRs can be superseded. After implementation, the v14 schema and export changes become forward-only state formats.
5. **Execution outline.**
   1. Save this record as `docs/adrs/2026-11-08-shadow-readiness-inputs/round-1.md` and copy it to `docs/adrs/2026-11-08-gitleaks-policy-version/round-1.md`.
   2. ADR-011:
      - apply ADR011-R1-01: REQ-22 wording, the binary-segment invariant as a pending delivery condition, the Semgrep gap recorded in STATUS and the plan, and corrected "Already implemented" claims;
      - apply ADR011-R1-02;
      - apply R1-16's STATUS wording;
      - decide on decoupling, and if decoupling, remove ADR-012 edits from the shared files in ADR-011's submission;
      - write `round-1-response.md`.
   3. ADR-012: apply R1-01 to R1-15 with the recommended options. Add the GitHub `pull_request` source row to the research ledger. Add ADR-007/008 status lines, README rows and card `rel` spans to the signed set. Write `round-1-response.md`.
   4. The writer recomputes SHA-256 digests for every signed file and produces the diff stat against `ef6f50e8`. Each checks as follows:
      - every changed file is either signed or a review record;
      - ADR-011's and ADR-012's shared-file digests are identical, or decoupling is stated explicitly.
   5. Resolve or schedule the R-064 coordination with X1.
   6. Submit round 2. Each folder is adopted only on its own explicit zero-issue final round, with adoption recorded outside the signed set.
