"""X2b shared configuration: the pinned mapping configuration, the frozen-X2 reuse pins and the drift guard against the Go constants.

X2b reuses the signed-off X2 code BY IMPORT (not by copy): `analysis.py`, `build_pairs.py`, `label.py`, `common.py`, `replay/main.go`, `GUIDE.md`, the
credential check and the stub Clawpatch are the frozen files of `evaluation/x2-area-stability/`; `verify_pins()` refuses unless each has the exact SHA-256 the X2
sign-off (`../x2-area-stability/reviews/round-10.md`) names. The only copied-and-changed file is the collector (`collect_x2b.py`; its provenance is a pinned
diff against X2's `collect.py`, see `PINNED_COLLECT_BASE` and validation). Standard library only (run with .tools/venv/bin/python).
"""
import hashlib
import os
import re
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", ".."))
X2_DIR = os.path.join(ROOT, "evaluation", "x2-area-stability")
DATA = os.path.join(HERE, "data")

# Frozen X2 files X2b imports or runs unchanged: path (relative to the X2 directory) -> SHA-256 named by X2's council round 10 (and unchanged since).
X2_PINS = {
    "PREREGISTRATION.md": "ed0a5203648b4da26d87effc4ccf207b9e8923311503571ea4b7bb4e1340e355",
    "GUIDE.md": "9a26286545c6c124ffc54d390af9b7d5a3e1f729ce01b1cb0d7266d6c251f799",
    "common.py": "e526804c331f3eab07f7003e797b86b4e793c3cf0a351e69cdf695286f83efa3",
    "analysis.py": "7d70e6304d9771bfc82669906aaf1e8fb9f79a7eb039d9c2ae37f0e7725fed63",
    "build_pairs.py": "face9e5b0cefe6c2fcf8fb9a3a568d28227ca0b57f85c0eeda9b562b251d222a",
    "label.py": "87b278014f75c215ffd5fe497572427a641b362fd82fe1ea217dfb152956800f",
    "replay/main.go": "f3614d6f6d19cf68d203ac46caadf18a9595204ba2f2d296154472285a971f7e",
    "validation/validate.py": "29e5208646d9a84804a7bd210cb07ad589d7acad48a1b85f31722076c2712784",
    "validation/stub_clawpatch.js": "0f1e9d8f751921cb64e17d1000845841e35510b9aab2320844b41dfa6c4e6226",
    "validation/credcheck/main.go": "e50633605ea763e2c4cb24ad3783cb58c356f71be4018961195d52807c4d5779",
}
# X2's collector, the base `collect_x2b.py` was copied from (the diff between the two is pinned in validation).
PINNED_COLLECT_BASE = "89948f3c6d1b308b91a5ac5a8060088269270806546fe5c577dea307a4dfc83d"
# X2's own result file, read only by the decision-combination block of analysis_x2b.py (a frozen artifact of a signed-off, run study).
X2_RESULTS_JSON = "data/results.json"
X2_RESULTS_JSON_SHA256 = "a9806fe3d0d9467372129e2ecb1dde092944ad5d373aa4d3e62803ff7cc030a3"

# ---- The pinned mapping configuration (PREREGISTRATION section 2). Scout runs the agent map on codex with exactly these two values
# (internal/clawpatch/client.go: mapModel, mapReasoningEffort). `map_flags()` refuses when the Go constants differ.
ARMS = ("codex",)
MAP_PROVIDER = "codex"
MAP_MODEL = "gpt-6-luna"
MAP_REASONING_EFFORT = "xhigh"
GO_CLIENT = os.path.join(ROOT, "internal", "clawpatch", "client.go")
GO_PROVIDER = os.path.join(ROOT, "internal", "clawpatch", "provider.go")
# What Scout appends after `map --source agent --provider codex` (client.go Map): the argv tail the collector passes, derived from the Go source.
ARGV_LINE = '"--model", mapModel, "--reasoning-effort", mapReasoningEffort'


def sha256_file(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1 << 20), b""):
            h.update(chunk)
    return h.hexdigest()


def verify_pins(x2_dir=None):
    """-> list of mismatch descriptions (empty when every reused X2 file has its pinned digest)."""
    bad = []
    for rel, want in sorted(X2_PINS.items()):
        p = os.path.join(x2_dir or X2_DIR, rel)
        got = sha256_file(p) if os.path.exists(p) else None
        if got != want:
            bad.append("%s: %s != pinned %s" % (rel, got, want))
    return bad


def go_map_config(client=None, provider=None):
    """The mapping configuration as Scout's Go source states it: {model, reasoning_effort, argv_tail_present, provider_resolver_present, provider}.
    Parsed from the source text (a drift guard, not a build): `mapModel`/`mapReasoningEffort` constants, the argv line that appends them after
    `--source agent` + the provider arguments, and resolveMapProvider returning ProviderCodex."""
    src = open(client or GO_CLIENT).read()
    m = re.search(r'^\s*mapModel\s*=\s*"([^"]*)"', src, re.M)
    e = re.search(r'^\s*mapReasoningEffort\s*=\s*"([^"]*)"', src, re.M)
    psrc = open(provider or GO_PROVIDER).read()
    fn = re.search(r"func \(c \*Client\) resolveMapProvider\(\) \(string, error\) \{.*?\n\}\n", psrc, re.S)
    return {"model": m.group(1) if m else None, "reasoning_effort": e.group(1) if e else None,
            "argv_tail_present": ('agentArgs = append(agentArgs, %s)' % ARGV_LINE) in src and "c.resolveMapProvider()" in src,
            "provider_resolver_present": bool(fn) and "return ProviderCodex, nil" in fn.group(0) and "modelcli.Detect(ProviderCodex)" in fn.group(0)}


def check_go_config(client=None, provider=None):
    """-> list of reasons the Go source does not state the preregistered configuration (empty when it does)."""
    g = go_map_config(client, provider)
    bad = []
    if g["model"] != MAP_MODEL:
        bad.append("client.go mapModel is %r, preregistered %r" % (g["model"], MAP_MODEL))
    if g["reasoning_effort"] != MAP_REASONING_EFFORT:
        bad.append("client.go mapReasoningEffort is %r, preregistered %r" % (g["reasoning_effort"], MAP_REASONING_EFFORT))
    if not g["argv_tail_present"]:
        bad.append("client.go no longer appends `%s` to the agent map argv via resolveMapProvider" % ARGV_LINE)
    if not g["provider_resolver_present"]:
        bad.append("provider.go resolveMapProvider no longer returns codex only")
    return bad


def map_flags():
    """The argv tail the collector appends after `map --source agent --provider codex`, exactly as Scout does. SystemExit unless the Go source states
    the preregistered configuration."""
    bad = check_go_config()
    if bad:
        sys.exit("refusing to start: Scout's mapping configuration drifted from the preregistration: " + "; ".join(bad))
    return ["--model", MAP_MODEL, "--reasoning-effort", MAP_REASONING_EFFORT]


def codex_probe_args(workdir):
    """The model probe's codex arguments: the provider invocation Clawpatch's codex provider makes for these options (providers/codex.ts runCodexJson +
    addCodexModelArgs: sandbox, --skip-git-repo-check, no codexConfig, `--model`, `-c model_reasoning_effort="<effort>"`, `-`), without the schema and
    last-message outputs."""
    return ["exec", "--cd", workdir, "--sandbox", "read-only", "--skip-git-repo-check", "--model", MAP_MODEL, "-c", 'model_reasoning_effort="%s"' % MAP_REASONING_EFFORT, "-"]


# Reuse by import: the X2 modules (common, analysis, build_pairs, label) resolve from the frozen X2 directory.
if X2_DIR not in sys.path:
    sys.path.insert(0, X2_DIR)
