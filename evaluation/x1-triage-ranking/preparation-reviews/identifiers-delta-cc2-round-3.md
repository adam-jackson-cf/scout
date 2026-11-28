# X1 preparation (d) change-cost-v2 identifier delta — confirming implementation review — round 3

Paths are relative to `evaluation/x1-triage-ranking/` unless absolute. This is the confirming round after round 2's R2-01. It is a static, read-only review. No CC2 data payload was opened, nothing was executed, and no model was called beyond two read-only peer analyses.

## Reviewed digests

Digests were computed by Main with `shasum -a 256`. The council had no shell and did not recompute them. It cross-checked them against the round-2 response's final packet (`identifiers-delta-cc2-round-2-response.md`, final 16-pin table) and the REQUEST's 16-row table (`REQUEST.md:103–118`). All agree.

| File | Bytes | SHA-256 | Status since round 2 |
|---|---:|---|---|
| `execution/identifiers/cc2_delta.py` | 14393 | `657aeb47ee11cbe4db4fdba3948639cf211f8ff032844cbe6d77a815086c0ea6` | unchanged |
| `execution/identifiers/cc2_readers.py` | 8012 | `1ad375a97b43a771245f3f746768dc72814656c1c81c8c4d8793a9c88644144b` | unchanged since round 1 |
| `execution/identifiers/cc2_schemas.py` | 4464 | `db845451e7dbefc8f32ea7b50fbe86cdd5c702330f755e39cee345b2cfc5799f` | unchanged since round 1 |
| `execution/tests/test_cc2_delta.py` | 22402 | `4844221de3ecb6c52b81a770ad7dadbcc42d8e0122636a29de2cdb1fc0e9d8d7` | changed (was `ba96da08…`) |
| `preparation-reviews/identifiers-delta-cc2-REQUEST.md` | 20847 | `c32bd40006bacdf77eb8970d8d4aeaee5be28a6c9bdc243910983c8dba233fb4` | changed (was `288073ce…`); self-pin taken from the round-2 response |
| `execution/identifiers/projection.py` | 7637 | `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072` | unchanged; equals the baseline journal's own `producer_pins` row |
| `execution/preparation.py` | 50213 | `7f8d390fd3ebb50140b5725d7704ccf75b5d157fd74fa73810864a625962753e` | changed (was `858031…`); Core seal integration, reviewed in Core round 3 |
| `execution/verify_inputs.py` | 11211 | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` | unchanged |
| `execution/run-config.json` | 5948 | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | unchanged |
| `preparation/exposure-sources.json` | 7792170 | `67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60` | signed source manifest |
| `preparation-run/journal/906389d7212a401b8c7f781d37960d84.json` | 10301 | `6aa8802a23ac0de75eea9a78de5e6422e8cb16c4762f875dc7ebf0787d77dc6c` | baseline action-d journal |
| `preparation-run/identifiers/manifest.json` | 27924749 | `23a911ae22084db67a5069a49c6db8ef4b0ec7a555c71348446d3372213c738d` | baseline output |
| `preparation-run/identifiers/source-coverage.json` | 30151508 | `24f93c47ab0ab356c7bdafe7bda906c55508370deefbe75e6fad2c13fc682724` | baseline output |
| `preparation-run/identifiers/delta-review.json` | 4635 | `240aa16d942f73e37d1c77b5055aad4a3158f08efcc63555833461e21276db04` | baseline output |
| `preparation-run/identifiers/unresolved.json` | 91278 | `19d10781ceb31b7f4ae58d1a786de24a8af6b09fde925bafe134a398f395dcd9` | baseline output |
| `preparation-run/identifiers/familiarity.jsonl` | 20588100 | `967518114b8187e8fcc4fb92615892d1cdab3a50e3cff6ab8fb10f3eb75e3f30` | baseline output |

Context, not bound: `identifiers-delta-cc2-round-2.md` `d01dbd46b5ba91e1fe2592a8eefd42c3878fc71c81e434d826903509eac62fb4`; `identifiers-delta-cc2-round-2-response.md` `4192e156703f22ad48e0183488d78fb68461a156a66d88190b6ffc85a416a09a`; baseline start record `906389d7….start.json` `c400e46e3f2ba85767165257b425a9719a223303b71cf78700d3111ae06f8738` (8,524 bytes), pinned inside the baseline journal.

Pin check: the REQUEST's 15 literal rows (`REQUEST.md:103–106,108–118`) equal the values above in path, bytes and digest. Its sixteenth row (`:107`) defers to the round-2 response, which gives 20,847 bytes and `c32bd400…`, equal to Main's digest. `REQUEST.md:79` names `preparation.py` `7f8d390f…` and `verify_inputs.py` `7d6f89ce…`, equal to Main's helper pins. A superseded Core-side request digest (`402f0a09…`, `core-impl-round-2-response.md`) is not the current request and is not bound here.

## Prior-issue resolution

| ID | Status | Evidence |
|---|---|---|
| R1-01, R1-02, R1-03 | Closed in round 2; not reopened | `cc2_delta.py` is byte-identical to round 2 (`657aeb47…`). No new evidence. |
| R2-01 (shared code could be frozen by a passed CC2 journal while still changing) | **Closed** | **Fix as prescribed.** `REQUEST.md:83` requires zero-issue Core implementation and delta seal-integration reviews first, both naming the same final `preparation.py` and `verify_inputs.py` digests as the CC2 sign-off and receipt. It then orders: complete edits, confirming CC2 round at the final bytes naming all 16 digests, authentic receipt, single run. It states that an earlier CC2 round at intermediate shared-code bytes is not execution authority.<br>`REQUEST.md:85` forbids changing any of the 16 bound files between the run and the seal, and states that a newer receipt cannot repair stale passed-journal pins. `REQUEST.md:79` carries the final helper digests. `REQUEST.md:53` quotes the exact verdict line, the single scope line and the full-digest scope-block rule enforced by `preparation.py:437–462`. `REQUEST.md:87` names the output-review path and scope.<br>**Catching test.** `test_request_requires_final_shared_code_freeze_order` (`test_cc2_delta.py:277–289`) asserts the prerequisite reviews, equal final helper digests, the confirming-round order, the all-16 freeze, the verdict and scope-block wording, and the output-review path. At `:287–288` it requires the request to contain the live `pin()` digest of `preparation.py` and `verify_inputs.py`. The pre-fix request (`288073ce…`) lacked the freeze-order wording and named `858031…`, so it fails `:280–283` and `:287–288` [INFERENCE: the old bytes were not re-read].<br>**Safety does not depend on the text.** A changed control before the run makes `review_gate` refuse (`cc2_delta.py:168–173,243–244`) before `root.mkdir()` (`:187`), so the root stays free. After the run, the seal re-verifies every receipt row (`preparation.py:558`) and every journal pin (`:618`). Stale shared code therefore cannot produce a passing seal. The ordering text protects liveness, which was R2-01's concern. |
| Round-1 CB-1 (`raw/q` transport-cache boundary) and the ID-01 allowlist | Accepted; carried forward | `cc2_readers.py` and `cc2_schemas.py` are unchanged since round 1. |

## CC2 side against the seal integration (`preparation.py` `7f8d390f…`)

| Contract | Producer | Seal | Result |
|---|---|---|---|
| Imports | `cc2_delta.py:14–18`: `Journal`, `load_config`, `operation_lock`, `paths`, `review_receipt`, `pin` | All present. `preparation.py:20` imports only `verify_inputs` at module level. `fetch_objects` and `attest_closure` are imported lazily inside `collect_seal`, which the producer never calls. | Executed modules: 6, all among the 16 controls. |
| 16 controls | 8 code/test/request pins (`:231–236`) + source + config + baseline journal + 5 outputs (`:243`) | 7 EXEC files + request (`:539–542`) + source + config + same 6 baseline pins (`:543–553`) | Same set. Both compare `(path, sha256)` sets, with paths from `verify_inputs.pin` (`verify_inputs.py:25–31`). |
| Implementation receipt | Canonical path and scope `identifier-delta-cc2-implementation` (`:21,222–227,168–169`) | Same path and scope (`:535–538`). Every output must carry `implementation_review == pin(review_path)` (`:567`). | The seal requires the byte-identical receipt the producer verified; a substituted receipt fails. |
| Output root and kinds | `OUT/identifiers-delta-cc2/{manifest.json, coverage.json, familiarity.jsonl}` with schemas `scout.x1.identifier-delta-cc2.v1`, `scout.x1.identifier-delta-cc2-coverage.v1` and `scout.x1.familiarity-delta-cc2.v1` (`:185–214`) | Same names, order and schemas (`:554–560`). Schemas are used as canonical roles (`:570`). | Match. |
| Common fields | `status`, `clock`, `implementation_review`, `sources_manifest`, `producer_code`, `predecessors`, `development_freshness_filter=False` (`:271–274`). The manifest adds the union, freshness and no-exclusion fields (`:188–191`). The familiarity first line repeats `common` with `entry_kind='source_table'` and the same `sources` (`:201–202`). | Checks at `:563–575`. | Every checked field is produced with the required value. |
| Producer identification | Journal action `d`, `source_pins` = 8 implementation pins + source manifest (`:238`). Passed `output_pins` = role-free pins in creation order (`:214,276`). | Exactly one passed action-`d` journal whose `output_pins` equal the three role-free pins (`:576–580`). Code + source must be a subset of its `source_pins` (`:582–583`). | Baseline `906389d7…` has five different outputs, and refused journals are not in `passed` (`:619–635`), so there is no confusion. |
| Output review | — | `OUT/reviews/identifier-delta-cc2-output.json`, scope `identifier-delta-cc2-output`, binding all three outputs and the producer journal (`:584–588`) | Matches `REQUEST.md:87`. |
| Output accounting | Creates only the three files under its root | The three outputs and the output-review receipt are added to the seal outputs (`:570,590`). The implementation receipt lives in `preparation-reviews/`, outside `OUT`. | No unaccounted bytes (`verify_output_accounting`, `:499–514`). |
| Clock | `--clock` is not in the `expected` path check (`:221–227`) | `Journal` → `read_clock` refuses a non-canonical clock (`preparation.py:205–208`). The seal compares every output's `clock` with `pin(clock_path)` (`:563`). | Closed. |

The seal fixture (`test_binding_seal.py:121–147`) calls the real `cc2_delta.publish`, `Journal` and `review_receipt` record format. Its hand-built `common` and `source_pins` equal the producer's at `cc2_delta.py:238,271–274` key for key. `test_cc2_seal.py` exercises the canonical roles, both scopes, and each refusal branch of `collect_cc2_delta`.

## New issues

None.

One peer proposed a sequencing condition: Core round 3 had not yet reported zero issues at `7f8d390f…` when this round was held, and `review_receipt` does not read prose conditions. The council does not count this as an issue in the reviewed bytes, for three reasons:
- `REQUEST.md:83` already forbids the run until both Core-side reviews pass at the same digests.
- If Core round 3 changes `preparation.py` before the run, the receipt's `7f8d390f…` row no longer matches. `review_gate` then refuses before the output root is reserved, so the run is not wasted.
- The residual failure needs Main to break the signed order: run first, then let `preparation.py` change. That is the documentary control R2-01 prescribed.

The condition is recorded in the sign-off below.

## Suggestions (not blocking)

1. **Freeze the receipt and this record as well.** Name `preparation-reviews/identifiers-delta-cc2-implementation.json` and its council record (this file) in the post-run freeze at `REQUEST.md:85`. Both become immutable once the run pins them: see `cc2_delta.py:272` and `preparation.py:441,558,567`.
2. **Check all 15 literal rows.** Extend `test_cc2_delta.py:287–288` to compare every literal row of the REQUEST packet with `pin()`, not just the two helpers. Also consider asserting that the superseded `858031…` digest is absent.
3. **Add an end-to-end test** that runs the real `cc2_delta.run` into `collect_cc2_delta` with the real `review_receipt`. Today the producer test mocks `review_gate` (`test_cc2_delta.py:344`), and the seal fixture hand-copies the producer's `common`. Drift between them would go unseen; it is absent at these bytes.
4. **Bind executed modules at the seal.** Require the CC2 producer journal's `producer_pins` to be a subset of the reviewed controls. Today they are only re-verified for unchanged bytes (`preparation.py:618`).
5. **Carried forward from round 2, still open:**
   - Validate the baseline, the receipt and root absence before reserving the journal.
   - Use `except BaseException` at `cc2_delta.py:277`.
   - Add a changed-baseline-output refusal test.
   - Note that any failure after `root.mkdir()` (`cc2_delta.py:187`) permanently blocks a rerun.

## Peers and where they changed the recommendation

- **innovation-challenger (Cc2R3Challenger):**
  - Confirmed independently that R2-01 is closed, the producer and seal contracts match, and the request table is consistent with the packet.
  - Suggestion 3 is partly its contribution.
  - It changed nothing in the verdict.
- **experiment-peer (Cc2R3Peer):**
  - Confirmed the same points, and that the baseline journal's `projection.py` producer pin equals the bound digest.
  - It contributed suggestions 1, 2 and 4.
  - It raised the sequencing condition. The council kept that as a sign-off condition rather than an issue, for the reasons under New issues.
- **Council:** the field-by-field trace of the producer's `common` dict and output pins against `collect_cc2_delta`, and the observation that a pre-run helper change refuses before the root is reserved.

## Verification

- Static reading only: `cc2_delta.py`, `preparation.py:1–25,320–800`, `verify_inputs.py:19–35`, `test_cc2_delta.py`, `test_binding_seal.py:110–148`, `test_cc2_seal.py` (grep), the REQUEST, and the round-2 record and response.
- Main reports 287 Python tests OK and `go vet` / `go test -count=1` OK on `execution/model_free` at these bytes. The council did not re-run them.
- **Main should run before building the receipt:** `shasum -a 256` and `wc -c` on all 16 bound paths. Then generate the receipt rows with `verify_inputs.pin()` on the exact absolute paths in `REQUEST.md:103–118`.

Verdict: ZERO ISSUES — signed off

**Zero issues. Signed off.**

## Sign-off

The innovation council signs off the CC2 identifier-delta implementation, its allowlist and its consistency with the seal integration, at the 16 digests named below.

The following conditions apply. They come from `REQUEST.md:83–87`:
- **Before the receipt and the run:** this record is execution authority only once Core round 3 and the delta seal-integration review have both recorded zero issues naming `preparation.py` `7f8d390fd3ebb50140b5725d7704ccf75b5d157fd74fa73810864a625962753e` and `verify_inputs.py` `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`.
- **If any bound file changes before the run:** this sign-off lapses, and a new confirming round is required.
- **After the run:** none of the 16 files, the receipt or this record may change before the seal.

Signed-off scope: identifier-delta-cc2-implementation

- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/identifiers/cc2_delta.py` — 14393 bytes — `657aeb47ee11cbe4db4fdba3948639cf211f8ff032844cbe6d77a815086c0ea6`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/identifiers/cc2_readers.py` — 8012 bytes — `1ad375a97b43a771245f3f746768dc72814656c1c81c8c4d8793a9c88644144b`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/identifiers/cc2_schemas.py` — 4464 bytes — `db845451e7dbefc8f32ea7b50fbe86cdd5c702330f755e39cee345b2cfc5799f`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/tests/test_cc2_delta.py` — 22402 bytes — `4844221de3ecb6c52b81a770ad7dadbcc42d8e0122636a29de2cdb1fc0e9d8d7`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-reviews/identifiers-delta-cc2-REQUEST.md` — 20847 bytes — `c32bd40006bacdf77eb8970d8d4aeaee5be28a6c9bdc243910983c8dba233fb4`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/identifiers/projection.py` — 7637 bytes — `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/preparation.py` — 50213 bytes — `7f8d390fd3ebb50140b5725d7704ccf75b5d157fd74fa73810864a625962753e`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/verify_inputs.py` — 11211 bytes — `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/run-config.json` — 5948 bytes — `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation/exposure-sources.json` — 7792170 bytes — `67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/journal/906389d7212a401b8c7f781d37960d84.json` — 10301 bytes — `6aa8802a23ac0de75eea9a78de5e6422e8cb16c4762f875dc7ebf0787d77dc6c`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/manifest.json` — 27924749 bytes — `23a911ae22084db67a5069a49c6db8ef4b0ec7a555c71348446d3372213c738d`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/source-coverage.json` — 30151508 bytes — `24f93c47ab0ab356c7bdafe7bda906c55508370deefbe75e6fad2c13fc682724`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/delta-review.json` — 4635 bytes — `240aa16d942f73e37d1c77b5055aad4a3158f08efcc63555833461e21276db04`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/unresolved.json` — 91278 bytes — `19d10781ceb31b7f4ae58d1a786de24a8af6b09fde925bafe134a398f395dcd9`
- `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/familiarity.jsonl` — 20588100 bytes — `967518114b8187e8fcc4fb92615892d1cdab3a50e3cff6ab8fb10f3eb75e3f30`
