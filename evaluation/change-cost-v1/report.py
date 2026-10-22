#!/usr/bin/env python3
"""E1b: print markdown tables from model-results.json / dataset-counts.json (pasted into RESULTS.md)."""
import json
import sys
from pathlib import Path
sys.stdout = open(Path(__file__).resolve().parent / "tables.md", "w")

HERE = Path(__file__).resolve().parent
R = json.loads((HERE / "model-results.json").read_text())
C = json.loads((HERE / "dataset-counts.json").read_text())
REPOS = ["fastapi", "webpack", "nest", "caddy", "axum", "junit5", "eShopOnWeb"]


def pct(x):
    return f"{100 * x:+.2f}%"


def row(c):
    return (f"{pct(c['rel_reduction'])} (MSE {c['mse_base']:.4f} -> {c['mse_model']:.4f}) | "
            f"[{pct(c['ci95'][0])}, {pct(c['ci95'][1])}] | {c['p_one_sided']:.3f} | "
            f"{c['repos_improved_big']}/{c['n_big']} (need {c['repos_needed']}) | {'**yes**' if c['rule_met'] else 'no'}")


print("### dataset counts")
print("| repo | raw areas T0 | eligible areas (>=3 files) | exist at T-1 | >=5 outcome changes | included | included-area changes (sum) | PRs fetched | PRs linked | areas O4 unmeasured | files (file-level) |")
print("|---|---|---|---|---|---|---|---|---|---|---|")
for c in C:
    if "skipped" in c:
        print(f"| {c['repo']} | skipped: {c['skipped']} | | | | | | | | | |")
    else:
        print(f"| {c['repo']} | {c['raw_areas_t0']} | {c['areas_eligible']} | {c['areas_exist_t1']} | {c['areas_ge5_changes']} | "
              f"{c['areas_included']} | {c['changes_in_included']} | {c['prs_fetched']} | {c['prs_linked_outcome']} | "
              f"{c['areas_o4_unmeasured']} | {c['files_included']} |")
for level in ("area", "file"):
    print(f"\n### {level}-level comparisons")
    print("| outcome | n | comparison | pooled relative MSE reduction | 95% CI | p (1-sided) | repos improved (>=10 units) | rule |")
    print("|---|---|---|---|---|---|---|---|")
    for oc, r in R[level].items():
        if "comparisons" not in r:
            print(f"| {oc} | {r['n']} | skipped | | | | | |"); continue
        for k, c in r["comparisons"].items():
            print(f"| {oc} | {r['n']} | {c['name']} | {row(c)} |")
    print(f"\n### {level}-level ablations (baseline + one predictor)")
    print("| outcome | predictor | reduction | 95% CI | p | repos | rule |")
    print("|---|---|---|---|---|---|---|")
    for oc, r in R[level].items():
        for q, c in r.get("ablations", {}).items():
            print(f"| {oc} | {q} | {row(c)} |")
print("\n### per-repo relative MSE reduction (area level)")
print("| outcome | comparison | " + " | ".join(REPOS) + " |")
print("|---|---|" + "---|" * len(REPOS))
for oc, r in R["area"].items():
    for k, c in r.get("comparisons", {}).items():
        print(f"| {oc} (n by repo {r['counts']}) | {k} | " + " | ".join(
            pct(c["per_repo"][g]) if g in c["per_repo"] else "-" for g in REPOS) + " |")
print("\n### bootstrap coefficients (area level, z-scored ridge)")
print("| outcome | model | predictor | coef | 95% CI | construct-invalid |")
print("|---|---|---|---|---|---|")
for oc, r in R["area"].items():
    for key in ("coefs_level_model", "coefs_full_model", "coefs_coupling_model"):
        for t, v in r.get(key, {}).items():
            print(f"| {oc} | {key} | {t} | {v['coef']:+.3f} | [{v['ci95'][0]:+.3f}, {v['ci95'][1]:+.3f}] | {v['construct_invalid']} |")
print("\n### decision JSON")
print(json.dumps(R["decision"], indent=1))
print("\n### file secondary rule_met")
print(json.dumps(R["file_secondary_rule_met"]))
