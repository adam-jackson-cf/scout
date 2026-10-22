# E2 — Can Jev triage deterministic analyzer signals? (route-only)

Date 2026-10-17. All numbers are measured from files in this directory (`raw/*.jsonl`, `metrics.json`, `llm_metrics.json`, `unlabelled_metrics.json`). n = 22 labelled cases (9 real_defect = `validated_true`, 13 `false_positive`), so every figure carries a wide interval.

## Method
- **Cases** (`cases.jsonl`, built by `build_cases.py`): one per labelled issue in `reference-v1.json`. State = `{tool_signal, path, excerpts[]}`; excerpts are numbered source lines from `git show <rev>:<path>` at the entry's pinned revision (cited lines ± 15–25, plus the other referenced file(s)). The case id (which contains words like `false-positive`), `behavior`, `evidence` and mechanism text are **never** in the state.
- **De-leak** (per-case note in `cases.jsonl` field `deleak`): the tool signal is rewritten as "<analyzer> rule <ID>: <neutral claim the rule makes>" taken from the provenance text; removed from every signal: verdict/reasoning words (e.g. "loopback-only", "generated migration", "test fixture", "exhaustive switch"), CVE/GHSA ids, fix commits, patched versions, permission commentary. Source excerpts are raw file lines. The file *path* is kept (a real analyzer input; it legitimately reveals `tests/`, `Migrations/`, `.github/workflows/`).
- **Split**: seeded (`random.seed(7)`), stratified: held-out = 5 real + 7 FP (12); tuning = 4 real + 6 FP (10).
- **Jev**: `POST https://openrouter.ai/api/alpha/decisions`, key from Infisical (presence-checked only, never printed/stored). Both `typesafe/jev-1.13-20260917` and `typesafe/jev-1.14-20261014` were reachable; every response's resolved `model` equalled the requested one. 3 repetitions per case per variant per model. Billed `usage.cost` and latency recorded per call in `raw/`.
- **Variants** (`questions.py`):
  - `v1_choice` — Choice {real_defect, false_positive, cannot_tell}, literal criteria with boundary cases. Written before any run.
  - `v2_noul` — Noul "does the excerpt show the flagged mechanism occurring in shipped, hand-written, non-test code with no visible guard?" Written before any run. Routing band applied by me: p ≥ 0.7 → real_defect, p ≤ 0.3 → false_positive, else cannot_tell.
  - `v3_choice` — revision of v1 after seeing **v1/v2 tuning-split errors only**; adds boundary examples (mutable action tags, vulnerable dependency range, loopback literal only formatted, generated/declarative schema, repetitive declarative metric). Held-out cases never informed its text, **but** see leakage caveat 1.
- **Metrics** (`analyze.py`): case-level label = majority of 3 reps (no majority → cannot_tell); accuracy counts cannot_tell as wrong ("acc_all"), plus coverage and accuracy on decided; AUROC on mean P(real_defect) (Choice: `probabilities.real_defect`; Noul: `noul`); rep agreement = all 3 reps gave the same label; Wilson 95% intervals.

## Jev results (case-level; 3 reps each)
Cost per 1k decisions is billed (`usage.cost`): ≈ **$0.049–0.054 / 1k** for every variant/model (~$5e-5 per decision). Latency p50 ≈ 275–310 ms, p95 ≈ 360–540 ms (4 concurrent calls). 0 failed calls out of 396 labelled-set calls (6 files × 66).

| model | variant | split | acc_all (Wilson) | coverage | acc on decided | AUROC | rep agree |
|---|---|---|---|---|---|---|---|
| 1.13 | v1_choice | all | 17/22 [.57,.90] | 21/22 | 17/21 | .957 | 21/22 |
| 1.13 | v1_choice | tuning | 7/10 [.40,.89] | 9/10 | 7/9 | .958 | 10/10 |
| 1.13 | v1_choice | **held-out** | 10/12 [.55,.95] | 12/12 | 10/12 | .971 | 11/12 |
| 1.14 | v1_choice | all | 17/22 [.57,.90] | 21/22 | 17/21 | .940 | 21/22 |
| 1.14 | v1_choice | tuning | 7/10 | 9/10 | 7/9 | .917 | 9/10 |
| 1.14 | v1_choice | **held-out** | 10/12 [.55,.95] | 12/12 | 10/12 | .943 | 12/12 |
| 1.13 | v2_noul | all | 14/22 [.43,.80] | 16/22 | 14/16 | .932 | 18/22 |
| 1.13 | v2_noul | tuning | 7/10 | 8/10 | 7/8 | .917 | 9/10 |
| 1.13 | v2_noul | **held-out** | 7/12 [.32,.81] | 8/12 | 7/8 | 1.000 | 9/12 |
| 1.14 | v2_noul | all | 16/22 [.52,.87] | 18/22 | 16/18 | .949 | 21/22 |
| 1.14 | v2_noul | tuning | 7/10 | 8/10 | 7/8 | .917 | 9/10 |
| 1.14 | v2_noul | **held-out** | 9/12 [.47,.91] | 10/12 | 9/10 | .971 | 12/12 |
| 1.13 | v3_choice | all | 21/22 [.78,.99] | 22/22 | 21/22 | .991 | 21/22 |
| 1.13 | v3_choice | tuning (tuned on) | 10/10 [.72,1.0] | 10/10 | 10/10 | 1.000 | 10/10 |
| 1.13 | v3_choice | **held-out** | 11/12 [.65,.99] | 12/12 | 11/12 | .971 | 11/12 |
| 1.14 | v3_choice | all | 21/22 [.78,.99] | 22/22 | 21/22 | .991 | 21/22 |
| 1.14 | v3_choice | tuning (tuned on) | 10/10 | 10/10 | 10/10 | 1.000 | 9/10 |
| 1.14 | v3_choice | **held-out** | 11/12 [.65,.99] | 12/12 | 11/12 | .971 | 12/12 |

Confusions (all 22, case-level, rows = label → prediction):
- v1 (both models): FP→FP 9, FP→real 3, FP→cannot_tell 1, real→real 8, real→FP 1. Errors: python-multipart range (real→FP, the only dependency-advisory case), nest loopback URL, nest route-params complexity, axum documented-query duplication (FP→real), eShop generated-migration architecture (FP→cannot_tell on 1.13).
- v2 1.13: FP→cannot_tell 5, FP→real 2, real→cannot_tell 1; v2 1.14: FP→cannot_tell 2, real→cannot_tell 2, FP→real 2. Noul is more conservative and routes more to cannot_tell.
- v3 (both models): FP→FP 12, FP→real 1 (axum documented-query duplication), real→real 9.

**1.13 vs 1.14**: indistinguishable at this n (identical v1/v3 accuracy; v2 1.14 slightly better, difference within noise). Same price.

## Routing rate / confidence
`cannot_tell` rate: v1 1/22 (1.13) / 1/22 (1.14); v2 6/22 / 4/22; v3 0/22. A route-only use does not need `cannot_tell`; the more useful signal is P(real_defect), where AUROC is .93–.99 across variants.

## Unlabelled sample (descriptive only; no ground truth)
30 `pending_review`/`incomplete` observations from `scout-requests-v2.json` (stratified round-robin over repo×source, seed 11; `unlabelled.jsonl`), 1.14, 3 reps (90 calls each), pooled rep labels:
- v1_choice: FP 42, real 45, cannot_tell 3. By source: gitleaks generic-api-key 27 FP/3 real; mutable-action-ref 27 real/15 FP; ruff PERF 9 real; ruff C901 6 real/3 cannot_tell.
- v3_choice: FP 27, real 63 (gitleaks 27 FP/3 real; mutable-action 42 real; ruff 18 real) — v3 never says cannot_tell and calls all ruff/action findings real.
- v2_noul: FP 35, real 11, cannot_tell 44 (mutable-action: 34 cannot_tell/8 FP).
Observation: v1 labelled 15/42 mutable-action reps "false_positive" where the adjudicated set has 0 FPs of this family (6/6 real) — [INFERENCE] some of those are inconsistent routing rather than real discrimination; no labels exist to verify. Gitleaks → mostly FP is plausible for docs/test files but unverified.

## LLM baseline
`run_llm.py`: `claude -p --tools ""` (no tool use, no session persistence, cwd /tmp), same instruction + criteria + de-leaked state as the Jev variants, asked for a one-sentence reason plus the 3-way label. Resolved model per `modelUsage`: `claude-opus-5-5`. 1 rep, 22 calls each.

| variant text | acc_all (Wilson) | tuning | held-out | cost / 1k | latency p50 / p95 |
|---|---|---|---|---|---|
| v1 criteria | 21/22 [.78,.99] | 9/10 | 12/12 [.76,1.0] | ≈ $64.8 (total $1.43) | 8.2 s / 10.4 s |
| v3 criteria | 20/22 [.72,.98] | 9/10 | 11/12 | ≈ $54.9 | 9.1 s / 12.0 s |
Both miss the python-multipart dependency case (v3 text additionally misses webpack DOM-clobbering). Cost includes Claude Code CLI system-prompt overhead, so it overstates a lean API call; [INFERENCE] a bare API call would still be well above Jev.

Versus Jev on the same cases: LLM v1 21/22 vs Jev v1 17/22 and Jev v3 21/22 (the latter benefits from criteria written with label knowledge). Jev is ≈ 1000× cheaper ($0.05 vs ≈ $55–65 per 1k) and ≈ 25–30× faster (0.3 s vs 8–9 s).

## Caveats
1. **Leakage / tuning contamination**: I read every adjudication's mechanism text while building the cases (needed to find the right source lines), so the v1/v2/v3 criteria and boundary examples (CI mutable tags, test/fixture/generated artifacts) reflect the label classes. The held-out split is held out from *iteration on results*, not from my knowledge. Treat v3 held-out 11/12 as optimistic.
2. **n = 22**, 12 held-out; Wilson intervals span ~±25 points. 1.13 vs 1.14 or v1 vs v3 differences are not statistically meaningful.
3. **Label mix is narrow**: 6 of 9 positives are the same "mutable GitHub Action tag" pattern, which any rule engine already identifies; the interesting negatives (generated/test/duplication) are mostly "file-type" signals visible in the path. The set under-tests subtle semantic cases (only the multipart dependency and DOM-clobbering are non-trivial positives; Jev v1 missed multipart, LLM missed multipart too).
4. Excerpts are bounded windows; a few signals (lizard long-function) reference code larger than the excerpt.
5. Labels are one adjudicator's judgement ("hand-adjudicated"), not independently double-labelled.
6. Choice `confidence` and Noul thresholds were not calibrated; routing band 0.3/0.7 was fixed a priori, not tuned.

## Recommendation (route-only)
- Jev is cheap and fast enough (~$0.05/1k, ~0.3 s) to run on every deterministic signal as a **triage prioritiser**: use `P(real_defect)` (AUROC ≈ .93–.99 here) to order the review queue and `false_positive`/`cannot_tell` to lower visibility, **never to close, dismiss or suppress**. Every finding stays persisted with its deterministic signal; Jev output is stored as an advisory `triage` annotation (variant version, model, probabilities, run id) that a human or Scout's own disposition can override.
- Prefer a Choice with explicit criteria over Noul (Noul was more conservative and noisier here: 6/22 cannot_tell on 1.13 and weaker agreement). Persist the question-variant version with each decision; re-measure on any model bump (1.14 behaved the same as 1.13 here).
- The single-model reasoning LLM is not needed for volume triage (≈ 1000× cost); keep it as an optional second opinion on the highest-severity or Jev-uncertain signals only.
- Before relying on it: build a larger, de-correlated labelled set (≥ 100, many more semantic cases), keep a blind hold-out authored by someone who has not seen the criteria, and check the over-confident "real_defect" tendency (v3 called 63/90 unlabelled reps real, v1 called ~half of mutable-action reps FP).

## Reproduce
`python3 build_cases.py && python3 build_unlabelled.py`; Jev runs need `OPENROUTER_API_KEY` (e.g. `cd ~/.omp/jev-lab/fastino-bench && infisical run --silent -- python3 <abs>/run_jev.py cases.jsonl raw/jev_<v>_<m>.jsonl typesafe/jev-<model> <variant> 3`; note `infisical run` from the scout dir fails — no project configured there); then `python3 analyze.py raw/jev_*.jsonl`, `python3 run_llm.py <variant> raw/llm_<variant>.jsonl`, `python3 analyze_llm.py raw/llm_*.jsonl`, `python3 unlabelled_dist.py raw/unlabelled_*.jsonl`.
