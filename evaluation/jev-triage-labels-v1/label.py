"""Agent first-pass labelling. Run from this folder: python3 label.py
A = claude -p (default model), B = codex exec, C = claude -p --model sonnet (adjudicator).
Appends to calls.jsonl and labels.jsonl (never rewrites). Refuses if labels.jsonl has rows for this guide_version unless --new-round <name>. Stdlib only."""
import json, os, re, subprocess, tempfile, hashlib, sys
from concurrent.futures import ThreadPoolExecutor
from datetime import datetime, timezone

HERE = os.path.dirname(os.path.abspath(__file__))
GUIDE = open(os.path.join(HERE, "GUIDE.md")).read()
GUIDE_VERSION = "triage-v1_choice/guide-1"
LABELS = ("real_defect", "false_positive", "cannot_tell")
WORKERS = 4
now = lambda: datetime.now(timezone.utc).isoformat(timespec="seconds")

BASE = ("You are labelling one static-analyzer finding. Apply the label guide below to the case state. "
        "Use only the case state; you have no tools and no other context.\n\n=== LABEL GUIDE ===\n" + GUIDE +
        "\n=== CASE STATE (JSON) ===\n")
FMT = ('Answer with ONLY one JSON object: {"label": "real_defect"|"false_positive"|"cannot_tell", '
       '"rule": "<guide rule id/text that decided it>", "reason": "<=2 sentences"}')

def prompt_ab(state):
    return BASE + json.dumps(state, indent=1) + "\n\n" + FMT

def prompt_c(state, a, b):
    ev = lambda x: json.dumps({k: x.get(k) for k in ("label", "rule", "reason")}) if x else "unparsable (no valid answer)"
    return (BASE + json.dumps(state, indent=1) +
            "\n\n=== TWO INDEPENDENT LABELLERS DISAGREED (or failed) ===\nLabeller A: " + ev(a) +
            "\nLabeller B: " + ev(b) +
            "\n\nYou are the adjudicator. Decide the label by the guide yourself. " + FMT[:-1] +
            ', "which": "A"|"B"|"neither" (which labeller you agree with)}')

def run_claude(p, model):
    cmd = ["claude", "-p", "--tools", "", "--no-session-persistence", "--setting-sources", "",
           "--output-format", "json"] + (["--model", model] if model else [])
    with tempfile.TemporaryDirectory() as d:
        r = subprocess.run(cmd, input=p, capture_output=True, text=True, cwd=d, timeout=600)
    j = json.loads(r.stdout)
    return j["result"], ",".join(j.get("modelUsage", {}).keys())

def run_codex(p, model):
    with tempfile.TemporaryDirectory() as d:
        out = os.path.join(d, "last.txt")
        r = subprocess.run(["codex", "exec", "-s", "read-only", "--skip-git-repo-check", "--ephemeral",
                            "--ignore-user-config", "--ignore-rules", "-C", d, "-o", out, "-"],
                           input=p, capture_output=True, text=True, cwd=d, timeout=600)
        txt = open(out).read()
        m = re.search(r"^model:\s*(\S+)", r.stdout + r.stderr, re.M)
    return txt, m.group(1) if m else "unknown"

def parse(txt, want_which=False):
    t = txt.strip()
    t = re.sub(r"^```(?:json)?\s*|\s*```$", "", t)
    try:
        o = json.loads(t)
    except Exception:
        return None
    if not isinstance(o, dict) or o.get("label") not in LABELS: return None
    if not isinstance(o.get("rule"), str) or not isinstance(o.get("reason"), str): return None
    if want_which and o.get("which") not in ("A", "B", "neither"): return None
    return o

def call(runner, model, p, want_which=False):
    for attempt in (1, 2):
        try:
            txt, resolved = runner(p, model)
        except Exception as e:
            txt, resolved = "ERROR %r" % e, "unknown"
        o = parse(txt, want_which)
        if o:
            return {"ok": True, **o, "model": resolved, "attempts": attempt, "timestamp": now()}
    return {"ok": False, "label": "unparsable", "model": resolved, "attempts": 2, "raw": txt[:500], "timestamp": now()}

def work(row):
    cid, st = row["case_id"], row["state"]
    pa, pb = prompt_ab(st), prompt_ab(st)
    a = call(run_claude, None, pa)
    b = call(run_codex, None, pb)
    c = None
    if not (a["ok"] and b["ok"] and a["label"] == b["label"]):
        c = call(run_claude, "sonnet", prompt_c(st, a if a["ok"] else None, b if b["ok"] else None), True)
    return {"case_id": cid, "A": a, "B": b, "C": c}

if __name__ == "__main__":
    rnd = None
    if "--new-round" in sys.argv:
        i = sys.argv.index("--new-round")
        if i + 1 >= len(sys.argv): sys.exit("error: --new-round needs a name")
        rnd = sys.argv[i + 1]
    LP = os.path.join(HERE, "labels.jsonl")
    if os.path.exists(LP):
        for l in open(LP):
            if l.strip() and json.loads(l).get("guide_version") == GUIDE_VERSION and rnd is None:
                sys.exit("error: labels.jsonl already has rows for %s; refusing (append-only). "
                         "Use --new-round <name> to append a new round." % GUIDE_VERSION)
    rows = [json.loads(l) for l in open(os.path.join(HERE, "blinded.jsonl"))]
    with ThreadPoolExecutor(WORKERS) as ex:
        res = list(ex.map(work, rows))
    with open(os.path.join(HERE, "calls.jsonl"), "a") as f:
        for r in res: f.write(json.dumps({**r, **({"round": rnd} if rnd else {})}) + "\n")
    with open(LP, "a") as f:
        def row(cid, label, by, prov, rule, note, ts):
            f.write(json.dumps({"case_id": cid, "label": label, "by": by, "provenance": prov, "rule": rule,
                                "note": note, "guide_version": GUIDE_VERSION, "timestamp": ts,
                                **({"round": rnd} if rnd else {})}) + "\n")
        for r in res:
            cid = r["case_id"]
            for k, who in (("A", "labeller_A:" ), ("B", "labeller_B:"), ("C", "adjudicator_C:")):
                x = r[k]
                if x is None: continue
                note = x.get("reason", x.get("raw", "")) + ("" if k != "C" else " [which=%s]" % x.get("which"))
                row(cid, x["label"], who + x["model"], "agent", x.get("rule", ""), note, x["timestamp"])
            a, b, c = r["A"], r["B"], r["C"]
            if a["ok"] and b["ok"] and a["label"] == b["label"]:
                lab, by, rule, note = a["label"], "A+B agree", a["rule"], "A and B agreed"
            elif c and c["ok"]:
                lab, by, rule, note = c["label"], "C adjudicated", c["rule"], c["reason"]
            else:
                lab, by, rule, note = "unparsable", "none", "", "no valid proposal"
            row(cid, lab, by, "agent_reviewed", rule, note, now())
