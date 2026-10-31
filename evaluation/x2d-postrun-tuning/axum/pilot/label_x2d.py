#!/usr/bin/env python3
"""Agent labelling of X2d units (CLAUDE.md 'Jev labelling' protocol; PREREGISTRATION section 8). Adapted from ../x2-area-stability/label.py.

A = `claude -p` (default model, no tools), B = `codex exec` (read-only sandbox, empty working directory, no config/rules), C = `claude -p --model sonnet`
adjudicator, only when A and B disagree (per `agree`) or fail. Every session is ephemeral, runs with Scout's credential scrub (common.scrub_env) and sees only
the evidence built from an explicit allow-list of fields plus the round's frozen guide (data/rounds/<r>/inputs/GUIDE.md, R2-01; its version is x2d.guide_version(r)). Real units, controls and the D/P
sets are labelled indistinguishably: the evidence never carries arm outputs, siblings, stratum, score, round, control or set flags.

Kinds
  wide (default)  label_evidence + the catalogue (up to 20 owned files per area) -> {covered, acceptable<=4, primary, none_kind, rule, reason}. The catalogue shown for a
                  unit is its designated one: the main catalogue, or for a round-0 control the `catalogue` of its control-key.jsonl row (catalogue-p1, catalogue-p2, ...);
                  the control kind is never shown. With --tag contrast: the contrast catalogue; the units are D/rounds/0/units.contrast.jsonl in round 0 and
                  D/rounds/R/units.contrast_window.jsonl (<= 30 hash-sampled window units, written by build_units.py) in rounds >= 1; output labels.contrast.jsonl.
  jev-state       >= --perception-n (20) scored window units (set window, not a control; smallest sha256('x2d-perception\\0'+case_id)), each asked both Jev questions from the
                  state `jev -dump-state` wrote to --states (default D/rounds/R/states.jsonl; the V-sem row: the exact merge.covered state, the merge.which_area state and its
                  option texts); no label guide is shown, only the question instructions and criteria. Rows carry `question` (merge.covered -> covered; merge.which_area -> primary)
  intruder        items of D/onboard/intruder-items.jsonl -> {"intruder": path|null}; output D/onboard/{labels,calls}.intruder.jsonl, lock labels.intruder.lock
Files (round dir D/rounds/<r>/): units[.<tag>].jsonl in; labels[.<tag>].jsonl, calls[.<tag>].jsonl, labels[.<tag>].lock, sessions.json out. All append-only except sessions.json
(the latest {cap, used, plan}); a case's label rows are written first, then its call row, under a lock. Refuses when labels of this guide version and kind already exist unless
--new-round NAME (which labels only the cases without a usable final label). Session cap (R1-05, R3-01): cap_for(r, W, J, R, I, C) = ceil(1.25 * (2.2 * W * (1 + C) + 4.4 * J + R + 2.2 * I)), ceiling 500,
W = wide units of the round (units.jsonl + the contrast units), J = perception units, R = reference clusters, I = intruder items (round 0), C = recorded guide clarifications of the round (<= 1); a slot is reserved before every session
attempt; the planned 2 sessions per case must fit in the cap minus the sessions already in the round's calls*.jsonl (round 0: and in D/onboard/calls.intruder.jsonl); a case that
hits the cap mid-way writes its call row only (no labels). A codex session that ran any command (`exec` events) is unparsable: its answer is quarantined in the call row
(`quarantined`), never used, never retried, and the case goes to the adjudicator.

Usage: label_x2d.py --round R [--kind wide|jev-state|intruder] [--tag contrast] [--workers 4] [--new-round NAME] [--catalogue FILE] [--perception-n 20] [--states FILE]
       X2D_LABEL_STUB=<truth.json> label_x2d.py ...   validation only: deterministic stub sessions, no model call.
         truth = {case_id: {"A": x, "B": x, "C": x}} (jev-state keys: "<case_id>|merge.covered" / "<case_id>|merge.which_area"); x = answer dict | None
         (unparsable session); an answer dict may carry "omit": [fields], "extra": {k: v} and, for B (the codex session), "exec": N (N exec events).
       X2D_STUB_SESSION_CAP=N  honoured by cap_for only while that caller's own stub env is set (label_x2d: X2D_LABEL_STUB; reference_arm: X2D_REFERENCE_STUB); ignored otherwise.
       X2D_LABEL_PROMPT_LOG=<file>  (stub only) appends {"cid","who","prompt"} per session prompt (blindness probe).
       X2D_SKIP_SIGNOFF=1  verify_pins(require_signoff=False) (validation only).
"""
import argparse
import hashlib
import json
import os
import re
import subprocess
import sys
import tempfile
import threading
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime, timezone

import x2d
import common as C  # X2's frozen common (scrub_env); X2's GUIDE.md and labels are NOT used

_ROUND = {"r": None, "data": None}  # the round whose inputs (guide) the prompts use; set_round() is called by main() and by every importer before building prompts
now = lambda: datetime.now(timezone.utc).isoformat(timespec="seconds")
PERCEPTION_Q = ("merge.covered", "merge.which_area")
SESSION_CEILING = 500  # hard ceiling of cap_for (R1-05)
CONTRAST_WINDOW_MAX = 30  # window units labelled against the contrast catalogue in rounds 1 and 10
WIDE_KEYS = ("covered", "acceptable", "primary", "none_kind", "rule", "reason")
EVIDENCE_FIELDS = {  # allow-list of label_evidence fields
    "unit": ("paths", "change_kind", "role", "members"), "pr": ("title", "body", "changed_paths"),
    "diff_summary": ("files", "added_lines", "deleted_lines", "top_symbols")}


# ---------------------------------------------------------------- evidence (allow-list only)
def allow(src, spec):
    src = src or {}
    out = {}
    for k, fields in spec.items():
        d = src.get(k) or {}
        out[k] = {f: d.get(f) for f in fields}
    if isinstance(out["unit"].get("members"), list):
        out["unit"]["members"] = [{"path": m.get("path"), "text": m.get("text")} for m in out["unit"]["members"]]
    return out


def wide_catalogue(cat):
    return [{"id": a["area_id"], "title": a.get("title"), "summary": a.get("summary"), "files": list(a.get("files", []))[:20]} for a in cat["areas"]]


def set_round(r, data=None):
    _ROUND["r"], _ROUND["data"] = r, data


def guide_text():
    if _ROUND["r"] is None:
        raise RuntimeError("label_x2d.set_round(r) must be called before a prompt is built")
    return open(x2d.guide_path(_ROUND["r"], _ROUND["data"])).read()


def perception_pick(units, n):
    """The n scored window units (set window, not a control, no audit/D/P items) with the smallest sha256('x2d-perception\\0' + case_id)."""
    pool = [u for u in units if u.get("set", "window") == "window" and not u.get("control")]
    return sorted(pool, key=lambda u: hashlib.sha256(("x2d-perception\0" + u["case_id"]).encode()).hexdigest())[:n]


# ---------------------------------------------------------------- parsing
def _json(txt):
    m = re.search(r"\{.*\}", (txt or "").strip(), re.S)
    if not m:
        return None
    try:
        o = json.loads(m.group(0))
    except Exception:  # noqa: BLE001
        return None
    return o if isinstance(o, dict) else None


def _ids(o, ids):
    a = o.get("acceptable")
    if not isinstance(a, list) or not all(isinstance(x, str) for x in a):
        return None
    out = []
    for x in a:
        if x in ids and x not in out:
            out.append(x)
    return out


def _answer(covered=None, acceptable=(), primary=None, none_kind=None, rule="", reason=""):
    return {"covered": covered, "acceptable": list(acceptable), "primary": primary, "none_kind": none_kind, "rule": rule, "reason": reason}


def _txt(o, k, n):
    return str(o.get(k, "") or "")[:n]


def parse_wide(txt, ctx):
    o = _json(txt)
    if o is None or not isinstance(o.get("covered"), bool):
        return None
    rule, reason = _txt(o, "rule", 200), _txt(o, "reason", 1000)
    if not o["covered"]:
        nk = o.get("none_kind")
        if nk not in ("gap", "new", None):
            return None
        return _answer(False, (), None, nk, rule, reason)
    acc = _ids(o, ctx["ids"])
    if acc is None or len(acc) > 4 or not acc:  # unknown ids are dropped; none left -> unparsable
        return None
    p = o.get("primary")
    if p is None and len(acc) == 1:
        p = acc[0]
    if p not in acc:
        return None
    return _answer(True, acc, p, None, rule, reason)


def parse_q1(txt, ctx):
    o = _json(txt)
    if o is None or not isinstance(o.get("covered"), bool):
        return None
    return _answer(o["covered"], (), None, None, _txt(o, "rule", 200), _txt(o, "reason", 1000))


def parse_q2(txt, ctx):
    o = _json(txt)
    if o is None or o.get("primary") not in ctx["ids"]:
        return None
    return _answer(True, [o["primary"]], o["primary"], None, _txt(o, "rule", 200), _txt(o, "reason", 1000))


def parse_intruder(txt, ctx):
    o = _json(txt)
    if o is None or "intruder" not in o or not (o["intruder"] is None or o["intruder"] in ctx["files"]):
        return None
    return {"intruder": o["intruder"], "rule": "", "reason": _txt(o, "reason", 1000)}


# ---------------------------------------------------------------- agreement
def agree(a, b, sizes=None):
    """Contract: ("agreed", row) | ("disagree", None); a, b = parsed wide answers or None (unparsable). sizes = {area id: catalogue file count} (tie-break).
    Covered: common = acceptable(A) & acceptable(B) must be non-empty; primary = A's primary if in common, else B's if in common, else the common id with the fewest
    catalogue files (tie: lexicographic); `primary_rule` records which ("both" when A and B name the same primary, "A", "B", "fewest_files")."""
    if not a or not b or a["covered"] != b["covered"]:
        return "disagree", None
    if a["covered"]:
        common = sorted(set(a["acceptable"]) & set(b["acceptable"]))
        if not common:
            return "disagree", None
        if a["primary"] in common:
            p, prule = a["primary"], "both" if a["primary"] == b["primary"] else "A"
        elif b["primary"] in common:
            p, prule = b["primary"], "B"
        else:
            sz = sizes or {}
            p, prule = min(common, key=lambda i: (sz.get(i, float("inf")), i)), "fewest_files"
        return "agreed", {"covered": True, "acceptable": common, "primary": p, "primary_rule": prule, "none_kind": None, "rule": a["rule"], "reason": a["reason"]}
    if a["none_kind"] != b["none_kind"]:
        return "disagree", None
    return "agreed", {"covered": False, "acceptable": [], "primary": None, "primary_rule": None, "none_kind": a["none_kind"], "rule": a["rule"], "reason": a["reason"]}


def agree_q1(a, b):
    if a and b and a["covered"] == b["covered"]:
        return "agreed", dict(a)
    return "disagree", None


def agree_intruder(a, b):
    if a and b and a["intruder"] == b["intruder"]:
        return "agreed", dict(a)
    return "disagree", None


# ---------------------------------------------------------------- prompts
FMT_WIDE = ('Answer with ONLY one JSON object: {"covered": true|false, "acceptable": ["<area id>", ...] (at most 4; empty when covered is false), '
            '"primary": "<area id>"|null (one of acceptable), "none_kind": "gap"|"new"|null (only when covered is false), "rule": "<guide rule id(s)>", '
            '"reason": "<=2 sentences"}')
FMT_Q1 = 'Answer with ONLY one JSON object: {"covered": true|false, "rule": "<guide rule id(s)>", "reason": "<=2 sentences"}'
FMT_Q2 = 'Answer with ONLY one JSON object: {"primary": "<one area id from the catalogue>", "rule": "<guide rule id(s)>", "reason": "<=2 sentences"}'
FMT_INTRUDER = 'Answer with ONLY one JSON object: {"intruder": "<one of the listed file paths>"|null, "reason": "<=2 sentences"}'


def intro(case):
    kind = case["kind"]
    if kind == "wide":
        return ("You are labelling one unit (files added or moved by a merged change). Apply the label guide below to the evidence. Use only the evidence; "
                "you have no tools and no other context.\n\n=== LABEL GUIDE ===\n" + guide_text() + "\n\n=== EVIDENCE (JSON) ===\n")
    if kind == "jev-state":
        ins, crit = case["instructions"]  # the effective instructions/criteria of the V-sem question, from the dumped state (R2-17)
        return ("You are answering one question from exactly the state shown, as an automated reader would. Use only that state; you have no tools and no other context.\n\n"
                "=== QUESTION ===\n" + ins + "\n" + ("Criteria: " + json.dumps(crit) + "\n" if crit else "") + "\n=== STATE (JSON) ===\n")
    return ("One file in the list below does not belong to the area described by the title and summary; the other files share one responsibility. "
            "Use only this evidence; you have no tools and no other context.\n\n=== EVIDENCE (JSON) ===\n")


def fmt_for(case):
    return {"wide": FMT_WIDE, "intruder": FMT_INTRUDER}.get(case["kind"]) or (FMT_Q1 if case["question"] == "merge.covered" else FMT_Q2)


def prompt_ab(case):
    return intro(case) + json.dumps(case["evidence"], indent=1) + "\n\n" + fmt_for(case)


def prompt_c(case, a, b):
    keys = ("intruder", "reason") if case["kind"] == "intruder" else WIDE_KEYS
    ev = lambda x: json.dumps({k: x.get(k) for k in keys}) if x else "unparsable (no valid answer)"
    return (intro(case) + json.dumps(case["evidence"], indent=1) + "\n\n=== TWO INDEPENDENT LABELLERS DISAGREED (or failed) ===\n"
            "Labeller A: " + ev(a) + "\nLabeller B: " + ev(b) + "\n\nYou are the adjudicator. Decide by the guide yourself.\n" + fmt_for(case))


# ---------------------------------------------------------------- sessions
def session_env():
    """Scout's credential scrub (common.scrub_env; CLAUDE.md model-use rule): a key or base URL in the shell can neither switch to API billing nor re-route."""
    return C.scrub_env(os.environ)


def run_claude(p, model, cid=None):
    cmd = ["claude", "-p", "--tools", "", "--no-session-persistence", "--setting-sources", "", "--output-format", "json"]
    cmd += ["--model", model] if model else []
    with tempfile.TemporaryDirectory() as d:
        r = subprocess.run(cmd, input=p, capture_output=True, text=True, cwd=d, timeout=600, env=session_env())
    j = json.loads(r.stdout)
    return j["result"], ",".join(j.get("modelUsage", {}).keys())


def run_codex(p, model, cid=None):
    """-> (answer, model, exec events): a session that ran any command (read-only sandbox can read the disk) is unparsable (`call` quarantines it)."""
    with tempfile.TemporaryDirectory() as d:
        out = os.path.join(d, "last.txt")
        r = subprocess.run(["codex", "exec", "-s", "read-only", "--skip-git-repo-check", "--ephemeral", "--ignore-user-config", "--ignore-rules",
                            "-C", d, "-o", out, "-"], input=p, capture_output=True, text=True, cwd=d, timeout=600, env=session_env())
        txt = open(out).read()
        m = re.search(r"^model:\s*(\S+)", r.stdout + r.stderr, re.M)
        execs = len(re.findall(r"^exec\b", r.stdout + r.stderr, re.M))
    return txt, m.group(1) if m else "unknown", execs


def make_stub(truth):
    """truth: {cid: {"A": x, "B": x, "C": x}}, x = answer dict | None (unparsable); dict may hold "omit": [fields], "extra": {k: v}."""
    log = os.environ.get("X2D_LABEL_PROMPT_LOG")
    lock = threading.Lock()

    def runner_for(which):
        def run(p, model, cid=None):
            if log:
                with lock, open(log, "a") as f:
                    f.write(json.dumps({"cid": cid, "who": which, "prompt": p}) + "\n")
            x = truth[cid][which]
            if x is None:
                return ("not json", "stub-" + which, 0) if which == "B" else ("not json", "stub-" + which)
            o = {k: v for k, v in x.items() if k not in ("omit", "extra", "exec")}
            o.setdefault("rule", "stub")
            o.setdefault("reason", "stub")
            for k in x.get("omit", []):
                o.pop(k, None)
            o.update(x.get("extra", {}))
            if which == "B":  # the codex runner reports its exec events (a stub session with an `exec` line)
                return json.dumps(o), "stub-" + which, int(x.get("exec", 0))
            return json.dumps(o), "stub-" + which
        return run
    return runner_for("A"), runner_for("B"), runner_for("C")


def call(runner, model, p, case, reserve):
    """One labeller: up to 2 sessions, a slot reserved from `reserve(1)` before EVERY attempt (False: the cap is reached, the result is {"capped": True} with the attempts made).
    A session with exec events (codex ran a command) is quarantined (`quarantined`: attempt, exec count, raw text), its answer is never parsed, and it is not retried."""
    parse = case["parse"]
    resolved, txt, execs, sessions, quarantined = "unknown", "", None, 0, []
    for attempt in (1, 2):
        if not reserve(1):
            return {"ok": False, "ans": None, "model": str(resolved), "attempts": sessions, "exec_events": execs, "quarantined": quarantined, "capped": True,
                    "raw": txt[:500], "timestamp": now()}
        sessions += 1
        ex = 0
        try:
            res = runner(p, model, case["cid"])
            txt, resolved = res[0], res[1]
            if len(res) > 2:
                ex = int(res[2] or 0)
                execs = (execs or 0) + ex
        except Exception as e:  # noqa: BLE001 - a failed session is recorded, never fabricated
            txt, resolved = "ERROR %r" % e, "unknown"
        if ex > 0:
            quarantined.append({"attempt": attempt, "exec_events": ex, "raw": (txt or "")[:500]})
            break
        o = parse(txt, case["ctx"])
        if o:
            return {"ok": True, "ans": o, "model": str(resolved), "attempts": sessions, "exec_events": execs, "quarantined": quarantined, "timestamp": now()}
    return {"ok": False, "ans": None, "model": str(resolved), "attempts": sessions, "exec_events": execs, "quarantined": quarantined, "raw": txt[:500], "timestamp": now()}


# ---------------------------------------------------------------- cases
def _units_file(rd, r, tag):
    if tag == "contrast":
        return os.path.join(rd, "units.contrast.jsonl" if r == 0 else "units.contrast_window.jsonl")
    return os.path.join(rd, "units.jsonl")


def load_states(path):
    """states.jsonl of `jev -dump-state` -> {case_id: the V-sem row}."""
    if not os.path.exists(path):
        sys.exit("error: %s is missing: run `jev -dump-state -in units.jsonl -catalogue catalogue.json -settings settings.json -out %s` first" % (path, path))
    return {r["case_id"]: r for r in x2d.read_jsonl(path) if r.get("variant") == "V-sem"}


def build_cases(a, rd):
    """-> (cases, paths). case = {cid, case_id, kind, question, evidence, parse, agree, ctx, label_set}."""
    suffix = "." + a.tag if a.tag else ""
    if a.kind == "intruder":
        base = x2d.data_path("onboard")
        paths = {"dir": base, "labels": os.path.join(base, "labels.intruder.jsonl"), "calls": os.path.join(base, "calls.intruder.jsonl"),
                 "lock": os.path.join(base, "labels.intruder.lock")}
        cases = []
        for it in x2d.read_jsonl(os.path.join(base, "intruder-items.jsonl")):
            cid = it.get("case_id") or it["item_id"]
            files = list(it["files"])
            ev = {"area_title": it.get("area_title", it.get("title")), "area_summary": it.get("area_summary", it.get("summary")), "files": files}
            cases.append({"cid": cid, "case_id": cid, "kind": "intruder", "question": None, "evidence": ev, "parse": parse_intruder, "agree": agree_intruder,
                          "ctx": {"files": set(files)}, "label_set": "main"})
        return cases, paths
    paths = {"dir": rd, "labels": os.path.join(rd, "labels%s.jsonl" % suffix), "calls": os.path.join(rd, "calls%s.jsonl" % suffix),
             "lock": os.path.join(rd, "labels%s.lock" % suffix)}
    default_cat = "contrast-catalogue.json" if a.tag == "contrast" else "catalogue.json"
    cat_path = a.catalogue or os.path.join(x2d.data_path("catalogue"), default_cat)
    cat = json.load(open(cat_path))
    ids = {x["area_id"] for x in cat["areas"]}
    ufile = _units_file(rd, a.round, a.tag)
    units = x2d.read_jsonl(ufile)
    if not units:
        sys.exit("error: no units in %s" % ufile)
    if a.tag == "contrast" and a.round > 0 and len(units) > CONTRAST_WINDOW_MAX:
        sys.exit("error: %s has %d units; the contrast window sample is <= %d" % (ufile, len(units), CONTRAST_WINDOW_MAX))
    label_set = "contrast" if a.tag == "contrast" else "main"
    cases = []
    if a.kind == "wide":
        # a round-0 control is labelled against its designated catalogue (control-key.jsonl `catalogue`); the control kind is never shown
        designated = {}
        if not a.tag:
            designated = {k["case_id"]: k.get("catalogue", "catalogue") for k in x2d.read_jsonl(os.path.join(rd, "control-key.jsonl"))}
        shown = {"catalogue": (wide_catalogue(cat), ids, {x["area_id"]: len(x.get("files", [])) for x in cat["areas"]})}

        def catalogue_of(name):
            if name not in shown:
                p = os.path.join(rd, name + ".json")
                if not os.path.exists(p):
                    sys.exit("error: the control catalogue %s is missing" % p)
                c = json.load(open(p))
                shown[name] = (wide_catalogue(c), {x["area_id"] for x in c["areas"]}, {x["area_id"]: len(x.get("files", [])) for x in c["areas"]})
            return shown[name]

        for u in units:
            wc, cids, sizes = catalogue_of(designated.get(u["case_id"], "catalogue"))
            ev = allow(u.get("label_evidence"), EVIDENCE_FIELDS)
            ev["catalogue"] = wc
            cases.append({"cid": u["case_id"] + ("|contrast" if a.tag == "contrast" else ""), "case_id": u["case_id"], "kind": "wide", "question": None, "evidence": ev, "parse": parse_wide,
                          "agree": (lambda x, y, sizes=sizes: agree(x, y, sizes)), "ctx": {"ids": cids}, "label_set": label_set})
        return cases, paths
    states = load_states(a.states or os.path.join(rd, "states.jsonl"))
    picked = perception_pick(units, a.perception_n)
    if not picked:
        sys.exit("error: no scored window units in %s for the perception check" % ufile)
    for u in picked:
        s = states.get(u["case_id"])
        if s is None:
            sys.exit("error: %s has no V-sem state for %s" % (a.states or os.path.join(rd, "states.jsonl"), u["case_id"]))
        options = s.get("choice_criteria") or s["choice_options"]
        cids = set(s["choice_criteria"]) if s.get("choice_criteria") else ids
        for q in PERCEPTION_Q:
            ev = s["noul_state"] if q == "merge.covered" else {"state": s["choice_state"], "options": options}
            ins, crit = (s["noul_instructions"], s.get("noul_criteria")) if q == "merge.covered" else (s["choice_instructions"], None)
            cases.append({"cid": u["case_id"] + "|" + q, "case_id": u["case_id"], "kind": "jev-state", "question": q, "evidence": ev, "instructions": (ins, crit),
                          "parse": parse_q1 if q == "merge.covered" else parse_q2, "agree": agree_q1 if q == "merge.covered" else agree,
                          "ctx": {"ids": cids}, "label_set": label_set})
    return cases, paths


def session_count(rd, extra_dir=None):
    n = 0
    d = extra_dir or rd
    if not d or not os.path.isdir(d):
        return 0
    for fn in os.listdir(d):
        if fn.startswith("calls") and fn.endswith(".jsonl"):
            for r in x2d.read_jsonl(os.path.join(d, fn)):
                n += sum((r[k] or {}).get("attempts", 0) for k in ("A", "B", "C") if r.get(k))
    return n


def used_sessions(r, rd):
    """Sessions already spent in round r: every calls*.jsonl of its directory (labels, contrast, reference) and, in round 0, the onboarding intruder calls."""
    n = session_count(rd)
    if r == 0:
        n += session_count(None, x2d.data_path("onboard"))
    return n


def plan_for(r, rd=None, perception_n=20):
    """The cap inputs of round r read from its files: W wide units (units.jsonl + the contrast units: round 0 units.contrast.jsonl, later rounds units.contrast_window.jsonl),
    J perception units, R reference clusters (rounds >= 1), I intruder items (round 0), C recorded guide clarifications (rows of guide-revisions.jsonl, R3-01; at most 1)."""
    rd = rd or x2d.round_dir(r)
    units = x2d.read_jsonl(os.path.join(rd, "units.jsonl"))
    C = len(x2d.read_jsonl(os.path.join(rd, "guide-revisions.jsonl")))
    if r == 0:
        return {"W": len(units) + len(x2d.read_jsonl(os.path.join(rd, "units.contrast.jsonl"))), "J": 0, "R": 0,
                "I": len(x2d.read_jsonl(os.path.join(x2d.data_path("onboard"), "intruder-items.jsonl"))), "C": C}
    window = [u for u in units if u.get("set", "window") == "window" and not u.get("control")]
    return {"W": len(units) + len(x2d.read_jsonl(os.path.join(rd, "units.contrast_window.jsonl"))), "J": min(len(window), perception_n),
            "R": min(len({u["cluster_id"] for u in window}), x2d.REFERENCE_SUBSET), "I": 0, "C": C}


def cap_for(r, W=0, J=0, R=0, I=0, C=0, stub_flag="X2D_LABEL_STUB"):
    """Session cap of round r: ceil(1.25 * (2.2 * W * (1 + C) + 4.4 * J + R + 2.2 * I)) (integer arithmetic: (22 W (1 + C) + 44 J + 10 R + 22 I) / 8), ceiling 500; C = recorded guide
    clarifications of the round (each relabels every wide case, 2.2 sessions per case; R3-01). X2D_STUB_SESSION_CAP overrides it
    ONLY while the caller's own stub env `stub_flag` is set (label_x2d: X2D_LABEL_STUB; reference_arm: X2D_REFERENCE_STUB, R2-15); a real run never sets a stub."""
    if os.environ.get(stub_flag) and os.environ.get("X2D_STUB_SESSION_CAP"):
        return int(os.environ["X2D_STUB_SESSION_CAP"])
    return min(SESSION_CEILING, -(-(22 * W * (1 + C) + 44 * J + 10 * R + 22 * I) // 8))


def clarify_precheck(r):
    """R3-01 (2): refuse BEFORE a guide revision is recorded when relabelling the round's wide cases (units.jsonl and, in a contrast round, the contrast units: 2 sessions each, adjudications
    and retries extra) would not fit in the cap computed with one more clarification (C + 1) minus the sessions already used. R4-02: the refusal writes create-once
    rounds/R/clarify-refused.json {utc, reason, cap, used, planned}: the round then proceeds scored and non-informative like a clarified one (run.sh require_labels_check, analysis --labels-check) and exits non-zero."""
    rd = x2d.round_dir(r)
    refused = os.path.join(rd, "clarify-refused.json")
    plan = plan_for(r, rd)
    plan["C"] += 1
    cap, used = cap_for(r, **plan), used_sessions(r, rd)
    n = len(x2d.read_jsonl(os.path.join(rd, "units.jsonl")))
    if r in x2d.CONTRAST_ROUNDS:
        n += len(x2d.read_jsonl(os.path.join(rd, "units.contrast_window.jsonl")))
    if used + 2 * n > cap:
        msg = "a clarification of round %d cannot be relabelled within the session cap: %d used + %d planned (2 per case) > cap %d with C = %d (plan %s); no revision recorded" % (r, used, 2 * n, cap, plan["C"], plan)
        if os.path.exists(refused):  # create-once: the first refusal stays the record
            sys.exit("error: " + msg + "; %s already refused it (exists, create-once)" % refused)
        x2d.write_json_new(refused, {"utc": now(), "reason": msg, "cap": cap, "used": used, "planned": 2 * n})
        sys.exit("error: " + msg + "; wrote " + refused)
    print("clarification of round %d fits the session cap: %d used + %d planned <= %d" % (r, used, 2 * n, cap))




def write_sessions(r, rd, cap, plan):
    """D/rounds/R/sessions.json = the latest {cap, used, plan} (overwritten; analysis copies it into results.json)."""
    os.makedirs(rd, exist_ok=True)
    rec = {"round": r, "cap": cap, "used": used_sessions(r, rd), "plan": plan, "ceiling": SESSION_CEILING,
           "formula": "ceil(1.25 * (2.2 * W * (1 + C) + 4.4 * J + R + 2.2 * I)), ceiling %d" % SESSION_CEILING}
    tmp = os.path.join(rd, "sessions.json.tmp")
    with open(tmp, "w") as f:
        json.dump(rec, f, indent=1, sort_keys=True)
    os.replace(tmp, os.path.join(rd, "sessions.json"))
    return rec


# ---------------------------------------------------------------- main
def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--round", type=int)
    ap.add_argument("--kind", choices=("wide", "jev-state", "intruder"), default="wide")
    ap.add_argument("--tag")
    ap.add_argument("--workers", type=int, default=4)
    ap.add_argument("--new-round")
    ap.add_argument("--catalogue")
    ap.add_argument("--perception-n", type=int, default=20)
    ap.add_argument("--states")
    ap.add_argument("--cases", help="comma-separated case ids: label only these (e.g. an earlier round's units 'for the record' under another round's guide, with --guide-round; never absorbed)")
    ap.add_argument("--guide-round", type=int, help="use the guide of this round's inputs (default: --round); rows carry that guide version")
    ap.add_argument("--precheck-clarify", action="store_true", help="R3-01: only check that relabelling the round's wide cases fits the cap with one more clarification (C + 1); writes rounds/R/clarify-refused.json (create-once) and exits non-zero when not")
    a = ap.parse_args(argv)
    if a.kind == "intruder" and a.round is None:
        a.round = 0
    if a.round is None:
        ap.error("--round is required")
    if a.tag not in (None, "contrast"):
        ap.error("--tag must be contrast")
    if a.kind == "intruder" and a.tag:
        ap.error("--tag does not apply to --kind intruder")
    if a.kind == "jev-state" and a.perception_n < 20:
        ap.error("--perception-n must be >= 20")
    bad = x2d.verify_pins(require_signoff=x2d.require_signoff())
    if bad:
        sys.exit("pin check failed:\n" + "\n".join(bad))
    if a.precheck_clarify:
        clarify_precheck(a.round)
        return
    rd = x2d.round_dir(a.round)
    set_round(a.guide_round if a.guide_round is not None else a.round)
    cases, P = build_cases(a, rd)
    if a.cases:
        want = {c for c in a.cases.split(",") if c}
        unknown = want - {c["case_id"] for c in cases}
        if unknown:
            sys.exit("error: --cases names units not in this label set: %s" % sorted(unknown)[:5])
        cases = [c for c in cases if c["case_id"] in want]
    stub = os.environ.get("X2D_LABEL_STUB")
    ra, rb, rc = make_stub(json.load(open(stub))) if stub else (run_claude, run_codex, run_claude)
    os.makedirs(P["dir"], exist_ok=True)
    try:
        os.close(os.open(P["lock"], os.O_CREAT | os.O_EXCL | os.O_WRONLY))
    except FileExistsError:
        sys.exit("error: %s exists: another labelling run is active, or a killed run left it (confirm none is running, delete it)" % P["lock"])
    try:
        label_all(a, cases, P, rd, ra, rb, rc)
    finally:
        os.unlink(P["lock"])


def label_all(a, cases, P, rd, ra, rb, rc):
    gv = x2d.guide_version(a.guide_round if a.guide_round is not None else a.round)
    existing = [r for r in x2d.read_jsonl(P["labels"]) if r.get("guide_version") == gv and r.get("kind") == a.kind]
    if a.new_round is None and existing:
        sys.exit("error: %s already has %s rows for %s; refusing (append-only). Use --new-round <name>." % (P["labels"], a.kind, gv))
    if a.new_round is not None:
        have = {(r["case_id"], r.get("question")) for r in existing if r.get("agreement") in ("agreed", "adjudicated")}
        cases = [c for c in cases if (c["case_id"], c["question"]) not in have]
        print("new round %r: %d case(s) without a usable label" % (a.new_round, len(cases)))
    plan = plan_for(a.round, rd, a.perception_n)
    cap = cap_for(a.round, **plan)
    used = used_sessions(a.round, rd)
    if used + 2 * len(cases) > cap:
        sys.exit("error: session cap: %d used + %d planned (2 per case, adjudications and retries extra) > cap %d (plan %s)" % (used, 2 * len(cases), cap, plan))
    rnd = a.new_round
    lock = threading.Lock()
    counter = {"n": 0}
    stop = threading.Event()

    def reserve(k=1):
        with lock:
            if used + counter["n"] + k > cap:
                stop.set()
                return False
            counter["n"] += k
            return True

    def add_row(rows, case, by, prov, s, ts, extra):
        base = {"case_id": case["case_id"], "kind": case["kind"], "label_set": case["label_set"], "by": by, "provenance": prov,
                "guide_version": gv, "timestamp": ts, "round": rnd}
        if case["question"]:
            base["question"] = case["question"]
        ans = s if s is not None else ({"intruder": None, "rule": "", "reason": ""} if case["kind"] == "intruder" else _answer())
        row = {**base, **ans, **extra}
        if case["kind"] == "intruder":
            row["answer"] = {"intruder": ans["intruder"]}
        rows.append(row)

    def work(case):
        if stop.is_set():
            return None
        x = call(ra, None, prompt_ab(case), case, reserve)
        y = call(rb, None, prompt_ab(case), case, reserve) if not x.get("capped") else None
        capped = bool(x.get("capped") or (y or {}).get("capped"))
        z, status, final = None, "unparsable", None
        if not capped:
            status, final = case["agree"](x["ans"], y["ans"])
            if status != "agreed":
                z = call(rc, "sonnet", prompt_c(case, x["ans"], y["ans"]), case, reserve)
                capped = bool(z.get("capped"))
        rows = []
        if not capped:
            for k, who, s in (("A", "labeller_A:", x), ("B", "labeller_B:", y), ("C", "adjudicator_C:", z)):
                if s is not None:
                    add_row(rows, case, who + s["model"], "agent", s["ans"], s["timestamp"], {"parsed": s["ok"]})
            if status == "agreed":
                add_row(rows, case, "A+B agree", "agent_reviewed", final, now(), {"agreement": "agreed"})
            elif z and z["ok"]:
                add_row(rows, case, "C adjudicated", "agent_reviewed", z["ans"], now(), {"agreement": "adjudicated"})
            else:
                add_row(rows, case, "none", "agent_reviewed", None, now(), {"agreement": "unparsable"})
        call_row = {"case_id": case["case_id"], "kind": case["kind"], "label_set": case["label_set"], "question": case["question"],
                    "A": x, "B": y, "C": z, "round": rnd}
        if capped:
            call_row["capped"] = True  # the cap was reached mid-case: the sessions are counted, no label is written (a later --new-round labels the case)
        attempted = sum(s["attempts"] for s in (x, y, z) if s)
        with lock:  # label rows FIRST, then the call row: a kill in between leaves a labelled case without a call row, never an orphan call row
            if rows:
                x2d.append_jsonl(P["labels"], rows)
            if attempted:
                x2d.append_jsonl(P["calls"], [call_row])
        return None if capped else case["cid"]

    with ThreadPoolExecutor(a.workers) as ex:
        done = [c for c in ex.map(work, cases) if c]
    if stop.is_set() and len(done) < len(cases):
        print("session cap reached: %d of %d cases labelled" % (len(done), len(cases)), file=sys.stderr)
    rec = write_sessions(a.round, rd, cap, plan)
    print("labelled %d cases (%s); sessions this run: %d; round sessions %d of cap %d" % (len(done), a.kind, counter["n"], rec["used"], cap))
    if len(done) < len(cases):
        sys.exit(1)


if __name__ == "__main__":
    main()
