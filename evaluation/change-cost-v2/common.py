#!/usr/bin/env python3
"""E1b-v2 shared definitions: repositories, pins, eligibility, git helper. No analysis here."""
import json
import re
import subprocess
from pathlib import Path

HERE = Path(__file__).resolve().parent
REPOS = Path("/tmp/scout-evidence/repos")
WORK = Path("/tmp/scout-evidence/e1b-v2")
DAY = 86400
WINDOW_DAYS = 730          # PR merge window length ending at T_end
TRAIL_DAYS = 365           # degradation look-back
FOLLOW_DAYS = 14           # follow-up horizon for incompleteness
COCHANGE_DAYS = 365        # co-change history window for hidden-coupling exposure
BULK_FILES = 50            # PRs/commits touching more files are bulk and excluded
PIN_BEFORE = "2025-10-02T00:00:00Z"   # new repositories are pinned at the last first-parent commit before this

# name -> (GitHub slug, local dir, pin mode). "head" = the already-pinned clone HEAD (reference-v1);
# "date" = last first-parent commit strictly before PIN_BEFORE on the default branch of the fresh clone.
REPO_CFG = {
    "fastapi": ("fastapi/fastapi", "fastapi", "head"),
    "webpack": ("webpack/webpack", "webpack", "head"),
    "nest": ("nestjs/nest", "nest", "head"),
    "caddy": ("caddyserver/caddy", "caddy", "head"),
    "axum": ("tokio-rs/axum", "axum", "head"),
    "junit5": ("junit-team/junit-framework", "junit5", "head"),
    "eShopOnWeb": ("dotnet-architecture/eShopOnWeb", "eShopOnWeb", "head"),
    "django": ("django/django", "django", "date"),
    "pandas": ("pandas-dev/pandas", "pandas", "date"),
    "scikit-learn": ("scikit-learn/scikit-learn", "sklearn", "date"),
    "vite": ("vitejs/vite", "vite", "date"),
    "fastify": ("fastify/fastify", "fastify", "date"),
    "prometheus": ("prometheus/prometheus", "prometheus", "date"),
    "terraform": ("hashicorp/terraform", "terraform", "date"),
    "jellyfin": ("jellyfin/jellyfin", "jellyfin", "date"),
    "tokio": ("tokio-rs/tokio", "tokio", "date"),
}
NAMES = list(REPO_CFG)
EXTS = {".py", ".js", ".mjs", ".cjs", ".ts", ".tsx", ".go", ".rs", ".java", ".cs"}

# One universal production-source rule for every repository (no per-repository tuning).
_DIR_RE = re.compile(
    r"(^|/)(tests?|__tests__|testing|testdata|test-data|spec|specs|e2e|fixtures?|examples?|samples?|docs?|"
    r"documentation|benchmarks?|bench|vendor|third_party|node_modules|dist|build|generated|mocks?|"
    r"scripts?|tools?|bin|migrations|playground|integration-tests?)/", re.I)
_FILE_RE = re.compile(
    r"(_test\.(go|py)|(^|/)test_[^/]*\.py|(^|/)conftest\.py|\.(test|spec|stories)\.(ts|tsx|js|mjs|cjs)|"
    r"(Test|Tests|Spec)\.(java|cs)|\.d\.ts|\.pb\.go|(_|\.)generated\.[a-z]+|\.min\.js|(^|/)setup\.py|\.mock\.[a-z]+)$", re.I)


def eligible(path):
    p = Path(path)
    return p.suffix in EXTS and not _DIR_RE.search(path) and not _FILE_RE.search(path)


def git(repo, *args, check=True):
    r = subprocess.run(["git", "-c", "core.quotepath=off", "-C", str(REPOS / REPO_CFG[repo][1]), *args],
                       capture_output=True, text=True, errors="replace")
    if check and r.returncode != 0:
        raise RuntimeError(f"git {' '.join(args)[:100]} failed: {r.stderr[:200]}")
    return r.stdout


def pin(repo):
    """-> (sha, committer_ts). Deterministic given the clone."""
    mode = REPO_CFG[repo][2]
    if mode == "head":
        sha = git(repo, "rev-parse", "HEAD").strip()
    else:
        sha = git(repo, "rev-list", "-1", "--first-parent", f"--before={PIN_BEFORE}", "HEAD").strip()
    ts = int(git(repo, "log", "-1", "--format=%ct", sha))
    return sha, ts


def load_pins():
    return json.loads((HERE / "pins.json").read_text())


# One bot predicate for every script (GraphQL `__typename == "Bot"` or exact login in this list; never substring matching).
BOT_LOGINS = {"dependabot", "renovate", "github-actions", "pre-commit-ci", "allcontributors", "greenkeeper",
              "snyk-bot", "mergify", "codecov", "copilot", "imgbot", "web-flow"}


def is_bot(login, author_type):
    return author_type == "Bot" or (login or "").lower() in BOT_LOGINS
