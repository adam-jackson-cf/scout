# X1 preparation (d) identifier projector — implementation review — round 1

**Scope.** This round reviews only the implementation of the action-(d) projector, before Main runs it. It does not sign off identifier completeness or the delta; a later review of the retained outputs does that. It grants no inventory, SRS, model, development, full-OC or GUIDE authority.

## Reviewed files (SHA-256 computed by Main; this round did not recompute them)

All paths are under `evaluation/x1-triage-ranking/execution/`.

| File | SHA-256 |
|---|---|
| `review_identifiers.py` | `7d1cf1cf7ea4ca0af05c9df1477dc51188543d3a80ed354010d1e97210aa2e5c` |
| `identifiers/projection.py` | `d7d1524888a2ed9758576c511ca68fae1412449a0dff3e56e0e74d8dc3b6e207` |
| `identifiers/REVIEW_REQUEST.md` | `9aca36c06a19938f0778617917d5b643f956a35dee0b8317c596a3500c60463d` |
| `tests/test_identifiers.py` | `f27594944241425c40a8ba3d4e682296f2dc4b8b78349892780f4d64a823a0e6` |
| `run-config.json` (integration) | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |
| `preparation.py` (integration) | `d75e40e0948351412c6446eeddd81b363d44a059c689c37ce15af82af6341bf9` |

## Basis digests

| Document | SHA-256 |
|---|---|
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |
| `PREREGISTRATION.md` (l.23, l.194, l.222) | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| `preparation/exposure-recipe.json` | `c065347e08a8571228764dae19fc686f3a6909e612a4119aad7bb58d86c5d383` |
| `preparation/exposure-sources.json` | `67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60` |
| `preparation/frozen-exposure-inputs.json` | `ec92fba6d3f91c671301244539a8caefc5025c0d6b03efca77c5bdfc4bd96425` |
| `reviews/rev9-digests.json` | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |

## What the implementation gets right

- **The schema assumptions match the real frozen v4 files.** Checked against key sets and counts only:
  - exposure: 50,016 records, each with `repository,study,task` and exactly one of commit, blob or path;
  - additional: 8,796 PR or repository-identity records;
  - unattributed: 262,080 records with one key set, and kind/matchability always paired as the projector expects;
  - coverage: the v2 envelope, row keys inside `COVER_KEYS`, and 24,251 Scout plus 74 inertia rows.

  Every record should therefore become either a row or an unresolved entry.
- **Read-only reuse holds.**
  - `extract_exposures.py` is hashed as bytes and never imported or run.
  - No prior-study raw source is opened.
  - Coverage `studies`, `counts` and `limitations` are not carried into the outputs, and exposure `studies` is not projected. No score, rank, label, verdict, disagreement, threshold or outcome field is selected.
- **Prometheus familiarity is in the frozen data and is kept.** There are repository-identity records from `change-cost-v2` and the `inertia-bench/docs/design*` studies, plus 5,533 commit, 6,149 blob and 1,096 path rows scoped to prometheus.
- **Development repositories are not freshness-filtered.** Their rows are kept with empty test applicability, and the outputs say so.
- **Outputs are created once.** The directory is created with `mkdir`, files with exclusive create, and the journal records refusals.
- **Source reconciliation reads metadata only.** It compares Git blob, SHA-256 and byte count. Matches on blob alone are labelled separately.
- **Marking missing provenance unresolved is the correct conservative handling.** v4 dropped source, field and mechanism provenance when it deduplicated scoped records. The projector records that gap per output and index. It does not infer provenance from study names and does not drop the records. Missing provenance must never become freshness, and must never by itself count as a whole-study failure.

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R1-01 | major | `projection.py` l.36–37 (`provenance.output`) and l.84 (`identifier=row`); `review_identifiers.py` `produce()` per-row unresolved entries; `preparation.py` `collect_seal` → `verify_tree_pins` / `validate_pin` (l.158–170, l.60–64) | Every manifest row, every per-row unresolved entry and every nested unattributed-scope row embeds the full input pin `{path,sha256,bytes}`. At seal time, action (e) walks every required record and re-reads and re-hashes each such dict; `pin()` has no cache. That is 320,892 × 2 + 23,462 × 2 ≈ 688,700 checks. About 571,000 of them re-hash the 76,255,763-byte unattributed file, roughly 43 TB of SHA-256, while the seal is reserved and the exclusive lock is held. The seal's `sources` list would also gain about 689,000 duplicate pins. The same per-row duplication inflates the outputs: each of `manifest.json`, `unresolved.json` and `familiarity.jsonl` is likely well over GitHub's 100 MB per-file limit, which blocks the planned writer-PR commit [INFERENCE: estimated from 291 bytes per record in the source file at `indent=2` and about 3× more keys per row]. Because outputs are create-once, finding this after the run needs a reviewed recovery. | Rows refer to their source output by file name only, with no pin. The pins appear once, in the existing top-level `inputs`. Emit the manifest and the familiarity log grouped by (output, study, task, canonical repository, kind, matching rule, applicability), each group holding parallel `record_indices` and `values` lists. Add a once-per-source table for unattributed source digests. Record the provenance-gap reason once per output, with its count and index coverage, not once per row. Keep record-level exceptions (schema or structure issues) individual. **Tests:** run `preparation.verify_tree_pins` on every produced payload and assert that each frozen pin is found exactly once per output; assert that the grouped indices cover every record exactly once. |
| R1-02 | major | `projection.py` l.74 (`applies_to_test_repositories`) | A commit or blob record scoped to a non-test repository gets `applies_to_test_repositories=[]`. v4 assigns scope by heuristic: inherited prefix contexts, filename aliases, and nested repository keys, including forks and mirrors, which share object IDs with upstream. That assignment is known to be wrong at least once (X11 `exposure-source-index.json` l.101: `human-confirmation-51.csv` mis-scoped to microsoft/TypeScript). Failure case: a django or prometheus SHA in an `x2d-*` file without its own repository key inherits fastapi scope, so an X1 case at that commit or blob looks fresh. This breaks P l.23's "conservative commit/blob matches". An equal SHA means the same object, so applying every SHA to every test repository costs no false matches beyond true object identity. | Apply every commit and every non-empty blob to all three test repositories whatever its scope. Use a distinct rule, e.g. `commit_or_parent_either_image_any_scope` / `blob_either_image_any_scope`, and keep the original scope as provenance. Paths and PRs stay repository-scoped, and the empty blob stays a non-match. **Tests:** a fastapi-scoped commit and a fastapi-scoped blob apply to all of `TESTS`; a fastapi-scoped PR or path still applies to none. |
| R1-03 | minor | `review_identifiers.py` `produce()` (non-parsed coverage → `source_not_identifier_parsed`; `coverage_by_source` lookup); `projection.py` `project()` / `canonical()` | Several freshness decision boundaries have no test, so a regression in any of them passes all 12 tests: <br>• non-parsed coverage statuses becoming unresolved (324 real rows; the pipeline fixture's only coverage row is parsed); <br>• ambiguous or missing source lookup; <br>• scoped test-repository path or PR applying to its own repository only; <br>• `repository_identity` mismatch; <br>• unknown matchability; <br>• invalid commit or blob hex; <br>• `unknown_repository`; <br>• the `typescript` alias and case variants of the test slugs; <br>• a non-empty unattributed blob; <br>• repository familiarity never acting as clearance. <br>Example: removing the non-parsed loop would make the outputs look complete. | Add table-driven synthetic tests asserting the reason or rule and applicability for each case above, including all of `non_data_source_or_document`, `unclassified_non_data`, `could_not_parse` and `unsupported_git_object`. |
| R1-04 | minor | `review_identifiers.py` `frozen_inputs()`; `run()` | The checks that enforce the safeguard "only the exact pinned frozen v4 outputs, read-only" have no tests. Untested checks: <br>• pin mismatch; <br>• duplicate name; <br>• not exactly eight inputs; <br>• seal identity, and seal differing from `upstream_seal_metadata`; <br>• recipe, context and output cross-links; <br>• coverage or enumeration identity; <br>• interpreter pin; <br>• non-canonical path; <br>• signed pins; <br>• source count other than 24,254; <br>• output directory already existing; <br>• refused journal record. <br>A check made vacuous (wrong key or wrong list) would pass silently. | Build a fabricated eight-file set and frozen manifest in a temp directory, patching `SIGNED` and the config paths. Assert each refusal. Assert that a run against an existing output root refuses, writes a refused journal, and leaves the original bytes unchanged. Assert that a failure partway through publishing keeps the partial outputs. |
| R1-05 | minor | `review_identifiers.py` `produce()` and payload construction | Nothing checks that records are conserved, and no count tables are emitted. The output review is a read-only council with no shell (round 9 l.24, l.402). It cannot verify by hand that roughly 320,000 records were neither dropped nor double-counted. Inertia completeness is checked one way only: enumeration `data_files` must be a subset of coverage. It is never checked that `recursive_listing` equals the inertia coverage paths. | In `produce()`: <br>• assert per output that rows plus record-level issues equal the record count; <br>• refuse unless the counts equal the frozen seal summary (50,016 compatible; 262,080 unattributed; unscoped 61,135 / 177,483 / 23,462 / 0; coverage status counts; 36 inertia data files); <br>• assert that the inertia coverage path set equals `recursive_listing`, with no duplicates. <br>Emit counts by test repository × study × kind × rule, unresolved counts by reason, and an explicit familiarity-source list per test repository (Prometheus included). **Test:** a planted dropped record and a planted listing mismatch each cause a refusal. |

## Suggestions (not issues)

1. Also list new and changed source files in `unresolved.json`; today they appear only in `delta-review.json`.
2. Keep the allowlisted string values inside record-level unresolved entries.
3. Always emit `source_file_blob`, as `null` when unmatched.
4. Rename `field_provenance`, which names the output record rather than the source field.
5. Split the provenance-gap reason into "source resolved, field and mechanism missing" and "source missing".
6. Copy the known `human-confirmation-51.csv` mis-scope into `limitations`.
7. Make `canonical()` strip `.git` and compare case-insensitively; treat near-matches to a test repository as unresolved.
8. Hash and parse the same bytes in one read, so a file cannot change between `pin` and `produce`.
9. fsync the parent directory after writing `familiarity.jsonl`, and hoist `import os` to the top of the file.
10. Pin the imported `preparation.py` and `verify_inputs.py` in the journal's `source_pins`. This matches the existing convention, so it is not an issue.
11. Validate the keys of `enumeration_result` instead of copying them wholesale.

## What the later completeness/delta output review must decide

1. **Source attribution.** Whether attribution at study level, plus output and index, plus coverage-level source digests, satisfies "binding source/file digests" and "record every exclusion and familiarity source". The signed recipe step 4 expected "rows with source-file hashes and schema/field provenance", and v4 has none. **Recommend** recording this as a deviation, accepting it on condition that every inventory exclusion cites output, index and study.
2. **Matching without a mechanism.** Test-scoped paths: django 1,833, prometheus 1,096, TypeScript 1. **Recommend:** any case touching such a path is ineligible.
3. **The 23,462 unscoped paths**, 18,029 of them from `change-cost-v2`, which has django and prometheus familiarity. The extractor reads these as "unmatchable"; the projector reads them as "unknown, ineligible". **Recommend** the ineligible reading, with the impact quantified at inventory, unless a separately authorised per-schema delta resolves their scope.
4. **Criteria absent from v4.** v4 extracted no issue IDs, Scout fingerprints, abbreviated SHAs, spans, rename links or excerpt hashes. The review must decide whether commit, blob and path coverage stands in for P l.23's issue-ID, evidence, span, rename and excerpt criteria for `change-cost-v2` and the inertia-bench screening/design work. If not, a separate per-schema request is required. Extraction must never be rerun.
5. **Files not parsed for identifiers, and the delta.** 324 files were not parsed for identifiers (286 Scout, 38 inertia), and there are new, changed and missing files. One example: 3 files in `d1-triage-minimums-exploratory` are signed but absent from frozen coverage.
6. **X11 directories.** `evaluation/x11-merge-association/` and `x11b-guard-recovery/` exist in the worktree but are outside the 24,254-file signed universe at `7bb73e52`. The review must confirm they hold only "X11 planning" and no examined case-level units, or treat them as unresolved.
7. **Conservation.** Confirm the R1-05 counts against the seal summary.

## Recommendation, alternatives, risks, execution

- **Recommendation.** Do not run the projector until R1-01 to R1-05 are fixed and a round 2 reports zero issues.
- **Alternatives considered and rejected.**
  - For R1-01, caching `validate_pin` in `preparation.py` fixes the re-hashing but leaves outputs too large to commit, a bloated seal, and an edit to a shared integration file.
  - For R1-02, keeping scoped-only SHA matching with a disclosure leaves a known mis-scoping mechanism able to clear cases at no saving.
- **Risks.**
  - The size figures in R1-01 are inference, not measurements.
  - Fixing R1-02 may slightly increase test exclusions for identical blobs shared across repositories, which is intended.
  - Everything is still pre-run and fully reversible: this round created no outputs.
- **Execution outline.**
  1. Implement the R1-01 grouped, pin-once encoding.
  2. Implement the R1-02 any-scope SHA rule.
  3. Implement the R1-05 conservation and listing checks and count tables.
  4. Add the tests for R1-01 to R1-05.
  5. Main runs `.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py'`.
  6. Recompute SHA-256s and submit round 2.
  7. If round 2 reports zero issues, run the reviewed command once at its canonical paths.

## Verification

- Read the reviewed files, the preregistration (l.23, l.194, l.222), round 9, the recipe, the sources manifest and the frozen-inputs manifest.
- Read the frozen v4 extractor's source as code only, and `exposure-contexts-v4.json`.
- Ran jq metadata queries on the pinned frozen outputs: key sets, repository and study names, counts, and coverage status. No score, label or outcome values were read.
- Grepped X11 for the mis-scope disclosure, confirmed at `exposure-source-index.json` l.101 and `PREREGISTRATION.md` l.73.
- This round ran no tests, scripts, model calls or digest recomputation.
- Main reports 27 tests OK with the pinned interpreter. None of those tests would catch R1-01 to R1-05.
- **Peers.**
  - The innovation-challenger independently raised the test gaps behind R1-03 and R1-04, and confirmed the schemas match and that nothing leaks.
  - The experiment-peer independently found R1-01 and R1-02, and supplied the mis-scope evidence and the R1-05 inertia-listing check.
  - All three analyses agree that marking missing provenance unresolved is correct.

## Verdict

**5 issues open.** The projector may not run yet. A later round reporting zero issues must name the revised SHA-256s.
