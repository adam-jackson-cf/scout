# Environment variables

Use this lexicon to see which environment variables Scout and its toolchain read, what each one changes, and which ones are secrets.

## Reader Notes

- These are the Scout-specific variables found in `cmd/` and `internal/` (plus the ones the toolchain sets). Settings that need review belong in `.scout/policy.yaml` instead ([policy-file.md](policy-file.md)).
- `OPENROUTER_API_KEY` is the only secret. Never print, export or commit it. Locally, inject it into the single command that needs it through Infisical, as described in `AGENTS.md` ("Secrets").
- Model CLIs (`claude`, `codex`) use their own subscription login. Scout passes them no API keys and scrubs credentials from their environment (`internal/modelcli`).

## Entries

### Set by `source .tools/activate`

- `SCOUT_CLAWPATCH_BINARY`
  Purpose: absolute path to the pinned, patched Clawpatch `dist/cli.js` under `.tools/clawpatch/<revision>/`. It is the default for every command's `--clawpatch` flag.
  Source: `scripts/bootstrap-tools.sh`; `cmd/scout/main.go` `common`

- `GOTOOLCHAIN`
  Purpose: pins Go to the version in `toolchain/versions.env` (`go1.25.6`).
  Source: `scripts/bootstrap-tools.sh`

- `SCOUT_TOOLS_ACTIVE`, `SCOUT_TOOLS_DIR`
  Purpose: mark the toolchain as active and record its directory (default `.tools/`). Setting `SCOUT_TOOLS_DIR` before bootstrapping installs the toolchain elsewhere, as `release.yml` does. `scout_deactivate` restores the previous `PATH`, `SCOUT_CLAWPATCH_BINARY` and `GOTOOLCHAIN`.
  Source: `scripts/bootstrap-tools.sh`; `.github/workflows/release.yml`

### Model-CLI lanes

- `SCOUT_MODEL_CLI`
  Purpose: `claude` or `codex`: a local choice of CLI for the opted-in `contextual` lane. It applies only when the committed policy leaves `model_cli.provider` unset. With neither set, the first CLI found on `PATH` is used. It never turns a lane on; consent comes from `model_cli.lanes`.
  Source: `cmd/scout/main.go` `providerConfig`

- `SCOUT_MODEL`
  Purpose: optional model name passed to the model CLI. Without it, the CLI uses its own default.
  Source: `cmd/scout/main.go` `providerConfig`

- `CLAWPATCH_PROVIDER`
  Purpose: Clawpatch provider for agent mapping and `clawpatch-review` when none is configured. Only `codex` or `claude` are accepted; any other value is rejected. When it is unset, Scout tries `codex` first, then `claude`, and reports `skipped (no model CLI)` if neither is on `PATH`.
  Source: `internal/clawpatch/provider.go`

- `CLAWPATCH_CODEX_TIMEOUT_MS`
  Purpose: overrides the agent-mapper timeout (default `900000`, 15 minutes).
  Source: `internal/clawpatch/client.go` `mapEnvironment`

- `CLAWPATCH_CLAUDE_AUTH_CONTEXT`
  Purpose: Scout sets it to `host` on every model-using Clawpatch call, so Claude runs on the host login. You don't set it yourself.
  Source: `internal/clawpatch/provider.go`

### Data sources and secrets

- `GITHUB_TOKEN`
  Purpose: read-only token for `scout assess --fetch-pr-metadata --pr-repo owner/name`. It needs `pull-requests: read`. Without PR metadata, `pr_effort` and `batch_size` are `unmeasured`.
  Source: `cmd/scout/main.go` `assess`; `.github/workflows/scout-guardian-trusted.yml`

- `OPENROUTER_API_KEY`
  Purpose: Jev key for the `decisions` lane. It is used only when the policy consents (`decisions.enabled`, `egress_acknowledged`), egress is allowed, and the run is not a fork. In CI only the trusted workflow receives it.
  Source: `cmd/scout/ranking.go`; `.github/workflows/scout-guardian-trusted.yml`; `AGENTS.md`
