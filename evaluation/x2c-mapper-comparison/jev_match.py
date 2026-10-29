"""The Jev matching method (X2c PREREGISTRATION section 6): Jev scores -> replay-schema rows that X2's frozen analysis consumes unchanged.

A candidate pair's Jev score is s = mean over the two presentation orders of P(same_concern) + P(overlapping_concern) (the probability of an answer X2's C2 does not count
as spurious; PREREGISTRATION section 6 item 5); a pair with no answer row for either order has no score and is not matched (a failed call is never a pass; a resumed run's later answer row for that order is its score; the unscored
share is gated). For a mapper arm A and each grid threshold tau (percent, the X2 grid 30..90) the method matches prior areas to next features GREEDILY BEST-FIRST AND
ONE-TO-ONE over the candidates with s >= tau/100, ordered by (-s, -P(same_concern), prior feature id, next feature id) - the same greedy best-first one-to-one rule as
`areas.Reconcile`, with s in place of the Jaccard and ties broken by feature ids rather than Reconcile's (area id, feature id). The rows are the Jaccard replay rows of arm A
with `arm = "<A>+jev"`, `matched`/`matches` replaced (`jaccard` carries s, so X2's tables and diagnostics read it as the match score), `created = nextAgentFeatures - matched`,
`kept = prevAreas - matched`, `retired = 0` (placeholders: they enter no decision, and are unmeasured).
"""
import json

import x2c
import common as C


def read_scores(path):
    """-> (header, {pair_id: {order: (s, p_same)}}, answers, failures). s = P(same_concern) + P(overlapping_concern) of the stored Choice probabilities. Rows are append-only; the last answer per (pair, order) wins (a resumed run retries failures)."""
    header, answers, failures = None, [], []
    scores = {}
    for r in C.read_jsonl(path):
        if r["type"] == "header":
            header = r
        elif r["type"] == "answer":
            answers.append(r)
            p = r["probabilities"]
            scores.setdefault(r["pair_id"], {})[r["order"]] = (p.get("same_concern", 0.0) + p.get("overlapping_concern", 0.0), p.get("same_concern", 0.0))
        elif r["type"] == "failure":
            failures.append(r)
    return header, scores, answers, failures


def pair_score(orders):
    """Mean over both presentation orders -> (s, p_same); None unless both answered."""
    if "ab" in orders and "ba" in orders:
        return ((orders["ab"][0] + orders["ba"][0]) / 2, (orders["ab"][1] + orders["ba"][1]) / 2)
    return None


def greedy(cands, pct):
    """cands: [(s, p_same, prev id, next id)] -> matches at tau = pct/100: best-first by (-s, -p_same, prev id, next id), one-to-one, s >= tau."""
    used_p, used_n, out = set(), set(), []
    for s, _ps, p, n in sorted(cands, key=lambda c: (-c[0], -c[1], c[2], c[3])):
        if s < pct / 100.0:
            break
        if p in used_p or n in used_n:
            continue
        used_p.add(p)
        used_n.add(n)
        out.append({"areaId": "", "prevFeatureId": p, "nextFeatureId": n, "jaccard": s})
    return out


def jev_rows(replay_rows, key, scores, arms=None):
    """-> (rows, unscored) where unscored = {(arm, repo): [scored-less candidate occurrences, all candidate occurrences]}."""
    by_tr = {}
    for k in key:
        by_tr.setdefault((k["arm"], k["repo"], k["transition"]), []).append(k)
    rows, unscored = [], {}
    for r in replay_rows:
        if arms is not None and r["arm"] not in arms:
            continue
        j = dict(r, arm=r["arm"] + x2c.JEV_SUFFIX)
        if r["status"] == "ok":
            cands = by_tr.get((r["arm"], r["repo"], r["transition"]), [])
            scored = [(pair_score(scores.get(k["pair_id"], {})), k["prev_feature_id"], k["next_feature_id"]) for k in cands]
            if r["thresholdPct"] == C.GRID_PCT[0]:
                u = unscored.setdefault((r["arm"], r["repo"]), [0, 0])
                u[0] += sum(1 for sc, _p, _n in scored if sc is None)
                u[1] += len(scored)
            m = greedy([(sc[0], sc[1], p, n) for sc, p, n in scored if sc is not None], r["thresholdPct"])
            j.update({"matched": len(m), "matches": m, "created": r["nextAgentFeatures"] - len(m), "kept": r["prevAreas"] - len(m), "retired": 0})
        rows.append(j)
    return rows, unscored


def write_jsonl(path, rows):
    with open(path, "w") as f:
        for r in rows:
            f.write(json.dumps(r) + "\n")
