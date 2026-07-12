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
		return nil, fmt.Errorf("error initializing Firebase app: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting Firestore client: %v", err)
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
		return nil, fmt.Errorf("failed to get document: %v", err)
	}

	return docSnap.Data(), nil
}

// GetAll retrieves all documents from the specified collection
func (a *FirestoreAdapter) GetAll(ctx context.Context, collectionPath string) ([]Document, error) {
	iter := a.client.Collection(collectionPath).Documents(ctx)
	var results []Document

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
			return nil, fmt.Errorf("failed to iterate documents: %v", err)
		}
		data := doc.Data()
		data["id"] = doc.Ref.ID
		results = append(results, data)
	}

	return results, nil
}

// Create creates a new document in the specified collection
func (a *FirestoreAdapter) Create(ctx context.Context, collectionPath string, id string, data Document) (string, error) {
	// Use provided ID or generate new one
	if id == "" {
		// Firestore will auto-generate ID if we use Add()
		_, _, err := a.client.Collection(collectionPath).Add(ctx, data)
		if err != nil {
			return "", fmt.Errorf("failed to create document: %v", err)
		}
		// For Firestore, we need to get the generated ID
		// This is a limitation - we'll return empty ID and let caller handle it
		return "", nil
	}

	// Set document with specific ID
	docRef := a.client.Doc(collectionPath + "/" + id)
	if _, err := docRef.Set(ctx, data); err != nil {
		return "", fmt.Errorf("failed to create document: %v", err)
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
		return fmt.Errorf("failed to update document: %v", err)
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
		return fmt.Errorf("failed to delete document: %v", err)
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
		return fmt.Errorf("firestore ping failed: %v", err)
	}
	return nil
}
