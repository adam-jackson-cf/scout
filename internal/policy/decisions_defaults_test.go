package policy

import (
	"strings"
	"testing"
)

func TestDefaultModelIsAPinnedJev114Slug(t *testing.T) {
	if DefaultModel != "typesafe/jev-1.14-20261014" || IsModelAlias(DefaultModel) {
		t.Fatalf("default model %q must be the dated Jev 1.14 OpenRouter slug and not an alias", DefaultModel)
	}
}

func TestIsModelAliasRejectsFloatingNamesAndUndatedReleaseLines(t *testing.T) {
	for _, model := range []string{"jev-latest", "typesafe/jev-latest", "jev-preview", "typesafe/jev-preview", "auto", "x:latest", "typesafe/jev-1.14", "jev-1.13", "TypeSafe/Jev-1.14"} {
		if !IsModelAlias(model) {
			t.Errorf("%q must be an alias", model)
		}
	}
	for _, model := range []string{"typesafe/jev-1.14-20261014", "typesafe/jev-1.13-20260917", "jev-1.14.0", "jev-2026-10-01"} {
		if IsModelAlias(model) {
			t.Errorf("%q is an exact build and must be accepted", model)
		}
	}
}

func TestDecisionsCostCapDefaultsOverridesAndRejectsNegatives(t *testing.T) {
	p, err := Parse([]byte("profile: report\n"))
	if err != nil || p.Decisions().MaxCostUSD != DefaultMaxCostUSD {
		t.Fatalf("default %+v %v", p.Decisions(), err)
	}
	p, err = Parse([]byte("profile: report\ndecisions:\n  enabled: true\n  model: typesafe/jev-1.14-20261014\n  max_cost_usd: 0.05\n  egress_acknowledged: true\n"))
	if err != nil || p.Decisions().MaxCostUSD != 0.05 || p.Decisions().MaxPerRun != DefaultMaxPerRun {
		t.Fatalf("override %+v %v", p.Decisions(), err)
	}
	if _, err := Parse([]byte("profile: report\ndecisions:\n  max_cost_usd: -1\n")); err == nil || !strings.Contains(err.Error(), "max_cost_usd") {
		t.Fatalf("negative cap accepted: %v", err)
	}
	if _, err := Parse([]byte("profile: report\ndecisions:\n  enabled: true\n  model: typesafe/jev-1.14\n  egress_acknowledged: true\n")); err == nil {
		t.Fatal("an undated release line must be rejected at load")
	}
}

func TestApplyConsentWritesTheCostCap(t *testing.T) {
	cost := 0.25
	out, changed, err := ApplyConsent(nil, ConsentUpdate{MaxCostUSD: &cost})
	if err != nil || !changed || !strings.Contains(string(out), "max_cost_usd: 0.25") {
		t.Fatalf("%s %v %v", out, changed, err)
	}
	again, changed, err := ApplyConsent(out, ConsentUpdate{MaxCostUSD: &cost})
	if err != nil || changed || string(again) != string(out) {
		t.Fatalf("not idempotent: %v %v", changed, err)
	}
}
