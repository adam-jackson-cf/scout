#!/usr/bin/env python3
"""E1b-v2 stage 6: attach the incomplete-change outcome to the panel -> panel-final.csv.gz.

Usage: uv run --with pandas,numpy python finalize_panel.py
y_incomplete = 1 if any non-censored candidate follow-up pair of the parent has mean P(corrective) >= 0.5 over reps;
0 if all of its candidate pairs were labelled with mean P < 0.5 (or it has no candidate pair); NaN if the parent is
right-censored (merged < 14 d before T_end) or any of its pairs lacks a successful label.
y_incomplete_explicit (exploratory; deterministic regex rule, no model call): 1 if a candidate follow-up names the parent (#N) AND the follow-up
title/body matches fix|revert|regress|follow-up|forgot|missing|oversight; 0 otherwise; NaN if censored.
"""
import gzip
import json
import re
from collections import defaultdict

import numpy as np
import pandas as pd

from common import HERE

EXPL = re.compile(r"\b(fix(es|ed)?|revert(s|ed)?|regress\w*|follow[- ]?up|forgot\w*|missing|oversight|missed)\b", re.I)
TAU = 0.5


def outcome_rule(censored, ps, labelled):
    """None (NaN) if the parent is right-censored or any candidate pair lacks a successful label; 1.0 if any pair has mean
    P(corrective) >= TAU; else 0.0 (including parents with no candidate pair)."""
    if censored or not all(labelled):
        return None
    return float(any(p >= TAU for p in ps))


def main():
    panel = pd.read_csv(HERE / "panel.csv.gz")
    cands = [json.loads(l) for l in gzip.open(HERE / "candidates.jsonl.gz", "rt")]
    lab = defaultdict(list)
    lf = HERE / "labels.jsonl"
    if lf.exists():
        for l in lf.read_text().splitlines():
            r = json.loads(l)
            if r.get("ok"):
                lab[(r["repo"], r["parent"], r["follow"])].append(r["p_corrective"])
    raws = {}
    pairs = defaultdict(list)
    for c in cands:
        pairs[(c["repo"], c["parent"])].append(c)
    y, yx = [], []
    for repo, n, cens in zip(panel.repo, panel.number, panel.censored_followup):
        if cens:
            y.append(np.nan)
            yx.append(np.nan)
            continue
        cs = pairs.get((repo, n), [])
        ps = [float(np.mean(lab[(repo, n, c["follow"])])) if lab.get((repo, n, c["follow"])) else 0.0 for c in cs]
        lb = [bool(lab.get((repo, n, c["follow"]))) for c in cs]
        v = outcome_rule(False, ps, lb)
        y.append(np.nan if v is None else v)
        if repo not in raws:
            raws[repo] = {p["number"]: p for p in json.loads(gzip.open(HERE / "raw" / f"prs-{repo}.json.gz", "rt").read())}
        yx.append(float(any(c["mention"] and EXPL.search(raws[repo][c["follow"]]["title"] + "\n" + raws[repo][c["follow"]]["body"])
                            for c in cs)))
    panel["y_incomplete"] = y
    panel["y_incomplete_explicit"] = yx
    panel.to_csv(HERE / "panel-final.csv.gz", index=False)
    print(len(panel), "rows; y_incomplete mean", np.nanmean(y) if np.isfinite(y).any() else None)


if __name__ == "__main__":
    main()
