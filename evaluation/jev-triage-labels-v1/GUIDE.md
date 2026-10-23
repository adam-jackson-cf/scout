# Label guide `triage-v1_choice/guide-1`

## Purpose
Decide, from a case's `state` alone (`tool_signal`, `path`, `excerpts`), which of three labels describes a static-analyzer finding. Use only the excerpts; do not assume code outside them.

## Question (verbatim)
A static analyzer reported the tool_signal on this file. Using only the excerpts, which describes the finding?

## Label definitions (criteria, verbatim)
- **real_defect**: The excerpt shows the flagged mechanism literally present in shipped, hand-written, non-test code, and nothing visible in the excerpt guards or neutralises it.
- **false_positive**: The excerpt shows the flagged mechanism is absent, guarded, type-only, or happens in test, fixture, example, documentation or generated files, or the metric does not reflect real maintainability or runtime cost.
- **cannot_tell**: The excerpt does not contain enough to decide, for example the flagged code is outside the excerpt.

## Exclusions and tie-break rules (derived only from the criteria above)
- R1 Presence: real_defect needs the flagged mechanism visibly, literally in the excerpt. If the excerpt shows no trace of it, and the excerpt is otherwise sufficient to see it is absent, use false_positive; if the excerpt simply does not cover the flagged location, use cannot_tell.
- R2 Guarded vs unguarded: a visible guard, check, pin, sanitiser or neutralising construct anywhere in the excerpt that addresses the mechanism makes it false_positive. A guard that is only hinted at but not shown does not count.
- R3 Path context: a path that is clearly a test, fixture, example, documentation or generated file (by directory or file name, e.g. `tests/`, `fixtures/`, `examples/`, `docs/`, `*.generated.*`, `*_pb2.py`, `dist/`) makes it false_positive when the mechanism is present only there. Paths that merely resemble these words but are shipped code do not qualify.
- R4 Type-only: an import, reference or declaration used only for types is false_positive for runtime or dependency claims.
- R5 Metrics: for size, duplication, complexity or similar signals, use false_positive only when the excerpt shows the metric is inflated by repetitive declarative content or otherwise does not reflect real maintainability or runtime cost; if the excerpt shows genuinely hand-written logic that drives the metric, real_defect; if the excerpt is too small to judge, cannot_tell.
- R6 Outside the excerpt: if deciding requires code the excerpt does not contain (flagged line, caller, other file, configuration), use cannot_tell.
- R7 Tie-break order: if a rule for false_positive applies (R2, R3, R4, R5), it wins over real_defect. If real_defect conditions are not all visible and no false_positive rule applies, use cannot_tell. Never choose real_defect on plausibility alone.
- R8 Do not use the severity, importance or likelihood of exploit; only the criteria decide.

## Worked examples (synthetic)
1. real_defect. Signal: `hardcoded-password`; path `src/billing/client.py`; excerpt shows `DB_PASSWORD = "hunter2-prod"` at module level, used in `connect(...)`, no environment lookup. Mechanism literally present, shipped code, unguarded (R1).
2. real_defect. Signal: `shell-injection`; path `lib/deploy/run.js`; excerpt shows `exec("rm -rf " + req.query.dir)` with no validation before it. Present, unguarded (R1, R2).
3. false_positive. Signal: `eval-usage`; path `src/plugin/loader.py`; excerpt shows `eval(expr)` directly preceded by `ast.literal_eval`-style whitelist validation that rejects everything but numerals. Guard visible (R2).
4. false_positive. Signal: `hardcoded-secret`; path `tests/fixtures/sample_config.yaml`; excerpt shows `api_key: "dummy-key"`. Fixture path (R3).
5. cannot_tell. Signal: `sql-injection at line 412`; path `app/db/queries.py`; excerpts show only lines 1-40, imports and a docstring. Flagged code is outside the excerpt (R6).
6. cannot_tell. Signal: `unsafe-deserialization`; path `core/cache.py`; excerpt shows a call `load(blob)` whose definition and origin of `blob` are not visible, and no guard or test context. Real_defect conditions not all visible, no false_positive rule applies (R7).
