package sameorigin

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequire(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*http.Request)
		status int
	}{
		{
			name:   "allows matching origin",
			setup:  func(r *http.Request) { r.Header.Set("Origin", "http://example.test") },
			status: http.StatusNoContent,
		},
		{
			name:   "allows matching referer",
			setup:  func(r *http.Request) { r.Header.Set("Referer", "http://example.test/form") },
			status: http.StatusNoContent,
		},
		{
			name: "allows forwarded HTTPS",
			setup: func(r *http.Request) {
				r.Header.Set("X-Forwarded-Proto", "https")
				r.Header.Set("Origin", "https://example.test")
			},
			status: http.StatusNoContent,
		},
		{
			name:   "rejects foreign origin",
			setup:  func(r *http.Request) { r.Header.Set("Origin", "http://evil.test") },
			status: http.StatusForbidden,
		},
		{
			name:   "rejects scheme mismatch",
			setup:  func(r *http.Request) { r.Header.Set("Origin", "https://example.test") },
			status: http.StatusForbidden,
		},
		{
			name:   "rejects missing headers",
			setup:  func(*http.Request) {},
			status: http.StatusForbidden,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "http://example.test/resource", nil)
			req.Host = "example.test"
			test.setup(req)
			recorder := httptest.NewRecorder()

			Require(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})(recorder, req)

			if recorder.Code != test.status {
				t.Fatalf("status = %d, want %d", recorder.Code, test.status)
			}
		})
	}
}
