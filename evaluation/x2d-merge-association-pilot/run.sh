#!/usr/bin/env bash
# X2d stages (README.md). Run from anywhere; needs the pinned toolchain (`source .tools/activate`). Only the stages marked [model] call a model.
#   run.sh validate                    model-free implementation validation (validation/validate.py); no data, no model call
#   run.sh onboard-collect [SRC]       [model] 15 scoped Clawpatch maps + 1 contrast draw at C0 (sequential)
#   run.sh onboard                     [model: Jev, infisical-wrapped] code-first consensus, Jev for contested files, freezes data/catalogue/ and its SHA-256
#   run.sh quality                     [model: labellers] section 4.3 report incl. the intruder audit
#   run.sh gates                       [model: Jev + labellers] round-0 gates K2-K6, P1, N1, planted items on reserve sets D/P
#   run.sh regate-build|regate-label|regate-jev|regate-analyse   [model: Jev + labellers] the ONE permitted re-gate of P1 and P2 on fresh items (PREREGISTRATION 5, rev 13; R-053), one stage at a time, the build alone first, all inside
#                                      D/rounds/0/regate (a self-contained data root; the round-0 files are never written); one shot; regate-jev checks its label-only denominators before any Jev request; round 1 needs its PASS
#   run.sh power                       [model-free] P3 / N3 / N3-PR / N4a / N4b / N4b-cap / N4b-rho0 on the live frozen-round structure (rounds/0/k23.json) -> data/power.json (after gates; read by the round-0 critique); exit 3 = a null cell above its bar (stop, new council round)
#   run.sh round R                     [model] R in 1..10: units -> label -> labels-check -> gold -> baselines -> jev -> tune -> analysis (each stage also callable alone, see below)
#   run.sh units|label|labels-check|gold|baselines|jev|tune|analysis R
#   run.sh clarify-precheck R          rounds 1-7, BEFORE editing the guide: the clarification's session-cap pre-check alone (`label_x2d.py --precheck-clarify`; a refusal writes rounds/R/clarify-refused.json and exits non-zero: do not edit the guide)
#   run.sh clarify R "reason"          [model] rounds 1-7 only, before Jev scored round R: record the edited rounds/R/inputs/GUIDE.md, relabel round R's wide cases under the new guide, re-run labels-check
#   run.sh rescore-v7                  [model: Jev, infisical-wrapped] after the round-7 critique: re-score rounds 0-7 under v_7 (rounds/<rr>/jev-v7/scores.jsonl, sem + fan) for the round-8 tuning
#   run.sh verdict                     GO / INCONCLUSIVE / NO-GO once, on the pooled frozen rounds 8-10 (needs the round-10 critique, $X2D_REVIEWS_DIR/pilot-round-10.md)
# Environment: X2D_DATA (data dir, every stage), X2D_SRC (a clone containing the pinned revisions; default /tmp/scout-x2c-smoke/_src/fastapi), X2D_JEV_ENDPOINT (validation
# points it at a stub; the real run leaves it unset), X2D_SKIP_SIGNOFF=1 and X2D_REVIEWS_DIR (validation only: honoured only with a loopback X2D_JEV_ENDPOINT and an X2D_DATA outside
# S/data, else refused). The Jev key is never read here: run the
# stages that need it as `infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- run.sh <stage>`.
# A stage refuses when its outputs exist (append-only), when labels already exist for a changed unit set, when a digest of a frozen file differs from the sign-off
# record, or (jev) when the cumulative billed cost exceeds the cap.
set -euo pipefail
cd "$(dirname "$0")/../.."
S=evaluation/x2d-merge-association-pilot
PY=.tools/venv/bin/python
D=${X2D_DATA:-$S/data}
case "$D" in /*) ;; *) D="$PWD/$D" ;; esac
export X2D_DATA="$D"
SRC=${X2D_SRC:-/tmp/scout-x2c-smoke/_src/fastapi}
C0=1d50bad4555b951d5fa48eb8787fd5c39ed7effb

py() { "$PY" "$@"; }
pyx() { "$PY" -c "import sys; sys.path.insert(0, '$S'); import x2d; $1" "${@:2}"; }

verify() {
  pyx "bad = x2d.verify_pins(require_signoff=x2d.require_signoff()); sys.exit('pin check failed: ' + '; '.join(bad)) if bad else 0"
}

is_contrast_round() { [ "$1" = 1 ] || [ "$1" = 10 ]; }

need_round() { [ -n "${1:-}" ] && [ "$1" -ge 1 ] 2>/dev/null && [ "$1" -le 10 ] || { echo "usage: run.sh $STAGE R   (R = 1..10)" >&2; exit 2; }; }

# Jev driver run: jev_run ROUND_KEY(or -) UNITS CATALOGUE OUT ARMS [extra driver flags]. Writes the code digests, then runs the driver with the remaining budget. The round's questions are
# its frozen inputs (rounds/<r>/inputs/questions, R2-01) and its frozen settings.json; rescore-v7 sets JEV_SETTINGS / JEV_QUESTIONS (shell variables, never read from the environment: a live
# environment cannot bypass the freeze) to round 7's, while the spend is charged to the round of OUT.
JEV_SETTINGS=""
JEV_QUESTIONS=""
jev_run() {
  local rk=$1 units=$2 cat=$3 out=$4 arms=$5; shift 5
  mkdir -p "$(dirname "$out")"
  pyx "import json, time; open('$D/code-digests.jsonl', 'a').write(json.dumps({'utc': time.strftime('%Y-%m-%dT%H:%M:%SZ', time.gmtime()), 'digests': x2d.go_code_digests()}) + '\n')"
  local budget settings="" qsrc=""
  if [ "$rk" = "-" ]; then budget=$(pyx "print(x2d.jev_budget(None, '$out'))"); else budget=$(pyx "print(x2d.jev_budget($rk, '$out'))"); fi
  if [ "$rk" != "-" ]; then
    settings=${JEV_SETTINGS:-$D/rounds/${rk}/settings.json}
    [ -f "$settings" ] || settings=""
    qsrc=${JEV_QUESTIONS:-$D/rounds/${rk}/inputs/questions}
    [ -d "$qsrc" ] || { echo "refusing: $qsrc is missing: versioning.py --round $rk creates the round's inputs" >&2; exit 1; }
  fi
  go run "./$S/jev" -in "$units" -catalogue "$cat" -out "$out" -arms "$arms" -questions-dir "$(dirname "$out")/questions" -workers 8 -max-cost-usd "$budget" \
    ${qsrc:+-questions-src "$qsrc"} ${settings:+-settings "$settings"} ${X2D_JEV_ENDPOINT:+-endpoint "$X2D_JEV_ENDPOINT"} "$@"
}

# The round-0 gates must have been evaluated and passed before any round (PREREGISTRATION section 5). Rev 13 (R-053): a round-0 FAIL of P1 and/or P2 alone is lifted only by the one re-gate on fresh items
# (D/rounds/0/regate/gates.json, gate set x2d-regate-v1) showing PASS for BOTH; a FAIL of any other gate, or a re-gate that is not PASS, stops the pilot (x2d.gate_problems).
require_gates() {
  pyx "bad = x2d.gate_problems(); sys.exit('refusing: ' + '; '.join(bad)) if bad else 0"
}

# Council critique between rounds (PREREGISTRATION section 10; R1-09, R2-16): round R in 1..10 needs $X2D_REVIEWS_DIR/pilot-round-<R-1>.md (default $S/reviews); x2d.critique_problems lists
# what is wrong with it: exactly one line starting `Verdict:` equal to `Verdict: ZERO ISSUES` (R-1 <= 7, the refinement rounds; the critique of round 7 freezes v_7) or
# `Verdict: DESCRIPTIVE` (R-1 >= 8), exactly one `version_sha256: <SHA-256 of rounds/<R-1>/version.json>`, for R-1 <= 7 exactly one `next_inputs_sha256: <inputs digest>` (the inputs the
# next round uses; R-1 = 7: round 7's own), for R-1 = 0 also `real_none_proxy: <int>` (INFO), `p3_real_structure: <float>`, `p3_abstention: 0.05=<f> 0.10=<f>`, `n4_max: <float>` (equal to gates.json / power.json,
# which must exist), `frozen_go_population: 8=<n>/<prs> 9=.. 10=..` (equal to rounds/0/k23.json), `user_decision: proceed` when a power.json check is false or a frozen round has < 20 clusters or < 6 PRs, and
# `p2_unmeasured: accepted` when the P2 gate is UNMEASURED (R3-16, R3-04, R4-13 (7)). `verdict` needs the critique of round 10.
critique_ok() {
  pyx "bad = x2d.critique_problems(int(sys.argv[1])); sys.exit('refusing: ' + '; '.join(bad)) if bad else 0" "$1"
}
require_critique() {
  local r=$1
  if [ "$r" -ge 1 ] && [ "$r" -le 10 ]; then critique_ok "$((r - 1))"; fi
}

# The round's frozen version (inputs, settings, thresholds, state builder, catalogue, code) must be unchanged since the freeze (R2-03).
require_version() { py "$S/versioning.py" --round "$1" --drift; }

# labels-check.json of the round must exist, be for the round's CURRENT guide version, and be bound to the CURRENT label rows (labels_sha256 = digest of the current-guide wide rows of the checked
# units, R3-07: after `label --new-round NAME` run `labels-check R` again). A failed check (kappa_answer < 0.6 on the GO population; the rule concentration is report-only, R3-02) refuses only in rounds 1-7
# while no guide revision is recorded (the one clarification, `clarify`) and no clarify-refused.json exists (R4-02: a clarification whose relabel cannot fit the session cap); after either, in rounds 8-10 and in
# round 0 (K4 on D is its kappa gate) the round proceeds scored and non-informative.
require_labels_check() {
  pyx "
import json, os
import analysis_x2d as A
r = int(sys.argv[1])
p = os.path.join(x2d.round_dir(r), 'labels-check.json')
if not os.path.isfile(p):
    sys.exit('refusing: %s is missing: run labels-check %d first' % (p, r))
j = json.load(open(p))
if j.get('guide_version') != x2d.guide_version(r):
    sys.exit('refusing: the labels-check of round %d is for guide %s, the current guide is %s' % (r, j.get('guide_version'), x2d.guide_version(r)))
if j.get('labels_sha256') != A.labels_digest(r):
    sys.exit('refusing: the labels of round %d changed after its labels-check (a label fill?): re-run labels-check %d' % (r, r))
revised = bool(x2d.read_jsonl(os.path.join(x2d.round_dir(r), 'guide-revisions.jsonl'))) or os.path.isfile(os.path.join(x2d.round_dir(r), 'clarify-refused.json'))  # R4-02: a refused pre-check proceeds like a recorded revision
if not j.get('ok') and r in x2d.REFINEMENT_ROUNDS and not revised:
    sys.exit('refusing: the labels-check of round %d failed: %s (clarify the guide once: run.sh clarify %d \"reason\")' % (r, '; '.join(j.get('failed', [])), r))
" "$1"
}

# every per-round stage after the freeze
prereq() { local r=$1; need_round "$r"; verify; require_gates; require_critique "$r"; require_version "$r"; }

# Futility stop (PREREGISTRATION section 9): exit 3 of analysis_x2d.py --futility means three consecutive informative futile rounds; the frozen rounds are then not run.
check_futility() {
  local rc=0
  py "$S/analysis_x2d.py" --futility || rc=$?
  if [ "$rc" -eq 3 ]; then echo "refusing: futility rule fired (three consecutive informative rounds): NO-GO (futility); the remaining rounds are not run" >&2; exit 3; fi
  [ "$rc" -eq 0 ] || exit "$rc"
}

# Round 8 tunes once on the v_7 re-scores of rounds 0-7 (rescore-v7, R2-02): all eight files must exist, with the headers of round 7's Jev run, before round 8 starts.
require_v7_rescore() {
  pyx "import tune; rs = tune.dev_rounds(8, x2d.DATA, x2d.V7_RESCORE_SUB); sys.exit('refusing: the v_7 re-score covers rounds %s, not 0-7 (run.sh rescore-v7)' % rs) if rs != list(range(8)) else 0"
}

stage_units() {
  local r=$1
  need_round "$r"; verify; require_gates; require_critique "$r"
  if [ "$r" -ge 4 ] && [ "$r" -le 8 ]; then check_futility; fi
  if [ "$r" -eq 8 ]; then require_v7_rescore; fi
  py "$S/versioning.py" --round "$r" --check   # every r >= 1 (round 1 against round 0, R3-12), BEFORE anything is frozen or written: the live catalogue / state builders / thresholds / round table / code must be within the envelope
  py "$S/versioning.py" --round "$r"
  py "$S/build_units.py" --repo "$SRC" --round "$r"
  py "$S/build_units.py" --repo "$SRC" --round "$r" --tag contrast   # every round (R1-13): the window units restratified against the contrast catalogue (+ the D units in rounds 1 and 10)
}

# Baselines run AFTER gold exists (R1-08): against the membership at each unit's parent commit (round start + earlier same-round gold + inherit ops). Contrast-B2 every round.
stage_baselines() {
  local r=$1; prereq "$r"
  [ -s "$D/rounds/$r/gold.jsonl" ] || { echo "refusing: $D/rounds/$r/gold.jsonl missing: baselines run after gold" >&2; exit 1; }
  py "$S/baselines.py" --repo "$SRC" --round "$r"
  py "$S/baselines.py" --repo "$SRC" --round "$r" --tag contrast
}

# The exact Jev states of the round's units (R1-16): `jev -dump-state` (no key, no request), from the round's inputs (R2-01) and settings. The perception labelling and the reference arm read
# them. Written once.
ensure_states() {
  local r=$1 R="$D/rounds/$1" settings=""
  [ -s "$R/states.jsonl" ] && return 0
  [ -f "$R/settings.json" ] && settings="$R/settings.json" || true
  go run "./$S/jev" -dump-state -in "$R/units.jsonl" -catalogue "$D/catalogue/catalogue.json" -arms sem -out "$R/states.jsonl" -questions-src "$R/inputs/questions" ${settings:+-settings "$settings"}
}

# stage_label forwards only --new-round NAME, --workers N and --cases IDS to label_x2d (R2-16: never --catalogue, --states, ...)
check_label_args() {
  while [ $# -gt 0 ]; do
    case "$1" in
      --new-round|--workers|--cases) [ $# -ge 2 ] || { echo "refusing: $1 needs a value" >&2; exit 2; }; shift 2 ;;
      *) echo "refusing: stage label forwards only --new-round, --workers and --cases (got $1)" >&2; exit 2 ;;
    esac
  done
}
stage_label() {
  local r=$1; prereq "$r"; check_label_args "${@:2}"
  py "$S/label_x2d.py" --round "$r" "${@:2}"
  ensure_states "$r"
  py "$S/label_x2d.py" --round "$r" --kind jev-state "${@:2}"
  if is_contrast_round "$r"; then   # rounds 1 and 10: <= 30 hash-sampled window units against the contrast catalogue (R1-13)
    py "$S/label_x2d.py" --round "$r" --tag contrast "${@:2}"
  fi
}

# Pre-score label gate (R1-09, R2-03, R3-02): writes rounds/R/labels-check.json (kappa of the joint answer on the GO population; the rule concentration is reported, never a trigger) and exits non-zero
# only when kappa_answer < 0.6 in rounds 1-7 before the one guide clarification (`clarify`, runs before Jev scores) and no `clarify-refused.json` is recorded (R4-02). jev, gold, tune and analysis refuse without a
# labels-check of the round's current guide version over the current label rows (re-run it after `label --new-round`).
stage_labels_check() { local r=$1; prereq "$r"; py "$S/analysis_x2d.py" --labels-check "$r"; }

# Gold and the next round's membership (R2-04: nothing is written when a window unit lacks usable gold; fill it with `label --new-round NAME`, then run again).
stage_gold() {
  local r=$1; prereq "$r"; require_labels_check "$r"
  # R4-01: each gold file only when absent/empty (append-only; gold.py itself refuses an existing file): the contrast gold (the last labelling step, the first to be capped) can complete after a fill
  # while the main gold and the next round's membership already stand
  local g="$D/rounds/$r/gold.jsonl" gc="$D/rounds/$r/gold.contrast.jsonl"
  if [ -s "$g" ]; then echo "$g exists (append-only): skipped"; else py "$S/gold.py" --round "$r"; fi
  if is_contrast_round "$r"; then
    if [ -s "$gc" ]; then echo "$gc exists (append-only): skipped"; else py "$S/gold.py" --round "$r" --tag contrast; fi
  fi
}

# Guide clarification (PREREGISTRATION 10 (v); R2-01 (4)): rounds 1-7, before Jev scored the round. Run `clarify-precheck R` BEFORE editing the guide (a refusal means no edit; R5-05); then the drafter edits
# rounds/R/inputs/GUIDE.md; this records the revision, relabels the round's wide cases under the new guide version and re-runs the labels-check. Earlier rounds' gold and memberships are never rewritten.
# R4-02: when the relabel cannot fit the session cap the pre-check (`label_x2d.py --precheck-clarify`, run by `clarify-precheck` and again here) refuses, records create-once rounds/R/clarify-refused.json and
# exits non-zero; the round then proceeds scored and non-informative like a clarified one (require_labels_check), and gold/jev/tune/analysis are unblocked. If the guide was edited before a refusal, restore
# the bytes whose SHA-256 is guide_sha256 of rounds/R/version.json: rounds/<R-1>/inputs/GUIDE.md (round 1: the study GUIDE.md) when round R's inputs were copied from it, otherwise the proposal approved by
# next_inputs_sha256 in reviews/pilot-round-<R-1>.md (its versioned diff is in that critique); `require_version R` must pass before any later stage; keep the refused text under reviews/ (R5-05, R6-04).
stage_clarify_precheck() {
  local r=$1
  need_round "$r"; verify; require_gates; require_critique "$r"
  [ "$r" -le 7 ] || { echo "refusing: a guide clarification is permitted in rounds 1-7 only" >&2; exit 2; }
  [ ! -e "$D/rounds/$r/jev/scores.jsonl" ] || { echo "refusing: round $r already has Jev scores; a clarification must precede scoring" >&2; exit 1; }
  [ ! -e "$D/rounds/$r/gold.jsonl" ] || { echo "refusing: round $r already has gold; a clarification must precede it" >&2; exit 1; }
  py "$S/label_x2d.py" --round "$r" --precheck-clarify
}

stage_clarify() {
  local r=$1 reason=${2:-}
  need_round "$r"; verify; require_gates; require_critique "$r"
  [ "$r" -le 7 ] || { echo "refusing: a guide clarification is permitted in rounds 1-7 only" >&2; exit 2; }
  [ -n "$reason" ] || { echo "usage: run.sh clarify R \"reason\"" >&2; exit 2; }
  [ ! -e "$D/rounds/$r/jev/scores.jsonl" ] || { echo "refusing: round $r already has Jev scores; a clarification must precede scoring" >&2; exit 1; }
  [ ! -e "$D/rounds/$r/gold.jsonl" ] || { echo "refusing: round $r already has gold; a clarification must precede it" >&2; exit 1; }
  py "$S/label_x2d.py" --round "$r" --precheck-clarify   # R3-01: refuses BEFORE a revision is recorded when the relabel does not fit the cap with C + 1; R4-02: the refusal writes create-once rounds/R/clarify-refused.json and exits non-zero
  py "$S/versioning.py" --round "$r" --clarify "$reason"
  local gv; gv=$(pyx "print(x2d.guide_version($r).split('/')[-1])")
  require_version "$r"
  py "$S/label_x2d.py" --round "$r" --new-round "$gv"
  if is_contrast_round "$r"; then py "$S/label_x2d.py" --round "$r" --tag contrast --new-round "$gv"; fi
  py "$S/analysis_x2d.py" --labels-check "$r"
}

stage_jev() {
  local r=$1 R="$D/rounds/$1"; prereq "$r"; require_labels_check "$r"
  jev_run "$r" "$R/units.jsonl" "$D/catalogue/catalogue.json" "$R/jev/scores.jsonl" "sem,sib,title,fan"
  if is_contrast_round "$r"; then
    jev_run "$r" "$R/units.contrast.jsonl" "$D/catalogue/contrast-catalogue.json" "$R/jev/scores.contrast.jsonl" "sem"
  fi
  ensure_states "$r"
  py "$S/reference_arm.py" --round "$r"
}

stage_tune() {
  local r=$1; prereq "$r"; require_labels_check "$r"
  local f; for f in gold.jsonl jev/scores.jsonl; do [ -s "$D/rounds/$r/$f" ] || { echo "refusing: $D/rounds/$r/$f is missing: tune runs after gold and jev (R3-14)" >&2; exit 1; }; done
  py "$S/tune.py" --round "$r"
}

stage_analysis() {
  local r=$1; prereq "$r"; require_labels_check "$r"
  local f; for f in arms/B2.jsonl arms/reference.jsonl jev/scores.jsonl; do [ -s "$D/rounds/$r/$f" ] || { echo "refusing: $D/rounds/$r/$f is missing: analysis needs the baselines and the jev stage (R3-14)" >&2; exit 1; }; done
  if is_contrast_round "$r"; then for f in arms/contrast-B2.jsonl jev/scores.contrast.jsonl; do [ -s "$D/rounds/$r/$f" ] || { echo "refusing: $D/rounds/$r/$f is missing (R3-14)" >&2; exit 1; }; done; fi
  py "$S/compose.py" --round "$r"
  if is_contrast_round "$r"; then py "$S/compose.py" --round "$r" --tag contrast; fi
  py "$S/analysis_x2d.py" --round "$r"
  py "$S/analysis_x2d.py" --round "$r" --pack
}

# rescore-v7 (R2-02): after the round-7 critique (v_7 frozen), re-score rounds 0-7 under round 7's questions, settings and model (sem + fan) for the single round-8 tuning. Round 0: the manifest
# `main` units restricted to D and P; rounds 1-7: the window units. Each unit's excerpt is RE-RENDERED under round 7's `excerpt_lines` from the same member text at the unit's commit (R3-05:
# jev-v7/units.jsonl rows carry `excerpt_lines`; the header's input digest covers them; tune checks them). The spend is charged to the total cap (`x2d.CAPS['usd_total']`, $20 since rev 12), not to the round caps (R3-17). A round whose file has the `.done` marker is complete and is
# skipped; the stage refuses when all eight are done; a killed run resumes (the driver is resumable, append-only), and tune/stage_units re-check headers and completeness.
stage_rescore_v7() {
  verify; require_gates; critique_ok 7
  local rr ndone=0 line units cat flip R7="$D/rounds/7"
  [ -f "$R7/settings.json" ] && [ -d "$R7/inputs/questions" ] || { echo "refusing: round 7's frozen settings and inputs are missing" >&2; exit 1; }
  require_version 7
  for rr in 0 1 2 3 4 5 6 7; do [ -e "$D/rounds/$rr/jev-v7/.done" ] && ndone=$((ndone + 1)) || true; done
  [ "$ndone" -lt 8 ] || { echo "refusing: the v_7 re-score is complete (append-only; it runs once)" >&2; exit 1; }
  JEV_SETTINGS="$R7/settings.json"
  JEV_QUESTIONS="$R7/inputs/questions"
  for rr in 0 1 2 3 4 5 6 7; do
    [ -e "$D/rounds/$rr/jev-v7/.done" ] && continue
    line=$(pyx "
import json, os
rr = int(sys.argv[1]); R = x2d.round_dir(rr); out = os.path.join(R, x2d.V7_RESCORE_SUB, 'units.jsonl')
cat, flip, src = os.path.join(x2d.DATA, 'catalogue', 'catalogue.json'), '0', os.path.join(R, 'units.jsonl')
if rr == 0:
    man = {e['name']: e for e in json.load(open(os.path.join(R, 'jev-in', 'manifest.json')))}['main']
    cat, flip, src = man['catalogue'], '1' if man.get('flip_all') else '0', man['units']
keep = (lambda u: u.get('set') in ('D', 'P') and not u.get('control')) if rr == 0 else (lambda u: u.get('set') == 'window' and not u.get('control'))
if not os.path.exists(out):
    import build_units as BU
    x2d.write_jsonl_new(out, BU.rerender_units(sys.argv[2], [u for u in x2d.read_jsonl(src) if keep(u)], x2d.round_settings(7)['excerpt_lines']))
print('|'.join([out, cat, flip]))
" "$rr" "$SRC")
    IFS='|' read -r units cat flip <<< "$line"
    local extra=()
    if [ "$flip" = 1 ]; then extra=(-flip-all); fi
    jev_run "$rr" "$units" "$cat" "$D/rounds/$rr/jev-v7/scores.jsonl" "sem,fan" ${extra[@]+"${extra[@]}"}
    : > "$D/rounds/$rr/jev-v7/.done"
  done
  JEV_SETTINGS=""
  JEV_QUESTIONS=""
}

# quality: build (report + intruder items) -> label (intruder audit; model) -> finish (kappa and detection rate). Each sub-stage is callable alone.
stage_quality_build() { verify; py "$S/quality.py" --repo "$SRC" --data "$D"; }
stage_quality_label() { check_label_args "$@"; verify; py "$S/versioning.py" --round 0 --inputs; py "$S/label_x2d.py" --round 0 --kind intruder "$@"; }
stage_quality_finish() { verify; py "$S/quality.py" --repo "$SRC" --data "$D" --finish; }

# gates: build (controls, K2/K3 inputs, round-0 version and provisional thresholds, baselines) -> label (wide labels and gold of D/P/controls and the contrast set; model)
# -> jev (the manifest-driven round-0 Jev runs, then compose; model) -> analyse (K2-K6, P1, N1, planted items).
gates_closed() {  # R17-06: after the round-0 gates were analysed, round 0's files are never written again by a legacy stage (the re-gate lives in rounds/0/regate)
  [ ! -s "$D/rounds/0/gates.json" ] || { echo "refusing: $D/rounds/0/gates.json exists: round 0's gates were analysed and its files are never written again (rev 13; the re-gate lives in rounds/0/regate)" >&2; exit 1; }
}
stage_gates_build() {
  verify
  gates_closed
  py "$S/build_controls.py" --repo "$SRC"
  py "$S/build_units.py" --repo "$SRC" --gates
  py "$S/versioning.py" --round 0
  py "$S/tune.py" --round 0
  py "$S/baselines.py" --repo "$SRC" --round 0
  py "$S/baselines.py" --repo "$SRC" --round 0 --tag contrast
}
stage_gates_label() {
  verify
  gates_closed
  require_version 0
  check_label_args "$@"
  py "$S/label_x2d.py" --round 0 "$@"
  py "$S/label_x2d.py" --round 0 --tag contrast "$@"
  py "$S/analysis_x2d.py" --labels-check 0   # §8: round 0's labels-check is report-only (kappa and the rule concentration are recorded; K4 on D is the round-0 kappa gate, R3-02)
  # R4-01: each gold file only when absent/empty (append-only; gold.py itself refuses an existing file): after a fill the contrast gold can complete while the main gold + membership stand
  local g="$D/rounds/0/gold.jsonl" gc="$D/rounds/0/gold.contrast.jsonl"
  if [ -s "$g" ]; then echo "$g exists (append-only): skipped"; else py "$S/gold.py" --round 0; fi
  if [ -s "$gc" ]; then echo "$gc exists (append-only): skipped"; else py "$S/gold.py" --round 0 --tag contrast; fi
}
runs_of() {  # manifest-driven round-0 Jev runs (written by build_controls.py): name|units|catalogue|out|arms|flip_all
  pyx "import json; [print('|'.join([e['name'], e['units'], e['catalogue'], e['out'], e['arms'], '1' if e.get('flip_all') else '0'])) for e in json.load(open('$D/rounds/0/jev-in/manifest.json'))]"
}
stage_gates_jev() {
  verify
  gates_closed
  require_version 0; require_labels_check 0
  local R0="$D/rounds/0" name units cat out arms flip extra
  runs_of > "$R0/jev-in/.runs"
  while IFS='|' read -r name units cat out arms flip; do
    extra=()
    if [ "$flip" = 1 ]; then extra=(-flip-all); fi
    jev_run 0 "$units" "$cat" "$R0/$out" "$arms" ${extra[@]+"${extra[@]}"}
  done < "$R0/jev-in/.runs"
  while IFS='|' read -r name units cat out arms flip; do
    case "$name" in
      main) py "$S/compose.py" --round 0 --units "$units" --catalogue "$cat" ;;   # the main run answers the D / P units and the dup, trap, prose and planted_new controls only (p1 / p2 have their own runs); the fan-out must have answered every area of the run's catalogue (R3-06)
      *) py "$S/compose.py" --round 0 --scores "$R0/$out" --units "$units" --arms-dir "arms-$name" --catalogue "$cat" ;;
    esac
  done < "$R0/jev-in/.runs"
}
stage_gates_analyse() { verify; py "$S/analysis_x2d.py" --gates --repo "$SRC"; }

# ---- the one permitted re-gate of P1 and P2 on fresh items (PREREGISTRATION 5, rev 13; R-053). It lives in $D/rounds/0/regate, a self-contained data root (catalogue/ and rounds/0/ with the frozen Jev side
# copied byte for byte): every stage after the build runs with D and X2D_DATA re-pointed at it, so nothing outside it can be written. One shot: regate-build refuses when the directory exists, regate-label and
# regate-jev refuse once regate/gates.json exists, regate-analyse writes it once. Jev spend is charged to the round-0 cap and the $20 total (x2d.jev_budget maps the re-gate root to the study data directory).
in_regate() { ( D="$D/rounds/0/regate"; export X2D_DATA="$D"; "$@" ); }
regate_ready() {
  [ -s "$D/rounds/0/regate/regate.json" ] || { echo "refusing: $D/rounds/0/regate/regate.json missing: run.sh regate-build first" >&2; exit 1; }
  pyx "bad = x2d.regate_frozen_problems(); sys.exit('refusing: ' + '; '.join(bad)) if bad else 0"
}
regate_open() {
  regate_ready
  [ ! -e "$D/rounds/0/regate/gates.json" ] || { echo "refusing: $D/rounds/0/regate/gates.json exists: the re-gate was analysed, it is one shot and no further re-gate stage runs" >&2; exit 1; }
}
regate_inner_baselines() { py "$S/baselines.py" --repo "$SRC" --round 0; }
stage_regate_build() {
  verify
  [ ! -e "$D/rounds/0/regate" ] || { echo "refusing: $D/rounds/0/regate exists: the re-gate is one shot" >&2; exit 1; }
  py "$S/build_controls.py" --repo "$SRC" --regate
  in_regate regate_inner_baselines
}
regate_inner_label() {
  py "$S/label_x2d.py" --round 0 "$@"
  local g="$D/rounds/0/gold.jsonl"
  if [ -s "$g" ]; then echo "$g exists (append-only): skipped"; else py "$S/gold.py" --round 0; fi
}
stage_regate_label() { verify; regate_open; check_label_args "$@"; in_regate regate_inner_label "$@"; }
regate_inner_jev() {
  local R0="$D/rounds/0" name units cat out arms flip extra
  [ -s "$R0/gold.jsonl" ] || { echo "refusing: $R0/gold.jsonl missing: the re-gate's labels and gold come before its Jev run" >&2; exit 1; }
  py "$S/analysis_x2d.py" --regate-precheck || { echo "refusing: the re-gate's label-only denominators or construction checks fail (no Jev request was made): stop, report; the council and the user decide (PREREGISTRATION 5)" >&2; exit 1; }
  runs_of > "$R0/jev-in/.runs"
  while IFS='|' read -r name units cat out arms flip; do
    extra=()
    if [ "$flip" = 1 ]; then extra=(-flip-all); fi
    jev_run 0 "$units" "$cat" "$R0/$out" "$arms" ${extra[@]+"${extra[@]}"}
  done < "$R0/jev-in/.runs"
  while IFS='|' read -r name units cat out arms flip; do
    py "$S/compose.py" --round 0 --scores "$R0/$out" --units "$units" --arms-dir "arms-$name" --catalogue "$cat"
  done < "$R0/jev-in/.runs"
}
stage_regate_jev() { verify; regate_open; in_regate regate_inner_jev; }
regate_inner_analyse() { py "$S/analysis_x2d.py" --regate-gates; }
stage_regate_analyse() { verify; regate_open; in_regate regate_inner_analyse; }

# power (model-free): P3 / N3 / N3-PR / N4 cells on the live frozen-round structure and the real-none proxy (R2-11); read by the council's round-0 critique; exit 3 when a null cell is above its bar (R5-02)
stage_power() { verify; [ -s "$D/rounds/0/k23.json" ] || { echo "refusing: $D/rounds/0/k23.json missing: run gates first" >&2; exit 1; }; py "$S/power_sim.py" --structure-k23 "$D/rounds/0/k23.json" --out "$D/power.json"; }

STAGE=${1:-}
case "$STAGE" in
  validate) exec "$PY" "$S/validation/validate.py" ;;
  onboard-collect)
    verify
    py "$S/collect_onboard.py" --source "${2:-$SRC}" --data "$D" ;;
  onboard)
    verify
    mkdir -p "$D/onboard-jev"
    budget=$(pyx "print(x2d.jev_budget(None, '$D/onboard-jev/scores.jsonl'))")
    go run "./$S/onboard" -draws "$D/onboard/draws" -scopes "$S/prepass/scopes.json" -repo "$SRC" -rev "$C0" -out "$D/catalogue" -contrast "$D/onboard/contrast" \
      -jev-out "$D/onboard-jev/scores.jsonl" -questions-src "$S/questions" -workers 8 -max-cost-usd "$budget" ${X2D_JEV_ENDPOINT:+-endpoint "$X2D_JEV_ENDPOINT"} ;;
  quality) stage_quality_build; stage_quality_label; stage_quality_finish ;;
  quality-build|quality-label|quality-finish) "stage_${STAGE//-/_}" "${@:2}" ;;
  gates) stage_gates_build; stage_gates_label; stage_gates_jev; stage_gates_analyse ;;
  regate-build|regate-label|regate-jev|regate-analyse) "stage_${STAGE//-/_}" "${@:2}" ;;
  gates-build|gates-label|gates-jev|gates-analyse) "stage_${STAGE//-/_}" "${@:2}" ;;
  round) need_round "${2:-}"; for s in units label labels_check gold baselines jev tune analysis; do "stage_$s" "$2"; done ;;
  labels-check) need_round "${2:-}"; stage_labels_check "$2" ;;
  clarify-precheck) need_round "${2:-}"; stage_clarify_precheck "$2" ;;
  clarify) need_round "${2:-}"; stage_clarify "$2" "${3:-}" ;;
  rescore-v7) stage_rescore_v7 ;;
  power) stage_power ;;
  units|baselines|label|gold|jev|tune|analysis) need_round "${2:-}"; "stage_$STAGE" "$2" "${@:3}" ;;
  verdict)
    verify
    critique_ok 10
    py "$S/analysis_x2d.py" --verdict ;;
  *) sed -n '2,15p' "$S/run.sh"; exit 2 ;;
esac
