package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"scout/internal/domain"
)

// schemaAssociations is the merge-time association table (design §4, §7.2): one
// payload row per decided cluster per run and per inherited file. The refresh
// generation is the LastAgentRefreshAt in force when decided ("" before
// onboarding); the refresh guard reads rows of one generation in decision order.
var schemaAssociations = []string{
	`CREATE TABLE IF NOT EXISTS associations (id TEXT PRIMARY KEY, run_id TEXT NOT NULL, decided_at INTEGER NOT NULL, refresh_generation TEXT NOT NULL, payload BLOB NOT NULL)`,
	`CREATE INDEX IF NOT EXISTS associations_by_decision ON associations(decided_at, id)`,
	`CREATE INDEX IF NOT EXISTS associations_by_generation ON associations(refresh_generation, decided_at, id)`,
	`CREATE INDEX IF NOT EXISTS associations_by_run ON associations(run_id)`,
}

// ExportAssociations is the repository-relative path of the exported association records.
const ExportAssociations = ExportDir + "/associations.jsonl"

// InsertAssociations appends records; an existing ID is an error (records are append-only).
func (s *Store) InsertAssociations(ctx context.Context, list []domain.Association) error {
	if s.readOnly {
		return errors.New("state store is read-only")
	}
	normalized, err := normalizeAssociations(list)
	if err != nil {
		return err
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error { return insertAssociations(ctx, tx, normalized) })
}

// ListAssociations returns every record ordered by decision time, then ID.
func (s *Store) ListAssociations(ctx context.Context) ([]domain.Association, error) {
	var result []domain.Association
	err := s.withRead(ctx, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT payload FROM associations ORDER BY decided_at, id`)
		if err != nil {
			return fmt.Errorf("list associations: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var payload []byte
			if err := rows.Scan(&payload); err != nil {
				return fmt.Errorf("scan association: %w", err)
			}
			var record domain.Association
			if err := decode(payload, &record); err != nil {
				return err
			}
			result = append(result, record)
		}
		return rows.Err()
	})
	return result, err
}

// MarkAssociationsStale sets Stale on the named records; an unknown ID is an error.
func (s *Store) MarkAssociationsStale(ctx context.Context, ids []string) error {
	return s.updateAssociations(ctx, ids, func(a *domain.Association) { a.Stale = true })
}

// MarkAssociationsSuperseded sets Superseded on the named records; an unknown ID is an error.
func (s *Store) MarkAssociationsSuperseded(ctx context.Context, ids []string) error {
	return s.updateAssociations(ctx, ids, func(a *domain.Association) { a.Superseded = true })
}

func (s *Store) updateAssociations(ctx context.Context, ids []string, change func(*domain.Association)) error {
	if s.readOnly {
		return errors.New("state store is read-only")
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		for _, id := range ids {
			var payload []byte
			if err := tx.QueryRowContext(ctx, `SELECT payload FROM associations WHERE id = ?`, id).Scan(&payload); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return fmt.Errorf("association %q not found", id)
				}
				return fmt.Errorf("read association %q: %w", id, err)
			}
			var record domain.Association
			if err := decode(payload, &record); err != nil {
				return err
			}
			change(&record)
			if payload, err := encode(record); err != nil {
				return err
			} else if _, err := tx.ExecContext(ctx, `UPDATE associations SET payload = ? WHERE id = ?`, payload, id); err != nil {
				return fmt.Errorf("update association %q: %w", id, err)
			}
		}
		return nil
	})
}

// normalizeAssociations validates records and fixes their encoding: UTC times
// and sorted lists, so an export round trip is byte-stable.
func normalizeAssociations(in []domain.Association) ([]domain.Association, error) {
	out := make([]domain.Association, len(in))
	seen := make(map[string]bool, len(in))
	for i, record := range in {
		if err := record.Validate(); err != nil {
			return nil, err
		}
		if seen[record.ID] {
			return nil, fmt.Errorf("duplicate association %q", record.ID)
		}
		seen[record.ID] = true
		record.DecidedAt = record.DecidedAt.UTC()
		if record.LastAgentRefreshAt != nil {
			refreshed := record.LastAgentRefreshAt.UTC()
			record.LastAgentRefreshAt = &refreshed
		}
		record.Files = sortedCopy(record.Files)
		record.AreaIDs = sortedCopy(record.AreaIDs)
		record.InheritedAreaIDs = sortedCopy(record.InheritedAreaIDs)
		out[i] = record
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].DecidedAt.Equal(out[j].DecidedAt) {
			return out[i].DecidedAt.Before(out[j].DecidedAt)
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func sortedCopy(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := append([]string{}, in...)
	sort.Strings(out)
	return out
}

func insertAssociations(ctx context.Context, tx *sql.Tx, list []domain.Association) error {
	for _, record := range list {
		payload, err := encode(record)
		if err != nil {
			return err
		}
		generation := ""
		if record.LastAgentRefreshAt != nil {
			generation = record.LastAgentRefreshAt.Format(time.RFC3339Nano)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO associations(id, run_id, decided_at, refresh_generation, payload) VALUES (?, ?, ?, ?, ?)`,
			record.ID, record.RunID, unixMillis(record.DecidedAt), generation, payload); err != nil {
			return fmt.Errorf("record association %q: %w", record.ID, err)
		}
	}
	return nil
}

// exportAssociations renders associations.jsonl, one canonical record per line
// in decision order; nil when the store holds none.
func (s *Store) exportAssociations(ctx context.Context) ([]byte, error) {
	list, err := s.ListAssociations(ctx)
	if err != nil || len(list) == 0 {
		return nil, err
	}
	lines := make([][]byte, 0, len(list))
	for _, record := range list {
		line, err := canonicalRecord(record)
		if err != nil {
			return nil, fmt.Errorf("encode association: %w", err)
		}
		lines = append(lines, line)
	}
	return joinLines(lines), nil
}

// decodeAssociations reads associations.jsonl of an export, validated, in
// decision order. An export without the file (every one before revision 2.8)
// yields no records.
func decodeAssociations(files map[string][]byte) ([]domain.Association, error) {
	var list []domain.Association
	for number, line := range SplitLines(files[ExportAssociations]) {
		var record domain.Association
		if err := decode(line, &record); err != nil {
			return nil, fmt.Errorf("associations.jsonl line %d: %w", number+1, err)
		}
		list = append(list, record)
	}
	normalized, err := normalizeAssociations(list)
	if err != nil {
		return nil, fmt.Errorf("associations.jsonl: %w", err)
	}
	return normalized, nil
}

// canonicalRecord renders v as one canonical, unescaped JSON line.
func canonicalRecord(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	fields := map[string]any{}
	if err := unmarshalNumbers(raw, &fields); err != nil {
		return nil, err
	}
	return CanonicalLine(fields)
}
