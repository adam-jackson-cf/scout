# X1 preparation (d) identifier projector — implementation review — round 9 (dependency)

**Question.** The identifier projector already ran once, under the signed-off `preparation.py` `599a128d…`. Its journal is `906389d7212a401b8c7f781d37960d84` and its outputs are under `preparation-run/identifiers/`. Core is now `b1759ec7…`. Can that journal and its outputs still pass seal accounting under the new Core?

**Out of scope:**
- identifier completeness and the change-cost-v2 delta (the separate output review);
- binding-v2;
- whether Core `b1759ec7` is acceptable on its own merits.

**Method.** Read-only. I did not run the projector, seal or the tests. I made no model calls apart from the two council peers. I did not open the contents of the identifier output files; I read only journal metadata, the code and directory listings. Main computed the digests in `local://x1-packet-digests.md`; this council did not recompute them.

## Files and SHA-256

Paths are under `evaluation/x1-triage-ranking/`.

| File | SHA-256 | Since round 8 |
|---|---|---|
| `execution/preparation.py` | `b1759ec710fa4b71dec1ef91800e84d285c733693b0cb5f4ffd7fe68f147c916` | changed (was `599a128d…`) |
| `execution/tests/test_core.py` | `4f4bfdaf50fc0d56ff775f8c86bb6bc579c932acc1a76764a55ba4a6f36da07e` | changed (was `94afb763…`) |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` | used for `pin` |
| `execution/review_identifiers.py` | `adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2` | unchanged |
| `execution/identifiers/projection.py` | `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072` | unchanged |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` | unchanged |
| `execution/identifiers/cc2_delta.py` | `f0cb9763d3fcb94dcd700c166b89a9c34149bd027ab4e264f665fe949d8a18b2` | read for blast radius only |
| `preparation-reviews/binding-v2-round-3-response.md` | `44a391c3b8d7a1aff2c0743fca2af2e642356e675a7bb9e15caa3b41d2cfea0a` | describes the Core change |

The existing action-(d) artifacts below are values recorded in the journal. I did not recompute them.

| Artifact | Bytes | SHA-256 |
|---|---|---|
| journal start record `906389d7….start.json` | 8524 | `c400e46e3f2ba85767165257b425a9719a223303b71cf78700d3111ae06f8738` |
| `identifiers/manifest.json` | 27924749 | `23a911ae22084db67a5069a49c6db8ef4b0ec7a555c71348446d3372213c738d` |
| `identifiers/source-coverage.json` | 30151508 | `24f93c47ab0ab356c7bdafe7bda906c55508370deefbe75e6fad2c13fc682724` |
| `identifiers/delta-review.json` | 4635 | `240aa16d942f73e37d1c77b5055aad4a3158f08efcc63555833461e21276db04` |
| `identifiers/unresolved.json` | 91278 | `19d10781ceb31b7f4ae58d1a786de24a8af6b09fde925bafe134a398f395dcd9` |
| `identifiers/familiarity.jsonl` | 20588100 | `967518114b8187e8fcc4fb92615892d1cdab3a50e3cff6ab8fb10f3eb75e3f30` |

Main's verification at these bytes: 237 Python tests OK; `go vet` clean; `go test -count=1` ok. I did not rerun them.

## Basis

The signed basis has not changed:
- PREREGISTRATION `abd206c3…`
- rev9-digests `2253911b…`
- round-9 authority `6f1d6ca2…`
- pins.json `6d556fb5…`
- model-free-policy `87de3949…`
- binding-check v1 `5056ae74…`
- runtime-attestation-rev6 `19463957…`
- exposure-recipe `c065347e…`, exposure-sources `67dc5a4b…`, frozen-exposure-inputs `ec92fba6…`

The 45-day clock is still the one started at `2026-11-09T03:22:58.701909Z`.

## Prior disposition

Round 8 (`identifiers-impl-round-8.md`) recorded zero issues. It authorised one run at canonical paths with Core `599a128d…`. Its risk note said that changing `verify_journal_configuration` or the `Journal` record shape after d runs would invalidate d's journal. That condition is what this round checks.

## Trace: existing d journal → seal under `b1759ec7`

1. **The change is limited to seal accounting.**
   - New `verify_output_accounting` (`preparation.py:444–458`):
     - re-verifies every journal output pin (`:447`);
     - adds a path as `journal-output` only if it is not already referenced (`:449–451`);
     - otherwise refuses any unreferenced file under `OUT` (`:453–457`).
   - `collect_seal` now collects `output_pins` from every terminal journal (`:548`, `:564`) and calls the accounting at `:572`.
   - Round 8 cited `599a128d` line numbers. Every one up to `:443` still lands on the same code at the same line: `:140–149`, `:160–174`, `:277–318`, `:332–339`, `:394–405`, `:411–412`, `:418–426`.
   - After the insertion, every cited line has moved by exactly +17 or +18. That matches the inserted 17-line function and the line added at `:548`.
   - This is consistent with `Journal`, `effective_configuration` and `verify_journal_configuration` not having changed. [INFERENCE: line-number evidence, not a byte diff.]
2. **d's outputs are all canonical before accounting runs.**
   - The four `identifiers/*.json` files are `required_records` (`run-config.json:94–109`). They are pinned with their kind as role at `:487–494`.
   - `familiarity.jsonl` is in `required_files` and is pinned as `identifier-output` at `:527–528`.
   - Both happen before `:572`, which passes `sources + outputs` as the canonical set.
   - The journal pins use the same absolute path strings as `pin()` produces from `paths.OUT` (`verify_inputs.py:25–28`). For example: `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/manifest.json`.
   - So, for d, `verify_output_accounting` returns `[]`. It adds no rows and changes no roles.
3. **The identifier evidence keeps its review requirement.**
   - `required_reviewed` (`:544`) is computed before the accounting call. The canonical roles `scout.x1.identifier-manifest.v1`, `…source-coverage.v1`, `…delta-review.v1`, `…unresolved.v1` and `identifier-output` are kept. This is the fix from 7afeaa0d to b1759ec7 (`binding-v2-round-3-response.md:62–66`).
   - The new tests cover it: `test_canonical_output_role_preserved_when_journal_also_pins_it` and `test_duplicate_journal_outputs_are_deduplicated_but_all_verified` (`test_core.py:89–109`).
4. **There is no new way for d to fail.**
   - Re-verifying at `:447` repeats a check `599a128d` already made: `verify_tree_pins(value)` over the whole terminal record, now at `:563`.
   - It refuses only if the output bytes changed, which the old Core refused too.
   - The d journal and start record are counted as `journal` and `journal-start` (`:566`).
   - `preparation-run/identifiers/` holds exactly the five pinned files, so there is no stray output.
   - The reserved paths (clock, config, pending seal start) do not involve d.
5. **The Core digest change does not invalidate d's pins.**
   - d's `source_pins` are only `review_identifiers.py` `adc42d38…` and `projection.py` `2408ed45…`. The configuration pin is `c817f678…` and the clock pin is `26a1212a…`.
   - No pin in the journal, clock or input receipt names `preparation.py` bytes (peer grep).
   - The effective configuration has `amendments: []` and `no-adopted-amendments`. That still satisfies `verify_journal_configuration` (`:409–426`) before adoption, and after it, as round 8 found.
6. **Blast radius: the later cc2 delta.**
   - `cc2_delta.baseline` (`cc2_delta.py:157–165`) requires the d journal's `output_pins` to equal fresh pins exactly. It does not rewrite the baseline files.
   - The delta publishes to `OUT/identifiers-delta-cc2` (`:219`).
   - That leaves d's sealability unchanged.

## New issues

None.

## Suggestions (not issues)

1. **Diff Core against `599a128d`, not against `7afeaa0d`.** Main should run a byte diff from `599a128d` to `b1759ec7` on `preparation.py`. It should touch only `:444–458`, `:548`, `:564` and `:569–572`. That closes the remaining line-count blind spot in point 1 of the trace (peer H6). Core `b1759ec7` still needs its own Core/closure round; `closure-impl-round-4.md` asks for one whenever Core is not `599a128d`.
2. **Re-pin d before seal.** Re-pin the five outputs and the start record and compare them with the journal values above. This is the same check `cc2_delta.baseline` makes.
3. **Add a test that seals a d journal shaped like the real one.**
   - Today's d-journal fixtures have no `output_pins` (`test_binding_seal.py:95–98`, `test_core.py:397–402`).
   - Add one synthetic `collect_seal` case whose d `output_pins` duplicate the canonical identifier paths. It should assert that roles are unchanged and that no `journal-output` rows appear.
4. **For the cc2 review track, not d:**
   - cc2's outputs under `identifiers-delta-cc2/` are accounted only as `journal-output`. That role is outside `required_reviewed` (`:544`), so no scoped review binds them at seal unless config or a review adds them.
   - A refused cc2 terminal record blocks seal (`:555`).
5. **Adoption fixes Core's bytes.** Adoption pins `writer=pin(__file__)` (`:313`). Core must be final before `adopt`.

## Verification and peers

**What I read:**
- `preparation.py` `:322–583` and the grep anchors;
- `verify_inputs.py:25–37`;
- `test_core.py:70–130`;
- `run-config.json` paths, records and files;
- the d journal metadata;
- the `preparation-run/` listing, with hidden files;
- `cc2_delta.py` grep;
- `identifiers-impl-round-8.md`;
- `identifiers-output-round-1.md`;
- the Core change section of `binding-v2-round-3-response.md`.

**innovation-challenger:** 0 issues. It ruled out path mismatch, role loss, duplicate rows or bytes mismatches, journal shape, stray files, and the cc2 interaction. It suggested a combined synthetic fixture, which I merged into suggestion 3.

**experiment-peer:** 0 issues. It ruled out hypotheses H1–H8, and its line-offset arithmetic matches mine. Its suggestions became 1–5 here.

**Where the peers changed my recommendation:** they did not change the verdict. The peer added suggestions 1, 4 and 5. Its point that Core `b1759ec7` needs its own Core round is recorded as suggestion 1. It is not a d issue, because d's journal pins no Core bytes.

## Verdict

**zero issues — existing action-(d) journal and outputs remain sealable under preparation.py b1759ec7…**

Signed off at these reviewed digests:
- `execution/preparation.py` `b1759ec710fa4b71dec1ef91800e84d285c733693b0cb5f4ffd7fe68f147c916` (signed off only for this dependency question)
- `execution/tests/test_core.py` `4f4bfdaf50fc0d56ff775f8c86bb6bc579c932acc1a76764a55ba4a6f36da07e`
- `execution/verify_inputs.py` `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`
- `execution/review_identifiers.py` `adc42d38ae3bd9871abedf424f9b28852ecf1863c17fcc9a0343ec71bbd4f4e2`
- `execution/identifiers/projection.py` `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072`
- `execution/run-config.json` `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54`
- `execution/identifiers/cc2_delta.py` `f0cb9763d3fcb94dcd700c166b89a9c34149bd027ab4e264f665fe949d8a18b2` (blast radius only)
- `preparation-reviews/binding-v2-round-3-response.md` `44a391c3b8d7a1aff2c0743fca2af2e642356e675a7bb9e15caa3b41d2cfea0a`
- d start record `c400e46e3f2ba85767165257b425a9719a223303b71cf78700d3111ae06f8738`
- d outputs, as recorded in the journal:
  - `manifest.json` `23a911ae22084db67a5069a49c6db8ef4b0ec7a555c71348446d3372213c738d`
  - `source-coverage.json` `24f93c47ab0ab356c7bdafe7bda906c55508370deefbe75e6fad2c13fc682724`
  - `delta-review.json` `240aa16d942f73e37d1c77b5055aad4a3158f08efcc63555833461e21276db04`
  - `unresolved.json` `19d10781ceb31b7f4ae58d1a786de24a8af6b09fde925bafe134a398f395dcd9`
  - `familiarity.jsonl` `967518114b8187e8fcc4fb92615892d1cdab3a50e3cff6ab8fb10f3eb75e3f30`
