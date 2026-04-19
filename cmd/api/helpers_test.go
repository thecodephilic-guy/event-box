package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"thecodephilic-guy/eventbox/internal/jsonlog"
	"thecodephilic-guy/eventbox/internal/validator"
)

func newTestApplication() *application {
	logger := jsonlog.New(bytes.NewBuffer(nil), jsonlog.LevelInfo)
	return &application{
		config: config{
			env: "testing",
		},
		logger: logger,
	}
}

func TestWriteJSON(t *testing.T) {
	app := newTestApplication()

	t.Run("basic response", func(t *testing.T) {
		rr := httptest.NewRecorder()
		data := envelop{"status": "ok"}
		err := app.writeJSON(rr, http.StatusOK, data, nil)
		if err != nil {
			t.Fatalf("writeJSON() returned error: %v", err)
		}

		if rr.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rr.Code)
		}

		if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %q", ct)
		}

		var result map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &result); err != nil {
			t.Fatalf("failed to parse JSON response: %v", err)
		}

		if result["status"] != "ok" {
			t.Errorf("expected status 'ok', got %v", result["status"])
		}
	})

	t.Run("with custom headers", func(t *testing.T) {
		rr := httptest.NewRecorder()
		headers := make(http.Header)
		headers.Set("X-Custom", "test-value")

		err := app.writeJSON(rr, http.StatusCreated, envelop{"id": 1}, headers)
		if err != nil {
			t.Fatalf("writeJSON() returned error: %v", err)
		}

		if rr.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", rr.Code)
		}
		if h := rr.Header().Get("X-Custom"); h != "test-value" {
			t.Errorf("expected X-Custom header 'test-value', got %q", h)
		}
	})
}

func TestReadJSON(t *testing.T) {
	app := newTestApplication()

	t.Run("valid JSON", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name": "test"}`)
		r := httptest.NewRequest(http.MethodPost, "/", body)
		w := httptest.NewRecorder()

		var dst struct {
			Name string `json:"name"`
		}

		err := app.readJSON(w, r, &dst)
		if err != nil {
			t.Fatalf("readJSON() returned error: %v", err)
		}
		if dst.Name != "test" {
			t.Errorf("expected name 'test', got %q", dst.Name)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		body := bytes.NewBufferString("")
		r := httptest.NewRequest(http.MethodPost, "/", body)
		w := httptest.NewRecorder()

		var dst struct{}
		err := app.readJSON(w, r, &dst)
		if err == nil {
			t.Error("expected error for empty body")
		}
		if err.Error() != "body must not be empty" {
			t.Errorf("expected 'body must not be empty', got %q", err.Error())
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		body := bytes.NewBufferString(`{bad json}`)
		r := httptest.NewRequest(http.MethodPost, "/", body)
		w := httptest.NewRecorder()

		var dst struct{}
		err := app.readJSON(w, r, &dst)
		if err == nil {
			t.Error("expected error for malformed JSON")
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		body := bytes.NewBufferString(`{"unknown": "field"}`)
		r := httptest.NewRequest(http.MethodPost, "/", body)
		w := httptest.NewRecorder()

		var dst struct {
			Name string `json:"name"`
		}
		err := app.readJSON(w, r, &dst)
		if err == nil {
			t.Error("expected error for unknown field")
		}
	})

	t.Run("multiple JSON values", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name":"a"}{"name":"b"}`)
		r := httptest.NewRequest(http.MethodPost, "/", body)
		w := httptest.NewRecorder()

		var dst struct {
			Name string `json:"name"`
		}
		err := app.readJSON(w, r, &dst)
		if err == nil {
			t.Error("expected error for multiple JSON values")
		}
		if err.Error() != "body must only contain a single JSON value" {
			t.Errorf("expected 'body must only contain a single JSON value', got %q", err.Error())
		}
	})

	t.Run("wrong type", func(t *testing.T) {
		body := bytes.NewBufferString(`{"name": 123}`)
		r := httptest.NewRequest(http.MethodPost, "/", body)
		w := httptest.NewRecorder()

		var dst struct {
			Name string `json:"name"`
		}
		err := app.readJSON(w, r, &dst)
		if err == nil {
			t.Error("expected error for wrong type")
		}
	})
}

func TestReadString(t *testing.T) {
	app := newTestApplication()

	r := httptest.NewRequest(http.MethodGet, "/?title=hello", nil)
	qs := r.URL.Query()

	if got := app.readString(qs, "title", "default"); got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
	if got := app.readString(qs, "missing", "default"); got != "default" {
		t.Errorf("expected 'default', got %q", got)
	}
}

func TestReadInt(t *testing.T) {
	app := newTestApplication()

	t.Run("valid int", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/?page=5", nil)
		qs := r.URL.Query()
		v := &validator.Validator{Errors: make(map[string]string)}
		if got := app.readInt(qs, "page", 1, v); got != 5 {
			t.Errorf("expected 5, got %d", got)
		}
		if len(v.Errors) != 0 {
			t.Errorf("expected no errors, got %v", v.Errors)
		}
	})

	t.Run("missing key returns default", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		qs := r.URL.Query()
		v := &validator.Validator{Errors: make(map[string]string)}
		if got := app.readInt(qs, "page", 42, v); got != 42 {
			t.Errorf("expected 42, got %d", got)
		}
	})

	t.Run("non-int adds error", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/?page=abc", nil)
		qs := r.URL.Query()
		v := &validator.Validator{Errors: make(map[string]string)}
		got := app.readInt(qs, "page", 1, v)
		if got != 1 {
			t.Errorf("expected default 1, got %d", got)
		}
		if _, ok := v.Errors["page"]; !ok {
			t.Error("expected validation error for non-int")
		}
	})
}

func TestReadCSV(t *testing.T) {
	app := newTestApplication()

	t.Run("comma separated", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/?genres=action,drama,comedy", nil)
		qs := r.URL.Query()
		got := app.readCSV(qs, "genres", []string{})
		if len(got) != 3 || got[0] != "action" || got[1] != "drama" || got[2] != "comedy" {
			t.Errorf("expected [action drama comedy], got %v", got)
		}
	})

	t.Run("missing returns default", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		qs := r.URL.Query()
		def := []string{"default"}
		got := app.readCSV(qs, "genres", def)
		if len(got) != 1 || got[0] != "default" {
			t.Errorf("expected [default], got %v", got)
		}
	})
}
