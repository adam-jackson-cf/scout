"""Run Jev variants over cases.jsonl. Usage (key via infisical env):
  python3 run_jev.py cases.jsonl raw/jev_<tag>.jsonl <model> <variant> [reps]
Stores request-free raw answers (no key, no state echo) + usage + latency."""
import os, sys, json, time, urllib.request, urllib.error
from concurrent.futures import ThreadPoolExecutor
from questions import VARIANTS
URL = "https://openrouter.ai/api/alpha/decisions"
cases_f, out_f, model, variant = sys.argv[1:5]
reps = int(sys.argv[5]) if len(sys.argv) > 5 else 3
key = os.environ["OPENROUTER_API_KEY"]
cases = [json.loads(l) for l in open(cases_f)]

def call(job):
    c, rep = job
    body = json.dumps({"model": model, "state": c["state"], "questions": VARIANTS[variant]}).encode()
    t = time.time()
    rec = {"case_id": c["case_id"], "variant": variant, "model_requested": model, "rep": rep}
    for attempt in range(3):
        try:
            r = urllib.request.Request(URL, body, {"Authorization": "Bearer " + key, "Content-Type": "application/json"})
            d = json.load(urllib.request.urlopen(r, timeout=60))
            rec.update(ok=True, model_resolved=d.get("model"), answers=d["answers"], usage=d.get("usage"), id=d.get("id"))
            break
        except urllib.error.HTTPError as e:
            rec.update(ok=False, http=e.code)
            if e.code in (401, 402): break
            time.sleep(2)
        except Exception as e:
            rec.update(ok=False, error=type(e).__name__); time.sleep(2)
    rec["latency_ms"] = round((time.time() - t) * 1000)
    return rec

jobs = [(c, r) for c in cases for r in range(reps)]
with ThreadPoolExecutor(4) as ex:
    res = list(ex.map(call, jobs))
os.makedirs(os.path.dirname(out_f) or ".", exist_ok=True)
open(out_f, "w").write("\n".join(json.dumps(r) for r in res) + "\n")
print(out_f, sum(r["ok"] for r in res), "/", len(res), "ok")
