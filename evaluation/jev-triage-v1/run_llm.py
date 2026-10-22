"""Reasoning-LLM baseline via `claude -p` (no tools). Usage: python3 run_llm.py <variant> <out.jsonl> [model]
Uses the same instruction + criteria text as the Jev variant, same de-leaked state."""
import sys, json, subprocess, time
from concurrent.futures import ThreadPoolExecutor
from questions import VARIANTS
variant, out = sys.argv[1], sys.argv[2]
model = sys.argv[3] if len(sys.argv) > 3 else None
q = VARIANTS[variant]["q"]
cases = [json.loads(l) for l in open("cases.jsonl")]
crit = "\n".join(f"- {k}: {v}" for k, v in q["criteria"].items())

def call(c):
    prompt = (f"{q['instructions']}\n\nOptions:\n{crit}\n\nState (JSON):\n{json.dumps(c['state'])}\n\n"
              'Reply with ONLY a JSON object: {"label": "<one option name>", "reason": "<one sentence>"}')
    cmd = ["claude", "-p", "--output-format", "json", "--tools", "", "--no-session-persistence", "--disable-slash-commands"]
    if model: cmd += ["--model", model]
    t = time.time()
    p = subprocess.run(cmd, input=prompt, capture_output=True, text=True, timeout=300, cwd="/tmp")
    rec = {"case_id": c["case_id"], "variant": variant, "latency_ms": round((time.time() - t) * 1000), "rc": p.returncode}
    try:
        d = json.loads(p.stdout); rec["cost_usd"] = d.get("total_cost_usd"); rec["result"] = d.get("result")
        rec["model_usage"] = list((d.get("modelUsage") or {}).keys())
        import re; m = re.search(r'"label"\s*:\s*"(\w+)"', d.get("result") or ""); rec["label"] = m.group(1) if m else None
    except Exception as e:
        rec["error"] = p.stdout[:200] + p.stderr[:200]
    return rec

with ThreadPoolExecutor(4) as ex: res = list(ex.map(call, cases))
open(out, "w").write("\n".join(json.dumps(r) for r in res) + "\n")
print(out, sum(1 for r in res if r.get("label")), "/", len(res))
