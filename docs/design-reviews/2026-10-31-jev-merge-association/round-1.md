# Design review round 1 — DRAFT revision 2.8: Jev merge-time area association with a reasoning-model fallback

Date: 2026-10-31. Rules: CLAUDE.md design rules 1–3. Reviewed by the innovation council: lead analysis, plus independent peers `innovation-challenger` (GPT, adversarial) and `experiment-peer` (Opus, evidence and confounds). The review was read-only. It ran no shell commands and no model calls, and it did not touch `.codegraph`.

Reviewed files (SHA-256): `docs/scout-design.md` `62dde5d309693c232491e04a8ec1852da305d891f3795634834839c1bec77586`; `docs/design-reviews/2026-10-31-jev-merge-association/proposal.md` `56e823eb2b6325db94a8500267d97ad7cca330ed5c4a06c1440e45a91f0bc17d`

## Inputs reviewed

- `docs/scout-design.md`: status line (:5), the rev 2.8 paragraph (:19), §4 merge-time association, evidence and refresh guard (:114–116), §6 rows (:164–166), §12 R18 (:427), §13 X11 (:444), §14 A14 (:463), Appendix 21 (:498) and the new item 23 (:500). For contradictions we also checked Appendix items 1–22, §3.4 (:102), §6 adapter rules (:173–184) and §7.2 (:200).
- `docs/design-reviews/2026-10-31-jev-merge-association/proposal.md`.
- `evaluation/x2d-postrun-tuning/README.md`, `FROZEN-8-10.md`, `axum/README.md`, `axum/FROZEN-AXUM.md`, and their referenced files: `axum/pilot/{gold.py, frontend.py, GUIDE.md, label_x2d.py, questions/merge.which_area@v1.json}`.
- `docs/research-ledger.md` R-055 to R-059; `docs/hypothesis-ledger.md` HD6, HD6b, HD7.
- Code: `internal/areas/areas.go`, `internal/state/{annotations.go, areas.go, export.go, store.go, mapping_refresh.go}`, `internal/domain/gate.go`, `internal/decisions/question.go`, `internal/modelcli/modelcli.go`, `docs/jev-adapter-audit.md` (row 30, B-3).

Settled user decisions (R-059), which this review does not reopen: a pending area is created at merge time; a maintainer must accept it; the lane is a provisional opt-in, off by default; the development evidence is accepted as an exception to design rule 3; the fallback is `codex exec gpt-6.1-terra` at high effort only.

## Verification table

| Claim in the draft | Evidence | Result |
|---|---|---|
| fastapi rounds 8–10 Jev+Terra P_assign 0.922 [0.840, 0.964] (71/77), recall 0.922, unresolved 0, new-area P/R 0.867 (13/15) | `README.md` rounds 8–10 table, row "Jev+Terra 8-10" | correct |
| fastapi Jev-only 0.955 / 0.831 / 0.109; B2 0.766 / 0.636 | same table | correct |
| fastapi escalations 25/92 (27.2 per 100), Terra right on 20/25, 0 failed calls, latency 4.3 / 4.9 / 5.9 s, Jev $0.1178 | `README.md` rounds 8–10 bullets | correct |
| Rounds 1–7 P_assign 0.899, recall 0.856; 12 of 14 rounds 6–7 `none` answers gold-covered | `README.md:113,116` | correct |
| axum Jev+Terra 0.622 [0.514, 0.719] (51/82), resolved recall 0.810, new-area P 0.852 (23/27), R 0.500 (23/46); Jev-only 0.610; B2 0.556, overlapping | `axum/README.md` three-arm table | correct |
| 23 of 31 axum errors on gold-`none`; 46/109 = 42 % gold-`none` against fastapi 15/92 = 16 % | `axum/README.md` "Wrong assignments by cause"; `:10`; `README.md` rounds 8–10 | numbers correct; reading wrong (R1-08) |
| axum gold-`none` are "new features" | `axum/README.md:10` "none kinds: 37 gap, 9 new"; `GUIDE.md:22-23` defines gap as a responsibility that already existed | **wrong** (R1-08) |
| Catalogue 0–4 years before its clusters | C0 2022-12-19; round 5 ends 2026-05-16 (`FROZEN-AXUM.md`) | **wrong**: up to about 3.4 years (R1-08) |
| axum escalations 29 per 100, Jev ≈ $0.0009 per cluster, fallback median 5.6 s; "25 per 100 `none` outputs" | `axum/README.md:11-12`; 27 `none` outputs / 109 | correct |
| Thresholds t_choice 0.535; `none` iff P(true) ≤ 0.40 (s = 0.60); Terra argv, prompt sha `ca24fd76…5e76`, 900 s | `FROZEN-8-10.md` "Jev", "Terra"; `FROZEN-AXUM.md` | correct; the draft omits parts of the measured setting (R1-04) |
| Measured unit is "one cluster per merged change" | `axum/pilot/frontend.py` `clusters` (key = normalised parent directory, variant stem); `README.md` 92 units / 41 PRs | **wrong** (R1-04) |
| Same frozen settings measure the product flow | `README.md:19` and `pilot/gold.py:7`: membership teacher-forced from gold, gold `none` minted no area | **not disclosed** (R1-09) |
| Areas are labels only; `Reconcile` matches by Jaccard | `internal/areas/areas.go:1-7,40-51,85-90` | correct; `Area` has no pending state and `Active()` means only "not retired" (R1-02) |
| Decision log as `annotation` rows | `internal/state/annotations.go:20` (`issue_id TEXT NOT NULL`); `export.go` exports no annotations | **unimplementable as stated** (R1-03) |
| `LastAgentRefreshAt` available to the guard | `internal/state/mapping_refresh.go:19`; exported (`export.go:134`) | correct |
| One `requires` per lane | `internal/domain/gate.go:26-33,53` (scalar `LaneRequires`) | proposal's double `requires` cannot be expressed (R1-01) |
| No Jev key or egress → "Jev-only" (§4) and → "skipped" (proposal) | `scout-design.md:114` against `proposal.md` "Lane placement" | **contradiction** (R1-01) |
| Whole-catalogue Choice | `internal/decisions/question.go:73-74` (2–255 options) | unhandled outside that range (R1-04) |
| Catalogue in the `which_area` version key | `questions/merge.which_area@v1.json` "the catalogue digest is part of the version key"; design :173 | staleness rule misfires (R1-05) |
| Model-CLI subprocess scrubs the Jev key | `internal/modelcli/modelcli.go:65-84` (drops `OPENROUTER_*`, `*API_KEY*`, `*SECRET*`) | correct |
| PR and fork runs read-only, no secrets; Jev key only on trusted runs | design :184, :200, Appendix 8, 17; the lane is push-only | correct |
| Model-CLI consent | design :184 "model-CLI lanes capture their own consent" | the draft's "`decision-egress` covers Terra" contradicts it (R1-07) |
| B-3 before default-on | Appendix 19 (:496) "B-3 … on the first Choice that gates anything"; `jev-adapter-audit.md:85` | **contradiction** (R1-06) |
| Agent-map provider pin and Appendix 22(c) | Appendix 22(c): one `claude`→`codex` order, opted-in lanes run the policy provider | fallback exception not stated (R1-01) |
| Guard defaults marked [ASSUMPTION], written after the axum figures were known | `:116`, `:463`, `proposal.md` "Refresh guard" | disclosed; the comparison it gives is incomplete (R1-10) |
| `max_per_run` 40 ≈ "1.5 × the clusters of an average fastapi push" | about 2.2 clusters per merged PR (92/41; 200 in rounds 1–7) | **wrong** (R1-11) |
| Proposal B2 rounds 1–7 "0.452 / 0.420 (unresolved counted as a miss)" | `README.md:114`: 0.420 is resolved recall (71/169); abstention-as-miss = 71/187 = 0.380 | **wrong** (R1-11) |
| User accepts R-057 and R-058 as the design rule 3 exception | R-059 limits column: "Taken before the axum replication result (R-058) was reported" | timing not disclosed (R1-12) |
| HD6 stays `inconclusive` | `hypothesis-ledger.md:39` | correct |
| Appendix 21 reopened only for pending areas; agent-map schedule and Appendix 20 pins unchanged | :498, :500 | correct |
| Appendix 4 (Jev never changes issue state or the gate) | the lane sets labels only; areas never enter identity | holds |
| Other Appendix items 1–3, 5–18 | checked | no contradiction |

## Issues

### R1-01 (high): lane prerequisites and failure handling are contradictory or incomplete

`scout-design.md:114` says "where no Jev key or egress, no `codex` on `PATH` or a failed call, the result is Jev-only". Without Jev there is no Jev result. The proposal says the opposite ("No Jev key or egress: the whole lane is `skipped (reason)`").

The proposal's "`requires: decision-egress` (Jev) and `model-cli` (fallback)" cannot be expressed: `requires` is one scalar per lane (`internal/domain/gate.go:26-33`). The draft also misses three things:
- the measured rule's voiding of command-running sessions (`FROZEN-8-10.md` escalation rule; `-s read-only` alone does not enforce "no tools");
- Jev failures and the shared `policy.decisions` caps;
- Appendix 22(c): "an opted-in lane runs the policy provider".

**Fix.** In `scout-design.md:114`, replace from "A Terra error, timeout, unparsable answer…" to "…never guessed from a folder." with the text below, and make the proposal's "Lane placement" match:

> The flow is two optional push-only `signal` lanes (§3.4 allows one `requires` per lane): `association` (`requires: decision-egress`, Jev) and `association-fallback` (`requires: model-cli`; it runs only on clusters `association` produced in the same run). With `mapping.association.enabled` false, no Jev key, `decision-egress` not acknowledged, or egress not allowed for the trigger, both lanes report `skipped (reason)`, no model is called, and residual files keep their directory scope as in revision 2.7. A Jev call failure, a Jev model mismatch or the `policy.decisions` call or cost cap leaves a cluster `unresolved (jev_failure)` or `unresolved (jev_cap)`. Such a cluster is not escalated; it is counted and listed. Later runs decide only their own push's clusters. If `association` ran but there is no `codex` on `PATH` (the normal case on GitHub-hosted runners, R13), or the fallback lane's consent is absent, `association-fallback` reports `skipped (no model CLI)` or `skipped (no consent)`, and the clusters Jev left `unresolved` or `none` stay unowned, counted and listed. The fallback looks for `codex` only and never runs `claude`. Like the agent map, this is an exception to Appendix 22(c): the policy provider is not read. After one retry, any of the following leaves the cluster `unresolved (terra_failure)`: a Terra error, a timeout, a session that ran any command, a CLI-reported model or effort other than `gpt-6.1-terra` / `high`, an unparsable answer, or an unknown area id. Clusters beyond `fallback.max_per_run` are `unresolved (fallback_cap)`. Each lane's health follows §6 Lane health: `ran` with counts, `failed` only when every attempted call fails. Neither lane changes the exit code. No path guesses an area from a folder.

In Appendix 23(d), append: "…; the fallback finds `codex` only and is an exception to item 22(c), as the agent map is."

### R1-02 (high): the pending-area and assignment state model does not fit `internal/areas`

The draft says pending areas are "never created or accepted by a model" (`:114`), yet a fallback `new` answer creates them. It also says "the next manual agent map's `Reconcile` may match it by Jaccard like any area". The draft does not say:
- where an `existing` assignment is stored;
- whether pending areas are offered to Jev or the fallback.

Today `Area` has no status, `Active()` is true for anything not retired, and `Reconcile` overwrites a matched area's file set and title and retires an unmatched area whose files are all heuristic-owned (`internal/areas/areas.go:40-51,85-90`). Reusing these semantics unchanged would let an unaccepted proposal own files through deterministic lookups, be matched, or be retired without a maintainer. It would also put agent-map discoveries behind maintainer acceptance, which rev 2.7 does not do. Neither run measured pending areas in the catalogue: gold `none` "minted" nothing (`axum/pilot/gold.py:7`).

**Fix.** In `:114`, replace "never created or accepted by a model" with "recorded on a fallback `new` answer and never accepted by a model". After "the next manual agent map's `Reconcile` may match it by Jaccard like any area", replace that clause with:

> State: an `existing` assignment is an association record (cluster id, files, area id, decision). The files' area label comes from that record; the files are not written into the area's `FileSet`, so `Reconcile` and its Jaccard (A1) read agent-map file sets only. A pending area is an `areas` record with `status: pending` and its provenance (cluster id, decision id), stored in the `areas` payload and exported in `areas.jsonl`. `Area.Active()` is false for a pending area, so no deterministic lookup treats its files as owned. Pending areas are not offered to Jev or the fallback; this is the measured setting, since no area was minted between rounds (`pilot/gold.py:7`). One new concern can therefore produce several pending areas; the summary lists pending areas with their files for a maintainer. `Reconcile` neither matches nor retires a pending area. After a completed agent map:
> - a pending area whose files are all owned by an active area or a heuristic feature is marked `superseded` and listed; it is never accepted or deleted automatically;
> - an association record whose files the map now owns is superseded by the map.
>
> Only a maintainer's accept, or merge into an existing area, changes a pending area's status. An accepted area becomes active with origin `association`, and from then on `Reconcile` matches it like any area.

Mirror this in the proposal's flow step 4 and in Appendix 23(a).

### R1-03 (high): the decision log and the guard's history cannot persist as designed

"Every fallback decision is an `annotation` row" (`:114`) does not fit the schema: `annotations` rows require `issue_id TEXT NOT NULL` (`internal/state/annotations.go:20`), and an association decision has no issue. Annotations are also not in the export (`internal/state/export.go`; design §7.2 :200). A trusted default-branch run on a hosted runner starts from the published `scout/state`, so the log, the guard's 20-merge window and the `catalogue-stale` flag would be lost between runs.

A new table also changes the schema. The store refuses every `user_version` except 12 and the v1 re-anchor (`internal/state/store.go:25,82,255`), which matches Appendix 22(a).

**Fix.** In `:114`, replace "Every fallback decision is an `annotation` row with…" with:

> Every association decision (Jev-only or with fallback) is a row of a new `associations` table, one per cluster per run, holding: merge commit, cluster key and files, Jev scores and composed outcome, escalation reason, CLI name and version, reported model and effort, prompt digest, inputs digest, catalogue digest, raw answer, retries, latency, outcome and `stale`. It is exported as `state/v1/associations.jsonl`, which is added to the §7.2 export list. Jev request bodies are stored redacted as §6 requires. The refresh guard's window and the `catalogue-stale` flag are computed from these exported rows and the exported `LastAgentRefreshAt`; no other guard state is stored.

In §3.3, the current schema becomes `user_version` 13. Under Appendix 22(a), unchanged, a store at 12 is refused with the existing re-baseline message. Extend R17 and A13 to name 12, and add a Phase test: a 12 store is refused, and `scout state restore` round-trips `associations.jsonl` and pending `areas.jsonl` records.

### R1-04 (high): the specified flow is not the measured one

"groups them into one cluster per merged change" (`:114`; proposal step 1) is wrong. The measured unit is several clusters per change: residual files keyed by normalised parent directory and variant stem (`axum/pilot/frontend.py` `clusters`; fastapi rounds 8–10 have 92 clusters in 41 PRs).

The draft also leaves out other parts of the measured setting:
- the residual rules applied before clustering (`frontend.py` `classify`; renames and variant siblings inherit, `gold.py:7-8`);
- the cap of at most 10 scored clusters per PR (`FROZEN-AXUM.md`);
- the input settings: 60 excerpt lines and 5 sample files per area (`FROZEN-8-10.md`, `FROZEN-AXUM.md`);
- the second permutation, asked for a 20 % hash-selected subset only (`questions/merge.which_area@v1.json` `second_permutation`);
- the dependency on PR metadata (Appendix 17).

The Choice also accepts only 2–255 options (`internal/decisions/question.go:73-74`). Implementing the draft as written would deploy a decision unit and inputs that were never measured.

**Fix.** In `:114`, replace "takes the files the push added or moved that deterministic lookups leave unowned (no heuristic feature or active area owns them), groups them into one cluster per merged change, and" with:

> takes, for each first-parent merge in the push, the files it added or moved. It applies the measured catalogue-free residual rules (`evaluation/x2d-postrun-tuning/axum/pilot/frontend.py` `classify`): a heuristic feature or active area already owns the file, a rename keeps its owner, a new variant sibling inherits its sibling's owner, and empty package markers are inert. It groups the remaining files by normalised parent directory and variant stem (`frontend.py` `clusters`), so one merge can produce several clusters (fastapi: 92 clusters in 41 PRs). At most 10 clusters per merge are decided, hash-sampled as measured; the rest are `unresolved (pr_cap)`. Jev and the fallback get the measured state:
> - unit paths, change kind, role, and the first 60 lines of each member;
> - PR title, body and changed paths from `--fetch-pr-metadata` (Appendix 17). When these are absent, the `pr` fields are empty and the decision records `pr_metadata: absent`, an unmeasured condition;
> - the diff summary;
> - every active area with id, title, summary and at most 5 sample files.
>
> `merge.which_area` is asked in a second permutation for the measured 20 % hash-selected subset only, and disagreement there is `unresolved`. With fewer than 2 or more than 255 active areas (the Choice limit, `internal/decisions/question.go:73`), `association` reports `skipped (catalogue size)`. Then it

Use the same wording in proposal step 1.

### R1-05 (medium): staleness misfires on every catalogue change, and stale labels are undefined

The draft says "a threshold, question, prompt or model change marks prior decisions stale" (`:114`). But the `merge.which_area` version key contains the catalogue digest (`questions/merge.which_area@v1.json`; design :173 hashes `criteria`). So every agent map, accepted pending area or merge would mark every earlier association decision stale. The draft also does not say what a stale decision does to its label.

**Fix.** In `:114`, replace "a threshold, question, prompt or model change marks prior decisions stale" with:

> each association record stores the catalogue digest it was decided against. A catalogue change alone does not make a record stale. A change to a threshold, the question text, the prompt digest, the Jev pin or the fallback model marks prior records `stale`. A stale record keeps its label, is shown `stale`, and is never re-decided automatically.

### R1-06 (medium): B-3 is deferred past what Appendix 19 requires

`:164` says "the B-3 flip-rate harness runs on `merge.which_area` before default-on". Appendix 19 (:496) runs B-3 "on the first Choice that gates anything". The draft concedes that this Choice's threshold gates escalation and the label applied, and Appendix 23 says item 19 is "otherwise unchanged".

**Fix.** In `:164`, replace "and the B-3 flip-rate harness runs on `merge.which_area` before default-on" with "and, by Appendix 19 (B-3), the `orderflip` harness (`docs/jev-adapter-audit.md` row 30) runs on `merge.which_area` and its flip rate is recorded before a release lets the opt-in lane be enabled; the 20 % second permutation does not replace it". Make proposal Q2 match.

### R1-07 (medium): the consent assumption contradicts §6

`:114` says "the existing `decision-egress` acknowledgement covers the cluster text sent to Terra… [ASSUMPTION; the council or user may require the stronger `authorize-egress`]". §6 (:184) already decides this: "model-CLI lanes capture their own consent". Terra is a second recipient (OpenAI through `codex`), and it receives source excerpts plus the PR title and body. `authorize-egress` is Clawpatch's `--allow-egress` for the agent map (:109), and Clawpatch is not invoked here.

**Fix.** Replace that sentence with:

> Consent: `association` needs the `decision-egress` acknowledgement, as every Jev use does. `association-fallback` captures its own model-CLI consent at `scout init`, as every model-CLI lane does (§6), and that consent names `codex` / OpenAI and the content sent (file excerpts, PR title and body, catalogue). `authorize-egress` does not apply, because Clawpatch is not invoked.

Remove "and the `decision-egress` consent sufficing for the fallback" from A14. Mark proposal Q4 resolved by §6.

### R1-08 (high): gold-`none` is read as "new feature", and staleness is asserted against the data

`:115`, `:427`, `:444` and `proposal.md` "Evidence"/"Risks" call axum's gold-`none` clusters new features. In fact 37 of the 46 are `gap`: the responsibility existed before the change, but no catalogue area describes it (`axum/README.md:10`; `GUIDE.md:22-23`). Only 9 are `new`.

The draft states freshness as the mechanism in R18, A14 and the guard rationale ("precision depends on … catalogue freshness"; proposal "Because precision depends on catalogue freshness"). The data in hand weigh against age as the sole cause. Round 1 ends 2023-03-28, 0–3 months after C0 (`FROZEN-AXUM.md`), and already has 13 of 23 gold-`none`. Per round the counts are 13/23, 5/20, 3/22, 15/24 and 10/20, with no rise over time (`axum/README.md:10`). Clusters run up to about 3.4 years after C0, not 4.

**Fix.**
- **`:115`.** Replace "(a new feature forced into an existing area)" with "(a cluster no catalogue area describes, assigned to an existing area; of the 46 gold-`none`, 37 are `gap`, a responsibility that existed before the change, and 9 are `new`)". Replace from "That the axum catalogue (built at 2022-12-19, 0–4 years before its clusters) was the cause is **[INFERENCE]**; it was not tested." with: "Whether catalogue age caused it is **[INFERENCE]**, and the data weigh against age as the sole cause: round 1, 0–3 months after the catalogue, already has 13 of 23 gold-`none`, and the share does not rise with age (13/23, 5/20, 3/22, 15/24, 10/20). An incomplete onboarding catalogue (37 of 46 `gap`; 179 files unowned at C0) is an equally open explanation. Clusters run 0–3.4 years after the catalogue."
- **`:427` (R18).** Replace "because new features were forced into existing areas (23 of 31 errors; 42 % of axum clusters gold-`none` against 16 %; staleness as cause is **[INFERENCE]**" with "because clusters no catalogue area covers were assigned to existing areas (23 of 31 errors; 42 % of axum clusters gold-`none`, 37 `gap` / 9 `new`, against 16 %; catalogue age or incompleteness as the cause is **[INFERENCE]**, untested". Replace "when the catalogue is stale or thin" with "when much of the change is not covered by the catalogue".
- **`:463` (A14).** Replace "precision depends on the new-feature share and catalogue freshness" with "precision depends on the gold-`none` share; whether catalogue age or an incomplete onboarding catalogue drives it is untested **[INFERENCE]**".
- **`:116`.** Append: "If the cause is an incomplete onboarding catalogue, a refresh with the same mapper may not restore precision (X11 tests it)."
- **Proposal and ledger.** Apply the same wording in `proposal.md`. Correct R-058's "(new features)" to "(37 gap, 9 new)" with a dated note.

### R1-09 (medium): teacher-forced membership is not disclosed

Both runs started each round from the previous round's gold (`README.md:19`; `FROZEN-AXUM.md` "teacher-forced"), and gold `none` minted no area (`pilot/gold.py:7`). The product instead feeds its own assignments forward and keeps pending areas out of the catalogue (R1-02). So the reported figures do not measure error build-up across merges.

**Fix.** Append to `:115`, before "Costs:":

> Both runs teacher-forced membership: each round began from the previous round's gold, and gold `none` minted no area. The product feeds its own decisions forward, so error accumulation over merges and duplicate pending areas are unmeasured.

Add the same sentence to A14 and to the proposal's "Evidence and its limits".

### R1-10 (high): the refresh-guard disclosure is incomplete, and its rates are undefined on the failure and cap paths

`:116` compares the limits with pooled axum rates, but the guard uses a trailing 20-merge window. At round granularity the guard would also fire on fastapi **[INFERENCE: no per-merge replay was run]**. Rounds 8–9 (21 PRs, 44 clusters) give a fallback rate of 20/44 = 0.45 and a pending-new rate of 14/44 = 0.32, on the run that scored P_assign 0.922 (`README.md` escalation bullet; Jev+Terra `none` outputs in rounds 8 and 9).

The pooled escalation rates do not separate the repositories (29.4 against 27.2 per 100). Neither rate sees axum's main error. Of the 31 wrong Jev+Terra assignments, 30 are Jev's own assignments, which are never escalated: Jev-only has 30 wrong (22 gold-`none` + 8), and Jev+Terra has 31 (23 + 8) (`axum/README.md` "Wrong assignments by cause").

Separately, the draft does not say how clusters left unresolved by a cap, a skip or a failure count, so the rates would fall exactly when answers are missing. The "evaluated only once the window holds at least 20 clusters" minimum is not a policy key, although the proposal says "All four numbers are policy keys".

**Fix.** In `:116` (and in the matching proposal text), replace from "at these defaults axum's escalation rate would not trip…" to "…an observation, not a validation." with:

> Compared at round granularity with the development data **[INFERENCE: no per-merge replay was run]**, it also fires on fastapi: rounds 8–9 (21 PRs, 44 clusters) give a fallback rate of 0.45 (20/44) and a pending-new rate of 0.32 (14/44), on the run that scored P_assign 0.922. Pooled escalation rates do not separate the repositories (29.4 against 27.2 per 100). Neither rate sees axum's main error: 30 of the 31 wrong assignments are Jev assignments, which are never escalated. Expect `catalogue-stale` flags on healthy catalogues. The guard's recall and false-fire rate are unknown (X11).

After the rate definitions, insert:

> The fallback-rate numerator counts clusters Jev left `unresolved` or `none`, whether or not the fallback ran. The pending-new rate is computed only over clusters the fallback answered; it counts valid `new` answers before the pending cap and is `unmeasured` when no cluster was answered (for example, on hosted runners without `codex`). Clusters `unresolved` by a Jev failure, a cap or the per-merge cluster cap enter neither rate.

Make the minimum a policy key: `mapping.association.guard.{window_merges, min_clusters, fallback_rate_max, pending_new_rate_max}`, with `min_clusters` 20 **[ASSUMPTION]**. List all four keys in A14.

### R1-11 (low): two figures and one key list misstate the evidence

1. `:114` "[ASSUMPTION: about 1.5 × the clusters of an average fastapi push; not measured]". fastapi averaged about 2.2 clusters per merged PR (92/41 in rounds 8–10). Replace with "[ASSUMPTION; not measured: fastapi averaged about 2.2 clusters per merged PR and about 27 of 100 escalate; 40 is a cost ceiling for large pushes]".
2. `:114` policy list `mapping.association.{enabled, t_choice, covered_score, fallback.{model, effort, timeout, max_per_run}}` omits the pending cap that the proposal names. Add `max_pending_per_run` (default 10 **[ASSUMPTION]**).
3. `proposal.md` "Problem": B2 rounds 1–7 "0.452 / 0.420" sits under "unresolved counted as a miss", but 0.420 is resolved recall (71/169, `README.md:114`). Replace with "0.452 / 0.380 (71/187; 0.420 on resolved clusters)".

### R1-12 (medium): the R-059 acceptance is presented as covering a result the user had not seen

`:19` and Appendix 23(c) state that the user accepts R-057 and R-058 (fastapi 0.922, axum 0.622) as the design rule 3 exception. R-059's own limits column says "Taken before the axum replication result (R-058) was reported". This does not reopen the decision, but the record must be accurate.

**Fix.** Append to `:19` and to Appendix 23(c): "(R-059 was taken before the axum result, R-058, was reported; the user's confirmation that the exception stands after seeing axum P_assign 0.622 is recorded in the review round that signs off)". Add the confirmation to the proposal's "Remaining questions" (open question U1).

### R1-13 (medium): X11 cannot run as written, and its failure branch keeps a lane that fails the same way

The X11 row (`:444`) has three problems:
- **Failure branch.** It says "else the lane stays opt-in or is removed (Jev-only with unresolved left unowned)". Jev-only has the same failure: axum P_assign 0.610, with 22 of its 30 errors on gold-`none` (`axum/README.md`).
- **Repository selection.** "one with a high new-feature share (gold-`none` ≥ 40 %)" can only be known after labelling.
- **Coverage.** It omits the closed-loop product state (R1-09), the guard's blind spot and remedy (R1-10), and the possibility that a labeller shares Terra's model: labeller B is `codex exec` with no `-m`, `axum/pilot/label_x2d.py:257`. It also has no hypothesis-ledger entry (ledger rule 2).

**Fix.**
- **Failure branch.** Replace "else the lane stays opt-in or is removed (Jev-only with unresolved left unowned)" with "else the lane stays opt-in or is removed (no merge-time association: residual files keep a directory scope, as in revision 2.7); a Jev-only variant stays only if it meets the same bar".
- **Selection.** Replace "including one with a high new-feature share (gold-`none` ≥ 40 %, as axum)" with "including one chosen by a model-free proxy fixed before labelling (e.g. the share of residual clusters in directories no area owns) expected to give a high gold-`none` share; the gold-`none` share, split into `gap` and `new`, is reported afterwards".
- **Method.** Append: "Replay merge sequences closed-loop, as the product runs (own prior assignments, pending areas kept out of the catalogue, a preregistered maintainer-action policy, no gold membership updates); report duplicate pending areas per real new concern; report the share of wrong assignments on clusters Jev assigned without escalation; test whether P_assign recovers after the recommended `scout map --refresh`; score clusters against catalogues mapped at two or more commits; record each labelling session's reported model and use labellers other than `gpt-6.1-terra`, or report agreement by labeller model; count clusters beyond the per-merge cap as `unmeasured`."
- **Ledger.** Record HD8 (proposed) in `docs/hypothesis-ledger.md`: "Jev + `gpt-6.1-terra` merge-time association reaches P_assign and recall ≥ 0.80 in every repository tested, including a high gold-`none` one; test X11". Cite HD8 in the X11 row.

### R1-14 (low): the relation to X3 is claimed in the proposal but not in the design

The proposal says "This proposal supersedes X3's file-level shape by a cluster-level one; … X3 is not run". The design leaves X3 (:438) and the §6 row (:163, "unowned-file → area suggestion (X3) … on from launch") unchanged. This leaves two Jev paths that could act on the same files.

**Fix.** Append to the X3 row: "(DRAFT rev 2.8: files added or moved by a default-branch merge are decided by the §4 merge-time association lane, validated by X11; X3 covers the other unowned files and its rule is unchanged)".

## Considered and not raised

- **Jev compared with labels outside a signed-off preregistration (R-056 to R-058).** The user directed this (R-055), and §4 discloses "Development runs, no preregistration, no council review". The design itself runs no such comparison. No design change is needed.
- **Trust boundary.** The lane runs only on trusted default-branch pushes. PR and fork runs stay read-only with no secrets (:184, :200). The Jev key is stripped from the `codex` environment (`internal/modelcli/modelcli.go:65-84`). The prompt treats PR and file text as untrusted data. Voiding command-running sessions (R1-01) closes the remaining gap. No further issue.
- **Appendix 4.** Labels never change issue state or the gate. This holds.

## Where the peers changed the lead's view

- **Challenger.** Proposed two staged lanes instead of one with a double `requires` (R1-01). Flagged Appendix 19 / B-3 (R1-06) and the guard's numerator definitions (R1-10). Its option of separate consent naming Codex agreed with §6 :184, which settled R1-07.
- **Experiment peer.** Found the gap/new split and the round-1 data against staleness (R1-08), the guard firing on fastapi and its blind spot (R1-10), the R-059 timing (R1-12), the X11 failure branch, selection and labeller-model problems (R1-13), the Choice limit, the second permutation and command-session voiding (R1-04, R1-01), and the `max_per_run` arithmetic (R1-11).
- **Lead.** Found the persistence and schema problem (R1-03), the catalogue digest in the version key (R1-05), the undisclosed teacher forcing (R1-09, also raised by both peers) and the X3 overlap (R1-14). Proposed deciding pending-area visibility from the measured setting rather than leaving it open (R1-02).

## Open questions that need the user

- **U1 (R1-12).** R-059 was decided before the axum result was reported. Does the design rule 3 exception for the provisional opt-in lane still stand now that axum gave Jev+Terra P_assign 0.622 [0.514, 0.719], did not replicate fastapi, and its interval overlaps B2's? This confirms the decision; it does not reopen it.

The proposal's Q2 (Appendix 4/19), Q4 (consent) and the pending-area visibility question are settled by existing design text (R1-06, R1-07) and by the measured setting (R1-02). They need no user decision unless the user wants something stricter.

## Verdict

**14 ISSUES** (R1-01 to R1-14: high R1-01, R1-02, R1-03, R1-04, R1-08, R1-10; medium R1-05, R1-06, R1-07, R1-09, R1-12, R1-13; low R1-11, R1-14). No sign-off in this round. A later sign-off must name the SHA-256 digests of the revised `docs/scout-design.md` and `proposal.md`, and holds only while those digests are unchanged. Round 2 reviews the revised draft and the record of U1.
