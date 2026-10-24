package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// ExportMappingRefresh is the exported mapping-refresh record (design §4).
const ExportMappingRefresh = ExportDir + "/mapping-refresh.json"

const metadataMappingRefresh = "mapping_refresh"

// MappingRefresh records when an agent map last completed (absent = onboarding
// has not happened).
type MappingRefresh struct {
	LastAgentRefreshAt *time.Time `json:"last_agent_refresh_at,omitempty"`
}

// LoadMappingRefresh returns the recorded refresh state (zero value if none).
func (s *Store) LoadMappingRefresh(ctx context.Context) (MappingRefresh, error) {
	raw, err := s.metadataValue(ctx, metadataMappingRefresh)
	if err != nil || raw == "" {
		return MappingRefresh{}, err
	}
	var record MappingRefresh
	if err := json.Unmarshal([]byte(raw), &record); err != nil {
		return MappingRefresh{}, fmt.Errorf("decode mapping refresh state: %w", err)
	}
	return record, nil
}

// SaveMappingRefresh durably records the refresh state.
func (s *Store) SaveMappingRefresh(ctx context.Context, record MappingRefresh) error {
	line, err := canonicalRefresh(record)
	if err != nil {
		return err
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, `INSERT INTO metadata(key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, metadataMappingRefresh, string(line))
		if err != nil {
			return fmt.Errorf("record mapping refresh state: %w", err)
		}
		return nil
	})
}

// ExportMappingRefreshFile renders the state/v1 mapping-refresh.json bytes, or
// nil when nothing has been recorded.
func (s *Store) ExportMappingRefreshFile(ctx context.Context) ([]byte, error) {
	record, err := s.LoadMappingRefresh(ctx)
	if err != nil {
		return nil, err
	}
	if record.LastAgentRefreshAt == nil {
		return nil, nil
	}
	return canonicalRefresh(record)
}

// importMappingRefresh restores the record inside an import transaction.
func importMappingRefresh(ctx context.Context, tx *sql.Tx, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	var record MappingRefresh
	if err := unmarshalNumbers(data, &record); err != nil {
		return fmt.Errorf("decode %s: %w", ExportMappingRefresh, err)
	}
	line, err := canonicalRefresh(record)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO metadata(key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`, metadataMappingRefresh, string(line)); err != nil {
		return fmt.Errorf("record imported mapping refresh state: %w", err)
	}
	return nil
}

// normalizeRefresh pins UTC times so export bytes are stable.
func normalizeRefresh(record MappingRefresh) MappingRefresh {
	if record.LastAgentRefreshAt != nil {
		t := record.LastAgentRefreshAt.UTC()
		record.LastAgentRefreshAt = &t
	}
	return record
}

// canonicalRefresh renders the record with sorted keys like every export line.
func canonicalRefresh(record MappingRefresh) ([]byte, error) {
	raw, err := json.Marshal(normalizeRefresh(record))
	if err != nil {
		return nil, err
	}
	var generic map[string]any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return nil, err
	}
	return CanonicalLine(generic)
}
