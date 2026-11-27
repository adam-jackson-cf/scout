package baseline

import (
	"encoding/json"
	"strings"
	"testing"

	"scout/internal/state"
)

func issueLedger(t *testing.T, updatedAt string, ids ...string) Tree {
	t.Helper()
	files := Tree{state.ExportWouldBlock: nil, state.ExportCoverageSummary: []byte("{}\n"), state.ExportPRRuns: nil, state.ExportShadowTriage: nil}
	var issues, dispositions []byte
	for _, id := range ids {
		issue, _ := state.CanonicalLine(map[string]any{"issueId": id, "title": "t"})
		disposition, _ := state.CanonicalLine(map[string]any{"issueId": id, "state": "open", "updatedAt": updatedAt})
		issues, dispositions = append(issues, issue...), append(dispositions, disposition...)
	}
	files[state.ExportIssues], files[state.ExportDispositions] = issues, dispositions
	files[state.ExportManifest] = state.RenderManifest(state.ManifestMeta{RepositoryID: "repo", WouldBlock: state.WouldBlockPresent}, files)
	return files
}

func issueIDs(t *testing.T, tree Tree) []string {
	t.Helper()
	var ids []string
	for _, line := range state.SplitLines(tree[state.ExportIssues]) {
		var row struct {
			IssueID string `json:"issueId"`
		}
		if err := json.Unmarshal(line, &row); err != nil {
			t.Fatal(err)
		}
		ids = append(ids, row.IssueID)
	}
	return ids
}

func TestMergeKeepsNewerDispositionPerIssue(t *testing.T) {
	localWithAAndOlderB := issueLedger(t, "2026-10-01T00:00:00Z", "issue:a", "issue:b")
	tipWithNewerB := issueLedger(t, "2026-10-02T00:00:00Z", "issue:b")
	merged, err := MergeState(tipWithNewerB, localWithAAndOlderB)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(issueIDs(t, merged), ","); got != "issue:a,issue:b" {
		t.Fatalf("merged issues = %q", got)
	}
	if !strings.Contains(string(merged[state.ExportDispositions]), "2026-10-02T00:00:00Z") {
		t.Fatalf("newer tip disposition lost: %s", merged[state.ExportDispositions])
	}
}

func TestMergeRejectsUnparseableDispositionTimestamp(t *testing.T) {
	tip := issueLedger(t, "2026-10-01T00:00:00Z", "issue:a")
	for name, bad := range map[string]string{"garbage": "not-a-time", "empty": ""} {
		t.Run(name, func(t *testing.T) {
			local := issueLedger(t, bad, "issue:a")
			if _, err := MergeState(tip, local); err == nil {
				t.Fatal("a disposition without a parseable updatedAt must fail the merge, not count as infinitely old")
			}
			if _, err := MergeState(local, tip); err == nil {
				t.Fatal("an unparseable tip timestamp must fail the merge")
			}
		})
	}
}
