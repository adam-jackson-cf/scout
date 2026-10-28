# Sourced by scripts/bootstrap-tools.sh, scripts/build-clawpatch.sh and scripts/install-release.sh (never executed directly).
# Defines `sha256 FILE`, which prints the file's SHA-256 with shasum or sha256sum, whichever the
# host has. The sourcing script must define `die`.
if command -v shasum >/dev/null 2>&1; then
  sha256() { shasum -a 256 "$1" | cut -d ' ' -f1; }
elif command -v sha256sum >/dev/null 2>&1; then
  sha256() { sha256sum "$1" | cut -d ' ' -f1; }
else
  die "required command is unavailable: shasum or sha256sum"
fi
