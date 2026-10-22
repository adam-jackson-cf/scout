# Design review — E1b v2 revision 2.1 — round 2

Scope: `docs/design-reviews/2026-10-21-e1b-v2/scout-design.proposed.md` (sha256 `040e39f9…` per `round-1-response.md`; not independently hashed — read-only, no shell) vs `scout-design.base.md` (`901b97cb…`), `proposed.diff`, `round-1.md`, `round-1-response.md`; evidence `evaluation/change-cost-v2/{PREREGISTRATION.md, RESULTS.md, results.json, sensitivity.json, validation-gates.json, power-results.json, common.py, build_panel.py, fetch_prs.py}`, `docs/{hypothesis,research}-ledger.md`, `AGENTS.md`. Read-only; `.codegraph` untouched; no secrets; no API calls. Independent challenger `history://DesignE1bV2R2.E1bV2R2Challenger` ran blind to this analysis; synthesis in §4.

## 1. Verdict: ISSUES FOUND (1 major, 7 minor). Not signed off.

The evidential core is now correct. Every number re-checked against `results.json`/`sensitivity.json`/`validation-gates.json`/`power-results.json` matches: T2 +0.023 (≈ +2.3 %), 90 % CI [+0.004, +0.041], Holm 0.165; T4 +0.006 [−0.010, +0.023]; T1 −0.011; T3 +0.018; P1 −0.15 % [−0.69, +0.38] (ci95), 0.333; 7,717 PRs / 12 repos / 480 areas; MDE80 T1 0.0449, T2 0.0629, T3 0.0419, T4 0.0528 (range 0.042–0.063; A8's 0.045/0.063 with n 7,028/451 and 6,110/417); G5 0.422, G7 κ 0.514 / 0.630; sensitivity max UB 0.0603 (`d_cohort_merged_before_2024.T2`), all other listed variants' T2/T4 UBs ≤ 0.0539; per-repo pandas +0.096 [+0.048, +0.144], vite +0.061, tokio −0.114; G3b Spearman 0.460 (`RESULTS.md:30`); "council decision b" exists (`PREREGISTRATION.md:193`); area rule 2 components / > 15 % (`PREREGISTRATION.md:52-53`, `build_panel.py:20`); `burden_density` formula equals prereg `Q` (`PREREGISTRATION.md:59-60`); active-hours rule matches `build_panel.py:18-19,56-62`; panel join keys `repo`,`number` exist (`build_panel.py:125`). Rule-3 compliance holds: every result-driven change cites E1b v2 (R-035, signed `reviews/round-10.md`, ledger history 2026-10-21); `:5` correctly scopes E1/E2/E3/CP07 to pre-rule decisions (ledger `:35`). Editing `RESULTS.md:70` is permitted — experiment rule 5 freezes only the preregistration and scripts. §15 item 5 remains a user decision and the Appendix does not pre-empt it.

What remains is residual unresolved round-1 text plus small implementability/consistency gaps introduced by the round-1 fixes.

## 2. Round-1 resolution check

| # | Status | Evidence |
|---|---|---|
| 1 | Resolved | `:201` scope-level only, "not interchangeable"; `:328` Q-equality removed, active-hours criterion with join, coverage and 1e-9 tolerance |
| 2 | Resolved | `:207` commit vs PR populations separated; `:216`, `:328`, `:402` consistent on API-sourced PR commits, `unmeasured` locally |
| 3 | **Partly** | Git predicate given (`:207`, `:108`); but "which predicate applied" is not recorded (`:257` stores only `bot flag`), and the no-reply seed is not derivable from `common.BOT_LOGINS` (`common.py:81-82` holds bare logins, no numeric IDs; `web-flow` is not a `[bot]` account) → issue 4 |
| 4 | Resolved | `:92` partition fixed per map version, E1b rule not used for scopes; `:351` no longer claims sharing |
| 5 | Resolved | `:261` re-estimation withdrawn, missing controls named |
| 6 | Resolved as specified | `:19`, `:247`, `:403` carry size/task caveat; amplification and floor added at `:247` (see issue 8 for a residual) |
| 7 | **Partly** | `:11`, `:229`, `:249` fixed; `:39` still "Not followed … structural series is first-class" with no §15 item 5 marker → issue 2 |
| 8 | **Partly** | `:389` fixed; `:233`, `:247`, `:403` still apply log-hours to T3/T4 → issue 1 |
| 9 | **Partly** | `:242-245` fixed; `:241` T1/T3 still without intervals → issue 3 |
| 10 | Resolved | `:216` cites prereg §3.3 construct, not a v2 result |
| 11 | Resolved | `:5` |
| 12 | Resolved | `:374`, `:251`, `:400` all HA2-only |

## 3. Issues

1. **Major — SESOI and MDE units wrong for T3/T4 in three places, including the user's decision text.** `:233` "SESOI 0.10 log-hours per SD … empirical MDE80 0.042–0.063 log-hours per SD for T1–T4"; `:247` and `:403` "material (≥ 0.10 log-h per SD) author-effort **or predictability** penalty". Prereg fixes T3/T4's SESOI as `0.10` *relative dispersion per SD* (`PREREGISTRATION.md:128`) and T4 is a relative-|residual| slope (`:191`); only T1/T2 are log-hours. Unresolved from round-1 issue 8. **Fix:** `:233` "SESOI 0.10 per SD (log-hours for T1/T2; relative dispersion for T3/T4); MDE80 0.045/0.063 log-h (T1/T2), 0.042/0.053 relative dispersion (T3/T4)"; `:247`/`:403` "≥ 0.10 log-h per SD on effort or ≥ 0.10 relative dispersion per SD on predictability".

2. **Minor — `:39` contradicts §15 item 5.** It rejects the outcome-first framing as settled while `:403` offers exactly that rebalance ("inertia series as a secondary panel") as open, and `:249` is "provisional pending §15 item 5". Required in round-1 issue 7. **Fix:** append "Primary-family priority is provisional pending §15 item 5."

3. **Minor — T1/T3 reported without uncertainty (`:241`).** `results.json .estimates.T1.ci90` = [−0.029, +0.007], `T3.ci90` = [−0.003, +0.038]. Required in round-1 issue 9. **Fix:** add both 90 % CIs.

4. **Minor — Bot seed not derivable; predicate provenance not stored.** `:207` seeds `history.bots` "as `<login>[bot]` names and their GitHub no-reply addresses" from `common.BOT_LOGINS`; GitHub bot no-reply addresses carry the numeric account ID (`<id>+<login>[bot]@users.noreply.github.com`), absent from `common.py:81-82`; `web-flow` has no `[bot]` form. `:257` keeps a `bot flag` only, whereas round-1 issue 3 required recording which predicate matched. **Fix:** ship an explicit seed table (name, e-mail) in `.scout/policy.yaml` defaults, state that e-mails are listed literally (no derivation), drop or special-case `web-flow`; add `bot_reason ∈ {author_name, author_email, api_bot_type, api_login}` to the change table.

5. **Minor — PR→scope linkage underspecified (`:207`).** "merge, squash or rebased commits touch ≥ 1 eligible file in scope (linkage from the API)" names no API object: E1b used the PR `files` list and `mergeCommit` (`fetch_prs.py:27-31,95-104`); rebased PRs leave no single merge commit, so "rebased commits" needs `associatedPullRequests` or the PR file list. `:257` stores "linkage completeness" with no stated consequence. **Fix:** define scope membership by the PR's API changed-file list (as E1b), keep `mergeCommit`/`headRefOid` for timestamps only, and state that incomplete linkage (> 100 files/commits truncated, missing merge commit) makes the PR `unmeasured` for I4/I5/`batch_size`, counted.

6. **Minor — I5 missingness not propagated.** `:217` `dispersion` of "PR active hours" is "Phase 3" with no local/without-`pull-requests: read` condition, while `:216` makes its input `unmeasured` there; `:358` (R11) and `:402` (item 4) list only I4 and `batch_size`. **Fix:** `:217` "active-hours component `unmeasured` whenever I4 is; churn component unaffected"; add I5 to `:358` and `:402`.

7. **Minor — Floor share lacks its denominator (`:247`).** "38 % of rows" follows `:233`'s 7,717 PRs; the measured value is 38.2 % of 7,030 PRs with a non-missing `Y_active` (`RESULTS.md:29`; `results.json .gates.G3.floor_share`). **Fix:** "38.2 % of the 7,030 PRs with a measured active time".

8. **Minor — Associational nature omitted from the limits paragraph (`:247`, `:403`).** "do not carry … penalty" / "excluded a … penalty" reads as a causal exclusion; the prereg states T2/T4 "do not identify causation" (`PREREGISTRATION.md:180-181`) and RESULTS lists "Association, not causation" as a threat (`RESULTS.md:81`). The §8.5 limits list names size control and the floor but not this. **Fix:** add a third limit — "E1b is observational: area fixed effects remove time-invariant area difficulty, time-varying confounding remains (prereg §6)" — and use "no material within-area association" at `:403`.

## 4. Challenger synthesis

Agreed and adopted: unresolved residue of issues 3, 7, 8, 9 (my issues 1–4); new findings on linkage (5), I5 missingness (6), floor denominator (7). Its "major" on observational framing is accepted in substance but downgraded to minor (8): round-1 issue 6's required fix (size/task caveat, amplification, floor) was implemented; the causal qualifier was rationale, not a required edit, so this is a new wording gap, not a defect in compliance. Its linkage finding changed my recommendation: I had read `:207` as implementable via `associatedPullRequests`; the challenger's point that E1b's own collector relies on the PR file list makes the simpler rule the right one for comparability. Numeric checks agree in full.

## 5. Alternatives

- **Sign off now** — rejected: design rule 1 requires zero issues; issues 1–3 are explicitly unresolved round-1 items.
- **Narrow 2.1 (drop PR-level measures to a later revision)** — removes issues 4–6 but reopens §15 item 4 wording; not recommended, the fixes are one-line edits.
- **Fix all eight (recommended)** — wording and definition edits only, no evidence change.

## 6. Risks

- Issue 1 touches the user's §15 item 5 text; the decision must be presented with correct units before it is taken.
- Issue 4's e-mail seed may need a hard-coded ID table; if omitted, name-only matching is acceptable if stated.
- All fixes are documentation edits; fully reversible.

## 7. Execution outline (round 3)

1. `:233`, `:247`, `:403` units (issue 1); `:39` marker (2); `:241` CIs (3).
2. `:207`, `:257` bot seed and `bot_reason` (4); `:207`, `:257` linkage rule and missingness (5).
3. `:217`, `:358`, `:402` I5 (6); `:247` denominator (7); `:247`, `:403` associational limit (8).
4. Regenerate `proposed.diff`, record the new SHA-256, request round 3.
