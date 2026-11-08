package main

import "testing"

func TestHeterogeneousInexactRatioClassification(t *testing.T) {
	values := []Triple{{10, 2, 0}, {30, 6, 0}}
	for _, indices := range [][]int{{0, 1}, {0, 0}, {1, 1}, {0, 1, 1}} {
		sample := make([]Triple, len(indices))
		for i, index := range indices { sample[i] = values[index] }
		_, variance, ok := fit(sample)
		if !ok || variance != 0 { t.Fatalf("zero classification: indices=%v variance=%g", indices, variance) }
	}
	_, variance, ok := fit([]Triple{{10, 2, 0}, {30, 6.000000000001, 0}})
	if !ok || variance <= 0 { t.Fatalf("nonzero classification: variance=%g", variance) }
}
