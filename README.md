# go-web

`go-web` provides small HTTP server helpers: a standard server configuration,
request logging and panic recovery middleware, Windows service support, and
same-origin protection for state-changing routes.

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

## License

MIT
