"""Independent recipe: no controller or signed-kernel imports; OpenSSL stdin."""
from decimal import Decimal, localcontext
import itertools
import json
from pathlib import Path
import subprocess

HERE = Path(__file__).resolve().parent


def compact(value):
    return json.dumps(value, sort_keys=True, separators=(",", ":"), allow_nan=False).encode()


def openssl(data, key=None):
    command = ["openssl", "dgst", "-sha256"]
    if key is not None:
        command += ["-hmac", key]
    return subprocess.check_output(command, input=data).decode().strip().split()[-1]


def derive():
    manifest = json.loads((HERE.parents[1] / "oc-manifest.json").read_bytes())
    dimensions = manifest["final_safety_OC"]["dimensions"]
    names = ("cap_M", "prevalence", "realised_gain", "scanner_control_true_at_prevalence", "rank_structure", "size_dependence")
    assert set(dimensions) == set(names)
    rows = []
    for values in itertools.product(*(dimensions[name] for name in names)):
        for kappa in manifest["final_safety_OC"]["unresolved_kappa"]:
            cell = compact(["safety-interval-group", {**dict(zip(names, values)), "unresolved_kappa": kappa}]).decode()
            rows.append((openssl(cell.encode(), "169-x1-safety-order"), cell))
    rows.sort(key=lambda pair: (pair[0], pair[1].encode()))
    ordered = [cell for _, cell in rows]
    first = ordered[0]
    specifications = [(first, 0, 0, "population"), (first, 1, 0, "audit"),
                      (first, 0, 17, "bootstrap"), ({"kind": "sampling-gate-estimate"}, 0, 0, "sampling-gate-estimate")]
    seeds = []
    for cell, population, attempt, phase in specifications:
        cell_bytes = cell.encode() if isinstance(cell, str) else compact(cell)
        full = openssl(b"\0".join((cell_bytes, str(population).encode(), str(attempt).encode(), phase.encode())), "169-x1-oc-v4")
        seeds.append(dict(cell=cell, population=population, attempt=attempt, phase=phase, hmac_sha256=full, seed64=int(full[:16], 16)))
    with localcontext() as context:
        context.prec = 60
        z, k, n = Decimal("1.959963984540054"), Decimal(50), Decimal(2000)
        p = k/n
        centre = (p+z*z/(2*n))/(1+z*z/n)
        half = z*(p*(1-p)/n+z*z/(4*n*n)).sqrt()/(1+z*z/n)
        bounds = [str(centre-half), str(centre+half)]
    return {"method": "Independent recipe in derive_known_answers.py; every order HMAC and SHA uses openssl; policy is a checksum of published policy.json, not an independent design derivation",
            "group_count": len(ordered), "first_identity": first, "last_identity": ordered[-1],
            "hmac_order_first": rows[0][0], "ordered_groups_sha256": openssl(compact(ordered)),
            "policy_sha256": openssl(compact(json.loads((HERE / "policy.json").read_bytes()))),
            "seeds": seeds, "wilson_50_2000_decimal60": bounds}


if __name__ == "__main__":
    print(json.dumps(derive(), indent=2, sort_keys=True))
