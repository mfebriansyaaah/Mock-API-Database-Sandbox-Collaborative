package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/internal/logger"
)

// fakeSubmitter is a thread-safe in-memory logger.Submitter used to verify
// what the middleware hands off without spinning up a real *logger.Logger
// (and therefore a real Firestore client).
type fakeSubmitter struct {
	mu      sync.Mutex
	entries []*logger.LogEntry
	// allow toggles the return value of Submit. When false, the entry is
	// not appended — this simulates a full internal channel.
	allow bool
}

func (f *fakeSubmitter) Submit(entry *logger.LogEntry) bool {
	if !f.allow {
		return false
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries = append(f.entries, entry)
	return true
}

func (f *fakeSubmitter) snapshot() []*logger.LogEntry {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]*logger.LogEntry, len(f.entries))
	copy(out, f.entries)
	return out
}

// okHandler is a trivial downstream handler used to make the middleware
// advance past next.ServeHTTP.
var okHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
})

// failingHandler is a downstream handler that writes a non-2xx status, so we
// can verify the middleware records the actual response status (not just 200).
var failingHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "boom", http.StatusInternalServerError)
})

// panicHandler is a downstream handler that panics, so we can verify the
// middleware still emits a log entry (latency must be measured in defer).
var panicHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	panic("kaboom")
})

// newReq builds a minimal GET request against an in-memory httptest server.
func newReq(t *testing.T, method, target string) *http.Request {
	t.Helper()
	r := httptest.NewRequest(method, target, nil)
	return r
}

func TestLogging_RecordsMethodPathAndProjectID(t *testing.T) {
	sub := &fakeSubmitter{allow: true}
	handler := Logging(sub, "proj-abc")(okHandler)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, newReq(t, http.MethodGet, "/hello"))

	entries := sub.snapshot()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]
	if e.ProjectID != "proj-abc" {
		t.Errorf("ProjectID = %q, want %q", e.ProjectID, "proj-abc")
	}
	if e.Method != http.MethodGet {
		t.Errorf("Method = %q, want %q", e.Method, http.MethodGet)
	}
	if e.Path != "/hello" {
		t.Errorf("Path = %q, want %q", e.Path, "/hello")
	}
}

func TestLogging_RecordsNonDefaultStatus(t *testing.T) {
	sub := &fakeSubmitter{allow: true}
	handler := Logging(sub, "proj-xyz")(failingHandler)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, newReq(t, http.MethodPost, "/explode"))

	entries := sub.snapshot()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if got, want := entries[0].Status, http.StatusInternalServerError; got != want {
		t.Errorf("Status = %d, want %d", got, want)
	}
	if got, want := entries[0].Method, http.MethodPost; got != want {
		t.Errorf("Method = %q, want %q", got, want)
	}
}

func TestLogging_RecordsLatencyAndTimestamp(t *testing.T) {
	sub := &fakeSubmitter{allow: true}
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(15 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})
	handler := Logging(sub, "proj-time")(slow)

	before := time.Now()
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, newReq(t, http.MethodGet, "/slow"))
	after := time.Now()

	entries := sub.snapshot()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	e := entries[0]

	// Timestamp should be a real time inside the request window.
	if e.Timestamp.Before(before) || e.Timestamp.After(after) {
		t.Errorf("Timestamp %v not in [%v, %v]", e.Timestamp, before, after)
	}
	// Latency should be at least the sleep duration; we don't assert an
	// upper bound to keep the test robust under CI load.
	if e.Latency < 10*time.Millisecond {
		t.Errorf("Latency = %v, want >= 10ms", e.Latency)
	}
}

func TestLogging_DropsEntryWhenSubmitterReturnsFalse(t *testing.T) {
	// allow=false simulates an internal channel that is full; the middleware
	// must not panic, must not call into a nil submitter, and must still
	// complete the response successfully.
	sub := &fakeSubmitter{allow: false}
	handler := Logging(sub, "proj-full")(okHandler)

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, newReq(t, http.MethodGet, "/drop"))

	if got := rr.Code; got != http.StatusOK {
		t.Errorf("response status = %d, want %d", got, http.StatusOK)
	}
	if got := rr.Body.String(); got != "ok" {
		t.Errorf("response body = %q, want %q", got, "ok")
	}
	if entries := sub.snapshot(); len(entries) != 0 {
		t.Errorf("expected 0 entries (dropped), got %d", len(entries))
	}
}

func TestLogging_EmitsEntryOnDownstreamPanic(t *testing.T) {
	// The latency measurement lives in a defer, so a panic in the handler
	// must still result in exactly one log entry being submitted. We recover
	// here only to keep the test from crashing the runner; the real value of
	// this test is verifying the entry was emitted.
	sub := &fakeSubmitter{allow: true}
	handler := Logging(sub, "proj-panic")(panicHandler)

	rr := httptest.NewRecorder()

	defer func() {
		// Swallow the panic once we've verified the entry.
		_ = recover()
		entries := sub.snapshot()
		if len(entries) != 1 {
			t.Fatalf("expected 1 entry after panic, got %d", len(entries))
		}
		if entries[0].Path != "/panic" {
			t.Errorf("Path = %q, want %q", entries[0].Path, "/panic")
		}
	}()

	handler.ServeHTTP(rr, newReq(t, http.MethodGet, "/panic"))
}

func TestLogging_OneEntryPerRequest(t *testing.T) {
	sub := &fakeSubmitter{allow: true}
	handler := Logging(sub, "proj-count")(okHandler)

	for i := 0; i < 5; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, newReq(t, http.MethodGet, "/n"))
	}

	if got, want := len(sub.snapshot()), 5; got != want {
		t.Errorf("entries = %d, want %d", got, want)
	}
}

// Compile-time check: *logger.Logger must satisfy logger.Submitter.
// If the interface drifts (e.g. signature change in Submit), this will
// fail to compile, alerting us before runtime.
var _ logger.Submitter = (*logger.Logger)(nil)
