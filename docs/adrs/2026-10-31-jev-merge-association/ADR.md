# ADR-008: Add provisional merge-time area association

- **Status:** accepted (provisional); amended by ADR-012 (schema 14, readiness export/merge)
- **Design revision:** 2.8 (adopted 2026-11-01)
- **Decided:** 2026-10-31; **sources:** R-055–R-061, HD8, user direction confirmed after the axum result.

## Decision

- Add two opt-in, off-by-default signal lanes on default-branch pushes only: `association` and `association-fallback`.
- Jev routes residual clusters through `merge.covered` and `merge.which_area` at provisional policy thresholds of 0.60 and 0.535 respectively.
- Send Jev's unresolved or `none` clusters to `codex exec` with `gpt-6.1-terra` at effort `high`, with separate model-CLI consent and no `claude` fallback.
- Store existing-area labels in association records rather than changing agent-map file sets.
- Record fallback `new` answers as pending areas that only a maintainer accepts or merges into an existing area.
- Keep pending areas out of ownership lookups and model catalogues, and preserve maintainer acceptance through later map reconciliation.
- Export association decisions and inherited-file records in `state/v1/associations.jsonl`, and move the current SQLite schema to `user_version` 13.
- Recommend a manual `scout map --refresh` through a refresh guard without starting an agent map or changing issue state, the gate, identity or signal series.

## Informed by

- R-055–R-058 record the merge-association development work and its fastapi and axum results.
- Fastapi rounds 8–10 covered 92 unseen clusters and gave Jev+Terra assignment precision 0.922 [0.840, 0.964] (71/77), recall 0.922 and zero unresolved clusters.
- Axum rounds 1–5 covered 109 clusters and gave assignment precision 0.622 [0.514, 0.719] (51/82), resolved recall 0.810 and zero unresolved clusters.
- Axum did not replicate fastapi, and its precision interval overlapped the directory baseline's interval.
- R-060 corrects the interpretation of axum's 46 gold-none clusters: 37 were existing coverage gaps and nine were new concerns.
- Both runs used agent-reviewed labels and teacher-forced membership, so closed-loop error accumulation and duplicate pending areas remain unmeasured.
- Neither development run had a preregistration or council review, so HD6 remains `inconclusive`.
- R-059 grants an exception to design rule 3 for this provisional opt-in only, and R-061 confirms that choice after the user saw axum's result.

## Changes

- Amend ADR-000's mapping policy as changed by ADR-006's Appendix 21 in `docs/scout-design.md`@9a3d19f5 for pending areas at merge time only, while retaining onboarding and manual-only agent-map runs.
- Amend ADR-007's Appendix 22(a) in `docs/scout-design.md`@9a3d19f5 from schema 12 to 13 and add 12 to the refused store versions, while retaining the v1 re-anchor and re-baseline remedy.
- Add an exception to ADR-007's Appendix 22(c) in `docs/scout-design.md`@9a3d19f5 so `association-fallback` ignores the policy provider and uses only `codex`.
- Extend `docs/scout-design.md`@9a3d19f5 §3.3, §4, §6, §7.2, §10.2, §11, R17–R18, X3, X11, A13–A14 and Appendix 23.
- Amend ADR-002's experiment list by adding X11: keep X3 for other unowned files and give merge-time association its own planned test, X11, tied to proposed hypothesis HD8.
- Sync the header and REQ-10, REQ-13, REQ-28 and REQ-31 in `SCOUT_REBUILD_REQUIREMENTS.md`@9a3d19f5 to revision 2.8 through `requirements.proposed.diff`.

## Conditions

- Record the `orderflip` flip rate on `merge.which_area` before a release permits enabling the opt-in lane; the 20 % second permutation does not replace this gate.
- Obtain council sign-off on X11's preregistration before any run; X11 determines whether to upgrade, retain provisionally or remove the lane.
- Treat the fallback-call cap of 40, pending-area cap of 10 and guard defaults of 20 merges, 20 clusters, 0.40 fallback rate and 0.20 pending-new rate as assumptions rather than validated values.
- Report a pending-new rate as `unmeasured` when its eligible merges lack the cluster floor or any answered fallback cluster.
- Disclose the guard's unknown recall and false-fire rate, including its blind spot for Jev's wrong assignments that never escalate.
- Keep the round-11 adoption sign-off conditional on the recorded digests, exact requirements diff, restricted design diff and documents-only adoption commit.

## Review

- Content review took eight rounds: 14 → 10 → 6 → 4 → 3 → 1 → 1 → 0 issues.
- Adoption review took three further rounds: 3 → 2 → 0 issues.
- Content sign-off: `round-8.md`; signed digests: `docs/scout-design.md` draft as reviewed in round 8: `7b102ebbabddebb9f49bdec5cd559bf6908ed32b1bd136f01ee1f5a4dc4b0551`; `docs/design-reviews/2026-10-31-jev-merge-association/proposal.md` as reviewed in round 8: `36d082ea61b6024091a14dad5120e0ef5d97d486ec83a8523eecb6d6fce52671`.
- Final sign-off: `round-11.md`; signed digests: `docs/scout-design.md`@9a3d19f5: `d1fb127446be290f75d4de5416ebf40b214552f18eedc1d1a335819f314f9b75`; `SCOUT_REBUILD_REQUIREMENTS.md`@9a3d19f5: `dab8fa330eecb90a1e4865b7842751306eb3addeadf3058806a106859fe7c797`.
- Adoption record: `round-11.md` signs the adopted design and requirements with zero issues on 2026-11-01.

## Records in this folder

- `ADR.md`: decision summary and record index.
- `proposal.md`: merge-time flow, policy values, development evidence, risks and user decisions reviewed for content sign-off.
- `requirements.proposed.diff`: requirements changes reviewed for adoption.
- `round-1.md`: fourteen issues in lane contracts, state, measured flow, evidence and validation.
- `round-2.md`: ten remaining issues after the first fixes.
- `round-3.md`: six remaining issues in inheritance, reconciliation, guard accounting and schema references.
- `round-4.md`: four remaining issues in pending renames, acceptance preservation and guard accounting.
- `round-5.md`: three remaining issues in rename precedence, obsolete paths and the guard's sample basis.
- `round-6.md`: one remaining issue in fallback-skip filtering and exported run-level provenance.
- `round-7.md`: one remaining issue when every fallback call fails.
- `round-8.md`: zero-issue content sign-off bound to the draft design and proposal digests.
- `round-9.md`: three adoption issues covering status wording, requirements sync and Phase 1 implementation disclosure.
- `round-10.md`: two adoption issues covering Phase 1 disclosure and pending-area acceptance wording.
- `round-11.md`: zero-issue adoption sign-off with signed digests and explicit conditions.
