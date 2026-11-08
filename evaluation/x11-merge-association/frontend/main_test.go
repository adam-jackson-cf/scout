package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"scout/internal/association"
)

const fixtureGitPath = "/opt/homebrew/Cellar/git/2.56.0_2/bin/git"

func git(t *testing.T, repo string, args ...string) string {
	t.Helper()
	cmd := exec.Command(fixtureGitPath, append([]string{"-C", repo}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL=/dev/null", "GIT_NO_LAZY_FETCH=1", "GIT_NO_REPLACE_OBJECTS=1", "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func put(t *testing.T, repo, path, content string) {
	t.Helper()
	full := filepath.Join(repo, path)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

type fixtureRequest struct {
	request
	Heuristic map[string][]string
}

func (r fixtureRequest) stream() ([]byte, error) {
	var input bytes.Buffer
	encoder := json.NewEncoder(&input)
	if err := encoder.Encode(r.request); err != nil {
		return nil, err
	}
	command := exec.Command(fixtureGitPath, "-C", r.Repo, "rev-list", "--first-parent", "--reverse", r.Base+".."+r.Head)
	command.Env = append(os.Environ(), "GIT_NO_REPLACE_OBJECTS=1")
	raw, err := command.Output()
	if err != nil {
		return nil, err
	}
	order := append([]string{r.Base}, strings.Fields(string(raw))...)
	for _, sha := range order {
		paths, ok := r.Heuristic[sha]
		if !ok {
			continue
		}
		if err := encoder.Encode(ownershipRecord{Revision: sha, Paths: paths}); err != nil {
			return nil, err
		}
	}
	return input.Bytes(), nil
}

func fixture(t *testing.T) fixtureRequest {
	t.Helper()
	repo := t.TempDir()
	git(t, repo, "init", "--quiet")
	git(t, repo, "config", "user.name", "Synthetic parity")
	git(t, repo, "config", "user.email", "parity@example.invalid")
	git(t, repo, "config", "commit.gpgsign", "false")
	git(t, repo, "config", "diff.renameEmpty", "true")
	put(t, repo, "type-change", "ordinary file\n")
	put(t, repo, "empty-before", "")
	put(t, repo, "owned.go", "package before\n")
	git(t, repo, "add", "--all")
	git(t, repo, "commit", "--quiet", "-m", "synthetic base")
	base := git(t, repo, "rev-parse", "HEAD")

	if err := os.Remove(filepath.Join(repo, "type-change")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("owned.go", filepath.Join(repo, "type-change")); err != nil {
		t.Fatal(err)
	}
	git(t, repo, "mv", "empty-before", "empty-after")
	for _, path := range []string{".gitignore", ".editorconfig", "trailing.", "dotted.dir/source.go", "dotted.png/code", "odd\tline\n雪.go", "empty-after.go"} {
		put(t, repo, path, "synthetic content\n")
	}
	// More than ten live clusters exercise Cap's ordering and excluded tail.
	for i := range 14 {
		put(t, repo, fmt.Sprintf("many/file%02d.go", i), "package synthetic\n")
	}
	git(t, repo, "add", "--all")
	// A gitlink is a commit object, not a blob. No blob-content read is valid.
	git(t, repo, "update-index", "--add", "--cacheinfo", "160000,"+base+",modules/dependency")
	git(t, repo, "commit", "--quiet", "-m", "synthetic edge cases")
	first := git(t, repo, "rev-parse", "HEAD")
	put(t, repo, "owned.go", "package after\n")
	put(t, repo, "later/new.go", "package later\n")
	git(t, repo, "add", "--all")
	git(t, repo, "commit", "--quiet", "-m", "synthetic per-SHA ownership")
	head := git(t, repo, "rev-parse", "HEAD")
	return fixtureRequest{request: request{Repo: repo, Base: base, Head: head}, Heuristic: map[string][]string{base: {"nested_an/absent.go", "root.go"}, first: {}, head: {"owned.go"}}}
}

// expectedOwnership is deliberately independent of the harness adapter.
type expectedOwnership struct{ paths []string }

func (expectedOwnership) AreaOwners(string) []string { return nil }
func (expectedOwnership) PendingAreas(string) []string { return nil }
func (expectedOwnership) SiblingFiles(association.SiblingKey) []string { return nil }
func (o expectedOwnership) HeuristicOnly(path string) bool {
	for _, p := range o.paths {
		if p == path {
			return true
		}
	}
	return false
}

func compareClusters(t *testing.T, got []cluster, want []association.Cluster) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("cluster lengths: got %d, want %d", len(got), len(want))
	}
	for i, c := range want {
		if got[i].ID != c.ID || got[i].Directory != c.Dir || len(got[i].Members) != len(c.Members) {
			t.Fatalf("cluster %d metadata differs: got %+v, want %+v", i, got[i], c)
		}
		for j, m := range c.Members {
			g := got[i].Members[j]
			if g.Path != m.Path || g.Kind != m.Kind || g.Old != m.Old || g.Empty != m.Empty {
				t.Fatalf("cluster %d member %d: got %+v, want %+v", i, j, g, m)
			}
		}
	}
}

func TestJSONParityWithProduct(t *testing.T) {
	req := fixture(t)
	input, err := req.stream()
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run(context.Background(), bytes.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	var got response
	if err := json.Unmarshal(output.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	product, err := (association.Repo{Dir: req.Repo}).Merges(context.Background(), req.Base, req.Head)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Merges) != len(product) || len(product) != 2 {
		t.Fatalf("merge lengths: got %d, product %d", len(got.Merges), len(product))
	}
	seen := map[string]association.Change{}
	for i, m := range product {
		g := got.Merges[i]
		if g.Commit != m.SHA || g.Parent != m.Parent || len(g.Changes) != len(m.Changes) {
			t.Fatalf("merge %d metadata differs", i)
		}
		for j, c := range m.Changes {
			seen[c.Path] = c
			actual := g.Changes[j]
			if actual.Path != c.Path || actual.Old != c.OldPath || actual.Kind != string(c.Kind) || actual.Empty != c.Empty {
				t.Fatalf("merge %d change %d differs: got %+v, product %+v", i, j, actual, c)
			}
		}
		results, err := association.Run(context.Background(), association.Repo{Dir: req.Repo}, m.Parent, m.SHA, expectedOwnership{paths: req.Heuristic[m.SHA]}, nil)
		if err != nil || len(results) != 1 {
			t.Fatalf("product Run: results %d, error %v", len(results), err)
		}
		result := results[0]
		if !reflect.DeepEqual(g.Counts, result.Counts) {
			t.Fatalf("merge %d classification counts: got %v, product %v", i, g.Counts, result.Counts)
		}
		compareClusters(t, g.Selected, result.Decided)
		compareClusters(t, g.Capped, result.Capped)
	}
	for _, path := range []string{".gitignore", ".editorconfig", "trailing.", "dotted.dir/source.go", "modules/dependency"} {
		if _, ok := seen[path]; !ok {
			t.Fatalf("fixture did not exercise %q", path)
		}
	}
	if seen["type-change"].Kind != association.Modified {
		t.Fatal("fixture did not exercise product T-to-Modified handling")
	}
	if c := seen["empty-after"]; c.Kind != association.Renamed || !c.Empty {
		t.Fatalf("fixture did not exercise empty rename: %+v", c)
	}
	if len(got.Merges[0].Capped) == 0 || got.Merges[1].Counts["heuristic_only"] != 1 {
		t.Fatal("fixture did not exercise cap or exact-SHA heuristic ownership")
	}
	if !reflect.DeepEqual(got.OwnedDirectories, []string{"", "nested"}) {
		t.Fatalf("product normalized owned directories: %v", got.OwnedDirectories)
	}
}

func TestInvalidGitBytePathRefusesWithoutPartialOutput(t *testing.T) {
	req := fixture(t)
	// APFS refuses non-UTF-8 working-tree filenames. Construct the git tree
	// directly so the byte-path fixture exercises git objects on every host.
	cmd := exec.Command(fixtureGitPath, "-C", req.Repo, "ls-tree", "-z", req.Head)
	treeBytes, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	blob := git(t, req.Repo, "rev-parse", req.Head+":owned.go")
	treeBytes = append(treeBytes, []byte("100644 blob "+blob+"\tinvalid-\xff.go\x00")...)
	cmd = exec.Command(fixtureGitPath, "-C", req.Repo, "mktree", "-z")
	cmd.Stdin = bytes.NewReader(treeBytes)
	tree, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	req.Head = git(t, req.Repo, "commit-tree", strings.TrimSpace(string(tree)), "-p", req.Head, "-m", "synthetic invalid UTF-8 path")
	req.Heuristic[req.Head] = []string{}
	input, err := req.stream()
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err = run(context.Background(), bytes.NewReader(input), &output)
	if err == nil || !strings.Contains(err.Error(), "invalid UTF-8 git path") || output.Len() != 0 {
		t.Fatalf("want explicit UTF-8 refusal and no partial output, got error %v, bytes %d", err, output.Len())
	}
}

func TestMissingExactSHAOwnershipRefusesWithoutOutput(t *testing.T) {
	req := fixture(t)
	delete(req.Heuristic, req.Head)
	input, err := req.stream()
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	err = run(context.Background(), bytes.NewReader(input), &output)
	if err == nil || !strings.Contains(err.Error(), "exact merge SHA") || output.Len() != 0 {
		t.Fatalf("want exact-SHA refusal and no output, got error %v, bytes %d", err, output.Len())
	}
}

func TestUnknownRequestFieldsRefuse(t *testing.T) {
	for _, field := range []string{"label", "verdict", "model"} {
		var output bytes.Buffer
		err := run(context.Background(), strings.NewReader(`{"`+field+`":"injected"}`), &output)
		if err == nil || !strings.Contains(err.Error(), "unknown field") || output.Len() != 0 {
			t.Fatalf("unknown %s: error %v, output %q", field, err, output.String())
		}
	}
}

func TestPreparationOwnershipMatchesProductLoader(t *testing.T) {
	state := t.TempDir()
	features := filepath.Join(state, "features")
	if err := os.Mkdir(features, 0755); err != nil {
		t.Fatal(err)
	}
	for i, spec := range []struct{ source, status, path string }{
		{"python-source-group", "active", "absent/from/tree.py"},
		{"package-json-bin", "active", "bin/tool.js"},
		{"python-source-group", "skipped", "stale.py"},
		{"agent-mapper", "active", "agent.py"},
		{"python-source-group", "active", "bin/tool.js"},
	} {
		raw, err := json.Marshal(map[string]any{
			"schemaVersion": 1, "featureId": fmt.Sprintf("fixture-%d", i),
			"kind": "module", "source": spec.source, "status": spec.status,
			"confidence": "high", "ownedFiles": []map[string]string{{"path": spec.path, "reason": "synthetic"}},
		})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(features, fmt.Sprintf("%d.json", i)), raw, 0644); err != nil {
			t.Fatal(err)
		}
	}
	product, err := loadHeuristic(state)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(product, []string{"absent/from/tree.py", "bin/tool.js"}) {
		t.Fatalf("product absent-path/agent/stale/dedup rule: %v", product)
	}
	metadata, err := os.ReadFile("../tool-pins-v6/python-runtime-v6b.json")
	if err != nil {
		t.Fatal(err)
	}
	var runtime struct { Path string `json:"path"` }
	if err := json.Unmarshal(metadata, &runtime); err != nil {
		t.Fatal(err)
	}
	script := `
import hashlib,json,runpy,sys
from pathlib import Path
here=Path.cwd().parent
config=json.loads((here/'inventory-inputs.json').read_text())
pins={str(Path(p['path']).resolve()):p['sha256'] for p in config['recipe_sources']}
bootstrap=here/'bootstrap_v6.py'
if hashlib.sha256(bootstrap.read_bytes()).hexdigest()!=pins[str(bootstrap)]:
    raise RuntimeError('reviewed bootstrap source drift')
pin=runpy.run_path(str(bootstrap))['verified_helper'](config,here)
pin.verify_python_runtime(config)
pin.verify_preparation_closure(config)
source=here/'prepare_exports.py'
if hashlib.sha256(source.read_bytes()).hexdigest()!=pins[str(source)]:
    raise RuntimeError('reviewed preparation source drift')
from prepare_exports import heuristic_paths
print(json.dumps(heuristic_paths(Path(sys.argv[1]))))
`
	cmd := exec.Command(runtime.Path, "-I", "-B", "-c", script, state)
	raw, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("preparation helper: %v: %s", err, raw)
	}
	var preparation []string
	if err := json.Unmarshal(raw, &preparation); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(preparation, product) {
		t.Fatalf("preparation %v differs from product %v", preparation, product)
	}
}
