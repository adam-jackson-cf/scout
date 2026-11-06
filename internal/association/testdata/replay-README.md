# Frozen replay slices

Source: evaluation/x2d-merge-association-pilot/data/postrun/rounds/{8,9,10}. Copied from the frozen pilot; no pilot file is changed. V-sem only, window population; gold uses covered + acceptable (not primary). Missing/ungolded rows are excluded by pilot counts_for; all 92 retained rows have usable gold. No core_gate filtering beyond the supplied units.window file. Thresholds: covered score .60, t_choice .535, hash-selected second permutation (not flip_all).

Each `replay-r<N>-{units,gold,scores}.jsonl` fixture retains only consumed columns from round N's `units.window.jsonl`, `gold.jsonl` or `jev/scores.jsonl`, respectively. The following paths are relative to the frozen source's `rounds/` directory; SHA-256 values cover full original bytes:

- `8/units.window.jsonl`: `3a669b845ec13d6410dc073269b15a4596d3bb5cf565bf570127eccb33a6c93b`
- `8/gold.jsonl`: `24516d64f6cb1e31eb9a59f40d21cd4611a99d5a698386e5f6442871179ae5a4`
- `8/jev/scores.jsonl`: `d86ac7fb4abc6363aa908deda9930cd4479f9421e84d507b95d21a842271928b`
- `9/units.window.jsonl`: `1e029f778f391c813e27fe2e897b76bbf9542a34ef78bae8d91ea193a9732342`
- `9/gold.jsonl`: `4c8526dd714d0401ad28086034489545de585a722844674d0555c5d5a7adcfa9`
- `9/jev/scores.jsonl`: `51f4c5802d69ca3cb4709d944088cfb1f3f67b62b7235ae5436f4809b816ac19`
- `10/units.window.jsonl`: `9b6c86ae2e2f98759141704f54401e4cbf8c0db22e74dffaa88a985c9bc20098`
- `10/gold.jsonl`: `1157c088fc22f7b5a612075f4844e90cb5d289b08f31f67701d4e2d62f088630`
- `10/jev/scores.jsonl`: `147118b71fa262eb2a8d9e95018bd2f1d003a3aa2dbdd9a3dfe179aeee5e61f2`
