<div align="center">

<h1>Scout</h1>

**A discovery-only CI guardian for a single repository: it finds evidence-backed issues and measures code structure over time, and blocks a pull request only for checks you have switched on.**

![Status](https://img.shields.io/badge/status-pre--release-orange.svg?style=flat-square)
![Go](https://img.shields.io/badge/go-1.25.6-00ADD8.svg?style=flat-square)
![Primary CLI](https://img.shields.io/badge/cli-scout-blue.svg?style=flat-square)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Linux-lightgrey.svg?style=flat-square)

</div>

## Quick Install

Prerequisites:

- `bash`, `curl`, `tar`, `git`, and `shasum` or `sha256sum`
- a Go toolchain (Go then fetches the version pinned in `go.mod`)
- the scanner sandbox: `sandbox-exec` (built into macOS) or `bubblewrap` (Linux)

Bootstrap the pinned toolchain and build:

```bash
git submodule update --init          # third_party/clawpatch at the pinned revision
scripts/bootstrap-tools.sh           # pinned, checksum-verified tools into .tools/
source .tools/activate               # like a venv; undo with: scout_deactivate
go build ./... && go test ./...
go install ./cmd/scout               # puts scout in $(go env GOPATH)/bin; add that to PATH
```

`scripts/bootstrap-tools.sh` installs uv, a Python 3.11 venv with Semgrep, Ruff and lizard, Node with pnpm and jscpd, Gitleaks, and the patched Clawpatch bundle. Activation puts them on `PATH` and sets `SCOUT_CLAWPATCH_BINARY`, so you don't need to pass `--clawpatch`. `claude`, `codex` and `infisical` are optional: the bootstrap reports whether they are present but does not install them.

## Quick Start

Inside any git repository, with the toolchain active:

```bash
scout init --yes                    # writes .scout/policy.yaml (profile: report)
scout assess                        # full assessment into the local ledger (.scout/state.sqlite)
scout issues                        # list issues in the ledger
scout diff --base main --summary scout-summary.md   # what this branch introduced
```

`assess` and `diff` print a JSON result with `gate`, `health`, `would_block` and `ranking`, and exit `0` clean, `1` gated, `2` incomplete or `3` unavailable. Under the default `report` profile, findings never fail the run. A branch that adds an unscoped `# noqa` gets a summary listing `guardrails.SUPPRESS.added.unscoped` as *introduced*, under "Would block if activated", plus a lane-health table. `scout serve` starts a read-only web view on `127.0.0.1:8347`.

## What Scout Does

Scout runs on pull requests, default-branch pushes, nightly runs and locally. It records what it finds in a durable ledger that you can triage. It does not fix code. It only blocks a pull request when the pull request itself introduced or re-introduced a finding from an allow-listed check that the repository has activated (`docs/scout-design.md` §1).

- `cmd/scout`: the CLI (`init`, `assess`, `diff`, `issues`, `triage`, `signals`, `state`, `serve`, …).
- `internal/engine`: runs the lanes, compares base and head, applies the gate and profiles.
- `internal/analyzers`, `internal/guardrails`, `internal/measures`, `internal/history`, `internal/artifacts`: the checks and measurements.
- `internal/clawpatch`, `internal/areas`: feature mapping through the pinned Clawpatch submodule.
- `internal/issues`, `internal/state`: stable issue identity, the SQLite ledger, and the `state/v1` export and import.
- `internal/decisions`: optional Jev ranking through OpenRouter, with a deterministic fallback.
- `.github/workflows/scout-guardian*.yml`, `.github/actions/scout/`: reusable CI that installs a pinned, checksum-verified release.
- `evaluation/`: product experiments and their review rounds; `docs/`: design, status and ledgers.

## Core Concepts

- **Issue**: a concrete finding with sanitised evidence at a source location. It has a durable ID (`issue:<family>:<ulid>`) that survives line shifts and renames. Only issues can gate.
- **Signal**: a descriptive measurement, such as cyclomatic complexity, duplication or import cycles, for a scope at a snapshot. Signals never gate, and a value that could not be measured is reported as `unmeasured`, never as 0.
- **Lane**: one family of checks (guardrails, security, measures, history, …). Each run reports a health status for every lane: ran, reused, skipped, blocked or failed.
- **Attribution**: `introduced`, `regressed`, `pre-existing` or `unknown` relative to the base. Only `introduced` and `regressed` can block.
- **Profile**: `report`, the default, never fails CI on findings. `guardian` blocks on the checks activated under `blocking` in `.scout/policy.yaml`.
- **would_block**: a finding that would block if its check were activated. Shadow runs use it to collect evidence before activation.
- **Disposition**: a human triage state set with `scout triage` (`investigating`, `accepted`, `false_positive`, `wont_fix`). Dispositions carry over when the issue matches on later runs.
- **Baseline**: the `scout/state` git branch. Trusted default-branch runs publish to it, and PR runs read it.
- **Feature map / areas**: the heuristic Clawpatch map is the canonical one. Agent-mapped areas are labels only; they never decide identity.

## Go Deeper

- [docs/references/cli-commands.md](docs/references/cli-commands.md): every `scout` command, its key flags, and the exit codes.
- [docs/references/lanes.md](docs/references/lanes.md): what each lane checks, when it runs, and whether it can block.
- [docs/references/policy-file.md](docs/references/policy-file.md): the keys in `.scout/policy.yaml` and what `scout init` writes.
- [docs/references/ci-integration.md](docs/references/ci-integration.md): wiring the reusable guardian workflows into a repository.
- [docs/references/environment-variables.md](docs/references/environment-variables.md): the environment variables that Scout and its toolchain read.
- `docs/scout-design.md`: the authoritative design (revision 2.2). `docs/STATUS.md`: current state and remaining work. `AGENTS.md`: project rules, including council sign-off for any design change or experiment.
