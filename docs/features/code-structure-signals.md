# Code-structure signals

## Sub-features

- **Descriptive measurements.** `scout signals` shows recorded measurements such as `scout.cc`, `dup_pct` and `cycle_mass` for a scope. Signals never gate.
- **Scoped view.** `--scope <dir>` or `--scope feature:<id>` limits output to one scope; other scopes are left out.
- **Latest value with its parts.** Each signal shows its numerator, denominator and population digest for the latest snapshot.
- **Unmeasured is explicit.** A measurement that could not be taken (for example `scout.crap` without a coverage artifact) is reported as `unmeasured` with its reason, never as `0`.

Proved by:

- [signals_test.go](../../cmd/scout/signals_test.go) `TestSignalsCommandShowsNumeratorDenominatorAndPopulationForAScope`.

## How to get to it (user POV)

1. Run `scout assess` so the ledger holds measurements.
2. Run `scout signals --scope <dir>`; add `--metric <id>` for one measurement or `--history` for its trend.

## Driving it with `go run ./cmd/scout`

```bash
go run ./cmd/scout signals --root <fixture-repo> --scope pkg/sub --json
```

Outcome test:

```bash
go test ./cmd/scout -run TestSignalsCommandShowsNumeratorDenominatorAndPopulationForAScope
```

## Gotchas

- Signals describe; they never block. A threshold that should block belongs in an activated check, not a signal.
- A missing value means "not measured", not zero. Check `missing` and its reason before comparing snapshots.
