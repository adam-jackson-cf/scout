# Identifier implementation round 3 — response

R3-01 is fixed with tests only. No production code, config, integration helper, signed input or previous review record changed. No real projector run occurred. This response is not sign-off.

Paths below are relative to `evaluation/x1-triage-ranking/`.

## Required fix and evidence

* `execution/tests/test_identifiers.py:117–137`: the multi-record fixture now includes a django-only scoped path; same-output, same-identifier records differing only in task and, separately, only in study; and a valid Prometheus-scoped repository-identity record in its own `prometheus-study`. The fabricated compatible-record summary increases to 6. The additional output now has 4 records. Rejects, repeated source references, distinct unattributed sources and non-adjacent unattributed paths remain.
* `:302–323`, `expected_groups`: independently spells out the complete ordered group dictionaries, including every header (`output`, `study`, `task`, `repository`, `kind`, `matching_rule`, `applies_to_test_repositories`) and every parallel index/value/source-ID list. Expected django-only and Prometheus-only familiarity groups cannot masquerade as any-scope groups.
* `:241–282`, `assert_multirecord_encoding`: asserts exact equality of the complete groups, every repository's ordered `familiarity_sources`, and the complete ordered repository × study × kind × rule count table. Prometheus has an explicit `additional-identifiers.json / prometheus-study / other` source, distinct from its any-scope SHA sources. Counts include multiplicity (for example, old-study commit count 3 per test repository), not presence alone.
* `:223–239`: `test_group_conservation_pin_once_and_source_table` calls that complete oracle and retains its pin-once, exact source-table and conservation assertions.
* `:436–459`: `test_successful_publication_groups_and_pin_once` now checks both the published manifest and the complete familiarity-group line list against the independent oracle, rather than relying only on agreement between two potentially identically corrupted outputs. Existing pin-once, unresolved and multiplicity checks remain.
* `:325–347`: `test_required_header_and_table_mutants_fail_oracle` executes the requested mutants in memory against fabricated inputs. It compiles only a mutated `produce` function in a private namespace; it does not edit or replace any source file or import raw data. Each mutation target must occur exactly once. Each mutant produces a result and must raise `AssertionError` in the independent complete oracle.

### Mutation results actually exercised

All six subtests passed, meaning each planted mutant was rejected by the oracle (not merely predicted to fail):

| Requested mutant | Planted mutation | Observed oracle result |
|---|---|---|
| (i) | Every group header uses `list(TESTS)` applicability | Assertion failed as required |
| (ii) | Familiarity/count accumulation uses only the first applicable repository | Assertion failed as required |
| (iii) | Count accumulation assigns 1 instead of incrementing | Assertion failed as required |
| (iv) | Familiarity-source insertion replaced by `pass` | Assertion failed as required |
| (v), task | Task dimension normalised to `other`, merging same-key task variants | Assertion failed as required |
| (v), study | Study dimension normalised to `old-study`, merging same-key study variants | Assertion failed as required |

## Disclosure of removed or changed assertions

The round-2 revision did remove the earlier `assertIn(TESTS[2], result['counts']['familiarity_sources'])` and `assertTrue(result['counts']['familiarity_sources'][TESTS[2]])` assertions while replacing that test's body with the multi-record helper. That omission was not disclosed in the round-2 response. It weakened the safeguard, and the prior response's R1-05 evidence claim was consequently stale. This revision restores and strengthens the safeguard with exact whole-table equality, including a genuinely Prometheus-specific source rather than only any-scope SHA familiarity.

Every assertion changed in this revision is listed here:

1. Replaced the partial ordered reconstruction assertion (output, kind and zipped triples only) with exact complete group-dictionary equality. No header is discarded now.
2. Added exact `familiarity_sources` and count-table assertions; these supersede the omitted coarse Prometheus assertions.
3. Updated exposure conservation from input/projected 3/2 to 6/5 and additional conservation from 3/2 to 4/3; both still require one record-level rejection. Unattributed conservation remains 5/4 with one rejection.
4. Updated exposure provenance ranges to `[[0,0],[2,5]]`, additional provenance ranges to `[[0,0],[2,3]]`, respective gap counts to 5 and 3, and total weighted provenance gaps from 8 to 12. Unattributed ranges/counts and all rejection-reason checks remain.
5. Retained manifest-versus-familiarity list equality and added independent complete-oracle assertions for each. Existing list multiplicity, source-table and pin-once assertions were not removed.
6. Added six mutation subtests with exact single-target checks and required assertion failures.

No other assertion was removed or weakened. The review request and production implementation are unchanged in this round.

## Verification command and complete output

Working directory: `/Users/adamjackson/Projects/work/scout-x1prep`.

```sh
.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py'
```

Complete process output:

```text
..........................................................................................................................
----------------------------------------------------------------------
Ran 122 tests in 5.717s

OK
```

Exit code 0; tool-reported wall time 5.93 seconds. This includes the six in-memory mutation subtests. All execution unit tests passed; no real identifier projection, candidate scan, network operation or model call was performed by this assignment. No Go source changed, and no Go check was run for this test-only fix.

## Recomputed SHA-256s

```sh
shasum -a 256 evaluation/x1-triage-ranking/execution/review_identifiers.py evaluation/x1-triage-ranking/execution/identifiers/projection.py evaluation/x1-triage-ranking/execution/tests/test_identifiers.py evaluation/x1-triage-ranking/preparation-reviews/identifiers-REVIEW_REQUEST.md evaluation/x1-triage-ranking/execution/preparation.py evaluation/x1-triage-ranking/execution/run-config.json
```

Complete process output:

```text
adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2  evaluation/x1-triage-ranking/execution/review_identifiers.py
2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072  evaluation/x1-triage-ranking/execution/identifiers/projection.py
6b7f64b3d283180fae29c4bd54eb22f949e10bc9c85342f08f81d440053ea371  evaluation/x1-triage-ranking/execution/tests/test_identifiers.py
10879ebd6a61029fb88bafd5924af6fb154b7a7571a33c651d6849c0641210f9  evaluation/x1-triage-ranking/preparation-reviews/identifiers-REVIEW_REQUEST.md
1a116d86b98b088c474a0f4e974af9be8299484ac1deee99dba8f3e7123bc9bf  evaluation/x1-triage-ranking/execution/preparation.py
c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54  evaluation/x1-triage-ranking/execution/run-config.json
```

Exit code 0. Only the test file in that list changed this round. This response was newly written after the digest command and does not include a circular self-digest.

## Handoff

Regenerate complete execution-source coverage and submit for round 4. Do not run the real projector until zero-issue implementation sign-off. The subsequent real-output completeness/delta review and its unresolved-provenance decisions remain separate requirements.
