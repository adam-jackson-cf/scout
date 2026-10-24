# `.scout/policy.yaml`

Use this catalog when you need to change what Scout gates, who may override it, or which optional lanes are consented. Source: `internal/policy/policy.go`, `internal/policy/consent.go`, `internal/policy/mapping.go`.

The file is parsed strictly, so an unknown key is an error. Scout reads it from the **base ref** of a run, never from the working tree, so a pull request cannot loosen its own gate. Without the file, Scout uses the defaults: profile `report`, no activations, decisions and model-CLI lanes off (`policy.go` `Default`, `Load`).

`scout init --yes` writes the minimal file:

```yaml
profile: report
```

## Keys

| Key | What it means | Default / rules | Evidence source |
| --- | --- | --- | --- |
| `profile` | `report`: findings never fail CI. `guardian`: activated checks can gate. | `report` | `policy.go` `Parse` |
| `blocking` | The activated checks. Either a list of activations, or a mapping with `shadow` and `checks`. | empty | `policy.go` `blockingSection` |
| `blocking.shadow.min_days`, `blocking.shadow.min_pr_runs` | How long a check must collect `would_block` evidence in shadow before activation. | 14 days, 5 PR runs | `policy.go:66-69` |
| `blocking.checks[].id` | An eligible check ID, for example `guardrails.SUPPRESS.added.unscoped` or `security.secrets.<rule>`. `delta.*` is rejected. | required | `policy.go` `validateActivation` |
| `blocking.checks[].version` | The rule or rule-pack version that was activated. If the running version differs, the check returns to shadow. | required | same; `docs/STATUS.md` §1 |
| `blocking.checks[].activated_at` | RFC 3339 timestamp or `YYYY-MM-DD`. | required | same |
| `blocking.checks[].evidence` | `would_block`, `warranted`, `window_days`, `pr_runs`, `dossier_digest`: the shadow evidence behind the activation. | required unless `provisional: true` | same |
| `blocking.checks[].provisional` | Activation without full evidence, for example a rare security check with zero shadow events. | `false` | `policy.go:413-414` |
| `maintainers` | Actors whose `scout-override` is honoured. Overrides from anyone else are rejected and logged. | empty; CODEOWNERS is also read | `policy.go` `Parse`, `:444` |
| `history.bots` | Exact author names or e-mails to exclude from history measures, in addition to the `[bot]` suffix. | empty | design §8.3 |
| `decisions.enabled`, `.model`, `.max_per_run`, `.max_cost_usd`, `.egress_acknowledged` | Consent for Jev ranking through OpenRouter. `scout init --decisions --decisions-model <id> --acknowledge-egress` sets these. Undated model aliases are rejected. | off; 2000 calls; $0.50 per run | `consent.go` |
| `model_cli.provider` | `claude` or `codex` for the opted-in model-CLI lanes. Takes precedence over `SCOUT_MODEL_CLI`. | unset | `consent.go`; `cmd/scout/main.go` `providerConfig` |
| `model_cli.lanes` | Opted-in lanes: `contextual`, `clawpatch-review`. | none | `consent.go` |
| `mapping.areas.match_jaccard` | File-set Jaccard threshold at which a new agent feature matches an existing area, in (0, 1]. Applies at onboarding and manual agent maps. | `0.5` | `mapping.go` |
| `architecture` | Reserved for declared boundaries (design §5.1 row 7). | — | `policy.go` `rawPolicy` |

## Example

```yaml
profile: guardian
maintainers: [alice]
blocking:
  shadow: {min_days: 14, min_pr_runs: 5}
  checks:
    - id: guardrails.SUPPRESS.added.unscoped
      version: guardrails.SUPPRESS.added.unscoped@1
      activated_at: 2026-10-23
      evidence: {would_block: 6, warranted: 6, window_days: 14, pr_runs: 9}
```

## Notes

- `scout init` is idempotent: it keeps existing activations and comments, and keeps an existing `profile` unless you pass `--profile`.
- Model-CLI lanes and Jev stay off for a repository without a committed policy, even if `SCOUT_MODEL_CLI` is set.
- The `mapping` keys are part of Phase 4 work that is in the working tree but not yet committed (`internal/policy/mapping.go`).
- `mapping.refresh` was removed in revision 2.6; a policy that still sets it (for example `mapping.refresh.debounce`) fails to parse because the decoder is strict.
