# X1 preparation runtime closure attestation (b) implementation review — round 2

**Scope.** This round re-reviews action (b), the read-only runtime closure attester, against round 1 (`preparation-reviews/closure-impl-round-1.md`) and the author response (`closure-impl-round-1-response.md`). It reads the binding consumer (`binding/acceptance.py`) and the seal (`preparation.py`) only where they use this component's output. The review was read-only: no attester, test, build, model call or hash computation was run, and no real input was processed. The SHA-256s below are Main's `shasum -a 256` values.

## Reviewed files (relative to `evaluation/x1-triage-ranking/execution/`)

| File | SHA-256 |
|---|---|
| `attest_closure.py` | `22f23e4c83b458035a081b18fc79fc5f4d8f82a450300a44333657c2aa7f18ff` |
| `closure/native.py` | `5ae267b380c3a841d4058685593a5fab9ca4e351616207d6b0e5a15f34859d07` |
| `tests/test_closure.py` | `e19d3193908101811dd022711489198fe0f68a9d77ae1f7585205e65ecfaa447` |
| `tests/test_closure_review.py` | `3fa002a1362e754d6528bc296d774f995bc1ac095bf1b28d446a44b53dbe6ea1` |
| `preparation.py` (shared) | `1a116d86b98b088c474a0f4e974af9be8299484ac1deee99dba8f3e7123bc9bf` |
| `verify_inputs.py` (shared) | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` (full value from the response's `shasum` output) |
| `run-config.json` (shared) | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |
| `binding/acceptance.py` (consumer; read for context only, not signed here) | `5e75e907392663394f01afa9d9d0f365568c3ae0f8bdc7cd0c222aa5ec59180c` |
| Records: `closure-impl-round-1.md`, `closure-impl-round-1-response.md` | `a03bd5db…`, `373d645f…` (Main supplied these digests truncated) |

## Basis (unchanged)

- `reviews/round-9.md` `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc`. It authorises (a)–(e). The clock started 2026-11-09T03:22:58.701909Z.
- `PREREGISTRATION.md` `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb`. §11 item 1 (action (b)): "bootstrap/attest installed scanner, dependency and numerical-interpreter closure". The §11 closing paragraph: "Seal stage/authentic council record/predecessors/…/all source and output hashes"; "Refuse substituted sources, stale versions".
- `reviews/rev9-digests.json` `2253911b…9c2da`
- `preparation/pins.json` `6d556fb5…100fc`
- `preparation/runtime-attestation-rev6.json` `19463957…62d62`
- `preparation/model-free-policy.json` `87de3949…2f9f`
- `preparation/model-free-binding-check.json` `5056ae74…c59c`

## Recommendation

**Do not run the preflight or the canonical closure yet.** Ten of the eleven round-1 fixes are in place, and their parsers now accept the real signed bytes. Two fixes introduced new defects:

- The R1-10 fix sends every enumerated file to the Mach-O classifier. That makes the R1-06 recipe (pass the git dependency kegs with `--external-root`) refuse on real bytes: the gettext keg contains Java class files.
- The R1-07 fix reads its review authority from an unpinned seal-time receipt. That brings back the R1-08 coupling, or else leaves the closure's authority untraceable.

Fix R2-01 to R2-04, then re-review in round 3.

## Prior-round disposition

| ID | Status | Evidence | Does the test catch the old defect? |
|---|---|---|---|
| R1-01 | **Fixed** | `attest_closure.py:160–203`. Only `python_full_version` and `platform_python_implementation` are accepted, with the operators allowed for each. Rows whose marker does not apply are dropped. A name that appears twice among applicable rows is refused. `python_packages` compares installed names with the applicable set. The lock's only `;` rows are l.277, 283 and 321 (peer-confirmed). | Yes. `test_r101_real_marker_lines` uses the three real lines. |
| R1-02 | **Fixed** | `native.py:13–41` reads the filetype, using the arm64 slice of fat files. l.44–56 expands `@executable_path`. l.98–125 carries the executable context through the queue. Executables are processed first, and the context-free repeat of a library is skipped. The dry check resolved the real `@executable_path/../lib/libpython3.11.dylib`. | Yes. `test_r102_*` covers executable → `@rpath` → `@executable_path`, and refuses when there is no executable context. |
| R1-03 | **Fixed** | `attest_closure.py:320` regex `^  (?! )(['"]?)(\S.*?)\1:$`. | Yes. The test uses the real `@vitest/mocker` peerDependencies block. |
| R1-04 | **Fixed** | l.272–289 scans only `node_modules` directly under the root and under lock-listed packages. | Yes. One internal case must pass and two stray cases must refuse. |
| R1-05 | **Fixed** | `verify_claw_assets` l.501–504 is called at l.640, before `pnpm_packages` at l.644. | The required test is present and catches substitution in the helper. It does not cover the call in `attest()`; see R2-04. |
| R1-06 | **Fixed in the closure**; a regression is open as R2-01 | `object_git` l.507–527 checks five passing receipts with an identical identity, refuses a shim and requires `libexec/git-core`. The keg git is used for `--version` and the submodule HEAD check (l.624–629). The keg is a root (l.668), and the output records `git=` and the receipt pins. On real bytes, all five receipts name `/opt/homebrew/Cellar/git/2.56.0_2/bin/git` with sha `371fc6bf…`. The receipt version is the stripped `git --version` output (`fetch_objects.py:45,145`), which matches the l.625 comparison. | Agreement and helper enumeration are tested. The shim, `git-core`, version and HEAD branches are not (R2-04). |
| R1-07 | **Partly fixed** → R2-02, R2-03 | `selected_source` l.554–560 is applied to every selected file. | The tests call the helper directly; no `go_metadata` route is tested (R2-04). |
| R1-08 | **Fixed** | `own_sources` l.461–463 is exactly the five files, also used for the journal pins at l.780. `execution_sources` and `source_manifest` were removed from the attester and from `verify_installed_closure`. | Yes. `test_r111_recheck_planted_null_pass` passes without the removed fields. |
| R1-09 | **Fixed** | l.530–537 requires the literal signed command and pinned `-I -B`. l.70 enforces `resolved_path`. l.540–543 and l.670 check interpreter roots. Real launchers (venv Python shebangs, `env node` → `.tools/bin/node`; `activate` has no shebang) resolve inside the roots, so there is no false refusal. | Yes, for each guard. |
| R1-10 | **Selection fixed** (l.675 passes every file; non-Mach-O files are skipped). **Catching test is ineffective** → R2-04. Regression → R2-01 | | No. `test_r110_*` calls `native_closure([addon])` directly. The defective selection was in `attest()` (old l.580–581), so putting the mode/suffix filter back would leave the test green. |
| R1-11 | **Fixed** for the boundaries listed in round 1 | Tests cover recheck drift and null cases, the deferred-row cases and the Go `go.sum`/`replace`/null cases. | Yes. The fixes added new refusal branches without tests; those are R2-04. |

## New issues

| ID | Severity | Location | Finding (input → wrong outcome) | Required fix |
|---|---|---|---|---|
| R2-01 | blocking | `closure/native.py:9–35` (the `CAFEBABE` fat branch), reached from `attest_closure.py:675` and the `pending.sort` call to `filetype_of` | **Java class files are taken for fat Mach-O images.** Since the R1-10 fix, every file in every root goes through `image_filetype`. That function treats `0xCAFEBABE` as a fat header and reads the Java version word as an architecture count, so it raises `truncated fat Mach-O header`, `invalid … count` or `no arm64 slice`. The R1-06 recipe passes git's dependency kegs with `--external-root` (round 1 R1-06; the response's R1-06 row). `INSTALL_RECEIPT.json` lists gettext among git's runtime dependencies, and the real `/opt/homebrew/Cellar/gettext/1.0` keg contains `share/gettext/javaversion.class` (510 bytes) and three `share/doc/gettext/examples/**/m4/*.class` files (glob-verified). Following the recipe therefore writes a refused canonical `closure.json`. | Recognise a fat Mach-O only when it is structurally valid: `nfat_arch` within a small bound (Java class versions start at 45), and every entry's offset/size inside the file and aligned. Treat bytes that parse as a Java class file (`u16 minor`, `u16 major` in the class-file range) as non-Mach-O. Refuse anything else that starts with a Mach-O magic but is malformed. Add fixtures for: a Java class (skipped), thin arm64, fat32 and fat64 containing arm64 (that slice is used), and truncated or out-of-bounds fat (refuses). |
| R2-02 | minor | `attest_closure.py:422–423`, `:554–560` | **R1-07 can still be bypassed with a symlink.** `go_metadata` resolves each selected file before `selected_source` checks whether it lies in ROOT. A selected `ROOT/internal/<pkg>/extra.go` that is a symlink to `/tmp/extra.go` resolves outside ROOT, so it is not checked against the pins or the reviewed files. It is hashed and retained as if it were authorised. Go compiles file symlinks (installed `src/go/build/build.go:886–900` excludes only directory symlinks). This is the same substitution round 1 required to be refused (§11 closing paragraph: "Refuse substituted sources"). | Authorise against the logical selected path (`Dir/name`) before resolving it. Refuse a main-module source whose link chain is non-empty or leaves ROOT unless the logical path, the target and the chain are each pinned. Add a `go_metadata` test that plants exactly this symlink and must refuse. |
| R2-03 | major | `attest_closure.py:546–552`, `:631`, `:692`, `:780` | **The Go-source authority comes from an unpinned seal-time receipt.** The R1-07 fix reads `OUT/reviews/implementation.json` (scope `execution-implementation`). That is the seal's `required_reviews` entry (`run-config.json:111–114`), and `binding/acceptance.py:261–271` requires the same file to bind the whole source manifest and `binding/runtime.json`. The receipt and its council record are pinned in neither the closure's `inputs` (l.692) nor the journal's `source_pins` (l.780). This gives two concrete failures. (i) Writing the full-execution receipt before (b) requires every binding and identifier byte to be final first, which restores the R1-08 coupling. (ii) Writing a receipt that covers only `model_free` means the canonical OUT file must later be replaced for (c), and the sealed evidence then has no record of which council record authorised (b)'s Go bytes. The second contradicts the §11 closing paragraph: "Seal … authentic council record/predecessors/… all source … hashes". A preflight before the receipt exists also refuses. | Read a dedicated receipt whose path and scope are constants in the attester, for example `OUT/reviews/closure-model-free.json` with scope `closure-model-free-go`. Validate it with `review_receipt`. Require its reviewed set to cover exactly the selected `model_free/*.go` files. Pin the receipt and its council record into the closure's `inputs` and the journal's `source_pins`. The seal then accounts for both through `verify_tree_pins` (`preparation.py:229`), and `run-config.json` does not change. Test: a missing, wrong-scope or drifted receipt refuses; a valid receipt is pinned. |
| R2-04 | minor | `tests/test_closure_review.py` | **Several fixes and new refusal points have no tests that would fail if the code were reverted.** (a) The R1-10 test bypasses the selection in `attest()`. (b) Nothing tests that `attest()` calls `verify_claw_assets` before `pnpm_packages`, adds the keg to the roots, or uses the keg git for the HEAD check. (c) `object_git` has untested branches: shim or not-`bin`, missing `git-core`, a receipt that did not pass. (d) The git `--version` drift and submodule HEAD mismatch refusals are untested. (e) No test has `go_metadata` refuse an unpinned file inside ROOT. (f) `image_filetype` classification is untested (covered by R2-01's fixtures). Failure: reverting any of these lines leaves the 121-test suite green (CLAUDE.md experiment rule 4). | Move the native candidate selection, the claw-asset-then-pnpm order and the git root/HEAD wiring out of `attest()` into small functions that `attest()` calls, or add one faked `attest()` test. Each item (a)–(e) needs a planted null that passes and a planted case that refuses. |

## Suggestions (not required for sign-off)

- **S1. A refused preflight is not recoverable without review.** The seal refuses any terminal journal that did not pass (`preparation.py:286–287`), so a refused journalled preflight needs reviewed recovery just like a refused canonical run. Its only saving is that `closure.json` stays unused. Recommend:
  - give the preflight its own retention directory (`SCRATCH/closure-preflight-retained/<id>`) rather than `SCRATCH/closure-retained/sha256` (l.685). Otherwise a preflight interrupted mid-copy leaves a partial object that makes the canonical run refuse.
  - write Main's runbook so that any preflight refusal is followed by a reviewed recovery before the canonical run.
- **S2 (round-1 S2 again, now concrete).** Homebrew git loads its dependencies through `/opt/homebrew/opt/<formula>` symlinks; its receipt lists pcre2, gettext, curl, openssl@4 and others. `verify_installed_closure` re-hashes the resolved Cellar files but never re-resolves the install names. If an upgrade keeps the old keg (`HOMEBREW_NO_INSTALL_CLEANUP`), git loads a new library while the recheck still passes. Recommend storing each on-disk native dependency's link chain and re-resolving it in `verify_installed_closure`. With default cleanup the old keg is removed and the recheck refuses, which is fail-closed.
- **S3. Treat non-host architectures consistently.** A thin x86_64 image is walked; for example, the inferred `pnpm/dist/node_modules/@reflink/reflink-darwin-x64/*.node`. A fat image without an arm64 slice refuses the whole closure. Recommend recording images with no arm64 code as `non_host_architecture`, covered by enumeration but not walked for loads.
- **S4. Inherited `@rpath`.** `native.py` resolves `@rpath` only against the image's own `LC_RPATH`. dyld also uses the rpaths of the loading chain (`dyld(1)`). Add an inherited-rpath fixture before relying on this. Today it fails closed.
- **S5. A binding-packet gap that affects this create-once output.** `acceptance.py:46,54` permits only `executable_resolutions` rows, and the keg git is not among them; nothing reads the closure's `git` field. Recommend that the binding review fix this on the binding side: permit the `closure.git` (path, sha) and other `original_runtime_files` pairs, except `claude`/`codex`. No closure change is needed, but settle it before the canonical closure run.
- **S6. Model-CLI cask drift (round-1 S5).** `resolved_path` is now enforced for `claude` 2.1.311 and `codex` 0.166.0. A cask upgrade before (b) refuses the closure; one after (b) breaks every later recheck.
- **S7.** Round-1 S7 (streaming `pin()` for the GB-scale dyld cache) and S8 (`closure.json` size) still apply.

## Alternatives

| Option | Tradeoff | Decision |
|---|---|---|
| A. Fix the classifier (R2-01) and keep enumerating whole kegs | Keeps R1-06's tamper evidence over the kegs; only `image_filetype` changes. | **Recommended** |
| B. Leave the gettext keg out of `--external-root` | Avoids the refusal but departs from the reviewed R1-06 recipe, and other kegs may hold similar files. | Rejected |
| C. Revert to selecting native images by mode or suffix | Brings back R1-10 (missed `.node` addons). | Rejected |
| D. Write `implementation.json` before (b) covering the full execution | Brings back the R1-08 coupling. | Rejected in favour of R2-03 |
| E. Run the preflight without a journal | An unrecorded operation under preparation authority, against the §11 closing paragraph's provenance requirement. | Rejected; S1 instead |

## Evidence

- **Real bytes (read-only):**
  - All five `preparation-run/objects/*.json` have identical `git_identity` (`/opt/homebrew/Cellar/git/2.56.0_2/bin/git`, sha `371fc6bf…a11f`).
  - The git keg has `libexec/git-core`. The `bin/git-{receive,upload}-pack` / `git-upload-archive` entries are 0-byte, so likely in-keg links [INFERENCE].
  - `INSTALL_RECEIPT.json` l.52–192 lists git's runtime dependencies, including pcre2 10.49 and gettext 1.0.
  - Glob found the gettext `.class` files listed in R2-01. There are no `.class` files in `scout-x1/.tools`, GOROOT `go/1.26.10` or the git keg.
  - The GOROOT `src/debug/macho/testdata/*` files are `.base64` text.
  - `.tools/activate` has no shebang.
  - The `pins.json` entrypoints at l.4069–4195 resolve inside the roots.
- **Code:** `attest_closure.py:422–423, 501–504, 507–527, 546–560, 624–631, 668–692, 779–783`; `native.py:9–41, 98–125`; `preparation.py:172–185, 229, 261–287`; `binding/acceptance.py:41–55, 261–271`; `run-config.json:107–123`.
- **Measurement record (CLAUDE.md experiment rule 3):** the author ran `attest_closure.py --dry-check` against the real `scout-x1` runtime. It is read-only, prints to stdout only, and its only subprocess was `otool -L` on the interpreter. It ran no scanner or model and computed no predictor–outcome association. Recorded here as a measurement/construct check. Its scope is limited to R1-01 to R1-04; it is not evidence that the native walk, dyld cache walk or Go walk succeed.
- **Main's verification (reported, not rerun):** 121 tests OK; `go vet` and `go test` clean. The eslogger sudoers drop-in is absent; that is a binding prerequisite and does not affect (b).

**Where peers changed this record:**
- The innovation-challenger found the gettext `.class` files, which I had only hypothesised as a GOROOT risk; I verified them by glob, and they are now R2-01. It also found the R1-07 symlink escape (R2-02), which I accepted at R1-07's minor severity.
- The experiment-peer independently raised the unpinned review receipt that I had drafted (R2-03); its acceptance.py evidence raised the severity to major.
- Both peers found that the R1-10 test misses the fixed line, which became R2-04 together with the peer's untested branches.
- I downgraded the peer's N1 (preflight) to S1 because it fails closed, and its N2 (binding git allowlist) to cross-packet note S5, because the fix belongs to the binding side and needs no closure change.

## Risks

- **Irreversibility.** A canonical refusal uses up `OUT/closure.json`, and a refused preflight blocks the seal (S1). Both need reviewed recovery.
- **Unverified on real bytes** [INFERENCE]: `otool` and `dyld_info` output across every Mach-O in the roots (including `.tools/cache`); dependencies that need an inherited rpath; offline `go list -m all` / `go mod verify`; whether other git dependency kegs contain files that look like Mach-O.
- **45-day dependency** on unchanged `scout-x1/.tools`, the Homebrew git keg and its dependency kegs, mise Go 1.26.10 and the model-CLI casks (S2, S6). After (b), no `brew upgrade` or `mise` changes until the seal.

## Execution outline

1. R2-01: rework the fat/Java/thin classification in `native.py`, with fixtures.
2. R2-02: authorise the logical selected path before resolving it, with a `go_metadata` symlink test.
3. R2-03: read a dedicated `closure-model-free` receipt, pin it and its record into the closure's `inputs` and the journal's `source_pins`, with tests.
4. R2-04: move the wiring into small functions, or add a faked `attest()` test, with the planted-null and planted-refusal tests listed. Adopt S1 (separate preflight retention).
5. Main reruns `.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py'`, recomputes every component SHA-256 and submits round 3.
6. After a zero-issue round: write the dedicated review receipt and keep `scout-x1`, the Homebrew kegs and the casks unchanged. Run the journalled preflight with the round-1 command, `--external-root` for every git dependency keg and the required `--dyld-cache-root`. If it passes, run the canonical command once.

## Verdict

**4 issues open**: R2-01 blocking; R2-03 major; R2-02 and R2-04 minor. The runtime closure attestation component must not be run, preflight or canonical, at its canonical paths. No sign-off is given for the digests above.
