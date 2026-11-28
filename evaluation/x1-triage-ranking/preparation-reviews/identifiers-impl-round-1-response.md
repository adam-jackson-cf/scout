# Identifier implementation round 1 — response

All five requested implementation fixes are applied. This is a response, not council sign-off. The real action-(d) projector has not been run. No frozen input, candidate repository, `run-config.json` or `preparation.py` was modified.

Read current `execution/preparation.py`, including the live journal/lock and recursive pin APIs, before integration. Its observed SHA-256 is `1a116d86b98b088c474a0f4e974af9be8299484ac1deee99dba8f3e7123bc9bf`.

Paths below are relative to `evaluation/x1-triage-ranking/`.

## Required-fix mapping

| Issue | Fix and code location | Catching synthetic tests |
|---|---|---|
| R1-01 | `execution/identifiers/projection.py:34–37,87–88`: output provenance is a filename/index, never a repeated pin; unattributed-scope issues no longer nest the full row. `execution/review_identifiers.py:96–117`: provenance and scope gaps aggregate once per output, retaining counts and exact index ranges; schema/structure exceptions remain individual. `:128–158`: group by output/study/task/canonical repository/kind/rule/applicability, with parallel record indices, values and source-table IDs; source digests occur once per source. `:186–203,231–249`: top-level input pins only; manifest and familiarity use grouped encoding. | `PipelineTests.test_group_conservation_pin_once_and_source_table` (`tests/test_identifiers.py:200`) calls real `preparation.verify_tree_pins` on every JSON payload, checks each frozen input exactly once and all original indices exactly once. `test_successful_publication_groups_and_pin_once` (`:298`) additionally exercises actual fabricated publication, all four JSON outputs and the complete familiarity JSONL stream, checking each frozen pin once and grouped indices. |
| R1-02 | `execution/identifiers/projection.py:65–78`: commit and non-empty blob rules explicitly end in `_any_scope` and apply to all three tests, retaining canonical original repository scope in provenance. Paths/PRs retain repository-scoped applicability; empty blobs have a non-match rule and no applicability. | `ProjectionTests.test_any_scope_sha_and_repository_scoped_path_pr` (`tests/test_identifiers.py:121`) plants fastapi-scoped commit/blob/path/PR/empty-blob records and checks every required boundary. |
| R1-03 | Existing finite freshness boundaries are retained through the grouped representation. Non-parsed rows remain individual unresolved coverage entries (`execution/review_identifiers.py:159–163`); the source table records unique/missing/ambiguous resolution (`:136–145`). | Table-driven `ProjectionTests.test_scoped_paths_prs_and_familiarity` (`:136`), `test_rejection_boundaries` (`:148`), `test_unattributed_boundaries` (`:162`), `test_aliases_and_case` (`:177`); plus `PipelineTests.test_nonparsed_statuses_all_unresolved` (`:218`) for all four statuses and `test_missing_and_ambiguous_source_lookup` (`:231`). These cover all listed reasons, rule/applicability boundaries, TypeScript/case aliases, invalid commit/blob hex, non-empty unattributed blobs and repository familiarity not being clearance. |
| R1-04 | `tests/test_identifiers.py:34–114` builds eight fabricated files, the manifest and matching seal, source metadata, private config/output/clock locations. Synthetic runs use the real journal/lock/publisher, patch only the runtime/config/clock trust anchors and fixture signed hashes. No real study input is exercised. Existing safeguards at `execution/review_identifiers.py:35–62,206–227,229–230,259–262` are now caught. | `test_frozen_valid_eight_file_fixture` (`:284`), `test_frozen_pin_duplicate_and_count_refusals` (`:287`), `test_seal_identity_and_crosslinks_refuse` (`:318`), `test_context_identity_refuses` (`:329`), `test_coverage_and_enumeration_identity_refuse` (`:274`), `test_runtime_noncanonical_signed_and_source_count_refuse` (`:335`), `test_existing_root_refuses_and_preserves_bytes` (`:352`), `test_partial_publication_retained` (`:366`). Refused terminal journal records and retained original/partial bytes are asserted. |
| R1-05 | `execution/review_identifiers.py:84–94`: exact inertia listing/coverage path-set equality and duplicate refusal, plus enumeration consistency. `:96–127`: each output's projected indices plus record-level rejected indices must equal every input index exactly once; compatible/unattributed/unscoped/coverage/inertia counts must match the pinned seal summary, and coverage's own counts must agree. `:133–158,174–183,187–202`: emit conservation and repository × study × kind × rule tables, unresolved weighted counts by reason and explicit familiarity sources for every test repository, including Prometheus. | `PipelineTests.test_dropped_record_refuses` (`:247`) plants a dropped projection; `test_listing_mismatch_and_duplicates_refuse` (`:266`) plants omitted/extra/duplicate listing paths; `test_summary_count_refusals` (`:256`) plants each required summary mismatch. `test_group_conservation_pin_once_and_source_table` (`:200`) checks emitted counts, complete index coverage and nonempty Prometheus familiarity sources. |

## Review packet placement

Moved and updated the draft request from `execution/identifiers/REVIEW_REQUEST.md` to `preparation-reviews/identifiers-REVIEW_REQUEST.md`. It documents grouping, pin-once references, any-scope SHA matching, conservation and the unchanged conservative unknown-freshness rule. No review record remains among my `execution/` files. Main must regenerate the execution-source manifest to cover every current byte under `execution/`; the review request and this response are outside that tree.

## Commands and complete output

Working directory for both test invocations and the digest command: `/Users/adamjackson/Projects/work/scout-x1prep`.

First invocation, after the required fix implementation and expanded safeguards:

```sh
.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py'
```

Complete process output:

```text
...................................................................
----------------------------------------------------------------------
Ran 67 tests in 4.032s

OK
```

Tool-reported wall time: 4.49 seconds; exit code 0.

Final invocation, after adding the successful-publication/familiarity pin-once test and making empty-blob applicability explicitly empty:

```sh
.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py'
```

Complete process output:

```text
....................................................................
----------------------------------------------------------------------
Ran 68 tests in 1.399s

OK
```

Tool-reported wall time: 1.54 seconds; exit code 0. These are synthetic unit tests only, not real identifier projection or completeness evidence.

Digest command:

```sh
shasum -a 256 evaluation/x1-triage-ranking/execution/review_identifiers.py evaluation/x1-triage-ranking/execution/identifiers/projection.py evaluation/x1-triage-ranking/execution/tests/test_identifiers.py evaluation/x1-triage-ranking/preparation-reviews/identifiers-REVIEW_REQUEST.md evaluation/x1-triage-ranking/execution/preparation.py evaluation/x1-triage-ranking/execution/run-config.json
```

Complete process output:

```text
fcb4ae28b3448d95d1e2642d9faeebef2550508c50a3d2336fbf6ada0072d133  evaluation/x1-triage-ranking/execution/review_identifiers.py
2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072  evaluation/x1-triage-ranking/execution/identifiers/projection.py
fa37ff16800474771ef8a8a8ae845688897fadd761c05ff7a43c5f54cc9e5907  evaluation/x1-triage-ranking/execution/tests/test_identifiers.py
f13e25ea529d6cbf4ca5fd2cadf84e42f7c25d1566287b3d1a34c4cea95d012c  evaluation/x1-triage-ranking/preparation-reviews/identifiers-REVIEW_REQUEST.md
1a116d86b98b088c474a0f4e974af9be8299484ac1deee99dba8f3e7123bc9bf  evaluation/x1-triage-ranking/execution/preparation.py
c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54  evaluation/x1-triage-ranking/execution/run-config.json
```

Exit code 0. The last two files are integration dependencies observed, not edited. This response is newly written after that digest command and does not contain a circular self-digest.

## Remaining authority boundary

Submit the revised packet for round 2. Do not run the real projector until scoped implementation review reports zero issues. The later real-output completeness/delta review must still decide the historical source/field/mechanism gaps; grouping does not resolve or waive them, and no new raw schema handling is authorised.
