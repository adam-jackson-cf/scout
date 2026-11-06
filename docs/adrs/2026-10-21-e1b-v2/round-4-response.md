# Design review E1b v2 — Response to round 4

All three round-4 issues are fixed. `scout-design.proposed.md` is now `7ff21a7f…`; this digest was computed with `shasum -a 256`. The base and `docs/scout-design.md` are unchanged (`901b97cb…`, also computed with `shasum`).

| # | Fix |
|---|---|
| 1 | Proposed `:207`: PR-level measures now count only "merged, **non-bot** PRs …". Bot PRs are counted and excluded, as in E1b's `collect.py:144-145`. |
| 2 | Proposed `:207`: "A PR over either cap, **or** whose file list or commit list the API returns incomplete …, is marked `linkage: incomplete`, counted and excluded." A fully paginated PR that is over a cap is therefore counted, as in E1b (`skip="bulk"` / `"many_commits"`). |
| 3 | In `docs/hypothesis-ledger.md:21`, the HA1 Tests field now reads "— (X8 tests HA2 only; HA1 needs a human comparison)". A history line records the change. |

`proposed.diff` has been regenerated.
