# Pull request gate

## Sub-features

- **Differential run.** `scout diff --base <ref>` compares the checked-out `HEAD` against the base and prints a JSON outcome with `gate`, `health`, `would_block`, `ranking` and `issues`. Each issue carries an attribution: `introduced`, `regressed`, `pre-existing` or `unknown`.
- **Report profile never fails.** Under the default `report` profile the run exits `0` and lists findings that would block under `would_block`, persisted in the ledger.
- **Guardian profile blocks only activated checks.** When the base policy activates a check, an `introduced` issue from that check sets `gate.status` to `gated`, lists the issue in `gate.blocking`, and exits `1`. Findings from checks that are not activated stay in `would_block`.
- **Rule version bump.** A policy that pins an older check version does not block on the current version.
- **Maintainer override.** `--override-actor` with `--override-text "scout:allow <check> reason=… ref=…"` is accepted from a listed maintainer and recorded as a disposition; a non-maintainer's override is rejected and logged.
- **Descriptive deltas stay report-only.** A pull request that raises a function's cyclomatic complexity past the threshold yields one `delta.cc-threshold@1` issue with claim `descriptive`; it never blocks and never changes the exit code. A docs-only change yields no delta.
- **Guardrail erosion.** A change that adds an unscoped suppression (for example a bare `# noqa`) is reported as an introduced `guardrails` issue.
- **Published baseline reuse.** With a reachable `scout/state` baseline whose digests match, `diff` reuses it; otherwise it reports `comparison: recomputed` and rescans the base.
- **Markdown summary.** `--summary <file.md>` writes the pull request summary with the "Would block if activated" list and the lane-health table.

Proved by:

- [diff_scenario_test.go](../../cmd/scout/diff_scenario_test.go) `TestDiffGateScenario`: report profile, guardian blocking, rule version bump and overrides on a real git fixture.
- [delta_scenario_test.go](../../cmd/scout/delta_scenario_test.go) `TestDiffDeltaIsReportOnlyUnderGuardianProfile`: descriptive deltas.
- [guardrails_diff_test.go](../../cmd/scout/guardrails_diff_test.go) `TestDiffReportsGuardrailErosion`: guardrail erosion.
- [diff_baseline_test.go](../../cmd/scout/diff_baseline_test.go) `TestDiffUsesPublishedBaselineUnlessDigestsDiffer`: baseline reuse.

## How to get to it (user POV)

1. In a repository with Scout initialised (`scout init --yes`), check out the pull request branch.
2. Run `scout diff --base main --summary scout-summary.md`.
3. Read the JSON on stdout and the Markdown summary. In CI, the reusable guardian workflow runs the same command on every pull request; see [ci-integration.md](../references/ci-integration.md).
4. To block on a check, activate it under `blocking` in `.scout/policy.yaml` and set `profile: guardian`; see [policy-file.md](../references/policy-file.md).

## Driving it with `go run ./cmd/scout`

From the repository root, with the toolchain active (`source .tools/activate`):

```bash
go run ./cmd/scout diff --root <fixture-repo> --base <base-ref> --json
```

Use a throwaway git repository as `--root`; run state lands in its `.scout/` directory. The scenario tests build such fixtures with `git init` in `t.TempDir()` and call the CLI entry point in process:

```bash
go test ./cmd/scout -run 'TestDiffGateScenario|TestDiffDeltaIsReportOnlyUnderGuardianProfile|TestDiffReportsGuardrailErosion|TestDiffUsesPublishedBaselineUnlessDigestsDiffer'
```

## Gotchas

- The scenario tests skip, not fail, when `ruff`, `gitleaks`, `semgrep`, `lizard` or `jscpd` are missing from `PATH`. Source `.tools/activate` first, or a green run proves nothing.
- `--head` must be the checked-out `HEAD`; check out the branch before running `diff`.
- The gate reads committed policy, never an uncommitted working-tree copy ([policy_sources_test.go](../../cmd/scout/policy_sources_test.go)). Commit `.scout/policy.yaml` before expecting an activation to block.
- Exit codes combine: `3` (unavailable) beats `2` (incomplete) beats `1` (gated) beats `0`. See [cli-commands.md](../references/cli-commands.md#exit-codes).
