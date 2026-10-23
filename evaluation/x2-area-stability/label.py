#!/usr/bin/env python3
"""Agent labelling of match pairs (AGENTS.md 'Jev labelling' protocol; PREREGISTRATION section 6).

A = `claude -p` (default model), B = `codex exec`, C = `claude -p --model sonnet` adjudicator, only when A and B disagree or fail.
Each session is ephemeral and sees only the pair's evidence and the frozen GUIDE.md (claude: no tools; codex: read-only sandbox, empty working
directory, no network access, user config and rules ignored; the same session setup as evaluation/jev-triage-labels-v1). Real pairs and control
pairs are labelled indistinguishably. Appends to calls.jsonl and labels.jsonl (never rewrites); refuses when labels.jsonl already holds rows
for this guide version unless --new-round <name>. A new round labels only the pairs of blinded.jsonl that have NO usable final label yet (final
label `unparsable`, or never labelled after an interrupted run); a pair's label rows are written before its call row, under a lock, after the pair is
complete, and a lock file (labels.lock) refuses a concurrent run.
Final label provenance is `agent_reviewed`. Stdlib only.

Usage: label.py --pairs DIR [--workers 4] [--new-round NAME]
       X2_LABEL_STUB=<truth.json> label.py ...   validation only: deterministic stub sessions, no model calls
"""
import argparse
import json
import os
import re
import subprocess
import sys
import tempfile
import threading
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime, timezone

import common as C

GUIDE = open(os.path.join(C.HERE, "GUIDE.md")).read()
now = lambda: datetime.now(timezone.utc).isoformat(timespec="seconds")

BASE = ("You are labelling one pair of automated-mapper features. Apply the label guide below to the pair evidence. "
        "Use only the pair evidence; you have no tools and no other context.\n\n=== LABEL GUIDE ===\n" + GUIDE +
        "\n=== PAIR EVIDENCE (JSON) ===\n")
FMT = ('Answer with ONLY one JSON object: {"label": "same_concern"|"overlapping_concern"|"different_concern"|"cannot_tell", '
       '"rule": "<guide rule id(s) that decided it>", "reason": "<=2 sentences"}')
FIELDS = ("label", "rule", "reason", "which")  # the only fields copied out of a session answer


def prompt_ab(state):
    return BASE + json.dumps(state, indent=1) + "\n\n" + FMT


def prompt_c(state, a, b):
    ev = lambda x: json.dumps({k: x.get(k) for k in ("label", "rule", "reason")}) if x else "unparsable (no valid answer)"
    return (BASE + json.dumps(state, indent=1) + "\n\n=== TWO INDEPENDENT LABELLERS DISAGREED (or failed) ===\nLabeller A: " + ev(a) +
            "\nLabeller B: " + ev(b) + "\n\nYou are the adjudicator. Decide the label by the guide yourself. " + FMT[:-1] +
            ', "which": "A"|"B"|"neither" (which labeller you agree with)}')


def session_env():
    """Label sessions run with Scout's credential scrub (common.scrub_env; AGENTS.md model-use rule, R-036): an API key or base URL in the operator's
    shell can neither switch the sessions to API billing nor re-route them. The CLIs use their own subscription login."""
    return C.scrub_env(os.environ)


def run_claude(p, model, cid=None):
    cmd = ["claude", "-p", "--tools", "", "--no-session-persistence", "--setting-sources", "", "--output-format", "json"]
    cmd += ["--model", model] if model else []
    with tempfile.TemporaryDirectory() as d:
        r = subprocess.run(cmd, input=p, capture_output=True, text=True, cwd=d, timeout=600, env=session_env())
    j = json.loads(r.stdout)
    return j["result"], ",".join(j.get("modelUsage", {}).keys())


def run_codex(p, model, cid=None):
    """-> (answer, model, exec events). Exec events = lines `exec` in codex's transcript (the sandbox is read-only but can read the disk): a session
    that ran any command is counted in the report. [INFERENCE: transcript format of `codex exec`.]"""
    with tempfile.TemporaryDirectory() as d:
        out = os.path.join(d, "last.txt")
        r = subprocess.run(["codex", "exec", "-s", "read-only", "--skip-git-repo-check", "--ephemeral", "--ignore-user-config",
                            "--ignore-rules", "-C", d, "-o", out, "-"], input=p, capture_output=True, text=True, cwd=d, timeout=600, env=session_env())
        txt = open(out).read()
        m = re.search(r"^model:\s*(\S+)", r.stdout + r.stderr, re.M)
        execs = len(re.findall(r"^exec\b", r.stdout + r.stderr, re.M))
    return txt, m.group(1) if m else "unknown", execs


def make_stub(truth):
    """truth: {pair_id: {"A": x, "B": x, "C": x}} with x = label | None (unparsable session) | dict {"label", "omit": [fields], "extra": {k: v}}
    (a session answer with fields missing and/or extra reserved keys such as "ok" or "model")."""
    def runner_for(which):
        def run(p, model, cid=None):
            x = truth[cid][which]
            if x is None:
                return "not json", "stub-" + which
            if isinstance(x, dict):
                o = {"label": x["label"], "rule": "stub", "reason": "stub", "which": "A"}
                for k in x.get("omit", []):
                    o.pop(k, None)
                o.update(x.get("extra", {}))
                return json.dumps(o), "stub-" + which
            return json.dumps({"label": x, "rule": "stub", "reason": "stub", "which": "A"}), "stub-" + which
        return run
    return runner_for("A"), runner_for("B"), runner_for("C")


def parse(txt, want_which=False):
    m = re.search(r"\{.*\}", txt.strip(), re.S)
    if not m:
        return None
    try:
        o = json.loads(m.group(0))
    except Exception:
        return None
    if not isinstance(o, dict) or o.get("label") not in C.LABELS:
        return None
    if want_which and o.get("which") not in ("A", "B", "neither"):
        return None
    return {"label": o["label"], "rule": str(o.get("rule", ""))[:200], "reason": str(o.get("reason", ""))[:1000],
            "which": o.get("which") if o.get("which") in ("A", "B", "neither") else None}


def call(runner, model, p, cid, want_which=False):
    resolved, txt, execs = "unknown", "", None
    for attempt in (1, 2):
        try:
            res = runner(p, model, cid)
            txt, resolved = res[0], res[1]
            if len(res) > 2:
                execs = (execs or 0) + res[2]
        except Exception as e:  # noqa: BLE001 - a failed session is recorded, never fabricated
            txt, resolved = "ERROR %r" % e, "unknown"
        o = parse(txt, want_which)
        if o:
            return {"ok": True, **o, "model": str(resolved), "attempts": attempt, "exec_events": execs, "timestamp": now()}
    return {"ok": False, "label": "unparsable", "rule": "", "reason": "", "which": None, "model": str(resolved), "attempts": 2, "exec_events": execs,
            "raw": txt[:500], "timestamp": now()}


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--pairs", required=True)
    ap.add_argument("--workers", type=int, default=4)
    ap.add_argument("--new-round")
    a = ap.parse_args()
    stub = os.environ.get("X2_LABEL_STUB")
    ra, rb, rc = (make_stub(json.load(open(stub)))) if stub else (run_claude, run_codex, run_claude)
    lp, cp = os.path.join(a.pairs, "labels.jsonl"), os.path.join(a.pairs, "calls.jsonl")
    lock_path = os.path.join(a.pairs, "labels.lock")
    try:
        os.close(os.open(lock_path, os.O_CREAT | os.O_EXCL | os.O_WRONLY))
    except FileExistsError:
        sys.exit("error: %s exists: another labelling run is active, or a killed run left it (confirm none is running, delete it, follow PREREGISTRATION section 6)" % lock_path)
    try:
        label_all(a, ra, rb, rc, lp, cp)
    finally:
        os.unlink(lock_path)


def label_all(a, ra, rb, rc, lp, cp):
    if a.new_round is None and any(r.get("guide_version") == C.GUIDE_VERSION for r in C.read_jsonl(lp)):
        sys.exit("error: labels.jsonl already has rows for %s; refusing (append-only). Use --new-round <name>." % C.GUIDE_VERSION)
    rows = C.read_jsonl(os.path.join(a.pairs, "blinded.jsonl"))
    if a.new_round is not None:
        have = C.final_labels(a.pairs)
        rows = [r for r in rows if r["pair_id"] not in have]
        print("new round %r: %d pair(s) without a usable label" % (a.new_round, len(rows)))
    rnd = {"round": a.new_round} if a.new_round else {}
    lock = threading.Lock()

    def work(row):
        cid, st = row["pair_id"], row["state"]
        x = call(ra, None, prompt_ab(st), cid)
        y = call(rb, None, prompt_ab(st), cid)
        z = None
        if not (x["ok"] and y["ok"] and x["label"] == y["label"]):
            z = call(rc, "sonnet", prompt_c(st, x if x["ok"] else None, y if y["ok"] else None), cid, True)
        res = {"pair_id": cid, "A": x, "B": y, "C": z}
        lab_rows = []

        def add(label, by, prov, rule, note, ts):
            lab_rows.append(json.dumps({"pair_id": cid, "label": label, "by": by, "provenance": prov, "rule": rule, "note": note,
                                        "guide_version": C.GUIDE_VERSION, "timestamp": ts, **rnd}) + "\n")
        for k, who in (("A", "labeller_A:"), ("B", "labeller_B:"), ("C", "adjudicator_C:")):
            s = res[k]
            if s is None:
                continue
            note = (s.get("reason") or s.get("raw", "")) + ("" if k != "C" else " [which=%s]" % s.get("which"))
            add(s["label"], who + s["model"], "agent", s.get("rule", ""), note, s["timestamp"])
        if x["ok"] and y["ok"] and x["label"] == y["label"]:
            add(x["label"], "A+B agree", "agent_reviewed", x.get("rule", ""), "A and B agreed", now())
        elif z and z["ok"]:
            add(z["label"], "C adjudicated", "agent_reviewed", z.get("rule", ""), z.get("reason", ""), now())
        else:
            add("unparsable", "none", "agent_reviewed", "", "no valid proposal", now())
        with lock:  # label rows FIRST, then the call row: a kill in between leaves a labelled pair without a call row (reported), never an orphan call row
            with open(lp, "a") as f:
                f.write("".join(lab_rows))
            with open(cp, "a") as f:
                f.write(json.dumps({**res, **rnd}) + "\n")
        return cid

    with ThreadPoolExecutor(a.workers) as ex:
        done = list(ex.map(work, rows))
    print("labelled %d pairs" % len(done))


if __name__ == "__main__":
    main()
