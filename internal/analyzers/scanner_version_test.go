//go:build boundary

package analyzers

import (
	"context"
	"os"
	"testing"
	"time"

	"scout/internal/sandbox"
)

type versionBudgetCase struct {
	name    string
	timeout time.Duration
	caller  time.Duration
	command string
	wantErr bool
}

func TestScannerVersionHonorsDeclaredBudget(t *testing.T) {
	if err := sandbox.Available(); err != nil {
		t.Fatal(err)
	}
	for _, test := range []versionBudgetCase{
		{"slow startup within budget", 30 * time.Second, 0, "sleep 11; printf '0.15.8\\n'", false},
		{"scanner budget exhausted", 20 * time.Millisecond, 0, "sleep 1; printf '0.15.8\\n'", true},
		{"caller budget exhausted", 30 * time.Second, 20 * time.Millisecond, "sleep 1; printf '0.15.8\\n'", true},
		{"wrong version", 30 * time.Second, 0, "printf '0.0.0\\n'", true},
	} {
		t.Run(test.name, func(t *testing.T) { checkVersionBudget(t, test) })
	}
}

func checkVersionBudget(t *testing.T, test versionBudgetCase) {
	spec := qualitySpecs()[0]
	spec.Timeout = test.timeout
	spec.Binary.VersionArgs = []string{"-c", test.command}
	runner, err := NewRunner(spec, "run-version-budget")
	if err != nil {
		t.Fatal(err)
	}
	root := t.TempDir()
	state, err := sandbox.NewState("scout-version-test-")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(state); err != nil {
			t.Errorf("remove scanner state: %v", err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	if test.caller != 0 {
		var callerCancel context.CancelFunc
		ctx, callerCancel = context.WithTimeout(ctx, test.caller)
		defer callerCancel()
	}
	err = runner.verifyVersion(ctx, root, state, "", "/bin/sh")
	if (err != nil) != test.wantErr {
		t.Fatalf("version probe = %v, want error %v (scanner budget %v)", err, test.wantErr, test.timeout)
	}
	if test.wantErr && runner.verifiedBinary != "" {
		t.Fatal("failed probe cached a verified scanner")
	}
	if !test.wantErr && runner.verifiedBinary != "/bin/sh" {
		t.Fatal("successful probe did not cache the verified scanner")
	}
}
