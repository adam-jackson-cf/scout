# Round 10: ADR-012 (shadow-window readiness inputs)

Date: 2026-11-08. This round follows design rules 1–3 and the ledger rules in `CLAUDE.md`, and the ADR rules in `docs/adrs/README.md`. ADR-012 is an agent decision made under the owner's standing delegation (R-064). Where a fix involves a choice, the council names the option it recommends under that delegation.

The round reviews the bytes revised after round 9 (`round-9-response.md`). It has three parts:

1. Check every ADR012-R9-01..05 fix against its required fix.
2. Re-confirm that every earlier closure (R1–R8) still holds in the current bytes.
3. Look for new issues, numbered ADR012-R10-xx.

The council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review. It looked for false-ready paths, readiness that a compliant repository could never reach, and loopholes in these areas: per-point lateness and visibility; the rule that C passes a point only through a normal gap; lineage and whole-history fast-forward; per-event reopening; the import-refusal lifecycle; the cancel-in-progress condition; fixture strength. It also checked the mirrors, the untouched requirements and the historical ADRs.
- **`experiment-peer` (Opus):** methodological review. It built concrete timelines to test whether each revised rule gives exactly one answer, and whether each acceptance fixture can expose the failure it targets.

The review was read-only. Nobody edited files, ran product code, made model calls outside the council, or touched `.codegraph`.

- The lead read the required codegraph skill references (`readiness-and-bootstrap.md`, `query-selection.md`). Codegraph was not used, because the material under review is documents and the council had no shell. Searches used `read`, `grep` and `glob`.
- The lead ran one web search and read two GitHub documentation pages: the workflow-syntax `concurrency` section and the Concurrency concepts page.
- The experiment-peer cited two GitHub documentation pages: control-workflow-concurrency and events-that-trigger-workflows.
- The challenger cited no external source.

**Main computed and verified the digests below with `shasum`. The council did not: the lead and both peers had read-only tools and no shell.** This round is not a sign-off, so the digests only identify the bytes reviewed. All paths are under `docs/`.

| File | SHA-256 (Main) |
|---|---|
| `adrs/2026-11-08-shadow-readiness-inputs/ADR.md` | `b5c4baa840d937ce51545eda1de3c3e810e785b581820bab27c735be618d5063` |
| `scout-design.html` | `2ed54b929b8700da36b70fda556a7b5aabcf3351146e1ef63b14b1457a87b210` |
| `adrs/README.md` | `cde415ff10b659a073d010f8a48c281d6f75a5a1cba466962d589e937b94587f` |
| `STATUS.md` | `17e6e4bf7f131111897600dcdbd17e5d51c13eace4f38b65cf7622357d3d563c` |
| `plans/2026-11-06-remaining-work.md` | `a0449bedb727d06372973f1cfa9706c00a3ef00636755516f54edb4e757e51e9` |
| `references/policy-file.md` | `333e22902221eef75ac3f356c4bccb477db6b7e7a1747a6d15062189609f602e` |
| `references/ci-integration.md` | `b3c2631f4dd4a0396dc5b14f23a016c0dd54cb6ff0e7b48d0466c7ae8e977fd6` |
| `references/cli-commands.md` | `08b73b27f907cdef2610af0a857e55ac70f3c4ffa5460a6f124ba366ec12d5c4` |
| `research-ledger.md` | `635fc9f53b2e258dfa29630a4b7c12a96c4e5d541675f295ece500b41440741a` |
| `adrs/2026-10-19-v1-decisions/ADR.md` (ADR-000) | `1cd521fd96439d87275ba353f7b05451c3e0c4d556ee665529aab85970ed81cb` |
| `adrs/2026-10-23-decisions/ADR.md` (ADR-003) | `973d24c4cfdfdd98acb4070a7bd72ca0c0e1ddf303bd4ccec8f92fe1cf117cc5` |
| `adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) | `ee593b6241814c24035671d666566bc2b7c399135867226c26e9626f3936abc8` |
| `adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) | `80e3375f81656a8856cf662cbec99a36d8e00bc451ff3c3592925d2293f44d95` |

The following record is not signed. Main computed its SHA-256:

| File | SHA-256 (Main) |
|---|---|
| `round-9-response.md` | `bf9928183ea68a3b4ce89fd8eac765c4436d15873529d516653abcd846406d7b` |

README and the four historical ADR files have the same digests as in round 9.

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11`, branch `agent/policy-shadow-adrs`, based on origin/main `7bb73e52`. Every change is staged and uncommitted.
- Base bytes come from `/Users/adamjackson/Projects/work/scout-adr11b`. The base has no ADR-012 folder, and its ledger has no R-07x rows.

**Files read.** Line numbers below refer to the raw worktree files.
- ADR-012 `ADR.md`, all lines:
  - `:16–22` items 1–7, where `:17` is population/exclusion and `:19` is "Clocks, frontier and late evidence";
  - `:26–30` triage, history import and ShadowReady;
  - `:34–40` persistence, manifest/merge, cutover and writer boundary;
  - `:44` CLI, live envelope and refusal scope;
  - `:48–52` amendments, sources and route;
  - `:56–62` implementation acceptance invariants.
- `round-9.md` (all) and `round-9-response.md` (all).
- `round-8-response.md:19–24`, `round-7-response.md:20–24`, `round-6.md:169–173`, `round-7.md:145` and `round-8.md:235`, all via grep context.
- `scout-design.html`, raw: REQ-22 `:589`, REQ-78 `:600`, REQ-27 `:612`, REQ-48 `:615`, REQ-51 `:618`, Activation callout `:449`, ADR-012 card `:1009`; grep context for `:610–611, 704–714, 736, 807, 901, 925, 1007–1010`.
- `cli-commands.md:126–134`, `policy-file.md:78–84` and `ci-integration.md:97–106`, all raw.
- Plan `:61,90`; STATUS `:121,131`; `research-ledger.md:75–76`.
- `.github/workflows/scout-guardian.yml:29–63` and `.github/workflows/scout-nightly.yml:38`.
- The challenger compared the bodies of the four historical ADRs and REQ-160/162/163/168 with base by reading the raw text.

## Issue summary

**4 distinct issues are open, all minor. None is major.**

| ID | Severity | Summary |
|---|---|---|
| ADR012-R10-01 | minor | Two R9-05 fixture fixes are incomplete. Plan B4-S and `ci:106` still carry the combined pin/inspection rerun fixture. The ADR's deadline-lateness fixture asserts only that the late facts are invisible before their stamp, so an implementation that drops them passes |
| ADR012-R10-02 | minor | The new normal-boundary fixture allows point = new C, but then requires a selected interval that the exclusive-boundary rule makes impossible |
| ADR012-R10-03 | minor | "Destination C at publication" has two readings when the publishing commit also advances C. They give different digests and gate results, and one keeps certification refused in an active repository |
| ADR012-R10-04 | minor | Inferred supersession tests the cancelled run's cancel-in-progress. GitHub applies the newly queued (superseding) run's setting, and a PR can edit the workflow it runs |

## Closure of round-9 issues

"Fixed" means the required document fix is present in the reviewed bytes at every named location. It does not mean the implementation is validated. "Fixed as prescribed" means the council's own prescription was applied but left a gap, which is carried as a new R10 issue.

| Issue | Status | Evidence |
|---|---|---|
| ADR012-R9-01 | **Fixed as prescribed** | **Per-point lateness:** `ADR.md:19` "each gap point independently is late when point≤that C".<br>**Single predicate:** "facts and gap points require their provider coordinate≤as_of and, only if late, stamp≤as_of; pending facts are never visible".<br>**Frontier corollary:** "C may pass a point only through a published gap in which that point is normal, hence visible at every cutoff the advance permits".<br>**Blanket sentence replaced:** "Late points and late facts use only the same visibility predicate; earlier dossiers stay identical".<br>**Mirrors:** REQ-78 `:600` (verbatim); `cli:126`; `policy-file:82`; `ci:101`.<br>**Fixtures:** `ADR.md:58`, plan `:61`, `ci:106`.<br>The peer re-ran both R9 timelines (S and P) and both now give one answer.<br>**Carried:** the fixture contradiction at equality is ADR012-R10-02; the meaning of "C at publication" is ADR012-R10-03. |
| ADR012-R9-02 | **Fixed** | **Lineage:** `ADR.md:28` "Lineage is the trusted publication chain, identified at least by provider repository ID, state ref and genesis publication ID … authenticated chain ancestry must be proved".<br>**Prefix test:** the "entire published history (all frontiers, stamped facts and rows, including post-C ones)" must be an identical prefix.<br>**No vacuous pass:** "A non-empty destination never qualifies vacuously"; missing or unverifiable lineage is cross-lineage.<br>**Per-event reopening:** "makes only that event Undecided"; "a verdict between import publications cannot resolve subsequent rows". `ADR.md:36` adds "lack of C alone never grants a non-empty destination fast-forward status".<br>**Mirrors:** REQ-48 `:615`, REQ-78, `cli:130`, `policy-file:82`.<br>**Fixtures (all four):** `ADR.md:60`, plan, ci.<br>Both peers re-ran the diverged-prefix, E′-only and between-publications timelines; each gives one answer. |
| ADR012-R9-03 | **Fixed** | **Record:** `ADR.md:28` "immutable import-refusal record (action ID, event ID, reason, lineage, CAS recorded_at), idempotent by action ID and exported/imported under these same rules"; "filter history/refusal records by stamp≤as_of".<br>**Trigger:** `ADR.md:44` "triage/history records stamped>C that change this exact cohort's veto-event triage state, whether authorized verdicts or import-refusal records".<br>**Mirrors:** REQ-48 ("Persist/export/import idempotent stamped import-refusal records"), REQ-78, `cli:130`, `policy-file:82`.<br>**Fixture:** the revoked-author fixture (`ADR.md:60`, plan, ci) fails an implementation that drops the refused row, one that stays refused forever, and one that stays Undecided forever (peer). |
| ADR012-R9-04 | **Fixed as prescribed** | `ADR.md:17` "the cancelled run's pinned lane declares cancel-in-progress for that group". The same condition is in REQ-27 `:612`, REQ-78 and `ci:101`; `cli:126` has "the cancelled run's pinned cancel-in-progress". Fixture: `ADR.md:57`, plan, ci.<br>**Carried:** the condition tests the wrong run (ADR012-R10-04). |
| ADR012-R9-05 | **Not fully fixed** | **(a)** Fixed only in `ADR.md:58`. Plan `:61` and `ci:106` still have the combined fixture.<br>**(b)** Fixed everywhere: "a new authorized destination verdict after the last import publication must restore ready".<br>**(c)** `ADR.md:57` narrows the prescribed "late and stamp-gated" to "late and invisible before their stamp".<br>**(d)** Fixed everywhere: the pending-local, pending-import and unrelated-check fixtures.<br>**Carried:** (a) and (c) are ADR012-R10-01. |

## Re-confirmation of earlier closures (R1–R8)

| Closure | Holds? | Evidence |
|---|---|---|
| R1-04: started_at membership; elapsed-days formula | Yes | `ADR.md:19–21` |
| R1-05, R7-01, R8-07: recoverable frontier, deadline point gaps | Yes, apart from ADR012-R10-03 | `ADR.md:19`. The parent-C reading keeps C recoverable. The other reading of "C at publication" threatens liveness |
| R2-05, R6-01/08, R7-04, R8-05: population, exclusions, causal supersession | Yes, apart from ADR012-R10-04 | `ADR.md:16–17`; REQ-27/78; `ci:101` |
| R5-09: scanner-output failure is comparison-none | Yes | `ADR.md:17` |
| R1-08/09/11/14, R2-10, R3-01, R4-07/10: CI/local/legacy separation, migration, v2 cutover, query-only v13 | Yes | `ADR.md:34–38,44`; REQ-10 `:610`; REQ-48 `:615` |
| R3-02/07, R4-02/08/11, R8-02: authorization, current-tip authority, replay, writer isolation, import authorization | Yes | `ADR.md:26,28,40`; REQ-26 `:611`, REQ-51 `:618`, REQ-161 `:707`; `ci:102–103`. R9-02/03 now close the import edges |
| R3-03, R4-06: guardian comparison, empty override inputs | Yes | `ADR.md:17–18`; REQ-27 |
| R2-11, R3-05, R4-03/04, R5-03/05/07, R6-02/07, R7-01, R8-03: interval selection, cutoffs, frontier minima, dossier invariance, single predicate | Yes, apart from ADR012-R10-03 | `ADR.md:19–20,28,44`; REQ-78; `cli:126–132` |
| R8-04: points are exclusive interval boundaries | Yes | `ADR.md:19–20`. The new fixture contradicts it at equality (ADR012-R10-02); the rule itself stands |
| R5-06, R6-03, R8-01: liveness, skewed-clock triage, scoped refusal | Yes | `ADR.md:19,44`. The trigger list now also covers import-refusal records |
| R5-01, R6-10: wiring-ADR gate | Yes | `ADR.md:12,56`; REQ-51/161; `ci:99`; plan `:61`; `policy-file:84`; `cli:134` |
| R5-04, R6-04, R7-02, R8-06: cohort-wide zero-not_warranted/zero-Undecided, including provisional; fixtures cannot pass by refusal alone | Yes | `ADR.md:20,28,30,58–60`; REQ-22 `:589` |
| R5-08, R6-05: Days, representative origin | Yes | `ADR.md:21,48,62`; card `:1009` |
| R5-10, R7-05/06/07: exact cohort, default version, dossier encoding, attempt-ID order, pin tuples | Yes | `ADR.md:21,44,61–62`; REQ-78; `cli:126,130` |
| R7-03: event-to-attempt binding | Yes | `ADR.md:17,57`; REQ-27/78 |
| R7-08/09: mirror drift; Activation callout | Yes, apart from ADR012-R10-01(a) | REQ-27 normative reference; callout `:449`; card `:1009`; STATUS `:121`; plan `:90` |
| Historical ADRs and untouched requirements | Yes | ADR-000/003/007/008 and README digests are the same as in round 9. The challenger's raw comparison found that only status lines differ from base. REQ-160/162/163/168 (`:706,708,709,714`) match base |
| Ledger | Yes | R-077 (`research-ledger.md:76`) links `round-9.md`. STATUS `:131` names next free R-078. No R-ID or ADR ID is reused |

## Are the invariants testable and enough for REQ-22, REQ-27 and REQ-78?

| Requirement need | Testable and sufficient? | Where / gap |
|---|---|---|
| REQ-22: counting unit, deduplication, exact cohort, unknown ≤ 10 % | Yes | `ADR.md:20–22` |
| REQ-22: 14 days, inspected origin, `as_of` endpoint | Yes | `ADR.md:19,21,58,62` |
| REQ-22: zero `not_warranted`/zero Undecided across gaps, late and imported evidence | Yes in the rules; fixtures **partly** | ADR012-R10-01(c): a late terminal veto fixture can pass with the facts dropped |
| REQ-22: maintainer-only, append-only, idempotent triage | Yes | `ADR.md:26,28,60` |
| REQ-27: population independent of content, author and approval | **Partly** | ADR012-R10-04 |
| REQ-78: complete-interval accounting, recoverable frontier | **Partly** | ADR012-R10-03 |
| REQ-78: historical reproduction, `dossier_digest` | **Partly** | ADR012-R10-02 (fixture); ADR012-R10-03 (two digests for one state) |
| REQ-78: current certification and liveness | **Partly** | ADR012-R10-03 |
| Governance of deferred wiring | Yes | `ADR.md:12,56`; REQ-51/161 |

## Issues

### ADR012-R10-01: Two R9-05 fixture fixes are incomplete (minor)

- **Location:**
  - Plan `:61` and `ci:106`: "Historical day30 default-version cohort/digest stays identical before/after pending import and publication of a later-started late rerun with changed pin/inspection."
  - `ADR.md:57`: "Advance C past its started_at before publishing terminal facts, then assert they are late and invisible before their stamp." Plan `:61` and `ci:106` say "late/stamp-gated".
  - Claims that this was done: `ADR.md:48` ("replace ineffective combined-pin fixtures with independent assertions") and the `round-9-response.md` mapping ("ADR acceptance, B4-S and CI carry the strengthened assertions").
- **Problem:**
  - **(a) Combined fixture still in two mirrors** (lead and both peers). R9-05(a) found that a different-pin rerun falls outside the v1 cohort whatever its visibility, so a leak into representative selection alone passes. Round 9 required the replacement at "ADR.md:57–60, plan B4-S and ci:106". Only `ADR.md:58` has the two independent assertions. Both mirrors still list the ineffective fixture under "Required fixtures", which contradicts the ADR and the response. The general "Prove all ADR-012 acceptance invariants" clause stops the mirrors from authorizing a weaker implementation, but it does not apply the required fix.
  - **(c) One-sided deadline fixture** (experiment-peer) [INFERENCE]. An implementation can treat the deadline point gap as the attempt's final record and discard the later terminal facts, for example under "reject conflicting finals" (`ADR.md:36`). It then satisfies "late and invisible before their stamp" and never vetoes on that attempt's would_block event. That contradicts "later terminal facts follow the late lifecycle" (`ADR.md:19`). The generic late would_block fixture (`ADR.md:58`) does not exercise this deadline path. Round 9 prescribed "late and stamp-gated", which covers both halves; the ADR kept only the first.
- **Required fix:**
  - **(a)** In plan `:61` and `ci:106`, replace that sentence with the two independent assertions from `ADR.md:58`: the same-pin/different-inspection late rerun must preserve the historical representative, and the different-pin late rerun must preserve the historical default version.
  - **(c)** In `ADR.md:57`, plan `:61` and `ci:106`, add the second half. With C ≥ the terminal facts' stamp and no refusal trigger, the late terminal attempt and its linked would_block event are visible, giving computed exit 0, not ready, with an Undecided reason. Pair it with an otherwise identical ready control without the late terminal facts.

### ADR012-R10-02: The normal-boundary fixture requires an impossible result at point = new C (minor)

- **Location:** `ADR.md:58`: "With parent C<point≤new C<stamp, the new-C query shows the normal boundary: the previous interval ends before the point and its selected interval starts after it." Plan `:61` and `ci:106` have the same text.
- **Problem** (challenger) [INFERENCE]: take parent C = 30, point = new C = 31, stamp = 32.
  - `ADR.md:19` says "previous interval ends strictly before it, next starts strictly after it, and attempts at it belong to no interval". So no interval contains `as_of` = 31.
  - `ADR.md:44` maps "no complete containing interval" to unmeasured, exit 2.
  - At the boundary the fixture permits, a correct implementation therefore fails it. An implementation that puts `as_of` = point inside the next interval, which violates the exclusive boundary, passes. This is the R9-05(c) defect class: a fixture that either fails a correct implementation or rewards a wrong one.
- **Required fix (recommended option):** in all three places, split the fixture:
  - **parent C < point < new C < stamp:** the new-C query selects the interval starting strictly after the point, and the previous interval ends strictly before it;
  - **parent C < point = new C < stamp:** the new-C query reports no complete containing interval (unmeasured, exit 2), and a later cutoff after the point selects the next interval.

### ADR012-R10-03: "Destination C at publication" is ambiguous when the publishing commit also advances C (minor)

- **Location:**
  - `ADR.md:19`: "A fact is late when its started_at (attempt) or event_at (event) is ≤ destination C at publication; each gap point independently is late when point≤that C".
  - `ADR.md:19`: "C may pass a point only through a published gap in which that point is normal, hence visible at every cutoff the advance permits".
  - `ADR.md:19` also lets one publication both publish facts and advance C: "Advance C at triage publication and periodic trusted reconcile"; "A stamping publication reaches its stamp only if its attested bound minus lag permits".
  - By contrast, the CAS formula and the fixtures use "parent C" (`ADR.md:19,58`).
  - Mirrors: REQ-78 `:600`; `cli:126` ("only if coordinate≤destination C at publication (late)"); `policy-file:82`; `ci:101`.
- **Problem** (experiment-peer) [INFERENCE]:
  1. Parent C = 30.0.
  2. A trusted collection run starts at 32.0 with an attested bound of 32.0 and a lag of 0.05.
  3. In one CAS commit it publishes terminal attempt A (started 31.0, linked to an Undecided would_block event W) and new C = 31.95. A's stamp is 32.0.
  - **Reading "parent C":** A is normal and visible at `as_of` = 31.95. The result is computed exit 0, not ready (W), and certification is `current`.
  - **Reading "C after this commit":** A is late. Its stamp is always later than the commit's own new C, because stamp ≥ run start > bound − lag ≥ new C. So A and W are invisible at every cutoff ≤ 31.95, and A is a "repository late record stamped>C", which gives refused, exit 2.
  - One state therefore has two `dossier_digest` values and two gate results. In an active repository where every collection commit both advances C and carries new attempts, the second reading keeps certification refused indefinitely. That is the R5-06/R8-01 liveness harm, and it would make readiness unreachable for a compliant repository.
  - Under the same reading, a gap published in the advancing commit is late, so the commit cannot pass its own points. That collides with the frontier corollary.
  - **Severity:** the lead adjudicates minor, because the CAS formula and the R9-01 fixture both point to the parent reading. The rule still has to give one answer.
- **Related boundary** (challenger, folded in here). Read literally, the frontier corollary could also apply to a late point exactly equal to parent C. The challenger rated that as a major deadlock. The council does not raise it separately (see "Considered and not raised"), but the rewrite below removes it.
- **Required fix (recommended option):**
  - In `ADR.md:19`, REQ-78, `cli:126`, `policy-file:82` and `ci:101`, define "destination C at publication" as the frontier in the publication's CAS parent, the same parent C the stamp formula uses. A frontier advance carried by the same commit never makes that commit's own facts or points late.
  - State that the normal-gap condition governs only points in (parent C, new C]. Points ≤ parent C, including equality, are late and never block an advance.
  - **Fixtures** (`ADR.md:58`, plan, ci):
    - One commit carries the C advance, a terminal attempt with parent C < started_at ≤ new C < stamp, and a point gap with parent C < point < new C. The attempt is normal and visible at new C, the boundary is visible, and certification is `current`.
    - A late point equal to parent C, previously accounted without a gap, does not stop C advancing to its stamp. Pre-stamp dossiers stay identical.

### ADR012-R10-04: Inferred supersession tests the cancelled run's setting, but GitHub applies the superseding run's (minor)

- **Location:** `ADR.md:17` ("the cancelled run's pinned lane declares cancel-in-progress for that group and the superseding run was created before cancellation with no actor-initiated cancel recorded"); REQ-27 `:612`; REQ-78 `:600`; `ci:101`; `cli:126`. Fixture: `ADR.md:57`, plan `:61`, `ci:106`.
- **Evidence:**
  - GitHub workflow syntax: "To also cancel any currently running job or workflow in the same concurrency group, specify `cancel-in-progress: true`." The same page's example uses an expression evaluated in the new push's context: "multiple pushes to a `release/1.2.3` branch would not cancel in-progress runs" ([workflow syntax](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax)). The setting therefore belongs to the newly queued run [INFERENCE from the documentation wording].
  - The Concurrency concepts page says pending runs are replaced by default: "any additional pending runs cancel the previous one" ([concurrency](https://docs.github.com/en/actions/concepts/workflows-and-actions/concurrency)).
  - `.github/workflows/scout-nightly.yml:38` calls the local `./.github/workflows/scout-guardian.yml`, whose concurrency block is at `:41–43`. `pull_request` runs use the PR's merge context (R-070), so a PR can edit the configuration its own run executes.
- **Problem** (experiment-peer; the lead found it independently) [INFERENCE]:
  1. H1's run R1 executes with `cancel-in-progress: true`, the pinned default-branch setting.
  2. The author pushes H2 with the setting changed to `false`, or a base change merged between the two runs has the same effect.
  3. R2 is created, and nothing cancels R1.
  4. The author manually cancels R1, and the provider records no canceller.
  5. Every condition in the branch holds, so H1 is excluded by the author's choice. This is the harm R8-05/R9-04 set out to close.
  - The council's own R9-04 prescription named the cancelled run. GitHub's documented semantics are new evidence.
- **Required fix (recommended option):**
  - The inferred branch applies only when the superseding run's executed lane configuration equals the pinned protected-default configuration and evaluates `cancel-in-progress` to true for the same evaluated concurrency group as the cancelled run. Otherwise, or when this cannot be verified, the cancellation is assessed or gapped. The cancelled run's own setting plays no part.
  - Mirror in REQ-27/78, `ci:101` and `cli:126`.
  - Record the GitHub concurrency documentation link and fact in R-070. R-070 is ADR-012's own unadopted GitHub-semantics row (absent from the base ledger).
  - **Fixture** (`ADR.md:57`, plan, ci): R1 runs with `true`; R2's executed configuration has `false` (edited in the PR, or changed on the base between the runs); a manual cancel with no recorded canceller leaves H1 assessed or gapped. Keep the existing all-`false` fixture.

## Checks that found no issue

- **Fast-forward integrity.** A copied lineage ID cannot pass, because authenticated ancestry and a whole-history prefix including post-C records are both required. A destination-only post-C correction makes the import cross-lineage (both peers).
- **Per-event reopening.** E′-only imports leave E resolved, and a verdict between two import publications cannot resolve E. A replay reopens nothing (both peers).
- **Import-refusal lifecycle.** The record is stamp-filtered, triggers refusal, round-trips through export/import and is idempotent by action ID. Importing a refusal record cross-lineage yields Undecided under either reading of its authorization, so the gate result is the same (peer).
- **R9-01 timelines.** Reading S: a normal gap point at 31 is visible at new C = 31.9. Reading P: the 08:00 late point stays invisible before its stamp, and earlier dossiers are unchanged (peer).
- **Exit codes and the historical bypass.** These are unchanged: `ADR.md:44`; `cli:130–132`; `policy-file:82`.
- **Mirrors.** REQ-22 `:589`, REQ-51 `:618`, callout `:449`, card `:1009`, STATUS `:121` and plan `:90` agree with the ADR. Apart from ADR012-R10-01(a), every round-9 fixture appears in plan `:61` and `ci:106`.

## Considered and not raised

- **"A late point equal to existing C permanently blocks advancement"** (challenger, rated major). The corollary's own purpose clause, "visible at every cutoff the advance permits", limits it to newly permitted cutoffs (parent C, new C]. The R9-01 fixture's "parent C<point≤new C" treats reaching a point as passing it. A point equal to parent C was therefore already passed by parent C, and the experiment-peer reached the same reading independently. It is not a separate defect. ADR012-R10-03's fix states the range explicitly anyway.
- **A mixed gap record (normal started_at, late event_at) counted as a "repository late record"** (peer). Because started_at ≥ event_at, the late point becomes visible only where the normal point already ends the interval, so dossiers never differ. The worst case is a temporary, fail-closed refusal.
- **Whether per-event reopening applies to fast-forwards** (lead, peer). It does not: "exactly as fresh restore" and the "sealed local fast-forward equals fresh-restore dossiers" fixture both rule it out.
- **Import of an export whose lineage cannot be authenticated into an empty store, or a v13-migrated store with only legacy rows** (lead, peer). Either import is treated as cross-lineage, so its rows stay pending and that local reader is refused. This fails closed, the REQ-46 fresh-restore remedy clears it, and it is consistent with "byte-identically where no new destination publication is needed".
- **Import-refusal records are not named in the schema-14 table list or the manifest-v2 file list** (`ADR.md:34,36`) (lead). REQ-48 places them "alongside history", and the ADR leaves storage layout to the implementation. This is wording.
- **The rule that late facts' points are gaps is stated only implicitly** (peer open question). It survives in "Late points" and "Report late gaps by stamp" (`ADR.md:19`), "Counts include late gaps" (`:44`), REQ-48 "count their points as gaps" and the day-20 fixture (`:58`). The deleted round-9 sentence governed timing, not whether points are created. R7-01 and R8-04 were closed on the same text.
- **Default replacement of pending runs when cancel-in-progress is false** (lead). Without a provider-recorded concurrency cause such runs are gapped. That fails closed and does not make readiness unreachable.
- **"Each test advances C≥all stamps" placed before the R9-01 fixtures, whose new C < stamp** (peer). This is about sentence scope and is wording.
- **The R9-01 fixture does not assert certification = current** (peer). A refusal there would fail closed and clear once C ≥ stamp. ADR012-R10-03's fixture asserts `current` for the same-commit case.

## Peers and where they changed the outcome

- **`innovation-challenger`** raised 1 major and 2 minor findings. It rated R9-02/03/04 fixed, R9-01 fixed as prescribed and R9-05 not fixed at every location.
  - The R9-05(a) mirror gap became ADR012-R10-01(a). The lead had found it independently.
  - The equality fixture contradiction became ADR012-R10-02. **This was new to the lead.**
  - Not adopted: the equality deadlock as a major issue. It was folded into ADR012-R10-03's fix as a clarification.
- **`experiment-peer`** raised 4 minor findings. It rated R9-02/03 fixed, R9-01/04 fixed as prescribed and R9-05(a) not fixed.
  - The ambiguity of "C at publication" became ADR012-R10-03. **This was new to the lead and changed the lead's view** of whether R9-01 had fully closed.
  - The wrong-run supersession condition became ADR012-R10-04. The lead had found it independently; the peer's `scout-nightly.yml:38` evidence that a PR can edit its own run's configuration strengthened it.
  - The mirror fixture gap became ADR012-R10-01(a), as above.
  - The one-sided deadline fixture became ADR012-R10-01(c). **This was new to the lead.**
- **Lead-only:**
  - the R-070 ledger obligation in ADR012-R10-04's fix;
  - the import-refusal storage check (not raised) and the default pending-replacement check (not raised);
  - all severity calls, including down-rating the challenger's major.

## Verdict

**ADR-012: 4 issues open (0 major, 4 minor). Not signed off.** Open: ADR012-R10-01, -02, -03 and -04.

Round-9 closure:
- R9-02 and R9-03 are fixed.
- R9-01 and R9-04 are fixed as prescribed. Their remaining gaps are carried as ADR012-R10-02, -03 and -04.
- R9-05 is not fully fixed. Items (a) and (c) are carried as ADR012-R10-01.

Earlier closures hold. The frontier and dossier-invariance groups are qualified by ADR012-R10-03, and the population group by ADR012-R10-04.

No digests are signed in this round. The 13 digests above identify the bytes reviewed.

## Council memo

1. **Recommendation.** Do not sign off. Apply the 4 fixes using the recommended options, then submit round 11 with fresh digests from Main for all 13 signed files and the response.
   - The round-9 structural fixes worked: lineage, whole-history fast-forward, per-event reopening, import-refusal records and the single per-point predicate are sound.
   - What remains is narrow: one definitional anchor (the parent C), one causal-source correction (the superseding run), and three fixture repairs.
2. **Alternatives.**
   - **R10-03:**
     - Recommended: anchor lateness to the CAS-parent C.
     - Alternative: forbid a commit from both publishing facts and advancing C, so advances happen only in a separate reconcile. That removes the ambiguity but slows C and adds a publication per advance. Not recommended.
   - **R10-04:**
     - Recommended: the superseding run's executed configuration must equal the pinned one and evaluate true.
     - Alternative B: require both runs' configurations to be true. It is safe but adds a condition with no causal role.
     - Alternative C: drop the inferred branch and rely only on provider-recorded concurrency cause. That is the most fail-closed option, but under today's `cancel-in-progress: true` every superseded run would gap and reset intervals whenever GitHub exposes no structured cause, which can deflate readiness severely in active repositories. Not recommended.
   - **R10-02:**
     - Recommended: split the fixture at strict inequality and equality.
     - Alternative: restrict the fixture to point < new C only. It is simpler, but leaves the equality behaviour untested.
   - **Signing off with known minor issues** is not permitted under design rule 1.
3. **Evidence.**
   - `ADR.md:17,19,36,44,48,57–58`; REQ-78 `:600`; REQ-27 `:612`; `cli:126`; `policy-file:82`; `ci:101,106`; plan `:61`.
   - `.github/workflows/scout-nightly.yml:38`; `scout-guardian.yml:41–43`; `research-ledger.md:75`.
   - GitHub [workflow syntax](https://docs.github.com/en/actions/reference/workflows-and-actions/workflow-syntax) and [concurrency](https://docs.github.com/en/actions/concepts/workflows-and-actions/concurrency).
   - `round-9.md` required fixes and `round-9-response.md` claims.
4. **Risks.**
   - ADR012-R10-04 relies on a reading of GitHub's documentation [INFERENCE]; the implementation must confirm the provider's behaviour.
   - All timelines were constructed, not executed.
   - The council could not compute digests.
   - Every fix is a document change and fully reversible. Implementation remains pending, so no shipped behaviour depends on these bytes.
   - Residual risks accepted in earlier rounds (author-forced interval resets, scanner self-falsification, provider limits) are unchanged.
5. **Execution outline.**
   1. R10-01: in plan `:61` and `ci:106`, replace the combined rerun fixture with the two assertions from `ADR.md:58`. In `ADR.md:57`, plan and ci, add the post-stamp visibility and Undecided-veto half of the deadline fixture, with its ready control.
   2. R10-02: in `ADR.md:58`, plan and ci, split the normal-boundary fixture into strict-inequality and equality cases.
   3. R10-03: in `ADR.md:19`, REQ-78, `cli:126`, `policy-file:82` and `ci:101`, define "C at publication" as the CAS-parent frontier and bound the normal-gap condition to (parent C, new C]. Add the same-commit and equal-to-parent-C fixtures.
   4. R10-04: in `ADR.md:17`, REQ-27/78, `ci:101` and `cli:126`, condition the inferred branch on the superseding run's executed configuration being the pinned one and evaluating true. Add the mixed-configuration fixture. Record the GitHub concurrency documentation in R-070.
   5. Write `round-10-response.md` mapping each issue. Main supplies the SHA-256 of all 13 signed files and of the response, then round 11 follows.
