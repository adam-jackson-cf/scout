#!/usr/bin/env python3
"""Record the X2d round-0 onboarding maps (X2d PREREGISTRATION sections 4.1, 4.2, 12; README stage 1). Needs a logged-in `codex` CLI; NOT run during review.

    source .tools/activate
    .tools/venv/bin/python evaluation/x2d-merge-association-pilot/collect_onboard.py --source /tmp/scout-x2c-smoke/_src/fastapi [--data D] [--work DIR] [--node node] [--cli $SCOUT_CLAWPATCH_BINARY]

One workspace at C0 (`x2d.C0`), then sequentially: for every scope of `prepass/scopes.json` (in file order) and draw k = 1..3 one fresh Clawpatch state directory outside the
workspace: `--root WS --state-dir S --json init`, a rewrite of `<S>/config.json` `include` to the scope's `include` entries (nothing else in the file is touched, in particular
never `codexConfig`; the procedure of `measurement/lever_check.py`), then `map --source agent --provider codex --model gpt-6-luna --reasoning-effort xhigh` (x2d.MAPPER);
finally the unscoped contrast draw (the default include, same flags). 15 + 1 = 16 maps, the cap `x2d.CAPS["clawpatch_maps_round_0"]`. The environment of every call is Scout's
(`common.clawpatch_map_env`: credentials scrubbed, host auth context, both provider timeouts). Each map's `features/` and a `meta.json` are copied to
`D/onboard/draws/<scope>/<k>/` (contrast: `D/onboard/contrast/`); every attempt (a failed map is retried once) is kept as `D/onboard/attempts/<scope>-<k>.attempt<N>.json`
(contrast: `contrast.attempt<N>.json`); `D/onboard/versions.json` records the tool versions, the pins and the environment names (never values).

Preflight refusals, before any clone or model call (as X2c's collect_x2c.py): data of this stage already present (`D/onboard/{draws,contrast,attempts,versions.json,stopped.json}`;
append-only), the workspace directory exists, any CLAWPATCH_MODEL/REASONING_EFFORT/PROVIDER/CONFIG (`common.REFUSED_MODEL_ENV`) or CLAWPATCH_CODEX_SANDBOX name set, `codex` not on
PATH, Clawpatch `--version` not 0.7.3, the CLI is not the review-safety build (`common.build_safety_check`), `x2d.verify_pins()` problems (no sign-off record included), the scopes file
not the C0 partition that fits the caps. Stop rule (X2c): before the next map after --wall-max seconds or three consecutive failed maps `D/onboard/stopped.json` is written and the
remaining maps are not attempted (`unmeasured`); an abort (Ctrl-C, SIGTERM, SIGHUP unless ignored, an exception) ends the running process group and writes the same record.

Validation-only options: `--rev`, `--scopes`, `--expect-clawpatch-version`, `--wall-max`, `--call-timeout`, `--term-grace`; environment `X2D_SKIP_SIGNOFF=1` makes the pin check
`x2d.verify_pins(require_signoff=False)` (stub runs, before the council's sign-off record exists). Departures from collect_x2c.py: no model probe (the maps name the model
explicitly; the stage makes no model call other than the maps), and the process helpers (`sh`, `run_limited`, ...) are copied from it (X2c is not a frozen dependency of X2d).
"""
import argparse
import json
import os
import re
import shutil
import signal
import subprocess
import sys
import time

import x2d  # noqa: E402 (puts the frozen X2 directory and this directory on sys.path)
import common as C  # noqa: E402 (X2's frozen common.py)

ROOT = x2d.ROOT
SCOPES_PATH = os.path.join(x2d.HERE, "prepass", "scopes.json")


# ---- process helpers (copied from evaluation/x2c-mapper-comparison/collect_x2c.py)
def sh(args, cwd=None, env=None, check=True, timeout=None):
    try:
        r = subprocess.run(args, cwd=cwd, env=env, capture_output=True, text=True, timeout=timeout)
    except subprocess.TimeoutExpired:
        sys.exit("command timed out after %ss: %s" % (timeout, " ".join(args)))
    if check and r.returncode != 0:
        sys.exit("command failed: %s\n%s" % (" ".join(args), r.stderr[-2000:]))
    return r


def version_of(args):
    return sh(args, check=False, timeout=C.VERSION_TIMEOUT_SECONDS).stdout.strip() or "unknown"


def terminate_group(p, grace):
    """SIGTERM the process group, wait `grace` seconds, then SIGKILL; a final SIGKILL sweeps leftovers. -> the signal that ended the process (SIGTERM, SIGKILL or None)."""
    sig = "SIGTERM"
    try:
        os.killpg(p.pid, signal.SIGTERM)
    except ProcessLookupError:
        return None
    try:
        p.wait(timeout=grace)
    except subprocess.TimeoutExpired:
        sig = "SIGKILL"
    try:
        os.killpg(p.pid, signal.SIGKILL)
    except ProcessLookupError:
        pass
    p.wait()
    return sig


def drain(p, grace):
    try:
        out, err = p.communicate(timeout=grace)
    except subprocess.TimeoutExpired:  # a surviving child still holds the pipes
        return "", "(output not drained: a surviving child holds the pipe)"
    return out or "", err or ""


def run_limited(cmd, env, timeout, cwd=None, grace=C.TERM_GRACE_SECONDS):
    """Run cmd in its own process group with a wall-clock limit. -> (returncode, stdout, stderr, timed_out, ended_by); rc 124 on timeout."""
    p = subprocess.Popen(cmd, cwd=cwd, env=env, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, start_new_session=True)
    try:
        out, err = p.communicate(timeout=timeout)
        return p.returncode, out, err, False, None
    except subprocess.TimeoutExpired:
        sig = terminate_group(p, grace)
        out, err = drain(p, grace)
        return 124, out, err, True, sig
    except BaseException:
        terminate_group(p, grace)
        drain(p, grace)
        raise


def progress_events(stderr):
    """Clawpatch progress lines are `clawpatch map <event> k=v ...` on stderr (third_party/clawpatch/src/progress.ts)."""
    ev = []
    for line in stderr.splitlines():
        m = re.match(r"^clawpatch map (\S+)(?: (.*))?$", line)
        if m:
            ev.append({"event": m.group(1), "fields": dict(kv.split("=", 1) for kv in (m.group(2) or "").split() if "=" in kv)})
    return ev[:200]


# ---- the plan and its paths
def onboard_dir(a, *parts):
    return os.path.join(a.data, "onboard", *parts)


def plan_of(scopes):
    """[(map_id, scope id | None, k | None, include list | None, destination dir relative to D/onboard, attempt file stem)]: scopes in file order, k = 1..3, then the contrast."""
    plan = []
    for s in scopes:
        for k in (1, 2, 3):
            plan.append(("%s/%d" % (s["id"], k), s["id"], k, list(s["include"]), os.path.join("draws", s["id"], str(k)), "%s-%d" % (s["id"], k)))
    plan.append(("contrast", None, None, None, "contrast", "contrast"))
    return plan


def load_scopes(path, rev):
    with open(path) as f:
        sc = json.load(f)
    if sc.get("rev") != rev:
        sys.exit("refusing to start: %s is for revision %s, not %s" % (path, sc.get("rev"), rev))
    if not sc.get("all_fit"):
        sys.exit("refusing to start: %s says the scopes do not all fit the inventory caps" % path)
    ids = [s["id"] for s in sc["scopes"]]
    if not ids or len(set(ids)) != len(ids) or any(not s.get("include") for s in sc["scopes"]):
        sys.exit("refusing to start: %s has no scopes, duplicate scope ids or an empty include" % path)
    return sc["scopes"]


def preflight(a, scopes):
    clash = [p for p in (onboard_dir(a, "draws"), onboard_dir(a, "contrast"), onboard_dir(a, "attempts"), onboard_dir(a, "versions.json"), onboard_dir(a, "stopped.json"))
             if os.path.exists(p)]
    if clash:
        sys.exit("refusing to start: already recorded (append-only): %s" % ", ".join(sorted(clash)))
    base = os.path.join(a.work, "onboard")
    if os.path.exists(base):
        sys.exit("workspace exists; remove %s to start a clean collection" % base)
    set_names = [n for n in (*C.REFUSED_MODEL_ENV, *C.REFUSED_SANDBOX_ENV) if n in os.environ]
    if set_names:
        sys.exit("refusing to start: %s set in the environment (they change the model, reasoning effort, provider or sandbox Clawpatch uses; defaults are required)" % ", ".join(set_names))
    if not shutil.which(x2d.MAPPER["provider"]):
        sys.exit("refusing to start: the %s CLI is not on PATH (the maps need it)" % x2d.MAPPER["provider"])
    if len(plan_of(scopes)) > x2d.CAPS["clawpatch_maps_round_0"]:
        sys.exit("refusing to start: %d maps exceed the cap of %d" % (len(plan_of(scopes)), x2d.CAPS["clawpatch_maps_round_0"]))
    out = sh([a.node, a.cli, "--version"], check=False, timeout=C.VERSION_TIMEOUT_SECONDS).stdout
    if not re.search(r"(?<![\w.])" + re.escape(a.expect_clawpatch_version) + r"(?![\w.])", out):
        sys.exit("refusing to start: Clawpatch --version %r is not the pinned %s" % (out.strip(), a.expect_clawpatch_version))
    chk = C.build_safety_check(a.cli)
    if not chk["ok"]:
        sys.exit("refusing to start: Clawpatch CLI %s is not the review-safety build (%s); rebuild with scripts/bootstrap-tools.sh" % (a.cli, chk["reason"]))
    bad = x2d.verify_pins(require_signoff=x2d.require_signoff())
    if bad:
        sys.exit("refusing to start: pins do not hold: %s" % "; ".join(bad))


def tool_versions(a, ws, scopes_sha):
    env = C.clawpatch_map_env(os.environ)
    head = sh(["git", "rev-parse", "HEAD"], cwd=ROOT, check=False)
    return {"schema": "x2d-onboard-versions-v1", "clawpatch": version_of([a.node, a.cli, "--version"]), "clawpatch_cli_sha256": C.sha256_file(a.cli),
            "clawpatch_build_safety": C.build_safety_check(a.cli), "node": version_of([a.node, "--version"]),
            "codex": version_of(["codex", "--version"]) if shutil.which("codex") else None,
            "scout_head": head.stdout.strip() or "unknown",
            "areas_go_sha256": C.sha256_file(os.path.join(ROOT, "internal", "areas", "areas.go")),
            "records_go_sha256": C.sha256_file(os.path.join(ROOT, "internal", "clawpatch", "records.go")),
            "env_names": sorted(env), "env_presence": {n: n in os.environ for n in C.RECORDED_PRESENCE_ENV},  # names only, never values
            "claude_auth_context": env.get("CLAWPATCH_CLAUDE_AUTH_CONTEXT"), "codex_timeout_ms": env.get("CLAWPATCH_CODEX_TIMEOUT_MS"),
            "call_timeout_seconds": a.call_timeout, "wall_max_seconds": a.wall_max, "term_grace_seconds": a.term_grace,
            "workspace_config_presence": {p: os.path.exists(os.path.join(ws, p)) for p in C.WORKSPACE_CONFIG_PATHS},
            "map_flags": map_flags(), "rev": a.rev, "scopes_sha256": scopes_sha, "collect_onboard_sha256": C.sha256_file(os.path.abspath(__file__)),
            "started": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime()), "finished": None, "maps": None}


def map_flags():
    m = x2d.MAPPER
    return ["--source", "agent", "--provider", m["provider"], "--model", m["model"], "--reasoning-effort", m["effort"]]


def write_versions(a, v):
    """First write creates the file exclusively (append-only across invocations); the end of the run rewrites it once to add `finished` and `maps`."""
    path = onboard_dir(a, "versions.json")
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w" if v.get("_created") else "x") as f:
        v["_created"] = True
        json.dump({k: x for k, x in v.items() if k != "_created"}, f, indent=1)


def record(a, dest_rel, state, meta):
    """Copy the state's features/ and write meta.json (append-only)."""
    dest = onboard_dir(a, dest_rel)
    if os.path.exists(dest):
        sys.exit("refusing to overwrite recorded draw %s" % dest)
    os.makedirs(dest)
    src = os.path.join(state, "features")
    if os.path.isdir(src):
        shutil.copytree(src, os.path.join(dest, "features"))
    else:
        os.makedirs(os.path.join(dest, "features"))
    with open(os.path.join(dest, "meta.json"), "x") as f:
        json.dump(meta, f, indent=1)


def set_include(state, include):
    """The lever of measurement/lever_check.py: rewrite only `include` of <state>/config.json (never `codexConfig` or anything else)."""
    cfg_path = os.path.join(state, "config.json")
    with open(cfg_path) as f:
        cfg = json.load(f)
    cfg["include"] = list(include)
    with open(cfg_path, "w") as f:
        json.dump(cfg, f, indent=2)


def run_map(a, ws, base, entry):
    map_id, scope, k, include, dest_rel, stem = entry
    env = C.clawpatch_map_env(os.environ)
    state = os.path.join(base, "state-" + stem)
    attempts_dir = onboard_dir(a, "attempts")
    os.makedirs(attempts_dir, exist_ok=True)
    m = x2d.MAPPER
    meta = {"schema": "x2d-onboard-draw-v1", "map_id": map_id, "scope": scope, "draw": k, "contrast": scope is None, "include": include,
            "provider": m["provider"], "map_model": m["model"], "map_reasoning_effort": m["effort"], "map_flags": map_flags(), "rev": a.rev,
            "tree": sh(["git", "rev-parse", "HEAD^{tree}"], cwd=ws).stdout.strip()}
    attempts, rc, tail = [], 1, ""
    for attempt in (1, 2):  # one retry for an infrastructure failure; both attempts are kept
        t0, started_utc = time.time(), time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
        shutil.rmtree(state, ignore_errors=True)
        base_cmd = [a.node, a.cli, "--root", ws, "--state-dir", state, "--json"]
        rc, out, err, timed_out, ended_by = run_limited(base_cmd + ["init"], env, a.call_timeout, grace=a.term_grace)
        if rc == 0 and include is not None:
            set_include(state, include)
        if rc == 0:
            rc, out, err, timed_out, ended_by = run_limited(base_cmd + ["map", *map_flags()], env, a.call_timeout, grace=a.term_grace)
        seconds = round(time.time() - t0, 1)
        ev = progress_events(err)
        start = next((e["fields"] for e in ev if e["event"] == "agent-start"), {})
        tail = err[-3000:]
        rec = {"attempt": attempt, "started_utc": started_utc, "rc": rc, "seconds": seconds, "timed_out": timed_out, "ended_by": ended_by, "provider": start.get("provider"),
               "model_option": start.get("model")}
        attempts.append(rec)
        with open(os.path.join(attempts_dir, "%s.attempt%d.json" % (stem, attempt)), "x") as f:
            json.dump({**meta, **rec, "progress": ev, "stderr_tail": tail, "stdout_tail": out[-3000:]}, f, indent=1)
        if rc == 0:
            break
    meta.update({"rc": rc, "attempt_count": len(attempts), "attempts": attempts, "seconds_total": round(sum(x["seconds"] for x in attempts), 1),
                 "seconds_last": attempts[-1]["seconds"], "provider_reported": attempts[-1]["provider"], "model_option_reported": attempts[-1]["model_option"],
                 "stderr_tail": tail[-1500:], "workspace_dirty": bool(sh(["git", "status", "--porcelain"], cwd=ws).stdout.strip())})
    record(a, dest_rel, state, meta)
    shutil.rmtree(state, ignore_errors=True)
    print("[onboard] %s rc=%s %ss attempts=%d" % (map_id, rc, meta["seconds_total"], len(attempts)), flush=True)
    return rc == 0


def _term(signum, _frame):
    raise KeyboardInterrupt("signal %d" % signum)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--source", required=True, help="path or URL of a clone containing the pinned revision")
    ap.add_argument("--data", default=x2d.DATA)
    ap.add_argument("--work", default="/tmp/scout-x2d")
    ap.add_argument("--node", default=shutil.which("node"))
    ap.add_argument("--cli", default=os.environ.get("SCOUT_CLAWPATCH_BINARY"))
    ap.add_argument("--rev", default=x2d.C0, help="revision to map; preregistered C0 (validation overrides)")
    ap.add_argument("--scopes", default=SCOPES_PATH, help="scopes file; prepass/scopes.json (validation overrides)")
    ap.add_argument("--wall-max", type=float, default=C.WALL_MAX_SECONDS, help="seconds; X2 default 8 h (validation overrides)")
    ap.add_argument("--call-timeout", type=float, default=C.CALL_TIMEOUT_SECONDS, help="seconds per Clawpatch call; X2 default 1800")
    ap.add_argument("--term-grace", type=float, default=C.TERM_GRACE_SECONDS, help="seconds between SIGTERM and SIGKILL; X2 default 10")
    ap.add_argument("--expect-clawpatch-version", default=x2d.CLAWPATCH_VERSION, help="pinned Clawpatch version 0.7.3 (validation overrides for its stub)")
    a = ap.parse_args()
    t_start = time.time()
    for sig in (signal.SIGTERM, signal.SIGHUP):
        if sig == signal.SIGHUP and signal.getsignal(sig) is signal.SIG_IGN:  # `nohup` protection wins
            continue
        signal.signal(sig, _term)
    if not a.node or not a.cli:
        sys.exit("need --node and --cli (or SCOUT_CLAWPATCH_BINARY)")
    scopes = load_scopes(a.scopes, a.rev)
    plan = plan_of(scopes)
    preflight(a, scopes)
    base = os.path.join(a.work, "onboard")  # the preflight has refused an existing `base`
    ws = os.path.join(base, "ws")
    os.makedirs(base)
    sh(["git", "clone", "-q", "--no-local", a.source, ws])
    sh(["git", "checkout", "-q", "--detach", a.rev], cwd=ws)
    if sh(["git", "rev-parse", "HEAD"], cwd=ws).stdout.strip() != a.rev:
        sys.exit("pinned revision mismatch")
    tv = tool_versions(a, ws, C.sha256_file(a.scopes))
    stop_path = onboard_dir(a, "stopped.json")

    def write_stop(reason, before, not_attempted):
        os.makedirs(os.path.dirname(stop_path), exist_ok=True)
        with open(stop_path, "x") as f:
            json.dump({"reason": reason, "elapsed_seconds": round(time.time() - t_start, 1), "stopped_before": before, "not_attempted": not_attempted}, f, indent=1)

    ok, consecutive, stopped, current = {}, 0, None, 0
    try:
        write_versions(a, tv)  # first statement of the protected block: from here an abort writes the stop record
        for current, entry in enumerate(plan):
            elapsed = time.time() - t_start
            if elapsed > a.wall_max:
                stopped = "wall-clock limit %.0fs exceeded (%.0fs elapsed)" % (a.wall_max, elapsed)
            elif consecutive >= C.CONSECUTIVE_FAIL_MAX:
                stopped = "%d consecutive failed maps" % consecutive
            if stopped:
                write_stop(stopped, entry[0], [p[0] for p in plan[current:]])
                print("[onboard] STOPPED before %s: %s" % (entry[0], stopped), flush=True)
                break
            ok[entry[0]] = run_map(a, ws, base, entry)
            consecutive = 0 if ok[entry[0]] else consecutive + 1
        tv["finished"] = time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
        tv["maps"] = ok
        write_versions(a, tv)
    except BaseException as e:  # Ctrl-C, SIGTERM/SIGHUP, any exception: run_limited has already ended the Clawpatch process group
        if not os.path.exists(stop_path):
            write_stop("aborted: %s: %s" % (type(e).__name__, e), plan[current][0], [p[0] for p in plan[current:]])
        tv["finished"], tv["maps"] = "aborted", ok
        write_versions(a, tv)
        raise
    print(json.dumps(ok))
    if not all(ok.get(p[0]) for p in plan):
        sys.exit(3)


if __name__ == "__main__":
    main()
