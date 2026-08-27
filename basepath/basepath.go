// Package basepath builds application URLs and mounts applications below a URL path.
package basepath

import (
	"net/http"
	"path"
	"strings"
)

// Builder builds URLs for an application mounted below BasePath.
type Builder struct {
	BasePath          string
	RootTrailingSlash bool
}

// New returns a Builder with a normalized base path.
func New(raw string, rootTrailingSlash bool) Builder {
	return Builder{BasePath: Normalize(raw), RootTrailingSlash: rootTrailingSlash}
}

// Normalize returns a clean, absolute base path.
func Normalize(raw string) string {
	if raw == "" {
		return "/"
	}
	clean := path.Clean(raw)
	if clean == "." {
		return "/"
	}
	if !strings.HasPrefix(clean, "/") {
		return "/" + clean
	}
	return clean
}

// URL returns target prefixed with the application base path.
func (b Builder) URL(target string) string {
	target = Normalize(target)
	if b.BasePath == "/" {
		return target
	}
	if target == "/" {
		if b.RootTrailingSlash {
			return b.BasePath + "/"
		}
		return b.BasePath
	}
	return b.BasePath + target
}

// Redirect redirects to an application-relative target.
func (b Builder) Redirect(w http.ResponseWriter, r *http.Request, target string, code int) {
	http.Redirect(w, r, b.URL(target), code)
}

// TrailingSlashRedirect redirects a bare base path to its trailing-slash form
// while retaining the original query string.
func TrailingSlashRedirect(basePath string) http.Handler {
	target := Normalize(basePath) + "/"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		location := target
		if r.URL.RawQuery != "" {
			location += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, location, http.StatusFound)
	})
}

// Mount serves app below basePath. bare handles requests to the base path
// without its trailing slash; fallback handles paths outside the base path.
// A nil handler leaves that route unregistered.
func Mount(basePath string, app, bare, fallback http.Handler) http.Handler {
	basePath = Normalize(basePath)
	if basePath == "/" {
		return app
	}

	mux := http.NewServeMux()
	mux.Handle(basePath+"/", http.StripPrefix(basePath, app))
	if bare != nil {
		mux.Handle("GET "+basePath, bare)
	}
	if fallback != nil {
		mux.Handle("/", fallback)
	}
	return mux
}
