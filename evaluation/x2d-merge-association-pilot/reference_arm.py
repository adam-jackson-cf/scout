#!/usr/bin/env python3
"""Reference arm (README Modules): a `claude -p` session (no tools, ephemeral, scrubbed env, empty cwd; label_x2d's runner) answers BOTH of Jev's questions
(merge.covered, merge.which_area) from the exact state `jev -dump-state` wrote to states.jsonl (--states, default D/rounds/R/states.jsonl; the V-sem row: the merge.covered
state with the catalogue, the merge.which_area state and its option texts), for the REFERENCE_SUBSET (20) clusters of the round's scored window units with the smallest
x2d.reference_key(cluster_id). Writes arms/reference.jsonl (arm `reference-claude`; output assigned|none, or unresolved `unparsable` when no valid answer or
`session_cap` when the cap left no slot (the arm never fails the round for the cap, R3-01); confidence null) and calls.reference.jsonl (session log, counted against the round's session cap: a slot is reserved before every
attempt; cap = label_x2d.cap_for(r, W, J, R, I, C)). Updates sessions.json. Append-only; not used in round 0.

Usage: reference_arm.py --round R [--states FILE]
       X2D_REFERENCE_STUB=<json {case_id: {"covered": bool, "area_id": id|null}}>  validation only: no model call (a case missing from it is unparsable)
       X2D_SKIP_SIGNOFF=1  verify_pins(require_signoff=False) (validation only).
"""
import argparse
import json
import os
import sys

import x2d
import label_x2d as L

ARM = "reference-claude"


def pick(units, n=x2d.REFERENCE_SUBSET):
    window = [u for u in units if not u.get("control") and u.get("set", "window") == "window"]
    return sorted(window, key=lambda u: x2d.reference_key(u["cluster_id"]))[:n]


def prompt(srow):
    q1, c1 = srow["noul_instructions"], srow.get("noul_criteria")  # the effective V-sem instructions/criteria of the dumped state (R2-17)
    q2, c2 = srow["choice_instructions"], None
    return ("You are answering two questions about one change from exactly the state shown. Use only that state; you have no tools and no other context.\n\n"
            "=== QUESTION 1 (covered) ===\n" + q1 + "\n" + ("Criteria: " + json.dumps(c1) + "\n" if c1 else "") +
            "\n=== STATE 1 (JSON) ===\n" + json.dumps(srow["noul_state"], indent=1) +
            "\n\n=== QUESTION 2 (which area; only if the answer to 1 is true) ===\n" + q2 + ("\nCriteria: " + json.dumps(c2) if c2 else "") +
            "\nEach option is an area id; its text is its criterion.\n\n=== STATE 2 and OPTIONS (JSON) ===\n" +
            json.dumps({"state": srow["choice_state"], "options": srow.get("choice_criteria") or srow["choice_options"]}, indent=1) +
            '\n\nAnswer with ONLY one JSON object: {"covered": true|false, "area_id": "<one area id from the options>"|null (null when covered is false)}')


def parse(txt, ids):
    o = L._json(txt)
    if o is None or not isinstance(o.get("covered"), bool):
        return None
    if not o["covered"]:
        return {"covered": False, "area_id": None}
    return {"covered": True, "area_id": o.get("area_id")} if o.get("area_id") in ids else None


def main(argv=None):
    ap = argparse.ArgumentParser()
    ap.add_argument("--round", type=int, required=True)
    ap.add_argument("--states")
    a = ap.parse_args(argv)
    if a.round == 0:
        sys.exit("error: the reference arm is not used in round 0")
    bad = x2d.verify_pins(require_signoff=x2d.require_signoff())
    if bad:
        sys.exit("pin check failed:\n" + "\n".join(bad))
    rd = x2d.round_dir(a.round)
    out_p, call_p = os.path.join(rd, "arms", "reference.jsonl"), os.path.join(rd, "calls.reference.jsonl")
    x2d.refuse_if_exists(out_p, call_p)
    units = pick(x2d.read_jsonl(os.path.join(rd, "units.jsonl")))
    if not units:
        sys.exit("error: no scored window units in %s" % os.path.join(rd, "units.jsonl"))
    plan = L.plan_for(a.round, rd)
    cap, used = L.cap_for(a.round, stub_flag="X2D_REFERENCE_STUB", **plan), L.used_sessions(a.round, rd)
    if used + len(units) > cap:  # R3-01: the arm is report-only; clusters the cap leaves no slot for get `unresolved: session_cap` rows below instead of failing the round
        print("warning: session cap: %d used + %d planned > cap %d (plan %s): clusters without a slot are written as unresolved session_cap" % (used, len(units), cap, plan), file=sys.stderr)
    cat = x2d.read_catalogue()
    ids = {x["area_id"] for x in cat["areas"]}
    stub = os.environ.get("X2D_REFERENCE_STUB")
    truth = json.load(open(stub)) if stub else None
    states_p = a.states or os.path.join(rd, "states.jsonl")
    states = L.load_states(states_p) if truth is None else (L.load_states(states_p) if os.path.exists(states_p) else {})
    spent = {"n": 0}
    arm_rows, call_rows = [], []
    for u in units:
        cid, txt, model, attempts, ans, capped = u["case_id"], "", "unknown", 0, None, False
        for attempt in (1, 2):
            if used + spent["n"] + 1 > cap:  # a slot is reserved before every attempt
                capped = True
                break
            spent["n"] += 1
            attempts = attempt
            try:
                if truth is not None:
                    t = truth[cid]
                    txt, model = json.dumps({"covered": t["covered"], "area_id": t.get("area_id")}), "stub-reference"
                else:
                    if cid not in states:
                        sys.exit("error: %s has no V-sem state for %s" % (states_p, cid))
                    txt, model = L.run_claude(prompt(states[cid]), None, cid)
            except Exception as e:  # noqa: BLE001 - recorded, never fabricated
                txt, model = "ERROR %r" % e, "unknown"
            ans = parse(txt, ids)
            if ans:
                break
        if ans is None:
            row = {"output": "unresolved", "area_id": None, "unresolved_reason": "session_cap" if capped else "unparsable"}
        elif ans["covered"]:
            row = {"output": "assigned", "area_id": ans["area_id"], "unresolved_reason": None}
        else:
            row = {"output": "none", "area_id": None, "unresolved_reason": None}
        arm_rows.append({"case_id": cid, "arm": ARM, "confidence": None, **row})
        if attempts:
            call_rows.append({"case_id": cid, "kind": "reference", "A": {"ok": ans is not None, "ans": ans, "model": model, "attempts": attempts,
                                                                        "raw": None if ans else txt[:500], "timestamp": L.now()}, "B": None, "C": None})
    x2d.write_jsonl_new(call_p, call_rows)
    x2d.write_jsonl_new(out_p, arm_rows)
    L.write_sessions(a.round, rd, cap, plan)
    print("reference arm: %d clusters, sessions: %d" % (len(units), sum(r["A"]["attempts"] for r in call_rows)))


if __name__ == "__main__":
    main()
