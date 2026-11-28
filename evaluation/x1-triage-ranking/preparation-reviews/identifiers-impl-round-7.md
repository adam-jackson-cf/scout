# X1 preparation (d) identifier projector — implementation review — round 7 (integration)

**Scope.** This round checks only the R6-01 fix. It asks two things:
- Does a d journal written now, before any adoption, survive a later `preparation.py adopt` of 001 followed by seal?
- Does the change add a concrete, plausible failure?

Identifier completeness, the delta, and whether amendment 001 is a good amendment are out of scope. I only read files. I did not run the projector, adopt, seal or the tests. The only model calls were the two council peers.

## Reviewed files

Main supplied these SHA-256 values. This council did not recompute them. Paths are under `evaluation/x1-triage-ranking/execution/`.

| File | SHA-256 | Since round 6 |
|---|---|---|
| `preparation.py` | `599a128d64dbd0617e7110daddf77c3b1b8677e516e4f0c9b3cf6a1ea389e427` | changed (was `e883abc4…`) |
| `tests/test_core.py` | `94afb7638acbb0316cc019077a87d924fda0e665b752d3d57acdfadbd9657ace` | changed (was `2fa57759…`) |
| `run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | unchanged |
| `run-config-amendments-proposed/001-binding-v2.json` | `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121` | unchanged |
| `review_identifiers.py` | `adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2` | unchanged |
| `identifiers/projection.py` | `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072` | unchanged |
| `tests/test_identifiers.py` | `6b7f64b3d283180fae29c4bd54eb22f949e10bc9c85342f08f81d440053ea371` | unchanged |

I also read two files outside this round, because they share the adoption contract: `binding/acceptance.py` (lines 309–335) and `preparation/model-free-binding-amendment.md` (line 66).

## R6-01: resolved

**What the code now does to stop backdating:**
- **The adopt command writes the adoption timestamp, and only it.**
  - `adopt` takes the exclusive lock and waits for it (`preparation.py:262`).
  - It refuses if any journal has no terminal record (`:284–287`).
  - It takes `adopted_at = now()` once. It refuses if any existing journal's `started_at` or `ended_at` is at or after that time (`:289–294`).
  - It refuses if the adoption has already been reserved (`:297–298`).
  - It then writes the copied amendment, the operation record and the receipt itself (`:301–317`).
- **Every load re-checks the adoption.** `verify_adoption` (`:68–100`) runs on every `load_config` and again at seal (`verify_amendment_reviews`).
  - It accepts only the command's schema. The receipt and command must pin each other, and the timestamps must be equal.
  - The writer and config pins must match, and so must the argv.
  - It re-checks every `prior_journals` record for being strictly earlier than `adopted_at`.
  - A hand-written or backdated receipt is therefore refused (`test_core.py:300–313`).
- **Load checks existing journals before anything new is written.** `load_config` runs `verify_journal_configuration` against every existing start record (`:148–149`). This happens before any new record is written, so a refusal can be recovered from.
- **No interleaving.** `Journal.__init__` reloads the config, captures `started_at` and writes the start record all inside the operation lock (`:326–339`). An exclusive adopt therefore cannot fall between those steps (`test_core.py:315–338`, which uses a real `Journal('d')`).

**What happens to d if it runs now:**
1. `review_identifiers.py:251` loads the config. `run-config-amendments/` is empty, and the seven legacy a/b starts have no `effective_configuration`, so they return early (`:414–416`).
2. Line 258 takes the shared lock. The `Journal` lock re-enters through the ContextVar (`:239–245`).
3. The journal records `amendments: []`, `no-adopted-amendments`, and the digest of the base config.

**What happens at a later adopt and seal:**
- `adopted_at` is later than d's `ended_at`, so the expected set at `:412` is empty. That matches d's record at `:418–424`.
- The prior-journal re-check passes, because the journals are create-once.
- Timestamps are compared as parsed UTC datetimes, never as strings (`:26–30`). All seven a/b pairs are terminal, and the latest `ended_at` is `2026-11-09T03:35:20.727619Z`.

**Result:** d's journals stay valid through adopt and seal. I found no failure in the d run or in the legacy a/b journals.

## New issues

### R7-01 (issue): the adopt order makes 001's binding acceptance impossible, so no later adopt → binding → seal can complete

**The conflict:**
- `adopt` validates 001's authority receipt (`preparation-run/reviews/binding-v2-amendment.json`) through `review_receipt` (`preparation.py:277–279`). That happens *before* it copies the amendment (`:301–304`).
- `review_receipt` re-pins every `reviewed` path, so every path in the receipt must already exist (`:394–405`).
- Binding v2 requires that **same** receipt to list the adopted copy at its adopted path, `EXEC/run-config-amendments/001-binding-v2.json` (`binding/acceptance.py:316–327`).
- The adoption permanently pins the receipt (`preparation.py:313–317`). Every load then rejects any change to it (`:79–81`).

**Why every workaround fails:**
- If the receipt lists the adopted path, adopt fails with `FileNotFoundError`, because the copy does not exist yet.
- If the receipt lists only the proposal, adopt succeeds, but binding then refuses with "does not bind recipe, amendment and implementation".
- If the receipt is edited after adoption, every `load_config` refuses.
- If the copy is placed by hand first, adopt's first `load_config` (`:263`) fails on the missing adoption receipt, and `:297` refuses anyway.

**A second circularity with the same root [INFERENCE].** The same receipt must also bind the source manifest (`acceptance.py` `required`). Seal requires that manifest to list exactly the files in EXEC (`preparation.py:514–516`), and after adoption that includes the adopted copy.

**Consequence.** Seal needs passing binding records (`:466–470`, `:492–498`). With these bytes, no sequence of adopt, binding and seal can succeed. d's journals are not affected, but the "later adopt + seal" this round asks about cannot happen. The round-6 procedure, where Main copied the file and then wrote the receipt, did not have this ordering conflict.

**Required fix (choose one and record which):**
- **(a) Preferred: change only `binding/acceptance.py`, in the binding track.** Binding round 2 still has 4 issues open, so that file is changing anyway.
  - Accept the reviewed **proposal** pin, with a digest equal to the adopted copy's.
  - Also require the adopted copy's path and digest in `config.amendments`, with `adoption_review == pin(binding-v2-amendment.json)`.
  - Bind the post-adoption source manifest through the post-adoption `reviews/implementation.json`, not through the pre-adoption amendment receipt.
  - None of the round-7 digests changes.
- **(b) Separate the two authorities.** Keep the immutable pre-adoption receipt as the adoption authority. Add a distinct binding-authorization receipt, created after adoption, for acceptance to read. This also changes `acceptance.py`, and possibly `preparation.py`.
- **Closing round.** Round 8 must confirm the fix and that the round-7 files are unchanged, or else re-review them.

## Suggestions (not issues)

1. **Stale procedure document** (both peers). `model-free-binding-amendment.md:66` still says to copy the file by hand and write the `adopted_at` receipt by hand. Following it is refused but recoverable (`:75`, `:263`, `:297`). Replace it with the exact `preparation.py adopt --amendment … --review …` invocation as part of 001's review. That document is in binding's reviewed set.
2. **Adoption freezes `preparation.py` in both bytes and path** (peer S-1, plus my own path point).
   - `writer = pin(__file__)` is re-checked on every load (`:89–90`, `:95`). Any later edit to `preparation.py` would make every command refuse, unless that edit migrates the pinned writer.
   - Adopt only with the final signed-off `preparation.py`. Never edit or delete `run-config-amendments-proposed/001-binding-v2.json` (pinned at `:313`).
   - `pin` uses `.absolute()`, not `resolve()` (`verify_inputs.py:26`). `__main__.__file__` is not resolved, but importers such as `review_identifiers.py:8` and `acceptance.py` use resolved paths. Invoke adopt and seal by the canonical absolute path, as the a/b argv already do.
3. **No guard against the wall clock going backward.** `Journal.__init__` does not refuse when its `started_at` is earlier than the latest `adopted_at`. If the clock stepped back after adoption, a post-adoption journal would permanently fail `:418`. Add a check; it is cheap.
4. **Test gaps.**
   - `test_load_and_seal_accept_legacy_ab_and_d_before_adoption` (`test_core.py:340–356`) never calls `collect_seal`. Rename it, or add an assertion at seal level.
   - Add tests for these refusals:
     - a hand-written backdated receipt/command pair that is internally consistent
     - "clock not later" (`:294`)
     - nonterminal (`:287`)
     - writer pin (`:89–90`)
   - Add an integration test of adoption → binding authority → seal. It would have caught R7-01, because the fixture reviews only the proposal (`test_core.py:177–187`).
   - Round-6 suggestion 1 still stands: `test_identifiers.py` stubs `load_config`.
5. **Source manifest timing** (round-6 suggestion 4, now confirmed). Adoption adds an EXEC file, so the source manifest must be produced after adoption (`:514–516`).

## Recommendation, alternatives, risks, execution

- **Recommendation.** Do not record zero issues yet. Close R7-01 with fix (a) in the binding track. Then hold a short round 8 confirming that the round-7 digests are unchanged and that the fix leaves `effective_configuration` and `verify_journal_configuration` alone. If it does, sign off d at those digests.
- **Alternatives.**
  - *Fix R7-01 in `preparation.py`* (for example, adopt validates only the proposal-bound part of the receipt and binding reads a later receipt). This changes the digest under review, so a full re-review is needed.
  - *Treat R7-01 as belonging to the binding track only and sign off d now.* Rejected under rule 1: it is a concrete failure on the exact path this round was asked about. It does, however, leave d's journals untouched.
- **Risks.**
  - Every finding comes from reading code [INFERENCE]; nothing was reproduced. Line numbers for `preparation.py` agree across the challenger, the peer and my own reading.
  - Once d has run, a future change to `verify_journal_configuration` could invalidate its journals, so any R7-01 fix must leave that function unchanged.
  - Adoption is irreversible: there is no path to re-adopt.
- **Execution outline.**
  1. Pick (a) or (b) for R7-01, then implement it with an end-to-end test of adopt → binding authority → `collect_seal`.
  2. Update `model-free-binding-amendment.md:66`.
  3. Main runs the execution suite and records digests.
  4. Round 8 with zero issues.
  5. Before running d, confirm that `run-config-amendments/` and `preparation-run/config-adoptions/` are empty and that all seven journal pairs are terminal and passed.
  6. Run d once at the canonical paths with the pinned CPython.
  7. Do the separate completeness/delta output review.

## Verification and peers

**What I read:**
- all of `preparation.py`
- `test_core.py:150–357`
- the `run()` call sites in `review_identifiers.py`
- `verify_inputs.pin` and `exclusive_json`
- `run-config.json` paths
- the proposed 001
- the timestamps of all seven journal pairs
- `acceptance.py:289–348`
- `model-free-binding-amendment.md:55–80`
- the round-6 record
- the binding round-2 verdict (4 issues open)

**innovation-challenger:**
- Found R7-01 and the stale procedure document. It rated the document an issue; I downgraded it to a suggestion because the failure is fail-closed and recoverable, with no immutable record written.
- Confirmed there is no deadlock, no interleaving and no string-based timestamp comparison.

**experiment-peer:**
- Its 11-hypothesis table rejected every way d's journals could fail.
- It reported **0 issues**, but it explicitly left the review receipt out of scope (its H10).
- Contributed suggestions on adoption freezing `preparation.py` and on the test name.

**Where the peers changed my conclusion:**
- My own analysis had reached the same R6-01 result and the path/clock suggestions.
- I had not traced the binding receipt until the challenger raised it. I then verified R7-01 directly at `acceptance.py:316–327` against `preparation.py:277–304`. That changed my verdict from zero issues to one.

## Verdict

**1 issue open (R7-01).** R6-01 is resolved: a d journal written now, before adoption, is accepted by `adopt` and by seal's journal checks. The adoption order, however, makes 001's binding acceptance and therefore seal impossible, so the projector must not be run yet. These digests stay under review:

- `execution/preparation.py` `599a128d64dbd0617e7110daddf77c3b1b8677e516e4f0c9b3cf6a1ea389e427`
- `execution/tests/test_core.py` `94afb7638acbb0316cc019077a87d924fda0e665b752d3d57acdfadbd9657ace`
- `execution/run-config.json` `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`
- `execution/run-config-amendments-proposed/001-binding-v2.json` `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121`
- `execution/review_identifiers.py` `adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2`
- `execution/identifiers/projection.py` `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072`
- `execution/tests/test_identifiers.py` `6b7f64b3d283180fae29c4bd54eb22f949e10bc9c85342f08f81d440053ea371`
