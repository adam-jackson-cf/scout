"""X2d power simulations of the resolved-cluster GO rule (PREREGISTRATION 11, thresholds.json study.power; R4-13 (6), R5-01) and the planting helpers validate.py reuses.

    .tools/venv/bin/python evaluation/x2d-merge-association-pilot/power_sim.py [--pilots 1000] [--seed 20261029] [--boot-B 1000] [--out D/power.json] [--structure-k23 D/rounds/0/k23.json]

Model (rev 6; the rev-1..4 "positives outside the population, P_new = R_new = 1" and "no abstention" models are retired -- superseded, kept in DRAFT-NOTES). Every simulated pilot has the
frozen rounds' clusters (91 clusters in 40 PRs from prepass/rounds.json pre-sign-off, the live k23.json frozen_pr_counts at round 0), of which n_pos are gold-`none` positives INSIDE the population
(10 pre-sign-off, real_none_proxy floored at 0 and capped at 20 at round 0), spread evenly over the PR-sorted cluster list. Stub Jev is correct with probability p_J per cluster, B2 with p_B2
(a Gaussian copula: shared latent normal s, arm latent sqrt(corr) * s + sqrt(1 - corr) * e_arm, correct iff latent < inv_cdf(p_arm); corr 0.5, 0 for the nulls; pr_w > 0 adds a PR-level latent
sqrt(pr_w) * u(PR, arm) + sqrt(1 - pr_w) * copula latent). Covered cluster: a wrong Jev answer is a false `none` with probability fn_share (0.5; resolved, not assigned) else a wrong assignment
(assigned); positive cluster: `none` (true none) when Jev is correct, else a wrong assignment, which COUNTS in `assigned` (the analysis's P_assign counts assigned_gold_none). B2 is drawn on every
cluster and never abstains. ABSTENTION (`u`, `mode`): "random" = each cluster is abstained (unresolved) independently of correctness with probability u (the honest cells); "b2_errors" = Jev
abstains on exactly the clusters B2 gets wrong, at most the first int(0.15 * n) of them in PR order (N4a); "own_errors" = Jev abstains on exactly the clusters it would get wrong (N4b, N4b-rho0, N4c);
"own_errors_cap" = Jev abstains on its first int(0.15 * n) own errors in PR order and answers (wrongly) the rest (N4b-cap). The raw (unthresholded) guess of an abstained cluster is correct iff J,
whether the cluster is abstained or not (R5-01), so an own-error abstention is scored by a wrong guess.
Every figure the rule reads is built as analysis_x2d.summarise builds it (P_assign = assigned-correct / assigned; R_assign_res = assigned-correct / (gold-covered AND resolved); gain_res over the
resolved clusters with the same B2 outcomes, its 90 % PR-cluster-bootstrap lower bound over the PRs with a resolved cluster; gain_all with every abstention a miss; gain_guess = (sum of J - sum of B2) / n over
every cluster, its 95 % PR-cluster-bootstrap lower bound (the guard h) over the PRs with a cluster; u; n_res) and the verdict is analysis_x2d.go_verdict itself. Perception, N1 and informativeness are
TAKEN AS PASSING. The bootstrap bounds use B = boot_B (1,000; the real verdict 10,000) with the SAME resampling scheme as analysis_x2d.pr_cluster_bootstrap (PRs sorted by key, one random.Random(seed),
randrange(m)); the index draws are made once per m and reused across pilots. Exact short-circuit: the bounds (conditions e and h) are computed only when every other condition holds with both bounds
set to pass (a failed point condition gives NO-GO, a failed non-point condition INCONCLUSIVE whatever the bounds).
Deterministic given --seed. Append-only output; calls x2d.verify_pins() (X2D_SKIP_SIGNOFF=1 -> verify_pins(require_signoff=False), used by validation). A binding cell on the wrong side of its bar is
printed, recorded as checks[<cell>] = False and all_pass = False in power.json, never moved. Exit code (R5-02): 3 when any NULL check (N3, N3_PR, N4a, N4b, N4b_cap, N4b_rho0) is false (the GO rule is
unsound on this structure: a new council round, never a spend decision); 0 when only P3 cells miss (power.json carries them; the round-0 critique then requires `user_decision: proceed`).

Output schema x2d-power-v5: cells P3_u0 / P3_u05 / P3_u10 (binding), P3_PR / P3_rho0 / P3_curve_delta_0.15 (report-only, at u = 0.05), N3 and N3_PR (u in {0, 0.10}, binding), N4a / N4b / N4b_cap /
N4b_rho0 (binding), N4c (report-only); `checks` {P3_u0, P3_u05, P3_u10, N3, N3_PR, N4a, N4b, N4b_cap, N4b_rho0}, `all_pass`, `p3_real_structure` {share: the P3-u0 share, ...}, `p3_abstention`
{"0.05", "0.10"}, `n4_max` (over N4a, N4b, N4b_cap, N4b_rho0).

Helpers: empirical_structure(), simulate(), plant() (synthetic units/gold/arm rows of the end-to-end plant), permute_gold() (N2 within-round label permutation), n2_permutation().
"""
import argparse
import json
import math
import os
import random
import sys
import time
from operator import itemgetter
from statistics import NormalDist

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import x2d  # noqa: E402
import analysis_x2d as A  # noqa: E402

HERE = os.path.dirname(os.path.abspath(__file__))
P3 = {"p_b2": 0.80, "p_j": 0.95}
BARS = {"P3_u0": 0.80, "P3_u05": 0.75, "P3_u10": 0.60, "N3": 0.10, "N3_PR": 0.10, "N4a": 0.10, "N4b": 0.10, "N4b_cap": 0.10, "N4b_rho0": 0.10}  # P3_*: minimum GO share; N*: maximum (equal to thresholds.json study.power.bars, checked by validation)
N3 = {"p_b2": [0.5, 0.6, 0.7, 0.8, 0.85, 0.9, 0.95], "u": [0.0, 0.10]}
N4A_P = [0.80, 0.85, 0.90, 0.95]
N4B_P = [0.85, 0.88, 0.90, 0.95]
N4B_CAP_P = [0.80, 0.85]  # N4b-cap: own-error abstention truncated at the cap, rho 0.5
N4B_RHO0_P = [0.85, 0.88, 0.90]  # N4b-rho0: own-error abstention, rho 0 (the correlation that maximises the false GO)
NULL_CHECKS = ("N3", "N3_PR", "N4a", "N4b", "N4b_cap", "N4b_rho0")  # a false null check is a rule-4 failure (exit 3, a round-0 stop); a false P3 check is a user decision
N4C = {"p_j": 0.90, "p_b2": 0.80, "u": 0.10}
CORR = 0.5
PR_W = 0.3  # N3-PR / P3-PR: share of latent variance that is PR-level (latent = sqrt(0.3)*u(PR, arm) + sqrt(0.7)*e)
P3_CURVE = [0.70, 0.75, 0.80, 0.85]  # report-only: delta = 0.15 curve over p_B2 (at u = 0.05)
U_P3_REPORT = 0.05
N_POS = 10  # gold-none positives inside the population pre-sign-off; at round 0 real_none_proxy, floor 0, cap N_POS_CAP
N_POS_CAP = 20
FN_SHARE = 0.5
ABST_CAP = 0.15  # N4a truncation: at most int(0.15 * n) clusters
P3_NAME = "P3 (conditional on perception, N1 and informativeness; positives inside the population)"


def empirical_structure(rounds_json=None, rounds=None):
    """[{"round","pr","n"}] scored clusters per PR of the frozen rounds (default 8-10), from prepass/rounds.json."""
    with open(rounds_json or os.path.join(HERE, "prepass", "rounds.json")) as f:
        j = json.load(f)
    out = []
    for rd in j["rounds"]:
        if rd["round"] in (rounds or x2d.FROZEN_ROUNDS):
            for pr in rd["prs"]:
                if pr["scored"] > 0:
                    out.append({"round": rd["round"], "pr": str(pr["pr"] if pr.get("pr") is not None else pr["sha"][:8]), "n": int(pr["scored"])})
    return out


def _group(structure):
    """PRs sorted by key (as pr_cluster_bootstrap does) -> (keys, sizes)."""
    prs = sorted(({"key": "%s:%s" % (s["round"], s["pr"]), "n": s["n"]} for s in structure), key=lambda x: x["key"])
    return [p["key"] for p in prs], [p["n"] for p in prs]


def _bootstrap_getters(m, boot_B, seed):
    """Index draws of boot_B replicates (random.Random(seed), randrange(m) m times) as itemgetters."""
    rng = random.Random(seed)
    getters = []
    for _ in range(boot_B):
        idx = [rng.randrange(m) for _ in range(m)]
        getters.append(itemgetter(*idx) if m > 1 else (lambda t, i=idx[0]: (t[i],)))
    return getters


def gain_res_lower_bound(d_by_pr, n_by_pr, getters, level=0.90):
    """One-sided bootstrap lower bound of the resolved-cluster paired gain: the same value pr_cluster_bootstrap(rows with n_res > 0, gain_res_stat)["lower"][level] gives with the same seed
    (replicates with no resolved cluster are dropped, as there)."""
    d, n = tuple(d_by_pr), tuple(n_by_pr)
    vals = []
    for g in getters:
        den = sum(g(n))
        if den:
            vals.append(sum(g(d)) / den)
    vals.sort()
    return A._quantile(vals, 1 - level) if vals else None


def _thresholds():
    with open(os.path.join(HERE, "thresholds.json")) as f:
        return json.load(f)


def simulate(p_j, p_b2, structure, pilots=1000, boot_B=1000, seed=x2d.SEED, corr=CORR, thresholds=None, pr_w=0.0, u=0.0, mode="random", n_pos=N_POS, fn_share=FN_SHARE, trace=None):
    """GO share of `pilots` simulated pilots under the model of the module docstring. `trace` (a list): each pilot's latent (J, B, per-cluster outcome kinds, PR of each cluster) is appended (validation only). -> {"p_j","p_b2","corr","pr_w","u","mode","n_pos","fn_share","pilots","boot_B","go","share","wilson95",
    "verdicts": {GO/INCONCLUSIVE/NO-GO counts},"bootstrap_run","u_mean","structure": {clusters, covered, prs}}."""
    thr = thresholds or _thresholds()
    keys, sizes = _group(structure)
    n_all = sum(sizes)
    if n_pos >= n_all:
        raise ValueError("n_pos %d >= clusters %d" % (n_pos, n_all))
    pos_idx = {int((i + 0.5) * n_all / n_pos) for i in range(n_pos)} if n_pos else set()
    pr_of, is_pos = [], []
    for pi, n in enumerate(sizes):
        for c in range(n):
            pr_of.append(pi)
            is_pos.append(len(pr_of) - 1 in pos_idx)
    n_cov = n_all - n_pos
    getters = {}
    zj, zb = NormalDist().inv_cdf(min(max(p_j, 1e-12), 1 - 1e-12)), NormalDist().inv_cdf(min(max(p_b2, 1e-12), 1 - 1e-12))
    a, b = math.sqrt(corr), math.sqrt(1 - corr)
    wu, we = math.sqrt(pr_w), math.sqrt(1 - pr_w)
    rng = random.Random("x2d-power-v5|%s|%s|%s|%s|%s|%s|%s|%s" % (seed, p_j, p_b2, corr, pr_w, u, mode, n_pos))
    cap_n = int(ABST_CAP * n_all)
    counts = {"GO": 0, "INCONCLUSIVE": 0, "NO-GO": 0}
    boot_runs = 0
    u_sum = 0.0
    m_pr = len(sizes)
    for _ in range(pilots):
        J, Bb = [], []
        for n in sizes:
            uj, ub = (rng.gauss(0, 1), rng.gauss(0, 1)) if pr_w else (0.0, 0.0)
            for _c in range(n):
                s = rng.gauss(0, 1)
                J.append((wu * uj + we * (a * s + b * rng.gauss(0, 1))) < zj)
                Bb.append((wu * ub + we * (a * s + b * rng.gauss(0, 1))) < zb)
        if mode == "b2_errors":
            wrong = [i for i in range(n_all) if not Bb[i]]
            ab = set(wrong[:cap_n])
        elif mode == "own_errors":
            ab = {i for i in range(n_all) if not J[i]}
        elif mode == "own_errors_cap":  # N4b-cap: the gamer that stops at the cap: abstain on its first int(0.15 n) own errors in PR order, answer (wrongly) the rest
            ab = set([i for i in range(n_all) if not J[i]][:cap_n])
        elif u > 0:
            ab = {i for i in range(n_all) if rng.random() < u}
        else:
            ab = set()
        correct = assigned = abst_cov = 0
        d_res, n_res_pr, spread_prs = [0] * m_pr, [0] * m_pr, set()
        d_guess = [0] * m_pr  # R5-01: every cluster, an abstained one scored by Jev's own unthresholded guess (correct iff J, abstained or not)
        d_all_sum = 0
        kinds = ["abstain"] * n_all  # per-cluster outcome of this pilot, recorded into `trace` (validation: rebuild the pilot's cases and arm rows and compare analysis_x2d.summarise)
        for i in range(n_all):
            j, bb, pi = J[i], Bb[i], pr_of[i]
            d_guess[pi] += int(j) - int(bb)
            if i in ab:
                abst_cov += 0 if is_pos[i] else 1
                d_all_sum -= int(bb)  # an abstention is a Jev miss
                continue
            d_all_sum += int(j) - int(bb)
            n_res_pr[pi] += 1
            d_res[pi] += int(j) - int(bb)
            if j and not bb:
                spread_prs.add(pi)
            if is_pos[i]:
                kinds[i] = "pos_right" if j else "pos_wrong"
                if not j:
                    assigned += 1  # a wrong assignment on a gold-none positive counts in P_assign
            elif j:
                kinds[i] = "right"
                assigned += 1
                correct += 1
            elif rng.random() >= fn_share:
                kinds[i] = "wrong_assigned"
                assigned += 1  # a wrong assignment (the other wrong answers are false `none`: resolved, not assigned)
            else:
                kinds[i] = "false_none"
        n_res = sum(n_res_pr)
        gc_res = n_cov - abst_cov
        pa = (correct / assigned) if assigned else None
        ra = (correct / gc_res) if gc_res else None
        gain_res = (sum(d_res) / n_res) if n_res else None
        gain_all = d_all_sum / n_all
        u_rate = len(ab) / n_all
        u_sum += u_rate
        pa_lb = A.wilson_bounds_one_sided(correct, assigned, A.GO["level"])[0] if assigned else None
        gr = {"point": gain_res, "lb90": None, "n_jev_only_prs": len(spread_prs)}
        gg = {"point": sum(d_guess) / n_all, "n": n_all, "lb95": None}
        pooled = {"p_assign": {"k": correct, "n": assigned, "point": pa, "wilson_lb90": pa_lb}, "r_assign_res": {"k": correct, "n": gc_res, "point": ra},
                  "gain_res": gr, "gain_all": {"point": gain_all, "n": n_all, "lb90": None}, "gain_guess": gg, "n_res": n_res,
                  "unresolved": {"k": len(ab), "n": n_all, "rate": u_rate}, "perception": {"pass": True}, "n1": {"pass": True}, "informative_all": True}
        if A.go_verdict(dict(pooled, gain_res=dict(gr, lb90=1.0), gain_guess=dict(gg, lb95=1.0)), thr)["verdict"] == "GO":  # exact short-circuit: the bounds matter only when every other condition holds
            idx = [pi for pi in range(m_pr) if n_res_pr[pi]]
            if idx:
                gt = getters.get(len(idx))
                if gt is None:
                    gt = getters[len(idx)] = _bootstrap_getters(len(idx), boot_B, seed)
                gr["lb90"] = gain_res_lower_bound([d_res[pi] for pi in idx], [n_res_pr[pi] for pi in idx], gt, A.GO["level"])
                gga = getters.get(m_pr)  # every PR has >= 1 cluster: the rows with n_all > 0 are all m_pr PRs
                if gga is None:
                    gga = getters[m_pr] = _bootstrap_getters(m_pr, boot_B, seed)
                gg["lb95"] = gain_res_lower_bound(d_guess, sizes, gga, A.GO["guard_level"])
                boot_runs += 1
        counts[A.go_verdict(pooled, thr)["verdict"]] += 1
        if trace is not None:
            trace.append({"J": list(J), "B": list(Bb), "kinds": kinds, "pr_of": list(pr_of), "is_pos": list(is_pos), "sizes": list(sizes)})
    go = counts["GO"]
    lo, hi = A.wilson(go, pilots)
    return {"p_j": p_j, "p_b2": p_b2, "corr": corr, "pr_w": pr_w, "u": u, "mode": mode, "n_pos": n_pos, "fn_share": fn_share, "pilots": pilots, "boot_B": boot_B, "go": go, "share": go / pilots,
            "wilson95": [lo, hi], "verdicts": counts, "bootstrap_run": boot_runs, "u_mean": u_sum / pilots, "structure": {"clusters": n_all, "covered": n_cov, "prs": len(sizes)}}


def structure_from_k23(k23_path):
    """-> ([{"round","pr","n"}], real_none_proxy) from rounds/0/k23.json: `frozen_pr_counts` ({"8": {pr_key: n_scored_clusters}, ...}) and `real_none_proxy`
    (top level, or under "gates"). Refuses a missing key, empty structure or a non-positive count."""
    with open(k23_path) as f:
        j = json.load(f)
    j = {**j.get("gates", {}), **j} if isinstance(j.get("gates"), dict) else j
    for k in ("frozen_pr_counts", "real_none_proxy"):
        if k not in j:
            sys.exit("error: %s lacks %r (rebuild with the R2-11 build_units)" % (k23_path, k))
    out = []
    for rd in sorted(j["frozen_pr_counts"], key=int):
        for pr, n in sorted(j["frozen_pr_counts"][rd].items()):
            if int(n) < 0:
                sys.exit("error: negative cluster count for %s:%s in %s" % (rd, pr, k23_path))
            if int(n) > 0:
                out.append({"round": int(rd), "pr": str(pr), "n": int(n)})
    if not out:
        sys.exit("error: %s has no scored clusters" % k23_path)
    return out, j["real_none_proxy"]





# ------------------------------------------------------------------ planting helpers (validate.py)
def plant(structure, p_j, p_b2, seed=x2d.SEED, corr=CORR, n_areas=40):
    """Synthetic end-to-end plant: units (set window), gold (covered, one acceptable area), arm rows for V-sem and B2 correct with probability p_j / p_b2 (Gaussian copula as
    simulate; wrong answers are a uniformly drawn other area). -> {"units", "gold": {case_id: row}, "arms": {"V-sem": rows, "B2": rows}, "areas": [ids]}."""
    rng = random.Random("x2d-plant|%s|%s|%s|%s" % (seed, p_j, p_b2, corr))
    areas = ["area%03d" % i for i in range(n_areas)]
    zj, zb = NormalDist().inv_cdf(min(max(p_j, 1e-12), 1 - 1e-12)), NormalDist().inv_cdf(min(max(p_b2, 1e-12), 1 - 1e-12))
    a, b = math.sqrt(corr), math.sqrt(1 - corr)
    units, gold, arms = [], {}, {"V-sem": [], "B2": []}
    i = 0
    for s in structure:
        for _c in range(s["n"]):
            cid = "%s:plant-%04d" % (s["round"], i)
            i += 1
            true = rng.choice(areas)
            units.append({"case_id": cid, "round": s["round"], "cluster_id": "plant-%04d" % i, "pr": int(s["pr"]) if str(s["pr"]).isdigit() else None, "pr_key": str(s["pr"]),
                          "members": ["f%d.py" % i], "role": "test", "stratum": 1, "set": "window", "control": False})
            gold[cid] = {"case_id": cid, "covered": True, "acceptable": [true], "primary": true, "none_kind": None, "agreement": "agreed", "control_kind": None}
            sh = rng.gauss(0, 1)
            for arm, z in (("V-sem", zj), ("B2", zb)):
                ok = (a * sh + b * rng.gauss(0, 1)) < z
                area = true if ok else rng.choice([x for x in areas if x != true])
                arms[arm].append({"case_id": cid, "arm": arm, "output": "assigned", "area_id": area, "confidence": 0.9, "unresolved_reason": None})
    return {"units": units, "gold": gold, "arms": arms, "areas": areas}


def permute_gold(units, gold, seed=x2d.SEED):
    """N2: within-round permutation of the gold labels among the round's units (labels move, case ids stay). -> new {case_id: gold}."""
    rng = random.Random("x2d-n2|%s" % seed)
    out = {}
    by_round = {}
    for u in units:
        by_round.setdefault(u["round"], []).append(u["case_id"])
    for r, ids in sorted(by_round.items(), key=lambda kv: str(kv[0])):
        shuffled = ids[:]
        rng.shuffle(shuffled)
        for i in range(len(ids)):  # derangement: no unit keeps its own label (a fixed point would put the true label back and lift P_assign above chance)
            if len(ids) > 1 and shuffled[i] == ids[i]:
                j = (i + 1) % len(ids)
                shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
        for cid, src in zip(ids, shuffled):
            g = dict(gold[src])
            g["case_id"] = cid
            out[cid] = g
    return out


def n2_permutation(structure, p_j, p_b2, seed=x2d.SEED, reps=200, n_areas=40):
    """N2 check on a planted dataset: over `reps` permutations of the gold labels, the mean paired gain (should be ~0) and the mean P_assign (should be ~1/n_areas)."""
    pl = plant(structure, p_j, p_b2, seed=seed, n_areas=n_areas)
    gains, pas = [], []
    for k in range(reps):
        g = permute_gold(pl["units"], pl["gold"], seed=seed * 1000 + k)
        cases = A.build_cases(pl["units"], g, pl["arms"])
        gains.append(A.paired(cases, "V-sem", "B2")["gain"])
        pas.append(A.rates(A.counts_for([c["out"]["V-sem"] for c in cases]))["P_assign"]["value"])
    mean = lambda v: sum(v) / len(v)
    sd = (sum((x - mean(gains)) ** 2 for x in gains) / (len(gains) - 1)) ** 0.5 if len(gains) > 1 else 0.0
    return {"reps": reps, "mean_gain": mean(gains), "gain_sd": sd, "gain_se_of_mean": sd / math.sqrt(len(gains)), "mean_P_assign": mean(pas), "chance": 1.0 / n_areas}


# ------------------------------------------------------------------ CLI
def _cell(p_j, p_b2, structure, pilots, seed, boot_B, corr, pr_w=0.0, u=0.0, mode="random", n_pos=N_POS):
    return simulate(p_j, p_b2, structure, pilots, boot_B, seed, corr, pr_w=pr_w, u=u, mode=mode, n_pos=n_pos)


def run(pilots=1000, seed=x2d.SEED, boot_B=1000, corr=CORR, structure=None, k23=None):
    """-> dict (schema x2d-power-v5). k23 = path of rounds/0/k23.json: the structure is then the live frozen-round PR counts (`frozen_pr_counts`) and n_pos = real_none_proxy (floor 0, cap
    N_POS_CAP); otherwise the empirical 91 / 40 structure and N_POS. Binding: P3_u0 / P3_u05 / P3_u10 (p_J 0.95, p_B2 0.80, corr, u 0 / 0.05 / 0.10), N3 and N3_PR (every grid p, u in {0, 0.10}),
    N4a, N4b, N4b_cap, N4b_rho0; report-only: P3_PR, P3_rho0, the delta = 0.15 curve (all at u = 0.05), N4c."""
    proxy = None
    if k23:
        structure, proxy = structure_from_k23(k23)
    structure = structure or empirical_structure()
    n_pos = N_POS if proxy is None else min(max(int(proxy), 0), N_POS_CAP)
    t0 = time.time()
    cell = lambda p_j, p_b2, c, pr_w=0.0, u=0.0, mode="random": _cell(p_j, p_b2, structure, pilots, seed, boot_B, c, pr_w, u, mode, n_pos)
    out_cells = {}
    for name, u in (("P3_u0", 0.0), ("P3_u05", 0.05), ("P3_u10", 0.10)):
        c = cell(P3["p_j"], P3["p_b2"], corr, u=u)
        c["bar"] = {"min_go_share": BARS[name]}
        c["pass"] = c["share"] >= BARS[name]
        out_cells[name] = c
    n3, n3pr = [], []
    for p in N3["p_b2"]:
        for uu in N3["u"]:
            for lst, pw in ((n3, 0.0), (n3pr, PR_W)):
                c = cell(p, p, 0.0, pw, uu)
                c["bar"] = {"max_go_share": BARS["N3"]}
                c["pass"] = c["share"] <= BARS["N3"]
                lst.append(c)
    n4a, n4b = [], []
    for p in N4A_P:
        c = cell(p, p, corr, mode="b2_errors")
        c["bar"] = {"max_go_share": BARS["N4a"]}
        c["pass"] = c["share"] <= BARS["N4a"]
        n4a.append(c)
    for p in N4B_P:
        c = cell(p, p, corr, mode="own_errors")
        c["bar"] = {"max_go_share": BARS["N4b"]}
        c["pass"] = c["share"] <= BARS["N4b"]
        n4b.append(c)
    n4b_cap, n4b_rho0 = [], []
    for p in N4B_CAP_P:
        c = cell(p, p, corr, mode="own_errors_cap")
        c["bar"] = {"max_go_share": BARS["N4b_cap"]}
        c["pass"] = c["share"] <= BARS["N4b_cap"]
        n4b_cap.append(c)
    for p in N4B_RHO0_P:
        c = cell(p, p, 0.0, mode="own_errors")
        c["bar"] = {"max_go_share": BARS["N4b_rho0"]}
        c["pass"] = c["share"] <= BARS["N4b_rho0"]
        n4b_rho0.append(c)
    n4c = cell(N4C["p_j"], N4C["p_b2"], corr, mode="own_errors")
    p3_pr = cell(P3["p_j"], P3["p_b2"], corr, PR_W, U_P3_REPORT)  # report-only
    p3_rho0 = cell(P3["p_j"], P3["p_b2"], 0.0, 0.0, U_P3_REPORT)  # report-only
    curve = [cell(round(pb + 0.15, 2), pb, corr, 0.0, U_P3_REPORT) for pb in P3_CURVE]  # report-only, delta = 0.15
    checks = {"P3_u0": out_cells["P3_u0"]["pass"], "P3_u05": out_cells["P3_u05"]["pass"], "P3_u10": out_cells["P3_u10"]["pass"], "N3": all(c["pass"] for c in n3),
              "N3_PR": all(c["pass"] for c in n3pr), "N4a": all(c["pass"] for c in n4a), "N4b": all(c["pass"] for c in n4b), "N4b_cap": all(c["pass"] for c in n4b_cap), "N4b_rho0": all(c["pass"] for c in n4b_rho0)}
    out = {"schema": "x2d-power-v5", "seed": seed, "pilots": pilots, "boot_B": boot_B, "corr_P3": corr, "corr_N3": 0.0, "pr_weight": PR_W, "n_pos": n_pos, "fn_share": FN_SHARE, "bars": BARS,
           "structure": {"rounds": x2d.FROZEN_ROUNDS, "clusters": sum(s["n"] for s in structure), "prs": len(structure), "clusters_per_pr": [s["n"] for s in structure]},
           "structure_source": "k23" if k23 else "empirical", **out_cells, "N3": n3, "N3_PR": n3pr, "N4a": n4a, "N4b": n4b, "N4b_cap": n4b_cap, "N4b_rho0": n4b_rho0, "N4c": n4c, "P3_PR": p3_pr, "P3_rho0": p3_rho0, "P3_curve_delta_0.15": curve,
           "checks": checks, "all_pass": all(checks.values()), "seconds": round(time.time() - t0, 1), "p3_name": P3_NAME,
           "p3_real_structure": {"share": out_cells["P3_u0"]["share"], "wilson95": out_cells["P3_u0"]["wilson95"], "pass": out_cells["P3_u0"]["pass"], "structure": out_cells["P3_u0"]["structure"], "n_pos": n_pos,
                                 "note": "the P3-u0 cell on the %s structure" % ("live frozen-round" if k23 else "empirical 91 / 40")},
           "p3_abstention": {"0.05": out_cells["P3_u05"]["share"], "0.10": out_cells["P3_u10"]["share"]},
           "n4_max": max(c["share"] for c in n4a + n4b + n4b_cap + n4b_rho0),
           "assumed_passing_in_simulation": ["perception check", "N1", "every round informative"]}
    if k23:
        out["real_none_proxy"] = proxy
    return out


def main(argv=None):
    ap = argparse.ArgumentParser(description=__doc__.split("\n")[0])
    ap.add_argument("--pilots", type=int, default=1000)
    ap.add_argument("--seed", type=int, default=x2d.SEED)
    ap.add_argument("--boot-B", type=int, default=1000)
    ap.add_argument("--corr", type=float, default=CORR)
    ap.add_argument("--out", default=x2d.data_path("power.json"))
    ap.add_argument("--structure-k23", help="rounds/0/k23.json: use the live frozen-round PR counts and real_none_proxy (n_pos) instead of the empirical 91 / 40 structure")
    a = ap.parse_args(argv)
    bad = x2d.verify_pins(require_signoff=x2d.require_signoff())
    if bad:
        sys.exit("pin verification failed:\n  " + "\n  ".join(bad))
    x2d.refuse_if_exists(a.out)
    res = run(a.pilots, a.seed, a.boot_B, a.corr, k23=a.structure_k23)
    x2d.write_json_new(a.out, res)
    for name in ("P3_u0", "P3_u05", "P3_u10"):
        c = res[name]
        print("%s u=%.2f GO share %.3f [%.3f, %.3f] (need >= %.2f) %s" % (name, c["u"], c["share"], c["wilson95"][0], c["wilson95"][1], BARS[name], "PASS" if c["pass"] else "FAIL"))
    for key in ("N3", "N3_PR", "N4a", "N4b", "N4b_cap", "N4b_rho0"):
        for c in res[key]:
            print("%s p=%.2f u=%.2f GO share %.3f [%.3f, %.3f] (need <= %.2f) %s" % (key, c["p_j"], c["u"], c["share"], c["wilson95"][0], c["wilson95"][1], BARS[key], "PASS" if c["pass"] else "FAIL"))
    for key in ("N4c", "P3_PR", "P3_rho0"):
        print("%s (report-only) GO share %.3f [%.3f, %.3f] verdicts %s" % (key, res[key]["share"], res[key]["wilson95"][0], res[key]["wilson95"][1], json.dumps(res[key]["verdicts"])))
    for c in res["P3_curve_delta_0.15"]:
        print("P3 curve (report-only) p_b2=%.2f p_j=%.2f GO share %.3f [%.3f, %.3f]" % (c["p_b2"], c["p_j"], c["share"], c["wilson95"][0], c["wilson95"][1]))
    print("n4_max %.3f; checks %s; seconds %.1f" % (res["n4_max"], json.dumps(res["checks"]), res["seconds"]))
    missed = [k for k, ok in res["checks"].items() if not ok]
    null_missed = [k for k in missed if k in NULL_CHECKS]
    if null_missed:  # R5-02: a planted null above its bar means the GO rule is unsound on this structure: never a spend decision
        print("NULL CELL ABOVE ITS BAR (the GO rule is unsound here; new council round, exit 3): %s" % ", ".join(null_missed))
        return 3
    if missed:  # a P3 cell below its bar is reported, never moved; power.json carries it and the round-0 critique requires `user_decision: proceed` (R5-02 (3))
        print("BAR MISS (P3 cell; the round-0 critique requires `user_decision: proceed`): %s" % ", ".join(missed))
    return 0


if __name__ == "__main__":
    sys.exit(main())
