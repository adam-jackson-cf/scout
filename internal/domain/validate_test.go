package domain

import "testing"

func TestValidateResolution(t *testing.T) {
	good := Verification{RunID: "run", Relevant: true, FailureModeAbsent: true, InspectionStatus: CoverageInspected}
	cases := []struct {
		name   string
		mutate func(*Verification)
		ok     bool
	}{
		{"inspected recheck", func(*Verification) {}, true},
		{"covered recheck", func(v *Verification) { v.InspectionStatus = CoverageCovered }, true},
		{"flagged recheck", func(v *Verification) { v.InspectionStatus = CoverageFlagged }, true},
		{"no run", func(v *Verification) { v.RunID = "" }, false},
		{"not relevant", func(v *Verification) { v.Relevant = false }, false},
		{"failure mode still present", func(v *Verification) { v.FailureModeAbsent = false }, false},
		{"reused coverage is not an inspection", func(v *Verification) { v.InspectionStatus = CoverageReused }, false},
		{"blocked coverage is not an inspection", func(v *Verification) { v.InspectionStatus = CoverageBlocked }, false},
		{"missing status", func(v *Verification) { v.InspectionStatus = "" }, false},
		{"deleted file needs no inspection", func(v *Verification) { v.Reason, v.InspectionStatus = ResolutionFileDeleted, "" }, true},
		{"deleted file still needs a run", func(v *Verification) { v.Reason, v.RunID = ResolutionFileDeleted, "" }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := good
			tc.mutate(&v)
			if err := v.ValidateResolution(); (err == nil) != tc.ok {
				t.Fatalf("ValidateResolution() = %v, want ok=%v", err, tc.ok)
			}
		})
	}
}

func TestSignalValidateRejectsMissingMeasurementsThatLookMeasured(t *testing.T) {
	base := Signal{MetricID: "m", Version: 1, ScopeKind: SignalScopeRepository, Snapshot: "s", Claim: ClaimDescriptive, ToolDigest: "t", ConfigDigest: "c"}
	measured := base
	measured.Value, measured.Numerator, measured.Denominator = Float(1), Float(1), Float(1)
	missing := base
	missing.Missing, missing.MissingReason = SignalUnmeasured, "no data"

	cases := []struct {
		name   string
		signal func() Signal
		ok     bool
	}{
		{"measured", func() Signal { return measured }, true},
		{"unmeasured with reason", func() Signal { return missing }, true},
		{"measured without numbers", func() Signal { return base }, false},
		{"unmeasured without reason", func() Signal { s := missing; s.MissingReason = ""; return s }, false},
		{"unmeasured carrying a value", func() Signal { s := missing; s.Value = Float(0); return s }, false},
		{"measured carrying a reason", func() Signal { s := measured; s.MissingReason = "x"; return s }, false},
		{"zero version", func() Signal { s := measured; s.Version = 0; return s }, false},
		{"no snapshot", func() Signal { s := measured; s.Snapshot = ""; return s }, false},
		{"no claim", func() Signal { s := measured; s.Claim = ""; return s }, false},
		{"no tool digest", func() Signal { s := measured; s.ToolDigest = ""; return s }, false},
		{"unknown missingness", func() Signal { s := missing; s.Missing = "bogus"; return s }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.signal().Validate(); (err == nil) != tc.ok {
				t.Fatalf("Validate() = %v, want ok=%v", err, tc.ok)
			}
		})
	}
}

func TestAccountedStatusesAgreeWithAccounted(t *testing.T) {
	listed := map[CoverageStatus]bool{}
	for _, status := range AccountedStatuses() {
		listed[status] = true
		if !status.Accounted() {
			t.Errorf("%q is listed but not accounted", status)
		}
	}
	for _, status := range []CoverageStatus{CoverageSkipped, CoverageBlocked, CoverageIncomplete, CoverageRequiresManualVerification} {
		if listed[status] || status.Accounted() {
			t.Errorf("%q must not count as accounted", status)
		}
	}
}
