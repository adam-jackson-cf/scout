package policy

import (
	"fmt"
	"time"
)

// Mapping defaults. Both values are [ASSUMPTION, X2] in design §4: they are
// configurable defaults, not validated thresholds.
const (
	DefaultRefreshDebounce   = time.Hour
	DefaultAreasMatchJaccard = 0.5
)

// Mapping is the effective `mapping:` policy block.
type Mapping struct {
	// RefreshDebounce is the quiet time since the last triggering merge before
	// one agent refresh runs (mapping.refresh.debounce).
	RefreshDebounce time.Duration
	// AreasMatchJaccard is the file-set Jaccard at which an agent feature
	// matches an existing area (mapping.areas.match_jaccard).
	AreasMatchJaccard float64
}

// DefaultMapping returns the documented defaults.
func DefaultMapping() Mapping {
	return Mapping{RefreshDebounce: DefaultRefreshDebounce, AreasMatchJaccard: DefaultAreasMatchJaccard}
}

// Mapping returns the effective mapping policy.
func (p Policy) Mapping() Mapping {
	if p.mapping.RefreshDebounce == 0 && p.mapping.AreasMatchJaccard == 0 {
		return DefaultMapping()
	}
	return p.mapping
}

type rawMapping struct {
	Refresh *struct {
		Debounce string `yaml:"debounce"`
	} `yaml:"refresh"`
	Areas *struct {
		MatchJaccard *float64 `yaml:"match_jaccard"`
	} `yaml:"areas"`
}

func parseMapping(raw *rawMapping) (Mapping, error) {
	m := DefaultMapping()
	if raw == nil {
		return m, nil
	}
	if raw.Refresh != nil && raw.Refresh.Debounce != "" {
		d, err := time.ParseDuration(raw.Refresh.Debounce)
		if err != nil || d < 0 {
			return Mapping{}, fmt.Errorf("mapping.refresh.debounce %q must be a non-negative duration", raw.Refresh.Debounce)
		}
		m.RefreshDebounce = d
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
