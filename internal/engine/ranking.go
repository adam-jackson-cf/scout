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
	Now            func() time.Time
	AuditFraction  float64
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
	reason, ok, resolution := rankingEligibility(in, cfg)
	store, closeStore, err := in.openStore()
	if err != nil {
		return Ranking{}, domain.LaneHealth{}, err
	}
	defer closeStore()
	run := rankingRun{in: in, cfg: cfg, formula: formula, formulaVersion: formula.Version(cfg.Model), resolution: resolution, store: store, now: rankingClock(in.Now)}

	var result rankingResult
	if !ok {
		result.ranking, result.ranked = baselineRanking(in.Issues, reason)
		result.health = decisions.SkippedHealth(reason)
	} else if result, err = run.jev(ctx); err != nil {
		return Ranking{}, domain.LaneHealth{}, err
	}
	ranking := result.ranking
	ranking.RankFormulaVersion = run.formulaVersion
	ranking.BaselineVersion = BaselineVersion().String()
	if ok {
		discloseJevRanking(&ranking, resolution, result.outcome)
	}

	if err := run.record(ctx, result, ranking.BaselineVersion, ok); err != nil {
		return Ranking{}, domain.LaneHealth{}, err
	}
	return ranking, result.health, nil
}

func rankingEligibility(in RankingInput, cfg decisions.Config) (string, bool, thresholds.Resolution) {
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
		var err error
		resolution, err = rankingThresholds(in, cfg)
		if err != nil {
			reason, ok = "thresholds invalid: "+err.Error(), false
		}
	}
	return reason, ok, resolution
}

func rankingThresholds(in RankingInput, cfg decisions.Config) (thresholds.Resolution, error) {
	var override []byte
	var err error
	if in.ReadThresholds != nil {
		override, err = in.ReadThresholds()
	}
	if err != nil {
		return thresholds.Resolution{}, err
	}
	return thresholds.Resolve(override, decisions.TriageRealDefect.ID, decisions.TriageRealDefect.Version(cfg.Model))
}

func rankingClock(now func() time.Time) func() time.Time {
	if now == nil {
		return func() time.Time { return time.Now().UTC() }
	}
	return now
}

func discloseJevRanking(ranking *Ranking, resolution thresholds.Resolution, outcome decisions.Outcome) {
	ranking.Thresholds = resolution.Status
	if resolution.Stale {
		ranking.Thresholds += " (stale: version key changed)"
	}
	if outcome.Annotated > 0 {
		ranking.JevQuality = "insufficient"
	}
}

type rankingRun struct {
	in             RankingInput
	cfg            decisions.Config
	formula        decisions.Formula
	formulaVersion string
	resolution     thresholds.Resolution
	store          *state.Store
	now            func() time.Time
}

type rankingResult struct {
	ranking Ranking
	ranked  []decisions.Ranked
	health  domain.LaneHealth
	outcome decisions.Outcome
	results map[string]decisions.UnitResult
}

func (r rankingRun) jev(ctx context.Context) (rankingResult, error) {
	var questionVersions []string
	for _, q := range r.formula.Questions {
		questionVersions = append(questionVersions, q.Version(r.cfg.Model))
	}
	if r.store != nil {
		if _, err := r.store.MarkAnnotationsStale(ctx, questionVersions); err != nil {
			return rankingResult{}, err
		}
	}
	units, err := r.units(ctx)
	if err != nil {
		return rankingResult{}, err
	}
	client := r.in.Client
	if client == nil {
		client = &decisions.Client{Endpoint: r.cfg.Endpoint, Key: r.in.Key}
	}
	var sink decisions.Sink
	if r.store != nil {
		sink = r.store
	}
	outcome := decisions.Run(ctx, decisions.RunInput{Config: r.cfg, Client: client, RunID: r.in.RunID, Units: units, Sink: sink, Now: r.now})
	outcome.ThresholdsStale = r.resolution.Stale
	result := rankingResult{outcome: outcome, health: outcome.Health(), results: map[string]decisions.UnitResult{}}
	baselines := map[string]domain.AttributedIssue{}
	for _, item := range r.in.Issues {
		baselines[item.IssueID] = item
	}
	result.ranked = decisions.Rank(r.formula, jevRankingItems(outcome, result.results))
	result.ranking = jevRanking(outcome, result.ranked, result.results, baselines)
	return result, nil
}

func (r rankingRun) units(ctx context.Context) ([]decisions.Unit, error) {
	units := make([]decisions.Unit, 0, len(r.in.Issues))
	for _, item := range rankable(r.in.Issues) {
		evidence, err := r.in.evidence(ctx, r.store, item)
		if err != nil {
			return nil, fmt.Errorf("evidence for %s: %w", item.IssueID, err)
		}
		encoded, err := decisions.BuildState(evidence)
		if err != nil {
			return nil, fmt.Errorf("state for %s: %w", item.IssueID, err)
		}
		units = append(units, decisions.Unit{Family: decisions.FamilyTriage, Key: item.IssueID, IssueID: item.IssueID, Questions: []decisions.Question{decisions.TriageRealDefect}, State: encoded, Baseline: item.RankingScore})
	}
	return units, nil
}

func jevRankingItems(outcome decisions.Outcome, results map[string]decisions.UnitResult) []decisions.Item {
	items := make([]decisions.Item, 0, len(outcome.Results))
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
	return items
}

func (r rankingRun) record(ctx context.Context, result rankingResult, baselineVersion string, ok bool) error {
	if r.store == nil {
		return nil
	}
	for _, ranked := range result.ranked {
		if err := r.store.RecordRanking(ctx, r.rankingRecord(ranked, result.results, baselineVersion)); err != nil {
			return err
		}
	}
	if ok && r.in.Trigger == domain.TriggerNightly {
		return r.in.drawAudit(ctx, r.store, r.formulaVersion, result.ranked, r.now())
	}
	return nil
}

func (r rankingRun) rankingRecord(ranked decisions.Ranked, results map[string]decisions.UnitResult, baselineVersion string) domain.RankingRecord {
	record := domain.RankingRecord{
		RunID: r.in.RunID, IssueID: ranked.IssueID, Mode: ranked.Mode, Disclosure: ranked.Disclosure, BaselineScore: ranked.Baseline,
		BaselineVersion: baselineVersion, RankFormulaVersion: r.formulaVersion, RankPosition: ranked.Position, CreatedAt: r.now(),
	}
	if result, found := results[ranked.IssueID]; found && len(result.Annotations) > 0 {
		if annotation, weighted := formulaAnnotation(r.formula, result.Annotations); weighted {
			p := annotation.Probability
			record.Probability, record.QuestionVersion = &p, annotation.QuestionVersion
		}
	}
	return record
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
		ranking.Items = append(ranking.Items, jevRankingItem(r, results, issues))
	}
	jev := decisions.JevDisclosure
	if outcome.ThresholdsStale {
		jev = decisions.JevStaleDisclosure
	}
	switch {
	case outcome.Total == 0:
		ranking.Mode, ranking.Method, ranking.Reason = domain.RankingModeDeterministic, "deterministic (no units)", "no units"
	case outcome.Annotated == 0:
		reason := firstUncappedReason(outcome)
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

func jevRankingItem(r decisions.Ranked, results map[string]decisions.UnitResult, issues map[string]domain.AttributedIssue) RankingItem {
	item := RankingItem{IssueID: r.IssueID, Position: r.Position, Reason: issues[r.IssueID].RankingReason, Mode: r.Mode, Ranking: r.Disclosure, BaselineScore: r.Baseline}
	if r.Mode == domain.RankingModeDeterministic {
		score := r.Score
		item.Score = &score
	}
	if annotation, weighted := formulaAnnotation(decisions.LaunchFormula(), results[r.IssueID].Annotations); weighted {
		item.QuestionVersion = annotation.QuestionVersion
		item.Reason = "jev P(real_defect), ordinal; baseline " + issues[r.IssueID].RankingReason
	}
	return item
}

func firstUncappedReason(outcome decisions.Outcome) string {
	reason := outcome.Results[0].Reason
	for _, result := range outcome.Results {
		if result.Status != decisions.UnitCapped {
			reason = result.Reason
			break
		}
	}
	return reason
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
	draw := func() float64 {
		value, err := secureFloat()
		if err != nil {
			sourceErr = errors.Join(sourceErr, err)
		}
		return value
	}
	_, err := decisions.DrawAudit(ctx, store, decisions.AuditRequest{RunID: in.RunID, FormulaVersion: formulaVersion, Units: units, Fraction: fraction, Rand: draw, Now: now})
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
