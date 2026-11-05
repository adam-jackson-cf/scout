#!/usr/bin/env python3
"""Regenerate the deterministic association fixtures from frozen pilot inputs.

Reads recorded X2d inputs and fixture clones; never changes frozen study modules.
Use --out to choose the output directory (which must already exist).
"""

from __future__ import annotations

import argparse
import collections
import gzip
import hashlib
import importlib.util
import json
import os
import re
from collections.abc import Iterable
from dataclasses import dataclass
from typing import Any

PR_CAP = 10
NEVER = re.compile(r"(?!)")


def load(name: str, path: str) -> Any:
    spec = importlib.util.spec_from_file_location(name, path)
    if spec is None or spec.loader is None:
        raise ImportError(path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def cap_key(cid: str) -> str:
    return hashlib.sha256(("x2d-cap\0" + cid).encode()).hexdigest()


@dataclass
class FrontInput:
    entries: Any
    state: Any
    own: Any
    blobs: Any
    commit: str


def run_front(front: Any, inputs: FrontInput, neutralise: tuple[str, ...] = ()) -> Any:
    """Run one pilot commit, restoring every temporarily neutralised rule."""
    saved = (front.scope, front._LANG_DIR)
    if "scope" in neutralise:
        front.scope = lambda p: "x"
    if "lang" in neutralise:
        front._LANG_DIR = NEVER
    try:
        resolved: list[Any] = []
        counts, residual = front.owned_classify(
            inputs.entries, inputs.state, inputs.own, resolved
        )
        live, inert = front.clusters(residual, inputs.blobs)
    finally:
        front.scope, front._LANG_DIR = saved
    info = {p: d for p, _k, d in residual}
    for cluster in live:
        cluster["id"] = front.cluster_id(inputs.commit, cluster)
    order = sorted(live, key=lambda c: cap_key(c["id"]))
    counts = dict(counts, inert_clusters=len(inert), residual_clusters=len(live))
    return {
        "counts": {k: v for k, v in sorted(counts.items()) if v},
        "clusters": [
            dict(c, heuristic=[f for f in c["files"] if "heuristic_only" in info[f]])
            for c in (
                {k: c[k] for k in ("id", "dir", "stem", "files", "kinds")} for c in live
            )
        ],
        "inert": [{k: c[k] for k in ("dir", "stem", "files")} for c in inert],
        "decided": [c["id"] for c in order[:PR_CAP]],
        "capped": [c["id"] for c in order[PR_CAP:]],
        "resolved": sorted(
            (
                {k: x[k] for k in ("path", "rule", "owners", "old", "pool")}
                for x in resolved
            ),
            key=lambda x: x["path"],
        ),
    }


@dataclass
class RepoInput:
    name: str
    front: Any
    repo: str
    data: str
    rounds_info: Any
    rounds: Iterable[int]


def read_json(path: str) -> Any:
    with open(path) as stream:
        return json.load(stream)


def recorded_clusters(directory: str) -> tuple[Any, Any]:
    recorded: Any = collections.defaultdict(list)
    capped: Any = collections.defaultdict(list)
    with open(os.path.join(directory, "units.jsonl")) as stream:
        for line in stream:
            unit = json.loads(line)
            if unit.get("set") == "window" and not unit.get("control"):
                recorded[unit["commit"]].append(unit["cluster_id"])
    with open(os.path.join(directory, "capped.jsonl")) as stream:
        for line in stream:
            if line.strip():
                row = json.loads(line)
                capped[row["pr"]].append(row["cluster_id"])
    return recorded, capped


def commit_case(front: Any, inputs: FrontInput, parent: str) -> Any:
    pilot = run_front(front, inputs)
    expected = run_front(front, inputs, ("scope", "lang"))
    affected = [
        rule
        for rule, names in (
            ("cross_scope_rename", ("scope",)),
            ("translation", ("lang",)),
        )
        if run_front(front, inputs, names) != pilot
    ]
    if pilot != expected and not affected:
        affected = ["cross_scope_rename+translation"]
    added = [e[3] for e in inputs.entries if e[0] in "AC"]
    return {
        "sha": inputs.commit,
        "parent": parent,
        "entries": [
            [e[0], e[2], e[3], 1 if e[4] == front.EMPTY_BLOB else 0]
            for e in inputs.entries
        ],
        "tree": sorted({
            f
            for p in added
            for f in inputs.own._by_sib.get(front.sibling_key(p), [])
            if f != p and f in inputs.state.paths
        }),
        "pilot": pilot,
        "expected": expected,
        "affected": affected,
    }


def interesting(case: Any) -> bool:
    expected = case["expected"]
    return bool(
        case["affected"]
        or case["pilot"] != expected
        or expected["clusters"]
        or expected["inert"]
        or expected["resolved"]
        or set(expected["counts"]) - {"modified_owned", "deleted"}
    )


def build_round(job: RepoInput, round_number: int) -> Any:
    front = job.front
    info = next(x for x in job.rounds_info["rounds"] if x["round"] == round_number)
    directory = os.path.join(job.data, "rounds", str(round_number))
    membership = read_json(os.path.join(directory, "membership.json"))
    own = front.Ownership(membership)
    recorded, capped = recorded_clusters(directory)
    state = front.TreeState(front.tree_paths(job.repo, info["after_commit"]))
    commits: list[Any] = []
    touched: set[str] = set()
    added_paths: list[str] = []
    total = 0
    for sha, parent, _date, subject in front.first_parent_commits(
        job.repo, info["after_commit"], info["through_commit"]
    ):
        total += 1
        entries = front.diff_entries(job.repo, parent, sha)
        case = commit_case(
            front,
            FrontInput(
                entries, state, own, {e[3]: e[4] for e in entries if e[0] in "ARC"}, sha
            ),
            parent,
        )
        assert sorted(recorded.get(sha, [])) == sorted(case["pilot"]["decided"]), (
            job.name,
            round_number,
            sha,
            "recorded units differ from the pilot run",
        )
        assert sorted(
            capped.get(str(front.pr_number(subject) or sha[:8]), [])
        ) == sorted(case["pilot"]["capped"]), (
            job.name,
            round_number,
            sha,
            "recorded capped differ from the pilot run",
        )
        if interesting(case):
            commits.append(case)
            for entry in entries:
                touched.update((entry[2], entry[3]))
            added_paths += [e[3] for e in entries if e[0] in "AC"]
        state.apply(entries)
    keep = set(touched)
    for path in added_paths:
        keep.update(own._by_sib.get(front.sibling_key(path), []))
    return {
        "round": round_number,
        "commits_total": total,
        "commits": commits,
        "owners": {f: sorted(own._owners[f]) for f in sorted(keep) if f in own._owners},
        "heuristic_only": sorted(touched & set(membership.get("heuristic_only", []))),
        "pool": sorted(touched & set(membership.get("pool", []))),
    }


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--x2d", default="evaluation/x2d-merge-association-pilot")
    parser.add_argument("--scout", default=".")
    parser.add_argument("--fastapi", default="/tmp/scout-x2c-smoke/_src/fastapi")
    parser.add_argument("--axum", default="/tmp/scout-x2c-smoke/_src/axum")
    parser.add_argument(
        "--out", default=os.path.join(os.path.dirname(os.path.abspath(__file__)), "..")
    )
    args = parser.parse_args()
    axum_pilot = os.path.join(args.scout, "evaluation/x2d-postrun-tuning/axum/pilot")
    jobs = [
        RepoInput(
            "fastapi",
            load("frontend_fastapi", os.path.join(args.x2d, "frontend.py")),
            args.fastapi,
            os.path.join(args.x2d, "data/postrun"),
            read_json(os.path.join(args.x2d, "prepass/rounds.json")),
            range(1, 11),
        ),
        RepoInput(
            "axum",
            load("frontend_axum", os.path.join(axum_pilot, "frontend.py")),
            args.axum,
            os.path.join(args.x2d, "data/postrun-axum"),
            read_json(os.path.join(axum_pilot, "prepass/rounds.json")),
            range(1, 6),
        ),
    ]
    for job in jobs:
        result = {
            "repo": job.name,
            "rounds": [build_round(job, number) for number in job.rounds],
        }
        path = os.path.join(args.out, f"parity-{job.name}.json.gz")
        with gzip.GzipFile(path, "wb", mtime=0) as stream:
            stream.write(
                json.dumps(result, sort_keys=True, separators=(",", ":")).encode()
            )
        print(
            job.name,
            "commits stored",
            sum(len(r["commits"]) for r in result["rounds"]),
            "of",
            sum(r["commits_total"] for r in result["rounds"]),
            "bytes",
            os.path.getsize(path),
        )


if __name__ == "__main__":
    main()
