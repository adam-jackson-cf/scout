# Core implementation round 3 — response

Date: 2026-11-10. **Implementation response; awaiting Core round 4 and a CC2 confirming round. Not sign-off or execution authority.**

Responds to [X1C-R3-01](core-impl-round-3.md). Paths below are relative to `evaluation/x1-triage-ranking/`.

## Required fix

`execution/preparation.py:357–371` now acquires the journal-directory publication lock **before** the `RunConfig` configuration re-check. It holds that same lock continuously through `load_config`, effective-configuration comparison, predecessor glob/pins and exclusive start publication. The ordering remains shared operation lock, then exclusive publication lock. There is no operation-lock upgrade.

No lock was added inside `load_config`; there is no second directory descriptor/recursive flock acquisition. No parse exception is caught or suppressed, and binding reservation code was not changed. Terminal publication and parallel operation bodies retain their prior behavior. This is only a lock-scope correction; the CC2 seal implementation is otherwise unchanged. Standalone initial CLI loads remain outside the publication guard as round 3's nonblocking S1 describes.

## Catching regression and counterfactual

`execution/tests/test_journal_publication.py` retains the previous start/terminal snapshot regressions and shared-lock-only counterfactual, and adds:

- `test_run_config_reload_waits_for_start_publication` (`:168`): a minimal valid config is loaded through the real `prep.load_config` before either process starts, and the real `RunConfig` is passed into both spawned workers. Its clock has the canonical config's `execution_config` pin. Worker A pauses after creating/fsyncing only the first byte of its start record. B reserves an empty synthetic binding output before constructing its Journal, mirroring the binding reservation order. It obtains the shared operation lock but waits for A's publication; after release, the actual loader re-check runs, B completes, its predecessor pin equals A's final bytes, its start tree verifies, and the synthetic canonical output contains the passed JSON record rather than remaining empty. The child asserts the config's `RunConfig` type and that `load_config` was actually called. The loader is observed, not replaced with a stub; only the existing temporary-clock validation seam is mocked.
- `test_unlocked_run_config_reload_reproduces_parse_failure_and_orphan` (`:171`): a wrapper deliberately invokes the real reload just before acquiring the unchanged publication guard, reproducing the reviewed ordering. With A still paused and its file still exactly `{`, B terminates with **JSONDecodeError**, before predecessor pinning. B has no start or terminal journal, and its already-created synthetic canonical output remains exactly empty. This is an executable counterfactual of the reported irreversible failure, confined to a temporary directory; it is not a run of the real binding controller.

The original publication tests still pass. The dedicated suite now has **5 passing tests**; the full execution suite has **289 passing tests**. The new counterfactual waits for B's actual parse refusal while A remains paused, rather than inferring the failure from a missing final digest.

## REQUEST pins and CC2 confirming review

Only the necessary REQUEST entries changed:

1. `preparation-reviews/identifiers-delta-cc2-REQUEST.md:79`: final preparation helper digest.
2. `:109`: helper digest and byte count in the 16-row table.
3. `:107`: the request's external self-pin pointer now names this response instead of the obsolete round-2 response; the actual request pin is in the table below.

No other REQUEST prose or bound implementation file changed. The helper is now **50,257 bytes**, SHA-256 `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38`. `verify_inputs.py` remains unchanged at `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`.

The existing CC2 round-3 sign-off at `7f8d390f…` does **not** authorize this revised helper/request pair. After Core round 4 reports zero issues, Main/Identifiers must bring the current 16-pin packet (including this request's new pin) to the CC2 confirming round and create the implementation receipt only after that confirmation. Do not modify historical sign-offs or receipts. No real journal may start before the required reviews pass; bound bytes remain frozen from the subsequent producer run through seal.

## Complete validation commands and output

All commands ran from `/Users/adamjackson/Projects/work/scout-x1prep`. Python used the worktree's pinned interpreter with `-I -B`. Go used the attested 1.26.10 binary, with `PATH=/Users/adamjackson/.local/share/mise/installs/go/1.26.10/bin:/usr/bin:/bin`, `GOTOOLCHAIN=local`, `GOENV=off`, `GOWORK=off`, `GOFLAGS=-mod=readonly`, `CGO_ENABLED=0`, `GOPROXY=off`, `GOSUMDB=off` and `GOCACHE=/Users/adamjackson/Projects/work/x1-preparation-scratch/core-review-go-cache`. HOME/TMPDIR were inherited if present. No dependency acquisition was enabled.

```sh
/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p test_journal_publication.py
```

```text
.....
----------------------------------------------------------------------
Ran 5 tests in 3.537s

OK
exit_code: 0
```

```sh
/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests
```

```text
.................................................................................................................................................................................................................................................................................................
----------------------------------------------------------------------
Ran 289 tests in 34.199s

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
ok  	scout/evaluation/x1-triage-ranking/execution/model_free	0.779s
exit_code: 0
```

## Final file-byte pins

These are the three existing files changed in this round. This response is the only new file; its own SHA-256 is reported separately to avoid self-reference.

| File | Bytes | SHA-256 |
| --- | ---: | --- |
| `execution/preparation.py` | 50257 | `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38` |
| `execution/tests/test_journal_publication.py` | 8900 | `7ea4fa923b963993067782f4aad5a13aaa9d2c0159c7dc49b5112cb5ea97ee0c` |
| `preparation-reviews/identifiers-delta-cc2-REQUEST.md` | 20873 | `ba3e5816eadbda4af7a541e66ea6e75fc0d525cd2c3e8a439608c45be4856119` |

No production clock, receipt, journal, adopted amendment, baseline identifier file or other create-once artifact was changed or created. No candidate fetch, bootstrap, scanner, model, real identifier producer, adoption, closure, preflight or seal ran. Tests used throwaway local fixtures only. The previous executed evidence was not rerun. Changes remain uncommitted and await review.
