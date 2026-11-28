# X1 preparation (d) identifier projector — implementation review — round 4

**Scope.** This round reviews only the test-only fix for R3-01. The production code has not changed since round 3, and round 3 cleared it apart from R3-01. This round does not sign off identifier completeness or the delta. That needs a separate review of the retained outputs. It grants no authority for inventory, SRS, models, development, full-OC or GUIDE. No files were edited. The projector was not run. The only model calls were the two council peers.

## Reviewed files

Main computed these SHA-256 values with `shasum -a 256`. This round did not recompute them. Paths are under `evaluation/x1-triage-ranking/`.

| File | SHA-256 | Change since round 3 |
|---|---|---|
| `execution/review_identifiers.py` | `adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2` | unchanged |
| `execution/identifiers/projection.py` | `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072` | unchanged |
| `execution/tests/test_identifiers.py` | `6b7f64b3d283180fae29c4bd54eb22f949e10bc9c85342f08f81d440053ea371` | changed (was `955bc57a…`) |
| `preparation-reviews/identifiers-REVIEW_REQUEST.md` | `10879ebd6a61029fb88bafd5924af6fb154b7a7571a33c651d6849c0641210f9` | unchanged |
| `execution/preparation.py` (integration) | `1a116d86b98b088c474a0f4e974af9be8299484ac1deee99dba8f3e7123bc9bf` | unchanged |
| `execution/run-config.json` (integration) | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | unchanged |
| `preparation-reviews/identifiers-impl-round-3.md` | `4ae11f6c550c808aaa1518e87dfc682ef97e20f01b8e43136dbe199527338821` | prior record |
| `preparation-reviews/identifiers-impl-round-3-response.md` | `8cf66ac6a8c000ee0b0811a2d72ec017c1024fbd62fe7aed11ac7de6fdafc370` | author response |

## Basis (unchanged)

| Document | SHA-256 |
|---|---|
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |
| `PREREGISTRATION.md` | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| `reviews/rev9-digests.json` | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |
| `preparation/exposure-recipe.json` | `c065347e08a8571228764dae19fc686f3a6909e612a4119aad7bb58d86c5d383` |
| `preparation/exposure-sources.json` | `67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60` |
| `preparation/frozen-exposure-inputs.json` | `ec92fba6d3f91c671301244539a8caefc5025c0d6b03efca77c5bdfc4bd96425` |

## R3-01 disposition: **fixed**

Line numbers below refer to `test_identifiers.py` unless another file is named.

| Required fix | Evidence |
|---|---|
| **1. Fixture covers each header dimension** | `multirecords` (l.117–137) adds these records: <br>• `exposure[5]`, a django-scoped path, giving `repository_scoped_path_unknown_mechanism` with applicability `[django]` only; <br>• `exposure[3]` (`task='association'`) and `exposure[4]` (`study='second-study'`). Apart from that one field, each matches `exposure[0]`/`[2]`, so they are same-key records that must not merge; <br>• `additional[3]`, `repository='prometheus'`, in `prometheus-study`, giving applicability `[prometheus]`. <br>The seal summary is updated to `compatible_records=6`. `unattributed_records=5` and `unscoped_by_kind` (blob 3, path 2) are consistent with how `produce` counts every unattributed record (`review_identifiers.py` l.170–178). |
| **2. Exact oracle in `test_group_conservation_pin_once_and_source_table`** | l.223–239 calls `assert_multirecord_encoding` (l.241–300). That helper asserts exact equality of: <br>• the complete ordered group dicts from `expected_groups` (l.302–323), with all seven header fields plus the index, value and source-ID lists; <br>• all three `familiarity_sources` lists, including the Prometheus-only `(additional, prometheus-study, other)` entry; <br>• the full sorted 15-row `by_test_repository_study_kind_rule` table, with multiplicity (commit = 3 per repository). <br>The expectations are written as literals and are not derived from production code; the only shared values are the constants `TESTS` and `OUTPUTS`. I recomputed every expected value by hand against `projection.py` l.66–80 and `review_identifiers.py` l.181–223 and found no discrepancy. That covers group insertion order; `source_ids` `[0,1]`; familiarity tuple sort order (`additional` < `exposure` < `unattributed`; `old-study` < `prometheus-study` < `second-study`); count-table sort order; the gap ranges `[[0,0],[2,5]]`, `[[0,0],[2,3]]` and `[[0,0],[2,4]]`; and the gap total of 12. |
| **3. Publication test compared against independent expectations** | l.436–459 runs both the published `manifest['groups']` and the familiarity-line groups through the same literal oracle, together with `manifest['counts']`. It keeps the list check `groups == manifest['groups']` (l.454), which catches duplicated lines. |
| **4. Mutants (i)–(v) fail** | `test_required_header_and_table_mutants_fail_oracle` (l.325–347) has six subtests that map to round-3 mutants (i), (ii), (iii), (iv), (v)-task and (v)-study. The harness cannot pass vacuously: <br>• each target string must occur exactly once (l.342); <br>• the mutated source is compiled into a copied namespace (`dict(vars(runner))`), so `runner.produce` is never replaced; <br>• the test calls `namespace['produce']`, the mutated function; <br>• the mutated `produce` runs **outside** `assertRaises`, so a `ValueError` or `validate_emission` refusal would error the test rather than count as a kill; <br>• only the oracle runs inside `assertRaises(AssertionError)`; <br>• the unmutated oracle passes in l.223–239, so an oracle that always fails cannot satisfy the harness. <br>None of the mutants changes the (index, value) pairs, so each result reaches the oracle. The author reports all six subtests executed and passed inside Main's 122-test run. |

**Undisclosed-change check.** I recovered the pre-fix test bytes from the round-3 reviewer's raw read, made at 12:07:08 on 2026-11-09. That is before the author's round-3 edit at 12:17 [INFERENCE: this snapshot is the `955bc57a…` file; its digest was not recomputed]. I compared it against the current file. The differences are exactly the response's disclosed items 1–6:

- the new `inspect` import;
- the fixture and summary changes;
- the conservation changes 3/2→6/5 and 3/2→4/3;
- the new gap ranges and counts, with the total 8→12;
- the full-oracle replacement of the triple reconstruction;
- the added familiarity and count-table assertions, `expected_groups`, the mutation test, and the added manifest-oracle call.

The experiment peer asked whether `unresolved.json`'s own `counts` field had been asserted before. It had not: the prior publication test already passed `counts=manifest['counts']` (prior l.371–372). So that is not an undisclosed weakening. The tail tests (runtime refusals, sources digest, existing root, partial publication, metadata delta) are present in both versions.

## New issues

None.

The new test code adds no production surface. The `exec`/`compile` in the mutation test runs only in memory against fabricated inputs and writes no source file.

## Suggestions (not issues)

1. **Null control for the mutation harness.** Add a subtest that compiles the unmodified source and requires the oracle to pass. It would show directly that the copied namespace behaves the same as `runner.produce`; static reading says it does today.
2. **Two dimensions the fixture still does not exercise.**
   - *Empty applicability:* a valid `EMPTY_BLOB` record, or a fastapi-scoped path. This would pin the rule that rows applying to no repository add nothing to familiarity or counts.
   - *Same key except `repository`, `output` or `matching_rule`:* a second-repository commit in the same study and task, for example. This would catch a grouping key that drops one of those dimensions. Either addition needs `compatible_records` updated.
3. **Per-mutant attribution.** Pair each mutant with the specific assertion it should break, so a future kill cannot come from an unrelated mismatch. Today's broad oracle is adequate.
4. **Published fields not yet checked against the literal oracle.** These are `manifest['unattributed_sources']`, `unresolved.json['counts']`, and the familiarity `source_table` line's content and single occurrence (round-3 suggestion 3), plus the per-line wrappers (`entry_kind`, `unknown_freshness`, `development_freshness_filter`).
5. **Line reference in the response.** It cites `:241–282` for `assert_multirecord_encoding`, which actually spans l.241–300.
6. Round-3 suggestions 1–2 and round-2 suggestions 1 and 3–8 remain open, still as suggestions.

## Recommendation, alternatives, risks, execution

- **Recommendation.** Accept the R3-01 fix. Run the reviewed projector once at its canonical paths, then submit the retained outputs for the separate completeness/delta output review.
- **Alternatives considered.**
  - *Require suggestion 2 (empty-applicability and repository-collision fixtures) before the run.* Rejected. Round 3 did not require it. The mutants it would catch are speculative: the code is correct by inspection and the round-3 grouping key includes every dimension. It also adds churn before a single, reversible-until-sealed run.
  - *Treat the missing `unresolved.json['counts']` assertion as an undisclosed weakening.* Rejected, because the recovered prior bytes show it was never asserted.
- **Risks.**
  - The oracle, fixture and harness were checked statically by all three reviewers. Execution evidence comes from Main and the author (122 tests OK, six mutation subtests); this round did not reproduce it.
  - The prior-bytes comparison relies on a transcript snapshot, not a recomputed digest.
  - Once the run happens, it cannot be repeated without reviewed recovery, because outputs are created exclusively. Any defect would surface at the output review, not silently.
- **Execution outline.**
  1. Main confirms the six reviewed digests are unchanged at run time.
  2. Main runs `.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py'` once more against those bytes.
  3. Main runs the reviewed projector once at its canonical paths with the pinned CPython.
  4. Main retains the outputs and journal and submits them for the completeness/delta output review, including the unresolved-provenance decisions.

## Verification

- **Read in full (`:raw`):** the current `test_identifiers.py` (565 lines), `review_identifiers.py` (l.1–300; the remaining l.301–318 hold only the end of `run()`'s refusal path and the CLI entry, which this test-only round does not depend on), and `projection.py`.
- **Prior records read:** the round-3 record (R2 and R3 rows read whole), the round-3 response, and round-2 response l.1–43.
- **Prior test bytes:** recovered from session log `X1IdentReview3.jsonl` line 25 (the round-3 raw read). The final assertions of the tail tests were confirmed present by grep.
- **Hand recomputation:** of every expected group, familiarity entry, count row, unresolved entry and conservation figure.
- **Not run by this round:** tests, mutants, the projector, digest recomputation.
- **Main's report:** full execution suite, `Ran 122 tests OK`.
- **Peers.**
  - *innovation-challenger:* R3-01 fixed, no new issues. It confirmed the harness is not vacuous and contributed suggestions 2 (output/rule collisions) and 3.
  - *experiment-peer:* R3-01 fixed, and it independently recomputed the oracle by hand. It raised the `unresolved['counts']` question, which this round resolved against the prior bytes, and contributed suggestions 1, 2 (empty applicability, second repository), 4 and 5.
  - *Own analysis:* reached the same conclusion and did the prior-bytes recovery.
  - No peer changed the verdict. The peer's open question was the only point that could have, and the evidence closed it.

## Verdict

**zero issues — projector may be run once at its canonical paths; completeness/delta sign-off remains a separate output review**

**Sign-off.** The innovation council signs off the X1 preparation (d) identifier projector implementation for one run at its canonical paths. The reviewed SHA-256s are:

- `execution/review_identifiers.py` `adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2`
- `execution/identifiers/projection.py` `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072`
- `execution/tests/test_identifiers.py` `6b7f64b3d283180fae29c4bd54eb22f949e10bc9c85342f08f81d440053ea371`
- `preparation-reviews/identifiers-REVIEW_REQUEST.md` `10879ebd6a61029fb88bafd5924af6fb154b7a7571a33c651d6849c0641210f9`
- `execution/preparation.py` `1a116d86b98b088c474a0f4e974af9be8299484ac1deee99dba8f3e7123bc9bf`
- `execution/run-config.json` `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`

The basis is listed above. Any change to these bytes needs a new review round with zero issues before the run.
