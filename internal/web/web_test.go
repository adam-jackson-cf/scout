package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"scout/internal/domain"
	"scout/internal/state"
)

func TestFeaturesLoadsCanonicalClawpatchRecords(t *testing.T) {
	root := t.TempDir()
	store, err := state.Open(root)
	if err != nil {
		t.Fatalf("open state: %v", err)
	}
	defer store.Close()

	featuresDir := filepath.Join(root, ".scout", "clawpatch", "features")
	if err := os.MkdirAll(featuresDir, 0o700); err != nil {
		t.Fatalf("create canonical features directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(featuresDir, "feature-1.json"), []byte(`{"schemaVersion":1,"featureId":"feature-1","kind":"service","source":"node","confidence":"high"}`), 0o600); err != nil {
		t.Fatalf("write canonical feature: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/features", nil)
	response := httptest.NewRecorder()
	New(store).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("features status = %d, want %d", response.Code, http.StatusOK)
	}
	var payload struct {
		Features []domain.Feature `json:"features"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode features response: %v", err)
	}
	if len(payload.Features) != 1 || payload.Features[0].FeatureID != "feature-1" {
		t.Fatalf("features = %#v, want canonical feature-1", payload.Features)
	}
}
