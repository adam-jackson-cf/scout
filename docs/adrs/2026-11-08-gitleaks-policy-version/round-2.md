# Round 2: ADR-011 (Gitleaks policy `version` contract) and ADR-012 (shadow-window readiness inputs)

Date: 2026-11-08. This round follows design rules 1–3 and the ledger rules in `CLAUDE.md`, and the ADR rules in `docs/adrs/README.md`. Both ADRs are agent decisions made under the owner's standing delegation (R-064). Where a fix involves a choice, the council names the option it recommends under that delegation.

The council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review of the round-1 closures, the trust model of the `ADOPTION.patch` split, the historical status lines, and the arithmetic and exit codes.
- **`experiment-peer` (Opus):** review of competing interpretations and confounds. It covered counting validity, attempt identity, migration, delegation, the ledger, adoption mechanics and the status lines.

The review was read-only. Nobody edited files, ran model calls outside the council, executed product code, or touched `.codegraph`.

**Main computed and verified the digests below with `shasum`. The council did not, because the lead and both peers had read-only tools and no shell.** This round is not a sign-off, so the digests only identify the bytes that were reviewed.

| File | SHA-256 (Main) |
|---|---|
| `docs/adrs/2026-11-08-gitleaks-policy-version/ADR.md` | `03de1928c789e2c6785d0313ce0dbd383ddf4d325213c9c9e8f725b881f5eeb1` |
| `docs/adrs/2026-11-08-gitleaks-policy-version/ADOPTION.patch` | `b1cd1aefd39fbe25cadeb351a83269e2fab17b3bca381ef30fa16f53dcea0c1d` |
| `docs/adrs/2026-11-08-gitleaks-policy-version/ADOPTION-SHA256.txt` | `9863d4a22fd0a90b7cfa020b467fc913152fdce96dc6874b102690fdc1951aa6` |
| `round-1-response.md` (identical in both folders) | `b014fca2177bf0b23266a046156d11b08e3bce7a61a862fd811ff4185cfc6ecd` |
| `docs/adrs/2026-11-08-shadow-readiness-inputs/ADR.md` | `ee414adfb52c2ef461275eeae0d24c35cbaa6c0e79aaf25dfca14721632b2980` |
| `docs/scout-design.html` | `054bd41b7c551842c4ad5bb173f2112c2a91b5f2c27e886c3dc0f5d21be415bb` |
| `docs/adrs/README.md` | `f024154a3dc495f759d02857ed7fc1742a07a464a5b639d1adf30dd907d383ef` |
| `docs/STATUS.md` | `f8149f017345e5754d6bccef91fd1cd0a5188129e75280a7eb39dcf97c66b193` |
| `docs/plans/2026-11-06-remaining-work.md` | `5f9136161136cc065b8c912475a2f2f24fc03db8ebe7386d70329c08b1766b7b` |
| `docs/references/policy-file.md` | `5c2f49ce4682fe3bd4a670ab3f605603eb10d91f70b6eb4657cbb63775d48311` |
| `docs/research-ledger.md` | `71edb4006d87dc56766e61e56071e3a5fc5f834b8e59c4f960d4a1fcc820d08c` |
| `docs/adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) | `ba819ac228dbe38465a313f186baad33b14a63ae177ebbc2c33da8ffcb621e6e` |
| `docs/adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) | `9224f15862f360a88cc9936469b8ad8ac38023c0fd96b343a0311a87e5c50278` |

**Result hashes of `ADOPTION.patch` not verified.** Main ran `git apply --check` on `/tmp/adr11-patchcheck` (`ef6f50e8`), and it passed. Nobody has yet shown that the six files the patch produces match `ADOPTION-SHA256.txt`. Only `policy-file.md` is cross-checked: its listed result hash `5c2f49ce…` equals Main's digest of the canonical file, which is consistent with ADR-012 leaving that file unchanged.

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11`, branch `agent/policy-shadow-adrs`, with uncommitted changes on top of `ef6f50e8`.
- Base bytes from `/tmp/adr11-patchcheck`.
- ADR-011 is reviewed as `ADR.md` plus the hunks in `ADOPTION.patch`. ADR-012 is reviewed as the canonical working files.

**Files read:**
- Both `ADR.md` files, `ADOPTION.patch`, `ADOPTION-SHA256.txt`, `round-1.md` and `round-1-response.md`.
- Canonical files:
  - `scout-design.html` REQ-10/22/27/39/47/48/51/70/78/161/166 (`:589,592,600,610,612,614,615,618,704,707,712`);
  - the cards for ADR-000 `:736`, ADR-003 `:807`, ADR-007 `:901`, ADR-008 `:925`, ADR-011 `:991` and ADR-012 `:1001`;
  - `README.md`, `STATUS.md:7-9,117-123,133`, plan `:84-99` and `research-ledger.md:3,72-74`.
- Every ADR status line under `docs/adrs/*/ADR.md:3`, compared with base for ADR-007 and ADR-008.
- `adrs/2026-10-23-decisions/ADR.md:11,14,34` (ADR-003).
- Code:
  - `.github/workflows/scout-guardian.yml:5-7,42-43,50,58-60,79,82-85,102-105,116-118`;
  - `scout-guardian-trusted.yml:9-10`;
  - `scripts/ci-toolchain.sh` and `scripts/bootstrap-tools.sh:59-73,93-97,104-118`;
  - `internal/engine/diff.go:106-157,221-228,275,341-368`;
  - `internal/state/export.go:40,258,683-695`;
  - `internal/policy/policy.go:300-330`;
  - `internal/guardrails/{suppress.go:49,focused_tests.go:9,config.go:39}`;
  - `internal/analyzers/catalog_security.go:221-225`.
- X1 worktree: `scout-x1/docs/research-ledger.md:71-78`.
- External: GitHub REST documentation for artifacts and workflow runs, `actions/upload-artifact` v4 migration notes, the GitHub dependency-caching reference, and the `pull_request` merge-branch documentation (links under Evidence).

## Issue summary

**19 distinct issues are open: 2 blocking, 7 major and 10 minor.**

- **ADR-011 has 5 issues** (1 major, 4 minor). Three of them concern shared files, so they block ADR-012 as well: ADR011-R2-01, -02 and -04.
- **ADR-012 has 14 issues of its own** (2 blocking, 6 major, 6 minor). With the three shared issues, 17 issues block ADR-012.

| ID | Severity | Blocks | Short title |
|---|---|---|---|
| ADR011-R2-01 | major | both | The `ADOPTION.patch`/hash split has no workable adoption order with X1 or with a squash merge |
| ADR011-R2-02 | minor | both | Shared ledger rows R-064/R-069 freeze "owner review pending" and fix ADR-012's design choices in place |
| ADR011-R2-03 | minor | ADR-011 | REQ-22 "auto-demote it" now points at the wrong thing |
| ADR011-R2-04 | minor | both | The reset-gap statement leaves out Scout-native guardrails checks, whose versions are maintained by hand |
| ADR011-R2-05 | minor | ADR-011 | The patch leaves false ADR ranges and puts an ADR note under the "design rules" heading |
| ADR012-R2-01 | blocking | ADR-012 | The producer is not authenticated: scripts, inputs, callee revision, PR-scoped toolchain cache and non-default bases |
| ADR012-R2-02 | blocking | ADR-012 | Artifacts belong to the run, not the attempt or job; a fixed name collides on reruns and sibling jobs can upload it |
| ADR012-R2-03 | major | ADR-012 | Rejecting every workflow-directory change leaves the action-ref check unmeasurable and inflates unknowns |
| ADR012-R2-04 | major | ADR-012 | CI "head" and "merge-base" are undefined; head validation and fork PR keys cannot be done as written |
| ADR012-R2-05 | major | ADR-012 | Provider-state classification is incomplete and biased: cancellations, pre-assessment failures, approval waits, `started_at` |
| ADR012-R2-06 | major | ADR-012 | A reconciliation deadline tied to artifact retention stalls coverage and mixes collector failure with producer outcome |
| ADR012-R2-07 | major | ADR-012 | The REQ-22 rewrite drops owner-decided clauses, and the ADR-000/ADR-003 amendments are not recorded |
| ADR012-R2-08 | major | ADR-012 | Collector deployment: adding `actions: read` breaks existing callers, and the producer configuration source is missing |
| ADR012-R2-09 | minor | ADR-012 | The canonical revision rolls back ADR-011's text and leaves the STATUS ADR range stale |
| ADR012-R2-10 | minor | ADR-012 | Local-scope would-block events are exported and later miscounted as legacy |
| ADR012-R2-11 | minor | ADR-012 | Only the latest contiguous interval can be computed, so an activation dossier cannot be reproduced after a later gap |
| ADR012-R2-12 | minor | ADR-012 | `readiness --shadow` takes the version from the installed binary, not the CI pin |
| ADR012-R2-13 | minor | ADR-012 | Inspection predicate edge cases: guardrails base condition and adapter `not_applicable` units |
| ADR012-R2-14 | minor | ADR-012 | No-op re-pushes add observations, and how many PRs are behind the count is not disclosed |

## Round-1 closure

| Round-1 issue | Status | Evidence / residual |
|---|---|---|
| ADR011-R1-01 | **Closed** | ADR-011 `ADR.md:15,17` separates implemented rule-content resets from the pending Gitleaks binary-segment enforcement and the open Semgrep gap. The same split is in patch REQ-22 `:8`, REQ-78 `:17`, the card `:25-34`, the policy reference, STATUS and the plan. One residual on the same claim, the omitted guardrails checks, is new: ADR011-R2-04. |
| ADR011-R1-02 | **Closed** | `ADR.md:3,43` and the card `rel` carry no pending state. Pending state appears only in the README, STATUS and plan trackers. The ledger-row counterpart is new: ADR011-R2-02. |
| ADR012-R1-01 | **Closed** | `ADR.md:16,22`: version-scoped new IDs, no overwrite, `pr_run_events` many-to-many, sorted `eventIds`, `would_block.jsonl` unchanged. Mirrored in REQ-27/48. |
| ADR012-R1-02 | **Partly closed** | The gates and the `untrusted_producer` terminal state are present (`ADR.md:34-38`), and R-070 was added. They still do not authenticate the producer or the artifact: ADR012-R2-01 and -02. |
| ADR012-R1-03 | **Closed** | Epochs removed; exact `(check_id, rule_version)` cohorts (`ADR.md:50`, REQ-78). |
| ADR012-R1-04 | **Closed** | `started_at` membership, the elapsed-days formula, default `as_of = collectedThrough`, and exit 3 for a future cutoff (`ADR.md:52,65,68`). The meaning of `started_at` in CI is a residual inside ADR012-R2-05. |
| ADR012-R1-05 | **Partly closed** | Terminal states and gap boundaries exist (`ADR.md:42-44`), but provider states are not exhaustive (ADR012-R2-05) and the retention-based deadline stalls coverage (ADR012-R2-06). |
| ADR012-R1-06 | **Closed** | Distinct `(PR key, head, merge-base)` observations, with the latest terminal assessed attempt as representative (`ADR.md:52`). Residuals: identity, rerun collisions and disclosure (ADR012-R2-04, -02, -14). |
| ADR012-R1-07 | **Closed** | Three statuses; "count toward PRRuns" leaves comparability unchanged (`ADR.md:54-60`). Edge cases in ADR012-R2-13. |
| ADR012-R1-08 | **Closed** | Scope and provider on facts and coverage; local facts never exported (`ADR.md:16,46`). Local *events* are a residual: ADR012-R2-10. |
| ADR012-R1-09 | **Closed** | Manifest `scout.state/v2`, with `state/v1/` kept as a layout name (`ADR.md:82-84`). The old-binary refusal follows `export.go:693`. |
| ADR012-R1-10 | **Closed** | pending<final, earliest-wins provenance, named `MergeState` changes (`ADR.md:20,86`). |
| ADR012-R1-11 | **Closed** | The folders hold ADR-007 and ADR-008, and only line 3 differs from base. README rows and cards `:901,925,1001` agree. Refuse-13 is rejected (`ADR.md:78`). The ADR-000/003 amendments are new: ADR012-R2-07. |
| ADR012-R1-12 | **Closed** | REQ-166 `:712`, REQ-47 `:614`, REQ-39 `:704` and REQ-161 `:707` are aligned and attributed. |
| ADR012-R1-13 | **Partly closed** | Command, triggers, artifact schema, caps and pagination are specified (`ADR.md:26-30`). Not yet specified: binding an artifact to its attempt, the producer configuration source, and caller permissions (ADR012-R2-02, -08). |
| ADR012-R1-14 | **Closed** | `ADR.md:74,92`, REQ-10 `:610` and the REQ-78 exit table. |
| ADR012-R1-15 | **Closed** | `ADR.md:3`, card `:1001` and README tracker row. |
| ADR012-R1-16 | **Closed** | `STATUS.md:9` uses conditional wording. Stale ADR ranges elsewhere are new: ADR011-R2-05 and ADR012-R2-09. |

## Issues: ADR-011

### ADR011-R2-01: The `ADOPTION.patch`/hash split has no workable adoption order with X1 or with a squash merge (major; shared, blocks both)

- **Location:** ADR-011 `ADR.md:41` ("Adoption applies that isolated revision against `ef6f50e8`"); `ADOPTION.patch` ledger hunk `@@ -69,3 +69,5 @@` (`:121`), which has no trailing context; `ADOPTION-SHA256.txt`; ADR-012 Review ("Apply/adopt in that order"); round-1 "Considered and not raised" condition ("X1's R-065 to R-068 land before or alongside R-069").
- **Problem:**
  - A hunk with no trailing context must match at end of file under `git apply` [INFERENCE from git's `match_end` rule]. Any row appended to the ledger on main makes the patch fail, or makes its result differ from the signed hash.
  - `scout-x1/docs/research-ledger.md:72-78` already carries R-064, almost certainly the same row: same title, same verbatim quote, and a link into ADR-011's folder. It also carries R-065 to R-068, R-071 and R-072.
  - If X1 lands first, as round 1's condition allows, ADR-011's signed result can never exist on main, and X1's R-064 links to a file that does not exist yet.
  - If both ADRs land in one PR that is squash-merged, the `ef6f50e8` + patch state never exists as a commit. ADR-011's signed revision would then be a virtual state, not an adopted one.
  - Neither ADR's signed Review states the merge method or the X1 order.
- **Required fix (recommended option):**
  1. In both Review sections, state the adoption sequence.
     - Commit A applies `ADOPTION.patch` and adds ADR-011's folder records. Its parent is `ef6f50e8`, or any commit whose six shared files are byte-identical to `ef6f50e8`'s.
     - Commit B adds ADR-012's canonical bytes, its folder and the historical status lines.
     - Merge so that both commits are kept (merge commit or fast-forward; no squash).
     - Main verifies all six result hashes on commit A and the canonical digests on commit B.
  2. X1 lands after commit B. It drops its duplicate R-064 and appends R-065 to R-068, R-071 and R-072 after R-070. Add one sentence to the ledger introduction or STATUS saying that rows are in append order and that IDs were pre-allocated.
  3. If any of the six files changes on main before adoption, regenerate the patch and the result hashes and re-review.
  4. This round corrects round 1's coordination condition to "X1 lands after both ADR revisions".

### ADR011-R2-02: Shared ledger rows R-064/R-069 freeze "owner review pending" and fix ADR-012's design choices in place (minor; shared, blocks both)

- **Location:** the R-064 and R-069 rows (`ADOPTION.patch:125-126`), byte-identical in canonical `research-ledger.md:72-73`; ledger rule "append only" (`research-ledger.md:3`); ADR-011 `ADR.md:24,36`.
- **Problem:**
  - R-064's caveat says "owner review pending". That is the same frozen-state defect as ADR011-R1-02, but in an append-only row.
  - R-069's key point records ADR-012 *selections*: "select new version-scoped IDs and attempt↔event links…", "manifest v2", "Distinct revision observations…".
  - ADR-011's adoption would put ADR-012 design choices that are not yet signed into an append-only record. That contradicts ADR-011's own claim of isolation (`:24`, "prevents ADR-012 redesign from renewing ADR-011 review"). Any later change to ADR-012 would need a correction row; R-060 is the precedent.
- **Required fix:**
  - R-064 caveat: "Agent-made choices under standing delegation; the owner may review later, recorded as a new `user` row. Does not waive…".
  - R-069 key point: keep only the static code facts (what the code does today) and remove every "select…" clause. ADR-012's `ADR.md` already records the selections and cites R-069 for the facts.
  - Apply the identical revised rows in both revisions.

### ADR011-R2-03: REQ-22 "auto-demote it" now points at the wrong thing (minor)

- **Location:** `ADOPTION.patch:8`: "…Semgrep action-ref parser-bump reset remains a separate open ADR gap; at least 3 confirmed-unwarranted blocks in the trailing 20 blocks auto-demote it to report for that repo…".
- **Problem:** The new clause sits in front of the auto-demotion rule, so "it" now reads as the Semgrep gap or the action-ref check rather than "that repo's check". That changes the apparent scope of an ADR-000 rule.
- **Required fix:** End the Semgrep clause with a full stop, and write "…auto-demote that repo's check to report for that repo and record it". The canonical text (`scout-design.html:589`) already splits the sentences, so this fix is ADR-011-only. ADR012-R2-07 restores "for that repo" in the canonical text.

### ADR011-R2-04: The reset-gap statement leaves out Scout-native guardrails checks, whose versions are maintained by hand (minor; shared, blocks both)

- **Location:**
  - REQ-22 in `ADOPTION.patch:8` and canonical `:589` ("Rule-content bumps reset through exact rule-pack equality (implemented)"; only Semgrep is named as an open gap);
  - STATUS paragraph (`ADOPTION.patch:63`) and canonical `STATUS.md:121-123`;
  - plan `:87-89`;
  - `policy-file.md` note.
- **Problem:**
  - Exact equality is implemented, but content is bound to the version only for bundled fingerprinted rule files (`spec.go:224`).
  - The eligible guardrails checks (`policy.go:319-320` plus the CONFIG families) use version constants maintained by hand: `suppress.go:49` `…unscoped@1`, `focused_tests.go:9` `…@1`, `config.go:39` `@1`.
  - Their "rule" and "parser" are Scout's own Go code, so a logic change that does not bump the constant does not return the check to shadow.
  - Naming only Semgrep as the open gap implies that every other check's parser bump is covered.
- **Required fix:**
  - Scope "(implemented)" to exact-version equality, with content binding enforced only for bundled fingerprinted packs.
  - Add Scout-native guardrails checks (reset only when someone bumps the `@N` constant, a convention nothing enforces) to the open known gaps next to Semgrep, in STATUS, the plan, REQ-22 and the policy note, for a later ADR.
  - Mirror the change in the canonical revision.

### ADR011-R2-05: The patch leaves false ADR ranges and puts an ADR note under the "design rules" heading (minor)

- **Location:**
  - `STATUS.md:7` in the patch result ("decisions ADR-000…ADR-010 in `docs/adrs/`", "Latest amendment: … ADR-010");
  - plan `:96` in the patch result ("Appendix B indexes ADR-000…ADR-010"; the patch does not touch this line);
  - README: the patch inserts an ADR-011 bullet into the list headed "Rules (`AGENTS.md` design rules)".
- **Problem:**
  - As soon as the patch is applied, Appendix B contains an ADR-011 card and `docs/adrs/` contains its folder, so both range statements are false in the signed result bytes.
  - The README bullet is not an `AGENTS.md` design rule but sits under a heading that says it is.
- **Required fix:**
  - Update both ranges with conditional wording, for example "ADR-000…ADR-011 (ADR-011 effective on its final zero-issue round)".
  - Move the ADR-011 note out of the rules list into a separate paragraph below it.

## Issues: ADR-012

### ADR012-R2-01: The producer is not authenticated: scripts, inputs, callee revision, PR-scoped toolchain cache and non-default bases (blocking)

- **Location:** ADR-012 `ADR.md:30,34-38`; REQ-51 (`:618`); REQ-161 (`:707`).
- **Problem:** Gates 1–3 check two directories, workflow and job *names*, and pins that the producer reports about itself. They do not establish that only trusted code produced the payload. Any of the following lets a PR fabricate an accepted "inspected, zero-event" payload, copying the public version and digest strings to pass gate 3:
  1. **Unprotected inputs and dependencies.** An unchanged caller workflow can:
     - run a repository script outside `.github/**`;
     - use a local action outside `.github/actions/**`;
     - take `scout-version`/`release-repo` from an expression such as `${{ vars.X }}`, which is not versioned.

     The callee's resolved revision (`referenced_workflows[].sha` on the workflow-run object) is not checked either. These inputs control which code runs (`scout-guardian.yml:69-79`).
  2. **PR-writable toolchain cache.** The PR lane restores `actions/cache` keyed `scout-toolchain-<os>-<arch>-<scout-version>` (`scout-guardian.yml:82-85`).
     - A cache saved during a `pull_request` run is scoped to `refs/pull/<n>/merge` and is restored first by later runs on that PR (GitHub dependency-caching reference).
     - Any workflow run on that PR that executes PR code can save that key [INFERENCE]: an existing test workflow, or, once ADR012-R2-03 narrows the gate, a newly added workflow.
     - `bootstrap-tools.sh:94` accepts a cached Gitleaks if *executing* it prints the pinned version, and node and Python tools are reused on a matching stamp (`:73,104-118`). The poisoned binary therefore runs inside the producer job.
  3. **Non-default bases.** Enumeration takes every `pull_request` run of the path, whatever its base ref. A PR into an unprotected branch that a collaborator controls gets pins chosen by the attacker "at merge-base".
  4. **Pins of other versions.** Gate 3 does not say where the tool digest and catalogue versions for a Scout version other than the collector's own come from. It also does not say which catalogue assigns the cohort for an untrusted observation. During a version transition, attempts on the old version could fall into the wrong cohort.
- **Required fix (recommended option: a verified, dedicated producer):**
  1. **Eligibility.** Only `pull_request` runs whose base ref is the default branch, and whose comparison base is reachable from the protected default branch, are eligible. All others are final `not_assessed` with reason `non_default_base`, counted and disclosed.
  2. **Producer recipe.** The producer is declared in trusted base policy (key defined in ADR012-R2-08). At the comparison base, the collector parses the caller without executing it. The caller must contain exactly one job, which calls `<release-repo>/.github/workflows/scout-guardian.yml@<tag>` with literal `scout-version` and `release-repo` values, no expressions. If not, every attempt is `untrusted_producer`.
  3. **Callee revision.** The run's `referenced_workflows` path and SHA must equal the pinned release tag's commit in the pinned release repository.
  4. **No PR-writable state.** Remove the `actions/cache` restore from the PR lane, which amends REQ-162 (ADR-007, already amended by ADR-012) and becomes a delivery condition. Verifying cached executables against per-file digests is the rejected alternative: it is more complex and easy to get wrong.
  5. **Pin source.** Resolve the tool digest and catalogue rule versions from the pinned release's published metadata, SHA-256-verified as in REQ-163, not from the collector's binary. Untrusted observations get `not_inspected` statuses from that release's catalogue. An unresolvable pin is a collection gap.
  6. **Records.**
     - Update REQ-51/161/162.
     - Add delivery proofs:
       - a modified script or local action outside `.github/**`;
       - an input given as an expression;
       - a changed callee SHA;
       - a poisoned PR-scoped cache;
       - a PR with a non-default base.
     - Add research-ledger rows for the GitHub workflow-run and dependency-caching documentation.

### ADR012-R2-02: Artifacts belong to the run, not the attempt or job; a fixed name collides on reruns and sibling jobs can upload it (blocking)

- **Location:** ADR-012 `ADR.md:18,28,35`; REQ-51.
- **Problem:**
  - The artifacts REST object carries `workflow_run.id` but no run attempt and no job (GitHub artifacts API).
  - Under `upload-artifact@v4`, a second upload of the same name in the same run conflicts unless `overwrite: true` is set, and that option deletes the earlier artifact (v4 migration notes).
  - **Reruns.** With the fixed name `scout-pr-run`, a rerun attempt either fails to upload and becomes an `artifact_unavailable` representative (turning a clean observation unknown) or erases the earlier attempt's evidence.
  - **Attempt binding.** The rule "validate run/attempt … not self-report" (`:35`) cannot be applied to artifacts.
  - **Sibling jobs.** Any other job in the same run, such as an adopter's PR test job running PR code, can upload the only artifact with that name [INFERENCE about adopter workflows]. The rule "reject multiple same-name artifacts" (`:28`) does not catch a single forged upload.
- **Required fix (recommended option):**
  - The pinned callee names the artifact `scout-pr-run-<run_attempt>`.
  - Accept exactly one artifact per attempt, with:
    - a name suffix equal to the provider attempt;
    - payload run and attempt equal to the provider's;
    - `created_at` between the provider start and completion of that attempt's assessment job.

    Record the artifact ID and digest in provenance.
  - ADR012-R2-01's single-job caller makes the producer job the only job in the run that can upload. Jobs from other workflows run in other runs and cannot reach this run's artifacts.
  - An attempt in which the assessment job did not execute (for example, a partial "re-run failed jobs") is not a producer attempt and can never be a representative.
  - Ambiguous, misnamed or out-of-window artifacts mean `untrusted_producer`.
  - The provider OIDC attestation alternative is rejected because fork PR runs do not receive `id-token: write` [INFERENCE].
  - Add a ledger row for the GitHub artifacts and `upload-artifact` documentation.

### ADR012-R2-03: Rejecting every workflow-directory change leaves the action-ref check unmeasurable and inflates unknowns (major)

- **Location:** ADR-012 `ADR.md:34,38,58,63,65,70`; REQ-70 (`:592`); `catalog_security.go:221-225`.
- **Problem:**
  - `security.ci-mutable-action-ref` only inspects `.github/workflows/*.y*ml`. Every observation with one of its units fails gate 1 and becomes `not_inspected`; every other observation is `not_applicable`.
  - The check therefore never gets an inspected observation. `Days` stays unmeasured, PRRuns stays 0, it can never have an accepted event, and the rare-security provisional route (`:70`, and `ShadowReady` applying days first at `policy.go:412-416`) is unreachable.
  - Separately, UnknownAttribution is counted per observation for all checks (`:63`). Every Dependabot GitHub Actions bump adds an unknown to every check's cohort, and in small repositories that alone can exceed the 10% bound.
- **Required fix:** With ADR012-R2-01 in place, replace gate 1 with "the PR does not change the declared caller file". Changes to other workflows cannot affect the producer run, since they run as other workflows in other runs, and they become ordinary scanned input. Add delivery proofs for an action-ref PR and a Dependabot workflow bump. This fix depends on removing the cache (R2-01 step 4); without that, narrowing the gate opens the cache channel.

### ADR012-R2-04: CI "head" and "merge-base" are undefined; head validation and fork PR keys cannot be done as written (major)

- **Location:** ADR-012 `ADR.md:16,35,52`; REQ-22/27/78; `scout-guardian.yml:58-60,102-103`; `diff.go:129-151`.
- **Problem:**
  - The checkout has no `ref`, so a `pull_request` run checks out the synthetic merge commit M (GitHub `pull_request` documentation). Scout records `head = HEAD = M` and computes `mergeBase = merge-base(BASE_SHA, M)`, which is the base tip (`diff.go:140-151`).
  - The provider run's `head_sha` is the PR head H, not M. Read literally, "validate head against provider data" rejects honest payloads.
  - Old merge commits are generally not fetchable later [INFERENCE], so "unavailable git objects fail closed" turns them into unknowns.
  - For fork PRs the run object's `pull_requests` list is empty [INFERENCE], so the PR key cannot come from the run.
  - Implementations could choose incompatible meanings, which would change deduplication and "trusted pins at merge-base".
- **Required fix (recommended option):**
  - Persist `pr_head_sha` (provider `head_sha`), `base_sha` (the supplied comparison base and policy source), `assessed_revision` (M, provenance only) and `merge_base` (Scout's comparison base).
  - Observation key = (PR key, `pr_head_sha`, `merge_base`).
  - Validate `pr_head_sha` against the provider. Evaluate the caller-file gate on `merge_base..pr_head_sha` using the provider compare API or the fetched PR head. Never require M.
  - An unverifiable revision is final unknown with reason `unverifiable_revision`.
  - PR key = (base repository ID, PR number), from the run's `pull_requests` or, when that is empty, from the commit-to-PR API filtered by head repository. If the key is ambiguous, the attempt is `not_assessed` and disclosed.
  - Mirror this in REQ-22/27/78.

### ADR012-R2-05: Provider-state classification is incomplete and biased: cancellations, pre-assessment failures, approval waits, `started_at` (major)

- **Location:** ADR-012 `ADR.md:18,42,44`; REQ-51; `scout-guardian.yml:42-43,50,72-79`.
- **Problem:**
  1. **Superseded pushes count as unknown.** The PR lane uses `cancel-in-progress: true`, and assessment is the long step (`timeout-minutes: 30`). Each new push therefore cancels the previous head *after* assessment began, which makes it an assessed `artifact_unavailable` unknown. The unknown ratio ends up tracking how often developers push, not attribution.
  2. **Pre-assessment failures have no class.** A job that fails or times out before the assessment step (for example, the install step exiting 3) is not classified.
  3. **Unstarted runs stall coverage.** Runs that are `action_required` (fork approval), waiting or queued may never complete. Because "the provider cutoff precedes any still-pending start" (`:44`), they hold `collectedThrough` indefinitely.
  4. **`started_at` is ambiguous.** It does not say whether this is run creation, run start or job start.
- **Required fix (recommended option): an exhaustive matrix keyed on the producer's assessment job.**
  - Job never started (run cancelled, skipped, `action_required`, or failed before the job), or job ended before the assessment step began: `not_assessed` with a reason, excluded and disclosed.
  - Assessment step began, then cancelled: `not_assessed`, reason `cancelled`, disclosed.
  - Assessment step began, then `failure`, `timed_out` or `success` without a valid artifact: assessed unknown.
  - Missing or renamed producer job in a verified producer run: `untrusted_producer`, which keeps the job-rename protection.
  - CI `started_at` is the provider start time of the assessment job. Runs whose job has not started do not hold the cutoff.
  - Mirror in REQ-51/78.

### ADR012-R2-06: A reconciliation deadline tied to artifact retention stalls coverage and mixes collector failure with producer outcome (major)

- **Location:** ADR-012 `ADR.md:42,44`; REQ-51/161 ("reconciles through retention deadlines").
- **Problem:**
  - Uploads happen inside the run, so once the producer job has completed its artifact set is final. Waiting "run completion plus retention" holds `collectedThrough` for up to the retention period (30 days in the current PR lane, `scout-guardian.yml:118`) after any single failed upload.
  - Repository-default retention needs Administration read access (GitHub Actions permissions API), which the token lacks. Under `:42`, missing retention metadata becomes a collection gap, so ordinary upload failures reset cohorts. That contradicts `:44` ("an actual terminal unknown … cannot reset a cohort").
  - An artifact that expires before the collector downloads it is a collector failure, but the current rule counts it as a producer unknown.
- **Required fix (recommended option):**
  - Finalize at the first collector invocation at least 1 hour after the provider reports the assessment job complete. If no valid artifact is listed then, the attempt is `artifact_unavailable` (measured unknown).
  - An artifact that existed but expired before download is a collection gap, not an unknown.
  - Drop the deadline derived from retention, and the reads of retention metadata.

### ADR012-R2-07: The REQ-22 rewrite drops owner-decided clauses, and the ADR-000/ADR-003 amendments are not recorded (major)

- **Location:**
  - canonical REQ-22 `:589` compared with base (`ADOPTION.patch:7`);
  - ADR-000 `ADR.md:3` and card `:736`; ADR-003 `ADR.md:3,11,34` and card `:807`; README rows;
  - ADR-012 card `rel` `:1001` and Changes.
- **Problem:**
  - The rewrite drops "during the window, **maintainers** triage", "return **that repo's** check", "auto-demote … **for that repo**" and "zero events **at window end**". All four are decided items, and none of the four drops is listed in Changes.
  - ADR-012 changes ADR-000's "5 comparable-base PR runs" and ADR-003's decision to "count a PR run once toward the ≤ 10 % unknown-attribution limit" (ADR-003 `:34`) into distinct observations.
  - It also adds collector and `actions: read` authority to the trusted runs, which ADR-003 `:11` defined.
  - ADR-010 is the precedent: changing REQ-98 made ADR-003 "amended by ADR-010".
  - R-064 delegates *recommended choices captured as ADRs*. Unrecorded changes to items the owner decided are not captured for the owner's later review.
- **Required fix:**
  - Restore the four clauses verbatim.
  - ADR-000 status: "…; amended by ADR-001, ADR-002, ADR-003, ADR-012 (readiness counting unit)".
  - ADR-003 status: "partly superseded by ADR-007; amended by ADR-010 (triage minimums), ADR-012 (readiness counting unit, trusted collector permission)".
  - Update the README rows, the ADR-000/003 card `rel` spans and ADR-012's `rel` and Changes.
  - Add both `ADR.md` files to the signed set.
  - List the owner-decided items ADR-012 changes under one heading for the owner's review.

### ADR012-R2-08: Collector deployment: adding `actions: read` breaks existing callers, and the producer configuration source is missing (major)

- **Location:** ADR-012 `ADR.md:26,30,35`; REQ-161; `scout-guardian.yml:5-7`; `scout-guardian-trusted.yml:9-10`.
- **Problem:**
  - GitHub checks every job of a called workflow against the caller's permissions, regardless of `if:` (`scout-guardian.yml:5-7`), and today's callers grant only `contents: write` and `pull-requests: read` (`scout-guardian-trusted.yml:9-10`).
  - A release that adds `actions: read` to the trusted workflow therefore fails for every existing caller before any step runs, including baseline publish.
  - The ADR also does not say where "the PR workflow path configured on the default branch" and the job identity are declared.
- **Required fix:**
  - Declare a breaking caller change. Callers must grant `actions: read`, recorded in release notes, the adopter checklist (REQ-168) and the CI reference, with a minimum-version note.
  - Define a trusted base-policy key, for example `ci.pull_request_workflow: {path, job}`, documented in `policy-file.md`. If the key is absent, collection is `unmeasured` with a reason, never guessed.
  - Update REQ-161.

### ADR012-R2-09: The canonical revision rolls back ADR-011's text and leaves the STATUS ADR range stale (minor)

- **Location:**
  - README ADR-011 row: the patch says "…enforce binary-segment invariant in later code"; canonical says "…no migration";
  - canonical REQ-78 `:600` drops the explicit statement that the Gitleaks pack version segment must equal the binary pin;
  - canonical `STATUS.md:7` still says "ADR-000…ADR-010" and "Latest amendment: … ADR-010".
- **Problem:** The canonical revision is meant to build on ADR-011's revision, but it removes ADR-011 commitments from tracker and requirement text. It also repeats the false range from ADR011-R2-05.
- **Required fix:**
  - Restore ADR-011's index decision text and the explicit REQ-78 invariant.
  - Make `STATUS.md:7` conditional on ADR-011/012 taking effect.
  - Submission condition: supply the diff from ADR-011's result to the canonical files as a review aid.

### ADR012-R2-10: Local-scope would-block events are exported and later miscounted as legacy (minor)

- **Location:** ADR-012 `ADR.md:22,46,66,82`; `diff.go:221-228`.
- **Problem:**
  - Local `scout diff` writes would-block rows. Local *facts* and links are excluded from export, but events are not.
  - After a restore, local events show up unlinked and are counted and disclosed as "legacy" (`:66`).
- **Required fix:**
  - Mark legacy explicitly at migration (a pre-v14 marker) rather than inferring it from "unlinked".
  - Give events created without a CI fact `scope: local`, and exclude them from export and publish, as for local facts.
  - Disclose legacy and local counts separately.

### ADR012-R2-11: Only the latest contiguous interval can be computed, so an activation dossier cannot be reproduced after a later gap (minor)

- **Location:** ADR-012 `ADR.md:44,52,68,82`.
- **Problem:**
  - After a later gap, an `as_of` inside an earlier interval exits 2, so the evidence behind an activation (`dossier_digest`) cannot be recomputed.
  - Export carries provenance for the latest interval only, although `:44` preserves old gaps.
- **Required fix:**
  - Compute over "the contiguous reconciled interval containing `as_of`".
  - Export all coverage intervals and gaps.

### ADR012-R2-12: `readiness --shadow` takes the version from the installed binary, not the CI pin (minor)

- **Location:** ADR-012 `ADR.md:90`.
- **Problem:** A maintainer whose local Scout differs from the CI pin queries a different cohort and gets `unmeasured` or zeros, with no warning.
- **Required fix:**
  - Default to the rule version of the pinned release recorded by the latest CI collection.
  - Add `--version <exact>`.
  - Print any mismatch with the installed version. An unknown version exits 3.

### ADR012-R2-13: Inspection predicate edge cases: guardrails base condition and adapter `not_applicable` units (minor)

- **Location:** ADR-012 `ADR.md:54-58`; REQ-27 `:612` ("non-guardrails finding"); `STATUS.md:115-116`.
- **Problem:**
  - Guardrails do not run on `assess` (they are `skipped`) and read the base tree inline in `diff`. So "the base inspected the check's family" can be read as excluding guardrails entirely, which would be stricter than REQ-27's guardrails exemption.
  - It is unstated whether units with adapter coverage `not_applicable` count as eligible units. Gitleaks allowlisted paths such as lockfiles are an example. If they count, a PR that touches a lockfile is never inspected for secrets.
- **Required fix:** State both rules:
  - a guardrails check meets the base condition when its inline lane completed;
  - units whose adapter coverage is `not_applicable` are not eligible units.

### ADR012-R2-14: No-op re-pushes add observations, and how many PRs are behind the count is not disclosed (minor)

- **Location:** ADR-012 `ADR.md:52,64`; REQ-22 `:589`.
- **Problem:**
  - An amend and force-push with an identical tree creates a new head and a new inspected observation of the same diff.
  - Five pushes to one PR meet `min_pr_runs`, and the output does not reveal that concentration.
- **Required fix:**
  - Deduplicate observations by (PR key, head tree OID, `merge_base`).
  - Report `distinctPRs` in output.
  - State in REQ-22 that successive distinct heads of one PR count separately [ASSUMPTION].

## Checks that found no issue

- **Patch structure.** The patch changes exactly the six shared files and no `ADR.md`. The hash list names those six paths, so nothing hashes itself. The card, index, STATUS and plan additions in the patch do not depend on ADR-012.
- **Ledger rows.** R-064 and R-069 are byte-identical in the patch and the canonical ledger. R-064's quote is identical in ADR-011 `:28`, ADR-012's Informed by and the ledger. R-070's format and links are fine. IDs do not collide: X1 holds R-065 to R-068, R-071 and R-072; this unit holds R-064, R-069 and R-070. The duplicate R-064 is handled in ADR011-R2-01.
- **Historical status lines.**
  - `2026-10-24-cutover-cleanup` holds ADR-007 and `2026-10-31-jev-merge-association` holds ADR-008. Compared with `/tmp/adr11-patchcheck`, only line 3 differs in each.
  - ADR-007 reads "partly superseded by ADR-008; amended by ADR-012 (13→14 migration)" and ADR-008 reads "accepted (provisional); amended by ADR-012 (schema 14)". Both match the README rows and the card `rel` spans `:901,925`.
  - Their bodies and review records are intact. Changing their digests through the status line is permitted by design rule 2.
- **Arithmetic and exits.** 14 days and 5 observations, the integer ≤10% bound, the order of the provisional exception and the `--shadow` exit table agree across ADR-012 `:62-70,92`, REQ-22/78 and `policy.go:68-69,402-424`. The old-binary refusal relies on `export.go:693` (`parsed.Schema != ExportSchema`).
- **Delegation record.**
  - Both ADRs say they are agent decisions under R-064, that delegation is not owner review and does not waive council review, and that the owner confirms through a new ledger row and rejects through a superseding ADR.
  - The agent-selected caps (`ADR.md:28`) are disclosed as agent choices, which is within R-064.
  - The one delegation-scope defect, unrecorded changes to owner-decided clauses, is ADR012-R2-07.

## Considered and not raised

- **Cohort key ignores tool and policy digests (peer).** ADR-012 `:50` states explicitly that the cohort key is exactly `(check_id, rule_version)`, and with ADR-011's invariant the Gitleaks binary is bound to it. This is a stated choice, not an ambiguity.
- **Plan `:99` "review candidates until council adoption" (peer).** Round 1 accepted this conditional wording, which stays true after adoption. Only the false ranges are raised (ADR011-R2-05, ADR012-R2-09).
- **Sign an `ADR-012.patch` as well (peer).** Canonical file hashes identify ADR-012's bytes exactly. The ADR-011→ADR-012 diff is required as a review aid (submission condition), not as a signed artifact.
- **ADR-011 amending ADR-000.** ADR-011 keeps ADR-000's parser-bump intent, adds a Gitleaks mechanism to meet it, and discloses gaps. "Clarifies" is accepted, and no status-line change is required.

## Peers and where they changed the outcome

- **`innovation-challenger`** raised 5 ADR-012 findings and none against ADR-011. It confirmed 14 round-1 closures and 4 partial closures, the patch structure, and the ADR-007/008 line-3-only changes. Adopted:
  - producer recipe and `referenced_workflows` (into R2-01);
  - artifact provenance (R2-02);
  - action-ref unmeasurability (R2-03);
  - revision conflation (R2-04);
  - pre-assessment failures (into R2-05).
- **`experiment-peer`** raised 23 findings. Adopted or merged:
  - adoption order and squash (ADR011-R2-01);
  - ledger rows (ADR011-R2-02, with the R-064 caveat merged in);
  - "it" antecedent (R2-03); guardrails versions (R2-04, severity lowered to minor); stale ranges (R2-05);
  - gate 1 and Dependabot (R2-03); fork PR key (into R2-04); attempt-named artifacts and partial reruns (into R2-02);
  - retention metadata (into R2-06); cancellations and `action_required` (into R2-05);
  - non-default bases (into R2-01); sibling jobs (into R2-02);
  - re-push inflation (R2-14, as a disclosure fix); dropped clauses and ADR-003 (R2-07); ADR-011 rollback (R2-09);
  - caller permissions and configuration (R2-08); local events (R2-10); interval reproducibility (R2-11); CLI version (R2-12); inspection edges (R2-13).

  Declined: the cohort tuple and plan `:99` wording. The peer's claim that X1 lacks R-064 was superseded by the lead's current read of `scout-x1/docs/research-ledger.md:72`.
- **Lead:** independently found:
  - toolchain-cache poisoning: PR-scoped cache, and a cached Gitleaks binary that is executed rather than verified (`bootstrap-tools.sh:94`);
  - mid-assessment cancellation bias from `cancel-in-progress`;
  - the retention-deadline stall;
  - the merge-commit versus PR-head conflation;
  - the end-of-file anchoring of the patch;
  - the false ADR ranges and the README heading placement.
- **Changes to the lead's position:**
  - The challenger's `referenced_workflows` and dynamic-input points widened R2-01 from cache and sibling jobs to a full producer recipe.
  - The peer's non-default-base attack and the cost of reading retention metadata were added to R2-01 and R2-06.
  - The peer's evidence of dropped clauses turned the lead's ADR-003 status-line concern into the major R2-07.
- **Severity calls:**
  - The peer's ADR011-R2-01 (R-069 content) went from major to minor: a correction row is a recoverable remedy (R-060 precedent).
  - The peer's guardrails finding went from major to minor: the equality mechanism is implemented, and only the gap list is incomplete.
  - The peer's blocking gate-1 finding went to major: it fails safe, and the action-ref check is unmeasurable rather than forgeable.
- **Expired artifacts:** the council chose collection gap over the peer's "unknown" for artifacts that expire before download. That keeps collector failure apart from producer outcome, as `ADR.md:44` intends.

## Verdict

**ADR-011: 5 issues open (1 major, 4 minor). Not signed off.** Open: ADR011-R2-01, -02, -03, -04 and -05.

**ADR-012: 17 issues open (2 blocking, 7 major, 8 minor). Not signed off.** These are its own 14 (ADR012-R2-01 to -14) plus the shared ADR011-R2-01, -02 and -04.

The digests above identify the reviewed bytes only. No sign-off is given for any file.

**Submission conditions for round 3:**
1. A `round-2-response.md` in each folder, with one row per issue: the fix made, or the recommended option chosen under R-064.
2. **Patch result hashes.** Main applies `ADOPTION.patch` to `ef6f50e8` and records that all six resulting SHA-256 values equal `ADOPTION-SHA256.txt`.
3. **Fresh digests.** Main-computed digests for every signed file, including ADR-000 and ADR-003 `ADR.md` if ADR012-R2-07 is applied as recommended, plus a diff stat against `ef6f50e8`.
4. **Review aid.** The diff from ADR-011's result to ADR-012's canonical files.
5. **New ledger rows.** Rows for the GitHub artifacts, workflow-run, `upload-artifact`, dependency-caching and merge-branch sources, with IDs allocated clear of X1 (next free is R-073 unless X1 adds more).
6. **X1 order.** X1 lands after both ADR revisions (ADR011-R2-01).

---

## Council memo

1. **Recommendation.** Sign off neither ADR this round.
   - **ADR-011:** fix its five text issues. The main one is adding the adoption sequence to its Review: commit A at `ef6f50e8` + patch, commit B with ADR-012, merged without squash, X1 after. ADR-011 can then reach zero issues in round 3.
   - **ADR-012:** adopt the verified dedicated-producer model:
     - default-base PRs only;
     - a single-job caller with literal pins and a matched callee SHA;
     - no PR-writable cache in the producer;
     - artifacts named per attempt and bound to the job window;
     - a caller-file-only gate.

     Then fix revision identity, the provider-state matrix, completion-based finalization, the REQ-22 clause restorations with the ADR-000/003 amendments, and the caller-permission break.
2. **Alternatives.**
   - **(A) Verified dedicated producer (recommended).** No privileged rescanning, works for forks, closes every channel found. Cost: adopters must use a dedicated caller, and the PR lane loses its toolchain cache, which makes runs slower.
   - **(B) Provider attestation (OIDC).** Strong binding, but fork PR runs get no `id-token: write` [INFERENCE], so fork observations would all be unknown. Rejected as the primary mechanism.
   - **(C) Narrow ADR-012 to facts, producer and migration, and defer the collector to a later ADR.** Removes the trust surface now, but CI readiness stays unmeasured, so known gap 4 is not closed. A fallback if (A) does not converge in round 3.
   - **(D) Privileged rescan in the collector.** Rejected again, as in round 1 (REQ-51/161).
   - **(E) Adoption mechanics.** Land ADR-011 then ADR-012 before X1 (recommended), or let X1 land first and regenerate and re-review ADR-011's patch. The second costs a review round.
3. **Evidence.**
   - **Workflows:** `scout-guardian.yml:5-7,42-43,50,58-60,79,82-85,102-105,116-118`; `scout-guardian-trusted.yml:9-10`.
   - **Toolchain:** `bootstrap-tools.sh:73,94,104-118`.
   - **Code:** `diff.go:129-157,221-228`; `export.go:40,693`; `policy.go:300-330,412-416`; `guardrails/suppress.go:49`, `focused_tests.go:9`, `config.go:39`; `catalog_security.go:221-225`.
   - **Design text:** REQ-22/27/51/70/78/161 (`scout-design.html:589,612,618,592,600,707`); cards `:736,807,901,925,991,1001`.
   - **ADRs:** ADR-003 `ADR.md:11,34`; ADR-007/008 `ADR.md:3` compared with base.
   - **Ledgers:** `research-ledger.md:3,72-74`; `scout-x1/docs/research-ledger.md:72-78`.
   - **Patch:** `ADOPTION.patch:8,17,121,125-126`.
   - **External:**
     - [GitHub artifacts REST API](https://docs.github.com/en/rest/actions/artifacts) (no attempt or job on the artifact);
     - [workflow-runs REST API](https://docs.github.com/en/rest/actions/workflow-runs) (`head_sha`, `run_attempt`, `referenced_workflows`);
     - [upload-artifact v4 migration](https://github.com/actions/upload-artifact/blob/main/docs/MIGRATION.md) (same-name conflict, `overwrite`);
     - [toolkit artifact lookup](https://github.com/actions/toolkit/blob/main/packages/artifact/src/internal/find/get-artifact.ts);
     - [dependency caching reference](https://docs.github.com/en/actions/reference/workflows-and-actions/dependency-caching) (PR caches scoped to `refs/pull/<n>/merge`);
     - [pull_request merge branch](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows#how-the-merge-branch-affects-your-workflow);
     - [Actions permissions REST API](https://docs.github.com/en/rest/actions/permissions) (retention settings need Administration access).
4. **Risks.**
   - **Trust:** even with (A), a compromised release repository, tag or default branch defeats the collector. REQ-163 hashing and branch protection remain the root of trust.
   - **Adopter burden:** a dedicated caller and a new `actions: read` grant are breaking changes, and adopters who miss them get `untrusted_producer`, unmeasured collection or a failed trusted run. Release notes are essential.
   - **Fail-safe direction:** an untrusted or unverifiable revision becomes an unknown, which delays activation and never forces it. Hostile PR authors can raise the unknown ratio. Excluding cancellations as `not_assessed` stops push frequency from doing the same.
   - **Unverified GitHub behaviour [INFERENCE]:** fork `pull_requests` being empty, job listings for partial reruns, availability of the compare API for old heads, and runtime-token cache writes. Implementation proofs must exercise each.
   - **Adoption:** if main changes any of the six shared files before ADR-011 lands, the patch must be regenerated and re-reviewed.
   - **Reversibility:** everything is documentation until implemented. After implementation, schema 14 and the v2 manifest are forward-only.
5. **Execution outline.**
   1. Save this record as `round-2.md` in both ADR folders.
   2. **ADR-011:**
      - add the adoption sequence to the Review (R2-01);
      - revise the R-064 caveat and cut R-069 to facts in the patch (R2-02);
      - fix the REQ-22 antecedent (R2-03);
      - add the guardrails gap to REQ-22, STATUS, the plan and the policy note (R2-04);
      - fix the ranges and the README placement (R2-05);
      - regenerate `ADOPTION.patch` and `ADOPTION-SHA256.txt`.
   3. **ADR-012:**
      - rewrite the trust boundary per R2-01/02/03 and update REQ-51/161/162;
      - add the revision identities (R2-04), the provider-state matrix (R2-05) and completion-based finalization (R2-06);
      - restore the REQ-22 clauses and amend ADR-000/003 status lines, README and cards (R2-07);
      - add the caller-permission break and the policy key (R2-08);
      - apply R2-09 to R2-14;
      - mirror the shared ADR-011 fixes byte-identically;
      - add the new GitHub source ledger rows.
   4. Main checks the regenerated patch with `git apply --check` and verifies all six result hashes. It computes digests for every signed file, including ADR-000/003/007/008 `ADR.md`, produces the diff stat and the ADR-011→ADR-012 diff, and confirms with X1 that X1 lands last.
   5. Submit round 3. Each folder is adopted only on its own explicit zero-issue final round, with adoption recorded outside the signed set, in the commit order of ADR011-R2-01.
