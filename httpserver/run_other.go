//go:build !windows

package httpserver

import (
	"log"
	"net/http"
)

// Run starts an HTTP server outside the Windows service manager.
func Run(addr string, handler http.Handler, _ string) {
	server := NewServer(addr, handler)
	log.Printf("Listening on http://%s", addr)
	log.Fatal(server.ListenAndServe())
}
