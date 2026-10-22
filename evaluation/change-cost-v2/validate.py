#!/usr/bin/env python3
"""E1b-v2 measurement validation: gates G4 (task type), G5 (follow-up labeller), G7 (hidden-coupling proxy).

Usage:
  uv run --with pandas python validate.py sample                    # validation/{type,pair,dep}-sample[-blinded].jsonl, RUBRIC.json, strata
  (annotate.py labels the blinded files with pinned models)         # validation/annotations-{A,B,C}.jsonl
  uv run --with pandas python validate.py gates                     # reads annotations + samples, writes validation-gates.json

Blinding: item ids are assigned AFTER shuffling, so position carries no information; the blinded files contain only the id and the
content the annotator needs (no pipeline label, no Jev output/probability, no stratum, no candidate kind, no repository).
Gates are mechanical: analysis.py reads validation-gates.json; nobody sets a flag by hand. Thresholds (PREREGISTRATION section 5):
  G4: kappa(A,B) >= 0.5, Jev-vs-consensus accuracy >= 0.70, unresolved <= 10 %.
  G5: kappa >= 0.6, Jev precision >= 0.70, inverse-probability-weighted recall >= 0.60, >= 20 consensus positives,
      consensus `cannot_tell` items are EXCLUDED from precision/recall and must be <= 15 % of the sample.
  G7: annotators label (via, partner) PAIRS; pairs aggregate to PARTNERS (any `yes` -> yes; else any unresolved/`cannot_tell` -> unresolved;
      else no) and every metric below is partner-level: kappa >= 0.6 (pair-level); proxy-hidden partners judged `no static dependence`
      >= 0.80 (precision of "hidden"); proxy-explained partners judged `yes` >= 0.70; hidden precision >= 0.60 in EVERY sampled language
      group (a group with no decided partner fails); unresolved partners <= 15 %.
"""
import gzip
import json
import random
import sys
from collections import Counter, defaultdict

import pandas as pd

from common import HERE, git

SEED = 20261020
TAU = 0.5
VAL = HERE / "validation"
LANG = {".py": "python", ".js": "js_ts", ".mjs": "js_ts", ".cjs": "js_ts", ".ts": "js_ts", ".tsx": "js_ts", ".go": "go", ".rs": "rust",
        ".java": "jvm_dotnet", ".cs": "jvm_dotnet"}
RUBRIC = {
    "type": """TASK TYPE (title + labels + description + changed files). Answer exactly one of:
fix = repairs incorrect behaviour: a bug, crash, panic, regression, security flaw, wrong result or missing error handling;
feature = adds a capability, API, option, command or support that did not exist before;
refactor = restructures, renames, cleans up, simplifies, deprecates, removes dead code or tunes performance without intending to
change observable behaviour;
other_dev = other source-code development work not clearly one of the above (intentional behaviour/API changes, compatibility,
configuration, mixed changes);
nondev = not development effort on the product code: dependency/version bump, release/changelog, documentation-only or comment-only,
tests-only, CI/build configuration, backport or cherry-pick of another pull request, automated/generated change.""",
    "pair": """CORRECTIVE FOLLOW-UP (earlier PR, later PR merged within 14 days). Answer exactly one of:
corrective = the later PR fixes a defect/regression introduced by the earlier one, completes something the earlier one omitted, or reverts it;
independent = anything else (ordinary continued development, unrelated change, refactor, chore);
cannot_tell = the descriptions do not give enough information.""",
    "dep": """STATIC DEPENDENCE between two source files of one repository (head of each file shown). Answer exactly one of:
yes = one file statically depends on the other: imports/includes/uses it by module, package or symbol name, inherits from it, or
calls something it defines (including same-package symbol use that the language resolves without an import);
no = neither file uses anything defined in the other; any link between them is only that they tend to change together;
cannot_tell = the heads shown are insufficient to decide.""",
}
LABELS = {"type": ["fix", "feature", "refactor", "other_dev", "nondev"], "pair": ["corrective", "independent", "cannot_tell"],
          "dep": ["yes", "no", "cannot_tell"]}


def jev_means():
    d = defaultdict(list)
    lf = HERE / "labels.jsonl"
    if lf.exists():
        for l in lf.read_text().splitlines():
            r = json.loads(l)
            if r.get("ok"):
                d[(r["repo"], r["parent"], r["follow"])].append(r["p_corrective"])
    return {k: sum(v) / len(v) for k, v in d.items()}


def head_at(repo, rev, path, n=150):
    try:
        return "\n".join(git(repo, "show", f"{rev}:{path}", check=False).splitlines()[:n])
    except Exception:
        return ""


def sample():
    from build_panel import load_types
    VAL.mkdir(exist_ok=True)
    (VAL / "RUBRIC.json").write_text(json.dumps(RUBRIC, indent=1))
    rng = random.Random(SEED)
    raws = {}

    def raw(repo):
        if repo not in raws:
            raws[repo] = {p["number"]: p for p in json.loads(gzip.open(HERE / "raw" / f"prs-{repo}.json.gz", "rt").read())}
        return raws[repo]

    # ---- G4: 30 PRs per Jev class (dev classes from panel PRs; `nondev` from all typed PRs, since nondev PRs never enter the panel)
    panel = pd.read_csv(HERE / "panel-final.csv.gz")
    inpanel = set(zip(panel.repo, panel.number))
    types = {k: v for k, v in load_types().items() if k in inpanel or v == "nondev"}
    by_class = defaultdict(list)
    for k in sorted(types):
        by_class[types[k]].append(k)
    items = []
    for t in sorted(by_class):
        rng.shuffle(by_class[t])
        items += [(t, k) for k in by_class[t][:30]]
    rng.shuffle(items)
    type_rows = []
    for i, (t, (repo, n)) in enumerate(items):
        p = raw(repo)[n]
        type_rows.append(dict(id=f"T{i:03d}", repo=repo, number=n, pipeline=t, title=p["title"], labels=p["labels"],
                              description=p["body"][:700], files=[f["path"] for f in p["files"]][:15]))
    # ---- G5: 100 Jev-positive + 100 Jev-negative candidate pairs
    cands = [c for c in (json.loads(l) for l in gzip.open(HERE / "candidates.jsonl.gz", "rt")) if not c["censored"]]
    jm = jev_means()
    strata = {"jev_pos": [c for c in cands if jm.get((c["repo"], c["parent"], c["follow"]), -1) >= TAU],
              "jev_neg": [c for c in cands if 0 <= jm.get((c["repo"], c["parent"], c["follow"]), -1) < TAU]}
    pair_items, sizes = [], {}
    for k in sorted(strata):
        pool = strata[k]
        rng.shuffle(pool)
        sizes[k] = dict(population=len(pool), sampled=min(100, len(pool)))
        pair_items += [(k, c) for c in pool[:100]]
    rng.shuffle(pair_items)
    pair_rows = []
    for i, (k, c) in enumerate(pair_items):
        a, b = raw(c["repo"])[c["parent"]], raw(c["repo"])[c["follow"]]
        view = lambda p: dict(title=p["title"], description=p["body"][:700], files=[f["path"] for f in p["files"]][:15])
        pair_rows.append(dict(id=f"P{i:03d}", repo=c["repo"], parent=c["parent"], follow=c["follow"], stratum=k, kind=c["kind"],
                              earlier=view(a), later=view(b), hours_between=round(c["gap_h"], 1)))
    # ---- G7: partner pairs stratified by language group x proxy decision, 12 per cell
    cells = defaultdict(list)
    for f in sorted((HERE / "data").glob("pr-*.jsonl.gz")):
        repo = f.name[3:-9]
        for r in (json.loads(l) for l in gzip.open(f, "rt")):
            if "skip" in r or (repo, r["number"]) not in inpanel:
                continue
            for pt in r.get("partners", []):
                ext = "." + pt["g"].rsplit(".", 1)[-1]
                cells[(LANG.get(ext, "other"), "explained" if pt["explained"] else "hidden")].append((repo, r["S"], pt))
    dep_items = []
    for key in sorted(cells):
        if key[0] == "other":
            continue
        rng.shuffle(cells[key])
        dep_items += [(key, x) for x in cells[key][:12]]
    rng.shuffle(dep_items)
    # PAIR-LEVEL items (one per (via, partner) pair; explained partner -> its via_ref pair, hidden partner -> up to 3 rng-drawn vias),
    # opaque ids assigned after a global shuffle; the partner grouping lives only in the unblinded sample (`partner_key`)
    dep_pairs = []
    for pk, ((lang, proxy), (repo, S, pt)) in enumerate(dep_items):
        vias = [pt["via_ref"]] if proxy == "explained" else rng.sample(pt["vias"], min(3, len(pt["vias"])))
        for v in vias:
            dep_pairs.append(dict(repo=repo, lang=lang, proxy=proxy, partner_key=f"K{pk:03d}",
                                  file_a=v, file_b=pt["g"], head_a=head_at(repo, S, v), head_b=head_at(repo, S, pt["g"])))
    rng.shuffle(dep_pairs)
    dep_rows = [dict(id=f"D{i:03d}", **r) for i, r in enumerate(dep_pairs)]
    hidden_keys = ("pipeline", "kind", "repo", "number", "parent", "follow", "stratum", "lang", "proxy", "partner_key")
    for name, rows in (("type", type_rows), ("pair", pair_rows), ("dep", dep_rows)):
        with (VAL / f"{name}-sample.jsonl").open("w") as fh:
            for r in rows:
                fh.write(json.dumps(r) + "\n")
        with (VAL / f"{name}-blinded.jsonl").open("w") as fh:
            for r in rows:
                fh.write(json.dumps({k: v for k, v in r.items() if k not in hidden_keys}) + "\n")
    (VAL / "pair-strata.json").write_text(json.dumps(sizes, indent=1))
    print(len(type_rows), "type;", len(pair_rows), "pair;", len(dep_rows), "dep items;", sizes)


def disagree(a, b):
    """A and B disagree when their labels differ or either is `unparsable` (prereg §5: an unparsable answer is a disagreement)."""
    return a != b or a == "unparsable"


def kappa(a, b):
    """Cohen's kappa where an `unparsable` answer never agrees (not even with another `unparsable`) and adds nothing to chance
    agreement pe."""
    n = len(a)
    po = sum(not disagree(x, y) for x, y in zip(a, b)) / n
    ca, cb = Counter(a), Counter(b)
    pe = sum(ca[k] * cb[k] for k in (set(ca) | set(cb)) - {"unparsable"}) / n ** 2
    return (po - pe) / (1 - pe) if pe < 1 else 1.0


def read_ann(letter):
    f = VAL / f"annotations-{letter}.jsonl"
    return {json.loads(l)["id"]: json.loads(l)["label"] for l in f.read_text().splitlines() if l.strip()} if f.exists() else {}


def consensus(A, B, C, ids):
    """A when A and B agree (`disagree` is False), else C's label (None if C is missing/unparsable)."""
    def ok(x):
        return None if x == "unparsable" else x
    return {i: (A[i] if not disagree(A[i], B[i]) else ok(C.get(i))) for i in ids}


def rows_of(name):
    return {json.loads(l)["id"]: json.loads(l) for l in (VAL / f"{name}-sample.jsonl").read_text().splitlines()}


def gates():
    A, B, C = read_ann("A"), read_ann("B"), read_ann("C")
    out = {}
    # ---- G4
    rows = rows_of("type")
    ids = [i for i in rows if i in A and i in B]
    cons = consensus(A, B, C, ids)
    ok = [i for i in ids if cons[i] is not None]
    k = kappa([A[i] for i in ids], [B[i] for i in ids]) if ids else 0.0
    acc = sum(rows[i]["pipeline"] == cons[i] for i in ok) / max(len(ok), 1)
    unres = (len(ids) - len(ok)) / max(len(ids), 1)
    out["G4"] = dict(n=len(ids), kappa=round(k, 3), accuracy=round(acc, 3), unresolved_share=round(unres, 3),
                     confusion=dict(Counter(f"{rows[i]['pipeline']}->{cons[i]}" for i in ok)),
                     **{"pass": bool(ids and k >= 0.5 and acc >= 0.70 and unres <= 0.10)})
    # ---- G5
    rows = rows_of("pair")
    sizes = json.loads((VAL / "pair-strata.json").read_text())
    w = {s: v["population"] / max(v["sampled"], 1) for s, v in sizes.items()}
    ids = [i for i in rows if i in A and i in B]
    cons = consensus(A, B, C, ids)
    k = kappa([A[i] for i in ids], [B[i] for i in ids]) if ids else 0.0
    ct = sum(cons[i] == "cannot_tell" or cons[i] is None for i in ids) / max(len(ids), 1)
    dec = [i for i in ids if cons[i] in ("corrective", "independent")]
    tp = sum(w[rows[i]["stratum"]] for i in dec if rows[i]["stratum"] == "jev_pos" and cons[i] == "corrective")
    fp = sum(w[rows[i]["stratum"]] for i in dec if rows[i]["stratum"] == "jev_pos" and cons[i] == "independent")
    fn = sum(w[rows[i]["stratum"]] for i in dec if rows[i]["stratum"] == "jev_neg" and cons[i] == "corrective")
    prec = tp / (tp + fp) if tp + fp else 0.0
    rec = tp / (tp + fn) if tp + fn else 0.0
    npos = sum(cons[i] == "corrective" for i in dec)
    out["G5"] = dict(n=len(ids), kappa=round(k, 3), jev_precision=round(prec, 3), jev_recall_ipw=round(rec, 3),
                     consensus_positives=npos, cannot_tell_or_unresolved_share=round(ct, 3),
                     **{"pass": bool(ids and k >= 0.6 and prec >= 0.70 and rec >= 0.60 and npos >= 20 and ct <= 0.15)})
    # ---- G7: annotators label PAIRS (opaque ids); partner label: any `yes` pair => yes; else any unresolved/cannot_tell pair => unresolved; else no
    rows = rows_of("dep")
    ids = [i for i in rows if i in A and i in B]
    cons = consensus(A, B, C, ids)
    k = kappa([A[i] for i in ids], [B[i] for i in ids]) if ids else 0.0
    groups = defaultdict(list)
    for i, r in rows.items():
        groups[r["partner_key"]].append(i)
    plab, pmeta = {}, {}
    for pk, members in groups.items():
        labs = [cons.get(i) if i in cons else None for i in members]
        plab[pk] = "yes" if "yes" in labs else ("unresolved" if any(x in (None, "cannot_tell") for x in labs) else "no")
        pmeta[pk] = (rows[members[0]]["lang"], rows[members[0]]["proxy"])
    ct = sum(v == "unresolved" for v in plab.values()) / max(len(plab), 1)
    hid = [pk for pk in plab if pmeta[pk][1] == "hidden" and plab[pk] != "unresolved"]
    exp = [pk for pk in plab if pmeta[pk][1] == "explained" and plab[pk] != "unresolved"]
    hp = sum(plab[pk] == "no" for pk in hid) / max(len(hid), 1)
    ep = sum(plab[pk] == "yes" for pk in exp) / max(len(exp), 1)
    by_lang = {}
    for lg in sorted({pmeta[pk][0] for pk in plab if pmeta[pk][1] == "hidden"}):   # every SAMPLED hidden language, decided or not
        hl = [pk for pk in hid if pmeta[pk][0] == lg]
        by_lang[lg] = dict(n=len(hl), hidden_precision=round(sum(plab[pk] == "no" for pk in hl) / len(hl), 3) if hl else None)
    cells = Counter(f"{pmeta[pk][0]}/{pmeta[pk][1]}" for pk in plab)   # partner-level cell sizes
    out["G7"] = dict(n_partners=len(plab), n_pairs=len(ids), kappa_pairs=round(k, 3), hidden_precision=round(hp, 3), explained_precision=round(ep, 3),
                     cell_sizes=dict(cells), by_language=by_lang, cannot_tell_or_unresolved_share=round(ct, 3),
                     **{"pass": bool(ids and k >= 0.6 and hp >= 0.80 and ep >= 0.70 and ct <= 0.15 and
                                     by_lang and all(v["hidden_precision"] is not None and v["hidden_precision"] >= 0.60
                                                     for v in by_lang.values()))})
    (HERE / "validation-gates.json").write_text(json.dumps(out, indent=1))
    print(json.dumps(out, indent=1))


if __name__ == "__main__":
    sample() if sys.argv[1] == "sample" else gates()
