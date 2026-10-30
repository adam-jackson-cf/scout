"""X2d deterministic front end (shape-decision section 1, rules (a)-(f) and the residual-cluster rule), catalogue-free part.

Pure functions plus a git reader. Standard library only (run with .tools/venv/bin/python). The same module is used by the pre-pass (`prepass.py`, distributions only)
and, with an ownership oracle added by `build_units.py`, by the real run. Everything here is deterministic; no model is called.

Catalogue-free proxies (the pre-pass has no catalogue): "owned" in rule (d) is proxied by "a variant-stem sibling exists in the parent tree"; rules (a) and (f), and
the still-unowned modified files of the residual definition, need the catalogue and are not evaluated here (counted as `modified`).
"""
import collections, hashlib, re, subprocess

EMPTY_BLOB = "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391"  # git's empty blob: the 100 % "rename" of an empty file is a trap (shape-decision 0.c, #9268)
RENAME_MIN_SIMILARITY = 50  # percent; git -M50% (the 0.c guard)

# ---- scopes of the onboarding draws (shape-decision 0.a step 1); a rename inherits only within one scope
def scope(path):
    """Rename domain (PREREGISTRATION section 3 rule (c)): `core` = fastapi/**, `examples-tests` = docs_src/** + tests/**, `infra-config` = everything else.
    Not the five onboarding scopes of section 4.1 (those are inventory shards)."""
    if path.startswith("fastapi/"):
        return "core"
    if path.startswith("docs_src/") or path.startswith("tests/"):
        return "examples-tests"
    return "infra-config"


def role(path):
    """Reporting role (shape-decision 0.b): core, example, test, infra, config, docs-config."""
    if path.startswith("fastapi/"):
        return "core"
    if path.startswith("docs_src/"):
        return "example"
    if path.startswith("tests/"):
        return "test"
    if path.startswith(".github/") or path.startswith("scripts/"):
        return "infra"
    if path.startswith("docs/"):
        return "docs-config"
    return "config"


# ---- policy (e): files outside the area-eligible population. Counted and reported by reason, never judged.
_MEDIA = {".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".webp", ".woff", ".woff2", ".ttf", ".eot", ".pdf", ".mp4", ".webm"}
_LANG_DIR = re.compile(r"^docs/([a-z]{2,3}(?:-[A-Za-z]{2,4})?)/")


def excluded_reason(path):
    """-> None (eligible) or the reason: `markdown` (any .md), `translation` (docs/<lang>/** with lang != en), `media` (binary assets: no text excerpt exists)."""
    low = path.lower()
    if low.endswith(".md"):
        return "markdown"
    m = _LANG_DIR.match(path)
    if m and m.group(1) != "en":
        return "translation"
    dot = low.rfind(".")
    if dot != -1 and low[dot:] in _MEDIA:
        return "media"
    return None


# ---- variant stems (rule (d) and clustering)
_VARIANT = re.compile(r"(?:_an)?(?:_py3\d+)?(?:_pv[12])?$")
_NUM = re.compile(r"_\d{1,2}$")


def split_name(path):
    d, _, f = path.rpartition("/")
    stem, dot, ext = f.rpartition(".") if "." in f else (f, "", "")
    return d, stem, ext


def variant_stem(stem):
    """tutorial003_an_py39 -> tutorial003; test_tutorial003_py310 -> test_tutorial003; tutorial003_01 -> tutorial003 (`_NN`)."""
    s = stem
    while True:
        s2 = _NUM.sub("", _VARIANT.sub("", s))
        if s2 == s or not s2:
            break
        s = s2
    return s


def norm_dir(d):
    """Directory variants too: docs_src/app_testing/app_b_an -> docs_src/app_testing/app_b; tests/test_tutorial/test_x_py39 -> .../test_x."""
    return "/".join(_VARIANT.sub("", seg) or seg for seg in d.split("/"))


def cluster_key(path):
    d, stem, ext = split_name(path)
    return norm_dir(d), variant_stem(stem)


def sibling_key(path):
    d, stem, ext = split_name(path)
    return norm_dir(d), variant_stem(stem), ext


# ---- git reading
def git(repo, *args, check=True):
    return subprocess.run(["git", "-C", repo, *args], capture_output=True, check=check).stdout


def first_parent_commits(repo, base, tip):
    """[(sha, first_parent, iso_date, subject)] oldest first, for base (exclusive) .. tip (inclusive), first-parent line."""
    out = git(repo, "rev-list", "--first-parent", "--reverse", "--format=%H%x1f%P%x1f%cI%x1f%s", "--no-commit-header", "%s..%s" % (base, tip)).decode("utf-8", "replace")
    rows = []
    for line in out.splitlines():
        if not line.strip():
            continue
        h, p, d, s = line.split("\x1f", 3)
        rows.append((h, p.split(" ")[0] if p else None, d, s))
    return rows


def diff_entries(repo, parent, sha):
    """Raw name-status with rename detection (-M50%): [(status, similarity, old_path, new_path, new_blob)]; status in A D M T R."""
    if parent is None:
        return []
    raw = git(repo, "diff-tree", "-r", "-z", "--raw", "--no-abbrev", "-M%d%%" % RENAME_MIN_SIMILARITY, parent, sha)
    tok = raw.decode("utf-8", "surrogateescape").split("\0")
    out, i = [], 0
    while i < len(tok):
        t = tok[i]
        if not t.startswith(":"):
            i += 1
            continue
        meta = t[1:].split(" ")
        status = meta[4]
        new_blob = meta[3]
        if status[0] in "RC":
            out.append((status[0], int(status[1:] or 0), tok[i + 1], tok[i + 2], new_blob))
            i += 3
        else:
            out.append((status[0], 100, tok[i + 1], tok[i + 1], new_blob))
            i += 2
    return out


def tree_paths(repo, rev):
    return git(repo, "ls-tree", "-r", "--name-only", "-z", rev).decode("utf-8", "surrogateescape").split("\0")[:-1]


class TreeState:
    """The set of paths and the (normdir, variant stem, ext) sibling counts of the parent tree, advanced commit by commit."""

    def __init__(self, paths):
        self.paths = set(paths)
        self.sib = collections.Counter(sibling_key(p) for p in self.paths)

    def add(self, p):
        if p not in self.paths:
            self.paths.add(p)
            self.sib[sibling_key(p)] += 1

    def remove(self, p):
        if p in self.paths:
            self.paths.discard(p)
            self.sib[sibling_key(p)] -= 1

    def apply(self, entries):
        for st, _, old, new, _b in entries:
            if st == "D":
                self.remove(old)
            elif st == "R":
                self.remove(old)
                self.add(new)
            elif st in "AC":
                self.add(new)


def classify(entries, state):
    """Apply rules (a)-(f) (catalogue-free proxies) to one commit's entries against the PARENT tree `state` (not mutated).
    -> (counts: Counter of rule names, residual: list of (path, kind, detail)). Rule names: modified(a) deleted(b) renamed_inherit(c) variant_inherit(d)
    excluded:<reason>(e) residual_added residual_renamed:<cause> inert_empty."""
    counts, residual = collections.Counter(), []
    # siblings that this same commit adds do not count as pre-existing owners: only `state` is consulted
    for st, sim, old, new, blob in entries:
        if st == "D":
            counts["deleted"] += 1
            continue
        if st in "MT":
            counts["modified"] += 1
            continue
        why = excluded_reason(new)
        if why:
            counts["excluded:" + why] += 1
            continue
        if st == "R":
            if blob == EMPTY_BLOB:
                counts["renamed_empty_trap"] += 1
                residual.append((new, "renamed", "empty-trap from " + old))
                continue
            if scope(old) == scope(new):
                counts["renamed_inherit"] += 1
                continue
            counts["renamed_cross_scope"] += 1
            residual.append((new, "renamed", "cross-scope from " + old))
            continue
        # added (A) or copied (C)
        if state.sib.get(sibling_key(new), 0) > 0:
            counts["variant_inherit"] += 1
            continue
        residual.append((new, "added", ""))
        counts["residual_added"] += 1
    return counts, residual


def clusters(residual, blobs=None, dir_only=False):
    """Residual files -> clusters keyed (normalised parent directory, variant stem) [or directory only]. Clusters made only of empty files are `inert` (package
    markers) and returned separately. blobs: {path: blob sha} to identify empties."""
    groups = collections.OrderedDict()
    for path, kind, detail in residual:
        k = (norm_dir(split_name(path)[0]), "*") if dir_only else cluster_key(path)
        groups.setdefault(k, []).append((path, kind, detail))
    live, inert = [], []
    for k, members in groups.items():
        c = {"dir": k[0], "stem": k[1], "files": sorted(m[0] for m in members), "kinds": sorted({m[1] for m in members}),
             "role": role(members[0][0]), "scope": scope(members[0][0])}
        if blobs is not None and all(blobs.get(m[0]) == EMPTY_BLOB for m in members):
            inert.append(c)
        else:
            live.append(c)
    return live, inert


def cluster_id(commit, c):
    return hashlib.sha256(("%s\0%s\0%s" % (commit, c["dir"], c["stem"])).encode()).hexdigest()[:16]


def pr_number(subject):
    m = re.search(r"\(#(\d+)\)\s*$", subject)
    return int(m.group(1)) if m else None


def cut_rounds(eligible, n_rounds, min_clusters, min_prs):
    """Chronological cut. `eligible`: [(commit_index, pr_key, n_clusters)] oldest first (only commits with >= 1 cluster). A round closes at the earliest eligible commit at which
    clusters >= min_clusters AND distinct PRs >= min_prs. -> (rounds: list of [eligible items], leftover items, ok)."""
    rounds, cur, clus, prs = [], [], 0, set()
    for item in eligible:
        cur.append(item)
        clus += item[2]
        prs.add(item[1])
        if clus >= min_clusters and len(prs) >= min_prs:
            rounds.append(cur)
            cur, clus, prs = [], 0, set()
            if len(rounds) == n_rounds:
                break
    used = sum(len(r) for r in rounds)
    return rounds, eligible[used:], len(rounds) == n_rounds


# ---- ownership oracle (the run's front end; added by build_units' owner, existing behaviour above is unchanged)
class Ownership:
    """Reads ownership only from a round-start membership layer {"areas": {area_id: [files]}, "pool": [...], "heuristic_only": [...]}."""

    def __init__(self, membership):
        self.membership = membership
        self._owners = collections.defaultdict(set)
        for aid, files in membership.get("areas", {}).items():
            for f in files:
                self._owners[f].add(aid)
        self._heur = set(membership.get("heuristic_only", []))
        self._pool = set(membership.get("pool", []))
        self._by_dir = collections.defaultdict(list)
        self._by_sib = collections.defaultdict(list)
        for f in sorted(self._owners):
            self._by_dir[norm_dir(split_name(f)[0])].append(f)
            self._by_sib[sibling_key(f)].append(f)

    def owners(self, path):
        return set(self._owners.get(path, ()))

    def is_heuristic_only(self, path):
        return path in self._heur and path not in self._owners

    def is_pooled(self, path):
        """The file is in the membership's pool (gold `none` of an earlier unit) and owned by no area."""
        return path in self._pool and path not in self._owners

    def dir_files(self, norm_directory):
        """Owned files whose normalised parent directory is `norm_directory`."""
        return self._by_dir.get(norm_directory, [])

    def sibling_owners(self, path, tree_paths):
        """Owned siblings of `path` in the same normalised directory present in `tree_paths`: [{"path","area_id"}] sorted (one row per file and owner)."""
        tp = tree_paths if isinstance(tree_paths, (set, frozenset, dict)) else set(tree_paths)
        out = []
        for f in self._by_dir.get(norm_dir(split_name(path)[0]), []):
            if f != path and f in tp:
                for a in sorted(self._owners[f]):
                    out.append({"path": f, "area_id": a})
        return out

    def variant_owners(self, path, tree_paths):
        """Rule (d): owners of owned files with the same (normalised dir, variant stem, extension) present in `tree_paths`."""
        tp = tree_paths if isinstance(tree_paths, (set, frozenset, dict)) else set(tree_paths)
        out = set()
        for f in self._by_sib.get(sibling_key(path), []):
            if f != path and f in tp:
                out |= self._owners[f]
        return out


def owned_classify(entries, state, ownership, resolved=None):
    """Rules (a)-(g) against the PARENT tree `state` (not mutated) and the round-start `ownership`. -> (counts, residual) like `classify`.
    Counts: modified_owned (a) / modified_unowned (total, NOT residual; split `modified_unowned:pooled` = the file is in the membership pool after an earlier gold `none`,
    `modified_unowned:never_owned` = no area and no pool; reported as a coverage figure, never a unit, R1-04) / deleted (b) / renamed_inherit (c) / renamed_pooled (c, the old
    path is pooled: the new path stays pooled, an `op: move` within the pool) / renamed_unowned (residual) / renamed_empty_trap, renamed_cross_scope (residual) /
    variant_inherit (d, only with an OWNED sibling) / variant_unowned + residual_added / excluded:<reason> (e) / heuristic_only (f, also residual; a modified heuristic-only
    file is a unit).
    Rule (g) (inert clusters) is applied by `clusters`. `resolved`, if a list, receives {"path","rule":"c"|"d","owners":[area ids],"old":path|None,"blob","pool":bool}
    for every file resolved by (c)/(d) (`pool` true: a pooled rename, owners empty). Residual detail strings: `heuristic_only`, `unowned from <old>`, `empty-trap from <old>`,
    `cross-scope from <old>`."""
    counts, residual = collections.Counter(), []

    def add_residual(path, kind, detail):
        if ownership.is_heuristic_only(path):
            counts["heuristic_only"] += 1
            detail = (detail + "; " if detail else "") + "heuristic_only"
        residual.append((path, kind, detail))

    for st, sim, old, new, blob in entries:
        if st == "D":
            counts["deleted"] += 1
            continue
        why = excluded_reason(new)
        if why:
            counts["excluded:" + why] += 1
            continue
        if st in "MT":
            if ownership.owners(new):
                counts["modified_owned"] += 1
            elif ownership.is_heuristic_only(new):  # rule (f): a heuristic-only file is judged, modified or not
                add_residual(new, "modified", "")
            else:  # still unowned after onboarding / earlier gold: coverage figure, not a unit (R1-04)
                counts["modified_unowned"] += 1
                counts["modified_unowned:pooled" if ownership.is_pooled(new) else "modified_unowned:never_owned"] += 1
            continue
        if st == "R":
            if blob == EMPTY_BLOB:
                counts["renamed_empty_trap"] += 1
                add_residual(new, "renamed", "empty-trap from " + old)
                continue
            if scope(old) != scope(new):
                counts["renamed_cross_scope"] += 1
                add_residual(new, "renamed", "cross-scope from " + old)
                continue
            own = ownership.owners(old)
            if own:
                counts["renamed_inherit"] += 1
                if resolved is not None:
                    resolved.append({"path": new, "rule": "c", "owners": sorted(own), "old": old, "blob": blob, "pool": False})
                continue
            if ownership.is_pooled(old):  # a pooled file stays pooled when renamed (op: move within the pool)
                counts["renamed_pooled"] += 1
                if resolved is not None:
                    resolved.append({"path": new, "rule": "c", "owners": [], "old": old, "blob": blob, "pool": True})
                continue
            counts["renamed_unowned"] += 1
            add_residual(new, "renamed", "unowned from " + old)
            continue
        # added (A) or copied (C)
        own = ownership.variant_owners(new, state.paths)
        if own:
            counts["variant_inherit"] += 1
            if resolved is not None:
                resolved.append({"path": new, "rule": "d", "owners": sorted(own), "old": None, "blob": blob, "pool": False})
            continue
        if state.sib.get(sibling_key(new), 0) > 0:
            counts["variant_unowned"] += 1
        counts["residual_added"] += 1
        add_residual(new, "added", "")
    return counts, residual
