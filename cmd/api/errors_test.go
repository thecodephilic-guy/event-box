package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorResponse(t *testing.T) {
	app := newTestApplication()

	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	app.errorResponse(rr, r, http.StatusBadRequest, "test error")

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}

	var result map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if result["error"] != "test error" {
		t.Errorf("expected error 'test error', got %v", result["error"])
	}
}

func TestNotFoundResponse(t *testing.T) {
	app := newTestApplication()
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	app.notFoundResponse(rr, r)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", rr.Code)
	}
}

func TestMethodNotAllowedResponse(t *testing.T) {
	app := newTestApplication()
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	app.methodNotAllowedResponse(rr, r)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status 405, got %d", rr.Code)
	}
}

func TestBadRequestResponse(t *testing.T) {
	app := newTestApplication()
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	app.badRequestResponse(rr, r, errForTest("bad input"))

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestFailedValidationResponse(t *testing.T) {
	app := newTestApplication()
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	errors := map[string]string{"name": "must be provided"}
	app.failedValidationResponse(rr, r, errors)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status 422, got %d", rr.Code)
	}
}

func TestEditConflictResponse(t *testing.T) {
	app := newTestApplication()
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPatch, "/", nil)

	app.editConflictResponse(rr, r)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", rr.Code)
	}
}

func TestRateLimitExceededResponse(t *testing.T) {
	app := newTestApplication()
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	app.rateLimitExceededResponse(rr, r)

	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", rr.Code)
	}
}

func TestInvalidCredentialsResponse(t *testing.T) {
	app := newTestApplication()
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/", nil)

	app.invalidCredentialsResponse(rr, r)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestInvalidAuthenticationTokenResponse(t *testing.T) {
	app := newTestApplication()
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	app.invalidAuthenticationTokenResponse(rr, r)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
	if h := rr.Header().Get("WWW-Authenticate"); h != "Bearer" {
		t.Errorf("expected WWW-Authenticate header 'Bearer', got %q", h)
	}
}

func TestAuthenticationRequiredResponse(t *testing.T) {
	app := newTestApplication()
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	app.authenticationRequiredResponse(rr, r)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", rr.Code)
	}
}

func TestNotPermittedResponse(t *testing.T) {
	app := newTestApplication()
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	app.notPermittedResponse(rr, r)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}
}

func TestRequireRoleResponse(t *testing.T) {
	app := newTestApplication()
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	app.requireRoleResponse(rr, r, "organizer")

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rr.Code)
	}

	var result map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	errMsg, ok := result["error"].(string)
	if !ok {
		t.Fatal("expected error to be a string")
	}
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

// errForTest is a simple error type for testing.
type errForTest string

func (e errForTest) Error() string {
	return string(e)
}
