// Package policy loads and evaluates the repository policy file
// (.scout/policy.yaml). It is pure: no store access. The file is always read
// from the base ref so a pull request cannot change the rules that gate it.
//
// Schema:
//
//	profile: guardian            # report (default) | guardian
//	blocking:                    # either a list of activations ...
//	  - id: guardrails.SUPPRESS.added.unscoped
//	    version: "1"
//	    activated_at: 2026-01-01
//	    provisional: false
//	    evidence: {would_block: 3, warranted: 3, window_days: 14, pr_runs: 9, dossier_digest: sha256:...}
//	# ... or a mapping carrying the shadow window next to the activations:
//	blocking:
//	  shadow: {min_days: 14, min_pr_runs: 5}
//	  checks: [ ...same entries... ]
//	maintainers: [alice, "@bob"]
//	history: {bots: [dependabot[bot]]}
//	decisions:                   # consent for Jev (OpenRouter) ranking, design §6
//	  enabled: true
//	  provider: jev
//	  model: <pinned Jev model id; aliases such as jev-latest are rejected>
//	  max_per_run: 2000
//	  egress_acknowledged: true  # required when enabled
//	model_cli:                   # opt-in for the model-CLI lanes (§6, lane 9)
//	  provider: claude           # claude | codex
//	  lanes: [contextual, clawpatch-review]
//
// The architecture key is owned by a later phase and is accepted and ignored.
package policy

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Path is the policy file location inside the repository.
const Path = ".scout/policy.yaml"

// Profiles.
const (
	ProfileReport   = "report"
	ProfileGuardian = "guardian"
)

// Status is a check's blocking state for one repository.
type Status string

const (
	NotEligible       Status = "not_eligible"
	Shadow            Status = "shadow"
	Active            Status = "active"
	ActiveProvisional Status = "active_provisional"
)

// Default shadow window (design §7.3).
const (
	DefaultMinDays   = 14
	DefaultMinPRRuns = 5
)

// Evidence records why a check was activated.
type Evidence struct {
	WouldBlock    int    `yaml:"would_block"`
	Warranted     int    `yaml:"warranted"`
	WindowDays    int    `yaml:"window_days"`
	PRRuns        int    `yaml:"pr_runs"`
	DossierDigest string `yaml:"dossier_digest"`
}

// Activation is one `blocking` entry.
type Activation struct {
	ID          string    `yaml:"id"`
	Version     string    `yaml:"version"`
	ActivatedAt string    `yaml:"activated_at"`
	Provisional bool      `yaml:"provisional"`
	Evidence    *Evidence `yaml:"evidence"`
}

// ShadowConfig is the configurable shadow window.
type ShadowConfig struct {
	MinDays   int `yaml:"min_days"`
	MinPRRuns int `yaml:"min_pr_runs"`
}

// Policy is the validated policy. The zero value is not usable; use Default,
// Parse or Load.
type Policy struct {
	profile     string
	shadow      ShadowConfig
	activations map[string]Activation
	order       []string
	maintainers map[string]bool
	decisions   Decisions
	modelCLI    ModelCLI
	// HistoryBots is the history.bots passthrough.
	HistoryBots []string
}

// Default returns the policy used when no file exists at the base ref.
func Default() Policy {
	return Policy{
		profile:     ProfileReport,
		shadow:      ShadowConfig{MinDays: DefaultMinDays, MinPRRuns: DefaultMinPRRuns},
		activations: map[string]Activation{},
		maintainers: map[string]bool{},
		decisions:   Decisions{Provider: ProviderJev, MaxPerRun: DefaultMaxPerRun},
	}
}

// Profile returns "report" or "guardian".
func (p Policy) Profile() string { return p.profile }

// Shadow returns the effective shadow window.
func (p Policy) Shadow() ShadowConfig { return p.shadow }

// Activations returns the activation entries in file order.
func (p Policy) Activations() []Activation {
	out := make([]Activation, 0, len(p.order))
	for _, id := range p.order {
		out = append(out, p.activations[id])
	}
	return out
}

// Load reads .scout/policy.yaml from baseRef (never the working tree). An
// empty baseRef or a file absent at baseRef yields Default. A ref that does
// not resolve is an error.
func Load(repoRoot, baseRef string) (Policy, error) {
	if baseRef == "" {
		return Default(), nil
	}
	ls, err := git(repoRoot, "ls-tree", "--name-only", baseRef, "--", Path)
	if err != nil {
		return Policy{}, fmt.Errorf("policy: resolve %s at %s: %w", Path, baseRef, err)
	}
	if strings.TrimSpace(string(ls)) == "" {
		return Default(), nil
	}
	data, err := git(repoRoot, "show", baseRef+":"+Path)
	if err != nil {
		return Policy{}, fmt.Errorf("policy: read %s at %s: %w", Path, baseRef, err)
	}
	p, err := Parse(data)
	if err != nil {
		return Policy{}, fmt.Errorf("policy: %s at %s: %w", Path, baseRef, err)
	}
	return p, nil
}

func git(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

type blockingSection struct {
	Shadow *ShadowConfig
	Checks []Activation
}

// UnmarshalYAML accepts the list form or the {shadow, checks} mapping form.
func (b *blockingSection) UnmarshalYAML(n *yaml.Node) error {
	switch n.Kind {
	case yaml.SequenceNode:
		return decodeStrict(n, &b.Checks)
	case yaml.MappingNode:
		var m struct {
			Shadow *ShadowConfig `yaml:"shadow"`
			Checks []Activation  `yaml:"checks"`
		}
		if err := decodeStrict(n, &m); err != nil {
			return err
		}
		b.Shadow, b.Checks = m.Shadow, m.Checks
		return nil
	}
	return fmt.Errorf("line %d: blocking must be a list or a mapping", n.Line)
}

func decodeStrict(n *yaml.Node, v any) error {
	// yaml.Node.Decode has no KnownFields switch; re-encode through a decoder.
	raw, err := yaml.Marshal(n)
	if err != nil {
		return err
	}
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("line %d: %w", n.Line, err)
	}
	return nil
}

type rawPolicy struct {
	Profile     string          `yaml:"profile"`
	Blocking    blockingSection `yaml:"blocking"`
	Maintainers []string        `yaml:"maintainers"`
	History     *struct {
		Bots []string `yaml:"bots"`
	} `yaml:"history"`
	Decisions    *rawDecisions `yaml:"decisions"`
	ModelCLI     *rawModelCLI  `yaml:"model_cli"`
	Architecture yaml.Node     `yaml:"architecture"`
}

// Parse validates policy bytes strictly.
func Parse(data []byte) (Policy, error) {
	p := Default()
	if len(bytes.TrimSpace(data)) == 0 {
		return p, nil
	}
	var raw rawPolicy
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&raw); err != nil {
		return Policy{}, fmt.Errorf("invalid policy: %w", err)
	}
	switch raw.Profile {
	case "":
	case ProfileReport, ProfileGuardian:
		p.profile = raw.Profile
	default:
		return Policy{}, fmt.Errorf("invalid policy: profile %q must be %q or %q", raw.Profile, ProfileReport, ProfileGuardian)
	}
	if s := raw.Blocking.Shadow; s != nil {
		if s.MinDays < 0 || s.MinPRRuns < 0 {
			return Policy{}, errors.New("invalid policy: blocking.shadow values must not be negative")
		}
		if s.MinDays > 0 {
			p.shadow.MinDays = s.MinDays
		}
		if s.MinPRRuns > 0 {
			p.shadow.MinPRRuns = s.MinPRRuns
		}
	}
	for i, a := range raw.Blocking.Checks {
		if err := validateActivation(a); err != nil {
			return Policy{}, fmt.Errorf("invalid policy: blocking[%d] (%q): %w", i, a.ID, err)
		}
		if _, dup := p.activations[a.ID]; dup {
			return Policy{}, fmt.Errorf("invalid policy: blocking[%d]: duplicate id %q", i, a.ID)
		}
		p.activations[a.ID] = a
		p.order = append(p.order, a.ID)
	}
	for i, m := range raw.Maintainers {
		k := normalizeActor(m)
		if k == "" {
			return Policy{}, fmt.Errorf("invalid policy: maintainers[%d] is empty", i)
		}
		p.maintainers[k] = true
	}
	if raw.History != nil {
		p.HistoryBots = append([]string(nil), raw.History.Bots...)
	}
	var err error
	if p.decisions, err = parseDecisions(raw.Decisions); err != nil {
		return Policy{}, fmt.Errorf("invalid policy: %w", err)
	}
	if p.modelCLI, err = parseModelCLI(raw.ModelCLI); err != nil {
		return Policy{}, fmt.Errorf("invalid policy: %w", err)
	}
	return p, nil
}

func validateActivation(a Activation) error {
	if err := ValidateEligibleID(a.ID); err != nil {
		return err
	}
	if strings.TrimSpace(a.Version) == "" {
		return errors.New("version is required")
	}
	if a.ActivatedAt == "" {
		return errors.New("activated_at is required")
	}
	if _, err := time.Parse(time.RFC3339, a.ActivatedAt); err != nil {
		if _, err2 := time.Parse("2006-01-02", a.ActivatedAt); err2 != nil {
			return fmt.Errorf("activated_at %q is not RFC 3339 or YYYY-MM-DD", a.ActivatedAt)
		}
	}
	if e := a.Evidence; e != nil {
		if e.WouldBlock < 0 || e.Warranted < 0 || e.WindowDays < 0 || e.PRRuns < 0 {
			return errors.New("evidence counts must not be negative")
		}
	} else if !a.Provisional {
		return errors.New("evidence is required unless provisional: true")
	}
	return nil
}

var (
	slugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)

	fixedEligible = map[string]bool{
		"security.ci-mutable-action-ref":      true,
		"guardrails.SUPPRESS.added.unscoped":  true,
		"guardrails.TESTQ.focused-test-added": true,
	}
)

const (
	secretsPrefix    = "security.secrets."
	ignoresPrefix    = "guardrails.CONFIG.ignores-added."
	strictnessPrefix = "guardrails.CONFIG.strictness-lowered."
)

// ValidateEligibleID reports why id is not on the AL table-1 list; nil means
// eligible. There are no wildcards: parameterised families take one exact
// segment, and generic-api-key is excluded.
func ValidateEligibleID(id string) error {
	if id == "" {
		return errors.New("empty check id")
	}
	if strings.ContainsAny(id, "*?[]") {
		return fmt.Errorf("wildcard check id %q is not allowed", id)
	}
	if strings.HasPrefix(id, "delta.") {
		return fmt.Errorf("check id %q is a descriptive delta rule and can never block", id)
	}
	if fixedEligible[id] {
		return nil
	}
	for _, f := range []struct{ prefix, what string }{
		{secretsPrefix, "gitleaks rule id"},
		{ignoresPrefix, "tool"},
		{strictnessPrefix, "tool"},
	} {
		if seg, ok := strings.CutPrefix(id, f.prefix); ok {
			if !slugRe.MatchString(seg) {
				return fmt.Errorf("check id %q: invalid %s %q", id, f.what, seg)
			}
			if f.prefix == secretsPrefix && seg == "generic-api-key" {
				return fmt.Errorf("check id %q is excluded from blocking (report-only)", id)
			}
			return nil
		}
	}
	return fmt.Errorf("check id %q is not on the blocking allow-list", id)
}

// CheckStatus returns the blocking state of checkID at ruleVersion. Only the
// guardian profile can activate; an activation whose version differs from the
// current rule version returns to shadow.
func (p Policy) CheckStatus(checkID, ruleVersion string) Status {
	if ValidateEligibleID(checkID) != nil {
		return NotEligible
	}
	if p.profile != ProfileGuardian {
		return Shadow
	}
	a, ok := p.activations[checkID]
	if !ok || a.Version != ruleVersion {
		return Shadow
	}
	if a.Provisional {
		return ActiveProvisional
	}
	return Active
}

// ShadowStats is the window evidence for one check in one repository.
type ShadowStats struct {
	CheckID string
	// Days since the first shadow run with a comparable base.
	Days int
	// PRRuns with a comparable base.
	PRRuns int
	// WouldBlock events recorded in the window.
	WouldBlock int
	// Warranted, NotWarranted, Undecided are maintainer triage counts.
	Warranted, NotWarranted, Undecided int
	// UnknownAttribution runs out of TotalRuns.
	UnknownAttribution, TotalRuns int
}

// ShadowReady implements design §7.3. A reason is always returned. A ready
// result for a rare security check with zero events starts with "provisional:";
// the activation must then carry provisional: true.
func (p Policy) ShadowReady(s ShadowStats) (bool, string) {
	if err := ValidateEligibleID(s.CheckID); err != nil {
		return false, err.Error()
	}
	if s.NotWarranted > 0 {
		return false, fmt.Sprintf("%d not_warranted would_block event(s)", s.NotWarranted)
	}
	if s.TotalRuns > 0 && s.UnknownAttribution*100 > 10*s.TotalRuns {
		return false, fmt.Sprintf("unknown attribution %d/%d runs exceeds 10%%", s.UnknownAttribution, s.TotalRuns)
	}
	if s.Days < p.shadow.MinDays {
		return false, fmt.Sprintf("shadow window %d day(s) < min_days %d", s.Days, p.shadow.MinDays)
	}
	if isRareSecurity(s.CheckID) && s.WouldBlock == 0 {
		return true, "provisional: rare security check with zero would_block events at window end; activate with provisional: true"
	}
	if s.PRRuns < p.shadow.MinPRRuns {
		return false, fmt.Sprintf("%d PR run(s) < min_pr_runs %d", s.PRRuns, p.shadow.MinPRRuns)
	}
	if s.Undecided > 0 {
		return false, fmt.Sprintf("%d undecided would_block event(s) await maintainer triage", s.Undecided)
	}
	return true, "shadow window satisfied"
}

func isRareSecurity(id string) bool {
	return id == "security.ci-mutable-action-ref" || strings.HasPrefix(id, secretsPrefix)
}

func normalizeActor(s string) string {
	return strings.ToLower(strings.TrimPrefix(strings.TrimSpace(s), "@"))
}

// AcceptOverride reports whether actor may record a scout-override: a
// policy-declared maintainer or an owner named in CODEOWNERS. Team owners
// (@org/team) match only by exact name; membership is not resolved here.
func (p Policy) AcceptOverride(actor string, codeowners []byte) (bool, string) {
	a := normalizeActor(actor)
	if a == "" {
		return false, "empty actor"
	}
	if p.maintainers[a] {
		return true, "declared in policy maintainers"
	}
	for _, line := range strings.Split(string(codeowners), "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		f := strings.Fields(line)
		for _, owner := range f[min(1, len(f)):] {
			if normalizeActor(owner) == a {
				return true, "listed in CODEOWNERS"
			}
		}
	}
	return false, fmt.Sprintf("%s is neither a policy maintainer nor a CODEOWNERS entry", actor)
}
