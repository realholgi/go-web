//go:build windows

package httpserver

import (
	"context"
	"log"
	"net/http"
	"time"

	"golang.org/x/sys/windows/svc"
)

type service struct {
	server *http.Server
}

func (s *service) Execute(_ []string, requests <-chan svc.ChangeRequest, status chan<- svc.Status) (bool, uint32) {
	status <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server: %v", err)
		}
	}()
	for request := range requests {
		switch request.Cmd {
		case svc.Stop, svc.Shutdown:
			status <- svc.Status{State: svc.StopPending}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.server.Shutdown(ctx); err != nil {
				log.Printf("shutdown: %v", err)
			}
			return false, 0
		}
	}
	return false, 0
}

// Run starts an HTTP server directly or through the Windows service manager.
func Run(addr string, handler http.Handler, serviceName string) {
	server := NewServer(addr, handler)
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("service detection: %v", err)
	}
	if isService {
		if err := svc.Run(serviceName, &service{server: server}); err != nil {
			log.Fatalf("service: %v", err)
		}
		return
	}
	log.Printf("Listening on http://%s", addr)
	log.Fatal(server.ListenAndServe())
}
