# Scout — design and planned approach, version 2

Status: council decision memo v2 (2026-10-19), revision 2.1 (2026-10-21), **revision 2.2 (2026-10-23)**. Supersedes v1 (2026-10-17). v2 folds in the user's decisions (Appendix, authoritative, not reopened), the inertia literature review, the Clawpatch 0.7.3 upgrade and a P1 performance defect found during the upgrade smoke. An independent adversarial analysis by a different model family (`InertiaChallenger`, transcript `history://ScoutDesignV2.InertiaChallenger`) was developed in parallel on the inertia section, the defect and the plan; §2 records where it changed the recommendation. Revision 2.1 replaces every use of the invalidated E1b v1 study with E1b v2 (R-035), the first study run under the `AGENTS.md` experiment rules (design review `docs/design-reviews/2026-10-21-e1b-v2/`). Phase 1 is implemented in the working tree (`docs/STATUS.md` §1); every later acceptance scenario is future work.

**Revision 2.2** (design review `docs/design-reviews/2026-10-22-rescope/`; R-036, R-037, R-038; no new experiment is cited) re-scopes Scout to a product and a future inertia-bench subject, following the user's decisions of 2026-10-22 (`docs/STATUS.md` §6). (1) Reasoning-LLM lanes (`contextual`, `clawpatch-review`, agent mapping) run only through `claude -p` / `codex exec` (Scout's own adapter, and Clawpatch's `codex`/`claude` providers in host auth context), report `skipped (no model CLI)` wherever no CLI is logged in, and CI carries no Anthropic/OpenAI keys; Jev stays on OpenRouter (§§3.1, 4, 5.1, 6, 7.4, 10.2, 10.3; risks R13, R14). (2) Hypothesis studies move to inertia-bench: §8.5's list of further tests, X7, X8, §15.2 and assumption A8 are removed, the claim-upgrade rule stays, and `scout export --study` stays a product feature with Scout-owned consent (§§8.5, 8.6). (3) New §8.7 adopts subject-interface principles only, with Phase 3b; the exact schema waits for its own council round. (4) The Phase-1 performance bar is decided (§10.1, Appendix 14). (5) E1b-v2 is grandfathered. (6) §15 is renumbered and the Appendix extended. (7) The Jev adapter is sharpened per the `jev-actions` skill review (`docs/design-reviews/2026-10-22-rescope/jev-actions-review.md` §A; R-038): a defined question version key, per-unit failure handling mapped to lane health, one Noul per rubric pattern and the X3/X10 primitive choices, probability-sampled blind audits, the labelling protocol of `docs/STATUS.md` §6.8, and "calibration unchecked; ranking only" wording. The estimator, split, interval and population mechanics of X1, X3, X4, X5 and X10 are not set here: each experiment's council-signed preregistration fixes them, together with every numeric minimum, cost bound and escalation policy that depends on the user's open answers. Minimums carried over from revision 2.1 (X1, X3, X4, X5, X10) are flagged "rev 2.1 value, pending the user's minimums; the preregistration confirms or replaces it". The Appendix decisions of 2026-10-19 (including Appendix 4, Jev on from launch, route/rank only) are unchanged.

Evidence notation. `brief §n` = `docs/scout-design-brief.md`; `req §n` = `SCOUT_REBUILD_REQUIREMENTS.md` (to be archived); `AL §n` = `docs/scout-allowlist.md`; `IR §n` = `docs/inertia-research.md`; E1 = `evaluation/inertia-v1/RESULTS.md`; **E1b = `evaluation/change-cost-v2/{PREREGISTRATION,AMENDMENT-1,RESULTS}.md` (R-035; council-signed, run 2026-10-21; grandfathered: it used OpenRouter annotators, including Gemini, before the model-use rule of 2026-10-22 and stays citable, `evaluation/change-cost-v2/RESULTS.md:5`, `AGENTS.md` design rule 3)**; E1b-v1 = `evaluation/change-cost-v1/` (superseded and invalidated, R-033; not cited as evidence); E2 = `evaluation/jev-triage-v1/RESULTS.md`; E3 = `evaluation/map-stability-v1/RESULTS.md`; CP07 = `evaluation/clawpatch-0.7-compat/RESULTS.md`; `SCB` = `/Users/adamjackson/Projects/inflight/slop-code-bench`; `FXL` = `/Users/adamjackson/Projects/inflight/factory-x-loops/factory-x-loops`; inertia-bench = `/Users/adamjackson/Projects/tools/inertia-bench` (research lab for hypothesis studies; its subject-provider contract is `docs/design/05-harness-design.md` §9, R-037). E1, E2, E3 and CP07 pre-date the experiment rules (adopted 2026-10-20, `docs/hypothesis-ledger.md` history); they are cited only as engineering observations supporting v2 decisions taken before those rules, never as evidence for or against an H-hypothesis, and no change in revision 2.1 or 2.2 rests on them. Evidence for design changes may come only from studies signed off under Scout's experiment rules (product experiments X1–X6, X9, X10) or under inertia-bench's rules, plus the grandfathered E1b-v2. Scout paths are repository-relative. Claims without a cited file or measurement are tagged **[ASSUMPTION]**. Scout never reads, writes or deletes `.codegraph`; it is read-only on the assessed code.

---

## 1. Product definition and non-goals

**Scout is a discovery-only CI guardian for a single repository.** It runs on pull requests, default-branch pushes, nightly and locally; it surfaces evidence-backed issues and descriptive measurements into a durable, triageable ledger; it blocks a PR only for a short, mechanism-certified allow-list of checks that the repository has activated and that the PR itself introduced or re-introduced (AL). **Code quality, as the hypothesised driver of long-term codebase inertia (H1–H4; §8.5 records what E1b found), is the primary family**: Scout monitors, per repository and per scope, the structural *principal* and the observed change-cost *interest* over time (§8). Architecture, security and performance remain families with lower priority.

- **Ledger**: stable issue IDs anchored to source (never to feature IDs or line numbers), audited dispositions, sanitized evidence, append-only signals with history.
- **Guardian**: differential gate (head vs. same-policy base) on `introduced`/`regressed` issues in *activated* checks; everything else reported, including `would_block` shadow annotations.
- **Mapping**: Clawpatch 0.7.3 heuristic map is the canonical feature vocabulary; the agent mapper runs at onboarding and on debounced default-branch structural events; Scout-side *areas* absorb agent-ID churn; agent IDs never enter issue identity.
- **Inertia monitoring**: per-scope time series of principal and interest, `claim: descriptive`, trend direction reported, never gated; the ledger accumulates what a future pooled analysis (run by inertia-bench from consented exports, §8.6) needs.
- **Judgment placement**: deterministic code decides eligibility, identity, gates and measurements; Jev ranks and routes (on from launch, deterministic fallback); reasoning LLMs (model CLI only, §6) are opt-in discovery; humans own dispositions, policy and activation.

**Non-goals.** (1) No remediation; req §7's approval machinery is cut. (2) No execution of the target's tests; CI artifacts are *read* when provenance-matched. (3) **No inertia score** as gate, severity input, forecast or headline: for PRs of comparable size and task type, E1b found that complexity burden and its one-year change add no predictive gain for PR author effort (E1b's exploratory P1: −0.15 %, 95 % CI [−0.69, +0.38] relative MSE) and that any degradation effect on effort or its predictability is below the 10 %-per-SD smallest effect of interest (§8.5, including what E1b did not test); Scout ships measurements and trends, not a debt-interest estimator (challenger §6, adopted). (4) No claim that agents cause degradation in an adopter's repo: the agent-era hypothesis (IR §5) is Scout's motivation, not a lane; agent provenance is recorded only when the repository supplies it. (5) No seven-language parity; a capability matrix states what is measured vs. detected vs. reviewed (§5.3). (6) No estate-level dashboarding. (7) No Enaible parity as acceptance.

**Scope (revision 2.2).** Scout is a product and a future inertia-bench subject (§8.7), not a research lab. Hypothesis studies of codebase inertia (H\*, HA\*, HQ\*) run in inertia-bench under its rules; Scout runs only product experiments (X1–X6, X9, X10; §13) under its own experiment rules.

---

## 2. Council synthesis — where the challenger changed v2

v1 §2's synthesis (agent cadence, attribution, Jev baseline, identity defects, supply chain) stands. v2 additions:

| Topic | Council leaning | Challenger | Resolution |
|---|---|---|---|
| CC-mass measure | Keep E1's `erosion` (high = healthy) | Ship the complement, higher = worse | **Adopted.** `scout.complex_burden@1`; E1 `erosion` = 1 − burden; E1b's PR-level `Q` is a related but different ratio (`scout.burden_density@1`, §8.2) |
| Clone denominator | jscpd span ÷ lizard NLOC, capped (E1 "Deviations" 6) | Same line population top and bottom, no cap | **Adopted** (`scout.dup_pct@1`) |
| Perf defect scope | Plan decode per commit | Also the run payload rewritten per commit | **Adopted.** §10.1 fixes both |
| Artifacts/PR ingestion | Phase 5 | Earlier, or principal and interest stay disconnected | **Adopted.** Phase 3 |
| Scope for inertia series | Scout areas (Phase 4) | Must not wait on agent reconciliation | **Adopted.** Heuristic features + directory fallback (E3 Jaccard 1.00; CP07 §4 100 % ID retention on tier-1 repos) |
| Trend reporting | Direction over N runs | Fixed non-overlapping windows; ≥ 3 adverse windows past a practical delta; < 20 changes = `insufficient` | **Adopted** (§8.4) |
| Pooled-study data | Export from state branch | Local collection ≠ research consent: separate opt-in, pseudonymous | **Adopted** (§8.6) |
| Moved-vs-pasted | Ship nightly | Defer until diff attribution is evaluated | **Partly.** Shipped Phase 5 as `exploratory` (X9) |
| Agent-specific signals | Jev-scored "competing implementations" | No validated deterministic measurement; narrow checks only | **Adopted.** `stale_reference_rate` ships; competing patterns is X10 |

Not followed: "outcome-first dashboard with optional structural drill-down" as the primary framing. The user's goal is to monitor *increase* of inertia, so the structural series is first-class; v2 shows both sides side by side with outcomes never derived from structure. (Revision 2.1: provisional pending §15 item 3, after E1b, §8.5.)

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
           ├─ contextual (LLM discovery through the model-CLI adapter, `claude -p`/`codex exec`; opt-in; kept, demoted; §6)
           ├─ issues (identity/matching; rewritten; §3.3)
           ├─ state (SQLite WAL + artifacts; kept; progress path fixed §10.1) + export/import
           ├─ baseline (NEW: git state-branch publish/restore; §7.2)
           ├─ selfeval (eval renamed; scorecard; §9)
           ├─ subject (NEW: `scout subject capabilities|measure`, bench-facing, snapshot-only; principles §8.7; Phase 3b)
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

`{family, trigger ∈ {pr, push, nightly}, mode ∈ {blocking-eligible, report, signal}, requires ∈ {none, tool, model-cli, decision-egress}}`. Lane *health* (ran / reused / skipped / blocked / failed) drives exit 2/3 for required lanes; findings drive exit 1. A `model-cli` lane on a runner where no model CLI is logged in reports `skipped (no model CLI)`; for these optional lanes the skip is reported in `health`, the summary and `readiness --matrix` and does not change the exit code (§7.3). The Jev decisions lane (`decision-egress`) reports health per run as specified in §6. Replaces `internal/engine/assess.go:139-171`'s hard-coded lanes.

---

## 4. Feature mapping (DECIDED 1 and 7 applied)

- **Heuristic map is canonical** (E3 §2 Jaccard 1.00; CP07 §4: 0.4.0→0.7.3 retained 100 % of IDs on fastapi, caddy, nest, webpack; axum gained 25 `rust-source-group` features and 50 content changes → one-time Rust re-review).
- **Refresh triggers (DECIDED):** both mappers (`map --source agent`) at onboarding, then on default-branch merges that change monitored areas — new unowned files, new/removed manifests or entrypoints, structural changes to owned paths — debounced into one refresh per burst (`mapping.refresh.debounce`, default 1 h **[ASSUMPTION]**). Never on PR runs; `scout map --refresh` forces one. PR/push runs reuse the active snapshot (E3 §4: 2–5 min remap; CP07 §5: 182–251 s on fastapi; 300 s default timeout marginal → `defaultMapCodexTimeoutMS = "900000"`, `internal/clawpatch/client.go:22`).
- **Areas** (kept): `{scout_area_id, title, file_set, origin, first_seen_snapshot, clawpatch_ids}`; new agent features match areas by file-set Jaccard ≥ 0.5 **[ASSUMPTION, X2]**; matched areas update file set and provenance; unmatched become new; fully heuristic-owned areas retire. Labels over Clawpatch output, never a second mapper (req §3), never identity. CP07 §5: agent-ID Jaccard 0.06–0.16 between fresh 0.7.3 runs, so reconciliation stays.
- **Measurement scopes** (new): §8 signals are keyed to heuristic features plus directory scopes for files no heuristic feature owns (the directory truncated to 2 components below the source root, fixed for a map version so a series never re-keys). E1b's per-PR area rule (prereg §2: the largest-churn file's directory, deepened while an area holds > 15 % of panel PRs) is a study clustering rule and is not used for scopes. Areas attach as labels; the time series never re-keys when an area changes.
- **Active snapshot membership**: ownership reads only the latest refresh's merged output (E3 §1: Clawpatch never deletes).
- **Unowned files**: mapping-coverage gap; still get every file-level lane and a directory scope.
- **Failure semantics**: agent mode throwing on zero features → exploration incomplete, heuristic output kept; `registry-verifier` stays disabled (CP07 §2).
- **Model CLI**: agent mapping is a Clawpatch subprocess; Scout invokes it only with Clawpatch provider `codex` or `claude`, the latter in host auth context (`CLAWPATCH_CLAUDE_AUTH_CONTEXT=host`, Claude Code ≥ 2.1.169; `third_party/clawpatch/docs/providers.md:147-163`), rejects any other configured provider, and passes no `CLAWPATCH_*_API_KEY` or provider API-key environment. Where no `claude`/`codex` login exists (for example GitHub-hosted runners), the agent refresh reports `skipped (no model CLI)`; the heuristic map stays canonical, areas and issue identity are unchanged, and the run is not failed.
- Pin: `Revision = "734f81bc1514b52a0627edb0c24ebe1a18b36105"`, `Version = "0.7.3"` (`internal/clawpatch/client.go:19-20`); patch `scout-review-sanitize-v2` sha256 `f39d73da…a865` (`review.go:19-20`, `scripts/build-clawpatch.sh:5-9`, pnpm 11.25.0). 0.8.x is a later review (CP07 §8).

---

## 5. Analyzer lanes and languages (DECIDED 5 applied)

### 5.1 Lane catalogue (priority order)

| # | Lane | Family | Trigger | Mode | Substrate | Launch languages |
|---|---|---|---|---|---|---|
| 1 | `guardrails` — `CONFIG.ignores-added.<tool>`, `CONFIG.strictness-lowered.<tool>` (effective, through `extends`), `SUPPRESS.added.unscoped`, `TESTQ.focused-test-added`; report-only: scoped suppressions, `GATE.coverage-pragma-added`, `GATE.ci-step-neutered`, hollow tests | quality | pr, push | blocking-eligible per AL table 1; report until activated | Go-native parsers for `.golangci`, `staticcheck.conf`, `pyproject`/`ruff.toml`/`.flake8`/mypy/pyright, `eslint`/`biome`/`tsconfig*`/jest; base/head diff; per-language suppression regex | **Go, Python, TypeScript** |
| 2 | `measures` — CC/NLOC (lizard), duplication (jscpd), import graph + cycles | quality / architecture | push, nightly (full); pr (changed files) | signal; `delta` rules report-only | pinned subprocess adapters (E1 "Deviations" 5–6 ran on all 7 languages); import graph per language | CC/NLOC/dup all 7; import graph **Go, Python, TS** (`tsconfig` `paths`/`baseUrl` resolution; `import type` edges flagged type-only **[ASSUMPTION]**, AL §2) |
| 3 | `history` — change-cost, coupling, ownership, hotspot, batch size, rework (§8.3) | quality | push, nightly | signal | git (non-merge, ≤ 50-file changes; bot commits excluded by the `[bot]` author suffix or an exact `history.bots` entry, never substring — §8.3) | any |
| 4 | `security` — Gitleaks 8.30.1 exact ids; Semgrep `p/ci`, `p/secrets`, `p/<lang>` | security | pr, push | blocking-eligible (`security.secrets.<rule>`, `security.ci-mutable-action-ref`) | existing adapters | any |
| 5 | `artifacts` — LCOV/Cobertura/JUnit + PR metadata, revision/producer matched | quality | pr, push | signal; `unmeasured` when absent | ingestion with provenance | any producer |
| 6 | `rubric` — 12 erosion-type SCB patterns decidable from a changed-function excerpt (`silent_fallbacks`, `retry_loops_hiding_failures`, `error_message_destruction`, `validation_theater`, `stub_implementations`, `scaffold_remnants`, `lenient_parsing`, `silent_truncation`, `auto_correction_without_warning`, `log_level_hiding`, `sentinel_values_as_errors`, `ignoring_return_values`); 45-pattern cross-product rejected (`SCB/configs/rubrics/llm_judge.jsonl` 6, 19, 20, 27, 43 need a spec or penalise Go idioms) | quality | pr (changed functions), nightly sample | report; Jev-scored; cost measured as billed `usage` per function in Phase 5 (≈ $0.12 per 200-function PR **[ASSUMPTION]**, unverified) | Semgrep where a parser decides (`ignoring_return_values`, `silent_truncation` forms, bare `except`); otherwise one Jev Noul per pattern, all fanned out in one request per changed function, each with a true/false checklist that carries the Go idiom exclusions and a per-pattern named threshold from X4 labels (never a literal), "any pattern" composed in code; a pattern that cannot reach X4's preregistered positive count is `unmeasured` and stays out of the lane | Go, Python, TS |
| 7 | `architecture-policy` — declared boundaries (`.scout/policy.yaml`: layers, cycle prohibition, max CC) | architecture | pr, push | blocking after per-repo activation | import graph + policy evaluator | Go, Python, TS |
| 8 | `performance` — Ruff PERF, CSS rule | performance | push, nightly | report | existing | Python, CSS |
| 9 | `contextual` / `clawpatch-review` | all | nightly, opt-in; only where a model CLI is logged in (`requires: model-cli`) | report | `contextual` through the model-CLI adapter (§6); `clawpatch-review` through Clawpatch provider `codex` or `claude` in host auth context (§4) | n/a |

Rule packs: corpus-tuned rules (`rules/security/semgrep-python-multipart-cve-2024-24762.yaml`, `semgrep-javascript-current-script-dom-clobbering.yaml`, `rules/performance/semgrep-css-height-transition.yaml`) become `selfeval` regression fixtures; `semgrep-actions-mutable-ref.yaml` and `semgrep-python-unsafe-yaml-load.yaml` stay.

### 5.2 TypeScript at tier 1

Guardrails = the AL table-1 TypeScript column (eslint/biome rule → `off`, ignore-list growth, `tsconfig` `exclude` growth; effective `strict`-family flags true→false; `error`→`warn`; jest `coverageThreshold` ↓; `@ts-nocheck`, bare `eslint-disable`, bare `@ts-ignore`; `.only`/`fit`/`fdescribe`). Import graph: ESM/CJS, path aliases, type-only edges separated. Fixtures and regression from webpack and nest (E1 dataset). No TS-specific precision evidence exists; TS rows follow AL track A/B like the rest.

### 5.3 Capability matrix (`scout readiness --matrix`)

| Capability | Go | Python | TS/JS | Rust | Java | C# |
|---|---|---|---|---|---|---|
| CC/NLOC/duplication; history (git); secrets/CI security | ✓ all | | | | | |
| Import graph / cycles; guardrails; rubric | P1 | P1 | P1 | P3 | P3 | – |
| CRAP (needs coverage artifact) | when produced | | | | | |
| Semantic review (contextual) | opt-in via model CLI, any | | | | | |

Rejected: seven-language parity; porting lizard; wrapping Python SCB.

---

## 6. Placement of deterministic code, decision models, LLMs, humans (DECIDED 4 applied)

| Decision | Owner |
|---|---|
| Eligibility, cache keys, identity, gate, measurements, trend flags, redaction | Deterministic Go |
| Triage ranking `P(real_defect)` (narrow questions composed in code), rubric presence (one Noul per pattern), commit classification (X5, Noul `is_fix`), unowned-file → area suggestion (X3) | Jev — advisory, **on from launch** (Appendix 4) |
| Discovery in context, FN discovery, agent mapping, second opinion on high-severity items (a deterministic rule attribute) — through `claude -p` / `codex exec` only | Reasoning LLM — opt-in; the second opinion is requested by a maintainer, never routed by Jev confidence at launch |
| Dispositions, policy, activation/demotion, area acceptance, overrides (maintainer) | Humans |

**Model use (REQ-31; R-036).** Reasoning-LLM use goes only through `claude -p` / `codex exec` with Anthropic or OpenAI models: Scout's own lanes (`contextual`) through a subprocess adapter that replaces the vendor-neutral HTTP contextual client, and Clawpatch-driven lanes (agent mapping, `clawpatch-review`) through Clawpatch's `codex` or `claude` provider in host auth context (§4). These lanes run only where the model CLI is logged in (local or self-hosted runner); the "model CLI logged in" check covers both paths, and `scripts/bootstrap-tools.sh:210-214` reports the installed CLI versions. Otherwise the lane reports `skipped (no model CLI)` (§3.4, R13). CI carries no Anthropic/OpenAI API keys. Jev is the only exception: it stays on OpenRouter, with the key from Infisical locally and a GitHub secret on trusted CI runs, never on fork PRs (§7.4).

**Adapter guardrails (`internal/decisions`).**
- *Typed questions* `noul|choice|score` as data in code. `question_id` is for code only and is never sent to the model. The version is `question_id@<sha256 of canonical JSON {instructions, criteria, state_shape, option_order_rule, resolved_model_id}>`, where `option_order_rule` is a fixed list or a permutation seeded by unit id; the realised option order is stored in the request body, not in the key. The exact request body is stored (redacted) beside every annotation. Any stored threshold, band, label-set summary or audit statistic keyed to a different version is stale: it is disclosed, and ranking falls back to the deterministic baseline until re-measured or the maintainer accepts `provisional`.
- *Pinned model.* The Jev model ID is a required `.scout/policy.yaml` field; aliases (`jev-latest`, `jev-preview`) are rejected at config load; the response's reported model is compared with the pin on every call and a mismatch is discarded (unit `incomplete`, counted). A model bump creates a new version key, so prior annotations are marked stale.
- *State and questions.* State is bounded and redacted by the `contextual` sanitizer, which is not an injection defence and is not described as one. Path role, test, fixture and generated status are computed in code and supplied as facts in state. Every question carries a data-not-instructions clause and refers to state fields by backticked path. Questions are narrow and independent ones are fanned out in one request per unit. The launch rank is a named formula composed in code over Jev outputs only (ordinal use), so changing a weight re-ranks stored answers with no model call; any term that mixes `rule weight` with a Jev probability falls under the reliability check below. A **rank-formula version** (formula id, weights, constituent question versions) and a **baseline version** (baseline formula id, rule-weight table digest, path-role classifier version) are stored with every `ranking` disclosure and every disposition; how Jev-quality statistics are keyed to them is fixed by X1's preregistration. Each Choice stores its realised option order and the harness runs it in two orders on that Choice's own labelled set to report the flip rate.
- *Output* is an `annotation` row only (`issues` has no import of `decisions`); an attacker can at most lower a rank.
- *Cap* `policy.decisions.max_per_run` default 2,000 calls (≈ $0.10 at E2's $0.05/1k) **[ASSUMPTION]**; only the call cap applies until the Phase 5 measurement of billed `usage` sets a cost bound, which will cover every question (triage and rubric); triage units are chosen in deterministic-baseline order (ties broken by issue id; `changed-in-PR` = 1 outside PR runs; both part of the baseline formula id), then the other question families (rubric functions; commit and unowned-file units once X5/X3 admit them) in a stable unit-key order (changed-file path then function start line; commit hash; file path); units a family does not reach are counted in that lane's health as `n/N`. How X1 treats capped and failed units is fixed by its preregistration.
- *Failures are failed units, never a pass or a low score.* Network errors, timeouts, HTTP 4xx/5xx, schema-invalid answers, 429/529 after exponential backoff that honours `retry-after`, `max_tokens_exceeded` and 422 (non-retryable), a missing key, a fork PR and egress off each leave the unit with no annotation, and each affected finding is disclosed `ranking: deterministic (reason)`. Units not reached because of the cap are disclosed `ranking: partial (n/N)`. Whether the OpenRouter route returns the same codes and headers as TypeSafe's documented API is undocumented **[INFERENCE]**; the adapter treats any non-success as a failed unit.
- *Lane health.* The decisions lane is optional (§7.3) and its health never changes the exit code: every unit annotated → `ran`; some units unannotated → `ran` with the per-finding disclosures above; none attempted because consent, key, egress or fork status forbids it → `skipped (reason)`; every call failing → `failed`.
- *Audit.* The nightly audit sample is a probability sample: each audited unit's inclusion probability is stored with it, and every metric that uses audit dispositions or labels on audited units is inverse-inclusion weighted. Audit items are presented blind (annotation, stratum and rank hidden). The interim draw is an independent per-unit draw at the unit's stored probability (so the stored value is the realised inclusion probability), made under a stored audit-sampler version; sampling unit, strata, fractions and floor are **[ASSUMPTION]** until X1's preregistration is signed off, which may discard interim draws. Until then a provisional audit fraction (rev 2.1's 5 % total, spread over top-ranked, bottom-ranked and unannotated units) is drawn, rank-bucket cut points are unset (the product shows rank order only) and every Jev-quality line is `insufficient`. The unranked view is always available.
- *Escalation.* No confidence band is set at launch; "high-severity" is a deterministic rule attribute. Setting a band needs a preregistered measurement of both stages on the same labelled cases, with the second opinion's cost and latency included (`claude -p`); until then the second opinion is requested by the maintainer.
- *Calibration.* `P(real_defect)` is used for ranking only. Its calibration is unchecked on Scout data (Jev's output is trained to be calibrated; Scout has not verified it). Once X1's preregistration is signed off, the product shows a rank bucket (high / normal / low) cut by rank position within the run, never the number; rank-position cuts, including the audit floor, do not read `P` and need no reliability check. Three uses would read the number and so need a preregistered reliability check first: displaying P as a percent or likelihood, a hybrid rank that weights P against `rule weight`, and any threshold on the value of P (for example an "uncertain" band). Probabilities are stored in full, never only the label.

**Launch default and fallback.** `scout init` captures consent (`decisions.enabled`, provider, egress acknowledgement; model-CLI lanes capture their own consent) into `.scout/policy.yaml`; the Jev key is a CI secret on trusted runs and comes from Infisical locally. Jev runs when enabled, key present and egress allowed for the trigger; on fork PRs or when key/egress is absent, the engine uses the **deterministic ranking baseline** (`rule weight × changed-in-PR × path role`) and discloses `ranking: deterministic (reason)`. Until X1 is recorded every Jev ranking is disclosed `ranking: jev (unvalidated, X1 pending)`; the E2 set (n = 22) is an engineering observation, not evidence for the default (Appendix 4 is unchanged). The baseline is stored for every finding regardless so X1 can measure Jev's incremental gain (E2 "Caveats" 3).

**Labels.** Maintainer dispositions are a separate `human` stream; the UI records, with every disposition, the view it was made in (audit view or not) and whether the annotation was visible. Labels for any Jev question (eval sets, threshold tuning, X1/X3/X4/X5/X10) follow the protocol of `docs/STATUS.md` §6.8: blind (labellers and the user's review sheet see only the unit's evidence, never Jev's answer, rank, score, stratum or a prior verdict); a one-page label guide per question, versioned with the question's version key and frozen before labelling, with synthetic worked examples; two independent sessions with no shared context and no tools (one Anthropic via `claude -p`, one OpenAI via `codex exec`) and a third adjudicating session that records why where they disagree; then the user confirms or changes every case, with no bulk accept. A label is `agent_reviewed` before that review and `human` after it (`by: user`, plus the agent proposal); labels are appended, never edited, and the latest `human` label wins. No Jev-versus-label comparison, threshold, band or default change is made outside a council-signed preregistration (`AGENTS.md`). Each experiment's preregistration fixes its estimator, labels, volumes, minimum precision and recall, costs of false actions and misses, written policy and any escalation accuracy loss.

Rejected: Jev as identity oracle; Jev closing `false_positive` (E2 "Recommendation"); a single bundled Noul for 3-way routing, rejected for 3-way routing only (E2, n = 22, pre-rules, did not show a Noul worse for ranking, `RESULTS.md:27-38`, and never tested one Noul per reason); `P` read as calibrated confidence before a reliability check on Scout data (E2 Caveat 6).

---

## 7. CI guardian contract (DECIDED 2 and 3 applied)

### 7.1 Triggers

`pull_request` (differential; `report` or `guardian`), `push` to default branch (full or incremental; publishes baseline; evaluates refresh triggers), `schedule` nightly (full; trend rules; opt-in model-CLI lanes where a CLI is logged in), `workflow_dispatch` (`full`, `refresh-map`, `authorize-egress`). Local: `scout assess`, `scout diff --base <ref>`.

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

`0 clean / 1 gated / 2 incomplete / 3 unavailable`, precedence 3 > 2 > 1 > 0; JSON always has `gate`, `health`, `would_block`, `ranking`. Outputs: `scout-assessment.json`, ≤ 60-line Markdown summary (introduced/regressed, would-block, pre-existing count, lane health, signal deltas), optional SARIF. Reusable workflow `scout-guardian.yml` + composite action; binary and Clawpatch bundle from a **pinned release**, never built from the PR checkout (today's `scout-nightly.yml` builds `./cmd/scout` from the checkout); policy read from the *base* branch; `contents: read` on PR, `contents: write` scoped to `scout/state` on trusted runs. CI carries no Anthropic/OpenAI API keys; the only secret is the Jev key (`OPENROUTER_API_KEY`), a GitHub secret on trusted runs and absent on fork PRs; model-CLI lanes run only where a CLI is logged in (local or self-hosted runner), and a skipped optional lane is reported in `health` and the summary without changing the exit code (R13). First adoption: `scout init` → consent → full default-branch run → baseline; nothing is `introduced` before a baseline exists.

---

## 8. Inertia monitoring: principal and interest over time

### 8.1 Framing and claim level

Inertia = a codebase's accumulated resistance to change (IR §1). Scout operationalises the debt metaphor: **principal** = structural measures of a scope at a snapshot; **interest** = observed change-cost outcomes for the scope over a trailing window; **exposure** = how often the scope changes (debt costs only where code moves; IR §3). The user's hypotheses — degradation raises later change effort/complexity (H1), unpredictability (H2), bug introduction (H3) and maintenance complexity, and agents magnify this — are the *reason to measure*, not something Scout asserts per repository.

Every measure here is **`descriptive`**. Scout reports levels, deltas and trend direction; it never gates, sets severity, forecasts, computes a composite, or upgrades `claim` on its own (§8.5). What launches is a versioned, comparable record of both sides over time per scope, with population and missingness stated. Within its scope (§8.5), E1b found no material per-PR author-effort penalty from complexity burden or its change, so principal series are presented as *structure to watch*, not as cost already being paid.

### 8.2 Principal measures (snapshot; function → file → scope)

Scope = heuristic feature or directory scope (§4). Eligible population = production source by E1's extension/exclusion rules (`evaluation/inertia-v1/exclusions.json` pattern), NLOC ≥ 10, generated/vendored excluded by path role; the population digest is stored with every signal.

| Metric | Formula | Source | Ships |
|---|---|---|---|
| `scout.cc@1` | per function CC, NLOC; scope `n_high_cc` = #functions CC > 10, `cc_p90`, `cc_max` | lizard (pinned) | Phase 3 |
| `scout.complex_burden@1` | B = Σ_{f: CC_f>10} CC_f·√NLOC_f ÷ Σ_f CC_f·√NLOC_f, higher = worse (E1 `erosion` = 1 − B) | lizard | Phase 3 |
| `scout.dup_pct@1` | duplicated eligible source-line positions ÷ eligible source-line positions, one tokenizer population, no cap; clone groups kept with cross-scope flag | jscpd 4.0.8 (min-tokens 50, min-lines 5, `--max-lines 10000 --max-size 500kb`) | Phase 3 |
| `scout.cycle_mass@1` | internal import edges on a directed cycle ÷ internal import edges; `cross_boundary_cycles` separate; `unresolved_imports` reported; type-only edges excluded and counted | import graph (Go, Python, TS) | Phase 3 tier 1; `unmeasured` elsewhere |
| `scout.crap@1` | per function CRAP = CC² × (1 − cov)³ + CC, cov ∈ [0,1]; scope `crap_load` = Σ_{CRAP>30} CRAP; `uncovered_cc_mass` = Σ CC_f·(1−cov_f) ÷ Σ CC_f | lizard + revision-matched LCOV/Cobertura | Phase 3 when the artifact exists; else `unmeasured`, never 0 |
| `scout.propagation_cost@1` | mean fraction of scope nodes reachable via imports (MacCormack 2006); Decoupling Level deferred | import graph | Phase 5 |
| `scout.cognitive_cc@1` | SonarSource nesting-weighted complexity | gocognit / complexipy / eslint-sonarjs **[ASSUMPTION: adapters]** | later |
| `scout.stale_reference_rate@1` | backticked identifiers/paths in `AGENTS.md`, `CLAUDE.md`, `README*`, docstrings resolving to no file/symbol ÷ references checked | regex + inventory + lizard symbols | Phase 5, `exploratory` (IR §5; deterministic) |
| `scout.burden_density@1` | Σ_f hi_f ÷ Σ_f NLOC_f, hi_f = Σ_{fn: CC>10} CC·√NLOC_fn, over the scope's eligible files ≥ 10 NLOC — the same ratio as E1b's per-PR `Q` (prereg §3.1), computed on the scope population under §8.2 eligibility, so values are not interchangeable with E1b's | lizard | Phase 3 |

Deltas are computed between consecutive snapshots on the *same population digest*; membership change is a separate `population_delta`, never degradation (challenger §6). PR runs apply `delta` rules on changed files only (e.g. "CC 9 → 14"), report-only.

### 8.3 Exposure and interest measures (trailing window; per scope)

Windows: fixed, non-overlapping, length `W` (default 90 days **[ASSUMPTION]**), calendar-aligned so successive windows compare; a trailing 365-day view is also stored. **Commit-level measures**: change = non-merge commit, non-bot, ≤ 50 files, touching ≥ 1 eligible file in scope. Bot commit = author name ending in the literal suffix `[bot]` (GitHub's reserved naming for App identities, the git-side counterpart of E1b's API `Bot` type) or author name/e-mail exactly equal to an entry of `history.bots` (repository policy; empty by default); no other pattern matching (E1b-v1's substring regex `(?i)(bot|renovate|dependabot)` is retired). **PR-level measures** (I4, `batch_size`, the active-hours part of I5): merged, non-bot PRs with ≤ 50 changed files and ≤ 100 commits whose file list (GraphQL `PullRequest.files`, as E1b `fetch_prs.py`) contains ≥ 1 eligible file in scope; bot PR = API author type `Bot` or login in E1b's exact list `common.BOT_LOGINS` (prereg §2). A PR over either cap, or whose file list or commit list the API returns incomplete (fewer entries than its total count, or no commits), is marked `linkage: incomplete`, counted and excluded (as E1b `collect.py:146-149`); bot PRs are counted and excluded (`collect.py:144-145`). The change table records `bot_reason`: `app_suffix` or `configured` for commits, `api_bot_type` or `login_list` for PRs.

| Metric | Formula | Source | Ships |
|---|---|---|---|
| `scout.changes@1` (exposure) | changes; distinct authors | git | Phase 3 |
| `scout.hotspot@1` | Σ_files (Σ_f CC_f) · log1p(changes to file) ÷ scope NLOC | lizard + git | Phase 3 |
| `scout.churn_per_change@1` (I1) | p50, p90 of added + removed lines in scope files per change | git | Phase 3 |
| `scout.amplification@1` (I2) | p50 eligible files touched repo-wide per change touching the scope; p50 scopes touched | git | Phase 3 |
| `scout.follow_up_rate@1` (I3) | changes followed ≤ 7 d by a change touching ≥ 1 same file ÷ changes with a complete 7-day observation | git | Phase 3 |
| `scout.pr_effort@1` (I4) | p50 and p90 of **author active hours** (git-hours sessions on the author timestamps of the PR's own commits, from the PR commit list: gaps ≤ 4 h add the gap, a longer gap starts a session, 0.5 h credit per session; single-commit PRs sit at the 0.5 h floor and are counted separately — E1b prereg §3.3; G3b ρ 0.46 with churn); commits/PR; review submissions + comments/PR; created→merged hours as **queue time**, reported separately and never as effort (E1b's effort construct excludes review and queue waiting, prereg §3.3); `unmeasured` < 5 linked PRs | PR commit list, `pull_request` payload + API (`pull-requests: read`) | Phase 3 in CI; `unmeasured` locally (squash merges remove PR commits from local history) |
| `scout.dispersion@1` (I5) | p90 ÷ p50 of churn (git, always) and of PR active hours (`unmeasured` whenever I4 is), when p50 > 0 | git, PR | Phase 3 |
| `scout.change_coupling@1` | changes touching scope and ≥ 1 other scope ÷ changes touching scope; pairwise support and P(b\|a) for top pairs; `unexplained` subset = pairs with no import edge (tier 1 only) | git (+ import graph) | Phase 3; `unexplained` Phase 5 |
| `scout.ownership@1` | top author's changes ÷ scope changes; author count; minor contributors < 5 %; pseudonymised keys, never exported as names | git | Phase 3 |
| `scout.rework@1` | removed lines with blame origin inside the window ÷ removed lines (E1 "Deviations" 7) | git blame | Phase 3 |
| `scout.batch_size@1` | lines and files per PR (p50, p90) | PR metadata | Phase 3 |
| `scout.moved_vs_pasted@1` | per change: added blocks ≥ 5 lines matching a removed block in the same change = `moved`, matching base-tree lines = `pasted`, else `new`; window shares | git diff + jscpd on hunks | Phase 5, `exploratory` (X9) |
| `scout.fix_touches@1` (H3 proxy) | `fix`-classified changes touching scope ÷ changes | Jev Noul `is_fix` with a tuned threshold (X5); a binary, not a 5-class argmax | after X5; regex precision 0.52 (E1) is disqualifying; E1b-v2's G4 (`evaluation/change-cost-v2/validation-gates.json`) has `fix→fix` 24 of 30 pipeline-`fix` rows (precision 0.80, Wilson 95 % [0.63, 0.91], derived here) on PR-level state: a prior, not evidence for commit-level input |

### 8.4 Reporting over time per repository

- **Side-by-side panel per scope**: principal levels and Δ, exposure, interest, with `n changes`, population digest, missingness. `n < 20` changes in a window ⇒ interest = `insufficient`, never shown as good **[ASSUMPTION]**.
- **Adverse-trend flag** (`trend` rule, informational, nightly): a metric moves adversely across ≥ 3 consecutive comparable windows by more than a practical delta (defaults `complex_burden` +0.05, `dup_pct` +2 pts, `cycle_mass` +0.05, `n_high_cc` +20 % **[ASSUMPTION]**) on a stable population; same for interest metrics. Lehman's mixed evidence (IR §1) means *direction* is the finding; no inevitability implied.
- **Inertia watch list** (descriptive, never a composite): scopes with an adverse principal trend AND top-quartile exposure — an adverse structural trend in frequently changed code (IR §4), presented as where to look. It is not a cost claim (§8.5).
- **Interventions**: `scout annotate --window … intervention: refactor|migration|…` keeps a legitimately churn-heavy window out of trend flags and records it as an event (challenger §6).
- Every panel carries `claim: descriptive` and a link to §8.5.

### 8.5 Honest reading of the evidence, and what would upgrade a claim

**E1b (R-035; council-signed preregistration and Amendment 1, run 2026-10-21).** Unit = merged PR; 7,717 PRs in 12 open-source repositories, 480 directory areas; author active time (git-hours sessions) as the effort outcome (G3b-validated); within-repository standardised predictors at the PR's start snapshot; task type, size, author experience, quarter and repository/area fixed effects; delete-one-area jackknife with full refit; one Holm layer; SESOI 0.10 per SD — log-hours for the effort tests (T1, T2), relative dispersion for the predictability tests (T3, T4); power simulated on the real design (empirical MDE80 0.042–0.063 per SD in each test's units for T1–T4, all adequate).

| Question | Result |
|---|---|
| H1 — change to code whose burden **rose** over the prior year costs more active time (T2, area FE) | **equivalent**: +0.023 log-h/SD (≈ +2.3 %), 90 % CI [+0.004, +0.041], Holm p 0.165 — an effect ≥ 0.10 log-h/SD is excluded; a small positive association is compatible |
| H2 — such a change is less predictable (T4) | **equivalent**: +0.006 relative dispersion/SD, 90 % CI [−0.010, +0.023] |
| Burden *level* → effort / predictability (T1, T3; associational) | equivalent: T1 −0.011, 90 % CI [−0.029, +0.007]; T3 +0.018, 90 % CI [−0.003, +0.038] |
| Incremental prediction of PR effort from burden and its change (P1, exploratory) | −0.15 %, 95 % CI [−0.69, +0.38] relative MSE; baseline beats a constant (0.333) |
| H4 — hidden coupling → corrective follow-up (T5) | **not decided**: the Jev follow-up labeller's IPW recall (G5, 0.422 < 0.60) and the hidden-coupling proxy (G7, pair κ 0.514 < 0.6, hidden precision 0.630 < 0.80) failed validation |

Sensitivity analyses (session gap 2 h/8 h, leave-one-repository-out, cohorts, author clustering) keep T2 and T4 at or below 0.061 at the 95 % upper bound (largest: T2 in PRs merged before 2024, 0.0603); per-repository T2 varies (pandas +0.096 [+0.048, +0.144], vite +0.061, tokio −0.114), exploratory.

**What this does and does not settle.** For merged PRs of comparable size and task type in these repositories, complexity-burden density and its one-year change carry no *material* author-effort penalty (≥ 0.10 log-hours per SD) and no material predictability penalty (≥ 0.10 relative dispersion per SD). Three limits are built into that statement: the tests are associational, not causal (time-varying confounding remains; prereg §6, RESULTS §6); E1b controls for size, so a burden effect that works by making changes larger (amplification) is outside its estimand (prereg §6); and 38 % of the 7,030 PRs with a measured `Y_active` sit at the single-commit floor, so T2 and T4 are not separable and are read together (council decision b). E1b does not test: (1) other quality measures — duplication, import cycles, cohesion, coverage, architecture erosion (it used one, CC-weighted density); (2) effort outside commits — reading, design, review (active time is a commit-timestamp proxy; a comprehension cost before the first commit is unobserved) [INFERENCE]; (3) bug introduction (H3); (4) work that never merges; (5) agent-authored change (HA1–HA3; inertia-bench); (6) hidden coupling, whose measurement failed validation and needs a better dependency oracle before it is retested.

Consequences for Scout (provisional pending §15 item 3): principal measures stay `descriptive` and are framed as structure to watch, never as cost (§8.1, §8.4); the interest side uses author active time as effort and reports queue time separately (§8.3 I4); `change_coupling.unexplained` stays Phase 5 and `exploratory` until a validated dependency check exists (G7 hidden precision 0.630 < 0.80); no metric's `claim` changes.

**Further tests belong to inertia-bench.** The tests that would go beyond E1b — a pooled adopter panel with Scout's own definitions and the quality measures E1b did not test, controlled continuations (HA2), within-scope event studies around recorded interventions, and defect introduction (H3) once X5 validates a fix classifier — are hypothesis studies (H\*/HA\*/HQ\*; the Scout ledger is frozen as of 2026-10-22) and run in inertia-bench under its rules (R-037; `AGENTS.md`). Scout's part is the data it accumulates (§8.6, with Scout-owned consent) and the subject interface (§8.7).

Decision rule, fixed now: a metric's `claim` moves to `predictive` for a named outcome only by a preregistered, council-signed pooled test showing ≥ 5 % relative MSE reduction with CI lower bound > 0, consistent across ≥ all-but-one repositories with ≥ 10 units, multiplicity-controlled. The qualifying study may be signed off under Scout's experiment rules or under inertia-bench's rules (`AGENTS.md` design rule 3); no such study is planned in Scout (see above). Causal claims need a controlled-continuation or within-scope event-study design signed off the same way. Gate use is a further separate decision (AL). Repeated dashboard significance never upgrades a claim.

### 8.6 Accumulating evidence from adopters without overclaiming

Per default-branch run Scout stores (SQLite and `signals/inertia.jsonl` in `scout/state`): snapshot revision/time, window bounds, scope ID, population digest, map version and scope lineage (split/merge), every §8.2/§8.3 value with numerator, denominator, tool/config digests, import-resolution coverage, artifact revision/producer, missingness reason, outcome maturity (right-censoring), and a compact sanitized change table (deduplicated change/PR keys → scope, churn, files/scopes touched, timestamps, workflow type, bot flag and `bot_reason` (§8.3), linkage completeness, optional task category, human/agent provenance only when the repository supplies it). Medians alone cannot support re-analysis; the change table makes alternative outcome definitions possible later. Bulky evidence stays in artifacts.

`scout export --study` is a **separate opt-in** from local collection, a product feature with Scout-owned consent (§15 item 1): schema in `docs/evidence/study-export.md`; repository-scoped pseudonyms; scope-level aggregates plus the change table; no source, commit messages, emails or reviewer identities; retention and withdrawal stated. Pooled analyses of consented exports are run by inertia-bench, not by Scout. Public `scout/state` branches carry a notice that git history makes erasure hard. Overlapping scopes are flagged so one change never multiplies pooled observations. Ownership is stored as coarse aggregates, never a personnel ranking.

Per-repository re-estimation of E1b's tests is not offered: the change table does not carry E1b's predictors and controls (start-snapshot burden and its one-year change, task type, author experience). Adding them is a separate design change.

### 8.7 Subject interface (principles only)

Scout is a future *subject* of inertia-bench: a measurement provider that the bench runs on a read-only snapshot, offline (R-037; `inertia-bench/docs/design/05-harness-design.md:879-920`, whose design is not yet signed off, review round 6: issues found). The Scout binary will provide `scout subject capabilities` and `scout subject measure`. This revision adopts only the principles below; the exact schema (`ib-subject-*/1`) is **deliberately not adopted here** and needs its own council round after the bench's design signs off. Implementation follows Phase 3 `measures` (Phase 3b, §11).

1. **Snapshot-only measures.** Only §8.2 measures, identified as `metric@version`, computed from the snapshot alone; each bench measure id is one scalar derived from a `metric@version` (for example `scout.cc@1` yields several scope values), with the ids fixed in the schema round. History and PR measures (§8.3) are out of the first version.
2. **Offline.** No network access and no model call; a measure whose tool or input is unavailable is `unmeasured`.
3. **Deterministic.** Two runs on the same snapshot and request produce byte-identical output (sorted, with no timestamps or run identifiers).
4. **Read-only on the snapshot.** The command never writes into the snapshot; any working state lives in a temporary directory removed on exit.
5. **Explicit missingness.** One result per requested (unit, measure) pair; `unmeasured` is never 0. Scout's `unmeasured` reasons map onto the bench's closed `unmeasured_reason` set (as that set stands in the unsigned bench design: unsupported language, parse error, timeout, not applicable, other). In the first version the subject interface has no channel for a coverage artifact (the bench request carries none), so `scout.crap@1` is `not_applicable`; supplying an artifact and establishing its revision match wait for the schema round. The exact mapping is fixed by the schema round.
6. **Blind.** Scout receives opaque snapshot and unit identifiers and never sees arm labels or outcomes.

Subject measures never change a metric's `claim` (§8.5).

---

## 9. Scout's own outcome scorecard

Each line `count/attempts` with a Wilson 95 % interval where the line is an unweighted binomial (for example precision on an equal-probability sample of the hit population, as E1b's G5), and `unmeasured` counts (FXL `docs/references/scoring.md`); weighted or paired estimators (inverse-inclusion recall, AUROC, paired differences) use a design-based interval (stratified or clustered bootstrap, level 95 %) named in each preregistration; no single solve rate.

1. **Actionable yield** per lane/rule: `accepted` or resolved-by-change ÷ dispositioned; `open` separate.
2. **Follow-through**: confirmed issues resolved within 30/90 days ÷ confirmed with a mature window.
3. **Scoped assurance**: random `(file, lane)` units called clean, independently audited → false-assurance rate per lane; no repository-wide FN claim.
4. **Cost of use**: unwarranted blocks (AL retention), gate-disable events, overrides, duplicate/re-opened issues per 100, time-to-disposition, runtime and spend per PR, agent-refresh value.
5. **Inertia-monitoring usefulness** (new): adverse-trend flags and watch-list entries marked `useful|not useful` by a maintainer; `undecided` separate. Attention value, not validity.
6. **Change-success rate** (experimental, descriptive): behind X5 only.
7. **Triage-model quality**: Jev vs. the deterministic baseline (AUROC, routing precision) from audit-view blind dispositions only, with inverse-inclusion weights; otherwise `insufficient`, and withheld until X1's preregistration (which defines the line's inputs, repositories and period) is signed off. Never a threshold, band or default.
8. **Corpus regression**: 9 true / 13 FP (`evaluation/reference-v1.json`) stay detected / non-flagged.

---

## 10. Keep / cut / rewrite

### 10.1 P1 performance defect — `scout assess` run-job persistence

Observation (user-reported, ground truth): after the 0.7.3 upgrade smoke, `scout assess` on fastapi ran > 1 h without finishing; the goroutine dump shows `CommitRunJob → loadPlanTx` re-reading and decoding a ≈ 5 MB plan per job commit.

Code (verified): `internal/state/progress.go:117-200` `CommitRunJob` runs per job in one transaction: `loadRunProgressTx` (decodes the run payload), `loadPlanTx` (`:331-346`: `SELECT payload FROM plans`, `decode`, `plan.Validate()`), `plannedJob` (`:389`, linear scan), then appends the commit's coverage/failures/artifacts to `run`, re-encodes the whole run and `UPDATE runs SET payload` (`:189-197`). Plans are insert-only (`internal/state/issues.go:34-37`, `ErrImmutable`), so a plan never changes after `SavePlan`; `Plan` carries the full `Snapshot` inventory (`internal/domain/types.go:639-647`), the source of the 5 MB. Cost for J jobs, plan size P, mean result r ≈ O(J·P + J²·r) [INFERENCE from code; not profiled].

Fix (minimal, both paths): (1) at `SavePlan` also write `plan_jobs(plan_id, job_id, payload)` plus the plan's identity fields; `CommitRunJob` reads one job row and checks `plan_id/run_id` without decoding the snapshot (an in-process decoded-plan cache keyed by `plan_id` is an acceptable first step because plans are immutable, if private to the `Store`, bounded, synchronised and hydrated on restart); (2) stop rewriting the run payload per commit — coverage already lives in `coverage`, the commit in `run_job_commits`; `LoadRun`/`LoadInProgressRun`/`FinishRun` reconstruct progress from those tables and assemble the final payload once. Preserve initial reused coverage from `BeginRun` (`:79-114`), atomicity, duplicate-commit rejection, resume-after-interruption equivalence.

Acceptance (Phase 1): a `go test ./internal/state` benchmark commits 5,000 jobs against a plan with a 5 MB snapshot and asserts per-commit time and allocations flat in J, with no `plans` read and no `runs` payload write in the commit path (SQL trace counts); interruption-resume test yields the same finalized run; `scout assess` on the pinned fastapi revision (E3 `d666ccb`, CP07 §1), heuristic map, egress off, completes end-to-end in < 10 min on the development Mac **[DECIDED 2026-10-22, Appendix 14; the > 1 h failure is accepted as reported, not rerun]**.

### 10.2 Code verdicts

| Package / file | Verdict |
|---|---|
| `internal/inventory/*`, `internal/analyzers/{runner,spec,parse,sanitize,ruff,semgrep,gitleaks}.go`, `internal/state/{store,artifacts}.go`, `internal/web/web.go`, `scripts/build-clawpatch.sh` + patch | Keep (inventory: lanes must not depend on `MappingRequired`; add paired base/head plan entry; state: add export/import; web: inertia panels later) |
| `internal/contextual/*` | Keep the sanitizer (reused by `decisions`), lane logic and opt-in nightly use; **rewrite the provider**: replace the HTTP client (`client.go`, `SCOUT_PROVIDER_*`) with a subprocess adapter over `claude -p` / `codex exec` (§6) |
| `internal/state/progress.go` | **Fix** (§10.1), then keep |
| `internal/state/{issues,issue_observations}.go` | Rewrite identity-sensitive parts: positions/quote out of `observationSupportKey` (`issue_observations.go:94-102`); wipe (`issues.go:455-461`) → `evidence_drift`; drop `ApprovalRef`/`ProposedAction` |
| `internal/issues/correlate.go` | Rewrite matching (§3.3); keep `SanitizeText`/provenance |
| `internal/domain/types.go` | Extend: `Signal` (window, population digest), `Area`, `Scope`, `Annotation`, `Attribution`, `LaneHealth`, `WouldBlock`; remove approval types |
| `internal/engine/{assess,readiness,config}.go` | Rewrite orchestration: lane registry, profiles, base-branch policy, consent/fallback |
| `internal/analyzers/catalog_*.go` | Re-curate: lizard/jscpd specs; import graph + policy evaluator (architecture currently `return nil`); packs + exact Gitleaks ids; corpus-tuned rules → fixtures |
| `internal/clawpatch/{client,records,review}.go` | Keep boundary; add explicit `source`/`refresh` + debounced trigger predicate; heuristic must run offline (`client.go:168-170` blocks it without egress today); **provider constraint**: map and review invoke Clawpatch only with `--provider codex` or `--provider claude` (host auth context for `claude`), reject any other provider and pass no API-key environment (§4) |
| `internal/{measures,history,artifacts,guardrails,decisions,baseline,subject}` | Create (§§5–8; `subject` after Phase 3 `measures`, §8.7) |
| `internal/eval/*` → `internal/selfeval` | Rename; scorecard (§9) |
| `internal/codex/exec.go` | Cut (no callers) |
| `cmd/scout/main.go` | Rewrite gate (`--fail-on`, `:202-207` → profile/attribution); add `init`, `diff`, `map --refresh`, `export/import [--study]`, `signals`, `trends`, `annotate`, `selfeval`, `readiness --matrix`, `subject capabilities|measure` (§8.7); remove `approve` |
| `.github/workflows/scout-nightly.yml`, `scripts/bundle-state.sh` | Rewrite → `scout-guardian.yml` + dogfood caller (drop AWS/OIDC and the `SCOUT_PROVIDER_*` settings and `SCOUT_PROVIDER_TOKEN`, the `OPENAI_API_KEY`/`ANTHROPIC_API_KEY` entries fed from `CLAWPATCH_{OPENAI,ANTHROPIC}_API_KEY`, and constrain `CLAWPATCH_PROVIDER` to `codex`/`claude`; add `pull_request`); bundle script → `scout export/import` |

### 10.3 Spec cutover (DECIDED 6)

`SCOUT_REBUILD_REQUIREMENTS.md` is archived under `docs/evidence/` and replaced by `docs/scout-product-contract.md` (§§1, 3, 7, 8.4), `docs/scout-adapter-contracts.md` (tool adapter; Clawpatch command contract amended for refresh triggers and for the provider constraint of §4 (`codex`/`claude` only, host auth context, no API-key environment); typed-decision adapter; model-CLI adapter for reasoning lanes and the Jev (OpenRouter) provider; identity/matching; incremental planning; signal/measure contract) and `docs/evidence/` (Enaible inventory, corpus notes, `study-export.md`). The subject-interface contract (§8.7) enters those documents only through its own council round. Kept from req: §8 gap→rule promotion with FXL's FCR record shape; §12.8 fixture classes plus AL attribution fixtures. Until Phase 6 the requirements file stays live; its REQ-29, 31, 35, 36, 38, 39, 41 and §10 are synchronised with this revision in the same change set.

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
Acceptance: `scout signals --scope` shows numerator/denominator/population digest; CRAP present only with a matching artifact, `unmeasured` otherwise; `pr_effort` `unmeasured` locally, present in CI; `delta` issues carry `claim: descriptive`; all four guardrail mechanisms fire with base/head evidence in all three languages; TS import graph resolves `paths` aliases and separates type-only edges; 9/13 corpus regression holds; `readiness --matrix` prints §5.3; given the PR commit author timestamps in `evaluation/change-cost-v2/raw/prs-{nest,webpack}.json.gz`, Scout's active-hours function reproduces `active_hours` in `evaluation/change-cost-v2/panel-final.csv.gz` for every panel PR of those two repositories (joined on `repo`, `number`) within 1e-9 h (reference: `build_panel.active_hours`).

**Phase 3b — Subject interface** (`subject`; after Phase 3 `measures`; schema gated on inertia-bench's design sign-off, §8.7)
Scenario: on a Go, Python and TypeScript fixture snapshot mounted read-only with the network denied, run `scout subject capabilities` and `scout subject measure` twice on the same request, including a file in an unsupported language.
Acceptance (principles only): the two runs are byte-identical; the snapshot tree digest is unchanged; no network or model call is made; every requested (unit, measure) pair has a result and `unmeasured` is never 0; `scout.crap@1` is `not_applicable` (no artifact channel in the first version, §8.7 principle 5); history and PR measures are absent from `capabilities`. Entry: Phase 3 `measures` done. The schema (`ib-subject-*/1`) is adopted by a separate council round after inertia-bench's design signs off, and Phase 3b does not complete before that.

**Phase 4 — Mapping refresh triggers and areas**
Scenario: fastapi and axum fixtures; onboarding refresh; three default-branch merges in quick succession (new unowned package, new manifest, content-only); then a content-only PR.
Acceptance: exactly one debounced refresh after the burst, none after the content-only merge or on the PR; prior issues keep IDs; area file-set Jaccard across refreshes ≥ 0.8 **[ASSUMPTION, X2]**; stale Clawpatch records excluded from ownership; unowned files listed with directory scopes; inertia series keys unchanged by the refresh; agent failure leaves heuristic output and marks exploration incomplete.

**Phase 5 — Trends, exploratory measures, rubric lane, self-evaluation**
Scenario: replay four Phase-3 windows plus a fifth adverse one; mark one window `intervention: refactor`; prompt-injection fixture; scorecard on Phase 2–3 ledgers; `moved_vs_pasted`, `stale_reference_rate`, `propagation_cost` on the corpus. A Jev replay of labelled sets runs only as an analysis named by a signed-off preregistration.
Acceptance: trend flag fires only after three adverse comparable windows and not across the intervention window; `n < 20` windows show `insufficient`; watch list sorted without a composite; annotations carry the resolved model and the question version key; a model bump marks prior annotations stale and ranking falls back disclosed; forced mismatch → `incomplete`; a simulated 429, timeout or cap hit never yields a low rank and is disclosed `ranking: deterministic (reason)` or `partial (n/N)`; the order-flip rate of each Choice is reported; no state/gate change from any annotation; the injection fixture (a critical-severity finding) changes at most its rank; the nightly audit sample is drawn with a stored inclusion probability per unit and presented blind (annotation, stratum and rank hidden), and on a synthetic frame the per-unit selection frequency matches the stored probability; scorecard prints the §9 intervals (Wilson or design-based per line); Jev-quality lines are `insufficient` until X1's preregistration is signed off; exploratory metrics tagged and excluded from trend flags.

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
| R4 | Hypotheses never testable from Scout data (few adopters, no consent) | medium / medium | Export schema fixed by Phase 6; PR effort defined as in E1b (author active hours); inertia-bench runs adopter-independent studies (controlled continuations on curated repositories) and the pooled analyses of any consented exports |
| R5 | State-branch permissions unavailable or history bloat | medium / medium | Squash-safe JSONL; S3 adapter later (DECIDED order) |
| R6 | Metric drift mistaken for change (tool upgrade, re-partition, population churn) | medium / medium | `metric@version` + tool and population digests; scope keys independent of agent areas |
| R7 | Guardian supply chain: PR controls its examiner | low / high | Pinned release, base-branch policy, read-only PR tokens, no secrets on fork PRs |
| R8 | Persistence fix regresses resume semantics | low / high | Interruption-resume equivalence test in Phase 1 |
| R9 | Clawpatch pin ages (0.8.x); patch re-anchor | medium / medium | `build-clawpatch.sh` verifies patch; each upgrade = compatibility review + reduced E3 (CP07) |
| R10 | Jev ranking hides findings, or is silently absent | low–medium / medium | Route-only, unranked view, disclosed deterministic fallback, X1; version key with stale-on-bump; blind probability-sampled audit with stored inclusion probabilities |
| R11 | Adopters refuse `pull-requests: read` | medium / low | `pr_effort`, `batch_size` and the active-hours part of `dispersion` `unmeasured`, never 0 |
| R12 | Ownership or change tables expose contributors | low / high | Pseudonyms, coarse aggregates, separate study consent, public-branch notice |
| R13 | GitHub-hosted adopters get no reasoning-LLM lanes (`contextual`, `clawpatch-review`, agent mapping): the model CLIs need a login and CI carries no Anthropic/OpenAI keys | high / medium | Lanes report `skipped (no model CLI)` in `health`, the summary and `readiness --matrix` without changing the exit code; heuristic map, deterministic lanes and Jev ranking are unaffected; the lanes run locally or on a self-hosted runner |
| R14 | Jev runs through OpenRouter (REQ-31): the `alpha/decisions` route, its dated model slugs and their retention are not documented by TypeSafe (whose documented endpoint is `https://api.typesafe.ai/v1/systemone`; R-038, review §D q12); a re-pointed or retired slug, or different error codes and headers, would break the pin | unknown / medium | Resolved-model check on every call (mismatch → discarded, `incomplete`); the resolved model is in the version key so a roll marks annotations stale; deterministic fallback keeps the product working; the route is re-checked at each model bump; moving to the TypeSafe endpoint is a user decision not taken here |

---

## 13. Experiments with decision rules

| ID | Question | Design | Decision rule |
|---|---|---|---|
| X1 | Jev vs deterministic ranking on triage retrieval | Question: does Jev's rank retrieve true defects better than the deterministic baseline at a fixed review budget? Primitive: narrow Noul questions composed in code (§6). Labels: §6.8 protocol on probability-sampled audit units. Estimator, split unit, population, sampling strata and draws, audit floor, `k`, interval, treatment of capped and failed units, versioning of verdicts and test-set reuse, and sample size are fixed by its council-signed preregistration (`AGENTS.md` experiment rules) | Jev stays default-on iff held-out recall@k gain ≥ 10 points with the interval's lower bound > 0, and no critical true issue lies below the audit floor; else default = deterministic baseline, `--rank jev` opt-in (route-only is not reopened). The 10-point gain is a rev 2.1 value, pending the user's minimums; the preregistration confirms or replaces it |
| X2 | Do areas stabilise agent output? | E3 protocol with reconciliation, fastapi/axum, 3 fresh + 2 same-state + 2 mutations | Adopt the Jaccard threshold yielding survival ≥ 0.8 and ≤ 5 % spurious merges; else refresh only at onboarding + manual |
| X3 | Jev assigns unowned files to features? | Primitive: code retrieves top-K candidate features per file (path proximity, import edges) or supplies the full list when ≤ 255 features; a Choice over them with a permutation option-order rule (§6); `none` decided by a companion Noul ("does any listed feature own this file?") read before the Choice, with no `none` option. Labels: unowned files (rev 2.1: 200), §6.8 protocol. Estimator, split, threshold tuning, candidate-coverage and recall reporting, interval and sample size are fixed by its council-signed preregistration | Reconciliation aid iff precision ≥ 0.9 (files assigned the correct feature ÷ files assigned any feature). Rev 2.1 value, pending the user's minimums (review §D q5, q6); the preregistration confirms or replaces it |
| X4 | Rubric-lane precision per pattern | Primitive: one Noul per pattern (§5.1 lane 6). Labels: §6.8 protocol, double-labelled. Estimator, split, enrichment and non-hit sampling for recall, threshold tuning, interval, positive counts and the `unmeasured` rule for patterns that cannot reach them are fixed by its council-signed preregistration | Keep a pattern iff precision lower bound ≥ 0.6 and the preregistered recall minimum is met; never blocking. The 0.6 bound is a rev 2.1 value, pending the user's minimums; the preregistration confirms or replaces it |
| X5 | Jev commit classifier vs regex | Primitive: Noul `is_fix` (binary, tuned threshold), not a 5-class argmax. Labels: blind-labelled commits across 7 repos, §6.8 protocol (rev 2.1: 200). Commit state shape, split, sampling, interval, sample size and error reporting by path role and repository are fixed by its council-signed preregistration; E1b-v2's G4 (`fix→fix` 24 of 30, a derived Wilson interval) saw PR-level state and is a prior only | Enable `fix_touches`/change-success iff precision ≥ 0.85, recall ≥ 0.8 (LBs ≥ 0.75/0.7); rev 2.1 values, pending the user's minimums; the preregistration confirms or replaces them |
| X6 | Guardrail unwarranted-block rate (track-B dry run) | Replay last 200 merged PRs of 3 adopter repos, three languages; double-label `would_block` | Feeds AL dossiers and shadow defaults; activation stays per repo |
| X9 | Exploratory-measure attribution validity | Hand-check 200 `moved/pasted/new` labels and 200 `stale_reference` hits on the corpus | `exploratory` → `descriptive` iff label precision Wilson LB ≥ 0.9; else keep or drop |
| X10 | Competing-implementation detection | Primitive: one Noul over a pair ("same concern, different implementation?"), each pair judged in both presentation orders and averaged. Labels: 200 (rev 2.1), double-labelled, §6.8 protocol. Candidate-pair population, strata, split, threshold tuning, interval and sample size are fixed by its council-signed preregistration | Ship as exploratory iff precision ≥ 0.8; else research note (IR §5). Rev 2.1 value, pending the user's minimums (review §D q5, q6); the preregistration confirms or replaces it. No recall is measured, so the result can license `exploratory` only and never a prevalence claim |

X7 (pooled predictive re-test) and X8 (controlled continuations) were removed in revision 2.2; both are hypothesis studies that run in inertia-bench (§8.5). Their IDs are not reused.

---

## 14. Assumptions register

- A1 Area Jaccard 0.5 match / 0.8 survival (§4, X2).
- A2 Activation condition `unknown` ≤ 10 % of runs; retention demotion 3/20 (§7.3).
- A3 Protected branch lower friction than object storage; JSONL history squash-safe (§7.2).
- A4 Deterministic ranking captures much of Jev's apparent gain (§6, X1).
- A5 Rubric cost ≈ $0.12 per 200-function PR (unverified; to be measured as billed per function in Phase 5); precision unknown (§5.1, X4).
- A6 Decision budget 2,000/run; refresh debounce 1 h (§4, §6).
- A7 Window 90 days; `insufficient` floor 20 changes; trend deltas and three-window rule (§8.3–8.4).
- A9 TS type-only import handling and `tsconfig` path resolution suffice for cycle measurement (§5.1).
- A11 Cognitive-complexity adapters exist per tier-1 language (§8.2; not in launch set).
- A12 Naive SCB idiom patterns penalise Go/TS idioms (§5.1 rubric list).
- A8 and A10 were removed in revision 2.2 (pooled-test sizing belongs to inertia-bench; the Phase-1 performance bar is decided, Appendix 14); their numbers are not reused.

---

## 15. Decisions for the user (remaining, genuinely open)

1. **Study-export consent contract** (§8.6): confirm the export contents (scope-level aggregates + sanitized change table, pseudonymised authors, no messages/identities) and whether Scout's own dogfood ledger becomes the first pooled dataset. (Export as a separate opt-in with Scout-owned consent, and bench-run pooled analyses, are decided: Appendix 11.)
2. **PR-metadata ingestion** (§8.3 `pr_effort`, `batch_size`, `dispersion`): accept `pull-requests: read` in the reusable workflow for trusted runs, or keep interest measures git-only (I4 including author active hours, `batch_size` and I5's active-hours part always `unmeasured`).
3. **Inertia framing after E1b** (§8): for merged PRs of comparable size and task type, E1b found no material association between complexity burden or its change and author effort (≥ 0.10 log-hours per SD) or predictability (≥ 0.10 relative dispersion per SD); it is associational, and other quality measures, defects and agent effects are untested (§8.5). This revision keeps code quality as the primary family and frames principal series as *structure to watch* rather than cost, provisionally. Confirm that framing, or choose to rebalance Scout's priority toward families whose findings are concrete defects (guardrails, security, architecture policy), with inertia series as a secondary panel.

All other open items from v1 §15 are decided (Appendix) and applied above. Revision 2.1's items 2 (X8) and 3 (performance target) left the list in revision 2.2: the first moved to inertia-bench, the second is decided (Appendix 11, 14).

---

## Appendix — Decisions taken (authoritative; items 1–7 v1 §15, 2026-10-19; items 8–15 user decisions of 2026-10-22/23, `docs/STATUS.md` §6, R-036)

1. Map refresh: both mappers at onboarding, then on default-branch merges that change monitored areas (new unowned files, new/removed manifests or entrypoints, structural changes to owned paths), debounced into one refresh; never on PR runs; areas layer stays.
2. Baseline: protected `scout/state` git branch; S3 deferred to an optional adapter after the branch backend is proven.
3. Blocking allow-list: `docs/scout-allowlist.md` adopted (eligible-not-active; two-track promotion); shadow window ≥ 14 days **and** ≥ 5 PR runs with comparable base, both configurable; overrides require a maintainer; rare security checks may activate `provisional: true` on track A alone (conservative defaults, to be confirmed by use).
4. Jev: on from launch, route/rank only; never changes issue state or the gate.
5. Language tiers: tier 1 = Go, Python, TypeScript.
6. Spec cutover: archive `SCOUT_REBUILD_REQUIREMENTS.md` in favour of the §10.3 documents.
7. Clawpatch: v0.7.3, commit `734f81bc1514b52a0627edb0c24ebe1a18b36105`, ported patch `scout-review-sanitize-v2` (CP07 §7); 0.8.x is a later separate review.
8. Model use: reasoning-LLM lanes only through `claude -p` / `codex exec` (Anthropic/OpenAI models, subscription login, no API keys in CI); `skipped (no model CLI)` where no CLI is logged in; Jev stays on OpenRouter (Infisical locally, GitHub secret on trusted CI runs, none on fork PRs).
9. E1b-v2 is grandfathered (`evaluation/change-cost-v2/RESULTS.md:5`); its preregistration, amendment and scripts are untouched.
10. Subject interface in-repo, two-step: principles adopted (§8.7); the exact schema is a separate council round after inertia-bench's design signs off; implementation after Phase 3 `measures` (Phase 3b).
11. Research moves to inertia-bench: §8.5's further tests, X7, X8, §15.2 and A8 are removed; the claim-upgrade rule stays and is satisfiable by a signed-off bench study; `scout export --study` stays a product feature with Scout-owned consent; Scout keeps product experiments X1–X6, X9, X10.
12. Spec cutover at Phase 6 as designed (Appendix 6 unchanged); `SCOUT_REBUILD_REQUIREMENTS.md` stays the live product spec until then.
13. `AGENTS.md` updated 2026-10-23: model-use rule; experiment rules scoped to product experiments; hypothesis ledger frozen for H\*/HA\*/HQ\*; Jev labelling protocol.
14. Phase-1 performance bar (former §15.3): `scout assess` on fastapi `d666ccb`, heuristic map, egress off, < 10 min on the development Mac.
15. Jev labelling: agents do a blind first pass (two independent sessions with no shared context and no tools, `claude -p` and `codex exec`, a frozen guide, an adjudicator) and the user reviews every case; labels are `agent_reviewed` before that review and `human` after it (`docs/STATUS.md` §6.8; §6).
