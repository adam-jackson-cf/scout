# Design review E1b v2 — Response to round 1

All 12 issues are fixed in `scout-design.proposed.md` (now `040e39f9…`; base and `docs/scout-design.md` are unchanged at `901b97cb…`). I chose the council's recommended direction: keep comparability with E1b only where the shared definition is exact and checkable, and drop every comparability claim that would need E1b's per-PR populations or controls.

| # | Fix (proposed file line) |
|---|---|
| 1 | `burden_density` is now scope-level only. `:201` states that it is the same ratio as E1b's `Q`, but computed on the scope population under §8.2 eligibility, so its values are not interchangeable with E1b's. The Phase-3 criterion against `Q` is removed (`:328`). |
| 2 | **Data sources.** Active hours now come from the PR's own commit list via the API; local runs report `unmeasured`, because squash merges drop the PR's commits from history (`:216`). §15 item 4 is restored: without `pull-requests: read`, all of I4 is `unmeasured` (`:402`). **Contradiction resolved.** Commit-level and PR-level inclusion rules are now separate (`:207`): commits are non-merge; PRs are linked through their merge, squash or rebased commits. |
| 3 | **Commit-level bots:** matched against an exact author name or e-mail list (`history.bots`), seeded from `common.BOT_LOGINS` as `<login>[bot]` names plus their no-reply addresses, with no substring matching. **PR-level bots:** API `Bot` type or the same login list (`:207`, `:108`). |
| 4 | E1b's area rule is no longer used for scopes. Scopes are directory scopes at 2 components, fixed per map version, and `:92` explains why the study's area rule is not reused. R4 no longer mentions it (`:351`). |
| 5 | The per-repository E1b re-estimation promise is withdrawn: the change table lacks the predictors and controls it needs, so adding them is a separate change (`:261`). |
| 6 | Scope caveats are added. Each claim now reads "for merged PRs of comparable size and task type" (`:19`, `:247`, `:403`). Two further caveats are stated: amplification is outside the estimand (size is controlled, prereg §6), and T2 and T4 cannot be separated because of the floor, so they are read together (`:247`). |
| 7 | `:11` now says "hypothesised driver". `:229` now says "adverse structural trend" and "not a cost claim (§8.5)". `:249` changes "adopted" to "provisional pending §15 item 5". |
| 8 | A8 (`:389`) and §8.5 (a) now give exact figures: T1 0.045 (7,028 PRs, 451 clusters) and T2 0.063 (6,110 PRs, 417 clusters). Both are labelled scale-only; a pooled test is sized by its own simulation, and X7's MSE rule has no power estimate yet. The "≈ 0.05" wording is removed. |
| 9 | The sensitivity bound is now "≤ 0.061 (largest 0.0603, T2, pre-2024)" (`:245`). G5 and G7 are reported exactly (0.422, 0.514, 0.630), and P1 is labelled 95 % CI (`:242`, `:243`). The same rounding error in `evaluation/change-cost-v2/RESULTS.md:70` is corrected; RESULTS.md is not a frozen artefact. |
| 10 | Queue time is now justified by E1b's effort construct, which excludes review and queue waiting by definition (prereg §3.3). It is no longer presented as a v2 result (`:216`). |
| 11 | `:5` now explains why the pre-rule studies stay: E1, E2, E3 and CP07 support v2 decisions taken before the rules were adopted (2026-10-20, ledger history), and no change in revision 2.1 rests on them. |
| 12 | X8 now tests HA2 only, and says it does not test HA1 or HA3 (`:374`). §8.5 (b) and §15 item 2 are aligned with that (`:251`, `:400`). |

## Verification

- **New Phase-3 criterion is executable.** `build_panel.active_hours` was run on the PR commit author timestamps in `evaluation/change-cost-v2/raw/prs-{nest,webpack}.json.gz` and reproduces `panel-final.csv.gz` `active_hours` for all 598 panel PRs of those two repositories, with 0 mismatches above 1e-9.
- **Word diff:** `proposed.diff`, regenerated.
