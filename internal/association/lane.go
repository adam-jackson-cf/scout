package association

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"scout/internal/areas"
	"scout/internal/clawpatch"
	"scout/internal/decisions"
	"scout/internal/domain"
	"scout/internal/modelcli"
	"scout/internal/policy"
	"scout/internal/state"
)

// LaneOptions supplies the push range and the independently consented Jev transport.
type LaneOptions struct {
	Base, Head, RunID string
	Mapping           policy.Association
	Config            decisions.Config
	Eligibility       decisions.Eligibility
	Client            *decisions.Client
	Fingerprint       Fingerprint
	FallbackConsent   bool
}

// Execute runs the provisional, advisory association lane. Skips do not mutate state.
func Execute(ctx context.Context, store *state.Store, root string, trigger domain.LaneTrigger, opts LaneOptions) ([]domain.LaneHealth, error) {
	healths := []domain.LaneHealth{
		{Lane: "association", Mode: domain.ModeSignal, Requires: domain.RequiresDecisionEgress, Status: domain.LaneSkipped},
		{Lane: "association-fallback", Mode: domain.ModeSignal, Requires: domain.RequiresModelCLI, Status: domain.LaneSkipped},
	}
	run := &laneRun{opts: opts, store: store, h: &healths[0], fallback: &healths[1]}
	reason, err := run.prepare(ctx, root, trigger)
	if err != nil {
		return healths, err
	}
	if reason != "" {
		run.h.Reason = reason
		run.fallback.Reason = reason
		return healths, nil
	}
	run.repo = Repo{Dir: root}
	run.budget = jevBudget{opts: opts}
	run.terra = NewTerraDecider(opts.Mapping.Fallback)
	run.fallback.Reason = fallbackSkipReason(opts.FallbackConsent)
	results, err := Run(ctx, RunInput{Repo: run.repo, Base: opts.Base, Head: opts.Head, Own: run.own, Decide: run.decide})
	if err != nil {
		return healths, err
	}
	run.recordMergeResults(results)
	if err := run.persist(ctx, results); err != nil {
		return healths, err
	}
	run.finishHealth()
	return healths, nil
}

type laneRun struct {
	opts                            LaneOptions
	store                           *state.Store
	h, fallback                     *domain.LaneHealth
	catalogue, active               []areas.Area
	own                             Ownership
	generation                      *time.Time
	repo                            Repo
	budget                          jevBudget
	terra                           *TerraDecider
	pendingCount, fallbackSucceeded int
	rows                            []domain.Association
}

func laneSkipReason(trigger domain.LaneTrigger, opts LaneOptions) string {
	switch {
	case trigger != domain.TriggerPush:
		return "trigger"
	case opts.Eligibility.Fork:
		return "fork"
	case !opts.Mapping.Enabled:
		return "disabled"
	}
	return jevSkipReason(opts)
}

func jevSkipReason(opts LaneOptions) string {
	switch {
	case !opts.Eligibility.HasKey || opts.Client == nil || opts.Client.Key == "":
		return "no jev key"
	case !opts.Config.EgressAcknowledged:
		return "no acknowledgement"
	case !opts.Config.Enabled || !opts.Eligibility.Enabled || !opts.Eligibility.EgressAllowed:
		return "egress not allowed"
	}
	return ""
}

func fallbackSkipReason(consent bool) string {
	if !consent {
		return "no consent"
	}
	if _, _, err := modelcli.Detect("codex"); err != nil {
		return "no model CLI"
	}
	return ""
}

func activeAreas(catalogue []areas.Area) []areas.Area {
	active := make([]areas.Area, 0, len(catalogue))
	for _, a := range catalogue {
		if a.Active() {
			active = append(active, a)
		}
	}
	return active
}

func (run *laneRun) prepare(ctx context.Context, root string, trigger domain.LaneTrigger) (string, error) {
	if reason := laneSkipReason(trigger, run.opts); reason != "" {
		return reason, nil
	}
	catalogue, err := run.store.ListAreas(ctx)
	if err != nil {
		return "", err
	}
	run.catalogue, run.active = catalogue, activeAreas(catalogue)
	if len(run.active) < 2 || len(run.active) > 255 {
		return "catalogue size", nil
	}
	if run.opts.Base == "" || run.opts.Head == "" {
		return "push range unavailable", nil
	}
	if err := run.opts.Config.Validate(); err != nil {
		return "", err
	}
	return "", run.loadOwnership(ctx, root)
}

func (run *laneRun) loadOwnership(ctx context.Context, root string) error {
	features, err := clawpatch.LoadFeatures(clawpatch.StateDir(root))
	if err != nil {
		return err
	}
	if run.own, err = NewStoreOwnership(ctx, run.store, features); err != nil {
		return err
	}
	if err := MarkStale(ctx, run.store, run.opts.Fingerprint); err != nil {
		return err
	}
	generation, err := run.store.LoadMappingRefresh(ctx)
	if err != nil {
		return err
	}
	run.generation = generation.LastAgentRefreshAt
	return nil
}

func (run *laneRun) baseRecord(m Merge, key string, files []string) domain.Association {
	return domain.Association{ID: "association:" + laneDigest(run.opts.RunID+"\x00"+m.SHA+"\x00"+key), RunID: run.opts.RunID, DecidedAt: time.Now().UTC(), LastAgentRefreshAt: run.generation, MergeCommit: m.SHA, ClusterKey: key, Files: files, PRMetadata: domain.PRMetadataAbsent}
}

func (run *laneRun) decide(ctx context.Context, m Merge, clusters []Cluster) ([]Decision, error) {
	out := make([]Decision, len(clusters))
	for i, c := range clusters {
		a, decision, err := run.decideCluster(ctx, m, c)
		if err != nil {
			return nil, err
		}
		out[i] = decision
		run.rows = append(run.rows, a)
	}
	return out, nil
}

func (run *laneRun) decideCluster(ctx context.Context, m Merge, c Cluster) (domain.Association, Decision, error) {
	input, catDigest, err := boundedState(ctx, run.repo, m, c, run.active)
	if err != nil {
		return domain.Association{}, Decision{}, err
	}
	a := run.baseRecord(m, c.ID, c.Files)
	a.ThresholdsProvisional = true
	a.DecisionDigest = run.opts.Fingerprint.Digest()
	a.InputsDigest, a.CatalogueDigest = laneDigest(string(input)), catDigest
	asker := &clusterAsker{budget: &run.budget, input: input, a: &a}
	run.askJev(ctx, asker, c)
	if asker.attempted {
		run.h.Units++
	}
	composed := Compose(c.ID, asker.scores, run.opts.Mapping.CoveredScore, run.opts.Mapping.TChoice)
	decision := applyComposed(&a, composed, asker.scores)
	if composed.Outcome == "none" || a.Reason == domain.AssociationReasonBelowTChoice || a.Reason == domain.AssociationReasonPermutationNo {
		run.fallbackCluster(ctx, &a, &decision, input, c)
	}
	if err := run.combineRawAnswer(&a, asker.audit); err != nil {
		return domain.Association{}, Decision{}, err
	}
	return a, decision, nil
}

type clusterAsker struct {
	budget    *jevBudget
	input     json.RawMessage
	a         *domain.Association
	scores    Scores
	audit     []jevAudit
	attempted bool
}

func (k *clusterAsker) ask(ctx context.Context, q decisions.Question, key string) (decisions.Response, bool) {
	reply, reason, called := k.budget.ask(ctx, q, key, k.input)
	k.attempted = k.attempted || called
	if called {
		k.audit = append(k.audit, jevAudit{Request: reply.RequestBody, Answers: reply.Answers, Model: reply.Model, Usage: reply.Usage})
	}
	k.a.Retries += max(0, reply.Attempts-1)
	k.a.LatencyMillis += reply.Latency.Milliseconds()
	if reply.Model != "" {
		k.a.ReportedModel = reply.Model
	}
	if reason != "" {
		k.scores.FailureReason = reason
		return decisions.Response{}, false
	}
	return reply.Answers[0], true
}

func (run *laneRun) askJev(ctx context.Context, k *clusterAsker, c Cluster) {
	covered, ok := k.ask(ctx, decisions.MergeCovered, c.ID+":covered")
	if !ok {
		return
	}
	p := covered.Probability
	k.scores.Covered, k.a.JevCovered = &p, &p
	// A negative covered answer is terminal; no choice call is needed.
	if 1-p < run.opts.Mapping.CoveredScore-1e-12 {
		askChoices(ctx, k, decisions.MergeWhichArea(run.active), c.ID)
	}
}

func askChoices(ctx context.Context, k *clusterAsker, q decisions.Question, id string) {
	choice, ok := k.ask(ctx, q, id+":choice:0")
	if !ok {
		return
	}
	k.scores.Choices = append(k.scores.Choices, ChoiceScore{AreaID: choice.Choice, Confidence: choice.Probability})
	if !FlipSelected(id) {
		return
	}
	if second, ok := k.ask(ctx, q, id+":choice:1"); ok {
		k.scores.Choices = append(k.scores.Choices, ChoiceScore{AreaID: second.Choice, Confidence: second.Probability})
	}
}

func applyComposed(a *domain.Association, composed Composed, scores Scores) Decision {
	var decision Decision
	a.JevOutcome = composed.Outcome
	if len(scores.Choices) > 0 {
		confidence := composed.Confidence
		a.JevChoice = &confidence
	}
	a.Outcome, a.Reason = domain.AssociationUnresolved, composed.Reason
	switch composed.Outcome {
	case "assigned":
		a.Outcome, a.Reason, a.AreaIDs = domain.AssociationAssigned, "", []string{composed.AreaID}
		decision.Owners = a.AreaIDs
	case "none":
		a.Reason = domain.AssociationReasonNoneUnsent
	}
	if composed.Outcome != "assigned" {
		a.EscalationReason = a.Reason
	}
	return decision
}

func (run *laneRun) fallbackCluster(ctx context.Context, a *domain.Association, decision *Decision, input json.RawMessage, c Cluster) {
	switch {
	case run.fallback.Reason != "":
		a.FallbackSkipped, a.FallbackSkipReason = true, run.fallback.Reason
	case run.fallback.Units >= run.opts.Mapping.Fallback.MaxPerRun:
		a.Reason = domain.AssociationReasonFallbackCap
	default:
		run.fallback.Units++
		answer, terraErr := run.terra.Decide(ctx, input, run.active)
		terra := run.terra
		a.CLIName, a.CLIVersion = "codex", terra.version
		a.PromptDigest = TerraPromptDigest
		a.ReportedModel, a.ReportedEffort = terra.model, terra.effort
		a.Retries += max(0, terra.attempts-1)
		a.LatencyMillis += terra.latency.Milliseconds()
		if terraErr != nil {
			a.Reason = domain.AssociationReasonTerraFailure
		} else {
			run.fallbackSucceeded++
			run.placeTerraAnswer(a, decision, answer, c)
		}
		a.RawAnswer = decisions.Redact(strings.ReplaceAll(terra.raw, run.opts.Client.Key, "[REDACTED]"))
	}
}

func (run *laneRun) placeTerraAnswer(a *domain.Association, decision *Decision, answer TerraAnswer, c Cluster) {
	switch {
	case answer.Decision == "existing":
		a.Outcome, a.Reason, a.AreaIDs = domain.AssociationAssigned, "", []string{answer.AreaID}
		decision.Owners = a.AreaIDs
	case run.pendingCount >= run.opts.Mapping.MaxPendingPerRun:
		a.Reason = domain.AssociationReasonPendingCap
	default:
		run.pendingCount++
		id := "area:association:" + laneDigest(a.ID)
		a.Outcome, a.Reason, a.PendingAreaID = domain.AssociationPendingNew, "", id
		decision.Pending = id
		run.catalogue = append(run.catalogue, areas.Area{ScoutAreaID: id, Title: answer.NewTitle, Summary: answer.NewSummary, FileSet: append([]string(nil), c.Files...), Origin: "association", Status: areas.StatusPending, ClusterID: c.ID, DecisionID: a.ID, FirstSeenSnapshot: run.opts.Head})
	}
}

func (run *laneRun) combineRawAnswer(a *domain.Association, audit []jevAudit) error {
	encoded, err := json.Marshal(audit)
	if err != nil {
		return err
	}
	jevRaw := decisions.Redact(strings.ReplaceAll(string(encoded), run.opts.Client.Key, "[REDACTED]"))
	if a.RawAnswer == "" {
		a.RawAnswer = jevRaw
		return nil
	}
	combined, err := json.Marshal(struct {
		Jev   json.RawMessage `json:"jev"`
		Terra string          `json:"terra"`
	}{json.RawMessage(jevRaw), a.RawAnswer})
	if err != nil {
		return err
	}
	a.RawAnswer = string(combined)
	return nil
}

func (run *laneRun) recordMergeResults(results []MergeResult) {
	for _, mr := range results {
		for _, inherited := range mr.Inherited {
			a := run.baseRecord(mr.Merge, "inherited:"+inherited.Path, []string{inherited.Path})
			a.Outcome, a.InheritedRule, a.InheritedFrom, a.InheritedAreaIDs = domain.AssociationInherited, inherited.Rule, inherited.From, inherited.Areas
			run.rows = append(run.rows, a)
		}
		for _, c := range mr.Capped {
			a := run.baseRecord(mr.Merge, c.ID, c.Files)
			a.Outcome, a.Reason = domain.AssociationUnresolved, domain.AssociationReasonPRCap
			run.rows = append(run.rows, a)
		}
		for _, move := range mr.PendingMoves {
			movePendingPath(run.catalogue, move)
		}
	}
}

func movePendingPath(catalogue []areas.Area, move PendingMove) {
	for i := range catalogue {
		for _, id := range move.Pending {
			if catalogue[i].ScoutAreaID != id {
				continue
			}
			for j, p := range catalogue[i].FileSet {
				if p == move.Old {
					catalogue[i].FileSet[j] = move.New
				}
			}
		}
	}
}

func (run *laneRun) persist(ctx context.Context, results []MergeResult) error {
	if len(run.rows) > 0 {
		if err := run.store.InsertAssociations(ctx, run.rows); err != nil {
			return err
		}
	}
	pendingChanged := run.pendingCount > 0
	for _, mr := range results {
		pendingChanged = pendingChanged || len(mr.PendingMoves) > 0
	}
	if pendingChanged {
		return run.store.ReplaceAreas(ctx, run.catalogue)
	}
	return nil
}

func (run *laneRun) finishHealth() {
	h, fallback := run.h, run.fallback
	if run.budget.succeeded > 0 {
		h.Status = domain.LaneRan
	} else if h.Units > 0 {
		h.Status = domain.LaneFailed
	} else {
		h.Reason = "no residual clusters"
	}
	if run.fallbackSucceeded > 0 {
		fallback.Status = domain.LaneRan
	} else if fallback.Units > 0 {
		fallback.Status = domain.LaneFailed
	} else if fallback.Reason == "" {
		fallback.Reason = "no residual clusters"
	}
}

type jevAudit struct {
	Request json.RawMessage      `json:"request,omitempty"`
	Answers []decisions.Response `json:"answers,omitempty"`
	Model   string               `json:"model,omitempty"`
	Usage   json.RawMessage      `json:"usage,omitempty"`
}

type jevBudget struct {
	opts             LaneOptions
	calls, succeeded int
	cost             float64
	unknownCost      bool
}

func (b *jevBudget) ask(ctx context.Context, q decisions.Question, key string, input json.RawMessage) (decisions.Reply, string, bool) {
	if b.calls >= b.opts.Config.MaxPerRun || b.unknownCost || b.cost >= b.opts.Config.MaxCostUSD {
		return decisions.Reply{}, domain.AssociationReasonJevCap, false
	}
	b.calls++
	reply, err := b.opts.Client.Ask(ctx, []decisions.Question{q}, b.opts.Config.Model, key, input)
	if reply.CostKnown {
		b.cost += reply.CostUSD
	} else if reply.Model != "" || len(reply.Usage) > 0 {
		b.unknownCost = true
	}
	if err != nil {
		return reply, domain.AssociationReasonJevFailure, true
	}
	if q.Kind == decisions.KindChoice {
		p, ok := reply.Answers[0].Probabilities[reply.Answers[0].Choice]
		if !ok {
			return reply, domain.AssociationReasonJevFailure, true
		}
		reply.Answers[0].Probability = p
	}
	b.succeeded++
	return reply, "", true
}

func laneDigest(s string) string { sum := sha256.Sum256([]byte(s)); return hex.EncodeToString(sum[:]) }

type stateMember struct {
	Path       string `json:"path"`
	ChangeKind Kind   `json:"change_kind"`
	Role       string `json:"role"`
	Text       string `json:"text"`
}

type stateArea struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	Files   []string `json:"files"`
}

func boundedState(ctx context.Context, repo Repo, m Merge, cluster Cluster, catalogue []areas.Area) (json.RawMessage, string, error) {
	members := make([]stateMember, 0, len(cluster.Files))
	for _, path := range cluster.Files {
		content, err := repo.File(ctx, m.SHA, path)
		if err != nil {
			return nil, "", err
		}
		lines := strings.SplitN(string(content), "\n", 61)
		if len(lines) > 60 {
			lines = lines[:60]
		}
		kind := Added
		for _, c := range m.Changes {
			if c.Path == path {
				kind = c.Kind
				break
			}
		}
		members = append(members, stateMember{Path: path, ChangeKind: kind, Role: memberRole(path), Text: decisions.Redact(strings.Join(lines, "\n"))})
	}
	cat := make([]stateArea, 0, len(catalogue))
	for _, a := range catalogue {
		files := append([]string(nil), a.FileSet...)
		sort.Strings(files)
		if len(files) > 5 {
			files = files[:5]
		}
		cat = append(cat, stateArea{ID: a.ScoutAreaID, Title: decisions.Redact(a.Title), Summary: decisions.Redact(a.Summary), Files: files})
	}
	catRaw, err := json.Marshal(cat)
	if err != nil {
		return nil, "", err
	}
	args := append([]string{"diff", "--stat", m.Parent, m.SHA, "--"}, cluster.Files...)
	diff, err := repo.git(ctx, args...)
	if err != nil {
		return nil, "", err
	}
	payload := struct {
		Unit []stateMember `json:"unit"`
		PR   struct {
			Title        string   `json:"title"`
			Body         string   `json:"body"`
			ChangedPaths []string `json:"changed_paths"`
		} `json:"pr"`
		PRMetadata  string      `json:"pr_metadata"`
		DiffSummary string      `json:"diff_summary"`
		Catalogue   []stateArea `json:"catalogue"`
	}{Unit: members, PRMetadata: domain.PRMetadataAbsent, DiffSummary: decisions.Redact(string(diff)), Catalogue: cat}
	input, err := json.Marshal(payload)
	return input, laneDigest(string(catRaw)), err
}

func memberRole(path string) string {
	lower := "/" + strings.ToLower(path)
	for _, marker := range []struct{ path, role string }{{"/testdata/", "fixture"}, {"/test/", "test"}, {"/tests/", "test"}, {"_test.", "test"}, {"test_", "test"}, {"/vendor/", "vendor"}, {"/docs/", "docs"}} {
		if strings.Contains(lower, marker.path) {
			return marker.role
		}
	}
	return "source"
}
