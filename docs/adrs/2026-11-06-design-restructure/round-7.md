# Round 7: ADR-009, one design document with requirements and ADR appendices

Date: 2026-11-06. This round follows design rules 1–3 in `AGENTS.md`/`CLAUDE.md`. The innovation council lead ran the review with two independent peers:
- `innovation-challenger` (GPT) reviewed normative content, numbers and the ADR column.
- `experiment-peer` (Opus) reviewed provenance, ADR records, rules, ledgers and consistency across documents.

The review was read-only. It made no model calls outside the council and did not touch `.codegraph`.

**The main agent recomputed the digests; the council did not.** The lead and both peers had read-only tools and no shell, so none of them hashed a file. Earlier rounds accepted a main-agent recomputation, so this round does too. The main agent ran these commands on the staged worktree and recorded the results:
- `shasum -a 256` on each file below;
- `test -L CLAUDE.md && readlink CLAUDE.md` → `AGENTS.md`;
- `go build ./... && go vet ./...` → OK;
- `git diff --cached --stat` → `245 files changed, 5921 insertions(+), 1781 deletions(-)`. That is round 6's 244 files plus the new `round-6.md`.

The main agent also states that only `docs/scout-design.html` (REQ-92) and `proposal.md` changed since round 6, plus the saved `round-6.md` record. The peer's file modification times agree: the html and the proposal are recent, and the other governed files are older.

| File | SHA-256 (main agent, `shasum -a 256`) | vs round 6 |
|---|---|---|
| `docs/scout-design.html` | `94253aceefb3e564fcc6bdd823e0a5fd8de22ab9c4556c12e15f6e4e8d8c965c` | changed (R6-01) |
| `docs/adrs/2026-11-06-design-restructure/proposal.md` | `8bb693a148fd16c43d1f6ebf0f5dbb9163b3b34ec9c7c38bca410ff526ebc914` | changed (R6-02, R6-03, "Round 6 fixes") |
| `…/requirements-map.md` | `8a46cea4f0281ad3b27c11e37e167ba658d03e98286cd0aac267035f3b48b3dd` | unchanged |
| `…/design-section-map.json` | `3cc0225e4bb98d6e39fbb6e71d870937adb84093cd9d301a85c026fb204da974` | unchanged |
| `…/ADR.md` (ADR-009) | `e0d2d7f12d9185527e5c48138cdf54aa738c5235aae1e11ccf1ceeb1a17e1bfc` | unchanged |
| `AGENTS.md` (`CLAUDE.md` symlink) | `aff517d34b71ac2ed189dcc9f4d3b03380e9b91e9b754a2bd98aceeaa4718511` | unchanged |
| `docs/adrs/README.md` | `7c90063cd3050dd05984990dade0553e47e4fc7d43051b9e9c92f241c2e307e0` | unchanged |
| `docs/jev-adapter-audit.md` | `30894dbf17bf9f815ceaadce2c3c9ae75d2cf0e31d4a91753985dbfe6818630d` | unchanged |
| ADR-000 `2026-10-19-v1-decisions/ADR.md` | `4d631ce92da11fc7c42467d2b18cc12902f4e193d92170a7e5e33a9a8547b36f` | unchanged |
| ADR-001 `2026-10-21-e1b-v2/ADR.md` | `e5db1e7659fb562bd3c89cb36483d789d0e669d3208c8f2a9e0617346d2be8fa` | unchanged |
| ADR-002 `2026-10-22-rescope/ADR.md` | `4adb8f70b90d2ca4deb941b56d161ccce424549f34c323b53cd14c4c80a9ce2b` | unchanged |
| ADR-003 `2026-10-23-decisions/ADR.md` | `f42a2ea88e1b02d9a41b558841f32c7393a02cce58f7ccdd0c401caa679fdfe7` | unchanged |
| ADR-004 `2026-10-23-clawpatch-patch-v3/ADR.md` | `cbaf321eed0957fa29c2ac7bca02b981f3acd8c8efc895e10bc7a1bf685e577a` | unchanged |
| ADR-005 `2026-10-23-codex-luna-mapping/ADR.md` | `ff8a06e034a20dc0242732a71743b7f99c1cc82d1a5293eab6d29a272e34e00c` | unchanged |
| ADR-006 `2026-10-24-area-refresh-fallback/ADR.md` | `371a870bd4302d5bf3a6f224679b6ac2ad321a0c3b1dd4aeb6e70ce4941e9a82` | unchanged |
| ADR-007 `2026-10-24-cutover-cleanup/ADR.md` | `de9fffdf63fc0ec61a98bdde7327617682d7b336b1199711537d48b26ee04721` | unchanged |
| ADR-008 `2026-10-31-jev-merge-association/ADR.md` | `76dbd3aec2503d447270a91bdfb9a9c44e35ca730c45199add44a34847c45ab0` | unchanged |

**The digest changes match what round 6 required.** Only the html and `proposal.md` changed. The other 15 digests match the `round-6.md` table byte for byte.

## Scope

- **Under review:** worktree `/Users/adamjackson/Projects/work/scout-req`, branch `agent/requirements-path`, with all changes staged. The review covered:
  - the "Round 6 fixes" (`proposal.md:102-108`) at every location they touched;
  - a check that the fixes introduced no new inconsistency;
  - a final broad pass over proposal questions 1–5.
- **Base:** `main` at `9a3d19f5` in `/Users/adamjackson/Projects/work/scout/`, including the deleted originals and the frozen `docs/design-reviews/` records.
- **Checks made:**
  - **R6-01:** REQ-92 (`html:574`) against:
    - base design `:178`;
    - REQ-93, REQ-97 and REQ-179 (`html:575,579,716`);
    - the Overview rank stage (`html:266-268`);
    - the origin of the paragraph (rescope `proposed.diff:113`);
    - map `:211`.
  - **R6-02 and R6-03:**
    - `proposal.md:26-27` against base `docs/STATUS.md:170` (§6) and `:186` (§6.9);
    - ADR-002 and ADR-003 `ADR.md:5`;
    - worktree `hypothesis-ledger.md:39`.
  - **"Round 6 fixes" section:** its counts, the REQ-92 text, and its "(base design §6)" claim. Base headings put §6 at `:158` and §7 at `:192`.
  - **Broad pass:**
    - **Ledger, coverage and identity:** base `:87-103` traced to `html:532-539,553,609,613-614`.
    - **Gate and exit contract:** base `:192-215` traced to REQ-16, 19–27, 49–51, 77, 78, 84, 85, 160 and 161, and the Overview `:389-450,494-505`.
    - **Allow-list:** all six eligible rows, base `scout-allowlist.md:19-24`, traced to REQ-70–77.
    - **REQ-1…REQ-41:** base requirements `:9-75` traced through map `:298-338`.
    - **Formulas and numeric bounds:** base `:229-271` traced to `html:620-651`.
    - **Other defaults, caps and experiment rows:** base `:109-120,177-184,293,320-329,343,435-444`.
    - **ADR records:** card relations, review counts and digests for ADR-006/007/008 against their base round files.
    - **Live tree:** a grep for deleted paths and legacy citation forms, and the tooling files named in proposal item 6.
    - **Rules and ledgers:** experiment rules 1–7, and both ledgers checked as append-only.
    - **Counts:** 312 JSON keys, and 388/323/43/22 map rows.
- **Not reopened:** decisions settled in ADR-000…ADR-008, and rounds 1–6 "Considered and not raised" items unless there was new evidence.

## Verification of round-6 fixes

| # | Evidence | Result |
|---|---|---|
| R6-01 | **REQ-92 (`html:574`) now ends:** "never a pass or low score; any other non-success response is likewise a failed unit. Caps disclose `ranking: partial (n/N)`."<br>• This restores the catch-all in base `:178`: "the adapter treats any non-success as a failed unit".<br>• "Likewise" carries over the no-annotation treatment, the `ranking: deterministic (reason)` disclosure and the never-pass rule. It is consistent with REQ-93/97/179 and the Overview's "else deterministic" (`:266-268`).<br>• The ADR column is still "ADR-002, ADR-003, ADR-008". That is correct, because ADR-002 introduced the paragraph (rescope `proposed.diff:113`). | fixed |
| R6-02 | **Item 8 (`proposal.md:27`) now reads:** "§6 (user decisions of 2026-10-22 and 2026-10-23; now ADR-002 and ADR-003)".<br>• This matches base STATUS `:170` (§6, 2026-10-22, R-036) and `:186` (§6.9, 2026-10-23, R-039).<br>• It also matches ADR-002 `ADR.md:5` (cites STATUS §6) and ADR-003 `ADR.md:5` (cites §6.9). | fixed |
| R6-03 | **Item 7 (`proposal.md:26`) no longer uses quotation marks.**<br>• Its paraphrase ("no hypothesis changes state and … paths resolve to `docs/adrs/<folder>/…`") accurately describes `hypothesis-ledger.md:39`. | fixed |
| "Round 6 fixes" | **The section (`proposal.md:102-108`) is accurate.**<br>• It counts 1 minor and 2 nits, which matches round 6.<br>• The quoted R6-01 text matches `html:574`.<br>• "(base design §6)" is correct: `:178` lies between the §6 heading at `:158` and the §7 heading at `:192`.<br>• "The ADR column is unchanged" is true. | accurate |
| Conditions from round 6 | **The main agent recorded the outputs:**<br>• recomputed digests;<br>• `readlink` → `AGENTS.md`;<br>• `git diff --cached --stat`;<br>• `go build ./... && go vet ./...` → OK. | met (main-agent outputs accepted) |

## Issues

| ID | Severity | Location | Finding | Required fix |
|---|---|---|---|---|
| R7-01 | minor | `docs/scout-design.html:691` (REQ-174) | **X6 has lost its sampling-frame qualifier.**<br>• Base `docs/scout-design.md:441` (§13 X6) reads "Replay **last** 200 merged PRs of 3 adopter repos, three languages; double-label `would_block`". This has been unchanged since v1 (`docs/archive/scout-design-v1.md:358`) and the e1b-v2 base (`2026-10-21-e1b-v2/scout-design.base.md:357`).<br>• REQ-174 reads "X6 replays 200 merged PRs from 3 adopter repositories/three languages". Map `:124` sends "Design §13 X6" only to REQ-174, and the proposal lists no change of meaning.<br>• Without "last", any 200 merged PRs satisfy the requirement, including a favourable historical sample. The base fixed the most recent 200 to prevent that. **[INFERENCE]** on the practical effect: the eventual X6 preregistration would still have to fix the sample. | Restore "the last 200 merged PRs" in REQ-174. Leave the other wording and the ADR column (ADR-000, ADR-002, ADR-003) unchanged. |
| R7-02 | minor | `docs/scout-design.html:553` (REQ-16), `:536` (REQ-9) | **Appendix A no longer says which exit code a required lane that could not run produces, or that optional lanes never affect required lanes.**<br>• Base `:211` (§7.3) reads "optional lanes never turn a required lane red; a required blocking lane that could not run on head → exit 2". Base `:90` (§3.2 Coverage record) says Coverage gates "Only as 'required lane incomplete' (exit 2)". Both date from v1 (`scout-design-v1.md:76,220`) and are unchanged through every snapshot.<br>• REQ-16 says only "Required lane health drives exit 2/3", and REQ-24 lists the code labels. Map `:373` sends "Design §3.2 Coverage" to REQ-9 and REQ-11, and neither states Coverage's gate role. Apart from the REQ-24 label, the exit-2 mapping appears only in the Overview (`html:441,495,505`). The header says the Overview does not prevail over Appendix A.<br>• The generic optional-lane rule is carried only lane by lane: REQ-18 covers model-CLI skips and missing optional tools, REQ-93 the decisions lane, and REQ-154 the association lanes. No REQ covers a failed, as opposed to skipped, opted-in `contextual`/`clawpatch-review` lane.<br>• The base kept "required lane cannot run → 2" separate from "unavailable → 3". The rescope council treated this exact exit-2/3 ambiguity as a minor issue (`2026-10-22-rescope/round-1.md:23`, issue 11). The standard is the same as R5-02 and R6-01: the base states the rule explicitly, and the prevailing appendix now carries only a looser form. | In REQ-16, add: "a required blocking lane that could not run on head gives exit 2; optional lanes never turn a required lane red." In REQ-9, add: "Coverage gates only as 'required lane incomplete' (exit 2)." Keep the ADR columns: both clauses are v1, so ADR-000 is already cited. |

## Answers to the proposal's questions

1. **Is normative content lost, weakened or changed?** Two more losses were found (R7-01, R7-02). Otherwise no change of meaning was found:
   - the R6-01 restoration matches the base;
   - the challenger's trace found no loss or numeric mismatch. It covered the ledger, coverage and identity, the gate and exit contract, all six eligible allow-list rows, REQ-1…41, the metric formulas and bounds (CC > 10, NLOC ≥ 10, clone 50/5/10000/500 kB, CRAP > 30, 90/365-day windows, 50-file/100-commit caps, 4 h gaps, 0.5 h sessions, practical deltas) and the defaults and caps (2,000 calls/$0.50, 5 % audit, 0.8/0.8/0.9 minimums, association guard 20/20/0.40/0.20, experiment thresholds).
2. **Do the ADR summaries and cards match their records?** Yes:
   - all ten status lines agree with the card relations (`html:736…949`) and README `:15-25`, and every supersedes and amends edge appears on both ADRs involved;
   - card review counts match each `ADR.md`;
   - ADR-006/007/008 digests match base `area-refresh-fallback/round-5.md`, `cutover-cleanup/round-7.md`, and `jev-merge-association/round-8.md` and `round-11.md`;
   - the ADR-009 relation text ("Base design §1 and Appendix 11 said 'X1–X6, X9, X10'") matches base `:37` and `:488`.
3. **Does the Overview agree with Appendix A?** There is no contradiction. The Overview's exit legend (`html:440-442`) and incomplete arrow (`:495,505`) are in fact more precise than Appendix A on exit 2; R7-02 restores that precision to the appendix that prevails.
4. **Are rule, tooling and citation changes complete, with no live reference to a deleted file?** Yes:
   - deleted-file names match nothing in `internal/`, `cmd/`, `.github/`, `scripts/`, `toolchain/`, `.omp/`, `.prettierignore` or `.agents/`;
   - the live docs mention them only intentionally: `@9a3d19f5` pins, `jev-adapter-audit.md:7`, `html:630,953,956`, `plans/2026-11-06-remaining-work.md:101`, README `:27,29`, and earlier ledger rows;
   - the only Go string change is `cmd/scout/main.go:302`, `design §5.3` → `REQ-18`;
   - `.omp/gate.json`, `.omp/ledger.d/c.json`, `.prettierignore`, `infrastructure.py` and `test_gate_contract.py` agree with proposal item 6.
5. **Experiment rules, ledgers, frozen records:** no violation.
   - Experiment rules 1–7 match base `CLAUDE.md` word for word.
   - The research ledger is unchanged up to `:69`, and R-062 is appended at `:70`.
   - The hypothesis ledger adds one newest-first History line at `:39`; its header and HD8 row are unchanged.
   - No round, response, snapshot or diff under ADR-000…ADR-008 cites `docs/adrs`, `ADR-00n` or REQ-42+.
   - The eight renamed folders hold the same file sets as base.

## Considered and not raised

- **Prettier still checks `jev-adapter-audit.md`** (peer). Base handled the file the same way, and `prettier --check` does not modify files.
- **`STATUS.md:124` still says "HA2 proposed (X8)"** (peer). The same stale text is at base `:149`. STATUS is a living file outside the governed set, as settled in round 6 (F3).
- **The X6 rewording "feeding dossiers and shadow defaults, never bypassing per-repo activation"** has the same meaning as base "Feeds AL dossiers and shadow defaults; activation stays per repo". Only the missing "last" is raised.
- **REQ-24's label "2 incomplete" might already imply R7-02.** It is not enough. REQ-16's "drives exit 2/3" leaves the choice open, and REQ-9 says nothing about Coverage's gate role. The council has raised this exact ambiguity before (rescope round 1, issue 11).
- **Round 6's "Considered and not raised" items** are unchanged and still not issues.

## Peers

- **`innovation-challenger`:**
  - Verified R6-01 and its ADR column.
  - Traced the ledger, coverage, identity, gate and exit, allow-list, REQ-1…41, formula and default areas listed above.
  - Found R7-01.
  - Reported the gate and exit contract, including "required-lane incompleteness", as preserved. The lead disagrees (R7-02, below).
- **`experiment-peer`:**
  - Verified R6-02, R6-03 and the "Round 6 fixes" section, including the §6 location of base `:178`.
  - Checked all supersession edges, review counts, the ADR-006/007/008 digests, the live-tree and Go-literal greps, the tooling files, rules, ledgers, frozen-record integrity and the counts.
  - Found no issues and reported two items that are not raised (above).
- **Lead:**
  - Verified all three R6 fixes against the base.
  - Confirmed R7-01: base `:441`, v1 `:358` and e1b-v2 base `:357` all say "last"; the allow-list has no replay wording; map `:124`.
  - Found R7-02 independently. The lead traced base `:211`, which has no named map row of its own. That led to the Coverage row `:90`, map `:373` and the origin in v1 `:76,220`, and then to the absence of any exit-2 mapping in Appendix A (grep: only Overview `:441,495,505` and REQ-16's "exit 2/3").
- **Where the peers changed the outcome:** R7-01 came from the challenger; without it, the lead would have raised only R7-02. The challenger reported the gate area as clean. The lead rejected that for R7-02 on line-level evidence: the REQs the challenger cited for the gate (`html:586-591`) do not contain an exit-2 mapping.

## Verdict

**2 issues (R7-01 minor, R7-02 minor): NOT signed off.**

No sign-off applies to the digests above.

### Required with the next submission

- **Changed digests:** expected only for:
  - `docs/scout-design.html` (R7-01, R7-02);
  - `proposal.md` (a "Round 7 fixes" section).

  All other digests in the table above should stay as they are. If the main agent updates `requirements-map.md` to note the restored Coverage clause, that file's digest changes too, and the submission must say so.
- **Recomputed digests and checks:** the main agent must recompute every digest with `shasum -a 256` on the staged worktree. It must also re-run:
  - `test -L CLAUDE.md && readlink CLAUDE.md`, expecting `AGENTS.md`;
  - `git -C /Users/adamjackson/Projects/work/scout-req diff --cached --stat`.

  Record the outputs in the submission. Since no code changes, `go build`/`go vet` need not be re-run unless Go files are touched.

---

## Council memo

1. **Recommendation:** fix R7-01 and R7-02 in one pass and submit round 8.
   - Both are text restorations in Appendix A, and neither needs a new decision.
   - Before resubmitting, grep the base design for other "→ exit" clauses and for qualifiers on the §13 experiment rows ("last", "fresh", "unseen", "blind"). Confirm each appears in Appendix A, so the next round does not find another instance of the same class.
2. **Alternatives:**
   - (a) **Sign off with conditions.** Rejected: design rule 1 requires zero issues.
   - (b) **Treat R7-02 as implied** by REQ-24's "2 incomplete" label and the Overview. Rejected: Appendix A prevails over the Overview, REQ-16 leaves the choice between 2 and 3 open, and the council has raised this exact ambiguity before (rescope round 1, issue 11).
   - (c) **Treat R7-01 as a preregistration matter.** Rejected: the design fixes the sampling frame, and the restructure must not loosen it silently.
3. **Evidence:** each issue row cites the exact worktree and base `file:line`, and the verification table cites each R6 fix location.
4. **Risks:**
   - **R7-01:** an X6 preregistration could choose a non-recent PR sample and still cite REQ-174 as satisfied.
   - **R7-02:** an implementer could map a required lane failure to exit 3, or let an optional lane's failure turn the run red. CI callers that branch on exit codes (REQ-39 "stable exit codes") would then behave differently from the base contract.
   - **Reversibility:** all fixes are text edits and fully reversible.
   - **Unknown:** the council saw no per-file staged diff. Round 8 again relies on the main agent's statement about which files changed.
5. **Execution outline:**
   1. In REQ-174 (`html:691`), change "replays 200 merged PRs" to "replays the last 200 merged PRs".
   2. In REQ-16 (`html:553`), add "a required blocking lane that could not run on head gives exit 2; optional lanes never turn a required lane red". In REQ-9 (`html:536`), add "Coverage gates only as 'required lane incomplete' (exit 2)". Leave both ADR columns unchanged.
   3. Optionally, add REQ-16 to map row `:373` ("Design §3.2 Coverage"). If you do, report the changed `requirements-map.md` digest.
   4. Add "Round 7 fixes" to `proposal.md`.
   5. Recompute every digest with `shasum -a 256`, and run `test -L CLAUDE.md` and `git diff --cached --stat`. Confirm which files changed.
   6. Submit round 8 with the digest list and the command outputs.
