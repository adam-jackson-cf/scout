# X1 preparation runtime closure (b): weak-link implementation review — round 2

## Scope and method

**Scope.** This round checks round 2 of the remedy A′ implementation against the required fixes in `preparation-reviews/closure-weaklink-impl-round-1.md` (SHA-256 `b50e016181651e8b98c33973c3a2869c563403025f91cd2785707c0940c7bc99`). The fixes are X1CW-R1-01 to 03, R3′a–c and the round-2 evidence requirements. The writer's response is `preparation-reviews/closure-weaklink-impl-round-2-response.md` (`e817cf22222798eeef4782fb5a5912dd95ffb6ce892b840d57e3fe031a9bad49`).

This round signs no code off. It does not authorise O0, GR, a journalled preflight or a canonical closure. Under R-064, Main adopts these recommendations without asking the owner.

**Method.**
- The council was read-only. The lead ran no command, test, hash or journal.
- Two independent peer analyses ran: `innovation-challenger` and `experiment-peer`.
- All digests are Main's `shasum -a 256` or the pinned suite record. The peers had no hashing tool: their digest checks compare records with one another, not fresh hashes.
- `SCRATCH` = `/Users/adamjackson/Projects/work/x1-preparation-scratch`.

## Reviewed bytes

| File | SHA-256 |
|---|---|
| `execution/closure/native_v2.py` | `5e320ed37a5b106e3a37984e6137e78f87422aa525792018032117d66f4dee09` (15849 bytes) |
| `execution/attest_closure_v2.py` | `c28a9853f826d46e05ddebbd207ac85cc112aec90179a84b7a0800aeecfd6d84` (unchanged) |
| `execution/tests/test_closure_weak_links.py` | `cf0385d711d97fae5dc7656c4368bf2badb66cb2a07b0bab44e7f27cdebde417` (32601 bytes) |
| `SCRATCH/closure-remedy-checks/r2-impl2/full-suite-round2.json` | `95bd2c0ac7562e41186bd54944032ce247915f9188a09bfa4debc0d11c0c659b` |
| `SCRATCH/closure-remedy-checks/r2-impl2/identity-measurement.json` | `fc26f3de0ba5d5f5f138c161970c8e97df0e51eed18c635027c5bfd839cc7f2a` |
| `SCRATCH/closure-remedy-checks/r2-impl2/seed-v1-measurement.json` | `1ac03a9761f91ab7bc1c60faf40570d8a13ec5eea299cef15664cded41f57ccb` |

The signed files are unchanged. The suite record pins each one before and after the run:
- `attest_closure.py` `113d2d72…`
- `closure/native.py` `c264d61a…`
- `preparation.py` `9572aed3…`
- `verify_inputs.py` `7d6f89ce…`
- `run-config.json` `c817f678…`

## Required fixes

| Fix | Status | Evidence |
|---|---|---|
| **X1CW-R1-01**: identity of the loaded image | **Met in code. Test gap: X1CW-R2-01** | See "X1CW-R1-01 evidence" below. |
| **X1CW-R1-02**: R3′a–c | **Met** | See "X1CW-R1-02 evidence" below. |
| **X1CW-R1-03**: a present non-member stays on disk | **Met** | `tests:263–279`: `image_filetype` is called on the fixture (`assert_any_call(disk)`), the fixture pin is `images[0]`, the target is not cached, and no `-linked_dylibs` call is made for it. Under the round-1 mutant, the fake `dyld_info` returns `''` and the parse raises, so the test catches that mutant. |
| **Round-2 evidence** | **Met** | See "Round-2 evidence" below. |

### X1CW-R1-01 evidence

**The helper, `loaded_cache_image_uuid`** (`native_v2.py:26–67`), checks that:
- exactly one loaded image has the exact name;
- the full 32-byte header lies inside `_dyld_get_shared_cache_range` (`:45`);
- the magic is `MH_MAGIC_64` and `sizeofcmds` and `ncmds` are bounded (`:49`);
- the command walk is bounded;
- there is exactly one `LC_UUID`, with `cmdsize` 24;
- all bytes are consumed.

Every failure becomes `cannot establish active-cache identity`. Python integers cannot wrap around, and `uuid.UUID(bytes=…)` keeps the byte order.

**The I7 branch:**
- `-arch` is passed only when the file is on disk.
- The comparison is case-insensitive (`:213–215`), and a mismatch refuses with the required message.
- `identity_check` is added only to I7 records (`:216–228`).

**Tests:**
- match and field emission (`tests:220–233`);
- mismatch (`:251–256`);
- an absent member is queried without `-arch` (`:258–261`);
- helper refusals (`:337–384`).

R1 asserts the identity check on all three images (`:475`, `:486–491`). The rule-3 measurement gives equal UUIDs for each image:

| Image | UUID |
|---|---|
| `libsystem_kernel` | `7CFB3A10-ADB2-32C4-99B5-922A111B3224` |
| `libsystem_platform` | `4C6BD0FE-25F2-38D8-BE5C-24DAC1DEDCFB` |
| `libsystem_pthread` | `12342372-0084-37A5-9EA6-2E4DC940C685` |

### X1CW-R1-02 evidence

- **R3′a** (`:493–496`): an exact string match on the refusal.
- **R3′b** (`:498–510`):
  - the seed sha `6ca6a32e…` is pinned (`:42`) and checked first;
  - both collectors complete, and their `without_v2` outputs are equal;
  - one image, no dependencies, no cached images, no absent images;
  - equal `cache_files` and UUID, and identical ordered argv.

  Signed v1 completed on the seed (`seed-v1-measurement.json`).
- **R3′c** (`:512–550`):
  - v1's `cached` is captured from the `collect` frame;
  - UUIDs and dependency lists are compared;
  - the multiset `(argv, exit_code, stdout)` from v1 is included in v2's;
  - RemoteProcessingBlock (RPB) is absent, and every absent row's evidence is re-checked live;
  - the RPB probe records exist;
  - each I7 image has `identity_check`.

  The retained records hold 189 v1 records (2 `otool`, 187 `dyld_info`) and 605 v2 records. The RPB probe is v1 `000189` and v2 `000192`. No record mentions iOSSupport.

### Round-2 evidence

- `full-suite-round2.json`:
  - pins 3 + 5 files in both `digests_before` and `digests_after`, and they are identical (`digests_unchanged: true`);
  - exit code 0;
  - `Ran 332 tests in 71.030s`, `OK`, with no skips;
  - all five real-byte tests `ok`.
- Every real-byte record timestamp falls inside the suite window, 17:16:37–17:17:48.
- R2 leaves an empty run directory. This is consistent with the code: RemoteProcessingBlock is neither a cache member nor on disk, so the walk refuses before running any command.

**Scope.** No change goes beyond the required fixes, and the wrapper is unchanged. The response restates P3-D faithfully. It adopts P3-D for future execution but performs no observation.

## Issues

| ID | Severity | Location | Failure | Required fix |
|---|---|---|---|---|
| **X1CW-R2-01** | Minor | `tests/test_closure_weak_links.py:340–358`, `:371–373`; `native_v2.py:45`, `:49` | `test_header_outside_cache_refuses` places the header at `start+size`, just past the end of the reported range. Two checks refuse that case: `:45` and the `sizeofcmds` bound at `:49`, which is `> -32` there. The test therefore passes even if `:45`, or only its lower bound `start <=`, is removed [INFERENCE from the arithmetic; not executed]. For an address **below** the range start, the lower bound is the only guard. An image loaded from disk lies outside the cache range. If that guard regressed, the helper would return that image's UUID. That UUID equals the disk slice's, so the record would say `loaded_from_active_cache_range: true` falsely, and no test would fail. The guard is what gives the I7 identity claim its meaning, and X1CW-R1-01 required a catching test for "header outside the cache range". | Add a `LoadedCacheIdentityTests` case whose header is a fully valid `MH_MAGIC_64` Mach-O, readable, with exactly one `LC_UUID`, placed **below** the reported range start. Allocate one buffer and report the range as starting after the header block, with a size that keeps `:49` satisfied. Assert the refusal `cannot establish active-cache identity`. Keep the existing case beyond the end. |

## Peer findings resolved as suggestions

**S1. Checking the `dyld_info` probe for every absent row** (innovation-challenger). R3′c inspects the actual probe records only for RPB. For the other absent rows it relies on the collector's `dyld_info_stdout_empty` flag.
- A collector that skipped the probe would already fail S1, which asserts exactly one `-linked_dylibs` call (`tests:122`), and S7, where non-empty stdout must refuse.
- The real records are retained, so every probe can be audited.
- Recommended: for each absent target, assert a retained record with argv `/usr/bin/dyld_info -uuid -linked_dylibs <target>`, exit 0 and stdout `''`.

**S2. A helper failure must refuse at the branch level** (experiment-peer). Wrapping the call at `:213` in `except ValueError` and falling back to the CLI UUID passes the suite. Recommended: patch the helper with `side_effect=ValueError(...)` and assert that the collector refuses. The current code does not swallow the error (`:213–215`).

**S3. Isolate each bound check** (experiment-peer). Add cases for:
- `ncmds > sizeofcmds//8` alone;
- trailing slack, which only `offset != sizeofcmds` catches;
- an in-bounds `LC_UUID` with `cmdsize` 32, which only `command_size != 24` catches;
- a non-symmetric, mixed-case UUID fixture.

Also add an eight-byte alignment check on `cmdsize` (innovation-challenger). Misaligned commands are not plausible in a loaded Apple cache header.

**S4. Make R3′c non-vacuous.** Assert `len(prefix) == ` (the number of v1 `dyld_info` records − 1). The real prefix is not empty: there were 187 calls.

**S5. Record the environment of the measurement.** Record the outer process environment in rule-3 and P3-D observation records; `DYLD_*` variables can change where images load from. Pin `native_v2.py` in any measurement that imports it.

**S6. Pin the fixture modules.** Pin `test_closure_review.py` and `test_closure_round2.py`, which the new tests import, in the suite record.

**S7. UUID equality is the identity policy, not byte equality** (innovation-challenger). A direct comparison of the loaded image's dependency load commands would be stronger. It is not required: the round-1 policy stands, and the measurement shows equal UUIDs.

## Recommendation

1. Fix X1CW-R2-01 with the one test case specified, in the test file only. `native_v2.py` and the wrapper must keep their round-2 digests.
2. Adopt S1–S4 in the same edit; they cost little and touch only the test file. They are not conditions of sign-off.
3. Re-run the full suite as specified in step 2 of the execution outline.
4. Round 3 reviews only that delta: the test file and the new suite record.

## Alternatives

| Option | Why rejected |
|---|---|
| Treat X1CW-R2-01 as a suggestion and sign off now | The closure writes `loaded_from_active_cache_range: true` as evidence, and no test checks the only guard for addresses below the range. This conflicts with experiment rule 4 (the implementation must be validated against planted effects). Round 1 treated the same kind of gap the same way (X1CW-R1-03). |
| Require a comparison of in-memory load commands (S7) | It would change the identity policy fixed in round 1. Nothing shows a defect in the policy. |
| Treat innovation-challenger's probe-record gap as an issue | S1 and S7 already catch that mutant (see S1 above). Recorded as a suggestion. |

## Evidence

- `native_v2.py`:
  - `:26–67` (helper), `:45` (range), `:49` (bounds);
  - `:206–218` (I7 identity), `:228` (field).
- `tests/test_closure_weak_links.py`:
  - `:40–42` (record root, seed pin), `:122` (S1 probe count);
  - `:220–280` (I7 and X1CW-R1-03), `:337–384` (helper);
  - `:465–550` (real-byte tests).
- `full-suite-round2.json`: identical pins before and after, 332 tests OK.
- `identity-measurement.json`: three equal UUIDs.
- `seed-v1-measurement.json`: v1 completes on the seed.

## Risks

- **The delta could touch more than the test file.** If a fix edits `native_v2.py` or the wrapper, round 3 must review that code again in full.
- **Re-runs pollute the evidence.** Each run creates new directories under the hard-coded `REAL_RECORD_ROOT` (`tests:40`). Step 2 of the execution outline points the next run at a fresh root and lists the run directories.
- **P3-D.** It is unchanged from round 1 Q4 and still not in force. It takes effect when the zero-issue round restates it with the v2 argv.
- **Clock.** The deadline is 2026-12-24. One narrow round remains.
- **Reversibility.** All code is in new files, and no signed byte has changed.

## Execution outline

1. **Writer:** add the X1CW-R2-01 case, and optionally S1–S4, to `tests/test_closure_weak_links.py` only.
2. **Writer:** change `REAL_RECORD_ROOT` to `SCRATCH/closure-remedy-checks/r2-impl3`.
3. **Main:** run the full suite with `/Users/adamjackson/Projects/work/scout-x1/.tools/venv/bin/python -I -B -m unittest discover -s …/execution/tests -p 'test_*.py' -v`. The suite record must:
   - pin, at start and end, the three new files, the five signed files and the two imported fixture modules;
   - list every real-byte run directory with a manifest digest;
   - show zero errors, zero failures and zero skips.
4. **Main:** write `closure-weaklink-impl-round-3-response.md`.
5. **Council:** round 3, scoped to the delta. On zero issues, that round signs off `closure-weaklink-implementation` and restates P3-D with the v2 argv. Steps 4–10 are then authorised.

## How the peers shaped this record

- **innovation-challenger:**
  - confirmed that the helper's bounds, magic, UUID handling and case handling are sound, and that no new path widens acceptance;
  - raised the gap in probe records for every absent row, which became S1;
  - raised alignment, the ctypes fault limit and the stronger-identity alternative, which became S3 and S7.
- **experiment-peer:**
  - found X1CW-R2-01 by working the range arithmetic;
  - verified the suite record line by line, including the timestamps;
  - counted the retained records;
  - checked the seed pin;
  - found the swallowed-failure and redundant-bound mutants (S2, S3), the vacuous prefix check (S4), the missing outer environment (S5) and the re-run pollution, which became step 2 of the execution outline.
- **My own analysis:**
  - verified the code against each round-1 requirement;
  - judged innovation-challenger's probe gap a suggestion, because S1 and S7 catch that mutant;
  - kept X1CW-R2-01 as an issue, for consistency with X1CW-R1-03 and rule 4;
  - limited round 3 to the delta.

## New issues

X1CW-R2-01 (Minor). This round signs nothing off.

Verdict: 1 ISSUES
