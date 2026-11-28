# X1 preparation (d) identifier projector — implementation review — round 5 (integration)

**Scope.** This round covers only the integration change in `execution/preparation.py` and the new amendment file it reads. The question is whether that change alters the projector's behaviour, its clock and config binding, or its create-once and journal guarantees, in a way that could plausibly make the action-(d) run fail. It does not cover identifier completeness or the delta, and it does not review the amendment itself. No files were edited. The projector, the tests and seal were not run. The only model calls were the two council peers.

## Reviewed files

Main computed these SHA-256 values; this round did not recompute them. Paths are under `evaluation/x1-triage-ranking/`.

| File | SHA-256 | Change since round 4 |
|---|---|---|
| `execution/preparation.py` (integration) | `745823c76cac2a632468b366e0168284162645b094378283ce6f6bb1d2f982e5` | changed (was `1a116d86…`) |
| `execution/run-config-amendments/001-binding-v2.json` | `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121` | new |
| `execution/run-config.json` (integration) | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | unchanged |
| `execution/tests/test_core.py` | `0feb6447a24498389925b0c4c9cc96b4a91e7be001d92ae0016969b97e3f7f93` | amendment tests added |
| `execution/review_identifiers.py` | `adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2` | unchanged |
| `execution/identifiers/projection.py` | `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072` | unchanged |
| `execution/tests/test_identifiers.py` | `6b7f64b3d283180fae29c4bd54eb22f949e10bc9c85342f08f81d440053ea371` | unchanged |

The basis documents are unchanged from round 4. The old `preparation.py` bytes were not available, so the diff below is inferred from the new code. Everything that refers to amendments or effective configuration is new. Old journals without an `effective_configuration` field are handled as a legacy case at l.377–378, which confirms the field did not exist before.

## Findings

### What the change leaves alone

- **What the projector produces is unchanged.** An amendment may only touch `binding_record_kind`, `binding_acceptance_keys`, `binding_producer_contract` and `binding_recipe` (`preparation.py` l.74–75). `paths`, `required_files`, the `identifiers/*` record kinds and everything `run()` reads are outside that set.
- **The clock and config binding is unchanged.** `load_config` adds a check against the clock's original config path and pin (l.46–52) that matches what `read_clock` already enforces (l.142–167). The clock's `execution_config` pins the original `run-config.json`, whose bytes are unchanged.
- **Amendment 001 passes at load.** Its schema, `amends` (`execution/run-config.json`, `c817f678…`), authority reference and change keys all pass l.60–91. So the d run itself will not refuse because of it.
- **Lock and create-once behaviour is unchanged.** `operation_lock`, `exclusive_json`, the exclusive `root.mkdir()` and the seal reservation all mean what they did before.
- **The existing journals are still accepted.** The seven journal pairs from actions a and b have no `effective_configuration` field (checked by grep). The start/terminal comparison at l.375 uses `.get`, so both sides are `None`, and the new check at l.377–378 is skipped.
- **The config now copied into journals contains nothing seal will re-pin.** `run-config.json` has no `{path, sha256}` objects. `verify_tree_pins` will therefore only re-check the absolute amendment pins inside `effective_configuration`.

### R5-01 (issue): running d while the unadopted amendment 001 is present can permanently block seal

**What goes wrong:**
1. `review_identifiers.run()` calls `load_config`. `load_config` applies every `*.json` in `run-config-amendments/` (l.56–94) and does not check for a review; that check only happens at seal (l.103–111, called at l.284). So the d run today applies 001, which `preparation/model-free-binding-amendment.md` l.3 calls "Proposed; unadopted; no execution authority".
2. `Journal.__init__` (l.188–191) writes `effective_configuration` into d's start and terminal records. That field holds 001's pin (absolute path, SHA-256 and byte count), the amended config value and 001's `effective_sha256`. Both records are create-once.
3. At seal, `collect_seal` checks each journal in two ways:
   - its amendment list must be a prefix of the amendments present at seal time (l.379–382, "journal amendment chain differs");
   - every pin in the journal is checked again (l.386, then l.235–246, "pin mismatch" or `FileNotFoundError`).

**When it triggers.** d runs with 001 present, and 001 does not then get signed off with identical bytes at the same path. That covers any byte edit during 001's review (even the `reason` text or the contract string), withdrawal, renaming, or a replacement that amends `run-config.json` directly. A later 002 cannot reverse 001, because l.76–91 only accept the v2 values. So rejecting v2 would mean deleting 001, which also triggers this.

**Why this is plausible.** 001 is still under its own council review, and rounds repeat until there are zero issues. Edits during review are normal in this study: the binding implementation alone has had two rounds.

**Why it cannot be fixed by rerunning.** The d journal pair is immutable. A second d run would fail at `root.mkdir()` and write a `refused` journal, which also blocks seal (l.370). Trying to seal leaves a permanent reservation (l.411–428). Recovery would need a reviewed recovery procedure.

**Which orderings are safe.**
- **d runs with no amendment files present.** Its journal records `amendments: []` and the base-config digest. An empty list is a prefix of any later chain (l.381), and the digest falls back to the base config (l.383), whatever happens to 001.
- **d runs after 001 is signed off and frozen.** The prefix check holds, and a correctly chained 002 added later also passes.

**Required fix (procedural; no projector or `preparation.py` change).** Make one of these conditions a recorded precondition of the single d run:
- **(a) Keep proposals out of the directory.** `execution/run-config-amendments/` holds no `*.json` when d runs. The 001 proposal is kept outside `execution/` until it is adopted.
- **(b) Run d only after 001 is adopted.** 001 has been signed off with zero issues, `preparation-run/reviews/binding-v2-amendment.json` exists, and `verify_amendment_reviews` passes against the bytes present.

The run record must name which condition held and list the directory contents.

## Suggestions (not issues)

1. **Seal tests for the journal lifecycle.** Add synthetic tests for:
   - a legacy journal without the field being accepted;
   - a journal with no amendments being accepted under a later chain;
   - rejection after an amendment is edited (including a reason-only edit), deleted or renamed;
   - a correctly chained 002 being accepted;
   - the digest mismatch at l.384.

   `test_core.py` only reaches the missing-receipt refusal (l.207–212). `test_identifiers.py` l.99 replaces `load_config` with a plain dict, so d is never run against a real amendment directory.
2. **Nothing enforces "append-only".** Nothing in the code stops an existing amendment from being edited in place; the edit only shows up later as a seal refusal. A malformed or editor-backup `*.json` in the directory makes d refuse before it writes a journal, but a valid stray file is applied without warning.
3. **Design point for 001's own review.** `load_config` applies an amendment whose own text says it is not authority before review. One option is to apply only amendments that have a receipt, or to check receipts before any journal is created. This belongs in the amendment's review, not in d.
4. **Accept and record the asymmetry.** The a and b journals, and a d journal with no amendments, bind only the base config. That is sound only because amendments cannot reach anything outside the binding contract (l.74–75), and the d run record should say so.

## Recommendation, alternatives, risks, execution

- **Recommendation.** Keep the projector bytes as they are. Before the run, close R5-01 by adopting fix (a) or (b) as a written precondition. Fix (a) is preferred: it does not make d wait on the binding-v2 review, and it is robust whatever that review decides. Then hold round 6 to confirm the precondition and the digests.
- **Alternatives.**
  - *Run d now with 001 present and rely on 001 being signed off unchanged.* Rejected: a single edit needs reviewed recovery.
  - *Loosen the seal checks at l.381/l.386.* Rejected: it weakens provenance, and the independent pin re-check would still fail.
  - *Code change so `load_config`, or `Journal`, refuses amendments without a review before creating anything.* Valid, but it changes `preparation.py`, needs its own review, and overlaps suggestion 3.
- **Risks.**
  - All findings come from reading the code [INFERENCE]; nothing was reproduced.
  - The diff against `1a116d86` is inferred, not byte-compared.
  - Fix (a) depends on the operator's discipline unless the directory contents are recorded in the run record.
  - The d run itself stays irreversible once its outputs are reserved.
- **Execution outline.**
  1. Main chooses (a) or (b). For (a), 001 is moved out of `execution/run-config-amendments/` until it is adopted. That changes the execution source closure, which is checked at seal, not at d.
  2. Main records the precondition, the directory listing and the unchanged digests for `preparation.py`, `run-config.json` and the projector files.
  3. Round 6 confirms it with zero issues.
  4. Main runs the projector once at its canonical paths with the pinned CPython.
  5. The retained outputs and journal go to the separate completeness/delta output review.

## Verification

- **Read in full:** `preparation.py` (496 lines), `review_identifiers.py` (318 lines), `run-config.json`, `001-binding-v2.json`.
- **Read in part:** `test_core.py` amendment tests (l.158–215), the round-4 record, and the header of `model-free-binding-amendment.md`.
- **Run state:** listing of `preparation-run/` (clock present; seven journal pairs for actions a and b; no `identifiers/` directory; no amendment receipt). Grep confirmed no existing journal has `effective_configuration`.
- **Main's report:** `test_identifiers.py` 25 OK and `test_core.py` 21 OK. These are not reproduced here and do not exercise R5-01.
- **Peers.**
  - *innovation-challenger:* found the same issue independently, showed that the pin re-check at l.386 blocks a "relax the prefix check" fix, and contributed the sequence table and suggestion 1.
  - *experiment-peer:* found the same issue independently and ruled out the clock, legacy-journal, `verify_tree_pins`-on-config and lock hypotheses. It found the `test_identifiers.py` l.99 stub and contributed suggestions 2–4.
  - *Own analysis:* reached R5-01 before either peer reported. No peer changed the verdict; both strengthened the remedy, since fix (a) was shown to tolerate any outcome of 001's review.

## Verdict

**1 issue open (R5-01).** The projector must not be run yet. Round 6 must confirm a recorded precondition that either keeps `execution/run-config-amendments/` free of `*.json` files during the d run, or requires 001 to be signed off and frozen first, with zero issues. These digests stay under review:

- `execution/preparation.py` `745823c76cac2a632468b366e0168284162645b094378283ce6f6bb1d2f982e5`
- `execution/run-config-amendments/001-binding-v2.json` `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121`
- `execution/run-config.json` `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`
- `execution/tests/test_core.py` `0feb6447a24498389925b0c4c9cc96b4a91e7be001d92ae0016969b97e3f7f93`
- `execution/review_identifiers.py` `adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2`
- `execution/identifiers/projection.py` `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072`
- `execution/tests/test_identifiers.py` `6b7f64b3d283180fae29c4bd54eb22f949e10bc9c85342f08f81d440053ea371`
