package engine

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"scout/internal/domain"
	"scout/internal/measures"
)

// Delta rule identifiers (REQ-8, REQ-125). A rule is named and versioned:
// changing its threshold or matching bumps the version.
const (
	DeltaCCRule  = "delta.cc-threshold"
	DeltaDupRule = "delta.dup-added"

	deltaRuleVersion = 1
	// deltaMode: delta issues never block, whatever the policy says.
	deltaMode = "report-only"
	// deltaMaxFiles bounds the whole-tree clone population; a larger tree is
	// compared over the changed files only.
	deltaMaxFiles = 3000
	deltaMaxBlob  = 4 << 20
)

// DeltaIssue is a `delta`-rule issue: this PR moved a unit across a stated
// threshold. It is descriptive and report-only: it is never an input to the
// gate, the exit code, or the would-block set.
type DeltaIssue struct {
	ID       string             `json:"id"`
	Rule     string             `json:"rule"` // name@version
	Claim    domain.SignalClaim `json:"claim"`
	Mode     string             `json:"mode"`
	Family   domain.Family      `json:"family"`
	Path     string             `json:"path"`
	Function string             `json:"function,omitempty"`
	Line     int                `json:"line"`
	Title    string             `json:"title"`
	// Base and Head are the measured values either side of the merge base;
	// Threshold is the stated line that was crossed (CC rule only).
	Base      int    `json:"base"`
	Head      int    `json:"head"`
	Threshold int    `json:"threshold,omitempty"`
	Partner   string `json:"partner,omitempty"` // other file of an added clone
	BaseRev   string `json:"baseRevision"`
}

// DeltaReport lists the delta issues of a PR run. Unmeasured names each rule
// that could not run (tool absent); an absent measure never reads as no delta.
type DeltaReport struct {
	Issues     []DeltaIssue `json:"issues"`
	Unmeasured []string     `json:"unmeasured,omitempty"`
}

// ComputeDeltas applies the delta rules to the changed files between the merge
// base and the tree at root. Functions are matched by file and name in both
// sides; files or functions present on one side only are a population change,
// never a delta (REQ-125).
func ComputeDeltas(ctx context.Context, root, mergeBase string, changed []string, tools measures.Tools) DeltaReport {
	report := DeltaReport{Issues: []DeltaIssue{}}
	changedProd := deltaChangedProduction(root, changed)
	if len(changedProd) == 0 {
		return report
	}
	headPop, basePop, err := deltaPopulations(ctx, root, mergeBase, changedProd)
	if err != nil {
		report.Unmeasured = append(report.Unmeasured, "delta rules: "+err.Error())
		return report
	}
	tmp, err := os.MkdirTemp("", "scout-delta-")
	if err != nil {
		report.Unmeasured = append(report.Unmeasured, "delta rules: "+err.Error())
		return report
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	extracted, err := extractBlobs(ctx, root, mergeBase, basePop, tmp)
	if err != nil {
		report.Unmeasured = append(report.Unmeasured, "delta rules: base content: "+err.Error())
		return report
	}

	run := deltaRun{root: root, baseRoot: tmp, mergeBase: mergeBase, changedProd: changedProd, extracted: extracted, tools: tools, report: &report}
	report.Issues = append(report.Issues, run.ccDeltas(ctx)...)
	report.Issues = append(report.Issues, run.dupDeltas(ctx, headPop)...)
	sortDeltaIssues(report.Issues)
	return report
}

func deltaChangedProduction(root string, changed []string) map[string]bool {
	changedProd := map[string]bool{}
	for _, path := range changed {
		if measures.IsProductionSource(path) {
			if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err == nil && info.Mode().IsRegular() {
				changedProd[path] = true
			}
		}
	}
	return changedProd
}

func deltaPopulations(ctx context.Context, root, mergeBase string, changedProd map[string]bool) ([]string, []string, error) {
	headPop, err := deltaHeadPopulation(ctx, root, changedProd)
	if err != nil {
		return nil, nil, err
	}
	basePop, err := deltaBasePopulation(ctx, root, mergeBase)
	if err != nil {
		return nil, nil, err
	}
	if len(headPop) > deltaMaxFiles || len(basePop) > deltaMaxFiles {
		headPop, basePop = sortedKeys(changedProd), sortedKeys(changedProd)
	}
	return headPop, basePop, nil
}

func sortDeltaIssues(issues []DeltaIssue) {
	sort.Slice(issues, func(i, j int) bool {
		a, b := issues[i], issues[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Rule < b.Rule
	})
}

type deltaRun struct {
	root, baseRoot, mergeBase string
	changedProd               map[string]bool
	extracted                 map[string]bool
	tools                     measures.Tools
	report                    *DeltaReport
}

func (r deltaRun) ccDeltas(ctx context.Context) []DeltaIssue {
	basePaths := ccDeltaPaths(r.changedProd, r.extracted)
	if len(basePaths) == 0 {
		return nil
	}
	head, err := r.tools.RunLizard(ctx, r.root, basePaths)
	if err != nil {
		r.report.Unmeasured = append(r.report.Unmeasured, DeltaCCRule+": lizard: "+err.Error())
		return nil
	}
	base, err := r.tools.RunLizard(ctx, r.baseRoot, basePaths)
	if err != nil {
		r.report.Unmeasured = append(r.report.Unmeasured, DeltaCCRule+": lizard: "+err.Error())
		return nil
	}
	var out []DeltaIssue
	for _, path := range basePaths {
		out = append(out, ccFileDeltas(path, head, base, r.mergeBase)...)
	}
	return out
}

func ccDeltaPaths(changedProd map[string]bool, extracted map[string]bool) []string {
	var basePaths []string
	for _, path := range sortedKeys(changedProd) {
		if extracted[path] {
			basePaths = append(basePaths, path)
		}
	}
	return basePaths
}

func ccMeasuredPair(path string, head, base map[string]measures.FileMeasure) (measures.FileMeasure, measures.FileMeasure, bool) {
	hm, hok := head[path]
	bm, bok := base[path]
	// A file outside the eligible population on either side is a
	// population change, not a delta.
	if !hok || !bok || hm.NLOC < measures.MinFileNLOC || bm.NLOC < measures.MinFileNLOC {
		return hm, bm, false
	}
	return hm, bm, true
}

func ccFileDeltas(path string, head, base map[string]measures.FileMeasure, mergeBase string) []DeltaIssue {
	hm, bm, ok := ccMeasuredPair(path, head, base)
	if !ok {
		return nil
	}
	baseCC, baseDup := functionCC(bm.Functions)
	headCC, headDup := functionCC(hm.Functions)
	var out []DeltaIssue
	for _, name := range sortedFunctionNames(headCC) {
		if baseDup[name] || headDup[name] {
			continue // ambiguous overloads cannot be matched
		}
		b, ok := baseCC[name]
		if !ok {
			continue
		}
		if issue, crossed := ccCrossing(path, name, b, headCC[name], mergeBase); crossed {
			out = append(out, issue)
		}
	}
	return out
}

func sortedFunctionNames(byName map[string]measures.Function) []string {
	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func ccCrossing(path, name string, b, h measures.Function, mergeBase string) (DeltaIssue, bool) {
	if b.CC > measures.HighCC || h.CC <= measures.HighCC {
		return DeltaIssue{}, false
	}
	return DeltaIssue{
		ID: deltaID(DeltaCCRule, path, name, ""), Rule: deltaRule(DeltaCCRule), Claim: domain.ClaimDescriptive, Mode: deltaMode,
		Family: domain.FamilyQuality, Path: path, Function: name, Line: h.Line,
		Title: fmt.Sprintf("function %s cyclomatic complexity %d → %d crosses %d", name, b.CC, h.CC, measures.HighCC),
		Base:  b.CC, Head: h.CC, Threshold: measures.HighCC, BaseRev: mergeBase,
	}, true
}

func functionCC(functions []measures.Function) (map[string]measures.Function, map[string]bool) {
	byName := make(map[string]measures.Function, len(functions))
	dup := map[string]bool{}
	for _, fn := range functions {
		if _, seen := byName[fn.Name]; seen {
			dup[fn.Name] = true
		}
		byName[fn.Name] = fn
	}
	return byName, dup
}

func (r deltaRun) dupDeltas(ctx context.Context, headPop []string) []DeltaIssue {
	headClones, err := r.tools.RunJscpd(ctx, r.root, headPop)
	if err != nil {
		r.report.Unmeasured = append(r.report.Unmeasured, DeltaDupRule+": jscpd: "+err.Error())
		return nil
	}
	baseClones, err := r.tools.RunJscpd(ctx, r.baseRoot, deltaExtractedPaths(r.extracted))
	if err != nil {
		r.report.Unmeasured = append(r.report.Unmeasured, DeltaDupRule+": jscpd: "+err.Error())
		return nil
	}
	counts := countDeltaClones(baseClones, headClones, r.changedProd)
	return counts.issues(r.changedProd, r.mergeBase)
}

func deltaExtractedPaths(extracted map[string]bool) []string {
	var basePop []string
	for path := range extracted {
		basePop = append(basePop, path)
	}
	sort.Strings(basePop)
	return basePop
}

func cloneDeltaKey(c measures.Clone) string {
	a, b := c.A.Path, c.B.Path
	if b < a {
		a, b = b, a
	}
	return a + "\x00" + b
}

type cloneDeltaCounts struct {
	base  map[string]int
	head  map[string]int
	first map[string]measures.Clone
}

func countDeltaClones(baseClones, headClones []measures.Clone, changedProd map[string]bool) cloneDeltaCounts {
	counts := cloneDeltaCounts{base: map[string]int{}, head: map[string]int{}, first: map[string]measures.Clone{}}
	for _, clone := range baseClones {
		counts.base[cloneDeltaKey(clone)]++
	}
	for _, clone := range headClones {
		if !changedProd[clone.A.Path] && !changedProd[clone.B.Path] {
			continue
		}
		k := cloneDeltaKey(clone)
		counts.head[k]++
		if _, ok := counts.first[k]; !ok {
			counts.first[k] = clone
		}
	}
	return counts
}

func (c cloneDeltaCounts) issues(changedProd map[string]bool, mergeBase string) []DeltaIssue {
	keys := make([]string, 0, len(c.head))
	for k := range c.head {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []DeltaIssue
	for _, k := range keys {
		if c.head[k] <= c.base[k] {
			continue
		}
		out = append(out, cloneDeltaIssue(c.first[k], changedProd, c.base[k], c.head[k], mergeBase))
	}
	return out
}

func cloneDeltaIssue(clone measures.Clone, changedProd map[string]bool, baseCount, headCount int, mergeBase string) DeltaIssue {
	mine, other := clone.A, clone.B
	if !changedProd[mine.Path] {
		mine, other = other, mine
	}
	lines := mine.End - mine.Start + 1
	return DeltaIssue{
		ID: deltaID(DeltaDupRule, mine.Path, "", other.Path), Rule: deltaRule(DeltaDupRule), Claim: domain.ClaimDescriptive, Mode: deltaMode,
		Family: domain.FamilyQuality, Path: mine.Path, Line: mine.Start, Partner: other.Path,
		Title: fmt.Sprintf("clone added: %s duplicates %s (%d lines)", mine.Path, other.Path, lines),
		Base:  baseCount, Head: headCount, BaseRev: mergeBase,
	}
}

func deltaRule(name string) string { return name + "@" + strconv.Itoa(deltaRuleVersion) }

func deltaID(rule, path, function, partner string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{deltaRule(rule), path, function, partner}, "\x00")))
	return "delta:" + hex.EncodeToString(sum[:12])
}

func sortedKeys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}

// deltaHeadPopulation is the head clone population: tracked and untracked
// non-ignored production source plus the changed files.
func deltaHeadPopulation(ctx context.Context, root string, changedProd map[string]bool) ([]string, error) {
	set := map[string]bool{}
	for path := range changedProd {
		set[path] = true
	}
	out, err := git(ctx, root, "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	if err != nil {
		return nil, fmt.Errorf("list head files: %w", err)
	}
	for _, name := range strings.Split(string(out), "\x00") {
		if name == "" || !measures.IsProductionSource(name) {
			continue
		}
		if info, statErr := os.Stat(filepath.Join(root, filepath.FromSlash(name))); statErr == nil && info.Mode().IsRegular() {
			set[name] = true
		}
	}
	return sortedKeys(set), nil
}

// deltaBasePopulation is the production source tracked at the merge base.
func deltaBasePopulation(ctx context.Context, root, mergeBase string) ([]string, error) {
	out, err := git(ctx, root, "ls-tree", "-r", "-z", "--name-only", mergeBase)
	if err != nil {
		return nil, fmt.Errorf("list merge-base files: %w", err)
	}
	var paths []string
	for _, name := range strings.Split(string(out), "\x00") {
		if name != "" && measures.IsProductionSource(name) {
			paths = append(paths, name)
		}
	}
	sort.Strings(paths)
	return paths, nil
}

// extractBlobs writes the merge-base content of paths below dir with a single
// `git cat-file --batch`; paths absent at the base are left out of the result.
func extractBlobs(ctx context.Context, root, rev string, paths []string, dir string) (map[string]bool, error) {
	extracted := map[string]bool{}
	if len(paths) == 0 {
		return extracted, nil
	}
	cmd := exec.CommandContext(ctx, "git", "-C", root, "cat-file", "--batch")
	var input bytes.Buffer
	for _, path := range paths {
		input.WriteString(rev + ":" + path + "\n")
	}
	cmd.Stdin = &input
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	readErr := readBlobBatch(bufio.NewReader(stdout), paths, dir, extracted)
	if readErr != nil {
		_ = cmd.Process.Kill()
	}
	_ = cmd.Wait()
	if readErr != nil {
		return nil, readErr
	}
	return extracted, nil
}

func readBlobBatch(reader *bufio.Reader, paths []string, dir string, extracted map[string]bool) error {
	for _, path := range paths {
		written, err := readBlob(reader, path, dir)
		if err != nil {
			return err
		}
		if written {
			extracted[path] = true
		}
	}
	return nil
}

func readBlob(reader *bufio.Reader, path, dir string) (bool, error) {
	header, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}
	fields := strings.Fields(header)
	if len(fields) != 3 { // "<spec> missing" or ambiguous
		return false, nil
	}
	size, err := strconv.Atoi(fields[2])
	if err != nil {
		return false, err
	}
	if fields[1] != "blob" || size > deltaMaxBlob {
		_, err := io.CopyN(io.Discard, reader, int64(size)+1)
		return false, err
	}
	content := make([]byte, size+1)
	if _, err := io.ReadFull(reader, content); err != nil {
		return false, err
	}
	if err := writeDeltaBlob(dir, path, content[:size]); err != nil {
		return false, err
	}
	return true, nil
}

func writeDeltaBlob(dir, path string, content []byte) error {
	target := filepath.Join(dir, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return err
	}
	return os.WriteFile(target, content, 0o600)
}

// deltaLane promotes each delta to a ledger observation through the ordinary
// inline-lane path (REQ-8). The observation carries the rule id, version
// and `descriptive` claim, so the issue it correlates into is report-only: its
// check id `delta.*` is never blocking-eligible. Rules that could not run are
// reported as lane health, never as an empty result.
func deltaLane(root, mergeBase string, changed []string, tools measures.Tools) InlineLane {
	return InlineLane{
		Lane: LaneDelta, Units: len(changed),
		Run: func(ctx context.Context, runID string) (InlineResult, error) {
			if len(changed) == 0 {
				return InlineResult{}, nil
			}
			report := ComputeDeltas(ctx, root, mergeBase, changed, tools)
			now := time.Now().UTC()
			result := InlineResult{Observations: make([]domain.Observation, 0, len(report.Issues))}
			for _, issue := range report.Issues {
				result.Observations = append(result.Observations, issue.observation(runID, now))
			}
			if len(report.Unmeasured) > 0 {
				result.Unmeasured = "unmeasured: " + strings.Join(report.Unmeasured, "; ")
			}
			return result, nil
		},
	}
}

// observation renders the delta as a validated ledger observation. The anchor
// is file + function (REQ-5): evidence carries the measured values, the
// threshold and the base revision, so a re-run on the same PR matches the same
// issue.
func (d DeltaIssue) observation(runID string, now time.Time) domain.Observation {
	line := d.Line
	if line < 1 {
		line = 1
	}
	quote := fmt.Sprintf("base=%d head=%d", d.Base, d.Head)
	if d.Threshold != 0 {
		quote += fmt.Sprintf(" threshold=%d", d.Threshold)
	}
	quote += " baseRevision=" + d.BaseRev
	category := "complexity-delta"
	if d.Partner != "" {
		category = "clone-delta"
	}
	rule, version, _ := strings.Cut(d.Rule, "@")
	sum := sha256.Sum256([]byte(strings.Join([]string{runID, d.Rule, d.Path, d.Function, d.Partner}, "\x00")))
	return domain.Observation{
		ObservationID: "delta:" + hex.EncodeToString(sum[:12]), RunID: runID, Source: "delta",
		Family: domain.FamilyQuality, OriginalCategory: category, RuleID: d.Rule, RuleVersion: version, Claim: d.Claim,
		SourceSeverity: domain.SeverityInformational, EffectiveSeverity: domain.SeverityInformational, Confidence: domain.ConfidenceHigh,
		Mechanism: rule + ": " + d.Title, Validated: true, ObservedAt: now,
		Evidence: []domain.Evidence{{Location: domain.SourceRange{Path: d.Path, StartLine: line, EndLine: line, Symbol: d.Function}, RedactedQuote: quote}},
	}
}
