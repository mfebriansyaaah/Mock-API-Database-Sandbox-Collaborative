package logger

import (
	"context"
	"log"
	"time"

	"cloud.google.com/go/firestore"
)

// runCleanup periodically prunes projects whose log count exceeds the cap.
// It exits when cleanupStop is closed.
func (l *Logger) runCleanup(ctx context.Context) {
	defer l.wg.Done()
	ticker := time.NewTicker(l.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for _, pid := range l.snapshotProjects() {
				if err := l.enforceLogLimit(ctx, pid); err != nil {
					log.Printf("logger: cleanup failed for project %s: %v", pid, err)
				}
			}
		case <-l.cleanupStop:
			return
		case <-ctx.Done():
			return
		}
	}
}

// enforceLogLimit ensures a project has at most MaxLogsPerProject entries.
// Strategy (FIFO):
//  1. Read the MaxLogsPerProject most recent entries (ordered by timestamp desc).
//  2. Compute the timestamp cutoff of the oldest of those entries.
//  3. Delete every entry with timestamp strictly older than the cutoff.
//
// Note: entries with the exact cutoff timestamp are kept (boundary is inclusive
// on the "keep" side). This may occasionally prune one extra entry when many
// logs share the same timestamp, but the cap is never violated.
func (l *Logger) enforceLogLimit(ctx context.Context, projectID string) error {
	col := l.firestore.Collection("projects").Doc(projectID).Collection("logs")

	// Step 1: gather the most recent MaxLogsPerProject documents.
	recentIter := col.
		OrderBy("timestamp", firestore.Desc).
		Limit(l.config.MaxLogsPerProject).
		Documents(ctx)

	var oldestKept time.Time
	count := 0
	for {
		snap, err := recentIter.Next()
		if err != nil {
			break // iterator.Done
		}
		count++
		var entry LogEntry
		if err := snap.DataTo(&entry); err != nil {
			return err
		}
		if oldestKept.IsZero() || entry.Timestamp.Before(oldestKept) {
			oldestKept = entry.Timestamp
		}
	}

	// If we have fewer than the cap, nothing to delete.
	if count < l.config.MaxLogsPerProject {
		return nil
	}

	// Step 2 & 3: delete entries strictly older than oldestKept.
	deleteIter := col.
		Where("timestamp", "<", oldestKept).
		Documents(ctx)

	return commitBatchedDeletes(ctx, l.firestore, deleteIter, 500)
}
