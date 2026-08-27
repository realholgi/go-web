package render

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/realholgi/go-web/basepath"
)

func TestRenderUsesBasePathAndConfiguredFunctions(t *testing.T) {
	renderer := New(fstest.MapFS{
		"templates/base.html": &fstest.MapFile{Data: []byte(`{{define "base"}}{{template "content" .}}{{end}}`)},
		"templates/page.html": &fstest.MapFile{Data: []byte(`{{define "content"}}{{url "/items"}} {{appTitle}}{{end}}`)},
	}, Config{
		BasePath: basepath.New("/app", false),
		Funcs: template.FuncMap{
			"appTitle": func() string { return "Example" },
		},
	})

	recorder := httptest.NewRecorder()
	renderer.Render(recorder, "page.html", nil)
	if recorder.Code != http.StatusOK || recorder.Body.String() != "/app/items Example" {
		t.Fatalf("render = (%d, %q)", recorder.Code, recorder.Body.String())
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
}

func TestRenderUsesRequestFunctionsWithoutLeakingValues(t *testing.T) {
	renderer := New(fstest.MapFS{
		"templates/base.html": &fstest.MapFile{Data: []byte(`{{define "base"}}{{template "content" .}}{{end}}`)},
		"templates/page.html": &fstest.MapFile{Data: []byte(`{{define "content"}}{{clientName}}{{end}}`)},
	}, Config{
		BasePath: basepath.New("/", false),
		Funcs: template.FuncMap{
			"clientName": func() string { return "" },
		},
		RequestFuncs: func(w http.ResponseWriter) template.FuncMap {
			return template.FuncMap{"clientName": func() string { return w.Header().Get("X-Client-Name") }}
		},
	})

	for _, name := range []string{"Ada", "Berta"} {
		recorder := httptest.NewRecorder()
		recorder.Header().Set("X-Client-Name", name)
		renderer.Render(recorder, "page.html", nil)
		if got := recorder.Body.String(); got != name {
			t.Errorf("rendered client name = %q, want %q", got, name)
		}
	}
}
