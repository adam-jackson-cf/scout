package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"scout/internal/areas"
)

// schemaV12 adds the area labels table (design §4, §7.2). Areas are labels over
// Clawpatch output: nothing here feeds issue identity or signal scope keys.
var schemaV12 = []string{
	`CREATE TABLE IF NOT EXISTS areas (area_id TEXT PRIMARY KEY, payload BLOB NOT NULL)`,
}

// ExportAreas is the repository-relative path of the exported area labels.
const ExportAreas = ExportDir + "/areas.jsonl"

// ListAreas returns every area, retired ones included, sorted by area ID.
func (s *Store) ListAreas(ctx context.Context) ([]areas.Area, error) {
	var result []areas.Area
	err := s.withRead(ctx, func(db *sql.DB) error {
		rows, err := db.QueryContext(ctx, `SELECT payload FROM areas ORDER BY area_id`)
		if err != nil {
			return fmt.Errorf("list areas: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var payload []byte
			if err := rows.Scan(&payload); err != nil {
				return fmt.Errorf("scan area: %w", err)
			}
			var area areas.Area
			if err := decode(payload, &area); err != nil {
				return err
			}
			result = append(result, area)
		}
		return rows.Err()
	})
	return result, err
}

// ReplaceAreas atomically replaces the area set with next (Reconcile output).
func (s *Store) ReplaceAreas(ctx context.Context, next []areas.Area) error {
	if s.readOnly {
		return errors.New("state store is read-only")
	}
	normalized, err := normalizeAreas(next)
	if err != nil {
		return err
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		if _, err := tx.ExecContext(ctx, `DELETE FROM areas`); err != nil {
			return fmt.Errorf("clear areas: %w", err)
		}
		return insertAreas(ctx, tx, normalized)
	})
}

func normalizeAreas(in []areas.Area) ([]areas.Area, error) {
	out := make([]areas.Area, len(in))
	seen := make(map[string]bool, len(in))
	for i, area := range in {
		if area.ScoutAreaID == "" || area.Origin == "" {
			return nil, errors.New("area requires an ID and an origin")
		}
		if seen[area.ScoutAreaID] {
			return nil, fmt.Errorf("duplicate area %q", area.ScoutAreaID)
		}
		seen[area.ScoutAreaID] = true
		area.FileSet = append([]string{}, area.FileSet...)
		area.ClawpatchIDs = append([]string{}, area.ClawpatchIDs...)
		sort.Strings(area.FileSet)
		sort.Strings(area.ClawpatchIDs)
		out[i] = area
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ScoutAreaID < out[j].ScoutAreaID })
	return out, nil
}

func insertAreas(ctx context.Context, tx *sql.Tx, list []areas.Area) error {
	for _, area := range list {
		payload, err := encode(area)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO areas(area_id, payload) VALUES (?, ?)`, area.ScoutAreaID, payload); err != nil {
			return fmt.Errorf("record area %q: %w", area.ScoutAreaID, err)
		}
	}
	return nil
}

// exportAreas renders areas.jsonl, one canonical record per line sorted by
// area ID; nil when the store holds no areas.
func (s *Store) exportAreas(ctx context.Context) ([]byte, error) {
	list, err := s.ListAreas(ctx)
	if err != nil || len(list) == 0 {
		return nil, err
	}
	lines := make([][]byte, 0, len(list))
	for _, area := range list {
		raw, err := json.Marshal(area)
		if err != nil {
			return nil, fmt.Errorf("encode area: %w", err)
		}
		fields := map[string]any{}
		if err := unmarshalNumbers(raw, &fields); err != nil {
			return nil, err
		}
		line, err := canonicalLine(fields)
		if err != nil {
			return nil, err
		}
		lines = append(lines, line)
	}
	return joinLines(lines), nil
}

// decodeAreas reads areas.jsonl of an export, validated, in area-ID order.
func decodeAreas(files map[string][]byte) ([]areas.Area, error) {
	var list []areas.Area
	for number, line := range splitLines(files[ExportAreas]) {
		var area areas.Area
		if err := decode(line, &area); err != nil {
			return nil, fmt.Errorf("areas.jsonl line %d: %w", number+1, err)
		}
		list = append(list, area)
	}
	normalized, err := normalizeAreas(list)
	if err != nil {
		return nil, fmt.Errorf("areas.jsonl: %w", err)
	}
	return normalized, nil
}
