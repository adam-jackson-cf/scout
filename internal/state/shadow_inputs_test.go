package state

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"scout/internal/domain"
)

func TestShadowSupplementalInputsImmutableAndAtomic(t *testing.T) {
	ctx := context.Background()
	s := adr012Open(t, t.TempDir())
	f := adr012Fact(domain.ShadowScopeCI, "999")
	adr012Record(t, s, f, nil, nil)
	obligation := domain.ProviderEventObligation{ID: "provider-event", RepositoryID: f.ID.RepositoryID, Scope: f.Scope, PRNumber: f.PRNumber, HeadCommit: f.HeadCommit, EventAt: f.StartedAt, Selector: "selector-1", ProviderProof: "enumeration-proof"}
	accounting := domain.PRRunObligationLink{AttemptID: f.ID, Scope: f.Scope, CheckID: f.CheckID, ObligationID: obligation.ID}
	original := domain.ShadowTriage{ActionID: "123", EventID: "finding", RepositoryID: f.ID.RepositoryID, Scope: f.Scope, Verdict: "warranted", Actor: "revoked-maintainer", PolicyRevision: "source-tip", TriagedAt: adr012Time}
	refusal := domain.ShadowImportRefusal{ID: "refusal-1", RepositoryID: f.ID.RepositoryID, Scope: f.Scope, SourceLineage: "source-genesis", Original: original, Reason: "author revoked", PolicyRevision: "destination-tip", RecordedAt: adr012Time}
	pending := domain.ShadowPendingImport{ID: "import-1", RepositoryID: f.ID.RepositoryID, Scope: f.Scope, SourceLineage: "source-genesis", Publication: domain.ShadowPublication{ID: "source-publication", RepositoryID: f.ID.RepositoryID, RecordedAt: adr012Time, Triage: []domain.ShadowTriage{original}}}
	input := domain.ShadowInputs{Obligations: []domain.ProviderEventObligation{obligation}, Accounting: []domain.PRRunObligationLink{accounting}, ImportRefusals: []domain.ShadowImportRefusal{refusal}, PendingImports: []domain.ShadowPendingImport{pending}}
	for range 2 {
		if err := s.RecordShadowInputs(ctx, input); err != nil {
			t.Fatal(err)
		}
	}
	before, err := s.ReadShadowInputs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, input) {
		t.Fatalf("round trip: %#v", before)
	}
	for _, kind := range []string{"obligation", "refusal", "pending"} {
		t.Run(kind, func(t *testing.T) {
			changed := domain.ShadowInputs{}
			switch kind {
			case "obligation":
				o := obligation
				o.HeadCommit = "different-head"
				changed.Obligations = []domain.ProviderEventObligation{o}
			case "refusal":
				r := refusal
				r.Reason = "rewritten"
				changed.ImportRefusals = []domain.ShadowImportRefusal{r}
			case "pending":
				p := pending
				p.SourceLineage = "rewritten"
				changed.PendingImports = []domain.ShadowPendingImport{p}
			}
			if err := s.RecordShadowInputs(ctx, changed); !errors.Is(err, ErrImmutable) {
				t.Fatalf("want immutable: %v", err)
			}
			after, err := s.ReadShadowInputs(ctx)
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatalf("mutation: %#v %v", after, err)
			}
		})
	}
	// A later invalid link rolls back the earlier new obligation in this batch.
	fresh := obligation
	fresh.ID = "new-event"
	dangling := accounting
	dangling.ObligationID = "missing"
	if err := s.RecordShadowInputs(ctx, domain.ShadowInputs{Obligations: []domain.ProviderEventObligation{fresh}, Accounting: []domain.PRRunObligationLink{dangling}}); err == nil {
		t.Fatal("dangling accounting accepted")
	}
	after, err := s.ReadShadowInputs(ctx)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatalf("partial insert: %#v %v", after, err)
	}
	if ledger := adr012Ledger(t, s); len(ledger.Frontiers) != 0 || len(ledger.Publications) != 0 || len(ledger.Triage) != 0 {
		t.Fatal("pending/refused import minted authority")
	}
}

func TestShadowObligationAccountingBindingAndScope(t *testing.T) {
	ctx := context.Background()
	s := adr012Open(t, t.TempDir())
	for _, scope := range []domain.ShadowScope{domain.ShadowScopeCI, domain.ShadowScopeLocal, domain.ShadowScopeLegacy} {
		f := adr012Fact(scope, "999")
		adr012Record(t, s, f, nil, nil)
		o := domain.ProviderEventObligation{ID: "provider-event", RepositoryID: f.ID.RepositoryID, Scope: scope, PRNumber: f.PRNumber, HeadCommit: f.HeadCommit, EventAt: f.StartedAt, Selector: "selector", ProviderProof: "proof"}
		if err := s.RecordShadowInputs(ctx, domain.ShadowInputs{Obligations: []domain.ProviderEventObligation{o}, Accounting: []domain.PRRunObligationLink{{AttemptID: f.ID, Scope: scope, CheckID: f.CheckID, ObligationID: o.ID}}}); err != nil {
			t.Fatal(err)
		}
	}
	before, err := s.ReadShadowInputs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(before.Obligations) != 3 || len(before.Accounting) != 3 {
		t.Fatal("scope collision")
	}
	for _, kind := range []string{"head", "PR", "future", "scope", "repository"} {
		t.Run(kind, func(t *testing.T) {
			f := adr012Fact(domain.ShadowScopeCI, "999")
			o := domain.ProviderEventObligation{ID: "bad-" + kind, RepositoryID: f.ID.RepositoryID, Scope: f.Scope, PRNumber: f.PRNumber, HeadCommit: f.HeadCommit, EventAt: f.StartedAt, Selector: "selector", ProviderProof: "proof"}
			switch kind {
			case "head":
				o.HeadCommit = "H2"
			case "PR":
				o.PRNumber = "18"
			case "future":
				o.EventAt = o.EventAt.Add(time.Millisecond)
			case "scope":
				o.Scope = domain.ShadowScopeLocal
			case "repository":
				o.RepositoryID = "foreign"
			}
			if err := s.RecordShadowInputs(ctx, domain.ShadowInputs{Obligations: []domain.ProviderEventObligation{o}, Accounting: []domain.PRRunObligationLink{{AttemptID: f.ID, Scope: f.Scope, CheckID: f.CheckID, ObligationID: o.ID}}}); err == nil {
				t.Fatal("invalid binding accepted")
			}
			after, err := s.ReadShadowInputs(ctx)
			if err != nil || !reflect.DeepEqual(before, after) {
				t.Fatal("partial invalid accounting")
			}
		})
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	ro, err := OpenExistingReadOnly(s.Root())
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	if err := ro.RecordShadowInputs(ctx, domain.ShadowInputs{}); err == nil {
		t.Fatal("read-only input write accepted")
	}
	after, err := ro.ReadShadowInputs(ctx)
	if err != nil || !reflect.DeepEqual(before, after) {
		t.Fatal("read-only round trip")
	}
}

func TestShadowPublicationParentAndNumericPrecision(t *testing.T) {
	ctx := context.Background()
	s := adr012Open(t, t.TempDir())
	// Arbitrarily large numeric identifiers are preserved, not cast to uint64.
	f := adr012Fact(domain.ShadowScopeCI, "184467440737095516160000")
	adr012Record(t, s, f, nil, nil)
	for _, id := range []string{"1e3", "01", "1.5", "-1", ""} {
		bad := f
		bad.ID.RunID = id
		if err := s.RecordPRRun(ctx, bad, nil, nil); err == nil {
			t.Fatalf("noncanonical provider ID %q accepted", id)
		}
	}
	p := domain.ShadowPublication{ID: "p1", RepositoryID: f.ID.RepositoryID, RecordedAt: adr012Time}
	if err := s.ApplyShadowPublication(ctx, p); err != nil {
		t.Fatal(err)
	}
	before := adr012Ledger(t, s)
	for _, kind := range []string{"fork", "missing-parent", "non-increasing"} {
		bad := p
		bad.ID = kind
		bad.RecordedAt = bad.RecordedAt.Add(time.Millisecond)
		switch kind {
		case "missing-parent":
			bad.ParentID = "missing"
		case "non-increasing":
			bad.ParentID = p.ID
			bad.RecordedAt = p.RecordedAt
		}
		if err := s.ApplyShadowPublication(ctx, bad); err == nil {
			t.Fatalf("invalid publication %s accepted", kind)
		}
		adr012Unchanged(t, s, before)
	}
}

func TestShadowAttemptRetainsIndependentSelectedChecks(t *testing.T) {
	ctx := context.Background()
	s := adr012Open(t, t.TempDir())
	first := adr012Fact(domain.ShadowScopeCI, "999")
	second := first
	second.CheckID = "security:other"
	second.RuleVersion = "2"
	for _, f := range []domain.PRRunFact{first, second} {
		e := adr012Event(f, "event-"+f.CheckID)
		adr012Record(t, s, f, []domain.ProviderPREvent{e}, []domain.PRRunEventLink{adr012Link(f, e)})
	}
	before := adr012Ledger(t, s)
	if len(before.Facts) != 2 || len(before.Links) != 2 {
		t.Fatal("selected checks collided within one provider job")
	}
	changed := second
	changed.RuleVersion = "3"
	if err := s.RecordPRRun(ctx, changed, nil, nil); !errors.Is(err, ErrImmutable) {
		t.Fatalf("rule version changed within an immutable selected check: %v", err)
	}
	adr012Unchanged(t, s, before)
}
