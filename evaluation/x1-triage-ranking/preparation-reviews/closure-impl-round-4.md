# X1 preparation runtime closure attestation (b) implementation review — round 4

**Scope.** This round checks only two things:

- the fix for R3-01 and what that fix can affect;
- whether the shared `preparation.py`, now at `599a128d…` with the `adopt` command and journal lock, can cause a concrete failure in the closure run.

Code from earlier rounds that has been cleared and has not changed was not audited again.

**How the review was done.**

- Nothing was executed: no attester, test, build, hash or model call, and no real input was processed. Model calls were limited to the two council peers, `innovation-challenger` and `experiment-peer`, both read-only.
- I read one public source: dyld's `mach_o/Universal.h`.
- Every digest below is one of Main's `shasum -a 256` values from `local://x1-packet-digests.md`. The council recomputed none of them.

## Reviewed files (paths relative to `evaluation/x1-triage-ranking/`)

| File | SHA-256 (Main) | Role this round |
|---|---|---|
| `execution/closure/native.py` | `c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb` | changed since round 3 (was `2e0f111b…`); under review |
| `execution/tests/test_closure_round3.py` | `f6eec06e95a7227a02ffcd7ded9598957cffe2f840fe7976bde7e9e1e9e599e8` | new; under review |
| `execution/attest_closure.py` | `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892` | unchanged; read for call sites into `native.py` and `preparation.py` |
| `execution/tests/test_closure.py` | `3a1eae6082d0a9f42074e7ea5db8a2d74720808a91c2e2d5770bf44a82e599dd` | unchanged |
| `execution/tests/test_closure_review.py` | `8762b45d3e1db762b9094dfb174ec79c65ae5efa262a2903a2f3894bf8804743` | unchanged |
| `execution/tests/test_closure_round2.py` | `7b22a59cbaf43cb0c97381d8ee6faf7fcc6a180c05b3d081ed88f04a4b99cc0c` | unchanged |
| `execution/preparation.py` (shared) | `599a128d64dbd0617e7110daddf77c3b1b8677e516e4f0c9b3cf6a1ea389e427` | dependency check only (see below); this is **not** a Core sign-off |
| `execution/verify_inputs.py` (shared) | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` | unchanged; dependency only |
| `execution/run-config.json` (clock-pinned) | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | unchanged |
| `preparation-reviews/closure-impl-round-3.md` | `32b4b78e0f73887e83c0a1cf1e33c5efc7c5cbf3805d01b8b96f11637adc9ab3` | prior record |
| `preparation-reviews/closure-impl-round-3-response.md` | `f82335907b2efecfcb0241da0bd9006ffc4c17527cbfe067f31774798da9c982` | author response |
| `execution/binding/acceptance.py` | `61a95ede2c423c50b1ae0df130160608f72632484f41f7705f3e46f6cceb6b9b` | context only, not signed here. How it reads `closure.json` is unchanged: `:41–58`, `:81–82`, `:347–352`. |
| `execution/model_free/main.go` | `edb37ec2…` | **truncated on purpose; not signed here** (round-3 S9) |

## Basis (unchanged)

- `reviews/round-9.md` `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` authorises actions (a)–(e). The clock started 2026-11-09T03:22:58.701909Z.
- `PREREGISTRATION.md` `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb`:
  - §11 item 1(b): "bootstrap/attest installed scanner, dependency and numerical-interpreter closure".
  - §11 closing: "Seal … all source and output hashes"; "Refuse substituted sources, stale versions".
- `reviews/rev9-digests.json` `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`
- `preparation/pins.json` `6d556fb5156e8ced3c17ae7526a841e51688145faad4a3f8a6a2ce0c939100fc`
- `preparation/model-free-policy.json` `87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f`
- `preparation/model-free-binding-check.json` (v1) `5056ae748e05717e7a1e16dd8974f11519045c01bdc3ce6498adf3b14aedc59c`
- `preparation/runtime-attestation-rev6.json` `19463957becfc23a391eac03582b32eaa0ff392069bccd9aed9000f390ad2d62`
- Exposure recipe `c065347e…`, exposure sources `67dc5a4b…`, frozen exposure inputs `ec92fba6…`

## Recommendation

**Accept the round-4 closure implementation with zero issues.**

- R3-01 is fixed exactly as round 3 required.
- `preparation.py` `599a128d…` causes no concrete failure in the closure run.
- **The closure must not run yet.** Two prerequisites are not met today:
  - **P1:** `identifiers-impl-round-7.md` (`6d279752…`) still has R7-01 open against `599a128d…`.
  - **P2:** `binding-v2-round-1.md` (`91105b45…`) has 5 issues open, and `main.go` `edb37ec2…` is not signed.
- Run the closure only once P1–P4 below hold.

## Prior-round disposition

| ID | Status | Evidence | Would a test fail if the fix were reverted? |
|---|---|---|---|
| R3-01 (minor) | **Fixed** | See the breakdown after this table. | **Yes.** Two tests fail on revert; two more guard against later changes. See the test list after this table. |
| R2-01…R2-04 | Fixed in round 3; not reopened | `attest_closure.py` is unchanged at `113d2d72…`. | Not applicable |

**R3-01 evidence.** `native.py:43–55`: when `CAFEBABE` is followed by a Java-range major version (the low 16 bits, 45–100), `image_filetype` returns `None` in only two cases:

- the declared fat32 table runs past the end of the file (`:48–49`);
- one of the declared slices runs past the end of the file (`:50–54`). This uses exact Python integer comparisons, so there is no overflow.

In every other case it raises `ambiguous Java/fat Mach-O image` (`:55`). The stricter checks for slice count, alignment, thin header and CPU (`:58–85`) are never used as evidence that a file is Java.

Across the whole function, `None` now comes from only three places: unknown magic (`:19–20`) and the two Java exits. So nothing that dyld can load can reach the skip at `:173–174`. Both Java exits correspond to errors dyld itself raises: "slice extends beyond end of file", or a table that runs past the end of the file with page-aligned slices (peer analysis, [INFERENCE]).

The other paths also hold:

- Fat64 magic (`CAFEBABF`/`BFBAFECA`) and byte-swapped fat32 (`BEBAFECA`) never enter the Java branch.
- A non-zero minor version (for example `0x00010034`) does enter the branch, but the full declared count is still checked against the file.
- Equality at the end of the file counts as in bounds, so it refuses.

**R3-01 tests** (`test_closure_round3.py`):

- `:43–48`: the 45-slice collision must refuse; the old code returned `None`, so this test fails on revert.
- `:65–68`: a bounded, zero-filled 52-entry table must refuse; this also fails on revert.
- `:59–63`: only the last entry escapes the file. This stops a regression to checking only the table or only the first entry.
- `:50–57`: Java 52.0, 45.3 and 52.1 return `None`. This also covers round-3 S5.

**Round-3 suggestions.**

- **S5:** adopted.
- **S4:** the classifier part is adopted (`test_closure_round3.py:70–118`). Retention-destination tests and the `go_metadata` argument assertions are still open.
- **S1, S2, S3, S6, S7, S8, S9:** still open as suggestions. None of them blocks this sign-off.

## Dependency on `preparation.py` 599a128d: no concrete failure

**Locking.**

- `attest_closure.py:836` takes a shared, non-blocking lock.
- Inside it, `Journal.__init__` (`preparation.py:326`) and `finish` (`:342`) re-enter through the held-key path at `:239–244`. That path refuses only an exclusive upgrade, which the closure never asks for.
- `adopt` needs an exclusive, blocking lock (`:262`), so it waits for the closure to finish.
- If `adopt` or `seal` already holds the exclusive lock when the closure starts, the non-blocking acquire fails before any journal is written. That is a clean refusal that leaves no record behind.

**An adoption between the closure's config load and its lock.**

- `Journal` reloads the effective configuration and compares it with the one loaded earlier (`:328–329`). This happens before the start record is written and before the `try` at `attest_closure.py:842`.
- So a mismatch raises with no start record and no output: no reviewed recovery is needed.

**Journal checks when the config loads, and at the seal.**

- `load_config` (`:148–149`) and the seal run `verify_journal_configuration` (`:409–426`) over every start record.
- All seven existing start records lack `effective_configuration`: grep over `preparation-run/journal/` found no match. They pass for as long as no amendment was adopted before they started, and `adopt` enforces that by giving `adopted_at` a later time than every existing journal (`:289–294`).
- A closure journal started before 001 is adopted expects an empty chain and the raw `run-config.json` digest. Adoption copies the amendment in and does not modify `run-config.json`. A closure journal started after adoption expects the chain it recorded. Both orders pass.

**Receipt schema.** `review_receipt` requires exactly `scope`, `record` and `reviewed` (`:396`). That matches what `reviewed_go_sources` reads (`attest_closure.py:554–564`).

**Config keys.**

- Amendments can change only `binding_*` keys (`:42–45`). The closure reads `paths`, `repositories`, `deferred_to_b_contract` and `execution_commit`.
- `adopt` writes nothing that `own_sources` (`attest_closure.py:469–471`) pins.

**Freeze.** The closure pins `preparation.py` into its journal and into `closure.json`, and `verify_adoption:89` requires the adopt writer to match the current `preparation.py`. Any later change to `preparation.py` therefore invalidates both the closure and the adoption. This is prerequisite P1, not a defect.

## New issues

None.

## Suggestions (not required for sign-off)

**S10. Fixtures at the Java/fat boundary.** None of the following exposes a defect today.

- Add a Java class longer than `8+20*major` bytes with minor version 0 and realistic constant-pool bytes. Today only the zero-filled table fixture reaches the per-entry loop at `:50–54`.
- Add an entry whose offset is in bounds but whose size overflows.
- Add a slice that ends exactly at the end of the file.
- Add declared counts of 44, 45, 100 and 101.

**S11. Make the all-roots dry check (P4/S3) show Java-branch decisions.**

- For every `CAFEBABE` file, print the version word and which exit it took: table bound, slice escape or refusal.
- The gettext evidence is weaker than round 3 assumed. The four classes are 449, 510, 449 and 1,125 bytes (response stdout). Three are smaller than the smallest Java-range table (908 B), so they leave at `:48–49` regardless of their version. At most `Test15.class` reached the per-entry loop.
- Every enumerated file is classified upfront (`collect`'s `pending.sort`). So a refusal anywhere in `.tools`, `sys.base_prefix`, GOROOT, the module directories, the git keg or the 16 dependency kegs would stop the preflight journal. The read-only dry check is the only place to find these without leaving a record.
- In the pinned Go 1.26.10, the Mach-O test data under `src/debug/macho/testdata/` is stored as `.base64` (found by glob). Raw `CAFEBABE` files elsewhere in the roots have not been checked.

**S12. Correct the round-3 evidence.**

- The round-3 R3-01 row cites dyld's limit of 204 fat entries. That is the old `FatFile` reader.
- The current upstream [`mach_o/Universal.h`](https://github.com/apple-oss-distributions/dyld/blob/main/mach_o/Universal.h) declares `enum { kMaxSliceCount = 16 }`. So the 45-slice construction is probably not loadable by the current dyld; which reader the installed OS uses was not checked.
- The fix is still correct and safe: it never returns `None` when every declared slice is in bounds. The attester's normal path also accepts counts 17–32. That over-includes files rather than omitting them.

**S13. Run discipline (falls under P3).** If the wall clock steps backwards after `adopt`, a later journal's `started_at` could fall before `adopted_at`. Every later `load_config` would then refuse. Keep the host clock steady until the seal.

**Round-3 S1, S2, S3, S6, S7, S8 and S9 still apply.**

## Run prerequisites (restated; all four must hold)

**P1. Shared code is final.**

- Today `identifiers-impl-round-7.md` leaves R7-01 open against `599a128d…`. Its preferred fix, option (a), changes only `binding/acceptance.py`.
- P1 is met only when Core/identifiers is signed off with zero issues at `preparation.py` `599a128d…` and `verify_inputs.py` `7d6f89ce…`.
- If the Core sign-off lands on any other digest of either file, a short closure round confirming this dependency check must pass before (b) runs.
- Nothing that the closure pins may change until the seal.

**P2. The frontend is final.**

- The binding v2 implementation review must sign off `model_free/main.go` with zero issues, naming its digest. Today 5 issues are open and the digest is `edb37ec2…`.
- `OUT/reviews/closure-model-free.json` must contain exactly `scope` = `closure-model-free-go`, `record` (a pin of that binding record) and `reviewed` = `[pin(main.go)]`, and not `main_test.go`.
- This record is not the authority for `main.go`, `preparation.py` or `verify_inputs.py`.

**P3. The environment is frozen until the seal.**

- No change to the OS, the dyld cache, Homebrew/mise kegs, casks, Go 1.26.10, `claude` 2.1.311 or `codex` 0.166.0.
- Keep the host clock steady.
- Keep preflight outputs, retained copies and `closure-modules`.

**P4. Run sequence.**

1. Run the read-only dry check with `--external-root` set to every root: `.tools`, `sys.base_prefix`, GOROOT, the Go module directories, the git keg and the 16 kegs from `/opt/homebrew/Cellar/git/2.56.0_2/INSTALL_RECEIPT.json`.
2. Run one journalled `--preflight` with those external roots and the active `--dyld-cache-root`.
3. Only if the preflight passes, run the canonical command once, writing `OUT/closure.json`.

## Alternatives

| Option | Tradeoff | Decision |
|---|---|---|
| A. Sign off now with zero issues, with P1–P4 as hard prerequisites | No closure defect is left open; the run still waits on the Core and binding tracks | **Recommended** |
| B. Withhold sign-off until P1 and P2 close | Adds another round for code that has no defects, and the review time comes out of the 45-day clock; P1–P4 already gate the run | Rejected |
| C. Add a full Java class parser so that ambiguous files are classified instead of refused | More parser code and more room for disagreement. Worth it only if a real required Java file is refused (S11 will show this) | Rejected for now |
| D. Lower the fat count limit to match dyld's 16 | Changes cleared code without any concrete failure, since over-inclusion fails closed | Rejected; noted under S12 |

## Evidence

**Code read:**

- `native.py:13–98` and `:103–212`
- `test_closure_round3.py:1–122`
- `attest_closure.py:408–452`, `:554–604`, `:752–859`
- `preparation.py:22–175`, `:236–350`, `:394–428`, `:570–636`
- `acceptance.py` (the parts that read `closure.json`)
- `run-config.json:1–15`

**Repository state:**

- `preparation-run/journal/` holds 7 start/end pairs. Grep finds no `effective_configuration` in them.
- `identifiers-impl-round-7.md` `:151–156`: 1 issue open (R7-01) against `599a128d…`.
- `binding-v2-round-1.md` `:7`, `:37`, `:272`: 5 issues open; `main.go` is `edb37ec2…`.
- Glob of `~/.local/share/mise/installs/go/1.26.10/src/debug/macho/testdata/`: the Mach-O test data is stored as `.base64`.
- A binary byte-pattern grep cannot detect `CAFEBABE` files: it also missed the known gettext classes. So the dry check (S11) is still required.

**External source:** dyld `mach_o/Universal.h` (main) declares `kMaxSliceCount = 16`.

**How the peers changed this record:**

- **innovation-challenger** agreed that R3-01 is fixed and found no fail-open. It flagged that the 204-entry citation is outdated (S12, which I checked against the source) and proposed the boundary fixtures in S10.
- **experiment-peer** independently found no issues, and its H2, H3 and H4 match my analysis. It showed that the gettext dry check barely reaches the per-entry loop, which became S11. It also found the backward clock-step trap (S13) and showed that an adoption racing the closure leaves no journal.
- My own reading added the current P1/P2 status from the identifiers round 7 and binding v2 round 1 records, and the GOROOT `.base64` observation.
- Neither peer changed the verdict.

## Risks

- **Records cannot be undone in code.** A refused or stale preflight journal, or a canonical `closure.json`, blocks the seal until there is a reviewed recovery. S11's dry check is the only check that leaves no record.
- **Not verified on real bytes** [INFERENCE]:
  - the full native load graph;
  - `dyld_info` over the cache;
  - offline `go list` and `go mod verify` under the closure;
  - fat files that contain both arm64 and arm64e slices, or no arm64 slice (these refuse);
  - Java classes outside gettext.
- **Shared code may still change.** If R7-01 is fixed in `preparation.py`, which is option (b) of round 7, this dependency sign-off lapses (P1).
- **The 45-day window.** The run relies on the environment staying frozen (P3).

## Execution outline

1. Close R7-01, preferably in `binding/acceptance.py`, and get a zero-issue Core/identifiers round. Main confirms `preparation.py` is `599a128d…` and `verify_inputs.py` is `7d6f89ce…`. If either differs, hold a short closure round.
2. Get a zero-issue binding v2 sign-off naming the `main.go` digest. Main writes the three-key `closure-model-free.json`.
3. Main reruns `.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py'`, `go vet ./evaluation/x1-triage-ranking/execution/...` and `go test -count=1 ./evaluation/x1-triage-ranking/execution/...`, then recomputes the digests in the table above. If any closure file has changed, a new round is needed.
4. Run the all-roots read-only dry check (P4 step 1, with S11 output if adopted). Every root must classify without refusal.
5. Run the journalled preflight once. If it passes, run the canonical closure once at `OUT/closure.json`. Keep the environment frozen until the seal.

## Verification

- The council ran nothing: it reviewed only, and computed no hashes.
- Main reported, but the council did not rerun: 198 Python tests OK, `go vet` clean, `go test` ok, all at the bytes in `local://x1-packet-digests.md`.
- The author reported, but the council did not rerun: the read-only gettext dry check with `-I -B` passed, classifying 2,493 files as 32 Mach-O images and 4 Java classes. `preparation.py` was `599a128d…` before and after that run.

## Verdict

**zero issues — closure attestation may be run at its canonical paths under the stated prerequisites**

**Signed off.** The innovation council signs off the runtime closure attestation (b) implementation at these digests:

- `execution/attest_closure.py` `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892`
- `execution/closure/native.py` `c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb`
- `execution/tests/test_closure.py` `3a1eae6082d0a9f42074e7ea5db8a2d74720808a91c2e2d5770bf44a82e599dd`
- `execution/tests/test_closure_review.py` `8762b45d3e1db762b9094dfb174ec79c65ae5efa262a2903a2f3894bf8804743`
- `execution/tests/test_closure_round2.py` `7b22a59cbaf43cb0c97381d8ee6faf7fcc6a180c05b3d081ed88f04a4b99cc0c`
- `execution/tests/test_closure_round3.py` `f6eec06e95a7227a02ffcd7ded9598957cffe2f840fe7976bde7e9e1e9e599e8`
- `execution/run-config.json` `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`

The prior-round records reviewed are `closure-impl-round-3.md` `32b4b78e0f73887e83c0a1cf1e33c5efc7c5cbf3805d01b8b96f11637adc9ab3` and `closure-impl-round-3-response.md` `f82335907b2efecfcb0241da0bd9006ffc4c17527cbfe067f31774798da9c982`.

The closure was checked against shared `execution/preparation.py` `599a128d64dbd0617e7110daddf77c3b1b8677e516e4f0c9b3cf6a1ea389e427` and `execution/verify_inputs.py` `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`. That was a dependency check only: this record is not a Core sign-off, and it must not be cited as a review receipt for those two files or for `model_free/main.go`.

The run is permitted only when P1–P4 all hold. P1 and P2 are not met today.
