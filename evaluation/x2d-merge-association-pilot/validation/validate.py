#!/usr/bin/env python3
"""X2d implementation validation (CLAUDE.md experiment rule 4; PREREGISTRATION section 11). Synthetic data, stubs and fake CLIs only: no model call, no real data,
no Jev key, no Clawpatch map. Every check prints PASS/FAIL; the script ends with a count and "all validation checks passed" (exit 1 otherwise).

Parts (the letter prefixes the check names):
  P  provenance and pins: X2 pins, Jev pin and its drift guard, frozen-file list, thresholds.json vs x2d constants, question files vs GUIDE.md, round table.
  F  front end: rules (a)-(g) on a synthetic git history (variant stems, directory normalisation, the empty-file rename trap, markdown/translation policy, a #9268-scale
     cluster, the PR cap's deterministic selection), the model-free pre-pass and scopes reproduce their committed files byte for byte.
  K  onboarding consensus: planted draw sets at two granularities recover their clusters at tau = 0.5; permuted memberships yield none above chance.
  J  the Jev driver and the onboarding driver against a stub endpoint: request shape, pin, key hygiene, resume/append-only, 429 retry, cost stop, failure rows.
  L  the labeller harness on stub sessions: blindness, identical evidence, adjudication, append-only, refusals; the reference arm.
  S  statistics: the GO / INCONCLUSIVE / NO-GO rule and its precedence, the primary-arm switch and core-gate exclusion through verdict(), perception (both question rows, any order), joint-answer kappa, N1 on D u P, the
     futility rule, P2 ablation, P3 planted effect, N2 label permutation, N3 and N3-PR planted nulls (power report schema x2d-power-v2).
  R  round machinery: units, baselines, gold, membership, tuning, composition, versioning and the refinement-envelope refusal on synthetic data.
  G  round-0 machinery on synthetic data: controls, P1/N1/K gates.
  G13 rev 13 (R-053) re-gate of P1 and P2 on fresh items (validation/regate_checks.py): planted effect and planted null of the rebuilt gates, direct-label versus derived P2 gold, the leak fix, fresh-pool exclusions, build refusal on a short
     supply, one shot, round 1 refusing without a re-gate PASS, the budget charge, the frozen Jev side, an end-to-end re-gate on stubs and the unchanged round-0 bytes.
  E  end to end: round 0, all ten rounds and the verdict through run.sh on stubs (stub Clawpatch, stub Jev endpoint, stub labellers): critique/version enforcement in a temporary X2D_REVIEWS_DIR,
     controls, contrast, labels-check, replayed baselines, sessions, a planted effect, a planted null and a flipped K5 primary arm. Only the named K3 gate is overridden.
Usage (repo root, toolchain active): .tools/venv/bin/python evaluation/x2d-merge-association-pilot/validation/validate.py
Environment: X2D_SRC (a fastapi clone containing the pinned revisions; default /tmp/scout-x2c-smoke/_src/fastapi) for the pre-pass reproduction and the end-to-end run.
"""
import hashlib
import collections
import copy
import json
import math
import os
import random
import re
import shutil
import subprocess
import sys
import tempfile

HERE = os.path.dirname(os.path.abspath(__file__))
S = os.path.dirname(HERE)
sys.path.insert(0, S)
sys.path.insert(0, HERE)
os.environ.setdefault("X2D_SKIP_SIGNOFF", "1")
os.environ.setdefault("X2D_JEV_ENDPOINT", "http://127.0.0.1:9/")  # R2-16: X2D_SKIP_SIGNOFF is honoured only with a loopback endpoint and an X2D_DATA outside S/data
_live = os.path.join(S, "data")  # run.sh exports X2D_DATA=S/data when unset: validation never points a stage at the live data directory (R2-16)
if not os.environ.get("X2D_DATA") or os.path.abspath(os.environ["X2D_DATA"]).startswith(_live):
    os.environ["X2D_DATA"] = os.path.join(__import__("tempfile").gettempdir(), "x2d-validate-default-data")
import x2d  # noqa: E402 (puts X2's directory and the study directory on sys.path)
import common as C  # noqa: E402 (X2's frozen common.py)
import frontend as F  # noqa: E402

ROOT, PY = x2d.ROOT, sys.executable
SRC = os.environ.get("X2D_SRC") or "/tmp/scout-x2c-smoke/_src/fastapi"
KEY = "sk-or-v1-" + "validation-stub-key-0123456789"  # a fake key, only ever given to a driver subprocess and the stub
FAILS = []
COUNT = [0]


def check(name, cond, detail=""):
    COUNT[0] += 1
    ok = bool(cond)
    print(("PASS " if ok else "FAIL ") + name + ((" :: " + str(detail)[:600]) if (not ok and detail != "") else ""))
    if not ok:
        FAILS.append(name)


def run(cmd, cwd=None, env=None, inp=None, timeout=1800):
    return subprocess.run(cmd, cwd=cwd, env=env, input=inp, capture_output=True, text=True, timeout=timeout)


def clean_env(**extra):
    env = {k: v for k, v in os.environ.items() if not C.is_credential(k)}
    env["X2D_SKIP_SIGNOFF"] = "1"
    env.setdefault("X2D_JEV_ENDPOINT", "http://127.0.0.1:9/")
    env["X2D_DATA"] = os.environ["X2D_DATA"]
    env.update(extra)
    return env


_BIN = {}


def go_bin(tmp, name, pkg):
    if name not in _BIN:
        p = os.path.join(tmp, name)
        r = subprocess.run(["go", "build", "-o", p, pkg], cwd=ROOT, capture_output=True, text=True)
        if r.returncode != 0:
            print("go build %s failed:\n%s" % (pkg, r.stderr[-2000:]))
            _BIN[name] = None
        else:
            _BIN[name] = p
    return _BIN[name]


def jl(path):
    return x2d.read_jsonl(path)


def wjl(path, rows):
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        for r in rows:
            f.write(json.dumps(r, sort_keys=True) + "\n")



def make_inputs(data, r):
    """rounds/<r>/inputs from the study base (GUIDE.md + questions/*.json), whatever the previous round holds (fixtures that skip rounds). -> the inputs directory."""
    d = x2d.inputs_dir(r, data)
    if not os.path.isdir(d):
        os.makedirs(d)
        shutil.copytree(os.path.join(S, "questions"), os.path.join(d, "questions"))
        shutil.copyfile(os.path.join(S, "GUIDE.md"), os.path.join(d, "GUIDE.md"))
    return d


def squash(s):
    return re.sub(r"\s+", " ", s).strip()


# ---------------------------------------------------------------- synthetic git history
class Repo:
    """A throw-away git repository. `commit(subject, writes, deletes, renames)` makes one first-parent commit with a fixed date."""

    def __init__(self, path):
        self.path, self.n = path, 0
        os.makedirs(path, exist_ok=True)
        self.git("init", "-q", "-b", "master")
        self.git("config", "user.name", "x2d")
        self.git("config", "user.email", "x2d@example.invalid")
        self.git("config", "commit.gpgsign", "false")

    def git(self, *a, check=True):
        r = subprocess.run(["git", "-C", self.path, *a], capture_output=True, text=True, env=dict(os.environ, GIT_AUTHOR_DATE=self.date(), GIT_COMMITTER_DATE=self.date()))
        if check and r.returncode != 0:
            raise RuntimeError("git %s: %s" % (" ".join(a), r.stderr))
        return r.stdout.strip()

    def date(self):
        return "2025-01-%02dT12:00:00+00:00" % (1 + (self.n % 27))

    def commit(self, subject, writes=None, deletes=(), renames=None, body=""):
        for old, new in (renames or {}).items():
            os.makedirs(os.path.dirname(os.path.join(self.path, new)) or self.path, exist_ok=True)
            self.git("mv", old, new)
        for p, text in (writes or {}).items():
            fp = os.path.join(self.path, p)
            os.makedirs(os.path.dirname(fp), exist_ok=True)
            with open(fp, "w") as f:
                f.write(text)
            self.git("add", p)
        for p in deletes:
            self.git("rm", "-q", p)
        self.n += 1
        self.git("commit", "-q", "--allow-empty", "-m", subject + ("\n\n" + body if body else ""))
        return self.git("rev-parse", "HEAD")


def body_text(tag, n=30):
    return "".join("line %d of %s: value = compute(%d)\n" % (i, tag, i * 7) for i in range(n))


def py_text(tag, funcs=("handler", "helper")):
    return '"""%s"""\n\n' % tag + "".join("def %s_%s(x):\n    return x + %d\n\n" % (f, tag.replace("-", "_").replace("/", "_").replace(".", "_"), i) for i, f in enumerate(funcs)) + body_text(tag, 10)


# ---------------------------------------------------------------- P provenance
def provenance_checks(tmp):
    check("P every reused X2 file (common.py, label.py, stub_clawpatch.js, credcheck) has the digest X2's sign-off names", not [b for b in x2d.verify_pins(require_signoff=False) if b.startswith("X2 ")],
          x2d.verify_pins(require_signoff=False))
    check("P the Jev pin of internal/policy/consent.go is the preregistered typesafe/jev-1.14-20261014", x2d.go_jev_pin() == x2d.JEV_MODEL == "typesafe/jev-1.14-20261014")
    bad = os.path.join(tmp, "consent.go")
    open(bad, "w").write(open(x2d.GO_POLICY).read().replace(x2d.JEV_MODEL, "typesafe/jev-1.15-20270101"))
    check("P a changed Jev pin is caught (drift guard names the changed constant)", x2d.go_jev_pin(bad) == "typesafe/jev-1.15-20270101" != x2d.JEV_MODEL)
    probs = x2d.verify_pins(require_signoff=True)
    check("P a stage that writes data refuses without the council's sign-off record (verify_pins names the missing record)", any("no sign-off record" in p for p in probs) or os.path.exists(x2d.SIGNOFF), probs)
    ff = x2d.frozen_files()
    mods = ["x2d.py", "collect_onboard.py", "onboard/main.go", "quality.py", "build_units.py", "b2.py", "baselines.py", "build_controls.py", "label_x2d.py", "gold.py", "jev/main.go",
            "reference_arm.py", "tune.py", "compose.py", "analysis_x2d.py", "power_sim.py", "versioning.py", "run.sh", "frontend.py", "prepass.py", "scopes.py", "GUIDE.md", "thresholds.json",
            "validation/validate.py", "validation/stub_jev.py", "validation/stub_labeller.py", "validation/stub_clawpatch_x2d.js", "prepass/rounds.json", "prepass/commits.jsonl",
            "questions/merge.covered@v1.json", "measurement/lever_check.py"]
    check("P the frozen-file list covers every module of the README contract, run.sh, validation/*, questions, prepass data and measurement", all(m in ff for m in mods), [m for m in mods if m not in ff])
    # sign-off record: a tampered frozen file and a tampered code-under-test blob are both refused
    rec_dir = os.path.join(tmp, "signoff")
    os.makedirs(rec_dir)
    shutil.copy(os.path.join(S, "GUIDE.md"), os.path.join(rec_dir, "GUIDE.md"))
    old = x2d.SIGNOFF
    try:
        x2d.SIGNOFF = os.path.join(tmp, "signoff.json")
        tune_sha = x2d.sha256_file(x2d.TUNE_SCRIPT) if os.path.exists(x2d.TUNE_SCRIPT) else "0" * 64
        all_files = {rel: x2d.sha256_file(os.path.join(S, rel)) for rel in x2d.frozen_files()}
        json.dump({"files": all_files, "code_under_test": go_digest_subset(), "tune_script_sha256": tune_sha}, open(x2d.SIGNOFF, "w"))
        ok_probs = [p for p in x2d.verify_pins(require_signoff=True) if not p.startswith("X2 ")]
        json.dump({"files": {k: v for k, v in all_files.items() if k != "GUIDE.md"}, "code_under_test": go_digest_subset(), "tune_script_sha256": tune_sha}, open(x2d.SIGNOFF, "w"))
        added_probs = [p for p in x2d.verify_pins(require_signoff=True) if not p.startswith("X2 ")]
        json.dump({"files": all_files, "code_under_test": go_digest_subset(), "tune_script_sha256": tune_sha}, open(x2d.SIGNOFF, "w"))
        open(os.path.join(rec_dir, "GUIDE.md"), "a").write("tampered")
        bad_files = x2d.verify_pins(require_signoff=True, root=rec_dir)
        json.dump({"files": all_files, "code_under_test": dict(go_digest_subset(), **{"go.mod": "0" * 64}), "tune_script_sha256": tune_sha}, open(x2d.SIGNOFF, "w"))
        bad_code = x2d.verify_pins(require_signoff=True)
    finally:
        x2d.SIGNOFF = old
    check("P the sign-off record is honoured: matching digests give no problem; a changed frozen file and a changed code-under-test blob are each refused",
          not ok_probs and any("frozen file GUIDE.md" in p for p in bad_files) and any("code under test go.mod" in p for p in bad_code), (ok_probs, bad_files, bad_code))
    check("P R2-16 a frozen file added after sign-off (present on disk, absent from the record) is refused by verify_pins", any("frozen file set differs" in p and "GUIDE.md" in p for p in added_probs), added_probs)
    dig = x2d.go_code_digests()
    need = ["go.mod", "go.sum", "internal/policy/consent.go", "internal/areas/areas.go", "internal/clawpatch/records.go", "internal/clawpatch/client.go", "scripts/clawpatch-safe.patch",
            "scripts/build-clawpatch.sh"]
    check("P the code-under-test digests cover every file PREREGISTRATION section 2 names plus every non-test file of internal/decisions (incl. orderflip and thresholds)",
          all(n in dig for n in need) and any(k.startswith("internal/decisions/") for k in dig) and not any(k.endswith("_test.go") for k in dig)
          and any("orderflip/orderflip.go" in k for k in dig) and any("thresholds/thresholds.go" in k for k in dig), sorted(dig))


def go_digest_subset():
    d = x2d.go_code_digests()
    return {"go.mod": d["go.mod"]}


def thresholds_checks():
    th = json.load(open(os.path.join(S, "thresholds.json")))
    check("P thresholds.json minimums, PR cap, round table, caps and seed equal the x2d constants", th["minimums"]["P_assign"] == x2d.MINIMUMS["P_assign"] == 0.90
          and th["minimums"]["R_assign"] == x2d.MINIMUMS["R_assign"] and th["study"]["pr_cap"] == x2d.PR_CAP == 10 and th["study"]["round_size"]["n_rounds"] == x2d.N_ROUNDS
          and th["study"]["round_size"]["frozen_rounds"] == x2d.FROZEN_ROUNDS and th["caps"]["usd_per_round"] == x2d.CAPS["usd_per_round"] == 2.0 and th["caps"]["usd_total"] == x2d.CAPS["usd_total"] == 20.0
          and th["caps"]["clawpatch_maps_round_0"] == x2d.CAPS["clawpatch_maps_round_0"] == 16 and "cli_sessions_round_0" not in th["caps"] and "cli_sessions_per_round" not in th["caps"]
          and "cli_sessions_round_0" not in x2d.CAPS and "cli_sessions_per_round" not in x2d.CAPS
          and "20261029" in th["tuning"]["tool"] and x2d.SEED == 20261029)
    check("P thresholds.json study.go caps, minimums (P_assign / R_assign only: P_new and R_new are no conditions), study.power.bars equal analysis_x2d.GO, x2d (UNRESOLVED_RATE_MAX, RESOLVED_MIN, NEW_AREA_GATE_MIN) and power_sim.BARS (R4-13 (9))",
          th["study"]["go"]["unresolved_rate_max"] == x2d.UNRESOLVED_RATE_MAX == A.GO["unresolved_rate_max"] == 0.15 and th["study"]["go"]["resolved_min"] == x2d.RESOLVED_MIN == A.GO["resolved_min"] == 50
          and "P_new" not in th["minimums"] and "R_new" not in th["minimums"] and set(x2d.MINIMUMS) == {"P_assign", "R_assign"} and x2d.NEW_AREA_GATE_MIN == 0.80
          and th["study"]["power"]["bars"] == PS.BARS and th["study"]["power"]["schema"] == "x2d-power-v5" and A.GO["gain_min"] == 0.10 and A.GO["p_assign_wilson_lb"] == 0.80 and A.GO["spread_min"] == 3
          and th["study"]["go"]["guard_level"] == x2d.GO_GUARD_LEVEL == A.GO["guard_level"] == 0.95)
    check("P no decision in thresholds.json has a 0.5 threshold and every tuned decision's provisional value is 0.7", all(d["threshold"] in (0.7, None) for d in th["decisions"]))
    rj = json.load(open(os.path.join(S, "prepass", "rounds.json")))
    sc = [r["scored_clusters"] for r in rj["rounds"]]
    check("P the round table has 10 rounds of >= 20 scored clusters from >= 6 PRs, 91 scored clusters from 40 PRs in rounds 8-10 (round 10 = 48 clusters / 20 PRs) and no spare PRs (R1-02)",
          len(sc) == 10 and min(sc) >= 20 and min(r["eligible_prs"] for r in rj["rounds"]) >= 6 and rj["frozen_8_10"] == {"scored_clusters": 91, "prs": 40} and rj["spare_after_round_10"] == []
          and sc[-1] == 48 and rj["rounds"][-1]["eligible_prs"] == 20, sc)
    check("P every scored id of rounds.json is capped at 10 per PR", all(p["scored"] == min(p["clusters"], 10) and len(p["scored_ids"]) == p["scored"] for r in rj["rounds"] for p in r["prs"]))


def guide_checks():
    guide = open(os.path.join(S, "GUIDE.md")).read()
    g = squash(guide)
    qc = json.load(open(os.path.join(S, "questions", "merge.covered@v1.json")))
    check("P merge.covered's criteria are the guide's Q1 words for words (true and false)", squash(qc["criteria"]["true"]).replace("`", "") in g.replace("`", "")
          and squash(qc["criteria"]["false"]) in g, qc["criteria"])
    check("P merge.covered's instructions carry the guide's question sentence and its five checklist rules (test, example, infrastructure, variant, package marker)",
          "Would this repository's maintainers file an issue about the files" in qc["instructions"] and "Would this repository's maintainers file an issue about the files" in g
          and all(w in qc["instructions"] for w in ("a test belongs with the responsibility it tests", "a documentation example belongs", "repository-wide automation", "new variant", "package markers")))
    qw = json.load(open(os.path.join(S, "questions", "merge.which_area@v1.json")))
    check("P merge.which_area carries the guide's Q2 rule (narrowest area whose responsibility covers the files) and the guide's R2/R3 in its instructions",
          "Choose the narrowest area whose responsibility covers them" in g and "narrowest area whose responsibility covers them" in qw["instructions"]
          and "A test belongs with the responsibility it tests" in qw["instructions"] and "a documentation example belongs with the documented responsibility" in qw["instructions"])
    names = ["merge.covered", "merge.which_area", "merge.area_covers", "onboard.covered", "onboard.which_file_area", "onboard.same_responsibility"]
    qs = {n: json.load(open(os.path.join(S, "questions", n + "@v1.json"))) for n in names}
    check("P the six question files exist, are v1, name the pinned model where they carry one and merge.which_area's V-title shape holds only the title",
          all(q["question_id"] == n and q["label_version"] == "v1" and q.get("model", x2d.JEV_MODEL) == x2d.JEV_MODEL for n, q in qs.items())
          and "-state.pr.body" in qw["variants"]["V-title"]["state_shape_delta"] and "+state.siblings" in qw["variants"]["V-sib"]["state_shape_delta"])
    check("P the guide's version string is the preregistered x2d-association-v1/guide-1 and GUIDE.md names it", x2d.GUIDE_VERSION in guide)


# ---------------------------------------------------------------- F front end
def frontend_checks(tmp):
    check("F variant stems: _an, _py3NN, _pv1/_pv2 and a trailing _NN are stripped repeatedly (tutorial003_an_py39 -> tutorial003, test_tutorial003_py310 -> test_tutorial003, tutorial003_01 -> tutorial003)",
          F.variant_stem("tutorial003_an_py39") == "tutorial003" and F.variant_stem("test_tutorial003_py310") == "test_tutorial003" and F.variant_stem("tutorial003_01") == "tutorial003"
          and F.variant_stem("main_pv2") == "main" and F.variant_stem("tutorial003_an_py39_02") == "tutorial003" and F.variant_stem("tutorial003") == "tutorial003")
    check("F directory normalisation strips the same variant suffixes from every segment (app_b_an -> app_b; test_x_py39 -> test_x); an unvaried path is unchanged",
          F.norm_dir("docs_src/app_testing/app_b_an") == "docs_src/app_testing/app_b" and F.norm_dir("tests/test_tutorial/test_x_py39") == "tests/test_tutorial/test_x"
          and F.norm_dir("fastapi/security") == "fastapi/security")
    check("F cluster key and sibling key: variants of one example share a key; a different extension differs only in the sibling key",
          F.cluster_key("docs_src/t/tutorial001_an_py39.py") == F.cluster_key("docs_src/t/tutorial001.py") == ("docs_src/t", "tutorial001")
          and F.sibling_key("docs_src/t/tutorial001.py") != F.sibling_key("docs_src/t/tutorial001.json"))
    check("F policy (e): any .md, docs/<lang>/ with lang != en and binary media are outside the population; docs/en/** non-markdown and code are eligible",
          F.excluded_reason("docs/en/docs/x.md") == "markdown" and F.excluded_reason("README.md") == "markdown" and F.excluded_reason("docs/de/docs/x.yml") == "translation"
          and F.excluded_reason("docs/zh-hant/mkdocs.yml") == "translation" and F.excluded_reason("docs/en/img/a.PNG") == "media" and F.excluded_reason("docs/en/mkdocs.yml") is None
          and F.excluded_reason("fastapi/a.py") is None)
    check("F roles and coarse scopes: core, example, test, infra, docs-config and config", F.role("fastapi/a.py") == "core" and F.role("docs_src/x/t.py") == "example" and F.role("tests/t.py") == "test"
          and F.role(".github/w.yml") == "infra" and F.role("scripts/x.py") == "infra" and F.role("docs/en/mkdocs.yml") == "docs-config" and F.role("pyproject.toml") == "config"
          and F.scope("fastapi/a.py") == "core" and F.scope("tests/t.py") == F.scope("docs_src/x.py") != F.scope("scripts/x.py"))

    repo = Repo(os.path.join(tmp, "fe-repo"))
    base = {"fastapi/a.py": py_text("a"), "fastapi/b.py": body_text("b", 40), "fastapi/util.py": body_text("util", 40), "docs_src/topic/tutorial001.py": py_text("t1"),
            "tests/test_tutorial/test_topic/test_tutorial001.py": py_text("tt1"), "docs/en/docs/x.md": "# x\n", "tests/test_tutorial/test_annotated/__init__.py": "",
            "docs_src/app_testing/app_b/main.py": py_text("appb"), "README.md": "# r\n", "docs_src/pkg/__init__.py": "", "tests/test_tutorial/test_old.py": py_text("old"),
            "docs_src/known/tutorial001.py": py_text("known"), "docs_src/other/tutorial001.json": "{}\n"}
    c0 = repo.commit("base", base)
    writes = {"docs_src/topic/tutorial001_an.py": py_text("t1an"), "docs_src/topic/tutorial001_an_py39.py": py_text("t1an39"), "docs_src/newtopic/tutorial001.py": py_text("nt1"),
              "docs_src/newtopic/tutorial001_an.py": py_text("nt1an"), "docs/en/docs/y.md": "# y\n", "docs/de/docs/z.yml": "a: b\n", "docs/en/img/a.png": "PNG",
              "fastapi/a.py": py_text("a") + "def more():\n    return 1\n", "docs_src/pkg/sub/marker.py": "", "docs_src/pkg/other/marker2.py": ""}
    c1 = repo.commit("Add things (#101)", writes, deletes=["docs_src/app_testing/app_b/main.py"],
                     renames={"fastapi/b.py": "fastapi/c.py", "fastapi/util.py": "docs_src/util.py", "tests/test_tutorial/test_annotated/__init__.py": "docs_src/app_testing/app_b_an/__init__.py"})
    entries = F.diff_entries(repo.path, c0, c1)
    state = F.TreeState(F.tree_paths(repo.path, c0))
    counts, residual = F.classify(entries, state)
    rp = sorted(r[0] for r in residual)
    want = {"modified": 1, "deleted": 1, "excluded:markdown": 1, "excluded:translation": 1, "excluded:media": 1, "renamed_inherit": 1, "variant_inherit": 2, "residual_added": 4,
            "renamed_cross_scope": 1}
    check("F catalogue-free rules (a)-(e) on a synthetic history: modified 1, deleted 1, markdown/translation/media each excluded once, same-scope rename inherits, two variants inherit, cross-scope rename is residual",
          all(counts.get(k) == v for k, v in want.items()), dict(counts))
    check("F the empty-file rename trap: git reports the empty __init__.py move across directories as a 100 % rename; the front end does not let it inherit and keeps it residual (`renamed_empty_trap`)",
          any(e[0] == "R" and e[1] == 100 and e[4] == F.EMPTY_BLOB for e in entries) and counts.get("renamed_empty_trap") == 1
          and "docs_src/app_testing/app_b_an/__init__.py" in rp, (dict(counts), [(e[0], e[1], e[2], e[3]) for e in entries]))
    blobs = {e[3]: e[4] for e in entries if e[0] in "ARC"}
    live, inert = F.clusters(residual, blobs)
    live_keys = sorted((c["dir"], c["stem"]) for c in live)
    check("F rule (g): clusters made only of empty blobs (package markers, the trapped rename) are dropped and counted; the live clusters are the new topic (2 variants in one cluster) and the cross-scope move",
          len(inert) >= 2 and ("docs_src/newtopic", "tutorial001") in live_keys and ("docs_src", "util") in live_keys and not any(c["stem"] == "__init__" for c in live)
          and next(c for c in live if c["stem"] == "tutorial001")["files"] == ["docs_src/newtopic/tutorial001.py", "docs_src/newtopic/tutorial001_an.py"], (live_keys, len(inert)))
    check("F cluster ids are deterministic and depend on commit, directory and stem; a cluster-id change would change the PR-cap order",
          F.cluster_id(c1, live[0]) == F.cluster_id(c1, dict(live[0])) and F.cluster_id(c1, live[0]) != F.cluster_id(c0, live[0]) and len(F.cluster_id(c1, live[0])) == 16)

    # ownership oracle (rules (a), (d), (f), still-unowned modified files)
    mem = {"areas": {"A": ["fastapi/a.py", "docs_src/topic/tutorial001.py", "fastapi/b.py"], "B": ["tests/test_tutorial/test_topic/test_tutorial001.py", "docs_src/known/tutorial001.py"]},
           "pool": [], "heuristic_only": ["docs_src/newtopic/tutorial001.py", "fastapi/zzz.py"]}
    own = F.Ownership(mem)
    resolved = []
    oc, ores = F.owned_classify(entries, state, own, resolved=resolved)
    check("F owned rules: a modified owned file keeps its owner (a); owned variants inherit (d) with the owner; a same-scope rename of an owned file inherits (c); the unowned heuristic-only package is residual and counted (f)",
          oc.get("modified_owned") == 1 and oc.get("variant_inherit") == 2 and oc.get("renamed_inherit") == 1 and oc.get("heuristic_only", 0) >= 1
          and {r["rule"] for r in resolved} == {"c", "d"} and all(r["owners"] == ["A"] for r in resolved), (dict(oc), resolved))
    writes2 = {"fastapi/a.py": py_text("a") + "# again\n", "docs_src/other/tutorial001_an.json": "{}\n", "docs_src/known/tutorial001_an.py": py_text("kan"), "fastapi/c.py": body_text("b", 40) + "x\n",
               "fastapi/d.py": body_text("d", 40), "docs_src/util2.py": body_text("u2", 40)}
    c2 = repo.commit("More (#102)", writes2)
    e2 = F.diff_entries(repo.path, c1, c2)
    st2 = F.TreeState(F.tree_paths(repo.path, c1))
    oc2, ores2 = F.owned_classify(e2, st2, own)
    check("F rule (d) needs an OWNED sibling: the variant of an owned example inherits; the variant of an unowned base file is residual (`variant_unowned`); a still-unowned modified file is NOT a unit (R1-04): "
          "counted `modified_unowned` (and `modified_unowned:never_owned`), never residual",
          oc2.get("variant_inherit") == 1 and oc2.get("variant_unowned") == 1 and oc2.get("modified_unowned") == 1 and oc2.get("modified_unowned:never_owned") == 1
          and oc2.get("modified_unowned:pooled", 0) == 0 and oc2.get("modified_owned") == 1
          and any(r[0] == "docs_src/other/tutorial001_an.json" for r in ores2) and not any(r[0] == "fastapi/c.py" for r in ores2), (dict(oc2), ores2))
    own_p = F.Ownership(dict(mem, pool=["fastapi/c.py"]))
    oc3, ores3 = F.owned_classify(e2, st2, own_p)
    check("F a modified file in the membership pool counts `modified_unowned:pooled` (not never_owned) and is never residual", oc3.get("modified_unowned:pooled") == 1 and oc3.get("modified_unowned:never_owned", 0) == 0
          and not any(r[0] == "fastapi/c.py" for r in ores3), dict(oc3))
    check("F the ownership oracle: owners, sibling owners in the normalised directory and heuristic-only flag are computed from the membership layer alone",
          own.owners("fastapi/a.py") == {"A"} and own.owners("nope.py") == set() and own.is_heuristic_only("fastapi/zzz.py") and not own.is_heuristic_only("fastapi/a.py")
          and [s["path"] for s in own.sibling_owners("fastapi/new.py", {"fastapi/a.py", "fastapi/b.py"})] == ["fastapi/a.py", "fastapi/b.py"])

    # #9268-scale cluster and the PR cap's deterministic selection
    big = {}
    for i in range(347):
        big["docs_src/mig/tutorial%03d_an%s.py" % (i % 60, "" if i < 60 else "_py3%d" % (9 + i % 3))] = py_text("mig%d" % i, ("h",))
    for i in range(60):
        big["tests/test_tutorial/test_mig%02d/test_tutorial001.py" % i] = py_text("tmig%d" % i, ("t",))
    c3 = repo.commit("Migrate (#9268)", big)
    e3 = F.diff_entries(repo.path, c2, c3)
    st3 = F.TreeState(F.tree_paths(repo.path, c2))
    _, res3 = F.classify(e3, st3)
    live3, _ = F.clusters(res3, {e[3]: e[4] for e in e3})
    ids = [F.cluster_id(c3, c) for c in live3]
    capped = sorted(ids, key=x2d.cap_key)[:x2d.PR_CAP]
    check("F a #9268-scale PR (347 near-duplicate variants + 60 tests) collapses to clusters by (directory, variant stem), not per file, and the PR cap keeps exactly the 10 smallest sha256('x2d-cap\\0' + id)",
          len(live3) == 120 and len(ids) == len(set(ids)) and len(capped) == 10 and capped == sorted(ids, key=lambda i: hashlib.sha256(("x2d-cap\0" + i).encode()).hexdigest())[:10]
          and sorted(ids, key=x2d.cap_key)[:10] == sorted(reversed(ids), key=x2d.cap_key)[:10], len(live3))
    rj = json.load(open(os.path.join(S, "prepass", "rounds.json")))
    big_pr = max((p for r in rj["rounds"] for p in r["prs"]), key=lambda p: p["clusters"])
    check("F the committed round table applies the cap to the largest real PR (#14569, 79 clusters): 10 scored, the rest reported", big_pr["pr"] == 14569 and big_pr["clusters"] == 79 and big_pr["scored"] == 10)
    return repo, (c0, c1, c2, c3)


def prepass_checks(tmp):
    if not os.path.isdir(SRC):
        check("F the pre-pass and scopes reproduce byte for byte (needs the fastapi clone X2D_SRC)", False, "clone %s absent" % SRC)
        return
    out = os.path.join(tmp, "prepass")
    os.makedirs(out)
    r = run([PY, os.path.join(S, "prepass.py"), "--repo", SRC, "--out", out], timeout=1800)
    same = r.returncode == 0 and all(open(os.path.join(out, n), "rb").read() == open(os.path.join(S, "prepass", n), "rb").read()
                                      for n in ("summary.json", "rounds.json", "reserve.json", "k0.json", "commits.jsonl"))
    check("F prepass.py re-run reproduces prepass/{summary,rounds,reserve,k0}.json and commits.jsonl byte for byte (the front-end changes did not alter existing behaviour)", same, r.stderr[-500:])
    sc = os.path.join(tmp, "scopes.json")
    r = run([PY, os.path.join(S, "scopes.py"), "--repo", SRC, "--rev", x2d.C0, "--out", sc])
    same = r.returncode == 0 and open(sc, "rb").read() == open(os.path.join(S, "prepass", "scopes.json"), "rb").read()
    check("F scopes.py reproduces prepass/scopes.json byte for byte", same, r.stderr[-500:])
    sj = json.load(open(os.path.join(S, "prepass", "scopes.json")))
    check("F every scope fits the three caps (<= 400 files, <= 500 source files, <= 200 test files) and there are five scopes", len(sj["scopes"]) == 5 and sj["all_fit"]
          and all(s["files"] <= 400 and s["source"] <= 500 and s["tests"] <= 200 for s in sj["scopes"]), [(s["id"], s["files"], s["source"], s["tests"]) for s in sj["scopes"]])


# ---------------------------------------------------------------- K onboarding consensus (Go `onboard -consensus-only`)
def write_feature(draws, scope, k, fid, title, files, source="agent-mapper"):
    d = os.path.join(draws, scope, str(k), "features")
    os.makedirs(d, exist_ok=True)
    rec = {"schemaVersion": 1, "featureId": fid, "title": title, "summary": "Handles " + title, "kind": "module", "source": source, "confidence": "medium", "entrypoints": [],
           "ownedFiles": [{"path": p, "reason": "plant"} for p in files], "contextFiles": [], "tests": [], "status": "mapped"}
    with open(os.path.join(d, fid + ".json"), "w") as f:
        json.dump(rec, f)
    json.dump({"rc": 0}, open(os.path.join(draws, scope, str(k), "meta.json"), "w"))


def write_scopes(path, ids, include=None):
    json.dump({"scopes": [{"id": i, "include": (include or {}).get(i) or [i + "/"]} for i in ids]}, open(path, "w"))


def consensus(tmp, name, draws_spec, scope="core", k=3, include=None):
    """draws_spec: [ {fid: (title, files)} ] one dict per draw. -> (rc, catalogue dict|None, stderr, out dir). `include`: {scope id: [prefixes]} for scopes.json."""
    base = os.path.join(tmp, "cons-" + name)
    draws = os.path.join(base, "draws")
    for i, spec in enumerate(draws_spec, 1):
        for fid, (title, files) in spec.items():
            write_feature(draws, scope, i, "%s_%d" % (fid, i), title, files)
    sc = os.path.join(base, "scopes.json")
    write_scopes(sc, [scope], {scope: include} if include else None)
    out = os.path.join(base, "catalogue")
    r = run([_BIN["onboard"], "-draws", draws, "-scopes", sc, "-out", out, "-consensus-only", "-k", str(k)], cwd=ROOT)
    cat = json.load(open(os.path.join(out, "catalogue.json"))) if r.returncode == 0 else None
    return r.returncode, cat, r.stderr, out


def partition_of(cat):
    return sorted(sorted(a["files"]) for a in cat["areas"])


def ari(a, b):
    import quality as Q
    return Q.ari(a, b)


def consensus_checks(tmp):
    if go_bin(tmp, "onboard", "./evaluation/x2d-merge-association-pilot/onboard") is None:
        check("K the onboarding program builds", False)
        return
    files = ["core/f%02d.py" % i for i in range(24)]
    coarse = [files[i:i + 6] for i in range(0, 24, 6)]
    fine = [files[i:i + 3] for i in range(0, 24, 3)]
    mk = lambda groups, tag: {"%s%d" % (tag, j): ("%s area %d" % (tag, j), g) for j, g in enumerate(groups)}
    want = lambda groups: sorted(sorted(g) for g in groups)
    rc, cat, err, out = consensus(tmp, "coarse", [mk(coarse, "c")] * 3)
    check("K planted draws at the coarse granularity (4 areas of 6 files, three identical draws) are recovered exactly at tau = 0.5", rc == 0 and partition_of(cat) == want(coarse), err[-300:])
    rc, cat, err, _ = consensus(tmp, "fine", [mk(fine, "f")] * 3)
    check("K planted draws at the fine granularity (8 areas of 3 files) are recovered exactly", rc == 0 and partition_of(cat) == want(fine), err[-300:])
    rc, cat, err, _ = consensus(tmp, "mixed-fine", [mk(fine, "f"), mk(fine, "f"), mk(coarse, "c")])
    check("K two fine draws and one coarse draw: sibling fine clusters co-occur in 1/3 < tau, so the fine areas are recovered", rc == 0 and partition_of(cat) == want(fine), err[-300:])
    rc, cat, err, _ = consensus(tmp, "mixed-coarse", [mk(coarse, "c"), mk(coarse, "c"), mk(fine, "f")])
    check("K two coarse draws and one fine draw: sibling fine clusters co-occur in 2/3 >= tau, so the coarse areas are recovered", rc == 0 and partition_of(cat) == want(coarse), err[-300:])
    plant = {f: j for j, g in enumerate(coarse) for f in g}
    aris, pure = [], 0
    for seed in range(6):
        rng = random.Random("x2d-perm-%d" % seed)
        draws = []
        for _ in range(3):
            sh = files[:]
            rng.shuffle(sh)
            draws.append(mk([sh[i:i + 6] for i in range(0, 24, 6)], "p"))
        rc, cat, err, _ = consensus(tmp, "perm%d" % seed, draws)
        lab = {f: i for i, a in enumerate(cat["areas"]) for f in a["files"]} if rc == 0 else {}
        if rc == 0 and len(lab) == len(files):
            aris.append(ari([plant[f] for f in files], [lab[f] for f in files]))
            pure += sum(1 for a in cat["areas"] if len(a["files"]) >= 3 and len({plant[f] for f in a["files"]}) == 1)
        else:
            aris.append(None)
    check("K draws with permuted memberships (independent random partitions) recover nothing: the consensus' adjusted Rand index against the plant stays near 0 and no cluster of >= 3 files is a pure planted group",
          len(aris) == 6 and None not in aris and max(abs(a) for a in aris) < 0.2 and pure == 0, (aris, pure))
    spec = [{"a": ("Area A", files[:6])}, {"a": ("Area A", files[:6]), "x": ("Other", [files[6]])}, {"a": ("Area A", files[:5])}]
    rc, cat, err, out = consensus(tmp, "contested", spec)
    cid = {r["path"] for r in cat["contested"]} if rc == 0 else set()
    owned = {f for a in cat["areas"] for f in a["files"]} if rc == 0 else set()
    check("K ownership vote: a file in >= ceil(K/2) = 2 of 3 draws is owned (f00-f04 in 3, f05 in 2); a file in one draw is contested and not in any area; the title is selected from a draw's feature",
          rc == 0 and owned == set(files[:6]) and cid == {files[6]} and all(a["title"] in ("Area A",) and a["area_id"] for a in cat["areas"]), (sorted(owned), cid, err[-200:]))
    rid = hashlib.sha256(("x2d-area" + "core" + "\0".join(sorted(files[:6]))).encode()).hexdigest()[:12]
    check("K area ids are sha256('x2d-area' + scope + sorted consensus members)[:12]; sample_files (R1-17) of a single-directory area of 6 are its 6 files in sorted order (<= 8, one per cluster key); the catalogue carries heuristic_only",
          rc == 0 and cat["areas"][0]["area_id"] == rid and cat["areas"][0]["sample_files"] == files[:6] and cat["areas"][0]["files"] == files[:6] and "heuristic_only" in cat, (cat["areas"][0] if rc == 0 else err))
    rc2, cat2, _, out2 = consensus(tmp, "contested2", spec)
    check("K the consensus is deterministic (two runs on identical draws give a byte-identical catalogue.json)", rc == 0 and rc2 == 0 and open(os.path.join(out, "catalogue.json"), "rb").read() == open(os.path.join(out2, "catalogue.json"), "rb").read())
    r = run([_BIN["onboard"], "-draws", os.path.join(tmp, "cons-contested", "draws"), "-scopes", os.path.join(tmp, "cons-contested", "scopes.json"), "-out", out, "-consensus-only"], cwd=ROOT)
    check("K onboarding refuses to overwrite an existing catalogue (append-only)", r.returncode != 0 and "refus" in (r.stderr + r.stdout).lower(), r.stderr[-200:])
    json.dump({"rc": 1}, open(os.path.join(tmp, "cons-coarse", "draws", "core", "2", "meta.json"), "w"))
    r = run([_BIN["onboard"], "-draws", os.path.join(tmp, "cons-coarse", "draws"), "-scopes", os.path.join(tmp, "cons-coarse", "scopes.json"), "-out", os.path.join(tmp, "cons-coarse", "c2"), "-consensus-only"], cwd=ROOT)
    check("K a failed draw (meta.json rc != 0) is refused rather than consolidated with a smaller quorum", r.returncode != 0, r.stderr[-200:])
    div = ["docs_src/topic/tutorial001.py", "docs_src/topic/tutorial001_an.py", "docs_src/topic/tutorial001_an_py39.py", "docs_src/topic/tutorial001_py310.py",
           *["docs_src/topic/tutorial%03d.py" % i for i in range(2, 9)], "tests/test_tutorial/test_topic/test_tutorial001.py", "fastapi/topic.py", "fastapi/other.py"]
    rc, cat, err, _ = consensus(tmp, "sample-div", [{"d": ("Topic", div)}] * 3, include=["docs_src/", "tests/", "fastapi/"])
    sf = cat["areas"][0]["sample_files"] if rc == 0 else []
    keys = [F.cluster_key(p) for p in sf]
    check("K sample_files (R1-17): <= 8 files, round-robin over the top-level directories (docs_src, fastapi, tests first), ONE file per cluster key (the four tutorial001 variants give one), so the test file and the core file are shown, not variant clones",
          rc == 0 and len(sf) == 8 and len(set(keys)) == len(keys) and [p.split("/")[0] for p in sf[:3]] == ["docs_src", "fastapi", "tests"] and "tests/test_tutorial/test_topic/test_tutorial001.py" in sf
          and sum(1 for p in sf if F.variant_stem(p.rsplit("/", 1)[1].rsplit(".", 1)[0]) == "tutorial001") == 1 and set(sf) <= set(div), (sf, err[-200:]))
    clones = ["docs_src/x/tutorial001.py", "docs_src/x/tutorial001_an.py", "docs_src/x/tutorial001_an_py39.py", "docs_src/x/tutorial001_py310.py", "docs_src/x/tutorial001_py39.py", "docs_src/x/tutorial001_an_py310.py"]
    rc, cat, err, _ = consensus(tmp, "sample-clones", [{"c": ("Clones", clones)}] * 3, include=["docs_src/"])
    sf = cat["areas"][0]["sample_files"] if rc == 0 else []
    check("K sample_files top-up (R1-17): an area whose files are all variants of one example (one cluster key < min(8, 6)) is topped up with the remaining sorted files, so small or clone-heavy areas still reach n = 5", rc == 0 and sorted(sf) == sorted(clones) and len(sf) == 6, (sf, err[-200:]))
    a_, b_ = ["core/a%d.py" % i for i in range(3)], ["core/b%d.py" % i for i in range(3)]
    bridge = "core/zbridge.py"
    spec = [{"fa": ("Group A", a_ + [bridge]), "fb": ("Group B", b_ + [bridge])}, {"fa": ("Group A", a_ + [bridge]), "fb": ("Group B", b_ + [bridge])}, {"fall": ("All", a_ + b_ + [bridge])}]
    rc, cat, err, out = consensus(tmp, "bridge", spec)
    rep = json.load(open(os.path.join(out, "consensus-report.json"))) if rc == 0 else {}
    check("K consensus-report.json lists the bridge files (R1-27): the articulation points of the co-association graph at tau (the file that alone links two groups), and no other file",
          rc == 0 and [b["path"] for b in rep.get("bridge_files", [])] == [bridge] and len(cat["areas"]) == 1 and rep["bridge_files"][0]["area_id"] == cat["areas"][0]["area_id"] and rep["bridge_files"][0]["scope"] == "core", (rep.get("bridge_files"), err[-200:]))


# ---------------------------------------------------------------- J the Jev driver on a stub endpoint
import stub_jev  # noqa: E402
import r4_checks  # noqa: E402 (council round-4 fixtures; takes this module as `V`, imports nothing from it)
import regate_checks  # noqa: E402 (rev-13 re-gate fixtures; takes this module as `V`, imports nothing from it)


def jev_fixture(base, n_units=12):
    """-> (catalogue dict, units, truth dict). Area ids are 12-hex; unit i lives in `docs_src/t<i>/tutorial001.py`; truth: area A[i % 4], covered unless i % 3 == 0."""
    os.makedirs(base, exist_ok=True)
    areas = []
    for j in range(4):
        aid = hashlib.sha256(("area%d" % j).encode()).hexdigest()[:12]
        files = ["docs_src/a%d/f%d.py" % (j, k) for k in range(10)]
        areas.append({"area_id": aid, "scope": "core", "title": "Area %d responsibility" % j, "summary": "Handles topic number %d of the library." % j, "sample_files": files[:8], "files": files,
                      "clawpatch_ids": []})
    cat = {"schema": "x2d-catalogue-v1", "c0": x2d.C0, "scopes": ["core"], "tau": 0.5, "areas": sorted(areas, key=lambda a: a["area_id"]), "contested": [], "unowned": [], "jev_assigned": [],
           "heuristic_only": []}
    units, truth = [], {}
    for i in range(n_units):
        p = "docs_src/t%d/tutorial001.py" % i
        cid = hashlib.sha256(("unit%d" % i).encode()).hexdigest()[:16]
        a = cat["areas"][i % 4]["area_id"]
        title = "Change number %d" % i
        units.append({"case_id": "1:" + cid, "cluster_id": cid, "pr_key": str(100 + i // 3), "members": [p], "state": {
            "unit": {"paths": [p], "change_kind": "added", "role": "example", "excerpt": "def f():\n    return %d\n" % i},
            "pr": {"title": title, "body": "body %d" % i, "changed_paths": [{"path": p, "status": "added"}]}, "diff_summary": {"files": 1, "added_lines": 2, "deleted_lines": 0, "top_symbols": ["f"]}},
            "siblings": [{"path": "docs_src/a0/f0.py", "area_id": cat["areas"][0]["area_id"]}]})
        truth[p] = {"area": a, "covered": i % 3 != 0}
        truth["title:" + title] = {"area": a, "covered": i % 3 != 0}
    return cat, units, truth


def jev_run(tmp, name, units, cat, stub, extra=(), key=KEY, out=None, arms="sem,sib,title,fan", model=None):
    d = os.path.join(tmp, "jev-" + name)
    os.makedirs(d, exist_ok=True)
    up, cp = os.path.join(d, "units.jsonl"), os.path.join(d, "catalogue.json")
    if not os.path.exists(up):
        wjl(up, units)
        json.dump(cat, open(cp, "w"))
    out = out or os.path.join(d, "scores.jsonl")
    env = clean_env()
    if key is not None:
        env["OPENROUTER_API_KEY"] = key
    cmd = [_BIN["jev"], "-in", up, "-catalogue", cp, "-out", out, "-arms", arms, "-questions-dir", os.path.join(d, "questions"), "-endpoint", stub.url, "-workers", "4", "-model", model or x2d.JEV_MODEL, *extra]
    r = run(cmd, cwd=ROOT, env=env)
    return r, out, d


def jev_checks(tmp):
    if go_bin(tmp, "jev", "./evaluation/x2d-merge-association-pilot/jev") is None:
        check("J the Jev driver builds", False)
        return
    cat, units, truth = jev_fixture(os.path.join(tmp, "jevfix"))
    areas = [a["area_id"] for a in cat["areas"]]
    log = os.path.join(tmp, "jev-stub.log")
    stub = stub_jev.Stub(x2d.JEV_MODEL, truth=truth, p_correct=1.0, seed="j1", log_path=log).start()
    try:
        d0 = os.path.join(tmp, "jev-pv")
        os.makedirs(d0)
        wjl(os.path.join(d0, "u.jsonl"), units[:1])
        json.dump(cat, open(os.path.join(d0, "c.json"), "w"))
        pv = lambda m=x2d.JEV_MODEL, arms="sem,sib,title,fan": run([_BIN["jev"], "-print-versions", "-catalogue", os.path.join(d0, "c.json"), "-model", m, "-arms", arms], cwd=ROOT)
        a, b = pv(), pv()
        lines = [ln for ln in a.stdout.splitlines() if ln.strip()]
        keys = [re.findall(r"[0-9a-f]{64}", ln) for ln in lines]
        check("J -print-versions prints one version key per question and variant (merge.covered, merge.which_area x V-sem/V-sib/V-title and merge.area_covers), deterministic across runs, no key needed",
              a.returncode == 0 and a.stdout == b.stdout and len(lines) >= 7 and all(keys) and len({k[0] for k in keys}) == len(keys), a.stdout + a.stderr)
        c = pv("typesafe/jev-1.15-20270101")
        check("J the question version key changes with the resolved model (the pin is part of the key)", c.returncode == 0 and c.stdout != a.stdout)
        r = pv("typesafe/jev-1.14")
        check("J an alias or undated release line is refused as the model", r.returncode == 2 and "alias" in r.stderr.lower(), r.stderr[-200:])
        r, out, d = jev_run(tmp, "nokey", units, cat, stub, key=None)
        check("J without OPENROUTER_API_KEY the driver exits 2, sends nothing and writes no output", r.returncode == 2 and stub.requests == 0 and not os.path.exists(out), (r.returncode, r.stderr[-200:]))
        before = stub.requests
        r, out, d = jev_run(tmp, "full", units, cat, stub, extra=["-max-cost-usd", "5"])
        rows = jl(out)
        hdr, ans, fail = rows[0], [x for x in rows if x["type"] == "answer"], [x for x in rows if x["type"] == "failure"]
        n = len(units)
        flipped = [u for u in units if x2d.flip_selected(u["cluster_id"])]
        want_rows = n * 3 + n * 3 + len(flipped) * 3 + n * len(areas)
        check("J a full run answers every request: per unit merge.covered x 3 variants, merge.which_area x 3 variants (+ the second permutation for exactly the sha256('x2d-flip') mod 5 == 0 clusters) and one merge.area_covers per area; no failures, exit 0",
              r.returncode == 0 and not fail and len(ans) == want_rows and len({x["req_id"] for x in ans}) == len(ans) and stub.requests - before == want_rows, (r.returncode, r.stderr[-300:], len(ans), want_rows, len(flipped)))
        check("J the header records model, catalogue digest, input digest, arms and one question version key per question/variant, and the catalogue digest is the file's SHA-256",
              hdr["type"] == "header" and hdr["model"] == x2d.JEV_MODEL and hdr["catalogue_sha256"] == x2d.sha256_file(os.path.join(d, "catalogue.json"))
              and hdr["input_sha256"] == x2d.sha256_file(os.path.join(d, "units.jsonl")) and len(hdr["versions"]) >= 7 and all(re.fullmatch(r"merge\.[a-z_]+@[0-9a-f]{64}", v) for v in hdr["versions"].values()), hdr)
        qv = {(x["question"], x["variant"]): x["question_version"] for x in ans if x["question"] != "merge.area_covers"}
        check("J every answer row carries the question version key of its variant (distinct per variant) and the questions as sent are written to <dir>/<id>@<key>.json",
              len(set(qv.values())) == 6 and all(os.path.exists(os.path.join(d, "questions", v + ".json")) for v in qv.values()), qv)
        cov = {(x["case_id"], x["variant"]): x["p_true"] for x in ans if x["question"] == "merge.covered"}
        pick = {(x["case_id"], x["variant"], x["perm"]): x["choice"] for x in ans if x["question"] == "merge.which_area"}
        ok_cov = all((cov[(u["case_id"], "V-sem")] > 0.5) == truth[u["members"][0]]["covered"] and (cov[(u["case_id"], "V-sib")] > 0.5) == truth[u["members"][0]]["covered"] for u in units)
        ok_ch = all(pick[(u["case_id"], "V-sem", 1)] == truth[u["members"][0]]["area"] for u in units)
        ok_title = all((cov[(u["case_id"], "V-title")] > 0.5) == truth[u["members"][0]]["covered"] for u in units)
        check("J rows hold the planted answers (stub p_correct = 1): Noul p_true and Choice top option for V-sem, V-sib and V-title (the title-only variant reaches the truth through its own key)", ok_cov and ok_ch and ok_title)
        perm2 = {(x["case_id"], x["variant"]) for x in ans if x["perm"] == 2 and x["question"] == "merge.which_area"}
        check("J the second permutation is asked for exactly the flip-selected clusters under every variant, with unit id <case_id>#2 (a different realised option order)",
              perm2 == {(u["case_id"], v) for u in flipped for v in ("V-sem", "V-sib", "V-title")} and len(flipped) >= 1)
        reqs = [json.loads(ln) for ln in open(log)]
        ch = [q for q in reqs if q["type"] == "choice"]
        noul = [q for q in reqs if q["type"] == "noul"]
        check("J request shape (stub-checked): model equals the pin, one question under key q1, Choice options are exactly the catalogue's area ids, Noul options true/false, an Authorization Bearer header and no key in any body, no protocol violation",
              stub.violations == [] and all(q["model"] == x2d.JEV_MODEL and q["question_keys"] == ["q1"] and q["auth_present"] and not q["body_has_key"] for q in reqs)
              and all(sorted(q["criteria_keys"]) == sorted(areas) for q in ch) and all(sorted(q["criteria_keys"]) == ["false", "true"] for q in noul), stub.violations[:3])
        orders = {tuple(q["criteria_keys"]) for q in ch}
        check("J the Choice option order is a permutation per unit (many distinct orders across units, never the catalogue order for every unit) while each unit's order is the same in every variant",
              len(orders) >= 3 and orders != {tuple(areas)}, len(orders))
        sem_titles = [q for q in ch if q["state"].get("pr", {}).get("title") and "unit" in q["state"]]
        title_only = [q for q in ch if "unit" not in q["state"]]
        check("J V-title sends only the pull-request title (no unit, body, changed paths or diff summary); V-sem sends no owner facts (no siblings); V-sib adds state.siblings",
              title_only and all(set(q["state"]) == {"pr"} and set(q["state"]["pr"]) == {"title"} for q in title_only)
              and any("siblings" in q["state"] for q in sem_titles) and any("siblings" not in q["state"] and "unit" in q["state"] for q in sem_titles), (len(title_only), len(sem_titles)))
        nl = [q for q in noul if "unit" in q["state"] and "area" not in q["state"]]
        check("J the Noul carries the catalogue in its state; the Choice carries it in its options only; the fan-out Noul carries exactly one area (title, summary, sample files)",
              all("catalogue" in q["state"] for q in nl) and all("catalogue" not in q["state"] for q in ch)
              and all(set(q["state"]["area"]) >= {"id", "title", "summary", "sample_files"} for q in noul if "area" in q["state"]), len(nl))
        crit_txt = next(iter(ch))["criteria"]
        check("J option criteria render the preregistered template '{title}: {summary} Sample files: ...' with exactly the default 5 of the catalogue's sample_files (`sample_files[:n]`, 8 stored) and opaque ids as options",
              all(re.match(r"^Area \d responsibility: Handles topic number \d of the library\. Sample files: ", t) and t.count(".py") == 5 for t in crit_txt.values()), crit_txt)
        # resume and append-only
        before_n, before_bytes = stub.requests, open(out, "rb").read()
        r2, _o, _d = jev_run(tmp, "full", units, cat, stub, extra=["-max-cost-usd", "5"])
        check("J re-running with the same input resumes: no request is sent and the output file is byte-identical (append-only, resumable)", r2.returncode == 0 and stub.requests == before_n and open(out, "rb").read() == before_bytes, r2.stderr[-200:])
        r3, _o, _d = jev_run(tmp, "full", units, cat, stub, arms="sem")
        check("J resuming against an output whose header names other arms (or input, catalogue, model, question versions) is refused with exit 2", r3.returncode == 2 and stub.requests == before_n, r3.stderr[-200:])
        cat2 = json.loads(json.dumps(cat))
        cat2["areas"][0]["title"] = "Changed title"
        d_ = os.path.join(tmp, "jev-full")
        json.dump(cat2, open(os.path.join(d_, "catalogue2.json"), "w"))
        r4 = run([_BIN["jev"], "-in", os.path.join(d_, "units.jsonl"), "-catalogue", os.path.join(d_, "catalogue2.json"), "-out", out, "-arms", "sem,sib,title,fan", "-endpoint", stub.url], cwd=ROOT, env=dict(clean_env(), OPENROUTER_API_KEY=KEY))
        check("J a changed catalogue (title text) refuses to resume the same output (the catalogue digest is in the header)", r4.returncode == 2 and stub.requests == before_n, r4.stderr[-200:])
        every = open(out).read() + r.stdout + r.stderr
        check("J key hygiene: the key value appears in no output file, stdout or stderr of the driver", KEY not in every and "sk-or-v1" not in every)
        rows_costs = [x for x in ans if x["cost_known"]]
        check("J every answer row records the billed cost and the resolved model", len(rows_costs) == len(ans) and all(x["model"] == x2d.JEV_MODEL for x in ans) and x2d.spent_in(out) > 0)
    finally:
        stub.stop()
    # faults
    fu = units[:4]
    p0 = fu[0]["members"][0]
    stub = stub_jev.Stub(x2d.JEV_MODEL, truth=truth, p_correct=1.0, seed="j2", fail_ids={p0}).start()
    try:
        r, out, d = jev_run(tmp, "fail", fu, cat, stub, arms="sem")
        rows = jl(out)
        fails = [x for x in rows if x["type"] == "failure"]
        check("J a failed request (HTTP 500) is a failure row by kind with the unit, question and variant, exit 3, and the other units still answer", r.returncode == 3 and fails and all(x["kind"] == "http" and x["case_id"] == fu[0]["case_id"] for x in fails)
              and any(x["type"] == "answer" and x["case_id"] == fu[1]["case_id"] for x in rows), (r.returncode, fails[:1]))
        stub.fail_ids = set()
        r, out, d = jev_run(tmp, "fail", fu, cat, stub, arms="sem")
        rows = jl(out)
        done = {x["req_id"] for x in rows if x["type"] == "answer"}
        check("J a re-run retries the failed requests (and only those), appends their answers and exits 0", r.returncode == 0 and all(u["case_id"] in " ".join(done) for u in fu) and sum(1 for x in rows if x["type"] == "failure") == len(fails), r.stderr[-200:])
    finally:
        stub.stop()
    for nm, kw, kind in (("model", {"model_for": {p0: "other/model-9"}}, "model-mismatch"), ("drop", {"drop_option": "$first"}, "schema-invalid"), ("noul", {"drop_option": "noul"}, "schema-invalid"),
                         ("scale", {"scale": 0.5}, "schema-invalid")):
        stub = stub_jev.Stub(x2d.JEV_MODEL, truth=truth, p_correct=1.0, seed="j3", **kw).start()
        try:
            r, out, d = jev_run(tmp, "f-" + nm, fu[:2], cat, stub, arms="sem")
            fails = [x for x in jl(out) if x["type"] == "failure"]
            label = {"model": "a reply from another model is a model-mismatch failure, never an answer", "drop": "a Choice reply that omits an option is a schema failure",
                     "noul": "a Noul reply without its probability is a schema failure", "scale": "an unnormalised Choice reply (probabilities do not sum to 1) is a schema failure"}[nm]
            check("J " + label, fails and any(x["kind"] == kind or (kind == "schema-invalid" and x["kind"] == "schema") for x in fails) and r.returncode == 3, (r.returncode, [x["kind"] for x in fails][:4], r.stderr[-200:]))
        finally:
            stub.stop()
    stub = stub_jev.Stub(x2d.JEV_MODEL, truth=truth, p_correct=1.0, seed="j4", rate_limit_first=2).start()
    try:
        r, out, d = jev_run(tmp, "429", fu[:1], cat, stub, arms="sem")
        check("J a 429 reply is retried by the product client (Retry-After honoured) and the request is answered", r.returncode == 0 and not [x for x in jl(out) if x["type"] == "failure"] and stub.requests > len([x for x in jl(out) if x["type"] == "answer"]), r.stderr[-200:])
    finally:
        stub.stop()
    stub = stub_jev.Stub(x2d.JEV_MODEL, truth=truth, p_correct=1.0, seed="j5", no_cost=True).start()
    try:
        r, out, d = jev_run(tmp, "nocost", fu[:1], cat, stub, arms="sem")
        check("J a reply without a billed cost is recorded as cost_known = false and the run exits 3 (an unknown cost is never silently zero)", r.returncode == 3 and all(not x["cost_known"] for x in jl(out) if x["type"] == "answer"), r.returncode)
    finally:
        stub.stop()
    stub = stub_jev.Stub(x2d.JEV_MODEL, truth=truth, p_correct=1.0, seed="j6").start()
    stub2 = stub_jev.Stub(x2d.JEV_MODEL, truth=truth, p_correct=1.0, seed="j6b").start()  # the re-run's endpoint: a request the first run cancelled in flight can only reach `stub`
    try:
        r, out, d = jev_run(tmp, "cap", units, cat, stub, arms="sem,sib,title,fan", extra=["-max-cost-usd", "0.00005"])
        spent1 = x2d.spent_in(out)
        n_after = len([x for x in jl(out) if x["type"] == "answer"])
        r2, _o, _d = jev_run(tmp, "cap", units, cat, stub2, arms="sem,sib,title,fan", extra=["-max-cost-usd", "0.00005"])
        total = len(units) * (6 + len(areas)) + sum(3 for u in units if x2d.flip_selected(u["cluster_id"]))
        check("J the cumulative cost stop: a run whose billed cost exceeds -max-cost-usd stops early with exit 4 (fewer answers than requests) and a re-run refuses to start (exit 2, no request) once the file's spend exceeds the cap",
              r.returncode == 4 and 0 < n_after < total and spent1 > 0.00005 and r2.returncode == 2 and stub2.requests == 0 and "budget" in r2.stderr.lower(),
              (r.returncode, r2.returncode, n_after, total, spent1, stub2.requests, r2.stderr[-200:]))
    finally:
        stub.stop()
        stub2.stop()
    check("J x2d.jev_budget subtracts the other score files of the round and of the study from the per-round and total caps (a file's own spend is the driver's)",
          jev_budget_test(tmp))
    check("J R3-17 the v_7 re-score is charged to the x2d.CAPS total only: a round at its per-round cap still re-scores (budget = the total left, CAPS total - cap - 0.8), a run of the round itself gets 0, and the re-score spend does not reduce another run's per-round allowance",
          jev_budget_v7_test(tmp))
    dump_state_checks(tmp, cat, units, truth)


def dump_state_checks(tmp, cat, units, truth):
    """R1-16, R1-17, R1-28: `jev -dump-state` writes exactly the states the driver sends, at sample counts 3, 5 and 8; the driver refuses n > the longest sample_files; an N1-style
    catalogue (no sample files) shows none; include_pr_title: false drops state.pr.title from V-sem."""
    d = os.path.join(tmp, "jev-dump")
    os.makedirs(d, exist_ok=True)
    us = units[:6]
    up, cp = os.path.join(d, "units.jsonl"), os.path.join(d, "catalogue.json")
    wjl(up, us)
    json.dump(cat, open(cp, "w"))

    def sent_and_dumped(tag, extra, cpath=cp):
        sp, out, log = (os.path.join(d, tag + ext) for ext in (".states.jsonl", ".scores.jsonl", ".stub.log"))
        r0 = run([_BIN["jev"], "-dump-state", "-in", up, "-catalogue", cpath, "-out", sp, "-arms", "sem", *extra], cwd=ROOT, env=clean_env())  # no key in the environment
        stub = stub_jev.Stub(x2d.JEV_MODEL, truth=truth, p_correct=1.0, seed="dump-" + tag, log_path=log).start()
        try:
            r1 = run([_BIN["jev"], "-in", up, "-catalogue", cpath, "-out", out, "-arms", "sem", "-endpoint", stub.url, "-workers", "2", "-model", x2d.JEV_MODEL, "-flip-all", *extra],
                     cwd=ROOT, env=dict(clean_env(), OPENROUTER_API_KEY=KEY))
        finally:
            stub.stop()
        reqs = [json.loads(ln) for ln in open(log)] if os.path.exists(log) else []
        return r0, r1, (jl(sp) if r0.returncode == 0 else []), reqs

    def same(rows, reqs, units_):
        bad = []
        for u in units_:
            p = u["members"][0]
            row = next((x for x in rows if x["case_id"] == u["case_id"] and x["variant"] == "V-sem"), None)
            noul = [q for q in reqs if q["type"] == "noul" and "catalogue" in q["state"] and q["state"].get("unit", {}).get("paths", [None])[0] == p]
            ch = [q for q in reqs if q["type"] == "choice" and q["state"].get("unit", {}).get("paths", [None])[0] == p]
            if row is None or not noul or not ch:
                bad.append((u["case_id"], "missing"))
                continue
            if not all(q["state"] == row["noul_state"] for q in noul) or not all(q["state"] == row["choice_state"] and q["criteria"] == row["choice_criteria"] for q in ch):
                bad.append((u["case_id"], "differs"))
            if row["choice_options"] != [row["choice_criteria"][a["area_id"]] for a in cat["areas"]]:
                bad.append((u["case_id"], "options"))
        return bad

    for n in (3, 5, 8):
        r0, r1, rows, reqs = sent_and_dumped("n%d" % n, ["-sample-files", str(n)])
        shown = [len(a["sample_files"]) for x in rows for a in x["noul_state"]["catalogue"]]
        check("J -dump-state at -sample-files %d (no key, no request): each unit's dumped Noul state, Choice state and option texts equal the states the driver sends, %d sample files per area (R1-16)" % (n, n),
              r0.returncode == 0 and r1.returncode == 0 and len(rows) == len(us) and not same(rows, reqs, us) and shown and set(shown) == {n}, (r0.stderr[-200:], r1.stderr[-200:], same(rows, reqs, us)[:3], set(shown)))
    r = run([_BIN["jev"], "-dump-state", "-in", up, "-catalogue", cp, "-out", os.path.join(d, "n9.jsonl"), "-arms", "sem", "-sample-files", "9"], cwd=ROOT, env=clean_env())
    check("J the driver refuses n > the longest sample_files of the catalogue (-sample-files 9 on a catalogue storing 8): exit 2 (R1-17)", r.returncode == 2 and "sample_files" in r.stderr, (r.returncode, r.stderr[-200:]))
    cat_n1 = json.loads(json.dumps(cat))
    for a in cat_n1["areas"]:
        a["sample_files"] = []
    cn = os.path.join(d, "catalogue-n1.json")
    json.dump(cat_n1, open(cn, "w"))
    r0, r1, rows, reqs = sent_and_dumped("n1", ["-sample-files", "5"], cpath=cn)
    check("J a catalogue whose areas all have empty sample_files (the N1 control) is accepted and shows no file names: no fallback to `files` (R1-17)",
          r0.returncode == 0 and rows and all(a["sample_files"] == [] for x in rows for a in x["noul_state"]["catalogue"]) and not any("docs_src/a0/f" in json.dumps(x) for x in rows), (r0.stderr[-200:], len(rows)))
    st = os.path.join(d, "settings-notitle.json")
    json.dump({"sample_files_per_area": 3, "include_pr_title": False}, open(st, "w"))
    r0, r1, rows, reqs = sent_and_dumped("notitle", ["-settings", st])
    check("J include_pr_title: false removes state.pr.title from V-sem; the dump still equals what the driver sends, at the settings' 3 sample files (R1-28)",
          r0.returncode == 0 and r1.returncode == 0 and rows and all("title" not in x["noul_state"].get("pr", {}) and "title" not in x["choice_state"].get("pr", {}) for x in rows) and not same(rows, reqs, us)
          and {len(a["sample_files"]) for x in rows for a in x["noul_state"]["catalogue"]} == {3}, (r0.stderr[-200:], r1.stderr[-200:]))
    # R2-17: a non-empty V-sem instructions_delta (a permitted wording refinement) reaches the dumped state's effective instructions and, from there, the labeller and reference prompts
    qsrc = os.path.join(d, "questions-delta")
    shutil.copytree(os.path.join(S, "questions"), qsrc)
    for qname, tag in (("merge.covered", " DELTA-COVERED-8f3a"), ("merge.which_area", " DELTA-CHOICE-5b1c")):
        qp = os.path.join(qsrc, qname + "@v1.json")
        qj = json.load(open(qp))
        qj["variants"]["V-sem"]["instructions_delta"] = tag
        json.dump(qj, open(qp, "w"), indent=1)
    sp_d = os.path.join(d, "delta.states.jsonl")
    r_d = run([_BIN["jev"], "-dump-state", "-in", up, "-catalogue", cp, "-out", sp_d, "-arms", "sem", "-questions-src", qsrc], cwd=ROOT, env=clean_env())
    rows_d = jl(sp_d) if r_d.returncode == 0 else []
    rows_0 = jl(os.path.join(d, "n5.states.jsonl"))
    import reference_arm as RA
    ok_d = bool(rows_d) and all("DELTA-COVERED-8f3a" in x["noul_instructions"] and "DELTA-CHOICE-5b1c" in x["choice_instructions"] for x in rows_d) and not any("DELTA-" in x["noul_instructions"] + x["choice_instructions"] for x in rows_0)
    lab_p = LX.intro({"kind": "jev-state", "instructions": (rows_d[0]["noul_instructions"], rows_d[0].get("noul_criteria"))}) if rows_d else ""
    lab_c = LX.intro({"kind": "jev-state", "instructions": (rows_d[0]["choice_instructions"], None)}) if rows_d else ""
    ref_p = RA.prompt(rows_d[0]) if rows_d else ""
    check("J R2-17 a non-empty V-sem instructions_delta (temp questions dir passed to `-dump-state -questions-src`) reaches the dumped noul_instructions / choice_instructions (and is absent without it), the labeller's jev-state prompt and the reference arm prompt",
          r_d.returncode == 0 and ok_d and "DELTA-COVERED-8f3a" in lab_p and "DELTA-CHOICE-5b1c" in lab_c and "DELTA-COVERED-8f3a" in ref_p and "DELTA-CHOICE-5b1c" in ref_p, (r_d.stderr[-300:], len(rows_d), ref_p[:200]))


def jev_budget_test(tmp):
    data = os.path.join(tmp, "budget-data")
    row = lambda c: {"type": "answer", "cost_known": True, "cost_usd": c}
    wjl(os.path.join(data, "rounds", "3", "jev", "scores.jsonl"), [row(0.30), row(0.10)])
    wjl(os.path.join(data, "rounds", "3", "jev", "scores.contrast.jsonl"), [row(0.20)])
    wjl(os.path.join(data, "rounds", "4", "jev", "scores.jsonl"), [row(x2d.CAPS["usd_total"] - 1.0)])  # total - 1.0 (9.0 under the rev-11 caps)
    wjl(os.path.join(data, "onboard-jev", "scores.jsonl"), [row(0.05), {"type": "failure", "cost_known": False, "cost_usd": 5}])
    own = os.path.join(data, "rounds", "3", "jev", "scores.jsonl")
    b_round = x2d.jev_budget(3, own, data)       # per-round: CAPS per round - 0.20 ; total: CAPS total - (0.20 + (CAPS total - 1.0) + 0.05) = 0.75 -> 0.75 (the total binds)
    b_other = x2d.jev_budget(5, os.path.join(data, "rounds", "5", "jev", "scores.jsonl"), data)  # total: CAPS total - (CAPS total - 1.0) - 0.60 - 0.05 = 0.35 ; round cap CAPS per round
    b_onb = x2d.jev_budget(None, os.path.join(data, "onboard-jev", "scores.jsonl"), data)         # total: CAPS total - (CAPS total - 1.0) - 0.60 = 0.40
    pr = os.path.join(tmp, "budget-data-pr")
    wjl(os.path.join(pr, "rounds", "3", "jev", "scores.jsonl"), [row(x2d.CAPS["usd_per_round"] - 0.5)])
    wjl(os.path.join(pr, "rounds", "3", "jev", "scores.contrast.jsonl"), [row(0.20)])
    b_pr = x2d.jev_budget(3, os.path.join(pr, "rounds", "3", "jev", "scores.jsonl"), pr)  # per-round binds: CAPS per round - 0.20; with the own file counted it would be 0.30
    return (abs(b_round - 0.75) < 1e-9 and abs(b_other - 0.35) < 1e-9 and abs(b_onb - 0.40) < 1e-9
            and abs(b_pr - (x2d.CAPS["usd_per_round"] - 0.20)) < 1e-9 and b_pr < x2d.CAPS["usd_total"] - 0.20)


def jev_budget_v7_test(tmp):
    """R3-17: files under rounds/<rr>/jev-v7/ are charged to the x2d.CAPS total only: a round whose own scoring is at its per-round cap still re-scores, and the re-score spend does not count against the round's own runs."""
    data = os.path.join(tmp, "budget-v7-data")
    row = lambda c: {"type": "answer", "cost_known": True, "cost_usd": c}
    wjl(os.path.join(data, "rounds", "3", "jev", "scores.jsonl"), [row(x2d.CAPS["usd_per_round"])])  # the round's own scoring: at the per-round cap
    wjl(os.path.join(data, "rounds", "4", "jev", "scores.jsonl"), [row(0.3)])
    wjl(os.path.join(data, "rounds", "4", x2d.V7_RESCORE_SUB, "scores.jsonl"), [row(0.5)])
    v7_at_cap = x2d.jev_budget(3, os.path.join(data, "rounds", "3", x2d.V7_RESCORE_SUB, "scores.jsonl"), data)   # total only: CAPS total - cap - 0.3 - 0.5 (not 0)
    own_at_cap = x2d.jev_budget(3, os.path.join(data, "rounds", "3", "jev", "scores.contrast.jsonl"), data)    # a new run of round 3 itself: round_other = the per-round cap -> 0
    other_run = x2d.jev_budget(4, os.path.join(data, "rounds", "4", "jev", "scores.contrast.jsonl"), data)     # round 4: round_other 0.3 (the v_7 spend 0.5 is excluded) -> cap - 0.3; the total left is far larger
    caps = x2d.CAPS
    return (abs(v7_at_cap - (caps["usd_total"] - caps["usd_per_round"] - 0.3 - 0.5)) < 1e-9 and own_at_cap == 0.0
            and abs(other_run - (caps["usd_per_round"] - 0.3)) < 1e-9 and other_run < caps["usd_total"] - caps["usd_per_round"] - 0.3 - 0.5)


# ---------------------------------------------------------------- L labeller harness and reference arm on stub sessions
import stub_labeller  # noqa: E402
import label_x2d as LX  # noqa: E402

POISON = ["ARMANSWER-V-sem-xyz", "SIBLING-OWNER-QQQ", "STRATUM-LEAK-777", "CONTROL-KIND-p1-leak", "ROUND-LEAK-424242", "SCORE-LEAK-0.987654"]


def lab_data(tmp, name, n=25, window=True):
    """A throw-away data directory with a catalogue and round-1 units that carry poison fields no labeller may ever see. -> (data dir, units, gold, area ids)."""
    data = os.path.join(tmp, "lab-" + name)
    cat, units, truth = jev_fixture(os.path.join(data, "fx"), n)
    os.makedirs(os.path.join(data, "catalogue"), exist_ok=True)
    json.dump(cat, open(os.path.join(data, "catalogue", "catalogue.json"), "w"))
    shutil.rmtree(os.path.join(data, "fx"))
    gold = {}
    for i, u in enumerate(units):
        t = truth[u["members"][0]]
        gold[u["case_id"]] = {"case_id": u["case_id"], "covered": t["covered"], "acceptable": [t["area"]] if t["covered"] else [], "primary": t["area"] if t["covered"] else None,
                              "none_kind": None if t["covered"] else "gap"}
        p = u["members"][0]
        u["label_evidence"] = {"unit": {"paths": [p], "change_kind": "added", "role": "example", "members": [{"path": p, "text": "def f():\n    return %d\n" % i}]},
                               "pr": u["state"]["pr"], "diff_summary": u["state"]["diff_summary"], "poison": POISON[0]}
        u.update({"set": "window", "control": False, "round": 1, "stratum": 2, "stratum_reason": POISON[2], "arm_output": POISON[0], "control_kind": POISON[3], "score": POISON[5],
                  "siblings": [{"path": "docs_src/a0/f0.py", "area_id": POISON[1]}], "pr_title": u["state"]["pr"]["title"]})
    wjl(os.path.join(data, "rounds", "1", "units.jsonl"), units)
    VR.ensure_inputs(1, data)
    return data, units, gold, [a["area_id"] for a in cat["areas"]]


def lab_env(data, truth_path=None, prompt_log=None, **extra):
    env = clean_env(X2D_DATA=data, **extra)
    if truth_path:
        env["X2D_LABEL_STUB"] = truth_path
    if prompt_log:
        env["X2D_LABEL_PROMPT_LOG"] = prompt_log
    return env


def label_checks(tmp):
    data, units, gold, ids = lab_data(tmp, "a")
    truth = stub_labeller.make_truth(units, gold, agree_rate=0.6, seed=7, area_ids=ids)
    # jev-state keys
    for u in units:
        g = gold[u["case_id"]]
        q1 = {"covered": g["covered"]}
        q2 = {"primary": g["primary"] or ids[0]}
        for q, a in (("merge.covered", q1), ("merge.which_area", q2)):
            truth[u["case_id"] + "|" + q] = {"A": dict(a), "B": dict(a), "C": dict(a)}
    exec_cid = next(c for c in sorted(truth) if "|" not in c and truth[c]["A"] and truth[c]["B"] and truth[c]["A"] == truth[c]["B"])
    stub_labeller.plant_exec(truth, exec_cid, 2)  # B (the codex session) ran two commands on an otherwise valid answer (R1-15)
    tp = os.path.join(tmp, "lab-truth.json")
    json.dump(truth, open(tp, "w"))
    log = os.path.join(tmp, "lab-prompts.jsonl")
    r = run([PY, os.path.join(S, "label_x2d.py"), "--round", "1", "--workers", "3"], cwd=ROOT, env=lab_env(data, tp, log))
    rd = os.path.join(data, "rounds", "1")
    lab = jl(os.path.join(rd, "labels.jsonl"))
    final = [x for x in lab if x["provenance"] == "agent_reviewed"]
    calls = jl(os.path.join(rd, "calls.jsonl"))
    kinds = {k: sum(1 for x in final if x["agreement"] == k) for k in ("agreed", "adjudicated", "unparsable")}
    check("L wide labelling on stub sessions: one final agent_reviewed row per unit (agreed / adjudicated / exactly one unparsable), A, B and (when they disagree) C session rows, one call row per unit",
          r.returncode == 0 and len(final) == len(units) and kinds["unparsable"] == 1 and kinds["adjudicated"] >= 1 and kinds["agreed"] >= 10 and len(calls) == len(units)
          and sum(1 for x in lab if x["by"].startswith("labeller_A")) == len(units) and sum(1 for x in lab if x["by"].startswith("adjudicator_C")) == kinds["adjudicated"] + kinds["unparsable"], (r.returncode, r.stderr[-300:], kinds, len(calls)))
    cx, fx = next(c for c in calls if c["case_id"] == exec_cid), next(x for x in final if x["case_id"] == exec_cid)
    check("L a codex session that ran a command (exec events) is unparsable: quarantined in the call row (attempt, exec count, raw text), not retried, never parsed into a label; the case goes to the adjudicator (R1-15)",
          cx["B"]["ok"] is False and cx["B"]["attempts"] == 1 and len(cx["B"]["quarantined"]) == 1 and cx["B"]["quarantined"][0]["exec_events"] == 2 and cx["C"] and fx["agreement"] == "adjudicated"
          and fx["covered"] == gold[exec_cid]["covered"] and any(x["case_id"] == exec_cid and x["by"].startswith("labeller_B") and x.get("parsed") is False for x in lab), (cx["B"], fx))
    ss = json.load(open(os.path.join(rd, "sessions.json")))
    used_calls = sum((c[k] or {}).get("attempts", 0) for c in calls for k in ("A", "B", "C") if c.get(k))
    check("L sessions.json records the computed cap, the used count and the plan: cap = ceil(1.25 (2.2 W (1 + C) + 4.4 J + R + 2.2 I)) for W = 25 units, J = 20, R = 20, C = 0 (204), used = every attempt of the call rows (R1-05, R3-01)",
          ss["cap"] == LX.cap_for(1, **ss["plan"]) == 204 and ss["plan"] == {"W": 25, "J": 20, "R": 20, "I": 0, "C": 0} and ss["used"] == used_calls, (ss, used_calls))
    prompts = [json.loads(ln) for ln in open(log)]
    leaks = stub_labeller.probe_leaks(log, POISON)
    check("L blindness: no prompt of any session (A, B or the adjudicator) contains an arm answer, a sibling's owner, the stratum, a control kind, the round or a score (evidence is built from an explicit allow-list)", prompts and not leaks, leaks[:3])
    by = {}
    for p in prompts:
        by.setdefault(p["cid"], {}).setdefault(p["who"], p["prompt"])
    same = all(v["A"] == v["B"] for v in by.values() if "A" in v and "B" in v)
    check("L both labellers' prompts are byte-identical in evidence (the same prompt for A and B on every unit)", same and len(by) == len(units))
    check("L the prompt carries the frozen guide, the unit's member text and the catalogue's area ids and nothing about Jev; a final row's gold answer for an agreed unit equals the planted one",
          all("x2d-association-v1" in p["prompt"] or "Label guide" in p["prompt"] or "R10" in p["prompt"] for p in prompts[:5]) and "typesafe" not in prompts[0]["prompt"]
          and all(x["covered"] == gold[x["case_id"]]["covered"] for x in final if x["agreement"] == "agreed"))
    before = open(os.path.join(rd, "labels.jsonl"), "rb").read()
    r2 = run([PY, os.path.join(S, "label_x2d.py"), "--round", "1"], cwd=ROOT, env=lab_env(data, tp))
    check("L labelling refuses to run again for the same guide version (append-only): labels.jsonl is untouched", r2.returncode != 0 and "refus" in (r2.stderr + r2.stdout).lower() and open(os.path.join(rd, "labels.jsonl"), "rb").read() == before, r2.stderr[-200:])
    bad_cid = next(x["case_id"] for x in final if x["agreement"] == "unparsable")
    t2 = json.loads(json.dumps(truth))
    t2[bad_cid] = {"A": _ans(gold[bad_cid]), "B": _ans(gold[bad_cid]), "C": _ans(gold[bad_cid])}
    tp2 = os.path.join(tmp, "lab-truth2.json")
    json.dump(t2, open(tp2, "w"))
    r3 = run([PY, os.path.join(S, "label_x2d.py"), "--round", "1", "--new-round", "r2"], cwd=ROOT, env=lab_env(data, tp2))
    after = open(os.path.join(rd, "labels.jsonl"), "rb").read()
    new_rows = jl(os.path.join(rd, "labels.jsonl"))[len(lab):]
    check("L --new-round labels only the units without a usable final label, appends (the earlier bytes are untouched) and marks its rows with the round name",
          r3.returncode == 0 and after.startswith(before) and new_rows and {x["case_id"] for x in new_rows} == {bad_cid} and all(x["round"] == "r2" for x in new_rows)
          and new_rows[-1]["provenance"] == "agent_reviewed" and new_rows[-1]["agreement"] == "agreed", (r3.returncode, r3.stderr[-200:], len(new_rows)))
    lock = os.path.join(rd, "labels.lock")
    open(lock, "w").close()
    r4 = run([PY, os.path.join(S, "label_x2d.py"), "--round", "1", "--new-round", "r3"], cwd=ROOT, env=lab_env(data, tp2))
    os.unlink(lock)
    check("L a lock file (another labelling run, or a killed one) refuses a concurrent run", r4.returncode != 0 and "labels.lock" in (r4.stderr + r4.stdout), r4.stderr[-200:])
    # the agreement rule
    a = {"covered": True, "acceptable": ["x", "y"], "primary": "x", "none_kind": None, "rule": "R1", "reason": "r"}
    check("L agree(): equal covered with >= 1 shared acceptable id is agreed, acceptable = the intersection, primary = A's when in it, else B's (never None while the sets share an id, R1-11)",
          LX.agree(a, dict(a, acceptable=["x"]))[0] == "agreed" and LX.agree(a, dict(a, acceptable=["x"]))[1]["acceptable"] == ["x"] and LX.agree(a, dict(a, acceptable=["x"]))[1]["primary"] == "x"
          and LX.agree(a, dict(a, acceptable=["y", "z"], primary="y"))[1]["acceptable"] == ["y"] and LX.agree(a, dict(a, acceptable=["y", "z"], primary="y"))[1]["primary"] == "y")
    pr = lambda x, y, sizes=None: LX.agree(dict(a, acceptable=x[0], primary=x[1]), dict(a, acceptable=y[0], primary=y[1]), sizes)[1]
    p_both, p_a, p_b = pr((["x", "y"], "x"), (["x", "y"], "x")), pr((["x", "y"], "x"), (["x", "y"], "y")), pr((["x", "y"], "x"), (["y", "z"], "y"))
    p_few, p_tie = pr((["x", "y", "w"], "w"), (["x", "y", "z"], "z"), {"x": 5, "y": 3}), pr((["x", "y", "w"], "w"), (["x", "y", "z"], "z"), {"x": 4, "y": 4})
    check("L agree() primary rule (R1-11): A's primary if in the common set, else B's, else the common id with the fewest catalogue files (tie: lexicographic); `primary_rule` is recorded",
          (p_both["primary"], p_both["primary_rule"]) == ("x", "both") and (p_a["primary"], p_a["primary_rule"]) == ("x", "A") and (p_b["primary"], p_b["primary_rule"]) == ("y", "B")
          and (p_few["primary"], p_few["primary_rule"]) == ("y", "fewest_files") and (p_tie["primary"], p_tie["primary_rule"]) == ("x", "fewest_files"), (p_both, p_a, p_b, p_few, p_tie))
    check("L agree(): disjoint acceptable sets, different covered, an unparsable session or a different none_kind go to the adjudicator",
          LX.agree(a, dict(a, acceptable=["z"], primary="z"))[0] == "disagree" and LX.agree(a, dict(a, covered=False, acceptable=[], primary=None, none_kind="gap"))[0] == "disagree" and LX.agree(a, None)[0] == "disagree"
          and LX.agree({"covered": False, "acceptable": [], "primary": None, "none_kind": "gap", "rule": "", "reason": ""}, {"covered": False, "acceptable": [], "primary": None, "none_kind": "new", "rule": "", "reason": ""})[0] == "disagree"
          and LX.agree({"covered": False, "acceptable": [], "primary": None, "none_kind": None, "rule": "", "reason": ""}, {"covered": False, "acceptable": [], "primary": None, "none_kind": None, "rule": "", "reason": ""})[0] == "agreed")
    # jev-state perception labelling: >= 20 units per question, both questions, from the dumped state (R1-16)
    stub_labeller.write_states(os.path.join(rd, "states.jsonl"), units, json.load(open(os.path.join(data, "catalogue", "catalogue.json"))), 5)
    r5 = run([PY, os.path.join(S, "label_x2d.py"), "--round", "1", "--kind", "jev-state"], cwd=ROOT, env=lab_env(data, tp, log + ".js"))
    lab5 = [x for x in jl(os.path.join(rd, "labels.jsonl")) if x["kind"] == "jev-state" and x["provenance"] == "agent_reviewed"]
    qs = {q: {x["case_id"] for x in lab5 if x["question"] == q} for q in ("merge.covered", "merge.which_area")}
    js_prompts = [json.loads(ln) for ln in open(log + ".js")]
    check("L perception labelling answers Jev's exact question from Jev's exact state: >= 20 units per question, selected by the smallest sha256('x2d-perception'), prompts carry state and catalogue and not the wide evidence (members' text)",
          r5.returncode == 0 and len(qs["merge.covered"]) >= 20 and qs["merge.covered"] == qs["merge.which_area"]
          and qs["merge.covered"] == {u["case_id"] for u in LX.perception_pick(units, 20)} and js_prompts and not any("poison" in p["prompt"] for p in js_prompts) and not stub_labeller.probe_leaks(log + ".js", POISON[1:]), (r5.returncode, r5.stderr[-300:]))
    check("L the perception prompt has the question instructions and criteria and the dumped state, and NOT the label guide (Jev never sees it, R1-16); a missing states.jsonl is refused",
          js_prompts and not any("=== LABEL GUIDE ===" in p["prompt"] for p in js_prompts) and any("Sample files:" in p["prompt"] for p in js_prompts)
          and run([PY, os.path.join(S, "label_x2d.py"), "--round", "1", "--kind", "jev-state", "--new-round", "nostates", "--states", os.path.join(tmp, "absent-states.jsonl")], cwd=ROOT, env=lab_env(data, tp)).returncode != 0)
    check("L R2-17 the perception prompts carry the dumped state's effective instructions and criteria (the V-sem question's wording incl. its delta), one per question",
          any("STUB-NOUL-INSTRUCTIONS" in p["prompt"] and "STUB-TRUE" in p["prompt"] for p in js_prompts) and any("STUB-CHOICE-INSTRUCTIONS" in p["prompt"] for p in js_prompts), js_prompts[:1])
    u0 = units[0]
    pick_pool = units + [dict(u0, case_id="1:aud", set="audit"), dict(u0, case_id="1:ctl", control=True, set="ctl"), dict(u0, case_id="1:ctl2", control=True)]
    check("L the perception sample is drawn from scored window units only (R1-16): audit units and controls are never picked", not ({"1:aud", "1:ctl", "1:ctl2"} & {u["case_id"] for u in LX.perception_pick(pick_pool, 1000)})
          and len(LX.perception_pick(pick_pool, 1000)) == len(units))
    # gold
    import gold as G
    env_old = dict(os.environ)
    os.environ["X2D_DATA"] = data
    try:
        grows = G.build_gold(1, data=data)
        gl = G.load_gold(1, data=data)
    finally:
        os.environ.clear()
        os.environ.update(env_old)
    bad_ok = all(gl[c]["agreement"] in ("agreed", "adjudicated") for c in gl) and bad_cid in gl
    check("L gold.jsonl is built from the wide final rows only (one per unit, agreed or adjudicated, the --new-round fill included) with covered / acceptable / primary / none_kind", len(gl) == len(units) and bad_ok
          and all(gl[c]["covered"] == gold[c]["covered"] for c in gl if gl[c]["agreement"] == "agreed"), len(gl))
    try:
        G.build_gold(1, data=data)
        refused = False
    except SystemExit:
        refused = True
    check("L gold.jsonl is append-only: building it twice is refused", refused)
    # session cap (R1-05)
    big, bunits, bgold, _i = lab_data(tmp, "cap", n=25)
    rows = jl(os.path.join(big, "rounds", "1", "units.jsonl"))
    extra = []
    for k in range(80):
        c = json.loads(json.dumps(rows[k % len(rows)]))
        c["case_id"] = "1:cap%03d" % k
        extra.append(c)
    wjl(os.path.join(big, "rounds", "1", "units.jsonl"), rows + extra)
    r6 = run([PY, os.path.join(S, "label_x2d.py"), "--round", "1"], cwd=ROOT, env=lab_env(big, tp, X2D_STUB_SESSION_CAP="50"))
    check("L the session cap refuses a plan of 2 sessions per unit beyond it before any session starts (stub cap 50, 105 units) and writes no labels", r6.returncode != 0 and "cap" in (r6.stderr + r6.stdout).lower()
          and not os.path.exists(os.path.join(big, "rounds", "1", "labels.jsonl")), r6.stderr[-200:])
    envs = ("X2D_LABEL_STUB", "X2D_REFERENCE_STUB", "X2D_STUB_SESSION_CAP")
    saved = {k: os.environ.pop(k, None) for k in envs}
    try:
        os.environ["X2D_STUB_SESSION_CAP"] = "5"
        c_real = LX.cap_for(1, W=105, J=20, R=20)
        os.environ["X2D_LABEL_STUB"] = "x"
        c_lab = LX.cap_for(1, W=105, J=20, R=20)
        del os.environ["X2D_LABEL_STUB"]
        os.environ["X2D_REFERENCE_STUB"] = "x"
        c_ref_cross = LX.cap_for(1, W=105, J=20, R=20)
        c_ref_own = LX.cap_for(1, W=105, J=20, R=20, stub_flag="X2D_REFERENCE_STUB")
        del os.environ["X2D_REFERENCE_STUB"]
        os.environ["X2D_LABEL_STUB"] = "x"
        c_lab_cross = LX.cap_for(1, W=105, J=20, R=20, stub_flag="X2D_REFERENCE_STUB")
        del os.environ["X2D_LABEL_STUB"]
        del os.environ["X2D_STUB_SESSION_CAP"]
        formula = (LX.cap_for(1, W=25, J=20, R=20), LX.cap_for(0, W=100, I=20), LX.cap_for(1, W=1000, J=20, R=20))
    finally:
        for k, v in saved.items():
            os.environ.pop(k, None)
            if v is not None:
                os.environ[k] = v
    check("L X2D_STUB_SESSION_CAP is validation-only: ignored when no stub env is set (cap = the formula 424 for W = 105, J = 20, R = 20), honoured under the caller's own stub flag (R1-05)",
          c_real == 424 and c_lab == 5 and c_ref_own == 5, (c_real, c_lab, c_ref_own))
    check("L R2-15 a cross stub-flag case keeps the formula cap: X2D_REFERENCE_STUB set with the default flag (label) -> 424; X2D_LABEL_STUB set with stub_flag=X2D_REFERENCE_STUB -> 424",
          c_ref_cross == 424 and c_lab_cross == 424, (c_ref_cross, c_lab_cross))
    check("L the session cap formula: ceil(1.25 (2.2 W + 4.4 J + R)) = 204 for (25, 20, 20); round 0 adds 2.2 per intruder item (ceil(1.25 * 2.2 * 120) = 330); the ceiling is 500", formula == (204, 330, 500), formula)
    c_formula = (LX.cap_for(1, W=35, J=20, R=20), LX.cap_for(1, W=35, J=20, R=20, C=1), LX.cap_for(1, W=1000, J=20, R=20, C=1))
    check("L R3-01 the clarification relabel is in the cap: ceil(1.25 (2.2 W (1 + C) + 4.4 J + R)) = 232 for (35, 20, 20) with C = 0 and 328 with C = 1 (§12's illustrative round), ceiling 500", c_formula == (232, 328, 500), c_formula)
    pdat, _pu, _pg, _pi = lab_data(tmp, "preclar", n=25)
    pre = lambda: run([PY, os.path.join(S, "label_x2d.py"), "--round", "1", "--precheck-clarify"], cwd=ROOT, env=lab_env(pdat))
    r_fit = pre()
    wjl(os.path.join(pdat, "rounds", "1", "calls.jsonl"), [{"case_id": "1:x", "kind": "wide", "label_set": "main", "A": {"attempts": 250, "ok": True}, "B": None, "C": None}])
    r_nofit = pre()
    check("L R3-01 `label_x2d --precheck-clarify` (run.sh clarify runs it BEFORE `versioning --clarify`): relabelling 25 wide cases fits the cap with C + 1 (273) when nothing is used (exit 0), is refused when 250 sessions are used (250 + 50 > 273) with the message that no revision was recorded, and writes nothing",
          r_fit.returncode == 0 and "fits" in r_fit.stdout and r_nofit.returncode != 0 and "no revision recorded" in (r_nofit.stderr + r_nofit.stdout) and not os.path.exists(os.path.join(pdat, "rounds", "1", "guide-revisions.jsonl"))
          and not os.path.exists(os.path.join(pdat, "rounds", "1", "labels.jsonl")), (r_fit.returncode, r_fit.stdout[-150:], r_nofit.returncode, (r_nofit.stderr + r_nofit.stdout)[-200:]))
    huge, _u, _g, _i2 = lab_data(tmp, "capreal", n=25)
    hr = jl(os.path.join(huge, "rounds", "1", "units.jsonl"))
    wjl(os.path.join(huge, "rounds", "1", "units.jsonl"), [dict(json.loads(json.dumps(hr[k % len(hr)])), case_id="1:h%04d" % k) for k in range(600)])
    nobin = os.path.join(tmp, "no-model-bin")
    os.makedirs(nobin, exist_ok=True)
    r7 = run([PY, os.path.join(S, "label_x2d.py"), "--round", "1"], cwd=ROOT, env=dict(lab_env(huge, None, X2D_STUB_SESSION_CAP="100000"), PATH=nobin))  # no stub env: the stub cap is ignored; no claude or codex on PATH
    check("L without a stub env a stub session cap of 100000 is ignored: 600 units (1200 planned sessions) are refused against the real ceiling of 500, before any session", r7.returncode != 0 and "cap" in (r7.stderr + r7.stdout).lower()
          and not os.path.exists(os.path.join(huge, "rounds", "1", "labels.jsonl")), (r7.returncode, r7.stderr[-200:]))
    # a case that reaches the cap mid-way writes its call row only (no labels) and still counts its sessions
    mid, munits, mgold, mids = lab_data(tmp, "capmid", n=3)
    mt = {u["case_id"]: {"A": _ans(mgold[u["case_id"]]), "B": stub_labeller._deviate(mgold[u["case_id"]], mids), "C": _ans(mgold[u["case_id"]])} for u in munits}
    mtp = os.path.join(tmp, "lab-truth-mid.json")
    json.dump(mt, open(mtp, "w"))
    r8 = run([PY, os.path.join(S, "label_x2d.py"), "--round", "1", "--workers", "1"], cwd=ROOT, env=lab_env(mid, mtp, X2D_STUB_SESSION_CAP="7"))
    mrd = os.path.join(mid, "rounds", "1")
    mcalls, mlab = jl(os.path.join(mrd, "calls.jsonl")), jl(os.path.join(mrd, "labels.jsonl"))
    msess = json.load(open(os.path.join(mrd, "sessions.json")))
    capped_calls = [c for c in mcalls if c.get("capped")]
    check("L a case that hits the cap mid-way writes its call row only (marked capped, no label rows) and its sessions are counted: stub cap 7, three units each needing A, B and C -> 2 labelled, 1 capped, used = 7 (R1-05)",
          r8.returncode != 0 and len(capped_calls) == 1 and capped_calls[0]["case_id"] not in {x["case_id"] for x in mlab} and len({x["case_id"] for x in mlab}) == 2 and msess["used"] == 7 and msess["cap"] == 7,
          (r8.returncode, len(mcalls), len(capped_calls), msess, r8.stderr[-200:]))
    reference_checks(tmp, data, units, gold, ids)


def primary_disagreement_checks(tmp):
    """R1-11: A and B agree on `covered` and share acceptable areas but name different primaries: the agreed primary is A's when in the common set, else B's, else the fewest-files area; gold absorbs it (never None -> unabsorbed)."""
    data, units, gold, ids = lab_data(tmp, "prim", n=4)
    cov = [u for u in units if gold[u["case_id"]]["covered"]]
    truth = {u["case_id"]: {k: _ans(gold[u["case_id"]]) for k in "ABC"} for u in units}
    c1, c2 = cov[0]["case_id"], cov[1]["case_id"]
    stub_labeller.plant_primary_disagreement(truth, c1, [ids[0], ids[1]], ids[0], ids[1])  # A's primary is in the common set
    mk = lambda acc, p: {"covered": True, "acceptable": list(acc), "primary": p, "none_kind": None, "rule": "stub", "reason": "stub"}
    truth[c2] = {"A": mk([ids[0], ids[1]], ids[0]), "B": mk([ids[1], ids[2]], ids[1]), "C": mk([ids[1]], ids[1])}  # only ids[1] is common: A's primary is not in it, B's is
    tp = os.path.join(tmp, "lab-truth-prim.json")
    json.dump(truth, open(tp, "w"))
    r = run([PY, os.path.join(S, "label_x2d.py"), "--round", "1", "--workers", "1"], cwd=ROOT, env=lab_env(data, tp))
    final = {x["case_id"]: x for x in jl(os.path.join(data, "rounds", "1", "labels.jsonl")) if x["by"] == "A+B agree"}
    check("L A and B agreeing on covered with different primaries are AGREED with a defined primary: A's when in the common set (rule A), else B's (rule B) (R1-11)",
          r.returncode == 0 and final[c1]["agreement"] == "agreed" and final[c1]["primary"] == ids[0] and final[c1].get("primary_rule") == "A" and final[c2]["agreement"] == "agreed"
          and final[c2]["primary"] == ids[1] and final[c2].get("primary_rule") == "B" and final[c2]["acceptable"] == [ids[1]], (r.returncode, r.stderr[-200:], {k: (v["primary"], v.get("primary_rule")) for k, v in final.items() if k in (c1, c2)}))
    import gold as G_
    env_old = dict(os.environ)
    os.environ["X2D_DATA"] = data
    try:
        G_.build_gold(1, data=data)
        gl = G_.load_gold(1, data=data)
        outs = G_._gold_outputs(units, gl)
    finally:
        os.environ.clear()
        os.environ.update(env_old)
    check("L the gold of those units carries the agreed primary, so they are absorbed into the next membership (not listed as unabsorbed)", gl[c1]["primary"] == ids[0] and gl[c2]["primary"] == ids[1] and outs[c1] == ids[0] and outs[c2] == ids[1], (outs.get(c1), outs.get(c2)))


def _ans(g):
    return {"covered": g["covered"], "acceptable": list(g["acceptable"]), "primary": g["primary"], "none_kind": g["none_kind"], "rule": "stub", "reason": "stub"}


def reference_checks(tmp, data, units, gold, ids):
    ref = {u["case_id"]: {"covered": gold[u["case_id"]]["covered"], "area_id": gold[u["case_id"]]["primary"]} for u in units}
    rp = os.path.join(tmp, "ref-truth.json")
    json.dump(ref, open(rp, "w"))
    env = lab_env(data, X2D_REFERENCE_STUB=rp)
    r = run([PY, os.path.join(S, "reference_arm.py"), "--round", "1"], cwd=ROOT, env=env)
    rows = jl(os.path.join(data, "rounds", "1", "arms", "reference.jsonl"))
    sel = [u["case_id"] for u in sorted(units, key=lambda u: x2d.reference_key(u["cluster_id"]))[:20]]
    check("L reference arm: exactly the 20 clusters with the smallest sha256('x2d-reference') are answered (assigned / none per the stub), one row each, arm reference-claude",
          r.returncode == 0 and [x["case_id"] for x in rows] == sel and all(x["arm"] == "reference-claude" for x in rows)
          and all((x["output"] == "assigned") == ref[x["case_id"]]["covered"] and (x["area_id"] == ref[x["case_id"]]["area_id"]) for x in rows), (r.returncode, r.stderr[-300:]))
    r2 = run([PY, os.path.join(S, "reference_arm.py"), "--round", "1"], cwd=ROOT, env=env)
    r0 = run([PY, os.path.join(S, "reference_arm.py"), "--round", "0"], cwd=ROOT, env=env)
    check("L the reference arm is append-only (a second run is refused) and is not used in round 0", r2.returncode != 0 and r0.returncode != 0, (r2.stderr[-100:], r0.stderr[-100:]))


# ---------------------------------------------------------------- S statistics: GO rule, futility, P2, P3, N2, N3
import analysis_x2d as A  # noqa: E402
import power_sim as PS  # noqa: E402
import gold as G  # noqa: E402

THR = json.load(open(os.path.join(S, "thresholds.json")))


def rg(k):
    return {"case_id": "c%d" % k, "arm": "x"}


def stats_primitives():
    lo, hi = A.wilson(18, 20)
    check("S Wilson 95 % interval: 18/20 -> [0.699, 0.972] (the memo's figure); 0/0 is undefined", abs(lo - 0.699) < 0.002 and abs(hi - 0.972) < 0.002 and math.isnan(A.wilson(0, 0)[0]))
    lo1, _ = A.wilson_bounds_one_sided(57, 63, 0.90)
    check("S the one-sided 90 % Wilson lower bound is above the two-sided 90 % lower bound's z = 1.2816 form and below the point", 0.78 < lo1 < 57 / 63 and abs(lo1 - A.wilson(57, 63, 1.2815515655446004)[0]) < 1e-9, lo1)
    check("S exact McNemar: (b, c) = (5, 0) -> p = 0.0625; (3, 3) -> p = 1; (0, 0) -> 1; symmetric", abs(A.mcnemar_exact(5, 0) - 0.0625) < 1e-12 and A.mcnemar_exact(3, 3) == 1.0 and A.mcnemar_exact(0, 0) == 1.0 and A.mcnemar_exact(0, 7) == A.mcnemar_exact(7, 0))
    check("S Cohen's kappa: perfect = 1, chance-level agreement ~ 0, systematic disagreement < 0",
          A.cohen_kappa([1, 0, 1, 0], [1, 0, 1, 0]) == 1.0 and abs(A.cohen_kappa([1, 1, 0, 0], [1, 0, 1, 0])) < 1e-9 and A.cohen_kappa([1, 1, 0, 0], [0, 0, 1, 1]) < 0)
    rows = [{"pr_key": "p%d" % (i % 8), "n": 1, "d": 1 if i % 4 else 0} for i in range(40)]
    b1 = A.pr_cluster_bootstrap(rows, A.gain_stat, B=500, seed=1)
    b2 = A.pr_cluster_bootstrap(rows, A.gain_stat, B=500, seed=1)
    b3 = A.pr_cluster_bootstrap(rows, A.gain_stat, B=500, seed=2)
    pt = sum(r["d"] for r in rows) / len(rows)
    check("S the PR-cluster bootstrap resamples whole PRs, is deterministic for a seed (differs for another), has the point estimate of the data and nests its levels (80 % within 90 % within 95 %)",
          b1 == b2 and b1["two_sided"] != b3["two_sided"] and abs(b1["point"] - pt) < 1e-12 and b1["two_sided"][0.95][0] <= b1["two_sided"][0.9][0] <= b1["two_sided"][0.8][0] <= pt <= b1["two_sided"][0.8][1] <= b1["two_sided"][0.9][1] <= b1["two_sided"][0.95][1])
    gold = {"covered": True, "acceptable": ["a", "b"], "primary": "a", "none_kind": None}
    gnone = {"covered": False, "acceptable": [], "primary": None, "none_kind": "new"}
    so = A.score_outcome
    row = lambda out, a=None: {"case_id": "x", "arm": "x", "output": out, "area_id": a}
    check("S outcome scoring (section 9): assigned in the acceptable set is correct; assigned outside it is wrong; assigned to a gold-none cluster is wrong; none is correct only for gold none (gap and new kept apart); unresolved is never correct; missing and ungolded are not counted",
          so(row("assigned", "b"), gold)["kind"] == "assigned_correct" and so(row("assigned", "z"), gold)["kind"] == "assigned_wrong" and so(row("assigned", "a"), gnone)["kind"] == "assigned_gold_none"
          and so(row("none"), gnone)["kind"] == "none_correct_new" and so(row("none"), dict(gnone, none_kind="gap"))["kind"] == "none_correct_gap" and so(row("none"), gold)["kind"] == "none_wrong"
          and so(row("unresolved"), gold)["correct"] is False and so(None, gold)["kind"] == "missing" and so(row("none"), None)["kind"] == "ungolded")
    outs = [so(row("assigned", "a"), gold), so(row("assigned", "z"), gold), so(row("none"), gold), so(row("unresolved"), gold), so(row("none"), gnone), so(row("assigned", "a"), gnone)]
    c = A.counts_for(outs)
    rt = A.rates(c)
    check("S the section 9 rates: P_assign = correct / assigned-to-any-area, R_assign = correct / gold-covered, P_new = none-and-gold-none / none, R_new = none-and-gold-none / gold-none, unresolved counted apart",
          (c["assigned"], c["assigned_correct"], c["gold_covered"], c["none"], c["none_correct"], c["gold_none"], c["unresolved"]) == (3, 1, 4, 2, 1, 2, 1)
          and abs(rt["P_assign"]["value"] - 1 / 3) < 1e-12 and abs(rt["R_assign"]["value"] - 1 / 4) < 1e-12 and abs(rt["P_new"]["value"] - 1 / 2) < 1e-12 and abs(rt["R_new"]["value"] - 1 / 2) < 1e-12)
    check("S R4-13 rates: R_assign_res = correct / (gold-covered AND resolved): an abstained covered cluster leaves the denominator of R_assign_res and stays in R_assign's; a false `none` on a covered cluster is a resolved wrong answer; unresolved counted by reason",
          c["gold_covered_res"] == 3 and abs(rt["R_assign_res"]["value"] - 1 / 3) < 1e-12 and rt["R_assign_res"]["n"] == 3 and rt["R_assign"]["n"] == 4,
          (c.get("gold_covered_res"), rt["R_assign_res"], rt["R_assign"]))


def pooled_from(structure, p_j, p_b2, seed, corr=PS.CORR, perception=True, n1=True, informative=True, B=2000):
    pl = PS.plant(structure, p_j, p_b2, seed=seed, corr=corr)
    gold = pl["gold"]
    # 10 real gold-none positives, correctly handled by both arms, so the P_new/R_new rows exist
    pooled = A.summarise(pl["arms"], gold, pl["units"], informative_all=informative, perception={"pass": perception}, n1={"pass": n1}, B=B)
    return pl, pooled


def mini_world():
    """R4-13 hand-made world (4 PRs, 8 clusters): the figures the resolved-cluster rule must report, worked out by hand in the comments of abstention_checks."""
    spec = [  # (case, pr, gold covered / acceptable, Jev (output, area, reason), B2 (output, area))
        ("c1", "p1", ["A"], ("assigned", "A", None), ("assigned", "B")),
        ("c2", "p1", ["A"], ("unresolved", None, "below_t_choice"), ("assigned", "A")),
        ("c3", "p2", ["A"], ("assigned", "A", None), ("assigned", "A")),
        ("c4", "p2", None, ("none", None, None), ("assigned", "A")),
        ("c5", "p3", ["A"], ("assigned", "B", None), ("assigned", "A")),
        ("c6", "p3", ["A"], ("unresolved", None, "failure"), ("assigned", "B")),
        ("c7", "p4", ["A"], ("none", None, None), ("assigned", "A")),
        ("c8", "p4", ["A"], ("unresolved", None, "permutation_disagreement"), ("assigned", "A")),
    ]
    units = [{"case_id": c, "pr_key": pr, "control": False, "set": "window", "stratum": 1} for c, pr, *_ in spec]
    gold = {c: ({"case_id": c, "covered": True, "acceptable": acc, "primary": acc[0], "none_kind": None} if acc else {"case_id": c, "covered": False, "acceptable": [], "primary": None, "none_kind": "gap"}) for c, _pr, acc, *_ in spec}
    arms = {"V": [{"case_id": c, "arm": "V", "output": j[0], "area_id": j[1], "unresolved_reason": j[2]} for c, _pr, _a, j, _b in spec],
            "B2": [{"case_id": c, "arm": "B2", "output": b[0], "area_id": b[1], "unresolved_reason": None} for c, _pr, _a, _j, b in spec]}
    return units, gold, arms


def abstention_checks():
    """R4-13 (2), (3), R4-10: the resolved-cluster pairing, the abstention figures, the zero-assignment and zero-resolved summaries."""
    units, gold, arms = mini_world()
    # Jev: assigned c1 (right) c3 (right) c5 (wrong); none c4 (right) c7 (false none); unresolved c2 c6 c8. Resolved = c1 c3 c4 c5 c7; paired d on them +1 0 +1 -1 -1 -> gain_res 0, jev-only c1 c4 (2 PRs), base-only c5 c7.
    # B2 correct on c2 c3 c5 c7 c8 (5 of 8); Jev correct on c1 c3 c4 -> gain_all = (3 - 5) / 8 = -0.25. gold-covered 7, of them resolved 4 (c1 c3 c5 c7): R_assign_res = 2 / 4, R_assign = 2 / 7. P_assign = 2 / 3.
    p = A.summarise(arms, gold, units, jev="V", base="B2", B=50)
    ab, un = p["abstained"], p["unresolved"]
    check("S R4-13 (3) resolved pairing: gain_res is over exactly the 5 clusters Jev resolved (B2 scored on the same ones; the 3 abstained are excluded, not scored as Jev misses): n_res 5, gain_res 0, Jev-only / B2-only 2 / 2 from 2 PRs",
          p["n_res"] == 5 and p["gain_res"]["n"] == 5 and p["gain_res"]["point"] == 0.0 and (p["gain_res"]["jev_only"], p["gain_res"]["base_only"]) == (2, 2) and p["gain_res"]["n_jev_only_prs"] == 2, p["gain_res"])
    check("S R4-13 guard figure: gain_all scores every abstained cluster as a Jev miss over all 8 clusters ((3 - 5) / 8 = -0.25), with its McNemar counts",
          p["gain_all"]["n"] == 8 and abs(p["gain_all"]["point"] + 0.25) < 1e-12 and (p["gain_all"]["jev_only"], p["gain_all"]["base_only"]) == (2, 4), p["gain_all"])
    check("S R4-13 estimators: P_assign = 2 / 3 (a wrong assignment counts), R_assign_res = 2 / 4 (resolved covered clusters), R_assign = 2 / 7 (abstention = miss)",
          (p["p_assign"]["k"], p["p_assign"]["n"]) == (2, 3) and (p["r_assign_res"]["k"], p["r_assign_res"]["n"]) == (2, 4) and (p["r_assign"]["k"], p["r_assign"]["n"]) == (2, 7), (p["p_assign"], p["r_assign_res"], p["r_assign"]))
    check("S R4-13 unresolved block: u = 3 / 8 with Wilson 95 % and the per-reason split (below_t_choice 1, permutation_disagreement 1, failure 1, below_t_cov 0), nothing missing or ungolded",
          (un["k"], un["n"]) == (3, 8) and abs(un["rate"] - 3 / 8) < 1e-12 and un["by_reason"] == {"below_t_cov": 0, "below_t_choice": 1, "permutation_disagreement": 1, "failure": 1, "other": 0}
          and un["wilson95"][0] < 3 / 8 < un["wilson95"][1] and un["missing"] == un["ungolded"] == 0, un)
    check("S R4-13 abstention table: B2 is right on 2 of the 3 Jev-abstained clusters, on 3 of the 5 resolved ones and on 5 of all 8 (descriptive; decides nothing)",
          (ab["n_abstained"], ab["b2_acc_abstained"]["k"], ab["b2_acc_abstained"]["n"], ab["b2_acc_resolved"]["k"], ab["b2_acc_resolved"]["n"], ab["b2_acc_all"]["k"], ab["b2_acc_all"]["n"]) == (3, 2, 3, 3, 5, 5, 8), ab)
    check("S R4-13 real new-area figures are descriptive with intervals: 1 real positive (c4), P_new = 1 / 2 and R_new = 1 / 1 with Wilson 95 %, the note 'no veto'",
          p["real_new"]["n"] == 1 and p["real_new"]["P_new"] == 0.5 and p["real_new"]["R_new"] == 1.0 and len(p["real_new"]["P_new_wilson95"]) == 2 and "no veto" in p["real_new"]["note"], p["real_new"])
    cases = A.build_cases(units, gold, arms)
    rows = A.pooling_rows([k for k in cases if not k["control"]], "V", "B2")
    keys = {"pr_key", "n_m", "n_res", "d_res", "a", "ac", "gc", "gc_res", "n_all", "d_all", "d_guess", "abst", "abst_b2_n", "abst_b2_correct", "b2_res"}
    check("S R4-13 per-PR pooling rows carry n_res, d_res, a, ac, gc_res, n_all, d_all, abst (and the B2 outcomes on abstained / resolved clusters) and sum to the pooled figures",
          all(set(r) == keys for r in rows) and sum(r["n_res"] for r in rows) == 5 and sum(r["d_res"] for r in rows) == 0 and sum(r["n_all"] for r in rows) == 8 and sum(r["d_all"] for r in rows) == -2
          and sum(r["d_guess"] for r in rows) == -2 and sum(r["abst"] for r in rows) == 3 and sum(r["gc_res"] for r in rows) == 4 and sum(r["a"] for r in rows) == 3 and sum(r["ac"] for r in rows) == 2, rows)
    # R5-01: gain_guess scores an abstained cluster by Jev's own unthresholded guess (a missing raw answer is wrong). The mini world has no raw fields: every abstention is a wrong guess, so gain_guess = gain_all = -0.25.
    check("S R5-01 gain_guess with no raw answer on the 3 abstained clusters (a missing raw answer is WRONG): equals gain_all, (3 - 5) / 8 = -0.25 over all 8 clusters; Jev-only / B2-only 2 / 4",
          p["gain_guess"]["n"] == 8 and abs(p["gain_guess"]["point"] + 0.25) < 1e-12 and (p["gain_guess"]["jev_only"], p["gain_guess"]["base_only"]) == (2, 4), p["gain_guess"])
    raw = {"c2": (True, "A"), "c6": (True, "B"), "c8": (False, None)}  # right (covered, area A), wrong area, wrong `none` on covered gold
    arms_raw = {"V": [dict(r, **({"raw_covered": raw[r["case_id"]][0], "raw_choice": raw[r["case_id"]][1]} if r["case_id"] in raw else {})) for r in arms["V"]], "B2": arms["B2"]}
    pr_ = A.summarise(arms_raw, gold, units, jev="V", base="B2", B=50)
    check("S R5-01 gain_guess with raw answers on the abstained clusters (c2 right area A, c6 wrong area B, c8 `none` on covered gold = wrong): the guess is right on c1 c3 c4 c2 -> (4 - 5) / 8 = -0.125; gain_all and gain_res are unchanged (-0.25, 0)",
          abs(pr_["gain_guess"]["point"] + 0.125) < 1e-12 and abs(pr_["gain_all"]["point"] + 0.25) < 1e-12 and pr_["gain_res"]["point"] == 0.0 and pr_["gain_guess"]["lb95"] is not None, (pr_["gain_guess"], pr_["gain_all"]))
    so_ = [A.score_outcome({"case_id": "g", "output": "unresolved", "unresolved_reason": "failure", "raw_covered": rc, "raw_choice": ch}, g_) for rc, ch, g_ in
           ((False, None, {"covered": False, "acceptable": []}), (False, None, {"covered": True, "acceptable": ["A"]}), (True, "A", {"covered": True, "acceptable": ["A", "B"]}), (True, "C", {"covered": True, "acceptable": ["A"]}),
            (True, "A", {"covered": False, "acceptable": []}), (None, None, {"covered": True, "acceptable": ["A"]}), (True, None, {"covered": True, "acceptable": ["A"]}))]
    check("S R5-01 score_outcome raw_correct / guess_correct: `none` iff raw_covered is False (right on gold none, wrong on covered), an area right iff gold covered and in acceptable, a missing raw answer is wrong; an abstained row's guess_correct is its raw_correct, a resolved row's is its correct",
          [o["raw_correct"] for o in so_] == [True, False, True, False, False, False, False] and [o["guess_correct"] for o in so_] == [True, False, True, False, False, False, False]
          and A.score_outcome({"output": "assigned", "area_id": "A", "raw_covered": False, "raw_choice": None}, {"covered": True, "acceptable": ["A"]})["guess_correct"] is True
          and A.score_outcome({"output": "assigned", "area_id": "A", "raw_covered": False, "raw_choice": None}, {"covered": True, "acceptable": ["A"]})["raw_correct"] is False, [(o["raw_correct"], o["guess_correct"]) for o in so_])
    # R4-10 / R4-13 (8): an arm with no assignment over covered gold, and an arm with no resolved cluster, still summarise and give a verdict
    none_only = {"V": [dict(r, output="none", area_id=None, unresolved_reason=None) for r in arms["V"]], "B2": arms["B2"]}
    q = A.summarise(none_only, gold, units, jev="V", base="B2", B=50)
    vq = A.go_verdict(q, THR)
    check("S R4-10: an arm that never assigns (every output `none`) over covered gold summarises (P_assign undefined, R_assign_res = 0.0) and verdict() completes",
          q["p_assign"]["point"] is None and q["r_assign_res"]["point"] == 0.0 and q["r_assign"]["point"] == 0.0 and vq["verdict"] in ("NO-GO", "INCONCLUSIVE"), (q["p_assign"], q["r_assign_res"], vq["verdict"]))
    all_unres = {"V": [dict(r, output="unresolved", area_id=None, unresolved_reason="failure") for r in arms["V"]], "B2": arms["B2"]}
    z = A.summarise(all_unres, gold, units, jev="V", base="B2", B=50)
    vz = A.go_verdict(z, THR)
    check("S R4-13 (8): an arm with zero resolved clusters summarises (n_res 0, gain_res undefined, u = 1) and the verdict is INCONCLUSIVE when the round is not informative (informative_all unknown); with every round informative the cap breach is NO-GO",
          z["n_res"] == 0 and z["gain_res"]["point"] is None and z["gain_res"]["lb90"] is None and z["unresolved"]["rate"] == 1.0 and vz["verdict"] == "INCONCLUSIVE"
          and A.go_verdict(dict(z, informative_all=True), THR)["verdict"] == "NO-GO", (z["gain_res"], vz))

    # R5-09: every GO-population case needs a primary-arm row and a B2 outcome (pooling_rows / paired silently drop a cluster otherwise)
    cases_all = A.build_cases(units, gold, arms)
    A._require_complete([k for k in cases_all if not k["control"]], "V", "B2", "fixture")
    arms_no_b2 = {"V": arms["V"], "B2": [r for r in arms["B2"] if r["case_id"] != "c5"]}
    cases_gap = A.build_cases(units, gold, arms_no_b2)
    try:
        A._require_complete([k for k in cases_gap if not k["control"]], "V", "B2", "fixture")
        refused_c5 = None
    except SystemExit as ex_:
        refused_c5 = str(ex_)
    check("S R5-09 an unfinished GO population is refused, naming the case: with the B2 row of c5 removed `_require_complete` exits naming c5 (a complete population passes silently), and the gain_guess / gain_all / unresolved populations of a complete world are all 8",
          refused_c5 is not None and "c5" in refused_c5 and "B2" in refused_c5 and (p["gain_guess"]["n"], p["gain_all"]["n"], p["unresolved"]["n"], p["n"]) == (8, 8, 8, 8), (refused_c5, p["gain_guess"]["n"], p["gain_all"]["n"], p["unresolved"]["n"], p["n"]))
    # R5-07: planted-new joint counts: a trap labelled COVERED counts in neither numerator; R_new = sum k / sum n (arm none AND gold not covered / gold not covered), P_new = sum k / sum none_n (arm none)
    t_units = [{"case_id": "t%d" % i, "control": True, "control_kind": "trap", "pr_key": "ptrap"} for i in range(4)]
    t_gold = {"t0": {"case_id": "t0", "covered": False, "acceptable": [], "none_kind": "new"}, "t1": {"case_id": "t1", "covered": False, "acceptable": [], "none_kind": "new"},
              "t2": {"case_id": "t2", "covered": True, "acceptable": ["A"], "primary": "A"}}  # t3 has no gold
    t_arm = {"t0": {"case_id": "t0", "output": "none"}, "t1": {"case_id": "t1", "output": "assigned", "area_id": "A"}, "t2": {"case_id": "t2", "output": "none"}, "t3": {"case_id": "t3", "output": "none"}}
    cs = A.controls_section(t_units, t_gold, t_arm, "V")
    pn = A._planted_new_from_rounds({8: {"controls": json.loads(json.dumps(cs, default=dict))}, 9: {"controls": json.loads(json.dumps(cs, default=dict))}})
    check("S R5-07 / R6-05 the trap controls carry the joint counts none_correct {k: arm none AND gold not covered, n: gold not covered} and none_n (arm none), only over traps with gold and an arm row: t0 right, t1 (gold none, assigned) wrong, t2 (labelled covered, arm none) in NEITHER numerator, t3 (no gold) in no count and unmeasured; "
          "pooled over two rounds R_new = 2 / 4 and P_new = 2 / 4 (arm none on t0 and t2 per round), Wilson 95 %, with the source text",
          cs["trap"]["none_correct"] == {"k": 1, "n": 2} and cs["trap"]["none_n"] == 2 and cs["unmeasured"]["trap_gold"] == 1 and (pn["R_new"]["k"], pn["R_new"]["n"]) == (2, 4) and (pn["P_new"]["k"], pn["P_new"]["n"]) == (2, 4) and pn["unmeasured"] == 2
          and "empty-file rename traps" in pn["source"] and "gates.json" in pn["source"] and len(pn["R_new"]["wilson95"]) == 2, (cs["trap"], cs["unmeasured"], pn))



def verdict_checks():
    structure = PS.empirical_structure()
    pl, pooled = pooled_from(structure, 0.99, 0.55, seed=11)
    mins = THR
    v = A.go_verdict(pooled, mins)
    check("S GO fires on a planted large effect (Jev 0.99, B2 0.55 on the real round structure) when every condition holds: P_assign point and Wilson bound, R_assign_res, gain_res point and bootstrap bound, >= 3 PRs, u <= 0.15, gain_guess 95 % bound > 0 (guard h), n_res >= 50, perception, N1, informative", v["verdict"] == "GO" and len(v["conditions"]) == 12, v)
    base = pooled
    mod = lambda **kw: A.go_verdict(json.loads(json.dumps(dict(base, **kw))), mins)
    pa, rr, gr, ga, gg, un = dict(base["p_assign"]), dict(base["r_assign_res"]), dict(base["gain_res"]), dict(base["gain_all"]), dict(base["gain_guess"]), dict(base["unresolved"])
    cases = [
        ("P_assign point below 0.90 -> NO-GO", mod(p_assign=dict(pa, point=0.89))["verdict"], "NO-GO"),
        ("R_assign_res point below 0.80 -> NO-GO", mod(r_assign_res=dict(rr, point=0.79))["verdict"], "NO-GO"),
        ("gain_res point below +0.10 -> NO-GO", mod(gain_res=dict(gr, point=0.09))["verdict"], "NO-GO"),
        ("unresolved rate 0.16 (the cap breached) -> NO-GO even though the bounds, the spread and the guard pass", mod(unresolved=dict(un, rate=0.16))["verdict"], "NO-GO"),
        ("unresolved rate exactly 0.15 -> still GO (the cap is <=)", mod(unresolved=dict(un, rate=0.15))["verdict"], "GO"),
        ("P_assign 90 % Wilson lower bound below 0.80 with the point conditions met -> INCONCLUSIVE", mod(p_assign=dict(pa, wilson_lb90=0.79))["verdict"], "INCONCLUSIVE"),
        ("gain_res bootstrap lower bound not above 0 -> INCONCLUSIVE", mod(gain_res=dict(gr, lb90=0.0))["verdict"], "INCONCLUSIVE"),
        ("Jev-correct/B2-wrong resolved clusters from only 2 PRs -> INCONCLUSIVE", mod(gain_res=dict(gr, n_jev_only_prs=2))["verdict"], "INCONCLUSIVE"),
        ("the guard alone (gain_guess 95 % lower bound 0.0, every other condition true) -> INCONCLUSIVE", mod(gain_guess=dict(gg, lb95=0.0))["verdict"], "INCONCLUSIVE"),
        ("the guard is gain_guess, not gain_all: gain_all point 0.0 and its bound below 0 with gain_guess lb95 > 0 -> still GO", mod(gain_all=dict(ga, point=0.0, lb90=-0.1))["verdict"], "GO"),
        ("the guard bound missing (None: no resamplable row) -> INCONCLUSIVE", mod(gain_guess=dict(gg, lb95=None))["verdict"], "INCONCLUSIVE"),
        ("n_res 49 (< 50 resolved paired clusters) -> INCONCLUSIVE", mod(n_res=49)["verdict"], "INCONCLUSIVE"),
        ("P_new 0.3 on 12 real positives -> no effect on the verdict (new-area figures are descriptive: ~10 cases cannot veto)", mod(real_new={"n": 12, "P_new": 0.3, "R_new": 0.3})["verdict"], "GO"),
        ("no real positives at all -> no effect on the verdict", mod(real_new={"n": 0, "P_new": None, "R_new": None})["verdict"], "GO"),
        ("perception check failed -> NO-GO", mod(perception={"pass": False})["verdict"], "NO-GO"),
        ("N1 not at chance -> NO-GO", mod(n1={"pass": False})["verdict"], "NO-GO"),
        ("perception unmeasured (None) -> INCONCLUSIVE", mod(perception={"pass": None})["verdict"], "INCONCLUSIVE"),
        ("a frozen round not informative -> INCONCLUSIVE", mod(informative_all=False)["verdict"], "INCONCLUSIVE"),
        ("a frozen round not informative AND a point condition failing (cap breach) -> INCONCLUSIVE (R1-19: informativeness precedes every point test)", mod(informative_all=False, unresolved=dict(un, rate=0.5))["verdict"], "INCONCLUSIVE"),
        ("a point condition failing AND a perception failure -> NO-GO", mod(p_assign=dict(pa, point=0.5), perception={"pass": False})["verdict"], "NO-GO"),
        ("perception failed while only a bound fails -> NO-GO (a measured failure beats INCONCLUSIVE)", mod(perception={"pass": False}, p_assign=dict(pa, wilson_lb90=0.79))["verdict"], "NO-GO"),
        ("N1 failed while the PR spread fails -> NO-GO", mod(n1={"pass": False}, gain_res=dict(gr, n_jev_only_prs=2))["verdict"], "NO-GO"),
        ("N1 unmeasured while a bound fails -> INCONCLUSIVE", mod(n1={"pass": None}, gain_res=dict(gr, lb90=0.0))["verdict"], "INCONCLUSIVE"),
    ]
    for name, got, want in cases:
        check("S GO rule: " + name, got == want, got)
    vg = mod(gain_guess=dict(gg, lb95=0.0))
    kinds = {k: c["kind"] for k, c in vg["conditions"].items()}
    check("S GO rule kinds: the cap and the three other minimums are `point` conditions, the guard is `guard`, n_res is `data`; the INCONCLUSIVE reason for the guard says Jev shows no accuracy edge once its abstentions are answered by its own unthresholded guess",
          sorted(k for k, t in kinds.items() if t == "point") == sorted(k for k in kinds if k.startswith(("P_assign point", "R_assign_res", "gain_res point", "unresolved rate"))) and list(kinds.values()).count("guard") == 1
          and list(kinds.values()).count("data") == 1 and any("no accuracy edge over B2" in r for r in vg["reasons"]) and A.GO["guard_level"] == x2d.GO_GUARD_LEVEL == 0.95, (kinds, vg["reasons"]))
    pl2, pooled2 = pooled_from(structure, 0.80, 0.80, seed=12, corr=0.0)
    check("S the same rule on a planted null (Jev = B2 = 0.80, independent errors) does not fire GO", A.go_verdict(pooled2, mins)["verdict"] != "GO")
    # exact counts of a planted effect through summarise
    cases_ = A.build_cases(pl["units"], pl["gold"], pl["arms"])
    okj = sum(1 for c in cases_ if c["out"]["V-sem"]["correct"])
    okb = sum(1 for c in cases_ if c["out"]["B2"]["correct"])
    n = len(cases_)
    check("S planted effect, exact counts: summarise reports P_assign = R_assign_res = R_assign = Jev correct / n, gain_res = gain_all = (Jev correct - B2 correct) / n, n_res = n and u = 0 for the planted arms on 91 clusters from 40 PRs",
          n == 91 and pooled["n"] == 91 and pooled["n_prs"] == 40 and abs(pooled["p_assign"]["point"] - okj / n) < 1e-12 and abs(pooled["r_assign"]["point"] - okj / n) < 1e-12 and abs(pooled["r_assign_res"]["point"] - okj / n) < 1e-12
          and abs(pooled["gain_res"]["point"] - (okj - okb) / n) < 1e-12 and abs(pooled["gain_all"]["point"] - (okj - okb) / n) < 1e-12 and pooled["n_res"] == 91 and pooled["unresolved"]["k"] == 0, (pooled["p_assign"], okj, okb))
    gains = []
    for s in range(30):
        p_, pd = pooled_from(structure, 0.8, 0.8, seed=100 + s, corr=0.0, B=50)
        gains.append(pd["gain_res"]["point"])
    mg = sum(gains) / len(gains)
    check("S a planted null (p_J = p_B2 = 0.8) shows no effect: the mean resolved-cluster paired gain over 30 plants is within 0.04 of 0", abs(mg) < 0.04, mg)
    gains = [pooled_from(structure, 0.95, 0.55, seed=200 + s, B=50)[1]["gain_res"]["point"] for s in range(10)]
    check("S a planted effect (0.95 vs 0.55) is recovered: the mean resolved-cluster paired gain over 10 plants is within 0.08 of 0.40", abs(sum(gains) / len(gains) - 0.40) < 0.08, sum(gains) / len(gains))
    abstention_checks()


def perception_order_checks():
    """R1-01: a jev-state labeller writes two rows per unit (one per question); the perception counts must not depend on the order they were written in."""
    areas = ["a%d" % i for i in range(6)]
    cases, q1_rows, q2_rows = [], [], []
    row = lambda cid, q, by, **kw: dict({"case_id": cid, "kind": "jev-state", "label_set": "main", "by": by, "question": q, "provenance": "agent"}, **kw)
    for i in range(22):
        cid = "8:u%02d" % i
        covered, area = i % 5 != 0, areas[i % 6]
        cases.append({"case_id": cid, "control": False, "gold": {"covered": covered, "acceptable": [area] if covered else []},
                      "out": {"V-sem": {"output": "assigned" if covered else "none", "area_id": area if covered else None, "raw_covered": covered, "raw_choice": area if covered else None}}})
        for who in ("labeller_A:claude", "labeller_B:codex", "A+B agree"):
            q1_rows.append(row(cid, "merge.covered", who, covered=covered, acceptable=[], primary=None))
            q2_rows.append(row(cid, "merge.which_area", who, covered=True, acceptable=[area], primary=area))
    interleaved = q1_rows + q2_rows
    random.Random(3).shuffle(interleaved)
    by_unit = sorted(q1_rows + q2_rows, key=lambda r: (r["case_id"], r["question"] != "merge.which_area"))
    orders = {"q1 then q2": q1_rows + q2_rows, "q2 then q1": q2_rows + q1_rows, "interleaved (shuffled)": interleaved, "per unit, q2 before q1": by_unit}
    counts = {k: A.perception_counts(v, cases, "V-sem") for k, v in orders.items()}
    first = counts["q1 then q2"]
    check("S perception counts are identical for the two question rows written in either order and interleaved (R1-01): 22 units with both questions final, Jev agrees with the combined answer on all, labeller pairs agree on all; Q1 on 22 units, Q2 on the 17 covered ones",
          all(c == first for c in counts.values()) and first["n"] == 22 and first["jev_agree_k"] == 22 and first["ll_n"] == 22 and first["ll_agree_k"] == 22 and first["q1"]["n"] == 22 and first["q2"]["n"] == 17, counts)
    only_q1 = A.perception_counts(q1_rows, cases, "V-sem")
    check("S a unit counts in the perception check only when BOTH questions are final: with only the covered-question rows no unit counts and the pooled check is unmeasured (never pass by a half answer)", only_q1["n"] == 0 and A.perception_pass([only_q1])["pass"] is None)
    wrong = [dict(c, out={"V-sem": {"output": "assigned", "area_id": "zz", "raw_covered": True, "raw_choice": "zz"}}) for c in cases]
    pw = A.perception_counts(q1_rows + q2_rows, wrong, "V-sem")
    check("S Jev's raw answer must equal the combined labeller answer (none, or the which_area primary); a wrong area is a disagreement", pw["jev_agree_k"] == 0 and A.perception_pass([pw])["pass"] is False)
    no_raw = [dict(c, out={"V-sem": {"output": "unresolved", "area_id": None}}) for c in cases]
    check("S a V-sem row without a raw answer (request failed) counts as disagreement in the perception check", A.perception_counts(q1_rows + q2_rows, no_raw, "V-sem")["jev_agree_k"] == 0)
    # R2-09: 15 % unresolved (every raw answer correct): the headline agreement is the RAW one, the thresholded one is lower and reported
    unres = [dict(c, out={"V-sem": dict(c["out"]["V-sem"], output="unresolved", area_id=None)}) if i % 7 == 1 else c for i, c in enumerate(cases)]
    pu = A.perception_counts(q1_rows + q2_rows, unres, "V-sem", primary="V-sem")
    pp = A.perception_pass([pu])
    check("S R2-09 a stub world with ~14 % unresolved outputs and every raw answer correct passes perception on raw agreement (jev_agree 1.0); the thresholded agreement is lower (19 of 22) and reported with the unresolved rate",
          pu["jev_agree_k"] == 22 and pu["jev_thresholded_k"] == 19 and pu["unresolved_k"] == 3 and pp["pass"] is True and pp["jev_agree"] == 1.0 and abs(pp["jev_agree_thresholded"] - 19 / 22) < 1e-12
          and abs(pp["unresolved_rate"] - 3 / 22) < 1e-12 and pp["primary_agree"] is not None, (pu, pp))
    pm = A.perception_counts(q1_rows + q2_rows, cases, "V-sem", primary="fan-out")
    check("S R2-17 perception compares V-sem even when the primary arm is fan-out: the primary arm's agreement is only reported (here fan-out rows are absent: 0 agreements), V-sem's headline stays 22 of 22",
          pm["jev_agree_k"] == 22 and pm["primary_k"] == 0 and pm["primary_arm"] == "fan-out", pm)


def kappa_answer_checks():
    """R1-06: informativeness and K4 use kappa of the JOINT answer; kappa on `covered` degenerates at high prevalence and is only reported."""
    labels = []
    for i in range(20):
        area = "a%d" % (i % 5)
        for who, covered in (("labeller_A:claude", True), ("labeller_B:codex", i != 7)):
            labels.append({"case_id": "1:k%02d" % i, "kind": "wide", "label_set": "main", "by": who, "covered": covered, "acceptable": [area] if covered else [], "primary": area if covered else None,
                           "none_kind": None if covered else "gap", "rule": "R%d" % (1 + i % 4)})
    la = A.label_agreement(labels, None)
    inf = A.informative({"n_clusters": 25, "n_prs": 8, "kappa_covered": la["kappa_covered"], "kappa_answer": la["kappa_answer"], "raw_agreement_covered": la["raw_agreement_covered"], "prevalence_covered": la["prevalence_covered"]})
    check("S joint-answer kappa (R1-06): 19/20 covered by both with one disagreement gives kappa_covered = 0 (prevalence artefact) but kappa_answer > 0.9, so the round stays informative; raw agreement and prevalence are reported",
          abs(la["kappa_covered"]) < 1e-9 and la["kappa_answer"] > 0.9 and inf["informative"] is True and abs(la["raw_agreement_covered"] - 0.95) < 1e-9 and abs(la["prevalence_covered"] - 0.975) < 1e-9
          and inf["kappa_covered"] == la["kappa_covered"], la)


def labels_check_checks(tmp):
    """R1-09, R3-02: the pre-score label gate (kappa_answer >= 0.6 on the GO population of the round's scored window units; the rule concentration is reported only) through analysis_x2d.py --labels-check R."""
    def mk(name, rule, area_b, n=20, core=()):
        data = os.path.join(tmp, "lc-" + name)
        wjl(os.path.join(data, "rounds", "1", "units.jsonl"), [{"case_id": "1:k%02d" % i, "set": "window", "control": False, "stratum_reason": "core_gate" if i in core else "one_area"} for i in range(n)]
            + [{"case_id": "1:ctl-1", "set": "ctl", "control": True}])
        VR.ensure_inputs(1, data)
        gv = x2d.guide_version(1, data)
        rows = []
        for i in range(n):
            for who, a in (("labeller_A:claude", "a%d" % (i % 5)), ("labeller_B:codex", area_b(i))):
                rows.append({"case_id": "1:k%02d" % i, "kind": "wide", "label_set": "main", "by": who, "covered": True, "acceptable": [a], "primary": a, "none_kind": None, "rule": rule(i), "guide_version": gv})
        rows += [{"case_id": "1:ctl-1", "kind": "wide", "label_set": "main", "by": w, "covered": True, "acceptable": ["zz"], "primary": "zz", "none_kind": None, "rule": "R1", "guide_version": gv} for w in ("labeller_A:claude", "labeller_B:codex")]
        rows += [{"case_id": "1:k00", "kind": "wide", "label_set": "main", "by": "labeller_A:claude", "covered": False, "acceptable": [], "primary": None, "none_kind": "gap", "rule": "R9", "guide_version": "x2d-association-v1/guide-stale00000000"}]
        wjl(os.path.join(data, "rounds", "1", "labels.jsonl"), rows)
        return data
    cli = lambda data: run([PY, os.path.join(S, "analysis_x2d.py"), "--labels-check", "1"], cwd=ROOT, env=clean_env(X2D_DATA=data))
    good = mk("good", lambda i: "R%d" % (1 + i % 5), lambda i: "a%d" % (i % 5))
    conc = mk("conc", lambda i: "R1", lambda i: "a%d" % (i % 5))
    lowk = mk("lowk", lambda i: "R%d" % (1 + i % 5), lambda i: "a%d" % ((i + 1) % 5) if i % 2 == 0 else "a%d" % (i % 5))
    mixed = mk("mixed", lambda i: "R1,R2" if i % 2 == 0 else "R1,R3", lambda i: "a%d" % (i % 5))  # R2-08: R1 is cited in EVERY session, in two different raw strings
    core = mk("core", lambda i: "R%d" % (1 + i % 5), lambda i: "a%d" % ((i + 1) % 5) if i >= 20 else "a%d" % (i % 5), n=28, core=set(range(20, 28)))  # R3-10: the 8 disagreeing units are core-gate units
    lg, lc, ll, lm, lk = A.labels_check(good, 1), A.labels_check(conc, 1), A.labels_check(lowk, 1), A.labels_check(mixed, 1), A.labels_check(core, 1)
    rg_, rc_, rl_, rm_ = cli(good), cli(conc), cli(lowk), cli(mixed)
    check("R labels-check passes on healthy labels (varied rules, joint-answer kappa >= 0.6); the control unit's labels and a row of a stale guide version are never counted", lg["ok"] is True and lg["n_units"] == 20 and rg_.returncode == 0, (lg, rg_.stderr[-200:]))
    lcj = os.path.join(good, "rounds", "1", "labels-check.json")
    check("R2-03 labels-check.json is written by the CLI and carries the round's guide version (a stale one is refused by run.sh's prereq)", os.path.exists(lcj) and json.load(open(lcj))["guide_version"] == x2d.guide_version(1, good) == lg["guide_version"] and json.load(open(lcj))["ok"] is True)
    check("R2-08 / R3-02 mixed citations (\"R1,R2\" / \"R1,R3\") with one dominant rule (R1 in every session) give ok TRUE: the concentration (top share > 0.40) is reported (`rule_top_share_over_0_40`), never a trigger; the raw strings are kept as they were written",
          lm["ok"] is True and lm["rule_top"] == "R1" and lm["rule_top_share"] > 0.40 and lm["rule_top_share_over_0_40"] is True and rm_.returncode == 0 and set(lm["rule_counts"]) == {"R1,R2", "R1,R3"} and lm["rule_sessions_parsed"] == 40,
          (lm["ok"], lm["rule_top"], lm["rule_top_share"], lm["rule_counts"], rm_.returncode))
    check("R3-02 sessions citing R1 in every session (the general rule) pass: ok true, share 1.0 reported, the CLI exits 0; the CLI exits 1 only for kappa_answer < 0.6 (rounds 1-7, no guide revision recorded)",
          lc["ok"] is True and lc["rule_top_share"] == 1.0 and lc["rule_top_share_over_0_40"] is True and rc_.returncode == 0 and ll["ok"] is False and rl_.returncode == 1 and "kappa_answer" in " ".join(ll["failed"]) and "rule-citation" not in " ".join(ll["failed"]),
          (lc["ok"], rc_.returncode, ll["failed"], rl_.returncode))
    wjl(os.path.join(lowk, "rounds", "1", "guide-revisions.jsonl"), [{"utc": "t", "guide_sha256": "x" * 64, "guide_version": x2d.guide_version(1, lowk), "inputs_sha256": "y" * 64, "reason": "r"}])
    rl2 = cli(lowk)
    check("R3-02 after the one recorded guide clarification a still-failing kappa lets the round proceed (CLI exit 0, labels-check.json ok false: the round is scored and non-informative)",
          rl2.returncode == 0 and json.load(open(os.path.join(lowk, "rounds", "1", "labels-check.json")))["ok"] is False and "non-informative" in rl2.stdout, (rl2.returncode, rl2.stdout[-200:]))
    check("R3-10 labels-check decides on the GO population: a kappa that fails only on core-gate units stays ok (kappa_answer over the 20 other units = 1.0) and the all-unit kappa is reported lower (kappa_answer_all_units); `informative` follows",
          lk["ok"] is True and lk["kappa_answer"] == 1.0 and lk["n_units"] == 20 and lk["n_units_all"] == 28 and lk["kappa_answer_all_units"] < 1.0, lk)
    # R3-07: the check is bound to the label rows it judged
    d0 = A.labels_digest(1, good)
    rows_g = jl(os.path.join(good, "rounds", "1", "labels.jsonl"))
    gvg = x2d.guide_version(1, good)
    stale_ctl = [{"case_id": "1:k05", "kind": "wide", "label_set": "main", "by": "labeller_A:claude", "covered": False, "acceptable": [], "primary": None, "none_kind": "gap", "rule": "R9", "guide_version": "x2d-association-v1/guide-stale00000000"},
                 {"case_id": "1:ctl-1", "kind": "wide", "label_set": "main", "by": "labeller_B:codex", "covered": False, "acceptable": [], "primary": None, "none_kind": "gap", "rule": "R9", "guide_version": gvg}]
    wjl(os.path.join(good, "rounds", "1", "labels.jsonl"), rows_g + stale_ctl)
    d_stale = A.labels_digest(1, good)
    wjl(os.path.join(good, "rounds", "1", "labels.jsonl"), rows_g + stale_ctl + [{"case_id": "1:k03", "kind": "wide", "label_set": "main", "by": "C adjudicated", "covered": True, "acceptable": ["a3"], "primary": "a3",
                                                                                  "none_kind": None, "rule": "R1", "guide_version": gvg, "agreement": "adjudicated", "round": "fill"}])
    check("R3-07 labels-check.json carries labels_sha256 = the digest of the current-guide wide rows of the checked units; an appended label row (a `label --new-round` fill) changes the recomputed digest, a row of a stale guide or of a control does not",
          lg["labels_sha256"] == d0 == json.load(open(lcj))["labels_sha256"] and d_stale == d0 and A.labels_digest(1, good) != d0, (lg["labels_sha256"], d0, d_stale, A.labels_digest(1, good)))


def write_verdict_data(tmp, name, pl, vsem, fan, fan_primary, core_gate_n=0):
    """A data directory holding only what `verdict` reads for rounds 8-10 (units, gold, arms, results.json) and rounds/0/gates.json (K5's primary switch, N1)."""
    data = os.path.join(tmp, "vd-" + name)
    by_round = {r: [u for u in pl["units"] if u["round"] == r] for r in x2d.FROZEN_ROUNDS}
    cg = {u["case_id"] for u in by_round[x2d.FROZEN_ROUNDS[0]][:core_gate_n]}
    for r, us in by_round.items():
        rd = os.path.join(data, "rounds", str(r))
        make_inputs(data, r)
        ids = {u["case_id"] for u in us}
        wjl(os.path.join(rd, "units.jsonl"), [dict(u, stratum=3 if u["case_id"] in cg else 1, stratum_reason="core_gate" if u["case_id"] in cg else "one_area") for u in us])
        wjl(os.path.join(rd, "gold.jsonl"), [pl["gold"][c] for c in sorted(ids)])
        for arm, rows in (("V-sem", vsem), ("B2", pl["arms"]["B2"]), ("fan-out", fan)):
            wjl(os.path.join(rd, "arms", arm + ".jsonl"), [dict(x, arm=arm) for x in rows if x["case_id"] in ids])
        json.dump({"informative": {"informative": True}, "arms": {}}, open(os.path.join(rd, "results.json"), "w"))
    os.makedirs(os.path.join(data, "rounds", "0"), exist_ok=True)
    json.dump({"gates": {"K6_N1": {"status": "PASS", "value": 0.0, "bound": 0.2}}, "fan_out_primary": fan_primary}, open(os.path.join(data, "rounds", "0", "gates.json"), "w"))
    return data


def verdict_primary_checks(tmp):
    """R1-03 / R1-14 / R1-19 through the real `verdict()`: V-sem and fan-out disagree and K5 flips -> the verdict, the pooled summary and per-round reports follow the primary arm; core-gate units leave the GO population."""
    structure = PS.empirical_structure()
    pl = PS.plant(structure, 0.99, 0.55, seed=21, corr=0.0)
    strong, weak = pl["arms"]["V-sem"], pl["arms"]["B2"]  # `weak` is B2 itself: a paired gain of exactly 0
    run_v = lambda name, **kw: A.verdict(write_verdict_data(tmp, name, pl, **kw), B=300)
    a = run_v("a", vsem=strong, fan=weak, fan_primary=False)
    b = run_v("b", vsem=strong, fan=weak, fan_primary=True)
    c = run_v("c", vsem=weak, fan=strong, fan_primary=True)
    d = run_v("d", vsem=weak, fan=strong, fan_primary=False)
    check("S K5 not flipped: the verdict is on V-sem (strong, gain >= +0.10, perception unmeasured -> INCONCLUSIVE); K5 flipped with a weak fan-out: the verdict follows the fan-out (gain 0 -> NO-GO) (R1-03)",
          a["primary_arm"] == "V-sem" and a["pooled"]["jev_arm"] == "V-sem" and a["pooled"]["gain_res"]["point"] >= 0.10 and a["verdict"] == "INCONCLUSIVE"
          and b["primary_arm"] == "fan-out" and b["pooled"]["jev_arm"] == "fan-out" and abs(b["pooled"]["gain_res"]["point"]) < 1e-12 and b["verdict"] == "NO-GO", (a["verdict"], b["verdict"], a["pooled"]["gain_res"]["point"], b["pooled"]["gain_res"]["point"]))
    check("S the mirror image (V-sem weak, fan-out strong): primary fan-out -> INCONCLUSIVE on its strong gain, primary V-sem -> NO-GO: the primary arm decides, V-sem never silently does (R1-03)",
          c["primary_arm"] == "fan-out" and c["verdict"] == "INCONCLUSIVE" and c["pooled"]["gain_res"]["point"] >= 0.10 and d["primary_arm"] == "V-sem" and d["verdict"] == "NO-GO", (c["verdict"], d["verdict"]))
    e = run_v("e", vsem=strong, fan=weak, fan_primary=False, core_gate_n=7)
    check("S core-gate units (stratum_reason core_gate) are excluded from the pooled GO population and counted: 7 of 91 clusters leave it (R1-14)", e["n_core_gate_excluded"] == 7 and e["pooled"]["n"] == 84 and a["pooled"]["n"] == 91, (e["n_core_gate_excluded"], e["pooled"]["n"]))


def n1_checks(tmp):
    """R1-18: N1 runs on D u P (>= 40 gold-covered units here), is measured per unit (a unit without an N1 arm row fails the gate), and keeps the bound max(2/K, 0.20)."""
    if not os.path.isdir(SRC):
        check("S N1 on D u P needs the fastapi clone X2D_SRC for the gates' core gate", False, SRC)
        return
    ids = ["0:d%02d" % i for i in range(20)] + ["0:p%02d" % i for i in range(20)]
    dummy = {"case_id": "0:zz", "arm": "x", "output": "none", "area_id": None}

    def mk(name, n1_ids, p2=()):
        data = os.path.join(tmp, "n1-" + name)
        d0 = os.path.join(data, "rounds", "0")
        wjl(os.path.join(d0, "units.jsonl"), [{"case_id": c, "set": "D" if ":d" in c else "P", "control": False} for c in ids])
        wjl(os.path.join(d0, "control-key.jsonl"), [{"case_id": "0:ctl", "kind": "none"}])
        wjl(os.path.join(d0, "gold.jsonl"), [{"case_id": c, "covered": True, "acceptable": ["a1"], "primary": "a1", "none_kind": None} for c in ids])
        wjl(os.path.join(d0, "labels.jsonl"), [{"case_id": ids[0], "kind": "wide", "label_set": "main", "by": "labeller_A:x", "covered": True, "acceptable": ["a1"], "primary": "a1", "rule": "R1"}])
        json.dump({"k2_share": 0.5, "k3_share": 0.3, "real_none_proxy": 7, "frozen_pr_counts": {"8": {"p1": 3, "p2": 4}, "9": {"p3": 5}, "10": {"p4": 2}}}, open(os.path.join(d0, "k23.json"), "w"))
        make_inputs(data, 0)
        for p in ("jev/scores.jsonl", "jev-n1/scores.jsonl", "jev-p/scores.p1.jsonl", "jev-p/scores.p2.jsonl"):
            wjl(os.path.join(d0, p), [{"type": "header"}])
        for p in ("arms/V-sem.jsonl", "arms/B1.jsonl", "arms-p1/V-sem.jsonl", "arms-p2/V-sem.jsonl"):
            wjl(os.path.join(d0, p), [dummy])
        b2 = [{"case_id": c, "arm": "B2", "output": "none", "area_id": None} for c in ids]
        wjl(os.path.join(d0, "arms", "B2.jsonl"), b2)
        wjl(os.path.join(d0, "arms-n1", "B2.jsonl"), b2)
        wjl(os.path.join(d0, "arms-n1", "V-sem.jsonl"), [{"case_id": c, "arm": "V-sem", "output": "assigned", "area_id": "zz", "confidence": 0.9, "unresolved_reason": None} for c in n1_ids])
        json.dump({"areas": [{"area_id": "a%d" % i} for i in range(40)]}, open(os.path.join(d0, "catalogue-n1.json"), "w"))
        os.makedirs(os.path.join(data, "catalogue"), exist_ok=True)
        json.dump({"schema": "x2d-catalogue-v1", "areas": [], "heuristic_only": []}, open(os.path.join(data, "catalogue", "catalogue.json"), "w"))
        if p2:  # R2-05: P2 clones (case id, derived gold covered?, arm output) with their control-key rows
            wjl(os.path.join(d0, "control-key.jsonl"), [{"case_id": "0:ctl", "kind": "none"}] + [{"case_id": c, "kind": "p2", "source_case_id": ids[0], "ablated_area_ids": ["a1"]} for c, _cov, _out in p2])
            wjl(os.path.join(d0, "gold.jsonl"), jl(os.path.join(d0, "gold.jsonl")) + [({"case_id": c, "covered": True, "acceptable": ["a2"], "primary": "a2", "none_kind": None} if cov else
                                                                                     {"case_id": c, "covered": False, "acceptable": [], "primary": None, "none_kind": "gap"}) for c, cov, _out in p2 if cov is not None])
            wjl(os.path.join(d0, "arms-p2", "V-sem.jsonl"), [{"case_id": c, "arm": "V-sem", "output": out, "area_id": "a2" if out == "assigned" else None} for c, _cov, out in p2])
        return A.gates(data, repo=SRC)
    gfull, gpart = mk("full", ids), mk("partial", ids[:20])
    full, partial = gfull["gates"]["K6_N1"], gpart["gates"]["K6_N1"]
    check("S N1 runs on D u P (R1-18): all 40 gold-covered units are measured, none correct (at chance) -> PASS, bound = max(2/K, 0.20) = 0.20 for K = 40 areas",
          full["status"] == "PASS" and full["n"] == 40 and abs(full["bound"] - 0.20) < 1e-12 and full["value"]["units_without_arm_row"] == 0, full)
    check("S N1 fails when D u P units lack an N1 arm row (the P half unmeasured): the gate does not pass on a subset", partial["status"] == "FAIL" and partial["value"]["units_without_arm_row"] == 20, partial)
    rnp = gfull["gates"]["real_none_proxy"]
    check("S R2-11 gates() records real_none_proxy as an INFO row (never a stop): value from k23.json, n = the frozen-round scored clusters, no bound (R4-13: ~10 real positives cannot veto; the proxy only limits the descriptive new-area table); PASS/FAIL gates are unaffected",
          rnp["status"] == "INFO" and rnp["value"] == 7 and rnp["n"] == 14 and rnp["bound"] is None, rnp)
    gap10 = [("0:ctl-%d" % i, False, "none") for i in range(10)] + [("0:ctl-r1", True, "assigned"), ("0:ctl-r2", True, "none")]
    g_ok = mk("p2ok", ids, gap10)["gates"]["P2"]
    g_few = mk("p2few", ids, gap10[:7] + gap10[10:])["gates"]["P2"]
    check("S R2-05 the P2 gate is evaluated over the gap positives only (10 clones whose derived gold is none / gap, all answered none -> PASS, n = 10); clones whose derived gold stays covered are reported in retained_covered (2 clones, 1 answered with the retained area) and are in no denominator",
          g_ok["status"] == "PASS" and g_ok["n"] == 10 and g_ok["value"]["gap_positives"] == 10 and g_ok["value"]["retained_covered"] == {"n": 2, "assigned_correct": 1} and g_ok["value"]["p2_clones"] == 12 and g_ok["value"]["none_rate"] == 1.0, g_ok)
    check("S R3-04 / R2-05 fewer than 8 label-derived gap positives (7) make the P2 gate UNMEASURED (not FAIL: a thin denominator is no Jev result), reported with the counts; the round-0 critique must rule on it",
          g_few["status"] == "UNMEASURED" and g_few["n"] == 7 and g_few["value"]["gap_positives"] == 7 and g_few["value"]["none_rate"] == 1.0, g_few)
    p2_thin = lambda name, k_assigned: mk(name, ids, [(c, cov, "assigned" if i < k_assigned else out) for i, (c, cov, out) in enumerate(gap10[:7])] + gap10[10:])["gates"]["P2"]
    g_zero, g_six, g_five = p2_thin("p2zero", 7), p2_thin("p2six", 1), p2_thin("p2five", 2)
    check("S R4-04 below 8 gap positives a decisive failure is never hidden: 0 of 7 answered `none` -> FAIL (n = 7, none rate 0.0), 5 of 7 (2 not answered `none`: no denominator of 8 could reach 0.80) -> FAIL, 6 of 7 -> UNMEASURED (none rate 6/7, counts reported), 7 of 7 -> UNMEASURED (the R3-04 case above)",
          g_zero["status"] == "FAIL" and g_zero["n"] == 7 and g_zero["value"]["none_rate"] == 0.0 and g_five["status"] == "FAIL" and g_five["n"] == 7 and abs(g_five["value"]["none_rate"] - 5 / 7) < 1e-12
          and g_six["status"] == "UNMEASURED" and g_six["n"] == 7 and abs(g_six["value"]["none_rate"] - 6 / 7) < 1e-12 and g_six["value"]["gap_positives"] == 7 and g_few["status"] == "UNMEASURED", (g_zero, g_five, g_six))
    g_slack = mk("p2slack", ids, gap10[:8] + [("0:ctl-r1", True, "assigned"), ("0:ctl-nog", None, "none")])["gates"]["P2"]
    g_fail = mk("p2fail", ids, [(c, cov, "assigned" if i < 3 else out) for i, (c, cov, out) in enumerate(gap10[:10])])["gates"]["P2"]
    check("S R3-04 one clone whose source gold keeps a retained acceptable area (counted in retained_covered, answered correctly) and one clone whose source has no agreed gold (no_agreed_gold) do not make the gate fail: it is decided over the 8 gap positives (PASS, n = 8); "
          "the gate FAILs only when the none rate over >= 8 gap positives is < 0.80 (3 of 10 answered with an area: 0.70)",
          g_slack["status"] == "PASS" and g_slack["n"] == 8 and g_slack["value"]["retained_covered"] == {"n": 1, "assigned_correct": 1} and g_slack["value"]["no_agreed_gold"] == 1 and g_slack["value"]["p2_clones"] == 10
          and g_fail["status"] == "FAIL" and g_fail["n"] == 10 and abs(g_fail["value"]["none_rate"] - 0.7) < 1e-12, (g_slack, g_fail))


def futility_checks():
    def rnd(r, gain_rows, pa_rows, informative=True, pause=False, abst=0):
        return {"round": r, "informative": informative, "integrity_pause": pause,
                "pooling": [{"pr_key": "r%dp%d" % (r, i), "n_m": 2, "n_res": 2 - abst, "d_res": d, "n_all": 2, "d_all": d - abst, "a": 2, "ac": ac, "gc": 2, "gc_res": 2 - abst, "abst": abst} for i, (d, ac) in enumerate(zip(gain_rows, pa_rows))]}
    poor = lambda r, **kw: rnd(r, [0, 0, 0, -1, 0, 0], [1, 1, 0, 1, 1, 1], **kw)   # gain 0 or below, P_assign 5/6 = 0.83
    good = lambda r, **kw: rnd(r, [1, 1, 1, 1, 0, 1], [2, 2, 2, 2, 2, 2], **kw)   # gain 0.42, P_assign 1.0
    f = lambda rs: A.futility(rs, B=500)["fired"]
    check("S futility rule: three consecutive informative rounds with pooled gain 80 % upper bound <= +0.02 and pooled P_assign 80 % upper bound below 0.90 fire it", f([poor(1), poor(2), poor(3)]))
    check("S futility rule: it does not fire on two such rounds, nor when one of three has a real gain, nor when P_assign is high", not f([poor(1), poor(2)]) and not f([poor(1), good(2), poor(3)]) and not f([good(1), good(2), good(3)]))
    check("S futility rule: a non-informative round breaks the run (no firing across it) and an integrity-pause round is neither counted nor breaking",
          not f([poor(1), poor(2, informative=False), poor(3), poor(4)][:3]) and f([poor(1), poor(2, pause=True), poor(3), poor(4)]) and not f([poor(1), poor(2, pause=True), poor(3)]))
    fw = A.futility([poor(1), poor(2), poor(3)], B=500)["windows"][0]
    check("S futility R4-13 (5): the gain term is gain_res (resolved clusters); the window reports the unresolved rate and the all-cluster gain beside it", fw["fires"] and "gain_res_ub80" in fw and fw["unresolved_rate"] == 0.0 and fw["gain_all_point"] is not None, fw)
    hi_u = lambda r: rnd(r, [1, 1, 1, 1, 0, 1], [2, 2, 2, 2, 2, 2], abst=1)  # real resolved gain, half of every PR's clusters abstained
    check("S futility R4-13 (5): no disjunct on the unresolved rate -- three rounds with a real resolved gain and u = 0.5 do not fire it (and the rate is reported), while the same rounds with no resolved gain do", not A.futility([hi_u(1), hi_u(2), hi_u(3)], B=500)["fired"]
          and abs(A.futility([hi_u(1), hi_u(2), hi_u(3)], B=500)["windows"][0]["unresolved_rate"] - 0.5) < 1e-12 and A.futility([poor(1), poor(2), poor(3)], B=500)["fired"])
    for info, ok in (({"n_clusters": 20, "n_prs": 6, "kappa_covered": 0.6, "kappa_answer": 0.6}, True), ({"n_clusters": 19, "n_prs": 6, "kappa_covered": 0.9, "kappa_answer": 0.9}, False),
                     ({"n_clusters": 20, "n_prs": 5, "kappa_covered": 0.9, "kappa_answer": 0.9}, False), ({"n_clusters": 25, "n_prs": 8, "kappa_covered": 0.0, "kappa_answer": 0.9}, True),
                     ({"n_clusters": 25, "n_prs": 8, "kappa_covered": 0.9, "kappa_answer": 0.59}, False), ({"n_clusters": 25, "n_prs": 8, "kappa_covered": 0.9, "kappa_answer": None}, False)):
        check("S informative round (>= 20 clusters, >= 6 PRs, kappa >= 0.6 on the JOINT ANSWER only; kappa_covered is reported, decides nothing, R1-06): %s -> %s" % (info, ok), A.informative(info)["informative"] is ok)


def p2_checks():
    g = {"case_id": "u", "covered": True, "acceptable": ["A", "B"], "primary": "A", "none_kind": None}
    d1 = G.derive_ablation(g, "A")
    d2 = G.derive_ablation(dict(g, acceptable=["A"], primary="A"), "A")
    d3 = G.derive_ablation(g, "Z")
    check("S P2 catalogue ablation: gold pointing only to the deleted area derives to none/gap automatically; with another acceptable area it keeps that one (primary cleared when it was the ablated area); an unrelated ablation changes nothing",
          d2["covered"] is False and d2["none_kind"] == "gap" and d2["acceptable"] == [] and d2["primary"] is None and d1["covered"] is True and d1["acceptable"] == ["B"] and d1["primary"] != "A" and d3 == g)
    arm_none = {"case_id": "u", "arm": "V-sem", "output": "none", "area_id": None}
    so = A.score_outcome(arm_none, d2)
    check("S P2 flows into scoring: with the area ablated, Jev's `none` on the unit is correct and reported as a catalogue gap (none_correct_gap); an assignment to any remaining area is wrong", so["kind"] == "none_correct_gap" and A.score_outcome(dict(arm_none, output="assigned", area_id="B"), d2)["kind"] == "assigned_gold_none")


def _bar_met(cells, bar, low):
    """R5-02: a binding power cell set meets its bar iff every cell is on the right side of it (>= for a P3 cell, <= for a null cell)."""
    return all((c["share"] >= bar) if low else (c["share"] <= bar) for c in cells)


def _power_main_exit(tmp, checks):
    """Exit code of power_sim.main with run() stubbed to return `checks` (the rest of the checks true): 3 when a null check is false, else 0 (R5-02 (2))."""
    full = {k: True for k in x2d.POWER_CHECK_KEYS}
    full.update(checks)
    cells = [{"p_j": 0.9, "u": 0.0, "share": 0.0, "wilson95": [0.0, 0.0], "pass": True}]
    fake = {k: cells for k in ("P3_u0", "P3_u05", "P3_u10", "N3", "N3_PR", "N4a", "N4b", "N4b_cap", "N4b_rho0")}
    fake.update({"P3_u0": cells[0], "P3_u05": cells[0], "P3_u10": cells[0], "checks": full, "n4_max": 0.0, "seconds": 0.0, "N4c": dict(cells[0], verdicts={}), "P3_PR": dict(cells[0], verdicts={}), "P3_rho0": dict(cells[0], verdicts={}),
                 "P3_curve_delta_0.15": []})
    for k in ("P3_u0", "P3_u05", "P3_u10"):
        fake[k] = dict(cells[0], u=0.0)
    orig_run, orig_write, orig_pins, orig_refuse = PS.run, PS.x2d.write_json_new, PS.x2d.verify_pins, PS.x2d.refuse_if_exists
    try:
        PS.run = lambda *a_, **k_: fake
        PS.x2d.write_json_new = lambda *a_, **k_: None
        PS.x2d.verify_pins = lambda *a_, **k_: []
        PS.x2d.refuse_if_exists = lambda *a_, **k_: None
        import contextlib
        import io
        with contextlib.redirect_stdout(io.StringIO()):
            return PS.main(["--out", os.path.join(tmp, "never-written-power.json")])
    finally:
        PS.run, PS.x2d.write_json_new, PS.x2d.verify_pins, PS.x2d.refuse_if_exists = orig_run, orig_write, orig_pins, orig_refuse


def power_checks(tmp):
    structure = PS.empirical_structure()
    check("S power structure: the empirical structure of rounds 8-10 is 91 scored clusters in 40 PRs (round 10 holds the former spare PRs; no spare remains), none above the cap of 10", sum(s["n"] for s in structure) == 91 and len(structure) == 40 and max(s["n"] for s in structure) <= 10)
    n2 = PS.n2_permutation(structure, 0.9, 0.7, reps=100)
    check("S N2 label permutation (gold labels permuted within each round, no fixed points): the mean paired gain is centred on 0 (within 3 standard errors and 0.02) and the mean P_assign is at chance (<= 0.06 for 40 areas)",
          abs(n2["mean_gain"]) < max(0.02, 3 * n2["gain_se_of_mean"]) and n2["mean_P_assign"] <= 0.06, n2)
    res = PS.run(pilots=1000, seed=x2d.SEED, boot_B=1000)
    json.dump(res, open(os.path.join(tmp, "power.json"), "w"))
    for name in ("P3_u0", "P3_u05", "P3_u10"):
        c = res[name]
        print("INFO power %s planted effect p_B2 = 0.80, p_J = 0.95, u = %.2f, %d positives inside, 91 clusters / 40 PRs: GO %d of %d = %.3f (Wilson 95 %% [%.3f, %.3f]; GO / INCONCLUSIVE / NO-GO %s; bar >= %.2f); %d bootstrap runs" % (name, c["u"], c["n_pos"], c["go"], c["pilots"], c["share"], c["wilson95"][0], c["wilson95"][1], json.dumps(c["verdicts"]), PS.BARS[name], c["bootstrap_run"]))
    for key in ("N3", "N3_PR", "N4a", "N4b", "N4b_cap", "N4b_rho0"):
        for cell in res[key]:
            print("INFO power %s p_J = p_B2 = %.2f u = %.2f: GO share %.3f (%d of %d; GO / INCONCLUSIVE / NO-GO %s; bar <= %.2f)" % (key, cell["p_j"], cell["u"], cell["share"], cell["go"], cell["pilots"], json.dumps(cell["verdicts"]), PS.BARS[key]))
    for key in ("N4c", "P3_PR", "P3_rho0"):
        print("INFO report-only %s: GO share %.3f [%.3f, %.3f]; GO / INCONCLUSIVE / NO-GO %s" % (key, res[key]["share"], res[key]["wilson95"][0], res[key]["wilson95"][1], json.dumps(res[key]["verdicts"])))
    for cell in res["P3_curve_delta_0.15"]:
        print("INFO report-only delta = 0.15 curve p_B2 = %.2f, p_J = %.2f: GO share %.3f" % (cell["p_b2"], cell["p_j"], cell["share"]))
    th = json.load(open(os.path.join(S, "thresholds.json")))
    check("S the power report is schema x2d-power-v5 with its checks dict (P3_u0, P3_u05, P3_u10, N3, N3_PR, N4a, N4b, N4b_cap, N4b_rho0), the bars of thresholds.json study.power, every report-only cell (P3-PR, P3 at rho 0, the delta = 0.15 curve, N4c), the live-structure lines p3_real_structure / p3_abstention / n4_max, and the structure and the 10 positives inside",
          res["schema"] == "x2d-power-v5" and set(res["checks"]) == set(x2d.POWER_CHECK_KEYS) == {"P3_u0", "P3_u05", "P3_u10", "N3", "N3_PR", "N4a", "N4b", "N4b_cap", "N4b_rho0"} and res["bars"] == th["study"]["power"]["bars"] == PS.BARS and res["p3_name"] == PS.P3_NAME
          and [c["p_b2"] for c in res["P3_curve_delta_0.15"]] == PS.P3_CURVE and "P3_PR" in res and "P3_rho0" in res and "N4c" in res and res["structure"]["clusters"] == 91 and res["structure"]["prs"] == 40 and res["n_pos"] == 10
          and res["p3_real_structure"]["share"] == res["P3_u0"]["share"] and res["p3_abstention"] == {"0.05": res["P3_u05"]["share"], "0.10": res["P3_u10"]["share"]} and res["n4_max"] == max(c["share"] for c in res["N4a"] + res["N4b"] + res["N4b_cap"] + res["N4b_rho0"]),
          (res["schema"], list(res["checks"])))
    check("S the power cells cover the preregistered grids: N3 / N3-PR p in {0.5 .. 0.95} x u in {0, 0.10}, N4a p in {0.80, 0.85, 0.90, 0.95}, N4b p in {0.85, 0.88, 0.90, 0.95}, N4b-cap p in {0.80, 0.85} (rho 0.5, own errors truncated at the cap), N4b-rho0 p in {0.85, 0.88, 0.90} (rho 0), N4c (0.90, 0.80, own-error abstention), PR latent on the PR cells",
          [(c["p_j"], c["u"]) for c in res["N3"]] == [(p, u) for p in PS.N3["p_b2"] for u in PS.N3["u"]] and [(c["p_j"], c["u"]) for c in res["N3_PR"]] == [(p, u) for p in PS.N3["p_b2"] for u in PS.N3["u"]]
          and all(c["pr_w"] == PS.PR_W for c in res["N3_PR"]) and all(c["corr"] == 0.0 for c in res["N3"] + res["N3_PR"]) and [c["p_j"] for c in res["N4a"]] == PS.N4A_P and [c["p_j"] for c in res["N4b"]] == PS.N4B_P
          and [c["p_j"] for c in res["N4b_cap"]] == [0.80, 0.85] == PS.N4B_CAP_P and [c["p_j"] for c in res["N4b_rho0"]] == [0.85, 0.88, 0.90] == PS.N4B_RHO0_P
          and all(c["mode"] == "own_errors_cap" and c["corr"] == PS.CORR for c in res["N4b_cap"]) and all(c["mode"] == "own_errors" and c["corr"] == 0.0 for c in res["N4b_rho0"])
          and all(c["mode"] == "b2_errors" for c in res["N4a"]) and all(c["mode"] == "own_errors" for c in res["N4b"]) and (res["N4c"]["p_j"], res["N4c"]["p_b2"]) == (0.90, 0.80) and res["N4c"]["mode"] == "own_errors")
    cell_specs = (("P3_u0", [res["P3_u0"]], True), ("P3_u05", [res["P3_u05"]], True), ("P3_u10", [res["P3_u10"]], True), ("N3", res["N3"], False), ("N3_PR", res["N3_PR"], False), ("N4a", res["N4a"], False),
                  ("N4b", res["N4b"], False), ("N4b_cap", res["N4b_cap"], False), ("N4b_rho0", res["N4b_rho0"], False))
    synth = [dict(res["N4b"][0], share=0.12, **{"pass": False})]
    check("S R5-02 the bar evaluation treats a null cell at 0.12 against the bar 0.10 as NOT met (a missed bar is a failed check, never a pass), and power_sim's exit code is 3 for a null check false, 0 for a P3 check false",
          not _bar_met(synth, PS.BARS["N4b"], False) and _bar_met([dict(synth[0], share=0.10)], PS.BARS["N4b"], False) and not _bar_met([dict(synth[0], share=0.79)], PS.BARS["P3_u0"], True)
          and set(PS.NULL_CHECKS) == set(x2d.POWER_NULL_KEYS) == {"N3", "N3_PR", "N4a", "N4b", "N4b_cap", "N4b_rho0"} and _power_main_exit(tmp, {"N4b": False}) == 3 and _power_main_exit(tmp, {"P3_u10": False}) == 0 and _power_main_exit(tmp, {}) == 0)
    for name, cells, low in cell_specs:
        bar = PS.BARS[name]
        met = _bar_met(cells, bar, low)
        worst = min(c["share"] for c in cells) if low else max(c["share"] for c in cells)
        consistent = res["checks"][name] is met and all(c["pass"] is ((c["share"] >= bar) if low else (c["share"] <= bar)) for c in cells)
        check("S binding power bar %s: %s GO share %.3f %s the preregistered bar %.2f (every cell; a cell on the wrong side of its bar FAILS validation, R5-02)" % (name, "lowest" if low else "highest", worst, ">=" if low else "<=", bar), met and consistent, [c["share"] for c in cells])
    check("S the power simulation's overall pass flag is the conjunction of its nine binding checks", res["all_pass"] == all(res["checks"].values()))
    again = PS.simulate(0.95, 0.80, structure, pilots=200, boot_B=300, seed=x2d.SEED, u=0.05)
    again2 = PS.simulate(0.95, 0.80, structure, pilots=200, boot_B=300, seed=x2d.SEED, u=0.05)
    check("S the power simulation is deterministic for a seed", again == again2)
    check("S the abstention modes behave as specified: random u = 0.10 abstains ~10 % of clusters, N4a truncates at int(0.15 x 91) = 13 clusters, N4b abstains exactly on Jev's own errors (the resolved clusters are all Jev-correct: P_assign = 1), N4b-cap on at most int(0.15 x 91) = 13 of them",
          abs(PS.simulate(0.9, 0.8, structure, pilots=100, boot_B=50, u=0.10)["u_mean"] - 0.10) < 0.03 and abs(PS.simulate(0.60, 0.60, structure, pilots=50, boot_B=50, mode="b2_errors")["u_mean"] - 13 / 91) < 1e-9
          and PS.simulate(0.80, 0.80, structure, pilots=50, boot_B=50, mode="own_errors")["u_mean"] > 0.1 and abs(PS.simulate(0.60, 0.60, structure, pilots=50, boot_B=50, mode="own_errors_cap")["u_mean"] - 13 / 91) < 1e-9)


# ---------------------------------------------------------------- R round machinery: units, baselines, gold/membership, tuning, composition, versioning
import build_units as BU  # noqa: E402
import baselines as BL  # noqa: E402
import b2 as B2  # noqa: E402
import compose as CM  # noqa: E402
import tune as TU  # noqa: E402
import versioning as VR  # noqa: E402


def units_checks(tmp, repo, shas):
    c0, c1, c2, c3 = shas
    mem = {"round": 1, "areas": {"A": ["fastapi/a.py", "docs_src/topic/tutorial001.py", "fastapi/b.py"], "B": ["tests/test_tutorial/test_topic/test_tutorial001.py", "docs_src/known/tutorial001.py"]},
           "pool": [], "heuristic_only": ["docs_src/newtopic/tutorial001.py", "fastapi/zzz.py"]}
    info = {"after_commit": c0, "through_commit": c3, "prs": []}
    res = BU.build_round_units(repo.path, 1, mem, rounds_info=info, settings=dict(x2d.DEFAULT_SETTINGS), core_gate=False, check_prepass=False)
    units, aud, capped = res["units"], res["audit_units"], res["capped"]
    byk = {(u["dir"], u["stem"]): u for u in units}
    big = [u for u in units if u["pr_key"] == "9268"]
    check("R build_units on a synthetic history: scored window units are the residual clusters of the round's PRs (2 in #101, 3 in #102, 10 of the 120 in #9268; the still-unowned MODIFIED fastapi/c.py is not a unit, R1-04); the PR cap reports the other 110 as capped, never labelled",
          len(units) == 15 and len(big) == 10 and len(capped) == 110 and all(c["reason"] == "pr_cap" for c in capped) and not ({c["cluster_id"] for c in capped} & {u["cluster_id"] for u in units}), (len(units), len(big), len(capped)))
    keep = sorted((F.cluster_id(c3, {"dir": u["dir"], "stem": u["stem"]}) for u in big))
    allc = sorted((cc["cluster_id"] for cc in capped)) + keep
    check("R the capped selection is the cap hash order: the 10 scored clusters of #9268 have the 10 smallest sha256('x2d-cap' + id) of its 120 clusters", sorted(allc, key=x2d.cap_key)[:10] == sorted(keep, key=x2d.cap_key) and len(allc) == 120)
    u_d, u_new, u_util, u_other, u_util2 = byk[("fastapi", "d")], byk[("docs_src/newtopic", "tutorial001")], byk[("docs_src", "util")], byk[("docs_src/other", "tutorial001")], byk[("docs_src", "util2")]
    check("R ownership-aware residual (R1-04): a modified file no area owns is NOT a unit (fastapi/c.py: renamed in #101, modified in #102, counted modified_unowned), a variant of an UNOWNED example is residual, owned variants and owned modified files are not clusters",
          ("fastapi", "c") not in byk and res["counts"].get("modified_unowned") == 1 and ("docs_src/topic", "tutorial001") not in byk and ("fastapi", "a") not in byk and u_other["members"] == ["docs_src/other/tutorial001_an.json"], (sorted(byk), res["counts"]))
    check("R catalogue strata (code side): owned siblings in one area -> 1 (a new fastapi/d.py next to A's a.py), none -> 3 (new topic, cross-scope move, a new root file; the heuristic-only package keeps reason no_owned_siblings or heuristic_only); the rule-resolved files never reach Jev",
          u_d["stratum"] == 1 and u_d["stratum_reason"] == "one_area" and u_new["stratum"] == 3 and u_util["stratum"] == 3 and u_other["stratum"] == 3 and u_util2["stratum"] == 3, (u_d["stratum"], u_new["stratum"], u_util["stratum"]))
    check("R the unit state: change kind, role, excerpt of the first non-empty member within settings.excerpt_lines (60), PR title and body <= 1500 chars with changed paths and statuses, a diff summary with files and line counts and <= 8 symbols; the two-member cluster lists both paths sorted",
          set(u_new["state"]) == {"unit", "pr", "diff_summary"} and u_new["state"]["unit"]["paths"] == ["docs_src/newtopic/tutorial001.py", "docs_src/newtopic/tutorial001_an.py"] and u_new["state"]["unit"]["change_kind"] == "added"
          and len(u_new["state"]["unit"]["excerpt"].splitlines()) <= 60 and u_new["state"]["pr"]["title"] == "Add things (#101)" and len(u_new["state"]["pr"]["body"]) <= 1500
          and u_new["state"]["diff_summary"]["files"] >= 2 and len(u_new["state"]["diff_summary"]["top_symbols"]) <= 8 and u_util["state"]["unit"]["change_kind"].startswith("renamed_from fastapi/util.py"), u_util["state"]["unit"])
    check("R the labeller evidence holds the full text of every member (<= 300 lines) and the PR's changed paths; the unit row carries both views and the case id <round>:<cluster id>",
          [m["path"] for m in u_new["label_evidence"]["unit"]["members"]] == u_new["members"] and all(m["text"] for m in u_new["label_evidence"]["unit"]["members"])
          and u_new["case_id"] == "1:" + u_new["cluster_id"] and u_new["pr"] == 101 and u_new["pr_key"] == "101" and u_new["set"] == "window" and u_new["control"] is False)
    check("R audit: the (c)/(d)-resolved clusters (renamed fastapi/c.py, the two topic variants, the known variant) are offered as audit units with ids <round>:aud-<n> (set audit), at most 10 per round, and are listed in the lookups with the inherited area",
          len(aud) == 3 and sorted(a["case_id"] for a in aud) == ["1:aud-1", "1:aud-2", "1:aud-3"] and all(a["set"] == "audit" for a in aud)
          and sum(len(l["audit"]) for l in res["lookups"]) == 3 and {i["inherited_area_id"] for l in res["lookups"] for i in l["audit"]} == {"A", "B"}, [a["case_id"] for a in aud])
    check("R lookups count the rules per PR (modified_owned, modified_unowned as a coverage figure, variant_inherit, renamed_inherit, renamed_cross_scope, excluded markdown/translation/media, inert clusters dropped by rule (g))",
          res["counts"].get("variant_inherit", 0) >= 3 and res["counts"].get("renamed_inherit", 0) >= 1 and res["counts"].get("modified_owned", 0) >= 2 and res["counts"].get("excluded:markdown", 0) == 1
          and res["counts"].get("modified_unowned", 0) == 1 and res["counts"].get("modified_unowned:never_owned", 0) == 1
          and res["counts"].get("excluded:translation", 0) == 1 and res["counts"].get("excluded:media", 0) == 1 and res["counts"].get("inert_clusters", 0) >= 2, res["counts"])
    ops = {(i["op"], i["path"]) for i in res["inherit"]}
    check("R rule effects on membership are recorded (deleted file leaves, (c) rename moves with its owners, (d) variants join the owner's area) so gold can replay them without the repo",
          ("remove", "docs_src/app_testing/app_b/main.py") in ops and ("move", "fastapi/c.py") in ops and ("add", "docs_src/topic/tutorial001_an.py") in ops, sorted(ops)[:8])
    for u in units + aud:
        for k in ("arm_output", "arm", "gold", "label", "covered"):
            assert k not in u
    check("R unit rows carry no arm output and no gold (labels and arms are produced later from the unit rows)", True)
    # ---- baselines
    tree1 = set(F.tree_paths(repo.path, c1))
    own = F.Ownership(mem)
    r0 = BL.b0(["fastapi/c.py"], tree1, own)
    check("R B0: the majority owner of owned siblings in the same normalised directory (fastapi/c.py -> A); none when no owned sibling; unresolved on a tie",
          r0[0] == "assigned" and r0[1] == "A" and BL.b0(u_other["members"], tree1, own)[0] == "none"
          and BL.b0(["fastapi/q.py"], tree1 | {"fastapi/a.py", "fastapi/z.py"}, F.Ownership({"areas": {"A": ["fastapi/a.py"], "B": ["fastapi/z.py"]}}))[0] == "unresolved", (r0, BL.b0(["fastapi/q.py"], tree1 | {"fastapi/z.py"}, F.Ownership({"areas": {"A": ["fastapi/a.py"], "B": ["fastapi/z.py"]}}))))
    mem2 = {"areas": {"A": ["docs_src/topic2/tutorial001.py", "tests/test_tutorial/test_deep/x/y/test_own.py"], "B": ["fastapi/deep/x.py"]}, "pool": [], "heuristic_only": []}
    own2 = F.Ownership(mem2)
    tree2 = {"docs_src/topic2/tutorial001.py", "tests/test_tutorial/test_topic2/test_tutorial001.py", "fastapi/deep/x.py", "fastapi/deep/y/z/w.py"}
    tm = ["tests/test_tutorial/test_topic2/test_tutorial002.py"]
    parent_unit = ["fastapi/deep/y/new.py"]
    check("R B1 walks at most 2 parent levels to the first directory with owned files (fastapi/deep/y/new.py -> fastapi/deep -> B, but not 3 levels up); B2 adds the tutorial test-mirror rule written before any label (tests/test_tutorial/test_X <-> docs_src/X) that B0 and B1 lack",
          BL.b1(parent_unit, tree2, own2)[1] == "B" and BL.b1(["fastapi/deep/y/z/q/new.py"], tree2, own2)[0] == "none" and BL.b1(tm, tree2, own2)[0] == "none" and BL.b2(tm, tree2, own2)[:2] == ("assigned", "A")
          and BL.b2(["docs_src/topic2/tutorial002_an.py"], tree2 | {"tests/test_tutorial/test_topic2/test_tutorial001.py"}, F.Ownership({"areas": {"T": ["tests/test_tutorial/test_topic2/test_tutorial001.py"]}}))[1] == "T"
          and B2.b2_rule(["fastapi/x/new.py"], tree2, own2) is None, (BL.b1(parent_unit, tree2, own2), BL.b2(tm, tree2, own2)))
    check("R baseline arm rows follow the contract (case_id, arm, output assigned|none|unresolved, area_id, confidence null, unresolved_reason)", BL.arm_row("c", "B2", BL.b2(tm, tree2, own2)) == {"case_id": "c", "arm": "B2", "output": "assigned", "area_id": "A", "confidence": None, "unresolved_reason": None})
    return res, mem


def gold_membership_checks(tmp, res, mem, repo):
    data = os.path.join(tmp, "gm-data")
    rd = os.path.join(data, "rounds", "1")
    units = res["units"]
    wjl(os.path.join(rd, "units.jsonl"), units + res["audit_units"])
    wjl(os.path.join(rd, "inherit.jsonl"), res["inherit"])
    json.dump(mem, open(os.path.join(rd, "membership.json"), "w"))
    byk = {(u["dir"], u["stem"]): u for u in units}
    golds = []
    for u in units + res["audit_units"]:
        g = {"case_id": u["case_id"], "covered": True, "acceptable": ["B"], "primary": "B", "none_kind": None, "agreement": "agreed", "control_kind": None}
        if u["dir"] == "docs_src/newtopic":
            g.update(covered=False, acceptable=[], primary=None, none_kind="new")
        if u["dir"] == "docs_src" and u["stem"] == "util":
            g.update(acceptable=["A", "B"], primary=None)
        golds.append(g)
    wjl(os.path.join(rd, "gold.jsonl"), golds)
    nm = G.next_membership(1, data=data)
    check("R teacher forcing at round granularity: the next membership absorbs gold primaries (the cluster's files join the area), gold none goes to the pool, a covered cluster with several acceptable areas and no primary is not absorbed (listed)",
          set(byk[("fastapi", "d")]["members"]) <= set(nm["areas"]["B"]) and set(byk[("docs_src/newtopic", "tutorial001")]["members"]) <= set(nm["pool"]) and "docs_src/util.py" in nm["unabsorbed"], (nm["pool"][:3], nm["unabsorbed"][:3]))
    check("R rule effects are replayed on the next membership: the deleted file leaves, the (c) rename moves with its owner, (d) variants join their owner's area",
          "docs_src/app_testing/app_b/main.py" not in sum(nm["areas"].values(), []) and "docs_src/topic/tutorial001_an.py" in nm["areas"]["A"] and "docs_src/known/tutorial001_an.py" in nm["areas"]["B"], nm["areas"]["A"])
    check("R the next membership never mentions an arm: it is a function of the start membership, the units, the gold and the rule effects only; round r + 1's units are built against it, so one label set serves every arm",
          nm["round"] == 2 and "areas" in nm and nm["pool_areas"] is not None)
    # free running secondary and divergence
    os.makedirs(os.path.join(rd, "arms"), exist_ok=True)
    wjl(os.path.join(rd, "arms", "V-sem.jsonl"), [{"case_id": u["case_id"], "arm": "V-sem", "output": "assigned", "area_id": "A", "confidence": 0.9, "unresolved_reason": None} for u in units])
    cat = {"areas": [{"area_id": "A", "files": mem["areas"]["A"]}, {"area_id": "B", "files": mem["areas"]["B"]}]}
    os.makedirs(os.path.join(data, "catalogue"), exist_ok=True)
    json.dump(dict(cat, schema="x", heuristic_only=mem["heuristic_only"]), open(os.path.join(data, "catalogue", "catalogue.json"), "w"))
    fm = G.free_membership(1, "V-sem", data=data)
    div = G.divergence(nm, fm)
    check("R free-running secondary: an arm's own membership absorbs its own assignments; divergence from the gold membership is the file-set Jaccard per area (1.0 when identical, below 1 where the arm's wrong assignments persist)",
          fm["areas"] != nm["areas"] and 0 <= div["mean"] < 1 and G.divergence(nm, nm)["mean"] == 1.0 and div["areas_diverged"] >= 1, div)


def round_fix_checks(tmp, repo, shas, res, mem):
    """Council round 1 on synthetic data: R1-07 per-round controls, R1-14 core gate decided once, R1-11 chronological absorption, R1-08 baseline replay, R1-12 end-to-end and divergence."""
    c0, c1, _c2, c3 = shas
    info = {"after_commit": c0, "through_commit": c3, "prs": []}
    settings = dict(x2d.DEFAULT_SETTINGS)
    cat_min = {"areas": [{"area_id": "A", "title": "Core routing responsibility", "summary": "Routes requests. It also validates."},
                         {"area_id": "B", "title": "Known topic examples", "summary": "Shows the known topic. More text."}]}
    check("R build_round_units builds no controls unless asked (controls=False); the CLI asks", res["controls"] == [] and "controls=True" in open(os.path.join(S, "build_units.py")).read())
    rc = BU.build_round_units(repo.path, 1, mem, rounds_info=info, settings=settings, core_gate=False, check_prepass=False, controls=True, catalogue=cat_min)
    win, ctl = rc["units"], rc["controls"]
    by_id = {u["case_id"]: u for u in win}
    kinds = sorted(c["control_kind"] for c in ctl)
    check("R per-round controls (R1-07): ceil(10 %) dup clones of window units, one empty-rename trap and one misleading-prose clone; control true, set ctl, ids <r>:ctl-<n>, unique, never in the window set",
          kinds == ["dup"] * math.ceil(0.10 * len(win)) + ["prose", "trap"] and all(c["control"] is True and c["set"] == "ctl" and re.fullmatch(r"1:ctl-\d+", c["case_id"]) for c in ctl)
          and len({c["case_id"] for c in ctl}) == len(ctl) and not ({c["case_id"] for c in ctl} & set(by_id)), (kinds, len(win)))
    dups = [c for c in ctl if c["control_kind"] == "dup"]
    check("R a dup control is a test-retest clone: same members and unit state as its window source (source_case_id), a new case and cluster id",
          dups and all(c["source_case_id"] in by_id and c["members"] == by_id[c["source_case_id"]]["members"] and c["state"]["unit"] == by_id[c["source_case_id"]]["state"]["unit"]
                       and c["cluster_id"] != by_id[c["source_case_id"]]["cluster_id"] for c in dups))
    pr = next(c for c in ctl if c["control_kind"] == "prose")
    ps = by_id[pr["source_case_id"]]
    check("R the prose control keeps its source's members but its PR title and body talk about another area (decoy, not the source's first sibling area)",
          pr["members"] == ps["members"] and pr["state"]["pr"]["title"] != ps["state"]["pr"]["title"] and pr["decoy_area_id"] in ("A", "B") and (not ps["siblings"] or pr["decoy_area_id"] != ps["siblings"][0]["area_id"]),
          (pr["state"]["pr"]["title"], pr["decoy_area_id"]))
    tr = next(c for c in ctl if c["control_kind"] == "trap")
    check("R the trap control is an empty-file rename across directories (the package marker moved next to the docs) with no source unit", tr["members"] == ["docs_src/x2d_control_r1/pkg/__init__.py"] and tr["state"]["unit"]["change_kind"].startswith("renamed")
          and tr["source_case_id"] is None, tr["state"]["unit"])
    m1 = G._absorb(mem, win, {u["case_id"]: "B" for u in win}, [], 2, True)
    m2 = G._absorb(mem, win + ctl, dict({u["case_id"]: "B" for u in win}, **{c["case_id"]: "A" for c in ctl}), [], 2, True)
    check("R controls are never absorbed into the next membership (gold._absorb takes window rows only)", m1["areas"] == m2["areas"] and m1["pool"] == m2["pool"] and m1["unabsorbed"] == m2["unabsorbed"])
    # R1-14: the core gate is decided ONCE at round 0 and read by every round
    d_on, d_off = os.path.join(tmp, "cg-on"), os.path.join(tmp, "cg-off")
    os.makedirs(os.path.join(d_on, "rounds", "0"), exist_ok=True)
    os.makedirs(d_off, exist_ok=True)
    absent = BU.load_core_gate(d_off)
    json.dump({"gates": {}, "core_gate": {"enabled": True, "core_share_owned": 0.1}}, open(os.path.join(d_on, "rounds", "0", "gates.json"), "w"))
    on = BU.build_round_units(repo.path, 1, mem, rounds_info=info, data=d_on, settings=settings, core_gate=None, check_prepass=False)
    off = BU.build_round_units(repo.path, 1, mem, rounds_info=info, data=d_off, settings=settings, core_gate=None, check_prepass=False)
    dkey = lambda us: next(u for u in us if (u["dir"], u["stem"]) == ("fastapi", "d"))
    check("R the core gate is read from rounds/0/gates.json core_gate.enabled (absent = off), never recomputed per round: enabled -> core units are stratum 3 with reason core_gate; disabled -> stratum 1 (R1-14)",
          absent is False and BU.load_core_gate(d_on) is True and on["core_gate"] is True and off["core_gate"] is False and dkey(on["units"])["stratum"] == 3 and dkey(on["units"])["stratum_reason"] == "core_gate"
          and dkey(off["units"])["stratum"] == 1, (absent, dkey(on["units"])["stratum"], dkey(off["units"])["stratum"]))
    cu = [{"case_id": "1:a", "pr_key": "1", "set": "window", "control": False, "stratum": 3, "stratum_reason": "core_gate"},
          {"case_id": "1:b", "pr_key": "1", "set": "window", "control": False, "stratum": 1, "stratum_reason": "one_area"}]
    check("R the GO population excludes stratum_reason == core_gate units (they stay in the stratum report): analysis._go_population keeps only the other cluster (R1-14)", [k["case_id"] for k in A._go_population(A.build_cases(cu, {}, {}))] == ["1:b"])
    # R1-11: chronological absorption
    mem0 = {"areas": {"A": ["f1.py"], "B": []}, "pool": [], "heuristic_only": []}
    units0 = [{"case_id": "u1", "set": "window", "members": ["f1.py"], "seq": 1}, {"case_id": "u2", "set": "window", "members": ["f2.py"], "seq": 0}]
    inh = [{"op": "remove", "path": "f1.py", "seq": 0}, {"op": "remove", "path": "f2.py", "seq": 1}]
    mm = G._absorb(mem0, units0, {"u1": "A", "u2": "B"}, inh, 2, True)
    check("R gold absorption is chronological (R1-11): a path deleted at commit 0 and re-created at commit 1 (gold A) is in A at the end; a path absorbed at commit 0 and deleted at commit 1 is gone (an all-adds-then-all-removes order gets both wrong)",
          mm["areas"]["A"] == ["f1.py"] and mm["areas"]["B"] == [] and "f2.py" not in mm["pool"], mm)
    # R1-08: baselines replayed against the membership at each unit's parent commit
    gm, gm2 = os.path.join(tmp, "gm-data"), os.path.join(tmp, "gm2-data")
    shutil.copytree(gm, gm2)
    util_case = next(u for u in res["units"] if (u["dir"], u["stem"]) == ("docs_src", "util"))["case_id"]
    golds = jl(os.path.join(gm2, "rounds", "1", "gold.jsonl"))
    for g_ in golds:
        if g_["case_id"] == util_case:
            g_.update(covered=True, acceptable=["B"], primary="B", none_kind=None)
    os.remove(os.path.join(gm2, "rounds", "1", "gold.jsonl"))
    wjl(os.path.join(gm2, "rounds", "1", "gold.jsonl"), golds)
    mems = G.replay_memberships(1, data=gm2)
    u2 = next(u for u in res["units"] if (u["dir"], u["stem"]) == ("docs_src", "util2"))
    un = next(u for u in res["units"] if (u["dir"], u["stem"]) == ("docs_src/newtopic", "tutorial001"))
    flat = lambda m: sum(m["areas"].values(), [])
    check("R replayed memberships (R1-08): a unit of the later PR (#102) sees the earlier PR's gold (docs_src/util.py joined B) and rule effects ((c) rename fastapi/c.py moved into A); a unit of the first PR sees the round start only",
          "docs_src/util.py" in mems[u2["case_id"]]["areas"]["B"] and "fastapi/c.py" in mems[u2["case_id"]]["areas"]["A"] and "docs_src/util.py" not in flat(mems[un["case_id"]]) and "fastapi/c.py" not in flat(mems[un["case_id"]]))
    rb = BL.run_round(gm2, 1, repo.path)
    row = lambda arm: next(x for x in rb[arm] if x["case_id"] == u2["case_id"])
    start_b0 = BL.b0(u2["members"], set(F.tree_paths(repo.path, u2["parent"])), F.Ownership(G.read_membership(1, gm2)))
    check("R baselines run on the replayed membership (R1-08): B0 on docs_src/util2.py assigns B (its sibling util.py was absorbed by gold at the earlier PR) where the round-start membership gives none; the round-start B2 is kept as b2_round_start",
          row("B0")["output"] == "assigned" and row("B0")["area_id"] == "B" and start_b0[0] == "none" and row("b2_round_start")["area_id"] != "B" and len(rb["b2_round_start"]) == len(rb["B2"]) == len(res["units"] + res["audit_units"]))
    try:
        BL.run_round(os.path.join(tmp, "gm-nogold"), 1, repo.path)
        refused = False
    except SystemExit:
        refused = True
    check("R baselines for rounds >= 1 refuse to run before gold exists", refused)
    # R1-12: end-to-end and divergence on the planted round
    ee = A._end_to_end([], res["lookups"], res["audit_units"], {}, arms=())
    check("R end-to-end (R1-12): rule_exact_files counts only modified_owned (%d), modified_unowned is reported apart (%d), neither enters an accuracy" % (res["counts"]["modified_owned"], res["counts"]["modified_unowned"]),
          ee["rule_exact_files"] == res["counts"]["modified_owned"] and ee["modified_unowned"].get("modified_unowned") == res["counts"]["modified_unowned"] == 1, ee)
    dv = A._divergence(1, gm, ["V-sem", "B2"])
    check("R replayed free-running divergence (R1-12): measured for an arm with rows (G.divergence(next_membership, free_membership), mean in [0, 1)), reported 'unmeasured' with its reason for an arm without a file, never silently dropped",
          dv["name"] == "replayed free-running divergence" and dv["arms"]["V-sem"]["status"] == "measured" and 0 <= dv["arms"]["V-sem"]["value"]["mean"] < 1 and dv["arms"]["B2"]["status"] == "unmeasured"
          and dv["arms"]["B2"]["reason"] and dv["status"] == "partly unmeasured", dv)


def compose_tie_checks():
    """R1-25: none iff 1 - p_true >= s (the tuner's tie rule); a tuned 0.5 is replaced by the next grid value BELOW in score space for merge.covered, ABOVE for the others."""
    ans = lambda cid, q, perm, **kw: dict({"type": "answer", "case_id": cid, "question": q, "variant": "V-sem", "perm": perm, "area_id": None}, **kw)
    rows = []
    for cid, p in (("t1", 0.5), ("t2", 0.51), ("t3", 0.69), ("t4", 0.71)):
        rows += [ans(cid, "merge.covered", 1, p_true=p), ans(cid, "merge.which_area", 1, choice="A", confidence=0.9)]
    us = [{"case_id": c, "cluster_id": cluster_for(False, c)} for c in ("t1", "t2", "t3", "t4")]
    th_s = {"decisions": {"merge.covered": {"threshold": 0.5, "score_threshold": 0.5}, "merge.which_area": {"threshold": 0.6}, "merge.area_covers": {"threshold": 0.5}}}
    out = {r["case_id"]: r["output"] for r in CM.compose(rows, us, th_s, "V-sem")}
    check("R composition tie rule (R1-25): with the stored score threshold s = 0.5 a cluster with 1 - p_true = 0.5 is `none` (>= s), 0.49 is assigned", out["t1"] == "none" and out["t2"] == "assigned", out)
    th_p = {"decisions": {"merge.covered": {"threshold": 0.7}, "merge.which_area": {"threshold": 0.6}, "merge.area_covers": {"threshold": 0.5}}}
    outp = {r["case_id"]: r["output"] for r in CM.compose(rows, us, th_p, "V-sem")}
    check("R a provisional covered decision (threshold 0.7, no stored s) uses s = 1 - threshold: p_true 0.69 -> none, 0.71 -> assigned", outp["t3"] == "none" and outp["t4"] == "assigned", outp)
    sc = [{"score": s} for s in (0.2, 0.4, 0.7, 0.9)]
    below, above = TU.next_half(sc, True), TU.next_half(sc, False)
    check("R a tuned 0.5 is replaced by the next grid value below it for merge.covered (toward `none`, the cheaper error: 0.30) and above it for the other decisions (0.55) (R1-25)", abs(below - 0.3) < 1e-9 and abs(above - 0.55) < 1e-9, (below, above))


def cluster_for(flip, tag):
    """A cluster id whose second-permutation selection (sha256('x2d-flip' + id) mod 5 == 0) equals `flip` (deterministic search)."""
    i = 0
    while x2d.flip_selected("%s-%d" % (tag, i)) != flip:
        i += 1
    return "%s-%d" % (tag, i)


def compose_checks(tmp):
    th = {"decisions": {"merge.covered": {"threshold": 0.7}, "merge.which_area": {"threshold": 0.6}, "merge.area_covers": {"threshold": 0.5}}}
    ans = lambda cid, q, v, perm, **kw: dict({"type": "answer", "case_id": cid, "question": q, "variant": v, "perm": perm, "area_id": kw.pop("area_id", None)}, **kw)
    rows = [ans("u1", "merge.covered", "V-sem", 1, p_true=0.9), ans("u1", "merge.which_area", "V-sem", 1, choice="A", confidence=0.8),
            ans("u2", "merge.covered", "V-sem", 1, p_true=0.69), ans("u2", "merge.which_area", "V-sem", 1, choice="A", confidence=0.99),
            ans("u3", "merge.covered", "V-sem", 1, p_true=0.95), ans("u3", "merge.which_area", "V-sem", 1, choice="A", confidence=0.59),
            ans("u4", "merge.covered", "V-sem", 1, p_true=0.95), ans("u4", "merge.which_area", "V-sem", 1, choice="A", confidence=0.9), ans("u4", "merge.which_area", "V-sem", 2, choice="B", confidence=0.9),
            ans("u5", "merge.covered", "V-sem", 1, p_true=0.95), ans("u5", "merge.which_area", "V-sem", 1, choice="A", confidence=0.9), ans("u5", "merge.which_area", "V-sem", 2, choice="A", confidence=0.5),
            ans("u6", "merge.covered", "V-sem", 1, p_true=0.95), {"type": "failure", "case_id": "u6", "question": "merge.which_area", "variant": "V-sem", "perm": 1, "area_id": None, "kind": "http"},
            {"type": "failure", "case_id": "u7", "question": "merge.covered", "variant": "V-sem", "perm": 1, "area_id": None, "kind": "http"}]
    for a_, p in (("A", 0.8), ("B", 0.2), ("C", 0.6)):
        rows.append(ans("f1", "merge.area_covers", "V-sem", 1, area_id=a_, p_true=p))
    for a_ in ("A", "B"):
        rows.append(ans("f2", "merge.area_covers", "V-sem", 1, area_id=a_, p_true=0.1))
    rows += [ans("f3", "merge.area_covers", "V-sem", 1, area_id="A", p_true=0.9), {"type": "failure", "case_id": "f3", "question": "merge.area_covers", "variant": "V-sem", "perm": 1, "area_id": "B", "kind": "http"}]
    flipset = {"u4", "u5", "u8"}
    us = [{"case_id": c, "cluster_id": cluster_for(c in flipset, c)} for c in ("u1", "u2", "u3", "u4", "u5", "u6", "u7", "u8")]
    rows.append(ans("u8", "merge.covered", "V-sem", 1, p_true=0.95))
    rows.append(ans("u8", "merge.which_area", "V-sem", 1, choice="A", confidence=0.9))  # flip-selected cluster, permutation-2 row omitted
    out = {r["case_id"]: r for r in CM.compose(rows, us, th, "V-sem")}
    check("R composition (V-sem): p_true >= t_cov then the Choice top option is assigned when its confidence >= t_choice; p_true < t_cov -> none (the pool), never an assignment; confidence below t_choice -> unresolved (reported, not an assignment)",
          out["u1"]["output"] == "assigned" and out["u1"]["area_id"] == "A" and out["u2"]["output"] == "none" and out["u3"]["output"] == "unresolved" and out["u3"]["unresolved_reason"] == "below_t_choice")
    check("R composition: a cluster asked under two permutations is assigned only if both agree (the confidence is their mean); disagreement -> unresolved; a failed request is unresolved (reason failure) and never dropped or counted as none",
          out["u4"]["output"] == "unresolved" and out["u4"]["unresolved_reason"] == "permutation_disagreement" and out["u5"]["output"] == "assigned" and abs(out["u5"]["confidence"] - 0.7) < 1e-12
          and out["u6"]["output"] == "unresolved" and out["u6"]["unresolved_reason"] == "failure" and out["u7"]["output"] == "unresolved" and len(out) == 8)
    check("R2-13 a V-sem flip-selected cluster without a permutation-2 row (neither answer nor failure) composes `unresolved:failure`, never an assignment on permutation 1 alone",
          out["u8"]["output"] == "unresolved" and out["u8"]["unresolved_reason"] == "failure" and out["u8"]["raw_choice"] == "A" and out["u8"]["raw_covered"] is True, out["u8"])
    full_fan = lambda c: [ans(c, "merge.area_covers", "V-sem", 1, area_id=a_, p_true=0.9 if a_ == "A" else 0.1) for a_ in ("A", "B", "C")]
    m_rows = [ans("m1", "merge.covered", "V-sem", 1, p_true=0.9), ans("m1", "merge.which_area", "V-sem", 1, choice="A", confidence=0.9)] + full_fan("m1") + \
             [ans("m2", "merge.covered", "V-sem", 1, p_true=0.9), ans("m2", "merge.which_area", "V-sem", 1, choice="A", confidence=0.9)] + full_fan("m2") + \
             [ans("m3", "merge.covered", "V-sem", 1, p_true=0.9), ans("m3", "merge.which_area", "V-sem", 1, choice="A", confidence=0.9),
              ans("m4", "merge.covered", "V-sem", 1, p_true=0.9), ans("m4", "merge.which_area", "V-sem", 1, choice="A", confidence=0.9), ans("m4", "merge.area_covers", "V-sem", 1, area_id="A", p_true=0.9)]
    m_units = [{"case_id": "m1", "cluster_id": cluster_for(False, "m1")}, {"case_id": "m2", "cluster_id": cluster_for(True, "m2")}, {"case_id": "m3", "cluster_id": cluster_for(False, "m3")}, {"case_id": "m4", "cluster_id": cluster_for(False, "m4")}]
    miss = CM.missing_required(m_rows, m_units, area_ids=["A", "B", "C"])
    check("R2-13 / R3-06 missing_required: when the file has fan-out rows, EVERY area of the catalogue needs an answer or failure row per unit: m3 (no area row) lacks A, B, C and m4 (only A answered: a partial fan-out) lacks B and C; the flip-selected m2 lacks permutation 2; m1 lacks nothing",
          miss == [("m2", "V-sem/merge.which_area", 2)] + [("m3", "fan-out/merge.area_covers", a_) for a_ in "ABC"] + [("m4", "fan-out/merge.area_covers", a_) for a_ in "BC"], miss)
    fan_out = {r["case_id"]: r for r in CM.compose(m_rows, m_units, th, "fan-out", area_ids=["A", "B", "C"])}
    check("R3-06 pure `compose` of a partial fan-out (m4: A answered, B and C never asked) is `unresolved:failure`, never an assignment or `none`; a complete fan-out (m1: A 0.9 >= t_fan) is assigned A; a unit without any area row (m3) is unresolved",
          fan_out["m4"]["output"] == "unresolved" and fan_out["m4"]["unresolved_reason"] == "failure" and fan_out["m1"]["output"] == "assigned" and fan_out["m1"]["area_id"] == "A" and fan_out["m3"]["output"] == "unresolved", fan_out)
    cd = os.path.join(tmp, "compose-cli")
    wjl(os.path.join(cd, "rounds", "1", "jev", "scores.jsonl"), [{"type": "header", "flip_all": False}] + [r for r in rows if r["case_id"] in ("u1", "u8")])
    wjl(os.path.join(cd, "rounds", "1", "units.jsonl"), [u for u in us if u["case_id"] in ("u1", "u8")])
    json.dump(th, open(os.path.join(cd, "rounds", "1", "thresholds.json"), "w"))
    rcli = run([PY, os.path.join(S, "compose.py"), "--round", "1"], cwd=ROOT, env=clean_env(X2D_DATA=cd))
    check("R2-13 compose.py refuses a scores file whose required request has neither answer nor failure row (non-zero exit, names the missing request, writes no arm file)",
          rcli.returncode != 0 and "required Jev request" in (rcli.stdout + rcli.stderr) and not os.path.exists(os.path.join(cd, "rounds", "1", "arms")), (rcli.returncode, (rcli.stdout + rcli.stderr)[-300:]))
    fo = {r["case_id"]: r for r in CM.compose(rows, [{"case_id": c} for c in ("f1", "f2", "f3")], th, "fan-out")}
    check("R composition (Noul fan-out): the set of areas with p_true >= t_fan is the acceptable set and the argmax is assigned; an empty set -> none; a failed area request -> unresolved",
          fo["f1"]["output"] == "assigned" and fo["f1"]["area_id"] == "A" and fo["f1"]["acceptable_set"] == ["A", "C"] and fo["f2"]["output"] == "none" and fo["f3"]["output"] == "unresolved")
    flips = CM.compose_k5([ans("a", "merge.which_area", "V-sem", 1, choice="A", confidence=0.9), ans("a", "merge.which_area", "V-sem", 2, choice="B", confidence=0.9),
                           ans("b", "merge.which_area", "V-sem", 1, choice="A", confidence=0.9), ans("b", "merge.which_area", "V-sem", 2, choice="A", confidence=0.9),
                           ans("c", "merge.which_area", "V-sem", 1, choice="A", confidence=0.9)])
    check("R K5 order-flip statistics: the flip rate is over units asked under both permutations only (1 flip of 2 = 0.5)", flips["n"] == 2 and flips["flips"] == 1 and flips["rate"] == 0.5)


def tune_checks(tmp):
    script = x2d.TUNE_SCRIPT
    if not os.path.exists(script):
        check("R the jev-actions tune_threshold.py script is present", False, script)
        return
    rng = random.Random(5)

    def write_round(data, r, n, sep=0.35, label_noise=True, ok_p=0.75):
        rd = os.path.join(data, "rounds", str(r))
        units, gold, rows = [], [], [{"type": "header", "versions": {"merge.covered/V-sem": "v1-a", "merge.which_area/V-sem": "v1-a", "merge.area_covers/V-sem": "v1-a"}, "model": x2d.JEV_MODEL, "catalogue_sha256": "c" * 64}]
        for i in range(n):
            cid = "%d:u%03d" % (r, i)
            units.append({"case_id": cid, "cluster_id": "u%d_%d" % (r, i), "pr_key": str(i // 3), "set": "D" if r == 0 and i % 2 else ("P" if r == 0 else "window"), "control": False, "stratum": 1, "members": ["f%d_%d.py" % (r, i)]})
            cov = rng.random() < 0.6
            gold.append({"case_id": cid, "covered": cov, "acceptable": ["A"] if cov else [], "primary": "A" if cov else None, "none_kind": None if cov else "gap", "agreement": "agreed", "control_kind": None})
            ok = rng.random() < ok_p
            pt = min(0.99, max(0.01, (0.8 if cov else 0.2) + rng.uniform(-sep, sep) * (1 if label_noise else 0)))
            rows.append({"type": "answer", "case_id": cid, "question": "merge.covered", "variant": "V-sem", "perm": 1, "area_id": None, "p_true": pt, "cost_known": True, "cost_usd": 0,
                         "model": x2d.JEV_MODEL, "question_version": "q"})
            rows.append({"type": "answer", "case_id": cid, "question": "merge.which_area", "variant": "V-sem", "perm": 1, "area_id": None, "choice": "A" if ok else "B",
                         "confidence": min(0.99, max(0.3, (0.9 if ok else 0.55) + rng.uniform(-0.15, 0.15))), "cost_known": True, "cost_usd": 0, "model": x2d.JEV_MODEL, "question_version": "q"})
            if x2d.flip_selected(units[-1]["cluster_id"]):  # R4-11: the round's own scores must be complete: the second permutation of the flip-selected clusters (same answer)
                rows.append(dict(rows[-1], perm=2))
        wjl(os.path.join(rd, "units.jsonl"), units)
        wjl(os.path.join(rd, "gold.jsonl"), gold)
        wjl(os.path.join(rd, "jev", "scores.jsonl"), rows)

    data = os.path.join(tmp, "tune-data")
    write_round(data, 0, 14)
    t1 = TU.tune_round(1, data=data, script=script)
    check("R tuning round 1 uses only the round-0 D and P units: fewer than 30 cases per label keeps the provisional 0.7 for every decision (source provisional-0.7), never 0.5",
          all(d["threshold"] == 0.7 and d["source"] == "provisional-0.7" for d in t1["decisions"].values()) and t1["dev_rounds"] == [0], t1["decisions"]["merge.which_area"])
    t0 = TU.tune_round(0, data=data, script=script)
    check("R tuning round 0 (no dev rows) writes the provisional 0.7 thresholds so the round-0 arms can be composed", all(d["threshold"] == 0.7 for d in t0["decisions"].values()) and t0["dev_rounds"] == [])
    try:
        TU.tune_round(1, data=data, script=script)
        refused = False
    except SystemExit:
        refused = True
    check("R tuning is append-only: a second run for a round with thresholds.json is refused", refused)
    for r in (1, 2, 3):
        write_round(data, r, 90)
    t4 = TU.tune_round(4, data=data, script=script)
    wc = t4["decisions"]["merge.which_area"]
    comp = t4.get("composed_dev")
    prim_q = ("merge.covered", "merge.which_area")
    rejected = comp is not None and comp["minimums_met"] is False
    tuned_fine = all(t4["decisions"][q]["threshold"] != 0.5 and t4["decisions"][q]["fold_report"] is not None for q in prim_q if t4["decisions"][q]["source"] == "tuned")
    reverted = all(t4["decisions"][q]["threshold"] == 0.7 and t4["decisions"][q]["source"] == "provisional-0.7" for q in prim_q)
    check("R tuning round 4 pools rounds 0-3 (>= 50 cases per label): tuned by the skill's script (stratified 5-fold, seed 20261029) with the fold report, never 0.5; the composed rule is evaluated at the tuned pair on the dev rows "
          "(composed_dev: P_assign, R_assign, P_new, R_new) and the primary arm's decisions are either tuned and meeting the minimums or reverted to the provisional 0.7 (R1-10)",
          t4["dev_rounds"] == [0, 1, 2, 3] and comp is not None and all(k in comp for k in ("P_assign", "R_assign", "P_new", "R_new", "minimums_met")) and tuned_fine and (reverted if rejected else True)
          and t4["seed"] == x2d.SEED and t4["skill_script_sha256"] == x2d.sha256_file(script) and t4["dev_scores_sub"] == "jev" and t4["reused_from"] is None, (comp, wc))
    bad = os.path.join(tmp, "tune-data-bad")
    for r_ in (0, 1, 2, 3):
        write_round(bad, r_, 14 if r_ == 0 else 90, ok_p=0.5)
    tb = TU.tune_round(4, data=bad, script=script)
    check("R composed-rule check (R1-10): when the Choice is right only half the time no tuned pair can meet R_assign >= 0.80 on the dev rows, so every decision of the primary arm is the provisional 0.7 and composed_dev records the failure",
          all(tb["decisions"][q]["threshold"] == 0.7 and tb["decisions"][q]["source"] == "provisional-0.7" for q in prim_q) and tb["composed_dev"]["minimums_met"] is not True, tb["composed_dev"])
    write_round(data, 4, 60)
    import shutil as sh
    d2 = os.path.join(tmp, "tune-data2")
    sh.copytree(data, d2)
    os.remove(os.path.join(d2, "rounds", "4", "thresholds.json"))
    for fn in os.listdir(os.path.join(d2, "rounds", "4")):
        if fn.startswith("scores") and fn.endswith(".csv"):
            os.remove(os.path.join(d2, "rounds", "4", fn))
    gold4 = jl(os.path.join(d2, "rounds", "4", "gold.jsonl"))
    for g in gold4:
        g["covered"] = not g["covered"]  # flip every gold label of the round being scored
        g["acceptable"], g["primary"], g["none_kind"] = (["A"], "A", None) if g["covered"] else ([], None, "gap")
    os.remove(os.path.join(d2, "rounds", "4", "gold.jsonl"))
    wjl(os.path.join(d2, "rounds", "4", "gold.jsonl"), gold4)
    t4b = TU.tune_round(4, data=d2, script=script)
    check("R tuning never uses the scores of the round being scored: flipping every gold label of round 4 leaves round 4's thresholds unchanged (dev set = rounds < 4 only)",
          {q: d["threshold"] for q, d in t4b["decisions"].items()} == {q: d["threshold"] for q, d in t4["decisions"].items()} and t4b["dev_rounds"] == [0, 1, 2, 3])
    cols = open(os.path.join(data, "rounds", "4", "scores.merge.covered.csv")).readline().strip()
    check("R scores.csv has the preregistered columns, one file per question plus the pooled file", cols == "case_id,label,score,model,question_version,split,stratum,fallback" and os.path.exists(os.path.join(data, "rounds", "4", "scores.csv")))
    for r in (5, 6, 7, 8, 9):
        write_round(data, r, 30)
    check("R the frozen rounds 8-10 never feed tuning (dev rounds of round 10 exclude 8 and 9)", TU.dev_rounds(10, data) == [0, 1, 2, 3, 4, 5, 6, 7], TU.dev_rounds(10, data))
    d8 = os.path.join(tmp, "tune-data8")
    sh.copytree(data, d8)
    try:
        TU.tune_round(8, data=d8, script=script)
        refused8 = False
    except SystemExit:
        refused8 = True
    check("R round 8 refuses to tune until rounds 0-7 are re-scored under v_7 (D/rounds/<rr>/jev-v7/scores.jsonl, R1-10) and writes nothing", refused8 and not os.path.exists(os.path.join(d8, "rounds", "8", "thresholds.json")))
    cat_v7 = os.path.join(d8, "catalogue", "catalogue.json")  # the fan-out must have answered every area of the round's catalogue (R3-18)
    os.makedirs(os.path.dirname(cat_v7), exist_ok=True)
    json.dump({"areas": [{"area_id": "A"}, {"area_id": "B"}]}, open(cat_v7, "w"))
    os.makedirs(os.path.join(d8, "rounds", "0", "jev-in"), exist_ok=True)
    json.dump([{"name": "main", "catalogue": cat_v7, "units": os.path.join(d8, "rounds", "0", "units.jsonl")}], open(os.path.join(d8, "rounds", "0", "jev-in", "manifest.json"), "w"))
    for rr in range(0, 8):
        src = os.path.join(d8, "rounds", str(rr), "jev", "scores.jsonl")
        if os.path.exists(src):
            os.makedirs(os.path.join(d8, "rounds", str(rr), x2d.V7_RESCORE_SUB), exist_ok=True)
            rows_v7 = jl(src)
            devu = [u_ for u_ in jl(os.path.join(d8, "rounds", str(rr), "units.jsonl")) if not u_.get("control")]
            for u_ in devu:
                for a_ in ("A", "B"):
                    rows_v7.append({"type": "answer", "case_id": u_["case_id"], "question": "merge.area_covers", "variant": "V-sem", "perm": 1, "area_id": a_, "p_true": 0.8 if a_ == "A" else 0.1,
                                    "cost_known": True, "cost_usd": 0, "model": x2d.JEV_MODEL, "question_version": "q"})
                if x2d.flip_selected(u_["cluster_id"]):  # the second permutation of the flip-selected clusters
                    p1 = next((r_ for r_ in rows_v7 if r_.get("case_id") == u_["case_id"] and r_.get("question") == "merge.which_area" and r_.get("perm") == 1), None)
                    if p1:
                        rows_v7.append(dict(p1, perm=2))
            up_v7 = os.path.join(d8, "rounds", str(rr), x2d.V7_RESCORE_SUB, "units.jsonl")
            wjl(up_v7, [dict(u_, excerpt_lines=int(x2d.round_settings(7, d8)["excerpt_lines"])) for u_ in devu])
            rows_v7[0] = dict(rows_v7[0], input_sha256=x2d.sha256_file(up_v7))  # R4-06: the header's input digest is the digest of the units file the re-score read
            wjl(os.path.join(d8, "rounds", str(rr), x2d.V7_RESCORE_SUB, "scores.jsonl"), rows_v7)

    def broken(tag, mutate):
        dd = os.path.join(tmp, "tune-data8-" + tag)
        sh.copytree(d8, dd)
        mutate(os.path.join(dd, "rounds"))
        try:
            TU.dev_rounds(8, dd, x2d.V7_RESCORE_SUB)
            return False
        except SystemExit as ex:
            return str(ex.code)
    v7p = lambda rr, base: os.path.join(base, str(rr), x2d.V7_RESCORE_SUB, "scores.jsonl")

    def mut_header(base):
        rows_ = jl(v7p(3, base))
        rows_[0] = dict(rows_[0], versions={"merge.covered/V-sem": "v1-OLD", "merge.which_area/V-sem": "v1-a", "merge.area_covers/V-sem": "v1-a"})
        wjl(v7p(3, base), rows_)

    def mut_model(base):
        rows_ = jl(v7p(3, base))
        rows_[0] = dict(rows_[0], model="typesafe/jev-0.0")
        wjl(v7p(3, base), rows_)

    def mut_catalogue(base):
        rows_ = jl(v7p(3, base))
        rows_[0] = dict(rows_[0], catalogue_sha256="d" * 64)
        wjl(v7p(3, base), rows_)

    def mut_empty(base):
        open(v7p(2, base), "w").close()

    def mut_headerless(base):
        wjl(v7p(2, base), [r_ for r_ in jl(v7p(2, base)) if r_.get("type") != "header"])

    def mut_incomplete(base):
        rows_ = jl(v7p(5, base))
        drop = next(r_["case_id"] for r_ in rows_ if r_.get("question") == "merge.covered")
        wjl(v7p(5, base), [r_ for r_ in rows_ if not (r_.get("case_id") == drop and r_.get("question") == "merge.covered")])
    def mut_extra(base):
        rows_ = jl(v7p(3, base))
        rows_[0] = dict(rows_[0], versions=dict(rows_[0]["versions"], **{"merge.area_covers/V-sib": "v1-extra"}))
        wjl(v7p(3, base), rows_)

    def mut_novers(base):
        rows_ = jl(v7p(3, base))
        rows_[0] = dict(rows_[0], versions={})
        wjl(v7p(3, base), rows_)

    def mut_subset(base):  # R3-18: a header with only merge.covered (a hand-run covered-only re-score) is refused
        rows_ = jl(v7p(3, base))
        rows_[0] = dict(rows_[0], versions={"merge.covered/V-sem": rows_[0]["versions"]["merge.covered/V-sem"]})
        wjl(v7p(3, base), rows_)

    def mut_choice(base):  # a required merge.which_area permutation row missing
        rows_ = jl(v7p(4, base))
        drop = next(r_ for r_ in rows_ if r_.get("question") == "merge.which_area" and r_.get("perm") == 1)
        wjl(v7p(4, base), [r_ for r_ in rows_ if r_ is not drop])

    def mut_fan(base):  # one area row of one unit's fan-out missing
        rows_ = jl(v7p(6, base))
        drop = next(r_ for r_ in rows_ if r_.get("question") == "merge.area_covers" and r_.get("area_id") == "B")
        wjl(v7p(6, base), [r_ for r_ in rows_ if r_ is not drop])

    def mut_excerpt(base):  # the re-score units were rendered under another excerpt_lines than round 7's (the header digest is re-bound so that THIS check, not R4-06's, fires)
        up = os.path.join(base, "5", x2d.V7_RESCORE_SUB, "units.jsonl")
        wjl(up, [dict(u_, excerpt_lines=30) for u_ in jl(up)])
        hdr_rows = jl(v7p(5, base))
        hdr_rows[0] = dict(hdr_rows[0], input_sha256=x2d.sha256_file(up))
        wjl(v7p(5, base), hdr_rows)
    refusals = {tag: broken(tag, fn) for tag, fn in (("versions", mut_header), ("model", mut_model), ("catalogue", mut_catalogue), ("empty", mut_empty), ("headerless", mut_headerless), ("incomplete", mut_incomplete),
                                                      ("extra", mut_extra), ("novers", mut_novers), ("coveredonly", mut_subset), ("choice", mut_choice), ("fan", mut_fan), ("excerpt", mut_excerpt))}
    check("R2-02 / R3-18 tune.dev_rounds(8, data, 'jev-v7') refuses a header whose versions (a changed or an extra version, none, or a covered-only header) / model / catalogue differ from round 7's, an empty file, a headerless file, a re-score missing a dev unit's merge.covered row, "
          "a Choice permutation row or one fan-out area row, and units rendered under another excerpt_lines (R3-05); the complete re-score (three versions equal round 7's, everything answered, round 7's excerpt_lines) is accepted",
          all(v is not False for v in refusals.values()) and "versions" in refusals["versions"] and "model" in refusals["model"] and "catalogue_sha256" in refusals["catalogue"] and "empty" in refusals["empty"]
          and "no header" in refusals["headerless"] and "incomplete" in refusals["incomplete"] and "versions" in refusals["extra"] and "versions" in refusals["novers"] and "versions" in refusals["coveredonly"]
          and "incomplete" in refusals["choice"] and "merge.which_area" in refusals["choice"] and "incomplete" in refusals["fan"] and "merge.area_covers" in refusals["fan"] and "excerpt_lines" in refusals["excerpt"], refusals)
    t8 = TU.tune_round(8, data=d8, script=script)
    t9 = TU.tune_round(9, data=d8, script=script)
    check("R2-02 a correct re-score (the same header as round 7's, every dev unit answered) tunes once on the v_7 rows (dev rounds 0-7, dev_scores_sub jev-v7); round 9 reuses round 8's decisions unchanged (reused_from 8) (R1-10)",
          t8["dev_scores_sub"] == x2d.V7_RESCORE_SUB and t8["dev_rounds"] == list(range(8)) and t9["reused_from"] == 8 and t9["decisions"] == t8["decisions"] and t9["dev_rounds"] == t8["dev_rounds"], (t8["dev_rounds"], t9["reused_from"]))
    # R2-13: audit units are absent from tune dev rows (round >= 1: window only)
    adata = os.path.join(tmp, "tune-audit")
    write_round(adata, 1, 12)
    us_a = jl(os.path.join(adata, "rounds", "1", "units.jsonl"))
    wjl(os.path.join(adata, "rounds", "1", "units.jsonl"), [dict(u, set="audit") if i % 2 else u for i, u in enumerate(us_a)])
    dv = TU.round_rows(1, adata, "dev")
    ids_dev = {r_["case_id"] for rows_ in dv.values() for r_ in rows_}
    audit_ids = {u["case_id"] for u in jl(os.path.join(adata, "rounds", "1", "units.jsonl")) if u["set"] == "audit"}
    check("R2-13 audit units (set audit) never enter the tuning dev rows of a refinement round (only window units; D and P in round 0)", ids_dev and not (ids_dev & audit_ids) and len(ids_dev) == 6, (len(ids_dev), len(audit_ids)))


def versioning_checks(tmp):
    data = os.path.join(tmp, "ver-data")
    os.makedirs(os.path.join(data, "catalogue"), exist_ok=True)
    json.dump({"areas": []}, open(os.path.join(data, "catalogue", "catalogue.json"), "w"))
    v1 = VR.freeze_version(1, data=data)
    check("R freezing a round writes version.json and settings.json (question files and structure digests, state builder, guide, thresholds, catalogue, settings, code digests) and refuses to overwrite", os.path.exists(os.path.join(data, "rounds", "1", "version.json"))
          and os.path.exists(os.path.join(data, "rounds", "1", "settings.json")) and v1["catalogue_sha256"] and v1["code_digests"] and set(v1["question_files"]) >= {"merge.covered@v1.json", "merge.which_area@v1.json"} and v1["guide_sha256"], list(v1))
    try:
        VR.freeze_version(1, data=data)
        again = False
    except SystemExit:
        again = True
    check("R a round's version is append-only", again)
    mk = lambda **kw: json.loads(json.dumps(dict(v1, round=2, **kw)))
    ok_settings = dict(v1["settings"], sample_files_per_area=8, excerpt_lines=120, v_sib=False, include_pr_title=False, field_order=["catalogue", "unit", "pr", "diff_summary"], order_rule="fixed")
    check("R envelope: the menu is allowed (wording of question instructions/criteria, sample files 3-8, excerpt 30-120 lines, V-sib off, PR title in/out of V-sem and V-sib (R1-28; the V-title arm is always on), field order, order rule within permutation/fixed): no violation",
          VR.check_envelope(v1, mk(settings=ok_settings)) == [] and VR.check_envelope(v1, mk(question_files=dict(v1["question_files"], **{"merge.covered@v1.json": "0" * 64}))) == [], VR.check_envelope(v1, mk(settings=ok_settings)))
    check("R the refinement menu (R1-28): include_pr_title is a bool setting (default true), v_title is gone (the V-title arm is always on), and a non-bool include_pr_title is a violation",
          x2d.DEFAULT_SETTINGS["include_pr_title"] is True and "v_title" not in x2d.DEFAULT_SETTINGS and VR.settings_violations(v1["settings"]) == [] and len(VR.settings_violations(dict(v1["settings"], include_pr_title="yes"))) >= 1)
    refused = {
        "a changed catalogue digest": mk(catalogue_sha256="1" * 64),
        "a changed state builder (build_units.py)": mk(state_builder=dict(v1["state_builder"], **{"build_units.py": "2" * 64})),
        "a changed frozen thresholds.json (minimums or the GO rule)": mk(thresholds_sha256="3" * 64),
        "a changed round table": mk(prepass_rounds_sha256="4" * 64),
        "a changed model pin": mk(model="typesafe/jev-1.15-20270101"),
        "a changed PR cap": mk(pr_cap=8),
        "changed code under test": mk(code_digests=dict(v1["code_digests"], **{"go.mod": "5" * 64})),
        "sample files outside 3-8": mk(settings=dict(v1["settings"], sample_files_per_area=9)),
        "an excerpt outside 30-120 lines": mk(settings=dict(v1["settings"], excerpt_lines=29)),
        "a field order that drops a field": mk(settings=dict(v1["settings"], field_order=["unit", "pr", "catalogue"])),
        "an order rule outside the menu": mk(settings=dict(v1["settings"], order_rule="random")),
        "a changed question kind": mk(question_meta=dict(v1["question_meta"], **{"merge.covered@v1.json": dict(v1["question_meta"]["merge.covered@v1.json"], kind="choice")})),
        "a question state shape outside the menu": mk(question_meta=dict(v1["question_meta"], **{"merge.covered@v1.json": dict(v1["question_meta"]["merge.covered@v1.json"], state_shape=["state.x"])})),
        "a question structure change that is not wording": mk(question_structure=dict(v1["question_structure"], **{"merge.which_area@v1.json": "6" * 64})),
    }
    for name, nv in refused.items():
        viol = VR.check_envelope(v1, nv)
        check("R envelope refusal: " + name + " is refused (a violation is reported)", len(viol) >= 1, viol)
    v7 = json.loads(json.dumps(dict(v1, round=7)))
    v8 = json.loads(json.dumps(dict(v1, round=8)))
    check("R frozen rounds: rounds 8-10 refuse ANY change relative to v_7, even a permitted menu change, and accept an identical version", VR.check_envelope(v7, v8) == [] and len(VR.check_envelope(v7, dict(v8, settings=ok_settings))) >= 1
          and len(VR.check_envelope(v7, dict(v8, question_files=dict(v1["question_files"], **{"merge.covered@v1.json": "0" * 64})))) >= 1)
    # the CLI path run.sh uses
    env = clean_env(X2D_DATA=data)
    r = run([PY, os.path.join(S, "versioning.py"), "--round", "2"], cwd=ROOT, env=env)
    r2 = run([PY, os.path.join(S, "versioning.py"), "--round", "2", "--check"], cwd=ROOT, env=env)
    os.makedirs(os.path.join(data, "rounds", "3"), exist_ok=True)
    json.dump(dict(ok_settings, sample_files_per_area=99), open(os.path.join(data, "rounds", "3", "settings.json"), "w"))
    r3 = run([PY, os.path.join(S, "versioning.py"), "--round", "3"], cwd=ROOT, env=env)
    check("R versioning CLI: a clean round freezes and passes --check against the previous round; a settings.json outside the menu is refused at freeze time (non-zero exit, nothing written)",
          r.returncode == 0 and r2.returncode == 0 and r3.returncode != 0 and not os.path.exists(os.path.join(data, "rounds", "3", "version.json")), (r.stderr[-200:], r2.stderr[-200:], r3.stderr[-200:]))
    # R2-01: in-envelope question wording change between rounds, pinned by the previous round's critique (`next_inputs_sha256`)
    vp_ = os.path.join(S, "versioning.py")
    vr = lambda *a: run([PY, vp_, *a], cwd=ROOT, env=env)
    os.remove(os.path.join(data, "rounds", "3", "settings.json"))
    shutil.rmtree(x2d.inputs_dir(3, data))  # the refused freeze had already created round 3's inputs
    ri = vr("--round", "3", "--inputs")
    qf = os.path.join(x2d.questions_dir(3, data), "merge.covered@v1.json")
    qd = json.load(open(qf))
    qd["instructions"] = qd["instructions"] + "\nRound-3 wording refinement."
    json.dump(qd, open(qf, "w"), indent=1)
    rf, rc = vr("--round", "3"), vr("--round", "3", "--check")
    v2_, v3_ = json.load(open(os.path.join(data, "rounds", "2", "version.json"))), json.load(open(os.path.join(data, "rounds", "3", "version.json")))
    check("R2-01 an in-envelope wording change of round 3's inputs (question instructions) freezes and passes `versioning --check`; only that question file and the inputs digest differ from round 2, the guide and the guide version do not",
          ri.returncode == 0 and rf.returncode == 0 and rc.returncode == 0 and v3_["inputs_sha256"] != v2_["inputs_sha256"] and v3_["guide_version"] == v2_["guide_version"]
          and [k for k in v3_["question_files"] if v3_["question_files"][k] != v2_["question_files"][k]] == ["merge.covered@v1.json"], (ri.stderr[-200:], rf.stderr[-200:], rc.stderr[-200:]))
    rev = os.path.join(tmp, "ver-reviews")
    os.makedirs(rev, exist_ok=True)

    def write_crit(n, nxt, verdict=None):
        with open(x2d.critique_path(n, rev), "w") as f:
            f.write("critique\n%s\n%s%s\n%s%s\n" % (verdict or x2d.CRITIQUE_VERDICT_REFINEMENT, x2d.CRITIQUE_VERSION_LINE, x2d.sha256_file(os.path.join(data, "rounds", str(n), "version.json")), x2d.CRITIQUE_INPUTS_LINE, nxt))
    write_crit(2, x2d.inputs_digest(x2d.inputs_dir(3, data)))
    pinned_ok = x2d.critique_problems(2, data, rev)
    write_crit(2, x2d.inputs_digest(x2d.inputs_dir(2, data)))
    pinned_stale = x2d.critique_problems(2, data, rev)
    write_crit(2, x2d.inputs_digest(x2d.inputs_dir(3, data)))
    open(qf, "a").write(" ")
    pinned_after_freeze = x2d.critique_problems(2, data, rev)  # a post-freeze edit does not change what the critique approved; versioning --drift refuses it
    drift_edit = vr("--round", "3", "--drift")
    json.dump(qd, open(qf, "w"), indent=1)
    check("R2-01 the round-2 critique's next_inputs_sha256 pins the inputs round 3 was frozen with: the approved digest passes, round 2's own (stale) digest is refused; an edit of round 3's inputs after the freeze leaves the critique valid but is refused by `versioning --drift`",
          pinned_ok == [] and any(x2d.CRITIQUE_INPUTS_LINE.strip() in p for p in pinned_stale) and pinned_after_freeze == [] and drift_edit.returncode != 0 and "question_files" in drift_edit.stderr, (pinned_ok, pinned_stale, pinned_after_freeze, drift_edit.stderr[-200:]))
    # an out-of-envelope change (a question's kind) is refused by `--check`
    vr("--round", "4", "--inputs")
    q4 = os.path.join(x2d.questions_dir(4, data), "merge.covered@v1.json")
    j4 = json.load(open(q4))
    j4["kind"] = "choice"
    json.dump(j4, open(q4, "w"), indent=1)
    f4, c4 = vr("--round", "4"), vr("--round", "4", "--check")
    check("R2-01 an out-of-envelope change (a question's kind) in round 4's inputs is refused by `versioning --check` (non-zero, names the violation)", f4.returncode == 0 and c4.returncode != 0 and "VIOLATION" in c4.stderr and "kind" in c4.stderr, (f4.stderr[-200:], c4.stderr[-300:]))
    # a post-sign-off edit of a study-directory question file is refused by the pins (the v_0 base cannot move)
    old_so, tune_sha = x2d.SIGNOFF, (x2d.sha256_file(x2d.TUNE_SCRIPT) if os.path.exists(x2d.TUNE_SCRIPT) else "0" * 64)
    qrel = "questions/merge.covered@v1.json"
    rec_q = os.path.join(tmp, "signoff-q")
    os.makedirs(os.path.join(rec_q, "questions"))
    shutil.copy(os.path.join(S, qrel), os.path.join(rec_q, qrel))
    try:
        x2d.SIGNOFF = os.path.join(tmp, "signoff-q.json")
        json.dump({"files": {qrel: x2d.sha256_file(os.path.join(S, qrel))}, "code_under_test": go_digest_subset(), "tune_script_sha256": tune_sha}, open(x2d.SIGNOFF, "w"))
        ok_q = [p for p in x2d.verify_pins(require_signoff=True, root=rec_q) if not p.startswith("X2 ")]
        jq = json.load(open(os.path.join(rec_q, qrel)))
        jq["instructions"] += " edited after sign-off"
        json.dump(jq, open(os.path.join(rec_q, qrel), "w"))
        bad_q = x2d.verify_pins(require_signoff=True, root=rec_q)
    finally:
        x2d.SIGNOFF = old_so
    check("R2-01 a post-sign-off edit of a study-directory question file is refused by the pins (frozen file questions/merge.covered@v1.json named)", not ok_q and any("frozen file " + qrel in p for p in bad_q), (ok_q, bad_q))
    # mid-round clarification (rounds 1-7, before scoring): recorded revision, new guide version, earlier rounds untouched
    v1_before = open(os.path.join(data, "rounds", "1", "version.json"), "rb").read()
    vr("--round", "5")
    gp5 = x2d.guide_path(5, data)
    gv_old = x2d.guide_version(5, data)
    c_same = vr("--round", "5", "--clarify", "no edit")
    open(gp5, "a").write("\nClarification: a test belongs with the responsibility it tests.\n")
    d_after_edit = vr("--round", "5", "--drift")
    c_empty = vr("--round", "5", "--clarify", " ")
    c_ok = vr("--round", "5", "--clarify", "clarify test placement")
    d_after = vr("--round", "5", "--drift")
    revs = jl(VR.revisions_path(5, data))
    gv_new = x2d.guide_version(5, data)
    eff = VR.effective_version(5, data)
    check("R2-01 a clarification (versioning --clarify) is refused without a guide edit and without a reason, an unrecorded guide edit is drift; once recorded the round's effective guide version is the new one, version.json is untouched and there is no drift; earlier rounds are untouched",
          c_same.returncode != 0 and d_after_edit.returncode != 0 and c_empty.returncode != 0 and c_ok.returncode == 0 and d_after.returncode == 0 and len(revs) == 1 and revs[0]["guide_version"] == gv_new != gv_old
          and eff["guide_version"] == gv_new and json.load(open(os.path.join(data, "rounds", "5", "version.json")))["guide_version"] == gv_old and open(os.path.join(data, "rounds", "1", "version.json"), "rb").read() == v1_before,
          (c_same.stderr[-120:], d_after_edit.stderr[-120:], c_ok.stderr[-160:], d_after.stderr[-160:], revs))
    wjl(os.path.join(data, "rounds", "5", "jev", "scores.jsonl"), [{"type": "header"}])
    gtext5 = open(gp5).read()
    open(gp5, "a").write("\nA second clarification after scoring.\n")
    c_late = vr("--round", "5", "--clarify", "too late")
    open(gp5, "w").write(gtext5)  # the refused edit is reverted: the round's recorded guide stays the current one
    vr("--round", "8", "--inputs")
    c8 = vr("--round", "8", "--clarify", "frozen")
    check("R2-01 a clarification is refused once the round has Jev scores and in the frozen rounds 8-10", c_late.returncode != 0 and "scores" in c_late.stderr and c8.returncode != 0 and "rounds 1-7" in c8.stderr, (c_late.stderr[-200:], c8.stderr[-200:]))
    # R2-03: a settings.json edit after the freeze is drift
    sp5 = os.path.join(data, "rounds", "5", "settings.json")
    s5 = json.load(open(sp5))
    json.dump(dict(s5, excerpt_lines=61), open(sp5, "w"))
    d_set = vr("--round", "5", "--drift")
    json.dump(s5, open(sp5, "w"))
    d_clean = vr("--round", "3", "--drift")
    check("R2-03 `versioning --drift` refuses a settings.json edited after the freeze (names `settings`) and accepts the frozen state", d_set.returncode != 0 and "settings" in d_set.stderr and d_clean.returncode == 0, (d_set.stderr[-200:], d_clean.stderr[-200:]))
    # before a round is frozen its proposed inputs are pinned live: an edit after the critique is refused
    vr("--round", "6", "--inputs")
    write_crit(5, x2d.inputs_digest(x2d.inputs_dir(6, data)))
    pre_ok = x2d.critique_problems(5, data, rev)
    open(os.path.join(x2d.questions_dir(6, data), "merge.covered@v1.json"), "a").write(" ")
    pre_edit = x2d.critique_problems(5, data, rev)
    check("R2-01 before round 6 is frozen the round-5 critique pins round 6's proposed inputs live: the approved digest passes, an edit of the proposed inputs after the critique is refused", pre_ok == [] and any(x2d.CRITIQUE_INPUTS_LINE.strip() in p for p in pre_edit), (pre_ok, pre_edit))
    # R2-16: a critique file that merely contains the wanted line (two Verdict lines, or a quoted draft line) is not accepted
    two = x2d.critique_path(3, rev)
    base_lines = open(x2d.critique_path(2, rev)).read()
    open(two, "w").write(base_lines.replace("critique\n", "critique\nVerdict: DESCRIPTIVE\n", 1))
    p_two = x2d.critique_problems(3, data, rev)
    open(two, "w").write(base_lines.replace("critique\n", "critique\nVerdict: ZERO ISSUES (draft: the drafter's proposal, quoted)\n", 1))
    p_quote = x2d.critique_problems(3, data, rev)
    check("R2-16 critique_problems rejects a file that merely contains the wanted line: two `Verdict:` lines (one of them the wrong phase) and a quoted draft line starting `Verdict:` are both refused (round 3 here lacks the matching version digest too, so the verdict problem is asserted by name)",
          any("exactly one line starting 'Verdict:'" in p for p in p_two) and any("exactly one line starting 'Verdict:'" in p for p in p_quote), (p_two, p_quote))


# ---------------------------------------------------------------- O onboarding Jev steps (contested files, duplicate check, freeze) on a stub endpoint
def onboarding_checks(tmp):
    if go_bin(tmp, "onboard", "./evaluation/x2d-merge-association-pilot/onboard") is None:
        check("O the onboarding program builds", False)
        return
    base = os.path.join(tmp, "onb")
    repo = Repo(os.path.join(base, "repo"))
    files = ["core/f%02d.py" % i for i in range(24)] + ["core/c_contested.py", "core/u_unowned.py"]
    c0 = repo.commit("onboarding base", {f: py_text(f) for f in files})
    coarse = [files[i:i + 6] for i in range(0, 24, 6)]
    draws = os.path.join(base, "draws")
    for k in (1, 2, 3):
        for j, g in enumerate(coarse):
            write_feature(draws, "core", k, "a%d_%d" % (j, k), "Planted area %d" % j, g + (["core/c_contested.py"] if (k == 1 and j == 0) else []))
        write_feature(draws, "core", k, "heur_%d" % k, "Heuristic group", ["core/f00.py"], source="python-source-group")
    cdir = os.path.join(base, "contrast")
    for j, g in enumerate(coarse):
        write_feature(os.path.join(cdir, "d"), "x", 1, "k%d" % j, "Contrast feature %d" % j, g)
    shutil.move(os.path.join(cdir, "d", "x", "1", "features"), os.path.join(cdir, "features"))
    json.dump({"rc": 0}, open(os.path.join(cdir, "meta.json"), "w"))
    sc = os.path.join(base, "scopes.json")
    write_scopes(sc, ["core"])
    ids = [hashlib.sha256(("x2d-area" + "core" + "\0".join(sorted(g))).encode()).hexdigest()[:12] for g in coarse]
    truth = {"core/c_contested.py": {"area": ids[0], "covered": True}, "core/u_unowned.py": {"area": None, "covered": False}, "_duplicates": [["Planted area 0", "Planted area 1"]]}
    log = os.path.join(base, "stub.log")
    stub = stub_jev.Stub(x2d.JEV_MODEL, truth=truth, p_correct=1.0, seed="o1", log_path=log, fail_ids={"core/c_contested.py"}).start()
    out = os.path.join(base, "catalogue")
    jout = os.path.join(base, "onboard-jev", "scores.jsonl")
    cmd = [_BIN["onboard"], "-draws", draws, "-scopes", sc, "-repo", repo.path, "-rev", c0, "-out", out, "-contrast", cdir, "-jev-out", jout, "-endpoint", stub.url, "-max-cost-usd", "5", "-workers", "4"]
    env = dict(clean_env(), OPENROUTER_API_KEY=KEY)
    try:
        r = run(cmd, cwd=ROOT, env=env)
        check("O a failed onboarding Jev request (HTTP 500) exits non-zero and writes NO catalogue: the catalogue is never frozen with unanswered requests", r.returncode in (3, 4) and not os.path.exists(os.path.join(out, "catalogue.json")), (r.returncode, r.stderr[-300:]))
        stub.fail_ids = set()
        r = run(cmd, cwd=ROOT, env=env)
        cat = json.load(open(os.path.join(out, "catalogue.json"))) if os.path.exists(os.path.join(out, "catalogue.json")) else {}
        check("O a re-run resumes the onboarding Jev output (answered requests are not asked again), completes, and freezes: catalogue.json, consensus-report.json, contrast-catalogue.json, catalogue.sha256",
              r.returncode == 0 and all(os.path.exists(os.path.join(out, n)) for n in ("catalogue.json", "consensus-report.json", "contrast-catalogue.json", "catalogue.sha256")), (r.returncode, r.stderr[-300:]))
        a0 = next((a for a in cat.get("areas", []) if "core/c_contested.py" in a["files"]), None)
        check("O Jev placed the contested file (covered, then the Choice over the scope's catalogue) in the planted area; the unowned file whose Noul is false stays unowned; the assignment is recorded in jev_assigned",
              a0 is not None and a0["area_id"] == ids[0] and any(x["path"] == "core/c_contested.py" and x["area_id"] == ids[0] for x in cat.get("jev_assigned", []))
              and all("core/u_unowned.py" not in a["files"] for a in cat["areas"]), cat.get("jev_assigned"))
        check("O catalogue.sha256 is the SHA-256 of catalogue.json as written (the catalogue digest); sample files are the diverse sample_files of the consensus-owned files (a Jev-assigned file is not a sample file)",
              open(os.path.join(out, "catalogue.sha256")).read().strip() == x2d.sha256_file(os.path.join(out, "catalogue.json")) and a0 is not None
              and a0["sample_files"] == sorted(set(a0["files"]) - {"core/c_contested.py"})[:8], a0["sample_files"] if a0 else None)
        rep = json.load(open(os.path.join(out, "consensus-report.json")))
        dups = json.dumps(rep)
        check("O the duplicate check (report only) covers pairs of clusters with title-token overlap and flags only the planted duplicate pair", dups.count('"duplicate": true') == 1 and dups.count('"duplicate": false') >= 3, dups[:300])
        reqs = [json.loads(ln) for ln in open(log)]
        covered_files = [q["state"]["file"]["path"] for q in reqs if q["type"] == "noul" and "file" in q["state"]]
        choice_files = [q["state"]["file"]["path"] for q in reqs if q["type"] == "choice" and "file" in q["state"]]
        check("O only contested and unowned files go to Jev: the covered read for both, the Choice only for the file whose covered read passed t_onboard_cov (0.7), never for a consensus-owned file",
              set(covered_files) == {"core/c_contested.py", "core/u_unowned.py"} and set(choice_files) == {"core/c_contested.py"} and stub.violations == [] and all(q["auth_present"] and not q["body_has_key"] for q in reqs), (set(covered_files), set(choice_files), stub.violations[:2]))
        cc = json.load(open(os.path.join(out, "contrast-catalogue.json")))
        check("O the contrast catalogue has one area per agent feature of the unconsolidated contrast draw (area ids sha256('x2d-contrast' + feature id)[:12], titles from the features)",
              len(cc["areas"]) == 4 and {a["area_id"] for a in cc["areas"]} == {hashlib.sha256(("x2d-contrast" + "k%d" % j).encode()).hexdigest()[:12] for j in range(4)}
              and {a["title"] for a in cc["areas"]} == {"Contrast feature %d" % j for j in range(4)}, [a["area_id"] for a in cc["areas"]])
        rows = jl(jout)
        check("O the onboarding Jev output records its header (model, thresholds), keeps its failure row (append-only), records every billed cost, and never holds the key",
              rows[0]["type"] == "header" and rows[0]["model"] == x2d.JEV_MODEL and rows[0]["t_onboard_cov"] == 0.7 and all(x["cost_known"] for x in rows if x["type"] == "answer")
              and sum(1 for x in rows if x["type"] == "failure") >= 1 and KEY not in open(jout).read(), rows[0])
        before = stub.requests
        r2 = run(cmd, cwd=ROOT, env=env)
        check("O onboarding refuses to run again over a frozen catalogue (append-only) and sends nothing", r2.returncode != 0 and stub.requests == before, r2.stderr[-200:])
    finally:
        stub.stop()


# ---------------------------------------------------------------- C the onboarding collector on a stub Clawpatch and a fake codex
def stub_build(tmp, name="stubbuild", marker=True, sanitizer=True):
    sb = os.path.join(tmp, name)
    os.makedirs(os.path.join(sb, "dist"))
    cli = os.path.join(sb, "dist", "cli.js")
    shutil.copy(os.path.join(HERE, "stub_clawpatch_x2d.js"), cli)
    if sanitizer:
        open(os.path.join(sb, "dist", "prompt.js"), "w").write("// fixture: %s\n" % C.SANITIZER_TOKEN)
    if marker:
        pv, ps = C.review_safety_pins()
        open(os.path.join(sb, C.BUILD_MARKER_FILE), "w").write("%s %s\n" % (pv, ps))
    return cli


def fake_codex_bin(tmp):
    bins = os.path.join(tmp, "fakebin")
    os.makedirs(bins, exist_ok=True)
    p = os.path.join(bins, "codex")
    open(p, "w").write("#!%s\nimport sys\nif sys.argv[1:2] == ['--version']:\n    print('codex-cli 0.0.0-fake')\n" % PY)
    os.chmod(p, 0o755)
    return bins


def collector_env(tmp, tag, fail_calls="", features=None, **extra):
    d = os.path.join(tmp, "col-" + tag)
    os.makedirs(d, exist_ok=True)
    env = clean_env(X2_STUB_COUNTER=d + "/counter", X2_STUB_FAIL_CALLS=fail_calls, X2_STUB_ENVLOG=d + "/envlog.jsonl", X2D_STUB_ARGVLOG=d + "/argv.jsonl",
                    X2D_BASE_STUB=os.path.join(x2d.X2_DIR, "validation", "stub_clawpatch.js"), **extra)
    if features is not None:
        env["X2D_STUB_FEATURES"] = str(features)
    env["PATH"] = fake_codex_bin(tmp) + os.pathsep + os.environ["PATH"]
    for n in (*C.TIMEOUT_ENV_NAMES, *C.REFUSED_MODEL_ENV, *C.REFUSED_SANDBOX_ENV, "CLAWPATCH_CLAUDE_BIN", "CODEX_HOME"):
        if n not in extra:
            env.pop(n, None)
    return env, d


def collect_run(tmp, tag, src, rev, scopes, env, d, cli=None, extra=(), data=None, work=None):
    cmd = [PY, os.path.join(S, "collect_onboard.py"), "--source", src, "--data", data or os.path.join(d, "data"), "--work", work or os.path.join(d, "work"), "--cli", cli or stub_build(tmp, "sb-" + tag),
           "--node", shutil.which("node"), "--rev", rev, "--scopes", scopes, "--expect-clawpatch-version", "stub-0", *extra]
    return run(cmd, cwd=ROOT, env=env)


def collector_checks(tmp):
    base = os.path.join(tmp, "colrepo")
    repo = Repo(base)
    files = {"core/a/x.py": py_text("cx"), "core/a/y.py": py_text("cy"), "core/b/z.py": py_text("cz"), "tut/t1.py": py_text("t1"), "tut/t2.py": py_text("t2"), "tut/t3.py": py_text("t3")}
    rev = repo.commit("collector base", files)
    scopes = os.path.join(tmp, "col-scopes.json")
    json.dump({"rev": rev, "all_fit": True, "scopes": [{"id": "core", "include": ["core"]}, {"id": "tut", "include": ["tut"]}]}, open(scopes, "w"))
    poison = {"OPENROUTER_API_KEY": "sk-or-v1-" + "poison", "ANTHROPIC_API_KEY": "sk-ant-" + "poison", "GITHUB_TOKEN": "gh" + "p_poison", "ANTHROPIC_BASE_URL": "http://127.0.0.1:1"}
    env, d = collector_env(tmp, "ok", features=2, **poison)
    r = collect_run(tmp, "ok", base, rev, scopes, env, d)
    argv = [json.loads(ln) for ln in open(d + "/argv.jsonl")] if os.path.exists(d + "/argv.jsonl") else []
    maps = [a for a in argv if "map" in a["args"]]
    inits = [a for a in argv if "init" in a["args"]]
    flags = ["map", "--source", "agent", "--provider", "codex", "--model", "gpt-6-luna", "--reasoning-effort", "xhigh"]
    check("C the collector runs 2 scopes x 3 draws + 1 contrast draw = 7 maps sequentially, each preceded by an init of a fresh state directory, with exactly `map --source agent --provider codex --model gpt-6-luna --reasoning-effort xhigh --json` and the state outside the workspace",
          r.returncode == 0 and len(maps) == 7 and len(inits) == 7 and all(a["args"][a["args"].index("map"):][:9] == flags for a in maps) and all("--json" in a["args"] for a in maps)
          and all(os.path.dirname(a["args"][a["args"].index("--state-dir") + 1]) != a["args"][a["args"].index("--root") + 1] for a in maps), (r.returncode, r.stderr[-400:], len(maps)))
    check("C the scope lever: each scoped map sees `include` rewritten to the scope's entries (core x3, tut x3) and the contrast draw keeps the default include; the config's codexConfig is never written by the collector",
          [a["include"] for a in maps] == [["core"]] * 3 + [["tut"]] * 3 + [["**"]] and not any("codexConfig" in json.dumps(a["include"]) for a in maps))
    dd = os.path.join(d, "data", "onboard")
    metas = {s: [json.load(open(os.path.join(dd, "draws", s, str(k), "meta.json"))) for k in (1, 2, 3)] for s in ("core", "tut")}
    feats = {s: [sorted(os.listdir(os.path.join(dd, "draws", s, str(k), "features"))) for k in (1, 2, 3)] for s in ("core", "tut")}
    own = {s: {o["path"] for fn in os.listdir(os.path.join(dd, "draws", s, "1", "features")) for o in json.load(open(os.path.join(dd, "draws", s, "1", "features", fn))).get("ownedFiles", []) if json.load(open(os.path.join(dd, "draws", s, "1", "features", fn)))["source"] == "agent-mapper"} for s in ("core", "tut")}
    check("C every draw is recorded under data/onboard/draws/<scope>/<k>/ (features/ and meta.json with rc 0, the scope's include, the pinned provider, model and effort); the contrast draw under contrast/; agent features own only files inside the scope's include",
          all(m["rc"] == 0 and m["include"] == [s] and m["map_model"] == "gpt-6-luna" and m["map_reasoning_effort"] == "xhigh" and m["provider"] == "codex" for s, ms in metas.items() for m in ms) and all(feats[s][k] for s in feats for k in range(3))
          and os.path.isdir(os.path.join(dd, "contrast", "features")) and all(p.startswith("core/") for p in own["core"]) and all(p.startswith("tut/") for p in own["tut"]) and own["core"] and own["tut"], (metas["core"][0], own))
    ver = json.load(open(os.path.join(dd, "versions.json")))
    envlog = [json.loads(ln) for ln in open(d + "/envlog.jsonl")]
    leaked = [n for e in envlog for n in e["names"] if C.is_credential(n)]
    check("C credentials are scrubbed from every Clawpatch call (no *API_KEY / OPENROUTER_* / *TOKEN / BASE_URL names reach the stub), Scout's host auth context and both provider timeouts are set; versions.json records names and presence, never values",
          not leaked and all(e["claudeAuth"] == "host" and e["codexTimeout"] == "900000" and e["claudeTimeout"] == "900000" for e in envlog if e["cmd"] == "map")
          and all(v not in json.dumps(ver) for v in poison.values()) and ver["clawpatch"] == "stub-0", (leaked, ver.get("env_names")))
    check("C attempts are kept under data/onboard/attempts/ (one per map) and the data is complete: 7 draw directories", len(os.listdir(os.path.join(dd, "attempts"))) == 7 and os.path.exists(os.path.join(dd, "versions.json")))
    r2 = collect_run(tmp, "ok2", base, rev, scopes, env, d, data=os.path.join(d, "data"), work=os.path.join(d, "work2"), cli=stub_build(tmp, "sb-ok2"))
    r3 = collect_run(tmp, "ok3", base, rev, scopes, env, d, data=os.path.join(d, "data2"), work=os.path.join(d, "work"), cli=stub_build(tmp, "sb-ok3"))
    check("C refusals before any clone or model call: existing onboarding data (append-only) and an existing workspace directory",
          r2.returncode != 0 and ("already" in (r2.stderr + r2.stdout).lower() or "refusing" in (r2.stderr + r2.stdout).lower())
          and r3.returncode != 0 and "workspace exists" in (r3.stderr + r3.stdout), (r2.stderr[-200:], r3.stderr[-200:]))
    env2, d2 = collector_env(tmp, "refuse", features=2, CLAWPATCH_MODEL="x")
    r4 = collect_run(tmp, "refuse", base, rev, scopes, env2, d2)
    env3, d3 = collector_env(tmp, "nocodex", features=2)
    env3["PATH"] = os.pathsep.join(p for p in env3["PATH"].split(os.pathsep) if p != os.path.join(tmp, "fakebin") and not shutil.which("codex", path=p))
    r5 = collect_run(tmp, "nocodex", base, rev, scopes, env3, d3)
    env4, d4 = collector_env(tmp, "unsafe", features=2)
    r6 = collect_run(tmp, "unsafe", base, rev, scopes, env4, d4, cli=stub_build(tmp, "sb-unsafe", marker=False))
    r7 = collect_run(tmp, "nosan", base, rev, scopes, env4, os.path.join(tmp, "col-nosan"), cli=stub_build(tmp, "sb-nosan", sanitizer=False), data=os.path.join(tmp, "col-nosan-data"), work=os.path.join(tmp, "col-nosan-w"))
    r8 = collect_run(tmp, "ver", base, rev, scopes, env4, os.path.join(tmp, "col-ver"), extra=("--expect-clawpatch-version", "9.9.9"), data=os.path.join(tmp, "col-ver-data"), work=os.path.join(tmp, "col-ver-w"), cli=stub_build(tmp, "sb-ver"))
    check("C refusals: CLAWPATCH_MODEL set in the environment; codex not on PATH; a Clawpatch build without the review-safety marker or without the sanitizer; a Clawpatch version other than the pin; none of them writes data",
          r4.returncode != 0 and "CLAWPATCH_MODEL" in r4.stderr + r4.stdout and r5.returncode != 0 and "codex" in r5.stderr + r5.stdout and r6.returncode != 0 and "review-safety" in r6.stderr + r6.stdout
          and r7.returncode != 0 and "review-safety" in r7.stderr + r7.stdout and r8.returncode != 0 and "9.9.9" in r8.stderr + r8.stdout and not os.path.exists(os.path.join(d4, "data", "onboard", "draws")), ((r4.stderr + r4.stdout)[-150:], (r5.stderr + r5.stdout)[-150:], (r8.stderr + r8.stdout)[-150:]))
    env5 = dict(env, X2D_SKIP_SIGNOFF="")
    env5.pop("X2D_SKIP_SIGNOFF")
    r9 = collect_run(tmp, "nosign", base, rev, scopes, env5, os.path.join(tmp, "col-nosign"), data=os.path.join(tmp, "col-nosign-data"), work=os.path.join(tmp, "col-nosign-w"), cli=stub_build(tmp, "sb-nosign"))
    check("C without the council's sign-off record the collector refuses before any clone or model call", r9.returncode != 0 and "sign-off" in (r9.stderr + r9.stdout), (r9.stderr + r9.stdout)[-200:])
    env6, d6 = collector_env(tmp, "retry", features=2, fail_calls="2")
    r = collect_run(tmp, "retry", base, rev, scopes, env6, d6)
    att = sorted(os.listdir(os.path.join(d6, "data", "onboard", "attempts")))
    m = json.load(open(os.path.join(d6, "data", "onboard", "draws", "core", "2", "meta.json")))
    check("C a failed map is retried once and both attempts are kept (the second draw's meta records attempt_count 2, rc 0); the run completes", r.returncode == 0 and m["attempt_count"] == 2 and m["rc"] == 0 and len(att) == 8, (r.returncode, att, r.stderr[-200:]))
    env7, d7 = collector_env(tmp, "stop", features=2, fail_calls="1,2,3,4,5,6")
    r = collect_run(tmp, "stop", base, rev, scopes, env7, d7)
    sp = os.path.join(d7, "data", "onboard", "stopped.json")
    st = json.load(open(sp)) if os.path.exists(sp) else {}
    check("C the stop rule: three consecutive failed maps stop the collection before the next map and write stopped.json naming the maps not attempted (unmeasured); the exit status is non-zero",
          r.returncode != 0 and "consecutive" in st.get("reason", "") and len(st.get("not_attempted", [])) == 4 and st.get("stopped_before") == "tut/1", (r.returncode, st))
    many = os.path.join(tmp, "col-many.json")
    json.dump({"rev": rev, "all_fit": True, "scopes": [{"id": "s%d" % i, "include": ["core"]} for i in range(6)]}, open(many, "w"))
    env8, d8 = collector_env(tmp, "cap", features=2)
    r = collect_run(tmp, "cap", base, rev, many, env8, d8)
    check("C the cap of 16 maps in round 0 is enforced before any map (6 scopes x 3 + 1 = 19 > 16)", r.returncode != 0 and ("16" in (r.stderr + r.stdout) or "cap" in (r.stderr + r.stdout).lower()) and not os.path.exists(os.path.join(d8, "data", "onboard", "draws")), (r.stderr + r.stdout)[-200:])


# ---------------------------------------------------------------- E end to end: round 0, all ten rounds and the verdict through run.sh on stubs
class E2E:
    """Drives run.sh stage by stage on the real fixture clone (git only) with a stub Clawpatch, an in-process stub Jev endpoint and stub labellers: no model call."""

    def __init__(self, tmp, p_correct=1.0):  # round 0's gates run on a noise-free stub (planted_new has only 3 items); e2e_rounds sets 0.95 for the rounds
        self.tmp = tmp
        self.data = os.path.join(tmp, "e2e-data")
        self.reviews = os.path.join(tmp, "e2e-reviews")  # X2D_REVIEWS_DIR: the governed S/reviews is never written by validation (R1-09)
        self.stub = stub_jev.Stub(x2d.JEV_MODEL, truth={}, p_correct=p_correct, seed="e2e", log_path=os.path.join(tmp, "e2e-jev.log")).start()
        self.env = clean_env(X2D_DATA=self.data, X2D_SRC=SRC, X2D_JEV_ENDPOINT=self.stub.url, OPENROUTER_API_KEY=KEY, X2D_REVIEWS_DIR=self.reviews)
        self.log = []

    def rs(self, *args, **extra_env):
        env = dict(self.env, **extra_env)
        r = run([os.path.join(S, "run.sh"), *map(str, args)], cwd=ROOT, env=env, timeout=3600)
        self.log.append((args, r.returncode, (r.stdout + r.stderr)[-600:]))
        return r

    def py(self, script, *args, **extra_env):
        """A study script (relative to S) with the stage environment (no run.sh) -> CompletedProcess."""
        r = run([PY, os.path.join(S, script), *map(str, args)], cwd=ROOT, env=dict(self.env, **extra_env), timeout=3600)
        self.log.append(((script,) + args, r.returncode, (r.stdout + r.stderr)[-600:]))
        return r

    def path(self, *p):
        return os.path.join(self.data, *p)

    def cat(self, contrast=False):
        return json.load(open(self.path("catalogue", "contrast-catalogue.json" if contrast else "catalogue.json")))

    def oracle(self, units):
        """Planted truth by the unit's first member path (the stub Jev's truth key and the stub labellers' truth): covered unless hash % 5 == 0, area = hash-chosen catalogue area."""
        areas = sorted(self.cat()["areas"], key=lambda a: a["area_id"])
        out = {}
        for u in units:
            p = u["members"][0]
            h = x2d.hash_int("e2e-oracle", p)
            a = areas[(h // 5) % len(areas)]
            out[u["case_id"]] = {"covered": h % 5 != 0, "area": a["area_id"], "area_title": a["title"], "path": p, "title": u["state"]["pr"]["title"]}
        return out

    def plant_jev(self, units, truth):
        for u in units:
            t = truth[u["case_id"]]
            self.stub.truth[t["path"]] = {"area": t["area"] if t["covered"] else None, "area_title": t["area_title"] if t["covered"] else None, "covered": t["covered"]}
            self.stub.truth["title:" + t["title"]] = self.stub.truth[t["path"]]

    def gold_of(self, t):
        return {"covered": t["covered"], "acceptable": [t["area"]] if t["covered"] else [], "primary": t["area"] if t["covered"] else None, "none_kind": None if t["covered"] else "gap",
                "rule": "stub", "reason": "stub"}

    def label_truth(self, units, truth, name, deviate_every=5, extra_keys=(), contrast=()):
        """X2D_LABEL_STUB / X2D_REFERENCE_STUB files: A = gold, B deviates (another acceptable area, same covered) for every 5th unit, C = gold. -> label stub path."""
        ids = [a["area_id"] for a in self.cat()["areas"]]
        t_all = {}
        for i, u in enumerate(sorted(units, key=lambda u: u["case_id"])):
            g = self.gold_of(truth[u["case_id"]])
            g["rule"] = "R1, R%d" % (2 + i % 4)  # R3-02: every session cites the general rule R1 beside a specific one (top-rule share 1.0): the concentration is reported, never a trigger, so the labels-check must still pass
            b = dict(g)
            if i % deviate_every == 0:
                if g["covered"]:
                    other = next(x for x in ids if x != g["acceptable"][0])
                    b.update(acceptable=[other], primary=other)
                else:
                    b.update(none_kind="new")
            t_all[u["case_id"]] = {"A": g, "B": b, "C": dict(g)}
            q1, q2 = {"covered": g["covered"]}, {"primary": g["primary"] or ids[0]}
            t_all[u["case_id"] + "|merge.covered"] = {"A": q1, "B": dict(q1), "C": dict(q1)}
            t_all[u["case_id"] + "|merge.which_area"] = {"A": q2, "B": dict(q2), "C": dict(q2)}
        cat_c = sorted(self.cat(True)["areas"], key=lambda a: a["area_id"])  # contrast labelling (cid "<case>|contrast") answers in the CONTRAST catalogue's area ids
        for i, u in enumerate(sorted(contrast, key=lambda u: u["case_id"])):
            cov = truth[u["case_id"]]["covered"]
            a = cat_c[(x2d.hash_int("e2e-oracle", u["members"][0]) // 5) % len(cat_c)]["area_id"]
            g = {"covered": cov, "acceptable": [a] if cov else [], "primary": a if cov else None, "none_kind": None if cov else "gap", "rule": "R1, R%d" % (2 + i % 4), "reason": "stub"}
            b = dict(g)
            if i % deviate_every == 0 and cov:
                other = next(x["area_id"] for x in cat_c if x["area_id"] != a)
                b.update(acceptable=[other], primary=other)
            t_all[u["case_id"] + "|contrast"] = {"A": g, "B": b, "C": dict(g)}
        p = os.path.join(self.tmp, "e2e-label-%s.json" % name)
        json.dump(t_all, open(p, "w"))
        return p

    def units_of(self, r, tag=None):
        return jl(self.path("rounds", str(r), "units%s.jsonl" % ("." + tag if tag else "")))

    def close(self):
        self.stub.stop()


def e2e_round0(tmp, e):
    if not os.path.isdir(SRC):
        check("E the end-to-end run needs the fastapi clone X2D_SRC", False, SRC)
        return False
    cli = stub_build(tmp, "e2e-sb")
    env, d = collector_env(tmp, "e2e", features=3)  # 3 stub agent features per scope draw: coarse enough that >= 28 P units have one unique owner (P1 needs 20 + P2 needs 8 gap positives, R2-05)
    r = run([PY, os.path.join(S, "collect_onboard.py"), "--source", SRC, "--data", e.data, "--work", os.path.join(tmp, "e2e-work"), "--cli", cli, "--node", shutil.which("node"),
             "--expect-clawpatch-version", "stub-0"], cwd=ROOT, env=dict(env, X2D_DATA=e.data), timeout=3600)
    draws = [os.path.join(e.data, "onboard", "draws", s, str(k)) for s in ("core", "tutorial-1", "tutorial-2", "tutorial-3", "infra") for k in (1, 2, 3)]
    check("E round 0 onboarding collection with the real scope partition on the real C0 tree and a stub Clawpatch: 15 scoped draws + the contrast draw are recorded, no failure", r.returncode == 0 and all(os.path.exists(os.path.join(p, "meta.json")) for p in draws)
          and os.path.isdir(os.path.join(e.data, "onboard", "contrast", "features")), (r.returncode, (r.stdout + r.stderr)[-400:]))
    if r.returncode != 0:
        return False
    r = e.rs("onboard")
    ok = r.returncode == 0 and os.path.exists(e.path("catalogue", "catalogue.sha256"))
    check("E run.sh onboard: code-first consensus per scope, onboarding Jev for any contested files, the freeze (catalogue.json, consensus-report.json, contrast-catalogue.json, catalogue.sha256)", ok, (r.stdout + r.stderr)[-500:])
    if not ok:
        return False
    cat = e.cat()
    check("E the catalogue digest on disk equals x2d.catalogue_digest() and the catalogue has areas in all five scopes, each with an opaque 12-hex id", open(e.path("catalogue", "catalogue.sha256")).read().strip() == x2d.catalogue_digest(e.data)
          and {a["scope"] for a in cat["areas"]} == {"core", "tutorial-1", "tutorial-2", "tutorial-3", "infra"} and all(re.fullmatch(r"[0-9a-f]{12}", a["area_id"]) for a in cat["areas"]), len(cat["areas"]))
    return True


def world(e, units, key=None):
    """Truth for a set of units: the hash oracle, then the control kind (round 0: the control key; rounds >= 1: the unit row's `control_kind`): p1 -> gold = the removed area;
    planted_new -> new; p2 (planted gap: the clone's area is ablated from its catalogue) -> gap; trap -> not covered, no kind; dup and prose clones keep their source's answer (same
    members). The source units of p2 clones are 'not covered' too (a clone shares its source's files, so both must agree)."""
    t = e.oracle(units)
    key = key or {}
    kind_of = lambda u: (key.get(u["case_id"]) or {}).get("kind") or u.get("control_kind")
    gap_members = {tuple(sorted(u["members"])) for u in units if kind_of(u) == "p2"}
    for u in units:
        k = key.get(u["case_id"]) or {}
        kind = kind_of(u)
        o = t[u["case_id"]]
        if kind == "p1":
            a = next(x for x in e.cat()["areas"] if x["area_id"] == k["area_id"])
            o.update(covered=True, area=a["area_id"], area_title=a["title"])
        elif kind in ("planted_new", "p2", "trap") or tuple(sorted(u["members"])) in gap_members:
            o.update(covered=False)
            o["none_kind"] = {"planted_new": "new", "trap": None}.get(kind, "gap")
    return t


def e2e_gates(tmp, e):
    """quality (build, label, finish) and the round-0 gates stage by stage on stubs."""
    r = e.rs("quality-build")
    check("E run.sh quality-build writes the quality report and the intruder items", r.returncode == 0 and os.path.exists(e.path("onboard", "quality.json")), (r.stdout + r.stderr)[-300:])
    items = jl(e.path("onboard", "intruder-items.jsonl"))
    keys = jl(e.path("onboard", "intruder-key.jsonl"))
    tr = {k["case_id"]: {s: {"intruder": k["intruder"], "reason": "stub"} for s in "ABC"} for k in keys}
    p = os.path.join(tmp, "e2e-intruder.json")
    json.dump(tr, open(p, "w"))
    rq_bad = e.rs("quality-label", "--catalogue", "/nonexistent.json", X2D_LABEL_STUB=p)
    check("E R3-13 `quality-label --catalogue X` is refused (R2-16's restriction extended): non-zero, names it, no intruder call row written",
          rq_bad.returncode != 0 and "forwards only" in (rq_bad.stdout + rq_bad.stderr) and not os.path.exists(e.path("onboard", "calls.intruder.jsonl")), (rq_bad.stdout + rq_bad.stderr)[-200:])
    r = e.rs("quality-label", X2D_LABEL_STUB=p)
    r2 = e.rs("quality-finish")
    q = json.load(open(e.path("onboard", "quality.intruder.json"))) if os.path.exists(e.path("onboard", "quality.intruder.json")) else {}
    check("E quality-label + quality-finish on stub labellers: every intruder is detected by the consensus, no unmeasured item", r.returncode == 0 and r2.returncode == 0 and q.get("unmeasured") == 0 and len(items) == len(keys), (r.stderr[-200:], r2.stderr[-200:], q))

    r = e.rs("gates-build")
    check("E run.sh gates-build: controls, K2/K3 inputs, round-0 version, provisional thresholds, baselines", r.returncode == 0 and os.path.exists(e.path("rounds", "0", "control-key.jsonl")), (r.stdout + r.stderr)[-400:])
    if r.returncode != 0:
        return False
    k23 = json.load(open(e.path("rounds", "0", "k23.json"))) if os.path.exists(e.path("rounds", "0", "k23.json")) else {}
    per = k23.get("per_round", {})
    frozen_s3 = sum((per.get(str(r_)) or {}).get("k3_stratum3", 0) + (per.get(str(r_)) or {}).get("new_dir", 0) + (per.get(str(r_)) or {}).get("core_gate", 0) for r_ in x2d.FROZEN_ROUNDS)
    print("INFO stratum-3 clusters of rounds 8-10 on the stub catalogue (the real-`none` proxy: INFO only, ~10 new-area cases cannot veto the verdict, R4-13; the real run reports this count in the round-0 critique): %d" % frozen_s3)
    check("E k23.json (R1-23/R1-24): K2 (B0 assigns) and its B2 share, K3 (stratum 3 on an existing C0 directory, upper bound) with new_dir and core_gate reported apart, per-round counts for rounds 1-10, core gate enabled flag",
          all(k in k23 for k in ("k2_share", "k2_b2_share", "k3_share", "new_dir", "core_gate", "per_round", "counts", "real_none_proxy", "frozen_pr_counts")) and sorted(map(int, per)) == list(range(1, 11)) and isinstance(k23["core_gate"].get("enabled"), bool), list(k23))
    pf, pa = k23["frozen_pr_counts"], k23["frozen_pr_counts_all_units"]
    check("E R3-11 k23.json `frozen_pr_counts` (the power structure) counts the verdict's population: core_gate units are excluded (its total = the frozen rounds' scored clusters minus their core_gate units); `frozen_pr_counts_all_units` keeps every scored cluster for the record",
          sum(sum(v.values()) for v in pf.values()) == sum(per[str(r_)]["n"] - per[str(r_)]["core_gate"] for r_ in x2d.FROZEN_ROUNDS) and sum(sum(v.values()) for v in pa.values()) == sum(per[str(r_)]["n"] for r_ in x2d.FROZEN_ROUNDS)
          and all(set(pf[r_]) <= set(pa[r_]) and all(pf[r_][p_] <= pa[r_][p_] for p_ in pf[r_]) for r_ in pf), (pf, pa, {r_: (per[str(r_)]["n"], per[str(r_)]["core_gate"]) for r_ in x2d.FROZEN_ROUNDS}))
    # R2-05 / R3-04 / rev 10: P2 sources = P leftovers, then D (hash order); n_p2 = min(12, eligible_P - 20 + eligible_D); refusal below 8; P1 unchanged
    import build_controls as BC
    full_res = json.load(open(os.path.join(S, "prepass", "reserve.json")))
    own0 = F.Ownership(G.catalogue_membership(e.cat(), 0))
    sett0 = x2d.round_settings(0, e.data)
    p_rows = BU.build_reserve_units(SRC, "P", own0, e.data, sett0, False, full_res)
    d_rows = BU.build_reserve_units(SRC, "D", own0, e.data, sett0, False, full_res)
    elig_ids = [r_["cluster_id"] for r_ in p_rows if BC._unique_owner(r_, own0)]
    elig_d_ids = [r_["cluster_id"] for r_ in d_rows if BC._unique_owner(r_, own0)]
    d_case_ids = {r_["case_id"] for r_ in d_rows}
    real_p2 = [k_ for k_ in jl(e.path("rounds", "0", "control-key.jsonl")) if k_["kind"] == "p2"]
    check("E rev 10 the real stub gates-build built n_p2 = min(12, eligible_P - 20 + eligible_D) = %d P2 clones (eligible_P %d, eligible_D %d), all gap positives" % (min(12, len(elig_ids) - 20 + len(elig_d_ids)), len(elig_ids), len(elig_d_ids)),
          len(real_p2) == min(12, len(elig_ids) - 20 + len(elig_d_ids)) and all(k_.get("gap_positive") for k_ in real_p2), len(real_p2))
    drop = lambda ids: dict(full_res, clusters_hash_order=[c for c in full_res["clusters_hash_order"] if c["id"] not in set(ids)])

    def p_only(res_):  # the legacy P-only build: no D unit is eligible, the gap-positive floor lifted so that a short P pool still builds
        orig_uo, orig_min = BC._unique_owner, BC.P2_MIN_GAP_POSITIVES
        BC._unique_owner = lambda row, own: None if row.get("set") == "D" else orig_uo(row, own)
        BC.P2_MIN_GAP_POSITIVES = 0
        try:
            return BC.make_controls(SRC, e.data, reserve=res_)
        finally:
            BC._unique_owner, BC.P2_MIN_GAP_POSITIVES = orig_uo, orig_min

    dump = lambda o: json.dumps(o, sort_keys=True)
    fill = base = None
    p_slots = 0
    for keep in range(min(27, len(elig_ids)), 19, -1):  # P short: 20 + p_slots eligible P units, p_slots < 8
        trimmed = drop(elig_ids[keep:])
        try:
            base = p_only(trimmed)
        except SystemExit:
            continue
        p_slots = keep - 20
        fill = BC.make_controls(SRC, e.data, reserve=trimmed)
        break

    def sig(res_, kinds):  # controls by (kind, source cluster id), case ids dropped
        kmap = {k_["case_id"]: k_ for k_ in res_["control_key"]}
        return {(u["control_kind"], u["cluster_id"]): (dump({a_: b_ for a_, b_ in u.items() if a_ != "case_id"}), dump({a_: b_ for a_, b_ in kmap[u["case_id"]].items() if a_ != "case_id"}))
                for u in res_["units"] if u.get("control") and u["control_kind"] in kinds}
    same_rows = fill is not None and base is not None
    if same_rows:
        real_f = {u["case_id"]: dump(u) for u in fill["units"] if not u.get("control")}
        real_b = {u["case_id"]: dump(u) for u in base["units"] if not u.get("control")}
        fixed_f, fixed_b = sig(fill, ("p1", "trap", "prose", "planted_new")), sig(base, ("p1", "trap", "prose", "planted_new"))
        dup_f, dup_b = sig(fill, ("dup",)), sig(base, ("dup",))
        small, large = (dup_f, dup_b) if len(dup_f) <= len(dup_b) else (dup_b, dup_f)
        same_rows = (real_f == real_b and fixed_f == fixed_b and len(fixed_f) == 20 + 2 + len(BC.PLANTED) and all(large.get(k_) == v_ for k_, v_ in small.items())
                     and dump(fill["catalogues"]["catalogue-p1"]) == dump(base["catalogues"]["catalogue-p1"]) and dump(fill["catalogues"]["catalogue-n1"]) == dump(base["catalogues"]["catalogue-n1"])
                     and dump(fill["contrast_units"]) == dump(base["contrast_units"]) and dump(fill["membership"]) == dump(base["membership"]))
    n_p1 = sum(1 for k_ in (fill or {"control_key": []})["control_key"] if k_["kind"] == "p1")
    fk = [k_ for k_ in (fill or {"control_key": []})["control_key"] if k_["kind"] == "p2"]
    check("E rev 10 P short, D fills (P pool 20 + %d eligible, %d eligible D): 12 clones, all membership gap positives, %d from P leftovers then %d from D in hash order; the D / P / P1 / trap / prose / planted_new rows, the dup rows common to both builds "
          "(a prefix of the same hash order), catalogue-p1, catalogue-n1, the contrast units and the membership are byte-identical to the P-only build of the same reserve; P1 stays exactly the P-only 20 items"
          % (p_slots, len(elig_d_ids), p_slots, 12 - p_slots),
          fill is not None and len(fk) == 12 and all(k_["gap_positive"] for k_ in fk) and fill["p2_donors"]["D"] == 12 - p_slots and fill["p2_donors"]["P_unplaceable"] + fill["p2_donors"]["P_leftover"] == p_slots
          and [k_["source_case_id"] in d_case_ids for k_ in fk].count(True) == 12 - p_slots and n_p1 == 20 and len(base["p2_source_ids"]) == p_slots and fill["p2_source_ids"][:p_slots] == base["p2_source_ids"]
          and fill["p2_source_ids"][p_slots:] == [r_["case_id"] for r_ in d_rows if r_["cluster_id"] in set(elig_d_ids)][:12 - p_slots] and same_rows,
          (None if fill is None else (len(fk), fill["p2_donors"], n_p1, same_rows)))
    refusal_pd, built8 = None, None
    if fill is not None:  # P + D short: p_slots + eligible_D = 7 refuses with the P-only text; = 8 builds 8 clones
        trimmed_p = drop(elig_ids[20 + p_slots:])
        for tgt in (8, 7):
            keep_d = tgt - p_slots
            trimmed = dict(trimmed_p, clusters_hash_order=[c for c in trimmed_p["clusters_hash_order"] if c["id"] not in set(elig_d_ids[keep_d:])])
            try:
                res_ = BC.make_controls(SRC, e.data, reserve=trimmed)
                if tgt == 8:
                    built8 = res_
            except SystemExit as ex:
                if tgt == 7:
                    refusal_pd = str(ex.code)
    check("E R2-05 / R3-04 / rev 10 build_controls refuses when eligible P beyond 20 plus eligible D leave fewer than %d P2 clones (7 of 8: %s); 8 clones build exactly at the floor" % (BC.P2_MIN_GAP_POSITIVES, (refusal_pd or "no refusal")[:80]),
          refusal_pd is not None and "planted-gap gate" in refusal_pd and "council and the user decide" in refusal_pd and refusal_pd.startswith("refusing: ") and built8 is not None
          and sum(1 for k_ in built8["control_key"] if k_["kind"] == "p2") == 8 and all(k_["gap_positive"] for k_ in built8["control_key"] if k_["kind"] == "p2"), (refusal_pd, None if built8 is None else built8["p2_donors"]))
    # R10-01 (alpha): on the real stub build every P1 item's gold area is an area of catalogue-p1, which keeps every area of the catalogue
    cp1_ids = {a_["area_id"] for a_ in json.load(open(e.path("rounds", "0", "catalogue-p1.json")))["areas"]}
    real_key = jl(e.path("rounds", "0", "control-key.jsonl"))
    real_p1 = [k_ for k_ in real_key if k_["kind"] == "p1"]
    check("E R10-01 (alpha) the real stub gates-build: %d P1 items, each gold area_id is an area of catalogue-p1, which has exactly as many areas as the catalogue (%d)" % (len(real_p1), len(e.cat()["areas"])),
          len(real_p1) >= 20 and all(k_["area_id"] in cp1_ids for k_ in real_p1) and len(cp1_ids) == len(e.cat()["areas"]), (len(real_p1), len(cp1_ids)))
    # R10-01 (beta): a P unit whose members are the only files of its area is not a P1 item (A would vanish from catalogue-p1); it goes to the P2 source pool as unplaceable
    real_p1_src = {k_["source_case_id"] for k_ in real_p1}
    cat_real = e.cat()
    beta = None
    cands = [r_ for r_ in p_rows if r_["case_id"] in real_p1_src and BC._unique_owner(r_, own0)]
    for r_ in cands[:6]:
        a_id = BC._unique_owner(r_, own0)
        mod = copy.deepcopy(cat_real)
        for a_ in mod["areas"]:
            if a_["area_id"] == a_id:
                a_["files"] = sorted(r_["members"])
                a_["sample_files"] = [f_ for f_ in a_.get("sample_files", []) if f_ in a_["files"]]
        orig_rc = x2d.read_catalogue
        x2d.read_catalogue = lambda data=None, _m=mod: copy.deepcopy(_m)
        try:
            res_b = BC.make_controls(SRC, e.data)
        except SystemExit:
            continue
        finally:
            x2d.read_catalogue = orig_rc
        b_p1 = [k_ for k_ in res_b["control_key"] if k_["kind"] == "p1"]
        b_ids = {a_["area_id"] for a_ in res_b["catalogues"]["catalogue-p1"]["areas"]}
        beta = (r_["case_id"], len(b_p1))
        if (r_["case_id"] in res_b["p2_source_ids"] and r_["case_id"] not in {k_["source_case_id"] for k_ in b_p1} and len(b_p1) == 20 and a_id in b_ids
                and len(b_ids) == len(mod["areas"]) and all(k_["area_id"] in b_ids for k_ in b_p1)):
            beta = True
            break
    check("E R10-01 (beta) planted: a P unit that is the sole owner-set of its area is not a P1 item, is a P2 source, P1 still has 20 items and catalogue-p1 keeps every area", beta is True, beta)
    units = e.units_of(0) + e.units_of(0, "contrast")
    key = {k["case_id"]: k for k in jl(e.path("rounds", "0", "control-key.jsonl"))}
    base_units = e.units_of(0)
    truth = world(e, base_units, key)
    for u in e.units_of(0, "contrast"):
        truth.setdefault(u["case_id"], truth.get(u["case_id"]) or e.oracle([u])[u["case_id"]])
    lp = e.label_truth(units, truth, "r0", contrast=e.units_of(0, "contrast"))
    rg_bad = e.rs("gates-label", "--catalogue", "/nonexistent.json", X2D_LABEL_STUB=lp)
    check("E R3-13 `gates-label --catalogue X` is refused (only --new-round, --workers, --cases are forwarded, as for `label`): non-zero, names the restriction, no label or call row written",
          rg_bad.returncode != 0 and "forwards only" in (rg_bad.stdout + rg_bad.stderr) and not any(os.path.exists(e.path("rounds", "0", f)) for f in ("labels.jsonl", "calls.jsonl", "labels.contrast.jsonl", "calls.contrast.jsonl")), (rg_bad.stdout + rg_bad.stderr)[-200:])
    r = e.rs("gates-label", X2D_LABEL_STUB=lp)
    check("E run.sh gates-label: stub labellers + adjudicator produce gold for the D/P units and the controls", r.returncode == 0 and os.path.exists(e.path("rounds", "0", "gold.jsonl")), (r.stdout + r.stderr)[-500:])
    if r.returncode != 0:
        return False
    lc0 = e.path("rounds", "0", "labels-check.json")
    lc0_ok = json.load(open(lc0)) if os.path.exists(lc0) else {}
    check("E run.sh gates-label writes rounds/0/labels-check.json carrying the round-0 guide version, ok (R2-03)", lc0_ok.get("ok") is True and lc0_ok.get("guide_version") == x2d.guide_version(0, e.data), lc0_ok)
    e2e_round0_gold_gap(tmp, e, lp)
    r4_checks.e2e_r4_round0(sys.modules[__name__], tmp, e, lp)
    shutil.move(lc0, lc0 + ".away")
    r = e.rs("gates-jev")
    shutil.move(lc0 + ".away", lc0)
    stale = dict(lc0_ok, guide_version="x2d-association-v1/guide-stale00000000")
    json.dump(stale, open(lc0, "w"))
    r_stale = e.rs("gates-jev")
    json.dump(lc0_ok, open(lc0, "w"))
    check("E R2-03 round-0 Jev (gates-jev) refuses without rounds/0/labels-check.json and with one for a stale guide version, and writes no Jev scores", r.returncode != 0 and "labels-check" in (r.stdout + r.stderr) and r_stale.returncode != 0
          and "labels-check" in (r_stale.stdout + r_stale.stderr) and not os.path.exists(e.path("rounds", "0", "jev", "scores.jsonl")), ((r.stdout + r.stderr)[-200:], (r_stale.stdout + r_stale.stderr)[-200:]))
    # Jev truth: every run reads the same planted world (the runs share first-member paths and titles by construction of the controls)
    e.plant_jev(base_units, truth)
    r = e.rs("gates-jev")
    check("E run.sh gates-jev: the manifest-driven round-0 Jev runs (main, N1, P1, P2; no gap run) and compose", r.returncode == 0 and os.path.exists(e.path("rounds", "0", "arms-p1", "V-sem.jsonl")) and os.path.exists(e.path("rounds", "0", "arms-p2", "V-sem.jsonl"))
          and not os.path.exists(e.path("rounds", "0", "arms-gap")), (r.stdout + r.stderr)[-500:])
    if r.returncode != 0:
        return False
    r = e.rs("gates-analyse")
    g = json.load(open(e.path("rounds", "0", "gates.json")))["gates"] if os.path.exists(e.path("rounds", "0", "gates.json")) else {}
    check("E run.sh gates-analyse writes gates.json (K2..K6, P1, planted_new, P2 = the planted-gap gate) with the core gate decided once", bool(g) and {"K2", "K3", "K4", "K5", "K6_N1", "P1", "planted_new", "P2"} <= set(g) and "planted_gap" not in g
          and isinstance(json.load(open(e.path("rounds", "0", "gates.json"))).get("core_gate", {}).get("enabled"), bool), (r.stdout + r.stderr)[-500:])
    for n in ("P1", "K6_N1", "planted_new", "P2", "K4", "K5"):
        check("E gate %s on the planted stub world passes" % n, (g.get(n) or {}).get("status") == "PASS", g.get(n))
    rnp = g.get("real_none_proxy") or {}
    check("E R2-11 gates.json records real_none_proxy (INFO, never a stop, no bound) equal to k23.json's", rnp.get("status") == "INFO" and rnp.get("value") == k23.get("real_none_proxy") and rnp.get("bound") is None, rnp)
    r = e.rs("power")
    pw = json.load(open(e.path("power.json"))) if os.path.exists(e.path("power.json")) else {}
    r_again = e.rs("power")
    cell_keys = ("P3_u0", "P3_u05", "P3_u10", "P3_PR", "P3_rho0", "P3_curve_delta_0.15", "N3", "N3_PR", "N4a", "N4b", "N4b_cap", "N4b_rho0", "N4c")
    check("E R2-11 / R4-13 / R5-01 run.sh power (after gates): P3 / N3 / N3-PR / N4 cells on the live frozen-round structure (rounds/0/k23.json) -> power.json schema x2d-power-v5 with the real_none_proxy, `checks` (the 9 binding cells, booleans), all_pass, "
          "p3_real_structure.share, p3_abstention (0.05 / 0.10), n4_max and the 13 cells; it exits 0 (no null cell above its bar on the stub world) and is append-only",
          r.returncode == 0 and pw.get("schema") == "x2d-power-v5" and pw.get("real_none_proxy") == k23.get("real_none_proxy") and all(k_ in pw for k_ in cell_keys)
          and set(pw.get("checks") or {}) == set(r4_checks.POWER_CHECK_KEYS) and all(isinstance(v_, bool) for v_ in pw["checks"].values()) and pw.get("all_pass") == all(pw["checks"].values())
          and "share" in pw.get("p3_real_structure", {}) and set(pw.get("p3_abstention") or {}) == {"0.05", "0.10"} and isinstance(pw.get("n4_max"), float)
          and pw["structure"]["clusters"] == sum(sum(v.values()) for v in k23["frozen_pr_counts"].values()) and r_again.returncode != 0, ((r.stdout + r.stderr)[-300:], pw.get("schema"), sorted(pw)))
    if pw:
        print("INFO live-structure power (rounds/0/k23.json, %d clusters / %d PRs): P3-u0 GO share %.3f; P3 at 5 %% / 10 %% unresolved %s; n4_max %.3f; binding checks %s; real_none_proxy %s" % (
            pw["structure"]["clusters"], pw["structure"]["prs"], pw["p3_real_structure"]["share"], pw.get("p3_abstention"), pw["n4_max"], pw["checks"], pw["real_none_proxy"]))
    return g


def write_critique(e, n, verdict=None, digest=None, nxt=None, lines=(), round0_lines=True):
    """The council critique of round n in the validation reviews dir: the verdict line of its phase (or `verdict`), the version digest line (or `digest`), for n <= 7 the digest of the
    inputs the next round will use (`nxt`, default: x2d.inputs_digest of the proposed inputs; n = 7: round 7's own), for n = 0 the recorded round-0 lines of r4_checks.round0_lines (unless
    round0_lines is False); `lines` are appended verbatim (e.g. a second Verdict line)."""
    os.makedirs(e.reviews, exist_ok=True)
    v = verdict or (x2d.CRITIQUE_VERDICT_REFINEMENT if n <= 7 else x2d.CRITIQUE_VERDICT_FROZEN)
    d = digest or x2d.sha256_file(e.path("rounds", str(n), "version.json"))
    out = ["validation critique of round %d" % n, "", v, x2d.CRITIQUE_VERSION_LINE + d]
    if n <= 7:
        out.append(x2d.CRITIQUE_INPUTS_LINE + (nxt or x2d.inputs_digest(x2d.inputs_dir(7, e.data) if n == 7 else x2d.proposed_inputs_dir(n + 1, e.data))))
    if n == 0 and round0_lines:
        out += r4_checks.round0_lines(sys.modules[__name__], e.data)  # R4-13 (7), R4-08: proxy (INFO), p3_real_structure, p3_abstention, n4_max, frozen_go_population, user_decision / p2_unmeasured when required
    with open(x2d.critique_path(n, e.reviews), "w") as f:
        f.write("\n".join(out + list(lines)) + "\n")


def refuses_units(e, rnd, what):
    r = e.rs("units", rnd)
    out = r.stdout + r.stderr
    check("E run.sh units %d refuses %s (R1-09): non-zero, names the critique file, writes no version.json" % (rnd, what),
          r.returncode != 0 and ("pilot-round-%d.md" % (rnd - 1)) in out and not os.path.exists(e.path("rounds", str(rnd), "version.json")), out[-300:])


def refusal_check(e, rnd, name, r, needle):
    out = r.stdout + r.stderr
    check("E round %d %s is refused (non-zero, names %r)" % (rnd, name, needle), r.returncode != 0 and needle in out, out[-300:])


def e2e_refusals_pre_label(e, rnd, lp):
    """R2-03 / R2-16 on the frozen round `rnd` before any label exists: a settings.json edited after the freeze, a missing previous critique, and forbidden `label` arguments are refused, nothing is written."""
    rs_ = str(rnd)
    st = e.path("rounds", rs_, "settings.json")
    s_ok = json.load(open(st))
    json.dump(dict(s_ok, excerpt_lines=s_ok["excerpt_lines"] + 1), open(st, "w"))
    r_set = e.rs("label", rnd, X2D_LABEL_STUB=lp)
    json.dump(s_ok, open(st, "w"))
    cp = x2d.critique_path(rnd - 1, e.reviews)
    os.replace(cp, cp + ".away")
    r_crit = e.rs("label", rnd, X2D_LABEL_STUB=lp)
    os.replace(cp + ".away", cp)
    r_cat = e.rs("label", rnd, "--catalogue", "/nonexistent.json", X2D_LABEL_STUB=lp)
    r_sta = e.rs("label", rnd, "--states", "/nonexistent.jsonl", X2D_LABEL_STUB=lp)
    refusal_check(e, rnd, "`label` after a settings.json edit post-freeze (R2-03, versioning --drift)", r_set, "settings")
    refusal_check(e, rnd, "`label` without the previous round's critique (R2-03)", r_crit, "pilot-round-%d.md" % (rnd - 1))
    refusal_check(e, rnd, "`label --catalogue` (R2-16: only --new-round, --workers, --cases are forwarded)", r_cat, "forwards only")
    refusal_check(e, rnd, "`label --states` (R2-16)", r_sta, "forwards only")
    check("E round %d the refused `label` calls wrote no labels" % rnd, not os.path.exists(e.path("rounds", rs_, "labels.jsonl")))


def e2e_refusals_post_label(e, rnd):
    """R2-03: after `label`, before `labels-check`: jev, gold, tune and analysis refuse (no labels-check.json); baselines refuse without gold."""
    for st in ("jev", "gold", "tune", "analysis"):
        refusal_check(e, rnd, "`%s` without rounds/%d/labels-check.json (R2-03)" % (st, rnd), e.rs(st, rnd), "labels-check")
    refusal_check(e, rnd, "`baselines` before gold (R1-08)", e.rs("baselines", rnd), "gold")
    check("E round %d the refused stages wrote nothing (no gold, no Jev scores, no thresholds, no results)" % rnd,
          not any(os.path.exists(e.path("rounds", str(rnd), f)) for f in ("gold.jsonl", "jev/scores.jsonl", "thresholds.json", "results.json")))


def e2e_refusal_stale_check(e, rnd):
    """R2-03: a labels-check of another guide version (stale) is refused by jev / gold."""
    p = e.path("rounds", str(rnd), "labels-check.json")
    good = json.load(open(p))
    json.dump(dict(good, guide_version="x2d-association-v1/guide-stale00000000"), open(p, "w"))
    rj, rg = e.rs("jev", rnd), e.rs("gold", rnd)
    json.dump(good, open(p, "w"))
    refusal_check(e, rnd, "`jev` with a labels-check of a stale guide version (R2-03)", rj, "labels-check")
    refusal_check(e, rnd, "`gold` with a labels-check of a stale guide version (R2-03)", rg, "labels-check")


def sha_of(*paths):
    return [x2d.sha256_file(p) if os.path.exists(p) else None for p in paths]


def e2e_clarify(e, rnd, lp, truth, units, tmp):
    """R2-01 (4): a mid-round guide clarification before scoring: `run.sh clarify` records the revision, relabels the round's wide cases under the new guide version (here one covered unit's
    answer changes), re-runs the labels-check; `gold` then uses the new rows (the LAST usable final row wins); earlier rounds are untouched; a second clarification is refused once gold exists."""
    rs_ = str(rnd)
    prev = [e.path("rounds", str(rnd - 1), f) for f in ("labels.jsonl", "gold.jsonl", "membership.json", "version.json")]
    before_prev = sha_of(*prev)
    ver_before = open(e.path("rounds", rs_, "version.json"), "rb").read()
    gp = x2d.guide_path(rnd, e.data)
    gv_old = x2d.guide_version(rnd, e.data)
    open(gp, "a").write("\nClarification (validation): a test belongs with the responsibility it tests, whatever its file name.\n")
    drift = e.py("versioning.py", "--round", rnd, "--drift")
    no_reason = e.rs("clarify", rnd)
    win = sorted((u for u in units if u.get("set") == "window" and not u.get("control") and truth[u["case_id"]]["covered"]), key=lambda u: u["case_id"])
    x = win[0]["case_id"]
    ids = [a["area_id"] for a in e.cat()["areas"]]
    old_area = truth[x]["area"]
    new_area = next(i for i in ids if i != old_area)
    t2 = json.load(open(lp))
    ans = {"covered": True, "acceptable": [new_area], "primary": new_area, "none_kind": None, "rule": "R1", "reason": "stub"}
    t2[x] = {"A": dict(ans), "B": dict(ans), "C": dict(ans)}
    lp2 = os.path.join(tmp, "e2e-label-clarified-%d.json" % rnd)
    json.dump(t2, open(lp2, "w"))
    r_cap = e.rs("clarify", rnd, "validation: relabel beyond the cap", X2D_LABEL_STUB=lp2, X2D_STUB_SESSION_CAP="1")
    check("E R3-01 round %d: a `clarify` whose relabel does not fit the session cap is refused BEFORE any revision is recorded (no guide-revisions.jsonl, no relabel rows, the round stays on its guide version)" % rnd,
          r_cap.returncode != 0 and "cannot be relabelled within the session cap" in (r_cap.stdout + r_cap.stderr) and not os.path.exists(e.path("rounds", rs_, "guide-revisions.jsonl"))
          and {x_.get("guide_version") for x_ in jl(e.path("rounds", rs_, "labels.jsonl")) if x_.get("kind") == "wide"} == {gv_old}, (r_cap.stdout + r_cap.stderr)[-300:])
    r = e.rs("clarify", rnd, "validation: test placement wording", X2D_LABEL_STUB=lp2)
    gv_new = x2d.guide_version(rnd, e.data)
    revs = jl(e.path("rounds", rs_, "guide-revisions.jsonl"))
    wide = [x_ for x_ in jl(e.path("rounds", rs_, "labels.jsonl")) if x_.get("kind") == "wide" and x_.get("label_set") == "main"]
    lc = json.load(open(e.path("rounds", rs_, "labels-check.json")))
    new_final = {x_["case_id"]: x_ for x_ in wide if x_.get("guide_version") == gv_new and x_.get("provenance") == "agent_reviewed" and x_.get("agreement") in ("agreed", "adjudicated")}
    check("E R2-01 round %d: an unrecorded guide edit is drift (versioning --drift refuses) and `clarify` without a reason is refused" % rnd, drift.returncode != 0 and no_reason.returncode != 0, (drift.stderr[-200:], no_reason.stdout[-200:] + no_reason.stderr[-200:]))
    check("E R2-01 round %d: run.sh clarify records the revision (guide-revisions.jsonl: the new guide version), relabels every wide case under it (rows of two guide versions on file), re-runs the labels-check for the new version and leaves version.json and the earlier rounds untouched" % rnd,
          r.returncode == 0 and len(revs) == 1 and revs[0]["guide_version"] == gv_new != gv_old and {x_.get("guide_version") for x_ in wide} == {gv_old, gv_new} and x in new_final
          and lc["guide_version"] == gv_new and lc["ok"] is True and open(e.path("rounds", rs_, "version.json"), "rb").read() == ver_before and sha_of(*prev) == before_prev, ((r.stdout + r.stderr)[-400:], revs, lc.get("guide_version")))
    sess = json.load(open(e.path("rounds", rs_, "sessions.json")))
    check("E R3-01 round %d: the clarified round runs under the REAL cap (no X2D_STUB_SESSION_CAP): sessions.json records C = 1 in the plan, the cap equals the formula with the clarification term, and the sessions used (labels, relabel, perception) fit it" % rnd,
          sess["plan"].get("C") == 1 and sess["cap"] == LX.cap_for(rnd, **sess["plan"]) and 0 < sess["used"] <= sess["cap"], sess)
    try:
        VR.clarify(rnd, "a second clarification", e.data)
        second = None
    except SystemExit as ex:
        second = str(ex.code)
    check("E R3-01 round %d: a second clarification is refused (C <= 1) and records nothing" % rnd, second is not None and "one recorded guide clarification" in second and len(jl(e.path("rounds", rs_, "guide-revisions.jsonl"))) == 1, second)
    r_g = e.rs("gold", rnd)
    gold_x = {g_["case_id"]: g_ for g_ in jl(e.path("rounds", rs_, "gold.jsonl"))}.get(x)
    old_final_x = [x_ for x_ in wide if x_["case_id"] == x and x_.get("guide_version") == gv_old and x_.get("agreement") in ("agreed", "adjudicated")]
    check("E R2-01 round %d: gold uses the relabelled rows of the new guide version (the LAST usable final row wins: unit %s moves from area %s to %s), not the first labelling's" % (rnd, x, old_area, new_area),
          r_g.returncode == 0 and gold_x is not None and gold_x["primary"] == new_area != old_area and old_final_x and old_final_x[-1]["primary"] == old_area and gold_x["acceptable"] == [new_area], ((r_g.stdout + r_g.stderr)[-300:], gold_x))
    r_c = e.rs("clarify", rnd, "again, after gold")
    refusal_check(e, rnd, "a second `clarify` once gold exists (R2-01)", r_c, "already has gold")


def e2e_label_gap(e, rnd, units, lp, truth, tmp):
    """R2-04 on round `rnd`: one unit with an unparsable adjudication and capped cases (a session cap of exactly 2 sessions per unit): `gold` refuses cleanly (no gold.jsonl, no next membership),
    baselines refuse, the divergence reports `unmeasured` with a reason; `label --new-round` fills the gap, then gold succeeds."""
    rs_ = str(rnd)
    t = json.load(open(lp))
    bad = sorted((u for u in units if u.get("set") == "window" and not u.get("control")), key=lambda u: u["case_id"])[0]["case_id"]
    t[bad] = {"A": None, "B": None, "C": None}
    lp_bad = os.path.join(tmp, "e2e-label-gap-%d.json" % rnd)
    json.dump(t, open(lp_bad, "w"))
    r = e.rs("label", rnd, X2D_LABEL_STUB=lp_bad, X2D_STUB_SESSION_CAP=str(2 * len(units)))
    calls = jl(e.path("rounds", rs_, "calls.jsonl"))
    capped = [c for c in calls if c.get("capped")]
    lab = jl(e.path("rounds", rs_, "labels.jsonl"))
    labelled = {x_["case_id"] for x_ in lab}
    unlabelled = [u_["case_id"] for u_ in units if u_["case_id"] not in labelled]
    check("E R2-04 round %d: with a session cap of 2 per unit the label stage stops non-zero with cases left unlabelled (a case capped mid-way has a call row only, no label rows; a case never started has neither)" % rnd,
          r.returncode != 0 and "session cap reached" in r.stderr and bool(unlabelled) and all(c["case_id"] not in labelled for c in capped), ((r.stdout + r.stderr)[-300:], len(capped), len(unlabelled)))
    rc = e.rs("labels-check", rnd)
    nxt = e.path("rounds", str(rnd + 1), "membership.json")
    r_g = e.rs("gold", rnd)
    r_b = e.rs("baselines", rnd)
    check("E R2-04 round %d: `gold` refuses cleanly when a window unit has no usable gold (names them, non-zero) and writes neither gold.jsonl nor the next round's membership.json; baselines refuse without gold" % rnd,
          rc.returncode == 0 and r_g.returncode != 0 and "nothing written" in (r_g.stdout + r_g.stderr) and not os.path.exists(e.path("rounds", rs_, "gold.jsonl")) and not os.path.exists(nxt) and r_b.returncode != 0,
          ((rc.stdout + rc.stderr)[-200:], (r_g.stdout + r_g.stderr)[-300:], (r_b.stdout + r_b.stderr)[-200:]))
    dv = A._divergence(rnd, e.data, ["V-sem", "B2"])
    try:
        G.replay_memberships(rnd, e.data)
        replay_refused = False
    except (ValueError, FileNotFoundError):
        replay_refused = True
    check("E R2-04 round %d: without gold the replayed divergence reports `unmeasured` with the reason for every arm (never a silent value) and replay_memberships refuses" % rnd,
          bool(dv["arms"]) and all(v["status"] == "unmeasured" and v["reason"] for v in dv["arms"].values()) and dv["status"] == "partly unmeasured" and replay_refused, dv)
    r_f = e.rs("label", rnd, "--new-round", "fill", X2D_LABEL_STUB=lp)
    r_stale = e.rs("gold", rnd)
    check("E R3-07 round %d: after the fill the earlier labels-check is stale (its labels_sha256 no longer matches the current-guide wide rows): `gold` is refused until `labels-check %d` is re-run, and writes nothing" % (rnd, rnd),
          r_stale.returncode != 0 and "changed after its labels-check" in (r_stale.stdout + r_stale.stderr) and not os.path.exists(e.path("rounds", rs_, "gold.jsonl")), (r_stale.stdout + r_stale.stderr)[-250:])
    lab2 = jl(e.path("rounds", rs_, "labels.jsonl"))
    final = {x_["case_id"]: x_ for x_ in lab2 if x_.get("kind") == "wide" and x_.get("provenance") == "agent_reviewed" and x_.get("agreement") in ("agreed", "adjudicated")}
    check("E R2-04 round %d: `label --new-round fill` labels the units without a usable final label (the capped ones and the unparsable one) and appends; every unit then has a usable final row" % rnd,
          r_f.returncode == 0 and all(u["case_id"] in final for u in units) and bad in final and any(x_.get("round") == "fill" for x_ in lab2), ((r_f.stdout + r_f.stderr)[-300:], len(final), len(units)))
    return r_f


def e2e_round0_gold_gap(tmp, e, lp):
    """R3-03 at round 0 (D/P units and the contrast set) on a copy of the data after `gates-label`: one D unit's wide labels and one contrast unit's labels are removed (a capped / unparsable case) and the
    gold files deleted: `gold.py --round 0` and `--tag contrast` refuse and write nothing; `label --new-round fill` fills both; both golds then succeed."""
    d = os.path.join(tmp, "e2e-r0gap")
    shutil.copytree(e.data, d)
    rd = os.path.join(d, "rounds", "0")
    for f in ("gold.jsonl", "gold.contrast.jsonl"):
        os.remove(os.path.join(rd, f))
    d_unit = next(u["case_id"] for u in jl(os.path.join(rd, "units.jsonl")) if u.get("set") == "D" and not u.get("control"))
    c_unit = jl(os.path.join(rd, "units.contrast.jsonl"))[0]["case_id"]
    wjl(os.path.join(rd, "labels.jsonl"), [x_ for x_ in jl(os.path.join(rd, "labels.jsonl")) if x_["case_id"] != d_unit])
    wjl(os.path.join(rd, "labels.contrast.jsonl"), [x_ for x_ in jl(os.path.join(rd, "labels.contrast.jsonl")) if x_["case_id"] != c_unit])
    r_main = e.py("gold.py", "--round", 0, X2D_DATA=d)
    r_con = e.py("gold.py", "--round", 0, "--tag", "contrast", X2D_DATA=d)
    check("E R3-03 round 0: `gold.py --round 0` refuses (non-zero, nothing written) when a D unit has no usable gold, and the contrast tag likewise: independent of any next membership",
          r_main.returncode != 0 and "nothing written" in (r_main.stdout + r_main.stderr) and d_unit in (r_main.stdout + r_main.stderr) and r_con.returncode != 0 and "nothing written" in (r_con.stdout + r_con.stderr)
          and c_unit in (r_con.stdout + r_con.stderr) and not os.path.exists(os.path.join(rd, "gold.jsonl")) and not os.path.exists(os.path.join(rd, "gold.contrast.jsonl")), ((r_main.stdout + r_main.stderr)[-200:], (r_con.stdout + r_con.stderr)[-200:]))
    f1 = e.py("label_x2d.py", "--round", 0, "--new-round", "fill", X2D_DATA=d, X2D_LABEL_STUB=lp)
    f2 = e.py("label_x2d.py", "--round", 0, "--tag", "contrast", "--new-round", "fill", X2D_DATA=d, X2D_LABEL_STUB=lp)
    g1 = e.py("gold.py", "--round", 0, X2D_DATA=d)
    g2 = e.py("gold.py", "--round", 0, "--tag", "contrast", X2D_DATA=d)
    gold_d = {g_["case_id"] for g_ in jl(os.path.join(rd, "gold.jsonl"))} if os.path.exists(os.path.join(rd, "gold.jsonl")) else set()
    gold_c = {g_["case_id"] for g_ in jl(os.path.join(rd, "gold.contrast.jsonl"))} if os.path.exists(os.path.join(rd, "gold.contrast.jsonl")) else set()
    check("E R3-03 round 0: `label --new-round fill` (wide and contrast) labels the removed units and `gold` then succeeds for both files (D unit %s and contrast unit %s have gold)" % (d_unit, c_unit),
          f1.returncode == 0 and f2.returncode == 0 and g1.returncode == 0 and g2.returncode == 0 and d_unit in gold_d and c_unit in gold_c, [(x_.returncode, (x_.stdout + x_.stderr)[-150:]) for x_ in (f1, f2, g1, g2)])


def e2e_rounds(tmp, e, g):
    """All ten rounds stage by stage (units -> label -> labels-check -> gold -> baselines -> jev -> tune -> analysis, rescore-v7 before round 8) on the unmodified integrated pipeline, then
    the verdict: the gates file and the critique files are enforced, only the named K3 gate is overridden, the planted world is recovered, controls / contrast / sessions / divergence /
    end-to-end are exercised, standalone stages refuse without their prerequisites, a clarification, a gap in the labels, a planted null and a flipped primary arm are exercised."""
    rev_dir = os.path.join(S, "reviews")
    reviews_before = sorted(os.listdir(rev_dir)) if os.path.isdir(rev_dir) else []
    gate_fail = sorted(k for k, v in (g or {}).items() if v.get("status") == "FAIL")
    refuses_units(e, 1, "without the council critique file of round 0")
    write_critique(e, 0, round0_lines=False)
    refuses_units(e, 1, "a round-0 critique lacking the real_none_proxy / p3_real_structure lines (R2-11)")
    write_critique(e, 0, lines=["Verdict: ZERO ISSUES (draft quoted from the drafter, not the council's)"])
    refuses_units(e, 1, "a round-0 critique with a second line starting 'Verdict:' (R2-16: a file that merely contains the wanted line)")
    write_critique(e, 0)
    d_env = os.path.join(tmp, "e2e-env1")  # R3-12: a catalogue byte changed after the gates is refused at `units 1` by the envelope check (round 1 against round 0), nothing written (a copy with the gates forced to PASS, so the refusal is the envelope's)
    shutil.copytree(e.data, d_env)
    gdoc = json.load(open(os.path.join(d_env, "rounds", "0", "gates.json")))
    for gv_ in gdoc["gates"].values():
        if gv_.get("status") == "FAIL":
            gv_["status"] = "PASS"
    json.dump(gdoc, open(os.path.join(d_env, "rounds", "0", "gates.json"), "w"))
    cat_env = os.path.join(d_env, "catalogue", "catalogue.json")
    open(cat_env, "a").write("\n")
    r_env = run([os.path.join(S, "run.sh"), "units", "1"], cwd=ROOT, env=dict(e.env, X2D_DATA=d_env), timeout=3600)
    check("E R3-12 `units 1` refuses a catalogue changed after the gates (the refinement-envelope check now runs for every r >= 1): non-zero, names the catalogue digest, writes no version.json and no units",
          r_env.returncode != 0 and "catalogue digest" in (r_env.stdout + r_env.stderr) and not os.path.exists(os.path.join(d_env, "rounds", "1", "version.json")) and not os.path.exists(os.path.join(d_env, "rounds", "1", "units.jsonl")),
          (r_env.stdout + r_env.stderr)[-300:])
    shutil.rmtree(d_env)
    r = e.rs("units", 1)
    first_units = r if r.returncode == 0 else None  # no gate failed: this run IS round 1's units stage
    check("E run.sh units 1 refuses while a round-0 gate FAILs (K3 on the coarse stub catalogue is expected to) or proceeds when none does", (r.returncode != 0) == bool(gate_fail), (gate_fail, (r.stdout + r.stderr)[-300:]))
    other = [k for k in gate_fail if k != "K3"]
    check("E every round-0 gate except K3 passes on the stub world: ONLY the named K3 gate (new directories of the coarse stub catalogue) is overridden, never a blanket override of failed gates (R1-20)", not other, other)
    if other:
        return False
    if gate_fail:
        gp = e.path("rounds", "0", "gates.json")
        doc = json.load(open(gp))
        doc["gates"]["K3"]["status"] = "PASS"
        doc["gates"]["K3"]["note"] = "validation-only override of the stub-catalogue K3 FAIL: " + doc["gates"]["K3"].get("note", "")
        json.dump(doc, open(gp, "w"))
        print("      (validation override of the stub-catalogue K3 gate)")
    prim = x2d.primary_arm(e.data)
    e.stub.p_correct = 0.95  # the planted effect of the rounds: stub V-sem 0.95 against the real B2
    res_all, ok = {}, True
    for rnd in range(1, 11):
        rs_ = str(rnd)
        if rnd == 2:
            os.makedirs(e.path("rounds", "2"), exist_ok=True)
            json.dump(dict(x2d.DEFAULT_SETTINGS, excerpt_lines=30), open(e.path("rounds", "2", "settings.json"), "w"))  # R3-05: round 2 is built at 30 excerpt lines (menu item (ii)); the v_7 re-score re-renders its units under round 7's 60
            refuses_units(e, 2, "without the council critique file of round 1")
            write_critique(e, 1, digest="0" * 64)
            refuses_units(e, 2, "a critique whose version_sha256 is not the digest of rounds/1/version.json")
            write_critique(e, 1, verdict="Verdict: APPROVED")
            refuses_units(e, 2, "a critique without the line 'Verdict: ZERO ISSUES'")
            write_critique(e, 1, lines=["Verdict: ZERO ISSUES"])
            refuses_units(e, 2, "a critique with two 'Verdict: ZERO ISSUES' lines (R2-16)")
            write_critique(e, 1, nxt="0" * 64)
            refuses_units(e, 2, "a critique whose next_inputs_sha256 does not pin the inputs round 2 uses (R2-01)")
        if rnd == 9:
            write_critique(e, 8, verdict=x2d.CRITIQUE_VERDICT_REFINEMENT)
            refuses_units(e, 9, "the refinement verdict where the frozen phase needs 'Verdict: DESCRIPTIVE' (critique of round 8)")
        if rnd >= 2:
            write_critique(e, rnd - 1)
        if rnd == 8:
            r = e.rs("units", 8)
            check("E run.sh units 8 refuses until rounds 0-7 are re-scored under v_7 (run.sh rescore-v7, R2-02) and writes no version.json", r.returncode != 0 and "v_7" in (r.stdout + r.stderr) and not os.path.exists(e.path("rounds", "8", "version.json")), (r.stdout + r.stderr)[-300:])
            cp7 = x2d.critique_path(7, e.reviews)
            os.replace(cp7, cp7 + ".away")
            r = e.rs("rescore-v7")
            os.replace(cp7 + ".away", cp7)
            check("E run.sh rescore-v7 refuses without the council critique of round 7 (v_7 is frozen by it) and writes nothing", r.returncode != 0 and "pilot-round-7.md" in (r.stdout + r.stderr) and not os.path.exists(e.path("rounds", "0", "jev-v7")), (r.stdout + r.stderr)[-300:])
            r = e.rs("rescore-v7")
            done = all(os.path.exists(e.path("rounds", str(rr), x2d.V7_RESCORE_SUB, ".done")) and os.path.exists(e.path("rounds", str(rr), x2d.V7_RESCORE_SUB, "scores.jsonl")) for rr in range(8))
            try:
                dev_ok = TU.dev_rounds(8, e.data, x2d.V7_RESCORE_SUB) == list(range(8))
            except SystemExit as ex:
                dev_ok = str(ex)
            r2 = e.rs("rescore-v7")
            check("E run.sh rescore-v7 (R2-02): rounds 0-7 re-scored under round 7's questions, settings and model with the round-7 header (tune.dev_rounds accepts all eight files); the re-score is complete once and refused a second time",
                  r.returncode == 0 and done and dev_ok is True and r2.returncode != 0 and "complete" in (r2.stdout + r2.stderr), ((r.stdout + r.stderr)[-400:], done, dev_ok, (r2.stdout + r2.stderr)[-200:]))
            if r.returncode != 0:
                return False
            hdr7 = TU._header(e.path("rounds", "7", "jev", "scores.jsonl"))
            hdr0 = TU._header(e.path("rounds", "0", x2d.V7_RESCORE_SUB, "scores.jsonl"))
            check("E the round-0 v_7 re-score carries exactly the three question versions of the sem and fan arms (merge.covered, merge.which_area, merge.area_covers, each as <question>/V-sem), each round 7's, and round 7's model and catalogue digest, in its header (R3-18)", bool(hdr0) and set(hdr0.get("versions") or {}) == {q_ + "/V-sem" for q_ in TU.V7_QUESTIONS} and all((hdr7.get("versions") or {}).get(q) == v for q, v in hdr0["versions"].items()) and all(hdr0.get(k) == hdr7.get(k) for k in ("model", "catalogue_sha256")), (hdr0, hdr7))
            n7 = int(x2d.round_settings(7, e.data)["excerpt_lines"])
            old2 = {u_["case_id"]: u_ for u_ in e.units_of(2) if u_.get("set") == "window" and not u_.get("control")}
            new2 = jl(e.path("rounds", "2", x2d.V7_RESCORE_SUB, "units.jsonl"))
            exp_ex = lambda u_: "\n".join(BU.read_text(SRC, u_["commit"], next((p_ for p_ in sorted(u_["members"]) if BU.read_text(SRC, u_["commit"], p_).strip()), sorted(u_["members"])[0])).splitlines()[:n7])
            grew = [u_["case_id"] for u_ in new2 if len(u_["state"]["unit"]["excerpt"].splitlines()) > 30]
            check("E R3-05 round 2 was built at excerpt_lines 30 and its v_7 re-score units are re-rendered under round 7's %d (the 120-line-vs-30 case of the council): every row carries excerpt_lines %d and the excerpt equals the first %d lines of the representative member at the unit's commit, "
                  "%d unit(s) grew past 30 lines, ids and membership unchanged; the round's own units.jsonl keeps its 30-line excerpts" % (n7, n7, n7, len(grew)),
                  bool(new2) and {u_["case_id"] for u_ in new2} == set(old2) and all(u_["excerpt_lines"] == n7 and u_["state"]["unit"]["excerpt"] == exp_ex(u_) and u_["members"] == old2[u_["case_id"]]["members"] for u_ in new2)
                  and all(len(o_["state"]["unit"]["excerpt"].splitlines()) <= 30 for o_ in old2.values()) and bool(grew), (len(new2), len(old2), grew[:3]))
        r = first_units if (rnd == 1 and first_units is not None) else e.rs("units", rnd)
        ok = r.returncode == 0 and os.path.exists(e.path("rounds", rs_, "units.jsonl"))
        check("E run.sh units %d: version freeze (and envelope check), units + per-round controls on the real window, the contrast units" % rnd, ok, (r.stdout + r.stderr)[-500:])
        if not ok:
            return False
        units = e.units_of(rnd)
        ctr = e.units_of(rnd, "contrast")
        cw = e.units_of(rnd, "contrast_window")
        win = [u for u in units if u.get("set") == "window" and not u.get("control")]
        ctl = [u for u in units if u.get("control")]
        kinds = sorted(u.get("control_kind") for u in ctl)
        check("E round %d: per-round controls (R1-07): ceil(10 %%) duplicates of window units, one empty-rename trap, one misleading-prose clone, all `control` rows with ids <r>:ctl-<n>, never in the window set" % rnd,
              len(win) > 0 and kinds.count("dup") == math.ceil(0.10 * len(win)) and kinds.count("trap") == 1 and kinds.count("prose") == 1 and len(kinds) == len(ctl)
              and all(u["set"] == "ctl" and u["case_id"].startswith("%d:ctl-" % rnd) for u in ctl), (len(win), kinds))
        d_units = [u for u in ctr if u.get("set") == "D"]
        d0 = [u for u in e.units_of(0) if u.get("set") == "D"]
        check("E round %d: units.contrast.jsonl holds the round's window units restratified (every round); the round-0 D units (%d) are added in rounds 1 and 10 only (R2-06); units.contrast_window.jsonl (<= 30 hash-sampled window units) exists in rounds 1 and 10 only (R1-13)" % (rnd, len(d0)),
              len(ctr) - len(d_units) == len(win) and all(u["set"] == "window" for u in ctr if u.get("set") != "D") and ((len(d_units) == len(d0) > 0) if rnd in x2d.CONTRAST_ROUNDS else not d_units)
              and (rnd in x2d.CONTRAST_ROUNDS) == bool(cw) and len(cw) <= 30 and {u["case_id"] for u in cw} <= {u["case_id"] for u in win}, (len(ctr), len(d_units), len(win), len(cw)))
        cnt_path = e.path("rounds", rs_, "counts.json")
        truth = world(e, units)
        for u in ctr + cw:
            truth[u["case_id"]] = truth.get(u["case_id"]) or e.oracle([u])[u["case_id"]]
        lp = e.label_truth(units, truth, "r%d" % rnd, contrast=cw)
        if rnd == 3:
            e2e_refusals_pre_label(e, rnd, lp)
        if rnd in (4, 10):  # R2-04 / R3-03: a capped case and an unparsable adjudication leave the round without complete gold (round 10 is 48 of the 91 verdict clusters)
            r = e2e_label_gap(e, rnd, units, lp, truth, tmp)
        else:
            r = e.rs("label", rnd, X2D_LABEL_STUB=lp)
        ok = r.returncode == 0 and os.path.exists(e.path("rounds", rs_, "labels.jsonl")) and os.path.exists(e.path("rounds", rs_, "states.jsonl")) and os.path.exists(e.path("rounds", rs_, "sessions.json"))
        check("E run.sh label %d: wide + jev-state labels (from the dumped states.jsonl); gold is its own stage, not written by `label`; contrast window labelled in rounds 1 and 10 only" % rnd,
              ok and (rnd == 4 or not os.path.exists(e.path("rounds", rs_, "gold.jsonl"))) and (rnd in x2d.CONTRAST_ROUNDS) == os.path.exists(e.path("rounds", rs_, "labels.contrast.jsonl")), (r.stdout + r.stderr)[-500:])
        if not ok:
            return False
        if rnd == 3:
            e2e_refusals_post_label(e, rnd)
        r = e.rs("labels-check", rnd)
        check("E run.sh labels-check %d passes (rule concentration <= 0.40 and kappa_answer >= 0.6 on the stub labellers) and writes labels-check.json for the round's guide version" % rnd,
              r.returncode == 0 and json.load(open(e.path("rounds", rs_, "labels-check.json")))["guide_version"] == x2d.guide_version(rnd, e.data), (r.stdout + r.stderr)[-300:])
        if rnd == 3:
            e2e_refusal_stale_check(e, rnd)
            r_tg = e.rs("tune", rnd)
            check("E R3-14 round %d: `tune` without gold.jsonl is refused (names it) and writes no thresholds.json or scores.csv" % rnd,
                  r_tg.returncode != 0 and "gold.jsonl" in (r_tg.stdout + r_tg.stderr) and not any(os.path.exists(e.path("rounds", rs_, f)) for f in ("thresholds.json", "scores.csv")), (r_tg.stdout + r_tg.stderr)[-200:])
            e2e_clarify(e, rnd, lp, truth, units, tmp)
        else:
            if rnd == 8:
                refusal_check(e, 8, "`clarify` in a frozen round (rounds 1-7 only, R2-01)", e.rs("clarify", 8, "frozen"), "rounds 1-7 only")
            r = e.rs("gold", rnd)
            check("E run.sh gold %d: gold.jsonl and the next round's membership.json%s; contrast gold in rounds 1 and 10 only" % (rnd, "" if rnd < x2d.N_ROUNDS else " (none after round 10)"),
                  r.returncode == 0 and os.path.exists(e.path("rounds", rs_, "gold.jsonl")) and (rnd == x2d.N_ROUNDS or os.path.exists(e.path("rounds", str(rnd + 1), "membership.json")))
                  and (rnd in x2d.CONTRAST_ROUNDS) == os.path.exists(e.path("rounds", rs_, "gold.contrast.jsonl")), (r.stdout + r.stderr)[-500:])
            if r.returncode != 0:
                return False
        r = e.rs("baselines", rnd)
        check("E run.sh baselines %d (after gold, R1-08): B0/B1/B2 against the membership at each unit's parent commit, the round-start B2 sensitivity, and contrast-B2 (every round, R1-13)" % rnd,
              r.returncode == 0 and all(os.path.exists(e.path("rounds", rs_, "arms", b + ".jsonl")) for b in ("B0", "B1", "B2", "b2_round_start", "contrast-B2")), (r.stdout + r.stderr)[-300:])
        if rnd == 3:  # R3-14: the standalone stages refuse without their data inputs, before any output
            r_an = e.rs("analysis", rnd)
            r_tu = e.rs("tune", rnd)
            check("E R3-14 round %d: `analysis` without arms/reference.jsonl and jev/scores.jsonl is refused (names the missing file) and writes no results.json / report.md / arms of the composed variants; `tune` without jev/scores.jsonl is refused and writes no thresholds.json" % rnd,
                  r_an.returncode != 0 and "arms/reference.jsonl" in (r_an.stdout + r_an.stderr) and r_tu.returncode != 0 and "jev/scores.jsonl" in (r_tu.stdout + r_tu.stderr)
                  and not any(os.path.exists(e.path("rounds", rs_, f)) for f in ("results.json", "report.md", "thresholds.json", "arms/V-sem.jsonl")), ((r_an.stdout + r_an.stderr)[-200:], (r_tu.stdout + r_tu.stderr)[-200:]))
        e.plant_jev(units, truth)
        ref = {c: {"covered": t["covered"], "area_id": t["area"] if t["covered"] else None} for c, t in truth.items()}
        rp = os.path.join(tmp, "e2e-ref-%d.json" % rnd)
        json.dump(ref, open(rp, "w"))
        r = e.rs("jev", rnd, X2D_REFERENCE_STUB=rp, X2D_LABEL_STUB=lp)
        check("E run.sh jev %d: V-sem, V-sib, V-title, fan-out (contrast V-sem in rounds 1 and 10) and the reference arm" % rnd, r.returncode == 0 and os.path.exists(e.path("rounds", rs_, "jev", "scores.jsonl"))
              and (rnd in x2d.CONTRAST_ROUNDS) == os.path.exists(e.path("rounds", rs_, "jev", "scores.contrast.jsonl")), (r.stdout + r.stderr)[-500:])
        if r.returncode != 0:
            return False
        ref_rows = jl(e.path("rounds", rs_, "arms", "reference.jsonl"))
        check("E R3-01 round %d: the reference arm completed under the real cap (%d clusters, none unresolved for `session_cap`)" % (rnd, len(ref_rows)), bool(ref_rows) and not any(x_.get("unresolved_reason") == "session_cap" for x_ in ref_rows), len(ref_rows))
        r = e.rs("tune", rnd)
        th = json.load(open(e.path("rounds", rs_, "thresholds.json"))) if os.path.exists(e.path("rounds", rs_, "thresholds.json")) else {}
        check("E run.sh tune %d (prior rounds only)%s" % (rnd, {8: ": tuned once on the v_7 re-scored rows", 9: ": reuses round 8's thresholds", 10: ": reuses round 8's thresholds"}.get(rnd, "")), r.returncode == 0
              and (rnd not in (8, 9, 10) or ((th.get("dev_scores_sub") == x2d.V7_RESCORE_SUB) if rnd == 8 else (th.get("reused_from") == 8))), (r.stdout + r.stderr)[-300:])
        r = e.rs("analysis", rnd)
        res = json.load(open(e.path("rounds", rs_, "results.json"))) if os.path.exists(e.path("rounds", rs_, "results.json")) else None
        pack = os.path.join(e.reviews, "pilot-round-%d-pack.md" % rnd)
        check("E run.sh analysis %d: compose, results.json with intervals, the critique pack (in X2D_REVIEWS_DIR)" % rnd, r.returncode == 0 and res is not None and os.path.exists(pack), (r.stdout + r.stderr)[-500:])
        if res is None:
            return False
        res_all[rnd] = res
        vs = res["arms"][prim]["real"]
        check("E round %d: the primary arm (%s, stub p_correct 0.95) recovers the planted world: P_assign and R_assign >= 0.6 on measured clusters" % (rnd, prim),
              vs["P_assign"]["n"] > 0 and vs["P_assign"]["value"] >= 0.6 and vs["R_assign"]["n"] > 0 and vs["R_assign"]["value"] >= 0.6, {k: vs[k] for k in ("P_assign", "R_assign")})
        b2 = res["arms"]["B2"]["real"]
        check("E round %d B2 is measured on the same clusters" % rnd, b2["accuracy"]["n"] == vs["accuracy"]["n"], (b2["accuracy"], vs["accuracy"]))
        lk = jl(e.path("rounds", rs_, "lookups.jsonl"))
        owned = sum((row.get("counts") or {}).get("modified_owned", 0) for row in lk)
        unowned = sum((row.get("counts") or {}).get("modified_unowned", 0) for row in lk)
        ee = res["end_to_end"]
        check("E round %d end-to-end (R1-12): only modified_owned files count as exact lookups (%d), modified_unowned files (%d) are reported in no accuracy" % (rnd, owned, unowned),
              ee["rule_exact_files"] == owned and ee["modified_unowned"].get("modified_unowned", 0) == unowned, (ee["rule_exact_files"], owned, ee.get("modified_unowned"), unowned))
        total = json.load(open(cnt_path)) if os.path.exists(cnt_path) else None
        have = collections.Counter()
        for row in lk:
            have.update({k: int(v) for k, v in (row.get("counts") or {}).items()})
        check("E round %d R2-07: rounds/%d/counts.json exists and the lookups rows sum to it key by key (a modification-only commit's counts are not dropped); the analysis ran its equality assertion" % (rnd, rnd),
              total is not None and {k: v for k, v in have.items() if v} == {k: int(v) for k, v in total.items() if int(v)} and sum(int(v) for v in total.values()) > 0, (dict(have), total))
        dv = res["divergence"]
        check("E round %d replayed free-running divergence is MEASURED for every arm (the two-membership call, R1-12), not 'unmeasured'" % rnd,
              dv["name"] == "replayed free-running divergence" and dv["arms"] and all(v["status"] == "measured" and 0 <= v["value"]["mean"] <= 1 for v in dv["arms"].values()), dv)
        sp = res.get("session_plan") or {}
        formula = LX.cap_for(rnd, **sp) if sp else None
        check("E round %d results.json records session_cap, sessions_used and session_plan (R1-05, R3-01): the cap is the FORMULA (%s; C = %s recorded clarification) with no stub override, and the sessions the stub world used (%s) fit it with no allowance" % (rnd, formula, (sp or {}).get("C"), res.get("sessions_used")),
              sp and sp.get("C") == (1 if rnd == 3 else 0) and res.get("session_cap") == formula and isinstance(res.get("sessions_used"), int) and 0 < res["sessions_used"] <= res["session_cap"], (res.get("session_cap"), res.get("sessions_used"), sp, formula))
        con = res.get("contrast") or {}
        dpart = (con.get("parts") or {}).get("D") or {}
        in_c = rnd in x2d.CONTRAST_ROUNDS
        check("E round %d reports the contrast section (every round): contrast-B2 on all %d contrast units; the D part is non-empty and scored in rounds 1 and 10 only (R2-06); the window part is scored where contrast gold exists (the window sample in rounds 1 and 10, none otherwise)" % (rnd, len(ctr)),
              con.get("n_units") == len(ctr) and ((con.get("n_scored", 0) > 0) == in_c) and "contrast-B2" in json.dumps(con) and (dpart.get("n_units", 0) > 0) == in_c and (dpart.get("n_scored", 0) > 0) == in_c,
              ({k: con.get(k) for k in ("n_units", "n_scored")}, {k: dpart.get(k) for k in ("n_units", "n_scored")}))
        ctl_res = res.get("controls") or {}
        report_txt, pack_txt = open(e.path("rounds", rs_, "report.md")).read(), open(pack).read()
        n_dup = kinds.count("dup")
        check("E round %d R2-10 control figures on the stub world (results.json, report.md, pack): dup gold agreement with the source %d/%d, dup arm output equality (n > 0), trap gold-not-covered and arm none rate (n 1), prose arm equality (n 1)" % (rnd, n_dup, n_dup),
              ctl_res.get("dup", {}).get("n") == n_dup and ctl_res["dup"]["gold_covered_agrees"]["n"] == n_dup and ctl_res["dup"]["gold_covered_agrees"]["k"] == n_dup and ctl_res["dup"]["arm_output_equals_source"]["n"] > 0
              and ctl_res["trap"]["gold_not_covered"]["n"] == 1 and ctl_res["trap"]["arm_none"]["n"] == 1 and ctl_res["prose"]["arm_output_equals_source"]["n"] == 1
              and "Controls (report-only, R2-10)" in report_txt and "Controls (test-retest" in pack_txt and "arm_output_equals_source" in report_txt and "arm_output_equals_source" in pack_txt, ctl_res)
    after = sorted(os.listdir(rev_dir)) if os.path.isdir(rev_dir) else []
    check("E validation never writes the governed %s (critiques and packs went to X2D_REVIEWS_DIR)" % rev_dir, after == reviews_before, (reviews_before, after))
    e2e_omitted_rows(tmp, e)
    e2e_go_population_kappa(tmp, e, res_all)
    e2e_failed_kappa(tmp, e)
    r4_checks.e2e_r4_rounds(sys.modules[__name__], tmp, e)
    return e2e_verdict(tmp, e, prim, res_all)


def verdict_json(path):
    return json.load(open(os.path.join(path, "verdict.json"))) if os.path.exists(os.path.join(path, "verdict.json")) else None


def reanalyse(tmp_data, reviews, rounds=(8, 9, 10)):
    """Recompute the frozen rounds' results (delete results.json / report.md, run analysis --round) and the verdict in a copy of the finished run. -> returncode of the last step."""
    env = clean_env(X2D_DATA=tmp_data, X2D_REVIEWS_DIR=reviews)
    for r in rounds:
        for fn in ("results.json", "report.md"):
            p = os.path.join(tmp_data, "rounds", str(r), fn)
            if os.path.exists(p):
                os.remove(p)
        x = run([PY, os.path.join(S, "analysis_x2d.py"), "--round", str(r)], cwd=ROOT, env=env)
        if x.returncode != 0:
            return x
    vp = os.path.join(tmp_data, "verdict.json")
    if os.path.exists(vp):
        os.remove(vp)
    return run([PY, os.path.join(S, "analysis_x2d.py"), "--verdict"], cwd=ROOT, env=env)


def e2e_verdict(tmp, e, prim, res_all):
    """The verdict through run.sh on the planted-effect world, then the planted null and the flipped primary arm re-analysed from copies."""
    r = e.rs("verdict")
    check("E run.sh verdict refuses without the council critique of round 10 (R1-09)", r.returncode != 0 and "pilot-round-10.md" in (r.stdout + r.stderr) and verdict_json(e.data) is None, (r.stdout + r.stderr)[-300:])
    write_critique(e, 10, verdict=x2d.CRITIQUE_VERDICT_REFINEMENT)
    r = e.rs("verdict")
    check("E run.sh verdict refuses a round-10 critique that carries the refinement verdict instead of 'Verdict: DESCRIPTIVE'", r.returncode != 0 and verdict_json(e.data) is None, (r.stdout + r.stderr)[-300:])
    write_critique(e, 10)
    null_dir, flip_dir, redraw_dir = os.path.join(tmp, "e2e-null"), os.path.join(tmp, "e2e-flip"), os.path.join(tmp, "e2e-redraw")
    shutil.copytree(e.data, null_dir)
    shutil.copytree(e.data, flip_dir)
    shutil.copytree(e.data, redraw_dir)
    common_dir = os.path.join(tmp, "e2e-common")
    shutil.copytree(e.data, common_dir)
    r = e.rs("verdict")
    v = verdict_json(e.data)
    check("E run.sh verdict writes verdict.json on the pooled frozen rounds 8-10 (one evaluation)", r.returncode == 0 and v is not None, (r.stdout + r.stderr)[-400:])
    if v is None:
        return False
    pooled = v["pooled"]
    thr = json.load(open(os.path.join(S, "thresholds.json")))
    recomputed = A.go_verdict(pooled, thr)
    cond_fail = [k for k, c in v["go_verdict"]["conditions"].items() if not c["pass"]]
    print("INFO planted effect verdict on the stub world (primary %s, stub V-sem p_correct 0.95 vs real B2): %s; failing conditions: %s; gain_res point %s (n_res %s); gain_all %s; u %s; P_assign %s; R_assign_res %s; real positives %s (P_new %s, R_new %s); perception %s" %
          (prim, v["verdict"], cond_fail, pooled["gain_res"]["point"], pooled["n_res"], pooled["gain_all"]["point"], pooled["unresolved"]["rate"], pooled["p_assign"]["point"], pooled["r_assign_res"]["point"], pooled["real_new"]["n"], pooled["real_new"].get("P_new"), pooled["real_new"].get("R_new"), json.dumps(v["perception"])[:300]))
    check("E R2-12 planted effect (stub V-sem 0.95 vs the real B2 on the full pipeline) -> the verdict is GO (every one of the 12 conditions holds: informative rounds, P_assign / R_assign_res points and bounds, gain_res >= +0.10 with a positive bound, spread over >= 3 PRs, u <= 0.15, gain_guess 95 % bound > 0 (guard h), n_res >= 50, perception, N1); the stored verdict equals the pure rule on its pooled summary",
          v["verdict"] == "GO" == recomputed["verdict"] and not cond_fail and len(v["go_verdict"]["conditions"]) == 12 and pooled["gain_res"]["point"] >= 0.10 and pooled["gain_guess"]["lb95"] > 0 and pooled["n_res"] >= 50, (v["verdict"], cond_fail, pooled["gain_res"], pooled["real_new"], v["perception"]))
    check("E the verdict's primary arm is the one K5 chose (%s) and every frozen round's results.json agrees" % prim, v["primary_arm"] == prim and pooled["jev_arm"] == prim and all(res_all[r_]["primary_arm"] == prim for r_ in x2d.FROZEN_ROUNDS), v["primary_arm"])
    loo = v.get("leave_one_round_out") or {}
    check("E R3-15 verdict.json carries leave_one_round_out for each of rounds 8, 9, 10 (descriptive, no verdict): the pooled figures over the other two frozen rounds, whose cluster counts add up to twice the pooled count",
          set(loo) == {"8", "9", "10"} and all(loo[str(r_)]["left_out"] == r_ and loo[str(r_)]["rounds"] == [q for q in x2d.FROZEN_ROUNDS if q != r_] and "gain_res" in loo[str(r_)]["pooled"] and "verdict" not in loo[str(r_)] for r_ in x2d.FROZEN_ROUNDS)
          and sum(loo[str(r_)]["pooled"]["n"] for r_ in x2d.FROZEN_ROUNDS) == 2 * pooled["n"], {k_: (x_.get("rounds"), x_["pooled"]["n"]) for k_, x_ in loo.items()} if loo else None)
    # R2-12 planted null: the primary arm's frozen-round correctness is redrawn per cluster, independently, at B2's realised accuracy (seeded); the raw answers stay (so perception is unaffected)
    rng = random.Random(x2d.SEED)
    area_ids = [a["area_id"] for a in e.cat()["areas"]]
    b2c = [(k["case_id"], k["out"]["B2"]["correct"]) for r_ in x2d.FROZEN_ROUNDS for k in A.build_cases(A.scored_units(A.load_round(r_, redraw_dir)["units"]), A.load_round(r_, redraw_dir)["gold"], {"B2": A.load_round(r_, redraw_dir)["arms"]["B2"]}) if k["out"]["B2"]["correct"] is not None]
    acc_b2 = sum(1 for _c, ok_ in b2c if ok_) / len(b2c)
    n_redrawn = n_ok = 0
    for r_ in x2d.FROZEN_ROUNDS:
        Lr = A.load_round(r_, redraw_dir)
        scored = {u["case_id"] for u in A.scored_units(Lr["units"])}
        out_rows = []
        for row in sorted(Lr["arms"][prim], key=lambda x: x["case_id"]):
            gd = Lr["gold"].get(row["case_id"])
            if row["case_id"] in scored and gd and isinstance(gd.get("covered"), bool):
                good = rng.random() < acc_b2
                n_redrawn += 1
                n_ok += good
                if gd["covered"]:
                    right = gd.get("primary") or (gd["acceptable"] or [None])[0]
                    wrong = next(a for a in area_ids if a not in (gd["acceptable"] or []))
                    row = dict(row, output="assigned", area_id=right if good else wrong, unresolved_reason=None)
                else:
                    row = dict(row, output="none", area_id=None, unresolved_reason=None) if good else dict(row, output="assigned", area_id=area_ids[0], unresolved_reason=None)
            out_rows.append(row)
        wjl(os.path.join(redraw_dir, "rounds", str(r_), "arms", prim + ".jsonl"), out_rows)
    x = reanalyse(redraw_dir, e.reviews)
    vr_ = verdict_json(redraw_dir)
    print("INFO redrawn primary null (B2 accuracy %.3f, %d clusters redrawn, %d correct): verdict %s, gain_res %s, P_assign %s" % (acc_b2, n_redrawn, n_ok, vr_ and vr_["verdict"], vr_ and vr_["pooled"]["gain_res"]["point"], vr_ and vr_["pooled"]["p_assign"]["point"]))
    check("E R2-12 planted null (the primary arm's frozen-round correctness redrawn per cluster, independently, at B2's realised accuracy %.3f; seeded random.Random(%d)): the verdict is not GO" % (acc_b2, x2d.SEED),
          x.returncode == 0 and vr_ is not None and vr_["verdict"] != "GO" and n_redrawn >= 80, (x.returncode, x.stderr[-300:], vr_ and (vr_["verdict"], vr_["pooled"]["gain_res"]["point"])))
    # R3-08 common-accuracy null: BOTH arms' frozen-round correctness redrawn per cluster, independently, at p = 0.95 (seeded): P_assign passes its point, so the verdict is decided by the gain conditions
    def redraw_arm(d_, arm, p, rng_):
        n_ = 0
        for r_ in x2d.FROZEN_ROUNDS:
            Lr = A.load_round(r_, d_)
            scored = {u["case_id"] for u in A.scored_units(Lr["units"])}
            out_rows = []
            for row in sorted(Lr["arms"][arm], key=lambda x: x["case_id"]):
                gd = Lr["gold"].get(row["case_id"])
                if row["case_id"] in scored and gd and isinstance(gd.get("covered"), bool):
                    good = rng_.random() < p
                    n_ += 1
                    if gd["covered"]:
                        right = gd.get("primary") or (gd["acceptable"] or [None])[0]
                        wrong = next(a for a in area_ids if a not in (gd["acceptable"] or []))
                        row = dict(row, output="assigned", area_id=right if good else wrong, unresolved_reason=None)
                    else:
                        row = dict(row, output="none", area_id=None, unresolved_reason=None) if good else dict(row, output="assigned", area_id=area_ids[0], unresolved_reason=None)
                out_rows.append(row)
            wjl(os.path.join(d_, "rounds", str(r_), "arms", arm + ".jsonl"), out_rows)
        return n_
    rng2 = random.Random(x2d.SEED)
    n_prim, n_b2 = redraw_arm(common_dir, prim, 0.95, rng2), redraw_arm(common_dir, "B2", 0.95, rng2)
    xc = reanalyse(common_dir, e.reviews)
    vcn = verdict_json(common_dir)
    ccond = (vcn or {}).get("go_verdict", {}).get("conditions", {})
    gain_fail = [k for k, c in ccond.items() if k.startswith("gain") and not c["pass"]]
    print("INFO common-accuracy null (both arms redrawn at 0.95, %d + %d clusters): verdict %s, P_assign %s, gain_res %s, failing gain conditions %s" % (n_prim, n_b2, vcn and vcn["verdict"], vcn and vcn["pooled"]["p_assign"]["point"], vcn and vcn["pooled"]["gain_res"]["point"], gain_fail))
    check("E R3-08 common-accuracy null (primary arm and B2 both redrawn per cluster, independently, at p = 0.95; seeded random.Random(%d)): the P_assign point condition PASSES, the verdict is not GO and a gain condition (`gain point >= +0.10` / its 90 %% bootstrap lower bound) is among the failing ones: the gain conditions are exercised, not only the P_assign point" % x2d.SEED,
          xc.returncode == 0 and vcn is not None and n_prim >= 80 and n_b2 == n_prim and ccond.get("P_assign point >= 0.90", {}).get("pass") is True and vcn["verdict"] != "GO" and bool(gain_fail), (xc.returncode, xc.stderr[-300:], vcn and (vcn["verdict"], vcn["pooled"]["p_assign"]["point"], gain_fail)))
    # sanity: the primary arm's frozen-round rows are replaced by B2's -> exactly zero paired gain, never GO
    for r_ in x2d.FROZEN_ROUNDS:
        rd = os.path.join(null_dir, "rounds", str(r_))
        wjl(os.path.join(rd, "arms", prim + ".jsonl"), [dict(x, arm=prim) for x in jl(os.path.join(rd, "arms", "B2.jsonl"))])
    x = reanalyse(null_dir, e.reviews)
    vn = verdict_json(null_dir)
    check("E copy-B2 sanity null (the primary arm answers exactly as B2 in rounds 8-10): zero paired gain (resolved and all-cluster) and the verdict is not GO", x.returncode == 0 and vn is not None and vn["verdict"] != "GO" and abs(vn["pooled"]["gain_res"]["point"]) < 1e-12 and abs(vn["pooled"]["gain_all"]["point"]) < 1e-12,
          (x.returncode, x.stderr[-300:], vn and (vn["verdict"], vn["pooled"]["gain_res"]["point"])))
    # flipped K5: the primary arm switches and the verdict follows it (R1-03)
    gp = os.path.join(flip_dir, "rounds", "0", "gates.json")
    doc = json.load(open(gp))
    doc["fan_out_primary"] = prim != x2d.FAN_OUT_ARM
    json.dump(doc, open(gp, "w"))
    other = x2d.FAN_OUT_ARM if prim != x2d.FAN_OUT_ARM else x2d.VSEM_ARM
    x = reanalyse(flip_dir, e.reviews)
    vf = verdict_json(flip_dir)
    rf = [json.load(open(os.path.join(flip_dir, "rounds", str(r_), "results.json"))) for r_ in x2d.FROZEN_ROUNDS] if x.returncode == 0 else []
    check("E flipping K5 (fan_out_primary) re-points everything at the other arm (%s): verdict.primary_arm, the pooled summary and every frozen round's results.json follow it (R1-03)" % other,
          x.returncode == 0 and vf is not None and vf["primary_arm"] == other and vf["pooled"]["jev_arm"] == other and all(q["primary_arm"] == other for q in rf), (x.returncode, x.stderr[-300:], vf and vf["primary_arm"]))
    check("E R2-17 with the fan-out as primary arm perception still compares V-sem's raw answer with the labels (the headline agreement is V-sem's, unchanged; the primary arm's agreement is reported beside it)",
          vf is not None and vf["perception"]["primary_arm"] == other and vf["perception"]["jev_agree"] == v["perception"]["jev_agree"] and vf["perception"]["primary_agree"] is not None, (vf and vf["perception"], v["perception"]))
    e2e_abstention_worlds(tmp, e, prim)
    return True


def frozen_go_clusters(d_):
    """[(round, case_id, pr_key)] of the scored GO-population clusters of rounds 8-10 with a usable gold row in data dir d_, sorted by case id."""
    out = []
    for r_ in x2d.FROZEN_ROUNDS:
        Lr = A.load_round(r_, d_)
        for u in A.scored_units(Lr["units"]):
            gd = Lr["gold"].get(u["case_id"])
            if u.get("stratum_reason") != "core_gate" and gd and isinstance(gd.get("covered"), bool):
                out.append((r_, u["case_id"], A._pr_key(u)))
    return sorted(out, key=lambda t: t[1])


def rewrite_arm(d_, arm, decide, area_ids, abstain_raw="keep"):
    """Rewrite the frozen rounds' composed rows of `arm` in data dir d_: decide(case_id) -> "right" | "wrong" | "abstain" | None (unchanged). right = an acceptable area (gold covered) or `none` (gold not covered);
    wrong = another area (covered) or an assignment (not covered); abstain = unresolved (below_t_choice). Rows of other clusters are written as they were. R5-04: the raw (unthresholded) fields are set consistently
    with `decide`: a right row carries the gold answer as its raw guess, a wrong row a wrong area (a wrong `assigned` on gold none), an abstained row by `abstain_raw` ("keep" = the stub's own raw fields,
    "right", "wrong"; or a callable case_id -> one of those)."""
    n = 0
    for r_ in x2d.FROZEN_ROUNDS:
        Lr = A.load_round(r_, d_)
        scored = {u["case_id"] for u in A.scored_units(Lr["units"]) if u.get("stratum_reason") != "core_gate"}
        out_rows = []
        for row in sorted(Lr["arms"][arm], key=lambda x: x["case_id"]):
            gd = Lr["gold"].get(row["case_id"])
            what = decide(row["case_id"]) if row["case_id"] in scored else None
            if what and gd and isinstance(gd.get("covered"), bool):
                n += 1
                raw_kind = what
                if what == "abstain":
                    row = dict(row, output="unresolved", area_id=None, unresolved_reason="below_t_choice")
                    raw_kind = abstain_raw(row["case_id"]) if callable(abstain_raw) else abstain_raw
                elif gd["covered"]:
                    right = gd.get("primary") or (gd["acceptable"] or [None])[0]
                    wrong = next(a for a in area_ids if a not in (gd["acceptable"] or []))
                    row = dict(row, output="assigned", area_id=right if what == "right" else wrong, unresolved_reason=None)
                else:
                    row = dict(row, output="none", area_id=None, unresolved_reason=None) if what == "right" else dict(row, output="assigned", area_id=area_ids[0], unresolved_reason=None)
                if raw_kind in ("right", "wrong"):
                    if gd["covered"]:
                        right = gd.get("primary") or (gd["acceptable"] or [None])[0]
                        wrong = next(a for a in area_ids if a not in (gd["acceptable"] or []))
                        row = dict(row, raw_covered=True, raw_choice=right if raw_kind == "right" else wrong)
                    else:
                        row = dict(row, raw_covered=False, raw_choice=None) if raw_kind == "right" else dict(row, raw_covered=True, raw_choice=area_ids[0])
            out_rows.append(row)
        wjl(os.path.join(d_, "rounds", str(r_), "arms", arm + ".jsonl"), out_rows)
    return n


def hash_first(case_ids, tag, k):
    """The first k case ids in the order of sha256(tag + case id): a deterministic hash-selected subset."""
    return set(sorted(case_ids, key=lambda c: hashlib.sha256((tag + "\0" + c).encode()).hexdigest())[:k])


def e2e_abstention_worlds(tmp, e, prim):
    """R4-13 (8) / R4-05 / R5-04 on copies of the finished planted-effect run (the primary arm's composed frozen-round rows rewritten in the copies, `below_t_choice` abstentions with their raw fields set per world, then
    analysis rounds 8-10 and `verdict` re-run): 5 % hash-selected abstentions on the planted effect -> GO; 20 % -> NO-GO with the cap the only failing point condition; a gaming world (both arms 0.88 correct, equal
    counts, Jev abstains on exactly the clusters it would get wrong, wrong raw guesses) -> INCONCLUSIVE with the guard h the only failing condition and gain_guess 0; a near-null (the gaming world with the raw guess right
    on exactly 2 abstained clusters) -> INCONCLUSIVE, gain_guess point > 0 but its 95 % bound <= 0 (a point-only guard would pass); an honest 10 % Jev against a strong B2 -> GO with gain_guess > gain_all; the
    concentrated-gain null (Jev correct on B2's errors in exactly 2 PRs) -> INCONCLUSIVE, `lb90` equals an independent bootstrap and is <= 0."""
    area_ids = [a["area_id"] for a in e.cat()["areas"]]
    base_dir = os.path.join(tmp, "e2e-abst-base")
    shutil.copytree(e.data, base_dir)
    cl = frozen_go_clusters(base_dir)
    ids = [c for _r, c, _p in cl]
    n = len(ids)

    def world(name, prim_decide=None, b2_decide=None, abstain_raw="keep"):
        d_ = os.path.join(tmp, "e2e-abst-" + name)
        shutil.copytree(base_dir, d_)
        if b2_decide:
            rewrite_arm(d_, "B2", b2_decide, area_ids)
        if prim_decide:
            rewrite_arm(d_, prim, prim_decide, area_ids, abstain_raw)
        x = reanalyse(d_, e.reviews)
        return x, verdict_json(d_), d_

    def indep_rows(dc):
        real = [k for r_ in x2d.FROZEN_ROUNDS for k in A.build_cases([u for u in A.scored_units(A.load_round(r_, dc)["units"]) if u.get("stratum_reason") != "core_gate"], A.load_round(r_, dc)["gold"],
                                                                      {prim: A.load_round(r_, dc)["arms"][prim], "B2": A.load_round(r_, dc)["arms"]["B2"]}) if not k["control"]]
        return [r for r in A.pooling_rows(real, prim, "B2") if r["n_all"]]


    k5, k20 = round(0.05 * n), round(0.20 * n)
    ab5 = hash_first(ids, "x2d-abst-5", k5)
    x, v5, _d = world("u05", lambda c: "abstain" if c in ab5 else None)
    print("INFO abstention world u = 5 %%: %d of %d clusters abstained; verdict %s; u %s; failing %s" % (k5, n, v5 and v5["verdict"], v5 and v5["pooled"]["unresolved"]["rate"], v5 and [k for k, c in v5["go_verdict"]["conditions"].items() if not c["pass"]]))
    check("E R4-13 (8) planted effect with 5 %% hash-selected abstentions (%d of %d clusters, u = %.3f) -> GO: every condition holds, including the cap (g), the guard (h) and n_res >= 50 (i)" % (k5, n, k5 / n),
          x.returncode == 0 and v5 is not None and v5["verdict"] == "GO" and all(c["pass"] for c in v5["go_verdict"]["conditions"].values()) and v5["pooled"]["unresolved"]["k"] == k5 and v5["pooled"]["n_res"] >= 50
          and v5["pooled"]["unresolved"]["by_reason"]["below_t_choice"] == k5 and set(v5["unresolved_by_round"]) == {"8", "9", "10"}, (x.returncode, x.stderr[-300:], v5 and (v5["verdict"], v5["go_verdict"]["reasons"])))
    ab20 = hash_first(ids, "x2d-abst-20", k20)
    x, v20, _d = world("u20", lambda c: "abstain" if c in ab20 else None)
    pt20 = [k for k, c in (v20 or {}).get("go_verdict", {}).get("conditions", {}).items() if c["kind"] == "point" and not c["pass"]]
    print("INFO abstention world u = 20 %%: %d of %d clusters abstained; verdict %s; failing point conditions %s" % (k20, n, v20 and v20["verdict"], pt20))
    check("E R4-13 (8) the same world with 20 %% abstentions (%d of %d, u = %.3f > the 0.15 cap) -> NO-GO with the unresolved-rate cap the ONLY failing point condition (a cap breach is a measured product failure)" % (k20, n, k20 / n),
          x.returncode == 0 and v20 is not None and v20["verdict"] == "NO-GO" and pt20 == ["unresolved rate <= 0.15 (cap)"], (x.returncode, x.stderr[-300:], v20 and (v20["verdict"], pt20)))
    # gaming world: Jev and B2 are both right on exactly round(0.88 n) clusters (independent seeded draws); Jev abstains on exactly the clusters it would get wrong, and its raw (unthresholded) guess there is
    # WRONG (R5-04: abstain_raw="wrong"), so gain_guess = (k_ok - k_ok) / n = 0 exactly while gain_res is a selection artefact
    k_ok = round(0.88 * n)
    rg = random.Random(x2d.SEED)
    jev_ok = set(rg.sample(ids, k_ok))
    b2_ok = set(rg.sample(ids, k_ok))
    x, vg, _d = world("gaming", lambda c: "right" if c in jev_ok else "abstain", lambda c: "right" if c in b2_ok else "wrong", abstain_raw="wrong")
    cg = (vg or {}).get("go_verdict", {}).get("conditions", {})
    guard = {k: c for k, c in cg.items() if c["kind"] == "guard"}
    failing_g = [k for k, c in cg.items() if not c["pass"]]
    print("INFO gaming world (Jev abstains on its own %d errors of %d, wrong raw guesses; B2 right on %d): verdict %s; gain_res %s; gain_all %s; gain_guess %s (lb95 %s); u %s; failing %s" % (
        n - k_ok, n, k_ok, vg and vg["verdict"], vg and vg["pooled"]["gain_res"]["point"], vg and vg["pooled"]["gain_all"]["point"], vg and vg["pooled"]["gain_guess"]["point"], vg and vg["pooled"]["gain_guess"]["lb95"], vg and vg["pooled"]["unresolved"]["rate"], failing_g))
    check("E R5-04 gaming world (both arms right on %d of %d clusters, Jev abstains on exactly the clusters it would answer wrongly and its raw guess there is wrong, u = %.3f <= the cap): INCONCLUSIVE, the guard h is the ONLY failing condition, gain_guess.point == 0.0, gain_res >= 0.10 reported (the point condition d passes: selection alone)" % (k_ok, n, (n - k_ok) / n),
          x.returncode == 0 and vg is not None and vg["verdict"] == "INCONCLUSIVE" and len(guard) == 1 and failing_g == list(guard) and abs(vg["pooled"]["gain_guess"]["point"]) < 1e-12 and vg["pooled"]["gain_res"]["point"] is not None
          and vg["pooled"]["gain_res"]["point"] >= 0.10 and cg["unresolved rate <= 0.15 (cap)"]["pass"] is True and vg["pooled"]["gain_guess"]["n"] == n, (x.returncode, x.stderr[-300:], vg and (vg["verdict"], vg["pooled"]["gain_res"], vg["pooled"]["gain_guess"], failing_g)))
    # near-null: the gaming world with the raw guess RIGHT on exactly 2 hash-chosen abstained clusters: gain_guess.point = 2 / n > 0 (a point-only guard would pass) but its 95 % bound is <= 0
    abst_g = sorted(set(ids) - jev_ok)
    near2 = hash_first(abst_g, "x2d-near-null", 2)
    x, vn_, dn_ = world("nearnull", lambda c: "right" if c in jev_ok else "abstain", lambda c: "right" if c in b2_ok else "wrong", abstain_raw=lambda c: "right" if c in near2 else "wrong")
    cn_ = (vn_ or {}).get("go_verdict", {}).get("conditions", {})
    failing_n = [k for k, c in cn_.items() if not c["pass"]]
    gg_n = (vn_ or {}).get("pooled", {}).get("gain_guess") or {}
    indep_g = A.pr_cluster_bootstrap(indep_rows(dn_), A.gain_guess_stat, B=10000)["lower"][0.95] if vn_ is not None else None
    would_go = A.go_verdict(dict(vn_["pooled"], gain_guess=dict(gg_n, lb95=1.0)), json.load(open(os.path.join(S, "thresholds.json"))))["verdict"] if vn_ is not None else None
    print("INFO near-null world (raw guess right on 2 abstained clusters %s): verdict %s; gain_guess %s lb95 %s (independent %s); failing %s; the same pooled dict with lb95 := 1.0 -> %s" % (sorted(near2), vn_ and vn_["verdict"], gg_n.get("point"), gg_n.get("lb95"), indep_g, failing_n, would_go))
    check("E R5-04 near-null world (the gaming world with the raw guess right on exactly 2 hash-chosen abstained clusters): INCONCLUSIVE with the guard h the ONLY failing condition, gain_guess.point = 2 / %d > 0, its stored 95 %% lower bound is <= 0 and equals an independently recomputed pr_cluster_bootstrap(rows, gain_guess_stat, B=10000) bound, and go_verdict on the same pooled dict with lb95 := 1.0 is GO (a point-only guard would have passed)" % n,
          x.returncode == 0 and vn_ is not None and vn_["verdict"] == "INCONCLUSIVE" and failing_n == [k for k, c in cn_.items() if c["kind"] == "guard"] and len(failing_n) == 1 and abs(gg_n["point"] - 2 / n) < 1e-12
          and gg_n["lb95"] is not None and gg_n["lb95"] <= 0 and indep_g is not None and abs(gg_n["lb95"] - indep_g) < 1e-12 and would_go == "GO", (x.returncode, x.stderr[-300:], vn_ and (vn_["verdict"], gg_n, failing_n, would_go)))
    # honest 10 % Jev against a strong B2 (planted effect: Jev right on every cluster it answers, B2 right on 80 % hash-chosen clusters, 9 hash-chosen abstentions keeping the stub's own raw guesses): GO; the guard is not gain_all
    k9 = round(0.10 * n)
    ab9 = hash_first(ids, "x2d-honest-10", k9)
    b2_80 = hash_first(ids, "x2d-b2-80", round(0.80 * n))
    x, vh, _d = world("honest10", lambda c: "abstain" if c in ab9 else "right", lambda c: "right" if c in b2_80 else "wrong", abstain_raw="keep")
    ch_ = (vh or {}).get("go_verdict", {}).get("conditions", {})
    gh_ = (vh or {}).get("pooled", {})
    print("INFO honest 10 %% world (%d abstentions, B2 right on %d): verdict %s; gain_res %s; gain_all %s; gain_guess %s (lb95 %s); u %s; failing %s" % (k9, len(b2_80), vh and vh["verdict"], gh_.get("gain_res", {}).get("point"), gh_.get("gain_all", {}).get("point"),
                                                                                                                                       gh_.get("gain_guess", {}).get("point"), gh_.get("gain_guess", {}).get("lb95"), gh_.get("unresolved", {}).get("rate"), [k for k, c in ch_.items() if not c["pass"]]))
    check("E R5-04 honest 10 %% Jev against a strong B2 (%d hash-chosen abstentions keeping the stub's raw guesses, B2 right on %d of %d clusters): GO, the guard h passing, and gain_guess.point > gain_all.point (the guard is not gain_all)" % (k9, len(b2_80), n),
          x.returncode == 0 and vh is not None and vh["verdict"] == "GO" and all(c["pass"] for c in ch_.values()) and gh_["gain_guess"]["point"] > gh_["gain_all"]["point"] and gh_["unresolved"]["k"] == k9, (x.returncode, x.stderr[-300:], vh and (vh["verdict"], gh_.get("gain_guess"), gh_.get("gain_all"))))
    # R4-05 concentrated-gain null: B2 is wrong on every cluster of the two largest frozen PRs (and right elsewhere), Jev is right everywhere: the gain point passes (>= +0.10), P_assign passes, but only 2 PRs carry the gain
    by_pr = collections.Counter(p for _r, _c, p in cl)
    top2 = [p for p, _k in sorted(by_pr.items(), key=lambda kv: (-kv[1], kv[0]))[:2]]
    conc = {c for _r, c, p in cl if p in top2}
    x, vc, dc = world("conc", lambda c: "right", lambda c: "wrong" if c in conc else "right")
    cc = (vc or {}).get("go_verdict", {}).get("conditions", {})
    m_prs = len(by_pr)
    p_miss = ((m_prs - 2) / m_prs) ** m_prs
    lb = vc and vc["pooled"]["gain_res"]["lb90"]
    indep = None
    if vc is not None:
        real = [k for r_ in x2d.FROZEN_ROUNDS for k in A.build_cases([u for u in A.scored_units(A.load_round(r_, dc)["units"]) if u.get("stratum_reason") != "core_gate"], A.load_round(r_, dc)["gold"],
                                                                      {prim: A.load_round(r_, dc)["arms"][prim], "B2": A.load_round(r_, dc)["arms"]["B2"]}) if not k["control"]]
        rows = [r for r in A.pooling_rows(real, prim, "B2") if r["n_res"]]
        indep = A.pr_cluster_bootstrap(rows, lambda rs: (sum(q["d_res"] for q in rs) / sum(q["n_res"] for q in rs)) if sum(q["n_res"] for q in rs) else None, B=10000)["lower"][0.9]
    print("INFO concentrated-gain null (2 PRs %s with %d of %d clusters; %d PRs, P(resample misses both) = %.3f): verdict %s; gain_res %s; lb90 %s (independent %s); jev-only PRs %s" % (
        top2, len(conc), n, m_prs, p_miss, vc and vc["verdict"], vc and vc["pooled"]["gain_res"]["point"], lb, indep, vc and vc["pooled"]["gain_res"]["n_jev_only_prs"]))
    check("E R4-05 concentrated-gain null (Jev right on B2's errors in exactly 2 PRs, >= 10 clusters, gain point >= +0.10, P_assign 1.0): the verdict is INCONCLUSIVE (never GO), the gain point and P_assign point conditions PASS, the spread fails with n_jev_only_prs == 2, "
          "and the stored lb90 equals an independently recomputed PR-cluster bootstrap lower bound (and is <= 0 when a resample misses both PRs with probability > 0.11)",
          x.returncode == 0 and vc is not None and vc["verdict"] == "INCONCLUSIVE" and len(conc) >= 10 and vc["pooled"]["gain_res"]["point"] >= 0.10 and cc["P_assign point >= 0.90"]["pass"] is True
          and cc["gain_res point >= +0.10 (resolved clusters)"]["pass"] is True and vc["pooled"]["gain_res"]["n_jev_only_prs"] == 2 and cc["Jev-correct/B2-wrong resolved clusters from >= 3 distinct PRs"]["pass"] is False
          and lb is not None and indep is not None and abs(lb - indep) < 1e-12 and (p_miss <= 0.11 or lb <= 0), (x.returncode, x.stderr[-300:], vc and (vc["verdict"], vc["pooled"]["gain_res"], cc)))


def e2e_omitted_rows(tmp, e):
    """R2-13 on a copy of the finished run: a V-sem required row (covered) and a fan-out unit's whole area_covers set are removed from round 9's scores: missing_required names both, compose.py and
    `run.sh analysis 9` refuse, and nothing is composed."""
    d = os.path.join(tmp, "e2e-omit")
    shutil.copytree(e.data, d)
    sp = os.path.join(d, "rounds", "9", "jev", "scores.jsonl")
    rows = jl(sp)
    units = jl(os.path.join(d, "rounds", "9", "units.jsonl"))
    win = sorted((u for u in units if u.get("set") == "window" and not u.get("control")), key=lambda u: u["case_id"])
    a_, b_, c_ = win[0]["case_id"], win[1]["case_id"], win[2]["case_id"]
    area_ids = [a["area_id"] for a in e.cat()["areas"]]
    kept = [r_ for r_ in rows if not ((r_.get("case_id") == a_ and r_.get("question") == "merge.covered" and r_.get("variant") == "V-sem") or (r_.get("case_id") == b_ and r_.get("question") == "merge.area_covers"))]
    wjl(sp, kept)
    miss = CM.missing_required(kept, units, area_ids=area_ids)
    env = dict(e.env, X2D_DATA=d)
    c = run([PY, os.path.join(S, "compose.py"), "--round", "9"], cwd=ROOT, env=env)
    ra = run([os.path.join(S, "run.sh"), "analysis", "9"], cwd=ROOT, env=env, timeout=3600)
    check("E R2-13 / R3-06 omitted required rows: missing_required names the V-sem merge.covered row of one unit and every area row of another unit's fan-out; compose.py and `run.sh analysis 9` refuse (non-zero, 'required Jev request')",
          (a_, "V-sem/merge.covered", 1) in miss and all((b_, "fan-out/merge.area_covers", ar) in miss for ar in area_ids) and c.returncode != 0 and "required Jev request" in (c.stdout + c.stderr) and ra.returncode != 0 and "required Jev request" in (ra.stdout + ra.stderr),
          (miss[:3], (c.stdout + c.stderr)[-200:], (ra.stdout + ra.stderr)[-200:]))
    drop = next(r_ for r_ in rows if r_.get("case_id") == c_ and r_.get("question") == "merge.area_covers")  # R3-06: ONE area row of ONE unit (a cost stop between a unit's area requests)
    wjl(sp, [r_ for r_ in rows if r_ is not drop])
    miss1 = CM.missing_required([r_ for r_ in rows if r_ is not drop], units, area_ids=area_ids)
    c1 = run([PY, os.path.join(S, "compose.py"), "--round", "9"], cwd=ROOT, env=env)
    ra1 = run([os.path.join(S, "run.sh"), "analysis", "9"], cwd=ROOT, env=env, timeout=3600)
    check("E R3-06 a partial fan-out alone (one area row of one unit missing: %s / %s) is refused by compose.py and `run.sh analysis 9` (non-zero, 'required Jev request')" % (c_, drop.get("area_id")),
          miss1 == [(c_, "fan-out/merge.area_covers", drop.get("area_id"))] and c1.returncode != 0 and "required Jev request" in (c1.stdout + c1.stderr) and ra1.returncode != 0 and "required Jev request" in (ra1.stdout + ra1.stderr),
          (miss1, (c1.stdout + c1.stderr)[-200:], (ra1.stdout + ra1.stderr)[-200:]))
    wjl(sp, kept)
    dev = TU.round_rows(9, d, "dev")
    ids_dev = {r_["case_id"] for rows_ in dev.values() for r_ in rows_}
    check("E R2-13 tuning dev rows hold neither controls nor audit units, and a window unit with a missing required row (%s) drops out of the dev rows instead of being scored on a partial answer" % a_,
          a_ not in ids_dev and not any(u["case_id"] in ids_dev for u in units if u.get("control") or u.get("set") == "audit"), len(ids_dev))


def e2e_go_population_kappa(tmp, e, res_all):
    """R2-14 on a copy of the finished run: one scored unit of round 9 whose labellers agreed becomes a core-gate unit and its A label is flipped: the GO population leaves it, so
    `informative` (and futility) use the kappa of the remaining units, while the all-units figure (report) shows the disagreement."""
    d = os.path.join(tmp, "e2e-go")
    shutil.copytree(e.data, d)
    L = A.load_round(9, d)
    su = sorted(A.scored_units(L["units"]), key=lambda u: u["case_id"])
    ss = A._sessions(L["labels"], "wide", "main")
    u = next(u_ for u_ in su if all(s in ss.get(u_["case_id"], {}) for s in "AB") and A._pair_agree(ss[u_["case_id"]]["A"], ss[u_["case_id"]]["B"]))
    ids_all, ids_go = {x["case_id"] for x in su}, {x["case_id"] for x in su} - {u["case_id"]}
    want_go = A.label_agreement(L["labels"], ids_go)["kappa_answer"]
    want_all_before = A.label_agreement(L["labels"], ids_all)["kappa_answer"]
    units_p = os.path.join(d, "rounds", "9", "units.jsonl")
    wjl(units_p, [dict(x, stratum=3, stratum_reason="core_gate") if x["case_id"] == u["case_id"] else x for x in jl(units_p)])
    lab_p = os.path.join(d, "rounds", "9", "labels.jsonl")
    rows = jl(lab_p)
    last_a = max(i for i, x in enumerate(rows) if x["case_id"] == u["case_id"] and x.get("kind") == "wide" and str(x.get("by", "")).startswith("labeller_A"))
    rows[last_a] = dict(rows[last_a], covered=not rows[last_a]["covered"], acceptable=[] if rows[last_a]["covered"] else ["zz"], primary=None if rows[last_a]["covered"] else "zz", none_kind="gap" if rows[last_a]["covered"] else None)
    wjl(lab_p, rows)
    res9 = A.round_results(9, d, B=100)
    got_go, got_all, got_inf = res9["labels_go_population"]["kappa_answer"], res9["labels"]["kappa_answer"], res9["informative"]["kappa_answer"]
    check("E R2-14 informative / futility use the GO-population kappa: a core-gate unit's labeller disagreement does not move it (GO kappa_answer %.4f = the kappa of the other units %.4f), while the all-units report figure does (%.4f -> %.4f)" % (got_go, want_go, want_all_before, got_all),
          got_go == want_go and got_inf == want_go and got_all != want_go and got_all < want_all_before and res9["n_core_gate_excluded"] == 1 and res9["informative"]["informative"] == res_all[9]["informative"]["informative"], (got_go, want_go, got_all, got_inf))


def e2e_failed_kappa(tmp, e):
    """R3-02 (3) on a copy of the finished run: in round 5 (no guide revision) a labels-check that fails on kappa < 0.6 refuses `gold` (the clarification path); in round 3 (one clarification recorded) the same failure lets the round proceed:
    labels-check exits 0, `gold` is not refused by it, and the analysis scores the round as NON-INFORMATIVE."""
    d = os.path.join(tmp, "e2e-kfail")
    shutil.copytree(e.data, d)

    def flip_a(r):
        p = os.path.join(d, "rounds", str(r), "labels.jsonl")
        rows = jl(p)
        last = {}
        for i, x in enumerate(rows):
            if x.get("kind") == "wide" and str(x.get("by", "")).startswith("labeller_A"):
                last[x["case_id"]] = i
        for i in last.values():
            x = rows[i]
            cov = not x["covered"]
            rows[i] = dict(x, covered=cov, acceptable=["zz"] if cov else [], primary="zz" if cov else None, none_kind=None if cov else "gap")
        wjl(p, rows)
    flip_a(5)
    l5 = e.rs("labels-check", 5, X2D_DATA=d)
    g5 = e.rs("gold", 5, X2D_DATA=d)
    check("E R3-02 round 5 (no guide revision recorded): a labels-check failing on kappa_answer < 0.6 exits 1 and `gold` is refused on it (the clarification path); the rule concentration (every session cites R1) never triggers",
          l5.returncode == 1 and "kappa_answer" in (l5.stdout + l5.stderr) and g5.returncode != 0 and "labels-check of round 5 failed" in (g5.stdout + g5.stderr), ((l5.stdout + l5.stderr)[-200:], (g5.stdout + g5.stderr)[-200:]))
    flip_a(3)
    l3 = e.rs("labels-check", 3, X2D_DATA=d)
    lc3 = json.load(open(os.path.join(d, "rounds", "3", "labels-check.json")))
    g3 = e.rs("gold", 3, X2D_DATA=d)
    for fn in ("results.json", "report.md"):
        if os.path.exists(os.path.join(d, "rounds", "3", fn)):
            os.remove(os.path.join(d, "rounds", "3", fn))
    xa = run([PY, os.path.join(S, "analysis_x2d.py"), "--round", "3"], cwd=ROOT, env=clean_env(X2D_DATA=d, X2D_REVIEWS_DIR=e.reviews))
    res3 = json.load(open(os.path.join(d, "rounds", "3", "results.json"))) if xa.returncode == 0 else {}
    check("E R3-02 round 3 (the one guide clarification recorded): the same failure (kappa_answer %s < 0.6) exits 0 with ok false, is not refused by `gold`'s labels-check gate, and the round is scored (%s clusters) and NON-INFORMATIVE" % (lc3.get("kappa_answer"), res3.get("n_scored_clusters")),
          l3.returncode == 0 and lc3["ok"] is False and "non-informative" in l3.stdout and "labels-check of round 3 failed" not in (g3.stdout + g3.stderr) and xa.returncode == 0
          and res3["informative"]["informative"] is False and res3["informative"]["labels_check_ok"] is False and res3["n_scored_clusters"] > 0, ((l3.stdout + l3.stderr)[-200:], (g3.stdout + g3.stderr)[-200:], xa.stderr[-200:]))


def modification_only_checks(tmp, repo, shas):
    """R2-07: a commit that only modifies files still gets a lookups row (known modified_owned / modified_unowned totals); the lookups must sum to counts.json (`_end_to_end` asserts it)."""
    c3 = shas[3]
    c4 = repo.commit("Tweak (#103)", {"fastapi/a.py": py_text("a") + "# tweak\n", "fastapi/c.py": body_text("b", 40) + "y\n"})
    mem = {"round": 1, "areas": {"A": ["fastapi/a.py", "docs_src/topic/tutorial001.py", "fastapi/b.py"], "B": ["tests/test_tutorial/test_topic/test_tutorial001.py", "docs_src/known/tutorial001.py"]},
           "pool": [], "heuristic_only": ["docs_src/newtopic/tutorial001.py", "fastapi/zzz.py"]}
    res = BU.build_round_units(repo.path, 1, mem, rounds_info={"after_commit": c3, "through_commit": c4, "prs": []}, settings=dict(x2d.DEFAULT_SETTINGS), core_gate=False, check_prepass=False)
    row = next((l_ for l_ in res["lookups"] if l_["commit"] == c4), None)
    check("R2-07 a modification-only commit (fastapi/a.py owned, fastapi/c.py still unowned) has a lookups row with the known totals (modified_owned 1, modified_unowned 1) and no unit; counts equal the sum of the lookups",
          row is not None and row["counts"].get("modified_owned") == 1 and row["counts"].get("modified_unowned") == 1 and not res["units"] and {k: v for k, v in res["counts"].items() if v and isinstance(v, int)} == {k: v for k, v in row["counts"].items() if v}, (row["counts"], res["counts"]))
    ee = A._end_to_end([], res["lookups"], [], {}, total=res["counts"])
    try:
        A._end_to_end([], [], [], {}, total=res["counts"])
        dropped = False
    except ValueError as ex:
        dropped = "dropped" in str(ex)
    check("R2-07 _end_to_end counts the modification-only commit's owned file as an exact lookup (1) and refuses (ValueError) when the lookups rows do not sum to counts.json (a dropped commit)", ee["rule_exact_files"] == 1 and ee["modified_unowned_total"] == 1 and dropped, (ee, dropped))


def guard_checks(tmp):
    """R2-16: X2D_SKIP_SIGNOFF=1 and X2D_REVIEWS_DIR are honoured only with a loopback X2D_JEV_ENDPOINT and an X2D_DATA outside S/data; the sign-off record is otherwise required."""
    live = os.path.join(S, "data")
    ok_data = os.path.join(tmp, "guard-data")
    os.makedirs(ok_data, exist_ok=True)
    cmd = [PY, os.path.join(S, "versioning.py"), "--round", "1", "--drift"]
    base = lambda **kw: clean_env(X2D_DATA=ok_data, **kw)
    env_noep = base()
    env_noep.pop("X2D_JEV_ENDPOINT", None)
    runs = {"a non-loopback X2D_JEV_ENDPOINT": run(cmd, cwd=ROOT, env=base(X2D_JEV_ENDPOINT="https://openrouter.ai/api/alpha/decisions")), "no X2D_JEV_ENDPOINT": run(cmd, cwd=ROOT, env=env_noep),
            "X2D_DATA inside S/data": run(cmd, cwd=ROOT, env=clean_env(X2D_DATA=os.path.join(live, "validation-must-not-exist"))),
            "a loopback endpoint and X2D_DATA outside S/data": run(cmd, cwd=ROOT, env=base())}
    out = {k: v.stdout + v.stderr for k, v in runs.items()}
    check("R2-16 X2D_SKIP_SIGNOFF=1 is refused with a non-loopback endpoint, without an endpoint and with X2D_DATA inside S/data (non-zero, names the loopback / S/data condition); a loopback endpoint with X2D_DATA outside passes the guard (it stops later on the missing round)",
          all(runs[k].returncode != 0 and "loopback" in out[k] for k in list(runs)[:3]) and "loopback" not in out["a loopback endpoint and X2D_DATA outside S/data"] and not os.path.exists(os.path.join(live, "validation-must-not-exist")),
          {k: v[-160:] for k, v in out.items()})
    probe = [PY, "-c", "import sys; sys.path.insert(0, %r); import x2d; print(x2d.reviews_dir())" % S]
    r_bad = run(probe, cwd=ROOT, env=base(X2D_REVIEWS_DIR=os.path.join(tmp, "rv"), X2D_JEV_ENDPOINT="https://example.com/x"))
    r_ok = run(probe, cwd=ROOT, env=base(X2D_REVIEWS_DIR=os.path.join(tmp, "rv")))
    r_data = run(probe, cwd=ROOT, env=clean_env(X2D_REVIEWS_DIR=os.path.join(tmp, "rv"), X2D_DATA=os.path.join(live, "x")))
    check("R2-16 X2D_REVIEWS_DIR is refused with a non-loopback endpoint and with X2D_DATA inside S/data, and honoured in a validation environment",
          r_bad.returncode != 0 and "X2D_REVIEWS_DIR" in r_bad.stderr and r_data.returncode != 0 and "X2D_REVIEWS_DIR" in r_data.stderr and r_ok.returncode == 0 and r_ok.stdout.strip() == os.path.join(tmp, "rv"),
          (r_bad.stderr[-120:], r_data.stderr[-120:], r_ok.stdout, r_ok.stderr[-120:]))
    check("R2-16 x2d.require_signoff() is False here (X2D_SKIP_SIGNOFF=1 in a validation environment) and True without the flag", x2d.require_signoff() is False and run([PY, "-c", "import sys; sys.path.insert(0, %r); import x2d; print(x2d.require_signoff())" % S], cwd=ROOT,
          env={k: v for k, v in clean_env().items() if k != "X2D_SKIP_SIGNOFF"}).stdout.strip() == "True")


def r2_misc_checks(tmp):
    """R2-11 (power on a live-shaped k23), R2-18 (text and config consistency)."""
    k23 = os.path.join(tmp, "k23-fixture.json")
    json.dump({"frozen_pr_counts": {"8": {"a": 6, "b": 6}, "9": {"c": 6}, "10": {"d": 6}}, "real_none_proxy": 12}, open(k23, "w"))
    st, proxy = PS.structure_from_k23(k23)
    res = PS.run(pilots=100, boot_B=100, k23=k23)
    check("R2-11 / R4-13 / R5-01 power_sim --structure-k23 on a tiny k23 fixture: schema x2d-power-v5, structure = 24 clusters in 4 PRs, the real_none_proxy (12) is n_pos (positives inside the population), p3_real_structure (the P3-u0 share), p3_abstention (0.05 / 0.10), n4_max and the cells P3_u0 / P3_u05 / P3_u10 / N3 (14) / N3_PR (14) / N4a (4) / N4b (4) / N4b_cap (2) / N4b_rho0 (3) / N4c are recorded",
          res["schema"] == "x2d-power-v5" and res["structure"]["clusters"] == 24 and res["structure"]["prs"] == 4 and proxy == 12 and res["real_none_proxy"] == 12 and res["n_pos"] == 12 and sum(s["n"] for s in st) == 24
          and res["structure_source"] == "k23" and res["p3_real_structure"]["share"] == res["P3_u0"]["share"] and set(res["p3_abstention"]) == {"0.05", "0.10"} and res["n4_max"] == max(c["share"] for c in res["N4a"] + res["N4b"] + res["N4b_cap"] + res["N4b_rho0"])
          and (len(res["N3"]), len(res["N3_PR"]), len(res["N4a"]), len(res["N4b"]), len(res["N4b_cap"]), len(res["N4b_rho0"])) == (14, 14, 4, 4, 2, 3) and "N4c" in res and all(c["n_pos"] == 12 for c in [res["P3_u0"], res["P3_u05"], res["P3_u10"], res["N4c"]]),
          (res["schema"], res["structure"], res["n_pos"]))
    k23b = os.path.join(tmp, "k23-fixture-cap.json")
    json.dump({"frozen_pr_counts": {"8": {"a": 8, "b": 8}, "9": {"c": 8}, "10": {"d": 8}}, "real_none_proxy": 30}, open(k23b, "w"))
    k23c = os.path.join(tmp, "k23-fixture-zero.json")
    json.dump({"frozen_pr_counts": {"8": {"a": 8, "b": 8}, "9": {"c": 8}, "10": {"d": 8}}, "real_none_proxy": 0}, open(k23c, "w"))
    check("R4-13 (6) at round 0 the positives inside the population are real_none_proxy, floor 0 and cap 20 (proxy 30 -> n_pos 20; proxy 0 -> n_pos 0)",
          PS.run(pilots=5, boot_B=10, k23=k23b)["n_pos"] == 20 and PS.run(pilots=5, boot_B=10, k23=k23c)["n_pos"] == 0)
    bad = {}
    for name, doc in (("missing proxy", {"frozen_pr_counts": {"8": {"a": 1}}}), ("missing counts", {"real_none_proxy": 3}), ("empty structure", {"frozen_pr_counts": {"8": {}}, "real_none_proxy": 3})):
        p = os.path.join(tmp, "k23-bad.json")
        json.dump(doc, open(p, "w"))
        try:
            PS.structure_from_k23(p)
            bad[name] = False
        except SystemExit:
            bad[name] = True
    check("R2-11 structure_from_k23 refuses a k23 without real_none_proxy, without frozen_pr_counts and with an empty structure", all(bad.values()) and len(bad) == 3, bad)
    g = {"case_id": "u", "covered": True, "acceptable": ["A", "B", "C"], "primary": "B", "none_kind": None}
    m1, m2, m3 = G.derive_ablation(g, {"A", "B"}), G.derive_ablation(g, {"A", "B", "C"}), G.derive_ablation(g, ["A"])
    check("R2-05 derive_ablation with a multi-acceptable source and a multi-area ablation (the SAME set the catalogue removes): {A, B} leaves C covered and clears the ablated primary B; {A, B, C} derives none / gap; {A} keeps primary B",
          m1["covered"] is True and m1["acceptable"] == ["C"] and m1["primary"] is None and m2["covered"] is False and m2["none_kind"] == "gap" and m2["acceptable"] == [] and m3["acceptable"] == ["B", "C"] and m3["primary"] == "B", (m1, m2, m3))
    doc = BL.__doc__
    order = "units -> label -> labels-check -> gold -> baselines -> jev -> tune -> analysis"
    runsh = open(os.path.join(S, "run.sh")).read()
    prereg = open(os.path.join(S, "PREREGISTRATION.md")).read()
    check("R2-18 text and config consistency: the baselines docstring names the run.sh stage order (units -> label -> labels-check -> gold -> baselines -> jev -> tune -> analysis, as `round)` runs it), the review files are named reviews/round-<N>.md, and neither thresholds.json nor x2d.CAPS carries the dead cli_sessions_per_round / cli_sessions_round_0 entries",
          order in doc and "for s in units label labels_check gold baselines jev tune analysis" in runsh and "pilot-design-round" not in prereg and "cli_sessions_per_round" not in json.dumps(x2d.CAPS)
          and not any(k in open(os.path.join(S, "thresholds.json")).read() for k in ('"cli_sessions_per_round"', '"cli_sessions_round_0"')), (order in doc, doc[:200]))


def gates_refusal_check(tmp):
    d = os.path.join(tmp, "empty-data")
    os.makedirs(d, exist_ok=True)
    r = run([os.path.join(S, "run.sh"), "units", "1"], cwd=ROOT, env=clean_env(X2D_DATA=d, X2D_SRC=SRC))
    check("E run.sh units 1 refuses without rounds/0/gates.json", r.returncode != 0 and "gates.json" in (r.stdout + r.stderr), (r.stdout + r.stderr)[-300:])


def r3_checks(tmp, repo, res):
    """R3-04 .. R3-18 unit checks that need no pipeline: R3-05 the excerpt re-render, R3-09 the inside-population estimators, R3-16 the round-0 critique guards."""
    # R3-05: rerender_units re-renders state.unit.excerpt under another excerpt_lines from the same member text at the unit's commit; nothing else changes
    def n_lines(u):
        rep_ = next((p for p in sorted(u["members"]) if BU.read_text(repo.path, u["commit"], p).strip()), sorted(u["members"])[0])
        return len(BU.read_text(repo.path, u["commit"], rep_).splitlines())
    src = max(res["units"], key=n_lines)
    short, long_ = BU.rerender_units(repo.path, [src], 1)[0], BU.rerender_units(repo.path, [src], 120)[0]
    rep = next((p for p in sorted(src["members"]) if BU.read_text(repo.path, src["commit"], p).strip()), sorted(src["members"])[0])
    full_text = BU.read_text(repo.path, src["commit"], rep).splitlines()
    strip = lambda u: {k: v for k, v in u.items() if k not in ("excerpt_lines",) and k != "state"}
    check("R3-05 build_units.rerender_units re-renders the excerpt under another excerpt_lines from the same representative member at the unit's commit (1 line vs 120 lines of a %d-line file), records excerpt_lines, and leaves ids, membership, strata and every other field unchanged" % len(full_text),
          len(full_text) > 1 and short["state"]["unit"]["excerpt"] == "\n".join(full_text[:1]) and long_["state"]["unit"]["excerpt"] == "\n".join(full_text[:120]) and short["excerpt_lines"] == 1 and long_["excerpt_lines"] == 120
          and strip(short) == strip(src) and strip(long_) == strip(src) and {k: v for k, v in short["state"].items() if k != "unit"} == {k: v for k, v in src["state"].items() if k != "unit"}
          and {k: v for k, v in short["state"]["unit"].items() if k != "excerpt"} == {k: v for k, v in src["state"]["unit"].items() if k != "excerpt"}, (short["state"]["unit"]["excerpt"], len(full_text)))
    # R3-09 / R4-13 (6): simulate builds the figures the rule reads as the analysis does (spy on A.go_verdict, whose pooled argument is the simulation's summary)
    st = [{"round": 8, "pr": "a", "n": 46}, {"round": 9, "pr": "b", "n": 45}]

    def pooled_of(*a, **kw):
        seen = []
        orig_go = A.go_verdict
        A.go_verdict = lambda pooled, thr: (seen.append(pooled) or {"verdict": "NO-GO"})
        try:
            PS.simulate(*a, **kw)
        finally:
            A.go_verdict = orig_go
        return seen[-1]
    ok_all = pooled_of(1 - 1e-12, 1e-12, st, pilots=1, boot_B=10)
    wrong_pos = pooled_of(1e-12, 1e-12, st, pilots=1, boot_B=10, fn_share=1.0)
    wrong_all = pooled_of(1e-12, 1e-12, st, pilots=1, boot_B=10, fn_share=0.0)
    all_ab = pooled_of(1 - 1e-12, 1e-12, st, pilots=1, boot_B=10, u=1.0)
    own = pooled_of(1e-12, 1 - 1e-12, st, pilots=1, boot_B=10, mode="own_errors")
    b2e = pooled_of(1 - 1e-12, 1e-12, st, pilots=1, boot_B=10, mode="b2_errors")
    check("R3-09 / R4-13 simulate matches the analysis's estimators: with Jev always right and B2 never, 81 covered are assigned and correct (P_assign 81 / 81, R_assign_res 81 / 81), the paired gain is over ALL 91 clusters (gain_res = gain_all = 1.0, n_res 91, u 0); with Jev always wrong and every wrong covered answer "
          "a false `none` (fn_share 1) only the 10 wrong assignments on the positives are `assigned` (0 correct) while R_assign_res is 0 / 81; with fn_share 0 all 91 are assigned",
          (ok_all["p_assign"]["n"], ok_all["p_assign"]["k"], ok_all["r_assign_res"]["n"], ok_all["r_assign_res"]["k"], ok_all["n_res"], ok_all["unresolved"]["k"]) == (81, 81, 81, 81, 91, 0) and abs(ok_all["gain_res"]["point"] - 1.0) < 1e-12 and abs(ok_all["gain_all"]["point"] - 1.0) < 1e-12
          and (wrong_pos["p_assign"]["n"], wrong_pos["p_assign"]["k"], wrong_pos["r_assign_res"]["n"], wrong_pos["r_assign_res"]["k"]) == (10, 0, 81, 0) and wrong_all["p_assign"]["n"] == 91,
          (ok_all["p_assign"], ok_all["gain_res"], wrong_pos["p_assign"], wrong_pos["r_assign_res"], wrong_all["p_assign"]))
    check("R4-13 (6) simulate's abstention: u = 1 abstains every cluster (u 91 / 91, n_res 0, gain_res undefined, gain_all 0, R_assign_res denominator 0); Jev abstaining on its own errors against an always-right B2 has gain_all = -1 (an abstention is a Jev miss); N4a abstains on "
          "B2's errors only up to int(0.15 x 91) = 13 clusters (78 resolved, gain_res 1.0, gain_all 78 / 91, 12 covered clusters leave R_assign_res)",
          (all_ab["unresolved"]["k"], all_ab["unresolved"]["n"], all_ab["n_res"], all_ab["r_assign_res"]["n"]) == (91, 91, 0, 0) and all_ab["gain_res"]["point"] is None and all_ab["gain_all"]["point"] == 0.0 and own["unresolved"]["k"] == 91 and own["gain_all"]["point"] == -1.0
          and b2e["unresolved"]["k"] == 13 and b2e["n_res"] == 78 and abs(b2e["gain_res"]["point"] - 1.0) < 1e-12 and abs(b2e["gain_all"]["point"] - 78 / 91) < 1e-12 and b2e["r_assign_res"]["n"] == 69,
          (all_ab["unresolved"], own["gain_all"], b2e["unresolved"], b2e["n_res"], b2e["r_assign_res"]))
    # R5-10: a non-degenerate traced pilot is rebuilt as cases and arm rows (raw fields from the latent) and analysis_x2d.summarise must equal the simulation's pooled dict
    def trace_pilot(verdict, *a, **kw):
        seen, tr = [], []
        orig_go = A.go_verdict
        A.go_verdict = lambda pooled, thr: (seen.append(pooled) or {"verdict": verdict})
        try:
            PS.simulate(*a, pilots=1, trace=tr, **kw)
        finally:
            A.go_verdict = orig_go
        return seen[-1], tr[0]

    def rebuild(tr):
        units, gold, jev_rows, b2_rows = [], {}, [], []
        for i, kind in enumerate(tr["kinds"]):
            cid, pos = "c%03d" % i, tr["is_pos"][i]
            units.append({"case_id": cid, "pr_key": "p%03d" % tr["pr_of"][i], "control": False, "set": "window", "stratum": 1})
            gold[cid] = ({"case_id": cid, "covered": False, "acceptable": [], "primary": None, "none_kind": "gap"} if pos else {"case_id": cid, "covered": True, "acceptable": ["A"], "primary": "A", "none_kind": None})
            rawr = (False, None) if pos and tr["J"][i] else ((True, "A") if tr["J"][i] else (True, "B"))
            out, area, why = {"abstain": ("unresolved", None, "below_t_choice"), "right": ("assigned", "A", None), "wrong_assigned": ("assigned", "B", None), "false_none": ("none", None, None),
                              "pos_right": ("none", None, None), "pos_wrong": ("assigned", "B", None)}[kind]
            jev_rows.append({"case_id": cid, "arm": "V", "output": out, "area_id": area, "unresolved_reason": why, "raw_covered": rawr[0], "raw_choice": rawr[1]})
            b2_rows.append({"case_id": cid, "arm": "B2", "output": ("none" if pos else "assigned") if tr["B"][i] else "assigned", "area_id": None if (pos and tr["B"][i]) else ("A" if tr["B"][i] else "B"), "unresolved_reason": None})
        return units, gold, {"V": jev_rows, "B2": b2_rows}

    st91 = PS.empirical_structure()
    spy_ok = True
    spy_detail = []
    for label, kw in (("random u = 0.10", dict(u=0.10)), ("own_errors", dict(mode="own_errors")), ("b2_errors", dict(mode="b2_errors"))):
        pl_, tr_ = trace_pilot("NO-GO", 0.88, 0.80, st91, boot_B=50, **kw)
        un_, gd_, ar_ = rebuild(tr_)
        sm_ = A.summarise(ar_, gd_, un_, jev="V", base="B2", B=50)
        same = ((pl_["p_assign"]["k"], pl_["p_assign"]["n"], pl_["r_assign_res"]["k"], pl_["r_assign_res"]["n"], pl_["n_res"], pl_["unresolved"]["k"]) == (sm_["p_assign"]["k"], sm_["p_assign"]["n"], sm_["r_assign_res"]["k"], sm_["r_assign_res"]["n"], sm_["n_res"], sm_["unresolved"]["k"])
                and all(abs(pl_[k][f] - sm_[k][f]) < 1e-12 for k, f in (("gain_res", "point"), ("gain_all", "point"), ("gain_guess", "point"), ("unresolved", "rate"))) and pl_["gain_guess"]["n"] == sm_["gain_guess"]["n"] == 91)
        spy_ok = spy_ok and same and sm_["unresolved"]["k"] > 0
        spy_detail.append((label, same, sm_["unresolved"]["k"], sm_["gain_guess"]["point"], sm_["gain_all"]["point"]))
    # at u = 0 the guard's rows are the resolved rows: the simulation's and summarise's lb95 of gain_guess equal the 95 % bootstrap bound of gain_res over the same rows (same seed), and the lb90 of gain_res agree
    pl0, tr0 = trace_pilot("GO", 0.95, 0.80, st91, boot_B=60)
    un0, gd0, ar0 = rebuild(tr0)
    sm0 = A.summarise(ar0, gd0, un0, jev="V", base="B2", B=60)
    rows0 = [r for r in A.pooling_rows([k for k in A.build_cases(un0, gd0, ar0) if not k["control"]], "V", "B2") if r["n_res"]]
    lb95_res = A.pr_cluster_bootstrap(rows0, A.gain_res_stat, B=60, levels=(0.95,))["lower"][0.95]
    check("R5-10 simulate matches analysis_x2d.summarise on a non-degenerate traced pilot (p_J 0.88, p_B2 0.80; random u = 0.10, own_errors, b2_errors): p_assign, r_assign_res, n_res, unresolved, gain_res / gain_all / gain_guess points and the gain_guess population agree, "
          "abstentions present; at u = 0 the guard's 95 % lower bound equals the gain_res 95 % bound over the same rows in the simulation AND in summarise, and the 90 % gain_res bounds agree",
          spy_ok and pl0["gain_guess"]["lb95"] is not None and abs(pl0["gain_guess"]["lb95"] - lb95_res) < 1e-12 and abs(sm0["gain_guess"]["lb95"] - lb95_res) < 1e-12 and pl0["gain_res"]["lb90"] is not None
          and abs(pl0["gain_res"]["lb90"] - sm0["gain_res"]["lb90"]) < 1e-12 and sm0["unresolved"]["k"] == 0, (spy_detail, pl0["gain_guess"], sm0["gain_guess"], lb95_res))
    # R3-16 (as extended by R4-08 and R4-13 (7)): the round-0 critique guards live in validation/r4_checks.py `critique_checks` (called by r4_infra_unit_checks at the end of pre_e2e_checks)


def pre_e2e_checks(tmp, skip):
    provenance_checks(tmp)
    thresholds_checks()
    guide_checks()
    repo, shas = frontend_checks(tmp)
    if "prepass" not in skip:
        prepass_checks(tmp)
    if "consensus" not in skip:
        consensus_checks(tmp)
    if "jev" not in skip:
        jev_checks(tmp)
    label_checks(tmp)
    primary_disagreement_checks(tmp)
    stats_primitives()
    verdict_checks()
    perception_order_checks()
    kappa_answer_checks()
    labels_check_checks(tmp)
    verdict_primary_checks(tmp)
    n1_checks(tmp)
    futility_checks()
    p2_checks()
    if "power" not in skip:
        power_checks(tmp)
    res, mem = units_checks(tmp, repo, shas)
    gold_membership_checks(tmp, res, mem, repo)
    round_fix_checks(tmp, repo, shas, res, mem)
    modification_only_checks(tmp, repo, shas)
    compose_checks(tmp)
    compose_tie_checks()
    tune_checks(tmp)
    versioning_checks(tmp)
    onboarding_checks(tmp)
    collector_checks(tmp)
    gates_refusal_check(tmp)
    guard_checks(tmp)
    r2_misc_checks(tmp)
    r3_checks(tmp, repo, res)
    r4_checks.r4_infra_unit_checks(sys.modules[__name__], tmp, repo, res)
    regate_checks.unit_checks(sys.modules[__name__], tmp)


def main():
    tmp = tempfile.mkdtemp(prefix="x2d-validate-")
    e = None
    try:
        skip = set(filter(None, os.environ.get("X2D_VALIDATE_SKIP", "").split(",")))  # development only (prepass, consensus, jev, power, pre = everything before the end-to-end part); the reviewed run skips nothing
        if "pre" not in skip:
            pre_e2e_checks(tmp, skip)
        e = E2E(tmp)
        if e2e_round0(tmp, e):
            g = e2e_gates(tmp, e)
            if g:
                regate_checks.e2e_regate(sys.modules[__name__], tmp, e)
                e2e_rounds(tmp, e, g)
    finally:
        if e is not None:
            e.close()
        if not os.environ.get("X2D_VALIDATE_KEEP"):  # development only: keep the temporary directory for inspection
            shutil.rmtree(tmp, ignore_errors=True)
        else:
            print("kept " + tmp)
    print("\n%d checks, %d failed" % (COUNT[0], len(FAILS)))
    for f in FAILS:
        print("FAILED: " + f)
    if FAILS:
        sys.exit(1)
    print("all validation checks passed")


if __name__ == "__main__":
    main()
