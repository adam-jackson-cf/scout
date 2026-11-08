# ADR-012 round 14 formatting addendum response

Date: 2026-11-08. Round 14’s signed bytes failed the commit hook’s formatting check on exactly two files:

- `docs/plans/2026-11-06-remaining-work.md`
- `docs/references/policy-file.md`

The fix was a rewrite of only those files by the repository’s pinned Prettier (`.tools/node-tools/node_modules/.bin/prettier --write`). Main verified that, after collapsing whitespace, table pipes and dash/colon rule rows, the output is identical to the previously signed bytes: whitespace and table padding only, with no content change. That comparison is main’s reported evidence, not an independent comparison by this writer.

The hook’s own `prettier(files, root)` function was run over all staged Markdown, applying its normal governed/frozen exclusions and pinned-version check. It passed: all matched files use Prettier code style; the function returned no errors. No commit was attempted after formatting, and no hook was bypassed.

No other signed file was modified. Historical review/response records remain unchanged. The two revised digests and all other signed-file digests accompany the handoff, together with this addendum’s digest and a saved worktree diff against the round-14 staged bytes. Changes are staged for narrow round-15 review; no commit is made and no renewed sign-off is claimed here.
