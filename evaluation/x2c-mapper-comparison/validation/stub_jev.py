"""A stub of OpenRouter's alpha/decisions route for X2c validation: no model, no network beyond 127.0.0.1.

It answers the one question the X2c scorer sends (a Choice over the four guide outcomes) from the PLANTED structure of the synthetic features, never from a model:
  mode "planted"  : titles `<family> <i> <token>`; both sides the same family and index -> P(same) = FAMILY_P[family]; any other pair -> P(same) = OTHER_P.
  mode "constant" : P(same) = CONSTANT_P for every pair (an uninformative scorer: the planted null).
  mode "reverse"  : the planted scores inverted (1 - p): the planted adverse scorer.
  mode "order"    : P(same) depends on which side is `feature_a` (a deliberate order artefact the two-order average must remove).
The stub validates what the product client must send (model, `q1` key only, criteria, no key in the body) and logs, per request, the question text, the state, the
Authorization header's PRESENCE and length (never its value). Fault injection: `fail_pairs` (HTTP 500 for those pair ids), `model_for` (answer from another model for
those pair ids), `no_cost` (omit usage.cost), `rate_limit_first` (429 with Retry-After: 0 for the first N requests), `drop_option` (omit that option from `probabilities`), `scale` (multiply every
probability by it, so the four options sum to `scale` instead of 1: an unnormalised reply).
"""
import hashlib
import json
import re
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

FAMILY_P = {"Concern": 0.92, "Redrawn": 0.88, "Volatile": 0.10, "Trap": 0.04}
OTHER_P = 0.03
CONSTANT_P = 0.5
TITLE = re.compile(r"^(Concern|Redrawn|Volatile|Trap)\b.*?\b(\d+)\b")


def p_same(state, mode):
    a, b = state["feature_a"], state["feature_b"]
    ma, mb = TITLE.match(a["title"]), TITLE.match(b["title"])
    if mode == "constant":
        return CONSTANT_P
    if mode == "order":
        return 0.95 if a["title"] < b["title"] else 0.05
    p = FAMILY_P[ma.group(1)] if ma and mb and ma.group(1) == mb.group(1) and ma.group(2) == mb.group(2) else OTHER_P
    return 1 - p if mode == "reverse" else p


class Stub:
    def __init__(self, model, mode="planted", log_path=None, fail_pairs=(), model_for=None, no_cost=False, rate_limit_first=0, drop_option=None, scale=None):
        self.model, self.mode, self.log_path = model, mode, log_path
        self.fail_pairs, self.model_for, self.no_cost, self.rate_limit_first, self.drop_option = set(fail_pairs), dict(model_for or {}), no_cost, rate_limit_first, drop_option
        self.scale = scale
        self.requests, self.lock, self.server, self.thread = 0, threading.Lock(), None, None
        outer = self

        class H(BaseHTTPRequestHandler):
            def log_message(self, *a):
                pass

            def do_POST(self):
                body = json.loads(self.rfile.read(int(self.headers.get("Content-Length", "0"))))
                with outer.lock:
                    outer.requests += 1
                    n = outer.requests
                if n <= outer.rate_limit_first:
                    self.send_response(429)
                    self.send_header("Retry-After", "0")
                    self.end_headers()
                    return
                auth = self.headers.get("Authorization", "")
                state = body["state"]
                questions = body["questions"]
                q = questions["q1"]
                shared = state["shared_files"]
                pair_hint = hashlib.sha256(json.dumps(state, sort_keys=True).encode()).hexdigest()[:12]
                if outer.log_path:
                    with outer.lock, open(outer.log_path, "a") as f:
                        f.write(json.dumps({"n": n, "model": body["model"], "question_keys": sorted(questions), "type": q["type"], "criteria_keys": list(q["criteria"]),
                                            "instructions": q["instructions"], "criteria": q["criteria"], "state": state, "auth_present": auth.startswith("Bearer "),
                                            "auth_len": len(auth), "body_has_key": "sk-" in json.dumps(body), "pair_hint": pair_hint}) + "\n")
                titles = (state["feature_a"]["title"], state["feature_b"]["title"])
                if any(t in outer.fail_pairs for t in titles):
                    self.send_response(500)
                    self.end_headers()
                    return
                p = p_same(state, outer.mode)
                # the residual mass goes to different_concern: the planted score s = P(same) + P(overlapping) equals the planted P(same)
                probs = {"same_concern": p, "overlapping_concern": 0.0, "different_concern": 1 - p, "cannot_tell": 0.0}
                if outer.scale is not None:
                    probs = {k: v * outer.scale for k, v in probs.items()}
                if outer.drop_option:
                    probs.pop(outer.drop_option)
                model = next((m for t, m in outer.model_for.items() if t in titles), outer.model)
                usage = {"prompt_tokens": 100 + len(json.dumps(state)) // 4, "completion_tokens": 1}
                if not outer.no_cost:
                    usage["cost"] = 0.00001 * len(json.dumps(state)) / 1000
                out = {"id": "gen-%d" % n, "model": model, "usage": usage, "answers": {"q1": {"type": "choice", "choice": max(probs, key=probs.get), "probabilities": probs, "confidence": 0.9}}}
                raw = json.dumps(out).encode()
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(raw)))
                self.end_headers()
                self.wfile.write(raw)

        self.server = ThreadingHTTPServer(("127.0.0.1", 0), H)

    @property
    def url(self):
        return "http://127.0.0.1:%d/api/alpha/decisions" % self.server.server_address[1]

    def start(self):
        self.thread = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.thread.start()
        return self

    def stop(self):
        self.server.shutdown()
        self.server.server_close()
