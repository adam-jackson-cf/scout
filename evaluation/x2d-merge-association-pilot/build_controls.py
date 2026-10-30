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
Importable: make_controls.
"""
import argparse, collections, copy, hashlib, json, math, os, sys

P2_MIN_GAP_POSITIVES = 8  # PREREGISTRATION 5: the P2 gate's denominator (clones whose derived gold is none/gap) must be at least this; fewer label-derived gap positives = UNMEASURED (R3-04)
P2_MAX_CLONES = 12  # clones built: n_p2 = min(12, eligible_P - 20 + eligible_D), refusal below P2_MIN_GAP_POSITIVES (R3-04; D top-up, rev 10)

import x2d
import frontend as F
import gold as G
import build_units as BU

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


def _p1_target(src_row, area, p1_own, tree, c0_state, stratum):
    """-> (new paths, directory) placing the members in a directory of the wanted kind; None when impossible."""
    role, scope = src_row["role"], src_row["scope"]
    own_dirs = collections.defaultdict(set)
    for f in tree:
        own_dirs[f.rpartition("/")[0]] |= p1_own.owners(f)
    cands = {1: [], 2: [], 3: []}
    for d in sorted(own_dirs):
        if not d or "__pycache__" in d or F.norm_dir(d) == src_row["dir"]:
            continue
        if F.role(d + "/x") != role or F.scope(d + "/x") != scope:
            continue
        o = own_dirs[d]
        if area in o:
            continue
        cands[1 if len(o) == 1 else 2 if len(o) >= 2 else 3].append(d)
    base = [F.split_name(m)[0] for m in src_row["members"]]
    names = [m.rpartition("/")[2] for m in src_row["members"]]
    for s in [stratum] + [x for x in (1, 2, 3) if x != stratum]:
        lst = cands[s]
        if not lst:
            continue
        start = int(_h("x2d-p1dir", src_row["cluster_id"]), 16) % len(lst)
        for k in range(len(lst)):
            d = lst[(start + k) % len(lst)]
            new = ["%s/%s" % (d, n) for n in names]
            if any(p in tree or c0_state.sib.get(F.sibling_key(p), 0) > 0 for p in new):
                continue
            ent = [("A", 100, p, p, "1" * 40) for p in new]
            _, residual = F.owned_classify(ent, c0_state, p1_own)
            if len(residual) == len(new):
                return new, d
    return None


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
        new, d = placements[i]
        old_of = dict(zip(new, r["members"]))
        texts = {p: BU.read_text(repo, r["commit"], old_of[p]) for p in new}
        nd, stem = F.cluster_key(new[0])
        c = {"dir": nd, "stem": stem, "files": new, "role": F.role(new[0]), "scope": F.scope(new[0])}
        cp = [{"path": p, "status": "added"} for p in sorted(new)]
        ds = {"files": len(new), "added_lines": sum(len(t.splitlines()) for t in texts.values()), "deleted_lines": 0}
        subj, body = r["pr_title"], r["state"]["pr"]["body"]
        row = BU.make_unit(repo, 0, None, _h("x2d-ctl-p1", r["cluster_id"])[:16], r["commit"], x2d.C0, r["date"], subj, body, c, {p: ("added", "") for p in new}, {}, cp, ds, c0_tree,
                           p1_own, settings, "ctl", gate, True, reader=texts.get, symbols_from_text=True)
        items.append(("p1", row, {"gold": {"covered": True, "acceptable": [a], "primary": a, "none_kind": None}, "source_case_id": r["case_id"], "catalogue": "catalogue-p1",
                                  "original_paths": r["members"], "target_dir": d, "area_id": a}))
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


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--repo", required=True)
    ap.add_argument("--data", default=None)
    a = ap.parse_args(argv)
    bad = x2d.verify_pins(require_signoff=x2d.require_signoff())
    if bad:
        sys.exit("pins: " + "; ".join(bad))
    data = a.data
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
