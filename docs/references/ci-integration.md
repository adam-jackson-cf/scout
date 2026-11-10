# CI integration

Use this guide to run Scout in GitHub Actions: pull-request checks, default-branch baselines and nightly runs.

## Purpose

Scout ships two reusable workflows. Each workflow declares its own permission ceiling, so the pull-request lane never holds write access or secrets:

| Piece                                          | Runs on                                               | Does                                                                                                                                                               | Permissions                                                                    | Evidence source                        |
| ---------------------------------------------- | ----------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------ | -------------------------------------- |
| `.github/workflows/scout-guardian.yml`         | `pull_request`                                        | `scout diff` against the PR base, reading the `scout/state` baseline; writes the Markdown summary to the job summary.                                              | `contents: read`; no secrets                                                   | workflow header and job `pull-request` |
| `.github/workflows/scout-guardian-trusted.yml` | push, schedule or dispatch on the default branch only | `state restore` → `map` (onboarding or manual) → `assess --fetch-pr-metadata` → `state publish --trusted` on exit 0 or 1 → uploads evidence artifacts for 90 days. | `contents: write`, `pull-requests: read`; optional `OPENROUTER_API_KEY` secret | job `trusted`                          |
| `.github/workflows/release.yml`                | tag `v*`                                              | Builds `scout-{linux,darwin}-{amd64,arm64}` and the Clawpatch bundle, writes `.sha256` files, and verifies them before creating the release.                       | `contents: write`                                                              | `release.yml`                          |

Every lane downloads the Scout binary and the Clawpatch bundle from a release by tag and checks their sha256 sums. Nothing is built from the PR checkout.

## Preconditions

- A Scout release tag at or after `v0.1.6` exists in the repository that publishes releases (the tag must carry `scripts/install-release.sh` and `scripts/ci-toolchain.sh`; an older tag ends the install step with exit 3). By default that is the repository hosting the called workflow, so a fork or mirror uses its own releases without configuration.
- `.scout/policy.yaml` is committed on the default branch. Run `scout init` locally first; see [policy-file.md](policy-file.md).
- To use Jev ranking, set the `OPENROUTER_API_KEY` repository secret and enable `decisions` in the policy. Only the trusted workflow reads the secret.

## Steps

1. Add a caller workflow. This repository's dogfood caller, `.github/workflows/scout-nightly.yml`, shows the shape:

   ```yaml
   on:
     pull_request:
     push:
       branches: [main]
     schedule:
       - cron: '23 3 * * *'
     workflow_dispatch:
       inputs:
         full: {type: boolean, default: false}
         refresh-map: {type: boolean, default: false}
         authorize-egress: {type: boolean, default: false}

   permissions:
     contents: read

   jobs:
     pull-request:
       if: github.event_name == 'pull_request'
       permissions:
         contents: read
       uses: ./.github/workflows/scout-guardian.yml
       with:
         scout-version: v0.1.6

     trusted:
       if: github.event_name != 'pull_request'
       permissions:
         contents: write
         pull-requests: read
       uses: ./.github/workflows/scout-guardian-trusted.yml
       with:
         scout-version: v0.1.6
         full: ${{ inputs.full || false }}
         refresh-map: ${{ inputs.refresh-map || false }}
         authorize-egress: ${{ inputs.authorize-egress || false }}
       secrets:
         OPENROUTER_API_KEY: ${{ secrets.OPENROUTER_API_KEY }}
   ```

   Another repository references the workflows as `<owner>/<repo>/.github/workflows/<file>@<ref>` instead of `./.github/workflows/<file>`; releases are then fetched from `<owner>/<repo>` unless `release-repo` says otherwise.

2. Merge to the default branch. The first trusted run finds no `scout/state` branch, runs a full assessment, and publishes the first baseline.
3. Open a pull request. The PR lane compares against the baseline. If the baseline is missing or its digests don't match, it rescans the base and reports `comparison: recomputed`.
4. Under the `report` profile, collect `would_block` evidence during the shadow window. Then activate checks under `blocking.checks` and switch to `profile: guardian`.

## Inputs

| Input                | Workflow | Default                                                                 | Effect                                                                                                                                                        |
| -------------------- | -------- | ----------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `scout-version`      | both     | required                                                                | Release tag at or after `v0.1.6`, for example `v0.1.6`.                                                                                                       |
| `release-repo`       | both     | empty → the repository hosting the workflow (`job.workflow_repository`) | `owner/name` that publishes releases.                                                                                                                         |
| `fail-on-incomplete` | both     | `false`                                                                 | Assessment exits 2/3 fail when set. ADR-015 reserved wiring: shadow reconcile 2 is always neutral; shadow 3 consults this flag; triage/start 2/3 always fail. |
| `full`               | trusted  | `false`                                                                 | Force `assess --full`.                                                                                                                                        |
| `refresh-map`        | trusted  | `false`                                                                 | Run `scout map --refresh` before assessing.                                                                                                                   |
| `authorize-egress`   | trusted  | `false`                                                                 | Pass `--allow-egress` to `map` and `assess`.                                                                                                                  |

## What To Check

- The job summary: the ≤ 60-line Markdown summary on PRs, or `status: … (exit N)` on trusted runs.
- Exit handling: assessment 0 passes, 1 fails, and 2/3 warn unless fail-on-incomplete. Under reserved ADR-015 wiring, reconcile 2 is always neutral, 3 warns/fails with that flag; triage/start require publication/idempotent success.
- The evidence artifact: `scout-assessment-<run_id>-pr` (kept 30 days) holds the assessment JSON, its stderr and the summary. `scout-assessment-<run_id>-<event>` (kept 90 days) holds the trusted run's JSON, stderr and `scout-map.json`.
- The `scout/state` branch advances after each trusted run that exits 0 or 1.
- On GitHub-hosted runners, the model-CLI lanes and the agent map report `skipped (no model CLI)`. That is expected and does not change the exit code.

## Related Docs

- `README.md`: local setup and first run.
- [cli-commands.md](cli-commands.md): the `diff`, `assess`, `map` and `state` flags that the workflows use.

## Planned readiness evidence (ADR-012; implementation pending)

[ADR-012](../adrs/2026-11-08-shadow-readiness-inputs/ADR.md) specifies evidence, not a new workflow. Trigger/caller/transport/retention/sandbox wiring and writer commands belong to the implementation PR. Existing workflows, grants and installation contract remain unchanged here. Any new/changed trigger, dispatch input, reusable workflow, job permission, policy key, release artifact or REQ-160–163/168 text is a design change requiring its own adopted wiring ADR and zero-issue council review before the implementation PR merges, meeting ADR-012 invariants; wiring within existing requirements needs no additional ADR.

[ADR-015 W1–W7](../adrs/2026-11-10-shadow-readiness-wiring/ADR.md), effective on its explicit final zero-issue round, adopts read-only diagnostic inventory/credential hardening. V1 writes nothing, performs no reassessment, Complete=false and C unchanged. No certifying source/event set exists; source/reassessment, triage/start transport/control storage and certifying S6 cutover are explicitly deferred to a later ADR and its source/identity/B/time/capacity/authority/provider/isolation/operational proofs. ADR-012 readiness remains incomplete until then; preserve existing/imported vetoes, triage and lineage. Round-3 envelope controls are a known viable option only, not adopted.

**Breaking upgrade:** every trusted caller upgrading to the wiring-enabled reusable revision must grant actions:read alongside contents:write/pull-requests:read or startup fails. Pair it with a release shipping read-only reconcile; implementation records the exact minimum tag. Old binaries report unavailable/3 after assessment, a warning unless fail-on-incomplete. All existing inputs and assessment/map/publish/artifact/enforcement guards remain unchanged. No normalization/action step is adopted. Release artifacts/scripts remain unchanged.

Set trusted checkout persist-credentials:false. Restore/publish create fresh per-step token-free askpass helpers with mktemp and an exit cleanup trap; no credential argv/config/files or shared helper. Inventory gets GITHUB_TOKEN only, no Git token. Restore fails closed on ls-remote status other than present 0/absent 2. Before workflow implementation merges, Linux/macOS S6/P23 must prove own/sibling-process confidentiality and integrity of runner command files, helper, installed Scout/Clawpatch and toolchain bytes. Inventory goes last after upload/enforcement, with 90-second binary deadline, 2-minute backstop, 100 requests and rate floor 200. Stop without waiting/retry on 403/429/retry-after, missing/malformed rate headers or insufficient budget. Step continue-on-error is the inverse of fail-on-incomplete; whole-job timeout can still cancel after assessment reporting and is unavailable inventory, never coverage.

No marker/horizon, attempt/event deadline or finite lag is selected. Inventory is newest-first by original created_at then descending numeric IDs. A later adopted ADR—not a clock or larger lag—must authorize complete event/attempt collection, reassessment, authenticated triage/start and all-writer/rerun cutover.

- Apply ADR-012/REQ-78’s population, event binding, provider-job-start, causal-supersession and single query-wide visibility/completeness rules. Unknown causes, manual cancel followed by later push, skip instructions and approval waiting are assessed/gapped, never presumed exclusions. Supersession exclusion requires an identified superseding run of the pinned guardian lane for the same PR key, created (run created_at at its first attempt, never rerun-attempt time; a rerun attempt of an earlier-created run never qualifies) after the cancelled attempt’s job started and before that attempt’s cancellation, whose executed configuration verifiably equals the pinned protected-default configuration and evaluates cancel-in-progress=true for the same evaluated concurrency group as the cancelled run. Additionally require either a provider-recorded concurrency cause or no actor-initiated cancel recorded; the provider cause replaces only that last condition, never the run/configuration checks. If no qualifying run can be identified within both bounds, including when original run creation cannot be distinguished from rerun-attempt time, assess or gap the head, never exclude it; missing proof cannot certify readiness. Cancellation by another workflow or job, an unidentified canceller, or a PR-edited superseding run is assessed/gapped. Deadline-expired non-terminal jobs receive interval-ending point gaps; later terminal facts follow the late lifecycle. Destination C at publication is the CAS-parent frontier, recomputed on retry; the same commit’s advance never makes its own facts/points late. Each gap point independently is late iff point≤that parent C. REQ-78’s sole visibility predicate requires coordinate≤as_of and stamp≤as_of only for late facts/points. Only points in (parent C, new C] require published normal gaps visible at the enabled frontier; late points≤parent C, including equality, never block advancement; earlier dossiers stay identical. Missing/unverifiable accounting is a gap or evidence refusal/exit 2, never presumed completeness. Compare pinned guardian merge result versus B with exactly its override inputs (currently none).
- Only trusted protected-default-branch writers append readiness facts/events/triage, not PR/fork jobs. The authenticated original and executing triage authors must match and still be maintainers under policy/CODEOWNERS at the current protected default tip when publishing. Reject non-default/unreachable writer source, revoked or cross-actor replay. Stable repository-qualified logical action IDs make replay a no-op. Existing scout/state ruleset protection against write-capable collaborators remains essential.
- Untrusted scans cannot modify trusted authorization, recipe, other attempts or files later executed by trusted steps. Prove this integrity boundary; a compromised scanner's own-attempt event/status falsification remains an accepted residual risk. This document does not prescribe a sandbox/transport.
- Before enabling collection, every shared-state reader/writer must support schema 14/manifest v2, with no old in-flight/rerun writer. On every CAS retry, refuse downgraded hybrid v1: recover last intact v2 and reapply subsequent state. Refuse genuinely newer schemas: upgrade writer, never roll back the ledger. Incomplete cutover history stays unmeasured.

The implementation must demonstrate completeness, guardian parity, stable event identity, authorized/idempotent triage, publication-safe historical cutoffs, storage compatibility and trusted-writer integrity before claiming readiness implemented. Only current_window with live certification=current supplies activation evidence; REQ-78 scopes refusal and binds identical-prefix restores/cross-lineage Undecided imports. Prove all [ADR-012 acceptance invariants](../adrs/2026-11-08-shadow-readiness-inputs/ADR.md#implementation-acceptance-invariants), not just a refusal. Required fixtures: clearing last Undecided refuses/exit 2 until C reaches stamp, then ready/exit 0; frontier-only quiet reconcile computes; another-check post-C triage does not refuse; late post-C record while an in-flight attempt holds C gives refused/exit 2, while --as-of C has the same digest/historical status. Day25/day32 cross-lineage verdicts in both orders remain Undecided until an authorized destination verdict after the last import publication, which restores ready; multi-row imports keep one row/event/publication, repeated import is a no-op, sealed local fast-forward equals fresh restore. At historical day30, independently assert that a same-pin/different-inspection late rerun preserves the representative, and a different-pin late rerun preserves the default version; cohort/digest stay identical before pending import, after pending import and after publication. A day20 point ends the interval so day35 interval/Days origin starts after it. Manual cancel then later push is assessed/gapped; approval-waiting fork cannot hold C, and deadline-expired job is point-gapped, then C advances past started_at before terminal publication, whose facts must be late/invisible before stamp. With C≥terminal stamp and no refusal trigger, assert the late terminal attempt and its linked would_block event are visible, yielding computed exit 0 not-ready/Undecided, paired with an otherwise identical ready control without the late terminal facts. Late would_block, imported not_warranted and skewed-clock correction veto fixtures advance C≥all stamps with no refusal trigger, assert computed exit 0 not-ready with the named Undecided/not_warranted reason, and pair otherwise identical ready controls without the late event/import/correction. Retain every earlier fixture in ADR-012, including numeric attempt order, distinct-cutoff digests and Scout/tool/catalogue/policy pin-mixture disclosure. Round-9 fixtures: parent C<point<new C<stamp selects the interval strictly after the point and ends the previous interval strictly before it; parent C<point=new C<stamp has no complete containing interval (unmeasured/exit 2), then a later cutoff after the point selects the next interval; a rerun deadline gap with event_at≤parent C preserves pre-stamp dossiers and activates its late point at C≥stamp. Diverged-prefix destination post-C correction and fork-restored chain follow cross-lineage rules/current-tip authorization. Rows only for E′ leave resolved E resolved; verdict between two import publications cannot resolve E after the second. Revoked-author import preserves C dossier, refuses until refusal stamp≤C, then computes exit 0 not-ready/Undecided; new destination verdict restores ready, replay is no-op. With pinned cancel-in-progress=false, H2 before H1 manual cancel/no recorded canceller still assesses/gaps H1. Pending local late facts and pending imported history separately refuse then clear after trusted publication/C advance. Unrelated-check late record refuses this cohort and clears once C≥stamp. Wiring choices remain in the implementation PR, subject to REQ-161. One commit publishes new C, a terminal attempt with parent C<started_at≤new C<stamp and a gap with parent C<point<new C: attempt normal/visible, boundary visible, certification=current. A late point equal to parent C, previously accounted without a gap, permits advance to its stamp and preserves pre-stamp dossiers. R1=true, R2 executed configuration=false (PR edit or base change), then manual cancel/no recorded canceller still assesses/gaps H1; unverifiable R2 configuration also assesses/gaps H1. Keep the all-false fixture. PR-added workflow job joins guardian group with cancel-in-progress=true, cancels H1 before assessment with provider cause and no H2: H1 is assessed/gapped. With pinned=false, H2 PR-edits R2 executed configuration to true in the same group and cancels R1 before assessment: H1 is assessed/gapped both with no recorded canceller and, separately, with provider-recorded concurrency cause. Retain the existing all-false, R1-true/R2-false and unverifiable-R2 fixtures. Supersession-window fixtures (all other exclusion conditions satisfied): (1) pinned=true, unedited configuration, no actor recorded/no provider cause; R1 is cancelled before assessment and H2’s R2 is created only after that cancellation: assess/gap H1. (2) Pinned=true; R2 legitimately supersedes R1, then a PR-added job cancels R2 before assessment with provider cause and no H3: assess/gap H2, never accept older R1 as its superseder. (3) Repeat (2) with R2 manually cancelled, no actor recorded and no provider cause: assess/gap H2. These isolate the upper bound and both lower-bound evidence paths; keep every existing fixture. (4) Pinned=true, unedited configuration; R1 on H1 is terminal; R2 on H2 has started; a maintainer reruns R1 and the rerun attempt cancels R2 before assessment; no H3. Once with provider concurrency cause and once with no actor recorded/no provider cause: assess/gap H2, never accept R1’s rerun attempt as its superseder.
