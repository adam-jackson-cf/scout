"""Fail-closed synthetic evaluator draft. Never emits binding_gate_established."""
import itertools
import json
import math
from provenance import (verify_bundle, maximum_completed_inputs, digest, encode,
                        decode, number, INPUTS, validate_host)


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
    if any(type(flag) is not bool for flag in (integrity, violating, adequate, unresolved)):
        raise ValueError("verdict flags must be bool")
    if integrity is False:
        return "inconclusive"
    if violating is True:
        return "not supported"
    if (adequate is False or not all(type(x) in (int, float) and math.isfinite(x)
                                    for x in (point, lower, upper)) or lower > upper):
        return "inconclusive"
    if upper < .10:
        return "not supported"
    if unresolved:
        return "inconclusive"
    if point >= .10 and lower > 0:
        return "support"
    return "inconclusive"


def resource_fit(inputs, inventory, manifest, host):
    """Signed rev8 binding arithmetic; maxima are per operation, never run totals."""
    validate_host(host)
    if (set(inputs) != INPUTS or not all(number(v) for v in inputs.values())
        or not number(inventory["p_gate"], positive=True) or inventory["p_gate"] > 1):
        raise ValueError("malformed resource measurement or p_gate")
    units = sum(sum(sizes) for sizes in inventory["eligible_sizes"].values())
    p_gate = inventory["p_gate"]
    requested = manifest["requested_latent_coverage_cell_count"]
    references = manifest["calibrated_power_and_null"]["additional_reference_cells"]
    missing = manifest["missing_label_grid"]["cell_count"]
    safety = manifest["final_safety_OC"]["base_profile_count"]
    populations = manifest["populations"]["per_cell"]
    complete = requested + references
    calibrated = (references + missing + safety) * populations
    fraction = (references + missing) / (references + missing + safety)
    seconds = (
        (complete + missing)*populations*units*inputs["construction_cpu_per_unit"]
        + requested*populations*units*inputs["requested_base_cpu_per_unit"]
        + calibrated*(units*inputs["calibration_cpu_per_unit"]
                      + inputs["enumeration_cpu_per_population"])*fraction
        + populations*4000*(complete*inputs["complete_trial_cpu"] + missing*inputs["full512_trial_cpu"])
        + (complete+missing)*populations*math.ceil(4000/p_gate)*inputs["gate_cpu_per_check"])
    hours = 2*seconds/3600
    wall = hours/192
    rss = max(inputs[k] for k in ("construction_rss", "calibration_rss", "trial_plus_helper_rss"))
    return {"passed": hours <= 2880 and wall <= 15 and 8*rss <= host["hw.memsize"]-8*1024**3,
            "workers": 8, "core_hours": hours, "wall_days": wall,
            "memory_bytes": host["hw.memsize"], "eight_worker_rss_bytes": 8*rss,
            "maximum_inputs": inputs, "units_per_population": units}


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


def evaluate(manifest_bytes, *, provenance_payloads, provenance_digests, synthetic_only):
    """Pre-dev only: every evaluated input is decoded from verified bundle bytes."""
    if synthetic_only is not True:
        raise ValueError("draft accepts synthetic fixtures only")
    provenance = verify_bundle(provenance_payloads, provenance_digests)
    if type(manifest_bytes) is not bytes or manifest_bytes != provenance_payloads["signed_manifest"]:
        raise ValueError("manifest is not the verified signed bytes")
    manifest = decode(manifest_bytes)
    if manifest["bootstrap_replicates"] != 1999:
        raise ValueError("B must be 1999")
    inventory = decode(provenance_payloads["locked_inventory"])
    p_gate = inventory["p_gate"]
    successes = inventory["gate_successes"]
    if (type(inventory["gate_redraws"]) is not int or inventory["gate_redraws"] != 100000 or type(successes) is not int
        or not 0 < successes <= 100000 or not number(p_gate, positive=True)
        or p_gate != successes/100000):
        raise ValueError("unmeasured/inconsistent p_gate")
    sizes = inventory["eligible_sizes"]
    if (type(sizes) is not dict or len(sizes) != 3
        or any(type(v) is not list or not 20 <= len(v) <= 2000
               or any(type(n) is not int or not 20 <= n <= 2000 for n in v) for v in sizes.values())):
        raise ValueError("incomplete locked eligible sizes")
    oc = decode(provenance_payloads["sealed_oc_output"])
    if (oc["locked_inventory_sha256"] != provenance["locked_inventory"]
        or oc["signed_manifest_sha256"] != provenance["signed_manifest"]
        or oc["records_sha256"] != digest(encode(oc["records"]))):
        raise ValueError("OC input/record seal mismatch")
    records = oc["records"]
    measurements = decode(provenance_payloads["measurement_runs"])
    if measurements["locked_inventory_sha256"] != provenance["locked_inventory"]:
        raise ValueError("measurement inventory mismatch")
    maximum, run_reports = maximum_completed_inputs(measurements["runs"], oc["host"])
    resources = resource_fit(maximum, inventory, manifest, oc["host"])
    expected = {identity(c): c for c in expected_cells(manifest)}
    found = {}
    for record in records:
        key = (identity(record["cell"]), record["population"])
        if key in found or key[0] not in expected or type(key[1]) is not int or key[1] not in range(5):
            raise ValueError("duplicate or unexpected cell/population")
        found[key] = record
    if len(found) != 5*len(expected):
        raise ValueError("missing cell/population")
    passed = {m: resources["passed"] for m in manifest["coverage_dimensions"]["cap_M"]}
    reports = {}
    for key, cell in expected.items():
        rows = [found[key, p] for p in range(5)]
        good = True
        for r in rows:
            counts = [r[n] for n in ("attempts", "executions", "eligible", "inadequate", "gate_failures", "lower_miss", "upper_miss", "support")]
            if any(type(n) is not int or n < 0 for n in counts):
                raise ValueError("noninteger/negative count")
            e, a, n = r["executions"], r["attempts"], r["eligible"]
            if (type(r["B"]) is not int or r["B"] != 1999 or n != 2000 or not n <= e <= 4000 or not e <= a <= math.ceil(4000/p_gate)
                or r["inadequate"] != e-n or r["gate_failures"] != a-e
                or any(r[name] > n for name in ("lower_miss", "upper_miss", "support"))):
                good = False
            if not good:
                continue
            # Tail gate is pooled; report each population separately below.
            if cell["kind"] == "calibrated-reference":
                realised = r["realised_gain"]
                if not number(realised) and not (type(realised) in (int, float) and math.isfinite(realised) and realised < 0):
                    good = False
                elif r["reachable"] is not True:
                    good = False
                elif cell["target"] == 0:
                    good &= realised <= 0
                else:
                    good &= realised >= .20 and wilson(r["inadequate"], e)[1] <= .05
            elif cell["kind"] == "missing":
                good &= r["mask_reachable"] is True and r["reachable"] is True
                realised = r["realised_gain"]
                target = cell["realised_gain"]
                good &= (type(realised) in (int, float) and math.isfinite(realised)
                         and (realised <= 0 if target == 0 else realised >= target))
        if good:
            good &= all(wilson(sum(r[name] for r in rows), 10000)[1] <= .035 for name in ("lower_miss", "upper_miss"))
            if cell["kind"] == "calibrated-reference":
                interval = wilson(sum(r["support"] for r in rows), 10000)
                good &= interval[1] <= .05 if cell["target"] == 0 else interval[0] >= .80
        passed[cell["cap_M"]] &= bool(good)
        reports[key] = {"passed": bool(good), "pooled": summarize(rows),
                        "populations": [summarize([r]) for r in rows]}
    result = {"status": "draft_synthetic_rows_passed" if any(passed.values()) else "not testable before dev",
              "rows": passed, "cell_reports": reports, "resources": resources,
              "measurement_runs": run_reports, "provenance_sha256": provenance,
              "authority": "none; synthetic draft, scoped review required"}
    sealed = encode(result)
    return {"result": result, "sealed_result": sealed, "result_sha256": digest(sealed)}


def select_postdev(sealed_predev, predev_sha256, *, provenance_payloads, provenance_digests,
                   synthetic_only):
    """Select only from externally sealed pre-dev rows; never recompute the gate."""
    if synthetic_only is not True or type(sealed_predev) is not bytes or digest(sealed_predev) != predev_sha256:
        raise ValueError("substituted pre-dev result")
    provenance = verify_bundle(provenance_payloads, provenance_digests, postdev=True)
    prior = decode(sealed_predev)
    if prior["provenance_sha256"] != {k: v for k, v in provenance.items() if k != "mhat_evidence"}:
        raise ValueError("pre-dev input bundle changed")
    evidence = decode(provenance_payloads["mhat_evidence"])
    if (evidence["predev_sha256"] != predev_sha256
        or evidence["locked_inventory_sha256"] != provenance["locked_inventory"]
        or not number(evidence["c_U"], positive=True) or type(evidence["Mhat"]) is not int
        or not 0 <= evidence["Mhat"] <= 2000
        or not evidence["bootstrap_evidence"]):
        raise ValueError("unbound/unmeasured Mhat/c_U evidence")
    expected_mhat = 2000 if evidence["c_U"] <= .50/2000 else math.floor(.50/evidence["c_U"])
    if evidence["Mhat"] != expected_mhat:
        raise ValueError("Mhat contradicts sealed c_U")
    candidates = [int(m) for m in prior["rows"] if int(m) <= evidence["Mhat"]]
    selected = max(candidates) if candidates else None
    return {"status": "draft_synthetic_gate_passed" if selected is not None and prior["rows"][str(selected)] is True else "not testable",
            "selected_M": selected, "predev_sha256": predev_sha256,
            "mhat_evidence_sha256": provenance["mhat_evidence"], "authority": prior["authority"]}
