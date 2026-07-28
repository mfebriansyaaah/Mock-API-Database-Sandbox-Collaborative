// Package database provides Firestore adapter implementation
// This adapter wraps the existing Firestore client to implement the DatabaseClient interface

package database

import (
	"context"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// FirestoreAdapter implements DatabaseClient interface for Firestore
type FirestoreAdapter struct {
	client    *firestore.Client
	projectID string
}

// NewFirestoreAdapter creates a new Firestore database adapter
func NewFirestoreAdapter(ctx context.Context, config *FirestoreConfig) (*FirestoreAdapter, error) {
	if config == nil {
		config = &FirestoreConfig{
			ProjectID:       "mockapi-sandbox-dev",
			CredentialsFile: "",
		}
	}

	var app *firebase.App
	var err error

	if config.CredentialsFile != "" {
		app, err = firebase.NewApp(ctx, nil, option.WithCredentialsFile(config.CredentialsFile))
	} else {
		app, err = firebase.NewApp(ctx, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase app: %w", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting Firestore client: %w", err)
	}

	return &FirestoreAdapter{
		client:    client,
		projectID: config.ProjectID,
	}, nil
}

// Type returns the database type
func (a *FirestoreAdapter) Type() DatabaseType {
	return DatabaseTypeFirestore
}

// Get retrieves a single document by ID from the specified collection
func (a *FirestoreAdapter) Get(ctx context.Context, collectionPath string, id string) (Document, error) {
	if id == "" {
		return nil, fmt.Errorf("document ID is required")
	}

	docRef := a.client.Doc(collectionPath + "/" + id)
	docSnap, err := docRef.Get(ctx)
	if err != nil {
		// Check if this is a "not found" error from Firestore
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "NotFound") {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return docSnap.Data(), nil
}

// GetAll retrieves documents from the specified collection.
// Supports optional pagination via opts (Limit / Offset).
// We iterate manually so that a hard cap (Limit) is enforced even on backends
// that don't natively support LIMIT in their iterator (Firestore is one of
// them — its .Documents(ctx) call has a 1 MiB response cap that surfaces as
// "rpc error: code = ResourceExhausted desc = Quota exceeded.").
func (a *FirestoreAdapter) GetAll(ctx context.Context, collectionPath string, opts *GetAllOptions) ([]Document, error) {
	iter := a.client.Collection(collectionPath).Documents(ctx)
	var results []Document

	// Skip the first `Offset` documents that come back. Firestore's query
	// language does not support OFFSET directly, so we drop them client-side.
	skipped := 0
	if opts != nil && opts.Offset > 0 {
		skipped = opts.Offset
	}
	limit := -1
	if opts != nil && opts.Limit > 0 {
		limit = opts.Limit
	}

	for {
		doc, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			// Check for "no more items" error which is equivalent to EOF
			if err.Error() == "no more items in iterator" {
				break
			}
			return nil, fmt.Errorf("failed to iterate documents: %w", err)
		}
		if skipped > 0 {
			skipped--
			continue
		}
		data := doc.Data()
		data["id"] = doc.Ref.ID
		results = append(results, data)
		if limit > 0 && len(results) >= limit {
			break
		}
	}

	return results, nil
}

// CountAll returns the total number of documents in a Firestore collection.
// Firestore does not have a native COUNT aggregate in the client SDK, so we
// iterate through all documents and count them.
func (a *FirestoreAdapter) CountAll(ctx context.Context, collectionPath string) (int64, error) {
	iter := a.client.Collection(collectionPath).Documents(ctx)
	var count int64
	for {
		_, err := iter.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			if strings.Contains(err.Error(), "no more items in iterator") {
				break
			}
			return 0, fmt.Errorf("failed to count documents: %w", err)
		}
		count++
	}
	return count, nil
}

// Create creates a new document in the specified collection
func (a *FirestoreAdapter) Create(ctx context.Context, collectionPath string, id string, data Document) (string, error) {
	// Use provided ID or generate new one
	if id == "" {
		// Firestore will auto-generate ID if we use Add()
		_, _, err := a.client.Collection(collectionPath).Add(ctx, data)
		if err != nil {
			return "", fmt.Errorf("failed to create document: %w", err)
		}
		// For Firestore, we need to get the generated ID
		// This is a limitation - we'll return empty ID and let caller handle it
		return "", nil
	}

	// Set document with specific ID
	docRef := a.client.Doc(collectionPath + "/" + id)
	if _, err := docRef.Set(ctx, data); err != nil {
		return "", fmt.Errorf("failed to create document: %w", err)
	}

	return id, nil
}

// Update updates an existing document in the specified collection
func (a *FirestoreAdapter) Update(ctx context.Context, collectionPath string, id string, data Document) error {
	if id == "" {
		return fmt.Errorf("document ID is required for update")
	}

	docRef := a.client.Doc(collectionPath + "/" + id)
	// Use MergeAll to update only the provided fields
	if _, err := docRef.Set(ctx, data, firestore.MergeAll); err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

// Delete deletes a document from the specified collection
func (a *FirestoreAdapter) Delete(ctx context.Context, collectionPath string, id string) error {
	if id == "" {
		return fmt.Errorf("document ID is required for delete")
	}

	docRef := a.client.Doc(collectionPath + "/" + id)
	if _, err := docRef.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// Close closes the database connection
func (a *FirestoreAdapter) Close() error {
	// Firestore client doesn't have an explicit Close method
	// The connection is managed by the Firebase app
	return nil
}

// Ping checks if the database connection is alive
func (a *FirestoreAdapter) Ping(ctx context.Context) error {
	// Try to get a document from a known collection
	// This is a simple way to test the connection
	_, err := a.client.Collection("__ping__").Doc("test").Get(ctx)
	// We expect this to fail (collection doesn't exist), but it tests the connection
	if err != nil && !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "NotFound") {
		return fmt.Errorf("firestore ping failed: %w", err)
	}
	return nil
}

// ListCollections returns the names of all subcollections under the given
// parent document path. For the sandbox this is used to discover tables
// within a project (parentPath = "sandbox/{projectId}").
func (a *FirestoreAdapter) ListCollections(ctx context.Context, parentPath string) ([]string, error) {
	iter := a.client.Doc(parentPath).Collections(ctx)
	var names []string
	for {
		collRef, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list collections: %w", err)
		}
		names = append(names, collRef.ID)
	}
	return names, nil
}

// DeleteProject removes all data associated with a project: sandbox
// collections/documents and API keys, then cleans up the parent documents.
func (a *FirestoreAdapter) DeleteProject(ctx context.Context, projectId string) error {
	// 1. Delete all sandbox collections under sandbox/{projectId}
	sandboxPath := fmt.Sprintf("sandbox/%s", projectId)
	sandboxColls, err := a.ListCollections(ctx, sandboxPath)
	if err != nil {
		// Document may not exist — that's fine
		if !strings.Contains(err.Error(), "not found") &&
			!strings.Contains(err.Error(), "NotFound") &&
			!strings.Contains(err.Error(), "Document") {
			return fmt.Errorf("failed to list sandbox collections: %w", err)
		}
	}
	for _, coll := range sandboxColls {
		if err := a.deleteCollection(ctx, fmt.Sprintf("%s/%s", sandboxPath, coll)); err != nil {
			return fmt.Errorf("failed to delete collection %s/%s: %w", sandboxPath, coll, err)
		}
	}

	// 2. Delete sandbox/{projectId} document
	if _, err := a.client.Doc(sandboxPath).Delete(ctx); err != nil {
		if !strings.Contains(err.Error(), "not found") &&
			!strings.Contains(err.Error(), "NotFound") {
			return fmt.Errorf("failed to delete sandbox document: %w", err)
		}
	}

	// 3. Delete all API keys under _api_keys/{projectId}/keys
	keysCollPath := fmt.Sprintf("_api_keys/%s/keys", projectId)
	if err := a.deleteCollection(ctx, keysCollPath); err != nil {
		return fmt.Errorf("failed to delete API keys: %w", err)
	}

	// 4. Delete _api_keys/{projectId} document
	if _, err := a.client.Doc(fmt.Sprintf("_api_keys/%s", projectId)).Delete(ctx); err != nil {
		if !strings.Contains(err.Error(), "not found") &&
			!strings.Contains(err.Error(), "NotFound") {
			return fmt.Errorf("failed to delete API keys document: %w", err)
		}
	}

	return nil
}

// deleteCollection recursively deletes every document in a Firestore
// collection (including all sub-collections of each document).
func (a *FirestoreAdapter) deleteCollection(ctx context.Context, collectionPath string) error {
	iter := a.client.Collection(collectionPath).DocumentRefs(ctx)
	for {
		docRef, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to iterate document refs: %w", err)
		}

		// Recurse into sub-collections of this document first.
		subIter := docRef.Collections(ctx)
		for {
			subColl, err := subIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to list sub-collections: %v", err)
			}
			subPath := fmt.Sprintf("%s/%s/%s", collectionPath, docRef.ID, subColl.ID)
			if err := a.deleteCollection(ctx, subPath); err != nil {
				return fmt.Errorf("failed to delete sub-collection %s: %v", subColl.ID, err)
			}
		}

		// Now delete the document itself.
		if _, err := docRef.Delete(ctx); err != nil {
			return fmt.Errorf("failed to delete document %s: %v", docRef.ID, err)
		}
	}
	return nil
}
