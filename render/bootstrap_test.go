package render

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStaticHandlerServesEmbeddedAssets(t *testing.T) {
	for _, path := range []string{"/bootstrap/bootstrap.min.css", "/bootstrap/bootstrap.bundle.min.js"} {
		t.Run(path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			StaticHandler().ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", path, recorder.Code, http.StatusOK)
			}
			if recorder.Body.Len() == 0 {
				t.Fatalf("GET %s returned an empty body", path)
			}
		})
	}
}
