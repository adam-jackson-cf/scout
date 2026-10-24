package engine

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"scout/internal/decisions"
	"scout/internal/decisions/thresholds"
	"scout/internal/domain"
	"scout/internal/policy"
	"scout/internal/state"
)

// LaneDecisions is the decisions lane id in health. It is optional: its
// health never changes the exit code.
const LaneDecisions = decisions.LaneID

// BaselineFormulaID names the deterministic ranking baseline
// (`rule weight × changed-in-PR × path role`). `changed-in-PR` is 1 outside PR
// runs and ties break by issue id; both are part of this id.
const BaselineFormulaID = "severity-x-changed-in-pr-x-path-role/v1;changed=1-outside-pr;ties=issue-id"

// BaselineVersion is stored with every ranking disclosure and disposition.
func BaselineVersion() decisions.BaselineVersion {
	var table strings.Builder
	for _, severity := range []domain.Severity{domain.SeverityInformational, domain.SeverityLow, domain.SeverityMedium, domain.SeverityHigh, domain.SeverityCritical} {
		fmt.Fprintf(&table, "%s=%d;", severity, severity.Rank())
	}
	sum := sha256.Sum256([]byte(table.String()))
	return decisions.BaselineVersion{FormulaID: BaselineFormulaID, RuleWeightDigest: hex.EncodeToString(sum[:8]), PathRoleVersion: PathRoleVersion}
}

// RankingInput is everything the ranking hook needs.
type RankingInput struct {
	Issues  []domain.AttributedIssue
	Root    string
	RunID   string
	Trigger domain.LaneTrigger
	Policy  policy.Decisions
	// Key is the Jev key; only its presence is ever reported.
	Key string
	// Fork marks a fork pull request: Jev never runs on one.
	Fork bool
	// NoEgress forbids the Jev call for this trigger even when consented.
	NoEgress bool
	// Store is where annotations, rankings and audit draws are kept. When nil
	// and Root is set, the state at Root is opened for the call.
	Store *state.Store
	// Client is the decisions client; nil builds one from Key.
	Client *decisions.Client
	// Evidence returns the bounded evidence of one issue; nil reads the stored
	// observations.
	Evidence func(ctx context.Context, issue domain.AttributedIssue) (decisions.Evidence, error)
	// ReadThresholds returns the repository's committed override of the
	// decisions thresholds (thresholds.ReadCommitted at the same ref the policy
	// came from); nil uses the embedded provisional defaults. An error is a
	// disclosed deterministic fallback.
	ReadThresholds func() ([]byte, error)
	// Rand draws the audit sample; nil uses crypto/rand.
	Rand          func() float64
	Now           func() time.Time
	AuditFraction float64
}

// DeterministicRanking is the baseline ranking with a disclosed reason, for
// callers that have no run (an unavailable report).
func DeterministicRanking(issues []domain.AttributedIssue, reason string) Ranking {
	ranking, _ := baselineRanking(issues, reason)
	return ranking
}

func rankable(issues []domain.AttributedIssue) []domain.AttributedIssue {
	var out []domain.AttributedIssue
	for _, item := range issues {
		if item.Attribution == domain.AttributionIntroduced || item.Attribution == domain.AttributionRegressed {
			out = append(out, item)
		}
	}
	return out
}

func baselineRanking(issues []domain.AttributedIssue, reason string) (Ranking, []decisions.Ranked) {
	disclosure := "deterministic (" + reason + ")"
	items := make([]decisions.Item, 0, len(issues))
	reasons := map[string]string{}
	for _, item := range rankable(issues) {
		items = append(items, decisions.Item{IssueID: item.IssueID, Baseline: item.RankingScore, Mode: domain.RankingModeDeterministic, Disclosure: disclosure})
		reasons[item.IssueID] = item.RankingReason
	}
	ranked := decisions.Rank(decisions.LaunchFormula(), items)
	ranking := Ranking{
		Method: disclosure, Mode: domain.RankingModeDeterministic, Reason: reason, Items: make([]RankingItem, 0, len(ranked)),
		BaselineVersion: BaselineVersion().String(), RankFormulaVersion: decisions.LaunchFormula().Version(""),
	}
	for _, r := range ranked {
		score := r.Score
		ranking.Items = append(ranking.Items, RankingItem{
			IssueID: r.IssueID, Position: r.Position, Score: &score, Reason: reasons[r.IssueID], Mode: r.Mode, Ranking: r.Disclosure, BaselineScore: r.Baseline,
		})
	}
	return ranking, ranked
}

// RankingOf ranks introduced and regressed issues. Jev ranks them when
// decisions are enabled, a key is present, egress is allowed and the run is
// trusted; otherwise the deterministic baseline ranks them and the reason is
// disclosed. Annotations only ever change this ranking, never an issue state
// or the gate, and the lane health it returns is optional.
func RankingOf(ctx context.Context, in RankingInput) (Ranking, domain.LaneHealth, error) {
	cfg := decisions.Config{
		Enabled: in.Policy.Enabled, Model: in.Policy.Model, MaxPerRun: in.Policy.MaxPerRun, MaxCostUSD: in.Policy.MaxCostUSD, EgressAcknowledged: in.Policy.EgressAcknowledged,
	}
	formula := decisions.LaunchFormula()
	formulaVersion := formula.Version(cfg.Model)
	eligibility := decisions.Eligibility{Enabled: cfg.Enabled, HasKey: in.Key != "", EgressAllowed: cfg.EgressAcknowledged && !in.NoEgress, Fork: in.Fork}
	reason, ok := eligibility.Skip()
	var resolution thresholds.Resolution
	if ok {
		if err := cfg.Validate(); err != nil {
			reason, ok = err.Error(), false
		}
	}
	if ok {
		// Thresholds are configuration, never literals: the record in force is
		// bound to the triage question's current version key. A file that is
		// unreadable or invalid is a disclosed fallback, not a guess.
		var override []byte
		var err error
		if in.ReadThresholds != nil {
			override, err = in.ReadThresholds()
		}
		if err == nil {
			resolution, err = thresholds.Resolve(override, decisions.TriageRealDefect.ID, decisions.TriageRealDefect.Version(cfg.Model))
		}
		if err != nil {
			reason, ok = "thresholds invalid: "+err.Error(), false
		}
	}
	store, closeStore, err := in.openStore()
	if err != nil {
		return Ranking{}, domain.LaneHealth{}, err
	}
	defer closeStore()
	now := in.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}

	var ranking Ranking
	var ranked []decisions.Ranked
	var health domain.LaneHealth
	var outcome decisions.Outcome
	results := map[string]decisions.UnitResult{}
	if !ok {
		ranking, ranked = baselineRanking(in.Issues, reason)
		health = decisions.SkippedHealth(reason)
	} else {
		var questionVersions []string
		for _, q := range formula.Questions {
			questionVersions = append(questionVersions, q.Version(cfg.Model))
		}
		if store != nil {
			if _, err := store.MarkAnnotationsStale(ctx, questionVersions); err != nil {
				return Ranking{}, domain.LaneHealth{}, err
			}
		}
		units := make([]decisions.Unit, 0, len(in.Issues))
		for _, item := range rankable(in.Issues) {
			evidence, err := in.evidence(ctx, store, item)
			if err != nil {
				return Ranking{}, domain.LaneHealth{}, fmt.Errorf("evidence for %s: %w", item.IssueID, err)
			}
			encoded, err := decisions.BuildState(evidence)
			if err != nil {
				return Ranking{}, domain.LaneHealth{}, fmt.Errorf("state for %s: %w", item.IssueID, err)
			}
			units = append(units, decisions.Unit{Family: decisions.FamilyTriage, Key: item.IssueID, IssueID: item.IssueID, Questions: []decisions.Question{decisions.TriageRealDefect}, State: encoded, Baseline: item.RankingScore})
		}
		client := in.Client
		if client == nil {
			client = &decisions.Client{Endpoint: cfg.Endpoint, Key: in.Key}
		}
		var sink decisions.Sink
		if store != nil {
			sink = store
		}
		outcome = decisions.Run(ctx, decisions.RunInput{Config: cfg, Client: client, RunID: in.RunID, Units: units, Sink: sink, Now: now})
		outcome.ThresholdsStale = resolution.Stale
		health = outcome.Health()
		items := make([]decisions.Item, 0, len(outcome.Results))
		baselines := map[string]domain.AttributedIssue{}
		for _, item := range in.Issues {
			baselines[item.IssueID] = item
		}
		for _, result := range outcome.Results {
			results[result.Unit.IssueID] = result
			mode, text := outcome.Disclosure(result)
			item := decisions.Item{IssueID: result.Unit.IssueID, Baseline: result.Unit.Baseline, Mode: mode, Disclosure: text}
			if len(result.Annotations) > 0 {
				item.Answers = map[string]float64{}
				for _, annotation := range result.Annotations {
					item.Answers[annotation.QuestionID] = annotation.Probability
				}
			}
			items = append(items, item)
		}
		ranked = decisions.Rank(formula, items)
		ranking = jevRanking(outcome, ranked, results, baselines)
	}
	ranking.RankFormulaVersion = formulaVersion
	ranking.BaselineVersion = BaselineVersion().String()
	if ok {
		ranking.Thresholds = resolution.Status
		if resolution.Stale {
			ranking.Thresholds += " (stale: version key changed)"
		}
		if outcome.Annotated > 0 {
			ranking.JevQuality = "insufficient"
		}
	}

	if store != nil {
		for _, r := range ranked {
			record := domain.RankingRecord{
				RunID: in.RunID, IssueID: r.IssueID, Mode: r.Mode, Disclosure: r.Disclosure, BaselineScore: r.Baseline,
				BaselineVersion: ranking.BaselineVersion, RankFormulaVersion: formulaVersion, RankPosition: r.Position, CreatedAt: now(),
			}
			if result, found := results[r.IssueID]; found && len(result.Annotations) > 0 {
				if annotation, weighted := formulaAnnotation(formula, result.Annotations); weighted {
					p := annotation.Probability
					record.Probability, record.QuestionVersion = &p, annotation.QuestionVersion
				}
			}
			if err := store.RecordRanking(ctx, record); err != nil {
				return Ranking{}, domain.LaneHealth{}, err
			}
		}
		if ok && in.Trigger == domain.TriggerNightly {
			if err := in.drawAudit(ctx, store, formulaVersion, ranked, now()); err != nil {
				return Ranking{}, domain.LaneHealth{}, err
			}
		}
	}
	return ranking, health, nil
}

// formulaAnnotation is the annotation of the highest-weighted question of the
// formula (ties by question id), the one a ranking record stores.
func formulaAnnotation(formula decisions.Formula, annotations []domain.Annotation) (domain.Annotation, bool) {
	var best domain.Annotation
	found := false
	for _, annotation := range annotations {
		weight, weighted := formula.Weights[annotation.QuestionID]
		if !weighted {
			continue
		}
		if !found || weight > formula.Weights[best.QuestionID] || weight == formula.Weights[best.QuestionID] && annotation.QuestionID < best.QuestionID {
			best, found = annotation, true
		}
	}
	return best, found
}

func jevRanking(outcome decisions.Outcome, ranked []decisions.Ranked, results map[string]decisions.UnitResult, issues map[string]domain.AttributedIssue) Ranking {
	ranking := Ranking{Items: make([]RankingItem, 0, len(ranked))}
	for _, r := range ranked {
		item := RankingItem{IssueID: r.IssueID, Position: r.Position, Reason: issues[r.IssueID].RankingReason, Mode: r.Mode, Ranking: r.Disclosure, BaselineScore: r.Baseline}
		if r.Mode == domain.RankingModeDeterministic {
			score := r.Score
			item.Score = &score
		}
		if annotation, weighted := formulaAnnotation(decisions.LaunchFormula(), results[r.IssueID].Annotations); weighted {
			item.QuestionVersion = annotation.QuestionVersion
			item.Reason = "jev P(real_defect), ordinal; baseline " + issues[r.IssueID].RankingReason
		}
		ranking.Items = append(ranking.Items, item)
	}
	jev := decisions.JevDisclosure
	if outcome.ThresholdsStale {
		jev = decisions.JevStaleDisclosure
	}
	switch {
	case outcome.Total == 0:
		ranking.Mode, ranking.Method, ranking.Reason = domain.RankingModeDeterministic, "deterministic (no units)", "no units"
	case outcome.Annotated == 0:
		reason := outcome.Results[0].Reason
		for _, result := range outcome.Results {
			if result.Status != decisions.UnitCapped {
				reason = result.Reason
				break
			}
		}
		ranking.Mode, ranking.Method, ranking.Reason = domain.RankingModeDeterministic, "deterministic ("+reason+")", reason
	case outcome.Annotated == outcome.Total:
		ranking.Mode, ranking.Method = domain.RankingModeJev, jev
	default:
		ranking.Mode = domain.RankingModeJev
		var rest []string
		if outcome.Capped > 0 {
			rest = append(rest, fmt.Sprintf("partial (%d/%d)", outcome.Attempted, outcome.Total))
		}
		if failed := outcome.Failed + outcome.Incomplete; failed > 0 {
			rest = append(rest, fmt.Sprintf("%d deterministic (call failed)", failed))
		}
		ranking.Method = jev + "; " + strings.Join(rest, "; ")
	}
	return ranking
}

func (in RankingInput) openStore() (*state.Store, func(), error) {
	if in.Store != nil {
		return in.Store, func() {}, nil
	}
	if in.Root == "" || in.RunID == "" {
		return nil, func() {}, nil
	}
	store, err := state.Open(in.Root)
	if err != nil {
		return nil, func() {}, fmt.Errorf("open state for ranking: %w", err)
	}
	return store, func() { _ = store.Close() }, nil
}

func (in RankingInput) evidence(ctx context.Context, store *state.Store, item domain.AttributedIssue) (decisions.Evidence, error) {
	if in.Evidence != nil {
		return in.Evidence(ctx, item)
	}
	evidence := decisions.Evidence{
		Severity: string(item.Severity), Path: item.Path,
		Facts: pathFacts(item.Path),
	}
	if store == nil {
		return evidence, errors.New("no state to read observations from")
	}
	observations, err := store.ListIssueObservations(ctx, item.IssueID)
	if err != nil {
		return evidence, err
	}
	for _, observation := range observations {
		evidence.Rule, evidence.Category, evidence.Mechanism = observation.RuleID, observation.OriginalCategory, observation.Mechanism
		for _, e := range observation.Evidence {
			if e.RedactedQuote != "" {
				evidence.Quotes = append(evidence.Quotes, e.RedactedQuote)
			}
		}
	}
	return evidence, nil
}

func pathFacts(path string) decisions.Facts {
	role := pathRoleName(path)
	return decisions.Facts{PathRole: role, Test: role == "test", Fixture: role == "fixture", Generated: role == "generated"}
}

func (in RankingInput) drawAudit(ctx context.Context, store *state.Store, formulaVersion string, ranked []decisions.Ranked, now time.Time) error {
	severity := map[string]domain.Severity{}
	for _, item := range in.Issues {
		severity[item.IssueID] = item.Severity
	}
	units := make([]decisions.AuditUnit, 0, len(ranked))
	for _, r := range ranked {
		units = append(units, decisions.AuditUnit{IssueID: r.IssueID, Position: r.Position, Annotated: r.Annotated, Critical: severity[r.IssueID] == domain.SeverityCritical})
	}
	fraction := in.AuditFraction
	if fraction <= 0 {
		fraction = decisions.ProvisionalFraction
	}
	// DrawAudit takes a bare draw function, so a failed random source is
	// carried out of the closure and returned: the draw never silently skips.
	var sourceErr error
	draw := in.Rand
	if draw == nil {
		draw = func() float64 {
			value, err := secureFloat()
			if err != nil {
				sourceErr = errors.Join(sourceErr, err)
			}
			return value
		}
	}
	_, err := decisions.DrawAudit(ctx, store, in.RunID, formulaVersion, units, fraction, draw, now)
	return errors.Join(err, sourceErr)
}

// secureFloat is a uniform value in [0, 1) from crypto/rand.
func secureFloat() (float64, error) {
	var buffer [8]byte
	if _, err := rand.Read(buffer[:]); err != nil {
		return 1, fmt.Errorf("audit draw random source: %w", err)
	}
	return float64(binary.BigEndian.Uint64(buffer[:])>>11) / (1 << 53), nil
}

// ApplyDecisionsHealth appends the decisions lane row to health. The lane is
// optional, so this never changes the exit code.
func ApplyDecisionsHealth(health []domain.LaneHealth, row domain.LaneHealth) []domain.LaneHealth {
	return append(append([]domain.LaneHealth(nil), health...), row)
}

// CommittedThresholds reads the thresholds override committed at ref, the ref
// the policy was loaded from, never the working tree.
func CommittedThresholds(root, ref string) func() ([]byte, error) {
	return func() ([]byte, error) { return thresholds.ReadCommitted(root, ref) }
}
