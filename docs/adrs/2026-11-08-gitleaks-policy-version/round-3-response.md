# Round 3 response: ADR-011

Date: 2026-11-08. Writer response to the four minor issues. This records fixes, not council closure or sign-off. All recommended options selected under R-064; no product code, tests, commits or experiments.

| Issue | Fix made |
|---|---|
| ADR011-R3-01 | STATUS latest research entry is R-069 with append-order disclosure. The ledger pointer lists R-001–R-064 and R-069, reservations R-065–R-068/R-070–R-077 across X1/ADR-012/X11, and next free R-078. Ledger introduction and plan unchanged as recommended. |
| ADR011-R3-02 | Reduced unmerged R-069 to rule-pack activation/version-binding facts only: specified title, activation/spec/policy/guardrails/reference/proposal links, Informed ADR-011; REQ-22/78, first two factual sentences and the exact static-evidence caveat. Readiness implementation facts now reside only in ADR-012's unmerged R-077, expanded with comparison enum/report, consumer, event upsert, state/export/merge and CLI citations. No ADR-012 file other than that R-077 row was edited while its review is running. |
| ADR011-R3-03 | Card Changes now says this canonical PR depends on no later ADR and supersedes the patch submission. ADR.md Options uses the recommended ADR-012-free wording. Review cites route (a) to ADR-009 round-13 step 5, says this PR depends on no later ADR, and identifies the main agent's separate-PR direction via round-2-response.md. Editable sequencing trackers remain unchanged. |
| ADR011-R3-04 | Submission inventory is exactly the seven canonical signed files declared in ADR.md. Responses are separately hashed reviewed records, not signed files. No response was added to the canonical declaration. |

## Exact canonical signed set (seven files)

1. `docs/adrs/2026-11-08-gitleaks-policy-version/ADR.md`
2. `docs/scout-design.html`
3. `docs/adrs/README.md`
4. `docs/references/policy-file.md`
5. `docs/STATUS.md`
6. `docs/plans/2026-11-06-remaining-work.md`
7. `docs/research-ledger.md`

`round-1-response.md`, `round-2-response.md` and this `round-3-response.md` are reviewed response records only; their separately supplied SHA-256 values do not put them in the signed set. Prior rounds/responses are unchanged. ADR-012's R-077 adjustment is outside this PR's signed set and is separately reported to its integration owner.

Condition C1 remains a merge-order coordination action for the main agent: X1 must not merge before ADR-011, then must drop its duplicate R-064 and reuse the adopted row under its own review. This response does not claim that external coordination has occurred. If the canonical bytes change before adoption, renew council review on the actual final bytes. Adoption is recorded outside signed files only after an explicit zero-issue final round.
