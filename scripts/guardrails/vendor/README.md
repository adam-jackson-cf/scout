# Upstream guardrail coverage snapshot

These files are vendored byte-for-byte, not repository-maintained Python or
formatting inputs. Update them only by copying a reviewed upstream skill snapshot
and updating these hashes. The check-only CI gate uses this local validator and
catalog; it does not depend on a contributor’s home directory.

| File | Source path | SHA-256 |
| --- | --- | --- |
| `validate_rule_coverage.py` | `/Users/adamjackson/.omp/agent/skills/setup-repository-guardrails/scripts/validate_rule_coverage.py` | `ddce1136f9c0b00076939aced1bb83ced0a021124eecb5024efeece9703e33a0` |
| `guardrail-rules.json` | `/Users/adamjackson/.omp/agent/skills/setup-repository-guardrails/assets/guardrail-rules.json` | `60639e533a6d3ed0f59bd45a2979f4026bcdf288ffcecadb0823fd8a9ec220fb` |

`guardrail-rules.json` differs from its upstream snapshot in one line: the
`GATE.local-ci-parity` contract names the host-neutral `ci` field
(`"ci_contract_field": "ci"`) instead of the retired host enum field. Re-vendor the
upstream file once its catalog carries the same change.
