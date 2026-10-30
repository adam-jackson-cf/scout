"""A stub of OpenRouter's alpha/decisions route for X2d validation: no model, no network beyond 127.0.0.1.

It answers the questions the X2d drivers send (`jev/main.go`: merge.covered, merge.which_area, merge.area_covers; `onboard/main.go`: onboard.covered,
onboard.which_file_area, onboard.same_responsibility) from a PLANTED truth, never from a model. Each request carries exactly one question, key `q1`:

  Noul   (criteria keys true/false) kinds, recognised from the state:
    state.area_a and state.area_b present      -> onboard.same_responsibility: p_true 0.95 when the pair of titles is listed in truth["_duplicates"], else 0.05
    state.area present (and state.unit)        -> merge.area_covers (fan-out): truth "area" == state.area.id
    state.file present                         -> onboard.covered: truth "covered"
    otherwise (state.unit, state.catalogue)    -> merge.covered: truth "covered"
  Choice (criteria keys = area ids) kinds: state.file -> onboard.which_file_area, otherwise merge.which_area: the correct option is truth "area".

Truth file format (JSON, `Stub(truth=...)` takes the loaded dict): {"<path>": {"area": "<area id>" | null, "covered": true|false}, ...}. The key is the unit's FIRST path,
`state.unit.paths[0]` (merge) or `state.file.path` (onboard); a V-title request has no files (state.pr.title only), so for it the key is "title:<state.pr.title>". Optional key
"_duplicates": [[titleA, titleB], ...] (order irrelevant). A request whose key is not in the truth is answered as {"area": null, "covered": false} and recorded in `unknown_keys`.
With the driver's `include_pr_title: false` setting the V-sem/V-sib states carry no `pr.title` (the key stays `unit.paths[0]`; the logged `state` shows it); V-title requests always carry it.

Answer correctness is drawn deterministically from sha256(seed + sha256 of the canonical request body): the answer is the planted one with probability `p_correct`, otherwise the
opposite. Noul: p_true 0.95 for a planted true / 0.05 for a planted false (wrong draw: swapped). Choice: probability 0.9 on the selected option, the rest spread evenly; the selected
option is the planted area, or, when the draw is wrong or the planted area is null or not an option, another option picked by the hash. `confidence` is the selected probability.
Modes: "planted" (above), "constant" (Noul p_true 0.5, Choice uniform over the options with the hash-picked option as `choice`: the planted null), "reverse" (Noul 1 - p; Choice never
the planted area), "order" (Choice: the FIRST option of the request's option order gets 0.9, whatever the truth: an order artefact the second permutation must expose).

The stub validates what the product client must send and records violations in `violations` (and answers HTTP 400 so the driver shows a failure): exactly one question with key q1, a
known type, true/false criteria for a Noul and 2..255 criteria for a Choice, `model` equal to the stub's pin, a `state` object, an Authorization Bearer header, and no key material
("sk-or") anywhere in the body. It logs, per request, to `log_path` (JSONL): n, model, question keys, type, option order (`criteria_keys`), instructions, criteria, state, key
PRESENCE and length (never the value), `body_has_key`, the truth key.

Fault injection (as X2c's stub): `fail_ids` (HTTP 500 for requests whose truth key or state.area.id is in the set), `model_for` ({truth key or area id: model}: answer from another
model), `no_cost` (omit usage.cost), `rate_limit_first` (429 with Retry-After: 0 for the first N requests), `drop_option` (an option id, or "$first" for the first listed option: omit it
from `probabilities`; "noul": omit the Noul probability), `scale` (multiply every Choice probability, so they sum to `scale` instead of 1). `requests` counts requests.

    s = Stub("typesafe/jev-1.14-20261014", truth={...}, p_correct=1.0, seed="s").start(); url = s.url; ...; s.stop()
"""
import hashlib
import json
import threading
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

P_HIGH, P_LOW = 0.95, 0.05
CHOICE_MASS = 0.9


def _h(text):
    return int(hashlib.sha256(text.encode("utf-8")).hexdigest()[:16], 16)


def truth_key(state):
    """The truth key of a request state: first path of the unit, the file path, or `title:<pr title>` for a V-title request."""
    if isinstance(state.get("unit"), dict) and state["unit"].get("paths"):
        return state["unit"]["paths"][0]
    if isinstance(state.get("file"), dict) and state["file"].get("path"):
        return state["file"]["path"]
    pr = state.get("pr")
    if isinstance(pr, dict) and "title" in pr:
        return "title:" + pr["title"]
    return None


class Stub:
    def __init__(self, model, truth=None, p_correct=1.0, seed="x2d", log_path=None, fail_ids=(), model_for=None, no_cost=False, rate_limit_first=0, drop_option=None,
                 scale=None, mode="planted"):
        assert mode in ("planted", "constant", "order", "reverse"), mode
        self.model, self.truth, self.p_correct, self.seed, self.log_path, self.mode = model, dict(truth or {}), p_correct, str(seed), log_path, mode
        self.fail_ids, self.model_for, self.no_cost, self.rate_limit_first = set(fail_ids), dict(model_for or {}), no_cost, rate_limit_first
        self.drop_option, self.scale = drop_option, scale
        self.requests, self.violations, self.unknown_keys = 0, [], []
        self.lock, self.server, self.thread = threading.Lock(), None, None
        outer = self

        class H(BaseHTTPRequestHandler):
            def log_message(self, *a):
                pass

            def _send(self, code, obj=None, retry_after=None):
                raw = json.dumps(obj).encode() if obj is not None else b""
                self.send_response(code)
                if retry_after is not None:
                    self.send_header("Retry-After", str(retry_after))
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(raw)))
                self.end_headers()
                try:
                    self.wfile.write(raw)
                except (BrokenPipeError, ConnectionResetError):  # the driver cancelled this request (cost stop): nothing to answer
                    pass

            def do_POST(self):
                body_raw = self.rfile.read(int(self.headers.get("Content-Length", "0")))
                body = json.loads(body_raw)
                with outer.lock:
                    outer.requests += 1
                    n = outer.requests
                if n <= outer.rate_limit_first:
                    return self._send(429, retry_after=0)
                auth = self.headers.get("Authorization", "")
                state = body.get("state") if isinstance(body.get("state"), dict) else {}
                questions = body.get("questions") if isinstance(body.get("questions"), dict) else {}
                q = questions.get("q1") if isinstance(questions.get("q1"), dict) else {}
                criteria = q.get("criteria") if isinstance(q.get("criteria"), dict) else {}
                bad = []
                if body.get("model") != outer.model:
                    bad.append("model %r is not the pin %r" % (body.get("model"), outer.model))
                if sorted(questions) != ["q1"]:
                    bad.append("question keys %r are not exactly ['q1']" % sorted(questions))
                if q.get("type") not in ("noul", "choice"):
                    bad.append("unknown question type %r" % q.get("type"))
                if q.get("type") == "noul" and sorted(criteria) != ["false", "true"]:
                    bad.append("noul criteria keys %r" % sorted(criteria))
                if q.get("type") == "choice" and not 2 <= len(criteria) <= 255:
                    bad.append("choice with %d options" % len(criteria))
                if not isinstance(body.get("state"), dict):
                    bad.append("state is not an object")
                if not auth.startswith("Bearer ") or len(auth) <= len("Bearer "):
                    bad.append("no bearer key")
                if "sk-or" in body_raw.decode("utf-8", "replace"):
                    bad.append("key material in the body")
                key = truth_key(state)
                if outer.log_path:
                    with outer.lock, open(outer.log_path, "a") as f:
                        f.write(json.dumps({"n": n, "model": body.get("model"), "question_keys": sorted(questions), "type": q.get("type"), "criteria_keys": list(criteria),
                                            "instructions": q.get("instructions"), "criteria": criteria, "state": state, "auth_present": auth.startswith("Bearer "),
                                            "auth_len": len(auth), "body_has_key": "sk-or" in body_raw.decode("utf-8", "replace"), "truth_key": key}) + "\n")
                if bad:
                    with outer.lock:
                        outer.violations.append({"n": n, "problems": bad})
                    return self._send(400, {"error": "; ".join(bad)})
                area_id = state["area"]["id"] if isinstance(state.get("area"), dict) and "id" in state["area"] else None
                if outer.fail_ids & {key, area_id}:
                    return self._send(500)
                model = next((m for k, m in outer.model_for.items() if k in (key, area_id)), outer.model)
                draw = (_h(outer.seed + hashlib.sha256(json.dumps(body, sort_keys=True).encode()).hexdigest()) % 10**6) / 10**6
                right = draw < outer.p_correct
                usage = {"prompt_tokens": 100 + len(json.dumps(state)) // 4, "completion_tokens": 1}
                if not outer.no_cost:
                    usage["cost"] = 0.00001 * len(json.dumps(state)) / 1000
                if q["type"] == "noul":
                    answer = outer._noul(state, key, area_id, right)
                else:
                    answer = outer._choice(state, key, list(criteria), right, draw, criteria)
                if outer.drop_option == "noul":
                    answer.pop("noul", None)
                self._send(200, {"id": "gen-%d" % n, "model": model, "usage": usage, "answers": {"q1": answer}})

        self.server = ThreadingHTTPServer(("127.0.0.1", 0), H)

    def _truth(self, key):
        if key in self.truth:
            return self.truth[key]
        with self.lock:
            self.unknown_keys.append(key)
        return {"area": None, "covered": False}

    def _noul(self, state, key, area_id, right):
        if isinstance(state.get("area_a"), dict) and isinstance(state.get("area_b"), dict):
            pair = {state["area_a"].get("title"), state["area_b"].get("title")}
            planted = any(set(p) == pair for p in self.truth.get("_duplicates", []))
        elif area_id is not None:
            t = self._truth(key)
            planted = t["area"] == area_id or (t.get("area_title") is not None and isinstance(state.get("area"), dict) and state["area"].get("title") == t["area_title"])
        else:
            planted = bool(self._truth(key)["covered"])
        if self.mode == "constant":
            p = 0.5
        else:
            p = P_HIGH if planted == right else P_LOW
            if self.mode == "reverse":
                p = 1 - p
        return {"type": "noul", "noul": p, "confidence": max(p, 1 - p)}

    def _choice(self, state, key, options, right, draw, criteria=None):
        t = self._truth(key)
        planted = t["area"]
        if t.get("area_title") is not None and criteria:  # semantic truth: the option whose criterion text starts with the planted TITLE (follows a catalogue whose titles were permuted)
            planted = next((o for o in options if str(criteria.get(o, "")).startswith(t["area_title"] + ":")), None)
        pick = options[int(draw * 10**6) % len(options)]
        if self.mode == "order":
            sel = options[0]
        elif self.mode == "constant":
            sel = pick
        elif self.mode == "reverse":
            sel = pick if pick != planted else options[(options.index(pick) + 1) % len(options)]
        elif planted in options and right:
            sel = planted
        elif planted is None or planted not in options:  # a gold-none unit has no right option: a fixed one (the order-independent answer a stable reader would give), so K5's flip rate reflects covered units only
            sel = min(options)
        else:
            others = [o for o in options if o != planted]
            sel = others[int(draw * 10**6) % len(others)]
        if self.mode == "constant":
            probs = {o: 1.0 / len(options) for o in options}
        else:
            rest = (1 - CHOICE_MASS) / (len(options) - 1)
            probs = {o: (CHOICE_MASS if o == sel else rest) for o in options}
        conf = probs[sel]
        if self.scale is not None:
            probs = {k: v * self.scale for k, v in probs.items()}
        if self.drop_option:
            probs.pop(options[0] if self.drop_option == "$first" else self.drop_option, None)
        return {"type": "choice", "choice": sel, "probabilities": probs, "confidence": conf}

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
