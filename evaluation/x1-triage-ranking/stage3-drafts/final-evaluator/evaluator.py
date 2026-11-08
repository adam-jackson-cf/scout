"""Fail-closed synthetic evaluator draft. Never emits binding_gate_established."""
import itertools
import json
import math
import hashlib
from provenance import verify_bundle, maximum_completed_inputs


def identity(cell):
    return json.dumps(cell, sort_keys=True, separators=(",", ":"), allow_nan=False)


def expected_cells(manifest):
    dimensions = manifest["coverage_dimensions"]
    for values in itertools.product(*dimensions.values()):
        yield {"kind": "requested-coverage", **dict(zip(dimensions, values))}
    ref = manifest["calibrated_power_and_null"]
    names = ref["dimensions_from_coverage"] + ["size_dependence"]
    axes = [dimensions[n] if n != "size_dependence" else ref[n] for n in names]
    for values in itertools.product(*axes):
        for target in ref["realised_targets"]:
            yield {"kind": "calibrated-reference", **dict(zip(names, values)), "target": target,
                   "scanner_control_true_at_prevalence": False}
    missing = manifest["missing_label_grid"]
    names = ["cap_M", "prevalence", "realised_gain", "weighted_unknown_share", "mechanism"]
    for values in itertools.product(*(missing[n] for n in names)):
        yield {"kind": "missing", **dict(zip(names, values))}


def wilson(k, n):
    if type(k) is not int or type(n) is not int or not 0 <= k <= n or n <= 0:
        raise ValueError("invalid binomial count")
    z = 1.959963984540054
    p = k/n
    centre = (p+z*z/(2*n))/(1+z*z/n)
    half = z*math.sqrt(p*(1-p)/n+z*z/(4*n*n))/(1+z*z/n)
    return centre-half, centre+half


def verdict(*, integrity, violating, adequate, upper, unresolved, point, lower):
    if not all(math.isfinite(x) for x in (point, lower, upper)) or lower > upper:
        raise ValueError("unmeasured/malformed interval")
    if not integrity:
        return "inconclusive"
    if violating:
        return "not supported"
    if not adequate:
        return "inconclusive"
    if upper < .10:
        return "not supported"
    if unresolved:
        return "inconclusive"
    if point >= .10 and lower > 0:
        return "support"
    return "inconclusive"


def resource_fit(resources):
    required = ("construction", "calibration", "enumeration", "requested_base", "trials", "gate_checks")
    cpu = resources["cpu_seconds"]
    rss = resources["peak_rss_bytes"]
    if set(cpu) != set(required) or set(rss) != {"construction", "calibration", "trial_plus_helper"}:
        return False
    numbers = [*cpu.values(), *rss.values(), resources["wall_days"], resources["memory_bytes"]]
    if not all(type(x) in (int, float) and math.isfinite(x) and x >= 0 for x in numbers):
        return False
    return (resources["workers"] == 8 and 2*sum(cpu.values())/3600 <= 2880
            and resources["wall_days"] <= 15
            and 8*max(rss.values()) <= resources["memory_bytes"]-8*1024**3)


def summarize(rows):
    totals = {name: sum(r[name] for r in rows) for name in
              ("attempts", "executions", "eligible", "inadequate", "gate_failures",
               "lower_miss", "upper_miss", "support")}
    intervals = {}
    for numerator, denominator in (("gate_failures", "attempts"), ("inadequate", "executions"),
                                   ("lower_miss", "eligible"), ("upper_miss", "eligible"),
                                   ("support", "eligible")):
        k, n = totals[numerator], totals[denominator]
        intervals[numerator] = wilson(k, n) if 0 <= k <= n and n else None
    return {"counts": totals, "Wilson_MC95": intervals}


def evaluate(manifest, records, *, p_gate, resources, mhat, synthetic_only,
             provenance_payloads, provenance_digests, measurement_runs, binding_host):
    """Exact membership, counters, coverage, references and largest-row selection.

    Synthetic provenance is explicit; authentic sealed provenance cannot be
    inferred from data or a JSON approval flag. This draft cannot authorize OC.
    """
    if synthetic_only is not True or manifest["bootstrap_replicates"] != 1999:
        raise ValueError("draft accepts synthetic fixtures only, B=1999")
    if hashlib.sha256(identity(manifest).encode()).hexdigest() != "bfc7bb2d8b88712cecd3a21f0ab0c4d3bb4fe654ce1695baff0e3eafe4004ec7":
        raise ValueError("substituted signed manifest")
    provenance = verify_bundle(provenance_payloads, provenance_digests)
    maximum = maximum_completed_inputs(measurement_runs, binding_host)
    flat = {**{"cpu:"+k: v for k, v in resources["cpu_seconds"].items()},
            **{"rss:"+k: v for k, v in resources["peak_rss_bytes"].items()},
            "wall_days": resources["wall_days"]}
    if maximum != flat or resources["memory_bytes"] != binding_host["hw.memsize"]:
        raise ValueError("resources are not maxima on the binding host")
    if not math.isfinite(p_gate) or not 0 < p_gate <= 1:
        raise ValueError("unmeasured p_gate")
    expected = {identity(c): c for c in expected_cells(manifest)}
    found = {}
    for record in records:
        key = (identity(record["cell"]), record["population"])
        if key in found or key[0] not in expected or type(key[1]) is not int or key[1] not in range(5):
            raise ValueError("duplicate or unexpected cell/population")
        found[key] = record
    if len(found) != 5*len(expected):
        raise ValueError("missing cell/population")
    passed = {m: resource_fit(resources) for m in manifest["coverage_dimensions"]["cap_M"]}
    reports = {}
    for key, cell in expected.items():
        rows = [found[key, p] for p in range(5)]
        good = True
        for r in rows:
            counts = [r[n] for n in ("attempts", "executions", "eligible", "inadequate", "gate_failures", "lower_miss", "upper_miss", "support")]
            if any(type(n) is not int or n < 0 for n in counts):
                raise ValueError("noninteger/negative count")
            e, a, n = r["executions"], r["attempts"], r["eligible"]
            if (r["B"] != 1999 or n != 2000 or not n <= e <= 4000 or not e <= a <= math.ceil(4000/p_gate)
                or r["inadequate"] != e-n or r["gate_failures"] != a-e
                or any(r[name] > n for name in ("lower_miss", "upper_miss", "support"))):
                good = False
            if not good:
                continue
            # Tail gate is pooled; report each population separately below.
            if cell["kind"] == "calibrated-reference":
                realised = r["realised_gain"]
                if not math.isfinite(realised) or not r["reachable"]:
                    good = False
                elif cell["target"] == 0:
                    good &= realised <= 0
                else:
                    good &= realised >= .20 and wilson(r["inadequate"], e)[1] <= .05
            elif cell["kind"] == "missing":
                good &= r["mask_reachable"] and r["reachable"]
                realised = r["realised_gain"]
                target = cell["realised_gain"]
                good &= math.isfinite(realised) and (realised <= 0 if target == 0 else realised >= target)
        if good:
            good &= all(wilson(sum(r[name] for r in rows), 10000)[1] <= .035 for name in ("lower_miss", "upper_miss"))
            if cell["kind"] == "calibrated-reference":
                interval = wilson(sum(r["support"] for r in rows), 10000)
                good &= interval[1] <= .05 if cell["target"] == 0 else interval[0] >= .80
        passed[cell["cap_M"]] &= bool(good)
        reports[key] = {"passed": bool(good), "pooled": summarize(rows),
                        "populations": [summarize([r]) for r in rows]}
    if not math.isfinite(mhat) or mhat < 0:
        raise ValueError("unmeasured Mhat")
    candidates = [m for m in passed if m <= mhat]
    selected = max(candidates) if candidates else None
    return {"status": "draft_synthetic_gate_passed" if selected is not None and passed[selected] else "not testable",
            "rows": passed, "selected_M": selected, "any_row_passed_before_dev": any(passed.values()),
            "cell_reports": reports, "provenance_sha256": provenance,
            "authority": "none; synthetic draft, scoped review required"}
