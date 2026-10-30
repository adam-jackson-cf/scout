// Command onboard builds the X2d catalogue from the round-0 onboarding draws (X2d PREREGISTRATION section 4.2), code first, Jev only for the files the draws could not
// settle. It imports Scout's own packages: internal/clawpatch (LoadFeatures), internal/areas (IsAgent, FeatureFiles, Jaccard) and internal/decisions (client, Question, version
// key, Redact). It is the study's onboarding driver and changes no product file.
//
//	onboard -draws D/onboard/draws -scopes prepass/scopes.json -out D/catalogue [-contrast D/onboard/contrast] [-tau 0.5] [-k 3] -consensus-only
//	onboard -draws ... -scopes ... -out D/catalogue -contrast D/onboard/contrast -no-jev [-repo CLONE -rev C0]
//	OPENROUTER_API_KEY=... onboard -draws ... -scopes ... -out D/catalogue -contrast D/onboard/contrast -repo CLONE -rev C0 -jev-out D/onboard-jev/scores.jsonl
//	    [-questions-src DIR] [-questions-dir DIR] [-thresholds FILE] [-endpoint URL] [-model M] [-workers 8] [-max-cost-usd 1] [-timeout 60s]
//
// Steps (PREREGISTRATION 4.2):
//  1. Per scope of -scopes, draws -draws/<scope>/<k> for k = 1..-k (default 3) are read with clawpatch.LoadFeatures; only agent features (areas.IsAgent) count and each feature's
//     file set is areas.FeatureFiles. A draw directory that is absent, or whose meta.json records rc != 0, is refused (rerun the collector; the quorum depends on K).
//  2. Code-first consensus per scope: a file is owned when it lies in >= ceil(K/2) of the scope's draws, contested when it lies in some draw but fewer, unowned when it lies in
//     none (needs -repo: the scope's files are the tracked files of -rev inside the Clawpatch inventory and the scope's include entries). Co-association of two owned files =
//     (draws in which some feature holds both) / K, so a file in several features of one draw contributes to every pair it forms; clusters are the connected components of the
//     pairs with co-association >= tau. A cluster takes title and summary from the draw feature with the highest Jaccard to it (ties: lowest draw, then feature id; selected,
//     never generated); sample_files are at most 8 files: round-robin over the cluster's top-level directories (sorted), one file per frontend.cluster_key (normalised directory,
//     variant stem; first in sorted order), topped up with the remaining sorted files when fewer than min(8, files) keys exist (list order = round-robin order, so a prefix of
//     n files is itself diverse); area_id = sha256("x2d-area" + scope + "\0".join(sorted members))[:12]. A file in several features of one draw contributes one pair vote per
//     draw; consensus-report.json lists the bridge files (articulation points of the co-association graph at tau: removing one splits its cluster).
//  3. Jev (skipped with -no-jev and -consensus-only; needs -repo, -rev, OPENROUTER_API_KEY, -jev-out): every contested or unowned file of a scope with 2..255 clusters goes to
//     onboard.covered (Noul, the scope catalogue in the state); when p_true >= t_onboard_cov (thresholds.json) to onboard.which_file_area (Choice over the scope catalogue,
//     permuted order, unit id "<scope>:<path>"); the file is assigned when the top option's probability >= t_onboard_choice, otherwise it stays unowned. The file under judgement is
//     removed from any sample_files it appears in. The state carries no owner facts. Binary files (NUL or invalid UTF-8) are not asked and stay unowned (reported).
//  4. Duplicate check (report only): onboard.same_responsibility for every pair of consensus clusters of one scope whose titles share a token (>= -dup-min-shared, default 1),
//     both presentation orders; a pair is reported as a duplicate when both orders give p_true >= -dup-threshold (0.7).
//  5. New areas: none. The unowned residue is reported by leaf directory.
//
// Outputs (refused when any exists): -out/catalogue.json, -out/consensus-report.json, -out/contrast-catalogue.json (-contrast given; one area per agent feature of the contrast
// draw), and last -out/catalogue.sha256 (the hex digest of catalogue.json, no newline). -consensus-only writes only catalogue.json and consensus-report.json and needs no repo, key or
// endpoint. The Jev output -jev-out is append-only and resumable (same conventions as jev/main.go: header, `answer` and `failure` rows, cumulative -max-cost-usd stop, header
// compatibility refusal); the catalogue is written only when every Jev request has an answer, otherwise the run exits 3 (failures) or 4 (cost cap) and a rerun resumes.
package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"scout/internal/areas"
	"scout/internal/clawpatch"
	"scout/internal/decisions"
	"scout/internal/policy"
)

const (
	defaultC0    = "1d50bad4555b951d5fa48eb8787fd5c39ed7effb"
	qOnCovered   = "onboard.covered"
	qOnWhich     = "onboard.which_file_area"
	qOnSame      = "onboard.same_responsibility"
	sampleCap    = 8
	excerptLines = 60
	siblingCap   = 10
)

// ---- helpers shared with jev/main.go (duplicated, identical; X2c does the same)

type variantDef struct {
	InstructionsDelta string   `json:"instructions_delta"`
	StateShapeDelta   []string `json:"state_shape_delta"`
}

type questionFile struct {
	QuestionID       string                    `json:"question_id"`
	Kind             string                    `json:"kind"`
	Instructions     string                    `json:"instructions"`
	Criteria         json.RawMessage           `json:"criteria"`
	CriteriaTemplate string                    `json:"criteria_template"`
	StateShape       []string                  `json:"state_shape"`
	OrderRule        decisions.OptionOrderRule `json:"option_order_rule"`
	Positive         *string                   `json:"positive"`
	Variants         map[string]variantDef     `json:"variants"`
}

func loadQuestionFile(dir, id string) (*questionFile, error) {
	raw, err := os.ReadFile(filepath.Join(dir, id+"@v1.json"))
	if err != nil {
		return nil, err
	}
	var q questionFile
	if err := json.Unmarshal(raw, &q); err != nil {
		return nil, fmt.Errorf("%s@v1.json: %w", id, err)
	}
	if q.QuestionID != id {
		return nil, fmt.Errorf("%s@v1.json names question %q", id, q.QuestionID)
	}
	return &q, nil
}

func redactValue(v any) any {
	switch t := v.(type) {
	case string:
		return decisions.Redact(t)
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = redactValue(e)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, e := range t {
			out[k] = redactValue(e)
		}
		return out
	default:
		return v
	}
}

func redactAll(in []string) []string {
	out := make([]string, len(in))
	for i, s := range in {
		out[i] = decisions.Redact(s)
	}
	return out
}

func toAny(in []string) []any {
	out := make([]any, len(in))
	for i, s := range in {
		out[i] = s
	}
	return out
}

// choiceComplete: the Choice contract returns a probability for every option and the probabilities sum to 1; the product client checks only the supplied values, so the
// driver requires every option and a sum within 1e-6 of 1 (PREREGISTRATION section 6).
func choiceComplete(probs map[string]float64, q decisions.Question) bool {
	sum := 0.0
	for _, p := range probs {
		sum += p
	}
	return len(probs) == len(q.Criteria) && math.Abs(sum-1) <= 1e-6
}

// confidenceOf reads the route's `confidence` of an answer (nil when it is absent).
func confidenceOf(raw json.RawMessage) *float64 {
	var a struct {
		Confidence *float64 `json:"confidence"`
	}
	if json.Unmarshal(raw, &a) != nil || a.Confidence == nil || math.IsNaN(*a.Confidence) || *a.Confidence < 0 || *a.Confidence > 1 {
		return nil
	}
	return a.Confidence
}

func die(code int, format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(code)
}

// ---- catalogue schema (README / contract)

type catArea struct {
	AreaID       string   `json:"area_id"`
	Scope        string   `json:"scope"`
	Title        string   `json:"title"`
	Summary      string   `json:"summary"`
	SampleFiles  []string `json:"sample_files"`
	Files        []string `json:"files"`
	ClawpatchIDs []string `json:"clawpatch_ids"`
}

type contestedRow struct {
	Path    string `json:"path"`
	Scope   string `json:"scope"`
	OwnedIn []int  `json:"owned_in"`
}

type unownedRow struct {
	Path  string `json:"path"`
	Scope string `json:"scope"`
}

type assignedRow struct {
	Path       string  `json:"path"`
	AreaID     string  `json:"area_id"`
	PCov       float64 `json:"p_cov"`
	Confidence float64 `json:"confidence"`
}

type catalogue struct {
	Schema        string         `json:"schema"`
	C0            string         `json:"c0"`
	Scopes        []string       `json:"scopes"`
	Tau           float64        `json:"tau"`
	Areas         []catArea      `json:"areas"`
	Contested     []contestedRow `json:"contested"`
	Unowned       []unownedRow   `json:"unowned"`
	JevAssigned   []assignedRow  `json:"jev_assigned"`
	HeuristicOnly []string       `json:"heuristic_only"`
}

// ---- reading the draws

type scopesFile struct {
	Scopes []struct {
		ID      string   `json:"id"`
		Include []string `json:"include"`
	} `json:"scopes"`
}

type feature struct {
	ID, Title, Summary string
	Files              []string
}

type draw struct {
	K     int
	Feats []feature // agent features with at least one owned file, sorted by id
	Heur  []string  // owned files of the non-agent (heuristic) features
}

func loadDraws(drawsDir, scope string, k int) ([]draw, error) {
	var out []draw
	for i := 1; i <= k; i++ {
		dir := filepath.Join(drawsDir, scope, strconv.Itoa(i))
		if st, err := os.Stat(dir); err != nil || !st.IsDir() {
			return nil, fmt.Errorf("scope %s: draw %d is missing (%s); the quorum needs K=%d draws", scope, i, dir, k)
		}
		if raw, err := os.ReadFile(filepath.Join(dir, "meta.json")); err == nil {
			var meta struct {
				RC *int `json:"rc"`
			}
			if err := json.Unmarshal(raw, &meta); err != nil {
				return nil, fmt.Errorf("%s/meta.json: %w", dir, err)
			}
			if meta.RC != nil && *meta.RC != 0 {
				return nil, fmt.Errorf("scope %s: draw %d failed (meta.json rc=%d); rerun the collection before consolidating", scope, i, *meta.RC)
			}
		}
		recs, err := clawpatch.LoadFeatures(dir)
		if err != nil {
			return nil, fmt.Errorf("scope %s draw %d: %w", scope, i, err)
		}
		d := draw{K: i}
		for _, f := range recs {
			if !areas.IsAgent(f) {
				d.Heur = append(d.Heur, areas.FeatureFiles(f)...)
				continue
			}
			files := areas.FeatureFiles(f)
			if len(files) == 0 {
				continue
			}
			d.Feats = append(d.Feats, feature{ID: f.FeatureID, Title: f.Title, Summary: f.Summary, Files: files})
		}
		sort.Slice(d.Feats, func(a, b int) bool { return d.Feats[a].ID < d.Feats[b].ID })
		out = append(out, d)
	}
	return out, nil
}

// ---- consensus

type cluster struct {
	AreaID      string
	Members     []string
	Title       string
	Summary     string
	TitleSource string
	Jaccard     float64
	IDs         []string
	Sample      []string // sample_files: sampleFilesOf(Members)
}

type scopeResult struct {
	Scope      string
	Draws      []draw
	Quorum     int
	Owned      []string
	Contested  []contestedRow
	Unowned    []string
	Clusters   []*cluster
	Universe   []string
	InAnyDraw  int
	Singletons int
	FeaturesBy map[int]int
	UniverseOK bool
	SkippedJev string
	Bridges    []bridgeRow
}

type unionFind []int

// bridgeRow: a bridge file of the co-association graph (an articulation point at tau: removing the file splits its cluster).
type bridgeRow struct {
	Path   string `json:"path"`
	Scope  string `json:"scope"`
	AreaID string `json:"area_id"`
}

// articulationPoints returns the sorted articulation points of the undirected graph on n nodes (Tarjan, one DFS per component).
func articulationPoints(n int, edges [][2]int) []int {
	adj := make([][]int, n)
	for _, e := range edges {
		adj[e[0]] = append(adj[e[0]], e[1])
		adj[e[1]] = append(adj[e[1]], e[0])
	}
	disc, low := make([]int, n), make([]int, n)
	ap := make([]bool, n)
	timer := 0
	var dfs func(u, parent int)
	dfs = func(u, parent int) {
		timer++
		disc[u], low[u] = timer, timer
		children := 0
		for _, v := range adj[u] {
			if disc[v] == 0 {
				children++
				dfs(v, u)
				if low[v] < low[u] {
					low[u] = low[v]
				}
				if parent != -1 && low[v] >= disc[u] {
					ap[u] = true
				}
			} else if v != parent && disc[v] < low[u] {
				low[u] = disc[v]
			}
		}
		if parent == -1 && children > 1 {
			ap[u] = true
		}
	}
	for i := 0; i < n; i++ {
		if disc[i] == 0 {
			dfs(i, -1)
		}
	}
	out := []int{}
	for i, b := range ap {
		if b {
			out = append(out, i)
		}
	}
	return out
}

func (u unionFind) find(x int) int {
	for u[x] != x {
		u[x] = u[u[x]]
		x = u[x]
	}
	return x
}

func (u unionFind) union(a, b int) {
	ra, rb := u.find(a), u.find(b)
	if ra == rb {
		return
	}
	if ra < rb {
		u[rb] = ra
	} else {
		u[ra] = rb
	}
}

func areaID(scope string, members []string) string {
	sum := sha256.Sum256([]byte("x2d-area" + scope + strings.Join(members, "\x00")))
	return hex.EncodeToString(sum[:])[:12]
}

func consensus(scope string, draws []draw, tau float64) *scopeResult {
	k := len(draws)
	quorum := (k + 1) / 2
	res := &scopeResult{Scope: scope, Draws: draws, Quorum: quorum, FeaturesBy: map[int]int{}}
	inDraws := map[string][]int{}
	for _, d := range draws {
		res.FeaturesBy[d.K] = len(d.Feats)
		seen := map[string]bool{}
		for _, f := range d.Feats {
			for _, p := range f.Files {
				seen[p] = true
			}
		}
		for p := range seen {
			inDraws[p] = append(inDraws[p], d.K)
		}
	}
	res.InAnyDraw = len(inDraws)
	for p, ks := range inDraws {
		sort.Ints(ks)
		if len(ks) >= quorum {
			res.Owned = append(res.Owned, p)
		} else {
			res.Contested = append(res.Contested, contestedRow{Path: p, Scope: scope, OwnedIn: ks})
		}
	}
	sort.Strings(res.Owned)
	sort.Slice(res.Contested, func(a, b int) bool { return res.Contested[a].Path < res.Contested[b].Path })
	index := make(map[string]int, len(res.Owned))
	for i, p := range res.Owned {
		index[p] = i
	}
	n := len(res.Owned)
	pairCount := map[uint64]int{}
	for _, d := range draws {
		seen := map[uint64]bool{}
		for _, f := range d.Feats {
			var ix []int
			for _, p := range f.Files {
				if i, ok := index[p]; ok {
					ix = append(ix, i)
				}
			}
			for a := 0; a < len(ix); a++ {
				for b := a + 1; b < len(ix); b++ {
					x, y := ix[a], ix[b]
					if x > y {
						x, y = y, x
					}
					seen[uint64(x)*uint64(n)+uint64(y)] = true
				}
			}
		}
		for key := range seen {
			pairCount[key]++
		}
	}
	uf := make(unionFind, n)
	for i := range uf {
		uf[i] = i
	}
	need := tau*float64(k) - 1e-9
	var edges [][2]int
	for key, c := range pairCount {
		if float64(c) >= need {
			x, y := int(key/uint64(n)), int(key%uint64(n))
			uf.union(x, y)
			edges = append(edges, [2]int{x, y})
		}
	}
	groups := map[int][]string{}
	for i, p := range res.Owned {
		r := uf.find(i)
		groups[r] = append(groups[r], p)
	}
	for _, members := range groups {
		sort.Strings(members)
		c := &cluster{AreaID: areaID(scope, members), Members: members, Sample: sampleFilesOf(members)}
		for _, d := range draws { // best feature per draw; the overall best (highest Jaccard, then lowest draw and id) donates title and summary
			var best *feature
			bj := 0.0
			for fi := range d.Feats {
				if j := areas.Jaccard(members, d.Feats[fi].Files); j > bj {
					best, bj = &d.Feats[fi], j
				}
			}
			if best != nil {
				c.IDs = append(c.IDs, fmt.Sprintf("%s/%d/%s", scope, d.K, best.ID))
				if bj > c.Jaccard {
					c.Jaccard, c.Title, c.Summary, c.TitleSource = bj, best.Title, best.Summary, fmt.Sprintf("%s/%d/%s", scope, d.K, best.ID)
				}
			}
		}
		sort.Strings(c.IDs)
		if len(members) == 1 {
			res.Singletons++
		}
		res.Clusters = append(res.Clusters, c)
	}
	sort.Slice(res.Clusters, func(a, b int) bool { return res.Clusters[a].AreaID < res.Clusters[b].AreaID })
	areaOf := map[string]string{}
	for _, c := range res.Clusters {
		for _, p := range c.Members {
			areaOf[p] = c.AreaID
		}
	}
	for _, i := range articulationPoints(n, edges) {
		res.Bridges = append(res.Bridges, bridgeRow{Path: res.Owned[i], Scope: scope, AreaID: areaOf[res.Owned[i]]})
	}
	sort.Slice(res.Bridges, func(a, b int) bool { return res.Bridges[a].Path < res.Bridges[b].Path })
	return res
}

// ---- the repository (inventory, roles, siblings, excerpts)

var skipRe = regexp.MustCompile(`(^|/)(node_modules|dist|build|coverage|\.build|\.git|\.clawpatch|\.worktrees|\.turbo|\.next|\.vercel|\.venv(?:-[^/]+)?|venv|Pods|Carthage|SourcePackages|DerivedData|__pycache__)(/|$)`)
var variantRe = regexp.MustCompile(`(?:_an)?(?:_py3\d+)?(?:_pv[12])?$`)

// inInventory mirrors prepass.inventory (Clawpatch repoInventory, agent-mapper.ts:363-400).
func inInventory(p string) bool {
	return p != "" && !skipRe.MatchString(p) && !strings.HasPrefix(p, "target/") && p != "target" && !strings.HasPrefix(p, ".build/")
}

func inScope(p string, include []string) bool {
	for _, e := range include {
		e = strings.TrimSuffix(e, "/")
		if p == e || strings.HasPrefix(p, e+"/") {
			return true
		}
	}
	return false
}

// normDir is frontend.norm_dir.
func normDir(d string) string {
	segs := strings.Split(d, "/")
	for i, s := range segs {
		if v := variantRe.ReplaceAllString(s, ""); v != "" {
			segs[i] = v
		}
	}
	return strings.Join(segs, "/")
}

// roleOf is frontend.role.
func roleOf(p string) string {
	switch {
	case strings.HasPrefix(p, "fastapi/"):
		return "core"
	case strings.HasPrefix(p, "docs_src/"):
		return "example"
	case strings.HasPrefix(p, "tests/"):
		return "test"
	case strings.HasPrefix(p, ".github/") || strings.HasPrefix(p, "scripts/"):
		return "infra"
	case strings.HasPrefix(p, "docs/"):
		return "docs-config"
	}
	return "config"
}

type repo struct {
	dir, rev string
	tree     []string
	byDir    map[string][]string
}

func openRepo(dir, rev string) (*repo, error) {
	out, err := exec.Command("git", "-C", dir, "ls-tree", "-r", "--name-only", "-z", rev).Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-tree %s: %w", rev, err)
	}
	r := &repo{dir: dir, rev: rev, byDir: map[string][]string{}}
	for _, p := range strings.Split(string(out), "\x00") {
		if p == "" {
			continue
		}
		r.tree = append(r.tree, p)
		nd := normDir(path.Dir(p))
		r.byDir[nd] = append(r.byDir[nd], p)
	}
	sort.Strings(r.tree)
	for _, v := range r.byDir {
		sort.Strings(v)
	}
	return r, nil
}

func (r *repo) universe(include []string) []string {
	var out []string
	for _, p := range r.tree {
		if inInventory(p) && inScope(p, include) {
			out = append(out, p)
		}
	}
	return out
}

func (r *repo) siblings(p string) []string {
	out := []string{}
	for _, q := range r.byDir[normDir(path.Dir(p))] {
		if q != p {
			out = append(out, q)
			if len(out) == siblingCap {
				break
			}
		}
	}
	return out
}

// excerpt reads the first lines of the file at the revision; ok is false for a binary file.
func (r *repo) excerpt(p string, lines int) (text string, ok bool, err error) {
	out, err := exec.Command("git", "-C", r.dir, "show", r.rev+":"+p).Output()
	if err != nil {
		return "", false, fmt.Errorf("git show %s:%s: %w", r.rev, p, err)
	}
	if bytes.IndexByte(out, 0) >= 0 || !utf8.Valid(out) {
		return "", false, nil
	}
	all := strings.Split(string(out), "\n")
	if len(all) > lines {
		all = all[:lines]
	}
	return strings.Join(all, "\n"), true, nil
}

// ---- Jev rows

type jevHeader struct {
	Type            string            `json:"type"`
	Model           string            `json:"model"`
	ConsensusSHA256 string            `json:"consensus_sha256"`
	Rev             string            `json:"rev"`
	TOnboardCov     float64           `json:"t_onboard_cov"`
	TOnboardChoice  float64           `json:"t_onboard_choice"`
	DupThreshold    float64           `json:"dup_threshold"`
	Versions        map[string]string `json:"versions"`
	Endpoint        string            `json:"endpoint"`
	Started         string            `json:"started_utc"`
	Tau             float64           `json:"tau"`
	ExcerptLines    int               `json:"excerpt_lines"`
}

// jevRow is an answer or failure row of the onboarding Jev output.
type jevRow struct {
	Type            string             `json:"type"`
	ReqID           string             `json:"req_id"`
	Question        string             `json:"question"`
	Scope           string             `json:"scope"`
	Path            string             `json:"path,omitempty"`
	AreaA           string             `json:"area_a,omitempty"`
	AreaB           string             `json:"area_b,omitempty"`
	Order           string             `json:"order,omitempty"`
	PTrue           *float64           `json:"p_true,omitempty"`
	Choice          string             `json:"choice,omitempty"`
	PChoice         *float64           `json:"p_choice,omitempty"`
	Confidence      *float64           `json:"confidence,omitempty"`
	Probabilities   map[string]float64 `json:"probabilities,omitempty"`
	Kind            string             `json:"kind,omitempty"`
	Status          int                `json:"status,omitempty"`
	Reason          string             `json:"reason,omitempty"`
	CostUSD         float64            `json:"cost_usd"`
	CostKnown       bool               `json:"cost_known"`
	Model           string             `json:"model"`
	QuestionVersion string             `json:"question_version,omitempty"`
	ResponseID      string             `json:"response_id,omitempty"`
	RequestSHA256   string             `json:"request_sha256,omitempty"`
	Usage           json.RawMessage    `json:"usage,omitempty"`
	LatencyMS       int64              `json:"latency_ms"`
	Attempts        int                `json:"attempts,omitempty"`
	Timestamp       string             `json:"timestamp"`
}

// existing reads a previous Jev output: its header, the answer rows by request id and the billed cost already spent (answer and failure rows that report a cost).
func existing(file string) (*jevHeader, map[string]jevRow, float64, error) {
	f, err := os.Open(file)
	if os.IsNotExist(err) {
		return nil, map[string]jevRow{}, 0, nil
	}
	if err != nil {
		return nil, nil, 0, err
	}
	defer f.Close()
	var hd *jevHeader
	answers := map[string]jevRow{}
	spent := 0.0
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 64<<20)
	for line := 1; sc.Scan(); line++ {
		var row jevRow
		if err := json.Unmarshal(sc.Bytes(), &row); err != nil {
			return nil, nil, 0, fmt.Errorf("%s:%d: malformed line (a killed run left a partial line: truncate it)", file, line)
		}
		switch row.Type {
		case "header":
			var h jevHeader
			if err := json.Unmarshal(sc.Bytes(), &h); err != nil {
				return nil, nil, 0, err
			}
			hd = &h
		case "answer":
			answers[row.ReqID] = row
			if row.CostKnown {
				spent += row.CostUSD
			}
		case "failure":
			if row.CostKnown {
				spent += row.CostUSD
			}
		}
	}
	return hd, answers, spent, sc.Err()
}

type runner struct {
	client      *decisions.Client
	model       string
	ctx         context.Context
	cancel      context.CancelFunc
	mu          sync.Mutex
	enc         *json.Encoder
	answers     map[string]jevRow
	cost        float64
	maxCost     float64
	failures    int
	unknownCost int
	answered    int
	stopped     bool
	skipped     map[string]string // path -> reason a file was not asked
}

func (r *runner) write(v any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.enc.Encode(v); err != nil {
		fmt.Fprintln(os.Stderr, "write failed:", err)
		os.Exit(2)
	}
}

func (r *runner) get(id string) (jevRow, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	row, ok := r.answers[id]
	return row, ok
}

// ask sends one question; it appends an answer or failure row and returns the answer row when there is one.
func (r *runner) ask(q decisions.Question, version, unitID string, state json.RawMessage, base jevRow) (jevRow, bool) {
	reply, err := r.client.Ask(r.ctx, []decisions.Question{q}, r.model, unitID, state)
	now := time.Now().UTC().Format(time.RFC3339)
	r.mu.Lock()
	if reply.CostKnown {
		r.cost += reply.CostUSD
	} else if reply.Model != "" || err == nil {
		r.unknownCost++
	}
	over := r.cost > r.maxCost
	r.mu.Unlock()
	base.Model, base.CostUSD, base.CostKnown, base.LatencyMS, base.Timestamp = reply.Model, reply.CostUSD, reply.CostKnown, reply.Latency.Milliseconds(), now
	fail := func(kind string, status int, reason string) {
		base.Type, base.Kind, base.Status, base.Reason = "failure", kind, status, reason
		r.write(base)
		r.mu.Lock()
		r.failures++
		r.mu.Unlock()
	}
	var out jevRow
	ok := false
	switch {
	case err != nil:
		var fe *decisions.Failure
		if errors.As(err, &fe) {
			fail(string(fe.Kind), fe.Status, fe.Reason)
		} else {
			fail("other", 0, err.Error())
		}
	case len(reply.Answers) != 1:
		fail("schema", 0, "reply does not carry exactly one answer")
	case q.Kind == decisions.KindChoice && !choiceComplete(reply.Answers[0].Probabilities, q):
		fail("schema", 0, "choice probabilities incomplete or not normalised")
	case q.Kind == decisions.KindNoul && (reply.Answers[0].Kind != decisions.KindNoul || math.IsNaN(reply.Answers[0].Probability)):
		fail("schema", 0, "noul answer carries no probability")
	default:
		a := reply.Answers[0]
		sum := sha256.Sum256(reply.RequestBody)
		base.Type, base.QuestionVersion, base.ResponseID, base.Usage, base.Attempts = "answer", version, reply.ID, reply.Usage, reply.Attempts
		base.RequestSHA256, base.Confidence = hex.EncodeToString(sum[:]), confidenceOf(a.Answer)
		if a.Kind == decisions.KindNoul {
			p := a.Probability
			base.PTrue = &p
		} else {
			p := a.Probabilities[a.Choice]
			base.Choice, base.Probabilities, base.PChoice = a.Choice, a.Probabilities, &p
		}
		r.write(base)
		r.mu.Lock()
		r.answers[base.ReqID] = base
		r.answered++
		r.mu.Unlock()
		out, ok = base, true
	}
	if over {
		r.mu.Lock()
		r.stopped = true
		r.mu.Unlock()
		r.cancel()
	}
	return out, ok
}

// ---- questions

type onboardQuestions struct {
	src       string
	model     string
	covered   decisions.Question
	same      decisions.Question
	whichFile *questionFile
	cache     map[string]decisions.Question
	mu        sync.Mutex
}

func noulQuestion(f *questionFile) (decisions.Question, error) {
	var crit map[string]string
	if err := json.Unmarshal(f.Criteria, &crit); err != nil {
		return decisions.Question{}, fmt.Errorf("%s: criteria is not a map: %w", f.QuestionID, err)
	}
	q := decisions.Question{ID: f.QuestionID, Kind: decisions.KindNoul, Instructions: f.Instructions, Criteria: crit, StateShape: append([]string(nil), f.StateShape...), Order: f.OrderRule}
	if f.Positive != nil {
		q.Positive = *f.Positive
	}
	return q, q.Validate()
}

func loadOnboardQuestions(src, model string) (*onboardQuestions, error) {
	oq := &onboardQuestions{src: src, model: model, cache: map[string]decisions.Question{}}
	for _, id := range []string{qOnCovered, qOnSame} {
		f, err := loadQuestionFile(src, id)
		if err != nil {
			return nil, err
		}
		q, err := noulQuestion(f)
		if err != nil {
			return nil, err
		}
		if id == qOnCovered {
			oq.covered = q
		} else {
			oq.same = q
		}
	}
	f, err := loadQuestionFile(src, qOnWhich)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(f.CriteriaTemplate, "{title}: {summary} Sample files: ") || f.Kind != "choice" {
		return nil, fmt.Errorf("%s: criteria_template %q or kind %q is not the frozen one", qOnWhich, f.CriteriaTemplate, f.Kind)
	}
	oq.whichFile = f
	return oq, nil
}

// which builds the Choice over one scope's clusters; `exclude` (a file under judgement) is removed from every sample_files.
func (oq *onboardQuestions) which(scope string, clusters []*cluster, exclude string) (decisions.Question, error) {
	q := decisions.Question{ID: qOnWhich, Kind: decisions.KindChoice, Instructions: oq.whichFile.Instructions, StateShape: append([]string(nil), oq.whichFile.StateShape...),
		Criteria: map[string]string{}, Order: decisions.OptionOrderRule{Rule: decisions.OrderPermutation}}
	for _, c := range clusters {
		q.Criteria[c.AreaID] = decisions.Redact(fmt.Sprintf("%s: %s Sample files: %s", c.Title, c.Summary, strings.Join(sampleOf(c, exclude), ", ")))
	}
	return q, q.Validate()
}

// sampleOf: the cluster's sample_files without `exclude` (the file under judgement); the list is not refilled.
func sampleOf(c *cluster, exclude string) []string {
	out := []string{}
	for _, p := range c.Sample {
		if p != exclude {
			out = append(out, p)
		}
	}
	return out
}

var numRe = regexp.MustCompile(`_\d{1,2}$`)

// clusterKey is frontend.cluster_key: (normalised directory, variant stem).
func clusterKey(p string) string {
	d, f := "", p
	if i := strings.LastIndexByte(p, '/'); i >= 0 {
		d, f = p[:i], p[i+1:]
	}
	stem := f
	if j := strings.LastIndexByte(f, '.'); j >= 0 {
		stem = f[:j]
	}
	s := stem
	for {
		s2 := numRe.ReplaceAllString(variantRe.ReplaceAllString(s, ""), "")
		if s2 == s || s2 == "" {
			break
		}
		s = s2
	}
	return normDir(d) + "\x00" + s
}

// sampleFilesOf picks at most sampleCap sample files from `files` (sorted): round-robin over the top-level directories (sorted), one file per cluster key within the
// directory order; when fewer than min(sampleCap, len(files)) keys exist the remaining sorted files top the list up in the same round-robin order.
func sampleFilesOf(files []string) []string {
	want := sampleCap
	if len(files) < want {
		want = len(files)
	}
	byTop := map[string][]string{}
	for _, p := range files {
		top := ""
		if i := strings.IndexByte(p, '/'); i >= 0 {
			top = p[:i]
		}
		byTop[top] = append(byTop[top], p)
	}
	tops := make([]string, 0, len(byTop))
	for t := range byTop {
		tops = append(tops, t)
	}
	sort.Strings(tops)
	chosen := map[string]bool{}
	out := []string{}
	roundRobin := func(cands map[string][]string) {
		idx := map[string]int{}
		for len(out) < want {
			progressed := false
			for _, t := range tops {
				if len(out) >= want {
					break
				}
				if idx[t] < len(cands[t]) {
					p := cands[t][idx[t]]
					idx[t]++
					progressed = true
					out = append(out, p)
					chosen[p] = true
				}
			}
			if !progressed {
				return
			}
		}
	}
	seen := map[string]bool{}
	distinct := map[string][]string{}
	for _, t := range tops {
		for _, p := range byTop[t] {
			if k := clusterKey(p); !seen[k] {
				seen[k] = true
				distinct[t] = append(distinct[t], p)
			}
		}
	}
	roundRobin(distinct)
	if len(out) < want {
		rest := map[string][]string{}
		for _, t := range tops {
			for _, p := range byTop[t] {
				if !chosen[p] {
					rest[t] = append(rest[t], p)
				}
			}
		}
		roundRobin(rest)
	}
	return out
}

func catalogueState(clusters []*cluster, exclude string) []any {
	out := make([]any, 0, len(clusters))
	for _, c := range clusters {
		out = append(out, map[string]any{"id": c.AreaID, "title": c.Title, "summary": c.Summary, "sample_files": toAny(sampleOf(c, exclude))})
	}
	return out
}

func marshalState(v map[string]any) (json.RawMessage, error) {
	red := redactValue(v)
	return json.Marshal(red)
}

// ---- thresholds

func loadThresholds(file string) (tCov, tChoice float64, err error) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return 0, 0, err
	}
	var t struct {
		Decisions []struct {
			QuestionID string   `json:"question_id"`
			Threshold  *float64 `json:"threshold"`
		} `json:"decisions"`
	}
	if err := json.Unmarshal(raw, &t); err != nil {
		return 0, 0, fmt.Errorf("%s: %w", file, err)
	}
	find := func(id string) (float64, error) {
		for _, d := range t.Decisions {
			if d.QuestionID == id && d.Threshold != nil {
				return *d.Threshold, nil
			}
		}
		return 0, fmt.Errorf("%s: no threshold for %s", file, id)
	}
	if tCov, err = find(qOnCovered); err != nil {
		return
	}
	tChoice, err = find(qOnWhich)
	return
}

// ---- duplicate-check tokens

var tokenRe = regexp.MustCompile(`[a-z0-9]+`)
var stopTokens = map[string]bool{"the": true, "and": true, "for": true, "with": true, "from": true, "into": true, "that": true, "this": true, "its": true, "are": true}

func titleTokens(title string) map[string]bool {
	out := map[string]bool{}
	for _, t := range tokenRe.FindAllString(strings.ToLower(title), -1) {
		if len(t) >= 3 && !stopTokens[t] {
			out[t] = true
		}
	}
	return out
}

type dupPair struct {
	Scope   string   `json:"scope"`
	AreaA   string   `json:"area_a"`
	AreaB   string   `json:"area_b"`
	TitleA  string   `json:"title_a"`
	TitleB  string   `json:"title_b"`
	Shared  []string `json:"shared_tokens"`
	PAB     *float64 `json:"p_true_ab"`
	PBA     *float64 `json:"p_true_ba"`
	Duplica bool     `json:"duplicate"`
}

// ---- outputs

func writeNew(file string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func marshalIndent(v any) []byte {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", " ")
	if err := enc.Encode(v); err != nil {
		panic(err)
	}
	return b.Bytes()
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func main() {
	drawsDir := flag.String("draws", "evaluation/x2d-merge-association-pilot/data/onboard/draws", "directory of the onboarding draws: <scope>/<k>/{features/,meta.json}")
	scopesPath := flag.String("scopes", "evaluation/x2d-merge-association-pilot/prepass/scopes.json", "scopes file (ids and include entries)")
	repoDir := flag.String("repo", "", "clone containing -rev (needed for the unowned set and for Jev)")
	rev := flag.String("rev", defaultC0, "revision of the onboarding (C0)")
	outDir := flag.String("out", "evaluation/x2d-merge-association-pilot/data/catalogue", "catalogue directory")
	contrastDir := flag.String("contrast", "", "contrast draw directory (features/, meta.json); empty skips contrast-catalogue.json")
	consensusOnly := flag.Bool("consensus-only", false, "steps 1-2 only: write catalogue.json and consensus-report.json; no repo, key or endpoint")
	noJev := flag.Bool("no-jev", false, "skip the Jev steps (3 and 4)")
	jevOut := flag.String("jev-out", "", "onboarding Jev output JSONL (append-only; resumable)")
	qsrc := flag.String("questions-src", "evaluation/x2d-merge-association-pilot/questions", "directory of the frozen questions/*@v1.json")
	qdir := flag.String("questions-dir", "", "write each question as sent to <dir>/<id>@<version>.json")
	thresholdsPath := flag.String("thresholds", "evaluation/x2d-merge-association-pilot/thresholds.json", "thresholds.json (t_onboard_cov, t_onboard_choice)")
	endpoint := flag.String("endpoint", decisions.DefaultEndpoint, "decisions route URL (validation points it at a local stub)")
	model := flag.String("model", policy.DefaultModel, "pinned Jev model id (aliases and undated release lines are refused)")
	workers := flag.Int("workers", 8, "concurrent requests")
	maxCost := flag.Float64("max-cost-usd", 1, "stop when the cumulative billed cost of the Jev output exceeds this many USD")
	timeout := flag.Duration("timeout", 60*time.Second, "per-request HTTP timeout")
	tau := flag.Float64("tau", 0.5, "co-association threshold")
	kDraws := flag.Int("k", 3, "draws per scope (K)")
	dupThreshold := flag.Float64("dup-threshold", 0.7, "duplicate flag: p_true in both presentation orders")
	dupShared := flag.Int("dup-min-shared", 1, "title tokens two clusters must share to be checked for duplication")
	flag.Parse()

	if *kDraws < 1 || *tau <= 0 || *tau > 1 {
		die(2, "need -k >= 1 and 0 < -tau <= 1")
	}
	outputs := []string{filepath.Join(*outDir, "catalogue.json"), filepath.Join(*outDir, "consensus-report.json")}
	if !*consensusOnly {
		outputs = append(outputs, filepath.Join(*outDir, "contrast-catalogue.json"), filepath.Join(*outDir, "catalogue.sha256"))
	}
	for _, o := range outputs {
		if _, err := os.Stat(o); err == nil {
			die(2, "refusing: output exists (append-only): %s", o)
		}
	}
	doJev := !*consensusOnly && !*noJev
	if doJev {
		if *repoDir == "" || *jevOut == "" {
			die(2, "the Jev steps need -repo, -rev and -jev-out (or use -no-jev / -consensus-only)")
		}
		if policy.IsModelAlias(*model) {
			die(2, "refusing: %q is an alias or an undated release line, not a pinned build", *model)
		}
	}

	rawScopes, err := os.ReadFile(*scopesPath)
	if err != nil {
		die(2, "%v", err)
	}
	var sf scopesFile
	if err := json.Unmarshal(rawScopes, &sf); err != nil || len(sf.Scopes) == 0 {
		die(2, "%s: no scopes (%v)", *scopesPath, err)
	}
	var rp *repo
	if *repoDir != "" {
		if rp, err = openRepo(*repoDir, *rev); err != nil {
			die(2, "%v", err)
		}
	}

	// steps 1-2
	var results []*scopeResult
	scopeIDs := []string{}
	include := map[string][]string{}
	for _, s := range sf.Scopes {
		draws, err := loadDraws(*drawsDir, s.ID, *kDraws)
		if err != nil {
			die(2, "%v", err)
		}
		res := consensus(s.ID, draws, *tau)
		include[s.ID] = s.Include
		if rp != nil {
			res.UniverseOK = true
			res.Universe = rp.universe(s.Include)
			inDraw := map[string]bool{}
			for _, p := range res.Owned {
				inDraw[p] = true
			}
			for _, c := range res.Contested {
				inDraw[c.Path] = true
			}
			for _, p := range res.Universe {
				if !inDraw[p] {
					res.Unowned = append(res.Unowned, p)
				}
			}
		}
		results = append(results, res)
		scopeIDs = append(scopeIDs, s.ID)
	}

	cat := catalogue{Schema: "x2d-catalogue-v1", C0: *rev, Scopes: scopeIDs, Tau: *tau, Areas: []catArea{}, Contested: []contestedRow{}, Unowned: []unownedRow{}, JevAssigned: []assignedRow{}, HeuristicOnly: []string{}}
	areaIndex := map[string]int{}
	for _, res := range results {
		for _, c := range res.Clusters {
			if _, dup := areaIndex[c.AreaID]; dup {
				die(2, "area id collision %s (two scopes with the same members?)", c.AreaID)
			}
			cat.Areas = append(cat.Areas, catArea{AreaID: c.AreaID, Scope: res.Scope, Title: c.Title, Summary: c.Summary, SampleFiles: sampleOf(c, ""),
				Files: append([]string{}, c.Members...), ClawpatchIDs: c.IDs})
			areaIndex[c.AreaID] = 0
		}
		cat.Contested = append(cat.Contested, res.Contested...)
		for _, p := range res.Unowned {
			cat.Unowned = append(cat.Unowned, unownedRow{Path: p, Scope: res.Scope})
		}
	}
	sort.Slice(cat.Areas, func(a, b int) bool { return cat.Areas[a].AreaID < cat.Areas[b].AreaID })
	for i, a := range cat.Areas {
		areaIndex[a.AreaID] = i
	}
	clustersOf := map[string]*scopeResult{}
	for _, res := range results {
		clustersOf[res.Scope] = res
	}

	// steps 3-4
	var dups []dupPair
	jevSummary := map[string]any{"ran": false}
	if doJev {
		dups, jevSummary = runJev(results, &cat, areaIndex, rp, *rev, *jevOut, *qsrc, *qdir, *thresholdsPath, *endpoint, *model, *workers, *maxCost, *timeout, *tau, *dupThreshold, *dupShared, clustersOf)
	}

	// step 5 and the report
	assigned := map[string]bool{}
	for _, a := range cat.JevAssigned {
		assigned[a.Path] = true
	}
	residue := map[string]int{}
	for _, c := range cat.Contested {
		if !assigned[c.Path] {
			residue[path.Dir(c.Path)]++
		}
	}
	for _, u := range cat.Unowned {
		if !assigned[u.Path] {
			residue[path.Dir(u.Path)]++
		}
	}
	scopeReports := []map[string]any{}
	clusterReports := []map[string]any{}
	for _, res := range results {
		sr := map[string]any{"scope": res.Scope, "k": len(res.Draws), "quorum": res.Quorum, "features_per_draw": res.FeaturesBy, "files_in_any_draw": res.InAnyDraw,
			"owned": len(res.Owned), "contested": len(res.Contested), "unowned": len(res.Unowned), "unowned_universe_known": res.UniverseOK, "clusters": len(res.Clusters),
			"singleton_clusters": res.Singletons}
		if res.UniverseOK {
			sr["universe_files"] = len(res.Universe)
		}
		scopeReports = append(scopeReports, sr)
		for _, c := range res.Clusters {
			clusterReports = append(clusterReports, map[string]any{"area_id": c.AreaID, "scope": res.Scope, "n_files": len(c.Members), "files": c.Members, "title": c.Title, "summary": c.Summary,
				"title_source": c.TitleSource, "title_source_jaccard": c.Jaccard, "clawpatch_ids": c.IDs})
		}
	}
	sort.Slice(clusterReports, func(a, b int) bool {
		return clusterReports[a]["area_id"].(string) < clusterReports[b]["area_id"].(string)
	})
	if dups == nil {
		dups = []dupPair{}
	}
	bridges := []bridgeRow{}
	for _, res := range results {
		bridges = append(bridges, res.Bridges...)
	}
	report := map[string]any{"schema": "x2d-consensus-report-v1", "c0": *rev, "tau": *tau, "k": *kDraws, "unowned_universe_known": rp != nil, "scopes": scopeReports, "clusters": clusterReports,
		"bridge_files": bridges, "contested": cat.Contested, "unowned": cat.Unowned, "jev_assigned": cat.JevAssigned, "residue_by_leaf_dir": residue, "duplicate_check": dups, "jev": jevSummary}

	// contrast catalogue (validated before anything is written)
	var contrast *catalogue
	if *contrastDir != "" && !*consensusOnly {
		if contrast, err = contrastCatalogue(*contrastDir, *rev); err != nil {
			die(2, "%v", err)
		}
	}
	inArea := map[string]bool{}
	for _, a := range cat.Areas {
		for _, p := range a.Files {
			inArea[p] = true
		}
	}
	heurSet := map[string]bool{}
	for _, res := range results {
		for _, d := range res.Draws {
			for _, p := range d.Heur {
				if inScope(p, include[res.Scope]) && !inArea[p] {
					heurSet[p] = true
				}
			}
		}
	}
	for p := range heurSet {
		cat.HeuristicOnly = append(cat.HeuristicOnly, p)
	}
	sort.Strings(cat.HeuristicOnly)
	catBytes := marshalIndent(cat)
	if err := writeNew(outputs[0], catBytes); err != nil {
		die(2, "%v", err)
	}
	if err := writeNew(outputs[1], marshalIndent(report)); err != nil {
		die(2, "%v", err)
	}
	if *consensusOnly {
		fmt.Printf("onboard: consensus only: %d areas, %d contested, %d unowned (catalogue.json, consensus-report.json)\n", len(cat.Areas), len(cat.Contested), len(cat.Unowned))
		return
	}
	if contrast != nil {
		if err := writeNew(outputs[2], marshalIndent(contrast)); err != nil {
			die(2, "%v", err)
		}
	}
	if err := writeNew(outputs[3], []byte(sha256Hex(catBytes))); err != nil {
		die(2, "%v", err)
	}
	fmt.Printf("onboard: %d areas in %d scopes, %d contested, %d unowned, %d assigned by Jev; catalogue sha256 %s\n", len(cat.Areas), len(cat.Scopes), len(cat.Contested), len(cat.Unowned),
		len(cat.JevAssigned), sha256Hex(catBytes))
}

// contrastCatalogue: one area per agent feature of the contrast draw (area_id = sha256("x2d-contrast" + feature id)[:12], no consensus).
func contrastCatalogue(dir, rev string) (*catalogue, error) {
	if raw, err := os.ReadFile(filepath.Join(dir, "meta.json")); err == nil {
		var meta struct {
			RC *int `json:"rc"`
		}
		if err := json.Unmarshal(raw, &meta); err != nil {
			return nil, fmt.Errorf("%s/meta.json: %w", dir, err)
		}
		if meta.RC != nil && *meta.RC != 0 {
			return nil, fmt.Errorf("the contrast draw failed (meta.json rc=%d); rerun the collection or omit -contrast", *meta.RC)
		}
	}
	recs, err := clawpatch.LoadFeatures(dir)
	if err != nil {
		return nil, fmt.Errorf("contrast draw: %w", err)
	}
	c := &catalogue{Schema: "x2d-catalogue-v1", C0: rev, Scopes: []string{"contrast"}, Tau: 0, Areas: []catArea{}, Contested: []contestedRow{}, Unowned: []unownedRow{}, JevAssigned: []assignedRow{}, HeuristicOnly: []string{}}
	for _, f := range recs {
		if !areas.IsAgent(f) {
			continue
		}
		files := areas.FeatureFiles(f)
		if len(files) == 0 {
			continue
		}
		sorted := append([]string{}, files...)
		sort.Strings(sorted)
		sample := sampleFilesOf(sorted)
		c.Areas = append(c.Areas, catArea{AreaID: sha256Hex([]byte("x2d-contrast" + f.FeatureID))[:12], Scope: "contrast", Title: f.Title, Summary: f.Summary,
			SampleFiles: append([]string{}, sample...), Files: files, ClawpatchIDs: []string{"contrast/" + f.FeatureID}})
	}
	if len(c.Areas) == 0 {
		return nil, errors.New("the contrast draw has no agent feature with files")
	}
	sort.Slice(c.Areas, func(a, b int) bool { return c.Areas[a].AreaID < c.Areas[b].AreaID })
	return c, nil
}

type fileTask struct {
	res   *scopeResult
	path  string
	scope string
}

// runJev performs steps 3 and 4. It returns the duplicate-check rows and a summary for the report, and updates cat (areas' files, jev_assigned). It exits 3/4 (writing nothing
// else) when a request has no answer.
func runJev(results []*scopeResult, cat *catalogue, areaIndex map[string]int, rp *repo, rev, jevOut, qsrc, qdir, thresholdsPath, endpoint, model string, workers int, maxCost float64,
	timeout time.Duration, tau, dupThreshold float64, dupShared int, clustersOf map[string]*scopeResult) ([]dupPair, map[string]any) {
	key := os.Getenv("OPENROUTER_API_KEY")
	if key == "" {
		die(2, "OPENROUTER_API_KEY is not set (run through `infisical run ... --`; the key is never printed)")
	}
	if rp == nil {
		die(2, "the Jev steps need -repo")
	}
	tCov, tChoice, err := loadThresholds(thresholdsPath)
	if err != nil {
		die(2, "%v", err)
	}
	oq, err := loadOnboardQuestions(qsrc, model)
	if err != nil {
		die(2, "question invalid: %v", err)
	}
	versions := map[string]string{qOnCovered: oq.covered.Version(model), qOnSame: oq.same.Version(model)}
	whichFor := map[string]decisions.Question{}
	type questionOut struct {
		q       decisions.Question
		version string
	}
	sent := []questionOut{{oq.covered, versions[qOnCovered]}, {oq.same, versions[qOnSame]}}
	for _, res := range results {
		if len(res.Clusters) < 2 || len(res.Clusters) > 255 {
			res.SkippedJev = fmt.Sprintf("%d clusters: a Choice needs 2 to 255 options", len(res.Clusters))
			continue
		}
		q, err := oq.which(res.Scope, res.Clusters, "")
		if err != nil {
			die(2, "question invalid: %v", err)
		}
		whichFor[res.Scope] = q
		versions[qOnWhich+"/"+res.Scope] = q.Version(model)
		sent = append(sent, questionOut{q, versions[qOnWhich+"/"+res.Scope]})
	}
	consensusDigest := sha256Hex(marshalIndent(cat))
	want := jevHeader{Type: "header", Model: model, ConsensusSHA256: consensusDigest, Rev: rev, TOnboardCov: tCov, TOnboardChoice: tChoice, DupThreshold: dupThreshold, Versions: versions,
		Endpoint: endpoint, Started: time.Now().UTC().Format(time.RFC3339), Tau: tau, ExcerptLines: excerptLines}
	prev, prior, spent, err := existing(jevOut)
	if err != nil {
		die(2, "%v", err)
	}
	if prev != nil {
		same := prev.Model == want.Model && prev.ConsensusSHA256 == want.ConsensusSHA256 && prev.Rev == want.Rev && prev.TOnboardCov == want.TOnboardCov &&
			prev.TOnboardChoice == want.TOnboardChoice && prev.DupThreshold == want.DupThreshold && len(prev.Versions) == len(want.Versions) && prev.ExcerptLines == want.ExcerptLines
		for k, v := range want.Versions {
			if prev.Versions[k] != v {
				same = false
			}
		}
		if !same {
			die(2, "refusing to resume: the output header names another consensus, revision, model, threshold or question version")
		}
	}
	if prev == nil && len(prior) > 0 {
		die(2, "refusing: output has answer rows but no header")
	}
	if spent > maxCost {
		die(2, "refusing: budget already exhausted (billed cost %.6f USD in the output exceeds -max-cost-usd %.4f)", spent, maxCost)
	}
	if err := os.MkdirAll(filepath.Dir(jevOut), 0o755); err != nil {
		die(2, "%v", err)
	}
	f, err := os.OpenFile(jevOut, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		die(2, "%v", err)
	}
	defer f.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	r := &runner{client: &decisions.Client{Endpoint: endpoint, Key: key, HTTP: &http.Client{Timeout: timeout}}, model: model, ctx: ctx, cancel: cancel, enc: json.NewEncoder(f),
		answers: prior, cost: spent, maxCost: maxCost, skipped: map[string]string{}}
	if prev == nil {
		r.write(want)
	}
	if qdir != "" {
		if err := os.MkdirAll(qdir, 0o755); err != nil {
			die(2, "%v", err)
		}
		for _, s := range sent {
			raw, _ := json.MarshalIndent(map[string]any{"question_id": s.q.ID, "version": s.version, "model": model, "kind": s.q.Kind, "instructions": s.q.Instructions, "criteria": s.q.Criteria,
				"state_shape": s.q.StateShape, "option_order_rule": s.q.Order, "positive": s.q.Positive}, "", " ")
			if err := os.WriteFile(filepath.Join(qdir, s.version+".json"), raw, 0o644); err != nil {
				die(2, "%v", err)
			}
		}
	}

	// tasks: one per contested/unowned file, two per duplicate-check pair
	var tasks []func()
	var pairs []*dupPair
	fileReq := func(q, scope, p string) string { return q + "|" + scope + "|" + p }
	for _, res := range results {
		if res.SkippedJev != "" {
			continue
		}
		res := res
		files := make([]string, 0, len(res.Contested)+len(res.Unowned))
		for _, c := range res.Contested {
			files = append(files, c.Path)
		}
		files = append(files, res.Unowned...)
		sort.Strings(files)
		for _, p := range files {
			p := p
			tasks = append(tasks, func() {
				covID := fileReq(qOnCovered, res.Scope, p)
				cov, ok := r.get(covID)
				if !ok {
					text, isText, err := rp.excerpt(p, excerptLines)
					if err != nil {
						r.write(jevRow{Type: "failure", ReqID: covID, Question: qOnCovered, Scope: res.Scope, Path: p, Kind: "input", Reason: err.Error(), Timestamp: time.Now().UTC().Format(time.RFC3339)})
						r.mu.Lock()
						r.failures++
						r.mu.Unlock()
						return
					}
					if !isText {
						r.mu.Lock()
						r.skipped[p] = "binary"
						r.mu.Unlock()
						return
					}
					file := map[string]any{"path": p, "role": roleOf(p), "excerpt": text, "sibling_paths": toAny(rp.siblings(p))}
					state, err := marshalState(map[string]any{"file": file, "catalogue": catalogueState(res.Clusters, p)})
					if err != nil {
						die(2, "%v", err)
					}
					if cov, ok = r.ask(oq.covered, versions[qOnCovered], res.Scope+":"+p, state, jevRow{ReqID: covID, Question: qOnCovered, Scope: res.Scope, Path: p}); !ok {
						return
					}
				}
				if cov.PTrue == nil || *cov.PTrue < tCov {
					return
				}
				chID := fileReq(qOnWhich, res.Scope, p)
				if _, have := r.get(chID); have {
					return
				}
				text, isText, err := rp.excerpt(p, excerptLines)
				if err != nil || !isText {
					r.write(jevRow{Type: "failure", ReqID: chID, Question: qOnWhich, Scope: res.Scope, Path: p, Kind: "input", Reason: "file text unavailable", Timestamp: time.Now().UTC().Format(time.RFC3339)})
					r.mu.Lock()
					r.failures++
					r.mu.Unlock()
					return
				}
				q := whichFor[res.Scope]
				version := versions[qOnWhich+"/"+res.Scope]
				for _, c := range res.Clusters { // the file under judgement is never among a sample_files of its catalogue: a per-file question when it would be
					for _, s := range sampleOf(c, "") {
						if s == p {
							if q, err = oq.which(res.Scope, res.Clusters, p); err != nil {
								die(2, "%v", err)
							}
							version = q.Version(model)
						}
					}
				}
				state, err := marshalState(map[string]any{"file": map[string]any{"path": p, "role": roleOf(p), "excerpt": text, "sibling_paths": toAny(rp.siblings(p))}})
				if err != nil {
					die(2, "%v", err)
				}
				r.ask(q, version, res.Scope+":"+p, state, jevRow{ReqID: chID, Question: qOnWhich, Scope: res.Scope, Path: p})
			})
		}
		// duplicate check: clusters of this scope whose titles share tokens
		toks := make([]map[string]bool, len(res.Clusters))
		for i, c := range res.Clusters {
			toks[i] = titleTokens(c.Title)
		}
		for i := 0; i < len(res.Clusters); i++ {
			for j := i + 1; j < len(res.Clusters); j++ {
				var shared []string
				for t := range toks[i] {
					if toks[j][t] {
						shared = append(shared, t)
					}
				}
				if len(shared) < dupShared {
					continue
				}
				sort.Strings(shared)
				ca, cb := res.Clusters[i], res.Clusters[j]
				dp := &dupPair{Scope: res.Scope, AreaA: ca.AreaID, AreaB: cb.AreaID, TitleA: ca.Title, TitleB: cb.Title, Shared: shared}
				pairs = append(pairs, dp)
				for _, order := range []string{"ab", "ba"} {
					order := order
					x, y := ca, cb
					if order == "ba" {
						x, y = cb, ca
					}
					tasks = append(tasks, func() {
						id := fmt.Sprintf("%s|%s|%s|%s|%s", qOnSame, res.Scope, ca.AreaID, cb.AreaID, order)
						if _, have := r.get(id); have {
							return
						}
						side := func(c *cluster) map[string]any {
							return map[string]any{"title": c.Title, "summary": c.Summary, "sample_files": toAny(sampleOf(c, ""))}
						}
						state, err := marshalState(map[string]any{"area_a": side(x), "area_b": side(y)})
						if err != nil {
							die(2, "%v", err)
						}
						r.ask(oq.same, versions[qOnSame], id, state, jevRow{ReqID: id, Question: qOnSame, Scope: res.Scope, AreaA: ca.AreaID, AreaB: cb.AreaID, Order: order})
					})
				}
			}
		}
	}
	queue := make(chan func())
	var wg sync.WaitGroup
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range queue {
				if ctx.Err() != nil {
					continue
				}
				t()
			}
		}()
	}
	for _, t := range tasks {
		queue <- t
	}
	close(queue)
	wg.Wait()
	fmt.Printf("onboard jev: %d tasks, %d answered this run, %d failed, %d with unknown billed cost, cumulative billed cost %.6f USD of which %.6f earlier (model %s)\n",
		len(tasks), r.answered, r.failures, r.unknownCost, r.cost, spent, model)
	if r.stopped {
		die(4, "stopped: billed cost %.4f USD exceeded -max-cost-usd %.4f (no catalogue written; rerun resumes)", r.cost, maxCost)
	}
	if r.failures > 0 || r.unknownCost > 0 {
		die(3, "%d requests failed or have unknown billed cost (no catalogue written; rerun retries the failed requests)", r.failures)
	}

	// decisions from the answer rows (this run's and earlier runs')
	covOf := func(scope, p string) (jevRow, bool) { return r.get(fileReq(qOnCovered, scope, p)) }
	counts := map[string]int{"files_asked": 0, "below_cov": 0, "below_choice": 0, "assigned": 0, "binary": 0}
	skippedScopes := map[string]string{}
	var assigned []assignedRow
	for _, res := range results {
		if res.SkippedJev != "" {
			skippedScopes[res.Scope] = res.SkippedJev
			continue
		}
		files := make([]string, 0)
		for _, c := range res.Contested {
			files = append(files, c.Path)
		}
		files = append(files, res.Unowned...)
		sort.Strings(files)
		for _, p := range files {
			if r.skipped[p] == "binary" {
				counts["binary"]++
				continue
			}
			cov, ok := covOf(res.Scope, p)
			if !ok {
				die(3, "no answer for %s (no catalogue written)", fileReq(qOnCovered, res.Scope, p))
			}
			counts["files_asked"]++
			if cov.PTrue == nil || *cov.PTrue < tCov {
				counts["below_cov"]++
				continue
			}
			ch, ok := r.get(fileReq(qOnWhich, res.Scope, p))
			if !ok {
				die(3, "no answer for %s (no catalogue written)", fileReq(qOnWhich, res.Scope, p))
			}
			if ch.PChoice == nil || *ch.PChoice < tChoice {
				counts["below_choice"]++
				continue
			}
			counts["assigned"]++
			assigned = append(assigned, assignedRow{Path: p, AreaID: ch.Choice, PCov: *cov.PTrue, Confidence: *ch.PChoice})
		}
	}
	sort.Slice(assigned, func(a, b int) bool { return assigned[a].Path < assigned[b].Path })
	for _, a := range assigned {
		i, ok := areaIndex[a.AreaID]
		if !ok {
			die(3, "Jev chose %s, which is not a catalogue area", a.AreaID)
		}
		cat.Areas[i].Files = append(cat.Areas[i].Files, a.Path)
	}
	for i := range cat.Areas {
		sort.Strings(cat.Areas[i].Files)
	}
	cat.JevAssigned = append(cat.JevAssigned, assigned...)
	var dups []dupPair
	dupCount := 0
	for _, dp := range pairs {
		for _, order := range []string{"ab", "ba"} {
			row, ok := r.get(fmt.Sprintf("%s|%s|%s|%s|%s", qOnSame, dp.Scope, dp.AreaA, dp.AreaB, order))
			if !ok {
				die(3, "no answer for a duplicate-check pair (no catalogue written)")
			}
			if order == "ab" {
				dp.PAB = row.PTrue
			} else {
				dp.PBA = row.PTrue
			}
		}
		dp.Duplica = dp.PAB != nil && dp.PBA != nil && *dp.PAB >= dupThreshold && *dp.PBA >= dupThreshold
		if dp.Duplica {
			dupCount++
		}
		dups = append(dups, *dp)
	}
	summary := map[string]any{"ran": true, "model": model, "t_onboard_cov": tCov, "t_onboard_choice": tChoice, "dup_threshold": dupThreshold, "versions": versions, "counts": counts,
		"duplicate_pairs_checked": len(pairs), "duplicates": dupCount, "skipped_scopes": skippedScopes, "billed_cost_usd": r.cost}
	return dups, summary
}
