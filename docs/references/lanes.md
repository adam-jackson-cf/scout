# Scout lanes

Use this catalog to see what each lane checks, when it runs, and whether it can ever block a pull request. The lane IDs, triggers and modes come from the registry in `internal/engine/lanes.go`. The intended scope of each lane is in `docs/scout-design.md` §5.1. Where the design and the code differ, the code's current behaviour is noted below.

Modes: **blocking-eligible** lanes can block once a check is activated. **report** lanes record issues but never block. **signal** lanes record measurements only.

| Lane | What it checks | Triggers | Mode | Evidence source |
| --- | --- | --- | --- | --- |
| `guardrails` | Changes against the base that weaken checks: `CONFIG.ignores-added.<tool>`, `CONFIG.strictness-lowered.<tool>` (resolved through `extends`), `SUPPRESS.added.unscoped` (for example bare `# noqa`, `//nolint`, `@ts-nocheck`), `TESTQ.focused-test-added` (`.only`, `fit`). Go, Python, TypeScript. | pr | blocking-eligible | `internal/engine/lanes.go` `LaneGuardrails`; design §5.1 row 1, §5.2 |
| `delta` | `delta.cc-threshold@1` and `delta.dup-added@1` on changed files. Promoted to issues with `claim: descriptive`. `delta.*` cannot be activated. | pr | report | `lanes.go` `LaneDelta`; `docs/STATUS.md` §1 |
| `security` | Gitleaks default credential rules (exact rule ids), plus four bundled Semgrep rules: unsafe `yaml.load` (Python), python-multipart CVE-2024-24762, DOM clobbering via `document.currentScript` (JavaScript), and mutable GitHub Actions refs. | pr, push | blocking-eligible | `lanes.go` `LaneSecurity`; `internal/analyzers/catalog_security.go`, `internal/analyzers/rules/security/` |
| `quality` | Ruff `C90` complexity on Python files. | pr, push, nightly | report | `lanes.go` `LaneQuality`; `internal/analyzers/catalog_quality.go` |
| `architecture` | Planned (design §5.1 row 7: declared boundaries, layers, cycle prohibition and max CC from `.scout/policy.yaml`). Nothing implements it yet: no scanner is registered, so the lane reports `skipped — no eligible inputs`. The policy's `architecture` section is reserved: it is accepted and ignored until the lane is implemented. | pr, push, nightly | report | `lanes.go` `LaneArchitecture`; design §5.1 row 7 |
| `performance` | Ruff `PERF` rules and a bundled CSS Semgrep rule. | push, nightly | report | `lanes.go` `LanePerformance`; `internal/analyzers/catalog_performance.go` |
| `contextual` | LLM discovery through `claude -p` or `codex exec`. Opt-in via `model_cli.lanes`. | nightly | report | `lanes.go` `LaneContextual`; design §5.1 row 9 |
| `clawpatch-review` | Clawpatch review with provider `codex` or `claude` (host auth). Opt-in via `model_cli.lanes`. | nightly | report | `lanes.go` `LaneClawpatchReview`; design §4 |
| `measures` | Cyclomatic complexity and NLOC (lizard), duplication (jscpd), import graph and cycles (Go, Python, TS), and CRAP when a coverage artifact for the same revision is supplied. | pr, push, nightly | signal | `lanes.go` `LaneMeasures`; design §8.2 |
| `history` | Git-mined change cost, coupling, ownership, hotspots and batch size per scope, in calendar-quarter windows. Bot commits are excluded. | push, nightly | signal | `lanes.go` `LaneHistory`; design §8.3 |
| `decisions` | Jev ranking of findings through OpenRouter. Runs only when consented in policy, a key is present, egress is allowed and the run is not a fork. Otherwise ranking is `deterministic (<reason>)`. | per run | signal | `cmd/scout/ranking.go`; design §6 |

## Notes

- Blocking applies only under the `guardian` profile, only for checks listed under `blocking.checks` in the policy, and only to `introduced` or `regressed` findings. Until then, matches show up as `would_block`.
- `guardrails` and `delta` run on `scout diff` only, so a `scout assess` run lists them as `skipped` (`docs/STATUS.md` §1).
- An opt-in lane that cannot run reports `skipped (not opted in)` or `skipped (no model CLI)` and does not change the exit code. If a required lane does not finish, the exit code is `2`.
- A value that could not be measured is `unmeasured` with a reason, never 0. For example, CRAP without a coverage artifact for the same revision, or `pr_effort` without PR metadata.
- The `rubric` lane (design §5.1 row 6) is planned for Phase 5 and is not in the registry.
- `scout readiness --matrix` prints which capabilities are available for each language in the target repository.
