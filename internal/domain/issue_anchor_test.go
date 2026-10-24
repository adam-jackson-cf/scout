package domain

import (
	"strings"
	"testing"
	"time"
)

func TestNewIssueIDIsUniqueTimeOrderedULID(t *testing.T) {
	early := NewIssueID(FamilyQuality, time.UnixMilli(1_700_000_000_000))
	late := NewIssueID(FamilyQuality, time.UnixMilli(1_700_000_000_000+60_000))
	if !strings.HasPrefix(early, "issue:quality:") || len(early) != len("issue:quality:")+26 {
		t.Fatalf("unexpected ID shape %q", early)
	}
	for _, char := range strings.TrimPrefix(early, "issue:quality:") {
		if !strings.ContainsRune(crockford, char) {
			t.Fatalf("non-Crockford character %q in %q", char, early)
		}
	}
	if early >= late || early == NewIssueID(FamilyQuality, time.UnixMilli(1_700_000_000_000)) {
		t.Fatal("IDs must sort by time and never repeat")
	}
}

func TestAnchorIgnoresLocatorsButTracksEvidence(t *testing.T) {
	observation := func(startLine, endByte int, quote string) Observation {
		return Observation{Source: "scanner", RuleID: "rule", Mechanism: "Unsafe  Path", Evidence: []Evidence{{
			Location:      SourceRange{Path: "a.go", StartLine: startLine, EndLine: startLine, StartByte: 1, EndByte: endByte, Symbol: "pkg.F"},
			RedactedQuote: quote,
		}}}
	}
	moved, _ := AnchorOf(observation(10, 40, "eval(x)"))
	shifted, _ := AnchorOf(observation(30, 90, "  EVAL(x)  "))
	changed, _ := AnchorOf(observation(10, 40, "eval(y)"))
	if moved != shifted {
		t.Fatalf("line/byte locators or cosmetic whitespace changed the anchor: %#v vs %#v", moved, shifted)
	}
	if moved == changed {
		t.Fatal("different evidence produced the same fingerprint")
	}
}

func TestIssueWithoutAnchorIsInvalid(t *testing.T) {
	now := time.UnixMilli(1_700_000_000_000)
	issue := Issue{
		IssueID: "issue:security:x", Title: "t", Mechanism: "m", State: IssueOpen, OriginalSeverity: SeverityHigh, EffectiveSeverity: SeverityHigh,
		ObservationIDs: []string{"obs"}, History: []IssueHistoryEntry{{At: now, To: IssueOpen, Event: EventCreated, Reason: "first observed"}},
		CreatedAt: now, UpdatedAt: now,
	}
	if err := issue.Validate(); err == nil {
		t.Fatal("issue without a source anchor validated")
	}
	issue.Anchor = IssueAnchor{RuleKey: "scanner:rule", Path: "a.go"}
	if err := issue.Validate(); err != nil {
		t.Fatalf("anchored issue rejected: %v", err)
	}
}
