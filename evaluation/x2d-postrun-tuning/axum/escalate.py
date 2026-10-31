"""Development: escalate the window clusters that V-sem leaves unresolved or answers `none` (at t_choice 0.535, covered score threshold s 0.60) to a reasoning model
through `codex exec` (gpt-6.1-terra, reasoning effort high), then compose and score Jev-only, Jev+Terra and B2 against gold. Terra sees the V-sem merge.covered
state (`jev -dump-state`: unit excerpt, PR, diff summary, the whole catalogue) under the prompt template file. Append-only, resumable per (prompt sha, case).

  X2D_DATA=<worktree>/evaluation/x2d-merge-association-pilot/data/postrun <worktree>/.tools/venv/bin/python escalate.py <study dir> --prompt FILE --tag TAG --rounds 1-5 [--score-only] [--workers 6]

Writes escalation/<tag>.calls.jsonl, escalation/<tag>.report.json and escalation/<tag>.errors.txt next to this file. Model calls: `codex exec` only (scrubbed env, read-only
sandbox, ephemeral, empty working directory, no config or rules); Terra answers that ran a command are discarded.
"""
import argparse, concurrent.futures as cf, hashlib, json, os, re, subprocess, sys, tempfile, threading, time

HERE = os.path.dirname(os.path.abspath(__file__))
ap = argparse.ArgumentParser()
ap.add_argument("study")
ap.add_argument("--prompt", required=True)
ap.add_argument("--tag", required=True)
ap.add_argument("--rounds", required=True)  # e.g. 1-5, 6-7, 8-10
ap.add_argument("--score-only", action="store_true")
ap.add_argument("--workers", type=int, default=6)
A_ = ap.parse_args()
S = os.path.abspath(A_.study)
sys.argv = [sys.argv[0], S]  # compose_window reads argv[1]
sys.path.insert(0, HERE)
import compose_window as cw
import label_x2d as L
x2d, compose, A, G = cw.x2d, cw.compose, cw.A, cw.G

MODEL, EFFORT = "gpt-6.1-terra", "high"
T_CHOICE, S_COV = 0.535, 0.60
lo, hi = (int(x) for x in A_.rounds.split("-"))
ROUNDS = list(range(lo, hi + 1))
EDIR = os.path.join(HERE, "escalation")
os.makedirs(EDIR, exist_ok=True)
CALLS = os.path.join(EDIR, A_.tag + ".calls.jsonl")
TEMPLATE = open(A_.prompt).read()
PSHA = hashlib.sha256(TEMPLATE.encode()).hexdigest()
TH = {"decisions": {"merge.covered": {"threshold": None, "score_threshold": S_COV}, "merge.which_area": {"threshold": T_CHOICE}}}
lock = threading.Lock()


def states(r):
    p = cw.rd(r, "states.window.jsonl")
    if not os.path.exists(p):
        wt = os.path.dirname(os.path.dirname(S))
        d = cw.rd(r)
        subprocess.run(["go", "run", "./evaluation/x2d-merge-association-pilot/jev", "-dump-state", "-in", os.path.join(d, "units.window.jsonl"), "-catalogue", os.path.join(DATA, "catalogue", "catalogue.json"),
                        "-arms", "sem", "-out", p, "-questions-src", os.path.join(d, "inputs", "questions"), "-settings", os.path.join(d, "settings.json")], cwd=wt, check=True)
    return {s["case_id"]: s for s in x2d.read_jsonl(p) if s.get("variant") == "V-sem"}


DATA = cw.DATA


def build_prompt(st):
    return TEMPLATE + "\n=== STATE (JSON) ===\n" + json.dumps(st["noul_state"], separators=(",", ":")) + "\n"


def run_codex(p):
    with tempfile.TemporaryDirectory() as d:
        out = os.path.join(d, "last.txt")
        t0 = time.time()
        r = subprocess.run(["codex", "exec", "-m", MODEL, "-c", "model_reasoning_effort=%s" % EFFORT, "-s", "read-only", "--skip-git-repo-check", "--ephemeral",
                            "--ignore-user-config", "--ignore-rules", "-C", d, "-o", out, "-"], input=p, capture_output=True, text=True, cwd=d, timeout=900, env=L.session_env())
        dt = time.time() - t0
        txt = open(out).read() if os.path.exists(out) else ""
        both = r.stdout + r.stderr
        m = re.search(r"^model:\s*(\S+)", both, re.M)
        e = re.search(r"^reasoning effort:\s*(\S+)", both, re.M)
        execs = len(re.findall(r"^exec\b", both, re.M))
    return txt, (m.group(1) if m else None), (e.group(1) if e else None), execs, dt


def parse(txt, ids):
    o = L._json(txt)
    if not isinstance(o, dict) or o.get("decision") not in ("existing", "new"):
        return None
    if o["decision"] == "existing":
        return o if o.get("area_id") in ids else None
    return o


def one(case_id, rnd, st, ids):
    prompt = build_prompt(st)
    for attempt in (1, 2):  # one retry for an unparsable answer or an error
        try:
            txt, model, eff, execs, dt = run_codex(prompt)
        except Exception as ex:  # timeout or CLI failure
            row = {"case_id": case_id, "round": rnd, "prompt_sha256": PSHA, "attempt": attempt, "error": repr(ex)[:200], "parsed": None}
            continue
        parsed = parse(txt, ids) if execs == 0 else None
        row = {"case_id": case_id, "round": rnd, "prompt_sha256": PSHA, "attempt": attempt, "model": model, "effort": eff, "execs": execs, "latency_s": round(dt, 1),
               "raw": txt[:1500], "parsed": parsed}
        if parsed is not None:
            break
    with lock:
        with open(CALLS, "a") as f:
            f.write(json.dumps(row) + "\n")
    return row


def jev_rows(data):
    out = {}
    for r in ROUNDS:
        g, units, scores, b2 = data[r]
        for u, row in zip(units, compose.compose(scores, units, TH, "V-sem")):
            out[u["case_id"]] = (r, u, row)
    return out


def main():
    data = {r: cw.load(r) for r in ROUNDS}
    jev = jev_rows(data)
    esc = [c for c, (r, u, row) in jev.items() if row["output"] in ("none", "unresolved")]
    ids = {a["area_id"] for a in json.load(open(os.path.join(DATA, "catalogue", "catalogue.json")))["areas"]}
    done = {}
    for row in (x2d.read_jsonl(CALLS) if os.path.exists(CALLS) else []):
        if row["prompt_sha256"] == PSHA:
            if row["case_id"] not in done or row.get("parsed") is not None:
                done[row["case_id"]] = row
    todo = [c for c in esc if c not in done]
    if not A_.score_only and todo:
        st = {r: states(r) for r in ROUNDS}
        with cf.ThreadPoolExecutor(A_.workers) as ex:
            futs = [ex.submit(one, c, jev[c][0], st[jev[c][0]][c], ids) for c in todo]
            for f in cf.as_completed(futs):
                done[f.result()["case_id"]] = f.result()
    elif todo:
        print("score-only: %d escalations have no call row" % len(todo))
    report(data, jev, esc, done)


def final_row(c, row, call):
    if row["output"] == "assigned":
        return row
    p = (call or {}).get("parsed")
    base = {"case_id": c, "arm": "V-sem+Terra", "confidence": None}
    if p is None:
        return dict(base, output="unresolved", area_id=None, unresolved_reason="terra_failure", raw_covered=None, raw_choice=None)
    if p["decision"] == "existing":
        return dict(base, output="assigned", area_id=p["area_id"], unresolved_reason=None, raw_covered=True, raw_choice=p["area_id"])
    return dict(base, output="none", area_id=None, unresolved_reason=None, raw_covered=False, raw_choice=None)


def score(data, rows_by_case, rs):
    outs = []
    for r in rs:
        g, units, scores, b2 = data[r]
        outs += [A.score_outcome(rows_by_case(r, u), g[u["case_id"]]) for u in units]
    rt = A.rates(A.counts_for(outs))
    f = cw.figures(A.counts_for(outs))
    f["wilson95"] = {k: rt[k]["wilson95"] for k in ("P_assign", "R_assign_res", "unresolved_rate", "P_new", "R_new")}
    f["gold_covered_answered_none"] = sum(1 for o in outs if o["kind"] == "none_wrong")
    return f


def report(data, jev, esc, done):
    calls = [done[c] for c in esc if c in done]
    lat = sorted(c["latency_s"] for c in calls if c.get("latency_s") is not None)
    fin = {c: final_row(c, jev[c][2], done.get(c)) for c in jev}
    parts = {"all": ROUNDS}
    if len(ROUNDS) > 1:
        parts.update({"r%d" % r: [r] for r in ROUNDS} if len(ROUNDS) <= 3 else {})
    res = {"tag": A_.tag, "prompt_sha256": PSHA, "model": MODEL, "effort": EFFORT, "jev": {"t_choice": T_CHOICE, "s_cov": S_COV}, "rounds": ROUNDS,
           "units": sum(len(data[r][1]) for r in ROUNDS), "escalations": len(esc), "escalations_per_100": round(100 * len(esc) / sum(len(data[r][1]) for r in ROUNDS), 1),
           "calls": len(calls), "attempts": len([1 for x in (x2d.read_jsonl(CALLS) if os.path.exists(CALLS) else []) if x["prompt_sha256"] == PSHA and x["case_id"] in set(esc)]),
           "terra_failures": sum(1 for c in esc if (done.get(c) or {}).get("parsed") is None),
           "reported_models": sorted({c.get("model") for c in calls if c.get("model")}), "reported_efforts": sorted({c.get("effort") for c in calls if c.get("effort")}),
           "latency_s": {"n": len(lat), "median": lat[len(lat) // 2] if lat else None, "p90": lat[int(0.9 * (len(lat) - 1))] if lat else None, "max": lat[-1] if lat else None,
                         "total": round(sum(lat), 1)}, "arms": {}}
    for name, fn in (("Jev-only", lambda r, u: jev[u["case_id"]][2]), ("Jev+Terra", lambda r, u: fin[u["case_id"]]), ("B2", lambda r, u: data[r][3][u["case_id"]])):
        res["arms"][name] = {p: score(data, fn, rs) for p, rs in parts.items()}
    # escalated-only accuracy and error list
    errs = []
    esc_ok = [0, 0]
    for c in esc:
        r, u, row = jev[c]
        g = data[r][0][c]
        p = (done.get(c) or {}).get("parsed")
        ok = None if p is None else ((p["decision"] == "existing" and g["covered"] and p["area_id"] in set(g.get("acceptable") or [])) or (p["decision"] == "new" and not g["covered"]))
        esc_ok[0] += 1 if ok else 0
        esc_ok[1] += 1
        if not ok:
            errs.append("r%d %s jev=%s gold=%s%s terra=%s reason=%s" % (r, c, row["output"], "none" if not g["covered"] else sorted(g["acceptable"]), "", json.dumps({k: p[k] for k in ("decision", "area_id")}) if p else None,
                                                                           (p or {}).get("reason")))
    res["escalated_correct"] = esc_ok
    json.dump(res, open(os.path.join(EDIR, A_.tag + ".report.json"), "w"), indent=1)
    open(os.path.join(EDIR, A_.tag + ".errors.txt"), "w").write("\n".join(errs) + "\n")
    print(json.dumps({k: v for k, v in res.items() if k != "arms"}))
    for arm, d in res["arms"].items():
        for p, f in d.items():
            print(arm, p, {k: (round(v, 3) if isinstance(v, float) else v) for k, v in f.items() if k not in ("wilson95",)})


if __name__ == "__main__":
    main()
