# go-web

`go-web` provides small HTTP server helpers: a standard server configuration,
request logging and panic recovery middleware, Windows service support,
same-origin protection for state-changing routes, and embedded shared HTML
template rendering.

## Installation

```sh
go get github.com/realholgi/go-web
```

## Usage

The runnable example starts a server on `http://localhost:8080`:

```sh
go run ./example
```

```sh
curl http://localhost:8080/
curl -X POST -H 'Origin: http://localhost:8080' http://localhost:8080/settings
```

It assembles the helpers as follows:

```go
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", home)
	mux.HandleFunc("POST /settings", sameorigin.Require(updateSettings))

	handler := httpserver.WithRequestLoggingAndRecovery(mux)
	httpserver.Run(":8080", handler, "go-web-example")
}
```

`httpserver.WithRequestLoggingAndRecovery` logs each request and returns a 500
response if a handler panics before its response begins. `httpserver.Run`
starts the server directly on non-Windows platforms. On Windows, it detects the
service manager and uses `go-web-example` as the service name when applicable.

`sameorigin.Require` protects state-changing `http.HandlerFunc` routes. It
requires a matching `Origin` or `Referer` header, returning 403 otherwise. It
honors `X-Forwarded-Proto` for TLS-terminating proxies; only trust that header
when the proxy removes values supplied by clients.

## Shared templates

`render.New` embeds the shared `render/base.html`. Applications provide their
page templates plus `templates/chrome.html`, which defines `home_url`,
`navigation`, `navbar_right`, and `app_styles`. The renderer parses those files
for each page and supplies application-relative URLs through `url`.

```go
renderer := render.New(templateFS, render.Config{
    BasePath: basepath.New("/app", false),
    Funcs: template.FuncMap{
        "appTitle":         func() string { return "Example" },
        "companyName":      func() string { return "Example GmbH" },
        "brandColor":       func() string { return "#123456" },
        "brandColorHover":  func() string { return "#0d47a1" },
        "changelogTrigger": func() template.HTML { return "" },
        "changelogModal":   func() template.HTML { return "" },
    },
})
```

Assets below `render/static` are embedded in the package and served by
`render.StaticHandler`. Applications mount it separately from app-owned assets:

```go
mux.Handle(
    "GET /static/shared/",
    http.StripPrefix("/static/shared/", render.StaticHandler()),
)
```

## License

MIT
