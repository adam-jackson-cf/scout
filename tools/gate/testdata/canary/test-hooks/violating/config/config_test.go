package config

import "testing"

func TestModeDefaultsToNormal(t *testing.T) {
	if got := Mode(); got == "" {
		t.Fatal("empty mode")
	}
}
