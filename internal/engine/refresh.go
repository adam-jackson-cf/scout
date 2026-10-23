package engine

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path"
	"sort"
	"strings"
	"time"

	"scout/internal/clawpatch"
	"scout/internal/domain"
	"scout/internal/inventory"
	"scout/internal/policy"
	"scout/internal/state"
)

// RefreshEvent is the kind of run asking whether to refresh the agent map.
type RefreshEvent string

const (
	RefreshPR      RefreshEvent = "pr"
	RefreshPush    RefreshEvent = "push"
	RefreshNightly RefreshEvent = "nightly"
)

// RefreshStatus is the outcome of one refresh evaluation.
type RefreshStatus string

const (
	RefreshNone       RefreshStatus = "none"       // nothing triggered, nothing pending
	RefreshPending    RefreshStatus = "pending"    // trigger recorded or debounce still open
	RefreshDone       RefreshStatus = "refreshed"  // one agent refresh ran and reconciled
	RefreshSkipped    RefreshStatus = "skipped"    // no model CLI; pending kept, run not failed
	RefreshIncomplete RefreshStatus = "incomplete" // agent failed or found nothing; pending kept
	RefreshBlocked    RefreshStatus = "blocked"    // agent mapping not authorized; pending kept
)

// ChangeStatus is a commit-diff entry kind (git --name-status, no rename detection).
type ChangeStatus string

const (
	ChangeAdded    ChangeStatus = "A"
	ChangeDeleted  ChangeStatus = "D"
	ChangeModified ChangeStatus = "M"
)

// FileChange is one path touched by the commit diff.
type FileChange struct {
	Path   string       `json:"path"`
	Status ChangeStatus `json:"status"`
}

// RefreshStore is the durable refresh state. *state.Store satisfies it.
type RefreshStore interface {
	LoadMappingRefresh(ctx context.Context) (state.MappingRefresh, error)
	SaveMappingRefresh(ctx context.Context, record state.MappingRefresh) error
}

// AgentMapper runs the Clawpatch mapper. *clawpatch.Client satisfies it.
type AgentMapper interface {
	Map(ctx context.Context, inputsChanged bool, source clawpatch.MapSource) (clawpatch.MapResult, error)
}

// Refresher decides and performs agent-map refreshes. The clock is injected.
type Refresher struct {
	Store     RefreshStore
	Mapper    AgentMapper
	Reconcile func(ctx context.Context, mapped clawpatch.MapResult) error
	Now       func() time.Time
	Mapping   policy.Mapping
}

// RefreshRun describes the run being evaluated.
type RefreshRun struct {
	Event         RefreshEvent
	DefaultBranch bool // push runs: true only for pushes to the default branch
	Revision      string
	Changes       []FileChange
	// Features are the active snapshot's features: ownership and entrypoints.
	Features []domain.Feature
	// Force is `scout map --refresh`: refresh now, ignoring triggers and debounce.
	Force bool
}

// RefreshOutcome reports what happened.
type RefreshOutcome struct {
	Status                RefreshStatus         `json:"status"`
	Reason                string                `json:"reason,omitempty"`
	Triggers              []string              `json:"triggers,omitempty"`
	Pending               *state.PendingRefresh `json:"pending,omitempty"`
	ExplorationIncomplete bool                  `json:"explorationIncomplete,omitempty"`
	Map                   *clawpatch.MapResult  `json:"map,omitempty"`
}

// Run evaluates triggers, updates the pending record, and runs at most one
// agent refresh. PR runs and non-default pushes never refresh and never record.
func (r *Refresher) Run(ctx context.Context, run RefreshRun) (RefreshOutcome, error) {
	now := r.Now().UTC()
	record, err := r.Store.LoadMappingRefresh(ctx)
	if err != nil {
		return RefreshOutcome{}, err
	}
	if !run.Force {
		switch {
		case run.Event == RefreshPR:
			return r.report(RefreshNone, "pull-request runs reuse the active snapshot", record, nil), nil
		case run.Event == RefreshPush && !run.DefaultBranch:
			return r.report(RefreshNone, "only default-branch pushes refresh the map", record, nil), nil
		case run.Event != RefreshPush && run.Event != RefreshNightly:
			return RefreshOutcome{}, fmt.Errorf("unknown refresh event %q", run.Event)
		}
	}
	var triggers []string
	if run.Event == RefreshPush && run.DefaultBranch && !run.Force {
		triggers = RefreshTriggers(run.Changes, run.Features)
		if len(triggers) > 0 {
			record.Pending = mergePending(record.Pending, now, run.Revision, triggers)
			if err := r.Store.SaveMappingRefresh(ctx, record); err != nil {
				return RefreshOutcome{}, err
			}
		}
	}
	onboarding := record.LastAgentRefreshAt == nil
	due := record.Pending != nil && now.Sub(record.Pending.LastTriggerAt) >= r.Mapping.RefreshDebounce
	switch {
	case run.Force, onboarding, due:
	case record.Pending != nil:
		return r.report(RefreshPending, "debounce window open", record, triggers), nil
	default:
		return r.report(RefreshNone, "no monitored change", record, triggers), nil
	}
	return r.refresh(ctx, record, now, triggers)
}

func (r *Refresher) refresh(ctx context.Context, record state.MappingRefresh, now time.Time, triggers []string) (RefreshOutcome, error) {
	mapped, err := r.Mapper.Map(ctx, true, clawpatch.MapSourceAgent)
	out := r.report(RefreshIncomplete, mapped.Reason, record, triggers)
	out.Map = &mapped
	out.ExplorationIncomplete = mapped.ExplorationIncomplete
	switch {
	case mapped.Status == clawpatch.MapSkipped:
		out.Status = RefreshSkipped
		return out, nil
	case err != nil || mapped.Status == clawpatch.MapBlocked:
		out.Status = RefreshBlocked
		if err != nil {
			out.Reason = err.Error()
		}
		return out, nil
	case mapped.ExplorationIncomplete || mapped.Source != clawpatch.MapSourceAgent || len(mapped.Features) == 0:
		out.ExplorationIncomplete = true
		return out, nil
	}
	if r.Reconcile != nil {
		if err := r.Reconcile(ctx, mapped); err != nil {
			return out, fmt.Errorf("reconcile areas: %w", err)
		}
	}
	record.Pending = nil
	record.LastAgentRefreshAt = &now
	if err := r.Store.SaveMappingRefresh(ctx, record); err != nil {
		return out, err
	}
	out.Status, out.Reason, out.Pending = RefreshDone, "", nil
	return out, nil
}

func (r *Refresher) report(status RefreshStatus, reason string, record state.MappingRefresh, triggers []string) RefreshOutcome {
	return RefreshOutcome{Status: status, Reason: reason, Triggers: triggers, Pending: record.Pending}
}

func mergePending(prev *state.PendingRefresh, now time.Time, revision string, triggers []string) *state.PendingRefresh {
	next := state.PendingRefresh{FirstTriggerAt: now, LastTriggerAt: now, Reasons: []string{}, Revisions: []string{}}
	if prev != nil {
		next = *prev
		next.LastTriggerAt = now
		next.Reasons = append([]string{}, prev.Reasons...)
		next.Revisions = append([]string{}, prev.Revisions...)
	}
	seen := map[string]bool{}
	for _, reason := range next.Reasons {
		seen[reason] = true
	}
	for _, reason := range triggers {
		if !seen[reason] {
			next.Reasons = append(next.Reasons, reason)
			seen[reason] = true
		}
	}
	sort.Strings(next.Reasons)
	if revision != "" {
		next.Revisions = append(next.Revisions, revision)
	}
	return &next
}

var manifestNames = map[string]bool{
	"package.json": true, "go.mod": true, "Cargo.toml": true, "pyproject.toml": true,
	"setup.py": true, "requirements.txt": true, "pom.xml": true, "build.gradle": true,
	"build.gradle.kts": true, "Gemfile": true, "composer.json": true, "Package.swift": true,
	"mix.exs": true, "deno.json": true, "Dockerfile": true,
}

var entrypointNames = map[string]bool{
	"main.go": true, "main.py": true, "__main__.py": true, "manage.py": true, "app.py": true,
	"main.rs": true, "lib.rs": true, "main.ts": true, "main.js": true, "index.ts": true,
	"index.js": true, "server.ts": true, "server.js": true,
}

// RefreshTriggers is the trigger predicate over a commit diff: new unowned
// files, new or removed manifests or entrypoints, and structural changes
// (removal) of owned paths. Content-only changes never trigger.
func RefreshTriggers(changes []FileChange, features []domain.Feature) []string {
	owned := map[string]bool{}
	entry := map[string]bool{}
	for _, f := range features {
		for _, ref := range f.OwnedFiles {
			owned[ref.Path] = true
		}
		for _, e := range f.Entrypoints {
			entry[e.Path] = true
		}
	}
	set := map[string]bool{}
	for _, c := range changes {
		if (c.Status != ChangeAdded && c.Status != ChangeDeleted) || inventory.IsScoutState(c.Path) {
			continue
		}
		base := path.Base(c.Path)
		kind := ""
		switch {
		case manifestNames[base]:
			kind = "manifest"
		case entrypointNames[base] || entry[c.Path]:
			kind = "entrypoint"
		}
		verb := map[ChangeStatus]string{ChangeAdded: "added", ChangeDeleted: "removed"}[c.Status]
		switch {
		case kind != "":
			set[fmt.Sprintf("%s-%s:%s", kind, verb, c.Path)] = true
		case c.Status == ChangeDeleted && owned[c.Path]:
			set["owned-removed:"+c.Path] = true
		case c.Status == ChangeAdded && !owned[c.Path]:
			set["unowned-added:"+c.Path] = true
		}
	}
	out := make([]string, 0, len(set))
	for reason := range set {
		out = append(out, reason)
	}
	sort.Strings(out)
	return out
}

// GitChanges lists added/deleted/modified paths between two revisions
// (renames are reported as delete + add).
func GitChanges(ctx context.Context, root, base, head string) ([]FileChange, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", "--name-status", "--no-renames", "-z", base, head)
	cmd.Dir = root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	raw, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git diff %s %s: %w: %s", base, head, err, strings.TrimSpace(stderr.String()))
	}
	parts := strings.Split(string(raw), "\x00")
	var out []FileChange
	for i := 0; i+1 < len(parts); i += 2 {
		status := ChangeStatus(parts[i][:1])
		if status != ChangeAdded && status != ChangeDeleted {
			status = ChangeModified
		}
		out = append(out, FileChange{Path: parts[i+1], Status: status})
	}
	return out, nil
}
