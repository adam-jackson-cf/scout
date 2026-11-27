# Council round 9: X11a, main-branch compatibility edit to the signed `frontend/main_test.go` call site and the product `association.Run` signature

| Field | Value |
|---|---|
| Scope | Main's copy of `evaluation/x11-merge-association/frontend/main_test.go:181`, plus the product change from `Run(ctx, repo, base, head, own, decide)` to `Run(ctx, RunInput{…})` in `internal/association/orchestrate.go`, `lane.go:52` and `orchestrate_test.go` (3 call sites). |
| Owner decision | 2026-11-27: change the signature rather than add a lint exclusion for the six-argument `Run` (the repository gate's revive argument-limit is 5). This reverses d13e5d38, which had restored the six-argument signature to protect this file. |
| Base | Branch `agent/repo-proofs` at HEAD `365bb588`, with uncommitted changes in exactly these 4 files. |
| Authority | Round 8. The signed set is the 1,978 files at `scout-x11` absolute paths in `reviews/rev8-digests.json` (SHA-256 056ddf7f428364230020800e2e4c36d4f8d2a2643640c91664b93686ba9cd135). |
| Date | 2026-11-27 |
| Reviewer | innovation-council lead (anthropic/claude-opus-5-5), read-only; it recomputed every digest with `shasum` and `git show`. |
| Independent peers | innovation-challenger (OpenAI Codex): no issues. experiment-peer (anthropic/claude-opus-5-5): I1 and I2 (minor) plus a doc-comment note, all addressed in packet revision 2. Neither was told the lead's view. |
| Packet | Outside the repository at `/Users/adamjackson/Projects/work/scout-proofs-candidate/council-x11-r9/`: `packet.md`, `packet-r2.md`, `challenger.md`, `peer.md` (the experiment-peer text was saved by the lead after the round). |

## 1. Status of prior items

| ID | Status | Evidence |
|---|---|---|
| R1-01 | **Open (carried; deferred as planned)** | Not touched by this scope. The implementation-round machinery is still absent and `stages.py` still refuses later stages. |
| R9-I1 (peer, minor): wrong "before" baseline | **Closed** | Revision 2 separates the pinned, `57e14923`, HEAD and after digests; the lead recomputed all of them (§5). |
| R9-I2 (peer, minor): lane.go and orchestrate_test.go digests missing | **Closed** | Both are in revision 2 and match the lead's recomputation. |
| Peer doc-comment note | **Closed** | The `Run` doc comment sits directly above `func Run` again; `RunInput` sits above it (`orchestrate.go:31–45`). |

## 2. New issues

**None.** No critical, major or minor issue was found.

**Considered and not raised:**

- **Arguments are equivalent.** At `main_test.go:181` the keyed fields match the old positions: `Repo` is `association.Repo{Dir: req.Repo}`; `Base` is `m.Parent` and `Head` is `m.SHA` (no swap; a swap would also fail the `len(results) != 1` check at :182); `Own` is the same `expectedOwnership`; `Decide` is omitted, so its zero value `nil` equals the old literal `nil`. `Run` unpacks the fields unchanged (`orchestrate.go:46–47`), and the guard `decide != nil && len(mr.Decided) > 0` is unchanged. The Counts, Selected and Capped assertions are byte-identical.
- **The study binary cannot be affected.** `frontend/main.go` never calls `association.Run`; it uses only `Repo.Merges`, `Classify`, `BuildClusters`, `Cap` and `SiblingKeyOf` (`main.go:66–147`).
- **Authority is intact.** The manifest, the frontend build recipe and the v7 Go closure all resolve `internal/association` and the frontend under `scout-x11`, whose bytes are unchanged (§5). Main's copy is a non-execution copy; [INFERENCE] under experiment rule 5 it still needs this round, because it is the same signed file at a repository path.
- **Main had already drifted before this change.** HEAD `orchestrate.go` (5dcbb04e…) already differed from the pinned 48d958b8… through commits 58601274, d13e5d38, b0876464, a515f88b and 2222f070, including the split of `overlay.apply` into `dropChanged`/`decide`, which the challenger and lead read as equivalent. Main's parity results are therefore not X11a evidence (H1).
- **gofmt.** `frontend/main.go` and `main_test.go` are unformatted in pinned `scout-x11` too (main.go:117–118, main_test.go:301); the new line 181 is not among the gofmt diffs.
- **Pre-existing test failure.** `TestPreparationOwnershipMatchesProductLoader` fails on main with a KeyError, with and without this change, because its pins are keyed by `scout-x11` absolute paths; it passed in `scout-x11` per round 8. The other four frontend tests (including `TestJSONParityWithProduct`, which holds the edited call) pass on main.
- **No RESULTS.md deviation is needed.** No runtime departure occurs, and no RESULTS.md exists for X11a yet.

## 3. Conditions

- **H1:** a passing `TestJSONParityWithProduct` on main shows only that main is consistent with itself. X11a evidence comes only from the pinned `scout-x11` closure.
- **H2:** replay must use a single product commit and show it is equivalent to the preparation closure (the existing R1-01 obligation, restated).
- **H3:** `RunInput`'s fields are exactly Repo, Base, Head, Own and Decide. A later field reaches the frozen call through its zero value; a field whose zero value is not neutral needs a new council round before it lands on main.
- **H4:** this verdict is limited to main-copy compatibility and grants no inventory, replay, label or model authority.
- **H5:** the `scout-x11` bytes stay unchanged, and main's changed files are never copied into or substituted in the frontend build or ownership closure.

## 4. Sign-off statement

**X11a round 9, main-branch compatibility scope: zero issues; signed off under conditions H1–H5.** Overall X11a status is unchanged: R1-01 stays open and deferred, and HD8 stays `proposed`.

| File | Main (after) SHA-256 | `scout-x11` pinned SHA-256 |
|---|---|---|
| `evaluation/x11-merge-association/frontend/main_test.go` | `20551692cd73ee754f5c37303ca96c45f2978152d1ca1019120bafb03d527d1e` | `1b49b60b85568ef8d3b440d12590335e5a9e17eb7b9997514af9827b833cf3c0` |
| `internal/association/orchestrate.go` | `86bac326d4fc154f8ba54c78830446ff3f25591f775aec6d5e7ac8e12ba7a361` | `48d958b8a389628b51b34ba7e5914225b6125e1679ca5ef312a808e3acd84a2d` |
| `internal/association/lane.go` | `f508a34a09679c7b2e609c7a8c36075593a31a0ade7047e102f888958d7dfc96` | `8f54f8d361f65b74b061df79d96cd99e5cd728e7faf000a2eb68bd5d1946f334` |
| `internal/association/orchestrate_test.go` | `6fdd4dfbc2673329ce4538f0a3f88a117ce1359cc3b9d045aa68228c6a8bab81` | not pinned |
| `PREREGISTRATION.md` (unchanged) | `9b4be6ef008c057e7891865fafeb1fbf6b0246dde0b7a0ab6552e9312378a082` | same |
| `reviews/rev8-digests.json` (unchanged) | `056ddf7f428364230020800e2e4c36d4f8d2a2643640c91664b93686ba9cd135` | same |

Revision 1 cited `orchestrate.go` as `0ffb7567…`; it is superseded by `86bac326…` after the revision-2 doc-comment reorder.
