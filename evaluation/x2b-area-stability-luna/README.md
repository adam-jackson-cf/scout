# X2b — area stability on Scout's production mapper (codex `gpt-6-luna`, effort `xhigh`)

Question: is there a Jaccard match threshold at which `internal/areas.Reconcile` keeps ≥ 80 % of prior areas with ≤ 5 % spurious merges when the agent map is made as Scout now makes it (`--provider codex --model gpt-6-luna --reasoning-effort xhigh`; design revision 2.5)? A codex-only repeat of X2 (`../x2-area-stability/`); everything is in `PREREGISTRATION.md` (revision 4; sign-off state lives in `reviews/`, the latest round is the record). Nothing here has been run on real data; two pre-run measurement checks (one probe, two single smoke maps) are disclosed in `PREREGISTRATION.md` §2.

## Layout
| File | Role |
|---|---|
| `PREREGISTRATION.md` | the design: what is reused from X2, the changes, the decision combination with X2 (§4a), the HD5b state mapping (§4b) |
| `x2b.py` | the pinned configuration, the SHA-256 pins of every reused frozen X2 file, the drift guard against `internal/clawpatch/client.go` / `provider.go`, the probe arguments |
| `collect_x2b.py` | the collector: a **copy** of X2's `collect.py` changed only as `validation/collect-x2b.diff` shows (Scout's two flags, the probe with effort, one arm, pin and drift refusals) |
| `analysis_x2b.py` | X2's `analysis.py` **by import** + the served/flags check, the X2 cross-check and the decision combination |
| `run.sh` | `validate`, `replay`, `pairs`, `label`, `analysis`; replay, pairs and labelling are X2's frozen code on X2b's `data/` |
| `validation/validate.py`, `stub_clawpatch_x2b.js`, `collect-x2b.diff` | P/X/S/C/E checks (below) |
| `DRAFT-NOTES.md` | historical digests, check counts, judgement calls and the ledger rows as applied (not a signed document) |
| `reviews/` | council rounds |
| `data/` | empty until sign-off |

Reused from `../x2-area-stability/` unchanged and verified against pinned digests before every stage: `common.py`, `analysis.py`, `build_pairs.py`, `label.py`, `GUIDE.md`, `replay/main.go`, `PREREGISTRATION.md`, `validation/validate.py`, `validation/stub_clawpatch.js`, `validation/credcheck/main.go`.

## Reproduce
Toolchain (`AGENTS.md`): `git submodule update --init && scripts/bootstrap-tools.sh && source .tools/activate` (the review-safety Clawpatch build; the collector refuses any other).

Validation (no model calls, no data; about 3 min; exits non-zero on any failed check; it also re-runs X2's own validation unchanged as check `X`):

    source .tools/activate && evaluation/x2b-area-stability-luna/run.sh validate

Real run (**only after council sign-off**; code under test `2a2a346`, design rev 2.5 adopted): `git worktree add ../scout-x2b <run commit> && cd ../scout-x2b`, where the run commit is a descendant of `2a2a346` containing the signed-off study files (`git diff --quiet 2a2a346 <run commit> -- internal/ scripts/ third_party/clawpatch` must report no difference); bootstrap the toolchain there as above, and `git status --porcelain internal/ scripts/ evaluation/x2b-area-stability-luna/` must be empty; no session edits that tree during collection (the collector re-reads `client.go`/`provider.go` before every map; a concurrent edit that breaks a guarded string aborts the collection, which this append-only study cannot repair). `data/` is produced in that worktree and committed from it unchanged. `SOURCE_*` = any local clone or URL containing the pinned revisions. Needs a logged-in `codex` (the model named `gpt-6-luna` must be available) and no `CLAWPATCH_MODEL`/`CLAWPATCH_REASONING_EFFORT`/`CLAWPATCH_PROVIDER`/`CLAWPATCH_CONFIG`/`CLAWPATCH_CODEX_SANDBOX` in the environment (the collector refuses them). `rm -rf /tmp/scout-x2b` first (disposable clones). Order: codex/fastapi, then codex/axum, sequential, under `nohup` with its output redirected (an ignored SIGHUP is kept) or `tmux`:

    PY=.tools/venv/bin/python; C=evaluation/x2b-area-stability-luna/collect_x2b.py
    nohup sh -c "$PY $C --arm codex --repo fastapi --source $SOURCE_FASTAPI &&
                 $PY $C --arm codex --repo axum    --source $SOURCE_AXUM" > x2b-collect.log 2>&1 &
    # once x2b-collect.log shows both collections finished:
    evaluation/x2b-area-stability-luna/run.sh replay
    evaluation/x2b-area-stability-luna/run.sh pairs
    evaluation/x2b-area-stability-luna/run.sh label       # appends data/pairs/{calls,labels}.jsonl
    evaluation/x2b-area-stability-luna/run.sh analysis    # data/results.json, data/analysis_output.txt (verdict, served/flags check, combination PROPOSAL)

The collector refuses to start (before any clone or model call, except the before-probe where stated) when: a map directory, attempt file, versions file or stop record of the repository exists; its workspace `/tmp/scout-x2b/codex/<repo>` exists; any of the five Clawpatch env names is set (even empty); `codex` is not on PATH; Clawpatch `--version` is not 0.7.3; the Clawpatch build is not the review-safety build; any reused X2 file differs from its pin; Scout's Go source no longer states `gpt-6-luna`/`xhigh` via the codex-only map resolver; or the before-probe (which runs `exec --cd <tmp> --sandbox read-only --skip-git-repo-check --model gpt-6-luna -c model_reasoning_effort="xhigh" -`) fails or names no model or no reasoning effort. `--arm claude` does not exist.

## Frozen files (digests in the final review round)
`PREREGISTRATION.md`, `README.md`, `x2b.py`, `collect_x2b.py`, `analysis_x2b.py`, `run.sh`, `validation/validate.py`, `validation/stub_clawpatch_x2b.js`, `validation/collect-x2b.diff`; by pin the X2 files above; code under test (commit `2a2a346`) `internal/clawpatch/client.go`, `internal/clawpatch/provider.go`, `internal/areas/areas.go`, `internal/clawpatch/records.go`, `scripts/clawpatch-safe.patch`, `scripts/build-clawpatch.sh`. The final round names the design round, `2a2a346`'s blob digests and the run-commit rule (the run commit itself is recorded in `RESULTS.md` when it is created after sign-off); RESULTS shows the recorded Go digests (`client_go_sha256`, `provider_go_sha256`, `areas_go_sha256`, `records_go_sha256`) equal to the `2a2a346` blobs, `collect_x2b_sha256` equal to its sign-off digest, `scout_head` equal to the run commit and the empty `git diff` (PREREGISTRATION §2.6).
