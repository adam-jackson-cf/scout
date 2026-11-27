package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"scout/internal/domain"
	"scout/internal/state"
)

func openStoreWithFeature(t *testing.T, featureJSON string) *state.Store {
	t.Helper()
	root := t.TempDir()
	store, err := state.Open(root)
	if err != nil {
		t.Fatalf("open state: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close state: %v", err)
		}
	})
	featuresDir := filepath.Join(root, ".scout", "clawpatch", "features")
	if err := os.MkdirAll(featuresDir, 0o700); err != nil {
		t.Fatalf("create canonical features directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(featuresDir, "feature-1.json"), []byte(featureJSON), 0o600); err != nil {
		t.Fatalf("write canonical feature: %v", err)
	}
	return store
}

func TestFeaturesLoadsCanonicalClawpatchRecords(t *testing.T) {
	store := openStoreWithFeature(t, `{"schemaVersion":1,"featureId":"feature-1","kind":"service","source":"node","confidence":"high"}`)

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

func assertRedactedWithSecurityHeaders(t *testing.T, target string, response *httptest.ResponseRecorder, secret string) {
	t.Helper()
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), secret) {
		t.Fatalf("%s = %d, secret visible: %v", target, response.Code, strings.Contains(response.Body.String(), secret))
	}
	if response.Header().Get("Content-Security-Policy") == "" || response.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("%s lacks security headers: %v", target, response.Header())
	}
}

func TestHandlerRejectsUnsafeRequestsAndRedactsStateOutput(t *testing.T) {
	const secret = "hunter2hunter2"
	store := openStoreWithFeature(t, `{"schemaVersion":1,"featureId":"feature-1","kind":"password=`+secret+`","source":"node","confidence":"high"}`)
	serve := func(method, target string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		New(store).ServeHTTP(response, httptest.NewRequest(method, target, nil))
		return response
	}

	if response := serve(http.MethodPost, "/api/features"); response.Code != http.StatusMethodNotAllowed || response.Header().Get("Allow") != http.MethodGet {
		t.Fatalf("POST = %d Allow %q, want 405 with Allow: GET", response.Code, response.Header().Get("Allow"))
	}
	if response := serve(http.MethodGet, "/api/issues/not%20valid"); response.Code != http.StatusBadRequest {
		t.Fatalf("invalid issue ID = %d, want 400", response.Code)
	}
	for _, target := range []string{"/api/features", "/features"} {
		assertRedactedWithSecurityHeaders(t, target, serve(http.MethodGet, target), secret)
	}
}
