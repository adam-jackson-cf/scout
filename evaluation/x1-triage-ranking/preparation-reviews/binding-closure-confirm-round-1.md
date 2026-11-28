# X1 preparation — binding v2 and runtime closure, confirming round at the final Core — round 1

**Scope.** This round checks two tracks against the final Core:
- binding v2, last signed in `binding-v2-round-4.md` against Core `b1759ec7…`;
- the runtime closure (b), last signed in `closure-impl-round-4.md` and `closure-impl-round-5.md`, whose prerequisite P1 required a Core sign-off.

Core was signed off with zero issues at `preparation.py` `9572aed3…` (`core-impl-round-5.md`). This round also has to provide a record that `review_receipt` will accept as authority, because `binding-v2-round-4.md` cannot serve.

**Method.** Read-only. The council ran no command, test, hash, adoption, attestation, rehearsal, acceptance, seal or model call. All digests are Main's (`shasum -a 256`); the council did not recompute any. Paths are relative to `evaluation/x1-triage-ranking/`. Three analyses fed this record:
- **My own trace:**
  - all of `execution/preparation.py`;
  - in `execution/attest_closure.py`: `reviewed_go_sources`, `attest` and `main`;
  - in `execution/binding/acceptance.py`: `reviewed_inputs`, `verify_rehearsal`, `rehearse`, `complete_operation` and `main`;
  - all of `test_binding_adoption.py` and `test_binding_seal.py`;
  - the authority fixtures in `test_closure_review.py` and `test_closure_round2.py`;
  - the proposal, amendment `:64–74`, `run-config.json` paths and reviews, the journal directory, CC2 journal `0378e156…`, and the CC2 implementation receipt;
  - Core rounds 1–5, closure rounds 4–5 and binding round 4.
- **`innovation-challenger` (ChallengerConfirm):** an adversarial trace of every receipt consumer and the exact parser rules.
- **`experiment-peer` (PeerConfirm):** one falsification test for each Core change (H1–H8), and a search for which digests have been signed.

**Outcome: 1 issue (X1BC-R1-01, Minor).** No Core change breaks `adopt`, `load_config`, closure preflight or canonical, the rehearsal, or the seal's handling of binding and closure records. Two closure-signed test files have changed since closure round 5 signed them, and no zero-issue record names their current bytes. Main also did not supply current digests for the other closure-signed files. This record is therefore not authority. A round 2 that receives those digests can close the issue without any code change.

## Reviewed digests (Main's)

| File | SHA-256 | Since last binding/closure sign-off |
|---|---|---|
| `execution/preparation.py` | `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38` | changed (binding r4 / closure r5: `b1759ec7…`); Core-signed in `core-impl-round-5.md` |
| `execution/model_free/main.go` | `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078` | unchanged |
| `execution/model_free/main_test.go` | `54e7bc66d4df750796ae4201b1dbeba172373a34dc46c3d4cce629d17e54d617` | unchanged |
| `execution/tests/model_free_contract.json` | `bc88c34ae1f8edc800e2c6a6d7eb86af4fcda205899deaba085630edab0a334d` | unchanged |
| `execution/tests/test_binding_contract.py` | `dc817736d39551e01c62acb2e6b3dffeb861de825b4fb56c9b5c3198dc0a9efa` | unchanged |
| `execution/binding/acceptance.py` | `2144999a8aece522e72a86a429a1628a2833ada1490fa6e0fe7600e152373be1` | unchanged |
| `execution/binding/interlock.py` | `19464e8940d0458be3606ed0438ae778fd01b48b73f8c2fee801856b80b2e6a1` | unchanged |
| `execution/binding/fixtures.py` | `48552a2bd5f8492b83237f147916bde12f2137a6aee6fbae99b01cda9e40381b` | unchanged |
| `execution/binding/fixtures/**` (9 files) | as listed in `binding-v2-round-4.md` (`:201–209` there) | unchanged (Main) |
| `execution/tests/test_binding.py` | `2458e4289d67a1a061fcf43a2de6e42f1d586daee6cb4452f17f857ee1d36ab5` | unchanged |
| `execution/tests/test_binding_adoption.py` | `0396765fc3d51de609bf2d935d5b0b8979cd3b84ff65c5f12080e384912ce432` | changed (was `4b186aa1…`) |
| `execution/tests/test_binding_boundaries.py` | `c5c1d042ee440e39eab66bb3bdb0d0d13b4756d4fe0c1018a16d7e63b7c2b177` | unchanged |
| `execution/tests/test_binding_seal.py` | `7c05eaf2f577210361bbd18a93087e7f59b8fffa687c09d98fad2824d8f4b593` | changed (was `b21763f3…`); also named in `core-impl-round-5.md` |
| `execution/tests/test_core.py` | `7bdc801955675437a095121b28cfbf2d85ca75c004abbe870f1e14ca5a1d2628` | changed (was `4f4bfdaf…`) |
| `execution/attest_closure.py` | `113d2d7272db8bdcda3abb7a745a7c8f4e8e49a447b7f36c9d14449c00420892` | unchanged |
| `execution/fetch_objects.py` | `64bb3dd3b99d5227d0abd8e460f4dbabda18e15b37c16c71358241566e8c6a9f` | changed (was `737b5ec7…`) |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` | unchanged |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | unchanged |
| `execution/run-config-amendments-proposed/001-binding-v2.json` (proposal) | `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121` | unchanged |
| `preparation/model-free-binding-check-v2.json` (recipe) | `9c1340da248ca25d99053f81309ff4c5d13f433c2faf2f22e055d2197c5020dc` | unchanged |
| `preparation/model-free-binding-amendment.md` (amendment) | `27f1100122f9867974049a88dfae14587aa6a9db7d96ae8065faca45fca284f7` | unchanged |

**Closure-signed files whose current digests were not supplied** (see X1BC-R1-01). The second column is the digest last signed in `closure-impl-round-5.md:197–202`:

| File | Last closure-signed digest | Recorded since |
|---|---|---|
| `execution/closure/native.py` | `c264d61acb33645a6d36b09fad794023db5306a96f50f9ce61698a83d396cabb` | no later record found |
| `execution/tests/test_closure.py` | `3a1eae6082d0a9f42074e7ea5db8a2d74720808a91c2e2d5770bf44a82e599dd` | no later record found |
| `execution/tests/test_closure_review.py` | `8762b45d3e1db762b9094dfb174ec79c65ae5efa262a2903a2f3894bf8804743` | **changed:** `534e4754f6100a69e60c7e12b6e7740a39fc84c8f0d81b094ab4b19f26cb2906` (`core-impl-round-1-response.md:100`; "Context" only in `core-impl-round-2.md:24`) |
| `execution/tests/test_closure_round2.py` | `7b22a59cbaf43cb0c97381d8ee6faf7fcc6a180c05b3d081ed88f04a4b99cc0c` | **changed:** `70efac053f0b92b5686455481bab1cd074874d96019091bb493f76cff465f274` (`core-impl-round-1-response.md:101`; "Context" only in `core-impl-round-2.md:24`) |
| `execution/tests/test_closure_round3.py` | `f6eec06e95a7227a02ffcd7ded9598957cffe2f840fe7976bde7e9e1e9e599e8` | no later record found |

**Main's reported checks, not re-run here:**
- the full Python suite: 289 tests OK;
- `go vet` and `go test -count=1` on `execution/model_free`: pass.

These are recorded in `core-impl-round-3-response.md` and `core-impl-round-4-response.md`.

## Q1. Do the Core changes (`b1759ec7` → `9572aed3`) preserve the binding and closure contracts?

**Yes.** The peer's falsification table (H1–H8) and my own trace agree; the challenger found no counterexample. Change by change:

1. **`sources` copy and predecessors (X1C-R1-02).**
   - `collect_seal` now starts from `list(gate['pins'])` (`preparation.py:648`), and `predecessors` is built from the gate pins only (`:745`).
   - Binding and closure record pins still reach `sources` through `verify_tree_pins(value)` in the `required_records` loop.
   - No binding or closure code reads `predecessors`.
   - **Preserved.**
2. **Reviewed recovery and passed producers (X1C-R1-03).**
   - A failed or refused journal now needs `OUT/reviews/operation-recovery-<id>.json`. That receipt must bind the start record, the terminal record and every output pin, and the failure is reported as a deviation (`:621–630`).
   - Every required record now needs a passed producer (`:726–731`), and both tracks provide one:
     - `closure.json` carries `journal_start=pin(journal.start_path)`, set at the end of `attest()` in `attest_closure.py`, which matches `passed_starts`.
     - Each binding receipt is the `output_pins=[pin(canonical)]` of its journal (`acceptance.py:293`), which matches `passed_outputs`.
   - **Side effect, an improvement:** a refused closure preflight or canonical journal now has a reviewed recovery route. Closure round 4 had treated that case as a dead end.
   - **Preserved.**
3. **Shared-file input receipt (X1C-R1-04).**
   - `verify_input_receipt` (`:408–421`) drops only whole-file pins of the five `SHARED` paths.
   - The closure reads only `deferred_to_b`, and pins `input-verification.json` whole in `attest()` `inputs`. That file is create-once.
   - **Preserved.**
4. **Publication lock.**
   - `Journal.__init__` takes the shared operation lock, then the directory publication lock, then reloads and compares the config (`:353–371`, check at `:359–360`).
   - `finish` takes the same lock (`:373–382`). `adopt` takes the operation lock exclusively and never the publication lock (`:264`).
   - **Closure:** `main()` constructs its `Journal` before the `try` block and before writing any output, so an `adopt` racing it refuses the closure with no output left behind.
   - **Binding:** the canonical binding path still has Core round 4 S3, the adoption-to-binding orphan. Core round 5 requires that to be resolved before canonical binding (c). It is disposed of below as runbook rule RB-1.
   - **Preserved for the closure; binding is safe under RB-1.**
5. **CC2 seal integration.**
   - The seal now requires CC2 evidence unconditionally (`collect_cc2_delta`, `:532–594`, called at `:732`). It binds `run-config.json` `c817f678…`, which adoption never modifies: adoption writes only `execution/run-config-amendments/` and `OUT/config-adoptions/`.
   - CC2 journal `0378e1563617446ea7bdfc993819aa51` passed at `2026-11-10T15:25:30Z`. It pins `preparation.py` `9572aed3…` (50,257 bytes) in both `source_pins` and `producer_pins`, with no amendments in its effective configuration.
   - Core is therefore already frozen in practice: any edit would break the seal's `verify_tree_pins` over that journal.
   - **Preserved.**

**Each consumer named in the brief:**
- **`adopt`** (`:259–321`). The contract is unchanged:
  - canonical proposed path;
  - `amends` equal to `execution/run-config.json` / `c817f678…`;
  - proposal `authority.path` equal to the `--review` argument;
  - `review_receipt`, with the proposal digest in `reviewed`;
  - all journals terminal;
  - writer pin `pin(__file__)`.
  
  All 10 journal pairs on disk are terminal (directory listing).
- **`load_config`** (`:105–151`) checks the receipt of each adopted amendment, then `verify_adoption`, then `verify_journal_configuration` for every start record.
- **`review_receipt`** (`:437–461`). The rules are under Q4.
- **Seal handling of binding records.** `verify_amendment_reviews` (`:161–176`) applies. The binding loop checks:
  - `key_present` is false;
  - `runtime_closure` is the closure pin;
  - the exact v2 `binding_acceptance_keys`;
  - the Django revision pair.
  
  `verify_output_accounting` (`:499–514`) keeps the canonical roles.
- **Closure verification.** `verify_installed_closure` runs at seal (`:702–703`) and before every binding launch (`acceptance.py:81–82`). Core is pinned dynamically through `own_sources()`; neither track names `b1759ec7…`.
- **`reviewed_go_sources`** (`attest_closure.py:554–564`):
  - It calls `review_receipt` with no `study`, and the study is derived from the receipt's `preparation-run` ancestor (`:442–444`).
  - Rows must be unique, direct `model_free/*.go` files. `go_metadata` then requires the reviewed set to equal the selected Go set.

## Q2. Do the changed binding tests still catch what binding round 4 said they catch?

**Yes.**
- **`test_binding_adoption.py` (`0396765f…`).**
  - `review()` (`:44–50`) now writes a council record under `STUDY/preparation-reviews/` containing the exact verdict line, a `Signed-off scope:` marker and the reviewed digests. No test mocks `review_receipt`.
  - Each test round 4 relied on is still present and aimed at the same guard:
    - binding round-3 S1, `test_duplicate_fourth_amendment_row_refuses_binding`, with the regex `exactly proposal, recipe and amendment`;
    - missing proposal digest refused at `adopt` (`does not bind proposed amendment`);
    - proposal path substitution, extra row, missing recipe and missing amendment;
    - adopted-SHA mismatch, exact adopted row, and adoption-review pin mismatch.
- **`test_binding_seal.py` (`7c05eaf2…`).**
  - `prepare()` now:
    - adds `closure.json` and the Django object receipt as required records;
    - links `closure.json` to the start pin of the (b) journal;
    - stages CC2 controls, and `publish_cc2()` uses the real `cc2_delta.publish`.
  - The seal path is therefore the real `collect_seal` at `9572aed3…`, including `collect_cc2_delta` and the passed-producer check.
  - The round-4 assertions remain:
    - both v2 output roles;
    - a per-binding refusal when a binding is missing from the review (`scoped reviews do not bind all required evidence`);
    - invalidation of the rehearsal record when it is changed or deleted;
    - round-3 S2's runtime and source omission (`does not bind report, runtime and source pins`);
    - S5's production `binding-acceptance` scope.
- **`test_core.py` (`7bdc8019…`).** It still contains:
  - `test_canonical_output_role_preserved_when_journal_also_pins_it`;
  - `test_duplicate_journal_outputs_are_deduplicated_but_all_verified`;
  - `test_tampered_journal_pinned_output_is_refused`;
  - the tests for pre- and post-adoption journal validation and for the amendment chain.
- **Limits, unchanged in kind:** these tests mock `verify_installed_closure`, `documentary_gate`, the scanners and the Git inspection, and they never exercise the seal's chmod or write. Several adoption negatives use bare `assertRaises` (S4).

## Q3. Closure P1, and do closure rounds 4 and 5 still hold at these bytes?

- **P1.**
  - **Core is final:** it is signed off with zero issues at `preparation.py` `9572aed3…` and `verify_inputs.py` `7d6f89ce…` (`core-impl-round-5.md`), and in practice it is frozen by CC2 journal `0378e156…`.
  - **Dependency:** closure round 5's dependency analysis still holds at `9572aed3…`. Under Q1 no Core change affects the closure's journal, preflight receipt, accounting of `closure.json`, `own_sources` or `verify_installed_closure`. Closure round 5's S17 has in effect been exercised: the start of `0378e156…` ran `load_config` → `verify_journal_configuration` over every start record, including `906389d7…` [INFERENCE from its passed status].
  - **Not closed:** closure rounds 4 and 5 sign a set of closure files. Two of them have since changed and are unsigned, and three more are unconfirmed (X1BC-R1-01). Closure round 4's outline (`closure-impl-round-4.md:232`, "If any closure file has changed, a new round is needed") and CLAUDE.md experiment rule 5 therefore block (b) until a zero-issue round names their current digests.
- **P2.** This is met only once a zero-issue record carries the `closure-model-free-go` scope block naming `main.go` `edb37ec2…`, and `OUT/reviews/closure-model-free.json` is exactly `{scope, record, reviewed:[pin(main.go)]}`, with no `main_test.go`. This record is not that authority.
- **P3 and P4: standing.** P3 is the environment freeze through seal. It now also covers the Git keg pinned with `brew pin git` (`RESULTS.md`) and the byte stability of the preflight receipt and its retained copies. P4 is the order: the all-roots dry check (closure S11), then one journalled preflight, then one canonical run before any binding run (closure S16).

## Q4. What else would make `adopt`, the closure, the rehearsal or the seal refuse?

**Nothing in code.** The only refusal found is the known one: `binding-v2-round-4.md` lacks both the exact verdict line and the scope marker, and is refused, as `test_preparation_evidence.py:45–54` asserts. The authority record must satisfy every rule below.

**Record name and location.** A record at `preparation-reviews/binding-closure-confirm-round-1.md`, or a later round's name, is acceptable. No code checks a filename, extension or round number. The record's resolved path must lie under `STUDY/reviews/` or `STUDY/preparation-reviews/` (`preparation.py:445–447`).

**Receipts.**
- They must sit at:
  - `OUT/reviews/binding-v2-amendment.json`, fixed by the proposal's `authority.path`;
  - `OUT/reviews/closure-model-free.json`, fixed by `GO_REVIEW_RELATIVE`.
- `adopt --review` must name the JSON receipt, not the Markdown record (`:276–277`).
- Each receipt has exactly the keys `scope`, `record` and `reviewed` (`:439`).
- The amendment receipt has exactly three rows: proposal, recipe and amendment, at those canonical paths (`acceptance.py:319–336`).
- The closure receipt has exactly `[pin(ROOT/evaluation/x1-triage-ranking/execution/model_free/main.go)]`.

**Record text.** Line-level parsing is exact:
- The verdict line must be the whole line `**Zero issues. Signed off.**` (`:449`).
- Each scope marker must be one whole line, `Signed-off scope: <scope>`, with:
  - no backticks;
  - no trailing spaces;
  - LF line endings only;
  - exactly one such line for each scope (`:451–454`).
- A scope block runs to the next `Signed-off scope:` line or to end of file (`:456`); headings and blank lines do not end it.
- Each reviewed digest must appear in full, in lowercase, inside its own block, with no hex character next to it (`:457–460`).

**Fail-safe order.** `adopt` validates the receipt before writing anything (`:279` comes before `:299`). The closure validates it before constructing its `Journal` (`attest_closure.py` `main()`). A malformed record therefore refuses without leaving a create-once artefact.

## Issues

### X1BC-R1-01 — Minor — closure-signed tests changed without a zero-issue round naming them; other closure digests unconfirmed

- **Location.**
  - `closure-impl-round-5.md:197–202` signs `test_closure_review.py` `8762b45d…` and `test_closure_round2.py` `7b22a59c…`.
  - `core-impl-round-1-response.md:100–101` records them as `534e4754…` and `70efac05…` after their authority fixtures moved to the new receipt format. In their current text, `test_closure_review.py:180–181` and `test_closure_round2.py:114–115` write `**Zero issues. Signed off.**` and `Signed-off scope:` fixtures, a format that did not exist when closure round 5 signed them.
  - `core-impl-round-2.md:24` lists both only as "Context". Core rounds 3–5 do not name them.
  - The brief supplies no current digests for `closure/native.py`, `test_closure.py` or `test_closure_round3.py` either.
- **Failure.** Closure rounds 4 and 5 no longer hold for the closure's signed file set. Running (b) now would run a signed experiment implementation whose changed scripts have no zero-issue round. That breaks CLAUDE.md experiment rule 5 and closure round 4's outline (`:232`). This round cannot name the current bytes because it has no digests for them.
- **Content check, done so round 2 can be short.** The changed authority tests still reach the production `review_receipt` and still catch what the closure rounds claimed:
  - `test_r107_reviewed_frontend_bytes_must_stay_signed` (`test_closure_review.py:174–188`): drift of `main.go` → `pin mismatch`.
  - `test_r203_*` (`test_closure_round2.py:110–186`):
    - missing receipt → `FileNotFoundError`;
    - wrong scope and an extra key → `invalid scoped review`;
    - drifted record → `pin mismatch`;
    - unrelated reviewed file → `only unique model_free`;
    - selected-set equality;
    - the journal binds both the receipt and the record.
- **Required fix (no code change).**
  1. Main supplies `shasum -a 256` for:
     - `execution/closure/native.py`;
     - `execution/tests/test_closure.py`;
     - `execution/tests/test_closure_review.py`;
     - `execution/tests/test_closure_round2.py`;
     - `execution/tests/test_closure_round3.py`;
     - `execution/tests/test_preparation_evidence.py`.
     
     The last is the Core-owned parser test. The only round that names `634f5242…` is `core-impl-round-2.md`, which closed with an issue open, and Core round 5's sign-off list omits it.
  2. A confirming round 2 confirms that the changes to the closure tests are limited to the receipt fixtures. It names all closure-signed digests and the digests in the table above in its human-readable sign-off. It also carries the two authority scope blocks.

## Dispositions

- **Core round 4 S3 / Core round 5 "resolve before canonical binding (c)".** The failure path:
  1. `acceptance.py main()` loads the config (`:423`).
  2. `adopt` completes.
  3. Binding reserves `canonical.open('x')` (`:436–437`).
  4. `Journal` refuses (`preparation.py:359–360`). That leaves an empty create-once `binding/*.json`, and the seal will refuse it.
  
  The council resolves this with **runbook rule RB-1**: run exactly one `adopt` of `001-binding-v2.json` and confirm it through `load_config` (`amendments` holds the adopted row) before any `acceptance.py` process starts, rehearsal included. Never run `adopt` while a binding CLI may be starting, and make no further amendment once canonical (c) has begun. RB-1 takes effect when the zero-issue round 2 restates it.
- **Binding round 4 S1 (the closure P1 confirming round).** It is discharged for the Core dependency by Q1 and Q3. For the closure's own file set it is still open under X1BC-R1-01.
- **Binding round-4 S2–S7 and closure-round-5 S14–S17:** still suggestions; not reopened.

## Suggestions (not blocking)

- **S1. Pre-check without side effects.** Before `adopt` or the preflight, call `preparation.review_receipt(<amendment receipt>, 'model-free-binding-v2-amendment')` and `attest_closure.reviewed_go_sources(ROOT, OUT)` from the pinned interpreter. Neither writes anything.
- **S2. Freeze the authority artefacts.** Never edit or delete:
  - the authority record;
  - either receipt;
  - the proposal `001-binding-v2.json`;
  - the adopted copy.
  
  Every `load_config` re-hashes the `proposal` pin in the adoption command through `verify_tree_pins(command)`. The closure pins the record and the receipt into its journal and into `closure.json`. Deleting the proposal as a "clean cutover" would make every later operation refuse, the seal included.
- **S3. Writer path and wrapper.** Run `adopt` only as the canonical absolute `$EXEC/preparation.py`. The writer pin compares path strings (`:90`), so a spelling with `..` or a symlink would make every later `load_config` refuse. Never run `adopt` under `preparation.py journal` (Core round 4 S4).
- **S4. Message-specific negatives.** Several negatives in `test_binding_adoption.py` use bare `assertRaises(InterlockFailure)` or `assertRaises((ValueError, InterlockFailure))`. Message regexes would pin each one to its intended guard. This is optional, and it changes a bound test.
- **S5. A closure-shaped journal fixture (carried from closure round 5 S14).** `test_binding_seal.py:96–104` finishes the (b) journal with no outputs. The real closure pins `closure.json` and an external preflight receipt.
- **S6. Layout of the authority record.** Put both scope blocks at the end, after the `Verdict:` line, each holding only its own digests. Separate per-scope records are an equally valid alternative.

## Recommendation

Do not create receipts, adopt, or run (b) on this record. Main supplies the six digests listed under X1BC-R1-01. Then hold a confirming round 2, limited to confirming those bytes and restating RB-1. If it finds zero issues, its record carries the two authority blocks.

## Alternatives

- **A. Round 2 with the missing digests (recommended).** Low cost: no code changes, and the content of the closure tests has already been checked here.
- **B. Sign binding v2 now and leave the closure tests to a separate closure round.** Rejected. The brief asks one record to carry both scopes, and the closure scope would still be blocked by rule 5. Splitting adds a round without removing any work.
- **C. Treat the test changes as outside rule 5 because they are test-only.** Rejected. Closure rounds 4 and 5 sign the tests by digest, and the rule makes no exception for tests.

## Evidence

- **Code:** in `preparation.py`:
  - `:43–67`, `:74–103`, `:105–151`, `:161–176`;
  - `:237–257`, `:259–321`, `:340–382`, `:408–421`;
  - `:437–461`, `:464–482`, `:499–514`;
  - `:532–594`, `:596–639`, `:641–750`.
- **`attest_closure.py`:** `:26–33`, `:554–564`, `attest()`, `main()`.
- **`acceptance.py`:** `:13–14`, `:81–82`, `:276–294`, `:311–356`, `:359–380`, `:383–451`.
- **Tests:**
  - `test_binding_adoption.py:22–296`;
  - `test_binding_seal.py:26–326`;
  - the `test_core.py` test list;
  - `test_closure_review.py:174–188`;
  - `test_closure_round2.py:110–186`;
  - `test_preparation_evidence.py:31–75`.
- **Records:**
  - `preparation-run/journal/` (10 terminal pairs);
  - `journal/0378e1563617446ea7bdfc993819aa51.json` (`source_pins` and `producer_pins` `9572aed3…`; no amendments);
  - `preparation-reviews/identifiers-delta-cc2-implementation.json`;
  - `run-config.json` paths and `required_reviews`;
  - the proposal `authority`;
  - amendment `:64`, `:66–74`;
  - `RESULTS.md`.
- **Review records:**
  - `binding-v2-round-4.md`;
  - `closure-impl-round-4.md:154–176`, `:232`;
  - `closure-impl-round-5.md:18–26`, `:197–203`;
  - `core-impl-round-1.md:80`, `:156`;
  - `core-impl-round-1-response.md:91–101`;
  - `core-impl-round-2.md:19–24`, `:179–187`;
  - `core-impl-round-4.md:129–143`;
  - `core-impl-round-5.md`.

## Risks

- All behaviour above comes from reading the code [INFERENCE]. Main's 289 tests and Go checks support it, but no real closure, rehearsal or seal has run.
- **Irreversibility:**
  - `adopt`, every `attest_closure.py` journal (preflight included) and every canonical binding reservation are create-once.
  - The authority record and receipts are frozen from the first closure journal or from adoption, whichever comes first.
  - `main.go` is frozen from the first closure journal.
- **Core cannot change:** journal `0378e156…` already pins it.
- **Runtime-only frontend defects** remain a disclosed residual risk (amendment `:74`).
- **Clock:** every round uses time on the 45-day clock (deadline `2026-12-24T03:22:58.701909Z`).

## Execution outline

1. Main supplies `shasum -a 256` for the six files in X1BC-R1-01 and re-confirms the 21 rows above. No file is edited.
2. Hold confirming round 2. Its record needs:
   - the human-readable sign-off naming every digest;
   - the line `Verdict: ZERO ISSUES — signed off`;
   - the exact verdict line;
   - two scope blocks at the end: the binding v2 amendment scope (proposal, recipe and amendment digests) and the closure Go scope (`main.go` only).
3. Run both contract tests on `main.go` `edb37ec2…` and record the argv, exit codes and digests (binding round 4 S2). Then run the S1 pre-checks on both receipts.
4. Write `OUT/reviews/closure-model-free.json` and the three-row `OUT/reviews/binding-v2-amendment.json`, checking each row against `pin()`.
5. Run the closure all-roots dry check (S11), one journalled preflight, then the single canonical `OUT/closure.json`.
6. Under RB-1, run the canonical `adopt`, before any `acceptance.py` process, and confirm it with `load_config`.
7. Produce the manifest, `binding/runtime.json` and `reviews/implementation.json`. Then run the scratch rehearsal and write its review.
8. Run canonical synthetic, then Django, then `reviews/binding.json`.
9. Finish the CC2 output review and the remaining reviews. Run `seal --check`, then `seal`.

## How the peers shaped this record

- **experiment-peer:**
  - Found X1BC-R1-01, which changed my verdict from zero issues to one. I had checked the content of the closure-test fixtures but not compared their digests with closure round 5. I confirmed the digest drift from `closure-impl-round-5.md` and `core-impl-round-1-response.md`.
  - Supplied the H1–H8 falsification table.
  - Raised the open requirement from Core round 4 S3, disposed of here as RB-1.
  - Raised the unsigned Core-owned `test_preparation_evidence.py`, folded into the fix for X1BC-R1-01.
  - Raised S2–S5.
- **innovation-challenger:**
  - Found no issue.
  - Independently confirmed the exact parser and location rules, the routing of the receipts, and that the record's filename is acceptable.
  - Confirmed that both tracks satisfy the passed-producer check.
  - Supplied S6 (separate records as an alternative) and the reminder that the RB-1 race still stands.
- **Mine:**
  - The Q1 trace for each change against the real journal `0378e156…`.
  - The Q2 test mapping.
  - The fail-safe order of `adopt` and the closure.
  - The decision to keep this record free of authority markers while an issue is open.

## Sign-off

None. With X1BC-R1-01 open, the innovation council does not sign off any file in this round. Reading found no defect in the binding v2 amendment and implementation, or in the closure's dependency on Core, at the 21 digests in the "Reviewed digests" table. That finding is not an approval and gives no receipt authority. A zero-issue round 2 must name those digests together with the six missing closure and Core test digests.

Verdict: 1 ISSUES
