package baseline

import "testing"

func TestMergeRejectsUnparseableDispositionTimestamp(t *testing.T) {
	tip := ledger(t, "2026-10-01T00:00:00Z", "issue:a")
	for name, bad := range map[string]string{"garbage": "not-a-time", "empty": ""} {
		t.Run(name, func(t *testing.T) {
			local := ledger(t, bad, "issue:a")
			if _, err := MergeState(tip, local); err == nil {
				t.Fatal("a disposition without a parseable updatedAt must fail the merge, not count as infinitely old")
			}
			if _, err := MergeState(local, tip); err == nil {
				t.Fatal("an unparseable tip timestamp must fail the merge")
			}
		})
	}
}
