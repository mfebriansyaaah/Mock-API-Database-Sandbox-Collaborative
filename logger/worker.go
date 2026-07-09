package logger

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
)

var errMissingProjectID = errors.New("logger: ProjectID is required")

// runWorker drains logChan and writes entries to Firestore.
// It exits when the channel is closed (by Close).
func (l *Logger) runWorker(ctx context.Context) {
	defer l.wg.Done()
	for entry := range l.logChan {
		if err := l.writeLog(ctx, entry); err != nil {
			log.Printf("logger: failed to write log: %v", err)
			continue
		}
		l.markProjectSeen(entry.ProjectID)
	}
}

// writeLog persists a single entry to /projects/{projectId}/logs/{logId}.
// Each write is bounded by WriteTimeout so a Firestore outage cannot stall
// a worker indefinitely.
func (l *Logger) writeLog(parent context.Context, entry *LogEntry) error {
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	ctx, cancel := context.WithTimeout(parent, l.config.WriteTimeout)
	defer cancel()

	docRef := l.firestore.
		Collection("projects").
		Doc(entry.ProjectID).
		Collection("logs").
		Doc(entry.ID)

	_, err := docRef.Set(ctx, entry)
	return err
}
