# Design review round 1: ADR-009, one design document with requirements and ADR appendices

Date: 2026-11-06. Rules: `AGENTS.md`/`CLAUDE.md` design rules 1–3. The innovation council (lead) ran the review with two independent peers: `innovation-challenger` (GPT, adversarial) and `experiment-peer` (Opus, evidence). The review was read-only. It used no shell, made no model calls outside the council, and did not touch `.codegraph`.

Reviewed files (SHA-256; the main agent computed these and the council computed none):

| File | SHA-256 |
|---|---|
| `docs/scout-design.html` | `469f8531818cde5f7109e5a8e2baf7e50b25402d392feb87eb36136a4850c287` |
| `docs/adrs/2026-11-06-design-restructure/proposal.md` | `72076e8df31c7be0f7852c532ff05ee59d795b66b14583f9ea9138c0ed3dd0d6` |
| `docs/adrs/2026-11-06-design-restructure/requirements-map.md` | `82a42e075e13d0b388e99b8f1f48fce26b9dfe88279461c7f54c5617532c2c30` |
| `docs/adrs/2026-11-06-design-restructure/design-section-map.json` | `3cc0225e4bb98d6e39fbb6e71d870937adb84093cd9d301a85c026fb204da974` |
| `AGENTS.md` | `552d212a5b16b60657238aa8b2d747129acb2016d0b0045311e6f16d2ce64780` |

## Scope

- **Under review:** worktree `/Users/adamjackson/Projects/work/scout-req`, branch `agent/requirements-path`.
- **Base:** `main` at `9a3d19f5` in `/Users/adamjackson/Projects/work/scout/`. This covers:
  - `docs/scout-design.md` rev 2.8, `d1fb1274…`;
  - `SCOUT_REBUILD_REQUIREMENTS.md`;
  - `docs/scout-allowlist.md`;
  - `docs/design-reviews/`;
  - `docs/STATUS.md`;
  - `CLAUDE.md`/`AGENTS.md`.
- **Checks made:**
  - lost, weakened or changed normative content, sampled by threshold, exit code, schema version, pin, consent and egress, gate attribution, association lane, and allow-list eligibility and promotion;
  - REQ-1…REQ-41 against the base requirements;
  - `ADR.md` summaries and Appendix B cards against the round records;
  - the Overview against Appendix A;
  - live references to deleted files;
  - rule changes and violations.
- **Not reopened:** decisions settled in ADR-000…ADR-008.

## Verification

| # | Check | Evidence | Result |
|---|---|---|---|
| 1 | Exit codes and outputs | REQ-24 (`html:591`) keeps `0/1/2/3`, precedence 3 > 2 > 1 > 0, the JSON `gate`/`health`/`would_block`/`ranking` fields, the ≤ 60-line summary and SARIF (base design :215) | preserved |
| 2 | Schema version | REQ-10 (`:609`): new stores at 13; v1 re-anchor from 3; versions 1, 2 and 4–12 refused after a read-only preflight. REQ-46/47 hold the remedy and fixture rules (base Appendix 22(a), 23(e)) | preserved |
| 3 | Shadow window and activation | REQ-22/78 keep ≥ 14 days AND ≥ 5 comparable-base PR runs, `unknown` ≤ 10 % [ASSUMPTION], 3/20 demotion [ASSUMPTION], the version bump returning a check to shadow, and the `blocking:` evidence shape including `pr_runs` (base :207, :209) | preserved, except R1-03 |
| 4 | Allow-list eligibility | REQ-70…REQ-77: every Go/Python/TS mechanism; `generic-api-key` report-only; no wildcards; `--ignore-gitleaks-allow --gitleaks-ignore-path /dev/null`; `trusted_owners` [ASSUMPTION]. REQ-79…REQ-83 hold the AL §2 table | preserved |
| 5 | Two-track promotion | REQ-84/85: Track A fixtures plus attribution fixtures; Track B retention; corpus precision as an audit instrument only, with no fixed 50 (AL §4, base :210) | preserved |
| 6 | Pins | REQ-57 (0.7.3, `734f81bc…`, `scout-review-sanitize-v3` `dc867503…e6ce3`, pnpm 11.25.0); Gitleaks 8.30.1 (REQ-63); jscpd 4.0.8 (REQ-105); Jev `typesafe/jev-1.14-20261014` via `alpha/decisions` (REQ-87); Claude Code ≥ 2.1.169 (REQ-59) | preserved |
| 7 | Model use, consent and egress | REQ-31/32/58/60/61; REQ-139/140/147/149 (association: Jev egress, separate fallback consent, codex/Terra/high only, never `claude`); REQ-97 (per-lane consent) | preserved |
| 8 | Trust boundary | REQ-25/161/162/163 match design §7.4: PR job `contents: read` with no secrets; the trusted job alone gets `contents: write`, `pull-requests: read` and the Jev secret; the release-tag scripts; exit 3 naming the minimum tag | preserved |
| 9 | Association numbers | 10 clusters per merge; catalogue 2–255; 60 lines and 5 samples; `t_choice` 0.535, `none` iff ≤ 0.40, covered 0.60; 20 % second permutation; caps 40 and 10; guard 20/20/0.40/0.20 [ASSUMPTION] (REQ-144–150, 159; base :114, :116) | preserved |
| 10 | REQ-1…REQ-41 meaning | Each compared with base `SCOUT_REBUILD_REQUIREMENTS.md:11-79`. Text that follows the design checks out against it: REQ-13 refresh-map dispatch (:109, Appendix 21), REQ-29 2,000 [ASSUMPTION] (:177), REQ-34 [ASSUMPTION] (:267), REQ-4 non-goals 5–7 (:35) | kept, except REQ-35 (R1-02) and REQ-37 (R1-10) |
| 11 | ADR digests and issue counts | The peer read every final round. Final-round digests and the sequences 12→0, 25→0, 16→0, 10→0, 17→0, 12→0, 16→0, 14→0 plus 3→0 match the `ADR.md` files and cards. The adoption commits `110388d`, `9640370` and `1d4e0cc` match base STATUS | match |
| 12 | Ledgers append-only | R-062 appended; R-001 and R-059–R-061 unchanged; the hypothesis ledger has one History line added (81 → 82 lines) and the HD8 row is unchanged | ok |
| 13 | Frozen records | All 8 moved folders hold the base file set plus `ADR.md`. No `docs/adrs`, `ADR-00x` or `REQ-1xx` text appears in any round, response, snapshot or diff. `evaluation/**` is untouched (peer spot check, no hashing) | ok, within that limit |
| 14 | Tooling paths | `.omp/gate.json:39-40`, `.omp/coverage-ledger.json:841-842`, `.omp/ledger.d/c.json:431-432`, `.prettierignore:4-5`, `scripts/guardrails/infrastructure.py:18,409` and `test_gate_contract.py:31-32,122` name the new paths | ok |
| 15 | Code and doc citations | No `design §`, `Appendix N`, `AL §` or `req §` remains in `internal/`, `cmd/`, `.github/`, `scripts/`, `toolchain/`, `docs/references`, `docs/plans`, README, CONTRIBUTING or STATUS | ok |
| 16 | Experiment rules in `AGENTS.md` | Rules 1–7 are word for word, including `evaluation/<experiment>/reviews/round-<N>.md` and the SHA-256 sign-off. Zero issues and E1b-v2 grandfathering are kept | ok |
| 17 | Overview disposition values | "accepted, false positive, won't fix" match `internal/domain/types.go:208-210` | ok |

## Issues

### R1-01 (major): the report profile loses "exit 2/3 as neutral check"

- **Location:** `docs/scout-design.html:588` (REQ-21); map row "Design §7.3 Profiles → REQ-21".
- **Evidence:**
  - Base design :205: "`report` (adoption default; findings never fail CI; **exit 2/3 as neutral check**)".
  - REQ-21: "Adoption defaults to `report`, which never fails on findings".
  - REQ-24 gives the numeric codes only.
  - No REQ or ADR carries the neutral-check conclusion.
- **Required fix:** add to REQ-21 (or REQ-24) that under `report`, exits 2/3 are surfaced as a neutral check, not a failure.

### R1-02 (major): REQ-35 weakens the causal-claim condition

- **Location:** `html:622` (REQ-35).
- **Evidence:**
  - Base :293: "Causal claims need a **controlled-continuation or within-scope event-study design** signed off the same way. Gate use is a further separate decision (AL). Repeated dashboard significance never upgrades a claim."
  - REQ-35: "causal/gate claims need separate signed-off decisions."
- **Required fix:** restore the required study designs for causal claims. Keep gate use as its own separate decision, and keep "repeated dashboard significance never upgrades a claim".

### R1-03 (minor): rare-security activation loses its fire-rate proxy

- **Location:** `html:589` (REQ-22).
- **Evidence:**
  - Base :208: "first-baseline `pre-existing` counts are the fire-rate proxy".
  - AL §3: "…in non-generated paths are the proxy for expected fire rate".
  - REQ-22 keeps only "rare secrets/action-ref checks with zero events may activate `provisional: true` on Track A alone".
- **Required fix:** add the proxy rule (first-baseline `pre-existing` counts in non-generated paths) to REQ-22.

### R1-04 (minor): base-gap attribution loses "one gate reason per family"

- **Location:** `html:612` (REQ-27) and `:587` (REQ-20).
- **Evidence:**
  - Base :204: "Such a finding is excluded from `would_block` and the gate, **with one gate reason per family**".
  - REQ-27 keeps the reclassification. REQ-20 keeps the exclusion, but neither keeps the reporting rule.
- **Required fix:** add the one-reason-per-uninspected-family rule to REQ-27.

### R1-05 (minor): over-cap PRs lose the `linkage: incomplete` classification

- **Location:** `html:647` (REQ-127).
- **Evidence:**
  - Base :247: "A PR **over either cap**, or whose file list or commit list the API returns incomplete…, is marked `linkage: incomplete`, counted and excluded".
  - REQ-127 marks incomplete linkage only "if file/commit lists are shorter than totals or commits absent".
- **Required fix:** include exceeding either cap (50 files, 100 commits) among the conditions that set `linkage: incomplete`.

### R1-06 (minor): the Jev backlog commitment and its timing are dropped

- **Location:** `html:581` (REQ-99); `requirements-map.md` ("Unplaced statements: none").
- **Evidence:**
  - Base :9, :184 and :496 (Appendix 19), and base REQ-29: "backlog B-1–B-4 records the follow-ups; B-3 runs the row-30 flip-rate harness **on the first Choice that gates anything**".
  - REQ-99 keeps the `merge.which_area` gate and per-Choice flip rates. It loses the B-1–B-4 pointer and the "first gating Choice" trigger.
- **Required fix:** add both to REQ-99, or add a drop row with its reason.

### R1-07 (major): ADR status and provenance claims are false or inconsistent

- **Location:** `html:949`, `html:729`, `docs/adrs/README.md:22`.
- **Evidence:**
  - The ADR-009 card reads `accepted`, while the README says `proposed` and STATUS:7 says "pending council sign-off".
  - The Appendix B heading says "Each one is council-signed". ADR-000 never had a council round: its `ADR.md:51-52` says "no round files exist… Signed digests: none".
- **Required fix:**
  - Make ADR-009's status the same in the card, README and STATUS. If the signed bytes are to read `accepted`, the adoption commit must update README and STATUS at the same time, as rounds 9–11 did for revision 2.8.
  - Reword `:729` so ADR-000 reads as user decisions from the v1 memo with no council round.

### R1-08 (minor): the ADR-009 folder has no `ADR.md`, and another file cites it

- **Location:** `docs/adrs/2026-11-06-design-restructure/` (holds only `proposal.md`, `requirements-map.md` and `design-section-map.json`); `docs/plans/2026-11-06-remaining-work.md:14`.
- **Evidence:**
  - New rule 2 (`AGENTS.md:39-43`) requires "an `ADR.md` summary (ID `ADR-NNN`, status, decision, …, review outcome and signed digests)".
  - The README says the cards "are generated from these summaries".
  - The plan links to `…/design-restructure/ADR.md`, which does not exist.
- **Required fix:** add ADR-009's `ADR.md`, include it in the final round's digest list, and fill in the review outcome and digests at adoption.

### R1-09 (major): design rule 1 narrows what needs council sign-off

- **Location:** `AGENTS.md:35-38`.
- **Evidence:**
  - Base: "Any change to Scout's design (`docs/scout-design.md` **and the documents it governs**)".
  - New: "(`docs/scout-design.html`: overview, Appendix A requirements, Appendix B ADRs)".
  - Requirements still draw force from other files:
    - REQ-99: "Follow Jev adapter practices recorded Met/Fixed in `docs/jev-adapter-audit.md`";
    - REQ-133 validates against `docs/evidence/study-export.md`;
    - the cards are generated from `ADR.md` files.
  - Under the new wording, editing an audit row's status changes REQ-99 with no council review.
- **Required fix:** restore "and the documents it governs", or list the governed documents explicitly, at least `docs/jev-adapter-audit.md` (its Met/Fixed status and Design-level rows), `docs/adrs/*/ADR.md`, and the future `docs/evidence/study-export.md`.

### R1-10 (major): "no requirement changes meaning" is false, and ADR-009's own decisions are not recorded

- **Location:** `proposal.md` (change-set item 3, "Not changed"); `html:962`; `html:949` header ("supersedes the document layout of ADR-000 to ADR-008").
- **Evidence:**
  - **REQ-37:** `requirements-map.md:5` itself says "REQ-37 is corrected… the legacy exact schema is not adopted". Base REQ-37 required `ib-subject-*/1`, a wall-clock cap and output sorted by (`unit_id`, `measure_id`). The new text, principles only, follows design §8.7 (:305). The change is legitimate but it is a change.
  - **REQ-35:** see R1-02.
  - **Decisions ADR-009 makes without recording them:**
    - it replaces Appendix 6 / §10.3, whose cutover planned `docs/scout-product-contract.md`, `docs/scout-adapter-contracts.md` and an Enaible inventory under `docs/evidence/` (:370, :483);
    - it ends Appendix 12, which kept the requirements file live until Phase 6 (:489);
    - it deletes `docs/enaible-findings.md` instead of moving it.
  - **REQ-42:** it settles base §1's "X1–X6, X9, X10" (:37) in favour of §13's list, which includes X11.
- **Required fix:**
  - Correct the "no change in meaning" claim in the proposal and the card.
  - List each of the above as an explicit ADR-009 decision on the card and in its `ADR.md`.
  - Mark ADR-000 item 6 and ADR-002's Appendix 12 as superseded by ADR-009.

### R1-11 (minor): supersession records are incomplete, and the "never edited" rule conflicts with status fields

- **Location:** `html:736`, the ADR-005 and ADR-007 card headers, their `ADR.md` status lines, and `AGENTS.md:43`.
- **Evidence:**
  - ADR-000's status reads "partly superseded by ADR-006", yet its own Changes list amendments by ADR-004, ADR-005 and ADR-008. `ADR-007/ADR.md:32-33` amends ADR-000. ADR-009 is missing (R1-10).
  - ADR-005's status omits ADR-007 (`ADR-005/ADR.md:41`: "ADR-007 removes `CLAWPATCH_PROVIDER`").
  - ADR-002's status omits ADR-005 (the ADR-005 card says "Narrows ADR-002's mapping-provider policy").
  - ADR-004's header omits "supersedes part of ADR-000" (Appendix 7 patch).
  - "An accepted ADR is never edited" means these status fields can never be updated.
- **Required fix:**
  - Complete the supersession relations in both directions.
  - Either allow status-line-only edits on supersession, or record supersession only in the superseding ADR and the README index.

### R1-12 (major): the Overview contradicts Appendix A

- **Location and evidence:**
  - **`html:450` popover:** "A finding is `introduced` only if… [the base's] scanner inspected the finding's family… Without such a base… nothing is classified as introduced". This drops the `guardrails.*` exemption that REQ-27 (`:612`) and base :204 keep. It also merges the whole-run case with the per-family case.
  - **`:442`:** "3 unavailable: bad input or missing tool". This conflicts with REQ-18 ("Missing optional tools never block") and REQ-16 (only required-lane health drives exit 2/3).
  - **`:324`:** the nightly trigger omits the onboarding agent-map attempts that REQ-13 requires.
  - **`:354`:** the LLM column says "Area proposals at onboarding". This omits manual refresh (REQ-13) and the merge-time fallback's pending areas (REQ-147, REQ-151).
  - **`:449`:** the activation callout omits REQ-22's `unknown` ≤ 10 % condition. The popover states it, but the callout reads as complete.
  - **`:454`:** "Four kinds of record" conflicts with REQ-8's three record types, and associations appear in no REQ-8 record list.
- **Required fix:** align each Overview statement with its REQ.

### R1-13 (minor): the document has no precedence rule

- **Location:** `docs/scout-design.html` as a whole; no "authoritative", "prevails" or "wins" anywhere.
- **Evidence:**
  - The base had explicit precedence: the requirements header said "where this file and the design disagree, the design wins", and the design Appendix was "authoritative".
  - The single file now holds a simplified Overview, normative Appendix A and summarising cards, with nothing saying which governs when they disagree (see R1-12).
- **Required fix:** state that Appendix A is the normative requirement set, the Overview is explanatory, and ADR cards summarise decisions, with Appendix A prevailing on conflict.

### R1-14 (minor): `AGENTS.md` cites REQ-42 for an experiment list REQ-42 does not hold

- **Location:** `AGENTS.md:5`; ledger rule 2.
- **Evidence:**
  - `AGENTS.md` says "(design REQ-42: X1–X6, X9, X10)".
  - REQ-42 (`html:529`) says "X1–X6, X9–X11".
  - Ledger rule 2 also says "X1–X6, X9, X10".
- **Required fix:** use one list in all three places (X1–X6, X9–X11), and record in the map that REQ-42 follows base §13.

### R1-15 (minor): `AGENTS.md` cites a STATUS section this change deletes

- **Location:** `AGENTS.md:70-71`.
- **Evidence:**
  - "(user decision 2026-10-23, `docs/STATUS.md` §6.9)".
  - Worktree STATUS has only §§1–5, because proposal item 8 removed §6. Base STATUS:186 held §6.9.
- **Required fix:** cite R-039 / ADR-003 (or REQ-100–101) instead.

### R1-16 (minor): new live text has dangling references to deleted documents

- **REQ-110 (`html:630`)** cites "IR §5". `IR` meant the deleted `docs/inertia-research.md`, and the alias is no longer defined.
- **REQ-111 (`:631`)** says "computed on the scope population under §8.2 eligibility". §8.2 no longer exists; the rule is now REQ-125.
- **New `ADR.md` files** cite deleted or removed content without a commit pin:
  - `ADR-000/ADR.md:5,15,32,53` ("now listed in design §10.3"; "carries the decisions forward as authoritative");
  - `ADR-001:39-40`, `ADR-002:5,39-40`, `ADR-003:5,53` (STATUS §2/§6.9), `ADR-004:34`, `ADR-005:40-41`, `ADR-006:40`, `ADR-007:38,56`.
- **Remaining-work plan:** `docs/plans/2026-11-06-remaining-work.md:90-93` lists "stale statements in governed text" about the design's status line, which no longer exists.
- **Required fix:**
  - Point REQ-110 and REQ-111 to current IDs and paths (REQ-125, and a resolvable source for IR §5 or drop the alias).
  - Pin historical citations in the `ADR.md` files to `@9a3d19f5`, and turn present-tense "now"/"authoritative" statements into past tense or point them to ADR-009 and the current REQs.
  - Remove or rewrite the stale plan lines.

### R1-17 (minor): the README path map does not resolve section-style citations in frozen records

- **Location:** `docs/adrs/README.md:24`.
- **Evidence:**
  - The README maps only `docs/design-reviews/` → `docs/adrs/`.
  - Frozen records and earlier ledger rows also cite `design §n`, `Appendix N`, `AL §n`, `req §n`, `docs/STATUS.md §6.x`, `docs/scout-design-brief.md` and `docs/enaible-findings.md`, all now deleted or removed.
- **Required fix:**
  - State that deleted files are read at base commit `9a3d19f5`.
  - Point to `design-section-map.json` for the section → REQ/ADR mapping.

## Answers to the proposal's questions

1. **Normative content lost, weakened or changed?** Yes:
   - lost or weakened: R1-01 to R1-06;
   - changed but not disclosed as a change: REQ-35 (R1-02), REQ-37 (R1-10), REQ-42 (R1-10/R1-14).

   The rest of the sample in Verification checks 1–10 is preserved.
2. **Do the ADR summaries match their records?**
   - Decisions, numbers, digests and issue counts: yes (check 11).
   - Supersession: no (R1-11).
   - Provenance: ADR-000 is wrongly called council-signed (R1-07).
3. **Does the Overview agree with Appendix A?** No (R1-12). There is also no precedence rule (R1-13).
4. **Are the rule, tooling and citation changes complete, with no live reference to a deleted file?**
   - Tooling and code: yes (checks 14–15).
   - Not complete in four places: `AGENTS.md` (R1-09, R1-14, R1-15), the new `ADR.md` files and REQ-110/111 (R1-16), the missing ADR-009 `ADR.md` (R1-08), and the README path map (R1-17).
5. **Do experiment rules, ledgers or frozen records break any rule?**
   - Ledgers, frozen records and experiment rules: no violation (checks 12, 13, 16).
   - Rule-level problems: the narrowed design rule 1 (R1-09) and the "never edited" conflict with status fields (R1-11).

## Considered and not raised

- **Cache-key reuse** (the challenger asked to split dropped row 12): REQ-9 (`reused` coverage, reused-from run), REQ-16 (`reused` health) and REQ-28 (deterministic cache keys) carry what the base design states. The base requirements appendix explicitly removed the zero-call contract.
- **Refresh-guard rate populations** (REQ-157/158): the base :116 says "both counted over…" for both rates. Two later base sentences say otherwise for the fallback rate: "the fallback-rate numerator counts… whether or not the fallback ran", and no-CLI/no-consent clusters "count in the fallback rate only". REQ-157/158 follow those two sentences, which is the coherent reading. This is a pre-existing base tension, not caused by ADR-009.
- **The REQ-37 content itself:** principles only is authoritative under base §8.7. Only its disclosure is raised (R1-10).
- **REQ-25 "PR/fork jobs… no secrets":** matches base §7.4 ("PR job: `contents: read`, no secrets").
- **The abbreviated patch digest `dc867503…e6ce3`:** copied from the authoritative source. Inventing a full digest would be worse.
- **A3 "JSONL history squash-safe [ASSUMPTION]" and `retention-days: 90`:** an assumption-register entry and an operational value, not a requirement. REQ-183 keeps assumption-marked values provisional.
- **Historical filenames in signed-digest inventories, `ADR-008/ADR.md`'s digests under the new path, and ADR-000's commit-pinned `docs/archive/` citation:** legitimate provenance.
- **No revision or digest status line in the HTML:** future rounds bind the file digest directly.
- **REQ-1 drops "lightweight, extensible" and REQ-2 drops "opt-in and minimal":** wording only.
- **Limits:** the council and peers had no shell. Nobody hashed the moved round files or verified `go build`/`go vet`. The 389 map rows and 312 citation keys were sampled, not counted one by one.

## Peers

- **`innovation-challenger`:** 9 issues.
  - Added R1-02 (causal design) and R1-05 (over-cap linkage), which the lead had not found, and the guardrails-exemption contradiction in R1-12.
  - Confirmed R1-01, R1-03 and R1-04, which the lead had found independently.
  - Its cache-key issue was not adopted (see above). Its "accepted before sign-off" issue was merged into R1-07, kept with the rounds 9–11 precedent that signed bytes may carry the adopted status.
- **`experiment-peer`:** 11 issues.
  - Added R1-09 (governance narrowing), the ADR-000 "council-signed" error (R1-07), R1-06 (backlog timing) and R1-17 (path map).
  - Supplied the digest and issue-count verification (check 11) and the ledger and frozen-record checks (checks 12–13).
- **Lead:** added R1-13 (no precedence rule), REQ-110/111 and the stale plan lines (R1-16), and the supersession gaps for ADR-002 and ADR-004 (R1-11).
- **Effect of the peers:** both moved the recommendation from "minor fixes" to "not signed off". R1-02 and R1-09 are major findings the lead had not reached.

## Verdict

**17 issues (6 major: R1-01, R1-02, R1-07, R1-09, R1-10, R1-12; 11 minor): NOT signed off.**

No sign-off applies to the digests above. The single-file direction is sound. Fix every issue, then:

- regenerate `requirements-map.md` (remove "Unplaced statements: none" until it is true);
- add ADR-009's `ADR.md`;
- submit a round 2 naming new digests for `docs/scout-design.html`, `proposal.md`, `requirements-map.md`, `design-section-map.json`, `AGENTS.md` and the ADR-009 `ADR.md`.

Round 2 re-checks every fix and anything it touches.
