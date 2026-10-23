# Round 1 response — revision 2.2

Proposed file after this response: 462 lines (line numbers unchanged from the round-1 text), SHA-256 `2ae9bbdfffc35233fc3ca54955dc778d5fa67753bd25c8502f358686ecc31c52`. Base digest unchanged (`7ff21a7f…`). `docs/scout-design.md` untouched. All line numbers are in `scout-design.proposed.md`.

| # | Resolution | Lines |
|---|---|---|
| 1 | Clawpatch bound to providers `codex`/`claude` only; `claude` in host auth context (`CLAWPATCH_CLAUDE_AUTH_CONTEXT=host`, Claude Code ≥ 2.1.169, verified at `third_party/clawpatch/docs/providers.md:147-163`); other providers rejected; no `CLAWPATCH_*_API_KEY`/provider key env. The "model CLI logged in" check covers both Scout's adapter and Clawpatch; bootstrap reports CLI versions (verified `scripts/bootstrap-tools.sh:210-214`). Applied to §4, lane 9, §6 Model use, §10.2 Clawpatch row, §10.3 contract, header. | 5, 101, 120, 150, 337, 346 |
| 2 | Phase 3b runs the same request twice; the first version has no artifact channel (bench request carries none), so `scout.crap@1` is `not_applicable`; artifact supply and revision match wait for the schema round. | 291, 365–366 |
| 3 | "Audit labels" (annotation and stratum hidden, stored inclusion probability) are the only labels in weighted recall/AUROC; non-audit hidden-annotation dispositions are reported separately, unweighted, never in recall/AUROC. X1 and §9.7 restricted to audit labels. | 166, 308, 407 |
| 4 | Ban kept. The §9.7 line is `insufficient`/withheld until X1's preregistration is signed off, descriptive afterwards, never sets a threshold, band or default. | 166, 308 |
| 5 | X3/X10 reverted to rev 2.1 point rules (0.9, 0.8) at a dev-tuned threshold, flagged pending §D q5/q6. Uniform flag "rev 2.1 value, pending the user's minimums; the preregistration confirms or replaces it" on X1, X3, X4, X5, X10; header says so. | 5, 407, 409–411, 414 |
| 6 | Key holds `option_order_rule` (fixed list or permutation seeded by unit id); realised order stored in the request body. Flip-rate harness runs on each Choice's own labelled set (X1, X3). X3 states the seeded rule. | 153, 155, 409 |
| 7 | §9 preamble: Wilson only for unweighted binomial lines; weighted/paired estimators use a design-based bootstrap interval (95 %) named per preregistration. X4 precision stated as unweighted on the fully labelled hit stratum; X5 names a design-based interval. | 300, 410–411 |
| 8 | X5: split by repository; state shape fixed by the preregistration. X10: split by repository, threshold tuned on dev, scored once on test. | 411, 414 |
| 9 | Audit floor = a rank position fixed in the X1 preregistration; "≥ 20 injected true defects" counted separately from the ≥ 30 semantic positives; interval level 95 %. Rank-position cuts (incl. the floor) do not read P and need no reliability check. | 162, 407 |
| 10 | Interim: provisional stratified audit fraction, rev 2.1's 5 % total split across strata, tagged `[ASSUMPTION]`; bucket cut points unset (rank order only); per-rule lines `insufficient` until X1's preregistration. | 160, 162 |
| 11 | Exit-code effect stated once: an optional skipped lane is reported in `health`, summary and `readiness --matrix` and does not change the exit code; "never `clean`" removed everywhere (§3.4, §7.4, R13). | 88, 195, 398 |
| 12 | "and no tools" added. | 166, 462 |
| 13 | Cited by key name: drop `SCOUT_PROVIDER_*`/`SCOUT_PROVIDER_TOKEN`, the `OPENAI_API_KEY`/`ANTHROPIC_API_KEY` entries fed from `CLAWPATCH_{OPENAI,ANTHROPIC}_API_KEY`; constrain `CLAWPATCH_PROVIDER` (verified `scout-nightly.yml:135-142`). | 342 |
| 14 | Reworded as rejected for 3-way routing only; E2 (n = 22, pre-rules) did not show a Noul worse for ranking, cited by file lines with no figures, so no mixed-row ranges are quoted. | 168 |
| 15 | Research-ledger row R-038 added (`docs/research-ledger.md:46`); cited in the header and R14. STATUS §5 row 8 candidate renumbered R-038 → R-039. | 5, 399 |
| 16 | REQ-29 (stratified audit, version key, provisional 5 %) and REQ-39 (`subject capabilities|measure`, after Phase 3, §8.7) synchronised in `SCOUT_REBUILD_REQUIREMENTS.md:52,71`; §10.3 says so. | 346 |
| 17 | "proposed" removed from the status line. | 3 |
| 18 | Launch rank composes Jev outputs only (ordinal); any term mixing `rule weight` with a Jev probability falls under the reliability check. | 155, 162 |
| 19 | `PROPOSAL.md` amended: C1 section list completed; C7 adds §11 Phase 5 and §12 R14; out-of-scope sentence excepts the Phase 5 entry; round-1 note records the user's waiver of the §D entry condition except q8. | `PROPOSAL.md:9,15,17` |
| 20 | §15 item 1 limited to export contents and the dogfood ledger; consent and bench-run analyses cited as decided (Appendix 11). | 438 |
| 21 | Added "no such study is planned in Scout (see above)". | 273 |
| 22 | R4 mitigation names inertia-bench's adopter-independent studies (controlled continuations on curated repositories). | 389 |
| 23 | "share of owned files assigned the correct feature". | 409 |
| 24 | G4 cited as `validation-gates.json` `fix→fix` 24 of 30; interval marked derived. | 243, 411 |
| 25 | One scalar per bench measure id derived from `metric@version`; ids fixed in the schema round. | 287 |

## Other changes in this response
- Fixed a slip while editing: the E2 caveat cited in §6 "Rejected" is Caveat 6, as in the base.
- `DRAFT-NOTES.md` updated (digests, round-1 section).

## Files changed outside the proposed file
`docs/design-reviews/2026-10-22-rescope/{proposed.diff, DRAFT-NOTES.md, PROPOSAL.md, round-1-response.md}`, `docs/research-ledger.md`, `docs/STATUS.md` (row 8), `SCOUT_REBUILD_REQUIREMENTS.md` (REQ-29, REQ-39), `docs/hypothesis-ledger.md` (round-0 edit, unchanged).
