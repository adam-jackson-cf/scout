// Package engine coordinates the persisted assessment without modifying target source.
package engine

import (
	"scout/internal/analyzers"
	"scout/internal/clawpatch"
	"scout/internal/contextual"
	"scout/internal/domain"
	"scout/internal/inventory"
)

type Config struct {
	Root                      string
	ClawpatchBinary           string
	ClawpatchEgressAuthorized bool
	Provider                  contextual.Config
	Registry                  *analyzers.Registry
	Dependencies              []inventory.DependencyEdge
}

type Options struct {
	Scope      domain.Scope
	ForceFull  bool
	Exclusions []string
}

type Result struct {
	Run         domain.Run             `json:"run"`
	Plan        domain.Plan            `json:"plan"`
	Diff        inventory.DiffResult   `json:"diff"`
	Map         clawpatch.MapResult    `json:"map"`
	MappingGaps []inventory.MappingGap `json:"mappingGaps,omitempty"`
	Adapters    []domain.AdapterResult `json:"adapters,omitempty"`
	Issues      []domain.Issue         `json:"issues,omitempty"`
}

// Capability reports whether a prerequisite for the selected work is available.
type Capability struct {
	Name     string `json:"name"`
	Required bool   `json:"required"`
	Ready    bool   `json:"ready"`
	Reason   string `json:"reason,omitempty"`
}

// Readiness describes prerequisites for work that would actually execute.
type Readiness struct {
	Ready        bool         `json:"ready"`
	Capabilities []Capability `json:"capabilities"`
}

// Contextual categories are independent of concrete scanner rule categories.
// An unavailable provider leaves their feature units blocked, not clean.
var contextualCategories = map[domain.Family][]string{
	domain.FamilyArchitecture: {"domain_alignment", "team_ownership", "cross_cutting_concerns", "runtime_deployment_boundaries"},
	domain.FamilySecurity:     {"business_logic_authorization", "implicit_trust_boundaries", "sensitive_data_flows", "configuration_supply_chain"},
	domain.FamilyQuality:      {"semantic_clarity", "abstraction_level", "domain_model_fit", "language_runtime_idioms"},
	domain.FamilyPerformance:  {"algorithmic_complexity", "cache_invalidation", "async_coordination", "resource_query_efficiency"},
}
