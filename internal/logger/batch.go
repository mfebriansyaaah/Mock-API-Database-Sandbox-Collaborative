package logger

import (
	"context"

	"cloud.google.com/go/firestore"
)

// commitBatchedDeletes commits the given document iterator's snapshots as
// deletes in batches of at most `batchSize` (Firestore caps writes at 500).
// Empty iterators commit no batches.
func commitBatchedDeletes(ctx context.Context, client *firestore.Client, iter *firestore.DocumentIterator, batchSize int) error {
	if batchSize <= 0 || batchSize > 500 {
		batchSize = 500
	}

	batch := client.Batch()
	pending := 0
	for {
		snap, err := iter.Next()
		if err != nil {
			break // iterator.Done
		}
		batch.Delete(snap.Ref)
		pending++
		if pending >= batchSize {
			if _, err := batch.Commit(ctx); err != nil {
				return err
			}
			batch = client.Batch()
			pending = 0
		}
	}
	if pending > 0 {
		if _, err := batch.Commit(ctx); err != nil {
			return err
		}
	}
	return nil
}
