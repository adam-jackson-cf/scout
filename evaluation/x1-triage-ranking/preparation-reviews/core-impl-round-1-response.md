# Core implementation round 1 — response

Date: 2026-11-09. **Implementation response; awaiting council re-review. This is not sign-off.**

Responds to [core-impl-round-1.md](core-impl-round-1.md). Paths below are relative to `evaluation/x1-triage-ranking/`. The original configuration, documentary verifier, signed study inputs and create-once evidence were not rewritten. No real fetch, bootstrap, scanner, identifier extraction, adoption, closure or seal was run for these fixes.

## Issue disposition and catching tests

All listed tests passed in the final full suite. New regressions are in `execution/tests/test_preparation_evidence.py`.

| Issue | Fix and production anchor | Catching tests |
| --- | --- | --- |
| X1C-R1-01 | `execution/preparation.py:423` requires the exact standalone bold verdict line, one matching explicit scope block, all reviewed digests in that block, and a record inside this study's `reviews/` or `preparation-reviews/`. `:502` and `:638` retain separate scope-to-evidence sets; implementation, development boundaries, binding and identifiers cannot borrow one another's digests. | `:41–87`: exact acceptance; authentic identifiers-round-5 and binding-v2-round-4 nonapproval records; negated verdict; other scope; digest in another scope; authority outside the study review directories; cross-scope pooling. |
| X1C-R1-02 | `execution/preparation.py:570` copies documentary pins instead of aliasing the gate list. `:663` therefore reports exactly documentary predecessors, not appended runtime or output pins. | `CoreSealReviewTests.test_r102_predecessors_are_exactly_documentary_gate_pins` (`:286`) checks the exact predecessor list and that the gate object remains unchanged. |
| X1C-R1-03 | `execution/preparation.py:518` refuses unreviewed failed/refused operations. Explicit `operation-recovery` authority must bind the immutable start, terminal and every output pin; recovered failures remain deviations, not passed operations. New starts pin prior journal files, detecting removal of an earlier pair. `:649–654` requires each canonical required record to have a passed producer. `execution/fetch_objects.py:141` refuses an existing object receipt before constructing a Journal. | `:125–167`: unreviewed failure, reviewed recovery, missing recovery-output pin, removed predecessor pair, duplicate fetch without Journal. `:300` refuses a canonical record attributed to a failed operation. |
| X1C-R1-04 | `execution/preparation.py:394` excludes historical whole-file observations for the five allowed shared paths while preserving immutable pins and deferred-submodule handling. `:569` still obtains a fresh documentary gate, enforcing current owned extracts and permitted transitions. `verify_inputs.py` is unchanged. | `test_r104_shared_whole_file_observations_can_change` (`:169`) permits shared-file drift but refuses an immutable-file byte change. Existing verifier exact/exception/owned-extract tests also passed. |
| X1C-R1-05 | `execution/preparation.py:237` supports an existing-lock read-only open. `:670–673` and CLI `seal --check` perform complete collection under an exclusive lock without a reservation, journal or writes. `RESULTS.md` requires the preflight immediately before real seal and retention of the pinned Git keg. | `test_r105_seal_check_has_no_reservation_journal_or_writes` (`:179`) compares bytes and mtimes and asserts no Journal or exclusive JSON writer. `brew pin git` succeeded, output below. |
| X1C-R1-06 | `execution/preparation.py:324` and `:339` capture the entry script and loaded execution-module pins at each new operation start. `:518` revalidates them and emits `producer bytes unrecorded` deviations for legacy passed journals. `RESULTS.md` names all eight historical operation IDs; their records are retained, not backfilled with later source bytes. | `test_r106_new_journal_pins_entry_and_loaded_execution_modules` (`:189`) and `test_r106_legacy_producer_gap_disclosed` (`:199`). |
| X1C-R1-07 | Added refusal-specific assertions rather than broad success expectations. The minimal canonical-record fixture in `execution/tests/test_core.py:140` now supplies ROOT and names the expected missing-file/schema failures. Shared binding/closure fixtures use canonical review locations, exact verdicts and explicit scope blocks; binding seal fixtures link closure to a passed producer and review current canonical evidence in the correct scope. | `test_r107_*` at `:207–276`, `:296` and `:311`: unfinished journals; id/stem; stale clock; before-clock and reversed timelines; start/terminal mismatch; required action; noncanonical start; adoption reservation, unfinished operation, duplicate adoption and filename order; first-seal refusal preserving the empty reservation and refused e journal; unaccounted output; clock create-once, canonical path, schema/study, execution commit, first action, deadline, gitlink, authority/manifest/config pins, future clock, input-receipt path and hash. Existing Core adoption/hash/order tests and full binding/closure/identifier suites passed. |

## Executed evidence and producer provenance

The documentary verifier was **not changed**. The fetcher **was changed**, solely to move the existing-receipt guard before Journal creation. The Core changes do not rewrite or rerun the historical input gate or the five executed action-(a) records. Their validity is retained as stated by the round-1 review; this response does not claim that current producer bytes produced them. `RESULTS.md` discloses the five a, two b and one d legacy journals with unrecorded producer bytes. Future starts capture producer pins automatically.

An initial integrated test attempt reported 272 tests with 1 failure and 20 errors: synthetic binding verdicts omitted the literal Markdown stars and the minimal Core record fixture lacked ROOT. Those fixtures were corrected, followed by a passing 272-test run. Additional clock/reservation refusal coverage brought the final suite to 273 tests. A test-log write initially hit a skill prerequisite; that write was retried after reading the required guidance, without rerunning the successful test command for that reason. The new regression file is named for preparation evidence rather than its review ticket.

## Complete final validation commands and output

All Python tests used the worktree's pinned interpreter with `-I -B`; no network candidate repositories were fetched. Working directory for the following commands was `/Users/adamjackson/Projects/work/scout-x1prep`.

### Python execution suite

```sh
.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests
```

```text
.................................................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 273 tests in 24.289s

OK
exit_code: 0
```

Full merged stdout/stderr is retained at `/Users/adamjackson/Projects/work/x1-preparation-scratch/core-r1-python-suite-final.txt`.

### Pinned Go frontend

The Go executable is the signed runtime-attestation's Go 1.26.10 binary, not the ambient mise shim. Both commands used `PATH=/Users/adamjackson/.local/share/mise/installs/go/1.26.10/bin:/usr/bin:/bin`, `GOTOOLCHAIN=local`, `GOENV=off`, `GOWORK=off`, `GOFLAGS=-mod=readonly`, `CGO_ENABLED=0`, `GOPROXY=off`, `GOSUMDB=off`, and `GOCACHE=/Users/adamjackson/Projects/work/x1-preparation-scratch/core-review-go-cache`; HOME and TMPDIR were inherited if present. Dependency/network acquisition was disabled.

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
ok  	scout/evaluation/x1-triage-ranking/execution/model_free	0.591s
exit_code: 0
```

Full merged stdout/stderr and the explicit Go settings are retained at `/Users/adamjackson/Projects/work/x1-preparation-scratch/core-r1-go-suite.txt`.

### Git keg retention

```sh
brew pin git
```

```text
(no output)
exit_code: 0
```

Do not upgrade or clean up this Git keg before seal. Main must run `seal --check` immediately before any real seal, using the commands in `RESULTS.md`; neither command was executed here. A passing synthetic suite is not authority to adopt, inventory or seal.

## Recomputed SHA-256

These are file-byte digests, not an approval statement. The original config, verifier and proposal remain unchanged.

| File | SHA-256 |
| --- | --- |
| `execution/preparation.py` | `858031133e1d28ecf71cb9b4e79b64df9ad03f572c2ab4eca32d5fc48f019255` |
| `execution/fetch_objects.py` | `64bb3dd3b99d5227d0abd8e460f4dbabda18e15b37c16c71358241566e8c6a9f` |
| `execution/verify_inputs.py` (unchanged) | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` |
| `execution/run-config.json` (unchanged) | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |
| `execution/run-config-amendments-proposed/001-binding-v2.json` (unchanged) | `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121` |
| `execution/tests/test_core.py` | `7bdc801955675437a095121b28cfbf2d85ca75c004abbe870f1e14ca5a1d2628` |
| `execution/tests/test_preparation_evidence.py` | `634f5242d5d4190b5373bc741398a834cd3ee5296ec86a59b4748ce74b080aeb` |
| `execution/tests/test_binding_adoption.py` | `0396765fc3d51de609bf2d935d5b0b8979cd3b84ff65c5f12080e384912ce432` |
| `execution/tests/test_binding_seal.py` | `4c3b58ea22d1325e439efdef19842a530ba5fb158f6dc61ff9c9d73cd0c50637` |
| `execution/tests/test_closure_review.py` | `534e4754f6100a69e60c7e12b6e7740a39fc84c8f0d81b094ab4b19f26cb2906` |
| `execution/tests/test_closure_round2.py` | `70efac053f0b92b5686455481bab1cd074874d96019091bb493f76cff465f274` |
| `RESULTS.md` | `21b6fc131a131e09e98c231fd9bd1792a8ff1b961def1fd7cba06c9565931c43` |
| Scratch `core-r1-python-suite-final.txt` | `31b95b238946e29d93c81ee9ec0c86b354665fd57fae563d2a3071b2a2ada81b` |
| Scratch `core-r1-go-suite.txt` | `f006b2e36ff7350749acbef30ea8838a321c13db1cf9b33c024151b4de0395ca` |

The response document's own digest is reported separately to avoid a self-referential digest. Changes remain uncommitted. A council re-review must cover the new production bytes and fixtures; this response supplies no authority receipt.
