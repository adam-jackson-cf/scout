# Design review round 10: adoption of revision 2.8 (merge-time area association)

Date: 2026-11-01. Rules: CLAUDE.md design rules 1–3. Reviewers: the innovation council (lead), with independent peers `innovation-challenger` (GPT, adversarial) and `experiment-peer` (Opus, evidence). The review was read-only: no shell, no model calls, and `.codegraph` was not touched.

Reviewed files (SHA-256): `docs/scout-design.md` `0ced69d8c4cbea206a1bc628d0793c06c06443ce281ec860eea18ea47da42019`; `SCOUT_REBUILD_REQUIREMENTS.md` `9edadcbab9dff30e08541478125d469819e092209c16a88c8d3cef8c5ca452b1`

## Scope

This round checks the round-9 fixes against HEAD `62760e0b23959feec4aab723e6ec211ffc9d4091` for both files:

- **R9-01:** `docs/scout-design.md:19`.
- **R9-02:** the `SCOUT_REBUILD_REQUIREMENTS.md` sync, which changes the header, REQ-10, REQ-13, REQ-28 and REQ-31; the diff is in `requirements.proposed.diff`.
- **R9-03:** `docs/scout-design.md:5`.
- **Optional I-5:** the lower-case fix at `docs/scout-design.md:378`.

Settled decisions are not reopened (R-059, R-061). The adoption is to be committed as documents only; the Phase 1 and Phase 2 code in the working tree is not part of that commit.

## Verification

HEAD versions were read from `raw.githubusercontent.com/adam-jackson-cf/scout/62760e0b…/`.

`docs/scout-design.md`: 25 lines differ from HEAD, the same set as in round 9 (3, 5, 19, 98, 114–116, 164–166, 173, 176, 200, 361, 378, 426, 427, 438, 444, 462, 463, 476, 498–500). Lines without markers were compared by length only, not byte by byte.

| Item | Line | Finding | Result |
|---|---|---|---|
| R9-01 | :19 | `adds merge-time area association as a provisional opt-in lane`; "proposes a pending new one" is unchanged | applied |
| I-5 | :378 | `(§3.3); revision 2.8 adds:` | applied |
| R9-03 | :5 | "Nothing of revision 2.8 is implemented yet: neither the lanes nor the schema-13 state changes it makes to Phase 1 (§3.3, §10.2, §11 Phase 1 acceptance)" is accurate at HEAD (`internal/state/store.go:25` `schemaVersion = 12`; no association lane code). The rest of the sentence is wrong | **R10-01** |
| Other differing lines | 3, 98, 114–116, 164–166, 173, 176, 200, 361, 426, 427, 438, 444, 462, 463, 476, 498–500 | marker changes only, as in round 9; no requirement, number or rule changed | consistent |
| DRAFT or pending markers left | whole file | none for rev 2.8 | consistent |

`SCOUT_REBUILD_REQUIREMENTS.md`: the `-` lines of the diff file match HEAD, and its `+` lines match the working tree. The other non-blank lines are identical, so the diff file is the whole change.

| Line | Change | Checked against | Result |
|---|---|---|---|
| 1 | aligned to revision 2.8 | — | consistent |
| 3 | rev-2.8 sync entry; revision 2.8 is authoritative, with a sha256 and the signing file | — | consistent in form; the digest and signing file must follow the final round (see Issues) |
| 21, REQ-10 | `user_version` 13 adds `associations`; refuses 1, 2 and 4–12; cites Appendix 22 and 23(e) | :98, :361, :378, :426, :499, :500(e) | faithful |
| 27, REQ-13 | association lane between agent maps; never starts an agent map; the guard only recommends a refresh | :114, :116, :498, :500(a)–(b) | **R10-02** |
| 51, REQ-28 | issue-triage output is an `annotation` row; association writes `associations` rows and pending areas that only a maintainer accepts; never changes issue state, the gate, identity or a signal series | :114, :164, :176, :500(b) | faithful |
| 54, REQ-31 | `association-fallback` runs `codex exec` with `gpt-6.1-terra` at `high` only, never `claude`; an exception to Appendix 22(c) like the agent map | :114, :165, :500(d) | faithful |

## Issues

### R10-01 (medium): the new :5 sentence overstates which acceptance Phases 1–4 meet

- **Where:** `docs/scout-design.md:5`.
- **Current:** `so "Phases 1–4 are implemented" above refers to the revision 2.7 acceptance;`
- **Problem:** this contradicts :3 and `docs/STATUS.md`:
  - :3 says Phase 4 is implemented only "to the pre-2.6 acceptance", and that Phase 3 passed except that `pr_effort` has not been observed in CI;
  - `docs/STATUS.md` §1 :84 says "Phase 4 is done to the pre-2.6 acceptance";
  - §5 row 4 (:165) says the rev 2.6 acceptance needs a re-run, and row 4c (:167), that re-run, is still open.

  The revision 2.7 acceptance for Phase 4 is the one revision 2.6 amended, so this sentence turns a Phase 1 disclosure into a claim about Phases 3 and 4. R9-03 asked for a statement about Phase 1 only.
- **Fix:** replace `so "Phases 1–4 are implemented" above refers to the revision 2.7 acceptance;` with `so "Phases 1–4 are implemented" above does not cover revision 2.8's Phase 1 changes;`.

### R10-02 (medium): REQ-13 says the lane records pending areas that are already maintainer-accepted

- **Where:** `SCOUT_REBUILD_REQUIREMENTS.md:27` (REQ-13), and the same `+` line in `docs/design-reviews/2026-10-31-jev-merge-association/requirements.proposed.diff`.
- **Current:** `labels residual files of default-branch merges with existing areas or records maintainer-accepted pending areas;`
- **Problem:** the lane records a pending area on a fallback `new` answer, and the area stays pending until a maintainer accepts it or merges it into an existing area:
  - design :114: "recorded on a fallback `new` answer and never accepted by a model … until then it stays pending";
  - design :114: `Area.Active()` is false for a pending area;
  - Appendix 23(a)–(b), :500.

  "Maintainer-accepted pending areas" mixes up recording with acceptance. It could be read as making the lane's pending areas owners. REQ-28 states this correctly.
- **Fix:** replace `or records maintainer-accepted pending areas;` with `or records pending areas, which only a maintainer accepts or merges into an existing area;`.

### What follows from both fixes (not a separate issue)

R10-01 changes the design bytes, and this round does not sign off. So `SCOUT_REBUILD_REQUIREMENTS.md:3` and its line in `requirements.proposed.diff` must name the recomputed SHA-256 of `docs/scout-design.md` and `docs/design-reviews/2026-10-31-jev-merge-association/round-11.md` in place of `0ced69d8…2019` and `round-10.md`.

## Considered and not raised

- **:5 "the lanes" (plural) next to "the lane it adds":** the design defines two lanes, `association` and `association-fallback` (:114).
- **:5 citations `(§3.3, §10.2, §11 Phase 1 acceptance)`:** these match :98, :361 and :378. §7.2 is covered by :378's `associations.jsonl` round-trip.
- **:3 "the revision 2.6 agent-map schedule awaits the companion code change":** stale at HEAD (STATUS row 4c records `ead1a3f` as committed), but it is in the bytes round 8 signed off and the adoption did not touch it. Fixing it is a separate design edit.
- **REQ-13 "its refresh guard only recommends `scout map --refresh`":** in context "only" contrasts with starting an agent map, which is accurate (:116, Appendix 21 as amended). The `catalogue-stale` flag is a mechanism the design defines and this file cites.
- **REQ-13 "residual files" leaves out inherited and rule-(f) files:** the same level of summary as :19.
- **REQ-28 "Jev … only annotates" next to association rows:** the parenthetical is scoped to issue-triage output, and the next clause states the association output (:164, :176).
- **REQ-28 "reasoning LLMs are opt-in discovery":** the association fallback is covered by the association clause just before it, and REQ-31.
- **REQ-31 "with none set" after the inserted exception:** the antecedent is still `model_cli.provider`. Readability only.
- **REQ-13 added beyond R9-02's list:** the header lists it, and the change is accurate apart from R10-02.
- **Working-tree `internal/state/store.go:25` `schemaVersion = 13`:** ":5 nothing … is implemented" holds only if the uncommitted Phase 1/2 code stays out of the docs-only adoption commit, as main stated.
- **`docs/STATUS.md` still names revision 2.7 and `00ffd367…` (:7, :120, :208):** plan Phase 0 bookkeeping, outside the two files under review.

## Peers

- **`innovation-challenger`:** 2 issues, the same as the lead's (R10-01 and R10-02). Its wording for R10-01 ("so Phase 1 is implemented to the pre-2.8 acceptance; the Phase 2–4 qualifications above remain…") was not adopted, because it adds a new claim.
- **`experiment-peer`:** 2 issues, the same as the lead's. Its R10-01 replacement is adopted because it makes no new acceptance claim. It also confirmed that the requirements diff file is the whole change to that file.
- **Lead:** found R10-01 and R10-02 before either peer reported, so no peer changed the outcome. The REQ-13 fix adds "or merges into an existing area" from Appendix 23(b).

## Review history

Revision 2.8 draft, rounds 1–8: 14 → 10 → 6 → 4 → 3 → 1 → 1 → 0 issues. Adoption, rounds 9–10: 3 → 2 issues.

## Next round

Round 11 reviews `docs/scout-design.md` with R10-01 applied, and `SCOUT_REBUILD_REQUIREMENTS.md` (plus `requirements.proposed.diff`) with R10-02 and the updated digest and signing file. It is bound to the SHA-256 digests of both files and confirms nothing else changed against HEAD `62760e0`.

## Verdict

The council does not sign off the adoption of revision 2.8 at the reviewed digests of `docs/scout-design.md` and `SCOUT_REBUILD_REQUIREMENTS.md`.

**2 ISSUES**
