# go-web

`go-web` provides small HTTP server helpers: a standard server configuration,
request logging and panic recovery middleware, Windows service support,
same-origin protection for state-changing routes, and embedded shared HTML
template rendering.

## Installation

```sh
go get github.com/realholgi/go-web
```

## Packages

| Package | Purpose |
| --- | --- |
| `httpserver` | Configures HTTP server timeouts, logs requests, recovers panics, and integrates with the Windows service manager. |
| `sameorigin` | Rejects cross-origin state-changing browser requests. |
| `basepath` | Builds URLs and mounts an application below a path prefix. |
| `render` | Renders application templates with the shared layout and serves shared static assets. |

## Runnable example

The example assembles a `ServeMux`, protects its state-changing route, then adds
request logging and panic recovery:

```sh
go run ./example
```

In another terminal, verify the public route and the origin check:

```sh
curl -i http://localhost:8080/
# HTTP/1.1 200 OK
# ...
# go-web example is running

curl -i -X POST http://localhost:8080/settings
# HTTP/1.1 403 Forbidden

curl -i -X POST -H 'Origin: http://localhost:8080' \
  http://localhost:8080/settings
# HTTP/1.1 204 No Content
```

The full source is [`example/main.go`](example/main.go):

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /", home)
mux.HandleFunc("POST /settings", sameorigin.Require(updateSettings))

handler := httpserver.WithRequestLoggingAndRecovery(mux)
httpserver.Run(":8080", handler, "go-web-example")
```

`httpserver.NewServer` applies read-header, read, write, and idle timeouts.
`httpserver.Run` uses that configuration and runs directly on non-Windows
platforms. On Windows, it detects the service manager and uses
`go-web-example` as the service name when applicable.

`httpserver.WithRequestLoggingAndRecovery` logs each request. It returns 500
when a handler panics before beginning its response; a panic after a response
has begun is re-raised so the server terminates the broken response.

## Same-origin protection

Wrap each state-changing `http.HandlerFunc` that is reachable from a browser:

```go
mux.HandleFunc("POST /settings", sameorigin.Require(updateSettings))
```

`sameorigin.Require` accepts a matching `Origin` header, or a matching
`Referer` header when `Origin` is absent; otherwise it returns 403. It compares
the request scheme and host. For TLS-terminating proxies, it honors
`X-Forwarded-Proto`; configure the proxy to remove client-supplied values before
forwarding requests.

## Base paths

Use `basepath.Builder` to keep redirects and template URLs correct when an
application is mounted below a path:

```go
app := http.NewServeMux()
app.HandleFunc("GET /", home)

handler := basepath.Mount(
	"/app",
	app,
	basepath.TrailingSlashRedirect("/app"),
	http.NotFoundHandler(),
)
```

`Mount` strips `/app` before dispatching to `app`. Pass the same base path to
the renderer so its `url` function and `Redirect` method generate `/app/...`
URLs.

## Templates and static assets

`render.New` parses `render/base.html`, your `templates/chrome.html`, and a
requested page template (`templates/<name>`). The page templates are cached
after their first successful parse. `templates/chrome.html` must define
`home_url`, `navigation`, `navbar_right`, and `app_styles`.

```go
renderer := render.New(templateFS, render.Config{
	BasePath: basepath.New("/app", true),
	Funcs: template.FuncMap{
		"appTitle":         func() string { return "Example" },
		"companyName":      func() string { return "Example GmbH" },
		"brandColor":       func() string { return "#123456" },
		"brandColorHover":  func() string { return "#0d47a1" },
		"changelogTrigger": func() template.HTML { return "" },
		"changelogModal":   func() template.HTML { return "" },
	},
})

func home(w http.ResponseWriter, r *http.Request) {
	renderer.Render(w, "home.html", nil)
}
```

`RequestFuncs` can provide response-specific template functions. They replace
functions of the same name while rendering without mutating the cached template.
The renderer supplies `url`, which turns an application-relative target into a
URL below `BasePath`.

The `template.HTML` values shown above must contain only trusted, application-controlled HTML.

Serve static assets through the same application mux:

```go
app.Handle(
	"GET /static/",
	http.StripPrefix("/static/", render.StaticHandler(appFS)),
)
```

`appFS` must contain a `static` directory. Application files take precedence
over the shared assets embedded below `render/static`.

## License

MIT
