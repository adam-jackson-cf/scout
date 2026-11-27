package policy

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultAreasMatchJaccard is a configurable default, not a validated
// threshold: [ASSUMPTION; X2b `fallback`, X2's 0.30 is for another model] in
// REQ-53. It applies at onboarding and manual agent maps.
const DefaultAreasMatchJaccard = 0.5

// Association configures the provisional, opt-in merge association lane.
type Association struct {
	Enabled          bool
	CoveredScore     float64
	TChoice          float64
	MaxPendingPerRun int
	Fallback         AssociationFallback
	Guard            AssociationGuard
}

// AssociationFallback configures the bounded Codex association decision pass.
type AssociationFallback struct {
	Model     string
	Effort    string
	Timeout   time.Duration
	MaxPerRun int
}

// AssociationGuard configures the advisory catalogue refresh guard.
type AssociationGuard struct {
	WindowMerges      int
	MinClusters       int
	FallbackRateMax   float64
	PendingNewRateMax float64
}

// Mapping is the effective `mapping:` policy block.
type Mapping struct {
	// AreasMatchJaccard is the file-set Jaccard at which an agent feature
	// matches an existing area (mapping.areas.match_jaccard).
	AreasMatchJaccard float64
	Association       Association
}

// DefaultMapping returns the documented defaults.
func DefaultMapping() Mapping {
	return Mapping{
		AreasMatchJaccard: DefaultAreasMatchJaccard,
		Association: Association{
			CoveredScore: .60, TChoice: .535, MaxPendingPerRun: 10,
			Fallback: AssociationFallback{Model: "gpt-6.1-terra", Effort: "high", Timeout: 900 * time.Second, MaxPerRun: 40},
			Guard:    AssociationGuard{WindowMerges: 20, MinClusters: 20, FallbackRateMax: .40, PendingNewRateMax: .20},
		},
	}
}

// Mapping returns the effective mapping policy.
func (p Policy) Mapping() Mapping { return p.mapping }

type rawMapping struct {
	Areas *struct {
		MatchJaccard *float64 `yaml:"match_jaccard"`
	} `yaml:"areas"`
	Association *rawAssociation `yaml:"association"`
}

type rawAssociation struct {
	Enabled          bool                    `yaml:"enabled"`
	CoveredScore     *float64                `yaml:"covered_score"`
	TChoice          *float64                `yaml:"t_choice"`
	MaxPendingPerRun *int                    `yaml:"max_pending_per_run"`
	Fallback         *rawAssociationFallback `yaml:"fallback"`
	Guard            *rawAssociationGuard    `yaml:"guard"`
}

type rawAssociationFallback struct {
	Model     *string `yaml:"model"`
	Effort    *string `yaml:"effort"`
	Timeout   *string `yaml:"timeout"`
	MaxPerRun *int    `yaml:"max_per_run"`
}

type rawAssociationGuard struct {
	WindowMerges      *int     `yaml:"window_merges"`
	MinClusters       *int     `yaml:"min_clusters"`
	FallbackRateMax   *float64 `yaml:"fallback_rate_max"`
	PendingNewRateMax *float64 `yaml:"pending_new_rate_max"`
}

func (r *rawAssociationGuard) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("mapping.association.guard must be a mapping")
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		key, value := n.Content[i].Value, n.Content[i+1]
		switch key {
		case "window_merges", "min_clusters":
			if value.Kind != yaml.ScalarNode || value.Tag != "!!int" {
				return fmt.Errorf("mapping.association.guard.%s must be an integer", key)
			}
		case "fallback_rate_max", "pending_new_rate_max":
			if value.Kind != yaml.ScalarNode || (value.Tag != "!!int" && value.Tag != "!!float") {
				return fmt.Errorf("mapping.association.guard.%s must be a number", key)
			}
		}
	}
	type plain rawAssociationGuard
	return decodeStrict(n, (*plain)(r))
}

func (r *rawAssociationFallback) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("mapping.association.fallback must be a mapping")
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		key, value := n.Content[i].Value, n.Content[i+1]
		switch key {
		case "model", "effort", "timeout":
			if value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
				return fmt.Errorf("mapping.association.fallback.%s must be a string", key)
			}
		case "max_per_run":
			if value.Kind != yaml.ScalarNode || value.Tag != "!!int" {
				return fmt.Errorf("mapping.association.fallback.max_per_run must be an integer")
			}
		}
	}
	type plain rawAssociationFallback
	return decodeStrict(n, (*plain)(r))
}

// UnmarshalYAML rejects scalar coercions while retaining strict unknown-key
// decoding. Explicit zero thresholds remain distinct from absent thresholds.
func (r *rawAssociation) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("mapping.association must be a mapping")
	}
	for i := 0; i+1 < len(n.Content); i += 2 {
		if err := checkAssociationValue(n.Content[i].Value, n.Content[i+1]); err != nil {
			return err
		}
	}
	type plain rawAssociation
	return decodeStrict(n, (*plain)(r))
}

func scalarTagged(n *yaml.Node, tags ...string) bool {
	return n.Kind == yaml.ScalarNode && slices.Contains(tags, n.Tag)
}

func checkAssociationValue(key string, value *yaml.Node) error {
	switch key {
	case "enabled":
		if !scalarTagged(value, "!!bool") {
			return fmt.Errorf("mapping.association.enabled must be a boolean")
		}
	case "covered_score", "t_choice":
		if !scalarTagged(value, "!!int", "!!float") {
			return fmt.Errorf("mapping.association.%s must be a number", key)
		}
	case "max_pending_per_run":
		if !scalarTagged(value, "!!int") {
			return fmt.Errorf("mapping.association.max_pending_per_run must be an integer")
		}
	case "fallback", "guard":
		if value.Kind != yaml.MappingNode {
			return fmt.Errorf("mapping.association.%s must be a mapping", key)
		}
	}
	return nil
}

func parseMapping(raw *rawMapping) (Mapping, error) {
	m := DefaultMapping()
	if raw == nil {
		return m, nil
	}
	if raw.Areas != nil && raw.Areas.MatchJaccard != nil {
		j := *raw.Areas.MatchJaccard
		if j <= 0 || j > 1 {
			return Mapping{}, fmt.Errorf("mapping.areas.match_jaccard %v must be in (0, 1]", j)
		}
		m.AreasMatchJaccard = j
	}
	if raw.Association != nil {
		if err := parseAssociation(raw.Association, &m.Association); err != nil {
			return Mapping{}, err
		}
	}
	return m, nil
}

func parseAssociation(raw *rawAssociation, a *Association) error {
	a.Enabled = raw.Enabled
	if raw.MaxPendingPerRun != nil {
		if *raw.MaxPendingPerRun < 0 {
			return fmt.Errorf("mapping.association.max_pending_per_run must not be negative")
		}
		a.MaxPendingPerRun = *raw.MaxPendingPerRun
	}
	if r := raw.Fallback; r != nil {
		if err := parseFallback(r, &a.Fallback); err != nil {
			return err
		}
	}
	if r := raw.Guard; r != nil {
		if err := parseGuard(r, &a.Guard); err != nil {
			return err
		}
	}
	return parseAssociationThresholds(raw, a)
}

func parseFallback(r *rawAssociationFallback, f *AssociationFallback) error {
	if err := parseFallbackChoices(r, f); err != nil {
		return err
	}
	return parseFallbackLimits(r, f)
}

func parseFallbackChoices(r *rawAssociationFallback, f *AssociationFallback) error {
	if r.Model != nil {
		model := strings.TrimSpace(*r.Model)
		if model == "" || IsModelAlias(model) {
			return fmt.Errorf("mapping.association.fallback.model must pin an exact model id")
		}
		f.Model = model
	}
	if r.Effort != nil {
		switch *r.Effort {
		case "minimal", "low", "medium", "high", "xhigh":
			f.Effort = *r.Effort
		default:
			return fmt.Errorf("mapping.association.fallback.effort %q is unsupported", *r.Effort)
		}
	}
	return nil
}

func parseFallbackLimits(r *rawAssociationFallback, f *AssociationFallback) error {
	if r.Timeout != nil {
		timeout, err := time.ParseDuration(*r.Timeout)
		if err != nil || timeout <= 0 {
			return fmt.Errorf("mapping.association.fallback.timeout must be a positive duration string")
		}
		f.Timeout = timeout
	}
	if r.MaxPerRun != nil {
		if *r.MaxPerRun < 0 {
			return fmt.Errorf("mapping.association.fallback.max_per_run must not be negative")
		}
		f.MaxPerRun = *r.MaxPerRun
	}
	return nil
}

func outsideUnitInterval(value float64) bool {
	return math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1
}

func parseGuard(r *rawAssociationGuard, g *AssociationGuard) error {
	for _, count := range []struct {
		name  string
		value *int
		dest  *int
	}{
		{"window_merges", r.WindowMerges, &g.WindowMerges},
		{"min_clusters", r.MinClusters, &g.MinClusters},
	} {
		if count.value == nil {
			continue
		}
		if *count.value <= 0 {
			return fmt.Errorf("mapping.association.guard.%s must be a positive integer", count.name)
		}
		*count.dest = *count.value
	}
	for _, threshold := range []struct {
		name  string
		value *float64
		dest  *float64
	}{
		{"fallback_rate_max", r.FallbackRateMax, &g.FallbackRateMax},
		{"pending_new_rate_max", r.PendingNewRateMax, &g.PendingNewRateMax},
	} {
		if threshold.value == nil {
			continue
		}
		value := *threshold.value
		if outsideUnitInterval(value) {
			return fmt.Errorf("mapping.association.guard.%s %v must be finite and in [0, 1]", threshold.name, value)
		}
		*threshold.dest = value
	}
	return nil
}

func parseAssociationThresholds(raw *rawAssociation, a *Association) error {
	for _, threshold := range []struct {
		name  string
		value *float64
		dest  *float64
	}{
		{"covered_score", raw.CoveredScore, &a.CoveredScore},
		{"t_choice", raw.TChoice, &a.TChoice},
	} {
		if threshold.value == nil {
			continue
		}
		value := *threshold.value
		if outsideUnitInterval(value) {
			return fmt.Errorf("mapping.association.%s %v must be finite and in [0, 1]", threshold.name, value)
		}
		*threshold.dest = value
	}
	return nil
}
