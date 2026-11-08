# Locked-size input — Option G draft

**DRAFT, synthetic only. No authority until this artefact's scoped council review has zero issues.** Nothing here authorises preparation, inventory, remeasurement or OC. No preparation action (a)–(e), scanner, model or real-repository data was used.

Implements PREREGISTRATION.md lines **153, 161, 163, 179–185, 224**; round 8 X1-R8-01 and round 9 §§6, 8–9. The frozen validation sources are unchanged.

- `size_input.py`: complete ordered integer-size vectors for three synthetic repository slots, with global min/max offsets, canonical SHA-256, keyed deterministic SRS20 and exact size-only gate. No truncation, padding, replication or selected-frame substitution.
- `population.py`: executable construction variant of the signed `prepare_population`, changing only its size parameter, run-loop bound, size lookup and global offset lookup. All other construction/randomisation code is retained, with the original immutable numerical classes/helpers imported.
- `test_size_input.py`: exact original-versus-new population equality for identical vectors; asymmetric complete vectors; global offsets; accepting/rejecting gates; deterministic digest; AST equality outside the four size-source changes.

The input API is `SizeInput({synthetic_repository: complete_size_vector})`; pass it as `prepare_population(size_input=...)`. `measurement_sizes()` exposes every `(run_index, size)` without a hidden L40/L400 subset. This is an input artefact, not permission to run the benchmark/calibration commands. Their eight-cell/2,000-trial/B=1,999/one-two-worker structure must remain unchanged when this source is connected under stage-3 review. No measurements or binding resource fit are claimed here.

## Open choices for review

1. Use the three existing synthetic repository names as positional aliases for the three reviewed locked repositories, preserving the numerical kernel's traversal order. Names are not inferred from inventory data.
2. Keep complete, ordered vectors, including unequal repository lengths, rather than choose prefixes or resample/replicate sizes into L40/L400 templates. Review the concrete benchmark size-source mapping before any locked remeasurement; this artefact does not select a smaller subset from known sizes.
3. Retain the reviewed worst-case 60,000-unit trial fixtures and all benchmark replicates unchanged; the new source replaces independent population construction/calibration inputs, not the scientific trial algorithm.

## Synthetic validation

From repository root:

```sh
PYTHONDONTWRITEBYTECODE=1 .tools/venv/bin/python evaluation/x1-triage-ranking/stage3-drafts/locked-size-input/test_size_input.py
```

Result: **4 tests passed** (CPython 3.11.11). No full benchmark, calibration timing, RSS measurement, inventory or experiment was run. Baseline unavailable; this is deterministic implementation validation, not evidence for HD9.
