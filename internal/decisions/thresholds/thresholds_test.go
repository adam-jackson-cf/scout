package thresholds

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const question = "triage.real_defect"

func TestDefaultsAreProvisionalWithNoLabelsAndTheAssumedMinimums(t *testing.T) {
	r, err := Resolve(nil, question, "triage.real_defect@v1")
	if err != nil {
		t.Fatal(err)
	}
	e := r.Entry
	if r.Source != "default" || r.Stale || r.Status != StatusProvisional || e.Status != StatusProvisional {
		t.Fatalf("resolution %+v", r)
	}
	if e.LabelSet.Summary != "none" || e.ScoreSpace != SpaceRaw || e.MinimumsProvenance != "assumed" {
		t.Fatalf("entry %+v", e)
	}
	// Precision 0.8; recall raised to 0.9 because a miss costs more.
	if e.MinPrecision != 0.8 || e.MinRecall != 0.9 || e.Metric != "fbeta" || e.Beta != 2 {
		t.Fatalf("minimums %+v", e)
	}
	if e.Gating || e.Threshold == nil || *e.Threshold != 0.3 || e.CostRule == "" || len(e.Assumptions) == 0 {
		t.Fatalf("a provisional ungated threshold must record its value, cost rule and assumptions: %+v", e)
	}
	if e.VersionKey != "triage.real_defect@v1" {
		t.Fatalf("the default binds to the current key, got %q", e.VersionKey)
	}
}

func TestAnOverrideWithAnotherVersionKeyIsStaleAndNeverReady(t *testing.T) {
	override := []byte(`{"schema":1,"decisions":[{"question_id":"triage.real_defect","version_key":"triage.real_defect@old","gating":true,
	"threshold":0.62,"min_precision":0.9,"min_recall":0.8,"minimums_provenance":"agreed","metric":"f1","label_set":{"summary":"40/40 human, dev 60 test 20"},
	"score_space":"raw","status":"ready","cost_rule":"false action costs more","assumptions":[]}]}`)
	r, err := Resolve(override, question, "triage.real_defect@new")
	if err != nil {
		t.Fatal(err)
	}
	if !r.Stale || r.Status != StatusProvisional || r.Source != Path || r.Entry.Status != StatusReady {
		t.Fatalf("a key mismatch must be stale and provisional: %+v", r)
	}
	matching, err := Resolve(override, question, "triage.real_defect@old")
	if err != nil || matching.Stale || matching.Status != StatusReady || *matching.Entry.Threshold != 0.62 {
		t.Fatalf("a matching key keeps the file's entry: %+v %v", matching, err)
	}
}

func TestAnOverrideThatDoesNotListTheQuestionFallsBackToTheDefault(t *testing.T) {
	r, err := Resolve([]byte(`{"schema":1,"decisions":[]}`), question, "k")
	if err != nil || r.Source != "default" || r.Stale {
		t.Fatalf("%+v %v", r, err)
	}
}

func TestInvalidConfigurationIsAnErrorNotAGuess(t *testing.T) {
	cases := map[string]string{
		"not json":           `{`,
		"unknown field":      `{"schema":1,"decisions":[],"extra":1}`,
		"bad schema":         `{"schema":2,"decisions":[]}`,
		"ready without data": `{"schema":1,"decisions":[{"question_id":"triage.real_defect","min_precision":0.8,"min_recall":0.8,"minimums_provenance":"assumed","metric":"f1","label_set":{"summary":"none"},"score_space":"raw","status":"ready","cost_rule":"c","assumptions":[]}]}`,
		"bad minimum":        `{"schema":1,"decisions":[{"question_id":"triage.real_defect","min_precision":0,"min_recall":0.8,"minimums_provenance":"assumed","metric":"f1","label_set":{"summary":"none"},"score_space":"raw","status":"provisional","cost_rule":"c","assumptions":[]}]}`,
		"bad space":          `{"schema":1,"decisions":[{"question_id":"triage.real_defect","min_precision":0.8,"min_recall":0.8,"minimums_provenance":"assumed","metric":"f1","label_set":{"summary":"none"},"score_space":"isotonic","status":"provisional","cost_rule":"c","assumptions":[]}]}`,
	}
	for name, body := range cases {
		if _, err := Resolve([]byte(body), question, "k"); err == nil || !strings.Contains(err.Error(), Path) {
			t.Fatalf("%s: want an error naming the file, got %v", name, err)
		}
	}
	if _, err := Resolve(nil, "no.such.question", "k"); err == nil {
		t.Fatal("an unknown question must be an error")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// The override is read from the committed ref, so a working-tree edit (a pull
// request's own change) cannot move the thresholds it is ranked under.
func TestReadCommittedReadsTheRefNotTheWorkingTree(t *testing.T) {
	root := t.TempDir()
	runGit(t, root, "init", "-q")
	if err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-q", "-m", "one")
	if got, err := ReadCommitted(root, "HEAD"); err != nil || got != nil {
		t.Fatalf("no committed file must read as nil, got %q %v", got, err)
	}
	path := filepath.Join(root, filepath.FromSlash(Path))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(`{"schema":1,"decisions":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := ReadCommitted(root, "HEAD"); err != nil || got != nil {
		t.Fatalf("an uncommitted working-tree file must not be read, got %q %v", got, err)
	}
	runGit(t, root, "add", ".")
	runGit(t, root, "commit", "-q", "-m", "two")
	if got, err := ReadCommitted(root, "HEAD"); err != nil || string(got) != `{"schema":1,"decisions":[]}` {
		t.Fatalf("committed file %q %v", got, err)
	}
	if _, err := ReadCommitted(root, "no-such-ref"); err == nil {
		t.Fatal("an unresolvable ref must be an error")
	}
}
