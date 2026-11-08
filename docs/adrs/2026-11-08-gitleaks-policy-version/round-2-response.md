# Round 2 response: ADR-011

Date: 2026-11-08. Writer response to ADR011-R2-01…05 only. This is not a council verdict or sign-off. ADR-012 is a separate PR and review unit.

## Route change and scope

Main explicitly superseded the patch mechanism: separate canonical PRs, ADR-011 first, ADR-012 rebases after ADR-011 merges. This revision is in `/Users/adamjackson/Projects/work/scout-adr11b`, branch `agent/gitleaks-version-adr`, based on `origin/main` `d2479d54`. The canonical documents contain ADR-011 changes only, plus source prerequisites R-064 and static R-069. No ADOPTION.patch or ADOPTION-SHA256.txt is included. There is no exact-parent or no-squash restriction: each PR exposes its own reviewed canonical bytes. Canonical changes after review require renewed council review, regardless of merge method.

The signed set is ADR.md in this folder and the six canonical files: `docs/scout-design.html`, `docs/adrs/README.md`, `docs/STATUS.md`, `docs/plans/2026-11-06-remaining-work.md`, `docs/references/policy-file.md`, `docs/research-ledger.md`. Route (a) takes effect on this folder's explicit final zero-issue council round, which names signed SHA-256 digests; adoption is recorded outside signed files. Accepted-form status is not a claim of sign-off.

`round-1.md`, `round-2.md` and `round-1-response.md` were copied byte-identically from the source worktree. Historical patch/hash, base and combined-review references are intentionally preserved as history, not current instructions.

## Issue mapping

| Issue | Severity | Fix / route selected | Location |
|---|---|---|---|
| ADR011-R2-01 | major | Replace patch/hash split with separate canonical PRs under Main's superseding direction. ADR-011 lands first; ADR-012 rebases after it merges and is reviewed independently. No patch artifacts, fixed-parent requirement or merge-method restriction. R-064 lands here first; dependent work reuses it. Preserve existing source allocations; ledger introduction explains append order/pre-allocation. | ADR.md Options/Changes/Review; index procedure paragraph; STATUS/plan; ledger introduction |
| ADR011-R2-02 | minor | R-064 contains verbatim direction and the later-review/new-user-row caveat, not frozen owner-pending state. R-069 records only existing code facts and limitations, with no ADR-012 selections. No main ledger row is edited and no new ID is allocated. Owner-pending state remains only in editable README/STATUS/plan trackers. | R-064/R-069; ADR.md Review; trackers |
| ADR011-R2-03 | minor | Auto-demotion explicitly names “that repo's check … for that repo”. Retain original maintainer triage, window end, repository scope, counting unit and rare-security clauses. Separate version-binding gaps from auto-demotion with a full stop. | REQ-22 |
| ADR011-R2-04 | minor | Scope implemented claim to exact-version equality, with content binding enforced only for bundled fingerprinted packs. Record Semgrep action-ref and native guardrails binding gaps, including unenforced hand-bumped @N constants, for a later ADR. Gitleaks binary-segment invariant enforcement/test and short-version diagnostic remain pending later code; full activation string is unchanged. | ADR.md Decision/Options; REQ-22/78 and ADR-011 card; policy note; STATUS gap 3; plan gap 3; R-069 |
| ADR011-R2-05 | minor | Make STATUS and plan ranges ADR-000…ADR-011 with conditional final-zero-issue effectiveness; name ADR-011 as the latest conditional amendment. Put the ADR-specific route paragraph outside the index Rules list. Preserve all accepted historical ADR bodies/status lines; this clarification does not amend ADR-000. | STATUS sections 1/2; plan Docs that stay; index paragraph; ADR.md Changes |

## Submission conditions and evidence boundary

- Round-2 patch-result verification and commit-preservation conditions are superseded by the canonical separate-PR route above. No patch artifacts need verification.
- Writer SHA-256 values for every changed canonical file and new folder file, and `git diff --stat d2479d54`, accompany this submission. Ordinary diff stat excludes the five untracked ADR-folder files; their hashes are supplied separately.
- R-064/R-069 are prerequisites shared with later work, not an adoption of ADR-012's unsigned design. ADR-012 must preserve these append-only rows after rebasing; its selected designs belong in its own ADR.
- No product code, tests, experiments, builds, render checks, formatters or commits were run/made. Existing code is cited as static evidence reviewed by prior council rounds, not exercised implementation proof.
- Main should obtain ADR-011's independent zero-issue council sign-off on the canonical signed set, then land it first; ADR-012's subsequent review must identify its own rebased canonical bytes. No issue is claimed council-closed by this writer response.
