#!/usr/bin/env python3
"""Record the X2c agent maps (X2c PREREGISTRATION section 3; X2 PREREGISTRATION sections 2 and 9 otherwise). Needs logged-in `codex` and `claude` CLIs; NOT run during review.

COPY OF evaluation/x2-area-stability/collect.py (sha256 89948f3c...), changed ONLY as the pinned diff in validation/ records: the arm is one of the four preregistered mapper
arms (x2c.ARMS: provider, model, reasoning effort), every map call is Clawpatch invoked directly with `--provider P --model M --reasoning-effort E`, the model probe mirrors
that invocation (codex: banner model and effort; claude: modelUsage keys, effort as requested) and records both, the frozen-X2 pins are verified before anything runs, and the
`--arm/--work/--data` arguments are X2c's. Every X2 safeguard (preflight refusals, build-safety check, append-only bookkeeping, snapshot retry, per-call timeout, stop rule,
abort record, claude host-auth version check) is the X2 code. Original X2 text follows.

Per (arm, repo): one workspace at the pinned revision; maps in the fixed order F1 F2 F3 R1 R2 M1 M2.
  F1, R1, R2, M1, M2 share one Clawpatch state dir (same-state re-maps and mutation maps); F2 and F3 use fresh state dirs.
  M1 and M2 are structural mutations applied to the workspace and committed before their map; M2 accumulates on M1 (its tree contains M1).
Clawpatch is invoked with Scout's environment (common.clawpatch_map_env: modelcli.ScrubEnv port, CLAWPATCH_CLAUDE_AUTH_CONTEXT=host, and BOTH
CLAWPATCH_CODEX_TIMEOUT_MS and CLAWPATCH_CLAUDE_TIMEOUT_MS, default 900000, Scout's own value wins) and Scout's arguments (`map --source agent --provider <arm>`, `--json`, state outside the workspace).
Each map's complete state `features/` directory and a meta.json are copied under <data>/maps/<arm>/<repo>/<map id>/ (append-only: every
map directory, attempt file and versions file must be absent before anything is cloned). A failed map is retried once; both attempts are
kept under <data>/attempts/. A non-fresh map is retried from a snapshot of the chain state taken just before its first attempt.
Stop rule (section 9): this invocation (one arm, one repo) stops before the next map after --wall-max seconds or three consecutive failed
maps and writes <data>/stopped-<arm>-<repo>.json; maps not attempted are `unmeasured`. An abort (Ctrl-C, SIGTERM, SIGHUP unless it is inherited as ignored
(`nohup`), an exception) terminates the running Clawpatch process group (SIGTERM, grace, SIGKILL) and writes the same record with reason `aborted: <cause>`. A skipped M2 mutation runs no map:
its meta.json has `skipped` and `rc: null`. A failure or abort during the setup between the before-probe and the first map (workspace creation, clone, checkout,
version collection) writes no stop record and no data file; the rerun needs only the workspace removed (section 9).
Preflight (amendment 1) also refuses unless the Clawpatch CLI is the review-safety build Scout ships: `<build>/.scout-review-safety` (next to dist/) must equal the
pinned "<version> <sha256>" (common.review_safety_pins) and the built dist/ must contain the sanitizer (common.build_safety_check); the check is recorded in the versions file.
"""
import argparse
import glob
import json
import os
import re
import shutil
import signal
import subprocess
import sys
import tempfile
import time

import x2c  # noqa: E402 (puts the frozen X2 directory on sys.path)
import common as C  # noqa: E402 (X2's frozen common.py)

ROOT = os.path.abspath(os.path.join(C.HERE, "..", ".."))

NEWPKG = {
    ".py": {
        "x2_newpkg/__init__.py": '"""Currency and unit formatting helpers (X2 mutation M1)."""\n',
        "x2_newpkg/currency.py": 'def format_money(amount_cents: int, currency: str = "USD") -> str:\n    """Render an integer amount of cents as a currency string."""\n    return f"{amount_cents / 100:.2f} {currency}"\n\n\ndef parse_money(text: str) -> int:\n    """Parse a string such as "12.34 USD" into integer cents."""\n    return round(float(text.split()[0]) * 100)\n',
        "x2_newpkg/units.py": 'def km_to_miles(km: float) -> float:\n    """Convert kilometres to miles."""\n    return km * 0.621371\n\n\ndef celsius_to_fahrenheit(c: float) -> float:\n    """Convert degrees Celsius to Fahrenheit."""\n    return c * 9 / 5 + 32\n',
        "x2_newpkg/slugs.py": 'import re\n\n\ndef slugify(title: str) -> str:\n    """Turn a title into a lowercase URL slug."""\n    return re.sub(r"[^a-z0-9]+", "-", title.lower()).strip("-")\n',
        "x2_newpkg/pyproject.toml": '[project]\nname = "x2-newpkg"\nversion = "0.0.1"\n',
    },
    ".rs": {
        "x2_newpkg/Cargo.toml": '[package]\nname = "x2-newpkg"\nversion = "0.0.1"\nedition = "2021"\n',
        "x2_newpkg/src/lib.rs": "//! Currency and unit formatting helpers (X2 mutation M1).\npub mod currency;\npub mod units;\npub mod slugs;\n",
        "x2_newpkg/src/currency.rs": '/// Render an integer amount of cents as a currency string.\npub fn format_money(amount_cents: i64, currency: &str) -> String {\n    format!("{:.2} {}", amount_cents as f64 / 100.0, currency)\n}\n',
        "x2_newpkg/src/units.rs": "/// Convert kilometres to miles.\npub fn km_to_miles(km: f64) -> f64 {\n    km * 0.621371\n}\n\n/// Convert degrees Celsius to Fahrenheit.\npub fn celsius_to_fahrenheit(c: f64) -> f64 {\n    c * 9.0 / 5.0 + 32.0\n}\n",
        "x2_newpkg/src/slugs.rs": "/// Turn a title into a lowercase URL slug.\npub fn slugify(title: &str) -> String {\n    title.to_lowercase().split(|c: char| !c.is_ascii_alphanumeric()).filter(|s| !s.is_empty()).collect::<Vec<_>>().join(\"-\")\n}\n",
    },
}
# M2 exclusion: a directory is excluded when any of its path COMPONENTS (lower-cased) is one of these, or starts with "x2_".
EXCLUDED_COMPONENTS = {"test", "tests", "testing", "doc", "docs", "example", "examples", "bench", "benches", "benchmark", "benchmarks", ".github"}


def sh(args, cwd=None, env=None, check=True, timeout=None):
    try:
        r = subprocess.run(args, cwd=cwd, env=env, capture_output=True, text=True, timeout=timeout)
    except subprocess.TimeoutExpired:
        sys.exit("command timed out after %ss: %s" % (timeout, " ".join(args)))
    if check and r.returncode != 0:
        sys.exit("command failed: %s\n%s" % (" ".join(args), r.stderr[-2000:]))
    return r


def version_of(args):
    """`<tool> --version` text, bounded (section 9); 'unknown' when empty."""
    return sh(args, check=False, timeout=C.VERSION_TIMEOUT_SECONDS).stdout.strip() or "unknown"


def terminate_group(p, grace):
    """SIGTERM the process group (Clawpatch's own handler kills its detached provider children on SIGTERM, third_party/clawpatch/src/exec.ts), wait
    `grace` seconds, then SIGKILL; a final SIGKILL sweeps leftovers of the group. -> the signal that ended the process: SIGTERM, SIGKILL or None."""
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


def run_limited(cmd, env, timeout, cwd=None, grace=C.TERM_GRACE_SECONDS, inp=None):
    """Run cmd in its own process group with a wall-clock limit. -> (returncode, stdout, stderr, timed_out, ended_by); rc 124 on timeout, `ended_by` the
    signal that ended it (SIGTERM, grace, SIGKILL). Any abort (Ctrl-C, SIGTERM to the collector, an exception) terminates the group the same way."""
    p = subprocess.Popen(cmd, cwd=cwd, env=env, stdin=subprocess.PIPE if inp is not None else None, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True,
                         start_new_session=True)
    try:
        out, err = p.communicate(input=inp, timeout=timeout)
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
    """Clawpatch progress lines are `clawpatch map <event> k=v ...` on stderr (third_party/clawpatch/src/progress.ts), not JSON."""
    ev = []
    for line in stderr.splitlines():
        m = re.match(r"^clawpatch map (\S+)(?: (.*))?$", line)
        if m:
            ev.append({"event": m.group(1), "fields": dict(kv.split("=", 1) for kv in (m.group(2) or "").split() if "=" in kv)})
    return ev[:200]


def record(data, arm, repo, map_id, state, meta):
    """Copy the chain/fresh state's features/ and write meta.json (append-only). state None (a skipped mutation) records an empty features/."""
    dest = C.map_dir(data, arm, repo, map_id)
    if os.path.exists(dest):
        sys.exit("refusing to overwrite recorded map %s" % dest)
    os.makedirs(dest)
    src = os.path.join(state, "features") if state else None
    if src and os.path.isdir(src):
        shutil.copytree(src, os.path.join(dest, "features"))
    else:
        os.makedirs(os.path.join(dest, "features"))
    with open(os.path.join(dest, "meta.json"), "x") as f:
        json.dump(meta, f, indent=1)


def base_meta(a, spec, ws, map_id, fresh, extra_meta):
    arm = x2c.ARMS[a.arm]
    return {"schema": "x2c-map-v1", "arm": a.arm, "repo": a.repo, "map_id": map_id, "fresh": fresh, "provider": arm["provider"],
            "map_model": arm["model"], "map_reasoning_effort": arm["effort"], "map_flags": x2c.arm_flags(a.arm),
            "rev": spec["rev"], "tree": sh(["git", "rev-parse", "HEAD^{tree}"], cwd=ws).stdout.strip(), **extra_meta}


def run_map(a, spec, ws, state, map_id, fresh, extra_meta):
    env = C.clawpatch_map_env(os.environ)
    meta = base_meta(a, spec, ws, map_id, fresh, extra_meta)
    attempts_dir = os.path.join(a.data, "attempts", a.arm, a.repo)
    os.makedirs(attempts_dir, exist_ok=True)
    snapshot = None
    if not fresh and os.path.isdir(state):  # a failed attempt may leave partial writes in the chain state: retry from this copy
        snapshot = state + ".snapshot-" + map_id
        shutil.rmtree(snapshot, ignore_errors=True)
        shutil.copytree(state, snapshot)
    attempts, rc, tail = [], 1, ""
    for attempt in (1, 2):  # one retry for an infrastructure failure; both attempts are kept
        restored = False
        if attempt == 2 and snapshot is not None:
            shutil.rmtree(state, ignore_errors=True)
            shutil.copytree(snapshot, state)
            restored = True
        t0, started_utc = time.time(), time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())
        if fresh:
            shutil.rmtree(state, ignore_errors=True)
        rc, out, err, timed_out, ended_by = 0, "", "", False, None
        if fresh or not os.path.isdir(state):  # Scout's Init passes no map environment; the study gives init the same environment (section 2)
            rc, out, err, timed_out, ended_by = run_limited([a.node, a.cli, "--root", ws, "--state-dir", state, "--json", "init"], env, a.call_timeout, grace=a.term_grace)
        if rc == 0:
            rc, out, err, timed_out, ended_by = run_limited([a.node, a.cli, "--root", ws, "--state-dir", state, "--json", "map", "--source", "agent",
                                                             "--provider", x2c.ARMS[a.arm]["provider"], *x2c.arm_flags(a.arm)], env, a.call_timeout, grace=a.term_grace)
        seconds = round(time.time() - t0, 1)
        ev = progress_events(err)
        start = next((e["fields"] for e in ev if e["event"] == "agent-start"), {})
        tail = err[-3000:]
        rec = {"attempt": attempt, "started_utc": started_utc, "rc": rc, "seconds": seconds, "timed_out": timed_out, "ended_by": ended_by, "provider": start.get("provider"),
               "model_option": start.get("model"), "restored_snapshot": restored}
        attempts.append(rec)
        with open(os.path.join(attempts_dir, "%s.attempt%d.json" % (map_id, attempt)), "x") as f:
            json.dump({**meta, **rec, "progress": ev, "stderr_tail": tail, "stdout_tail": out[-3000:]}, f, indent=1)
        if rc == 0:
            break
    if snapshot is not None:
        shutil.rmtree(snapshot, ignore_errors=True)
    meta.update({"rc": rc, "attempt_count": len(attempts), "attempts": attempts, "seconds_total": round(sum(x["seconds"] for x in attempts), 1),
                 "seconds_last": attempts[-1]["seconds"], "provider_reported": attempts[-1]["provider"],
                 "model_option_reported": attempts[-1]["model_option"],  # Clawpatch's option value ("default"), NOT the served model (see the probes)
                 "stderr_tail": tail[-1500:],
                 "workspace_dirty": bool(sh(["git", "status", "--porcelain"], cwd=ws).stdout.strip())})  # a map must leave the checkout unchanged (section 2; a writable sandbox could not)
    record(a.data, a.arm, a.repo, map_id, state, meta)
    print("[%s/%s] %s rc=%s %ss attempts=%d" % (a.arm, a.repo, map_id, rc, meta["seconds_total"], len(attempts)), flush=True)
    return rc == 0


def record_skipped(a, spec, ws, map_id, fresh, extra_meta):
    """A skipped mutation runs no map: meta.json carries `skipped` and rc null, features/ is empty (the replay then reads the map as unmeasured)."""
    meta = {**base_meta(a, spec, ws, map_id, fresh, extra_meta), "rc": None, "attempt_count": 0, "attempts": []}
    record(a.data, a.arm, a.repo, map_id, None, meta)
    print("[%s/%s] %s SKIPPED (no map run): %s" % (a.arm, a.repo, map_id, extra_meta["skipped"]), flush=True)


def mutate_m1(ws, ext):
    for rel, text in NEWPKG[ext].items():
        os.makedirs(os.path.dirname(os.path.join(ws, rel)), exist_ok=True)
        with open(os.path.join(ws, rel), "w") as f:
            f.write(text)
    commit(ws, "X2 M1: add x2_newpkg (new unowned package with manifest)")
    return {"mutation": "M1", "added": sorted(NEWPKG[ext])}


def excluded_dir(d):
    return any(part.lower() in EXCLUDED_COMPONENTS or part.lower().startswith("x2_") for part in d.split("/"))


def select_rename_dir(ws, ext, data, arm, repo):
    """M2 target (deterministic, section 2): a leaf directory of the M1 tree with >= 3 tracked source files of the repo's language, no path
    component in EXCLUDED_COMPONENTS or starting with `x2_`, and >= 1 source file owned by an active agent-mapper feature of the M1 map;
    fewest source files first, then path ascending. None when no directory qualifies; "M1 not recorded" is signalled by the caller."""
    tracked = [t for t in sh(["git", "ls-files"], cwd=ws).stdout.split("\n") if t]
    dirs = {}
    for t in tracked:
        dirs.setdefault(os.path.dirname(t), []).append(t)
    m1 = C.load_active_features(data, arm, repo, "M1") or []
    owned = {p for f in m1 if f.get("source") == C.AGENT_SOURCE for p in C.owned_files(f)}
    cands = []
    for d, files in dirs.items():
        if not d or excluded_dir(d):
            continue
        if any(o != d and o.startswith(d + "/") for o in dirs):
            continue  # not a leaf
        src = [f for f in files if f.endswith(ext)]
        if len(src) >= 3 and any(f in owned for f in src):
            cands.append((len(src), d))
    if not cands:
        return None
    return sorted(cands)[0][1]


def mutate_m2(ws, ext, data, arm, repo):
    if C.load_active_features(data, arm, repo, "M1") is None:
        return {"mutation": "M2", "skipped": "M1 map not recorded or failed: the M2 target is selected from the M1 map"}
    d = select_rename_dir(ws, ext, data, arm, repo)
    if d is None:
        return {"mutation": "M2", "skipped": "no leaf directory qualifies"}
    sh(["git", "mv", d, d + "_x2moved"], cwd=ws)
    commit(ws, "X2 M2: rename leaf directory %s" % d)
    return {"mutation": "M2", "renamed": d, "to": d + "_x2moved"}


def commit(ws, msg):
    sh(["git", "add", "-A"], cwd=ws)
    sh(["git", "-c", "user.name=x2", "-c", "user.email=x2@example.invalid", "commit", "-q", "-m", msg], cwd=ws)


def preflight(a):
    """Refuse before any clone or model call: recorded maps, attempts, versions or a stop record already present for this (arm, repo), a workspace
    directory left over for it, or the arm's CLI not on PATH (Scout's resolveProvider requires it), so a missing CLI cannot burn attempts and leave an
    unrepeatable stop record and a leftover directory cannot cost a before-probe call."""
    clash = [C.map_dir(a.data, a.arm, a.repo, m) for m in C.MAP_IDS if os.path.exists(C.map_dir(a.data, a.arm, a.repo, m))]
    clash += glob.glob(os.path.join(a.data, "attempts", a.arm, a.repo, "*"))
    for name in ("tool-versions-%s-%s.json", "stopped-%s-%s.json"):
        p = os.path.join(a.data, name % (a.arm, a.repo))
        if os.path.exists(p):
            clash.append(p)
    if clash:
        sys.exit("refusing to start: already recorded (append-only): %s" % ", ".join(sorted(clash)))
    base = os.path.join(a.work, a.arm, a.repo)
    if os.path.exists(base):
        sys.exit("workspace exists; remove %s to start a clean collection" % base)
    set_names = [n for n in (*C.REFUSED_MODEL_ENV, *C.REFUSED_SANDBOX_ENV) if n in os.environ]
    if set_names:  # section 2: no Clawpatch-level model/effort/provider/config override and no sandbox override (Clawpatch's `??` lets even an empty value override)
        sys.exit("refusing to start: %s set in the environment (they change the model, reasoning effort, provider or sandbox Clawpatch uses; section 2 requires defaults)" % ", ".join(set_names))
    if not shutil.which(x2c.ARMS[a.arm]["provider"]):
        sys.exit("refusing to start: the %s CLI is not on PATH (arm %s needs it)" % (x2c.ARMS[a.arm]["provider"], a.arm))
    expect = a.expect_clawpatch_version
    out = sh([a.node, a.cli, "--version"], check=False, timeout=C.VERSION_TIMEOUT_SECONDS).stdout
    if not re.search(r"(?<![\w.])" + re.escape(expect) + r"(?![\w.])", out):
        sys.exit("refusing to start: Clawpatch --version %r is not the pinned %s" % (out.strip(), expect))
    chk = C.build_safety_check(a.cli)
    if not chk["ok"]:  # amendment 1: the unpatched local build the first codex/fastapi collection ran on is not what Scout invokes
        sys.exit("refusing to start: Clawpatch CLI %s is not the review-safety build (%s); rebuild with scripts/bootstrap-tools.sh" % (a.cli, chk["reason"]))
    if x2c.ARMS[a.arm]["provider"] == "claude":
        exe = C.claude_executable(os.environ)  # CLAWPATCH_CLAUDE_BIN is honoured by Clawpatch's claude provider (providers/claude.ts claudeExecutable)
        if not shutil.which(exe):
            sys.exit("refusing to start: the claude executable %r (CLAWPATCH_CLAUDE_BIN or `claude`) is not runnable" % exe)
        out = sh([exe, "--version"], check=False, timeout=C.VERSION_TIMEOUT_SECONDS).stdout
        ver = C.parse_version(out)
        if ver is None or ver < C.CLAUDE_MIN_VERSION:
            sys.exit("claude arm needs Claude Code >= %s for CLAWPATCH_CLAUDE_AUTH_CONTEXT=host; found %r" % (".".join(map(str, C.CLAUDE_MIN_VERSION)), out.strip()))
    bad = x2c.verify_pins()
    if bad:  # X2c reuses the frozen X2 files by import: any change to them voids the reuse
        sys.exit("refusing to start: frozen X2 files differ from their pinned digests: %s" % "; ".join(bad))


def probe_model(arm, grace=C.TERM_GRACE_SECONDS):
    """One trivial model call through the arm's CLI, to record the model AND reasoning effort it serves in the context of the map calls (Clawpatch only exposes its
    option; section 2 [INFERENCE]: the maps' model is taken to be the probes' model; a transient roll or a weight change under the same name is invisible). The call
    mirrors Clawpatch's provider invocation for the arm's options (providers/codex.ts:77-90,130-141, providers/claude.ts:170-212,234-275) so that user configuration,
    settings and environment resolve the model exactly as for the maps: codex `exec --cd <tmp> --sandbox read-only --skip-git-repo-check --model M -c model_reasoning_effort="E" -`
    under Clawpatch's own (process) environment, with NO --ignore-user-config/--ignore-rules; claude `-p --output-format json ... --safe-mode ... --model M [--effort E]` in
    claudeEnv(host)'s replaced, allow-listed environment, with NO --setting-sources, the executable CLAWPATCH_CLAUDE_BIN when set. Values: codex the `model:` and
    `reasoning effort:` banner lines (`reasoning_effort_source` "banner"); claude the sorted, comma-joined `modelUsage` keys and the REQUESTED effort ("requested": the
    claude CLI does not report the effort it served; for effort `none` no flag is passed and the recorded effort is `none`)."""
    t0 = time.time()
    env = C.clawpatch_map_env(os.environ)
    provider, want_effort = x2c.ARMS[arm]["provider"], x2c.ARMS[arm]["effort"]
    with tempfile.TemporaryDirectory() as d:
        if provider == "claude":
            cmd = [C.claude_executable(env), "-p", "--output-format", "json", "--tools", "", "--no-session-persistence", "--safe-mode", "--strict-mcp-config",
                   "--mcp-config", json.dumps({"mcpServers": {}}, separators=(",", ":")), "--disable-slash-commands", "--no-chrome", *x2c.claude_probe_tail(arm)]
            env = C.claude_provider_env(env, d)
        else:
            cmd = ["codex", *x2c.codex_probe_args(arm, d)]
        rc, out, err, timed_out, _sig = run_limited(cmd, env, C.PROBE_TIMEOUT_SECONDS, cwd=d, grace=grace, inp="Reply with the single word: ok")
    res = {"model": None, "reasoning_effort": None, "reasoning_effort_source": "requested" if provider == "claude" else "banner", "error": None,
           "seconds": round(time.time() - t0, 1), "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())}
    try:
        if rc != 0:
            res["error"] = "rc=%s%s" % (rc, " (timed out)" if timed_out else "")
        elif provider == "claude":
            res["model"] = ",".join(sorted(json.loads(out).get("modelUsage", {}))) or None
            res["reasoning_effort"] = want_effort
        else:
            m = re.search(r"^model:\s*(\S+)", out + err, re.M)
            e = re.search(r"^reasoning effort:\s*(\S+)", out + err, re.M)
            res["model"] = m.group(1) if m else None
            res["reasoning_effort"] = e.group(1) if e else None
    except ValueError as e:
        res["error"] = "unparsable probe output: %s" % e
    if res["model"] is None and res["error"] is None:
        res["error"] = "no model named in the probe output"
    if res["reasoning_effort"] is None and res["error"] is None:
        res["error"] = "no reasoning effort named in the probe output"
    return res


def cli_versions():
    v = {t: version_of([t, "--version"]) for t in ("claude", "codex") if shutil.which(t)}
    exe = C.claude_executable(os.environ)
    if exe != "claude" and shutil.which(exe):  # the maps (and the probe) run CLAWPATCH_CLAUDE_BIN, not the `claude` on PATH
        v["claude_executable_version"] = version_of([exe, "--version"])
    return v


def tool_versions(a, probe_before, ws):
    v = cli_versions()
    v["clawpatch"] = version_of([a.node, a.cli, "--version"])
    v["clawpatch_cli_sha256"] = C.sha256_file(a.cli)
    v["clawpatch_build_safety"] = C.build_safety_check(a.cli)
    v["node"] = version_of([a.node, "--version"])
    head = sh(["git", "rev-parse", "HEAD"], cwd=ROOT, check=False)
    env = C.clawpatch_map_env(os.environ)
    v.update({"scout_head": head.stdout.strip() or "unknown",
              "areas_go_sha256": C.sha256_file(os.path.join(ROOT, "internal", "areas", "areas.go")),
              "records_go_sha256": C.sha256_file(os.path.join(ROOT, "internal", "clawpatch", "records.go")),
              "env_names": sorted(env), "env_presence": {n: n in os.environ for n in C.RECORDED_PRESENCE_ENV},  # names only, never values
              "claude_auth_context": env.get("CLAWPATCH_CLAUDE_AUTH_CONTEXT"),
              "codex_timeout_ms": env.get("CLAWPATCH_CODEX_TIMEOUT_MS"), "claude_timeout_ms": env.get("CLAWPATCH_CLAUDE_TIMEOUT_MS"),
              "call_timeout_seconds": a.call_timeout, "wall_max_seconds": a.wall_max, "term_grace_seconds": a.term_grace,
              "workspace_config_presence": {p: os.path.exists(os.path.join(ws, p)) for p in C.WORKSPACE_CONFIG_PATHS},  # names only; the probes run outside the workspace
              "x2c": {"arm": a.arm, **x2c.ARMS[a.arm], "map_flags": x2c.arm_flags(a.arm), "collect_x2c_sha256": C.sha256_file(os.path.abspath(__file__)), "x2_pins": x2c.X2_PINS},
              "model_probe_before": probe_before, "model_probe_after": None, "cli_versions_after": None, "started": time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())})
    return v


def write_versions(a, v):
    """First write creates the file exclusively (append-only across invocations); this invocation then rewrites it once to add model_probe_after.
    `_created` is set as soon as the exclusive open succeeds, so an interrupt during the dump leaves a file the abort path rewrites, not one it trips over."""
    path = os.path.join(a.data, "tool-versions-%s-%s.json" % (a.arm, a.repo))
    with open(path, "w" if v.get("_created") else "x") as f:
        v["_created"] = True
        json.dump({k: x for k, x in v.items() if k != "_created"}, f, indent=1)


def _term(signum, _frame):
    raise KeyboardInterrupt("signal %d" % signum)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--arm", required=True, choices=list(x2c.ARMS))
    ap.add_argument("--repo", required=True)
    ap.add_argument("--source", required=True, help="path or URL of a clone containing the pinned revision")
    ap.add_argument("--work", default="/tmp/scout-x2c")
    ap.add_argument("--data", default=x2c.DATA)
    ap.add_argument("--node", default=shutil.which("node"))
    ap.add_argument("--cli", default=os.environ.get("SCOUT_CLAWPATCH_BINARY"))
    ap.add_argument("--spec", help="JSON overriding the pinned repository table (validation only)")
    ap.add_argument("--wall-max", type=float, default=C.WALL_MAX_SECONDS, help="seconds; preregistered default 8 h (validation overrides)")
    ap.add_argument("--call-timeout", type=float, default=C.CALL_TIMEOUT_SECONDS, help="seconds per Clawpatch call; preregistered default 1800")
    ap.add_argument("--term-grace", type=float, default=C.TERM_GRACE_SECONDS, help="seconds between SIGTERM and SIGKILL; preregistered default 10 (validation overrides)")
    ap.add_argument("--expect-clawpatch-version", default=C.CLAWPATCH_VERSION, help="pinned Clawpatch version; preregistered 0.7.3 (validation overrides for its stub)")
    a = ap.parse_args()
    t_start = time.time()
    # a dropped terminal (SIGHUP) or a kill ends the running process group like Ctrl-C, then writes the stop record. An inherited SIGHUP ignore (set by
    # `nohup`) wins: replacing it would turn the protection the README recommends into an abort.
    for sig in (signal.SIGTERM, signal.SIGHUP):
        if sig == signal.SIGHUP and signal.getsignal(sig) is signal.SIG_IGN:
            continue
        signal.signal(sig, _term)
    repos = json.loads(a.spec) if a.spec else C.REPOS
    spec = repos[a.repo]
    if not a.node or not a.cli:
        sys.exit("need --node and --cli (or SCOUT_CLAWPATCH_BINARY)")
    preflight(a)
    probe_before = probe_model(a.arm, a.term_grace)
    if probe_before["error"]:
        sys.exit("refusing to start: the %s CLI model probe failed: %s" % (a.arm, probe_before["error"]))
    base = os.path.join(a.work, a.arm, a.repo)  # the preflight has refused an existing `base`
    ws = os.path.join(base, "ws")
    os.makedirs(base)
    sh(["git", "clone", "-q", "--no-local", a.source, ws])
    sh(["git", "checkout", "-q", spec["rev"]], cwd=ws)
    if sh(["git", "rev-parse", "HEAD"], cwd=ws).stdout.strip() != spec["rev"]:
        sys.exit("pinned revision mismatch")
    sh(["git", "checkout", "-q", "-B", "x2"], cwd=ws)
    os.makedirs(a.data, exist_ok=True)
    tv = tool_versions(a, probe_before, ws)
    chain, ext = os.path.join(base, "state-chain"), spec["ext"]
    plan = [("F1", chain, True, None), ("F2", os.path.join(base, "state-F2"), True, None), ("F3", os.path.join(base, "state-F3"), True, None),
            ("R1", chain, False, None), ("R2", chain, False, None), ("M1", chain, False, lambda: mutate_m1(ws, ext)),
            ("M2", chain, False, lambda: mutate_m2(ws, ext, a.data, a.arm, a.repo))]
    stop_path = os.path.join(a.data, "stopped-%s-%s.json" % (a.arm, a.repo))

    def write_stop(reason, before, not_attempted, phase="maps"):
        with open(stop_path, "x") as f:
            json.dump({"arm": a.arm, "repo": a.repo, "reason": reason, "phase": phase, "elapsed_seconds": round(time.time() - t_start, 1),
                       "stopped_before": before, "not_attempted": not_attempted}, f, indent=1)

    ok, consecutive, stopped, current, phase = {}, 0, None, 0, "maps"
    try:
        write_versions(a, tv)  # first statement of the protected block: from here an abort writes the stop record (before it, nothing is recorded in data/)
        for current, (map_id, state, fresh, mutation) in enumerate(plan):
            elapsed = time.time() - t_start
            if elapsed > a.wall_max:
                stopped = "wall-clock limit %.0fs exceeded (%.0fs elapsed)" % (a.wall_max, elapsed)
            elif consecutive >= C.CONSECUTIVE_FAIL_MAX:
                stopped = "%d consecutive failed maps" % consecutive
            if stopped:
                write_stop(stopped, map_id, [p[0] for p in plan[current:]])
                print("[%s/%s] STOPPED before %s: %s" % (a.arm, a.repo, map_id, stopped), flush=True)
                break
            extra = mutation() if mutation else {}
            if extra.get("skipped"):  # no map is run for a skipped mutation (it is neither a success nor a failure for the stop rule)
                record_skipped(a, spec, ws, map_id, fresh, extra)
                ok[map_id] = None
                continue
            ok[map_id] = run_map(a, spec, ws, state, map_id, fresh, extra)
            consecutive = 0 if ok[map_id] else consecutive + 1
        phase = "after-probe"  # inside the protected block: an abort here still writes the stop record (all maps are already recorded)
        tv["model_probe_after"] = probe_model(a.arm, a.term_grace)
        tv["cli_versions_after"] = cli_versions()  # Claude Code self-updates: a version change inside the window is recorded, beside the after-probe
        write_versions(a, tv)
    except BaseException as e:  # Ctrl-C, SIGTERM/SIGHUP to the collector, any exception: run_limited has already ended the Clawpatch process group
        if not os.path.exists(stop_path):
            in_maps = phase == "maps"
            write_stop("aborted: %s: %s" % (type(e).__name__, e), plan[current][0] if in_maps else "after-probe", [p[0] for p in plan[current:]] if in_maps else [], phase)
        for k_ in ("model_probe_after", "cli_versions_after"):  # only what was not taken: a probe taken before the abort is kept
            if tv[k_] is None:
                tv[k_] = "not taken: aborted"
        write_versions(a, tv)
        raise
    print(json.dumps(ok))


if __name__ == "__main__":
    main()
