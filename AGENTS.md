# Scout project rules

## Experiment rules

These apply to every Scout **product** experiment under `evaluation/` (design §13: X1–X6, X9, X10) and to
any other study run in this repository that produces evidence for Scout's design. Hypothesis studies about
codebase inertia (H*, HA*, HQ*) are not run here: they belong to inertia-bench
(`~/Projects/tools/inertia-bench`), under that repository's rules, with Scout as a subject.

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
   run under the experiment rules above or under inertia-bench's rules. Studies run before
   2026-10-22 that were signed off here (E1b-v2) stay citable (grandfathered).

## Ledgers

1. **Research ledger** (`docs/research-ledger.md`): every source that shaped Scout's approach or
   formed a hypothesis is recorded with a link: papers, reports, repositories, tool documentation,
   internal experiments and user direction. Record it when the source is first used to justify a
   decision or a hypothesis. Skill: `.agents/skills/research-ledger/`.
2. **Hypothesis ledger** (`docs/hypothesis-ledger.md`): inertia hypotheses (H*, HA*, HQ*) are
   frozen as of 2026-10-22 as a historical record; new and continuing inertia hypotheses are tracked
   in inertia-bench. Product hypotheses (HD*, and any tied to X1–X6, X9, X10) stay live here: record
   state, tests and outcome whenever one is proposed, preregistered, tested, revised or retired.
   Skill: `.agents/skills/hypothesis-ledger/`.
3. Experiment preregistrations and design changes cite ledger IDs (`R-NNN`, `H-…`) for their sources
   and hypotheses.

## General

- Never modify or delete any `.codegraph` directory.
- **Toolchain.** Work inside the pinned toolchain: `git submodule update --init`, `scripts/bootstrap-tools.sh`,
  `source .tools/activate`. Tool versions change only in `toolchain/versions.env` (Clawpatch: the
  `third_party/clawpatch` submodule); a scanner or Clawpatch version change needs a compatibility review.
- **Jev labelling.** Labels for any Jev question come from the agent protocol: a blind pass by two
  independent sessions (one `claude -p`, one `codex exec`) following a frozen guide, plus an adjudicating
  session where they disagree. The user does not review each case (user decision 2026-10-23,
  `docs/STATUS.md` §6.9); these labels stay provenance `agent_reviewed`, never `human`. Labels are
  appended, never edited; no Jev-versus-label comparison runs outside a signed-off preregistration.
- Never print, log or commit secrets; use presence checks only.
- **Model use.** OpenRouter is used only for TypeSafe Jev. Every other model call (Scout lanes, scripts,
  experiments, agents) goes through `claude -p` or `codex exec` with Anthropic or OpenAI models only; no
  other vendors and no Anthropic/OpenAI API keys. These CLIs use their own subscription login.
- **Secrets.** Scout's only secret is the Jev key (`OPENROUTER_API_KEY`). In CI it is a GitHub Actions secret
  on trusted runs only, never on fork PRs. Locally it comes only from Infisical, injected into the one
  command that needs it: `infisical run --projectId 052b4c81-b9cc-4694-9019-d131ec3817cc --env dev --silent -- <command>`
  (project `dev`; `mac-os-cf`, `ae2a56a4-28d9-429f-bb7c-7d45e26b9ffa`, holds the same key as a fallback). Never export the
  key into the shell, write it to a file, or pass it as an argument; never use another project's secrets (e.g.
  finance-planner). Check availability by presence or length only. Never print Infisical config or vault files
  (`~/.infisical/*`); list key names only.
