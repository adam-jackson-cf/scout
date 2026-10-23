"""Run ONLY after proposals are final. Writes E2-COMPARISON.md. python3 compare_e2.py"""
import json
prop = {}
for l in open("labels.jsonl"):
    r = json.loads(l)
    if r["provenance"] == "agent_reviewed": prop[r["case_id"]] = r["label"]
e2 = {}
for l in open("../jev-triage-v1/cases.jsonl"):
    r = json.loads(l); e2[r["case_id"]] = r["label"]
diff = sorted(i for i in e2 if prop[i] != e2[i])
n = len(e2)
o = ["# E2 comparison\n\n**Read after your review to avoid anchoring.**\n\n",
     "Proposed (agent) labels vs the original E2 adjudication on the %d labelled cases. No Jev output used.\n\n" % n,
     "- agree: %d/%d (%.1f%%)\n- differ: %d\n\n" % (n - len(diff), n, 100 * (n - len(diff)) / n, len(diff))]
if diff:
    o.append("| case_id | proposed | E2 |\n|---|---|---|\n")
    o += ["| %s | %s | %s |\n" % (i, prop[i], e2[i]) for i in diff]
open("E2-COMPARISON.md", "w").write("".join(o))
print(len(diff), n)
