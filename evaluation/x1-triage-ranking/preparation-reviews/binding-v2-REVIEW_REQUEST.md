# Binding v2 — experiment-rule-5 review request and prior-finding response

Date: 2026-11-09. **Proposed, unadopted, unexecuted.** Request a scoped innovation-council review of the exact files below, iterating until zero issues. Do not infer sign-off from this request or the passing source checks. The first sign-off authorizes only the scratch synthetic rehearsal; canonical acceptance needs the subsequent authentic rehearsal-report review described below. No rehearsal, acceptance, candidate scan, real scanner, model CLI, Infisical command or network fetch was executed by this implementation work.

## Authority and requested review

- [Amendment](../preparation/model-free-binding-amendment.md) and [v2 recipe](../preparation/model-free-binding-check-v2.json): four-layer deterministic proof replaces only v1's complete descendant-process-start/argv observation requirement. All other fixture, policy, parity, provenance and stage-refusal requirements remain.
- Authentic [X1 round 9](../reviews/round-9.md), `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc`: actions (a)–(e) only. The clock remains `2026-11-09T03:22:58.701909Z`. No frame/development or inventory authority is conferred.
- **R-064** delegated conservative decision; **R-087 (pending ledger row)** reserved by Main for the owner's managed-Mac/no-root challenge and seam audit. Main appends the row on adoption; the ledger is untouched here. This is not a product-design change or a new product hypothesis.
- Current response: [binding v2 round-3 response](binding-v2-round-3-response.md), addressing X1B2-R3-01 and trivial S1–S6. Earlier responses/checks remain historical. No sign-off is claimed.

Review `execution/run-config-amendments-proposed/001-binding-v2.json`, SHA `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121`. After zero-issue review, materialize `OUT/reviews/binding-v2-amendment.json` with `{scope, record, reviewed}`, scope `model-free-binding-v2-amendment`, and **exactly three reviewed rows: proposal, recipe and amendment**. No source, manifest or extra row is allowed. Compare those rows with actual `pin()` output before adoption: paths must be the exact absolute proposal `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/run-config-amendments-proposed/001-binding-v2.json`, recipe `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation/model-free-binding-check-v2.json` and amendment `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation/model-free-binding-amendment.md`, with their current SHA-256/byte values. A wrong path can be adopted by the digest-only adoption gate but then refuses irreversibly at binding. The council record still names every reviewed digest; protocol authority stays frozen from adoption through seal. Independently, `model_free/main.go` and every closure-pinned file freeze from the first `attest_closure.py` journal, **preflight included**; (b) must precede rehearsal. Only `execution/binding/**` and the manifest/runtime/implementation review are repairable as described below. A rehearsal frontend defect has no in-code recovery and needs a reviewed recovery of (b), outside this amendment; this is a residual risk.

## Changes from v1

1. **Effective API inputs.** The frontend validates actual Config/Options before registry construction/scanners and emits all API fields, complete registry specs/digest, original/frozen policy, empty lanes, disabled decisions/association, no ranking/history/PR-fetch seam and false egress flags. Catalogue construction is pure and invokes no scanner; resolve-only returns before state access/Assess. All required revision records resolve before the first scanner, and later effective bytes must match exactly.
2. **Actual runtime proof.** Retain `engine.Result.Lanes` verbatim, not a reconstructed report. Require exactly one contextual, Clawpatch-review, association and fallback row, reject unexpected model/decision-egress lanes, and inspect actual coverage/job rows. Association/fallback are disabled; Jev ranking is absent. `heuristic_only` also requires actual **`mapping.Source == "heuristic"`** (capital `Source`, the existing Go API JSON), not just a constant argv claim. The scanner-only reference reports structural absence instead of fabricated engine execution.
3. **Audited-seam guards.** Fresh private allowlisted HOME/XDG/TMP/PATH; no inherited credentials/overrides; key presence is refused and only its boolean recorded. Both failing stand-ins exist first on PATH. The actual `CLAWPATCH_CLAUDE_BIN` points to its stub. Codex has no binary override; its literal PATH command is guarded and `CLAWPATCH_CODEX_SANDBOX=read-only`. Removed the fictitious `CLAWPATCH_CODEX_BIN` setting. Attempt logs are exclusively precreated mode0600, read unconditionally and pinned even when empty; missing logs refuse.
4. **Genuine control.** Only synthetic unbound parent/Claude and head/Codex exercise enabled contextual model paths against actual failing stand-ins. Structured logs contain argv/PID/PPID/kind/fixed exit97. Production argument prefixes must match, and the exact same bound predicate must reject those real logs. The frontend's overall exit need not be the stub's exit; the control proves the stub's fixed failure and expected **interlock** rejection. No real CLI/version probe is allowed. Actual stub subprocess tests cover both providers, including Claude `-p` as script argv.
5. **No privileged observer.** Removed `binding/observer.py`, `observer_proof.py`, `descendant_fixture.py`, `gate.py`, and `tests/test_binding_observer.py`; removed process-tree, executable-allowlist, observer-proof/session and complete-start claims. Historical probe JSON moved byte-identically to `preparation-reviews/binding-v1-unprivileged-probe.json`; v1 review records are retained. Its successful allow-default nesting probe (exit0) is not misrepresented as failure. No outer experiment sandbox; the product analyzer sandbox remains unchanged.
6. **Clean contracts.** Acceptance kind `scout.x1.binding-acceptance.v2`; `deterministic_model_free_proof` replaces `complete_process_observation`. The other five acceptance keys are retained. Runtime kind is **`scout.x1.binding-runtime.v2`**, exactly `schema/git/python/frontend/clawpatch/executables`; removed unused `read_roots` baggage. Runtime/closure/source reviews remain mandatory.
7. **Scratch rehearsal.** `--rehearsal --fixture synthetic --output <SCRATCH/report.json>` invokes the same reviewed implementation, runtime, closure, routes and interlock but never opens/reserves OUT/binding, creates a Journal/OUT lock, or asserts acceptance status passed. The report kind is `scout.x1.binding-rehearsal.v2`, status `rehearsal`; six derived flags are `acceptance_predicate`, with full route results, keys/multiplicity, scanner-health reasons and pinned diagnostics/logs. Clock is validated read-only and pinned. Canonical synthetic requires `--rehearsal-report` plus authentic `OUT/reviews/binding-v2-rehearsal.json`, scope `model-free-binding-v2-rehearsal`, binding the actual report/current runtime/source pins, before lock/reservation. Symlink/traversal/OUT output paths refuse. A rehearsal is never a seal producer.

**Residual limitations:** sandboxed scanners are trusted through closure and the fail-closed analyzer network sandbox; their rebuilt PATH does **not** contain the controller stand-ins. Separately, the unsandboxed frontend/Node/Clawpatch tree and its git/go-list children have network available and are trusted through the non-transitive source audit, hash-pinned closure and configured stand-ins, not through that sandbox. Positive-control liveness is proved only at Scout's contextual seam. Clawpatch-provider exclusion rests on empty opt-in lanes, false egress and heuristic null-provider mapping; its stand-ins are statically checked defense in depth, not exercised controls. Unexpected descendants/absolute-path launches are not globally observed. No whole-program, malicious-software or OS/hardware-compromise guarantee is claimed.

## Round-1 and round-2 disposition

| Finding | V2 disposition and catching evidence |
|---|---|
| R1-01 | Applicable direct-launch/private-env portion retained. Actual direct subprocess argv, PATH/provider controls and no outer wrapper are tested (`EnvironmentAndSeamTests`). ES executable allowlist portion **superseded** by observation removal. |
| R1-02, R1-03 | **Superseded:** no ES ancestry/session/shebang proof, root command, sudoers or FDA gate remains. Their replacement is the four layers, not a bypassed privilege flag. |
| R1-04 / R2-01 Critical | **Fixed in source:** shared health accepts `AdapterIneligible` only with nonempty entirely `not_applicable` rows and no failures. Real Runner on a synthetic nonmatching file returns before sandbox/scanner; tests require healthy. Footprint/incomplete/failure variants refuse. Overall RunComplete is still not incorrectly required. |
| R1-05 residual Minor | **Source fix supplied, measurement still pending:** noncanonical synthetic rehearsal and no-canonical-write tests now exist. Main must measure the trigger set after scoped rehearsal authority; no measured findings are claimed here. Rule name corrected: append loop is **PERF402**, not PERF401; unchanged YAML and head-only complexity12 remain. Canonical acceptance cannot be consumed before reviewed rehearsal. |
| R1-06 | Preserved repo_path/content_sha256 manifests; synthetic fixture test calls verify_tree_pins and verifies deterministic materialization. |
| R1-07 | Preserved full occurrence identity, independently fixed expected key/hash, real bound-store provenance links, parent bijection and carried-ID disclosure; Go identity/link boundary tests retained. |
| R1-08 | Preserved planner ordering/batch positions, effective applicable/readable inputs, exact coverage/omission/gap comparison before findings, equal-length roots. Direct off-file observation ownership test added. |
| R1-09 | All revisions resolve before scanners and later bytes match. Actual Config/Options are asserted; pure catalogue identity is now included consistently in resolution/execution. Tests cover bad head resolution preventing every scanner. |
| R1-10 | Private head state directory, sorted carry and policy/symlink/conflict refusal retained and tested. |
| R1-11 | **Superseded:** no ES script:null parser exists. Genuine stand-in logging tests replace observer-shape tests. |
| R1-12 | Canonical create-once/fsync and ordinary-exception terminal refused receipts retained. CalledProcessError test retained; new rehearsal/adoption prerequisites refuse before irreversible reservation. |
| R1-13 | Applicable independent identity/health/policy/control/parity tests retained; new config/options, lane cardinality, actual map source, actual stub, missing-log, adopted-overlay and rehearsal-boundary tests added. Observer-only tests removed. |
| R2-02 Critical | **Superseded:** no interpreter/executable observation allowlist can misread Claude `-p`; it was deleted, not globally weakened. Direct real stub tests cover production prefixes for both providers; bound predicate rejects their logs. |
| R2-03 Minor | **Superseded:** no root-owned observer stop/fragment/signal path or finish-time gate remains. |

Nonblocking suggestions: S1 privileged check, S2/S3/S5 observer parsing/payload rules are superseded; S4's direct real launch-argv test is implemented (no new unrequested timeout policy); S6 PERF402 naming corrected; S7 documented without weakening complete gap comparison; S8's heuristic intent is proven from actual map output plus source gates rather than obsolete observed argv; S9 direct off-file test added.

## Exercised source checks

Fresh complete check history: [binding-v2-round-3-checks.json](binding-v2-round-3-checks.json), SHA-256 **`7272f8d2c7e0256bc86e44211bf89b9aebc3578f7f4627b03ae94bbe7e379e1e`**. Final full suite: **237 tests in 15.568s**, exit0, after correcting two new synthetic negative-fixture create-once errors (initial full output retained). Go vet and uncached `go test -count=1` both exited0. Core owner fixed canonical-role preservation; final Core is `b1759ec710fa4b71dec1ef91800e84d285c733693b0cb5f4ffd7fe68f147c916`. Binding assertions were retained, not weakened.

Cwd `/Users/adamjackson/Projects/work/scout-x1prep`:

```sh
/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B -m unittest discover -s evaluation/x1-triage-ranking/execution/tests -p 'test_*.py' -v
source .tools/activate && GOPROXY=off GOSUMDB=off GOFLAGS=-mod=readonly GOWORK=off GOENV=off CGO_ENABLED=0 go vet ./evaluation/x1-triage-ranking/execution/...
source .tools/activate && GOPROXY=off GOSUMDB=off GOFLAGS=-mod=readonly GOWORK=off GOENV=off CGO_ENABLED=0 go test -count=1 ./evaluation/x1-triage-ranking/execution/...
```

Go vet: exit0, stdout/stderr empty. Go test: exit0, stderr empty, complete stdout:

```text
ok  	scout/evaluation/x1-triage-ranking/execution/model_free	1.363s
```

The suite exercises real synthetic adopt → rehearse → verify_rehearsal → canonical synthetic main → Django main → collect_seal. Its source-repair case proves repair only for `execution/binding/**`, not the frozen Go frontend or closure-pinned files. New Go/Python shared-golden tests exercise the frontend field/predicate contract before (b). Authority, pins, clock, journals and seal logic stay real in lifecycle tests; only permitted scanner/non-binding prerequisites are substituted. No real preparation action or candidate scan ran. See round-3 response for complete outputs and the Core-owner role correction.

## Main-only prerequisites and commands — DO NOT RUN before staged sign-off

No root, sudoers drop-in, FDA change, ES proof or log-stream privilege is required. Main first needs a zero-issue review naming the recipe, amendment, proposal and binding implementation source pins. Run both shared golden contract tests on `main.go` `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078` **before the first `attest_closure.py` journal**. After clean review/tests, write `closure-model-free.json` and the exact-three authority receipt; run (b), then canonical adoption below. Preserve the proposal and final signed-off writer bytes/paths. Never perform (b) or adoption before final sign-off.

```sh
/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B /Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/preparation.py adopt --config /Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/run-config.json --amendment /Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/execution/run-config-amendments-proposed/001-binding-v2.json --review /Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/reviews/binding-v2-amendment.json
```

Only the command writes the adopted copy, timestamp, operation and adoption receipt. Keep the three-row authority and its record immutable. **After adoption**, generate the manifest including the adopted file and its scoped implementation/runtime review. If rehearsal exposes a defect in `execution/binding/**`, obtain a new zero-issue council round, regenerate the source manifest, `binding/runtime.json` and implementation review at their fixed paths, then run a fresh rehearsal at a **new unused `--output` and `.work` path** and replace the fixed-path rehearsal review, all **before any canonical reservation**. This does not permit changes to `model_free/main.go` or any closure-pinned file: those freeze from the first `attest_closure.py` journal, preflight included, and (b) must precede rehearsal. A frontend defect discovered then has no in-code recovery and requires a reviewed recovery of (b), outside this amendment. Runtime-only frontend defects are a residual risk even after golden tests. Canonical `reviewed_rehearsal` pins report, receipt and council record; original protocol/config/clock stay unchanged.

After the amendment is reviewed/adopted and those prerequisites are actually present, the first permitted component run is:

```sh
/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B evaluation/x1-triage-ranking/execution/binding/acceptance.py --config evaluation/x1-triage-ranking/execution/run-config.json --clock evaluation/x1-triage-ranking/preparation-run/clock.json --fixture synthetic --rehearsal --output /Users/adamjackson/Projects/work/x1-preparation-scratch/binding-v2-rehearsal/report.json
```

Main must review the real report/trigger set and materialize the authentic `model-free-binding-v2-rehearsal` receipt, then (not before):

```sh
/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B evaluation/x1-triage-ranking/execution/binding/acceptance.py --config evaluation/x1-triage-ranking/execution/run-config.json --clock evaluation/x1-triage-ranking/preparation-run/clock.json --fixture synthetic --rehearsal-report /Users/adamjackson/Projects/work/x1-preparation-scratch/binding-v2-rehearsal/report.json --output /Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/binding/synthetic.json
/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B evaluation/x1-triage-ranking/execution/binding/acceptance.py --config evaluation/x1-triage-ranking/execution/run-config.json --clock evaluation/x1-triage-ranking/preparation-run/clock.json --fixture django-pinned-pair --objects /Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/objects/django__django.json --output /Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/preparation-run/binding/django.json
```

The final command additionally requires the passing canonical synthetic receipt and the genuine authorized Django object receipt with sealed first-parent SHA. These commands are instructions only, not exercised output. Runtime/closure/sign-off and the trigger measurement remain prerequisites, not claims of completion or pass. All files are uncommitted.

## Review SHA-256 manifest

Paths in the first two tables are relative to `evaluation/x1-triage-ranking/`. This request's own hash is supplied in the handoff, avoiding self-reference.

### Binding-owned current files and evidence

| File | SHA-256 |
|---|---|
| `execution/binding/acceptance.py` | `2144999a8aece522e72a86a429a1628a2833ada1490fa6e0fe7600e152373be1` |
| `execution/binding/fixtures/common/go.mod` | `ec75b6e87719a439d7ec4959302e25fd588319c7f7cfd1e6e20b39a324aff0b0` |
| `execution/binding/fixtures/common/index.js` | `2fa9c2c5a47269dd22d1da68d24423b1a2d3a2da635c4c60798c59e17ade5d23` |
| `execution/binding/fixtures/common/main.go` | `d37528f6027aa09adca50fa54e40ffc247c5597ffb7d4effa633ccba6134536c` |
| `execution/binding/fixtures/common/package.json` | `514ef8299051029ada3268f1fb2eba8b698548d59bb6d438d71c42c03a627f9e` |
| `execution/binding/fixtures/common/unchanged.py` | `534f740e38f9a8c5e7226a676a427bc5b44ec64ab605c33c595a68aedf5a97c4` |
| `execution/binding/fixtures/head/.scout/policy.yaml` | `6189fea6ad039f0eb0206a7986aa94a44e3bdf46669184298b8a558d06fa7255` |
| `execution/binding/fixtures/head/changed.py` | `0950bd4989dfc76b1c8ed8641c91b4994cefee6bc5bb6d5105510f99f25ec109` |
| `execution/binding/fixtures/parent/.scout/policy.yaml` | `3d3dae7d063c4fe94127de4e67ae9315e1495e4fa510ea587ae4c5a058604ec3` |
| `execution/binding/fixtures/parent/changed.py` | `0ecf4e2fd0b76f8305f0729807a67bc7d2eeabe27ed637a98bda8b9d87738e3b` |
| `execution/binding/fixtures.py` | `48552a2bd5f8492b83237f147916bde12f2137a6aee6fbae99b01cda9e40381b` |
| `execution/binding/interlock.py` | `19464e8940d0458be3606ed0438ae778fd01b48b73f8c2fee801856b80b2e6a1` |
| `execution/model_free/main.go` | `edb37ec2a2b8efb16c6f38bb4c5ec10a299a6e4ce9b7fcd0424bfa2845265078` |
| `execution/model_free/main_test.go` | `54e7bc66d4df750796ae4201b1dbeba172373a34dc46c3d4cce629d17e54d617` |
| `execution/tests/test_binding.py` | `2458e4289d67a1a061fcf43a2de6e42f1d586daee6cb4452f17f857ee1d36ab5` |
| `execution/tests/test_binding_adoption.py` | `4b186aa1c672bd7353d014e056ae881cd0e5a05613130ff44a4791450ce0065d` |
| `execution/tests/test_binding_boundaries.py` | `c5c1d042ee440e39eab66bb3bdb0d0d13b4756d4fe0c1018a16d7e63b7c2b177` |
| `execution/tests/test_binding_seal.py` | `b21763f3a94b229722381a36ed0b2a2a6054bdd6c476e326445bdf76aad38450` |
| `execution/tests/test_binding_contract.py` | `dc817736d39551e01c62acb2e6b3dffeb861de825b4fb56c9b5c3198dc0a9efa` |
| `execution/tests/model_free_contract.json` | `bc88c34ae1f8edc800e2c6a6d7eb86af4fcda205899deaba085630edab0a334d` |
| `preparation/model-free-binding-amendment.md` | `27f1100122f9867974049a88dfae14587aa6a9db7d96ae8065faca45fca284f7` |
| `preparation/model-free-binding-check-v2.json` | `9c1340da248ca25d99053f81309ff4c5d13f433c2faf2f22e055d2197c5020dc` |
| `preparation-reviews/binding-v1-unprivileged-probe.json` | `10a0c930a3b3d006e7a9c368ea7a93f709ebfda13a149b7521dc1c5721ef4881` |
| `preparation-reviews/binding-v2-checks.json` | `22a5d5bff095260f39a710fdd2f5b8161293f301a688cac0cf505377e0817ea6` |
| `preparation-reviews/binding-v2-round-1-checks.json` | `8d567ea225c388c19bfb6e82c7e93076f2f5c27e1f26572cf8b2ff3a24ce10e4` |
| `preparation-reviews/binding-v2-round-2-checks.json` | `07dc03e1c55c4a93b246d6ef60f7c5d3879b1b7321437f0df61a38d446d92c55` |
| `preparation-reviews/binding-v2-round-3-checks.json` | `7272f8d2c7e0256bc86e44211bf89b9aebc3578f7f4627b03ae94bbe7e379e1e` |

### Shared contract, immutable authority and prior review (read; not edited by Binding)

| File | SHA-256 |
|---|---|
| `execution/preparation.py` | `b1759ec710fa4b71dec1ef91800e84d285c733693b0cb5f4ffd7fe68f147c916` |
| `execution/run-config.json` | `c817f6789f1cd112242664c33ddd9bb4193510869c14027bb269ce7a23feee54` |
| `execution/run-config-amendments-proposed/001-binding-v2.json` | `88eb95a55c65d415a3ca4ef50b43f2513231c72fe5008820631d07527575f121` |
| `execution/tests/test_core.py` | `4f4bfdaf50fc0d56ff775f8c86bb6bc579c932acc1a76764a55ba4a6f36da07e` |
| `execution/verify_inputs.py` | `7d6f89ce90986c44f3cad2128344a7c90bf4670301ba5811e59886babd0d9310` |
| `preparation/model-free-binding-check.json` | `5056ae748e05717e7a1e16dd8974f11519045c01bdc3ce6498adf3b14aedc59c` |
| `preparation/model-free-policy.json` | `87de394937a225f87f1d53fed00c09bff88372a5de17b6669656478a32245f9f` |
| `reviews/round-9.md` | `6f1d6ca271055ce4953009a2e229d4fb2bd99c84901918ae524760982533c6cc` |
| `preparation-reviews/binding-impl-round-2.md` | `fcacab9a6be95e4ce70d01c61eca09f8313c64a2af8050a0d6a5c4b978a9190b` |
| `preparation-reviews/binding-v2-round-1.md` | `91105b4598f4d971ee78933a6578b5560476b966b27260109adc15fd4b0eabcd` |
| `preparation-reviews/identifiers-impl-round-7.md` | `6d279752335d4982754fd5639f729ecc008a23eacd66fa1207268caf1706b01f` |
| `preparation-reviews/binding-v2-round-2.md` | `29c1651f01bc5c87543f182de84c2008307e4e17dcb69ec9b3f60e4890327dc0` |
| `preparation-reviews/closure-impl-round-4.md` | `5e9fcd1bcd1a28ae3487505df11d1bb95c378d6c2d1a33a059b85a9ffd349627` |
| `preparation-reviews/binding-v2-round-3.md` | `d3fcde7cfec9a9f3e9126b3284903d0be6796f0f414e947fcf71d27f1354985b` |

### Audited product-source seams (repository-relative; unchanged)

| File | SHA-256 |
|---|---|
| `internal/engine/assess.go` | `dbb20a1eccea29f7f46b88337fbdea73a1ee2c4f54e64174f9d03916e59fd1d1` |
| `internal/engine/model_cli_policy.go` | `cb7694f7f44af75d9e71370431de568c1e460aef7a9d8f73efb43c9359ead734` |
| `internal/association/lane.go` | `8f54f8d361f65b74b061df79d96cd99e5cd728e7faf000a2eb68bd5d1946f334` |
| `internal/decisions/client.go` | `0f46173a3d787e1577d97699f64623ebddedb62c3bf31b9428ac5c23538c2fa0` |
| `internal/decisions/run.go` | `bc8d5d38c5281446d6e9ed54c151a6e4fcddb6c9fad3a5b30ac9cf5a113d849c` |
| `internal/engine/ranking.go` | `ca49c3c088df15e5c3e4d2227099e8c1030e7cc909514c7e1dfa196f7cd70380` |
| `internal/clawpatch/client.go` | `a53ca6056b7de521d1dc4c262e593aff00a744ef331440136179e7315b74f180` |
| `internal/clawpatch/review.go` | `1a0bf3825d0228c8372c81c3d79b65edd86b2633b83acf346f2ef558df1ac117` |
| `third_party/clawpatch/src/app.ts` | `c6336d07cf2c763cc677523cc26f98fed90fe3dc926d927aa95a174ad34ad575` |
| `third_party/clawpatch/src/detect.ts` | `9bcd0a495a4f794661c9aff3600d2825ae730d7a80cca24683430fd76e477e46` |
| `third_party/clawpatch/src/git.ts` | `dfc98e3764b1cafa1c7f5a46490009ffef10e0e6efb3ceadcd0431841108efcb` |
| `internal/modelcli/modelcli.go` | `b02b4692a8a6e44db48fc07c3cab4e09b843cb8fb6db8850d7d5838f9e7f9096` |
| `third_party/clawpatch/src/providers/claude.ts` | `fadf1e447d9b0b9a32ba7db6ca9101d29a84acce5c575f37423995456d03d059` |
| `third_party/clawpatch/src/providers/codex.ts` | `994e8066be12c3051c4cb63b32e2144bcf4bd495c5d17a8033f94814bd7a6e92` |
| `internal/sandbox/sandbox.go` | `18886a0b81e57c1e1eb992ae43844750a0538777b1f7e26324a3120a1cc25c73` |
| `internal/sandbox/run.go` | `d78faa34ec691361900d986ff4aeb4cba11f88e7815bbb8cc481393ecba82afd` |
| `internal/analyzers/runner.go` | `8ec374d7b1e67fa04999d5ce8934c5428285ab3b1eeaa4d109342990b6e979f3` |
| `internal/measures/tools.go` | `966e94868d45bae749b8fed74a237cf8a1348b21db6512ca19e74e786e2864f1` |
| `internal/prmeta/fetch.go` | `2cac2d7813b5a602180cc247096824a93527f042216e19de6a03a9a59e9d24cd` |
| `internal/baseline/baseline.go` | `bdc94bb7876d03203110bcff82b4780f204e53abd2a0750ccbdb2e9ffef9d1b7` |
| `internal/inventory/plan.go` | `becf04b94d3b175de6f62fb67fdb63dd3c09f6b526d616978cb1e481ce1a278a` |
| `internal/engine/lanes.go` | `7bcddebbfea020b8e67a1d1b17e2557545f38f03bd6af1619213657925c98ca9` |
