# Requirements migration map

Authority: `docs/scout-design.md` revision 2.8, including its authoritative Appendix; `SCOUT_REBUILD_REQUIREMENTS.md` and adopted `docs/scout-allowlist.md`. This is a normative-content migration, not a new product design. ADR-009 governs the document restructure; older adopted decisions retain their ADR IDs. Parent-section mappings are unions so broad code citations do not silently lose subordinate requirements: a parent row contains every REQ and ADR ID named by its child rows (subsections, sub-items, rows whose source lies under the parent's heading in the base document, and the `req REQ-n` rows of the old section that held REQ-n), including the targets named by `dropped` children, and a row whose source names several sections adds its IDs to each of them. Exception: the document-title rows ("Design title / status / adoption and implementation chronology", "AL Scout PR blocking allow-list at launch — council decision memo", "req title / revision/status authority paragraphs"), "Design Appendix heading" and "req Appendix — Removed / superseded heading and provenance note" map only their own text (the heading and the prose named in their Source), not their children's IDs; the generic range provenance in the removed-item rows ("REQ-1–REQ-41 … ADR-000–ADR-008") is not a union target. Numeric thresholds and provisional markers retain their original force.

All 41 legacy IDs remain live. Their wording is rewritten into compact requirements; detailed clauses split into new requirements without reusing or withdrawing IDs. REQ-37 is corrected to the authoritative principles-only subject decision: the legacy exact schema is not adopted. The Appendix pin `dc867503…e6ce3` is deliberately copied in its source abbreviation rather than inventing a full digest. [INFERENCE] No product requirement is inferred from implementation-status prose, risk likelihood ratings or observed performance/accuracy figures.

`req §n` denotes sections of the current requirements source; explicitly marked historical `req §11`–`§13` entries cover old section references that the source still discusses. `REQ` targets may also name the governing ADR; `ADR` means the decision/evidence record, not an additional runtime requirement.

| Source | Disposition | Target |
|---|---|---|
| Design title / status / adoption and implementation chronology | ADR | ADR-000, ADR-001, ADR-002, ADR-003, ADR-004, ADR-005, ADR-006, ADR-007, ADR-008, ADR-009 |
| Design §1 — 1. Product definition and non-goals | REQ | REQ-1, REQ-2, REQ-4, REQ-42, REQ-43 |
| Design §2 — 2. Council synthesis — where the challenger changed v2 | ADR | ADR-000, ADR-003 |
| Design §3 — 3. Architecture and data model | REQ | REQ-3, REQ-39, REQ-40, REQ-5, REQ-6, REQ-7, REQ-10, REQ-44, REQ-45, REQ-46, REQ-47, REQ-48, REQ-49, REQ-50, REQ-51, REQ-8, REQ-9, REQ-11, REQ-33, REQ-16, REQ-18, REQ-28 |
| Design §3.1 — 3.1 Components | REQ | REQ-3, REQ-39, REQ-40, REQ-28 |
| Design §3.2 — 3.2 Record types | REQ | REQ-8, REQ-9, REQ-11, REQ-16, REQ-33, REQ-44, REQ-45, REQ-5, REQ-7 |
| Design §3.3 — 3.3 Issue identity and anchoring | REQ | REQ-5, REQ-6, REQ-7, REQ-10, REQ-44, REQ-45, REQ-46, REQ-47, REQ-48, REQ-49, REQ-50, REQ-51 |
| Design §3.4 — 3.4 Lanes | REQ | REQ-9, REQ-16, REQ-18 |
| Design §4 — 4. Feature mapping (DECIDED 1, 7 and 21 applied) | REQ | REQ-12, REQ-13, REQ-14, REQ-15, REQ-31, REQ-52, REQ-53, REQ-54, REQ-55, REQ-56, REQ-57, REQ-58, REQ-59, REQ-60, REQ-61, REQ-139, REQ-140, REQ-141, REQ-142, REQ-143, REQ-144, REQ-145, REQ-146, REQ-147, REQ-148, REQ-149, REQ-150, REQ-151, REQ-152, REQ-153, REQ-154, REQ-155, REQ-156, REQ-157, REQ-158, REQ-159, ADR-000, ADR-005, ADR-006, ADR-008 |
| Design §5 — 5. Analyzer lanes and languages (DECIDED 5 applied) | REQ | REQ-2, REQ-17, REQ-62, REQ-63, REQ-64, REQ-65, REQ-66, REQ-67, REQ-68, REQ-69, REQ-18, ADR-000 |
| Design §5.1 — 5.1 Lane catalogue (priority order) | REQ | REQ-2, REQ-17, REQ-62, REQ-63, REQ-64, REQ-65, REQ-66, REQ-67, REQ-68, REQ-69 |
| Design §5.2 — 5.2 TypeScript at tier 1 | REQ | REQ-62, REQ-63, REQ-64, REQ-65, REQ-66 |
| Design §5.3 — 5.3 Capability matrix (`scout readiness --matrix`) | REQ | REQ-2, REQ-18, REQ-62, REQ-63, REQ-64, REQ-65, REQ-66, ADR-000 |
| Design §6 — 6. Placement of deterministic code, decision models, LLMs, humans (DECIDED 4 applied) | REQ | REQ-7, REQ-28, REQ-29, REQ-30, REQ-31, REQ-32, REQ-59, REQ-60, REQ-61, REQ-67, REQ-68, REQ-69, REQ-86, REQ-87, REQ-88, REQ-89, REQ-90, REQ-91, REQ-92, REQ-93, REQ-94, REQ-95, REQ-96, REQ-97, REQ-98, REQ-99, REQ-100, REQ-101, REQ-102, REQ-139, REQ-140, REQ-141, REQ-142, REQ-143, REQ-144, REQ-145, REQ-146, REQ-147, REQ-148, REQ-149, REQ-150, REQ-151, REQ-152, REQ-153, REQ-154, REQ-155, REQ-156, REQ-157, REQ-158, REQ-159, ADR-000, ADR-002 |
| Design §7 — 7. CI guardian contract (DECIDED 2 and 3 applied) | REQ | REQ-1, REQ-3, REQ-160, REQ-161, REQ-162, REQ-163, REQ-164, REQ-165, REQ-166, REQ-167, REQ-168, REQ-26, REQ-27, REQ-46, REQ-47, REQ-48, REQ-49, REQ-50, REQ-51, REQ-7, REQ-11, REQ-16, REQ-19, REQ-20, REQ-21, REQ-22, REQ-23, REQ-70, REQ-71, REQ-72, REQ-73, REQ-74, REQ-75, REQ-76, REQ-77, REQ-78, REQ-79, REQ-80, REQ-81, REQ-82, REQ-83, REQ-84, REQ-85, REQ-24, REQ-25, REQ-32, REQ-40 |
| Design §7.1 — 7.1 Triggers | REQ | REQ-1, REQ-3, REQ-160, REQ-161, REQ-162, REQ-163, REQ-164, REQ-165, REQ-166, REQ-167, REQ-168 |
| Design §7.2 — 7.2 Baseline — protected `scout/state` git branch (DECIDED) | REQ | REQ-3, REQ-26, REQ-27, REQ-46, REQ-47, REQ-48, REQ-49, REQ-50, REQ-51 |
| Design §7.3 — 7.3 Gate semantics | REQ | REQ-7, REQ-11, REQ-16, REQ-19, REQ-20, REQ-21, REQ-22, REQ-23, REQ-27, REQ-70, REQ-71, REQ-72, REQ-73, REQ-74, REQ-75, REQ-76, REQ-77, REQ-78, REQ-79, REQ-80, REQ-81, REQ-82, REQ-83, REQ-84, REQ-85 |
| Design §7.4 — 7.4 Exit codes, outputs, trust boundary | REQ | REQ-24, REQ-25, REQ-27, REQ-32, REQ-40, REQ-160, REQ-161, REQ-162, REQ-163, REQ-164, REQ-165, REQ-166, REQ-167, REQ-168 |
| Design §8 — 8. Inertia monitoring: principal and interest over time | REQ | REQ-33, REQ-103, REQ-104, REQ-105, REQ-106, REQ-107, REQ-108, REQ-109, REQ-110, REQ-111, REQ-125, REQ-126, REQ-127, REQ-128, REQ-129, REQ-130, REQ-131, REQ-112, REQ-113, REQ-114, REQ-115, REQ-116, REQ-117, REQ-118, REQ-119, REQ-120, REQ-121, REQ-122, REQ-123, REQ-124, REQ-34, REQ-35, REQ-36, REQ-132, REQ-133, REQ-134, REQ-135, REQ-37, REQ-42, ADR-001, ADR-003 |
| Design §8.1 — 8.1 Framing and claim level | REQ | REQ-33, REQ-35 |
| Design §8.2 — 8.2 Principal measures (snapshot; function → file → scope) | REQ | REQ-33, REQ-103, REQ-104, REQ-105, REQ-106, REQ-107, REQ-108, REQ-109, REQ-110, REQ-111, REQ-125, REQ-126, REQ-127, REQ-128, REQ-129, REQ-130, REQ-131 |
| Design §8.3 — 8.3 Exposure and interest measures (trailing window; per scope) | REQ | REQ-33, REQ-112, REQ-113, REQ-114, REQ-115, REQ-116, REQ-117, REQ-118, REQ-119, REQ-120, REQ-121, REQ-122, REQ-123, REQ-124, REQ-125, REQ-126, REQ-127, REQ-128, REQ-129, REQ-130, REQ-131 |
| Design §8.4 — 8.4 Reporting over time per repository | REQ | REQ-33, REQ-34, REQ-125, REQ-126, REQ-127, REQ-128, REQ-129, REQ-130, REQ-131 |
| Design §8.5 — 8.5 Honest reading of the evidence, and what would upgrade a claim | REQ | REQ-35, REQ-42, REQ-36, REQ-129, ADR-001, ADR-003 |
| Design §8.6 — 8.6 Accumulating evidence from adopters without overclaiming | REQ | REQ-36, REQ-132, REQ-133, REQ-134, REQ-135 |
| Design §8.7 — 8.7 Subject interface (principles only) | REQ | REQ-37, REQ-132, REQ-133, REQ-134, REQ-135 |
| Design §9 — 9. Scout's own outcome scorecard | REQ | REQ-41, REQ-136, REQ-137, REQ-138, REQ-19 |
| Design §10 — 10. Keep / cut / rewrite | REQ | REQ-14, REQ-39, REQ-59, REQ-60, REQ-61, REQ-160, REQ-161, REQ-162, REQ-163, REQ-164, REQ-165, REQ-166, REQ-167, REQ-168, REQ-38, REQ-84, REQ-85, ADR-009, ADR-000, ADR-002, ADR-003, ADR-007, ADR-008 |
| Design §10.1 — 10.1 P1 performance defect — `scout assess` run-job persistence | REQ | REQ-38, REQ-160, REQ-161, REQ-162, REQ-163, REQ-164, REQ-165, REQ-166, REQ-167, REQ-168, ADR-000, ADR-002 |
| Design §10.2 — 10.2 Code verdicts | REQ | REQ-14, REQ-39, REQ-59, REQ-60, REQ-61, REQ-160, REQ-161, REQ-162, REQ-163, REQ-164, REQ-165, REQ-166, REQ-167, REQ-168, ADR-000, ADR-002, ADR-003, ADR-007, ADR-008 |
| Design §10.3 — 10.3 Spec cutover (DECIDED 6) | REQ | REQ-85, ADR-009 |
| Design §11 — 11. Phased plan — every phase ends in a runnable scenario | REQ | REQ-10, REQ-37, REQ-46, REQ-47, REQ-48, REQ-49, REQ-50, REQ-51, REQ-160, REQ-161, REQ-162, REQ-163, REQ-164, REQ-165, REQ-166, REQ-167, REQ-168, REQ-5, REQ-6, REQ-38, REQ-20, REQ-21, REQ-22, REQ-23, REQ-24, REQ-25, REQ-26, REQ-27, REQ-2, REQ-18, REQ-19, REQ-33, REQ-103, REQ-104, REQ-105, REQ-106, REQ-107, REQ-108, REQ-109, REQ-110, REQ-111, REQ-112, REQ-113, REQ-114, REQ-115, REQ-116, REQ-117, REQ-118, REQ-119, REQ-120, REQ-121, REQ-122, REQ-123, REQ-124, REQ-62, REQ-63, REQ-64, REQ-65, REQ-125, REQ-126, REQ-127, REQ-128, REQ-129, REQ-130, REQ-131, REQ-135, REQ-13, REQ-14, REQ-52, REQ-53, REQ-54, REQ-55, REQ-56, REQ-29, REQ-34, REQ-41, REQ-86, REQ-87, REQ-88, REQ-89, REQ-90, REQ-91, REQ-92, REQ-93, REQ-94, REQ-95, REQ-96, REQ-97, REQ-98, REQ-99, REQ-100, REQ-101, REQ-102, REQ-136, REQ-137, REQ-138, REQ-36, REQ-39, REQ-40, REQ-132, REQ-133, REQ-178, REQ-179, REQ-180, REQ-181, REQ-182, REQ-183, ADR-009 |
| Design §12 — 12. Top risks | REQ | REQ-40, REQ-5, REQ-6, REQ-10, REQ-21, REQ-22, REQ-23, REQ-33, REQ-34, REQ-35, REQ-36, REQ-26, REQ-54, REQ-25, REQ-38, REQ-29, REQ-30, REQ-94, REQ-131, REQ-133, REQ-18, REQ-31, REQ-87, REQ-58, REQ-56, REQ-13, REQ-14, REQ-27, REQ-46, REQ-159, REQ-183 |
| Design §13 — 13. Experiments with decision rules | REQ | REQ-67, REQ-68, REQ-69, REQ-169, REQ-170, REQ-171, REQ-172, REQ-173, REQ-174, REQ-175, REQ-176, REQ-177, ADR-002 |
| Design §14 — 14. Assumptions register | REQ | REQ-53, REQ-58, REQ-22, REQ-26, REQ-30, REQ-69, REQ-29, REQ-34, REQ-126, REQ-128, ADR-002, REQ-65, REQ-38, REQ-109, REQ-67, REQ-68, REQ-10, REQ-46, REQ-159, REQ-146, REQ-147, REQ-148, REQ-149, REQ-150, ADR-000, ADR-005, ADR-006, ADR-007, ADR-008 |
| Design §15 — 15. Decisions for the user (remaining) | REQ | REQ-98, ADR-002, ADR-003, ADR-005, ADR-006, ADR-000 |
| Design Appendix heading | ADR | ADR-000, ADR-002, ADR-003, ADR-004, ADR-005, ADR-006, ADR-007, ADR-008, ADR-009 |
| Design revision 2.1 paragraph | ADR | ADR-001 |
| Design revision 2.2 paragraph | ADR | ADR-002 |
| Design revision 2.3 paragraph | ADR | ADR-003 |
| Design revision 2.4 paragraph | ADR | ADR-004 |
| Design revision 2.5 paragraph | ADR | ADR-005 |
| Design revision 2.6 paragraph | ADR | ADR-006 |
| Design revision 2.7 paragraph | ADR | ADR-007 |
| Design revision 2.8 paragraph | ADR | ADR-008 |
| Design revision 2.8 adoption notice (before revision 2.2) | ADR | ADR-008 |
| Design evidence notation (aliases and study admissibility) | ADR | ADR-009, ADR-001, ADR-002, ADR-008 |
| Design evidence notation (.codegraph / read-only constraint) | REQ | REQ-43 |
| Design Appendix 22 | REQ | REQ-10, REQ-25, REQ-27, REQ-31, REQ-46, REQ-47, REQ-48, REQ-49, REQ-50, REQ-51, REQ-59, REQ-60, REQ-61, ADR-007, REQ-162, REQ-163 |
| Design Appendix 23 | REQ | REQ-10, REQ-13, REQ-28, REQ-31, REQ-139, REQ-140, REQ-141, REQ-142, REQ-143, REQ-144, REQ-145, REQ-146, REQ-147, REQ-148, REQ-149, REQ-150, REQ-151, REQ-152, REQ-153, REQ-154, REQ-155, REQ-156, REQ-157, REQ-158, REQ-159, ADR-008 |
| Design Appendix 7 | REQ | REQ-12, REQ-57, REQ-58, ADR-004 |
| Design Appendix 1 | REQ | REQ-13, REQ-52, REQ-53, REQ-54, REQ-55, REQ-56, ADR-000 |
| Design Appendix 21 | REQ | REQ-13, REQ-52, REQ-53, ADR-006 |
| Design §9.8 | REQ | REQ-41, REQ-19 |
| Design Appendix 3 | REQ | REQ-19, REQ-21, REQ-22, REQ-23, REQ-84, REQ-85, ADR-000 |
| Design Appendix 17 | REQ | REQ-25, REQ-161, ADR-003 |
| Design Appendix 2 | REQ | REQ-26, ADR-000 |
| Design Appendix 4 | REQ | REQ-28, REQ-30, ADR-000 |
| Design Appendix 8 | REQ | REQ-31, REQ-32, REQ-59, REQ-60, REQ-61, ADR-002 |
| Design Appendix 20 | REQ | REQ-31, REQ-58, ADR-005, REQ-53, REQ-170, ADR-006 |
| Design Appendix 11 | REQ | REQ-35, REQ-36, REQ-42, ADR-002 |
| Design Appendix 18 | REQ | REQ-2, REQ-35, ADR-003 |
| Design Appendix 16 | REQ | REQ-36, REQ-132, REQ-133, ADR-003 |
| Design Appendix 10 | REQ | REQ-37, REQ-135, ADR-002 |
| Design Appendix 14 | REQ | REQ-38, ADR-002 |
| Design Appendix 5 | REQ | REQ-2, ADR-000 |
| Design Appendix 6 | ADR | ADR-009 |
| Design Appendix 9 | ADR | ADR-002 |
| Design Appendix 12 | ADR | ADR-009 |
| Design Appendix 13 | REQ | REQ-31, REQ-100, REQ-101, REQ-102, ADR-003 |
| Design Appendix 15 | REQ | REQ-100, REQ-101, ADR-003 |
| Design Appendix 19 | REQ | REQ-29, REQ-30, REQ-86, REQ-87, REQ-88, REQ-89, REQ-90, REQ-91, REQ-92, REQ-93, REQ-94, REQ-95, REQ-96, REQ-97, REQ-98, REQ-99, REQ-100, REQ-101, REQ-102, ADR-003 |
| Design Appendix 22(a) | REQ | REQ-10, REQ-46, REQ-47, ADR-007 |
| Design Appendix 22(b) | REQ | REQ-49, REQ-50, ADR-007 |
| Design Appendix 22(c) | REQ | REQ-59, REQ-60, REQ-61, ADR-007 |
| Design Appendix 22(d) | REQ | REQ-162, REQ-163, ADR-007 |
| Design Appendix 23(a) | REQ | REQ-151, REQ-152, REQ-153, ADR-008 |
| Design Appendix 23(b) | REQ | REQ-152, ADR-008 |
| Design Appendix 23(c) | REQ | REQ-159, ADR-008 |
| Design Appendix 23(d) | REQ | REQ-147, REQ-148, REQ-149, REQ-150, ADR-008 |
| Design Appendix 23(e) | REQ | REQ-10, ADR-008 |
| Design §3.3(a) | REQ | REQ-6 |
| Design §3.3(b) | REQ | REQ-6 |
| Design §3.3(c) | REQ | REQ-6 |
| Design §3.3(d) | REQ | REQ-6 |
| Design §3.3(e) | REQ | REQ-6 |
| Design §8.7.1 | REQ | REQ-37, REQ-135 |
| Design §8.7.2 | REQ | REQ-37, REQ-135 |
| Design §8.7.3 | REQ | REQ-37, REQ-135 |
| Design §8.7.4 | REQ | REQ-37, REQ-135 |
| Design §8.7.5 | REQ | REQ-37, REQ-135 |
| Design §8.7.6 | REQ | REQ-37, REQ-135 |
| Design §9.1 | REQ | REQ-41, REQ-136 |
| Design §9.2 | REQ | REQ-41, REQ-136 |
| Design §9.3 | REQ | REQ-41, REQ-136 |
| Design §9.4 | REQ | REQ-41, REQ-137 |
| Design §9.5 | REQ | REQ-41, REQ-137 |
| Design §9.6 | REQ | REQ-41, REQ-137 |
| Design §9.7 | REQ | REQ-41, REQ-138 |
| Design §1 non-goal 1 | REQ | REQ-4 |
| Design §1 non-goal 2 | REQ | REQ-4 |
| Design §1 non-goal 3 | REQ | REQ-4 |
| Design §1 non-goal 4 | REQ | REQ-4 |
| Design §1 non-goal 5 | REQ | REQ-4 |
| Design §1 non-goal 6 | REQ | REQ-4 |
| Design §1 non-goal 7 | REQ | REQ-4 |
| Design §13 X1 | REQ | REQ-169 |
| Design §13 X2 | REQ | REQ-170 |
| Design §13 X2b | REQ | REQ-170 |
| Design §13 X3 | REQ | REQ-171 |
| Design §13 X4 | REQ | REQ-172 |
| Design §13 X5 | REQ | REQ-173 |
| Design §13 X6 | REQ | REQ-174 |
| Design §13 X9 | REQ | REQ-174 |
| Design §13 X10 | REQ | REQ-175 |
| Design §13 X11 | REQ | REQ-176, REQ-177 |
| Design §13 X7 | ADR | ADR-002 |
| Design §13 X8 | ADR | ADR-002 |
| Design §12 R1 | REQ | REQ-5, REQ-6, REQ-10 |
| Design §12 R2 | REQ | REQ-21, REQ-22, REQ-23 |
| Design §12 R3 | REQ | REQ-33, REQ-34, REQ-35, REQ-183 |
| Design §12 R4 | REQ | REQ-36 |
| Design §12 R5 | REQ | REQ-26 |
| Design §12 R6 | REQ | REQ-33, REQ-54 |
| Design §12 R7 | REQ | REQ-25 |
| Design §12 R8 | REQ | REQ-38 |
| Design §12 R9 | REQ | REQ-40 |
| Design §12 R10 | REQ | REQ-29, REQ-30, REQ-94 |
| Design §12 R11 | REQ | REQ-131 |
| Design §12 R12 | REQ | REQ-36, REQ-133 |
| Design §12 R13 | REQ | REQ-18, REQ-31 |
| Design §12 R14 | REQ | REQ-87, REQ-29, REQ-30 |
| Design §12 R15 | REQ | REQ-58, REQ-56 |
| Design §12 R16 | REQ | REQ-13, REQ-14, REQ-54 |
| Design §12 R17 | REQ | REQ-10, REQ-27, REQ-46 |
| Design §12 R18 | REQ | REQ-159 |
| Design §14 A1 | REQ | REQ-53, REQ-58 |
| Design §14 A2 | REQ | REQ-22 |
| Design §14 A3 | REQ | REQ-26 |
| Design §14 A4 | REQ | REQ-30 |
| Design §14 A5 | REQ | REQ-69 |
| Design §14 A6 | REQ | REQ-29 |
| Design §14 A7 | REQ | REQ-34, REQ-126, REQ-128 |
| Design §14 A8 | ADR | ADR-002 |
| Design §14 A9 | REQ | REQ-65 |
| Design §14 A10 | ADR | ADR-002 |
| Design §14 A11 | REQ | REQ-109 |
| Design §14 A12 | REQ | REQ-67, REQ-68 |
| Design §14 A13 | REQ | REQ-10, REQ-46 |
| Design §14 A14 | REQ | REQ-159, REQ-146, REQ-147, REQ-148, REQ-149, REQ-150 |
| Design §15.1 | REQ | REQ-98 |
| Design §11 Phase 1 | REQ | REQ-5, REQ-6, REQ-10, REQ-38, REQ-46, REQ-47, REQ-164, REQ-165, REQ-178 |
| Design §11 Phase 1 Scenario | REQ | REQ-5, REQ-6, REQ-10, REQ-38, REQ-46, REQ-47, REQ-164, REQ-165, REQ-178 |
| Design §11 Phase 1 Acceptance | REQ | REQ-5, REQ-6, REQ-10, REQ-38, REQ-46, REQ-47, REQ-164, REQ-165, REQ-178 |
| Design §11 Phase 2 | REQ | REQ-20, REQ-21, REQ-22, REQ-23, REQ-24, REQ-25, REQ-26, REQ-27, REQ-51, REQ-179 |
| Design §11 Phase 2 Scenario | REQ | REQ-20, REQ-21, REQ-22, REQ-23, REQ-24, REQ-25, REQ-26, REQ-27, REQ-51, REQ-179 |
| Design §11 Phase 2 Acceptance | REQ | REQ-20, REQ-21, REQ-22, REQ-23, REQ-24, REQ-25, REQ-26, REQ-27, REQ-51, REQ-179 |
| Design §11 Phase 3 | REQ | REQ-2, REQ-18, REQ-19, REQ-33, REQ-103, REQ-104, REQ-105, REQ-106, REQ-107, REQ-108, REQ-109, REQ-110, REQ-111, REQ-112, REQ-113, REQ-114, REQ-115, REQ-116, REQ-117, REQ-118, REQ-119, REQ-120, REQ-121, REQ-122, REQ-123, REQ-124, REQ-62, REQ-63, REQ-64, REQ-65, REQ-125, REQ-126, REQ-127, REQ-128, REQ-129, REQ-130, REQ-131, REQ-180 |
| Design §11 Phase 3 Scenario | REQ | REQ-2, REQ-18, REQ-19, REQ-33, REQ-103, REQ-104, REQ-105, REQ-106, REQ-107, REQ-108, REQ-109, REQ-110, REQ-111, REQ-112, REQ-113, REQ-114, REQ-115, REQ-116, REQ-117, REQ-118, REQ-119, REQ-120, REQ-121, REQ-122, REQ-123, REQ-124, REQ-62, REQ-63, REQ-64, REQ-65, REQ-125, REQ-126, REQ-127, REQ-128, REQ-129, REQ-130, REQ-131, REQ-180 |
| Design §11 Phase 3 Acceptance | REQ | REQ-2, REQ-18, REQ-19, REQ-33, REQ-103, REQ-104, REQ-105, REQ-106, REQ-107, REQ-108, REQ-109, REQ-110, REQ-111, REQ-112, REQ-113, REQ-114, REQ-115, REQ-116, REQ-117, REQ-118, REQ-119, REQ-120, REQ-121, REQ-122, REQ-123, REQ-124, REQ-62, REQ-63, REQ-64, REQ-65, REQ-125, REQ-126, REQ-127, REQ-128, REQ-129, REQ-130, REQ-131, REQ-180 |
| Design §11 Phase 3b | REQ | REQ-37, REQ-135, REQ-181 |
| Design §11 Phase 3b Scenario | REQ | REQ-37, REQ-135, REQ-181 |
| Design §11 Phase 3b Acceptance | REQ | REQ-37, REQ-135, REQ-181 |
| Design §11 Phase 4 | REQ | REQ-13, REQ-14, REQ-52, REQ-53, REQ-54, REQ-55, REQ-56, REQ-167 |
| Design §11 Phase 4 Scenario | REQ | REQ-13, REQ-14, REQ-52, REQ-53, REQ-54, REQ-55, REQ-56, REQ-167 |
| Design §11 Phase 4 Acceptance | REQ | REQ-13, REQ-14, REQ-52, REQ-53, REQ-54, REQ-55, REQ-56, REQ-167 |
| Design §11 Phase 5 | REQ | REQ-29, REQ-34, REQ-41, REQ-86, REQ-87, REQ-88, REQ-89, REQ-90, REQ-91, REQ-92, REQ-93, REQ-94, REQ-95, REQ-96, REQ-97, REQ-98, REQ-99, REQ-100, REQ-101, REQ-102, REQ-136, REQ-137, REQ-138, REQ-182 |
| Design §11 Phase 5 Scenario | REQ | REQ-29, REQ-34, REQ-41, REQ-86, REQ-87, REQ-88, REQ-89, REQ-90, REQ-91, REQ-92, REQ-93, REQ-94, REQ-95, REQ-96, REQ-97, REQ-98, REQ-99, REQ-100, REQ-101, REQ-102, REQ-136, REQ-137, REQ-138, REQ-182 |
| Design §11 Phase 5 Acceptance | REQ | REQ-29, REQ-34, REQ-41, REQ-86, REQ-87, REQ-88, REQ-89, REQ-90, REQ-91, REQ-92, REQ-93, REQ-94, REQ-95, REQ-96, REQ-97, REQ-98, REQ-99, REQ-100, REQ-101, REQ-102, REQ-136, REQ-137, REQ-138, REQ-182 |
| Design §11 Phase 6 | REQ | REQ-36, REQ-39, REQ-40, REQ-132, REQ-133, REQ-168, ADR-009 |
| Design §11 Phase 6 Scenario | REQ | REQ-36, REQ-39, REQ-40, REQ-132, REQ-133, REQ-168, ADR-009 |
| Design §11 Phase 6 Acceptance | REQ | REQ-36, REQ-39, REQ-40, REQ-132, REQ-133, REQ-168, ADR-009 |
| Design §11 Phase 3b Entry | REQ | REQ-37 |
| Design §4 Heuristic map is canonical | REQ | REQ-12 |
| Design §4 Agent-map schedule | REQ | REQ-13, REQ-52 |
| Design §4 Areas | REQ | REQ-53 |
| Design §4 Measurement scopes | REQ | REQ-54 |
| Design §4 Active snapshot membership | REQ | REQ-55 |
| Design §4 Unowned files | REQ | REQ-14 |
| Design §4 Merge-time association | REQ | REQ-139, REQ-140, REQ-141, REQ-142, REQ-143, REQ-144, REQ-145, REQ-146, REQ-147, REQ-148, REQ-149, REQ-150, REQ-151, REQ-152, REQ-153, REQ-154, REQ-155, REQ-156, REQ-157, REQ-158, REQ-159 |
| Design §4 Merge-time association evidence and limits | REQ | REQ-159 |
| Design §4 Refresh guard | REQ | REQ-156, REQ-157, REQ-158, REQ-159 |
| Design §4 Failure semantics | REQ | REQ-56 |
| Design §4 Model CLI | REQ | REQ-31, REQ-59, REQ-60, REQ-61, REQ-58 |
| Design §4 Pin | REQ | REQ-57 |
| Design §5.1 lane 1 | REQ | REQ-17, REQ-62, REQ-63, REQ-64, REQ-65, REQ-66 |
| Design §5.1 lane 2 | REQ | REQ-17, REQ-62, REQ-63, REQ-64, REQ-65, REQ-66 |
| Design §5.1 lane 3 | REQ | REQ-17, REQ-62, REQ-63, REQ-64, REQ-65, REQ-66 |
| Design §5.1 lane 4 | REQ | REQ-17, REQ-62, REQ-63, REQ-64, REQ-65, REQ-66 |
| Design §5.1 lane 5 | REQ | REQ-17, REQ-62, REQ-63, REQ-64, REQ-65, REQ-66 |
| Design §5.1 lane 6 | REQ | REQ-17, REQ-67, REQ-68, REQ-69 |
| Design §5.1 lane 7 | REQ | REQ-17, REQ-62, REQ-63, REQ-64, REQ-65, REQ-66 |
| Design §5.1 lane 8 | REQ | REQ-17, REQ-62, REQ-63, REQ-64, REQ-65, REQ-66 |
| Design §5.1 lane 9 | REQ | REQ-17, REQ-62, REQ-63, REQ-64, REQ-65, REQ-66 |
| Design §6 Typed questions | REQ | REQ-86, REQ-90 |
| Design §6 Pinned model | REQ | REQ-87 |
| Design §6 State and questions | REQ | REQ-88, REQ-89 |
| Design §6 Output | REQ | REQ-28 |
| Design §6 Cap | REQ | REQ-29, REQ-91 |
| Design §6 Failures are failed units | REQ | REQ-92 |
| Design §6 Lane health | REQ | REQ-93 |
| Design §6 Audit | REQ | REQ-94 |
| Design §6 Escalation | REQ | REQ-95 |
| Design §6 Calibration | REQ | REQ-96 |
| Design §6 Launch default and fallback | REQ | REQ-97, REQ-98, REQ-99 |
| Design §6 Labels | REQ | REQ-100, REQ-101, REQ-102 |
| Design §6 Model use | REQ | REQ-31, REQ-32, REQ-59, REQ-60, REQ-61 |
| Design §7.3 Attribution | REQ | REQ-20, REQ-27, REQ-78 |
| Design §7.3 Profiles | REQ | REQ-21 |
| Design §7.3 Allow-list | REQ | REQ-21, REQ-70, REQ-71, REQ-72, REQ-73, REQ-74, REQ-75, REQ-76, REQ-77, REQ-78 |
| Design §7.3 Shadow window | REQ | REQ-22, REQ-78 |
| Design §7.3 Rare security checks | REQ | REQ-22 |
| Design §7.3 Overrides | REQ | REQ-23, REQ-77 |
| Design §7.3 Two-track promotion | REQ | REQ-19, REQ-22, REQ-84, REQ-85 |
| Design §8.5 Framing | REQ | REQ-35 |
| Design §8.5 Further tests belong to inertia-bench | REQ | REQ-42, REQ-35, REQ-36 |
| Design §8.5 Decision rule | REQ | REQ-35 |
| Design §8.5 Consequences for Scout | REQ | REQ-35, REQ-129 |
| Design §10.1 Fix | REQ | REQ-38, REQ-164, REQ-165 |
| Design §10.1 Acceptance | REQ | REQ-38, REQ-164, REQ-165 |
| Design metric scout.cc@1 | REQ | REQ-103 |
| Design metric scout.complex_burden@1 | REQ | REQ-104 |
| Design metric scout.dup_pct@1 | REQ | REQ-105 |
| Design metric scout.cycle_mass@1 | REQ | REQ-106 |
| Design metric scout.crap@1 | REQ | REQ-107 |
| Design metric scout.propagation_cost@1 | REQ | REQ-108 |
| Design metric scout.cognitive_cc@1 | REQ | REQ-109 |
| Design metric scout.stale_reference_rate@1 | REQ | REQ-110 |
| Design metric scout.burden_density@1 | REQ | REQ-111 |
| Design metric scout.changes@1 | REQ | REQ-112 |
| Design metric scout.hotspot@1 | REQ | REQ-113 |
| Design metric scout.churn_per_change@1 | REQ | REQ-114 |
| Design metric scout.amplification@1 | REQ | REQ-115 |
| Design metric scout.follow_up_rate@1 | REQ | REQ-116 |
| Design metric scout.pr_effort@1 | REQ | REQ-117 |
| Design metric scout.dispersion@1 | REQ | REQ-118 |
| Design metric scout.change_coupling@1 | REQ | REQ-119 |
| Design metric scout.ownership@1 | REQ | REQ-120 |
| Design metric scout.rework@1 | REQ | REQ-121 |
| Design metric scout.batch_size@1 | REQ | REQ-122 |
| Design metric scout.moved_vs_pasted@1 | REQ | REQ-123 |
| Design metric scout.fix_touches@1 | REQ | REQ-124 |
| AL §1 eligible table row 1 | REQ | REQ-70 |
| AL §1 eligible table row 2 | REQ | REQ-71 |
| AL §1 eligible table row 3 | REQ | REQ-72, REQ-73 |
| AL §1 eligible table row 4 | REQ | REQ-74, REQ-73 |
| AL §1 eligible table row 5 | REQ | REQ-75 |
| AL §1 eligible table row 6 | REQ | REQ-76 |
| AL §2 report-only table row 1 | REQ | REQ-79, REQ-80, REQ-81, REQ-82, REQ-83 |
| AL §2 report-only table row 2 | REQ | REQ-79, REQ-80, REQ-81, REQ-82, REQ-83 |
| AL §2 report-only table row 3 | REQ | REQ-79, REQ-80, REQ-81, REQ-82, REQ-83 |
| AL §2 report-only table row 4 | REQ | REQ-79, REQ-80, REQ-81, REQ-82, REQ-83 |
| AL §2 report-only table row 5 | REQ | REQ-79, REQ-80, REQ-81, REQ-82, REQ-83 |
| AL §2 report-only table row 6 | REQ | REQ-79, REQ-80, REQ-81, REQ-82, REQ-83 |
| AL §2 report-only table row 7 | REQ | REQ-79, REQ-80, REQ-81, REQ-82, REQ-83 |
| AL §2 report-only table row 8 | REQ | REQ-79, REQ-80, REQ-81, REQ-82, REQ-83 |
| AL §2 report-only table row 9 | REQ | REQ-79, REQ-80, REQ-81, REQ-82, REQ-83 |
| AL Scout PR blocking allow-list at launch — council decision memo | ADR | ADR-000 |
| AL §1 — 1. Recommendation | REQ | REQ-21, REQ-23, REQ-45, REQ-70, REQ-71, REQ-72, REQ-73, REQ-74, REQ-75, REQ-76, REQ-77, REQ-78, REQ-79, REQ-80, REQ-81, REQ-82, REQ-83 |
| AL §1 Eligible blocking checks — Eligible blocking checks (default profile `guardian`, per-repo activation) | REQ | REQ-70, REQ-71, REQ-72, REQ-73, REQ-74, REQ-75, REQ-76, REQ-77, REQ-78, REQ-21, REQ-79, REQ-80, REQ-81, REQ-82, REQ-83 |
| AL §2 — 2. Report-only at launch | REQ | REQ-79, REQ-80, REQ-81, REQ-82, REQ-83 |
| AL §3 — 3. Start in `report`; graduation | REQ | REQ-22, REQ-21, REQ-23, ADR-003, REQ-78, REQ-48, ADR-000 |
| AL §4 — 4. Promotion rule verdict: **replace** | REQ | REQ-19, REQ-22, REQ-23, REQ-84, REQ-85, ADR-000 |
| AL §5 — 5. Assumptions | REQ | REQ-22, REQ-70, ADR-000, REQ-23, REQ-81, REQ-183 |
| AL introduction / challenger history | ADR | ADR-000 |
| AL §1 disagreement / wildcard rejection | REQ | REQ-21, REQ-70, REQ-71, REQ-72, REQ-73, REQ-74, REQ-75, REQ-76 |
| AL §1 prerequisites | REQ | REQ-45 |
| AL §1 escape hatch | REQ | REQ-23, REQ-77 |
| AL §1 demoted seed checks | REQ | REQ-21, REQ-79, REQ-80, REQ-81, REQ-82, REQ-83 |
| AL §3 old 20-run window, old activation evidence / signals path | dropped | superseded by REQ-22 and REQ-78 (14 days AND 5 PR runs), and REQ-48 (`state/v1/would_block.jsonl`); authority: ADR-000, ADR-003 |
| AL §4 old fixed-50 promotion arithmetic / objections (a)–(c) | ADR | ADR-000 |
| AL §4 Track A | REQ | REQ-19, REQ-22, REQ-84, REQ-85 |
| AL §4 Track B | REQ | REQ-19, REQ-22, REQ-84, REQ-85 |
| AL §4 Corpus-level audit | REQ | REQ-19, REQ-22, REQ-84, REQ-85 |
| AL §5 unresolved authority/provisional questions | dropped | decided: maintainer-only REQ-23 and provisional rare-security REQ-22; ADR-000 |
| AL §5 working-tree/test-execution statements | dropped | review-session observations, not product behavior; review transcript/history |
| req §1 — 1. Product | REQ | REQ-1, REQ-2, REQ-3, REQ-4 |
| req §2 — 2. Ledger: issues, identity, dispositions, evidence | REQ | REQ-5, REQ-6, REQ-7, REQ-8, REQ-9, REQ-10, REQ-11, REQ-44, REQ-46, REQ-47 |
| req §3 — 3. Feature mapping | REQ | REQ-12, REQ-13, REQ-14, REQ-15, REQ-57, REQ-52, REQ-53, REQ-54, REQ-55, REQ-56 |
| req §4 — 4. Analyzer lanes | REQ | REQ-16, REQ-17, REQ-18, REQ-19, REQ-62, REQ-63, REQ-64, REQ-65, REQ-66, REQ-67, REQ-68, REQ-69, REQ-84 |
| req §5 — 5. CI guardian contract | REQ | REQ-20, REQ-21, REQ-22, REQ-23, REQ-24, REQ-25, REQ-26, REQ-27, REQ-70, REQ-71, REQ-72, REQ-73, REQ-74, REQ-75, REQ-76, REQ-77, REQ-78, REQ-79, REQ-80, REQ-81, REQ-82, REQ-83, REQ-84, REQ-161, REQ-162, REQ-163, REQ-48, REQ-49, REQ-50, REQ-51 |
| req §6 — 6. Placement of judgment | REQ | REQ-28, REQ-29, REQ-30, REQ-31, REQ-32, REQ-139, REQ-140, REQ-86, REQ-87, REQ-88, REQ-89, REQ-90, REQ-91, REQ-92, REQ-93, REQ-94, REQ-95, REQ-96, REQ-97, REQ-98, REQ-99, REQ-100, REQ-101, REQ-102, REQ-59, REQ-60, REQ-61, REQ-58, REQ-147, REQ-148, REQ-149, REQ-150 |
| req §7 — 7. Inertia monitoring (descriptive only) | REQ | REQ-33, REQ-34, REQ-35, REQ-36, REQ-103, REQ-104, REQ-105, REQ-106, REQ-107, REQ-108, REQ-109, REQ-110, REQ-111, REQ-112, REQ-113, REQ-114, REQ-115, REQ-116, REQ-117, REQ-118, REQ-119, REQ-120, REQ-121, REQ-122, REQ-123, REQ-124, REQ-125, REQ-126, REQ-127, REQ-128, REQ-129, REQ-132, REQ-133, REQ-134 |
| req §8 — 8. Subject interface for inertia-bench (new; not yet in design 2.1) | REQ | REQ-37, REQ-135, ADR-002 |
| req §9 — 9. Non-functional | REQ | REQ-38, REQ-39, REQ-40, REQ-41, REQ-164, REQ-165, REQ-160, REQ-57, REQ-136, REQ-137, REQ-138 |
| req §10 — 10. Open decisions (not requirements until closed) | REQ | REQ-29, REQ-169, REQ-170, REQ-171, REQ-172, REQ-173, REQ-174, REQ-175, REQ-176, REQ-177, ADR-002, ADR-003 |
| req REQ-1 | REQ | REQ-1 |
| req REQ-2 | REQ | REQ-2 |
| req REQ-3 | REQ | REQ-3 |
| req REQ-4 | REQ | REQ-4 |
| req REQ-5 | REQ | REQ-5, REQ-44 |
| req REQ-6 | REQ | REQ-6 |
| req REQ-7 | REQ | REQ-7, REQ-44 |
| req REQ-8 | REQ | REQ-8, REQ-44 |
| req REQ-9 | REQ | REQ-9 |
| req REQ-10 | REQ | REQ-10, REQ-46, REQ-47 |
| req REQ-11 | REQ | REQ-11 |
| req REQ-12 | REQ | REQ-12, REQ-57 |
| req REQ-13 | REQ | REQ-13, REQ-52, REQ-53, REQ-54, REQ-55, REQ-56 |
| req REQ-14 | REQ | REQ-14 |
| req REQ-15 | REQ | REQ-15 |
| req REQ-16 | REQ | REQ-16, REQ-62, REQ-63 |
| req REQ-17 | REQ | REQ-17, REQ-62, REQ-63, REQ-64, REQ-65, REQ-66, REQ-67, REQ-68, REQ-69 |
| req REQ-18 | REQ | REQ-18 |
| req REQ-19 | REQ | REQ-19, REQ-84 |
| req REQ-20 | REQ | REQ-20 |
| req REQ-21 | REQ | REQ-21, REQ-70, REQ-71, REQ-72, REQ-73, REQ-74, REQ-75, REQ-76, REQ-77, REQ-78, REQ-79, REQ-80, REQ-81, REQ-82, REQ-83 |
| req REQ-22 | REQ | REQ-22, REQ-84, REQ-78 |
| req REQ-23 | REQ | REQ-23 |
| req REQ-24 | REQ | REQ-24 |
| req REQ-25 | REQ | REQ-25, REQ-161, REQ-162, REQ-163 |
| req REQ-26 | REQ | REQ-26, REQ-48, REQ-49, REQ-50, REQ-51 |
| req REQ-27 | REQ | REQ-27, REQ-49, REQ-50 |
| req REQ-28 | REQ | REQ-28, REQ-139, REQ-140 |
| req REQ-29 | REQ | REQ-29, REQ-86, REQ-87, REQ-88, REQ-89, REQ-90, REQ-91, REQ-92, REQ-93, REQ-94, REQ-95, REQ-96, REQ-97, REQ-98, REQ-99, REQ-100, REQ-101, REQ-102 |
| req REQ-30 | REQ | REQ-30 |
| req REQ-31 | REQ | REQ-31, REQ-59, REQ-60, REQ-61, REQ-58, REQ-147, REQ-148, REQ-149, REQ-150 |
| req REQ-32 | REQ | REQ-32 |
| req REQ-33 | REQ | REQ-33, REQ-103, REQ-104, REQ-105, REQ-106, REQ-107, REQ-108, REQ-109, REQ-110, REQ-111, REQ-112, REQ-113, REQ-114, REQ-115, REQ-116, REQ-117, REQ-118, REQ-119, REQ-120, REQ-121, REQ-122, REQ-123, REQ-124, REQ-125, REQ-126, REQ-127 |
| req REQ-34 | REQ | REQ-34, REQ-128, REQ-129 |
| req REQ-35 | REQ | REQ-35, REQ-129 |
| req REQ-36 | REQ | REQ-36, REQ-132, REQ-133, REQ-134 |
| req REQ-37 | REQ | REQ-37, REQ-135 |
| req REQ-38 | REQ | REQ-38, REQ-164, REQ-165 |
| req REQ-39 | REQ | REQ-39, REQ-160 |
| req REQ-40 | REQ | REQ-40, REQ-57 |
| req REQ-41 | REQ | REQ-41, REQ-136, REQ-137, REQ-138 |
| req title / revision/status authority paragraphs | ADR | ADR-009, ADR-008 |
| req §8 exact caps/request/measure schemas, flags, scalar-ID mapping and closed reason set | dropped | not adopted by authoritative design §8.7; REQ-37 adopts principles only and defers exact schema to its own council round (ADR-002) |
| req §10 former open decisions / experiment relocation history | ADR | ADR-002, ADR-003 |
| req Appendix — Removed / superseded heading and provenance note | ADR | ADR-009 |
| req removed Appendix item 1 — Clawpatch **0.4.0** pin, build and `map --source auto` default (§§3, 10, 12.1, 13) | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 2 — "Clawpatch features are the LLM work queue; every feature × category row accountable" (§§3, 6, 12.3, 12.5) | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 3 — Mandatory contextual inspection for completeness of the four families | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 4 — Severity rubrics `security-impact-likelihood@v1` etc. (§12.3) and "persist all low-to-critical findings" as severity-driven design | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 5 — Enaible issue-level parity as the acceptance bar (§§5, 11, 12.2, 12.8; 28-entry corpus incomplete) | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 6 — Seven-language parity (§5) | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 7 — Wide tool selection roadmap (Checkov, OSV, Trivy, CodeQL, Knip, SpotBugs, dependency-cruiser, Lighthouse, EXPLAIN, pprof, etc., §5) | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 8 — Remediation: issue states `pending verification`, `approval`, `ProposedAction`, `ApprovalRef`, authorised-change verification and `scout approve` (§7, §12.6) | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 9 — Gap→rule promotion workflow with explicit approval (§8) | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 10 — "Exit 1 when finding meets severity threshold" gate (§12.7) | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 11 — Durable baseline "bundle from persistent storage"/CI cache and the AWS/OIDC prerequisite (§12.7; `scout-nightly.yml`) | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 12 — Zero-call no-change nightly as a hard contract (§§4, 12.4) | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 13 — Full rescans as the default nightly shape and full-repo `.scout/clawpatch` state attachment options (§§3, 12.1) | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 14 — Research/hypothesis-testing duties (predictive-validity studies of inertia, controlled agent continuations) and "weak-model labelling" | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 15 — Non-Jev OpenRouter/vendor-neutral model use as a product assumption (§10 "Contextual models": generic JSON/HTTP provider; workflow API-key secrets) | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 16 — Sections 11.1–11.2 measured offline corpus statistics (6/9 matches, 28 incomplete comparisons) | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| req removed Appendix item 17 — Section 13 source baselines (Enaible paths, Harness revision, Clawpatch 0.4.0 checkout) | dropped | already removed/superseded source material; surviving product substance in REQ-1–REQ-41 and their supplemental REQs, decision history in ADR-000–ADR-008; inventory/corpus provenance remains in research ledger/evidence/git history |
| Design §8.5 E1b results table, sensitivity and limitations | ADR | ADR-001, ADR-003 |
| Design §4 observed E3/CP07/X2/X2b/development timings and accuracy narratives | ADR | ADR-000, ADR-005, ADR-006, ADR-008 |
| Design §10.1 observation / pre-fix code analysis / complexity inference | ADR | ADR-000, ADR-002 |
| Design §10.2 historical implementation line references and keep/cut rationale | ADR | ADR-000, ADR-002, ADR-003, ADR-007, ADR-008 |
| Design §10.3 / Appendix 6,12 old multi-file contract/archive layout | ADR | ADR-009 |
| Design §12 likelihood/impact ratings and descriptive risk statements | dropped | risk assessments, not product requirements; every behavioral mitigation is mapped individually under §12 R1–R18 |
| Design §14 evidentiary limits / study statistics / historical removed IDs | ADR | ADR-000, ADR-005, ADR-006, ADR-007, ADR-008 |
| Design §15 former numbered decisions | ADR | ADR-002, ADR-003, ADR-000 |
| Design Appendix 20 pending study clause | ADR | ADR-005, ADR-006 |
| Design §2 rejected outcome-first primary framing | ADR | ADR-003 |
| Design §5.3 rejected parity/port/wrapper alternatives | ADR | ADR-000 |
| Design §6 rejected identity/state-change/bundled-routing alternatives | ADR | ADR-000, ADR-002 |
| Design §3.2 Signal | REQ | REQ-8, REQ-33, REQ-44 |
| Design §3.2 Issue | REQ | REQ-5, REQ-7, REQ-44 |
| Design §3.2 Coverage | REQ | REQ-9, REQ-11, REQ-16 |
| Design §3.2 Signal promotion | REQ | REQ-8 |
| Design §6 Adapter guardrails | REQ | REQ-29, REQ-86, REQ-87, REQ-88, REQ-89, REQ-90, REQ-91, REQ-92, REQ-93, REQ-94, REQ-95, REQ-96, REQ-97, REQ-98, REQ-99, REQ-100, REQ-101, REQ-102 |
| Design §6 Rejected | ADR | ADR-000, ADR-002 |
| Design §8.4 Side-by-side panel | REQ | REQ-33, REQ-34 |
| Design §8.4 Adverse-trend flag | REQ | REQ-34, REQ-128 |
| Design §8.4 Inertia watch list | REQ | REQ-34, REQ-128 |
| Design §8.4 Interventions | REQ | REQ-129 |
| Design §8.5 What this does and does not settle | ADR | ADR-001, ADR-003 |
| Design §8.5 E1b | ADR | ADR-001, ADR-003 |
| Design §14 A8 and A10 | ADR | ADR-002 |
| AL table 1 | REQ | REQ-70, REQ-71, REQ-72, REQ-73, REQ-74, REQ-75, REQ-76, REQ-77, REQ-78 |
| AL table 2 | REQ | REQ-79, REQ-80, REQ-81, REQ-82, REQ-83 |
| Historical req §11 | ADR | ADR-000 |
| Historical req §11.1 | ADR | ADR-000 |
| Historical req §11.2 | ADR | ADR-000 |
| Historical req §12 | REQ | REQ-5, REQ-13, REQ-17, REQ-19, REQ-20, REQ-28, REQ-14, REQ-8, REQ-16, REQ-4, REQ-24, REQ-26 |
| Historical req §12.1 | REQ | REQ-13, REQ-14 |
| Historical req §12.2 | REQ | REQ-17, REQ-19 |
| Historical req §12.3 | REQ | REQ-8, REQ-17, REQ-28 |
| Historical req §12.4 | REQ | REQ-16 |
| Historical req §12.5 | REQ | REQ-17, REQ-28 |
| Historical req §12.6 | REQ | REQ-4 |
| Historical req §12.7 | REQ | REQ-20, REQ-24, REQ-26 |
| Historical req §12.8 | REQ | REQ-19 |
| Historical req §13 | REQ | REQ-12, ADR-000 |

## Counts

- Requirements: 183 live rows; 41 existing IDs retained with changed wording (40 preserve meaning; REQ-37 authority-corrected to design §8.7's principles-only scope, disclosed in ADR-009), 142 new (REQ-42–REQ-183), 0 withdrawn; 0 verbatim-kept rows. REQ-42 follows design §13's experiment list (X1–X6, X9–X11) where design §1 still said "X1–X6, X9, X10"; ADR-009 records the choice.
- Source inventory: 389 rows (323 REQ, 44 ADR, 22 dropped).
- Citation index: 312 section/subsection/legacy citation keys, each resolving to REQ/ADR IDs.
- Every design heading, all 23 Appendix items (including 22(a)–(d), 23(a)–(e)), all lane/scorecard/subject numbered items, all R1–R18/A1–A14 register entries (including removed IDs), all allow-list sections and all REQ-1–REQ-41 are represented.
- Council round 1 found six statements lost or weakened in the first draft. Each is now restored in the REQ named: §7.3 report profile "exit 2/3 as neutral check" → REQ-21 (R1-01); §8.5 causal-claim study designs and "repeated dashboard significance never upgrades a claim" → REQ-35 (R1-02); §7.3 / AL §3 rare-security fire-rate proxy → REQ-22 (R1-03); §7.3 "one gate reason per family" → REQ-27 (R1-04); §8.3 over-cap PRs marked `linkage: incomplete` → REQ-127 (R1-05); Appendix 19 backlog B-1–B-4 and "B-3 on the first Choice that gates anything" → REQ-99 (R1-06). Any further statement a review finds unplaced is a defect of this map, not a deliberate drop.
