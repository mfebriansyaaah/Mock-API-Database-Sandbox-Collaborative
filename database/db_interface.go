// Package database provides interfaces and implementations for database operations.
// This file defines the common interface that all database adapters must implement.

package database

import (
	"context"
	"fmt"
)

// DatabaseType represents the type of database being used
type DatabaseType string

const (
	// DatabaseTypeFirestore represents Google Cloud Firestore
	DatabaseTypeFirestore DatabaseType = "firestore"
	// DatabaseTypePostgreSQL represents PostgreSQL
	DatabaseTypePostgreSQL DatabaseType = "postgresql"
	// DatabaseTypeMySQL represents MySQL
	DatabaseTypeMySQL DatabaseType = "mysql"
)

// Document represents a generic document/row in the database
type Document map[string]interface{}

// QueryResult represents the result of a query operation
type QueryResult struct {
	Documents []Document
	Error     error
}

// DatabaseClient is the interface that all database adapters must implement
// This provides a common abstraction layer for different database backends
type DatabaseClient interface {
	// Type returns the type of database
	Type() DatabaseType

	// Get retrieves a single document by ID from the specified collection/table
	Get(ctx context.Context, collectionPath string, id string) (Document, error)

	// GetAll retrieves all documents from the specified collection/table
	GetAll(ctx context.Context, collectionPath string) ([]Document, error)

	// Create creates a new document in the specified collection/table
	// If id is empty, the database should generate a new ID
	Create(ctx context.Context, collectionPath string, id string, data Document) (string, error)

	// Update updates an existing document in the specified collection/table
	Update(ctx context.Context, collectionPath string, id string, data Document) error

	// Delete deletes a document from the specified collection/table
	Delete(ctx context.Context, collectionPath string, id string) error

	// Close closes the database connection
	Close() error

	// Ping checks if the database connection is alive
	Ping(ctx context.Context) error
}

// DatabaseConfig holds configuration for database connections
type DatabaseConfig struct {
	// Type of database: "firestore", "postgresql", or "mysql"
	Type DatabaseType

	// Firestore specific configuration
	FirestoreConfig *FirestoreConfig

	// PostgreSQL specific configuration
	PostgreSQLConfig *PostgreSQLConfig

	// MySQL specific configuration
	MySQLConfig *MySQLConfig
}

// FirestoreConfig holds configuration for Firestore
type FirestoreConfig struct {
	ProjectID       string
	CredentialsFile string
	UseEmulator     bool
	EmulatorHost    string
}

// PostgreSQLConfig holds configuration for PostgreSQL
type PostgreSQLConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string // disable, allow, prefer, require
}

// MySQLConfig holds configuration for MySQL
type MySQLConfig struct {
	Host      string
	Port      string
	User      string
	Password  string
	Database  string
	ParseTime bool // Whether to parse time values to time.Time
}

// NewDatabaseClient creates a new DatabaseClient based on the configuration
func NewDatabaseClient(ctx context.Context, config *DatabaseConfig) (DatabaseClient, error) {
	if config == nil {
		return nil, fmt.Errorf("database config cannot be nil")
	}
	switch config.Type {
	case DatabaseTypeFirestore:
		return NewFirestoreAdapter(ctx, config.FirestoreConfig)
	case DatabaseTypePostgreSQL:
		return NewPostgreSQLClient(ctx, config.PostgreSQLConfig)
	case DatabaseTypeMySQL:
		return NewMySQLClient(ctx, config.MySQLConfig)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", config.Type)
	}
}
