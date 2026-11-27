package association

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"testing"
)

type fxResult struct {
	Counts   map[string]int `json:"counts"`
	Clusters []struct {
		ID        string   `json:"id"`
		Dir       string   `json:"dir"`
		Stem      string   `json:"stem"`
		Files     []string `json:"files"`
		Kinds     []string `json:"kinds"`
		Heuristic []string `json:"heuristic"`
	} `json:"clusters"`
	Inert []struct {
		Dir   string   `json:"dir"`
		Stem  string   `json:"stem"`
		Files []string `json:"files"`
	} `json:"inert"`
	Decided  []string `json:"decided"`
	Capped   []string `json:"capped"`
	Resolved []struct {
		Path   string   `json:"path"`
		Rule   string   `json:"rule"`
		Owners []string `json:"owners"`
		Old    *string  `json:"old"`
		Pool   bool     `json:"pool"`
	} `json:"resolved"`
}

type fxCommit struct {
	SHA      string          `json:"sha"`
	Entries  [][]interface{} `json:"entries"` // [status, old, new, empty]
	Tree     []string        `json:"tree"`
	Pilot    fxResult        `json:"pilot"`
	Expected fxResult        `json:"expected"`
	Affected []string        `json:"affected"`
}

type fxRound struct {
	Round         int                 `json:"round"`
	CommitsTotal  int                 `json:"commits_total"`
	Commits       []fxCommit          `json:"commits"`
	Owners        map[string][]string `json:"owners"`
	HeuristicOnly []string            `json:"heuristic_only"`
	Pool          []string            `json:"pool"`
}

type fxRepo struct {
	Repo   string    `json:"repo"`
	Rounds []fxRound `json:"rounds"`
}

// fixtureOwnership is the round-start membership: owners, heuristic-only files
// and the pool (a pending area with id "pool").
func newFixtureOwnership(r fxRound) *storeOwnership {
	o := &storeOwnership{owners: r.Owners, pending: map[string][]string{}, heuristic: map[string]bool{}, siblings: map[SiblingKey][]string{}}
	for _, p := range r.HeuristicOnly {
		o.heuristic[p] = true
	}
	for _, p := range r.Pool {
		o.pending[p] = []string{"pool"}
	}
	for p := range r.Owners {
		k := SiblingKeyOf(p)
		o.siblings[k] = append(o.siblings[k], p)
	}
	return o
}

func loadFixture(t *testing.T, name string) fxRepo {
	t.Helper()
	f, err := os.Open("testdata/parity-" + name + ".json.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}()
	zr, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	var fx fxRepo
	if err := json.NewDecoder(zr).Decode(&fx); err != nil {
		t.Fatal(err)
	}
	return fx
}

func changesOf(t *testing.T, c fxCommit) []Change {
	var out []Change
	for _, e := range c.Entries {
		st, old, nw, empty := e[0].(string), e[1].(string), e[2].(string), e[3].(float64) == 1
		switch st {
		case "D":
			out = append(out, Change{Kind: Deleted, Path: old})
		case "R":
			out = append(out, Change{Kind: Renamed, OldPath: old, Path: nw, Empty: empty})
		case "A", "C":
			out = append(out, Change{Kind: Added, Path: nw, Empty: empty})
		case "M", "T":
			out = append(out, Change{Kind: Modified, Path: nw})
		default:
			t.Fatalf("%s: status %q", c.SHA, st)
		}
	}
	return out
}

// intentionalDifferences pins the cases where the pilot's two unported rules
// change the result, keyed "<repo>:<sha>" -> rules. Everything else must equal the pilot.
//   - fastapi round 8, 6e6c74fc: a new Hindi translation tree (docs/hi/...). The pilot excludes
//     docs/hi/mkdocs.yml as a translation; Scout has no translation rule, so it is one residual cluster.
//
// The pilot's cross-scope rename rule never fires in the recorded windows (no case); it is covered by
// TestUnportedRules.
var intentionalDifferences = map[string][]string{
	"fastapi:6e6c74fc162a4cc6421deb928ad403c9ee0f9d4f": {"translation"}, // gitleaks:allow (public fastapi commit SHA, not a secret)
}

func TestParityWithPilot(t *testing.T) {
	for _, name := range []string{"fastapi", "axum"} {
		fx := loadFixture(t, name)
		compared, identical := 0, 0
		diffs := map[string][]string{}
		byRule := map[string]int{}
		for _, r := range fx.Rounds {
			own := newFixtureOwnership(r)
			for _, c := range r.Commits {
				compared++
				tree := map[string]bool{}
				for _, p := range c.Tree {
					tree[p] = true
				}
				got := summarise(c.SHA, Classify(changesOf(t, c), own, func(p string) bool { return tree[p] }))
				if len(c.Affected) == 0 {
					identical++
					if !reflect.DeepEqual(c.Pilot, c.Expected) {
						t.Fatalf("%s %s: fixture pilot and expected differ without an affected rule", name, c.SHA)
					}
				} else {
					diffs[c.SHA] = c.Affected
					for _, a := range c.Affected {
						byRule[a]++
					}
					if reflect.DeepEqual(c.Pilot, c.Expected) {
						t.Errorf("%s %s: marked affected but pilot == expected", name, c.SHA)
					}
				}
				// Scout's result must be the expected one (for parity cases that is the pilot's own result).
				compareResult(t, fmt.Sprintf("%s round %d %s", name, r.Round, c.SHA), got, c.Expected)
			}
		}
		t.Logf("%s: %d cases compared (interesting commits of %d first-parent commits), %d identical to the pilot, %d intentional differences %v",
			name, compared, total(fx), identical, len(diffs), byRule)
		if want := intentionalDifferences; !sameDiffs(diffs, want, name) {
			t.Errorf("%s: intentional differences changed:\n got %v", name, diffs)
		}
	}
}

func total(fx fxRepo) int {
	n := 0
	for _, r := range fx.Rounds {
		n += r.CommitsTotal
	}
	return n
}

// sameDiffs compares the observed differences of one repository with the pinned
// ones (pinned keys are "<repo>:<sha>").
func sameDiffs(got map[string][]string, pinned map[string][]string, repo string) bool {
	want := map[string][]string{}
	for k, v := range pinned {
		if len(k) > len(repo) && k[:len(repo)+1] == repo+":" {
			want[k[len(repo)+1:]] = v
		}
	}
	return reflect.DeepEqual(got, want) || (len(got) == 0 && len(want) == 0)
}

// summary is the comparable form of Classify + clustering + cap.
type summary struct {
	Counts   map[string]int
	Clusters []string // id|dir|stem|files|kinds|heuristic
	Inert    []string
	Decided  []string
	Capped   []string
	Resolved []string // path|rule|owners|old
}

func join(s []string) string {
	out, _ := json.Marshal(s)
	return string(out)
}

func summarise(sha string, r Result) summary {
	live, inert := BuildClusters(sha, r.Residual)
	s := summary{Counts: summaryCounts(r.Counts, len(inert), len(live))}
	for _, c := range live {
		s.Clusters = append(s.Clusters, clusterLine(c))
	}
	for _, c := range inert {
		s.Inert = append(s.Inert, c.Dir+"|"+c.Stem+"|"+join(c.Files))
	}
	dec, capd := Cap(live)
	for _, c := range dec {
		s.Decided = append(s.Decided, c.ID)
	}
	for _, c := range capd {
		s.Capped = append(s.Capped, c.ID)
	}
	s.Resolved = resolvedLines(r)
	return s
}

func summaryCounts(counts map[string]int, inert, live int) map[string]int {
	out := map[string]int{}
	for k, v := range counts {
		out[k] = v
	}
	out["inert_clusters"] = inert
	out["residual_clusters"] = live
	for k, v := range out {
		if v == 0 {
			delete(out, k)
		}
	}
	return out
}

func clusterLine(c Cluster) string {
	var heur []string
	for _, m := range c.Members {
		if m.HeuristicOnly {
			heur = append(heur, m.Path)
		}
	}
	sort.Strings(heur)
	return c.ID + "|" + c.Dir + "|" + c.Stem + "|" + join(c.Files) + "|" + join(c.Kinds) + "|" + join(heur)
}

func resolvedLines(r Result) []string {
	var resolved []string
	for _, in := range r.Inherited {
		old := ""
		if in.Rule == "rename" {
			old = in.From
		}
		rule := map[string]string{"rename": "c", "variant": "d"}[in.Rule]
		resolved = append(resolved, in.Path+"|"+rule+"|"+join(in.Areas)+"|"+old+"|false")
	}
	for _, pm := range r.PendingMoves {
		resolved = append(resolved, pm.New+"|c|[]|"+pm.Old+"|true")
	}
	sort.Strings(resolved)
	return resolved
}

func fixtureSummary(f fxResult) summary {
	// variant_unowned is a pilot diagnostic (an unowned sibling exists in the tree); Scout does not count it.
	counts := map[string]int{}
	for k, v := range f.Counts {
		if k != "variant_unowned" {
			counts[k] = v
		}
	}
	s := summary{Counts: counts, Decided: f.Decided, Capped: f.Capped}
	for _, c := range f.Clusters {
		heur := append([]string(nil), c.Heuristic...)
		sort.Strings(heur)
		s.Clusters = append(s.Clusters, c.ID+"|"+c.Dir+"|"+c.Stem+"|"+join(c.Files)+"|"+join(c.Kinds)+"|"+join(heur))
	}
	for _, c := range f.Inert {
		s.Inert = append(s.Inert, c.Dir+"|"+c.Stem+"|"+join(c.Files))
	}
	for _, x := range f.Resolved {
		old := ""
		if x.Old != nil {
			old = *x.Old
		}
		owners := x.Owners
		if owners == nil {
			owners = []string{}
		}
		pool := "false"
		if x.Pool {
			pool = "true"
		}
		s.Resolved = append(s.Resolved, x.Path+"|"+x.Rule+"|"+join(owners)+"|"+old+"|"+pool)
	}
	sort.Strings(s.Resolved)
	return s
}

func compareResult(t *testing.T, label string, got summary, want fxResult) {
	t.Helper()
	w := fixtureSummary(want)
	// The pilot also counts rules with no Scout equivalent or no Scout count; compare the shared names.
	if !reflect.DeepEqual(norm(got), norm(w)) {
		t.Errorf("%s:\n got  %+v\n want %+v", label, got, w)
	}
}

func norm(s summary) summary {
	if len(s.Counts) == 0 {
		s.Counts = nil
	}
	for _, p := range []*[]string{&s.Clusters, &s.Inert, &s.Decided, &s.Capped, &s.Resolved} {
		if len(*p) == 0 {
			*p = nil
		}
	}
	return s
}
