# X1 stage-3 draft review — oc-controller — round 5

**Scope.** This round reviews `stage3-drafts/oc-controller/` on PR #21 at head `66bfb1408d24f3587ff4afc02bc6afbb1b54fba9`, in worktree `/Users/adamjackson/Projects/work/scout-x1g`. The review was read-only: no edits, no shell, no model calls and no real data. It covers only the round-4 fixes and the code they affect. **It gives no run authority**; round 9 §12 still lists the blocking stage-3 prerequisites. C = `controller.py`, T = `test_controller.py`, M = `mutate_round2.py`. Paths are relative to the artefact.

## Reviewed files (SHA-256 recomputed by Main)

| File | SHA-256 | Status |
|---|---|---|
| `README.md` | `e6727b1fa2787ad3fcad3dcd50e2b8c816f5527a39e3421f4df07ccf2e802093` | changed |
| `controller.py` | `4627b3bb419dcf0fe5de230609e9e768570b8a616a485015b7c766c02c90f8a0` | changed |
| `mutate_round2.py` | `f7df46e237ed69b5d6598a0828ea18ed770923560694060892957794b955cb10` | changed |
| `test_controller.py` | `16cc62ea3d8b497ee301e300a41103ecd197cf92129157abfad248e66ba62726` | changed |
| `benchmark_journal.py` | `36d692e228d5f3afbb8951bfdb12aacc6a07a257a48a8f4a2e2f234a269fdf55` | unchanged |
| `derive_known_answers.py` | `c789ea6b2b7d4dbef54ed70ae2ec96a02b89fa6822e3d77181fc443c280edc33` | unchanged |
| `known_answers.json` | `ceae315543f0e775dd1d24657ddd5c225f917b368dfb08655a527d89020634d5` | unchanged |
| `policy.json` | `ceb2ca53ffee1699ffabe03d10bd3a067d58f2bd92bdc3e20369a4daec88ebba` | unchanged |
| `reviews/round-4.md` | `9d9a0d52937a60f7fcdeb1945d5c242cca4abde0c80f2e72ab7e27bae71d466f` | record |
| `reviews/round-4-response.md` | `bcc2aa8913fdc1cc8eb301f9cbdbfe7c05e08291374b50063af8eef168c16925` | record |
| `reviews/round-4-test-output.txt` | `a2070f3f5ae1e0e8ccc54c87501a2abdc667f6bd05a9a5fdbc184f95211dd0b2` | evidence |
| `reviews/round-4-mutant-output.txt` | `702ad39561cfc84eef0f2f8f9d150b9f59e9b0c270e82ed632d132e206846744` | evidence |
| `reviews/round-4-derivation-output.json` | `ceae315543f0e775dd1d24657ddd5c225f917b368dfb08655a527d89020634d5` (byte-identical to `known_answers.json`) | evidence |
| `reviews/round-4-evaluator-preservation.json` | `c6266f87a13019f656ed26f5490a8c739fa0664a79f51588408a965ee9cfd48c` | evidence |

The four unchanged files are byte-identical to round 4 and were not re-audited. The pinned constants at C:22–26 are unchanged.

## Signed basis (unchanged)

| Item | SHA-256 |
|---|---|
| `PREREGISTRATION.md` | `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` |
| `README.md` | `22e2f368db17ed1a5c6ca788e1cba4419fee99863626032853689d5c8c3202c3` |
| `oc-manifest.json` | `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d` |
| `reviews/rev9-digests.json` | `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da` |
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |

## Round-4 disposition

### R4-01 — fixed: yes

**Problem.** A complete final record that lost only its newline was read one way on the first restart and another way on the second.

**Code.** In `_replay` (C:292–321):
- A final line without `\n` counts as torn only if it fails to parse, or parses but does not chain on `seq` and `previous_sha256` (C:303–304).
- Otherwise it goes through the normal chain check (C:315) and is applied (C:317–318).
- `_tail_needs_newline` stays set (C:319), and `_previous` is hashed over `raw + b"\n"` (C:320).
- The constructor writes and fsyncs the missing `\n` before the process-loss pause (C:251–255). Because `_tail_sha` is None, that pause has `tail_sha256` null (C:256–257).

**Why it is correct (traced; peers agree):**
- **Digests match.** The writer hashes `canonical(event)+b"\n"` (C:361, C:368). One missing byte leaves `raw[:-1]` on disk, and replay hashes `raw[:-1]+b"\n"`, which is the same bytes. Any shorter prefix of a compact JSON object leaves the outer `}` unclosed, so it cannot parse. The only newline-less tail that parses is therefore exactly `raw[:-1]`, and it always chains, because appends are never retried after poisoning.
- **The second restart agrees with the first.** On the second restart the same line ends in `\n`, so `_previous = digest(raw)` gives the same value. The recovery pause (N+1) chains, and the counters match.
- **Failures during recovery stay stable:**
  - If writing the `\n` raises, the journal is unchanged and the next restart repeats the same recovery.
  - If the `\n` is written but the pause is torn, the next restart takes the existing single-torn-tail path.
  - If the pause itself loses only its newline, it parses, chains and is treated as authoritative.
- **A newline-less `terminate` record** is applied. C:250 then skips both the newline restoration and the pause. No code path appends after termination: `_boundary` C:437, `pause` C:587, `resume` C:597, `run` C:689, `_terminate` C:773 and `stop_integrity` C:785 all check for it. Every later replay therefore reads the line the same way.

**Catching test.** The `newline` stage at T:815 writes `raw[:-1]`, and `_emit` poisons the writer with "short journal write".
- T:851–859 rebuild population 0 independently from the journal bytes, using `splitlines`, `json.loads` and dict updates, so the newline-less record is included.
- T:860–870 restart twice. Each time they require no constructor error, a process-loss pause, a null `tail_sha256` and population 0 equal to the rebuilt state.
- The lost record is the trial-start transition (C:553–560). It changes `executions`, `pending` and the carried gate deltas, so the state assertion can tell the two outcomes apart.

**Mutant.** M:28 restores newline-dependent parsing. It is killed by an `AssertionError` at T:868 in the `newline` subtest only (mutant output lines 423–440).

By trace [INFERENCE], two other reversions would also be caught at the second restart, though neither has its own mutant:
- dropping the `+ b"\n"` normalisation at C:320 gives a chain mismatch and fails T:864;
- dropping the flag at C:319 merges the pause onto the record's line and fails T:868–869.

### R4-02 — fixed: yes

**Problem.** The interpreter refusal test did not isolate the guard.

**Test.** The subprocess cases (T:768–788) set up each condition before construction (T:780) and before the guard at C:160–161:
- ample free space is mocked (T:776);
- the `-O` case uses the real flag, the version case sets `version_info` (T:777), and the implementation case patches it to "PyPy" (T:778).

Each case makes exactly one guard clause true. The child exits 0 only if it sees the exact diagnostic and no `freeze.json`, and the parent asserts both the return code (T:787) and the exact stdout (T:788). Host free space is no longer a factor.

**Mutants.** M:29–31 delete each guard clause separately. Each is killed by `AssertionError: 3 != 0` at T:787 (mutant output lines 442–494). Exit 3 is the path where construction succeeds, which is the evidence that the guard was bypassed rather than some unrelated error.

The harness only counts a kill when the output contains FAIL and AssertionError and no ERROR (M:54). The recorded summary is 20/20 (mutant output line 496).

### Response and README claims

- The response's line citations (C:292–320, M:28, M:29–31) and the "20/20 assertion-killed" claim match the committed files.
- README l.13 and l.62 describe the actual behaviour. l.13's "every durable population field" overstates slightly: the test checks population 0 only. This is wording (suggestion 4).

## New issues

None.

## Rejected or downgraded

- **A newline-less tail that parses but does not chain is unstable across restarts** (C:303–304, C:315–316). Raised by the peer and the challenger, and found independently by the lead. The first restart treats it as torn and restores the `\n`; the second restart sees a newline-terminated line that does not chain and raises "journal chain mismatch".
  - No path from the writer produces such a tail: no truncation of a record parses unless it is complete, and appends are never retried.
  - The same behaviour existed before the fix, and both outcomes refuse to continue while keeping the evidence.
  - Downgraded to suggestion 1.
- **A newline-less tail that parses to a non-object, or to an object without `seq`, raises `TypeError`/`KeyError`** at C:304 instead of being treated as torn (peer). Such a tail can only be filesystem garbage, not part of a record, and the controller refuses to continue. This is the accepted fail-closed default; see suggestion 2.

## Non-blocking suggestions

1. Make the parseable-but-non-chaining branch stable. Either refuse it outright on the first restart, or treat a newline-terminated line as torn when the next line is a pause whose `tail_sha256` matches its digest.
2. Check `isinstance(event, dict)` and use `.get("seq")` before the chain comparison at C:304.
3. Tests:
   - Add a mutant for the `+ b"\n"` normalisation at C:320, so the second-restart path has its own evidence.
   - Add a newline-less `terminate` stage to cover the skip at C:250.
   - Add an interrupted-recovery matrix (challenger): the recovery pause fails before writing, mid-record, and just before its newline; reopen twice each time.
4. Assert population 1 in the poison-test rebuild, or narrow the wording of README l.13.
5. Round-4 suggestions 1–9 remain open as suggestions, unchanged.

## Recommendation, risks and next steps

**Recommendation:** accept the draft for its stage-3 role. Keep the minimal `_replay` fix as it is.

**Alternatives considered:**
- Require suggestion 1 first: it would harden a branch that only filesystem corruption can reach, so it is not required.
- Truncate the torn tail during recovery: rejected again, because it destroys evidence.

**Risks:**
- Corruption that is not a prefix of a record, or two corrupted tails in a row, still makes the run unrecoverable. It fails closed and keeps the evidence; this is the accepted round-4 suggestion 3.
- Everything is a reversible draft.

**Next steps:**
1. Main records this round as `reviews/round-5.md`.
2. Any optional suggestion that is adopted needs a new scoped round.
3. Before any later stage, round 9 §12 prerequisites still apply.

## Peer contributions

- **Challenger (GPT Astra):** independently confirmed both fixes and the mutant kills. Traced the torn-recovery-pause and lock/reopen cycles. Proposed the interrupted-recovery matrix (suggestion 3). Treated the non-chaining tail as out of scope. None of this changed the recommendation.
- **Experiment peer (Opus):** confirmed the two fixes, the independence of the rebuild and the absence of confounds in R4-02. Raised the non-chaining instability and the non-object `TypeError`, both downgraded (suggestions 1–2). Proposed the normalisation mutant, the terminate stage and the README wording fix (suggestions 3–4). Agreed that there are zero issues.
- **Lead:** independently traced the non-chaining instability and decided its downgrade. Checked every code path that runs after termination, and checked how each R4-02 case isolates its guard clause.

## Verification

- **This session and both peer sessions used no shell.** Every behaviour claim is a static trace [INFERENCE].
- **Main's independent run** at `66bfb140`: clean tree, pinned `.tools/venv/bin/python` 3.11.11, `PYTHONDONTWRITEBYTECODE=1`.
  - `oc-controller/test_controller.py`: 35 tests in 1.366s, OK.
  - `locked-size-input/test_size_input.py`: 32 tests in 593.805s, OK.
  - `locked-size-input/test_cli.py`: 1 test in 602.449s, OK.
  - All file digests are Main's.
- **Author's records, not re-run here:** the mutant run (20/20 assertion-killed), the derivation output and the evaluator-preservation record.
- **Files read:**
  - C:140–400 and C:400–563, plus every `_terminated` and emit guard in C;
  - T:740–876 and M in full;
  - README lines touching replay and the interpreter guard;
  - the round-4 record and response;
  - the summary of the mutant output.

## Verdict

**zero issues — draft accepted for its stage-3 role, conferring no run authority.**

Sign-off: the council reviewed the following against the unchanged signed basis (`PREREGISTRATION.md` `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb`, `README.md` `22e2f368db17ed1a5c6ca788e1cba4419fee99863626032853689d5c8c3202c3`, `oc-manifest.json` `1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d`, `reviews/rev9-digests.json` `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`, `reviews/round-9.md` `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc`) and reports zero open issues:

- `oc-controller/README.md` `e6727b1fa2787ad3fcad3dcd50e2b8c816f5527a39e3421f4df07ccf2e802093`
- `controller.py` `4627b3bb419dcf0fe5de230609e9e768570b8a616a485015b7c766c02c90f8a0`
- `test_controller.py` `16cc62ea3d8b497ee301e300a41103ecd197cf92129157abfad248e66ba62726`
- `mutate_round2.py` `f7df46e237ed69b5d6598a0828ea18ed770923560694060892957794b955cb10`
- `benchmark_journal.py` `36d692e228d5f3afbb8951bfdb12aacc6a07a257a48a8f4a2e2f234a269fdf55`
- `derive_known_answers.py` `c789ea6b2b7d4dbef54ed70ae2ec96a02b89fa6822e3d77181fc443c280edc33`
- `known_answers.json` `ceae315543f0e775dd1d24657ddd5c225f917b368dfb08655a527d89020634d5`
- `policy.json` `ceb2ca53ffee1699ffabe03d10bd3a067d58f2bd92bdc3e20369a4daec88ebba`
- `reviews/round-4.md` `9d9a0d52937a60f7fcdeb1945d5c242cca4abde0c80f2e72ab7e27bae71d466f`
- `reviews/round-4-response.md` `bcc2aa8913fdc1cc8eb301f9cbdbfe7c05e08291374b50063af8eef168c16925`
- `reviews/round-4-test-output.txt` `a2070f3f5ae1e0e8ccc54c87501a2abdc667f6bd05a9a5fdbc184f95211dd0b2`
- `reviews/round-4-mutant-output.txt` `702ad39561cfc84eef0f2f8f9d150b9f59e9b0c270e82ed632d132e206846744`
- `reviews/round-4-derivation-output.json` `ceae315543f0e775dd1d24657ddd5c225f917b368dfb08655a527d89020634d5`
- `reviews/round-4-evaluator-preservation.json` `c6266f87a13019f656ed26f5490a8c739fa0664a79f51588408a965ee9cfd48c`

Acceptance gives no preparation or run authority. Any change to these files needs a new scoped round with zero issues.
