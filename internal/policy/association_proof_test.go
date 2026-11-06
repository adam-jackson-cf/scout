package policy

import (
	"fmt"
	"strings"
	"testing"
)

func TestAssociationPolicyDefaults(t *testing.T) {
	got := mustParse(t, "").Mapping().Association
	if got.Enabled || got.CoveredScore != .60 || got.TChoice != .535 {
		t.Fatalf("defaults %+v", got)
	}
	if got != DefaultMapping().Association {
		t.Fatalf("parse/default diverge: %+v", got)
	}
}

func TestAssociationPolicyParse(t *testing.T) {
	got := mustParse(t, "mapping:\n  association:\n    enabled: true\n    covered_score: 0.75\n    t_choice: 0.625\n").Mapping().Association
	if !got.Enabled || got.CoveredScore != .75 || got.TChoice != .625 {
		t.Fatalf("parsed %+v", got)
	}
	// Explicit false and boundary zero must not be confused with absent values.
	for _, bound := range []string{"0", "1"} {
		p := mustParse(t, fmt.Sprintf("mapping: {association: {enabled: false, covered_score: %s, t_choice: %s}}", bound, bound)).Mapping().Association
		want := 0.
		if bound == "1" {
			want = 1
		}
		if p.Enabled || p.CoveredScore != want || p.TChoice != want {
			t.Fatalf("boundary %s: %+v", bound, p)
		}
	}
}

func TestAssociationPolicyBounds(t *testing.T) {
	for _, key := range []string{"covered_score", "t_choice"} {
		for _, value := range []string{"-0.01", "1.01", ".nan", ".inf", "-.inf"} {
			t.Run(key+"/"+value, func(t *testing.T) {
				_, err := Parse([]byte(fmt.Sprintf("mapping: {association: {%s: %s}}", key, value)))
				if err == nil || !strings.Contains(err.Error(), "mapping.association."+key) {
					t.Fatalf("%s=%s error %v", key, value, err)
				}
			})
		}
	}
	for _, value := range []string{"[true]", "{bad: true}", "not-a-bool"} {
		if _, err := Parse([]byte("mapping: {association: {enabled: " + value + "}}")); err == nil {
			t.Fatalf("enabled accepts %s", value)
		}
	}
}
