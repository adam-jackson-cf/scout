package engine

import (
	"strings"

	"scout/internal/domain"
)

func short(value string) string {
	if value == "" {
		return "(empty)"
	}
	if len(value) > 12 {
		return value[:12]
	}
	return value
}

// comparisonReason explains why a comparison is not the imported baseline.
func comparisonReason(comparison domain.Comparison, banners []string) string {
	if comparison == domain.ComparisonBaseline {
		return ""
	}
	return strings.Join(banners, "; ")
}
