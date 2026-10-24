package policy

import "fmt"

// DefaultAreasMatchJaccard is a configurable default, not a validated
// threshold: [ASSUMPTION; X2b `fallback`, X2's 0.30 is for another model] in
// design §4 A1. It applies at onboarding and manual agent maps.
const DefaultAreasMatchJaccard = 0.5

// Mapping is the effective `mapping:` policy block.
type Mapping struct {
	// AreasMatchJaccard is the file-set Jaccard at which an agent feature
	// matches an existing area (mapping.areas.match_jaccard).
	AreasMatchJaccard float64
}

// DefaultMapping returns the documented defaults.
func DefaultMapping() Mapping {
	return Mapping{AreasMatchJaccard: DefaultAreasMatchJaccard}
}

// Mapping returns the effective mapping policy.
func (p Policy) Mapping() Mapping { return p.mapping }

type rawMapping struct {
	Areas *struct {
		MatchJaccard *float64 `yaml:"match_jaccard"`
	} `yaml:"areas"`
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
	return m, nil
}
