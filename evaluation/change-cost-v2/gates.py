#!/usr/bin/env python3
"""E1b-v2 gates G1-G3 (PREREGISTRATION §5). Distribution/coverage gates only, EXCEPT G3b which is a covariate-outcome check
and is computed only with --g3b (run time, after council sign-off; never during design review).

Usage: uv run --with pandas,numpy python gates.py [--g3b] > gates.json
G4 and G5 (annotator validation) are scored by validate.py; G6 is tests.py + harness.py.
"""
import json
import sys

import numpy as np
import pandas as pd

from common import HERE

FLOOR = np.log(0.5)


def main():
    funnel = json.loads((HERE / "funnel.json").read_text())
    gate = json.loads((HERE / "repo-gate.json").read_text())
    df = pd.read_csv(HERE / "panel.csv.gz")
    out = dict(G1={}, G2=gate)
    for repo, f in funnel.items():
        # `ok` already contains the PRs later classed nondev/type_unlabelled (those are subsets of ok, not extra rows);
        # reached = every PR that passed the bot/bulk/commit/eligible-file filters and therefore needed a snapshot
        reach = sum(f.get(k, 0) for k in ("ok", "no_snapshot", "no_preexisting_file", "no_measurable_file"))
        ns = f.get("no_snapshot", 0)
        out["G1"][repo] = dict(reached=reach, no_snapshot=ns, coverage=(1 - ns / reach) if reach else None,
                               pass_=bool(reach and 1 - ns / reach >= 0.85))
    y = df.y_active.dropna()
    out["G3"] = dict(n=len(y), floor_share=float((y <= FLOOR + 1e-9).mean()), pass_=bool((y <= FLOOR + 1e-9).mean() <= 0.60))
    if "--g3b" in sys.argv:
        d = df.dropna(subset=["y_active"])
        r = lambda c: d[c] - d.groupby("repo")[c].transform("mean")
        rho = pd.Series(r("y_active").to_numpy()).rank().corr(pd.Series(r("log_churn").to_numpy()).rank())
        out["G3b"] = dict(spearman_y_active_log_churn=float(rho), pass_=bool(rho >= 0.15))
    print(json.dumps(out, indent=1))


if __name__ == "__main__":
    main()
