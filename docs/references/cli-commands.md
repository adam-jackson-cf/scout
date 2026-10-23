# Scout CLI commands

Use this catalog to find the right `scout` subcommand and its key flags. Run `scout <command> -h` for the full flag list. Source: `cmd/scout/main.go` and the `-h` output of a build of the current tree.

Every command accepts `--root` (repository root, default `.`), `--clawpatch` (absolute path to the pinned Clawpatch `dist/cli.js`; defaults to `SCOUT_CLAWPATCH_BINARY`) and `--json` (machine-readable output).

## Exit codes

| Code | Meaning | Source |
| --- | --- | --- |
| `0` | clean | `cmd/scout/main.go:30-35`, design §7.4 |
| `1` | gated: an activated check found an `introduced` or `regressed` issue | same |
| `2` | incomplete: a required lane did not finish | same |
| `3` | unavailable: bad input, missing tool or refused operation | same |

When more than one applies, the highest code wins: 3 > 2 > 1 > 0. Under the `report` profile, findings never produce `1`.

## Commands

| Command | What it does | Key flags | Evidence source |
| --- | --- | --- | --- |
| `init` | Writes or updates `.scout/policy.yaml` and records consent (profile, Jev decisions, model-CLI lanes). Idempotent; keeps existing activations and comments. | `--yes`, `--profile report\|guardian`, `--decisions --decisions-model <id> --acknowledge-egress`, `--model-cli claude\|codex`, `--lane contextual\|clawpatch-review` | `cmd/scout/init.go` |
| `readiness` | Reports which lanes and tools can run here; `--matrix` prints the capability matrix by language (design §5.3). | `--matrix`, `--scope`, `--path`, `--family` | `cmd/scout/main.go` `readiness` |
| `assess` | Runs a full or scoped assessment and records it in the local ledger. | `--scope repository\|subpath\|staged\|commit_range\|paths`, `--path`, `--exclude`, `--family`, `--from/--to`, `--full`, `--allow-egress`, `--artifact <lcov\|cobertura\|junit>`, `--artifact-revision`, `--artifact-producer`, `--pr-metadata <file>`, `--fetch-pr-metadata --pr-repo owner/name` | `cmd/scout/main.go` `assess` |
| `diff` | Differential run of the checked-out `HEAD` against a base ref, with attribution against the `scout/state` baseline. | `--base <ref>` (required), `--head`, `--baseline-remote`, `--baseline-branch`, `--summary <file.md>`, `--fork`, `--no-egress`, `--override-text`, `--override-actor` | `cmd/scout/diff.go` |
| `issues` | Lists ledger issues. | `--status`, `--family`, `--file`, `--feature`, `--min-severity`, `--run` | `cmd/scout/main.go` `listIssues` |
| `issue` | Shows one issue, optionally with the evidence from one run. | `--id`, `--run` | `cmd/scout/main.go` `showIssue` |
| `triage` | Sets a human disposition on an issue. The change is recorded in an audit trail. | `--id`, `--state investigating\|accepted\|false_positive\|wont_fix`, `--actor`, `--reason`, `--authorize`, `--owner`, `--run`, `--artifact`, `--disposition` | `cmd/scout/main.go` `triage` |
| `coverage` | Shows what a run examined, reused, skipped or blocked, and why. | `--run` (default latest) | `cmd/scout/main.go` `coverage` |
| `mapping` | Shows the feature map; `--areas` lists area labels, `--unowned` lists directory scopes that no heuristic feature owns. | `--areas`, `--unowned` | `cmd/scout/main.go` `mapping`, `cmd/scout/mapping_areas.go` |
| `map` | Evaluates the map refresh triggers for a run, or forces an agent map refresh. | `--refresh`, `--event pr\|push\|nightly`, `--base`, `--head`, `--default-branch`, `--allow-egress` | `cmd/scout/map.go` |
| `signals` | Shows the descriptive measurements (`scout.cc`, `dup_pct`, `cycle_mass`, history measures, …). | `--metric`, `--scope feature:<id>\|<dir>`, `--snapshot`, `--history` | `cmd/scout/signals.go` |
| `export` | Writes the ledger as `state/v1/*` files in deterministic order. | `--out <dir>` | `cmd/scout/state_cmds.go` |
| `import` | Loads `state/v1/*` into the local ledger. Export → import → export round-trips byte for byte. | `--from <dir>` | `cmd/scout/state_cmds.go` |
| `state publish` | Publishes the ledger to the `scout/state` branch with an expected-parent compare-and-swap. Refused unless `--trusted`. | `--remote`, `--branch`, `--trusted` | `cmd/scout/state_cmds.go` |
| `state restore` | Restores the ledger from the `scout/state` branch. | `--remote`, `--branch` | `cmd/scout/state_cmds.go` |
| `serve` | Read-only web view of the ledger, on loopback only. | `--listen` (default `127.0.0.1:8347`) | `cmd/scout/main.go` `serve` |

## Notes

- `scout` with no arguments prints a usage line that lists every command except `map` (`cmd/scout/main.go:89`).
- `diff --head` must be the checked-out `HEAD`; check out the branch before you run it.
- Without a reachable `scout/state` baseline, `diff` reports `comparison: recomputed` and rescans the base itself.
- Local state lives in `.scout/`: `state.sqlite`, `artifacts/`, `clawpatch/` and `policy.yaml`. Commit only `.scout/policy.yaml`; `scout init` says so in its `next` hint. This repository's `.gitignore` excludes all of `.scout/`.
