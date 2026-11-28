# X1 preparation (d) identifier projector — implementation review — round 2

**Scope.** This round reviews only the implementation of the action-(d) projector, before Main runs it. It does not sign off identifier completeness or the delta; a later review of the retained outputs does that. It grants no inventory, SRS, model, development, full-OC or GUIDE authority. It made no model calls and did not run the projector on real inputs.

## Reviewed files

All SHA-256 values were computed by Main. This round did not recompute them. Paths are under `evaluation/x1-triage-ranking/`.

| File | SHA-256 |
|---|---|
| `execution/review_identifiers.py` | `fcb4ae28b3448d95d1e2642d9faeebef2550508c50a3d2336fbf6ada0072d133` |
| `execution/identifiers/projection.py` | `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072` |
| `execution/tests/test_identifiers.py` | `fa37ff16800474771ef8a8a8ae845688897fadd761c05ff7a43c5f54cc9e5907` |
| `preparation-reviews/identifiers-REVIEW_REQUEST.md` (moved out of `execution/`) | `f13e25ea529d6cbf4ca5fd2cadf84e42f7c25d1566287b3d1a34c4cea95d012c` |
| `execution/preparation.py` (integration; round 1 saw `d75e40e0…`) | `1a116d86b98b088c474a0f4e974af9be8299484ac1deee99dba8f3e7123bc9bf` |
| `execution/run-config.json` (integration) | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |
| `preparation-reviews/identifiers-impl-round-1.md` | `33d57430e503438b6e3620f847a69a082e35dcff12aa3674df2f5e53f43a6a73` |
| `preparation-reviews/identifiers-impl-round-1-response.md` | `83a42155ea030f1fd277045ae47c098dfac6277d018d998f41716ac96fb85108` |

## Basis (unchanged from round 1)

| Document | SHA-256 |
|---|---|
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |
| `PREREGISTRATION.md` | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| `preparation/exposure-recipe.json` | `c065347e08a8571228764dae19fc686f3a6909e612a4119aad7bb58d86c5d383` |
| `preparation/exposure-sources.json` | `67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60` |
| `preparation/frozen-exposure-inputs.json` | `ec92fba6d3f91c671301244539a8caefc5025c0d6b03efca77c5bdfc4bd96425` |
| `reviews/rev9-digests.json` | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |

## Round-1 disposition

| ID | Disposition | Evidence | Catching test(s) |
|---|---|---|---|
| R1-01 | **Implementation fixed; required test incomplete (see R2-01).** | Each row now identifies its source only by output name and record index (`projection.py` l.36–37); the unattributed-scope issue no longer nests the row (l.88). Provenance gaps and scope gaps are each recorded once per output, with a count and index ranges (`review_identifiers.py` l.109–117). The manifest and familiarity log are grouped, with parallel lists of record indices, values and source IDs. Source digests sit once in a source table, under the renamed keys `source_relative_path` and `source_bytes_sha256` (l.132–155). Delta and enumeration metadata use the same renamed keys (l.164–173). Real pins appear only once, in `common.inputs` and `clock`. At seal time, `collect_seal` (`preparation.py` `verify_tree_pins`) now re-hashes 11 pins per record across 4 records, about 0.4 GB, instead of the ~43 TB found in round 1 [INFERENCE]. | `test_group_conservation_pin_once_and_source_table` and `test_successful_publication_groups_and_pin_once` run the real `preparation.verify_tree_pins` and would fail if per-row pins or the unrenamed `path`/`sha256` metadata came back. However, the fixture has one valid record per output (`test_identifiers.py` l.54–57). The required "grouped indices cover every record exactly once" assertion is therefore vacuous for groups with two or more records. |
| R1-02 | **Fixed.** | Commits and non-empty blobs use the `_any_scope` rules and apply to all of `TESTS`. Paths and PRs apply only to their own repository. The empty blob applies to nothing. The original scope is kept in `repository` (`projection.py` l.65–78). | `test_any_scope_sha_and_repository_scoped_path_pr` plants fastapi-scoped commit, blob, path, PR and empty-blob records. Reverting to scoped-only SHA matching fails it. |
| R1-03 | **Fixed.** | Non-parsed coverage rows become unresolved (`review_identifiers.py` l.159–163). Source resolution records unique, missing or ambiguous (l.136–145). Every listed rejection and rule boundary is in `projection.py`. | `test_scoped_paths_prs_and_familiarity`, `test_rejection_boundaries`, `test_unattributed_boundaries`, `test_aliases_and_case`, `test_nonparsed_statuses_all_unresolved` (all four statuses) and `test_missing_and_ambiguous_source_lookup`. The ambiguous fixture supplies two real matching `(study, path)` entries. |
| R1-04 | **Implementation fixed; two of the listed refusals not independently tested (see R2-02).** | The checks are at `review_identifiers.py` l.35–62, l.206–227 and l.229–262. | Eleven planted refusals catch their guards: frozen pin mismatch, duplicate, count, snapshot identity and seal differing from snapshot; seal identity and all cross-links; context identity; coverage/enumeration identity; interpreter; existing root with bytes preserved; partial publication retained. Not tested independently: the signed digest of `exposure-sources.json`, and canonical paths for `--frozen` and `--sources`. |
| R1-05 | **Fixed as requested.** | Inertia listing set equality and duplicate refusal (l.84–94). Per-output sorted-index conservation, which catches both drops and double counts (l.103–105). Counts checked against the seal summary and against coverage's own counts (l.122–127). Count tables and familiarity sources (l.174–183). | `test_dropped_record_refuses`, `test_summary_count_refusals` (all five keys) and `test_listing_mismatch_and_duplicates_refuse` (omitted, duplicate and extra). |

## New issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R2-01 | minor | `review_identifiers.py` l.103–117 (conservation and unresolved aggregation), l.147–155 (grouping), l.241–249 (familiarity lines); `test_identifiers.py` l.54–57, l.200–216, l.298–315 | Conservation is checked on the intermediate `projected` + `record_issues` lists, before encoding. Nothing checks the emitted groups or the emitted unresolved entries, and the pipeline fixture never exercises the code that writes them. It has one valid record per output, no record-level reject and no unattributed path; every range is `[[0,0]]`, and `len(unresolved)==3` counts only the provenance-gap entries. Any of these changes would pass all 68 tests and the runtime conservation check, while the emitted tables still report conservation as complete: <br>(a) re-creating a group on every record (l.149–151) keeps only the last record of each group; <br>(b) dropping `unresolved.extend(record_issues)` (l.109) removes rejected records from every output; <br>(c) building `scope` from `record_issues` (l.110) drops the aggregated `unattributed_scope` entry, which covers 23,462 real unscoped paths; this aggregation is new code from the R1-01 fix; <br>(d) wrong `count` or `record_index_ranges`; <br>(e) familiarity group lines written twice, because the test's `{output: record_indices}` dictionary collapses duplicates. <br>The output review has no shell, so it depends on these emitted tables. This breaks R1-01's "with its count and index coverage" and its grouped-coverage test requirement, and P l.23's binding of unresolved identifiers. | 1. In `produce()`, after grouping, refuse with `emitted identifier conservation failure` unless both hold per output: <br>• the flattened group `record_indices` plus the indices of record-level unresolved entries for that output equal `range(n)` exactly once, and the flattened `(record_index, value)` pairs equal the projected rows; <br>• the `unattributed_scope` entry's `count`, and the expansion of its ranges, equal the projected unattributed path/PR indices. <br>2. Extend the fixture: at least two same-group records per identifier output; a repeated source and a distinct source in the unattributed output; at least one interleaved record-level reject per output; at least two unattributed paths at non-adjacent indices; summary updated to match. <br>3. Assert the exact ordered reconstruction of the groups, the source-table size, the exact unresolved entries with counts and ranges, the weighted `unresolved_by_reason`, and the multiplicity of familiarity group lines as a list. <br>4. Add one planted emitted-encoding loss (e.g. `index_ranges` patched to drop a range) that must refuse. |
| R2-02 | minor | `review_identifiers.py` l.211–215 (canonical-path loop), l.221–223 (signed-digest loop); `test_identifiers.py` l.335–350 | R1-04 asked for each refusal to be asserted, specifically so that a check over the wrong list could not pass unnoticed. The test corrupts only the frozen manifest's `SIGNED` digest and only `--output-root`'s path. If the digest check covered only `args.frozen`, any unsigned `exposure-sources.json` with 24,254 rows and the right schema would pass. That file drives the matched/changed/missing/new delta (recipe source binding; P l.23 source digests). If the canonical-path check covered only `--output-root`, relocated input copies would be accepted and recorded as non-canonical pins. All tests would still pass. | Add to the runtime test, keeping every other prerequisite valid so the intended guard is what refuses: <br>• `SIGNED` with a valid frozen digest and a wrong sources digest must give the `signed recipe` refusal and a refused journal; <br>• `--frozen`, and separately `--sources`, replaced by a byte-identical copy at a non-canonical path with the same basename must give the `noncanonical` refusal and no journal. <br>Update the expected terminal-journal list to match. |

Neither issue is a live defect in the current code. By inspection, the current grouping appends correctly, both signed digests are checked, and record issues and scope are emitted [INFERENCE: static reading, not mutation runs]. Both issues leave a required safeguard unverified, so they are open.

## Integration (`preparation.py` 1a116d86, `run-config.json` c817f678)

I found no defect affecting the projector:
- The APIs the projector calls match the current code: `Journal(config, config_path, clock, action, source_pins)`, `finish`, the shared `operation_lock` (which refuses once the seal is reserved), `load_config` and `paths`.
- The seal's record contract is met:
  - schema equals `required_records.kind` for all four JSON outputs;
  - `status: passed`;
  - `clock` equals `pin(clock_path)`, since `pin()` makes paths absolute;
  - no `failures`, `unmeasured` or `key_present` fields.
- `familiarity.jsonl` is in `required_files`, so the unaccounted-output walk passes.
- `--output-root` canonical `OUT/identifiers` matches the config.
- The frozen inputs live under `scout-x11/…`, outside OUT and EXEC, so seal re-verification hashes them in place.
- The response says it did not modify `preparation.py`, and I could not diff the two revisions. The checks above are against the current bytes.
- An existing design property, unchanged by this revision: any refused projector journal blocks the seal (`collect_seal` requires every journal to have passed), so a refusal on real data needs a reviewed recovery.

## Real-data precondition checks (metadata only; no predictor–outcome association)

These checks used read-only jq queries on the signed `frozen-exposure-inputs.json` and the pinned v4 `coverage.json` and `inertia-enumeration.json`.
- The seal summary has every key the projector compares:
  - `compatible_records` 50,016
  - `unattributed_records` 262,080
  - `unscoped_by_kind` {commit 61,135, blob 177,483, path 23,462, pr 0}
  - `coverage_status_counts` {non_data 321, parsed 24,001, unclassified 3}
  - `inertia_data_files` 36
- `coverage.counts` is the same flat status map.
- `recursive_listing` has 74 unique strings. The inertia coverage rows are 74 distinct paths.
- The 36 enumeration `data_files` have the same 13-key set as the parsed inertia coverage rows.
- `enumeration_result` holds only `data_files`, `screening_json_jsonl_csv_files` and `study_groups`: no outcome values and nothing that looks like a pin.

The R1-05 refusals should therefore not fire on the real inputs [INFERENCE: set equality and row equality across the two files were not computed directly].

## Suggestions (not issues)

1. Keep the allowlisted value in record-level unresolved entries (round-1 suggestion 2), or apply the any-scope rule before resolving the repository for commits and blobs. Today an `unknown_repository` commit or blob loses its SHA, so a case at that SHA could not be matched. v4 should emit no such records [INFERENCE from extractor normalisation], and the conservation table would show any.
2. Split `familiarity_sources` into scoped sources and any-scope SHA sources, and test with a Prometheus-scoped `repository_identity` record. After R1-02 every SHA-bearing study appears under all three repositories, so the Prometheus assertion at l.213–214 cannot detect a Prometheus-specific loss.
3. Hash and parse the same bytes in one read (round-1 suggestion 8). `unattributed-identifiers.json` is now parsed twice after it is pinned.
4. Pin a partially written `familiarity.jsonl` in the refused journal.
5. Add separate tests for: coverage's own count mismatch, duplicate inertia coverage paths, a planted double count, an `enumeration_result` count mismatch, and `data_files` absent from coverage.
6. Remove the unused `output_pin` parameter (`projection.py` l.29).
7. Document that the count table counts each any-scope SHA once per test repository.
8. Add missing and ambiguous source-resolution counts to `unresolved_by_reason`.

## Recommendation, alternatives, risks, execution

- **Recommendation.** Fix R2-01 with a runtime check on the emitted encoding plus a multi-record fixture, and fix R2-02 with independent sources-digest and input-path refusal tests. Then submit round 3. Do not run the projector before round 3 reports zero issues.
- **Alternatives considered and rejected.**
  - *Tests only for R2-01.* Rejected: conservation would still be checked on intermediate lists, and nothing at runtime would bind the artefact the shell-less output review relies on.
  - *Accept as suggestions because the current code is correct by inspection.* Rejected: round 1 required these specific tests, and the severity calibration treats untested decision boundaries as issues.
  - *Revert to per-row pins with a verifier cache.* Rejected again: it would restore the size and seal-bloat problems.
- **Risks.**
  - A new runtime check could itself refuse wrongly on real data. That would create a refused journal that blocks the seal. Mitigate by deriving the check from the same projected lists and covering it with the planted fixtures.
  - All "would survive" claims about mutants come from static inspection, not mutation runs.
  - Everything is still pre-run and fully reversible: no outputs exist.
- **Execution outline.**
  1. Add the post-grouping and post-aggregation conservation refusal in `produce()`.
  2. Extend the fixture and assertions as R2-01 sets out, including the planted encoding loss.
  3. Add the R2-02 refusal tests and update the expected journal statuses.
  4. Main runs `.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py'`.
  5. Recompute SHA-256s, regenerate the execution-source manifest, and submit round 3.
  6. If round 3 reports zero issues, run the reviewed command once at its canonical paths.

## Verification

- **Files read in full:** `projection.py`, `review_identifiers.py`, `test_identifiers.py`, `preparation.py` l.1–340, `run-config.json` (records, files, reviews), `identifiers-REVIEW_REQUEST.md`, the round-1 record and the round-1 response.
- **`verify_inputs.pin` and `exclusive_json`:** `pin()` makes paths absolute; outputs are written with `indent=2` and exclusive create.
- **jq queries:** metadata only, as listed above. No score, label or outcome values were read.
- **Not run:** no tests, scripts, model calls or digest recomputation.
- **Main's report:** the full execution suite with the pinned interpreter gave 68 tests OK. That suite would not catch R2-01 or R2-02.
- **Peers.**
  - *innovation-challenger:* raised R2-02 and the grouping and familiarity-multiplicity half of R2-01. Its R2-02 finding turned R1-04 from fixed into partially fixed in this record.
  - *experiment-peer:* raised the unresolved-emission half of R2-01 (record-level rejects and aggregated `unattributed_scope`). It also supplied suggestions 1, 2, 4 and 8 and the extractor-based [INFERENCE] that real-data record issues are zero. It judged R1-04 fixed. I side with the challenger, because the sources digest is a separate boundary and R1-04 named the wrong-list failure class.
  - *My own analysis:* I had the single-record grouping gap as a suggestion. Both peers' evidence that conservation runs before encoding, with no output-level test, changed it to an issue.
  - All three analyses agree: R1-02, R1-03 and R1-05 are fixed; there is no outcome leak, no change of units, and no integration regression.

## Verdict

**2 issues open** (R2-01 minor, R2-02 minor). The projector may not run yet. A later round reporting zero issues must name the revised SHA-256s.
