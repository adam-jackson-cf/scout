// Package issues correlates immutable observations with durable issues.
package issues

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"scout/internal/domain"
	"scout/internal/state"
)

const (
	correlationActor = "issue-correlation"
	maxIssueText     = 4096
)

var (
	privateKeyPattern    = regexp.MustCompile(`(?s)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----.+?-----END [A-Z0-9 ]*PRIVATE KEY-----`)
	assignmentPattern    = regexp.MustCompile(`(?i)(\b(?:api[_-]?key|access[_-]?key|secret|token|password|passwd|credential|client[_-]?secret)\b\s*[:=]\s*["']?)[^\s"']+`)
	urlCredentialPattern = regexp.MustCompile(`://[^\s/@:]+:[^\s/@]+@`)
	jwtPattern           = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b`)
	knownTokenPattern    = regexp.MustCompile(`\b(?:AKIA[0-9A-Z]{16}|gh[pousr]_[A-Za-z0-9_]{20,}|xox[baprs]-[A-Za-z0-9-]{20,})\b`)
	longTokenPattern     = regexp.MustCompile(`\b[A-Za-z0-9+/_=-]{32,}\b`)
)

type findingGroup struct {
	key          string
	mechanism    string
	category     string
	family       domain.Family
	sourceKeys   []string
	files        []string
	featureIDs   []string
	observations []domain.Observation
}

// Correlate links observations from run to durable issues. It only creates or
// links issues: it never changes a human disposition or resolves a finding.
func Correlate(ctx context.Context, store *state.Store, run domain.Run, observations []domain.Observation) ([]domain.Issue, error) {
	if store == nil {
		return nil, errors.New("issue correlation requires a state store")
	}
	if err := run.Validate(); err != nil {
		return nil, fmt.Errorf("correlation run: %w", err)
	}

	groups, err := groupFindings(run, observations)
	if err != nil {
		return nil, err
	}
	if len(groups) == 0 {
		return []domain.Issue{}, nil
	}

	existing, err := store.ListIssues(ctx)
	if err != nil {
		return nil, fmt.Errorf("list durable issues: %w", err)
	}

	issues := make([]domain.Issue, 0, len(groups))
	for _, group := range groups {
		candidates, err := matchingIssues(ctx, store, group, existing)
		if err != nil {
			return nil, err
		}
		switch len(candidates) {
		case 0:
			issue, err := createIssue(ctx, store, group)
			if err != nil {
				return nil, err
			}
			existing = append(existing, issue)
			issues = append(issues, issue)
		case 1:
			issue, err := linkObservations(ctx, store, candidates[0], group.observations)
			if err != nil {
				return nil, err
			}
			for index := range existing {
				if existing[index].IssueID == issue.IssueID {
					existing[index] = issue
					break
				}
			}
			issues = append(issues, issue)
		default:
			// Multiple durable issues share the available identity signals. Linking
			// would merge findings without proof, so retain a separate correlation.
			issue, err := createIssue(ctx, store, group)
			if err != nil {
				return nil, err
			}
			existing = append(existing, issue)
			issues = append(issues, issue)
		}
	}
	return issues, nil
}

func groupFindings(run domain.Run, observations []domain.Observation) ([]findingGroup, error) {
	featureIDs := featuresByObservation(run)
	byKey := make(map[string]*findingGroup, len(observations))
	for _, observation := range observations {
		if err := observation.Validate(); err != nil {
			return nil, fmt.Errorf("correlation observation %q: %w", observation.ObservationID, err)
		}
		if observation.RunID != run.RunID {
			return nil, fmt.Errorf("correlation observation %q belongs to run %q, not %q", observation.ObservationID, observation.RunID, run.RunID)
		}
		mechanism, err := sanitizeIssueText(observation.Mechanism)
		if err != nil || mechanism == "" {
			return nil, fmt.Errorf("correlation observation %q has unsafe mechanism", observation.ObservationID)
		}
		category, err := sanitizeIssueText(observation.OriginalCategory)
		if err != nil || category == "" {
			return nil, fmt.Errorf("correlation observation %q has unsafe category", observation.ObservationID)
		}
		files := observationFiles(observation)
		features := featureIDs[observation.ObservationID]
		sourceKey := observationSourceKey(observation)
		key := correlationKey(observation.Family, category, mechanism, []string{sourceKey}, files, features)
		group := byKey[key]
		if group == nil {
			group = &findingGroup{key: key, family: observation.Family, category: category, mechanism: mechanism, sourceKeys: []string{sourceKey}, files: files, featureIDs: features}
			byKey[key] = group
		}
		group.observations = append(group.observations, observation)
	}

	keys := make([]string, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	groups := make([]findingGroup, 0, len(keys))
	for _, key := range keys {
		group := byKey[key]
		sort.Slice(group.observations, func(i, j int) bool {
			return group.observations[i].ObservationID < group.observations[j].ObservationID
		})
		groups = append(groups, *group)
	}
	return groups, nil
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
	for observationID, featureIDs := range byObservation {
		byObservation[observationID] = sortedUnique(featureIDs)
	}
	return byObservation
}

func observationFiles(observation domain.Observation) []string {
	files := make([]string, 0, len(observation.Evidence))
	for _, evidence := range observation.Evidence {
		files = append(files, evidence.Location.Path)
	}
	return sortedUnique(files)
}

func matchingIssues(ctx context.Context, store *state.Store, group findingGroup, issues []domain.Issue) ([]domain.Issue, error) {
	matches := make([]domain.Issue, 0, 1)
	for _, issue := range issues {
		mechanism, err := sanitizeIssueText(issue.Mechanism)
		if err != nil || mechanism != group.mechanism {
			continue
		}
		if !sameCategory(issue.Title, group.category) || !sameFamily(issue, group.family) {
			continue
		}
		if !sameLocationOrFeature(group, issue) {
			continue
		}
		observations, err := store.ListIssueObservations(ctx, issue.IssueID)
		if err != nil {
			return nil, fmt.Errorf("load issue %q provenance: %w", issue.IssueID, err)
		}
		if hasMatchingSourceRule(group, observations) {
			matches = append(matches, issue)
		}
	}
	return matches, nil
}

func sameCategory(title, category string) bool {
	prefix, err := sanitizeIssueText(category + ":")
	if err != nil {
		return false
	}
	return strings.HasPrefix(title, prefix)
}

func sameFamily(issue domain.Issue, family domain.Family) bool {
	return strings.HasPrefix(issue.IssueID, issueIDPrefix(family))
}

func sameLocationOrFeature(group findingGroup, issue domain.Issue) bool {
	if len(group.featureIDs) != 0 || len(issue.FeatureIDs) != 0 {
		return len(group.featureIDs) != 0 && len(issue.FeatureIDs) != 0 && intersects(group.featureIDs, issue.FeatureIDs)
	}
	return intersects(group.files, issue.Files)
}

func createIssue(ctx context.Context, store *state.Store, group findingGroup) (domain.Issue, error) {
	issue := newIssue(group)
	if err := store.CreateIssue(ctx, issue); err == nil {
		return issue, nil
	} else {
		// The deterministic ID can collide only when another correlator created
		// the same issue concurrently. Load it and require the same full identity
		// before linking, rather than merging an unrelated durable record.
		existing, loadErr := store.LoadIssue(ctx, issue.IssueID)
		if loadErr != nil || !sameStableIdentity(group, existing) {
			return domain.Issue{}, fmt.Errorf("create correlated issue %q: %w", issue.IssueID, err)
		}
		return linkObservations(ctx, store, existing, group.observations)
	}
}

func newIssue(group findingGroup) domain.Issue {
	now := time.Now().UTC()
	observationIDs := make([]string, 0, len(group.observations))
	originalSeverity := domain.SeverityInformational
	effectiveSeverity := domain.SeverityInformational
	uncertainty := ""
	for _, observation := range group.observations {
		observationIDs = append(observationIDs, observation.ObservationID)
		if observation.SourceSeverity.Rank() > originalSeverity.Rank() {
			originalSeverity = observation.SourceSeverity
		}
		if observation.EffectiveSeverity.Rank() > effectiveSeverity.Rank() {
			effectiveSeverity = observation.EffectiveSeverity
		}
		if uncertainty == "" {
			uncertainty, _ = sanitizeIssueText(observation.Uncertainty)
		}
	}
	observationIDs = sortedUnique(observationIDs)
	title, _ := sanitizeIssueText(group.category + ": " + group.mechanism)
	return domain.Issue{
		IssueID:           issueID(group),
		State:             domain.IssueOpen,
		Title:             title,
		Mechanism:         group.mechanism,
		OriginalSeverity:  originalSeverity,
		EffectiveSeverity: effectiveSeverity,
		Uncertainty:       uncertainty,
		ObservationIDs:    observationIDs,
		FeatureIDs:        append([]string(nil), group.featureIDs...),
		Files:             append([]string(nil), group.files...),
		History: []domain.IssueHistoryEntry{{
			At:             now,
			To:             domain.IssueOpen,
			Actor:          correlationActor,
			Reason:         "created from validated observations",
			ObservationIDs: observationIDs,
		}},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func linkObservations(ctx context.Context, store *state.Store, issue domain.Issue, observations []domain.Observation) (domain.Issue, error) {
	existing := make(map[string]struct{}, len(issue.ObservationIDs))
	for _, observationID := range issue.ObservationIDs {
		existing[observationID] = struct{}{}
	}
	observationIDs := make([]string, 0, len(observations))
	for _, observation := range observations {
		if _, found := existing[observation.ObservationID]; !found {
			observationIDs = append(observationIDs, observation.ObservationID)
		}
	}
	if len(observationIDs) == 0 {
		return issue, nil
	}
	updated, err := store.LinkIssueObservations(ctx, issue.IssueID, sortedUnique(observationIDs), correlationActor, "linked repeated validated observation")
	if err == nil {
		return updated, nil
	}
	refreshed, loadErr := store.LoadIssue(ctx, issue.IssueID)
	if loadErr == nil && containsObservations(refreshed, observationIDs) {
		return refreshed, nil
	}
	return domain.Issue{}, fmt.Errorf("link observations to issue %q: %w", issue.IssueID, err)
}

func containsObservations(issue domain.Issue, observationIDs []string) bool {
	existing := make(map[string]struct{}, len(issue.ObservationIDs))
	for _, observationID := range issue.ObservationIDs {
		existing[observationID] = struct{}{}
	}
	for _, observationID := range observationIDs {
		if _, found := existing[observationID]; !found {
			return false
		}
	}
	return true
}

func sameStableIdentity(group findingGroup, issue domain.Issue) bool {
	mechanism, err := sanitizeIssueText(issue.Mechanism)
	if err != nil || mechanism != group.mechanism {
		return false
	}
	return issue.IssueID == issueID(group) && sameLocationOrFeature(group, issue)
}

func issueID(group findingGroup) string {
	sum := sha256.Sum256([]byte(group.key))
	return issueIDPrefix(group.family) + hex.EncodeToString(sum[:])
}

func issueIDPrefix(family domain.Family) string { return "issue:" + string(family) + ":" }

func correlationKey(family domain.Family, category, mechanism string, sourceKeys, files, featureIDs []string) string {
	anchors := make([]string, 0, len(sourceKeys)+len(files)+len(featureIDs))
	for _, sourceKey := range sourceKeys {
		anchors = append(anchors, "source:"+sourceKey)
	}
	for _, featureID := range featureIDs {
		anchors = append(anchors, "feature:"+featureID)
	}
	for _, file := range files {
		anchors = append(anchors, "file:"+file)
	}
	return strings.Join(append([]string{string(family), category, mechanism}, anchors...), "\x00")
}


func observationSourceKey(observation domain.Observation) string {
	return observation.Source + "\x00" + observation.RuleID
}

func hasMatchingSourceRule(group findingGroup, observations []domain.Observation) bool {
	for _, observation := range observations {
		for _, sourceKey := range group.sourceKeys {
			if observationSourceKey(observation) == sourceKey {
				return true
			}
		}
	}
	return false
}
func intersects(left, right []string) bool {
	for _, first := range left {
		for _, second := range right {
			if first == second {
				return true
			}
		}
	}
	return false
}

func sortedUnique(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	result := append([]string(nil), values...)
	sort.Strings(result)
	write := 1
	for read := 1; read < len(result); read++ {
		if result[read] != result[write-1] {
			result[write] = result[read]
			write++
		}
	}
	return result[:write]
}

// SanitizeText redacts credential-shaped values before human-supplied audit
// text is committed or printed. It rejects invalid UTF-8 and NUL bytes.
func SanitizeText(value string) (string, error) {
	return sanitizeIssueText(value)
}

func sanitizeIssueText(value string) (string, error) {
	if !utf8.ValidString(value) || strings.ContainsRune(value, '\x00') {
		return "", errors.New("unsafe issue text")
	}
	value = privateKeyPattern.ReplaceAllString(value, "[REDACTED PRIVATE KEY]")
	value = assignmentPattern.ReplaceAllString(value, "$1[REDACTED]")
	value = urlCredentialPattern.ReplaceAllString(value, "://[REDACTED]@")
	value = jwtPattern.ReplaceAllString(value, "[REDACTED JWT]")
	value = knownTokenPattern.ReplaceAllString(value, "[REDACTED TOKEN]")
	value = longTokenPattern.ReplaceAllString(value, "[REDACTED]")
	value = strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, value)
	value = strings.TrimSpace(value)
	if len(value) > maxIssueText {
		value = value[:maxIssueText]
		for !utf8.ValidString(value) {
			value = value[:len(value)-1]
		}
		value += "…"
	}
	return value, nil
}
