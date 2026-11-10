package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"scout/internal/domain"
)

// RecordShadowInputs atomically inserts supplemental control-plane inputs.
// Exact replay is a no-op; every changed immutable identity is rejected. In
// particular this does not authorize imports or resolve pending imports: those
// decisions are represented by later trusted publications, never row updates.
// Import refusals are trusted-writer records, so callers must authenticate their
// source and current policy before passing them here.
func (s *Store) RecordShadowInputs(ctx context.Context, inputs domain.ShadowInputs) error {
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		return recordShadowInputs(ctx, tx, inputs)
	})
}

func recordShadowInputs(ctx context.Context, tx *sql.Tx, inputs domain.ShadowInputs) error {
	for _, o := range inputs.Obligations {
		if o.ID == "" || o.RepositoryID == "" || !validShadowScope(o.Scope) || !shadowDecimal(o.PRNumber) || o.HeadCommit == "" || !shadowTime(o.EventAt) || o.Selector == "" || o.ProviderProof == "" {
			return errors.New("invalid provider event obligation")
		}
		if err := insertShadowJSON(ctx, tx, "shadow_event_obligations", []string{"scope", "repository_id", "event_id"}, []any{o.Scope, o.RepositoryID, o.ID}, o); err != nil {
			return err
		}
	}
	for _, a := range inputs.Accounting {
		if err := validateAttempt(a.AttemptID); err != nil {
			return err
		}
		if !validShadowScope(a.Scope) || a.ObligationID == "" || a.CheckID == "" {
			return errors.New("invalid provider event accounting")
		}
		b, err := shadowPayload(ctx, tx, "pr_run_facts", []string{"scope", "repository_id", "attempt_id", "check_id"}, []any{a.Scope, a.AttemptID.RepositoryID, shadowAttemptKey(a.AttemptID), a.CheckID})
		if err != nil {
			return fmt.Errorf("accounting attempt: %w", err)
		}
		var f domain.PRRunFact
		if err := json.Unmarshal(b, &f); err != nil {
			return err
		}
		b, err = shadowPayload(ctx, tx, "shadow_event_obligations", []string{"scope", "repository_id", "event_id"}, []any{a.Scope, a.AttemptID.RepositoryID, a.ObligationID})
		if err != nil {
			return fmt.Errorf("accounting obligation: %w", err)
		}
		var o domain.ProviderEventObligation
		if err := json.Unmarshal(b, &o); err != nil {
			return err
		}
		if f.PRNumber != o.PRNumber || f.HeadCommit != o.HeadCommit || f.StartedAt.Before(o.EventAt) {
			return errors.New("attempt cannot account a different PR/head or a future provider event")
		}
		if err := insertShadowJSON(ctx, tx, "pr_run_obligations", []string{"scope", "repository_id", "attempt_id", "check_id", "event_id"}, []any{a.Scope, a.AttemptID.RepositoryID, shadowAttemptKey(a.AttemptID), a.CheckID, a.ObligationID}, a); err != nil {
			return err
		}
	}
	for _, r := range inputs.ImportRefusals {
		if r.ID == "" || r.RepositoryID == "" || r.Scope != domain.ShadowScopeCI || r.SourceLineage == "" || r.Original.ActionID == "" || r.Original.EventID == "" || r.Original.RepositoryID != r.RepositoryID || r.Original.Scope != domain.ShadowScopeCI || r.Reason == "" || r.PolicyRevision == "" || !shadowTime(r.RecordedAt) {
			return errors.New("invalid import refusal provenance")
		}
		if err := rejectUnsafePersistence(r.Reason); err != nil {
			return err
		}
		if err := insertShadowJSON(ctx, tx, "shadow_import_refusals", []string{"scope", "repository_id", "refusal_id"}, []any{r.Scope, r.RepositoryID, r.ID}, r); err != nil {
			return err
		}
	}
	for _, p := range inputs.PendingImports {
		if err := validatePendingShadowImport(p); err != nil {
			return err
		}
		if err := insertShadowJSON(ctx, tx, "shadow_pending_imports", []string{"scope", "repository_id", "import_id"}, []any{p.Scope, p.RepositoryID, p.ID}, p); err != nil {
			return err
		}
	}
	return nil
}

// ReadShadowInputs returns one consistent snapshot of the insert-only input
// tables. Pending material is not published CI evidence, regardless of source
// stamps. The consumer must prove ancestry or obtain current authorization.
func (s *Store) ReadShadowInputs(ctx context.Context) (domain.ShadowInputs, error) {
	var inputs domain.ShadowInputs
	err := s.withRead(ctx, func(db *sql.DB) error {
		tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if err := readShadowInputsTx(ctx, tx, &inputs); err != nil {
			return err
		}
		return tx.Commit()
	})
	return inputs, err
}

func readShadowInputsTx(ctx context.Context, tx *sql.Tx, inputs *domain.ShadowInputs) error {
	if err := readShadowRows(ctx, tx, "shadow_event_obligations", "scope,repository_id,event_id", &inputs.Obligations); err != nil {
		return err
	}
	if err := readShadowRows(ctx, tx, "pr_run_obligations", "scope,repository_id,attempt_id,check_id,event_id", &inputs.Accounting); err != nil {
		return err
	}
	if err := readShadowRows(ctx, tx, "shadow_import_refusals", "scope,repository_id,refusal_id", &inputs.ImportRefusals); err != nil {
		return err
	}
	if err := readShadowRows(ctx, tx, "shadow_pending_imports", "scope,repository_id,import_id", &inputs.PendingImports); err != nil {
		return err
	}
	return nil
}
