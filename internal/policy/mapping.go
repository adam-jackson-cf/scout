package policy

import (
	"fmt"
	"math"

	"gopkg.in/yaml.v3"
)

// DefaultAreasMatchJaccard is a configurable default, not a validated
// threshold: [ASSUMPTION; X2b `fallback`, X2's 0.30 is for another model] in
// REQ-53. It applies at onboarding and manual agent maps.
const DefaultAreasMatchJaccard = 0.5

// Association configures the provisional, opt-in merge association lane.
type Association struct {
	Enabled      bool
	CoveredScore float64
	TChoice      float64
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
		Association:       Association{CoveredScore: .60, TChoice: .535},
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
	Enabled      bool     `yaml:"enabled"`
	CoveredScore *float64 `yaml:"covered_score"`
	TChoice      *float64 `yaml:"t_choice"`
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
