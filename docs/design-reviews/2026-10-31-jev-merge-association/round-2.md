# Design review round 2: DRAFT revision 2.8, Jev merge-time area association with a reasoning-model fallback

Date: 2026-10-31. Rules: CLAUDE.md design rules 1–3. Reviewers: the innovation council (lead analysis) plus two independent peers, `innovation-challenger` (GPT, adversarial) and `experiment-peer` (Opus, evidence). The review was read-only: no shell, no model calls, `.codegraph` untouched.

Reviewed files (SHA-256): `docs/scout-design.md` `fc26115cdcf01cb99af12be2057bd8434cb0c5a1f7d143bb481ae7c7a27a8a36`; `docs/design-reviews/2026-10-31-jev-merge-association/proposal.md` `adb9bcff212139abeee845e8dc193d81760e59bc20fcf664fb21b8d6e7ca7591`; `docs/research-ledger.md` `6e1960a3ae70cbd67108b675554ed127da88c21f5784a11529dafacf25cb7495`; `docs/hypothesis-ledger.md` `d35525a18feeb36683db3b197c1acb86a05e4c11bde342b1c11adac3a1d5bd25`

## Scope

This round checks:
- the lines changed since round 1 in `docs/scout-design.md`: :19, §3.3 :98, §4 :114–116, §6 :164–166, §7.2 :200, :378, R17 :426, R18 :427, X3 :438, X11 :444, A13 :462, A14 :463, Appendix 21 :498 and 23 :500;
- `docs/design-reviews/2026-10-31-jev-merge-association/proposal.md`;
- the new ledger rows R-060 (`docs/research-ledger.md:68`) and HD8 (`docs/hypothesis-ledger.md:35,39`);
- anything the fixes broke.

Settled decisions that this round does not reopen (R-059, confirmed after axum, U1): a pending area is created at merge time; a maintainer must accept it; the lane is a provisional opt-in and off by default; the evidence is accepted as an exception to design rule 3; the fallback is `codex exec gpt-6.1-terra` at high effort only.

## Round-1 fixes: verification

| R1 | Status | Evidence |
|---|---|---|
| 01 lanes, skips, failures, 22(c) | applied; consent contradiction left (R2-03) | :114 two lanes, `jev_failure`/`jev_cap`, command-session void, codex-only; :500(d) |
| 02 pending-area state | applied | :114, :500(a), proposal step 4; fits `Area.Active()` / `Reconcile` (`internal/areas/areas.go:40-51,118-168`) |
| 03 persistence, schema | partly (R2-04, R2-06, R2-07) | :98, :114, :200, :378, :426, :462 |
| 04 measured flow | partly (R2-01) | clusters, ≤ 10 per merge, 60 lines, 5 samples, 20 % second permutation, PR-metadata absence and the 2–255 Choice limit are correct; the residual rules cite the wrong function |
| 05 staleness | record rule applied; threshold-version rule left (R2-05) | :114 |
| 06 B-3 before an enabling release | applied | :164; proposal Q2 |
| 07 own model-CLI consent | applied; prerequisite contradiction (R2-03) | :114; proposal Q4; removed from A14 |
| 08 gap/new, age as cause, 3.4 years | applied | :115, :116, :427, :463; R-060 appended, R-058 untouched |
| 09 teacher forcing disclosed | applied | :115, :463; proposal "Evidence" |
| 10 guard disclosure and keys | partly (R2-02) | :116 disclosure, `min_clusters`; the denominator contradicts the cited figures |
| 11 caps, keys, 0.380 | partly (R2-09) | `max_pending_per_run` and 71/187 correct; "per merged PR" misleads |
| 12 R-059 timing | partly (R2-08) | :19, :500(c), proposal U1; no ledger row; :19 wording overstates |
| 13 X11, HD8 | applied | :444; HD8 matches X11's starting bar |
| 14 X3 | applied | :438; proposal "Interactions" |

## Numbers re-derived

| Figure | Derivation | Result |
|---|---|---|
| fastapi rounds 8–9 fallback rate 20/44 = 0.45 | escalations 7 + 13; clusters 21 + 23 (`README.md` rounds 8–10) | correct |
| fastapi rounds 8–9 pending-new 14/44 = 0.32 | `none` outputs 3 (2/3) + 11 (10/11) over 44 residual clusters | numerator correct; denominator differs from the stated definition (R2-02) |
| axum: 30 of 31 wrong assignments are Jev's | Jev-only 77 − 47 = 30; Jev+Terra 82 − 51 = 31; Terra's 5 `existing` answers include 1 wrong | correct |
| 37 gap / 9 new; per round 13/23, 5/20, 3/22, 15/24, 10/20 | `axum/README.md:10`; `GUIDE.md:22-23` | correct |
| round 1 0–3 months after C0; up to 3.4 years | C0 2022-12-19; round 1 ends 2023-03-28; round 5 ends 2026-05-16 (`FROZEN-AXUM.md`) | correct |
| B2 rounds 1–7 0.380 = 71/187 | `README.md:114` | correct |
| 2.2 clusters per PR | 92/41; 200/92 | arithmetic correct; wording misleads (R2-09) |
| Jev-only none recall on axum | `axum/README.md` table: 0.217 (10/46) | the draft cites the combined-arm 0.500 instead (R2-10) |

## Issues

### R2-01 (high): the residual rules cite the pre-pass proxy, and the population differs from the measured one

This comes from R1-04's fix text.

`:114` and proposal step 1 say: "the files it added or moved. It applies the measured catalogue-free residual rules (`…frontend.py` `classify`): a heuristic feature or active area already owns the file, …". That function is the wrong one. `classify` is the catalogue-free pre-pass proxy (`frontend.py:6-7`: rules (a) and (f) "need the catalogue and are not evaluated here"). Both runs used `owned_classify` (`build_units.py:247`; `frontend.py:311-380`).

Where the draft and the measured front end differ:
- **Heuristic-only files.** A file owned only by a heuristic feature was decided, whether added, moved or modified (rule (f), `:340`). The draft excludes these files.
- **Excluded files.** Markdown, non-English translations and media were dropped (`excluded_reason`). Media files have no text excerpt to send.
- **Pooled files.** The measured "pool" (gold `none`) corresponds to pending areas, and the draft does not say so. In the measured runs a renamed pooled file stayed pooled, and a modified pooled file was a coverage figure, not a decided unit.
- **Cross-scope renames.** The measured cross-scope rename rule used repository-specific prefixes.

**Fix.** At `:114` replace from "takes, for each first-parent merge in the push, the files it added or moved." to "…and empty package markers are inert." with the text below, and make the same change in proposal step 1:

> takes, for each first-parent merge in the push, the files it added, moved or modified, and applies the measured residual rules (`evaluation/x2d-postrun-tuning/axum/pilot/frontend.py` `owned_classify`, the front end of both runs, `build_units.py:247`; `classify` is only the catalogue-free pre-pass proxy). An active area or an association record counts as an owner, and a pending area takes the role of the measured pool:
> - deleted files are dropped, and so are markdown, non-English translation and media files (`excluded_reason`);
> - a modified file that an owner holds is not decided, and a modified file that no owner holds and no heuristic feature owns is counted as a coverage gap and is not decided;
> - a renamed file inherits its old path's owner, or stays in its pending area, unless it is empty;
> - an added file with an owned variant sibling (same normalised directory, variant stem and extension) inherits that owner;
> - a file owned only by a heuristic feature is decided whether added, moved or modified (rule (f));
> - every other added or moved file is residual, and clusters made only of empty files are inert.
>
> The measured cross-scope rename rule (`frontend.py` `scope`, repository-specific prefixes) has no Scout equivalent and is not applied; this is an unmeasured difference.

### R2-02 (high): the pending-new rate's definition contradicts its figures and its limit

This comes from R1-10's fix text.

`:116` and proposal "Refresh guard" define the rate as "valid `new` answers ÷ clusters the fallback answered". The cited figures divide by all residual clusters: "0.32 (14/44)", "1.25 × the fastapi gold-`none` share of about 16 %", and axum's "25 per 100". Under the stated definition the rate would be 14/20 = 0.70 for fastapi rounds 8–9, 15/25 = 0.60 for rounds 8–10, 22/54 = 0.41 for rounds 1–7, and 27/32 = 0.84 for axum. The 0.20 limit would fire in every period.

There is a second problem. "Clusters `unresolved` by a Jev failure, a cap or the per-merge cluster cap enter neither rate" would also remove `fallback_cap` clusters, which contradicts "whether or not the fallback ran".

**Fix.** At `:116`, and the same in the proposal:
- Replace "(valid `new` answers ÷ clusters the fallback answered)" with "(valid `new` answers ÷ the residual clusters counted in the fallback rate, less those left `unresolved (fallback_cap)` or `unresolved (terra_failure)`)".
- Replace from "The pending-new rate is computed only over clusters the fallback answered;" to "…enter neither rate." with:

  > The pending-new rate counts valid `new` answers before the pending cap. It is `unmeasured` when the fallback answered no cluster in the window (for example, on hosted runners without `codex`). Clusters `unresolved (jev_failure)`, `unresolved (jev_cap)` or `unresolved (pr_cap)` enter neither rate. Clusters `unresolved (fallback_cap)` or `unresolved (terra_failure)` count in the fallback rate only.

With this definition, 14/44, the 16 % basis and axum's 25 per 100 are all consistent.

### R2-03 (medium): enabling the lane still requires a consent that the lane's skip rules say it does not need

`:114` opens with "`mapping.association.enabled` defaults to false and needs `decision-egress` and `model-cli` consent". The proposal's policy keys say the same. Yet the same paragraph runs Jev-only when the fallback consent is absent (`skipped (no consent)`). The two texts give different behaviour when only the Jev consent exists.

**Fix.** At `:114` replace that sentence with:

> `mapping.association.enabled` defaults to false. The `association` lane needs the `decision-egress` acknowledgement; `association-fallback` also needs its own model-CLI consent (below). Without the fallback consent, Jev association still runs and no fallback model is called.

In the proposal's policy keys, replace "needs `decision-egress` and `model-cli` consent" with "`association` needs `decision-egress`; `association-fallback` separately needs its own model-CLI consent".

### R2-04 (medium): §6 still says Jev output is an annotation row only

`:176`: "*Output* is an `annotation` row only (`issues` has no import of `decisions`); an attacker can at most lower a rank." The new §4 contract stores association decisions in `associations` rows, not annotations (`annotations` requires `issue_id`, `internal/state/annotations.go:20`). Untrusted input can now change a label or propose a pending area.

**Fix.** Replace `:176` with:

> *Output* of issue-triage questions is an `annotation` row only (`issues` has no import of `decisions`); an attacker can at most lower a rank. DRAFT revision 2.8: merge-time association decisions are `associations` rows (§4), not annotations. There, untrusted input can at most change an association label or propose a pending area, which only a maintainer accepts. It never changes issue state, the gate, issue identity or a signal series.

### R2-05 (medium): §6's threshold-version rule still applies to catalogue-only changes

R1-05 fixed the staleness of records. The thresholds themselves are a separate case. `:173` says "Any stored threshold … keyed to a different version is stale … once a threshold gates behaviour, ranking falls back to the deterministic baseline until re-measured or the maintainer accepts `provisional`", and ends "No launch threshold gates behaviour". The `merge.which_area` version key contains the catalogue (`question.go`; `questions/merge.which_area@v1.json` "the catalogue digest is part of the version key"). So every accepted area or manual map would make `t_choice` stale. The lane would then fall back to a ranking baseline that does not apply to association. `:164` itself concedes these thresholds gate behaviour.

**Fix.** After "…never re-decided automatically." at `:114`, insert:

> The association thresholds are policy values of this provisional lane, not tuned operating points keyed to a question version. Enabling the lane is the maintainer's acceptance of `provisional` under §6, so a `merge.which_area` version change caused only by the catalogue (its options) neither makes them stale nor stops the lane. Every association record is disclosed `thresholds provisional`.

At `:173` replace "No launch threshold gates behaviour (ranking is ordinal; audit rows 23, 25), so Jev ranking continues with the disclosure and the fallback has no subject until one does." with:

> No launch ranking threshold gates behaviour (ranking is ordinal; audit rows 23, 25), so Jev ranking continues with the disclosure and the ranking fallback has no subject until one does. DRAFT revision 2.8: the merge-time association thresholds (§4) gate label routing, not ranking or the gate. Enabling that opt-in lane is the maintainer's acceptance of `provisional`, and a catalogue-only version change does not make them stale (§4).

Mirror this in the proposal's policy keys.

### R2-06 (medium): the exported rows cannot rebuild the guard's window or its reset

`:114` says the window and the `catalogue-stale` flag "are computed from these exported rows and the exported `LastAgentRefreshAt`; no other guard state is stored". The row fields have no run id, no decision time and no refresh marker. `LastAgentRefreshAt` is a wall-clock time (`internal/state/mapping_refresh.go:18-20`). After a restore, decisions from before a refresh cannot be told apart from decisions after it, so ":116 … until a manual agent map completes … after which the window restarts" cannot be implemented.

**Fix.** At `:114` (and proposal step 6) add "run id, decision time, and the `LastAgentRefreshAt` in force when decided (null before onboarding)" to the row fields. After "no other guard state is stored", insert:

> The window uses only rows whose recorded `LastAgentRefreshAt` equals the current one, in decision-time order. The flag is set by the first window in that generation that exceeds a limit, and holds until `LastAgentRefreshAt` changes. Commit timestamps are never used.

Add to the acceptance test at `:378` that already carries the associations round-trip: "after export and restore the guard reproduces the same window and flag; a completed manual map resets both".

### R2-07 (medium): Appendix 22(a) still names schema 12, and older exports are unhandled

Five texts are out of step:
- Appendix 22(a) (`:499`) is authoritative and still says "one current schema (`user_version` 12)".
- Appendix 23 ends "Items 4, 19, 20 and 21 are otherwise unchanged". It omits item 22, although 22(a) changes and 22(c) gets an exception.
- The proposal (step 6) says "under Appendix 22(a), unchanged".
- No text says how an export without `associations.jsonl` imports. That is every export written before revision 2.8, and it is exactly what the re-baseline remedy restores from. The manifest's `absent` list handles only areas and signals today (`internal/state/export.go` `absentFiles`).
- A13 says the refused stores were "written by releases v0.1.0–v0.1.5", but the dogfood caller now pins v0.1.6 (`docs/STATUS.md` §1).

**Fix.**
- Append to Appendix 22(a): "**DRAFT amendment (item 23(e), pending council sign-off):** the current schema becomes `user_version` 13 for `associations`; 12 joins the refused versions; the v1 re-anchor stays the only migration and the re-baseline remedy is unchanged."
- Add Appendix 23 "(e) the current state schema becomes `user_version` 13 (item 22(a) amended; its rule otherwise unchanged)", and change its last sentence to "Items 4, 19, 20, 21 and 22 are otherwise unchanged."
- In proposal step 6 replace "under Appendix 22(a), unchanged" with "under Appendix 22(a), whose schema number becomes 13 (Appendix 23(e)) and whose refusal rule is unchanged".
- At `:114` after "added to the §7.2 export list", insert: "An export with no `associations.jsonl` (every export before revision 2.8) imports as no association records, and the manifest lists the file as `absent` like `areas.jsonl`."
- In A13 replace "(written by releases v0.1.0–v0.1.5)" with "(written by published releases, at least v0.1.0–v0.1.5, R-046)".

### R2-08 (medium): the user's confirmation has no ledger row, R-060 is not cited, and :19 overstates the confirmation

The post-axum confirmation now justifies the design rule 3 exception. It appears only as "(user confirmation 2026-10-31 after R-058)" at `:19`, `:500(c)` and proposal U1, with no ledger ID. That breaks ledger rules 1 and 3. R-060's corrected reading is used at `:115`, `:427` and `:463`, but R-060 is never cited. `:19` also reads as if the user confirmed the evidence findings: "the user confirmed … that it did not replicate fastapi and that its interval overlaps B2's, that the exception … stand".

**Fix.**
- Append to `docs/research-ledger.md`:

  > | R-061 | User confirmation after the axum result (R-058) was reported: shown Jev+Terra P_assign 0.622 [0.514, 0.719] (fastapi not replicated; interval overlaps B2's) and options (a) provisional opt-in with the axum limit stated, (b) improve detection first, (c) stop; chose (a) | This project's session, 2026-10-31; `docs/design-reviews/2026-10-31-jev-merge-association/proposal.md` "U1 (answered)" | user | 2026-10-31 | Scout design rev 2.8 (DRAFT) revision paragraph, Appendix 23(c) | The design rule 3 exception and the provisional opt-in of R-059 stand after axum | Covers the provisional opt-in only; default-on needs a signed-off X11 |

- At `:19` replace from "the user confirmed on 2026-10-31, after seeing" to the end of the paragraph with: "after being shown that result (axum P_assign 0.622 [0.514, 0.719], not replicating fastapi, interval overlapping B2's) and the options (a) provisional opt-in with the axum limit stated, (b) improve detection first, (c) stop, the user chose (a) on 2026-10-31: the exception and the provisional opt-in stand (R-061)."
- In the source lists at `:19`, at the start of the `:114` bullet and at proposal :3, change "R-055, R-056, R-057, R-058, R-059" to "R-055, R-056, R-057, R-058, R-059, R-060, R-061".
- At `:500(c)` and proposal U1, replace "user confirmation 2026-10-31 after R-058" with "R-061".

### R2-09 (low): "about 2.2 clusters per merged PR" misstates the evidence

At `:114` and in the proposal (policy keys, Q7), the 41 PRs are only the PRs that left residual clusters; most merges leave none (axum: 120 clusters from 69 of 905 first-parent commits, `FROZEN-AXUM.md`).

**Fix.** Replace "fastapi averaged about 2.2 clusters per merged PR" with "fastapi averaged about 2.2 clusters per merged PR that left residual files (92 in 41 such PRs; most merges leave none: axum 120 clusters from 69 of 905 first-parent commits)".

### R2-10 (low): "Jev rarely answers `none`" cites the combined arm

At `:115`, "and Jev rarely answers `none` (Jev + fallback new-area recall 0.500)": the 0.500 includes the fallback's answers.

**Fix.** Replace that phrase with "and Jev rarely answers `none` (Jev-only new-area recall 0.217, 10/46; the fallback raises it to 0.500, 23/46, but never sees clusters Jev assigns)". Make the same change in the proposal's "Evidence and its limits".

## Considered and not raised

- HD8 does not say that recall counts abstention as a miss. X11 does, and HD8 names X11 as its test. Too small to raise.
- The challenger's alternative for R2-05 was rejected: treating the thresholds as stale and requiring renewed maintainer acceptance after every catalogue change would stop the lane after every accepted area or map. X11's closed-loop replay measures catalogue-change effects.
- R-060 marks itself as superseding R-058's cause clause and leaves R-058 unedited, so the append-only rule holds.
- No settled decision is reopened. Appendix 1–21 shows no new contradiction, apart from item 22 (R2-07).

## Where the peers changed the lead's view

- **Challenger:**
  - R2-04 (§6 output contract), R2-05 (§6 threshold-version rule) and R2-06 (guard rebuild) were not in the lead's pass.
  - On R2-02 it proposed restating the figures as 0.70. The lead kept the 16 % basis instead, because that version of the limit would fire in every period.
- **Experiment peer:**
  - R2-01 (`owned_classify`, rule (f), exclusions, pool), R2-08's wording overstatement, R2-09 and R2-10 came from this peer.
  - It supplied the R2-02 replacement text, which matches the lead's.
- **Lead:**
  - found R2-02 independently;
  - raised R2-03 (also found by the challenger);
  - raised R2-07 (also found by the challenger), adding the older-export import case and the A13 release range;
  - raised R2-08's missing ledger row (also found by the peer).
- R2-01 and R2-02 come from fix wording the council wrote in round 1.

## Open questions that need the user

None. U1 is answered (R-061 to be appended). Every remaining issue has a text fix that follows existing design rules or the measured setting.

## Verdict

**10 ISSUES** (high: R2-01, R2-02; medium: R2-03, R2-04, R2-05, R2-06, R2-07, R2-08; low: R2-09, R2-10). No sign-off in this round. A later sign-off must name the SHA-256 digests of the revised `docs/scout-design.md` and `proposal.md`, and holds only while those digests are unchanged. Round 3 checks only the lines these fixes change, plus R-061.
