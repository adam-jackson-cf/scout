package baseline

import (
	"fmt"

	"scout/internal/domain"
	"scout/internal/state"
)

// PlanShadowImport is pure. A fresh restore or an authenticated whole-history
// fast-forward retains the original source exactly. Every other source remains
// pending; this function cannot authorize its authors or mint destination stamps.
func PlanShadowImport(dst, src domain.ShadowLedger, proof domain.ChainProof) (domain.ShadowImportPlan, error) {
	// A store may have final material without a publication chain. Validate its
	// graph as staged material, but never use its claimed stamps as lineage proof.
	destination := dst
	if len(dst.Publications) == 0 {
		destination.Facts = append([]domain.PRRunFact(nil), dst.Facts...)
		destination.Events = append([]domain.ProviderPREvent(nil), dst.Events...)
		for i := range destination.Facts {
			destination.Facts[i].RecordedAt = nil
		}
		for i := range destination.Events {
			destination.Events[i].RecordedAt = nil
		}
	}
	if err := state.ValidateShadowLedger(destination); err != nil {
		return domain.ShadowImportPlan{}, fmt.Errorf("destination history: %w", err)
	}
	if err := state.ValidateShadowLedger(src); err != nil {
		return domain.ShadowImportPlan{}, fmt.Errorf("source history: %w", err)
	}
	if err := state.ValidateShadowOverlap(dst, src); err != nil {
		return domain.ShadowImportPlan{}, err
	}
	fresh := emptyShadowLedger(dst)
	proven := proof.Authenticated && proof.DestinationAncestor && proof.RepositoryID != "" && proof.StateRef != "" && proof.GenesisCommit != "" && proof.SourceTip != "" && proof.DestinationTip != "" && len(dst.Publications) != 0
	for _, l := range []domain.ShadowLedger{dst, src} {
		for _, p := range l.Publications {
			if p.RepositoryID != proof.RepositoryID {
				proven = false
			}
		}
		for _, f := range l.Facts {
			if f.ID.RepositoryID != proof.RepositoryID || f.RecordedAt == nil {
				proven = false
			}
		}
		for _, e := range l.Events {
			if e.Scope == domain.ShadowScopeCI && e.RepositoryID != proof.RepositoryID {
				proven = false
			}
		}
	}
	if fresh || (proven && state.ShadowHistoryContains(dst, src)) {
		return domain.ShadowImportPlan{BytePreserving: true, Adopt: src}, nil
	}
	return domain.ShadowImportPlan{Pending: src}, nil
}

func emptyShadowLedger(l domain.ShadowLedger) bool {
	return len(l.Facts)+len(l.Events)+len(l.Links)+len(l.Triage)+len(l.Frontiers)+len(l.GapPoints)+len(l.Intervals)+len(l.Publications) == 0
}

// mergeShadow keeps the current published authority intact. Without a provider
// chain proof even a byte-prefix source cannot grant itself fast-forward status.
// Pending import IDs are content-addressed for deterministic retry/replay.
func mergeShadow(current, local Tree, merged Tree) error {
	dst, dstInputs, err := state.DecodeShadow(current)
	if err != nil {
		return err
	}
	src, srcInputs, err := state.DecodeShadow(local)
	if err != nil {
		return err
	}
	if emptyShadowLedger(dst) && len(dstInputs.Obligations)+len(dstInputs.Accounting)+len(dstInputs.ImportRefusals)+len(dstInputs.PendingImports) == 0 {
		files, err := state.EncodeShadow(src, srcInputs)
		if err != nil {
			return err
		}
		for path, data := range files {
			merged[path] = data
		}
		return nil
	}
	// Exact already-present material is a replay, not a new import publication.
	if !state.ShadowHistoryContains(src, dst) || !state.ShadowInputsContains(srcInputs, dstInputs) {
		if err := state.ValidateShadowOverlap(dst, src); err != nil {
			return err
		}
		pending, err := state.PendingShadowImport(src, srcInputs)
		if err != nil {
			return err
		}
		found := false
		for _, existing := range dstInputs.PendingImports {
			if existing.ID == pending.ID {
				found = true
				break
			}
		}
		if !found {
			dstInputs.PendingImports = append(dstInputs.PendingImports, pending)
		}
	}
	files, err := state.EncodeShadow(dst, dstInputs)
	if err != nil {
		return err
	}
	for path, data := range files {
		merged[path] = data
	}
	// Revalidate the union too: a valid source and destination can conflict.
	_, _, err = state.DecodeShadow(merged)
	return err
}
