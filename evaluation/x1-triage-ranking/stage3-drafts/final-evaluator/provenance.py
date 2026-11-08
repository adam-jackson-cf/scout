"""Draft provenance checks over supplied synthetic records, not host attestation."""
import hashlib
import math


def verify_bundle(payloads, expected_digests):
    """Trust anchors come from the external scoped review, never the bundle itself.

    Exact membership prevents omitted closures. Byte hashes bind locked inputs,
    source/runtime closure, manifest, controller freeze and council record alike.
    This checks integrity, not whether a council record is authentically approved.
    """
    required = {"locked_inventory", "signed_manifest", "source_runtime_closure", "controller_freeze", "scoped_council_record"}
    if set(payloads) != required or set(expected_digests) != required:
        raise ValueError("incomplete provenance bundle")
    if any(hashlib.sha256(payloads[k]).hexdigest() != expected_digests[k] for k in required):
        raise ValueError("substituted/stale provenance")
    return {k: expected_digests[k] for k in sorted(required)}


def maximum_completed_inputs(runs, binding_host):
    """One run; a qualifying failure alone permits the next create-once record.

    Completed closure-invalid runs still enter the maxima (R9-reading-1).
    Host load/termination never permit a rerun; binding host cannot change.
    """
    if not runs:
        raise ValueError("unmeasured resource fit")
    required_host = {"hw.model", "hw.memsize", "logical_cpus", "os_build"}
    if set(binding_host) != required_host or any(not v for v in binding_host.values()):
        raise ValueError("missing host identity")
    maxima = {}
    previous_failure = None
    seen = set()
    for i, run in enumerate(runs):
        if run["id"] in seen or (i and previous_failure is None):
            raise ValueError("duplicate or unauthorised rerun")
        seen.add(run["id"])
        if run["host"] != binding_host or len(run["load_averages"]) != 3 or run["concurrent_process_count"] < 1:
            raise ValueError("host mismatch or missing load provenance")
        cause = run["integrity_failure"]
        if cause not in (None, "non-resource crash", "closure mismatch", "independent OS malfunction"):
            raise ValueError("failed/unmeasured fit cannot be rerun")
        if cause and not run["preserved_evidence_sha256"]:
            raise ValueError("failure evidence absent")
        if cause == "independent OS malfunction" and not run["contemporaneous_os_evidence_sha256"]:
            raise ValueError("independent OS evidence absent")
        if run["completed"]:
            values = run["binding_inputs"]
            if maxima and set(values) != set(maxima):
                raise ValueError("inconsistent binding input membership")
            if not values or any(not math.isfinite(v) or v < 0 for v in values.values()):
                raise ValueError("unmeasured binding inputs")
            maxima = {k: max(maxima.get(k, 0), v) for k, v in values.items()}
        elif cause is None:
            raise ValueError("interrupted without qualifying integrity failure")
        previous_failure = cause
    if previous_failure is not None or not maxima:
        raise ValueError("remeasurement unresolved")
    return maxima
