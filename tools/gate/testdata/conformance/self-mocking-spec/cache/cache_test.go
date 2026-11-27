package cache

import (
	"testing"

	"canary/store"
)

type partialStore struct {
	store.Store
}

func TestWarmReportsAMissingKey(t *testing.T) {
	if Warm(partialStore{Store: store.NewMemory(nil)}, "a") {
		t.Fatal("warm")
	}
}
