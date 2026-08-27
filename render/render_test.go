package render

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/realholgi/go-web/basepath"
)

const chrome = `{{define "home_url"}}{{url "/"}}{{end}}
{{define "navigation"}}
<li><a class="nav-link" href="{{url "/items"}}">Items</a></li>
{{end}}
{{define "navbar_right"}}<li class="nav-item">{{companyName}}</li>{{end}}
{{define "app_styles"}}.custom { color: red; }{{end}}`

func testFS(page string) fstest.MapFS {
	return fstest.MapFS{
		"templates/chrome.html": &fstest.MapFile{Data: []byte(chrome)},
		"templates/page.html":   &fstest.MapFile{Data: []byte(page)},
	}
}

func testConfig(title string) Config {
	return Config{
		BasePath: basepath.New("/app", false),
		Funcs: template.FuncMap{
			"appTitle":        func() string { return title },
			"appVersion":      func() string { return "v1" },
			"brandColor":      func() string { return "#112233" },
			"brandColorHover": func() string { return "#223344" },
			"companyName":     func() string { return "Example GmbH" },
			"changelogTrigger": func() template.HTML {
				return template.HTML(`<button type="button" data-changelog-trigger="1">v1</button>`)
			},
			"changelogModal": func() template.HTML { return template.HTML("") },
		},
	}
}

func TestRenderUsesSharedBaseAndApplicationChrome(t *testing.T) {
	renderer := New(testFS(`{{define "title"}}Items{{end}}
{{define "content"}}<main>{{url "/items"}} {{appTitle}}</main>{{end}}`), testConfig("Example"))

	recorder := httptest.NewRecorder()
	renderer.Render(recorder, "page.html", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		`<title>Items</title>`,
		`<a class="navbar-brand p-0 d-flex align-items-center gap-2" href="/app">`,
		`<link href="/app/static/vendor/bootstrap/bootstrap.min.css" rel="stylesheet">`,
		`alt="Example GmbH"`,
		`<a class="nav-link" href="/app/items">Items</a>`,
		`<li class="nav-item">Example GmbH</li>`,
		`.custom { color: red; }`,
		`<main>/app/items Example</main>`,
		`<script src="/app/static/vendor/bootstrap/bootstrap.bundle.min.js">`,
		`<button type="button" data-changelog-trigger="1">v1</button>`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered page does not contain %q:\n%s", want, body)
		}
	}
	for _, unwanted := range []string{"KGW", "Isotherm", "clientName"} {
		if strings.Contains(body, unwanted) {
			t.Errorf("rendered page must not contain %q", unwanted)
		}
	}
	if got := recorder.Header().Get("Content-Type"); got != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q", got)
	}
}

func TestRenderUsesDefaultTitleAndNavigationHighlighting(t *testing.T) {
	renderer := New(testFS(`{{define "content"}}{{end}}`), testConfig("Example"))

	recorder := httptest.NewRecorder()
	renderer.Render(recorder, "page.html", nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d", recorder.Code)
	}
	body := recorder.Body.String()
	for _, want := range []string{
		`<title>Example</title>`,
		`location.pathname.replace(/\/+$/, '')`,
		`a.dataset.exact !== '1'`,
	} {
		if !strings.Contains(body, want) {
			t.Errorf("rendered page does not contain %q:\n%s", want, body)
		}
	}
}

func TestRenderUsesRequestFunctionsWithoutLeakingValues(t *testing.T) {
	config := testConfig("Example")
	config.RequestFuncs = func(w http.ResponseWriter) template.FuncMap {
		return template.FuncMap{"companyName": func() string { return w.Header().Get("X-Client-Name") }}
	}
	renderer := New(testFS(`{{define "content"}}{{end}}`), config)

	for _, name := range []string{"Ada", "Berta"} {
		recorder := httptest.NewRecorder()
		recorder.Header().Set("X-Client-Name", name)
		renderer.Render(recorder, "page.html", nil)
		if got := recorder.Body.String(); !strings.Contains(got, ">"+name) {
			t.Fatalf("rendered page = %q, want request value %q", got, name)
		}
	}
}
