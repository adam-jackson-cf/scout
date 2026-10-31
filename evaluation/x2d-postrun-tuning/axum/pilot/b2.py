"""X2d B2 mirror rule (PREREGISTRATION 6; frozen text, written before any label).

B2 as a baseline (baselines.py:b2) = B0 (majority owner of the owned siblings in the directory); if B0 does not assign (none or tie), THIS mirror rule; if the mirror
yields nothing, B1 (parent-directory walk) (R1-29).

The rule is the fastapi tutorial test mirror only: `tests/test_tutorial/test_X/ <-> docs_src/X/`. The topic X is the normalised name (variant suffixes stripped with
`frontend.norm_dir`: `test_x_py39` -> `test_x`, `x_an` -> `x`). A member in `tests/test_tutorial/test_X/...` takes the majority owner of the owned files under
`docs_src/X/...` and vice versa; owned files count once per owning area, the winner must be the unique maximum (a tie -> None). The pre-registration contains no
other mirror (in particular no `tests/test_<name>.py <-> fastapi/<name>.py` rule), so none is implemented. Deterministic; no randomness.

    b2_rule(unit_members, tree_paths, ownership) -> area_id | None
"""
import collections

import frontend as F

_cache = {"tree": None, "idx": None}


def _topic_test(path):
    parts = path.split("/")
    if len(parts) >= 4 and parts[0] == "tests" and parts[1] == "test_tutorial":
        n = F.norm_dir(parts[2])
        if n.startswith("test_") and len(n) > 5:
            return n[5:]
    return None


def _topic_docs(path):
    parts = path.split("/")
    if len(parts) >= 3 and parts[0] == "docs_src":
        return F.norm_dir(parts[1])
    return None


def _index(tree_paths):
    if _cache["tree"] is tree_paths and _cache.get("n") == len(tree_paths):
        return _cache["idx"]
    idx = collections.defaultdict(list)
    for p in sorted(tree_paths):
        t = _topic_test(p)
        if t:
            idx[("test", t)].append(p)
        d = _topic_docs(p)
        if d:
            idx[("docs", d)].append(p)
    _cache["tree"], _cache["idx"], _cache["n"] = tree_paths, idx, len(tree_paths)
    return idx


def b2_rule(unit_members, tree_paths, ownership):
    idx = _index(tree_paths)
    mem = set(unit_members)
    votes = collections.Counter()
    seen = set()
    for m in sorted(mem):
        t, d = _topic_test(m), _topic_docs(m)
        key = ("docs", t) if t else ("test", d) if d else None
        if key is None or key in seen:
            continue
        seen.add(key)
        for f in idx.get(key, []):
            if f in mem:
                continue
            for a in ownership.owners(f):
                votes[a] += 1
    if not votes:
        return None
    top = votes.most_common()
    if len(top) > 1 and top[0][1] == top[1][1]:
        return None
    return top[0][0]
