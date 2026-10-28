package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"scout/internal/domain"
)

// schemaSignals adds the append-only signals ledger (design §3.2). The primary key
// is the signal identity plus tool and config digests, so recomputing a
// snapshot with the same tool is idempotent and a different tool appends.
var schemaSignals = []string{
	`CREATE TABLE IF NOT EXISTS signals (seq INTEGER PRIMARY KEY AUTOINCREMENT, metric TEXT NOT NULL, scope_kind TEXT NOT NULL, scope_anchor TEXT NOT NULL, snapshot TEXT NOT NULL, window TEXT NOT NULL, tool_digest TEXT NOT NULL, config_digest TEXT NOT NULL, payload BLOB NOT NULL, UNIQUE(metric, scope_kind, scope_anchor, snapshot, window, tool_digest, config_digest))`,
	`CREATE INDEX IF NOT EXISTS signals_by_scope ON signals(scope_anchor, metric, seq)`,
}

// SignalFilter selects signals; empty fields match everything.
type SignalFilter struct {
	MetricID    string // "scout.cc" or "scout.cc@1"
	ScopeKind   string
	ScopeAnchor string
	Snapshot    string
	Window      string
	// Latest keeps, per (metric, scope, window), only the most recently
	// recorded snapshot.
	Latest bool
}

// RecordSignals appends signals. A signal whose identity, tool and config
// digests are already recorded is left untouched (append-only).
func (s *Store) RecordSignals(ctx context.Context, signals []domain.Signal) error {
	if s.readOnly {
		return errors.New("state store is read-only")
	}
	for _, signal := range signals {
		if err := signal.Validate(); err != nil {
			return err
		}
		for _, value := range []string{signal.ScopeAnchor, signal.MissingReason} {
			if err := rejectUnsafePersistence(value); err != nil {
				return err
			}
		}
	}
	return s.withTransaction(ctx, func(tx *sql.Tx) error {
		return insertSignals(ctx, tx, signals)
	})
}

func insertSignals(ctx context.Context, tx *sql.Tx, signals []domain.Signal) error {
	for _, signal := range signals {
		payload, err := json.Marshal(signal)
		if err != nil {
			return fmt.Errorf("encode signal: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO signals(metric, scope_kind, scope_anchor, snapshot, window, tool_digest, config_digest, payload)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT DO NOTHING`,
			signal.ID(), signal.ScopeKind, signal.ScopeAnchor, signal.Snapshot, signal.Window, signal.ToolDigest, signal.ConfigDigest, payload); err != nil {
			return fmt.Errorf("record signal %s: %w", signal.ID(), err)
		}
	}
	return nil
}

// ListSignals returns signals in deterministic order: metric, scope kind,
// scope anchor, snapshot, window, tool digest, config digest.
func (s *Store) ListSignals(ctx context.Context, filter SignalFilter) ([]domain.Signal, error) {
	type row struct {
		seq    int64
		signal domain.Signal
	}
	var rows []row
	err := s.withRead(ctx, func(db *sql.DB) error {
		query := `SELECT seq, payload FROM signals WHERE 1=1`
		var args []any
		add := func(column, value string) {
			if value != "" {
				query += ` AND ` + column + ` = ?`
				args = append(args, value)
			}
		}
		metric := filter.MetricID
		if metric != "" && !containsVersion(metric) {
			query += ` AND metric LIKE ? ESCAPE '\'`
			args = append(args, escapeLike(metric)+"@%")
		} else {
			add("metric", metric)
		}
		add("scope_kind", filter.ScopeKind)
		add("scope_anchor", filter.ScopeAnchor)
		add("snapshot", filter.Snapshot)
		add("window", filter.Window)
		result, err := db.QueryContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("query signals: %w", err)
		}
		defer result.Close()
		for result.Next() {
			var seq int64
			var payload []byte
			if err := result.Scan(&seq, &payload); err != nil {
				return fmt.Errorf("scan signal: %w", err)
			}
			var signal domain.Signal
			if err := json.Unmarshal(payload, &signal); err != nil {
				return fmt.Errorf("decode signal: %w", err)
			}
			rows = append(rows, row{seq, signal})
		}
		return result.Err()
	})
	if err != nil {
		return nil, err
	}
	if filter.Latest {
		newest := map[string]int64{}
		key := func(signal domain.Signal) string {
			return signal.ID() + "\x00" + signal.ScopeKind + "\x00" + signal.ScopeAnchor + "\x00" + signal.Window
		}
		for _, r := range rows {
			if r.seq > newest[key(r.signal)] {
				newest[key(r.signal)] = r.seq
			}
		}
		kept := rows[:0]
		for _, r := range rows {
			if r.seq == newest[key(r.signal)] {
				kept = append(kept, r)
			}
		}
		rows = kept
	}
	signals := make([]domain.Signal, len(rows))
	for i, r := range rows {
		signals[i] = r.signal
	}
	SortSignals(signals)
	return signals, nil
}

// SortSignals orders signals by identity (the export order).
func SortSignals(signals []domain.Signal) {
	sort.SliceStable(signals, func(i, j int) bool { return signals[i].Key() < signals[j].Key() })
}

func containsVersion(metric string) bool {
	for _, r := range metric {
		if r == '@' {
			return true
		}
	}
	return false
}

func escapeLike(value string) string {
	out := make([]rune, 0, len(value))
	for _, r := range value {
		if r == '%' || r == '_' || r == '\\' {
			out = append(out, '\\')
		}
		out = append(out, r)
	}
	return string(out)
}
