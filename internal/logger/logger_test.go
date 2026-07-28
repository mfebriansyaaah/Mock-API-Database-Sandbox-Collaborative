package logger

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// newTestLogger builds a Logger without a Firestore client. This is safe for
// tests that exercise the in-process channel/metrics only — Submit/Close
// never touch the firestore field directly.
func newTestLogger(channelBuffer, numWorkers int) *Logger {
	cfg := &LoggerConfig{
		ProjectID:         "test-project",
		ChannelBuffer:     channelBuffer,
		NumWorkers:        numWorkers,
		CleanupInterval:   1 << 30, // never tick in tests
		MaxLogsPerProject: 100,
		WriteTimeout:      1 * time.Second,
	}
	l := &Logger{
		config:      cfg,
		logChan:     make(chan *LogEntry, channelBuffer),
		cleanupStop: make(chan struct{}),
		projects:    make(map[string]struct{}),
	}
	return l
}

func TestSubmit_NonBlockingOnFullChannel(t *testing.T) {
	l := newTestLogger(2, 1)
	defer close(l.cleanupStop) // prevent cleanup goroutine leak if any was started

	// Fill the channel to capacity without starting any worker.
	if !l.Submit(&LogEntry{ID: "a"}) {
		t.Fatalf("first submit should succeed (channel empty)")
	}
	if !l.Submit(&LogEntry{ID: "b"}) {
		t.Fatalf("second submit should succeed (channel at capacity)")
	}

	// Channel is now full. Submit must return false without blocking.
	done := make(chan bool, 1)
	go func() {
		done <- l.Submit(&LogEntry{ID: "c"})
	}()

	select {
	case got := <-done:
		if got {
			t.Errorf("Submit on full channel should return false, got true")
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Submit on full channel blocked instead of dropping")
	}
}

func TestSubmit_NilEntryDropped(t *testing.T) {
	l := newTestLogger(1, 1)
	defer close(l.cleanupStop)

	if l.Submit(nil) {
		t.Errorf("Submit(nil) should return false")
	}
}

func TestMarkProjectSeen_ThreadSafe(t *testing.T) {
	l := newTestLogger(1, 1)
	defer close(l.cleanupStop)

	const goroutines = 50
	const iters = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iters; j++ {
				pid := "p-a"
				if (id+j)%2 == 0 {
					pid = "p-b"
				}
				l.markProjectSeen(pid)
			}
		}(i)
	}
	wg.Wait()

	projects := l.snapshotProjects()
	if len(projects) != 2 {
		t.Errorf("expected 2 distinct projects, got %d: %v", len(projects), projects)
	}
}

func TestSnapshotProjects_Empty(t *testing.T) {
	l := newTestLogger(1, 1)
	defer close(l.cleanupStop)

	if got := l.snapshotProjects(); len(got) != 0 {
		t.Errorf("expected empty snapshot, got %v", got)
	}
}

func TestClose_Idempotent(t *testing.T) {
	l := newTestLogger(1, 1)

	// Close should be safe to call multiple times and must not deadlock.
	var counter int32
	done := make(chan struct{})
	go func() {
		l.Close()
		atomic.AddInt32(&counter, 1)
		l.Close() // second call must be a no-op
		atomic.AddInt32(&counter, 1)
		close(done)
	}()

	select {
	case <-done:
		if got := atomic.LoadInt32(&counter); got != 2 {
			t.Errorf("expected 2 close callbacks, got %d", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("Close deadlocked")
	}
}
