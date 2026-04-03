package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIsOriginAllowed(t *testing.T) {
	allowed := []string{
		"https://fitness-tracker.me",
		"http://localhost:5173",
		"capacitor://localhost",
		"*.staging.fitness-tracker.me",
	}
	c := NewCORSMiddleware(allowed)

	tests := []struct {
		origin string
		want   bool
	}{
		// Exact matches
		{"https://fitness-tracker.me", true},
		{"http://localhost:5173", true},
		{"capacitor://localhost", true},
		// www variants should be treated the same as apex
		{"https://www.fitness-tracker.me", true},
		{"http://www.localhost:5173", true}, // strips to http://localhost:5173 which is in list
		// Unlisted origins
		{"https://evil.com", false},
		{"https://www.evil.com", false},
		// Wildcard subdomain
		{"https://app.staging.fitness-tracker.me", true},
		// Empty origin
		{"", false},
		// www on a different domain
		{"https://www.other-tracker.me", false},
	}

	for _, tt := range tests {
		t.Run(tt.origin, func(t *testing.T) {
			got := c.isOriginAllowed(tt.origin)
			if got != tt.want {
				t.Errorf("isOriginAllowed(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}

func TestStripWWW(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://www.foo.com", "https://foo.com"},
		{"http://www.foo.com", "http://foo.com"},
		{"https://foo.com", "https://foo.com"},
		{"https://www.foo.com/path", "https://foo.com/path"},
		{"https://notawww.foo.com", "https://notawww.foo.com"},
		{"capacitor://localhost", "capacitor://localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := stripWWW(tt.input)
			if got != tt.want {
				t.Errorf("stripWWW(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCORSMiddlewareHandler_WWW(t *testing.T) {
	c := NewCORSMiddleware([]string{"https://fitness-tracker.me"})
	handler := c.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// www variant should get the ACAO header set
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://www.fitness-tracker.me")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://www.fitness-tracker.me" {
		t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, "https://www.fitness-tracker.me")
	}
}
