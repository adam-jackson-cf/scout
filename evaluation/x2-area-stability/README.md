# X2 — area stability (design §13 X2)

Question: does Scout's area reconciliation (`internal/areas.Reconcile`, file-set Jaccard match) stabilise Clawpatch agent-mapper output, and at what threshold?
Everything is specified in `PREREGISTRATION.md` (draft until the innovation council signs off with zero issues, `reviews/`). **Rev 10 (Amendment 1, PREREGISTRATION §11, pending council round 10, `reviews/round-8-amendment.md`): the first collection ran on an unpatched Clawpatch build and is excluded; no usable data exists until the restart below.**

## Layout
| File | Role |
|---|---|
| `PREREGISTRATION.md`, `GUIDE.md` | frozen design and the frozen label guide `x2-concern-v1/guide-1` |
| `common.py` | constants (grid, gates, repos, transitions), Scout's credential scrub (port of `modelcli.ScrubEnv`), Wilson interval, loaders |
| `collect.py` | records the 7 agent maps per (arm, repo) with Scout's Clawpatch environment, stop rule and provenance: `data/maps/<arm>/<repo>/<map>/{features/,meta.json}` (needs a logged-in `codex`/`claude`) |
| `replay/main.go` | replays `areas.Reconcile` per transition and threshold over the recorded maps (reads them with `clawpatch.LoadFeatures`); coverage diagnostics, chained fold (`chain.jsonl`), provenance (`replay-meta.json`) |
| `build_pairs.py` | blinded labelling units (`data/pairs/blinded.jsonl`, real and known-answer control pairs), the unblinded key and the sealed control key |
| `label.py` | agent labelling: `claude -p` + `codex exec` + adjudicator, provenance `agent_reviewed`, append-only |
| `analysis.py` | survival, spurious merges, Wilson intervals, decision rule, `unmeasured` accounting |
| `validation/validate.py`, `validation/stub_clawpatch.js`, `validation/credcheck/main.go` | planted-effect / planted-null / replay-fidelity / environment / collector checks on synthetic data |
| `run.sh` | offline stages (`validate`, `replay`, `pairs`, `label`, `analysis`) |
| `data/` | the restarted collection (empty until round 10 signs off); `data-rev8-unpatched/` is the excluded first run (codex/fastapi, unpatched build), preserved untouched and never read by a stage |
| `reviews/` | council rounds (`round-<N>.md`) and the author's responses (`round-<N>-response.md`) |

## Reproduce
Toolchain (`AGENTS.md`): `git submodule update --init && scripts/bootstrap-tools.sh && source .tools/activate`. Python is the pinned venv (`.tools/venv/bin/python`, standard library only); Go is the pinned `go1.25.6`.

Validation (runs now; no model calls, no data, about 3 min; exits non-zero on any failed check):

    source .tools/activate && evaluation/x2-area-stability/run.sh validate

Real run (**only after council sign-off**; paths from the repo root; `SOURCE_*` = any local clone or URL containing the pinned revisions). **Restart (Amendment 1).** The toolchain must hold the review-safety Clawpatch build (`scripts/bootstrap-tools.sh` rebuilds it when its marker differs from the pins; `collect.py` refuses any other build). The workspace left by the first run must be removed (`rm -rf /tmp/scout-x2`: a disposable clone) because `collect.py` refuses an existing workspace; `data/` is fresh (the first run's data is in `data-rev8-unpatched/`). The four collections run **sequentially** in the preregistered ABBA order (PREREGISTRATION §2), under `nohup` with its output redirected (as below; its ignored SIGHUP is kept) or `tmux`, so that a dropped terminal cannot abort a collection (without such protection a SIGHUP is routed to the abort record, which refuses a rerun). `setsid` is a protection only as `setsid … > log 2>&1 < /dev/null`: the collector prints after each map, and a write to a hung-up terminal would fail and abort it [INFERENCE: tty hangup semantics; not exercised]:

    PY=.tools/venv/bin/python; C=evaluation/x2-area-stability/collect.py
    nohup sh -c "$PY $C --arm codex  --repo fastapi --source $SOURCE_FASTAPI &&
                 $PY $C --arm claude --repo fastapi --source $SOURCE_FASTAPI &&
                 $PY $C --arm claude --repo axum    --source $SOURCE_AXUM    &&
                 $PY $C --arm codex  --repo axum    --source $SOURCE_AXUM" > x2-collect.log 2>&1 &
    # once x2-collect.log shows all four collections finished:
    evaluation/x2-area-stability/run.sh replay
    evaluation/x2-area-stability/run.sh pairs
    evaluation/x2-area-stability/run.sh label        # both arms' pairs in one pass; appends data/pairs/{calls,labels}.jsonl
    evaluation/x2-area-stability/run.sh analysis     # data/results.json, data/analysis_output.txt

`collect.py` refuses to start when any map directory, attempt file, versions file or stop record of the (arm, repo) exists, when its workspace directory (`/tmp/scout-x2/<arm>/<repo>`, with or without a clone in it) exists (checked before the before-probe), when the arm's CLI is not on PATH, when Clawpatch's `--version` is not 0.7.3, when the Clawpatch build is not the review-safety build (its `<build>/.scout-review-safety` marker, next to `dist/`, differs from the `<version> <sha256>` pinned in `scripts/build-clawpatch.sh`/`internal/clawpatch/review.go`, or `dist/` lacks `SCOUT_REDACTED_SECRET`; the check is recorded as `clawpatch_build_safety` in the versions file), when `CLAWPATCH_MODEL`, `CLAWPATCH_REASONING_EFFORT`, `CLAWPATCH_PROVIDER`, `CLAWPATCH_CONFIG` or `CLAWPATCH_CODEX_SANDBOX` is set (even empty), or when the model probe fails; it stops an (arm, repo) after 8 h wall, three consecutive failed maps or an abort (SIGTERM, SIGHUP, Ctrl-C, an exception, also during the after-probe and the CLI-version collection after it; `data/stopped-<arm>-<repo>.json`), ends a timed-out or aborted Clawpatch process group with SIGTERM, a 10 s grace, then SIGKILL, and runs no map for a skipped M2 mutation. A failure during the setup after the before-probe (clone, checkout, version collection) writes no record or data file: remove the workspace and rerun. `label.py` refuses a second run for the same guide version (`--new-round <name>`
labels only the pairs without a usable label) and a concurrent run (`labels.lock`); `build_pairs.py`, `run.sh pairs` and `run.sh replay` refuse once `labels.jsonl` has rows. Model calls go through `claude -p` / `codex exec` only (`AGENTS.md`), including the label sessions, which run with Scout's credential scrub; no API keys, no Infisical secret is needed or used.

## Frozen files (digests in the final review round)
`PREREGISTRATION.md`, `GUIDE.md`, `common.py`, `collect.py`, `replay/main.go`, `build_pairs.py`, `label.py`, `analysis.py`, `run.sh`, `validation/validate.py`, `validation/stub_clawpatch.js`, `validation/credcheck/main.go`; code under test (from rev 9): `scripts/clawpatch-safe.patch`, `scripts/build-clawpatch.sh`.
