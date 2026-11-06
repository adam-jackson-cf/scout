package association

import (
	"crypto/sha256"

	"scout/internal/domain"
)

// Scores holds the covered answer and choice answers in permutation order.
// FailureReason records a Jev failure or cap, overriding any partial answers.
type Scores struct {
	Covered       *float64
	Choices       []ChoiceScore
	FailureReason string
}

// ChoiceScore is the winning area and its probability for one permutation.
type ChoiceScore struct {
	AreaID     string
	Confidence float64
}

// Composed is the Jev-only decision, before any fallback lane.
type Composed struct {
	Outcome    string
	AreaID     string
	Confidence float64
	Reason     string
}

// Compose applies the frozen pilot covered tie rule and choice threshold.
func Compose(clusterID string, scores Scores, coveredScore, tChoice float64) Composed {
	if scores.FailureReason != "" {
		return Composed{Outcome: "unresolved", Reason: scores.FailureReason}
	}
	if scores.Covered == nil {
		return Composed{Outcome: "unresolved", Reason: domain.AssociationReasonJevFailure}
	}
	if confidence := 1 - *scores.Covered; confidence >= coveredScore-1e-12 {
		return Composed{Outcome: "none", Confidence: confidence}
	}
	n := 1
	if FlipSelected(clusterID) {
		n = 2
	}
	if len(scores.Choices) < n {
		return Composed{Outcome: "unresolved", Reason: domain.AssociationReasonJevFailure}
	}
	first := scores.Choices[0]
	confidence := first.Confidence
	if n == 2 {
		second := scores.Choices[1]
		if first.AreaID != second.AreaID {
			return Composed{Outcome: "unresolved", Reason: domain.AssociationReasonPermutationNo}
		}
		confidence = (confidence + second.Confidence) / 2
	}
	if confidence >= tChoice {
		return Composed{Outcome: "assigned", AreaID: first.AreaID, Confidence: confidence}
	}
	return Composed{Outcome: "unresolved", Confidence: confidence, Reason: domain.AssociationReasonBelowTChoice}
}

// FlipSelected selects the frozen 20% subset: the full big-endian SHA-256
// integer modulo five, without allocating an arbitrary-precision integer.
func FlipSelected(clusterID string) bool {
	sum := sha256.Sum256([]byte("x2d-flip" + clusterID))
	remainder := 0
	for _, b := range sum {
		remainder = (remainder*256 + int(b)) % 5
	}
	return remainder == 0
}
