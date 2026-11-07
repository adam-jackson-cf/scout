# Terra replay provenance

Recorded fastapi rounds 8–10 only. Source of answers: frozen-r8-10.calls.jsonl; Jev rows and gold are composed with the frozen pilot modules (covered score 0.60, choice 0.535). No calls were made. The slice retains complete composed rows, gold and recorded answers; no outcome was edited.

Reproduction: pinned Python /tmp/a4-reproduce-pilot.py imports escalate.py, cw.load, jev_rows, final_row and score without running main or writing evaluation files. Observed: 92 units, 77 assigned, 71 correct assignments, zero unresolved, 15 new, 13 correct new. Product replay raises pending cap to preserve all new candidates.

SHA-256 sources:
- `evaluation/x2d-postrun-tuning/escalation/frozen-r8-10.calls.jsonl`: `84c5e399098e3fca2f78718eea1cf42ab05e5980a50f6a48d5d0b5e8967a76ac`
- `evaluation/x2d-postrun-tuning/escalate.py`: `29109416640c0576451f339b5cd350680c116c50cf446584c17a611295dad70c`
- `evaluation/x2d-postrun-tuning/terra_prompt_v1.txt`: `ca24fd764b745dbcb261a1aa6a0d740710809e82a13a8a103e55bbc4d8555e76`
- `evaluation/x2d-merge-association-pilot/data/postrun/rounds/8/units.jsonl`: `05be7fa024903d0a1e296e55005dbdd9a95dc6d6d8df55ccb17ec05008f0c6ff`
- `evaluation/x2d-merge-association-pilot/data/postrun/rounds/8/gold.jsonl`: `24516d64f6cb1e31eb9a59f40d21cd4611a99d5a698386e5f6442871179ae5a4`
- `evaluation/x2d-merge-association-pilot/data/postrun/rounds/8/jev/scores.jsonl`: `d86ac7fb4abc6363aa908deda9930cd4479f9421e84d507b95d21a842271928b`
- `evaluation/x2d-merge-association-pilot/data/postrun/rounds/9/units.jsonl`: `736e3127d48449958c06935c408d37091d52317993051f6e9ae436319a895e86`
- `evaluation/x2d-merge-association-pilot/data/postrun/rounds/9/gold.jsonl`: `4c8526dd714d0401ad28086034489545de585a722844674d0555c5d5a7adcfa9`
- `evaluation/x2d-merge-association-pilot/data/postrun/rounds/9/jev/scores.jsonl`: `51f4c5802d69ca3cb4709d944088cfb1f3f67b62b7235ae5436f4809b816ac19`
- `evaluation/x2d-merge-association-pilot/data/postrun/rounds/10/units.jsonl`: `d2e866d99b7fc075f64d885b60ea0817e8f5817002af3128de5063aabd1b3e8b`
- `evaluation/x2d-merge-association-pilot/data/postrun/rounds/10/gold.jsonl`: `1157c088fc22f7b5a612075f4844e90cb5d289b08f31f67701d4e2d62f088630`
- `evaluation/x2d-merge-association-pilot/data/postrun/rounds/10/jev/scores.jsonl`: `147118b71fa262eb2a8d9e95018bd2f1d003a3aa2dbdd9a3dfe179aeee5e61f2`

Slice SHA-256: `5dc49bf721a41504a0072d7cd9b502ded64cb32915cf2e83089e5c028da3ae1e`.
