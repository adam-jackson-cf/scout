package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"scout/internal/domain"
)

func recordShadowInputs(ctx context.Context, tx *sql.Tx, inputs domain.ShadowInputs) error {
	if err := recordShadowObligationsTx(ctx, tx, inputs.Obligations); err != nil {
		return err
	}
	for _, a := range inputs.Accounting {
		if err := recordShadowAccountingTx(ctx, tx, a); err != nil {
			return err
		}
	}
	if err := recordShadowRefusalsTx(ctx, tx, inputs.ImportRefusals); err != nil {
		return err
	}
	for _, p := range inputs.PendingImports {
		if err := validatePendingShadowImport(p); err != nil {
			return err
		}
		if err := insertShadowJSON(ctx, tx, shadowRow{"shadow_pending_imports", []string{"scope", "repository_id", "import_id"}, []any{p.Scope, p.RepositoryID, p.ID}}, p); err != nil {
			return err
		}
	}
	return nil
}

func recordShadowObligationsTx(ctx context.Context, tx *sql.Tx, obligations []domain.ProviderEventObligation) error {
	for _, o := range obligations {
		if !validShadowObligation(o) {
			return errors.New("invalid provider event obligation")
		}
		if err := insertShadowJSON(ctx, tx, shadowRow{"shadow_event_obligations", []string{"scope", "repository_id", "event_id"}, []any{o.Scope, o.RepositoryID, o.ID}}, o); err != nil {
			return err
		}
	}
	return nil
}

func validShadowObligation(o domain.ProviderEventObligation) bool {
	return o.ID != "" && o.RepositoryID != "" && validShadowScope(o.Scope) && shadowDecimal(o.PRNumber) && o.HeadCommit != "" && shadowTime(o.EventAt) && o.Selector != "" && o.ProviderProof != ""
}

func recordShadowAccountingTx(ctx context.Context, tx *sql.Tx, a domain.PRRunObligationLink) error {
	if err := validateAttempt(a.AttemptID); err != nil {
		return err
	}
	if !validShadowScope(a.Scope) || a.ObligationID == "" || a.CheckID == "" {
		return errors.New("invalid provider event accounting")
	}
	f, o, err := accountedFactAndObligationTx(ctx, tx, a)
	if err != nil {
		return err
	}
	if f.PRNumber != o.PRNumber || f.HeadCommit != o.HeadCommit || f.StartedAt.Before(o.EventAt) {
		return errors.New("attempt cannot account a different PR/head or a future provider event")
	}
	return insertShadowJSON(ctx, tx, shadowRow{"pr_run_obligations", []string{"scope", "repository_id", "attempt_id", "check_id", "event_id"}, []any{a.Scope, a.AttemptID.RepositoryID, shadowAttemptKey(a.AttemptID), a.CheckID, a.ObligationID}}, a)
}

func accountedFactAndObligationTx(ctx context.Context, tx *sql.Tx, a domain.PRRunObligationLink) (domain.PRRunFact, domain.ProviderEventObligation, error) {
	var f domain.PRRunFact
	var o domain.ProviderEventObligation
	b, err := shadowPayload(ctx, tx, "pr_run_facts", []string{"scope", "repository_id", "attempt_id", "check_id"}, []any{a.Scope, a.AttemptID.RepositoryID, shadowAttemptKey(a.AttemptID), a.CheckID})
	if err != nil {
		return f, o, fmt.Errorf("accounting attempt: %w", err)
	}
	if err := json.Unmarshal(b, &f); err != nil {
		return f, o, err
	}
	b, err = shadowPayload(ctx, tx, "shadow_event_obligations", []string{"scope", "repository_id", "event_id"}, []any{a.Scope, a.AttemptID.RepositoryID, a.ObligationID})
	if err != nil {
		return f, o, fmt.Errorf("accounting obligation: %w", err)
	}
	if err := json.Unmarshal(b, &o); err != nil {
		return f, o, err
	}
	return f, o, nil
}

func recordShadowRefusalsTx(ctx context.Context, tx *sql.Tx, refusals []domain.ShadowImportRefusal) error {
	for _, r := range refusals {
		if !validImportRefusal(r) {
			return errors.New("invalid import refusal provenance")
		}
		if err := rejectUnsafePersistence(r.Reason); err != nil {
			return err
		}
		if err := insertShadowJSON(ctx, tx, shadowRow{"shadow_import_refusals", []string{"scope", "repository_id", "refusal_id"}, []any{r.Scope, r.RepositoryID, r.ID}}, r); err != nil {
			return err
		}
	}
	return nil
}

func validImportRefusal(r domain.ShadowImportRefusal) bool {
	return validRefusalRecord(r) && validRefusalOriginal(r)
}

func validRefusalRecord(r domain.ShadowImportRefusal) bool {
	return r.ID != "" && r.RepositoryID != "" && r.Scope == domain.ShadowScopeCI && r.SourceLineage != "" && r.Reason != "" && r.PolicyRevision != "" && shadowTime(r.RecordedAt)
}

func validRefusalOriginal(r domain.ShadowImportRefusal) bool {
	return r.Original.ActionID != "" && r.Original.EventID != "" && r.Original.RepositoryID == r.RepositoryID && r.Original.Scope == domain.ShadowScopeCI
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
