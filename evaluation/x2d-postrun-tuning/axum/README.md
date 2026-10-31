# X2d post-run replication on axum (Rust): the frozen fastapi settings, unchanged (development, not evidence)

One run, as specified in `FROZEN-AXUM.md` (SHA-256 `b696e6d10e41047c45c624c5d2decb0d54934573b6cec29e2816fb76abd5df4a`, written before any data was collected and not changed since). Same Jev thresholds (V-sem, t_choice 0.535, covered score threshold s 0.60), same escalation rule, same Terra prompt (`../terra_prompt_v1.txt`, SHA-256 `ca24fd76…5e76`), same metrics and three arms as `../FROZEN-8-10.md`. Descriptive figures; no hypothesis state moves; no pass/fail bar.

Result in one line: **the fastapi result does not replicate on axum.** Jev+Terra P_assign is 0.622 [0.514, 0.719] (fastapi 0.922 [0.840, 0.964]) because 23 of its 31 wrong assignments are on gold-`none` units (axum has 46 gold-`none` of 109, fastapi 15 of 92) and Jev rarely answers `none`; on gold-covered units resolved recall is 0.81 and Terra removes all abstention, as on fastapi.

## Setup
- Repository `tokio-rs/axum`, tip `bebe175c334d2a4154b32cd9e477274c2aab5953` (2026-10-30). C0 `978ae6335862b264b4368e01a52177d98c42b2d9` (2022-12-19). **Window 2023-01-01 -> tip, not 2025-01-01** (deviation by design, recorded in the freeze file): since 2025-01-01 the pilot's front end finds only 22 clusters in 17 PRs; from 2023-01-01 it finds 120. The pilot's round cut gives five rounds (rounds 1-5, all run); the units merged on or after 2025-01-01 are the last 10 units and are reported as a descriptive slice.
- Catalogue (pilot method: 4 scopes x 3 Clawpatch 0.7.3 `codex gpt-6-luna xhigh` agent maps + 1 contrast map, all 13 maps rc 0, 1 attempt each, 19 min wall; consensus tau 0.5; Jev for contested files): **96 areas** (the fastapi catalogue has 114), 54 contested files, 179 unowned files, 164 assigned by Jev; catalogue SHA-256 `c4cfb2f252b9e72b009dffa46e6bcbcc6cc2106eeb87313ed941f348197e9b9e`. Onboarding cost: Jev 0.080323 USD (1,743 requests; the 13 Luna maps run on the CLI subscription and are not metered). Scopes: `libs` (axum, axum-core, axum-extra, 131 files), `macros` (axum-macros, 148), `examples` (111), `infra` (9).
- Units, labels, sessions (window units; labelled cases include controls and audit items): round 1: 23 units / 11 PRs, 28 cases, 65 sessions of cap 212; round 2: 20 / 14, 25, 56 of 204; round 3: 22 / 8, 27, 57 of 210; round 4: 24 / 9, 29, 68 of 215; round 5: 20 / 19, 28, 63 of 212. Total **109 units** (the membership-free pre-pass counted 108; round 3 has 22 against 21), 309 sessions. Gold none / covered: 13 / 10, 5 / 15, 3 / 19, 15 / 9, 10 / 10 = **46 / 63** (none kinds: 37 gap, 9 new). Of the 109 window units, 81 agreed and 28 were adjudicated.
- Jev (V-sem, pinned `typesafe/jev-1.14-20261014`): 244 requests, 0 failures, cost 0.020981 + 0.018128 + 0.020089 + 0.020750 + 0.017181 = **0.097129 USD** (budget 0.5); with the onboarding Jev step 0.177452 USD in total. About 0.089 USD per 100 window units (fastapi rounds 8-10: 0.128).
- Escalations: **32 of 109 = 29.4 per 100** (fastapi 27.2); Terra correct on 27 of 32; 0 failed or unparsable calls; reported model `gpt-6.1-terra`, effort `high` on all calls; latency per call median 5.6 s, p90 7.0 s, max 10.3 s, total 186.3 s (6 in parallel).

## Three arms (rounds 1-5, 109 units)
Cells: value [Wilson 95 %] (count). `R_assign` counts abstention as a miss; `R_assign_res` = assigned-correct / (gold-covered and resolved). none P = correct `none` / `none` outputs; none R = correct `none` / gold-none units. Covered-none = gold-covered units answered `none`.

| arm | P_assign | R_assign_res | R_assign | unresolved | none P (new-area P) | none R (new-area R) | covered-none |
|---|---|---|---|---|---|---|---|
| Jev-only | 0.610 [0.499, 0.712] (47/77) | 0.839 [0.722, 0.913] (47/56) | 0.746 [0.627, 0.837] (47/63) | 0.193 [0.130, 0.277] (21/109) | 0.909 [0.623, 0.984] (10/11) | 0.217 [0.123, 0.356] (10/46) | 1 |
| Jev+Terra | 0.622 [0.514, 0.719] (51/82) | 0.810 [0.696, 0.888] (51/63) | 0.810 [0.696, 0.888] (51/63) | 0.000 [0.000, 0.034] (0/109) | 0.852 [0.675, 0.941] (23/27) | 0.500 [0.361, 0.639] (23/46) | 4 |
| B2 | 0.556 [0.396, 0.705] (20/36) | 0.345 [0.236, 0.473] (20/58) | 0.317 [0.216, 0.440] (20/63) | 0.101 [0.057, 0.172] (11/109) | 0.484 [0.364, 0.606] (30/62) | 0.652 [0.508, 0.773] (30/46) | 32 |

Wrong assignments by cause (`split.json`): Jev-only 30 = 22 on gold-`none` units + 8 wrong area; Jev+Terra 31 = 23 + 8; B2 16 = 10 + 6. The residual wrong-area count is the same (8) with and without Terra: the extra P_assign errors come from gold-`none` units that Jev, and then Terra, assign to an area; the `none` output has high precision but low recall (Jev-only 10 of 46 gold-none, Jev+Terra 23 of 46).

fastapi rounds 8-10 for comparison (`../README.md`): Jev-only P_assign 0.955, R_assign 0.831, unresolved 0.109, new-area P/R 0.933 / 0.933; Jev+Terra 0.922, 0.922, 0, 0.867 / 0.867; B2 0.766, 0.636, 0.207, 0.333 / 0.200. On axum Jev+Terra beats B2 on every measured rate except none R (0.500 against 0.652), and its resolved recall (0.810) is above B2's (0.345), but P_assign intervals of Jev (0.514-0.719) and B2 (0.396-0.705) overlap.

Per round (Jev+Terra, P_assign / R_assign / new-area R; full per-arm, per-round table in `report.md`): round 1 0.562 (9/16) / 0.900 (9/10) / 0.462 (6/13); round 2 0.733 (11/15) / 0.733 / 0.800 (4/5); round 3 0.789 (15/19) / 0.789 / 0.667 (2/3); round 4 0.529 (9/17) / 1.000 (9/9) / 0.467 (7/15); round 5 0.467 (7/15) / 0.700 (7/10) / 0.400 (4/10). Rounds 1 and 4 and 5 are gold-none heavy (13, 15, 10 of 23, 24, 20).

Units merged on or after 2025-01-01 (the originally requested window; 10 units, descriptive only): Jev-only and Jev+Terra identical, P_assign 0.500 [0.215, 0.785] (4/8), R_assign 1.000 [0.510, 1.000] (4/4), none R 0.333 (2/6); B2 P_assign 0.333 (2/6), R_assign 0.500 (2/4).

`unmeasured`: none. Cells with no `none` output are `n/a` (round 3 Jev-only none P, 0/0); small cells (rounds, the 2025 slice) have wide intervals and should not be read as differences.

## Deviations and adaptations (all disclosed; none changes a frozen setting)
- Window start 2023-01-01 instead of 2025-01-01 (reason above; decided and recorded before any data collection).
- Onboarding Jev step: the first `run_axum.sh onboard` run had 8 of 1,743 requests fail (no catalogue written); it was resumed once, append-only as the freeze allows, answering the 8 requests (cumulative billed cost 0.080323 USD). No other failure, resume or refusal.
- `run_axum.sh` was run with `env -u SCOUT_TOOLS_ACTIVE` (the persistent shell already had the toolchain activated; the activation script is called by the script itself). No script edit.
- Axum adaptations of the pilot code, listed in `FROZEN-AXUM.md` section "Axum adaptations": `x2d.py` (constants, `ROOT`, sign-off check off), `frontend.py` and `onboard/main.go` (Rust `scope()` / `role()`), `prepass.py` (window, tip), `scopes.py` (4 axum scopes), orchestration in `run_axum.sh` (the frozen `run_window.sh` / `run_jev.sh` / `run_partB.sh` hard-code the fastapi pilot path and data root). `escalate.py` and `compose_window.py` are byte-identical copies of the frozen post-run scripts. `GUIDE.md`, the question files, `thresholds.json`, `b2.py` and every Go/Python module not listed are byte-identical to the pilot's (the B2 tutorial-mirror rule never fires on axum, so B2 = directory majority, then parent walk).
- The label guide keeps its fastapi worked examples (rules R1-R11 are repository-independent). Whether that moved the axum labels is not tested.
- `report_axum.py` (written before collection, renders tables) and `split_axum.py` (written after the results, model-free description of wrong assignments) are not in the freeze file's hash list: SHA-256 `e1f11a1082d2cde5fa04a9b0220d75dfa2a183a15c0953c22064999730f6efee` and `f54c1a4e550110a443b447c204fdb0442161c28b5c92098a6ac6355256480727`.
- Labels are agent-protocol, provenance `agent_reviewed`; a fresh run produces different labels (model sessions are not deterministic), the numbers above are for the labels in the data root.

## Files
`FROZEN-AXUM.md`, `run_axum.sh`, `escalate.py`, `compose_window.py`, `report_axum.py`, `split_axum.py`; `pilot/` (adapted copy of the pilot modules, `prepass/` outputs, `scopes.json`); `report.md`, `report.json` (every arm / part with counts and Wilson intervals), `split.json`; `escalation/axum-r1-5.{calls.jsonl,report.json,errors.txt}` (the 32 Terra calls and the 5 Terra errors). Data: `/Users/adamjackson/Projects/inflight/scout-x2d/evaluation/x2d-merge-association-pilot/data/postrun-axum/` (catalogue, onboard draws and attempts, `onboard-jev/`, `rounds/1..5/` units, labels, gold, arms, jev scores, states).

## Reproduce
From the scout checkout (`env -u SCOUT_TOOLS_ACTIVE` if a toolchain is already active; needs the run worktree `scout-x2d` with the pinned toolchain, `claude -p`, `codex exec`, the Jev key via Infisical only):
```
R=evaluation/x2d-postrun-tuning/axum/run_axum.sh
git -C /tmp/scout-x2c-smoke/_src/axum fetch origin && git -C /tmp/scout-x2c-smoke/_src/axum checkout --detach bebe175c334d2a4154b32cd9e477274c2aab5953
$R onboard-collect                       # 13 maps at C0
infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- $R onboard
$R window                                # rounds 1-5: versioning, units, labels, gold
$R baselines && $R states
infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- $R jev
$R escalate && $R report                 # Terra, then report.md / report.json
X2D_DATA=<data root> <worktree>/.tools/venv/bin/python evaluation/x2d-postrun-tuning/axum/split_axum.py evaluation/x2d-postrun-tuning/axum/pilot 1-5
```
Model-free pre-pass (reproduces `pilot/prepass/` byte for byte): `<venv python> pilot/prepass.py --repo $SRC --out pilot/prepass` and `pilot/scopes.py --repo $SRC --rev 978ae6335862b264b4368e01a52177d98c42b2d9 --out pilot/prepass/scopes.json`.
