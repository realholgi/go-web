// Package render renders embedded HTML templates with application-relative URLs.
package render

import (
	"bytes"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"sync"

	"github.com/realholgi/go-web/basepath"
)

//go:embed base.html
var baseFS embed.FS

// Config configures a Renderer. Funcs are available while templates are parsed.
// RequestFuncs may replace those functions for an individual response.
type Config struct {
	BasePath     basepath.Builder
	Funcs        template.FuncMap
	RequestFuncs func(http.ResponseWriter) template.FuncMap
}

// Renderer parses each page template once and safely reuses it across requests.
type Renderer struct {
	templateFS    fs.FS
	basePath      basepath.Builder
	funcs         template.FuncMap
	requestFuncs  func(http.ResponseWriter) template.FuncMap
	templateCache sync.Map
}

// New creates a renderer for application templates rooted at templates/chrome.html.
func New(templateFS fs.FS, cfg Config) *Renderer {
	return &Renderer{
		templateFS:   templateFS,
		basePath:     cfg.BasePath,
		funcs:        cloneFuncs(cfg.Funcs),
		requestFuncs: cfg.RequestFuncs,
	}
}

// BasePath returns the normalized application mount path.
func (r *Renderer) BasePath() string {
	return r.basePath.BasePath
}

// AppURL returns an application-relative URL.
func (r *Renderer) AppURL(target string) string {
	return r.basePath.URL(target)
}

// Redirect redirects to an application-relative URL.
func (r *Renderer) Redirect(w http.ResponseWriter, req *http.Request, target string, code int) {
	r.basePath.Redirect(w, req, target, code)
}

// Render renders name using the shared base template, templates/chrome.html, and templates/name.
func (r *Renderer) Render(w http.ResponseWriter, name string, data any) {
	cachedValue, ok := r.templateCache.Load(name)
	if !ok {
		funcs := cloneFuncs(r.funcs)
		funcs["url"] = r.AppURL
		parsed, err := parseTemplate(funcs, r.templateFS, name)
		if err != nil {
			serverError(w, err, "render parse "+name)
			return
		}
		cachedValue, _ = r.templateCache.LoadOrStore(name, parsed)
	}
	cached := cachedValue.(*template.Template)

	if r.requestFuncs != nil {
		if funcs := r.requestFuncs(w); len(funcs) > 0 {
			clone, err := cached.Clone()
			if err != nil {
				serverError(w, err, "render clone "+name)
				return
			}
			cached = clone.Funcs(funcs)
		}
	}

	var buf bytes.Buffer
	if err := cached.ExecuteTemplate(&buf, "base", data); err != nil {
		serverError(w, err, "render template "+name)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}

func parseTemplate(funcs template.FuncMap, templateFS fs.FS, name string) (*template.Template, error) {
	parsed, err := template.New("").Funcs(funcs).ParseFS(baseFS, "base.html")
	if err != nil {
		return nil, err
	}
	return parsed.ParseFS(templateFS, "templates/chrome.html", "templates/"+name)
}

func cloneFuncs(source template.FuncMap) template.FuncMap {
	clone := make(template.FuncMap, len(source))
	for name, function := range source {
		clone[name] = function
	}
	return clone
}

func serverError(w http.ResponseWriter, err error, operation string) {
	log.Printf("%s: %v", operation, err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}
