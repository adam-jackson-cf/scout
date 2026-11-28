# X1 execution-implementation council — round 4 (binding repair under `:74` and restore review)

**Study:** `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/` (STUDY). OUT = `STUDY/preparation-run`. SCRATCH = `/Users/adamjackson/Projects/work/x1-preparation-scratch`.
**Date:** 2026-11-10.
**Action:** preparation action (c), binding v2.

This round is step 8 of `execution-implementation-round-3.md`. It covers three things:
- **Code and tests.** It reviews the implementation of round 3's C1–C3 and tests T1–T3.
- **Restore.** It reviews the restore R and the regeneration under `preparation/model-free-binding-amendment.md:74`.
- **Authority.** It is the zero-issue round for the `execution/binding/**` repair and issues the `execution-implementation` authority again.

## Inputs

The council did not recompute any SHA-256. The digests are Main's, cross-checked across the artefacts.

| Input | SHA-256 | Bytes |
|---|---|---:|
| Round-3 record `preparation-reviews/execution-implementation-round-3.md` | `c3e2b9febb91d5361f227447e295f6976f3fe07b0e8b2a53f9ada6502e18f28e` | 32831 |
| Writer response `preparation-reviews/execution-implementation-round-3-response.md` | `f620b1a08021ba4e3879af483a87902e9d9c98c5a844aad2247fa20137c2031e` | — |
| Restore record `preparation-reviews/binding-v2-restore-2026-11-10.json` | `1c5a5dce4c6ddb2ec2aea34be391dfa7964f9b7fe96933de0174c982c3c74690` | 14186 |
| Restore script `SCRATCH/binding-v2-restore/restore_pyc.py` | `93901ccc3c9ff0a545b0bf819b831be7e2ca0071c5ebf242761b0b83b5e4e9ed` | 2893 |
| Restore stdout `SCRATCH/binding-v2-restore/restore.json` | `5d1264abfcdc3193c70ca3393b73ed0e2fda338dd7fd34bed512398200f90777` | 1781 |
| Drifted copy `SCRATCH/binding-v2-restore/drifted/9c815f6b…` | `9c815f6bb55ad4e93c1b74a6c729e234f3838ef8067f201a2441c32cfffdabaf` | 12018 |
| Checks artefact `preparation-reviews/execution-implementation-round-3-checks.json` | `60bd2b27a6c51c5ba4c68fe0a1b6f0df1e1a64d67ef799e7262501fefc148db7` | 63272 |
| Regeneration evidence `preparation-reviews/execution-implementation-round-3-regeneration.json` | `fefbf005c85ddbd2022ea48157978e763698db290167856fce32f6f031406ef8` | 6760 |
| `STUDY/RESULTS.md` | `ce5fde8f7aeeab00fbc4a11adcbd192d6cd661d9c459b67238753be368dee5a1` | 15084 |
| `docs/research-ledger.md` (R-094, R-095) | `d5fe8cfb7af21d7b3a52f58eaeb8e797f68781fc75e46ef1054c5dd2821e17e7` | 74594 |
| Superseded manifest `SCRATCH/binding-v2-build/superseded-r2/execution-source-manifest.json` | `25ef2b4e6387f5ed496acdc4d28c187fcf1a38f71fdb2d783bda5bbf78a3d147` | 11373 |
| Superseded receipt `SCRATCH/binding-v2-build/superseded-r2/implementation.json` | `8711ffaecc7662609f6afa75a9e50ba6ff1fcf8f1b9d41312a0639efed11a0d4` | 2581 |

**Method:** three independent read-only analyses: Main council (Opus), `innovation-challenger` (GPT Astra) and `experiment-peer` (Opus). No launch, build, test or file mutation was done.

## C1–C3: the code matches round 3 exactly

- **C1.** `binding/interlock.py:16` is `'#!'+str(python)+' -ISB\n'`. The rest of `stub_source`, `make_stubs` and `evaluate` is unchanged.
- **C2.** `binding/acceptance.py:30` adds `'PYTHONDONTWRITEBYTECODE': '1'` to the `private_env` literal.
- **C3.** `binding/acceptance.py:123–127`:
  - The new `verify_installed_closure(closure_path)` is the first statement after the `with … subprocess.run(…)` block and comes before `attempts=`.
  - It runs on every route, including the positive control's allowed nonzero exit.
  - The pre-launch call (`:115`) is kept.
- **Size evidence.** Only the prescribed additions fit:
  - `interlock.py` grows from 4472 to 4477 B, which is ` -ISB`.
  - `acceptance.py` grows from 29957 to 30043 B, which is 43 B for the indented environment line plus 43 B for the indented check line.
  - This supports "no other behavioural change", but it is not a hash proof. Step 2 below adds the hash proof as a condition.

## T1–T3: each test has a planted effect and a passing null, and each repair is load-bearing

**Tests that would fail if a repair were deleted** (by source reading; neither peer ran a mutation):

| Repair removed | Test that would fail |
|---|---|
| C1 | T1a's null (`test_binding.py:80–84`). T1b's repaired branch would write `.pyc` files under its temp `PYTHONPYCACHEPREFIX`, so `assertFalse(cache.exists())` fails (`:86–111`). |
| C2 | T2's `env['PYTHONDONTWRITEBYTECODE']` assertion (`:136`). T2's subprocess pair separately shows suppression against real `.pyc` creation (`:131–146`). |
| C3 | T3's event list becomes `['verify','launch']`, and the three planted writes no longer raise (`test_binding_boundaries.py:121–164`). |

**Coverage:**
- **T1b** covers both providers. It uses the pre-repair stub line as the planted effect; in both branches the stub exits 97 and logs one valid attempt row.
- **T3:**
  - It builds a real installed-closure receipt with `Inventory`, `retain` and `checked` (`:130–137`), and the spy calls the real `verify_installed_closure` (`:138–141`).
  - It has three nulls: bound exit 0, positive exit 1 and reference exit 0.
  - It has three planted effects, each refusing with `file/link/mode set drift`: bound append, positive append with exit 1 and a valid claude attempt, and reference with an added file.
  - Each case asserts verify → launch → verify, two real checks and one launch.
- **Existing assertions updated:** `test_binding.py:180` and `test_binding_boundaries.py:92`. The key-present branch still asserts no check and no launch.
- **Writes stay in temp directories.** T1b's unflagged planted stub and T2 honour the temp prefix. The repaired stub's `-ISB` suppresses all writes. T3's runtime, receipt and run directories are all under the temp directory.
- **Known limit (S3).** Round 3 put the `-B` guarantee on T1a, because T1b's repaired branch cannot tell `-IS` from `-ISB`: `-I` ignores the prefix variable.

## Restore R: consistent with the round-3 authority

- **The script is verbatim.** `restore_pyc.py` (2893 B) matches round 3's R block line for line, with no CR and no trailing whitespace. It was copied by Main council and checked line by line by the peer.
- **It ran exactly once,** under `scout-x1prep/.tools/venv/bin/python -I -B`, from 18:54:01.237Z to 18:54:01.552Z, exit 0, with no OpenRouter key present.
- **Preconditions held:**
  - the retained object was 12015 B, mode 0444 and `71093d3e…`;
  - the target was 12018 B, mode 0644 and `9c815f6bb55a…`, with mtime 18:32:59.20Z.
- **Postconditions held:**
  - the target is `71093d3e…`, 12015 B and mode 0644;
  - the inode, 441352178, is unchanged, so the write was in place;
  - mtime and ctime are 18:54:01.536Z.
- **The retained object was untouched.** Its ctime is 17:42:34Z.
- **The drifted bytes are kept** read-only on their own inode, 442831735.
- **Order of events:**
  1. P3-D items 1–3 at 18:53:54. All three stdout digests equal O0: sw_vers `46c87c2b…`, sysctl `f9a1754a…`, cache UUID `284cdac9…`. Item 3 used `scout-x1/.tools/venv/bin/python -I -B`, which rule O allows.
  2. The restore.
  3. Round-1 `.checks[0]`, with byte-identical argv, from 18:54:01.6Z to 18:55:32Z, exit 0. Real `verify_installed_closure` took 58.59 s; `verify_tree_pins` passed over 92,066 pins in 31.62 s.
  4. `compare_enumeration.py` `5b23f238…` under `-I -B`: 32 roots, 63,459 records, zero differences.
- **The environment record holds presence flags only;** no secret value is stored.

## Tests, ordering and regeneration

**Suites.** All three ran under `scout-x1prep/.tools/venv/bin/python -I -B`, exited 0, and reported plain `OK` with no skips:

| Suite | Tests |
|---|---:|
| `test_binding*.py` | 65 (was 61; the 4 new methods are T1a, T1b, T2 and T3) |
| `test_binding_contract.py` | 6 |
| Full Python suite | 347 (was 343) |

**Post-suite check.** `.checks[0]` passed again: closure check 63.40 s, 92,066 pins 24.44 s.

**The tested code is the manifested code.** The checks artefact does not pin source digests (S2). However, the writer's session log, read by the peer, shows:
- one patch under `execution/` at 18:56:39.7;
- the suites from 18:56:54.7 to 18:59:37;
- no edit before the regeneration at 19:00:49.

The fresh rehearsal (T4) will bind these exact source pins in its report.

**Regeneration:**
- Round 2's producer was run with byte-identical argv, exit 0.
- Both manifests have 46 rows over the same paths. Exactly four rows differ:
  - `acceptance.py`: `6e56fb5f` → `781a72ec`;
  - `interlock.py`: `19464e89` → `f6841a4b`;
  - `test_binding.py`: `be91c678` → `5f008766`;
  - `test_binding_boundaries.py`: `c5c1d042` → `b8b4b7ea`.
- The superseded manifest and receipt were moved byte-identically to `superseded-r2/`.
- `OUT/reviews/implementation.json` is absent.
- `OUT/binding/runtime.json` is still `647381c5…` (4891 B).

## The 9-row scope

Rows 2–9 equal the round-2 receipt in path, SHA-256 and bytes: the closure, the three test-role object receipts, the runtime and the three I6 files. Only the manifest row changes, from `25ef2b4e…` to `50e8c2b1…` (11373 B in both).

## Deviation D and the ledger

**D as applied** (`RESULTS.md:60–82`) is round-3 D word for word, with the measured values filled in. Two wording changes go beyond the placeholders, and both are **approved by name**:
1. "Round-4 sign-off remains pending; no binding launch is authorised before that review." This was true when it was written.
2. "Sources: R-094–R-095." Ledger rule 3 supports citing the ledger IDs.

After this record is in place, Main replaces sentence 1 with exactly:

> Signed off with zero issues in [execution-implementation round 4](preparation-reviews/execution-implementation-round-4.md).

"Sources: R-094–R-095." stays as written. Rule O is recorded at `RESULTS.md:106`.

**Ledger.** R-094 and R-095 are appended in the 9-column format, with type `docs`, links to the Python documentation and to the round-3 record, and caveats that do not claim independent retrieval. No earlier row was edited.

## Round-3 issues

- **X1EI-R3-01** — Resolved by C1, C2, T1 and T2.
- **X1EI-R3-02** — Resolved by C3 and T3.
- **X1EI-R3-03** — Resolved:
  - the restore R was verified;
  - D is applied;
  - rule O is in force.

No new issues.

## How the peers shaped this record

- **innovation-challenger:**
  - confirmed C1–C3 at the line level;
  - did the guard-deletion analysis for each repair;
  - confirmed that T3 uses the real check and that all writes stay in temp directories;
  - proposed reversing C1–C3 in memory and hashing the result against the round-2 pins. This became condition step 2.
- **experiment-peer:**
  - checked the restore script line by line, and the restore, check and ordering timeline;
  - checked P3-D against O0;
  - did the row-by-row manifest comparison and the 9-row comparison with round 2;
  - established from the session log that the tested code is the manifested code;
  - raised the D wording as a low-severity point. The council disposes of it above by approving both changes by name and fixing the post-sign-off sentence; it is not an issue;
  - measured the real C3 cost (see Risks).
- **Main council:**
  - independent reads of `interlock.py:13–21`, `acceptance.py:26–38,108–137`, both test files, `restore_pyc.py`, `restore.json`, the restore evidence summary, the checks artefact and the regeneration rows, and `RESULTS.md:60–82`.
- No peer changed the verdict.

## Risks

- **Duration.** The closure check measured 58.6–63.4 s, above round 3's 28.4–54.5 s range. C3 therefore adds about 12–13 min per synthetic operation and about 8.5 min per Django operation. `acceptance.py` has no timeout, so this costs time, not correctness.
- **Detection limits.** Write-then-restore between checks remains undetectable, as D discloses. Outer write confinement still requires its own amendment.

## Suggestions (not conditions)

- **S1.** Round-2 S1 is still open: node shadowing or absence, runtime git ≠ closure git, closure-side duplicate support rows, and a refusal through `reviewed_inputs`.
- **S2.** Future suite records should pin the reviewed source digests at the start and end of the run, as closure-weaklink rounds 1–2 did, rather than relying on session logs.
- **S3.** A behavioural test would need a probe outside temp to tell `-IS` from `-ISB`. Keep T1a as the `-B` guard.

## Next authorised steps

Stop at the first failure. Any failure means: retain everything, create nothing more, and open a new council round.

1. **Write this record.** Main writes it to `STUDY/preparation-reviews/execution-implementation-round-4.md` (UTF-8, LF) and computes its sha256 and byte count.
2. **Pre-receipt hash checks.** Read-only; Main records them in a checks file.
   - **Recompute** the 9 scope rows and the binding repair digests below. Each must equal the value in this record.
   - **Reverse in memory, never on disk:**
     - In `interlock.py`, replace the single occurrence of `' -ISB\n'` in the shebang expression with `'\n'`. The SHA-256 must equal `19464e8940d0458be3606ed0438ae778fd01b48b73f8c2fee801856b80b2e6a1` (4472 B).
     - In `acceptance.py`, delete the line `           'PYTHONDONTWRITEBYTECODE': '1',\n`. Also delete the `    verify_installed_closure(closure_path)\n` line that directly follows the line ending `check=False).returncode\n`. The SHA-256 must equal `6e56fb5f6f2f8f1041c2d06498b6f2d01c0e7480a2911def12902094e955de63` (29957 B).
3. **Write the receipt.** Main writes `OUT/reviews/implementation.json` with exactly these keys:
   - `scope`: `execution-implementation`;
   - `record`: path, sha256 and bytes of this record;
   - `reviewed`: exactly the 9 rows below, as absolute `{path, sha256, bytes}`.

   Then confirm it with `review_receipt`.
4. **Update D.** Main replaces D's pending sentence with the approved replacement above.
5. **Freeze until seal.** Nothing may change in:
   - the 9 reviewed files;
   - all `binding/**` source pins and `main.go`;
   - the frontend binary;
   - `SCRATCH/binding-v2-build/superseded-r1/` and `superseded-r2/`;
   - `SCRATCH/binding-v2-restore/`;
   - the failed rehearsal tree, its diagnosis directory and the two `binding-v2-rehearsal-*` records.

   Rule O and RB-1 stay in force: no further `adopt`. A further `binding/**` defect follows `:74` again.
6. **P3-D session checks 1–3.**
7. **Fresh scratch rehearsal (T4).** Synthetic only, with round 2's step-5 argv except `--output /Users/adamjackson/Projects/work/x1-preparation-scratch/binding-v2-rehearsal-2/report.json`. Neither that path nor its `.work` may exist. Expect about 12–13 min longer than the failed rehearsal.
8. **After the rehearsal.** Rerun round-1 `.checks[0]` and `compare_enumeration.py` under `-I -B`. Record the run with argv, exit codes, timestamps and digests. The report must have:
   - all six predicates true;
   - both providers' positive controls present;
   - its `sources` equal to the binding repair digests below plus `main.go`.
9. **Council rehearsal review.** On zero issues, Main writes `OUT/reviews/binding-v2-rehearsal.json` with scope `model-free-binding-v2-rehearsal`, binding:
   - the report;
   - `OUT/binding/runtime.json` `647381c5…`;
   - every source pin.
10. **Then:** canonical synthetic with `--rehearsal-report`, Django, `reviews/binding.json`, the remaining reviews, `seal --check` and `seal`.

## Binding repair sign-off (amendment `:74`)

This council signs off the repaired binding implementation with zero issues:
- `execution/binding/acceptance.py` — 30043 bytes — `781a72ec97fb9b7edf78437aa8aea914de3297c0e62f1de7c1f7e1f51234341f`
- `execution/binding/interlock.py` — 4477 bytes — `f6841a4b68d7bdb14f84efc08e9525b0fb46b57996a9ee92ad38c3fec535c9f0`
- `execution/tests/test_binding.py` — 33253 bytes — `5f008766807c973084f9b760cccc74857a43fcef7924e89061dea89af360f9ee`
- `execution/tests/test_binding_boundaries.py` — 20660 bytes — `b8b4b7ea1a9ca137a7401fd8964f98ec160cca88598b797373b10b2f4b1ab85c`

It also covers the unchanged binding files:
- `execution/binding/fixtures.py` `48552a2bd5f8492b83237f147916bde12f2137a6aee6fbae99b01cda9e40381b`
- `execution/tests/test_binding_adoption.py` `985b78b2294680c3a564b80b2f92562bda2ceada0419d85d1ecc0c1f398e42aa`
- `execution/tests/test_binding_seal.py` `7c05eaf2f577210361bbd18a93087e7f59b8fffa687c09d98fad2824d8f4b593`
- `execution/tests/test_binding_contract.py` `dc817736d39551e01c62acb2e6b3dffeb861de825b4fb56c9b5c3198dc0a9efa`
- `execution/model_free/main.go` `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078`

It accepts the restore evidence:
- the restore record `1c5a5dce4c6ddb2ec2aea34be391dfa7964f9b7fe96933de0174c982c3c74690`;
- `restore_pyc.py` `93901ccc3c9ff0a545b0bf819b831be7e2ca0071c5ebf242761b0b83b5e4e9ed`;
- `restore.json` `5d1264abfcdc3193c70ca3393b73ed0e2fda338dd7fd34bed512398200f90777`;
- the restored installed file `71093d3edaad0475b353d8b8515d0888fe84fbe74f0fac3080a907dbc5b01d12`.

Round-3 issues X1EI-R3-01, X1EI-R3-02 and X1EI-R3-03 are resolved. No new issues.

**Zero issues. Signed off.**

Signed-off scope: execution-implementation

- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/execution-source-manifest.json` — 11373 bytes — `50e8c2b16907513f10cefd74232e93166b2363ca1375e37e080c85a17058a4ce`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/closure.json` — 61479449 bytes — `b5df46b1b92a3cc9ef2905589db120a008881bc6b5519f37b6d38e3705cc35b2`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/objects/django__django.json` — 3401 bytes — `b10d9c486e36c482833f0684c50c3ca6b2a38f0f711e0747588d136e6e274d50`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/objects/microsoft__TypeScript.json` — 3208 bytes — `d1ef2d34aa953e855e8cc147f280118c64bb23cd157aa3615e169a1184a8b7e2`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/objects/prometheus__prometheus.json` — 3264 bytes — `8135e50e45c7039b237a7cabe617ca0f95a4c98bc9ff169d98d25feeb1355e2a`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/binding/runtime.json` — 4891 bytes — `647381c553c78516fe19fe7d83e7401be1692e6a6922e93f4218606b52ebbc6d`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/attest_closure_v2.py` — 1168 bytes — `c28a9853f826d46e05ddebbd207ac85cc112aec90179a84b7a0800aeecfd6d84`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/closure/native_v2.py` — 15849 bytes — `5e320ed37a5b106e3a37984e6137e78f87422aa525792018032117d66f4dee09`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/tests/test_closure_weak_links.py` — 35588 bytes — `2a7ee81554000968590cfbcd8ce4649a757a346f3ed24f496db37345e6b89c87`
