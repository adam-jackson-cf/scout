# X2d post-run threshold tuning (development, not evidence)

User direction 2026-10-31 (R-055): tune the X2d Jev thresholds on the labelled data the stopped pilot produced, per the jev-actions skill. This is instrument development after the pilot ended; it moves no hypothesis state and is not a preregistered test.

- Data: every gold-labelled V-sem case with saved scores: round-0 main run (64), round-0 P1 (20), re-gate P1 (40), re-gate P2 (40) = 164 cases. Round-0 P2 excluded (derived gold, shown mis-built). Labels: agent protocol, provenance `agent_reviewed`.
- Method: the pilot's pinned `tune_threshold.py` (sha `2f850933…2e4c`), the pilot's minimums and betas from `thresholds.json`, `--folds 5 --repeats 5 --seed 20261029`, conservative and point-estimate runs. No model calls.
- Reproduce: from the run worktree, `.tools/venv/bin/python evaluation/x2d-postrun-tuning/tune_pool.py $PWD/evaluation/x2d-merge-association-pilot` (copy this directory there first, or pass the scout checkout path of the script); outputs land in `/tmp/x2d-tune/`.

Results:
- `merge.which_area` (t_choice; min P 0.90, R 0.80, beta 0.5): 124 positive / 40 negative. Point estimate: shipped threshold 0.6875; held-out P 0.909, R 0.818 (default 0.5: P 0.836, R 0.944); fold range 0.615–0.6875 (spread > 0.05) and 15 of 25 folds below the minimums held-out, so status `tuned-provisional`. Conservative (95 % lower bounds): no threshold qualifies at this size.
- `merge.covered` (t_cov; min P 0.80, R 0.80): only 16 `none` positives against 148 negatives (below the skill's ~30 per label). No threshold meets the minimums in any fold, even on point estimates (default 0.5: P 1.00, R 0.38). Stays provisional 0.7.
- Effect on abstention (choice step only): re-gate P1 0.47 at 0.70, 0.45 at 0.6875, 0.35 at 0.60; round-0 main 0.23 / 0.20 / 0.12. The tuned value is effectively the starting 0.70, so threshold tuning does not explain the re-gate abstention.
