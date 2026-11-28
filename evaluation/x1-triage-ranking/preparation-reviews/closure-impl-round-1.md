# X1 preparation runtime closure attestation (b) implementation review — round 1

**Scope.** This round reviews action (b) only: the read-only runtime closure attester and its shared helpers. It does not review the binding (c), identifier (d) or seal (e) packets, except where they consume this component's output. It is read-only: no attester, scanner, test, build, model call or hash computation was run. All findings come from reading files.

## Reviewed files (SHA-256 computed by Main with `shasum -a 256`; not recomputed here because this session had no shell)

| File (relative to `evaluation/x1-triage-ranking/execution/`) | SHA-256 |
|---|---|
| `attest_closure.py` | `61b98abb9f3ea5822800cf05b9c6b40ff15620d90b69978ee2dbaf82bfa58694` |
| `closure/native.py` | `f2bad611ec65021cc3623dccdddd56edf32cc659c6a139e3a579eeb3e3b15b94` |
| `tests/test_closure.py` | `ee83374f996eaecffac8bbb47ae8f57b4c2819c7442298ded339573d27e89efa` |
| `preparation.py` (shared) | `1a116d86b98b088c474a0f4e974af9be8299484ac1deee99dba8f3e7123bc9bf` |
| `run-config.json` (shared) | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |
| `verify_inputs.py` (imported helpers `pin`, `working_pin`, `exclusive_json`; read for context) | no digest supplied; not in this sign-off |

## Basis digests

- `reviews/round-9.md` `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` (authorises actions (a)–(e) only; clock started 2026-11-09T03:22:58.701909Z)
- `PREREGISTRATION.md` `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb` (l.218–228; action (b) at l.222: "bootstrap/attest installed scanner, dependency and numerical-interpreter closure"; l.228: "Refuse substituted sources, stale versions")
- `reviews/rev9-digests.json` `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`
- `preparation/pins.json` `6d556fb5156e8ced3c17ae7526a841e51688145faad4a3f8a6a2ce0c939100fc`
- `preparation/runtime-attestation-rev6.json` `19463957becfc23a391eac03582b32eaa0ff392069bccd9aed9000f390ad2d62`
- `preparation/model-free-policy.json` `87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f`
- `preparation/model-free-binding-check.json` `5056ae748e05717e7a1e16dd8974f11519045c01bdc3ce6498adf3b14aedc59c`
- Main's plan `local://x1-prep-exec-plan.md` §4 was used as context only; it is not authority.

## 1. Recommendation

**Do not run yet.** Use the historical `scout-x1/.tools` installation read-only. Do not bootstrap. Fix the 11 issues below, then re-review in round 2.

Four of the issues each make the attester refuse on every run against the signed runtime: R1-01 to R1-04, all confirmed from the real bytes. `closure.json` is create-once, so the first run would write a refused record at the canonical path. Every later run would stop at `attest_closure.py:633–634` and need a reviewed recovery.

**Reuse versus bootstrap.** Read-only reuse is the option the signed text points to.
- P l.222 allows "bootstrap/**attest**"; it does not require reinstalling.
- The signed numerical interpreter is literally `/Users/adamjackson/Projects/work/scout-x1/.tools/venv/bin/python` (`runtime-attestation-rev6.json` l.7–12).
- Every pinned entrypoint `resolved_path` is under `scout-x1/.tools` (`pins.json` l.4069–4180).
- Bootstrapping would rewrite launchers and `activate` (which explains the prep worktree's 5 failing pins). It would also run `claude`/`codex --version` (`scripts/bootstrap-tools.sh` l.214–215, `report_system`) and do a networked `go install` (l.225). Both are outside what this action should do.

**Coupling to the whole execution manifest.** The closure should not require a manifest of every execution byte (R1-08). That requirement is real, but it belongs to the binding receipt (`binding/acceptance.py:222–230`) and the seal (`preparation.py:261–266`), which already enforce it. If the create-once runtime receipt also binds unrelated binding and identifier bytes that are still under review, any later fix to them invalidates it permanently.

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R1-01 | blocking | `attest_closure.py:160–177` (regex l.167), called at l.548 | **The parser rejects the signed Python lock.** The regex has no room for environment markers. The pinned `toolchain/python-requirements.lock` has `deprecated==1.3.1 ; python_full_version < '3.12'` (l.277), `deprecated==3.0.0 ; python_full_version >= '3.12'` (l.283) and `librt==0.16.0 ; platform_python_implementation != 'PyPy'` (l.321). `requirements()` raises `unsupported frozen requirement syntax` on every run. Fixing only the regex is not enough: the duplicate-name check (l.171–172) and `set(observed) != set(wanted)` (l.212) would then fail on the inapplicable `deprecated 3.0.0`. | Parse an optional `; marker`. Evaluate only `python_full_version` and `platform_python_implementation` against the running interpreter (3.11.11, CPython), and refuse any other marker variable or operator. Keep only rows whose marker applies, require exactly one applicable row per name, and compare installed distributions with that applicable set. The 53 installed `*.dist-info` then match 53 applicable names. Add a test that uses these three real lock lines. |
| R1-02 | blocking | `closure/native.py:91–95`; interpreter enters at `attest_closure.py:580–583` | **The native pass cannot resolve `@executable_path`, and the pinned CPython uses it.** The raw bytes of `scout-x1/.tools/python/cpython-3.11.11-macos-aarch64-none/bin/python3.11` (48.8 KB; `lib/libpython3.11.dylib` is present) contain the load command `@executable_path/../lib/libpython3.11.dylib`. Line 92 raises on it, so every run refuses on the numerical interpreter itself. | Read the Mach-O header filetype. For `MH_EXECUTE`, resolve `@executable_path` to the image's own directory, and carry that executable context with each queued dependency so `@executable_path` and `@executable_path`-based `@rpath` entries inside the dylibs it loads also resolve. Refuse only when no executable context exists. Add a fake-`otool` test for exactly this layout. |
| R1-03 | blocking | `attest_closure.py:299` (also l.309) | **The pnpm header regex matches nested keys.** `^  ['"]?(.+?)['"]?:$` also matches 4- and 6-space lines such as `    peerDependencies:`, `    peerDependenciesMeta:`, `      msw:` and `      vite:`. These exist in the installed (and pinned) Clawpatch `pnpm-lock.yaml` `packages:` section at l.539–545, 576–580, 680–686, 693–699, 769–785 and 812–828. Those keys collect no integrity, so l.309 raises `missing pnpm package integrity declaration` on every run. | Match only exactly-two-space headers, e.g. `^  (?! )(['"]?)(\S.*?)\1:$`, and strip one balanced pair of quotes. Add a test built from a real `peerDependencies`/`peerDependenciesMeta` block. |
| R1-04 | blocking | `attest_closure.py:244–252` (condition l.249) | **The npm "unlocked package" scan treats files inside a package as installed packages.** The pinned `pnpm@11.25.0` tarball ships its own `node_modules/pnpm/dist/node_modules/{yallist,which,semver,tar,undici,…}/package.json` and `…/@reflink/*`. None of these are lock rows (no `inBundle`, `bundleDependencies` or `pnpm/dist` in `.tools/node-tools/package-lock.json`). Because `parent.parent.name == 'node_modules'`, l.252 raises `unlocked npm package: node_modules/pnpm/dist/node_modules/…` on every run. | Walk the npm install layout structurally. A dependency directory is `<P>/node_modules/<name>` or `<P>/node_modules/@scope/<name>`, where `<P>` is the node-tools root or itself a lock-listed package directory. Bytes inside a package's own tarball are covered by the full enumeration, not by lock rows. Add a test with a package-internal `dist/node_modules/x/package.json` that must pass and a stray top-level package that must refuse. |
| R1-05 | major | `attest_closure.py:283–291, 548–550` | **Clawpatch's installed lock is only checked against itself.** The installed `pnpm-lock.yaml` is compared only with its own `node_modules/.pnpm/lock.yaml`, and the `allowed` package identities come from that same unpinned file. The signed lock (`pins.json` l.4564–4568, sha listed there) and the signed `package.json` / `pnpm-workspace.yaml` (l.4557, l.4571) are never compared with the installation. Failure: an install whose two locks were changed consistently passes, which contradicts P l.228 "Refuse substituted sources". `build-clawpatch.sh` l.105 copies these files from the gitlink commit via `git archive`, so they must equal the working-file pins. | Before interpreting the installed files, require `claw/package.json`, `claw/pnpm-lock.yaml` and `claw/pnpm-workspace.yaml` to match the sha256/bytes of the corresponding `pins.submodule.working_files` rows. Add a test where two installed locks match each other but are substituted, which must refuse. |
| R1-06 | major | `attest_closure.py:535–536, 577–579` | **The git that actually runs is not attested.** `/usr/bin/git` is the xcrun shim. The git it hands off to, with its `git-core` helpers and dylibs, is never enumerated. Action (a) used `/opt/homebrew/Cellar/git/2.56.0_2/bin/git`, sha `371fc6bf…a11f`, version 2.56.0 (`preparation-run/objects/django__django.json` l.47–52). Scout's engine and the scanners resolve git through PATH (`internal/engine/matrix.go:96`). `binding/acceptance.py:49–54` accepts only a git listed in `original_runtime_files`, which currently forces (c) onto the shim. An `xcode-select` change or a Command Line Tools update would change the git in use while `verify_installed_closure` still passes (P l.228 "stale versions"). | Attest the git recorded by action (a): require the five object receipts to agree on its path and sha256, and pin it. Enumerate its keg (`/opt/homebrew/Cellar/git/2.56.0_2`, including `libexec/git-core`) as an explicit root, include its Mach-O helpers in the native pass, and pass its dylib kegs via `--external-root`. Use it for the submodule HEAD check, and record it as `closure.git` so (c) can bind to it. |
| R1-07 | minor | `attest_closure.py:379–387` versus l.480–482 | **Go files outside the pins are compiled without refusal.** `go list -deps` from `ROOT` hashes and retains every selected main-module file, but nothing requires those files to be pinned. An untracked `.go` file, or one added after `7bb73e52`, in a selected package would pass silently (P l.228). | For every selected file inside `ROOT`, require its relative path and sha256 to be in `pins.source.files`, or to be one of the reviewed `execution/model_free/*.go` files. Refuse otherwise, and add a test. |
| R1-08 | major | `attest_closure.py:425–439, 479, 450–453, 598–600, 640` | **The closure binds every execution byte into a create-once runtime receipt.** That includes binding, identifier and test code that is still under separate review. Any later fix invalidates `execution_sources` in `verify_installed_closure` and at the seal, and `closure.json` cannot be recreated without a reviewed recovery. The attester's file set also contradicts the seal's: `attest_closure.py:430–436` includes `__pycache__`, while `preparation.py:263` excludes it. No producer for `execution-source-manifest.json` exists. | Decouple. The closure pins only its own import closure (`attest_closure.py`, `closure/native.py`, `preparation.py`, `verify_inputs.py`, `run-config.json`) in the journal `source_pins` and the output, plus the Go selected files. Remove `execution_sources` / `source_manifest` from the attester and from `verify_installed_closure`. Whole-execution coverage stays where it is already enforced (`acceptance.py:222–230`, `preparation.py:261–266`). Run prerequisite: `execution/model_free/*.go` must be signed off and unchanged before the closure runs, because the Go selected files bind them. |
| R1-09 | minor | `attest_closure.py:470, 500, 498, 519–534` | **The runtime location is not tied to the signed location.** `--runtime-root` is free. The interpreter check uses `runtime_root/.tools/venv/bin/python`, not the signed command path, and launcher interpreter resolutions are pinned but never required to lie inside the enumerated roots. Failure: pointing `--runtime-root` at a byte-identical copy (for example the "immutable copy" the plan proposes) passes, while the launchers' absolute shebangs still run the `scout-x1` interpreter, which is then neither enumerated nor rechecked. | Require `runtime_root/.tools/venv/bin/python` to equal `runtime['numerical_interpreter']['command']`. Require each `checked()` realpath to equal the pinned `resolved_path` where one is given. Require every `interpreter_resolution` path to lie inside the Inventory roots. Add tests. |
| R1-10 | minor | `attest_closure.py:580–581` | **Native images are selected by mode or suffix.** That misses Mach-O addons that are present in the closure, e.g. `.tools/clawpatch/734f81bc…/node_modules/.pnpm/@oxlint+binding-darwin-arm64@1.81.0/…/oxlint.darwin-arm64.node`, yet the output still says `complete: true`. | Select native images by Mach-O magic across every enumerated regular file. Add a test with a non-executable `.node` file. |
| R1-11 | minor | `tests/test_closure.py` | **Decision boundaries have no tests.** There is no planted-drift / planted-null test for `verify_installed_closure` (added file, mode change, retargeted link, writable retained object, entrypoint resolution drift, and a no-drift pass). The deferred-row branches (count ≠ 176, set mismatch, field mismatch, empty deferral), `go_metadata` refusals (go.sum mismatch, `replace`), `attest()` interpreter/location refusals, and the R1-01 to R1-10 cases are also untested (CLAUDE.md experiment rule 4). | Add synthetic tests for each listed boundary. Fixture copies of the real lock syntax are fine. No runtime execution is needed. |

## Suggestions (not required for sign-off)

- **S1. Add a preflight.** Add `--preflight`, which runs the same `attest()` but writes only under `SCRATCH/closure-preflight/<journal-id>.json` and is journalled as action (b). Make `--dyld-cache-root` required: leaving it out guarantees a refusal (`native.py:41`). Run the preflight before the canonical run.
- **S2. Keep on-disk native dependencies inside approved roots.** Refuse non-system on-disk native dependencies that are outside the Inventory roots, so their link chains are enumerated and rechecked. This applies the Inventory's own escaped-root policy to `native.py`.
- **S3. Take the runtime baseline before any subprocess.** Enumerate `.tools` and `sys.base_prefix` before the first subprocess runs; Go roots can follow.
- **S4. Validate the dyld cache family.** Check the main cache's declared subcache list (UUID and bounds) rather than taking files by name prefix. Recheck the active cache UUID, `sw_vers` and Darwin/arm64 in `verify_installed_closure`.
- **S5. Exclude `claude`/`codex` from the drift gate.** Record their bytes as metadata, but leave the rows with `execution_allowed=False` out of the drift check (`attest_closure.py:457–460`). A routine Homebrew upgrade could otherwise block the seal with no safety benefit; the binding sandbox denies those paths anyway.
- **S6. Refuse unowned startup files.** Refuse `site-packages` files that no RECORD owns and that run at interpreter start (`*.pth`, `sitecustomize`), except an allowlisted `_virtualenv.{pth,py}`.
- **S7. Hash large files by streaming.** Use streaming hashing in `pin()`. It currently reads whole files into memory, including GB-scale dyld cache files.
- **S8. Keep `closure.json` small.** Move the bulk enumeration and command stdout into content-addressed SCRATCH artifacts pinned from `closure.json`. The record lists files roughly three times, so it may exceed GitHub's 100 MB file limit [INFERENCE].
- **S9. Record download checksum provenance.** Record `.tools/downloads/*` against the pinned `toolchain/checksums.sha256`.
- **S10. Keep `scout-x1` untouched until the seal.** Every recheck needs the `scout-x1/.tools` tree to stay unchanged; any non-`-B` Python run there writes `__pycache__` into `site-packages`. Main should use `scout-x1` for nothing else until the seal.
- **S11. Fix two provenance details.** Record an identity mapping for the runtime root: `historical_path_mapping` currently implies a relocation that did not happen. Add a seal-side producer for the execution manifest that refuses `__pycache__`.

## 2. Alternatives

| Option | Tradeoff | Decision |
|---|---|---|
| **A. Read-only reuse of `scout-x1/.tools`** | Matches the signed interpreter path and all pinned `resolved_path`s, with no relocation. It needs that tree to stay unchanged and must tie the run to the signed location (R1-09). | **Recommended** |
| B. Bootstrap in the prep worktree | Rewrites launchers and `activate` (the 5 failing pins). It also runs `claude`/`codex --version` and a networked `go install`. It would need a location-compatibility review. | Rejected |
| C. Rebuild at the historical path | The path is occupied, and overwriting it is forbidden. | Not available |
| D. Closure coupled to the full manifest, with a reviewed-manifest precondition (peer I6) | Enforces review order, but every later execution fix still breaks the create-once closure, and the seal already covers this. | Rejected in favour of R1-08 |
| E. Leave the coupling as it is | Breaks the closure permanently on any fix, and the `__pycache__` rules can't both be met. | Rejected |

## 3. Evidence

- **Signed text:** P l.222 (action (b)) and l.228 (refuse substituted or stale). `README.md` l.7 and l.20. `runtime-attestation-rev6.json` l.7–12. `pins.json` l.4069–4180 (entrypoints and assets), l.4557–4575 (Clawpatch manifest, lock and workspace pins), l.4037 (the Python lock is pinned).
- **Real bytes observed:**
  - `python-requirements.lock` marker rows at l.277, l.283 and l.321.
  - The `python3.11` load command `@executable_path/../lib/libpython3.11.dylib`.
  - `pnpm-lock.yaml` `peerDependencies:` blocks at l.539 and onward. `.pnpm/lock.yaml` has the same structure (equal byte-for-byte [INFERENCE]).
  - `node-tools/node_modules/pnpm/dist/node_modules/*/package.json` exists, and the npm lock has no rows for it.
  - Action (a)'s git is Homebrew 2.56.0 (`objects/django__django.json` l.47–52).
  - 53 installed dist-infos.
- **Build path:** `build-clawpatch.sh` l.105 (the installed source comes from `git archive` of the gitlink commit), l.138 (safety stamp), l.18–19 (patch marker).
- **Deferred rows:** `attest_closure.py:483–494` checks all 176 `working_files` rows in `ROOT` even when the deferral list is empty. When the list is not empty, it requires the count to be 176, the path set to be equal, and sha256/bytes/mode/blob to equal the pins. `preparation.py:187–199` rechecks this at the seal. This is correct for the deferral contract: the prep worktree is the gated checkout. The `runtime_root` working tree is not a build input (`git archive`), so R1-05 binds the installation to the same pins instead.
- **Retained copies:** `retain()` (l.127–157) writes each object with exclusive create, fsync, mode 0444, a re-hash of the source after copying, and an fsync of the directory. Directories, symlinks and modes exist only in `enumeration`. `verify_installed_closure` re-hashes the retained objects and re-enumerates the original roots. This is correct as tamper evidence. There is no restore path (S10).
- **Main's verification (reported, not rerun):** 58 unit tests pass; `go vet`, `go build -pgo=off` and `go test` are clean. These tests do not cover R1-01 to R1-11.

**Where peers changed this record:**
- The experiment-peer added R1-06 (git) and R1-07 (Go sources), which I had not found, and independently confirmed R1-01 and R1-02.
- The innovation-challenger supplied the bootstrap evidence (`claude`/`codex --version`). It raised R1-05 alongside me, and its `.node` finding moved R1-10 from a suggestion to a minor issue.
- I resolved the peer's I6 differently: decoupling (R1-08) instead of a reviewed-manifest precondition.
- I downgraded the challenger's dyld subcache, native link replay and baseline-timing findings to S4, S2 and S3. Each needs Main to supply a non-system cache copy or a planted write, which is not a plausible path when the attester is run against the system cache directory.
- I kept the peer's S7 (`claude`/`codex`) as a suggestion (S5) because it is a fail-closed default.

## 4. Risks

- **Irreversibility.** A canonical refusal permanently uses up `OUT/closure.json` and needs a reviewed recovery round. S1 (preflight) is the mitigation.
- **Other refusals not yet found** [INFERENCE]:
  - other Mach-O images may use `@rpath` / `@executable_path` forms in ways not yet seen;
  - the output format of `dyld_info -uuid -linked_dylibs` on this macOS build;
  - whether `go list -m all` and `go mod verify` succeed offline against the scratch module-cache copy;
  - whether `node_modules/.pnpm/lock.yaml` and `pnpm-lock.yaml` are byte-equal.
- **45-day dependency on an unchanged `scout-x1/.tools`.** Any stray write blocks the seal (S10).
- **`closure.json` size** may block the writer PR (S8) [INFERENCE].
- **Ordering.** After R1-08, the closure still has to wait until `model_free/*.go` is final.

## 5. Execution outline

1. Fix R1-01 to R1-04 and add a regression test per parser using the real syntax (the three marker lines, a `peerDependencies` block, a pnpm `dist/node_modules` layout, an `@executable_path` executable).
2. Fix R1-05 (pin the Clawpatch manifest, lock and workspace), R1-06 (attest the action-(a) git with its keg and dylibs; record `closure.git`) and R1-07 (refuse unpinned selected Go files).
3. Apply R1-08 (pin only the attester's own import closure; remove execution-manifest coupling) and R1-09 (tie the run to the signed runtime path and interpreter roots).
4. Apply R1-10 (select native images by magic) and R1-11 (boundary and planted-drift/null tests). Adopt S1 (`--preflight`, required `--dyld-cache-root`).
5. Main reruns `.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py'`, recomputes all component SHA-256s, and submits round 2.
6. After zero-issue round 2: confirm the binding packet's `model_free/*.go` are signed off and unchanged; leave `scout-x1` untouched; run the preflight (journalled action b). Then run the canonical command once, from cwd `ROOT`:
   `/Users/adamjackson/Projects/work/scout-x1/.tools/venv/bin/python -I -B "$EXEC/attest_closure.py" --config "$EXEC/run-config.json" --clock "$OUT/clock.json" --pins "$STUDY/preparation/pins.json" --output "$OUT/closure.json" --runtime-root /Users/adamjackson/Projects/work/scout-x1 --module-cache <SCRATCH Go module-cache copy> --dyld-cache-root <active OS dyld directory> --external-root /opt/homebrew/Cellar/git/2.56.0_2 [--external-root <git dylib kegs>]`
7. Bind (c) to `closure.git`, the attested Python and `original_runtime_files`; the seal then rechecks everything through `verify_installed_closure`.

## Verdict

**11 issues open** (R1-01 to R1-04 blocking; R1-05, R1-06 and R1-08 major; R1-07, R1-09, R1-10 and R1-11 minor). The runtime closure attestation component must not be run at its canonical paths. No sign-off is given for the digests listed above.
