# Frozen specification for the one axum replication run (development, not a preregistration)

Replicates the frozen fastapi result (`../FROZEN-8-10.md`, README "Rounds 8-10") on a second repository, axum (Rust), with the Jev thresholds, the escalation rule, the Terra prompt, the metrics and the three arms unchanged. Written before any axum data is collected (no onboarding map, no label, no Jev request, no Terra call has happened). Nothing below changes after data collection starts. This is the one run: each step runs once; a failed step is resumed (append-only), never re-run with other settings. Descriptive development figures, no pass/fail bar, no hypothesis state moves.

## Repository and window
- Repository `tokio-rs/axum`, clone `/tmp/scout-x2c-smoke/_src/axum` fetched 2026-10-31; tip = `origin/main` = `bebe175c334d2a4154b32cd9e477274c2aab5953` (2026-10-30).
- **C0** = `978ae6335862b264b4368e01a52177d98c42b2d9` (2022-12-19, "Update extract.md (#1653)"), the last first-parent commit before 2023-01-01. The catalogue is built at C0.
- **Window: first-parent merges after C0 up to the tip (2023-01-01 -> 2026-10-30), not 2025-01-01.** Reason, measured model-free before this file (the pilot's front end, no label, no model): since 2025-01-01 axum has 374 first-parent commits but only 17 PRs with a residual cluster and 22 clusters in total (most commits modify files the catalogue already owns), far below the ~120 asked for; from 2023-01-01 there are 905 commits, 69 such PRs and 120 clusters. The units merged on or after 2025-01-01 are reported as a descriptive slice (they are the last rounds' units).
- **Slicing = the pilot's round cut, earliest first** (`prepass.py`/`frontend.cut_rounds`: a round closes at the earliest merge where it holds >= 20 clusters and >= 6 PRs; at most 10 clusters scored per PR, hash-sampled). The cut gives 5 rounds, 108 scored clusters in 61 PRs (round 1 23 clusters / 11 PRs to 2023-03-28; 2 20 / 14 to 2023-11-24; 3 21 / 8 to 2024-05-02; 4 24 / 9 to 2024-10-06; 5 20 / 19 to 2026-05-16), plus 12 clusters in 8 PRs after round 5 that do not form a round and are not run. **All five rounds are built, labelled and scored (rounds 1-5, `FIRST=1 LAST=5`)**; gold of round r is teacher-forced into the membership of round r+1 exactly as in the fastapi post-run. Units counted by `build_units` (membership-aware) can differ from the 108 membership-free ones; whatever it builds is scored (window units only, no audit or control units).
- Model-free pre-pass outputs: `pilot/prepass/{rounds,scopes,summary,k0,reserve}.json`, `commits.jsonl`.

## Unchanged settings (as in `../FROZEN-8-10.md`)
- **Jev (V-sem)**: model `typesafe/jev-1.14-20261014`, v1 questions (`pilot/questions/*@v1.json`, byte-identical to the pilot's), each round's frozen inputs and `settings.json` (defaults: excerpt 60 lines, 5 sample files, PR cap 10), window units only, one pass, no fan-out. `merge.which_area` t_choice = 0.535 (permutation disagreement and below-threshold are unresolved); `merge.covered` covered score threshold s = 0.60 (`none` iff p_true <= 0.40). Budget: V-sem `-max-cost-usd` 0.5 USD across the rounds; onboarding Jev (contested files) separately capped at 0.25 USD.
- **Escalation**: every window cluster whose V-sem composed output is `none` or `unresolved` goes to Terra once (one retry on error/unparsable). Final answer = Terra's: `existing` + catalogue id -> assigned; `new` -> `none`; failure -> unresolved (`terra_failure`). V-sem `assigned` clusters are not escalated.
- **Terra**: `codex exec -m gpt-6.1-terra -c model_reasoning_effort=high -s read-only --skip-git-repo-check --ephemeral --ignore-user-config --ignore-rules -C <empty temp dir> -o <file> -`, credential scrub, 900 s; prompt `../terra_prompt_v1.txt` SHA-256 `ca24fd764b745dbcb261a1aa6a0d740710809e82a13a8a103e55bbc4d8555e76` (byte-identical, referenced in place); input = prompt + `=== STATE (JSON) ===` + the compact `noul_state` of the V-sem `merge.covered` state (unit, pr, diff_summary, the whole catalogue with id, title, summary, sample_files).
- **Gold and scoring**: agent-protocol labels (blind `claude -p` + `codex exec`, adjudicator on disagreement; provenance `agent_reviewed`), the pilot's `GUIDE.md` byte-identical (including its fastapi worked examples: they illustrate rules R1-R11, which are repository-independent), `analysis_x2d.score_outcome/counts_for/rates` (an assignment is correct iff the area is in the gold acceptable set; `none` correct iff gold is not covered; unresolved never correct).
- **Metrics (pooled over all rounds, per round, and the >= 2025-01-01 slice), each with Wilson 95 % and counts**: arms Jev-only, Jev+Terra, B2 (`baselines.py`, against the round-start membership plus earlier same-round gold). P_assign = assigned-correct / assigned; R_assign_res = assigned-correct / (gold-covered and resolved); R_assign (abstention = miss); unresolved rate; new-area P_new = correct `none` / `none` outputs and R_new = correct `none` / gold-none; gold-covered clusters answered `none`; escalations per 100 clusters; Terra correct / escalated; Terra latency (median, p90, max, total); Jev cost per round and total; catalogue size and onboarding cost; label sessions per round; unit counts, gold none / covered, every `unmeasured` item.

## Catalogue method (as the pilot, `run.sh onboard-collect` / `onboard`)
`collect_onboard.py` (unchanged): one workspace at C0, for every scope and draw k = 1..3 a fresh Clawpatch 0.7.3 state (pinned review-safety build via `.tools/activate` in the run worktree), `include` rewritten to the scope's entries, `map --source agent --provider codex --model gpt-6-luna --reasoning-effort xhigh`, plus one unscoped contrast draw; then `onboard/main.go`: code-first consensus (tau 0.5), Jev for the contested files. Output catalogue and its SHA-256 are recorded in the README.

## Axum adaptations (the only departures; all in `pilot/`, a copy of the pilot's modules at `scout-x2d` `8c7148c`; the pilot's own files are untouched)
1. `x2d.py`: `REPO_NAME`, `C0`, `TIP`; `ROOT` two levels deeper (the fork lives under `evaluation/x2d-postrun-tuning/axum/pilot`); `require_signoff()` returns False (no council sign-off record for a post-run replication; the X2 pins and the Jev model pin are still verified).
2. `frontend.py`: `scope()`, `role()` for Rust (new `CRATE_PREFIXES`, `is_test_path`): crates `axum/ axum-core/ axum-extra/ axum-macros/` = core, `examples/` = example, paths under `tests/` or `test_files/` or a `tests.rs` file = test, `.github/` = infra, rest config. Variant-stem rules (`_an`, `_py3x`, `_NN`) and `excluded_reason` (markdown, translation, media) are untouched; they are inert on axum, except `_NN` which strips a trailing 1-2 digit number from stems.
3. `onboard/main.go`: `roleOf` mirrors the new `frontend.role`; `defaultC0`.
4. `prepass.py`: `WINDOW_START` 2023-01-01 (the pilot: 2025-01-01), `PRE_START` 2022-01-01, `TIP` `origin/main` with `EXPECT_TIP`; the two `core_fastapi_*` summary keys count the four crates.
5. `scopes.py`: onboarding scopes derived at C0 for axum (the fastapi tutorial-topic sharding does not apply): `libs` (axum, axum-core, axum-extra: 131 files), `macros` (axum-macros: 148 files, 135 tests), `examples` (examples: 111), `infra` (.github and non-markdown root files: 9). All fit the Clawpatch inventory caps (400 / 500 / 200); 4 x 3 + 1 contrast = 13 maps (cap 16). The pilot's `core` / `tutorial-N` / `infra` shards are fastapi-specific.
6. `b2.py` is unchanged: its tutorial test-mirror rule (`tests/test_tutorial/test_X <-> docs_src/X`) never fires on axum, so B2 = B0, else B1 (directory majority owner, then a parent-directory walk). Disclosed, not altered.
7. Orchestration: `run_axum.sh` replaces `run_window.sh`, `run_jev.sh`, `run_partB.sh` (same commands, pointed at `pilot/` and the axum data root; the originals hard-code the fastapi pilot path and data root). `escalate.py` and `compose_window.py` are byte-identical copies of the frozen post-run scripts (their `escalation/` output lands in `axum/escalation/`). `report_axum.py` renders the tables (imports `escalate`).
- No pilot frozen file is edited. The unchanged copies in `pilot/` are byte-identical to `scout-x2d` `8c7148c`: `versioning.py`, `build_units.py`, `label_x2d.py`, `gold.py`, `baselines.py`, `compose.py`, `analysis_x2d.py`, `collect_onboard.py`, `tune.py`, `b2.py`, `jev/main.go`, `GUIDE.md`, `thresholds.json`, `questions/*`.
- Data root (new, nothing existing written): `/Users/adamjackson/Projects/inflight/scout-x2d/evaluation/x2d-merge-association-pilot/data/postrun-axum/`.

## Scripts (SHA-256)
In `evaluation/x2d-postrun-tuning/axum/`:
- `run_axum.sh` 6f76f256ace0f7bf73e79df54e048c6a82b1851e8c93d7c516c655e98a04d3d0
- `report_axum.py` (hash recorded in the README; it only renders results)
- `escalate.py` 29109416640c0576451f339b5cd350680c116c50cf446584c17a611295dad70c
- `compose_window.py` 986ba40cd8dd1726f4d80da45146f84f41b8f6d0015b546e27ffcb6d9105aca3
- `../terra_prompt_v1.txt` ca24fd764b745dbcb261a1aa6a0d740710809e82a13a8a103e55bbc4d8555e76

In `pilot/` (unchanged = equal to the pilot's frozen digests in `../FROZEN-8-10.md`):
- `versioning.py` 331c5f249d3f50fd30f3e2fba911b3bebada1717c6baf87f5486654276bcc168
- `build_units.py` c052246c828024a25fc6a143d513e111ecbf63b42b8b5912dd37e960a1c3e243
- `label_x2d.py` fa5d60ef2bbf2147102cf986048f7711cd2ad90d8cd7a0ddbe89aa5162729dcb
- `gold.py` a7831d65d40c0fa9ee15555a8488532dc646990725ac47a670f1fb6fec6bb8e5
- `baselines.py` a20c1824a1006ed37fb38cda123ce70262d8140ebcb96d44dff7f55dd165e7d0
- `jev/main.go` 63fbec0534a2ddf3d55d07c074c6a882e199ec2b90cf771a984c9fa9ee586ca0
- `compose.py` 296001923384652d7d31cdb33ca937ed409783023de2b6ee53c62e4ea8ab5bd1
- `analysis_x2d.py` 305848bff23a184c06f3d9ace8323014a7c6fc0155500858db318da38addaf19
- `GUIDE.md` 3e036ff81b1bbdd4907744fe7cd0d52c4b6323ccfe85019907036cd9ca7fa980

In `pilot/` (adapted for axum, hashes of the adapted files):
- `x2d.py` cfab705ec503599dbacd5f9127e092566867d52d5b30144e481020778141efaa
- `frontend.py` 30d3fa33cceabac0d90575fabf21212dde1cee8bbabfc2d5b80fb466eebed6fc
- `prepass.py` 9101f7d503e246adaed436a389259d220edba84ed7d8e3dc142854f582095993
- `scopes.py` f678b371fcbeda97ea9bb2b5b3515667858bfd9301ddb91e3741c4ed139ac004
- `onboard/main.go` 3859e29b5f4c52ba08aaea992636f9cc6ee65b741e6f781f488b11d83693fc0a
- `prepass/rounds.json` 77831c47191788757b8e9ce77d1aac33b09c983214ce8990075558842bb967a2
- `prepass/scopes.json` 78049bf705be34ccea6e9694a5b93af27db7b344c6cdf151ff5c5ce5b6df721f

## Statement
This is the one axum run on C0 `978ae633…`, rounds 1-5. Results are appended to `README.md` whatever they are. A deviation (a stopped map, a resumed step, a refusal) is recorded under "Deviations" in the README; no setting changes after the first onboarding map starts.
