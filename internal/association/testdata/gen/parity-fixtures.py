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
from collections import Counter
from collections.abc import Callable, Iterable
from dataclasses import dataclass
from typing import NotRequired, Protocol, TypedDict, cast

PR_CAP = 10
NEVER = re.compile(r"(?!)")

Entry = tuple[str, int, str, str, str]
Residual = tuple[str, str, str]
SiblingKey = tuple[str, str, str]
CommitRow = tuple[str, str | None, str, str]


class Membership(TypedDict, total=False):
    areas: dict[str, list[str]]
    pool: list[str]
    heuristic_only: list[str]


class RoundInfo(TypedDict):
    round: int
    after_commit: str
    through_commit: str


class RoundsInfo(TypedDict):
    rounds: list[RoundInfo]


class FrontCluster(TypedDict):
    dir: str
    stem: str
    files: list[str]
    kinds: list[str]
    role: str
    scope: str
    id: NotRequired[str]


class FrontResolved(TypedDict):
    path: str
    rule: str
    owners: list[str]
    old: str | None
    blob: str
    pool: bool


class ClusterRecord(TypedDict):
    id: str
    dir: str
    stem: str
    files: list[str]
    kinds: list[str]
    heuristic: list[str]


class InertRecord(TypedDict):
    dir: str
    stem: str
    files: list[str]


class ResolvedRecord(TypedDict):
    path: str
    rule: str
    owners: list[str]
    old: str | None
    pool: bool


class FrontResult(TypedDict):
    counts: dict[str, int]
    clusters: list[ClusterRecord]
    inert: list[InertRecord]
    decided: list[str]
    capped: list[str]
    resolved: list[ResolvedRecord]


class CommitCase(TypedDict):
    sha: str
    parent: str | None
    entries: list[list[str | int]]
    tree: list[str]
    pilot: FrontResult
    expected: FrontResult
    affected: list[str]


class RoundRecord(TypedDict):
    round: int
    commits_total: int
    commits: list[CommitCase]
    owners: dict[str, list[str]]
    heuristic_only: list[str]
    pool: list[str]


class RepoFixture(TypedDict):
    repo: str
    rounds: list[RoundRecord]


class TreeView(Protocol):
    paths: set[str]

    def apply(self, entries: list[Entry]) -> None: ...


class OwnershipView(Protocol):
    _owners: dict[str, set[str]]
    _by_sib: dict[SiblingKey, list[str]]


ScopeRule = Callable[[str], str]
OwnershipFactory = Callable[[Membership], OwnershipView]
TreeFactory = Callable[[list[str]], TreeView]


class FrontEnd(Protocol):
    EMPTY_BLOB: str
    _LANG_DIR: re.Pattern[str]
    scope: ScopeRule
    Ownership: OwnershipFactory
    TreeState: TreeFactory

    def owned_classify(
        self,
        entries: list[Entry],
        state: TreeView,
        ownership: OwnershipView,
        resolved: list[FrontResolved],
    ) -> tuple[Counter[str], list[Residual]]: ...

    def clusters(
        self, residual: list[Residual], blobs: dict[str, str]
    ) -> tuple[list[FrontCluster], list[FrontCluster]]: ...

    def cluster_id(self, commit: str, cluster: FrontCluster) -> str: ...

    def sibling_key(self, path: str) -> SiblingKey: ...

    def tree_paths(self, repo: str, rev: str) -> list[str]: ...

    def first_parent_commits(
        self, repo: str, base: str, tip: str
    ) -> list[CommitRow]: ...

    def diff_entries(self, repo: str, parent: str | None, sha: str) -> list[Entry]: ...

    def pr_number(self, subject: str) -> int | None: ...


def load(name: str, path: str) -> FrontEnd:
    spec = importlib.util.spec_from_file_location(name, path)
    if spec is None or spec.loader is None:
        raise ImportError(path)
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return cast(FrontEnd, module)


def cap_key(cid: str) -> str:
    return hashlib.sha256(("x2d-cap\0" + cid).encode()).hexdigest()


@dataclass
class FrontInput:
    entries: list[Entry]
    state: TreeView
    own: OwnershipView
    blobs: dict[str, str]
    commit: str


def cluster_fields(cluster: FrontCluster, info: dict[str, str]) -> ClusterRecord:
    return {
        "id": cluster["id"],
        "dir": cluster["dir"],
        "stem": cluster["stem"],
        "files": cluster["files"],
        "kinds": cluster["kinds"],
        "heuristic": [f for f in cluster["files"] if "heuristic_only" in info[f]],
    }


def inert_fields(cluster: FrontCluster) -> InertRecord:
    return {"dir": cluster["dir"], "stem": cluster["stem"], "files": cluster["files"]}


def resolved_fields(resolved: FrontResolved) -> ResolvedRecord:
    return {
        "path": resolved["path"],
        "rule": resolved["rule"],
        "owners": resolved["owners"],
        "old": resolved["old"],
        "pool": resolved["pool"],
    }


def run_front(
    front: FrontEnd, inputs: FrontInput, neutralise: tuple[str, ...] = ()
) -> FrontResult:
    """Run one pilot commit, restoring every temporarily neutralised rule."""
    saved = (front.scope, front._LANG_DIR)
    if "scope" in neutralise:
        front.scope = lambda p: "x"
    if "lang" in neutralise:
        front._LANG_DIR = NEVER
    try:
        resolved: list[FrontResolved] = []
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
    all_counts = dict(counts, inert_clusters=len(inert), residual_clusters=len(live))
    return {
        "counts": {k: v for k, v in sorted(all_counts.items()) if v},
        "clusters": [cluster_fields(c, info) for c in live],
        "inert": [inert_fields(c) for c in inert],
        "decided": [c["id"] for c in order[:PR_CAP]],
        "capped": [c["id"] for c in order[PR_CAP:]],
        "resolved": sorted(
            (resolved_fields(x) for x in resolved),
            key=lambda x: x["path"],
        ),
    }


@dataclass
class RepoInput:
    name: str
    front: FrontEnd
    repo: str
    data: str
    rounds_info: RoundsInfo
    rounds: Iterable[int]


def read_json(path: str) -> object:
    with open(path) as stream:
        loaded: object = json.load(stream)
        return loaded


def recorded_clusters(
    directory: str,
) -> tuple[dict[str, list[str]], dict[str, list[str]]]:
    recorded: collections.defaultdict[str, list[str]] = collections.defaultdict(list)
    capped: collections.defaultdict[str, list[str]] = collections.defaultdict(list)
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


def commit_case(front: FrontEnd, inputs: FrontInput, parent: str | None) -> CommitCase:
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


def interesting(case: CommitCase) -> bool:
    expected = case["expected"]
    return bool(
        case["affected"]
        or case["pilot"] != expected
        or expected["clusters"]
        or expected["inert"]
        or expected["resolved"]
        or set(expected["counts"]) - {"modified_owned", "deleted"}
    )


def build_round(job: RepoInput, round_number: int) -> RoundRecord:
    front = job.front
    info = next(x for x in job.rounds_info["rounds"] if x["round"] == round_number)
    directory = os.path.join(job.data, "rounds", str(round_number))
    membership = cast(Membership, read_json(os.path.join(directory, "membership.json")))
    own = front.Ownership(membership)
    recorded, capped = recorded_clusters(directory)
    state = front.TreeState(front.tree_paths(job.repo, info["after_commit"]))
    commits: list[CommitCase] = []
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
            cast(RoundsInfo, read_json(os.path.join(args.x2d, "prepass/rounds.json"))),
            range(1, 11),
        ),
        RepoInput(
            "axum",
            load("frontend_axum", os.path.join(axum_pilot, "frontend.py")),
            args.axum,
            os.path.join(args.x2d, "data/postrun-axum"),
            cast(
                RoundsInfo, read_json(os.path.join(axum_pilot, "prepass/rounds.json"))
            ),
            range(1, 6),
        ),
    ]
    for job in jobs:
        result: RepoFixture = {
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
