# CC2 identifier delta implementation round 1 — response

R1-01 to R1-03 are fixed with catching tests. This response is not sign-off. No real delta producer run occurred. No `preparation.py`, `verify_inputs.py`, config, baseline output or previous review record changed.

Paths below are relative to `evaluation/x1-triage-ranking/`.

## Required-fix mapping

| Issue | Fix | Catching tests |
|---|---|---|
| R1-01 | `execution/identifiers/cc2_delta.py:235–236` adds exact live pins for `identifiers/projection.py`, `preparation.py` and `verify_inputs.py` to `implementation`. Thus `:238` includes them in the immutable journal source pins, and `:243–244` includes them in the review gate's required binding set. `preparation-reviews/identifiers-delta-cc2-REQUEST.md:53` explicitly names all three imported dependencies and requires renewed review after any change. | `execution/tests/test_cc2_delta.py:258–267`, `test_changed_imported_dependency_refuses_review_gate`, plants a post-review byte change for each dependency in a temporary fixture and requires refusal. `:334–359`, `test_synthetic_journalled_run_and_refusal_keep_baseline`, independently asserts each actual imported helper pin is present in **both** journal records and in **both** review-gate calls. Omitting any dependency from either binding set fails the test. |
| R1-02 | `execution/identifiers/cc2_delta.py:176–182` registers each JSON filename immediately after its exclusive open, before any serialization. `:198–200` does the same for the familiarity log before writing its first line. `:240,275,278–279` maintains an invocation-local created-file list and uses only that list for terminal output pins. Existing-root refusal therefore records zero outputs; partial current-invocation files remain pinned. The no-op publish exception/rethrow and unused shared `exclusive_json` import were removed; core was not edited. Request `:73` states the corrected provenance rule. | `test_synthetic_journalled_run_and_refusal_keep_baseline` (`tests/test_cc2_delta.py:334–359`) now expects 3 output pins on success and **0** on the refused rerun, retaining exact baseline hashes. `test_partial_publication_retained` (`:232–245`) plants a failure inside coverage JSON serialization and asserts both newly created names are registered, including the partial coverage bytes. `test_journal_pins_partial_familiarity_from_current_run` (`:361–382`) plants a failure after the familiarity header; the refused journal must pin both JSON files and the partial familiarity log with its actual nonzero byte count. |
| R1-03 | Computed the existing baseline journal's file SHA-256 and byte count and supplied them to Main. Request `:55–65` now binds the exact file digest/size, start-record digest/size and five ordered output pins. The sentence deferring that binding to the next output review was removed. The next implementation sign-off must name the journal digest before execution. | `test_request_names_exact_baseline_journal_pin` (`tests/test_cc2_delta.py:269–275`) asserts the request contains the exact journal filename, 10,301-byte size, full journal digest and start digest, and contains no old deferral sentence. The synthetic journalled-run test (`:353–359`) asserts the **journal file's own pin**, including bytes and digest, reaches each required review binding set; checking only its filename or five output pins cannot pass. |

## Baseline metadata actually observed

This was a metadata-only read of the existing preparation journal, not a delta run or a raw-study read:

* `preparation-run/journal/906389d7212a401b8c7f781d37960d84.json`: **10,301 bytes**, SHA-256 **`6aa8802a23ac0de75eea9a78de5e6422e8cb16c4762f875dc7ebf0787d77dc6c`**.
* action `d`, status `passed`, exit code 0.
* start pin: `906389d7212a401b8c7f781d37960d84.start.json`, 8,524 bytes, SHA-256 `c400e46e3f2ba85767165257b425a9719a223303b71cf78700d3111ae06f8738`.
* Five output pins match the ordered list in the request. Their values were read from this journal's metadata, not recomputed by parsing identifier payloads.

Command, from `/Users/adamjackson/Projects/work/scout-x1prep`:

```sh
shasum -a 256 evaluation/x1-triage-ranking/preparation-run/journal/906389d7212a401b8c7f781d37960d84.json && wc -c evaluation/x1-triage-ranking/preparation-run/journal/906389d7212a401b8c7f781d37960d84.json
```

Complete process output:

```text
6aa8802a23ac0de75eea9a78de5e6422e8cb16c4762f875dc7ebf0787d77dc6c  evaluation/x1-triage-ranking/preparation-run/journal/906389d7212a401b8c7f781d37960d84.json
10301 evaluation/x1-triage-ranking/preparation-run/journal/906389d7212a401b8c7f781d37960d84.json
```

Exit code 0.

## Assertion changes disclosed

* Replaced the incorrect assertion that **both** the successful run and refused rerun have three output pins with separate exact expectations: success has three; rerun has an empty list. Existing action/status and unchanged-baseline checks remain.
* Replaced a failure planted **before** coverage creation (previously expecting coverage absent) with failure **during** coverage serialization. The new oracle requires the registered partial file and exact retained bytes. This strengthens the partial-current-invocation safeguard rather than removing it.
* Factored synthetic setup/context into helpers to reuse it for the partial-familiarity run. No source/schema/projection assertions were removed.
* Added dependency-mutation, both binding-set, baseline-journal-file-pin and request-metadata assertions.
* Added the partial familiarity journal oracle; it validates all three actual current-invocation pins and the single retained source-table line.

The request also corrects its stale integration sentence: the reviewed `preparation.py` at `b1759ec7…` accounts for delta files as `journal-output`; separately reviewed integration must still enforce the review chain/current source closure. This is documentary alignment with the round-1 finding, not a core edit.

## Full-suite command and complete output

Working directory: `/Users/adamjackson/Projects/work/scout-x1prep`.

```sh
.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py'
```

Complete process output:

```text
................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 240 tests in 24.453s

OK
```

Exit code 0; tool-reported wall time 25.03 seconds. All execution Python tests passed, including the new CC2 safeguard tests. The real producer, model calls, network fetches and candidate scans were not run. No Go source changed and no Go check was run in this assignment.

## Recomputed SHA-256s

Command:

```sh
shasum -a 256 evaluation/x1-triage-ranking/execution/identifiers/cc2_delta.py evaluation/x1-triage-ranking/execution/identifiers/cc2_readers.py evaluation/x1-triage-ranking/execution/identifiers/cc2_schemas.py evaluation/x1-triage-ranking/execution/tests/test_cc2_delta.py evaluation/x1-triage-ranking/preparation-reviews/identifiers-delta-cc2-REQUEST.md evaluation/x1-triage-ranking/execution/identifiers/projection.py evaluation/x1-triage-ranking/execution/preparation.py evaluation/x1-triage-ranking/execution/verify_inputs.py evaluation/x1-triage-ranking/execution/run-config.json evaluation/x1-triage-ranking/preparation-run/journal/906389d7212a401b8c7f781d37960d84.json
```

Complete process output:

```text
657aeb47ee11cbe4db4fdba3948639cf211f8ff032844cbe6d77a815086c0ea6  evaluation/x1-triage-ranking/execution/identifiers/cc2_delta.py
1ad375a97b43a771245f3f746768dc72814656c1c81c8c4d8793a9c88644144b  evaluation/x1-triage-ranking/execution/identifiers/cc2_readers.py
db845451e7dbefc8f32ea7b50fbe86cdd5c702330f755e39cee345b2cfc5799f  evaluation/x1-triage-ranking/execution/identifiers/cc2_schemas.py
ba96da085fe3cc1ad2d2433c199e0abedacebb6065f52ccea3d8be83441a20c2  evaluation/x1-triage-ranking/execution/tests/test_cc2_delta.py
288073ce3a47d952db958f33b8423bac8b4f4f0fce340ffd2f9a98ebb456ad0c  evaluation/x1-triage-ranking/preparation-reviews/identifiers-delta-cc2-REQUEST.md
2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072  evaluation/x1-triage-ranking/execution/identifiers/projection.py
b1759ec710fa4b71dec1ef91800e84d285c733693b0cb5f4ffd7fe68f147c916  evaluation/x1-triage-ranking/execution/preparation.py
7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310  evaluation/x1-triage-ranking/execution/verify_inputs.py
c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54  evaluation/x1-triage-ranking/execution/run-config.json
6aa8802a23ac0de75eea9a78de5e6422e8cb16c4762f875dc7ebf0787d77dc6c  evaluation/x1-triage-ranking/preparation-run/journal/906389d7212a401b8c7f781d37960d84.json
```

Exit code 0. Only `cc2_delta.py`, `test_cc2_delta.py` and the request changed among these files. The imported helper/config/baseline pins are observations, not edits. This newly written response does not include a circular self-digest.

## Handoff

Regenerate the complete execution-source manifest and obtain a new zero-issue allowlist/implementation round naming the complete binding set, including the baseline journal file digest. Only Main may then run the canonical command in the request. That real run and its output review remain pending; inventory remains unauthorised.
