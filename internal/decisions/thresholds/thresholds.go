// Package thresholds holds the named, per-question threshold configuration of
// the decisions adapter (jev-actions threshold-tuning: a threshold lives in
// configuration beside its minimums, metric, label set and version key, never
// as a literal in code).
//
// Defaults come from an embedded file; `.scout/decisions/thresholds.json`
// overrides an entry. This package stores and compares configuration only: it
// computes no quality metric, and it does not import the adapter.
package thresholds

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os/exec"
	"strings"
)

// Path is the repository-relative override file.
const Path = ".scout/decisions/thresholds.json"

// Statuses of an entry.
const (
	StatusProvisional = "provisional"
	StatusReady       = "ready"
)

// Score spaces a threshold applies to.
const (
	SpaceRaw   = "raw"
	SpacePlatt = "platt"
)

//go:embed defaults.json
var defaultsJSON []byte

// LabelSet summarises the labelled cases a threshold was tuned on. The
// summary is `none` when no case exists.
type LabelSet struct {
	Summary string `json:"summary"`
}

// Entry is one decision's threshold record.
type Entry struct {
	QuestionID string `json:"question_id"`
	// VersionKey is the question version key the entry was set against; empty
	// in the embedded defaults, which bind to the current key on load.
	VersionKey string `json:"version_key,omitempty"`
	// Gating is true only when some behaviour reads Threshold. Launch ranking
	// is ordinal, so nothing does.
	Gating    bool     `json:"gating"`
	Threshold *float64 `json:"threshold"`
	// Minimums are the agreed or assumed floors the threshold must meet.
	MinPrecision       float64  `json:"min_precision"`
	MinRecall          float64  `json:"min_recall"`
	MinimumsProvenance string   `json:"minimums_provenance"`
	Metric             string   `json:"metric"`
	Beta               float64  `json:"beta,omitempty"`
	LabelSet           LabelSet `json:"label_set"`
	ScoreSpace         string   `json:"score_space"`
	Status             string   `json:"status"`
	CostRule           string   `json:"cost_rule"`
	Assumptions        []string `json:"assumptions"`
}

// File is the document.
type File struct {
	Schema    int     `json:"schema"`
	Decisions []Entry `json:"decisions"`
}

func (e Entry) validate() error {
	switch {
	case e.QuestionID == "":
		return errors.New("entry requires a question_id")
	case e.Status != StatusProvisional && e.Status != StatusReady:
		return fmt.Errorf("%s: status %q must be provisional or ready", e.QuestionID, e.Status)
	case e.ScoreSpace != SpaceRaw && e.ScoreSpace != SpacePlatt:
		return fmt.Errorf("%s: score_space %q must be raw or platt", e.QuestionID, e.ScoreSpace)
	case !unit(e.MinPrecision) || !unit(e.MinRecall) || e.MinPrecision == 0 || e.MinRecall == 0:
		return fmt.Errorf("%s: minimums must be in (0, 1]", e.QuestionID)
	case e.Threshold != nil && !unit(*e.Threshold):
		return fmt.Errorf("%s: threshold must be in [0, 1]", e.QuestionID)
	case e.Gating && e.Threshold == nil:
		return fmt.Errorf("%s: a gating entry requires a threshold", e.QuestionID)
	case e.Status == StatusReady && (e.Threshold == nil || e.LabelSet.Summary == "" || e.LabelSet.Summary == "none"):
		return fmt.Errorf("%s: a ready entry requires a threshold and a label set", e.QuestionID)
	case e.Metric == "" || e.LabelSet.Summary == "" || e.CostRule == "" || e.MinimumsProvenance == "":
		return fmt.Errorf("%s: metric, label_set, cost_rule and minimums_provenance are required", e.QuestionID)
	}
	return nil
}

func unit(v float64) bool { return !math.IsNaN(v) && v >= 0 && v <= 1 }

func decode(raw []byte) (File, error) {
	var file File
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&file); err != nil {
		return File{}, err
	}
	if file.Schema != 1 {
		return File{}, fmt.Errorf("schema %d is not supported", file.Schema)
	}
	seen := map[string]bool{}
	for _, entry := range file.Decisions {
		if err := entry.validate(); err != nil {
			return File{}, err
		}
		if seen[entry.QuestionID] {
			return File{}, fmt.Errorf("%s: duplicate entry", entry.QuestionID)
		}
		seen[entry.QuestionID] = true
	}
	return file, nil
}

// Defaults returns the embedded entries.
func Defaults() (File, error) {
	file, err := decode(defaultsJSON)
	if err != nil {
		return File{}, fmt.Errorf("embedded thresholds: %w", err)
	}
	return file, nil
}

// Resolution is the threshold record in force for one question.
type Resolution struct {
	Entry Entry
	// Source is `default` or the override file's path.
	Source string
	// Stale is true when the entry was set against another version key.
	Stale bool
	// Status is the entry's status, forced to provisional when stale: a
	// threshold tuned for a question that has since changed is not ready.
	Status string
}

// Resolve returns the entry for questionID, bound to currentVersion. An entry
// in override (the raw override file; nil or empty when the repository has
// none) wins, and its version key must equal currentVersion or the resolution
// is stale. An absent override, or one that does not list the question, falls
// back to the default, which carries no pinned key and so is never stale. An
// invalid override is an error: it is configuration, not a guess.
func Resolve(override []byte, questionID, currentVersion string) (Resolution, error) {
	defaults, err := Defaults()
	if err != nil {
		return Resolution{}, err
	}
	if len(bytes.TrimSpace(override)) != 0 {
		file, err := decode(override)
		if err != nil {
			return Resolution{}, fmt.Errorf("%s: %w", Path, err)
		}
		for _, entry := range file.Decisions {
			if entry.QuestionID == questionID {
				return resolution(entry, Path, currentVersion), nil
			}
		}
	}
	for _, entry := range defaults.Decisions {
		if entry.QuestionID == questionID {
			entry.VersionKey = currentVersion
			return resolution(entry, "default", currentVersion), nil
		}
	}
	return Resolution{}, fmt.Errorf("no thresholds entry for question %q", questionID)
}

// ReadCommitted reads the override file as committed at ref (nil when the ref
// has no such file). It reads from git, never the working tree, for the same
// reason the policy is loaded from a ref: a pull request must not be able to
// change the thresholds it is ranked under.
func ReadCommitted(repoRoot, ref string) ([]byte, error) {
	listed, err := git(repoRoot, "ls-tree", "--name-only", ref, "--", Path)
	if err != nil {
		return nil, fmt.Errorf("%s at %s: %w", Path, ref, err)
	}
	if len(bytes.TrimSpace(listed)) == 0 {
		return nil, nil
	}
	data, err := git(repoRoot, "show", ref+":"+Path)
	if err != nil {
		return nil, fmt.Errorf("%s at %s: %w", Path, ref, err)
	}
	return data, nil
}

func git(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func resolution(entry Entry, source, currentVersion string) Resolution {
	r := Resolution{Entry: entry, Source: source, Stale: entry.VersionKey != currentVersion, Status: entry.Status}
	if r.Stale {
		r.Status = StatusProvisional
	}
	return r
}
