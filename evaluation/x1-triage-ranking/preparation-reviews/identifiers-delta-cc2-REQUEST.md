# Request: X1 append-only CC2 identifier delta, per-schema field allowlist

**Prospective request, not authority or sign-off.** ID-01 in `identifiers-output-round-1.md` blocks inventory. Review this field allowlist and its implementation to zero issues **before** Main runs the delta producer. The producer has not been run on real CC2 data. The implementation assignment examined producer source and signed source-file metadata only, never CC2 data payloads. The existing X11 v4 extraction and X1 identifier outputs remain immutable.

## Authority, scope and exclusions

This requests one action-d append-only delta, limited to prior change-cost-v2 identifiers relevant to Django and Prometheus, plus conservatively unscoped identifier-only paths in the corresponding blinded validation exports. It does not authorise inventory, candidate scanning, new model calls, comparisons, labels, analysis or re-extraction of v4. Existing output-review dispositions D-1–D-5 remain in force. Unknown freshness stays ineligible.

Source authority is the complete signed `preparation/exposure-sources.json`, SHA-256 `67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60`. Every selected table and every producer-code file must match that manifest's exact SHA-256 and byte count. No moving checkout, enumeration of current evaluations, source refresh, fetch or repin substitutes for these pins. Every selected source table is pinned once in a source table; records reference source ID, original record index and exact selected field path. Source code is verified as bytes but is never imported or executed.

Only repository context, PR numbers, snapshot/trailing/base/head/merge/parent commit identifiers and repository-relative file/partner paths may be materialised. **No score, label, rank, verdict, disagreement, threshold, stratum, outcome, Jev response/probability, prose, source excerpt, author, timestamp or measurement value may be decoded or used.** Known excluded field names occur only in the finite schema registry so their values can be skipped. Unknown projected-container keys or structures refuse; there is no generic recursive mining or permissive unknown-key fallback.

## Code-derived schema mapping

All paths in this table are relative to `evaluation/change-cost-v2/`. The implementation registry is `execution/identifiers/cc2_schemas.py`; it contains the finite selected fields and finite known-but-opaque excluded fields for each schema. The following mappings are derived from producer code, not from examining real rows.

| Sources | Producer evidence | Materialised identifier fields and meaning |
|---|---|---|
| `data/pr-django.jsonl.gz`, `data/pr-prometheus.jsonl.gz` | `collect.py:138–143,154–191,218–238,246–261` | `repo` for context; `number` as PR; `S` and `trail` as commits, including an S selected from the base fallback; `files[].p` as path; `partners[].g`, `via`, every `vias[]` entry and non-null `via_ref` as paths. Skipped/early-return rows still contribute their PR number. No predicate reads `skip` or any measurement. |
| `raw/prs-django.json.gz`, `raw/prs-prometheus.json.gz` | `fetch_prs.py:86–107,124–133` | Filename fixes repository scope; `number` as PR; non-null `base_oid`, `head_oid`, `merge_commit`, every `merge_parents[]`, `commits[].oid` and non-null `commits[].parent` as **commit**, never blob; `files[].path` as path. |
| `panel.csv.gz`, `panel-final.csv.gz` | `build_panel.py:124–150,174`; `finalize_panel.py:33–64` | `repo` and `number` only. The exact code-derived header is checked, but no other cell is decoded. `main_path` is not selected: it is already a path from the raw/data tables (`build_panel.py:121–122`). Appended final-panel outcome columns are opaque and never read as values. |
| `candidates.jsonl.gz` | `follow.py:36–81` | `repo`, `parent`, `follow`, where the latter two are **PR numbers**, not commit parents. `shared` is not selected: producer lines 70–76 show it is a subset of the raw touched paths and collected partner paths already fully covered above. Timing, candidate kind, mention and censoring values are opaque. |
| `types.jsonl` | `jev_type.py:44–61,85–97` | `repo`, `number` only. Every Jev call/response/probability/status field is opaque. No task class or label is inferred or emitted. |
| `labels.jsonl` | `jev_label.py:47–64,92–100` | `repo`, `parent`, `follow` PR numbers only. All Jev fields are opaque; no label, probability or decision is read. |
| `validation/type-sample.jsonl` | `validate.py:102–105,151–158` | `repo`, `number` and every `files[]` path string. All text, labels and pipeline class are opaque. |
| `validation/pair-sample.jsonl` | `validate.py:119–123,151–158` | `repo`, PR `parent`/`follow`, and every `earlier.files[]` / `later.files[]` path string. `stratum`, `kind`, text and timing are opaque. |
| `validation/dep-sample.jsonl` | `validate.py:131–150,151–158` | `repo`, partner paths `file_a` and `file_b`. Language, proxy, partner-group key and source excerpts are opaque. No sample or grouping ID is selected. |
| `validation/{type,pair,dep}-blinded.jsonl` | `validate.py:151–158` | Same path keys as the matching sample export, but no repository or PR context is present. Preserve these paths as unresolved scope, applicable conservatively to every X1 test repository. Do not recover scope from labels, strata or opaque sample IDs. |

Repository context is restricted to the sixteen aliases in `common.py:21–39`. Only `django` and `prometheus` scoped rows contribute identifiers from combined tables; other documented repository rows are counted separately and not used for this delta. Unrecognised or inconsistent scope refuses. Blinded path-only records remain unscoped and cannot clear freshness.

### Other files and why no new field handling is requested

The signed source metadata identifies the complete prior CC2 asset set. This request covers all consolidated PR/candidate/panel/type/pair/dependency case tables produced by the code above, including the tables carrying Jev fields without reading those fields.

* `validation/annotations-{A,B,C}.jsonl` and their `validation/run-1-failed/` predecessors contain opaque validation IDs and model annotations (`annotate.py:93–111`). They add no repository, PR, commit or path identifiers beyond the sample/blinded source tables. No value from them is parsed.
* Aggregate results, gates, constructs, power, sensitivity, prompts, rubric and strata documents contain analyses/configuration rather than additional case-level identifier tables. No value from them is parsed. Earlier accepted documentation/script dispositions are not reopened.
* `raw/q/*.json.gz` is the GraphQL transport cache (`fetch_prs.py:22–58`), not the consolidated PR export. This request does **not** authorise another GraphQL envelope projection or decoding cache values. The consolidated `raw/prs-*` files are the producer's identifier export (`fetch_prs.py:69–107,124–133`); the requested raw-table projection uses those exact pinned files. Council must explicitly accept this transport-cache boundary, or request a separately specified GraphQL-cache field allowlist before any cache handling. There is no implicit claim that an unreviewed cache schema grants freshness.
* Other-repository `data/pr-*` and `raw/prs-*` files are outside this Django/Prometheus-specific delta. No data payload from them is opened. Existing any-scope SHA evidence from v4 remains retained; nothing is removed or cleared.

These are explicit source-disposition requests, not self-authored completeness approval. If council finds any additional case table or identifier-bearing schema, stop that handling and specify its code-derived field allowlist in a new review before running it. Do not inspect forbidden values to decide scope.

## Selective-reading implementation

Files: `execution/identifiers/{cc2_delta.py,cc2_readers.py,cc2_schemas.py}` and `execution/tests/test_cc2_delta.py`.

`cc2_readers.py` scans JSON framing at the byte level. Only schema field names and selected scalar identifier tokens reach `json.loads`; excluded values and their descendants are structurally skipped without JSON decoding, string decoding, numeric conversion or per-value retention. Selected object/list shapes are finite; unknown keys, duplicate projected keys, unexpected containers and invalid identifiers refuse. For CSV, the exact producer-derived header is checked, cells are located as raw spans, and only `repo` and `number` cells are decoded. A full `json.load`, `csv.reader`, pandas load or generic nested traversal of the raw source is prohibited. Gzip decompression and whole-source byte hashing are transport operations, not permission to decode forbidden fields.

Synthetic tests plant forbidden sentinel values under known excluded fields and spy on JSON/CSV scalar decoding to prove they are not decoded; an unknown forbidden field refuses before its value is selected. They also test the formerly mislabelled `commits[].oid` as commit, snapshot/trail and every partner-path spelling, both panel schemas, PR-parent versus commit-parent context, finite schema refusal, source pins, review binding, create-once/partial-output retention, journal action d and unchanged baseline outputs.

## Execution interlock and retained records

The CLI requires canonical config, clock, sources, review receipt and output root. It requires the pinned CPython executable. Before source projection it verifies the authentic scoped receipt with `preparation.review_receipt` for scope `identifier-delta-cc2-implementation`. That receipt must bind all three CC2 implementation modules, `execution/identifiers/projection.py`, `execution/preparation.py`, `execution/verify_inputs.py`, synthetic tests, this request, config, signed source manifest and baseline identifier/journal pins. The imported dependency pins appear in both the journal's `source_pins` and the gate's `control_pins`; any changed dependency requires fresh review. The authentic council record must contain the exact standalone verdict line `**Zero issues. Signed off.**` and exactly one standalone line `Signed-off scope: identifier-delta-cc2-implementation`. After that scope line and before the next `Signed-off scope:` line (or end of record), the scope block must name every one of the 16 bound SHA-256 digests in full. Booleans, prose-only approval or a digest named only in another scope block do not grant authority. Generate the receipt's reviewed rows from `pin()` on the exact producer paths.

ID-02 is not a code change to prior evidence. The baseline run bound by this producer is specifically `preparation-run/journal/906389d7212a401b8c7f781d37960d84.json`, **10,301 bytes**, SHA-256 **`6aa8802a23ac0de75eea9a78de5e6422e8cb16c4762f875dc7ebf0787d77dc6c`**, not the action-b journal named in the earlier packet. It records action d, status passed and exit 0. Its start record is `preparation-run/journal/906389d7212a401b8c7f781d37960d84.start.json`, 8,524 bytes, SHA-256 `c400e46e3f2ba85767165257b425a9719a223303b71cf78700d3111ae06f8738`. This baseline journal digest must be named in the implementation sign-off and its receipt before the producer runs; it is not deferred to output review.

The baseline journal binds these five existing outputs, in order:

| Identifier output | Bytes | SHA-256 |
|---|---:|---|
| `manifest.json` | 27,924,749 | `23a911ae22084db67a5069a49c6db8ef4b0ec7a555c71348446d3372213c738d` |
| `source-coverage.json` | 30,151,508 | `24f93c47ab0ab356c7bdafe7bda906c55508370deefbe75e6fad2c13fc682724` |
| `delta-review.json` | 4,635 | `240aa16d942f73e37d1c77b5055aad4a3158f08efcc63555833461e21276db04` |
| `unresolved.json` | 91,278 | `19d10781ceb31b7f4ae58d1a786de24a8af6b09fde925bafe134a398f395dcd9` |
| `familiarity.jsonl` | 20,588,100 | `967518114b8187e8fcc4fb92615892d1cdab3a50e3cff6ab8fb10f3eb75e3f30` |

The producer journals action d with the existing shared operation lock and immutable start/terminal records. It verifies every selected source pin before decoding any table, and parses the same bytes it hashed. It writes create-once:

* `preparation-run/identifiers-delta-cc2/manifest.json`: identifier occurrences with source/schema/record/field provenance and the source-pin table;
* `coverage.json`: per-source input/selected/other-repository/output counts;
* `familiarity.jsonl`: initial source table followed by append-union evidence records, not actual candidate exclusions.

Output directory reservation, completed files and partial bytes survive failure. Each filename is registered immediately after exclusive creation, including a partly written familiarity log. A terminal journal pins only files created by that invocation: an existing-root refusal has no output pins and does not claim earlier outputs. Existing v4 outputs, all five existing X1 identifier outputs and prior journals are never edited or replaced.

## Inventory merge contract and unresolved prerequisites

Later authorised inventory consumes the **union** of v4, existing X1 identifier records and every CC2 delta occurrence. It must never choose one manifest in place of another. Every commit identifier matches commit/parent on either image in all three test repositories, regardless of its source scope. PR numbers and paths remain repository-scoped; unscoped blinded paths apply conservatively to all three tests and remain unknown-freshness ineligible. Existing rename-closure and exclusion-provenance dispositions still apply. No actual case exclusion is computed here.

The integration submitted for review is `preparation.py` at `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38`, with unchanged `verify_inputs.py` at `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310`. It requires the three canonical delta output kinds, their unique passed action-d producer, the complete 16-control implementation receipt, and the separate `identifier-delta-cc2-output` review of all three outputs and that producer journal. It preserves the original identifier outputs and their review scope, canonical output roles, and the current complete execution-source closure check. These bytes still need zero-issue Core and delta seal-integration reviews before the producer runs; this request is not approval. Neither the producer nor this request changes `verify_inputs.py` or `run-config.json`, and previously signed helper pins are not silently waived. Main must bind these exact integrated helper bytes in the confirming implementation review.

## Main command, after zero-issue review only

**Mandatory freeze order:** run the producer only after the Core implementation review and the CC2 delta seal-integration review have both passed with zero issues. Both sign-offs must name the same final `preparation.py` and `verify_inputs.py` digests that the confirming CC2 implementation sign-off and receipt bind. Complete all edits first; then hold a confirming CC2 round at those exact final bytes naming all 16 bound digests in its `identifier-delta-cc2-implementation` scope block; then create the authentic receipt; only then make the single canonical producer run. An earlier zero-issue CC2 round at intermediate shared-code bytes is not execution authority.

After the producer run, none of the 16 bound files may change before the seal. In particular, do not edit the shared helpers, imported modules, synthetic tests, this request or config after their pins enter the passed create-once journal. A newer receipt cannot repair stale pins in a passed journal. The immutable baseline source/journal/output pins must remain intact as well. New output-review records are separate append-only records, not modifications of the bound set.

After the run, bind the delta output review at `preparation-run/reviews/identifier-delta-cc2-output.json` with scope `identifier-delta-cc2-output`, before sealing. The Core integration must enforce that output review and the existing implementation-receipt chain; publishing delta files alone does not establish review or inventory authority.

From `/Users/adamjackson/Projects/work/scout-x1prep`:

```sh
.tools/venv/bin/python -I -B evaluation/x1-triage-ranking/execution/identifiers/cc2_delta.py --config evaluation/x1-triage-ranking/execution/run-config.json --clock evaluation/x1-triage-ranking/preparation-run/clock.json --sources evaluation/x1-triage-ranking/preparation/exposure-sources.json --review evaluation/x1-triage-ranking/preparation-reviews/identifiers-delta-cc2-implementation.json --output-root evaluation/x1-triage-ranking/preparation-run/identifiers-delta-cc2
```

Main must first create the authentic implementation review receipt at the canonical review path after zero-issue review. No privileged access, network, models or credentials are required. After the single real run, submit every output pin and the new action-d journal for a separate identifier-output completeness/delta round. That later review must resolve ID-01 and bind corrected ID-02 provenance before inventory can begin.

## Confirming implementation packet (prospective, not sign-off)

All 16 receipt paths are listed below. These are recomputed metadata pins, not authority. The request is itself the sixteenth control: its final SHA-256 and byte count are recorded in `core-impl-round-4-response.md` after this file is written, avoiding a circular self-digest. The confirming council record must include that actual final request digest along with the other 15 full digests in its scope block; a self-pin description cannot substitute for the digest in that record or receipt.

| Absolute receipt path | Bytes | SHA-256 |
|---|---:|---|
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/identifiers/cc2_delta.py` | 14393 | `657aeb47ee11cbe4db4fdba3948639cf211f8ff032844cbe6d77a815086c0ea6` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/identifiers/cc2_readers.py` | 8012 | `1ad375a97b43a771245f3f746768dc72814656c1c81c8c4d8793a9c88644144b` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/identifiers/cc2_schemas.py` | 4464 | `db845451e7dbefc8f32ea7b50fbe86cdd5c702330f755e39cee345b2cfc5799f` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/tests/test_cc2_delta.py` | 22402 | `4844221de3ecb6c52b81a770ad7dadbcc42d8e0122636a29de2cdb1fc0e9d8d7` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-reviews/identifiers-delta-cc2-REQUEST.md` | Final request pin in core-impl-round-4-response.md | Final request pin in core-impl-round-4-response.md |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/identifiers/projection.py` | 7637 | `2408ed459b98ca991bb834e1b4f46437b30364e7e81ff3879b7e631ad560e072` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/preparation.py` | 50257 | `9572aed3b10fe76999e503ef690518febda972340a56f1f1af33c24ad39c5f38` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/verify_inputs.py` | 11211 | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/run-config.json` | 5948 | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation/exposure-sources.json` | 7792170 | `67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/journal/906389d7212a401b8c7f781d37960d84.json` | 10301 | `6aa8802a23ac0de75eea9a78de5e6422e8cb16c4762f875dc7ebf0787d77dc6c` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/manifest.json` | 27924749 | `23a911ae22084db67a5069a49c6db8ef4b0ec7a555c71348446d3372213c738d` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/source-coverage.json` | 30151508 | `24f93c47ab0ab356c7bdafe7bda906c55508370deefbe75e6fad2c13fc682724` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/delta-review.json` | 4635 | `240aa16d942f73e37d1c77b5055aad4a3158f08efcc63555833461e21276db04` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/unresolved.json` | 91278 | `19d10781ceb31b7f4ae58d1a786de24a8af6b09fde925bafe134a398f395dcd9` |
| `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/identifiers/familiarity.jsonl` | 20588100 | `967518114b8187e8fcc4fb92615892d1cdab3a50e3cff6ab8fb10f3eb75e3f30` |
