"""X2d model-free pre-pass (shape-decision section 11 K0/K1; PREREGISTRATION section 3): distributions only, no model, no label, no predictor-outcome relationship.

    .tools/venv/bin/python evaluation/x2d-merge-association-pilot/prepass.py --repo /tmp/scout-x2c-smoke/_src/fastapi --out evaluation/x2d-merge-association-pilot/prepass

Writes (under --out): commits.jsonl (one row per first-parent commit from PRE_START to the window end: rule counts and residual clusters), k0.json, rounds.json (the
round table), reserve.json (pre-C0 clusters reserved for the round-0 gates and the dry batch), summary.json.
"""
import argparse, collections, hashlib, json, os, re, sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import frontend as F

WINDOW_START = "2025-01-01T00:00:00Z"
PRE_START = "2023-01-01T00:00:00Z"  # start of the pre-C0 reserve search window
TIP = "origin/master"
EXPECT_TIP = "79cab5bed28b24d92c185d236feb33f5783dfeb3"
N_ROUNDS, MIN_CLUSTERS, MIN_PRS = 10, 20, 6
PR_CAP = 10  # at most this many clusters per PR are scored (hash-sampled); the rest are reported (shape-decision 1: one migration must not dominate ten rounds)
SRC_EXT = {".c", ".cc", ".cpp", ".cs", ".cxx", ".ex", ".exs", ".fs", ".fsi", ".go", ".h", ".heex", ".hpp", ".java", ".js", ".jsx", ".kt", ".kts", ".m", ".mm", ".php", ".py",
           ".rb", ".rs", ".bash", ".sh", ".swift", ".ts", ".tsx", ".vb", ".yaml", ".yml"}  # third_party/clawpatch/src/agent-mapper.ts:60-93
SKIP = re.compile(r"(^|/)(node_modules|dist|build|coverage|\.build|\.git|\.clawpatch|\.worktrees|\.turbo|\.next|\.vercel|\.venv(?:-[^/]+)?|venv|Pods|Carthage|SourcePackages|DerivedData|__pycache__)(/|$)")
TEST = re.compile(r"(^|/)(test|tests|__tests__)(/|$)|(?:^|[._-])(?:test|spec)\.[^/]+$", re.I)  # agent-mapper.ts:509-511


def inventory(paths):
    """Clawpatch repoInventory order and samples (agent-mapper.ts:363-400, git ls-files order = byte order of the path)."""
    files = [p for p in paths if p and not SKIP.search(p) and not p.startswith("target/") and p != "target" and not p.startswith(".build/")]
    src = [p for p in files if os.path.splitext(p)[1].lower() in SRC_EXT and not TEST.search(p)]
    tests = [p for p in files if TEST.search(p)]
    return files, src, tests


def k0(repo, c0):
    paths = sorted(F.tree_paths(repo, c0), key=lambda s: s.encode("utf-8", "surrogateescape"))
    files, src, tests = inventory(paths)
    top = lambda lst: collections.Counter(p.split("/")[0] for p in lst)
    out = {"c0": c0, "files": len(files), "source_files": len(src), "test_files": len(tests),
           "window_files": 400, "window_source": 500, "window_tests": 200,
           "source_500th": src[499] if len(src) >= 500 else None, "source_499th_501st": src[498:501],
           "source_by_top_dir_all": dict(top(src)), "source_by_top_dir_in_window": dict(top(src[:500])),
           "core_fastapi_source_total": sum(1 for p in src if p.startswith("fastapi/")),
           "core_fastapi_source_in_window": sum(1 for p in src[:500] if p.startswith("fastapi/")),
           "files_400th": files[399] if len(files) >= 400 else None, "files_by_top_dir_in_window": dict(top(files[:400])),
           "tests_200th": tests[199] if len(tests) >= 200 else None, "tests_in_window": len(tests[:200]),
           "first_source_dir_order": []}
    seen = []
    for p in src:
        d = p.split("/")[0]
        if d not in seen:
            seen.append(d)
    out["first_source_dir_order"] = seen
    return out


def replay(repo, start_parent, tip):
    commits = F.first_parent_commits(repo, start_parent, tip)
    state = F.TreeState(F.tree_paths(repo, start_parent))
    rows = []
    for sha, parent, date, subject in commits:
        entries = F.diff_entries(repo, parent, sha)
        counts, residual = F.classify(entries, state)
        blobs = {e[3]: e[4] for e in entries if e[0] in "ARC"}
        live, inert = F.clusters(residual, blobs)
        live_dir, _ = F.clusters(residual, blobs, dir_only=True)
        for c in live + inert:
            c["id"] = F.cluster_id(sha, c)
        survivors = None
        rows.append({"sha": sha, "date": date, "subject": subject, "pr": F.pr_number(subject), "entries": len(entries), "counts": dict(counts),
                     "clusters": live, "n_clusters": len(live), "n_inert": len(inert), "n_clusters_dir_only": len(live_dir),
                     "residual_files": len(residual)})
        state.apply(entries)
    return rows, state


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--repo", required=True)
    ap.add_argument("--out", required=True)
    a = ap.parse_args()
    repo, out = a.repo, a.out
    os.makedirs(out, exist_ok=True)
    tip = F.git(repo, "rev-parse", TIP).decode().strip()
    if tip != EXPECT_TIP:
        sys.exit("tip %s != %s" % (tip, EXPECT_TIP))
    c0 = F.git(repo, "rev-list", "--first-parent", "--before=" + WINDOW_START, "-1", TIP).decode().strip()
    pre0 = F.git(repo, "rev-list", "--first-parent", "--before=" + PRE_START, "-1", TIP).decode().strip()
    c0_meta = F.git(repo, "log", "-1", "--format=%H %cI %s", c0).decode().strip()
    json.dump(k0(repo, c0), open(os.path.join(out, "k0.json"), "w"), indent=1)

    rows, final_state = replay(repo, pre0, tip)
    c0_idx = next(i for i, r in enumerate(rows) if r["sha"] == c0)
    pre, win = rows[: c0_idx + 1], rows[c0_idx + 1:]
    with open(os.path.join(out, "commits.jsonl"), "w") as f:
        for i, r in enumerate(rows):
            f.write(json.dumps(dict(r, split="pre" if i <= c0_idx else "window")) + "\n")

    # ---- the round table over the window (chronological, eligible = commit with >= 1 live cluster; a PR counts min(clusters, PR_CAP) towards the round size)
    def selected(r):
        return sorted(r["clusters"], key=lambda c: hashlib.sha256(("x2d-cap\0" + c["id"]).encode()).hexdigest())[:PR_CAP]
    elig = [(i, r["pr"] if r["pr"] is not None else r["sha"][:8], min(r["n_clusters"], PR_CAP)) for i, r in enumerate(win) if r["n_clusters"] > 0]
    rounds, leftover, ok = F.cut_rounds(elig, N_ROUNDS, MIN_CLUSTERS, MIN_PRS)
    if ok and leftover:  # R1-02: the spare PRs after round 10 are folded into round 10 (larger frozen tail); nothing is left over
        rounds[-1] = rounds[-1] + leftover
        leftover = []
    table, prev, prev_idx = [], c0, -1
    for k, rd in enumerate(rounds, 1):
        lo, hi = rd[0][0], rd[-1][0]
        table.append({"round": k, "after_commit": prev, "through_commit": win[hi]["sha"], "first_eligible_date": win[lo]["date"], "last_eligible_date": win[hi]["date"],
                      "commits_in_range": hi - prev_idx,
                      "eligible_prs": len({x[1] for x in rd}), "scored_clusters": sum(x[2] for x in rd), "uncapped_clusters": sum(win[x[0]]["n_clusters"] for x in rd),
                      "max_pr_share_scored": round(max(x[2] for x in rd) / sum(x[2] for x in rd), 3),
                      "prs": [{"pr": x[1], "sha": win[x[0]]["sha"], "date": win[x[0]]["date"], "clusters": win[x[0]]["n_clusters"], "scored": x[2],
                               "scored_ids": [c["id"] for c in selected(win[x[0]])]} for x in rd]})
        prev, prev_idx = win[hi]["sha"], hi
    spare = [{"pr": x[1], "sha": win[x[0]]["sha"], "date": win[x[0]]["date"], "clusters": win[x[0]]["n_clusters"], "scored": x[2]} for x in leftover]
    json.dump({"ok": ok, "pr_cap": PR_CAP, "min_clusters": MIN_CLUSTERS, "min_prs": MIN_PRS, "c0": c0, "tip": tip, "rounds": table, "spare_after_round_10": spare,
               "frozen_8_10": {"scored_clusters": sum(t["scored_clusters"] for t in table[7:]), "prs": len({p["pr"] for t in table[7:] for p in t["prs"]})}},
              open(os.path.join(out, "rounds.json"), "w"), indent=1)

    # ---- the pre-C0 reserve: clusters of pre-C0 PRs whose files all still exist at C0 (replayed tree state at c0)
    st = F.TreeState(F.tree_paths(repo, pre0))
    c0_paths = set(F.tree_paths(repo, c0))
    cands = []
    for r in reversed(pre):
        for c in r["clusters"]:
            if all(p in c0_paths for p in c["files"]):
                cands.append({"id": c["id"], "sha": r["sha"], "pr": r["pr"], "date": r["date"], "subject": r["subject"], "dir": c["dir"], "stem": c["stem"],
                              "role": c["role"], "n_files": len(c["files"]), "files": c["files"][:6], "kinds": c["kinds"]})
    cands.sort(key=lambda c: hashlib.sha256(("x2d-reserve\0" + c["id"]).encode()).hexdigest())
    for i, c in enumerate(cands):  # D = K4-K6 dry batch (20), P = P1/N1 gate pool (30), G = guide pool (worked examples, never labelled as units)
        c["set"] = "D" if i < 20 else "P" if i < 50 else "G"
    json.dump({"c0": c0, "pre_start": pre0, "n_pre_commits": len(pre), "n_surviving_clusters": len(cands), "n_surviving_prs": len({c["sha"] for c in cands}),
               "by_role": dict(collections.Counter(c["role"] for c in cands)), "clusters_hash_order": cands}, open(os.path.join(out, "reserve.json"), "w"), indent=1)

    # ---- summary distributions (window)
    tot = collections.Counter()
    for r in win:
        tot.update(r["counts"])
    per_year = collections.defaultdict(lambda: [0, 0, 0])
    for r in win:
        y = r["date"][:4]
        per_year[y][0] += 1
        per_year[y][1] += 1 if r["n_clusters"] else 0
        per_year[y][2] += r["n_clusters"]
    byrole = collections.Counter()
    for r in win:
        for c in r["clusters"]:
            byrole[c["role"]] += 1
    nclu = sorted((r["n_clusters"] for r in win if r["n_clusters"]), reverse=True)
    summary = {"c0": c0_meta, "tip": tip, "window_commits": len(win), "window_eligible_prs": sum(1 for r in win if r["n_clusters"]),
               "window_clusters": sum(r["n_clusters"] for r in win), "window_clusters_dir_only": sum(r["n_clusters_dir_only"] for r in win),
               "window_inert_clusters": sum(r["n_inert"] for r in win), "window_residual_files": sum(r["residual_files"] for r in win),
               "per_year[commits,eligible_prs,clusters]": per_year, "rule_counts_window": dict(tot), "clusters_by_role": dict(byrole),
               "clusters_per_pr_desc_top15": nclu[:15], "rounds_ok": ok, "rounds_cut": len(rounds), "leftover_prs": len(leftover),
               "leftover_clusters": sum(x[2] for x in leftover)}
    json.dump(summary, open(os.path.join(out, "summary.json"), "w"), indent=1, default=dict)
    print(json.dumps(summary, indent=1, default=dict))


if __name__ == "__main__":
    main()
