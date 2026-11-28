# Preparation evidence

The preparation files in `preparation-run/` (72 files, 240 MB uncompressed) are too large to keep in git. They are stored as a release asset, and `SHA256SUMS` in this directory records the SHA-256 of every file.

| Release | Asset | Size | Asset SHA-256 |
|---|---|---|---|
| [`evidence-x1-preparation-2026-11-28`](https://github.com/adam-jackson-cf/scout/releases/tag/evidence-x1-preparation-2026-11-28) | `x1-preparation-run.tar.gz` | 28010722 bytes | `6b1d32b57db56320edc2889eb382e874a302f6f32d36b07916c374306d3d6bc1` |

Restore and verify from the repository root:

```sh
cd evaluation/x1-triage-ranking
gh release download evidence-x1-preparation-2026-11-28 -R adam-jackson-cf/scout -p x1-preparation-run.tar.gz
shasum -a 256 x1-preparation-run.tar.gz   # must match the asset SHA-256 above
tar -xzf x1-preparation-run.tar.gz
shasum -a 256 -c --quiet evidence/SHA256SUMS
```

The last command prints nothing and exits 0 when all 72 files match. The extracted paths are ignored by git.
