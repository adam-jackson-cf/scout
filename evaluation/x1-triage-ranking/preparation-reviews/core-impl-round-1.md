# X1 preparation core implementation review — round 1

**Verdict: 7 issues open (1 Major, 3 Moderate, 3 Minor).** Core is not approved at `preparation.py` `b1759ec7…`. The executed action (a) evidence and the input-gate evidence remain valid. Closure prerequisite P1 is still unmet.

**Scope and method.** This review was read-only: no code, tests, hashes, Git commands, adoption or seal were run. Three analyses fed it:
- My own trace of the five Core files and the executed records under `preparation-run/`.
- `innovation-challenger` (CoreChallenger): adversarial review of the seal, adoption, receipt checking and fetch isolation.
- `experiment-peer` (CorePeer): competing failure hypotheses, record-by-record checks, and whether each test catches what it is named for.

This record avoids the two approval phrases that `preparation.py:400` searches for, so the current check cannot mistake it for authority.

## Files and SHA-256

The digests below are Main's (`local://x1-packet-digests.md`). I did not recompute them.

| File (relative to `evaluation/x1-triage-ranking/`) | SHA-256 | Role |
|---|---|---|
| `execution/preparation.py` | `b1759ec710fa4b71dec1ef91800e84d285c733693b0cb5f4ffd7fe68f147c916` | Under review |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` | Under review; already executed |
| `execution/fetch_objects.py` | `737b5ec7fa734bacbf3bd6be0b56ac3a1b64041f7442c2e5aef17a311b8d1831` | Under review; (a) already executed |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | Under review; pinned by the clock |
| `execution/tests/test_core.py` | `4f4bfdaf50fc0d56ff775f8c86bb6bc579c932acc1a76764a55ba4a6f36da07e` | Under review |
| `execution/run-config-amendments-proposed/001-binding-v2.json` | `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121` | Context only |

Main reports 237 Python tests passing, plus clean `go vet` and `go test`, at these bytes. That was not re-run here.

## Basis

- **Authority:** `reviews/round-9.md` `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc`, preparation actions (a)–(e).
- **Signed inputs:**
  - `PREREGISTRATION.md` `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb`, §11 (l.218–228).
  - `reviews/rev9-digests.json` `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`.
  - `preparation/pins.json` `6d556fb5156e8ced3c17ae7526a841e51688145faad4a3f8a6a2ce0c939100fc`.
- **Run records:**
  - `preparation-run/clock.json` `26a1212a…` (start `2026-11-09T03:22:58.701909Z`).
  - `preparation-run/input-verification.json` `22d93b44…`.
- **Execution plan** `local://x1-prep-exec-plan.md` §1–§3 and §7: used as context, not as authority.

## Findings: what is correct

- **Clock** (`:183–196`, `:204–228`, `:359–372`):
  - It is created once under the exclusive lock, and `load_config` and `read_clock` check its config path and digest.
  - The deadline is exactly 45 × 86,400 s after the start. On the record: `2026-11-09T03:22:58.701909Z` → `2026-12-24T03:22:58.701909Z`.
  - It pins the authority, the manifest, the input receipt and the gitlink `734f81bc…`.
  - `start` checks `HEAD == execution_commit`. The recorded value is `5a8c0a15…`.
- **Journals and locks** (`:236–256`, `:322–347`):
  - Each operation gets an exclusive start record and an exclusive terminal record.
  - Locks re-enter through `ContextVar`. A shared lock cannot be upgraded to exclusive. Every operation refuses once a seal reservation exists.
  - `adopt` takes the exclusive lock (blocking) and refuses while any operation is unfinished. The seal takes the exclusive lock (non-blocking) before it reserves.
- **Amendment adoption** (`:63–151`, `:257–319`):
  - Proposed amendments are ignored.
  - Adoption copies only from the canonical proposed directory, in filename order, with a predecessor digest chain. Adoption artefacts are create-once, and adoption must be later than every existing journal.
  - Journals record the effective configuration at their start. The seal checks the chain as it stood at each journal's start (`:409–427`).
  - `run-config.json` stays byte-identical.
- **Input gate** (`verify_inputs.py`):
  - It checks the exact manifest, authority and preregistration digests, and that the ADR copy is byte-equal to the authority.
  - It checks all 131 rows. Only the five `SHARED` files may differ, and only if their owned extracts match.
  - The HD9 state is reversed only for the allowed transition. The transition line must be the exact `TRANSITION`. Earlier HD9 history is authenticated against the adoption snapshot `584da4a2…`, whose reversed bytes must hash to the signed value.
  - It checks the 577 source rows and the policy.
  - `deferred_to_b` applies only to an uninitialised, empty gitlink directory. The seal re-checks every deferred row against the current bytes (`:429–441`) and refuses if any deferral is still outstanding (`:480`).
- **Fetch** (`fetch_objects.py`):
  - Uses an absolute Git binary with a private HOME. System, global and replacement config are disabled, hooks are routed to `/dev/null`, and objects are fsck'd during transfer.
  - Fetches only literal refspecs, with no depth, filter, tags or submodules.
  - `inspect_store` refuses shallow repositories, alternates, grafts, promisor packs, symlinks, unexpected config and extra refs.
  - Requires an exact commit type and SHA, and runs `fsck --full --strict`.
  - Proves first-parent membership by walking commit headers (`:107–119`), with the merge-base check on top.
  - Takes the Django parent as `tip^1 == tip_header.parents[0]`.
  - Sets the dev C0 to `tip~2000` when exactly 2,000 edges exist, otherwise to the root, and cross-checks this against the `rev-parse` shortcut. Tests cover the merge, off-path and both dev cases (`test_core.py:457–472`).
- **Seal** (`:461–604`):
  - Reserves at one fixed path before validating. There is no path or subset argument and no approval boolean.
  - Requires every record in `required_records` to be passing and pinned to the clock.
  - Re-derives each object store, its boundaries, headers and fsck from the retained databases (`:502–505`).
  - Accounts for every file under OUT.
  - Where a journal also pins a canonical record, the canonical role is kept (`:444–451`), and duplicate journal pins are all verified.

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| X1C-R1-01 | Major | `preparation.py:394–406` (`review_receipt`); `:536–546` (seal); consumers are `adopt` `:276–279` and `attest_closure.py:554–557, :835` | **The receipt check accepts records that are not approvals, and refuses records that are.** `:400` only needs both approval phrases somewhere in the text, and `:404` only needs each digest somewhere. The receipt's `scope` is never matched against the record. At seal, `:541`/`:545` pool the digests of all four scopes together.<br>**Wrong record accepted:** `preparation-reviews/identifiers-impl-round-5.md` names the current proposal digest `88eb95a5…` (l.12, l.106) and contains both phrases (l.53), but its verdict is "1 issue open (R5-01)" (l.103). A receipt `{scope: model-free-binding-v2-amendment, record: <that file>, reviewed: [pin(proposal)]}` passes `review_receipt` and the digest gate in `adopt`. The peer reports similar cases in `closure-impl-round-4.md` (l.26–27, 157, 165) and `identifiers-impl-round-9.md` (l.8, 131–139).<br>**Right record refused:** `binding-v2-round-4.md` is the record Main has been told to cite for `closure-model-free.json` and `binding-v2-amendment.json` (l.167–170). It contains the first phrase (l.18, l.188) but not the second; it says "signs" and "Sign-off" (l.192). So `reviewed_go_sources` and `adopt` would refuse it today, before any record is written. | 1. Accept only records under `STUDY/reviews/` or `STUDY/preparation-reviews/`.<br>2. Require the exact verdict line used at `reviews/round-9.md:461`, as a whole line.<br>3. Require one or more whole lines `Signed-off scope: <scope>`. The receipt's scope must have such a line, and every reviewed digest must appear in full between that line and the next `Signed-off scope:` line (or the end of the file).<br>4. Replace the pooled seal check with a constant map from scope to evidence: dev object receipts → `resolved-development-boundaries`; binding records → `binding-acceptance`; identifier records and `familiarity.jsonl` → `identifier-completeness-delta`; source manifest, `closure.json` and the test object receipts → `execution-implementation`.<br>5. Add negative tests using both real records above, negated wording, a record for another scope, a record outside the review directories, and cross-scope pooling.<br>6. Every pending authority record, binding v2 included, must carry this block. |
| X1C-R1-02 | Moderate | `preparation.py:468`, `:577` | `outputs, sources = [], gate['pins']` makes `sources` the same list as the gate's pins. Everything later added to `sources` (amendment pins, every pin inside every record, the whole source manifest, the Homebrew git keg) therefore ends up in `predecessors` with `kind: signed-input`. The seal is create-once, so this false provenance would be permanent. | Use `sources = list(gate['pins'])`, and add a test that `predecessors` equals exactly the documentary-gate pins. |
| X1C-R1-03 | Moderate | `preparation.py:549–556`, `:581`; `fetch_objects.py:141–144`, `:198`; `journal` subcommand `:655–664` | **After any failed operation, the preparation can no longer be sealed with all its records kept.** Any terminal journal that did not pass refuses the seal, and `refusals_or_deviations=[]` is hard-coded. No code path implements the "reviewed recovery" that earlier rounds rely on (`closure-impl-round-2.md:65`, `closure-impl-round-3.md:103`, `identifiers-impl-round-5.md:45`). The completeness check (`:551`) only compares the journal files that are present, so deleting a pair goes unnoticed. This conflicts with preregistration §11 l.223: report every failed scenario and append Deviations.<br>Each of these leaves such a journal: a duplicate `fetch_objects.py --repo …` (the receipt-exists check at `:143–144` runs after the journal is reserved at `:141`); a transient network failure during (b); a `journal`-wrapped command that exits non-zero. Once (b) pins `preparation.py`, the code fix would invalidate `closure.json`. | 1. The seal accepts a terminal journal that did not pass only when a receipt with scope `operation-recovery` (under the X1C-R1-01 rules) names its start and terminal records and all its output pins.<br>2. Such journals and their receipts go into `refusals_or_deviations`, not `operations`. Each record in `required_records` must still come from a passed operation.<br>3. Each `Journal` start record pins the journal files that already exist, as `adopt`'s `prior_journals` does (`:288`). The seal refuses if any pinned prior journal is missing.<br>4. Move fetch's receipt-exists check before the `Journal` is created.<br>5. Add tests. |
| X1C-R1-04 | Moderate | `preparation.py:375–377`, `:479`, with `verify_inputs.py:157` | The input receipt pins the whole bytes of all five `SHARED` files as they were at gate time, and the seal re-checks them. Round 9 §13 condition 2 allows edits outside the owned extracts, and the ledger rules require such edits. So any ledger or STATUS edit in this worktree before sealing refuses the seal. Because the seal reserves before it validates (`:586`), that refusal is permanent. | `verify_input_receipt` skips pins whose path is `ROOT/<path in verify_inputs.SHARED>`. The seal's fresh `documentary_gate` (`:467`) already re-checks the owned extracts and HD9 history and supplies the current pins. Add a test. `verify_inputs.py` does not change. |
| X1C-R1-05 | Minor | `preparation.py:584–604` | Nothing validates before the seal reserves, so any drift since the run causes a permanent refusal. Concrete cases:<br>• Every object receipt and fetch journal pins `/opt/homebrew/Cellar/git/2.56.0_2/bin/git`, which is re-checked at `:502–505` and by `verify_tree_pins`. The peer cites Homebrew's periodic cleanup ([Homebrew FAQ](https://docs.brew.sh/FAQ); not checked by me).<br>• A stray file such as `.DS_Store` under EXEC breaks the source-closure equality (`:529–533`). | Add `seal --check`: run `collect_seal` under the exclusive lock with no reservation, no journal and no writes. The runbook requires it to pass immediately before `seal`. Run `brew pin git` until the seal and record that this was done. |
| X1C-R1-06 | Minor | `Journal` `:322–347`; `fetch_objects.py:141` | No journal records which code ran. The fetch journals pin only the config and git. The (a) and (b) journals lack `effective_configuration`, so an earlier `preparation.py` wrote them. At seal, the source manifest will describe the code as it is then, so the seal would imply that the reviewed bytes produced (a), (b) and (d). | 1. `Journal` start records pin the resolved entry script and every loaded module file under EXEC.<br>2. The seal lists each passed journal that lacks these pins in `refusals_or_deviations`, with the reason "producer bytes unrecorded".<br>3. Record the same under Deviations in RESULTS. |
| X1C-R1-07 | Minor | `tests/test_core.py` | Several refusals are untested, or tested in a way that a different refusal satisfies:<br>• `:295` passes because of a pin mismatch, raised before the scope check runs.<br>• `:286` stops at `FileNotFoundError`.<br>• `:129` covers only a reservation that already exists.<br>• The `collect_seal` journal refusals (unfinished, failed, id/stem mismatch, stale clock, timeline before the clock, start/terminal mismatch, missing a–d) have no tests.<br>• No integrated test of `collect_seal` with an unaccounted file.<br>• Untested `adopt` refusals: adopting twice, reservation already present, unfinished operation, filename order.<br>• Re-creating the clock and the `read_clock` refusals are untested.<br>The peer reports these; I confirmed `:286`, `:295` and the fixture's council text at `:242`. | Add one negative test per refusal path, using `assertRaisesRegex` on the specific message, plus tests for the fixes in 01–06. |

## Effect on executed evidence

- **`clock.json`: valid.**
  - Pins: config `c817f678…`, authority `6f1d6ca2…`, manifest `2253911b…`, input receipt `22d93b44…`, commit `5a8c0a15…`, gitlink `734f81bc…`, `first_action` a.
  - The deadline is exact, and the clock predates the first journal (`03:23:04.443223Z`).
- **`input-verification.json`: valid.**
  - Status passed with no failures.
  - 126 exact rows plus 5 shared-file exceptions make 131. The exceptions are exactly the `SHARED` set.
  - 577 source rows and 713 pins.
  - 176 `deferred_to_b` rows, matching the contract's `working_file_count`.
  - `verifier_source` is `7d6f89ce…`, which matches the current file. The interpreter's digest is the pinned CPython `570b4105…`.
  - None of the issues touches `verify_inputs.py` logic. X1C-R1-04 affects only the future seal check.
- **`objects/*.json`: all five valid.**

  | Repository | Tip | C0 | C0 rule | Edges |
  |---|---|---|---|---|
  | django | `a05e16a7…` | `d88ec42b…` | fixed | 2791 |
  | TypeScript | `6161772c…` | `f2e9ebd1…` | fixed | 1666 |
  | prometheus | `ef78ff64…` | `756202aa…` | fixed | 3248 |
  | fastapi | `d666ccb6…` | `e6da96fb…` | 2000 first-parent edges | 2000 |
  | axum | `562fa206…` | `07294378…` | root | 1086 |

  - Every tip and C0 equals the pinned table.
  - The Django parent `f4c65fd6…` equals `tip_header.parents[0]`.
  - All four store booleans are true, and fsck exited 0. The challenger and peer report that `tip~2000` exited 0 for fastapi and 128 for axum, consistent with the derivations above.
  - X1C-R1-06 means the fetch code's bytes are not recorded. The evidence still stands, because the seal re-derives every boundary, header and fsck result from the retained stores using reviewed code, and the command logs record the exact argv. This needs only the deviation disclosure in X1C-R1-06.
- **`journal/*`: 8 pairs, all passed and valid.** That is five (a), two (b) and one (d), each with exit 0, matching ids and the clock pin `26a1212a…`.
- No issue invalidates an executed receipt. The fixes are all forward-looking, and nothing create-once is touched.

## Suggestions

- **S1.** `:456` exempts `.operation.lock` and `preparation.seal.json` by name at any depth. Exempt only those exact paths at the OUT root. This is cheap to fold into the fix round.
- **S2.** `inspect_store` should refuse `commondir`, `config.worktree` and `worktrees/`, and assert that `rev-parse --git-common-dir` equals the database. The challenger rated this Major. I rate it a suggestion: content addressing means a redirected store can still serve only the pinned objects, so no false boundary can result; what remains at risk is the no-alternates retention claim.
- **S3.** The seal should assert `django_first_parent.sha == tip_header.parents[0]`, and that each `start_record.path` is the canonical `<id>.start.json`.
- **S4.** `adopt` should apply `load_config`'s authority-path validation (absolute path or `..`) before writing anything. Today such a path is adopted and only refused afterwards, permanently.
- **S5.** `adopt` should hash the proposal bytes once and use that digest for both the review check and the copy.
- **S6.** If `Journal('e')` construction fails after the reservation, no record of the reason is left. Record it in a refused journal or a failure file.
- **S7.** `verify_adoption:89` ties every adoption to the current `preparation.py`. Land every Core fix before `adopt` and before the first `attest_closure` journal; this is the existing P1 condition.

## Recommendation

Do not approve Core at `b1759ec7…`. Fix X1C-R1-01 to -07 in one Core revision before any `adopt`, closure, binding or seal journal, then iterate to a zero-issue round 2. Nothing already executed needs repeating.

## Alternatives

- **A. Fix now, before (b) and adopt.** This is the recommendation.
- **B. Approve now and handle the issues by runbook discipline.** Rejected. X1C-R1-02 corrupts a create-once record whatever Main does. X1C-R1-01 already misfires on a real record (`binding-v2-round-4.md`).
- **C. Fix after (b) or (c).** Rejected. The closure pins `preparation.py`, so a later change invalidates `closure.json` and any adoption (`closure-impl-round-5.md:105, 120–124`).
- **D. A separate council-issued JSON decision file (challenger's option).** Rejected. Main would write that file too, so it is no more authentic. A fixed block inside the council record gives the same checkability without a new kind of artefact.

## Evidence

- Code: `preparation.py:89, 375–377, 394–406, 444–457, 461–581, 584–604`; `fetch_objects.py:48–105, 107–119, 125, 141–144, 198`; `verify_inputs.py:73–100, 141–200`.
- Records: `preparation-run/clock.json`; `input-verification.json` (summary fields above); `objects/*.json`; `journal/85d89edc….json` (source pins are config and the git keg only, and there is no `effective_configuration`).
- Review records: `identifiers-impl-round-5.md:12, 53, 103, 106`; `binding-v2-round-4.md:18, 167–170, 188, 192`; `reviews/round-9.md:461`; `PREREGISTRATION.md:223`; `closure-impl-round-5.md:105, 120–124`; `closure-impl-round-3.md:127` (S9 had already noted the substring check).

## Risks

- The fix round, the closure dependency round and a binding v2 confirming round in the new format all use time on the 45-day clock.
- Adoption and the seal are irreversible, so the fixes must land first.
- The Homebrew git keg must stay unchanged until the seal (X1C-R1-05).
- The record-level behaviour is my inference from reading the code, not from running it. That covers the counterexample's acceptance, the refusal of `binding-v2-round-4.md` and the peers' reports of command exit codes.

## Execution outline

1. Implement fixes 01–06 in `preparation.py`, plus the receipt-check move in `fetch_objects.py` (and S1–S6 if adopted), and add the 07 tests. Leave `verify_inputs.py` and `run-config.json` unchanged.
2. Main runs the full Python suite, `go vet` and `go test -count=1`, then recomputes the digests.
3. Hold Core round 2 and iterate to zero issues. Its record uses the new sign-off block.
4. Hold a short closure dependency round for the new digests, and a binding v2 confirming round whose sign-off block covers `model-free-binding-v2-amendment` and `closure-model-free-go`.
5. Run `brew pin git` and record it.
6. Continue `binding-v2-round-4.md`'s outline: receipts citing conforming records, then (b), then `adopt`, and the rest.
7. Immediately before sealing, `seal --check` must pass; then run `seal`.

## How the peers shaped this record

- **Challenger:**
  - Supplied the verified `identifiers-impl-round-5.md` counterexample, which raised X1C-R1-01 to Major and added the scope-in-record requirement.
  - Its `commondir` finding became S2, and its basename finding became S1; I had found the latter myself.
- **Peer:**
  - Changed my view on failure accounting from an accepted design to an issue (X1C-R1-03), citing preregistration l.223 and the absence of any recovery path.
  - Added X1C-R1-06 and X1C-R1-07.
  - Corrected my belief that the shared-file mismatch could be recovered: the seal reserves before it validates, which made it X1C-R1-04 and motivated X1C-R1-05.
- **Mine:**
  - The `binding-v2-round-4.md` wording gap.
  - The `predecessors` aliasing, which the peer found independently.
  - S4–S6.

## Verdict

**7 issues open** (X1C-R1-01 Major; X1C-R1-02, -03 and -04 Moderate; X1C-R1-05, -06 and -07 Minor). Core is not approved at `preparation.py` `b1759ec710fa4b71dec1ef91800e84d285c733693b0cb5f4ffd7fe68f147c916`, `verify_inputs.py` `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`, `fetch_objects.py` `737b5ec7fa734bacbf3bd6be0b56ac3a1b64041f7442c2e5aef17a311b8d1831`, `run-config.json` `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` or `test_core.py` `4f4bfdaf50fc0d56ff775f8c86bb6bc579c932acc1a76764a55ba4a6f36da07e`.

The executed (a) and input-gate evidence remain valid; X1C-R1-06 must be disclosed as a deviation. Closure prerequisite P1 is not met. No `adopt`, (b), (c) or seal may run until a later round reports zero open issues at new digests.
