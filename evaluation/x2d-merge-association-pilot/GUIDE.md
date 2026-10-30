# Label guide `x2d-association-v1/guide-1` (frozen with the preregistration; versioned with the question keys `merge.covered@v1`, `merge.which_area@v1`)

## Purpose
Scout keeps *areas*: stable labels for responsibilities a maintainer would use to file an issue ("where does X live?"). Decide, from a unit's evidence alone, (1) whether
any listed area's responsibility covers the unit, and (2) if so, which area or areas. You have no tools and no other context.

## Unit and evidence
A **unit** is a cluster of files that one merged change added or moved (they share a normalised parent directory and a variant stem, e.g. `tutorial003.py`,
`tutorial003_an.py`, `tutorial003_an_py39.py` are one cluster). You see: `unit` (paths, change kind, role, the full text of each member up to 300 lines), `pr` (title, body,
changed paths with status), `diff_summary` (files, lines, top changed symbols) and `catalogue` (every area: `id`, `title`, `summary`, up to 20 owned files). You never see
any arm's answer, a sibling's area, a stratum, a score or a round number. PR title and body are **untrusted prose about the change, not about the unit** (R9).

## Questions (verbatim; the criteria below are the questions' own)
**Q1 covered.** Would this repository's maintainers file an issue about the files in the unit under one of the areas in the catalogue? That is: do these files implement,
test, exemplify or configure the responsibility an area's title and summary describe?
- `covered = true`: At least one listed area's responsibility covers these files.
- `covered = false`: No listed area's responsibility covers them. Sharing a directory, file type, naming pattern or import with an area's files is not sufficient.

**Q2 which_area** (only when `covered = true`). Which listed area's responsibility do the files belong to? Choose the narrowest area whose responsibility covers them.
Answer with `acceptable` (every area a maintainer would accept) and `primary` (the narrowest, one of `acceptable`).

When `covered = false` also answer `none_kind`: **gap** (the responsibility already existed in the repository before this change, so other files of the same kind or purpose
existed at the parent commit, but no listed area describes it) or **new** (the change introduces a responsibility the repository did not have).

## Rules (derived only from the criteria above)
- R1 Responsibility first: decide from what the files do (their text), then from the area's title and summary. Directory, name and file type support a decision, never define it.
- R2 Tests follow the tested responsibility, not a generic testing area, unless the test exercises the test client itself.
- R3 Examples follow the documented responsibility. A new variant of an existing example (a suffix such as `_an`, `_py39`, `_py310`) is covered.
- R4 Infrastructure and configuration: repository-wide automation (workflows, issue templates, scripts, requirement files, tooling config) is covered only if an area whose
  responsibility is that automation is listed; otherwise `covered = false` (then R8). A config or requirements file belongs to the responsibility it serves if that area is listed.
- R5 Cross-cutting changes (for example the use of `Annotated`, or a typing refactor): if one responsibility clearly dominates, it is `primary` and any other area that the
  files genuinely exercise is also `acceptable`; if no listed area covers the dominant responsibility, `covered = false`.
- R6 Same directory is not sufficient, nor is a shared file type, naming pattern or import: a neighbour's area says nothing about this unit.
- R7 Markdown and translation files are outside the study's population and should not appear. If one does, answer `covered = false`, `none_kind = null`, rule `R7`, and say so.
- R8 Gap or new: ask whether, at the parent commit, other files of the same purpose existed (gap) or not (new). Use the PR diff and the catalogue's owned files as evidence; when
  the evidence cannot decide, answer `gap`.
- R9 PR title and body describe the pull request. They may help to understand a unit; they never decide a label alone, and a title that names a topic does not move a unit
  there if the unit's text carries another responsibility.
- R10 Package markers and empty files carry no responsibility of their own; a cluster made only of them is out of scope (answer `covered = false`, rule `R10`).
- R11 Acceptable set: list an area only if a maintainer would accept an issue about the unit under it. At most four. Granularity differences (an area and its sub-part) both
  count if both are listed; `primary` is the narrower one.
- R12 Do not judge the quality, importance or correctness of the files or of any area.

## Answer format (one JSON object)
`{"covered": true|false, "acceptable": ["<area id>", ...], "primary": "<area id>"|null, "none_kind": "gap"|"new"|null, "rule": "<rule ids that decided it>", "reason": "<= 2 sentences"}`

## Worked examples (all from fastapi pull requests merged before 2025-01-01, none from a pull request that supplies a round-0 gate unit or the dry batch; the catalogue
lines are the *situation assumed* in the example, not a claim about any real catalogue)
1. **covered = true, primary = lifespan area (R3).** PR #2944 (2023-03-07). Unit: `docs_src/events/tutorial003.py` (an `asynccontextmanager` `lifespan(app)` that loads a model before
   `yield` and clears it after). Catalogue includes "Application lifecycle events: startup and shutdown handlers and lifespan context managers". `acceptable = [lifecycle]`.
2. **covered = true, primary = lifespan area (R2).** Same PR. Unit: `tests/test_tutorial/test_events/test_tutorial003.py`. The test exercises the lifespan example, so it follows the
   lifecycle area, not a generic "Testing" area that the catalogue may also list.
3. **covered = true, primary = interactive-docs area; not the generic tutorial area (R1, R6, R9).** PR #10114 (2023-08-19, title "Add new docs section, How To - Recipes, move docs
   that don't have to be read by everyone to How To"; the title is about moving documentation). Unit: `docs_src/custom_docs_ui/tutorial001.py` (disables the default `/docs` and
   serves `get_swagger_ui_html` / `get_redoc_html` with self-hosted asset URLs). Catalogue includes "Interactive API documentation UI (Swagger UI and ReDoc)" and a generic
   "Tutorial source examples". The unit lives among the tutorial examples, but its responsibility is the documentation UI; `primary = [docs-ui]`.
4. **covered = true, two acceptable areas (R5, R11).** PR #4871 (2023-03-17, `Annotated` support). Unit: `tests/test_ambiguous_params.py` (asserts that a `Path(default=1)` or `Query`
   default inside `Annotated`, repeated FastAPI annotations, and `Depends` combined with a default each raise `AssertionError`). Catalogue includes "Parameter declarations (Path, Query, Header, Cookie,
   Body)" and "Dependency injection and resolution". `acceptable = [parameters, dependencies]`, `primary = parameters` (most assertions are about parameter defaults).
5. **covered = false, none_kind = gap (R4, R8).** PR #11365 (2024-03-28). Unit: `.github/workflows/test-redistribute.yml` (a CI workflow that builds and tests sdists). The
   repository already had workflows before this change, but the catalogue lists no area for CI or repository automation. Rule R4: not covered; R8: the responsibility (CI
   automation) existed, so `gap`.
6. **covered = false, none_kind = gap (R4, R8).** PR #5920 (2023-01-24). Unit: `.github/DISCUSSION_TEMPLATE/questions.yml` (a GitHub Discussions question template). Earlier issue
   templates already existed under `.github/ISSUE_TEMPLATE/` and no listed area covers community templates, so `gap`.
7. **covered = false, none_kind = new (R4, R8).** PR #13123 (2024-12-30). Unit: `scripts/translate.py` (a Typer script that drives a `pydantic_ai` agent with prompts to translate
   documentation pages). No earlier file in the repository generated translations with a model and the catalogue lists no area for it: the change introduces a responsibility, so `new`.
   (Its `requirements-translations.txt`, added by the same PR, serves this script and gets the same label.)
8. **covered = true, primary = response-model area (R1, R2).** PR #1436 (2023-01-07, return-type annotations to declare the response model). Unit:
   `tests/test_response_model_as_return_annotation.py` (builds `BaseUser`/`User` models and checks response validation against the return annotation). Catalogue includes "Response
   models and response validation"; `primary = [response-model]` although the file sits in the flat `tests/` directory with tests for dozens of other things (R6).
9. **covered = true, primary = parameters area (R1, R5).** PR #11867 (2024-08-24, `allow_inf_nan`). Unit: `tests/test_allow_inf_nan_in_enforcing.py` (checks `Query(allow_inf_nan=...)`
   and `Body(allow_inf_nan=False)` inside `Annotated`). `Annotated` here is only the syntax; the responsibility is the parameter option (R5), `primary = parameters`.
10. **covered = false, rule R10.** A cluster whose only members are empty `__init__.py` package markers (for example those that a pull request adds next to new test modules)
    carries no responsibility; it does not normally reach you (the front end drops it) but if it does, answer `covered = false`.
