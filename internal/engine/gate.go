package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"scout/internal/domain"
	"scout/internal/policy"
)

// Exit codes (design §7.4): precedence 3 > 2 > 1 > 0.
const (
	ExitClean       = 0
	ExitGated       = 1
	ExitIncomplete  = 2
	ExitUnavailable = 3
)

// ExitCode applies the precedence unavailable > incomplete > gated > clean.
func ExitCode(unavailable, incomplete, gated bool) int {
	switch {
	case unavailable:
		return ExitUnavailable
	case incomplete:
		return ExitIncomplete
	case gated:
		return ExitGated
	default:
		return ExitClean
	}
}

// StatusOf names an exit code.
func StatusOf(code int) string {
	switch code {
	case ExitGated:
		return "gated"
	case ExitIncomplete:
		return "incomplete"
	case ExitUnavailable:
		return "unavailable"
	default:
		return "clean"
	}
}

// CheckIDOf maps an observation's source and rule to the policy check id.
// The result is only meaningful to policy.CheckStatus, which decides
// eligibility; ids outside the allow-list come back as not_eligible there.
func CheckIDOf(source, ruleID string) string {
	switch {
	case strings.HasPrefix(ruleID, "delta."):
		return ruleID // promoted delta rule id: name@version, never blocking-eligible
	case strings.HasPrefix(ruleID, "guardrails."):
		return ruleID
	case strings.HasPrefix(source, "security.gitleaks"):
		return "security.secrets." + ruleID
	case ruleID == "scout.security.actions.mutable-ref":
		return "security.ci-mutable-action-ref"
	default:
		return source + ":" + ruleID
	}
}

// CheckFact is the check identity of one issue.
type CheckFact struct {
	CheckID     string
	RuleVersion string
}

// OverrideRequest is one parsed `scout:allow` directive, or the blanket
// `scout-override` label (All).
type OverrideRequest struct {
	CheckID string
	Reason  string
	Ref     string
	All     bool
}

// ParseOverrides extracts override requests from a PR body or label text:
// `scout:allow <id> reason=… ref=…` lines and the `scout-override` label.
func ParseOverrides(text string) []OverrideRequest {
	var requests []OverrideRequest
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "scout-override" {
			requests = append(requests, OverrideRequest{All: true, Reason: "label scout-override"})
			continue
		}
		index := strings.Index(line, "scout:allow ")
		if index < 0 {
			continue
		}
		fields := strings.TrimSpace(line[index+len("scout:allow "):])
		id, rest, _ := strings.Cut(fields, " ")
		request := OverrideRequest{CheckID: strings.TrimSpace(id)}
		if refIndex := strings.LastIndex(rest, " ref="); refIndex >= 0 {
			request.Ref = strings.TrimSpace(rest[refIndex+len(" ref="):])
			rest = rest[:refIndex]
		} else if strings.HasPrefix(rest, "ref=") {
			request.Ref, rest = strings.TrimSpace(strings.TrimPrefix(rest, "ref=")), ""
		}
		request.Reason = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(rest), "reason="))
		if request.CheckID != "" && request.Reason != "" {
			requests = append(requests, request)
		}
	}
	return requests
}

// GateInput is everything Evaluate needs; it performs no I/O.
type GateInput struct {
	Policy        policy.Policy
	Head          []domain.Issue
	Facts         map[string]CheckFact
	Base          []domain.Issue
	Comparison    domain.Comparison
	Changed       map[string]bool
	Lanes         []domain.LaneHealth
	Overrides     []OverrideRequest
	OverrideActor string
	Codeowners    []byte
	BaseRevision  string
	HeadRevision  string
	Unavailable   error
	Banners       []string
	// BaseGaps names the scanner families the base side did not inspect; an
	// `introduced` finding there is `unknown` and never gates.
	BaseGaps BaseGaps
}

// Acceptance is an override the gate accepted; the caller stores it as a disposition.
type Acceptance struct {
	IssueID string
	CheckID string
	Actor   string
	Reason  string
	Ref     string
}

// GateOutput is the pure verdict.
type GateOutput struct {
	Gate       domain.Gate
	Issues     []domain.AttributedIssue
	WouldBlock []domain.WouldBlock
	Accepted   []Acceptance
	Pre        int
}

func issuePath(issue domain.Issue) string {
	if issue.Anchor.Path != "" {
		return issue.Anchor.Path
	}
	if len(issue.Files) > 0 {
		return issue.Files[0]
	}
	return ""
}

func exactKey(issue domain.Issue) string {
	a := issue.Anchor
	if a.Secret {
		return a.RuleKey + "\x00secret\x00" + a.Fingerprint
	}
	return a.RuleKey + "\x00" + a.Path + "\x00" + a.Symbol + "\x00" + a.Fingerprint
}

func looseKey(issue domain.Issue) string {
	return issue.Anchor.RuleKey + "\x00" + issuePath(issue)
}

// Attribute classifies head issues against base issues. Exact identity
// matches come first, then a one-to-one pairing on rule and path (evidence
// drift). A match to a resolved base issue, or a head issue already marked
// regressed, is `regressed`. Without a baseline every issue is `unknown`.
func Attribute(head, base []domain.Issue, comparison domain.Comparison) map[string]domain.Attribution {
	out := make(map[string]domain.Attribution, len(head))
	if comparison == domain.ComparisonNone {
		for _, issue := range head {
			out[issue.IssueID] = domain.AttributionUnknown
		}
		return out
	}
	taken := make([]bool, len(base))
	matched := map[string]int{}
	pair := func(key func(domain.Issue) string, secretOnlyExact bool) {
		index := map[string][]int{}
		for i, b := range base {
			if !taken[i] {
				index[key(b)] = append(index[key(b)], i)
			}
		}
		for _, issue := range head {
			if _, done := matched[issue.IssueID]; done {
				continue
			}
			if secretOnlyExact && issue.Anchor.Secret {
				continue
			}
			list := index[key(issue)]
			if len(list) == 0 {
				continue
			}
			// Prefer a live base issue so a resolved twin does not mask debt.
			pick := 0
			for n, i := range list {
				if base[i].State != domain.IssueResolved {
					pick = n
					break
				}
			}
			taken[list[pick]] = true
			matched[issue.IssueID] = list[pick]
			index[key(issue)] = append(list[:pick:pick], list[pick+1:]...)
		}
	}
	pair(exactKey, false)
	pair(looseKey, true)
	for _, issue := range head {
		i, found := matched[issue.IssueID]
		switch {
		case found && base[i].State == domain.IssueResolved, !found && issue.State == domain.IssueRegressed:
			out[issue.IssueID] = domain.AttributionRegressed
		case found:
			out[issue.IssueID] = domain.AttributionPreExisting
		default:
			out[issue.IssueID] = domain.AttributionIntroduced
		}
	}
	return out
}

func dispositioned(state domain.IssueState) bool {
	return state == domain.IssueAccepted || state == domain.IssueFalsePositive || state == domain.IssueWontFix || state == domain.IssueResolved
}

// PathRoleVersion versions the path-role classifier of the deterministic
// baseline (design §6: stored with every ranking disclosure).
const PathRoleVersion = "path-role/v1"

var pathMarkers = []struct{ marker, role string }{
	{"/testdata/", "fixture"}, {"/test/", "test"}, {"/tests/", "test"}, {"_test.", "test"}, {"test_", "test"},
	{"/vendor/", "vendor"}, {"/node_modules/", "vendor"}, {"generated", "generated"}, {".md", "docs"}, {"/docs/", "docs"}, {"/fixtures/", "fixture"},
}

// pathRoleName classifies a path in code: source, test, fixture, generated,
// vendor or docs. Non-source roles halve the baseline weight.
func pathRoleName(path string) string {
	lower := "/" + strings.ToLower(path)
	for _, entry := range pathMarkers {
		if strings.Contains(lower, entry.marker) {
			return entry.role
		}
	}
	return "source"
}

func pathRole(path string) float64 {
	if pathRoleName(path) != "source" {
		return 0.5
	}
	return 1
}

// Evaluate applies attribution, policy, overrides and lane health.
func Evaluate(in GateInput) GateOutput {
	profile := in.Policy.Profile()
	out := GateOutput{Issues: []domain.AttributedIssue{}, WouldBlock: []domain.WouldBlock{}}
	gate := domain.Gate{Profile: profile, Comparison: in.Comparison, Blocking: []string{}}
	attribution := Attribute(in.Head, in.Base, in.Comparison)
	var candidates []int
	gapHits := map[domain.Family]int{}
	gapReason := map[domain.Family]string{}
	for _, issue := range in.Head {
		path := issuePath(issue)
		fact := in.Facts[issue.IssueID]
		att := attribution[issue.IssueID]
		if reason, gap := in.BaseGaps[issue.Family]; gap && att == domain.AttributionIntroduced && !strings.HasPrefix(fact.CheckID, "guardrails.") {
			// Guardrails findings come from the base tree diff, not a base lane scan.
			att = domain.AttributionUnknown
			gapHits[issue.Family]++
			gapReason[issue.Family] = reason
		}
		item := domain.AttributedIssue{
			IssueID: issue.IssueID, CheckID: fact.CheckID, RuleVersion: fact.RuleVersion, Path: path, Title: issue.Title,
			Severity: issue.EffectiveSeverity, State: issue.State, Attribution: att, ChangedInPR: in.Changed[path],
		}
		weight := float64(issue.EffectiveSeverity.Rank())
		changed := 1.0
		if item.ChangedInPR {
			changed = 2
		}
		item.RankingScore = weight * changed * pathRole(path)
		item.RankingReason = fmt.Sprintf("severity %.0f x changed-in-pr %.0f x path-role %.1f", weight, changed, pathRole(path))
		status := in.Policy.CheckStatus(fact.CheckID, fact.RuleVersion)
		item.CheckStatus = string(status)
		if att == domain.AttributionPreExisting {
			out.Pre++
		}
		if (att == domain.AttributionIntroduced || att == domain.AttributionRegressed) && !dispositioned(issue.State) && issue.Claim == "" {
			switch status {
			case policy.Active, policy.ActiveProvisional:
				item.Blocking = true
			case policy.Shadow:
				item.WouldBlock = true
			}
			if item.Blocking || item.WouldBlock {
				candidates = append(candidates, len(out.Issues))
			}
		}
		out.Issues = append(out.Issues, item)
	}
	// Overrides: only a maintainer's request takes effect; verdicts are kept.
	if len(in.Overrides) > 0 {
		ok, why := in.Policy.AcceptOverride(in.OverrideActor, in.Codeowners)
		for _, request := range in.Overrides {
			result := domain.OverrideResult{CheckID: request.CheckID, Actor: in.OverrideActor, Accepted: ok, Reason: why}
			if request.All {
				result.CheckID = "*"
			}
			if ok {
				for _, index := range candidates {
					item := &out.Issues[index]
					if !request.All && item.CheckID != request.CheckID {
						continue
					}
					if item.Overridden {
						continue
					}
					item.Overridden, item.Blocking, item.WouldBlock = true, false, false
					result.IssueIDs = append(result.IssueIDs, item.IssueID)
					out.Accepted = append(out.Accepted, Acceptance{IssueID: item.IssueID, CheckID: item.CheckID, Actor: in.OverrideActor, Reason: request.Reason, Ref: request.Ref})
				}
			}
			gate.Overrides = append(gate.Overrides, result)
		}
	}
	sort.SliceStable(out.Issues, func(i, j int) bool {
		if out.Issues[i].RankingScore != out.Issues[j].RankingScore {
			return out.Issues[i].RankingScore > out.Issues[j].RankingScore
		}
		return out.Issues[i].IssueID < out.Issues[j].IssueID
	})
	gated := false
	for _, item := range out.Issues {
		if item.Blocking {
			gated = true
			gate.Blocking = append(gate.Blocking, item.IssueID)
		}
		if item.WouldBlock {
			out.WouldBlock = append(out.WouldBlock, domain.WouldBlock{
				ID: wouldBlockID(in.BaseRevision, in.HeadRevision, item.CheckID, item.IssueID), HeadRevision: in.HeadRevision, BaseRevision: in.BaseRevision,
				IssueID: item.IssueID, CheckID: item.CheckID, RuleVersion: item.RuleVersion, Path: item.Path, Attribution: item.Attribution,
				Profile: profile, Triage: domain.TriageUndecided,
			})
		}
	}
	incomplete, reasons := Incomplete(in.Lanes)
	gate.Reasons = append(append(gate.Reasons, reasons...), in.Banners...)
	families := make([]string, 0, len(gapHits))
	for family := range gapHits {
		families = append(families, string(family))
	}
	sort.Strings(families)
	for _, family := range families {
		gate.Reasons = append(gate.Reasons, fmt.Sprintf("%s; %d %s finding(s) not in base are attributed unknown, not introduced", gapReason[domain.Family(family)], gapHits[domain.Family(family)], family))
	}
	if in.Unavailable != nil {
		gate.Reasons = append(gate.Reasons, "unavailable: "+in.Unavailable.Error())
	}
	gate.ExitCode = ExitCode(in.Unavailable != nil, incomplete, gated)
	gate.Status = StatusOf(gate.ExitCode)
	out.Gate = gate
	return out
}

func wouldBlockID(base, head, check, issue string) string {
	sum := sha256.Sum256([]byte(strings.Join([]string{base, head, check, issue}, "\x00")))
	return "wb-" + hex.EncodeToString(sum[:8])
}

// Ranking is the ranking of introduced and regressed findings and its
// disclosure (design §6): which method produced it, why a fallback happened,
// and the versions every stored ranking carries.
type Ranking struct {
	// Method is the disclosure line: `deterministic (<reason>)`,
	// `jev (unvalidated, X1 pending)`, or a partial form naming n/N.
	Method             string             `json:"method"`
	Mode               domain.RankingMode `json:"mode"`
	Reason             string             `json:"reason,omitempty"`
	RankFormulaVersion string             `json:"rank_formula_version"`
	BaselineVersion    string             `json:"baseline_version"`
	JevQuality         string             `json:"jev_quality,omitempty"`
	// Thresholds is the status of the threshold record in force for the Jev
	// question (`provisional`, `ready`, or `provisional (stale: ...)`); empty
	// when Jev did not run.
	Thresholds string        `json:"thresholds,omitempty"`
	Items      []RankingItem `json:"items"`
}

// RankingItem is one ranked issue. Ranking is the per-finding disclosure.
// Score is the deterministic baseline's score and is absent for Jev-ranked
// items: design §6 never shows Jev's probability (it stays in the state store
// for X1), only the position and its disclosure.
type RankingItem struct {
	IssueID         string             `json:"issueId"`
	Position        int                `json:"position"`
	Score           *float64           `json:"score,omitempty"`
	Reason          string             `json:"reason"`
	Mode            domain.RankingMode `json:"mode"`
	Ranking         string             `json:"ranking"`
	BaselineScore   float64            `json:"baselineScore"`
	QuestionVersion string             `json:"questionVersion,omitempty"`
}

const summaryLimit = 60

// Summary renders the Markdown PR summary, never longer than 60 lines.
func Summary(gate domain.Gate, ranking Ranking, issues []domain.AttributedIssue, wouldBlock []domain.WouldBlock, pre int, health []domain.LaneHealth) string {
	var lines []string
	add := func(format string, args ...any) { lines = append(lines, fmt.Sprintf(format, args...)) }
	add("## Scout — %s (exit %d)", gate.Status, gate.ExitCode)
	add("")
	add("Profile `%s`; comparison `%s`; ranking: %s.", gate.Profile, gate.Comparison, ranking.Method)
	for _, reason := range gate.Reasons {
		add("> %s", oneLine(reason))
	}
	var changed []domain.AttributedIssue
	unknown := 0
	for _, item := range issues {
		switch item.Attribution {
		case domain.AttributionIntroduced, domain.AttributionRegressed:
			changed = append(changed, item)
		case domain.AttributionUnknown:
			unknown++
		}
	}
	add("")
	add("### Introduced / regressed (%d)", len(changed))
	if len(changed) == 0 {
		add("None.")
	}
	for index, item := range changed {
		if index == 12 {
			add("- … %d more (see scout-assessment.json)", len(changed)-index)
			break
		}
		flag := ""
		switch {
		case item.Blocking:
			flag = " **blocks**"
		case item.Overridden:
			flag = " overridden"
		case item.WouldBlock:
			flag = " would block"
		case item.Claim != "":
			flag = " (" + string(item.Claim) + ", report-only)"
		}
		add("- `%s` %s `%s` %s [%s]%s", item.CheckID, item.Attribution, item.Path, oneLine(item.Title), item.Severity, flag)
	}
	add("")
	add("### Would block if activated (%d)", len(wouldBlock))
	if len(wouldBlock) == 0 {
		add("None.")
	}
	for index, row := range wouldBlock {
		if index == 8 {
			add("- … %d more", len(wouldBlock)-index)
			break
		}
		add("- `%s` in `%s` (triage: %s)", row.CheckID, row.Path, row.Triage)
	}
	for _, override := range gate.Overrides {
		verdict := "rejected"
		if override.Accepted {
			verdict = "accepted"
		}
		add("- override `%s` by `%s`: %s (%s)", override.CheckID, override.Actor, verdict, oneLine(override.Reason))
	}
	add("")
	add("Pre-existing: %d. Unknown attribution: %d.", pre, unknown)
	add("")
	add("### Lane health")
	add("| lane | mode | required | status |")
	add("|---|---|---|---|")
	for _, h := range health {
		reason := ""
		if h.Reason != "" {
			reason = " — " + oneLine(h.Reason)
		}
		add("| %s | %s | %t | %s%s |", h.Lane, h.Mode, h.Required, h.Status, reason)
	}
	if len(lines) > summaryLimit {
		lines = append(lines[:summaryLimit-1], "… summary truncated; see scout-assessment.json")
	}
	return strings.Join(lines, "\n") + "\n"
}

func oneLine(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 160 {
		value = value[:157] + "..."
	}
	return value
}
