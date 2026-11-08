# Round 5: ADR-012 (shadow-window readiness inputs)

Date: 2026-11-08. This round follows design rules 1–3 and the ledger rules in `CLAUDE.md`, and the ADR rules in `docs/adrs/README.md`. ADR-012 is an agent decision under the owner's standing delegation (R-064). Where a fix involves a choice, the council names the option it recommends under that delegation.

This is the first round on the narrowed ADR. After round 4, the writer cut ADR-012 back to readiness evidence, counting and triage invariants, and left workflow, transport and writer mechanics to the implementation PR (`round-4-response.md`, "Delegated scope cut").

The council lead ran the round with two independent peers. Neither peer was told the lead's view.

- **`innovation-challenger` (GPT):** adversarial review. It checked R4 closure, tried to build an implementation that meets the narrowed contract but breaks REQ-22/27/78, and looked for contradictions across the signed set.
- **`experiment-peer` (Opus):** methodological review. It treated readiness as a measurement and looked for confounds, inflation, deflation and gaming. It checked R4 closure and asked whether the deferred invariants can be tested.

The review was read-only. Nobody edited files, executed product code, ran model calls outside the council, or touched `.codegraph`.

**Main computed and verified the digests below with `shasum`. The council did not: the lead and both peers had read-only tools and no shell.** This round is not a sign-off, so the digests only identify the bytes reviewed. All paths are under `docs/`.

| File | SHA-256 (Main) |
|---|---|
| `adrs/2026-11-08-shadow-readiness-inputs/ADR.md` | `9857fd080d6db4c30464067b9fc31206954219d43217a4655b495b580db09618` |
| `scout-design.html` | `48f81b8ce938c5c464da27530290fdfb996dd9dc4bf8433457294d4f3282e3a6` |
| `adrs/README.md` | `cde415ff10b659a073d010f8a48c281d6f75a5a1cba466962d589e937b94587f` |
| `STATUS.md` | `65ca25eef914eaef0023c16b0fe56085bf1d1d76af95259940155b9a8f2e3e77` |
| `plans/2026-11-06-remaining-work.md` | `f6f51e541e18a756909695caccfdc4eb4dafa4d44ca215b70b6f843b70e7f21d` |
| `references/policy-file.md` | `77b2183e4eeb7268baa2312f83ab1086876c21d2ce950070511ca3616a7f4235` |
| `references/ci-integration.md` | `c7be63817a2709e54bdd11fad6bffbfb3c8121de673f07c863e805a5d3465044` |
| `references/cli-commands.md` | `4cd9b9b0a645a9ba299928bae836147e488a71f7bd55cb90c9bf7add48cf990b` |
| `research-ledger.md` | `ed665a05134e891c7e7347742f243ff18fb7bc82c87420b248225478992cf6d2` |
| `adrs/2026-10-19-v1-decisions/ADR.md` (ADR-000) | `1cd521fd96439d87275ba353f7b05451c3e0c4d556ee665529aab85970ed81cb` |
| `adrs/2026-10-23-decisions/ADR.md` (ADR-003) | `973d24c4cfdfdd98acb4070a7bd72ca0c0e1ddf303bd4ccec8f92fe1cf117cc5` |
| `adrs/2026-10-24-cutover-cleanup/ADR.md` (ADR-007) | `ee593b6241814c24035671d666566bc2b7c399135867226c26e9626f3936abc8` |
| `adrs/2026-10-31-jev-merge-association/ADR.md` (ADR-008) | `80e3375f81656a8856cf662cbec99a36d8e00bc451ff3c3592925d2293f44d95` |

The round-4 response record is not signed. Its SHA-256 (Main) is `ff738f11f4ff9c829e3ef4a5fcd3126c1bda55e605e5861257d879c021306716`.

## Scope

**Under review:**
- Worktree `/Users/adamjackson/Projects/work/scout-adr11`, branch `agent/policy-shadow-adrs`, based on origin/main `7bb73e52`.
- Every change is staged and uncommitted.
- Base bytes come from `/Users/adamjackson/Projects/work/scout-adr11b`.

**Files read.** Line numbers below refer to the raw worktree files.
- ADR-012 `ADR.md:1–58`, all lines:
  - `:3–4` status and revision;
  - `:12` delegated narrowing;
  - `:16–22` items 1–7 on which observations count;
  - `:26–30` triage, cutoffs and ShadowReady;
  - `:34–40` persistence, cutover and writer boundary;
  - `:44` CLI;
  - `:48–52` amendments, sources and route.
- `round-4.md` (all 299 lines), `round-4-response.md`, `round-3.md:212–224,335–346`, `round-2.md:195–206`, and `round-1-response.md:19–23`, located by grep.
- `scout-design.html`, worktree against base:
  - the Activation callout `:449`;
  - REQ-22 `:589`, REQ-78 `:600`, REQ-10 `:610`, REQ-26 `:611`, REQ-27 `:612`, REQ-48 `:615`, REQ-51 `:618`, REQ-160 `:706`, REQ-161 `:707`, REQ-162/163 `:708–709` and REQ-168 `:714`;
  - the cards for ADR-000 `:736`, ADR-003 `:807`, ADR-007 `:901`, ADR-008 `:925`, ADR-011 `:992–1003` and ADR-012 `:1006–1012`.
- `adrs/README.md`, all lines, worktree against base.
- Line 3 of ADR-000, ADR-003, ADR-007 and ADR-008 in both trees. The challenger compared all four bodies with base.
- ADR-011 `ADR.md:1–12`.
- `STATUS.md:7,9,121,131`; plan `:15,35,61–62,90,92,100`; `ci-integration.md:97–106`; `cli-commands.md:120–134`; `policy-file.md:76–84`; `research-ledger.md:73–76`.
- Code:
  - `internal/policy/policy.go:384–458`: `ShadowStats`, `ShadowReady` and `AcceptOverride`;
  - `internal/engine/gate.go:357,382–385`: `wouldBlockID`.

## Issue summary

**12 distinct issues are open: 5 major and 7 minor.**

| ID | Severity | Summary |
|---|---|---|
| ADR012-R5-01 | major | The wiring left to the implementation PR needs changes to REQ-160–163/168, which are governed design, but the ADR calls those choices "not additional product requirements" and gives them no review path |
| ADR012-R5-02 | major | The evidence population ("every eligible attempt") is undefined, and the round-2 default-branch/protected-base eligibility rule was lost in the narrowing |
| ADR012-R5-03 | major | Attempt facts are not publication-safe: how collectedThrough advances, terminal completeness, and the `started_at` clock are all undefined, so `--as-of` results can change |
| ADR012-R5-04 | major | A collection gap starts a new interval and forgets an earlier `not_warranted` verdict for the same exact cohort |
| ADR012-R5-05 | major | `recorded_at` does not order successive triage publications, so a later verdict can lose to an earlier one when runner clocks are skewed |
| ADR012-R5-06 | minor | A newly published verdict stays invisible to `current_window` until C advances, and nothing requires C to advance on a quiet repository |
| ADR012-R5-07 | minor | Historical reproduction does not fix which readiness policy (`min_days`, `min_pr_runs`) applies at a cutoff |
| ADR012-R5-08 | minor | Days has no endpoint, and the inspected-only Days origin and `not_inspected` PRRuns exclusion are not declared as counting changes |
| ADR012-R5-09 | minor | The boundary between a gap and comparison-none for missing or invalid output is unclear, so a PR could induce a gap |
| ADR012-R5-10 | minor | Cohort membership, "eligible links", the default version and the as_of > C exit code are ambiguous, and mid-window pin changes are not disclosed |
| ADR012-R5-11 | minor | The declared change set is inaccurate: the card says "Remove", the Activation callout is undeclared, and the amendments paragraph omits trusted readiness publication |
| ADR012-R5-12 | minor | The `round-4.md` record is truncated: it ends at an empty heading with no verdict section |

## Closure of round-4 issues

"Fixed" means the required document fix, or its explicitly narrowed invariant replacement, is present in the signed bytes. It does not mean the implementation is validated. The lead and both peers agree on every row.

| Issue | Status | Evidence / remainder |
|---|---|---|
| ADR012-R4-01 | **Fixed** | REQ-22 `:589` reads "auto-demote that repo's check to report for that repo and record it [ASSUMPTION]", byte-identical to base `:589`. Only the counting and triage sentences were appended. |
| ADR012-R4-02 | **Fixed (invariant level)** | `ADR.md:26` requires protected-default-branch execution with reachable source, authorization at the current protected tip at publication, a recorded policy revision, matching original and executing actors, a stable repository-qualified action ID, replay as a no-op, and a stated ruleset residual. REQ-22 `:589`, REQ-26 `:611`, REQ-161 `:707`, `ci-integration.md:102` and `cli-commands.md:134` carry it. |
| ADR012-R4-03 | **Fixed as prescribed** | `ADR.md:28`, REQ-78 `:600`, `cli-commands.md:130` and `policy-file.md:82` carry the publishing-CAS formula, reassignment on retry, the non-decreasing frontier and default cutoff C. The council's own formula leaves publication ordering open (ADR012-R5-05) and only covers triage, not attempts (ADR012-R5-03). |
| ADR012-R4-04 | **Fixed** | `ADR.md:20` ("**Every query**…"), REQ-78 `:600`, `cli-commands.md:126` and `policy-file.md:80`. |
| ADR012-R4-05 | **Fixed** | `ADR.md:22` and REQ-27 `:612` set base=merge_base(B,H) and head=H tree OID; B and the assessed merge result are provenance. |
| ADR012-R4-06 | **Fixed** | `ADR.md:18`, REQ-27 and `ci-integration.md:101` use exactly the guardian's override inputs (none) and require a parity proof with override-marker PR text. |
| ADR012-R4-07 | **Fixed** | `ADR.md:38`, REQ-48 `:615` and `ci-integration.md:104` give separate remedies for a downgraded hybrid and a newer schema, rechecked on every CAS retry. |
| ADR012-R4-08 | **Fixed (invariant level)** | `ADR.md:40` and REQ-51 `:618` keep the integrity invariant and the proof obligation. The sandbox recipe is deliberately not selected. |
| ADR012-R4-09 | **Fixed** | `ADR.md:21,30,44`, REQ-78 and `cli-commands.md:128` disclose inspection-status counts and the inspected count in the provisional reason. |
| ADR012-R4-10 | **Fixed** | REQ-10 `:610` and `ADR.md:34`. |
| ADR012-R4-11 | **Fixed** | REQ-51 `:618`: "PR/fork jobs have read-only access to `scout/state` and never publish to it". |
| ADR012-R4-12 | **Fixed (adapted to the scope cut)** | Status lines `:3` of ADR-000/003/007/008, README `:18,21,25,26` (row 18 keeps ADR-001/002/003) and card relations `:736,807,901,925,1007` agree. The ADR body's own amendments paragraph lags behind them (ADR012-R5-11). |
| ADR012-R4-13 | **Fixed** | Callout `:449` now counts "comparable-base PR observations (distinct PR, head tree and merge-base)". It is not declared in the ADR's revision list (ADR012-R5-11). |
| ADR012-R4-14 | **Withdrawn by the scope cut** | REQ-168 `:714` is byte-identical to base, and no named workflow is selected. The governance leftover is ADR012-R5-01. |
| ADR012-R4-15 | **Fixed** | Plan `:15,35,61–62` and STATUS `:7`. |
| ADR012-R4-16 | **Fixed** | R-070 `:75` has the `workflow_dispatch`, contexts and `git merge-tree` links. R-077 `:76` has the repository facts, rounds 2–4 and REQ-26/162/163/168 attribution. Its `gate.go:382-384` citation matches the code (`wouldBlockID` at `:382–385`). `ADR.md:50` and card `:1010` cite both. |

## Does the narrowing keep what REQ-22, REQ-27 and REQ-78 need?

Leaving mechanics to the implementation PR is acceptable only if the ADR binds every invariant that PR must meet, testably. Most invariants are bound. Some were bound by the topology that the scope cut removed and were not restated.

| Requirement need | Bound today? | Where / gap |
|---|---|---|
| REQ-22: counting unit, deduplication, unknown ≤ 10 %, exact cohort | Yes | `ADR.md:20–21`; REQ-22/78 |
| REQ-22: maintainer-only, append-only, replay-safe triage | Yes | `ADR.md:26`; REQ-22/161 |
| REQ-22: zero `not_warranted` per exact version | **No** | A gap resets it (ADR012-R5-04); publication ordering (ADR012-R5-05) |
| REQ-22: 14-day clause | **Partly** | No endpoint, and the origin change is undeclared (ADR012-R5-08) |
| REQ-27: guardian-equivalent comparison and inputs | Yes | `ADR.md:17–18`; REQ-27 |
| REQ-27/51: trusted, protected comparison base | **No** | Lost with the round-2 eligibility rule (ADR012-R5-02) |
| REQ-27: failed or unverifiable assessment is comparison-none, not a gap | **Partly** | The boundary is unclear (ADR012-R5-09) |
| REQ-78: complete-interval accounting | **No** | No population (ADR012-R5-02) and no frontier advancement rule (ADR012-R5-03) |
| REQ-78: historical reproduction | **Partly** | Triage cutoffs are bound. Attempt facts (ADR012-R5-03), policy minima (ADR012-R5-07) and cohort or default version (ADR012-R5-10) are not. |
| Governance of the deferred wiring | **No** | ADR012-R5-01 |

## Issues

### ADR012-R5-01: The deferred wiring is itself a design change, with no review path (major)

- **Location:**
  - `ADR.md:12`: "This ADR does not select a trigger, reusable workflow, caller layout, artifact transport … policy workflow key or writer/scan commands. Those are implementation-PR decisions bounded by this contract, not additional product requirements."
  - REQ-160 `:706` lists exactly `pull_request`, default-branch `push`, nightly `schedule` and `workflow_dispatch` inputs `full`, `refresh-map`, `authorize-egress`.
  - REQ-161 `:707`: "This ADR selects no additional workflow, trigger, permissions or policy key."
  - REQ-162 `:708`: "both workflows". REQ-168 `:714` lists two reusable workflows.
  - `ci-integration.md:99`: "Existing workflows, grants and installation contract remain unchanged here."
  - Plan B4-S `:61`: "choose collection/triage wiring under its evidence invariants".
- **Problem:**
  - Meeting `ADR.md:16`, trusted collection that accounts for every eligible attempt, needs a trusted trigger beyond REQ-160's list, because REQ-51 forbids PR-lane publication [INFERENCE].
  - Meeting `ADR.md:26`, a provider-authenticated maintainer triage action, needs a new `workflow_dispatch` input, a new trigger or a new workflow [INFERENCE].
  - Each of those changes REQ-160, 161, 162 or 168. `CLAUDE.md` design rule 1 makes any change to the governed design subject to council sign-off.
  - As written, the ADR and the B4-S row invite the implementation PR to make that design change without review. Or, read strictly, they make the ADR unimplementable.
- **Required fix (recommended option: a scoped wiring ADR):**
  - Add one sentence to `ADR.md:12`, REQ-161, `ci-integration.md:99`, the plan B4-S row and STATUS `:121`. Wiring that changes any requirement is a design change. That covers:
    - a new or changed trigger, `workflow_dispatch` input, reusable workflow, job permission, policy key or release artifact;
    - any text of REQ-160–163 or REQ-168.
  - Such wiring needs its own ADR with a zero-issue council round before the implementation PR merges, and that ADR must meet ADR-012's invariants. Wiring that stays within the existing requirements needs no ADR.
  - In the plan, the B4-S dependency becomes "final adopted ADR-012 and, where its wiring changes requirements, an adopted wiring ADR".

### ADR012-R5-02: The evidence population is undefined, and the protected-base eligibility rule was lost (major)

- **Location:**
  - `ADR.md:16`: "account for every eligible attempt". "Eligible" is defined nowhere in the ADR, REQ-22/27/78 or the references.
  - `ADR.md:19` treats undelivered provider events as a sampling limitation.
  - Earlier rounds bound the population: `round-2.md:201–202` (ADR012-R2-01, fix 1) said only runs whose base ref is the default branch, and whose comparison base is reachable from the protected default branch, are eligible; others are `not_assessed` with reason `non_default_base`. Round 3 also relied on the removed topology for forks, bots and retargeted PRs (`round-3.md:340,343`).
- **Problem:**
  - Two conforming collectors give different PRRuns, TotalRuns and unknown ratios on the same history. One assesses on every PR event. The other is a nightly sweep that never sees PRs merged within a day.
  - Worse, a protected selector (for example a `readiness-sample` label) can certify complete intervals over a chosen subset and leave unknown-attribution PRs out of the population [INFERENCE]. Protected configuration authenticates the selector; it does not make the sample representative.
  - Without the default-base restriction, a PR into an unprotected branch controlled by a collaborator supplies a B that an attacker chose.
  - "Complete accounting" cannot be tested without a defined frame.
- **Required fix (recommended option: restore the reviewed population):** state in `ADR.md:16`, REQ-27, REQ-78 and `ci-integration.md:101`:
  - An eligible attempt is each PR event on which the pinned guardian PR lane runs, or would run, for a PR whose base ref is the protected default branch. Fork and bot PRs are included. Collector-initiated reassessments only add attempts to existing observations.
  - An attempt whose base ref is not the default branch, or whose B is not reachable from the protected default-branch tip, is `not_assessed` with reason `non_default_base`, counted and disclosed.
  - Eligibility never depends on PR content, labels, author choice or the assessment outcome.
  - Completeness is judged against the provider's enumeration of eligible events. An eligible event that is not accounted for makes its period a gap. Changing the collection selector starts a new interval.
  - Add proofs that an omitted failing PR prevents certification and that a non-default-base PR is `not_assessed`.

### ADR012-R5-03: Attempt facts are not publication-safe (major)

- **Location:**
  - `ADR.md:19`: "Default cutoff C is latest completely reconciled collectedThrough, which never decreases". "Completely reconciled" is undefined.
  - `ADR.md:17` records "assessment start" without naming a clock.
  - The no-rewrite guarantee is stated only for triage (`ADR.md:28`: "Publication after C cannot change a dossier at/before C"; REQ-78 `:600`).
  - Round 1's fix for ADR012-R1-04 required "terminal completeness through cutoff" (`round-1-response.md:20`). That wording did not survive the narrowing.
- **Problem:**
  - An attempt that started at or before C but is still running or queued when C is published, or whose finalize is delayed, or whose runner clock runs ahead, can land after C.
  - It then changes TotalRuns, the representative, the event union or the Days origin for `--as-of C`.
  - Historical reproduction (`ADR.md:44`, REQ-78) then fails for attempt facts even though triage is protected.
- **Required fix:** state in `ADR.md:19`, REQ-78 and `cli-commands.md:126`:
  - `started_at` is the provider-recorded attempt start. Interval bounds and collectedThrough use the same provider clock.
  - collectedThrough may advance to t only when every eligible attempt with `started_at` ≤ t is terminal (assessed, comparison-none or `not_assessed`) and published in that same state, or lies in a recorded gap.
  - Every later state holds exactly the same attempt facts and links with `started_at` ≤ C. A late fact for that range is refused and logged, not appended.
  - Add a proof that an attempt started at or before C₁ and arriving after the state carrying C₁ leaves the `as_of = C₁` result unchanged.

### ADR012-R5-04: A collection gap forgets earlier `not_warranted` verdicts (major)

- **Location:**
  - `ADR.md:20`: "use the complete interval containing as_of", with event links restricted to `interval.start ≤ started_at`.
  - `ADR.md:19`: an irrecoverable gap ends an interval.
  - REQ-22 `:589` requires zero `not warranted`, and `ADR.md:22` says "No later empty rerun can erase a not-warranted event".
  - Rounds 1–4 do not discuss this interaction (peer grep).
- **Problem:**
  - A gap caused by an outage or a failed publication does not change the rule version, yet it starts a new interval.
  - Suppose a check has a maintainer-confirmed `not_warranted` event in interval 1, then a gap, then 14 days and 5 clean observations in interval 2. It reads ready, although the same exact `(check_id, rule_version)` has a confirmed false positive.
  - Version bumps are the designed reset (REQ-22). A collection artifact should not be one. Combined with ADR012-R5-09, a PR-induced gap could launder a known false positive.
- **Required fix (recommended option):**
  - Evaluate zero-`not_warranted` over triage at `as_of` for every persisted CI event of the same repository and exact `(check_id, rule_version)` whose linked assessed attempt started at or before `as_of`, whatever its interval.
  - Days, PRRuns, TotalRuns, UnknownAttribution, WouldBlock and Undecided stay scoped to the containing interval.
  - Report a prior-interval `not_warranted` count.
  - Update `ADR.md:20,30`, REQ-22, REQ-78, `policy-file.md:80` and `cli-commands.md:128`, and add a proof using the gap fixture above.
  - Disclosure alone is the rejected alternative.

### ADR012-R5-05: `recorded_at` does not order successive triage publications (major)

- **Location:** `ADR.md:28` ("Current triage is latest authorized row by recorded_at then action ID"; recorded_at = max(writer UTC now, parent collectedThrough + 1 ms)); REQ-78 `:600`; `cli-commands.md:130`; `policy-file.md:82`.
- **Problem:**
  - While collectedThrough stays at 10:00, action A publishes `warranted` with recorded_at 10:02. Then action B, a later correction to `not_warranted`, runs on a runner whose clock is slower and publishes with recorded_at 10:01.
  - Both obey the formula. A, the earlier publication, stays "latest", so the correction is ignored and readiness can pass.
  - The formula protects cutoffs at or below the frontier. It does not order publications above it. "Replay … cannot overtake a newer verdict" (`ADR.md:26`) covers replays only.
  - This is the same harm class as R4-02 item 4 (major).
- **Required fix (recommended option):**
  - recorded_at = max(writer UTC now, parent collectedThrough + 1 ms, parent maximum recorded_at + 1 ms), recomputed on every CAS retry. Current triage is then the latest authorized row in publication order, and the cutoff property (recorded_at > C) still holds.
  - Readiness also reports how many in-window events have an authorized history at `as_of` containing a `not_warranted` row that a later row superseded.
  - Mirror this in REQ-78, `cli-commands.md:130` and `policy-file.md:82`.
  - Add a proof that two writers with skewed clocks publishing in sequence leave the later publication current.
  - A durable monotonic publication sequence, with recorded_at kept only for cutoff membership, is an acceptable alternative.

### ADR012-R5-06: Newly published triage is invisible until C advances (minor)

- **Location:** `ADR.md:28`: recorded_at ≥ parent C + 1 ms; "current_window uses as_of=C". Nothing requires C to advance.
- **Problem:**
  - A verdict published now is outside every `current_window` query until collectedThrough passes its recorded_at.
  - On a quiet repository with no new eligible attempts, Undecided can stay above zero indefinitely and wrongly hold readiness down.
  - "Current triage" (`ADR.md:28`) then differs from what the default query shows, and nothing says so.
- **Required fix:**
  - Require collectedThrough to advance whenever the ADR012-R5-03 completeness condition holds: at least on every triage publication and on a periodic trusted reconcile, at an implementation-chosen cadence.
  - Readiness reports the frontier age and the count of authorized triage rows with recorded_at > `as_of`.
  - Name the default view "triage at as_of".
  - Update `ADR.md:28,44`, REQ-78 and `cli-commands.md:130`.

### ADR012-R5-07: Historical reproduction does not fix the readiness policy (minor)

- **Location:** `ADR.md:44` (explicit `--as-of` is `historical_reproduction`; the output reports minima); REQ-78 `:600` (`min_days`/`min_pr_runs` are configurable); `policy.go:412–419` evaluates against the receiving policy's minima. Attempts record their own policy pins (`ADR.md:17`), but nothing selects the policy for a query.
- **Problem:** a dossier that was ready at T under 14/5 becomes "not ready" when the same `--version … --as-of T` query is evaluated under a later 30/10 policy, although every observation and triage row was reproduced exactly.
- **Required fix (recommended option):**
  - Each published collection frontier records the trusted base-policy revision and effective minima from the protected default-branch tip.
  - Readiness evaluates against the revision recorded at the latest frontier at or before `as_of`, for both `current_window` and `historical_reproduction`, and reports that revision.
  - Evaluation under any other policy is labelled policy re-evaluation, never `historical_reproduction`.
  - Update `ADR.md:44`, REQ-78 and `cli-commands.md:128,130`, and add a proof that changing minima after T leaves reproduction at T unchanged.

### ADR012-R5-08: Days has no endpoint, and its origin and the PRRuns exclusion are undeclared (minor)

- **Location:**
  - `ADR.md:21`: "Days is floor elapsed 24-hour periods since the first in-interval inspected comparable observation", with no endpoint. REQ-78 `:600` says "first qualifying in-interval start" without defining "qualifying".
  - Base `policy.go:387`: "Days since the first shadow run with a comparable base".
  - The amendments `ADR.md:48`, REQ-22 `:589` and card `:1010` declare only the no-eligible-unit exclusion. The `not_inspected` exclusion from PRRuns (`ADR.md:21`, `policy-file.md:78`) and the inspected-only Days origin are not declared.
- **Problem:**
  - An implementation could measure Days to the reader's wall clock, which breaks reproduction, or to the latest observation.
  - For `security.ci-mutable-action-ref`, an inspected-only origin starts the 14-day clock at the first PR that touches a workflow file. That changes the owner-decided 14-day clause in a way readers cannot see.
- **Required fix (recommended option: keep the council-settled origin and declare it):**
  - Days = floor((as_of − origin)/24 h) on the provider clock. The origin is the earliest `started_at` among in-interval represented comparable observations whose selected check is inspected. This is round 1's ADR012-R1-07 choice (`round-1-response.md:23`).
  - Declare both counting changes in `ADR.md:48`, REQ-22 and card `:1010`: the inspected-only Days origin, and the exclusion of `not_inspected` and `not_applicable` observations from PRRuns.
  - Changing the origin to include `not_applicable` is the alternative. It is not recommended because it reopens a settled choice.

### ADR012-R5-09: The boundary between a gap and comparison-none is unclear for missing output (minor)

- **Location:** `ADR.md:17` ("an assessment that starts but fails/times out/has invalid output, is assessed comparison-none") against `ADR.md:16` ("failed publication") and `:19` ("irrecoverable evidence is a gap"); REQ-27 `:612`; REQ-51 `:618`.
- **Problem:**
  - If a compromised or PR-influenced scanner can make its own output missing or irrecoverable, and that is classified as a gap, then one PR can end the interval for every check.
  - That resets Days and PRRuns and, under the current text, forgets earlier `not_warranted` verdicts (ADR012-R5-04).
  - `ADR.md:40` accepts only own-attempt falsification as residual risk, not interval resets.
- **Required fix:**
  - An authenticated eligible attempt that started but produced missing, invalid or irrecoverable output is assessed comparison-none.
  - Gaps arise only when the trusted collector cannot enumerate or attest attempts, or when its own publication fails.
  - State this in `ADR.md:16–19`, REQ-27 and `ci-integration.md:101`, and add a proof that a scanner deleting its own output yields comparison-none, not a gap.

### ADR012-R5-10: Cohort membership, eligible links, default version and the as_of exit code are ambiguous (minor)

- **Location:**
  - `ADR.md:20` does not limit representatives to attempts pinned to the cohort's rule version, and "eligible links" (`:22`) is undefined.
  - `ADR.md:44` says "latest CI-pinned catalogue", "future cutoff … exits 3" and "no complete containing interval is unmeasured", and "Report pin" is singular.
  - REQ-78 `:600`; `cli-commands.md:126,132`.
- **Problem:**
  - Suppose an observation has attempt 1 pinned to v1 and attempt 2, a rerun after an upgrade, pinned to v2. Whether it is represented in cohort v1 is undetermined.
  - "Latest" could mean the newest version or the most recently pinned one, which differ after a rollback.
  - Whether C < as_of ≤ now is "future" (exit 3) or "outside coverage" (exit 2) is ambiguous. Round 1 chose exit 2 for outside coverage (`round-1-response.md:20`).
  - Scout, tool or policy pin changes inside one window can reduce events without anyone seeing it.
- **Required fix:**
  - Cohort attempts are those whose pinned rule version for the check equals the cohort version. The representative is chosen among them. Eligible links are links from those attempts to events of that check and rule version.
  - The default version is the one pinned by the latest-started in-scope attempt at or before `as_of`.
  - as_of > C, but not beyond the reader's clock, exits 2. as_of beyond the reader's clock exits 3.
  - Report each distinct Scout, tool and policy-revision pin with its observation count.
  - Update `ADR.md:20,22,44`, REQ-78 and `cli-commands.md:126,132`.

### ADR012-R5-11: The declared change set is inaccurate (minor)

- **Location:**
  - Card Changes `:1010`: "Remove selected workflow/key/commands/transport/retention/install mechanics". Main never contained them.
  - `ADR.md:4` and card `:1010` omit the changed overview Activation callout `:449`.
  - `ADR.md:48` names counting, triage history, targeted migration and manifest-v2 export/merge. But the status lines `:3` of ADR-000/003/007, README `:18,21,25` and card relations `:736,807,901,1007` also name "trusted readiness publication" (REQ-26/51/161).
- **Problem:** the ADR summary and card misstate what this ADR changes relative to adopted main, and the authoritative ADR body is narrower than the status lines it justifies.
- **Required fix:**
  - Card: "Selects no workflow/key/commands/transport/retention/install mechanics; REQ-162/163/168 unchanged."
  - Add "overview Activation callout" to `ADR.md:4` and the card's Changes.
  - Add "and trusted readiness publication (REQ-26/51/161) for ADR-000/003/007" to `ADR.md:48`.

### ADR012-R5-12: The round-4 record is truncated and has no verdict (minor)

- **Location:** `round-4.md:299`, which is the last line: "## `pull_request_target` safety recheck", with no body. There is no "Peers", "Verdict" or memo section; compare `round-3.md:412–414`.
- **Problem:** `CLAUDE.md` design rule 2 requires every round to record the council's verdict. The issue summary (`round-4.md:65`) states the counts but not the verdict, and the file ends mid-structure.
- **Required fix (outside the signed set):**
  - Main restores the missing sections of `round-4.md` from the round-4 council output.
  - If that output is unavailable, Main appends a dated "Record correction" addendum after the existing bytes. It states the verdict ("ADR-012: 16 issues open (2 major, 14 minor). Not signed off."), notes that the `pull_request_target` recheck text was lost, and notes that the round-4 scope cut removed that topology.
  - Do not change the existing bytes.

## Checks that found no issue

- **Accepted ADRs.** ADR-000, ADR-003, ADR-007 and ADR-008 differ from base only on line 3 (lead: line 3 in both trees; challenger: all four bodies compared). ADR-011's card `:992–1003` and README row match base.
- **REQ-162/163** `:708–709` and **REQ-168** `:714` are byte-identical to base. No named shadow workflow, writer command, artifact transport or policy key remains as a requirement. `policy-file.md:84` mentions the superseded key only to forbid it.
- **Ledger.** R-064 and R-069 match base. R-070 and R-077 are this ADR's own rows. Every R-ID cited in `ADR.md:50` and card `:1010` exists. STATUS `:131` gives the next free ID as R-078. No reused ADR or R-ID was found.
- **Status-line format.** ADR-012's "accepted; … implementation pending" status with effectiveness on its final round matches ADR-011's precedent (`ADR-011 ADR.md:3`; README rows).
- **Triage cutoff arithmetic.** With collectedThrough non-decreasing and recorded_at ≥ parent C + 1 ms, a row absent from a state at C has recorded_at > C in every later state (lead check). The open defects are ordering (ADR012-R5-05) and liveness (ADR012-R5-06), not this property.
- **Exit codes** agree across `ADR.md:44`, REQ-78 and `cli-commands.md:132`, apart from the boundary in ADR012-R5-10.

## Considered and not raised

- **Resampling the representative with a no-op push (experiment-peer, minor).** The latest-terminal-attempt representative was the council's own choice in ADR012-R1-06, and R3-06 kept it. Pushing a trivially different tree has the same diluting effect and was accepted with `distinctPRs` disclosure (`round-3.md:338`). This is not a new interaction created by the scope cut.
- **Excluding cancellations, and counting per tree rather than per PR.** Both were settled (`round-3.md:338,341`) and are disclosed.
- **Changing the Days origin to include `not_applicable` (experiment-peer's recommended option).** The council keeps the inspected-only origin settled in ADR012-R1-07 and requires it to be declared (ADR012-R5-08).
- **A compromised scanner falsifying its own attempt.** This is an accepted residual risk (`ADR.md:40`). ADR012-R5-09 only stops it from causing interval resets.
- **`dossier_digest` binding the cutoff.** Once ADR012-R5-03, -07 and -10 are fixed, the readiness output carries the cutoff, interval, pins and policy revision, so the digest binds them. No separate change is needed.

## Peers and where they changed the outcome

- **`innovation-challenger`** raised 3 major findings and confirmed all 16 R4 closures. Adopted:
  - population and eligibility, which became ADR012-R5-02, merged with the peer's sampling frame. Its counterexample, a protected sample-selector label, and the citation of `round-2.md:201–202` turned the lead's "eligible is vague" note into a major issue;
  - publication ordering, which became ADR012-R5-05. The lead had accepted the council's own R4-03 formula as sound;
  - historical policy binding, which became ADR012-R5-07. The lead had not raised it.
- **`experiment-peer`** raised 13 findings (4 major) and confirmed all 16 R4 closures, with R4-14 withdrawn. Adopted:
  - wiring governance, which became ADR012-R5-01. The lead found this independently, and the peer's REQ-162 "both workflows" evidence strengthened it to major;
  - attempt publication-safety, which became ADR012-R5-03 and matches the lead's in-flight-attempt finding;
  - gap reset of `not_warranted`, which became ADR012-R5-04. The lead had not raised it;
  - triage liveness, which became ADR012-R5-06;
  - Days, which became ADR012-R5-08. The lead's endpoint point is kept, the peer's origin change is replaced with declaration after the lead found ADR012-R1-07;
  - the gap versus comparison-none boundary, which became ADR012-R5-09;
  - cohort membership and the exit boundary, which became ADR012-R5-10 (peer findings 9 and 11 merged);
  - the change-set declaration, which became ADR012-R5-11 and is merged with the lead's amendments-paragraph finding;
  - the round-4 truncation, which became ADR012-R5-12. The lead found this independently.
- **Not adopted:** peer finding 5 (resampling; see above). Peer finding 8 is merged into ADR012-R5-05 as the disclosure of superseded verdicts.

## Verdict

**ADR-012: 12 issues open (5 major, 7 minor). Not signed off.** Open: ADR012-R5-01, -02, -03, -04, -05, -06, -07, -08, -09, -10, -11 and -12. All 16 round-4 issues are closed: 15 are fixed and R4-14 is withdrawn by the scope cut, with its governance leftover carried as ADR012-R5-01. No digests are signed in this round.

## Council memo

1. **Recommendation.** Keep the narrowed, invariants-only ADR, and do not sign off this round. Fix the 12 issues with the recommended options above, then submit round 6 on the revised signed set, with Main supplying new digests.
   - The narrowing is sound in principle. The defects are invariants that the removed topology used to supply implicitly, plus a governance gap that the narrowing created:
     - the population and protected base;
     - terminal completeness at the frontier;
     - publication order;
     - the policy used for reproduction;
     - the gap boundary.
   - All fixes are text-level and add no mechanism selection.

2. **Alternatives.**
   - **(A) Fix within the narrowed ADR, and send requirement-changing wiring to its own ADR (recommended).** This keeps ADR-012 small and testable, and the wiring ADR can be reviewed against concrete workflow text. The cost is one more ADR before B4-S merges.
   - **(B) Re-expand ADR-012 to select the wiring.** Rounds 1–4 show that specifying the workflow before implementing it was the main source of defects (`round-4-response.md`, "Delegated scope cut"). This option also still needs changes to REQ-160/161/162. Rejected.
   - **(C) Sign off now and carry these items as implementation obligations.** This breaks the zero-issue rule (design rule 1). Items R5-02 to R5-05 change what counts as evidence for activation, so they are not implementation details. Rejected.
   - **(D) Drop interval scoping and use cumulative per-version windows.** This would remove R5-04 and part of R5-03, but it reopens the reviewed interval and gap design (ADR012-R1-03/05) and loses explicit unmeasured periods. Not recommended. R5-04's narrower fix keeps intervals and makes only zero-`not_warranted` cohort-wide.

3. **Evidence.** The issue sections above cite exact lines. The key ones:
   - `ADR.md:12,16,19,20,21,28,44,48`;
   - `scout-design.html:449,589,600,612,706–709,714,1010`;
   - `ci-integration.md:99`; plan `:61`;
   - `round-1-response.md:20,23`; `round-2.md:201–202`; `round-3.md:338–343`; `round-4.md:299`;
   - `policy.go:387,412–419`.

4. **Risks.**
   - **Overspecification.** The R5 fixes could drift back into mechanism. Each fix states only an observable invariant plus a proof, and the trigger, cadence and transport stay unselected.
   - **The wiring ADR becomes a bottleneck.** This is reversible: if the implementation fits within the existing requirements, no ADR is needed.
   - **Provider-clock assumptions** (R5-03): if the provider exposes no reliable attempt start time, the implementation must report a gap rather than certify. This is conservative.
   - **R5-04 is stricter:** a single old `not_warranted` blocks the exact version until a version bump or a maintainer re-triages it, which is append-only and visible. This is intended.
   - All changes are document-level and reversible before adoption. Nothing has been adopted.

5. **Execution outline.**
   1. Add the wiring-governance sentence (R5-01) to `ADR.md:12`, REQ-161, `ci-integration.md:99`, plan B4-S and STATUS `:121`, and update the B4-S dependency.
   2. Restate the evidence population and the `non_default_base` rule (R5-02) in `ADR.md:16`, REQ-27, REQ-78 and `ci-integration.md:101`.
   3. Define the provider clock, terminal-complete frontier advancement and refusal of late facts (R5-03) in `ADR.md:19`, REQ-78 and `cli-commands.md:126`.
   4. Make zero-`not_warranted` cohort-wide with a prior-interval count (R5-04) in `ADR.md:20,30`, REQ-22, REQ-78, `policy-file.md:80` and `cli-commands.md:128`.
   5. Change the recorded_at formula to include the parent's maximum recorded_at, and add the superseded-`not_warranted` count (R5-05), in `ADR.md:28`, REQ-78, `cli-commands.md:130` and `policy-file.md:82`.
   6. Add frontier liveness, frontier age and the rows-after-cutoff count (R5-06), plus the policy revision recorded per frontier (R5-07), in `ADR.md:28,44`, REQ-78 and `cli-commands.md:128–130`.
   7. Set the Days endpoint and declare the origin and the PRRuns exclusions (R5-08) in `ADR.md:21,48`, REQ-22, REQ-78 and card `:1010`.
   8. Set the gap versus comparison-none boundary (R5-09) in `ADR.md:16–19`, REQ-27 and `ci-integration.md:101`.
   9. Set cohort membership, eligible links, default version, the exit boundary and pin disclosure (R5-10) in `ADR.md:20,22,44`, REQ-78 and `cli-commands.md:126,132`.
   10. Correct the change-set declaration (R5-11) in the card, `ADR.md:4` and `ADR.md:48`.
   11. Main repairs `round-4.md` with an addendum only (R5-12).
   12. Add the new proof obligations to `ADR.md:52` and to the B4-S acceptance column. Each must be testable as a fixture:
       - an omitted failing PR;
       - a non-default base;
       - a late attempt;
       - a gap with an earlier `not_warranted`;
       - skewed-clock publications;
       - a quiet-repository triage;
       - a policy change after T;
       - scanner-deleted output.
   13. Write `round-5-response.md` mapping each issue. Main supplies the SHA-256 of all signed files and of `round-4.md`, then round 6 follows.
