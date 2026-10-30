"""Validation helpers for label_x2d (no model call).

Truth format (env X2D_LABEL_STUB=<truth.json>): {cid: {"A": x, "B": x, "C": x}}; cid = case_id (wide, intruder) or "<case_id>|<question>" (jev-state);
x = answer dict {covered, acceptable, primary, none_kind, rule?, reason?} (intruder: {intruder, reason?}) | None (the session answers non-JSON = unparsable).
An answer dict may hold "omit": [fields] and "extra": {k: v}; for B (the stub codex session) also "exec": N = the session reports N `exec` events (a stub session with an
`exec` line: label_x2d quarantines it as unparsable, never labels from it). A cid missing from the truth makes the session fail (recorded as unparsable).

make_truth(units, gold, agree_rate, seed): units = rows with case_id; gold = {case_id: {covered, acceptable, primary, none_kind}} (acceptable ids are
catalogue ids). A and B both give the gold answer for a hash-chosen `agree_rate` of the units; for the others B deviates (different covered or a disjoint
acceptable set) and C gives the gold answer. Exactly one case (the last in hash order among the disagreeing ones; or the first unit when all agree) is
unparsable for A, B and C, so the final row is `unparsable`.

Blindness probe: the stub appends {"cid","who","prompt"} to the file in env X2D_LABEL_PROMPT_LOG for every session; probe_leaks(log, needles) returns the
prompts containing any needle string (arm answers, sibling owners, stratum labels, control kinds).
"""
import hashlib
import json


def _h(seed, cid):
    return hashlib.sha256(("%s\0%s" % (seed, cid)).encode()).hexdigest()


def _answer(g):
    return {"covered": g["covered"], "acceptable": list(g.get("acceptable") or []), "primary": g.get("primary"), "none_kind": g.get("none_kind"),
            "rule": "stub", "reason": "stub"}


def _deviate(g, area_ids):
    a = _answer(g)
    if g["covered"]:
        other = [i for i in area_ids if i not in a["acceptable"]]
        if other:
            a["acceptable"], a["primary"] = [other[0]], other[0]
        else:
            a.update(covered=False, acceptable=[], primary=None, none_kind="gap")
    else:
        a["none_kind"] = "new" if g.get("none_kind") != "new" else "gap"
    return a


def make_truth(units, gold, agree_rate=0.8, seed=0, area_ids=()):
    ids = sorted(u["case_id"] for u in units)
    order = sorted(ids, key=lambda c: _h(seed, c))
    n_agree = int(round(agree_rate * len(ids)))
    agreeing, disagreeing = set(order[:n_agree]), order[n_agree:]
    bad = disagreeing[-1] if disagreeing else (order[0] if order else None)
    truth = {}
    for cid in ids:
        g = _answer(gold[cid])
        if cid == bad:
            truth[cid] = {"A": None, "B": None, "C": None}
        elif cid in agreeing:
            truth[cid] = {"A": g, "B": dict(g), "C": dict(g)}
        else:
            truth[cid] = {"A": g, "B": _deviate(gold[cid], list(area_ids)), "C": dict(g)}
    return truth


def plant_exec(truth, cid, n=1):
    """B's session for `cid` reports n exec events (the codex session used a tool): the answer is quarantined, the case goes to the adjudicator C (R1-15)."""
    truth[cid]["B"] = dict(truth[cid]["B"] or {}, exec=n)
    return truth


def plant_primary_disagreement(truth, cid, acceptable, primary_a, primary_b, covered=True):
    """A and B agree on the covered answer and on `acceptable` (>= 2 ids) but name different primaries (R1-11: the agreed primary follows the `agree` primary rule, never None)."""
    mk = lambda p: {"covered": covered, "acceptable": list(acceptable), "primary": p, "none_kind": None, "rule": "stub", "reason": "stub"}
    truth[cid] = {"A": mk(primary_a), "B": mk(primary_b), "C": mk(primary_a)}
    return truth


def write_states(path, units, catalogue, n=5, noul_instructions="STUB-NOUL-INSTRUCTIONS: is the unit covered by an area?", noul_criteria=None, choice_instructions="STUB-CHOICE-INSTRUCTIONS: which area covers the unit?"):
    """A states.jsonl without the Go driver (the `jev -dump-state` row shape, V-sem only; the real validation runs the driver): per unit the merge.covered state
    {unit, pr, diff_summary, catalogue[id, title, summary, sample_files[:n]]}, the merge.which_area state (no catalogue), the option texts of the Choice and the effective
    instructions / criteria of both questions (R2-17: `noul_instructions`, `noul_criteria`, `choice_instructions`)."""
    cat = [{"id": a["area_id"], "title": a.get("title"), "summary": a.get("summary"), "sample_files": list(a.get("sample_files", []))[:n]} for a in catalogue["areas"]]
    crit = {a["id"]: "%s: %s Sample files: %s" % (a["title"], a["summary"], ", ".join(a["sample_files"])) for a in cat}
    with open(path, "w") as f:
        for u in units:
            base = {k: u["state"][k] for k in ("unit", "pr", "diff_summary")}
            f.write(json.dumps({"case_id": u["case_id"], "variant": "V-sem", "noul_state": dict(base, catalogue=cat), "choice_state": base,
                                "choice_options": list(crit.values()), "choice_criteria": crit, "noul_instructions": noul_instructions,
                                "noul_criteria": noul_criteria or {"true": "STUB-TRUE", "false": "STUB-FALSE"}, "choice_instructions": choice_instructions}) + "\n")


def probe_leaks(log_path, needles):
    out = []
    for line in open(log_path):
        r = json.loads(line)
        hit = [n for n in needles if n and n in r["prompt"]]
        if hit:
            out.append({"cid": r["cid"], "who": r["who"], "needles": hit})
    return out
