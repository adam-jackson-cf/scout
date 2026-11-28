# X1 preparation runtime closure (b): remedy for the refused weak-link preflight — round 2

**Scope.** This round does three things:
- It chooses the remedy for the refused journalled closure preflight `082476b0335b4deebe7390ca6b426c0a`, whose reason was `ambiguous cached native image identity: /System/Library/PrivateFrameworks/RemoteProcessingBlock.framework/Versions/A/RemoteProcessingBlock`.
- It specifies that remedy precisely enough for a writer.
- It signs off how that refusal is recovered at seal.

It does not sign off any code. The remedy needs new code, and that code needs its own zero-issue implementation round before any further journal runs. Under R-064, Main adopts this recommendation without asking the owner.

**Method.**
- Read-only. The council lead ran no command, test, hash, journal, attestation or seal. The only model calls were the two peer analyses (`innovation-challenger`, `experiment-peer`).
- All digests are Main's (`shasum -a 256`), as recorded in the journals, `closure-p4-attempt-3-2026-11-10.json` and the command records. Byte counts come from those records.
- External sources:
  - Apple's open-source dyld, read by `innovation-challenger`:
    - [`other-tools/dyld_info.cpp`](https://github.com/apple-oss-distributions/dyld/blob/fd8d0c4d52320ebf64db34f3cb280310d905c5ae/other-tools/dyld_info.cpp#L314-L345)
    - [`dyld/JustInTimeLoader.cpp`](https://github.com/apple-oss-distributions/dyld/blob/fd8d0c4d52320ebf64db34f3cb280310d905c5ae/dyld/JustInTimeLoader.cpp#L475-L518)
  - CPython's use of `_dyld_shared_cache_contains_path` ([bpo msg400797](https://bugs.python.org/msg400797)).

  Upstream source is not proof of how the installed binaries behave. The real-byte tests and the M-checks below establish that behaviour.
- Paths are relative to `evaluation/x1-triage-ranking/` unless absolute. The fixed paths are:
  - `SCRATCH` = `/Users/adamjackson/Projects/work/x1-preparation-scratch`
  - `ROOT` = `/Users/adamjackson/Projects/work/scout-x1prep`
  - `OUT` = `ROOT/evaluation/x1-triage-ranking/preparation-run`
  - `C` = `SCRATCH/closure/082476b0335b4deebe7390ca6b426c0a/commands/`
- Main's diagnosis (`dyld_info` on the missing image, the read-only walk of the link graph, the absent framework directory) and the M-checks below are measurement checks, which rule 3 permits. They look at no association between a predictor and an outcome.

## Evidence

| Item | SHA-256 (bytes) | Use |
|---|---|---|
| `execution/closure/native.py` | `c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb` (10692) | <ul><li>cache identification `:103–129`</li><li>`pending.sort` `:138`</li><li>`cached_image` `:140–167`: identity check `:145–148`, parser `:149–159`, per-image `cache_files` `:160–162`, recursion `:163–167`</li><li>on-disk walk `:169–210`</li></ul> |
| `execution/attest_closure.py` | `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892` (46996) | <ul><li>`Commands.run` `:379–396` (returns stdout only)</li><li>`native_closure` `:464–466`</li><li>`own_sources` `:469–471`</li><li>`verify_installed_closure` `:474–494`</li><li>`inventory_native_closure` `:601–604`</li><li>`attest` `:607–750`: `own_sources()` `:622`, native call `:715`, cache files into the inventory `:720–721`, `retain` `:724–726`</li><li>`main` `:809–856`: journal source pins `:838`</li></ul> |
| `execution/preparation.py` | `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38` | <ul><li>`validate_pin` `:198–201`</li><li>`producer_pins` `:324–336`</li><li>Journal `:363–367`</li><li>`verify_tree_pins` `:423–434`</li><li>`review_receipt` `:437–461`</li><li>`collect_journals` `:596–638`</li><li>seal record checks `:663–675`</li><li>closure recheck `:702–706`</li><li>source manifest `:709–713`</li></ul> |
| `execution/binding/acceptance.py` | (unchanged; digest per `binding-closure-confirm-round-2.md`) | <ul><li>`verify_installed_closure` at launch `:81–82`</li><li>source manifest/review `:340–346`</li><li>closure `verify_tree_pins` `:347–352`</li></ul> |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | `closure.json` kind `scout.x1.installed-closure.v1` `:79–80`; `source_manifest` `:144` |
| `preparation-reviews/closure-preflight-remedy-round-1.md` | `cb9fc32a92dfde04beb6f299297e0dd5d565562d207390fd040ff5e38b8e27c4` (25773) | F3; remedy B′; authority for the `4124…` recovery |
| `preparation-reviews/closure-p4-attempt-3-2026-11-10.json` | `9bf2ef06124551e9d1f064b6ca926d605a9661eb99b638414c85cbd768ea8246` | gates G0, J1, J2, G4, G5, G1 passed; 272 command records; failure trace |
| `preparation-reviews/closure-preflight-remedy-gates-2026-11-10.json` | `bd200ce998f80420bbf402a17f791635b325ef33f933d772b39b20bc3ae927d0` (1330583) | round-1 gate record |
| `OUT/reviews/operation-recovery-412402435e9f4ac9bb810a5520d102d5.json` | `ab7e30d20e9b5964423f79558ed42dda6a00e1939a6748519074d96a6faa7728` (1370) | validated recovery receipt for `4124…` |
| `OUT/journal/082476b0335b4deebe7390ca6b426c0a.start.json` | `3a13295dd6e4579ccd1ec5e86377227b7a9cf0bfe51406e4a6435e491885bd16` (20718) | refused operation, start |
| `OUT/journal/082476b0335b4deebe7390ca6b426c0a.json` | `808dfd0d762e64ac21bb32fd30b14666c29a97932376704c4f3544fbc74d1933` (21596) | <ul><li>`status: refused`, `exit_code: 1`</li><li>`source_pins` include `closure/native.py` `c264d61a…`</li><li>`producer_pins` = `attest_closure.py`, `preparation.py`, `verify_inputs.py`</li></ul> |
| `SCRATCH/closure-preflight/082476b0335b4deebe7390ca6b426c0a.json` | `d62fbe0483903f0742eb1c5dff83fb71d50b4b85a2d8b8d38bb9283bdff9c20f` (69649) | refused receipt (the sole `output_pins` row); pins all 272 command records |
| `C/000267.json` | `1d088008b0a0a0903b458dabe9de5937d5f5ffb0c88bee4640a3ad227f3ba06a` (3895) | CoreAudio `dyld_info`: `delay-init weak-link    /System/Library/PrivateFrameworks/RemoteProcessingBlock.framework/…`; present weak links TCC, AppServerSupport, perfdata, AudioDSPGraph, WatchdogClient, libAudioStatistics, SystemPolicy |
| `C/000271.json` | `9d48746ed4f1b0a7d441a8d449495ba061916f8e512001c69da821f2b97c61a6` (3084) | AudioDSPGraph `dyld_info`: `weak-link      …RemoteProcessingBlock`; present weak Swift overlays |
| `C/000272.json` | `4cc9758dd0c31bbbc01c8cbdb5c2f6369404081907c7f6a92a639925ec71b034` (1827) | failing command: `exit_code 0`, `stdout ""`, stderr `non-mach-o in fat file '…'` / `dyld_info: '…' file not found` |
| `tests/test_closure_review.py:192–193`, `tests/test_closure_round2.py:178–181,309–382` | not hashed | the signed five-file `own_sources` assertion; the pattern for patched-orchestration tests |

## Findings

**F1. Cause: an absent weak link treated as required.**
- `native.py:149–159` takes the load path from every `-linked_dylibs` line and drops the attribute column. `:163–167` then requires every target.
- RemoteProcessingBlock is linked only weakly: by CoreAudio as `delay-init weak-link` (`C/000267.json`) and by AudioDSPGraph as `weak-link` (`C/000271.json`).
- It is missing from both the cache and the disk. `dyld_info` printed nothing on stdout, gave "file not found" on stderr and exited 0 (`C/000272.json`), so the identity check at `:145–148` refused.
- dyld lets an effectively weak link be missing (`JustInTimeLoader.cpp:475–518`, `canBeMissing`). The closure claim was therefore not violated. The parser misread a legitimate graph.

**F2. No sound route without code changes (B).**
- The CLI (`attest_closure.py:812–822`) has no switch that excludes a dependency or makes it optional.
- `--dyld-cache-root` only selects which cache files to pin against the running UUID (`native.py:103–129`). `dyld_info` is called with the image name alone (`:145`).
- The seeds (`.tools`, Python, Git, GOROOT, OS tools) cannot be dropped (`attest_closure.py:595–604,708–715`). Every runtime binary that links CoreFoundation reaches CoreAudio (peer reading of `C/000014`–`000272`).
- Supplying the framework, swapping the cache or changing the OS breaks P3.

**F3. Editing `native.py` in place is fatal at seal. Confirmed.**
- The end record of `0824…` pins `closure/native.py` `c264d61a…` in `source_pins`, through `own_sources()` (`attest_closure.py:469–471,838`).
- It is not in `producer_pins`, because `native.py` is imported lazily inside `native_closure` (`:465`), after the Journal has snapshotted the loaded modules (`preparation.py:367`).
- `collect_journals` runs `verify_tree_pins` on every terminal record (`:618`) before the recovery branch (`:621–630`). Any byte change at that path therefore makes the seal refuse permanently, unless the original bytes are restored.
- The same applies to `attest_closure.py`, `preparation.py`, `verify_inputs.py`, `run-config.json`, `preparation/pins.json`, `OUT/reviews/closure-model-free.json` and `binding-closure-confirm-round-2.md`. All are pinned by `0824…`.

**F4. No consumer of `closure.json` needs a code change.** Round 1 assumed otherwise; this corrects it.
- `verify_installed_closure` (`attest_closure.py:474–494`) checks only:
  - the schema/status/complete flags;
  - re-enumeration;
  - `validate_pin` on `inputs`, `source_pins`, `original_runtime_files`, `submodule_working_files`, `deferred_working_files_verified` and `retained_files`;
  - retained-file modes;
  - entrypoint resolutions.

  It never reads `native_dependencies` and never reruns `dyld_info`. Binding acceptance calls it at every launch (`acceptance.py:81–82`) and runs `verify_tree_pins(closure)` (`:350`). The seal calls it (`preparation.py:702–703`) and runs `verify_tree_pins(value)` with schema/status/clock checks (`:663–675`).
- The only binding constraint: output must keep the top-level schema `scout.x1.installed-closure.v1` and every existing top-level field. `run-config.json:79–80` requires that kind, and `run-config.json` cannot change (F3).
- `reviewed_go_sources` (`attest_closure.py:554–563`) accepts only `model_free/*.go`. It and its receipt are untouched.
- New files under `execution/` are allowed. The seal requires the future `execution-source-manifest.json` to list every file under EXEC (`preparation.py:709–713`), and binding requires the implementation review to bind that manifest (`acceptance.py:340–346`). Neither exists yet, so both will include the new files.

**F5. Absence needs its own test.** `dyld_info`'s exit status and stderr do not establish absence: it exits 0 on "file not found" (`C/000272.json`), and `Commands.run` returns stdout only. The process-local dyld API `_dyld_shared_cache_contains_path` does answer membership in the active cache. The signed code already uses the same `ctypes.CDLL(None)` route for `_dyld_get_shared_cache_uuid` (`native.py:105–109`), and CPython uses this API for `ctypes.util`.

**F6. Absence holds at every launch without a new check.**
- The whole active cache family is in `original_runtime_files` (`attest_closure.py:720–721`), and `validate_pin` re-checks it at every launch and at seal (`:482–484`). Absence from the cache is a property of those bytes.
- Absence from `/System/Library` or `/usr/lib` on disk is protected by the sealed system volume. A file could appear only through an OS or cryptex change, which also replaces the pinned cache files [INFERENCE].
- So no launch-time absence check is added, and binding stays unchanged.

**F7. Latent next-refusal risks in the signed walk, none yet exercised on real bytes.** Only node's `otool` calls (`C/000012`–`000013`) and 259 `dyld_info` calls ran.
- **(a) Recursion depth.** `cached_image` recurses. The trace shows 49+ frames at the failure. The pinned interpreter is CPython 3.11.11 [INFERENCE from the external root name], whose default recursion limit is 1000, and a depth-first walk over the cache can exceed it. `RecursionError` would be one more refusal.
- **(b) Truncated paths.** The unanchored regex at `:155` would cut `/System/iOSSupport/System/Library/…` down to `/System/Library/…`.
- **(c) `-rpaths:` section.** Upstream `dyld_info` can append a `-rpaths:` section. `:151–159` would read its lines as dependencies.
- **(d) On-disk weak links.** The on-disk branch (`:189–205`) also ignores weak links.
- **(e) Aliases.** The identity check `text.startswith(name+' [')` refuses an alias spelling.

**F8. Avoidable cost at seal and binding.** Every cached-image record embeds the full `cache_files` pin list (`native.py:160–162`). `verify_tree_pins` at seal (`preparation.py:673`) and at binding (`acceptance.py:350`) therefore re-hashes the multi-GB cache family once per cached image. With about 10³ images this is plausibly hours per call [INFERENCE]. It is not a refusal, but it costs clock time. The family is already pinned once at `native_dependencies.cache_files` and in `original_runtime_files`.

**F9. Recovery path.** The Core operation-recovery path (`preparation.py:621–630`) covers `0824…`, exactly as it covered `4124…`, provided every pin in its end record stays valid until seal (F3). A′ at new paths edits none of them.

## Recommendation (decision)

**Adopt A′, append-only:**
- Add a new native collector, `execution/closure/native_v2.py`.
- Add a thin, digest-guarded entrypoint, `execution/attest_closure_v2.py`, that runs the unchanged signed `attest_closure` with exactly two module globals rebound: `native_closure` and `own_sources`.
- Add a new test module, `execution/tests/test_closure_weak_links.py`.
- Edit no existing file, and change no consumer code.

**Reasons.**
- **Correctness.** The new collector follows dyld's rule:
  - only an effectively weak edge may be missing;
  - absence is proved by three independent conditions: not a cache member, no filesystem entry including cryptex overlays, and `dyld_info` prints nothing;
  - every required edge stays strict;
  - a target that is present but cannot be inspected still refuses.

  The output records each absent weak link and its evidence, so `complete: true` keeps its meaning: every image dyld can load is closed.
- **Blast radius.**
  - No signed byte changes, and every pin of `4124…` and `0824…` stays valid.
  - Binding acceptance and the seal keep calling the unchanged `attest_closure.verify_installed_closure` (F4).
  - Schema `scout.x1.installed-closure.v1` is kept.
  - Absent-link rows carry no `path`/`sha256` pair, so `verify_tree_pins` never tries to pin a file that does not exist.
- **Clock.**
  - One implementation round is expected (more only if issues are found). The binding/seal consumer confirmation is folded into that round instead of a separate binding-confirm round, because no consumer code changes (F4).
  - The read-only M-checks run before implementation, so findings shape implementation round 1 and do not cost a refused journal.
  - Deadline 2026-12-24T03:22:58Z, about 44 days away.
- **Wrapper rather than copy.**
  - Copying `attest()` and `main()` (~250 lines) would duplicate signed logic and need a full re-review.
  - The call sites look up `native_closure` (`:603`) and `own_sources` (`:622`, `:838`) through module globals at call time. Rebinding two names is a narrow seam that is easy to audit, and the signed tests already patch the same seam (`test_closure_round2.py:374,379`).
  - Round 1's concern about monkeypatching is met in three ways:
    - rebinding happens only in `main()`, never at import;
    - the wrapper refuses unless the base files have their signed digests;
    - tests assert the exact rebinding and its absence after import.
- **Collector rather than intercepting commands.** Filtering `dyld_info` output to hide edges would split the evidence into two representations. A new collector that reuses the signed `image_filetype` and `resolve_install_name` is easier to audit.

## Alternatives rejected

| Option | Why rejected |
|---|---|
| **B: any route without code changes** | None exists (F2). |
| **A: edit `native.py` or `attest_closure.py` in place** | Fatal at seal (F3). |
| **A′ with a full copy `attest_closure_v2.py` (no rebinding)** | Duplicates ~250 signed lines, so the review surface and drift risk are larger; no correctness gain. |
| **Treat every weak link as optional** (never walk weak targets) | Under-closes: present weak targets (TCC, WatchdogClient, Swift overlays) do load and must be closed. |
| **Treat any `weak-link` token as optional** | dyld makes `re-export` + `weak-link` required (`JustInTimeLoader.cpp:475–518`). |
| **Use `dyld_info` stderr or exit status as the absence test** | Not a test: exit 0 on "file not found" (F5). |
| **Pass absent weak links only when called from cached images** | On-disk images can weak-link system images too (F7(d)). Applying the same proved-absence rule to both branches, limited to `/usr/lib/` and `/System/Library/` targets, avoids another refusal and is no less strict. |
| **Add a launch-time absence recheck to binding or seal** | Needs changes to pinned or signed consumers. The pinned cache bytes and the sealed system volume already bind absence (F6). |
| **Allow absent weak links outside `/usr/lib/` and `/System/Library/`** (`@rpath`, Homebrew) | dyld's fallback search for these is not modelled. They stay refused, which fails safe. |

## Implementation specification

### I1. Files

New files only:
- `execution/closure/native_v2.py`
- `execution/attest_closure_v2.py`
- `execution/tests/test_closure_weak_links.py`

No existing file under `execution/`, `preparation/` or `preparation-reviews/` is edited. In particular, `tests/test_closure_review.py:192–193` (the signed five-file `own_sources` assertion) stays as it is and must keep passing.

### I2. `execution/closure/native_v2.py`

- **Imports.** It imports `image_filetype` and `resolve_install_name` from `closure.native` at module top. It does not copy them.
- **Constants:**
  - `FORMAT = 'scout.x1.native-dependencies.v2'`
  - `SYSTEM_PREFIXES = ('/usr/lib/', '/System/Library/')`
  - `LINK_ATTRIBUTES = {'upward', 'delay-init', 'weak-link', 're-export', 'lazy-load'}`
  - `LOAD_COMMANDS = ('LC_LOAD_DYLIB', 'LC_LOAD_WEAK_DYLIB', 'LC_REEXPORT_DYLIB', 'LC_LOAD_UPWARD_DYLIB', 'LC_LAZY_LOAD_DYLIB')`
  - `OVERLAY_PARENTS = ('/System/Cryptexes', '/System/Volumes/Preboot/Cryptexes')`
- **Module-level helpers.** Tests patch these.
  - `shared_cache_contains(name) -> bool`
    - Uses `ctypes.CDLL(None)._dyld_shared_cache_contains_path` with `argtypes=[c_char_p]` and `restype=c_bool`.
    - If the symbol is missing, raise `ValueError('cannot query active dyld shared-cache membership')`.
  - `overlay_roots() -> list[Path]`
    - Returns the sorted child directories of each existing `OVERLAY_PARENTS` entry, via `os.scandir`.
    - Any `OSError` other than `FileNotFoundError` on a parent raises `ValueError('cannot establish native absence: '+path)`.
  - `filesystem_entries(name) -> list[str]`
    - Checks `name` and `str(root)+name` for each overlay root.
    - Returns those for which `os.lstat` succeeds.
    - `FileNotFoundError` and `NotADirectoryError` count as absent. Any other `OSError` raises `ValueError('cannot establish native absence: '+path)`.
- **`parse_linked_dylibs(name, text) -> (image_uuid, edges)`.** `edges` is an ordered list of `(target, attributes_tuple)`. Rules:
  1. **Identity, kept from v1.**
     - `text.startswith(name+' [')` is required.
     - Exactly one line matches the v1 UUID regex (`native.py:146`).
     - Exactly one line has the image header form `^\S.* \[[^\]]+\]:$`.
     - Otherwise raise `ValueError('ambiguous cached native image identity: '+name)`.
  2. **Sections.**
     - Section headers are lines matching `^    -[A-Za-z_]+:$`.
     - Allowed headers are `-uuid:`, `-linked_dylibs:` and `-rpaths:`; each may appear at most once, and `-linked_dylibs:` exactly once.
     - Any other header, or a duplicate, raises `ValueError('unexpected dyld_info section: '+line)`.
     - The `-linked_dylibs` section ends at the next section header or at EOF. Lines in `-rpaths:` are never dependencies.
  3. **Column header.** The first non-blank line of `-linked_dylibs` must satisfy `line.split() == ['attributes', 'load', 'path']`. Otherwise raise `ValueError('unexpected dyld_info link header: '+line)`. Substring matching of `'load path'` is not used.
  4. **Link lines.** Every later non-blank line must fully match `^\s+((?:[a-z][a-z-]* +)*)(/.*\S)$`. The anchored absolute path is kept in full, with no truncation.
     - A line that does not match raises `ValueError('unresolved cached-image load command: '+line)` (v1 message).
     - The attribute tokens are `group1.split()`. Any token outside `LINK_ATTRIBUTES`, or a repeated token, raises `ValueError('unknown cached-image link attribute: '+line)`.
- **`effective_weak(attributes)`** is `'weak-link' in attributes and 're-export' not in attributes`. If the same target occurs more than once in one image, the edge is weak only if every occurrence is effectively weak.
- **`collect(files, commands, cwd, cache_root, safe_resolve, pin)`.** Same signature and same seeding, `pending` and `contextual_sources` logic as `native.collect`. The cache-identification block is copied exactly from `native.py:103–129`. The differences are:
  - **(a) Iterative cached walk.**
    - Replace the recursive `cached_image` with an explicit stack of frames `(record, iterator over remaining edges)`.
    - Visit order and `pending.append` order must match v1's recursion exactly.
    - Insert each record into `cached` before its edges are processed, as v1 does.
    - No recursion is used anywhere.
  - **(b) Resolving an edge `(source, target, weak, provenance)`.** Shared by the cached and on-disk branches.
    1. If `Path(target).is_file()`: `pending.append((Path(target), None))` (v1).
    2. Else if `not cache_files`: raise `ValueError('unresolvable native runtime dependency: '+target)` (v1 message).
    3. Else if `shared_cache_contains(target)`: treat as a cached image. Run `dyld_info -uuid -linked_dylibs target` and parse it with `parse_linked_dylibs`. Any parse or identity failure refuses; it is never converted into absence.
    4. Else if `weak` and `target.startswith(SYSTEM_PREFIXES)`, prove absence. All three must hold:
       - `filesystem_entries(target) == []`;
       - `commands.run(['/usr/bin/dyld_info', '-uuid', '-linked_dylibs', target], cwd)` returns stdout `''`, so the command record is kept;
       - `shared_cache_contains(target)` is False (step 3).

       If all hold, record the edge in `absent[target]` and do not traverse it. The probe runs once per target, and the result is memoised. If any condition fails, raise `ValueError('weak native dependency not provably absent: '+source+' -> '+target)`.
    5. Otherwise raise `ValueError('required native dependency absent: '+source+' -> '+target)`.

    Absence is judged per edge. A required edge to a target that is already recorded as weak-absent refuses at step 5, whatever the visit order.
  - **(c) Cached-image records:**
    - `install_name`, `image_uuid`, `active_cache_uuid`;
    - `dependency_install_names`: all targets, in order (v1 field kept);
    - `linked_dylibs`: `[{install_name, attributes: [...], requirement: 'weak'|'required'}]`.

    The per-image `cache_files` field is removed (F8). The family is recorded once at the top level.
  - **(d) On-disk branch.**
    - Keep v1's `otool -L`/`-l` calls, rpath parsing, `LC_ID_DYLIB` exclusion and `resolve_install_name`.
    - Additionally parse `details` with `cmd (LC_[A-Z_]+)\s+cmdsize \d+\s+name (.+?) \(offset` into a map from name to the set of `LOAD_COMMANDS` that name it.
    - For each dependency:
      - **Exactly one candidate exists:** as v1.
      - **Several candidates exist:** refuse, as v1.
      - **No candidate exists and the dependency is absolute:** resolve through (b). `weak` is true only if the map entry equals exactly `{'LC_LOAD_WEAK_DYLIB'}`; a missing entry means required.
      - **No candidate exists and the dependency is not absolute:** v1's `'ambiguous/missing native dependency: '` refusal.
    - Dependency rows:
      - present on disk: as v1, plus `load_commands`;
      - cached: `{install_name, active_cache_uuid, load_commands}`;
      - weak-absent: `{install_name, active_cache_uuid, load_commands, resolution: 'weak-absent'}`.
  - **(e) Return value:**
    - `format`
    - `images`
    - `cached_images`, sorted by `install_name`
    - `active_cache_uuid`
    - `cache_files`
    - `absent_weak_links`: sorted by `install_name`. Each row has:
      - `install_name`
      - `active_cache_uuid`
      - `linked_from`: sorted by `source`; each entry has either `attributes` (cached parent) or `load_commands` (on-disk parent)
      - `evidence`: `{shared_cache_contains_path: false, filesystem_entries: [], probed_paths: [...], dyld_info_stdout_empty: true}`

    No absent row, at any depth, contains both `path` and `sha256`.

### I3. `execution/attest_closure_v2.py`

- **Imports.** In order:
  1. `HERE = Path(__file__).resolve().parent`
  2. `sys.path.insert(0, str(HERE))`, needed because `-I` drops the script directory
  3. `import attest_closure as base`
  4. `from closure import native_v2`, at top level so that `producer_pins` captures `native_v2.py` and `native.py` when the Journal starts
  5. `from verify_inputs import pin`
- **`SIGNED_BASE`:**
  - `'attest_closure.py': '113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892'`
  - `'closure/native.py': 'c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb'`
- **`_SIGNED_OWN_SOURCES = base.own_sources`**, captured at import.
- **`own_sources()`** returns `[*_SIGNED_OWN_SOURCES(), pin(HERE/'attest_closure_v2.py'), pin(HERE/'closure/native_v2.py')]`: seven rows, in that order.
- **`native_closure(files, commands, cwd, cache_root=None)`** returns `native_v2.collect(files, commands, cwd, cache_root, base.safe_resolve, base.pin)`.
- **`install()`:**
  - Checks `base.digest(HERE/name) == sha` for every `SIGNED_BASE` entry, else raises `ValueError('signed closure base drift: '+name)`.
  - Then sets `base.native_closure = native_closure` and `base.own_sources = own_sources`.
- **`main()`** calls `install()`, then `return base.main()`.
- **Entry point:** `if __name__ == '__main__': main()`.
- Importing the module mutates nothing.

### I4. Output contract

The canonical and preflight outputs keep:
- `schema: scout.x1.installed-closure.v1`;
- every existing top-level field and the semantics of `attest()`.

Only `native_dependencies` changes, as in I2(e), and `source_pins`, which gains the two v2 rows through `own_sources`. The journal `source_pins` gain the same two rows (`:838`).

### I5. Tests: `execution/tests/test_closure_weak_links.py`

All tests run with the pinned `$PY -I -B`.

**Real-byte tests.**
- They run only when all of these hold:
  - `platform.system() == 'Darwin'`;
  - `platform.machine() == 'arm64'`;
  - `/usr/bin/sw_vers -buildVersion` is `25G72`.

  Otherwise they call `skipTest` with the reason. The implementation response must show that they ran, not skipped, on the frozen host.
- Each uses a `Commands` instance with a closed environment (`PATH=/usr/bin:/bin`, `LANG=C`, `LC_ALL=C`) recording to a temp directory, and the real cache root `/System/Volumes/Preboot/Cryptexes/OS/System/Library/dyld`.

**Synthetic tests** use fake `dyld_info`/`otool` text in the observed format (`C/000267.json`) and patch `shared_cache_contains`, `overlay_roots` and `filesystem_entries` with `unittest.mock.patch.object`.

| ID | Kind | Test | Expected |
|---|---|---|---|
| R1 | real, planted effect | Cached walk seeded with a weak-free required edge to `/System/Library/Frameworks/CoreAudio.framework/Versions/A/CoreAudio`, through a v2 entry that accepts seed edges (writer's choice of name, reviewed) | <ul><li>Completes.</li><li>`absent_weak_links` contains RemoteProcessingBlock, with `linked_from` ⊇ {CoreAudio `['delay-init','weak-link']`, AudioDSPGraph `['weak-link']`}.</li><li>TCC, WatchdogClient and AudioDSPGraph are in `cached_images` with UUIDs.</li></ul> |
| R2 | real, negative | A required seed edge to RemoteProcessingBlock | Raises `required native dependency absent` |
| R3 | real, planted null | `native.collect` and `native_v2.collect` on `['/bin/sh']` with the real cache root | Equal, except for the documented differences: no per-image `cache_files`; added `format`, `linked_dylibs`, `load_commands`, `absent_weak_links` (empty). `images` order is identical. If v1 refuses on `/bin/sh`, that is a finding for the implementation round, not a reason to change the seed silently. |
| S1 | planted effect | Weak edge to an absent system target | Recorded, with evidence |
| S2 | negative | Required edge to an absent target | Refuses |
| S3 | negative | `delay-init` alone, target absent | Refuses |
| S4 | negative | `re-export weak-link`, target absent | Refuses |
| S5 | negative | Weak edge, membership true, but `dyld_info` output fails to parse or has no unique UUID | Refuses (never absent) |
| S6 | negative | Weak edge, membership false, `filesystem_entries` non-empty (on disk; overlay) | `not provably absent` |
| S7 | negative | Weak edge, membership false, no entries, `dyld_info` stdout non-empty | `not provably absent` |
| S8 | negative | Same target weak from A and required from B, both visit orders | Refuses |
| S9 | parser | <ul><li>unknown token</li><li>repeated token</li><li>duplicate `-linked_dylibs:`</li><li>unknown section</li><li>wrong column header</li><li>`-rpaths:` lines</li><li>`/System/iOSSupport/System/Library/…` path</li><li>two arch headers</li></ul> | <ul><li>refuse</li><li>refuse</li><li>refuse</li><li>refuse</li><li>refuse</li><li>`-rpaths:` lines are not dependencies</li><li>the iOSSupport path is kept whole</li><li>two arch headers refuse</li></ul> |
| S10 | on-disk | <ul><li>`LC_LOAD_WEAK_DYLIB` to an absent system target</li><li>`LC_LOAD_DYLIB` to the same</li><li>weak link to an absent `@rpath` or non-system absolute target</li><li>name named by both weak and non-weak commands</li></ul> | <ul><li>recorded</li><li>refuses</li><li>refuses</li><li>required</li></ul> |
| S11 | depth | Linear cached chain of 5000 images | Completes; no `RecursionError` |
| S12 | null | Synthetic graphs without weak-absent edges, including executable-context cases from `test_closure_review.py:59–78` | v1 and v2 equal modulo the documented differences |
| S13 | consumers | `preparation.verify_tree_pins` over a v2 result containing absent rows | Returns only real file pins and does not raise. The signed `attest_closure.verify_installed_closure` passes on a fixture closure (pattern `test_closure_review.recheck_fixture`) whose `native_dependencies` is in v2 format. |
| S14 | negative | `cache_root=None` with a weak absent system target | Refuses (`unresolvable native runtime dependency`) |
| W1 | wrapper | `install()` with a drifted base digest (patch `base.digest`) | Refuses before rebinding |
| W2 | wrapper | `install()` inside `patch.object` contexts that restore the base | `base.native_closure is v2.native_closure` and `base.own_sources is v2.own_sources`; `base.inventory_native_closure` routes into `native_v2.collect` (mocked) |
| W3 | wrapper | `own_sources()` | Exactly the 5 signed rows plus the 2 v2 rows, in order |
| W4 | wrapper | `v2.main()` with patched `attest`/`Journal`/`load_config`/`read_clock`/`operation_lock` (pattern `test_closure_round2.py:178–181`) | Journal `source_pins` contain both v2 files; the argv path is the v2 file |
| W5 | wrapper | After `import attest_closure_v2` | `attest_closure.native_closure` and `own_sources` are the originals; `test_closure_review` still passes in the same process |

The full existing execution suite must pass unchanged.

### I6. Consumers

There are no code or config changes to:
- `binding/acceptance.py`;
- `preparation.py` (seal);
- `attest_closure.verify_installed_closure`;
- `reviewed_go_sources`;
- `preparation/pins.json`;
- `run-config.json`;
- `OUT/reviews/closure-model-free.json`.

The implementation round confirms this against I4 and S13. The later execution-source manifest and `reviews/implementation.json` must include the three new files (F4).

### I7. Amendments from the M-checks

If M4 or M5 (below) shows a concrete refusal that this specification would still hit on a reachable edge, the writer may add the smallest change that removes it. Examples are an alias spelling, an unlisted attribute token or an on-disk `@rpath` chain. Each such change must be listed in the implementation response with its M-check evidence and a test, and implementation round 1 reviews it. Nothing else departs from I1–I6.

## Execution outline (review sequence)

Stop at the first failure and open a new council round. Write nothing under `OUT` except as stated.

**Step 0. Recovery.**
- Main writes this record verbatim (UTF-8, LF) and measures its SHA-256 and byte size.
- Main writes `OUT/reviews/operation-recovery-082476b0335b4deebe7390ca6b426c0a.json` (next section).
- Main runs `review_receipt(<receipt>, 'operation-recovery')` read-only, as round 1 did. It must pass before any later journal.

**Step 1. M-checks.** Unjournalled and read-only; no lock and no Journal. Outputs go under `SCRATCH/closure-remedy-checks/r2-m*/`, recorded in `preparation-reviews/closure-weaklink-prescan-2026-11-10.json` with argv, exit codes, stdout/stderr digests and timestamps.

| Check | What | Stop condition |
|---|---|---|
| **M1** (H1) | `$PY -I -B` + ctypes `_dyld_shared_cache_contains_path` on RemoteProcessingBlock, CoreAudio, AudioDSPGraph and TCC | Stop if RemoteProcessingBlock → True, or any of the other three → False |
| **M2** (calibration) | `/usr/bin/dyld_info -uuid -linked_dylibs /System/Library/PrivateFrameworks/X1ProbeAbsent.framework/Versions/A/X1ProbeAbsent` | Informational: record that this gives the same stdout/stderr shape as `C/000272.json`, with exit 0 |
| **M3** (H4) | `os.lstat` of the RemoteProcessingBlock binary and its `.framework` directory, at `/` and under every child of `/System/Cryptexes` and `/System/Volumes/Preboot/Cryptexes`. Record `AudioDSPGraph.framework` as a positive control. | Stop if any RemoteProcessingBlock entry exists or any listing raises |
| **M4** (census) | `dyld_info -help`. If `-all_dyld_cache` exists, run `/usr/bin/dyld_info -all_dyld_cache -uuid -linked_dylibs` to `r2-m4/all-cache.txt`. Report: header install names; edges with attributes; targets not among the headers, with membership and `lexists`; unknown tokens or sections; load paths outside `SYSTEM_PREFIXES`; alias spellings (member but not a header name). | Stop if any required-absent edge is reachable from the cached seeds found in M5. Reachable aliases, tokens or prefixes feed I7. |
| **M5** (on-disk census) | Harness: call `attest_closure.attest()` with a stub journal (id `remedy-r2-m5`, a SCRATCH-only start file) and patch `attest_closure.native_closure` to save the exact `files` list and raise a sentinel. Then, for every Mach-O in that list, run `otool -L`/`-l` and report: unresolvable `@rpath`/`@executable_path`/`@loader_path` dependencies under v1 rules; absolute dependencies that are neither regular files nor cache members, with their load commands; fat images; cached seed names. The harness script and its digest go in the prescan record. | Findings feed I7. Stop only if a required system dependency is absent. |
| **M6** | `df -k SCRATCH`; total bytes of the active cache family plus the M5 file list | Stop if free space < 2.2 × that total (preflight and canonical retention) |
| **M7** | `defaults read /Library/Preferences/com.apple.SoftwareUpdate` | If automatic installation of security responses or critical updates is enabled, Main turns it off before Step 5 and records the before/after values. This protects P3 and changes no OS bytes. |

**Step 2. Write the code.** Main writes I1–I5 plus any I7 amendments. Main runs the full execution suite and the real-byte tests on the host, then writes `preparation-reviews/closure-weaklink-impl-round-1-response.md`, giving:
- the digests of the three files;
- the suite output;
- the M-check record;
- each I7 amendment with its evidence.

**Step 3. Implementation round(s).** The council reviews in `preparation-reviews/closure-weaklink-impl-round-<N>.md` and iterates to zero issues. The round also contains the binding/seal consumer confirmation (I6). Its sign-off names:
- the three new digests;
- the unchanged signed digests;
- P3/P4 restated with the v2 argv.

**Step 4. GR native rehearsal.** Unjournalled.
- Harness:
  - `import attest_closure_v2 as v2`;
  - call `v2.install()` inside a context that restores the base globals;
  - call `base.attest(args, config, stub)`, where `args` holds exactly the parsed values of the Step 5 argv and `stub` has id `remedy-r2-gr`;
  - patch `base.retain` to raise a sentinel after saving the `native` result to `SCRATCH/closure-remedy-checks/r2-gr/native.json`.
- Pass conditions:
  - the sentinel is raised and nothing else;
  - `absent_weak_links` contains RemoteProcessingBlock;
  - every absent row's evidence meets I2(b)4.
- Record the argv, digests and timing in the gate record (Step 8).

**Step 5. Journalled preflight.** Use the exact argv of `0824…` with one substitution: the script path `…/execution/attest_closure.py` becomes `…/execution/attest_closure_v2.py`. Everything else is identical, including `--module-cache …/closure-modules-r2`, the 32 `--external-root` values and `--dyld-cache-root`. If it is refused, stop; the refusal needs its own recovery record.

**Step 6. GV verify rehearsal.** Read-only, after a `passed` preflight.
- Run `attest_closure.verify_installed_closure(<preflight output>)` and `preparation.verify_tree_pins(json.load(<preflight output>))`, and record wall time.
- Both must pass. Times over 1 h go into the gate record as clock risk.

**Step 7. Canonical closure.** Once, with the Step 5 argv minus `--preflight` plus `--output $OUT/closure.json` (P4, S16). Then follow the binding runbook (RB-1).

**Step 8. Gate record.** `preparation-reviews/closure-weaklink-gates-<date>.json`, never under `OUT`. It holds:
- Step 0, GR, Step 5, GV and Step 7 argv, exit codes, digests and timestamps;
- the M-check record digest;
- the journal IDs.

**Step 9. RESULTS.md, Deviations.** Add entries for:
1. `4124…` and B′ (round 1);
2. `0824…`: cause F1, recovered under this record;
3. the v2 entrypoint and collector: why, the absent weak links recorded, and the argv substitution;
4. the M-checks, GR and GV;
5. reproduce commands.

**Step 10. Ledgers.** Main adds research-ledger entries for:
- Apple dyld `dyld_info.cpp` (link attributes);
- `JustInTimeLoader.cpp` (`canBeMissing`, the `re-export` normalisation);
- `_dyld_shared_cache_contains_path` (bpo msg400797).

This is required because these sources justify the decision.

**P3, restated.** In addition to round 1's list, keep these byte-stable until seal:
- `SCRATCH/closure/082476b0335b4deebe7390ca6b426c0a/`;
- `SCRATCH/closure-preflight/082476b0335b4deebe7390ca6b426c0a.json`;
- `SCRATCH/closure-remedy-checks/r2-*`;
- `SCRATCH/closure/remedy-r2-*`.

No OS, cryptex or cache change. The rest of P3 and P4 is unchanged.

## Recovery handling for `082476b0335b4deebe7390ca6b426c0a`

- **Path.** The existing Core path covers this refusal (`preparation.py:621–630`). At seal it requires `OUT/reviews/operation-recovery-082476b0335b4deebe7390ca6b426c0a.json`, validated by `review_receipt(…, 'operation-recovery')`. It requires that the receipt's reviewed digests include the start record, the end record and the end record's `output_pins` row. It lists all four items under `refusals_or_deviations`.
- **Precondition** (`:618`). Every pin in the `0824…` end record stays valid until seal:
  - the 8 `source_pins`: `attest_closure.py`, `closure/native.py`, `preparation.py`, `verify_inputs.py`, `run-config.json`, `preparation/pins.json`, `OUT/reviews/closure-model-free.json` `baddb17350575bb118e734606ad846897d6d376c7ba4e40d964d57405d3fe804`, and `binding-closure-confirm-round-2.md` `c2bf06cff7830cc085b4a5f0496b4835394047332019ddfa3a76d0e7278ff944`;
  - the 3 `producer_pins`;
  - the clock and the configuration;
  - the `prior_journals`;
  - the SCRATCH preflight receipt.

  A′ (new files only) preserves all of them.
- **Authority.** This record is the authority: its single operation-recovery scope block is at the end.
- **Receipt.** Main writes it with exactly these keys; `bytes` are integers measured after writing:

```json
{
  "scope": "operation-recovery",
  "record": {"path": "/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-reviews/closure-preflight-remedy-round-2.md", "sha256": "<sha256 of this file as written>", "bytes": "<size as integer>"},
  "reviewed": [
    {"path": "/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/journal/082476b0335b4deebe7390ca6b426c0a.start.json", "sha256": "3a13295dd6e4579ccd1ec5e86377227b7a9cf0bfe51406e4a6435e491885bd16", "bytes": 20718},
    {"path": "/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/journal/082476b0335b4deebe7390ca6b426c0a.json", "sha256": "808dfd0d762e64ac21bb32fd30b14666c29a97932376704c4f3544fbc74d1933", "bytes": 21596},
    {"path": "/Users/adamjackson/Projects/work/x1-preparation-scratch/closure-preflight/082476b0335b4deebe7390ca6b426c0a.json", "sha256": "d62fbe0483903f0742eb1c5dff83fb71d50b4b85a2d8b8d38bb9283bdff9c20f", "bytes": 69649},
    {"path": "/Users/adamjackson/Projects/work/x1-preparation-scratch/closure/082476b0335b4deebe7390ca6b426c0a/commands/000267.json", "sha256": "1d088008b0a0a0903b458dabe9de5937d5f5ffb0c88bee4640a3ad227f3ba06a", "bytes": 3895},
    {"path": "/Users/adamjackson/Projects/work/x1-preparation-scratch/closure/082476b0335b4deebe7390ca6b426c0a/commands/000271.json", "sha256": "9d48746ed4f1b0a7d441a8d449495ba061916f8e512001c69da821f2b97c61a6", "bytes": 3084},
    {"path": "/Users/adamjackson/Projects/work/x1-preparation-scratch/closure/082476b0335b4deebe7390ca6b426c0a/commands/000272.json", "sha256": "4cc9758dd0c31bbbc01c8cbdb5c2f6369404081907c7f6a92a639925ec71b034", "bytes": 1827}
  ]
}
```

- **Rows.** The first three rows are what Core requires. Rows 4–6 bind the causal evidence (the two weak edges and the failing probe), following round 1's precedent. Each row becomes a seal-time `validate_pin` dependency, which P3 covers.
- **Disposition.** `0824…` is a retained refusal, not action-(b) evidence. `closure.json` still needs its own passed producer (`preparation.py:728–731`). Nothing about `0824…` is deleted, rewritten or relabelled. Recovery does not authorise resuming: Step 5 waits for Step 3's sign-off.

## Risks

- **A later stage refuses.** These stages have never run on real bytes: the on-disk walk beyond node, `retain()`, re-enumeration and the output write. M5 and GR cover the walk before any journal. `retain()`, re-enumeration and the output write can still refuse at Step 5, which costs one recovery round, not `closure.json` (P4).
- **The absence test misses a dyld search location** [INFERENCE]. The test requires three independent conditions and is limited to `/usr/lib/` and `/System/Library/` on the sealed volume. Anything outside that refuses. M3 and R1/R2 check behaviour on the installed host.
- **Alias spellings or new `dyld_info` grammar** cause another refusal. M4 feeds I7, and the parser fails closed.
- **A background security update replaces the cryptex cache.** The pinned cache files then break binding and seal. This risk exists regardless of the remedy and violates P3; M7 mitigates it.
- **Wrapper fragility.** Mitigated by the digest guard, rebinding only in `main()`, and W1–W5. The base files cannot change anyway (F3).
- **Seal and binding cost.** v2 removes the per-image re-hashing (F8). One hash of the cache family per verification remains; GV measures it.
- **Clock.** About 44 days remain. The plan is this round, then the M-checks (hours), implementation (about a day), one to three implementation rounds, GR, preflight and canonical. Each further refusal adds a recovery round.
- **Reversibility.** All code is new files and can be abandoned without touching a signed byte. Journals are create-once. The irreversible step is still the canonical `OUT/closure.json`, gated by a passed preflight and GV.

## How the peers shaped this record

- **Both peers** independently confirmed F2 and F3. They also showed that binding and seal never rerun the native walk, so no consumer code changes. That corrects round 1's assumption that A′ needed updates to the consumers and a separate binding-confirm round.
- **innovation-challenger** added:
  - the dyld source readings, including that `re-export` combined with `weak-link` is required, the full token set (`lazy-load`, the `-rpaths:` section), and that the column padding is a minimum, not a fixed width;
  - exact header matching;
  - "never turn a parse failure into absence";
  - per-edge requiredness;
  - the requirement that the evidence of absence be independent of `dyld_info`;
  - the choice of an explicit collector over intercepting commands.
- **experiment-peer** added:
  - H1–H6 with their falsifying checks, which became M1–M4;
  - the observation that only node's `otool` calls and one cached walk have run (M5);
  - the recursion-depth, truncated-regex and alias risks (F7);
  - the per-image `cache_files` re-hashing cost (F8, removed in I2(c));
  - the requirement that the output keep its schema;
  - GR and GV;
  - the background-update check (M7);
  - the full receipt row set.
- **Mine**:
  - combining the binding confirmation into the implementation round;
  - the wrapper's digest guard and rebinding only in `main()`, which avoids breaking `test_closure_review.py:192–193` in the same test process;
  - the identical-order explicit stack;
  - limiting absence to system prefixes and leaving all other absent weak links refused.

## New issues

None open against this decision or the recovery. The refusal is a defect in the signed parser's reading of weak links (F1). It is remedied by new code at new paths, which still needs its implementation round.

**Zero issues. Signed off.**

This sign-off covers the remedy decision and specification (A′ as I1–I7, with the sequence in Steps 0–10) and the operation-recovery scope below. It signs off no code and authorises no journal, preflight or canonical run; those wait for a zero-issue `closure-weaklink-impl-round-<N>.md`. The signed implementation stays at the digests in `binding-closure-confirm-round-2.md` (`c2bf06cff7830cc085b4a5f0496b4835394047332019ddfa3a76d0e7278ff944`):
- `execution/attest_closure.py` `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892`
- `execution/closure/native.py` `c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb`
- `execution/preparation.py` `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38`
- `execution/run-config.json` `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`

Evidence records reviewed:
- `closure-p4-attempt-3-2026-11-10.json` `9bf2ef06124551e9d1f064b6ca926d605a9661eb99b638414c85cbd768ea8246`
- `closure-preflight-remedy-gates-2026-11-10.json` `bd200ce998f80420bbf402a17f791635b325ef33f933d772b39b20bc3ae927d0`
- `closure-preflight-remedy-round-1.md` `cb9fc32a92dfde04beb6f299297e0dd5d565562d207390fd040ff5e38b8e27c4`

Verdict: REMEDY CHOSEN — implementation review required

Signed-off scope: operation-recovery

Refused closure preflight `082476b0335b4deebe7390ca6b426c0a` (action b): reviewed and recovered as a retained refusal. Cause F1: the signed parser read a weak link to an image absent from the active cache and the disk as required. The remedy is A′ under this record. The reviewed evidence:
- start record `preparation-run/journal/082476b0335b4deebe7390ca6b426c0a.start.json` `3a13295dd6e4579ccd1ec5e86377227b7a9cf0bfe51406e4a6435e491885bd16`
- end record `preparation-run/journal/082476b0335b4deebe7390ca6b426c0a.json` `808dfd0d762e64ac21bb32fd30b14666c29a97932376704c4f3544fbc74d1933`
- refused receipt `SCRATCH/closure-preflight/082476b0335b4deebe7390ca6b426c0a.json` `d62fbe0483903f0742eb1c5dff83fb71d50b4b85a2d8b8d38bb9283bdff9c20f`
- CoreAudio weak-edge record `SCRATCH/closure/082476b0335b4deebe7390ca6b426c0a/commands/000267.json` `1d088008b0a0a0903b458dabe9de5937d5f5ffb0c88bee4640a3ad227f3ba06a`
- AudioDSPGraph weak-edge record `SCRATCH/closure/082476b0335b4deebe7390ca6b426c0a/commands/000271.json` `9d48746ed4f1b0a7d441a8d449495ba061916f8e512001c69da821f2b97c61a6`
- failing probe record `SCRATCH/closure/082476b0335b4deebe7390ca6b426c0a/commands/000272.json` `4cc9758dd0c31bbbc01c8cbdb5c2f6369404081907c7f6a92a639925ec71b034`
