"""X2d unit builder (README Modules: build_units.py; PREREGISTRATION 3, 7, 8, 10).

CLI:
    build_units.py --repo CLONE --round R [--data D]
        Walks the first-parent commits after_commit..through_commit of prepass/rounds.json round R, classifies each commit against the ROUND-START membership
        (D/rounds/R/membership.json: round 1 = the catalogue, later rounds = gold.next_membership(R-1); an existing membership.json is verified against that, never
        rebuilt), clusters the residual (frontend.clusters), caps each PR at x2d.PR_CAP clusters (smallest sha256("x2d-cap\\0"+id)) and writes
        D/rounds/R/{units,capped,lookups}.jsonl + counts.json (the round's total rule counts; lookups hold a row for every commit with a non-zero count, R2-07) plus inherit.jsonl (the deterministic rule effects: deletions, (c) moves incl. pooled renames, (d) additions; read by gold.py)
        and membership.json. Rows carry `seq` (index of the commit within the round; inherit rows too) for chronological absorption. The membership-free pre-pass
        classification of every PR in rounds.json is recomputed and its capped cluster ids must equal `scored_ids` (loud refusal otherwise). Up to 10 rule-resolved (c)/(d)
        cluster-like items per round (smallest sha256("x2d-audit\\0"+id)) are also written as units with set "audit" (case_id "<r>:aud-<n>"), listed in lookups.jsonl.
        Per-round controls (R1-07), `control: true`, `control_kind`, `source_case_id`, set "ctl", case_id "<r>:ctl-<n>" (hash order), never metrics, never absorbed:
        `dup` = ceil(10 %) hash-chosen clones of window units (test-retest), `trap` = one empty-file rename across directories (gold: not covered, rule R10),
        `prose` = one clone of a window unit whose PR title/body talk about another area (same members: the answer must not move).
        The core gate (PREREGISTRATION 4.3) is decided ONCE at round 0 (D/rounds/0/gates.json core_gate.enabled, see `load_core_gate`; absent = off) and read here.
    build_units.py --repo CLONE --round R --tag contrast      (every round 1..10)
        D/rounds/R/units.contrast.jsonl = the round's scored window units re-stratified against the contrast catalogue (B2 against it runs every round; scored only where
        contrast labels exist). Rounds 1 and 10 also write D/rounds/R/units.contrast_window.jsonl = <= 30 hash-sampled (sha256("x2d-contrast\\0"+case_id)) of them: the
        units labelled against the contrast catalogue. Rounds 1 and 10 also append the 20 D units (round 0's units.contrast.jsonl rows `set: "D"`, re-stratified, ids unchanged; R2-06).
    build_units.py --repo CLONE --gates [--data D]
        D/rounds/0/k23.json {k2_share, k2_b2_share, k3_share, new_dir, core_gate, per_round, counts, real_none_proxy, frozen_pr_counts}: for rounds 1..10 against the ROUND-1 (catalogue) membership
        (`real_none_proxy` = sum over rounds 8..10 of k3_stratum3 + new_dir, core_gate excluded; `frozen_pr_counts` = {"8"|"9"|"10": {pr_key: scored clusters}}, R2-11)
        WITHOUT teacher forcing, no labels or outcomes. K2 (lower bound) = share of scored clusters on which B0 assigns (a unique modal owner among owned siblings);
        k2_b2_share = the same for B2. K3 (upper bound) = stratum-3 clusters whose normalised parent directory exists in the C0 tree and whose reason is not core_gate;
        the other stratum-3 clusters are counted as `new_dir`. Stratum 4 is not built here: it is assigned post hoc from gold `none_kind = new` (descriptive only).
Importable: build_round_units, make_unit, stratify, build_reserve_units, contrast_units, compute_core_gate, load_core_gate, prose_title, read_text, commit_message.
Append-only: refuses when an output exists. X2D_SKIP_SIGNOFF=1 -> verify_pins(require_signoff=False) (validation).
Rule counts in lookups: (a) modified_owned / modified_unowned (NOT units: coverage figure split `modified_unowned:pooled|never_owned`), (b) deleted,
(c) renamed_inherit / renamed_pooled (+ renamed_unowned/cross_scope/empty_trap residual), (d) variant_inherit (variant_unowned = sibling exists but is not owned),
(e) excluded:<reason>, (f) heuristic_only, (g) inert_clusters.
Rule effects use the round-start membership only (a file added by (d) in a round is not an owner within that round).
"""
import argparse, collections, copy, hashlib, json, math, os, re, sys

import x2d
import frontend as F
import gold as G

_SYM = re.compile(r"^\s*(?:async\s+def|def|class|function)\s+([A-Za-z_$][\w$]*)")
MAX_BLOB = 2_000_000
CONTRAST_WINDOW_MAX = 30  # rounds 1 and 10: at most this many window units are labelled against the contrast catalogue (R1-13)
FROZEN_ROUNDS = (8, 9, 10)  # K3 real-none proxy and the per-PR cluster counts (R2-11) are taken over these rounds
STATUS_WORD = {"A": "added", "M": "modified", "T": "modified", "D": "deleted", "R": "renamed", "C": "copied"}


# ---- git helpers
def read_text(repo, rev, path):
    """Text of `path` at `rev`; binary/oversized safe. -> str"""
    data = F.git(repo, "show", "%s:%s" % (rev, path), check=False)
    if len(data) > MAX_BLOB:
        data = data[:MAX_BLOB]
    if b"\0" in data[:8192]:
        return "[binary file: %d bytes]" % len(data)
    return data.decode("utf-8", "replace").replace("\r\n", "\n")


def commit_message(repo, sha):
    """-> (subject, body) from `git show -s --format=%B`."""
    msg = F.git(repo, "show", "-s", "--format=%B", sha).decode("utf-8", "replace").strip("\n")
    subject, _, body = msg.partition("\n")
    return subject.strip(), body.strip("\n")


def numstat(repo, parent, sha):
    raw = F.git(repo, "diff-tree", "-r", "-z", "--numstat", "-M%d%%" % F.RENAME_MIN_SIMILARITY, parent, sha).decode("utf-8", "surrogateescape").split("\0")
    files = add = dele = 0
    i = 0
    while i < len(raw):
        parts = raw[i].split("\t", 2)
        if len(parts) < 3:
            i += 1
            continue
        files += 1
        add += int(parts[0]) if parts[0].isdigit() else 0
        dele += int(parts[1]) if parts[1].isdigit() else 0
        i += 3 if parts[2] == "" else 1
    return {"files": files, "added_lines": add, "deleted_lines": dele}


def symbols_from_lines(lines):
    out = []
    for ln in lines:
        m = _SYM.match(ln)
        if m and m.group(1) not in out:
            out.append(m.group(1))
        if len(out) >= 8:
            break
    return out


def top_symbols_git(repo, parent, sha, path):
    d = F.git(repo, "diff", "-U0", "--no-color", parent, sha, "--", path, check=False).decode("utf-8", "replace")
    return symbols_from_lines([ln[1:] for ln in d.splitlines() if ln.startswith("+") and not ln.startswith("+++")])


# ---- stratum / siblings
def compute_core_gate(ownership, tree_paths):
    """Gate of PREREGISTRATION 4.3: core fastapi/ source files < 50 % area-owned -> core units are stratum 3 (reason core_gate)."""
    core = [p for p in tree_paths if p.startswith("fastapi/") and p.endswith(".py")]
    if not core:
        return False
    return sum(1 for p in core if ownership.owners(p)) / len(core) < 0.5


def load_core_gate(data=None):
    """The core gate as decided once at round 0 (D/rounds/0/gates.json core_gate.enabled); the file absent = False. Never recomputed per round (R1-14)."""
    p = os.path.join(x2d.round_dir(0, data), "gates.json")
    if not os.path.exists(p):
        return False
    with open(p) as f:
        return bool((json.load(f).get("core_gate") or {}).get("enabled"))


def prose_title(cat, avoid):
    """-> (decoy area id, title, body): PR prose that talks about an area other than `avoid` (the misleading-prose control)."""
    areas = [a for a in cat["areas"] if a["area_id"] != avoid] or cat["areas"]
    a = areas[int(_h("x2d-prose", avoid or ""), 16) % len(areas)]
    t = a["title"].strip().rstrip(".")
    summ = a["summary"].strip().split(". ")[0].strip()[:200]
    return a["area_id"], "🐛 Fix %s handling" % (t[:1].lower() + t[1:])[:60], "This change updates how %s works. %s." % ((t[:1].lower() + t[1:])[:80], summ.rstrip("."))


def _h(prefix, key):
    return hashlib.sha256((prefix + "\0" + key).encode()).hexdigest()


def _interleave(sibs, k=10):
    by = collections.OrderedDict()
    for s in sibs:
        by.setdefault(s["area_id"], []).append(s)
    out, i = [], 0
    while len(out) < k and any(i < len(v) for v in by.values()):
        for v in by.values():
            if i < len(v) and len(out) < k:
                out.append(v[i])
        i += 1
    return out


def stratify(members, role, tree_set, ownership, core_gate=False):
    """-> (stratum, reason, siblings<=10): owned siblings in the same normalised directory in `tree_set` (the parent tree)."""
    mem = set(members)
    sibs, seen = [], set()
    for m in members:
        for s in ownership.sibling_owners(m, tree_set):
            k = (s["path"], s["area_id"])
            if s["path"] in mem or k in seen:
                continue
            seen.add(k)
            sibs.append(s)
    sibs.sort(key=lambda s: (s["area_id"], s["path"]))
    areas = {s["area_id"] for s in sibs}
    heur = any(ownership.is_heuristic_only(m) for m in members)
    if role == "core" and core_gate:
        return 3, "core_gate", _interleave(sibs)
    if not areas:
        return 3, "heuristic_only" if heur else "no_owned_siblings", []
    if len(areas) == 1:
        return 1, "one_area", _interleave(sibs)
    return 2, "several_areas", _interleave(sibs)


# ---- the unit row
def make_unit(repo, r, case_id, cluster_id, commit, parent, date, subject, body, cluster, member_info, rename_old, changed_paths, diff_summary, tree_set, ownership,
              settings, set_name, core_gate=False, control=False, reader=None, symbols_from_text=False):
    """One units.jsonl row (contract). `cluster`: {"dir","stem","files","role","scope"}; member_info {path: (kind, detail)}; rename_old {new: old};
    reader(path) -> text overrides reading blobs at `commit` (controls)."""
    members = sorted(cluster["files"])
    rd = reader or (lambda p: read_text(repo, commit, p))
    texts = {p: rd(p) for p in members}
    kinds = {p: {"kind": member_info[p][0], "detail": member_info[p][1]} for p in members}
    parts = []
    if any(kinds[p]["kind"] == "added" for p in members):
        parts.append("added")
    if any(kinds[p]["kind"] == "modified" for p in members):
        parts.append("modified")
    for p in members:
        if kinds[p]["kind"] == "renamed":
            parts.append("renamed_from " + rename_old.get(p, ""))
    change_kind = "; ".join(parts)
    rep = next((p for p in members if texts[p].strip()), members[0])
    excerpt = "\n".join(texts[rep].splitlines()[: int(settings["excerpt_lines"])])
    if symbols_from_text:
        syms = symbols_from_lines(texts[rep].splitlines())
    else:
        syms = top_symbols_git(repo, parent, commit, rep)
    stratum, reason, sibs = stratify(members, cluster["role"], tree_set, ownership, core_gate)
    pr = F.pr_number(subject)
    ds = dict(diff_summary, top_symbols=syms)
    state = {"unit": {"paths": members[:10], "change_kind": change_kind, "role": cluster["role"], "excerpt": excerpt},
             "pr": {"title": subject, "body": body[:1500], "changed_paths": changed_paths[:100]}, "diff_summary": ds}
    evidence = {"unit": {"paths": members, "change_kind": change_kind, "role": cluster["role"],
                         "members": [{"path": p, "text": "\n".join(texts[p].splitlines()[:300])} for p in members]},
                "pr": {"title": subject, "body": body[:6000], "changed_paths": changed_paths[:300]}, "diff_summary": ds}
    return {"case_id": case_id, "round": r, "cluster_id": cluster_id, "commit": commit, "parent": parent, "pr": pr, "pr_key": str(pr) if pr else commit[:8], "date": date,
            "pr_title": subject, "pr_body": body[:1500], "dir": cluster["dir"], "stem": cluster["stem"], "members": members, "member_kinds": kinds, "change_kind": change_kind,
            "role": cluster["role"], "scope": cluster["scope"], "stratum": stratum, "stratum_reason": reason, "siblings": sibs, "state": state, "label_evidence": evidence,
            "control": control, "set": set_name}


def restratify(row, tree_set, ownership, core_gate=False):
    """Copy of `row` with siblings/stratum recomputed against another membership (contrast catalogue)."""
    st, reason, sibs = stratify(row["members"], row["role"], tree_set, ownership, core_gate)
    return dict(row, stratum=st, stratum_reason=reason, siblings=sibs)


def _changed_paths(entries):
    return [{"path": e[3] if e[0] != "D" else e[2], "status": STATUS_WORD.get(e[0], e[0])} for e in entries]


# ---- a round
def round_info(r):
    pr = json.load(open(os.path.join(x2d.HERE, "prepass", "rounds.json")))
    return next(x for x in pr["rounds"] if x["round"] == r)


def build_round_units(repo, r, membership, rounds_info=None, data=None, settings=None, core_gate=None, check_prepass=True, audit=True, hook=None, controls=False, catalogue=None):
    """Pure (no file written). -> {"units","audit_units","controls","capped","lookups","inherit","counts"}.
    rounds_info: {"after_commit","through_commit","prs":[{"pr","sha","scored_ids"}]} (default: prepass/rounds.json round r). hook(row, tree_set) is called for every
    scored unit as it is built (K2/K3 gates). core_gate None = `load_core_gate(data)` (decided once at round 0). controls=True also builds the per-round controls
    (dup, trap, prose; `catalogue` default = the study catalogue of `data`)."""
    info = rounds_info or round_info(r)
    settings = settings or x2d.round_settings(r, data)
    own = F.Ownership(membership)
    state = F.TreeState(F.tree_paths(repo, info["after_commit"]))
    tree0 = set(state.paths)
    if core_gate is None:
        core_gate = load_core_gate(data)
    prs = {p["sha"]: p for p in info.get("prs", [])}
    visited = set()
    units, capped, lookups, inherit = [], [], [], []
    audit_best = []  # (hash, id, row, lookup item), <= 10 smallest
    total = collections.Counter()
    first_date = None
    for seq, (sha, parent, date, subject0) in enumerate(F.first_parent_commits(repo, info["after_commit"], info["through_commit"])):
        first_date = first_date or date
        entries = F.diff_entries(repo, parent, sha)
        blobs = {e[3]: e[4] for e in entries if e[0] in "ARC"}
        if check_prepass and sha in prs and "scored_ids" in prs[sha]:
            visited.add(sha)
            _, res0 = F.classify(entries, state)
            live0, _i = F.clusters(res0, blobs)
            ids0 = sorted((F.cluster_id(sha, c) for c in live0), key=x2d.cap_key)[: x2d.PR_CAP]
            if ids0 != prs[sha]["scored_ids"]:
                sys.exit("refusing: membership-free scored ids of %s (%s) differ from prepass/rounds.json (code or repo drift): %s != %s" % (sha, prs[sha].get("pr"), ids0, prs[sha]["scored_ids"]))
        resolved = []
        counts, residual = F.owned_classify(entries, state, own, resolved)
        live, inert = F.clusters(residual, blobs)
        counts["inert_clusters"] += len(inert)
        counts["residual_clusters"] += len(live)
        subject, body = commit_message(repo, sha) if (live or resolved) else (subject0, "")
        res_map = {x["path"]: x for x in resolved}
        for st, _s, old, new, _b in entries:  # rule effects on the membership
            if st == "D":
                inherit.append({"commit": sha, "seq": seq, "op": "remove", "path": old, "old": None, "area_ids": []})
            elif st == "R":
                if new in res_map:
                    inherit.append({"commit": sha, "seq": seq, "op": "move", "path": new, "old": old, "area_ids": res_map[new]["owners"]})
                else:
                    inherit.append({"commit": sha, "seq": seq, "op": "remove", "path": old, "old": None, "area_ids": []})
            elif new in res_map:
                inherit.append({"commit": sha, "seq": seq, "op": "add", "path": new, "old": None, "area_ids": res_map[new]["owners"]})
        pr = F.pr_number(subject0)
        pr_key = str(pr) if pr else sha[:8]
        ctx = None

        def context():
            nonlocal ctx
            if ctx is None:
                ctx = (_changed_paths(entries), numstat(repo, parent, sha), {e[3]: e[2] for e in entries if e[0] in "RC"})
            return ctx

        for c in live:
            c["id"] = F.cluster_id(sha, c)
        order = sorted(live, key=lambda c: x2d.cap_key(c["id"]))
        scored, cut = order[: x2d.PR_CAP], order[x2d.PR_CAP:]
        counts["scored_clusters"] += len(scored)
        counts["capped_clusters"] += len(cut)
        for c in cut:
            capped.append({"cluster_id": c["id"], "pr": pr_key, "reason": "pr_cap", "round": r})
        res_info = {p: (k, d) for p, k, d in residual}
        for c in sorted(scored, key=lambda c: c["id"]):
            cp, ds, ren = context()
            row = make_unit(repo, r, x2d.case_id(r, c["id"]), c["id"], sha, parent, date, subject, body, c, res_info, ren, cp, ds, state.paths, own, settings, "window", core_gate)
            row["seq"] = seq
            units.append(row)
            if hook:
                hook(row, state.paths)
        # audit candidates: (c)/(d)-resolved files grouped like clusters
        item_ids = []
        if audit and resolved:
            groups = collections.OrderedDict()
            for x in resolved:
                if x.get("pool"):  # pooled renames have no inherited area to audit
                    continue
                groups.setdefault(F.cluster_key(x["path"]), []).append(x)
            for (d, stem), xs in groups.items():
                if all(x["blob"] == F.EMPTY_BLOB for x in xs):
                    continue
                c = {"dir": d, "stem": stem, "files": sorted(x["path"] for x in xs), "role": F.role(xs[0]["path"]), "scope": F.scope(xs[0]["path"])}
                cid = F.cluster_id(sha, c)
                h = hashlib.sha256(("x2d-audit\0" + cid).encode()).hexdigest()
                if len(audit_best) >= 10 and h >= audit_best[-1][0]:
                    continue
                cp, ds, ren = context()
                mi = {x["path"]: ("renamed" if x["rule"] == "c" else "added", "inherit rule " + x["rule"]) for x in xs}
                row = make_unit(repo, r, None, cid, sha, parent, date, subject, body, c, mi, ren, cp, ds, state.paths, own, settings, "audit", core_gate)
                row["seq"] = seq
                owners = sorted({a for x in xs for a in x["owners"]})
                item = {"path": c["files"][0], "rule": "+".join(sorted({x["rule"] for x in xs})), "inherited_area_id": owners[0] if owners else None, "inherited_area_ids": owners}
                audit_best.append((h, cid, row, item, pr_key))
                audit_best.sort(key=lambda t: t[0])
                del audit_best[10:]
        if live or resolved or sha in prs or any(counts.values()):  # every commit with a rule count has a row (R2-07: modification-only commits too)
            lookups.append({"pr": pr_key, "commit": sha, "counts": dict(counts), "audit": []})
        total.update(counts)
        state.apply(entries)
    if check_prepass:
        missing = sorted(set(prs) - visited)
        if missing:
            sys.exit("refusing: rounds.json PR commits not in %s..%s: %s" % (info["after_commit"], info["through_commit"], missing[:3]))
    audit_units = []
    by_commit = {l["commit"]: l for l in lookups}
    for n, (h, cid, row, item, pr_key) in enumerate(audit_best, 1):
        row["case_id"] = "%s:aud-%d" % (r, n)
        audit_units.append(row)
        item["case_id"] = row["case_id"]
        by_commit[row["commit"]]["audit"].append(item)
    ctl = []
    if controls:
        ctl = round_controls(repo, r, units, info, tree0, own, settings, core_gate, first_date, catalogue if catalogue is not None else x2d.read_catalogue(data))
    return {"units": units, "audit_units": audit_units, "controls": ctl, "capped": capped, "lookups": lookups, "inherit": inherit, "counts": dict(total), "core_gate": core_gate}


def round_controls(repo, r, units, info, tree0, own, settings, core_gate, first_date, cat):
    """Per-round controls (R1-07 (a)): ceil(10 %) duplicates of hash-chosen window units (test-retest), one empty-file rename trap, one misleading-prose clone. All rows have
    control true, set ctl, `control_kind`, `source_case_id`; ids `<r>:ctl-<n>` in a hash order. -> [rows]. Never metrics, never absorbed (gold._absorb takes window rows only)."""
    items = []
    by_hash = lambda prefix: sorted(units, key=lambda u: _h(prefix, u["case_id"]))
    n_dup = math.ceil(0.10 * len(units))
    dup_src = by_hash("x2d-dup")[:n_dup]
    for u in dup_src:
        items.append(("dup", _control_clone(u, "dup", _h("x2d-ctl-dup", u["cluster_id"])[:16])))
    taken = {u["case_id"] for u in dup_src}
    pool = [u for u in by_hash("x2d-prose") if u["case_id"] not in taken] or by_hash("x2d-prose")
    if pool:
        src = pool[0]
        avoid = src["siblings"][0]["area_id"] if src["siblings"] else None
        decoy, title, body = prose_title(cat, avoid)
        row = _control_clone(src, "prose", _h("x2d-ctl-prose", src["cluster_id"])[:16])
        row["pr_title"], row["pr_body"] = title, body[:1500]
        row["state"]["pr"]["title"], row["state"]["pr"]["body"] = title, body[:1500]
        row["label_evidence"]["pr"]["title"], row["label_evidence"]["pr"]["body"] = title, body[:6000]
        row["pr"], row["pr_key"] = None, row["commit"][:8]
        row["decoy_area_id"] = decoy
        items.append(("prose", row))
    new_p, old_p = "docs_src/x2d_control_r%d/pkg/__init__.py" % r, "tests/test_tutorial/x2d_control_r%d/__init__.py" % r
    nd, stem = F.cluster_key(new_p)
    c = {"dir": nd, "stem": stem, "files": [new_p], "role": F.role(new_p), "scope": F.scope(new_p)}
    sha = _h("x2d-trap", "%d\0%s" % (r, new_p))[:40]
    row = make_unit(repo, r, None, _h("x2d-ctl-trap", "%d\0%s" % (r, new_p))[:16], sha, info["after_commit"], first_date or "2025-01-01T00:00:00+00:00",
                    "♻️ Move the package marker of an example next to the docs", "Moves the package marker so that tests and docs examples live next to each other.", c,
                    {new_p: ("renamed", "empty-trap from " + old_p)}, {new_p: old_p}, [{"path": new_p, "status": "renamed"}], {"files": 1, "added_lines": 0, "deleted_lines": 0},
                    tree0, own, settings, "ctl", core_gate, True, reader={new_p: ""}.get, symbols_from_text=True)
    row["control_kind"], row["source_case_id"], row["seq"] = "trap", None, 0
    items.append(("trap", row))
    items.sort(key=lambda it: _h("x2d-ctl-order", "%d|%s|%s" % (r, it[0], it[1]["cluster_id"])))
    out = []
    for n, (_k, row) in enumerate(items, 1):
        row["case_id"] = "%d:ctl-%d" % (r, n)
        out.append(row)
    return out


def _control_clone(u, kind, cluster_id):
    c = copy.deepcopy(u)
    c["source_case_id"], c["case_id"], c["cluster_id"] = u["case_id"], None, cluster_id
    c["control"], c["control_kind"], c["set"] = True, kind, "ctl"
    return c


# ---- reserve (D/P) units, used by build_controls
def build_reserve_units(repo, which, ownership, data=None, settings=None, core_gate=False, reserve=None):
    """Real units for reserve set D or P (pre-C0 commits, trees at the parent commit). -> [rows] in hash order. Members are recomputed by replaying the
    membership-free front end on the commit (the reserve file lists only 6 files per cluster)."""
    reserve = reserve or json.load(open(os.path.join(x2d.HERE, "prepass", "reserve.json")))
    settings = settings or x2d.round_settings(0, data)
    rows = []
    for rc in [c for c in reserve["clusters_hash_order"] if c["set"] == which]:
        sha = rc["sha"]
        parent = F.git(repo, "rev-parse", sha + "^").decode().strip()
        entries = F.diff_entries(repo, parent, sha)
        state = F.TreeState(F.tree_paths(repo, parent))
        _, residual = F.classify(entries, state)
        blobs = {e[3]: e[4] for e in entries if e[0] in "ARC"}
        live, _ = F.clusters(residual, blobs)
        c = next((c for c in live if F.cluster_id(sha, c) == rc["id"]), None)
        if c is None:
            sys.exit("refusing: reserve cluster %s not reproduced at %s" % (rc["id"], sha))
        subject, body = commit_message(repo, sha)
        info = {p: (k, d) for p, k, d in residual}
        row = make_unit(repo, 0, x2d.case_id(0, rc["id"]), rc["id"], sha, parent, rc["date"], subject, body, c, info, {e[3]: e[2] for e in entries if e[0] in "RC"},
                        _changed_paths(entries), numstat(repo, parent, sha), state.paths, ownership, settings, which, core_gate)
        rows.append(row)
    return rows


def contrast_units(repo, r, data=None):
    """-> (rows, sample): rows = the round's scored window units re-stratified against the contrast catalogue (units.contrast.jsonl, every round); for r in CONTRAST_ROUNDS the
    rows also hold the 20 D units (round 0's `units.contrast.jsonl` rows with set "D", re-stratified, case ids unchanged; R2-06). sample = for r in CONTRAST_ROUNDS the
    <= 30 hash-sampled (sha256("x2d-contrast\\0"+case_id)) WINDOW units, sorted by case_id (units.contrast_window.jsonl: the units labelled against the contrast catalogue); [] otherwise."""
    cat = json.load(open(os.path.join(data or x2d.DATA, "catalogue", "contrast-catalogue.json")))
    own = F.Ownership(G.catalogue_membership(cat, 0))
    gate = compute_core_gate(own, F.tree_paths(repo, x2d.C0))  # static: the contrast catalogue and the C0 tree do not change
    trees, out = {}, []
    for u in x2d.read_jsonl(os.path.join(x2d.round_dir(r, data), "units.jsonl")):
        if u["set"] != "window":
            continue
        if u["parent"] not in trees:
            trees[u["parent"]] = set(F.tree_paths(repo, u["parent"]))
        out.append(restratify(u, trees[u["parent"]], own, gate))
    sample = []
    if r in x2d.CONTRAST_ROUNDS:
        sample = sorted(sorted(out, key=lambda u: _h("x2d-contrast", u["case_id"]))[:CONTRAST_WINDOW_MAX], key=lambda u: u["case_id"])
        for u in x2d.read_jsonl(os.path.join(x2d.round_dir(0, data), "units.contrast.jsonl")):  # the 20 D units join the contrast set (R2-06), restratified, ids unchanged
            if u["set"] == "D":
                if u["parent"] not in trees:
                    trees[u["parent"]] = set(F.tree_paths(repo, u["parent"]))
                out.append(restratify(u, trees[u["parent"]], own, gate))
    return out, sample


def _membership_for_round(r, data):
    p = os.path.join(x2d.round_dir(r, data), "membership.json")
    want = G.round_membership(r, data)
    if os.path.exists(p):
        have = json.load(open(p))
        if json.dumps(have, sort_keys=True) != json.dumps(want, sort_keys=True):
            sys.exit("refusing: %s differs from the %s" % (p, "catalogue membership" if r == 1 else "gold.next_membership(%d)" % (r - 1)))
        return have, False
    return want, True


def gates(repo, data=None):
    import baselines
    out = os.path.join(x2d.round_dir(0, data), "k23.json")
    x2d.refuse_if_exists(out)
    m = G.catalogue_membership(x2d.read_catalogue(data), 1)
    own = F.Ownership(m)
    c0_tree = F.tree_paths(repo, x2d.C0)
    c0_dirs = {F.norm_dir(F.split_name(p)[0]) for p in c0_tree}
    gate = compute_core_gate(own, c0_tree)
    per, tot, frozen_pr, frozen_pr_all = {}, collections.Counter(), {}, {}
    for r in range(1, x2d.N_ROUNDS + 1):
        c = collections.Counter()
        pr_c = collections.Counter()      # the verdict's population: core_gate units excluded (R3-11)
        pr_all = collections.Counter()    # every scored cluster, kept for the record

        def hook(row, tree_set):
            c["n"] += 1
            pr_all[row["pr_key"]] += 1
            if row["stratum_reason"] != "core_gate":
                pr_c[row["pr_key"]] += 1
            if baselines.b0(row["members"], tree_set, own)[0] == "assigned":
                c["k2"] += 1
            if baselines.b2(row["members"], tree_set, own)[0] == "assigned":
                c["k2_b2"] += 1
            if row["stratum"] == 3:
                if row["stratum_reason"] != "core_gate" and row["dir"] in c0_dirs:
                    c["k3"] += 1
                else:
                    c["core_gate" if row["stratum_reason"] == "core_gate" else "new_dir"] += 1

        build_round_units(repo, r, m, data=data, core_gate=gate, check_prepass=True, audit=False, hook=hook)
        if r in FROZEN_ROUNDS:
            frozen_pr[r] = dict(pr_c)
            frozen_pr_all[r] = dict(pr_all)
        n = c["n"]
        per[r] = {"n": n, "k2_unique_owner": c["k2"], "k2_b2_assigned": c["k2_b2"], "k3_stratum3": c["k3"], "new_dir": c["new_dir"], "core_gate": c["core_gate"],
                  "k2_share": c["k2"] / n if n else None, "k2_b2_share": c["k2_b2"] / n if n else None, "k3_share": c["k3"] / n if n else None}
        tot.update(c)
    res = {"k2_share": tot["k2"] / tot["n"] if tot["n"] else None, "k2_b2_share": tot["k2_b2"] / tot["n"] if tot["n"] else None,
           "k3_share": tot["k3"] / tot["n"] if tot["n"] else None, "new_dir": tot["new_dir"], "core_gate": {"enabled": gate}, "per_round": per,
           "real_none_proxy": sum(per[r]["k3_stratum3"] + per[r]["new_dir"] for r in FROZEN_ROUNDS),  # excludes core_gate (R2-11)
           "frozen_pr_counts": {str(r): frozen_pr[r] for r in FROZEN_ROUNDS},  # core_gate units excluded: the verdict's PR structure (R3-11)
           "frozen_pr_counts_all_units": {str(r): frozen_pr_all[r] for r in FROZEN_ROUNDS},
           "counts": {"scored_clusters": tot["n"], "b0_unique_owner": tot["k2"], "b2_assigned": tot["k2_b2"], "stratum_3": tot["k3"], "stratum_3_new_dir": tot["new_dir"],
                      "stratum_3_core_gate": tot["core_gate"]}}
    x2d.write_json_new(out, res)
    return res


def rerender_units(repo, units, excerpt_lines):
    """R3-05: copies of `units` whose `state.unit.excerpt` is re-rendered under `excerpt_lines` from the same member text at the unit's commit (same representative-member rule as make_unit:
    the first sorted member with non-blank text); ids, membership, strata, siblings and every other field are unchanged. Each row records `excerpt_lines`."""
    out = []
    for u in units:
        members = sorted(u["members"])
        texts = {p: read_text(repo, u["commit"], p) for p in members}
        rep = next((p for p in members if texts[p].strip()), members[0])
        n = copy.deepcopy(u)
        n["state"]["unit"]["excerpt"] = "\n".join(texts[rep].splitlines()[: int(excerpt_lines)])
        n["excerpt_lines"] = int(excerpt_lines)
        out.append(n)
    return out


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--repo", required=True)
    ap.add_argument("--round", type=int)
    ap.add_argument("--tag", default=None, choices=[None, "contrast"])
    ap.add_argument("--gates", action="store_true")
    ap.add_argument("--data", default=None)
    a = ap.parse_args(argv)
    bad = x2d.verify_pins(require_signoff=x2d.require_signoff())
    if bad:
        sys.exit("pins: " + "; ".join(bad))
    if a.gates:
        res = gates(a.repo, a.data)
        print(json.dumps({k: res[k] for k in ("k2_share", "k2_b2_share", "k3_share", "new_dir", "core_gate", "counts", "real_none_proxy", "frozen_pr_counts")}))
        return
    if a.round is None:
        sys.exit("--round or --gates required")
    rd = x2d.round_dir(a.round, a.data)
    if a.tag == "contrast":
        if a.round < 1:
            sys.exit("--tag contrast: rounds 1..10 (round 0's contrast units come from build_controls)")
        p = os.path.join(rd, "units.contrast.jsonl")
        pw = os.path.join(rd, "units.contrast_window.jsonl")
        x2d.refuse_if_exists(p, *([pw] if a.round in x2d.CONTRAST_ROUNDS else []))
        rows, sample = contrast_units(a.repo, a.round, a.data)
        x2d.write_jsonl_new(p, rows)
        if a.round in x2d.CONTRAST_ROUNDS:
            x2d.write_jsonl_new(pw, sample)
        print("wrote %d contrast units, %d window sample" % (len(rows), len(sample)))
        return
    outs = [os.path.join(rd, n) for n in ("units.jsonl", "capped.jsonl", "lookups.jsonl", "inherit.jsonl", "counts.json")]
    x2d.refuse_if_exists(*outs)
    membership, fresh = _membership_for_round(a.round, a.data)
    res = build_round_units(a.repo, a.round, membership, data=a.data, controls=True)
    if fresh:
        x2d.write_json_new(os.path.join(rd, "membership.json"), membership)
    x2d.write_jsonl_new(outs[0], res["units"] + res["audit_units"] + res["controls"])
    x2d.write_jsonl_new(outs[1], res["capped"])
    x2d.write_jsonl_new(outs[2], res["lookups"])
    x2d.write_jsonl_new(outs[3], res["inherit"])
    x2d.write_json_new(outs[4], res["counts"])
    print(json.dumps({"round": a.round, "scored": len(res["units"]), "audit": len(res["audit_units"]), "controls": len(res["controls"]), "capped": len(res["capped"]), "core_gate": res["core_gate"]}))


if __name__ == "__main__":
    main()
