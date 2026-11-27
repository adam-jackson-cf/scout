package engine

import (
	"context"
	"strings"
	"testing"

	"scout/internal/modelcli"
)

func TestCapabilityMatrixReportsPerRepoAvailability(t *testing.T) {
	root := t.TempDir()
	deltaWrite(t, root, "main.go", "package main\n")
	deltaWrite(t, root, "app.py", "x = 1\n")
	t.Setenv("PATH", t.TempDir()) // no model CLI on PATH
	matrix, err := CapabilityMatrix(context.Background(), Config{}, root)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(matrix.InRepo, ",") != "Go,Python" || len(matrix.Rows) != 4 {
		t.Fatalf("matrix: %+v", matrix)
	}
	rows := map[string]MatrixRow{}
	for _, row := range matrix.Rows {
		rows[row.Capability] = row
	}
	assertMatrixReviewSkippedWithoutModelCLI(t, rows["Semantic review (contextual)"])
	graph := rows["Import graph / cycles; guardrails; rubric"]
	if graph.Cells["Go"].State != CellAvailable || graph.Cells["Rust"].State != CellPlanned || graph.Cells["C#"].State != CellNotPlanned {
		t.Fatalf("import graph row: %+v", graph.Cells)
	}
	crap := rows["CRAP (needs coverage artifact)"]
	if crap.Cells["Go"].State != CellWhen {
		t.Fatalf("crap without artifact: %+v", crap.Cells["Go"])
	}
	deltaWrite(t, root, "coverage/lcov.info", "TN:\n")
	assertMatrixCrapAvailableWithArtifact(t, root)
	assertMatrixRenderMarksPresentLanguages(t, matrix.Render())
}

func assertMatrixReviewSkippedWithoutModelCLI(t *testing.T, review MatrixRow) {
	t.Helper()
	for _, language := range MatrixLanguages {
		if cell := review.Cells[language]; cell.State != CellSkipped || cell.Detail != skippedDetail(modelcli.NoCLIReason) {
			t.Fatalf("%s review = %+v", language, cell)
		}
	}
}

func assertMatrixCrapAvailableWithArtifact(t *testing.T, root string) {
	t.Helper()
	again, err := CapabilityMatrix(context.Background(), Config{}, root)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range again.Rows {
		if strings.HasPrefix(row.Capability, "CRAP") && row.Cells["Go"].State != CellAvailable {
			t.Fatalf("crap with artifact: %+v", row.Cells["Go"])
		}
	}
}

func assertMatrixRenderMarksPresentLanguages(t *testing.T, text string) {
	t.Helper()
	for _, want := range []string{"Go *", "Python *", "CRAP (needs coverage artifact)"} {
		if !strings.Contains(text, want) {
			t.Fatalf("render lacks %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "Rust *") {
		t.Fatalf("a language absent from the repository is marked present:\n%s", text)
	}
}
