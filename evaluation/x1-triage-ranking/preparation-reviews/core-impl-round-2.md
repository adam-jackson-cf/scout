# X1 preparation core implementation review — round 2

**Verdict: 1 issue open (X1C-R2-01, Moderate).** Core is not approved at `preparation.py` `858031133e1d28ecf71cb9b4e79b64df9ad03f572c2ab4eca32d5fc48f019255`. All seven round-1 issues are fixed as round 1 specified them. However, the X1C-R1-03 fix (prior-journal pins) adds a new way for the seal to become permanently impossible. The executed action (a) evidence, the action (d) evidence and the input-gate evidence remain valid.

**Scope and method.** This review was read-only: no code, tests, hashes, Git commands, adoption or seal were run, and no models were called. Three analyses fed it:
- **My own trace** of the five Core files, the new regression file, `RESULTS.md`, the eight executed journals and Homebrew's source code.
- **`innovation-challenger` (CoreR2Challenger):** an adversarial review of the receipt parser, recovery, preflight, shared-file exclusion and producer pins.
- **`experiment-peer` (CoreR2Peer):** for each round-1 issue, the observation that would show it is not fixed; whether each test fails for its stated reason; and whether the executed evidence still passes under the new code.

This record does not contain the exact round-9 verdict line or a whole-line scope marker, so `review_receipt` cannot accept it as authority.

## Files and SHA-256

The digests are Main's (`local://x1-packet-digests.md`). I did not recompute them.

| File (relative to `evaluation/x1-triage-ranking/`) | SHA-256 | Role |
|---|---|---|
| `execution/preparation.py` | `858031133e1d28ecf71cb9b4e79b64df9ad03f572c2ab4eca32d5fc48f019255` | Under review (changed) |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` | Under review (unchanged; executed) |
| `execution/fetch_objects.py` | `64bb3dd3b99d5227d0abd8e460f4dbabda18e15b37c16c71358241566e8c6a9f` | Under review (changed) |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | Under review (unchanged; pinned by the clock) |
| `execution/tests/test_core.py` | `7bdc801955675437a095121b28cfbf2d85ca75c004abbe870f1e14ca5a1d2628` | Under review (changed) |
| `execution/tests/test_preparation_evidence.py` | `634f5242d5d4190b5373bc741398a834cd3ee5296ec86a59b4748ce74b080aeb` | Under review (new regressions) |
| `execution/tests/test_binding_adoption.py`, `test_binding_seal.py`, `test_closure_review.py`, `test_closure_round2.py` | `0396765f…`, `4c3b58ea…`, `534e4754…`, `70efac05…` | Context (fixtures moved to the new receipt format) |
| `execution/review_identifiers.py`, `execution/identifiers/projection.py` | `adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2`, `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072` | Context: they equal the `source_pins` of executed (d) journal `906389d7…` |
| `preparation-reviews/core-impl-round-1-response.md` | `0d13e3356dcace31fa6d1c96db927d53347bcfbb7253b3a72c7a510d52f9c938` | Author response |
| `RESULTS.md` | `21b6fc13…` (author-reported, not in Main's list) | Deviations and runbook |

Main reports 273 Python tests passing, plus clean `go vet` and `go test -count=1`, at these bytes. I did not re-run them.

## Basis

- **Authority:** `reviews/round-9.md` `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc`, preparation actions (a)–(e).
- **Signed inputs, unchanged:**
  - `PREREGISTRATION.md` `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb`, §11 (l.218–228).
  - `reviews/rev9-digests.json` `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`.
  - `preparation/pins.json` `6d556fb5…`; exposure-recipe `c065347e…`; exposure-sources `67dc5a4b…`; frozen-exposure-inputs `ec92fba6…`.
- **Run records:** `preparation-run/clock.json` `26a1212a…` (start `2026-11-09T03:22:58.701909Z`) and `input-verification.json`.
- **Prior round:** `preparation-reviews/core-impl-round-1.md` (X1C-R1-01 to -07, S1–S7).

## Disposition of round-1 issues

| ID | Status | Evidence |
|---|---|---|
| X1C-R1-01 (Major) | **Fixed** | **Code** (`preparation.py`):<br>• `:423–447` requires a three-key receipt and a record under the study's `reviews/` or `preparation-reviews/` (`:429–433`).<br>• It requires the exact round-9 verdict as a whole line (`:435`; same text as `reviews/round-9.md:461`) and exactly one matching scope marker (`:437–440`).<br>• Each reviewed digest must appear inside that scope's block, with hex boundaries (`:442–446`).<br>• The seal keeps digests per scope (`:640–646`) and checks them against a fixed scope-to-evidence map (`verify_scoped_evidence` `:502–515`). That map matches the four `required_reviews` entries in `run-config.json`.<br>**Existing records:** none has a whole-line scope marker, and only `round-9.md:461` has the verdict line, so no record on disk passes today.<br>**Consumers:** `attest_closure.py:554–563`, `binding/acceptance.py:330/339/362` and `identifiers/cc2_delta.py:168–173` keep the two-argument call and the `reviewed` contract. `review_identifiers.py` does not call it.<br>**Tests:** `test_preparation_evidence.py:41–87` cover the real `identifiers-impl-round-5.md`, negated wording, another scope, a digest in another scope's block, a record outside the review directories, and cross-scope pooling. Each `assertRaisesRegex` matches the first refusal on its path. |
| X1C-R1-02 (Moderate) | **Fixed** | `:570` copies the list (`list(gate['pins'])`), and `:663` builds the predecessors from the gate pins only. Test `test_r102_predecessors_are_exactly_documentary_gate_pins` checks the exact list and that the gate object is unchanged. |
| X1C-R1-03 (Moderate) | **Fixed as specified** | **Code:**<br>• A failed or refused terminal needs `OUT/reviews/operation-recovery-<id>.json`, checked under the X1C-R1-01 rules. That receipt must bind the start record, the terminal record and every output pin (`:543–552`).<br>• Recovered failures go to `refusals_or_deviations`, not to `operations`.<br>• Every required record must have a passed producer (`:649–654`).<br>• New starts pin the journals that already exist (`:354`), and the seal checks those pins through `verify_tree_pins(value)` (`:540`).<br>• Fetch now refuses an existing receipt before it creates a journal (`fetch_objects.py:141–143`).<br>**Tests:** `:125–167` and `:300–309`.<br>**But:** the new prior-journal pin mechanism causes X1C-R2-01. |
| X1C-R1-04 (Moderate) | **Fixed** | **Code:** `verify_input_receipt` (`:394–406`) drops only pins whose absolute path is one of the five `SHARED` files, and the seal passes the configured ROOT (`:581`). The fresh `documentary_gate` (`:569`) still checks the owned extracts and the HD9 history.<br>**Peer checks:**<br>• In the executed receipt, the shared-file exception rows carry no `sha256`.<br>• No other record under `preparation-run/` pins a shared file.<br>• `RESULTS.md` is not in the manifest.<br>**Test:** `:169–177` refuses a change to an immutable file. |
| X1C-R1-05 (Minor) | **Fixed** | `seal --check` opens the existing lock read-only and exclusive, then runs `collect_seal` with no reservation, journal or write (`:247–250`, `:670–673`). No write path was found anywhere in the collection, including the closure re-check and the Git inspection. `RESULTS.md:21–28` makes the preflight mandatory. `brew pin git` is judged below. |
| X1C-R1-06 (Minor) | **Fixed** | **Code:** `producer_pins()` (`:322–336`) pins the resolved entry script and every loaded module under EXEC. Start and terminal records must agree on these pins (`:537`). Passed journals that lack them become `producer bytes unrecorded` deviations (`:556–559`).<br>**Records:** `RESULTS.md:7–13` names exactly the eight journals on disk (five a, two b, one d).<br>**Tests:** `:189` and `:199`. |
| X1C-R1-07 (Minor) | **Fixed for the listed paths** | Every refusal round 1 enumerated now has a test that asserts its specific message (`:207–276`, `:296–346`). Two bare assertions remain (S6). |

## Effect on executed evidence

- **Clock, input receipt and the five object receipts: unchanged and valid** (as round 1 found). `read_clock` still accepts the clock's pins of config `c817f678…`, authority, manifest and input receipt. After shared paths are dropped, the input receipt's remaining pins are immutable, including `verifier_source` `7d6f89ce…`.
- **Eight legacy journals: accepted, with one deviation each.**
  - None has `prior_journals` or `producer_pins`. The (a) and (b) journals also lack `effective_configuration`.
  - The start/terminal comparison sees `None` on both sides.
  - `verify_journal_configuration` returns early when a journal has no effective configuration and nothing was adopted before it started.
  - The (d) journal's "no-adopted-amendments" state matches.
- **(d) pins still match.** Journal `906389d7…` pins `review_identifiers.py` `adc42d38…` and `projection.py` `2408ed45…`, and both equal Main's current digests. The identifier records link to that passed journal through its `output_pins`.
- **No journal has been written since the fixes.** `preparation-run/journal/` holds exactly the eight legacy pairs, and there is no `reviews/` or `config-adoptions/` under OUT. So nothing create-once depends on the current Core bytes yet, and fixing X1C-R2-01 invalidates nothing.

## New issue

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| X1C-R2-01 | Moderate | `preparation.py:343–358` (`Journal.__init__`), `:361–366` (`finish`); `verify_inputs.py:33–44` (`exclusive_json`); seal check `preparation.py:540` | **Two journals running at the same time can make the seal permanently impossible.** This race came with the X1C-R1-03 fix.<br>**Mechanism:**<br>• `Journal` takes the predecessor snapshot (`prior_journals`: glob plus `pin` of every `journal/*.json`, `:354`) under the **shared** operation lock. It also publishes its start record (`:358`) and its terminal record (`:366`) under that shared lock.<br>• `exclusive_json` creates the final filename with `open('x')` before it writes and flushes the content. New start records also exceed the text buffer, so a reader can see a partial file.<br>• So operation B can pin operation A's start or terminal record while it is empty or half-written. A then completes its write, and B's immutable start keeps a digest that will never match.<br>**Consequence:** at seal, `verify_tree_pins(value)` (`:540`) refuses before the recovery branch runs. No code path can repair an immutable start record. `seal --check` detects the mismatch but cannot fix it.<br>**Why it is plausible:** concurrent journals are the established pattern. The five (a) starts fall within 23 ms (`03:23:04.443223Z`–`.466597Z`). The shared lock is designed to let operations run in parallel. Concurrent closure preflights, `journal`-wrapped helpers and binding runs are all permitted. | 1. Make taking the snapshot and publishing a record mutually exclusive across processes. Hold a short **exclusive publication lock**:<br>&nbsp;&nbsp;• in `__init__`, around the glob, the pins and `exclusive_json(start)`;<br>&nbsp;&nbsp;• in `finish`, around `exclusive_json(terminal)`.<br>Always acquire it after the operation lock, and do not upgrade the operation lock (`:243–244`). The lock must not leave an unaccounted file under OUT, because `:497` exempts only two names. A `flock` on a descriptor of the `journal/` directory, or a lock path outside OUT, both satisfy this.<br>2. Add a deterministic regression. Two writers overlap, with A's publication paused after the file is created. Assert that B waits, and that B's `prior_journals` pin equals A's final bytes. Cover both a start and a terminal publication. Both tests must fail on the current code.<br>3. Land the fix before any new journal (S5). |

**Severity.** The challenger rated this Moderate and the peer Minor. I rate it Moderate:
- The fix introduced it.
- The failure is irreversible: the seal is lost, with no recovery path.
- The code explicitly allows the concurrency that triggers it.

## Suggestions (not required for approval)

- **S1. Harden the receipt parser.**
  - **Example text.** The challenger rated this Major. I rate it a suggestion. The round-1 fix was delivered exactly as specified, and no record on disk can trigger it. A misfire needs a non-approval council record that contains, as whole lines, the exact verdict and a scope marker naming the current reviewed digest. That would be a fenced example of the format.
  - **Cheap hardening:** ignore fenced and quoted lines, require exactly one verdict line, exclude `*-response.md` (author responses share `preparation-reviews/`), take STUDY from the config rather than from where the receipt sits (`:429–430`), and read with `encoding='utf-8'`.
  - **Tests to add:** a duplicate scope marker ("ambiguous"), and a verdict phrase embedded in a longer line.
  - **Convention:** council records should never quote the verdict or scope marker as whole lines.
- **S2. Interrupted operations cannot be recovered.**
  - SIGHUP, SIGTERM, SIGKILL or a power loss leave a start record with no terminal. Nothing in the code can recover from that: `adopt` refuses (`:288–289`) and the seal refuses (`:523–524`).
  - Round 1 accepted the refusal itself.
  - Fix: install SIGTERM/SIGHUP handlers that raise, so the existing `BaseException` paths write a refused terminal. Also extend `operation-recovery` to a start record with no terminal.
- **S3. A failed canonical run cannot be recovered.** A failed canonical (b) or (c) run leaves a refused record at its create-once canonical path (`attest_closure.py:845–854`, `binding/acceptance.py:437–442`). Core recovery accepts the journal but can never accept that record. Run the closure `--preflight` and the binding rehearsal first, and record this limit in the runbook.
- **S4. Recent deletions go undetected.** Deleting the newest journal pair before the seal goes unnoticed, because no later journal pins it. Anchoring journal IDs outside OUT as they are created (for example by committing them) would close this.
- **S5. The code is frozen from the first new journal onwards.** Producer pins, the closure's `own_sources` and the binding `source_pins` freeze every loaded EXEC module from the next journal of any kind. That includes `--preflight` and `journal` wrappers.
  - Land X1C-R2-01 and any adopted suggestions in one revision before any new journal.
  - The `journal` wrapper should also pin the resolved `argv[0]` of the command it runs.
- **S6. Tests.**
  - Add message regexes at `test_core.py:136` (`assertRaises(ValueError)`) and `:293` (`assertRaises(FileNotFoundError)`).
  - Replace the mocked `collect_seal` in `test_r105` with a no-write fixture that runs the real collection.
  - Relabel the `binding-v2-round-4.md` subtest: that record approves, but in the superseded format; it is not a non-approval.
- **S7. `RESULTS.md`.**
  - Define `$PY` (the pinned `.tools/venv/bin/python`), `$EXEC` and `$OUT`.
  - Give each deviation the time field required by preregistration l.223.
  - Correct the wording at l.9: the (d) journal does pin its entry script, through `source_pins`.
  - Document the `operation-recovery` receipt procedure.
- **S8.** Round-1 S1, S2, S4, S5 and S6 remain open as suggestions. The basename exemption is still at `:497`.
- **S9. Homebrew retention evidence.** See the judgement below.

## Judgement on `brew pin git`

**Appropriate, authorised and adequate for this fix.**
- **Authorised:** round 1's X1C-R1-05 required fix asked for it.
- **Changes nothing that matters to the study:** it changes host state only through Homebrew's pin record outside the keg, so no pinned byte, clock or study artefact changes. `brew unpin git` reverses it.
- **Protects the keg:** Homebrew's `brew cleanup` skips the pinned keg ([`formula.rb` `eligible_kegs_for_cleanup`](https://github.com/Homebrew/brew/blob/main/Library/Homebrew/formula.rb): "Skipping (old) … due to it being pinned"), and `brew upgrade` skips pinned formulae ([Manpage, `pin`/`upgrade`](https://docs.brew.sh/Manpage)).
- **Host checks (by the peers, against the local Homebrew install):** the pin exists at `/opt/homebrew/var/homebrew/pinned/git`; the Git rack holds only `2.56.0_2`, matching the object receipts; and reinstall and uninstall are refused while the formula is pinned.

**What it does not cover.**
- **Dependencies:** pinning git does not pin its runtime dependencies (pcre2, gettext and others, per the keg's `INSTALL_RECEIPT.json` as the peer read it). This does not threaten Core's seal, which re-checks only `bin/git` bytes and needs only local Git operations to work [INFERENCE].
- **The closure:** (b) will inventory git's native dependencies (`attest_closure.py:712–718`). An upgrade plus cleanup of those formulae between (b) and the seal could break the closure re-check [INFERENCE].

**S9 recommendations:**
- Before (b), record `brew list --pinned --versions`, the "Pinned" line of `brew info git` and `otool -L` of `bin/git`.
- Pin git's linked dependency formulae, or record a no-`brew upgrade`/no-`brew cleanup` discipline until the seal.
- Unpin after the seal.

## Verification

- Read-only. Main's 273-test Python run and clean Go results are cited, not re-run.
- The behaviour of the X1C-R2-01 interleaving is inferred from reading the code. Both peers derived it independently.
- Homebrew behaviour is cited from upstream source and docs; the host facts are as the peers reported them.

## Recommendation

Do not approve Core at these bytes. Fix X1C-R2-01 in `preparation.py` with its regression, folding in any adopted suggestions (S1 and S2 especially) in the same revision. Then hold a short round 3 limited to journal publication and the folded-in changes. No new journal of any kind may start before that round reports no open issues.

## Alternatives

- **A. Publication lock in `Journal` (recommended).** It is small, keeps parallel operation bodies, and creates no new artefact.
- **B. Atomic publication: temp file, `fsync`, `os.link` to the final name, unlink.** It keeps records create-once, but a crash leaves temp files under OUT that become unaccounted outputs. It needs more handling than A.
- **C. Run every operation one at a time under the exclusive lock.** It closes the race, but it removes the designed parallelism and changes how locks re-enter (`:243–244`).
- **D. Approve now and forbid concurrency in the runbook.** Rejected: the code permits the race, and its consequence is permanent.

## Evidence

- **Code:** `preparation.py:243–250`, `:322–366`, `:394–447`, `:497`, `:502–515`, `:519–561`, `:570`, `:640–673`; `verify_inputs.py:17`, `:33–44`; `fetch_objects.py:135–143`; `attest_closure.py:554–563`, `:712–718`, `:828–854`; `binding/acceptance.py:330–366`, `:394–443`.
- **Records:**
  - `preparation-run/journal/*.json`: eight pairs. The (a) start times run from `03:23:04.443223Z` to `.466597Z`. The (d) `source_pins` equal `adc42d38…` and `2408ed45…`.
  - `run-config.json` `required_reviews`.
  - `reviews/round-9.md:461`.
  - `RESULTS.md:5–28`.
- **External:** [Homebrew Manpage](https://docs.brew.sh/Manpage); [Homebrew `formula.rb`](https://github.com/Homebrew/brew/blob/main/Library/Homebrew/formula.rb).

## Risks

- Every further round uses time on the 45-day clock (deadline `2026-12-24T03:22:58.701909Z`).
- From the first new journal onwards, any change to Core is irreversible in effect (S5).
- The receipt-format change means (b), (c) and adoption cannot proceed until council records exist in the new format. That includes a binding v2 confirming round and a closure dependency round, as round 1 outlined. Do not rewrite historical records.
- Homebrew dependencies of git are not covered by the pin (S9).

## Execution outline

1. Implement the X1C-R2-01 publication lock and its overlapping-writer regressions, plus any adopted suggestions. Change no create-once artefact.
2. Main runs the full Python suite, `go vet` and `go test -count=1`, then recomputes the digests.
3. Hold Core round 3 at the new digests.
4. Hold the closure dependency round and the binding v2 confirming round. Each approval record must use the exact verdict line and per-scope marker blocks that name the full digests.
5. Record the Homebrew retention evidence (S9).
6. Write receipts citing those records, then run (b) preflight, (b), `adopt`, (c), `seal --check`, and finally `seal`.

## How the peers shaped this record

- **Challenger:**
  - Raised the concurrent-publication race (now X1C-R2-01) and set its severity.
  - Raised the example-text counterexample, which I downgraded to S1. The required fix was delivered as specified and no record on disk triggers it.
  - Confirmed that the seal's own journal and adoption create no self-pin cycle.
- **Peer:**
  - Found the same race independently and showed that start records exceed the buffer.
  - Confirmed under the new code, record by record, that the executed evidence (including (d)) is still accepted.
  - Supplied S3–S7 and the Homebrew dependency limits.
- **Mine:**
  - The interrupted-operation gap (S2), the closure-retention angle of the pin (S9), and the check that no new journal exists yet.
  - The severity decision and the lock-placement constraint at `:497`.

## Verdict

**1 issue open** (X1C-R2-01, Moderate). Core is not approved at:
- `preparation.py` `858031133e1d28ecf71cb9b4e79b64df9ad03f572c2ab4eca32d5fc48f019255`
- `verify_inputs.py` `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`
- `fetch_objects.py` `64bb3dd3b99d5227d0abd8e460f4dbabda18e15b37c16c71358241566e8c6a9f`
- `run-config.json` `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`
- `test_core.py` `7bdc801955675437a095121b28cfbf2d85ca75c004abbe870f1e14ca5a1d2628`
- `test_preparation_evidence.py` `634f5242d5d4190b5373bc741398a834cd3ee5296ec86a59b4748ce74b080aeb`

The executed (a) and (d) evidence and the input-gate evidence remain valid. No new journal, `adopt`, (b), (c) or seal may run until a later round reports no open issues at new digests.
