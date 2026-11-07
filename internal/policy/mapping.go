package policy

import (
	"fmt"
	"math"
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
}

// AssociationFallback configures the bounded Codex association decision pass.
type AssociationFallback struct {
	Model     string
	Effort    string
	Timeout   time.Duration
	MaxPerRun int
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
}

type rawAssociationFallback struct {
	Model     *string `yaml:"model"`
	Effort    *string `yaml:"effort"`
	Timeout   *string `yaml:"timeout"`
	MaxPerRun *int    `yaml:"max_per_run"`
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
		key, value := n.Content[i].Value, n.Content[i+1]
		switch key {
		case "enabled":
			if value.Kind != yaml.ScalarNode || value.Tag != "!!bool" {
				return fmt.Errorf("mapping.association.enabled must be a boolean")
			}
		case "covered_score", "t_choice":
			if value.Kind != yaml.ScalarNode || (value.Tag != "!!int" && value.Tag != "!!float") {
				return fmt.Errorf("mapping.association.%s must be a number", key)
			}
		case "max_pending_per_run":
			if value.Kind != yaml.ScalarNode || value.Tag != "!!int" {
				return fmt.Errorf("mapping.association.max_pending_per_run must be an integer")
			}
		case "fallback":
			if value.Kind != yaml.MappingNode {
				return fmt.Errorf("mapping.association.fallback must be a mapping")
			}
		}
	}
	type plain rawAssociation
	return decodeStrict(n, (*plain)(r))
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
		m.Association.Enabled = raw.Association.Enabled
		if raw.Association.MaxPendingPerRun != nil {
			if *raw.Association.MaxPendingPerRun < 0 {
				return Mapping{}, fmt.Errorf("mapping.association.max_pending_per_run must not be negative")
			}
			m.Association.MaxPendingPerRun = *raw.Association.MaxPendingPerRun
		}
		if r := raw.Association.Fallback; r != nil {
			f := &m.Association.Fallback
			if r.Model != nil {
				model := strings.TrimSpace(*r.Model)
				if model == "" || IsModelAlias(model) {
					return Mapping{}, fmt.Errorf("mapping.association.fallback.model must pin an exact model id")
				}
				f.Model = model
			}
			if r.Effort != nil {
				switch *r.Effort {
				case "minimal", "low", "medium", "high", "xhigh":
					f.Effort = *r.Effort
				default:
					return Mapping{}, fmt.Errorf("mapping.association.fallback.effort %q is unsupported", *r.Effort)
				}
			}
			if r.Timeout != nil {
				timeout, err := time.ParseDuration(*r.Timeout)
				if err != nil || timeout <= 0 {
					return Mapping{}, fmt.Errorf("mapping.association.fallback.timeout must be a positive duration string")
				}
				f.Timeout = timeout
			}
			if r.MaxPerRun != nil {
				if *r.MaxPerRun < 0 {
					return Mapping{}, fmt.Errorf("mapping.association.fallback.max_per_run must not be negative")
				}
				f.MaxPerRun = *r.MaxPerRun
			}
		}
		for _, threshold := range []struct {
			name  string
			value *float64
			dest  *float64
		}{
			{"covered_score", raw.Association.CoveredScore, &m.Association.CoveredScore},
			{"t_choice", raw.Association.TChoice, &m.Association.TChoice},
		} {
			if threshold.value == nil {
				continue
			}
			value := *threshold.value
			if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 1 {
				return Mapping{}, fmt.Errorf("mapping.association.%s %v must be finite and in [0, 1]", threshold.name, value)
			}
			*threshold.dest = value
		}
	}
	return m, nil
}
