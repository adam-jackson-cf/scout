# X1 preparation core implementation review — round 3

**Verdict: 1 issue open (X1C-R3-01, Moderate).** Core is not approved at `preparation.py` `7f8d390fd3ebb50140b5725d7704ccf75b5d157fd74fa73810864a625962753e`.
- X1C-R2-01 is fixed as round 2 specified it.
- The Core part of CC2 R2-01 is implemented correctly.
- **But** a second reader of live journal start records is still outside the new publication lock: the configuration reload inside `Journal.__init__`. Through the binding run's canonical reservation, it can make the seal permanently impossible.

**Executed evidence remains valid:** the clock, the input receipt, the five object receipts, the eight legacy journal pairs and the baseline (d) journal `906389d7…`.

**No new journal has been written.** `preparation-run/journal/` still holds exactly the eight legacy pairs, and OUT has no `identifiers-delta-cc2/`, `reviews/` or `config-adoptions/`.

**Scope and method.** This review was read-only. Nothing was run: no code, tests, hashes, Git commands, adoption, preflight, producer or seal. No models were called beyond the two read-only peer analyses. Three analyses fed this record:
- **My own trace** of `preparation.py`, `verify_inputs.py`, the three test files, the CC2 producer `identifiers/cc2_delta.py`, `binding/acceptance.py`, the REQUEST and the run-directory listing.
- **`innovation-challenger` (CoreR3Challenger):** an inventory of every reader of `journal/*` records and what each caller has already reserved when that read fails; lock ordering; a search for CC2 bypasses.
- **`experiment-peer` (CoreR3Peer):** for each prior issue, the observation that would show it is not fixed; whether each new test fails for its stated reason; whether the seal accepts the producer's records under the exact REQUEST command.

This record does not contain the exact round-9 verdict line or a whole-line scope marker, so `review_receipt` cannot accept it as authority.

## Reviewed digests

The digests are Main's (`shasum -a 256`), from the brief and Main's later digest update. I did not recompute them. Paths are relative to `evaluation/x1-triage-ranking/`.

| File | SHA-256 | Role |
|---|---|---|
| `execution/preparation.py` | `7f8d390fd3ebb50140b5725d7704ccf75b5d157fd74fa73810864a625962753e` | Under review (changed) |
| `execution/tests/test_journal_publication.py` | `6d8a0f1a64370dc0775e3876ff539ca60510323d16ebdede87e3b07aa6a3f956` | Under review (new) |
| `execution/tests/test_cc2_seal.py` | `91e33dcac2d9ab6fcf0a9160cc01755b0fbb051f9c9f56e4e6504136f6190f82` | Under review (new) |
| `execution/tests/test_binding_seal.py` | `7c05eaf2f577210361bbd18a93087e7f59b8fffa687c09d98fad2824d8f4b593` | Under review (changed) |
| `preparation-reviews/identifiers-delta-cc2-REQUEST.md` | `c32bd40006bacdf77eb8970d8d4aeaee5be28a6c9bdc243910983c8dba233fb4` | Under review. From Main's update; it supersedes `402f0a09…` in the brief and in `core-impl-round-2-response.md` |
| `execution/tests/test_cc2_delta.py` | `4844221de3ecb6c52b81a770ad7dadbcc42d8e0122636a29de2cdb1fc0e9d8d7` | Context (changed by the Identifiers owner). Full digest from the REQUEST's 16-row table; Main gave the prefix `4844221d…` |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` | Unchanged; executed |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | Unchanged; pinned by the clock |
| `preparation-reviews/core-impl-round-2-response.md` | `9a82c2c91cb08927cf589922084d8aede6039fc5a30ed45da081569eb2003afb` | Author response |
| `preparation-reviews/core-impl-round-2.md` | `e6b7e726…` | Prior round |

**Main's test results at these bytes** (recorded output: `core-impl-round-2-response.md:78–124`). I did not re-run any of them.
- `test_journal_publication.py`: 3 tests OK, including the counterfactual with the lock disabled.
- Full Python suite: 287 OK. Main reports the same count after the REQUEST and `test_cc2_delta.py` update.
- `go vet` and `go test -count=1` on `execution/model_free` (Go 1.26.10): clean.

## Disposition of prior issues

| ID | Status | Evidence |
|---|---|---|
| X1C-R2-01 (Moderate) | **Resolved as specified** | **Code** (`preparation.py`):<br>• `journal_publication` (`:340–348`) takes `flock(LOCK_EX)` on a read-only `O_DIRECTORY` descriptor of `journal/`. It creates no file and releases the lock on close, including on exception.<br>• `Journal.__init__` takes the shared operation lock first (`:355`), then holds the publication lock around the predecessor glob and pins and the exclusive start write (`:369–371`).<br>• `finish` takes the operation lock (`:374`), then the publication lock around the terminal write (`:379–380`).<br>• The operation lock is never upgraded.<br>• `exclusive_json` writes, flushes, fsyncs and closes before it returns (`verify_inputs.py:34–46`), so the lock covers the whole write.<br>• The order is always operation lock, then publication lock, so there is no lock-order cycle. `adopt` and `seal` hold the operation lock exclusively, which excludes all publishers.<br>**Tests** (`test_journal_publication.py`):<br>• Writer A pauses one byte into its start or terminal record. Writer B signals that it holds the shared lock.<br>• The test asserts that B does not pin A's target within 0.5 s. It then checks that B's stored predecessor pin equals A's final bytes, and that OUT holds only `.operation.lock`, `clock.json` and `journal/`.<br>• The counterfactual replaces only `journal_publication` with `nullcontext` and reproduces the exact assertion in both phases. So both regressions fail for their stated reason.<br>• An early B failure cannot pass silently: the exit-code and `('b','passed')` checks would fail.<br>**Limit:** the fixture config is a plain `dict` (`test_journal_publication.py:20`), so the `RunConfig` reload at `preparation.py:357` never runs. See X1C-R3-01. |
| CC2 R2-01, Core part | **Resolved** | **Code:** `collect_cc2_delta` (`preparation.py:532–593`) is called unconditionally from `collect_seal` (`:732–735`), after the original required-producer check. It requires:<br>• the three outputs, each with its canonical schema, passed status, the canonical clock, the implementation-receipt pin, the signed source-manifest pin and the exact baseline predecessors (`:554–572`);<br>• no failures, no unmeasured entries and no freshness filtering;<br>• the conservative union flags and a matching familiarity source table (`:573–576`);<br>• all 16 controls, as path/hash pairs, in the `identifier-delta-cc2-implementation` receipt (`:539–553`);<br>• exactly one passed action-d producer, whose `output_pins` equal the three outputs and whose `source_pins` contain the code and source pins (`:577–582`);<br>• a separate `identifier-delta-cc2-output` receipt that binds all three outputs and that producer's terminal record (`:583–588`).<br>Both council records enter `scoped_reviews`. The delta files are under OUT, so the existing freeze loop covers them.<br>**Producer compatibility (traced, not tested):**<br>• The REQUEST command's relative arguments resolve to the configured absolute paths. `cc2_delta.py` refuses non-canonical review, sources and output paths; `read_clock` enforces the clock path; `load_config` enforces the original config path.<br>• The output order (manifest, coverage, familiarity) and the pin keys `{path, sha256, bytes}` match.<br>• The producer's `control_pins` equal the seal's `controls`.<br>• Identifier records carry no `path`+`sha256` keys, so the pin walker does not misread them.<br>• The baseline journal `906389d7…` was produced the same way, and it records absolute pins in config-path form.<br>**Tests** (`test_cc2_seal.py`): each of these fails with the intended message: dropping any of the 16 controls or any of the four output-review rows; mutating each class of manifest field; removing, duplicating or relabelling the producer.<br>**Request** (`identifiers-delta-cc2-REQUEST.md`, `c32bd400…`) now contains:<br>• the mandatory freeze order;<br>• the exact verdict and scope-block rule;<br>• the integration paragraph naming `7f8d390f…` and `7d6f89ce…`;<br>• the 16-row table, with `test_cc2_delta.py` at `4844221d…`.<br>The fix for X1C-R3-01 changes `preparation.py`, so that paragraph, the 16-row table and the confirming CC2 round must all be refreshed. |
| Round-2 S1–S9 | Still suggestions | Not reopened. S2 (interrupted operations) and S3 (failed canonical runs) remain runbook matters. |

## New issue

| ID | Severity | Location | Concrete failure | Required fix |
|---|---|---|---|---|
| X1C-R3-01 | Moderate | **Race:** `execution/preparation.py:357` (the reload inside `Journal.__init__`), which calls `load_config`; `:149–150` there parses every `journal/*.start.json`. The publication lock is taken only at `:369`.<br>**Consequence:** through `execution/binding/acceptance.py:435–439`.<br>**Test gap:** `execution/tests/test_journal_publication.py:20`. | **A live start record is still parsed outside the publication lock.** `Journal.__init__` holds only the shared operation lock when it calls `load_config(config_path)` (`:357`). It does this whenever the config is a `RunConfig`, which every production caller passes. `load_config` globs all start records and runs `json.loads` on each one (`:149–150`).<br>**Sequence:**<br>1. Operation A is writing its start record: it holds the shared and publication locks and is between `open('x')` and its final write. A could be the CC2 producer, a closure run or preflight, or a `journal`-wrapped helper. Start records are about 8.5–10 kB, larger than the text buffer, so the file is still empty or partial.<br>2. Canonical binding run B holds the shared operation lock. It has already created and fsynced an empty `binding/synthetic.json` or `binding/django.json` (`acceptance.py:437–438`).<br>3. B constructs its `Journal` (`acceptance.py:439`). The reload reads A's partial start record and raises `JSONDecodeError`.<br>4. The exception leaves the `with canonical.open('x')` block before `complete_operation` runs. B has no start record, no terminal record and no refused receipt. The canonical file stays empty.<br>5. A rerun fails at `canonical.open('x')`. At seal, `json.loads` of the required binding record fails. `operation-recovery` applies only to journals, so nothing can recover it. The seal is permanently impossible.<br>**Why it is plausible:** it has the same window and the same permitted concurrency that made X1C-R2-01 an issue. Binding (c), the CC2 producer and closure runs belong to separate lanes and may overlap.<br>**Scope:** for every other caller (fetch, closure, CC2, identifiers, the `journal` wrapper and CLI start-up loads), the same race causes only a retryable refusal before anything is reserved. So the issue is confined to binding, but there it is irreversible.<br>**Why the tests miss it:** the fixture's config is a plain `dict`, so the `isinstance(config, RunConfig)` branch at `:357` never runs. | 1. In `Journal.__init__`, acquire `journal_publication(self.directory)` before the configuration re-check. Hold it continuously through the re-check, the predecessor snapshot and the start publication. Keep the order operation lock, then publication lock, and do not upgrade the operation lock.<br>2. Do not take the publication lock inside `load_config` itself. A second `flock` on a new descriptor in the same process would block against the first and deadlock. Standalone CLI loads stay unprotected (S1).<br>3. Do not just catch the parse error after the binding reservation.<br>4. Add a start-phase overlap regression in which B uses a real `RunConfig`, so that `load_config` actually runs. This needs a minimal valid run config and a clock whose `execution_config` pin names that config path. A pauses mid-start. Assert that B waits, passes and pins A's final bytes. Add a counterfactual that runs the reload outside the lock and reproduces `JSONDecodeError`. A variant that drives the binding reservation into `Journal` and asserts no empty canonical file is preferred, but not required.<br>5. Then recompute the digests, refresh the REQUEST's integration paragraph and 16-row table, and hold a Core round 4 limited to this change. No journal of any kind may start before that round reports zero issues. |

**Severity.** All three analyses found this independently, and all rate it Moderate.
- It has the same class and consequence as X1C-R2-01: a race the code permits, an irreversible result, and no recovery path.
- It is new evidence, not a reopening. Round 2's required fix named only the predecessor snapshot and the two publications, and those are fixed.

## Suggestions (not required for approval)

- **S1. Standalone `load_config` scans.** These loads can refuse spuriously while another operation is publishing its start record:
  - CLI start-up loads (`preparation.py` `main`, binding, closure, fetch, CC2, identifiers);
  - the initial load of `seal --check`.
  
  Nothing is reserved at that point, so a retry recovers. Either:
  - document "retry on a JSON decode refusal" in the runbook; or
  - later, make record publication atomic (temp file, fsync, `os.link` to the final name, unlink), with explicit accounting for temp files left by a crash.
- **S2. Binding reservation order (binding lane).** Constructing the `Journal` before `canonical.open('x')` in `binding/acceptance.py` would also remove the irreversible consequence. It needs its own binding review, so it is not part of the Core fix.
- **S3. The missing-output test refuses earlier than intended.** In `test_full_seal_refuses_missing_delta_output` (`test_cc2_seal.py:43–48`), `FileNotFoundError('coverage.json')` comes first from `collect_journals`' `verify_tree_pins` on the producer journal, not from `collect_cc2_delta`. Add a seal-level case where CC2 never ran: no producer, no outputs and no output review.
- **S4. The changed-source test is generic.** In `test_full_seal_refuses_changed_execution_source`, the existing execution-source manifest check refuses before the CC2 code runs. It is a valid test, but it is not evidence for the CC2 controls. That evidence is the 16-control omission test.
- **S5. The baseline check is never reached in tests.** The explicit baseline-mismatch refusal (`preparation.py:546–547`) always comes after `review_receipt`'s `validate_pin` on the receipt's baseline rows. It is defence in depth, and the substitution test's message correctly names the earlier check.
- **S6. The real producer does not run in the seal tests.** `test_binding_seal.py:132–147` builds the producer's `common` fields, `source_pins` and receipt rows by hand, and publishes zero identifiers. Add:
  - one test that runs `cc2_delta.run`, or `publish` with a non-empty record set and the real `common`, into `collect_cc2_delta`;
  - one test that runs a real `seal` in a temporary directory and asserts the delta files are frozen read-only.
- **S7. The seal now depends hard on CC2.** The seal requires a passed CC2 run, so each of these now blocks the seal, not just inventory:
  - a failure after `root.mkdir()` (`cc2_delta.py` `publish`);
  - an I/O error;
  - an interrupt (the producer catches only `Exception`).
  
  Check free disk space before the run, and record this beside round-2 S2 and S3 in the runbook.
- **S8. The regressions depend on timing.** The publication tests rely on a 0.5 s window, so on a loaded host they could pass on defective code. The counterfactual shows they catch the race on this host.
- **S9. Spelling of the seal's config path.** `collect_cc2_delta` compares `pin(config_path)` by path string. Run the seal with the same absolute or repo-relative spelling as the producer. Otherwise it refuses, though a corrected rerun recovers.

## Recommendation

Do not approve Core at these bytes.
1. Apply the X1C-R3-01 fix in `preparation.py`, with its `RunConfig` overlap regression and counterfactual.
2. In the same revision, refresh the REQUEST's helper digests and 16-row table.
3. Hold a short Core round 4 limited to that change.

The CC2 confirming round, the receipt and the single producer run follow only after Core round 4 reports zero issues.

## Alternatives

- **A. Extend the publication lock over the reload in `Journal.__init__` (recommended).** A small change to the lock's scope in Core. It creates no new artefact and keeps operation bodies parallel.
- **B. Atomic publication (temp file, then `os.link`).** Protects every unlocked reader, including CLI loads. But a crash leaves temp files under OUT that need accounting.
- **C. Reserve the binding output after constructing the `Journal`.** Removes the irreversible consequence, but changes signed binding code and leaves the Core race in place.
- **D. Forbid concurrency in the runbook.** Rejected, as in round 2: the code permits it, and the consequence is permanent.

## Risks

- Every further round uses time on the 45-day clock (deadline `2026-12-24T03:22:58.701909Z`).
- The fix changes `preparation.py`, which forces changes to the CC2 REQUEST, its 16-row packet and the confirming CC2 round. From the first new journal onwards, Core is frozen (round-2 S5).
- That the seal accepts the producer's records rests on reading the code, not on a test (S6).

## Execution outline

1. Implement the X1C-R3-01 fix and the `RunConfig` overlap regression with its counterfactual. Change no create-once artefact.
2. Main runs the publication tests, the full Python suite, `go vet` and `go test -count=1`, and recomputes the digests.
3. Update the REQUEST's integration paragraph and 16-row table to the new `preparation.py` digest.
4. Hold Core round 4 at the new digests.
5. Then:
   1. the confirming CC2 round;
   2. the receipt;
   3. the single producer run;
   4. the output review;
   5. the remaining actions and the seal, following round 2's outline.

## How the peers shaped this record

- **Challenger:**
  - Found X1C-R3-01 independently.
  - Supplied the complete inventory of journal readers, which confirmed that binding is the only caller that reserves something irreversible before constructing its `Journal`.
  - Proposed atomic publication (Alternative B) and S6.
- **Peer:**
  - Found X1C-R3-01 independently.
  - Identified the deadlock constraint on locking inside `load_config`.
  - Traced whether the seal accepts the producer's records under the exact REQUEST command.
  - Supplied S3–S5, S7 and S8.
- **Mine:**
  - The initial trace from the plain-`dict` fixture to the unexercised `RunConfig` branch and the binding reservation.
  - The severity decision, S9, and the check that no new journal exists.

## Verdict

X1C-R2-01 and the Core part of CC2 R2-01 are resolved. X1C-R3-01 is open. None of these may start using these bytes: a new journal, `adopt`, (b), (c), the CC2 producer run or the seal.

Verdict: 1 ISSUES
