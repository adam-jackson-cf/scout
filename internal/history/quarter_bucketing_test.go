//go:build boundary

package history

import (
	"context"
	"testing"
	"time"
)

func TestQuarterBucketingAcrossLeapDayAndImmaturity(t *testing.T) {
	f := newFixture(t)
	times := []time.Time{
		utc(2024, 2, 29, 23, 59, 59), // leap day: Q1
		utc(2024, 3, 31, 23, 59, 59), // last second of Q1
		utc(2024, 4, 1, 0, 0, 0),     // first second of Q2
		utc(2024, 5, 10, 12, 0, 0),
	}
	for i, ts := range times {
		f.commit(ts, "dev", "dev@example.com", map[string]string{"a/f.go": lines(i + 1)})
	}
	commits, _ := Load(context.Background(), f.dir)
	sigs := Compute(commits, Options{Scopes: scopesAB()[:1], End: utc(2024, 5, 15, 0, 0, 0), Snapshot: "s"})
	q1 := find(t, sigs, "scout.changes@1", "a", "2024-Q1")
	q2 := find(t, sigs, "scout.changes@1", "a", "2024-Q2")
	near(t, "Q1", q1.Value, 2)
	near(t, "Q2", q2.Value, 2)
	if q1.Detail["immature"] != false || q2.Detail["immature"] != true {
		t.Errorf("immature flags: Q1 %v, Q2 %v", q1.Detail["immature"], q2.Detail["immature"])
	}
	if q1.Detail["window_start"] != "2024-01-01T00:00:00Z" || q1.Detail["window_end"] != "2024-04-01T00:00:00Z" {
		t.Errorf("Q1 bounds %v..%v", q1.Detail["window_start"], q1.Detail["window_end"])
	}
	for _, s := range sigs {
		if err := s.Validate(); err != nil {
			t.Fatal(err)
		}
	}
}
