package history

import (
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

// TestActiveHoursReproducesE1bPanel recomputes active_hours for every nest and
// webpack PR in panel-final.csv.gz from the PR commit author timestamps in
// raw/prs-<repo>.json.gz and requires agreement within 1e-9 h.
func TestActiveHoursReproducesE1bPanel(t *testing.T) {
	dir := filepath.Join("..", "..", "evaluation", "change-cost-v2")
	open := func(name string) io.ReadCloser {
		f, err := os.Open(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		z, err := gzip.NewReader(f)
		if err != nil {
			t.Fatal(err)
		}
		return struct {
			io.Reader
			io.Closer
		}{z, f}
	}

	type pr struct {
		Number  int `json:"number"`
		Commits []struct {
			Authored string `json:"authored"`
		} `json:"commits"`
	}
	raw := map[string]map[int][]float64{}
	for _, repo := range []string{"nest", "webpack"} {
		r := open("raw/prs-" + repo + ".json.gz")
		var prs []pr
		if err := json.NewDecoder(r).Decode(&prs); err != nil {
			t.Fatal(err)
		}
		r.Close()
		raw[repo] = map[int][]float64{}
		for _, p := range prs {
			var ts []float64
			for _, c := range p.Commits {
				at, err := time.Parse(time.RFC3339, c.Authored)
				if err != nil {
					t.Fatalf("%s #%d: %v", repo, p.Number, err)
				}
				ts = append(ts, float64(at.Unix()))
			}
			raw[repo][p.Number] = ts
		}
	}

	r := open("panel-final.csv.gz")
	defer r.Close()
	rows, err := csv.NewReader(r).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	col := map[string]int{}
	for i, h := range rows[0] {
		col[h] = i
	}
	compared, maxErr := 0, 0.0
	for _, row := range rows[1:] {
		repo := row[col["repo"]]
		if repo != "nest" && repo != "webpack" {
			continue
		}
		num, err := strconv.Atoi(row[col["number"]])
		if err != nil {
			t.Fatal(err)
		}
		want, err := strconv.ParseFloat(row[col["active_hours"]], 64)
		if err != nil {
			t.Fatalf("%s #%d active_hours: %v", repo, num, err)
		}
		ts, ok := raw[repo][num]
		if !ok {
			t.Errorf("panel PR %s #%d missing from raw PR file", repo, num)
			continue
		}
		e := math.Abs(ActiveHours(ts) - want)
		if e > 1e-9 {
			t.Errorf("%s #%d: got %v want %v", repo, num, ActiveHours(ts), want)
		}
		maxErr = math.Max(maxErr, e)
		compared++
	}
	if compared != 252+346 {
		t.Errorf("compared %d panel PRs, want 598 (nest 252 + webpack 346)", compared)
	}
	t.Logf("active_hours compared for %d panel PRs (nest+webpack); max abs error %.3g h", compared, maxErr)
}
