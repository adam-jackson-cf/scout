# Round 15: ADR-012 (shadow-window readiness inputs), formatting-only re-sign

Date: 2026-11-08. This round applies design rules 1–3 in `CLAUDE.md` and the ADR rules in `docs/adrs/README.md`. ADR-012 is an agent decision made under the owner's standing delegation (R-064).

Round 14 (`round-14.md`) signed 13 files with zero issues. When Main tried to commit them, the pre-commit hook's Prettier check failed on two signed files, so nothing was committed. The repository's pinned Prettier then rewrote those two files, and nothing else changed them. Because signed bytes changed, round 14's limits require a new zero-issue round. This round has a narrow scope:

1. Check whether the formatting change alters meaning in either file: wording, list or table structure, table cells, column count, alignment markers, links and anchors, code spans, or the line numbers that round records cite.
2. Check that the new signed set is complete and that round 14's sign-off conditions otherwise still hold.
3. Raise any new issue as ADR012-R15-xx.

The ADR-012 design is not reopened.

The council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial check of every changed row pair, character by character. It covered dashes, `≥`, `§`, curly quotes, `[bot]`, escapes, backticks and link targets, plus column counts, delimiter rows, hunk line counts and GitHub-flavoured Markdown (GFM) table parsing. It also searched all 13 signed files for references to a numbered round or to an old digest.
- **`experiment-peer` (Opus):** tested eight competing hypotheses against the diff, the worktree files, base bytes in `scout-adr11b`, the hook source, `round-14.md` and `round-14-addendum-response.md`. Examples: Prettier escaped or normalised characters, merged or split a row, the staged file differs from the diff, a third file changed, or the addendum misstates its scope.

The review was read-only. Nobody edited files, ran product code, made model calls outside the council, or touched `.codegraph`.

- The lead read the required codegraph skill references (`readiness-and-bootstrap.md`, `query-selection.md`). The council did not use codegraph, because the material under review is two Markdown tables and the council had no shell. Searches used `read`, `grep` and `glob`.

**Main computed and verified the digests below with `shasum`. The council did not, because the lead and both peers had read-only tools and no shell.** All paths are under `docs/`.

| File | SHA-256 (Main) | Since round 14 |
|---|---|---|
| `adrs/2026-11-08-shadow-readiness-inputs/ADR.md` | `a2869a076e0820cc13552ac9055cee22d0a57effb7d770c4501c9fb1c20343d8` | unchanged |
| `scout-design.html` | `54d769cbcaec081a6764ef35f2898a91a3f839537ab7ea79e978763de5d7f587` | unchanged |
| `adrs/README.md` | `cde415ff10b659a073d010f8a48c281d6f75a5a1cba466962d589e937b94587f` | unchanged |
| `STATUS.md` | `17e6e4bf7f131111897600dcdbd17e5d51c13eace4f38b65cf7622357d3d563c` | unchanged |
| `plans/2026-11-06-remaining-work.md` | `d78075702e3ef1f260b138572c2b5e454eed5b38327ccd2a1fa3699dacded88a` | **changed** (was `33afb05b4bd721e381475368e8354bc8e71aa24656e996a56819190f8154fe35`) |
| `references/policy-file.md` | `cdea3594df72661f64610f94c7ac49f07797749dc5bb81dd198c405f7c86d3b8` | **changed** (was `7b747b02563fb3fdfa41914b99b9a7fa25a12546ad23fe98e922b99cccc80727`) |
| `references/ci-integration.md` | `057594820d494584c004ba8d02a600fe1384196dcafa1f669500127aa1cf4ec4` | unchanged |
| `references/cli-commands.md` | `a5ce50bab10d6c3348053e2321ce9e88499830780dd117ea0279e303b19d21e8` | unchanged |
| `research-ledger.md` | `1ff9e7158b5619b41d1bb8a060d2b8ab9d58080a9c9620ba0524092f65313f77` | unchanged |
| `adrs/2026-10-19-v1-decisions/ADR.md` (ADR-000) | `1cd521fd96439d87275ba353f7b05451c3e0c4d556ee665529aab85970ed81cb` | unchanged |
| `adrs/2026-10-23-decisions/ADR.md` (ADR-003) | `973d24c4cfdfdd98acb4070a7bd72ca0c0e1ddf303bd4ccec8f92fe1cf117cc5` | unchanged |
| `adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) | `ee593b6241814c24035671d666566bc2b7c399135867226c26e9626f3936abc8` | unchanged |
| `adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) | `80e3375f81656a8856cf662cbec99a36d8e00bc451ff3c3592925d2293f44d95` | unchanged |

The lead compared all 11 "unchanged" digests above, character by character, with the round-14 table (`round-14.md:24–37`) and the round-14 sign-off list. They match.

Main also computed the SHA-256 of one unsigned record:

| File | SHA-256 (Main) |
|---|---|
| `round-14-addendum-response.md` | `1aba88545a5b13505d349e897c04c74603cd56cdaaf5cc89418b8d8b6fb8f578` |

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11`, branch `agent/policy-shadow-adrs`, based on origin/main `7bb73e52`. Every change is staged and uncommitted.
- Base bytes come from `/Users/adamjackson/Projects/work/scout-adr11b`.

**Files read:**
- `local://adr012-round15-formatting.diff`, all lines, with the long lines 24–32 read raw. It has three hunks:
  - plan `-12,7 +12,7`;
  - plan `-52,14 +52,14`;
  - policy-file `-12,35 +12,35`.
- Worktree `docs/references/policy-file.md:1–53` and `docs/plans/2026-11-06-remaining-work.md:8–20,50–66`.
- `round-14.md`, all lines, and `round-14-addendum-response.md`, all lines.
- ADR-012 `ADR.md:52` (adopted-form route (a)) and `:67–68`.
- `docs/STATUS.md:7,9`.
- `scripts/guardrails/infrastructure.py:17–20,390–410`, which holds the hook's `prettier()` function and its exclusions.

**Searches:**
- `(?i)round[- ]1[45]` across the 11 signed files outside `STATUS.md`: no matches. STATUS was covered by a separate docs-wide search, which found "round 14" only in the unrelated E1b-v2 row (`STATUS.md:140`).
- `final round|round-14|round 14|round-15|round 15` across `docs/`, excluding this folder's round files.
- `33afb05b|7b747b02|4ac36ff6|8a421f61` (the old digests and old blob IDs) across the worktree. These match only in this folder's round records (`round-11.md` to `round-14.md`), and no signed file contains them.
- `final|zero-issue|Review outcome|adopted-form|route \(a\)` in ADR-012 `ADR.md`.

## Issue summary

**No issues are open.** No ADR012-R15-xx issue was raised.

## Is the change formatting only?

| Check | Result | Evidence |
|---|---|---|
| Cell text in every changed row | Identical after trimming padding. The plan has 8 changed rows: the "Product roadmap" row, the Track B header, delimiter, B1–B4 and B5. The policy file has 29: header, delimiter and 27 key rows. The challenger compared all 37 pairs; the lead and the experiment-peer checked independently. No en dash, `≥`, `§`, `—`, `$0.50`, `[0, 1]`, `(0, 1]`, `<id>`, `[bot]`, `current_window` or curly quote changed. There is no backslash, so nothing was escaped | diff `:6–7,24–32,48–107`; worktree policy `:16–44`; plan `:15,55–62` |
| Rows and columns | No row was merged or split. Each hunk keeps its line count (7→7, 14→14, 35→35). The roadmap summary table keeps 3 columns, and the Track B and Keys tables keep 4 | hunk headers |
| Delimiter and alignment rows | Only the number of hyphens changed. Neither side has colons, so alignment is unchanged. The plan's first column widened from `---` to `----` to fit `B4-S` | diff `:25`, `:49` (raw) |
| Links, anchors, code spans | The B4-S row, which holds `../adrs/2026-11-08-shadow-readiness-inputs/ADR.md#implementation-acceptance-invariants`, is an unchanged context line. No other changed row contains a link. All backticked spans are unchanged | diff B4-S context line; plan `:61` |
| Line-number citations | Because no line was added or removed, round 14's plan `:61` (now read as the B4-S row) and older rounds' policy-file citations still point to the same content | plan `:55–62`; policy table `:14–44` |
| GFM rendering | The tables render the same. Both versions already had valid delimiter rows and matching cell counts, and padding cannot create or remove a cell [INFERENCE: not rendered] | the challenger's analysis |
| Staged bytes equal the diff's `+` side | The worktree reads match the `+` and context lines. That the index equals the worktree rests on Main's report that every change is staged [INFERENCE] | worktree reads |
| Why only these two files failed | The hook's `prettier()` checks only `.md`/`.yaml`/`.yml` paths. It skips `GOVERNED` files (the design html, the research ledger and others) and anything under `docs/adrs/`. Of the 13 signed files, it therefore checks only STATUS, the plan, policy-file, ci-integration and cli-commands. In base, the edited rows were padded. ADR-012 rewrote the policy `blocking.checks[].evidence` row (digest unchanged since round 11), and added B4-S and edited B5 and the roadmap row in the plan, all without padding. Those edits stayed uncommitted until the round-14 commit attempt [INFERENCE: no earlier commit attempt] | `infrastructure.py:17–20,390–410`; the experiment-peer's base comparison |

Main reports two further checks that the council did not repeat: the old and new bytes are identical after normalising whitespace, table pipes and rule rows, and the staged bytes equal Main's own pinned-Prettier output. They are consistent with every observation above.

## Do round 14's conditions still hold?

| Condition | Holds? | Evidence |
|---|---|---|
| The signed set is complete and has the same paths | Yes. The 13 paths match round 14 path for path. They are the set that `ADR.md:52` route (a) lists: the summary, the design, README, STATUS, the plan, the ledger, the policy-file, ci-integration and cli-commands references, and ADR-000/003/007/008 | `ADR.md:52`; `round-14.md:24–37` and its sign-off list |
| No signed byte names round 14 as the final round | Yes. `ADR.md:52` says "This folder's explicit zero-issue final round names outcome and SHA-256", and `:68` says "final zero-issue round names signed digests". `STATUS.md:9` says ADR-012 "takes effect only on its own final zero-issue round". None of them names a round number, so a round 15 does not make any signed byte stale. No signed file contains the old digests | searches above |
| The other 11 files are unchanged | Yes, on Main's digests [INFERENCE: not computed by the council]. They are identical to round 14's | digest table |
| "Any change to these bytes needs a new zero-issue round" | Met by this round. Round 14's execution step 2 (do not modify the files) was broken only by the hook-required reformat, which this round reviews | `round-14.md`, "Limits of the sign-off" |
| Adoption is recorded outside the signed files | Unchanged | `ADR.md:52` |
| Historical records stay byte-identical | Yes. The round records are not signed and the formatting change did not touch them. The addendum response lives under `docs/adrs/`, which the hook's Prettier check skips | `infrastructure.py:407` |
| Round 14's design closures (R1–R14) | Not affected. Neither changed file carries ADR-012 rule or fixture wording outside the unchanged B4-S row, and the policy-file evidence row's text is identical | the formatting checks above |
| Addendum response accuracy | Accurate. It names the same two files, the pinned Prettier, no bypassed hook and no claimed renewed sign-off, and it correctly calls the normalised comparison Main's evidence | `round-14-addendum-response.md` |

## Considered and not raised

- **The council cannot recompute the SHA-256 digests or compare the index with the worktree.** This limit is the same as in rounds 11–14, and Main's `shasum` is the stated source. It is not a defect.
- **Comparing the long B1–B5 Notes cells by eye rather than mechanically.** Three independent readings agree, and Main's normalised byte comparison covers the same ground. Not raised.
- **Alternative of a two-file supplementary sign-off that incorporates round 14's 11 signatures** (the challenger). A full 13-digest statement is more self-contained and is what `ADR.md:52` route (a) expects of the final round. It is adopted below.

## Peers and where they changed the outcome

- **`innovation-challenger`** reported zero issues.
  - It confirmed 37 changed row pairs with identical cell text, unchanged column counts and alignment, the untouched B4-S link, equal hunk counts, and no stale round or digest reference in the 13 signed files. It also found that design `:1011` defers sign-off to the final round in the same generic way.
  - It recommended renewing the full 13-file sign-off.
  - It did not change the verdict.
- **`experiment-peer`** reported zero issues.
  - It falsified H1–H5 and H7 and found no sign of H6.
  - It explained H8 (why the policy file was flagged only now) from the hook's exclusions and base bytes.
  - It confirmed the 13-path set and the generic "final zero-issue round" wording.
  - Its hook-exclusion finding was checked by the lead (`infrastructure.py:390–410`) and added to the evidence above.
  - It did not change the verdict.
- **Lead-only:**
  - the character-by-character comparison of the 11 unchanged digests with round 14;
  - the round-number and old-digest searches across the signed set;
  - the check that `ADR.md` and STATUS wording is generic;
  - the severity calls.

## Verdict

**ADR-012: zero issues. Signed off.**

The round-15 change is formatting only. It changes table padding and the hyphen counts in delimiter rows in two signed files, and it does not change their meaning. Round 14's sign-off conditions otherwise hold, and the signed set is complete.

**Sign-off statement.** The innovation council signs off ADR-012 (shadow-window readiness inputs) with zero open issues. This round supersedes round 14 as this folder's final zero-issue round. The sign-off covers exactly these 13 signed files, at the SHA-256 digests that Main computed and verified with `shasum` (the council could not compute digests itself). All paths are under `docs/`.

- `adrs/2026-11-08-shadow-readiness-inputs/ADR.md` `a2869a076e0820cc13552ac9055cee22d0a57effb7d770c4501c9fb1c20343d8`
- `scout-design.html` `54d769cbcaec081a6764ef35f2898a91a3f839537ab7ea79e978763de5d7f587`
- `adrs/README.md` `cde415ff10b659a073d010f8a48c281d6f75a5a1cba466962d589e937b94587f`
- `STATUS.md` `17e6e4bf7f131111897600dcdbd17e5d51c13eace4f38b65cf7622357d3d563c`
- `plans/2026-11-06-remaining-work.md` `d78075702e3ef1f260b138572c2b5e454eed5b38327ccd2a1fa3699dacded88a`
- `references/policy-file.md` `cdea3594df72661f64610f94c7ac49f07797749dc5bb81dd198c405f7c86d3b8`
- `references/ci-integration.md` `057594820d494584c004ba8d02a600fe1384196dcafa1f669500127aa1cf4ec4`
- `references/cli-commands.md` `a5ce50bab10d6c3348053e2321ce9e88499830780dd117ea0279e303b19d21e8`
- `research-ledger.md` `1ff9e7158b5619b41d1bb8a060d2b8ab9d58080a9c9620ba0524092f65313f77`
- `adrs/2026-10-19-v1-decisions/ADR.md` (ADR-000) `1cd521fd96439d87275ba353f7b05451c3e0c4d556ee665529aab85970ed81cb`
- `adrs/2026-10-23-decisions/ADR.md` (ADR-003) `973d24c4cfdfdd98acb4070a7bd72ca0c0e1ddf303bd4ccec8f92fe1cf117cc5`
- `adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) `ee593b6241814c24035671d666566bc2b7c399135867226c26e9626f3936abc8`
- `adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) `80e3375f81656a8856cf662cbec99a36d8e00bc451ff3c3592925d2293f44d95`

The unsigned `round-14-addendum-response.md` (`1aba88545a5b13505d349e897c04c74603cd56cdaaf5cc89418b8d8b6fb8f578`) was reviewed as context.

**Limits of the sign-off:**
- It adopts the design only. It does not validate the implementation.
- Any change to these bytes, including any further formatting, needs a new zero-issue round.
- Adoption is recorded outside the signed files, as `ADR.md` (adopted-form route (a)) requires.
- Owner review under R-064 remains pending.

## Council memo

1. **Recommendation.** Sign off ADR-012 at the 13 digests above, with round 15 as the final round, and commit without changing any signed byte.
2. **Alternatives.**
   - **Recommended: renew the full 13-digest sign-off in this round.** It is self-contained and matches `ADR.md:52`.
   - **B: a two-file supplement that incorporates round 14's 11 signatures.** It is equivalent in substance but splits the signed set across two rounds. Not recommended.
   - **C: revert the two files to the round-14 bytes and bypass or exclude them from the hook.** This would break the repository gate's rules for no gain in meaning. Not recommended.
3. **Evidence.**
   - The diff: hunks `-12,7 +12,7`, `-52,14 +52,14` and `-12,35 +12,35`, with lines 24–32 read raw.
   - Worktree plan `:15,55–62` and policy `:14–44`.
   - `ADR.md:52,68`; `STATUS.md:9`; `infrastructure.py:17–20,390–410`.
   - `round-14.md`, the digest table and the sign-off list.
   - The searches listed under Scope.
4. **Risks.**
   - The digests and index equality rest on Main's `shasum` and staging [INFERENCE].
   - GFM rendering was inferred, not rendered.
   - If any hook or tool rewrites the signed files again before the commit lands, the digests go stale and another round is needed.
   - The decision is document-only, is reversible by a superseding ADR, and owner review under R-064 remains open.
5. **Execution outline.**
   1. Save this record as `round-15.md` in `docs/adrs/2026-11-08-shadow-readiness-inputs/`.
   2. Before committing, check with `shasum -a 256` that all 13 signed files match the digests above.
   3. Commit with the hook enabled and without bypassing it. If the hook rewrites any signed file, stop and request a new round.
   4. Record adoption outside the signed files. ADR-012 takes effect on this round.
   5. Carry forward round 14's execution step 4 (the implementation-PR guidance for fixture (4) and provider timestamps) unchanged.
