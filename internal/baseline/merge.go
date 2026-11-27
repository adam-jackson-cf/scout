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
	if err := validateMergeInputs(current, local); err != nil {
		return nil, err
	}
	if len(current) == 0 || len(local) == 0 {
		return copyNonEmptyTree(current, local), nil
	}
	merged := overlayTrees(current, local)
	issues, dispositions, err := mergeIssues(current, local)
	if err != nil {
		return nil, err
	}
	merged[state.ExportIssues] = issues
	merged[state.ExportDispositions] = dispositions
	wouldBlock, err := mergeWouldBlock(current, local)
	if err != nil {
		return nil, err
	}
	merged[state.ExportWouldBlock] = wouldBlock
	if err := mergeSignalFiles(current, local, merged); err != nil {
		return nil, err
	}
	if err := mergeShadow(current, local, merged); err != nil {
		return nil, err
	}
	if err := renderMergedManifest(current, local, merged); err != nil {
		return nil, err
	}
	return merged, nil
}

func validateMergeInputs(current, local Tree) error {
	for _, tree := range []Tree{current, local} {
		if len(tree) != 0 {
			if _, err := state.ValidateState(tree); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyNonEmptyTree(current, local Tree) Tree {
	source := current
	if len(local) != 0 {
		source = local
	}
	out := make(Tree, len(source))
	for path, data := range source {
		out[path] = data
	}
	return out
}

func overlayTrees(current, local Tree) Tree {
	merged := Tree{}
	for path, data := range current {
		merged[path] = data
	}
	for path, data := range local {
		merged[path] = data
	}
	return merged
}

func indexSideRows(side string, data []byte, requireTime bool) (map[string]*row, error) {
	rows, err := indexRows(data, requireTime)
	if err != nil {
		return nil, fmt.Errorf("%s %w", side, err)
	}
	return rows, nil
}

func mergeIssues(current, local Tree) ([]byte, []byte, error) {
	localDisp, err := indexSideRows("local", local[state.ExportDispositions], true)
	if err != nil {
		return nil, nil, err
	}
	currentDisp, err := indexSideRows("tip", current[state.ExportDispositions], true)
	if err != nil {
		return nil, nil, err
	}
	localIssues, err := indexSideRows("local", local[state.ExportIssues], false)
	if err != nil {
		return nil, nil, err
	}
	currentIssues, err := indexSideRows("tip", current[state.ExportIssues], false)
	if err != nil {
		return nil, nil, err
	}

	var issueOut, dispOut bytes.Buffer
	for _, id := range unionIssueIDs(localIssues, currentIssues) {
		issueRow, disp := pickIssueSide(localIssues[id], localDisp[id], currentIssues[id], currentDisp[id])
		if issueRow == nil || disp == nil {
			return nil, nil, fmt.Errorf("issue %q lacks an issue or disposition row", id)
		}
		issueOut.Write(issueRow.line)
		dispOut.Write(disp.line)
	}
	return issueOut.Bytes(), dispOut.Bytes(), nil
}

func unionIssueIDs(localIssues, currentIssues map[string]*row) []string {
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
	return sorted
}

func pickIssueSide(issueRow, disp, currentIssue, currentDispRow *row) (*row, *row) {
	switch {
	case currentIssue == nil || currentDispRow == nil:
		// local only
	case issueRow == nil || disp == nil:
		issueRow, disp = currentIssue, currentDispRow
	case currentDispRow.updated.After(disp.updated):
		issueRow, disp = currentIssue, currentDispRow
	}
	return issueRow, disp
}

func mergeWouldBlock(current, local Tree) ([]byte, error) {
	blocks := map[string]*row{}
	for _, source := range []Tree{current, local} {
		for number, line := range state.SplitLines(source[state.ExportWouldBlock]) {
			id, candidate, err := wouldBlockRow(number, line)
			if err != nil {
				return nil, err
			}
			// The later maintainer triage wins; local wins ties (it is visited last).
			if existing := blocks[id]; existing == nil || !existing.updated.After(candidate.updated) {
				blocks[id] = candidate
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
	return blockOut.Bytes(), nil
}

func wouldBlockRow(number int, line []byte) (string, *row, error) {
	var fields struct {
		ID        string `json:"id"`
		TriagedAt string `json:"triagedAt"`
	}
	if err := json.Unmarshal(line, &fields); err != nil || fields.ID == "" {
		return "", nil, fmt.Errorf("would_block row %d is not a keyed record", number+1)
	}
	triaged, err := parseTime(fields.TriagedAt, true)
	if err != nil {
		return "", nil, fmt.Errorf("would_block row %d has an unparseable triagedAt: %w", number+1, err)
	}
	return fields.ID, &row{line: append(append([]byte{}, line...), '\n'), updated: triaged}, nil
}

func mergeSignalFiles(current, local, merged Tree) error {
	for path := range merged {
		if !strings.HasPrefix(path, state.ExportSignalsDir) {
			continue
		}
		unioned, err := unionSignals(current[path], local[path])
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		merged[path] = unioned
	}
	return nil
}

func renderMergedManifest(current, local, merged Tree) error {
	meta, err := state.ReadManifestMeta(local[state.ExportManifest])
	if err != nil {
		return err
	}
	if currentMeta, err := state.ReadManifestMeta(current[state.ExportManifest]); err == nil && currentMeta.WouldBlock == state.WouldBlockPresent {
		meta.WouldBlock = state.WouldBlockPresent
	}
	delete(merged, state.ExportManifest)
	merged[state.ExportManifest] = state.RenderManifest(meta, merged)
	return nil
}

type row struct {
	line    []byte
	updated time.Time
}

// parseTime reads a row timestamp. An empty value is the zero time only when
// optional; anything else that does not parse is a merge error.
func parseTime(raw string, optional bool) (time.Time, error) {
	if raw == "" && optional {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339Nano, raw)
}

// indexRows maps issueId to its JSONL row (kept with its trailing LF). The
// updatedAt of each row must parse when requireTime is set.
func indexRows(data []byte, requireTime bool) (map[string]*row, error) {
	rows := map[string]*row{}
	for number, line := range state.SplitLines(data) {
		var fields struct {
			IssueID   string `json:"issueId"`
			UpdatedAt string `json:"updatedAt"`
		}
		if err := json.Unmarshal(line, &fields); err != nil || fields.IssueID == "" {
			return nil, fmt.Errorf("row %d is not a keyed record", number+1)
		}
		updated, err := parseTime(fields.UpdatedAt, !requireTime)
		if err != nil {
			return nil, fmt.Errorf("row %d has an unparseable updatedAt: %w", number+1, err)
		}
		rows[fields.IssueID] = &row{line: append(append([]byte{}, line...), '\n'), updated: updated}
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
