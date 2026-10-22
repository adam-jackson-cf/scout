# Clawpatch 0.7.x compatibility review for Scout

**Recommendation: upgrade with listed changes.** Target `v0.7.3` (`c46c0c6c4822fdd46497c8c9c640dc9ee613954d`, 2026-09-05), the latest 0.7.x tag. Upstream already has `v0.8.0`–`v0.8.2` (HEAD is 0.8.3-unreleased), so "latest 0.7.x" is not the latest release; see §8.

No Scout parser change is needed. The cost is a mechanical re-pin, one ported patch (`clawpatch-safe-0.7.3.patch`, produced here and already building), and one expected migration effect on Rust repos. It is not zero-touch: the existing patch does not apply, and 50 of 52 axum features change record content on first map (§4).

All numbers below were measured in this session. Raw outputs are in `raw/`, `analysis_output.txt`, `loadcheck_*.txt`, `go_test_clawpatch.txt` and `build_*.log`. Reproduce with `run_compat.py <repo>`, `analyze_compat.py` and `go run ./evaluation/clawpatch-0.7-compat/loadcheck <state-dir>...`.

## 1. Setup
- Fresh clone of `https://github.com/openclaw/clawpatch.git` at `/tmp/scout-evidence/clawpatch-07`. The existing `/Users/adamjackson/Projects/examples/clawpatch` checkout was not touched. `git log 587b9a1..v0.7.3` is 116 commits; `src/` changes by 89 files (+12029/-5550), tests included.
- Node v25.6.1; pnpm 11.25.0, which is the `packageManager` pin in 0.7.3. Stock 0.7.3 builds clean and reports `0.7.3` (`build_0.7.3_stock.log`). The pnpm 11.1.2 under `/tmp/scout-evidence/pnpm11` was placed first on PATH for the E3-style runs only.
- Repos: fresh copies of `/tmp/scout-evidence/repos/{fastapi,axum,caddy,nest,webpack}` under `/tmp/scout-evidence/cp07-work/`; originals untouched. The fastapi and axum commits are the E3 ones (d666ccb, 562fa20).
- The 0.4.0 baseline is the existing patched build in `/tmp/scout-evidence/clawpatch-build/587b9a1…`. The 0.7.3 build is stock 0.7.3 plus the ported Scout patch (§3).
- Provider for agent runs: codex-cli 0.162.1, model default. Clawpatch reports no tokens or cost, so only wall time is available.

## 2. Compatibility table

| Item | 0.4.0 (pinned) | 0.7.3 | Impact on Scout | Migration action |
|---|---|---|---|---|
| CLI: `init`, `map --source heuristic\|auto\|agent`, `review --feature`, global `--root/--state-dir/--json` | as Scout uses | Unchanged. `map` flags: dry-run, source, provider, model, reasoning-effort, skip-git-repo-check. 0.8.0 adds `--link-http`. `cli.ts` was refactored into a `commandSpecs` table; `parseArgs` behaviour is equivalent. | None. | None. |
| New commands/flags | n/a | `review --feature-list`, `review --no-registry-verify`, `clean-locks --stale-only`; Claude host-auth context; provider split into `src/providers/*`; MiniMax provider added in 0.6.0 and removed in 0.7.0. | Scout only uses `review --feature <id>`; `--feature-list` could replace N serialized spawns but is not required. | Optional later. |
| `map` JSON | `{features,new,changed,stale,source,usedAgent,reason,next}` | Identical keys (measured). | Scout's `client.go` does not parse map stdout and reads `features/*.json`. | None. |
| `review --json` | `{run,reviewed,findings,jobs,report,next}` plus Scout-patch `reviewManifests`, `errors` | Stock 0.7.3 output is the same without the patch additions. With the ported patch, a real `review --feature feat_cli-command_ae468c27b2` run on axum returned `run, reviewed, findings, jobs, report, reviewManifests[{featureId,manifest{approximateTokens,includedFiles,maxContextFiles,maxOwnedFiles,omittedFiles,promptBytes}}], errors:[], next` (`raw/review_axum.stdout.json`). | Matches `reviewResponse` and `reviewPromptManifest` in `review.go:86-134`. | None once the patch is ported. |
| New drop code | `schema-drop`, `validation-drop` | Adds `registry-verifier-drop` (layer `registry-verifier`). The verifier is opt-in (`registryVerifier.enabled=false` by default) and sends package names to the public npm registry. | Scout's generic "any `errors` ⇒ review incomplete" rule (`review.go:356`) still holds. | Leave disabled; do not enable in Scout config. |
| featureId derivation | `stableId("feat",[kind,source,entryPath, identityKey??command??route??symbol??""])` | `seedIdentityParts()` in new `mapper-reconciliation.ts` yields the same four parts; `id.ts` is unchanged. The dedupe key moved from a `:` string to `JSON.stringify` of the parts, which only avoids collisions. | None for unchanged seeds. | None. |
| Feature record schema, `schemaVersion` | `schemaVersion: 1` | Still literal 1. `types.ts` diff is config-only: `provider.codexConfig`, `registryVerifier.enabled`. Feature and finding keys are identical (compared key sets; the only per-feature difference is data, §4). | `decodeFeature` / `Feature.Validate` accept every 0.7.3 record. | None. |
| State layout | `config.json project.json features/ findings/ locks/ patches/ reports/ runs/` | Same dirs. Locks gained `proper-lockfile` mutation locks and same-host dead-lock reclaim, with `clean-locks --stale-only`. Config adds `provider.codexConfig` and `registryVerifier`. `codexConfig` is rejected unless set via `--config` or `CLAWPATCH_CONFIG`. | Scout reads `config.json` for `schemaVersion` and `review.maxFindingsPerFeature` only (`review.go:402-411`), both unchanged. | None. |
| Reconciliation (changed/stale/missing) | Changed record ⇒ status reset via `statusForChangedFeature`; missing IDs counted as `stale`, never deleted | Same semantics. `stableFeatureJson()` was extracted, still ignoring createdAt/updatedAt/lock/analysisHistory. | Unchanged. A mapper that adds fields to a record resets reviewed features (§4). | Plan for one-time re-review of changed features. |
| Weak-map thresholds | zero features; config-only; `<25%` coverage with ≥4 source files; ≤2 meaningful with ≥12 files | Identical (`agent-mapper.ts:469-491`). fastapi still `weak=true` at 9%. | None. | None. |
| Scoped/partial agent mapping | none | **None added**, through 0.8.2 as well. The agent still maps the whole inventory and merges by union (`mapFeatureSeeds`). No path-scoped or delta map flag. | Agent churn is not fixed by upstream (§5). | Keep the Scout-side reconciliation plan from E3. |
| Review limits/defaults | maxContextFiles 24, maxOwnedFiles 12, maxFindingsPerFeature 10, file char limit 24 000, jobs `floor(cpu/2)` clamped to [1,10] | Unchanged (`config.ts`, `prompt.ts:REVIEW_PROMPT_FILE_CHAR_LIMIT`). New: CUDA prompt guidance; `--since` / `--feature` reviews now include non-pending features. | Same bounded prompt; `--feature` already allowed non-pending review. | None. |
| Provider config | codex default; `CLAWPATCH_CODEX_TIMEOUT_MS` default 300 000 | Same default 300 s (`providers/codex.ts`); env names unchanged. A 300 s timeout is shared with 0.4.0 (see fastapi note in §5). | Map on large repos can time out. | Set `CLAWPATCH_CODEX_TIMEOUT_MS` in Scout's environment for map. |
| Heuristic mappers | n/a | Rust: new `rust-source-group` seeds (+25 on axum) and Cargo-manifest context on existing features. Python, Go, React and node-routes were rewritten. fastapi, caddy, nest and webpack are byte-identical in content (§4). | Rust repos see one-time churn. Python/Go/TS (tier-1) were stable on the five repos measured. | See §4. |
| Node / pnpm | `engines.node >=22`; `packageManager pnpm@11.1.2` | `engines.node >=22`; `packageManager pnpm@11.25.0` (0.8.3-dev: 11.28.4). | `build-clawpatch.sh` hard-requires `pnpm@11.1.2` and checks `manifest.packageManager`. | Bump `PNPM_VERSION` (and workflows) to 11.25.0. |
| Licence | MIT, © 2026 openclaw | MIT, same text | None. | None. |

## 3. `clawpatch-safe.patch` status
- Original patch against 0.7.3: **does not apply**. `git apply --check` (`patch-check-0.7.3.txt`): per-file result is `src/prompt.ts` OK; `src/app.ts` fails at :23; `src/agent-mapper.ts` fails at :1; `src/mappers/shared.ts` fails at :175. A `--reject` run: prompt.ts clean; agent-mapper hunks 2 and 3 apply and only the import hunk fails; shared.ts only the first two hunks fail, the `shouldSkip` hunk applies; all 8 `app.ts` hunks fail.
- Cause: upstream did not add equivalent redaction; it split code. `grep` for redact/sanitize in 0.7.3 `src/` finds only provider env scrubbing and tests that stop failure messages from previewing stdout. **Content redaction of prompts, findings and errors is still absent upstream, so the patch must be ported, not dropped.**
  - Review logic moved from `app.ts` to the new `src/review.ts` (`reviewCommand`, `reviewFeature`).
  - `provider` types moved to `provider-types.js`; the import line changed.
  - `shared.ts` replaced the two walkers with a single `walkByPolicy`, so the original `.scout` exclusion hunks no longer match.
- **Ported patch: `clawpatch-safe-0.7.3.patch`** (4 files, +137/-20; sha256 `f39d73dad42b5d6701b1fee51f7784d3eb2b1548808d1ee47d8ef146ca97a865`).
  - Same behaviour: prompt/output/error redaction, agent-map metadata sanitisation, `.scout` excluded (shouldSkip regex plus `isScoutStatePath` in `walkByPolicy`), `reviewManifests` and `errors` in review JSON, finding-cap drop record.
  - `git apply --check` passes on a pristine `v0.7.3` archive.
  - Builds with `pnpm install --frozen-lockfile && pnpm run build`, and `tsc --noEmit` passes (`build_0.7.3_patched.log`). The built CLI reports `0.7.3`.
  - Exercised end to end by the axum agent runs, the migration map and the review run above. **Not run:** upstream vitest and any redaction unit test for the ported patch; sanitiser code in `prompt.ts` is unchanged text.
  - The ported `reviewFeature` also sanitises `droppedFindings` that now come with the registry-verifier layer: `sanitizeProviderReviewOutput(validated)` covers `validated.droppedFindings`.
  - The earlier patch's `sanitizeSensitiveText` only covers the paths the old patch covered. New 0.7.x paths that Scout does not call (revalidate prompts, `open-pr`, `fix`) are not sanitised; Scout does not call them.
- Against `v0.8.2` the original patch also fails and the ported patch fails at `review.ts:298` (upstream shifted lines); it needs another re-anchor.

## 4. ID migration (heuristic) — 0.4.0 → 0.7.3 on identical repo copies

| repo | 0.4.0 | 0.7.3 | common IDs | only 0.4.0 | only 0.7.3 | Jaccard | 0.4.0 IDs surviving |
|---|---|---|---|---|---|---|---|
| fastapi | 154 | 154 | 154 | 0 | 0 | 1.000 | 100% |
| axum | 52 | 77 | 52 | 0 | 25 (`rust-source-group`) | 0.675 | 100% |
| caddy (Go) | 44 | 44 | 44 | 0 | 0 | 1.000 | 100% |
| nest (TS) | 22 | 22 | 22 | 0 | 0 | 1.000 | 100% |
| webpack (TS) | 75 | 75 | 75 | 0 | 0 | 1.000 | 100% |

- **No feature ID that existed in 0.4.0 disappeared on any of the five repos.** Python, Go and TypeScript were stable.
- 0.7.3 heuristic runs are deterministic: H07_1 vs H07_2 Jaccard 1.000 on fastapi and axum.
- Content changes on common IDs: fastapi, caddy, nest, webpack **0**. axum **50 of 52** `contextFiles` gained a Cargo manifest, e.g. `examples/sse/Cargo.toml` (new "Rust seed context for Cargo manifests", changelog 0.7.2). Axum owned files rose from 52 to 157 distinct (`ownedSourceFiles` 51 → 149 of 158 sources) because the new `rust-source-group` features own previously unmapped sources.
- Migration run: a 0.4.0 state (`H04_1`) mapped by 0.7.3 on the same repo (`raw/migrate_*_04state_to_07map.json`):
  - fastapi: `new 0, changed 0, stale 0`.
  - axum: `new 25, changed 50, stale 0`.
  - **On a state with reviewed features, those 50 axum features would reset to pending** (`statusForChangedFeature`, `mapper.ts`) and be re-reviewed once. The measured state was fresh, so all features were already pending; the reset itself is inferred from that code path rather than observed.
- Scout impact: one-time cost on Rust repos only. Tier-1 repos measured no change. Per the requirements, Scout does not rewrite IDs and correlates issues across features, so nothing is orphaned.

## 5. Agent-map stability, 0.7.3 vs 0.4.0 (reduced E3: 2 fresh runs per repo, `--source agent`)

| repo | metric | 0.4.0 (E3, 3 runs) | 0.7.3 (2 runs) |
|---|---|---|---|
| fastapi | agent features per run | 33 / 31 / 48 | 36 / 33 |
| fastapi | agent-only Jaccard A1–A2 | 0.085 / 0.052 / 0.145 | **0.062** |
| fastapi | share of A1 agent IDs present in A2 | 15% / 12% | **11%** (4 of 36) |
| fastapi | all-feature Jaccard | 0.747 / 0.684 / 0.735 | 0.721 |
| fastapi | wall time | 140–266 s | 182 s and 251 s |
| axum | agent features per run | 49 / 45 / 45 | 35 / 46 |
| axum | agent-only Jaccard A1–A2 | 0.190 / 0.133 / 0.343 | **0.157** |
| axum | share of A1 agent IDs present in A2 | 31% / 22% | **31%** (11 of 35) |
| axum | all-feature Jaccard | 0.512 / 0.467 / 0.630 | 0.599 |
| axum | wall time | 157–184 s | 191 s and 243 s |

- **Agent-ID churn is unchanged.** 0.7.3 agent IDs recur between independent runs at about the 0.4.0 rate (6–16% Jaccard on agent IDs, 11–31% survival). No scoped, partial, or delta agent mapping exists, so E3's conclusion stands: do not anchor issues on agent IDs, re-attach by file-set overlap, and keep the heuristic map as the anchor.
- Owned-file churn: fastapi agent-owned files A1 317 vs A2 297, overlap 252; axum 183 vs 180, overlap 179. Not compared to E3's per-file owner-set metric; the numbers are not directly comparable.
- **fastapi first attempt failed:** both agent runs ran into `command timed out after 300000ms` (rc 4, 301 s; `raw/fastapi_first_attempt_300s_timeout/`), because the codex agent call exceeded the 300 s default (stderr: `command timed out after 300000ms`); that default is the same in both versions and the fastapi/axum runs were executing in parallel, so machine load may have contributed (not isolated) (0.4.0 E3 runs took 140–297 s, so this is the same margin). They were rerun with `CLAWPATCH_CODEX_TIMEOUT_MS=900000`, two in parallel, and both completed (`raw/fastapi/A07_*.json`). Treat the 300 s default as marginal for large repos on either version.
- Not measured: same-state reruns, content/structural mutations and stale-feature accumulation (E3 §2/§4). That behaviour depends on the unchanged `mapFeatureSeeds` merge and is expected to match 0.4.0, but this review did not measure it.
- `--source auto` on axum: 0.4.0 coverage was 32% (51/158, not weak). 0.7.3 coverage is 94% (149/158), still not weak. fastapi remains weak at 9% under both, so `auto` would run the agent identically.

## 6. Scout record parsing against 0.7.3 output
- `go test -count=1 ./internal/clawpatch/...` unchanged: **ok** (`go_test_clawpatch.txt`). These tests use fixtures, so they do not exercise 0.7.3 output.
- Throwaway `loadcheck/main.go` runs Scout's `clawpatch.DecodeFeature` on every record and `clawpatch.LoadFeatures` per state dir (`loadcheck_heuristic.txt`, `loadcheck_agent.txt`):
  - fastapi H04 154/154 and H07 154/154 decoded.
  - axum H04 52/52 and H07 77/77 decoded.
  - fastapi A07_1/A07_2 and axum A07_1 (112) and A07_2 (123) decoded; no errors.
  - `LoadFeatures` returned nil error for every state.
  - Review output shape matches `reviewResponse` (§2). A single review run produced 0 findings, so the finding-record parser (`rawFinding`, `review.go:136-155`) was **not** exercised on real 0.7.3 findings; `types.ts` finding schema is unchanged by diff.

## 7. Required Scout changes (file:line)
1. `internal/clawpatch/client.go:19-20`: `Revision = "c46c0c6c4822fdd46497c8c9c640dc9ee613954d"`, `Version = "0.7.3"`. These also drive `verifyIdentityLocked` (`:215`) and `validatePinnedBinaryPath` (`:241`, directory name must equal the revision).
2. `internal/clawpatch/review.go:19-20`: new safety-patch version string (e.g. `scout-review-sanitize-v2`) and `reviewSafetyPatchSHA256 = f39d73da…a865`. `review.go:253` compares the build marker, and `SafetyFingerprint()` at `:27` is the safety fingerprint Scout uses; I did not trace its consumers, so confirm that changing it invalidates cached review state as intended.
3. `internal/clawpatch/review.go:479`: `Source: "clawpatch@0.7.3+scout-review-sanitize-v2"`.
4. `scripts/build-clawpatch.sh:5-9`: revision, version, `PNPM_VERSION="11.25.0"` (also checked at `:81` and `:107`), patch version and SHA256. The identity check at `:99-112` needs no code change.
5. `scripts/clawpatch-safe.patch`: replace with `evaluation/clawpatch-0.7-compat/clawpatch-safe-0.7.3.patch` (same file name so `:13` is unchanged).
6. `.github/workflows/release.yml:18-20,102` and `scout-nightly.yml:37,39`: `PNPM_VERSION`, `CLAWPATCH_REVISION`, `CLAWPATCH_VERSION`, `safetyPatchSha256`.
7. `SCOUT_REBUILD_REQUIREMENTS.md:31,128,193,285-286` and `docs/scout-design.md:130,276,345,385`, `docs/scout-design-brief.md:15,70`: update pin; req §3 expects "explicit compatibility decision and migration". This review is that decision once accepted.
8. Fixtures and tests that embed `0.4.0` or the old SHA: none found outside the listed files by `grep`.
9. Migration step on first map: expect Rust repos to report 25 new and many changed features and a one-time re-review of changed features. No ID rewriting is needed.
10. Set `CLAWPATCH_CODEX_TIMEOUT_MS` (e.g. 900000) for Scout's `map` invocation, because the 300 s default timed out in this session's fastapi agent runs (separate from the upgrade, but it blocks agent maps on large repos).

Effort estimate: about 0.5 day. Constants and script/workflow edits are mechanical (items 1-6); docs are item 7. A rebuild plus the unchanged Go tests verifies items 1-5. The porting work is already done (item 5).

## 8. Notes on 0.8.x and recommendation
- Upstream has 0.8.0–0.8.2, all MIT, Node >=22. Changes after 0.7.3 (`git diff v0.7.3 v0.8.2`, 37 files in `src/`): `map/review --link-http`, rename-aware diff review, nested-project repair fixes, and shared mapper refactors (`workspace-patterns.ts`, `python-metadata.ts`, `source-comments.ts`). I did not build or measure 0.8.x. The ported patch fails there at `review.ts:298` and needs another re-anchor. Choose 0.7.3 now if the aim is the compatibility review requested; evaluate 0.8.2 only if its rename-aware diff review matters.
- **Upgrade with the changes above.** Rationale: ID derivation, feature/finding schema, CLI, review limits and weak-map rules are unchanged; Scout's loader decodes all 0.7.3 records; 0.4.0 IDs are 100% retained on five repos in three tier-1 languages plus Rust.
- **Risks that remain:**
  - Agent IDs are as unstable as before, so nothing upstream removes the need for Scout-side reconciliation.
  - Rust repos see one-time reset of changed features.
  - The ported patch has not been through upstream's vitest or a redaction unit test.
  - Provider timeout is still 300 s.
- **Blockers:** none for 0.7.3. One real review that yields findings (to exercise `rawFinding` on 0.7.3 output) was not obtained: the one review run returned 0 findings.
