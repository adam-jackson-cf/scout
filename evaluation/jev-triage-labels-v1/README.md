# jev-triage-labels-v1: agent first-pass labels

Protocol: `docs/STATUS.md` §6 item 8; `jev-actions` skill (`labeling-cases.md`, `building-eval-data.md`).
Question: E2 `v1_choice` (`evaluation/jev-triage-v1/questions.py:3-8`). 52 cases (22 `cases.jsonl` + 30 `unlabelled.jsonl`).

- Guide: `GUIDE.md`, version `triage-v1_choice/guide-1`, sha256 `376969542940b0a49e110ab4a27d39e121eb19679c0fc74213bf6e8ef0bf391f` (frozen before labelling).
- Blinding: `blinded.jsonl` = {case_id, state} only (label/split/deleak dropped), seeded shuffle, seed `20261022`; `sources.jsonl` maps case_id to source file.
- Labeller A: `claude -p`, resolved model `claude-opus-5-5` (CLI default), `--tools ""`, `--no-session-persistence`, `--setting-sources ""`, empty temp cwd.
- Labeller B: `codex exec`, resolved model `gpt-6.1-sol` (default, reasoning effort none), read-only sandbox, `--ephemeral --ignore-user-config --ignore-rules`, empty temp cwd.
- Adjudicator C: `claude -p --model sonnet`, resolved `claude-sonnet-5-5`; sees guide, state, A and B labels+reasons only; called where A != B or either unparsable.
- Strict JSON parse, one retry, then `unparsable` (none occurred). Worker pool of 4. Per-call model/timestamp in `calls.jsonl`.
- Proposed label = agreed label, else C's.

## Results
- Total calls: 108 (52 A + 52 B + 4 C), no retries needed.
- A/B agreement 48/52 = 92.3%, Cohen's kappa 0.856 (3 labels, n=52). 4 adjudicated by C, 0 unparsable.
- Proposed: real_defect 28, false_positive 23, cannot_tell 1. (A: 28/23/1; B: 26/23/3.) Detail: `ab_agreement.json`.
- Note: all four adjudications went to A's label, so proposed labels equal A's for every case.

## Files
`labels.jsonl` has 212 rows: the A, B and C first-pass rows (`agent`), one proposed row per case (`agent_reviewed`), and one user decision per case (`human`, `by: user`, 2026-10-23; all 52 proposals confirmed). The other files are `REVIEW.md` (the review sheet) and `E2-COMPARISON.md`.

## Reproduce (from this folder)
```
python3 blind.py        # blinded.jsonl, sources.jsonl
python3 label.py        # appends calls.jsonl, labels.jsonl (needs claude and codex CLIs); refuses if rows for this guide_version exist
python3 label.py --new-round <name>   # later rounds: appends rows tagged with "round"
python3 report.py       # REVIEW.md, ab_agreement.json
python3 compare_e2.py   # E2-COMPARISON.md (only after proposals are final)
```
`label.py` is append-only and never rewrites existing rows. `report.py`/`compare_e2.py` read the original (untagged) round; they have not been adapted to new rounds.

## NOT done
- No Jev-vs-label comparison, accuracy, AUROC or threshold work (needs a council-signed preregistration).
- User review done 2026-10-23: all 52 proposed labels confirmed and appended as provenance `human`. The latest human label overrides agent labels.
- Agent labels are not ground truth; the E2 comparison was computed only after proposals were final.
