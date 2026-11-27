// Package issues correlates immutable observations with durable issues.
//
// Identity follows REQ-5: an issue has a durable opaque ID minted once;
// matching uses the anchor set {rename-followed path, qualified symbol, rule
// key, normalized mechanism, normalized evidence fingerprint}. Line and byte
// ranges are locators on observations and never take part in matching.
package issues

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"scout/internal/domain"
	"scout/internal/redact"
	"scout/internal/state"
)

const (
	correlationActor = "issue-correlation"
	maxIssueText     = 4096
)

// Option adjusts one correlation.
type Option func(*options)

type options struct {
	renames map[string]string
	deleted map[string]bool
}

// WithRenames supplies content-stable renames from the inventory diff, keyed by
// previous path, so an issue follows its file (ladder step d).
func WithRenames(renames map[string]string) Option {
	return func(o *options) { o.renames = renames }
}

// WithDeleted supplies paths the inventory diff shows deleted (not renamed).
// Their open issues resolve with reason file_deleted, but only when the run is
// complete and its scope includes the path.
func WithDeleted(paths []string) Option {
	return func(o *options) {
		o.deleted = make(map[string]bool, len(paths))
		for _, path := range paths {
			o.deleted[path] = true
		}
	}
}

// occurrence is one distinct flagged location in a run.
type occurrence struct {
	family       domain.Family
	category     string
	mechanism    string
	anchor       domain.IssueAnchor
	observations []domain.Observation
	featureIDs   []string
	// ambiguous lists the existing issues this occurrence cannot be told apart
	// from; it is then created as a new issue and never merged.
	ambiguous []string
}

// candidate is an existing issue offered to the ladder for this run.
type candidate struct {
	issue domain.Issue
	// path is the issue's anchor path after following inventory renames.
	path string
	// taken: already matched to an occurrence this run.
	taken bool
	// held: ambiguous this run, so it is neither matched nor resolved.
	held bool
}

// Correlate links observations from run to durable issues (REQ-6), then
// resolves issues whose file was successfully re-checked without finding them.
// It never changes a human disposition.
func Correlate(ctx context.Context, store *state.Store, run domain.Run, observations []domain.Observation, opts ...Option) ([]domain.Issue, error) {
	if store == nil {
		return nil, errors.New("issue correlation requires a state store")
	}
	if err := run.Validate(); err != nil {
		return nil, fmt.Errorf("correlation run: %w", err)
	}
	settings := applyOptions(opts)
	occurrences, err := groupOccurrences(run, observations)
	if err != nil {
		return nil, err
	}
	pool, err := loadCandidates(ctx, store, settings.renames)
	if err != nil {
		return nil, err
	}

	matched := make([]*candidate, len(occurrences))
	drift := make([]bool, len(occurrences))
	matchExact(occurrences, pool, matched)
	matchLadder(occurrences, pool, matched, drift, true)
	matchLadder(occurrences, pool, matched, drift, false)

	result := make([]domain.Issue, 0, len(occurrences))
	for index, occ := range occurrences {
		issue, err := recordOccurrence(ctx, store, occ, matched[index], drift[index])
		if err != nil {
			return nil, err
		}
		result = append(result, issue)
	}
	if err := resolveAbsent(ctx, store, run, pool, settings); err != nil {
		return nil, err
	}
	return result, nil
}

func applyOptions(opts []Option) options {
	var settings options
	for _, option := range opts {
		option(&settings)
	}
	return settings
}

func loadCandidates(ctx context.Context, store *state.Store, renames map[string]string) ([]*candidate, error) {
	existing, err := store.ListIssues(ctx)
	if err != nil {
		return nil, fmt.Errorf("list durable issues: %w", err)
	}
	sortByCreation(existing)
	pool := make([]*candidate, 0, len(existing))
	for _, issue := range existing {
		if issue.Anchor.RuleKey == "" {
			continue // migrated legacy row listed as ambiguous; it has nothing to match on
		}
		path := issue.Anchor.Path
		if renamed, found := renames[path]; found {
			path = renamed
		}
		pool = append(pool, &candidate{issue: issue, path: path})
	}
	return pool, nil
}

func sortByCreation(existing []domain.Issue) {
	sort.Slice(existing, func(i, j int) bool {
		if !existing[i].CreatedAt.Equal(existing[j].CreatedAt) {
			return existing[i].CreatedAt.Before(existing[j].CreatedAt)
		}
		return existing[i].IssueID < existing[j].IssueID
	})
}

func recordOccurrence(ctx context.Context, store *state.Store, occ *occurrence, match *candidate, drift bool) (domain.Issue, error) {
	if match != nil {
		return store.RecordIssueMatch(ctx, state.IssueMatch{
			IssueID: match.issue.IssueID, ObservationIDs: observationIDs(occ), Anchor: occ.anchor,
			Files: []string{occ.anchor.Path}, FeatureIDs: occ.featureIDs, Drift: drift,
			Actor: correlationActor, Reason: "linked repeated validated observation",
		})
	}
	issue, err := createIssue(ctx, store, occ)
	if err != nil {
		return domain.Issue{}, err
	}
	for _, peer := range occ.ambiguous {
		if _, err := store.MarkIssueAmbiguous(ctx, peer, []string{issue.IssueID}, correlationActor); err != nil {
			return domain.Issue{}, err
		}
	}
	return issue, nil
}

// matchExact is ladder step (a): rule + path + symbol + fingerprint. A secret
// matches by its digest alone so a moved secret stays the same issue.
func matchExact(occurrences []*occurrence, pool []*candidate, matched []*candidate) {
	for _, resolvedOK := range []bool{false, true} {
		for index, occ := range occurrences {
			if matched[index] != nil {
				continue
			}
			for _, cand := range pool {
				if !exactMatch(occ, cand, resolvedOK) {
					continue
				}
				cand.taken, matched[index] = true, cand
				break
			}
		}
	}
}

func exactMatch(occ *occurrence, cand *candidate, resolvedOK bool) bool {
	if cand.taken || cand.held || (cand.issue.State == domain.IssueResolved) != resolvedOK || !sameRule(occ, cand) {
		return false
	}
	if occ.anchor.Secret != cand.issue.Anchor.Secret || occ.anchor.Fingerprint != cand.issue.Anchor.Fingerprint {
		return false
	}
	return occ.anchor.Secret || (cand.path == occ.anchor.Path && cand.issue.Anchor.Symbol == occ.anchor.Symbol)
}

// matchLadder runs steps (b) and (c) over occurrences left after (a), per
// bucket. withSymbol selects (b): same symbol, fingerprint changed; otherwise
// (c): no symbol, exactly one candidate in the file. A bucket with exactly one
// occurrence and one candidate matches with evidence drift; any other mix of
// candidates is ambiguous: the occurrences become new issues linked through
// ambiguous_with and the candidates are held, never merged or resolved.
func matchLadder(occurrences []*occurrence, pool []*candidate, matched []*candidate, drift []bool, withSymbol bool) {
	buckets, order := ladderBuckets(occurrences, matched, withSymbol)
	for _, key := range order {
		indexes := buckets[key]
		first := occurrences[indexes[0]]
		var candidates []*candidate
		for _, cand := range pool {
			if ladderCandidate(cand, first, withSymbol) {
				candidates = append(candidates, cand)
			}
		}
		settleBucket(occurrences, indexes, candidates, matched, drift)
	}
}

func ladderBuckets(occurrences []*occurrence, matched []*candidate, withSymbol bool) (map[string][]int, []string) {
	buckets := map[string][]int{}
	var order []string
	for index, occ := range occurrences {
		if matched[index] != nil || len(occ.ambiguous) != 0 || occ.anchor.Secret || (occ.anchor.Symbol != "") != withSymbol {
			continue
		}
		key := occ.anchor.RuleKey + "\x1f" + string(occ.family) + "\x1f" + occ.category + "\x1f" + occ.anchor.Path + "\x1f" + occ.anchor.Symbol
		if _, seen := buckets[key]; !seen {
			order = append(order, key)
		}
		buckets[key] = append(buckets[key], index)
	}
	return buckets, order
}

func ladderCandidate(cand *candidate, first *occurrence, withSymbol bool) bool {
	if cand.taken || cand.held || cand.issue.State == domain.IssueResolved || cand.issue.Anchor.Secret || cand.path != first.anchor.Path || !sameRule(first, cand) {
		return false
	}
	return !withSymbol || cand.issue.Anchor.Symbol == first.anchor.Symbol
}

func settleBucket(occurrences []*occurrence, indexes []int, candidates []*candidate, matched []*candidate, drift []bool) {
	first := occurrences[indexes[0]]
	switch {
	case len(candidates) == 0:
	case len(indexes) == 1 && len(candidates) == 1:
		candidates[0].taken, matched[indexes[0]] = true, candidates[0]
		drift[indexes[0]] = candidates[0].issue.Anchor.Fingerprint != first.anchor.Fingerprint
	default:
		ids := make([]string, 0, len(candidates))
		for _, cand := range candidates {
			cand.held = true
			ids = append(ids, cand.issue.IssueID)
		}
		for _, index := range indexes {
			occurrences[index].ambiguous = ids
		}
	}
}

func sameRule(occ *occurrence, cand *candidate) bool {
	return cand.issue.Family == occ.family && cand.issue.Category == occ.category && cand.issue.Anchor.RuleKey == occ.anchor.RuleKey
}

// scopeIncludes reports whether the run's scope covers path; commit-range and
// staged scopes are conservatively treated as not covering unlisted paths.
func scopeIncludes(scope domain.Scope, path string) bool {
	switch scope.Kind {
	case domain.ScopeRepository:
		return true
	case domain.ScopeSubpath, domain.ScopePaths:
		for _, prefix := range scope.Paths {
			if path == prefix || strings.HasPrefix(path, strings.TrimSuffix(prefix, "/")+"/") {
				return true
			}
		}
	}
	return false
}

// resolveAbsent resolves each issue that was neither matched nor held, is not a
// dispositioned false positive/won't fix, and either (a) had its file deleted
// (not renamed) in a complete run whose scope includes it, or (b) has relevant
// successful coverage in this run. Reused, skipped, blocked and incomplete
// coverage (and files absent from a partial run) leave the issue open.
func resolveAbsent(ctx context.Context, store *state.Store, run domain.Run, pool []*candidate, settings options) error {
	for _, cand := range pool {
		if err := followRename(ctx, store, cand); err != nil {
			return err
		}
		if !resolutionCandidate(cand) {
			continue
		}
		if err := resolveCandidate(ctx, store, run, cand, settings); err != nil {
			return err
		}
	}
	return nil
}

func followRename(ctx context.Context, store *state.Store, cand *candidate) error {
	if !cand.taken && cand.path != cand.issue.Anchor.Path {
		if err := store.FollowIssueRename(ctx, cand.issue.IssueID, cand.path); err != nil {
			return fmt.Errorf("follow rename for issue %q: %w", cand.issue.IssueID, err)
		}
	}
	return nil
}

func resolutionCandidate(cand *candidate) bool {
	switch cand.issue.State {
	case domain.IssueResolved, domain.IssueFalsePositive, domain.IssueWontFix:
		return false
	}
	return !cand.taken && !cand.held
}

func resolveCandidate(ctx context.Context, store *state.Store, run domain.Run, cand *candidate, settings options) error {
	if settings.deleted[cand.path] && run.State == domain.RunComplete && scopeIncludes(run.Scope, cand.path) {
		verification := domain.Verification{RunID: run.RunID, Reason: domain.ResolutionFileDeleted, Relevant: true, FailureModeAbsent: true, VerifiedAt: time.Now().UTC()}
		if _, err := store.ResolveIssue(ctx, state.Resolution{IssueID: cand.issue.IssueID, Verification: verification, Actor: correlationActor, Reason: domain.ResolutionFileDeleted + ": file removed in a complete run"}); err != nil {
			return fmt.Errorf("resolve issue %q: %w", cand.issue.IssueID, err)
		}
		return nil
	}
	followed := cand.issue
	followed.Anchor.Path = cand.path
	best := bestCoverage(run, followed)
	if best == nil {
		return nil
	}
	verification := domain.Verification{RunID: run.RunID, Relevant: true, InspectionStatus: best.Status, FailureModeAbsent: true, VerifiedAt: time.Now().UTC()}
	if _, err := store.ResolveIssue(ctx, state.Resolution{IssueID: cand.issue.IssueID, Verification: verification, FollowPath: cand.path, Actor: correlationActor, Reason: "failure mode absent after a successful relevant re-check"}); err != nil {
		return fmt.Errorf("resolve issue %q: %w", cand.issue.IssueID, err)
	}
	return nil
}

func bestCoverage(run domain.Run, followed domain.Issue) *domain.Coverage {
	var best *domain.Coverage
	for index := range run.Coverage {
		row := &run.Coverage[index]
		if followed.CoverageRelevant(*row) && (best == nil || best.Status == domain.CoverageFlagged) {
			best = row
		}
	}
	return best
}

type keyed struct {
	observation domain.Observation
	anchor      domain.IssueAnchor
	mechanism   string
	category    string
}

func groupOccurrences(run domain.Run, observations []domain.Observation) ([]*occurrence, error) {
	featureIDs := featuresByObservation(run)
	rows := make([]keyed, 0, len(observations))
	for _, observation := range observations {
		row, err := keyObservation(run, observation)
		if err != nil {
			return nil, err
		}
		rows = append(rows, row)
	}
	sortByLocation(rows)
	byLocation := map[string]*occurrence{}
	occurrences := make([]*occurrence, 0, len(rows))
	for _, row := range rows {
		location := row.observation.Evidence[0].Location
		key := fmt.Sprintf("%s\x1f%s\x1f%s\x1f%v\x1f%v\x1f%v\x1f%v\x1f%v", row.observation.Family, row.category, row.anchor.RuleKey, row.anchor.Secret, row.anchor.Fingerprint, location.Path, location.StartLine, location.StartByte)
		occ := byLocation[key]
		if occ == nil {
			occ = &occurrence{family: row.observation.Family, category: row.category, mechanism: row.mechanism, anchor: row.anchor}
			byLocation[key] = occ
			occurrences = append(occurrences, occ)
		}
		occ.observations = append(occ.observations, row.observation)
		occ.featureIDs = sortedUnique(append(occ.featureIDs, featureIDs[row.observation.ObservationID]...))
	}
	return occurrences, nil
}

func keyObservation(run domain.Run, observation domain.Observation) (keyed, error) {
	if err := observation.Validate(); err != nil {
		return keyed{}, fmt.Errorf("correlation observation %q: %w", observation.ObservationID, err)
	}
	if observation.RunID != run.RunID {
		return keyed{}, fmt.Errorf("correlation observation %q belongs to run %q, not %q", observation.ObservationID, observation.RunID, run.RunID)
	}
	mechanism, err := SanitizeText(observation.Mechanism)
	if err != nil || mechanism == "" {
		return keyed{}, fmt.Errorf("correlation observation %q has unsafe mechanism", observation.ObservationID)
	}
	category, err := SanitizeText(observation.OriginalCategory)
	if err != nil || category == "" {
		return keyed{}, fmt.Errorf("correlation observation %q has unsafe category", observation.ObservationID)
	}
	anchor, ok := domain.AnchorOf(observation)
	if !ok {
		return keyed{}, fmt.Errorf("correlation observation %q has no evidence", observation.ObservationID)
	}
	return keyed{observation: observation, anchor: anchor, mechanism: mechanism, category: category}, nil
}

func sortByLocation(rows []keyed) {
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i].observation.Evidence[0].Location, rows[j].observation.Evidence[0].Location
		switch {
		case a.Path != b.Path:
			return a.Path < b.Path
		case a.StartLine != b.StartLine:
			return a.StartLine < b.StartLine
		case a.StartByte != b.StartByte:
			return a.StartByte < b.StartByte
		}
		return rows[i].observation.ObservationID < rows[j].observation.ObservationID
	})
}

func featuresByObservation(run domain.Run) map[string][]string {
	byObservation := make(map[string][]string)
	for _, coverage := range run.Coverage {
		if coverage.Key.FeatureID == "" {
			continue
		}
		for _, observationID := range coverage.ObservationIDs {
			byObservation[observationID] = append(byObservation[observationID], coverage.Key.FeatureID)
		}
	}
	for observationID, ids := range byObservation {
		byObservation[observationID] = sortedUnique(ids)
	}
	return byObservation
}

func observationIDs(occ *occurrence) []string {
	ids := make([]string, 0, len(occ.observations))
	for _, observation := range occ.observations {
		ids = append(ids, observation.ObservationID)
	}
	return sortedUnique(ids)
}

func createIssue(ctx context.Context, store *state.Store, occ *occurrence) (domain.Issue, error) {
	now := time.Now().UTC()
	ids := observationIDs(occ)
	claim, rule := firstClaim(occ.observations)
	originalSeverity, effectiveSeverity := domain.SeverityInformational, domain.SeverityInformational
	uncertainty := ""
	for _, observation := range occ.observations {
		if observation.SourceSeverity.Rank() > originalSeverity.Rank() {
			originalSeverity = observation.SourceSeverity
		}
		if observation.EffectiveSeverity.Rank() > effectiveSeverity.Rank() {
			effectiveSeverity = observation.EffectiveSeverity
		}
		if uncertainty == "" {
			var err error
			if uncertainty, err = SanitizeText(observation.Uncertainty); err != nil {
				return domain.Issue{}, fmt.Errorf("observation %q uncertainty: %w", observation.ObservationID, err)
			}
		}
	}
	title, err := SanitizeText(occ.category + ": " + occ.mechanism)
	if err != nil {
		return domain.Issue{}, fmt.Errorf("issue title: %w", err)
	}
	issue := domain.Issue{
		IssueID: domain.NewIssueID(occ.family, time.Now()), State: domain.IssueOpen, Family: occ.family, Category: occ.category,
		Title: title, Mechanism: occ.mechanism, Anchor: occ.anchor,
		OriginalSeverity: originalSeverity, EffectiveSeverity: effectiveSeverity, Uncertainty: uncertainty,
		ObservationIDs: ids, FeatureIDs: append([]string(nil), occ.featureIDs...), Files: []string{occ.anchor.Path},
		Claim: claim, Rule: rule,
		AmbiguousWith: append([]string(nil), occ.ambiguous...),
		History:       []domain.IssueHistoryEntry{{At: now, To: domain.IssueOpen, Event: domain.EventCreated, Actor: correlationActor, Reason: "created from validated observations", ObservationIDs: ids}},
		CreatedAt:     now, UpdatedAt: now,
	}
	if len(occ.ambiguous) != 0 {
		issue.History = append(issue.History, domain.IssueHistoryEntry{At: now, From: domain.IssueOpen, To: domain.IssueOpen, Event: domain.EventAmbiguous, Actor: correlationActor, Reason: "cannot be told apart from existing issues; not merged"})
	}
	if err := store.CreateIssue(ctx, issue); err != nil {
		return domain.Issue{}, fmt.Errorf("create correlated issue %q: %w", issue.IssueID, err)
	}
	return issue, nil
}

func firstClaim(observations []domain.Observation) (domain.SignalClaim, string) {
	for _, observation := range observations {
		if observation.Claim != "" {
			return observation.Claim, observation.RuleID
		}
	}
	return "", ""
}

func sortedUnique(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	return slices.Compact(slices.Sorted(slices.Values(values)))
}

// SanitizeText redacts credential-shaped values before human-supplied audit
// text is committed or printed. It rejects invalid UTF-8 and NUL bytes.
func SanitizeText(value string) (string, error) {
	value, err := redact.Text(value)
	if err != nil {
		return "", errors.New("unsafe issue text")
	}
	return redact.Truncate(strings.TrimSpace(value), maxIssueText), nil
}
