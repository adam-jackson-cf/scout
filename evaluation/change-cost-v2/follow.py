#!/usr/bin/env python3
"""E1b-v2 stage 4: candidate follow-up pairs for the incomplete-change outcome (H4).

Usage: uv run --with pandas python follow.py
Reads panel.csv.gz, data/pr-<repo>.jsonl.gz (partners), raw/prs-<repo>.json.gz, pins.json.
Writes candidates.jsonl.gz: one record per (parent, follow-up) pair with the strict timing rules:
  follow-up F of parent P iff  F != P, F not bot, F has <= 50 files, F.created >= P.merged (strictly after the parent
  merge, so concurrent work is never a follow-up), F.merged <= P.merged + 14 d, and F shares an eligible file with P,
  or touches an eligible co-change partner of P's files, or names P (#<number>) in its title/body.
Parents merged within 14 d of T_end are right-censored (not given an outcome). The deterministic part decides only
which pairs are *judged*; whether a pair is a correction/completion is decided by the validated Jev label
(jev_label.py) -- hidden-coupling status of the parent is never shown to the labeller.
"""
import bisect
import gzip
import json
import re

import pandas as pd

from common import FOLLOW_DAYS, HERE, NAMES, eligible, is_bot, load_pins

DAY = 86400


def ts(s):
    return pd.Timestamp(s).timestamp()


def is_follow(P, F, same):
    """Strict timing rule: F is a different PR, was CREATED at or after P's merge (concurrent work never counts), and was
    merged within FOLLOW_DAYS after P (so merged >= created >= P.merged: no negative gaps)."""
    return (not same) and F["created"] >= P["merged"] and P["merged"] <= F["merged"] <= P["merged"] + FOLLOW_DAYS * DAY


def main():
    panel = pd.read_csv(HERE / "panel.csv.gz")
    pins = load_pins()
    out, stats = [], {}
    for repo in NAMES:
        f = HERE / "data" / f"pr-{repo}.jsonl.gz"
        if not f.exists():
            continue
        rows = {r["number"]: r for r in (json.loads(l) for l in gzip.open(f, "rt"))}
        raw = {p["number"]: p for p in json.loads(gzip.open(HERE / "raw" / f"prs-{repo}.json.gz", "rt").read())}
        info = {}
        for n, p in raw.items():
            if not p["merged"] or is_bot(p["author"], p["author_type"]) or p["n_files"] > 50:
                continue
            info[n] = dict(created=ts(p["created"]), merged=ts(p["merged"]),
                           files={x["path"] for x in p["files"] if eligible(x["path"])})
        order = sorted(info, key=lambda n: info[n]["merged"])
        mtimes = [info[n]["merged"] for n in order]
        t_end = pins[repo]["t_end_ts"]
        sub = panel[panel.repo == repo]
        npairs = 0
        for n in sub.number:
            P = info.get(n)
            if P is None:
                continue
            partners = {x["g"] for x in rows[n].get("partners", [])}
            mention = re.compile(r"(?<![\w/])#%d(?!\d)" % n)
            cens = P["merged"] > t_end - FOLLOW_DAYS * DAY
            lo = bisect.bisect_left(mtimes, P["merged"])
            hi = bisect.bisect_right(mtimes, P["merged"] + FOLLOW_DAYS * DAY)
            for m in order[lo:hi]:
                F = info[m]
                if not is_follow(P, F, m == n):
                    continue
                own, part = F["files"] & P["files"], F["files"] & partners
                ment = bool(mention.search(raw[m]["title"] + "\n" + raw[m]["body"]))
                if not (own or part or ment):
                    continue
                out.append(dict(repo=repo, parent=n, follow=m, gap_h=(F["merged"] - P["merged"]) / 3600.0,
                                kind=("own" if own else "partner" if part else "mention"),
                                mention=ment, censored=cens, shared=sorted(own | part)[:12]))
                npairs += 1
        stats[repo] = dict(parents=len(sub), pairs=npairs)
    with gzip.open(HERE / "candidates.jsonl.gz", "wt") as fh:
        for r in out:
            fh.write(json.dumps(r) + "\n")
    print(json.dumps(stats))


if __name__ == "__main__":
    main()
