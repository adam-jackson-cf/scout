package inventory

import (
	"testing"

	"scout/internal/domain"
)

func TestDiffSchedulesDependentClosureForRecordedSharedContract(t *testing.T) {
	previous := domain.File{Path: "shared/contract.go", ContentDigest: "before", Eligible: true, OwnedFeatureIDs: []string{"producer"}}
	current := previous
	current.ContentDigest = "after"

	result, err := Diff(DiffInput{
		PreviousFiles:   []domain.File{previous},
		CurrentFiles:    []domain.File{current},
		PreviousFeatures: []domain.Feature{testFeature("producer", "shared/contract.go"), testFeature("consumer", "consumer/handler.go")},
		CurrentFeatures:  []domain.Feature{testFeature("producer", "shared/contract.go"), testFeature("consumer", "consumer/handler.go")},
		Dependencies: []DependencyEdge{{
			FromFeatureID: "producer",
			ToFeatureID:   "consumer",
			Reason:        "consumer implements the shared contract",
			InputPaths:    []string{"shared/contract.go"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !containsFeature(result.AffectedFeatureIDs, "consumer") {
		t.Fatalf("recorded dependency was not invalidated by a non-manifest shared contract: %+v", result.Invalidations)
	}
}

func TestDiffSchedulesChangedOwnedFileWhenCanonicalStatusIsUnchanged(t *testing.T) {
	previous := domain.File{Path: "service/handler.go", ContentDigest: "before", Eligible: true, OwnedFeatureIDs: []string{"feature"}}
	current := previous
	current.ContentDigest = "after"
	feature := testFeature("feature", "service/handler.go")
	feature.Status = "reviewed"

	result, err := Diff(DiffInput{
		PreviousFiles:    []domain.File{previous},
		CurrentFiles:     []domain.File{current},
		PreviousFeatures: []domain.Feature{feature},
		CurrentFeatures:  []domain.Feature{feature},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !containsFeature(result.AffectedFeatureIDs, feature.FeatureID) {
		t.Fatalf("content-only owned-file edit was not scheduled: %+v", result.Invalidations)
	}
}

func TestDiffClassifiesAddedRemovedAndRenamedFiles(t *testing.T) {
	previous := []domain.File{
		{Path: "removed.go", ContentDigest: "removed", Eligible: true, OwnedFeatureIDs: []string{"feature"}},
		{Path: "before.go", ContentDigest: "rename", Eligible: true, OwnedFeatureIDs: []string{"feature"}},
	}
	current := []domain.File{
		{Path: "added.go", ContentDigest: "added", Eligible: true, OwnedFeatureIDs: []string{"feature"}},
		{Path: "after.go", ContentDigest: "rename", Eligible: true, OwnedFeatureIDs: []string{"feature"}},
	}
	result, err := Diff(DiffInput{PreviousFiles: previous, CurrentFiles: current, PreviousFeatures: []domain.Feature{testFeature("feature", "before.go")}, CurrentFeatures: []domain.Feature{testFeature("feature", "after.go")}})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Added) != 1 || result.Added[0].Path != "added.go" {
		t.Fatalf("added file was not classified: %+v", result.Added)
	}
	if len(result.Removed) != 1 || result.Removed[0].Path != "removed.go" {
		t.Fatalf("removed file was not classified: %+v", result.Removed)
	}
	if len(result.Renamed) != 1 || result.Renamed[0].Path != "after.go" || result.Renamed[0].PreviousPath != "before.go" {
		t.Fatalf("same-content rename was not classified: %+v", result.Renamed)
	}
	if !containsFeature(result.AffectedFeatureIDs, "feature") {
		t.Fatalf("file delta did not invalidate the owning feature: %+v", result.Invalidations)
	}
}

func TestLinkFeaturesReportsUnownedEligibleFileWithoutSynthesizingFeature(t *testing.T) {
	files, gaps, err := LinkFeatures([]domain.File{
		{Path: "mapped.go", ContentDigest: "mapped", Eligible: true},
		{Path: "unowned.go", ContentDigest: "unowned", Eligible: true},
	}, []domain.Feature{testFeature("feature", "mapped.go")})
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if file.Path == "unowned.go" && len(file.OwnedFeatureIDs) != 0 {
			t.Fatalf("unowned file received a synthetic owner: %+v", file)
		}
	}
	if !hasGap(gaps, "unowned.go", "") {
		t.Fatalf("unowned eligible file was not reported: %+v", gaps)
	}
}

func testFeature(id, ownedPath string) domain.Feature {
	return domain.Feature{
		SchemaVersion: 1,
		FeatureID:     id,
		Kind:          "service",
		Source:        "heuristic",
		Confidence:    domain.ConfidenceHigh,
		OwnedFiles:    []domain.FeatureFileRef{{Path: ownedPath, Reason: "test ownership"}},
		RawOriginal:   []byte(`{"schemaVersion":1,"featureId":"` + id + `","kind":"service","source":"heuristic"}`),
	}
}

func containsFeature(ids []string, wanted string) bool {
	for _, id := range ids {
		if id == wanted {
			return true
		}
	}
	return false
}

func hasGap(gaps []MappingGap, path, featureID string) bool {
	for _, gap := range gaps {
		if gap.Path == path && gap.FeatureID == featureID {
			return true
		}
	}
	return false
}
