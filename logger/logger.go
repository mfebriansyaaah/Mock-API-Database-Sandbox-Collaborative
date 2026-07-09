package logger

import (
	"context"
	"os"
	"sync"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

// Logger manages the asynchronous logging to Firestore.
//
// It owns:
//   - the in-process log channel (bounded, non-blocking submit)
//   - a pool of worker goroutines that drain the channel
//   - a FIFO cleanup ticker that prunes logs beyond MaxLogsPerProject
type Logger struct {
	config    *LoggerConfig
	firestore *firestore.Client
	logChan   chan *LogEntry

	wg          sync.WaitGroup
	closeOnce   sync.Once
	cleanupStop chan struct{}

	mu       sync.RWMutex
	projects map[string]struct{} // project IDs that have been seen
}

// NewLogger constructs a Logger and starts its worker pool and cleanup routine.
// The provided ctx is used for Firestore operations; cancel it on shutdown.
func NewLogger(ctx context.Context, config *LoggerConfig) (*Logger, error) {
	if config == nil {
		config = &LoggerConfig{}
	}
	if config.ProjectID == "" {
		return nil, errMissingProjectID
	}
	config.applyDefaults()

	client, err := newFirestoreClient(ctx, config.ProjectID)
	if err != nil {
		return nil, err
	}

	l := &Logger{
		config:      config,
		firestore:   client,
		logChan:     make(chan *LogEntry, config.ChannelBuffer),
		cleanupStop: make(chan struct{}),
		projects:    make(map[string]struct{}),
	}

	for i := 0; i < config.NumWorkers; i++ {
		l.wg.Add(1)
		go l.runWorker(ctx)
	}

	l.wg.Add(1)
	go l.runCleanup(ctx)

	return l, nil
}

// Submit hands off a log entry to the worker pool.
// The call is non-blocking: if the channel is full, the entry is dropped
// and false is returned. Callers should treat a `false` return as a
// signal that logging is degraded but should NOT fail the request.
func (l *Logger) Submit(entry *LogEntry) bool {
	if entry == nil {
		return false
	}
	select {
	case l.logChan <- entry:
		return true
	default:
		return false
	}
}

// Close stops the worker pool and cleanup routine and closes the Firestore client.
// Safe to call multiple times and safe to call on a Logger with a nil Firestore
// client (useful in tests that exercise only the in-process pipeline).
// Blocks until all workers exit.
func (l *Logger) Close() {
	l.closeOnce.Do(func() {
		close(l.logChan)
		close(l.cleanupStop)
		l.wg.Wait()
		if l.firestore != nil {
			_ = l.firestore.Close()
		}
	})
}

// newFirestoreClient creates a Firestore client, preferring a credentials file
// when GOOGLE_APPLICATION_CREDENTIALS is set, otherwise using Application
// Default Credentials (suitable for Cloud Run / GCE).
func newFirestoreClient(ctx context.Context, projectID string) (*firestore.Client, error) {
	if credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); credPath != "" {
		return firestore.NewClient(ctx, projectID, option.WithCredentialsFile(credPath))
	}
	return firestore.NewClient(ctx, projectID)
}

// markProjectSeen records that a log entry for projectID was processed.
func (l *Logger) markProjectSeen(projectID string) {
	l.mu.Lock()
	l.projects[projectID] = struct{}{}
	l.mu.Unlock()
}

// snapshotProjects returns a slice of currently-tracked project IDs.
func (l *Logger) snapshotProjects() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]string, 0, len(l.projects))
	for pid := range l.projects {
		out = append(out, pid)
	}
	return out
}
