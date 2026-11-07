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
	h, fallback := &healths[0], &healths[1]
	switch {
	case trigger != domain.TriggerPush:
		h.Reason = "trigger"
	case opts.Eligibility.Fork:
		h.Reason = "fork"
	case !opts.Mapping.Enabled:
		h.Reason = "disabled"
	case !opts.Eligibility.HasKey || opts.Client == nil || opts.Client.Key == "":
		h.Reason = "no jev key"
	case !opts.Config.EgressAcknowledged:
		h.Reason = "no acknowledgement"
	case !opts.Config.Enabled || !opts.Eligibility.Enabled || !opts.Eligibility.EgressAllowed:
		h.Reason = "egress not allowed"
	}
	if h.Reason != "" {
		fallback.Reason = h.Reason
		return healths, nil
	}
	catalogue, err := store.ListAreas(ctx)
	if err != nil {
		return healths, err
	}
	active := make([]areas.Area, 0, len(catalogue))
	for _, a := range catalogue {
		if a.Active() {
			active = append(active, a)
		}
	}
	if len(active) < 2 || len(active) > 255 {
		h.Reason = "catalogue size"
		fallback.Reason = h.Reason
		return healths, nil
	}
	if opts.Base == "" || opts.Head == "" {
		h.Reason = "push range unavailable"
		fallback.Reason = h.Reason
		return healths, nil
	}
	if err := opts.Config.Validate(); err != nil {
		return healths, err
	}
	features, err := clawpatch.LoadFeatures(clawpatch.StateDir(root))
	if err != nil {
		return healths, err
	}
	own, err := NewStoreOwnership(ctx, store, features)
	if err != nil {
		return healths, err
	}
	if err := MarkStale(ctx, store, opts.Fingerprint); err != nil {
		return healths, err
	}
	generation, err := store.LoadMappingRefresh(ctx)
	if err != nil {
		return healths, err
	}
	repo := Repo{Dir: root}
	budget := jevBudget{opts: opts}
	terra := NewTerraDecider(opts.Mapping.Fallback)
	switch {
	case !opts.FallbackConsent:
		fallback.Reason = "no consent"
	default:
		if _, _, err := modelcli.Detect("codex"); err != nil {
			fallback.Reason = "no model CLI"
		}
	}
	pendingCount, fallbackSucceeded := 0, 0
	var rows []domain.Association
	baseRecord := func(m Merge, key string, files []string) domain.Association {
		return domain.Association{ID: "association:" + laneDigest(opts.RunID+"\x00"+m.SHA+"\x00"+key), RunID: opts.RunID, DecidedAt: time.Now().UTC(), LastAgentRefreshAt: generation.LastAgentRefreshAt, MergeCommit: m.SHA, ClusterKey: key, Files: files, PRMetadata: domain.PRMetadataAbsent}
	}
	results, err := Run(ctx, repo, opts.Base, opts.Head, own, func(ctx context.Context, m Merge, clusters []Cluster) ([]Decision, error) {
		out := make([]Decision, len(clusters))
		for i, c := range clusters {
			input, catDigest, err := boundedState(ctx, repo, m, c, active)
			if err != nil {
				return nil, err
			}
			a := baseRecord(m, c.ID, c.Files)
			a.ThresholdsProvisional = true
			a.DecisionDigest = opts.Fingerprint.Digest()
			a.InputsDigest, a.CatalogueDigest = laneDigest(string(input)), catDigest
			scores := Scores{}
			var audit []jevAudit
			attempted := false
			ask := func(q decisions.Question, key string) (decisions.Response, bool) {
				reply, reason, called := budget.ask(ctx, q, key, input)
				attempted = attempted || called
				if called {
					audit = append(audit, jevAudit{Request: reply.RequestBody, Answers: reply.Answers, Model: reply.Model, Usage: reply.Usage})
				}
				a.Retries += max(0, reply.Attempts-1)
				a.LatencyMillis += reply.Latency.Milliseconds()
				if reply.Model != "" {
					a.ReportedModel = reply.Model
				}
				if reason != "" {
					scores.FailureReason = reason
					return decisions.Response{}, false
				}
				return reply.Answers[0], true
			}
			if covered, ok := ask(decisions.MergeCovered, c.ID+":covered"); ok {
				p := covered.Probability
				scores.Covered, a.JevCovered = &p, &p
				// A negative covered answer is terminal; no choice call is needed.
				if 1-p < opts.Mapping.CoveredScore-1e-12 {
					q := decisions.MergeWhichArea(active)
					if choice, ok := ask(q, c.ID+":choice:0"); ok {
						scores.Choices = append(scores.Choices, ChoiceScore{AreaID: choice.Choice, Confidence: choice.Probability})
						if FlipSelected(c.ID) {
							if second, ok := ask(q, c.ID+":choice:1"); ok {
								scores.Choices = append(scores.Choices, ChoiceScore{AreaID: second.Choice, Confidence: second.Probability})
							}
						}
					}
				}
			}
			if attempted {
				h.Units++
			}
			composed := Compose(c.ID, scores, opts.Mapping.CoveredScore, opts.Mapping.TChoice)
			a.JevOutcome = composed.Outcome
			if len(scores.Choices) > 0 {
				confidence := composed.Confidence
				a.JevChoice = &confidence
			}
			a.Outcome, a.Reason = domain.AssociationUnresolved, composed.Reason
			switch composed.Outcome {
			case "assigned":
				a.Outcome, a.Reason, a.AreaIDs = domain.AssociationAssigned, "", []string{composed.AreaID}
				out[i].Owners = a.AreaIDs
			case "none":
				a.Reason = domain.AssociationReasonNoneUnsent
			}
			if composed.Outcome != "assigned" {
				a.EscalationReason = a.Reason
			}
			if composed.Outcome == "none" || a.Reason == domain.AssociationReasonBelowTChoice || a.Reason == domain.AssociationReasonPermutationNo {
				switch {
				case fallback.Reason != "":
					a.FallbackSkipped, a.FallbackSkipReason = true, fallback.Reason
				case fallback.Units >= opts.Mapping.Fallback.MaxPerRun:
					a.Reason = domain.AssociationReasonFallbackCap
				default:
					fallback.Units++
					answer, terraErr := terra.Decide(ctx, input, active)
					a.CLIName, a.CLIVersion = "codex", terra.version
					a.PromptDigest = TerraPromptDigest
					a.ReportedModel, a.ReportedEffort = terra.model, terra.effort
					a.Retries += max(0, terra.attempts-1)
					a.LatencyMillis += terra.latency.Milliseconds()
					if terraErr != nil {
						a.Reason = domain.AssociationReasonTerraFailure
					} else {
						fallbackSucceeded++
						switch {
						case answer.Decision == "existing":
							a.Outcome, a.Reason, a.AreaIDs = domain.AssociationAssigned, "", []string{answer.AreaID}
							out[i].Owners = a.AreaIDs
						case pendingCount >= opts.Mapping.MaxPendingPerRun:
							a.Reason = domain.AssociationReasonPendingCap
						default:
							pendingCount++
							id := "area:association:" + laneDigest(a.ID)
							a.Outcome, a.Reason, a.PendingAreaID = domain.AssociationPendingNew, "", id
							out[i].Pending = id
							catalogue = append(catalogue, areas.Area{ScoutAreaID: id, Title: answer.NewTitle, Summary: answer.NewSummary, FileSet: append([]string(nil), c.Files...), Origin: "association", Status: areas.StatusPending, ClusterID: c.ID, DecisionID: a.ID, FirstSeenSnapshot: opts.Head})
						}
					}
					a.RawAnswer = decisions.Redact(strings.ReplaceAll(terra.raw, opts.Client.Key, "[REDACTED]"))
				}
			}
			encoded, err := json.Marshal(audit)
			if err != nil {
				return nil, err
			}
			jevRaw := decisions.Redact(strings.ReplaceAll(string(encoded), opts.Client.Key, "[REDACTED]"))
			if a.RawAnswer == "" {
				a.RawAnswer = jevRaw
			} else {
				combined, err := json.Marshal(struct {
					Jev   json.RawMessage `json:"jev"`
					Terra string          `json:"terra"`
				}{json.RawMessage(jevRaw), a.RawAnswer})
				if err != nil {
					return nil, err
				}
				a.RawAnswer = string(combined)
			}
			rows = append(rows, a)
		}
		return out, nil
	})
	if err != nil {
		return healths, err
	}
	for _, mr := range results {
		for _, inherited := range mr.Inherited {
			a := baseRecord(mr.Merge, "inherited:"+inherited.Path, []string{inherited.Path})
			a.Outcome, a.InheritedRule, a.InheritedFrom, a.InheritedAreaIDs = domain.AssociationInherited, inherited.Rule, inherited.From, inherited.Areas
			rows = append(rows, a)
		}
		for _, c := range mr.Capped {
			a := baseRecord(mr.Merge, c.ID, c.Files)
			a.Outcome, a.Reason = domain.AssociationUnresolved, domain.AssociationReasonPRCap
			rows = append(rows, a)
		}
		for _, move := range mr.PendingMoves {
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
	}
	if len(rows) > 0 {
		if err := store.InsertAssociations(ctx, rows); err != nil {
			return healths, err
		}
	}
	pendingChanged := pendingCount > 0
	for _, mr := range results {
		pendingChanged = pendingChanged || len(mr.PendingMoves) > 0
	}
	if pendingChanged {
		if err := store.ReplaceAreas(ctx, catalogue); err != nil {
			return healths, err
		}
	}
	if budget.succeeded > 0 {
		h.Status = domain.LaneRan
	} else if h.Units > 0 {
		h.Status = domain.LaneFailed
	} else {
		h.Reason = "no residual clusters"
	}
	if fallbackSucceeded > 0 {
		fallback.Status = domain.LaneRan
	} else if fallback.Units > 0 {
		fallback.Status = domain.LaneFailed
	} else if fallback.Reason == "" {
		fallback.Reason = "no residual clusters"
	}
	return healths, nil
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
