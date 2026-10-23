# Scout

Discovery-only CI guardian for a single repository. Design: `docs/scout-design.md`; current state and plan: `docs/STATUS.md`; project rules: `AGENTS.md`.

## Development setup

Prerequisites: `bash`, `curl`, `tar`, `git`, `shasum` or `sha256sum`, a Go toolchain (Go then fetches the pinned version from `go.mod`), and the scanner sandbox: `sandbox-exec` (built into macOS) or `bubblewrap` (Linux).

```sh
git submodule update --init          # third_party/clawpatch at the pinned revision
scripts/bootstrap-tools.sh           # pinned, checksum-verified toolchain into .tools/
source .tools/activate               # like a venv; undo with: scout_deactivate
go build ./... && go test ./...
```

`toolchain/versions.env` is the single list of tool versions; `go test ./internal/toolchain` fails if Scout's code and the pins drift apart. Optional account-bound CLIs (`claude`, `codex`, `infisical`) are reported by the bootstrap but not installed.
