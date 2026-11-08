# ADR-011: Keep the Gitleaks rule-pack activation version

- **Status:** accepted; decided by the agent under owner delegation (R-064)
- **Design revision:** `docs/scout-design.html`, REQ-22 and REQ-78, Appendix B
- **Decided:** 2026-11-08; **sources:** R-064, R-069

## Decision

Keep the exact-string rule-version activation contract. For eligible `security.secrets.<rule>` checks from the current bundled Gitleaks pack, `blocking.checks[].version` is:

```text
gitleaks-default-rules/8.30.1@sha256:e163e53b9e7e8a8511e77271e2b323ed057759542a6d988258afe3a1fa329caf
```

This is not the binary version `8.30.1`. Every eligible rule in this pack uses the same pack version. `CheckStatus` compares activation with the finding's rule version exactly. Exact-version equality is implemented; content binding is enforced only for bundled fingerprinted rule packs, whose rule-content digest changes return the check to shadow. Eligibility exclusions, including `generic-api-key`, do not change. Keeping today's string changes no current behaviour and needs no migration; equality alone does not guarantee a reset on a binary/parser or Scout-native logic bump.

Under the owner's delegation, select the council's recommended additional invariant: the Gitleaks pack label's version segment must equal the pinned `Binary.Version`. A later implementation PR must enforce that equality in spec validation with a test, so a binary bump forces a new activation version. In that same PR add a clear short-version diagnostic naming the required full string and explaining shadow fallback. Neither validation change ships in this documentation unit. Semgrep-backed `security.ci-mutable-action-ref` has no binary version in its rule-pack key. Scout-native guardrails reset only when someone hand-bumps their `@N` version constant, an unenforced convention. Those parser/logic-version binding gaps remain open for a later ADR, not guarantees delivered here.

## Options and rationale

- **Selected (a): retain the long pack string.** Exact-version equality and bundled-pack content binding are implemented; no migration is needed. Enforce the Gitleaks binary-segment invariant in later code rather than claim it is already checked.
- **Rejected (b): short tool version plus separate digest field.** Needs a new contract and migration without a demonstrated benefit. Short version alone cannot protect rule-content resets.
- **Rejected:** remove parser-bump intent. The existing intent remains; the Gitleaks invariant and the separately recorded Semgrep/native-guardrails decision gaps make the delivery boundary explicit.
- **Separate canonical PR:** ADR-011 lands with only its own canonical document changes and depends on no later ADR. This supersedes the earlier patch/hash split and its commit-preservation restriction; no patch application or merge-method restriction is required. A later revision touching the same files rebases onto ADR-011's adopted bytes and is reviewed on its own.

## Informed by

- R-064: owner direction, verbatim: "can you stop posing councile decisions to me please, go with your recommended each time and just capture them as adr's i can review later - so we can complete our entired implementation plan autonomously". This is agent authority to choose, not an owner-reviewed selection or a council waiver.
- R-069: `internal/analyzers/catalog_security.go:217,262,281` distinguishes binary and pack keys (and the Semgrep exception); `internal/analyzers/spec.go:199-226` validates bundled content fingerprints, not equality with the binary version. `internal/policy/policy.go:84,293,367-375` stores a nonempty version and compares it exactly. `internal/guardrails/suppress.go:49`, `focused_tests.go:9` and `config.go:39` use hand-maintained version constants. `docs/references/policy-file.md:21` describes rule/rule-pack activation; `docs/adrs/2026-10-23-decisions/PROPOSAL.md:43` left the ambiguity as an implementation note. Existing code is evidence only; no experiment evidence is used.

## Changes

- REQ-22/78 and Appendix B: distinguish implemented exact equality and bundled-pack content binding from pending Gitleaks binary-segment enforcement and the separate Semgrep/native-guardrails gaps. Preserve the existing maintainer, window, repository, counting and rare-check clauses; clarify auto-demotion's subject as that repo's check. This clarifies ADR-000's intent, without amending its accepted record.
- Policy reference: exact string, no migration and the same delivery boundary.
- Index, STATUS, remaining-work plan and ledger: delegated decision; remove gap 3's long-string ambiguity while retaining pending implementation and decision work. R-064 records direction and R-069 static facts only, not unsigned ADR-012 selections.
- Canonical signed set: this summary, `docs/scout-design.html`, `docs/adrs/README.md`, `docs/references/policy-file.md`, `docs/STATUS.md`, `docs/plans/2026-11-06-remaining-work.md` and `docs/research-ledger.md`.
- No product code, tests, tool pins or existing accepted ADR bodies/status lines change.

## Review

This summary is among the signed files, so it does not contain its own outcome or digests. Adopted-form route (a) (ADR-009 `round-13.md` step 5): this canonical revision takes effect on the explicit zero-issue final `round-<N>.md` in this folder (highest N), naming the review outcome and signed SHA-256 digests. Adoption is recorded outside the signed set without editing signed files. No sign-off is asserted here. ADR-011 is a separate PR and depends on no later ADR. The main agent's separate-PR direction (`round-2-response.md`) supersedes the historical patch/result-hash mechanism, exact-parent restriction and no-squash requirement. If canonical signed bytes change before adoption, submit the changed bytes for council review again. R-064 lands here first; dependent work reuses it without duplicating or editing the accepted row. Pre-allocated source IDs are retained; append order need not be numeric.

Owner confirmation is recorded as a new `user` research-ledger row and in editable README/STATUS/plan trackers; owner rejection requires a superseding ADR and its permitted status-line update. The pending review state belongs in those trackers, not this frozen status, card or append-only ledger rows. Delegation does not waive council or experiment rules.

## Records in this folder

- `ADR.md`: decision, delegated choices and canonical review route.
- `round-1.md`, `round-2.md`, `round-1-response.md`: byte-identical historical records from the earlier combined worktree; their patch references describe the superseded submission mechanism.
- `round-<N>.md`: council issues, severity, fixes and verdict; final zero-issue sign-off and digests.
- `round-<N>-response.md`: per-issue fixes and submission conditions.
