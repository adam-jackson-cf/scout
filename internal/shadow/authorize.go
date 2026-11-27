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
	if !trustedWriter(w) || !p.Authenticated || p.Revision == "" || !sameRepository(a, w, p) || !originalActorExecuted(a, w) || !authorizedActor(a.Actor, p) || !completeAction(a) {
		return domain.AuthorizedTriage{}, fmt.Errorf("triage refused: authenticated original actor and current protected policy required")
	}
	return domain.AuthorizedTriage{Row: domain.ShadowTriage{ActionID: a.RepositoryID + ":" + a.DispatchRunID, EventID: a.EventID, RepositoryID: a.RepositoryID, Scope: domain.ShadowScopeCI, Verdict: a.Verdict, Actor: a.Actor, Reason: a.Reason, PolicyRevision: p.Revision, TriagedAt: a.OccurredAt}}, nil
}

func trustedWriter(w WriterIdentity) bool {
	return w.Authenticated && w.DefaultBranch && w.SourceReachable && w.SourceRevision != ""
}

func sameRepository(a TriageAction, w WriterIdentity, p ProtectedPolicy) bool {
	return a.RepositoryID != "" && a.RepositoryID == w.RepositoryID && a.RepositoryID == p.RepositoryID
}

func originalActorExecuted(a TriageAction, w WriterIdentity) bool {
	return a.Actor == w.OriginalActor && w.OriginalActor == w.ExecutingActor
}

func completeAction(a TriageAction) bool {
	return a.DispatchRunID != "" && a.EventID != "" && domain.ValidTriage(a.Verdict) && !a.OccurredAt.IsZero()
}
