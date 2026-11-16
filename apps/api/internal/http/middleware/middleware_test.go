package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := RequestIDFromContext(r.Context())
		if id == "" {
			t.Error("expected request id in context, got empty")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get(headerRequestID) == "" {
		t.Error("expected X-Request-ID response header")
	}
}

func TestSecurityHeaders(t *testing.T) {
	handler := SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected nosniff, got %s", rec.Header().Get("X-Content-Type-Options"))
	}
	if rec.Header().Get("X-Frame-Options") != "DENY" {
		t.Errorf("expected DENY, got %s", rec.Header().Get("X-Frame-Options"))
	}
	if rec.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Errorf("expected strict-origin-when-cross-origin, got %s", rec.Header().Get("Referrer-Policy"))
	}
}

func TestCORS(t *testing.T) {
	allowed := []string{"https://app.mjengofiti.org", "http://localhost:5173"}
	corsMiddleware := CORS(allowed)

	handler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Allowed origin
	reqAllowed := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqAllowed.Header.Set("Origin", "https://app.mjengofiti.org")
	recAllowed := httptest.NewRecorder()
	handler.ServeHTTP(recAllowed, reqAllowed)

	if recAllowed.Header().Get("Access-Control-Allow-Origin") != "https://app.mjengofiti.org" {
		t.Errorf("expected allowed origin header, got %s", recAllowed.Header().Get("Access-Control-Allow-Origin"))
	}

	// Disallowed origin
	reqBlocked := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqBlocked.Header.Set("Origin", "https://malicious.example.com")
	recBlocked := httptest.NewRecorder()
	handler.ServeHTTP(recBlocked, reqBlocked)

	if recBlocked.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected empty CORS header for untrusted origin, got %s", recBlocked.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestRecover(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	recoverMiddleware := Recover(logger)

	handler := recoverMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("simulated critical crash")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 on panic recovery, got %d", rec.Code)
	}
}

func TestRateLimiter(t *testing.T) {
	// 60 requests per minute with burst capacity of 3
	limiter := NewRateLimiter(60, 3)
	handler := limiter.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	ip := "192.0.2.1:1234"

	// First 3 requests (burst) must pass
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = ip
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d should succeed within burst limit, got %d", i+1, rec.Code)
		}
	}

	// 4th request must be rejected with 429 Too Many Requests
	reqBlocked := httptest.NewRequest(http.MethodGet, "/test", nil)
	reqBlocked.RemoteAddr = ip
	recBlocked := httptest.NewRecorder()
	handler.ServeHTTP(recBlocked, reqBlocked)

	if recBlocked.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 Too Many Requests after burst exhaustion, got %d", recBlocked.Code)
	}

	if recBlocked.Header().Get("Retry-After") != "60" {
		t.Errorf("expected Retry-After header 60, got %s", recBlocked.Header().Get("Retry-After"))
	}
}
