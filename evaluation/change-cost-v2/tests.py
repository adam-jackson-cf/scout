#!/usr/bin/env python3
"""E1b-v2 implementation tests (synthetic git fixture + pure-function checks).

Usage: uv run --with lizard,numpy,pandas,scipy python tests.py
Covers the defects the v1 critique found: deletions must count as touched/churned files, follow-ups strictly after the parent,
stable-population degradation, trailing/co-change windows using only information before the snapshot, bot filter without
substring matching, task-type rules, area deepening, session-based active hours, Holm, within-repo standardisation.
"""
import gzip
import json
import os
import subprocess
import sys
import tempfile
from pathlib import Path

import numpy as np
import pandas as pd

import analysis as an
import build_panel as bp
import collect
import common

DAY = 86400
FAILS = []


def check(name, cond):
    print(("ok   " if cond else "FAIL ") + name)
    if not cond:
        FAILS.append(name)


def sh(cwd, *a, env=None):
    subprocess.run(["git", "-C", str(cwd), *a], check=True, capture_output=True, env={**os.environ, **(env or {})})


def commit(cwd, msg, ts):
    env = dict(GIT_AUTHOR_DATE=f"{ts} +0000", GIT_COMMITTER_DATE=f"{ts} +0000", GIT_AUTHOR_NAME="a", GIT_AUTHOR_EMAIL="a@x",
               GIT_COMMITTER_NAME="a", GIT_COMMITTER_EMAIL="a@x")
    sh(cwd, "add", "-A")
    sh(cwd, "commit", "-q", "-m", msg, env=env)
    return subprocess.run(["git", "-C", str(cwd), "rev-parse", "HEAD"], capture_output=True, text=True).stdout.strip()


def complex_py(n_branches, name="f"):
    body = "\n".join(f"    if x == {i}:\n        y += {i}" for i in range(n_branches))
    return f"def {name}(x):\n    y = 0\n{body}\n    return y\n"


def fixture():
    d = Path(tempfile.mkdtemp())
    sh(d, "init", "-q", "-b", "main")
    t0 = 1_700_000_000
    (d / "pkg").mkdir()
    (d / "pkg/a.py").write_text(complex_py(5) + "\n".join(f"# pad {i}\nz{i} = {i}" for i in range(12)) + "\n")
    (d / "pkg/b.py").write_text("import zeta_mod\n" + "\n".join(f"v{i} = {i}" for i in range(15)) + "\n")  # b.py references zeta_mod
    (d / "pkg/zeta_mod.py").write_text("\n".join(f"u{i} = {i}" for i in range(15)) + "\n")
    (d / "pkg/c.py").write_text("\n".join(f"w{i} = {i}" for i in range(15)) + "\n")
    commit(d, "init", t0)
    for k in range(4):   # a.py and c.py co-change 4 times in the 365 d before S (no textual reference between them)
        for nm in ("a", "b", "zeta_mod"):   # a.py, b.py, zeta_mod.py co-change (a.py alone has no reference to zeta_mod)
            (d / f"pkg/{nm}.py").write_text((d / f"pkg/{nm}.py").read_text() + f"q{k} = {k}\n")
        (d / "pkg/c.py").write_text((d / "pkg/c.py").read_text() + f"r{k} = {k}\n")
        commit(d, f"co-change {k}", t0 + (200 + k) * DAY)
    (d / "pkg/a.py").write_text(complex_py(14) + "\n".join(f"z{i} = {i}" for i in range(12)) + "\n")  # a.py gets MORE complex
    S = commit(d, "snapshot base", t0 + 400 * DAY)
    (d / "pkg/a.py").write_text((d / "pkg/a.py").read_text() + "extra = 1\n")
    (d / "pkg/b.py").unlink()                                    # PR deletes b.py
    pr_commit = commit(d, "pr work", t0 + 401 * DAY)
    return d, S, pr_commit, t0


def pr_dict(S, first_parent, t0):
    return dict(number=7, title="Fix thing", body="", labels=[], created="2023-01-01T00:00:00Z", merged="2023-01-02T00:00:00Z",
                assoc="MEMBER", additions=1, deletions=15, changed_files=2, base_oid=S, head_oid=None, author="u", author_type="User",
                merge_commit=None, merge_parents=[], n_commits=1,
                commits=[dict(oid="x", authored=pd.Timestamp(t0 + 401 * DAY, unit="s", tz="UTC").isoformat(), committed="2023-01-01T00:00:00Z", email="a@x", parent=first_parent)],
                n_files=2, files=[dict(path="pkg/a.py", add=1, dele=0, ct="MODIFIED"), dict(path="pkg/b.py", add=0, dele=15, ct="DELETED")],
                n_reviews=0, reviews=[])


def test_collect():
    d, S, prc, t0 = fixture()
    common.REPOS = d.parent
    collect.REPOS = d.parent
    common.REPO_CFG["fx"] = ("o/fx", d.name, "head")
    commits = collect.load_history("fx", "HEAD")
    collect.G.update(repo="fx", hist=commits, idx=collect.build_index(commits))
    row = collect.process(pr_dict(S, S, t0))
    check("snapshot parent selected", row.get("S") == S and row["S_src"] == "parent")
    check("deleted file counted as touched+measured (v1 gave deletions 0 churn)", row.get("n_old") == 2 and row["n_measured"] == 2)
    check("churn counts deletions", row["churn"] == 16)
    paths = {f["p"] for f in row["files"]}
    check("both files in F_S", paths == {"pkg/a.py", "pkg/b.py"})
    fa = next(f for f in row["files"] if f["p"] == "pkg/a.py")
    check("a.py high-CC measured at S", fa["nh"] == 1 and fa["hi"] > 0)
    check("stable population: files present at trailing snapshot with same path", row["n_stable"] == 2)
    check("degradation uses SAME files at both snapshots", fa["stable"] and fa["nh_t"] == 0 and fa["hi_t"] == 0)
    trail_ts = int(common.git("fx", "show", "-s", "--format=%ct", row["trail"]).strip())
    check("trailing snapshot is >=365 d before S", row["trail"] is not None and row["S_ts"] - trail_ts >= 365 * DAY)
    byg = {p["g"]: p for p in row["partners"]}
    check("hidden: c.py (no reference from any touched co-changing file)", byg["pkg/c.py"]["explained"] is False)
    check("ANY-via rule: zeta_mod explained because b.py references it although the best-confidence via is a.py",
          byg["pkg/zeta_mod.py"]["explained"] is True and "pkg/a.py" in byg["pkg/zeta_mod.py"]["vias"] and "pkg/b.py" in byg["pkg/zeta_mod.py"]["vias"])
    check("n_hidden counts only unexplained partners", row["n_hidden"] == 1 and row["n_partner_abs"] == 2)
    check("via_ref records the touched file whose reference explained the partner (b.py for zeta_mod; None for hidden c.py)",
          byg["pkg/zeta_mod.py"]["via_ref"] == "pkg/b.py" and byg["pkg/c.py"]["via_ref"] is None)
    check("author prior commits counted before first PR commit only", row["author_prior_commits"] == 6)
    # a PR whose author is a bot is skipped by exact type/login, not substring
    p = pr_dict(S, S, t0)
    p["author"], p["author_type"] = "dependabot", "User"
    check("login-list bot (type User, login dependabot) is a bot", collect.process(p).get("skip") == "bot")
    check("shared predicate: robot-fan is not a bot", not common.is_bot("robot-fan", "User") and common.is_bot("renovate", None))
    check("referenced(): whole-word, non-comment, non-generic only",
          collect.referenced("import pkg_c_mod\n", "pkg/pkg_c_mod.py") and not collect.referenced("# pkg_c_mod\n", "pkg/pkg_c_mod.py")
          and not collect.referenced("import utils\n", "pkg/utils.py") and not collect.referenced("import pkg_c_modx\n", "pkg/pkg_c_mod.py"))
    p = pr_dict(S, S, t0)
    p["author"], p["author_type"] = "robot-fan", "User"
    check("substring 'bot' in a human login is NOT a bot", "skip" not in collect.process(p) or collect.process(p).get("skip") != "bot")
    p["author_type"] = "Bot"
    check("GraphQL Bot type is a bot", collect.process(p).get("skip") == "bot")
    p = pr_dict(S, S, t0)
    p["n_files"] = 51
    check("bulk PR skipped", collect.process(p).get("skip") == "bulk")


def test_pure():
    check("eligible: production py", common.eligible("pkg/a.py"))
    check("not eligible: tests dir", not common.eligible("pkg/tests/test_a.py"))
    check("not eligible: _test.go", not common.eligible("x/y_test.go"))
    check("not eligible: .d.ts", not common.eligible("src/a.d.ts"))
    check("not eligible: docs", not common.eligible("docs/conf.py"))
    check("active hours: single commit = floor credit", abs(bp.active_hours([0]) - 0.5) < 1e-9)
    check("active hours: 1h gap within a session", abs(bp.active_hours([0, 3600]) - 1.5) < 1e-9)
    check("active hours: >4h gap starts a new session", abs(bp.active_hours([0, 5 * 3600]) - 1.0) < 1e-9)
    tmp = Path(tempfile.mkdtemp())
    recs = [dict(repo="r", number=1, rep=0, ok=True, probs=dict(fix=0.9, feature=0.05, refactor=0.0, other_dev=0.05, nondev=0.0)),
            dict(repo="r", number=2, rep=0, ok=True, probs=dict(fix=0.5, feature=0.5, refactor=0.0, other_dev=0.0, nondev=0.0)),
            dict(repo="r", number=3, rep=0, ok=False, http=500)]
    (tmp / "types.jsonl").write_text("\n".join(json.dumps(r) for r in recs) + "\n")
    old, bp.HERE = bp.HERE, tmp
    try:
        got = bp.load_types()
    finally:
        bp.HERE = old
    check("task type: argmax of Jev class probabilities", got.get(("r", 1)) == "fix")
    check("task type: tie falls to other_dev; failed calls are unlabelled", got.get(("r", 2)) == "other_dev" and ("r", 3) not in got)
    x = np.array([1.0, 2.0, 3.0, 10.0, 12.0, 14.0])
    r = np.array([0, 0, 0, 1, 1, 1])
    z = an.within_repo_std(x, r)
    check("within-repo std: zero mean per repo", abs(z[:3].mean()) < 1e-12 and abs(z[3:].mean()) < 1e-12)
    check("holm monotone", an.holm({"a": 0.001, "b": 0.02, "c": 0.5}) == {"a": 0.003, "b": 0.04, "c": 0.5})
    df = pd.DataFrame(dict(main_path=[f"x/y/f{i}.py" for i in range(8)] + ["x/z/g.py"] * 2, repo=["r"] * 10))
    ar = bp.assign_areas(df)
    check("area depth rule: area holding > 15% of PRs is deepened (x/y -> x/y at depth 3 is a directory, so stays; split only when deeper dirs exist)",
          set(ar[:8]) == {"r:x/y"} and set(ar[8:]) == {"r:x/z"})
    df2 = pd.DataFrame(dict(main_path=[f"x/y/s{i % 4}/f.py" for i in range(8)] + ["x/z/g.py"] * 2, repo=["r"] * 10))
    ar2 = bp.assign_areas(df2)
    check("area depth rule: x/y (80% of PRs) is deepened into x/y/sN", set(ar2[:8]) == {f"r:x/y/s{i}" for i in range(4)})
    # ---- statistics
    rng = np.random.default_rng(0)
    n = 4000
    Xs = rng.normal(size=(n, 2))
    yb = (rng.random(n) < 1 / (1 + np.exp(-(-1 + 0.5 * Xs[:, 0] + 0.2 * Xs[:, 1])))).astype(float)
    b = an.logit(yb, Xs, [rng.integers(0, 3, n)])
    check("logit recovers a planted log-odds coefficient (0.5 +- 0.12)", abs(b - 0.5) < 0.12)
    e = rng.normal(size=300) * np.repeat([1.0, 5.0, 2.0], 100)
    u = an.rel_abs_resid(e, np.repeat([0, 1, 2], 100))
    check("rel_abs_resid: mean 1 within every repo", all(abs(u[np.repeat([0, 1, 2], 100) == r].mean() - 1) < 1e-9 for r in range(3)))
    fake = lambda est, se, p: dict(est=est, se=se, p=p, ci95=[est - 2 * se, est + 2 * se], ci90=[est - 1.6 * se, est + 1.6 * se])
    res = {"T1": fake(0.30, 0.05, 1e-9), "T2": fake(0.05, 0.01, 1e-6), "T3": fake(-0.30, 0.05, 1e-9), "T4": fake(0.0, 0.02, 0.9),
           "T5": fake(0.0, 0.2, 0.9)}
    v, h = an.decide(res, adequate={"T1": True, "T2": True, "T3": True, "T4": True, "T5": False}, t5_decided=True)
    check("decide: supported / detected_below_SESOI / construct_invalid / equivalent / inconclusive branches",
          v["T1"]["verdict"] == "supported" and v["T2"]["verdict"] == "detected_below_SESOI"
          and v["T3"]["verdict"].startswith("construct_invalid") and v["T4"]["verdict"].startswith("equivalent") and v["T5"]["verdict"] == "inconclusive")
    v2, _ = an.decide(res, adequate={}, t5_decided=False)
    check("decide: no adequacy evidence => never equivalent; T5 undecided when G5/G7 fail and excluded from Holm",
          v2["T4"]["verdict"] == "inconclusive" and v2["T5"]["verdict"].startswith("not_decided") and "T5" in v2)
    # ---- finalize_panel NaN rules
    import finalize_panel as fp
    check("finalize_panel: censored parents and unlabelled pairs are NaN, not 0", fp.outcome_rule(True, [0.9], [True]) is None
          and fp.outcome_rule(False, [0.9], [False]) is None and fp.outcome_rule(False, [], []) == 0.0
          and fp.outcome_rule(False, [0.2, 0.7], [True, True]) == 1.0 and fp.outcome_rule(False, [0.2, 0.3], [True, True]) == 0.0)


def _g7(va, items, labels):
    """Run validate.gates() G7 on synthetic pair-level items {id, partner_key, lang, proxy} and a label dict (A = B)."""
    old = (va.HERE, va.VAL)
    tmp = Path(tempfile.mkdtemp())
    va.HERE, va.VAL = tmp, tmp / "validation"
    va.VAL.mkdir()
    wl = lambda name, rows: (va.VAL / name).write_text("".join(json.dumps(r) + "\n" for r in rows))
    wl("type-sample.jsonl", [dict(id="T000", pipeline="fix")])
    wl("pair-sample.jsonl", [dict(id="P000", stratum="jev_pos")])
    (va.VAL / "pair-strata.json").write_text(json.dumps(dict(jev_pos=dict(population=1, sampled=1), jev_neg=dict(population=1, sampled=1))))
    wl("dep-sample.jsonl", items)
    for L in "ABC":
        rows = [dict(id="T000", label="fix"), dict(id="P000", label="corrective")] + [dict(id=k, label=x) for k, x in labels.items()]
        wl(f"annotations-{L}.jsonl", rows if L != "C" else [])
    try:
        import io
        import contextlib
        with contextlib.redirect_stdout(io.StringIO()):
            va.gates()
        return json.loads((tmp / "validation-gates.json").read_text())["G7"]
    finally:
        va.HERE, va.VAL = old


def test_g7_pairs(va):
    # partner K0: 3 hidden pairs, mixed no/no/yes -> partner `yes` (depends); K1: 2 hidden pairs both no -> `no`;
    # K2: hidden pair no + pair cannot_tell -> unresolved; K3: explained single pair yes
    items = [dict(id="D0", partner_key="K0", lang="python", proxy="hidden"), dict(id="D1", partner_key="K0", lang="python", proxy="hidden"),
             dict(id="D2", partner_key="K0", lang="python", proxy="hidden"), dict(id="D3", partner_key="K1", lang="python", proxy="hidden"),
             dict(id="D4", partner_key="K1", lang="python", proxy="hidden"), dict(id="D5", partner_key="K2", lang="python", proxy="hidden"),
             dict(id="D6", partner_key="K2", lang="python", proxy="hidden"), dict(id="D7", partner_key="K3", lang="python", proxy="explained")]
    lab = dict(D0="no", D1="no", D2="yes", D3="no", D4="no", D5="no", D6="cannot_tell", D7="yes")
    g = _g7(va, items, lab)
    check("G7 partner aggregation: 4 partners / 8 pairs, partner-level cell sizes, unresolved share 1/4, hidden precision 1/2 (K1 no, K0 yes)",
          g["n_partners"] == 4 and g["n_pairs"] == 8 and g["cell_sizes"] == {"python/hidden": 3, "python/explained": 1}
          and abs(g["cannot_tell_or_unresolved_share"] - 0.25) < 1e-9 and abs(g["hidden_precision"] - 0.5) < 1e-9 and g["explained_precision"] == 1.0)
    items2 = [dict(id=f"D{i}", partner_key=f"K{i}", lang="python", proxy="hidden") for i in range(10)]
    items2 += [dict(id="D10", partner_key="K10", lang="go", proxy="hidden"), dict(id="D11", partner_key="K11", lang="python", proxy="explained")]
    lab2 = {f"D{i}": "no" for i in range(10)}
    lab2.update(D10="cannot_tell", D11="yes")
    g2 = _g7(va, items2, lab2)
    check("G7: a sampled hidden language with zero decided partners is reported (precision None) and fails the gate",
          g2["by_language"]["go"] == dict(n=0, hidden_precision=None) and g2["hidden_precision"] == 1.0 and g2["pass"] is False)
    check("G7: the same sample passes once that language has a decided precise partner", _g7(va, items2, dict(lab2, D10="no"))["pass"] is True)


def test_g1():
    ok, bad = dict(ok=90, no_snapshot=10), dict(ok=80, no_snapshot=20)
    check("G1: snapshot_coverage formula (90 % passes the 0.85 gate, 80 % fails it), empty funnel is 0, nondev not double counted",
          abs(bp.snapshot_coverage(ok) - 0.9) < 1e-12 and bp.snapshot_coverage(ok) >= bp.GATE_MIN_COVERAGE
          and bp.snapshot_coverage(bad) < bp.GATE_MIN_COVERAGE and bp.snapshot_coverage({}) == 0.0
          and abs(bp.snapshot_coverage(dict(ok=10, no_preexisting_file=5, no_measurable_file=5, nondev=4)) - 1.0) < 1e-12)


def test_round4():
    import harness
    # ---- jackknife: exploratory statistics count only clusters in their OWN complete-case sample (P1 needs y_active and D)
    df = harness.synth(np.random.default_rng(1), n_repos=4, size=300)
    r0 = df["repo"].unique()[0]
    df.loc[df["repo"] == r0, "y_active"] = np.nan          # django-style: one repository has no y_active
    res = an.analyse(df)
    p1_areas = df.dropna(subset=an.EXPLORATORY_COLS["P1"])["area"].nunique()
    t5_areas = df.dropna(subset=an.TEST_COLS["T5"])["area"].nunique()
    check("jackknife: P1 df = its own clusters - 1 (repo without y_active excluded); T5 keeps every cluster",
          res["P1_rel_gain"]["df"] == p1_areas - 1 and p1_areas < t5_areas and res["T5"]["df"] == t5_areas - 1)
    check("stat_cols: P1_* map to the P1 set, S1/S2 to their own outcome, unknown names raise",
          an.stat_cols("P1_mse_full") == an.EXPLORATORY_COLS["P1"] and "y_review" in an.stat_cols("S1")
          and "y_rework" in an.stat_cols("S2") and _raises(lambda: an.stat_cols("X9"), KeyError))
    # ---- logit: an exhausted IRLS iteration budget is a hard error, not a silent estimate
    rng = np.random.default_rng(5)
    X = rng.normal(size=(400, 2))
    y = (rng.random(400) < 1 / (1 + np.exp(-(0.8 * X[:, 0])))).astype(float)
    check("logit: non-convergence (iters=1) raises RuntimeError; default budget converges",
          _raises(lambda: an.logit(y, X, [rng.integers(0, 3, 400)], iters=1), RuntimeError)
          and np.isfinite(an.logit(y, X, [rng.integers(0, 3, 400)])))


def _raises(fn, exc):
    try:
        fn()
    except exc:
        return True
    return False



def test_round2():
    import harness
    import annotate
    import validate as va
    # ---- G3/G3b mechanics
    fake = lambda est, se, p: dict(est=est, se=se, p=p, ci95=[est - 2 * se, est + 2 * se], ci90=[est - 1.6 * se, est + 1.6 * se])
    res = {t: fake(0.30, 0.05, 1e-9) for t in an.FAMILY}
    v, h = an.decide(res, adequate={t: True for t in an.FAMILY}, t5_decided=True, h12_valid=False)
    check("decide: G3/G3b failure => T1-T4 outcome_construct_invalid, not in the Holm family; T5 still decided",
          all(v[t]["verdict"].startswith("outcome_construct_invalid") and v[t]["adj_p"] is None for t in ("T1", "T2", "T3", "T4"))
          and v["T5"]["verdict"] == "supported" and abs(v["T5"]["adj_p"] - 1e-9) < 1e-12 and h["H1"].startswith("outcome_construct_invalid"))
    # ---- annotate.parse
    check("annotate.parse: exact label (case/punctuation tolerant), anything else is `unparsable`",
          annotate.parse("pair", " Corrective. ") == "corrective" and annotate.parse("dep", "Yes, they import each other") == "unparsable"
          and annotate.parse("type", None) == "unparsable" and annotate.parse("dep", "cannot_tell") == "cannot_tell")
    # ---- annotate.answer: truncated/empty completions abort (infrastructure failure), never become `unparsable`
    resp = lambda content, fin="stop", model="google/gemini-2.5-pro": dict(model=model, choices=[dict(finish_reason=fin, message=dict(content=content))])
    check("annotate.answer: any finish_reason other than stop (length, content_filter, error, None) or empty content aborts; wrong model aborts; normal answer parsed",
          all(_raises(lambda f=f: annotate.answer("C", "dep", resp("yes" if f != "length" else "\n**Analyzing", f)), SystemExit)
              for f in ("length", "content_filter", "error", None))
          and _raises(lambda: annotate.answer("C", "dep", resp(None)), SystemExit)
          and _raises(lambda: annotate.answer("C", "dep", resp("yes", model="google/gemini-2.5-flash")), SystemExit)
          and annotate.answer("C", "dep", resp("Yes.")) == ("google/gemini-2.5-pro", "yes")
          and annotate.answer("C", "dep", resp("maybe")) == ("google/gemini-2.5-pro", "unparsable"))
    bc, ba = annotate.request_body("C", "q"), annotate.request_body("A", "q")
    check("annotate.request_body: C gets a bounded reasoning budget below max_tokens; A/B get no reasoning parameter",
          bc["reasoning"]["max_tokens"] < bc["max_tokens"] and "reasoning" not in ba and "reasoning" not in annotate.request_body("B", "q")
          and bc["temperature"] == 0 and bc["model"] == "google/gemini-2.5-pro")

    def _down(*a, **k):
        raise TimeoutError("synthetic outage")
    check("annotate.call: three failed attempts abort the run (never an `unparsable` label)",
          _raises(lambda: annotate.call(("C", "dep", dict(id="D999")), "k", urlopen=_down), SystemExit))
    # ---- validate.consensus / disagree / kappa: an `unparsable` answer is a disagreement (prereg §5)
    cons = va.consensus({"a": "x", "b": "unparsable", "c": "x", "d": "y", "e": "unparsable"},
                        {"a": "x", "b": "unparsable", "c": "z", "d": "y", "e": "unparsable"},
                        {"c": "unparsable", "e": "y"}, ["a", "b", "c", "d", "e"])
    check("consensus: A=B parsable agrees; identical unparsable goes to C (None without C, C's label with it); unparsable C is None",
          cons == {"a": "x", "b": None, "c": None, "d": "y", "e": "y"})
    check("disagree: differing labels or either unparsable (incl. both) => C adjudicates",
          va.disagree("x", "y") and va.disagree("unparsable", "unparsable") and va.disagree("x", "unparsable")
          and not va.disagree("x", "x"))
    # po = 1/2 (the identical unparsable pair is not agreement); pe = (1*1)/4 over parsable labels => kappa = (0.5-0.25)/0.75
    check("kappa: identical `unparsable` is not agreement and adds nothing to chance agreement",
          abs(va.kappa(["x", "unparsable"], ["x", "unparsable"]) - 1 / 3) < 1e-12
          and va.kappa(["x", "y"], ["x", "y"]) == 1.0 and va.kappa(["unparsable"] * 4, ["unparsable"] * 4) == 0.0)
    # ---- validate.gates() on synthetic samples and annotations
    old = (va.HERE, va.VAL)
    tmp = Path(tempfile.mkdtemp())
    va.HERE, va.VAL = tmp, tmp / "validation"
    va.VAL.mkdir()
    wl = lambda name, rows: (va.VAL / name).write_text("".join(json.dumps(r) + "\n" for r in rows))
    types = ["fix", "feature", "refactor", "other_dev"]
    wl("type-sample.jsonl", [dict(id=f"T{i:03d}", pipeline=types[i % 4]) for i in range(40)])
    pairs = [dict(id=f"P{i:03d}", stratum="jev_pos" if i < 30 else "jev_neg") for i in range(60)]
    wl("pair-sample.jsonl", pairs)
    (va.VAL / "pair-strata.json").write_text(json.dumps(dict(jev_pos=dict(population=300, sampled=30), jev_neg=dict(population=3000, sampled=30))))
    deps = [dict(id=f"D{i:03d}", partner_key=f"K{i:03d}", lang="python" if i % 2 == 0 else "go", proxy="hidden" if i < 20 else "explained") for i in range(30)]
    wl("dep-sample.jsonl", deps)
    # annotators: G4 A=B=pipeline except 4 items (A != B, C decides = pipeline); 2 items identical `unparsable` (unresolved)
    ann = {"A": {}, "B": {}, "C": {}}
    for i in range(40):
        iid, truth = f"T{i:03d}", types[i % 4]
        a = b = truth
        if i in (0, 1, 2, 3):
            a, ann["C"][iid] = "feature" if truth != "feature" else "fix", truth
        if i in (10, 11):
            a = b = "unparsable"
        ann["A"][iid], ann["B"][iid] = a, b
    # G5: 30 jev_pos (24 corrective by consensus, 6 independent); 30 jev_neg (3 corrective, 27 independent)
    for i in range(60):
        iid = f"P{i:03d}"
        lab = "corrective" if (i < 24 or 30 <= i < 33) else "independent"
        ann["A"][iid] = ann["B"][iid] = lab
    # G7: hidden items: python 8/10 no (D000, D002 yes), then D004 unparsable twice -> 7 no / 2 yes / 1 unresolved; go 7/10 no; explained: 8/10 yes
    for i in range(30):
        iid = f"D{i:03d}"
        if i < 20:
            lab = "no" if (i % 2 == 0 and i != 0 and i != 2) or (i % 2 == 1 and i not in (1, 3, 5)) else "yes"
        else:
            lab = "yes" if i < 28 else "no"
        ann["A"][iid] = ann["B"][iid] = lab
    ann["A"]["D004"] = ann["B"]["D004"] = "unparsable"
    for L in "ABC":
        wl(f"annotations-{L}.jsonl", [dict(id=k, label=x, model_resolved="m", prompt_sha256="0") for k, x in ann[L].items()])
    try:
        import io
        import contextlib
        with contextlib.redirect_stdout(io.StringIO()):
            va.gates()
        out = json.loads((tmp / "validation-gates.json").read_text())
    finally:
        va.HERE, va.VAL = old
    g4, g5, g7 = out["G4"], out["G5"], out["G7"]
    check("G4: n, unresolved share (2 identical unparsable of 40 = 0.05) and accuracy over resolved items (all pipeline-consistent)",
          g4["n"] == 40 and abs(g4["unresolved_share"] - 0.05) < 1e-9 and g4["accuracy"] == 1.0 and g4["pass"])
    # IPW: pos weight 10, neg weight 100; tp = 24*10, fp = 6*10, fn = 3*100 -> precision 0.8, recall 240/540
    check("G5: Jev precision and IPW recall use stratum weights; consensus positives counted",
          abs(g5["jev_precision"] - 0.8) < 1e-9 and abs(g5["jev_recall_ipw"] - round(240 / 540, 3)) < 1e-3
          and g5["consensus_positives"] == 27 and not g5["pass"])
    check("G7: cell sizes reported; unparsable counted in the unresolved share (1/30) and excluded from precision",
          g7["cell_sizes"] == {"python/hidden": 10, "go/hidden": 10, "python/explained": 5, "go/explained": 5}
          and abs(g7["cannot_tell_or_unresolved_share"] - round(1 / 30, 3)) < 1e-3 and g7["by_language"]["python"]["n"] == 9)
    check("G7: hidden precision by language is computed per language group", "python" in g7["by_language"] and "go" in g7["by_language"])
    test_g7_pairs(va)
    test_g1()
    # ---- estimate(): a test with an empty complete-case sample is skipped (django-style), others still run
    df = harness.synth(np.random.default_rng(1), n_repos=2, size=100)
    d2 = df.copy()
    d2["y_active"] = np.nan
    A = an.encode(d2)
    est = an.estimate(A, np.arange(len(d2)), A["area"])
    check("estimate: empty y_active sample => T1-T4 skipped, T5 still estimated", "T5" in est and not any(t in est for t in ("T1", "T2", "T3", "T4")))
    # ---- ridge: lambda 0 vs 1e-2 changes the T5-type coefficient by < 2 % on n = 4,000
    rng = np.random.default_rng(3)
    n = 4000
    X = rng.normal(size=(n, 3))
    y = (rng.random(n) < 1 / (1 + np.exp(-(-2.2 + 0.3 * X[:, 0] + 0.2 * X[:, 1])))).astype(float)
    fe = [rng.integers(0, 8, n)]
    b0, b1 = an.logit(y, X, fe, lam=0.0), an.logit(y, X, fe)
    check("logit ridge 1e-2 vs lambda=0 changes the coefficient by < 2 % (n = 4,000)", abs(b1 - b0) / abs(b0) < 0.02)
    # ---- analysis.main end to end on a synthetic panel with synthetic power / validation / gates files
    d = Path(tempfile.mkdtemp())
    df.to_csv(d / "panel.csv.gz", index=False)
    gate = lambda ok: dict(G3=dict(pass_=True), G3b=dict(pass_=ok))
    ok_val = dict(G4={"pass": True}, G5={"pass": True}, G7={"pass": True})
    (d / "power.json").write_text(json.dumps(dict(tests={t: dict(adequate=True) for t in an.FAMILY})))
    (d / "val.json").write_text(json.dumps(ok_val))
    (d / "g_ok.json").write_text(json.dumps(gate(True)))
    (d / "g_bad.json").write_text(json.dumps(gate(False)))
    (d / "g_none.json").write_text(json.dumps(dict(G3=dict(pass_=True))))
    run = lambda g, out="out.json": (setattr(sys, "argv", ["analysis.py", str(d / "panel.csv.gz"), str(d / out), "--power", str(d / "power.json"),
                                                           "--validation", str(d / "val.json"), "--gates", str(d / g)]), an.main())
    run("g_ok.json")
    o = json.loads((d / "out.json").read_text())
    check("analysis.main: all five tests estimated and given verdicts, gates recorded",
          all(t in o["verdicts"] for t in an.FAMILY) and o["gates"]["G3b"]["pass_"] and set(o["hypotheses"]) >= {"H1", "H2", "H4"})
    run("g_bad.json", "out2.json")
    o2 = json.loads((d / "out2.json").read_text())
    check("analysis.main: G3b failure => T1-T4 outcome_construct_invalid in the written results",
          all(o2["verdicts"][t]["verdict"].startswith("outcome_construct_invalid") for t in ("T1", "T2", "T3", "T4")))
    try:
        run("g_none.json", "out3.json")
        missing_exit = False
    except SystemExit as e:
        missing_exit = "G3b" in str(e)
    check("analysis.main: gates.json without G3b is refused", missing_exit)


def test_timing():
    import follow
    P = dict(created=0.0, merged=1000.0)
    mk = lambda c, m: dict(created=c, merged=m)
    check("follow-up: created after parent merge and merged within 14 d", follow.is_follow(P, mk(1500.0, 2000.0), False))
    check("follow-up: created BEFORE parent merge is not a follow-up (concurrent work)", not follow.is_follow(P, mk(900.0, 2000.0), False))
    check("follow-up: merged before parent (negative gap) rejected", not follow.is_follow(P, mk(1500.0, 900.0), False))
    check("follow-up: beyond 14 d rejected", not follow.is_follow(P, mk(1500.0, 1000.0 + 15 * 86400), False))
    check("follow-up: the parent itself rejected", not follow.is_follow(P, mk(1500.0, 2000.0), True))


if __name__ == "__main__":
    test_pure()
    test_collect()
    test_timing()
    test_round2()
    test_round4()
    print("FAILED:" if FAILS else "all tests passed", FAILS or "")
    sys.exit(1 if FAILS else 0)
