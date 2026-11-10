package shadow

import (
	"encoding/json"
	"fmt"
	"scout/internal/domain"
	"scout/internal/policy"
)

func authorizedActor(actor string, p ProtectedPolicy) bool {
	raw, err := json.Marshal(map[string]any{"maintainers": p.Maintainers})
	if err != nil {
		return false
	}
	parsed, err := policy.Parse(raw)
	if err != nil {
		return false
	}
	ok, _ := parsed.AcceptOverride(actor, []byte(p.CODEOWNERS))
	return ok
}

// AuthorizeTriage binds original inputs to matching authenticated execution and
// the policy read at the protected tip. Dispatch time is provenance only.
func AuthorizeTriage(a TriageAction, w WriterIdentity, p ProtectedPolicy) (domain.AuthorizedTriage, error) {
	if !w.Authenticated || !p.Authenticated || !w.DefaultBranch || !w.SourceReachable || w.SourceRevision == "" || p.Revision == "" || a.RepositoryID == "" || a.RepositoryID != w.RepositoryID || a.RepositoryID != p.RepositoryID || a.Actor != w.OriginalActor || w.OriginalActor != w.ExecutingActor || !authorizedActor(a.Actor, p) || a.DispatchRunID == "" || a.EventID == "" || !domain.ValidTriage(a.Verdict) || a.OccurredAt.IsZero() {
		return domain.AuthorizedTriage{}, fmt.Errorf("triage refused: authenticated original actor and current protected policy required")
	}
	return domain.AuthorizedTriage{Row: domain.ShadowTriage{ActionID: a.RepositoryID + ":" + a.DispatchRunID, EventID: a.EventID, RepositoryID: a.RepositoryID, Scope: domain.ShadowScopeCI, Verdict: a.Verdict, Actor: a.Actor, Reason: a.Reason, PolicyRevision: p.Revision, TriagedAt: a.OccurredAt}}, nil
}
