package domain

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// SecretMechanism is the mechanism text analyzers use for credential findings.
// Such findings carry a secret digest (Observation.SourceFindingID) as their
// identity so a moved secret stays the same issue.
const SecretMechanism = "secret diagnostic reported"

// AnchorOf derives the identity anchor of an observation (design §3.3). Line
// and byte ranges are deliberately ignored: they are locators, not identity.
// ok is false for an observation with no evidence location.
func AnchorOf(observation Observation) (IssueAnchor, bool) {
	if len(observation.Evidence) == 0 {
		return IssueAnchor{}, false
	}
	primary := observation.Evidence[0].Location
	ruleKey := observation.Source + ":" + observation.RuleID
	if observation.RuleID == "" {
		ruleKey = observation.Source + ":category:" + observation.OriginalCategory
	}
	anchor := IssueAnchor{
		Path:      primary.Path,
		Symbol:    primary.Symbol,
		RuleKey:   ruleKey,
		Mechanism: NormalizeText(observation.Mechanism),
	}
	if observation.Mechanism == SecretMechanism && observation.SourceFindingID != "" {
		anchor.Fingerprint = observation.SourceFindingID
		anchor.Secret = true
		return anchor, true
	}
	// Numbers in the message (e.g. a complexity value) are evidence, so the
	// normalized mechanism text is part of the fingerprint with the quotes.
	quotes := []string{anchor.Mechanism}
	for _, evidence := range observation.Evidence {
		if quote := NormalizeText(evidence.RedactedQuote); quote != "" {
			quotes = append(quotes, quote)
		}
	}
	sum := sha256.Sum256([]byte(strings.Join(quotes, "\n")))
	anchor.Fingerprint = hex.EncodeToString(sum[:])
	return anchor, true
}

// NormalizeText lowercases and collapses whitespace so cosmetic edits do not
// change a fingerprint or mechanism comparison.
func NormalizeText(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

const crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// NewIssueID mints the durable, opaque issue:<family>:<ulid> identifier.
func NewIssueID(family Family, now time.Time) string {
	return "issue:" + string(family) + ":" + newULID(now)
}

func newULID(now time.Time) string {
	var raw [16]byte
	millis := uint64(now.UTC().UnixMilli())
	for index := 5; index >= 0; index-- {
		raw[index] = byte(millis)
		millis >>= 8
	}
	if _, err := rand.Read(raw[6:]); err != nil {
		panic("domain: crypto/rand unavailable: " + err.Error())
	}
	// 128 bits -> 26 base32 characters (the first carries only 3 bits).
	var out [26]byte
	var accumulator uint
	bits := 0
	position := 25
	for index := 15; index >= 0; index-- {
		accumulator |= uint(raw[index]) << bits
		bits += 8
		for bits >= 5 && position >= 0 {
			out[position] = crockford[accumulator&31]
			accumulator >>= 5
			bits -= 5
			position--
		}
	}
	if position >= 0 {
		out[position] = crockford[accumulator&31]
	}
	return string(out[:])
}

// CoverageRelevant reports whether coverage is a successful re-check of the
// issue's own file, family and category: the only evidence that may resolve it.
// Skipped, blocked, incomplete and reused coverage is never relevant.
func (i Issue) CoverageRelevant(coverage Coverage) bool {
	switch coverage.Status {
	case CoverageInspected, CoverageCovered, CoverageFlagged:
	default:
		return false
	}
	if (i.Family != "" && coverage.Key.Family != i.Family) || coverage.Key.Category != i.Category || i.Anchor.Path == "" {
		return false
	}
	if coverage.Key.FilePath != "" {
		return coverage.Key.FilePath == i.Anchor.Path
	}
	for _, inspected := range coverage.InspectedRanges {
		if inspected.Path == i.Anchor.Path {
			return true
		}
	}
	return false
}
