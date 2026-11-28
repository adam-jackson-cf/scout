# X1 preparation runtime closure attestation (b) implementation review — round 5 (dependency)

**Scope.** This round checks two things only:
- Can the shared `execution/preparation.py`, now at `b1759ec7…`, cause a concrete failure in the closure's preflight run, canonical run, journals, or the seal of closure outputs?
- Has anything else that the closure pins or reads changed?

The closure files themselves are unchanged since round 4 and were not audited again.

**Method.** I only read files. I ran no attester, test, build, hash, adopt or seal, made no model calls beyond the two council peers (`innovation-challenger`, `experiment-peer`), and processed no real inputs. Every digest below comes from Main's `shasum -a 256` list in `local://x1-packet-digests.md`, except where a row says otherwise.

## Files and SHA-256 (paths relative to `evaluation/x1-triage-ranking/`)

| File | SHA-256 | Role this round |
|---|---|---|
| `execution/preparation.py` (shared) | `b1759ec710fa4b71dec1ef91800e84d285c733693b0cb5f4ffd7fe68f147c916` | **Changed** (was `599a128d…`). Dependency check only; this is not a Core sign-off. |
| `execution/tests/test_core.py` | `4f4bfdaf50fc0d56ff775f8c86bb6bc579c932acc1a76764a55ba4a6f36da07e` | Read at `:78–127` for the new accounting tests |
| `execution/binding/acceptance.py` | `2144999a8aece522e72a86a429a1628a2833ada1490fa6e0fe7600e152373be1` | Changed (was `61a95ede…`). Context only: checked how it reads `closure.json`. Not signed here. |
| `execution/attest_closure.py` | `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892` | Unchanged |
| `execution/closure/native.py` | `c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb` | Unchanged |
| `execution/tests/test_closure.py` | `3a1eae6082d0a9f42074e7ea5db8a2d74720808a91c2e2d5770bf44a82e599dd` | Unchanged |
| `execution/tests/test_closure_review.py` | `8762b45d3e1db762b9094dfb174ec79c65ae5efa262a2903a2f3894bf8804743` | Unchanged |
| `execution/tests/test_closure_round2.py` | `7b22a59cbaf43cb0c97381d8ee6faf7fcc6a180c05b3d081ed88f04a4b99cc0c` | Unchanged |
| `execution/tests/test_closure_round3.py` | `f6eec06e95a7227a02ffcd7ded9598957cffe2f840fe7976bde7e9e1e9e599e8` | Unchanged |
| `execution/verify_inputs.py` (shared) | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` | Unchanged; dependency only |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | Unchanged |
| `execution/model_free/main.go` | `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078` | Unchanged. Not signed here (P2). |
| `preparation-reviews/closure-impl-round-4.md` | `5e9fcd1bcd1a28ae3487505df11d1bb95c378d6c2d1a33a059b85a9ffd349627` | Prior record. Digest as cited at `binding-v2-round-3-response.md:168`; not in Main's list. |
| `preparation-reviews/binding-v2-round-3-response.md` | `44a391c3b8d7a1aff2c0743fca2af2e642356e675a7bb9e15caa3b41d2cfea0a` | Read `:60–66` for where the Core change came from, and for P2 status |

## Basis (unchanged)

- `reviews/round-9.md` `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` authorises preparation (a)–(e). The clock started 2026-11-09T03:22:58.701909Z.
- `PREREGISTRATION.md` `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb`:
  - §11 item 1(b);
  - §11 closing: "Seal … all source and output hashes"; "Refuse substituted sources, stale versions".
- `rev9-digests.json` `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`
- `pins.json` `6d556fb5…`
- `model-free-policy.json` `87de3949…`
- `model-free-binding-check.json` v1 `5056ae74…`
- `runtime-attestation-rev6.json` `19463957…`
- Exposure recipe `c065347e…`, exposure sources `67dc5a4b…`, frozen exposure inputs `ec92fba6…`

## Recommendation

**Accept the closure's dependency on `preparation.py` `b1759ec7…` with zero issues.** This is the short closure round that round-4 P1 requires when Core moves off `599a128d…`.

The closure still must not run. P1 needs a zero-issue Core sign-off at `b1759ec7…`, and P2 is still unmet.

## Prior-round disposition

- **Round 4:** zero issues; nothing to reopen. The closure implementation files are byte-identical to round 4.
- **Round-4 suggestions:** S10–S13 and round-3 S1–S4 and S6–S9 are still open as suggestions. S11, the all-roots dry check that shows Java-branch decisions, is still strongly recommended before any journalled run.

## Dependency on `preparation.py` b1759ec7: no concrete failure

**1. What the closure writes**
- One journal start record and one end record.
- The end record carries `output_pins=[pin(args.output)]` (`attest_closure.py:845`; the same on refusal at `:854`). For preflight, `args.output` is `SCRATCH/closure-preflight/<journal.id>.json` (`:840–841`). For the canonical run it is `OUT/closure.json` (`:828`).
- Retained copies, command records and the module cache all go to SCRATCH (`:724–726`).
- `run-config.json` places SCRATCH (`/Users/adamjackson/Projects/work/x1-preparation-scratch`) outside OUT and EXEC. So under OUT the closure writes only `closure.json`, its journal pair and `.operation.lock`.

**2. Change (1): terminal-journal output pins are now accounted at seal**
- `collect_seal` gathers every end record's `output_pins` (`preparation.py:564`).
- `verify_output_accounting` (`:444–458`) re-verifies each pin and adds only paths not already referenced, with role `journal-output`.
- **Preflight receipt:** its path is unique per journal ID and lives under SCRATCH, so it cannot collide with `OUT/closure.json`. It enters the seal outputs once as `journal-output` (`:448–451`).
  - The seal does not chmod it, because only OUT and EXEC paths are frozen (`:612`).
  - It is not added to the review-binding set, which is built earlier at `:544`.
  - It adds no new precondition: `verify_tree_pins(value)` at `:563` already re-hashed every pin in each end record. Keeping the receipt byte-stable is already part of P3.
- **Unaccounted-output scan:** the scan at `:453–457` covers OUT only. Every file the closure writes under OUT is referenced: `closure.json` as a required record (`:494`), the journal pair (`:566`), and the lock is excluded (`:456`).

**3. Change (2): canonical roles are kept when a journal also pins the path**
- `referenced` starts from `sources + outputs` (`:445`, called at `:572`).
- `OUT/closure.json` is already in `outputs` with role `scout.x1.installed-closure.v1` (`:494`), so the journal's pin of it is verified but not appended (`:449`). The canonical role is neither replaced nor duplicated.
- `required_reviewed` (`:544`) still requires a scoped review that binds `closure.json`.
- A journal pin with a stale digest for the same path still refuses with `pin mismatch`.
- Tests:
  - `test_core.py:89–98`: the role is preserved.
  - `test_core.py:100–109`: duplicates are deduplicated, and every pin is still verified.
  - `test_core.py:119–127`: a tampered journal-pinned output is refused.

**4. Behaviour round 4 relied on is unchanged.** Every round-4 anchor is on the same line with the same text:
- re-entrant held-lock path `:239–244`;
- `adopt` takes an exclusive, blocking lock `:262`;
- adoption time must be later than every journal `:289–294`;
- `Journal` reload and compare before the start record `:326–329`;
- `finish` lock `:342`;
- `load_config` start-record loop `:148–149`;
- three-key `review_receipt` `:396`;
- `verify_journal_configuration` `:409–427`;
- amendments may change only `binding_*` keys `:42–45`.

The changes are confined to `verify_output_accounting` and the journal loop in `collect_seal`. I did not byte-diff the two versions, so "unchanged" is an [INFERENCE] from positions; Main reports Core tests pass at `4f4bfdaf…`.

**5. Existing journals**
- `preparation-run/journal/` now holds 8 pairs, not the 7 that round 4 counted. The extra pair is the action-d journal `906389d7…`, which records `effective_configuration` with no adopted amendments.
- Both existing generic (b) journals (`3f969faf…`, `d8414b51…`) have `output_pins: []` and `source_pins: []`. They interact with no closure output.

**6. `acceptance.py` 2144999a**
- It reads `closure.json` at the same lines and with the same fields round 4 cleared: `:41–58`, `:81–82` (`verify_installed_closure` on every launch), and `:347–352` (passed and complete, `verify_tree_pins`, `runtime_closure` pin).
- The closure's `own_sources` does not pin `acceptance.py`, so this change does not affect the closure.

**7. Freeze still applies**
- The closure pins `preparation.py` into its journal `source_pins` and into `closure.json`.
- `verify_installed_closure` re-validates those pins before every binding launch and at the seal (`attest_closure.py:482–485`).
- So `b1759ec7…` must stay fixed from the first `attest_closure.py` journal until the seal. This is part of P1, not a defect.

## New issues

None.

## Suggestions (not required for sign-off)

- **S14.** Add a Core test in which a journal pins a file outside OUT, matching the preflight receipt. It should be accounted as `journal-output`, be refused if tampered, and not be chmodded by the seal. Also add a closure-shaped fixture that combines canonical `closure.json`, a duplicate journal pin of it, and an external preflight receipt. Today `test_core.py:78–127` uses only paths inside `out`.
- **S15. Wording.** In the sealed record, a `journal-output` row for the preflight receipt means "hash-recorded", not "frozen". The retained copies listed inside the preflight receipt are not re-checked at the seal, so the P3 retention duty stays as written.
- **S16. Run discipline (P4).** Start no binding run until the canonical closure's end record exists. `exclusive_json` does not write atomically, and a binding run that read a half-written `closure.json` would leave a refused receipt that cannot be replaced. This predates the change.
- **S17.** Main should confirm, with a hash comparison only, that `906389d7…`'s recorded `effective_configuration.sha256` equals `config_digest` of `run-config.json` `c817f678…`. If they differed, every `load_config` would refuse, the closure's included. This check is unaffected by `b1759ec7…` and was not computed here.

## Run prerequisites (restated; all four must hold)

**P1. Shared code is final — not met.**
- This round confirms the closure dependency at `preparation.py` `b1759ec7…` and `verify_inputs.py` `7d6f89ce…`.
- The last zero-issue Core sign-off, `identifiers-impl-round-8.md:162–165`, was at `599a128d…`. A zero-issue Core review at `b1759ec7…` is still needed.
- If Core signs off on any other digest of either file, another short closure round is needed.
- Nothing the closure pins may change from the first closure journal until the seal.

**P2. Frontend is final — not met.**
- `binding-v2-round-3.md:10` and `:218` leave X1B2-R3-01 (Major) open. A response has been filed (`binding-v2-round-3-response.md`), but there is no round-4 sign-off yet.
- `main.go` `edb37ec2…` is still unsigned.
- Once it is signed, `OUT/reviews/closure-model-free.json` must contain exactly `scope` = `closure-model-free-go`, `record` and `reviewed` = `[pin(main.go)]`, and must not include `main_test.go`.
- Per the binding response (`:101`), Main must run both shared contract tests against `main.go` before the first `attest_closure.py` journal, preflight included.

**P3. Environment frozen until the seal — standing.**
- No change to the OS, the dyld cache, Homebrew/mise kegs, casks, Go 1.26.10, `claude` 2.1.311 or `codex` 0.166.0.
- Keep the host clock steady.
- Keep preflight receipts, retained copies and `closure-modules` byte-stable. The seal now also re-hashes the preflight receipts.

**P4. Run sequence — standing.**
1. Run the read-only all-roots dry check (with S11 output) over every root.
2. Run one journalled `--preflight`.
3. Only if the preflight passes, run the canonical closure once at `OUT/closure.json`, before any binding run (S16).

## Alternatives

| Option | Tradeoff | Decision |
|---|---|---|
| A. Zero issues at `b1759ec7…`, with P1–P4 kept as hard gates | Records the dependency for the new Core digest; the run still waits on Core and binding | **Recommended** |
| B. Withhold until Core signs `b1759ec7…` | Adds a round with no closure defect, and the time comes out of the 45-day clock | Rejected |
| C. Ask Core to exclude SCRATCH `journal-output` pins from the seal | Weakens a safeguard ("all source and output hashes") with no failure to justify it | Rejected |

## Evidence

**Code read:**
- `preparation.py:40–46`, `:140–152`, `:206–643`
- `attest_closure.py`: grep of journal and output sites, `:724–726`, `:816–856`
- `acceptance.py`: closure reads `:41–58`, `:77–82`, `:205–231`, `:347–355`
- `test_core.py`: grep hits `:84–127`
- `run-config.json:1–15`

**Repository state:**
- `preparation-run/journal/`: 8 pairs. Both (b) pairs have empty `output_pins` and `source_pins`.
- `identifiers-impl-round-8.md:162–165` signs `599a128d…`.
- `binding-v2-round-3.md:10` and `:218`: 1 issue open; `main.go` is not signed.
- `binding-v2-round-3-response.md:60–66` explains the Core change: role regression found at `7afeaa0d`, fixed in `b1759ec7…`.

**How the peers shaped this record:**
- **innovation-challenger** found no failure. Its role-collision, preflight, retention and acceptance traces match mine, and it proposed the closure-shaped fixture in S14.
- **experiment-peer** refuted H1–H7 independently. It corrected the journal count (8 pairs, `906389d7…` carries `effective_configuration`), and its findings became S14 (pins outside OUT), S15, S16 and S17.
- My own reading added the empty `output_pins` of the two existing (b) journals and the current P1/P2 status.
- Neither peer changed the verdict.

## Risks

- **Records cannot be undone in code.** A refused preflight or canonical journal blocks the seal until there is a reviewed recovery. The S11 dry check is the only check that leaves no record.
- **Core may move again.** Any further change to `preparation.py` needs another dependency round and invalidates any closure already run.
- **Not verified on real bytes** [INFERENCE]: the load graph, the dyld cache, offline Go module checks and Java-range files beyond gettext, as in round 4.
- **45-day window.** The run depends on P3 holding through the seal.

## Execution outline

1. Core review: zero issues at `preparation.py` `b1759ec7…` and `verify_inputs.py` `7d6f89ce…`. If either digest differs, hold another short closure round.
2. Binding v2 round 4: zero issues naming `main.go` `edb37ec2…` (or a new digest). Main writes the three-key `closure-model-free.json` and runs both contract tests.
3. Main reruns the full Python suite, `go vet` and `go test`, and recomputes the digests in the table above. Any change to a closure-pinned file needs a new round.
4. Run the all-roots read-only dry check (S11). Every root must classify without refusal.
5. Run one journalled preflight. If it passes, run the canonical closure once at `OUT/closure.json`, then start binding work. Keep P3 until the seal.

## Verification

- The council ran nothing and computed no hashes.
- Reported by Main, not rerun here: 237 Python tests OK; `go vet` clean; `go test -count=1` ok, all at the bytes in `local://x1-packet-digests.md`.

## Verdict

**zero issues — closure attestation may be run at its canonical paths under the stated prerequisites**

**Signed off.** The innovation council confirms the runtime closure attestation (b) implementation, unchanged since round 4, against the changed shared dependency at these digests:

- `execution/attest_closure.py` `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892`
- `execution/closure/native.py` `c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb`
- `execution/tests/test_closure.py` `3a1eae6082d0a9f42074e7ea5db8a2d74720808a91c2e2d5770bf44a82e599dd`
- `execution/tests/test_closure_review.py` `8762b45d3e1db762b9094dfb174ec79c65ae5efa262a2903a2f3894bf8804743`
- `execution/tests/test_closure_round2.py` `7b22a59cbaf43cb0c97381d8ee6faf7fcc6a180c05b3d081ed88f04a4b99cc0c`
- `execution/tests/test_closure_round3.py` `f6eec06e95a7227a02ffcd7ded9598957cffe2f840fe7976bde7e9e1e9e599e8`
- `execution/run-config.json` `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`

**Dependency only** (not a Core sign-off, and not a review receipt for these files):
- `execution/preparation.py` `b1759ec710fa4b71dec1ef91800e84d285c733693b0cb5f4ffd7fe68f147c916`
- `execution/verify_inputs.py` `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`
- with `execution/tests/test_core.py` `4f4bfdaf50fc0d56ff775f8c86bb6bc579c932acc1a76764a55ba4a6f36da07e` read for the accounting tests

**Context only, not signed:**
- `execution/binding/acceptance.py` `2144999a8aece522e72a86a429a1628a2833ada1490fa6e0fe7600e152373be1`
- `execution/model_free/main.go` `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078`

**Prior records reviewed:**
- `closure-impl-round-4.md` `5e9fcd1bcd1a28ae3487505df11d1bb95c378d6c2d1a33a059b85a9ffd349627` (digest as cited at `binding-v2-round-3-response.md:168`)
- `binding-v2-round-3-response.md` `44a391c3b8d7a1aff2c0743fca2af2e642356e675a7bb9e15caa3b41d2cfea0a`

The run is permitted only when P1–P4 all hold. P1 (Core zero-issue sign-off at `b1759ec7…`) and P2 (signed `main.go`) are not met today.
