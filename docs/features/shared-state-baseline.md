# Shared state baseline

## Sub-features

- **Deterministic export.** `scout export --out <dir>` writes the ledger as `state/v1/*` files in a fixed order.
- **Byte-identical round trip.** `scout import --from <dir>` into a fresh ledger, then `export` again, reproduces the same files byte for byte.
- **Trusted publish only.** `scout state publish` refuses with exit `3` and writes no refs unless `--trusted` is given.
- **Publish and restore.** A trusted publish writes `state/v1/issues.jsonl`, `dispositions.jsonl` and `manifest.json` to the `scout/state` branch and never commits the SQLite file; `scout state restore` loads it back.
- **Import safety.** Import refuses a non-fresh store and a tampered export.

Proved by:

- [state_cmds_test.go](../../cmd/scout/state_cmds_test.go): `TestStatePublishWithoutTrustedRefusesWithExit3`, `TestStatePublishRestoreRoundTripNeverCommitsSQLite`, `TestExportThenImportCommandsReproduceFiles`.
- [export_test.go](../../internal/state/export_test.go): `TestExportImportExportIsByteIdentical`, `TestImportRefusesNonFreshStoreAndTamperedExport`.

## How to get to it (user POV)

1. A trusted default-branch CI run calls `scout state publish --trusted` after its assessment.
2. Pull request runs call `scout state restore` (or let `scout diff` read the baseline) to attribute findings against it.
3. Locally, `scout export --out <dir>` and `scout import --from <dir>` move a ledger between clones.

## Driving it with `go run ./cmd/scout`

```bash
go run ./cmd/scout export --root <fixture-repo> --out <out-dir>
go run ./cmd/scout state publish --root <fixture-repo> --remote <bare-repo> --trusted
```

Outcome tests:

```bash
go test ./cmd/scout -run 'TestStatePublish|TestExportThenImport'
go test ./internal/state -run 'TestExportImportExportIsByteIdentical|TestImportRefusesNonFreshStoreAndTamperedExport'
```

## Gotchas

- Only trusted default-branch runs may publish; pull request and fork runs read the baseline but never write it.
- Publishing uses an expected-parent compare-and-swap, so a concurrent publish fails instead of overwriting.
- Use a local bare repository as `--remote` when driving publish by hand; never publish to a shared remote from a test.
