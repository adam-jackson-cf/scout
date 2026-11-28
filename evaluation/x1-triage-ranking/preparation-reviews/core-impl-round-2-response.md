# Core implementation round 2 — response

Date: 2026-11-10. **Implementation response, awaiting council re-review; not approval or execution authority.**

Addresses [Core X1C-R2-01](core-impl-round-2.md) and the Core part of [CC2 R2-01](identifiers-delta-cc2-round-2.md). Paths below are relative to `evaluation/x1-triage-ranking/`. This is one integrated Core revision: publication locking and CC2 seal integration are both present before any new real journal is started.

No production clock, input receipt, journal, object receipt, identifier output or adopted configuration was rewritten. No candidate fetch, model call, scanner, CC2 producer, adoption, closure, preflight or real seal was run. Synthetic tests use temporary fixtures only. All edits are uncommitted and were made through absolute paths inside the dedicated worktree; no `.tools` symlink target was edited. Identifiers remained paused while these edits and checks ran.

## X1C-R2-01 — journal publication

`execution/preparation.py:339–348` adds a short exclusive publication lock on a read-only descriptor of the `journal/` directory. It creates no lock file or unaccounted output. Closing the descriptor releases the lock, including on exception.

- `Journal.__init__` (`:353–371`) first holds the existing operation lock, then takes the publication lock around the entire predecessor glob/pin snapshot and the exclusive start-record publication.
- `Journal.finish` (`:373–381`) likewise acquires the operation lock first, then holds the publication lock through terminal publication.
- The shared operation lock is never upgraded. Operation bodies and the five repository fetches remain parallel; only predecessor snapshots and record publication serialize.
- `verify_inputs.py` and its create-once/fsync implementation are unchanged. The fix prevents a live overlapping publisher from exposing a partial record to another snapshot; it does not claim to recover pre-existing interrupted operations.

### Catching regressions

`execution/tests/test_journal_publication.py` spawns independent processes. Writer A pauses after exclusively creating its start or terminal file, writing and fsyncing only its first byte. Writer B signals that it has obtained the shared operation lock. The test then checks that B cannot reach the target predecessor pin or finish while A is paused. After releasing A, B's stored predecessor pin must equal A's final bytes, the complete start tree must verify, and OUT must contain only the expected operation lock, clock and journal directory.

- `test_start_publication_blocks_predecessor_snapshot`
- `test_terminal_publication_blocks_predecessor_snapshot`
- `test_shared_operation_lock_alone_reproduces_both_publication_races`: disables only the new publication guard inside the child processes. Both start and terminal interleavings then trigger the exact assertion `writer B pinned writer A during publication`. This executable counterfactual confirms the two regressions catch the pre-fix shared-lock-only mechanism, rather than merely asserting successful final writes.

Final dedicated result: **3 tests passed**; both overlapping publication cases and both counterfactual interleavings were exercised.

## CC2 R2-01 — integrated Core part

`execution/preparation.py:532–593` adds mandatory `collect_cc2_delta`; `collect_seal` invokes it at `:732–735`, after original canonical producer validation and before scoped evidence/output accounting. It does not change the clock-pinned `run-config.json`, remove any original identifier requirement, or add a waiver for stale producer bytes.

### Canonical output kinds and chain

The seal now requires these three files and preserves their canonical output roles:

| Canonical path under OUT | Required schema/role |
| --- | --- |
| `identifiers-delta-cc2/manifest.json` | `scout.x1.identifier-delta-cc2.v1` |
| `identifiers-delta-cc2/coverage.json` | `scout.x1.identifier-delta-cc2-coverage.v1` |
| `identifiers-delta-cc2/familiarity.jsonl` | `scout.x1.familiarity-delta-cc2.v1` |

Both JSON documents and the JSONL source-table header must have passed status, the canonical clock, the canonical implementation receipt, the signed source-manifest pin and the exact retained baseline journal/output predecessors. Failed/unmeasured evidence or development-freshness filtering refuses. Manifest union/unknown-freshness/non-exclusion flags and the familiarity source-table header must retain the producer's conservative append-union contract. The whole JSONL file, not only its header, is pinned for producer linkage and output review.

The implementation receipt is canonical at `STUDY/preparation-reviews/identifiers-delta-cc2-implementation.json`, scope `identifier-delta-cc2-implementation`. Its authentic council record must bind all **16** current control pins: eight code/test/request files, config, signed source manifest, baseline action-d journal and five baseline outputs. Path/hash pairs, not pooled hashes alone, establish this coverage. The baseline terminal must still record passed action d/exit 0 and exactly its five current output pins.

Exactly one passed action-d journal must pin all three delta outputs in producer order and the implementation/source inputs. The separate output receipt is canonical at `OUT/reviews/identifier-delta-cc2-output.json`, scope `identifier-delta-cc2-output`; its authentic council record must bind every delta output and that passed producer terminal. The implementation and output council records are both retained in the seal's scoped reviews. Earlier `identifier-completeness-delta` or implementation scopes cannot substitute for the new output scope.

`review_receipt` (`:437–447`) accepts an explicit study context for this integration and the seal's configured reviews. This allows canonical implementation receipts outside OUT to be validated against the configured study's review directories; it retains the existing exact verdict, unique scope-block and digest checks. The existing execution-source manifest still must cover every current execution file and verify all bytes before seal collection completes. Required baseline identifiers and their old output roles remain untouched.

### Fixtures and catching tests

`execution/tests/test_binding_seal.py` stages only synthetic CC2 controls and baseline outputs. It uses the real CC2 `publish` function, real Journal, real receipt parser and both authentic-format synthetic council records. CC2 publication precedes canonical binding publication so the existing final-operation recovery regression continues to select the required binding producer, not the newly added delta producer.

`execution/tests/test_cc2_seal.py` covers:

- full seal collection retaining baseline bytes, the original clock, all three canonical delta roles and both additional scopes;
- missing delta output and changed execution-source refusals through the real seal collector;
- missing or wrong-scope implementation/output receipts;
- omission of each of the 16 implementation controls, and each of the three outputs or producer journal from output review;
- schema/status/clock, implementation/source/baseline chains, freshness/failure/unmeasured flags and union semantics;
- familiarity source-table mismatch;
- missing, duplicate or wrong-action producers and omitted producer source pins;
- baseline output substitution, refused by the exact immutable-pin check before any delta acceptance.

### Freeze order and remaining review work

The Identifiers owner's pre-pause edits already added the mandatory review order, exact verdict/scope-block rule and no-bound-edits-after-run rule to `preparation-reviews/identifiers-delta-cc2-REQUEST.md`. As sole writer, I completed its `:79` integration paragraph with the final helper pins and implemented contract so its existing helper-pin regression can run against the final bytes. The request now names:

- `preparation.py`: `7f8d390fd3ebb50140b5725d7704ccf75b5d157fd74fa73810864a625962753e`.
- unchanged `verify_inputs.py`: `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`.

No further helper-pin edit is needed unless Core changes again. Identifiers must now complete its round-2 response using these digests and the validation evidence below. Main must regenerate the full 16-pin confirming packet, obtain zero-issue Core **and** CC2 seal-integration sign-offs naming these same helpers, then obtain the confirming CC2 implementation round/receipt at the final bound bytes before the single producer run. Output review follows that run. This response supplies none of those approvals. No bound file may change between the producer run and seal.

## Validation history

The first publication harness attempt timed out waiting for writer A. The fixture was corrected to provide an explicit existing worker entry path and surface child failures; no production workaround was added. The initial full suite found two fixture/assertion failures: baseline tampering correctly hit the immutable pin check before a later baseline comparison, and adding CC2 after binding changed which operation an existing final-operation test selected. The baseline assertion now names its actual specific pin refusal; canonical required-producer checks run before delta collection, and synthetic CC2 is published before the binding pair. An intermediate full run still caught the final-operation ordering problem; after that fixture reorder the complete suite passed. No assertion was broadened to suppress the required refusal.

## Complete final commands and output

Working directory: `/Users/adamjackson/Projects/work/scout-x1prep`. Python used the pinned worktree interpreter with `-I -B`. Go used the attested 1.26.10 executable, not the ambient shim, with `PATH=/Users/adamjackson/.local/share/mise/installs/go/1.26.10/bin:/usr/bin:/bin`, `GOTOOLCHAIN=local`, `GOENV=off`, `GOWORK=off`, `GOFLAGS=-mod=readonly`, `CGO_ENABLED=0`, `GOPROXY=off`, `GOSUMDB=off` and `GOCACHE=/Users/adamjackson/Projects/work/x1-preparation-scratch/core-review-go-cache`. HOME/TMPDIR were inherited if present; Go dependency acquisition was disabled.

```sh
/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p test_journal_publication.py
```

```text
...
----------------------------------------------------------------------
Ran 3 tests in 1.594s

OK
exit_code: 0
```

```sh
/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests
```

```text
...............................................................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 287 tests in 7.516s

OK
exit_code: 0
```

```sh
/Users/adamjackson/.local/share/mise/installs/go/1.26.10/bin/go vet -pgo=off ./evaluation/x1-triage-ranking/execution/model_free
```

```text
(no output)
exit_code: 0
```

```sh
/Users/adamjackson/.local/share/mise/installs/go/1.26.10/bin/go test -pgo=off -count=1 ./evaluation/x1-triage-ranking/execution/model_free
```

```text
ok  	scout/evaluation/x1-triage-ranking/execution/model_free	0.560s
exit_code: 0
```

## SHA-256 packet

Every file changed by Core in this round is listed below; the response's own digest is reported separately to avoid self-reference. The Identifiers owner's previously edited `test_cc2_delta.py` was exercised but not changed by Core. Files from round 1 not listed as changed remain as previously reported.

| File | SHA-256 |
| --- | --- |
| `execution/preparation.py` | `7f8d390fd3ebb50140b5725d7704ccf75b5d157fd74fa73810864a625962753e` |
| `execution/tests/test_journal_publication.py` | `6d8a0f1a64370dc0775e3876ff539ca60510323d16ebdede87e3b07aa6a3f956` |
| `execution/tests/test_cc2_seal.py` | `91e33dcac2d9ab6fcf0a9160cc01755b0fbb051f9c9f56e4e6504136f6190f82` |
| `execution/tests/test_binding_seal.py` | `7c05eaf2f577210361bbd18a93087e7f59b8fffa687c09d98fad2824d8f4b593` |
| `preparation-reviews/identifiers-delta-cc2-REQUEST.md` | `402f0a09c6209fd37014d990378340cba7fa7d3d6aca592d7f1d0586bbc0a4e2` |
| `execution/verify_inputs.py` (unchanged anchor) | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` |
| `execution/run-config.json` (unchanged anchor) | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |

**Status:** implementation and required checks complete; awaiting council review at these bytes. All changes remain uncommitted. No real preparation action is authorised by this response.
