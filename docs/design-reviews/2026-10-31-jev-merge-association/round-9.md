# Design review round 9: adoption of revision 2.8 (merge-time area association)

Date: 2026-11-01. Rules: CLAUDE.md design rules 1–3. Reviewers: the innovation council (lead), with independent peers `innovation-challenger` (GPT, adversarial) and `experiment-peer` (Opus, evidence). The review was read-only: no shell, no model calls, and `.codegraph` was not touched.

Reviewed files (SHA-256): `docs/scout-design.md` `d53500e2aa6f913e8b830714abbd3bde6fc24442bf6add5d18f07c1693115c24`

## Scope

Round 8 signed off the DRAFT bytes of revision 2.8. That sign-off was bound to digest `7b102ebb…0551` and did not cover the adoption steps. This round reviews only the adoption diff: the working-tree `docs/scout-design.md` against HEAD `62760e0b23959feec4aab723e6ec211ffc9d4091`. It checks three things:

1. the diff changes no requirement, number, rule or decision;
2. no DRAFT or pending wording that contradicts adoption remains;
3. the adopted notice is accurate.

Settled decisions are not reopened (R-059, R-061; round 8 §Scope).

## Diff verification

The HEAD version was read from `raw.githubusercontent.com/adam-jackson-cf/scout/62760e0b…/docs/scout-design.md`. Both versions have 500 lines. Exactly 25 lines differ, and the experiment peer confirmed the count by comparing the length of every line.

| Lines | HEAD | Working tree | Substance changed? |
|---|---|---|---|
| 3 | `**revision 2.7 (2026-10-24)**` | `revision 2.7 (2026-10-24), **revision 2.8 (2026-10-31; adopted 2026-11-01)**` | status only |
| 5 | DRAFT notice: "proposed, not adopted … the adopted design is revision 2.7" | adopted notice: round 8, rounds 1–8 14 → 0, provisional opt-in, off by default, not yet implemented, plan path | status (R9-03) |
| 19 | `(DRAFT, not adopted;` … `get DRAFT amendments` | `(adopted 2026-11-01 after council round 8;` … `get amendments` | status (R9-01) |
| 98 | `DRAFT revision 2.8 makes it 13`; `(12; DRAFT 13)`; `(1, 2, 4–11; DRAFT 4–12, …)` | `revision 2.8 makes it 13`; `(13 from revision 2.8; 12 through revision 2.7)`; `(1, 2, 4–12 from revision 2.8, …; 4–11 through revision 2.7)` | no; 12/13 and 4–11/4–12 kept |
| 114 | `(DRAFT, revision 2.8; …; not adopted until council sign-off; R-055 …)` | `(revision 2.8; …; R-055 …)` | no; body identical |
| 115, 116 | `(DRAFT)`, `(DRAFT; design only, …)` | `(revision 2.8)`, `(revision 2.8; design only, …)` | no; bodies identical |
| 164–166, 438, 444, 463 | `DRAFT rev 2.8` | `rev 2.8` | no |
| 173, 176, 462 | `DRAFT revision 2.8:` | `Revision 2.8:` | no |
| 200 | `is DRAFT revision 2.8` | `from revision 2.8` | no |
| 361, 426 | `DRAFT revision 2.8` | `revision 2.8` | no |
| 378 | `(DRAFT revision 2.8: 4–12 …)`; `; DRAFT revision 2.8 adds:` | `(revision 2.8: 4–12 …)`; `; Revision 2.8 adds:` | no |
| 427 | `DRAFT rev 2.8:` | `Rev 2.8:` | no |
| 476 | `item 23 DRAFT 2026-10-31, R-059, pending council sign-off` | `item 23 2026-10-31, R-059, adopted with revision 2.8` | status |
| 498, 499 | `**DRAFT amendment (item 23[(e)], pending council sign-off)` | `**Amendment (item 23[(e)], revision 2.8)` | no |
| 500 | `23. **DRAFT, pending council sign-off.** Merge-time …` | `23. Merge-time …` | no |

Leftover markers: no rev-2.8 "DRAFT", "drafted", "pending council sign-off", "not adopted until", "until signed off" or "adopted design is revision 2.7" remains. The other "pending" and "proposal" uses are unrelated:

- pending areas and the pending-new rate;
- "pending the user's minimums" (:7, :435–443, :470) and "X1 pending" (:184);
- "Pending clause closed" and "resulting proposal" (:497), "HD8 (proposed, …)" (:444);
- `ProposedAction` (:352) and the earlier revisions' `PROPOSAL.md` references.

Adopted notice, checked against the evidence:

- **Round 8:** `round-8.md` exists, says "ZERO ISSUES — signed off", and its history is 14 → 10 → 6 → 4 → 3 → 1 → 1 → 0, so "14 → 0" is accurate.
- **Provisional opt-in, off by default:** matches :114 (`mapping.association.enabled` defaults to false) and Appendix 23(c).
- **Plan file:** `docs/plans/2026-11-01-rev-2.8-merge-association.md` exists. Its Phase 0 matches this diff.
- **Not yet implemented:** at HEAD, `internal/state/store.go:25` is `const schemaVersion = 12`, and no association lane, residual or cluster code, Jev question, fallback or guard exists. The working tree holds uncommitted schema-13 state code (`internal/state/associations.go`, `internal/domain/association.go`).
- **No false claim from the rewording:** X11 is still "Planned, not yet preregistered" (:444). The guard is still "design only" (:116). A14 keeps its defaults as **[ASSUMPTION]**. :98's "13 from revision 2.8" is a design statement, not a claim about the code.

## Issues

### R9-01 (low): the revision 2.8 paragraph still describes the revision as a proposal

- **Where:** `docs/scout-design.md:19`.
- **Current:** `…HD6 stays \`inconclusive\`)** proposes merge-time area association as a provisional opt-in lane, off by default:`
- **Problem:** the adopted revision's summary still says it "proposes" the lane. Earlier revision paragraphs use verbs like records, re-scopes or corrects, and :5 says "the lane it adds".
- **Fix:** replace `proposes merge-time area association as a provisional opt-in lane` with `adds merge-time area association as a provisional opt-in lane`. Leave the later "proposes a pending new one" unchanged: it describes what the model does.

### R9-02 (medium): the live requirements file was not synced to revision 2.8

- **Where:** `SCOUT_REBUILD_REQUIREMENTS.md:1`, `:3`, `:21` (REQ-10), `:51` (REQ-28), `:54` (REQ-31).
- **Why it matters:** this file stays the live product spec until Phase 6 (Appendix 12, `docs/scout-design.md:370`).
- **What is wrong:**
  - its header still names design revision 2.7 as authoritative, with 2.7's digest and `cutover-cleanup/round-7.md`;
  - REQ-10 states schema `user_version` 12 and refuses 1, 2 or 4–11, which contradicts :98, :361, :378, R17 and Appendix 22(a) as amended;
  - REQ-28 says Jev output is only an `annotation` row, which contradicts :176 (`associations` rows);
  - REQ-31 says an opted-in model-CLI lane runs only the policy provider, which contradicts :114 and Appendix 23(d) (`codex` only, an exception to Appendix 22(c)).
- **Precedent:** every earlier revision synced this file in the same change (its header records the syncs for 2.3, 2.5, 2.6 and 2.7), and the council reviewed that sync (`requirements.proposed.diff` in `2026-10-24-area-refresh-fallback/` and `2026-10-24-cutover-cleanup/`). Round 8 explicitly left this step out of its sign-off.
- **Fix:** apply the following and submit it for round 10 as `docs/design-reviews/2026-10-31-jev-merge-association/requirements.proposed.diff`.
  - Line 1: replace `(aligned to design revision 2.7)` with `(aligned to design revision 2.8)`.
  - Line 3, first edit: replace `and this header and REQ-10, REQ-25, REQ-27 and REQ-31 to revision 2.7 on 2026-10-24.` with `this header and REQ-10, REQ-25, REQ-27 and REQ-31 to revision 2.7 on 2026-10-24, and this header and REQ-10, REQ-28 and REQ-31 to revision 2.8 on 2026-11-01.`
  - Line 3, second edit: replace ``revision 2.7 (sha256 `00ffd3677acd7dbaec94d7c0a88be0b0d5a18841ceed5f7dbffa6df08d8bf005`, council-signed in `docs/design-reviews/2026-10-24-cutover-cleanup/round-7.md`)`` with ``revision 2.8 (sha256 `<adopted digest of docs/scout-design.md>`, council-signed in `docs/design-reviews/2026-10-31-jev-merge-association/round-<N>.md`)``. N is the round that signs off the adopted bytes.
  - Line 21 (REQ-10), first edit: replace ``(`user_version` 12), and a store at 1, 2 or 4–11 is refused`` with ``(`user_version` 13 from design revision 2.8, for the `associations` table), and a store at 1, 2 or 4–12 is refused``.
  - Line 21 (REQ-10), second edit: replace `(design §3.3, Appendix 22).` with `(design §3.3, Appendix 22(a) as amended by Appendix 23(e)).`
  - Line 51 (REQ-28): replace ``(output is an `annotation` row; `issues` does not import `decisions`)`` with ``(output is an `annotation` row; `issues` does not import `decisions`; in the opt-in merge-time association lane of design revision 2.8, off by default, the output is an `associations` row instead, a label that never changes issue state, the gate, issue identity or a signal series, design §4, §6, Appendix 23)``.
  - Line 54 (REQ-31): replace ``An opted-in model-CLI lane runs only the policy's `model_cli.provider`;`` with ``An opted-in model-CLI lane runs only the policy's `model_cli.provider`, except the opt-in `association-fallback` lane (design revision 2.8), which runs only `codex exec` with `gpt-6.1-terra` at effort `high` and does not read the policy provider (design §4, Appendix 23(d));``.

### R9-03 (low): the adopted notice leaves "Phases 1–4 are implemented" (:3) inaccurate for revision 2.8's Phase 1 changes

- **Where:** `docs/scout-design.md:5`.
- **Problem:** revision 2.8 amends the §10.2 row (:361, schema 13 including `associations`) and the Phase 1 acceptance (:378: refuse a store at 12; round-trip `associations.jsonl` and pending `areas.jsonl`). HEAD implements none of these (`internal/state/store.go:25`, `schemaVersion = 12`). The notice says only that "the lane" is not implemented, so :3 now overstates Phase 1. Rev 2.6 qualified :3 in the same situation ("Phase 4 to the pre-2.6 acceptance"), and rev 2.7 disclosed it in its paragraph at :17.
- **Fix:** in :5, replace ``The merge-time association lane it adds is a provisional opt-in, off by default (Appendix 23), and is not yet implemented; implementation plan`` with ``The merge-time association lane it adds is a provisional opt-in, off by default (Appendix 23). Revision 2.8 is not yet implemented, neither the lane nor its state changes (`user_version` 13, the `associations` table and `associations.jsonl`; §3.3, §7.2), so Phase 1 is implemented to the pre-2.8 acceptance (§11); implementation plan``.
- **Condition:** this wording is true only if the adoption commit carries no Phase 1 code (plan Phase 0, docs only). The uncommitted schema-13 working-tree code must not go into that commit.

## Considered and not raised

- **Citing round 9 or 10 as the sign-off of the adopted bytes (:5, :19):** :5's claim that round 8 signed off the revision's text is true. :19 cites the review directory, which will hold the signing round; revisions 2.2–2.7 cite their reviews the same way. Once R9-02 is applied, the requirements header records the signing round and the adopted digest. Not a defect.
- **`(proposal Q2)` at :164 without a file path:** the review directory cited at :19 contains `proposal.md`, so the reference still resolves.
- **`; Revision 2.8 adds:` at :378:** a capital letter after a semicolon, left over from deleting "DRAFT". Cosmetic, no change in meaning. If it is lowercased in the same edit, round 10 checks it.
- **:19 "It touches §3.3, §4, §6, §7.2, §12, §13, §14":** the list omits §10.2 (:361) and §11 (:378). This text is unchanged from the bytes signed in round 8 and is outside the adoption diff.
- **"rounds 1–8: 14 → 0":** a shortened but accurate form of the history.
- **The "adopted 2026-11-01" date (:3, :5, :19):** correct only if the zero-issue round falls on 2026-11-01; otherwise the next edit must update it.
- **`docs/STATUS.md` rev-2.8 entry and phase list (plan Phase 0):** not part of the design and outside this round.

## Peers

- **`innovation-challenger`:** 1 issue (the "proposes" wording), adopted as R9-01. Otherwise it found the diff to be status-only and the notice accurate.
- **`experiment-peer`:** 6 items.
  - I-4 (= R9-01) is adopted.
  - I-6 (requirements sync) is adopted as R9-02. It changed the lead's leaning: the lead had first treated the requirements file as outside this round, but the precedent and the REQ-10/28/31 contradictions make it an adoption defect.
  - I-2 (Phase 1 implementation status) is adopted as R9-03.
  - I-1 (cite round 9), I-3 (proposal path) and I-5 (capitalisation) are not raised, for the reasons above.
- **Lead:** found the same 25-line diff and had R9-01 as a borderline issue; both peers raising it settled it.

## Review history

Revision 2.8 draft, rounds 1–8: 14 → 10 → 6 → 4 → 3 → 1 → 1 → 0 issues. Adoption, round 9: 3 issues.

## Next round

Round 10 reviews:

- `docs/scout-design.md` with R9-01 and R9-03 applied;
- `SCOUT_REBUILD_REQUIREMENTS.md` with R9-02 applied, plus its `requirements.proposed.diff`.

The round is bound to the SHA-256 digests of both files and confirms that nothing else changed against HEAD `62760e0`. As revision 2.8 itself states:

- X11 needs a council-signed preregistration before any run;
- the `orderflip` flip rate on `merge.which_area` must be recorded before a release lets the opt-in lane be enabled (Appendix 19, B-3).

## Verdict

The council does not sign off the adoption of revision 2.8 at the reviewed digest of `docs/scout-design.md`.

**3 ISSUES**
