# Scout project rules

## Experiment rules

These apply to every study under `evaluation/` that produces evidence for Scout's design.

1. **Innovation council sign-off before any experiment is conducted.** The innovation council must
   review both the experiment design (preregistration) and its implementation (scripts, data
   collection, models), and report **zero issues**. Iterate review → fix → re-review until that is
   true. An experiment with any open council issue must not be run.
2. **Every review round is recorded** in `evaluation/<experiment>/reviews/round-<N>.md`: issues
   found, severity, the fix made, and the council's verdict. The final round contains an explicit
   sign-off statement naming the reviewed preregistration and script digests (SHA-256).
3. **Preregister before analysis.** The preregistration fixes the question, units, outcomes,
   predictors, models, decision rules and threats before any predictor–outcome relationship is
   computed on real data. Measurement and construct checks on outcome or predictor
   *distributions* are allowed during review, must be recorded in the review round, and must not
   look at predictor–outcome associations.
4. **Validate the implementation before the real run.** The pipeline must recover a planted effect
   and report no effect on a planted null, using synthetic or permuted data, and those checks are
   part of the reviewed implementation.
5. **Frozen after sign-off.** The preregistration and signed-off scripts are not changed after
   sign-off. Any change, however small, needs a new council review round with zero issues before
   the run continues. Unavoidable runtime deviations are recorded under `Deviations` in RESULTS.md.
6. **Report everything preregistered**, with uncertainty intervals, `unmeasured` counts, deviations
   and reproduce commands. Never fabricate or round away numbers; mark inferences `[INFERENCE]`.
7. **Superseded studies are kept**, marked superseded at the top of their RESULTS.md with a link to
   the replacement and the reason.

## Design rules

1. **Innovation council sign-off for design changes.** Any change to Scout's design
   (`docs/scout-design.md` and the documents it governs) must be reviewed by the innovation council
   and signed off with zero issues, iterating until true, before it is adopted.
2. Design review rounds are recorded in `docs/design-reviews/<YYYY-MM-DD>-<topic>/round-<N>.md`
   with the same sign-off requirements as experiments.
3. Design changes driven by experiment results may only cite experiments that were signed off and
   run under the experiment rules.

## Ledgers

1. **Research ledger** (`docs/research-ledger.md`): every source that shaped Scout's approach or
   formed a hypothesis is recorded with a link: papers, reports, repositories, tool documentation,
   internal experiments and user direction. Record it when the source is first used to justify a
   decision or a hypothesis. Skill: `.agents/skills/research-ledger/`.
2. **Hypothesis ledger** (`docs/hypothesis-ledger.md`): every hypothesis is recorded with its
   state, the experiments that test it and their outcome. Update it whenever a hypothesis is
   proposed, preregistered, tested, revised or retired. Skill: `.agents/skills/hypothesis-ledger/`.
3. Experiment preregistrations and design changes cite ledger IDs (`R-NNN`, `H-…`) for their sources
   and hypotheses.

## General

- Never modify or delete any `.codegraph` directory.
- Never print, log or commit secrets; use presence checks only.
- **Secrets for study and Jev runs** (`OPENROUTER_API_KEY`) come only from Infisical, injected into the one
  command that needs them: `infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- <command>`
  (project `dev`; `mac-os-cf`, `ae2a56a4-28d9-429f-bb7c-7d45e26b9ffa`, holds the same key as a fallback). Never export the
  key into the shell, write it to a file, or pass it as an argument; never use another project's secrets (e.g.
  finance-planner). Check availability by presence or length only. Never print Infisical config or vault files
  (`~/.infisical/*`); list key names only.
