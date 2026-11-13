# Onboarding policy

## Sub-features

- **Non-interactive setup.** `scout init --yes` writes `.scout/policy.yaml` with profile `report` and the consents given by flags; `policy.Load` reads the file back to the same values.
- **Safe defaults.** `scout init` with no flags records profile `report` and no consent.
- **Idempotent.** Running `init` again keeps existing activations under `blocking` and existing comments.
- **Consistency checks.** Inconsistent consent flags are rejected with exit `3` and nothing is written.
- **Interactive answers.** Interactive onboarding records the answers given at the prompts.
- **Decisions pin and cost cap.** `init --decisions` pins the default Jev model unless a pin already exists, and records a billed cost cap from `--max-cost-usd`.
- **Association fallback consent.** `init --model-cli codex --lane association-fallback` records separate fallback consent and discloses what leaves the machine.

Proved by:

- [init_test.go](../../cmd/scout/init_test.go): `TestInitYesWritesConsentThatPolicyLoadRoundTrips`, `TestInitWithoutFlagsRecordsReportAndNoConsent`, `TestInitIsIdempotentAndNeverTouchesActivatedBlocking`, `TestInitRejectsInconsistentConsent`, `TestInitInteractiveRecordsAnswers`, `TestInitDecisionsPinsJev114ByDefaultAndKeepsAnExistingPin`, `TestInitRecordsABilledCostCap`.
- [association_terra_proof_test.go](../../cmd/scout/association_terra_proof_test.go): `TestInitAssociationTerraConsent`, `TestInitAssociationTerraDisclosure`.

## How to get to it (user POV)

1. Install Scout and activate the toolchain (see the [README](../../README.md)).
2. In the target repository, run `scout init` and answer the prompts, or `scout init --yes` for defaults.
3. Commit `.scout/policy.yaml`. Edit it later to activate checks; see [policy-file.md](../references/policy-file.md).

## Driving it with `go run ./cmd/scout`

```bash
go run ./cmd/scout init --root <fixture-repo> --yes --json
```

Outcome tests:

```bash
go test ./cmd/scout -run 'TestInit'
```

## Gotchas

- `init` never rewrites activated `blocking` entries; activating a check is a hand edit of `blocking` in the policy.
- `--decisions` without `--acknowledge-egress` exits `3` and writes no policy file.
- The association fallback always uses Codex, even when `model_cli.provider` is `claude`.
