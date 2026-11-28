# X1 preparation runtime closure (b): weak-link implementation review — round 3

**Scope.** This round reviews only the delta named in the execution outline of round 2:
- the test file;
- the new suite record.

It checks the fix for X1CW-R2-01 from `preparation-reviews/closure-weaklink-impl-round-2.md` (SHA-256 `a2b11fbfbaff482326ed27333e22f36c88b576befea7768d9f8e1a99875983aa`), and the optional S1–S4. The writer's response is `preparation-reviews/closure-weaklink-impl-round-3-response.md` (`e5b8ae2997c3da608b3629aa93bbb4b936df82df6afa43f2dd137617db40d0a9`). Under R-064, Main adopts this record without asking the owner.

**Method.**
- The council was read-only. The lead ran no command, test, hash or journal.
- Two independent peer analyses ran: `innovation-challenger` and `experiment-peer`.
- Digests are Main's `shasum -a 256` or the pinned suite record. The peers compared the records with one another and did no fresh hashing.
- `SCRATCH` = `/Users/adamjackson/Projects/work/x1-preparation-scratch`.

## Reviewed bytes

| File | SHA-256 |
|---|---|
| `execution/tests/test_closure_weak_links.py` | `2a7ee81554000968590cfbcd8ce4649a757a346f3ed24f496db37345e6b89c87` (35588 bytes) |
| `execution/closure/native_v2.py` | `5e320ed37a5b106e3a37984e6137e78f87422aa525792018032117d66f4dee09` (unchanged since round 2) |
| `execution/attest_closure_v2.py` | `c28a9853f826d46e05ddebbd207ac85cc112aec90179a84b7a0800aeecfd6d84` (unchanged since round 1) |
| `SCRATCH/closure-remedy-checks/r2-impl3/full-suite-round3.json` | `ebcce6d490de2500ba5d991055eda42e081dba594239b38d334f57f6472bb2c6` |

The suite record also pins these files, unchanged, at start and end:
- `attest_closure.py` `113d2d72…`
- `closure/native.py` `c264d61a…`
- `preparation.py` `9572aed3…`
- `verify_inputs.py` `7d6f89ce…`
- `run-config.json` `c817f678…`
- `tests/test_closure_review.py` `534e4754…`
- `tests/test_closure_round2.py` `70efac05…`

The two fixture modules match their earlier pins in `binding-closure-confirm-round-2.md:34–35`.

## Findings

| Item | Result | Evidence |
|---|---|---|
| **X1CW-R2-01** | **Met** | See "X1CW-R2-01" below. |
| **S1** | Met | `tests:591–593`: for **every** absent target, R3′c requires a retained record with the exact argv `/usr/bin/dyld_info -uuid -linked_dylibs <target>`, exit 0 and stdout `''`. RemoteProcessingBlock (RPB) being absent guarantees the loop runs. |
| **S2** | Met | `tests:258–264`: the helper raises `ValueError`, and the test requires that exact refusal. A fallback to the CLI UUID would make the test fail. |
| **S3** | Met | Each case reaches only one guard:<br>• `ncmds` bound plus exactly one memory read (`:398–402`);<br>• trailing slack, caught only by `offset != sizeofcmds` (`:404–407`);<br>• an in-bounds `LC_UUID` with `cmdsize` 32, caught only by `!= 24` (`:409–412`);<br>• an asymmetric mixed-case UUID, for byte order and `.upper()` (`:414–416`);<br>• a case-insensitive comparison in the collector (`:266–273`).<br>The production alignment suggestion was correctly not applied: it would change the frozen collector. |
| **S4** | Met | `tests:570`: the captured prefix size equals the number of v1 `-linked_dylibs` records minus the final refusal probe. The retained v1 records `000188`/`000189` show a parsed image followed by the refusal, so the check is not vacuous on the frozen host. |
| **Evidence binding** | Met | See "Evidence binding" below. |
| **Scope** | Met | See "Scope" below. |

### X1CW-R2-01

The new test is at `tests:357–370` and `:394–396`. It uses one buffer:
- a 32-byte valid `MH_MAGIC_64` header;
- one `LC_UUID` command (24 bytes);
- 64 padding bytes and a terminator, 121 bytes in all.

The reported range starts at B+56 with size 65, and the header sits at B, below it. Against `native_v2.py:45` and `:49`:
- the lower bound `B+56 ≤ B` is false;
- the upper header bound `B ≤ B+89` is true;
- the `sizeofcmds` bound `24 > 89` is false;
- the `ncmds` bound `1 > 3` is false;
- the command walk returns the UUID.

Only `start <=` refuses [INFERENCE from the arithmetic, which all three council analyses worked independently]. Every read stays inside the live allocation. The case beyond the end (`:390–392`) is unchanged.

### Evidence binding

- `REAL_RECORD_ROOT` is `r2-impl3` (`tests:40`). It was a fresh root.
- Ten pins are identical at start and end (`digests_unchanged: true`).
- Exit code 0: `Ran 339 tests in 65.466s`, `OK`, with no skips, failures or errors. R1, R2 and R3′a–c ran and passed.
- `real_byte_runs` lists five run directories, each with its manifest digest. Record counts are 603 + 0 + 189 + 4 + 794 = 1590.
- No directories exist outside the manifests, and the record timestamps fall inside the suite window (17:26:28–17:27:34).

### Scope

- Only the test file changed. This is supported by the nine other digests matching their earlier records; the writer's "edited only" claim is otherwise unverified.
- The response matches the suite record and the code.
- It performed no O0, GR, journal, preflight or observation.

### Binding and seal consumer confirmation (I6)

This is confirmed for the signed collector and wrapper:
- `attest()` adds only rows that carry a `path` (`attest_closure.py:716–721`).
- `verify_installed_closure` (`:474–494`) never reads `native_dependencies`, and it re-checks the cache-family pins at every launch.
- Absent rows carry no `path`/`sha256` pair.
- I7 records add `identity_check`, which holds no file pin.
- S13 runs the signed `verify_tree_pins` and `verify_installed_closure` on v2 output.

None of the following changes: `binding/acceptance.py`, `preparation.py`, `run-config.json`, `preparation/pins.json`, `OUT/reviews/closure-model-free.json`. The later `execution-source-manifest.json` and `reviews/implementation.json` must include `attest_closure_v2.py`, `closure/native_v2.py` and `tests/test_closure_weak_links.py`.

## Suggestions (not conditions of sign-off)

1. Add `assertGreater(len(prefix), 0)` next to the S4 cardinality check. Both peers suggested it, and it is redundant on 25G72.
2. In later records, cite `binding-closure-confirm-round-2.md:34–35` whenever the fixture modules are called "unchanged".
3. Main should record `git status --porcelain` for `execution/` in the Step 8 gate record. This turns the writer's "only the test file was edited" claim into an observation.

## P3-D, P3 and P4 restated with the v2 argv

The paths are:
- `PY` = `/Users/adamjackson/Projects/work/scout-x1/.tools/venv/bin/python`
- `STUDY` = `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking`
- `OUT` = `STUDY/preparation-run`
- `CACHE_ROOT` = `/System/Volumes/Preboot/Cryptexes/OS/System/Library/dyld`

**v2 preflight argv (Step 5).**
- This is the argv of `OUT/journal/082476b0335b4deebe7390ca6b426c0a.start.json` (`3a13295dd6e4579ccd1ec5e86377227b7a9cf0bfe51406e4a6435e491885bd16`), with exactly one substitution: `argv[0]` becomes `STUDY/execution/attest_closure_v2.py`.
- It runs as `$PY -I -B STUDY/execution/attest_closure_v2.py` with:
  - `--config STUDY/execution/run-config.json`
  - `--clock OUT/clock.json`
  - `--pins STUDY/preparation/pins.json`
  - `--preflight`
  - `--runtime-root /Users/adamjackson/Projects/work/scout-x1`
  - `--module-cache SCRATCH/closure-modules-r2`
  - `--dyld-cache-root CACHE_ROOT`
  - the same 32 `--external-root` values, in the same order as that start record: `.tools`, cpython-3.11.11, go 1.26.10, the 12 module roots under `SCRATCH/closure-modules-r2`, and the 17 Homebrew Cellar roots.
- The environment must not contain `OPENROUTER_API_KEY`.

**Canonical argv (Step 7).**
- The same argv without `--preflight`, plus `--output OUT/closure.json`.
- It runs once, only after a passed Step 5 and a passed GV (P4).
- `OUT/closure.json` is create-once.

**GR (Step 4).** The `args` used by the harness are exactly the parsed values of the Step 5 argv. The stub id is `remedy-r2-gr`. Pass conditions are unchanged from remedy round 2, plus one: the three images `libsystem_{kernel,platform,pthread}` must carry `identity_check`.

**P3.** The following stay byte-stable until seal:
- the remedy-round-1 list;
- `SCRATCH/closure/082476b0335b4deebe7390ca6b426c0a/`;
- `SCRATCH/closure-preflight/082476b0335b4deebe7390ca6b426c0a.json`;
- `SCRATCH/closure-remedy-checks/r2-*`, including `r2-impl2/` and `r2-impl3/`;
- `SCRATCH/closure/remedy-r2-*`;
- the three signed-off files;
- the five signed base files.

No OS, cryptex or cache change is allowed. P3-D below enforces this by detection.

**P3-D** replaces M7's prevention requirement. It needs no root, no network and no settings change. Each observation O is recorded with its argv, exit code, complete stdout, digests and timestamps, and with the full outer process environment, with secrets checked by presence only:
1. `/usr/bin/sw_vers`: ProductVersion `26.6`, BuildVersion `25G72`, no `ProductVersionExtra`.
2. `/usr/sbin/sysctl -n kern.bootsessionuuid kern.boottime kern.osversion`.
3. A fresh `$PY -I -B` process reads `_dyld_get_shared_cache_uuid`. It must equal `d64a142b-8533-3e09-968a-6ad3e176e3f6`, and also bytes 88–104 of `CACHE_ROOT/dyld_shared_cache_arm64e`.
4. The SHA-256 of every file in the active cache family. Take this at O0, O2, O4 and before seal.
5. `/usr/bin/defaults read /Library/Preferences/com.apple.SoftwareUpdate`. For `/Library/Managed Preferences/com.apple.SoftwareUpdate.plist`, record whether it exists and its content if readable. Run `/usr/sbin/softwareupdate --history`, never `--list`.

**Schedule:**

| Observation | When |
|---|---|
| O0 | immediately before GR |
| O1 | before Step 5 |
| O2 | after Step 5 |
| O3 | before Step 7 |
| O4 | after Step 7 |
| Session checks | items 1–3 before each binding launch session (runbook) and before seal |

**Stop rules:**
- Items 1, 3 and 4 must equal O0 at every observation.
- `kern.bootsessionuuid` must stay constant from O1 to O4.
- The preflight's and the canonical closure's `native_dependencies.active_cache_uuid` and `cache_files` must equal O0. Their `platform.version` must equal O0's `sw_vers`.
- Any difference, or any failed observation: stop before the next step, retain the outputs, create nothing, and open a council round.
- From O1 until seal, the owner does not restart and does not accept update prompts.

**Non-blocking.** The owner asks the MDM administrator to defer updates, including Background Security Improvements, until 2026-12-24.

## Authorised sequence (Steps 4–10 of remedy round 2)

Stop at the first failure. Each refusal needs its own recovery record and council round.

1. **O0**, then **Step 4 (GR)**. This is unjournalled, as specified in remedy round 2, with the addition above.
2. **O1**, then **Step 5**: one journalled preflight with the v2 preflight argv.
3. **O2**, then **Step 6 (GV)**: run `attest_closure.verify_installed_closure` and `preparation.verify_tree_pins` read-only on the preflight output. Record the times; any time over 1 h goes into the gate record as a clock risk.
4. **O3**, then **Step 7**: the canonical closure, once, with the canonical argv. Then **O4**. Then follow the binding runbook (RB-1), with the P3-D session checks.
5. **Step 8.** Write the gate record `preparation-reviews/closure-weaklink-gates-<date>.json`. It holds:
   - O0–O4;
   - the argv, exit codes, digests and timestamps of GR, Step 5, GV and Step 7;
   - the journal IDs;
   - the M-check records and the `r2-impl2`/`r2-impl3` evidence digests;
   - `git status --porcelain` for `execution/`.
6. **Step 9.** Add `RESULTS.md` Deviations entries for:
   - `4124…` (B′);
   - `0824…` (F1);
   - the v2 entrypoint and collector, with the absent weak links;
   - the I7 amendment and its identity check;
   - R3 replaced by R3′a–c, with the reason;
   - M7 prevention not achieved and replaced by P3-D;
   - GR and GV;
   - reproduce commands.
7. **Step 10.** Add research-ledger entries for the sources that justified decisions:
   - Apple dyld `dyld_info.cpp`;
   - `JustInTimeLoader.cpp`;
   - `MiscFileUtils.cpp` (`forEachFileGetSlices`, disk first);
   - `_dyld_shared_cache_contains_path` (bpo msg400797);
   - Apple support notes 102657 and 101591.

## How the peers shaped this record

- **innovation-challenger** confirmed the below-range arithmetic with exact offsets. It also verified that every hypothetical read stays inside the live allocation, and checked that S1–S4 are not vacuous.
- **experiment-peer:**
  - verified the ten pins, the 339 OK result and the 1,590-record manifests against the run trees and the timestamps;
  - grounded the "unchanged" claim for the fixture modules in the earlier pins;
  - suggested that Main confirm the edit claim, which became suggestion 3.
- **My own analysis:**
  - checked the arithmetic and the guards for each new test against `native_v2.py`;
  - restated the v2 argv from the `0824…` start record and the interpreter used in the gate record;
  - confirmed I6;
  - added the `identity_check` pass condition for GR, the record of the outer environment for P3-D (from round-2 suggestion S5), and the new ledger sources.

## New issues

None.

**Zero issues. Signed off.**

Signed-off scope: closure-weaklink-implementation
- `execution/closure/native_v2.py` `5e320ed37a5b106e3a37984e6137e78f87422aa525792018032117d66f4dee09`
- `execution/attest_closure_v2.py` `c28a9853f826d46e05ddebbd207ac85cc112aec90179a84b7a0800aeecfd6d84`
- `execution/tests/test_closure_weak_links.py` `2a7ee81554000968590cfbcd8ce4649a757a346f3ed24f496db37345e6b89c87`

The signed base stays unchanged:
- `execution/attest_closure.py` `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892`
- `execution/closure/native.py` `c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb`
- `execution/preparation.py` `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38`
- `execution/verify_inputs.py` `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`
- `execution/run-config.json` `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`

The authority chain is:
- `closure-preflight-remedy-round-2.md` `6ebcc683fc32f960d275c4786b40a5e8bd50354f9b11075e25f2261dc268fa01`
- `closure-weaklink-impl-round-1.md` `b50e016181651e8b98c33973c3a2869c563403025f91cd2785707c0940c7bc99`
- `closure-weaklink-impl-round-2.md` `a2b11fbfbaff482326ed27333e22f36c88b576befea7768d9f8e1a99875983aa`

Suite evidence: `full-suite-round3.json` `ebcce6d490de2500ba5d991055eda42e081dba594239b38d334f57f6472bb2c6`.

P3-D, P3 and P4 are restated above with the v2 argv. Steps 4–10 are authorised in the order given.

Verdict: ZERO ISSUES — signed off
