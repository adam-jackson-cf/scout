# Label guide `x2-concern-v1/guide-1` (frozen with the preregistration)

## Purpose
Decide, from a pair's evidence alone, whether two Clawpatch agent-mapper features describe the same *concern*. Scout keeps an
**area** across refreshes by matching a new feature to an earlier one; this guide decides whether such a match joins the same concern
or two different concerns. Use only the evidence in the pair; you have no tools and no other context.

## Question (verbatim)
Two features were produced by an automated mapper for the same codebase. Using only the evidence shown, do they describe the same
concern (the same area of responsibility a maintainer would point to for a question like "where does X live?")?

## Evidence shown per pair
Two sides (`left`, `right`; their order is arbitrary): `title`, `summary`, `kind`, up to 8 `entrypoints`, up to 40 `owned_files`
(with `owned_files_total`, so truncation is visible), and for the pair `shared_files_total` and up to 20 `shared_files`. No identifiers, scores,
arm, transition, class or run information are shown. File overlap is visible in the evidence (rules R2/R3 need it), and file paths may reveal the repository.

## Label definitions (criteria, verbatim)
- **same_concern**: Both sides name the same area of responsibility. A maintainer asking where that responsibility lives would accept either
  as the answer, and issues about it would belong to either. Differences are boundary details (a few files in or out), wording or granularity of the title.
- **overlapping_concern**: Both sides share a core responsibility, but one side also carries a distinct responsibility that the other does not
  (roughly a third or more of its stated purpose), or one side is a clearly narrower sub-part of the other.
- **different_concern**: The stated responsibilities differ. Any file overlap is incidental: shared utilities, configuration, type or schema
  files, package markers, tests, documentation or generated files.
- **cannot_tell**: The evidence is too thin or too generic to decide, for example empty or contradictory summaries and no informative titles or paths.

## Rules (derived only from the criteria above)
- R1 Responsibility first: decide from `title`, `summary` and `entrypoints`; file overlap supports a decision but never defines it.
- R2 Generic files: shared files that are generic (`__init__`, `mod`, `lib`, `index`, `types`, `config`, `conftest`, manifests, READMEs) do not
  establish a shared concern.
- R3 Same files, different purpose: if the shared files are many but the stated responsibilities differ, label `different_concern`.
- R4 Different files, same purpose: if the files barely overlap but title and summary describe the same responsibility (the mapper redrew
  the boundary), label `same_concern` or `overlapping_concern` by R5.
- R5 Granularity: one side a strict sub-part of the other's responsibility, or the wider side adding a responsibility of its own that the
  narrower side lacks → `overlapping_concern`. A wider title that only restates the same responsibility → `same_concern`.
- R6 Tie-break: between `same_concern` and `overlapping_concern` choose `overlapping_concern`; between `overlapping_concern` and
  `different_concern` choose `overlapping_concern` only when a shared core responsibility is stated in both summaries, otherwise `different_concern`;
  use `cannot_tell` only when R1–R5 cannot reach any label.
- R7 Do not judge the quality, importance or correctness of either feature.

## Worked examples (synthetic)
1. same_concern. Left "HTTP request routing" (files `app/router.py`, `app/routes/*.py`); right "Request router and path matching"
   (files `app/router.py`, `app/routes/*.py`, `app/path.py`). Same responsibility, boundary differs by one file (R1, R4).
2. overlapping_concern. Left "Authentication and session handling" (auth.py, session.py); right "Authentication" (auth.py). Right is a sub-part of left (R5).
3. different_concern. Left "Database migrations" (`db/migrate.py`, `db/__init__.py`); right "Database connection pooling" (`db/pool.py`, `db/__init__.py`).
   The only shared file is a package marker (R2, R3).
4. same_concern. Left "CLI entrypoint" (`cli/main.py`); right "Command-line interface" (`cli/main.py`, `cli/args.py`, `cli/__init__.py`). Same responsibility (R1).
5. different_concern. Left "Serialization of response bodies" (`out/render.py`, `common/types.py`); right "Request validation" (`in/validate.py`, `common/types.py`). Shared
   `types.py` is generic (R2).
6. cannot_tell. Left title "Misc", empty summary, files `util/a.py`; right title "Helpers", empty summary, files `util/a.py`, `util/b.py`. No stated responsibility (R1, R6).
