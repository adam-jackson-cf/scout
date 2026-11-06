# Scout — design and planned approach, version 2

Status: council decision memo v2 (2026-10-19). Supersedes v1 (2026-10-17). v2 folds in the user's decisions (Appendix, authoritative, not reopened), the inertia literature review, the E1b change-cost study, the Clawpatch 0.7.3 upgrade and a P1 performance defect found during the upgrade smoke. An independent adversarial analysis by a different model family (`InertiaChallenger`, transcript `history://ScoutDesignV2.InertiaChallenger`) was developed in parallel on the inertia section, the defect and the plan; §2 records where it changed the recommendation. Nothing here is implemented; every acceptance scenario is future work.

Evidence notation. `brief §n` = `docs/scout-design-brief.md`; `req §n` = `SCOUT_REBUILD_REQUIREMENTS.md` (to be archived); `AL §n` = `docs/scout-allowlist.md`; `IR §n` = `docs/inertia-research.md`; E1 = `evaluation/inertia-v1/RESULTS.md`; E1b = `evaluation/change-cost-v1/{PREREGISTRATION,RESULTS}.md`; E2 = `evaluation/jev-triage-v1/RESULTS.md`; E3 = `evaluation/map-stability-v1/RESULTS.md`; CP07 = `evaluation/clawpatch-0.7-compat/RESULTS.md`; `SCB` = `/Users/adamjackson/Projects/inflight/slop-code-bench`; `FXL` = `/Users/adamjackson/Projects/inflight/factory-x-loops/factory-x-loops`. Scout paths are repository-relative. Claims without a cited file or measurement are tagged **[ASSUMPTION]**. Scout never reads, writes or deletes `.codegraph`; it is read-only on the assessed code.

---

## 1. Product definition and non-goals

**Scout is a discovery-only CI guardian for a single repository.** It runs on pull requests, default-branch pushes, nightly and locally; it surfaces evidence-backed issues and descriptive measurements into a durable, triageable ledger; it blocks a PR only for a short, mechanism-certified allow-list of checks that the repository has activated and that the PR itself introduced or re-introduced (AL). **Code quality as the driver of long-term codebase inertia is the primary family**: Scout monitors, per repository and per scope, the structural *principal* and the observed change-cost *interest* over time (§8). Architecture, security and performance remain families with lower priority.

- **Ledger**: stable issue IDs anchored to source (never to feature IDs or line numbers), audited dispositions, sanitized evidence, append-only signals with history.
- **Guardian**: differential gate (head vs. same-policy base) on `introduced`/`regressed` issues in *activated* checks; everything else reported, including `would_block` shadow annotations.
- **Mapping**: Clawpatch 0.7.3 heuristic map is the canonical feature vocabulary; the agent mapper runs at onboarding and on debounced default-branch structural events; Scout-side *areas* absorb agent-ID churn; agent IDs never enter issue identity.
- **Inertia monitoring**: per-scope time series of principal and interest, `claim: descriptive`, trend direction reported, never gated; the ledger accumulates what a future pooled test needs.
- **Judgment placement**: deterministic code decides eligibility, identity, gates and measurements; Jev ranks and routes (on from launch, deterministic fallback); reasoning LLMs are opt-in discovery; humans own dispositions, policy and activation.

**Non-goals.** (1) No remediation; req §7's approval machinery is cut. (2) No execution of the target's tests; CI artifacts are *read* when provenance-matched. (3) **No inertia score** as gate, severity input, forecast or headline: E1 and E1b found no predictive support; Scout ships measurements and trends, not a debt-interest estimator (challenger §6, adopted). (4) No claim that agents cause degradation in an adopter's repo: the agent-era hypothesis (IR §5) is Scout's motivation, not a lane; agent provenance is recorded only when the repository supplies it. (5) No seven-language parity; a capability matrix states what is measured vs. detected vs. reviewed (§5.3). (6) No estate-level dashboarding. (7) No Enaible parity as acceptance.

---

## 2. Council synthesis — where the challenger changed v2

v1 §2's synthesis (agent cadence, attribution, Jev baseline, identity defects, supply chain) stands. v2 additions:

| Topic | Council leaning | Challenger | Resolution |
|---|---|---|---|
| CC-mass measure | Keep E1's `erosion` (high = healthy) | Ship the complement, higher = worse | **Adopted.** `scout.complex_burden@1`; E1/E1b `erosion` = 1 − burden recorded (§8.2) |
| Clone denominator | jscpd span ÷ lizard NLOC, capped (E1 "Deviations" 6) | Same line population top and bottom, no cap | **Adopted** (`scout.dup_pct@1`) |
| Perf defect scope | Plan decode per commit | Also the run payload rewritten per commit | **Adopted.** §10.1 fixes both |
| Artifacts/PR ingestion | Phase 5 | Earlier, or principal and interest stay disconnected | **Adopted.** Phase 3 |
| Scope for inertia series | Scout areas (Phase 4) | Must not wait on agent reconciliation | **Adopted.** Heuristic features + directory fallback (E3 Jaccard 1.00; CP07 §4 100 % ID retention on tier-1 repos) |
| Trend reporting | Direction over N runs | Fixed non-overlapping windows; ≥ 3 adverse windows past a practical delta; < 20 changes = `insufficient` | **Adopted** (§8.4) |
| Pooled-study data | Export from state branch | Local collection ≠ research consent: separate opt-in, pseudonymous | **Adopted** (§8.6) |
| Moved-vs-pasted | Ship nightly | Defer until diff attribution is evaluated | **Partly.** Shipped Phase 5 as `exploratory` (X9) |
| Agent-specific signals | Jev-scored "competing implementations" | No validated deterministic measurement; narrow checks only | **Adopted.** `stale_reference_rate` ships; competing patterns is X10 |

Not followed: "outcome-first dashboard with optional structural drill-down" as the primary framing. The user's goal is to monitor *increase* of inertia, so the structural series is first-class; v2 shows both sides side by side with outcomes never derived from structure.

---

## 3. Architecture and data model

### 3.1 Components

```
cmd/scout ─┬─ engine (lanes, base/head comparison, gate, profiles)
           ├─ inventory (scan/diff/plan; kept)
           ├─ mapping (clawpatch 0.7.3 client + Scout-side areas; §4)
           ├─ analyzers (runner/spec kept; catalogs re-curated; §5)
           ├─ measures (NEW: lizard/jscpd/import-graph → principal signals; §8.2)
           ├─ history (NEW: git mining → change-cost, coupling, ownership, hotspot; §8.3)
           ├─ artifacts (NEW: LCOV/Cobertura/JUnit + PR metadata, provenance-matched)
           ├─ guardrails (NEW: against-the-base config/suppression/test rules; Go, Python, TS)
           ├─ decisions (NEW: typed adapter — Jev; deterministic fallback; §6)
           ├─ contextual (LLM discovery; opt-in; kept, demoted)
           ├─ issues (identity/matching; rewritten; §3.3)
           ├─ state (SQLite WAL + artifacts; kept; progress path fixed §10.1) + export/import
           ├─ baseline (NEW: git state-branch publish/restore; §7.2)
           ├─ selfeval (eval renamed; scorecard; §9)
           └─ web (read-only; kept)
```

### 3.2 Record types

| Record | Nature | Identity | Can gate? | Carries |
|---|---|---|---|---|
| **Signal** | Descriptive measurement at a scope (function / file / feature / directory scope / repo) for a snapshot and, for history measures, a window | `(metric_id@version, scope_kind, scope_anchor, snapshot, window)` — append-only | **Never** | numerator, denominator, eligible-population digest, exclusions, tool/config digests, missingness (`unmeasured` ≠ 0; `insufficient` ≠ good), `claim` |
| **Issue** | Concrete mechanism with sanitized evidence at a source anchor | Opaque `issue:<family>:<ulid>` | Only in *activated* checks with attribution `introduced`/`regressed` | mechanism, rule, anchor set, rule-defined severity, dispositions + audit, advisory annotations, secondary feature/area labels, `would_block` |
| **Coverage** | What was examined, reused, skipped, blocked, why | `(file or feature, family, category)` (`internal/domain/types.go` `CoverageKey`) | Only as "required lane incomplete" (exit 2) | status, reason, reused-from run, inspected ranges |

Signal → Issue promotion only through a named, versioned rule carrying `claim`: `delta` (this PR moved a unit across a stated threshold; report-only), `trend` (§8.4 flag; informational; nightly), `policy` (repository-declared boundary; may block once activated per AL). No rule derives severity from an aggregate.

### 3.3 Issue identity and anchoring

Verified defects, still P0: `internal/issues/correlate.go:129,326-340` folds `feature:<id>` into the hashed key and `sameLocationOrFeature` (`:214-219`) requires feature intersection; `internal/state/issue_observations.go:94-102` hashes line/byte positions and the quote; `internal/state/issues.go:455-461` wipes dispositions when they change. A PR inserting lines above an accepted finding, or any mapper churn, re-mints issues and manufactures `introduced` findings (AL §1 prerequisites).

Design: durable opaque ID minted once; anchor set `{rename-followed path, qualified symbol when available, rule key, normalized mechanism, normalized evidence fingerprint}`; line/byte ranges are locators on the observation, excluded from matching. Ladder: (a) rule+path+symbol+fingerprint → match; (b) fingerprint changed → match + `evidence_drift`, disposition intact; (c) no symbol, exactly one candidate in file → match; (d) renamed path via inventory diff → (a)–(c); (e) else new; ≥ 2 candidates for (c) → both open, `ambiguous_with`, never auto-merged. Distinct occurrences of one rule in one file are distinct issues. Secrets match by secret digest so moves stay `pre-existing` (AL table 1). Feature/area labels are secondary, re-derived each run. Migration keeps `legacy_id`.

### 3.4 Lanes

`{family, trigger ∈ {pr, push, nightly}, mode ∈ {blocking-eligible, report, signal}, requires ∈ {none, tool, provider-egress, decision-egress}}`. Lane *health* (ran / reused / skipped / blocked / failed) drives exit 2/3; findings drive exit 1. Replaces `internal/engine/assess.go:139-171`'s hard-coded lanes.

---

## 4. Feature mapping (DECIDED 1 and 7 applied)

- **Heuristic map is canonical** (E3 §2 Jaccard 1.00; CP07 §4: 0.4.0→0.7.3 retained 100 % of IDs on fastapi, caddy, nest, webpack; axum gained 25 `rust-source-group` features and 50 content changes → one-time Rust re-review).
- **Refresh triggers (DECIDED):** both mappers (`map --source agent`) at onboarding, then on default-branch merges that change monitored areas — new unowned files, new/removed manifests or entrypoints, structural changes to owned paths — debounced into one refresh per burst (`mapping.refresh.debounce`, default 1 h **[ASSUMPTION]**). Never on PR runs; `scout map --refresh` forces one. PR/push runs reuse the active snapshot (E3 §4: 2–5 min remap; CP07 §5: 182–251 s on fastapi; 300 s default timeout marginal → `defaultMapCodexTimeoutMS = "900000"`, `internal/clawpatch/client.go:22`).
- **Areas** (kept): `{scout_area_id, title, file_set, origin, first_seen_snapshot, clawpatch_ids}`; new agent features match areas by file-set Jaccard ≥ 0.5 **[ASSUMPTION, X2]**; matched areas update file set and provenance; unmatched become new; fully heuristic-owned areas retire. Labels over Clawpatch output, never a second mapper (req §3), never identity. CP07 §5: agent-ID Jaccard 0.06–0.16 between fresh 0.7.3 runs, so reconciliation stays.
- **Measurement scopes** (new): §8 signals are keyed to heuristic features plus directory scopes (≤ 2 components below the source root, E1b "Unit") for files no heuristic feature owns. Areas attach as labels; the time series never re-keys when an area changes.
- **Active snapshot membership**: ownership reads only the latest refresh's merged output (E3 §1: Clawpatch never deletes).
- **Unowned files**: mapping-coverage gap; still get every file-level lane and a directory scope.
- **Failure semantics**: agent mode throwing on zero features → exploration incomplete, heuristic output kept; `registry-verifier` stays disabled (CP07 §2).
- Pin: `Revision = "734f81bc1514b52a0627edb0c24ebe1a18b36105"`, `Version = "0.7.3"` (`internal/clawpatch/client.go:19-20`); patch `scout-review-sanitize-v2` sha256 `f39d73da…a865` (`review.go:19-20`, `scripts/build-clawpatch.sh:5-9`, pnpm 11.25.0). 0.8.x is a later review (CP07 §8).

---

## 5. Analyzer lanes and languages (DECIDED 5 applied)

### 5.1 Lane catalogue (priority order)

| # | Lane | Family | Trigger | Mode | Substrate | Launch languages |
|---|---|---|---|---|---|---|
| 1 | `guardrails` — `CONFIG.ignores-added.<tool>`, `CONFIG.strictness-lowered.<tool>` (effective, through `extends`), `SUPPRESS.added.unscoped`, `TESTQ.focused-test-added`; report-only: scoped suppressions, `GATE.coverage-pragma-added`, `GATE.ci-step-neutered`, hollow tests | quality | pr, push | blocking-eligible per AL table 1; report until activated | Go-native parsers for `.golangci`, `staticcheck.conf`, `pyproject`/`ruff.toml`/`.flake8`/mypy/pyright, `eslint`/`biome`/`tsconfig*`/jest; base/head diff; per-language suppression regex | **Go, Python, TypeScript** |
| 2 | `measures` — CC/NLOC (lizard), duplication (jscpd), import graph + cycles | quality / architecture | push, nightly (full); pr (changed files) | signal; `delta` rules report-only | pinned subprocess adapters (E1 "Deviations" 5–6 ran on all 7 languages); import graph per language | CC/NLOC/dup all 7; import graph **Go, Python, TS** (`tsconfig` `paths`/`baseUrl` resolution; `import type` edges flagged type-only **[ASSUMPTION]**, AL §2) |
| 3 | `history` — change-cost, coupling, ownership, hotspot, batch size, rework (§8.3) | quality | push, nightly | signal | git (non-bot, non-merge, ≤ 50-file changes; E1b rules) | any |
| 4 | `security` — Gitleaks 8.30.1 exact ids; Semgrep `p/ci`, `p/secrets`, `p/<lang>` | security | pr, push | blocking-eligible (`security.secrets.<rule>`, `security.ci-mutable-action-ref`) | existing adapters | any |
| 5 | `artifacts` — LCOV/Cobertura/JUnit + PR metadata, revision/producer matched | quality | pr, push | signal; `unmeasured` when absent | ingestion with provenance | any producer |
| 6 | `rubric` — 12 erosion-type SCB patterns decidable from a changed-function excerpt (`silent_fallbacks`, `retry_loops_hiding_failures`, `error_message_destruction`, `validation_theater`, `stub_implementations`, `scaffold_remnants`, `lenient_parsing`, `silent_truncation`, `auto_correction_without_warning`, `log_level_hiding`, `sentinel_values_as_errors`, `ignoring_return_values`); 45-pattern cross-product rejected (`SCB/configs/rubrics/llm_judge.jsonl` 6, 19, 20, 27, 43 need a spec or penalise Go idioms) | quality | pr (changed functions), nightly sample | report; Jev-scored; ≈ $0.12 per 200-function PR **[ASSUMPTION]** | Semgrep where structural; Jev Choice where semantic | Go, Python, TS |
| 7 | `architecture-policy` — declared boundaries (`.scout/policy.yaml`: layers, cycle prohibition, max CC) | architecture | pr, push | blocking after per-repo activation | import graph + policy evaluator | Go, Python, TS |
| 8 | `performance` — Ruff PERF, CSS rule | performance | push, nightly | report | existing | Python, CSS |
| 9 | `contextual` / `clawpatch-review` | all | nightly, opt-in | report | existing | n/a |

Rule packs: corpus-tuned rules (`rules/security/semgrep-python-multipart-cve-2024-24762.yaml`, `semgrep-javascript-current-script-dom-clobbering.yaml`, `rules/performance/semgrep-css-height-transition.yaml`) become `selfeval` regression fixtures; `semgrep-actions-mutable-ref.yaml` and `semgrep-python-unsafe-yaml-load.yaml` stay.

### 5.2 TypeScript at tier 1

Guardrails = the AL table-1 TypeScript column (eslint/biome rule → `off`, ignore-list growth, `tsconfig` `exclude` growth; effective `strict`-family flags true→false; `error`→`warn`; jest `coverageThreshold` ↓; `@ts-nocheck`, bare `eslint-disable`, bare `@ts-ignore`; `.only`/`fit`/`fdescribe`). Import graph: ESM/CJS, path aliases, type-only edges separated. Fixtures and regression from webpack and nest (E1 dataset). No TS-specific precision evidence exists; TS rows follow AL track A/B like the rest.

### 5.3 Capability matrix (`scout readiness --matrix`)

| Capability | Go | Python | TS/JS | Rust | Java | C# |
|---|---|---|---|---|---|---|
| CC/NLOC/duplication; history (git); secrets/CI security | ✓ all | | | | | |
| Import graph / cycles; guardrails; rubric | P1 | P1 | P1 | P3 | P3 | – |
| CRAP (needs coverage artifact) | when produced | | | | | |
| Semantic review (contextual) | opt-in, any | | | | | |

Rejected: seven-language parity; porting lizard; wrapping Python SCB.

---

## 6. Placement of deterministic code, decision models, LLMs, humans (DECIDED 4 applied)

| Decision | Owner |
|---|---|
| Eligibility, cache keys, identity, gate, measurements, trend flags, redaction | Deterministic Go |
| Triage ranking `P(real_defect)`, rubric presence, commit classification (X5), unowned-file → area suggestion (X3) | Jev — advisory, **on from launch** |
| Discovery in context, FN discovery, second opinion on Jev-uncertain high-severity items | Reasoning LLM — opt-in |
| Dispositions, policy, activation/demotion, area acceptance, overrides (maintainer) | Humans |

Adapter guardrails (`internal/decisions`): typed questions `noul|choice|score` with `question_id@version`; bounded redacted state from the `contextual` sanitizer; output is an `annotation` row only (`issues` has no import of `decisions`); resolved-model pinning (mismatch → discarded, unit `incomplete`); cap `policy.decisions.max_per_run` default 2,000 (≈ $0.10 at E2's $0.05/1k) **[ASSUMPTION]**; unranked view always available; 5 % random low-ranked audit sample nightly; an attacker can at most lower a rank.

**Launch default and fallback.** `scout init` captures consent (`decisions.enabled`, provider, egress acknowledgement) into `.scout/policy.yaml`; the key is a CI secret. Jev runs when enabled, key present and egress allowed for the trigger; on fork PRs or when key/egress is absent, the engine uses the **deterministic ranking baseline** (`rule weight × changed-in-PR × path role`) and discloses `ranking: deterministic (reason)`. The baseline is stored for every finding regardless so X1 can measure Jev's incremental gain (E2 "Caveats" 3). Every disposition on an annotated issue becomes a blind label for `scout selfeval triage`.

Rejected: Jev as identity oracle; Jev closing `false_positive` (E2 "Recommendation"); Noul primary; `P` as calibrated confidence (E2 Caveat 6).

---

## 7. CI guardian contract (DECIDED 2 and 3 applied)

### 7.1 Triggers

`pull_request` (differential; `report` or `guardian`), `push` to default branch (full or incremental; publishes baseline; evaluates refresh triggers), `schedule` nightly (full; trend rules; opt-in LLM lanes), `workflow_dispatch` (`full`, `refresh-map`, `authorize-egress`). Local: `scout assess`, `scout diff --base <ref>`.

### 7.2 Baseline — protected `scout/state` git branch (DECIDED)

Export `state/v1/{issues.jsonl, dispositions.jsonl, signals/<metric>.jsonl, areas.jsonl, would_block.jsonl, coverage-summary.json, manifest.json}`, deterministic order, one record per line; SQLite stays the local transactional store; workflow artifacts hold bulky evidence. Only trusted default-branch runs publish, with expected-parent compare-and-swap and retry; PR and fork runs are read-only. PR runs fetch the tip and resolve merge-base coverage; digest mismatch → paired base/head scan, `comparison: recomputed`. Never commit `.sqlite`/WAL (`scripts/bundle-state.sh` → `scout export/import`). The AWS/OIDC prerequisite in `.github/workflows/scout-nightly.yml` is dropped; **S3 is deferred** to an optional adapter after the branch backend is proven; cache/artifacts are never the sole baseline (`retention-days: 90`; req §10). JSONL history is squash-safe **[ASSUMPTION]**.

### 7.3 Gate semantics

- **Attribution**: `introduced` (absent in base under identical tool/rule/policy digests), `regressed` (matched a `resolved` issue), `pre-existing`, `unknown`. Only `introduced`/`regressed` can block; `unknown` reports with a banner.
- **Profiles**: `report` (adoption default; findings never fail CI; exit 2/3 as neutral check) and `guardian` (eligible allow-list, activated per repo).
- **Allow-list (DECIDED: AL adopted).** Eligible-not-active: `security.ci-mutable-action-ref`; `security.secrets.<exact rule>` (`generic-api-key` excluded); `guardrails.CONFIG.ignores-added.<tool>`; `guardrails.CONFIG.strictness-lowered.<tool>`; `guardrails.SUPPRESS.added.unscoped`; `guardrails.TESTQ.focused-test-added` (AL §1). No wildcards. AL §2 rows are report-only.
- **Shadow window (DECIDED; supersedes AL §3's 20 runs):** every repo starts in `report` with `would_block` persisted; a check activates after **≥ 14 days AND ≥ 5 PR runs with comparable base**, both configurable (`policy.blocking.shadow.min_days`, `min_pr_runs`); maintainers triage each event `warranted|not warranted|undecided`; activate when `not warranted` = 0 and `unknown` attribution ≤ 10 % of runs **[ASSUMPTION]**, by a maintainer PR to base `.scout/policy.yaml` (`blocking: [{id, version, activated_at, evidence: {would_block, warranted, window_days, pr_runs, dossier_digest}}]`). A parser/rule version bump returns the check to shadow.
- **Rare security checks (DECIDED, conservative default):** zero events at window end → `security.secrets.*` and `security.ci-mutable-action-ref` may activate on track A alone, `provisional: true`; first-baseline `pre-existing` counts are the fire-rate proxy.
- **Overrides (DECIDED, conservative default):** native justified suppressions where they exist; otherwise `scout:allow <id> reason=… ref=…` or label `scout-override`, **accepted only from a maintainer** (policy-declared team / `CODEOWNERS`), stored as `disposition: accepted{actor, pr, reason}`; findings never deleted; overrides trigger review, not FP labels. Gitleaks runs with `--ignore-gitleaks-allow --gitleaks-ignore-path /dev/null` (`internal/analyzers/catalog_security.go:266-267`), so the secrets escape is Scout-side only.
- **Two-track promotion (replaces v1 §7.4, unsatisfiable: zero errors at n = 50 gives Wilson LB 0.929; AL §4):** Track A certification (global): `selfeval` fixtures positive / clean negative / boundary / redaction *plus attribution fixtures* (line move, rename, version bump, config migration, tool upgrade, new file), no known systematic negative class, versioned dossier → eligibility. Track B (per repo): the shadow window above; retention ≥ 3 confirmed-unwarranted blocks in the trailing 20 → auto-demote **[ASSUMPTION]**; gate-disable events tracked. Corpus precision is an audit instrument, never a prerequisite.
- Pre-existing debt never blocks; optional lanes never turn a required lane red; a required blocking lane that could not run on head → exit 2.

### 7.4 Exit codes, outputs, trust boundary

`0 clean / 1 gated / 2 incomplete / 3 unavailable`, precedence 3 > 2 > 1 > 0; JSON always has `gate`, `health`, `would_block`, `ranking`. Outputs: `scout-assessment.json`, ≤ 60-line Markdown summary (introduced/regressed, would-block, pre-existing count, lane health, signal deltas), optional SARIF. Reusable workflow `scout-guardian.yml` + composite action; binary and Clawpatch bundle from a **pinned release**, never built from the PR checkout (today's `scout-nightly.yml` builds `./cmd/scout` from the checkout); policy read from the *base* branch; `contents: read` on PR, `contents: write` scoped to `scout/state` on trusted runs; no provider/decision secrets on fork PRs. First adoption: `scout init` → consent → full default-branch run → baseline; nothing is `introduced` before a baseline exists.

---

## 8. Inertia monitoring: principal and interest over time

### 8.1 Framing and claim level

Inertia = a codebase's accumulated resistance to change (IR §1). Scout operationalises the debt metaphor: **principal** = structural measures of a scope at a snapshot; **interest** = observed change-cost outcomes for the scope over a trailing window; **exposure** = how often the scope changes (debt costs only where code moves; IR §3). The user's hypotheses — degradation raises later change effort/complexity (H1), unpredictability (H2), bug introduction (H3) and maintenance complexity, and agents magnify this — are the *reason to measure*, not something Scout asserts per repository.

Every measure here is **`descriptive`** by E1 and E1b's preregistered rules. Scout reports levels, deltas and trend direction; it never gates, sets severity, forecasts, computes a composite, or upgrades `claim` on its own (§8.5). What launches is a versioned, comparable record of both sides over time per scope, with population and missingness stated — the data E1b lacked.

### 8.2 Principal measures (snapshot; function → file → scope)

Scope = heuristic feature or directory scope (§4). Eligible population = production source by E1's extension/exclusion rules (`evaluation/inertia-v1/exclusions.json` pattern), NLOC ≥ 10, generated/vendored excluded by path role; the population digest is stored with every signal.

| Metric | Formula | Source | Ships |
|---|---|---|---|
| `scout.cc@1` | per function CC, NLOC; scope `n_high_cc` = #functions CC > 10, `cc_p90`, `cc_max` | lizard (pinned) | Phase 3 |
| `scout.complex_burden@1` | B = Σ_{f: CC_f>10} CC_f·√NLOC_f ÷ Σ_f CC_f·√NLOC_f, higher = worse (E1/E1b `erosion` = 1 − B) | lizard | Phase 3 |
| `scout.dup_pct@1` | duplicated eligible source-line positions ÷ eligible source-line positions, one tokenizer population, no cap; clone groups kept with cross-scope flag | jscpd 4.0.8 (min-tokens 50, min-lines 5, `--max-lines 10000 --max-size 500kb`) | Phase 3 |
| `scout.cycle_mass@1` | internal import edges on a directed cycle ÷ internal import edges; `cross_boundary_cycles` separate; `unresolved_imports` reported; type-only edges excluded and counted | import graph (Go, Python, TS) | Phase 3 tier 1; `unmeasured` elsewhere |
| `scout.crap@1` | per function CRAP = CC² × (1 − cov)³ + CC, cov ∈ [0,1]; scope `crap_load` = Σ_{CRAP>30} CRAP; `uncovered_cc_mass` = Σ CC_f·(1−cov_f) ÷ Σ CC_f | lizard + revision-matched LCOV/Cobertura | Phase 3 when the artifact exists; else `unmeasured`, never 0 |
| `scout.propagation_cost@1` | mean fraction of scope nodes reachable via imports (MacCormack 2006); Decoupling Level deferred | import graph | Phase 5 |
| `scout.cognitive_cc@1` | SonarSource nesting-weighted complexity | gocognit / complexipy / eslint-sonarjs **[ASSUMPTION: adapters]** | later |
| `scout.stale_reference_rate@1` | backticked identifiers/paths in `AGENTS.md`, `CLAUDE.md`, `README*`, docstrings resolving to no file/symbol ÷ references checked | regex + inventory + lizard symbols | Phase 5, `exploratory` (IR §5; deterministic) |

Deltas are computed between consecutive snapshots on the *same population digest*; membership change is a separate `population_delta`, never degradation (challenger §6). PR runs apply `delta` rules on changed files only (e.g. "CC 9 → 14"), report-only.

### 8.3 Exposure and interest measures (trailing window; per scope)

Windows: fixed, non-overlapping, length `W` (default 90 days **[ASSUMPTION]**), calendar-aligned so successive windows compare; a trailing 365-day view is also stored for E1b comparability. Change = non-merge, non-bot (E1b regex `(?i)(bot|renovate|dependabot)`), ≤ 50 files, touching ≥ 1 eligible file in scope. Definitions follow E1b's preregistration so adopter data pools with the study.

| Metric | Formula | Source | Ships |
|---|---|---|---|
| `scout.changes@1` (exposure) | changes; distinct authors | git | Phase 3 |
| `scout.hotspot@1` | Σ_files (Σ_f CC_f) · log1p(changes to file) ÷ scope NLOC | lizard + git | Phase 3 |
| `scout.churn_per_change@1` (I1) | p50, p90 of added + removed lines in scope files per change | git | Phase 3 |
| `scout.amplification@1` (I2) | p50 eligible files touched repo-wide per change touching the scope; p50 scopes touched | git | Phase 3 |
| `scout.follow_up_rate@1` (I3) | changes followed ≤ 7 d by a change touching ≥ 1 same file ÷ changes with a complete 7-day observation | git | Phase 3 |
| `scout.pr_effort@1` (I4) | p50 commits/PR, review submissions + comments/PR, hours created→merged, for PRs whose merge/squash commit is a change in scope; `unmeasured` < 5 linked PRs | `pull_request` payload + API (`pull-requests: read`) | Phase 3 in CI; `unmeasured` locally |
| `scout.dispersion@1` (I5) | p90 ÷ p50 of churn and of PR hours, when p50 > 0 | git, PR | Phase 3 |
| `scout.change_coupling@1` | changes touching scope and ≥ 1 other scope ÷ changes touching scope; pairwise support and P(b\|a) for top pairs; `unexplained` subset = pairs with no import edge (tier 1 only) | git (+ import graph) | Phase 3; `unexplained` Phase 5 |
| `scout.ownership@1` | top author's changes ÷ scope changes; author count; minor contributors < 5 %; pseudonymised keys, never exported as names | git | Phase 3 |
| `scout.rework@1` | removed lines with blame origin inside the window ÷ removed lines (E1 "Deviations" 7) | git blame | Phase 3 |
| `scout.batch_size@1` | lines and files per PR (p50, p90) | PR metadata | Phase 3 |
| `scout.moved_vs_pasted@1` | per change: added blocks ≥ 5 lines matching a removed block in the same change = `moved`, matching base-tree lines = `pasted`, else `new`; window shares | git diff + jscpd on hunks | Phase 5, `exploratory` (X9) |
| `scout.fix_touches@1` (H3 proxy) | `fix`-classified changes touching scope ÷ changes | Jev classifier (X5) | after X5; regex precision 0.52 (E1) is disqualifying |

### 8.4 Reporting over time per repository

- **Side-by-side panel per scope**: principal levels and Δ, exposure, interest, with `n changes`, population digest, missingness. `n < 20` changes in a window ⇒ interest = `insufficient`, never shown as good **[ASSUMPTION]**.
- **Adverse-trend flag** (`trend` rule, informational, nightly): a metric moves adversely across ≥ 3 consecutive comparable windows by more than a practical delta (defaults `complex_burden` +0.05, `dup_pct` +2 pts, `cycle_mass` +0.05, `n_high_cc` +20 % **[ASSUMPTION]**) on a stable population; same for interest metrics. Lehman's mixed evidence (IR §1) means *direction* is the finding; no inevitability implied.
- **Inertia watch list** (descriptive, never a composite): scopes with an adverse principal trend AND top-quartile exposure — "high principal paid on frequent changes" (IR §4), presented as where to look, not as predicted cost.
- **Interventions**: `scout annotate --window … intervention: refactor|migration|…` keeps a legitimately churn-heavy window out of trend flags and records it as an event (challenger §6).
- Every panel carries `claim: descriptive` and a link to §8.5.

### 8.5 Honest reading of E1 and E1b, and what would upgrade a claim

What they show: E1 (1,732 files, 7 repos) — quality *levels* add +1.18 % [−0.62, +3.24] to file-level fix-proneness prediction beyond size and history; AUROC Δ −0.019. E1b (60 areas, 5 repos, 3 with ≥ 10 units; 35 areas for PR outcomes) — level, degradation, coupling, ownership give no preregistered support for effort (O1 −0.44 % [−2.72, +1.42]), predictability (O5) or follow-ups (H4); nothing survives Holm. These are failures to demonstrate *incremental* prediction under strict rules, not demonstrated absence. What keeps the hypotheses open (E1/E1b "Threats"): (1) **power** — CIs ±3 to ±20 points; a 5 % effect could not be reliably detected; (2) **the lagged-outcome control** — the baseline includes the prior window's value of the same outcome, so the test asks whether quality predicts *change in* effort beyond persistence; if persistent quality already caused past effort, that pathway is absorbed — a stricter estimand than "does degraded code cost more to change" (challenger §2 concurs; removing the lag would not establish causality either); (3) ridge chose the grid ceiling for several PR-outcome models, making base and candidate near-intercept; (4) E1's fix labels have precision 0.52 — attenuating, or biasing either way if error is differential; (5) directory areas ≠ features; incomplete PR linkage; squash-vs-merge differences; (6) E1's raw correlations are real (Spearman cc_max +0.35) but carried by size and history, which are partly consequences of quality — an incremental test cannot separate that. Hypothesis-generating only: O4c PR hours +9.05 % [−2.89, +18.04]; file-level `clone_pct` on churn dispersion +6.45 % [+2.02, +11.85] (36 uncorrected tests).

A better test, by strength: (a) **pooled adopter panel** with Scout's own definitions — repeated snapshots, repository-clustered uncertainty, held-out repositories *and* future periods, workflow/task covariates, preregistered power simulation; CI half-widths shrink ≈ 1/√n, so a 5 % effect with lower bound > 0 needs roughly 4–8× E1b's 60 areas: ≥ 300 scope-windows from ≥ 15 repositories **[ASSUMPTION]**; (b) **controlled continuations on SlopCodeBench** (external study, X8): same task from behaviour-equivalent healthy vs degraded snapshots, fixed budgets, repeated seeds; outcomes success, cost, steps, subsequent degradation — the only design that tests *agent magnification* (IR §5; SCB `measurement-validity.md` lists causal inertia as unvalidated); (c) **within-scope event studies** around recorded interventions with comparison scopes and pre-trend checks, acknowledging selection into refactoring; (d) effort measured directly (PR hours, review rounds) not churn; (e) a centrality control.

Decision rule, fixed now: a metric's `claim` moves to `predictive` for a named outcome only by a preregistered pooled test meeting E1b's rule (≥ 5 % relative MSE reduction, CI lower bound > 0, consistency across ≥ all-but-one repositories with ≥ 10 units, multiplicity-controlled). Causal claims need (b) or (c). Gate use is a further separate decision (AL). Repeated dashboard significance never upgrades a claim.

### 8.6 Accumulating evidence from adopters without overclaiming

Per default-branch run Scout stores (SQLite and `signals/inertia.jsonl` in `scout/state`): snapshot revision/time, window bounds, scope ID, population digest, map version and scope lineage (split/merge), every §8.2/§8.3 value with numerator, denominator, tool/config digests, import-resolution coverage, artifact revision/producer, missingness reason, outcome maturity (right-censoring), and a compact sanitized change table (deduplicated change/PR keys → scope, churn, files/scopes touched, timestamps, workflow type, bot flag, linkage completeness, optional task category, human/agent provenance only when the repository supplies it). Medians alone cannot support re-analysis; the change table makes alternative outcome definitions possible later. Bulky evidence stays in artifacts.

`scout export --study` is a **separate opt-in** from local collection: schema in `docs/evidence/study-export.md`; repository-scoped pseudonyms; scope-level aggregates plus the change table; no source, commit messages, emails or reviewer identities; retention and withdrawal stated. Public `scout/state` branches carry a notice that git history makes erasure hard. Overlapping scopes are flagged so one change never multiplies pooled observations. Ownership is stored as coarse aggregates, never a personnel ranking.

Per repository, Scout may compute the E1b comparisons on its own ledger once ≥ 2 years of windows exist, reported as `descriptive, n = …, single repository`; the upgrade rule never applies to them.

---

## 9. Scout's own outcome scorecard

Each line `count/attempts`, Wilson 95 % interval, `unmeasured` counts (FXL `docs/references/scoring.md`); no single solve rate.

1. **Actionable yield** per lane/rule: `accepted` or resolved-by-change ÷ dispositioned; `open` separate.
2. **Follow-through**: confirmed issues resolved within 30/90 days ÷ confirmed with a mature window.
3. **Scoped assurance**: random `(file, lane)` units called clean, independently audited → false-assurance rate per lane; no repository-wide FN claim.
4. **Cost of use**: unwarranted blocks (AL retention), gate-disable events, overrides, duplicate/re-opened issues per 100, time-to-disposition, runtime and spend per PR, agent-refresh value.
5. **Inertia-monitoring usefulness** (new): adverse-trend flags and watch-list entries marked `useful|not useful` by a maintainer; `undecided` separate. Attention value, not validity.
6. **Change-success rate** (experimental, descriptive): behind X5 only.
7. **Triage-model quality**: per-rule AUROC and routing precision of Jev vs. the deterministic baseline on blind post-launch labels.
8. **Corpus regression**: 9 true / 13 FP (`evaluation/reference-v1.json`) stay detected / non-flagged.

---

## 10. Keep / cut / rewrite

### 10.1 P1 performance defect — `scout assess` run-job persistence

Observation (user-reported, ground truth): after the 0.7.3 upgrade smoke, `scout assess` on fastapi ran > 1 h without finishing; the goroutine dump shows `CommitRunJob → loadPlanTx` re-reading and decoding a ≈ 5 MB plan per job commit.

Code (verified): `internal/state/progress.go:117-200` `CommitRunJob` runs per job in one transaction: `loadRunProgressTx` (decodes the run payload), `loadPlanTx` (`:331-346`: `SELECT payload FROM plans`, `decode`, `plan.Validate()`), `plannedJob` (`:389`, linear scan), then appends the commit's coverage/failures/artifacts to `run`, re-encodes the whole run and `UPDATE runs SET payload` (`:189-197`). Plans are insert-only (`internal/state/issues.go:34-37`, `ErrImmutable`), so a plan never changes after `SavePlan`; `Plan` carries the full `Snapshot` inventory (`internal/domain/types.go:639-647`), the source of the 5 MB. Cost for J jobs, plan size P, mean result r ≈ O(J·P + J²·r) [INFERENCE from code; not profiled].

Fix (minimal, both paths): (1) at `SavePlan` also write `plan_jobs(plan_id, job_id, payload)` plus the plan's identity fields; `CommitRunJob` reads one job row and checks `plan_id/run_id` without decoding the snapshot (an in-process decoded-plan cache keyed by `plan_id` is an acceptable first step because plans are immutable, if private to the `Store`, bounded, synchronised and hydrated on restart); (2) stop rewriting the run payload per commit — coverage already lives in `coverage`, the commit in `run_job_commits`; `LoadRun`/`LoadInProgressRun`/`FinishRun` reconstruct progress from those tables and assemble the final payload once. Preserve initial reused coverage from `BeginRun` (`:79-114`), atomicity, duplicate-commit rejection, resume-after-interruption equivalence.

Acceptance (Phase 1): a `go test ./internal/state` benchmark commits 5,000 jobs against a plan with a 5 MB snapshot and asserts per-commit time and allocations flat in J, with no `plans` read and no `runs` payload write in the commit path (SQL trace counts); interruption-resume test yields the same finalized run; `scout assess` on the pinned fastapi revision (E3 `d666ccb`, CP07 §1), heuristic map, egress off, completes end-to-end in < 10 min on the development machine **[ASSUMPTION target; the > 1 h failure is accepted as reported, not rerun]**.

### 10.2 Code verdicts

| Package / file | Verdict |
|---|---|
| `internal/inventory/*`, `internal/analyzers/{runner,spec,parse,sanitize,ruff,semgrep,gitleaks}.go`, `internal/state/{store,artifacts}.go`, `internal/contextual/*`, `internal/web/web.go`, `scripts/build-clawpatch.sh` + patch | Keep (inventory: lanes must not depend on `MappingRequired`; add paired base/head plan entry; state: add export/import; contextual: opt-in nightly, sanitizer reused by `decisions`; web: inertia panels later) |
| `internal/state/progress.go` | **Fix** (§10.1), then keep |
| `internal/state/{issues,issue_observations}.go` | Rewrite identity-sensitive parts: positions/quote out of `observationSupportKey` (`issue_observations.go:94-102`); wipe (`issues.go:455-461`) → `evidence_drift`; drop `ApprovalRef`/`ProposedAction` |
| `internal/issues/correlate.go` | Rewrite matching (§3.3); keep `SanitizeText`/provenance |
| `internal/domain/types.go` | Extend: `Signal` (window, population digest), `Area`, `Scope`, `Annotation`, `Attribution`, `LaneHealth`, `WouldBlock`; remove approval types |
| `internal/engine/{assess,readiness,config}.go` | Rewrite orchestration: lane registry, profiles, base-branch policy, consent/fallback |
| `internal/analyzers/catalog_*.go` | Re-curate: lizard/jscpd specs; import graph + policy evaluator (architecture currently `return nil`); packs + exact Gitleaks ids; corpus-tuned rules → fixtures |
| `internal/clawpatch/{client,records,review}.go` | Keep boundary; add explicit `source`/`refresh` + debounced trigger predicate; heuristic must run offline (`client.go:168-170` blocks it without egress today) |
| `internal/{measures,history,artifacts,guardrails,decisions,baseline}` | Create (§§5–8) |
| `internal/eval/*` → `internal/selfeval` | Rename; scorecard (§9) |
| `internal/codex/exec.go` | Cut (no callers) |
| `cmd/scout/main.go` | Rewrite gate (`--fail-on`, `:202-207` → profile/attribution); add `init`, `diff`, `map --refresh`, `export/import [--study]`, `signals`, `trends`, `annotate`, `selfeval`, `readiness --matrix`; remove `approve` |
| `.github/workflows/scout-nightly.yml`, `scripts/bundle-state.sh` | Rewrite → `scout-guardian.yml` + dogfood caller (drop AWS/OIDC; add `pull_request`); bundle script → `scout export/import` |

### 10.3 Spec cutover (DECIDED 6)

`SCOUT_REBUILD_REQUIREMENTS.md` is archived under `docs/evidence/` and replaced by `docs/scout-product-contract.md` (§§1, 3, 7, 8.4), `docs/scout-adapter-contracts.md` (tool adapter; Clawpatch command contract amended for refresh triggers; typed-decision adapter; LLM provider; identity/matching; incremental planning; signal/measure contract) and `docs/evidence/` (Enaible inventory, corpus notes, `study-export.md`). Kept from req: §8 gap→rule promotion with FXL's FCR record shape; §12.8 fixture classes plus AL attribution fixtures.

---

## 11. Phased plan — every phase ends in a runnable scenario

**Phase 1 — Persistence fix, identity, offline ledger** (`state/progress`, `issues`, `state`, `domain`, offline heuristic map)
Scenario: §10.1 benchmark and the pinned fastapi assess; on a credential-less fixture run `scout assess` twice (second after inserting 20 lines above a flagged function, renaming a file with a finding, duplicating a same-rule finding in one file); disposition one issue `accepted`; resolve one by deleting code; re-introduce it.
Acceptance: §10.1 criteria; one durable ID across line shift and rename; `evidence_drift` with disposition intact; two IDs for two occurrences; `regressed` on re-introduction; heuristic map ran with egress off; `go test ./internal/issues ./internal/state` green; migration test re-anchors a v1 store and lists ambiguous rows.

**Phase 2 — Differential guardian, state branch, consent** (`baseline`, `engine`, `cmd`, reusable workflow, `decisions` fallback)
Scenario: a fixture repo adopts the workflow in `report`; `scout init` records consent off; default-branch run publishes `scout/state`; PR A adds a GitHub PAT and a bare `# noqa`; PR B touches pre-existing debt only; PR C is a fork; a rule-digest bump is simulated; the shadow window is simulated (14 days, 5 PR runs) and a maintainer activates `guardrails.SUPPRESS.added.unscoped`; PR D from a non-maintainer carries `scout-override`.
Acceptance: PR A exits 0 in `report` with two `would_block` entries persisted; after activation an identical PR exits 1 listing exactly the activated `introduced` issue, the PAT still `would_block`; PR B exits 0; PR C cannot write the branch or read secrets and reports `ranking: deterministic (fork)`; digest mismatch → `attribution: unknown`, `comparison: recomputed`; PR D's override rejected and logged; two concurrent publishes land via CAS; `scout import` reproduces the ledger byte-identically; a rule-version bump returns the check to shadow.

**Phase 3 — Quality substrate for Go, Python, TypeScript: measures, history, artifacts, guardrails**
Scenario: on Scout's own repo, a Python fixture and the nest or webpack checkout: nightly writes all Phase-3 signals of §8.2/§8.3 for four synthetic quarterly windows; a PR raises a function CC 9 → 14, adds a clone, lowers `ruff` and `tsconfig` strictness, adds `//nolint` and `@ts-nocheck`, commits a `.only`; an LCOV artifact is attached to one run and not the next.
Acceptance: `scout signals --scope` shows numerator/denominator/population digest; CRAP present only with a matching artifact, `unmeasured` otherwise; `pr_effort` `unmeasured` locally, present in CI; `delta` issues carry `claim: descriptive`; all four guardrail mechanisms fire with base/head evidence in all three languages; TS import graph resolves `paths` aliases and separates type-only edges; 9/13 corpus regression holds; `readiness --matrix` prints §5.3; E1b's O1–O5 recomputed by Scout on the nest/webpack checkout match `evaluation/change-cost-v1/outcomes-area.csv` within rounding for overlapping directory scopes.

**Phase 4 — Mapping refresh triggers and areas**
Scenario: fastapi and axum fixtures; onboarding refresh; three default-branch merges in quick succession (new unowned package, new manifest, content-only); then a content-only PR.
Acceptance: exactly one debounced refresh after the burst, none after the content-only merge or on the PR; prior issues keep IDs; area file-set Jaccard across refreshes ≥ 0.8 **[ASSUMPTION, X2]**; stale Clawpatch records excluded from ownership; unowned files listed with directory scopes; inertia series keys unchanged by the refresh; agent failure leaves heuristic output and marks exploration incomplete.

**Phase 5 — Trends, exploratory measures, rubric lane, self-evaluation**
Scenario: replay four Phase-3 windows plus a fifth adverse one; mark one window `intervention: refactor`; replay E2 cases plus ≥ 100 blind labels through deterministic ranking and Jev; prompt-injection fixture; scorecard on Phase 2–3 ledgers; `moved_vs_pasted`, `stale_reference_rate`, `propagation_cost` on the corpus.
Acceptance: trend flag fires only after three adverse comparable windows and not across the intervention window; `n < 20` windows show `insufficient`; watch list sorted without a composite; annotations carry resolved model and question version; forced mismatch → `incomplete`; no state/gate change from any annotation; injection fixture changes only rank and appears in the audit sample; scorecard prints Wilson intervals; X1 rule evaluated and recorded; exploratory metrics tagged and excluded from trend flags.

**Phase 6 — Study export, spec cutover, web, release**
Scenario: `scout export --study` on the dogfood ledger; docs replace the archived spec; release ships binary + Clawpatch 0.7.3 bundle + reusable workflow; Scout dogfoods itself in `guardian` through its own shadow window.
Acceptance: export contains no paths outside scope IDs, no messages, no identities, validates against `docs/evidence/study-export.md`; adopter checklist completes in one PR; unwarranted-block count over the dogfood window reported; `SCOUT_REBUILD_REQUIREMENTS.md` archived.

---

## 12. Top risks

| # | Risk | L / I | Mitigation / reversibility |
|---|---|---|---|
| R1 | Identity rewrite still mis-merges or splits | medium / high | Ladder keeps ambiguity explicit; Phase 1 fixtures; `legacy_id` retained |
| R2 | Unwarranted blocks erode trust after activation | medium / high | Eligible-not-active, shadow window, retention auto-demote, maintainer overrides |
| R3 | Inertia panels read as judgments: growth or refactoring mistaken for degradation; sparse scopes look healthy; metric noise | medium / medium | `population_delta` separated; intervention marks; `insufficient` floor; no composite; §9.5 usefulness line; concrete issues first, trends second |
| R4 | Hypotheses never testable from Scout data (few adopters, no consent) | medium / medium | Export schema fixed by Phase 6; definitions pooled with E1b; X8 is independent of adopters |
| R5 | State-branch permissions unavailable or history bloat | medium / medium | Squash-safe JSONL; S3 adapter later (DECIDED order) |
| R6 | Metric drift mistaken for change (tool upgrade, re-partition, population churn) | medium / medium | `metric@version` + tool and population digests; scope keys independent of agent areas |
| R7 | Guardian supply chain: PR controls its examiner | low / high | Pinned release, base-branch policy, read-only PR tokens, no secrets on fork PRs |
| R8 | Persistence fix regresses resume semantics | low / high | Interruption-resume equivalence test in Phase 1 |
| R9 | Clawpatch pin ages (0.8.x); patch re-anchor | medium / medium | `build-clawpatch.sh` verifies patch; each upgrade = compatibility review + reduced E3 (CP07) |
| R10 | Jev ranking hides findings, or is silently absent | low–medium / medium | Route-only, unranked view, audit sample, disclosed deterministic fallback, X1 |
| R11 | Adopters refuse `pull-requests: read` | medium / low | `pr_effort`/`batch_size` `unmeasured`, never 0 |
| R12 | Ownership or change tables expose contributors | low / high | Pseudonyms, coarse aggregates, separate study consent, public-branch notice |

---

## 13. Experiments with decision rules

| ID | Question | Design | Decision rule |
|---|---|---|---|
| X1 | Jev vs deterministic ranking on triage retrieval | ≥ 100 de-correlated blind labels (≥ 30 semantic positives, ≥ 20 injection/indirection negatives); recall@k at fixed budget | Jev stays default-on iff held-out recall@k gain ≥ 10 points, Wilson LB > 0, no critical true issue below the audit floor; else default = deterministic baseline, `--rank jev` opt-in (route-only is not reopened) |
| X2 | Do areas stabilise agent output? | E3 protocol with reconciliation, fastapi/axum, 3 fresh + 2 same-state + 2 mutations | Adopt the Jaccard threshold yielding survival ≥ 0.8 and ≤ 5 % spurious merges; else refresh only at onboarding + manual |
| X3 | Jev assigns unowned files to features? | Choice on 200 unowned files, hand labels | Reconciliation aid iff precision ≥ 0.9 on `none`-exclusive decisions |
| X4 | Rubric-lane precision per pattern | 12 patterns × 300 changed Go/Python/TS functions; double-labelled hits | Keep iff Wilson LB ≥ 0.6; never blocking |
| X5 | Jev commit classifier vs regex | 200 blind-labelled commits across 7 repos | Enable `fix_touches`/change-success iff precision ≥ 0.85, recall ≥ 0.8 (LBs ≥ 0.75/0.7) |
| X6 | Guardrail unwarranted-block rate (track-B dry run) | Replay last 200 merged PRs of 3 adopter repos, three languages; double-label `would_block` | Feeds AL dossiers and shadow defaults; activation stays per repo |
| X7 | Pooled predictive re-test (post-launch; replaces v1 X7) | Preregistered panel over consenting adopters' study exports with §8 definitions; held-out repos and periods; power simulation first | `claim: predictive` for a (metric, outcome) only by E1b's rule, multiplicity-controlled; gate use a separate decision |
| X8 | Controlled continuations (external SCB/FXL study, not a Scout phase) | Same task from behaviour-equivalent healthy vs degraded snapshots, fresh sessions, fixed budgets, repeated seeds; outcomes success, cost, steps, subsequent erosion/verbosity | Supports agent magnification iff degraded starts show lower success or higher cost with CI excluding 0 across ≥ 2 agents; motivates adoption, never upgrades a Scout claim |
| X9 | Exploratory-measure attribution validity | Hand-check 200 `moved/pasted/new` labels and 200 `stale_reference` hits on the corpus | `exploratory` → `descriptive` iff label precision Wilson LB ≥ 0.9; else keep or drop |
| X10 | Competing-implementation detection | Jev Choice "same concern, different implementation?" over clone-group and signature-similar pairs; 200 labels | Ship as exploratory iff precision ≥ 0.8; else research note (IR §5) |

---

## 14. Assumptions register

- A1 Area Jaccard 0.5 match / 0.8 survival (§4, X2).
- A2 Activation condition `unknown` ≤ 10 % of runs; retention demotion 3/20 (§7.3).
- A3 Protected branch lower friction than object storage; JSONL history squash-safe (§7.2).
- A4 Deterministic ranking captures much of Jev's apparent gain (§6, X1).
- A5 Rubric cost ≈ $0.12 per 200-function PR; precision unknown (§5.1, X4).
- A6 Decision budget 2,000/run; refresh debounce 1 h (§4, §6).
- A7 Window 90 days; `insufficient` floor 20 changes; trend deltas and three-window rule (§8.3–8.4).
- A8 Pooled test needs ≥ 300 scope-windows from ≥ 15 repositories (§8.5).
- A9 TS type-only import handling and `tsconfig` path resolution suffice for cycle measurement (§5.1).
- A10 Perf acceptance < 10 min on the dev machine; O(J·P + J²·r) inferred from code, not profiled (§10.1).
- A11 Cognitive-complexity adapters exist per tier-1 language (§8.2; not in launch set).
- A12 Naive SCB idiom patterns penalise Go/TS idioms (§5.1 rubric list).

---

## 15. Decisions for the user (remaining, genuinely open)

1. **Study-export consent contract** (§8.6): confirm pooled export as a separate opt-in with the stated contents (scope-level aggregates + sanitized change table, pseudonymised authors, no messages/identities), and whether Scout's own dogfood ledger becomes the first pooled dataset.
2. **External controlled-continuations study (X8)**: commission now on SlopCodeBench/factory-x-loops in parallel with Phases 1–3 (the only design that tests the agent-magnification motivation), and who runs it.
3. **Performance acceptance target** (§10.1): confirm "< 10 min end-to-end on the pinned fastapi revision, heuristic map, egress off, development machine" as the Phase-1 bar, or give another machine/time.
4. **PR-metadata ingestion** (§8.3 `pr_effort`, `batch_size`): accept `pull-requests: read` in the reusable workflow for trusted runs, or keep interest measures git-only (I4 always `unmeasured`).

All other open items from v1 §15 are decided (Appendix) and applied above.

---

## Appendix — Decisions taken (authoritative; v1 §15, 2026-10-19)

1. Map refresh: both mappers at onboarding, then on default-branch merges that change monitored areas (new unowned files, new/removed manifests or entrypoints, structural changes to owned paths), debounced into one refresh; never on PR runs; areas layer stays.
2. Baseline: protected `scout/state` git branch; S3 deferred to an optional adapter after the branch backend is proven.
3. Blocking allow-list: `docs/scout-allowlist.md` adopted (eligible-not-active; two-track promotion); shadow window ≥ 14 days **and** ≥ 5 PR runs with comparable base, both configurable; overrides require a maintainer; rare security checks may activate `provisional: true` on track A alone (conservative defaults, to be confirmed by use).
4. Jev: on from launch, route/rank only; never changes issue state or the gate.
5. Language tiers: tier 1 = Go, Python, TypeScript.
6. Spec cutover: archive `SCOUT_REBUILD_REQUIREMENTS.md` in favour of the §10.3 documents.
7. Clawpatch: v0.7.3, commit `734f81bc1514b52a0627edb0c24ebe1a18b36105`, ported patch `scout-review-sanitize-v2` (CP07 §7); 0.8.x is a later separate review.
