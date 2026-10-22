"""Versioned Jev question variants for Scout finding triage (route-only)."""
VARIANTS = {
"v1_choice": {"q": {"type": "choice",
  "instructions": "A static analyzer reported the tool_signal on this file. Using only the excerpts, which describes the finding?",
  "criteria": {
    "real_defect": "The excerpt shows the flagged mechanism literally present in shipped, hand-written, non-test code, and nothing visible in the excerpt guards or neutralises it.",
    "false_positive": "The excerpt shows the flagged mechanism is absent, guarded, type-only, or happens in test, fixture, example, documentation or generated files, or the metric does not reflect real maintainability or runtime cost.",
    "cannot_tell": "The excerpt does not contain enough to decide, for example the flagged code is outside the excerpt."}}},
"v2_noul": {"q": {"type": "noul",
  "instructions": "Does the excerpt show the mechanism named in tool_signal actually occurring in shipped, hand-written, non-test code with no visible guard?",
  "criteria": {
    "true": "The flagged construct is literally present in code or configuration that is used for real builds or at runtime, and no guard, type-only import, or test/example/fixture/generated context is visible.",
    "false": "The flagged construct is absent, guarded, type-only, inside a test/example/fixture/documentation/generated file, or the metric (length, duplication, complexity) is high only because of repetitive declarative content."}}},
"v3_choice": {"q": {"type": "choice",
  "instructions": "A static analyzer reported the tool_signal on this file. Decide from the excerpts whether the reported mechanism really exists in code or configuration that is built, run or published, and is worth a human's attention.",
  "criteria": {
    "real_defect": "The flagged mechanism is visibly present and creates a concrete weakness. Examples: a CI workflow references actions by mutable tag or branch instead of a commit SHA; a dependency version range admits releases with known vulnerabilities and the code uses the affected feature.",
    "false_positive": "The finding is an artifact. Examples: the file is a test, example, fixture, snapshot, documentation or generated schema file; a length or duplication metric is high only because the content is repetitive declarative statements; a literal such as a loopback address is only formatted for display or comes from the runtime socket; the tool misreads what the file is.",
    "cannot_tell": "The excerpt does not contain enough to decide, for example the flagged line is outside it."}}},
}
