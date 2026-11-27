package eval

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// The v1 corpus regression (REQ-19): the 9
// adjudicated true issues stay detected and the 13 adjudicated false positives
// stay unflagged.
//
// Two checks, because the corpus targets are pinned remote repositories that
// tests cannot clone (no network):
//  1. Replay: the recorded Scout observations of the latest corpus run
//     (evaluation/scout-requests-v2.json) are compared with reference-v1.json
//     through CompareOutcomes.
//  2. Live rules: the bundled Semgrep rules behind the 9 true issues are run on
//     minimal fixtures reproducing each mechanism, plus the negative boundary.

const (
	wantValidatedTrue  = 9
	wantFalsePositives = 13
)

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCorpusReplayKeepsTrueIssuesAndDoesNotFlagFalsePositives(t *testing.T) {
	root := repoRoot(t)
	loaded, err := LoadCorpus(filepath.Join(root, "evaluation", "reference-v1.json"))
	if err != nil || loaded.Status != CorpusAvailable {
		t.Fatalf("load reference corpus: %s %v", loaded.Status, err)
	}
	data, err := os.ReadFile(filepath.Join(root, "evaluation", "scout-requests-v2.json"))
	if err != nil {
		t.Fatal(err)
	}
	var requests []EvaluationRequest
	if err := json.Unmarshal(data, &requests); err != nil {
		t.Fatal(err)
	}

	wantTrue, falsePositives, nFP := referenceTruth(loaded.Corpus)
	if len(wantTrue) != wantValidatedTrue || nFP != wantFalsePositives {
		t.Fatalf("reference corpus has %d validated true / %d false positives, want %d / %d", len(wantTrue), nFP, wantValidatedTrue, wantFalsePositives)
	}

	matched, flagged := replayRequests(t, loaded, requests, falsePositives)
	for id := range wantTrue {
		if !matched[id] {
			t.Errorf("true issue %s not detected", id)
		}
	}
	sort.Strings(flagged)
	if len(flagged) != 0 {
		t.Errorf("false positives flagged:\n%s", strings.Join(flagged, "\n"))
	}
	t.Logf("corpus replay: %d/%d true detected, %d/%d false positives unflagged", len(matched), len(wantTrue), nFP-len(flagged), nFP)
}

func referenceTruth(corpus *Corpus) (map[string]bool, map[string][]ReferenceIssue, int) {
	wantTrue := map[string]bool{}
	falsePositives := map[string][]ReferenceIssue{} // corpus entry id -> FP issues
	for _, entry := range corpus.Entries {
		for _, issue := range entry.Issues {
			switch issue.Adjudication {
			case AdjudicationValidatedTrue:
				wantTrue[issue.ID] = true
			case AdjudicationFalsePositive:
				falsePositives[entry.ID] = append(falsePositives[entry.ID], issue)
			}
		}
	}
	nFP := 0
	for _, issues := range falsePositives {
		nFP += len(issues)
	}
	return wantTrue, falsePositives, nFP
}

func replayRequests(t *testing.T, loaded CorpusLoadResult, requests []EvaluationRequest, falsePositives map[string][]ReferenceIssue) (map[string]bool, []string) {
	t.Helper()
	matched := map[string]bool{}
	flagged := []string{}
	for _, request := range requests {
		report, err := CompareOutcomes(loaded, request)
		if err != nil {
			t.Fatalf("compare %s/%s/%s: %v", request.Target.Repository, request.Language, request.Family, err)
		}
		for _, match := range report.Matches {
			matched[match.ReferenceIssueID] = true
		}
		for _, miss := range report.Misses {
			t.Errorf("true issue %s lost: %s %s", miss.ID, miss.Mechanism, miss.Path)
		}
		flagged = append(flagged, flaggedFalsePositives(falsePositives[report.CorpusEntryID], request.Observations)...)
	}
	return matched, flagged
}

func flaggedFalsePositives(issues []ReferenceIssue, observations []ScoutObservation) []string {
	var flagged []string
	for _, issue := range issues {
		for _, observation := range observations {
			if pathsRelevant(issue.Path, observation.Path) {
				flagged = append(flagged, issue.ID+" <- "+observation.Source+" "+observation.Path)
			}
		}
	}
	return flagged
}
