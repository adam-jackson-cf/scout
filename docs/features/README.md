# Scout features

One document per top user-facing feature of the `scout` CLI. Each document says what the feature delivers, how a user reaches it, how to drive it from this repository with `go run ./cmd/scout`, and which outcome tests or scenario tests prove it.

This is the product's own feature documentation. It is separate from the feature map Scout builds for the repositories it scans (Clawpatch features and areas, see [lanes.md](../references/lanes.md)).

| Feature                                                 | Commands                                             | Proved by                                                                                                                        |
| ------------------------------------------------------- | ---------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------- |
| [Pull request gate](pull-request-gate.md)               | `diff`                                               | [diff_scenario_test.go](../../cmd/scout/diff_scenario_test.go), [delta_scenario_test.go](../../cmd/scout/delta_scenario_test.go) |
| [Assessment and issue ledger](assessment-and-ledger.md) | `assess`, `issues`, `issue`, `triage`, `serve`       | [assess_identity_test.go](../../cmd/scout/assess_identity_test.go)                                                               |
| [Onboarding policy](onboarding-policy.md)               | `init`                                               | [init_test.go](../../cmd/scout/init_test.go)                                                                                     |
| [Shared state baseline](shared-state-baseline.md)       | `export`, `import`, `state publish`, `state restore` | [state_cmds_test.go](../../cmd/scout/state_cmds_test.go)                                                                         |
| [Code-structure signals](code-structure-signals.md)     | `signals`                                            | [signals_test.go](../../cmd/scout/signals_test.go)                                                                               |

When you add or change user-facing behaviour, update its document here and link the outcome test that proves it. A new feature gets its own document with the same four sections.
