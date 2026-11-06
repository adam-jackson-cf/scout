# Design review — E1b v2 revision 2.1 — round 1

Scope: `docs/design-reviews/2026-10-21-e1b-v2/scout-design.proposed.md` (sha256 af525b98…) vs `scout-design.base.md` (901b97cb…), `proposed.diff`; evidence `evaluation/change-cost-v2/{RESULTS.md,PREREGISTRATION.md,AMENDMENT-1.md,results.json,sensitivity.json,validation-gates.json,reviews/round-10.md,round-14.md}`, ledgers. Read-only; `.codegraph` untouched; no secrets; no API calls. Independent challenger (`history://DesignE1bV2R1.E1bV2DesignChallenger`, 12 findings) ran blind to this analysis; synthesis below.

## 1. Verdict: ISSUES FOUND (8 major, 4 minor). Not signed off.

**Recommendation:** return revision 2.1 for a round-2 edit. The evidential core is sound — every headline number checked against `results.json`/`RESULTS.md` matches (T2 +0.023, 90 % CI [+0.004, +0.041], Holm 0.165; T4 +0.006 [−0.010, +0.023]; T1 −0.011; T3 +0.018; P1 −0.15 % [−0.69, +0.38], 0.333; 7,717 PRs = Σ of the 12 gate-passing repos in `results.json .gates.G2`; 480 areas; MDE80 0.042–0.063; G5 recall, G7 κ/precision; per-repo T2), E1b-v1 is cited nowhere as evidence, sign-offs exist (`reviews/round-10.md:41-44` 7d87913c…, `round-14.md:34-37` 527bd820…), and §15 item 5 is correctly a user decision. The defects are in transplanted *definitions* (not implementable or not E1b's), one acceptance criterion that cannot be executed, residual over-reach/under-reaction wording, and small numeric inexactness.

## 2. Issues

1. **Major — `burden_density` is not E1b's `Q`, so Phase-3 acceptance cannot hold.** `proposed.md:201` defines a scope-level density over the §8.2 population (`:189`: E1's `inertia-v1/exclusions.json` rules); `:328` requires per-PR equality with `panel-final.csv.gz` (`active_hours`, `Q`). E1b's `Q` is per PR over *modified/deleted eligible files existing at snapshot S (parent of first commit, else `baseRefOid`) with ≥ 10 NLOC* under `common.eligible` (`PREREGISTRATION.md` §2, §3.1; `common.py:43-54`), which additionally excludes `scripts|tools|bin|migrations|mocks|dist|build|spec` dirs, `test_*.py`, `conftest.py`, `setup.py`, `.stories.*` — none in `exclusions.json`. "Within rounding" names no tolerance, join key or row-coverage rule. **Fix:** add a per-PR `scout.pr_burden_density@1` (or rename) with E1b's population, eligibility (`common.eligible` verbatim, versioned) and snapshot rule, citing `pins.json`; acceptance: join on (repo, PR number), ≥ 99 % of panel rows matched, |ΔQ| ≤ 1e-6, |Δhours| ≤ 0.01.

2. **Major — Author active hours "from git" is not implementable for squash merges, and the PR rule contradicts the change rule.** `:216` ("git (active hours)… also locally"), `:328`, `:402` promise local active hours for "PRs whose merge/squash commit is a change in scope"; squash-merged PRs' original author timestamps are not in the default-branch history — E1b fetched PR commits via GraphQL (`PREREGISTRATION.md` §2 `fetch_prs.py`, §3.3). `:207` defines Change = *non-merge*, so a merge commit can never be "a change in scope". **Fix:** specify sources (merge-commit second-parent ancestry locally; API PR-commit list otherwise; `unmeasured: pr_commits_unavailable`), and define PR inclusion separately from the non-merge change population; soften `:402`.

3. **Major — Bot rule has no git-side identity source.** `:108`, `:207` adopt "GitHub `Bot` type or exact login list"; `Bot` type and logins exist only in the API (`common.py:80-86`); git exposes author name/e-mail. **Fix:** define the git predicate (exact author e-mail/name set in `.scout/policy.yaml`, seeded from `common.BOT_LOGINS`; API `Bot` type when present; no patterns), and record which predicate applied per change.

4. **Major — §4 area rule is not E1b's and re-keys scopes.** `:92`: per *file*, "below the source root", share of "changes", deepened dynamically. E1b (`PREREGISTRATION.md` §2): per *PR*, directory of the largest-churn measured file, 2 path components, deepened on *panel-PR* share, independent of ownership. Dynamic deepening changes scope keys as traffic shifts, contradicting same-population deltas (`:203`), "never re-keys" (`:92`) and R6 (`:353`); R4 (`:351`) claims the rule is "shared". **Fix:** store E1b's PR-area as a per-PR attribute (needed for T2/T4 FE) separate from measurement scopes; freeze the directory partition per map version with lineage; correct R4.

5. **Major — §8.6 change table cannot support the promised T1–T4.** `:261` promises T1–T4 on Scout's PR table; `:257` stores churn, files, timestamps, workflow type, bot flag, optional task category — no per-PR `Q`, `D`, `Q_stable` at S, prior-author-commit count, maintainer status, PR-area (`PREREGISTRATION.md` §3.1–3.2, §4). **Fix:** enumerate the required per-PR predictors/controls with missingness reasons, or make `:261` conditional on their presence.

6. **Major — Conclusion stated beyond the estimand (over-reach).** `:19`, `:185`, `:229`, `:247`, `:403` say E1b "excluded a material PR-effort penalty" without the preregistered scope: within-area association *at comparable size and task*, so amplification-mediated cost is outside T1–T4 (`PREREGISTRATION.md` §6); floored proxy where location and dispersion are inseparable (§3.3); observational. §8.5's "does not test" list omits amplification — Scout's own I2. **Fix:** add "(7) cost acting through larger or more dispersed changes (I2)" and the floor note to `:247`; qualify `:19`/`:229`/`:403` with "for changes of comparable size and task".

7. **Major — Residual cost framing (under-reaction) and §15 item 5 cross-reference.** `:11` asserts "Code quality as *the driver* of long-term codebase inertia" as fact; `:229` "structural debt in frequently changed code" implies cost one clause before "not a cost claim"; `:39` justifies structural-first without noting it is now pending item 5; `:249` says "adopted" while `:403` asks confirmation. Item 5 itself is correctly a user decision (family priority is R-034 direction, not an Appendix item). **Fix:** `:11` "hypothesised driver (H1–H4; §8.5; §15 item 5)"; `:229` "adverse structural trend in frequently changed code"; `:39`/`:249` "adopted provisionally, priority pending §15 item 5". Keeping the principal/interest vocabulary is acceptable given `:185`.

8. **Major — A8 and §8.5(a) round away test-specific numbers and mix units.** `:389`, `:251`: "≈ 7,000 PRs, ≈ 450 clusters gave MDE80 ≈ 0.05" conflates panel (7,717/480), T1 (7,028/451) and T2 (6,110/417; MDE80 **0.063**, `RESULTS.md` §1, §3); `:235` "SESOI 0.10 log-hours per SD" applies to T1/T2 only (T3/T4: relative dispersion, prereg §4). X7's rule is MSE-reduction (`:253`), for which a slope MDE is not a power basis. **Fix:** cite per-test n, clusters and MDE80 exactly; A8 as historical scale reference, X7 sample by its own power simulation (already required at `:373`).

9. **Minor — Numeric inexactness.** `:245` "below 0.06 at the 95 % upper bound" is false: cohort < 2024 T2 UB = 0.0603 (`sensitivity.json d_cohort_merged_before_2024.T2.ci95[1]`; `RESULTS.md` §5 "≤ 0.060"). `:243` G5 0.42 / G7 κ 0.51 vs reported 0.422 / 0.514. `:242` P1 interval unlabelled (it is `ci95`); `:241` T1/T3 lack intervals. **Fix:** "≤ 0.060"; exact gate values; label CI levels; add T1/T3 90 % CIs [−0.029, +0.007], [−0.003, +0.038].

10. **Minor — "E1b: a backlog-dominated proxy" (`:216`) is not an E1b-v2 result.** It is a v1-critique premise carried into the v2 design (`PREREGISTRATION.md` §3.3 "fixes v1's backlog-latency problem"; `DESIGN-NOTES.md` design-review item 2, cycle time rejected; R-033). **Fix:** cite "E1b prereg §3.3 rationale / R-033", not E1b as evidence.

11. **Minor — Pre-rule label (`:5`) is narrower than rule 3.** It exempts E1/E2/E3/CP07 only "for or against H-hypotheses"; `AGENTS.md` design rule 3 covers any experiment-driven design change. The citations are acceptable only because each supports a v2 decision of 2026-10-19 that predates the rules (ledger created 2026-10-20, `hypothesis-ledger.md:35`) and no revision-2.1 change relies on them (checked: `:29`, `:33`, `:107`, `:148`, `:189`, `:194`, `:223`). **Fix:** say exactly that, with the dates, and that altering any such decision needs rule-compliant evidence.

12. **Minor (pre-existing, not E1b-driven) — X8 rule tests HA2, not magnification.** `:374` "lower success or higher cost across ≥ 2 agents" has no human/reference comparator or interaction (HA1/HA3, `hypothesis-ledger.md:21-23`). **Fix:** name HA2 or add the comparator; may be deferred to its own review if the user prefers a narrow 2.1.

## 3. Challenger synthesis

Agreed and adopted: its findings on estimand scope (6), units/A8 (8), numeric exactness (9), framing (7), git-only effort and bot identity (2, 3), Q population and tolerance (1), area rule (4), §8.6 table (5), X8 (12). **Disagreement resolved by evidence:** its "blocker" on pre-rule citations — rule 3 governs *design changes*; every E1/E2/E3/CP07 use supports an unchanged v2 decision predating the rules, so it is downgraded to the wording fix in issue 11. Its claim that item 5 is "already decided" is partly rejected: the framing must change to stay consistent with E1b; only the cross-references (`:39`, `:249`) need the "pending" marker. The challenger changed my recommendation on issues 2 (merge-vs-non-merge contradiction) and 5 (missing controls), which I had not found.

## 4. Alternatives

- **Sign off with issues 9–12 only** — rejected: rule 1 requires zero issues of any severity; issues 1–5 leave acceptance criteria unexecutable.
- **Drop E1b comparability (remove `:328` equality, `:261`, R4's "shared" claim)** — smaller, consistent revision; loses the pooled-study pathway §8.6 depends on. Viable if the user prefers a minimal 2.1.
- **Full comparability (recommended)** — implement issues 1–5 as specified; larger edit but the only option that makes R4/X7 true.

## 5. Risks

- Issues 2–3 may force `pr_effort` active hours to depend on `pull-requests: read` for squash repos, reopening §15 item 4's premise ("active hours still measured from git").
- Tolerances in issue 1 are proposals; lizard version drift vs `requirements.txt` (1.24.1) could break equality — pin it in the acceptance.
- All fixes are documentation edits; fully reversible.

## 6. Execution outline (round 2)

1. Edit `:201`/`:189`/`:328` per issue 1; cite `common.eligible`, snapshot rule, `pins.json`, tolerance.
2. Edit `:207`/`:216`/`:402` per issue 2; `:108`/`:207` per issue 3.
3. Edit `:92`/`:351` per issue 4; `:257`/`:261` per issue 5.
4. Wording: `:11`, `:19`, `:39`, `:229`, `:247`, `:249`, `:403` (issues 6–7).
5. Numbers: `:235`, `:241-245`, `:251`, `:389` (issues 8–9); `:216` citation (10); `:5` note (11); `:374` (12).
6. Regenerate `proposed.diff`, record new SHA-256, request round 2; re-check every number against `results.json`/`sensitivity.json`.
