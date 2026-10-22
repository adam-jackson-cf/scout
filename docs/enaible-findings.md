# Enaible: findings and step-change opportunities

Assessment date: 23 September 2026  
Repository reviewed: `/Users/adamjackson/Projects/complete/enaible`

## Executive assessment

The biggest opportunity is to turn Enaible’s evidence-backed assessments into a continuous workflow that tracks decisions through to verified resolution.

Enaible already combines shared agent instructions, deterministic analyzers, validated evidence artifacts, and agent-specific installation. Its strongest foundation is the separation between evidence and judgment: immutable findings, explicit dispositions, evidence references, validation, and retained reports.

Connecting these capabilities offers more potential value than expanding the skill catalogue. This is a source-based product judgment, not a conclusion from observed user behavior or adoption measurements.

## Scope and verification

The review explored installation, agent adapters, shared skills, analyzer execution, evidence contracts, reporting, context continuity, knowledge-base workflows, and test infrastructure. No source files were changed during the assessment.

- CLI help and basic diagnostics ran successfully.
- The required `scripts/run-ci-quality-gates.sh` stopped during tool-environment setup. `uv` encountered a filesystem permission error while installing Ruff, before the quality checks completed. The gate did not pass.
- `docs/reference` was excluded from content review under the repository’s access rules. Intentionally defective test fixtures were excluded from findings and remediation.
- End-to-end user behavior and performance were not measured. Expected benefits below require validation.

Source links point to the reviewed local checkout. Their line numbers describe the assessment snapshot and may drift as the repository changes.

## K1 — Enaible already has a strong foundation; the next step change is connecting it

Enaible has two distinct users: engineers invoking workflows inside their coding agent, and contributors maintaining those workflows across supported agents.

The main engineering journey is:

```text
Install CLI and agent skills
  → request a named workflow
  → collect deterministic evidence
  → agent inspects and interprets findings
  → validate and render assessment
  → investigate and recommend actions when authorized
```

Contributors author shared skills, render agent-specific outputs, and install them through the common CLI.

The evidence and judgment contracts provide a useful basis for trustworthy engineering decisions. The next opportunity is continuity: retaining the relationship between a finding, a decision, an authorized change, and evidence that the change resolved the issue.

Evidence: [assessment contract](/Users/adamjackson/Projects/complete/enaible/shared/skills/analyze-code-quality/SKILL.md:70).

## K2 — Highest priority: connect findings, decisions, changes, and verification across runs

### Observed state

Analyzer workflows retain substantial context and can reopen a retained run. Planning also has resumable state, while handoff and recent-context skills support session continuity. However, the inspected analyzer lifecycle does not provide a shared, persistent work item connecting:

```text
Finding → investigation → accepted action → implementation → verification
```

Normalized finding references use the source artifact and record index. Those identify evidence within a run, but are unsuitable as durable issue identities across reordered results.

Evidence: [finding normalization](/Users/adamjackson/Projects/complete/enaible/tools/enaible/src/enaible/workflows/analyzer_skills/phases/summaries.py:144) and [retained-run resolution](/Users/adamjackson/Projects/complete/enaible/tools/enaible/src/enaible/workflows/analyzer_skills/runner.py:198).

### Proposed improvement

Add persistent issue identities linked to immutable run evidence, plus a common workflow state recording the next action, approvals, dependencies, and verification results. Reuse the existing retained-run machinery rather than introducing another independent history system.

The user should be able to say:

> Continue the investigation, implement the approved action, and verify whether the original finding is resolved.

### First increment and success criteria

Connect one code-quality finding across assessment, investigation, a user-authorized change, and a verification run.

- A fresh agent session resumes correctly.
- Closure requires relevant verification evidence.
- A finding disappearing because its analyzer was skipped remains unresolved.

This would reduce the coordination users must perform between sessions and make the outcome of an assessment inspectable beyond its report.

## K3 — Make repeated reviews answer “what changed?”

### Observed state

Enaible supports path, staged, commit, and repository scopes. Selecting changed files is useful, but differs from comparing findings against a baseline.

The current deterministic phase regenerates inventory, reconnaissance, and analyzer outputs. Analyzer invocations run sequentially.

Evidence: [scope resolution](/Users/adamjackson/Projects/complete/enaible/tools/enaible/src/enaible/runtime/analyzer_scope.py:40), [workflow execution](/Users/adamjackson/Projects/complete/enaible/tools/enaible/src/enaible/workflows/analyzer_skills/runner.py:339), and [analyzer execution](/Users/adamjackson/Projects/complete/enaible/tools/enaible/src/enaible/workflows/analyzer_skills/phases/analyzer_execution.py:18).

### Proposed improvement

Build baseline comparison on the persistent identities in K2. Classify findings as new, unchanged, resolved, regressed, or unverified. Carry previous adjudications forward only when their supporting evidence remains applicable.

This would let users ask:

> What risk did this change introduce, and which previously accepted decisions need revisiting?

Add selective evidence reuse using source content, configuration, analyzer versions, and dependency information. Preserve explicit uncertainty when evidence cannot safely be reused.

### First increment and success criteria

Compare two runs of one analyzer before introducing caching or parallel execution.

- Unchanged findings do not demand repeated triage.
- Uncertain matches remain visible.
- Selective execution agrees with full execution on representative changes.

The expected benefit is less repeated work and a clearer daily review experience. Runtime savings need measurement.

## K4 — Test the quality of agent judgment as a product capability

### Observed state

Existing slow seams exercise real analyzers and Codex execution. However, the inspected Codex scenario instructs the agent to run an exact command and return artifact paths. That proves execution and delivery more directly than it proves independent inspection or useful judgment.

There is also a distinction between the full skill and CLI automation: `--phase all` can create empty inspection inputs and unresolved deterministic dispositions. That is a legitimate findings-first path, but does not demonstrate completion of the skill’s manual inspection requirements.

Evidence: [Codex seam scenario](/Users/adamjackson/Projects/complete/enaible/tests/seams/slow/test_full_skill_workflow_seams.py:403), [default assessment generation](/Users/adamjackson/Projects/complete/enaible/tools/enaible/src/enaible/workflows/analyzer_skills/runner.py:138), and [inspection requirements](/Users/adamjackson/Projects/complete/enaible/shared/skills/analyze-code-quality/SKILL.md:43).

### Proposed improvement

Add behavioral evaluations invoking the installed skill through realistic user requests. Include genuine issues, plausible false positives, insufficient evidence, changed source, interrupted sessions, and authorization boundaries.

Evaluate:

- Whether important issues were identified and supported.
- Whether clean code attracted unsupported findings.
- Whether recommendations address the actual mechanism.
- Whether the agent preserves uncertainty and respects scope.

### First increment and success criteria

Evaluate the code-quality journey alongside K2, using reviewed expected outcomes rather than exact report wording.

A workflow change should not improve formatting compliance while silently degrading finding accuracy or recommendation quality. This would provide evidence for the reliability of the judgment users are relying on, alongside the existing execution checks.

## K5 — Make readiness specific to the user’s intended task

### Observed state

Installation provisions a common analyzer baseline, including Python and Node tooling. The current `doctor` checks basic workspace and registry availability; it does not establish that a particular installed agent workflow can complete.

Evidence: [installer preparation](/Users/adamjackson/Projects/complete/enaible/tools/enaible/src/enaible/commands/install.py:474) and [diagnostics](/Users/adamjackson/Projects/complete/enaible/tools/enaible/src/enaible/commands/root.py:175).

### Proposed improvement

Offer capability-based installation and readiness checks: “ready to review this Python repository in Codex,” for example. Resolve required tools from the selected workflows and detected stack, reusing existing analyzer eligibility logic.

Readiness should identify the actual CLI, shared assets, rendered skill versions, required tools, and a small execution proof. Users should receive a precise explanation of any unavailable capability.

### First increment and success criteria

Implement workflow-specific diagnostics, followed by selective installation.

- A fresh user can complete their first useful assessment with clear prerequisites.
- A successful readiness result predicts that workflow’s execution.

This would connect installation success to the task the user actually wants to perform.

## Recommended implementation order

| Order | Improvement | Dependency | Initial proof |
| --- | --- | --- | --- |
| 1 | K2: persistent findings and workflow continuity | Durable issue identity linked to immutable evidence | Resume one finding through authorized action and verified resolution |
| Alongside 1 | K4: behavioral evaluation | Reviewed expected outcomes for the selected journey | Installed skill makes supported judgments and respects authorization boundaries |
| 2 | K3: baseline comparison and selective execution | K2 identities and evidence contracts | Compare two runs reliably before adding caching |
| Independent | K5: workflow-specific readiness | Existing dependency and eligibility information | Readiness predicts successful execution of one selected workflow |

Start with K2 and a small K4 evaluation suite, then build K3 on the resulting identities and evidence contracts. K5 can proceed independently. This order establishes reliable continuity before adding automation that depends on it.
