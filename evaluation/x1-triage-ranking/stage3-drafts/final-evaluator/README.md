# Final evaluator — Option G draft

**DRAFT, synthetic only; no experiment, preparation, model, scan or real-repository authority.** This revision addresses scoped round-1 residual and round-2 findings; it is not council sign-off and never emits `binding_gate_established`. Signed protocol sources and `check_frozen.py` (including its baseline exceptions) are unchanged.

## Byte-bound API

`evaluate(manifest_bytes, *, provenance_payloads, provenance_digests, synthetic_only)` evaluates every row **before dev**, without M̂. Each payload is bytes, and every SHA-256 is supplied in an external exact-membership digest map. Authenticity of those external anchors remains outside this synthetic draft. Manifest bytes must equal the bundle member and the signed byte digest, not a reserialized-object hash.

Pre-dev bundle members:

- `signed_manifest`: exact signed `oc-manifest.json` bytes.
- `locked_inventory`: JSON with `eligible_sizes` (three complete repository size vectors, not the selected sample; each vector has 20–2,000 integer sizes, each size 20–2,000, inclusive), `gate_redraws=100000`, integer `gate_successes`, and `p_gate=gate_successes/100000`. The evaluator obtains both sizes and p_gate here, never from parallel arguments. Synthetic fixtures use synthetic repository names and sizes only.
- `measurement_runs`: JSON with `locked_inventory_sha256` and ordered `runs`. Each run has an ID, host, three load averages, concurrent-process count, exact boolean completion flag, integrity-failure cause, evidence hashes where required, and per-operation `binding_inputs` for completed runs.
- `sealed_oc_output`: JSON with `records`, SHA-256 of the exact UTF-8 JSON value bytes of that top-level member (excluding surrounding separator whitespace), `locked_inventory_sha256`, `signed_manifest_sha256`, and OC `host`. Producers use compact sorted-key JSON as before: its digest is unchanged. The evaluator extracts the member with JSON parsing, not float reserialization, so sealed exponent-overflow gain bytes reach the finite-gain cell checks. Duplicate keys and nonstandard NaN/Infinity literals are refused. Host keys are `hw.model`, `hw.memsize`, `logical_cpus`, and `os_build`; every remeasurement must match the OC host or evaluation refuses with a deviation pause.
- `source_runtime_closure`, `controller_freeze`, `scoped_council_record`: externally pinned closure/freeze/review bytes (synthetic, non-authoritative stand-ins in these tests).

No parallel records, resources, p_gate, measurement-run or M̂ argument exists. Mutating any payload without changing its external digest is refused. Cross-member inventory, manifest and record seals are checked too.

The pre-dev return contains `result`, canonical `sealed_result` bytes and `result_sha256`. The result includes every row, every cell's pooled and per-population counts/Wilson intervals, resource arithmetic, provenance hashes, and **every run**, including its complete original inputs and classification. If no row passes, status is `not testable before dev`.

`select_postdev(sealed_predev, predev_sha256, *, provenance_payloads, provenance_digests, synthetic_only)` requires the externally sealed pre-dev digest and the identical pre-dev bundle, plus `mhat_evidence`. This JSON binds `Mhat`, positive finite `c_U`, typed complete `bootstrap_evidence`, `predev_sha256`, and `locked_inventory_sha256`. The bootstrap object has exactly these fields: `bootstrap_means` (int, 10000), `bootstrap_seed` (int, 169300), `resampling_unit` (str, `"run-cluster"`), `one_sided_level` (float, .95), and `unknown_or_incomplete_receipts` (int, 0). Missing/extra fields, wrong values/types and booleans are refused; this declaration does not recompute c_U from receipts (the later dev-gate stage owns that). M̂ must be an exact integer satisfying `min(2000, floor(.50/c_U))`; contradictory sealed evidence is refused. It selects the largest registered M ≤ M̂, then consults that row's sealed pass status; it cannot skip a failed nearest row or recompute rows using replacement records. Success remains `draft_synthetic_gate_passed`; failure is `not testable`.

## Arithmetic and identities

`provenance.INPUTS` specifies the measured maxima: construction/calibration/requested-base CPU per unit, worst-case enumeration CPU per calibrated population, gate CPU per check, complete/full512 trial CPU per execution, and construction/calibration/trial-plus-helper lifetime high-water RSS. Every completed run contributes each input's maximum, including completed closure-invalid runs. Only a documented qualifying integrity failure permits a rerun; unfinished failures cannot remain unresolved. Reports retain IDs, causes, completion flags, host/load evidence and inputs.

`resource_fit` computes the signed rev8 binding expression using locked units, manifest counts, 4,000 executions, `ceil(4000/p_gate)` gate checks, calibration binding fraction 2044/2380, and ×2 **once**. Wall days are derived as binding core-hours/192, never supplied. Workers are fixed at eight; memory is eight times the largest measured RSS ≤ host memory − 8 GiB. CPU and wall caps are inclusive at 2,880 core-hours and 15 days. No planning slope/intercept or supplied whole-run totals replace locked measured maxima.

The stage-3 generator **must import `expected_cells` and `identity` from `evaluator.py`**, rather than defining another binding identity schema. Compact sorted-key cell JSON is the seed-message identity. Populations are zero-based; gate rejections and inadequate executions have separate reconciled counters. Binding counts are 16,128 requested coverage, 1,792 calibrated reference, and 252 missing cells, with five populations each. Safety is not part of the gate.

Known pins computed directly from existing bytes / the canonical identity construction (not new experiment results):

| Basis | SHA-256 |
|---|---|
| Signed `oc-manifest.json` bytes | `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d` |
| Signed `validation/reproduce_reports.py` resource arithmetic | `023e2e5bc6ca1ee282e4ebc75ef6b3bccdb96111d0f218c0341189d6a400168e` |
| Compact sorted-key JSON array of sorted canonical identity strings | `3499e423b90df60972b1003c4fc8b65f77a03eada73d67afafa1288e8a43dda2` |

## Regression coverage and reproduction

`test_evaluator.py` contains full-grid planted effect/null fixtures; exact membership and B mutants; individually reconciled counter caps; both tails in each cell kind; support, null, inadequacy and realised-target last-pass/first-fail boundaries; malformed predicate/gain/measurement mutants; byte and cross-member seal substitutions; per-input maxima from different runs; CPU/p_gate/memory boundaries and resource failures propagating to every row; complete host/load provenance; rerun/evidence cases; pre-dev no-pass, post-dev seal tampering and no-row-bypass cases; typed bootstrap declarations; ordered typed §7 verdicts including NaN precedence, reversed intervals and both sides of the .10 point cutoff. These are deterministic synthetic behavioral fixtures, not outer-OC evidence.

Round-2 edits have not been verified during implementation. Parent integration commands, from the worktree root:

```sh
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/test_evaluator.py -v
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/check_frozen.py
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/mutate_round2.py
```

Historical round-1 evidence (not evidence for these changes): 14 tests passed in 149.162 seconds, recorded in `reviews/round-1-test-output.txt`; mappings and earlier mutation evidence are in `reviews/round-1-response.md`. `mutate_round2.py` creates independent temporary scratch trees containing only the evaluator, provenance, tests and the signed fixture files they read; it never changes source files or old review records. It runs a focused behavioral test for each requested guard mutant, requires at least one assertion failure (not solely an import/runtime error), prints the output and baseline file hashes, and exits nonzero if a mutant survives. No real inputs, authentic stage provenance or authority are exercised.

Round-2 observed validation: **20 tests passed in 256.070 seconds; 10/10 scratch guard mutants were caught by assertion failures** (the host-guard removal additionally produced downstream missing-field/type errors). Full outputs and issue-to-line/test mappings are committed as `reviews/round-2-test-output.txt`, `reviews/round-2-mutant-output.txt` and `reviews/round-2-response.md`. These are synthetic validation results, not stage authority.
