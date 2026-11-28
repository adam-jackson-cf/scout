# X1 preparation runtime closure attestation (b) implementation review — round 3

**Scope.** This round re-reviews action (b), the read-only runtime closure attester. It checks the round-2 record (`preparation-reviews/closure-impl-round-2.md`) and the author's response (`closure-impl-round-2-response.md`). It also checks whether the closure's dependency on the shared `preparation.py` can cause a concrete failure for the closure run. `binding/acceptance.py` was read only where it consumes `closure.json`.

The review was read-only. Nothing was run: no attester, test, build or model call, and no real input was processed. The only model calls were the two council peers. No session had a shell, so **no SHA-256 below was recomputed by the council**; all digests are Main's `shasum -a 256` values.

## Reviewed files (relative to `evaluation/x1-triage-ranking/`)

| File | SHA-256 (Main) | Note |
|---|---|---|
| `execution/attest_closure.py` | `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892` | under review |
| `execution/closure/native.py` | `2e0f111bfe1462ee3e3b3bc037b5eaf398df7d67478d9f1028a9013029895ecd` | under review |
| `execution/tests/test_closure.py` | `3a1eae6082d0a9f42074e7ea5db8a2d74720808a91c2e2d5770bf44a82e599dd` | under review |
| `execution/tests/test_closure_review.py` | `8762b45d3e1db762b9094dfb174ec79c65ae5efa262a2903a2f3894bf8804743` | under review |
| `execution/tests/test_closure_round2.py` | `7b22a59cbaf43cb0c97381d8ee6faf7fcc6a180c05b3d081ed88f04a4b99cc0c` | under review |
| `execution/preparation.py` (shared) | supplied `e883abc47ea9ea6a693a78ffe46e5793bb1a6b1ccdc61cb17bf2c815ab99db6e` | **The bytes changed during this review** (see "Dependency on `preparation.py`"). This digest is not confirmed as current. |
| `execution/verify_inputs.py` (shared) | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` | unchanged since round 2 |
| `execution/run-config.json` (shared, clock-pinned) | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | unchanged |
| `execution/model_free/main.go` (Go frontend selected by the closure) | `edb37ec2…` | **Deliberately truncated.** It must not be signed off here (Suggestion S9). |
| `execution/binding/acceptance.py` (consumer; context only, not signed here) | `916d3d732f2efbdfa98e1119d9ab5e49401cac7c8e17edc91a404bb48a4a21c1` | |
| Records `closure-impl-round-2.md` / `closure-impl-round-2-response.md` | `2e5983a8…` / `003bc726…` | Main supplied these truncated. |

## Basis (unchanged)

- `reviews/round-9.md` `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` authorises (a)–(e). The clock started 2026-11-09T03:22:58.701909Z.
- `PREREGISTRATION.md` `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb`:
  - §11 item 1, action (b): "bootstrap/attest installed scanner, dependency and numerical-interpreter closure".
  - §11 closing paragraph: "Seal stage/authentic council record/predecessors/…/all source and output hashes"; "Refuse substituted sources, stale versions".
- `reviews/rev9-digests.json` `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`
- `preparation/pins.json` `6d556fb5156e8ced3c17ae7526a841e51688145faad4a3f8a6a2ce0c939100fc`
- `preparation/runtime-attestation-rev6.json` `19463957becfc23a391eac03582b32eaa0ff392069bccd9aed9000f390ad2d62`
- `preparation/model-free-policy.json` `87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f`
- `preparation/model-free-binding-check.json` (v1, signed) `5056ae748e05717e7a1e16dd8974f11519045c01bdc3ce6498adf3b14aedc59c`

## Recommendation

**Do not run the preflight or the canonical closure yet.**

- R2-02, R2-03 and R2-04 are fixed.
- R2-01 is fixed for the reported failure: the real gettext Java classes no longer refuse. But the fix adds a gap where real native code can be skipped without any error (R3-01).
- Fix R3-01, then hold round 4 under the run prerequisites below.
- Those prerequisites include something outside this component. `preparation.py`, which the closure pins permanently, changed while this review was in progress. Its R6-01 fix has not yet been reviewed.

## Prior-round disposition

| ID | Status | Evidence | Would a test fail if the fix were reverted? |
|---|---|---|---|
| R2-01 | **Fixed for the reported failure; the fix adds R3-01** | `native.py:45–48` returns non-native for `CAFEBABE` with a Java-range version word. Fat tables are validated at `:50–77`: 1–32 entries, table inside the file, slices bounded, aligned and not overlapping, valid thin headers, CPU agreement, a unique arm64 slice. Real bytes: the author's dry check classified all 2,493 regular files in `/opt/homebrew/Cellar/gettext/1.0`, giving 32 Mach-O images and 4 non-Mach-O `.class` files (response l.502–783). | Yes, for the Java refusal. `test_r201_java_is_non_native_without_suffix_filter` raised under the old code. Every fixture required by round 2 is present (`test_closure_round2.py:43–73`). There is no fixture for the collision in R3-01. |
| R2-02 | **Fixed** | `selected_source` (`attest_closure.py:566–580`) authorises the logical `Dir/name` before resolving it. For main-module or in-ROOT files it refuses: an unpinned file, any symlink chain, an escape from ROOT, or a digest mismatch. `go_metadata` builds the logical path without resolving it (`:424–426`). | Yes. `test_r202_go_metadata_escaping_file_symlink_refuses` goes through the real `go_metadata` route and plants exactly the `internal/pkg/extra.go` → outside-file link. |
| R2-03 | **Fixed** | Constant path and scope at `:32–33`. `reviewed_go_sources` (`:554–564`) validates through `review_receipt`, accepts only unique `model_free/*.go` rows, and returns pins for the receipt and its council record. `go_metadata` requires the reviewed set to equal the selected frontend set (`:428–431`). Both pins go into `closure.json` `inputs` (`:733`) and into the journal `source_pins` (`:837–838`). `implementation.json` is no longer read. | Yes. The `test_r203_*` tests cover missing, wrong-scope and drifted receipts, a drifted council record, unrelated bytes and the exact-set rule. The journal test checks both pins in `source_pins`, and the producer test checks both in `inputs` (`:381`). |
| R2-04 | **Fixed for items (a)–(f)** | `verify_git_runtime`, `claw_runtime`, `runtime_inventory` and `inventory_native_closure` (`:582–604`) are called by `attest()`. | Yes. The producer test asserts each helper is called once, the keg root is enumerated, and both review pins are in `inputs` (`test_closure_round2.py:379–383`). Planted pass/refuse pairs cover the `object_git` branches, Git version drift and HEAD mismatch, the Clawpatch-before-pnpm order, and native addon selection. Gaps that only fail closed are listed under S4. |

Round-2 suggestions:
- **S1 (preflight retention):** adopted (`:724–725`) but untested (S4).
- **S5:** handled on the binding side. `acceptance.py:45–50` checks `git` against `original_runtime_files`.
- **S2, S3, S4, S6 and S7:** still open as suggestions. S7 is larger than estimated (S6 below).

## New issues

| ID | Severity | Location | Finding (input → wrong outcome) | Required fix |
|---|---|---|---|---|
| R3-01 | minor | `closure/native.py:45–48`; the skip happens at `:165–167` (`if filetype is None: continue`) | **Java check skips structurally valid fat Mach-O images.** A big-endian fat32 header with `nfat_arch` 45–100 reads as Java minor 0 / major 45–100. The function then returns `None` before it applies the count bound at `:51–53` or inspects the table. The challenger built such a file: 45 bounded, 16 KiB-aligned thin slices that do not overlap, exactly one of them arm64. Its constant-pool tag would be 0, so it is not a valid class file, yet `image_filetype` returns `None`. Apple's fat reader accepts up to 204 fat32 entries ([dyld `MachOFile.cpp` l.179–210](https://github.com/apple-oss-distributions/dyld/blob/main/common/MachOFile.cpp#L179-L210)). If such a dylib sits in an enumerated root and an attested executable loads it, the walk skips it, and dependencies it loads from outside the approved roots are never walked. The closure still reports `complete: true`. Round 2's code refused these bytes, so this weakens a safeguard. It is the same kind of planted-file bypass as R2-02, so the severity is the same: minor. | For `CAFEBABE` with a Java-range version word, return non-native only when the fat reading fails a check that dyld itself applies (the fat table or any slice falls outside the file). If the fat reading passes those checks, refuse it as ambiguous. Never return `None` for bytes dyld could load as a fat image. Tests: (i) the 45-entry collision refuses or is walked, never skipped; (ii) Java 52.0 and 45.3 class fixtures still return `None`; (iii) rerun the read-only dry check on `/opt/homebrew/Cellar/gettext/1.0`, which must still classify the four real `.class` files as non-Mach-O. |

## Dependency on `preparation.py`: requested check

**The calls the closure makes.** Each is compatible with both the supplied version and the version currently on disk [INFERENCE: read, not run]:

- `load_config`, `read_clock`, `paths`, `validate_pin`, `now`, `review_receipt`, `operation_lock` and `Journal` (`attest_closure.py:26–27, :556, :827–838`).
- The closure reads only `paths`, `repositories`, `deferred_to_b_contract` and `execution_commit`. Amendments may change only `binding_*` keys (`preparation.py:43–45`), and the proposed amendment 001 changes only those.
- On disk, `adopt` takes an exclusive blocking lock (`:262`), refuses while any operation is unfinished (`:286–287`), and stamps the time itself (`:289`). The closure holds a shared lock for its whole run, and `Journal`/`finish` re-enter it (`:236–254`, `:326–347`). So adoption cannot happen in the middle of (b).
- The seal accounts for `OUT/reviews/closure-model-free.json` through `closure.json` `inputs` (`verify_tree_pins` → `referenced`), so it is not flagged as an unaccounted output.

**The concrete failure is about when (b) runs, not a defect in the closure code.**
- (b) pins `preparation.py`, `verify_inputs.py`, `run-config.json`, `attest_closure.py` and `native.py` into its journal and into `closure.json` (`own_sources`, `:469–471`). `closure.json` also pins `model_free/main.go` and the review receipt and record.
- Those pins are rechecked before every binding scanner run (`verify_installed_closure` `:482–485`; `acceptance.py:82, :341`) and on every journal at seal (`preparation.py` journal loop, `verify_tree_pins`).
- After any adoption, `verify_adoption` also requires the current `preparation.py` to equal the adoption writer (`:89`).
- Existing (a)/(b) journals have empty `source_pins` or do not pin `preparation.py` (`preparation-run/journal/d8414b51….json`, `172d394f….json`). So the first closure journal would be the first record to pin it.

**What was observed during this review:**
- `preparation.py` changed. The first read showed copy-in adoption with a hand-written `adopted_at` that `review_receipt` accepted. Later reads show the R6-01 `adopt`/`verify_adoption` design, `review_receipt` accepting exactly three keys (`:396`), and a recheck of every journal start record in `load_config` (`:148–149`). The experiment-peer saw the file go from 528 to 662 lines.
- `identifiers-impl-round-6.md` (l.41–100, 129–132) leaves R6-01 open against `e883abc4…`. Under `e883abc4…`, a backdated `adopted_at` can make the a/b journals fail permanently at seal.
- `test_core.py:175–182` still writes `adopted_at` into review receipts, which the on-disk `review_receipt` rejects. The shared tree is therefore still being edited, and Main's 164-test result describes earlier bytes [INFERENCE].

**Consequence.** If (b) runs on any `preparation.py` that is later changed, for example by the R6-01 fix, both its journal and `closure.json` become permanently unverifiable. The binding refuses, and the seal refuses. This is prerequisite P1, not a closure defect: binding the code that was executed is correct provenance, and R1-08 accepted it.

## Run prerequisites (to be restated in the zero-issue round)

1. **P1. Shared code is final.**
   - `preparation.py` and `verify_inputs.py` are at digests signed with zero issues in their own review, after R6-01 is closed.
   - Main records those digests, and the full suite has been rerun against them.
   - Nothing the closure pins changes until the seal.
2. **P2. The frontend is final.**
   - `model_free/main.go` has been signed off with zero issues in the binding v2 implementation review, and that record names its digest.
   - `OUT/reviews/closure-model-free.json` has exactly the keys `scope` (`closure-model-free-go`), `record` (a pin of that binding record) and `reviewed`, which is exactly `[pin(<ROOT>/evaluation/x1-triage-ranking/execution/model_free/main.go)]`.
   - `main_test.go` is **not** included: `go list` excludes test files, and the exact-set check would refuse only after the journal has started (S2).
3. **P3. The environment is frozen until the seal.** No `brew`, `mise`, cask or macOS update, so the dyld cache, kegs, Go 1.26.10 and `claude` 2.1.311 / `codex` 0.166.0 stay as they are. Preflight outputs, retained copies and `closure-modules` are kept until the seal.
4. **P4. Run sequence.** First the read-only dry check over every root (S3). Then one journalled preflight, with `--external-root` for every runtime-dependency keg in `/opt/homebrew/Cellar/git/2.56.0_2/INSTALL_RECEIPT.json` (16 kegs, l.52–196, which the challenger confirmed by glob) and the required `--dyld-cache-root`. If that passes, run the canonical command once.

## Suggestions (not required for sign-off)

- **S1. Do the argument checks before the journal.**
  - Move the platform, `--pins` path and module-cache-under-SCRATCH checks (`attest_closure.py:609–616`) before `Journal` (`:837`).
  - Today a typo writes a refused journal, which blocks the seal, or uses up `closure.json`.
- **S2. Refuse `_test.go` in `reviewed_go_sources` before the journal starts.**
  - `reviewed_go_sources` (`:558–563`) currently accepts `model_free/main_test.go`.
  - With that file in the receipt, `go_metadata` refuses only after the journal has started (`:430`).
  - Rejecting `_test.go` names, or checking the exact frontend set with `go list` before the journal, closes the trap. P2 covers it in the meantime.
- **S3. Extend the dry check, and run it first.**
  - Classify every root in one inventory: `.tools`, `sys.base_prefix`, GOROOT, the Go module directories, the git keg, each dependency keg and the seven OS tools.
  - Also run `checked()` over `pins.source.files` for both roots.
  - GOROOT holds raw Mach-O test data (`src/debug/dwarf/testdata/typedef.macho`, `typedef.macho4`, `runtime/race/*.syso`). They have not been classified on real bytes [INFERENCE: expected thin, so they load or fail closed].
- **S4. Close the remaining test gaps. None of them is a fail-open.**
  - Assert `retain`'s destination for `preflight=True` and `False` (the S1 isolation is untested).
  - Assert that `go_metadata` receives `reviewed_go` and the updated `allowed` mapping.
  - Add a fixture for each `image_filetype` refusal branch, plus a fat x86_64 + arm64e null case.
- **S5. A real but malformed Java class refuses.** Java 52.1, for example, is valid under [JVMS §4.1](https://docs.oracle.com/javase/specs/jvms/se25/html/jvms-4.html#jvms-4.1) and currently refuses. That is fail-closed, and the R3-01 fix may also cover it.
- **S6. Recheck cost (round-2 S7, now measured).**
  - Every cached-image record repeats `cache_files` (`native.py:154`).
  - `verify_tree_pins(closure)` in the binding and at the seal re-reads the whole dyld cache family, about 5.5 GB per the peer, once per cached image. It uses `read_bytes` (`verify_inputs.py:25–28`).
  - Store the family once and stream `pin()`.
- **S7. A binding-side ordering trap (cross-packet).**
  - The binding rehearsal requires the canonical `OUT/closure.json` (`acceptance.py:338–343`, called from `rehearse` `:390`).
  - So a `main.go` defect found by the rehearsal forces reviewed recovery of (b).
  - Recommend that the binding review let the synthetic rehearsal accept a passed journalled preflight closure.
- **S8.** Round-2 S2 (re-resolving link chains), S3 (non-host architectures), S4 (inherited `@rpath`) and S6 (cask drift) still apply.
- **S9. Do not use this record as the authority for `closure-model-free-go`.**
  - `review_receipt` only checks that a record contains "zero issues", "signed off" and the reviewed digest. That is why this record truncates `main.go`'s digest.
  - The receipt must cite the binding review that actually reviewed `main.go`.

## Alternatives

| Option | Tradeoff | Decision |
|---|---|---|
| A. Check the fat structure first; treat the bytes as Java only when the fat reading fails dyld's own bounds; refuse when both readings are plausible | A small change; fails closed; the real gettext classes stay non-native | **Recommended** |
| B. Write a complete Java class-file parser | Sound, but adds parser surface and a new place for disagreement | Rejected |
| C. Refuse every `CAFEBABE` whose version word is in the Java range | Brings back R2-01's refusal of the real gettext keg | Rejected |
| D. Exempt by filename or suffix (`.class`) | Brings back R1-10's selection-by-suffix gap | Rejected |
| E. Sign off now and keep R3-01 as a suggestion | Leaves a fail-open in the completeness guarantee; R2-02 set the precedent of treating this kind of gap as an issue | Rejected |

## Evidence

- **Code read:**
  - `native.py:13–77, 90–196` and `attest_closure.py:1–900`, both complete.
  - `test_closure_round2.py:19–383`.
  - `preparation.py` across two snapshots: l.41–56 and 153–442 in the first; l.1–21, 77–152, 236–350 and 394–426 in the current.
  - `acceptance.py:41–58, 77–82, 308–346, 374–407`.
  - `run-config.json:3–9, 57–141`.
  - `internal/sandbox/sandbox.go:77–83` (writes outside the state directory are denied, so scanners cannot add `.pyc` files to the enumerated roots).
- **Repository state:**
  - `pins.json` `source.files` contains no `model_free` paths, and every `internal/*` package used is pinned. `internal/toolchain` contains only tests.
  - `model_free/` contains only `main.go` and `main_test.go`. `main.go` imports only the standard library and pinned `scout/internal/*` packages.
  - Existing journals: five (a) and two (b) `git submodule update --init` pairs, with no `preparation.py` source pin.
  - `identifiers-impl-round-6.md` leaves R6-01 open against `e883abc4…`.
- **Real-byte measurement record (CLAUDE.md experiment rule 3):** the round-2 response's read-only dry check (l.423–809) is a measurement and construct check only. It made no predictor–outcome computation, and it does not show that the native, dyld-cache or Go walks succeed.

**Where peers changed this record:**
- **innovation-challenger:** found the R3-01 collision. I had dismissed it, assuming the count bound refused it, but the Java return comes before the bound. I adopted it and lowered it from major to minor to match R2-02. Its test-gap observations became part of S4. It also saw `preparation.py` change, independently.
- **experiment-peer:** independently saw `preparation.py` change, and its H1 matches my P1. It rated this a "blocking precondition"; I kept it as a prerequisite, not an issue, because the closure behaviour is correct. Its peer-R3-01 and R3-02 test gaps (renumbered here to avoid a clash with R3-01 above) became S4, because they fail closed. Its H7 became S2 and P2, its S-a became S1, its S-c became S3, and its H10 measurements became S6. Its "R2-01 Fixed" rating was overturned by the challenger's construction.

## Risks

- **Irreversibility.** A refused or later-stale preflight journal, or a canonical `closure.json`, blocks the seal. The code has no recovery path, so only reviewed recovery can clear it.
- **Unknown digests.** The council could not recompute any digest. `preparation.py` is changing, and this record's sign-off basis depends on Main's recomputation.
- **Not verified on real bytes** [INFERENCE]:
  - the native load walk over `.tools`, GOROOT and the 16 kegs;
  - offline `go list` / `go mod verify`;
  - `dyld_info` over the cache;
  - whether any root contains fat binaries with both arm64 and arm64e slices, which refuse.
- **45-day window.** The run depends on Homebrew, mise, the casks and the OS staying unchanged (P3).

## Execution outline

1. **R3-01.** Rework the `CAFEBABE` classification in `native.py` (Alternative A). Add the collision, Java 52.0 and Java 45.3 fixtures, plus S4's branch fixtures. Rerun the read-only gettext dry check.
2. **P1.** The core/identifiers review closes R6-01 with zero issues on the final `preparation.py`. Main recomputes the digests of `preparation.py` and `verify_inputs.py`.
3. **P2.** The binding v2 review signs off `main.go`. Main writes the three-key `closure-model-free.json` receipt.
4. Main reruns `.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py'`, `go vet` and `go test`. Main recomputes all component digests and submits round 4 naming them.
5. After a zero-issue round 4: run the S3 dry check, then the journalled preflight (P4), then the canonical run once. Keep the environment frozen until the seal (P3).

## Verification

- **Read only.** Nothing was executed, and no hash was computed.
- **Main's reported checks (not rerun):** 164 tests OK, `go vet` clean, `go test` ok. These predate the `preparation.py` change seen during this review, so they may not describe the current shared bytes.

## Verdict

**1 issue open: R3-01 (minor).** The runtime closure attestation must not be run, either as a preflight or at its canonical paths. No sign-off is given for the digests above. Round 4 must name the final `preparation.py` and `verify_inputs.py` digests (P1) together with the revised closure files.
