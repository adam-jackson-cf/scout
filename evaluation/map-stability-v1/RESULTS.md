# E3 — Clawpatch map ID stability with `--source agent`

**Verdict: always-agent mapping is NOT safe as an anchor for Scout issues.** Heuristic feature IDs were 100% stable (reruns, content change, new file). Agent-mapper feature IDs were not: only 5–31% of agent IDs recur between independent runs, and only 6–29% recur when the agent is rerun against the same state dir with the prior features supplied. Stale IDs are never deleted, so persisted state keeps every orphaned ID.

## Setup (all measured, nothing fabricated)
- Clawpatch v0.4.0 @ 587b9a1, built via `scripts/build-clawpatch.sh` (patch applied, CLAWPATCH_OUTPUT_DIR=/tmp/scout-evidence/clawpatch-build). The machine had pnpm 10.23.0 and no corepack, so pnpm 11.1.2 was installed with `npm i pnpm@11.1.2` under /tmp/scout-evidence/pnpm11 and put first on PATH. The build succeeded and reported 0.4.0.
- Repos: copies of fastapi (d666ccb, Python) and axum (562fa20, Rust) under /tmp/scout-evidence/e3/. All runs used `--root <copy> --state-dir <external dir> --json`.
- Provider: Clawpatch default `codex`, model `default`, no reasoning-effort override. **Clawpatch reports no tokens or cost** in its JSON or progress output; none were measured. Wall time only.
- Runs per repo:
  - H1–H3: fresh state, `--source heuristic`.
  - A1–A3: fresh state each, `--source agent`.
  - A1_rerun1 and A1_rerun2: same state dir as A1, `--source agent`.
  - M1_content: edit a function body in an agent-owned file, remap `--source agent` on the A1 state.
  - M2_struct: add a new module file, remap `--source agent` on the A1 state.
- Scripts: `run_e3.py`, `mutate_e3.py`, `analyze_e3.py`. Raw per-run JSON (result, stderr tail, full feature snapshots) is in `raw/<repo>/*.json`. Derived metrics are in `metrics.json` and `analysis_output.txt`.
- ID derivation (clawpatch `src/mapper.ts:223`): `hash(kind, source, entryPath, identityKey|command|route|symbol)`. For agent features, kind and entry path are chosen by the LLM, so a different choice gives a different ID.
- Merge (`agent-mapper.ts:295`): union by featureId, heuristic first, agent overrides on collision. There is no semantic matching of agent features across runs.

## 1. Features per run

| repo | run | s | total features | heuristic-source | agent-mapper | result new/changed/stale |
|---|---|---|---|---|---|---|
| fastapi | H1/H2/H3 | ~1 | 154 | 154 | 0 | 154/0/0 |
| fastapi | A1 | 153 | 187 | 154 | 33 | 187/0/0 |
| fastapi | A2 | 140 | 185 | 154 | 31 | 185/0/0 |
| fastapi | A3 | 266 | 202 | 154 | 48 | 202/0/0 |
| fastapi | A1_rerun1 | 153 | 218 (state) | 154 | 64 | 31/3/30 |
| fastapi | A1_rerun2 | 155 | 244 (state) | 154 | 90 | 26/12/52 |
| axum | H1/H2/H3 | ~0.5 | 52 | 52 | 0 | 52/0/0 |
| axum | A1 | 184 | 101 | 52 | 49 | 101/0/0 |
| axum | A2 | 157 | 97 | 52 | 45 | 97/0/0 |
| axum | A3 | 160 | 97 | 52 | 45 | 97/0/0 |
| axum | A1_rerun1 | 142 | 130 (state) | 52 | 78 | 29/9/40 |
| axum | A1_rerun2 | 141 | 151 (state) | 52 | 99 | 21/23/55 |

On the rerun rows, "state" is the total persisted in the state dir. Clawpatch's own `features` field is the merged set for that run, e.g. fastapi rerun1 = 188.

## 2. ID stability

| metric | fastapi | axum |
|---|---|---|
| Heuristic Jaccard H1 vs H2, H1 vs H3 | 1.00, 1.00 | 1.00, 1.00 |
| Fresh agent runs, all-feature Jaccard (A1-A2 / A1-A3 / A2-A3) | 0.747 / 0.684 / 0.735 | 0.512 / 0.467 / 0.630 |
| Fresh agent runs, **agent-only** Jaccard | 0.085 / 0.052 / 0.145 | 0.190 / 0.133 / 0.343 |
| Share of A1 agent IDs present in A2 / A3 | 15% / 12% | 31% / 22% |
| Same state: agent IDs reused from the previous state in the next merged output | rerun1 3/33 = 9%; rerun2 12/64 = 19% | rerun1 9/49 = 18%; rerun2 23/78 = 29% |
| Same state: share of previous state IDs still present in state | 100% (never deleted) | 100% |
| Same state: per-run new agent IDs | 31, 26 | 29, 21 |
| Same state: cumulative orphan growth (`stale` in the result) | 30, then 52, then 76, then 102 | 40, then 55, then 80, then 86 |

"Reused" is computed as `result.features - result.new - heuristic_count`. Every reused agent ID had `changed > 0`: its file list or title differed, so the record was rewritten.

The all-feature Jaccard looks high only because 154 or 52 heuristic IDs are identical in every run. Heuristic IDs are fully stable, and agent IDs are largely not.

## 3. Owned-file coverage and ownership churn
Denominator: git-tracked `.py` files in fastapi (1132, including docs_src and tests) and `.rs` files in axum (264). Clawpatch's own inventory counted 716 and 158 source files respectively, so my denominators are broader.

| repo | H | A1 | A2 | A3 | A1_rerun2 |
|---|---|---|---|---|---|
| fastapi | 306 (27%) | 508 (45%) | 435 (38%) | 742 (66%) | 610 (54%) |
| axum | 50 (19%) | 153 (58%) | 153 (58%) | 153 (58%) | 153 (58%) |

Files whose owning-feature set differs between runs, among files owned in both runs:

| comparison | fastapi | axum |
|---|---|---|
| Fresh A1 vs A2 | 209/427 (49%) | 125/180 (69%) |
| Fresh A1 vs A3 | 294/569 (52%) | 145/180 (81%) |
| Fresh A2 vs A3 | 208/475 (44%) | 118/184 (64%) |
| Same state, A1 to rerun1 | 240/571 (42%) | 142/226 (63%) |
| Same state, rerun1 to rerun2 | 203/676 (30%) | 74/226 (33%) |

Part of the same-state change comes from the owner set growing as features accumulate (a file gains extra owners), not only from reassignment.

Coverage of files owned by at least one feature varies by up to 28 percentage points between fresh fastapi runs (435 to 742). Axum coverage was constant.

## 4. Small changes (same state, `--source agent`, after rerun2)
- **Content-only change.**
  - fastapi: inserted one statement in `get_graphql_issue_edges` in `.github/actions/people/app/main.py`.
  - axum: inserted one statement in `Clone::clone` in `axum/src/routing/mod.rs`.
  - Heuristic IDs were unchanged in both repos. Heuristic features owning the edited file: none (agent-only coverage).
  - The remap produced 23 new agent features (fastapi) and 19 new agent features (axum) that had nothing to do with the edit. The agent re-partitions the repo on every call.
  - The prior owners of the edited file stayed in state. In axum, a third owner (`feat_library_d581e03e52`, "Router topology and path dispatch") appeared for `routing/mod.rs`.
  - Of the previous state's features, 0 were marked `stale` or removed. The result's `stale` count rose to 76 (fastapi) and 80 (axum), but those are just not-in-merged-set.
- **Structural change (new module file).**
  - fastapi: `fastapi/e3_new_module.py` was picked up by the existing heuristic `python-source-group` feature `feat_library_999642215b`. That feature's ID is unchanged and only its file list grew. The agent also created `feat_unknown_bd6ed3e381` for it.
  - axum: `axum/src/e3_new_module.rs` was owned only by a new agent feature, `feat_unknown_4f11e64d75`. The heuristic mapper did not pick it up.
  - No heuristic IDs were added or removed in either repo. The remap added 29 (fastapi) and 13 (axum) new agent IDs.
- Agent-run time: fastapi 140–297 s, axum 119–184 s. The `map` call is one LLM invocation of 2–5 minutes, versus roughly 1 s for heuristic.

## 5. `--source auto`
From the inventory line, which the heuristic run itself reports:
- **fastapi**: `sourceFiles=716 ownedSourceFiles=65` gives 9% coverage, which is below the 25% threshold (`weakMap`, `agent-mapper.ts:496`). The map is **weak** with reason "heuristic map covers 9% of source files". `auto` would invoke the agent and show the same churn as `agent`.
- **axum**: `sourceFiles=158 ownedSourceFiles=51` gives 32% coverage, which is not weak. `auto` would stay on the heuristic map (52 features, perfectly stable).

## Conclusions and mitigations
1. Always running the agent mapper as the anchor source orphans issues. In every pair of fresh agent runs, 70–95% of agent IDs differ. Even when Clawpatch is given the prior features, 71–91% of the agent features it emits get new IDs. This is inferred from the ID formula: the agent picks kind and entryPath and nothing reconciles them against prior features.
2. Clawpatch never deletes old features, so persisted issues keep resolving to an ID, but to an ever-growing, overlapping set. fastapi went from 187 to 296 features and axum from 101 to 183 after five agent runs against one state dir. Anchoring to feature IDs makes the "which feature owns this issue" answer drift, and file-to-feature mapping becomes many-to-many.
3. Heuristic IDs are stable, including across a new file where the heuristic mapper picks it up. The heuristic mapper is the only safe feature-ID anchor from this data. It can still fail for newly added files (axum's new file got no heuristic owner).
4. Mitigations the data supports, in order of strength:
   - **Anchor issues to file path plus symbol or content fingerprint.** Treat the feature as a secondary, re-derivable label stored with the issue and updated on each map. File ownership is volatile, but file paths are not.
   - **Run the agent only for unowned files.** In axum, heuristic ownership covers 19% of `.rs` files and `auto` keeps it. Restricting the agent to files with no heuristic owner reduces the surface, but its own IDs still churn.
   - **Pin accepted agent features.** Snapshot an agent feature (ID plus a stable file set) once accepted, then re-attach by file-set overlap instead of by hash. Clawpatch does not do this today; it would need a Scout-side reconciliation layer.
   - **Use `auto`, not `agent`, as the default.** It avoids the agent entirely on repos like axum.
5. Limits: one repo per language, 3 fresh runs, a single provider and model, default reasoning effort, no token or cost data. Reuse rates for same-state reruns are from two reruns only.
