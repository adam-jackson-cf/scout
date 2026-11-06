# Handoff: continue Scout on another machine (2026-11-04)

Everything needed is on `origin/main` (https://github.com/adam-jackson-cf/scout). No other branch, worktree
or stash holds unmerged work (details in §5). Read order after this file: `CLAUDE.md`, `docs/STATUS.md`,
`docs/plans/2026-11-01-rev-2.8-merge-association.md`, `docs/scout-design.md` (rev 2.8, §4 lines 114–116, Appendix 23).

## 1. Set up the new machine

```sh
git clone https://github.com/adam-jackson-cf/scout.git && cd scout
git submodule update --init            # third_party/clawpatch at v0.7.3 (734f81bc)
scripts/bootstrap-tools.sh             # pinned toolchain into .tools/ (versions: toolchain/versions.env)
source .tools/activate
go build ./... && go vet ./... && go test -count=1 ./...   # was green at handoff
shasum -a 256 docs/scout-design.md     # must print d1fb127446be290f75d4de5416ebf40b214552f18eedc1d1a335819f314f9b75
```

`scripts/bootstrap-tools.sh` does not install Go. It needs a system `go` on `PATH` and stops with "Go is required" if none
is found. It then uses `GOTOOLCHAIN=go$GO_VERSION` (`toolchain/versions.env`, now 1.26.10) to fetch the pinned Go.
One way that does not change global config: `mise install go@1.26.10`, then put `$(mise where go@1.26.10)/bin` on `PATH`.

Install the repository pre-commit hook once per clone: `git config core.hooksPath .githooks`
(see `CONTRIBUTING.md`). Schema-v2 `.omp/gate.json` owns `checks.commit` and
`checks.ci`; run the complete check-only suite with `bash scripts/gate.sh ci`.
No pre-push hook is installed.

Accounts and tools the work needs (none are stored in the repo):
1. `claude` and `codex` CLIs logged in with their subscriptions. The fallback lane and post-run scripts use
   `codex exec -m gpt-6.1-terra` (effort high); labelling uses both CLIs.
2. Infisical CLI logged in. The Jev key (`OPENROUTER_API_KEY`) is only injected per command:
   `infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- <command>`. Never export or print it.
3. `gitleaks` (provided by `.tools`). Gate before every commit:
   `gitleaks git --pre-commit --staged --no-banner --redact .` (`.gitleaksignore` lists the reviewed public-example hits).
4. The council agent set-up (lead Opus 5.5 high, challenger GPT-6 Astra high, peer Opus 5.5 high) lives in
   `~/.omp/agent/config.yml` on the old machine, outside this repo. Copy it over if the same omp set-up is used.
5. Do not copy `.codegraph/` or the ignored contents of `.scout/` (`state.db` and other local state; `.gitignore` has `.scout/*`).
   `.scout/policy.yaml` is tracked and arrives with the clone. Scout's published state is on the `origin/scout/state` branch.

Fixture clones (used by replay generators and study scripts; the old `/tmp` copies are gone, re-clone):
```sh
mkdir -p /tmp/scout-x2c-smoke/_src && cd /tmp/scout-x2c-smoke/_src
git clone https://github.com/fastapi/fastapi.git
git clone https://github.com/tokio-rs/axum.git && git -C axum checkout --detach bebe175c334d2a4154b32cd9e477274c2aab5953
```
Every stage verifies pinned SHAs, not paths, so any clone holding them works (pass `--fastapi`/`--axum` or `X2D_SRC`).

## 2. Where the work stands

Design revision 2.8 (merge-time area association, provisional opt-in, off by default) is adopted: commit `1d4e0cc`,
council rounds in `docs/design-reviews/2026-10-31-jev-merge-association/` (content round 8, adoption round 11).

The flow being built:
1. Residual files of each default-branch merge go to Jev (`merge.covered`, then `merge.which_area`).
2. Clusters Jev leaves unresolved or answers `none` go to `codex exec gpt-6.1-terra` (effort high).
3. Terra picks an existing area or proposes a new feature; a new feature becomes a *pending* area that only a maintainer accepts.

| Plan phase | State |
|---|---|
| 0 Adopt rev 2.8 | done, `1d4e0cc` |
| 1 State: schema 13, `associations`, pending areas, export/import | done, `3e89f23` (export→import→export byte-identical) |
| 2 Residual files and clusters (no model), `internal/association/` | done, `3e89f23` (parity 2,805/2,806 with the pilot; the 1 difference is the deliberately unported translation rule) |
| 3 `association` lane (Jev) | **next** |
| 4 `association-fallback` lane (codex Terra) | pending |
| 5 `scout areas list --pending / accept / merge` | pending |
| 6 Refresh guard | pending |
| 7 `orderflip` B-3 record on `merge.which_area` + docs | pending; gates any release that lets the lane be enabled |

## 3. Next steps (Phase 3 onward)

Phase 3, `association` lane (plan §Phase 3 lists the files):
- Register `merge.covered@v1` and `merge.which_area@v1` in `internal/decisions` with the pilot's frozen question text and GUIDE criteria.
- Add `mapping.association.*` policy keys in `internal/policy/mapping.go`; push-only optional lane in `internal/engine/lanes.go`, `requires: decision-egress`.
- Wire `association.Run` into `scout assess`. Carry these Phase 2 departures:
  1. add a store-backed implementation of `association.Ownership` (`internal/association/residual.go`) with all four
     methods: `AreaOwners`, `PendingAreas`, `HeuristicOnly`, `SiblingFiles`. Phase 2 has only the `overlay` wrapper in
     `orchestrate.go`, which delegates to a base, and the test fixture `fixtureOwnership` in `parity_test.go`;
  2. its `AreaOwners` must include association-derived owners, not just map owners;
  3. the `association.Decider` callback is called once per merge;
  4. map the lane's results to `domain.Association`.
- Acceptance: a replay test reproducing the fastapi rounds 8–10 Jev-only outcomes, **64/67 correct and 10/92 unresolved**
  (`evaluation/x2d-postrun-tuning/README.md` "Rounds 8-10" table; settings frozen in `FROZEN-8-10.md`). Build the fixture the way `internal/association/testdata/gen/gen_parity.py` does.

Phase 4 acceptance: replay with a fake `codex` reproducing **71/77** (prompt `evaluation/x2d-postrun-tuning/terra_prompt_v1.txt`, digest-checked).
Phase 5 follows the `scout triage` `--actor/--reason/--authorize` pattern. Order: 3 and 4 need 1–2; 5 needs 1; 6 needs 1 and 3–4; 7 last.
Each phase: `go test ./...`, lint, the phase scenario run and recorded in `docs/STATUS.md`.

Known limit, stated in the design: the fastapi result (P_assign 0.922) did not replicate on axum (0.622), which is why the lane ships as a provisional opt-in (R-061).

## 4. Study data now in the repo

The X2d pilot and post-run data (`evaluation/x2d-merge-association-pilot/data/`, 97 MB, previously untracked in the
`scout-x2d` worktree) is committed on `main`, so replay fixtures can be regenerated without the old worktree.
`gen_parity.py` defaults `--x2d` to that in-repo path (run it from the repository root) and `--fastapi`/`--axum` to the
`/tmp/scout-x2c-smoke/_src` clones, so make the §1 fixture clones first or pass other clone paths:
```sh
python3 internal/association/testdata/gen/gen_parity.py \
    --fastapi /tmp/scout-x2c-smoke/_src/fastapi --axum /tmp/scout-x2c-smoke/_src/axum
```
Post-run study scripts in `evaluation/x2d-postrun-tuning/` (`run_window.sh`, `run_jev.sh`, `run_partB.sh`) are frozen
(their digests are in `FROZEN-8-10.md`), so their old-machine defaults stay: `X2D_WORKTREE` defaults to
`/Users/adamjackson/Projects/inflight/scout-x2d` and `X2D_SRC` (fastapi only) to `/tmp/scout-x2c-smoke/_src/fastapi`.
Override both with the environment variables. To re-run them, recreate that worktree and give it the data:
```sh
git worktree add --detach ../scout-x2d 8c7148c
cp -R evaluation/x2d-merge-association-pilot/data ../scout-x2d/evaluation/x2d-merge-association-pilot/
export X2D_WORKTREE=$PWD/../scout-x2d     # if the path differs from the old machine
(cd ../scout-x2d && git submodule update --init && scripts/bootstrap-tools.sh)
```
The pilot's files at `8c7148c` are frozen (experiment rule 5); do not edit them.

## 5. What was checked before migration

- Branches: only `main` (= `origin/main`) and `origin/scout/state` (Scout's state-publishing branch, intentionally separate).
- Worktrees `scout-x2b` (`c482b37`), `scout-x2c` (`9965410`), `scout-x2d` (`8c7148c`): all detached at commits already in `main`.
  Untracked content: x2c data is identical to the copy tracked on `main`; x2d data is now committed (§4); the x2b collection
  log is committed as `evaluation/x2b-area-stability-luna/data/x2b-collect.log`. The worktrees can be deleted after migration.
- Stash `stash@{0}` (`omp-isolation-15b2ace4a8a2fcd5`, 2026-10-24): an isolation snapshot of the rev 2.6 "remove refresh-on-merge"
  change, superseded by commit `ead1a3f` 37 minutes later. Not migrated; nothing to recover.

## 6. Open items outside this plan

- Design §15: Jev false-action costs and per-decision minimums await a user decision (provisional defaults in force).
- HD6 is `inconclusive` (X2d pilot), HD8 `proposed`; X11 needs a council-signed preregistration before any run.
- Inertia-bench (`~/Projects/tools/inertia-bench`) is a separate repo with its own `HANDOFF.md`; migrate it separately.

## 7. Rules to keep (full text in `CLAUDE.md`)

- Any change to `docs/scout-design.md` needs innovation-council sign-off with zero issues, recorded under `docs/design-reviews/`.
- Experiments: preregister, planted-effect and planted-null checks, council sign-off, frozen after sign-off.
- Model use: `claude -p` / `codex exec` only; OpenRouter only for Jev. Never print secrets. Never touch `.codegraph`.
- Ledgers are append-only: `docs/research-ledger.md` (next free R-062), `docs/hypothesis-ledger.md`.
