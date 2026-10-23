# Round 3 response — design revision 2.3

All six round-3 issues fixed, text-only (no code change, `docs/scout-design.md` untouched). Line numbers refer to `scout-design.proposed.md` (still 466 lines; every edit is in place).

| # | Sev | Fix |
|---|---|---|
| 1 | minor | Qualifier added at `:186`, `:7` item (4), `:360` and REQ-27: the per-finding rule applies to a finding that would otherwise be `introduced` and is not a `guardrails.*` check; it is `unknown` instead of `introduced`; `pre-existing` and `regressed` are unaffected. The rule is not widened to `regressed` (matches `gate.go:282`). PROPOSAL D4 mirrored. |
| 2 | minor | Option (a). `:372` and PROPOSAL D6 now report survival and one distribution of the surviving areas' Jaccard in [0.5, 1] (matched ≥ 0.5; kept contribute 1.0); the harness does not separate them because `scout map` discards the report (`cmd/scout/map.go:102`); X2's protocol separates matched from kept. DRAFT-NOTES corrected. No STATUS gap needed. |
| 3 | nit | `:7` item (4) now counts runs that are `comparison: none` or carry such an `unknown`; item (5) adds the CI toolchain provisioning (§7.4, `404f69b`, REQ-25). |
| 4 | nit | "shipped" replaced by "on `main`" with commits and release. Ancestry confirmed with `git merge-base --is-ancestor`: `480949a`, `404f69b` and `71dbf25` are in `v0.1.2` and `480949a`, `404f69b` are not in `v0.1.1`. The file parenthetical moved to the per-finding sentence at `:186`; the counting-unit sentence now says `ShadowStats` has no producer yet and `comparison` is not persisted per run. `:197`, `:7` item (4), PROPOSAL D4/D8 and DRAFT-NOTES adjusted. |
| 5 | nit | Qualified, not re-run: `:3` says the Phase 2 scenario ran before `480949a` and the per-finding base-gap clause is covered by `base_gaps_test.go`, not yet by the scenario. |
| 6 | nit | PROPOSAL X bullet now "where a point minimum applies (X3, X10, X5's point values; X1's gain and X4/X5 lower bounds have no default substitute)"; "§14 other than A2" (D4); DRAFT-NOTES adds `:442` and A2 (§14). |

## New facts reflected only where the text already touched them
- `v0.1.2` published and pinned by dogfood; CI scanner provisioning succeeded on a real runner (run 39966674494); `pr_effort` in CI still not observed. Edited in `:3`, `:197`, PROPOSAL D8 rows, `docs/STATUS.md` §1 (workflow assumptions, Published, Phase 3 `pr_effort` notes and §5 row 3).
- Gitleaks global-allowlist paths are `not_applicable` (`71dbf25`, the commit the task cites as `eda77f9`^): recorded in STATUS §1 only (new bullet after "Published"); the design does not touch Gitleaks coverage, so it is not added there.

## Reproduction and digests (`shasum -a 256`)
`proposed.diff` is `diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md` (236 lines, 42 insertions, 38 deletions; no timestamps). `patch --dry-run -p1 < requirements.proposed.diff` applies cleanly.

| File | SHA-256 |
|---|---|
| `scout-design.base.md` | `f81b26014e6d2e38a741249819d23fcb993541e2194f7b168c4970c3b415ded4` |
| `scout-design.proposed.md` | `5f3f02f4d29b166b9ebca1a1a3ed6e97d03c98bcc26a09c2606dd6bbd7be84fd` |
| `proposed.diff` | `50cda388e4ddfb564d0389894bb389ed5489eb0c54f0283c068699b119dbfcdf` |
| `requirements.proposed.diff` | `4e7adaccaa3ffe0e627fdd06c068a8469afbefc5a6dd09d985ca73803ab927c3` |
| `PROPOSAL.md` | `546407eef3e4e8119d987bb4587704e7cf7e592490245e3fe78b450fe359f5e1` |
| `DRAFT-NOTES.md` | `0a365bfecbb1a1d179eac4adc5ae9ce9094ebb984e1cac201e8ab3a42ca5c122` |
| `docs/scout-design.md` (unchanged, = base) | `f81b26014e6d2e38a741249819d23fcb993541e2194f7b168c4970c3b415ded4` |

`DRAFT-NOTES.md` records the other digests; this file's own digest is not self-recorded.
