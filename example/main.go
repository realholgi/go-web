package main

import (
	"fmt"
	"net/http"

	"github.com/realholgi/go-web/httpserver"
	"github.com/realholgi/go-web/sameorigin"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", home)
	mux.HandleFunc("POST /settings", sameorigin.Require(updateSettings))

	handler := httpserver.WithRequestLoggingAndRecovery(mux)
	httpserver.Run(":8080", handler, "go-web-example")
}

func home(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = fmt.Fprintln(w, "go-web example is running")
}

func updateSettings(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
