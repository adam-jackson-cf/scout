import glob, json, sys
print("arm repo agent_features agent_owned_paths(union) agent_owned_sum")
for st in sorted(glob.glob("/tmp/scout-x2c-smoke/*/*/state")):
    _, _, _, arm, repo, _ = st.split("/")
    fs = [json.load(open(p)) for p in glob.glob(st + "/features/*.json")]
    ag = [f for f in fs if f.get("source") == "agent-mapper" and f.get("status") != "stale"]
    o = [x["path"] for f in ag for x in f.get("ownedFiles", [])]
    print(arm, repo, len(ag), len(set(o)), len(o))
