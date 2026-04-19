package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"thecodephilic-guy/eventbox/internal/data"
)

func TestRecoverPanic(t *testing.T) {
	app := newTestApplication()

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	handler := app.recoverPanic(panicHandler)

	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "127.0.0.1:12345"

	handler.ServeHTTP(rr, r)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rr.Code)
	}

	if h := rr.Header().Get("Connection"); h != "close" {
		t.Errorf("expected Connection: close header, got %q", h)
	}
}

func TestRateLimitEnabled(t *testing.T) {
	app := newTestApplication()
	app.config.limiter.enabled = true
	app.config.limiter.rps = 1
	app.config.limiter.burst = 2

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := app.rateLimit(okHandler)

	// First two requests should pass (burst of 2).
	for i := 0; i < 2; i++ {
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = "10.0.0.1:12345"
		handler.ServeHTTP(rr, r)
		if rr.Code != http.StatusOK {
			t.Errorf("request %d: expected 200, got %d", i+1, rr.Code)
		}
	}

	// Third request should be rate-limited.
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:12345"
	handler.ServeHTTP(rr, r)
	if rr.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 (rate limited), got %d", rr.Code)
	}
}

func TestRateLimitDisabled(t *testing.T) {
	app := newTestApplication()
	app.config.limiter.enabled = false

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := app.rateLimit(okHandler)

	// Many requests should all pass when rate limiting is disabled.
	for i := 0; i < 100; i++ {
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.RemoteAddr = "10.0.0.1:12345"
		handler.ServeHTTP(rr, r)
		if rr.Code != http.StatusOK {
			t.Errorf("request %d: expected 200 with limiter disabled, got %d", i+1, rr.Code)
			break // No need to continue
		}
	}
}

func TestRequireAuthenticatedUser(t *testing.T) {
	app := newTestApplication()

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := app.requireAuthenticatedUser(okHandler)

	t.Run("anonymous user", func(t *testing.T) {
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = app.contextSetUser(r, data.AnonymousUser)

		handler.ServeHTTP(rr, r)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 for anonymous user, got %d", rr.Code)
		}
	})

	t.Run("authenticated user", func(t *testing.T) {
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		user := &data.User{ID: 1, Name: "Test", Role: "customer"}
		r = app.contextSetUser(r, user)

		handler.ServeHTTP(rr, r)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 for authenticated user, got %d", rr.Code)
		}
	})
}

func TestRequireRole(t *testing.T) {
	app := newTestApplication()

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	t.Run("correct role", func(t *testing.T) {
		handler := app.requireRole("organizer", okHandler)
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		user := &data.User{ID: 1, Name: "Organizer", Role: "organizer"}
		r = app.contextSetUser(r, user)

		handler.ServeHTTP(rr, r)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 for correct role, got %d", rr.Code)
		}
	})

	t.Run("wrong role", func(t *testing.T) {
		handler := app.requireRole("organizer", okHandler)
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		user := &data.User{ID: 1, Name: "Customer", Role: "customer"}
		r = app.contextSetUser(r, user)

		handler.ServeHTTP(rr, r)

		if rr.Code != http.StatusForbidden {
			t.Errorf("expected 403 for wrong role, got %d", rr.Code)
		}
	})

	t.Run("anonymous user", func(t *testing.T) {
		handler := app.requireRole("organizer", okHandler)
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = app.contextSetUser(r, data.AnonymousUser)

		handler.ServeHTTP(rr, r)

		if rr.Code != http.StatusUnauthorized {
			t.Errorf("expected 401 for anonymous user, got %d", rr.Code)
		}
	})
}

func TestEnableCORS(t *testing.T) {
	app := newTestApplication()
	app.config.cors.trustedOrigins = []string{"https://example.com"}

	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := app.enableCORS(okHandler)

	t.Run("trusted origin", func(t *testing.T) {
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Origin", "https://example.com")

		handler.ServeHTTP(rr, r)

		if h := rr.Header().Get("Access-Control-Allow-Origin"); h != "https://example.com" {
			t.Errorf("expected Access-Control-Allow-Origin 'https://example.com', got %q", h)
		}
	})

	t.Run("untrusted origin", func(t *testing.T) {
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Origin", "https://evil.com")

		handler.ServeHTTP(rr, r)

		if h := rr.Header().Get("Access-Control-Allow-Origin"); h != "" {
			t.Errorf("expected no Access-Control-Allow-Origin, got %q", h)
		}
	})

	t.Run("preflight request", func(t *testing.T) {
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodOptions, "/", nil)
		r.Header.Set("Origin", "https://example.com")
		r.Header.Set("Access-Control-Request-Method", "PUT")

		handler.ServeHTTP(rr, r)

		if rr.Code != http.StatusOK {
			t.Errorf("expected 200 for preflight, got %d", rr.Code)
		}
		if h := rr.Header().Get("Access-Control-Allow-Methods"); h == "" {
			t.Error("expected Access-Control-Allow-Methods header for preflight")
		}
	})

	t.Run("no origin header", func(t *testing.T) {
		rr := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/", nil)

		handler.ServeHTTP(rr, r)

		if h := rr.Header().Get("Access-Control-Allow-Origin"); h != "" {
			t.Errorf("expected no Access-Control-Allow-Origin without Origin header, got %q", h)
		}
	})
}
