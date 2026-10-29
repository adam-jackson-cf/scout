# X2c — mapper comparison (luna xhigh, terra none/low, haiku) × area-matching method (Jaccard, Jev)

Question: which agent mapper (codex `gpt-6-luna` `xhigh` as reference, codex `gpt-6.1-terra` at `none` and `low`, Claude Code `claude-haiku-5-5`) and which area-matching method (X2's Jaccard rule, or a Jev score s = P(same_concern) + P(overlapping_concern), mean of both presentation orders, on every file-sharing candidate pair) keep Scout's areas stable (X2's C1–C3), at what refresh time and Jev cost? Everything is in `PREREGISTRATION.md` (revision 3, submitted for council round 4; the latest round in `reviews/` is the sign-off record; `reviews/round-1.md` is the council's round 1 and `reviews/round-1-response.md` records the fixes). Nothing here has been run on real data; the pre-run measurement checks (arm probes, `smoke/SMOKE.md`, candidate counts, a 30-input reading) are disclosed in `PREREGISTRATION.md`.

## Layout
| File | Role |
|---|---|
| `PREREGISTRATION.md` | the design: arms, candidate set, the Jev method and its question design choices, the cell/selection decision rule, validation, threats |
| `x2c.py` | the four arms, the SHA-256 pins of every reused frozen X2 file, the Jev pin and question-version constants, the decision constants, probe arguments |
| `collect_x2c.py` | the collector: a **copy** of X2's `collect.py` changed only as `validation/collect-x2c.diff` shows (arm table → Clawpatch `--provider/--model/--reasoning-effort`, probes with model and effort, pin check) |
| `build_candidates.py` | the candidate pairs (every prior-area × next-feature pair sharing ≥ 1 owned file) + X2's controls, in X2's file formats, + `jev-input.jsonl` |
| `jev/main.go` | the Jev scorer: a thin driver over Scout's `internal/decisions` client (one Choice question, two orders, append-only resumable output, cumulative billed-cost stop, a `schema` failure for an incomplete or unnormalised Choice reply) |
| `jev_match.py` | Jev scores (s = P(same_concern) + P(overlapping_concern), mean of both orders) → replay-schema rows (greedy best-first one-to-one, per grid threshold) |
| `analysis_x2c.py` | X2's `analysis.py` **by import** on all cells + served check, Jev check, confirmation, coverage guard, refresh time, selection |
| `run.sh` | `validate`, `collect`, `replay`, `pairs`, `label`, `jev`, `analysis` |
| `validation/` | `validate.py`, `stub_clawpatch_x2c.js`, `stub_jev.py`, `collect-x2c.diff` |
| `smoke/` | the pre-run measurement check (disclosed, not a signed-off script; do not reuse unreviewed) |
| `DRAFT-NOTES.md` | digests, judgement calls, ledger drafts (not applied), open questions |
| `reviews/` | council rounds (`round-<N>.md`) and the drafter's responses (`round-<N>-response.md`) |
| `data/` | empty until sign-off |

Reused from `../x2-area-stability/` unchanged and verified against pinned digests before every stage: `common.py`, `analysis.py`, `build_pairs.py`, `label.py`, `GUIDE.md`, `replay/main.go`, `PREREGISTRATION.md`, `validation/validate.py`, `validation/stub_clawpatch.js`, `validation/credcheck/main.go`.

## Reproduce
Toolchain (`CLAUDE.md`): `git submodule update --init && scripts/bootstrap-tools.sh && source .tools/activate` (the review-safety Clawpatch build; the collector refuses any other).

Validation (no model calls, no data; X2c's checks plus X2's own re-run inside it; exits non-zero on any failed check):

    source .tools/activate && evaluation/x2c-mapper-comparison/run.sh validate

Real run (**only after council sign-off**; in a dedicated worktree at the run commit, a descendant of the sign-off commit that contains the signed-off files; `git status --porcelain` empty for `internal/ scripts/ evaluation/x2c-mapper-comparison/`):

    SRC=/tmp/scout-x2c-smoke/_src      # any clone containing the pinned revisions; the collector verifies the SHA
    nohup evaluation/x2c-mapper-comparison/run.sh collect $SRC/fastapi $SRC/axum > x2c-collect.log 2>&1 &
    # once x2c-collect.log shows all 8 collections finished (map failures and stop records: data/stopped-*.json):
    evaluation/x2c-mapper-comparison/run.sh replay
    evaluation/x2c-mapper-comparison/run.sh pairs
    evaluation/x2c-mapper-comparison/run.sh label --workers 8        # appends data/pairs/{calls,labels}.jsonl
    infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- evaluation/x2c-mapper-comparison/run.sh jev
    evaluation/x2c-mapper-comparison/run.sh analysis                 # data/results.json, data/analysis_output.txt

The Jev key is only ever injected by Infisical into the `jev` command; check availability by presence only. `run.sh jev` can be re-run: it resumes (answered units are skipped, failures retried) and refuses a changed input, question version or model, or a budget already exhausted (the 5 USD stop is cumulative over `data/jev/scores.jsonl`). Before each driver run it appends the SHA-256 of the Scout code compiled into the driver to `data/jev/code-digests.jsonl`. `X2C_DATA` overrides the data directory for every stage, `collect` included.
The collector refuses to start (before any clone or model call, except the before-probe) when: a map directory, attempt file, versions file or stop record of the (arm, repo) exists; its workspace `/tmp/scout-x2c/<arm>/<repo>` exists; any of the five Clawpatch env names is set (even empty); the arm's CLI is not on PATH or Claude Code is older than 2.1.169 (claude arms); Clawpatch `--version` is not 0.7.3; the build is not the review-safety build; any reused X2 file differs from its pin; or the before-probe fails or names no model or no reasoning effort.

## Frozen files (digests in the final review round)
`PREREGISTRATION.md`, `README.md`, `x2c.py`, `collect_x2c.py`, `build_candidates.py`, `jev_match.py`, `analysis_x2c.py`, `jev/main.go`, `run.sh`, `validation/validate.py`, `validation/stub_clawpatch_x2c.js`, `validation/stub_jev.py`, `validation/collect-x2c.diff`; by pin the X2 files above; code under test every non-test file of `internal/decisions/`, `internal/policy/consent.go`, `internal/areas/areas.go`, `internal/clawpatch/records.go`, `go.mod`, `go.sum`, `scripts/clawpatch-safe.patch`, `scripts/build-clawpatch.sh`.
