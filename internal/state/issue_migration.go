package state

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"scout/internal/domain"
)

type migratedIssue struct {
	issue   domain.Issue
	legacy  string
	reasons []string
}

// migrateV5 re-anchors the v1 issue store (design §3.3): every issue gets a
// durable opaque ID and an anchor derived from its linked observations, keeps
// its old ID as legacy_id, loses approval fields and states, and ambiguous rows
// are listed rather than merged.
func migrateV5(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `CREATE TABLE issue_migration_ambiguities (issue_id TEXT PRIMARY KEY, legacy_id TEXT NOT NULL, reason TEXT NOT NULL)`); err != nil {
		return fmt.Errorf("create migration ambiguity table: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `PRAGMA defer_foreign_keys = ON`); err != nil {
		return fmt.Errorf("defer foreign keys: %w", err)
	}
	rows, err := tx.QueryContext(ctx, `SELECT issue_id, payload FROM issues ORDER BY issue_id`)
	if err != nil {
		return fmt.Errorf("query legacy issues: %w", err)
	}
	type legacyRow struct {
		id      string
		payload []byte
	}
	var legacy []legacyRow
	for rows.Next() {
		var row legacyRow
		if err := rows.Scan(&row.id, &row.payload); err != nil {
			rows.Close()
			return fmt.Errorf("scan legacy issue: %w", err)
		}
		legacy = append(legacy, row)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	migrated := make([]*migratedIssue, 0, len(legacy))
	now := time.Now().UTC()
	for _, row := range legacy {
		item, err := reanchorLegacyIssue(ctx, tx, row.id, row.payload, now)
		if err != nil {
			return fmt.Errorf("re-anchor issue %q: %w", row.id, err)
		}
		migrated = append(migrated, item)
	}
	flagIndistinguishable(migrated)
	for _, item := range migrated {
		if len(item.reasons) != 0 {
			item.issue.History = append(item.issue.History, domain.IssueHistoryEntry{At: now, From: item.issue.State, To: item.issue.State, Event: domain.EventAmbiguous, Actor: "migration-v5", Reason: strings.Join(item.reasons, "; ")})
		}
		if err := writeMigratedIssue(ctx, tx, item); err != nil {
			return fmt.Errorf("write migrated issue %q: %w", item.legacy, err)
		}
	}
	return nil
}

func reanchorLegacyIssue(ctx context.Context, tx *sql.Tx, legacyID string, payload []byte, now time.Time) (*migratedIssue, error) {
	var raw map[string]any
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("decode legacy issue: %w", err)
	}
	if history, ok := raw["history"].([]any); ok {
		for index, entry := range history {
			fields, ok := entry.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("legacy issue %q: history entry %d is not an object", legacyID, index)
			}
			for _, key := range []string{"from", "to"} {
				if state, _ := fields[key].(string); state == "pending_verification" || state == "unverified" {
					fields[key] = string(domain.IssueOpen)
				}
			}
		}
	}
	delete(raw, "proposedAction")
	delete(raw, "approvalRef")
	if state, _ := raw["state"].(string); state == "pending_verification" || state == "unverified" {
		raw["state"] = string(domain.IssueOpen)
	}
	if verification, ok := raw["verification"].(map[string]any); ok {
		delete(verification, "authorizedChangeRef")
	}
	cleaned, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	var issue domain.Issue
	if err := json.Unmarshal(cleaned, &issue); err != nil {
		return nil, fmt.Errorf("decode legacy issue: %w", err)
	}
	idRows, err := tx.QueryContext(ctx, `SELECT observation_id FROM issue_observations WHERE issue_id = ?`, legacyID)
	if err != nil {
		return nil, fmt.Errorf("query legacy observations: %w", err)
	}
	var observationIDs []string
	for idRows.Next() {
		var id string
		if err := idRows.Scan(&id); err != nil {
			idRows.Close()
			return nil, err
		}
		observationIDs = append(observationIDs, id)
	}
	idRows.Close()
	observations, err := loadObservationsTx(ctx, tx, observationIDs)
	if err != nil {
		return nil, err
	}
	item := &migratedIssue{legacy: legacyID}
	family := domain.Family(strings.Split(legacyID, ":")[min(1, strings.Count(legacyID, ":"))])
	if len(observations) != 0 {
		latest := observations[0]
		for _, observation := range observations {
			if observation.ObservedAt.After(latest.ObservedAt) || observation.ObservedAt.Equal(latest.ObservedAt) && observation.ObservationID > latest.ObservationID {
				latest = observation
			}
		}
		family = latest.Family
		issue.Family, issue.Category = latest.Family, latest.OriginalCategory
		anchors := map[domain.IssueAnchor]struct{}{}
		var primary domain.IssueAnchor
		var primaryID string
		for _, observation := range observations {
			if observation.RunID != latest.RunID {
				continue
			}
			anchor, ok := domain.AnchorOf(observation)
			if !ok {
				continue
			}
			anchors[anchor] = struct{}{}
			if primaryID == "" || observation.ObservationID < primaryID {
				primary, primaryID = anchor, observation.ObservationID
			}
		}
		issue.Anchor = primary
		if primaryID == "" {
			item.reasons = append(item.reasons, "no linked observation carries an evidence location to anchor")
		}
		if len(anchors) > 1 {
			item.reasons = append(item.reasons, fmt.Sprintf("legacy issue merged %d distinct occurrences in its latest run", len(anchors)))
		}
	} else {
		item.reasons = append(item.reasons, "no linked observations to anchor")
	}
	if issue.LegacyID == "" {
		issue.LegacyID = legacyID
	}
	if family.Valid() {
		issue.IssueID = domain.NewIssueID(family, now)
	} else {
		// No family can be derived, so no durable ID is minted: the row keeps
		// its legacy ID and is listed as ambiguous.
		issue.IssueID = legacyID
		item.reasons = append(item.reasons, fmt.Sprintf("unknown family in legacy ID %q", legacyID))
	}
	issue.History = append(issue.History, domain.IssueHistoryEntry{At: now, From: issue.State, To: issue.State, Event: domain.EventMigrated, Actor: "migration-v5", Reason: "re-anchored from legacy issue " + legacyID})
	item.issue = issue
	return item, nil
}

// flagIndistinguishable marks issues that share the full anchor tuple: a later
// run could not tell which occurrence is which, so they are listed and linked
// through AmbiguousWith instead of being matched or merged.
func flagIndistinguishable(items []*migratedIssue) {
	groups := map[string][]*migratedIssue{}
	for _, item := range items {
		if item.issue.Anchor.RuleKey == "" {
			continue
		}
		key := fmt.Sprintf("%s\x1f%s\x1f%v", item.issue.Family, item.issue.Category, item.issue.Anchor)
		if item.issue.Anchor.Secret {
			key = fmt.Sprintf("%s\x1f%s\x1f%s\x1f%s", item.issue.Family, item.issue.Category, item.issue.Anchor.RuleKey, item.issue.Anchor.Fingerprint)
		}
		groups[key] = append(groups[key], item)
	}
	for _, group := range groups {
		if len(group) < 2 {
			continue
		}
		for _, item := range group {
			peers := make([]string, 0, len(group)-1)
			for _, other := range group {
				if other != item {
					peers = append(peers, other.issue.IssueID)
				}
			}
			sort.Strings(peers)
			item.issue.AmbiguousWith = peers
			item.reasons = append(item.reasons, fmt.Sprintf("indistinguishable from %d other issue(s) with the same rule, path, symbol and fingerprint", len(peers)))
		}
	}
}

func writeMigratedIssue(ctx context.Context, tx *sql.Tx, item *migratedIssue) error {
	issue := item.issue
	if err := issue.Validate(); err != nil {
		return fmt.Errorf("migrated issue is invalid: %w", err)
	}
	payload, err := encode(issue)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE issues SET issue_id = ?, state = ?, payload = ? WHERE issue_id = ?`, issue.IssueID, issue.State, payload, item.legacy); err != nil {
		return fmt.Errorf("update issue: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE issue_observations SET issue_id = ? WHERE issue_id = ?`, issue.IssueID, item.legacy); err != nil {
		return fmt.Errorf("update issue observations: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM issue_history WHERE issue_id IN (?, ?)`, item.legacy, issue.IssueID); err != nil {
		return fmt.Errorf("replace issue history: %w", err)
	}
	for ordinal, entry := range issue.History {
		entryPayload, err := encode(entry)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO issue_history(issue_id, ordinal, payload) VALUES (?, ?, ?)`, issue.IssueID, ordinal, entryPayload); err != nil {
			return fmt.Errorf("save issue history: %w", err)
		}
	}
	if len(item.reasons) != 0 {
		if _, err := tx.ExecContext(ctx, `INSERT INTO issue_migration_ambiguities(issue_id, legacy_id, reason) VALUES (?, ?, ?)`, issue.IssueID, item.legacy, strings.Join(item.reasons, "; ")); err != nil {
			return fmt.Errorf("record migration ambiguity: %w", err)
		}
	}
	return nil
}
