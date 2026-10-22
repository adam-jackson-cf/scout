# Codebase inertia: research notes (2026-10-19)

Purpose: ground Scout's quality/inertia hypotheses and measurements in published evidence. Every
finding below is **associational** unless stated; none is a controlled causal estimate.

## 1. The concept

"Codebase inertia" is not a standardised research term. It is an informal name for a codebase's
accumulated resistance to change. The established literature studies it under several names:

| Name | Source | What it contributes |
|---|---|---|
| Technical debt: principal vs **interest** | Cunningham; Besker, Martini & Bosch, JSS 2019 | Structure is the principal; the extra cost paid on every change is the interest. |
| Lehman's laws II (increasing complexity) and VII (declining quality) | Lehman 1980s; Neamtiu et al. 2013 (9 OSS projects) | Complexity tends to grow *unless work is done to reduce it*. Empirical support is mixed and project-dependent, so this is a trend to monitor, not an inevitability. |
| Architecture erosion and drift | Li, Liang, Soliman & Avgeriou 2022 (mapping study, 73 studies); de Silva & Balasubramaniam 2012 | Symptoms: structural (cycles, undesired dependencies), violation (broken layering/decisions), quality (rising defects), evolution (rigidity, brittleness). |
| Complexity symptoms | Ousterhout, *A Philosophy of Software Design* ch. 2 | **Change amplification**, **cognitive load**, **unknown unknowns** (the most dangerous); causes are dependencies and obscurity. |
| Architecture debt | Xiao, Cai, Kazman et al., ICSE 2016; Mo et al., ICSE 2016 | A few structural patterns concentrate most maintenance cost. |
| Socio-technical congruence (Conway) | Cataldo, Herbsleb & Carley 2008; Cataldo & Herbsleb 2013 | Mismatch between technical dependencies and who coordinates slows work and causes failures. |

## 2. Consequences, with evidence

| Consequence | Evidence | Strength |
|---|---|---|
| **More effort per change** | Tornhill & Borg, "Code Red", TechDebt 2022 (39 proprietary codebases, 30,737 files): low-health code took **124% longer** to resolve issues; warning-level code 78% longer. Besker et al. 2019: developers report **23%** of working time lost to technical debt. Xiao et al. 2016: identified architecture debts accounted for **51–85%** of bug-fixing churn in 7 Apache projects. | Large observational effects; effort is proxied (Jira transitions, churn, self-report). |
| **Less predictable delivery** | Code Red: maximum issue-resolution time **~9×** longer in low-health code. | Observational; a distinct outcome from the mean effort. |
| **More defects** | Code Red: **~15×** more defects in low-health files. D'Ambros, Lanza & Robbes, WCRE 2009: change coupling correlates with defects, often better than complexity metrics (but raw change count was stronger). Bird et al. 2011: ownership dispersion associated with defects (a 2024 replication across 9 projects found no strong correlation). | Mixed; strongly confounded by change frequency. E1 here found static complexity adds ~1–2%. |
| **Lower developer productivity** | Cheng et al., "What Improves Developer Productivity at Google? Code Quality", FSE 2022: lagged panel data showed improvements in perceived code quality **preceded** improvements in perceived productivity, not the reverse. | Strongest causal-direction evidence found; outcomes are perceptions. |
| **Slower comprehension** | Muñoz Barón, Wyrich & Wagner, ESEM 2020 (meta-analysis, ~24,000 evaluations, 427 snippets): higher Cognitive Complexity associated with longer comprehension time; a 2022 re-analysis found it no better than traditional metrics. | Moderate, for comprehension time. |
| **Slower coordination** | Cataldo et al. 2008: congruence between coordination needs and actual coordination associated with modification requests resolved **32% faster**; logical (co-change) dependencies identified coordination needs better than call/data dependencies. | Moderate, two industrial settings. |
| **Delivery instability** | DORA 2024: per 25% increase in AI adoption, delivery stability **−7.2%**, throughput −1.5%; proposed mechanism is larger batch sizes. | Survey association; mechanism hypothesised. |
| **Runtime performance / energy** | Palomba et al. 2019 and follow-ups (Android): *specific* smells (slow loops, leaking threads) raise energy use; a 2025 mapping study finds no conclusive general effect of refactoring on energy. | Weak for general degradation; holds only for specific mechanisms. |
| **AI-era degradation trend** | GitClear 2025 (211M changed lines, 2020–2024): moved (refactored) lines fell from 24.8% to 9.5%; copy/pasted lines rose to 12.3%; short-term churn +26%; duplicated blocks of 5+ lines ~8×. | Industry trend data; not causal. |

## 3. Measures found in the literature (candidates for Scout signals)

| Measure | Definition | Evidence of validity | Scout fit |
|---|---|---|---|
| **CRAP** (Change Risk Anti-Patterns) | `comp(m)² × (1 − cov(m)/100)³ + comp(m)`; > 30 flagged. Savoia & Evans 2007; Google Testing Blog 2011 "This Code is CRAP" | No published validation of the combined score was found; it is a triage heuristic. | Per-function, needs coverage from CI artifacts; descriptive. |
| Hotspots | complexity × change frequency (Tornhill) | Debt only costs where code changes; consistent with E1 where history carried most signal. | Deterministic from lizard + git. |
| Change coupling / logical coupling | files that co-change, esp. across feature boundaries and without a static dependency | D'Ambros 2009; Cataldo 2008. Unexplained coupling is a proxy for Ousterhout's "unknown unknowns". | Git-only, language-agnostic. |
| Change amplification | files/modules touched per change | Ousterhout; Code Red-style effort. | Git-only. |
| Propagation cost; Decoupling Level | share of the system reachable through dependencies (MacCormack, Rusnak & Baldwin 2006); DL (Mo et al. 2016, DL correlated −0.61 to −0.77 with change/bug coupling) | Method papers; association with maintenance coupling. | Needs import graph per language. |
| Architecture debt patterns | modularity violations, unstable interfaces, cycles (Xiao et al. 2016) | Concentrate 51–85% of bug-fix churn. | Import graph + co-change. |
| Cognitive Complexity | nesting-weighted control-flow complexity (SonarSource) | Comprehension time association. | Per function; tools exist for Go/Python/TS. |
| Ownership / truck factor | top-owner share, minor contributors, knowledge concentration | Bird 2011, mixed replication. | Git-only; Atlas commit-team already computes. |
| Moved vs copy/pasted lines, short-term churn | GitClear | Trend indicator. | Git diff analysis. |
| Batch size | lines/files per PR | DORA 2024 mechanism. | PR metadata. |
| Code health composite | CodeScene (used in Code Red) | Large effect sizes but proprietary definition. | Not reproducible; reference only. |

## 4. Implications for Scout's hypotheses

- E1 tested defect prediction from complexity *levels*. The literature's strongest associations are with
  **effort, predictability, and productivity**, not only defects, which supports testing change cost directly.
- The debt metaphor separates **principal** (structural signals) from **interest** (change-cost outcomes).
  Scout should measure both per area: high principal paid on frequent changes is the actionable inertia.
- New outcome: **variance** of change effort (predictability), separate from its mean.
- New predictors: change coupling (especially cross-boundary and unexplained by static dependencies),
  architecture debt patterns, CRAP/untested complexity, ownership concentration, batch size, copy/paste vs move.
- Performance should not be a general consequence hypothesis; keep it to specific mechanisms.
- Lehman's mixed evidence means increase is not guaranteed; trend direction is itself the thing to report.

## 5. Agent-era hypothesis (motivation for Scout, not a Scout lane)

User hypothesis: the tech-debt evidence above comes from human-written code. Agents now do much of the
work, misapply established patterns, are misled by contradictory context, and work far faster, so the
effects of debt are magnified.

| Strand | Evidence | Strength |
|---|---|---|
| A. Agents create degradation faster | SlopCodeBench (Orlanski et al., arXiv 2603.24755; 36 problems, 196 checkpoints, 15 agents): erosion rose in **77%** and verbosity in **75.5%** of trajectories; agent code **2.0× more eroded, 2.3× more verbose** than 473 sampled OSS Python repos; quality prompts improved the starting code but **did not stop later degradation**. GitClear 2025 trend; DORA 2024 batch-size mechanism. | Benchmark plus trend data; supports A. |
| B. Existing degradation harms agents more | Borg, Hagatulah, Tornhill & Söderberg, "Code for Machines, Not Just Humans" (arXiv 2601.02200, 2026): medium-sized LLMs broke behaviour **15–30% less often** when refactoring healthy code; Claude Code showed no significant difference. | One refactoring experiment on competitive-programming Python; partial support. |
| C. Velocity multiplies the interest | No direct study found. METR RCT (July 2025; 16 developers, 246 tasks): AI-allowed tasks took **19% longer** (95% CI 2–39%) while developers believed they were ~20% faster. | Indirect: perceived speed is not a reliable gauge, which argues for measured monitoring. |
| D. Pattern misapplication and contradictory context | No study found that measures this directly. | Untested; candidate signals below. |

The combination of A and B is a feedback loop: agents degrade code, and degraded code makes agents
less reliable, which degrades it further. That loop is the "magnification" claim; it is untested.

Candidate agent-specific signals (no published validation): competing implementations of the same
concern (pattern inconsistency), instructions/docs/comments that contradict the code (AGENTS.md, README,
docstrings), and stale examples that an agent would copy.

Controlled test available locally: SlopCodeBench / factory-x-loops "controlled continuations" (same task
started from a healthy vs a degraded snapshot, fresh agent session), measuring cost, steps, success and
subsequent erosion. SCB's `measurement-validity.md` lists this "causal inertia" claim as unvalidated.
