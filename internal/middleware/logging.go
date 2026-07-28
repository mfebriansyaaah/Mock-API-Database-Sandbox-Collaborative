// Package middleware contains HTTP middlewares used by the server.
package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/internal/logger"
)

// statusRecorder wraps http.ResponseWriter to capture the response status code
// without consuming or copying the body.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Logging returns an HTTP middleware that records every request to the given
// logger. The middleware is fully non-blocking: it only builds the LogEntry
// and hands it off via the Submitter interface (which drops on a full channel).
//
// Latency is measured in a deferred function so it captures panics in the
// downstream handler as well. The LogEntry's ProjectID is taken from the
// caller's projectID so every entry lands in the right Firestore
// subcollection.
//
// The Submitter is taken as an interface so tests can inject a fake without
// standing up a real *logger.Logger (and therefore a real Firestore client).
func Logging(l logger.Submitter, projectID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			defer func() {
				entry := &logger.LogEntry{
					ProjectID: projectID,
					Method:    r.Method,
					Path:      r.URL.Path,
					Status:    rec.status,
					Latency:   time.Since(start),
					Timestamp: start,
				}

				if !l.Submit(entry) {
					log.Printf("middleware: log channel full, dropping entry for %s %s", r.Method, r.URL.Path)
				}
			}()

			next.ServeHTTP(rec, r)
		})
	}
}
