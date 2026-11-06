# Round 2 response — revision 2.6 (agent map at onboarding and manual `scout map` only)

All 10 issues of `round-2.md` are accepted and fixed; the council's replacement text was used as given. Line numbers are in `scout-design.proposed.md` (477 lines, unchanged count). `proposed.diff` and `requirements.proposed.diff` were regenerated with `/usr/bin/diff -u` and verified. Not signed off: round 3 re-reviews.

| # | Sev | Fix | Where |
|---|---|---|---|
| 1 | minor | "§§3.1" → "§§1" in the Revision 2.6 paragraph; "§3.1 Mapping" → "§1 Mapping" in the PROPOSAL row (base 23), D2, D4, R-043 Informed and the DRAFT-NOTES "Line 25" bullet. | design 13; PROPOSAL; DRAFT-NOTES |
| 2 | minor | §4 (a)/(b) replaced with the council's text (`skipped` decided before the egress check, `client.go:203-213`; `--allow-egress` the only authorization; documented caller authorizes only from `workflow_dispatch`; a manual map also records `LastAgentRefreshAt`). PROPOSAL D5 "Now" and JC1 (both sentences) use the council's text. | design 103; PROPOSAL D5, JC1 |
| 3 | minor | `requirements.proposed.diff` regenerated: line 1 "aligned to design revision 2.6"; line 3 both substitutions, naming design digest `a6dc122b…ab84` and `round-3.md`; REQ-13 unchanged. PROPOSAL (c) replaced with the council's paragraph; REQ-13 row notes the header lines; DRAFT-NOTES row recounted (2 hunks, 3 insertions, 3 deletions) with the new patched-result digest. **If round 3 changes the design file, this response's successor must recompute the line-3 digest and round number.** | requirements diff; PROPOSAL (c); DRAFT-NOTES |
| 4 | minor | R15: "until the pin changes (a manual `scout map --refresh` uses the same pin and fails the same way) (§4)". D13 did not quote the clause (no change). | design 408 |
| 5 | nit | Revision 2.6 paragraph item (3): `R15 says "agent map run" and records the onboarding re-attempts (§4);`. | design 13 |
| 6 | nit | Phase 4 scenario: "(every step runs with `codex` on `PATH` and egress authorized, so no absent invocation comes from a `skipped` or `blocked` attempt)"; D12 mirrors. | design 377; PROPOSAL D12 |
| 7 | nit | `refresh.go:146-151` → `:147-152`; workflow `:171-173` → `:170-172` (JC1 and the `refresh.go` `report` row; the `:171-173` row that cites `refresh.go:171-173` is a different file reference and unchanged); `refresh.go:97-98` → `:98-99`. All re-checked in the working tree. | PROPOSAL JC1, Companion (b) |
| 8 | nit | STATUS list gains line 161 and line 166 with the council's wording. | PROPOSAL Companion (a) |
| 9 | nit | (a) "skipped/blocked/incomplete"; (b) `refresh_test.go:85` row asserts second push and nightly are `none` with `mapper.calls` at 1; (c) cutover bumps both pins in `scout-nightly.yml` (`:40`, `:49`) and both in `ci-integration.md` (`:51`, `:60`). | PROPOSAL Companion (b) |
| 10 | nit | DRAFT-NOTES line-3 bullet lists both Phase 4 qualifiers. | DRAFT-NOTES |

Also updated: PROPOSAL status line (round-3 draft); DRAFT-NOTES digests, the line 13, 103, 377 and 408 bullets, requirements generation directory (`/tmp/reqdiff4`), and a note that `DRAFT-NOTES.md`'s own digest lives here.

Code claims re-verified in the working tree: `client.go:203-206` (`MapSkipped` on `ErrNoCLI`), `:210-213` (egress `MapBlocked`), `:207-216` (blocked before invocation); `refresh.go:129`, `:147-152`, `:163`, `:98-99`; `scout-guardian-trusted.yml:170-172`; `scout-nightly.yml:40,49,52`; `ci-integration.md:51,60,63`.

## Digests (recomputed, `shasum -a 256`)

| File | Lines | SHA-256 |
|---|---|---|
| `scout-design.base.md` (= `docs/scout-design.md`, rev 2.5) | 473 | `032285b35e18f027aee0503af0890009fe13b45a36b56185586c5d075275acf9` |
| `scout-design.proposed.md` | 477 | `a6dc122b520239ea514fcaff0cb3d0ce2a4804a4e8fd0bf85ea688a01b37ab84` |
| `proposed.diff` (27 insertions, 23 deletions, 14 hunks) | 159 | `70ed6094ce1799c4711123a33465b25dccfeeec800e04539259ba426ccfbafeb` |
| `SCOUT_REBUILD_REQUIREMENTS.md` (base, unchanged) | | `9efcea1b0248d43d0ab831f0924f055f16bb39a3e008365d1e64c7d68a6ece2d` |
| `requirements.proposed.diff` (patched result `d689992f0f8aaa8784fb2227596a8ef9c089ded846c2dfc49be1aedb4127bf90`) | 20 | `313ba2d58021178ece4e068acac6e7969c5d860c621732dac44152afcea32590` |
| `PROPOSAL.md` | 179 | `ad31bb1760210a14c9937cbf2a2de3855da1d1f535b1de5a3c41b63abd45f10d` |
| `DRAFT-NOTES.md` | 63 | `cbf718d6bd2c333909cd604249c6993983c963d722c53c84ffe24f4a557f1b3c` |

Verification by command: `/usr/bin/diff -u --label … scout-design.base.md scout-design.proposed.md | /usr/bin/cmp - proposed.diff` matched; `/usr/bin/patch -p1 --dry-run -i requirements.proposed.diff` from the repository root applied to `SCOUT_REBUILD_REQUIREMENTS.md` (digest `9efcea1b…`, unchanged).

Not changed: `scout-design.base.md`, `round-*.md` council files, `docs/scout-design.md`, `SCOUT_REBUILD_REQUIREMENTS.md`, the ledgers, `docs/STATUS.md`, code, `.codegraph`.
