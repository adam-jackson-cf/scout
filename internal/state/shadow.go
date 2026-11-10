package state

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"scout/internal/domain"
)

// RecordPRRun stages immutable final assessment material. It cannot create
// publication stamps or completeness; only ApplyShadowPublication can do that.
func (s *Store) RecordPRRun(ctx context.Context, fact domain.PRRunFact, events []domain.ProviderPREvent, links []domain.PRRunEventLink) error {
	if fact.RecordedAt != nil {
		return errors.New("pending fact cannot carry recorded_at")
	}
	for _, event := range events {
		if event.RecordedAt != nil {
			return errors.New("pending event cannot carry recorded_at")
		}
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		if err := storeShadowFact(ctx, tx, fact, false); err != nil {
			return err
		}
		for _, event := range events {
			if event.RepositoryID != fact.ID.RepositoryID || event.Scope != fact.Scope {
				return errors.New("event repository/scope differs from attempt")
			}
			if err := storeShadowEvent(ctx, tx, event, false); err != nil {
				return err
			}
		}
		for _, link := range links {
			if link.AttemptID.RepositoryID != fact.ID.RepositoryID || link.Scope != fact.Scope {
				return errors.New("link repository/scope differs from attempt")
			}
			if err := storeShadowLink(ctx, tx, link); err != nil {
				return err
			}
		}
		return nil
	})
}

// ReadShadowLedger returns a single consistent snapshot, including pending facts.
// Consumers must select explicit CI scope and trusted stamps, not infer authority
// from the existence of a row. No frontier means collection is unmeasured.
func (s *Store) ReadShadowLedger(ctx context.Context) (domain.ShadowLedger, error) {
	var ledger domain.ShadowLedger
	err := s.withRead(ctx, func(db *sql.DB) error {
		tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if err := readShadowLedgerTx(ctx, tx, &ledger); err != nil {
			return err
		}
		return tx.Commit()
	})
	return ledger, err
}

func readShadowLedgerTx(ctx context.Context, tx *sql.Tx, ledger *domain.ShadowLedger) error {
	if err := readShadowRows(ctx, tx, "pr_run_facts", "scope,repository_id,attempt_id,check_id", &ledger.Facts); err != nil {
		return err
	}
	if err := readShadowRows(ctx, tx, "shadow_events", "scope,repository_id,event_id", &ledger.Events); err != nil {
		return err
	}
	if err := readShadowRows(ctx, tx, "pr_run_events", "scope,repository_id,attempt_id,event_id", &ledger.Links); err != nil {
		return err
	}
	if err := readShadowRows(ctx, tx, "would_block_triage", "scope,repository_id,action_id", &ledger.Triage); err != nil {
		return err
	}
	if err := readShadowRows(ctx, tx, "shadow_frontiers", "repository_id,complete_through,check_id,rule_version", &ledger.Frontiers); err != nil {
		return err
	}
	if err := readShadowRows(ctx, tx, "shadow_gap_points", "repository_id,point_id", &ledger.GapPoints); err != nil {
		return err
	}
	if err := readShadowRows(ctx, tx, "shadow_intervals", "repository_id,interval_id", &ledger.Intervals); err != nil {
		return err
	}
	if err := readShadowRows(ctx, tx, "shadow_publications", "repository_id,recorded_at,publication_id", &ledger.Publications); err != nil {
		return err
	}
	return nil
}

// ApplyShadowPublication atomically appends control-plane-assembled history. It
// enforces storage invariants, not provider authentication or CAS authorization:
// callers must establish those before invoking this trusted-writer API.
func (s *Store) ApplyShadowPublication(ctx context.Context, p domain.ShadowPublication) error {
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		return applyShadowPublication(ctx, tx, p)
	})
}

func applyShadowPublication(ctx context.Context, tx *sql.Tx, p domain.ShadowPublication) error {
	if p.ID == "" || p.RepositoryID == "" || !shadowTime(p.RecordedAt) {
		return errors.New("publication identity and UTC millisecond stamp required")
	}
	// Check exact replay before checking current parent: retries cannot reopen a
	// completed action or require the old parent still to be the latest tip.
	var existing []byte
	err := tx.QueryRowContext(ctx, `SELECT payload FROM shadow_publications WHERE repository_id=? AND publication_id=?`, p.RepositoryID, p.ID).Scan(&existing)
	if err == nil {
		return equalShadowJSON(existing, p)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	var parent sql.NullString
	if p.ParentID != "" {
		parent = sql.NullString{String: p.ParentID, Valid: true}
		var stamp int64
		if err := tx.QueryRowContext(ctx, `SELECT recorded_at FROM shadow_publications WHERE repository_id=? AND publication_id=?`, p.RepositoryID, p.ParentID).Scan(&stamp); err != nil {
			return fmt.Errorf("publication parent: %w", err)
		}
		if p.RecordedAt.UnixMilli() <= stamp {
			return errors.New("publication stamp must increase")
		}
	}
	if err := insertShadowJSON(ctx, tx, "shadow_publications", []string{"repository_id", "publication_id", "parent_id", "recorded_at"}, []any{p.RepositoryID, p.ID, parent, p.RecordedAt.UnixMilli()}, p); err != nil {
		return err
	}
	for _, f := range p.Facts {
		if f.Scope != domain.ShadowScopeCI || f.ID.RepositoryID != p.RepositoryID || !sameShadowStamp(f.RecordedAt, p.RecordedAt) {
			return errors.New("publication fact requires CI repository and publication stamp")
		}
		if err := storeShadowFact(ctx, tx, f, true); err != nil {
			return err
		}
	}
	for _, e := range p.Events {
		if e.Scope != domain.ShadowScopeCI || e.RepositoryID != p.RepositoryID || !sameShadowStamp(e.RecordedAt, p.RecordedAt) {
			return errors.New("publication event requires CI repository and publication stamp")
		}
		if err := storeShadowEvent(ctx, tx, e, true); err != nil {
			return err
		}
	}
	for _, link := range p.Links {
		if link.Scope != domain.ShadowScopeCI || link.AttemptID.RepositoryID != p.RepositoryID {
			return errors.New("publication link repository/scope mismatch")
		}
		if err := storeShadowLink(ctx, tx, link); err != nil {
			return err
		}
	}
	seen := make(map[string]bool, len(p.Triage))
	for _, row := range p.Triage {
		if seen[row.EventID] {
			return errors.New("only one triage row per event per publication")
		}
		seen[row.EventID] = true
		// Look for immutable collisions first, including attempted provenance edits.
		var old []byte
		err := tx.QueryRowContext(ctx, `SELECT payload FROM would_block_triage WHERE scope=? AND repository_id=? AND action_id=?`, row.Scope, row.RepositoryID, row.ActionID).Scan(&old)
		if err == nil {
			if err := equalShadowJSON(old, row); err != nil {
				return err
			}
			return fmt.Errorf("triage action already published: %w", ErrImmutable)
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if row.Scope != domain.ShadowScopeCI || row.RepositoryID != p.RepositoryID || row.ActionID == "" || row.EventID == "" || row.Actor == "" || row.PolicyRevision == "" || !domain.ValidTriage(row.Verdict) || !shadowTime(row.TriagedAt) || !sameShadowStamp(row.RecordedAt, p.RecordedAt) {
			return errors.New("invalid authorized triage provenance")
		}
		if err := rejectUnsafePersistence(row.Actor); err != nil {
			return err
		}
		if err := rejectUnsafePersistence(row.Reason); err != nil {
			return err
		}
		if err := insertShadowJSON(ctx, tx, "would_block_triage", []string{"scope", "repository_id", "action_id", "event_id", "publication_id"}, []any{row.Scope, row.RepositoryID, row.ActionID, row.EventID, p.ID}, row); err != nil {
			return err
		}
	}
	for _, f := range p.Frontiers {
		if f.RepositoryID != p.RepositoryID || f.CheckID == "" || f.RuleVersion == "" || f.PolicyRevision == "" || f.MinDays < 0 || f.MinObservations < 0 || !shadowTime(f.CompleteThrough) {
			return errors.New("invalid frontier policy")
		}
		keys := []string{"repository_id", "check_id", "rule_version", "complete_through"}
		values := []any{f.RepositoryID, f.CheckID, f.RuleVersion, f.CompleteThrough.UnixMilli()}
		old, err := shadowPayload(ctx, tx, "shadow_frontiers", keys, values)
		if err == nil {
			if err := equalShadowJSON(old, f); err != nil {
				return err
			}
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		var latest sql.NullInt64
		if err := tx.QueryRowContext(ctx, `SELECT max(complete_through) FROM shadow_frontiers WHERE repository_id=? AND check_id=? AND rule_version=?`, f.RepositoryID, f.CheckID, f.RuleVersion).Scan(&latest); err != nil {
			return err
		}
		if latest.Valid && f.CompleteThrough.UnixMilli() <= latest.Int64 {
			return errors.New("frontier must strictly increase")
		}
		if err := insertShadowJSON(ctx, tx, "shadow_frontiers", keys, values, f); err != nil {
			return err
		}
	}
	for _, g := range p.GapPoints {
		if g.RepositoryID != p.RepositoryID || g.ID == "" || g.CheckID == "" || g.Reason == "" || !shadowTime(g.At) || !sameShadowStamp(g.RecordedAt, p.RecordedAt) {
			return errors.New("invalid gap point")
		}
		if err := insertShadowJSON(ctx, tx, "shadow_gap_points", []string{"repository_id", "point_id"}, []any{g.RepositoryID, g.ID}, g); err != nil {
			return err
		}
	}
	for _, i := range p.Intervals {
		if i.RepositoryID != p.RepositoryID || i.ID == "" || i.CheckID == "" || i.RuleVersion == "" || !shadowTime(i.After) || !shadowTime(i.Through) || !i.Through.After(i.After) {
			return errors.New("invalid complete interval")
		}
		if err := insertShadowJSON(ctx, tx, "shadow_intervals", []string{"repository_id", "interval_id"}, []any{i.RepositoryID, i.ID}, i); err != nil {
			return err
		}
	}
	return nil
}

func shadowTime(t time.Time) bool {
	return !t.IsZero() && t.Location() == time.UTC && t.Nanosecond()%int(time.Millisecond) == 0
}
func sameShadowStamp(t *time.Time, stamp time.Time) bool {
	return t != nil && shadowTime(*t) && t.Equal(stamp)
}
func validShadowScope(scope domain.ShadowScope) bool {
	return scope == domain.ShadowScopeCI || scope == domain.ShadowScopeLocal || scope == domain.ShadowScopeLegacy
}
func shadowDecimal(id string) bool {
	if id == "" || (len(id) > 1 && id[0] == '0') {
		return false
	}
	for _, c := range id {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
func validateAttempt(id domain.AttemptID) error {
	if id.RepositoryID == "" || !shadowDecimal(id.RunID) || !shadowDecimal(id.JobID) || !shadowDecimal(id.Attempt) {
		return errors.New("provider identity requires exact canonical decimal run/job/attempt IDs")
	}
	return nil
}
func shadowAttemptKey(id domain.AttemptID) string {
	// An array encoding is unambiguous even when opaque repository IDs contain
	// punctuation. IDs remain text; no integer width or floating point is involved.
	b, _ := json.Marshal([3]string{id.RunID, id.JobID, id.Attempt})
	return string(b)
}
func validateShadowFact(f domain.PRRunFact) error {
	if err := validateAttempt(f.ID); err != nil {
		return err
	}
	if !validShadowScope(f.Scope) || !shadowDecimal(f.PRNumber) || f.HeadCommit == "" || f.HeadTree == "" || f.BaseCommit == "" || f.CheckID == "" || f.RuleVersion == "" || f.Status == "" || f.Comparison == "" || f.Inspection == "" || !shadowTime(f.CreatedAt) || !shadowTime(f.StartedAt) || !shadowTime(f.FinishedAt) || f.StartedAt.Before(f.CreatedAt) || f.FinishedAt.Before(f.StartedAt) {
		return errors.New("invalid final PR run fact")
	}
	if f.RecordedAt != nil && (f.Scope != domain.ShadowScopeCI || !shadowTime(*f.RecordedAt)) {
		return errors.New("invalid fact stamp/scope")
	}
	if f.Scope == domain.ShadowScopeCI && (f.Pins.Scout == "" || f.Pins.Tools == "" || f.Pins.Catalogue == "" || f.Pins.Policy == "" || f.Pins.Recipe == "") {
		return errors.New("CI fact requires pinned recipe provenance")
	}
	return nil
}

func storeShadowFact(ctx context.Context, tx *sql.Tx, f domain.PRRunFact, publish bool) error {
	if err := validateShadowFact(f); err != nil {
		return err
	}
	keys := []string{"scope", "repository_id", "attempt_id", "check_id"}
	values := []any{f.Scope, f.ID.RepositoryID, shadowAttemptKey(f.ID), f.CheckID}
	old, err := shadowPayload(ctx, tx, "pr_run_facts", keys, values)
	if err == nil {
		var prior domain.PRRunFact
		if err := json.Unmarshal(old, &prior); err != nil {
			return err
		}
		stamp := prior.RecordedAt
		prior.RecordedAt = nil
		bare := f
		bare.RecordedAt = nil
		if prior != bare {
			return ErrImmutable
		}
		if f.RecordedAt == nil {
			return nil
		}
		if stamp != nil {
			if stamp.Equal(*f.RecordedAt) {
				return nil
			}
			return ErrImmutable
		}
		if !publish {
			return errors.New("only publication may stamp facts")
		}
		return updateShadowStamp(ctx, tx, "pr_run_facts", keys, values, f)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	return insertShadowJSON(ctx, tx, "pr_run_facts", keys, values, f)
}
func validateShadowEvent(e domain.ProviderPREvent) error {
	if !validShadowScope(e.Scope) || e.EventID == "" || e.RepositoryID == "" || !shadowDecimal(e.PRNumber) || e.HeadCommit == "" || e.HeadTree == "" || e.MergeBase == "" || e.CheckID == "" || e.IssueID == "" || e.RuleVersion == "" || !shadowTime(e.EventAt) {
		return errors.New("invalid provider event")
	}
	if e.RecordedAt != nil && (e.Scope != domain.ShadowScopeCI || !shadowTime(*e.RecordedAt)) {
		return errors.New("invalid event stamp/scope")
	}
	return nil
}

func storeShadowEvent(ctx context.Context, tx *sql.Tx, e domain.ProviderPREvent, publish bool) error {
	if err := validateShadowEvent(e); err != nil {
		return err
	}
	keys := []string{"scope", "repository_id", "event_id"}
	values := []any{e.Scope, e.RepositoryID, e.EventID}
	old, err := shadowPayload(ctx, tx, "shadow_events", keys, values)
	if err == nil {
		var prior domain.ProviderPREvent
		if err := json.Unmarshal(old, &prior); err != nil {
			return err
		}
		stamp := prior.RecordedAt
		prior.RecordedAt = nil
		bare := e
		bare.RecordedAt = nil
		if prior != bare {
			return ErrImmutable
		}
		if e.RecordedAt == nil {
			return nil
		}
		if stamp != nil {
			if stamp.Equal(*e.RecordedAt) {
				return nil
			}
			return ErrImmutable
		}
		if !publish {
			return errors.New("only publication may stamp events")
		}
		return updateShadowStamp(ctx, tx, "shadow_events", keys, values, e)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	return insertShadowJSON(ctx, tx, "shadow_events", keys, values, e)
}
func storeShadowLink(ctx context.Context, tx *sql.Tx, link domain.PRRunEventLink) error {
	if err := validateAttempt(link.AttemptID); err != nil {
		return err
	}
	if !validShadowScope(link.Scope) || link.EventID == "" {
		return errors.New("invalid event link")
	}
	var f domain.PRRunFact
	var e domain.ProviderPREvent
	b, err := shadowPayload(ctx, tx, "shadow_events", []string{"scope", "repository_id", "event_id"}, []any{link.Scope, link.AttemptID.RepositoryID, link.EventID})
	if err != nil {
		return fmt.Errorf("link event: %w", err)
	}
	if err := json.Unmarshal(b, &e); err != nil {
		return err
	}
	b, err = shadowPayload(ctx, tx, "pr_run_facts", []string{"scope", "repository_id", "attempt_id", "check_id"}, []any{link.Scope, link.AttemptID.RepositoryID, shadowAttemptKey(link.AttemptID), e.CheckID})
	if err != nil {
		return fmt.Errorf("link attempt: %w", err)
	}
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	if f.PRNumber != e.PRNumber || f.HeadTree != e.HeadTree || f.MergeBase != e.MergeBase || f.CheckID != e.CheckID || f.RuleVersion != e.RuleVersion {
		return errors.New("event link differs from attempt cohort/observation")
	}
	return insertShadowJSON(ctx, tx, "pr_run_events", []string{"scope", "repository_id", "attempt_id", "check_id", "event_id"}, []any{link.Scope, link.AttemptID.RepositoryID, shadowAttemptKey(link.AttemptID), e.CheckID, link.EventID}, link)
}

// Table/column names below are exclusively package-owned constants, never input.
func shadowWhere(keys []string) string { return strings.Join(keys, "=? AND ") + "=?" }
func shadowPayload(ctx context.Context, tx *sql.Tx, table string, keys []string, values []any) ([]byte, error) {
	var b []byte
	err := tx.QueryRowContext(ctx, "SELECT payload FROM "+table+" WHERE "+shadowWhere(keys), values...).Scan(&b)
	return b, err
}
func equalShadowJSON(old []byte, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	if !bytes.Equal(old, b) {
		return ErrImmutable
	}
	return nil
}
func insertShadowJSON(ctx context.Context, tx *sql.Tx, table string, keys []string, values []any, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	args := append(values, b)
	result, err := tx.ExecContext(ctx, "INSERT INTO "+table+" ("+strings.Join(keys, ",")+",payload) VALUES ("+strings.Repeat("?,", len(keys))+"?) ON CONFLICT DO NOTHING", args...)
	if err != nil {
		return fmt.Errorf("insert %s: %w", table, err)
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		old, err := shadowPayload(ctx, tx, table, keys, values)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrImmutable
		}
		if err != nil {
			return err
		}
		if !bytes.Equal(old, b) {
			return ErrImmutable
		}
	}
	return nil
}
func updateShadowStamp(ctx context.Context, tx *sql.Tx, table string, keys []string, values []any, value any) error {
	b, err := json.Marshal(value)
	if err != nil {
		return err
	}
	args := make([]any, 0, len(values)+1)
	args = append(args, b)
	args = append(args, values...)
	_, err = tx.ExecContext(ctx, "UPDATE "+table+" SET payload=? WHERE "+shadowWhere(keys), args...)
	return err
}
func readShadowRows[T any](ctx context.Context, tx *sql.Tx, table, order string, out *[]T) error {
	rows, err := tx.QueryContext(ctx, "SELECT payload FROM "+table+" ORDER BY "+order)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var b []byte
		var value T
		if err := rows.Scan(&b); err != nil {
			return err
		}
		if err := json.Unmarshal(b, &value); err != nil {
			return fmt.Errorf("decode %s: %w", table, err)
		}
		*out = append(*out, value)
	}
	return rows.Err()
}
