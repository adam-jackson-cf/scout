package state

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"scout/internal/areas"
	"scout/internal/domain"
)

// AreaAction records a maintainer's authorized pending-area decision.
type AreaAction struct {
	Action       string    `json:"action"`
	AreaID       string    `json:"area_id"`
	TargetAreaID string    `json:"target_area_id,omitempty"`
	Actor        string    `json:"actor"`
	Reason       string    `json:"reason"`
	At           time.Time `json:"at"`
}

// ExportAreaActions is the audit trail's export path.
const ExportAreaActions = ExportDir + "/area_actions.jsonl"
const metadataAreaActions = "area_actions"

// ListAreaActions returns the append-ordered audit trail.
func (s *Store) ListAreaActions(ctx context.Context) ([]AreaAction, error) {
	value, err := s.metadataValue(ctx, metadataAreaActions)
	if err != nil {
		return nil, err
	}
	return decodeAreaActions([]byte(value))
}

func decodeAreaActions(data []byte) ([]AreaAction, error) {
	var actions []AreaAction
	for _, line := range SplitLines(data) {
		var a AreaAction
		if err := decode(line, &a); err != nil {
			return nil, err
		}
		if a.AreaID == "" || a.At.IsZero() || (a.Action != "accept" && a.Action != "merge") || (a.Action == "merge" && a.TargetAreaID == "") {
			return nil, errors.New("invalid area action")
		}
		if err := validateAreaActor(a.Actor, a.Reason); err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	return actions, nil
}

func validateAreaActor(actor, reason string) error {
	if strings.TrimSpace(actor) == "" || strings.TrimSpace(reason) == "" {
		return errors.New("area action requires actor and reason")
	}
	for _, value := range []string{actor, reason} {
		if err := rejectUnsafePersistence(value); err != nil {
			return err
		}
	}
	return nil
}

// AcceptArea activates a proposal and preserves its provenance.
func (s *Store) AcceptArea(ctx context.Context, id, actor, reason string) (areas.Area, error) {
	return s.actOnArea(ctx, id, "", actor, reason)
}

// MergeArea resolves a proposal by assigning its files to an active label.
func (s *Store) MergeArea(ctx context.Context, pendingID, targetID, actor, reason string) (areas.Area, error) {
	if targetID == "" {
		return areas.Area{}, errors.New("merge target required")
	}
	return s.actOnArea(ctx, pendingID, targetID, actor, reason)
}

func (s *Store) actOnArea(ctx context.Context, id, targetID, actor, reason string) (areas.Area, error) {
	if s.readOnly {
		return areas.Area{}, errors.New("state store is read-only")
	}
	if err := validateAreaActor(actor, reason); err != nil {
		return areas.Area{}, err
	}
	var updated areas.Area
	err := s.withTransaction(ctx, func(tx *sql.Tx) error {
		load := func(id string) (areas.Area, error) {
			var payload []byte
			var a areas.Area
			if err := tx.QueryRowContext(ctx, `SELECT payload FROM areas WHERE area_id = ?`, id).Scan(&payload); err != nil {
				return a, fmt.Errorf("area not found: %w", err)
			}
			err := decode(payload, &a)
			return a, err
		}
		source, err := load(id)
		if err != nil {
			return err
		}
		if source.Status != areas.StatusPending {
			return errors.New("source area is not pending")
		}
		now := time.Now().UTC()
		action := AreaAction{Action: "accept", AreaID: id, Actor: actor, Reason: reason, At: now}
		if targetID == "" {
			source.Status, source.Origin = areas.StatusActive, "association"
			source.AcceptedBy, source.AcceptedAt = actor, &now
		} else {
			target, err := load(targetID)
			if err != nil {
				return err
			}
			if !target.Active() {
				return errors.New("merge target is not active")
			}
			action.Action, action.TargetAreaID = "merge", targetID
			source.Status = areas.StatusSuperseded
			// The pending decision supplies the original run and mapped revision.
			var provenance domain.Association
			rows, err := tx.QueryContext(ctx, `SELECT payload FROM associations ORDER BY decided_at, id`)
			if err != nil {
				return err
			}
			for rows.Next() {
				var payload []byte
				if err := rows.Scan(&payload); err != nil {
					rows.Close()
					return err
				}
				var record domain.Association
				if err := decode(payload, &record); err != nil {
					rows.Close()
					return err
				}
				if record.PendingAreaID == id {
					provenance = record
				}
			}
			rowErr := rows.Err()
			rows.Close()
			if rowErr != nil {
				return rowErr
			}
			if provenance.ID == "" {
				// Restored area labels can predate association-row exports. The
				// maintainer action is its own run, anchored to the area snapshot.
				provenance.RunID = "maintainer-merge:" + id + ":" + now.Format(time.RFC3339Nano)
				provenance.MergeCommit = source.FirstSeenSnapshot
				provenance.PRMetadata = domain.PRMetadataAbsent
			}
			record := domain.Association{ID: "maintainer-merge:" + id + ":" + now.Format(time.RFC3339Nano), RunID: provenance.RunID, DecidedAt: now, LastAgentRefreshAt: provenance.LastAgentRefreshAt, MergeCommit: provenance.MergeCommit, ClusterKey: source.ClusterID, Files: append([]string(nil), source.FileSet...), Outcome: domain.AssociationAssigned, Reason: "maintainer_merge", AreaIDs: []string{targetID}, PRMetadata: provenance.PRMetadata}
			normalized, err := normalizeAssociations([]domain.Association{record})
			if err != nil {
				return err
			}
			if err := insertAssociations(ctx, tx, normalized); err != nil {
				return err
			}
		}
		payload, err := encode(source)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE areas SET payload = ? WHERE area_id = ?`, payload, id); err != nil {
			return err
		}
		var previous string
		err = tx.QueryRowContext(ctx, `SELECT value FROM metadata WHERE key = ?`, metadataAreaActions).Scan(&previous)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		line, err := canonicalRecord(action)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO metadata(key,value) VALUES (?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, metadataAreaActions, previous+string(joinLines([][]byte{line}))); err != nil {
			return err
		}
		updated = source
		return nil
	})
	return updated, err
}
