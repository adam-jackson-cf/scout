# Proposal (DRAFT, pending innovation council sign-off): merge-time area association

Draft revision 2.8 of `docs/scout-design.md` (CLAUDE.md design rule 1: not adopted until the council signs off with zero issues; review rounds go in this directory as `round-<N>.md`). Sources: R-055, R-056, R-057 (`docs/research-ledger.md`); `evaluation/x2d-postrun-tuning/README.md`, `FROZEN-8-10.md`.

## Problem

After onboarding the agent map runs only on a manual `scout map --refresh` (Appendix 21). Files that merges add or move in between have no agent-derived area. Today they are listed as unowned and get a directory scope (§4 Unowned files); `internal/areas` only folds a whole agent map into areas (`Reconcile`, file-set Jaccard ≥ 0.5) and has nothing for single merges. A folder guess is the only cheap alternative, and it is poor: on 92 unseen fastapi window clusters (rounds 8–10) the directory baseline B2 assigned correctly 0.766 of the time with recall 0.636 (unresolved counted as a miss); on the 200 clusters of rounds 1–7 it was 0.452 / 0.420.

## Proposed flow (default-branch push runs only; never PR runs)

1. Deterministic lookups first (a file already owned by a heuristic feature or an active area keeps that owner). Files still unowned, grouped into one cluster per merged change, are the residual clusters.
2. For each cluster, Jev (pinned model, v1 questions) reads `merge.covered` ("does any listed area cover this change?") and then `merge.which_area` (a permuted choice over the whole catalogue). Policy thresholds: `which_area` confidence ≥ **0.535** assigns; `covered` answers `none` iff P(true) ≤ **0.40** (covered score threshold s = **0.60**). Permutation disagreement or a score below threshold is `unresolved`.
3. Every cluster Jev leaves `unresolved` or calls `none` goes to the reasoning-model lane: `codex exec -m gpt-6.1-terra -c model_reasoning_effort=high`, read-only sandbox, ephemeral, scrubbed environment, no tools. Prompt `terra_prompt_v1.txt` (sha256 `ca24fd76…5e76`), input = cluster, PR title/body, diff summary and the full catalogue. It answers `existing` (an area id) or `new` (title, summary, reason).
4. `existing` assigns the cluster to that area. `new` records a **pending new area** (title, summary, file set, reason). A pending area is a label only: it is exported in `areas.jsonl` marked pending, never re-keys an issue or a signal series (areas are labels, §4), and does not change heuristic scopes. A maintainer accepts or rejects it (`area acceptance` is a human decision, §6); the next manual agent map's `Reconcile` may match it by Jaccard like any area, otherwise it is kept pending.
5. Nothing is left unowned silently: a Terra error, timeout, unparsable answer or unknown area id after one retry leaves the cluster `unresolved (terra_failure)`, reported in the summary and `readiness`, never a folder guess.
6. Every fallback decision is logged as an `annotation` row (§6): cluster id, Jev scores and composed outcome, reason for escalation, CLI name and version, reported model and effort, prompt digest, inputs digest, the raw answer, retry count, latency. Jev calls keep the existing redacted request body and question version.

## Policy keys and defaults (from FROZEN-8-10)

`mapping.association.enabled` (follows consent), `.t_choice` 0.535, `.covered_score` 0.60, `.fallback.provider` `codex`, `.fallback.model` `gpt-6.1-terra`, `.fallback.effort` `high`, `.fallback.timeout` 900 s, `.fallback.max_per_run` (cap, value to be set in preregistration). Model and effort are pinned like Jev's model: an alias or a model mismatch reported by the CLI voids the answer. A change to a threshold, question version, prompt digest or either model makes prior decisions stale (§6 versioning).

## Lane placement

A new lane family in §3.4 terms: trigger `push`, mode `signal`, `requires: decision-egress` (Jev) and `model-cli` (fallback), optional (§7.3), so health never changes the exit code. Fork and PR runs: not run. No Jev key or egress: the whole lane is `skipped (reason)`. Jev available but no `codex` on `PATH` or the call fails: Jev-only result; `unresolved` and `none` clusters stay unowned, are counted and listed, and the lane reports `ran` with the shortfall, never a folder guess. The fallback is `codex` only (as the frozen run), not the `claude`→`codex` detection order. Cluster text sent to the fallback is the same class of content the Jev call already sends, redacted by the existing sanitizer (not an injection defence; the prompt treats it as data and the output is only a label).

## Cost and latency (README, descriptive)

Jev ≈ $0.0013 per cluster in rounds 8–10 ($0.1178 / 92; $0.0014 over rounds 1–7). About 27 of 100 clusters escalate (25 of 92 in rounds 8–10; 54 of 200 in rounds 1–7). Terra latency per call: median 4.3 s, p90 4.9 s, max 5.9 s in rounds 8–10 (median 5.5 s in rounds 1–7). Terra's own cost is on the user's Codex subscription and was not measured in dollars. A push with 30 residual clusters would add about 8 calls, roughly 40 s sequential.

## Evidence and its limits

Rounds 8–10, 92 unseen clusters, one run, settings frozen and committed first (`c55a40d`): Jev+Terra P_assign 0.922 [0.840, 0.964] (71/77), recall 0.922, unresolved 0; new-area P / R 0.867 / 0.867 (13/15). Jev-only: P 0.955, recall 0.831, unresolved 0.109; B2 0.766 / 0.636. Terra was right on 20 of 25 escalations, 0 failed calls. Rounds 1–7 (where the prompt was iterated): Jev+Terra P_assign 0.899, recall 0.856, none P 0.318.

Limits: one repository (fastapi, Python); labels are agent-made (`agent_reviewed`); thresholds and prompt were developed on rounds 1–7 and the one test is 92 clusters (wide intervals; the new-area figure rests on 15 positives); this was a development run under no preregistration and no council review, so HD6 stays `inconclusive` and nothing here is a tested hypothesis. The axum replication (a second repository and language) is in progress and is not yet evidence. The figures are engineering observations supporting a provisional design, not a validated default.

## Risks

- **Spurious new areas.** In rounds 6–7 Terra answered `new` for 12 of 14 `none` answers that were gold-covered clusters of one PR (streaming / server-sent events) where labellers accept an existing area; rounds 8–10 had 2 gold-covered clusters answered `none`. Mitigation: pending status, maintainer acceptance, log of the reason; no identity impact.
- **Near-duplicate areas (HD7).** Catalogue duplicates split Jev's choice; cross-scope duplicate checking is unvalidated (R-054).
- **Provider drift.** Jev, `gpt-6.1-terra` and the `codex` CLI change under pins; thresholds were tuned on one Jev version. Mitigation: pins, staleness disclosure, re-measure on version change.
- **Cost / latency** per push, bounded by the cap; the lane is optional.
- **Prompt overfit** to fastapi rounds 1–7; v2 of the prompt was worse, v1 kept.
- **Per-decision minimums:** `merge.which_area` was `tuned-provisional` (held-out below minimums); `merge.covered` is untunable at the skill's minimums (13 gold-none of 200). The thresholds here are chosen policy, not a tuned and validated operating point.

## Interactions with Appendix decisions

- Appendix 21 (no automatic agent-map trigger): the Clawpatch agent map is untouched; this flow is not a map refresh and re-labels no existing area. It adds areas only as pending. See question 1.
- Appendix 4 / 19 (Jev advisory; thresholds gate nothing): here two thresholds route a cluster to the fallback or leave it unowned, so they gate behaviour of a label. §6's rule then applies (stale thresholds disclosed; B-3 flip-rate harness on this choice). See question 2.
- Appendix 20 / 8: the fallback is a reasoning lane through `codex exec` only; the agent map's `gpt-6-luna`/`xhigh` pin is unchanged.
- X3 (§13) already covers "unowned file → area suggestion" by Jev. This proposal supersedes X3's file-level shape by a cluster-level one; X3's 0.9 precision rule is not changed and X3 is not run.

## Open questions for the user

1. Does merge-time pending-area creation conflict with Appendix 21's intent (areas change only at onboarding and manual maps)? If yes, pending areas are not created and `new` clusters stay unowned and listed until the next manual map.
2. Is Jev (and a reasoning model) assigning a label on merges within Appendix 4's "route/rank only", given it changes neither issue state nor the gate?
3. Pending areas: maintainer accept/reject, or auto-accept after N merges? (The draft assumes maintainer.)
4. Source egress: is the existing `decision-egress` consent enough for the fallback, or does it need the stronger `authorize-egress` that the agent map uses?
5. Fallback provider: `codex` only (as measured), or also `claude` where only it is present? `claude` is unmeasured.
6. Wait for the axum replication before council review, or review now as provisional?
7. Per-push cap and maximum pending areas per run.
8. Design rule 3 (CLAUDE.md) lets design changes cite only signed-off experiments or inertia-bench studies. R-056/R-057 are development runs without council sign-off, so this draft cites them only to describe the proposal and its limits, and makes X11 the validating experiment. Should adoption wait for X11 (default off until then), or does the user accept the development evidence for a provisional, opt-in launch?

## User decisions (2026-10-31, R-059)

- Q1 new features: create a **pending area** at merge time (Appendix 21 reopened for this case only).
- Q3 acceptance: a **maintainer accepts** (or merges) each pending area.
- Q8 adoption: **provisional opt-in now**, lane off by default; X11 upgrades it later.
- Q5 fallback: **`codex exec` gpt-6.1-terra, effort high, only**.
- Questions 2, 4, 6 and 7 remain open for the council to settle or return to the user.

## Axum replication (R-058), received after these decisions

Zero-tuning replication on axum: Jev+Terra P_assign 0.622 [0.514, 0.719], recall 0.810, unresolved 0 (fastapi 0.922 / 0.922 / 0). The loss is new features forced into existing areas (23 of 31 errors; axum window 42 % gold-none against fastapi 16 %, catalogue 0–4 years stale). The design must state that precision depends on catalogue freshness.
