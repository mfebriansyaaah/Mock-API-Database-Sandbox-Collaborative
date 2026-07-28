// Package logger provides an asynchronous, non-blocking HTTP request logger
// that writes structured entries to Google Cloud Firestore using a bounded
// worker pool and a FIFO (capped) cleanup strategy.
package logger

import "time"

// Submitter is the minimal contract the middleware depends on for handing off
// a log entry to the underlying pipeline. It is satisfied by *Logger and by
// any test fake, which keeps the middleware decoupled from the concrete
// Firestore-backed implementation.
type Submitter interface {
	// Submit hands off an entry. Implementations MUST be non-blocking:
	// when the downstream buffer is full, the entry is dropped and false
	// is returned. A nil entry must always be treated as a drop (false).
	Submit(entry *LogEntry) bool
}

// LogEntry represents a single HTTP request log entry stored in Firestore.
//
// Firestore path: /projects/{projectId}/logs/{logId}
type LogEntry struct {
	ProjectID string        `firestore:"projectId,omitempty"`
	Method    string        `firestore:"method,omitempty"`
	Path      string        `firestore:"path,omitempty"`
	Status    int           `firestore:"status,omitempty"`
	Payload   string        `firestore:"payload,omitempty"`
	Latency   time.Duration `firestore:"latency,omitempty"`
	Timestamp time.Time     `firestore:"timestamp,omitempty"`
	ID        string        `firestore:"id,omitempty"`
}
