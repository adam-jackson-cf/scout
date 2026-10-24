package engine

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"scout/internal/modelcli"
)

// MatrixLanguages are the §5.3 columns, in order.
var MatrixLanguages = []string{"Go", "Python", "TS/JS", "Rust", "Java", "C#"}

var matrixExtensions = map[string]string{
	".go": "Go", ".py": "Python", ".ts": "TS/JS", ".tsx": "TS/JS", ".js": "TS/JS", ".jsx": "TS/JS", ".mjs": "TS/JS", ".cjs": "TS/JS",
	".rs": "Rust", ".java": "Java", ".cs": "C#",
}

// Cell states.
const (
	CellAvailable   = "available"
	CellUnavailable = "unavailable"
	CellPlanned     = "planned"
	CellNotPlanned  = "not planned"
	CellWhen        = "when produced"
	CellSkipped     = "skipped"
)

// MatrixCell is one capability for one language in this repository.
type MatrixCell struct {
	State  string `json:"state"`
	Detail string `json:"detail,omitempty"`
}

// MatrixRow is one §5.3 capability row.
type MatrixRow struct {
	Capability string                `json:"capability"`
	Cells      map[string]MatrixCell `json:"cells"`
}

// Matrix is the §5.3 capability matrix evaluated for one repository.
type Matrix struct {
	Languages []string `json:"languages"`
	// InRepo lists the languages detected in this repository.
	InRepo []string    `json:"inRepo"`
	Rows   []MatrixRow `json:"rows"`
}

// coverageArtifactNames are the conventional coverage report locations whose
// presence makes CRAP computable (artifacts are matched to a revision at
// assess time; the matrix reports only that a producer's output exists).
var coverageArtifactNames = []string{"lcov.info", "coverage/lcov.info", "coverage.xml", "cobertura.xml", "coverage/cobertura-coverage.xml", "coverage/coverage.xml"}

// CapabilityMatrix evaluates design §5.3 for the repository at root: which
// rows this machine and repository can deliver today. It runs nothing and
// calls no model; it only looks for files and executables.
func CapabilityMatrix(ctx context.Context, config Config, root string) (Matrix, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return Matrix{}, err
	}
	info, err := os.Stat(root)
	if err != nil {
		return Matrix{}, err
	}
	if !info.IsDir() {
		return Matrix{}, errors.New("assessment root is not a directory")
	}
	present, err := detectLanguages(ctx, root)
	if err != nil {
		return Matrix{}, err
	}
	matrix := Matrix{Languages: append([]string{}, MatrixLanguages...), InRepo: []string{}}
	for _, language := range MatrixLanguages {
		if present[language] {
			matrix.InRepo = append(matrix.InRepo, language)
		}
	}
	all := func(cell MatrixCell) map[string]MatrixCell {
		cells := make(map[string]MatrixCell, len(MatrixLanguages))
		for _, language := range MatrixLanguages {
			cells[language] = cell
		}
		return cells
	}

	// Row 1: measures, history and security need external tools or git.
	var missing []string
	for _, tool := range []string{"lizard", "jscpd", "git"} {
		if _, err := exec.LookPath(tool); err != nil {
			missing = append(missing, tool)
		}
	}
	if _, err := os.Stat(filepath.Join(root, ".git")); err != nil {
		missing = append(missing, "git history (.git)")
	}
	row1 := MatrixCell{State: CellAvailable}
	if len(missing) > 0 {
		row1 = MatrixCell{State: CellUnavailable, Detail: "missing: " + strings.Join(missing, ", ")}
	}
	// Row 2: import graph, guardrails, rubric by priority tier.
	row2 := map[string]MatrixCell{
		"Go": {State: CellAvailable, Detail: "P1"}, "Python": {State: CellAvailable, Detail: "P1"}, "TS/JS": {State: CellAvailable, Detail: "P1"},
		"Rust": {State: CellPlanned, Detail: "P3"}, "Java": {State: CellPlanned, Detail: "P3"}, "C#": {State: CellNotPlanned},
	}
	// Row 3: CRAP needs a coverage artifact.
	row3 := MatrixCell{State: CellWhen, Detail: "no coverage artifact found"}
	for _, name := range coverageArtifactNames {
		if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(name))); err == nil && info.Mode().IsRegular() {
			row3 = MatrixCell{State: CellAvailable, Detail: "artifact " + name}
			break
		}
	}
	// Row 4: semantic review needs an opted-in model CLI.
	row4 := modelReviewCell(config)

	matrix.Rows = []MatrixRow{
		{Capability: "CC/NLOC/duplication; history (git); secrets/CI security", Cells: all(row1)},
		{Capability: "Import graph / cycles; guardrails; rubric", Cells: row2},
		{Capability: "CRAP (needs coverage artifact)", Cells: all(row3)},
		{Capability: "Semantic review (contextual)", Cells: all(row4)},
	}
	return matrix, nil
}

// skippedDetail turns a "skipped (reason)" lane reason into its reason.
func skippedDetail(reason string) string {
	return strings.TrimSuffix(strings.TrimPrefix(reason, "skipped ("), ")")
}

func modelReviewCell(config Config) MatrixCell {
	name, _, err := modelcli.Detect(config.Provider.CLI)
	switch {
	case errors.Is(err, modelcli.ErrNoCLI):
		return MatrixCell{State: CellSkipped, Detail: skippedDetail(modelcli.NoCLIReason)}
	case err != nil:
		return MatrixCell{State: CellUnavailable, Detail: err.Error()}
	case !config.laneOptedIn(LaneContextual):
		return MatrixCell{State: CellSkipped, Detail: skippedDetail(ReasonNotOptedIn) + "; " + string(name) + " found"}
	}
	return MatrixCell{State: CellAvailable, Detail: "opted in via " + string(name)}
}

// detectLanguages finds the matrix languages with source files in the
// repository: tracked files when git answers, else a bounded walk.
func detectLanguages(ctx context.Context, root string) (map[string]bool, error) {
	found := map[string]bool{}
	note := func(path string) {
		if language, ok := matrixExtensions[strings.ToLower(filepath.Ext(path))]; ok {
			found[language] = true
		}
	}
	if out, err := git(ctx, root, "ls-files", "-z", "--cached", "--others", "--exclude-standard"); err == nil {
		for _, name := range strings.Split(string(out), "\x00") {
			note(name)
		}
		return found, nil
	}
	skip := map[string]bool{".git": true, "node_modules": true, "vendor": true, ".scout": true, ".codegraph": true}
	budget := 20000
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if skip[entry.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if budget--; budget < 0 {
			return filepath.SkipAll
		}
		note(path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("detect repository languages: %w", err)
	}
	return found, nil
}

// Render prints the matrix as a Markdown table. Cells read `✓ (detail)` for
// available, `–` for not planned, and the state text otherwise; a language
// present in the repository is marked with `*` in its heading.
func (m Matrix) Render() string {
	present := map[string]bool{}
	for _, language := range m.InRepo {
		present[language] = true
	}
	var b strings.Builder
	b.WriteString("| Capability |")
	for _, language := range m.Languages {
		mark := ""
		if present[language] {
			mark = " *"
		}
		b.WriteString(" " + language + mark + " |")
	}
	b.WriteString("\n|---|")
	for range m.Languages {
		b.WriteString("---|")
	}
	b.WriteString("\n")
	for _, row := range m.Rows {
		b.WriteString("| " + row.Capability + " |")
		for _, language := range m.Languages {
			b.WriteString(" " + renderCell(row.Cells[language]) + " |")
		}
		b.WriteString("\n")
	}
	languages := append([]string{}, m.InRepo...)
	sort.Strings(languages)
	if len(languages) == 0 {
		languages = []string{"none detected"}
	}
	b.WriteString("\n* present in this repository: " + strings.Join(languages, ", ") + "\n")
	return b.String()
}

func renderCell(cell MatrixCell) string {
	var text string
	switch cell.State {
	case CellAvailable:
		text = "✓"
	case CellNotPlanned:
		return "–"
	case "":
		return "–"
	default:
		text = cell.State
	}
	if cell.Detail != "" {
		text += " (" + cell.Detail + ")"
	}
	return text
}
