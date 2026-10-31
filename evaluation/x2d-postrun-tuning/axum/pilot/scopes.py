"""X2d onboarding scopes (PREREGISTRATION section 4): deterministic partition of the C0 tree into Clawpatch `include` scopes, each small enough that the agent
mapper's inventory sample covers it completely.

Clawpatch hands the agent only `files[:400]`, `sourceFiles[:500]` and `testFiles[:200]` of the include-filtered inventory (agent-mapper.ts:396-398), so a scope is
complete iff files <= 400, source files <= 500 and test files <= 200. The K0 check at C0 showed the council's three scopes do not fit (docs_src + tests alone is 1,224
files, 685 source files, 537 test files). The partition below is therefore derived, not chosen:

  core     fastapi/** + every test outside tests/test_tutorial (tests of the library itself)
  tutorial topic shards: docs_src/<topic> together with tests/test_tutorial/test_<topic>, topics sorted by name, packed greedily (a topic is never split)
  infra    .github, scripts, root files that are not markdown, docs-config yml named in INFRA_DOCS

These are the onboarding inventory shards (PREREGISTRATION 4.1), NOT the rename domains of rule (c): those are `frontend.scope` = core (`fastapi/**`), examples-tests
(`docs_src/**` + `tests/**`), infra-config (everything else) (R1-21).
Include entries are exact directory/file prefixes (Clawpatch matches `path == p or path startswith p + "/"` for a pattern without wildcards, mappers/shared.ts:77-79).
Standard library only. Usage: scopes.py --repo R --rev C0 --out scopes.json
"""
import argparse, json, os, re, sys

sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
import frontend as F
import prepass as P

CAP_FILES, CAP_SOURCE, CAP_TESTS = 400, 500, 200
INFRA_DOCS = ("docs/en/mkdocs.yml", "docs/language_names.yml")


def counts(paths):
    files, src, tests = P.inventory(paths)
    return {"files": len(files), "source": len(src), "tests": len(tests)}


def fits(c):
    return c["files"] <= CAP_FILES and c["source"] <= CAP_SOURCE and c["tests"] <= CAP_TESTS


def build(tree):
    """AXUM ADAPTATION: the C0 inventory partitioned into four scopes that fit the caps: `libs` (axum, axum-core, axum-extra), `macros` (axum-macros: its 159 trybuild tests),
    `examples` (examples/) and `infra` (.github and the root files that are not markdown). The fastapi tutorial-topic sharding does not apply."""
    tree = sorted(tree, key=lambda s: s.encode("utf-8", "surrogateescape"))
    inv = set(P.inventory(tree)[0])
    tree = [p for p in tree if p in inv]
    libs = [p for p in tree if p.startswith(("axum/", "axum-core/", "axum-extra/"))]
    macros = [p for p in tree if p.startswith("axum-macros/")]
    examples = [p for p in tree if p.startswith("examples/")]
    infra = [p for p in tree if p.startswith(".github/") or ("/" not in p and not p.lower().endswith(".md"))]
    scopes = [{"id": "libs", "include": ["axum", "axum-core", "axum-extra"], "paths": libs},
              {"id": "macros", "include": ["axum-macros"], "paths": macros},
              {"id": "examples", "include": ["examples"], "paths": examples},
              {"id": "infra", "include": sorted({p.split("/")[0] if p.startswith(".github/") else p for p in infra}), "paths": infra}]
    for s in scopes:
        s.update(counts(s["paths"]))
        s["fits"] = fits(s)
        s["n_paths"] = len(s["paths"])
        s["include_n"] = len(s["include"])
    covered = {p for s in scopes for p in s["paths"]}
    return scopes, [], [p for p in tree if p not in covered]



def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--repo", required=True)
    ap.add_argument("--rev", required=True)
    ap.add_argument("--out", required=True)
    a = ap.parse_args()
    scopes, leftovers, uncovered = build(F.tree_paths(a.repo, a.rev))
    out = {"rev": a.rev, "caps": [CAP_FILES, CAP_SOURCE, CAP_TESTS], "scopes": [{k: v for k, v in s.items() if k != "paths"} for s in scopes],
           "n_scopes": len(scopes), "all_fit": all(s["fits"] for s in scopes), "tutorial_leftover_paths": leftovers,
           "unscoped_inventory_paths": len(uncovered), "unscoped_by_top_dir": {}}
    for p in uncovered:
        k = p.split("/")[0]
        out["unscoped_by_top_dir"][k] = out["unscoped_by_top_dir"].get(k, 0) + 1
    json.dump(out, open(a.out, "w"), indent=1)
    print(json.dumps({k: v for k, v in out.items() if k != "scopes"}, indent=1))
    for s in out["scopes"]:
        print(s["id"], s["files"], s["source"], s["tests"], s["fits"], "include entries", s["include_n"], s.get("topics", ""))


if __name__ == "__main__":
    main()
