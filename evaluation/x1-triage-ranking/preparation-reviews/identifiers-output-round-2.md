# X1 preparation (d) identifier output review — round 2

**Scope.** This round reviews two of the three items round 1 required (`identifiers-output-round-1.md`, SHA-256 `9c6cdb66c4a6d2311848654275da839149c51f660e041418bde92bb34e0909d4`):
- the appended outputs of the change-cost-v2 (CC2) identifier delta producer;
- the corrected binding of the baseline action-(d) journal.

The third item, the per-schema request and its producer, was signed off in `identifiers-delta-cc2-round-5.md`. This round decides whether ID-01 and ID-02 close.

The review was static and read-only. The council ran nothing and recomputed no digest; every digest below is Main's `shasum -a 256` value. The council did check that the producer journal's `output_pins` and `predecessors` equal those digests. It decoded no CC2 source table and read no score, label, stratum or Jev value. All counts below are identifier **occurrences**, not unique identifiers and not case exclusions.

## Reviewed digests

Paths are relative to `evaluation/x1-triage-ranking/`.

| File | Bytes | SHA-256 |
|---|---:|---|
| `preparation-run/identifiers-delta-cc2/manifest.json` | 39502535 | `5ef5832cd49ca8dbfd0f4f0b192221012f0edb7d8dd076178b3ffe81b6cec4ea` |
| `preparation-run/identifiers-delta-cc2/coverage.json` | 9124 | `fde4052be405822a695d8c3fe7c0554bf2b73d1b05b3bd37ccc330e156174a2a` |
| `preparation-run/identifiers-delta-cc2/familiarity.jsonl` | 46132232 | `a9ef95b771fcb93edb5a56726f13233b6bb7874016e04300c2ef3639600d5191` |
| `preparation-run/journal/0378e1563617446ea7bdfc993819aa51.json` (CC2 producer, terminal) | 17458 | `ed5183888e744d68ce5599989f3a1564f65d38c2e5f15c327cc8b4c0ad8bbd11` |
| `preparation-run/journal/0378e1563617446ea7bdfc993819aa51.start.json` | 16169 | `2071c1e5d05dc55866c39e79ad768686b25b9e2598774ecf056b325746fd8c57` |
| `preparation-run/journal/906389d7212a401b8c7f781d37960d84.json` (baseline action d) | 10301 | `6aa8802a23ac0de75eea9a78de5e6422e8cb16c4762f875dc7ebf0787d77dc6c` |
| `preparation-run/journal/906389d7212a401b8c7f781d37960d84.start.json` | 8524 | `c400e46e3f2ba85767165257b425a9719a223303b71cf78700d3111ae06f8738` |
| `preparation-run/identifiers/manifest.json` | 27924749 | `23a911ae22084db67a5069a49c6db8ef4b0ec7a555c71348446d3372213c738d` |
| `preparation-run/identifiers/source-coverage.json` | 30151508 | `24f93c47ab0ab356c7bdafe7bda906c55508370deefbe75e6fad2c13fc682724` |
| `preparation-run/identifiers/delta-review.json` | 4635 | `240aa16d942f73e37d1c77b5055aad4a3158f08efcc63555833461e21276db04` |
| `preparation-run/identifiers/unresolved.json` | 91278 | `19d10781ceb31b7f4ae58d1a786de24a8af6b09fde925bafe134a398f395dcd9` |
| `preparation-run/identifiers/familiarity.jsonl` | 20588100 | `967518114b8187e8fcc4fb92615892d1cdab3a50e3cff6ab8fb10f3eb75e3f30` |

**Basis (not bound by this round):**
- Implementation receipt `preparation-reviews/identifiers-delta-cc2-implementation.json`: 4335 bytes, `155c9b5af5e1024e43ece698a2007bbf23d55905b4869d9fdc38303f367fcf14`.
- Its authority record `identifiers-delta-cc2-round-5.md`: 14854 bytes, `f5f453aa6f5437c855c144115145a952d28079255a6315098cdc340d236bbda8`.
- `identifiers-delta-cc2-REQUEST.md`: 20861 bytes, `ae6ba980…`.
- `preparation/exposure-sources.json`: `67dc5a4b…`.

## The run

Main ran the producer once with the canonical command (REQUEST `:92`). It exited 0 in about 4.5 s with no stdout or stderr.

The terminal journal `0378e156….json` records:
- action `d`, status `passed`, `exit_code` 0, `failure_reason` null;
- start 2026-11-10T15:25:30.030509Z, end 15:25:34.454528Z;
- canonical `argv`.

Its `output_pins` are exactly the three CC2 outputs above, in order manifest, coverage, familiarity, with the bytes and SHA-256 above.

Its nine `source_pins` are:
- the three CC2 modules, `test_cc2_delta.py`, the REQUEST (`ae6ba980…`), `projection.py` (`2408ed45…`), `preparation.py` (`9572aed3…`) and `verify_inputs.py` (`7d6f89ce…`);
- the source manifest `67dc5a4b…`.

These equal the receipt's reviewed rows. The other seven controls the gate requires are config `c817f678…` plus the baseline journal and its five outputs. The gate assembles them in memory (`cc2_delta.py:235–236`); they are not a journal field. Each output's `predecessors` lists exactly those six baseline pins, and each output's `implementation_review` pins the receipt `155c9b5a…`.

The start record `2071c1e5…` matches the terminal on every key that `collect_journals` compares (`preparation.py:617–619`):
- `id`, `action`, `started_at`, `argv`, `cwd`;
- `configuration`, `clock`, `source_pins`, `environment`, `effective_configuration`;
- `producer_pins` (six), `prior_journals` (16 files = the 8 legacy pairs).

The output root holds exactly the three files, and the journal directory holds nine start/terminal pairs. No refused or partial attempt exists.

## ID-01 — closed

Round 1 found these CC2 fields unextracted: `trail`, `S` (including the `base_oid` fallback), `base_oid`, `head_oid`, `merge_parents[]`, `files[].p`, partner paths, `commits[].oid` (which v4 labelled as blob), and the Django/Prometheus PR numbers.

### All 15 tables were projected

All 15 tables in the signed REQUEST schema table were projected, each pinned with source id, record index and exact field path (`manifest.json:1350599–1350750`; `coverage.json` `source_coverage`):

| id | Source (`evaluation/change-cost-v2/`) | Schema | Total | Selected | Other repo | Emitted |
|---:|---|---|---:|---:|---:|---:|
| 0 | `data/pr-django.jsonl.gz` | data | 1538 | 1538 | 0 | 5420 |
| 1 | `data/pr-prometheus.jsonl.gz` | data | 2229 | 2229 | 0 | 15003 |
| 2 | `raw/prs-django.json.gz` | raw | 1540 | 1540 | 0 | 17543 |
| 3 | `raw/prs-prometheus.json.gz` | raw | 2235 | 2235 | 0 | 38143 |
| 4 | `panel.csv.gz` | panel | 7717 | 1647 | 6070 | 1647 |
| 5 | `panel-final.csv.gz` | panel-final | 7717 | 1647 | 6070 | 1647 |
| 6 | `candidates.jsonl.gz` | candidate | 18083 | 1496 | 16587 | 2992 |
| 7 | `types.jsonl` | type-call | 10200 | 1855 | 8345 | 1855 |
| 8 | `labels.jsonl` | pair-call | 60342 | 4767 | 55575 | 9534 |
| 9 | `validation/type-sample.jsonl` | type-sample | 150 | 24 | 126 | 135 |
| 10 | `validation/type-blinded.jsonl` | type-blind | 150 | 150 | 0 | 712 |
| 11 | `validation/pair-sample.jsonl` | pair-sample | 200 | 19 | 181 | 192 |
| 12 | `validation/pair-blinded.jsonl` | pair-blind | 200 | 200 | 0 | 2293 |
| 13 | `validation/dep-sample.jsonl` | dep-sample | 128 | 16 | 112 | 32 |
| 14 | `validation/dep-blinded.jsonl` | dep-blind | 128 | 128 | 0 | 256 |

- The data-file byte counts (173309 and 275606) are the same files round 1 found yielding one identifier each.
- Every source byte count equals its `exposure-sources.json` entry (`:1640,1696,1871,1927,1815,1822,1591,11237,1808,11293–11377`).
- The producer refuses on any SHA-256 mismatch before decoding (`cc2_delta.py:144–152,256–258`), and it passed.

### Every missing field is now present for both repositories, with the correct kind

A jq aggregation of `manifest.json .records` by repository, normalised field path and kind gave these counts:

| Field → kind | django/django | prometheus/prometheus |
|---|---:|---:|
| `S` → commit | 735 | 1127 |
| `trail` → commit | 734 | 1121 |
| `base_oid` → commit | 1540 | 2235 |
| `head_oid` → commit | 1540 | 2235 |
| `merge_commit` → commit | 1540 | 2235 |
| `merge_parents[]` → commit | 1540 | 3526 |
| `commits[].oid` → **commit** | 1853 | 7515 |
| `commits[].parent` → commit | 1853 | 7515 |
| `files[].p` → path | 1286 | 2717 |
| `files[].path` → path | 6137 | 10647 |
| `partners[].g` / `.via` / `.via_ref` / `.vias[]` → path | 283 / 283 / 253 / 308 | 2117 / 2117 / 956 / 2619 |
| sample `files[]`, `earlier/later.files[]`, `file_a/file_b` → path | 44, 8/13, 2/2 | 67, 64/69, 14/14 |
| `number` → pr | 5194 | 7521 |
| candidate/label `parent` / `follow` → **pr** | 1103 / 1103 | 5179 / 5179 |

The unscoped blinded paths are 712 (`files[]`), 1340 + 953 (`earlier/later.files[]`) and 128 + 128 (`file_a/file_b`), for 3261 in total. Each has `repository: null`, rule `unresolved_repository_scope_unknown_freshness_ineligible`, and applies to all three test repositories (`familiarity.jsonl:93921–93930,94825–94833,97150–97151`).

**Matching rules.**
- Every commit uses rule `commit_or_parent_either_image_any_scope` and applies to all three test repositories (e.g. `familiarity.jsonl:3–4,5423–5424,20426–20431,37969–37975`).
- PRs use `repository_scoped_pr`, and paths use `repository_scoped_path_unknown_mechanism`.
- Skipped data rows keep their PR number (`familiarity.jsonl:63–64`).

**Other observations.**
- Django C0 `d88ec42b…` appears as a CC2 `merge_commit` (`familiarity.jsonl:20444`). This is consistent with round-1 finding 7.
- No repository other than the two targets or null appears.
- No `field_path` lies outside the allowlist.

### Counts reconcile

- Coverage rows sum to 97404 = `emitted_identifiers` = commit 38844 + path 33281 + pr 25279 (`manifest.json:8–12`).
- PR occurrences reconcile exactly: 3767 + 3775 + 1647 + 1647 + 2992 + 1855 + 9534 + 24 + 38 = 25279.
- Non-PR occurrences reconcile exactly: data and raw give 68567, and samples and blinded files give 3558, so 68567 + 3558 = 72125 = 38844 + 33281.
- `familiarity.jsonl` has the source-table header on line 1, then 97404 rows in source order. The source boundaries fall at lines 5422, 20425, 37968 and 97150, exactly as the counts predict.

### Contract

- `manifest.json` records:
  - `merge_semantics` `union_with_frozen_v4_and_existing_X1_identifiers_never_replace`;
  - `unknown_freshness` `ineligible`;
  - `candidate_exclusions_computed` false.
- `development_freshness_filter` is false in all three outputs.
- Case-insensitive key searches across the three outputs for score, rank, label, verdict, stratum, threshold, outcome, probability, choice, confidence, p_corrective, disagreement, title, description, author and similar returned no matches. "labels" occurs only as the source path `labels.jsonl`.

### What ID-01 closure covers

The delta records carry source, record and field provenance, so D-1's provenance gap does not apply to them. D-1 to D-5 remain in force for the v4 records.

The CC2 files that this delta does not project were disposed of in the signed REQUEST. That includes the CB-1 transport cache `raw/q`. `pins.json` was already parsed by v4 (`identifiers/source-coverage.json:8171–8183`, `parsed_identifier_projection`). No CC2 table with Django or Prometheus rows remains unprojected. The CC2 aliases contain no TypeScript entry, and an unknown alias would have refused the run.

## ID-02 — closed

The correct baseline is `906389d7212a401b8c7f781d37960d84.json`. It records:
- action `d` (`review_identifiers.py`), status `passed`, exit 0;
- start 2026-11-09T13:13:46.478896Z;
- `output_pins` equal to the five baseline digests (`:187–213`);
- `start_record` `c400e46e…` (`:227–231`).

The CC2 producer binds it in three places:
- it refuses unless its pins are current (`cc2_delta.py:155–163`);
- every output lists it as its first predecessor;
- its `prior_journals` lists the journal and its start record.

The seal hardcodes the same path (`preparation.py:544–551`). Journal `d8414b51…` is action b. It appears only in `prior_journals`, as part of the full journal history, and is not bound as the baseline anywhere.

`preparation-run/reviews/` does not exist yet, so no wrong receipt was ever written. This record binds `6aa8802a…` and `c400e46e…` in the `identifier-completeness-delta` scope block below. That is the fix round 1 prescribed.

## Seal readiness [INFERENCE: read from code, not run]

Based on `preparation.py:437–461,532–591,604–650,709–736`:
- The three outputs satisfy `collect_cc2_delta`'s checks: schema, status, clock (`26a1212a…`), receipt pin, source manifest, predecessors and append-union fields.
- The familiarity header's `sources` equals the manifest's.
- Exactly one passed action-d journal has these `output_pins`, and its `source_pins` include all code and source controls.
- The baseline outputs still match the `906389d7…` `output_pins`, which `verify_scoped_evidence` requires.
- The seal also re-hashes every pin embedded in the outputs: the receipt, the round-5 record, the 16 CC2 producer files and the 15 CC2 tables.

## New issues

None.

## Suggestions (not blocking)

1. **Report the blinded-path over-exclusion.** The 3261 unscoped blinded path occurrences include generic names such as `CHANGELOG.md`, `go.mod`, `go.sum` and `package.json`, and they apply to all three test repositories. At inventory, report per repository the exclusions caused only by these paths, as a sensitivity figure (CC2 round-1 suggestion 5).
2. **Deduplicate at inventory without losing provenance.** The 97404 counts are occurrences; for example `head_oid` often repeats `commits[].oid`. Deduplicate values but keep every source, record and field citation.
3. **Match SHAs regardless of kind.** Match every 40-hex value across v4 and this delta whatever kind it was assigned, including v4's `raw/q` `oid` values (output round-1 suggestion 1).
4. **Map PR numbers through commit identifiers.** Map PR numbers to cases through emitted commit identifiers, not by parsing `#N` from commit messages. Django messages cite Trac ticket numbers [INFERENCE].
5. **Optional `raw/q` recount** by object size (CC2 round-1 suggestion 7) remains open. The `pins.json` part of that suggestion is satisfied above.

## Peers and where they changed the recommendation

- **innovation-challenger (IdOutR2Challenger):** independently confirmed:
  - all 15 tables and per-source conservation;
  - every field kind for both repositories;
  - the journal pins and start/terminal agreement;
  - the baseline as the correct ID-02 target;
  - the exact receipt requirements.

  It pointed out that the journal has no literal `control_pins` field; that clarification is used above. It did not change the verdict.
- **experiment-peer (IdOutR2Peer):**
  - Supported H1–H6 (completeness, field kinds, no leakage, conservation, provenance, merge contract).
  - Ruled out the confounds: rows from other repositories, dropped Django or Prometheus rows, lost PR numbers on skipped rows, scoped blinded paths, and `parent` misread as a commit.
  - Contributed suggestions 1 and 4.

  It did not change the verdict.
- **Council:** made the same findings through direct reading, a jq aggregation and the count reconciliation.

## Conditions

- **Receipts.** Build both with `verify_inputs.pin()` on absolute paths, with `record` set to this file:
  - `preparation-run/reviews/identifiers.json`, scope `identifier-completeness-delta`: the five baseline outputs plus `906389d7….json` and its start record;
  - `preparation-run/reviews/identifier-delta-cc2-output.json`, scope `identifier-delta-cc2-output`: the three CC2 outputs plus `0378e156….json` and its start record.
- **Before building them**, re-run `shasum -a 256` and `wc -c` on the 12 files above and confirm they match.
- **Until the seal, change none of these:** the outputs, journals, receipts, this record, the 16 implementation controls, the round-5 record, or the CC2 source and producer files.
- **Inventory.** This lifts round 1's identifier blocker on inventory. Every other preparation prerequisite is unchanged.

Verdict: ZERO ISSUES — signed off

**Zero issues. Signed off.**

Signed-off scope: identifier-completeness-delta

- `preparation-run/identifiers/manifest.json` — 27924749 bytes — `23a911ae22084db67a5069a49c6db8ef4b0ec7a555c71348446d3372213c738d`
- `preparation-run/identifiers/source-coverage.json` — 30151508 bytes — `24f93c47ab0ab356c7bdafe7bda906c55508370deefbe75e6fad2c13fc682724`
- `preparation-run/identifiers/delta-review.json` — 4635 bytes — `240aa16d942f73e37d1c77b5055aad4a3158f08efcc63555833461e21276db04`
- `preparation-run/identifiers/unresolved.json` — 91278 bytes — `19d10781ceb31b7f4ae58d1a786de24a8af6b09fde925bafe134a398f395dcd9`
- `preparation-run/identifiers/familiarity.jsonl` — 20588100 bytes — `967518114b8187e8fcc4fb92615892d1cdab3a50e3cff6ab8fb10f3eb75e3f30`
- `preparation-run/journal/906389d7212a401b8c7f781d37960d84.json` — 10301 bytes — `6aa8802a23ac0de75eea9a78de5e6422e8cb16c4762f875dc7ebf0787d77dc6c`
- `preparation-run/journal/906389d7212a401b8c7f781d37960d84.start.json` — 8524 bytes — `c400e46e3f2ba85767165257b425a9719a223303b71cf78700d3111ae06f8738`

Signed-off scope: identifier-delta-cc2-output

- `preparation-run/identifiers-delta-cc2/manifest.json` — 39502535 bytes — `5ef5832cd49ca8dbfd0f4f0b192221012f0edb7d8dd076178b3ffe81b6cec4ea`
- `preparation-run/identifiers-delta-cc2/coverage.json` — 9124 bytes — `fde4052be405822a695d8c3fe7c0554bf2b73d1b05b3bd37ccc330e156174a2a`
- `preparation-run/identifiers-delta-cc2/familiarity.jsonl` — 46132232 bytes — `a9ef95b771fcb93edb5a56726f13233b6bb7874016e04300c2ef3639600d5191`
- `preparation-run/journal/0378e1563617446ea7bdfc993819aa51.json` — 17458 bytes — `ed5183888e744d68ce5599989f3a1564f65d38c2e5f15c327cc8b4c0ad8bbd11`
- `preparation-run/journal/0378e1563617446ea7bdfc993819aa51.start.json` — 16169 bytes — `2071c1e5d05dc55866c39e79ad768686b25b9e2598774ecf056b325746fd8c57`
