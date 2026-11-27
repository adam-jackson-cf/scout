package main

import (
	"context"
	"errors"

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
	all, err := store.ListAreas(ctx)
	if err = errors.Join(err, store.Close()); err != nil {
		return nil, err
	}
	return onlyPendingAreas(all), nil
}
