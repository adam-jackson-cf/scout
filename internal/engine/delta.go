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

// Delta rule identifiers (design §3.2, §8.2). A rule is named and versioned:
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
// never a delta (design §8.2).
func ComputeDeltas(ctx context.Context, root, mergeBase string, changed []string, tools measures.Tools) DeltaReport {
	report := DeltaReport{Issues: []DeltaIssue{}}
	changedProd := map[string]bool{}
	for _, path := range changed {
		if measures.IsProductionSource(path) {
			if info, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err == nil && info.Mode().IsRegular() {
				changedProd[path] = true
			}
		}
	}
	if len(changedProd) == 0 {
		return report
	}
	headPop := deltaHeadPopulation(ctx, root, changedProd)
	basePop := deltaBasePopulation(ctx, root, mergeBase)
	if len(headPop) > deltaMaxFiles || len(basePop) > deltaMaxFiles {
		headPop, basePop = sortedKeys(changedProd), sortedKeys(changedProd)
	}
	tmp, err := os.MkdirTemp("", "scout-delta-")
	if err != nil {
		report.Unmeasured = append(report.Unmeasured, "delta rules: "+err.Error())
		return report
	}
	defer os.RemoveAll(tmp)
	extracted, err := extractBlobs(ctx, root, mergeBase, basePop, tmp)
	if err != nil {
		report.Unmeasured = append(report.Unmeasured, "delta rules: base content: "+err.Error())
		return report
	}

	report.Issues = append(report.Issues, ccDeltas(ctx, root, tmp, mergeBase, changedProd, extracted, tools, &report)...)
	report.Issues = append(report.Issues, dupDeltas(ctx, root, tmp, mergeBase, changedProd, headPop, extracted, tools, &report)...)
	sort.Slice(report.Issues, func(i, j int) bool {
		a, b := report.Issues[i], report.Issues[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Rule < b.Rule
	})
	return report
}

func ccDeltas(ctx context.Context, root, baseRoot, mergeBase string, changedProd map[string]bool, extracted map[string]bool, tools measures.Tools, report *DeltaReport) []DeltaIssue {
	headPaths := sortedKeys(changedProd)
	var basePaths []string
	for _, path := range headPaths {
		if extracted[path] {
			basePaths = append(basePaths, path)
		}
	}
	if len(basePaths) == 0 {
		return nil
	}
	head, err := tools.RunLizard(ctx, root, basePaths)
	if err != nil {
		report.Unmeasured = append(report.Unmeasured, DeltaCCRule+": lizard: "+err.Error())
		return nil
	}
	base, err := tools.RunLizard(ctx, baseRoot, basePaths)
	if err != nil {
		report.Unmeasured = append(report.Unmeasured, DeltaCCRule+": lizard: "+err.Error())
		return nil
	}
	var out []DeltaIssue
	for _, path := range basePaths {
		hm, hok := head[path]
		bm, bok := base[path]
		// A file outside the eligible population on either side is a
		// population change, not a delta.
		if !hok || !bok || hm.NLOC < measures.MinFileNLOC || bm.NLOC < measures.MinFileNLOC {
			continue
		}
		baseCC, baseDup := functionCC(bm.Functions)
		headCC, headDup := functionCC(hm.Functions)
		names := make([]string, 0, len(headCC))
		for name := range headCC {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			if baseDup[name] || headDup[name] {
				continue // ambiguous overloads cannot be matched
			}
			b, ok := baseCC[name]
			if !ok {
				continue
			}
			h := headCC[name]
			if b.CC <= measures.HighCC && h.CC > measures.HighCC {
				out = append(out, DeltaIssue{
					ID: deltaID(DeltaCCRule, path, name, ""), Rule: deltaRule(DeltaCCRule), Claim: domain.ClaimDescriptive, Mode: deltaMode,
					Family: domain.FamilyQuality, Path: path, Function: name, Line: h.Line,
					Title: fmt.Sprintf("function %s cyclomatic complexity %d → %d crosses %d", name, b.CC, h.CC, measures.HighCC),
					Base:  b.CC, Head: h.CC, Threshold: measures.HighCC, BaseRev: mergeBase,
				})
			}
		}
	}
	return out
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

func dupDeltas(ctx context.Context, root, baseRoot, mergeBase string, changedProd map[string]bool, headPop []string, extracted map[string]bool, tools measures.Tools, report *DeltaReport) []DeltaIssue {
	var basePop []string
	for path := range extracted {
		basePop = append(basePop, path)
	}
	sort.Strings(basePop)
	headClones, err := tools.RunJscpd(ctx, root, headPop)
	if err != nil {
		report.Unmeasured = append(report.Unmeasured, DeltaDupRule+": jscpd: "+err.Error())
		return nil
	}
	baseClones, err := tools.RunJscpd(ctx, baseRoot, basePop)
	if err != nil {
		report.Unmeasured = append(report.Unmeasured, DeltaDupRule+": jscpd: "+err.Error())
		return nil
	}
	key := func(c measures.Clone) string {
		a, b := c.A.Path, c.B.Path
		if b < a {
			a, b = b, a
		}
		return a + "\x00" + b
	}
	baseCount := map[string]int{}
	for _, clone := range baseClones {
		baseCount[key(clone)]++
	}
	headCount := map[string]int{}
	first := map[string]measures.Clone{}
	for _, clone := range headClones {
		if !changedProd[clone.A.Path] && !changedProd[clone.B.Path] {
			continue
		}
		k := key(clone)
		headCount[k]++
		if _, ok := first[k]; !ok {
			first[k] = clone
		}
	}
	keys := make([]string, 0, len(headCount))
	for k := range headCount {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var out []DeltaIssue
	for _, k := range keys {
		if headCount[k] <= baseCount[k] {
			continue
		}
		clone := first[k]
		mine, other := clone.A, clone.B
		if !changedProd[mine.Path] {
			mine, other = other, mine
		}
		lines := mine.End - mine.Start + 1
		out = append(out, DeltaIssue{
			ID: deltaID(DeltaDupRule, mine.Path, "", other.Path), Rule: deltaRule(DeltaDupRule), Claim: domain.ClaimDescriptive, Mode: deltaMode,
			Family: domain.FamilyQuality, Path: mine.Path, Line: mine.Start, Partner: other.Path,
			Title: fmt.Sprintf("clone added: %s duplicates %s (%d lines)", mine.Path, other.Path, lines),
			Base:  baseCount[k], Head: headCount[k], BaseRev: mergeBase,
		})
	}
	return out
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
func deltaHeadPopulation(ctx context.Context, root string, changedProd map[string]bool) []string {
	set := map[string]bool{}
	for path := range changedProd {
		set[path] = true
	}
	if out, err := git(ctx, root, "ls-files", "-z", "--cached", "--others", "--exclude-standard"); err == nil {
		for _, name := range strings.Split(string(out), "\x00") {
			if name == "" || !measures.IsProductionSource(name) {
				continue
			}
			if info, statErr := os.Stat(filepath.Join(root, filepath.FromSlash(name))); statErr == nil && info.Mode().IsRegular() {
				set[name] = true
			}
		}
	}
	return sortedKeys(set)
}

// deltaBasePopulation is the production source tracked at the merge base.
func deltaBasePopulation(ctx context.Context, root, mergeBase string) []string {
	out, err := git(ctx, root, "ls-tree", "-r", "-z", "--name-only", mergeBase)
	if err != nil {
		return nil
	}
	var paths []string
	for _, name := range strings.Split(string(out), "\x00") {
		if name != "" && measures.IsProductionSource(name) {
			paths = append(paths, name)
		}
	}
	sort.Strings(paths)
	return paths
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
	reader := bufio.NewReader(stdout)
	var readErr error
	for _, path := range paths {
		header, err := reader.ReadString('\n')
		if err != nil {
			readErr = err
			break
		}
		fields := strings.Fields(header)
		if len(fields) != 3 { // "<spec> missing" or ambiguous
			continue
		}
		size, err := strconv.Atoi(fields[2])
		if err != nil {
			readErr = err
			break
		}
		if fields[1] != "blob" || size > deltaMaxBlob {
			if _, err := io.CopyN(io.Discard, reader, int64(size)+1); err != nil {
				readErr = err
				break
			}
			continue
		}
		content := make([]byte, size+1)
		if _, err := io.ReadFull(reader, content); err != nil {
			readErr = err
			break
		}
		target := filepath.Join(dir, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			readErr = err
			break
		}
		if err := os.WriteFile(target, content[:size], 0o600); err != nil {
			readErr = err
			break
		}
		extracted[path] = true
	}
	if readErr != nil {
		_ = cmd.Process.Kill()
	}
	_ = cmd.Wait()
	if readErr != nil {
		return nil, readErr
	}
	return extracted, nil
}

// deltaLane promotes each delta to a ledger observation through the ordinary
// inline-lane path (design §3.2). The observation carries the rule id, version
// and `descriptive` claim, so the issue it correlates into is report-only: its
// check id `delta.*` is never blocking-eligible. Rules that could not run are
// reported as lane health, never as an empty result.
func deltaLane(root, mergeBase string, changed []string, tools measures.Tools) InlineLane {
	var unmeasured []string
	return InlineLane{
		Lane: LaneDelta, Units: len(changed),
		Reason: func() string {
			if len(unmeasured) == 0 {
				return ""
			}
			return "unmeasured: " + strings.Join(unmeasured, "; ")
		},
		Run: func(ctx context.Context, runID string) ([]domain.Observation, error) {
			if len(changed) == 0 {
				return nil, nil
			}
			report := ComputeDeltas(ctx, root, mergeBase, changed, tools)
			unmeasured = report.Unmeasured
			now := time.Now().UTC()
			observations := make([]domain.Observation, 0, len(report.Issues))
			for _, issue := range report.Issues {
				observations = append(observations, issue.observation(runID, now))
			}
			return observations, nil
		},
	}
}

// observation renders the delta as a validated ledger observation. The anchor
// is file + function (design §3.3): evidence carries the measured values, the
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
