# X2d post-run threshold tuning (development, not evidence)

User direction 2026-10-31 (R-055): tune the X2d Jev thresholds on the labelled data the stopped pilot produced, per the jev-actions skill. This is instrument development after the pilot ended; it moves no hypothesis state and is not a preregistered test.

- Data: every gold-labelled V-sem case with saved scores: round-0 main run (64), round-0 P1 (20), re-gate P1 (40), re-gate P2 (40) = 164 cases. Round-0 P2 excluded (derived gold, shown mis-built). Labels: agent protocol, provenance `agent_reviewed`.
- Method: the pilot's pinned `tune_threshold.py` (sha `2f850933…2e4c`), the pilot's minimums and betas from `thresholds.json`, `--folds 5 --repeats 5 --seed 20261029`, conservative and point-estimate runs. No model calls.
- Reproduce: `source .tools/activate`, then from the run worktree `X2D_DATA=$PWD/evaluation/x2d-merge-association-pilot/data/postrun .tools/venv/bin/python <scout>/evaluation/x2d-postrun-tuning/tune_pool.py $PWD/evaluation/x2d-merge-association-pilot [OUT]` (default OUT `/tmp/x2d-tune`; it needs only the study data, not postrun). Round-0 results below were produced before the 2026-10-31 fix: `which_area` rows copied their `question_version` from the covered row; they now carry `asked[0]`'s, and mixed versions are blanked for the script as `tune.run_script` does. The shipped thresholds and counts do not depend on this.

Results:
- `merge.which_area` (t_choice; min P 0.90, R 0.80, beta 0.5): 124 positive / 40 negative. Point estimate: shipped threshold 0.6875; held-out P 0.909, R 0.818 (default 0.5: P 0.836, R 0.944); fold range 0.615–0.6875 (spread > 0.05) and 15 of 25 folds below the minimums held-out, so status `tuned-provisional`. Conservative (95 % lower bounds): no threshold qualifies at this size.
- `merge.covered` (t_cov; min P 0.80, R 0.80): only 16 `none` positives against 148 negatives (below the skill's ~30 per label). No threshold meets the minimums in any fold, even on point estimates (default 0.5: P 1.00, R 0.38). Stays provisional 0.7.
- Effect on abstention (choice step only): re-gate P1 0.47 at 0.70, 0.45 at 0.6875, 0.35 at 0.60; round-0 main 0.23 / 0.20 / 0.12. The tuned value is effectively the starting 0.70, so threshold tuning does not explain the re-gate abstention.

## Window merges, refinement rounds 1-7 (development, not evidence)

User direction 2026-10-31 (R-055, extended): tune on real window merges, rounds 1-7 only. Rounds 8-10 were never built, labelled or scored (`rounds/8/membership.json` exists only because `gold.py --round 7` writes the next membership; there are no round-8 units).

How the data was made (run worktree `scout-x2d`, detached at `8c7148c`, frozen files untouched; own data root `data/postrun/` with a byte copy of `data/catalogue/` and the round-0 `gates.json`, so the core gate is off):
- Units: `versioning.py --round r` then `build_units.py --round r` per round in order, each round's membership teacher-forced from the previous round's gold (`gold.py`). Default settings (excerpt 60 lines, PR cap 10). The membership-aware count (200 window units) exceeds the membership-free prepass count (158).
- Labels: `label_x2d.py --round r --workers 6` (blind `claude -p` and `codex exec`, plus adjudicator; provenance `agent_reviewed`) on every unit of the round (window, audit, controls), then `gold.py`. The perception (`jev-state`) labels were not made; tuning does not use them.
- Jev: `jev/main.go -arms sem` on the window units only (no audit or control units, no fan-out), pinned model, v1 questions, each round's frozen inputs and settings, one pass; `-max-cost-usd` 1.0 total across rounds.
- Scripts here: `run_window.sh` (units, labels, gold), `run_jev.sh`, `tune_window.py`, `compose_window.py`, `tune_pool.py` (the earlier 164 cases). `window/data/<r>/` archives what the reports need (gold, B2, slimmed window units, gzipped Jev scores, sessions); raw labels and prompts stay in the worktree.

Per round (window units are the scored units; labelled cases = wide label rows incl. audit and controls; label counts are over the window units):

| round | window units | PRs | audit | controls | labelled cases | sessions / cap | gold none / covered | which_area pos / neg | Jev requests | Jev cost USD |
|---|---|---|---|---|---|---|---|---|---|---|
| 1 | 23 | 16 | 5 | 5 | 33 | 69 / 226 | 3 / 20 | 16 / 7 | 51 | 0.029667 |
| 2 | 24 | 10 | 1 | 5 | 30 | 61 / 218 | 0 / 24 | 21 / 3 | 51 | 0.031130 |
| 3 | 22 | 17 | 0 | 5 | 27 | 55 / 210 | 1 / 21 | 19 / 3 | 50 | 0.029259 |
| 4 | 26 | 14 | 10 | 5 | 41 | 82 / 248 | 0 / 26 | 25 / 1 | 59 | 0.035622 |
| 5 | 42 | 12 | 10 | 7 | 59 | 122 / 298 | 5 / 37 | 27 / 15 | 91 | 0.060433 |
| 6 | 38 | 17 | 10 | 6 | 54 | 109 / 284 | 3 / 35 | 30 / 8 | 83 | 0.052759 |
| 7 | 25 | 6 | 3 | 5 | 33 | 76 / 226 | 1 / 24 | 18 / 7 | 56 | 0.033523 |
| total | 200 | | | | 277 | 574 | 13 / 187 | 156 / 44 | 441 | 0.272393 |

No Jev failures. Cost is about $0.136 per 100 window units (2.2 requests each). `merge.covered` positive = gold none (13 of 200); `which_area` positive = Jev's choices agree and are acceptable.

Tuning (`tune_window.py`; pinned script, the pilot's minimums and betas from `thresholds.json`, `--folds 5 --repeats 5 --seed 20261029`; rounds 1-5 `dev`, rounds 6-7 `test`; which_area rows are every gold-labelled window unit with both answers, as in `tune_pool.py`). Reports and CSVs are in `window/`.

| decision (minimums, beta) | rows | dev pos / neg | test pos / neg | point-estimate shipped | held-out dev P / R (default 0.5) | fold range, folds tuned | test P / R at shipped (at 0.5) | status |
|---|---|---|---|---|---|---|---|---|
| merge.covered, window (P 0.80, R 0.80) | 200 | 9 / 128 | 4 / 59 | none qualifies | none (0.5: P 0.27, R 0.33) | n/a, 0 of 25 | n/a | provisional 0.70 (< 30 per label, no tuning) |
| merge.covered, window + earlier | 364 | 25 / 276 | 4 / 59 | none qualifies | none (0.5: P 0.53, R 0.36) | n/a, 0 of 25 | n/a | provisional 0.70 |
| merge.which_area, window (P 0.90, R 0.80, beta 0.5) | 200 | 108 / 29 | 48 / 15 | t_choice 0.535 | 0.900 / 0.942 (0.904 / 0.954) | 0.4975-0.54, 24 of 25 | 0.822 / 0.771 (0.804 / 0.771) | tuned-provisional; test below minimums |
| merge.which_area, window + earlier | 364 | 232 / 69 | 48 / 15 | t_choice 0.6125 | 0.892 / 0.817 (0.866 / 0.948) | 0.577-0.688, 22 of 25 | 0.892 / 0.688 (0.804 / 0.771) | tuned-provisional; fold spread > 0.05, test R below minimum |

Conservative (95 % lower bounds) runs: no threshold qualifies for any decision or variant. Test by round (which_area, shipped threshold): window variant r6 P 0.839 / R 0.867, r7 P 0.786 / R 0.611; window + earlier r6 0.893 / 0.833, r7 0.889 / 0.444. Failed folds are listed in `window/report.*.txt`. In the pooled CSVs `question_version` is blank (earlier and window which_area versions differ); the window-only which_area CSV has one version.

Composed check (descriptive; `compose_window.py`): V-sem composed with `compose.compose` over the window units, scored with `analysis_x2d.score_outcome`, `counts_for` and `rates`. covered stays at 0.70 in every row because no run tunes it. P_assign = assigned-correct / assigned; R_assign_res = assigned-correct / (gold covered and resolved); abstention = unresolved / units; none P = correct `none` / `none` outputs; none R = correct `none` / gold-none units (an abstained gold-none unit is a miss). The tuned t_choice values come from dev rounds 1-5 (window variant) and from dev plus the earlier cases; "test" is rounds 6-7.

| setting | part | P_assign | R_assign_res | abstention | none P | none R |
|---|---|---|---|---|---|---|
| 0.70 / 0.70 | all 200 | 0.923 (84/91) | 0.613 (84/137) | 0.255 (51/200) | 0.155 (9/58) | 0.692 (9/13) |
| | test 63 | 1.000 (22/22) | 0.489 (22/45) | 0.238 (15/63) | 0.115 (3/26) | 0.750 (3/4) |
| tuned window, t_choice 0.535 | all 200 | 0.914 (106/116) | 0.654 (106/162) | 0.130 (26/200) | 0.155 (9/58) | 0.692 (9/13) |
| | test 63 | 0.931 (27/29) | 0.519 (27/52) | 0.127 (8/63) | 0.115 (3/26) | 0.750 (3/4) |
| tuned window + earlier, t_choice 0.6125 | all 200 | 0.907 (98/108) | 0.636 (98/154) | 0.170 (34/200) | 0.155 (9/58) | 0.692 (9/13) |
| | test 63 | 0.929 (26/28) | 0.510 (26/51) | 0.143 (9/63) | 0.115 (3/26) | 0.750 (3/4) |
| B2 baseline | all 200 | 0.452 (71/157) | 0.420 (71/169) | 0.105 (21/200) | 0.045 (1/22) | 0.077 (1/13) |
| | test 63 | 0.350 (14/40) | 0.286 (14/49) | 0.175 (11/63) | 0.083 (1/12) | 0.250 (1/4) |

Dev (rounds 1-5) rows for each setting are in `window/composed.json`. Reading: lowering t_choice from 0.70 to 0.535 cuts abstention from 0.255 to 0.130 for about 1 point of P_assign, but resolved recall stays far below 0.80. The binding problem is `merge.covered`: at 0.70 it emits `none` for 58 of 200 units against 13 gold-none (none P 0.155), and 13 positives is too few to tune it.

Reproduce (from the scout checkout; `WT=/Users/adamjackson/Projects/inflight/scout-x2d`, `source $WT/.tools/activate`; model steps use `claude -p`, `codex exec` and the Jev key via Infisical only):
```
evaluation/x2d-postrun-tuning/run_window.sh 1 7    # units, labels, gold (labelling refuses to redo a round that already has labels)
for r in 1 2 3 4 5 6 7; do X2D_DATA=$WT/evaluation/x2d-merge-association-pilot/data/postrun $WT/.tools/venv/bin/python $WT/evaluation/x2d-merge-association-pilot/baselines.py --round $r --repo /tmp/scout-x2c-smoke/_src/fastapi; done
infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- evaluation/x2d-postrun-tuning/run_jev.sh
X2D_DATA=$WT/evaluation/x2d-merge-association-pilot/data/postrun $WT/.tools/venv/bin/python evaluation/x2d-postrun-tuning/tune_window.py $WT/evaluation/x2d-merge-association-pilot
X2D_DATA=$WT/evaluation/x2d-merge-association-pilot/data/postrun $WT/.tools/venv/bin/python evaluation/x2d-postrun-tuning/compose_window.py $WT/evaluation/x2d-merge-association-pilot
```
A fresh run produces different labels (model sessions are not deterministic); the numbers above are for the labels in the worktree's `data/postrun`.

## Best-guess starting point (development)

Saved scores only (`sweep_window.py`, `window/sweep.json`; same units, gold and composition as above). The covered cut is stated as the score threshold s of `compose`: `none` iff 1 - p_true >= s, i.e. p_true <= 1 - s (s = 1 - t_cov; the compose `t_cov` 0.70 used so far is s 0.30). So the requested "t_cov 0.80-0.99" values, read as `none` iff p_true < 1 - t_cov, are s 0.80-0.99: they give no `none` at all (only p_true <= 0.2 or lower qualifies). I swept s over 0.30 (old setting), 0.40, 0.50, 0.60, 0.70, 0.80, 0.85, 0.90, 0.95, 0.99 and never-none; s 0.80 to 0.99 are identical to never-none on these units, so one row stands for them. "Covered-none" = gold-covered units answered `none`. Cells: all 200 / test 63 (rounds 6-7).

| t_choice | s | P_assign | R_assign_res | abstention | none count | none P | none R | covered-none |
|---|---|---|---|---|---|---|---|---|
| 0.535 | 0.30 | 0.914 / 0.931 | 0.654 / 0.519 | 0.130 / 0.127 | 58 / 26 | 0.155 / 0.115 | 0.692 / 0.750 | 49 / 23 |
| 0.535 | 0.40 | 0.921 / 0.939 | 0.745 / 0.608 | 0.165 / 0.143 | 40 / 21 | 0.175 / 0.143 | 0.538 / 0.750 | 33 / 18 |
| 0.535 | 0.50 | 0.906 / 0.865 | 0.822 / 0.696 | 0.195 / 0.222 | 23 / 12 | 0.174 / 0.083 | 0.308 / 0.250 | 19 / 11 |
| 0.535 | 0.60 | 0.904 / 0.875 | 0.880 / 0.795 | 0.205 / 0.254 | 13 / 7 | 0.231 / 0.143 | 0.231 / 0.250 | 10 / 6 |
| 0.535 | 0.70 | 0.888 / 0.837 | 0.918 / 0.878 | 0.225 / 0.302 | 3 / 1 | 0 / 0 | 0 / 0 | 3 / 1 |
| 0.535 | 0.80-0.99, never | 0.890 / 0.841 | 0.939 / 0.902 | 0.225 / 0.302 | 0 / 0 | n/a | 0 / 0 | 0 / 0 |
| 0.6125 | 0.30 | 0.907 / 0.929 | 0.636 / 0.510 | 0.170 / 0.143 | 58 / 26 | 0.155 / 0.115 | 0.692 / 0.750 | 49 / 23 |
| 0.6125 | 0.40 | 0.915 / 0.938 | 0.730 / 0.600 | 0.210 / 0.159 | 40 / 21 | 0.175 / 0.143 | 0.538 / 0.750 | 33 / 18 |
| 0.6125 | 0.50 | 0.913 / 0.912 | 0.817 / 0.705 | 0.250 / 0.270 | 23 / 12 | 0.174 / 0.083 | 0.308 / 0.250 | 19 / 11 |
| 0.6125 | 0.60 | 0.910 / 0.917 | 0.877 / 0.805 | 0.270 / 0.317 | 13 / 7 | 0.231 / 0.143 | 0.231 / 0.250 | 10 / 6 |
| 0.6125 | 0.70 | 0.898 / 0.892 | 0.925 / 0.917 | 0.300 / 0.397 | 3 / 1 | 0 / 0 | 0 / 0 | 3 / 1 |
| 0.6125 | 0.80-0.99, never | 0.899 / 0.892 | 0.947 / 0.943 | 0.305 / 0.413 | 0 / 0 | n/a | 0 / 0 | 0 / 0 |

Chosen on rounds 1-5 only (max R_assign_res with P_assign >= 0.90 and abstention <= 0.15): t_choice 0.535 with s 0.30 (compose t_cov 0.70), dev P 0.908, R_assign_res 0.718, abstention 0.131. It is the only pair meeting the abstention cap on dev (the next, s 0.40, has dev abstention 0.175, R 0.811). On rounds 6-7 the pair gives P_assign 0.931 (27/29), R_assign_res 0.519 (27/52), abstention 0.127 (8/63), `none` 26 (P 0.115, R 0.750), 23 gold-covered units answered `none`. Reading: removing the spurious `none` answers raises resolved recall sharply, but the cost moves to abstention (the unresolved units were being absorbed by `none`), so the abstention cap and R_assign_res >= 0.80 cannot both hold with this score and these questions. Reproduce: `X2D_DATA=$WT/evaluation/x2d-merge-association-pilot/data/postrun $WT/.tools/venv/bin/python evaluation/x2d-postrun-tuning/sweep_window.py $WT/evaluation/x2d-merge-association-pilot`.

## Escalation to Terra, built on rounds 1-7 (development)

Jev fixed at t_choice 0.535 and covered score threshold s 0.60 (none iff p_true <= 0.40). Every window cluster Jev leaves unresolved or answers `none` goes to `codex exec -m gpt-6.1-terra -c model_reasoning_effort=high` (read-only sandbox, ephemeral, scrubbed env, no config; the CLI reported `model: gpt-6.1-terra`, `reasoning effort: high` on all 54 calls). Terra sees the V-sem `merge.covered` state from `jev -dump-state` (unit excerpt, PR, diff summary, all 114 areas with id, title, summary, sample files) under `terra_prompt_v1.txt` and answers `existing`/`new` JSON. Scoring: `existing` correct iff the area is in the gold acceptable set; `new` correct iff gold is not covered. Files: `escalate.py`, `terra_prompt_v1.txt`, `terra_prompt_v2.txt`, `escalation/` (calls, reports, error lists). Frozen specification for rounds 8-10: `FROZEN-8-10.md`.

Prompt iteration (rounds 1-5 escalated clusters only, 31 clusters): v1 24/31 correct; one revision v2 22/31 (worse), v1 kept. v1 checked once on rounds 6-7: 11/23 correct. Failed Terra calls: 0 of 54. Latency per call: median 5.5 s, p90 7.2 s, max 8.8 s (54 calls, 304 s total, 6 in parallel). Escalations: 31 of 137 (22.6 per 100) in rounds 1-5, 23 of 63 (36.5 per 100) in rounds 6-7, 54 of 200 (27.0 per 100) overall.

Composed results, counts in brackets, Wilson 95 % in square brackets:

| arm | part | P_assign | R_assign_res | unresolved | none P | none R | gold-covered answered none |
|---|---|---|---|---|---|---|---|
| Jev-only | rounds 1-5 | 0.915 [0.846, 0.955] (97/106) | 0.915 [0.846, 0.955] (97/106) | 0.182 (25/137) | 0.333 (2/6) | 0.222 (2/9) | 4 |
| Jev+Terra | rounds 1-5 | 0.899 [0.835, 0.940] (116/129) | 0.906 [0.843, 0.946] (116/128) | 0 (0/137) | 0.625 (5/8) | 0.556 (5/9) | 3 |
| B2 | rounds 1-5 | 0.487 [0.398, 0.577] (57/117) | 0.475 [0.388, 0.564] (57/120) | 0.073 (10/137) | 0 (0/10) | 0 (0/9) | 10 |
| Jev-only | rounds 6-7 | 0.875 [0.739, 0.945] (35/40) | 0.795 [0.655, 0.888] (35/44) | 0.254 (16/63) | 0.143 (1/7) | 0.250 (1/4) | 6 |
| Jev+Terra | rounds 6-7 | 0.898 [0.782, 0.956] (44/49) | 0.746 [0.622, 0.839] (44/59) | 0 (0/63) | 0.143 (2/14) | 0.500 (2/4) | 12 |
| B2 | rounds 6-7 | 0.350 [0.221, 0.505] (14/40) | 0.286 [0.178, 0.424] (14/49) | 0.175 (11/63) | 0.083 (1/12) | 0.250 (1/4) | 11 |
| Jev-only | rounds 1-7 | 0.904 [0.845, 0.942] (132/146) | 0.880 [0.818, 0.923] (132/150) | 0.205 (41/200) | 0.231 (3/13) | 0.231 (3/13) | 10 |
| Jev+Terra | rounds 1-7 | 0.899 [0.846, 0.935] (160/178) | 0.856 [0.798, 0.899] (160/187) | 0 (0/200) | 0.318 (7/22) | 0.538 (7/13) | 15 |
| B2 | rounds 1-7 | 0.452 [0.376, 0.530] (71/157) | 0.420 [0.348, 0.495] (71/169) | 0.105 (21/200) | 0.045 (1/22) | 0.077 (1/13) | 21 |

Jev+Terra's R_assign_res equals R_assign (nothing abstains). Reading: Terra removes all abstention and raises R_assign (abstention = miss) from 0.706 to 0.856 over rounds 1-7 at about the same P_assign (0.90), but it is weaker on rounds 6-7. Twelve of the 14 `none` answers in rounds 6-7 are gold-covered clusters of one round-7 PR (streaming and server-sent-events tests and examples): Terra reads "no listed area covers streaming responses" while the labellers accept "Custom responses" / "Response classes and status behavior". Whether such a change is a new area or covered is where Terra and the gold differ.

Reproduce (`X2D_DATA=$WT/evaluation/x2d-merge-association-pilot/data/postrun`):
```
$WT/.tools/venv/bin/python evaluation/x2d-postrun-tuning/escalate.py $WT/evaluation/x2d-merge-association-pilot --prompt evaluation/x2d-postrun-tuning/terra_prompt_v1.txt --tag v1-r1-5 --rounds 1-5
$WT/.tools/venv/bin/python evaluation/x2d-postrun-tuning/escalate.py $WT/evaluation/x2d-merge-association-pilot --prompt evaluation/x2d-postrun-tuning/terra_prompt_v1.txt --tag v1-r6-7 --rounds 6-7
```
