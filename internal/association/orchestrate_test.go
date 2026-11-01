package association

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir, "-c", "user.name=t", "-c", "user.email=t@example.com", "-c", "commit.gpgsign=false"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return strings.TrimSpace(string(out))
}

func write(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

const body = "package svc\n\nfunc Handle() string {\n\treturn \"one\"\n}\n\nfunc Other() string {\n\treturn \"two\"\n}\n"

// TestRunThreeMerges: merge 1 (a real --no-ff merge) adds a residual file that
// a stub decision assigns; merge 2 modifies it; merge 3 renames it.
func TestRunThreeMerges(t *testing.T) {
	dir := t.TempDir()
	gitIn(t, dir, "init", "-q", "-b", "main")
	write(t, dir, "README.md", "hi\n")
	write(t, dir, "core/known.go", "package core\n")
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "base")
	base := gitIn(t, dir, "rev-parse", "HEAD")

	gitIn(t, dir, "checkout", "-q", "-b", "feature")
	write(t, dir, "svc/handler.go", body)
	write(t, dir, "docs/notes.md", "ignored\n")
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "feature")
	gitIn(t, dir, "checkout", "-q", "main")
	gitIn(t, dir, "merge", "-q", "--no-ff", "-m", "merge 1", "feature")
	m1 := gitIn(t, dir, "rev-parse", "HEAD")

	write(t, dir, "svc/handler.go", body+"\nfunc Third() {}\n")
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "merge 2")
	m2 := gitIn(t, dir, "rev-parse", "HEAD")

	gitIn(t, dir, "mv", "svc/handler.go", "svc/handlers.go")
	gitIn(t, dir, "commit", "-q", "-m", "merge 3")
	head := gitIn(t, dir, "rev-parse", "HEAD")

	repo := Repo{Dir: dir}
	own := newFixtureOwnership(fxRound{Owners: map[string][]string{"core/known.go": {"A0"}}})
	var calls []string
	decide := func(_ context.Context, m Merge, decided []Cluster) ([]Decision, error) {
		calls = append(calls, m.SHA)
		out := make([]Decision, len(decided))
		for i := range out {
			out[i] = Decision{Owners: []string{"A1"}}
		}
		return out, nil
	}

	res, err := Run(context.Background(), repo, base, head, own, decide)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) != 3 || res[0].Merge.SHA != m1 || res[1].Merge.SHA != m2 || res[2].Merge.SHA != head {
		t.Fatalf("merges not first-parent, oldest first: %+v", res)
	}
	// merge 1: the markdown file is excluded; svc/handler.go is one decided cluster, assigned by the stub.
	if len(res[0].Decided) != 1 || !reflect.DeepEqual(res[0].Decided[0].Files, []string{"svc/handler.go"}) || res[0].Counts["excluded:markdown"] != 1 {
		t.Fatalf("merge 1: %+v", res[0])
	}
	// merge 2: the modification of the now-owned file is not decided.
	if len(res[1].Decided) != 0 || res[1].Counts["modified_owned"] != 1 {
		t.Fatalf("merge 2: %+v", res[1])
	}
	// merge 3: the rename inherits the owner, with an association record's data, and is not decided.
	want := []Inherited{{Path: "svc/handlers.go", Rule: "rename", From: "svc/handler.go", Areas: []string{"A1"}}}
	if len(res[2].Decided) != 0 || !reflect.DeepEqual(res[2].Inherited, want) {
		t.Fatalf("merge 3: %+v", res[2])
	}
	if !reflect.DeepEqual(calls, []string{m1}) {
		t.Fatalf("decider called for %v, want only merge 1", calls)
	}

	// A later push starts at the earlier head: earlier merges are not listed, nothing is re-decided.
	own2 := newFixtureOwnership(fxRound{Owners: map[string][]string{"core/known.go": {"A0"}, "svc/handlers.go": {"A1"}}})
	calls = nil
	write(t, dir, "core/known.go", "package core\n\nvar X = 1\n")
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "merge 4")
	res2, err := Run(context.Background(), repo, head, "HEAD", own2, decide)
	if err != nil {
		t.Fatal(err)
	}
	if len(res2) != 1 || len(calls) != 0 || res2[0].Counts["modified_owned"] != 1 {
		t.Fatalf("second push: %+v calls %v", res2, calls)
	}
}

func TestRunCapAndPendingMove(t *testing.T) {
	dir := t.TempDir()
	gitIn(t, dir, "init", "-q", "-b", "main")
	write(t, dir, "pool/old.go", strings.Repeat("distinct line of text\n", 20))
	write(t, dir, "x/empty.txt", "")
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "base")
	base := gitIn(t, dir, "rev-parse", "HEAD")

	gitIn(t, dir, "mv", "pool/old.go", "pool/new.go")
	for i := range 12 { // 12 clusters in one merge: 10 decided, 2 pr_cap
		write(t, dir, filepath.Join("d"+string(rune('a'+i)), "f.go"), body+strings.Repeat("// x\n", i))
	}
	write(t, dir, "x/other_empty.txt", "") // inert: only empty files
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "commit", "-q", "-m", "merge 1")

	own := &fixtureOwnership{owners: map[string][]string{}, heur: map[string]bool{}, pool: map[string]bool{"pool/old.go": true}, sibs: map[SiblingKey][]string{}}
	res, err := Run(context.Background(), Repo{Dir: dir}, base, "HEAD", own, nil)
	if err != nil {
		t.Fatal(err)
	}
	r := res[0]
	if len(r.Decided) != 10 || len(r.Capped) != 2 || len(r.Inert) != 1 {
		t.Fatalf("decided %d capped %d inert %d", len(r.Decided), len(r.Capped), len(r.Inert))
	}
	if !reflect.DeepEqual(r.PendingMoves, []PendingMove{{Old: "pool/old.go", New: "pool/new.go", Pending: []string{"pool"}}}) {
		t.Fatalf("pending moves %+v", r.PendingMoves)
	}
}
