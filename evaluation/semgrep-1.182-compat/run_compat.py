"""Run bounded scanner compatibility samples; not a Scout design experiment."""
import json
import os
from pathlib import Path
import subprocess
import tempfile
import time

ROOT = Path(__file__).resolve().parents[2]
RAW = Path(__file__).resolve().parent / "raw"
SAMPLES = {
    "unsafe.py": "import yaml\nx = yaml.unsafe_load(data)\n",
    "safe.py": "import yaml\nx = yaml.safe_load(data)\n",
    ".hidden/unsafe.py": "import yaml\nx = yaml.unsafe_load(data)\n",
    "ignored/unsafe.py": "import yaml\nx = yaml.unsafe_load(data)\n",
    "vendor/unsafe.py": "import yaml\nx = yaml.unsafe_load(data)\n",
    ".github/workflows/mutable.yml": "jobs:\n  a:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n",
    "vuln/pyproject.toml": '[project]\ndependencies = ["python-multipart<0.0.7"]\n',
    "identity.css": ".menu {\n transition: height 0.3s ease;\n}\n",
    "AutoPublicPathRuntimeModule.js": 'const x = Template.indent([\n `if (document.currentScript)`,\n Template.indent(`scriptUrl = document.currentScript.src;`)\n]);\n',
}


def main():
    summaries = {}
    with tempfile.TemporaryDirectory(prefix="scout-semgrep-compat-") as tmp:
        work = Path(tmp)
        subprocess.run(["git", "init", "--quiet", tmp], check=True)
        for rel, text in {**SAMPLES, ".gitignore": "ignored/\n", ".semgrepignore": "vendor/\n"}.items():
            path = work / rel
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(text)
        for label, venv in [("before", "semgrep-baseline"), ("after", "venv")]:
            bin_dir = ROOT / ".tools" / venv / "bin"
            env = {**os.environ, "PATH": str(bin_dir) + os.pathsep + os.environ["PATH"], "SEMGREP_SEND_METRICS": "off"}
            args = [str(bin_dir / "semgrep"), "scan", "--config", str(ROOT / "internal/analyzers/rules/security"), "--config", str(ROOT / "internal/analyzers/rules/performance"), "--disable-version-check", "--metrics=off", "--no-git-ignore", "--json", "--error", *SAMPLES]
            start = time.monotonic()
            result = subprocess.run(args, cwd=work, env=env, capture_output=True)
            elapsed = time.monotonic() - start
            (RAW / f"sample-{label}.json").write_bytes(result.stdout)
            (RAW / f"sample-{label}.stderr.txt").write_bytes(result.stderr)
            out = json.loads(result.stdout)
            summaries[label] = {"version": out["version"], "exit": result.returncode, "seconds": elapsed, "scanned": sorted(out["paths"]["scanned"]), "errors": out["errors"], "findings": sorted((r["path"], r["check_id"], r["start"]["line"], r["end"]["line"], r["extra"]["message"], r["extra"]["severity"]) for r in out["results"]), "command": args}
        summaries["same_findings"] = summaries["before"]["findings"] == summaries["after"]["findings"]
        summaries["same_targets"] = summaries["before"]["scanned"] == summaries["after"]["scanned"]
        (RAW / "sample-summary.json").write_text(json.dumps(summaries, indent=2) + "\n")
    print(json.dumps(summaries, indent=2))


if __name__ == "__main__":
    main()
