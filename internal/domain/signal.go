package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Digest is the one digest format of signal population, tool and config
// digests: the bare lowercase hex SHA-256 of the parts joined by NUL.
func Digest(parts ...string) string {
	hash := sha256.New()
	for i, part := range parts {
		if i > 0 {
			hash.Write([]byte{0})
		}
		hash.Write([]byte(part))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

// IsDigest reports whether s is a 64-character lowercase hex digest.
func IsDigest(s string) bool {
	if len(s) != 64 {
		return false
	}
	for _, r := range s {
		if !('0' <= r && r <= '9' || 'a' <= r && r <= 'f') {
			return false
		}
	}
	return true
}

// SignalClaim is the evidential strength of a signal. Every Phase-3 measure is
// descriptive: it reports what was observed, never a cause or a verdict.
type SignalClaim string

const ClaimDescriptive SignalClaim = "descriptive"

// SignalMissing records why a signal has no value. The zero value means the
// signal was measured. `unmeasured` (the measure could not run: tool absent,
// artifact absent, input absent) and `insufficient` (it ran but the population
// is too small to support a value) are never zero and never "good".
type SignalMissing string

const (
	SignalMeasured     SignalMissing = ""
	SignalUnmeasured   SignalMissing = "unmeasured"
	SignalInsufficient SignalMissing = "insufficient"
)

// Signal scope kinds. A signal's scope anchor is `feature:<id>`, a repository
// directory for `directory`, or the empty string for `repository`.
const (
	SignalScopeFeature    = "feature"
	SignalScopeDirectory  = "directory"
	SignalScopeRepository = "repository"
	SignalScopeFile       = "file"
	SignalScopeFunction   = "function"
)

// Signal is one descriptive measurement at a scope for a snapshot and, for
// history measures, a window (design §3.2). Signals are append-only, can never
// gate, and carry what a reader needs to judge comparability: numerator,
// denominator, the eligible-population digest and tool/config digests.
//
// Identity is (MetricID@Version, ScopeKind, ScopeAnchor, Snapshot, Window);
// ToolDigest and ConfigDigest disambiguate recomputation by a different tool
// or configuration of the same snapshot.
type Signal struct {
	MetricID    string `json:"metricId"` // e.g. "scout.cc"
	Version     int    `json:"version"`  // metric formula version
	ScopeKind   string `json:"scopeKind"`
	ScopeAnchor string `json:"scopeAnchor"`
	// Snapshot identifies the measured tree: the commit for a clean tree, else
	// "worktree:<digest>".
	Snapshot string `json:"snapshot"`
	// Window is empty for snapshot (principal) measures; history measures use
	// "<start>/<end>" calendar bounds (RFC 3339 dates, end exclusive) or
	// "trailing-365d".
	Window string `json:"window,omitempty"`

	// Numerator, Denominator and Value are nil when Missing is set. Value is
	// the headline number of the metric (a ratio or a count); Detail carries the
	// remaining parts.
	Numerator   *float64 `json:"numerator,omitempty"`
	Denominator *float64 `json:"denominator,omitempty"`
	Value       *float64 `json:"value,omitempty"`

	PopulationDigest string `json:"populationDigest"`
	ToolDigest       string `json:"toolDigest"`
	ConfigDigest     string `json:"configDigest"`

	Missing       SignalMissing `json:"missing,omitempty"`
	MissingReason string        `json:"missingReason,omitempty"`
	Claim         SignalClaim   `json:"claim"`

	// Detail holds JSON-native values only (numbers, strings, bools, slices of
	// them, string-keyed maps).
	Detail map[string]any `json:"detail,omitempty"`
}

// ID renders the versioned metric identifier, e.g. "scout.cc@1".
func (s Signal) ID() string { return s.MetricID + "@" + strconv.Itoa(s.Version) }

// Key is the identity of a signal row (including tool and config digests).
func (s Signal) Key() string {
	return strings.Join([]string{s.ID(), s.ScopeKind, s.ScopeAnchor, s.Snapshot, s.Window, s.ToolDigest, s.ConfigDigest}, "\x00")
}

// Float returns a pointer to v, for Signal numerators, denominators and values.
func Float(v float64) *float64 { return &v }

// Validate rejects records that would make a missing measurement look like a
// measured one, or that lack the context needed to compare it.
func (s Signal) Validate() error {
	if s.MetricID == "" || s.Version <= 0 {
		return errors.New("signal requires a metric ID and a positive version")
	}
	if s.ScopeKind == "" || s.Snapshot == "" {
		return errors.New("signal requires a scope kind and a snapshot")
	}
	if s.Claim == "" {
		return errors.New("signal requires a claim")
	}
	if s.ToolDigest == "" || s.ConfigDigest == "" {
		return errors.New("signal requires tool and config digests")
	}
	switch s.Missing {
	case SignalMeasured:
		if s.Value == nil || s.Numerator == nil || s.Denominator == nil {
			return fmt.Errorf("measured signal %s requires numerator, denominator and value", s.ID())
		}
		if s.MissingReason != "" {
			return errors.New("measured signal must not carry a missing reason")
		}
	case SignalUnmeasured, SignalInsufficient:
		if s.Value != nil || s.Numerator != nil || s.Denominator != nil {
			return fmt.Errorf("%s signal %s must not carry numbers", s.Missing, s.ID())
		}
		if s.MissingReason == "" {
			return fmt.Errorf("%s signal %s requires a reason", s.Missing, s.ID())
		}
	default:
		return fmt.Errorf("unknown missingness %q", s.Missing)
	}
	return nil
}
