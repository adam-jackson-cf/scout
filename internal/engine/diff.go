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

	"scout/internal/domain"
	"scout/internal/guardrails"
	"scout/internal/issues"
	"scout/internal/policy"
	"scout/internal/state"
)

// BaseState is a previously computed base: an imported baseline.
type BaseState struct {
	Issues       []domain.Issue
	Revision     string
	ToolDigest   string
	PolicyDigest string
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
	return DiffReport{Comparison: domain.ComparisonNone, Gate: gate, Health: []domain.LaneHealth{}, WouldBlock: []domain.WouldBlock{}, Ranking: Ranking{Method: RankingLabel, Items: []RankingItem{}}, Issues: []domain.AttributedIssue{}, Changed: []string{}}
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
	mergeBase := baseSHA
	if out, mbErr := git(ctx, root, "merge-base", baseSHA, headSHA); mbErr == nil {
		mergeBase = strings.TrimSpace(string(out))
	}
	pol, err := policy.Load(root, baseSHA)
	if err != nil {
		return fail(fmt.Errorf("load policy from base ref: %w", err))
	}
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

	guard := guardrailsLane(ctx, root, mergeBase, changed)
	headResult, err := Assess(ctx, cfg.Config, Options{Trigger: domain.TriggerPR, Inline: []InlineLane{guard}})
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
	facts := make(map[string]CheckFact, len(headResult.Issues))
	for _, issue := range headResult.Issues {
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
	baseIssues, comparison, banners, err := resolveBase(ctx, cfg, root, mergeBase, plan, headResult)
	if err != nil {
		return fail(err)
	}

	changedSet := make(map[string]bool, len(changed))
	for _, path := range changed {
		changedSet[path] = true
	}
	overrides := ParseOverrides(cfg.OverrideText)
	output := Evaluate(GateInput{
		Policy: pol, Head: headResult.Issues, Facts: facts, Base: baseIssues, Comparison: comparison, Changed: changedSet,
		Lanes: headResult.Lanes, Overrides: overrides, OverrideActor: cfg.OverrideActor, Codeowners: codeowners,
		BaseRevision: mergeBase, HeadRevision: headSHA, Banners: banners,
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
	return DiffReport{
		Base: baseSHA, Head: headSHA, MergeBase: mergeBase, Comparison: comparison, ComparisonReason: comparisonReason(comparison, banners), Gate: output.Gate, Health: headResult.Lanes,
		WouldBlock: output.WouldBlock, Ranking: RankingOf(output.Issues), Issues: output.Issues, PreExisting: output.Pre, Changed: changed,
	}, nil
}

// resolveBase returns the base issues: an imported baseline when its digests
// and revision match head, else a paired scan of the merge-base worktree.
func resolveBase(ctx context.Context, cfg DiffConfig, root, mergeBase string, headPlan domain.Plan, head Result) ([]domain.Issue, domain.Comparison, []string, error) {
	var banners []string
	if cfg.Baseline != nil {
		imported, err := cfg.Baseline(ctx)
		if err != nil {
			return nil, "", nil, fmt.Errorf("load baseline: %w", err)
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
			return imported.Issues, domain.ComparisonBaseline, banners, nil
		}
		if imported == nil {
			banners = append(banners, "no baseline published; comparison recomputed")
		}
	} else {
		banners = append(banners, "no baseline remote configured; comparison recomputed")
	}
	baseIssues, err := pairedBaseScan(ctx, cfg, root, mergeBase)
	if err != nil {
		banners = append(banners, "no baseline available ("+err.Error()+"); nothing is classified introduced")
		return nil, domain.ComparisonNone, banners, nil
	}
	return baseIssues, domain.ComparisonRecomputed, banners, nil
}

// pairedBaseScan scans the base commit in a temporary worktree so the user's
// checkout is never modified.
func pairedBaseScan(ctx context.Context, cfg DiffConfig, root, revision string) ([]domain.Issue, error) {
	dir, err := os.MkdirTemp("", "scout-base-")
	if err != nil {
		return nil, err
	}
	worktree := filepath.Join(dir, "tree")
	if _, err := git(ctx, root, "worktree", "add", "--detach", "--quiet", worktree, revision); err != nil {
		_ = os.RemoveAll(dir)
		return nil, err
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
		return nil, fmt.Errorf("base assessment: %w", err)
	}
	return result.Issues, nil
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

// guardrailsLane wires guardrails.SUPPRESS as a pr-trigger blocking-eligible
// lane over the changed files: base from the merge base, head from the tree.
func guardrailsLane(ctx context.Context, root, mergeBase string, changed []string) InlineLane {
	return InlineLane{Lane: LaneGuardrails, Units: len(changed), Run: func(ctx context.Context, runID string) ([]domain.Observation, error) {
		if len(changed) == 0 {
			return nil, nil
		}
		base := map[string][]byte{}
		head := map[string][]byte{}
		for _, path := range changed {
			content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
			if err != nil || len(content) > maxGuardrailBytes {
				continue
			}
			head[path] = content
			if old, showErr := git(ctx, root, "show", mergeBase+":"+path); showErr == nil && len(old) <= maxGuardrailBytes {
				base[path] = old
			}
		}
		now := time.Now().UTC()
		var observations []domain.Observation
		for _, finding := range guardrails.SuppressAdded(base, head, changed) {
			quote, err := issues.SanitizeText(strings.TrimSpace(finding.Evidence.Head))
			if err != nil || quote == "" {
				quote = "[suppression line elided]"
			}
			sum := sha256.Sum256([]byte(strings.Join([]string{runID, finding.CheckID, finding.Path, fmt.Sprint(finding.Line)}, "\x00")))
			observations = append(observations, domain.Observation{
				ObservationID: "guardrails:" + hex.EncodeToString(sum[:12]), RunID: runID, Source: "guardrails",
				Family: domain.FamilyQuality, OriginalCategory: "suppression", RuleID: finding.CheckID, RuleVersion: finding.Version,
				SourceSeverity: domain.SeverityLow, EffectiveSeverity: domain.SeverityLow, Confidence: domain.ConfidenceHigh,
				Mechanism: finding.Mechanism, Validated: true, ObservedAt: now,
				Evidence: []domain.Evidence{{Location: domain.SourceRange{Path: finding.Path, StartLine: finding.Line, EndLine: finding.Line}, RedactedQuote: quote}},
			})
		}
		return observations, nil
	}}
}
