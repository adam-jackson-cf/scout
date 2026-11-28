# X1 preparation runtime closure (b): weak-link implementation review — round 1

**Scope.** This round reviews remedy A′, implemented under `preparation-reviews/closure-preflight-remedy-round-2.md` (SHA-256 `6ebcc683fc32f960d275c4786b40a5e8bd50354f9b11075e25f2261dc268fa01`). It covers:
- the three new files against I1–I7, including the binding and seal consumer confirmation (I6);
- the R3 disposition;
- the I7 amendment;
- the M7 disposition.

It signs off no code. It authorises no GR, no journalled preflight and no canonical closure. Under R-064, Main adopts these recommendations without asking the owner.

**Method.**
- The council was read-only. The lead ran no command, test, hash or journal.
- Two independent peer analyses were run: `innovation-challenger` (a different model family) and `experiment-peer`.
- All digests below are Main's `shasum -a 256`.
- Paths are relative to `evaluation/x1-triage-ranking/` unless absolute.
- `SCRATCH` = `/Users/adamjackson/Projects/work/x1-preparation-scratch`.
- One external source was read: Apple's open-source dyld [`other-tools/MiscFileUtils.cpp` `forEachFileGetSlices`](https://github.com/apple-oss-distributions/dyld/blob/fd8d0c4d52320ebf64db34f3cb280310d905c5ae/other-tools/MiscFileUtils.cpp). The peers cited Apple support notes [102657](https://support.apple.com/en-gb/102657) and [101591](https://support.apple.com/en-gb/101591) on Background Security Improvements.

## Reviewed bytes

| File | SHA-256 |
|---|---|
| `execution/closure/native_v2.py` | `a22a2c7f594672eae845a7c56412ab0c846030c08ee40a0919e7a945e975a739` |
| `execution/attest_closure_v2.py` | `c28a9853f826d46e05ddebbd207ac85cc112aec90179a84b7a0800aeecfd6d84` |
| `execution/tests/test_closure_weak_links.py` | `e19b0d25d3017923a2f7fba591db3865db48f35ea323526cf827f72b4b2d0ed7` |
| `preparation-reviews/closure-weaklink-impl-round-1-response.md` | `245531d71ce046652d289a5245ab87b56a9482843ce54088f5ede92adf0873c2` |
| `preparation-reviews/closure-weaklink-prescan-2026-11-10.json` | `2475337905376d64bba2aec29204a225edffc89abc416a4ed43943c5228aa19a` |
| `preparation-reviews/closure-weaklink-prescan-2026-11-10-rev2.json` | `79dabedcbafc06b037487b7614811253633461db5c8f0f49dc0c67fe2a089446` |
| `preparation-reviews/closure-weaklink-prescan-2026-11-10-rev3.json` | `4d982c94f3752c84463fd492c7de8f8d13d783042a3379cf6e5998482a69e1e0` |

The signed files are unchanged. Their digests are listed in `binding-closure-confirm-round-2.md`:
- `attest_closure.py` `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892`
- `closure/native.py` `c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb`
- `preparation.py` `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38`

Evidence read, with digests as the writer reported them:
- `SCRATCH/closure-remedy-checks/r2-m4/i7-active-architecture-analysis.json` `dd57dfd0…`
- `…/reachable-disk-classification.json` `b0dd1ebb…`
- `…/actual-loaded-system-image-source.json` `e0c4b783…`
- `…/present-kernel-dyld-info.json` `d801e8ce…`
- `…/present-platform-arm64e-dyld-info.json` `cf5853d4…`
- `…/alias-disk-classification.json` `2ed439b5…`
- `SCRATCH/closure-remedy-checks/r2-m7/command.json`
- `SCRATCH/closure-remedy-checks/r2-m5/census.json` and `commands/000027.json`, `000028.json`
- the three `closure-weaklink-impl-round-1-suite-*.json` records

## Q1. Conformance with I1–I7

**Met.**
- **I1, files.** Exactly three new files exist. No signed byte changed, and the five signed digests match the response.
- **Weak-absence proof** (`native_v2.py:141–192`). An edge is recorded as absent only when all of the following hold:
  - the target is not a cache member (`:144`);
  - the edge is weak, under `/usr/lib/` or `/System/Library/` (`:178`);
  - `filesystem_entries` is empty: the target and every overlay root were checked with `os.lstat`, and any `OSError` other than "not found" refuses (`:43–53`);
  - the `dyld_info` stdout is `''` (`:181–183`), and `Commands.run` raises on a non-zero exit (`attest_closure.py:393–394`).

  The probe is memoised per target, and parents are accumulated. A parse or identity failure on a cache member always refuses (`:165`); it is never turned into absence.
- **Strictness of required links.**
  - `effective_weak` excludes `re-export` (`:89–90`).
  - Duplicate edges in one image are weak only if every occurrence is weak (`:166–168`).
  - A required edge to a target already recorded as weak-absent refuses at `:192` in either visit order (test S8).
  - On disk, a link is weak only when its load commands are exactly `['LC_LOAD_WEAK_DYLIB']` (`:250`).
  - Missing `@rpath`, `@loader_path` and `@executable_path` dependencies keep v1's refusal.
- **Parser.**
  - The identity, header, section, column-header, anchored-path, token and duplicate rules follow I2.
  - Lines under `-rpaths:` are never dependencies, and the iOSSupport path is kept whole.
  - The walk uses an explicit stack in v1's depth-first pre-order, and each record is inserted before its edges are processed.
- **Wrapper** (`attest_closure_v2.py`).
  - `install()` checks both `SIGNED_BASE` digests before exactly two rebindings: `native_closure` and `own_sources`.
  - Rebinding happens only in `main()`; importing the module mutates nothing (W5).
  - `own_sources` returns 5 + 2 rows in order (W3).
  - `native_v2` is imported at top level, so `producer_pins` captures `native_v2.py` and `native.py`.
  - A refusal propagates as an exception, after `journal.finish('refused', exit_code=1)` (`attest_closure.py:854–855`).
- **I4 and I6, consumers.**
  - `attest()` adds only rows that carry a `path` (`:716–721`). Absent rows carry no `path`/`sha256` pair.
  - `verify_installed_closure` (`:474–494`) never reads `native_dependencies`.
  - S13 runs the signed `preparation.verify_tree_pins` and `verify_installed_closure` on a v2 fixture.
  - No consumer code or configuration changes are needed. The binding and seal consumer confirmation is satisfied for the reviewed design.
- **Suite records match the response.**
  - initial: 319 tests, 2 errors
  - intermediate: 321 tests, 2 errors
  - architecture: 322 tests in 26.309s, `FAILED (errors=1)`; the only error is R3, there are no skips, and R1 and R2 ran and passed.
- **Prescan revisions.**
  - rev1 stopped correctly on `-help`.
  - rev2 missed the multi-header present files. Its own census shows 3656 headers for 3649 names: `libsystem_{kernel,platform,pthread}` three times each, and `/usr/lib/dyld` twice. `/usr/lib/dyld` is not reachable, so this is harmless.
  - rev3 supersedes rev2 openly and does not overwrite it.

**Not met.** The I7 identity is not established (X1CW-R1-01). The R3 acceptance criterion cannot be met as written (X1CW-R1-02). One branch of the I7 amendment has no catching test (X1CW-R1-03).

## Issues

| ID | Severity | Location | Failure | Required fix |
|---|---|---|---|---|
| **X1CW-R1-01** | Major | `native_v2.py:150–177` (I7 branch); `tests/test_closure_weak_links.py` I7 tests; response § I7 | See X1CW-R1-01 below. | See X1CW-R1-01 below. |
| **X1CW-R1-02** | Major | `tests/test_closure_weak_links.py:394–399` (R3); I5 | See X1CW-R1-02 below. | Replace R3 with R3′a–c, as specified under Q2. |
| **X1CW-R1-03** | Minor | `native_v2.py:143–147`; tests | See X1CW-R1-03 below. | See X1CW-R1-03 below. |

### X1CW-R1-01 (Major): the I7 identity comes from the disk file

**Failure.**
- The amendment says it uses "the active-cache representation". In fact, `dyld_info -arch arm64e -uuid -linked_dylibs <present path>` reads the **on-disk fat file**. `forEachFileGetSlices` maps the disk file first and looks in the dyld cache only when the path cannot be opened.
- The writer's own evidence shows this: the call without `-arch` prints x86_64, arm64 and arm64e headers (`present-kernel-dyld-info.json`), while the cache is arm64e only.
- The UUID and link list of the disk slice are therefore stored as a cached record tagged `active_cache_uuid`, and the disk bytes are not pinned.
- `actual-loaded-system-image-source.json` proves only that the loaded headers lie inside the cache range. It never compares their UUID.
- This affects three real, exercised images: `/usr/lib/system/libsystem_{kernel,platform,pthread}.dylib`.
- Background Security Improvements can patch system libraries between OS builds (Apple 102657). Whether the disk and cache copies are equal today is unmeasured.

**Required fix.**
1. **New helper.** Add a patchable module-level helper, `loaded_cache_image_uuid(name)`, using ctypes on `CDLL(None)`. It:
   - requires exactly one loaded image whose `_dyld_get_image_name` equals `name`;
   - requires its `_dyld_get_image_header` address to lie inside `_dyld_get_shared_cache_range`;
   - requires the header magic to be `MH_MAGIC_64`;
   - walks `ncmds` load commands within `sizeofcmds`;
   - requires exactly one `LC_UUID`, which it returns in `dyld_info`'s format.

   Any other outcome raises `ValueError('cannot establish active-cache identity: '+name)`.
2. **When the helper runs.** In the I7 branch only (`on_disk` and cache member), after parsing, require `loaded_cache_image_uuid(target)` to equal the parsed UUID, case-insensitively. Otherwise raise `ValueError('on-disk alternate differs from active-cache image: '+target)`.
3. **New record field.** Add `identity_check: {'disk_alternate_architecture': <arch>, 'loaded_cache_image_uuid': <uuid>, 'loaded_from_active_cache_range': true}` to the cached record. Do not add it to any other record.
4. **Synthetic tests:**
   - match → the field is present;
   - mismatch → refuses;
   - the image is not loaded → refuses;
   - the header is outside the cache range → refuses;
   - an absent cache member is queried without `-arch`.
5. **Extend R1** to assert that all three images are in `cached_images` with `identity_check`, and that the check passes on real bytes.
6. **Rule-3 measurement in the round-2 response:** the `LC_UUID` of each of the three images loaded in memory, against its `dyld_info -arch arm64e` UUID.

### X1CW-R1-02 (Major): R3 cannot pass on this host

**Failure.** R3 can never compare v1 with v2 on this host:
- (a) Signed v1 refuses `/bin/sh` at RemoteProcessingBlock (traceback in `suite-architecture.json`).
- (b) Without (a), v1 would still queue the three present fat alternates to disk from `libSystem`'s links. There, `image_filetype` raises "no unique arm64 slice" (`native.py:163–165`; `reachable-disk-classification.json`). v1 therefore cannot complete any real walk that reaches `libSystem` [INFERENCE from (b) and the fact that every real macOS Mach-O executable links `libSystem`].
- (c) v2 on `/bin/sh` has non-empty `absent_weak_links`, so the assertion at `:398` is false for v2 as well.

As a result, the suite stays red and I5 is unmet. The writer was right not to change the seed silently.

**Required fix.** Replace R3 with R3′a–c, as specified under Q2.

### X1CW-R1-03 (Minor): no test that a non-member disk file stays on disk

**Failure.** No test covers a present `SYSTEM_PREFIXES` file that is **not** a cache member and is reached from a cached parent; that file must go to disk. Removing the membership condition for on-disk targets passes the whole suite, R1's real graph included. That mutant would record unpinned disk bytes as cached.

**Required fix.** Add a synthetic test with:
- `Path.is_file` true for the target;
- membership false.

It asserts that the target is queued and classified (`image_filetype` is called), no `dyld_info` call is made for it, and its image is pinned in `images`.

## Q2. R3: replaced, not accepted

R3′a–c replace R3. All three are real-byte tests behind the existing 25G72 gate. Each uses its own recording `Commands`, with the same closed environment, and `CACHE_ROOT`.

**R3′a. v1's refusal stays visible.**
- `original.collect(['/bin/sh'], …)` must raise `ValueError` with exactly `'ambiguous cached native image identity: '+REMOTE`.

**R3′b. A real null on which both collectors pass.**
- Seed: `/Users/adamjackson/.local/share/mise/installs/go/1.26.10/src/runtime/race/race_darwin_arm64.syso`. It is a real thin Mach-O, `MH_OBJECT`, with no dependencies and no rpaths (`r2-m5/census.json`; `commands/000027.json`, `000028.json`).
- Both collectors must complete. Assert:
  - `without_v2(v1) == without_v2(v2)`;
  - exactly one image, with an empty dependency list;
  - `cached_images == []` and `absent_weak_links == []`;
  - equal `cache_files` and `active_cache_uuid`;
  - an identical ordered argv list.
- Pin the seed's sha256 in the test. Measure it and report it in the round-2 response.
- If v1 does not complete on this seed, stop and report it. Do not substitute another seed.

**R3′c. Real cached-graph conformance up to v1's refusal.**
- Run v1 as in R3′a, and capture its `cached` dict from the `collect` frame of the exception traceback. This is read-only introspection; v1 is unchanged.
- Run `native.collect(['/bin/sh'], …)`, which must complete. Then assert:
  1. Every install name in v1's `cached` is in v2's `cached_images`, with an equal `image_uuid` and equal `dependency_install_names`. The only allowed difference is a v2 name equal to `'/System/iOSSupport'` plus v1's name (F7(b)).
  2. Every v1 command record `(argv, exit_code, stdout)` occurs in v2's records. This is a multiset inclusion, so the I7 reordering is allowed. The RemoteProcessingBlock probe record must match v2's absence probe.
  3. v2's `absent_weak_links` contains RemoteProcessingBlock, and every row meets I2(b)4.
  4. The three I7 images carry `identity_check` (X1CW-R1-01).

**Why not accept R3 as an unmeasured result?**
- R3′b gives a passing real-byte null on both collectors.
- R3′c compares the two parsers on every real image v1 reaches before its documented refusal.
- The synthetic S12 cases remain the proof of v1/v2 parity in visit order.

## Q3. I7 amendment: routing accepted, identity evidence rejected until X1CW-R1-01 is fixed

**Accepted.** A present `SYSTEM_PREFIXES` target that is proven to be a member of the active cache, and is reached from a cached parent, is resolved as a cached image. This is sound because:
- dyld loads these three images from the active cache range (`actual-loaded-system-image-source.json`);
- the cache family is pinned and re-verified at every launch;
- of the reachable names, exactly three are present files, all are cache members, and all have three slices (`reachable-disk-classification.json`);
- the architecture comes strictly from the magic of the UUID-matched cache, `dyld_v1  arm64e`;
- the strict parser is unchanged, and no fallback exists.

The disk route for on-disk parents is unchanged. It would refuse on these three images, which fails safe; M5 found zero such edges.

**Rejected as submitted.** The identity and edges are those of the disk slice, and nothing ties them to the image that is actually loaded (X1CW-R1-01). After the fix, I7 is accepted as amended.

## Q4. M7: protection without root is detection-and-stop

**What M7 recorded** (`r2-m7/command.json`):
- `AutomaticDownload = 1` and `SplatEnabled = 1`;
- updates on offer: macOS 26.7.3 (`MSU_UPDATE_25G342…`), macOS 27.2.1 and Safari 27.2;
- current build `26.6 (25G72)`.

The host is MDM-managed and has no root, so prevention cannot be achieved.

**A correction to the brief.**
- Binding re-checks the cache-family **file pins** at every launch, through `verify_installed_closure` → `validate_pin` on `original_runtime_files` (`attest_closure.py:482–484`; `acceptance.py:81–82`). It does not query the active cache UUID or the OS build.
- A change to the pinned bytes therefore causes a fail-closed refusal, not a false accept.
- A change to which cache is active, with the pinned files intact, is not excluded [INFERENCE].
- The real exposure is to **availability**: the canonical `closure.json` is create-once.

**Disposition.** The council replaces M7's requirement to "turn it off before Step 5" with protocol **P3-D**. P3-D needs no root, no network and no settings change. It takes effect when it is restated in the zero-issue implementation round.

**P3-D observation O.** Each observation is recorded with its argv, exit code, stdout, digests and timestamps:
1. `/usr/bin/sw_vers`: ProductVersion `26.6`, BuildVersion `25G72`, no `ProductVersionExtra`.
2. `/usr/sbin/sysctl -n kern.bootsessionuuid kern.boottime kern.osversion`.
3. The active cache UUID, from a fresh `$PY -I -B` process via `_dyld_get_shared_cache_uuid`. It must equal `d64a142b-8533-3e09-968a-6ad3e176e3f6` and the UUID in bytes 88–104 of `CACHE_ROOT/dyld_shared_cache_arm64e`.
4. The sha256 of every file in the active cache family. Take this at O0, O2, O4 and before seal.
5. `defaults read /Library/Preferences/com.apple.SoftwareUpdate`. Also read `/Library/Managed Preferences/com.apple.SoftwareUpdate.plist`: record whether it exists, and its content if it is readable. Run `softwareupdate --history`. Do **not** run `--list`.

**Schedule.**

| Observation | When |
|---|---|
| O0 | before GR |
| O1 | before Step 5 |
| O2 | after Step 5 |
| O3 | before Step 7 |
| O4 | after Step 7 |
| Session checks | items 1–3 before each binding launch session (runbook) and before seal |

**Stop rules.**
- Items 1, 3 and 4 must equal O0 at every observation.
- `kern.bootsessionuuid` must stay constant from O1 to O4.
- The preflight's and the canonical closure's `native_dependencies.active_cache_uuid` and `cache_files` must equal O0. Their `platform.version` must equal O0's `sw_vers`.
- Any difference, or any failed observation, means: stop before the next step, retain the outputs, create nothing, and open a council round.
- From O1 until seal, the owner does not restart and does not accept update prompts. These are user-level actions, not settings changes.

**Non-blocking.** The owner asks the MDM administrator to defer updates, including Background Security Improvements, until seal (2026-12-24).

**RESULTS.md Deviations.** Record that M7 prevention could not be achieved and that P3-D replaced it. The residual risk is a forced update: it costs a refusal and a recovery round, but not integrity.

## Recommendation

1. Fix X1CW-R1-01 to 03 exactly as specified.
2. Replace R3 with R3′a–c.
3. Adopt I7 as amended.
4. Adopt P3-D for M7.
5. Hold round 2 on the new digests.
6. Steps 4–10 of round 2 wait for a zero-issue implementation round.

## Alternatives

| Option | Tradeoff | Outcome |
|---|---|---|
| Accept R3 as unmeasured | No real-byte equivalence evidence at all, when R3′b and R3′c are feasible | Rejected |
| A pruned `/bin/sh` graph, or patching v1 to get past RemoteProcessingBlock | Changes signed behaviour, or deletes edges | Rejected |
| A census oracle (`r2-m4/census.json`) for v2's cached records (experiment-peer) | Independent and strong, but depends on the writer's census parser | Suggestion S3 |
| `/usr/lib/dyld` as the R3′b null (experiment-peer) | Exercises a fat file, but the `otool` behaviour on it is unmeasured | The syso was chosen because its real outputs are already recorded |
| I7: parse the cache file directly to locate the image | Most direct, but needs a large parser for the split cache | Rejected in favour of the UUID of the image loaded in memory |
| I7: pin the disk alternate only | Makes the record honest, but proves nothing about the loaded image | Suggestion S2, alongside X1CW-R1-01 |
| M7: keep Step 5 blocked until an MDM freeze | No prevention is in the team's control, and the deadline is 2026-12-24 | Rejected; P3-D instead |

## Evidence

- `native_v2.py:141–192` (edge rules), `:150–177` (I7 branch).
- `attest_closure.py:393–394` (non-zero exit raises), `:474–494` (verifier), `:716–721` (inventory), `:854–855` (refusal).
- `MiscFileUtils.cpp` `forEachFileGetSlices`: the disk file is read first, the cache only when the path is absent.
- `present-kernel-dyld-info.json`: three headers. `present-platform-arm64e-dyld-info.json`: one header, `4C6BD0FE-…`.
- `actual-loaded-system-image-source.json`: header addresses only, no UUID.
- `i7-active-architecture-analysis.json`: cache magic `dyld_v1  arm64e`, header UUID `d64a142b-…`.
- `suite-architecture.json`: 322 tests, R3 error at `native.py:148`.
- `r2-m5/commands/000027.json`, `000028.json`: the syso's real `otool` outputs.
- `r2-m7/command.json`: the update settings and the offered updates.

## Risks

- Seed parsing or `otool` on the syso may behave differently under v1 than the census suggests. R3′b then stops and is reported; the seed is not swapped.
- `loaded_cache_image_uuid` depends on the three images being loaded in the attesting process. This was measured for the pinned interpreter. A failure refuses at GR, before any journal.
- The census header count and multi-slice traps may hide another present member in the full seed set (not only CoreAudio). GR (Step 4) runs the full seeds unjournalled, so the cost is one more council round.
- P3-D detects drift and stops; it does not prevent updates. A forced MDM update inside the window costs availability.
- **Reversibility.** All code is in new files, and no signed byte changes. The irreversible step is still canonical Step 7.

## Execution outline

1. The writer implements X1CW-R1-01 to 03 and R3′a–c in the three new files only.
2. The writer takes the rule-3 measurements:
   - the in-memory `LC_UUID` against the `-arch arm64e` UUID for the three images;
   - the syso's sha256.
3. Main runs the full suite with the pinned `$PY -I -B`. The suite record must:
   - pin, at the start and the end of the run, the three new and five signed digests;
   - show zero errors and zero skips. A skip of R1, R2 or R3′ counts as a failure.

   The real-byte command records are retained under `SCRATCH/closure-remedy-checks/r2-impl2/`, not in a temporary directory.
4. Main writes `closure-weaklink-impl-round-2-response.md` with the digests, the suite record and the measurements.
5. The council holds round 2. Its sign-off restates P3-D and P3/P4 with the v2 argv.
6. Only after zero issues: O0, GR (Step 4), then Steps 5–10 of remedy round 2, with P3-D in place of M7.

## Suggestions (not issues)

- **S1.** Before allowing absence, require a canonical target (`os.path.normpath(target) == target`, no `..`).
  - Today this is not plausible: M5 found zero absolute missing on-disk dependencies, and the cached edges are Apple install names.
  - The challenger showed that `/usr/lib/../…` passes the prefix check.
- **S2.** Pin the disk alternate as `identity_check.disk_alternate`, so the bytes that supplied the edges are bound.
- **S3.** Add a census-oracle check of v2's cached records and absent set, as experiment-peer proposed.
- **S4.** Isolate the duplicate-section guard and the header-count guard in S9:
  - a duplicate `-linked_dylibs:` followed by a valid column header and a link line;
  - two headers with one UUID.
- **S5.** For a present target linked once weakly and once as required, assert `requirement == 'required'` on both rows.
- **S6.** W4's argv check is circular: the test asserts the `sys.argv` it set itself. Assert instead that `preparation.producer_pins()` includes `native_v2.py`.
- **S7.** Enumerate the overlay probe paths once (`native_v2.py:180` and `:45`).

## How the peers shaped this record

- **innovation-challenger**
  - Confirmed independently, from the dyld source, that `-arch` does not force the cache to be read (01).
  - Found the real null seed, the syso (R3′b).
  - Proposed the common-prefix design that became R3′c.
  - Showed that binding verifies file pins, not the active UUID. This corrected the brief's premise for M7.
  - Raised the lexical-prefix escape, which I downgraded to S1 because the frozen inputs contain no such path.
- **experiment-peer**
  - Separated R3's three independent causes, including v1's second refusal on the fat alternates. This established that real v1 parity is impossible past `libSystem`.
  - Found the untested non-member branch (03).
  - Traced the census header-count clue that rev2 missed.
  - Showed that the suite records are not bound to the reviewed digests (step 3).
  - Reported the offered updates, the unread managed preferences and the Background Security Improvements behaviour, which shaped P3-D.
- **My own contributions**
  - The in-memory `LC_UUID` check as the narrow I7 fix.
  - Capturing v1 `cached` from the traceback, so the comparison is against signed v1 rather than a reimplementation.
  - The multiset form of the command-record comparison, so the I7 reordering does not create false failures.
  - P3-D's stop rules and schedule.

## New issues

X1CW-R1-01 (Major), X1CW-R1-02 (Major), X1CW-R1-03 (Minor). This round signs nothing off.

Verdict: 3 ISSUES
