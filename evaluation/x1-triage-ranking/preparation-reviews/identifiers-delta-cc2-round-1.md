# X1 preparation (d) change-cost-v2 identifier delta — allowlist and implementation review — round 1

## Files reviewed (SHA-256)

These digests come from Main's packet (`local://x1-packet-digests.md`, from `shasum -a 256`). I have no shell, so I did not recompute them. Paths are relative to `evaluation/x1-triage-ranking/`.

| File | SHA-256 |
|---|---|
| `execution/identifiers/cc2_delta.py` | `f0cb9763d3fcb94dcd700c166b89a9c34149bd027ab4e264f665fe949d8a18b2` |
| `execution/identifiers/cc2_readers.py` | `1ad375a97b43a771245f3f746768dc72814656c1c81c8c4d8793a9c88644144b` |
| `execution/identifiers/cc2_schemas.py` | `db845451e7dbefc8f32ea7b50fbe86cdd5c702330f755e39cee345b2cfc5799f` |
| `execution/tests/test_cc2_delta.py` | `657fa365f5a7b75424a9b9d1f2e3eaab96d1d76101a634553f0791beb6f13faa` |
| `preparation-reviews/identifiers-delta-cc2-REQUEST.md` | `a018f153f1377a7f1c416bba0891a60675e118a34e037f7c2e237b7f151823d6` |
| Integration context: `execution/preparation.py` | `b1759ec710fa4b71dec1ef91800e84d285c733693b0cb5f4ffd7fe68f147c916` |
| Integration context: `execution/identifiers/projection.py` | `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072` |
| Integration context: `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` |
| Integration context: `execution/review_identifiers.py` | `adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2` |
| Integration context: `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |
| Signed source manifest: `preparation/exposure-sources.json` | `67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60` (source commit `7bb73e52…`) |

**Baseline that the producer binds.** Journal `preparation-run/journal/906389d7212a401b8c7f781d37960d84.json`, start record `c400e46e3f2ba85767165257b425a9719a223303b71cf78700d3111ae06f8738`. Its output pins:
- `manifest.json` `23a911ae…3c738d`
- `source-coverage.json` `24f93c47…682724`
- `delta-review.json` `240aa16d…6db04`
- `unresolved.json` `19d10781…dcd9`
- `familiarity.jsonl` `96751811…3e30`

The digest of the journal file itself is **not in the packet** (see R1-03).

## Basis

- X1 round 9 `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc`: preparation (a)–(e) only. The clock started 2026-11-09T03:22:58.701909Z.
- PREREGISTRATION `abd206c3ce3c64fd0176c4211e6c8613c9cde8aaaa2ad91e86b28306e9600fdb`.
- rev9-digests `2253911be62da66a8f5187a4ebe2c5e986c8b981bf619e872f7ebf65d0c9c2da`.
- exposure-recipe `c065347e…`, exposure-sources `67dc5a4b…`, frozen-exposure-inputs `ec92fba6…`.
- `preparation-reviews/identifiers-output-round-1.md`:
  - ID-01, the allowlist under review;
  - ID-02;
  - dispositions D-1 to D-5.
- Change-cost-v2 producer code, read as code only:
  - `collect.py`, `fetch_prs.py`, `common.py`, `build_panel.py`, `finalize_panel.py`;
  - `follow.py`, `jev_type.py`, `jev_label.py`, `validate.py`, `annotate.py`;
  - the aggregate writers.

## Prior disposition

- **ID-01** (major, output round 1). This round is the response. The allowlist and its implementation are **substantively correct** (see the checked items below). It is not yet authorised to run because R1-01 to R1-03 are open.
- **ID-02** (minor, output round 1). The code binds the correct baseline:
  - `cc2_delta.py:23` sets `BASELINE_JOURNAL = '906389d7…json'`;
  - `cc2_delta.py:157–165` requires action `d`, status `passed` and exit 0, and checks the five output pins in order;
  - those pins match journal `:187–212`.
  
  ID-02 itself closes at the output review.

## Checked and found correct

### Allowlist matches ID-01 exactly

| Schema (`cc2_schemas.py`) | Only selected (`ATOM`) values | Producer evidence |
|---|---|---|
| DATA `:9–14` | `repo`, `number`, `S`, `trail`, `files[].p`, `partners[].g/via/vias[]/via_ref` | `collect.py:140–143` (base); the early returns `:145–177`; `S` `:154–170`; `trail` `:179,191`; `files` `:187–190`; partners `:237`; `:242` |
| RAW `:15–19` | `number`, `base_oid`, `head_oid`, `merge_commit`, `merge_parents[]`, `commits[].oid/parent`, `files[].path` | `fetch_prs.py:86–107` (every top-level and child key in `slim`) |
| CANDIDATE | `repo`, `parent`, `follow` (PR numbers) | `follow.py:69–76` |
| TYPE_CALL / PAIR_CALL | `repo`, `number` / `repo`, `parent`, `follow` | `jev_type.py:46–61`; `jev_label.py:47–64`. The success, HTTP and exception key sets are all covered. |
| Samples / blinded tables | `repo`, PR numbers, `files[]`, `earlier/later.files[]`, `file_a/b` | `validate.py:102–105,119–123,146–158` (blinded files drop `hidden_keys`) |
| Panels | `repo`, `number` cells only | Header from `build_panel.py:124–150` plus `area`; `finalize_panel.py:62–63` adds the two outcome columns |

- Every value the code emits is one of ID-01's kinds:
  - PR number;
  - `S`, `trail`, `base_oid`, `head_oid`, `merge_commit`, `merge_parents[]` and `commits[].oid/parent`, all as **commit**;
  - `files[].path/p` and the partner path keys.
  
  `repo` is decoded only for scope.
- No key the producers write is missing from the schemas, and no identifier-bearing key is marked opaque:
  - `shared` is a subset of the raw paths plus the partner paths (`follow.py:69–76`);
  - `main_path` is one of `files.path` (`build_panel.py:121–122`);
  - `head_a`/`head_b` are source excerpts.
- Every value in the 12 schema dictionaries was checked by inspection. No forbidden key (score, label, stratum, outcome, Jev response or probability, text, timestamp, author) is `ATOM`.

### Forbidden values are never decoded

- Values under `SKIP` are only scanned for structure:
  - `string_token(False)`;
  - a regex framing check on numbers;
  - the literals `null`/`true`/`false`.
  
  The keys inside skipped containers are not decoded either (`cc2_readers.py:55–83,122–139`).
- Keys in projected objects are decoded and checked against the schema before their value is read. Unknown and duplicate keys refuse (`:89–95`).
- The CSV reader decodes only the header and the `repo`/`number` cells (`:199–216`).
- JSONL line splitting is on bytes, and JSON strings cannot contain raw CR or LF.

### Pins, scope and merge semantics

- Every producer pin and every table pin is checked before any table is decoded, and the same bytes are then parsed (`cc2_delta.py:246–262`).
- Symlinks and `.codegraph` paths are refused (`:148`).
- Scope:
  - the 16 aliases equal `common.py` `REPO_CFG`;
  - only `django` and `prometheus` rows emit;
  - rows from other repositories are counted;
  - an inconsistent fixed scope refuses.
- The rule names match the baseline (`projection.py:65–78`):
  - commits use `commit_or_parent_either_image_any_scope` and apply to all three tests;
  - PRs and paths are repository-scoped;
  - blinded paths use `unresolved_repository_scope_unknown_freshness_ineligible` and apply to all three tests.
- The manifest declares union semantics and computes no exclusions (`:179–183`).

### Create-once outputs

- `root.mkdir()` is exclusive (`:177`).
- `exclusive_json` and `open('x')` both create with `x` mode and fsync (`verify_inputs.py:34–47`).
- The baseline outputs live in another directory and are only read.

### Coverage of other change-cost-v2 tables

The remaining tables add no case identifiers:
- `annotations-*` and `run-1-failed/`: id, label and model/prompt provenance only (`annotate.py:78,110–116`);
- aggregate writers: `constructs`, `gates`, `analysis`, `power`, `sensitivity`, `harness`;
- `tests.py`: temporary fixtures only.

v4 parsed `pins.json` (`source-coverage.json`: `parsed_identifier_projection`), and its `sha` key is a v4 commit key (`extract_exposures.py:29`).

### Integration with `preparation.py` `b1759ec7`

- The delta outputs are accounted at seal as `journal-output` (`:444–451,564,572`).
- Canonical roles are untouched.
- The actual seal blockers are elsewhere:
  - the execution-source closure is stale because of the new `EXEC` files (`:531–533`);
  - any refused journal blocks the seal (`:555–557`).
- Both blockers belong to the separately reviewed seal integration, as the request says.

### Transport-cache disposition (CB-1: `raw/q`) — accepted

- **Conservation from the code.**
  - The cache key is a hash of the query and its variables (`fetch_prs.py:36–40`).
  - Every page the final run read is appended to `out` (`:79–80`).
  - `slim` copies every identifier the query requests: `baseRefOid`/`headRefOid` (`:93`), the merge commit and its parents (`:95`), each commit's oid and first parent (`:97–100`), and file paths (`:103`).
  - The cache and the export have the same `first:100` caps.
  - `seen` drops only duplicate PR numbers (`:130`). The date ranges are disjoint (`:69–83,110–129`), and a merged PR's identifiers do not change.
  - So the cache pages the final run read add no identifier beyond `raw/prs-*`.
- **Already extracted.** v4 parsed every `raw/q` file (`source-coverage.json:8655–8676`, `parsed_identifier_projection`):
  - its paths are unscoped, so they are ineligible in all three tests under D-2;
  - its `oid` values are retained as any-scope SHAs.
- **Remaining risk.** Only cache entries the final run never read (orphan pages). There is no evidence of them:
  - change-cost-v2 never re-pinned (`PREREGISTRATION.md:211–212`), and its reviews mention no fetch changes;
  - the peer's size tally of the metadata, 144 objects of 53–55 bytes = 16 repositories × 9 chunks, fits a single window configuration [INFERENCE].
- **Decision.** No GraphQL-cache allowlist is required.

## New issues

| ID | Severity | Finding | Concrete failure | Required fix |
|---|---|---|---|---|
| R1-01 | minor | **Some executed code is bound by neither the review gate nor the journal.** `source_pins`/`control_pins` (`cc2_delta.py:226–236`) cover only the three modules, the test and the request. The producer imports `TESTS` from `identifiers/projection.py` (`:18`), and `TESTS` decides `applies_to_test_repositories` for every record (`:64–78`). Gate and journal code come from `preparation.py`/`verify_inputs.py`. The signed baseline producer does pin `projection.py` in its journal (`review_identifiers.py:261–262`). | `projection.py` or `preparation.py` (recently changed) is edited after sign-off. The receipt still validates and the run proceeds with unreviewed `TESTS`, gate or journal code. Neither the journal nor the receipt can show which bytes produced the records. | Add `pin(EXEC/'identifiers/projection.py')`, `pin(EXEC/'preparation.py')` and `pin(EXEC/'verify_inputs.py')` to `implementation` (journal `source_pins`) and so to `control_pins`. List them in the request's binding sentence. |
| R1-02 | minor | **A refused run pins outputs it did not create.** The `except` block pins every existing `manifest/coverage/familiarity` file under `root` (`cc2_delta.py:270–271`), even when `root.mkdir()` (`:177`) failed because an earlier run created them. The test asserts this: the refused rerun has 3 `output_pins` (`test_cc2_delta.py:308–316`). | An accidental second invocation writes a refused action-d journal that claims the first run's three outputs as its own. The provenance of the reviewed outputs is now false. | Record each name in `publish` immediately after its exclusive create, including a partly written `familiarity.jsonl`, and pin only those files. This follows the baseline's `written` convention (`review_identifiers.py`). Change the test to expect 0 pins on the rerun and to keep pins for partial files from the current invocation. |
| R1-03 | minor | **The binding set cannot be signed from this packet.** `review_gate` requires the receipt to bind `prior` (`cc2_delta.py:235–236`), which includes the baseline journal's own pin (`:165`). `review_receipt` requires the council record to name every reviewed SHA-256 (`preparation.py:394–406`). Neither the request nor the packet gives the SHA-256 of `906389d7…json`. The request also defers that digest to "the next output review" (`REQUEST:55`), which contradicts its own receipt requirement (`:53`). | A receipt built as the code requires cannot be authentic against this round. `review_receipt` raises after the journal is reserved (`:231` then `:236`). That leaves a refused action-d journal, which the seal rejects (`preparation.py:555–557`). | Main supplies the SHA-256 and byte count of `preparation-run/journal/906389d7212a401b8c7f781d37960d84.json`. Put it in the request's baseline binding and delete the deferral sentence. The next round checks it against the journal (`passed`, the five output pins, start record `c400e46e…`) and names it in the sign-off. |

## Suggestions (not blocking)

1. **Correct the request's seal sentence** (`REQUEST:69`). Under `b1759ec7` the delta outputs *are* accounted, as `journal-output`. What is still missing is reviewed integration:
   - `required_reviewed` is computed before journal outputs are added (`preparation.py:544–546`);
   - in that integration, add the delta kinds to `required_records`/`required_files` and the delta output-review scope to `required_reviews`, so that sealing enforces review of the delta.
2. **Tests (code inspected correct; frozen after sign-off):**
   - put a sentinel under every excluded key for each schema, including numbers, booleans, nulls and escaped strings, and assert the exact sequence of `json.loads` calls;
   - give the last `TABLES` row a bad pin and assert zero projection calls;
   - add a refusal test for a changed baseline output;
   - assert the arguments to `review_gate` instead of mocking it with `{}`.
3. **Pre-flight before reserving the journal:** check that `root` does not exist and validate the receipt. Every refused journal blocks the seal.
4. **Possible older rows in append-mode Jev outputs.** `types.jsonl`/`labels.jsonl` are appended to (`jev_type.py`/`jev_label.py` `open('a')`), and change-cost-v2 round-1 mentions "pre-existing records". An older key, or `NaN`/`Infinity`, refuses closed: there is no output, but there is a seal-blocking refused journal. Plan the recovery path.
5. **Report the over-exclusion.** The blinded tables apply other repositories' paths to all three tests. Their paths are a subset of the sample paths (`validate.py:152–158`). Report the count of these over-exclusions in `coverage.json` and the output review.
6. **Add `study` to the records.** Put `study: change-cost-v2` explicitly on each record or source row, so D-1 citations do not have to derive it from `source_relative_path`.
7. **At the output review:**
   - recount the `raw/q` objects by size against Σ max(1, ⌈issueCount/20⌉);
   - confirm the Django and Prometheus `pins.json` SHAs are in the v4 manifest;
   - adopt suggestion 1 from output round 1 (match SHAs regardless of kind) for v4's `raw/q` `oid` values.
8. **Fix the request's wording on headers:** the CSV header is checked as a set, not in order.

## Peers and where they changed the recommendation

- **experiment-peer:**
  - **F2 changed my verdict.** I had accepted the journal binding. It became R1-01.
  - **F4** became R1-02. I had treated refusal pinning as following the convention; the peer showed the misattribution.
  - **F1 and F3** are kept as suggestions 1 and 2. The integration was deferred explicitly, and the code was inspected correct.
  - **Suggestion 1** (the metadata tally) strengthened CB-1.
- **innovation-challenger:**
  - **CC2-C1** (major: cache conservation not established) was rejected as an issue:
    - its duplicate-node scenario needs the identifiers of a merged PR to change between cached pages, and they do not;
    - pagination caps are the same in cache and export;
    - the orphan risk is unevidenced and already partly covered by v4.
    
    It led to the explicit CB-1 disposition and suggestion 7.
  - **CC2-C2** was merged into suggestion 2 and **CC2-C3** into suggestion 1.
- R1-03 is my own finding.

## Verification

- This was a static, read-only review of the producer code, the delta code, `preparation.py`, the X1 control documents (baseline journal, `source-coverage.json`) and the metadata in `exposure-sources.json`.
- No change-cost-v2 data payloads, `pins.json` or `raw/q` contents were opened.
- Nothing was executed, and there were no model calls.
- Main reports 237 Python tests OK and `go vet`/`go test` ok at these bytes. I did not re-run them, and I recomputed no digests.
- Whether the checkout's change-cost-v2 bytes match the pins is enforced at runtime (`cc2_delta.py:146–154`). The peer found that the pinned `fetch_prs.py` digest equals change-cost-v2's round-14 sign-off.

**What Main should run after the fixes:**
- `test_cc2_delta.py` and the full suite under the pinned CPython 3.11.11;
- the `shasum -a 256` digest and byte count of the baseline journal file, for R1-03.

## Verdict

**3 issues open** (R1-01, R1-02 and R1-03, all minor). The ID-01 allowlist and its code-derived schema registry are correct, and the `raw/q` transport-cache boundary (CB-1) is accepted. The delta producer must not be run until a later round reports zero issues. That round's sign-off must name every bound digest:
- the three modules, the test and the request;
- `projection.py`, `preparation.py` and `verify_inputs.py` (after R1-01);
- `exposure-sources.json` and `run-config.json`;
- the baseline journal and its five outputs.

No sign-off is given in this round.
