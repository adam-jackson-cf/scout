package baseline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"scout/internal/domain"
	"scout/internal/state"
)

// MergeState combines the tip's state/v1 files (current) with a local export so
// a publisher that lost a compare-and-swap race re-applies its ledger on top of
// the other publisher's instead of overwriting it. Issues are unioned by issue
// ID; when both sides carry an issue, the side whose disposition was updated
// last wins (a tie keeps the local side) and both of its rows are taken
// together. would_block rows are unioned. The coverage summary is local. The
// manifest is regenerated for the merged files.
func MergeState(current, local Tree) (Tree, error) {
	if len(current) == 0 {
		return local, nil
	}
	merged := Tree{}
	for path, data := range current {
		merged[path] = data
	}
	for path, data := range local {
		merged[path] = data
	}

	localDisp, err := indexRows(local[state.ExportDispositions])
	if err != nil {
		return nil, fmt.Errorf("local %w", err)
	}
	currentDisp, err := indexRows(current[state.ExportDispositions])
	if err != nil {
		return nil, fmt.Errorf("tip %w", err)
	}
	localIssues, err := indexRows(local[state.ExportIssues])
	if err != nil {
		return nil, fmt.Errorf("local %w", err)
	}
	currentIssues, err := indexRows(current[state.ExportIssues])
	if err != nil {
		return nil, fmt.Errorf("tip %w", err)
	}

	ids := map[string]bool{}
	for id := range localIssues {
		ids[id] = true
	}
	for id := range currentIssues {
		ids[id] = true
	}
	sorted := make([]string, 0, len(ids))
	for id := range ids {
		sorted = append(sorted, id)
	}
	sort.Strings(sorted)

	var issueOut, dispOut bytes.Buffer
	for _, id := range sorted {
		issueRow, disp := localIssues[id], localDisp[id]
		currentIssue, currentDispRow := currentIssues[id], currentDisp[id]
		switch {
		case currentIssue == nil || currentDispRow == nil:
			// local only
		case issueRow == nil || disp == nil:
			issueRow, disp = currentIssue, currentDispRow
		case updated(currentDispRow).After(updated(disp)):
			issueRow, disp = currentIssue, currentDispRow
		}
		if issueRow == nil || disp == nil {
			return nil, fmt.Errorf("issue %q lacks an issue or disposition row", id)
		}
		issueOut.Write(issueRow.line)
		dispOut.Write(disp.line)
	}
	merged[state.ExportIssues] = issueOut.Bytes()
	merged[state.ExportDispositions] = dispOut.Bytes()

	blocks := map[string]*row{}
	for _, source := range []Tree{current, local} {
		for number, line := range state.SplitLines(source[state.ExportWouldBlock]) {
			var fields struct {
				ID        string `json:"id"`
				TriagedAt string `json:"triagedAt"`
			}
			if err := json.Unmarshal(line, &fields); err != nil || fields.ID == "" {
				return nil, fmt.Errorf("would_block row %d is not a keyed record", number+1)
			}
			candidate := &row{line: append(append([]byte{}, line...), '\n'), updated: fields.TriagedAt}
			// The later maintainer triage wins; local wins ties (it is visited last).
			if existing := blocks[fields.ID]; existing == nil || !updated(existing).After(updated(candidate)) {
				blocks[fields.ID] = candidate
			}
		}
	}
	keys := make([]string, 0, len(blocks))
	for id := range blocks {
		keys = append(keys, id)
	}
	sort.Strings(keys)
	var blockOut bytes.Buffer
	for _, id := range keys {
		blockOut.Write(blocks[id].line)
	}
	merged[state.ExportWouldBlock] = blockOut.Bytes()

	for path := range merged {
		if !strings.HasPrefix(path, state.ExportSignalsDir) {
			continue
		}
		unioned, err := unionSignals(current[path], local[path])
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		merged[path] = unioned
	}

	meta, err := state.ReadManifestMeta(local[state.ExportManifest])
	if err != nil {
		return nil, err
	}
	if currentMeta, err := state.ReadManifestMeta(current[state.ExportManifest]); err == nil && currentMeta.WouldBlock == state.WouldBlockPresent {
		meta.WouldBlock = state.WouldBlockPresent
	}
	delete(merged, state.ExportManifest)
	merged[state.ExportManifest] = state.RenderManifest(meta, merged)
	return merged, nil
}

type row struct {
	line    []byte
	updated string
}

func updated(r *row) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, r.updated)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

// indexRows maps issueId to its JSONL row (kept with its trailing LF).
func indexRows(data []byte) (map[string]*row, error) {
	rows := map[string]*row{}
	for number, line := range state.SplitLines(data) {
		var fields struct {
			IssueID   string `json:"issueId"`
			UpdatedAt string `json:"updatedAt"`
		}
		if err := json.Unmarshal(line, &fields); err != nil || fields.IssueID == "" {
			return nil, fmt.Errorf("row %d is not a keyed record", number+1)
		}
		rows[fields.IssueID] = &row{line: append(append([]byte{}, line...), '\n'), updated: fields.UpdatedAt}
	}
	return rows, nil
}

// unionSignals merges two signal files. Signals are append-only, so the union
// keyed by signal identity (metric, scope, snapshot, window, tool and config
// digests) loses nothing; the local record wins an identity collision.
func unionSignals(current, local []byte) ([]byte, error) {
	rows := map[string][]byte{}
	for _, source := range [][]byte{current, local} {
		for number, line := range state.SplitLines(source) {
			var signal domain.Signal
			if err := json.Unmarshal(line, &signal); err != nil {
				return nil, fmt.Errorf("signal row %d: %w", number+1, err)
			}
			rows[signal.Key()] = append(append([]byte{}, line...), '\n')
		}
	}
	keys := make([]string, 0, len(rows))
	for key := range rows {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var out bytes.Buffer
	for _, key := range keys {
		out.Write(rows[key])
	}
	return out.Bytes(), nil
}
