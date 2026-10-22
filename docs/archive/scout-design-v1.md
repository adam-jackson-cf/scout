# Scout — final design and planned approach

Status: council decision memo (2026-10-17). Produced by an innovation-council pass over `docs/scout-design-brief.md`: an independent adversarial analysis by a different model family (`ScoutChallenger`, transcript `history://ScoutCouncil.ScoutChallenger`) was developed in parallel with the council's own, then synthesised against repository evidence. §2 records where the challenger changed the recommendation. Nothing here has been implemented; every acceptance scenario is future work.

Evidence notation. `brief §n` = `docs/scout-design-brief.md`; `req §n` = `SCOUT_REBUILD_REQUIREMENTS.md`; E1/E2/E3 = `evaluation/{inertia-v1,jev-triage-v1,map-stability-v1}/RESULTS.md`; `SCB` = `/Users/adamjackson/Projects/inflight/slop-code-bench`; `FXL` = `/Users/adamjackson/Projects/inflight/factory-x-loops/factory-x-loops`; `CP` = `/Users/adamjackson/Projects/examples/clawpatch` (pinned 0.4.0 source). Scout paths are repository-relative. Claims not backed by a cited file or measurement are tagged **[ASSUMPTION]**. Atlas (`/tmp/atlas-review`) was absent during this pass; Atlas statements are taken from brief §3 only.

---

## 1. Product definition and non-goals

**Scout is a discovery-only CI guardian for a single repository.** It runs on pull requests, on default-branch pushes, nightly, and locally; it surfaces evidence-backed issues and descriptive quality measurements into a durable, triageable ledger; and it blocks a PR only for a small allow-list of deterministic, high-precision mechanisms that the PR itself introduced or re-introduced. Code quality as the driver of long-term inertia is the first-class family; architecture, security and performance remain families with lower implementation priority (brief §1).

One sentence per surface:

- **Ledger**: stable issue IDs anchored to source (not to feature IDs), audited dispositions, sanitized evidence, per-file/per-feature descriptive signals with history.
- **Guardian**: differential gate (head vs. same-policy base) on `new`/`regressed` issues in blocking lanes; everything else is reported.
- **Mapping**: Clawpatch heuristic map is the canonical feature definition; the agent mapper is used on every *authorised map refresh* so every file gets a definition, but its IDs never enter issue identity.
- **Judgment placement**: deterministic code decides eligibility, identity, gates; Jev ranks and routes; reasoning LLMs are opt-in discovery; humans own dispositions.

**Non-goals** (explicit, so the spec can shrink):

1. No remediation: never `clawpatch fix`, never edit the target, no approval/implementation workflow in the product (req §7's approval machinery is cut; see §9).
2. No execution of the target's tests or scripts; CI artifacts (JUnit/coverage) may be *read* when present and provenance-matched (brief §4.9).
3. No inertia *score* as a gate or severity input: E1 found no predictive support (E1 "Decision"; brief §6 E1).
4. No claim of seven-language semantic parity at launch; a capability matrix states what is measured vs. detected vs. semantically reviewed per language (§5.4).
5. No estate-level (multi-repo) dashboarding; Scout's ledger is designed to be *consumed* by Atlas-style tooling, not to replace it (brief §3 Atlas).
6. No Enaible parity as an acceptance gate; the corpus becomes a regression fixture (brief §2 "overfit"; req §11.1–11.2 moved to appendix).

---

## 2. Council synthesis — where the challenger changed the recommendation

| Topic | Council's initial leaning | Challenger's position | Resolution and evidence |
|---|---|---|---|
| Agent mapper cadence | Heuristic canonical; agent only on *unowned files*, rarely; Jev "assign to existing feature" as default alternative | Both mappers on every authorised *refresh* (bootstrap, scheduled); snapshot reused across CI runs; "agent only for unowned regions" is not a supported Clawpatch behaviour | **Adopted challenger's cadence.** `CP/src/agent-mapper.ts:129-180` offers only `heuristic|auto|agent` over the whole root; there is no scoped agent call. Jev-assign is demoted to experiment X3 (§13). The council keeps its Scout-side reconciliation layer (§4.3) because E3 §2 shows 71–91 % agent-ID churn even with prior features supplied. |
| New import cycle as blocking | Block by default | Report-only unless the repo adopts a cycle policy | **Adopted challenger's.** No FP-rate evidence exists for cycle detection on real PRs; the blocking allow-list needs the promotion rule in §7.4, not a prior. |
| "New or regressed only" gate | Sufficient | Insufficient: "new in DB" ≠ "introduced by the PR" (tool upgrade, mapper churn, first adoption) | **Adopted.** Attribution requires base and head evaluated under the same tool/rule/policy digests; otherwise attribution = `unknown` → report-only (§7.3). |
| Jev ranking | Route-only, store annotation | Also: compare against a *deterministic* ranking baseline; keep unranked view; audit random low-ranked samples because ranking becomes suppression in practice | **Adopted all three.** E2 "Caveats" 3: many negatives are path-visible, so a path/rule heuristic may capture much of the apparent gain **[ASSUMPTION]** until X1 measures it. |
| Issue identity defect scope | Feature IDs in the ID hash (`internal/issues/correlate.go:129,333-340`) | Also: `internal/state/issue_observations.go:94-102` hashes line/byte positions and `issues.go:455-461` wipes the disposition when they change → a PR inserting lines above an accepted finding re-opens it | **Verified and adopted.** Both are P0 blockers before any gate (§3.3, §10 Phase 1). |
| First deep languages | Python + TS/JS + Go | Go + Python, TS/JS second | **Adopted challenger's.** Go dogfoods Scout; Python reuses Ruff and FXL's `python-guardrails` rule shapes (`FXL/docs/references/verify-gate.md:33-50`). No evidence favours TS first; corpus has webpack/nest so TS is tier 2. |
| Outcome metric | Flag precision + follow-through + change-success (descriptive, gated on classifier) | Scorecard of four, change-success experimental only | Converged; §8. |
| Supply-chain of the guardian | Not considered | Never build Scout from PR source; pinned release, read-only tokens, fork isolation | **Adopted** (§7.6). Current workflow builds Scout from the checked-out tree (`.github/workflows/scout-nightly.yml` "Build Scout outside the assessed repository" step builds `./cmd/scout` of the checkout). |

Where the council **did not** follow the challenger: the challenger proposed "paired deterministic base/head scans with no incremental cache" as the simplest viable alternative (its §6.7). The council keeps the existing planner because it is already finer than Atlas's HEAD-keyed cache (brief §3) and tested (`internal/inventory/*_test.go`); the paired scan is retained as the *fallback* when no comparable base coverage exists (§7.3).

---

## 3. Architecture and data model

### 3.1 Components

```
cmd/scout ─┬─ engine (lanes, base/head comparison, gate)
           ├─ inventory (scan/diff/plan; kept)
           ├─ mapping (clawpatch client + Scout-side feature reconciliation; §4)
           ├─ analyzers (runner/spec kept; catalogs re-curated; §5)
           ├─ measures (NEW: lizard/jscpd/import-graph adapters → signals)
           ├─ guardrails (NEW: against-the-base config/suppression/test-quality rules)
           ├─ decisions (NEW: typed decision adapter — Jev; §6)
           ├─ contextual (LLM discovery lane; opt-in; kept, demoted)
           ├─ issues (identity/matching; rewritten; §3.3)
           ├─ state (SQLite WAL + content-addressed artifacts; kept) + export/import (NEW)
           ├─ baseline (NEW: git state-branch publish/restore; §7.2)
           ├─ selfeval (eval package renamed; scorecard; §8)
           └─ web (read-only; kept)
```

### 3.2 Record types

Three durable record kinds plus the existing immutable `Run`/`Observation`/`Coverage` (req §7 lines 96-105 already separate observations, issues and coverage; the model is extended, not replaced — challenger Q1, adopted).

| Record | Nature | Identity | Can gate? | Carries |
|---|---|---|---|---|
| **Signal** (new) | Descriptive measurement at a scope (function / file / feature / repo) at a snapshot | `(metric_id@version, scope_kind, scope_anchor, snapshot)` — append-only time series | **Never** (E1) | numerator, denominator, exclusions manifest, tool version, missingness flag (`unmeasured` ≠ 0), claim level = `descriptive` |
| **Issue** (existing, identity rewritten) | Concrete mechanism with sanitized evidence at a source anchor | Opaque durable ID; matched by §3.3 | Only in blocking lanes, only when attribution = `introduced`/`regressed` | mechanism, rule/source, anchor set, severity (rule-defined, never score-derived), dispositions + audit, Jev annotations (advisory), feature labels (secondary, re-derivable) |
| **Coverage** (existing) | What was examined, reused, skipped, blocked, and why | `(file or feature, family, category)` (`internal/domain/types.go` `CoverageKey`) | Blocks only as "required lane incomplete" (exit 2) | status, reason, reused-from run, inspected ranges manifest |

**How a Signal becomes an Issue without overclaiming (Q1).** Promotion happens only through a named, versioned *promotion rule* that states a concrete mechanism and a claim level:

- `delta` rules: a change in this PR moved a function/file across a stated threshold (e.g. function CC ≤ 10 → > 10; new clone group ≥ 50 tokens introduced; new import-cycle edge). Claim = *descriptive*; mechanism = "this change introduced X"; lane = report-only by default.
- `trend` rules: a per-feature signal worsened monotonically over N default-branch runs by ≥ Δ. Claim = *descriptive*; informational severity; nightly only.
- `policy` rules: the repository has opted into a boundary (e.g. `no cycles in internal/`, `no CC > 15`). Claim = *policy violation*; may be blocking once promoted under §7.4.

Every issue stores `claim ∈ {descriptive, policy, mechanism}` and a `promotion_rule@version`. No rule may set severity from an aggregate inertia score (E1 "Decision": "may be reported and trended, but not used as CI gates or severity inputs"). SCB's canonical components (erosion `1 − Σ_{CC>10} CC·√SLOC / Σ CC·√SLOC`, verbosity, cyclic edge mass, rework; `SCB/src/slop_code/metrics/scoring/production_quality.py:669-703,779-803`, `production_graph.py:222-235`) are stored as Scout metrics named `scout.erosion@1` etc., **not** labelled SCB-equivalent, because Scout's inputs (lizard NLOC, jscpd spans) differ from SCB's AST-statement/source-line contract (challenger §2 "§4.4", adopted; E1 "Deviations" 5–7 describe the lizard/jscpd inputs).

### 3.3 Issue identity and anchoring (Q2, identity half)

**Current defects (verified):**

1. `internal/issues/correlate.go:129` builds `correlationKey(..., files, features)` and `:333-340` folds `feature:<id>` into the key that `issueID` (`:326-329`) hashes → any change in owning feature IDs changes the issue ID. `sameLocationOrFeature` (`:214-219`) requires feature intersection whenever either side has features, even when files match.
2. `internal/state/issue_observations.go:94-102` `observationSupportKey` includes `StartLine/EndLine/StartByte/EndByte/RedactedQuote`; `internal/state/issues.go:455-461` treats any change as "supporting evidence changed", sets `IssueUnverified` and clears disposition/verification. Line movement above a finding therefore wipes an accepted/false-positive disposition.

**Design:**

- **Durable ID** = opaque `issue:<family>:<ulid>` minted once. Never recomputed from content.
- **Anchor set** (stored, updated each run): `{path (rename-followed via inventory diff), qualified symbol (when the tool or lizard provides one), rule/source key, normalized mechanism, normalized evidence fingerprint}` where the fingerprint is the redacted quote with whitespace/identifiers-of-line-numbers stripped. Line/byte ranges are *locators*, stored on the observation, excluded from matching.
- **Matching ladder** (deterministic, in order): (a) same rule + same path + same symbol + same fingerprint → match; (b) same rule + same path + same symbol, fingerprint changed → match, flag `evidence_drift` (keeps disposition, records drift in history); (c) same rule + same path, symbol unavailable, exactly one candidate in file → match; (d) same rule + renamed path (inventory `diff.go` rename record) → apply (a)–(c) on the new path; (e) otherwise → new issue; if ≥ 2 candidates satisfy (c) → both stay open, linked `ambiguous_with`, never auto-merged (req §12.6 "only unambiguous matches merge automatically" — kept).
- **Distinct occurrences** of the same rule in one file are distinct issues (challenger §6.1 adopted); today they can merge because the key has no symbol/range discrimination.
- **Feature labels** are secondary: `issue.feature_refs = [{clawpatch_id, scout_area_id, source: heuristic|agent, snapshot}]`, refreshed from the active map snapshot each run; a change here never touches state or disposition.
- **State transitions on re-observation**: match (a)/(c)/(d) → unchanged; (b) → unchanged + `evidence_drift`; matched issue in `resolved` → `regressed` (existing `issues.go:462-468` branch kept); mechanism changed (different rule) → new issue.

Migration: existing IDs are preserved as `legacy_id` aliases; the one-time migration re-anchors using observation paths/quotes and leaves ambiguous rows `unverified` with a history entry.

### 3.4 Lanes

A **lane** is a named, versioned unit of discovery with `{family, trigger ∈ {pr, push, nightly}, mode ∈ {blocking, report, signal}, requires ∈ {none, tool, provider-egress, decision-egress}}`. The engine reports lane *health* (ran / reused / skipped / blocked / failed) separately from lane *findings*. Required-lane health drives exit 2/3; findings drive exit 1 (§7.5). This replaces `assess.go:139-171`'s hard-coded `ClawpatchReview: true` and unconditional family/category expansion.

---

## 4. Feature-mapping strategy (Q2, mapping half)

### 4.1 Decision

- **Heuristic map is canonical** for ownership and for the stable feature vocabulary (E3 §2: heuristic Jaccard 1.00 across reruns, content edits and a new file).
- **Both mappers run on every authorised map refresh** (`map --source agent`), where a refresh is: first adoption; a scheduled default-branch run (nightly, or weekly — user's call, §15); or an explicit `scout map --refresh`. Ordinary PR and push runs **reuse the active snapshot** and never invoke the agent (E3 §4: agent remap 2–5 min, re-partitions the repo on every call; `auto` would also trigger it on fastapi every time, E3 §5).
- **Scout-side reconciliation** turns volatile agent output into stable *areas*: an `area` is a Scout record `{scout_area_id, title, file_set, origin: agent, first_seen_snapshot, clawpatch_ids: [...]}`. On each refresh, new agent features are matched to existing areas by file-set Jaccard ≥ 0.5 (threshold **[ASSUMPTION]**, to be set by X2); matched areas update their file set and append the new Clawpatch ID to provenance; unmatched become new areas; areas whose files are all now heuristic-owned are retired. Original Clawpatch IDs are preserved; Scout never rewrites Clawpatch's own records (req §3 "Do not implement a second Go mapper or silently translate its feature identity" is honoured: areas are *labels* over Clawpatch output, not a competing mapper, and are never used for issue identity).
- **Active snapshot membership**: ownership queries read only features present in the latest refresh's merged output, not every JSON ever persisted (E3 §1: state grows 187 → 296 features over five runs because Clawpatch never deletes).
- **Unowned files** after reconciliation are reported as a mapping-coverage gap (`coverage.mapping = unowned`) and still receive every file-level lane. Nothing waits on a map.
- **Failure semantics**: `agent` mode throws on zero valid agent features (`CP/src/agent-mapper.ts:165-168`); Scout treats that as "exploration lane incomplete", keeps the heuristic output, and never marks the run partial on PR triggers (the agent isn't invoked there).

### 4.2 Why not "always both on every run" (disagreement with brief §1/§4.6)

The user's underlying goal — "areas the deterministic mappers miss still get a definition issues can attach to" — is met by 4.1: every refresh uses both mappers, every file ends with a feature, an area, or an explicit `unowned` label, and issues attach to source anchors regardless. The literal reading (invoke the agent on every assessment, anchor to its IDs) is rejected on E3 §2–§3 evidence: 5–31 % agent-ID survival across fresh runs, 44–81 % of files change owner, a one-line edit yields ~20 unrelated new IDs, and the current correlator would mint new issue IDs on each change (§3.3). If the user still wants per-run agent invocation, the design tolerates it (areas absorb churn; identity ignores features), but its 2–5 min cost should be off the PR path and its measured value per refresh (new useful regions, actionable findings) must be reported (§8.4).

### 4.3 Rejected alternatives

- `auto` only (current `internal/clawpatch/client.go:178`): skips the agent on axum-like repos (E3 §5), so unowned areas never get a definition — fails the user's goal.
- Jev "assign unowned file to existing feature or none" as default: cheaper and deterministic-ID, but unmeasured; kept as experiment X3 to *augment* reconciliation for small unowned residues.
- Second Go mapper: forbidden by req §3 and unnecessary.
- Clawpatch upgrade: superseded. The 0.7.x review (`evaluation/clawpatch-0.7-compat/RESULTS.md`) was done and v0.7.3 adopted (§15 item 7); agent-ID churn is unchanged there, so §4.1 still applies.

---

## 5. Analyzer lanes and priorities (Q6 and lane design)

### 5.1 Lane catalogue (priority order)

| # | Lane | Family | Trigger | Mode (default) | Substrate | Languages at launch |
|---|---|---|---|---|---|---|
| 1 | `guardrails` — against-the-base: `CONFIG.ignores-added`, `CONFIG.strictness-lowered`, `SUPPRESS.added` (noqa / eslint-disable / `//nolint` / `#[allow]` / `@SuppressWarnings` / `#pragma warning disable`), `TESTQ.*` hollow tests, `GATE.coverage-pragma-added` (`FXL/docs/references/verify-gate.md:33-50`) | quality | pr, push | **blocking candidates** after §7.4 promotion; report until then | Go-native parsers for lint/type/test configs + diff of base/head; per-language suppression regex/tree-sitter | Go, Python (tier 1); TS/JS (tier 2); Rust/Java/C# (tier 3) |
| 2 | `measures` — function CC/NLOC (lizard), clone spans (jscpd), import graph + cycle mass, churn/rework from git | quality / architecture | push, nightly (full); pr (changed files only) | **signal**; `delta` promotion rules report-only | lizard, jscpd as pinned subprocess adapters (E1 "Deviations" 5–6 prove they run across all 7 corpus languages); import graph per language | all 7 for CC/NLOC/clones; import graph Go + Python first |
| 3 | `security` — Gitleaks; Semgrep maintained packs (`p/secrets`, `p/ci`, language `p/<lang>` security subsets) | security | pr, push | blocking candidates (secrets; mutable-action-ref, which is 6/9 corpus positives, E2 "Caveats" 3) | existing adapters (`internal/analyzers/gitleaks.go`, `semgrep.go`) | lang-agnostic |
| 4 | `rubric` — SCB erosion-type patterns decidable from a changed-function excerpt | quality | pr (changed functions), nightly sample | report; Jev-scored | Semgrep where structural; Jev Choice per candidate function where semantic (§6) | Python, Go |
| 5 | `architecture-policy` — declared boundaries (`.scout/policy.yaml`: layer rules, cycle prohibition, max CC) | architecture | pr, push | blocking after promotion | import graph + policy evaluator | Go, Python |
| 6 | `performance` — Ruff PERF, CSS rule (existing) | performance | push, nightly | report | existing | Python, CSS |
| 7 | `contextual` — LLM feature×category discovery (existing `internal/contextual`) | all | nightly, opt-in | report | provider adapter | n/a |
| 8 | `clawpatch-review` — per-feature review | cross-cutting | nightly, opt-in | report | existing `internal/clawpatch/review.go` | n/a |
| 9 | `artifacts` — read JUnit/coverage when present → changed-line-uncovered, regression breadth | quality | pr | report; `unmeasured` when absent | artifact ingestion with revision/producer match | any producing LCOV/Cobertura/JUnit |

### 5.2 Rubric lane scope (disagreement with brief §4.5)

The 45-pattern cross-product per changed function is rejected. `SCB/configs/rubrics/llm_judge.jsonl` lines 19, 20, 43 (`unrequested_features`, `handling_excluded_cases`, `validation_mismatch`) require a spec; line 27 (`boolean_success_flags`) penalises Go's `(result, ok)` idiom; line 6 (`lonely_interface`) needs whole-repo evidence. Launch set (12, erosion-type, excerpt-decidable, language-neutral): `silent_fallbacks`, `retry_loops_hiding_failures`, `error_message_destruction`, `validation_theater`, `stub_implementations`, `scaffold_remnants`, `lenient_parsing`, `silent_truncation`, `auto_correction_without_warning`, `log_level_hiding`, `sentinel_values_as_errors`, `ignoring_return_values` (language-aware). Cost: 12 × ~$5e-5 ≈ $0.0006 per changed function (E2 cost), i.e. ≈ $0.12 per 200-function PR instead of brief §4.5's ≈ $0.90. Precision unmeasured → report-only, experiment X4.

### 5.3 Rule-pack decision (brief §4.7, partial disagreement)

Replace corpus-tuned rules (`internal/analyzers/rules/security/semgrep-python-multipart-cve-2024-24762.yaml`, `semgrep-javascript-current-script-dom-clobbering.yaml`, `rules/performance/semgrep-css-height-transition.yaml`) as *capability claims*; keep them as labelled regression detectors in `selfeval` fixtures until maintained packs demonstrably cover the same positives and negatives (challenger §2 "§4.7", adopted). Keep `semgrep-actions-mutable-ref.yaml` and `semgrep-python-unsafe-yaml-load.yaml` (general mechanisms).

### 5.4 Language strategy (Q6)

**Depth first on Go and Python; breadth where the tool is language-independent.** Launch capability matrix (published by `scout readiness --matrix`):

| Capability | Go | Python | TS/JS | Rust | Java | C# |
|---|---|---|---|---|---|---|
| CC/NLOC/clones signals | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Secrets / CI security | ✓ all | | | | | |
| Import graph / cycles | P1 | P1 | P2 | P3 | P3 | P3 |
| Guardrails (config/suppression/test quality) | P1 | P1 | P2 | P3 | P3 | P3 |
| Rubric lane | P1 | P1 | P2 | – | – | – |
| Semantic review (contextual) | opt-in, any | | | | | |

Rejected: seven-language equality as acceptance (req §5 line 58); wrapping Python SCB (Python-only metrics code, brief §3); porting lizard's parsers to Go (challenger §Q6: adapters over porting — adopted; `internal/analyzers/runner.go` already provides the bounded subprocess contract).

---

## 6. Placement of deterministic code, decision models, reasoning LLMs, humans (Q3)

| Decision | Owner | Why |
|---|---|---|
| Eligibility, dependency closure, cache keys, identity/matching, gate evaluation, quote validation, redaction | **Deterministic Go** | Must be reproducible and adversary-proof; current code already does most (`internal/inventory`, `internal/analyzers/sanitize.go`, `internal/contextual/validate.go`) |
| Triage ranking of deterministic findings (`P(real_defect)`), rubric-pattern presence per candidate function, commit classification (fix/revert/feature) for history signals, unowned-file → area suggestion, severity-rubric closed dimensions | **Jev (typed decisions) — advisory only** | E2: AUROC .93–.99, $0.05/1k, 0.3 s p50, 0/396 failures; vs. Opus 21/22 at ≈ $55–65/1k, 8–9 s |
| Finding discovery in context, new feature grouping, FN discovery, second opinion on Jev-uncertain high-severity items | **Reasoning LLM — opt-in** | E2 "LLM baseline": accuracy parity on this set, 1000× cost; discovery accuracy unmeasured |
| Dispositions, policy adoption, blocking-rule promotion, area acceptance | **Humans** | brief §3 "Keep human" |

**Guardrails on the decision adapter (`internal/decisions`)** — a second adapter contract distinct from req §10/§12.5's JSON LLM provider:

1. Typed questions only (`noul | choice | score`), versioned question text (`question_id@version`), bounded redacted state built by the same sanitizer as `contextual`.
2. **Route/rank only.** The adapter's output is an `annotation` row `{issue_id, question_id@version, model_requested, model_resolved, probabilities, cost, latency, run_id}`. No code path may change `issue.state`, severity, gate result, coverage status, or merge issues from an annotation. Enforced by type: annotations live in their own table and `issues` package has no import of `decisions`.
3. **Resolved-model pinning**: `model_resolved ≠ model_requested` → annotation discarded, unit marked `incomplete`, cache key invalidated (brief §3 "Safety rule").
4. **Budget**: per-run decision cap (`policy.decisions.max_per_run`, default 2,000 ≈ $0.10) and per-PR cap; exceeding → remaining items `unranked`, disclosed.
5. **Unranked view always available**; the CLI default `scout issues` sorts by severity then recency; `--rank jev` opts into ranking. 5 % random sample of low-ranked items is surfaced in nightly summaries for audit (challenger: ranking ≈ suppression).
6. **Deterministic ranking baseline** (`rule weight × changed-in-PR × path role (test/generated/docs/workflow)`) is computed for every finding and stored beside the Jev annotation so X1 can measure Jev's incremental retrieval gain.
7. **Prompt-injection posture**: state is data; Jev-1.13 weaknesses (indirection, numbers, adversarial content in state; brief §3) mean an attacker can at most *lower a finding's rank*, never close it — and the audit sample bounds that.
8. **Eval-set growth**: every human disposition on an annotated issue becomes a label `(state digest, question_id@version, model, P, disposition)`; `scout selfeval triage` reports per-rule AUROC/Wilson intervals over labels not seen by the question author (blind by construction after launch). Jev ranking is enabled by default only when X1's rule passes.

Rejected: Jev as identity oracle; Jev closing `false_positive` automatically (E2 "Recommendation": "never to close, dismiss or suppress"); Noul as primary question type (E2: more `cannot_tell`, weaker agreement); treating `P` as calibrated confidence (E2 Caveat 6).

---

## 7. CI guardian contract (Q4, Q5)

### 7.1 Triggers

`pull_request` (differential, blocking profile optional), `push` to default branch (full or incremental; publishes baseline), `schedule` nightly (full; map refresh; trend rules; contextual/review lanes if authorised), `workflow_dispatch` (`full`, `refresh-map`, `authorize-egress`). Local: `scout assess`, `scout diff --base <ref>`.

### 7.2 Baseline storage (Q5)

**Decision: a protected, dedicated git state branch (`scout/state`) holding a versioned, sanitized logical export; SQLite stays the local transactional store; workflow artifacts hold bulky run evidence; S3 is an optional adapter, not a prerequisite.**

- Export format: `state/v1/{issues.jsonl, dispositions.jsonl, signals/<metric>.jsonl, areas.jsonl, coverage-summary.json, manifest.json}` — deterministic ordering, one record per line, so diffs are reviewable and merges are append-friendly. Import rebuilds SQLite; never commit the `.sqlite`/WAL files (`scripts/bundle-state.sh` tars the live state dir today).
- Writers: only trusted `push`/`schedule`/`dispatch` runs on the default branch publish, with expected-parent (compare-and-swap) semantics; concurrent publishers retry on top of the new tip. PR and fork runs are read-only.
- Readers: PR runs fetch the branch tip *and* resolve the base-revision coverage for the merge-base; if the tip's tool/rule/policy digests differ from the PR's, the engine falls back to a paired base/head deterministic scan (§7.3) and reports `comparison: recomputed`.
- Why not S3/OIDC: current workflow needs four AWS variables plus OIDC and exits 3 without them (`.github/workflows/scout-nightly.yml` "Validate durable baseline configuration", "Restore the durable baseline") — the main adoption blocker. Why not committed-in-tree snapshot: every PR would conflict on it and PR bots could forge dispositions. Why not Actions cache/artifacts alone: eviction and 90-day retention (`upload-artifact ... retention-days: 90`) violate req §10 "no best-effort cache as the only baseline". **[ASSUMPTION]** a protected branch is lower friction than object storage for the target repos; if branch protection cannot be granted, the S3 adapter is the alternative and never a silent downgrade to cache.

### 7.3 Gate semantics

- **Attribution** per issue on a PR run: `introduced` (absent in base under the same tool/rule/policy digests, present in head), `regressed` (matched a `resolved` issue), `pre-existing` (present in base), `unknown` (base not comparable: digests differ, first adoption, lane skipped on base). Only `introduced` and `regressed` can block. `unknown` reports with a banner (challenger Q4, adopted).
- **Blocking lanes** (default profile `guardian`): the allow-list in `.scout/policy.yaml` seeded with `security.secrets`, `security.ci-mutable-action-ref`, `guardrails.CONFIG.*`, `guardrails.SUPPRESS.added`, `guardrails.GATE.coverage-pragma-added`. Every entry must have passed the promotion rule (§7.4). `architecture-policy` rules block only once the repo declares them. All `measures`, `rubric`, `contextual`, `performance`, `artifacts` findings are report-only (E1).
- **Report-only profile** (`report`) is the adoption default: findings never fail CI; exit codes 2/3 surface as neutral check + summary.
- Pre-existing debt: counted and listed, never blocks (brief §4.1).
- Optional lanes can never turn a required lane red; a required blocking lane that could not execute on head yields exit 2 (not pass).

### 7.4 Promotion rule for blocking lanes

A rule may enter the blocking allow-list when: all `selfeval` safety fixtures for it pass (positive, clean negative, false-positive boundary, redaction); blocking precision on ≥ 50 independently double-labelled, unseen PR findings has Wilson 95 % lower bound ≥ 0.95; and no known systematic negative class remains. Thresholds are proposals **[ASSUMPTION]**; no current rule is asserted to qualify.

### 7.5 Exit codes and outputs

Keep `0 clean / 1 gated / 2 incomplete / 3 unavailable` (req §12.7; `cmd/scout/main.go`). Precedence: 3 > 2 > 1 > 0 for the process; JSON always contains both `gate` and `health` blocks so a gated-and-incomplete run is not hidden. Outputs: `scout-assessment.json`, a ≤ 60-line Markdown summary (new/regressed first, then pre-existing count, then lane health, then signal deltas), SARIF for code-scanning upload (optional).

### 7.6 Adoption path and trust boundary

- Reusable workflow `scout/.github/workflows/scout-guardian.yml` (`workflow_call`) plus a composite action; adopter adds a 15-line workflow with `profile: report|guardian`, `languages`, `baseline_branch`.
- Scout binary and Clawpatch bundle come from a **pinned release** (`release.yml` already exists), never built from the PR checkout. Policy is read from the *base* branch, not the PR head, for blocking decisions.
- Tokens: `contents: read` for PR; `contents: write` restricted to the state branch for trusted runs; provider/decision secrets never exposed to `pull_request` from forks.
- First adoption: `scout init` → full run on default branch → publishes baseline; nothing is `introduced` until a baseline exists.

---

## 8. Scout's own outcome and self-evaluation metrics (Q7)

A **discovery-outcome scorecard**, each line with `count/attempts`, Wilson 95 % interval, and `unmeasured` counts (FXL `docs/references/scoring.md` discipline; brief §3). No single "solve rate".

1. **Actionable yield** per lane/rule: issues dispositioned `accepted`/`resolved`-by-change ÷ dispositioned issues; `open` counted separately as undecided.
2. **Follow-through**: confirmed issues resolved (verified) within 30/90 days ÷ confirmed issues with a mature window. Measures uptake, not causal benefit.
3. **Scoped assurance** (brief §4.8, scoped as the challenger insisted): for a random sample of `(file, lane)` units Scout called clean, an independent audit (LLM-assisted human) finds an in-scope miss or not → false-assurance rate per lane; plus flag precision/recall on that sample (FXL assurance matrix, brief §3). Repository-wide FN rates are not claimed.
4. **Cost of use**: false blocks (blocked PRs whose blocking issue was later `false_positive`), duplicate/re-opened issues per 100, median time-to-disposition, runtime and provider/decision spend per PR, agent-refresh value (new reviewed regions, actionable findings per refresh).
5. **Change-success rate (experimental, descriptive)**: per feature/area, `changes with no follow-up fix/revert touching the same lines within N days ÷ changes`, computed only when the commit classifier meets X5's rule; shown with `claim: descriptive` and never as a Scout-value headline (E1 classifier precision 0.52; challenger Q7 "quiet feature looks excellent"). This honours the user's liking for SCB's solve-rate framing while keeping to what the oracle supports (brief §1, §4.3).
6. **Triage-model quality**: per-rule AUROC and routing precision of Jev vs. the deterministic ranking baseline on post-launch blind labels (§6 item 8).
7. **Corpus regression**: the 9 adjudicated true issues and 13 FPs (`evaluation/reference-v1.json`) stay detectable/non-flagged under the current rule set — a regression check, not recall.

---

## 9. Keep / cut / rewrite (Q8)

### 9.1 Code

| Package / file | Verdict | Notes |
|---|---|---|
| `internal/inventory/{scan,diff,plan}.go` (+tests) | **Keep** | Finer than Atlas's HEAD cache (brief §3). Change: deterministic file lanes must not depend on `MappingRequired`; add base/head paired-plan entry point. |
| `internal/state/{store,artifacts,progress}.go` | **Keep** | Add `export`/`import` of the logical snapshot (§7.2). |
| `internal/state/{issues,issue_observations}.go` | **Rewrite identity-sensitive parts** | Remove position/quote from `observationSupportKey` (`issue_observations.go:94-102`); replace `supportingEvidenceChanged` wipe (`issues.go:455-461`) with `evidence_drift` history; keep audit, verification, `regressed` transition; drop `ApprovalRef`/`ProposedAction` fields with migration. |
| `internal/issues/correlate.go` | **Rewrite matching** (§3.3) | Keep `SanitizeText`/provenance handling. |
| `internal/domain/types.go` | **Keep; extend** | Add `Signal`, `Area`, `Annotation`, `Attribution`, `LaneHealth`; remove approval types. |
| `internal/engine/{assess,readiness,config}.go` | **Rewrite orchestration** | Lane registry, base/head comparison, profile/policy loading; readiness computed from the same lane plan. |
| `internal/analyzers/{runner,spec,parse,sanitize,ruff,semgrep,gitleaks}.go` (+tests) | **Keep** | Runner is the subprocess contract for lizard/jscpd too. |
| `internal/analyzers/catalog_{security,quality,performance,architecture}.go` | **Re-curate** | Quality: lizard/jscpd specs; architecture: import-graph + policy evaluator (currently `return nil`); security: maintained packs; corpus-tuned rule files → `selfeval` fixtures (§5.3). |
| `internal/clawpatch/{client,records,review}.go` | **Keep boundary; rewrite scheduling** | `Map` gets explicit `source` and `refresh` semantics (today hard-coded `auto`, `client.go:178`, and blocked whenever egress is unauthorised even for heuristic, `client.go:168-170` — heuristic must run offline). Review stays opt-in. |
| `internal/contextual/*` | **Keep; demote to opt-in nightly lane** | Reuse its sanitizer/manifest for `decisions`. |
| `internal/decisions` (new), `internal/measures` (new), `internal/guardrails` (new), `internal/baseline` (new) | **Create** | §§5–7. |
| `internal/eval/{corpus,outcomes}.go` | **Rename to `selfeval`; rewrite success criteria** | Keep miss/FP/ambiguous/incomplete distinctions; add scorecard (§8). |
| `internal/web/web.go` | **Keep** read-only | Add signals/areas views after CLI stabilises. |
| `internal/codex/exec.go` | **Cut** | No callers (brief §2). |
| `cmd/scout/main.go` | **Rewrite gate + add commands** | `--fail-on` severity gate (`main.go:202-207`) → profile/attribution gate; add `diff`, `map --refresh`, `export/import`, `selfeval`, `readiness --matrix`; remove `approve`. |
| `.github/workflows/scout-nightly.yml` | **Rewrite** into `scout-guardian.yml` (reusable) + self-dogfood caller | Drop AWS/OIDC prerequisite; add `pull_request`; partial ≠ failure in `report` profile. |
| `scripts/bundle-state.sh` | **Replace** by `scout export/import` | Tarring live WAL state is fragile. |
| `scripts/build-clawpatch.sh`, `clawpatch-safe.patch` | **Keep** | Pinned v0.7.3 + ported redaction patch (`evaluation/clawpatch-0.7-compat/`). |

### 9.2 The 62 KB spec

Replace with three documents (target ≤ 20 KB total):

- `docs/scout-product-contract.md` — §§1, 3.2–3.4, 7 of this design (product, record kinds, lanes, CI contract, exit codes). Rewrites req §§1, 2, 5, 9, 10, 12.7.
- `docs/scout-adapter-contracts.md` — tool adapter (keep req §5 contract text + §12.1 Clawpatch command contract, amended for `--source agent` refresh), **new** typed-decision adapter (§6), LLM provider (req §12.5, trimmed), identity/matching (req §12.6 rewritten per §3.3), incremental planning (req §12.4 kept nearly verbatim).
- `docs/evidence/` — req §§11.1–11.2, 12.2 (Enaible inventory), 13 moved as historical appendices; `evaluation/*/RESULTS.md` stay where they are.

Cut outright: req §7 approval/implementation-plan paragraphs, §12.3 mandatory feature×category coverage matrix (becomes the opt-in `contextual` lane's own ledger), four-equal-families framing, Enaible parity as acceptance, S3 as the baseline. Keep req §8 (gap→rule promotion, with FXL's FCR record shape: detection signal, evidence, hypothesis, measurable resolution criterion, status) and §12.8's fixture classes (positive, clean negative, FP, missing binary, ineligible, tool failure, redaction, evidence drift, contextual miss).

---

## 10. Phased plan — every phase ends in a runnable scenario

Phases are sequenced by dependency; each lists acceptance evidence the main agent can observe. No tests are claimed to pass today.

**Phase 1 — Identity and offline ledger** (`issues`, `state`, `domain`, `clawpatch.Map` offline-heuristic)
Scenario: on a fixture repo with no provider credentials, run `scout assess` twice (second after inserting 20 lines above a flagged function, renaming a file with a finding, and duplicating a same-rule finding in one file); disposition one issue `accepted`; resolve one by deleting the code and verify; re-introduce it.
Acceptance: one durable ID across line shift and rename; `evidence_drift` history entry, disposition intact; two distinct IDs for two occurrences; re-introduced issue → `regressed`; heuristic map ran with egress off; `go test ./internal/issues ./internal/state` green; a migration test re-anchors a v1 store and lists ambiguous rows.

**Phase 2 — Differential guardian with git-branch baseline** (`baseline`, `engine` comparison, `cmd` gate, reusable workflow)
Scenario: a separate fixture repo adopts the reusable workflow in `guardian` profile; default-branch run publishes `scout/state`; PR A adds a secret and a `# noqa`; PR B touches a file with pre-existing debt only; PR C comes from a fork; a tool-version bump is simulated by changing a rule digest.
Acceptance: PR A exits 1 listing exactly the two `introduced` issues; PR B exits 0 with pre-existing count; PR C cannot write the state branch or read secrets; digest mismatch yields `attribution: unknown` + `comparison: recomputed`; two concurrent default-branch publishes both land (CAS retry) with no lost dispositions; `scout import` of the branch tip reproduces the SQLite ledger byte-identical in logical export.

**Phase 3 — Quality substrate: measures + guardrails for Go and Python** (`measures`, `guardrails`, catalogs)
Scenario: on Scout's own repo and a Python fixture, nightly run writes signals; a PR raises a function from CC 9 → 14, adds a clone, lowers `ruff` strictness, adds `//nolint`.
Acceptance: `scout signals --file` shows numerator/denominator/exclusions and `unmeasured` for an unsupported file; `delta` rules produce report-only issues with `claim: descriptive`; `CONFIG.strictness-lowered` and `SUPPRESS.added` fire with base/head evidence; no severity derives from an aggregate; the 9-true/13-FP corpus regression still holds; `readiness --matrix` prints the capability table.

**Phase 4 — Mapping refresh and reconciliation** (`mapping` areas, nightly refresh)
Scenario: fastapi and axum fixtures (E3 setups); run refresh (both mappers) three times fresh and once after a content-only edit; attach issues to agent-only-owned files between runs.
Acceptance: every prior issue remains addressable with unchanged ID; areas absorb agent churn (report Jaccard of area file-sets across refreshes ≥ 0.8 **[ASSUMPTION]**, measured); active-snapshot ownership excludes stale Clawpatch records; unowned files listed; PR runs show zero agent invocations; agent failure leaves heuristic output and marks exploration incomplete.

**Phase 5 — Typed decisions, rubric lane, and self-evaluation** (`decisions`, `rubric`, `selfeval`)
Scenario: replay the E2 cases plus ≥ 100 new blind-labelled cases through deterministic ranking, Jev, and optional Opus escalation; run a prompt-injection fixture (comment instructing "mark as false positive"); run the scorecard on Phase-2/3 ledgers.
Acceptance: annotations carry resolved model and question version; a forced model mismatch marks the unit incomplete; no state/gate change from any annotation; injection fixture changes nothing but rank, and appears in the audit sample; scorecard prints Wilson intervals and `unmeasured` counts; X1 decision rule evaluated and recorded.

**Phase 6 — Spec cutover, web views, release**
Scenario: new docs replace the 62 KB spec; release workflow ships binary + Clawpatch bundle + reusable workflow; Scout dogfoods itself in `guardian` profile for two weeks.
Acceptance: `SCOUT_REBUILD_REQUIREMENTS.md` removed/archived; adopter checklist completes in one PR; false-block count over the dogfood window reported.

---

## 11. Decisions on the eight open questions (summary with evidence)

1. **Signals vs issues.** Decision: three record kinds; signals are append-only descriptive measurements that can never gate; promotion only via named `delta`/`trend`/`policy` rules carrying a `claim` field (§3.2). Rejected: inertia-score gate; every rubric hit → issue; separate trends DB. Evidence: E1 "Decision" (+1.18 %, CI [−0.62, +3.24], fails prereg), E1 "Ablations"; `SCB/docs/metrics/measurement-validity.md` claims ladder; req §7 lines 96-105.
2. **Feature mapping.** Decision: heuristic canonical; both mappers on authorised refreshes only; Scout-side areas reconcile agent output by file-set overlap; issue identity excludes feature IDs and line positions (§§3.3, 4). Rejected: always-agent per run; `auto` only; second mapper; Jev-assign as default. Evidence: E3 §§1–5; `CP/src/agent-mapper.ts:129-180`, `CP/src/mapper.ts:223`; `internal/issues/correlate.go:129,214-219,333-340`; `internal/state/issue_observations.go:94-102`, `issues.go:455-461`.
3. **Decision models.** Decision: separate typed adapter; route/rank only; resolved-model pinning; budget cap; unranked view + audit sample; deterministic ranking baseline; dispositions become labels (§6). Rejected: Jev closes/suppresses; Jev as identity oracle; Noul primary; calibrated-confidence assumption. Evidence: E2 tables (AUROC .93–.99; $0.05/1k; 0/396 failures), E2 "Caveats" 1–6, "Recommendation", "Unlabelled sample"; brief §3 Jev weaknesses.
4. **Gate.** Decision: block only `introduced`/`regressed` issues from a promoted allow-list (secrets, CI mutable refs, config/suppression/pragma guardrails; policy-declared architecture rules); everything quality-descriptive reports; `report` profile is the adoption default; exit 0/1/2/3 with health precedence (§7.3–7.5). Rejected: gate all unresolved ≥ severity (current `cmd/scout/main.go:202-207`); fail on partial by default (current workflow); block new cycles without policy. Evidence: E1; `FXL/docs/references/verify-gate.md:33-50`; `.github/workflows/scout-nightly.yml`.
5. **Baseline storage.** Decision: protected `scout/state` git branch with deterministic JSONL export, trusted-writer CAS publish, artifacts for bulk evidence, optional S3 adapter (§7.2). Rejected: S3/OIDC prerequisite; in-tree snapshot; cache/artifacts as sole store; tarring live WAL. Evidence: workflow AWS steps; `upload-artifact retention-days: 90`; req §10 "no best-effort cache as the only baseline"; `scripts/bundle-state.sh`.
6. **Languages.** Decision: Go + Python deep (guardrails, import graph, rubric); all seven get language-independent measures (lizard/jscpd) and secrets/CI security; capability matrix published (§5.4). Rejected: seven-language parity at launch; porting parsers; wrapping Python SCB. Evidence: E1 "Deviations" 5–6 (lizard/jscpd ran on all 7); `internal/analyzers/catalog_quality.go:9-22` (Python-only today), `catalog_architecture.go` (nil); `FXL` guardrails are Python-only.
7. **Outcome metric.** Decision: scorecard — actionable yield, follow-through, scoped assurance, cost of use, triage-model quality, corpus regression; change-success kept experimental behind X5 (§8). Rejected: single solve-rate headline; predictive claims. Evidence: E1 "Commit classifier validation" (P 0.52 / R 0.92); `FXL/docs/references/scoring.md` (count/attempts, Wilson, unmeasured ≠ failed, assurance matrix); SCB claims ladder.
8. **Scope.** Decision: refactor in place per §9; cut codex, approvals, S3, corpus-tuned capability claims, mandatory feature×category matrix; split spec into product contract + adapter contracts + evidence appendix. Rejected: greenfield rewrite; delivering the full 62 KB spec first. Evidence: brief §2 table (what is real vs. never-run), `internal/engine/assess.go:139-171` hard-coded lanes, `internal/codex/exec.go` no callers.

---

## 12. Top risks

| # | Risk | Likelihood / impact | Mitigation / reversibility |
|---|---|---|---|
| R1 | Identity rewrite still mis-merges (copies share fingerprints; edited defects change them) | medium / high — wrong dispositions carried or lost | Matching ladder keeps ambiguity explicit; Phase 1 fixtures (rename+edit, copy, two same-rule sites, mapper churn); legacy IDs retained → reversible |
| R2 | Gate fatigue from false blocks in `guardrails` | medium / high — adopters switch to `report` and ignore | Promotion rule §7.4; `report` default; false-block metric §8.4 |
| R3 | Agent refresh adds cost without discovered value | medium / medium | §8.4 agent-refresh value metric; cadence is config; E3 has wall time but no cost data — measure before claiming |
| R4 | Jev ranking hides real findings (adversarial comments, indirection) | low–medium / medium | Route-only, unranked view, audit sample, X1 vs. deterministic baseline |
| R5 | State-branch permissions unavailable or history bloat | medium / medium | S3 adapter fallback; export is JSONL so history squashing is safe |
| R6 | Metric drift mistaken for quality change (tool upgrades, feature re-partition) | medium / medium | Signals keyed by `metric@version` + tool digest; re-partition never rewrites history; areas vs. features distinction |
| R7 | Guardian supply chain: PR controls its examiner | low / high | Pinned release, base-branch policy, read-only PR tokens (§7.6) |
| R8 | Spec/doc cutover leaves contradictory sources | medium / low | Phase 6 archives the old spec; design doc links replace it |
| R9 | Clawpatch pin ages (upstream already at 0.8.2); redaction patch stops applying | medium / medium | `build-clawpatch.sh` verifies patch; each upgrade needs a compatibility review + reduced E3 re-run, as done for 0.7.3 |

---

## 13. Remaining experiments with decision rules

| ID | Question | Design | Decision rule |
|---|---|---|---|
| X1 | Does Jev ranking beat deterministic ranking on triage retrieval? | ≥ 100 de-correlated, blind-authored labelled signals (≥ 30 semantic positives, ≥ 20 injection/indirection negatives); compare actionable recall at fixed review budget (top-k) | Enable Jev ranking by default iff held-out recall@k gain ≥ 10 points absolute with Wilson lower bound > 0 and no critical-severity true issue drops below the audit floor; else keep `--rank jev` opt-in |
| X2 | Do Scout areas stabilise agent output? | Re-run E3 protocol with reconciliation on fastapi/axum, 3 fresh + 2 same-state + 2 mutations | Adopt file-set Jaccard threshold that yields area-ID survival ≥ 0.8 across fresh refreshes and ≤ 5 % spurious merges on hand-checked pairs; otherwise reduce refresh cadence to onboarding + manual |
| X3 | Can Jev assign unowned files to existing features? | Choice "existing feature or none" on 200 unowned files with hand labels | Use as reconciliation aid iff precision ≥ 0.9 on `none`-exclusive decisions; else drop |
| X4 | Rubric-lane precision per pattern | 12 patterns × 300 changed Go/Python functions from corpus history; double-labelled sample of hits | Keep a pattern in the lane iff flag precision Wilson lower bound ≥ 0.6; promotion to blocking never (descriptive claim) |
| X5 | Jev commit classifier vs regex | 200 blind-labelled commits across 7 repos, diff-aware state | Use for change-success signal iff precision ≥ 0.85 and recall ≥ 0.8 (Wilson lower bounds ≥ 0.75/0.7); else change-success stays unpublished |
| X6 | Guardrail rule false-positive rate on real history | Replay last 200 merged PRs of 3 adopter repos through `guardrails`; double-label hits | Promote to blocking per §7.4 (lower bound ≥ 0.95); else report-only |
| X7 | Predictive validity re-test (optional, post-launch) | E1 protocol with X5 classifier, feature-level units, centrality control | Only a pass of E1's preregistered rule (≥ 5 %, lower > 0, ≥ 5/7 repos) upgrades any quality signal's claim to `predictive`; severity/gate use still requires a separate decision |

---

## 14. Explicit disagreements with brief §4 and user preferences

- **§1 / §4.6 "always run both mappers"** — disagree with per-run agent invocation and with anchoring to agent IDs; agree with "both on every refresh" and "every file gets a definition" (§4.2). Evidence: E3 §§2–4; `correlate.go:333-340`.
- **§4.1 "gate only on new/regressed"** — insufficient alone; add same-policy attribution (§7.3). Evidence: workflow and tool-digest reasoning; E3 churn.
- **§4.3 change-success as the solve-rate analogue** — keep only as experimental descriptive signal (§8.5). Evidence: E1 classifier P 0.52; survivorship/centrality threats.
- **§4.4 reimplement SCB formulas in Go** — implement the arithmetic, but name them Scout metrics; do not claim SCB equivalence (§3.2). Evidence: SCB production_quality vs E1 lizard/jscpd inputs.
- **§4.5 45 patterns × changed functions** — rejected; 12-pattern language-neutral subset (§5.2). Evidence: `llm_judge.jsonl` lines 6, 19, 20, 27, 43.
- **§4.7 replace corpus rules wholesale** — replace as capability claims, keep as regression detectors (§5.3).
- **§4.8 false-assurance rate** — adopt only scoped, sampled version (§8.3).
- **§4.9 read CI artifacts** — agree, with provenance matching; cannot substantiate FXL's "base tests against head" claim (§5.1 lane 9).
- **Challenger's "no incremental cache"** — not adopted; planner kept, paired scan as fallback (§2).

---

## 15. Decisions for the user — status (updated 2026-10-19)

1. **Map-refresh cadence — DECIDED:** both mappers run at onboarding, then on events where monitored areas change (new unowned files, new/removed manifests or entrypoints, structural changes to owned paths). Triggers fire on default-branch merges only, never PR runs, and are debounced into one refresh. The "areas" reconciliation layer stays (agent-ID churn persists per E3 and the 0.7.3 review). Supersedes §4.1's nightly default.
2. **Baseline backend — DECIDED:** protected `scout/state` git branch. S3 is deferred; add it as an optional adapter only once the branch backend is proven.
3. **Blocking allow-list — DECIDED:** adopt `docs/scout-allowlist.md` (eligible-not-active list; two-track promotion replaces §7.4). Per-repo shadow window: minimum 14 days **and ≥ 5 PR runs** with comparable base, both configurable in `.scout/policy.yaml` (the memo's 20 is superseded). Sub-decisions taken as conservative defaults [ASSUMPTION, confirm]: overrides (`scout-override` label / `scout:allow`) require a maintainer, not the PR author; rare security checks may activate `provisional: true` on track A alone.
4. **Jev default — DECIDED:** on from launch, route/rank only; never changes issue state or the gate.
5. **Language tiers — DECIDED:** tier 1 is Go, Python and TypeScript.
6. **Spec cutover — DECIDED:** archive `SCOUT_REBUILD_REQUIREMENTS.md` in favour of the §9.2 documents.
7. **Clawpatch version — DONE:** upgraded to v0.7.3, pinned by commit `734f81bc1514b52a0627edb0c24ebe1a18b36105` (`c46c0c6` is the annotated tag object, not the commit) with the ported redaction patch, per `evaluation/clawpatch-0.7-compat/RESULTS.md` §7. 0.8.x is a later, separate review.

Related research added after this memo: `docs/inertia-research.md` (consequences of inertia, new candidate measures, agent-era hypothesis) and the planned E1b study of change cost at area level.

---

## Appendix A — Assumptions register

- A1 Area file-set Jaccard 0.5 match / 0.8 survival targets (§4.1, Phase 4) — to be set by X2.
- A2 Blocking promotion threshold 0.95 lower bound on ≥ 50 labels (§7.4).
- A3 Protected state branch is lower friction than S3 for target repos (§7.2).
- A4 Deterministic path/rule ranking captures much of Jev's apparent E2 gain (§6 item 6) — X1.
- A5 Rubric-lane cost ≈ $0.12 per 200-function PR extrapolated from E2 unit cost; precision unknown (§5.2) — X4.
- A6 Default decision budget 2,000/run (§6 item 4).
- A7 Naive application of SCB idiom patterns penalises Go/TS idioms (§5.2) — inferred from rubric text, not measured.
