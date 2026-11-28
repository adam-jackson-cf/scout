# X1 preparation core implementation review — round 4

**Verdict: 1 issue open (X1C-R4-01, Minor).** Core is not approved at these bytes.
- **X1C-R3-01 is resolved as round 3 required.** The publication lock is held continuously from before the `RunConfig` re-check through the exclusive start write. The new regression and counterfactual catch the defect.
- **The lock change is safe.** It introduces no deadlock, and it adds or moves no reader of journal records outside a lock.
- **`preparation.py` has no open issue.** The only open issue is in the REQUEST, which was revised in the same change: its prose names a stale source for its own pin.

**Scope and method.** This review was read-only. Nothing was run: no code, tests, hashes, Git commands, adoption, preflight, producer or seal. No models were called beyond the two read-only peer analyses. Three analyses fed this record:
- **My own trace** of:
  - `preparation.py`: `load_config`, `operation_lock`, `adopt`, `journal_publication`, `Journal`, `review_receipt`, `verify_journal_configuration`, `seal` and `main`;
  - `test_journal_publication.py`, in full;
  - the reservation order in `binding/acceptance.py:428–448`;
  - `identifiers/cc2_delta.py:157–248`;
  - REQUEST `:79–112`;
  - `identifiers-delta-cc2-round-2-response.md`;
  - every lock acquisition and every reader of the `journal/` directory under `execution/`.
- **`innovation-challenger` (CoreR4Challenger):** lock coverage, the lock order of every caller, an inventory of journal readers, exception safety, and whether the tests could pass vacuously.
- **`experiment-peer` (CoreR4Peer):** for X1C-R3-01, the observation that would show it is not fixed; whether each test fails for its stated reason; deadlock hypotheses; and whether the REQUEST changed only where intended and its pins are consistent.

This record does not contain the exact round-9 verdict line or a whole-line scope marker, so `review_receipt` cannot accept it as authority.

## Reviewed digests

The digests are Main's (`shasum -a 256`). I did not recompute them. Paths are relative to `evaluation/x1-triage-ranking/`.

| File | SHA-256 | Role |
|---|---|---|
| `execution/preparation.py` | `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38` | Under review (changed; 50,257 bytes per the response) |
| `execution/tests/test_journal_publication.py` | `7ea4fa923b963993067782f4aad5a13aaa9d2c0159c7dc49b5112cb5ea97ee0c` | Under review (changed) |
| `preparation-reviews/identifiers-delta-cc2-REQUEST.md` | `ba3e5816eadbda4af7a541e66ea6e75fc0d525cd2c3e8a439608c45be4856119` | Under review (changed: helper pin and 16-row table) |
| `execution/tests/test_cc2_seal.py` | `91e33dcac2d9ab6fcf0a9160cc01755b0fbb051f9c9f56e4e6504136f6190f82` | Unchanged since round 3 |
| `execution/tests/test_binding_seal.py` | `7c05eaf2f577210361bbd18a93087e7f59b8fffa687c09d98fad2824d8f4b593` | Unchanged since round 3 |
| `execution/tests/test_cc2_delta.py` | `4844221de3ecb6c52b81a770ad7dadbcc42d8e0122636a29de2cdb1fc0e9d8d7` | Unchanged since round 3 |
| `execution/identifiers/cc2_delta.py` | `657aeb47ee11cbe4db4fdba3948639cf211f8ff032844cbe6d77a815086c0ea6` | Context (unchanged) |
| `execution/identifiers/cc2_readers.py` | `1ad375a97b43a771245f3f746768dc72814656c1c81c8c4d8793a9c88644144b` | Context (unchanged) |
| `execution/identifiers/cc2_schemas.py` | `db845451e7dbefc8f32ea7b50fbe86cdd5c702330f755e39cee345b2cfc5799f` | Context (unchanged) |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` | Unchanged; executed |
| `preparation-reviews/core-impl-round-3.md` | `2e937a529e44bdf95500abe15ebe4f0e45dc38c754211076c4a3ce900fa3097b` | Prior round |
| `preparation-reviews/core-impl-round-3-response.md` | `900fbcfa5c73401fee4be45ff03275d95a3277ecaf07b1bbe4694b819f04a711` | Author response |
| `preparation-reviews/identifiers-delta-cc2-round-3.md` | `be86e13ef89695314d6a70e51fa86abd46a6dc8fc2ffc0d4c63dd7f188655dc8` | Context |

**Main's test results at these bytes** (recorded output: `core-impl-round-3-response.md`). I did not re-run any of them.
- `test_journal_publication.py`: 5 tests OK.
- Full Python suite: 289 OK.
- `go vet` and `go test -count=1` on `execution/model_free` (Go 1.26.10): clean.

## Disposition of prior issues

| ID | Status | Evidence |
|---|---|---|
| X1C-R3-01 (Moderate) | **Resolved as specified** | See the five required-fix points below. |
| X1C-R2-01, Core part of CC2 R2-01 | Still resolved (round 3) | Unaffected. The terminal publication in `finish` (`preparation.py:374–381`) and `collect_cc2_delta` are unchanged. |
| Round-3 S1–S9 | Still suggestions | Not reopened. S2 gets new supporting evidence in S3 below. |

**How X1C-R3-01 meets each of round 3's five required-fix points:**

1. **The lock now covers the re-check.** In `Journal.__init__`, the order is:
   - the shared operation lock (`preparation.py:355`);
   - `read_clock` (`:356`), which reads only `clock.json` and the files it pins (`:205–227`), nothing under `journal/`;
   - `journal_publication(self.directory)` (`:358`), which takes `flock(LOCK_EX)` on an `O_DIRECTORY` descriptor (`:340–348`).
   
   Everything after that runs inside the lock, which is released only when the `with` block ends after `:371`:
   - the `RunConfig` reload, `load_config(config_path)`, whose start-record glob and parse are at `:149–150`;
   - the effective-configuration comparison (`:359–360`);
   - the id and the base record (`:361–369`);
   - the predecessor glob and pins (`:370`);
   - `exclusive_json` (`:371`), which writes, flushes, fsyncs and closes before it returns (`verify_inputs.py:34–46`).
   
   The order is still operation lock, then publication lock, and the operation lock is never upgraded (`:242–246`).
2. **No lock was added inside `load_config`** (`:105–151`). It takes neither lock, and neither do `verify_adoption`, `review_receipt` (`:437–461`), `verify_journal_configuration` (`:464–482`), `effective_configuration`, `producer_pins` or `pin`. `exclusive_json` opens a second descriptor on the directory only to fsync it, never to flock it. So no process ever takes a second flock on `journal/`.
3. **The parse error is not caught.** No `except` was added, and binding code is unchanged.
4. **There is a `RunConfig` regression plus a counterfactual.**
   - **Regression:** `test_run_config_reload_waits_for_start_publication` (`test_journal_publication.py:168`).
     - The config is loaded by the real `prep.load_config` (`:105`). It is passed to the spawned workers, which assert that it is a `RunConfig` (`:73–74`).
     - Each worker installs its patches in its own process (`:71–72`).
     - B reserves an empty `binding/synthetic.json` with `open('x')` before it constructs the real `prep.Journal` (`:77–80`).
     - `:81` proves that the reload ran.
     - The test asserts that B does not pin A within 0.5 s and is not done (`:127–128`). After release, it asserts that both operations passed (`:145`), that B's predecessor pin equals A's final bytes (`:146–148`), that `verify_tree_pins` accepts B's start record (`:149`), and that the canonical output holds `{"status": "passed"}` (`:151–153`).
     - **On the round-3 ordering**, B would parse A's exposed `{` and raise `JSONDecodeError` a few milliseconds after `ready`. The check at `:128` or the one at `:145` would then fail.
   - **Counterfactual:** `test_unlocked_run_config_reload_reproduces_parse_failure_and_orphan` (`:171`).
     - In B only, a wrapper calls the real loader on the real config path **before** entering the unchanged `original_publication` (`:33–39`), while B holds only the shared operation lock. This is the round-3 ordering.
     - Before releasing A, the test waits up to 10 s for B to actually finish. It then asserts that A's file is still `{` and that B never reached predecessor pinning (`:122–125`).
     - After release, it asserts that B's error is `JSONDecodeError`, that B has no start or terminal record, and that the synthetic output is exactly `b''` (`:137–143`). This is the irreversible orphan.
     - The preferred variant of round-3 point 4, driving a binding-style reservation through `Journal`, is implemented. It is a synthetic stand-in, not the real binding controller.
5. **Digests refreshed.** The REQUEST's integration paragraph (`:79`) and the 16-row table (`:107`, `:109`) were refreshed, but one pointer was missed. See X1C-R4-01.

### Requested checks: deadlock and readers outside the lock

**Deadlock: none.** Every caller takes the locks in this order:

| Caller | Lock order |
|---|---|
| `Journal.__init__` | Blocking shared operation lock, then blocking exclusive publication lock (`:355`, `:358`) |
| `Journal.finish` | The same (`:374`, `:379`) |
| `seal` | Non-blocking exclusive operation lock (`:753`). The nested `Journal` takes the re-entrant `ContextVar` branch (`:242–246`), then the publication lock. |
| `adopt` (`:264`), `start` (`:404`) | Operation lock only, never the publication lock |
| binding (`acceptance.py:435`), CC2 (`cc2_delta.py:237`), identifiers (`review_identifiers.py:258`), closure (`attest_closure.py:836`), the `journal` wrapper (`preparation.py:829`) | Non-blocking shared operation lock, then the nested `Journal` |
| fetch (`fetch_objects.py:135–143`) | Shared operation lock, then a non-blocking per-database lock, then `Journal` |

- Every holder of the publication lock already holds an operation lock, and while holding the publication lock it never waits for any lock.
- So no process can hold the publication lock and wait on a lock held by another process, and there is no cycle.
- If `load_config` raises inside the lock, `finally: os.close(descriptor)` (`:347–348`) releases it. The operation-lock context then unwinds (`:248–256`).

**Readers outside the lock: no new ones.** The change moved the one problematic reader, the reload, **into** the lock. Every remaining reader of `journal/` falls into one of four groups:
- **Under the exclusive operation lock**, which excludes every publisher: `adopt`'s glob (`:286–295`), and `collect_journals` and `collect_seal`, which run only from `seal` (`:596–638`, `:753`).
- **Fixed, completed records:** the `prior_journals` read by `verify_adoption` (`:96–98`), the baseline journal read by CC2 (`cc2_delta.py:157–165`), and `finish`'s own start record (`:378`).
- **Accepted round-3 S1:** the standalone CLI loads, where nothing is reserved yet (`:821`, `acceptance.py:423`, and the other CLI entry points).
- **Binding's window:** between `canonical.open('x')` and `Journal` (`acceptance.py:437–439`), the only unlocked read is `read_clock`, of create-once files that no operation writes.

## New issue

| ID | Severity | Location | Concrete failure | Required fix |
|---|---|---|---|---|
| X1C-R4-01 | Minor | `preparation-reviews/identifiers-delta-cc2-REQUEST.md:99`, which conflicts with `:107` | **The REQUEST misnames where its own pin is recorded.** Line `:99` still says the request's "final SHA-256 and byte count are recorded in `identifiers-delta-cc2-round-2-response.md`". That file records the REQUEST as 20,847 bytes, `c32bd40006bacdf77eb8970d8d4aeaee5be28a6c9bdc243910983c8dba233fb4` (`identifiers-delta-cc2-round-2-response.md:63`), not the current `ba3e5816…`. Row `:107` names `core-impl-round-3-response.md`.<br>• **Introduced by this revision.** At `c32bd400…` the pointer was true. The response says the self-pin pointer was updated (`core-impl-round-3-response.md`, "REQUEST pins"), but only the table row was.<br>• **Permanent once pinned.** The REQUEST is the sixteenth bound control. After the producer's create-once journal pins it, the REQUEST cannot change before the seal (REQUEST, "Main command"). The sealed control would then contain a false statement about its own provenance.<br>• **Misdirects the confirming round.** The confirming CC2 round and the receipt author follow this pointer to cross-check the self-pin, as `identifiers-delta-cc2-round-3.md` did. Following `:99` gives a stale digest.<br>• **Fails closed, but at a cost.** If that digest reaches the receipt, `review_gate` (`cc2_delta.py:171–176`) refuses. By then `Journal` has already published a start record (`:238`), so the result is a refused canonical action-d journal plus a wrong create-once receipt. Both need reviewed recovery. | 1. Make `:99` and `:107` name the **same** response.<br>2. That response must record the REQUEST's final byte count and SHA-256. Because this fix changes the REQUEST again, the file named should be the response that records the new digest (e.g. `core-impl-round-4-response.md`). Alternatively, reword `:99` to say "the response named in the request row of the table below".<br>3. Change no other REQUEST text and no other bound file.<br>4. Recompute the REQUEST digest and record it in that response. Main re-runs `test_cc2_delta.py`, which reads the REQUEST, and the full suite.<br>5. Hold Core round 5, limited to the REQUEST bytes. It confirms that `preparation.py` is still `9572aed3…`. |

**Severity.** Minor:
- It fails closed.
- It touches no code path.
- The fix is a single sentence.

It is an issue rather than a suggestion because:
- this round's change introduced it into a bound control that will be sealed;
- the revision claimed to have fixed exactly this pointer.

## Suggestions (not required for approval)

- **S1. The counterfactual uses its own assertions** (peer). The counterfactual (`:122–125`, `:137–143`) does not run the regression's own assertions (`:127–128`, `:145`). By contrast, the shared-lock counterfactual wraps them in `assertRaisesRegex` (`:163–166`). Coverage follows from the code instead: B's `JSONDecodeError` is deterministic while A is paused, and `:145` requires B to pass. A wrapped run of the injected ordering, expecting "writer B did not wait for publication", would show this directly.
- **S2. The timing dependence remains** (round-3 S8, challenger). `ready` (`:44–45`) signals that B holds the operation lock, not that B is contending for the publication lock. A B descheduled for more than 0.5 s could let defective ordering pass the regression. The counterfactual has no such dependence. A deterministic version would record B's non-blocking `flock` attempt failing with `EWOULDBLOCK` before releasing A.
- **S3. Adoption can orphan a binding output** (my trace). This is new evidence for round-3 S2. It is binding-lane, pre-existing and not caused by this change.
  - **Sequence:**
    1. Binding loads the config at CLI start-up (`acceptance.py:423`).
    2. `adopt` then completes before binding takes its shared operation lock (`:435`).
    3. Binding reserves its canonical output (`:437`).
    4. The Core re-check correctly refuses with "adopted config changed before journal reservation" (`preparation.py:359–360`).
    5. The `with canonical.open('x')` block exits, leaving an empty canonical binding file. No recovery path exists (round 3, X1C-R3-01 consequence).
  - **Why not a Core issue:** Core behaves as designed. Only a fix to binding's reservation order (round-3 S2 / Alternative C) or a runbook rule removes it. The rule would be: never run `adopt` while a canonical binding run may be starting.
  - **What is needed:** the binding lane or the runbook must address this before the canonical (c) runs.
- **S4. A wrapped `adopt` can hang** (challenger). It is pre-existing and needs an operator to misuse the wrapper.
  - `preparation.py journal -- … preparation.py adopt …` would hang:
    - the wrapper holds a shared operation lock (`:829`) while it waits for its child;
    - `adopt` blocks forever trying to take the operation lock exclusively (`:264`, `blocking=True`).
  - An interrupt recovers it, leaving a refused wrapper journal (round-2 S3).
  - Either document "never run `adopt` under the `journal` wrapper", or later make `adopt` non-blocking.
- **S5. Line numbers in the response are off by one** (peer). The response cites `preparation.py:357–371`; the publication lock is at `:358` and the reload at `:359`.

## Recommendation

Do not approve Core at these bytes, but keep `preparation.py` at `9572aed3…`.
1. Fix X1C-R4-01 in the REQUEST only.
2. Record the new REQUEST digest in a response.
3. Hold Core round 5, limited to the REQUEST bytes.

The CC2 confirming round, the receipt and the single producer run follow only after Core round 5 reports zero issues.

## Alternatives

- **A. Fix the pointer and hold a short Core round 5 (recommended).** One sentence changes. `preparation.py` and the tests are unaffected.
- **B. Defer to the CC2 confirming round.** This would sign Core off with a known false statement in a bound control, which the zero-issue rule does not allow.
- **C. Atomic publication (temp file, then `os.link`).** Not needed for X1C-R3-01. It would also cover the standalone CLI loads (round-3 S1), at the cost of accounting for crash residue.

## Risks

- **Clock time.** Each round uses time on the 45-day clock (deadline `2026-12-24T03:22:58.701909Z`). Round 5 should be short, because no code changes.
- **REQUEST edits.** Every REQUEST edit changes the request's own pin. Both pointers must name the response that records the final digest, or the same mismatch recurs.
- **Untested seal path.** That the seal accepts the producer's records still rests on reading the code (round-3 S6).
- **Binding exposure.** Binding stays exposed to the adoption race (S3) until the binding lane or the runbook addresses it.
- **Reversibility.** Everything up to the first new journal is reversible. After the producer run, nothing bound is.

## Execution outline

1. Edit REQUEST `:99`, and `:107` if needed, so that both name the same response and that response records the final REQUEST bytes and SHA-256. Change nothing else.
2. Main:
   1. recomputes all digests and confirms `preparation.py` is still `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38`;
   2. runs `test_cc2_delta.py` and the full Python suite with the pinned interpreter;
   3. records the output in the response.
3. Hold Core round 5, limited to the REQUEST change.
4. After zero issues:
   1. the CC2 confirming round with all 16 actual digests, including the new REQUEST pin;
   2. the authentic receipt;
   3. the single producer run;
   4. the output review;
   5. the remaining actions and the seal, following round 2's outline.
5. Before canonical binding (c), resolve S3, either in the binding lane or as a runbook rule.

## How the peers shaped this record

- **Peer:**
  - Found X1C-R4-01 (`:99` against `:107` and `identifiers-delta-cc2-round-2-response.md:63`). I confirmed it by grep and accepted it as an issue because this revision introduced it into a bound control.
  - Supplied S1 and S5.
  - Independently confirmed the X1C-R3-01 resolution and that there is no deadlock.
- **Challenger:**
  - Supplied the full lock-order table and the reader inventory, which matched my trace.
  - Found S4. I downgraded it from the challenger's "Minor issue" to a suggestion: it is pre-existing, needs an operator to misuse the wrapper, is recoverable by interrupt, and seals no wrong bytes.
  - Restated S2.
- **Mine:**
  - The line-by-line check against round 3's five required-fix points.
  - Whether the counterfactual reproduces the round-3 ordering.
  - S3, from tracing binding's CLI load against `adopt`.
  - The severity and the scope of the required fix.

## Verdict

X1C-R3-01 is resolved, and the lock change introduces no deadlock and no reader outside the lock. X1C-R4-01 is open. None of these may start using these bytes: a new journal, `adopt`, (b), (c), the CC2 confirming round, the receipt, the producer run or the seal.

Verdict: 1 ISSUES
