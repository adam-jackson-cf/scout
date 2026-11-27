# Contributing

## Install pinned tools

Install a Go bootstrap toolchain first (for example, `mise use go@1.26.11`).
The installer fails clearly if Go is unavailable; Go's toolchain downloader
then selects the version in `toolchain/versions.env`.

```sh
git submodule update --init
scripts/bootstrap-tools.sh
source .tools/activate
```

Tool pins live only in `toolchain/versions.env`; when changing an allowed developer
tool pin, update its package metadata and run `scripts/bootstrap-tools.sh --relock`.
Scanner and Clawpatch upgrades require their compatibility review.

Build and test with the toolchain activated:

```sh
go build ./... && go vet ./... && go test -count=1 ./...
```

For what to work on, see `docs/HANDOFF.md`.

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

