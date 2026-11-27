package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"testing"

	"scout/internal/domain"
	"scout/internal/state"
)

func closeOrFail(t *testing.T, closer io.Closer) {
	t.Helper()
	if err := closer.Close(); err != nil {
		t.Errorf("close: %v", err)
	}
}

type capturedOutput struct {
	data []byte
	err  error
}

func captureStdout(t *testing.T, fn func() int) (string, int) {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, reader)
	captured := make(chan capturedOutput, 1)
	go func() {
		data, readErr := io.ReadAll(reader)
		captured <- capturedOutput{data, readErr}
	}()
	os.Stdout = writer
	defer func() { os.Stdout = original }()
	code := fn()
	os.Stdout = original
	closeOrFail(t, writer)
	out := <-captured
	if out.err != nil {
		t.Fatal(out.err)
	}
	return string(out.data), code
}

func recordScopeSignals(t *testing.T, ctx context.Context, root string) {
	t.Helper()
	store, err := state.Open(root)
	if err != nil {
		t.Fatal(err)
	}
	defer closeOrFail(t, store)
	base := domain.Signal{Version: 1, ScopeKind: domain.SignalScopeDirectory, ScopeAnchor: "pkg/sub", Snapshot: "old", PopulationDigest: "popdigest", ToolDigest: "t", ConfigDigest: "c", Claim: domain.ClaimDescriptive}
	old := base
	old.MetricID, old.Numerator, old.Denominator, old.Value = "scout.cc", domain.Float(0), domain.Float(4), domain.Float(0)
	latest := old
	latest.Snapshot, latest.Numerator, latest.Value = "new", domain.Float(1), domain.Float(1)
	crap := base
	crap.Snapshot, crap.MetricID, crap.Missing, crap.MissingReason = "new", "scout.crap", domain.SignalUnmeasured, "no coverage artifact"
	other := latest
	other.ScopeAnchor = "elsewhere"
	if err := store.RecordSignals(ctx, []domain.Signal{old, latest, crap, other}); err != nil {
		t.Fatal(err)
	}
}

type signalsReply struct {
	Result struct {
		Signals []domain.Signal `json:"signals"`
	} `json:"result"`
}

func assertLatestScopeSignals(t *testing.T, reply signalsReply) {
	t.Helper()
	byMetric := map[string]domain.Signal{}
	for _, signal := range reply.Result.Signals {
		byMetric[signal.MetricID] = signal
	}
	cc := byMetric["scout.cc"]
	if len(reply.Result.Signals) != 2 || cc.Snapshot != "new" || *cc.Numerator != 1 || *cc.Denominator != 4 || cc.PopulationDigest != "popdigest" {
		t.Fatalf("signals for the scope = %+v", reply.Result.Signals)
	}
	if crap := byMetric["scout.crap"]; crap.Missing != domain.SignalUnmeasured || crap.Value != nil {
		t.Fatalf("crap = %+v", crap)
	}
}

func TestSignalsCommandShowsNumeratorDenominatorAndPopulationForAScope(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	recordScopeSignals(t, ctx, root)

	out, code := captureStdout(t, func() int {
		return run(ctx, []string{"signals", "--root", root, "--scope", "pkg/sub", "--json"})
	})
	if code != exitComplete {
		t.Fatalf("exit %d: %s", code, out)
	}
	var reply signalsReply
	if err := json.Unmarshal([]byte(out), &reply); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	assertLatestScopeSignals(t, reply)

	out, _ = captureStdout(t, func() int {
		return run(ctx, []string{"signals", "--root", root, "--scope", "pkg/sub", "--metric", "scout.cc", "--history", "--json"})
	})
	if err := json.Unmarshal([]byte(out), &reply); err != nil || len(reply.Result.Signals) != 2 {
		t.Fatalf("history view = %s (%v)", out, err)
	}
}
