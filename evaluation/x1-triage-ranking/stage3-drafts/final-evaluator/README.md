# Final evaluator — Option G draft

**DRAFT, synthetic only. No authority until this artefact's scoped council review has zero issues.** It deliberately cannot emit `binding_gate_established`; successful synthetic fixtures emit `draft_synthetic_gate_passed`. No preparation action (a)–(e), models, scans or real-repository data was used.

Implements PREREGISTRATION.md lines **127–139, 153–177, 181–185, 206, 224–228**, round 8's resource/safety requirements and round 9 readings 1–2. Signed bytes are never rewritten.

- `evaluator.py`: exact manifest identity; enumerates all 18,172 binding cells × five populations (90,860 records); rejects absent, duplicate or extra identities; requires B=1,999, 2,000 eligible per population, both hard caps and reconciled gate/inadequacy counters. Applies both pooled Wilson tail upper ≤.035, calibrated realised null/power, null false-support upper ≤.05, power support lower ≥.80 and inadequacy upper ≤.05, missing-mask reachability, fixed-eight-worker CPU/memory/wall fit. Largest registered M≤Mhat is selected without skipping a failing row. Safety does not participate. Ordered §7 verdict boundaries preserve veto/integrity/negative/unresolved precedence.
- `provenance.py`: exact byte-digest bundle membership against **external** trust anchors (not an input approval flag); complete-run maximum inputs; recorded host identity; rerun classification and evidence refusal. Completed closure-mismatched runs still enter maxima. Resource-exhaustion/host-load/operator termination cannot buy a rerun. The final host cannot change.
- `test_evaluator.py`: full-grid planted effect/null summaries; missing/duplicate/B/cap mutants; wrong realised/unreachable references; resource limits; no lower-row bypass; ordered verdict boundaries; substituted input digests and completed-closure-failure maxima.
- `check_frozen.py`: reproduces the manifest/baseline/owned-extract checks below. It only reads existing signed protocol files and Git objects, never candidate repositories.

The trust-anchor checker establishes byte integrity, **not** council authenticity. In real use, authentic externally reviewed source/locked-input/stage records are mandatory; JSON approvals or a bundle's self-declared hashes cannot create authority. The synthetic guard refuses all non-synthetic evaluation. This draft therefore cannot be mistaken for the separately reviewed authoritative evaluator.

## Open choices for review

1. Keep this pre-clock executable synthetic-only and its status non-authoritative; the scoped review must bind the source, authentic locked input and trust anchors before any authoritative deployment. This is deliberately stricter than accepting self-declared stage flags.
2. Check power conditional-inadequacy Wilson upper per population, conservatively, as well as pooled coverage/reference support criteria; no population may hide behind a pooled adequate rate.
3. Supply resource components as measured, pre-overhead CPU seconds; apply ×2 once. Supply cumulative OC wall days already charged with overhead and max-phase lifetime high-water RSS bytes. Exact component maxima from every completed remeasurement must match the evaluator inputs. No planning slope/intercept substitutes for locked RSS.
4. Represent records with explicit canonical cell dictionaries, zero-based population indices and separate gate/adequacy counters. The concrete dictionary serialization and external provenance bundle membership need scoped review; they are not new study dimensions.

## Synthetic validation

From repository root:

```sh
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/test_evaluator.py
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/final-evaluator/check_frozen.py
```

Results: **6 tests passed** (CPython 3.11.11). Initial validation exposed an invalid synthetic test interval (lower .20 > upper .05); the fixture was corrected, not the decision rule. Planted summaries are deterministic behavioral fixtures, not claimed outer-OC evidence. No full OC, labels, real provenance attestation or product-backed rule-4 harness ran. Baseline unavailable.

Frozen-byte result: **126/131 literal rev9 file hashes match; 5/5 shared-file owned-extract groups valid; 131/131 files unchanged against baseline e2b94e3c.** Nineteen of 20 owned extracts match literally; HD9's state-only `proposed`→`preregistered` transition is round-9 condition 2 (normalising only that state reproduces the signed hash). The five shared files already changed at baseline; none is edited here. Manifest SHA-256 remains `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`.
