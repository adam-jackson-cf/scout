"""Build blinded.jsonl + sources.jsonl (seeded shuffle). Run from this folder: python3 blind.py"""
import json, random
SEED = 20261022
SRC = "../jev-triage-v1/"
rows = []
for f in ("cases.jsonl", "unlabelled.jsonl"):
    for l in open(SRC + f):
        r = json.loads(l)
        rows.append((r["case_id"], r["state"], "evaluation/jev-triage-v1/" + f))
random.Random(SEED).shuffle(rows)
with open("blinded.jsonl", "w") as b, open("sources.jsonl", "w") as s:
    for cid, st, f in rows:
        b.write(json.dumps({"case_id": cid, "state": st}) + "\n")
        s.write(json.dumps({"case_id": cid, "source_file": f}) + "\n")
