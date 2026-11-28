# X1 execution-implementation council — round 5 (rehearsal-2 failure: repository identity at bound/head)

**Study:** `/Users/adamjackson/Projects/work/scout-x1prep/evaluation/x1-triage-ranking/` (STUDY). OUT = `STUDY/preparation-run`. SCRATCH = `/Users/adamjackson/Projects/work/x1-preparation-scratch`. W = `SCRATCH/binding-v2-rehearsal-2/report.json.work`. R = `/Users/adamjackson/Projects/work/scout-x1` (the frozen, attested Scout runtime).
**Date:** 2026-11-10.
**Action:** preparation action (c), binding v2.

Fresh rehearsal 2 ran under round-4 step 7 and failed with `InterlockFailure: frontend/scanner exited nonzero`. This round finds the cause, chooses the remedy under `preparation/model-free-binding-amendment.md:74`, and states the sequence. R-064 is in force: the council recommends and does not pose a choice.

## Inputs

The council did not recompute any SHA-256. Values marked "…" are the prefixes Main supplied.

| Input | SHA-256 |
|---|---|
| Round-4 record `preparation-reviews/execution-implementation-round-4.md` | as written by Main |
| Implementation receipt `OUT/reviews/implementation.json` (written and validated before launch) | `ac3ccf37…` |
| `RESULTS.md` (D updated) | `e7c39081…` |
| Rehearsal-2 report `SCRATCH/binding-v2-rehearsal-2/report.json` | `1d8442d3…` |
| Run record `preparation-reviews/binding-v2-rehearsal-2-run-2026-11-10.json` | `8cca165a…` |
| Session record `preparation-reviews/binding-v2-rehearsal-2-session-2026-11-10.json` | `2013e0fa…` |
| `W/logs/bound/head/result.json` | `e81fa7ae…` |
| `compare_enumeration` stdout after rehearsal 2 (32 roots, 63,459 records, zero differences) | `3191f7bc…` |
| `execution/binding/fixtures.py` (signed, round 4) | `48552a2bd5f8492b83237f147916bde12f2137a6aee6fbae99b01cda9e40381b` (6956 B) |
| `execution/tests/test_binding.py` (signed, round 4) | `5f008766807c973084f9b760cccc74857a43fcef7924e89061dea89af360f9ee` (33253 B) |
| Execution source manifest (round 4) | `50e8c2b16907513f10cefd74232e93166b2363ca1375e37e080c85a17058a4ce` |

**Method:** three independent read-only analyses: Main council (Opus), `innovation-challenger` (GPT Astra) and `experiment-peer` (Opus). No launch, build, test or file mutation was done.

## Findings

### F1. Cause, confirmed: Scout keys state to the origin remote or else the root path, and the harness carries state between two different roots

**How Scout derives the identity:**
- `R/internal/inventory/scan.go:621–627` `repositoryID`:
  - The identity is `git config --get remote.origin.url` if that is set and non-empty.
  - Otherwise it is `filepath.Clean(root)`, where the root has already been resolved through symlinks (`scan.go:78`).
  - The result is `"repo-" + digestStrings([identity])`, a hash of that one string. It does not depend on HEAD, the commits or the object store.
- `scan.go:171` puts the ID in the snapshot.
- `R/internal/state/store.go:348–363` `ensureRepositoryIdentity`:
  - It stores the ID in `metadata.repository_id` the first time.
  - On any later mismatch it refuses with `ErrRepositoryIdentityMismatch`. The comment at `store.go:30–31` gives the purpose: "prevents state from one target being reused for another".
  - It is called from `SavePlan` (`state/issues.go:25`), and at run start and finalisation (`state/progress.go:88,252`).
- `main.go` does not derive or override the ID (`:406–410` only opens the store and lists issues).

**What the harness does:**
- `fixtures.py` `materialize` creates each root with a fresh `git init` and never sets a remote.
- No global or system config can supply one, because both `fixtures.git()` and `private_env` set `GIT_CONFIG_GLOBAL=/dev/null` and `GIT_CONFIG_NOSYSTEM=1`.
- `execute_fixture` materialises bound/parent at `W/boundroute/parent` and bound/head at `W/boundroute/head`.
- `carry_state` (`acceptance.py:144–158`, called from `execute_fixture`) copies the parent's `.scout`, except `policy.yaml`, into head.
- Both `.git/config` files are identical and have no `[remote]` section (read by Main council and both peers).

**The deciding observation (experiment-peer).**
- `W/positiveXX/parent` has the same commit, the same objects and a byte-identical `.git/config` as `W/boundroute/parent`; only the path differs.
- Its stored ID is `repo-eabe806…`, while `boundroute/parent` stored `repo-bfcf5cac…`.
- Same content at a different path gives a different ID.
- The head store's `metadata.repository_id` equals the parent's `repo-bfcf5cac…`, which is the "stored" value in the error. The value received at head is `repo-46593fad…`.
- `mapping.Status` is empty in the head `result.json`. So the run passed `state.Open` and `ListIssues` and was refused at `SavePlan`, before mapping or any scanner. "Runtime proof incomplete" is a consequence of that refusal, not a second fault.

**Hypotheses refuted:**
- A config or remote difference: none exists.
- An identity derived from commits or objects: `repositoryID` reads neither.
- State corrupted by the copy: the head store opened and listed its issues normally.

### F2. The defect is in `binding/**`; the frozen frontend and engine behave as designed

The engine's refusal is the product's intended protection. The fault is the harness: it materialises parent and head as two separate repositories whose identity falls back to their paths, then hands one repository's store to the other.

Consequences:
- No change to `main.go`, the engine or the closure is needed. The repair route is amendment `:74`.
- Any remedy that changed `main.go` or the engine would need a reviewed recovery outside `:74`. None is proposed.

### F3. The same defect would have hit Django at its canonical run

- Django uses the same `materialize`. Its object database has an origin remote (`fetch_objects.py:65`), but that is not copied into the materialised roots.
- Django is never rehearsed. If this were not fixed now, it would surface at canonical Django, where the create-once receipt allows no repair.

### F4. Why the reviews and tests missed it

- `test_binding.py:236–246` carries a fake `state.sqlite` with the content `b'synthetic state'`, never a real Scout store.
- The `execute_fixture` tests replace `materialize`, `carry_state` and `run_revision` with fakes (`test_binding.py:454–456`; `test_binding_boundaries.py:295–296,321–324`).
- `test_binding_seal.py:86–94` substitutes `generate`, `materialize` and `run_revision`.
- The Go contract fixture runs with no assessment and no real checkout (`main_test.go:414–415`).
- No review or preparation document mentions Scout's repository ID; both peers grepped `preparation-reviews/` and `preparation/`.
- Rehearsal 1 stopped earlier, at the closure drift. Rehearsal 2 was the first run ever to reach a carried head.

### F5. Integrity

- Rehearsal 2 refused correctly. It wrote only under SCRATCH, reserved nothing and produced no acceptance.
- Round 3's C1–C3 worked:
  - both repaired stubs ran: positive/parent made 80 claude attempts and positive/head 80 codex attempts;
  - bound/parent was healthy;
  - the post-launch closure check ran after the bound/head launch and passed, because it executes before the nonzero-exit check;
  - `compare_enumeration` afterwards found zero differences.
- No P3 breach occurred, so no `RESULTS.md` deviation is required. Amendment `:74` anticipates binding defects found in rehearsal.
- The round-4 step-8 post-checks were rightly not run, because the procedure stops at the first failure.

### F6. Constraints on the remedy

- **The recipe requires state carry.** `preparation/model-free-binding-check-v2.json` `required_checks_per_fixture[4]` says "Preserve independent fresh parent-to-head state lifecycles and equal-byte-length bound/reference roots" and requires that carried durable IDs be disclosed.
- **One shared root per route is unavailable.** `main.go:364` puts `Config.Root` into the effective record. Every effective record is resolved before any scanner runs, and is compared byte for byte at execution (`verify_effective_record`, `acceptance.py:96–100`). Per-root source manifests are also pinned.

## Issue

### X1EI-R5-01 — Major — the harness carries path-keyed Scout state between two separately materialised roots, so bound/head always refuses (synthetic and Django)

**Evidence:** F1–F4.

**Required fix:** C4, T5 and T6 below.

## Recommendation

### C4. Give every materialisation of one fixture the same origin identity (`binding/fixtures.py`, under `:74`)

In `materialize`, immediately after the existing `git(executable, destination, 'init', '--template=', '--initial-branch=x1')` line, insert exactly:

```python
    # Scout keys durable state to remote.origin.url, else to the root path
    # (inventory.repositoryID). Name the source database so parent and head
    # roots of one fixture share an identity; nothing ever fetches from it.
    git(executable, destination, 'config', 'remote.origin.url', str(Path(database).resolve()))
```

No other line in any file under `execution/` changes, and no file is added there.

**Properties:**
- **Used by every route.** It is set at materialisation, before any resolve-only call, for every route (bound, reference, positive control) and both revisions. Parity inputs stay symmetric and the equal-length root paths are unchanged.
- **Distinct per fixture, and true provenance.** The value is the resolved object database the root was materialised from: synthetic `SCRATCH/…/objects.git` and Django `database_realpath`.
  - Different fixtures get different identities; there is no universal dummy value.
  - It is a local path. Nothing in the engine or the frontend fetches from origin: there is no baseline or `ls-remote` call on these paths, and the PR-metadata fetch is structurally absent (`main.go:367`).
- **Bytes unchanged.** Only `.git/config` changes. The materialised file bytes, the manifests, `verify_materialization`, the ignore-rule digest (it reads `info/exclude` and `core.excludesFile`, not the remote) and the worktree digest are all unchanged.
- **Disclosed side effect.** Clawpatch reads the origin for `projectId` (`dist/detect.js:24`) and for project-name fallback (`dist/git.js:46–51`). The project name still comes from `package.json` when present.
  - `projectId` now depends on the database path rather than the root path, identically on every route.
  - Bound and reference therefore become more alike, not less.
- **Product-faithful.** The product prefers the origin over the path for identity. Real checkouts have an origin.

### T5 and T6. Tests (each with a planted effect and a passing null), in `execution/tests/test_binding.py` only

All tests use real git and temp directories only.

**T5. `materialize` sets one shared origin per fixture.**

Setup: `generate` a synthetic database in temp, then `materialize` its parent into `T/p` and its head into `T/h`.

- **Null:**
  - `git -C T/p config --get remote.origin.url` and the same for `T/h` both exit 0 and both equal `str(Path(database).resolve())`, a non-empty value.
  - `verify_materialization` passes for both roots.
  - The returned file manifests are byte-equal to those produced in the planted run below. This shows the scanned bytes are unchanged.
  - A second database, `generate(T/'other.git')`, materialised into `T/q`, has a different origin from `T/p`.
- **Planted effect (the pre-repair materialiser):**
  - Repeat with `fixtures.git` wrapped so that it skips any call whose first git argument is `config`.
  - `git config --get remote.origin.url` then exits 1 in both roots. Scout's identity input would fall back to two distinct paths (`scan.go:621–627`), and the shared-origin predicate is false.

**T6. Composition: every bound launch sees the parent's identity after the carry.**

Run the real `execute_fixture(config, {'git':{'path':shutil.which('git')}}, 'synthetic', None, T)`. Use the real `fixture_manifest`/`generate`, `materialize`, `carry_state` and `verify_materialization`. Only `run_revision` is faked, extending the existing `CompletionTests` fake (`test_binding.py:436–462`).

For each call, the fake records:
- route, revision and `resolve_only`;
- the root's `git config --get remote.origin.url`, or `None` on exit 1;
- whether `root/.scout/state-marker` exists.

For bound/parent and reference/parent execution calls (not resolve-only), the fake writes `root/.scout/state-marker`, so that `carry_state` has something to carry.

- **Null:**
  - Every one of the 12 calls sees the same non-`None` origin, equal to the database realpath.
  - The bound/head and reference/head execution calls each see `state-marker` present. That proves they ran after the carry, with the parent's identity.
- **Planted effect:** the same run under T5's suppression wrapper records `None` origins for bound/parent and bound/head. A predicate that requires equal, non-`None` origins for the carried pair is false.

**Mutation reasoning:**
- Deleting C4 makes T5's null and T6's null fail.
- Deleting the carry (an alternative remedy rejected below) makes T6's `state-marker` assertion fail.

**Real-engine acceptance is T4: the fresh rehearsal 3.** The rehearsal review must confirm all of the following:
- (a) There is no identity error at bound/head or reference/head.
- (b) `boundroute/{parent,head}/.scout/state.sqlite` `metadata.repository_id` are equal.
- (c) Bound/head `carried_issue_ids` are disclosed, as the recipe requires.
  - The synthetic `common/unchanged.py` finding is present at both revisions, so a non-empty carried set is expected [INFERENCE].
  - If the set is empty, the review must explain why before sign-off.
- (d) `referencer/*/.scout` contains no `state.sqlite`.
- (e) All six predicates are true.

## Alternatives

| Option | Tradeoff | Decision |
|---|---|---|
| **C4: shared origin = source database path, set in `materialize`** | One line plus tests. Uses the product's own preferred identity input, applies to every route and both fixtures, and leaves file bytes and paths unchanged. | **Adopted** |
| Opaque constant label as origin (e.g. `x1-binding/synthetic`) | Works, but invents a provenance string; the database path is the true source and is already distinct per fixture. | Rejected |
| Django's GitHub URL as origin | Puts a network-addressable origin in front of Clawpatch and other origin readers, and treats synthetic and Django differently. | Rejected |
| One root per route, advancing parent to head in place | Breaks resolve-before-scan, because `Config.Root` is in the pinned effective record (`main.go:364`). Also breaks the per-root manifests and needs much more orchestration. | Rejected |
| Rewrite `metadata.repository_id` in the carried store | Tampers with the product's guard. The `plans` and `runs` rows also store the ID (`progress.go:85,238`). | Rejected |
| Drop the state carry | Violates the recipe's "independent fresh parent-to-head state lifecycles" and hides continuity. | Rejected |
| Fail-fast origin check in `carry_state` (peer) | It would duplicate the engine's own explicit refusal (`store.go:361`) and add git calls to `carry_state`. A refusal inside a canonical operation consumes the receipt either way. T6 covers the composition. | Rejected |
| Change `main.go` or the engine | Not needed (F2), and outside `:74`. | Not applicable |

## Risks

- **Carried Clawpatch files name the parent path (experiment-peer, H5).**
  - The carried `.scout/clawpatch/config.json` has `stateDir=…/boundroute/parent/.scout/clawpatch`, and `project.json` has `rootPath=…/boundroute/parent`.
  - `--state-dir` overrides `stateDir` (`third_party/clawpatch/src/config.ts:76`), and `rootPath` is never read back. The council judges this harmless for heuristic mapping [INFERENCE].
  - Rehearsal 3 is the first run to map at a carried head. Any problem refuses in scratch.
- **Later stages are untested.** Bound/head and reference/{parent,head} have never run, so further integration defects may surface in rehearsal 3. Each costs one scratch rehearsal and one round, and nothing irreversible.
- **Django-specific carry is not rehearsed.** The amendment's synthetic-only rehearsal design accepts this. C4 applies the same identity rule to Django.
- **Duration.** Each rehearsal costs about 30 min with the post-launch checks. The deadline is 2026-12-24.

## Authorised sequence

Stop at the first failure. Any failure means: retain everything, create nothing more, and open a new council round.

1. **Write this record.** Main writes it to `STUDY/preparation-reviews/execution-implementation-round-5.md` (UTF-8, LF) and records its sha256 and bytes.
2. **Freeze.**
   - No binding launch until step 9.
   - Keep the rehearsal-2 report, its `.work` tree and both rehearsal-2 records unchanged. They are superseded and are not evidence.
   - Rule O and RB-1 stay in force.
3. **Implement** C4 in `binding/fixtures.py` and T5 and T6 in `tests/test_binding.py`. No other file under `execution/` changes.
4. **Run the tests** with `/Users/adamjackson/Projects/work/scout-x1prep/.tools/venv/bin/python -I -B -m unittest`:
   - the binding suites (`test_binding*.py`);
   - `test_binding_contract.py`;
   - the full Python suite.

   The suite record must pin, at the start and the end of the run, `fixtures.py`, `acceptance.py`, `interlock.py`, `test_binding.py` and `test_binding_boundaries.py` (round-4 S2). Then rerun round-1 `.checks[0]` unchanged; it must pass.
5. **Regenerate at fixed paths (`:74`).**
   - Move `OUT/execution-source-manifest.json` `50e8c2b1…` and `OUT/reviews/implementation.json` `ac3ccf37…` byte-identically to `SCRATCH/binding-v2-build/superseded-r3/`, with before/after pins.
   - Regenerate the manifest with round 2's producer and argv. Exactly two rows may differ: `binding/fixtures.py` and `tests/test_binding.py`.
   - `OUT/binding/runtime.json` stays `647381c5…`.
6. **Writer response.** The writer writes `execution-implementation-round-5-response.md` with the new digests and the evidence from steps 3–5.
7. **Council round 6.** It needs zero issues. Its pre-receipt conditions include an in-memory reversal of C4: deleting the four inserted lines from `fixtures.py` must give SHA-256 `48552a2bd5f8492b83237f147916bde12f2137a6aee6fbae99b01cda9e40381b` (6956 B). On zero issues, Main writes `OUT/reviews/implementation.json` naming the round-6 record (9 rows, with the new manifest).
8. **No `RESULTS.md` deviation** (F5). D's existing text stays accurate.
9. **Fresh rehearsal 3.**
   - Run P3-D session checks 1–3.
   - Run the fresh synthetic-only rehearsal with the authorised argv, except `--output /Users/adamjackson/Projects/work/x1-preparation-scratch/binding-v2-rehearsal-3/report.json`. Neither that path nor its `.work` may exist.
   - Then run the round-4 step-8 post-checks: `.checks[0]` and `compare_enumeration.py` under `-I -B`.
10. **Rehearsal review.** The council rehearsal review checks T4 (a)–(e). On zero issues, Main writes `OUT/reviews/binding-v2-rehearsal.json`. Then canonical synthetic with `--rehearsal-report`, Django, `reviews/binding.json`, the remaining reviews, `seal --check` and `seal`.

## How the peers shaped this record

- **experiment-peer:**
  - supplied the deciding observation (`positiveXX/parent` vs `boundroute/parent`) and the head-store metadata read;
  - showed that `mapping.Status` was empty;
  - found the Django origin in `fetch_objects.py:65` and the carried Clawpatch parent paths (H5);
  - proposed the remote-at-materialise fix and the composition test, which shaped T6.
  - The council declined its fail-fast `carry_state` check (see Alternatives).
- **innovation-challenger:**
  - traced the identity end to end (`assess.go`, `plan.go`, `scan.go:682–691`);
  - read the materialised `.git` state;
  - mapped the test fakes that hid the defect;
  - framed the shared-root alternative and its `Config.Root` cost;
  - warned that continuity must be shown, not merely the crash avoided. That became T4 (c) and the T6 carry assertion.
- **Main council:**
  - independent reads of `scan.go:600–643`, `store.go`, `main.go:395–443`, `fixtures.py`, the head `.git/config` and the recipe's parity requirement, which rules out dropping the carry;
  - the choice of the source database path as origin;
  - the test design and the sequence.

Verdict: REMEDY CHOSEN — implementation review required
