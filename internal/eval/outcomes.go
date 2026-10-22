package eval

import (
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"
	"unicode"
)

// EvaluationStatus states whether this comparison had enough evidence to assess
// the requested target input.
type EvaluationStatus string

const (
	EvaluationComplete         EvaluationStatus = "complete"
	EvaluationMissingCorpus    EvaluationStatus = "missing_corpus"
	EvaluationUnsupportedInput EvaluationStatus = "unsupported_input"
	EvaluationIncomplete       EvaluationStatus = "incomplete"
)

// ReferenceParityStatus never represents success based on invocation counts or
// unvalidated observations.
type ReferenceParityStatus string

const (
	ParityNotEvaluated ReferenceParityStatus = "not_evaluated"
	ParityIncomplete   ReferenceParityStatus = "incomplete"
	ParityFailed       ReferenceParityStatus = "failed"
	ParityMet          ReferenceParityStatus = "met"
)

// AssessmentEvidenceStatus states whether Scout completed the corresponding
// assessment input. It is independent from whether the corpus was adjudicated.
type AssessmentEvidenceStatus string

const (
	AssessmentEvidenceComplete   AssessmentEvidenceStatus = "complete"
	AssessmentEvidenceIncomplete AssessmentEvidenceStatus = "incomplete"
)

// AssessmentEvidence identifies the coverage state for the observations being
// compared. A missing reason for incomplete assessment is invalid.
type AssessmentEvidence struct {
	Status AssessmentEvidenceStatus `json:"status"`
	Reason string                   `json:"reason,omitempty"`
}

// EvaluationRequest selects one immutable target revision and language/family
// combination from a corpus.
type EvaluationRequest struct {
	Target       CorpusTarget       `json:"target"`
	Language     string             `json:"language"`
	Family       string             `json:"family"`
	Assessment   AssessmentEvidence `json:"assessment"`
	Observations []ScoutObservation `json:"observations"`
}

// ScoutObservation is a safe, evidence-backed Scout issue observation. The
// stable issue key joins observations already correlated by Scout; it is not a
// replacement for mechanism/path matching.
type ScoutObservation struct {
	ID             string             `json:"id"`
	StableIssueKey string             `json:"stableIssueKey,omitempty"`
	Source         string             `json:"source"`
	Mechanism      string             `json:"mechanism"`
	Path           string             `json:"path"`
	Behavior       string             `json:"behavior,omitempty"`
	Evidence       string             `json:"evidence,omitempty"`
	Validated      bool               `json:"validated"`
	Adjudication   AdjudicationStatus `json:"adjudication"`
}

// IssueMatch links one independently validated reference issue to one Scout
// issue group. Every linked source observation remains visible.
type IssueMatch struct {
	ReferenceIssueID string   `json:"referenceIssueId"`
	ScoutIssueKey    string   `json:"scoutIssueKey"`
	ObservationIDs   []string `json:"observationIds"`
	MatchBasis       string   `json:"matchBasis"`
}

// AmbiguousMatch remains unmerged because multiple Scout issue groups are
// equally relevant to the same reference issue.
type AmbiguousMatch struct {
	ReferenceIssueID string   `json:"referenceIssueId"`
	ScoutIssueKeys   []string `json:"scoutIssueKeys"`
}

// OutcomeReport records reference misses, Scout extras, and confirmed false
// positives independently. Extras and false positives never satisfy a miss.
type OutcomeReport struct {
	Status         EvaluationStatus      `json:"status"`
	Parity         ReferenceParityStatus `json:"parity"`
	Reason         string                `json:"reason,omitempty"`
	CorpusEntryID  string                `json:"corpusEntryId,omitempty"`
	Matches        []IssueMatch          `json:"matches,omitempty"`
	Misses         []ReferenceIssue      `json:"misses,omitempty"`
	Extras         []ScoutObservation    `json:"extras,omitempty"`
	FalsePositives []ScoutObservation    `json:"falsePositives,omitempty"`
	Unvalidated    []ScoutObservation    `json:"unvalidated,omitempty"`
	Ambiguous      []AmbiguousMatch      `json:"ambiguous,omitempty"`
}

// CompareOutcomes compares evidence-backed Scout observations with the validated
// true issues in the selected reference corpus entry. It does not infer success
// from tool execution, warning totals, or unvalidated observations.
func CompareOutcomes(corpus CorpusLoadResult, request EvaluationRequest) (OutcomeReport, error) {
	if err := validateEvaluationRequest(request); err != nil {
		return OutcomeReport{}, err
	}
	if corpus.Status == CorpusUnreadable {
		return OutcomeReport{
			Status: EvaluationIncomplete,
			Parity: ParityIncomplete,
			Reason: "reference corpus is unreadable",
		}, nil
	}
	if corpus.Status != CorpusAvailable || corpus.Corpus == nil {
		return OutcomeReport{
			Status: EvaluationMissingCorpus,
			Parity: ParityNotEvaluated,
			Reason: "reference corpus is unavailable",
		}, nil
	}
	if err := corpus.Corpus.Validate(); err != nil {
		return OutcomeReport{}, fmt.Errorf("validate corpus before comparison: %w", err)
	}

	entry := selectCorpusEntry(*corpus.Corpus, request)
	if entry == nil {
		return OutcomeReport{
			Status: EvaluationMissingCorpus,
			Parity: ParityNotEvaluated,
			Reason: "no corpus entry for the requested pinned target revision, language, and family",
		}, nil
	}
	if !entry.Input.Supported || entry.AdjudicationStatus == ReferenceAssessmentUnsupported {
		return OutcomeReport{
			Status:        EvaluationUnsupportedInput,
			Parity:        ParityNotEvaluated,
			Reason:        entry.Input.Reason,
			CorpusEntryID: entry.ID,
		}, nil
	}

	report := compareEntry(*entry, request.Observations)
	if request.Assessment.Status == AssessmentEvidenceIncomplete {
		report.Status = EvaluationIncomplete
		report.Parity = ParityIncomplete
		report.Reason = request.Assessment.Reason
		return report, nil
	}
	if entry.AdjudicationStatus != ReferenceAssessmentAdjudicated {
		report.Status = EvaluationIncomplete
		report.Parity = ParityIncomplete
		report.Reason = "reference corpus entry is not completely adjudicated"
		return report, nil
	}
	if len(report.Ambiguous) != 0 {
		report.Status = EvaluationIncomplete
		report.Parity = ParityIncomplete
		report.Reason = "ambiguous mechanism/path matches require adjudication"
		return report, nil
	}
	if len(report.Misses) != 0 {
		report.Status = EvaluationComplete
		report.Parity = ParityFailed
		return report, nil
	}
	report.Status = EvaluationComplete
	report.Parity = ParityMet
	return report, nil
}

func validateEvaluationRequest(request EvaluationRequest) error {
	if strings.TrimSpace(request.Target.Repository) == "" || strings.TrimSpace(request.Target.Revision) == "" {
		return errors.New("evaluation target repository and pinned revision are required")
	}
	if strings.TrimSpace(request.Language) == "" || strings.TrimSpace(request.Family) == "" {
		return errors.New("evaluation language and family are required")
	}
	switch request.Assessment.Status {
	case AssessmentEvidenceComplete:
	case AssessmentEvidenceIncomplete:
		if strings.TrimSpace(request.Assessment.Reason) == "" {
			return errors.New("incomplete assessment requires a reason")
		}
	default:
		return fmt.Errorf("unknown assessment status %q", request.Assessment.Status)
	}
	for index, observation := range request.Observations {
		if err := observation.validate(fmt.Sprintf("observation %d", index)); err != nil {
			return err
		}
	}
	return nil
}

func (o ScoutObservation) validate(location string) error {
	if strings.TrimSpace(o.ID) == "" || strings.TrimSpace(o.Source) == "" {
		return fmt.Errorf("%s: id and source are required", location)
	}
	switch o.Adjudication {
	case AdjudicationValidatedTrue, AdjudicationFalsePositive, AdjudicationUnsupported, AdjudicationIncomplete, AdjudicationPendingReview:
	default:
		return fmt.Errorf("%s: unknown adjudication %q", location, o.Adjudication)
	}
	if o.Validated && o.Adjudication == AdjudicationValidatedTrue {
		if strings.TrimSpace(o.Mechanism) == "" || strings.TrimSpace(o.Path) == "" || strings.TrimSpace(o.Evidence) == "" {
			return fmt.Errorf("%s: validated true observation requires mechanism, path, and evidence", location)
		}
	}
	return nil
}

func selectCorpusEntry(corpus Corpus, request EvaluationRequest) *CorpusEntry {
	for index := range corpus.Entries {
		entry := &corpus.Entries[index]
		if entry.Target == request.Target && entry.Language == request.Language && entry.Family == request.Family {
			return entry
		}
	}
	return nil
}

func compareEntry(entry CorpusEntry, observations []ScoutObservation) OutcomeReport {
	report := OutcomeReport{CorpusEntryID: entry.ID}
	groups, falsePositives, unvalidated := groupObservations(observations)
	report.FalsePositives = falsePositives
	report.Unvalidated = unvalidated

	references := validatedReferenceIssues(entry.Issues)
	claimed := make(map[int]struct{}, len(groups))
	for _, reference := range references {
		best, ties := bestMatch(reference, groups, claimed)
		if best < 0 {
			report.Misses = append(report.Misses, reference)
			continue
		}
		if len(ties) > 1 {
			keys := make([]string, 0, len(ties))
			for _, index := range ties {
				keys = append(keys, groups[index].key)
			}
			sort.Strings(keys)
			report.Ambiguous = append(report.Ambiguous, AmbiguousMatch{ReferenceIssueID: reference.ID, ScoutIssueKeys: keys})
			continue
		}
		claimed[best] = struct{}{}
		group := groups[best]
		report.Matches = append(report.Matches, IssueMatch{
			ReferenceIssueID: reference.ID,
			ScoutIssueKey:    group.key,
			ObservationIDs:   group.observationIDs(),
			MatchBasis:       matchBasis(reference, group.representative),
		})
	}
	for index, group := range groups {
		if _, matched := claimed[index]; !matched {
			report.Extras = append(report.Extras, group.representative)
		}
	}
	return report
}

type observationGroup struct {
	key            string
	representative ScoutObservation
	observations   []ScoutObservation
}

func (g observationGroup) observationIDs() []string {
	ids := make([]string, 0, len(g.observations))
	for _, observation := range g.observations {
		ids = append(ids, observation.ID)
	}
	sort.Strings(ids)
	return ids
}

func groupObservations(observations []ScoutObservation) ([]observationGroup, []ScoutObservation, []ScoutObservation) {
	groupsByKey := make(map[string]int)
	groupsByEvidence := make(map[string]int)
	groups := make([]observationGroup, 0, len(observations))
	falsePositives := make([]ScoutObservation, 0)
	unvalidated := make([]ScoutObservation, 0)
	for _, observation := range observations {
		if observation.Adjudication == AdjudicationFalsePositive {
			falsePositives = append(falsePositives, observation)
			continue
		}
		if !observation.Validated || observation.Adjudication != AdjudicationValidatedTrue {
			unvalidated = append(unvalidated, observation)
			continue
		}
		key := observationGroupKey(observation)
		evidenceKey := observationEvidenceKey(observation)
		if index, ok := groupsByKey[key]; ok {
			groups[index].observations = append(groups[index].observations, observation)
			continue
		}
		if index, ok := groupsByEvidence[evidenceKey]; ok {
			groups[index].observations = append(groups[index].observations, observation)
			groupsByKey[key] = index
			continue
		}
		index := len(groups)
		groupsByKey[key] = index
		groupsByEvidence[evidenceKey] = index
		groups = append(groups, observationGroup{key: key, representative: observation, observations: []ScoutObservation{observation}})
	}
	return groups, falsePositives, unvalidated
}

func observationGroupKey(observation ScoutObservation) string {
	if key := normalizeText(observation.StableIssueKey); key != "" {
		return "issue:" + key
	}
	return observationEvidenceKey(observation)
}

func observationEvidenceKey(observation ScoutObservation) string {
	return "evidence:" + normalizeText(observation.Mechanism) + "\x00" + normalizePath(observation.Path) + "\x00" + normalizeText(observation.Behavior)
}

func validatedReferenceIssues(issues []ReferenceIssue) []ReferenceIssue {
	validated := make([]ReferenceIssue, 0, len(issues))
	for _, issue := range issues {
		if issue.Adjudication == AdjudicationValidatedTrue {
			validated = append(validated, issue)
		}
	}
	return validated
}

func bestMatch(reference ReferenceIssue, groups []observationGroup, claimed map[int]struct{}) (int, []int) {
	best := -1
	bestScore := 0
	ties := make([]int, 0)
	for index, group := range groups {
		if _, used := claimed[index]; used {
			continue
		}
		score := matchScore(reference, group.representative)
		if score == 0 {
			continue
		}
		if score > bestScore {
			best, bestScore = index, score
			ties = []int{index}
			continue
		}
		if score == bestScore {
			ties = append(ties, index)
		}
	}
	return best, ties
}

func matchScore(reference ReferenceIssue, observation ScoutObservation) int {
	if normalizeText(reference.Mechanism) != normalizeText(observation.Mechanism) {
		return 0
	}
	if pathsRelevant(reference.Path, observation.Path) {
		if behaviorsRelevant(reference.Behavior, observation.Behavior) {
			return 3
		}
		return 2
	}
	if behaviorsRelevant(reference.Behavior, observation.Behavior) {
		return 1
	}
	return 0
}

func matchBasis(reference ReferenceIssue, observation ScoutObservation) string {
	if pathsRelevant(reference.Path, observation.Path) && behaviorsRelevant(reference.Behavior, observation.Behavior) {
		return "normalized_mechanism_path_behavior"
	}
	if pathsRelevant(reference.Path, observation.Path) {
		return "normalized_mechanism_path"
	}
	return "normalized_mechanism_behavior"
}

func pathsRelevant(left, right string) bool {
	left, right = normalizePath(left), normalizePath(right)
	return left != "" && right != "" && left == right
}

func behaviorsRelevant(left, right string) bool {
	left, right = normalizeText(left), normalizeText(right)
	if left == "" || right == "" {
		return false
	}
	return left == right
}

func normalizePath(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	if value == "" {
		return ""
	}
	value = path.Clean(value)
	value = strings.TrimPrefix(value, "./")
	if value == "." {
		return ""
	}
	return strings.ToLower(value)
}

func normalizeText(value string) string {
	var builder strings.Builder
	previousSpace := true
	for _, character := range strings.ToLower(strings.TrimSpace(value)) {
		if unicode.IsLetter(character) || unicode.IsDigit(character) {
			builder.WriteRune(character)
			previousSpace = false
			continue
		}
		if !previousSpace {
			builder.WriteByte(' ')
			previousSpace = true
		}
	}
	return strings.TrimSpace(builder.String())
}
