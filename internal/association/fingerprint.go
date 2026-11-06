package association

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math"

	"scout/internal/state"
)

// Fingerprint contains decision semantics, deliberately not catalogue content.
// ChoiceQuestion identifies the frozen template, not a realised option list.
type Fingerprint struct {
	CoveredScore    float64
	TChoice         float64
	CoveredQuestion string
	ChoiceQuestion  string
	JevModel        string
	PromptDigest    string
	FallbackModel   string
}

// Digest hashes exact threshold values and length-delimited question/model
// identities. Catalogue-only edits therefore cannot invalidate decisions.
func (f Fingerprint) Digest() string {
	texts := [...]string{f.CoveredQuestion, f.ChoiceQuestion, f.JevModel, f.PromptDigest, f.FallbackModel}
	size := 16 + 8*len(texts)
	for _, text := range texts {
		size += len(text)
	}
	payload := make([]byte, 0, size)
	payload = binary.BigEndian.AppendUint64(payload, math.Float64bits(f.CoveredScore))
	payload = binary.BigEndian.AppendUint64(payload, math.Float64bits(f.TChoice))
	for _, text := range texts {
		payload = binary.BigEndian.AppendUint64(payload, uint64(len(text)))
		payload = append(payload, text...)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

// MarkStale preserves every prior label and sets only the stale flag when a
// recorded decision fingerprint differs. Legacy rows without a digest stay put.
func MarkStale(ctx context.Context, store *state.Store, fingerprint Fingerprint) error {
	rows, err := store.ListAssociations(ctx)
	if err != nil {
		return err
	}
	digest := fingerprint.Digest()
	var ids []string
	for _, row := range rows {
		if !row.Stale && row.DecisionDigest != "" && row.DecisionDigest != digest {
			ids = append(ids, row.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return store.MarkAssociationsStale(ctx, ids)
}
