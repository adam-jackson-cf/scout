# X1 preparation (d) identifier completeness/delta — output review — round 1

**Scope.** This round reviews whether the retained action-(d) outputs are complete enough, against the three test repositories (django/django, microsoft/TypeScript, prometheus/prometheus), for inventory to proceed under the rule that unknown freshness is ineligible. The review was read-only. It did not run the projector, any tests or any extraction, and it computed no candidate exclusions.

## Output files and SHA-256

Paths are under `evaluation/x1-triage-ranking/preparation-run/identifiers/`. The digests come from Main's `shasum -a 256` and match `output_pins` in the actual action-(d) journal (see ID-02).

| File | Bytes | SHA-256 |
|---|---|---|
| `manifest.json` | 27,924,749 | `23a911ae22084db67a5069a49c6db8ef4b0ec7a555c71348446d3372213c738d` |
| `source-coverage.json` | 30,151,508 | `24f93c47ab0ab356c7bdafe7bda906c55508370deefbe75e6fad2c13fc682724` |
| `delta-review.json` | 4,635 | `240aa16d942f73e37d1c77b5055aad4a3158f08efcc63555833461e21276db04` |
| `unresolved.json` | 91,278 | `19d10781ceb31b7f4ae58d1a786de24a8af6b09fde925bafe134a398f395dcd9` |
| `familiarity.jsonl` | 20,588,100 | `967518114b8187e8fcc4fb92615892d1cdab3a50e3cff6ab8fb10f3eb75e3f30` |

**Run record.** The journal is `preparation-run/journal/906389d7212a401b8c7f781d37960d84.json`:
- action `d`, status `passed`, exit 0;
- ran 2026-11-09T13:13:46.478896Z → 13:13:48.798317Z;
- start record `c400e46e3f2ba85767165257b425a9719a223303b71cf78700d3111ae06f8738`.

Its pins match the implementation signed off in round 8:
- source pins: `review_identifiers.py` `adc42d38…` and `projection.py` `2408ed45…`;
- config `c817f678…`;
- clock `26a1212a…`.

## Basis

- round-9 `6f1d6ca2…`
- PREREGISTRATION `abd206c3…` (l.19–23, l.37, l.194, l.222)
- exposure-recipe `c065347e…`
- exposure-sources `67dc5a4b…` (source commit `7bb73e52…`)
- frozen-exposure-inputs `ec92fba6…` (upstream Scout `04d5e632…`, inertia `85e7329b…`)
- request `identifiers-REVIEW_REQUEST.md` `10879ebd…` (digest as supplied by Main; not recomputed)
- implementation sign-off `identifiers-impl-round-8.md`

## Findings

1. **Conservation matches the seal summary.** Every count in the outputs reconciles with `.upstream_seal_metadata.summary`:

   | Quantity | Count |
   |---|---|
   | compatible records | 50,016 |
   | additional records | 8,796 |
   | unattributed records | 262,080 |
   | unscoped commit / blob / path / PR | 61,135 / 177,483 / 23,462 / 0 |
   | coverage parsed / document / unclassified | 24,001 / 321 / 3 |
   | inertia data files | 36 |

   - `record_issues` is 0 for every output.
   - The provenance-gap ranges cover 50,016 + 8,796 + 262,080 = 320,892 records.
   - All 1,300 unattributed sources resolve as `unique`.
   - The 74-path inertia listing (36 parsed, 38 unparsed) is counted separately from the 24,254-file Scout universe (24,251 matched + 3 new).

2. **The delta is clean except for 3 new files.** `delta-review.json .counts`: matched 24,251, changed 0, missing 0, unknown_schema 0, new 3. The three new files are `d1-triage-minimums-exploratory/{README.md,scores.csv,tuning.json}`.
   - By P l.19–21 and `reviews/round-1.md:250` (which cites the strata in `scores.csv`/`tuning.json`), D1 covers only the seven examined non-test repositories.
   - **Disposition:** these files are outside the test scope. No per-schema authority is needed for them.

3. **Prometheus familiarity is captured.** Prometheus-scoped groups in `familiarity.jsonl`:
   - change-cost-v2 (E1b-v2, R-035): 5,531 commits, 6,149 blobs, 1,095 paths and the repository identity;
   - inertia-bench design/screening: 2 commits, 1 path and repository identity in 6 groups.

   Django and TypeScript familiarity are captured in the same way. Other-task familiarity is retained. `encode/django-rest-framework` is correctly not counted as a test repository.

4. **Aliases, case and the empty blob are handled correctly.**
   - `projection.py:4–26` applies the 7 aliases and canonical case for the three test slugs.
   - Every commit and blob group applies to all three test repositories.
   - The Git empty blob is treated as a non-match.

5. **No forbidden fields were read or copied.** A case-insensitive grep across all five outputs for keys matching `score|rank|label|verdict|disagree|threshold|outcome` returned no matches.

6. **The commit, blob and path rules work without the missing provenance.** The scoped rows keep output, index and study, so a later exclusion can cite them even though v4 dropped source, field and mechanism provenance.

7. **The frozen extractor misses change-cost-v2 fields because of its key allowlist** (`extract_exposures.py:28–32`). Checked against the change-cost-v2 producer code, which is code and not data:
   - **`collect.py`** writes several identifier fields into `data/pr-<repo>.jsonl.gz`:
     - `S` (snapshot commit, `:154–169`);
     - `trail`, a first-parent commit measured with lizard at least 365 days before `S` (`:179–191`);
     - `files[].p` (`:187`);
     - co-change `partners` (`:197–218`).
   - **`fetch_prs.py:93–97`** writes `base_oid`, `head_oid`, `merge_parents` and `commits[].oid`.
   - **What v4 does with them:** none of `S`, `trail`, `p`, `base_oid`, `head_oid` or `merge_parents` is an allowlisted key. `oid` is classified as a blob key. `number` gets no PR context, because the records use `merged`, not `merged_at`.
   - **Observed result:** `pr-django.jsonl.gz` (173 KB) and `pr-prometheus.jsonl.gz` (276 KB) each yield one identifier, and there are no Django or Prometheus PR identifiers at all.
   - **What is captured:** `commits[].parent` (usually `S`), `merge_commit` and `files[].path`.
   - **Overlap with the X1 frames:** the test frames are 2,000 first-parent commits after C0 (P l.37). The peer found that both the Django and the Prometheus C0 are change-cost-v2 commits, so the windows overlap.

8. **The X11 directories are planning-only.** `x11-merge-association/` and `x11b-guard-recovery/` are outside both source manifests.
   - The X11 README says "No candidate source scan, inventory ranking, real map/label/replay or model call".
   - P l.23 excludes X11 planning from the freshness definition.
   - The main checkout has no further evaluation directories.
   - Inertia-bench `refs/heads/main` is `85e7329b…`, which is the frozen snapshot.

9. **The packet cites the wrong journal.** The journal Main named, `d8414b51d4ae4420a1dbbf4e300f67b5`, is an **action-b** record (`git submodule update --init`, 03:35 UTC, empty pins). It is not the run being reviewed. The real action-(d) journal is `906389d7…` (ID-02).

10. **Reviewer disclosure.** One grep I ran over `evaluation/change-cost-v2` returned matching lines from `validation/pair-sample.jsonl` and `pair-blinded.jsonl` (P086–P090). Those lines include a `stratum` value derived from Jev for one Prometheus PR pair (#14939/#14951). This is a prior-study label, not X1 data. I did not use it in any decision. That pair is a change-cost-v2 case, so it is already exposed.

## Dispositions decided in this round

These apply to inventory and are not open issues.

- **D-1 Provenance gap.** This is accepted as a recorded deviation from recipe step 4, which expected rows with source-file hashes and field provenance. The 320,892 records keep output, index and study. Every inventory exclusion must cite output, record index and study, plus `source_id` for unattributed records. This deviation must be carried into RESULTS under Deviations.
- **D-2 Paths.** These are ineligible:
  - any case touching a test-scoped path (`repository_scoped_path_unknown_mechanism`: Django 1,833, Prometheus 1,096, TypeScript 1);
  - any case touching any of the 23,462 paths with unknown scope, in all three repositories.
- **D-3 Spans, issue IDs, renames and excerpts.**
  - Path-level exclusion is a superset of span-level exclusion.
  - Issue IDs are moot, because the only case-level prior study on the test repositories (change-cost-v2) is PR-level with no Scout issues (P l.19–21).
  - Inventory must compute rename closure for exposed paths from the pinned test-repository Git history alone, without reading prior data. A rename-linked path counts as exposed.
- **D-4 The 324 sources not parsed for identifiers.** They are accepted as documents and scripts.
  - Rationale: Django and Prometheus case-level exposure comes only from change-cost-v2 data files. TypeScript has only repository-level inertia-bench screening. Inertia-bench at `85e7329b` is at the design stage: repository screening, with no dataset or labels.
  - `raw-SHA256SUMS`, `.gitattributes` and `.gitignore` are harmless.
- **D-5 D1 and X11.** Both are outside the test scope (findings 2 and 8).

## Issues

| ID | Severity | Finding | Required fix |
|---|---|---|---|
| ID-01 | major | **Change-cost-v2 commits and paths on the test repositories that v4 never extracted** (finding 7). The uncaptured fields are `trail`, `S` when it falls back to `base_oid`, `base_oid`, `head_oid`, `merge_parents[]`, `files[].p`, `partners[]` paths, and `commits[].oid`, which v4 labelled as blob. Concrete failure: an X1 Django or Prometheus run at a `trail` commit (or its child) has a case on a file that change-cost-v2 did not measure, or on a co-change partner file. Under the preregistered rule it should be excluded (role-agnostic conservative commit match, path mechanism), but inventory would judge it fresh. | Pause inventory. Submit a separate per-schema request for the change-cost-v2 Django and Prometheus records. It should cover `data/pr-{django,prometheus}.jsonl.gz`, `raw/prs-{django,prometheus}.json.gz`, and every other change-cost-v2 table with Django or Prometheus rows (for example `panel.csv.gz` and the follow-up/pair tables). Derive each table's identifier keys from the producer code. The allowlist is only: PR number, `S`, `trail`, `base_oid`, `head_oid`, `merge_commit`, `merge_parents[]`, `commits[].oid`/`parent` (emitted as **commit**), `files[].path`/`p`, and partner path keys. Scores, labels, strata, outcomes and Jev fields stay excluded. Run only a zero-issue-reviewed producer that appends a delta, then re-review its outputs. |
| ID-02 | minor | **The packet binds the wrong run record.** Main's packet names journal `d8414b51…`, which is an action-b record with empty pins. Binding it in `preparation-run/reviews/identifiers.json` would attach the wrong provenance to this scope (P l.228). | In the next round and in the canonical `reviews/identifiers.json`, bind journal `906389d7212a401b8c7f781d37960d84.json` (with its SHA-256) and start record `c400e46e…`. |

## Suggestions

1. **Match SHAs regardless of kind at inventory.** Match each case's commit, its parent and both blob images against the union of all 40-hex values, whatever kind v4 assigned. 3,918 of the 6,149 Prometheus "blob" values also appear as commits. This adds no false matches. I did not raise it as an issue: oid-only SHAs from merged PRs are either already captured as a parent or merge commit, or lie off the first-parent chain.
2. **Quantify supply before collection scans.** The change-cost-v2 merge commits and the 23,462 paths with unknown scope may exclude most of the Django and Prometheus frames [INFERENCE].
3. **Publish a familiarity table by original repository** (the query in finding 3) as the P l.23 record. `counts.familiarity_sources` lists studies whose rows match any scope, not where familiarity originated.
4. **Re-check later activity at inventory.** Confirm the status of X11 (it targets the same three repositories), the unmerged inertia-bench branch `agent/handoff-machine-move` (`9c42859…`), and any newer Scout studies. Append any new exposure; it never becomes fresh again.
5. **Split the provenance-gap reason in counts** into "source resolved" (unattributed) and "source missing" (scoped). This repeats suggestion 5 from implementation round 1.

## Peers and where they changed the recommendation

- **experiment-peer.** Raised I-1 to I-5.
  - I-1 (the uncaptured change-cost-v2 fields) **changed my verdict** from acceptance with dispositions to issues open. I checked it against `collect.py` and `fetch_prs.py`, and it became ID-01.
  - I-2 (oid labelled as blob) became suggestion 1. A new per-schema delta is the place to re-emit `commits[].oid` as commit.
  - I-3 to I-5 became dispositions D-3, D-4 and D-1.
- **innovation-challenger.** ID-D-01 (whole-source unknowns) became disposition D-4. The documents and scripts cannot add test-repository case exposures beyond change-cost-v2's data. The challenger also found the low-yield `pr-*.jsonl.gz` files that led to ID-01.

## Verdict

**2 issues open** (ID-01 major, ID-02 minor). Inventory must not start. The coverage delta, the D1 files and the X11 directories are accepted without new authority. The unresolved provenance gap and the sources not parsed for identifiers are accepted under dispositions D-1 to D-4. Unresolved records stay excluded as unknown freshness.

Round 2 must review:
- the per-schema request and its delta producer for ID-01;
- the producer's appended outputs;
- the corrected journal binding for ID-02.

No sign-off is given in this round.
