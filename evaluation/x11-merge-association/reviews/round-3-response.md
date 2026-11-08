# Round 3 response — X11a revision 4

2026-11-08. **Drafter response, not council closure/sign-off.** Alternative A: address all preparation-stage findings before requesting export-preparation review. Neither preparation nor inventory is authorised. No ownership scan, inventory, real map/label/replay/model call or commit. Owner-directed extraction v4 ran create-once; v3 and previous reviews/manifests are preserved.

## Fixed design choices before any inventory facts

- **R3-04 option(a):** recall unknown cap `U_all/(E_known+U_all)≤.05`; precision unknown-assignment cap≤.05. Conservative estimators unchanged. Residual missing share reported, not used as recall cap. Removed invalid precision-only recall rationale.
- **R3-05 alternative D:** full-census replay, lowest-hash **2600 labels per repository×age×arm cell**, independent of outputs, no replacement for unknowns. Targets≥800 known assignments/existing per age unchanged. Unlabelled-unsampled cases are not missing. Concern/propagation semantic summaries explicitly sample-bound.
- Exact eligibility projection removes heuristic-modified supply: all-added Cold90%/C080%; rename-without-modified Cold75%/C050%; heuristic-modified0%. Fixed hash/salts, same across arms/roles; no fitting to mapped supply. All vectors/composition and thinned versus unthinned resource counts are emitted.
- Complete coherent OC generator fixed: none-share grid{.2,.4,.6} across all repositories, assignment prevalence derived from P/R/q, explicit infeasible-margin companion degradation, age/arm copula, beta ICC, missingness race, raw-label/gate generation, model DEFF function, all binding scenarios/seed/order. No filtering to favourable achieved DEFF. Every binding scenario must pass before maps; no numeric power claim now.
- Labels reserve149,769 sessions regardless of projected yield. The **draft model-hours ceiling changes1500→3000**, before inventory; other ceilings unchanged (150k sessions,40k fallbacks,USD200,6 maps,21days,6 parallel). This supports the fixed2600 cap plus documented latency allowances. Nominal proxy band widens from old total≈7758 ceiling to9000 by Jev/3000 per repository, not guaranteed feasible actual supply.
- Latency-only evidence: disjoint pinned X2d fallback logs n25/max5.9s and n32/max10.3s. Fallback31s=ceil(3×10.3); label60s is an explicit cross-task margin, **not measured label latency**. Rate limits/queueing/timeouts can stop before ceilings; no deadline/budget/vendor escalation.

## Per-issue response

| Issue | Revision 4 change | Evidence / remaining boundary |
| --- | --- | --- |
| R1-01 carried | Intentionally remains deferred: full GUIDE/driver/replay/label/group/scorer/seal and planted-effect/null/OC implementation are not claimed implemented. | All later stages refuse. Final implementation sign-off remains required before maps/labels. |
| R3-01 major | Evidence-backed contexts restore pilot fastapi and postrun fastapi/axum, nested overrides preserved. New create-once v4 extraction; inventory consumes every unscoped full commit/blob SHA, except empty blob. Unscoped paths never cross-match. Per-candidate/study resolved/unresolved counts, kind summaries and known-positive X2d SHA exclusion fixture. | **Fastapi association records8,866**, scoped total50,016. Unscoped commits61,135/blobs177,483/paths23,462; not silently ignored. V3 untouched, superseded for binding. |
| R3-02 major | All inertia paths use the same data suffix enumeration as Scout; screening data extracted task other. Replaced false zero-study/data statement. | **36 inertia data files,1,350 scoped records; screening per-file identifier total4,714.** Exact pinned listing has **27**, not council asserted28, JSON/JSONL/CSV files, plus8 parsed text/checksum files. This measured discrepancy is recorded, no invented28th file. |
| R3-03 major | PREREG §2/§9 fixes every generator input: proxy vector/population/thinning, q grid, arm prevalence, dependence, missingness/failure/gate generation and model DEFF. Binding scenarios cannot condition on achieved DEFF. | Numeric table remains unmeasured until authorised inventory/full implementation; any invalid/failing binding scenario stops before maps. |
| R3-04 major | Recall denominator cap chosen explicitly, precision cap separate; arithmetic counterexample acknowledged, none-share grid binding, conservative estimators retained. | No claim residual5% missingness guarantees recall≥.80; cells exceeding cap inadequate. |
| R3-05 major | Population kind composition and deterministic projection vectors; full-census replay/hash-label cap2600; fixed enlarged nominal band and resource limits; latency-only source evidence and subscription stop. | `latency-basis-v4.json` records source digests/n/means/quantiles and assumptions. Fixed sample targets unchanged, actual mapped yield may still fail. |
| R3-06 major | Rule3 has only per-age adequacy gates; accuracy point/bound thresholds apply to equal-age pooled repository estimator only. | Synthetic fixture .78/.92 with pooled lower≥.80 supports when adequate. |
| R3-07 major | Whole resolved scanner-install tree pin; PREP verifies it **before each scanner command**, not entrypoint only. Manifest and source generator bound in preparation source closure. | **2667 files/109 symlinks** including dist/dependencies/package/lock. No scanner run. |
| R3-08 minor | Preregister primary gold-none gap+new shares and selected-minus-each-mandatory block-bootstrap differences; exact failure wording and no re-selection. Omit cluster-level U×gold-none analysis. | “high-gold-none condition not realised” unless both pooled lower difference bounds>0; one-candidate margin limitation explicit. |
| R3-09 minor | `Run(...,nil)` Decided/Capped/Counts parity; product ownership-loader/IsAgent differential fixture including absent paths; exporter accepts stale absent paths; harness emits product-owned normalised directories, removes Python NormDir port. | Synthetic frontend tests passed. Product files unchanged. |
| R3-10 minor | Role code/text aligned precisely; repository asymmetries and Python-only basename patterns listed. | Fixtures include test_helpers.go/non-test, Django test/, TypeScript harness/testRunner and Prometheus testutil. |
| R3-11 minor | Git/frontend neutralise global/system config; independent per-merge rename-skip warning check reports count/affected merges. Future replay same isolation mandated. | Synthetic warning fixture; no real rename counts computed. |
| R3-12 minor | Per-commit tracked-path export-ignore/export-subst checks; require zero non-unspecified attributes before archive, record counts. | No attribute-altered archive silently scanned; no real archive/ownership scan run. |
| R3-13 minor | Refused export fails repository gates: mandatory→not testable, candidate→ineligible; changed recipe requires renewed review. | Absent owned paths no longer refused (product-compatible), other extraction refusals retained. |
| R3-14 minor | Preparation/inventory bytes freeze under original authority explicitly; later GUIDE/driver/OC/final pins are permitted completions only under fixed rules and enumerated inventory inputs. | Changed signed bytes need fresh preparation review, not retroactive digest adoption. |
| R3-15 minor | Pinned Python/Git; v4 runtime/version/argv/context seal; exact hermetic frontend build recipe; go-list dependency/cache/embed/compiler/include closure; future literal go-version-m buildinfo mandatory artifact field. | **251 packages/1463 closure files**, including terra_prompt.txt; input recipe_sources bind1478 source/metadata entries. No prepared executable or buildinfo fabricated/built. |
| R3-16 minor | Exact source/artifact/auth envelopes, FE DisallowUnknownFields; unmocked chain fixtures for omitted files/artifacts, stale source/preparation digest, wrong scope and injected label/model/verdict fields. | Pinned-Python synthetic suite **10 tests OK**; frontend synthetic tests **passed**. |
| R3-17 minor | PREREG/PLAN say REQ176/177 decision waits for X11b or signed amending ADR; advisory guard computed/recorded but recommendations ignored; X11b replay/label reuse declared exposure requiring fresh confirmation. | No adopted design/default change; X11b remains plan only, HD8 proposed. |

## Extraction v4 evidence and qualification

Pinned sources remain Scout `04d5e632ae34d8056c3df5116b9ba87cfbf99282`, inertia `85e7329bef230358740c2b7ae90b6364a15b675c`. Exact extraction and git argv, pinned Python/Git versions/digests, context evidence and all output digests are in `exposure-extraction-v4/seal.json` (schema-v2), pinned by inputs/source index. The earlier v3 artifacts were **not modified**.

Coverage: **24,001 parsed identifier projections,321 non-data sources/documents,3 explicitly unclassified non-data paths**, zero parse failures. The three are inertia `.gitattributes`, `.gitignore`, `docs/design/screening/raw-SHA256SUMS`; never falsely counted as parsed or silently omitted. Context documents are not opened except the explicitly pinned preregistration evidence used by the context mapper. Extraction computes no predictor–outcome relation and reads no candidate checkout/outcome. Snapshot coverage is not a proof that arbitrary scalar prose contains no other exposures.

Metadata generation only pinned scanner install and Go build dependency closure. No preparation authority or real generated ownership artifact exists. The next requested review is **export-preparation**, not inventory; after zero issues, ownership/build artifacts require their separate inventory-binding review.

## Exercised checks

After all code slices and input integration:

```sh
source /Users/adamjackson/Projects/work/scout-x11/.tools/activate
PYTHONDONTWRITEBYTECODE=1 /Users/adamjackson/.local/share/uv/python/cpython-3.14.10-macos-aarch64-none/bin/python3.14 /Users/adamjackson/Projects/work/scout-x11/evaluation/x11-merge-association/selftest.py
GOTOOLCHAIN=local GOFLAGS=-mod=readonly GOPROXY=off GOSUMDB=off CGO_ENABLED=0 go test ./evaluation/x11-merge-association/frontend
```

Observed: **10 Python tests passed**, Go frontend test package **passed**. No full plants/OC/replay validation claimed. Markdown formatted; final exact digests are in `rev4-digests.json`, self excluded. No commit.
