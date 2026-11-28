# X1 preparation (d) identifier projector — implementation review — round 6 (integration)

**Scope.** This round covers only the R5-01 fix. That means the `load_config` adoption gate, `effective_configuration`, the reload check in `Journal.__init__`, and how `verify_journal_configuration` and `collect_seal` judge d's journals. Identifier completeness and the delta are out of scope. So is the merit of amendment 001. Nothing was edited or run: not the projector, not seal, not the tests. The only model calls were the two council peers.

## Reviewed files

Main computed these SHA-256 values. Neither this reviewer nor either peer could recompute them, because no session had a shell. Paths are under `evaluation/x1-triage-ranking/`.

| File | SHA-256 | Change since round 5 |
|---|---|---|
| `execution/preparation.py` | `e883abc47ea9ea6a693a78ffe46e5793bb1a6b1ccdc61cb17bf2c815ab99db6e` | changed (was `745823c7…`) |
| `execution/tests/test_core.py` | `2fa57759d64e2ed19a533358b79dc011375fdef82f7908a6728daad968314e41` | changed (was `0feb6447…`) |
| `execution/run-config-amendments-proposed/001-binding-v2.json` | `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121` | moved, same bytes |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | unchanged |
| `execution/review_identifiers.py` | `adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2` | unchanged |
| `execution/identifiers/projection.py` | `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072` | unchanged |
| `execution/tests/test_identifiers.py` | `6b7f64b3d283180fae29c4bd54eb22f949e10bc9c85342f08f81d440053ea371` | unchanged |

**Correction to the brief.** `execution/run-config-amendments/` does exist; it is an empty directory. It is not absent. This does not change behaviour: `Path.glob('*.json')` yields nothing in either case (`preparation.py:56`). `adoption_state.directory_contents` is built from the amendments that were loaded, not from a directory listing (`:114`, `:298–299`). Seal's source-closure `rglob` only counts files (`:387`).

## R5-01: resolved

- **The proposal folder is never read.** `load_config` only globs `path.parent/'run-config-amendments'` (`preparation.py:56`). `run-config-amendments-proposed/` is referenced only by the test that checks it is ignored (`test_core.py:240–249`).
- **Adopted amendments are checked when loaded.** Each one now needs a scoped zero-issue receipt that binds its digest (`:64–71`, `review_receipt` `:269–281`), plus a non-future UTC `adopted_at` in chain order (`:72–76`).
  - A stray or unreviewed file therefore stops d before anything is written. `run()` calls `load_config` at `review_identifiers.py:251`, before `Journal` at `:259`.
  - `Journal.__init__` reloads the config and compares it before the create-once write (`preparation.py:204–205`). A refusal at either point leaves no record, so d can be rerun.
- **What d's journals would record today.** Both the start and terminal records would hold the following (`:110–114`, `:211`):
  - `amendments: []`
  - `adoption_state` = `{condition: 'no-adopted-amendments', directory_contents: []}`
  - `sha256` = `config_digest` of the base config, with `value` holding the base config itself

  `source_pins` covers only `review_identifiers.py` and `projection.py` (`review_identifiers.py:260`). Nothing pins 001's proposal file.
- **What seal now checks for d.** Seal no longer requires a d journal's chain to be a prefix of the chain present at seal. It compares the journal against amendments with `adopted_at ≤ started_at` (`:287`). It derives the expected digest from the base file when that set is empty (`:295`). With honest adoption timestamps, editing, withdrawing or renaming 001, or adding a later 002, can no longer touch d's journal. That was the R5-01 failure, and it is closed.
- **Supporting checks (all passed):**
  - `run-config.json` contains no `{path, sha256}` object (it has only prose mentions at `:133–143`). So `verify_tree_pins` (`:418`) re-checks nothing new inside `effective_configuration.value`.
  - The base digest matches after a JSON round trip: there are no floats, and `RunConfig` only adds an attribute that is not a dict key (`:31–33`).
  - Legacy a/b journals still pass `:289–292` and the start/terminal `.get` comparison (`:415`), provided no adoption predates them. All seven pairs lack `effective_configuration` (checked by grep); their starts are 2026-11-09T03:23:04.44Z to 03:35:10.39Z.

## New issues

### R6-01 (issue): a hand-written `adopted_at` can permanently fail d's journals, and the a/b journals, at seal

**Mechanism.**
- The seal check is now based on time. A journal is expected to contain every amendment with `adopted_at ≤ record['started_at']` (`preparation.py:287`).
- `adopted_at` is typed into a receipt by the operator. It is checked only for being UTC, not in the future (`:73`), and in chain order (`:75–76`). Nothing ties it to when the file is copied into `run-config-amendments/`. Nothing requires it to be later than journals that already exist.
- The adoption procedure only says Main copies the reviewed bytes and then creates "the authority receipt with its nonfuture UTC `adopted_at`" (`preparation/model-free-binding-amendment.md:66`).
- No code writes this receipt (`adopted_at` appears only in `preparation.py` and `test_core.py`).

The old prefix rule could never reject an empty-chain journal. The new partition can.

**How it triggers (any one of these):**
1. 001 is signed off at time T_s and d runs at T_d > T_s while 001 is still only in `-proposed`. Main then writes `adopted_at = T_s`, the natural "adopted at sign-off" reading. Seal expects `[001]` in d's journal, finds `[]`, and refuses with "journal amendment chain/adoption ordering differs" (`:293–294`).
2. Review records carry only a date, so the receipt may use a date-only stamp such as `2026-11-09T00:00:00Z`.
   - The test helper uses exactly that shape as its default (`test_core.py:175`).
   - That stamp also predates all seven a/b journals, which fail with "post-adoption journal lacks effective config" (`:290–291`). `test_core.py:265–268` confirms both refusal paths.
3. A narrow race (challenger):
   - d holds only the shared lock (`review_identifiers.py:258`), and adoption does not take the lock.
   - The reload at `:204` runs before `started_at` is captured at `:207`.
   - If 001 is adopted with a stamp of the current time between those two lines, the journal records `[]` but its `started_at` is later than `adopted_at`.

**Why this is permanent:**
- Nothing detects the mismatch before seal.
- Seal reserves `preparation.seal.json` before it validates, and the reservation stays after a refusal (`:445–460`, `:189–196`).
- Once any journal written after adoption holds 001's `adoption_review` pin and `adopted_at` (`:103–105`), correcting the receipt breaks that journal too (`:121–128`, `:418`).
- The d and a/b journals are create-once.

**Required fix (either one; record which).**
- **(a) Code, preferred.** Before any new immutable record is written, `load_config` runs the `verify_journal_configuration` logic against every existing journal start record. It refuses any adopted amendment whose `adopted_at` is at or before the `started_at` of an existing journal that does not record it.
  - This turns the seal-time dead end into a refusal at load. That refusal can be recovered while no journal yet pins the receipt.
  - It covers a, b and d.
  - It changes `preparation.py`, so it needs its own review round.
- **(b) Procedure, no change to d's bytes.** Add a binding precondition to 001's adoption: copy the file first, then write the receipt with `adopted_at = now()` while holding the exclusive `operation_lock`.
  - `adopted_at` must be strictly later than every existing journal's `started_at`, including d's terminal `ended_at`.
  - The round-6 response and `model-free-binding-amendment.md` must both state this, and 001's adoption review must verify it.
  - The cost is that it relies on operator discipline.

## Suggestions (not issues)

1. **Test the real config path in the projector tests.** `test_identifiers.py` still stubs `load_config` with a plain dict (l.96–100), so the `RunConfig` reload branch (`preparation.py:204`) never runs for d. Add one `run()` test against the real `load_config` with an empty adopted directory and a populated `-proposed`.
2. **Close the lifecycle test gaps.** Add tests for: loading with the directory absent, a journal validated after a 002, the timestamp boundary (`<=`), and an actual rename. `test_edit_delete_or_rename_adopted_amendment_refused` covers only edit and delete (`test_core.py:270–279`).
3. **Make `directory_contents` honest.** Rename it, or document it as "loaded amendment basenames". The d run record should include a real listing of both amendment directories.
4. **Note for 001's review [INFERENCE].** Seal requires the source manifest to match the current `EXEC` files exactly (`:385–389`). Adding the adopted copy after sign-off may make a manifest reviewed before adoption stale.
5. **Existing risk, not new.** Legacy journal a pins `/opt/homebrew/Cellar/git/2.56.0_2/bin/git`, which seal re-checks (`:418`). Don't upgrade or clean up Homebrew git before seal.

## Recommendation, alternatives, risks, execution

- **Recommendation.** Do not run d yet. Close R6-01 with fix (a), then hold round 7 on the new `preparation.py` digest. If the d run cannot wait for a code change, (b) is acceptable as long as it is written into the amendment document, cited from the round-6 response, and confirmed in round 7.
- **Alternatives.**
  - *Go back to the round-5 prefix rule.* With load-time review gating, an empty-chain journal is always a prefix. But this drops the check that no journal written after adoption omits the amendment.
  - *Record an adoption boundary tied to journal reservations (challenger).* This is sounder, but a larger redesign.
  - *Run d now and rely on 001's adoption review to choose a safe timestamp.* Rejected: the d journal is irreversible, and nothing currently requires a safe timestamp.
- **Risks.**
  - All findings come from reading the code [INFERENCE]; nothing was reproduced.
  - The digests were not recomputed by the council.
  - With (b), the guarantee is only as good as the operator.
  - The R6-01 trigger happens at adoption, outside this round's files.
- **Execution outline.**
  1. Implement (a): an existing-journal consistency check in `load_config`, plus tests for a backdated `adopted_at` (refused at load), a later `adopted_at` (accepted), and a/b legacy journals with a date-only stamp (refused at load).
  2. Main runs `test_core.py` and `test_identifiers.py` and records the new digests.
  3. Hold round 7 with zero issues.
  4. Record the precondition that `run-config-amendments/` holds no `*.json`, and include directory listings.
  5. Run d once at the canonical paths with the pinned CPython.
  6. Send the retained outputs to the separate completeness/delta output review.

## Verification and peers

- **Read:**
  - `preparation.py`: all 520 lines.
  - `test_core.py`: lines 149–300.
  - The round-5 record.
  - `review_identifiers.py`: imports, plus lines 250–262.
  - `run-config.json`: grep.
  - `model-free-binding-amendment.md`: lines 62–70.
- **Run state:**
  - `preparation-run/` has a clock, the input receipt, seven journal pairs, and object records. It has no `identifiers/` or `reviews/` directory.
  - The a/b start timestamps were confirmed by grep.
  - Main reports `test_identifiers.py` OK and `test_core.py` OK with 25 tests. These were not reproduced, and they do not exercise R6-01.
- **innovation-challenger:**
  - Found R6-01 independently, contributed the lock and reload race, and proposed an adoption boundary tied to reservations.
  - Confirmed that an empty or absent directory is harmless and that pins carry no mtime.
- **experiment-peer:**
  - Found R6-01 independently, including the date-only stamp that also breaks a/b, the failure being permanent once a post-adoption journal exists, and both fix options.
  - Its hypothesis table ruled out the other residual-failure paths.
  - It contributed suggestions 4 and 5.
- **Own analysis:** found the backdating trigger before either peer reported. The challenger's race and the peer's a/b extension widened the issue. Neither changed the verdict.

## Verdict

**1 issue open (R6-01).** R5-01 is resolved, but the projector must not be run yet. These digests stay under review:

- `execution/preparation.py` `e883abc47ea9ea6a693a78ffe46e5793bb1a6b1ccdc61cb17bf2c815ab99db6e`
- `execution/tests/test_core.py` `2fa57759d64e2ed19a533358b79dc011375fdef82f7908a6728daad968314e41`
- `execution/run-config-amendments-proposed/001-binding-v2.json` `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121`
- `execution/run-config.json` `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`
- `execution/review_identifiers.py` `adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2`
- `execution/identifiers/projection.py` `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072`
- `execution/tests/test_identifiers.py` `6b7f64b3d283180fae29c4bd54eb22f949e10bc9c85342f08f81d440053ea371`
