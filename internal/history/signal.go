package history

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"scout/internal/domain"
)

// The only file that touches domain.Signal's field layout.

func mkSignal(sc Scope, w Window, o Options, metric string, population []string) domain.Signal {
	id, ver, _ := strings.Cut(metric, "@")
	v, err := strconv.Atoi(ver)
	if err != nil { // metric IDs are package constants of the form name@version
		panic(fmt.Sprintf("history metric %q has no integer version: %v", metric, err))
	}
	return domain.Signal{
		MetricID:         id,
		Version:          v,
		ScopeKind:        sc.Kind,
		ScopeAnchor:      sc.Anchor,
		Snapshot:         o.Snapshot,
		Window:           w.Label,
		PopulationDigest: popDigest(metric, population),
		ToolDigest:       o.toolDigest(metric),
		ConfigDigest:     o.configDigest(),
		Claim:            domain.ClaimDescriptive,
		Detail: map[string]any{
			"window_start": w.Start.Format(time.RFC3339),
			"window_end":   w.End.Format(time.RFC3339), // exclusive
			"immature":     w.Immature,                 // right-censored: do not compare with complete windows
		},
	}
}

func setValue(s domain.Signal, value, num, den float64) domain.Signal {
	s.Value, s.Numerator, s.Denominator = domain.Float(value), domain.Float(num), domain.Float(den)
	return s
}

// markMissing turns a signal into unmeasured/insufficient: numbers are cleared
// (domain.Signal.Validate forbids them) and the reason recorded.
func markMissing(s domain.Signal, kind, reason string) domain.Signal {
	s.Value, s.Numerator, s.Denominator = nil, nil, nil
	s.Missing, s.MissingReason = domain.SignalMissing(kind), reason
	return s
}

func withDetail(s domain.Signal, d map[string]any) domain.Signal {
	if s.Detail == nil {
		s.Detail = map[string]any{}
	}
	for k, v := range d {
		s.Detail[k] = v
	}
	return s
}
