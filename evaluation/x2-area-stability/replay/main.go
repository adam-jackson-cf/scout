// Command replay folds recorded agent maps through areas.Reconcile for every grid threshold (X2 preregistration §3, §5).
//
// For each transition (prev map, next map) it builds the prior areas from the prev map's active agent features
// (Reconcile(nil, prev)), then reconciles the next map's active features against them at each threshold, using the same
// code Scout runs in production (a refresh reconciles Map(agent) output directly; internal/engine/refresh.go). Maps are read
// through clawpatch.LoadFeatures, so stale (`skipped`) records are excluded exactly as in Scout. Nothing here calls a model.
//
// Besides the one-step rows it writes (a) per-transition coverage diagnostics (descriptive, §3.5) into every row, (b) with -chain-out a
// label-free chained fold F1→R1→R2→M1→M2 with accumulating areas (§3.6), and (c) with -meta-out the provenance of the replayed code.
//
// Usage: go run ./evaluation/x2-area-stability/replay -maps <data>/maps -transitions <file> -grid 30,35,...,90 -out <jsonl>
//
//	[-chain-out <jsonl>] [-meta-out <json>] [-root <scout repo root>]
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"scout/internal/areas"
	"scout/internal/clawpatch"
	"scout/internal/domain"
)

type transition struct {
	Arm   string `json:"arm"`
	Repo  string `json:"repo"`
	ID    string `json:"id"`
	Class string `json:"class"`
	Sub   string `json:"sub"`
	Prev  string `json:"prev"`
	Next  string `json:"next"`
}

type match struct {
	AreaID        string  `json:"areaId"`
	PrevFeatureID string  `json:"prevFeatureId"`
	NextFeatureID string  `json:"nextFeatureId"`
	Jaccard       float64 `json:"jaccard"`
}

type row struct {
	Arm         string  `json:"arm"`
	Repo        string  `json:"repo"`
	Transition  string  `json:"transition"`
	Class       string  `json:"class"`
	Sub         string  `json:"sub"`
	ThresholdPc int     `json:"thresholdPct"`
	Threshold   float64 `json:"threshold"`
	Status      string  `json:"status"` // ok | unmeasured
	Reason      string  `json:"reason,omitempty"`
	PrevAreas   int     `json:"prevAreas"`
	NextAgent   int     `json:"nextAgentFeatures"`
	Matched     int     `json:"matched"`
	Created     int     `json:"created"`
	Retired     int     `json:"retired"`
	Kept        int     `json:"kept"`
	Matches     []match `json:"matches"`
	// Coverage diagnostics (descriptive; threshold independent): agent features with ≥1 owned file in each map, distinct owned files in each
	// map and in both, and prior areas with no file in common with any next agent feature (a survival ceiling no threshold can lift).
	PrevOwnedFiles   int `json:"prevOwnedFiles"`
	NextOwnedFiles   int `json:"nextOwnedFiles"`
	SharedOwnedFiles int `json:"sharedOwnedFiles"`
	ZeroOverlapPrior int `json:"zeroOverlapPriorAreas"`
}

type chainRow struct {
	Arm         string  `json:"arm"`
	Repo        string  `json:"repo"`
	ThresholdPc int     `json:"thresholdPct"`
	Threshold   float64 `json:"threshold"`
	Step        string  `json:"step"` // transition ID of the step: T3 F1→R1, T4 R1→R2, T5 R2→M1, T6 M1→M2
	PrevMap     string  `json:"prevMap"`
	NextMap     string  `json:"nextMap"`
	Status      string  `json:"status"`
	Reason      string  `json:"reason,omitempty"`
	ActivePrior int     `json:"activePrior"` // active areas entering the step (accumulated: kept areas persist, created areas join)
	Matched     int     `json:"matched"`
	Created     int     `json:"created"`
	Kept        int     `json:"kept"`
	Retired     int     `json:"retired"`
	F1Areas     int     `json:"f1Areas"`
	// F1MatchedAllSoFar: F1 areas matched at every step up to and including this one (the step-1..k survival of the original areas).
	F1MatchedAllSoFar int `json:"f1MatchedAllSoFar"`
}

type meta struct {
	ScoutHead         string `json:"scoutHead"`
	TreeDirty         bool   `json:"areasOrRecordsWorkingTreeDirty"`
	AreasGoSHA256     string `json:"areasGoSha256"`
	RecordsGoSHA256   string `json:"recordsGoSha256"`
	Grid              []int  `json:"gridPct"`
	GoVersion         string `json:"goVersion"`
	MapsDir           string `json:"mapsDir"`
	TransitionsSHA256 string `json:"transitionsSha256"`
}

func main() {
	maps := flag.String("maps", "", "directory holding <arm>/<repo>/<map id>/{meta.json,features/}")
	transFile := flag.String("transitions", "", "transitions JSON written by common.write_transitions")
	grid := flag.String("grid", "", "comma-separated thresholds in percent")
	out := flag.String("out", "", "output JSONL")
	chainOut := flag.String("chain-out", "", "optional chained-fold output JSONL")
	metaOut := flag.String("meta-out", "", "optional provenance JSON (git HEAD, sha256 of areas.go and records.go, grid, Go version)")
	root := flag.String("root", ".", "Scout repository root (provenance only)")
	flag.Parse()
	if *maps == "" || *transFile == "" || *grid == "" || *out == "" {
		flag.Usage()
		os.Exit(2)
	}
	var pcts []int
	for _, field := range strings.Split(*grid, ",") {
		p, err := strconv.Atoi(strings.TrimSpace(field))
		if err != nil || p <= 0 || p > 100 {
			fatal(fmt.Errorf("bad grid value %q", field))
		}
		pcts = append(pcts, p)
	}
	var transitions []transition
	data, err := os.ReadFile(*transFile)
	if err != nil {
		fatal(err)
	}
	if err := json.Unmarshal(data, &transitions); err != nil {
		fatal(err)
	}
	file, err := os.Create(*out)
	if err != nil {
		fatal(err)
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	defer w.Flush()
	enc := json.NewEncoder(w)

	for _, tr := range transitions {
		prev, prevReason := load(*maps, tr.Arm, tr.Repo, tr.Prev)
		next, nextReason := load(*maps, tr.Arm, tr.Repo, tr.Next)
		for _, pct := range pcts {
			r := row{Arm: tr.Arm, Repo: tr.Repo, Transition: tr.ID, Class: tr.Class, Sub: tr.Sub, ThresholdPc: pct, Threshold: float64(pct) / 100, Matches: []match{}}
			switch {
			case prev == nil:
				r.Status, r.Reason = "unmeasured", "prev map "+tr.Prev+": "+prevReason
			case next == nil:
				r.Status, r.Reason = "unmeasured", "next map "+tr.Next+": "+nextReason
			default:
				fillRow(&r, prev, next)
			}
			if err := enc.Encode(r); err != nil {
				fatal(err)
			}
		}
	}

	if *chainOut != "" {
		writeChains(*chainOut, *maps, transitions, pcts)
	}
	if *metaOut != "" {
		writeMeta(*metaOut, *root, *maps, data, pcts)
	}
}

// load returns the active features of one recorded map, or nil and a reason when the map was not recorded or its run failed.
func load(root, arm, repo, mapID string) ([]domain.Feature, string) {
	dir := filepath.Join(root, arm, repo, mapID)
	raw, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	if err != nil {
		return nil, "not recorded"
	}
	var m struct {
		RC      *int   `json:"rc"`
		Skipped string `json:"skipped"`
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, "meta.json unreadable"
	}
	if m.Skipped != "" { // a skipped mutation runs no map (rc null): the map and every transition touching it are unmeasured
		return nil, "mutation skipped: " + m.Skipped
	}
	if m.RC == nil {
		return nil, "meta.json unreadable"
	}
	if *m.RC != 0 {
		return nil, fmt.Sprintf("map run failed (rc=%d)", *m.RC)
	}
	features, err := clawpatch.LoadFeatures(dir)
	if err != nil {
		fatal(fmt.Errorf("%s: %w", dir, err))
	}
	return features, ""
}

func agentFiles(fs []domain.Feature) (n int, files map[string]struct{}) {
	files = map[string]struct{}{}
	for _, f := range fs {
		if !areas.IsAgent(f) || len(areas.FeatureFiles(f)) == 0 {
			continue
		}
		n++
		for _, p := range areas.FeatureFiles(f) {
			files[p] = struct{}{}
		}
	}
	return n, files
}

func fillRow(r *row, prev, next []domain.Feature) {
	none := func(string) bool { return false }
	prevAreas, _ := areas.Reconcile(nil, prev, none, areas.Config{Snapshot: "prev"})
	prevFeature := make(map[string]string, len(prevAreas))
	for _, a := range prevAreas {
		prevFeature[a.ScoutAreaID] = a.ClawpatchIDs[0]
	}
	_, report := areas.Reconcile(prevAreas, next, areas.HeuristicOwner(next), areas.Config{MatchJaccard: r.Threshold, Snapshot: "next"})
	r.Status = "ok"
	r.PrevAreas = len(prevAreas)
	var nextFiles map[string]struct{}
	r.NextAgent, nextFiles = agentFiles(next)
	_, prevFiles := agentFiles(prev)
	r.PrevOwnedFiles, r.NextOwnedFiles = len(prevFiles), len(nextFiles)
	for p := range prevFiles {
		if _, ok := nextFiles[p]; ok {
			r.SharedOwnedFiles++
		}
	}
	for _, a := range prevAreas {
		overlap := false
		for _, p := range a.FileSet {
			if _, ok := nextFiles[p]; ok {
				overlap = true
				break
			}
		}
		if !overlap {
			r.ZeroOverlapPrior++
		}
	}
	r.Matched, r.Created, r.Retired, r.Kept = len(report.Matched), len(report.Created), len(report.Retired), len(report.Kept)
	for _, m := range report.Matched {
		r.Matches = append(r.Matches, match{AreaID: m.ScoutAreaID, PrevFeatureID: prevFeature[m.ScoutAreaID], NextFeatureID: m.FeatureID, Jaccard: m.Jaccard})
	}
	sort.Slice(r.Matches, func(i, j int) bool {
		a, b := r.Matches[i], r.Matches[j]
		if a.PrevFeatureID != b.PrevFeatureID {
			return a.PrevFeatureID < b.PrevFeatureID
		}
		return a.NextFeatureID < b.NextFeatureID
	})
}

var chainIDs = []string{"T3", "T4", "T5", "T6"}

// writeChains folds F1→R1→R2→M1→M2 through Reconcile with accumulating areas at every grid threshold (label-free; §3.6).
func writeChains(path, mapsDir string, transitions []transition, pcts []int) {
	file, err := os.Create(path)
	if err != nil {
		fatal(err)
	}
	defer file.Close()
	w := bufio.NewWriter(file)
	defer w.Flush()
	enc := json.NewEncoder(w)
	type key struct{ arm, repo string }
	byKey := map[key]map[string]transition{}
	var order []key
	for _, tr := range transitions {
		k := key{tr.Arm, tr.Repo}
		if byKey[k] == nil {
			byKey[k] = map[string]transition{}
			order = append(order, k)
		}
		byKey[k][tr.ID] = tr
	}
	for _, k := range order {
		steps := make([]transition, 0, len(chainIDs))
		for _, id := range chainIDs {
			tr, ok := byKey[k][id]
			if !ok {
				steps = nil
				break
			}
			steps = append(steps, tr)
		}
		if steps == nil {
			continue
		}
		for _, pct := range pcts {
			foldChain(enc, mapsDir, k.arm, k.repo, steps, pct)
		}
	}
}

func foldChain(enc *json.Encoder, mapsDir, arm, repo string, steps []transition, pct int) {
	threshold := float64(pct) / 100
	mk := func(s transition) chainRow {
		return chainRow{Arm: arm, Repo: repo, ThresholdPc: pct, Threshold: threshold, Step: s.ID, PrevMap: s.Prev, NextMap: s.Next}
	}
	emit := func(r chainRow) {
		if err := enc.Encode(r); err != nil {
			fatal(err)
		}
	}
	start, reason := load(mapsDir, arm, repo, steps[0].Prev)
	broken := ""
	if start == nil {
		broken = "chain start map " + steps[0].Prev + ": " + reason
	}
	var current []areas.Area
	matchedAll := map[string]bool{}
	if broken == "" {
		current, _ = areas.Reconcile(nil, start, func(string) bool { return false }, areas.Config{Snapshot: "prev"})
		for _, a := range current {
			matchedAll[a.ScoutAreaID] = true
		}
	}
	for _, s := range steps {
		r := mk(s)
		r.F1Areas = len(matchedAll)
		if broken == "" {
			next, nextReason := load(mapsDir, arm, repo, s.Next)
			if next == nil {
				broken = "map " + s.Next + ": " + nextReason
			} else {
				for _, a := range current {
					if a.Active() {
						r.ActivePrior++
					}
				}
				var report areas.Report
				current, report = areas.Reconcile(current, next, areas.HeuristicOwner(next), areas.Config{MatchJaccard: threshold, Snapshot: s.Next})
				matchedNow := map[string]bool{}
				for _, m := range report.Matched {
					matchedNow[m.ScoutAreaID] = true
				}
				for id := range matchedAll {
					if !matchedNow[id] {
						matchedAll[id] = false
					}
				}
				for _, ok := range matchedAll {
					if ok {
						r.F1MatchedAllSoFar++
					}
				}
				r.Status = "ok"
				r.Matched, r.Created, r.Kept, r.Retired = len(report.Matched), len(report.Created), len(report.Kept), len(report.Retired)
			}
		}
		if broken != "" && r.Status != "ok" {
			r.Status, r.Reason = "unmeasured", broken
		}
		emit(r)
	}
}

func fileSHA(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		fatal(err)
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func writeMeta(path, root, mapsDir string, transJSON []byte, pcts []int) {
	git := func(args ...string) string {
		c := exec.Command("git", args...)
		c.Dir = root
		b, err := c.Output()
		if err != nil {
			return "unknown"
		}
		return strings.TrimSpace(string(b))
	}
	status := git("status", "--porcelain", "--", "internal/areas", "internal/clawpatch")
	th := sha256.Sum256(transJSON)
	m := meta{
		ScoutHead:         git("rev-parse", "HEAD"),
		TreeDirty:         status != "" && status != "unknown",
		AreasGoSHA256:     fileSHA(filepath.Join(root, "internal", "areas", "areas.go")),
		RecordsGoSHA256:   fileSHA(filepath.Join(root, "internal", "clawpatch", "records.go")),
		Grid:              pcts,
		GoVersion:         runtime.Version(),
		MapsDir:           mapsDir,
		TransitionsSHA256: hex.EncodeToString(th[:]),
	}
	b, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		fatal(err)
	}
	if err := os.WriteFile(path, append(b, '\n'), 0o644); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "replay:", err)
	os.Exit(1)
}
