# go-web

`go-web` provides small HTTP server helpers: a standard server configuration,
request logging and panic recovery middleware, Windows service support, and
same-origin protection for state-changing routes.

## Installation

```sh
go get github.com/realholgi/go-web
```

## Usage

```go
package main

import (
	"net/http"

	"github.com/realholgi/go-web/httpserver"
	"github.com/realholgi/go-web/sameorigin"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /settings", sameorigin.Require(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	handler := httpserver.WithRequestLoggingAndRecovery(mux)
	httpserver.Run(":8080", handler, "my-service")
}
```

`httpserver.Run` starts the server directly on non-Windows platforms. On
Windows, it detects the service manager and uses `my-service` as the service
name when applicable.

`sameorigin.Require` requires a matching `Origin` or `Referer` header. It
honors `X-Forwarded-Proto` for TLS-terminating proxies; only trust that header
when the proxy removes values supplied by clients.

## License

MIT
