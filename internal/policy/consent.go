package policy

import (
	"bytes"
	"errors"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Decisions provider and model-CLI constants.
const (
	ProviderJev             = "jev"
	CLIClaude               = "claude"
	CLICodex                = "codex"
	LaneContextual          = "contextual"
	LaneClawpatchRev        = "clawpatch-review"
	LaneAssociationFallback = "association-fallback"
	// DefaultMaxPerRun is the Jev call cap (REQ-29, [ASSUMPTION]).
	DefaultMaxPerRun = 2000
	// DefaultMaxCostUSD is the cap on billed Jev cost per run, in US dollars
	// [ASSUMPTION]: five times the ≈ $0.10 the default call cap costs at E2's
	// $0.05 per 1k calls, so it binds only on a cost surprise, not on the
	// call cap.
	DefaultMaxCostUSD = 0.50
	// DefaultModel is what `scout init --decisions` pins when no model is
	// given: Jev 1.14 through OpenRouter. The slug is the dated endpoint name
	// OpenRouter lists for `typesafe/jev-1.14` (the undated id and the
	// `jev-latest`/`jev-preview` aliases move and are rejected).
	DefaultModel = "typesafe/jev-1.14-20261014"
)

// Decisions is the consent and configuration for Jev (typed-decision) ranking.
// Enabled without Model and EgressAcknowledged is rejected at load.
type Decisions struct {
	Enabled            bool
	Provider           string
	Model              string
	MaxPerRun          int
	MaxCostUSD         float64
	EgressAcknowledged bool
}

// ModelCLI is the opt-in for the model-CLI lanes: which CLI and which lanes.
type ModelCLI struct {
	Provider string
	Lanes    []string
}

// Decisions returns the decisions consent; the zero consent is disabled.
func (p Policy) Decisions() Decisions { return p.decisions }

// ModelCLI returns the model-CLI lane opt-in; no lanes means opted out.
func (p Policy) ModelCLI() ModelCLI {
	return ModelCLI{Provider: p.modelCLI.Provider, Lanes: slices.Clone(p.modelCLI.Lanes)}
}

// LaneOptedIn reports whether the named model-CLI lane was opted in.
func (p Policy) LaneOptedIn(lane string) bool { return slices.Contains(p.modelCLI.Lanes, lane) }

type rawDecisions struct {
	Enabled            bool    `yaml:"enabled"`
	Provider           string  `yaml:"provider"`
	Model              string  `yaml:"model"`
	MaxPerRun          int     `yaml:"max_per_run"`
	MaxCostUSD         float64 `yaml:"max_cost_usd"`
	EgressAcknowledged bool    `yaml:"egress_acknowledged"`
}

type rawModelCLI struct {
	Provider string   `yaml:"provider"`
	Lanes    []string `yaml:"lanes"`
}

// undatedJev matches an OpenRouter Jev slug that names a release line, not a
// build (`typesafe/jev-1.14`): OpenRouter re-points it at new dated builds and
// reports the dated slug as the resolved model, so it can never be a pin.
var undatedJev = regexp.MustCompile(`^(?:[^/]+/)?jev-\d+\.\d+$`)

// IsModelAlias reports whether model is a floating name that would leave the
// pinned model unpinned: `jev-latest`, `jev-preview`, `latest`, `preview`,
// `auto`, any `-latest`/`-preview`/`:latest` suffix, or an undated release
// line such as `typesafe/jev-1.14`.
func IsModelAlias(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	switch m {
	case "jev-latest", "jev-preview", "latest", "preview", "auto":
		return true
	}
	return strings.HasSuffix(m, "-latest") || strings.HasSuffix(m, "-preview") || strings.HasSuffix(m, ":latest") || undatedJev.MatchString(m)
}

func parseDecisions(r *rawDecisions) (Decisions, error) {
	d := Decisions{Provider: ProviderJev, MaxPerRun: DefaultMaxPerRun, MaxCostUSD: DefaultMaxCostUSD}
	if r == nil {
		return d, nil
	}
	if r.Provider != "" && r.Provider != ProviderJev {
		return Decisions{}, fmt.Errorf("decisions.provider %q must be %q", r.Provider, ProviderJev)
	}
	if err := applyDecisionCaps(&d, r); err != nil {
		return Decisions{}, err
	}
	d.Enabled, d.Model, d.EgressAcknowledged = r.Enabled, strings.TrimSpace(r.Model), r.EgressAcknowledged
	if err := validateDecisionModel(d); err != nil {
		return Decisions{}, err
	}
	return d, nil
}

func applyDecisionCaps(d *Decisions, r *rawDecisions) error {
	if r.MaxCostUSD < 0 || math.IsNaN(r.MaxCostUSD) || math.IsInf(r.MaxCostUSD, 0) {
		return errors.New("decisions.max_cost_usd must be a non-negative number")
	}
	if r.MaxCostUSD > 0 {
		d.MaxCostUSD = r.MaxCostUSD
	}
	if r.MaxPerRun < 0 {
		return errors.New("decisions.max_per_run must not be negative")
	}
	if r.MaxPerRun > 0 {
		d.MaxPerRun = r.MaxPerRun
	}
	return nil
}

func validateDecisionModel(d Decisions) error {
	if d.Model != "" && IsModelAlias(d.Model) {
		return fmt.Errorf("decisions.model %q is an alias; pin an exact model id", d.Model)
	}
	if d.Enabled {
		if d.Model == "" {
			return errors.New("decisions.model is required when decisions.enabled is true")
		}
		if !d.EgressAcknowledged {
			return errors.New("decisions.egress_acknowledged must be true when decisions.enabled is true")
		}
	}
	return nil
}

func parseModelCLI(r *rawModelCLI) (ModelCLI, error) {
	if r == nil {
		return ModelCLI{}, nil
	}
	switch r.Provider {
	case CLIClaude, CLICodex:
	case "":
		if len(r.Lanes) > 0 {
			return ModelCLI{}, errors.New("model_cli.provider is required when model_cli.lanes is set")
		}
	default:
		return ModelCLI{}, fmt.Errorf("model_cli.provider %q must be %q or %q", r.Provider, CLIClaude, CLICodex)
	}
	seen := map[string]bool{}
	for _, l := range r.Lanes {
		if l != LaneContextual && l != LaneClawpatchRev && l != LaneAssociationFallback {
			return ModelCLI{}, fmt.Errorf("model_cli.lanes: unknown lane %q", l)
		}
		if seen[l] {
			return ModelCLI{}, fmt.Errorf("model_cli.lanes: duplicate lane %q", l)
		}
		seen[l] = true
	}
	return ModelCLI{Provider: r.Provider, Lanes: slices.Clone(r.Lanes)}, nil
}

// ConsentUpdate carries the values `scout init` sets. A nil field is left as
// it is in an existing file; for a new file only Profile defaults (to report).
type ConsentUpdate struct {
	Profile            *string
	DecisionsEnabled   *bool
	Model              *string
	MaxPerRun          *int
	MaxCostUSD         *float64
	EgressAcknowledged *bool
	ModelCLIProvider   *string
	ModelCLILanes      *[]string
}

// ApplyConsent merges u into the policy bytes existing (nil or empty for a
// new file) and returns the new bytes and whether anything changed. Only the
// profile, decisions and model_cli keys are touched; blocking activations,
// maintainers, history and comments are preserved. The result is validated
// with Parse; on error nothing is returned. When nothing changes the input is
// returned unchanged, so re-running init is idempotent.
func ApplyConsent(existing []byte, u ConsentUpdate) ([]byte, bool, error) {
	doc, err := consentDocument(existing)
	if err != nil {
		return nil, false, err
	}
	root := doc.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, false, errors.New("invalid policy: top level must be a mapping")
	}
	e := &consentEditor{root: root}
	e.applyProfile(u.Profile)
	e.applyDecisions(u)
	e.applyModelCLI(u)
	if !e.changed {
		return existing, false, nil
	}
	out, err := encodePolicy(&doc)
	if err != nil {
		return nil, false, err
	}
	return out, true, nil
}

func consentDocument(existing []byte) (yaml.Node, error) {
	var doc yaml.Node
	if len(bytes.TrimSpace(existing)) != 0 {
		if _, err := Parse(existing); err != nil {
			return yaml.Node{}, err
		}
		if err := yaml.Unmarshal(existing, &doc); err != nil {
			return yaml.Node{}, fmt.Errorf("invalid policy: %w", err)
		}
	}
	if doc.Kind == 0 {
		doc = yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}}
	}
	return doc, nil
}

func encodePolicy(doc *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	if _, err := Parse(buf.Bytes()); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type consentEditor struct {
	root    *yaml.Node
	changed bool
}

func (e *consentEditor) set(m *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			if !sameNode(m.Content[i+1], value) {
				m.Content[i+1] = value
				e.changed = true
			}
			return
		}
	}
	m.Content = append(m.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, value)
	e.changed = true
}

func (e *consentEditor) section(key string) *yaml.Node {
	root := e.root
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == key {
			if root.Content[i+1].Kind != yaml.MappingNode {
				root.Content[i+1] = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				e.changed = true
			}
			return root.Content[i+1]
		}
	}
	m := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	root.Content = append(root.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, m)
	e.changed = true
	return m
}

func (e *consentEditor) hasKey(key string) bool {
	for i := 0; i+1 < len(e.root.Content); i += 2 {
		if e.root.Content[i].Value == key {
			return true
		}
	}
	return false
}

func strNode(v string) *yaml.Node { return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v} }

func boolNode(v bool) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: fmt.Sprint(v)}
}

func (e *consentEditor) applyProfile(profile *string) {
	switch {
	case profile != nil:
		e.set(e.root, "profile", strNode(*profile))
	case !e.hasKey("profile"):
		e.set(e.root, "profile", strNode(ProfileReport))
	}
}

func decisionsRequested(u ConsentUpdate) bool {
	return u.DecisionsEnabled != nil || u.Model != nil || u.MaxPerRun != nil || u.MaxCostUSD != nil || u.EgressAcknowledged != nil
}

func (e *consentEditor) applyDecisions(u ConsentUpdate) {
	if !decisionsRequested(u) {
		return
	}
	d := e.section("decisions")
	if u.DecisionsEnabled != nil {
		e.set(d, "enabled", boolNode(*u.DecisionsEnabled))
		e.set(d, "provider", strNode(ProviderJev))
	}
	if u.Model != nil {
		e.set(d, "model", strNode(*u.Model))
	}
	if u.MaxPerRun != nil {
		e.set(d, "max_per_run", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprint(*u.MaxPerRun)})
	}
	if u.MaxCostUSD != nil {
		e.set(d, "max_cost_usd", &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!float", Value: strconv.FormatFloat(*u.MaxCostUSD, 'f', -1, 64)})
	}
	if u.EgressAcknowledged != nil {
		e.set(d, "egress_acknowledged", boolNode(*u.EgressAcknowledged))
	}
}

func (e *consentEditor) applyModelCLI(u ConsentUpdate) {
	if u.ModelCLIProvider == nil && u.ModelCLILanes == nil {
		return
	}
	m := e.section("model_cli")
	if u.ModelCLIProvider != nil {
		e.set(m, "provider", strNode(*u.ModelCLIProvider))
	}
	if u.ModelCLILanes != nil {
		seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Style: yaml.FlowStyle}
		for _, l := range *u.ModelCLILanes {
			seq.Content = append(seq.Content, strNode(l))
		}
		e.set(m, "lanes", seq)
	}
}

// sameNode reports whether two scalar or flow-sequence nodes carry the same value.
func sameNode(a, b *yaml.Node) bool {
	if a.Kind != b.Kind || len(a.Content) != len(b.Content) {
		return false
	}
	if a.Kind == yaml.ScalarNode {
		return a.Value == b.Value
	}
	for i := range a.Content {
		if !sameNode(a.Content[i], b.Content[i]) {
			return false
		}
	}
	return true
}
