// Package eval loads adjudicated reference evidence and compares it with Scout observations.
package eval

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const CorpusSchemaV1 = "scout.eval.corpus/v1"

// CorpusLoadStatus describes whether a corpus was available and valid for use.
type CorpusLoadStatus string

const (
	CorpusAvailable  CorpusLoadStatus = "available"
	CorpusMissing    CorpusLoadStatus = "missing_corpus"
	CorpusUnreadable CorpusLoadStatus = "unreadable_corpus"
)

// AdjudicationStatus is the independently reviewed truth status of an issue.
type AdjudicationStatus string

const (
	AdjudicationValidatedTrue AdjudicationStatus = "validated_true"
	AdjudicationFalsePositive AdjudicationStatus = "false_positive"
	AdjudicationUnsupported   AdjudicationStatus = "unsupported"
	AdjudicationIncomplete    AdjudicationStatus = "incomplete"
	AdjudicationPendingReview AdjudicationStatus = "pending_review"
)

// ReferenceAssessmentStatus records whether a language/family input received a
// complete independent adjudication, including a clean negative result.
type ReferenceAssessmentStatus string

const (
	ReferenceAssessmentAdjudicated ReferenceAssessmentStatus = "adjudicated"
	ReferenceAssessmentIncomplete  ReferenceAssessmentStatus = "incomplete"
	ReferenceAssessmentUnsupported ReferenceAssessmentStatus = "unsupported"
)

// Corpus is a versioned collection of independently adjudicated reference outcomes.
type Corpus struct {
	SchemaVersion string        `json:"schemaVersion"`
	CorpusVersion string        `json:"corpusVersion"`
	Entries       []CorpusEntry `json:"entries"`
}

// CorpusEntry describes one target revision and one language/family input.
type CorpusEntry struct {
	ID                 string                    `json:"id"`
	Target             CorpusTarget              `json:"target"`
	Language           string                    `json:"language"`
	Family             string                    `json:"family"`
	Input              SupportedInput            `json:"input"`
	AdjudicationStatus ReferenceAssessmentStatus `json:"adjudicationStatus"`
	Issues             []ReferenceIssue          `json:"issues"`
}

// CorpusTarget identifies an immutable reference target.
type CorpusTarget struct {
	Repository string `json:"repository"`
	Revision   string `json:"revision"`
}

// SupportedInput records whether the reference assessment could inspect this input.
type SupportedInput struct {
	Supported bool   `json:"supported"`
	Reason    string `json:"reason,omitempty"`
}

// ReferenceIssue is an independently adjudicated expected outcome. It contains no
// scanner invocation or warning-count proxy.
type ReferenceIssue struct {
	ID           string             `json:"id"`
	Mechanism    string             `json:"mechanism"`
	Path         string             `json:"path"`
	Behavior     string             `json:"behavior,omitempty"`
	Adjudication AdjudicationStatus `json:"adjudication"`
	Evidence     string             `json:"evidence"`
}

// CorpusLoadResult separates a missing corpus from an empty but valid corpus.
type CorpusLoadResult struct {
	Status CorpusLoadStatus `json:"status"`
	Corpus *Corpus          `json:"corpus,omitempty"`
}

// LoadCorpus reads and validates a versioned JSON corpus. A nonexistent path is
// reported as missing evidence rather than a successful empty corpus.
func LoadCorpus(path string) (CorpusLoadResult, error) {
	if strings.TrimSpace(path) == "" {
		return CorpusLoadResult{Status: CorpusMissing}, nil
	}

	file, err := os.Open(filepath.Clean(path))
	if errors.Is(err, os.ErrNotExist) {
		return CorpusLoadResult{Status: CorpusMissing}, nil
	}
	if err != nil {
		return CorpusLoadResult{Status: CorpusUnreadable}, fmt.Errorf("open corpus: %w", err)
	}
	defer file.Close()

	corpus, err := decodeCorpus(file)
	if err != nil {
		return CorpusLoadResult{Status: CorpusUnreadable}, err
	}
	if err := corpus.Validate(); err != nil {
		return CorpusLoadResult{Status: CorpusUnreadable}, err
	}
	return CorpusLoadResult{Status: CorpusAvailable, Corpus: &corpus}, nil
}

func decodeCorpus(reader io.Reader) (Corpus, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	var corpus Corpus
	if err := decoder.Decode(&corpus); err != nil {
		return Corpus{}, fmt.Errorf("decode corpus: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return Corpus{}, errors.New("decode corpus: multiple JSON values")
		}
		return Corpus{}, fmt.Errorf("decode corpus trailing data: %w", err)
	}
	return corpus, nil
}

// Validate checks the minimum evidence needed to compare outcomes without
// converting unsupported or unadjudicated evidence into a passing result.
func (c Corpus) Validate() error {
	if c.SchemaVersion != CorpusSchemaV1 {
		return fmt.Errorf("unsupported corpus schema version %q", c.SchemaVersion)
	}
	if strings.TrimSpace(c.CorpusVersion) == "" {
		return errors.New("corpus version is required")
	}

	entryIDs := make(map[string]struct{}, len(c.Entries))
	selectionKeys := make(map[string]struct{}, len(c.Entries))
	for entryIndex, entry := range c.Entries {
		location := fmt.Sprintf("entry %d", entryIndex)
		if strings.TrimSpace(entry.ID) == "" {
			return fmt.Errorf("%s: id is required", location)
		}
		if _, duplicate := entryIDs[entry.ID]; duplicate {
			return fmt.Errorf("%s: duplicate id %q", location, entry.ID)
		}
		entryIDs[entry.ID] = struct{}{}
		if strings.TrimSpace(entry.Target.Repository) == "" || strings.TrimSpace(entry.Target.Revision) == "" {
			return fmt.Errorf("%s: pinned target repository and revision are required", location)
		}
		if strings.TrimSpace(entry.Language) == "" || strings.TrimSpace(entry.Family) == "" {
			return fmt.Errorf("%s: language and family are required", location)
		}
		selectionKey := entry.Target.Repository + "\x00" + entry.Target.Revision + "\x00" + entry.Language + "\x00" + entry.Family
		if _, duplicate := selectionKeys[selectionKey]; duplicate {
			return fmt.Errorf("%s: duplicate target revision, language, and family", location)
		}
		selectionKeys[selectionKey] = struct{}{}
		if !entry.Input.Supported && strings.TrimSpace(entry.Input.Reason) == "" {
			return fmt.Errorf("%s: unsupported input requires a reason", location)
		}
		switch entry.AdjudicationStatus {
		case ReferenceAssessmentAdjudicated, ReferenceAssessmentIncomplete, ReferenceAssessmentUnsupported:
		default:
			return fmt.Errorf("%s: unknown assessment adjudication status %q", location, entry.AdjudicationStatus)
		}
		if entry.AdjudicationStatus == ReferenceAssessmentUnsupported && entry.Input.Supported {
			return fmt.Errorf("%s: unsupported assessment requires unsupported input", location)
		}
		if entry.AdjudicationStatus != ReferenceAssessmentUnsupported && !entry.Input.Supported {
			return fmt.Errorf("%s: unsupported input requires unsupported assessment", location)
		}

		issueIDs := make(map[string]struct{}, len(entry.Issues))
		for issueIndex, issue := range entry.Issues {
			if err := issue.validate(fmt.Sprintf("%s issue %d", location, issueIndex)); err != nil {
				return err
			}
			if _, duplicate := issueIDs[issue.ID]; duplicate {
				return fmt.Errorf("%s: duplicate issue id %q", location, issue.ID)
			}
			issueIDs[issue.ID] = struct{}{}
		}
	}
	return nil
}

func (i ReferenceIssue) validate(location string) error {
	if strings.TrimSpace(i.ID) == "" {
		return fmt.Errorf("%s: id is required", location)
	}
	switch i.Adjudication {
	case AdjudicationValidatedTrue, AdjudicationFalsePositive, AdjudicationUnsupported, AdjudicationIncomplete, AdjudicationPendingReview:
	default:
		return fmt.Errorf("%s: unknown adjudication %q", location, i.Adjudication)
	}
	if i.Adjudication == AdjudicationValidatedTrue {
		if strings.TrimSpace(i.Mechanism) == "" || strings.TrimSpace(i.Path) == "" {
			return fmt.Errorf("%s: validated true issue requires mechanism and path", location)
		}
		if strings.TrimSpace(i.Evidence) == "" {
			return fmt.Errorf("%s: validated true issue requires evidence", location)
		}
	}
	return nil
}
