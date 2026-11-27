package association

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"scout/internal/domain"
)

func readReplay[T any](t *testing.T, path string) []T {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}()
	var out []T
	s := bufio.NewScanner(f)
	for s.Scan() {
		var row T
		if err := json.Unmarshal(s.Bytes(), &row); err != nil {
			t.Fatal(err)
		}
		out = append(out, row)
	}
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

type replayUnit struct {
	CaseID    string `json:"case_id"`
	ClusterID string `json:"cluster_id"`
}

type replayGold struct {
	CaseID     string   `json:"case_id"`
	Covered    bool     `json:"covered"`
	Acceptable []string `json:"acceptable"`
}

type replayScore struct {
	Type, Question string
	CaseID         string `json:"case_id"`
	Perm           int
	PTrue          *float64 `json:"p_true"`
	Choice         string
	Confidence     float64
}

type replayTally struct {
	correct, assigned, unresolved int
}

func replayScores(scores []replayScore, caseID string) Scores {
	var in Scores
	byPerm := map[int]ChoiceScore{}
	for _, s := range scores {
		if s.CaseID != caseID || s.Type != "answer" {
			continue
		}
		switch s.Question {
		case "merge.covered":
			in.Covered = s.PTrue
		case "merge.which_area":
			byPerm[s.Perm] = ChoiceScore{AreaID: s.Choice, Confidence: s.Confidence}
		}
	}
	// Preserve permutation position; an absent required answer remains missing.
	if p, ok := byPerm[1]; ok {
		in.Choices = append(in.Choices, p)
		if p, ok := byPerm[2]; ok {
			in.Choices = append(in.Choices, p)
		}
	}
	return in
}

func (tally *replayTally) score(got Composed, g replayGold) {
	if got.Outcome == "assigned" {
		tally.assigned++
		if g.Covered && slices.Contains(g.Acceptable, got.AreaID) {
			tally.correct++
		}
	}
	if got.Outcome == "unresolved" {
		tally.unresolved++
	}
}

func replayRound(t *testing.T, round int) (replayTally, int) {
	t.Helper()
	fixture := func(kind string) string {
		return filepath.Join("testdata", fmt.Sprintf("replay-r%d-%s.jsonl", round, kind))
	}
	units := readReplay[replayUnit](t, fixture("units"))
	gold := readReplay[replayGold](t, fixture("gold"))
	scores := readReplay[replayScore](t, fixture("scores"))
	var tally replayTally
	for _, u := range units {
		got := Compose(u.ClusterID, replayScores(scores, u.CaseID), .60, .535)
		i := slices.IndexFunc(gold, func(g replayGold) bool { return g.CaseID == u.CaseID })
		if i < 0 {
			t.Fatalf("fixture lacks gold for %s", u.CaseID)
		}
		tally.score(got, gold[i])
		switch got.Outcome {
		case "assigned", "none", "unresolved":
		default:
			t.Fatalf("unknown output %+v", got)
		}
	}
	return tally, len(units)
}

// P1 counts assignments against any acceptable gold area, not only primary.
// As in score_outcome/counts_for, none is resolved, not an abstention.
func TestAssociationReplay(t *testing.T) {
	totalCorrect, totalAssigned, totalUnresolved, totalN := 0, 0, 0, 0
	for _, want := range []struct {
		name                                    string
		round, correct, assigned, unresolved, n int
	}{{"round 8", 8, 14, 14, 5, 21}, {"round 9", 9, 8, 10, 1, 23}, {"round 10", 10, 42, 43, 4, 48}} {
		t.Run(want.name, func(t *testing.T) {
			tally, n := replayRound(t, want.round)
			correct, assigned, unresolved := tally.correct, tally.assigned, tally.unresolved
			if correct != want.correct || assigned != want.assigned || unresolved != want.unresolved || n != want.n {
				t.Fatalf("got correct/assigned %d/%d unresolved/n %d/%d; want %+v", correct, assigned, unresolved, n, want)
			}
			totalCorrect += correct
			totalAssigned += assigned
			totalUnresolved += unresolved
			totalN += n
		})
	}
	if totalCorrect != 64 || totalAssigned != 67 || totalUnresolved != 10 || totalN != 92 {
		t.Fatalf("pooled %d/%d %d/%d", totalCorrect, totalAssigned, totalUnresolved, totalN)
	}
}

func choiceScores(values ...ChoiceScore) []ChoiceScore {
	return values
}

func TestAssociationComposeBranches(t *testing.T) {
	// Pin selection independently of Compose: these ids' full SHA256 mod5 values
	// are checked in TestAssociationFlipSelection.
	one, two := "cluster-0", "cluster-13"
	for _, tc := range []struct {
		name, id             string
		scores               Scores
		output, area, reason string
		confidence           float64
	}{
		{name: "covered missing", id: one, output: "unresolved", reason: domain.AssociationReasonJevFailure},
		{name: "covered failed", id: one, scores: Scores{FailureReason: domain.AssociationReasonJevFailure}, output: "unresolved", reason: domain.AssociationReasonJevFailure},
		{name: "cap", id: one, scores: Scores{FailureReason: domain.AssociationReasonJevCap}, output: "unresolved", reason: domain.AssociationReasonJevCap},
		{name: "covered tie none", id: two, scores: Scores{Covered: new(.40)}, output: "none", confidence: .60},
		{name: "covered tolerance none", id: one, scores: Scores{Covered: new(.4000000000005)}, output: "none", confidence: .5999999999995},
		{name: "covered above tie requires choice", id: one, scores: Scores{Covered: new(.400000000002)}, output: "unresolved", reason: domain.AssociationReasonJevFailure},
		{name: "first permutation absent", id: one, scores: Scores{Covered: new(.9)}, output: "unresolved", reason: domain.AssociationReasonJevFailure},
		{name: "second permutation absent", id: two, scores: Scores{Covered: new(.9), Choices: choiceScores(ChoiceScore{AreaID: "a", Confidence: .9})}, output: "unresolved", reason: domain.AssociationReasonJevFailure},
		{name: "disagreement", id: two, scores: Scores{Covered: new(.9), Choices: choiceScores(ChoiceScore{AreaID: "a", Confidence: .99}, ChoiceScore{AreaID: "b", Confidence: .99})}, output: "unresolved", reason: domain.AssociationReasonPermutationNo},
		{name: "choice tie assigned", id: one, scores: Scores{Covered: new(.9), Choices: choiceScores(ChoiceScore{AreaID: "a", Confidence: .535})}, output: "assigned", area: "a", confidence: .535},
		{name: "mean tie assigned", id: two, scores: Scores{Covered: new(.9), Choices: choiceScores(ChoiceScore{AreaID: "a", Confidence: .535}, ChoiceScore{AreaID: "a", Confidence: .535})}, output: "assigned", area: "a", confidence: .535},
		{name: "mean above threshold", id: two, scores: Scores{Covered: new(.9), Choices: choiceScores(ChoiceScore{AreaID: "a", Confidence: .7}, ChoiceScore{AreaID: "a", Confidence: .4})}, output: "assigned", area: "a", confidence: .55},
		{name: "mean below threshold", id: two, scores: Scores{Covered: new(.9), Choices: choiceScores(ChoiceScore{AreaID: "a", Confidence: .9}, ChoiceScore{AreaID: "a", Confidence: .1})}, output: "unresolved", reason: domain.AssociationReasonBelowTChoice, confidence: .5},
		{name: "single below threshold", id: one, scores: Scores{Covered: new(.9), Choices: choiceScores(ChoiceScore{AreaID: "a", Confidence: .534999})}, output: "unresolved", reason: domain.AssociationReasonBelowTChoice, confidence: .534999},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := Compose(tc.id, tc.scores, .60, .535)
			if got.Outcome != tc.output || got.AreaID != tc.area || got.Reason != tc.reason || math.Abs(got.Confidence-tc.confidence) > 1e-12 {
				t.Fatalf("got %+v want %+v", got, tc)
			}
		})
	}
}

func TestAssociationFlipSelection(t *testing.T) {
	for _, tc := range []struct {
		name, id string
		want     bool
	}{{name: "excluded cluster 0", id: "cluster-0", want: false}, {name: "excluded cluster 1", id: "cluster-1", want: false}, {name: "selected cluster", id: "cluster-13", want: true}, {name: "empty id", id: "", want: false}, {name: "alphabetic id", id: "abc", want: false}} {
		if got := FlipSelected(tc.id); got != tc.want {
			t.Errorf("FlipSelected(%q)=%v want %v", tc.id, got, tc.want)
		}
	}
}
