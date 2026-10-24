package analyzers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"scout/internal/domain"
)

func TestExternalScannerBinary(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "assessed")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	root, err := canonicalRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	external := filepath.Join(base, "scanner")
	if err := os.WriteFile(external, nil, 0700); err != nil {
		t.Fatal(err)
	}
	if err := externalScannerBinary(root, external); err != nil {
		t.Fatalf("external scanner rejected: %v", err)
	}
	internal := filepath.Join(root, "scanner")
	if err := os.WriteFile(internal, nil, 0700); err != nil {
		t.Fatal(err)
	}
	if err := externalScannerBinary(root, internal); err == nil {
		t.Fatal("assessed repository scanner accepted")
	}
}

func incompleteResult(reason string) domain.AdapterResult {
	key := domain.CoverageKey{FilePath: "a.py", Family: domain.FamilySecurity, Category: "credential-leakage"}
	return domain.AdapterResult{
		ResultID: "result", Status: domain.AdapterFailed,
		Coverage: []domain.Coverage{{Key: key, Status: domain.CoverageIncomplete, Reason: reason}},
	}
}

// A batch that left an input incomplete for a proof reason is rescanned; the
// retry attempt is observable as the result being re-finalized.
func TestRetryUnprovenRescansEveryReasonTheRunnerWritesForUnprovenInputs(t *testing.T) {
	runner, err := NewRunner(specForTest(t, "security.gitleaks.default-credentials"), "run-retry")
	if err != nil {
		t.Fatal(err)
	}
	files := []domain.File{{Path: "a.py", Language: "python", Eligible: true}, {Path: "b.py", Language: "python", Eligible: true}}
	missingRoot := filepath.Join(t.TempDir(), "absent")
	batchErr := errors.New("batch incomplete")
	// completeCoverage is what writes reasonNotInspected into a real batch result.
	written := completeCoverage([]domain.Coverage{{Key: incompleteResult("").Coverage[0].Key}}, map[string]sourceInput{"a.py": {}}, parsedOutput{}, domain.AdapterFailed)
	for _, reason := range []string{written[0].Reason, reasonOutputInvalid, reasonOutputBound, reasonEvidenceInvalid} {
		result, _, err := runner.retryUnproven(context.Background(), missingRoot, files, incompleteResult(reason), nil, batchErr)
		if result.FinishedAt == nil || !errors.Is(err, batchErr) {
			t.Errorf("reason %q was not retried (finished=%v err=%v)", reason, result.FinishedAt, err)
		}
	}
}

func TestRetryUnprovenLeavesTimeoutsAndSetupFailuresAlone(t *testing.T) {
	runner, err := NewRunner(specForTest(t, "security.gitleaks.default-credentials"), "run-retry")
	if err != nil {
		t.Fatal(err)
	}
	files := []domain.File{{Path: "a.py", Language: "python", Eligible: true}, {Path: "b.py", Language: "python", Eligible: true}}
	for _, reason := range []string{"scanner timed out", "required scanner binary is not installed", "private input staging failed"} {
		result, _, err := runner.retryUnproven(context.Background(), t.TempDir(), files, incompleteResult(reason), nil, errors.New("batch incomplete"))
		if result.FinishedAt != nil || err == nil || result.Coverage[0].Reason != reason {
			t.Errorf("reason %q was retried or altered: %+v", reason, result)
		}
	}
}

func TestSecretDigestIsStablePerValueAndRejectsUnusableRanges(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.py"), []byte("x = 'secretvalue'\nsecond = 'secretvalue'\nplain = 1\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Gitleaks 8.30.1 columns: start = offset+2, end = last offset+2.
	first := parsedFinding{path: "a.py", startLine: 1, endLine: 1, startColumn: 7, endColumn: 17}
	second := parsedFinding{path: "a.py", startLine: 2, endLine: 2, startColumn: 12, endColumn: 22}
	one, err := secretDigest(root, first)
	sum := sha256.Sum256([]byte("scout-secret-digest-v1\x00secretvalue"))
	if err != nil || one != hex.EncodeToString(sum[:]) {
		t.Fatalf("digest = %q, %v; want the digest of the bare secret", one, err)
	}
	two, err := secretDigest(root, second)
	if err != nil || one != two {
		t.Fatalf("the same secret on another line changed identity: %q vs %q (%v)", one, two, err)
	}
	multiLine, err := secretDigest(root, parsedFinding{path: "a.py", startLine: 1, endLine: 2})
	if err != nil || multiLine != "" {
		t.Fatalf("multi-line finding = %q, %v", multiLine, err)
	}
	for name, finding := range map[string]parsedFinding{
		"column outside the line": {path: "a.py", startLine: 3, endLine: 3, startColumn: 4, endColumn: 90},
		"no column range":         {path: "a.py", startLine: 1, endLine: 1},
		"line outside the file":   {path: "a.py", startLine: 9, endLine: 9, startColumn: 2, endColumn: 4},
		"unreadable source":       {path: "missing.py", startLine: 1, endLine: 1, startColumn: 2, endColumn: 4},
	} {
		if digest, err := secretDigest(root, finding); err == nil {
			t.Errorf("%s: digest %q returned instead of an error", name, digest)
		}
	}
}
