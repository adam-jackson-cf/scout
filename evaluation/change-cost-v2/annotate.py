#!/usr/bin/env python3
"""E1b-v2 scripted blinded annotation for gates G4, G5, G7. NOT run before council sign-off (PREREGISTRATION section 5).

Usage (after `validate.py sample`; requires OPENROUTER_API_KEY in the environment, presence-checked, never printed or stored):
  uv run python annotate.py                # A and B on every item, then C on A/B disagreements
  uv run python annotate.py --adjudicate   # keep validation/annotations-{A,B}.jsonl, (re)run only C (round-11 amendment)
Annotators are PINNED chat models of a different family from Jev (`typesafe/*`), temperature 0, one call per item, the prompt is
RUBRIC text + the blinded item JSON + a fixed answer instruction (no pipeline label, no Jev output, no stratum, no repository):
  A = anthropic/claude-sonnet-4.5   B = openai/gpt-4.1   C = google/gemini-2.5-pro   (C only labels items where A and B disagree,
  i.e. `validate.disagree`: labels differ or either is `unparsable`)
The resolved model name returned by the API must equal the pinned id or start with it (dated suffix), else the script aborts. A truncated
(`finish_reason` length) or empty completion also aborts: it is an infrastructure failure, not an answer. max_tokens = MAX_TOKENS.
Writes validation/annotations-{A,B,C}.jsonl ({id, label, model_resolved, prompt_sha256}; `--adjudicate` writes only C) and
validation/annotator-prompts.json (models, temperature, max_tokens, rubric digests).
An answer outside the label list is recorded as `unparsable` (a disagreement: C adjudicates; unresolved if C gives no parsable label; never silently dropped).
"""
import hashlib
import json
import os
import sys
import urllib.error
import urllib.request
from concurrent.futures import ThreadPoolExecutor

from validate import LABELS, RUBRIC, VAL, disagree

URL = "https://openrouter.ai/api/v1/chat/completions"
MODELS = {"A": "anthropic/claude-sonnet-4.5", "B": "openai/gpt-4.1", "C": "google/gemini-2.5-pro"}
INSTR = "Answer with exactly one label from the list and nothing else."
KINDS = {"T": "type", "P": "pair", "D": "dep"}
MAX_TOKENS = 8192   # completion budget; reasoning models (C = gemini-2.5-pro) spend part of it on thinking before the label
REASONING = {"C": {"max_tokens": 2048}}   # bounded thinking for C (OpenRouter `reasoning`), leaving >= 6,144 tokens for the label


def request_body(letter, p):
    b = {"model": MODELS[letter], "temperature": 0, "max_tokens": MAX_TOKENS, "messages": [{"role": "user", "content": p}]}
    if letter in REASONING:
        b["reasoning"] = REASONING[letter]
    return b


def prompt(kind, item):
    return f"{RUBRIC[kind]}\n\nItem:\n{json.dumps(item, ensure_ascii=False)}\n\nLabels: {', '.join(LABELS[kind])}\n{INSTR}"


def parse(kind, text):
    t = (text or "").strip().lower().strip(".`'\" ")
    return t if t in LABELS[kind] else "unparsable"


def answer(letter, kind, d):
    """Label from one API response. Any completion that did not finish normally (`finish_reason` other than "stop": length,
    content_filter, error, missing) or is empty is an INFRASTRUCTURE failure, not an annotator answer: the run aborts instead
    of recording `unparsable` (Amendment 1: max_tokens 20 truncated every reasoning-model answer)."""
    resolved = d.get("model", "")
    if not resolved.startswith(MODELS[letter]):   # exact pinned id (dated suffixes allowed); recorded in model_resolved
        sys.exit(f"annotator {letter}: resolved model {resolved!r} does not match the pinned model {MODELS[letter]!r}")
    c = d["choices"][0]
    if c.get("finish_reason") != "stop" or not (c["message"].get("content") or "").strip():
        sys.exit(f"annotator {letter}: incomplete or empty completion (finish_reason {c.get('finish_reason')!r}); return to council")
    return resolved, parse(kind, c["message"]["content"])


def call(job, key, urlopen=urllib.request.urlopen):
    """One label, or abort. Three failed attempts (transport error, timeout, malformed response) abort the run: an infrastructure
    failure is never recorded as an annotator answer. Only an in-list check on a complete answer can yield `unparsable`."""
    letter, kind, item = job
    p = prompt(kind, item)
    body = json.dumps(request_body(letter, p)).encode()
    last = None
    for _ in range(3):
        try:
            r = urllib.request.Request(URL, body, {"Authorization": "Bearer " + key, "Content-Type": "application/json"})
            d = json.load(urlopen(r, timeout=300))
            resolved, label = answer(letter, kind, d)
            return dict(id=item["id"], label=label, model_resolved=resolved, prompt_sha256=hashlib.sha256(p.encode()).hexdigest())
        except urllib.error.HTTPError as e:
            if e.code in (401, 402):
                sys.exit(f"HTTP {e.code} from the annotation endpoint")
            last = f"HTTP {e.code}"
        except Exception as e:
            last = type(e).__name__
    sys.exit(f"annotator {letter}: item {item['id']} failed 3 attempts ({last}); nothing written, rerun with identical scripts")


def main():
    adjudicate_only = sys.argv[1:] == ["--adjudicate"]
    if sys.argv[1:] and not adjudicate_only:
        sys.exit("usage: annotate.py [--adjudicate]")
    key = os.environ.get("OPENROUTER_API_KEY")
    if not key:
        sys.exit("OPENROUTER_API_KEY missing: export it in the environment before running (required by the preregistration)")
    items = []
    for name in ("type", "pair", "dep"):
        items += [(KINDS[json.loads(l)["id"][0]], json.loads(l)) for l in (VAL / f"{name}-blinded.jsonl").read_text().splitlines()]
    out = {}
    if adjudicate_only:   # reuse existing A/B labels (they must cover every item); only C is (re)run
        lab = {l: {json.loads(r)["id"]: json.loads(r)["label"] for r in (VAL / f"annotations-{l}.jsonl").read_text().splitlines() if r.strip()}
               for l in ("A", "B")}
        if any(set(lab[l]) != {it[1]["id"] for it in items} for l in lab):
            sys.exit("--adjudicate: annotations-A/B do not cover exactly the blinded items")
    else:
        for letter in ("A", "B"):
            with ThreadPoolExecutor(8) as ex:
                out[letter] = list(ex.map(lambda it: call((letter, it[0], it[1]), key), items))
        lab = {l: {r["id"]: r["label"] for r in out[l]} for l in out}
    dis = [it for it in items if disagree(lab["A"][it[1]["id"]], lab["B"][it[1]["id"]])]
    with ThreadPoolExecutor(8) as ex:
        out["C"] = list(ex.map(lambda it: call(("C", it[0], it[1]), key), dis))
    for l, rows in out.items():
        (VAL / f"annotations-{l}.jsonl").write_text("\n".join(json.dumps(r) for r in rows) + "\n")
    digests = {k: hashlib.sha256((RUBRIC[k] + INSTR).encode()).hexdigest() for k in RUBRIC}
    # provenance per annotator: in --adjudicate mode A/B labels come from the original run at max_tokens 20 (Amendment 1)
    max_tokens = {l: (20 if adjudicate_only and l in ("A", "B") else MAX_TOKENS) for l in MODELS}
    (VAL / "annotator-prompts.json").write_text(json.dumps(dict(models=MODELS, temperature=0, max_tokens=max_tokens, reasoning=REASONING,
                                                                ab_reused_from_run_1=adjudicate_only, rubric_digests=digests), indent=1))
    print({l: len(r) for l, r in out.items()}, "C unparsable:", sum(r["label"] == "unparsable" for r in out["C"]))


if __name__ == "__main__":
    main()
