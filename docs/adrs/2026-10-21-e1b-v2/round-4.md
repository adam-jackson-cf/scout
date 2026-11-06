# Design review — E1b v2 revision 2.1 — round 4

Scope: `docs/design-reviews/2026-10-21-e1b-v2/scout-design.proposed.md` (sha256 `55810f1a…` as reported in `round-3-response.md`; not independently hashed — read-only, no shell) vs `scout-design.base.md` (`901b97cb…` = `docs/scout-design.md`), `proposed.diff`, rounds 1–3 and responses; evidence `evaluation/change-cost-v2/{PREREGISTRATION.md, RESULTS.md, results.json, sensitivity.json, validation-gates.json, power-results.json, common.py, collect.py, build_panel.py, fetch_prs.py}`, `docs/{hypothesis,research}-ledger.md`, `AGENTS.md`. Read-only; `.codegraph` untouched; no secrets; no API calls. Independent challenger `history://DesignE1bV2R4.E1bV2R4Challenger` ran blind; synthesis in §4.

## 1. Verdict: ISSUES FOUND (3 minor). Not signed off.

All three round-3 fixes are present and the numbers are exact (§5). What remains are two one-clause gaps in the new PR-population sentence at `:207` and one ledger cell that the revised X8 scope leaves contradictory. None changes a conclusion or a §15 decision.

## 2. Round-3 resolution check

| # | Status | Evidence |
|---|---|---|
| 1 `:108` bot rule | **Resolved** | `:108` "bot commits excluded by the `[bot]` author suffix or an exact `history.bots` entry, never substring — §8.3" = `:207` "author name ending in the literal suffix `[bot]` … or author name/e-mail exactly equal to an entry of `history.bots`" |
| 2 commit-list completeness | **Partly** | `:207` adds "≤ 100 commits" to PR eligibility and "a PR whose file list or commit list the API returns incomplete (fewer entries than its total count, or no commits) is marked `linkage: incomplete`, counted and excluded (as E1b `collect.py:146-149`)". Truncation/empty guard now matches `collect.py:148-149`; the cap matches `PREREGISTRATION.md:42` and `fetch_prs.py:29` (`commits(first:100)`). Residue: round 3 asked that over-cap PRs be "(counted, excluded)"; the text counts only API-truncated lists (issue 2) |
| 3 `bot_reason` for PRs | **Resolved** | `:207` "`app_suffix` or `configured` for commits, `api_bot_type` or `login_list` for PRs" ↔ `common.py:85-86` `is_bot`: `author_type == "Bot" or login.lower() in BOT_LOGINS`; `:257` "bot flag and `bot_reason` (§8.3)" |

## 3. Issues

1. **Minor — `:207` defines "bot PR" but never excludes it.** The commit-level rule reads "change = non-merge commit, **non-bot**, ≤ 50 files …"; the PR-level rule reads "merged PRs with ≤ 50 changed files and ≤ 100 commits whose file list … contains ≥ 1 eligible file in scope" — no non-bot condition — and then "bot PR = API author type `Bot` or login in … `common.BOT_LOGINS`" with no consequence attached. In the base (`base:200`) PR membership went through the merge commit being a *change*, which carried the non-bot filter; the revision links PRs by file list instead, so the filter was lost. E1b excludes bot PRs outright (`collect.py:144-145`, `PREREGISTRATION.md:41-42`), and a dependabot PR at the 0.5 h floor would otherwise enter I4/I5 and the G3-style floor share. **Fix:** `:207` "**PR-level measures** (…): merged, **non-bot** PRs with ≤ 50 changed files …" (keep the bot-PR definition and `bot_reason`; bot PRs stay in the change table flagged, as §8.6 implies).

2. **Minor — over-cap PRs are counted only if the API truncates.** `:207` claims "the same rule as E1b `collect.py:146-149`", but E1b counts `n_files > 50` and `n_commits > len(commits)` in one skip each (`skip="bulk"`, `skip="many_commits"`), i.e. every over-cap PR appears in the funnel. The design does not fix a page size, so an implementation that paginates all commits (or filters on `totalCount` before fetching) drops a 150-commit PR silently: outside the population, not `linkage: incomplete`, not counted — contradicting §8.1 "population and missingness stated" and the round-3 fix text "(counted, excluded)". **Fix:** `:207` "… a PR over either cap, or whose file list or commit list the API returns incomplete (…), is marked `linkage: incomplete`, counted and excluded (as E1b `collect.py:146-149`)".

3. **Minor — `docs/hypothesis-ledger.md:21` contradicts the revised X8 scope.** Proposed `:374` X8: "Tests HA2 … Does not test HA1 (needs a human comparison) or HA3"; `:251` (b) and `:400` item 2 agree. The ledger's HA1 row still lists Tests = "X8 (external, not commissioned)". Ledger rule 2 (`AGENTS.md`) requires the ledger to follow a hypothesis revision; round 3's "ledger IDs … with matching meaning" check missed this cell. The design text is right; the ledger is stale. **Fix (ledger, not `proposed.md`):** HA1 Tests → "— (X8 tests HA2 only; HA1 needs a human-comparison design)"; HA2 Tests keeps X8; add a 2026-10-21 History line citing revision 2.1. This must land with adoption; it does not alter `proposed.md`'s digest.

Nothing else found in the changed text or whole-document consistency: E1b-v1 appears only as "superseded and invalidated, R-033; not cited" (`:5`) and "E1b-v1's substring regex … is retired" (`:207`); `change-cost-v1`, "O1–O5", "60 areas" absent elsewhere; §1 non-goal 3, §2 `:39`, §4 `:92`, §8.1 `:185`, §8.2 `burden_density`, §8.4, §8.5 `:233-252`, §8.6 `:261`, Phase 3 `:328`, R4/R11, X7/X8, A8, §15 items 2/4/5 agree with one another and with `:403`; header `:5` correctly confines E1/E2/E3/CP07 to pre-rule engineering observations; R-033/R-034/R-035 (`research-ledger.md:41-43`) and H1/H2/H4 states (`hypothesis-ledger.md:14-17`, `:32-33`) match the cited results; E1b sign-off recorded (`hypothesis-ledger.md:33`, `reviews/round-10.md`); Phase 3 acceptance remains executable (`build_panel.active_hours`, `build_panel.py:56-62`: 0.5 h first credit, gap ≤ 4 h added, else +0.5 h — exactly `:216`'s wording). Experiment-rule compliance: every cited number is in `RESULTS.md`/`results.json` and inside the preregistered/amended analyses; per-repository and cohort figures are labelled exploratory as in `RESULTS.md:72-74`.

## 4. Challenger synthesis

My independent pass found the three round-3 fixes present and all numbers exact, and would have signed off. The challenger raised issues 1–3; I adopted all three after verifying each against the text and evidence (`:207` sentence structure vs `collect.py:144-149`; `hypothesis-ledger.md:21` vs `:374`). The challenger changed the recommendation from sign-off to ISSUES FOUND. Its alternative — separate file-linkage from commit-list completeness so `batch_size` survives a truncated commit list — is viable but widens the change; round 3's uniform `linkage: incomplete` stands. Numeric checks agree in full; the challenger likewise accepts "38 %" (0.3815) as correct rounding.

## 5. Evidence (numeric cross-checks)

`results.json .estimates`: T1 −0.0112 ci90 [−0.0290, +0.0066]; T2 +0.0228 [+0.0044, +0.0412], p 0.0412 (Holm 0.165, `RESULTS.md:13`); T3 +0.0175 [−0.0032, +0.0383]; T4 +0.0061 [−0.0104, +0.0226]; P1_rel_gain −0.00153 ci95 [−0.0069, +0.0038]; P1_base_vs_const 0.333; `.n` n_T1 7,028 (451 clusters), n_T2 6,110 (417); `.G` 480; `.gates.G3` n 7,030, floor 0.3815; `.G3b` ρ 0.4597; `.G2` 12 passing repositories sum to 7,717 (= `power-results.json .n`). `power-results.json`: empirical MDE80 T1 0.0449, T2 0.0629, T3 0.0419, T4 0.0528, SESOI 0.1, all `adequate`. `sensitivity.json`: largest pooled T2/T4 95 % UB `d_cohort_merged_before_2024.T2` 0.06025; `d_per_repo` pandas +0.0960 [+0.0479, +0.1442], vite +0.0611, tokio −0.1143. `validation-gates.json`: G5 `jev_recall_ipw` 0.422, G7 `kappa_pairs` 0.514, `hidden_precision` 0.63. All match `:233-243`, `:389`.

## 6. Risks

- Issue 1 removes bot PRs from I4/I5 populations — a strict narrowing; `bot_reason` keeps provenance. Issue 2 adds counting only. Issue 3 is a ledger edit. All reversible; no §15 decision text changes.
- Digest `55810f1a…` is taken from the author's response; the main agent should confirm with `shasum -a 256` before the next round's sign-off statement.

## 7. Execution outline (round 5)

1. `:207` insert "non-bot" in the PR-population sentence (issue 1).
2. `:207` extend the counted-and-excluded clause to over-cap PRs (issue 2).
3. `docs/hypothesis-ledger.md:21` HA1 Tests cell + History line (issue 3).
4. Regenerate `proposed.diff`, record the new SHA-256 of `scout-design.proposed.md`, request round 5 (sign-off expected if nothing else changes).
