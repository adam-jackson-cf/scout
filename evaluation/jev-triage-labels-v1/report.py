"""Build REVIEW.md and ab_agreement.json from calls.jsonl/labels.jsonl/blinded.jsonl. No E2/Jev labels. python3 report.py"""
import json
from collections import Counter

L = lambda f: [json.loads(x) for x in open(f)]
calls = {r["case_id"]: r for r in L("calls.jsonl")}
blind = {r["case_id"]: r["state"] for r in L("blinded.jsonl")}
prop = {r["case_id"]: r for r in L("labels.jsonl") if r["provenance"] == "agent_reviewed"}
dis = lambda c: c["C"] is not None
ids = sorted(calls, key=lambda i: (not dis(calls[i]), i))

# A/B agreement + Cohen's kappa over cases where both parsed
both = [c for c in calls.values() if c["A"]["ok"] and c["B"]["ok"]]
n = len(both); po = sum(c["A"]["label"] == c["B"]["label"] for c in both) / n
labs = ("real_defect", "false_positive", "cannot_tell")
pe = sum((sum(c["A"]["label"] == l for c in both) / n) * (sum(c["B"]["label"] == l for c in both) / n) for l in labs)
kappa = (po - pe) / (1 - pe) if pe < 1 else float("nan")
ndis = sum(dis(c) for c in calls.values())
unp = [i for i in ids if prop[i]["label"] == "unparsable"]
counts = Counter(p["label"] for p in prop.values())
json.dump({"n_both_parsed": n, "agree": po, "kappa": kappa, "n_adjudicated": ndis, "unparsable": unp,
           "proposed_counts": counts, "A_counts": Counter(c["A"]["label"] for c in calls.values()),
           "B_counts": Counter(c["B"]["label"] for c in calls.values())}, open("ab_agreement.json", "w"), indent=1)

o = []
if unp: o.append("**UNPARSABLE (no proposal): " + ", ".join(unp) + "**\n")
o.append("""# Review sheet: agent first-pass labels

Task: for each case below, read the excerpt and decide the label yourself. Confirm the proposed label or change it.
Labels are blind: the agents saw only the case state and the guide, never any tool verdict or earlier label. Every case needs your own decision; there is no bulk accept.
Guide (`triage-v1_choice/guide-1`): [GUIDE.md](GUIDE.md). Labels: real_defect | false_positive | cannot_tell.
After review, append your decisions to `labels.jsonl` as provenance `human`, by `user`.

Summary: %d cases; proposed %s; A/B disagreements adjudicated by C: %d; unparsable: %d.
Disagreements are listed first, then the rest by case_id.
""" % (len(ids), ", ".join("%s=%d" % kv for kv in sorted(counts.items())), ndis, len(unp)))

def fmt(who, x):
    if x is None: return ""
    if not x["ok"]: return "- **%s**: unparsable\n" % who
    return "- **%s** (%s): `%s` — rule: %s — %s\n" % (who, x["model"], x["label"], x["rule"], x["reason"])

for i in ids:
    c, s = calls[i], blind[i]
    o.append("\n---\n\n## %s%s\n\n- tool_signal: %s\n- path: `%s`\n\n" % (i, " (A/B disagree)" if dis(c) else "", s["tool_signal"], s["path"]))
    for e in s["excerpts"]:
        o.append("`%s` lines %s\n```\n%s\n```\n\n" % (e.get("file"), e.get("lines"), e.get("text")))
    o.append(fmt("A", c["A"]) + fmt("B", c["B"]))
    if c["C"]:
        o.append(fmt("C", c["C"]).rstrip("\n") + (" — agrees with: %s\n" % c["C"].get("which") if c["C"]["ok"] else "\n"))
    o.append("\n**Proposed label: `%s`**\n\nYour decision: [ ] confirm  [ ] change to ______  note: ____\n" % prop[i]["label"])
open("REVIEW.md", "w").write("".join(o))
print(json.dumps({"agree": po, "kappa": kappa, "n": n}))
