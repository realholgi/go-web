package basepath

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalize(t *testing.T) {
	for input, want := range map[string]string{
		"": "/", ".": "/", "/": "/", "app": "/app", "/app/": "/app", "/app//v1": "/app/v1",
	} {
		if got := Normalize(input); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestBuilderURL(t *testing.T) {
	standard := New("/app/", false)
	if got := standard.URL("/"); got != "/app" {
		t.Fatalf("standard root URL = %q, want /app", got)
	}
	trailing := New("/app/", true)
	for target, want := range map[string]string{"/": "/app/", "items": "/app/items", "/items": "/app/items"} {
		if got := trailing.URL(target); got != want {
			t.Errorf("URL(%q) = %q, want %q", target, got, want)
		}
	}
}

func TestMountRoutesMountedBareAndFallbackPaths(t *testing.T) {
	app := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("app:" + r.URL.Path))
	})
	bare := TrailingSlashRedirect("/app")
	fallback := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("fallback:" + r.URL.Path))
	})
	handler := Mount("/app", app, bare, fallback)

	for requestPath, want := range map[string]struct {
		code int
		body string
	}{
		"/app/items": {http.StatusOK, "app:/items"},
		"/outside":   {http.StatusOK, "fallback:/outside"},
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, requestPath, nil))
		if recorder.Code != want.code || recorder.Body.String() != want.body {
			t.Errorf("GET %s = (%d, %q), want (%d, %q)", requestPath, recorder.Code, recorder.Body.String(), want.code, want.body)
		}
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/app?x=1", nil))
	if recorder.Code != http.StatusFound || recorder.Header().Get("Location") != "/app/?x=1" {
		t.Fatalf("bare redirect = (%d, %q)", recorder.Code, recorder.Header().Get("Location"))
	}

	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/app", nil))
	if got := recorder.Header().Get("Location"); got != "/app/" {
		t.Fatalf("second bare redirect location = %q, want /app/", got)
	}
}
