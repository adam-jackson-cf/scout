import subprocess


def list_files() -> str:
    return subprocess.run(["ls"], capture_output=True, text=True, check=True, timeout=30).stdout
