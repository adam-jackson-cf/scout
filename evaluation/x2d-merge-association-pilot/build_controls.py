"""X2d round-0 material: real reserve units (sets D and P), the controls of PREREGISTRATION 5/8, the modified catalogues and the Jev run manifest.

CLI:
    build_controls.py --repo CLONE [--data D]
Writes (all append-only; refuses when any exists):
    D/rounds/0/units.jsonl            D and P real units (pre-C0 commits, trees at the parent commit, membership = catalogue; `set` D/P) + controls (`0:ctl-<n>`,
                                      `control:true`, `set:"ctl"`), rows in a hash order so controls are not grouped
    D/rounds/0/control-key.jsonl      {"case_id","kind":"p1|p2|dup|trap|prose|planted_new","gold":{...}|null,"source_case_id","catalogue":"catalogue|catalogue-p1|
                                      catalogue-p2", ...}; labellers and Jev never see it. gold null = derived from the source unit's gold (p2 via gold.derive_ablation = the planted-gap gate,
                                      written by gold.build_gold; dup, prose). A control is labelled/scored against its designated `catalogue`, its kind is never shown.
    D/rounds/0/units.contrast.jsonl   the 20 D units, siblings/stratum recomputed against the contrast catalogue
    D/rounds/0/membership.json        catalogue membership (round 0)
    D/rounds/0/catalogue-p1.json      catalogue minus the files of the P1 items (unowned afterwards)
    D/rounds/0/catalogue-p2.json      catalogue minus the ablated areas (areas that own the P2 items)
    D/rounds/0/catalogue-n1.json      titles/summaries rotated across area ids (fixed order by sha256("x2d-n1\\0"+id)), sample_files dropped
    D/rounds/0/jev-in/{manifest.json,<name>.units.jsonl}   one run per entry: main, n1, p1, p2 (units subset, catalogue, out relative to D/rounds/0, arms, flip_all)
X2D_SKIP_SIGNOFF=1 -> verify_pins(require_signoff=False) (validation).

Construction (deterministic; no model; hash orders only):
  eligible_P = P units whose members are all owned by one single area A in the catalogue; eligible_D = the same for D units. n_p2 = min(12, eligible_P - 20 + eligible_D) (refusal below 8),
  P1 = 20 or more eligible placeable P units (loud refusal otherwise). P1: the members leave A in catalogue-p1 and re-enter as NEW files (same basenames and text) in a directory whose owned siblings belong to one
  other area (stratum 1), to two or more other areas (stratum 2) or to no area (stratum 3), cycling; rules (a)-(f) are verified not to resolve the new path.
  P2 (the planted-gap control, R1-07): clones of eligible units whose owning area is ablated. Sources, in order: unplaceable eligible P units, then placeable eligible P units beyond the 20 of P1, then eligible D units in hash order (rev 10;
  the D clones are source-dependent: their gold derives from D labels). gold = gold.derive_ablation(source gold, ablated_area_ids) with the SAME union the catalogue
  removes (R2-05; key rows carry `ablated_area_ids` and `gap_positive` = no member owned in catalogue-p2); the gate's denominator is the gap positives; build refuses below P2_MIN_GAP_POSITIVES = 8. There is no separate
  planted_gap control. planted_new: synthetic packages
  (fixed templates). trap: an empty `__init__.py` renamed across directories (#9268 shape). prose: a clone of a P unit whose PR title/body talk about another area.
  dup: ceil(10 %) clones of D/P units under new ids (test-retest). Control ids are 0:ctl-<n> in a hash order; control units carry the commit of their source (planted:
  a synthetic hash) and `parent` = C0 (the tree their paths were placed in). Per-round controls of rounds 1..10 are built by build_units (dup, trap, prose).
Importable: make_controls, fresh_pool, regate_plan, make_regate.

Rev 13 (R-053), the one permitted re-gate of P1 and P2 on FRESH items (PREREGISTRATION 5, rev 13; the round-0 gates stay as they are):
    build_controls.py --repo CLONE --regate-supply [--out FILE]   membership only, writes nothing under the data directory: the funnel of the fresh pool and the planned build
    build_controls.py --repo CLONE --regate                       writes D/rounds/0/regate/ (a self-contained data root: catalogue/, rounds/0/ with the frozen Jev side copied byte for byte, units.jsonl,
                                                                   control-key.jsonl, catalogue-p1/p2.json, jev-in/, membership.json, supply.json, regate.json) and nothing else; refuses when it exists (one shot)
Fresh pool (membership only, deterministic, fixed before any label): the residual clusters (front end of PREREGISTRATION 3, replayed on the first-parent commit that added their files) of every commit up to C0
whose files are all still in the C0 tree and all owned by ONE catalogue area; origin D / P / G (reserve sets, `prepass/reserve.json`) or `in_no_unit` (C0-owned files of no reserve unit). Excluded: a unit used by
the rev-12 P1 / P2 gates as source, any cluster sharing a file path with such a source (path twin), any cluster of a PR that contributed such a source (PR episode) and the PRs named in the guide (`#<digits>`).
Order: sha256("x2d-regate-order\\0" + cluster id); one cluster per PR and no shared path in the pool. P1: the first 40 clusters in that order that place (target stratum cycled over
x2d.REGATE_P1_STRATA = 1,1,1,2,3; the directory baselines B1 and B2 must not answer A (the ctl-23 leak); a stratum-1 directory is never reused); P2: from the clusters left, the areas that own at least
REGATE_P2_AREA_MIN of them in the hash order of sha256("x2d-regate-p2area\\0" + area id), up to REGATE_P2_PER_AREA clones each, until REGATE_P2_CLONES. P2 gold is the clone's OWN blind label on catalogue-p2 (gold.py); the source unit of every P2 clone
is labelled on the full catalogue as well (set "F") so that the derived gold can be reported beside it.
"""
import argparse, collections, copy, hashlib, json, math, os, re, shutil, sys, time

P2_MIN_GAP_POSITIVES = 8  # PREREGISTRATION 5: the P2 gate's denominator (clones whose derived gold is none/gap) must be at least this; fewer label-derived gap positives = UNMEASURED (R3-04)
P2_MAX_CLONES = 12  # clones built: n_p2 = min(12, eligible_P - 20 + eligible_D), refusal below P2_MIN_GAP_POSITIVES (R3-04; D top-up, rev 10)

import x2d
import frontend as F
import gold as G
import build_units as BU
import baselines as BL

PLANTED = [
    {"title": "✨ Add Kafka consumer integration for streaming endpoints", "body": "Adds a small consumer wrapper so that route handlers can read messages from a Kafka topic and acknowledge them after the response is sent.",
     "path": "fastapi/streaming/kafka_consumer.py",
     "code": '''import json
from typing import Any, Callable, Iterator, Optional


class TopicConsumer:
    """Reads JSON messages from a Kafka topic and hands them to a handler."""

    def __init__(self, brokers: str, topic: str, group_id: str = "fastapi") -> None:
        self.brokers = brokers
        self.topic = topic
        self.group_id = group_id
        self._offset: Optional[int] = None

    def poll(self, client: Any, timeout_ms: int = 1000) -> Iterator[dict]:
        for record in client.poll(timeout_ms=timeout_ms):
            self._offset = record.offset
            yield json.loads(record.value)

    def commit(self, client: Any) -> None:
        if self._offset is not None:
            client.commit(offset=self._offset + 1)


def consume_forever(consumer: TopicConsumer, client: Any, handler: Callable[[dict], None]) -> None:
    while True:
        for message in consumer.poll(client):
            handler(message)
        consumer.commit(client)
'''},
    {"title": "✨ Add spreadsheet export helpers for report endpoints", "body": "Adds helpers that turn a list of rows into an XLSX byte stream so that report endpoints can return downloadable spreadsheets.",
     "path": "fastapi/exports/spreadsheet.py",
     "code": '''import io
import zipfile
from typing import Iterable, List, Sequence
from xml.sax.saxutils import escape


def _cell(ref: str, value: object) -> str:
    if isinstance(value, (int, float)):
        return f'<c r="{ref}"><v>{value}</v></c>'
    return f'<c r="{ref}" t="inlineStr"><is><t>{escape(str(value))}</t></is></c>'


def _column(index: int) -> str:
    name = ""
    index += 1
    while index:
        index, rem = divmod(index - 1, 26)
        name = chr(65 + rem) + name
    return name


def render_sheet(rows: Iterable[Sequence[object]]) -> str:
    out: List[str] = []
    for r, row in enumerate(rows, start=1):
        cells = "".join(_cell(f"{_column(c)}{r}", v) for c, v in enumerate(row))
        out.append(f'<row r="{r}">{cells}</row>')
    return "<sheetData>" + "".join(out) + "</sheetData>"


class SpreadsheetResponseBody:
    def __init__(self, rows: Iterable[Sequence[object]]) -> None:
        self.rows = rows

    def to_bytes(self) -> bytes:
        buffer = io.BytesIO()
        with zipfile.ZipFile(buffer, "w") as archive:
            archive.writestr("xl/worksheets/sheet1.xml", render_sheet(self.rows))
        return buffer.getvalue()
'''},
    {"title": "✨ Add feature flag evaluation for request handling", "body": "Adds a minimal percentage-rollout feature flag evaluator keyed by a stable request attribute.",
     "path": "fastapi/rollout/flags.py",
     "code": '''import hashlib
from dataclasses import dataclass
from typing import Dict, Optional


@dataclass
class Flag:
    name: str
    percentage: int = 0
    enabled: bool = True


class FlagSet:
    def __init__(self) -> None:
        self._flags: Dict[str, Flag] = {}

    def register(self, flag: Flag) -> None:
        self._flags[flag.name] = flag

    def bucket(self, name: str, key: str) -> int:
        digest = hashlib.sha256(f"{name}:{key}".encode()).hexdigest()
        return int(digest[:8], 16) % 100

    def is_on(self, name: str, key: Optional[str] = None) -> bool:
        flag = self._flags.get(name)
        if flag is None or not flag.enabled:
            return False
        if key is None:
            return flag.percentage >= 100
        return self.bucket(name, key) < flag.percentage
'''},
]


def _h(prefix, key):
    return hashlib.sha256((prefix + "\0" + key).encode()).hexdigest()


def _filter_catalogue(cat, remove_paths=(), drop_areas=()):
    out = copy.deepcopy(cat)
    rm, da = set(remove_paths), set(drop_areas)
    areas = []
    for a in out["areas"]:
        if a["area_id"] in da:
            continue
        a["files"] = [f for f in a["files"] if f not in rm]
        a["sample_files"] = [f for f in a.get("sample_files", []) if f not in rm]
        if a["files"]:
            areas.append(a)
    out["areas"] = areas
    out["jev_assigned"] = [j for j in out.get("jev_assigned", []) if j["path"] not in rm and j["area_id"] not in da]
    return out


def _n1_catalogue(cat):
    out = copy.deepcopy(cat)
    order = sorted(range(len(out["areas"])), key=lambda i: _h("x2d-n1", out["areas"][i]["area_id"]))
    tit = [(out["areas"][i]["title"], out["areas"][i]["summary"]) for i in order]
    if len(tit) > 1:
        tit = tit[1:] + tit[:1]
    for i, (t, s) in zip(order, tit):
        out["areas"][i]["title"], out["areas"][i]["summary"] = t, s
    for a in out["areas"]:
        a["sample_files"] = []
    return out


def _clone(row, case_id, cluster_id):
    c = copy.deepcopy(row)
    c["case_id"], c["cluster_id"], c["control"], c["set"] = case_id, cluster_id, True, "ctl"
    return c


def _unique_owner(row, own):
    owners = set()
    for m in row["members"]:
        o = own.owners(m)
        if len(o) != 1:
            return None
        owners |= o
    return next(iter(owners)) if len(owners) == 1 else None


def _owned_dirs(p1_own, tree):
    """{directory: set of owning areas of the tree files directly in it}."""
    own_dirs = collections.defaultdict(set)
    for f in tree:
        own_dirs[f.rpartition("/")[0]] |= p1_own.owners(f)
    return own_dirs


def _candidate_dirs(src_row, area, own_dirs, taken=(), forbid_dirs=(), rej=None):
    """{1: [...], 2: [...], 3: [...]}: the directories of the same role and scope that do not hold A's own files, by kind: owned siblings of exactly one other area (1), of two or more (2), of none (3).
    `forbid_dirs` (rev 13, R17-01): the target directories of the rev-12 P1 items, never reused (counted in `rej["dir"]` when given)."""
    role, scope = src_row["role"], src_row["scope"]
    cands = {1: [], 2: [], 3: []}
    for d in sorted(own_dirs):
        if not d or "__pycache__" in d or F.norm_dir(d) == src_row["dir"]:
            continue
        if d in forbid_dirs:
            if rej is not None:
                rej["dir"] += 1
            continue
        if d in taken:
            continue
        if F.role(d + "/x") != role or F.scope(d + "/x") != scope:
            continue
        o = own_dirs[d]
        if area in o:
            continue
        cands[1 if len(o) == 1 else 2 if len(o) >= 2 else 3].append(d)
    return cands


def _p1_target(src_row, area, p1_own, tree, c0_state, stratum, leak_check=False, taken=(), own_dirs=None, forbid_paths=(), forbid_dirs=(), rej=None):
    """-> (new paths, directory) placing the members in a directory of the wanted kind; None when impossible. `leak_check` (rev 13, the ctl-23 leak): refuse a directory for which the directory
    baselines B1 or B2 would still answer the item's own area A (a stratum-3 directory whose first owned ancestor within two levels is A's, or any directory whose mirror/sibling vote is A), so a
    baseline can only recover A by a path cue that points at A, never through a parent-directory fallback to A; `taken` = directories already used by earlier items (never reused).
    `forbid_paths` / `forbid_dirs` (R17-01): the presented paths of every round-0 control and the target directories of the rev-12 P1 items; a directory is skipped when any new path is forbidden."""
    cands = _candidate_dirs(src_row, area, own_dirs if own_dirs is not None else _owned_dirs(p1_own, tree), taken, forbid_dirs, rej)
    names = [m.rpartition("/")[2] for m in src_row["members"]]
    for s in [stratum] + [x for x in (1, 2, 3) if x != stratum]:
        lst = cands[s]
        if not lst:
            continue
        start = int(_h("x2d-p1dir", src_row["cluster_id"]), 16) % len(lst)
        for k in range(len(lst)):
            d = lst[(start + k) % len(lst)]
            new = ["%s/%s" % (d, n) for n in names]
            if any(p in forbid_paths for p in new):
                if rej is not None:
                    rej["path"] += 1
                continue
            if any(p in tree or c0_state.sib.get(F.sibling_key(p), 0) > 0 for p in new):
                continue
            ent = [("A", 100, p, p, "1" * 40) for p in new]
            _, residual = F.owned_classify(ent, c0_state, p1_own)
            if len(residual) == len(new):
                if leak_check and area in (BL.b1(new, tree, p1_own)[1], BL.b2(new, tree, p1_own)[1]):
                    continue
                return new, d
    return None


def _p1_row(repo, r, a, placement, p1_own, c0_tree, settings, gate):
    """-> (control row, key extras) of one P1 item: the members of `r` leave A and re-enter as new files (same basenames and text) in the placement's directory (parent = C0)."""
    new, d = placement
    old_of = dict(zip(new, r["members"]))
    texts = {p: BU.read_text(repo, r["commit"], old_of[p]) for p in new}
    nd, stem = F.cluster_key(new[0])
    c = {"dir": nd, "stem": stem, "files": new, "role": F.role(new[0]), "scope": F.scope(new[0])}
    cp = [{"path": p, "status": "added"} for p in sorted(new)]
    ds = {"files": len(new), "added_lines": sum(len(t.splitlines()) for t in texts.values()), "deleted_lines": 0}
    subj, body = r["pr_title"], r["state"]["pr"]["body"]
    row = BU.make_unit(repo, 0, None, _h("x2d-ctl-p1", r["cluster_id"])[:16], r["commit"], x2d.C0, r["date"], subj, body, c, {p: ("added", "") for p in new}, {}, cp, ds, c0_tree,
                       p1_own, settings, "ctl", gate, True, reader=texts.get, symbols_from_text=True)
    return row, {"gold": {"covered": True, "acceptable": [a], "primary": a, "none_kind": None}, "source_case_id": r["case_id"], "catalogue": "catalogue-p1",
                 "original_paths": r["members"], "target_dir": d, "area_id": a}


def make_controls(repo, data=None, settings=None, reserve=None):
    """Pure: -> dict(units, control_key, contrast_units, catalogues {name: obj}, membership, manifest_runs {name: {rows, catalogue, out, arms, flip_all}})."""
    cat = x2d.read_catalogue(data)
    cont = json.load(open(os.path.join(data or x2d.DATA, "catalogue", "contrast-catalogue.json")))
    settings = settings or x2d.round_settings(0, data)
    mem = G.catalogue_membership(cat, 0)
    own = F.Ownership(mem)
    c0_tree = set(F.tree_paths(repo, x2d.C0))
    c0_state = F.TreeState(c0_tree)
    gate = BU.compute_core_gate(own, c0_tree)
    d_rows = BU.build_reserve_units(repo, "D", own, data, settings, gate, reserve)
    p_rows = BU.build_reserve_units(repo, "P", own, data, settings, gate, reserve)
    trees = {}

    def tree(rev):
        if rev not in trees:
            trees[rev] = set(F.tree_paths(repo, rev))
        return trees[rev]

    # ---- allocation of P units (P1 only from units that can be placed; placement is checked against the ownership the final P1 items use)
    elig_p = [(r, a) for r in p_rows for a in [_unique_owner(r, own)] if a]
    elig_d = [(r, a) for r in d_rows for a in [_unique_owner(r, own)] if a]

    def p1_ownership(p1_units):
        rm = {m for r, _ in p1_units for m in r["members"]}
        c1 = _filter_catalogue(cat, rm)
        return c1, F.Ownership(G.catalogue_membership(c1, 0))

    placeable, unplaceable, removed = [], [], set()
    for r, a in elig_p:
        c1 = _filter_catalogue(cat, removed | set(r["members"]))
        if a not in {x["area_id"] for x in c1["areas"]}:  # A would lose its last owned file: Jev could not recover an area that is not in catalogue-p1
            unplaceable.append((r, a))
            continue
        o = F.Ownership(G.catalogue_membership(c1, 0))
        if _p1_target(r, a, o, c0_tree, c0_state, [1, 2, 3][len(placeable) % 3]) is None:
            unplaceable.append((r, a))
        else:
            placeable.append((r, a))
            removed |= set(r["members"])
    if len(placeable) < 20:
        sys.exit("refusing: only %d P units can be placed for P1 (>= 20 needed; %d eligible, %d unplaceable); the gate cannot be built" % (len(placeable), len(elig_p), len(unplaceable)))
    p_slots = len(elig_p) - 20  # eligible P units beyond the 20 of P1
    need = min(P2_MAX_CLONES, p_slots + len(elig_d))  # R3-04 + rev 10: up to 12 clones so that the gate (>= 8 label-derived gap positives) has slack; P1 keeps >= 20; P leftovers first, then D
    if need < P2_MIN_GAP_POSITIVES:
        sys.exit("refusing: %d eligible P units and %d eligible D units leave only %d for P2 after the 20 P1 items (>= %d needed): the planted-gap gate cannot be built; stop, report, the council and the user decide (PREREGISTRATION 5)" % (
            len(elig_p), len(elig_d), max(need, 0), P2_MIN_GAP_POSITIVES))
    p_need = min(need, max(p_slots, 0))
    donors = list(unplaceable[:p_need])
    spare = max(0, min(p_need - len(donors), len(placeable) - 20))
    moved = placeable[len(placeable) - spare:] if spare else []
    p1 = placeable[: len(placeable) - spare]
    p2_p = (donors + moved)[:p_need]
    d_donors = elig_d[: need - len(p2_p)]
    p2 = p2_p + d_donors
    leftover_unplaceable = unplaceable[len(donors):]
    used_src = {r["case_id"] for r, _ in elig_p} | {r["case_id"] for r, _ in d_donors}
    items = []  # (kind, row, key_extra)

    # ---- P1: the members leave A, re-enter as new files elsewhere; drop items that stop placing under the final removal set until stable
    while True:
        cat_p1, p1_own = p1_ownership(p1)
        placements = [_p1_target(r, a, p1_own, c0_tree, c0_state, [1, 2, 3][i % 3]) for i, (r, a) in enumerate(p1)]
        bad = [i for i, t in enumerate(placements) if t is None or p1[i][1] not in {x["area_id"] for x in cat_p1["areas"]}]
        if not bad:
            break
        for i in reversed(bad):
            leftover_unplaceable.append(p1.pop(i))
        if len(p1) < 20:
            sys.exit("refusing: only %d P1 items remain placeable under the final catalogue (>= 20 needed)" % len(p1))
    for i, (r, a) in enumerate(p1):
        row, extra = _p1_row(repo, r, a, placements[i], p1_own, c0_tree, settings, gate)
        items.append(("p1", row, extra))
    # ---- P2: clones whose owning area is ablated
    ablated = sorted({a for _, a in p2})
    cat_p2 = _filter_catalogue(cat, (), ablated)
    p2_own = F.Ownership(G.catalogue_membership(cat_p2, 0))
    gap_positives = 0
    for r, a in p2:
        row = BU.restratify(_clone(r, None, _h("x2d-ctl-p2", r["cluster_id"])[:16]), tree(r["parent"]), p2_own, gate)
        gap = all(not p2_own.owners(m) for m in row["members"])
        gap_positives += gap
        items.append(("p2", row, {"gold": None, "gold_rule": "derive_ablation", "source_case_id": r["case_id"], "catalogue": "catalogue-p2", "ablated_area_id": a,
                                  "ablated_area_ids": ablated, "gap_positive": bool(gap)}))
    if gap_positives < P2_MIN_GAP_POSITIVES:
        sys.exit("refusing: only %d P2 clones are gap positives (no member owned in catalogue-p2; >= %d needed)" % (gap_positives, P2_MIN_GAP_POSITIVES))
    # ---- planted new: synthetic packages
    for i, t in enumerate(PLANTED):
        p = t["path"]
        nd, stem = F.cluster_key(p)
        c = {"dir": nd, "stem": stem, "files": [p], "role": F.role(p), "scope": F.scope(p)}
        sha = _h("x2d-planted", p)[:40]
        lines = t["code"].splitlines()
        ds = {"files": 1, "added_lines": len(lines), "deleted_lines": 0}
        row = BU.make_unit(repo, 0, None, _h("x2d-ctl-new", p)[:16], sha, x2d.C0, "2024-%02d-1%dT10:%02d:00+00:00" % (3 + i, i, 10 + i), t["title"], t["body"], c, {p: ("added", "")}, {},
                           [{"path": p, "status": "added"}], ds, c0_tree, own, settings, "ctl", gate, True, reader={p: t["code"]}.get, symbols_from_text=True)
        items.append(("planted_new", row, {"gold": {"covered": False, "acceptable": [], "primary": None, "none_kind": "new"}, "source_case_id": None, "catalogue": "catalogue"}))
    # ---- empty-file rename trap (#9268 shape)
    new_p, old_p = "docs_src/app_testing/app_b_an/__init__.py", "tests/test_tutorial/test_annotated/__init__.py"
    nd, stem = F.cluster_key(new_p)
    c = {"dir": nd, "stem": stem, "files": [new_p], "role": F.role(new_p), "scope": F.scope(new_p)}
    row = BU.make_unit(repo, 0, None, _h("x2d-ctl-trap", new_p)[:16], _h("x2d-trap", new_p)[:40], x2d.C0, "2024-05-14T09:30:00+00:00", "♻️ Move the Annotated app testing example next to the docs (#9268)",
                       "Moves the package marker of the Annotated testing example so that tests and docs examples live next to each other.", c, {new_p: ("renamed", "empty-trap from " + old_p)},
                       {new_p: old_p}, [{"path": new_p, "status": "renamed"}], {"files": 1, "added_lines": 0, "deleted_lines": 0}, c0_tree, own, settings, "ctl", gate, True,
                       reader={new_p: ""}.get, symbols_from_text=True)
    items.append(("trap", row, {"gold": {"covered": False, "acceptable": [], "primary": None, "none_kind": None, "rule": "R10"}, "source_case_id": None, "catalogue": "catalogue"}))
    # ---- misleading PR prose
    rest = [r for r in p_rows if r["case_id"] not in used_src] or [r for r in p_rows if r["case_id"] != p_rows[0]["case_id"]]
    src = rest[0]
    decoy, title, body = BU.prose_title(cat, _unique_owner(src, own))
    row = _clone(src, None, _h("x2d-ctl-prose", src["cluster_id"])[:16])
    row["pr_title"], row["pr_body"] = title, body[:1500]
    row["state"]["pr"]["title"], row["state"]["pr"]["body"] = title, body[:1500]
    row["label_evidence"]["pr"]["title"], row["label_evidence"]["pr"]["body"] = title, body[:6000]
    row["pr"], row["pr_key"] = None, row["commit"][:8]
    items.append(("prose", row, {"gold": None, "source_case_id": src["case_id"], "catalogue": "catalogue", "decoy_area_id": decoy}))
    # ---- duplicates (test-retest), 10 % of the round-0 batch
    n_dup = max(1, math.ceil(0.10 * (len(d_rows) + len(p_rows) + len(items))))
    pool = sorted(d_rows + p_rows, key=lambda r: _h("x2d-dup", r["case_id"]))
    for r in pool[:n_dup]:
        items.append(("dup", _clone(r, None, _h("x2d-ctl-dup", r["cluster_id"])[:16]), {"gold": None, "source_case_id": r["case_id"], "catalogue": "catalogue"}))
    # ---- ids in a hash order
    items.sort(key=lambda it: _h("x2d-ctl-order", it[0] + "|" + it[1]["cluster_id"]))
    controls, key = [], []
    for n, (kind, row, extra) in enumerate(items, 1):
        row["case_id"] = "0:ctl-%d" % n
        row["control_kind"], row["source_case_id"] = kind, extra.get("source_case_id")
        controls.append(row)
        key.append(dict({"case_id": row["case_id"], "kind": kind}, **extra))
    units = sorted(d_rows + p_rows + controls, key=lambda r: _h("x2d-order", r["case_id"]))
    cown = F.Ownership(G.catalogue_membership(cont, 0))
    cgate = BU.compute_core_gate(cown, c0_tree)
    contrast = [BU.restratify(r, tree(r["parent"]), cown, cgate) for r in d_rows]
    kinds = {k["case_id"]: k["kind"] for k in key}
    by = lambda *ks: [u for u in controls if kinds[u["case_id"]] in ks]
    runs = {
        "main": {"rows": d_rows + p_rows + by("dup", "trap", "prose", "planted_new"), "catalogue": "catalogue.json", "out": "jev/scores.jsonl", "arms": "sem,sib,title,fan", "flip_all": True},
        "n1": {"rows": d_rows + p_rows, "catalogue": "catalogue-n1.json", "out": "jev-n1/scores.jsonl", "arms": "sem", "flip_all": False},
        "p1": {"rows": by("p1"), "catalogue": "catalogue-p1.json", "out": "jev-p/scores.p1.jsonl", "arms": "sem", "flip_all": False},
        "p2": {"rows": by("p2"), "catalogue": "catalogue-p2.json", "out": "jev-p/scores.p2.jsonl", "arms": "sem", "flip_all": False},
    }
    return {"units": units, "control_key": key, "contrast_units": contrast, "catalogues": {"catalogue-p1": cat_p1, "catalogue-p2": cat_p2, "catalogue-n1": _n1_catalogue(cat)},
            "membership": mem, "runs": runs, "p1_unplaceable": len(unplaceable) + len(leftover_unplaceable) - (len(unplaceable) - len(donors)), "eligible": len(elig_p), "eligible_d": len(elig_d),
            "p2_donors": {"P_unplaceable": len(donors), "P_leftover": len(moved), "D": len(d_donors)}, "p2_source_ids": [r["case_id"] for r, _ in p2], "ablated_area_ids": ablated}


# ---------------------------------------------------------------- rev 13 (R-053): the fresh-gate build (the one permitted re-gate of P1 and P2)
def _guide_prs(data=None):
    """The PR numbers the round-0 guide names (every `#<digits>` of its worked examples): their PRs are exposed to the labellers and never fresh."""
    p = x2d.guide_path(0, data)
    if not os.path.isfile(p):
        p = os.path.join(x2d.HERE, "GUIDE.md")
    with open(p) as f:
        return sorted(set(re.findall(r"#(\d{2,6})\b", f.read())))


def exposed(data=None):
    """Membership only (units.jsonl and control-key.jsonl of round 0; no label, no Jev score is read): what the rev-12 P1 / P2 gates exposed. -> {"sources": case ids of the P1 items' and P2 clones'
    source units, "paths": their members and the P1 items' original paths, "prs": their PRs (the PR episodes), "guide_prs": the PRs of the guide's worked examples}."""
    rd = x2d.round_dir(0, x2d.main_data(data))
    urows = x2d.read_jsonl(os.path.join(rd, "units.jsonl"))
    units = {u["case_id"]: u for u in urows}
    key = x2d.read_jsonl(os.path.join(rd, "control-key.jsonl"))
    sources = sorted({k["source_case_id"] for k in key if k.get("kind") in ("p1", "p2") and k.get("source_case_id")})
    paths, prs = set(), set()
    for c in sources:
        paths |= set(units[c]["members"])
        prs.add(str(units[c]["pr_key"]))
    for k in key:
        if k.get("kind") == "p1":
            paths |= set(k.get("original_paths") or [])
    presented = set()  # R17-01: the paths at which the round-0 controls were presented to the labellers and Jev
    for u in urows:
        if u.get("control"):
            presented |= set(u["members"])
    target_dirs = {k["target_dir"] for k in key if k.get("kind") == "p1" and k.get("target_dir")}  # the rev-12 P1 target directories
    return {"sources": sources, "paths": paths, "prs": prs, "guide_prs": set(_guide_prs(data)), "presented": presented, "target_dirs": target_dirs}


def _add_commits(repo, wanted):
    """{path: sha} of the most recent first-parent commit up to C0 that ADDED the path (no rename detection: a rename's target counts as an addition); only paths in `wanted`."""
    out = F.git(repo, "-c", "core.quotepath=off", "log", "--first-parent", "--no-renames", "--diff-filter=A", "--name-only", "--format=%x01%H", x2d.C0).decode("utf-8", "surrogateescape")
    adds, sha = {}, None
    for line in out.split("\n"):
        if line.startswith("\x01"):
            sha = line[1:].strip()
        elif line and sha and line in wanted and line not in adds:  # newest first: the first sighting is the most recent addition
            adds[line] = sha
    return adds


def fresh_pool(repo, data=None, reserve=None):
    """Membership only, deterministic, no model, no label: the fresh candidates of the re-gate (module docstring). -> {"pool": [candidate dicts in hash order, one per PR, no shared path], "funnel": {origin: {reason: n}},
    "exposed": {...counts}}. A candidate: {cluster_id, case_id, commit, parent, date, pr, pr_key, dir, stem, members, role, scope, area_id, origin, pr_title}."""
    data = x2d.main_data(data)
    cat = x2d.read_catalogue(data)
    own = F.Ownership(G.catalogue_membership(cat, 0))
    c0_tree = set(F.tree_paths(repo, x2d.C0))
    reserve = reserve or json.load(open(os.path.join(x2d.HERE, "prepass", "reserve.json")))
    rset = {c["id"]: c["set"] for c in reserve["clusters_hash_order"]}
    ex = exposed(data)
    src_ids = set(ex["sources"])
    elig = {p for p in c0_tree if len(own.owners(p)) == 1 and F.excluded_reason(p) is None}
    adds = _add_commits(repo, elig)
    shas = sorted(set(adds.values()) | {c["sha"] for c in reserve["clusters_hash_order"]})
    funnel = collections.defaultdict(collections.Counter)
    cands = []
    for sha in shas:
        pp = F.git(repo, "rev-list", "--parents", "-n", "1", sha).decode().split()
        if len(pp) < 2:  # the root commit has no parent tree to diff against
            continue
        parent = pp[1]
        entries = F.diff_entries(repo, parent, sha)
        state = F.TreeState(F.tree_paths(repo, parent))
        _, residual = F.classify(entries, state)
        live, _inert = F.clusters(residual, {e[3]: e[4] for e in entries if e[0] in "ARC"})
        info = None
        for c in live:
            cid = F.cluster_id(sha, c)
            origin = rset.get(cid) or "in_no_unit"
            fn = funnel[origin]
            fn["clusters"] += 1
            mem = c["files"]
            if not all(m in c0_tree for m in mem):
                fn["not_all_files_at_C0"] += 1
                continue
            owners = [own.owners(m) for m in mem]
            if any(len(o) != 1 for o in owners) or len(set().union(*owners)) != 1:
                fn["not_one_owner_area"] += 1
                continue
            if info is None:
                subject, _body = BU.commit_message(repo, sha)
                info = (subject, F.pr_number(subject), F.git(repo, "show", "-s", "--format=%cI", sha).decode().strip())
            if info[1] is None:
                fn["no_pr_number"] += 1
                continue
            if x2d.case_id(0, cid) in src_ids:
                fn["excluded_used_as_source"] += 1
            elif set(mem) & ex["paths"]:
                fn["excluded_path_twin"] += 1
            elif str(info[1]) in ex["prs"]:
                fn["excluded_pr_episode"] += 1
            elif str(info[1]) in ex["guide_prs"]:
                fn["excluded_guide_example_pr"] += 1
            else:
                fn["eligible_before_dedup"] += 1
                cands.append({"cluster_id": cid, "case_id": x2d.case_id(0, cid), "commit": sha, "parent": parent, "date": info[2], "pr": info[1], "pr_key": str(info[1]), "dir": c["dir"], "stem": c["stem"],
                              "members": list(mem), "role": c["role"], "scope": c["scope"], "area_id": next(iter(owners[0])), "origin": origin, "pr_title": info[0]})
    cands.sort(key=lambda r: _h("x2d-regate-order", r["cluster_id"]))
    pool, prs_seen, paths_seen = [], set(), set()
    for r in cands:
        fn = funnel[r["origin"]]
        if r["pr_key"] in prs_seen:
            fn["dropped_same_pr_as_earlier"] += 1
        elif set(r["members"]) & paths_seen:
            fn["dropped_shared_path_with_earlier"] += 1
        else:
            pool.append(r)
            prs_seen.add(r["pr_key"])
            paths_seen |= set(r["members"])
            fn["pool"] += 1
    return {"pool": pool, "funnel": {o: dict(sorted(c.items())) for o, c in sorted(funnel.items())},
            "exposed": {"source_units": len(ex["sources"]), "paths": len(ex["paths"]), "pr_episodes": sorted(ex["prs"]), "guide_prs": sorted(ex["guide_prs"])}}


def _p1_placements(trial, cat, c0_tree, c0_state, gate, forbid_paths=(), forbid_dirs=(), rej=None):
    """-> (placements [(new paths, directory)], catalogue-p1, its ownership, strata) when every item of `trial` places under the removal set of the whole trial (its area keeps an owned file, a target
    directory of the wanted kind exists, rules (a)-(f) do not resolve the new path, B1 and B2 do not answer A), else None. Target stratum of item i: x2d.REGATE_P1_STRATA[i mod 5]; a directory that hosted a stratum-1 item is never reused.
    R17-01: `taken` starts as the rev-12 P1 target directories (`forbid_dirs`), skipped for every stratum, and a directory whose new path is a presented path of a round-0 control (`forbid_paths`) is skipped."""
    removed = {m for r in trial for m in r["members"]}
    c1 = _filter_catalogue(cat, removed)
    ids = {x["area_id"] for x in c1["areas"]}
    o = F.Ownership(G.catalogue_membership(c1, 0))
    own_dirs = _owned_dirs(o, c0_tree)
    out, taken, strata = [], set(forbid_dirs), []
    for i, r in enumerate(trial):
        if r["area_id"] not in ids:
            return None
        t = _p1_target(r, r["area_id"], o, c0_tree, c0_state, x2d.REGATE_P1_STRATA[i % len(x2d.REGATE_P1_STRATA)], leak_check=True, taken=taken, own_dirs=own_dirs,
                       forbid_paths=forbid_paths, forbid_dirs=forbid_dirs, rej=rej)
        if t is None:
            return None
        st = BU.stratify(t[0], r["role"], c0_tree, o, gate)[0]
        if st == 1:  # a stratum-1 directory carries ONE other area's cue for one item only; the flat directories of strata 2 and 3 can host several
            taken.add(t[1])
        out.append(t)
        strata.append(st)
    return out, c1, o, strata


def regate_plan(repo, data=None, reserve=None, info=None):
    """Membership only: the planned fresh build. -> dict(info, p1 [(candidate, area, placement, stratum)], p2 [(candidate, area)], ablated, supply (the recorded table), problems [why the build must refuse])."""
    data = x2d.main_data(data)
    info = info or fresh_pool(repo, data, reserve)
    cat = x2d.read_catalogue(data)
    own = F.Ownership(G.catalogue_membership(cat, 0))
    c0_tree = set(F.tree_paths(repo, x2d.C0))
    c0_state = F.TreeState(c0_tree)
    gate = BU.compute_core_gate(own, c0_tree)
    ex = exposed(data)
    fp, fd = ex["presented"], ex["target_dirs"]
    acc, rest, res = [], [], None
    for r in info["pool"]:
        if len(acc) >= x2d.REGATE_P1_N:
            rest.append(r)
            continue
        got = _p1_placements(acc + [r], cat, c0_tree, c0_state, gate, fp, fd)
        if got is None:
            rest.append(r)
        else:
            acc, res = acc + [r], got
    rej = collections.Counter()
    if acc:
        _p1_placements(acc, cat, c0_tree, c0_state, gate, fp, fd, rej)  # the final trial again, counting what the forbidden sets rejected
    problems = []
    p1 = [(r, r["area_id"], res[0][i], res[3][i]) for i, r in enumerate(acc)] if res else []
    strata = collections.Counter(s for *_, s in p1)
    if len(p1) < x2d.REGATE_P1_N:
        problems.append("only %d fresh P1 items can be placed (>= %d needed)" % (len(p1), x2d.REGATE_P1_N))
    if strata[1] < x2d.REGATE_P1_STRATUM1_MIN:
        problems.append("only %d of the P1 items are stratum 1 (>= %d needed: at least half)" % (strata[1], x2d.REGATE_P1_STRATUM1_MIN))
    by_area = collections.defaultdict(list)
    for r in rest:
        by_area[r["area_id"]].append(r)
    qual = sorted((a for a, v in by_area.items() if len(v) >= x2d.REGATE_P2_AREA_MIN), key=lambda a: _h("x2d-regate-p2area", a))
    p2, chosen = [], []
    for a in qual:  # areas in hash order, up to REGATE_P2_PER_AREA clones each (candidates in pool order), until REGATE_P2_CLONES clones
        if len(p2) >= x2d.REGATE_P2_CLONES:
            break
        take = by_area[a][: min(x2d.REGATE_P2_PER_AREA, x2d.REGATE_P2_CLONES - len(p2))]
        p2 += [(r, a) for r in take]
        chosen.append(a)
    chosen.sort()
    if len(p2) < x2d.REGATE_P2_CLONES:
        problems.append("only %d fresh P2 clones can be built from the %d areas that own >= %d candidates left after P1 (%d wanted, at most %d per area)" % (
            len(p2), len(qual), x2d.REGATE_P2_AREA_MIN, x2d.REGATE_P2_CLONES, x2d.REGATE_P2_PER_AREA))
    own_dirs = _owned_dirs(own, c0_tree)  # hostability over the FULL catalogue: a pool candidate can host a stratum when a directory of that kind exists for it (before the removal sets and the leak check)
    host = collections.Counter()
    for r in info["pool"]:
        cd = _candidate_dirs(r, r["area_id"], own_dirs, (), fd)
        for s in (1, 2, 3):
            host[s] += bool(cd[s])
    pool = info["pool"]
    short = lambda a: a[:6]
    supply = {
        "gate_set": x2d.REGATE_GATE_SET, "membership_only": True,
        "exposed": dict(info["exposed"], presented_paths=len(fp), rev12_p1_target_dirs=len(fd), p1_rejected_for_presented_path=rej["path"], p1_directory_candidates_rejected_as_rev12_target=rej["dir"]), "funnel": info["funnel"],
        "pool": {"clusters": len(pool), "prs": len({r["pr_key"] for r in pool}), "areas": len({r["area_id"] for r in pool}), "by_origin": dict(sorted(collections.Counter(r["origin"] for r in pool).items())),
                 "can_host_stratum": {str(s): host[s] for s in (1, 2, 3)}},
        "p1": {"n": len(p1), "wanted": x2d.REGATE_P1_N, "stratum1_min": x2d.REGATE_P1_STRATUM1_MIN, "strata": {str(s): strata[s] for s in (1, 2, 3)}, "target_pattern": list(x2d.REGATE_P1_STRATA),
               "items": [{"cluster_id": r["cluster_id"], "origin": r["origin"], "pr": r["pr"], "date": r["date"], "area": short(a), "members": r["members"], "target_dir": pl[1], "stratum": s} for r, a, pl, s in p1]},
        "p2": {"areas_qualifying": len(qual), "clones_wanted": x2d.REGATE_P2_CLONES, "per_area_max": x2d.REGATE_P2_PER_AREA, "clones": len(p2), "ablated_areas": [short(a) for a in chosen],
               "by_origin": dict(sorted(collections.Counter(r["origin"] for r, _ in p2).items())),
               "items": [{"cluster_id": r["cluster_id"], "origin": r["origin"], "pr": r["pr"], "date": r["date"], "area": short(a), "members": r["members"]} for r, a in p2]},
        "problems": problems,
    }
    return {"info": info, "p1": p1, "p2": p2, "ablated": chosen, "supply": supply, "problems": problems, "gate": gate}


def make_regate(repo, data=None, settings=None, reserve=None, plan=None):
    """Pure: the fresh build. -> dict(units, control_key, catalogues {catalogue-p1, catalogue-p2}, membership, runs {p1, p2: {rows, catalogue, out, arms, flip_all}}, supply, p2_source_ids, ablated_area_ids).
    Refuses (SystemExit, nothing written) when the plan has problems: the supply cannot meet the minimums and they are never shrunk."""
    data = x2d.main_data(data)
    cat = x2d.read_catalogue(data)
    settings = settings or x2d.round_settings(0, data)
    plan = plan or regate_plan(repo, data, reserve)
    if plan["problems"]:
        sys.exit("refusing: the fresh supply cannot meet the re-gate minimums (they are not shrunk; stop, report, the council and the user decide): " + "; ".join(plan["problems"]))
    mem = G.catalogue_membership(cat, 0)
    own = F.Ownership(mem)
    c0_tree = set(F.tree_paths(repo, x2d.C0))
    gate = plan["gate"]
    trees = {}

    def tree(rev):
        if rev not in trees:
            trees[rev] = set(F.tree_paths(repo, rev))
        return trees[rev]

    srcs = [r for r, *_ in plan["p1"]] + [r for r, _ in plan["p2"]]
    synth = {"clusters_hash_order": [{"set": "F", "sha": r["commit"], "id": r["cluster_id"], "date": r["date"]} for r in srcs]}
    full = {x["cluster_id"]: x for x in BU.build_reserve_units(repo, "F", own, data, settings, gate, synth)}
    p1_cands = [r for r, *_ in plan["p1"]]
    ex = exposed(data)
    _pl, cat_p1, p1_own, strata = _p1_placements(p1_cands, cat, c0_tree, F.TreeState(c0_tree), gate, ex["presented"], ex["target_dirs"])
    items = []
    for i, (r, a, pl, _s) in enumerate(plan["p1"]):
        row, extra = _p1_row(repo, full[r["cluster_id"]], a, pl, p1_own, c0_tree, settings, gate)
        nd = F.norm_dir(pl[1])
        sib_areas = sorted({x for f in p1_own.dir_files(nd) if f in c0_tree for x in p1_own.owners(f)})
        extra.update(gold_rule="labelled_acceptable_set_must_contain_A", target_stratum=x2d.REGATE_P1_STRATA[i % len(x2d.REGATE_P1_STRATA)], achieved_stratum=row["stratum"], leak_checked=True,
                     sibling_area_ids=sib_areas, origin=r["origin"])
        items.append(("p1", row, extra))
    ablated = list(plan["ablated"])
    cat_p2 = _filter_catalogue(cat, (), ablated)
    p2_own = F.Ownership(G.catalogue_membership(cat_p2, 0))
    p2_sources = []
    for r, a in plan["p2"]:
        src = full[r["cluster_id"]]
        p2_sources.append(src)
        row = BU.restratify(_clone(src, None, _h("x2d-ctl-p2", src["cluster_id"])[:16]), tree(src["parent"]), p2_own, gate)
        gap = all(not p2_own.owners(m) for m in row["members"])
        items.append(("p2", row, {"gold": None, "gold_rule": "direct_label", "source_case_id": src["case_id"], "catalogue": "catalogue-p2", "ablated_area_id": a, "ablated_area_ids": ablated,
                                  "gap_positive": bool(gap), "origin": r["origin"]}))
    items.sort(key=lambda it: _h("x2d-ctl-order", it[0] + "|" + it[1]["cluster_id"]))
    controls, key = [], []
    for n, (kind, row, extra) in enumerate(items, 1):
        row["case_id"] = "0:ctl-%d" % n
        row["control_kind"], row["source_case_id"] = kind, extra.get("source_case_id")
        controls.append(row)
        key.append(dict({"case_id": row["case_id"], "kind": kind}, **extra))
    units = sorted(p2_sources + controls, key=lambda r: _h("x2d-order", r["case_id"]))
    kinds = {k["case_id"]: k["kind"] for k in key}
    by = lambda k: [u for u in controls if kinds[u["case_id"]] == k]
    runs = {"p1": {"rows": by("p1"), "catalogue": "catalogue-p1.json", "out": "jev-p/scores.p1.jsonl", "arms": "sem", "flip_all": False},
            "p2": {"rows": by("p2"), "catalogue": "catalogue-p2.json", "out": "jev-p/scores.p2.jsonl", "arms": "sem", "flip_all": False}}
    return {"units": units, "control_key": key, "catalogues": {"catalogue-p1": cat_p1, "catalogue-p2": cat_p2}, "membership": mem, "runs": runs, "supply": plan["supply"],
            "p2_source_ids": [s["case_id"] for s in p2_sources], "ablated_area_ids": ablated}


def regate_precondition(data=None):
    """-> problems that forbid the re-gate: the rev-12 round-0 gates.json must exist and record a FAIL of P1 and / or P2 and of no other gate; the frozen Jev side (inputs, settings, thresholds,
    catalogue) must exist."""
    main = x2d.main_data(data)
    rd = x2d.round_dir(0, main)
    gp = os.path.join(rd, "gates.json")
    bad = []
    if not os.path.isfile(gp) or os.path.getsize(gp) == 0:
        return ["%s missing: the re-gate follows a round-0 gate stop" % gp]
    failed = sorted(k for k, v in (json.load(open(gp)).get("gates") or {}).items() if v.get("status") == "FAIL")
    if not failed:
        bad.append("no round-0 gate FAILed: there is nothing to re-gate")
    elif set(failed) - {"P1", "P2"}:
        bad.append("a gate other than P1 / P2 FAILed (%s): that stops the pilot, it is not re-gated" % ", ".join(failed))
    for rel in ("inputs/GUIDE.md", "settings.json", "thresholds.json"):
        if not os.path.isfile(os.path.join(rd, rel)):
            bad.append("%s missing" % os.path.join(rd, rel))
    if not os.path.isdir(os.path.join(rd, "inputs", "questions")):
        bad.append("%s missing" % os.path.join(rd, "inputs", "questions"))
    if not os.path.isfile(x2d.catalogue_path(main)):
        bad.append("%s missing" % x2d.catalogue_path(main))
    return bad


def write_regate(res, data=None):
    """Write D/rounds/0/regate/ once (the only directory this stage touches). Refuses when it exists. The frozen Jev side (the round's inputs, settings, thresholds and the catalogue) is copied byte for byte;
    x2d.regate_frozen_problems compares the copies with the originals at every later stage."""
    main = x2d.main_data(data)
    root, rd0 = x2d.regate_root(main), x2d.round_dir(0, main)
    x2d.refuse_if_exists(root)
    os.makedirs(root)
    rr = os.path.join(root, "rounds", "0")
    marker = {"utc": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()), "gate_set": x2d.REGATE_GATE_SET, "rev12_gates_sha256": x2d.sha256_file(os.path.join(rd0, "gates.json")),
              "one_shot": "a second build, a second analysis and any stage after the analysis refuse; a re-gate that does not PASS stops the pilot"}
    x2d.write_json_new(os.path.join(root, "regate.json"), marker)
    os.makedirs(os.path.join(root, "catalogue"))
    for fn in ("catalogue.json", "catalogue.sha256"):
        if os.path.isfile(os.path.join(main, "catalogue", fn)):
            shutil.copyfile(os.path.join(main, "catalogue", fn), os.path.join(root, "catalogue", fn))
    shutil.copytree(os.path.join(rd0, "inputs"), os.path.join(rr, "inputs"))
    for fn in ("settings.json", "thresholds.json"):
        shutil.copyfile(os.path.join(rd0, fn), os.path.join(rr, fn))
    x2d.write_jsonl_new(os.path.join(rr, "units.jsonl"), res["units"])
    x2d.write_jsonl_new(os.path.join(rr, "control-key.jsonl"), res["control_key"])
    x2d.write_json_new(os.path.join(rr, "membership.json"), res["membership"])
    for n in ("catalogue-p1", "catalogue-p2"):
        x2d.write_json_new(os.path.join(rr, n + ".json"), res["catalogues"][n])
    jin, manifest = os.path.join(rr, "jev-in"), []
    for name, run in res["runs"].items():
        up = os.path.join(jin, name + ".units.jsonl")
        x2d.write_jsonl_new(up, run["rows"])
        manifest.append({"name": name, "units": os.path.abspath(up), "catalogue": os.path.abspath(os.path.join(rr, run["catalogue"])), "out": run["out"], "arms": run["arms"], "flip_all": run["flip_all"]})
    x2d.write_json_new(os.path.join(jin, "manifest.json"), manifest)
    x2d.write_json_new(os.path.join(root, "supply.json"), res["supply"])
    return root


def _print_supply(s):
    print("fresh pool: %d clusters from %d PRs in %d areas, by origin %s; can host stratum %s" % (s["pool"]["clusters"], s["pool"]["prs"], s["pool"]["areas"], s["pool"]["by_origin"], s["pool"]["can_host_stratum"]))
    print("funnel (clusters seen -> pool) per origin: " + json.dumps(s["funnel"], sort_keys=True))
    print("P1: %d of %d items, strata %s" % (s["p1"]["n"], s["p1"]["wanted"], s["p1"]["strata"]))
    print("P2: %d of %d clones in %d ablated areas (%d areas own >= %d candidates, at most %d clones per area), by origin %s" % (s["p2"]["clones"], s["p2"]["clones_wanted"], len(s["p2"]["ablated_areas"]),
                                                                                                                              s["p2"]["areas_qualifying"], x2d.REGATE_P2_AREA_MIN, s["p2"]["per_area_max"], s["p2"]["by_origin"]))
    print("problems: %s" % (s["problems"] or "none"))


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--repo", required=True)
    ap.add_argument("--data", default=None)
    ap.add_argument("--regate-supply", action="store_true", help="rev 13: the fresh pool and the planned re-gate build, membership only; writes nothing under the data directory (--out FILE writes the record there)")
    ap.add_argument("--regate", action="store_true", help="rev 13: build the one permitted re-gate (D/rounds/0/regate/); refuses when it exists")
    ap.add_argument("--out", default=None)
    a = ap.parse_args(argv)
    bad = x2d.verify_pins(require_signoff=x2d.require_signoff())
    if bad:
        sys.exit("pins: " + "; ".join(bad))
    data = a.data
    if a.regate_supply or a.regate:
        if x2d.is_regate_root(data):
            sys.exit("refusing: the re-gate is built from the study data directory, not from inside D/rounds/0/regate")
        main_ = x2d.main_data(data)
        if a.regate:
            pre = regate_precondition(main_)
            if pre:
                sys.exit("refusing: " + "; ".join(pre))
            x2d.refuse_if_exists(x2d.regate_root(main_))
        plan = regate_plan(a.repo, main_)
        _print_supply(plan["supply"])
        if a.regate_supply:
            if a.out:
                x2d.write_json_new(a.out, plan["supply"])
            return
        res = make_regate(a.repo, main_, plan=plan)
        print("wrote", write_regate(res, main_))
        return
    if x2d.is_regate_root(data):
        sys.exit("refusing: the round-0 build does not run inside D/rounds/0/regate")
    rd = x2d.round_dir(0, data)
    jin = os.path.join(rd, "jev-in")
    outs = [os.path.join(rd, n) for n in ("units.jsonl", "control-key.jsonl", "units.contrast.jsonl", "membership.json", "catalogue-p1.json", "catalogue-p2.json", "catalogue-n1.json")]
    x2d.refuse_if_exists(*outs, os.path.join(jin, "manifest.json"))
    res = make_controls(a.repo, data)
    x2d.write_jsonl_new(outs[0], res["units"])
    x2d.write_jsonl_new(outs[1], res["control_key"])
    x2d.write_jsonl_new(outs[2], res["contrast_units"])
    x2d.write_json_new(outs[3], res["membership"])
    for n, p in zip(("catalogue-p1", "catalogue-p2", "catalogue-n1"), outs[4:]):
        x2d.write_json_new(p, res["catalogues"][n])
    manifest = []
    for name, run in res["runs"].items():
        up = os.path.join(jin, name + ".units.jsonl")
        x2d.write_jsonl_new(up, run["rows"])
        cpath = os.path.join(data or x2d.DATA, "catalogue", run["catalogue"]) if run["catalogue"] == "catalogue.json" else os.path.join(rd, run["catalogue"])
        manifest.append({"name": name, "units": os.path.abspath(up), "catalogue": os.path.abspath(cpath), "out": run["out"], "arms": run["arms"], "flip_all": run["flip_all"]})
    x2d.write_json_new(os.path.join(jin, "manifest.json"), manifest)
    kinds = collections.Counter(k["kind"] for k in res["control_key"])
    print(json.dumps({"units": len(res["units"]), "controls": dict(kinds), "p1_unplaceable": res["p1_unplaceable"], "eligible": res["eligible"], "eligible_d": res["eligible_d"], "p2_donors": res["p2_donors"],
                      "ablated_areas": len(res["ablated_area_ids"])}))


if __name__ == "__main__":
    main()
