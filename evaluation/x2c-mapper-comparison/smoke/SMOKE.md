# X2c measurement check (smoke), 2026-10-28

Rule 3 measurement check: one Clawpatch agent map per (arm, repo); no area matching, Jaccard, survival, label comparison or Jev. Each run is a single attempt (no retry), one run per cell: counts and seconds are single draws, not estimates.

## Setup
- Clawpatch: `$SCOUT_CLAWPATCH_BINARY` after `source .tools/activate` (`.tools/clawpatch/734f81bc.../dist/cli.js`); Claude Code 2.1.312 (`--effort` supported; haiku accepted `low`).
- Fixtures: `/tmp/scout-p1/fastapi` and `/tmp/scout-p4/axum-src` no longer exist; fresh clones of the pinned revisions were made at `/tmp/scout-x2c-smoke/_src/{fastapi,axum}` (`git clone` of github.com/fastapi/fastapi @ d666ccb62216e45ca78643b52c235ba0d2c53986, tokio-rs/axum @ 562fa206f09514e7503e898decf73024f5ce6fd4, `git rev-parse HEAD` verified), then a `git clone --no-hardlinks` per run to `/tmp/scout-x2c-smoke/<arm>/<repo>/ws`, state in `.../state`.
- Environment: X2's `common.clawpatch_map_env(os.environ)` (scrubbed env, `CLAWPATCH_CLAUDE_AUTH_CONTEXT=host`, both provider timeouts default 900000 ms); runner `smoke_run.py` imports frozen `evaluation/x2-area-stability/common.py` read-only. Per-call timeout 3300 s (none hit).
- Commands (per cell, `node` = the pinned node; `<E>` = env above):
  - `node $SCOUT_CLAWPATCH_BINARY --root <ws> --state-dir <state> --json init`
  - `node $SCOUT_CLAWPATCH_BINARY --root <ws> --state-dir <state> --json map --source agent --provider <codex|claude> --model <m> [--reasoning-effort <e>]`
  - arms: codex `gpt-6.1-terra` `--reasoning-effort none` / `low`; claude `claude-haiku-5-5` with no effort flag / `--reasoning-effort low`.
  - Launch: `.tools/venv/bin/python evaluation/x2c-mapper-comparison/smoke/smoke_run.py <arm> <repo>`, all 8 cells concurrently (separate workspaces); logs in `logs/`, per-cell JSON in `results/`, recount in `results/recount.txt` (`recount.py`).
- Served-model probes: codex `exec --cd <tmp> --sandbox read-only --skip-git-repo-check --model gpt-6.1-terra -c model_reasoning_effort="<e>" -` (banner `model:` / `reasoning effort:`), claude `-p --output-format json --model claude-haiku-5-5 [--effort low] --safe-mode ...` in claudeEnv(host) (`modelUsage` keys). Probes ran before and after each map.

## Results (all exit 0; workspace clean after every map)

"Agent features" = active features with `source == "agent-mapper"`; owned files = distinct paths in their `ownedFiles` (sum equals union except claude-default/fastapi 110 vs 109). Wall = `map` command only (init excluded; init succeeded rc 0 each time; JSON `seconds`). Concurrency (8 parallel maps) may inflate wall times.

| arm | repo | rc | wall s | agent features | owned files | total features (incl. deterministic) | served model (probe before/after) |
|---|---|---|---|---|---|---|---|
| codex terra none | fastapi | 0 | 124.1 | 28 | 412 | 182 | gpt-6.1-terra, effort none / same |
| codex terra none | axum | 0 | 108.2 | 41 | 216 | 118 | gpt-6.1-terra, none / same |
| codex terra low | fastapi | 0 | 177.9 | 47 | 585 | 201 | gpt-6.1-terra, low / same |
| codex terra low | axum | 0 | 67.0 | 32 | 169 | 109 | gpt-6.1-terra, low / same |
| claude haiku default | fastapi | 0 | 54.5 | 20 | 109 | 174 | claude-haiku-5-5 / same |
| claude haiku default | axum | 0 | 89.2 | 21 | 160 | 98 | claude-haiku-5-5 / same |
| claude haiku low | fastapi | 0 | 38.9 | 11 | 66 | 165 | claude-haiku-5-5 / same |
| claude haiku low | axum | 0 | 50.9 | 14 | 157 | 91 | claude-haiku-5-5 / same |

Inventory context (Clawpatch stderr): fastapi `files=1728 sourceFiles=716 ownedSourceFiles=65 weak=true`; axum `files=417 sourceFiles=158 ownedSourceFiles=149 weak=false`. The 'agent features' owned-file count is therefore not comparable to `ownedSourceFiles` (it includes non-source/test/doc files).

## Tokens / cost
- Clawpatch's `map` prints no token or cost data (stderr is progress events only: `agent-start provider= model=`, `agent-done features=N elapsed=`). Map-call tokens and cost are NOT measured here; X2c must obtain them another way (not found in the invocation used) or report them as unmeasured. [INFERENCE] not tested: the provider CLIs' own session logs.
- Probe-only: claude haiku trivial call costUSD ≈ 0.00046 (list basis), 4 output tokens; codex terra probe `tokens used` 8,939 / 10,946 (low) and 20,969 / 20,973 (none) — probe overhead, uninformative for the maps.

## Notes
- `model_option` in `agent-start` shows `--model` as passed (`gpt-6.1-terra`, `claude-haiku-5-5`); effort is not echoed by Clawpatch, only by the codex probe banner.
- claude `low` vs default produced fewer agent features on both fixtures (11 vs 20; 14 vs 21); not a comparison of quality, one run each.
- `results/*.json` field `features_agent`/`owned_files_agent` is 0: a runner bug (filtered `source == "agent"`, real value `agent-mapper`). The table above uses `recount.py`. `features_by_source` in those JSONs is correct. The runner is a measurement script, not a signed-off study script; X2c must not reuse it unreviewed.
- No arm failed or needed a different invocation. The claude `--reasoning-effort low` flag was accepted and the run succeeded; whether Clawpatch actually passes `--effort low` to claude was not inspected in the run (only X2's claim in the brief).
