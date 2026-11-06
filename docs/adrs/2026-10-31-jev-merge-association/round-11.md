# Design review round 11: adoption of revision 2.8 (merge-time area association)

Date: 2026-11-01. Rules: CLAUDE.md design rules 1–3. Reviewers: the innovation council (lead), with independent peers `innovation-challenger` (GPT, adversarial) and `experiment-peer` (Opus, evidence). The review was read-only: no shell, no model calls, and `.codegraph` was not touched.

Reviewed files (SHA-256): `docs/scout-design.md` `d1fb127446be290f75d4de5416ebf40b214552f18eedc1d1a335819f314f9b75`; `SCOUT_REBUILD_REQUIREMENTS.md` `dab8fa330eecb90a1e4865b7842751306eb3addeadf3058806a106859fe7c797`

## Scope

This round checks, against HEAD `62760e0b23959feec4aab723e6ec211ffc9d4091`, the round-10 fixes and anything they broke:

- **R10-01:** `docs/scout-design.md:5`.
- **R10-02:** REQ-13 at `SCOUT_REBUILD_REQUIREMENTS.md:27`, and its `+` line in `requirements.proposed.diff`.
- **The follow-on header update:** `SCOUT_REBUILD_REQUIREMENTS.md:3` names the recomputed design digest and `round-11.md`.

Settled decisions are not reopened (R-059, R-061). The adoption is committed as documents only; the Phase 1 and Phase 2 code in the working tree is not part of that commit.

## Verification

HEAD versions were read from `raw.githubusercontent.com/adam-jackson-cf/scout/62760e0b…/`.

| # | Check | Evidence | Result |
|---|---|---|---|
| 1 | R10-01 applied word for word | :5 reads `…(§3.3, §10.2, §11 Phase 1 acceptance), so "Phases 1–4 are implemented" above does not cover revision 2.8's Phase 1 changes; implementation plan …`; `refers to the revision 2.7 acceptance` is gone | applied |
| 2 | :5 agrees with :3 and `docs/STATUS.md` | :3 is unchanged except the status marker and keeps "Phase 4 to the pre-2.6 acceptance" and the Phase 3 `pr_effort` caveat; :5 now makes no claim about Phases 2–4 (STATUS §1 :84, §5 rows 4 and 4c) | consistent |
| 3 | Plan file named at :5 | `docs/plans/2026-11-01-rev-2.8-merge-association.md` exists | ok |
| 4 | R10-02 applied word for word | REQ-13 reads `…with existing areas or records pending areas, which only a maintainer accepts or merges into an existing area; it never starts an agent map…`; `maintainer-accepted` appears in neither the requirements file nor the diff file | applied |
| 5 | REQ-13 agrees with the design and REQ-28 | design :114 ("Only a maintainer can accept a pending area or merge it into an existing one; until then it stays pending"), Appendix 23(a)–(b) :500, REQ-28 ("pending areas that only a maintainer accepts") | consistent |
| 6 | Requirements header :3 | names revision 2.8, sha256 `d1fb127446be290f75d4de5416ebf40b214552f18eedc1d1a335819f314f9b75` (64 hex characters) and `docs/design-reviews/2026-10-31-jev-merge-association/round-11.md`; `0ced69d8` and `round-10.md` appear nowhere | correct form; the value is bound below |
| 7 | Diff file `-` lines match HEAD | lines 1, 3, REQ-10, REQ-13, REQ-28 and REQ-31 match HEAD exactly | match |
| 8 | Diff file `+` lines match the working tree | lines 1, 3, 21, 27, 51 and 54 match exactly, including the R10-02 text and the new header fields | match |
| 9 | Nothing else changed in the requirements file | every other line is the same in HEAD and the working tree; 99 lines; the hunks (`-1,6`, `-18,13`, `-48,10`) and their context match | only 1, 3, 21, 27, 51 and 54 differ |
| 10 | Unchanged since round 10 | REQ-10, REQ-28, REQ-31; design :19 (`adds merge-time area association…`, with "proposes a pending new one" kept) and :378 (`(§3.3); revision 2.8 adds:`) | unchanged |
| 11 | No DRAFT marker left in the design | no `DRAFT`, `pending council` or `proposed, not adopted` for rev 2.8; 500 lines | ok |

## Issues

None.

## Considered and not raised

- **:5 cites round 8 while the requirements header cites round 11:** round 8 is the content sign-off of the revision 2.8 text. Round 11 signs the adopted bytes. Line 19 cites the review directory, which holds both. Rounds 9 and 10 accepted this.
- **REQ-31 "with none set" after the inserted exception:** its antecedent is still `model_cli.provider`. Readability only (round 10).
- **:3 "the revision 2.6 agent-map schedule awaits the companion code change":** stale at HEAD (STATUS row 4c, `ead1a3f`). It is in the bytes round 8 signed off and the adoption did not touch it, so fixing it is a separate design edit (round 10).
- **`docs/STATUS.md` still names revision 2.7 and `00ffd367…` (:7, :120, :208):** plan Phase 0 bookkeeping, outside the two files under review.
- **Limits of the evidence:** the council had no shell. It could not recompute the design SHA-256, the diff's `index a73d523f..ceacfddf` blob hashes, or a byte-level diff of the design against round 10; round 10 compared lines without markers by length. The sign-off below is therefore bound to the digests the main agent records, with the checks listed under Conditions.

## Peers

- **`innovation-challenger`:** zero issues. It confirmed both fixes word for word, the header form, and the diff-file line pairs.
- **`experiment-peer`:** zero issues. It confirmed both fixes, the header fields, and that the diff file is the whole change. It listed the shell checks under Conditions.
- **Lead:** zero issues. No peer changed the outcome.

## Review history

Revision 2.8 draft, rounds 1–8: 14 → 10 → 6 → 4 → 3 → 1 → 1 → 0 issues. Adoption, rounds 9–11: 3 → 2 → 0 issues. Records are in `round-1.md` through `round-11.md`.

## Conditions

The main agent confirms, before recording the digests and committing:

1. `shasum -a 256 docs/scout-design.md` prints `d1fb127446be290f75d4de5416ebf40b214552f18eedc1d1a335819f314f9b75`, the digest named in `SCOUT_REBUILD_REQUIREMENTS.md:3`.
2. `git diff 62760e0 -- SCOUT_REBUILD_REQUIREMENTS.md` equals `requirements.proposed.diff`.
3. `git diff 62760e0 -- docs/scout-design.md` changes only lines 3, 5, 19, 98, 114–116, 164–166, 173, 176, 200, 361, 378, 426, 427, 438, 444, 462, 463, 476 and 498–500.
4. The adoption commit holds documents only (no Phase 1/2 code), so the :5 statement "Nothing of revision 2.8 is implemented yet" is true at that commit.

If any condition fails, this sign-off does not hold and a new round is needed.

## Sign-off

The innovation council signs off the adoption of revision 2.8, merge-time area association with a reasoning-model fallback, with zero issues, under CLAUDE.md design rules 1 and 2. The sign-off covers:

- `docs/scout-design.md`: the full adopted text, that is, the revision 2.8 content signed off in round 8 plus the adoption edits reviewed in rounds 9–11 (the status line, the adopted notice at :5, the revision paragraph at :19, and the DRAFT markers removed or reworded);
- `SCOUT_REBUILD_REQUIREMENTS.md`: its header and REQ-10, REQ-13, REQ-28 and REQ-31, synced to revision 2.8 (`requirements.proposed.diff`).

This sign-off holds only while both files match, byte for byte, the SHA-256 digests recorded in the `Reviewed files (SHA-256)` line of this record. The design digest must equal `d1fb127446be290f75d4de5416ebf40b214552f18eedc1d1a335819f314f9b75`. The council computed no digests; the main agent records them. Any later change to either file voids this sign-off and needs a new council round with zero issues.

As revision 2.8 itself states:
- the lane is a provisional opt-in, off by default (Appendix 23);
- X11 needs a council-signed preregistration before any run;
- the `orderflip` flip rate on `merge.which_area` must be recorded before a release lets the opt-in lane be enabled (Appendix 19, B-3).

## Verdict

**ZERO ISSUES — signed off**
