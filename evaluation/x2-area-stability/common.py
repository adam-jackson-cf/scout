"""X2 shared constants and helpers. Standard library only (run with .tools/venv/bin/python)."""
import glob
import hashlib
import json
import math
import os
import re
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
DATA = os.path.join(HERE, "data")
STALE_STATUS = "skipped"  # Clawpatch marks records the latest map did not emit (internal/clawpatch StaleStatus)
AGENT_SOURCE = "agent-mapper"  # internal/areas.AgentSource

# Preregistered (PREREGISTRATION.md section 4): grid in percent, step 5, so thresholds are exact p/100 floats.
GRID_PCT = list(range(30, 91, 5))
T_MIN = GRID_PCT[0] / 100.0
GRID = [p / 100.0 for p in GRID_PCT]

# Preregistered decision constants (sections 4, 5, 7, 9).
SURVIVAL_MIN = 0.80  # compared as the exact rational 4/5 (analysis.SURVIVAL_MIN_FRAC)
SPURIOUS_MAX = 0.05
SPURIOUS_UPPER_GUARD = 0.10
UNMEASURED_MAX = 0.10
CLASS_MIN_MEASURED = {"remap": 2, "mutation": 1}  # per repository, measured transitions of each decisive class
CONTROL_SWAP_FRACTION = 0.08  # content-swapped control pairs, as a fraction of unique real pairs (rounded up)
CONTROL_IDENT_FRACTION = 0.04  # identical-record control pairs
CONTROL_FAIL_MAX = 0.20  # label-validity gate (section 5)
WALL_MAX_SECONDS = 8 * 3600  # section 9: per (arm, repo) collect.py invocation
CALL_TIMEOUT_SECONDS = 1800  # section 9: per Clawpatch call
CONSECUTIVE_FAIL_MAX = 3  # section 9
CLAUDE_MIN_VERSION = (2, 1, 169)  # Clawpatch host auth context (third_party/clawpatch/docs/providers.md)
DEFAULT_MAP_TIMEOUT_MS = "900000"  # internal/clawpatch defaultMapTimeoutMS: default of BOTH CLAWPATCH_CODEX_TIMEOUT_MS and CLAWPATCH_CLAUDE_TIMEOUT_MS
TIMEOUT_ENV_NAMES = ("CLAWPATCH_CODEX_TIMEOUT_MS", "CLAWPATCH_CLAUDE_TIMEOUT_MS")  # internal/clawpatch mapEnvironment
CLAWPATCH_VERSION = "0.7.3"  # internal/clawpatch Version: the only Clawpatch the study accepts (section 2)
TERM_GRACE_SECONDS = 10  # section 9: SIGTERM, this grace, then SIGKILL
VERSION_TIMEOUT_SECONDS = 60  # section 9: bound on every `--version` call
PROBE_TIMEOUT_SECONDS = 300  # section 2: model probe before/after each (arm, repo) collection
CONTROL_MIN_USABLE = 10  # section 4 gate 3: minimum usable controls of each kind (pooled across arms), and at least half the generated ones
Z95 = 1.959963984540054

ARMS = ["codex", "claude"]  # codex = decision arm A; claude = replication arm B
REPOS = {
    "fastapi": {"rev": "d666ccb62216e45ca78643b52c235ba0d2c53986", "url": "https://github.com/fastapi/fastapi", "ext": ".py"},
    "axum": {"rev": "562fa206f09514e7503e898decf73024f5ce6fd4", "url": "https://github.com/tokio-rs/axum", "ext": ".rs"},
}
MAP_IDS = ["F1", "F2", "F3", "R1", "R2", "M1", "M2"]
# (transition id, class, sub-stratum, prev map, next map). Classes (all six transitions are decisive): remap = agent re-map of the
# unchanged tree (T1-T4), mutation = re-map after a structural change (T5-T6). Sub-strata (descriptive): fresh = F2/F3 on a new state
# directory, same_state = R1/R2 on the F1 state; they differ only in Clawpatch's stale-record bookkeeping (section 2).
TRANSITIONS = [
    ("T1", "remap", "fresh", "F1", "F2"),
    ("T2", "remap", "fresh", "F1", "F3"),
    ("T3", "remap", "same_state", "F1", "R1"),
    ("T4", "remap", "same_state", "R1", "R2"),
    ("T5", "mutation", "mutation", "R2", "M1"),
    ("T6", "mutation", "mutation", "M1", "M2"),
]
CHAIN_TRANSITIONS = ["T3", "T4", "T5", "T6"]  # F1 -> R1 -> R2 -> M1 -> M2, folded with accumulating areas (label-free secondary measure)
DECISIVE_CLASSES = ("remap", "mutation")
LABELS = ("same_concern", "overlapping_concern", "different_concern", "cannot_tell")
GUIDE_VERSION = "x2-concern-v1/guide-1"


# ---- Scout's credential scrub, ported name for name from internal/modelcli/modelcli.go (IsCredential, ScrubEnv). validation checks this
# port against the Go function on a name list (validation/credcheck).
CREDENTIAL_NAMES = {"ANTHROPIC_AUTH_TOKEN", "GITHUB_TOKEN", "GH_TOKEN", "INFISICAL_TOKEN", "ANTHROPIC_BASE_URL", "OPENAI_BASE_URL", "OPENAI_ORG_ID"}


def is_credential(name):
    upper = name.upper()
    return upper in CREDENTIAL_NAMES or upper.startswith("OPENROUTER_") or "API_KEY" in upper or "SECRET" in upper


def scrub_env(base, extra=()):
    """modelcli.ScrubEnv: base (a mapping) without credentials and without names overridden by extra, then extra ("K=V" strings) appended."""
    override = {e.split("=", 1)[0] for e in extra if "=" in e}
    out = {k: v for k, v in base.items() if not is_credential(k) and k not in override}
    for e in extra:
        k, _, v = e.partition("=")
        out[k] = v
    return out


def clawpatch_map_env(base):
    """Environment of one model-using Clawpatch call exactly as Scout builds it (internal/clawpatch/provider.go modelInvocation + client.go
    mapEnvironment + invokeEnvLocked scrub): host auth context, BOTH provider timeouts (Scout's own value wins, else 900000), no credentials."""
    extra = ["CLAWPATCH_CLAUDE_AUTH_CONTEXT=host"]
    for name in TIMEOUT_ENV_NAMES:
        extra.append(name + "=" + ((base.get(name) or "").strip() or DEFAULT_MAP_TIMEOUT_MS))
    return scrub_env(base, extra)


# ---- Clawpatch claude provider's environment, ported from third_party/clawpatch/src/providers/claude.ts (CLAUDE_AUTH_ENV_KEYS, claudeEnv with
# authContext "host", includeAuth true, `replaceEnv: true`). The model probe (collect.probe_model) runs in it so it resolves the model as the maps do.
CLAUDE_AUTH_ENV_KEYS = [
    "ANTHROPIC_API_KEY", "ANTHROPIC_BASE_URL", "ANTHROPIC_AUTH_TOKEN", "CLAUDE_CODE_USE_VERTEX", "ANTHROPIC_VERTEX_PROJECT_ID", "ANTHROPIC_VERTEX_REGION",
    "ANTHROPIC_VERTEX_BASE_URL", "CLOUD_ML_REGION", "GOOGLE_APPLICATION_CREDENTIALS", "GOOGLE_CLOUD_PROJECT", "GCLOUD_PROJECT", "CLOUDSDK_CORE_PROJECT",
    "CLAUDE_CODE_SKIP_VERTEX_AUTH", "CLAUDE_CODE_USE_BEDROCK", "CLAUDE_CODE_SKIP_BEDROCK_AUTH", "ANTHROPIC_BEDROCK_BASE_URL", "AWS_BEARER_TOKEN_BEDROCK",
    "AWS_REGION", "AWS_DEFAULT_REGION", "AWS_ACCESS_KEY_ID", "AWS_SECRET_ACCESS_KEY", "AWS_SESSION_TOKEN", "AWS_SHARED_CREDENTIALS_FILE", "AWS_CONFIG_FILE",
    "AWS_ROLE_ARN", "AWS_WEB_IDENTITY_TOKEN_FILE",
]
# Names that change which model Clawpatch's maps use (app.ts:243-256, config.ts:78-83,137-139: CLAWPATCH_CONFIG names a config file whose provider.model,
# provider.reasoningEffort and codexConfig become the provider options); `??` there means even an empty value overrides, so presence is refused.
REFUSED_MODEL_ENV = ("CLAWPATCH_MODEL", "CLAWPATCH_REASONING_EFFORT", "CLAWPATCH_PROVIDER", "CLAWPATCH_CONFIG")
# Names that change what a map call may do to the workspace (codex.ts:121-127: `bypass`/`none` -> --dangerously-bypass-approvals-and-sandbox, any other
# non-empty value -> --sandbox <value>). They do not change the model, but a writable map would break "re-map of the unchanged tree" (T1-T4); presence is refused.
REFUSED_SANDBOX_ENV = ("CLAWPATCH_CODEX_SANDBOX",)
# Workspace-level files that could change the model or settings of a map call but not of a probe (maps run with cwd = workspace, probes in a temp
# directory); `collect.tool_versions` records their presence (names only) after the clone.
WORKSPACE_CONFIG_PATHS = (".claude/settings.json", ".claude/settings.local.json", ".codex", "clawpatch.config.json", ".clawpatch/config.json")
# Names whose presence (never value) is recorded because they can change which executable or model a provider call resolves.
RECORDED_PRESENCE_ENV = ("CLAWPATCH_CLAUDE_BIN", "ANTHROPIC_MODEL", "CODEX_HOME")


def claude_provider_env(base, base_dir):
    """claudeEnv(includeAuth=true, baseDir, "host") over `base` (the environment Clawpatch itself receives, i.e. clawpatch_map_env): the replaced, allow-listed
    environment of a claude provider call. Values that are absent or empty are not copied (copyEnv)."""
    env = {}

    def copy(key):
        v = base.get(key)
        if v is not None and len(v) > 0:
            env[key] = v
    for key in base:  # copyPathEnv (claude.ts:304-311): the FIRST case-insensitive `path` key only, even when its value is empty and copies nothing
        if key.lower() == "path":
            copy(key)
            break
    for key in ("SystemRoot", "ComSpec", "PATHEXT"):
        copy(key)
    env["HOME"] = (base.get("HOME") or "").strip() or os.path.expanduser("~")
    copy("USERPROFILE")
    copy("CLAUDE_CONFIG_DIR")
    env["XDG_CONFIG_HOME"] = os.path.join(base_dir, "xdg-config")
    env["XDG_CACHE_HOME"] = os.path.join(base_dir, "xdg-cache")
    env["XDG_DATA_HOME"] = os.path.join(base_dir, "xdg-data")
    env["TMPDIR"] = env["TEMP"] = env["TMP"] = base_dir
    env["CLAUDE_CODE_SUBPROCESS_ENV_SCRUB"] = "1"
    for key in CLAUDE_AUTH_ENV_KEYS:
        copy(key)
    if base.get("AWS_CONFIG_FILE") is not None or base.get("AWS_SHARED_CREDENTIALS_FILE") is not None:
        copy("AWS_PROFILE")
    copy("CLAUDE_CODE_OAUTH_TOKEN")
    return env


def claude_executable(base):
    """claudeExecutable(): CLAWPATCH_CLAUDE_BIN (trimmed) when non-empty, else `claude`."""
    return (base.get("CLAWPATCH_CLAUDE_BIN") or "").strip() or "claude"


def parse_version(text):
    m = re.search(r"(\d+)\.(\d+)\.(\d+)", text or "")
    return tuple(int(x) for x in m.groups()) if m else None


def sha256_file(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def wilson(k, n, z=Z95):
    """Wilson score interval for k successes of n; (nan, nan) when n == 0."""
    if n == 0:
        return (float("nan"), float("nan"))
    p = k / n
    d = 1 + z * z / n
    c = p + z * z / (2 * n)
    r = z * math.sqrt(p * (1 - p) / n + z * z / (4 * n * n))
    return (max(0.0, (c - r) / d), min(1.0, (c + r) / d))


def map_dir(data_root, arm, repo, map_id):
    return os.path.join(data_root, "maps", arm, repo, map_id)


def read_meta(data_root, arm, repo, map_id):
    p = os.path.join(map_dir(data_root, arm, repo, map_id), "meta.json")
    if not os.path.exists(p):
        return None
    with open(p) as f:
        return json.load(f)


def load_active_features(data_root, arm, repo, map_id):
    """Active (non-stale) records of one recorded map, as the clawpatch boundary exposes them. None when the map was not recorded
    or failed (rc != 0)."""
    meta = read_meta(data_root, arm, repo, map_id)
    if meta is None or meta.get("rc") != 0:
        return None
    feats = []
    for p in sorted(glob.glob(os.path.join(map_dir(data_root, arm, repo, map_id), "features", "*.json"))):
        with open(p) as f:
            rec = json.load(f)
        if rec.get("status") != STALE_STATUS:
            feats.append(rec)
    return feats


def owned_files(rec):
    return sorted({o["path"] for o in rec.get("ownedFiles") or []})


def jaccard(a, b):
    a, b = set(a), set(b)
    u = len(a | b)
    return 0.0 if u == 0 else len(a & b) / u


def read_jsonl(path):
    """Rows of a JSONL file; a malformed line (e.g. the partial trailing line of a killed writer) is a loud error naming file and line."""
    out = []
    if os.path.exists(path):
        with open(path) as f:
            for n, line in enumerate(f, 1):
                if line.strip():
                    try:
                        out.append(json.loads(line))
                    except ValueError as e:
                        sys.exit("malformed JSONL line %s:%d (%s); recovery: PREREGISTRATION section 6" % (path, n, e))
    return out


def final_labels(pairs_dir, with_unparsable=False):
    """First usable `agent_reviewed` label per pair (labels are appended, never edited; later rounds only fill unparsable pairs).
    With `with_unparsable`, pairs whose only final rows are `unparsable` map to "unparsable"; pairs never labelled are absent."""
    out, unparsable = {}, set()
    for r in read_jsonl(os.path.join(pairs_dir, "labels.jsonl")):
        if r.get("provenance") != "agent_reviewed" or r.get("guide_version") != GUIDE_VERSION:
            continue
        if r["label"] in LABELS:
            out.setdefault(r["pair_id"], r["label"])
        elif r["label"] == "unparsable":
            unparsable.add(r["pair_id"])
    if with_unparsable:
        for pid in unparsable:
            out.setdefault(pid, "unparsable")
    return out


def write_transitions(path, arms=None, repos=None):
    rows = [{"arm": a, "repo": r, "id": t, "class": c, "sub": s, "prev": p, "next": n}
            for a in (arms or ARMS) for r in (repos or REPOS) for (t, c, s, p, n) in TRANSITIONS]
    with open(path, "w") as f:
        json.dump(rows, f, indent=1)
    return rows
