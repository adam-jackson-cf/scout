# Final evaluator — Option G draft

**DRAFT, synthetic only; no experiment, preparation, model, scan or real-repository authority.** This revision addresses scoped round-1 findings; it is not council sign-off and never emits `binding_gate_established`. Signed protocol sources and `check_frozen.py` (including its baseline exceptions) are unchanged.

## Byte-bound API

`evaluate(manifest_bytes, *, provenance_payloads, provenance_digests, synthetic_only)` evaluates every row **before dev**, without M̂. Each payload is bytes, and every SHA-256 is supplied in an external exact-membership digest map. Authenticity of those external anchors remains outside this synthetic draft. Manifest bytes must equal the bundle member and the signed byte digest, not a reserialized-object hash.

Pre-dev bundle members:

- `signed_manifest`: exact signed `oc-manifest.json` bytes.
- `locked_inventory`: JSON with `eligible_sizes` (three complete repository size vectors, not the selected sample; each vector has 20–2,000 integer sizes, each size 20–2,000, inclusive), `gate_redraws=100000`, integer `gate_successes`, and `p_gate=gate_successes/100000`. The evaluator obtains both sizes and p_gate here, never from parallel arguments. Synthetic fixtures use synthetic repository names and sizes only.
- `measurement_runs`: JSON with `locked_inventory_sha256` and ordered `runs`. Each run has an ID, host, three load averages, concurrent-process count, exact boolean completion flag, integrity-failure cause, evidence hashes where required, and per-operation `binding_inputs` for completed runs.
- `sealed_oc_output`: JSON with `records`, SHA-256 of compact sorted-key JSON encoding of those records, `locked_inventory_sha256`, `signed_manifest_sha256`, and OC `host`. Host keys are `hw.model`, `hw.memsize`, `logical_cpus`, and `os_build`; every remeasurement must match the OC host or evaluation refuses with a deviation pause.
- `source_runtime_closure`, `controller_freeze`, `scoped_council_record`: externally pinned closure/freeze/review bytes (synthetic, non-authoritative stand-ins in these tests).

No parallel records, resources, p_gate, measurement-run or M̂ argument exists. Mutating any payload without changing its external digest is refused. Cross-member inventory, manifest and record seals are checked too.

The pre-dev return contains `result`, canonical `sealed_result` bytes and `result_sha256`. The result includes every row, every cell's pooled and per-population counts/Wilson intervals, resource arithmetic, provenance hashes, and **every run**, including its complete original inputs and classification. If no row passes, status is `not testable before dev`.

`select_postdev(sealed_predev, predev_sha256, *, provenance_payloads, provenance_digests, synthetic_only)` requires the externally sealed pre-dev digest and the identical pre-dev bundle, plus `mhat_evidence`. This JSON binds `Mhat`, positive finite `c_U`, nonempty `bootstrap_evidence`, `predev_sha256`, and `locked_inventory_sha256`. M̂ must be an exact integer satisfying `min(2000, floor(.50/c_U))`; contradictory sealed evidence is refused. It selects the largest registered M ≤ M̂, then consults that row's sealed pass status; it cannot skip a failed nearest row or recompute rows using replacement records. Success remains `draft_synthetic_gate_passed`; failure is `not testable`.

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

`test_evaluator.py` contains full-grid planted effect/null fixtures; exact membership and B mutants; individually reconciled counter caps; both tails in each cell kind; support, null, inadequacy and realised-target last-pass/first-fail boundaries; malformed predicate/gain/measurement mutants; byte and cross-member seal substitutions; per-input maxima from different runs; CPU/p_gate/memory boundaries; rerun/host/evidence cases; pre-dev no-pass, post-dev seal tampering and no-row-bypass cases; ordered typed §7 verdicts including NaN precedence and the .10 upper boundary. These are deterministic synthetic behavioral fixtures, not outer-OC evidence.

Reproduced after integration using the pinned interpreter:

```sh
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/test_evaluator.py -v
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/check_frozen.py
```

**Result: 14 tests passed in 149.162 seconds.** Full output is in `reviews/round-1-test-output.txt`; issue mappings, scratch-only mutation evidence and file digests are in `reviews/round-1-response.md`. An initial fixture attribute named `run` shadowed `unittest.TestCase.run`; the fixture was renamed before the passing run. No real inputs, authentic stage provenance or authority were exercised.
