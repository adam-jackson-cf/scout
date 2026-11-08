# Round 4: ADR-011 (Gitleaks policy `version` contract)

Date: 2026-11-08. This round applies design rules 1–3 and the ledger rules in `CLAUDE.md`, and the ADR rules in `docs/adrs/README.md`. ADR-011 is an agent decision made under the owner's standing delegation (R-064). Where a choice was needed, the council chose under that delegation.

The council lead ran this round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review. It checked each R3 fix against round 3's required text, compared the six tracked documents with base and reconciled them with the diff stat. It also checked consistency across the signed set, the reservation list against the concurrent worktrees, the code citations, and whether any remaining text in frozen files depends on ADR-012.
- **`experiment-peer` (Opus):** competing-readings review. For each R3 fix it named an observation that would show the fix was not done, then looked for it. It gave competing readings of every sentence changed since round 3, checked the rules and the round-3 submission conditions, and looked for duplicate append-only rows across X1, X11 and the ADR-012 draft.

The review was read-only. Nobody edited files, ran product code, called models outside the council, or touched `.codegraph`. The lead and both peers had no shell.

**Main computed and verified the digests below with `shasum`. The council did not.**

| File | SHA-256 (Main) |
|---|---|
| `docs/adrs/2026-11-08-gitleaks-policy-version/ADR.md` | `f1b740afe30c59347a17382f37817cad2eb40480b8b0ec97e9a8e524b963cd7c` |
| `docs/scout-design.html` | `c9b7651bc06bbc757be0e26f64796a66136da68ba836f2a1891599a623a7cc71` |
| `docs/adrs/README.md` | `f7dce2bd2770f0c495a2306abe3605bea594d1a7bb50a3918bf594648555425e` |
| `docs/references/policy-file.md` | `1dc872b97e90b220df0e0939ff82ba06070f1a367743eb68aca324a9fa4b96c9` |
| `docs/STATUS.md` | `cc7fdc81918bfd80aa6108bdd5e71422760da386461093ac1bd60cfc2fec9275` |
| `docs/plans/2026-11-06-remaining-work.md` | `94ebb5b147bed2a9e3d0e09433daf6d0ae089b6e10dd117346a199622c40f570` |
| `docs/research-ledger.md` | `f35a1750095daed9fad22a74d0b15487c4e02aa788d1cec283f3d294bbc75693` |

These are the same seven values the writer reports after its fixes (`history://ADR11Writer`, final turn). Three digests are unchanged from round 3: `adrs/README.md`, `references/policy-file.md` and the plan. The other four changed in the R3 fixes, and this round reread their changed lines in full.

**Reviewed records.** These are identified by digest only and are **not signed** (R3-04):

| Record | SHA-256 | Source |
|---|---|---|
| `round-1-response.md` | `b014fca2177bf0b23266a046156d11b08e3bce7a61a862fd811ff4185cfc6ecd` | Writer (`history://ADR11Writer`, final turn). It equals the value in round 2 and round 3. |
| `round-2-response.md` | `2970ddd7baf93d9a9adcc5b896443c18f68b3012885c6067fc991824d20ef062` | Writer (same turn). It equals the value in `round-3.md`. |
| `round-3-response.md` | `3ae19e5d0e97a15fbe55124531103d30ad8d104990004f4223f923d6c1b11dfb` | Main, and the writer (same turn) |

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11b`, branch `agent/gitleaks-version-adr`.
- The changes are uncommitted, on top of `origin/main` `d2479d54`. Main fetched and confirmed that `origin/main` has not moved.
- Base bytes: `/Users/adamjackson/Projects/work/scout-gap1`, whose docs match `d2479d54`.

**Submission items that round 3 required:**

| # | Condition | Status |
|---|---|---|
| 1 | `round-3-response.md` with one row per issue | **Met.** Rows at `round-3-response.md:7-10` |
| 2 | Main-computed digests of exactly the seven files in `ADR.md:36`, plus separately listed digests of the response records | **Met.** Main supplied the seven signed digests plus `round-3-response.md`. The writer's final turn lists all three response digests separately from the signed set, and they match the earlier records. |
| 3 | `git diff --stat` against the actual base | **Met.** Main reports 6 tracked files, **+38 −9**, against `d2479d54`, which has not moved. The one added and one deleted line compared with round 3 (+37 −8) are both `STATUS.md:143`, which was unchanged from base in round 3 (reconciliation below). |
| 4 | Condition C1 confirmed | **Met for X1.** Main is the sole merger, X1 will not merge before ADR-011, and X1 will then drop its duplicate R-064 and reuse the adopted row under its own council. ADR-012 rebases onto ADR-011 before its own sign-off. X11 is handled in condition C2. |

**Files read:**
- This folder: `ADR.md` (in full, with lines 17 and 41 read raw), `round-3.md` and `round-3-response.md`.
- Canonical documents, compared with base:
  - `scout-design.html`: REQ-22 `:589` and REQ-78 `:600`, both raw; the ADR-011 card `:991-1004`;
  - `STATUS.md:7,118,126,128,143`, raw, against `scout-gap1/docs/STATUS.md`;
  - `research-ledger.md:1-8,70-74`, raw, against base `:1-8`;
  - `README.md:12`.
- Precedent and citations:
  - ADR-009 `2026-11-06-design-restructure/round-13.md:207-212` (step 5);
  - ADR-010 `2026-11-08-d1-jev-minimums/ADR.md:42-43` and `round-3.md:1-60,150-200` (format of the final round and sign-off);
  - `docs/adrs/2026-10-23-decisions/PROPOSAL.md` (exists).
- Concurrent ledgers, rows R-064 and above: `scout-x1`, `scout-adr11`, `scout-x11`, `scout-gap2`, `scout-d1` and `scout` (main).
- Code citations in R-069, verified by both peers:
  - `catalog_security.go:217,262,281`;
  - `spec.go:199-226`;
  - `policy.go:84,293,367-375`;
  - guardrails `suppress.go:49`, `focused_tests.go:9`, `config.go:39`.
- `history://ADR11Writer` (final turn: digests).

## Issue summary

**No issues are open.** All four round-3 issues are closed, and no new issue was found.

## Closure of round-3 issues

| Issue | Status | Evidence |
|---|---|---|
| ADR011-R3-01 | **Closed** | `STATUS.md:143` ends: "Latest research entry **R-069** (rows are in append order)." `STATUS.md:128` reads: "`docs/research-ledger.md` (R-001…R-064 and R-069; R-065–R-068 and R-070–R-077 are pre-allocated to concurrent X1, ADR-012 and X11 work; next free R-078)". Both match `round-3.md` word for word. The reservations are correct: X1 holds R-065–R-068, R-071 and R-072 (`scout-x1/docs/research-ledger.md:73-78`); ADR-012 holds R-070 and R-077 (`scout-adr11/docs/research-ledger.md:72-73`); X11 holds R-073–R-076 (`scout-x11/docs/research-ledger.md:72-75`). No ledger in any of the seven worktrees has a row at R-078 or above, so "next free R-078" is true. The ledger introduction and the plan are unchanged, as required. |
| ADR011-R3-02 | **Closed** | `research-ledger.md:74`, read raw:<br>• **Title:** "Repository source audit: rule-pack activation and version binding".<br>• **Links:** exactly `catalog_security.go:217,262,281`, `spec.go:199-226`, `policy.go:84,293,367-375`, the three guardrails files, `policy-file.md` and `ADR-003 proposal:43`.<br>• **Informed:** "ADR-011; REQ-22/78".<br>• **Key point:** two sentences, ending "…hand-bumped @N constants without enforced logic binding."<br>• **Caveat:** the required text, word for word.<br>No readiness link or sentence remains. R-069 matches `ADR.md:29` and the card's Informed-by callout `:1000`. The readiness facts now live only in ADR-012's draft R-077 (`scout-adr11/docs/research-ledger.md:73`), whose `policy.go:384-424` citation covers `ShadowStats`/`ShadowReady`, not version binding (peer). The two rows do not overlap. |
| ADR011-R3-03 | **Closed** | **Card `:1001`, second item:** "Lands as its own canonical PR and depends on no later ADR; supersedes the earlier patch-based submission." This is word for word.<br>**`ADR.md:24`:** the required text word for word, inside the existing bullet with a bold label.<br>**`ADR.md:41`:** "Adopted-form route (a) (ADR-009 `round-13.md` step 5)", "ADR-011 is a separate PR and depends on no later ADR." and "The main agent's separate-PR direction (`round-2-response.md`) supersedes…".<br>The cited step resolves to `round-13.md:212` ("Mark ADR-009 adopted only through files that are not in the signed set"), the same citation ADR-010 uses (`ADR.md:43`). In ADR.md, "ADR-012" now appears only at `:35`, which round 3 allowed. No "ADR-012" appears in the design HTML. The raw reads show that the "…" at `ADR.md:17,41` was reader truncation: the bytes end "not guarantees delivered here." and "append order need not be numeric." |
| ADR011-R3-04 | **Closed** | `ADR.md:36` lists seven files. Main's signed inventory and `round-3-response.md:12-20` list the same seven. Responses are hashed separately as records (`round-3-response.md:22`). No response was added to the canonical declaration. |

Every ADR011-R1 and -R2 issue closed in round 3 stays closed. The only signed bytes changed since round 3 are the R3 fix lines in `ADR.md`, the card, STATUS and the ledger, all reread above.

## Diff stat reconciliation

The challenger compared each document with base by hand and reconciled the result with Main's stat. The lead spot-checked STATUS and the ledger.

| Document | Changed lines | Added/deleted |
|---|---|---:|
| `STATUS.md` | `:7,118,126,128,143` | +5/−5 |
| `adrs/README.md` | `:12,27` plus a blank separator | +3/−0 |
| `plans/2026-11-06-remaining-work.md` | `:87,96` | +2/−2 |
| `references/policy-file.md` | `:61-68` | +8/−0 |
| `research-ledger.md` | `:4,73,74` | +3/−0 |
| `scout-design.html` | `:589,600` and the card `:990-1004` | +17/−2 |
| **Total** | | **+38/−9** |

Base `STATUS.md:143` read "Latest research entry **R-063** (`docs/research-ledger.md`)". It is the only line that is newly different from base since round 3. The other R3 fixes rewrote lines that were already counted as changed. `ADR.md` is untracked and so outside the stat. No change outside ADR-011's scope was found.

## New issues

None.

## Checks that found no issue

- **Competing readings of the changed sentences (peer):**
  - `STATUS.md:143` is true whether "latest" means the last row appended or the highest ID.
  - `STATUS.md:128` names no owner for each range, so the fact that X1 holds R-071/R-072 inside "R-070–R-077" makes nothing false. It will drift, but it is an editable tracker.
  - "Depends on no later ADR" (`ADR.md:24,41`, card `:1001`) stays true if ADR-012 is rejected, renumbered or lands second.
  - "Supersedes the earlier patch-based submission" refers to a submission mechanism, as `ADR.md:48` explains, not to an ADR.
- **Design rule 2:**
  - `ADR.md` has its ID, status, decision, sources, changes and a Review section.
  - Because `ADR.md` is itself signed, it names the highest-N zero-issue round in this folder and contains no digests (`ADR.md:41`), as ADR-009 and ADR-010 do.
- **Ledger rules 1 and 3:**
  - Every fact in R-069 is used at `ADR.md:29`.
  - R-064 justifies the delegated status (`ADR.md:3,28`).
  - Both are cited in `ADR.md:5,28-29` and card `:992,1000`.
  - No product hypothesis is affected, and no experiment is cited (design rule 3).
- **Code citations** (both peers, against the worktree):
  - `catalog_security.go`: `:217` is the Semgrep pack key, with no binary segment; `:262` is the Gitleaks binary `8.30.1`; `:281` is the exact pack string.
  - `spec.go:221-226` checks the `sha256:` content fingerprint and never compares it with `Binary.Version`.
  - `policy.go`: `:293` requires a nonempty version, and `:374-375` compares exactly with `a.Version != ruleVersion`.
  - The guardrails constants are hand-maintained `@1`.
- **No duplicate rows from ADR-012:** ADR-012's draft has no R-064 or R-069 row (`scout-adr11/docs/research-ledger.md:72-73`). Once it rebases onto ADR-011, no duplicate results.
- **Owner-decided clauses:** REQ-22 `:589` and REQ-78 `:600` are byte-identical to the text round 3 closed.

## Considered and not raised

- **Peer F1: X11 carries its own R-064 row, so C1 does not cover it (minor, as raised).**
  - **Finding:** the lead found this independently in the ledger-allocation search. `scout-x11/docs/research-ledger.md:71` is an older R-064, with Informed "…; ADR-012 shadow readiness inputs" and the pre-R2-02 caveat. The peer adds that X11 and X1 both claim their row was copied byte-identically from ADR-011 (`scout-x11/evaluation/x11-merge-association/reviews/round-5-response.md:11`, `scout-x11/docs/adrs/2026-11-08-x11a-delegated-decisions/ADR.md:69`, `scout-x1/docs/adrs/2026-11-08-x1-delegated-decisions/ADR.md:53`), and that claim is false.
  - **Why it is not an issue:**
    - No signed byte is wrong.
    - Main is the sole merger, so merge order is in Main's control.
    - If X11 merged first, its R-064 row would conflict with ADR-011's ledger append. Resolving that conflict changes signed bytes, which voids this sign-off and requires renewed review (`ADR.md:41`, condition 1 below). The frozen false text that C1 guards against therefore cannot land without a new council round.
  - **Disposition:** recorded as sign-off condition C2.
  - The false "copied byte-identically" claims belong to the X1 and X11 councils and are passed on in the memo.
- **Peer F2: response digests only partly supplied (minor, as raised).** Declined. The writer's final turn (`history://ADR11Writer`) gives full SHA-256 values for all three responses, kept apart from the seven signed files. The values for `round-1-response.md` and `round-2-response.md` equal those in `round-2.md` and `round-3.md`, and Main independently supplied `round-3-response.md`. Responses are reviewed records, not signed files (R3-04). Their digests rest on writer attestation, of the same kind as the digests this round accepts from Main.
- **R-064's Informed column includes "delegated design decisions", which is not an ID (lead).** Not raised:
  - Earlier rows are also descriptive: R-061 "…revision paragraph, Appendix 23(c)" and R-062 "`AGENTS.md` design rule 2".
  - An append-only row cannot list ADRs that do not exist yet.
  - Round 2 settled R-064's content.
- **`STATUS.md:143` no longer cites `docs/research-ledger.md` (peer, nit).** Round 3 prescribed this exact text, and `:128` names the file. Declined.
- **"Route (a)" is not literally defined in ADR-009 `round-13.md` (peer, nit).** The label comes from ADR-010's rounds. The citation points to the substance (step 5, `:212`) and matches ADR-010 `ADR.md:43`. Declined.
- **`Binary.Version` has no backticks in R-069 (lead, nit).** It is cosmetic, and the row is append-only once adopted. Declined.
- **The sequencing notes in the trackers will go stale if ADR-012 changes:** `README.md:12`, `STATUS.md:118` and plan `:87` say "ADR-012 rebases…". Round 3 allowed editable trackers to keep them. Not reopened.

## Peers and where they changed the outcome

- **`innovation-challenger`:**
  - raised **0 issues**;
  - found R3-01…R3-04 met exactly, quoting the bytes;
  - reconciled +38 −9 line by line, verified every code citation, and confirmed no remaining frozen dependency on ADR-012 and no literal truncation.
- **`experiment-peer`:**
  - raised **2 minor items** (F1, F2) and 4 considered-not-raised notes;
  - found R3-01…R3-04 met against explicit falsifiers.
  - **F1** confirmed the lead's X11 finding and added the false "copied byte-identically" evidence. It became condition C2, not an issue (reasons above).
  - **F2** was resolved from `history://ADR11Writer`.
- **Lead:**
  - checked every R3 fix against the raw bytes, base STATUS and ledger, and the ADR-009 step-5 and ADR-010 precedent;
  - searched all seven worktree ledgers for IDs R-064 and above, and found X11's duplicate R-064 and that no row holds R-078;
  - confirmed that the writer's seven signed digests match Main's;
  - considered the R-064 Informed wording and did not raise it.
- **Changes to the lead's position:** the peer's evidence that X1 and X11 both falsely claim byte-identical copies made C2 an explicit, named condition of the sign-off, where the lead had planned only a note in the memo. Neither peer changed the verdict.

## Verdict

**Zero issues. Signed off.**

The innovation council signs off ADR-011 ("Keep the Gitleaks rule-pack activation version"), worktree `scout-adr11b`, branch `agent/gitleaks-version-adr`, with the uncommitted changes on base `origin/main` `d2479d54`, with zero open issues. This round is the folder's final round. The sign-off covers exactly these seven files at these SHA-256 digests, computed and verified by Main:

- `docs/adrs/2026-11-08-gitleaks-policy-version/ADR.md`: `f1b740afe30c59347a17382f37817cad2eb40480b8b0ec97e9a8e524b963cd7c`
- `docs/scout-design.html`: `c9b7651bc06bbc757be0e26f64796a66136da68ba836f2a1891599a623a7cc71`
- `docs/adrs/README.md`: `f7dce2bd2770f0c495a2306abe3605bea594d1a7bb50a3918bf594648555425e`
- `docs/references/policy-file.md`: `1dc872b97e90b220df0e0939ff82ba06070f1a367743eb68aca324a9fa4b96c9`
- `docs/STATUS.md`: `cc7fdc81918bfd80aa6108bdd5e71422760da386461093ac1bd60cfc2fec9275`
- `docs/plans/2026-11-06-remaining-work.md`: `94ebb5b147bed2a9e3d0e09433daf6d0ae089b6e10dd117346a199622c40f570`
- `docs/research-ledger.md`: `f35a1750095daed9fad22a74d0b15487c4e02aa788d1cec283f3d294bbc75693`

The response records `round-1-response.md` (`b014fca2…6ecd`), `round-2-response.md` (`2970ddd7…f062`) and `round-3-response.md` (`3ae19e5d…1dfb`) are reviewed records. They are not signed.

**Conditions of this sign-off:**

1. **The bytes are frozen.** Any change to these seven files voids this sign-off and needs a new zero-issue round. This includes the declined nits and any rebase-conflict resolution.
2. **Adoption happens outside the signed set** (route (a); ADR-009 `round-13.md` step 5). No signed file is edited to mark adoption.
3. **C1 (X1), as Main confirmed:** X1 does not merge before ADR-011. It then drops its own R-064 row, reuses the adopted row, and is reviewed by X1's council.
4. **C2 (X11):** Main, as sole merger, does not merge X11 before ADR-011. X11 then drops its own R-064 row (`scout-x11/docs/research-ledger.md:71`), reuses the adopted row, and is reviewed by X11's council. If X1, X11 or any other change to the seven files lands first, condition 1 applies.
5. **ADR-012** rebases onto ADR-011's adopted bytes before its own sign-off.
6. **The digests, the diff stat and the unmoved base rest on attestation** by Main and the writer. The council had no shell.

---

## Council memo

1. **Recommendation.** Adopt ADR-011 on this round:
   - Save this record as `docs/adrs/2026-11-08-gitleaks-policy-version/round-4.md`.
   - Re-verify the seven digests, commit, and merge before X1 and X11.
   - Record adoption only outside the signed set.

   The decision stands:
   - the full pack string is kept and exact equality is implemented;
   - content binding applies only to bundled fingerprinted packs;
   - the Gitleaks binary-segment invariant is required, and its enforcement is pending later code;
   - the Semgrep and guardrails gaps are disclosed.
2. **Alternatives.**
   - **(A) Sign off now with C1/C2 (chosen).** No signed byte is defective, and the merge order is in Main's control.
   - **(B) Hold the sign-off until X11 confirms in writing.** This adds a round but no protection. A merge out of order already forces a ledger conflict, changed signed bytes and re-review.
   - **(C) Short tool version plus a separate digest field.** Rejected in earlier rounds: it needs a migration and a new contract, and no benefit has been shown.
   - **(D) Edit the signed bytes for the declined nits.** Rejected: it voids this round for cosmetic gain.
3. **Evidence.**
   - **ADR:** `ADR.md:3,17,24,28-29,35-36,41`.
   - **Design:** `scout-design.html:589,600,992-1003`.
   - **STATUS:** `STATUS.md:118,128,143`, with base `scout-gap1/docs/STATUS.md:126,128,143`.
   - **Ledger:** `research-ledger.md:4,73,74`.
   - **Concurrent ledgers:** `scout-x1:72-78`, `scout-adr11:72-73`, `scout-x11:71-75`.
   - **Precedent:** ADR-009 `round-13.md:212`; ADR-010 `ADR.md:43` and `round-3.md:173-195`.
   - **Digests:** `history://ADR11Writer` (final turn).
   - **Peers:** `history://ADR11Council4.R4Challenger`, `history://ADR11Council4.R4Peer`.
4. **Risks.**
   - **Merge order:** if X1 or X11 lands first, the sign-off lapses (conditions 1, 3 and 4). This is reversible at the cost of one more round.
   - **Ledger drift:** the reservation list in `STATUS.md:128` drifts as concurrent work allocates IDs. STATUS is editable, and each PR that lands should update it.
   - **Pending enforcement:** until the later validation PR ships, a Gitleaks binary bump that keeps the old pack label would not reset activation [INFERENCE: only a code change can create this case; the current guards are the toolchain parity test and the compatibility review required for scanner bumps].
   - **False claims in sibling records:** X1 and X11 claim their R-064 rows are byte-identical copies of ADR-011's, and they are not.
   - **ADR-012's ledger:** its R-077 row was edited while ADR-012's review was running (`round-3-response.md:8`). ADR-012's council must review the new bytes.
   - **Reversibility:** the change is documentation only. A later ADR can amend or supersede it, and no behaviour or data migrates.
5. **Execution outline.**
   1. Save this record as `round-4.md` in the ADR-011 folder.
   2. Run `shasum -a 256` on the seven files and compare each with the verdict list.
   3. Run `git fetch` and confirm `origin/main` is still `d2479d54`. Run `git diff --stat d2479d54` and expect 6 tracked files, +38 −9, plus the untracked ADR folder.
   4. Commit the six documents and the ADR folder, including `round-4.md`, as one PR. Merge it before X1 and X11. Record adoption only outside the signed files.
   5. Tell the X1 and X11 integration owners and councils to drop their R-064 rows and reuse the adopted row, and that their "copied byte-identically" claims are false (C1, C2).
   6. Rebase ADR-012 onto the adopted bytes and give its council the updated R-077 before its own sign-off.
   7. Later implementation PR: spec-validation enforcement that the Gitleaks pack label segment equals `Binary.Version`, with a test, plus the short-version diagnostic (`ADR.md:17`).
