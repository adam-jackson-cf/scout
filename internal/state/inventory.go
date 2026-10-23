package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"scout/internal/domain"
)

// LatestRun returns the most recently finalized complete assessment. A partial
// or interrupted run remains durable but can never become an incremental baseline.
func (s *Store) LatestRun(ctx context.Context) (domain.Run, error) {
	var result domain.Run
	err := s.withRead(ctx, func(db *sql.DB) error {
		var payload []byte
		err := db.QueryRowContext(ctx, `SELECT payload FROM runs WHERE finalized = 1 AND state = ? ORDER BY started_at DESC, run_id DESC LIMIT 1`, domain.RunComplete).Scan(&payload)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("load latest run: %w", err)
		}
		if err := decode(payload, &result); err != nil {
			return err
		}
		if result.State != domain.RunComplete {
			return errors.New("finalized baseline payload is not complete")
		}
		return result.Validate()
	})
	return result, err
}

// LatestRecordedRun is for reporting, not incremental baselines: it includes
// finalized partial assessments and never exposes an interrupted run as final.
func (s *Store) LatestRecordedRun(ctx context.Context) (domain.Run, error) {
	var result domain.Run
	err := s.withRead(ctx, func(db *sql.DB) error {
		var payload []byte
		err := db.QueryRowContext(ctx, `SELECT payload FROM runs WHERE finalized = 1 ORDER BY started_at DESC, run_id DESC LIMIT 1`).Scan(&payload)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("load latest recorded run: %w", err)
		}
		if err := decode(payload, &result); err != nil {
			return err
		}
		return result.Validate()
	})
	return result, err
}

// LatestFinalizedRun is the explicit baseline API. LatestRun is retained for
// existing callers and has identical finalized-only semantics.
func (s *Store) LatestFinalizedRun(ctx context.Context) (domain.Run, error) {
	return s.LatestRun(ctx)
}

// LoadInventory returns the immutable file index attached to one finalized run.
func (s *Store) LoadInventory(ctx context.Context, runID string) ([]domain.File, error) {
	if runID == "" {
		return nil, errors.New("run ID is required")
	}
	var files []domain.File
	err := s.withRead(ctx, func(db *sql.DB) error {
		var payload []byte
		err := db.QueryRowContext(ctx, `SELECT i.payload FROM inventory_snapshots i JOIN runs r ON r.run_id = i.run_id WHERE i.run_id = ? AND r.finalized = 1 AND r.state = ?`, runID, domain.RunComplete).Scan(&payload)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("load inventory: %w", err)
		}
		if err := decode(payload, &files); err != nil {
			return err
		}
		for _, file := range files {
			if err := file.Validate(); err != nil {
				return fmt.Errorf("stored inventory file: %w", err)
			}
		}
		return nil
	})
	return files, err
}

// LoadLatestInventory returns the file index atomically stored with LatestRun.
// A legacy run without an index is not a valid incremental baseline.
func (s *Store) LoadLatestInventory(ctx context.Context) ([]domain.File, error) {
	run, err := s.LatestFinalizedRun(ctx)
	if err != nil {
		return nil, err
	}
	return s.LoadInventory(ctx, run.RunID)
}

// LoadLatestRecordedInventory returns the newest finalized run's inventory
// snapshot whatever the run's state. The inventory is a full scan even when
// analysis lanes were partial, so it is valid for rename/delete diffs but
// never for coverage reuse (use LoadLatestInventory for that).
func (s *Store) LoadLatestRecordedInventory(ctx context.Context) (domain.Run, []domain.File, error) {
	var run domain.Run
	var files []domain.File
	err := s.withRead(ctx, func(db *sql.DB) error {
		var runPayload, filePayload []byte
		err := db.QueryRowContext(ctx, `SELECT r.payload, i.payload FROM runs r JOIN inventory_snapshots i ON i.run_id = r.run_id WHERE r.finalized = 1 ORDER BY r.finished_at DESC, r.started_at DESC LIMIT 1`).Scan(&runPayload, &filePayload)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("load recorded inventory: %w", err)
		}
		if err := decode(runPayload, &run); err != nil {
			return err
		}
		return decode(filePayload, &files)
	})
	return run, files, err
}
