package engine

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"scout/internal/decisions"
	"scout/internal/domain"
	"scout/internal/guardrails"
	"scout/internal/issues"
	"scout/internal/measures"
	"scout/internal/policy"
	"scout/internal/state"
)

// BaseState is a previously computed base: an imported baseline. Gaps names
// the scanner families the baseline run did not fully inspect; nil means the
// baseline carries no lane health (a legacy export), which forces a recompute.
type BaseState struct {
	Issues       []domain.Issue
	Revision     string
	ToolDigest   string
	PolicyDigest string
	Gaps         BaseGaps
}

// DiffConfig drives one differential run on the checked-out head.
type DiffConfig struct {
	Config        Config
	BaseRef       string
	HeadRef       string // must resolve to the checked-out HEAD; default HEAD
	OverrideActor string
	OverrideText  string
	// Baseline returns an imported baseline, or nil when none is present.
	Baseline func(ctx context.Context) (*BaseState, error)
	Log      io.Writer
	// JevKey is the Jev key, taken from the environment by the caller; only
	// its presence is ever reported.
	JevKey string
	// Fork marks a fork pull request: Jev never runs and the ranking falls
	// back to the deterministic baseline.
	Fork bool
	// NoEgress disables the Jev call for this run even when consented.
	NoEgress bool
	// JevClient is a test seam; nil builds the OpenRouter client from JevKey.
	JevClient *decisions.Client
	// DeltaTools locates lizard and jscpd for the delta rules; the zero value
	// finds them on PATH.
	DeltaTools measures.Tools
}

// DiffReport is the JSON payload of `scout diff`; gate, health, would_block
// and ranking are always present.
type DiffReport struct {
	Base       string            `json:"base"`
	Head       string            `json:"head"`
	MergeBase  string            `json:"mergeBase,omitempty"`
	Comparison domain.Comparison `json:"comparison"`
	// ComparisonReason says why the comparison is not the imported baseline.
	ComparisonReason string                   `json:"comparison_reason,omitempty"`
	Gate             domain.Gate              `json:"gate"`
	Health           []domain.LaneHealth      `json:"health"`
	WouldBlock       []domain.WouldBlock      `json:"would_block"`
	Ranking          Ranking                  `json:"ranking"`
	Issues           []domain.AttributedIssue `json:"issues"`
	PreExisting      int                      `json:"preExisting"`
	Changed          []string                 `json:"changedFiles"`
}

// UnavailableReport is the report for a run that could not start or finish.
func UnavailableReport(err error) DiffReport {
	gate := domain.Gate{Status: StatusOf(ExitUnavailable), ExitCode: ExitUnavailable, Comparison: domain.ComparisonNone, Blocking: []string{}}
	if err != nil {
		gate.Reasons = []string{"unavailable: " + err.Error()}
	}
	return DiffReport{Comparison: domain.ComparisonNone, Gate: gate, Health: []domain.LaneHealth{}, WouldBlock: []domain.WouldBlock{}, Ranking: DeterministicRanking(nil, "run unavailable"), Issues: []domain.AttributedIssue{}, Changed: []string{}}
}

func git(ctx context.Context, root string, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...)
	command.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	var stderr bytes.Buffer
	command.Stderr = &stderr
	out, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func revParse(ctx context.Context, root, ref string) (string, error) {
	out, err := git(ctx, root, "rev-parse", "--verify", "--quiet", ref+"^{commit}")
	if err != nil {
		return "", fmt.Errorf("cannot resolve %q to a commit: %w", ref, err)
	}
	return strings.TrimSpace(string(out)), nil
}

const maxGuardrailBytes = 1 << 20

// Diff attributes head issues against a same-policy base, applies the policy
// gate and persists would_block rows and accepted overrides in the head store.
func Diff(ctx context.Context, cfg DiffConfig) (DiffReport, error) {
	root, err := filepath.Abs(cfg.Config.Root)
	if err != nil {
		return UnavailableReport(err), err
	}
	cfg.Config.Root = root
	fail := func(err error) (DiffReport, error) { return UnavailableReport(err), err }
	if cfg.BaseRef == "" {
		return fail(errors.New("--base is required"))
	}
	headRef := cfg.HeadRef
	if headRef == "" {
		headRef = "HEAD"
	}
	baseSHA, err := revParse(ctx, root, cfg.BaseRef)
	if err != nil {
		return fail(err)
	}
	headSHA, err := revParse(ctx, root, headRef)
	if err != nil {
		return fail(err)
	}
	checkedOut, err := revParse(ctx, root, "HEAD")
	if err != nil {
		return fail(err)
	}
	if headSHA != checkedOut {
		return fail(fmt.Errorf("--head %s is not the checked-out HEAD; check it out so the head scan never touches another tree", headRef))
	}
	mbOut, err := git(ctx, root, "merge-base", baseSHA, headSHA)
	if err != nil {
		return fail(fmt.Errorf("merge base of %s and %s: %w", cfg.BaseRef, headRef, err))
	}
	mergeBase := strings.TrimSpace(string(mbOut))
	pol, err := policy.Load(root, baseSHA)
	if err != nil {
		return fail(fmt.Errorf("load policy from base ref: %w", err))
	}
	cfg.Config = ApplyModelCLIPolicy(cfg.Config, pol)
	changed, err := changedFiles(ctx, root, mergeBase)
	if err != nil {
		return fail(err)
	}
	var codeowners []byte
	for _, path := range []string{".github/CODEOWNERS", "CODEOWNERS", "docs/CODEOWNERS"} {
		if out, showErr := git(ctx, root, "show", baseSHA+":"+path); showErr == nil {
			codeowners = out
			break
		}
	}

	guard := guardrailsLane(root, mergeBase, changed)
	delta := deltaLane(root, mergeBase, changed, cfg.DeltaTools)
	headResult, err := Assess(ctx, cfg.Config, Options{Trigger: domain.TriggerPR, Inline: []InlineLane{guard, delta}, Measures: &MeasuresOptions{Only: append([]string{}, changed...)}})
	if err != nil {
		return fail(fmt.Errorf("head assessment: %w", err))
	}
	store, err := state.Open(root)
	if err != nil {
		return fail(err)
	}
	defer store.Close()
	plan, err := store.LoadPlan(ctx, headResult.Run.PlanID)
	if err != nil {
		return fail(err)
	}
	// Promoted delta issues are descriptive: they are partitioned out here so
	// they can never be an input to attribution, the gate or would_block.
	var headIssues, deltaIssues []domain.Issue
	for _, issue := range headResult.Issues {
		if issue.Claim != "" {
			deltaIssues = append(deltaIssues, issue)
		} else {
			headIssues = append(headIssues, issue)
		}
	}
	facts := make(map[string]CheckFact, len(headIssues))
	for _, issue := range headIssues {
		observations, listErr := store.ListIssueObservations(ctx, issue.IssueID)
		if listErr != nil {
			return fail(listErr)
		}
		if len(observations) == 0 {
			continue
		}
		last := observations[len(observations)-1]
		facts[issue.IssueID] = CheckFact{CheckID: CheckIDOf(last.Source, last.RuleID), RuleVersion: last.RuleVersion}
	}

	var banners []string
	baseIssues, baseGaps, comparison, banners, err := resolveBase(ctx, cfg, root, mergeBase, plan, headResult)
	if err != nil {
		return fail(err)
	}

	changedSet := make(map[string]bool, len(changed))
	for _, path := range changed {
		changedSet[path] = true
	}
	overrides := ParseOverrides(cfg.OverrideText)
	output := Evaluate(GateInput{
		Policy: pol, Head: headIssues, Facts: facts, Base: withoutClaimed(baseIssues), Comparison: comparison, Changed: changedSet,
		Lanes: headResult.Lanes, Overrides: overrides, OverrideActor: cfg.OverrideActor, Codeowners: codeowners,
		BaseRevision: mergeBase, HeadRevision: headSHA, Banners: banners, BaseGaps: baseGaps,
	})
	for index := range output.WouldBlock {
		output.WouldBlock[index].RecordedAt = time.Now().UTC()
		if err := store.RecordWouldBlock(ctx, output.WouldBlock[index]); err != nil {
			return fail(err)
		}
	}
	for _, accepted := range output.Accepted {
		pr := accepted.Ref
		if pr == "" {
			pr = "none"
		}
		actor, actorErr := issues.SanitizeText(accepted.Actor)
		reason, reasonErr := issues.SanitizeText(accepted.Reason)
		if actorErr != nil || reasonErr != nil || actor == "" || reason == "" {
			return fail(errors.New("override actor or reason is unsafe to persist"))
		}
		if _, err := store.TriageIssue(ctx, accepted.IssueID, domain.IssueAccepted, "", fmt.Sprintf("scout-override pr=%s", pr), actor, reason); err != nil {
			return fail(fmt.Errorf("record override disposition: %w", err))
		}
	}
	if cfg.Log != nil {
		for _, override := range output.Gate.Overrides {
			if !override.Accepted {
				fmt.Fprintf(cfg.Log, "scout: override for %s by %q rejected: %s\n", override.CheckID, override.Actor, override.Reason)
			}
		}
	}
	if output.WouldBlock == nil {
		output.WouldBlock = []domain.WouldBlock{}
	}
	if changed == nil {
		changed = []string{}
	}
	reportIssues := append(output.Issues, deltaAttributed(pol, deltaIssues)...)
	ranking, decisionsHealth, err := RankingOf(ctx, RankingInput{
		Issues: output.Issues, Root: root, RunID: headResult.Run.RunID, Trigger: domain.TriggerPR, Policy: pol.Decisions(),
		Key: cfg.JevKey, Fork: cfg.Fork, NoEgress: cfg.NoEgress, Store: store, Client: cfg.JevClient, ReadThresholds: CommittedThresholds(root, baseSHA),
	})
	if err != nil {
		return fail(fmt.Errorf("ranking: %w", err))
	}
	return DiffReport{
		Base: baseSHA, Head: headSHA, MergeBase: mergeBase, Comparison: comparison, ComparisonReason: comparisonReason(comparison, banners), Gate: output.Gate, Health: ApplyDecisionsHealth(headResult.Lanes, decisionsHealth),
		WouldBlock: output.WouldBlock, Ranking: ranking, Issues: reportIssues, PreExisting: output.Pre, Changed: changed,
	}, nil
}

// withoutClaimed drops rule-promoted issues: a baseline never attributes them.
func withoutClaimed(list []domain.Issue) []domain.Issue {
	out := make([]domain.Issue, 0, len(list))
	for _, issue := range list {
		if issue.Claim == "" {
			out = append(out, issue)
		}
	}
	return out
}

// deltaAttributed renders promoted delta issues for the report. By
// construction they are `introduced` relative to the merge base; their check id
// is never blocking-eligible, so they are neither blocking nor would-block.
func deltaAttributed(pol policy.Policy, list []domain.Issue) []domain.AttributedIssue {
	out := make([]domain.AttributedIssue, 0, len(list))
	for _, issue := range list {
		_, version, _ := strings.Cut(issue.Rule, "@")
		out = append(out, domain.AttributedIssue{
			IssueID: issue.IssueID, CheckID: issue.Rule, Claim: issue.Claim, RuleVersion: version, Path: issue.Anchor.Path, Title: issue.Title,
			Severity: issue.EffectiveSeverity, State: issue.State, Attribution: domain.AttributionIntroduced, ChangedInPR: true,
			CheckStatus:   string(pol.CheckStatus(issue.Rule, version)),
			RankingReason: "descriptive delta; report-only, not ranked",
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].IssueID < out[j].IssueID })
	return out
}

// BaseGaps maps a scanner family to why the base side did not inspect it
// (lane blocked or failed, or imported baseline lane not fully inspected).
// A head finding in such a family is absent from base for the wrong reason, so
// it is never `introduced` (design §7.3).
type BaseGaps map[domain.Family]string

// gapsOfLanes derives the base gaps of a paired recompute from its lane
// health: a scanner lane that is blocked or failed did not inspect the base.
func gapsOfLanes(health []domain.LaneHealth) BaseGaps {
	gaps := BaseGaps{}
	for _, h := range health {
		if h.Status != domain.LaneBlocked && h.Status != domain.LaneFailed {
			continue
		}
		for _, lane := range lanes {
			if lane.ID == h.Lane && lane.Family != "" {
				reason := "base lane " + h.Lane + " " + string(h.Status)
				if h.Reason != "" {
					reason += ": " + h.Reason
				}
				gaps[lane.Family] = reason
			}
		}
	}
	return gaps
}

// resolveBase returns the base issues and their coverage gaps: an imported
// baseline when its digests and revision match head and it carries lane
// health, else a paired scan of the merge-base worktree. A baseline without
// lane health cannot show which lanes inspected the base, so it is recomputed.
func resolveBase(ctx context.Context, cfg DiffConfig, root, mergeBase string, headPlan domain.Plan, head Result) ([]domain.Issue, BaseGaps, domain.Comparison, []string, error) {
	var banners []string
	if cfg.Baseline != nil {
		imported, err := cfg.Baseline(ctx)
		if err != nil {
			return nil, nil, "", nil, fmt.Errorf("load baseline: %w", err)
		}
		switch {
		case imported == nil:
		case imported.ToolDigest != headPlan.ToolDigest:
			banners = append(banners, "baseline tool digest "+short(imported.ToolDigest)+" differs from head "+short(headPlan.ToolDigest)+"; comparison recomputed")
		case imported.PolicyDigest != headPlan.PolicyDigest:
			banners = append(banners, "baseline policy digest "+short(imported.PolicyDigest)+" differs from head "+short(headPlan.PolicyDigest)+"; comparison recomputed")
		case imported.Revision != mergeBase:
			banners = append(banners, "baseline revision "+short(imported.Revision)+" is not the merge base "+short(mergeBase)+"; comparison recomputed")
		case imported.Gaps == nil:
			banners = append(banners, "baseline carries no lane health; comparison recomputed")
		default:
			return imported.Issues, imported.Gaps, domain.ComparisonBaseline, banners, nil
		}
		if imported == nil {
			banners = append(banners, "no baseline published; comparison recomputed")
		}
	} else {
		banners = append(banners, "no baseline remote configured; comparison recomputed")
	}
	baseResult, err := pairedBaseScan(ctx, cfg, root, mergeBase)
	if err != nil {
		banners = append(banners, "no baseline available ("+err.Error()+"); nothing is classified introduced")
		return nil, nil, domain.ComparisonNone, banners, nil
	}
	return baseResult.Issues, gapsOfLanes(baseResult.Lanes), domain.ComparisonRecomputed, banners, nil
}

// pairedBaseScan scans the base commit in a temporary worktree so the user's
// checkout is never modified.
func pairedBaseScan(ctx context.Context, cfg DiffConfig, root, revision string) (Result, error) {
	dir, err := os.MkdirTemp("", "scout-base-")
	if err != nil {
		return Result{}, err
	}
	worktree := filepath.Join(dir, "tree")
	if _, err := git(ctx, root, "worktree", "add", "--detach", "--quiet", worktree, revision); err != nil {
		_ = os.RemoveAll(dir)
		return Result{}, err
	}
	defer func() {
		_, _ = git(context.WithoutCancel(ctx), root, "worktree", "remove", "--force", worktree)
		_, _ = git(context.WithoutCancel(ctx), root, "worktree", "prune")
		_ = os.RemoveAll(dir)
	}()
	baseConfig := cfg.Config
	baseConfig.Root = worktree
	result, err := Assess(ctx, baseConfig, Options{Trigger: domain.TriggerPR})
	if err != nil {
		return Result{}, fmt.Errorf("base assessment: %w", err)
	}
	return result, nil
}

// changedFiles lists files changed between the merge base and the working tree.
func changedFiles(ctx context.Context, root, mergeBase string) ([]string, error) {
	out, err := git(ctx, root, "diff", "--name-only", "-z", "--diff-filter=ACMR", mergeBase)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var files []string
	for _, name := range strings.Split(string(out), "\x00") {
		if name != "" && !seen[name] {
			seen[name] = true
			files = append(files, name)
		}
	}
	sort.Strings(files)
	return files, nil
}

// guardrailsLane wires the guardrails mechanisms (SUPPRESS, TESTQ, CONFIG) as
// a pr-trigger blocking-eligible lane: base from the merge base, head from
// the working tree. Config erosion compares whole trees, so it runs even when
// no file was added or modified (a PR that only deletes a config).
func guardrailsLane(root, mergeBase string, changed []string) InlineLane {
	return InlineLane{Lane: LaneGuardrails, Units: len(changed), Run: func(ctx context.Context, runID string) (InlineResult, error) {
		baseTree, err := newGitTree(ctx, root, mergeBase)
		if err != nil {
			return InlineResult{}, err
		}
		headTree, err := newWorkTree(ctx, root)
		if err != nil {
			return InlineResult{}, err
		}
		now := time.Now().UTC()
		var observations []domain.Observation
		for _, finding := range guardrails.ChangesAll(baseTree, headTree, changed) {
			quote, err := issues.SanitizeText(strings.TrimSpace(finding.Evidence.Head))
			if err != nil || quote == "" {
				quote = "[guardrail evidence elided]"
			}
			category := "suppression"
			if strings.HasPrefix(finding.CheckID, "guardrails.CONFIG.") {
				category = "guardrail-config"
				if base, err := issues.SanitizeText(strings.TrimSpace(finding.Evidence.Base)); err == nil && base != "" {
					quote += " (base: " + base + ")"
				}
			} else if strings.HasPrefix(finding.CheckID, "guardrails.TESTQ.") {
				category = "focused-test"
			}
			sum := sha256.Sum256([]byte(strings.Join([]string{runID, finding.CheckID, finding.Path, fmt.Sprint(finding.Line), finding.Mechanism}, "\x00")))
			observations = append(observations, domain.Observation{
				ObservationID: "guardrails:" + hex.EncodeToString(sum[:12]), RunID: runID, Source: "guardrails",
				Family: domain.FamilyQuality, OriginalCategory: category, RuleID: finding.CheckID, RuleVersion: finding.Version,
				SourceSeverity: domain.SeverityLow, EffectiveSeverity: domain.SeverityLow, Confidence: domain.ConfidenceHigh,
				Mechanism: finding.Mechanism, Validated: true, ObservedAt: now,
				Evidence: []domain.Evidence{{Location: domain.SourceRange{Path: finding.Path, StartLine: finding.Line, EndLine: finding.Line}, RedactedQuote: quote}},
			})
		}
		return InlineResult{Observations: observations}, nil
	}}
}
