package httpserver

import (
	"log"
	"net/http"
	"time"
)

// WithRequestLoggingAndRecovery logs requests and returns an internal-server-error response for panics before a response starts.
func WithRequestLoggingAndRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &responseRecorder{ResponseWriter: w}
		path := requestPath(r)

		defer func() {
			var panicValue any
			abortResponse := false
			if recovered := recover(); recovered != nil {
				log.Printf("panic method=%s path=%q remote=%q value=%v", r.Method, path, r.RemoteAddr, recovered)
				if !recorder.wroteHeader {
					http.Error(recorder, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				} else {
					panicValue = recovered
					abortResponse = true
				}
			}
			log.Printf(
				"access method=%s path=%q status=%d bytes=%d duration=%s remote=%q",
				r.Method,
				path,
				recorder.Status(),
				recorder.bytesWritten,
				time.Since(start),
				r.RemoteAddr,
			)
			if abortResponse {
				panic(panicValue)
			}
		}()

		next.ServeHTTP(recorder, r)
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status       int
	bytesWritten int
	wroteHeader  bool
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.ResponseWriter.WriteHeader(status)
	if status >= 100 && status <= 199 && status != http.StatusSwitchingProtocols {
		return
	}
	r.status = status
	r.wroteHeader = true
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(data)
	r.bytesWritten += n
	return n, err
}

func (r *responseRecorder) Status() int {
	if r.status == 0 {
		return http.StatusOK
	}
	return r.status
}

func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func requestPath(r *http.Request) string {
	if r.URL == nil {
		return ""
	}
	if path := r.URL.RequestURI(); path != "" {
		return path
	}
	return r.URL.Path
}
