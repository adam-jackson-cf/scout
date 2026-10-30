"""X2d shared constants, pins, paths and small I/O helpers (README section Modules). Standard library only (run with .tools/venv/bin/python).

Every stage first calls `verify_pins()`. The X2 files that X2d imports (credential scrub, Clawpatch map environment, review-safety check) are the frozen X2 files, pinned
by digest exactly as X2c pins them (`evaluation/x2c-mapper-comparison/x2c.py`). Nothing here calls a model.
"""
import glob
import hashlib
import json
import os
import re
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", ".."))
X2_DIR = os.path.join(ROOT, "evaluation", "x2-area-stability")
DATA = os.environ.get("X2D_DATA") or os.path.join(HERE, "data")
SIGNOFF = os.path.join(HERE, "signoff.json")  # written by the council's final sign-off round: {"files": {rel: sha}, "code_under_test": {rel: sha}}

# ---- pins (PREREGISTRATION sections 2, 3, 7, 9)
JEV_MODEL = "typesafe/jev-1.14-20261014"
JEV_ENDPOINT = "https://openrouter.ai/api/alpha/decisions"  # internal/decisions.DefaultEndpoint
MAPPER = {"provider": "codex", "model": "gpt-6-luna", "effort": "xhigh"}
REPO_NAME = "fastapi"
C0 = "1d50bad4555b951d5fa48eb8787fd5c39ed7effb"
TIP = "79cab5bed28b24d92c185d236feb33f5783dfeb3"
SEED = 20261029
PR_CAP = 10
GUIDE_VERSION = "x2d-association-v1/guide-1"
QUESTION_VERSION = "v1"
TAU = 0.5  # consensus co-association threshold (PREREGISTRATION 4.2)
REFERENCE_SUBSET = 20  # reference arm: clusters per round
FLIP_FRACTION_MOD = 5  # second permutation for sha256("x2d-flip" + cluster id) mod 5 == 0
N_ROUNDS = 10
REFINEMENT_ROUNDS = list(range(1, 8))
FROZEN_ROUNDS = [8, 9, 10]
CONTRAST_ROUNDS = [1, 10]
MIN_KAPPA = 0.6
MINIMUMS = {"P_assign": 0.90, "R_assign": 0.80}  # R_assign = R_assign_res: assigned-correct / (gold-covered AND resolved), R4-13; P_new / R_new are reported, never a minimum
UNRESOLVED_RATE_MAX = 0.15  # R4-13 (1) g: abstained / GO population (the cap); tune.composed_dev enforces it on dev
RESOLVED_MIN = 50  # R4-13 (1) i: resolved paired clusters needed for GO
NEW_AREA_GATE_MIN = 0.80  # round-0 gate bound (planted_new, P2): formerly MINIMUMS["R_new"]; a round-0 gate, not a verdict condition
# ---- rev 13 (R-053): the one permitted re-gate of P1 and P2 on fresh items (PREREGISTRATION 5, rev 13). Every bar is an existing bar: P1's Jev recovery 0.60 / baseline 0.20, the R_assign minimum 0.80, the P2 gate
# bound 0.80 and the GO abstention cap 0.15 (reasoning in PREREGISTRATION 5, rev 13).
REGATE_SUBDIR = "regate"  # D/rounds/0/regate/ is a self-contained data root (catalogue/, rounds/0/): every re-gate stage runs with X2D_DATA pointing at it; nothing outside it is ever written
REGATE_GATE_SET = "x2d-regate-v1"  # the gate-set tag of D/rounds/0/regate/gates.json; the rev-12 gates.json carries no tag and is never rewritten
REGATE_P1_N = 40  # P1 items built (>= 40, PREREGISTRATION 5 rev 13)
REGATE_P1_STRATA = (1, 1, 1, 2, 3)  # target stratum cycled over the P1 items (24 / 8 / 8 of 40 when every target exists; a target falls back to another kind when none is free): stratum 1 decides the gate, at least half of the items must be stratum 1
REGATE_P1_STRATUM1_MIN = REGATE_P1_N // 2  # stratum-1 items needed in the build, and eligible (gold contains A) in the analysis; fewer = UNMEASURED
REGATE_P1_ACC_MIN = MINIMUMS["R_assign"]  # P1: V-sem accuracy on the answered stratum-1 items (assigned in the labellers' acceptable set / resolved): the R_assign minimum (R-051 judges accuracy on answered clusters)
REGATE_P1_ABSTAIN_MAX = UNRESOLVED_RATE_MAX  # P1 (stratum 1) and P2 (positives): abstained / items <= the GO abstention cap g
REGATE_BASELINE_MAX = 0.20  # P1: B1 and B2 recovery of an acceptable area over all eligible P1 items (the rev-12 bar, unchanged)
REGATE_P2_CLONES = 40  # P2 clones built. Sized from the round-0 yield of 4 direct-label positives in 12 clones (1/3). P(X >= 8) = 0.98 (0.82 at 1/4) for independent clones; the clones share 16 ablations, and with yield fully correlated within an area P(X >= 8) ~ 0.88 at 1/3 and ~ 0.71 at 1/4 [INFERENCE]; a shortfall is caught before any Jev request (run.sh regate-jev, analysis_x2d.py --regate-precheck)
REGATE_P2_AREA_MIN = 2  # an ablated area owns >= 2 fresh candidates left after P1 (its clones share one ablation: the catalogue loses few areas)
REGATE_P2_PER_AREA = 4  # at most this many clones per ablated area
REGATE_P2_POSITIVES_MIN = 8  # direct-label positives (a clone whose own blind label on catalogue-p2 is not covered, gap or new) needed for a decision; fewer = UNMEASURED, which does not pass
REGATE_P2_NONE_MIN = NEW_AREA_GATE_MIN  # P2: V-sem `none` / answered positives
REGATE_P2_ABSTAIN_MAX = UNRESOLVED_RATE_MAX
GO_GUARD_LEVEL = 0.95  # R5-01 h: one-sided level of the PR-cluster-bootstrap lower bound of gain_guess (thresholds.json study.go.guard_level)
POWER_CHECK_KEYS = ("P3_u0", "P3_u05", "P3_u10", "N3", "N3_PR", "N4a", "N4b", "N4b_cap", "N4b_rho0")  # power.json `checks`, exactly these (R5-01, R5-02)
POWER_NULL_KEYS = ("N3", "N3_PR", "N4a", "N4b", "N4b_cap", "N4b_rho0")  # a false null check is a stop no critique line clears (R5-02); a false P3 check needs `user_decision: proceed`
CAPS = {"usd_per_round": 2.0, "usd_total": 20.0, "clawpatch_maps_round_0": 16}
CLI_SESSION_CEILING = 500  # hard ceiling of label_x2d.cap_for (R1-05); the session cap is the per-round formula, never a fixed number
CLAWPATCH_VERSION = "0.7.3"
DEFAULT_SETTINGS = {  # the refinement menu (PREREGISTRATION 10 (ii), (iii)); a round's settings live in data/rounds/<r>/settings.json
    "sample_files_per_area": 5, "excerpt_lines": 60, "v_sib": True, "include_pr_title": True,  # include_pr_title=False removes state.pr.title from V-sem/V-sib (V-title arm always on)
    "field_order": ["unit", "pr", "diff_summary", "catalogue"], "order_rule": "permutation:unit_id",
}
SETTING_RANGES = {"sample_files_per_area": (3, 8), "excerpt_lines": (30, 120)}
TUNE_SCRIPT = os.environ.get("X2D_TUNE_SCRIPT") or os.path.join(os.path.expanduser("~"), ".omp", "agent", "skills", "jev-actions", "scripts", "tune_threshold.py")
# Council checkpoint critiques (PREREGISTRATION 10; R1-09, R2-16): a critique file is S/reviews/pilot-round-<n>.md (dir overridable by X2D_REVIEWS_DIR, honoured only in a
# validation environment, see validation_env_ok). run.sh gates round r (1..10) and `verdict` on the critique of round r-1 (verdict: pilot-round-10.md); `critique_problems`
# lists what a file lacks: exactly one line starting `Verdict:` equal to the wanted one, exactly one `version_sha256: <sha of rounds/<n>/version.json>`, for n <= 7 exactly one
# `next_inputs_sha256: <digest of the inputs the next round will use>` (n = 7: round 7's own inputs, v_7 frozen), for n = 0 also `real_none_proxy: <int>` (= gates.json; INFO, not a decision trigger),
# `p3_real_structure: <float>` (= power.json p3_real_structure.share), `p3_abstention: 0.05=<float> 0.10=<float>`, `n4_max: <float>` (= power.json, 4 decimals), `frozen_go_population: 8=<n>/<prs> 9=<n>/<prs>
# 10=<n>/<prs>` (= rounds/0/k23.json, core-gate units excluded, R4-08), `user_decision: proceed` when a power.json P3 check is false or a frozen round has < 20 clusters or < 6 PRs (a false NULL check
# (POWER_NULL_KEYS) is a stop no line clears, R5-02; `checks` must hold exactly POWER_CHECK_KEYS), and `p2_unmeasured: accepted` when the P2 gate is UNMEASURED (R3-16, R3-04, R4-13 (7)).
CRITIQUE_VERDICT_REFINEMENT = "Verdict: ZERO ISSUES"  # critique of rounds 0..7
CRITIQUE_VERDICT_FROZEN = "Verdict: DESCRIPTIVE"  # critique of rounds 8..10
CRITIQUE_VERSION_LINE = "version_sha256: "  # followed by the SHA-256 of rounds/<n>/version.json
CRITIQUE_INPUTS_LINE = "next_inputs_sha256: "  # followed by the inputs digest (inputs_digest) of the inputs the next round uses
CRITIQUE_NONE_PROXY_LINE = "real_none_proxy: "  # round-0 critique only
CRITIQUE_P3_LINE = "p3_real_structure: "  # round-0 critique only
CRITIQUE_ABSTAIN_LINE = "p3_abstention: "  # round-0 critique only: `p3_abstention: 0.05=<share> 0.10=<share>` (power.json p3_abstention, the P3 GO share at 5 % / 10 % unresolved)
CRITIQUE_N4_LINE = "n4_max: "  # round-0 critique only: the maximum GO share over the N4a / N4b / N4b-cap / N4b-rho0 abstention-gaming cells of power.json
CRITIQUE_POP_LINE = "frozen_go_population: "  # round-0 critique only: `8=<n>/<prs> 9=<n>/<prs> 10=<n>/<prs>` from rounds/0/k23.json (core-gate units excluded, R4-08)
CRITIQUE_DECISION_LINE = "user_decision: "  # round-0 critique only, required when a power.json P3 check is false or a frozen round has < FROZEN_MIN_CLUSTERS clusters or < FROZEN_MIN_PRS PRs (R4-13 (7), R4-08, R5-02)
CRITIQUE_P2_LINE = "p2_unmeasured: "  # round-0 critique only, required when the P2 gate is UNMEASURED (R3-04)
FROZEN_MIN_CLUSTERS = 20  # informative-round minimum (PREREGISTRATION 9): scored GO-population clusters per frozen round
FROZEN_MIN_PRS = 6  # informative-round minimum: distinct PRs per frozen round
FAN_OUT_ARM, VSEM_ARM = "fan-out", "V-sem"
V7_RESCORE_SUB = "jev-v7"  # D/rounds/<rr>/jev-v7/scores.jsonl: rounds 0-7 re-scored under v_7 (dev only) before round 8 by `run.sh rescore-v7`; round 8 tunes once on these rows (R1-10, R2-02)
GUIDE_VERSION_PREFIX = "x2d-association-v1/guide-"


def _loopback(url):
    from urllib.parse import urlparse
    try:
        return urlparse(url).hostname in ("127.0.0.1", "localhost", "::1")
    except ValueError:
        return False


def validation_env_ok():
    """R2-16: X2D_SKIP_SIGNOFF=1 and X2D_REVIEWS_DIR are honoured only in a validation environment: X2D_JEV_ENDPOINT a loopback URL and X2D_DATA outside S/data."""
    data = os.path.abspath(os.environ.get("X2D_DATA") or DATA)
    live = os.path.join(HERE, "data")
    outside = data != live and not data.startswith(live + os.sep)
    return bool(os.environ.get("X2D_JEV_ENDPOINT")) and _loopback(os.environ["X2D_JEV_ENDPOINT"]) and outside


def require_signoff():
    """-> whether the sign-off record is required: False only for X2D_SKIP_SIGNOFF=1 in a validation environment; the flag on a live run is refused (exit)."""
    if os.environ.get("X2D_SKIP_SIGNOFF") == "1":
        if not validation_env_ok():
            sys.exit("refusing: X2D_SKIP_SIGNOFF=1 is honoured only with a loopback X2D_JEV_ENDPOINT and X2D_DATA outside %s" % os.path.join(HERE, "data"))
        return False
    return True


def reviews_dir():
    d = os.environ.get("X2D_REVIEWS_DIR")
    if d:
        if not validation_env_ok():
            sys.exit("refusing: X2D_REVIEWS_DIR is honoured only with a loopback X2D_JEV_ENDPOINT and X2D_DATA outside %s" % os.path.join(HERE, "data"))
        return d
    return os.path.join(HERE, "reviews")


def critique_path(n, reviews_dir_=None):
    return os.path.join(reviews_dir_ or reviews_dir(), "pilot-round-%d.md" % n)


# ---- frozen X2 files that X2d imports or runs unchanged (digests = X2's sign-off, the same pins as X2c)
X2_PINS = {
    "common.py": "e526804c331f3eab07f7003e797b86b4e793c3cf0a351e69cdf695286f83efa3",
    "label.py": "87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f",
    "validation/stub_clawpatch.js": "0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226",
    "validation/credcheck/main.go": "e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779",
}

# ---- Scout code under test (PREREGISTRATION 2)
GO_POLICY = os.path.join(ROOT, "internal", "policy", "consent.go")
GO_DECISIONS_DIR = os.path.join(ROOT, "internal", "decisions")

# ---- study files (README "Frozen files"): everything but data/, reviews/, notes, results and the sign-off record
STUDY_MODULES = [
    "x2d.py", "collect_onboard.py", "onboard/main.go", "quality.py", "build_units.py", "b2.py", "baselines.py", "build_controls.py", "label_x2d.py", "gold.py",
    "jev/main.go", "reference_arm.py", "tune.py", "compose.py", "analysis_x2d.py", "power_sim.py", "versioning.py", "run.sh",
]
STUDY_STATIC = ["PREREGISTRATION.md", "README.md", "GUIDE.md", "thresholds.json", "frontend.py", "prepass.py", "scopes.py"]
STUDY_GLOBS = ["questions/*.json", "prepass/*.json", "prepass/*.jsonl", "measurement/*", "validation/*"]


def sha256_file(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def sha256_bytes(b):
    return hashlib.sha256(b).hexdigest()


def sha256_text(s):
    return hashlib.sha256(s.encode("utf-8")).hexdigest()


def frozen_files():
    """Sorted relative paths of every frozen study file that exists (a module not yet written is simply absent)."""
    out = set()
    for rel in STUDY_STATIC + STUDY_MODULES:
        if os.path.isfile(os.path.join(HERE, rel)):
            out.add(rel)
    for g in STUDY_GLOBS:
        for p in glob.glob(os.path.join(HERE, g)):
            if os.path.isfile(p) and "__pycache__" not in p and not p.endswith(".pyc"):
                out.add(os.path.relpath(p, HERE))
    return sorted(out)


def go_code_digests():
    """SHA-256 of the Scout code under test (PREREGISTRATION 2): every non-test file of internal/decisions, internal/policy/consent.go, internal/areas/areas.go,
    internal/clawpatch/records.go and client.go, go.mod, go.sum, scripts/clawpatch-safe.patch, scripts/build-clawpatch.sh. -> {relpath: sha256}."""
    files = sorted(os.path.join(dp, n) for dp, _dn, fn in os.walk(GO_DECISIONS_DIR) for n in fn if (n.endswith(".go") and not n.endswith("_test.go")) or n.endswith(".json"))
    files += [GO_POLICY, os.path.join(ROOT, "internal", "areas", "areas.go"), os.path.join(ROOT, "internal", "clawpatch", "records.go"),
              os.path.join(ROOT, "internal", "clawpatch", "client.go"), os.path.join(ROOT, "go.mod"), os.path.join(ROOT, "go.sum"),
              os.path.join(ROOT, "scripts", "clawpatch-safe.patch"), os.path.join(ROOT, "scripts", "build-clawpatch.sh")]
    return {os.path.relpath(p, ROOT): sha256_file(p) for p in files}


def go_jev_pin(policy=None):
    m = re.search(r'^\s*DefaultModel\s*=\s*"([^"]*)"', open(policy or GO_POLICY).read(), re.M)
    return m.group(1) if m else None


def verify_pins(require_signoff=True, x2_dir=None, root=None):
    """-> list of problems (empty when every pin holds): the X2 files, the Jev pin of internal/policy/consent.go, and, against the sign-off record,
    every frozen study file and every code-under-test blob. `require_signoff=False` (validation only) skips the record; a stage that spends or writes data
    runs with the default and refuses when the record is absent."""
    bad = []
    for rel, want in sorted(X2_PINS.items()):
        p = os.path.join(x2_dir or X2_DIR, rel)
        got = sha256_file(p) if os.path.exists(p) else None
        if got != want:
            bad.append("X2 %s: %s != pinned %s" % (rel, got, want))
    pin = go_jev_pin()
    if pin != JEV_MODEL:
        bad.append("internal/policy/consent.go DefaultModel is %r, preregistered %r" % (pin, JEV_MODEL))
    if not os.path.exists(SIGNOFF):
        if require_signoff:
            bad.append("no sign-off record %s: stages that write data refuse before the council's sign-off" % SIGNOFF)
        return bad
    rec = json.load(open(SIGNOFF))
    if root is None and sorted(frozen_files()) != sorted(rec.get("files", {})):  # R2-16: a frozen file added or removed after sign-off is a violation
        bad.append("frozen file set differs from the sign-off record: %s" % sorted(set(frozen_files()) ^ set(rec.get("files", {}))))
    for rel, want in sorted(rec.get("files", {}).items()):
        p = os.path.join(root or HERE, rel)
        got = sha256_file(p) if os.path.exists(p) else None
        if got != want:
            bad.append("frozen file %s: %s != signed-off %s" % (rel, got, want))
    code = go_code_digests()
    for rel, want in sorted(rec.get("code_under_test", {}).items()):
        if code.get(rel) != want:
            bad.append("code under test %s: %s != signed-off %s" % (rel, code.get(rel), want))
    want_tune = rec.get("tune_script_sha256")  # the external tune_threshold.py is pinned at sign-off (R1-10)
    got_tune = sha256_file(TUNE_SCRIPT) if os.path.exists(TUNE_SCRIPT) else None
    if want_tune is None:
        bad.append("sign-off record lacks tune_script_sha256 (tune_threshold.py must be pinned at sign-off)")
    elif got_tune != want_tune:
        bad.append("tune_threshold.py %s: %s != signed-off %s" % (TUNE_SCRIPT, got_tune, want_tune))
    return bad


# ---- paths of the data directory (every stage reads and writes here; X2D_DATA overrides)
def data_path(*parts):
    return os.path.join(DATA, *parts)


def round_dir(r, data=None):
    return os.path.join(data or DATA, "rounds", str(r))


def regate_root(data=None):
    """D/rounds/0/regate: the one permitted re-gate's own data root (rev 13, R-053). It holds `catalogue/` and `rounds/0/` like a data directory; every re-gate stage runs with X2D_DATA = this
    directory, so nothing outside it can be written; the study's round-0 files stay as they are."""
    return os.path.join(round_dir(0, data), REGATE_SUBDIR)


def is_regate_root(data=None):
    parts = os.path.abspath(data or DATA).split(os.sep)
    return len(parts) >= 4 and parts[-1] == REGATE_SUBDIR and parts[-2] == "0" and parts[-3] == "rounds"


def main_data(data=None):
    """The study data directory: `data` itself, or the one a re-gate root lives in (D/rounds/0/regate -> D)."""
    d = os.path.abspath(data or DATA)
    return os.path.dirname(os.path.dirname(os.path.dirname(d))) if is_regate_root(d) else d


def gate_problems(data=None):
    """-> what stops round 1 on the round-0 gates (empty = round 1 may start). Any round-0 gate other than P1 / P2 that FAILed stops the pilot. P1 / P2 are the two gates rev 13 may re-test once, on
    fresh items: when either FAILed in round 0 (rev 12 bytes), `rounds/0/regate/gates.json` must exist, carry the gate-set tag REGATE_GATE_SET and show PASS for BOTH P1 and P2 (UNMEASURED or FAIL
    stops the pilot; there is no second re-gate). The rev-12 gates.json is read, never rewritten."""
    main = main_data(data)
    gp = os.path.join(round_dir(0, main), "gates.json")
    if not os.path.isfile(gp) or os.path.getsize(gp) == 0:
        return ["%s missing: run the gates first" % gp]
    gates = (json.load(open(gp)).get("gates")) or {}
    if not gates:
        return ["%s has no gate records: run the gates first" % gp]
    failed = sorted(k for k, v in gates.items() if v.get("status") == "FAIL")
    if not failed:
        return []
    if set(failed) - {"P1", "P2"}:
        return ["round-0 gate(s) failed: " + ", ".join(failed)]
    rp = os.path.join(regate_root(main), "gates.json")
    head = "round-0 gate(s) failed: %s (rev 12 bytes); round 1 needs the one permitted re-gate on fresh items to PASS both P1 and P2 (PREREGISTRATION 5, rev 13)" % ", ".join(failed)
    if not os.path.isfile(rp) or os.path.getsize(rp) == 0:
        return [head + ": %s does not exist (run.sh regate-build / -label / -jev / -analyse)" % rp]
    rj = json.load(open(rp))
    bad = []
    if rj.get("gate_set") != REGATE_GATE_SET:
        bad.append("%s carries gate_set %r, not %r" % (rp, rj.get("gate_set"), REGATE_GATE_SET))
    rg = rj.get("gates") or {}
    for k in ("P1", "P2"):
        st = (rg.get(k) or {}).get("status")
        if st != "PASS":
            bad.append("re-gate %s is %s: a re-gate that does not PASS stops the pilot (one shot, no further attempt)" % (k, st))
    return [head + ": " + "; ".join(bad)] if bad else []



def regate_frozen_problems(data=None):
    """-> differences between the re-gate root's copy of the frozen Jev side (the round's inputs: guide and v1 questions, settings.json, thresholds.json, the catalogue) and the study's round-0 originals, plus a
    missing marker or a threshold that is not the round-0 provisional 0.70 (empty = the Jev side is the one round 0 used, byte for byte). Every re-gate stage after the build checks this."""
    main = main_data(data)
    root = regate_root(main)
    rd0, rr = round_dir(0, main), os.path.join(root, "rounds", "0")
    bad = []
    if not os.path.isfile(os.path.join(root, "regate.json")):
        bad.append("%s missing: run.sh regate-build first" % os.path.join(root, "regate.json"))
    pairs = [(os.path.join(rd0, "settings.json"), os.path.join(rr, "settings.json")), (os.path.join(rd0, "thresholds.json"), os.path.join(rr, "thresholds.json")),
             (catalogue_path(main), catalogue_path(root))]
    for a, b in pairs:
        if not os.path.isfile(a) or not os.path.isfile(b) or sha256_file(a) != sha256_file(b):
            bad.append("%s differs from %s" % (b, a))
    da, db = inputs_digest(os.path.join(rd0, "inputs")), inputs_digest(os.path.join(rr, "inputs"))
    if da is None or da != db:
        bad.append("%s (inputs digest %s) differs from %s (%s)" % (os.path.join(rr, "inputs"), db, os.path.join(rd0, "inputs"), da))
    tp = os.path.join(rr, "thresholds.json")
    if os.path.isfile(tp):
        for name, dec in (json.load(open(tp)).get("decisions") or {}).items():
            if dec.get("source") != "provisional-0.7" or dec.get("threshold") != 0.7:
                bad.append("%s: decision %s is not the round-0 provisional 0.70 threshold" % (tp, name))
    return bad


def primary_arm(data=None):
    """The preregistered primary Jev arm (PREREGISTRATION 6): "fan-out" iff rounds/0/gates.json `fan_out_primary` is truthy (K5 failed), else "V-sem"."""
    p = os.path.join(round_dir(0, data), "gates.json")
    if os.path.exists(p):
        with open(p) as f:
            if json.load(f).get("fan_out_primary"):
                return FAN_OUT_ARM
    return VSEM_ARM


def catalogue_path(data=None):
    return os.path.join(data or DATA, "catalogue", "catalogue.json")


def read_catalogue(data=None):
    with open(catalogue_path(data)) as f:
        return json.load(f)


def catalogue_digest(data=None):
    """SHA-256 of catalogue.json (the catalogue digest of every score-file header)."""
    return sha256_file(catalogue_path(data))


def round_settings(r, data=None):
    """The round's menu settings: data/rounds/<r>/settings.json over DEFAULT_SETTINGS (written by versioning.freeze_version)."""
    p = os.path.join(round_dir(r, data), "settings.json")
    s = dict(DEFAULT_SETTINGS)
    if os.path.exists(p):
        s.update(json.load(open(p)))
    return s


# ---- per-round inputs (R2-01): the guide and the question files a round is labelled and scored with live in data/rounds/<r>/inputs/{GUIDE.md,questions/*.json}; the study-directory
# copies are the frozen v_0 base. versioning.ensure_inputs creates them (round 0/1: copied from the study directory; r >= 2: from round r-1 unless the drafter's proposal sits there).
def inputs_dir(r, data=None):
    return os.path.join(round_dir(r, data), "inputs")


def guide_path(r, data=None):
    return os.path.join(inputs_dir(r, data), "GUIDE.md")


def questions_dir(r, data=None):
    return os.path.join(inputs_dir(r, data), "questions")


def inputs_digest(d):
    """SHA-256 of the sorted "<sha256>  <relative path>" list (one per line, '/' separators) of every file under directory d; None when d does not exist."""
    if not os.path.isdir(d):
        return None
    lines = sorted("%s  %s" % (sha256_file(os.path.join(dp, n)), os.path.relpath(os.path.join(dp, n), d).replace(os.sep, "/")) for dp, _dn, fn in os.walk(d) for n in fn)
    return sha256_text("\n".join(lines) + "\n")


def proposed_inputs_dir(r, data=None):
    """The inputs directory freeze will use for round r: rounds/<r>/inputs when it exists, else the one it is copied from (round 1: round 0's, r >= 2: round r-1's)."""
    if os.path.isdir(inputs_dir(r, data)):
        return inputs_dir(r, data)
    return inputs_dir(r - 1, data) if r >= 1 else None


def guide_version(r, data=None):
    """`x2d-association-v1/guide-<first 12 hex of the SHA-256 of rounds/<r>/inputs/GUIDE.md>` (R2-01)."""
    p = guide_path(r, data)
    if not os.path.isfile(p):
        sys.exit("error: %s is missing (versioning.py --round %s --inputs creates the round's inputs)" % (p, r))
    return GUIDE_VERSION_PREFIX + sha256_file(p)[:12]


def line_pop_text(pops):
    """The text after `frozen_go_population: ` for {round: (n, prs)}: `8=<n>/<prs> 9=<n>/<prs> 10=<n>/<prs>`."""
    return " ".join("%d=%d/%d" % (r, n, p) for r, (n, p) in sorted(pops.items()))


def critique_problems(n, data=None, reviews_dir_=None):
    """-> list of what is wrong with the council critique of round n (empty = it unlocks the next stage). See the comment at CRITIQUE_*."""
    p = critique_path(n, reviews_dir_)
    if not os.path.isfile(p) or os.path.getsize(p) == 0:
        return ["the council critique %s is missing" % p]
    lines = [l.strip() for l in open(p).read().splitlines()]
    bad = []

    def one(prefix):
        got = [l for l in lines if l.startswith(prefix.rstrip())]
        return got[0] if len(got) == 1 else None, len(got)

    want = CRITIQUE_VERDICT_REFINEMENT if n <= 7 else CRITIQUE_VERDICT_FROZEN
    v, k = one("Verdict:")
    if k != 1 or v != want:
        bad.append("%s must contain exactly one line starting 'Verdict:' and it must be %r (found %d)" % (p, want, k))
    vp = os.path.join(round_dir(n, data), "version.json")
    if not os.path.isfile(vp):
        bad.append("%s is missing" % vp)
    else:
        line, k = one(CRITIQUE_VERSION_LINE)
        if k != 1 or line != CRITIQUE_VERSION_LINE + sha256_file(vp):
            bad.append("%s must contain exactly one line %r (the SHA-256 of %s)" % (p, CRITIQUE_VERSION_LINE + sha256_file(vp), vp))
    if n <= 7:
        nv = os.path.join(round_dir(n + 1, data), "version.json") if n < 7 else None
        if nv and os.path.isfile(nv):  # round n+1 is frozen: the critique pins the inputs it was frozen with; later edits are policed by versioning --drift, a recorded clarification (R2-01 (4)) moves the live digest legitimately
            d, dig = nv, json.load(open(nv)).get("inputs_sha256")
        else:
            d = inputs_dir(n, data) if n == 7 else proposed_inputs_dir(n + 1, data)
            dig = inputs_digest(d) if d else None
        line, k = one(CRITIQUE_INPUTS_LINE)
        if dig is None:
            bad.append("the inputs the council approves (%s) do not exist" % d)
        elif k != 1 or line != CRITIQUE_INPUTS_LINE + dig:
            bad.append("%s must contain exactly one line %r (the inputs digest of %s)" % (p, CRITIQUE_INPUTS_LINE + dig, d))
    if n == 0:
        import re
        vals = {}
        for prefix, cast in ((CRITIQUE_NONE_PROXY_LINE, int), (CRITIQUE_P3_LINE, float), (CRITIQUE_N4_LINE, float)):
            line, k = one(prefix)
            try:
                if k != 1:
                    raise ValueError(prefix)
                vals[prefix] = cast(line[len(prefix):])
            except ValueError:
                bad.append("%s must contain exactly one parsable line %r<number>" % (p, prefix))
        line, k = one(CRITIQUE_ABSTAIN_LINE)
        m_ab = re.fullmatch(re.escape(CRITIQUE_ABSTAIN_LINE) + r"0\.05=(\S+) 0\.10=(\S+)", line or "") if k == 1 else None
        ab = None
        if m_ab:
            try:
                ab = (float(m_ab.group(1)), float(m_ab.group(2)))
            except ValueError:
                ab = None
        if ab is None:
            bad.append("%s must contain exactly one line '%s0.05=<share> 0.10=<share>'" % (p, CRITIQUE_ABSTAIN_LINE))
        line, k = one(CRITIQUE_POP_LINE)
        m_pop = re.fullmatch(re.escape(CRITIQUE_POP_LINE) + r"8=(\d+)/(\d+) 9=(\d+)/(\d+) 10=(\d+)/(\d+)", line or "") if k == 1 else None
        if not m_pop:
            bad.append("%s must contain exactly one line '%s8=<n>/<prs> 9=<n>/<prs> 10=<n>/<prs>'" % (p, CRITIQUE_POP_LINE))
        # R3-16, R4-13 (7): the lines equal the recorded gates.json / power.json / k23.json values; a user decision is recorded when a binding power cell is below its bar or a frozen round is thin (R4-08)
        gp, pp = os.path.join(round_dir(0, data), "gates.json"), os.path.join(data or DATA, "power.json")
        kp = os.path.join(round_dir(0, data), "k23.json")
        if not os.path.isfile(pp):
            bad.append("%s is missing (run.sh power)" % pp)
        if not os.path.isfile(gp):
            bad.append("%s is missing (run.sh gates)" % gp)
        if not os.path.isfile(kp):
            bad.append("%s is missing (run.sh gates)" % kp)
        if os.path.isfile(pp) and os.path.isfile(gp) and os.path.isfile(kp):
            gj, pj, kj = json.load(open(gp)), json.load(open(pp)), json.load(open(kp))
            proxy = (gj.get("gates", {}).get("real_none_proxy") or {}).get("value")
            p3 = (pj.get("p3_real_structure") or {}).get("share")
            pab = pj.get("p3_abstention") or {}
            n4 = pj.get("n4_max")
            for prefix, want, nd in ((CRITIQUE_NONE_PROXY_LINE, proxy, None), (CRITIQUE_P3_LINE, p3, 4), (CRITIQUE_N4_LINE, n4, 4)):
                got = vals.get(prefix)
                if got is None:
                    continue
                if want is None or (round(got, nd) != round(want, nd) if nd is not None else got != want):
                    bad.append("%s%s does not equal the recorded value %s (%s)" % (prefix, got, want, gp if prefix == CRITIQUE_NONE_PROXY_LINE else pp))
            if ab is not None:
                for got, key in zip(ab, ("0.05", "0.10")):
                    want = pab.get(key)
                    if want is None or round(got, 4) != round(want, 4):
                        bad.append("%s%s=%s does not equal the recorded value %s (%s)" % (CRITIQUE_ABSTAIN_LINE, key, got, want, pp))
            pops, thin = {}, False
            for r_ in FROZEN_ROUNDS:
                pr = (kj.get("per_round") or {}).get(str(r_)) or {}
                prs = (kj.get("frozen_pr_counts") or {}).get(str(r_))
                if "n" not in pr or prs is None:
                    bad.append("%s has no per_round[%d].n / frozen_pr_counts[%d]: rebuild it with build_units.py --gates" % (kp, r_, r_))
                    continue
                pops[r_] = (pr["n"] - (pr.get("core_gate") or 0), len(prs))  # the GO population: core-gate units excluded (R3-11)
                thin = thin or pops[r_][0] < FROZEN_MIN_CLUSTERS or pops[r_][1] < FROZEN_MIN_PRS
            if m_pop and len(pops) == len(FROZEN_ROUNDS):
                got_pop = {r_: (int(m_pop.group(1 + 2 * i)), int(m_pop.group(2 + 2 * i))) for i, r_ in enumerate(FROZEN_ROUNDS)}
                if got_pop != pops:
                    bad.append("%s%s does not equal the recorded values %s (%s)" % (CRITIQUE_POP_LINE, line_pop_text(got_pop), line_pop_text(pops), kp))
            checks = pj.get("checks")
            if not isinstance(checks, dict) or set(checks) != set(POWER_CHECK_KEYS) or any(not isinstance(v, bool) for v in checks.values()):
                bad.append("%s `checks` must be booleans for exactly %s (schema x2d-power-v5): re-run run.sh power" % (pp, ", ".join(POWER_CHECK_KEYS)))
                failing, null_failing = [], []
            else:
                failing = sorted(k_ for k_, v in checks.items() if not v and k_ not in POWER_NULL_KEYS)
                null_failing = sorted(k_ for k_, v in checks.items() if not v and k_ in POWER_NULL_KEYS)
            if null_failing:
                bad.append("null cell %s above its bar on the live structure: the GO rule is unsound here; new council round (no user_decision line clears it) (%s)" % (", ".join(null_failing), pp))
            if failing or thin:
                line, k = one(CRITIQUE_DECISION_LINE)
                if k != 1 or line != CRITIQUE_DECISION_LINE + "proceed":
                    bad.append("%s must contain exactly one line %r (%s: the council and the user decide before window spend)" % (
                        p, CRITIQUE_DECISION_LINE + "proceed", "; ".join(([("P3 cells below their bars: %s" % ", ".join(failing))] if failing else []) +
                                                                          ([("a frozen round has < %d clusters or < %d PRs: %s" % (FROZEN_MIN_CLUSTERS, FROZEN_MIN_PRS, line_pop_text(pops)))] if thin else []))))
            if (gj.get("gates", {}).get("P2") or {}).get("status") == "UNMEASURED":
                line, k = one(CRITIQUE_P2_LINE)
                if k != 1 or line != CRITIQUE_P2_LINE + "accepted":
                    bad.append("%s must contain exactly one line %r (the P2 planted-gap gate is UNMEASURED: the critique rules on it)" % (p, CRITIQUE_P2_LINE + "accepted"))
    return bad


# ---- ids and hashing
def case_id(r, cluster_id):
    return "%s:%s" % (r, cluster_id)


def hash_int(prefix, key):
    """sha256(prefix + key) as an integer (the spelled-out hash orders of the preregistration)."""
    return int(hashlib.sha256((prefix + key).encode("utf-8")).hexdigest(), 16)


def flip_selected(cluster_id):
    """Second-permutation selection (questions/merge.which_area@v1.json): sha256('x2d-flip' + cluster id) mod 5 == 0."""
    return hash_int("x2d-flip", cluster_id) % FLIP_FRACTION_MOD == 0


def cap_key(cluster_id):
    return hashlib.sha256(("x2d-cap\0" + cluster_id).encode()).hexdigest()


def reference_key(cluster_id):
    return hashlib.sha256(("x2d-reference\0" + cluster_id).encode()).hexdigest()


# ---- JSONL / JSON
def read_jsonl(path):
    """Rows of a JSONL file ([] when absent); a malformed line (the partial trailing line of a killed writer) is a loud error naming file and line."""
    out = []
    if os.path.exists(path):
        with open(path) as f:
            for n, line in enumerate(f, 1):
                if not line.strip():
                    continue
                try:
                    out.append(json.loads(line))
                except ValueError:
                    sys.exit("%s:%d: malformed JSONL line (a killed run left a partial line: truncate it)" % (path, n))
    return out


def append_jsonl(path, rows):
    os.makedirs(os.path.dirname(path) or ".", exist_ok=True)
    with open(path, "a") as f:
        for r in rows:
            f.write(json.dumps(r, sort_keys=True) + "\n")


def write_jsonl_new(path, rows):
    """Create a JSONL file that must not exist (append-only data: never overwritten)."""
    os.makedirs(os.path.dirname(path) or ".", exist_ok=True)
    with open(path, "x") as f:
        for r in rows:
            f.write(json.dumps(r, sort_keys=True) + "\n")


def write_json_new(path, obj):
    os.makedirs(os.path.dirname(path) or ".", exist_ok=True)
    with open(path, "x") as f:
        json.dump(obj, f, indent=1, sort_keys=True)
        f.write("\n")


def refuse_if_exists(*paths):
    """Append-only guard: exit when any path already exists and is non-empty."""
    clash = [p for p in paths if os.path.exists(p) and (not os.path.isfile(p) or os.path.getsize(p) > 0)]
    if clash:
        sys.exit("refusing: output exists (append-only): " + ", ".join(clash))


# ---- reuse of the frozen X2 modules by import (as X2c): common.scrub_env, common.clawpatch_map_env, common.build_safety_check, common.wilson, ...
if X2_DIR not in sys.path:
    sys.path.insert(0, X2_DIR)
if HERE not in sys.path:
    sys.path.insert(0, HERE)


# ---- Jev budget (PREREGISTRATION 12; thresholds.json caps): the driver's -max-cost-usd is cumulative over ITS OWN output file, so the value passed on the
# command line subtracts what the other score files of the same round (per-round cap) and of the whole study (total cap) have already billed.
def spent_in(path):
    """Billed USD recorded in one append-only scores file (answer and failure rows that report a cost)."""
    total = 0.0
    for row in read_jsonl(path):
        if row.get("type") in ("answer", "failure") and row.get("cost_known"):
            total += float(row.get("cost_usd") or 0.0)
    return total


def score_files(data=None):
    base = data or DATA
    out = set(glob.glob(os.path.join(base, "rounds", "*", "**", "scores*.jsonl"), recursive=True))
    out |= set(glob.glob(os.path.join(base, "onboard-jev", "scores*.jsonl")))
    return sorted(out)


def jev_budget(round_key, out_path, data=None):
    """-> the `-max-cost-usd` for a driver run writing `out_path` in round `round_key` (None for onboarding): the smaller of the per-round and the total
    allowance left after the OTHER score files, plus nothing for this file (the driver adds its own earlier spend itself). The v_7 re-score (files under
    rounds/<rr>/jev-v7/) is charged to the $20 total only, never to a round's cap, and does not count against the round's own runs (R3-17)."""
    out_path = os.path.abspath(out_path)
    data = main_data(data)  # rev 13: a re-gate root (D/rounds/0/regate) is charged to the study's own caps: its score files sit under D/rounds/0/, so the round-0 cap ($2) and the $20 total count them
    others = [p for p in score_files(data) if os.path.abspath(p) != out_path]
    total_other = sum(spent_in(p) for p in others)
    cap = CAPS["usd_total"] - total_other
    if round_key is not None and (os.sep + V7_RESCORE_SUB + os.sep) not in out_path:
        rdir = os.path.abspath(round_dir(round_key, data)) + os.sep
        round_other = sum(spent_in(p) for p in others if os.path.abspath(p).startswith(rdir) and (os.sep + V7_RESCORE_SUB + os.sep) not in os.path.abspath(p))
        cap = min(cap, CAPS["usd_per_round"] - round_other)
    return max(cap, 0.0)


def round_artifacts_exist(r, name, data=None):
    return os.path.exists(os.path.join(round_dir(r, data), name))
