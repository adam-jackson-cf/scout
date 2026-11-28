# X1 execution-implementation council — round 2 (also the binding repair round under amendment `:74`)

**Study:** `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/` (STUDY). OUT = `STUDY/preparation-run`. SCRATCH = `/Users/adamjackson/Projects/work/x1-preparation-scratch`.
**Date:** 2026-11-10.
**Action:** preparation action (c), binding v2.

This round reviews the repair of the round-1 issues X1EI-R1-01 and X1EI-R1-02. Under `preparation/model-free-binding-amendment.md:74` it is also the zero-issue council round for the `execution/binding/**` repair. It then issues the `execution-implementation` authority.

## Inputs

SHA-256 digests are cross-checked across documents but were not recomputed by the council.

| Input | SHA-256 |
|---|---|
| Round-1 record `preparation-reviews/execution-implementation-round-1.md` | `f57c945d9348332cea169796137e867ac8e71da81ec664309d4f095fdd01cdbb` |
| Writer response `preparation-reviews/execution-implementation-round-1-response.md` | `b8335c6c3a8fb24fc18651240adf18c764ef6b08a183df37e801d58c756e2252` |
| Checks artifact `preparation-reviews/execution-implementation-round-2-checks.json` | `8356567416acc9fd52bb6f061e8f36bacf87235383b3fd5a4c48237a90f122ff` |
| Integrity evidence `preparation-reviews/execution-implementation-round-1-evidence.json` | `e6b624b5c7eebcf590575d1fd8431fa062bbc82664236d3160321fd1cbd30d6c` |
| `execution/binding/acceptance.py` (29957 B; was `2144999a…`) | `6e56fb5f6f2f8f1041c2d06498b6f2d01c0e7480a2911def12902094e955de63` |
| `execution/tests/test_binding.py` (30358 B; was `2458e428…`) | `be91c678512402244ff27847177046511caa129ac6d8834b079c2add35ccc155` |
| `execution/tests/test_binding_adoption.py` (20936 B; was `0396765f…`) | `985b78b2294680c3a564b80b2f92562bda2ceada0419d85d1ecc0c1f398e42aa` |
| Regenerated `OUT/execution-source-manifest.json` (46 rows, 11373 B) | `25ef2b4e6387f5ed496acdc4d28c187fcf1a38f71fdb2d783bda5bbf78a3d147` |
| Regenerated `OUT/binding/runtime.json` (4891 B) | `647381c553c78516fe19fe7d83e7401be1692e6a6922e93f4218606b52ebbc6d` |
| Superseded manifest, moved to `SCRATCH/binding-v2-build/superseded-r1/execution-source-manifest.json` | `8be62162f5f13d6a5b410efa3d23f5a11dda1fa7c26e5af5521b90dcb3f9cb96` |
| Superseded runtime, moved to `SCRATCH/binding-v2-build/superseded-r1/runtime.json` | `4f80091c7356e6b8940c782b498603defaea6577171a3323e6fe598f504ec9c9` |
| Frontend `SCRATCH/binding-v2-build/model-free` (unchanged; not rebuilt) | `05fa048c4911666e107e8e272770308cef3b23535b6c0dd208f36f7768a0a139` |
| `OUT/closure.json` | `b5df46b1b92a3cc9ef2905589db120a008881bc6b5519f37b6d38e3705cc35b2` |
| `STUDY/RESULTS.md`, section "Frontend build outside the protected-cache command rule (X1EI-R1-02)" | (read in place) |

**Method:** three independent read-only analyses: Main council (Opus), `innovation-challenger` (GPT Astra) and `experiment-peer` (Opus). No launch, build, test or file mutation was done.

## X1EI-R1-01 — Resolved

**Steps 1–3: the code.**
- `private_path(runtime)` (`acceptance.py:41–42`) is exactly the prescribed expression. `run_revision` uses it (`:117`).
- `validate_runtime` (`:45–89`) checks the following:
  - `runtime['git']` must equal `closure['git']` and be in `original_runtime_files` (`:52–56`).
  - The Go pin is the unique `original_runtime_files` row at `closure.go.goroot/bin/go` (`:59`).
  - The `sandbox-exec` pin is the unique `os_executables` row `/usr/bin/sandbox-exec`, which must also be in the originals (`:60–63`).
  - Each support pin must appear exactly once in `executables` (`:65–68`).
  - Only allowed resolutions, the three support pins and the exact frontend are admitted (`:73–74`).
  - The resolution guard `shutil.which(name, path=private_path(runtime))` must resolve, via `Path.resolve()`, to the attested path for `git`, `go`, `sandbox-exec` and the unique allowed `node` row. `validate_pin` then checks that file's sha256 and size (`:79–86`).
  - `claude` and `codex` must not resolve at all (`:87–89`).
- The guard runs at every operation launch: `reviewed_inputs` calls it (`:384`) on both the rehearsal path and the canonical path.

**Real-data check.**
- Against the canonical closure, the regeneration run validated the new runtime with exit 0 and printed these resolutions:
  - `git` → `/opt/homebrew/Cellar/git/2.56.0_2/bin/git`;
  - `go` → `…/mise/installs/go/1.26.10/bin/go`;
  - `sandbox-exec` → `/usr/bin/sandbox-exec`;
  - `node` → `…/scout-x1/.tools/node/bin/node`;
  - `claude` and `codex` → none.
- The PATH order is safe:
  - `.tools/node/bin` comes before `.tools/bin`, and either would pass, because `resolve()` follows the `.tools/bin` symlink to the attested node.
  - The Homebrew git and mise go directories come before `/usr/bin`, so the Xcode `git` shim cannot shadow them.
  - The frontend directory contains no competing tool name.

**Step 4: the tests.** `test_binding_adoption.py:135–158` builds a closure with the real shape:
- git, go and sandbox-exec appear only in the originals and `os_executables`;
- node is an allowed resolution.

The seal fixture reuses this closure. Each required case has a passing null first, then a planted effect, and each planted effect asserts the specific refusal message:
- missing, wrong sha256, duplicate and unattested path for each of the three support pins (`:181–203`);
- an unrelated original file must not be admitted (`:205–216`);
- a shadow `git` in an earlier PATH directory (`:218–230`);
- a planted `claude` and a planted `codex` (`:232–243`).

`test_binding.py:117–138` checks that the launch PATH is exactly the stub directory, a colon, then `private_path(runtime)`, and that duplicate parent directories are removed.

Recorded results, with argv, timestamps and exit codes in the checks artifact:

| Suite | Tests | Exit |
|---|---|---|
| Binding (`test_binding*.py`) | 61 | 0 |
| `test_binding_contract.py` (round-4 S3) | 6 | 0 |
| Full Python suite | 343 | 0 |

**Step 5: regeneration.**
- **Manifest.** jq comparison, row by row against the superseded copy: both have 46 rows. Exactly three rows differ, and they are the three changed files above. All other rows are unchanged, including:
  - `preparation.py` `9572aed3…`;
  - `main.go` `edb37ec2…`;
  - `test_binding_seal.py` `7c05eaf2…`;
  - `test_binding_contract.py` `dc817736…`;
  - both amendment copies `88eb95a5…`;
  - the three I6 rows.
- **Runtime.** It keeps the same six keys, and the `git`, `python`, `frontend` and `clawpatch` pins are unchanged. `executables` is the old 16 rows in their old order, followed by:
  - git `371fc6bfbfea4d9ace7d122feefe100503d24f827bc94f54ab3973583ffea11f`;
  - go `5742f0004c49c005313d88cd06df964b09cad6bcf2eac4506bdb90abc923b1f4`;
  - sandbox-exec `e3d7a792c58a5d3783d2f7274c82d70062393830d8cb1ded713ca554a470bd2f`.

  Each equals the closure's own row.
- **Move.** The superseded copies were moved byte-identically: old and retained pins are equal in the checks artifact `.superseded`. OUT holds only the regenerated files.

**Pinning.** No journal, closure field or receipt pins `binding/**` or the changed tests, so the repair keeps the seal intact (F3). `main.go`, Core, the closure, the I6 files and the amendment authority are unchanged.

## X1EI-R1-02 — Resolved

- **(a) Closure checks.** The real `attest_closure.verify_installed_closure` passed in 54.46 s, and `preparation.verify_tree_pins` passed over 92,066 pins (evidence `.checks[0]`, exit 0). Neither was mocked.
- **(b) Cache listing.** The listing uses `lstat` and a walk that does not follow symlinks, over the root, every directory and every file of `SCRATCH/closure-modules-r2`, including `cache/download`.
  - Entries: 14,511 (2,014 directories, 12,497 files).
  - Cutoff: the `st_birthtime` of `binding-v2-build/` = `1794333048955812608` ns. The `mkdir` preceded the build as a separate command; the float rounding is about 256 ns and does not matter.
  - Violations: zero. The latest mtime or ctime in the listing is `1794327320728744155` ns, about 5,728 s before the cutoff.
  - This is timestamp evidence, not a byte-for-byte snapshot of the download files the closure does not enumerate. That limit is stated in `RESULTS.md`.
- **(c) Binary metadata.** `go version -m` ran under `env -i` with a fresh `GOMODCACHE` that was still empty afterwards. Exit 0, empty stderr; stdout `339b679c3aba00a9dfb84ace408ac70b705d9a2c739e49b3745af4c85badce1d`.
  - The binary reports go1.26.10, path `scout/evaluation/x1-triage-ranking/execution/model_free`, and 12 `dep` lines with no replacements. All 12 module/version/`h1:` triples equal `closure.go.module_graph[].verified_sum`.
  - Build settings: `-buildmode=exe -compiler=gc CGO_ENABLED=0 GOARCH=arm64 GOARM64=v8.0 GOOS=darwin`.
  - There is no `vcs.*` stamping. This settles round 1's [INFERENCE] about VCS stamping: none was recorded.
- **(d) Deviations entry.** Present in `RESULTS.md` under `## Deviations` as "Frontend build outside the protected-cache command rule (X1EI-R1-02)", with the outcomes of (a)–(c).
- **No rebuild.** The binary is accepted.

**S5 frontend provenance:**
- Original build argv: `/usr/bin/env -i PATH=/Users/adamjackson/.local/share/mise/installs/go/1.26.10/bin:/usr/bin:/bin HOME=SCRATCH/binding-v2-build/home GOCACHE=SCRATCH/binding-v2-build/gocache GOMODCACHE=SCRATCH/closure-modules-r2 GOTOOLCHAIN=local GOFLAGS=-mod=readonly GOWORK=off GOENV=off CGO_ENABLED=0 GOPROXY=off GOSUMDB=off …/go build -pgo=off -o SCRATCH/binding-v2-build/model-free STUDY/execution/model_free`. Here SCRATCH and STUDY stand for the absolute paths above; the full argv is in `preparation-report.json` `628382180538f13ea0f932f1b5befc76109d56bf1b01050bbe96e67709b04a00`.
- cwd: `/Users/adamjackson/Projects/work/scout-x1prep`; exit 0.
- Go driver: `5742f0004c49c005313d88cd06df964b09cad6bcf2eac4506bdb90abc923b1f4`.
- `main.go`: `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078`.
- Binary: `05fa048c4911666e107e8e272770308cef3b23535b6c0dd208f36f7768a0a139`.
- Metadata output: `339b679c3aba00a9dfb84ace408ac70b705d9a2c739e49b3745af4c85badce1d`.

## Round-1 execution outline, step 3 (`go vet`/`go test`) — waived by this council

The writer did not run `go vet` or `go test` for `model_free`. This council waives that outline step:
- The required fix in X1EI-R1-01 named only the binding suite, `test_binding_contract.py` and the full Python suite. All three passed.
- `main.go` `edb37ec2…` and `main_test.go` `54e7bc66…` are byte-unchanged, and frozen by the closure. Their golden contract tests passed on these bytes before the first `attest_closure.py` journal, as the amendment `:74` and REQUEST `:69` require. Rerunning them would add no evidence about the repair.
- Running them now is not justified given the standing restriction on Go commands against `-r2` (remedy round 1, Step 11).

## Questions

1. **The repair is correct for the real closure and runtime.** No guard bypass and no false refusal was found.
   - A copy or hardlink in an earlier PATH directory resolves to a different path and is refused.
   - A symlink to the attested file passes by design; the sha256 is checked on the attested path.
2. **The tests cover every case round 1 required, each with a passing null.**
   - The guard-deletion analysis found these untested: node resolution, closure-side uniqueness, runtime git equal to closure git, and refusal through `reviewed_inputs:384`.
   - None of these was a required case, and all are verified by reading the code and by the real-data regeneration run. Recorded as suggestion S1.
3. **The manifest and runtime are correct** (above).
4. **Evidence (a)–(d) is complete** (above).
5. **No separate scope is needed for the rehearsal chain.** `acceptance.py` consumes three review scopes:
   - `model-free-binding-v2-amendment` (`:362`), unchanged;
   - `execution-implementation` (`:371`). Its manifest must contain every `binding/**` plus `main.go` source pin (`:373–378`), and the runtime must be bound (`:385`);
   - `model-free-binding-v2-rehearsal` (`:394`). It is issued only after the fresh rehearsal and binds the report, `runtime.json` and the source pins (`:396–413`).

   This record is the binding repair sign-off for `acceptance.py` `6e56fb5f…`, `test_binding.py` `be91c678…` and `test_binding_adoption.py` `985b78b2…`. No code reads a separate repair receipt.
6. **The required rows are exactly 9:**
   - the manifest, `closure.json` and the three object receipts for test-role repositories (`preparation.py:516–529`). fastapi and axum have the `development` role and belong to `resolved-development-boundaries`;
   - the runtime, bound as an exact absolute (path, sha256) pair (`acceptance.py:371–386`);
   - the three I6 files (`closure-weaklink-impl-round-3.md:86`).

## How the peers shaped this record

- **innovation-challenger:**
  - confirmed no guard bypass and that the PATH order is safe;
  - checked every entry of the cache listing and the latest timestamp;
  - confirmed the three consumed scopes.
  - It raised two gaps: (d) not applied, and the Go checks not run. When I read `RESULTS.md` myself, the entry was present, so (d) is closed. The Go step is disposed of above as an explicit waiver. Neither became an issue.
- **experiment-peer:**
  - supplied the falsification table and the guard-deletion table. The four untested guards it found became S1;
  - confirmed the Python 3.11 birth-time rounding is immaterial and that the binary carries no VCS stamping;
  - agreed the Go-check omission was correct under Step 11.
- **Main council:**
  - independent reads of the new `acceptance.py:41–89`, the tests, the jq comparison of runtime and manifest, the evidence sections and `RESULTS.md`;
  - the scope-chain analysis.
- No peer changed the verdict.

## Suggestions (not conditions)

- **S1.** In any later `binding/**` repair, add planted effect and null cases for:
  - `node` shadowing or absence;
  - runtime git differing from closure git;
  - closure-side duplicate go or sandbox-exec rows;
  - a refusal that goes through `reviewed_inputs`.
- **S2.** The guard runs before each operation, not before each `run_revision`. Keep the operator from changing tool directories during a binding session; P3 already forbids such changes.
- **S3.** The accepted residual still stands: the unsandboxed frontend tree can reach `/usr/bin` as a whole. The guard covers only the named tools and is not a universal allowlist of executables.

## Next authorised steps

1. Main writes this record to `STUDY/preparation-reviews/execution-implementation-round-2.md` (UTF-8, LF) and computes its sha256 and byte count.
2. Main recomputes the nine pins below. Every one must equal the digest named in the scope block; any difference means stop and open a new round. Main then writes `OUT/reviews/implementation.json` with exactly these keys:
   - `scope`: `execution-implementation`;
   - `record`: path, sha256 and bytes of this record;
   - `reviewed`: exactly the nine rows below, as absolute `{path, sha256, bytes}`.
3. Freeze. From now until seal, nothing may change in:
   - the nine reviewed files;
   - the `binding/**` source pins and `main.go`;
   - the frontend binary and `SCRATCH/binding-v2-build/superseded-r1/`.

   A further `binding/**` defect follows amendment `:74` again: new zero-issue round, regeneration at fixed paths, fresh rehearsal path.
4. Before each binding launch session: P3-D session checks 1–3. RB-1 stays in force: there must be no further `adopt`.
5. The fresh scratch rehearsal, synthetic only, at an unused `--output` and `.work` path. For example, the REQUEST `:79` command with `--output /Users/adamjackson/Projects/work/x1-preparation-scratch/binding-v2-rehearsal/report.json`, provided that path and its `.work` directory do not exist.
6. A council rehearsal review. On zero issues, Main writes `OUT/reviews/binding-v2-rehearsal.json` with scope `model-free-binding-v2-rehearsal`, binding:
   - the report;
   - `OUT/binding/runtime.json` `647381c5…`;
   - every source pin.
7. Canonical synthetic with `--rehearsal-report`, then Django, then `reviews/binding.json`. Then the remaining reviews, `seal --check` and `seal`.

## Binding repair sign-off (amendment `:74`)

This council signs off the repaired binding implementation with zero issues:
- `execution/binding/acceptance.py` `6e56fb5f6f2f8f1041c2d06498b6f2d01c0e7480a2911def12902094e955de63`;
- `execution/tests/test_binding.py` `be91c678512402244ff27847177046511caa129ac6d8834b079c2add35ccc155`;
- `execution/tests/test_binding_adoption.py` `985b78b2294680c3a564b80b2f92562bda2ceada0419d85d1ecc0c1f398e42aa`.

It also covers the unchanged binding files:
- `interlock.py` `19464e8940d0458be3606ed0438ae778fd01b48b73f8c2fee801856b80b2e6a1`;
- `fixtures.py` `48552a2bd5f8492b83237f147916bde12f2137a6aee6fbae99b01cda9e40381b`;
- `test_binding_seal.py` `7c05eaf2f577210361bbd18a93087e7f59b8fffa687c09d98fad2824d8f4b593`;
- `test_binding_contract.py` `dc817736d39551e01c62acb2e6b3dffeb861de825b4fb56c9b5c3198dc0a9efa`;
- `model_free/main.go` `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078`.

Round-1 issues X1EI-R1-01 and X1EI-R1-02 are resolved. No new issues.

**Zero issues. Signed off.**

Signed-off scope: execution-implementation

- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/execution-source-manifest.json` — 11373 bytes — `25ef2b4e6387f5ed496acdc4d28c187fcf1a38f71fdb2d783bda5bbf78a3d147`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/closure.json` — 61479449 bytes — `b5df46b1b92a3cc9ef2905589db120a008881bc6b5519f37b6d38e3705cc35b2`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/objects/django__django.json` — 3401 bytes — `b10d9c486e36c482833f0684c50c3ca6b2a38f0f711e0747588d136e6e274d50`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/objects/microsoft__TypeScript.json` — 3208 bytes — `d1ef2d34aa953e855e8cc147f280118c64bb23cd157aa3615e169a1184a8b7e2`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/objects/prometheus__prometheus.json` — 3264 bytes — `8135e50e45c7039b237a7cabe617ca0f95a4c98bc9ff169d98d25feeb1355e2a`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/binding/runtime.json` — 4891 bytes — `647381c553c78516fe19fe7d83e7401be1692e6a6922e93f4218606b52ebbc6d`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/attest_closure_v2.py` — 1168 bytes — `c28a9853f826d46e05ddebbd207ac85cc112aec90179a84b7a0800aeecfd6d84`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/closure/native_v2.py` — 15849 bytes — `5e320ed37a5b106e3a37984e6137e78f87422aa525792018032117d66f4dee09`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/tests/test_closure_weak_links.py` — 35588 bytes — `2a7ee81554000968590cfbcd8ce4649a757a346f3ed24f496db37345e6b89c87`
