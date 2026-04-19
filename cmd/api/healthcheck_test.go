package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthcheckHandler(t *testing.T) {
	app := newTestApplication()

	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v1/healthcheck", nil)

	app.healthcheckHandler(rr, r)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}

	var result map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if result["status"] != "available" {
		t.Errorf("expected status 'available', got %v", result["status"])
	}

	systemInfo, ok := result["system_info"].(map[string]any)
	if !ok {
		t.Fatal("expected system_info to be a map")
	}

	if systemInfo["environment"] != "testing" {
		t.Errorf("expected environment 'testing', got %v", systemInfo["environment"])
	}

	if systemInfo["version"] != version {
		t.Errorf("expected version %q, got %v", version, systemInfo["version"])
	}
}
