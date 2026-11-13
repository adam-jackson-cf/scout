# Contributing

## Install pinned tools and native hooks

Install a Go bootstrap toolchain first (for example, `mise use go@1.26.11`).
The installer fails clearly if Go is unavailable; Go's toolchain downloader
then selects the version in `toolchain/versions.env`.

```sh
git submodule update --init
scripts/bootstrap-tools.sh
source .tools/activate
git config core.hooksPath .githooks
```

The last command installs this checkout's tracked native hooks. It is an explicit
local setup step, not a side effect of running checks. No hook bypass is supported.
Tool pins live only in `toolchain/versions.env`; when changing an allowed developer
tool pin, update its package metadata and run `scripts/bootstrap-tools.sh --relock`.
Scanner and Clawpatch upgrades require their compatibility review.

Build and check the design digest with the toolchain activated:

```sh
go build ./... && go vet ./... && go test -count=1 ./...
shasum -a 256 docs/scout-design.html
```

The digest must match the SHA-256 of `docs/scout-design.html` signed in the zero-issue final round of the latest ADR that changed the design (now `docs/adrs/2026-11-06-design-restructure/`).
For what to work on, see `docs/plans/2026-11-06-remaining-work.md`.

## Accounts, external tools and fixture data

Accounts and external tools are not stored in the repository:

- Log in to the `claude` and `codex` CLIs with their subscriptions. The fallback
  lane and post-run scripts use `codex exec -m gpt-6.1-terra` (effort high);
  labelling uses both CLIs.
- Log in to the Infisical CLI. Inject the Jev key (`OPENROUTER_API_KEY`) only
  into the command that needs it:
  `infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- <command>`.
  Never export or print the key.
- Do not copy `.codegraph/` or the ignored contents of `.scout/` (`state.db`
  and other local state). `.scout/policy.yaml` is tracked and arrives with the
  clone. Scout's published state is on `origin/scout/state`.

Replay generators and study scripts need fixture clones:

```sh
mkdir -p /tmp/scout-x2c-smoke/_src && cd /tmp/scout-x2c-smoke/_src
git clone https://github.com/fastapi/fastapi.git
git clone https://github.com/tokio-rs/axum.git && git -C axum checkout --detach bebe175c334d2a4154b32cd9e477274c2aab5953
```

Every stage verifies pinned SHAs, not paths, so any clone holding them works
(pass `--fastapi`/`--axum` or `X2D_SRC`). The X2d pilot and post-run data is
committed under `evaluation/x2d-merge-association-pilot/data/`. From the repository
root, regenerate the parity fixtures with:

```sh
python3 internal/association/testdata/gen/gen_parity.py \
    --fastapi /tmp/scout-x2c-smoke/_src/fastapi --axum /tmp/scout-x2c-smoke/_src/axum
```

`gen_parity.py` defaults `--x2d` to that in-repo data path and `--fastapi`/`--axum`
to the clones above. The post-run scripts in `evaluation/x2d-postrun-tuning/`
(`run_window.sh`, `run_jev.sh`, `run_partB.sh`) are frozen; their digests are in
`FROZEN-8-10.md`. Their old-machine defaults remain: `X2D_WORKTREE` is
`/Users/adamjackson/Projects/inflight/scout-x2d`, and `X2D_SRC` (fastapi only) is
`/tmp/scout-x2c-smoke/_src/fastapi`. Override both with environment variables.
To recreate the worktree and supply its data, run from the repository root:

```sh
git worktree add --detach ../scout-x2d 8c7148c
cp -R evaluation/x2d-merge-association-pilot/data ../scout-x2d/evaluation/x2d-merge-association-pilot/
export X2D_WORKTREE=$PWD/../scout-x2d
export X2D_SRC=/tmp/scout-x2c-smoke/_src/fastapi
(cd ../scout-x2d && git submodule update --init && scripts/bootstrap-tools.sh)
```

The pilot's files at `8c7148c` are frozen (experiment rule 5); do not edit them.

## One gate locally and in CI

```sh
bash scripts/gate.sh ci
bash scripts/gate.sh commit
bash scripts/gate.sh path internal/domain/types.go
```

The pre-commit hook runs `bash scripts/gate.sh commit`. CI's required `gate` job
runs `bash scripts/gate.sh ci` on pushes and pull requests, without modifying
sources. Failures propagate. CI reads the Go pin from `toolchain/versions.env`
before setup-go and runs `bash scripts/ci-sandbox.sh --strict` before bootstrap.
The shared script installs Linux bubblewrap and handles the Ubuntu user-namespace
restriction; an unusable sandbox blocks the gate immediately. Guardian
provisioning calls the same script without `--strict`, preserving its
warning-only behavior when the probe fails.

The gate's checks live in `scripts/gate/`:

- `run.py` runs the modes. The `path` mode checks file placement
  (`owned-path`, `test-placement`) and package dependency direction
  (`layering`) against `scripts/gate/structure.json`. `imports.go` reads each
  Go file's imports for the `layering` check.
- `tests.py` runs the Go test tiers declared in `scripts/gate/test-tiers.json`:
  `fast` packages on every commit and in CI, `seam` packages in CI and on commits
  that touch them. Every package with tests must declare a tier. The fast tier
  enforces the declared coverage floor.
- `python.py` runs pinned Ruff (lint and format check) over maintained Python,
  and mypy in CI.
- `infrastructure.py` checks toolchain pin parity (`toolchain-pin`,
  `tool-inventory`, `install-guidance`), the CI workflow (`ci-parity`), Prettier
  formatting of maintained Markdown/YAML (`prettier-pin`, `prettier-format`),
  and (in CI) a strict `govulncheck` dependency audit (`dependency-audit`).

Each finding prints as `path: rule: fix`.

The pre-commit hook transfers the staged index in bulk into an independent shared
clone, so partially staged files are checked as committed, not as edited. Secret
scanning uses the original staged index, including Git's alternate index. Deleted
files do not become formatter inputs. Every snapshot has its own Git directory,
index, refs and config, and checker/test subprocesses receive no inherited
`GIT_*` variables. There is no pre-push hook or push-range runner: CI owns the
complete fast and seam suite. Failed tests print concise failure blocks and a
`go test JSON:` path to the full temporary JSON log.

Prettier 3.6.2 checks maintained tracked Markdown/YAML. All `evaluation/**`,
`docs/adrs/**`, scanner rule assets under
`internal/analyzers/rules/**`, governed design and ledger documents, third-party
code, instruction/migration files and `.codegraph` directories remain excluded.
Markdown fixtures under `internal/**/testdata` are also excluded. Checks never run
frozen studies. The dependency audit is strict: unavailable tooling/network and
vulnerability findings both fail.

Branch protection on the default branch requires the `gate` check.
