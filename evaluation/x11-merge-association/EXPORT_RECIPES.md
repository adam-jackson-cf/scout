# X11a revision 8 preparation and identifier recipes

**DRAFT — preparation/inventory unauthorised.** [Round7](reviews/round-7.md)/[response](reviews/round-7-response.md). Identifier extraction, synthetic checks and Git-only eligibility reproduction grant no candidate source-scan/inventory/model authority. R1-01 full real implementation remains explicitly deferred.

## Authority and persistent inputs

1. A zero-issue **export-preparation** review freezes protocol/source/recipes/input/tool closures before offline candidate ownership or frontend build.
2. A separate **inventory** review binds generated completed/refused seals, artifacts and original preparation authority before inventory once.
3. Later protocol authority must adopt the exact original inventory authority/output seal. Full GUIDE/orchestration/calibration/plants/conditional OC review remains required before any real model request.

`stages.py` enforces exact schemas/digests; Main checks council provenance. No retroactive authority, empty exposure fallback or recipe hash in place of generated artifacts. Input schema remains `x11a-inventory-source-inputs-v4`; pin entries are exactly `{path,sha256}`.

Five ordinary full clones `/Users/adamjackson/Projects/work/x11-candidates-v6/{django,TypeScript,gin,cobra,prometheus}` preserve all Cold/C0/tip/resolved-tip pins. Earlier [restoration evidence](tool-pins-v6/object-restoration-v6.json) observed **eligibility segment lengths and object counts with an unreviewed producer**, not merely object connectivity; no diff/scanner/inventory ranking. Main's once-only reviewed [generator](envelope_check.py), [verbatim receipt](envelope-v7.json), [argv/environment record](envelope-v7-run.json) reproduce envelopes2791/1666/316/70/3248, capped2000/1666/316/70/2000. The1500–2000 amendment is informed by those distributions (R-084; round6§3). Fresh-clone Prometheus default HEAD is an unused separate observation from the original resolved remote tip, not an analysis-anchor change. Old clones/evidence untouched.

Read-only tool root: `/Users/adamjackson/Projects/work/x11-tools/83b941f8c8a994a6f02af7c08dfda99683279e0cc55ad456d3352335344b4c1e`. It contains relocated Git/executable helpers/templates/dylibs, Node and complete link-preserving Clawpatch dependency tree. `tool-pins-v6/` pins the whole immutable tree, OS build/architecture/sandbox/interpreters, Python whole install, and exact-selected-file frontend/OC Go closures. The trusted bootstrap checks the whole immutable tool/platform closure before entrypoint execution; scanner closure is rechecked before scanner commands. OS binary/build drift needs a council micro-round, not silent repinning. Git replacements/lazy fetch/system/global config are disabled; Node heap fixed8192MiB. Historical v4/v5 metadata remains unchanged.

## Stdlib-first isolated launcher

Use this launcher for every Python entrypoint. **Obtain INPUTS_SHA256 and BOOTSTRAP_SHA256 from the genuine reviewed authority/packet, not a live self-hash used as its own trust anchor.** For separately authorised synthetic checks they identify the exact source packet under test, not council approval. No helper is imported before its reviewed source digest is checked. Local `__pycache__`/`.pyc` is refused even with `-B`.

```sh
STUDY=/Users/adamjackson/Projects/work/scout-x11/evaluation/x11-merge-association
INPUTS="$STUDY/inventory-inputs.json"
PYTHON=/Users/adamjackson/Projects/work/x11-tools/python-4c17f7a00b8b57db0e61f0375d3c1345997a5f82739eb969b5ab82d8e7132c6e/bin/python3.14
PYDIR=/Users/adamjackson/Projects/work/x11-tools/python-4c17f7a00b8b57db0e61f0375d3c1345997a5f82739eb969b5ab82d8e7132c6e/bin
x11python() {
  : "${INPUTS_SHA256:?Set from reviewed input pin}"
  : "${BOOTSTRAP_SHA256:?Set from reviewed bootstrap source pin}"
  env -i HOME="$STUDY/artifacts/preparation-runtime/python-home" PATH="$PYDIR" LANG=C \
    TMPDIR="${FIXTURE_TMPDIR:-/tmp}" "$PYTHON" -I -B -c '
import hashlib,pathlib,runpy,sys
p=pathlib.Path(sys.argv[1])
if hashlib.sha256(p.read_bytes()).hexdigest()!=sys.argv[2]:
    raise RuntimeError("reviewed bootstrap source drift")
sys.argv=[str(p),"--config",sys.argv[3],"--config-sha256",sys.argv[4],"--script",sys.argv[5],"--",*sys.argv[6:]]
runpy.run_path(str(p),run_name="__main__")' \
    "$STUDY/bootstrap_v6.py" "$BOOTSTRAP_SHA256" "$INPUTS" "$INPUTS_SHA256" "$@"
}
```

Bootstrap checks the reviewed config, helper source **before import**, complete Python/tool/platform closure and target source before execution. The former helper-importing launcher/API is removed; no compatibility path. Integrity checks grant no execution permission.

## Offline ownership — candidate execution not authorised

```sh
AUTH=/ABS/export-preparation-authority.json
BUDGET="$STUDY/artifacts/preparation-budget-v7.json"
x11python "$STUDY/prepare_exports.py" --inputs "$INPUTS" \
  --authorization "$AUTH" start-budget --output "$BUDGET"

x11python "$STUDY/prepare_exports.py" --inputs "$INPUTS" \
  --authorization "$AUTH" --budget "$BUDGET" \
  ownership --repository django/django --scratch /ABS/scanner-scratch \
  --output "$STUDY/artifacts/preparation-v8/ownership--django__django.output.json" \
  --seal "$STUDY/artifacts/preparation-v8/ownership--django__django.seal.json"
```

Scan **Cold+C0+first L_r post-C0 positions**, L_r=min(2000,envelope), length eligibility≥1500; same L_r across ages/arms, never remaining envelope. One recursive `ls-tree -r -z`/`cat-file --batch` per revision, exact blobs/modes/safe symlinks/postverification; empty gitlink directories, no archive attributes. `.codegraph` excluded/count-reported, never read/modified. NFD/casefold aliases deterministic; unsuitable scratch operational.

Each revision gets fresh external state/HOME, sandboxed pinned `init` then `map --source heuristic`, no network/credentials/model CLIs/semantic state. Compact tree/batch/OID-list/count digests replace per-blob traces. Each revision's ownership is a disk record in a small digest index, not an all-revisions RAM union. The frontend reads base+revision ownership as NDJSON and preserves product filtering/normalised directories.

**Budget:** canonical create-once `artifacts/preparation-budget-v7.json` binds input/authority pins and one shared42day preparation clock. Every operation requires `--budget` at that exact path; new filenames cannot reset it. Up tofive repository operations via shared locked slots; per-attempt14days, every subprocess3600s/whole process-group kill. Ordinary operational failure retries at most3 fresh states, then `PreparationInfeasibleOperational`; concrete in-scope unsafe/unsupported content refuses immediately, structured revision/failing step bound to actual tree/batch trace. **Interrupt→`PreparationInterruptedOperational`, whole deadline→`PreparationBudgetExhausted`: both stop all preparation operationally**, never silent retry/selective candidate exclusion. Record host failure under Deviations; continuation requires review. Incomplete process-exit output never success.

## Frontend and exposure operations — not authorised

```sh
x11python "$STUDY/prepare_exports.py" --inputs "$INPUTS" \
  --authorization "$AUTH" --budget "$BUDGET" \
  build-frontend --output "$STUDY/artifacts/preparation-v8/build-frontend.output.json" \
  --seal "$STUDY/artifacts/preparation-v8/build-frontend.seal.json"
x11python "$STUDY/prepare_exports.py" --inputs "$INPUTS" \
  --authorization "$AUTH" --budget "$BUDGET" \
  exposures --output "$STUDY/artifacts/preparation-v8/exposures.output.json" \
  --seal "$STUDY/artifacts/preparation-v8/exposures.seal.json"
```

Active frontend/OC recipes are `tool-pins-v7/{frontend-build-recipe-v7,oc-build-recipe-v7}.json`, preserving v6 metadata. Same pinned Go/environment: `-pgo=off`, `GOTOOLCHAIN=local`, `GOFLAGS=-mod=readonly`, `GOPROXY=off`, `GOSUMDB=off`, `CGO_ENABLED=0`, `GOWORK=off`, `GOENV=off`, private HOME/GOCACHE/GOTMPDIR and exact current package/file closure. Atomic create-once executable/buildinfo promotion. Exposure verifies preserved extraction-v4 bindings, never reruns extraction.

## Seals and deterministic assembly

Create-once `x11a-export-preparation-seal-v7` binds authority/input/shared-budget/source/runtime/platform/heap, UTC/argv, attempts/trace, anchors/count and output/refusal. Successful ownership verifies per-revision tree→batch→materialised→init→map, OID/tree/batch digests/counts/argv. Content refusal requires real in-scope traced offending revision/step, not self-asserted class or preflight-only trace. Build verifies go-list/build/buildinfo. Digests are provenance, not arbitrary-JSON council authority.

Assembly discovers exactly the five ownership seals (including terminal refusals) and build/exposure seals at canonical paths; there is no caller-supplied `--seal` list:

```sh
x11python "$STUDY/assemble_preparation.py" --inputs "$INPUTS" \
  --authorization /ABS/export-preparation-authority.json \
  --output /ABS/new-artifacts.json
```

Assembler verifies operations/dependencies/traces and rejects missing/duplicate/nonterminal states; global interrupt/whole-budget codes immediately mean study-wide operational not-testable. Ordinary terminal ownership refusals retained; failed build/exposure operationally not-testable, no usable manifest fabricated. Inventory rechecks authority/refusals and uses unchanged gates except1500 length floor, no substitution after fixed winner. Calibration remains required; inventory cannot claim measured resource gates passed.

**Canonical attempt record:** output/seal paths are derived by `canonical_paths(inputs,operation,repository)` under `artifacts/preparation-v8/`, using `<operation>[--<owner>__<repo>].{output,seal}.json`. `execute` rejects every alternate path and any existing seal, including an empty reservation left by a failed host; changing filenames cannot reset≤three internal attempts. Starting work and assembling both check all canonical seals for interrupt/budget global stops. Assembly cannot omit a seal or use a selected subset. Host-failure continuation requires a council-reviewed deviation, not a fresh path.

## Conditional OC — still unauthorised on candidate inventory

OC wrapper requires `--inputs/--authorization/--inventory`, exact non-null vectors and pinned build recipe. Original inventory authority/seal/export binding includes `segment_lengths`1500..2000 and count/directory arrays exactly L_r for both ages. Official399 draws, no timing override; circular ceil(L_r/20) starts truncated to L_r. [Specification](oc-v5-spec.json) fixes56160/198 scenarios, seeds, one-way-robust paired bounds and IUT. Historical `v5` filenames are active revision7. Real10000 draws. Compute stop uses **minimum256 full-grid projection** versus384h, preserves first gate failure and separate compute reason;8h sensitivity. No two-way coverage/runtime/power guarantee.

## Historical extraction v4 — preserved

Seal SHA-256 `96fcfcbfada87faaa6624d9124fb3c88d55d714a8cd0bf79ce72a114e32f8510` records exact runtime/Git argv/source/output pins. Source Scout `04d5e632ae34d8056c3df5116b9ba87cfbf99282`, inertia `85e7329bef230358740c2b7ae90b6364a15b675c`. Do not rerun successful extraction or rewrite v3/v4.

Measured:50,016 scoped records; fastapi association8,866; inertia1,350 scoped records using `study.startswith("inertia-bench/")`. **36 inertia data files=27 screening JSON/JSONL/CSV+8 screening text/checksum+1 docs/design/candidates.csv.** Coverage24,001 parsed,321 source/docs, three unclassified non-data paths (`.gitattributes`, `.gitignore`, `docs/design/screening/raw-SHA256SUMS`), zero parse failures. Unscoped61,135 commits/177,483 blobs/23,462 paths; commit/blob matches compare both images except empty blob, paths never cross-match unknown repositories. Association-task overlap excludes scoring only; other-task and repository-only identities are descriptive per candidate/study. Preserve disclosure of other-task `human-confirmation-51.csv` mis-scoped to TypeScript.

## Authorised synthetic checks and resource evidence

After every source/pin lands, run the integrated entrypoint **once** with ordinary suitable synthetic scratch (path-alias policy is filesystem-independent):

```sh
FIXTURE_TMPDIR=/ABS/synthetic-scratch
x11python "$STUDY/selftest.py"
```

It includes preparation, bootstrap and OC contract fixtures; do not separately rerun those suites. They cover genuine synthetic build/exposure/ownership/refusal chains, trace tampering, interruption, aliases/gitlinks/codegraph counts, preimport sentinel/cache checks, exact inventory-vector authority, discriminating counts and gate thresholds. They are not deferred full experiment plants. Only synthetic repositories may use the real scanner offline.

Frontend tests: pinned `go test -pgo=off ./evaluation/x11-merge-association/frontend`, cwd `/Users/adamjackson/Projects/work/scout-x11`, recipe-v6 environment plus pinned Git directory appended to PATH. The ownership differential is isolated/bytecode-free. Do not run checks concurrently with sibling edits or the scale benchmark.

R-079: `prep_scale_v6.py` preserves the first80,000-file×20 TypeScript run: real init/map succeeded but zero owned paths, so not populated-output validation. `prep_scale_v6b.py` uses the frozen Go-package fixture (80×1000sourcefiles+go.mod), requires actual all-source ownership, and persists/readbacks the disk index and INV/FE NDJSON stream. Its seven-file positive construct produced three features/seven owned paths. Scripts accept `--tool-root`, `--scanner-tree`, `--scratch`, `--output`; v6b also has `--construct-only`. Evidence records elapsed/RSS/runtime/8192MiB heap, bytes/lines and fixed100k-file×2002 projections: whole elapsed including cleanup primary; processing-only separate. Identical small-file trees are not an arbitrary-complexity runtime guarantee. Original source/evidence remain unchanged; no candidate scan/inventory.

[Completed v6b receipt](prep-scale-v6b.json) has80,001ownedpaths/81features on every revision,2455.8070487079676s elapsed and188481536byte peak RSS. [Separate projection](prep-scale-projection-v6b.json) also discloses processing-only. The59,201,755byte stream exercise is Python emission/readback, not frontend binary execution.

Revision7 production-path runner `prep_scale_v7.py` uses genuine execute→seal verification and built frontend NDJSON consumption, changed-file and byte-heavy synthetic fixtures. Receipts pin original/snapshot source bytes and explicit location-only metadata fixtures; no mocked execute/authorise/verifier/frontend. [File-heavy](prep-scale-files-v7.json) and [byte-heavy](prep-scale-bytes-v7.json) completed; their original max(file ratio,byte ratio) projections **both fail** the unchanged14day/42day ceilings and remain unedited. The about1463day file-heavy result multiplies file-driven time by the byte ratio; it is not a measurement of candidate infeasibility.

**Capacity rule:** `preparation_capacity_v8.py` fits `seconds/revision=a·files+b·bytes+c` to the final-PREP construct/file-heavy/byte-heavy Go whole-elapsed/revision measurements, retaining v7 receipts unchanged. Three points/three unknowns give no residual/error estimate; fixed1.5× is the only safety factor, not a confidence bound. Reuse Main's once-only [Git-size receipt](preparation-capacity-v7.json), SHA256141c6764dee809dd0bb198d53e1a414c2f744b1403d9d99c1e282af530c8e037; collect no new candidate facts. Report additive projections with/without margin for Django and Go, disclosing Django's untested Python-mapper path. For TypeScript use the held-out direct bound: before Main runs once, create/hash a frozen prediction at the actual synthetic fixture's files/bytes. Fixture:≥81369files/380592869bytes per revision,50% tests across tests/ and sibling*.test.ts, nested≥three-level TS imports/exports/functions, several≥1MB code files, root package.json test script/tsconfig.json,3revisions/5%churn. Run final-PREP execute→seal verifier→builtfrontend; TypeScript per-attempt=1.5×1668×max(whole elapsed/3,frozen prediction), three attempts=3×that, largest revision=1.5×max. Compare unchanged14days/attempt,42days/three attempts and3600s largest revision,≤five concurrent operations. Any failed gate→**“preparation not feasible”**, no refit/rerun/fixture/repository/ceiling change. Report observed/predicted ratio and exact result in v8 receipts. These canonical rules become effective on the authorising zero-issue round; passing projections alone grant no execution authority.

**Recorded result:** a=.001485893272630103s/file,b=2.899084947714978e−9s/byte,c=4.392888931244594s/revision. [Frozen prediction](prep-scale-heldout-ts-prediction-v8.json)126.49289025945677s/revision; [Main's one held-out run](prep-scale-heldout-ts-v8.json)139.9957846949886s/revision (ratio1.1067482481255293). [Final capacity receipt](preparation-capacity-v8.json) reports all five pass, `projected-feasible-not-authorised`; TypeScript1.5× attempt350269.45330686145s,three attempts1050808.3599205844s,largest revision209.9936770424829s. [Clock/run provenance](prep-scale-heldout-ts-v8-run.json) explains419.9873540849658s charged elapsed (includes earlier setup/freeze) versus365.607s timed run wall; no model tuning. Actual40700source/0test/40702total owned paths each revision matches the frozen construct: test references are not guaranteed owned source; association discovery is capped at200tests by pinned mapper source.

Active Python is an exact read-only relocation of the previously reviewed2505files/nine relative links: [runtime v6b](tool-pins-v6/python-runtime-v6b.json), [immutable/platform metadata v6b](tool-pins-v6/persistent-locations-v6b.json). New ambient caches in the shared UV install caused a genuine whole-install refusal; they were neither deleted nor repinned. The immutable copy retains original reviewed bytes, and bootstrap checks its complete tree and modes. Historical v6 runtime/metadata and benchmark commands remain unchanged.

Version6 OC timing artifacts preserve earlier timings and distinguish measurements from projections. `cost-feasibility-v5.json` remains cost-only; `latency-basis-v4.json` remains fallback-only. PREREG§10 requires future180-session synthetic-GUIDE latency/throughput calibration before maps—none has run.
