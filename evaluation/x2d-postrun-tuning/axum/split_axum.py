"""Axum post-hoc description (model-free): what the assigned-wrong answers of each arm are (gold none vs wrong area), gold none kinds and agreement, catalogue counts.

  X2D_DATA=<axum data root> <venv python> split_axum.py <fork pilot dir> FIRST-LAST
"""
import collections, json, os, sys

HERE = os.path.dirname(os.path.abspath(__file__))
S, RNG = os.path.abspath(sys.argv[1]), sys.argv[2]
sys.argv = [sys.argv[0], S, "--prompt", os.path.join(HERE, "..", "terra_prompt_v1.txt"), "--tag", "axum-r" + RNG, "--rounds", RNG]
sys.path.insert(0, HERE)
import escalate as E  # noqa: E402

data = {r: E.cw.load(r) for r in E.ROUNDS}
jev = E.jev_rows(data)
done = {}
for row in E.x2d.read_jsonl(E.CALLS):
    if row["prompt_sha256"] == E.PSHA and (row["case_id"] not in done or row.get("parsed") is not None):
        done[row["case_id"]] = row
fin = {c: E.final_row(c, jev[c][2], done.get(c)) for c in jev}
out = {}
for name, get in (("Jev-only", lambda r, u: jev[u["case_id"]][2]), ("Jev+Terra", lambda r, u: fin[u["case_id"]]), ("B2", lambda r, u: data[r][3][u["case_id"]])):
    c = collections.Counter()
    for r in E.ROUNDS:
        g, units, s, b = data[r]
        for u in units:
            row, gg = get(r, u), g[u["case_id"]]
            if row["output"] == "assigned":
                ok = gg["covered"] and row["area_id"] in set(gg.get("acceptable") or [])
                c["assigned_correct" if ok else ("assigned_wrong_on_gold_none" if not gg["covered"] else "assigned_wrong_area")] += 1
    out[name] = dict(c)
kinds, agree = collections.Counter(), collections.Counter()
for r in E.ROUNDS:
    g = data[r][0]
    for u in data[r][1]:
        gg = g[u["case_id"]]
        agree[gg.get("agreement")] += 1
        if not gg["covered"]:
            kinds[gg.get("none_kind")] += 1
cat = json.load(open(os.path.join(E.DATA, "catalogue", "catalogue.json")))
out["gold_none_kinds"], out["gold_agreement_window_units"] = dict(kinds), dict(agree)
out["catalogue"] = {"areas": len(cat["areas"]), "contested": len(cat["contested"]), "unowned": len(cat["unowned"]), "jev_assigned": len(cat["jev_assigned"])}
json.dump(out, open(os.path.join(HERE, "split.json"), "w"), indent=1)
print(json.dumps(out, indent=1))
