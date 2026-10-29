"""X2c shared configuration: the candidate mapper arms, the frozen-X2 reuse pins, the Jev pin and the drift guards.

X2c reuses the signed-off X2 code BY IMPORT (not by copy), exactly as X2b did (`../x2b-area-stability-luna/x2b.py`): `analysis.py`, `build_pairs.py`, `label.py`,
`common.py`, `replay/main.go`, `GUIDE.md`, the credential check and the stub Clawpatch are the frozen files of `evaluation/x2-area-stability/`; `verify_pins()` refuses
unless each has the exact SHA-256 X2's sign-off (`../x2-area-stability/reviews/round-10.md`) names. The copied-and-changed file is the collector (`collect_x2c.py`;
provenance: a pinned diff against X2's `collect.py`, `PINNED_COLLECT_BASE` and validation). Standard library only (run with .tools/venv/bin/python).
"""
import hashlib
import os
import re
import sys

HERE = os.path.dirname(os.path.abspath(__file__))
ROOT = os.path.abspath(os.path.join(HERE, "..", ".."))
X2_DIR = os.path.join(ROOT, "evaluation", "x2-area-stability")
DATA = os.path.join(HERE, "data")

# Frozen X2 files X2c imports or runs unchanged: path (relative to the X2 directory) -> SHA-256 named by X2's council round 10 (the same pins X2b uses).
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
# X2's collector, the base `collect_x2c.py` was copied from (the diff between the two is pinned in validation).
PINNED_COLLECT_BASE = "89948f3c6d1b308b91a5ac5a8060088269270806546fe5c577dea307a4dfc83d"

# ---- The mapper arms (PREREGISTRATION section 2). Clawpatch is invoked directly with these three values: `--provider P --model M --reasoning-effort E`.
# Order = the reporting order; `luna-xhigh` is the reference arm (Scout's production configuration, design rev 2.7 section 4).
ARMS = {
    "luna-xhigh": {"provider": "codex", "model": "gpt-6-luna", "effort": "xhigh"},
    "terra-none": {"provider": "codex", "model": "gpt-6.1-terra", "effort": "none"},
    "terra-low": {"provider": "codex", "model": "gpt-6.1-terra", "effort": "low"},
    "haiku": {"provider": "claude", "model": "claude-haiku-5-5", "effort": "none"},
}
REFERENCE_ARM = "luna-xhigh"
# Collection order (sequential, never parallel): serpentine so that no arm is always earlier in time (fastapi forward, axum reversed).
COLLECTION_ORDER = [("luna-xhigh", "fastapi"), ("terra-none", "fastapi"), ("terra-low", "fastapi"), ("haiku", "fastapi"),
                    ("haiku", "axum"), ("terra-low", "axum"), ("terra-none", "axum"), ("luna-xhigh", "axum")]
METHODS = ("jaccard", "jev")
JEV_SUFFIX = "+jev"  # the replay arm name of the Jev method on a mapper arm: "<arm>+jev"

# ---- Jev (PREREGISTRATION section 6): the pin is Scout's (internal/policy.DefaultModel), the client is internal/decisions.
JEV_MODEL = "typesafe/jev-1.14-20261014"
JEV_ENDPOINT = "https://openrouter.ai/api/alpha/decisions"  # internal/decisions.DefaultEndpoint
GO_POLICY = os.path.join(ROOT, "internal", "policy", "consent.go")
GO_DECISIONS_DIR = os.path.join(ROOT, "internal", "decisions")
JEV_SCORER_PKG = "./evaluation/x2c-mapper-comparison/jev"
# The version key of the X2c question for the pin (`jev -print-version`): any edit to the instructions, criteria, state shape, option-order rule or model changes it.
JEV_QUESTION_VERSION = "x2c.concern@2ff0e36e9db56d928d5b9beac8cb642b111e1ae805b5e2bea4133ec9c98bf1b2"
JEV_WORKERS = 8  # concurrent requests ASSUMED for a product Jev matcher (decisions.Run is sequential today, run.go:186); the study's driver uses 8 and the divisor of the per-refresh latency estimate

# ---- Preregistered decision constants of X2c (PREREGISTRATION section 7); X2's C1-C3, gates and grid are X2's (common.py, analysis.py).
COVERAGE_FLOOR = 0.5  # mapper guard (PREREGISTRATION section 7, option B): mean owned files per map >= this fraction of the reference arm's, in each repository; the area ratio is reported, not gated
JEV_UNSCORED_MAX = 0.01  # Jev gate: share of candidate occurrences without a score (a failed request in either order)
JEV_COST_PER_REFRESH_MAX_USD = 0.10  # Jev gate [ASSUMPTION]: billed cost of scoring one refresh's candidates (mean over transitions, both repositories summed)
JEV_SECONDS_PER_REFRESH_MAX = 300.0  # Jev gate [ASSUMPTION]: sum of request latencies / JEV_WORKERS per refresh (same aggregation)
TIE_BAND = 0.10  # cells within 10 % of the fastest eligible refresh time are tied


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


def arm_flags(arm):
    """The argv tail the collector appends after `map --source agent --provider <provider>`: Clawpatch's own flags, explicit (never from user configuration)."""
    a = ARMS[arm]
    return ["--model", a["model"], "--reasoning-effort", a["effort"]]


def codex_probe_args(arm, workdir):
    """The model probe's codex arguments: the provider invocation Clawpatch's codex provider makes for the arm's options (providers/codex.ts runCodexJson +
    addCodexModelArgs: sandbox, --skip-git-repo-check, no codexConfig, `--model`, `-c model_reasoning_effort="<effort>"`, `-`), without the schema and last-message outputs."""
    a = ARMS[arm]
    return ["exec", "--cd", workdir, "--sandbox", "read-only", "--skip-git-repo-check", "--model", a["model"], "-c", 'model_reasoning_effort="%s"' % a["effort"], "-"]


def claude_effort_flag(effort):
    """providers/claude.ts addClaudeModelArgs/claudeEffort: `none` adds no flag, `minimal` is passed as `low`, any other value as is."""
    if effort == "none":
        return []
    return ["--effort", "low" if effort == "minimal" else effort]


def claude_probe_tail(arm):
    """The model/effort arguments Clawpatch's claude provider appends (addClaudeModelArgs: `--model <m>` then the effort flag)."""
    a = ARMS[arm]
    return ["--model", a["model"], *claude_effort_flag(a["effort"])]


def go_jev_pin(policy=None):
    """The Jev pin Scout's source states (`DefaultModel = "..."` in internal/policy/consent.go), None when absent."""
    m = re.search(r'^\s*DefaultModel\s*=\s*"([^"]*)"', open(policy or GO_POLICY).read(), re.M)
    return m.group(1) if m else None


def check_jev_pin(policy=None):
    """-> reasons the Go source does not state the preregistered Jev pin (empty when it does)."""
    got = go_jev_pin(policy)
    return [] if got == JEV_MODEL else ["internal/policy/consent.go DefaultModel is %r, preregistered %r" % (got, JEV_MODEL)]


def go_code_digests():
    """SHA-256 of the Scout code under test that the Jev stage and the replay run: every non-test file of internal/decisions (the whole package compiles into the Jev driver,
    Redact included), internal/policy/consent.go, internal/areas/areas.go, internal/clawpatch/records.go, go.mod and go.sum. `run.sh jev` appends it to <data>/jev/code-digests.jsonl
    before every driver run; the analysis reports those lines."""
    files = sorted(os.path.join(GO_DECISIONS_DIR, n) for n in os.listdir(GO_DECISIONS_DIR) if n.endswith(".go") and not n.endswith("_test.go")) + [GO_POLICY,
             os.path.join(ROOT, "internal", "areas", "areas.go"), os.path.join(ROOT, "internal", "clawpatch", "records.go"), os.path.join(ROOT, "go.mod"), os.path.join(ROOT, "go.sum")]
    return {os.path.relpath(p, ROOT): sha256_file(p) for p in files}


# Reuse by import: the X2 modules (common, analysis, build_pairs, label) resolve from the frozen X2 directory.
if X2_DIR not in sys.path:
    sys.path.insert(0, X2_DIR)
