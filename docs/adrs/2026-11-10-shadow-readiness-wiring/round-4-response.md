# ADR-015 round-4 response

No sign-off is claimed here. Round 4 stayed frozen and returned two Low issues; its record is preserved verbatim. Route B and all other resolved R3 boundaries remain unchanged.

| Issue | Fix |
|---|---|
| ADR015-R4-01 | Remove both stale triage/start clauses from ci-integration.md's fail-on-incomplete row and exit-handling bullet. Those clauses now describe only adopted reconcile behavior; future action exits remain unadopted. |
| ADR015-R4-02 | W6 now explicitly requires caller/reusable grants and diagnostic wiring in the same gated workflow merge only after restore, per-step credential and Linux/macOS S6 process/filesystem proofs pass. No permission-widening merge may precede those proofs. |

Only the ADR rollout paragraph and CI-reference clauses change in the signed set (plus the CI table's required pinned formatting). Round 5 must review the fresh full-set digests and these exact fixes before any adoption. No source, transport, control format, write, implementation proof or readiness acceptance is added.
