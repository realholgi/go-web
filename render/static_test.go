package render

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestStaticHandlerServesApplicationAndSharedAssets(t *testing.T) {
	handler := StaticHandler(fstest.MapFS{
		"static/logo.txt": &fstest.MapFile{Data: []byte("application logo")},
	})
	for _, test := range []struct {
		path string
		body string
	}{
		{path: "/logo.txt", body: "application logo"},
		{path: "/shared/bootstrap/bootstrap.min.css"},
		{path: "/shared/bootstrap/bootstrap.bundle.min.js"},
	} {
		t.Run(test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, test.path, nil))
			if recorder.Code != http.StatusOK {
				t.Fatalf("GET %s status = %d, want %d", test.path, recorder.Code, http.StatusOK)
			}
			if recorder.Body.Len() == 0 {
				t.Fatalf("GET %s returned an empty body", test.path)
			}
			if test.body != "" && recorder.Body.String() != test.body {
				t.Fatalf("GET %s body = %q, want %q", test.path, recorder.Body.String(), test.body)
			}
		})
	}
}
