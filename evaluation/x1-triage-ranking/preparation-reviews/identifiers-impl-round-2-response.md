# Identifier implementation round 2 — response

R2-01 and R2-02 are fixed with catching synthetic tests. This response is not sign-off. No real action-(d) projection was run, and no raw study source, frozen input, `preparation.py` or `run-config.json` was modified.

All relative locations below are under `evaluation/x1-triage-ranking/`.

## Required-fix mapping

### R2-01 — emitted encoding conservation and non-vacuous tests

* `execution/review_identifiers.py:76–114` adds `validate_emission`; `:200` invokes it after grouping and unresolved aggregation. It refuses with exactly `emitted identifier conservation failure` if, per output, flattened emitted group indices plus emitted record-level reject indices do not cover `range(n)` exactly once, or flattened index/value pairs differ from projected rows. Parallel value/index/source-ID lengths must agree. It independently checks the existence, count and expanded ranges of the emitted unattributed-scope aggregate against projected unscoped path/PR indices. The same aggregate checks also bind emitted provenance-gap ranges/counts.
* `execution/tests/test_identifiers.py:116–133` adds the multi-record fixture. Each identifier output has at least two same-group valid records and an interleaved record-level reject. The unattributed output has repeated `old.json` and distinct `other.json` source references, with paths at non-adjacent indices 2 and 4. The fabricated seal summary is updated accordingly.
* `PipelineTests.test_group_conservation_pin_once_and_source_table` (`execution/tests/test_identifiers.py:219–235`) now uses that fixture and asserts the exact two-entry source table, original count conservation and pin-once invariant.
* `assert_multirecord_encoding` (`:237–262`) asserts the exact ordered group reconstruction (including values and source IDs), the complete ordered unresolved entries with exact counts/ranges, and weighted `unresolved_by_reason`: provenance gaps 8, scope gaps 2, and one of each of the three record-rejection reasons.
* `test_successful_publication_groups_and_pin_once` (`:351–372`) publishes only fabricated data. It compares familiarity group lines as a **list**, without collapsing duplicates, with the manifest's groups and the same explicit multi-record reconstruction; it also validates exact emitted unresolved entries and weighted counts. All pin-once checks remain.
* `test_planted_emitted_range_loss_refuses` (`:264–269`) patches `index_ranges` to discard its last range and requires the exact post-encoding refusal. Thus intermediate-list conservation cannot conceal a loss in the emitted aggregation.
* Updated `preparation-reviews/identifiers-REVIEW_REQUEST.md:32` to document the post-encoding safeguards. Review records remain outside `execution/`.

### R2-02 — independent sources digest and canonical input paths

`PipelineTests.test_independent_sources_digest_and_canonical_input_refusals` (`execution/tests/test_identifiers.py:409–434`) supplies the valid 24,254-row fabricated source universe and all other valid prerequisites. It separately:

1. Copies the frozen manifest byte-for-byte to a noncanonical path with the same basename, asserts identical hashes, requires `noncanonical`, and verifies no journal exists.
2. Repeats that check independently for the source manifest.
3. Keeps the correct frozen digest but corrupts only the expected sources digest; requires `signed recipe`, then verifies the exact terminal status list `['refused']` and exact refusal reason. No output directory is created.

The earlier separate runtime-guard test still expects its own two refused journal records. No production guard change was necessary for R2-02; these tests independently exercise both input-path checks and the second signed-digest check.

## Commands and complete output

Working directory: `/Users/adamjackson/Projects/work/scout-x1prep`.

### Required complete execution suite

```sh
.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py'
```

Complete process output:

```text
E...................................................................
======================================================================
ERROR: test_binding (unittest.loader._FailedTest.test_binding)
----------------------------------------------------------------------
ImportError: Failed to import test module: test_binding
Traceback (most recent call last):
  File "/Users/adamjackson/Projects/work/scout/.tools/python/cpython-3.11.11-macos-aarch64-none/lib/python3.11/unittest/loader.py", line 419, in _find_test_path
    module = self._get_module_from_name(name)
             ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
  File "/Users/adamjackson/Projects/work/scout/.tools/python/cpython-3.11.11-macos-aarch64-none/lib/python3.11/unittest/loader.py", line 362, in _get_module_from_name
    __import__(name)
  File "/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/tests/test_binding.py", line 12, in <module>
    from interlock import InterlockFailure, compare, evaluate, profile, stub_source
ImportError: cannot import name 'profile' from 'interlock' (/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/binding/interlock.py)


----------------------------------------------------------------------
Ran 68 tests in 2.842s

FAILED (errors=1)
```

Exit code 1; tool-reported wall time 3.02 seconds. This is an integration blocker to claiming a passing complete suite: sibling binding tests import `profile`, which is absent from their `interlock` module. Main and X1PrepBinding were notified. I did not change their files or rerun this failing suite to confirm it.

### Isolated identifier safeguards

```sh
.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_identifiers.py'
```

Complete process output:

```text
........................
----------------------------------------------------------------------
Ran 24 tests in 0.722s

OK
```

Exit code 0. All identifier safeguard tests pass on fabricated inputs. This is not a real projector run or completeness evidence. X1PrepBinding subsequently confirmed the missing import is an intentional mid-edit state during its R1 fixes; that agent will update its tests and owns the final integrated execution suite plus Go checks requested by Main. I will not rerun the complete suite while those edits are in flight.

### Recomputed SHA-256s

```sh
shasum -a 256 evaluation/x1-triage-ranking/execution/review_identifiers.py evaluation/x1-triage-ranking/execution/identifiers/projection.py evaluation/x1-triage-ranking/execution/tests/test_identifiers.py evaluation/x1-triage-ranking/preparation-reviews/identifiers-REVIEW_REQUEST.md evaluation/x1-triage-ranking/execution/preparation.py evaluation/x1-triage-ranking/execution/run-config.json
```

Complete process output:

```text
adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2  evaluation/x1-triage-ranking/execution/review_identifiers.py
2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072  evaluation/x1-triage-ranking/execution/identifiers/projection.py
955bc57a9a8cd97e0291ed3582cc0d330ed6060186dda7ac38d1cbd6a033ebea  evaluation/x1-triage-ranking/execution/tests/test_identifiers.py
10879ebd6a61029fb88bafd5924af6fb154b7a7571a33c651d6849c0641210f9  evaluation/x1-triage-ranking/preparation-reviews/identifiers-REVIEW_REQUEST.md
1a116d86b98b088c474a0f4e974af9be8299484ac1deee99dba8f3e7123bc9bf  evaluation/x1-triage-ranking/execution/preparation.py
c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54  evaluation/x1-triage-ranking/execution/run-config.json
```

Exit code 0. The last two are observed integration dependencies, not my edits. This newly written response does not contain a circular self-digest. Earlier review/response records are retained unchanged.

## Handoff

Regenerate the execution-source manifest after integration, preserving complete execution-byte coverage. Submit revised digests for round 3; the real projector remains blocked until zero-issue implementation sign-off. The later identifier-output completeness/delta decision remains separate. No new raw schema extraction is authorised.
