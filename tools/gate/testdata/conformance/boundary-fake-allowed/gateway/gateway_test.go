package gateway

import "testing"

type recordingGateway struct{ charged int }

func (r *recordingGateway) Charge(cents int) error {
	r.charged += cents
	return nil
}

func TestBillChargesTheGateway(t *testing.T) {
	g := &recordingGateway{}
	if !Bill(g, 5) || g.charged != 5 {
		t.Fatal("charge")
	}
}
