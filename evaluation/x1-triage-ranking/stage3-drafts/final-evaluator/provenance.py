"""Byte-bound synthetic provenance; no authenticity or experiment authority."""
import hashlib
import json
import math

MANIFEST_SHA256 = "1caea1a5ddba3a3d9eae99bba780e5de987ad7771f6cce06af9b0171be96761d"
INPUTS = {"construction_cpu_per_unit", "calibration_cpu_per_unit", "enumeration_cpu_per_population",
          "requested_base_cpu_per_unit", "gate_cpu_per_check", "complete_trial_cpu", "full512_trial_cpu",
          "construction_rss", "calibration_rss", "trial_plus_helper_rss"}
BASE_MEMBERS = {"locked_inventory", "signed_manifest", "source_runtime_closure", "controller_freeze",
                "scoped_council_record", "measurement_runs", "sealed_oc_output"}


def digest(data):
    return hashlib.sha256(data).hexdigest()


def encode(value):
    return json.dumps(value, sort_keys=True, separators=(",", ":"), allow_nan=False).encode()


def decode(data):
    def unique(pairs):
        result = {}
        for key, value in pairs:
            if key in result:
                raise ValueError("duplicate JSON key")
            result[key] = value
        return result
    try:
        return json.loads(data, object_pairs_hook=unique,
                          parse_constant=lambda x: (_ for _ in ()).throw(ValueError("nonfinite JSON")))
    except (TypeError, UnicodeError, json.JSONDecodeError) as exc:
        raise ValueError("malformed byte payload") from exc

def member_bytes(data, member):
    """Extract a top-level JSON value's exact UTF-8 bytes after strict decoding."""
    decode(data)  # Reject duplicate keys, malformed JSON and nonstandard literals first.
    text = data.decode("utf-8")
    decoder = json.JSONDecoder()

    def skip_space(index):
        while index < len(text) and text[index] in " \t\r\n":
            index += 1
        return index

    index = skip_space(0)
    if text[index] != "{":
        raise ValueError("expected JSON object")
    index = skip_space(index+1)
    while text[index] != "}":
        key, index = decoder.raw_decode(text, index)
        index = skip_space(index)
        if text[index] != ":":
            raise ValueError("expected member separator")
        start = skip_space(index+1)
        _, index = decoder.raw_decode(text, start)
        if key == member:
            return text[start:index].encode("utf-8")
        index = skip_space(index)
        if text[index] == "}":
            break
        index = skip_space(index+1)
    raise ValueError("missing JSON member")



def number(value, positive=False):
    return type(value) in (int, float) and math.isfinite(value) and (value > 0 if positive else value >= 0)


def verify_bundle(payloads, expected_digests, *, postdev=False):
    required = BASE_MEMBERS | ({"mhat_evidence"} if postdev else set())
    if set(payloads) != required or set(expected_digests) != required:
        raise ValueError("incomplete provenance bundle")
    if any(type(payloads[k]) is not bytes or digest(payloads[k]) != expected_digests[k] for k in required):
        raise ValueError("substituted/stale provenance")
    if digest(payloads["signed_manifest"]) != MANIFEST_SHA256:
        raise ValueError("substituted signed manifest bytes")
    return {k: expected_digests[k] for k in sorted(required)}


def validate_host(host):
    if (set(host) != {"hw.model", "hw.memsize", "logical_cpus", "os_build"}
        or any(type(host[k]) is not str or not host[k] for k in ("hw.model", "os_build"))
        or any(type(host[k]) is not int or host[k] <= 0 for k in ("hw.memsize", "logical_cpus"))):
        raise ValueError("missing host identity")


def maximum_completed_inputs(runs, binding_host):
    """Every completed run contributes, including completed closure failures."""
    validate_host(binding_host)
    if not runs:
        raise ValueError("unmeasured resource fit")
    maxima, seen, reports = {}, set(), []
    previous_failure = None
    for i, run in enumerate(runs):
        if type(run["id"]) is not str or not run["id"] or run["id"] in seen or (i and previous_failure is None):
            raise ValueError("duplicate or unauthorised rerun")
        seen.add(run["id"])
        if (run["host"] != binding_host or type(run.get("load_averages")) is not list
            or len(run["load_averages"]) != 3 or not all(number(v) for v in run["load_averages"])
            or type(run.get("concurrent_process_count")) is not int or run["concurrent_process_count"] < 1):
            raise ValueError("host mismatch or missing load provenance; pause for deviation")
        cause = run["integrity_failure"]
        if cause not in (None, "non-resource crash", "closure mismatch", "independent OS malfunction"):
            raise ValueError("failed/unmeasured fit cannot be rerun")
        for field in (["preserved_evidence_sha256"] if cause else []) + (["contemporaneous_os_evidence_sha256"] if cause == "independent OS malfunction" else []):
            value = run.get(field)
            if type(value) is not str or len(value) != 64 or any(c not in "0123456789abcdef" for c in value):
                raise ValueError("failure evidence absent")
        if type(run["completed"]) is not bool:
            raise ValueError("malformed completion flag")
        if run["completed"] is True:
            values = run["binding_inputs"]
            if set(values) != INPUTS or not all(number(v, positive=True) for v in values.values()):
                raise ValueError("unmeasured binding inputs")
            maxima = {k: max(maxima.get(k, 0), v) for k, v in values.items()}
        elif cause is None:
            raise ValueError("interrupted without qualifying integrity failure")
        reports.append({"run": run, "classification": "completed-invalid-included" if run["completed"] and cause else "completed-valid-included" if run["completed"] else "interrupted-excluded"})
        previous_failure = cause
    if previous_failure is not None or not maxima:
        raise ValueError("remeasurement unresolved")
    return maxima, reports
