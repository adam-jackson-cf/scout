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

	"scout/internal/areas"
	"scout/internal/association"
	"scout/internal/domain"
	"scout/internal/guardrails"
	"scout/internal/issues"
	"scout/internal/measures"
	"scout/internal/policy"
	"scout/internal/state"
)

// BaseState is a previously computed base: an imported baseline. Gaps names
// the scanner families the baseline run did not fully inspect; it is never
// nil (empty means all were inspected).
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
	PendingAreas     []areas.Area             `json:"pendingAreas"`
	AssociationGuard association.GuardState   `json:"associationGuard"`
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
	report, err := diffAtRoot(ctx, cfg)
	if err != nil {
		return UnavailableReport(err), err
	}
	return report, nil
}

type diffRevisions struct {
	headRef   string
	baseSHA   string
	headSHA   string
	mergeBase string
}

type diffRun struct {
	ctx        context.Context
	cfg        DiffConfig
	rev        diffRevisions
	pol        policy.Policy
	changed    []string
	codeowners []byte
	head       Result
	store      *state.Store
}

func resolveDiffRevisions(ctx context.Context, cfg DiffConfig) (diffRevisions, error) {
	root := cfg.Config.Root
	if cfg.BaseRef == "" {
		return diffRevisions{}, errors.New("--base is required")
	}
	rev := diffRevisions{headRef: cfg.HeadRef}
	if rev.headRef == "" {
		rev.headRef = "HEAD"
	}
	var err error
	if rev.baseSHA, err = revParse(ctx, root, cfg.BaseRef); err != nil {
		return diffRevisions{}, err
	}
	if rev.headSHA, err = revParse(ctx, root, rev.headRef); err != nil {
		return diffRevisions{}, err
	}
	checkedOut, err := revParse(ctx, root, "HEAD")
	if err != nil {
		return diffRevisions{}, err
	}
	if rev.headSHA != checkedOut {
		return diffRevisions{}, fmt.Errorf("--head %s is not the checked-out HEAD; check it out so the head scan never touches another tree", rev.headRef)
	}
	mbOut, err := git(ctx, root, "merge-base", rev.baseSHA, rev.headSHA)
	if err != nil {
		return diffRevisions{}, fmt.Errorf("merge base of %s and %s: %w", cfg.BaseRef, rev.headRef, err)
	}
	rev.mergeBase = strings.TrimSpace(string(mbOut))
	return rev, nil
}

func baseCodeowners(ctx context.Context, root, baseSHA string) []byte {
	for _, path := range []string{".github/CODEOWNERS", "CODEOWNERS", "docs/CODEOWNERS"} {
		if out, showErr := git(ctx, root, "show", baseSHA+":"+path); showErr == nil {
			return out
		}
	}
	return nil
}

func diffAtRoot(ctx context.Context, cfg DiffConfig) (report DiffReport, err error) {
	root := cfg.Config.Root
	rev, err := resolveDiffRevisions(ctx, cfg)
	if err != nil {
		return DiffReport{}, err
	}
	pol, err := policy.Load(root, rev.baseSHA)
	if err != nil {
		return DiffReport{}, fmt.Errorf("load policy from base ref: %w", err)
	}
	cfg.Config = ApplyModelCLIPolicy(cfg.Config, pol)
	changed, err := changedFiles(ctx, root, rev.mergeBase)
	if err != nil {
		return DiffReport{}, err
	}
	codeowners := baseCodeowners(ctx, root, rev.baseSHA)

	guard := guardrailsLane(root, rev.mergeBase, changed)
	delta := deltaLane(root, rev.mergeBase, changed, cfg.DeltaTools)
	headResult, err := Assess(ctx, cfg.Config, Options{Trigger: domain.TriggerPR, Inline: []InlineLane{guard, delta}, Measures: &MeasuresOptions{Only: append([]string{}, changed...)}})
	if err != nil {
		return DiffReport{}, fmt.Errorf("head assessment: %w", err)
	}
	store, err := state.Open(root)
	if err != nil {
		return DiffReport{}, err
	}
	defer func() { err = errors.Join(err, store.Close()) }()
	run := diffRun{ctx: ctx, cfg: cfg, rev: rev, pol: pol, changed: changed, codeowners: codeowners, head: headResult, store: store}
	return run.report()
}

func (r diffRun) report() (DiffReport, error) {
	root := r.cfg.Config.Root
	plan, err := r.store.LoadPlan(r.ctx, r.head.Run.PlanID)
	if err != nil {
		return DiffReport{}, err
	}
	headIssues, deltaIssues := partitionDeltaIssues(r.head.Issues)
	facts, err := r.checkFacts(headIssues)
	if err != nil {
		return DiffReport{}, err
	}

	var banners []string
	baseIssues, baseGaps, comparison, banners, err := resolveBase(r.ctx, r.cfg, root, r.rev.mergeBase, plan)
	if err != nil {
		return DiffReport{}, err
	}

	changedSet := make(map[string]bool, len(r.changed))
	for _, path := range r.changed {
		changedSet[path] = true
	}
	overrides := ParseOverrides(r.cfg.OverrideText)
	output := Evaluate(GateInput{
		Policy: r.pol, Head: headIssues, Facts: facts, Base: withoutClaimed(baseIssues), Comparison: comparison, Changed: changedSet,
		Lanes: r.head.Lanes, Overrides: overrides, OverrideActor: r.cfg.OverrideActor, Codeowners: r.codeowners,
		BaseRevision: r.rev.mergeBase, HeadRevision: r.rev.headSHA, Banners: banners, BaseGaps: baseGaps,
	})
	if err := r.recordGateOutput(output); err != nil {
		return DiffReport{}, err
	}
	r.logRejectedOverrides(output)
	if output.WouldBlock == nil {
		output.WouldBlock = []domain.WouldBlock{}
	}
	changed := r.changed
	if changed == nil {
		changed = []string{}
	}
	reportIssues := append(output.Issues, deltaAttributed(r.pol, deltaIssues)...)
	ranking, decisionsHealth, err := RankingOf(r.ctx, RankingInput{
		Issues: output.Issues, Root: root, RunID: r.head.Run.RunID, Trigger: domain.TriggerPR, Policy: r.pol.Decisions(),
		Key: r.cfg.JevKey, Fork: r.cfg.Fork, NoEgress: r.cfg.NoEgress, Store: r.store, ReadThresholds: CommittedThresholds(root, r.rev.baseSHA),
	})
	if err != nil {
		return DiffReport{}, fmt.Errorf("ranking: %w", err)
	}
	pending, err := r.pendingAreas()
	if err != nil {
		return DiffReport{}, err
	}
	return DiffReport{
		Base: r.rev.baseSHA, Head: r.rev.headSHA, MergeBase: r.rev.mergeBase, Comparison: comparison, ComparisonReason: comparisonReason(comparison, banners), Gate: output.Gate, Health: ApplyDecisionsHealth(r.head.Lanes, decisionsHealth),
		WouldBlock: output.WouldBlock, Ranking: ranking, Issues: reportIssues, PreExisting: output.Pre, Changed: changed,
		PendingAreas:     pending,
		AssociationGuard: r.head.AssociationGuard,
	}, nil
}

// Promoted delta issues are descriptive: they are partitioned out here so
// they can never be an input to attribution, the gate or would_block.
func partitionDeltaIssues(list []domain.Issue) (headIssues, deltaIssues []domain.Issue) {
	for _, issue := range list {
		if issue.Claim != "" {
			deltaIssues = append(deltaIssues, issue)
		} else {
			headIssues = append(headIssues, issue)
		}
	}
	return headIssues, deltaIssues
}

func (r diffRun) checkFacts(headIssues []domain.Issue) (map[string]CheckFact, error) {
	facts := make(map[string]CheckFact, len(headIssues))
	for _, issue := range headIssues {
		observations, listErr := r.store.ListIssueObservations(r.ctx, issue.IssueID)
		if listErr != nil {
			return nil, listErr
		}
		if len(observations) == 0 {
			continue
		}
		last := observations[len(observations)-1]
		facts[issue.IssueID] = CheckFact{CheckID: CheckIDOf(last.Source, last.RuleID), RuleVersion: last.RuleVersion}
	}
	return facts, nil
}

func (r diffRun) recordGateOutput(output GateOutput) error {
	for index := range output.WouldBlock {
		output.WouldBlock[index].RecordedAt = time.Now().UTC()
		if err := r.store.RecordWouldBlock(r.ctx, output.WouldBlock[index]); err != nil {
			return err
		}
	}
	for _, accepted := range output.Accepted {
		if err := r.recordAcceptedOverride(accepted); err != nil {
			return err
		}
	}
	return nil
}

func (r diffRun) recordAcceptedOverride(accepted Acceptance) error {
	pr := accepted.Ref
	if pr == "" {
		pr = "none"
	}
	actor, actorErr := issues.SanitizeText(accepted.Actor)
	reason, reasonErr := issues.SanitizeText(accepted.Reason)
	if actorErr != nil || reasonErr != nil || actor == "" || reason == "" {
		return errors.New("override actor or reason is unsafe to persist")
	}
	if _, err := r.store.TriageIssue(r.ctx, state.Triage{IssueID: accepted.IssueID, State: domain.IssueAccepted, Owner: "", DispositionReason: fmt.Sprintf("scout-override pr=%s", pr), Actor: actor, Reason: reason}); err != nil {
		return fmt.Errorf("record override disposition: %w", err)
	}
	return nil
}

func (r diffRun) logRejectedOverrides(output GateOutput) {
	if r.cfg.Log == nil {
		return
	}
	for _, override := range output.Gate.Overrides {
		if !override.Accepted {
			_, _ = fmt.Fprintf(r.cfg.Log, "scout: override for %s by %q rejected: %s\n", override.CheckID, override.Actor, override.Reason)
		}
	}
}

func (r diffRun) pendingAreas() ([]areas.Area, error) {
	allAreas, err := r.store.ListAreas(r.ctx)
	if err != nil {
		return nil, err
	}
	pending := make([]areas.Area, 0)
	for _, area := range allAreas {
		if area.Status == areas.StatusPending {
			pending = append(pending, area)
		}
	}
	return pending, nil
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
// it is never `introduced` (REQ-27).
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
// baseline when its digests and revision match head, else a paired scan of the
// merge-base worktree.
func resolveBase(ctx context.Context, cfg DiffConfig, root, mergeBase string, headPlan domain.Plan) ([]domain.Issue, BaseGaps, domain.Comparison, []string, error) {
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
