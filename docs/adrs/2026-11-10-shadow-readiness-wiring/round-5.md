# ADR-015 council review — round 5 (final; zero issues; signed off)

## Scope, provenance and limits
- **Read only.** No edits, builds, tests, formatters, hashing or live provider calls. This is a design-only review. No P23/S6, P24 or P25 implementation proof, provider receipt or deployment/cutover claim is made.
- **Worktree:** `/Users/adamjackson/Projects/work/scout-adr015`. Folder: `docs/adrs/2026-11-10-shadow-readiness-wiring`.
- **Reviewed:**
  - `round-4.md` and `round-4-response.md`.
  - The current `ADR.md`, with W6 ¶1 at `:58` read raw, against the R4-02 fix text proposed at `round-4.md:57`.
  - `ci-integration.md:59-103`, with the pinned table at `:75-82` read raw.
  - In `wiring-contract.json`: `operation_exits.reconcile` (`:248-253`) and `implementation_proofs` (`:267-273`).
  - In `scout-design.html`: REQ-25 at `:696`, REQ-161 at `:707` and the card at `:1053-1061`.
  - A search of all 10 signed files for `triage/start | First land | before workflow merge | prepare-auth | SHADOW_OPERATION | action_id_encoding`.
- **Reused from round 4:** the settled route-B findings were reused, not re-explored. These are R3-01 to R3-05 resolved and parity with the actual workflows.
- **Digests:** the council computed none. The values below are Main's, from `local://adr015-round5-digests.json`.
  - Main computed them with both hashlib and `shasum -a 256`, and the two agreed on all 10 files.
  - Main matched the exact bytes again after the pre-commit hook passed at `aa518f87`. Main reports that no frozen file changed after that.
  - Neither the council nor the peers saw any drift between reads.
- **Manifest coherence:** compared with the round-4 table, only two hashes changed:
  - `ADR.md`: `833a249d…` → `30901df3…`
  - `ci-integration.md`: `e636e4da…` → `f1372ace…`

  The other 8 are identical, including `wiring-contract.json` (`6a75a5ea…`). That matches the declared round-5 scope exactly.

## Round-4 issues
| ID | Status | Evidence |
|---|---|---|
| ADR015-R4-01 (Low) | **Resolved** | `ci-integration.md:79` now ends "…shadow reconcile 2 is always neutral; shadow 3 consults this flag." `:87` now ends "…reconcile 2 is always neutral, 3 warns/fails with that flag." Neither line mentions triage/start any more. Both match contract `operation_exits.reconcile`: exit 2 is neutral regardless of fail-on-incomplete, and exit 3 warns and fails only when it is set. The pinned table at `:75-82` is still aligned: every Effect cell is padded to the separator width, and the other rows are unchanged. Every remaining `triage/start` hit in the signed set is deferral wording: ADR `:10,16,50,52`; ci `:101,107`; cli `:136,142`; policy `:86`; STATUS `:13`; README `:37`; plan `:94,96`; R-109; REQ-161; card `:1054`; contract `:272`. |
| ADR015-R4-02 (Low) | **Resolved** | W6 ¶1 (`ADR.md:58`) now reads: "Merge caller/reusable permission grants and diagnostic workflow wiring together **only after** restore, per-step credential and Linux/macOS S6 process/filesystem proofs pass; no permission-widening workflow merge may precede them." This matches W2 ("before implementation merges"), REQ-25 `:696` ("S6 … must pass before workflow implementation merges"), contract `implementation_proofs[0]` ("before workflow merge"), STATUS `:15`, plan `:94` and card `:1059`. W6 also requires restore and per-step credential proofs, which the contract's S6 entry does not list. That is not a disagreement: round 4's proposed fix included them and said no contract change was needed. W6 says "pair a compatible released binary" where round 4 proposed "first ship". That is still consistent. An old binary reports unavailable/3, which is neutral unless fail-on-incomplete is set (W1, W6, ci `:103`), so when the binary ships cannot widen the token before S6. |

## New issues
None. No other text in the signed companion files grants action or control authority. The following all read as round 4 settled them:
- route B
- `durable_write_types: []`
- route A as a known option that is not adopted
- REQ-48/160/162/163/168 unchanged
- existing vetoes, triage and lineage preserved in the same lineage

## Peers (real, independently spawned, finished before this verdict)
- **`innovation-challenger`, `ADR015Wiring.ADR015R5Adversarial`: 0 issues; recommends approval.**
  - R4-01 is resolved, with no remaining conflict over action authority in the signed companion files.
  - R4-02 is resolved and consistent with W2, REQ-25/161 and `implementation_proofs`.
  - The settled route-B boundaries are coherent.
  - It recomputed no hashes and claimed no live proofs.
- **`experiment-peer`, `ADR015Wiring.ADR015R5Integration`: 0 issues; recommends sign-off on the exact bytes.**
  - It confirmed independently that only `ADR.md` and `ci-integration.md` changed against round 4.
  - It confirmed that both ci clauses are removed and the table is still aligned.
  - It confirmed that W6 agrees with W2, REQ-25, the contract and the trackers.
  - It found the "pair" versus "first ship" wording consistent.
  - It stated one limit: it had no git diff, so "only W6 ¶1 changed in `ADR.md`" rests on the scope statement. The sections it did read match round 4.
  - It recomputed no hashes and claimed no live proofs.
- **Peer influence:** neither peer changed the lead's verdict.
  - The lead reached zero issues independently, including the "pair" wording analysis and the table check, before reading either peer's output.
  - Both peers then confirmed that result. The integration peer's own manifest comparison agreed with the lead's.
  - The integration peer's missing git diff is not a blocker, because the sign-off binds the exact current bytes reviewed, not a claim about what changed.

## Decision memo
1. **Recommendation:** adopt ADR-015 route B as signed below. It covers read-only diagnostic inventory plus credential and restore hardening, as a design only.
2. **Alternatives:**
   - (a) Route A publication-envelope controls remain a known viable option, not adopted. Adopting them needs a later ADR.
   - (b) Rewording W6 to "first ship" is rejected. It adds no safety, because old binaries fall back to unavailable/3.
   - (c) Requiring live proofs before design adoption is rejected. Those proofs are implementation gates (W2, W6, `implementation_proofs`), not adoption gates.
3. **Evidence:**
   - Signed set: `ADR.md:58`; `ci-integration.md:75-82,87,101-107`; `wiring-contract.json:248-253,267-273`; `scout-design.html:696,707,1053-1061`.
   - Round 4: `round-4.md:57`, and the round-5 manifest compared with the round-4 table.
   - Peer records: `agent://ADR015Wiring.ADR015R5Adversarial` and `agent://ADR015Wiring.ADR015R5Integration`.
4. **Risks:**
   - Requirement-row hashes have not been re-verified against `baseline_commit` by the council. They rest on Main's computation and the P25 implementation checks.
   - The council did not recompute any file digests.
   - The S6, restore, credential, provider and P25 proofs are still pending. The workflow implementation must not merge before they pass.
   - ADR-012 readiness remains incomplete.
   - The decision can be reversed by a superseding ADR, and owner review under R-064 is still pending.
5. **Execution outline:**
   1. Commit this round as the folder's highest `round-N.md`, with the block below, without changing any signed bytes.
   2. P25 then locates the commit that adds this round and hashes the signed blobs at that commit.
   3. Implementation follows W6: release the binary; pass the restore, credential and S6 proofs; then merge the grants and the wiring together in one gated merge.

## Verdict
**0 open issues. Final design sign-off for ADR-015** on exactly the bytes below. The digests are Main's: hashlib and `shasum -a 256` agreed on all 10, and they matched again after the hook passed at `aa518f87`. The council did not recompute them.

| Path | SHA-256 |
|---|---|
| docs/STATUS.md | 21a06fa4265a106d85b645a22efebce2155feac3610ef4da69594cff5a69ba65 |
| docs/adrs/2026-11-10-shadow-readiness-wiring/ADR.md | 30901df3d76432a62b71e388221bf7b7316c16af7ba487a0d27f57218fb28553 |
| docs/adrs/2026-11-10-shadow-readiness-wiring/wiring-contract.json | 6a75a5eaaa720c3c1220a948c2dedea60b1b6a1381e50c5ccbe51c7c368119e0 |
| docs/adrs/README.md | 4a7affe5f8a711dca78cef75e90b384f9cd25adb758c55cd43683f45c6e172ff |
| docs/plans/2026-11-06-remaining-work.md | 8281afc3b908b1a9c4c16aca84a5ad5cb7bafad845e16d5bfdce4253a6ffe69d |
| docs/references/ci-integration.md | f1372ace5782081cf5e99db10742340bfbc2b30280dc0bdaa01f7f6d93826388 |
| docs/references/cli-commands.md | 27e6f033508be7eb377893bf063b5edcd53d61fc275f1fe8fb2ec2483dde79f5 |
| docs/references/policy-file.md | 964c813eaca641b46c69b3a54049240046274e6427395dbb2545a8598bea2222 |
| docs/research-ledger.md | a6fe13d920e2589ced6e32834f86cd4142af1c852caf635b8ea08f528aee2904 |
| docs/scout-design.html | 9a877aaa37b8083892195af8d217743340e79aba6de4c210f75015069e0eb8ce |

```json
{
  "schema": "scout-signoff/v1",
  "adr": "ADR-015",
  "folder": "docs/adrs/2026-11-10-shadow-readiness-wiring",
  "round": 5,
  "open_issues": 0,
  "verdict": "approved",
  "files": [
    {"path": "docs/STATUS.md", "sha256": "21a06fa4265a106d85b645a22efebce2155feac3610ef4da69594cff5a69ba65"},
    {"path": "docs/adrs/2026-11-10-shadow-readiness-wiring/ADR.md", "sha256": "30901df3d76432a62b71e388221bf7b7316c16af7ba487a0d27f57218fb28553"},
    {"path": "docs/adrs/2026-11-10-shadow-readiness-wiring/wiring-contract.json", "sha256": "6a75a5eaaa720c3c1220a948c2dedea60b1b6a1381e50c5ccbe51c7c368119e0"},
    {"path": "docs/adrs/README.md", "sha256": "4a7affe5f8a711dca78cef75e90b384f9cd25adb758c55cd43683f45c6e172ff"},
    {"path": "docs/plans/2026-11-06-remaining-work.md", "sha256": "8281afc3b908b1a9c4c16aca84a5ad5cb7bafad845e16d5bfdce4253a6ffe69d"},
    {"path": "docs/references/ci-integration.md", "sha256": "f1372ace5782081cf5e99db10742340bfbc2b30280dc0bdaa01f7f6d93826388"},
    {"path": "docs/references/cli-commands.md", "sha256": "27e6f033508be7eb377893bf063b5edcd53d61fc275f1fe8fb2ec2483dde79f5"},
    {"path": "docs/references/policy-file.md", "sha256": "964c813eaca641b46c69b3a54049240046274e6427395dbb2545a8598bea2222"},
    {"path": "docs/research-ledger.md", "sha256": "a6fe13d920e2589ced6e32834f86cd4142af1c852caf635b8ea08f528aee2904"},
    {"path": "docs/scout-design.html", "sha256": "9a877aaa37b8083892195af8d217743340e79aba6de4c210f75015069e0eb8ce"}
  ]
}
```

This sign-off covers the design only. It claims no implementation, S6/P23/P24/P25 proof, provider receipt, deployment or ADR-012 readiness.
