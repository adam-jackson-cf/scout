# ADR-005: Pin agent mapping to codex gpt-6-luna at xhigh

- **Status:** partly superseded by ADR-006, ADR-007
- **Design revision:** 2.5 (adopted 2026-10-24)
- **Decided:** 2026-10-23; **sources:** R-040, R-041, HD5, user direction

## Decision

- Clawpatch agent mapping uses provider `codex`, model `gpt-6-luna`, and reasoning effort `xhigh` only.
- The model and effort are fixed in code, not configuration.
- Mapping ignores the configured provider and `CLAWPATCH_PROVIDER`, which govern only `clawpatch-review` at this revision.
- Mapping never uses `claude` or a model/provider fallback.
- `clawpatch-review` and `contextual` may still use `claude`.
- Mapping reports `skipped (no model CLI)` when `codex` is absent from `PATH`.
- A failed installed CLI call keeps the heuristic map and reports exploration incomplete without changing areas or issue identity.
- The area-match default stays Jaccard 0.5 as an assumption pending X2b and a later council decision.

## Informed by

- R-040 records the user's 2026-10-23 instruction to pin mapping and retain 0.5 pending a codex-only X2 repeat.
- R-041 records X2's signed-off codex result on model option `default`, with probe-served `gpt-6.1-sol` and unrecorded effort.
- X2 selected `adopt` at `t* = 0.30`, its only passing threshold and the lowest measured grid point, with a contiguous pass set.
- X2 reported survival margin 0.00354238984269944 and survival at 0.50 of 0.7207172342621259 for axum and 0.7662488708220415 for fastapi (`round-1.md`, Verified correct).
- X2's `claude-opus-5-5` arm returned `fallback`, with replication `not replicated`.
- X2 covered one-step clean-layout transitions in fastapi (Python) and axum (Rust), two clusters without a cluster-robust interval, and one run of seven maps per repository.
- X2 used model-made `agent_reviewed` labels unreviewed by a human under `x2-concern-v1/guide-1`.
- A mapper truly at 5 % spurious is not adopted at a passing threshold about three times in four (`scout-design.proposed.md`, A1).
- X2 does not establish an area default for the pinned mapper.
- E3 and CP07 recorded model option `default`, not a served model, so neither establishes timing or churn for the pinned configuration.
- The one-run smoke maps remain operational observations, not signed-off experimental evidence.

## Changes

- This revision narrows ADR-000’s agent-mapping policy and ADR-002’s mapping-provider policy without changing the subscription-CLI or no-API-key rules.
- It amends the rev 2.4 baseline from ADR-004 in `scout-design.proposed.md` (signed revision 2.5) §§4, 6, 10.2–10.3, 12–14 and Appendix items 8 and 20.
- It adds risk R15 for model retirement, account unavailability and undetected same-name provider changes.
- It records X2's result without applying 0.30 and adds the planned X2b experiment to `scout-design.proposed.md` (signed revision 2.5) §13.
- It changes `SCOUT_REBUILD_REQUIREMENTS.md` REQ-31 to distinguish mapping from review through the signed `requirements.proposed.diff`.
- Adoption appends R-040 and R-041 and adds a wording-only HD5 History entry without changing HD5's state (`round-6-response.md`).
- ADR-006 closes `docs/scout-design.md`@9a3d19f5 Appendix 20's pending X2b clause and limits agent mapping to onboarding and manual refresh (revision 2.6 paragraph).
- ADR-007 removes `CLAWPATCH_PROVIDER` entirely while retaining the codex-only mapping pin (`docs/scout-design.md`@9a3d19f5, rev 2.7 paragraph).

## Conditions

- X2b must fix the X2/X2b combination before collection, including `fallback` and `inconclusive` branches.
- A later council design round decides the resulting proposal rather than automatically adopting a threshold.
- A model change needs a code change and a new council-signed study for that model before adopting an area default.
- Same-name provider changes remain undetected by probes that echo the requested name.
- The adoption record confirms the final sign-off's digest and diff-equality conditions were met before adoption.

## Review

- Six rounds: 17 → 14 → 12 → 4 → 3 → 0 issues.
- Reviews corrected X2 scope claims, separated missing-CLI skips from failed calls, clarified X2b decision authority, and fixed ledger ordering and record accuracy.
- Final sign-off: `round-6.md`; signed digests: `scout-design.proposed.md`: `032285b35e18f027aee0503af0890009fe13b45a36b56185586c5d075275acf9`; `proposed.diff`: `778842853667a5a882b261ca8875e4a47c7697e72cebcf02eaf2f9bc749b0fda`; `requirements.proposed.diff`: `4cfdd24fb96bec040bc4112b79b5b80e75c92bc953bef22876b60f742a2b5463`.
- The final round also names `scout-design.base.md`: `5c16a807513d54667390278755b8daa504e8db5ae0733236f87938c51a9a90ae`; `SCOUT_REBUILD_REQUIREMENTS.md` base as reviewed in round 6: `cd84592fc287fd5e4041c13fed827900c4143bdd583aee447ac59c6f65f71da2`; `PROPOSAL.md`: `1a59951c3536e8ac82d377f48996f5b8ceb3eddba618889974267857314a6686`; `DRAFT-NOTES.md`: `6c5ec06f88615cf0a33a7aca8c3fa53e5f0a8802fe1235f1fbc718aed3a153ab`.
- Adoption record: `round-6-response.md` confirms the corrected requirements digest, pre-adoption checks, design replacement, requirements patch and companion ledger edits.
- `docs/STATUS.md`@9a3d19f5 §1 records adoption on 2026-10-24 without an adoption commit.

## Records in this folder

- `ADR.md`: Summary of the decision and its later amendments.
- `PROPOSAL.md`: User direction, code facts, evidence limits and proposed changes.
- `DRAFT-NOTES.md`: Artifact digests, changed-line inventory and adoption instructions.
- `scout-design.base.md`: Frozen rev 2.4 design baseline.
- `scout-design.proposed.md`: Signed rev 2.5 design snapshot.
- `proposed.diff`: Design changes from the baseline to the signed proposal.
- `requirements.proposed.diff`: REQ-31 mapping-provider carve-out in `SCOUT_REBUILD_REQUIREMENTS.md` as reviewed in round 6.
- `round-1.md`: Council review with 17 issues.
- `round-1-response.md`: Response to the first round's issues.
- `round-2.md`: Council review with 14 issues.
- `round-2-response.md`: Response to the second round's issues.
- `round-3.md`: Council review with 12 issues.
- `round-3-response.md`: Response including the directed ledger-adoption sequence.
- `round-4.md`: Council review with four issues.
- `round-4-response.md`: Response clarifying ledger ordering and later design authority.
- `round-5.md`: Council review with three record-keeping issues.
- `round-5-response.md`: Record corrections with the three signed change artifacts unchanged.
- `round-6.md`: Zero-issue sign-off with corrected requirements digest and adoption conditions.
- `round-6-response.md`: Condition checks and adoption record.
