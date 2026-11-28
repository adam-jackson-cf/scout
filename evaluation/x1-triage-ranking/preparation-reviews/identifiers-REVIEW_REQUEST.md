# X1 test-only frozen-identifier completeness/delta review request

Draft request, not approval or inventory authority. Round 9 authorises preparation (a)–(e) only. Review this additive implementation before Main runs action (d), then review its exact retained outputs for completeness. Do not execute extraction, scanners or candidate collection as part of this request.

## Review scope and packet

Review `execution/review_identifiers.py`, `execution/identifiers/projection.py`, `execution/tests/test_identifiers.py`, `execution/run-config.json`, the execution source manifest, the clock and action-(d) journal records. Bind exact SHA-256s in the authentic council round, including this request and every reviewed file. The request itself does not supply a council sign-off.

Authority inputs are signed X1 `preparation/frozen-exposure-inputs.json` (SHA-256 `ec92fba6d3f91c671301244539a8caefc5025c0d6b03efca77c5bdfc4bd96425`) and `preparation/exposure-sources.json` (`67dc5a4bd7322317cb3785e1ddf7ad1ccbe75a4c73a9cdeab5ccb4d6afbadf60`), together with authentic round 9. Review all eight input pins from the frozen manifest, its original seal/source/context/output cross-links, upstream Scout `04d5e632ae34d8056c3df5116b9ba87cfbf99282` and inertia `85e7329bef230358740c2b7ae90b6364a15b675c`. No fresh Git or filesystem enumeration of either snapshot is performed.

After Main runs the reviewed implementation, include `preparation-run/identifiers/{manifest.json,source-coverage.json,delta-review.json,unresolved.json,familiarity.jsonl}` and their complete hashes, plus source/clock/config/journal pins. Review complete tables, not just counts. `status: passed` denotes successful projection/metadata reconciliation only; it does not assert council sufficiency or freshness clearance.

## Frozen schema projection

Only the five existing identifier/coverage/enumeration JSON outputs and contexts/seal metadata are decoded. The extraction script is hashed as bytes, never imported or run. Prior Scout/inertia study sources are never opened or parsed. No score, rank, label, verdict, disagreement, threshold or outcome field is selected, copied or mined. Unknown top-level/record schemas become unresolved metadata or refusal; their values are not recursively extracted.

* Scoped records: exact `repository,study,task` plus one `commit`, `blob`, `path`, `pr` or `repository_identity` string. Exposure output additionally carries `studies`; the other two carry only `records`.
* Unattributed records: exact `source,study,task,kind,identifier,matchability`. Preserve every supported record, including unmatchable path/PR scope.
* Coverage: exact v2 envelope and the finite source row keys in `projection.COVER_KEYS`. Source metadata is compared by path/Git blob and available SHA-256/byte count, never by reading raw source files. Blob-only matches are explicitly distinguished from matches with SHA-256 metadata.
* Inertia: existing enumeration `repository,commit,recursive_listing,enumeration_result,scope,data_files`, checked against existing coverage. Account for it separately, never as part of the 24,254-file signed Scout universe.

The historical scoped output deduplicates records and has no per-source-file or source-field mapping. It also emits no issue/evidence hash, span, mechanism, rename link or excerpt hash. Provenance gaps are unresolved once per output with exact index ranges and counts, never invented. Manifest and familiarity entries group by output/study/task/canonical repository/kind/rule/applicability, preserving parallel record indices, values and source-table IDs. A once-per-source table records available unattributed source digests and unique/missing/ambiguous resolution. Every output references frozen files by name; actual pins appear once in its top-level inputs. Count tables report conservation, repository/study/kind/rule totals, unresolved reasons and familiarity sources for every test repository.

## Matching and completeness decisions

Only django/django, microsoft/TypeScript and prometheus/prometheus are X1 test repositories. Apply the seven historical aliases in `projection.ALIASES`, and canonical case for the three test slugs. Prometheus and other-task familiarity are retained. Original scope stays provenance, never a restriction on SHA matching; development never uses freshness filtering.

Every commit and non-empty blob applies to all three test repositories regardless of its original scope, using `commit_or_parent_either_image_any_scope` and `blob_either_image_any_scope`. This avoids inherited-context, fork, mirror and filename mis-scoping creating false freshness. The Git empty blob remains a non-match. Paths and PRs require documented repository scope; unknown scope remains unresolved. Repository familiarity never clears freshness. No actual candidate exclusion is computed during preparation. Later authorised inventory appends every actual case exclusion and source without rewriting initial evidence.

Review matched/changed/missing/new/unknown-schema source sets and all non-parsed statuses. The delta is a problem list, not permission to inspect raw data. **Unknown freshness is ineligible.** Explicitly decide whether disclosed unresolved provenance/coverage is sufficient for X1 test-only completeness under that conservative rule. A council refusal blocks inventory; unknowns must not silently become fresh or automatically imply a whole-study failure.

Conservation is checked per frozen output against every original index, both before and after encoding. Emitted group index/value pairs must reconstruct projected records exactly; those indices plus emitted record-level rejects must cover every input once. Aggregated scope and provenance counts/ranges must reconstruct their expected indices exactly. Compatible/unattributed/unscoped/coverage/inertia counts are checked against the pinned seal summary. The frozen inertia coverage path set must equal its recursive listing with no duplicates. Council must inspect these emitted tables as well as unresolved gaps.

## Separate authority required for any new raw schema

No new handling is requested or authorised by this packet. If completeness needs it, stop that handling and submit a separate per-schema request naming repository/source, exact schema and allowlisted fields: case/issue identifier, commit/parent, blob/evidence digest, repository-relative path/span, mechanism/rule, rename link and identical-excerpt hash only. Scores, ranks, labels, verdicts, disagreements, thresholds and outcomes remain excluded. Only a separately zero-issue reviewed append-only delta producer may then run. Never rerun `extract_exposures.py`, recursively mine arbitrary fields, replace frozen outputs or refresh inertia.

## Required authentic outcome

Report every issue with severity and exact fix. Iterate implementation review and output sufficiency review to **ZERO ISSUES**; record every round with exact reviewed SHA-256s and an explicit **Signed off** statement for scope `identifier-completeness-delta`. Bind that human-readable council record and reviewed pins through canonical `preparation-run/reviews/identifiers.json` per `run-config.json`'s review contract. State unresolved coverage/provenance disposition explicitly. This scoped review grants neither inventory/SRS nor model/dev/full-OC/GUIDE authority; the later preparation-to-inventory review remains required.

Main command after pinned Python and zero-issue implementation review:

```sh
.tools/venv/bin/python -I -B evaluation/x1-triage-ranking/execution/review_identifiers.py --config evaluation/x1-triage-ranking/execution/run-config.json --clock evaluation/x1-triage-ranking/preparation-run/clock.json --frozen evaluation/x1-triage-ranking/preparation/frozen-exposure-inputs.json --sources evaluation/x1-triage-ranking/preparation/exposure-sources.json --output-root evaluation/x1-triage-ranking/preparation-run/identifiers
```

A pre-existing output directory, missing/changed input or partial failed operation requires retained evidence and reviewed recovery, not deletion/retry under another path. No privileges, network or credentials are required. Main must run synthetic unit tests with the pinned interpreter before review; these are fabricated metadata tests, not study observations.
