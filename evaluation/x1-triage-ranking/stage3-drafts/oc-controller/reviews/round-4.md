# X1 stage-3 draft review — oc-controller — round 4

**Scope.** This round reviews `stage3-drafts/oc-controller/` on PR #21 at head `fcdfda5b566d89fe193b88bedcfe36419c902236`, in worktree `/Users/adamjackson/Projects/work/scout-x1g`. The review was read-only: no edits, no model calls and no real data. **It gives no run authority.** Round 9 §12 still lists the blocking stage-3 prerequisites. C = `controller.py`, T = `test_controller.py`. Paths are relative to the artefact.

## Reviewed files (SHA-256 recomputed by Main)

| File | SHA-256 |
|---|---|
| `README.md` | `8bf16f39f36f4afb6dd0986b8dc7ea8f4eaabe4e2faa46b6a4e3dfc355cb2bd7` |
| `controller.py` | `9fe453651816394a88272a98208f7e484c8bd73e895535638c68c273fc56478c` |
| `benchmark_journal.py` | `36d692e228d5f3afbb8951bfdb12aacc6a07a257a48a8f4a2e2f234a269fdf55` |
| `derive_known_answers.py` | `c789ea6b2b7d4dbef54ed70ae2ec96a02b89fa6822e3d77181fc443c280edc33` |
| `known_answers.json` | `ceae315543f0e775dd1d24657ddd5c225f917b368dfb08655a527d89020634d5` |
| `policy.json` | `ceb2ca53ffee1699ffabe03d10bd3a067d58f2bd92bdc3e20369a4daec88ebba` |
| `mutate_round2.py` | `2b944f4583fefbf654a9853023a22a2f02487f8c2c94b9a833239f50d1772f4c` |
| `test_controller.py` | `30c16339e173f8f925c1b7219e6587a4679b30a0addabed270573a6a516ca542` |
| `reviews/round-3.md` | `d8e708a3e133d7f73ae2584d80e5a874f51ffe6c28cfa26ce8b676e44010e4d6` |
| `reviews/round-3-response.md` | `aded0fa8fe622492cba94bad14528c28b7b81dbc4158231733274a2adb393d52` |
| `reviews/round-3-test-output.txt` | `3fd9a9f8a63538f974564ccaddffa05906107fea860115bb77ade987e5732959` |
| `reviews/round-3-derivation-output.json` | `ceae315543f0e775dd1d24657ddd5c225f917b368dfb08655a527d89020634d5` (byte-identical to `known_answers.json`) |
| `reviews/round-3-mutant-output.txt` | `7232edb21609784fbe8137e8a124c8a1f9cc7740a35cd08c48075f67f199a5d1` |
| `reviews/round-3-journal-output.json` | `24a698b5cc9e6e0642318d19a4395f75868c2e18d7aa783756ae275d5ea223dc` |
| `reviews/round-3-evaluator-preservation.json` | `c6266f87a13019f656ed26f5490a8c739fa0664a79f51588408a965ee9cfd48c` |

Constants pinned in C:22–26: `POLICY_SHA256` `9b6de56992b00129edcb09d8c0b1e95e7f3bfb538dc3fea02c0529e94886bfd5` (re-pinned; matches the OpenSSL output in `known_answers.json`). `ORDER_SHA256` `04aa2184…` is unchanged. `KERNEL_SHA256` `ed670a6b…` and `RANKING_SHA256` `8bc2e7cb…` are unchanged.

## Signed basis (unchanged)

| Item | SHA-256 |
|---|---|
| `PREREGISTRATION.md` | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| `README.md` | `22e2f368db17ed1a5c6ca788e1cba4419fee99863626032853689d5c8c3202c3` |
| `oc-manifest.json` | `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d` |
| `reviews/rev9-digests.json` | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |

## Round-3 disposition

| ID | Fixed | Code and catching test | Residual |
|---|---|---|---|
| R3-01 | **partly** | The poison flag is set before the append (C:360) and stays set on every exception from write, short write, flush, fsync or `_apply` (C:361–372). The deferred path is latched too (C:336–339). `_check_writer` gates `_emit`, `_boundary`, `pause`, `resume`, `run` and `_wait_ready`, and waiters are notified. T:785 covers five stages with a second lane in flight. It checks that no bytes are appended, that six operations are blocked, and that one reopen gives a process-loss pause. The two poison mutants fail an assertion (mutant output lines 195–303). | A short write that leaves out only the record's newline gives a journal that cannot be replayed after the next restart (**R4-01**). |
| R3-02 | **yes** | Records are compact absolute-value deltas keyed by a short population index. Gate changes are held back until the next durable record (C:334–378). A passing attempt writes 2 fsync'd records and a rejected attempt writes 0. There are no meter records. The reservation is computed at C:119–129 and stored in the freeze's `expected` block, so replay compares it (C:217, C:224). Too little free space gives the create-once `freeze-refused.json` and no freeze (C:230–238). Tests: T:720 (fsync counts, lost-gate replay), T:588 (loss between gate and trial start), T:749 (refusal). Benchmark arithmetic checked: 403,760,000 × 281.276 + 100,940 × (448 + 344) = 113,647,942,240 B, and 403,760,000 × 69.76975 = 28,170,234,260 µs. Both match the journal output. | Replay after a stop when only held-back gate changes were pending has no test. This is suggestion 1, not an issue: traced against the current code it behaves correctly. |
| R3-03 | **yes** | Binding leases are refused while the run lock is held (C:465). Completion is re-checked in the same critical section as the first dispatch (C:699–701). T:672 is deterministic because the test holds the condition across `resume` and the lease. The admission mutant fails an assertion (mutant output 305–328). | — |
| R3-04 | **yes** | The construction cell is disclosed in hashed `policy.json` `calibration` and in the README "Construction seed disclosure", with `POLICY_SHA256` re-pinned. `derive_known_answers.py:35–45` derives the base independently with OpenSSL: seed64 `9337516894895418925`, HMAC `81958624…`. T:701 checks that 336 bases are each shared by exactly six κ groups, and that the seed actually passed to preparation equals the vector. | Only population index 0 is pinned (suggestion 6). |
| R3-05 | **yes** | (a) T:614 gate-window gold, cap, pause and pause-plus-resume. (b) T:646 and T:649: `KeyboardInterrupt` from gate and from calibration. (c) T:660 safety-handle guard. All four mutants fail an assertion (mutant output 330–403). (a) is killed through the pause-plus-resume subtest, which isolates the epoch clause. | — |
| R3-06 | **partly** | C:160–161 refuses a non-CPython, non-3.11.11 or optimised interpreter before freezing. C:211–216 records executable path, SHA-256, version, implementation and `python_optimize` in `expected`, and replay refuses a mismatch (C:224–225). In-process, T:749 checks the recorded provenance and that replay refuses a changed version. | The subprocess refusal test passes even with the guard removed, and the implementation refusal has no test (**R4-02**). |

## New issues

| ID | Severity | Location | Finding (input → wrong outcome) | Required fix |
|---|---|---|---|---|
| R4-01 | major (rare trigger, outcome unrecoverable) | C:299, C:305–306, C:251–257, C:313–317; T:808, T:418 | **Recovery classifies a newline-less complete record differently on the first and second restarts.**<br>(1) A journal `write` returns `len(raw)-1`, so the complete JSON object lands without its `\n`. This is a partial write on ENOSPC/EIO, or a host crash that persists every byte but the last. C:363 correctly poisons the writer.<br>(2) First restart: C:299 parses a line only if it ends in `\n`, so the record is treated as torn, its digest is kept, and neither `seq` nor the hash advances. C:252 appends `\n`, then writes the process-loss pause with the record's `seq` N and the old `previous_sha256`.<br>(3) Any later restart: the line now ends in `\n`, parses, chains and is applied, so the expected sequence moves to N+1. The recovery pause, which carries `seq` N, then fails with "journal chain mismatch" (C:314).<br>The run directory can never be reopened, and binding and safety rows become unrecoverable. That is the R3-01 failure mode, against R3-01's required "rebuilt controller replays without a chain error" and the README's "a complete final record is authoritative" [INFERENCE, static trace].<br>T:808 writes half a record, which can never be valid JSON. T:418 reopens twice, but its tail `{"torn":` is also unparseable. Neither catches this. Found by both the lead and the challenger. | **Make the classification stable across restarts.** In `_replay`, a final line without `\n` that parses and chains (`seq` and `previous_sha256` match) is the authoritative complete record. Apply it, compute `_previous` over `line + b"\n"` so later replays agree, and leave `_tail_needs_newline` set so the constructor restores the delimiter before writing the process-loss pause, with `tail_sha256` null. Only a tail that fails to parse or to chain is torn.<br>**Test:** add a poison stage that writes `raw[:-1]`, then reopen, close and reopen again. Require no chain error, and require counters equal to the durable records *including* that record. Add a reversion mutant to `mutate_round2.py` and record that an assertion kills it. |
| R4-02 | minor | T:768–783 (subprocess), C:160–161, C:230–238 | **The `-O` and version refusal test does not isolate the guard.**<br>The subprocess builds `SafetySchedule` without mocking `shutil.disk_usage` and accepts *any* `RuntimeError` with no `freeze.json`.<br>Delete `or sys.flags.optimize` (or the version clause) from C:160. On any volume with less than 1,657,108,561,920 B free, construction then reaches disk admission, raises "insufficient journal free space" and writes no freeze, so both subprocess cases still exit 0 and T:749 passes. Whether the reversion is caught therefore depends on the host's free space [INFERENCE].<br>A regression would then let `python -O` freeze and run with assertions stripped on a large production volume, against P l.169 ("refuse non-zero optimization") and M l.10.<br>The round-3 response says implementation refusal is checked, but no test covers `python_implementation() != "CPython"`. Found by the challenger; the lead had missed it. | In the subprocess, mock ample free space (as `make` does at T:46) and assert the exact "requires CPython 3.11.11 with python_optimize=0" diagnostic. Add a non-CPython case by patching `platform.python_implementation`. Add the optimisation, version and implementation guard reversions to `mutate_round2.py`, and record that an assertion kills each. |

**Counts:** 0 blocking, 1 major (R4-01), 1 minor (R4-02).

### Rejected or downgraded

- **Challenger C3, no test that pause or terminate records carry held-back gate changes:** downgraded to suggestion 1.
  - Traced against the current code, every durable record, including nested `lane_stop`/`phase_stop` and `terminate`, carries `updates` before `_build_report` runs (C:357–358, C:376–379).
  - Deleting the `updates` line outright is caught by T:720, because `restored != durable`.
  - The only surviving mutant needs a new kind-specific branch to be added, so this is hardening rather than an untested existing guard.
- **Peer N4-01, a pause with non-serialisable evidence (NaN, datetime) during an active lane:** this writes a durable `lane_stop` with reason "pause" and then raises `ValueError` before the poison flag is set (C:341–359). Downgraded to suggestion 2.
  - The caller receives the error.
  - No counter or classification changes, and no pause took place.
  - The pattern predates round 3, and normal flow already writes a second `lane_stop` from the run loop's `finally`.

## Non-blocking suggestions

1. **Test replay after a stop that flushed only held-back gate changes** (challenger C3).
   - After termination nothing is re-executed, so the stop record is the only durable carrier of the final gate counts.
   - Test rejected gates followed by gold, a cap, pause/resume and an exhausted lane, with two population indices. Reopen and compare full reports, and add a mutant that omits the control-point flush.
2. **Validate evidence before writing anything.** Call `canonical(evidence)` in `pause`, `resume` and `stop_integrity`, or serialise the outer record before writing the nested `lane_stop`/`phase_stop` (peer N4-01).
3. **Interrupted recovery.** A partial write of the recovery pause after an earlier torn tail ends in "multiple corrupt journal records" (C:303–304). Cover it alongside the R4-01 fix (challenger).
4. **Record the cause of a poisoned halt.** Write a create-once sidecar outside the chain (exception type, time, last `seq`) and cite its digest in the restart pause, to serve P l.171's "with its cause" (peer).
5. **Reservation wording and tightening.** The 2,048 B figure is an aggregate allowance; single `complete` or criticality records can exceed it. Safety populations stop at 2,000 executions, which would save 82,575,360,000 B. Rename the README heading "Open choices for review" to "Author choice for council review": the content already recommends a single choice.
6. **Tests.**
   - Pin construction seeds for indices 1–4.
   - Add disk-refusal and dispatch re-check mutants.
   - A second refusal currently raises `FileExistsError` (`"xb"`) and loses the new evidence, and the directory is not fsync'd after the refusal file is written.
7. **Dead code.** The `kind == "meter"` branch, `"meter"` in the excluded-kinds tuple, and the `data.get("i", …)` fallback in `_apply` are unreachable.
8. **Benchmark caveats** [INFERENCE].
   - It runs one lane with a constant meter and clock, so `updates` from other lanes and real float widths are absent.
   - macOS `os.fsync` does not flush the drive cache. If round-3 suggestion 6 (`F_FULLFSYNC`) is adopted, re-measure first: about 2 × 201.88M flushes at millisecond latency could approach the 15-day cap.
9. Round-3 suggestions 1, 2, 4, 5 and 8 are still open as suggestions. In particular, binding cells never leased before dispatch remain stranded.

## Author choices (R-064)

- **Conservative disk reservation:** 1,657,108,561,920 B = (90,860 + 10,080) × (2 × 4,000 + 16) × 2,048, checked by hand.
  - **Accepted.** It is a fail-closed admission check: it can only refuse, never weaken, a signed safeguard.
  - It is checked once, at freeze. It is not allocated space or a runtime quota, and that is enough for R3-02's required fix.
  - It is recommended, not posed as an open choice. No operator discretion is introduced.
- **Held-back gate changes, re-executed with the same seed after a crash:** accepted as round 3 recommended.
- **The safety-profile construction cell:** accepted with its disclosure, as round 3 decided.

## Recommendation, risks, execution outline

**Recommendation:** keep the round-3 architecture unchanged. Fix R4-01 in `_replay` by treating a newline-less record that parses and chains as authoritative, with its digest normalised, then fix R4-02's test isolation, and resubmit for a scoped round 5.

**Alternatives:**
- Truncate the torn tail on recovery: rejected, because it destroys evidence.
- Keep the torn bytes in an immutable segment file: viable, but more moving parts.
- Write the recovery pause with a look-ahead match on `tail_sha256`: viable, but it conflicts with the stated "complete record is authoritative" rule.

**Risks:**
- R4-01's trigger is rare but the outcome cannot be undone, short of hand-editing an evidence journal.
- R4-02 is a test gap; the guard in C:160 is currently correct.
- Everything is a reversible draft.

**Execution outline:**
1. Change `_replay` and the constructor recovery as R4-01 requires, and add the `raw[:-1]` stage and the double-reopen assertion to T:785.
2. Optionally cover interrupted recovery (suggestion 3).
3. Isolate the subprocess refusal tests, and add the implementation case and three guard mutants.
4. Re-run the derivation, the suite and the mutants, commit the full outputs, and recompute the digests.

## Peer contributions

**Challenger (GPT Astra):**
- Independently found R4-01 and rated it major; the lead had reached the same trace and agrees.
- Found R4-02. The lead had not checked the subprocess test against disk admission, and this changed the issue count.
- C3 was downgraded to a suggestion.

**Experiment peer (Opus):**
- Independently confirmed all arithmetic and the reasons each mutant was killed.
- Rated R3-01 and R3-06 fixed; R4-01 and R4-02 overrule that.
- N4-01 was downgraded to a suggestion.
- Supplied suggestions 4–8.

**Lead:**
- Found R4-01, including the digest-normalisation requirement in the fix.
- Traced the correctness of the held-back gate changes, the index mapping and the terminate flush.
- Decided the downgrades of C3 and N4-01.

## Verification

- No shell was used in this council session or either peer session. Every behaviour and mutant-survival claim here is a static trace [INFERENCE].
- **Main's independent run** at `fcdfda5b`: clean tree, pinned `.tools/venv/bin/python` 3.11.11, `PYTHONDONTWRITEBYTECODE=1`.
  - `oc-controller/test_controller.py`: 35 tests in 3.400s, OK.
  - `locked-size-input/test_size_input.py`: 30 tests, OK.
  - `locked-size-input/test_cli.py`: 1 test, OK.
  - File digests are Main's.
- The mutant results (16/16), the derivation output and the benchmark figures are the author's records. The council checked the arithmetic but did not re-run them.
- **Files read:**
  - all of `controller.py`, `README.md`, `policy.json`, `derive_known_answers.py`, `mutate_round2.py` and `benchmark_journal.py`;
  - T:393–475 and T:580–866;
  - `round-3.md` and `round-3-response.md`;
  - the mutant output summary and the `round-3-journal-output.json` figures.
- **Main should run before round 5:**
  - `shasum -a 256` on the revised files;
  - the README test command;
  - `derive_known_answers.py`, compared byte-for-byte with `known_answers.json`;
  - `mutate_round2.py` with the R4-01 and R4-02 mutants.

## Verdict

**2 issues open** (0 blocking, 1 major R4-01, 1 minor R4-02). R3-01 and R3-06 are only partly fixed; R3-02 to R3-05 are fixed with tests that catch them. The draft is not accepted for its stage-3 role. A revised draft needs a new scoped round with zero issues, and acceptance would still give no run authority.
