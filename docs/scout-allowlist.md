# Scout PR blocking allow-list at launch — council decision memo

Independent challenger analysis: `local://challenger-allowlist.md` (transcript `history://AllowListCouncil.AllowlistChallenger`). Council and challenger converged independently on the key arithmetic (§4). Challenger changed the recommendation in four places (marked **[CH]**); one disagreement is resolved in §1.

## 1. Recommendation

**Ship `guardian` with a short, exactly-named, mechanism-certified allow-list that is *eligible*, not *active*: every adopting repo starts in `report`, sees `would_block` annotations, and activates each check by a maintainer PR to base `.scout/policy.yaml`.** Replace the §7.4 promotion rule (it is unsatisfiable as written) with a two-track rule (§4).

Disagreement resolved: the challenger proposed an *empty* active list at launch. Agreed for "active everywhere"; rejected for "eligible" — a `guardian` profile with nothing to activate is dead code, and the per-repo shadow window (§3) already supplies the repo-specific evidence the challenger wants. **[CH]** Wildcard `guardrails.CONFIG.*` is dropped in favour of exact per-tool ids so new parsers never inherit blocking rights.

Prerequisites before *any* check activates (design §3.3 defects, verified by challenger): feature IDs out of issue hashes (`internal/issues/correlate.go:326-340`), line positions out of evidence identity (`internal/state/issue_observations.go:94-102`), dispositions surviving evidence drift (`internal/state/issues.go:455-461`). Without these, mapping churn (E3: 5–31 % agent-ID survival) manufactures `introduced` findings and false blocks.

**Escape hatch (all rows, keeps discovery-only):** the ecosystem's native *justified* form is the escape — `//nolint:rule // reason`, `# noqa: CODE  # reason`, `# type: ignore[code]  # reason`, `// eslint-disable-next-line rule -- reason`, `// @ts-expect-error reason`, `// biome-ignore lint/x: reason`. Where no native form exists (secrets, configs, workflows) a `scout:allow <id> reason=… ref=…` comment or PR label `scout-override` records a `disposition: accepted{actor, pr, reason}` in the ledger. The finding is never deleted; overrides feed the retention metric (§4). Note: Scout's Gitleaks invocation passes `--ignore-gitleaks-allow --gitleaks-ignore-path /dev/null` (`internal/analyzers/catalog_security.go:266-267`), so the secrets escape **must** be Scout-side.

### Eligible blocking checks (default profile `guardian`, per-repo activation)

| id | Mechanism | Go | Python | TypeScript | Expected precision / why | FP classes → attribution/exemption |
|---|---|---|---|---|---|---|
| `security.ci-mutable-action-ref` | New remote `uses: owner/repo[/path]@ref` where ref ≠ 40-hex SHA (`rules/security/semgrep-actions-mutable-ref.yaml:10-15`); `.github/workflows/*.y*ml` only | lang-agnostic | — | — | Highest: 6/6 adjudicated real (E2 line 55); regex is a pure format test | Version bumps (`@v4`→`@v5`) must be `pre-existing`: identity = (workflow, owner/repo), not ref. Local `./` and `docker://` already don't match. Policy `trusted_owners: [actions, github]` → report-only **[ASSUMPTION]** |
| `security.secrets.<rule>` (exact ids only: `aws-access-token`, `github-pat`, `github-fine-grained-pat`, `github-app-token`, `private-key`, `slack-*`, `stripe-*`, `gcp-api-key`…) | Gitleaks 8.30.1 pinned (`gitleaks_rule_ids.go`), vendor-prefixed/checksummed formats | lang-agnostic | — | — | High for structured formats; **`generic-api-key` excluded** (E2 line 52: 27 FP/3 real by Jev, unverified) | Fixture/revoked keys: not path-exempted (a real key in `testdata/` is still a leak) → Scout-side escape. Redact evidence (`--redact=100`). Moves across files: match by secret digest |
| `guardrails.CONFIG.ignores-added.<tool>` | Diff base/head of parsed config; block when an ignore set *grows* for a tool enabled in base | `.golangci.y*ml`: `linters.disable` +, `issues.exclude-rules`/`exclude-dirs`/`exclude-files` +; `staticcheck.conf` `checks` gains `-SAxxxx` | `pyproject [tool.ruff(.lint)] ignore/extend-ignore/per-file-ignores/exclude` +, `ruff.toml`, `.flake8`/`setup.cfg` `ignore` +; `[tool.mypy] ignore_errors=true`, `[[tool.mypy.overrides]]` + | `.eslintrc.{json,yaml,cjs}`/`eslint.config.*` rule → `off`/`0`, `ignorePatterns`/`ignores`/`.eslintignore` +; `biome.json` rule off; `tsconfig*.json` `exclude` + | High for "effective disablement"; unmeasured for harm | Tool-version migration (golangci v2 `version: 2`, ruff rule renames) → same PR changes version key ⇒ `unknown`, report. New config file (new package) ⇒ no base ⇒ report. JS-code configs that can't be parsed statically ⇒ `unknown` |
| `guardrails.CONFIG.strictness-lowered.<tool>` | Named boolean/enum transitions only, resolved through `extends`/inheritance | golangci `linters.enable` −, `enable-all`→false | mypy `strict`/`disallow_untyped_defs`/`check_untyped_defs` true→false; pyright `typeCheckingMode` strict→standard/basic; `[tool.coverage.report] fail_under` ↓ | `tsconfig` effective `strict`/`strictNullChecks`/`noImplicitAny`/`noUncheckedIndexedAccess` true→false; eslint severity `error`→`warn`; jest `coverageThreshold` ↓ | High for the transition; a textual `strict:false` in a non-inherited file is not enough — must be *effective* **[CH]** | Replacement validator in same PR (mypy→pyright) ⇒ escape. Inherited defaults ⇒ resolve chain or `unknown` |
| `guardrails.SUPPRESS.added.unscoped` | Only file-level or rule-less suppressions in non-test, non-generated code | bare `//nolint` (no `:rule`), `//nolint:all` | bare `# noqa`, bare `# type: ignore`, `# pylint: disable=all`, `@no_type_check`, `# mypy: ignore-errors` | `// @ts-nocheck`, bare `/* eslint-disable */`, bare `// @ts-ignore` | High: these have no legitimate narrow use; scoped forms (`# type: ignore[attr-defined]`, `eslint-disable-next-line x`) are everyday code and stay **report-only** | Moved/unchanged suppression ⇒ `pre-existing` (identity excludes line). Generated/vendored paths (`path role`, §6 item 6) ⇒ report. Scoped+justified = escape |
| `guardrails.TESTQ.focused-test-added` (council addition) | `.only`/`fit`/`fdescribe`/`test.only`/`describe.only` committed in a test file | — (no construct) | — (no construct) | jest/vitest/mocha/playwright; same mechanism as `jest/no-focused-tests` | ≈1.0: a committed `.only` silently disables every sibling test in CI | None known beyond scratch files; `scout:allow` if intentional |

Demoted from the seed **[CH]**: `guardrails.GATE.coverage-pragma-added` → report-only (table 2); `generic-api-key` → report-only; wildcard `CONFIG.*` → exact ids.

## 2. Report-only at launch

| Check | Reason |
|---|---|
| Inertia deltas, CC/NLOC, clones, rework, any score threshold | E1: +1.18 % MSE, CI [−0.62, +3.24], fails prereg; AUROC Δ −0.019 (`evaluation/inertia-v1/RESULTS.md:9-18`) |
| Rubric / Jev-ranked / contextual / clawpatch-review | Jev route-only (§6 item 2); E2 n=22, Caveat 1 leakage, Caveat 3 narrow labels (`jev-triage-v1/RESULTS.md:69-71`) |
| Import cycles | Policy-dependent; type-only imports (Py `TYPE_CHECKING`, TS `import type`) make naive cycles wrong **[ASSUMPTION]**; becomes `architecture-policy` block only after the repo declares cycle prohibition (§7.3) |
| Hollow tests `TESTQ.{constant,literal,no}-assert`, `delegating`, `skipped` | Helper-based assertions (Go `t.Helper()`, testify), snapshot and "does not throw" tests defeat syntax heuristics; `t.Skip("reason")` is legitimate platform gating |
| `GATE.coverage-pragma-added` (`# pragma: no cover`, `/* istanbul|c8|v8 ignore */`, coverage `exclude_lines` +) | Legit uses (`TYPE_CHECKING`, `__main__`, abstract methods); Go has no pragma. Promotable per repo only when a coverage gate (`fail_under`, `coverageThreshold`) exists *and* the pragma is unjustified **[ASSUMPTION]** |
| Coverage-of-changed-lines (`artifacts`) | Producer/revision match required; absent ⇒ `unmeasured` (§5.1 row 9); generated/log lines uncovered legitimately |
| `GATE.ci-step-neutered` (`continue-on-error: true` on existing job, `\|\| true` appended, `if: false`) | Strong candidate, but GitHub's documented `experimental` matrix pattern is a real FP class; no fixtures yet |
| `security.ci-pull_request_target-checkout`, `ci-script-injection`, `unsafe-yaml-load`, other `p/ci`/`p/<lang>` rules | High-value, but zero corpus positives (E2); PyYAML ≥6 `yaml.load` without Loader is a TypeError, not a hazard. Promote via track A fixtures + track B |
| Scoped suppressions, scoped `//nolint:rule`, `@ts-expect-error`, `.gitleaksignore` edits, test-file deletion, performance (Ruff PERF), tech/EOL | Everyday legitimate code or refactor noise; visibility suffices |

## 3. Start in `report`; graduation

Yes — `report` is already the adoption default (§7.3). Add to it: a `would_block` block in `scout-assessment.json` and the Markdown summary ("this PR would have been blocked by …"), persisted to `scout/state` as `signals/would_block.jsonl`.

Graduation per check per repo: window = min 14 days **and** ≥ 20 PR runs with comparable base **[ASSUMPTION]**; during it, maintainers triage every `would_block` event (`warranted` / `not warranted` / `undecided`). Activate when `not warranted` = 0 and no `unknown`-attribution storm (> 10 % of runs). Rare checks (secrets, action refs) with zero events at window end activate on track A alone, flagged `provisional: true`; first-baseline `pre-existing` counts in non-generated paths are the proxy for expected fire rate. Activation = maintainer PR editing base `.scout/policy.yaml`:
`blocking: [{id, version, activated_at, evidence: {would_block, warranted, window_days, dossier_digest}}]`. A parser/rule version bump returns the check to shadow for that repo **[CH]**.

## 4. Promotion rule verdict: **replace**

The rule "≥ 50 double-labelled, Wilson 95 % LB ≥ 0.95" is unsatisfiable at n = 50: with zero errors LB = 1/(1 + 1.96²/50) = **0.929**; zero errors need n ≥ 73; one error needs n ≈ 120. Both analyses derived this independently. Further defects: (a) rare checks never accrue 73 independent unseen PR findings — findings from one template are correlated, so double-labelling doesn't double n; (b) mechanism precision is trivially ≈1 for a regex, while the trust-relevant quantity is *unwarranted blocks / blocks*, only observable live; (c) corpus precision says nothing about an adopter's monorepo of generated code.

Replacement:
- **Track A — mechanism certification (global, Scout maintainers):** `selfeval` fixtures positive / clean negative / boundary / redaction **plus attribution fixtures** (line move, file rename, version bump, config migration, tool upgrade, new file) all pass; no known systematic negative class; versioned evidence dossier. Grants *eligibility*.
- **Track B — per-repo activation and retention:** shadow window (§3) grants activation. Retention metric = confirmed-unwarranted blocks over the trailing 20 blocks; ≥ 3 ⇒ auto-demote to report for that repo and record it **[ASSUMPTION]**. Overrides trigger review, not automatic FP labels **[CH]**. Also track gate-disable events (the real cost signal).
- Corpus-level double-labelled precision stays as an *audit/demotion* instrument with honest arithmetic (state the LB target and compute the required n; no fixed 50), never as a launch prerequisite.

## 5. Assumptions

Marked inline. Unmeasured: live precision of every row; developer tolerance for unscoped-suppression blocks; `trusted_owners` default; window sizes; demotion threshold. Prerequisite decisions for the user: exception authority (author vs maintainer for `scout-override`), and whether `provisional` activation on track A alone is acceptable for rare security checks. No working-tree or `.codegraph` changes; no tests run.
