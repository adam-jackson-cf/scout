package main

import (
	"context"
	"scout/internal/areas"
	"scout/internal/clawpatch"
	"scout/internal/domain"
	"scout/internal/engine"
	"scout/internal/inventory"
)

type assessmentReport struct {
	Run           domain.Run             `json:"run"`
	Plan          domain.Plan            `json:"plan"`
	Diff          inventory.DiffResult   `json:"diff"`
	MappingStatus clawpatch.MapStatus    `json:"mappingStatus"`
	MappingGaps   []inventory.MappingGap `json:"mappingGaps"`
	Issues        []domain.Issue         `json:"issues"`
	Adapters      []domain.AdapterResult `json:"adapters"`
	Gate          domain.Gate            `json:"gate"`
	Health        []domain.LaneHealth    `json:"health"`
	WouldBlock    []domain.WouldBlock    `json:"would_block"`
	Ranking       engine.Ranking         `json:"ranking"`
	PendingAreas  []areas.Area           `json:"pendingAreas"`
	Associations  []domain.Association   `json:"associations"`
	associationGuardReport
}

func pendingAssessmentAreas(ctx context.Context, flags *commonFlags) ([]areas.Area, error) {
	store, err := openStore(flags)
	if err != nil {
		return nil, err
	}
	defer store.Close()
	all, err := store.ListAreas(ctx)
	if err != nil {
		return nil, err
	}
	pending := make([]areas.Area, 0)
	for _, area := range all {
		if area.Status == areas.StatusPending {
			pending = append(pending, area)
		}
	}
	return pending, nil
}
