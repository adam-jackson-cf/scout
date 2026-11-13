# Assessment and issue ledger

## Sub-features

- **Full assessment.** `scout assess` runs every enabled lane over the repository and records the run, its lane health and its issues in the local ledger (`.scout/state.sqlite`). A run where opt-in lanes are blocked is recorded as `partial`, not failed.
- **Stable issue identity.** An issue keeps its opaque ID (`issue:<family>:<ulid>`) when its file is renamed or lines shift above it. A secret keeps its issue across moves by digest.
- **Evidence drift.** When a finding's message changes but it still matches (for example a C901 complexity number), the issue records `evidence_drift` and keeps its human disposition.
- **Listing and inspection.** `scout issues` lists ledger issues, filterable by status, family, file and feature; `scout issue --id` shows one issue and its evidence.
- **Triage.** `scout triage` sets `investigating`, `accepted`, `false_positive` or `wont_fix`; mutations without `--authorize` are refused.
- **Read-only web view.** `scout serve` serves the ledger on loopback only, rejects unsafe requests and redacts state output.

Proved by:

- [assess_identity_test.go](../../cmd/scout/assess_identity_test.go) `TestAssessKeepsIdentityAcrossRenameAfterPartialRun`: partial run, rename and evidence drift through the real assess path.
- [secret_e2e_test.go](../../internal/issues/secret_e2e_test.go) `TestGitleaksSecretKeepsIssueAcrossMoveEndToEnd`: secret identity across a move.
- [main_test.go](../../cmd/scout/main_test.go) `TestRequireAuthorizationRejectsUnauditedMutation`: triage authorization.
- [web_test.go](../../internal/web/web_test.go) `TestHandlerRejectsUnsafeRequestsAndRedactsStateOutput`: the web view's safety.

## How to get to it (user POV)

1. Inside any git repository with the toolchain active, run `scout init --yes` once.
2. Run `scout assess` for a full assessment.
3. Run `scout issues` to list what it found, and `scout issue --id <id>` for one issue.
4. Run `scout triage --id <id> --state accepted --actor <you> --reason <why> --authorize` to record a decision.
5. Run `scout serve` and open `http://127.0.0.1:8347` for the read-only view.

## Driving it with `go run ./cmd/scout`

```bash
go run ./cmd/scout assess --root <fixture-repo> --json
go run ./cmd/scout issues --root <fixture-repo> --json
```

The outcome tests run against fixture repositories in `t.TempDir()`:

```bash
go test ./cmd/scout -run 'TestAssessKeepsIdentityAcrossRenameAfterPartialRun|TestRequireAuthorizationRejectsUnauditedMutation'
go test ./internal/issues -run TestGitleaksSecretKeepsIssueAcrossMoveEndToEnd
go test ./internal/web
```

## Gotchas

- `TestAssessKeepsIdentityAcrossRenameAfterPartialRun` skips unless `ruff` 0.15.8 is on `PATH`, and skips if the pinned Ruff emits no C901 or PERF finding for its fixture.
- Assessment needs the pinned Clawpatch bundle (`SCOUT_CLAWPATCH_BINARY`, set by `.tools/activate`) and the scanner sandbox.
- Commit only `.scout/policy.yaml`; the ledger and artifacts under `.scout/` are local state.
