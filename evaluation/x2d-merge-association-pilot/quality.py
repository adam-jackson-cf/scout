"""X2d onboarding quality report (PREREGISTRATION 4.3), computed once on the frozen catalogue. Label-free except the intruder audit; never touches a predictor-outcome association.

    .tools/venv/bin/python evaluation/x2d-merge-association-pilot/quality.py --repo CLONE [--data D] [--perms 200] [--since 2023-01-01] [--intruder-n 40]
    .tools/venv/bin/python evaluation/x2d-merge-association-pilot/quality.py --finish [--data D]

Main run writes (append-only; refuses when they exist) D/onboard/quality.json and D/onboard/intruder-items.jsonl (+ intruder-key.jsonl: the answers, never shown to labellers).
Items: {"case_id": "onboard:intruder-<n>", "item_id", "area_title", "title", "summary", "files": [6 paths, shuffled]} -- an area's title and summary, 5 of its files and 1 file
from another area (same scope when possible); seeded (x2d.SEED). Labelling is `label_x2d.py --round ... --kind intruder`, which writes D/onboard/labels.intruder.jsonl.
`--finish` reads those labels (per-session rows by labeller_A / labeller_B and final rows; answer in row["answer"]["intruder"], else row["intruder"], else row["primary"]),
and writes D/onboard/quality.intruder.json (new file, append-only): kappa A vs B, intruder-detection rate of A, B and of the consensus, unparsable counts.

Reported (quality.json): coverage by role x {heuristic-owned, area-owned, both, unowned} over the C0 inventory minus policy-(e) files (heuristic-owned = the file is owned by a
non-agent feature of some scoped draw; area-owned = in a catalogue area); area-owned share at C0 sorted source position > 500 (position in Clawpatch's sourceFiles list, as
prepass.py k0) and at inventory position > 500; areas at the 60-file cap; overlap share; co-change lift on first-parent commits (since --since, before C0) against a
size-preserving permutation of area labels over the owned files (seeded; commits touching > 30 owned files are skipped and counted; a multi-area file counts for its first
area id); leaf-directory label entropy; granularity (area count, files per area median/IQR/max/singletons, exp(H)); reproducibility between the K draws of each scope (owned-file
Jaccard, nestedness both ways, VI split H(X|Y)/H(Y|X) over files owned by exactly one feature in both); split-half consensus ARI: consensus of draws {1,2} vs draw {3}
(and the rotations) against the median single-draw ARI. The consensus is the GO consensus: `go run ./evaluation/x2d-merge-association-pilot/onboard -consensus-only` on a temp
copy of the draws restricted to the chosen draws (no Python port: one implementation); when the Go run fails the ARI figures are `unmeasured` with the error. Core gate: core
`fastapi/` source files area-owned < 50 % -> core_gate true. Every figure carries its `unmeasured` count.
Every CLI call runs x2d.verify_pins(); X2D_SKIP_SIGNOFF=1 -> verify_pins(require_signoff=False) (validation only).
"""
import argparse
import collections
import glob
import json
import math
import os
import random
import shutil
import statistics
import subprocess
import sys
import tempfile

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import x2d  # noqa: E402
import common as C  # noqa: E402
import frontend as F  # noqa: E402
import prepass as P  # noqa: E402
import analysis_x2d as A  # noqa: E402

AREA_CAP = 60
CO_CHANGE_MAX_OWNED = 30
ROLES = ["core", "example", "test", "docs-config", "infra", "config"]
STATES = ["heuristic-owned", "area-owned", "both", "unowned"]


# ------------------------------------------------------------------ small maths
def entropy(counts):
    n = sum(counts)
    return -sum(c / n * math.log(c / n) for c in counts if c) if n else 0.0


def ari(a, b):
    """Adjusted Rand index of two label lists over the same items (1.0 when both partitions are trivial and equal)."""
    n = len(a)
    if n < 2:
        return None
    tab = collections.Counter(zip(a, b))
    ra, rb = collections.Counter(a), collections.Counter(b)
    comb = lambda x: x * (x - 1) / 2
    s_ij = sum(comb(v) for v in tab.values())
    s_a, s_b = sum(comb(v) for v in ra.values()), sum(comb(v) for v in rb.values())
    expected = s_a * s_b / comb(n)
    mx = (s_a + s_b) / 2
    if mx == expected:
        return 1.0
    return (s_ij - expected) / (mx - expected)


def vi_split(a, b):
    """(H(X|Y), H(Y|X)) in nats."""
    joint = collections.Counter(zip(a, b))
    hxy = entropy(list(joint.values()))
    return hxy - entropy(list(collections.Counter(b).values())), hxy - entropy(list(collections.Counter(a).values()))


def quartiles(v):
    if not v:
        return None
    s = sorted(v)
    return [A._quantile(s, 0.25), A._quantile(s, 0.5), A._quantile(s, 0.75)]


# ------------------------------------------------------------------ draws
def load_draw(data, scope, k):
    """-> {"areas": {feature id: [files]}, "heuristic_files": set, "meta": dict} or None (draw absent or failed: unmeasured)."""
    base = os.path.join(data, "onboard", "draws", scope, str(k))
    mp = os.path.join(base, "meta.json")
    if not os.path.exists(mp):
        return None
    with open(mp) as f:
        meta = json.load(f)
    if meta.get("rc") not in (0, None):
        return None
    areas, heur = {}, set()
    for p in sorted(glob.glob(os.path.join(base, "features", "*.json"))):
        with open(p) as f:
            rec = json.load(f)
        if rec.get("status") == C.STALE_STATUS:
            continue
        files = C.owned_files(rec)
        if rec.get("source") == C.AGENT_SOURCE:
            areas[rec.get("id") or os.path.basename(p)] = files
        else:
            heur.update(files)
    return {"areas": areas, "heuristic_files": heur, "meta": meta}


def single_partition(draw):
    """{file: feature id} for files owned by exactly one agent feature of the draw."""
    owners = collections.defaultdict(list)
    for fid, files in draw["areas"].items():
        for p in files:
            owners[p].append(fid)
    return {p: v[0] for p, v in owners.items() if len(v) == 1}


def owned_set(draw):
    return {p for files in draw["areas"].values() for p in files}


def pair_stats(x, y):
    """Reproducibility between two draws of one scope."""
    ox, oy = owned_set(x), owned_set(y)
    px, py = single_partition(x), single_partition(y)
    common = sorted(set(px) & set(py))
    ax, ay = [px[p] for p in common], [py[p] for p in common]
    sx, sy = [set(v) for v in x["areas"].values() if v], [set(v) for v in y["areas"].values() if v]
    nest = lambda s1, s2: (sum(1 for a in s1 if any(a <= b for b in s2)) / len(s1)) if s1 else None
    hxy, hyx = vi_split(ax, ay) if common else (None, None)
    return {"jaccard_owned": (len(ox & oy) / len(ox | oy)) if (ox | oy) else None, "nested_x_in_y": nest(sx, sy), "nested_y_in_x": nest(sy, sx),
            "H_x_given_y": hxy, "H_y_given_x": hyx, "n_common_single_owned": len(common), "ari": ari(ax, ay) if len(common) >= 2 else None}


def consensus_via_go(data, ks, tmp):
    """Run the Go consensus on a temp copy of the draws restricted to `ks` (renumbered 1..len). -> {scope: {file: area_id}} or raises RuntimeError."""
    scopes_json = os.path.join(x2d.HERE, "prepass", "scopes.json")
    ids = [s["id"] for s in json.load(open(scopes_json))["scopes"]]
    dr = os.path.join(tmp, "draws")
    for sc in ids:
        for new, old in enumerate(ks, 1):
            src = os.path.join(data, "onboard", "draws", sc, str(old))
            if os.path.isdir(src):
                shutil.copytree(src, os.path.join(dr, sc, str(new)))
    out = os.path.join(tmp, "out")
    env = {k: v for k, v in os.environ.items() if k != "OPENROUTER_API_KEY"}
    cp = subprocess.run(["go", "run", "./evaluation/x2d-merge-association-pilot/onboard", "-consensus-only", "-draws", dr, "-scopes", scopes_json, "-out", out],
                        cwd=x2d.ROOT, env=env, capture_output=True, text=True, timeout=900)
    if cp.returncode != 0 or not os.path.exists(os.path.join(out, "catalogue.json")):
        raise RuntimeError("go consensus failed rc=%s: %s" % (cp.returncode, (cp.stderr or cp.stdout)[-400:]))
    with open(os.path.join(out, "catalogue.json")) as f:
        cat = json.load(f)
    res = collections.defaultdict(dict)
    for a in cat["areas"]:
        for p in a["files"]:
            res[a["scope"]].setdefault(p, []).append(a["area_id"])
    return {sc: {p: v[0] for p, v in m.items() if len(v) == 1} for sc, m in res.items()}


# ------------------------------------------------------------------ report sections
def coverage(elig, role_of, area_files, heur_files):
    cnt = {r: {s: 0 for s in STATES} for r in ROLES}
    for p in elig:
        h, a = p in heur_files, p in area_files
        s = "both" if (h and a) else ("area-owned" if a else ("heuristic-owned" if h else "unowned"))
        cnt[role_of(p)][s] += 1
    return cnt


def cochange(repo, since, owned_primary, perms, seed):
    out = subprocess.run(["git", "-C", repo, "log", "--first-parent", "-m", "--name-only", "--format=@@%H", "--since=" + since, x2d.C0], capture_output=True, text=True, check=True).stdout
    commits, cur = [], None
    for line in out.splitlines():
        if line.startswith("@@"):
            cur = []
            commits.append(cur)
        elif line and cur is not None:
            cur.append(line)
    files = sorted(owned_primary)
    index = {p: i for i, p in enumerate(files)}
    pairs = collections.Counter()
    skipped_large = used = 0
    for fl in commits:
        idx = sorted({index[p] for p in fl if p in index})
        if len(idx) < 2:
            continue
        if len(idx) > CO_CHANGE_MAX_OWNED:
            skipped_large += 1
            continue
        used += 1
        for i in range(len(idx)):
            for j in range(i + 1, len(idx)):
                pairs[(idx[i], idx[j])] += 1
    total = sum(pairs.values())
    if total == 0:
        return {"status": "unmeasured", "reason": "no first-parent commit before C0 co-changed two owned files", "commits_total": len(commits)}
    labels = [owned_primary[p] for p in files]
    obs_same = sum(w for (i, j), w in pairs.items() if labels[i] == labels[j])
    plist = list(pairs.items())
    rng = random.Random("x2d-cochange|%s" % seed)
    null = []
    for _ in range(perms):
        lab = labels[:]
        rng.shuffle(lab)
        null.append(sum(w for (i, j), w in plist if lab[i] == lab[j]) / total)
    mean_null = sum(null) / len(null)
    return {"status": "measured", "commits_total": len(commits), "commits_used": used, "commits_skipped_gt_%d_owned" % CO_CHANGE_MAX_OWNED: skipped_large, "pairs_weighted": total,
            "unique_pairs": len(pairs), "p_same_area_observed": obs_same / total, "p_same_area_null_mean": mean_null, "null_sd": statistics.pstdev(null) if len(null) > 1 else 0.0,
            "lift": (obs_same / total) / mean_null if mean_null else None, "perms": perms, "p_value_null_ge_observed": (1 + sum(1 for v in null if v >= obs_same / total)) / (len(null) + 1),
            "since": since, "multi_area_files_counted_once": True}


def leaf_entropy(owned_primary):
    by_dir = collections.defaultdict(list)
    by_area = collections.defaultdict(list)
    for p, a in owned_primary.items():
        d = os.path.dirname(p)
        by_dir[d].append(a)
        by_area[a].append(d)
    n = len(owned_primary)
    h_area_given_dir = sum(len(v) / n * entropy(list(collections.Counter(v).values())) for v in by_dir.values()) if n else None
    h_dir_given_area = sum(len(v) / n * entropy(list(collections.Counter(v).values())) for v in by_area.values()) if n else None
    return {"H_area_given_leaf_dir_nats": h_area_given_dir, "H_leaf_dir_given_area_nats": h_dir_given_area, "leaf_dirs": len(by_dir), "files": n}


def granularity(cat):
    sizes = [len(a["files"]) for a in cat["areas"]]
    if not sizes:
        return {"areas": 0}
    return {"areas": len(sizes), "files_per_area": {"median": statistics.median(sizes), "quartiles": quartiles(sizes), "max": max(sizes), "singletons": sum(1 for s in sizes if s == 1)},
            "exp_H_area_sizes": math.exp(entropy(sizes))}


# ------------------------------------------------------------------ intruder audit
def build_intruder_items(cat, n_items, seed):
    rng = random.Random("x2d-intruder|%s" % seed)
    areas = [a for a in cat["areas"] if len(a["files"]) >= 5]
    if not areas:
        raise SystemExit("no catalogue area has >= 5 files: no intruder item can be built")
    owners = collections.defaultdict(set)
    for a in cat["areas"]:
        for p in a["files"]:
            owners[p].add(a["area_id"])
    order = []
    while len(order) < n_items:
        batch = areas[:]
        rng.shuffle(batch)
        order.extend(batch)
    items, keys = [], []
    for n, a in enumerate(order[:n_items], 1):
        mine = rng.sample(a["files"], 5)
        pool = sorted(p for b in cat["areas"] if b["area_id"] != a["area_id"] and b["scope"] == a["scope"] for p in b["files"] if owners[p] == {b["area_id"]})
        if not pool:
            pool = sorted(p for b in cat["areas"] if b["area_id"] != a["area_id"] for p in b["files"] if owners[p] == {b["area_id"]})
        if not pool:
            raise SystemExit("no file owned by another area exists: no intruder item can be built")
        intr = rng.choice(pool)
        files = mine + [intr]
        rng.shuffle(files)
        cid = "onboard:intruder-%d" % n
        items.append({"case_id": cid, "item_id": n, "area_title": a["title"], "title": a["title"], "summary": a["summary"], "files": files})
        keys.append({"case_id": cid, "area_id": a["area_id"], "intruder": intr})
    return items, keys


def _intruder_answer(row):
    ans = row.get("answer")
    if isinstance(ans, dict) and "intruder" in ans:
        return ans["intruder"]
    if "intruder" in row:
        return row["intruder"]
    return row.get("primary")


def finish(data):
    out = os.path.join(data, "onboard", "quality.intruder.json")
    x2d.refuse_if_exists(out)
    lp = os.path.join(data, "onboard", "labels.intruder.jsonl")
    if not os.path.exists(lp):
        sys.exit("missing input: %s (label the items with label_x2d.py --kind intruder first)" % lp)
    key = {k["case_id"]: k for k in x2d.read_jsonl(os.path.join(data, "onboard", "intruder-key.jsonl"))}
    if not key:
        sys.exit("missing input: intruder-key.jsonl")
    sess = collections.defaultdict(dict)
    for r in x2d.read_jsonl(lp):
        if r.get("kind") != "intruder" or r.get("case_id") not in key:
            continue
        by = r.get("by") or ""
        slot = "final" if by in ("A+B agree", "C adjudicated") else ("A" if by.startswith("labeller_A") else "B" if by.startswith("labeller_B") else None)
        if slot:
            sess[r["case_id"]][slot] = r
    both = [c for c in key if "A" in sess[c] and "B" in sess[c] and _intruder_answer(sess[c]["A"]) and _intruder_answer(sess[c]["B"])]
    det = lambda slot: [(_intruder_answer(sess[c][slot]) == key[c]["intruder"]) for c in key if slot in sess[c] and _intruder_answer(sess[c][slot])]
    rate = lambda v: {"k": sum(v), "n": len(v), "rate": (sum(v) / len(v)) if v else None, "wilson95": [None if math.isnan(x) else x for x in A.wilson(sum(v), len(v))] if v else None}
    res = {"items": len(key), "items_with_A_and_B": len(both), "unmeasured": len(key) - len(both),
           "kappa_A_vs_B": A.cohen_kappa([_intruder_answer(sess[c]["A"]) for c in both], [_intruder_answer(sess[c]["B"]) for c in both]),
           "detection_A": rate(det("A")), "detection_B": rate(det("B")), "detection_consensus": rate(det("final")), "provenance": "agent_reviewed", "labels_sha256": x2d.sha256_file(lp)}
    x2d.write_json_new(out, res)
    return res


# ------------------------------------------------------------------ main run
def run(repo, data, perms=200, since="2023-01-01", intruder_n=40):
    cat = x2d.read_catalogue(data)
    paths = sorted(F.tree_paths(repo, x2d.C0), key=lambda s: s.encode("utf-8", "surrogateescape"))
    files, src, tests = P.inventory(paths)
    elig = [p for p in files if F.excluded_reason(p) is None]
    area_files = {p for a in cat["areas"] for p in a["files"]}
    scopes = [s["id"] for s in json.load(open(os.path.join(x2d.HERE, "prepass", "scopes.json")))["scopes"]]
    draws, missing = {}, []
    for sc in scopes:
        for k in (1, 2, 3):
            d = load_draw(data, sc, k)
            if d is None:
                missing.append("%s/%d" % (sc, k))
            else:
                draws[(sc, k)] = d
    heur = set()
    for d in draws.values():
        heur |= d["heuristic_files"]
    res = {"schema": "x2d-quality-v1", "c0": x2d.C0, "catalogue_sha256": x2d.catalogue_digest(data),
           "unmeasured": {"draws_missing_or_failed": missing, "inventory_files_excluded_by_policy_e": len(files) - len(elig)}}
    res["inventory"] = {"files": len(files), "eligible_files": len(elig), "source_files": len(src), "test_files": len(tests)}
    res["coverage_by_role"] = coverage(elig, F.role, area_files, heur)
    srcset = set(src)
    pos_src = {p: i + 1 for i, p in enumerate(src)}
    pos_inv = {p: i + 1 for i, p in enumerate(files)}
    for name, pos, universe in (("by_source_position", pos_src, [p for p in src if F.excluded_reason(p) is None]), ("by_inventory_position", pos_inv, elig)):
        late = [p for p in universe if pos[p] > 500]
        early = [p for p in universe if pos[p] <= 500]
        res.setdefault("area_owned_share_position_gt_500", {})[name] = {"late": {"files": len(late), "area_owned": sum(1 for p in late if p in area_files)},
                                                                       "early": {"files": len(early), "area_owned": sum(1 for p in early if p in area_files)}}
    for v in res["area_owned_share_position_gt_500"].values():
        for part in ("late", "early"):
            v[part]["share"] = (v[part]["area_owned"] / v[part]["files"]) if v[part]["files"] else None
    res["areas_at_cap"] = {"cap": AREA_CAP, "at_or_over_cap": sum(1 for a in cat["areas"] if len(a["files"]) >= AREA_CAP), "areas": len(cat["areas"])}
    owners = collections.defaultdict(list)
    for a in cat["areas"]:
        for p in a["files"]:
            owners[p].append(a["area_id"])
    res["overlap"] = {"files_in_ge2_areas": sum(1 for v in owners.values() if len(v) >= 2), "owned_files": len(owners), "share": (sum(1 for v in owners.values() if len(v) >= 2) / len(owners)) if owners else None}
    primary = {p: sorted(v)[0] for p, v in owners.items()}
    res["co_change"] = cochange(repo, since, primary, perms, x2d.SEED)
    res["leaf_directory_entropy"] = leaf_entropy(primary)
    res["granularity"] = granularity(cat)
    # reproducibility between draws
    rep = {}
    for sc in scopes:
        pr = {}
        for a_, b_ in ((1, 2), (1, 3), (2, 3)):
            if (sc, a_) in draws and (sc, b_) in draws:
                pr["%d-%d" % (a_, b_)] = pair_stats(draws[(sc, a_)], draws[(sc, b_)])
        rep[sc] = pr
    flat = [v for pr in rep.values() for v in pr.values()]
    med = lambda key: statistics.median([v[key] for v in flat if v[key] is not None]) if any(v[key] is not None for v in flat) else None
    res["reproducibility"] = {"per_scope": rep, "median_over_scope_pairs": {k: med(k) for k in ("jaccard_owned", "nested_x_in_y", "nested_y_in_x", "H_x_given_y", "H_y_given_x", "ari")}}
    # split-half consensus ARI via the Go consensus
    sh = {"method": "go consensus-only on a temp copy of draws restricted to the chosen ones; ARI over files in a single area of the consensus and in a single feature of the held-out draw"}
    rots = []
    for held in (3, 2, 1):
        ks = [k for k in (1, 2, 3) if k != held]
        try:
            with tempfile.TemporaryDirectory(prefix="x2d-quality-") as tmp:
                cons = consensus_via_go(data, ks, tmp)
        except (RuntimeError, OSError, subprocess.TimeoutExpired) as e:
            rots.append({"held_out": held, "draws": ks, "status": "unmeasured", "reason": str(e)})
            continue
        pooled_a, pooled_b, per = [], [], {}
        for sc in scopes:
            if (sc, held) not in draws:
                continue
            part = single_partition(draws[(sc, held)])
            common = sorted(set(part) & set(cons.get(sc, {})))
            la, lb = [cons[sc][p] for p in common], [part[p] for p in common]
            per[sc] = {"n": len(common), "ari": ari(la, lb) if len(common) >= 2 else None}
            pooled_a += [sc + "/" + x for x in la]
            pooled_b += [sc + "/" + x for x in lb]
        rots.append({"held_out": held, "draws": ks, "status": "measured", "per_scope": per, "pooled_ari": ari(pooled_a, pooled_b), "pooled_n": len(pooled_a)})
    sh["rotations"] = rots
    primary_rot = next((r for r in rots if r["held_out"] == 3), None)
    singles = []
    for a_, b_ in ((1, 2), (1, 3), (2, 3)):
        pa, pb = [], []
        for sc in scopes:
            if (sc, a_) in draws and (sc, b_) in draws:
                x, y = single_partition(draws[(sc, a_)]), single_partition(draws[(sc, b_)])
                common = sorted(set(x) & set(y))
                pa += [sc + "/" + x[p] for p in common]
                pb += [sc + "/" + y[p] for p in common]
        singles.append(ari(pa, pb) if len(pa) >= 2 else None)
    sh["single_draw_pooled_ari_pairs"] = singles
    sh["single_draw_median_ari"] = statistics.median([s for s in singles if s is not None]) if any(s is not None for s in singles) else None
    sh["consensus_12_vs_3_pooled_ari"] = primary_rot.get("pooled_ari") if primary_rot and primary_rot["status"] == "measured" else None
    sh["consensus_exceeds_single_draw_median"] = (sh["consensus_12_vs_3_pooled_ari"] > sh["single_draw_median_ari"]) if sh["consensus_12_vs_3_pooled_ari"] is not None and sh["single_draw_median_ari"] is not None else None
    res["split_half_consensus_ari"] = sh
    # core gate
    core_src = [p for p in src if p.startswith("fastapi/")]
    owned_core = [p for p in core_src if p in area_files]
    res["core_gate"] = {"core_source_files": len(core_src), "area_owned": len(owned_core), "share": (len(owned_core) / len(core_src)) if core_src else None,
                        "core_gate": bool(core_src and len(owned_core) / len(core_src) < 0.5), "rule": "core fastapi/ source files area-owned < 50 % -> core units are stratum 3 and the go/no-go excludes them"}
    items, keys = build_intruder_items(cat, intruder_n, x2d.SEED)
    res["intruder_audit"] = {"status": "items written; labels pending (label_x2d.py --kind intruder, then quality.py --finish)", "items": len(items)}
    return res, items, keys


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--repo")
    ap.add_argument("--data", default=x2d.DATA)
    ap.add_argument("--perms", type=int, default=200)
    ap.add_argument("--since", default="2023-01-01")
    ap.add_argument("--intruder-n", type=int, default=40)
    ap.add_argument("--finish", action="store_true")
    a = ap.parse_args(argv)
    bad = x2d.verify_pins(require_signoff=x2d.require_signoff())
    if bad:
        sys.exit("pin verification failed:\n  " + "\n  ".join(bad))
    if a.finish:
        print(json.dumps(finish(a.data), indent=1, sort_keys=True))
        return 0
    if not a.repo:
        ap.error("--repo is required")
    if a.intruder_n < 30:
        ap.error("--intruder-n must be >= 30")
    qp = os.path.join(a.data, "onboard", "quality.json")
    ip = os.path.join(a.data, "onboard", "intruder-items.jsonl")
    kp = os.path.join(a.data, "onboard", "intruder-key.jsonl")
    x2d.refuse_if_exists(qp, ip, kp)
    res, items, keys = run(a.repo, a.data, a.perms, a.since, a.intruder_n)
    x2d.write_jsonl_new(ip, items)
    x2d.write_jsonl_new(kp, keys)
    x2d.write_json_new(qp, res)
    print("wrote %s (%d intruder items); core_gate=%s" % (qp, len(items), res["core_gate"]["core_gate"]))
    return 0


if __name__ == "__main__":
    sys.exit(main())
