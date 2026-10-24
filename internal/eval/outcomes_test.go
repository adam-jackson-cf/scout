package eval

import (
	"reflect"
	"testing"
)

var (
	testTarget = CorpusTarget{Repository: "example/repo", Revision: "0123abc"}
	trueIssue  = ReferenceIssue{ID: "ref-1", Adjudication: AdjudicationValidatedTrue, Mechanism: "Mutable Action ref", Path: ".github/workflows/ci.yml", Behavior: "uses a tag", Evidence: "reviewed"}
)

func corpusWith(status ReferenceAssessmentStatus, supported bool, issues ...ReferenceIssue) CorpusLoadResult {
	entry := CorpusEntry{
		ID: "entry-1", Target: testTarget, Language: "yaml", Family: "ci",
		Input: SupportedInput{Supported: supported}, AdjudicationStatus: status, Issues: issues,
	}
	if !supported {
		entry.Input.Reason = "no scanner for this input"
	}
	return CorpusLoadResult{Status: CorpusAvailable, Corpus: &Corpus{SchemaVersion: CorpusSchemaV1, CorpusVersion: "1", Entries: []CorpusEntry{entry}}}
}

func request(observations ...ScoutObservation) EvaluationRequest {
	return EvaluationRequest{
		Target: testTarget, Language: "yaml", Family: "ci",
		Assessment:   AssessmentEvidence{Status: AssessmentEvidenceComplete},
		Observations: observations,
	}
}

func observation(id, mechanism, path string) ScoutObservation {
	return ScoutObservation{ID: id, Source: "scanner", Mechanism: mechanism, Path: path, Evidence: "e", Validated: true, Adjudication: AdjudicationValidatedTrue}
}

func compare(t *testing.T, corpus CorpusLoadResult, req EvaluationRequest) OutcomeReport {
	t.Helper()
	report, err := CompareOutcomes(corpus, req)
	if err != nil {
		t.Fatal(err)
	}
	return report
}

func TestCompareOutcomesScenarios(t *testing.T) {
	adjudicated := corpusWith(ReferenceAssessmentAdjudicated, true, trueIssue)
	match := observation("o1", "mutable action ref", ".GITHUB\\workflows/./ci.yml")

	t.Run("a normalised mechanism and path match meets parity", func(t *testing.T) {
		report := compare(t, adjudicated, request(match))
		if report.Status != EvaluationComplete || report.Parity != ParityMet || len(report.Matches) != 1 || len(report.Misses) != 0 || len(report.Extras) != 0 {
			t.Fatalf("report %+v", report)
		}
	})
	t.Run("a missed reference issue fails parity and stays a miss", func(t *testing.T) {
		report := compare(t, adjudicated, request())
		if report.Parity != ParityFailed || len(report.Misses) != 1 || report.Misses[0].ID != "ref-1" {
			t.Fatalf("report %+v", report)
		}
	})
	t.Run("an extra never satisfies a miss", func(t *testing.T) {
		report := compare(t, adjudicated, request(observation("o2", "other mechanism", "a.go")))
		if report.Parity != ParityFailed || len(report.Misses) != 1 || len(report.Extras) != 1 {
			t.Fatalf("report %+v", report)
		}
	})
	t.Run("a confirmed false positive and an unvalidated observation are kept apart", func(t *testing.T) {
		fp := observation("fp", "mutable action ref", trueIssue.Path)
		fp.Adjudication = AdjudicationFalsePositive
		pending := observation("pending", "mutable action ref", trueIssue.Path)
		pending.Validated, pending.Adjudication = false, AdjudicationPendingReview
		report := compare(t, adjudicated, request(fp, pending))
		if report.Parity != ParityFailed || len(report.FalsePositives) != 1 || len(report.Unvalidated) != 1 || len(report.Matches) != 0 || len(report.Extras) != 0 {
			t.Fatalf("report %+v", report)
		}
	})
	t.Run("observations sharing a stable key merge into one match", func(t *testing.T) {
		a, b := match, observation("o3", "mutable action ref", "elsewhere.yml")
		a.StableIssueKey, b.StableIssueKey = "issue-9", "issue-9"
		report := compare(t, adjudicated, request(a, b))
		if len(report.Matches) != 1 || !reflect.DeepEqual(report.Matches[0].ObservationIDs, []string{"o1", "o3"}) || len(report.Extras) != 0 {
			t.Fatalf("report %+v", report)
		}
	})
	t.Run("a mechanism and behaviour match on another path scores below a path match", func(t *testing.T) {
		reference := trueIssue
		pathMatch := observation("path", "mutable action ref", trueIssue.Path)
		behaviourOnly := observation("behaviour", "mutable action ref", "other.yml")
		behaviourOnly.Behavior = "uses a tag"
		report := compare(t, corpusWith(ReferenceAssessmentAdjudicated, true, reference), request(behaviourOnly, pathMatch))
		if len(report.Matches) != 1 || report.Matches[0].ObservationIDs[0] != "path" || len(report.Extras) != 1 || report.Extras[0].ID != "behaviour" {
			t.Fatalf("report %+v", report)
		}
		only := compare(t, adjudicated, request(behaviourOnly))
		if len(only.Matches) != 1 || only.Matches[0].MatchBasis != "normalized_mechanism_behavior" {
			t.Fatalf("behaviour-only report %+v", only)
		}
	})
	t.Run("equally relevant groups are ambiguous and make the run incomplete", func(t *testing.T) {
		a := observation("a", "mutable action ref", trueIssue.Path)
		a.Behavior = "different one"
		b := observation("b", "mutable action ref", trueIssue.Path)
		b.Behavior = "different two"
		report := compare(t, adjudicated, request(a, b))
		if report.Status != EvaluationIncomplete || report.Parity != ParityIncomplete || len(report.Ambiguous) != 1 || len(report.Ambiguous[0].ScoutIssueKeys) != 2 || len(report.Matches) != 0 {
			t.Fatalf("report %+v", report)
		}
	})
	t.Run("a group claimed by one reference cannot satisfy another", func(t *testing.T) {
		second := trueIssue
		second.ID = "ref-2"
		report := compare(t, corpusWith(ReferenceAssessmentAdjudicated, true, trueIssue, second), request(match))
		if len(report.Matches) != 1 || len(report.Misses) != 1 || report.Parity != ParityFailed {
			t.Fatalf("report %+v", report)
		}
	})
	t.Run("an incomplete Scout assessment is incomplete even when everything matches", func(t *testing.T) {
		req := request(match)
		req.Assessment = AssessmentEvidence{Status: AssessmentEvidenceIncomplete, Reason: "lane timed out"}
		report := compare(t, adjudicated, req)
		if report.Status != EvaluationIncomplete || report.Parity != ParityIncomplete || report.Reason != "lane timed out" {
			t.Fatalf("report %+v", report)
		}
	})
	t.Run("an incompletely adjudicated entry never meets parity", func(t *testing.T) {
		report := compare(t, corpusWith(ReferenceAssessmentIncomplete, true, trueIssue), request(match))
		if report.Status != EvaluationIncomplete || report.Parity != ParityIncomplete {
			t.Fatalf("report %+v", report)
		}
	})
	t.Run("an unsupported input is not evaluated", func(t *testing.T) {
		report := compare(t, corpusWith(ReferenceAssessmentUnsupported, false), request())
		if report.Status != EvaluationUnsupportedInput || report.Parity != ParityNotEvaluated || report.Reason != "no scanner for this input" {
			t.Fatalf("report %+v", report)
		}
	})
	t.Run("a clean adjudicated negative with no observations meets parity", func(t *testing.T) {
		report := compare(t, corpusWith(ReferenceAssessmentAdjudicated, true), request())
		if report.Parity != ParityMet {
			t.Fatalf("report %+v", report)
		}
	})
	t.Run("a missing, unreadable or non-matching corpus is never a pass", func(t *testing.T) {
		missing := compare(t, CorpusLoadResult{Status: CorpusMissing}, request(match))
		unreadable := compare(t, CorpusLoadResult{Status: CorpusUnreadable}, request(match))
		other := request(match)
		other.Target.Revision = "different"
		unknownTarget := compare(t, adjudicated, other)
		if missing.Status != EvaluationMissingCorpus || unreadable.Status != EvaluationIncomplete || unknownTarget.Status != EvaluationMissingCorpus {
			t.Fatalf("missing %+v unreadable %+v unknown %+v", missing, unreadable, unknownTarget)
		}
		for _, report := range []OutcomeReport{missing, unreadable, unknownTarget} {
			if report.Parity == ParityMet || report.Parity == ParityFailed {
				t.Fatalf("parity %q must not be decided without a corpus entry", report.Parity)
			}
		}
	})
}

func TestCompareOutcomesRejectsInvalidRequests(t *testing.T) {
	corpus := corpusWith(ReferenceAssessmentAdjudicated, true, trueIssue)
	valid := observation("o1", "m", "p")
	cases := map[string]func(*EvaluationRequest){
		"no repository":            func(r *EvaluationRequest) { r.Target.Repository = "" },
		"no revision":              func(r *EvaluationRequest) { r.Target.Revision = " " },
		"no language":              func(r *EvaluationRequest) { r.Language = "" },
		"no family":                func(r *EvaluationRequest) { r.Family = "" },
		"unknown assessment":       func(r *EvaluationRequest) { r.Assessment.Status = "maybe" },
		"incomplete without cause": func(r *EvaluationRequest) { r.Assessment = AssessmentEvidence{Status: AssessmentEvidenceIncomplete} },
		"observation without id": func(r *EvaluationRequest) {
			r.Observations = []ScoutObservation{{Source: "s", Adjudication: AdjudicationPendingReview}}
		},
		"observation unknown adjudication": func(r *EvaluationRequest) {
			o := valid
			o.Adjudication = "guessed"
			r.Observations = []ScoutObservation{o}
		},
		"validated observation without evidence": func(r *EvaluationRequest) {
			o := valid
			o.Evidence = ""
			r.Observations = []ScoutObservation{o}
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			req := request(valid)
			mutate(&req)
			if _, err := CompareOutcomes(corpus, req); err == nil {
				t.Fatal("an invalid request must be an error, not a report")
			}
		})
	}
}
