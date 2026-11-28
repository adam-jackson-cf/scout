# X1 preparation (d) identifier projector — implementation review — round 3

**Scope.** This round reviews only the action-(d) projector implementation, before Main runs it. It does not sign off identifier completeness or the delta; that needs a separate review of the retained outputs. It grants no authority for inventory, SRS, models, development, full-OC or GUIDE. No model calls were made beyond the two council peers. The projector was not run on real inputs, and no files were edited.

## Reviewed files

Main computed these SHA-256 values with `shasum -a 256`; this round did not recompute them. Paths are under `evaluation/x1-triage-ranking/`.

| File | SHA-256 | Change since round 2 |
|---|---|---|
| `execution/review_identifiers.py` | `adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2` | changed (`validate_emission`) |
| `execution/identifiers/projection.py` | `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072` | unchanged |
| `execution/tests/test_identifiers.py` | `955bc57a9a8cd97e0291ed3582cc0d330ed6060186dda7ac38d1cbd6a033ebea` | changed |
| `preparation-reviews/identifiers-REVIEW_REQUEST.md` | `10879ebd6a61029fb88bafd5924af6fb154b7a7571a33c651d6849c0641210f9` | changed (l.32 only) |
| `execution/preparation.py` (integration) | `1a116d86b98b088c474a0f4e974af9be8299484ac1deee99dba8f3e7123bc9bf` | unchanged |
| `execution/run-config.json` (integration) | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | unchanged |
| `preparation-reviews/identifiers-impl-round-2.md` | `77e077a3…` (Main-supplied prefix) | prior record |
| `preparation-reviews/identifiers-impl-round-2-response.md` | `aabbd0bb…` (Main-supplied prefix) | prior record |

## Basis (unchanged)

| Document | SHA-256 |
|---|---|
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |
| `PREREGISTRATION.md` | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| `reviews/rev9-digests.json` | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |
| `preparation/exposure-recipe.json` | `c065347e08a8571228764dae19fc686f3a6909e612a4119aad7bb58d86c5d383` |
| `preparation/exposure-sources.json` | `67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60` |
| `preparation/frozen-exposure-inputs.json` | `ec92fba6d3f91c671301244539a8caefc5025c0d6b03efca77c5bdfc4bd96425` |

## Round-2 disposition

| ID | Disposition | Evidence | Catching test(s) |
|---|---|---|---|
| R2-01 | **Runtime check fixed. Test oracle still incomplete: see R3-01.** | `validate_emission` is at `review_identifiers.py` l.76–114 and is called at l.200, after grouping and before coverage entries are appended. For each output it refuses unless all of the following hold: <br>• the three parallel group lists have equal length; <br>• the emitted (index, value) pairs equal the projected pairs as a multiset; <br>• group indices plus emitted record-level reject indices equal `range(input)` exactly once; <br>• the `unattributed_scope` entry and the provenance-gap entry are each present exactly when expected; <br>• each range has start ≤ end; <br>• each entry's `count` and its expanded ranges equal the expected indices. <br>Scope indices are recomputed from projected rows (`repository is None`, kind `path`/`pr`) rather than copied from the issue list, so the check is independent of the aggregation it verifies. | All five mutants named in round 2 are caught [INFERENCE: static, not mutation runs]: <br>(a) a group re-created on every record: runtime pair check, plus the reconstruction at `test_identifiers.py` l.239–245; <br>(b) `unresolved.extend` dropped: runtime coverage of `range(n)`, plus the exact unresolved list; <br>(c) scope built from `record_issues`: runtime entry-presence check; <br>(d) wrong count or ranges: runtime check plus `test_planted_emitted_range_loss_refuses` (l.264–269), which also catches deleting the call; <br>(e) familiarity group lines written twice: list equality at l.369. <br>The multi-record fixture (l.116–133) meets the round-2 fixture requirements, and the exact source table, unresolved entries and weighted `unresolved_by_reason` are asserted. <br>Two parts of the item-3 oracle are missing (R3-01): <br>• the group reconstruction drops the group header fields; <br>• an earlier familiarity assertion was removed. |
| R2-02 | **Fixed.** | `test_independent_sources_digest_and_canonical_input_refusals` (l.409–434) covers each guard separately. <br>• `--frozen` and `--sources` are each copied byte-for-byte to `relocated/<same basename>`, with the hashes asserted equal. Each must refuse `noncanonical` with no `journal/` directory; `Journal` creates that directory (`preparation.py` l.115–118), so the absence check is meaningful. <br>• The frozen digest is asserted correct and only the sources digest is set to `'0'*64`. The run must refuse `signed recipe`, leave terminal statuses of exactly `['refused']` with `failure_reason == 'signed recipe pin mismatch'`, and create no `identifiers/`. <br>If either loop skipped an argument, the run would publish, because the basename-keyed digest and the 24,254-row source pass. The test would then fail. `signed source universe mismatch` cannot match the `signed recipe` regex. The older test still expects `['refused','refused']`. | the named test |

Round-1 items R1-02 to R1-05 stay as dispositioned in round 2. R3-01 reopens part of the R1-05 test evidence.

## New issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R3-01 | minor | **Code:** `review_identifiers.py` l.186–199 (group header construction; `count_table` and `familiarity` accumulation) and l.215–223 (`by_test_repository_study_kind_rule`, `familiarity_sources`). <br>**Tests:** `test_identifiers.py` l.219–262 and l.351–372. | **What is untested.** The emitted group headers and the familiarity and count tables have no test oracle. An earlier familiarity assertion was removed without disclosure. <br>• R2-01 item 3 asked for an "exact ordered reconstruction of the groups". `assert_multirecord_encoding` keeps only `output`, `kind` and the (index, value, source_id) triples. It drops `study`, `task`, `repository`, `matching_rule` and `applies_to_test_repositories`. Applicability is the field the familiarity log feeds into freshness exclusion. <br>• No current test asserts `counts['familiarity_sources']` or `counts['by_test_repository_study_kind_rule']`. R1-05 required both tables. <br>• The round-1 response's R1-05 row says `test_group_conservation_pin_once_and_source_table` "checks emitted counts … and nonempty Prometheus familiarity sources". Round 2 (suggestion 2) cites "the Prometheus assertion at l.213–214" of that test. The round-3 rewrite of that test removed the assertion, and the round-2 response does not mention the removal. This weakens a safeguard. <br>**Failure cases.** `validate_emission` checks only index/value pairs, and the publication test compares the manifest with familiarity lines that carry the same corruption. Each of these regressions therefore passes every test and the runtime check, and the shell-less output review would see consistent but wrong tables [INFERENCE: static]: <br>(i) Group header built with `applies_to_test_repositories=list(TESTS)`. The django-only `repository_identity` group, any scoped path or PR, and the empty blob would then claim TypeScript and Prometheus. <br>(ii) `for repo in row['applies_to_test_repositories'][:1]`. Any-scope SHAs would lose TypeScript and Prometheus from `familiarity_sources` and the count table, and those familiarity tables would undercount, suggesting fresher cases than exist. <br>(iii) `count_table[...] = 1` instead of `+= 1`. Commit totals become 1 instead of 2 for the current fixture. <br>(iv) Line 199's `familiarity[repo].add` deleted. Every `familiarity_sources` list becomes empty. <br>(v) A grouping key without `task` or `study`. Different-task records merge under the first record's header. The fixture has one study and one task, so this is not caught. | Tests only; no production change is needed (the current code is correct by inspection). <br>1. Extend the multi-record fixture so that each header dimension is exercised: <br>• a valid django-scoped path or PR (applies to `[django]` only); <br>• a second `task` or `study` within one output, as same-key records that must not merge; <br>• a Prometheus-specific valid record, for example a `repository='prometheus'` `repository_identity`. <br>Update the seal summary to match. <br>2. In `test_group_conservation_pin_once_and_source_table`, assert exact equality of: <br>• `result['identifier_groups']` as complete ordered dicts with every header field; <br>• `result['counts']['familiarity_sources']`, including Prometheus; <br>• `result['counts']['by_test_repository_study_kind_rule']`. <br>3. In `test_successful_publication_groups_and_pin_once`, compare the published manifest groups and familiarity lines with the same independently written expectations, not only with each other. <br>4. Confirm that each of mutants (i)–(v) fails. |

## Checks for new defects (none found) [INFERENCE: static]

- **No false refusal on real-data shape.** The real data has 50,016 compatible records; 262,080 unattributed records; 23,462 unscoped paths; 0 PRs. <br>• Scoped outputs cannot yield `repository is None`, because a `canonical()` miss is rejected as `unknown_repository`. <br>• Unattributed scope rows and scope issues come from the same branch (`projection.py` l.83–87). <br>• Gap indices are projected indices, matching `produce()`. <br>• Values repeated at distinct indices are compared as a multiset. <br>• Outputs that are empty or fully rejected agree with `range(0)` and produce no entries. <br>• `unknown_output_schema` refuses earlier (l.142–143), and `source_not_identifier_parsed` entries are appended after validation and have no `output` key.
- **Empty blob.** The experiment peer reports a read-only value search showing the empty-blob SHA in neither real identifier output. That search read no outcome fields.
- **Integration.** Unchanged since round 2; the `Journal.finish` field `failure_reason` used by the new test exists (`preparation.py` l.123).

## Suggestions (not issues)

1. Add direct tests for the other `validate_emission` branches: parallel-length mismatch, a missing pair, a duplicate reject, a missing or extra scope entry, `start > end`, and a count-only mismatch.
2. Reword `identifiers-REVIEW_REQUEST.md` l.32. At runtime, `validate_emission` checks only (index, value) pairs and list lengths, not group headers or `source_ids`.
3. Assert the multiplicity of the familiarity `source_table` line. A duplicated first line is not caught today.
4. Round-2 suggestions 1 and 3–8 remain open. Suggestion 2 (separating scoped from any-scope familiarity) is partly addressed by the R3-01 fix.

## Recommendation, alternatives, risks, execution

- **Recommendation.** Fix R3-01 with the test-only changes above, then submit round 4. Do not run the projector until a round reports zero issues.
- **Alternatives considered.**
  - *Extend `validate_emission` to compare fully decoded rows, including headers and source identities, against the projected rows.* This would protect the real artefact at runtime. It is not required, because the header code is deterministic and a test oracle catches regressions before the single run. The cost is a larger new refusal surface on real data, and a refusal there would block the seal.
  - *Accept R3-01 as a suggestion because the current code is correct.* Rejected. A previously present assertion was removed, which weakens a safeguard. Applicability is the field that matters most for freshness, and the calibration treats untested decision boundaries as issues.
- **Risks.** The mutant outcomes come from static reading, not mutation runs. Extending the fixture changes many expected counts, so stale expectations could cause test churn but not false passes. Everything is still before the run and fully reversible: no outputs exist.
- **Execution outline.**
  1. Extend the fixture and summary.
  2. Add exact header, `familiarity_sources` and count-table assertions in both pipeline tests.
  3. Main runs `.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py'`.
  4. Recompute the SHA-256s and regenerate the execution-source manifest.
  5. Submit round 4 together with an author response that discloses every removed or changed assertion.
  6. After a zero-issue round, run the reviewed command once at its canonical paths.

## Verification

- **Read in full:** `review_identifiers.py` and `projection.py` (both as `:raw`), `test_identifiers.py` (`:raw`), `identifiers-REVIEW_REQUEST.md`, the round-2 record (all issue rows read whole) and its response, and the round-1 record's R1-01 to R1-05 rows. The round-1 response's R1-05 row was read whole, and it is the evidence for the removed Prometheus assertion. `preparation.py` `paths`, `operation_lock`, `Journal` and `finish` were checked by grep.
- **Grep of `test_identifiers.py`:** `familiarity_sources` and `by_test_repository` have no matches. `applies_to_test_repositories` and `matching_rule` appear only in `ProjectionTests` row-level assertions (l.149–192).
- **Not run by this round:** tests, the projector, digest recomputation, mutation runs.
- **Main's report:** 121 tests OK, plus clean `go vet` and `go test`. These do not catch R3-01.
- **Peers.**
  - *innovation-challenger* raised R3-01 with mutants (i), (iii) and (iv), and judged R2-01 and R2-02 fixed.
  - *experiment-peer* raised R3-01 independently and found the evidence for the removed assertion. It contributed mutant (ii), suggestions 1 and 2, and the empty-blob check.
  - *My own analysis* found the header gap, including the grouping-key mutant (v) and the applicability concern. I had held it at suggestion level until both peers produced further failure cases and the removed-assertion evidence, which made it an issue.
  - All three agree that R2-01's required runtime check and R2-02 are correctly implemented, and that there is no new production defect, no false refusal on the real data's shape, and no outcome leak.

## Verdict

**1 issue open** (R3-01, minor). The projector may not run yet. A later round reporting zero issues must name the revised SHA-256s, including `test_identifiers.py`.
