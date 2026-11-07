package engine

import (
	"fmt"
	"strings"
	"testing"

	"scout/internal/areas"
	"scout/internal/association"
	"scout/internal/domain"
)

func TestSummaryPendingAreasPreserveLaneHealthWithinBound(t *testing.T) {
	pending := summaryPendingAreas(40)
	health := summaryPendingHealth(5)
	text := Summary(domain.Gate{}, Ranking{}, nil, nil, 0, health, pending, association.GuardState{})
	if n := strings.Count(text, "\n"); n > 60 {
		t.Errorf("summary has %d lines, want at most 60", n)
	}
	for _, lane := range health {
		row := fmt.Sprintf("| %s | report | false | ran |\n", lane.Lane)
		if !strings.Contains(text, row) {
			t.Errorf("summary dropped lane-health row %q", strings.TrimSpace(row))
		}
	}
	block := summaryPendingBlock(t, text)
	if n := strings.Count(block, "- `"); n != 8 {
		t.Errorf("pending block lists %d areas, want 8 (cap)", n)
	}
	for _, area := range pending[:8] {
		row := fmt.Sprintf("- `%s`: %s\n", area.ScoutAreaID, strings.Join(area.FileSet, ", "))
		if !strings.Contains(block, row) {
			t.Errorf("pending block dropped complete area row %q", strings.TrimSpace(row))
		}
	}
	if !strings.HasSuffix(block, "- … 32 more\n") {
		t.Errorf("pending block must end with %q", "- … 32 more")
	}
}

func TestSummaryPendingAreasSmallListIsComplete(t *testing.T) {
	pending := summaryPendingAreas(2)
	text := Summary(domain.Gate{}, Ranking{}, nil, nil, 0, summaryPendingHealth(5), pending, association.GuardState{})
	want := "- `pending-00`: src/area-00.go, src/area-00_test.go\n" +
		"- `pending-01`: src/area-01.go, src/area-01_test.go\n"
	if got := summaryPendingBlock(t, text); got != want {
		t.Errorf("small pending block = %q, want %q", got, want)
	}
}

func summaryPendingAreas(count int) []areas.Area {
	var pending []areas.Area
	for index := range count {
		pending = append(pending, areas.Area{
			ScoutAreaID: fmt.Sprintf("pending-%02d", index),
			FileSet: []string{
				fmt.Sprintf("src/area-%02d.go", index),
				fmt.Sprintf("src/area-%02d_test.go", index),
			},
		})
	}
	return pending
}

func summaryPendingHealth(count int) []domain.LaneHealth {
	var health []domain.LaneHealth
	for index := range count {
		health = append(health, domain.LaneHealth{
			Lane:   fmt.Sprintf("proof-lane-%d", index),
			Mode:   domain.ModeReport,
			Status: domain.LaneRan,
		})
	}
	return health
}

func summaryPendingBlock(t *testing.T, text string) string {
	t.Helper()
	_, after, ok := strings.Cut(text, "### Pending areas\n")
	if !ok {
		t.Fatal("summary omitted Pending areas heading")
	}
	block, _, ok := strings.Cut(after, "\n### Introduced / regressed")
	if !ok {
		t.Fatal("summary omitted section following Pending areas")
	}
	return block
}
