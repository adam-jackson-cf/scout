package baseline

import (
	"slices"

	"scout/internal/domain"
	"scout/internal/state"
)

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
	if emptyShadowLedger(dst) && emptyShadowInputs(dstInputs) {
		return writeShadow(merged, src, srcInputs)
	}
	// Exact already-present material is a replay, not a new import publication.
	if !state.ShadowHistoryContains(src, dst) || !state.ShadowInputsContains(srcInputs, dstInputs) {
		if dstInputs, err = stagePendingImport(dst, src, srcInputs, dstInputs); err != nil {
			return err
		}
	}
	if err := writeShadow(merged, dst, dstInputs); err != nil {
		return err
	}
	// Revalidate the union too: a valid source and destination can conflict.
	_, _, err = state.DecodeShadow(merged)
	return err
}

func emptyShadowInputs(in domain.ShadowInputs) bool {
	return len(in.Obligations)+len(in.Accounting)+len(in.ImportRefusals)+len(in.PendingImports) == 0
}

func writeShadow(merged Tree, l domain.ShadowLedger, in domain.ShadowInputs) error {
	files, err := state.EncodeShadow(l, in)
	if err != nil {
		return err
	}
	for path, data := range files {
		merged[path] = data
	}
	return nil
}

func stagePendingImport(dst, src domain.ShadowLedger, srcInputs, dstInputs domain.ShadowInputs) (domain.ShadowInputs, error) {
	if err := state.ValidateShadowOverlap(dst, src); err != nil {
		return dstInputs, err
	}
	pending, err := state.PendingShadowImport(src, srcInputs)
	if err != nil {
		return dstInputs, err
	}
	found := slices.ContainsFunc(dstInputs.PendingImports, func(existing domain.ShadowPendingImport) bool { return existing.ID == pending.ID })
	if !found {
		dstInputs.PendingImports = append(dstInputs.PendingImports, pending)
	}
	return dstInputs, nil
}
