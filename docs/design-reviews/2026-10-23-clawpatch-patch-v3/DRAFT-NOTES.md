# Draft notes — revision 2.4 proposal

Round-5 draft (round-4 issue fixed, `round-4-response.md`; round 3 in `round-3-response.md`; round 2 in `round-2-response.md`). Not yet signed off. Line numbers refer to `scout-design.proposed.md`.

| File | Lines | SHA-256 |
|---|---|---|
| `scout-design.base.md` (byte copy of `docs/scout-design.md`, rev 2.3) | 466 | `75ef1e66521b1f129e872f651bd0f8bb9989e38867fb5335ec5ad43158fb3061` |
| `scout-design.proposed.md` (rev 2.4) | 468 | `5c16a807513d54667390278755b8daa504e8db5ae0733236f87938c51a9a90ae` |
| `proposed.diff` | 45 | 7 insertions (6 text, 1 blank), 5 deletions; hunks `-1,12 +1,14`, `-94,14 +96,14`, `-451,7 +453,7`; `101f0af4fd1d48cdb70c2f1139071768bd77fa12be283c09f52648e1b00f8759` |
| `PROPOSAL.md` | — | `7b298dbc782b2da6ae8a6c30a8af2f49af563010cf7644c1d035c9da9e923d1a` |

(`proposed.diff` content: header line 3 replaced (revision bookkeeping and dogfood pin); Revision 2.4 paragraph and its blank line inserted; evidence-notation line replaced; §4 map-refresh line replaced; §4 pin line replaced; Appendix 7 replaced. Counted from the diff: -5 lines, +7 lines (6 text, 1 blank).)

Digests via `shasum -a 256`. `proposed.diff` was generated with `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md > proposed.diff` (BSD diff at `/usr/bin/diff`; no timestamps in the labels). `docs/scout-design.md` digest equals the base digest at draft time.

No `requirements.proposed.diff`: `SCOUT_REBUILD_REQUIREMENTS.md` does not state the patch pin (no match for `sanitize`, `f39d73da`, `HOME`, `USERPROFILE`).

## Where each change is (proposed file)

- Line 3: status line, `**revision 2.4 (2026-10-23)**` (rev 2.3 no longer bold) (P1); dogfood pin now `v0.1.4` (P5).
- Line 9: Revision 2.4 paragraph (P2; names the two drift refreshes, no failure-mode claim).
- Line 11: evidence notation, "2.1, 2.2, 2.3 or 2.4" (P7).
- Line 99: §4 Map refresh, `defaultMapTimeoutMS`, `internal/clawpatch/client.go:23` (base 97). (P6)
- Line 106: §4 Pin bullet (base 104): `client.go:20-21`, pnpm cite `toolchain/versions.env:10`. (P3)
- Line 456: Appendix item 7 (base 454). (P4)

Every base line from the evidence-notation line (base line 9) onward shifts by +2 (the new paragraph and its blank line at proposed 9–10).

## Judgement calls for the council

1. Host-auth environment is not stated in the design, so no edit; recorded in PROPOSAL.md and the Revision 2.4 paragraph only (`USER`/`LOGNAME` named as the v3 addition; `scripts/clawpatch-safe.patch:332-336` verified).
2. The Revision 2.4 paragraph is added to follow the header convention (one paragraph per revision); it makes no claim beyond the pin, the v3 change and the two drift refreshes (P5, P6).
3. The old line cites (`review.go:19-20`, `scripts/build-clawpatch.sh:5-9`) were stale; replaced with the current `review.go:20-21` and `build-clawpatch.sh:18-19`, and `release.yml:89` added because the release manifest carries the digest.
4. CP07 §7 is not edited (historical port record).

## Checks for the main agent

`shasum -a 256 docs/scout-design.md docs/design-reviews/2026-10-23-clawpatch-patch-v3/scout-design.base.md` must both equal `75ef1e66…`; `/usr/bin/diff -u --label scout-design.base.md --label scout-design.proposed.md scout-design.base.md scout-design.proposed.md` must reproduce `proposed.diff` byte for byte; the council's final round must name the proposed file's SHA-256 (any later edit changes it).
