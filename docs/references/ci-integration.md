# CI integration

Use this guide to run Scout in GitHub Actions: pull-request checks, default-branch baselines and nightly runs.

## Purpose

Scout ships two reusable workflows and one composite action. Each workflow declares its own permission ceiling, so the pull-request lane never holds write access or secrets:

| Piece | Runs on | Does | Permissions | Evidence source |
| --- | --- | --- | --- | --- |
| `.github/workflows/scout-guardian.yml` | `pull_request` | `scout diff` against the PR base, reading the `scout/state` baseline; writes the Markdown summary to the job summary. | `contents: read`; no secrets | workflow header and job `pull-request` |
| `.github/workflows/scout-guardian-trusted.yml` | push, schedule or dispatch on the default branch only | `state restore` → `map` (onboarding or manual) → `assess --fetch-pr-metadata` → `state publish --trusted` on exit 0 or 1 → uploads evidence artifacts for 90 days. | `contents: write`, `pull-requests: read`; optional `OPENROUTER_API_KEY` secret | job `trusted` |
| `.github/actions/scout/` | any job | Installs a pinned release and runs `diff` (`mode: pr`) or `assess` (`mode: trusted`). | caller's | `action.yml` |
| `.github/workflows/release.yml` | tag `v*` | Builds `scout-{linux,darwin}-{amd64,arm64}` and the Clawpatch bundle, writes `.sha256` files, and verifies them before creating the release. | `contents: write` | `release.yml` |

Every lane downloads the Scout binary and the Clawpatch bundle from a release by tag and checks their sha256 sums. Nothing is built from the PR checkout.

## Preconditions

- A Scout release tag exists in `release-repo` (default `adam-jackson-cf/scout`).
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
         scout-version: v0.1.0
         release-repo: ${{ github.repository }}

     trusted:
       if: github.event_name != 'pull_request'
       permissions:
         contents: write
         pull-requests: read
       uses: ./.github/workflows/scout-guardian-trusted.yml
       with:
         scout-version: v0.1.0
         release-repo: ${{ github.repository }}
         full: ${{ inputs.full || false }}
         refresh-map: ${{ inputs.refresh-map || false }}
         authorize-egress: ${{ inputs.authorize-egress || false }}
       secrets:
         OPENROUTER_API_KEY: ${{ secrets.OPENROUTER_API_KEY }}
   ```

   Another repository references the workflows as `<owner>/scout/.github/workflows/<file>@<ref>` instead of `./.github/workflows/<file>`.
2. Merge to the default branch. The first trusted run finds no `scout/state` branch, runs a full assessment, and publishes the first baseline.
3. Open a pull request. The PR lane compares against the baseline. If the baseline is missing or its digests don't match, it rescans the base and reports `comparison: recomputed`.
4. Under the `report` profile, collect `would_block` evidence during the shadow window. Then activate checks under `blocking.checks` and switch to `profile: guardian`.

## Inputs

| Input | Workflow | Default | Effect |
| --- | --- | --- | --- |
| `scout-version` | both | required | Release tag, for example `v0.1.0`. |
| `release-repo` | both | `adam-jackson-cf/scout` | `owner/name` that publishes releases. |
| `fail-on-incomplete` | both | `false` | Fail the job on exit 2 or 3 instead of a neutral warning. |
| `full` | trusted | `false` | Force `assess --full`. |
| `refresh-map` | trusted | `false` | Run `scout map --refresh` before assessing. |
| `authorize-egress` | trusted | `false` | Pass `--allow-egress` to `map` and `assess`. |

## What To Check

- The job summary: the ≤ 60-line Markdown summary on PRs, or `status: … (exit N)` on trusted runs.
- Exit handling: `0` passes, `1` fails the job, and `2`/`3` give a neutral warning unless `fail-on-incomplete` is set.
- The evidence artifact: `scout-assessment-<run_id>-pr` (kept 30 days) holds the assessment JSON, its stderr, the summary and SARIF if one was produced. `scout-assessment-<run_id>-<event>` (kept 90 days) holds the trusted run's JSON, stderr, `scout-map.json` and SARIF.
- The `scout/state` branch advances after each trusted run that exits 0 or 1.
- On GitHub-hosted runners, the model-CLI lanes and the agent map report `skipped (no model CLI)`. That is expected and does not change the exit code.

## Related Docs

- `README.md`: local setup and first run.
- [cli-commands.md](cli-commands.md): the `diff`, `assess`, `map` and `state` flags that the workflows use.
