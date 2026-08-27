package httpserver

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewServerConfiguresAddressHandlerAndTimeouts(t *testing.T) {
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	server := NewServer("127.0.0.1:8080", handler)

	if server.Addr != "127.0.0.1:8080" || server.Handler == nil {
		t.Fatalf("server = %#v", server)
	}
	if server.ReadHeaderTimeout != 5*time.Second || server.ReadTimeout != 30*time.Second || server.WriteTimeout != 2*time.Minute || server.IdleTimeout != 2*time.Minute {
		t.Fatalf("unexpected server timeouts: %#v", server)
	}
}

func TestWithRequestLoggingAndRecoveryLogsSuccessfulRequest(t *testing.T) {
	logs := captureLogs(t, func() {
		handler := WithRequestLoggingAndRecovery(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("ok"))
		}))
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/test?x=1", nil))
	})

	for _, want := range []string{"access method=GET", `path="/test?x=1"`, "status=200", "bytes=2"} {
		if !strings.Contains(logs, want) {
			t.Fatalf("logs = %q, want %q", logs, want)
		}
	}
}

func TestWithRequestLoggingAndRecoveryLogsExplicitStatus(t *testing.T) {
	logs := captureLogs(t, func() {
		handler := WithRequestLoggingAndRecovery(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}))
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/resource", nil))
	})

	if !strings.Contains(logs, "status=201") {
		t.Fatalf("logs = %q, want explicit status", logs)
	}
}

func TestWithRequestLoggingAndRecoveryRecoversInvalidStatusBeforeResponse(t *testing.T) {
	logs := captureLogs(t, func() {
		handler := WithRequestLoggingAndRecovery(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(99)
		}))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/invalid-status", nil))
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
		}
	})

	if !strings.Contains(logs, "panic method=GET") || !strings.Contains(logs, "status=500") {
		t.Fatalf("logs = %q", logs)
	}
}

func TestWithRequestLoggingAndRecoveryPreservesFinalStatusAfterInformationalHeader(t *testing.T) {
	logs := captureLogs(t, func() {
		writer := &informationalResponseWriter{}
		handler := WithRequestLoggingAndRecovery(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusEarlyHints)
			w.WriteHeader(http.StatusNoContent)
		}))
		handler.ServeHTTP(writer, httptest.NewRequest(http.MethodGet, "/test", nil))
	})

	if !strings.Contains(logs, "status=204") {
		t.Fatalf("logs = %q, want final status", logs)
	}
}

func TestWithRequestLoggingAndRecoveryRecoversPanicBeforeResponse(t *testing.T) {
	logs := captureLogs(t, func() {
		handler := WithRequestLoggingAndRecovery(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic("boom")
		}))
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/test", nil))
		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
		}
	})

	if !strings.Contains(logs, "panic method=GET") || !strings.Contains(logs, "status=500") {
		t.Fatalf("logs = %q", logs)
	}
}

func TestWithRequestLoggingAndRecoveryRepanicsAfterResponseStarts(t *testing.T) {
	handler := WithRequestLoggingAndRecovery(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		panic("boom")
	}))

	defer func() {
		if recovered := recover(); recovered != "boom" {
			t.Fatalf("panic = %v, want boom", recovered)
		}
	}()
	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/test", nil))
}

type informationalResponseWriter struct {
	headers  http.Header
	statuses []int
}

func (w *informationalResponseWriter) Header() http.Header {
	if w.headers == nil {
		w.headers = make(http.Header)
	}
	return w.headers
}

func (w *informationalResponseWriter) WriteHeader(status int) {
	w.statuses = append(w.statuses, status)
}

func (w *informationalResponseWriter) Write(data []byte) (int, error) { return len(data), nil }

func captureLogs(t *testing.T, fn func()) string {
	t.Helper()
	var buffer bytes.Buffer
	originalWriter := log.Writer()
	originalFlags := log.Flags()
	log.SetOutput(&buffer)
	log.SetFlags(0)
	t.Cleanup(func() {
		log.SetOutput(originalWriter)
		log.SetFlags(originalFlags)
	})
	fn()
	return buffer.String()
}
