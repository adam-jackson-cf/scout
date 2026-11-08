# D1 triage minimums: exploratory record

**Exploratory, owner-directed, not preregistered, not citable as experiment evidence.** No council sign-off authorised this analysis as an experiment. R-063 and ADR-010 record the owner's decision, not a validation of the production triage question or threshold. The owner directed the analysis (“use the data not costs”) and accepted it after disclosure that it was unpreregistered and not citable. This was an owner-directed deviation from REQ-102 and CLAUDE.md General “Jev labelling” for this analysis only, not precedent or an exception to design rule 3.

## Inputs and construction

- `scores.csv` is the exact 51-row input retained from the session (28 positives, 23 negatives); one `cannot_tell` case is excluded. All rows are `dev`; there is no test split. Model is `typesafe/jev-1.14-20261014`, question-version identifier `e2-v2_noul`, and stratum is repository.
- Read `evaluation/jev-triage-labels-v1/labels.jsonl` in file order. Keep only `provenance: human`, replacing each case's value with its latest human row while preserving first-seen case order. Exclude `cannot_tell`; map `real_defect` to 1 and `false_positive` to 0.
- Combine successful rows from `evaluation/jev-triage-v1/raw/jev_v2_noul_1.14.jsonl` then `unlabelled_v2_noul_1.14.jsonl`. For each case take the arithmetic mean of the three `answers.q.noul` values and round to six decimal places (Python `round(mean, 6)`, serialized as a number).
- Map `case_id` to `repo` using `evaluation/jev-triage-v1/cases.jsonl` then `unlabelled.jsonl`. Emit the header `case_id,label,score,model,question_version,split,stratum`, with each retained human case in the order above and LF line endings. This recipe was reconstructed and matched the retained CSV byte for byte.

SHA-256 of `scores.csv`: `c7679a2da9a30b3b905e4f8be3c81f4eed6ad459e85426db719a356f3818c2d2`.

## Reproduce

The session used the jev-actions skill at `~/.omp/agent/skills/jev-actions/`, specifically `scripts/tune_threshold.py`, SHA-256 `2f8509337d582506203db9d8946a6490a6824921a58328550ad9bf79d7d72e4c`. The script is external to Scout; reproduction requires that exact script revision, Python 3 and the committed CSV. No model calls are made.

From the skill directory, with `scores.csv` referring to this folder's committed CSV, the original report command is:

```sh
python3 -B scripts/tune_threshold.py scores.csv --min-precision 0.01 --min-recall 0.01 --repeats 5 --seed 7 --errors
```

For this record, `tuning.json` was regenerated with the same arguments plus `--json`, passing the absolute CSV path:

```sh
cd ~/.omp/agent/skills/jev-actions
python3 -B scripts/tune_threshold.py /Users/adamjackson/Projects/work/scout-d1/evaluation/d1-triage-minimums-exploratory/scores.csv --min-precision 0.01 --min-recall 0.01 --repeats 5 --seed 7 --errors --json
```

Save stdout as `tuning.json`; SHA-256 of the recorded output is `03402fc8a1d183338ccf58ce7c723584a44f5fd615944b12509f519e0f587f9b`. The default objective is F1 (beta 1), not the production entry's F-beta beta 2. The 0.01 floors were exploratory tuning constraints, not the owner's accepted minimums.

## Report semantics and limits

`tuned_heldout` pools counts across 5 folds × 5 repeats, using each fold's training-selected threshold. Every case contributes five out-of-fold predictions. It reports TP 128, FP 33, FN 12: precision 0.7950310559006211, recall 0.9142857142857143, F1 0.8504983388704318. These figures are pooled across per-fold thresholds (0.385 in 13, 0.26 in 9, 0.24333349999999998 in 3 of 25 fold tests), not measured at one fixed threshold. `shipped_threshold: 0.385` is the script's name for the threshold chosen in-sample on all 51 dev cases; it was **not shipped**. At fixed 0.5, pooled counts are TP 95, FP 10, FN 45: precision 0.9047619047619048, recall 0.6785714285714286, F1 0.7755102040816326. Per-stratum and error rows in the JSON are in-sample at 0.385, not confirmatory results.

The session's reported intervals were descriptive Wilson 95% intervals, calculated separately from the script (which does not print intervals). Divide pooled counts by five before applying Wilson with z=1.96: precision uses successes 25.6 and denominator 32.2 (n≈32.2), recall successes 25.6 and denominator 28. For p=successes/n, centre is `(p + z²/(2n))/(1 + z²/n)` and half-width is `z*sqrt(p*(1-p)/n + z²/(4n²))/(1 + z²/n)`. Recomputed bounds are precision `[0.6280921804368921, 0.89907642473047]`, recall `[0.7549733289187971, 0.9736332737644854]`. These fractional-count intervals do not account for repeated-CV dependence or threshold-selection uncertainty; they are not confirmatory uncertainty intervals and do not treat the 255 predictions as independent cases. Earlier rounded recall bounds 0.76–0.97 are superseded by these explicit recomputed bounds.

There are fewer than 30 cases per label, no untouched test split, unstable fold thresholds (spread >0.05) and weak axum/nest/webpack strata. Axum's in-sample precision is 0.5 (TP 1, FP 1; n=8, one positive). Scores average three Jev calls per case, while production asks one request per unit (REQ-88); averaging reduces score noise, so a threshold tuned on means may not transfer. Pooled out-of-fold precision 0.7950310559006211 is below the agreed 0.80: joint attainability of 0.80 precision and 0.90 recall is not shown, and D1-R must re-check it. Labels were made under E2's `v1_choice` guide, scores under E2 `v2_noul`; neither matches production `triage.real_defect@v1` (data-not-instructions clause, `state.facts`, 8-line/600-byte bounded excerpts). F1 optimisation does not establish the production F2 policy. No readiness conclusion follows.

## Examined-data quarantine and decision

All 52 `evaluation/jev-triage-labels-v1` cases and the E2 `v2_noul` score files are examined data. Any X1 or D1-R preregistration must disclose them, may use them only for development, and must draw confirmatory test data from unexamined cases. Legacy `v1_choice`-guide labels do not count toward D1-R's floor of at least 30 human/`agent_reviewed` labels per label under the frozen current-question guide and blind two-labeller protocol with adjudication.

The owner accepted precision 0.80 / recall 0.90 for `triage.real_defect` only, and separately confirmed misses cost more. Threshold 0.3 stays provisional and ungated. X3/X5/X10 keep their own REQ-98 defaults; D1 remains open for them. D1-R is part of X1 or its own council-signed preregistration and does not change X1's ordinal decision rule.
