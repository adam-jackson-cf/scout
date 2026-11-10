package shadow

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"scout/internal/domain"
	"scout/internal/policy"
)

// EncodingVersion is retained as part of the immutable dossier contract.
const EncodingVersion = "scout.shadow/v1"

type Query struct {
	RepositoryID, CheckID, RuleVersion, EncodingVersion string
	AsOf                                                *time.Time
	Now                                                 time.Time
}

type Report struct {
	ExitCode int
	Dossier  Dossier
	Live     LiveEnvelope
}

// LiveEnvelope cannot contribute to an activation dossier's digest.
type LiveEnvelope struct {
	Certification string
	ReaderTime    time.Time
	FrontierAge   time.Duration
	Reason        string
}

type Dossier struct {
	EncodingVersion                        string
	RepositoryID                           string
	CheckID                                string
	RuleVersion                            string
	AsOf                                   time.Time
	Frontier                               domain.ShadowFrontier
	Interval                               domain.ShadowInterval
	Origin                                 *time.Time
	Stats                                  policy.ShadowStats
	DistinctPRs                            int
	Inspected, NotInspected, NotApplicable int
	LateGaps                               int
	PriorNotWarranted, PriorUndecided      int
	SupersededNotWarranted                 int
	Representatives                        []domain.AttemptID
	PinMixtures                            map[string][]string
	PinCounts                              []PinCount
	Triage                                 []EventVerdict
	Ready                                  bool
	Reason                                 string
}

type PinCount struct {
	Pins         domain.ShadowPins
	Observations int
}
type EventVerdict struct{ EventID, Verdict, ActionID string }

// CanonicalDossier encodes only immutable decision inputs, with sorted sets and
// UTC millisecond coordinates. Unsupported formats are never silently replaced.
func CanonicalDossier(d Dossier) ([]byte, string, error) {
	if d.EncodingVersion != EncodingVersion {
		return nil, "", fmt.Errorf("unsupported shadow encoding %q", d.EncodingVersion)
	}
	d.AsOf = canonicalTime(d.AsOf)
	d.Frontier.CompleteThrough = canonicalTime(d.Frontier.CompleteThrough)
	d.Interval.After = canonicalTime(d.Interval.After)
	d.Interval.Through = canonicalTime(d.Interval.Through)
	if d.Origin != nil {
		x := canonicalTime(*d.Origin)
		d.Origin = &x
	}
	d.Representatives = append([]domain.AttemptID(nil), d.Representatives...)
	sort.Slice(d.Representatives, func(i, j int) bool { return compareID(d.Representatives[i], d.Representatives[j]) < 0 })
	mixtures := make(map[string][]string, len(d.PinMixtures))
	for key, values := range d.PinMixtures {
		v := append([]string(nil), values...)
		sort.Strings(v)
		mixtures[key] = v
	}
	d.PinMixtures = mixtures
	d.PinCounts = append([]PinCount(nil), d.PinCounts...)
	sort.Slice(d.PinCounts, func(i, j int) bool {
		a, b := d.PinCounts[i].Pins, d.PinCounts[j].Pins
		for _, pair := range [][2]string{{a.Scout, b.Scout}, {a.Tools, b.Tools}, {a.Catalogue, b.Catalogue}, {a.Policy, b.Policy}, {a.Recipe, b.Recipe}} {
			if pair[0] != pair[1] {
				return pair[0] < pair[1]
			}
		}
		return false
	})
	d.Triage = append([]EventVerdict(nil), d.Triage...)
	sort.Slice(d.Triage, func(i, j int) bool { return d.Triage[i].EventID < d.Triage[j].EventID })
	b, err := json.Marshal(d)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(b)
	return b, "sha256:" + hex.EncodeToString(sum[:]), nil
}
func canonicalTime(t time.Time) time.Time { return t.UTC().Truncate(time.Millisecond) }
