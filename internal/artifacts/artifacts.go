// Package artifacts ingests CI-produced coverage (LCOV, Cobertura XML) and
// test-result (JUnit XML) files with provenance (design §8.2, §8.7).
//
// Coverage is only usable for a snapshot when the artifact's stated revision
// matches the snapshot's commit; an artifact with no stated revision, a
// different revision, or a snapshot that is a dirty worktree never matches.
// Nothing here fabricates a value: a non-matching artifact is reported with a
// reason and consumers record `unmeasured`.
package artifacts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"scout/internal/domain"
)

// Kinds of artifact.
const (
	KindLCOV      = "lcov"
	KindCobertura = "cobertura"
	KindJUnit     = "junit"
)

// Revision-match outcomes.
const (
	MatchMatch    = "match"
	MatchMismatch = "mismatch"
	MatchUnknown  = "unknown"
)

// SidecarSuffix names the provenance file next to an artifact:
// `coverage.lcov.scout.json` holds {"revision": "<sha>", "producer": "<name>"}.
const SidecarSuffix = ".scout.json"

const maxArtifactBytes = 256 << 20

// Spec names one artifact to ingest. Revision and Producer are defaults used
// when the artifact has no sidecar provenance file.
type Spec struct {
	Path     string
	Revision string
	Producer string
}

// Provenance records where an artifact came from.
type Provenance struct {
	Source   string `json:"source"`
	Digest   string `json:"digest"` // SHA-256 of the artifact bytes
	Kind     string `json:"kind"`
	Revision string `json:"revision,omitempty"`
	Producer string `json:"producer,omitempty"`
	// RevisionMatch is match, mismatch or unknown against the snapshot.
	RevisionMatch string `json:"revisionMatch"`
	Reason        string `json:"reason,omitempty"`
}

// Coverage is per-file line coverage: path (repository-relative) → line →
// hit count, over instrumented lines only.
type Coverage map[string]map[int]int

// TestResults summarises a JUnit report.
type TestResults struct {
	Suites   int     `json:"suites"`
	Tests    int     `json:"tests"`
	Failures int     `json:"failures"`
	Errors   int     `json:"errors"`
	Skipped  int     `json:"skipped"`
	Seconds  float64 `json:"seconds"`
}

// Artifact is one parsed artifact. Exactly one of Coverage and Tests is set.
type Artifact struct {
	Provenance
	Coverage Coverage
	Tests    *TestResults
}

// Failure is an artifact that could not be ingested.
type Failure struct {
	Source string
	Err    error
}

// Result is the outcome of ingesting a set of specs against a snapshot.
type Result struct {
	Artifacts []Artifact
	Failures  []Failure
}

// Load reads every spec (a path or a glob, relative to root), parses it and
// establishes its revision match against the snapshot. Failures never abort
// the load. Artifacts are ordered by source path.
func Load(root string, specs []Spec, snapshot domain.Snapshot) Result {
	var result Result
	seen := map[string]bool{}
	roots := []string{root}
	if resolved, err := filepath.EvalSymlinks(root); err == nil && resolved != root {
		roots = append(roots, resolved)
		root = resolved
	}
	for _, spec := range specs {
		paths, err := expand(root, spec.Path)
		if err != nil {
			result.Failures = append(result.Failures, Failure{Source: spec.Path, Err: err})
			continue
		}
		for _, path := range paths {
			if seen[path] {
				continue
			}
			seen[path] = true
			artifact, err := loadOne(root, roots, path, spec, snapshot)
			if err != nil {
				result.Failures = append(result.Failures, Failure{Source: path, Err: err})
				continue
			}
			result.Artifacts = append(result.Artifacts, artifact)
		}
	}
	sort.SliceStable(result.Artifacts, func(i, j int) bool { return result.Artifacts[i].Source < result.Artifacts[j].Source })
	return result
}

func expand(root, pattern string) ([]string, error) {
	if pattern == "" {
		return nil, errors.New("empty artifact path")
	}
	if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(root, pattern)
	}
	if !strings.ContainsAny(pattern, "*?[") {
		return []string{pattern}, nil
	}
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("artifact pattern: %w", err)
	}
	kept := matches[:0]
	for _, match := range matches {
		if !strings.HasSuffix(match, SidecarSuffix) {
			kept = append(kept, match)
		}
	}
	if len(kept) == 0 {
		return nil, errors.New("artifact pattern matched no file")
	}
	sort.Strings(kept)
	return kept, nil
}

func loadOne(root string, roots []string, path string, spec Spec, snapshot domain.Snapshot) (Artifact, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Artifact{}, err
	}
	if !info.Mode().IsRegular() {
		return Artifact{}, errors.New("artifact is not a regular file")
	}
	if info.Size() > maxArtifactBytes {
		return Artifact{}, fmt.Errorf("artifact exceeds %d bytes", maxArtifactBytes)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return Artifact{}, err
	}
	sum := sha256.Sum256(raw)
	artifact := Artifact{Provenance: Provenance{Source: displayPath(root, path), Digest: hex.EncodeToString(sum[:]), Revision: spec.Revision, Producer: spec.Producer}}
	if side, err := os.ReadFile(path + SidecarSuffix); err == nil {
		var meta struct {
			Revision string `json:"revision"`
			Producer string `json:"producer"`
		}
		if err := json.Unmarshal(side, &meta); err != nil {
			return Artifact{}, fmt.Errorf("provenance sidecar: %w", err)
		}
		artifact.Revision, artifact.Producer = meta.Revision, meta.Producer
	}
	switch detect(raw) {
	case KindLCOV:
		artifact.Kind = KindLCOV
		artifact.Coverage, err = ParseLCOV(raw, roots...)
	case KindCobertura:
		artifact.Kind = KindCobertura
		artifact.Coverage, err = ParseCobertura(raw, roots...)
	case KindJUnit:
		artifact.Kind = KindJUnit
		var tests TestResults
		tests, err = ParseJUnit(raw)
		artifact.Tests = &tests
	default:
		return Artifact{}, errors.New("unrecognized artifact format (expected LCOV, Cobertura XML or JUnit XML)")
	}
	if err != nil {
		return Artifact{}, err
	}
	artifact.RevisionMatch, artifact.Reason = MatchRevision(artifact.Revision, snapshot)
	return artifact, nil
}

func displayPath(root, path string) string {
	if rel, err := filepath.Rel(root, path); err == nil && !strings.HasPrefix(rel, "..") {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(path)
}

// detect classifies content: XML by root element, else LCOV by record tags.
func detect(raw []byte) string {
	head := string(raw[:min(len(raw), 4096)])
	trimmed := strings.TrimSpace(head)
	if strings.HasPrefix(trimmed, "<") {
		switch {
		case strings.Contains(head, "<coverage"):
			return KindCobertura
		case strings.Contains(head, "<testsuite"):
			return KindJUnit
		}
		return ""
	}
	for _, line := range strings.Split(head, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TN:") || strings.HasPrefix(line, "SF:") {
			return KindLCOV
		}
	}
	return ""
}

// MatchRevision compares an artifact's stated revision with the snapshot. Only
// a clean checkout of a commit can match; abbreviated revisions of at least
// seven characters match by prefix.
func MatchRevision(revision string, snapshot domain.Snapshot) (string, string) {
	revision = strings.ToLower(strings.TrimSpace(revision))
	switch {
	case revision == "":
		return MatchUnknown, "the artifact states no revision (provide a sidecar " + SidecarSuffix + " file or --artifact-revision)"
	case snapshot.Revision == "":
		return MatchMismatch, "the snapshot has no commit revision to compare with"
	case snapshot.Dirty:
		return MatchMismatch, "the working tree is dirty, so the snapshot is not the commit " + short(snapshot.Revision)
	}
	head := strings.ToLower(snapshot.Revision)
	if revision == head || (len(revision) >= 7 && strings.HasPrefix(head, revision)) {
		return MatchMatch, ""
	}
	return MatchMismatch, fmt.Sprintf("the artifact was produced at %s but the snapshot is %s", short(revision), short(head))
}

func short(revision string) string {
	if len(revision) > 12 {
		return revision[:12]
	}
	return revision
}

// Merged combines the revision-matched coverage artifacts of a result. Hit
// counts add across artifacts. The digest is stable for the artifact set; it
// is empty when no coverage artifact matched, and Reason then explains why.
type Merged struct {
	Lines   Coverage
	Digest  string
	Sources []string
	Reason  string
}

// MatchedCoverage merges the coverage artifacts whose revision matched.
func (r Result) MatchedCoverage() Merged {
	var digests, sources, reasons []string
	lines := Coverage{}
	for _, artifact := range r.Artifacts {
		if artifact.Coverage == nil {
			continue
		}
		if artifact.RevisionMatch != MatchMatch {
			reasons = append(reasons, fmt.Sprintf("%s: %s", artifact.Source, artifact.Reason))
			continue
		}
		digests = append(digests, artifact.Digest)
		sources = append(sources, artifact.Source)
		for path, perLine := range artifact.Coverage {
			target := lines[path]
			if target == nil {
				target = map[int]int{}
				lines[path] = target
			}
			for line, hits := range perLine {
				target[line] += hits
			}
		}
	}
	if len(digests) == 0 {
		reason := "no coverage artifact was provided"
		if len(r.Failures) > 0 && len(reasons) == 0 {
			parts := make([]string, len(r.Failures))
			for i, failure := range r.Failures {
				parts[i] = fmt.Sprintf("%s: %v", failure.Source, failure.Err)
			}
			reason = "coverage artifact could not be read: " + strings.Join(parts, "; ")
		} else if len(reasons) > 0 {
			reason = "no coverage artifact matches the snapshot revision: " + strings.Join(reasons, "; ")
		}
		return Merged{Reason: reason}
	}
	sort.Strings(digests)
	return Merged{Lines: lines, Digest: domain.Digest(digests...), Sources: sources}
}

// normalizePath maps a path in an artifact to a repository-relative slash
// path; absolute paths outside the root are returned cleaned as given.
func normalizePath(roots []string, path string) string {
	path = filepath.FromSlash(strings.TrimSpace(path))
	if filepath.IsAbs(path) {
		for _, root := range roots {
			if rel, err := filepath.Rel(root, path); err == nil && !strings.HasPrefix(rel, "..") {
				path = rel
				break
			}
		}
	}
	path = filepath.ToSlash(filepath.Clean(path))
	return strings.TrimPrefix(path, "./")
}
