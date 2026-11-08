# Round 3: ADR-011 (Gitleaks policy `version` contract)

Date: 2026-11-08. This round applies design rules 1–3 and the ledger rules in `CLAUDE.md`, and the ADR rules in `docs/adrs/README.md`. ADR-011 is an agent decision made under the owner's standing delegation (R-064). Where a fix needs a choice, the council names the option it recommends under that delegation.

This is the first round on ADR-011 alone. After round 2, ADR-011 and ADR-012 were split into separate PRs, and the `ADOPTION.patch` route was dropped (`round-2-response.md:7`). ADR-012 is reviewed by its own council and is out of scope here, except where its draft shows that something in ADR-011's bytes would duplicate it or become stale.

The council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review. It checked the round-2 fixes against the canonical bytes, compared all six documents with base, verified the code citations, checked internal consistency, and tested the adoption route against a squash merge, X1 landing first and ADR-012 landing first.
- **`experiment-peer` (Opus):** competing readings of every normative sentence that was added, compliance with the rules and ledger, observations that would show a fix is not done, and what happens if ADR-012 is rejected or X1 lands first.

The review was read-only. Nobody edited files, ran product code, called models outside the council, or touched `.codegraph`. The lead and both peers had no shell.

**Main computed and verified the digests below with `shasum`. The council did not.** This round is not a sign-off, so the digests only identify the bytes reviewed.

| File (under `docs/`) | SHA-256 (Main) |
|---|---|
| `scout-design.html` | `46407617a9e91d241f9c7e26e9c4f4d157226eaec1731e6c9fba0d463f6410e2` |
| `adrs/README.md` | `f7dce2bd2770f0c495a2306abe3605bea594d1a7bb50a3918bf594648555425e` |
| `STATUS.md` | `f194470a82d6f05bf9b17867a9e2b38e5a08d5110c63cb5ef3be4feae2f0b46a` |
| `plans/2026-11-06-remaining-work.md` | `94ebb5b147bed2a9e3d0e09433daf6d0ae089b6e10dd117346a199622c40f570` |
| `references/policy-file.md` | `1dc872b97e90b220df0e0939ff82ba06070f1a367743eb68aca324a9fa4b96c9` |
| `research-ledger.md` | `7e4bec781effe14da670b04ed31c1c8e722887045bbdb83e9e9f64083884d641` |
| `adrs/2026-11-08-gitleaks-policy-version/ADR.md` | `51e46c0934d1db68b3c71eb35d8b8337d12861da5773e614e8673359e318ae66` |
| `adrs/2026-11-08-gitleaks-policy-version/round-2-response.md` | `2970ddd7baf93d9a9adcc5b896443c18f68b3012885c6067fc991824d20ef062` |

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11b`, branch `agent/gitleaks-version-adr`. The changes are uncommitted, on top of `origin/main` `d2479d54`.
- Base bytes: `/Users/adamjackson/Projects/work/scout-gap1` (`f6cba734`, whose docs match `d2479d54`).

**Submission items:**
- **Per-issue response:** `round-2-response.md`.
- **Digests:** the table above.
- **Diff stat:** the writer's `git diff --stat d2479d54` (`history://ADR11Writer.ADR11Canonical`) reports **6 tracked files, +37 −8**. These are the six canonical documents; the five new ADR-folder files are untracked and were hashed separately. The council did not re-run the diff stat. Its line-by-line comparison of the six documents with base matches that count (8 deletions: STATUS 4, plan 2, REQ-22/78 2).
- **Historical records:**
  - The writer's digest of `round-1-response.md` (`b014fca2…`) equals the one recorded in `round-2.md`, so that file is byte-identical to the original.
  - `round-1.md` and `round-2.md` were not checked byte for byte. Their sizes match the copies in `scout-adr11`. They are review records and not in the canonical set.

**Files read:**
- This folder's `ADR.md`, `round-1.md`, `round-1-response.md`, `round-2.md` and `round-2-response.md`.
- Canonical documents, each compared with base:
  - `scout-design.html` REQ-22 `:589`, REQ-78 `:600`, the ADR-011 card `:991-1004`, and the ADR-003/010 cards `:807,972`;
  - `README.md:1-35`;
  - `STATUS.md:7,17,113-128,143`;
  - plan `:84-99`;
  - `policy-file.md:19-24,58-68`;
  - `research-ledger.md:1-8,70-74`.
- Code:
  - `internal/analyzers/catalog_security.go:204-293`, `spec.go:189-233`;
  - `internal/policy/policy.go:84,293,367-381`;
  - `internal/guardrails/{suppress.go:49,focused_tests.go:9,config.go:39}`;
  - `internal/engine/gate.go:350-385`, `diff.go:68-75,338-372`;
  - `internal/domain/gate.go:68-78`, `internal/state/would_block.go:30-48`, `internal/baseline/merge.go:25-138`.
- Precedent: ADR-010 `ADR.md:3,42-49` and `round-3.md` (verdict and signed list).
- Concurrent work:
  - `scout-x1/docs/research-ledger.md:71-78`;
  - `scout-adr11/docs/research-ledger.md:71-73`;
  - `scout-adr11/docs/adrs/2026-11-08-shadow-readiness-inputs/round-2-response.md:24-48`.
- A repository-wide search for stale `ADR-000…ADR-01x` ranges, and for Gitleaks version-contract statements outside the signed set.

## Issue summary

**4 issues are open, all minor. None is blocking or major.**

| ID | Severity | Short title |
|---|---|---|
| ADR011-R3-01 | minor | STATUS ledger pointers are false or incomplete after R-064/R-069 are appended |
| ADR011-R3-02 | minor | R-069 records readiness facts that ADR-011 does not use, and ADR-012 records them again as R-077 |
| ADR011-R3-03 | minor | The frozen card and `ADR.md` make promises about ADR-012, which has no index entry, and use an uncited route label and actor name |
| ADR011-R3-04 | minor | The submitted signed set (8 files) disagrees with the signed set `ADR.md` declares (7 files) |

## Closure of earlier issues

| Issue | Status | Evidence / residual |
|---|---|---|
| ADR011-R1-01 | **Closed** | REQ-22 `:589` no longer says equality covers parser bumps. Implemented exact equality and content binding for bundled fingerprinted packs are kept apart from the pending Gitleaks binary-segment enforcement and the open Semgrep and guardrails gaps. The split appears in `ADR.md:15,17,21`, REQ-78 `:600`, the card `:995-997`, `policy-file.md:67`, `STATUS.md:118` and plan `:87`. The code supports every claim (Checks). |
| ADR011-R1-02 | **Closed** | `ADR.md:3` and the card `rel` `:992` carry no pending state. "Owner review pending" appears only in the editable trackers `README.md:27`, `STATUS.md:118` and plan `:87`. The confirmation and rejection routes are at `ADR.md:43`. |
| ADR012-R1-16 (shared) | **Closed for ADR-011** | `STATUS.md:7,126` say "effective on its final zero-issue council round", which stays true after adoption. |
| ADR011-R2-01 (shared) | **Closed by the superseding route**, subject to condition C1 | No `ADOPTION.*` files remain, and the signed bytes contain no `ef6f50e8` parent rule and no squash restriction. `ADR.md:24,41` and `README.md:12` define separate canonical PRs, and say changed signed bytes need renewed review. `research-ledger.md:4` gives the append-order note round 2 asked for. A squash merge is compatible, because what is signed is the resulting file bytes. The X1 merge order is now condition C1, and the ADR-012 wording is ADR011-R3-03. |
| ADR011-R2-02 (shared) | **Closed** | R-064's caveat (`research-ledger.md:73`) matches the required text word for word. R-069 (`:74`) has no "select…" clauses, no "manifest v2" and no frozen pending-owner state. R-069's over-broad scope is a separate problem: ADR011-R3-02. |
| ADR011-R2-03 | **Closed** | `scout-design.html:589`: the gap clauses end with a full stop, followed by "auto-demote that repo's check to report for that repo and record it". Every other ADR-000/003 clause matches base word for word (compared with `scout-gap1:589`). |
| ADR011-R2-04 (shared) | **Closed** | The hand-bumped `@N` gap for Scout-native guardrails appears next to the Semgrep gap in REQ-22, the card `:997`, `policy-file.md:67`, `STATUS.md:118`, plan `:87`, `ADR.md:17` and R-069. "Implemented" is limited to exact equality plus content binding for bundled fingerprinted packs. |
| ADR011-R2-05 | **Closed** | `STATUS.md:7,126` and plan `:96` say "ADR-000…ADR-011 (ADR-011 effective on its final zero-issue council round)". `README.md:12` sits outside the rules list `:7-10`. No stale range remains anywhere else in the repository outside the round records. |
| ADR012-R2-09 (shared in round 2) | **Not applicable** | It concerned ADR-012's canonical revision rolling back ADR-011's text. That revision is no longer in this PR. |

## Issues

### ADR011-R3-01: STATUS ledger pointers are false or incomplete after R-064/R-069 are appended (minor)

- **Location:**
  - `STATUS.md:143` ("Latest research entry **R-063**");
  - `STATUS.md:128` ("R-001…R-064 and pre-allocated R-069");
  - `research-ledger.md:4`;
  - plan `:92` ("record it under the next free research-ledger ID").
- **Problem:**
  - This revision appends R-064 and R-069, so `:143` is false in the signed bytes and contradicts `:128`. This is the same kind of defect as the false ADR ranges in ADR011-R2-05.
  - Once the revision lands, main jumps from R-064 to R-069, and nothing on main says who holds the gap. The ledger introduction says only that gaps do not imply missing rows.
  - The plan tells authors to use "the next free" ID. The lowest unused ID (R-065) and the next ID above the highest (R-070) are both already held:
    - X1 holds R-065–R-068, R-071 and R-072 (`scout-x1/docs/research-ledger.md:73-78`);
    - ADR-012 holds R-070 and R-077 (`scout-adr11/docs/research-ledger.md:72-73`);
    - X11 holds R-073–R-076 (`shadow-readiness-inputs/round-2-response.md:44`).
  - A wrong "next free" call has already happened once: `round-2.md:456` named R-073.
- **Required fix:**
  - `STATUS.md:143`: "Latest research entry **R-069** (rows are in append order)."
  - `STATUS.md:128`: "`docs/research-ledger.md` (R-001…R-064 and R-069; R-065–R-068 and R-070–R-077 are pre-allocated to concurrent X1, ADR-012 and X11 work; next free R-078)". STATUS is an editable tracker, so later work updates this list.
  - The ledger and the plan need no change.

### ADR011-R3-02: R-069 records readiness facts that ADR-011 does not use, and ADR-012 records them again as R-077 (minor)

- **Location:** `research-ledger.md:74`: the title "…rule-pack activation, event identity and missing readiness inputs"; the links to `diff.go`, `domain/gate.go`, `engine/gate.go`, `would_block.go`, `store.go`, `export.go`, `baseline/merge.go` and `main.go`; the "Informed" text "ADR-011; activation and readiness requirements"; and the key-point sentences from "Schema 13 has no durable comparison/run facts…" to the end of the manifest sentence. For comparison: `ADR.md:29` and the card `:1000`.
- **Problem:**
  - Ledger rule 1 records a source when it is first used to justify a decision. ADR-011 cites only the activation and version-binding facts (`ADR.md:29`). No decision in this PR uses the readiness facts.
  - "Activation and readiness requirements" is not a decision or hypothesis ID (`research-ledger.md:5`).
  - ADR-012's current draft records the same readiness audit again as R-077 (`scout-adr11/docs/research-ledger.md:73`). Once both land, two append-only rows would describe one audit, and the readiness half of R-069 would have shaped no decision.
  - If ADR-012 is redesigned, that half would need a correction row (R-060 precedent). It cannot be trimmed after adoption.
  - The readiness facts are accurate: the council checked `gate.go:357,382-384` (event ID = base, head, check, issue), `would_block.go:38-40` (upsert overwrites `rule_version`) and `merge.go` (copy local-wins, metadata rebuilt from local). The problem is scope, not truth.
- **Required fix:** before sign-off, reduce R-069 to ADR-011's source:
  - **Title:** "Repository source audit: rule-pack activation and version binding".
  - **Links:** `catalog_security.go:217,262,281`, `spec.go:199-226`, `policy.go:84,293,367-375`, the three guardrails constants, `policy-file.md` and `ADR-003 proposal:43`.
  - **Informed:** "ADR-011; REQ-22/78".
  - **Key point:** only the first two sentences, ending with "…hand-bumped @N constants without enforced logic binding."
  - **Caveat:** "Static repository evidence, not experiment validation or evidence that version-binding gaps have been fixed."
  - The readiness facts stay in ADR-012's own row (R-077). `ADR.md:29` and the card already match the reduced row.

### ADR011-R3-03: The frozen card and `ADR.md` make promises about ADR-012, which has no index entry, and use an uncited route label and actor name (minor)

- **Location:**
  - card Changes `scout-design.html:1001` ("ADR-012 rebases after it merges and receives its own council review");
  - `ADR.md:24` ("ADR-012 rebases after ADR-011 merges and receives separate council review");
  - `ADR.md:41` ("ADR-011 is a separate PR and lands first; ADR-012 rebases after that merge…", "Adopted-form route (a)", "Main's separate-PR direction");
  - for comparison: `README.md:12`, `STATUS.md:118`, plan `:87`.
- **Problem:**
  - After adoption, the card is governed design text and `ADR.md` is frozen. ADR-012 has no folder or index row on main, so a reader there cannot resolve the ID.
  - ADR-012 had 17 open issues in round 2. If it is rejected, folded into other work or renumbered, these sentences become permanently false. Changing the card would need a council-signed design change, and `ADR.md` cannot be changed.
  - ADR-011's decision does not depend on ADR-012, so the frozen record should not depend on it either.
  - Unlike ADR-010 `ADR.md:43`, "route (a)" is not tied to its source.
  - "Main" names an agent session in a permanent record without saying what it is.
  - The editable trackers (README, STATUS, plan) may keep the sequencing note.
- **Required fix:**
  - **Card `:1001`, second item:** "Lands as its own canonical PR and depends on no later ADR; supersedes the earlier patch-based submission."
  - **`ADR.md:24`:** "Separate canonical PR: ADR-011 lands with only its own canonical document changes and depends on no later ADR. This supersedes the earlier patch/hash split and its commit-preservation restriction; no patch application or merge-method restriction is required. A later revision touching the same files rebases onto ADR-011's adopted bytes and is reviewed on its own."
  - **`ADR.md:41`:**
    - Write "route (a) (ADR-009 `round-13.md` step 5)".
    - Replace the ADR-012 sentence with "ADR-011 is a separate PR and depends on no later ADR."
    - Write "The main agent's separate-PR direction (`round-2-response.md`) supersedes…".
  - `ADR.md:35` ("not unsigned ADR-012 selections") can stay: it states what the rows do not contain.

### ADR011-R3-04: The submitted signed set (8 files) disagrees with the signed set `ADR.md` declares (7 files) (minor)

- **Location:** `ADR.md:36` ("Canonical signed set: this summary, … `docs/research-ledger.md`", seven files) and `round-2-response.md:9` (the same seven), compared with the submitted signed set, which adds `round-2-response.md`.
- **Problem:**
  - A final round that signed eight files would name a different set from the one the signed `ADR.md` declares, and `round-2-response.md` would be signing a statement that excludes itself.
  - Readers could not tell which bytes are the canonical revision that takes effect.
  - ADR-010's final round treated every response as a review record, not a signed file (`d1-jev-minimums/round-3.md:38,173`).
- **Required fix (recommended option):**
  - The final round signs exactly the seven files `ADR.md:36` lists.
  - `round-2-response.md` and `round-3-response.md` are identified by digest as reviewed records, not signed.
  - No file edit is needed; only the round-4 submission inventory changes.
  - Rejected alternative: add the response files to `ADR.md:36`. It breaks precedent, gains nothing, and every later response would have to be added too.

## Checks that found no issue

- **Scope of the six documents.** Each was compared with base:
  - design: REQ-22/78 and the new card only;
  - README: the route paragraph and the ADR-011 row;
  - STATUS: the ranges, gap 3 and the ledger pointer;
  - plan: gap 3 and the Appendix B range;
  - policy reference: the new note;
  - ledger: the introduction line plus R-064 and R-069.

  No schema-14, `pr_run_events`, collector, `shadowReadiness`, R-070 or ADR-007/008 content appears. ADR-012's requirements (`:592,610-618,704-712`) and every existing ADR status line and card match base.
- **Code citations.**
  - `catalog_security.go`: `:217` is the Semgrep pack key, with no binary segment; `:262` is the Gitleaks `Binary.Version` `8.30.1`; `:281` is the exact pack string.
  - `spec.go:224-226` requires the version to fingerprint the bundled content. It never compares the version with `Binary.Version`.
  - `policy.go:293` requires a nonempty version, and `:375` compares it exactly with `!=`.
  - The guardrails constants are `…unscoped@1` at `suppress.go:49`, `…@1` at `focused_tests.go:9` and `@1` at `config.go:39`.
  - `runner.go:612` takes the finding's `RuleVersion` from the spec, and `toolchain_test.go:63` ties `Binary.Version` to the pinned toolchain (peer). So "binary pin" is unambiguous, and no claim goes beyond what the code does.
- **Behaviour claim.** Keeping today's string needs no migration and changes no behaviour, because the activation compares equal to the current pack string.
- **Owner-decided clauses.** REQ-22 keeps the 14-day and 5-run window, maintainer triage, the zero-`not warranted` rule, the 10 % bound, the repository scope, the trailing-20 auto-demotion and the provisional route for rare checks. REQ-78 keeps the activation record and the comparable-run clauses. "Clarifies" ADR-000 stands, as round 2 accepted, and no status-line change is needed.
- **Folder and ID rules.** ADR-011 follows ADR-010. The folder is named `<YYYY-MM-DD>-<topic>`. Because `ADR.md` is itself signed, it points to the highest-N zero-issue round and holds no digests (`ADR.md:41`), as in ADR-009/010. The copied historical rounds are disclosed at `ADR.md:48`.
- **R-064.** The quote is identical in `ADR.md:28` and ledger `:73`. The row format matches R-062/R-063, and the `#informed-by` anchor resolves.
- **Delegation.** Delegation is not owner review and waives no council or experiment rule. Confirmation is a new `user` row; rejection is a superseding ADR (`ADR.md:28,43`; `README.md:12`).
- **Design rule 3 and the hypothesis ledger.** No experiment is cited, and no product hypothesis is affected.

## Considered and not raised

- **"Gap 3" is undefined, and it is the second bullet in the plan (peer, nit).** The plan's list is headed "from `docs/STATUS.md` §1", where it is the third gap. Declined.
- **REQ-78 says "not the tool version" next to "the version segment must equal the binary version" (peer, nit).** The full string is not the tool version and contains it as a segment, so the two do not conflict. Hard-coding the current string matches REQ-63's existing normative pin of Gitleaks 8.30.1: any scanner bump already needs a compatibility review and a design change. Declined.
- **"Digest changes return the check to shadow" compresses the mechanism (peer, nit).** A content change without a new label fails spec validation, and the required new label is what resets activation. The text is not false. Declined.
- **ADR-012 sequencing in the README, STATUS and plan.** These are editable trackers and may keep it. Only the frozen text is raised (R3-03).
- **Squash merge.** Compatible, because the signed objects are the resulting file bytes, not commits.

## Condition C1: X1 merge order

This is a coordination condition, not a defect in ADR-011's bytes.

- X1's draft R-064 (`scout-x1/docs/research-ledger.md:72`) is not ADR-011's revised row: its "Informed" column names ADR-012, and its caveat is the pre-R2-02 text.
- **If X1 merges first:**
  - ADR-011's R-064 collides with X1's;
  - `research-ledger.md:4` and `ADR.md:41` must change, which means renewed review (already provided for at `ADR.md:41`);
  - the frozen "owner review pending" text lands permanently on main.
- **Condition:** X1 does not merge before ADR-011. It then drops its own R-064 row, reuses ADR-011's bytes, and is reviewed again under its own rules.

## Peers and where they changed the outcome

- **`innovation-challenger`** raised 2 minor findings: the signed-set mismatch (→ R3-04) and the stale "latest research entry R-063" at `STATUS.md:143` (→ R3-01). The lead had missed `:143`. The challenger confirmed every closure, the document scope, the code citations and squash compatibility.
- **`experiment-peer`** raised 4 minor findings and 4 nits, and set condition C1.
  - Adopted:
    - the signed-set mismatch (merged into R3-04);
    - R-069 scope (R3-02);
    - frozen ADR-012 promises (R3-03, with its "route (a)"/"Main" nit merged in);
    - the reserved-ID gap (merged into R3-01).
  - Declined: three nits (above).
- **Lead:** independently checked the code citations and the readiness facts in R-069, the base comparisons and the stale-range search, and found the ADR-012 dangling-reference concern. It had first placed the signed-set mismatch and the R-069 scope under "considered and not raised".
- **Changes to the lead's position:**
  - The peer's evidence that ADR-012's draft records the same audit as R-077 turned R-069's scope from acceptable shared context into a duplicate, append-only row (R3-02).
  - Both peers' independent findings on the signed set, together with ADR-010's precedent, made R3-04 an issue.
  - The challenger's `:143` finding and the peer's reservation evidence became R3-01.
  - The lead had judged the frozen ADR-012 wording a nit. The peer's rejection scenario, given ADR-012's 17 open issues in round 2, raised it to minor.

## Verdict

**ADR-011: 4 issues open (4 minor). Not signed off.** Open: ADR011-R3-01, -02, -03 and -04.

The digests above identify the reviewed bytes only. No sign-off is given for any file.

**Submission conditions for round 4:**
1. A `round-3-response.md` with one row per issue.
2. Main-computed SHA-256 digests of exactly the seven canonical files in `ADR.md:36` (R3-04), plus separately listed digests of the response records.
3. `git diff --stat` against the PR's actual base. If `origin/main` has moved, rebase, name the new base, and confirm that the six documents still contain ADR-011's changes only.
4. Condition C1 confirmed with X1.

---

## Council memo

1. **Recommendation.** Do not sign off this round. Make the four text fixes in one pass, then resubmit for round 4, which should be able to reach zero issues. The fixes:
   - `STATUS.md:128,143` (R3-01);
   - R-069 reduced to the activation facts (R3-02);
   - the ADR-012-free wording in card `:1001` and `ADR.md:24,41` (R3-03);
   - a seven-file signed set (R3-04).

   The decision itself is sound. The string is kept, exact equality is implemented, the binary-segment invariant is pending later code, and the Semgrep and guardrails gaps are disclosed. Every R1/R2 issue is closed.
2. **Alternatives.**
   - **(A) Fix all four now (recommended).** Small edits to four files, one more round, and every frozen record stands on its own.
   - **(B) Sign off now and fix later.** Not allowed (design rule 1 requires zero issues). R-069 is append-only and the card is governed text, so neither could be trimmed after adoption.
   - **(C) Sign the response files too, and add them to `ADR.md:36`.** Rejected: it breaks the ADR-010 precedent and forces every later response into the signed set.
   - **(D) Keep the ADR-012 sequencing in the frozen text.** It is true only if ADR-012 lands as numbered. Rejected.
   - **(E) X1 lands first.** ADR-011 then rebases, reconciles R-064 without a duplicate row, and is reviewed again. That costs a round and lets the old R-064 caveat onto main.
3. **Evidence.**
   - **ADR:** `ADR.md:3,9-17,21-24,28-29,35-36,41,43,48`.
   - **Design:** `scout-design.html:589,600,992-1003`, with base `scout-gap1:589,600`.
   - **Trackers:** `README.md:12,27`; `STATUS.md:7,118,126,128,143`; plan `:87,92,96`; `policy-file.md:58-67`.
   - **Ledger:** `research-ledger.md:4,73,74`.
   - **Code:**
     - `catalog_security.go:217,262,281`;
     - `spec.go:224-226`;
     - `policy.go:293,367-375`;
     - guardrails `suppress.go:49`, `focused_tests.go:9`, `config.go:39`;
     - `engine/gate.go:357,382-384`;
     - `would_block.go:38-40`;
     - `merge.go:25-138`.
   - **Concurrent work:** `scout-x1/docs/research-ledger.md:72-78`; `scout-adr11/docs/research-ledger.md:72-73`; `shadow-readiness-inputs/round-2-response.md:44`.
   - **Precedent:** ADR-010 `ADR.md:42-49` and `round-3.md:38,171-195`.
   - **Diff stat:** `history://ADR11Writer.ADR11Canonical`.
4. **Risks.**
   - **Concurrent ledger allocation.** R3-01's reservation list can drift as X1, X11 and ADR-012 change. STATUS stays editable and should be updated by each landing PR.
   - **Merge order.** If X1 or a docs change to any of the six files lands first, ADR-011 must rebase and be reviewed again (C1; `ADR.md:41`).
   - **Pending enforcement.** Until the later validation PR ships, a Gitleaks binary bump that keeps the old pack label would not reset activation [INFERENCE: only a code change can create this case, and the toolchain parity test plus council compatibility review are the current guards]. The gap is disclosed in every governed location.
   - **Attestation.** The digests and diff stat rest on Main and the writer. The council had no shell.
   - **Reversibility.** The change is documentation only. A later ADR can amend or supersede it, and no behaviour or data migrates.
5. **Execution outline.**
   1. Save this record as `docs/adrs/2026-11-08-gitleaks-policy-version/round-3.md`.
   2. Writer:
      - edit `STATUS.md:128,143` (R3-01);
      - replace the R-069 row as specified (R3-02);
      - edit card `:1001` and `ADR.md:24,41` (R3-03);
      - write `round-3-response.md`, stating that the signed set is the seven files in `ADR.md:36` (R3-04).
   3. Main:
      - compute SHA-256 for the seven canonical files and for the response records, listed separately;
      - run `git diff --stat` against the actual base, and confirm that every changed tracked file is canonical;
      - confirm C1 with X1.
   4. Submit round 4. On a zero-issue verdict, adopt with adoption recorded outside the signed set (route (a), ADR-009 `round-13.md` step 5). Then ADR-012 rebases onto the adopted bytes, and X1 reuses ADR-011's R-064.
