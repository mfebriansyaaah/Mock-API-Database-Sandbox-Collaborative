package database

import (
	"context"
	"testing"
)

// MockDatabaseClient is a mock implementation of DatabaseClient for testing
func TestDatabaseInterface(t *testing.T) {
	// Test that all database adapters implement the DatabaseClient interface
	// This is a compile-time test

	// Test Firestore adapter
	firestoreConfig := &FirestoreConfig{
		ProjectID: "test-project",
	}
	firestoreClient, err := NewFirestoreAdapter(context.Background(), firestoreConfig)
	if err != nil {
		t.Logf("Firestore adapter creation failed (expected in test environment): %v", err)
		// This is expected to fail in test environment without Firebase credentials
	} else {
		defer firestoreClient.Close()
		// Verify it implements DatabaseClient
		var _ DatabaseClient = firestoreClient
		t.Log("✅ Firestore adapter implements DatabaseClient interface")
	}

	// Test PostgreSQL adapter (will fail without actual PostgreSQL server)
	postgresConfig := &PostgreSQLConfig{
		Host:     "localhost",
		Port:     "5432",
		User:     "test",
		Password: "test",
		Database: "test",
		SSLMode:  "disable",
	}
	postgresClient, err := NewPostgreSQLClient(context.Background(), postgresConfig)
	if err != nil {
		t.Logf("PostgreSQL adapter creation failed (expected without server): %v", err)
		// This is expected to fail without a running PostgreSQL server
	} else {
		defer postgresClient.Close()
		// Verify it implements DatabaseClient
		var _ DatabaseClient = postgresClient
		t.Log("✅ PostgreSQL adapter implements DatabaseClient interface")
	}

	// Test MySQL adapter (will fail without actual MySQL server)
	mysqlConfig := &MySQLConfig{
		Host:      "localhost",
		Port:      "3306",
		User:      "test",
		Password:  "test",
		Database:  "test",
		ParseTime: true,
	}
	mysqlClient, err := NewMySQLClient(context.Background(), mysqlConfig)
	if err != nil {
		t.Logf("MySQL adapter creation failed (expected without server): %v", err)
		// This is expected to fail without a running MySQL server
	} else {
		defer mysqlClient.Close()
		// Verify it implements DatabaseClient
		var _ DatabaseClient = mysqlClient
		t.Log("✅ MySQL adapter implements DatabaseClient interface")
	}
}

// TestNewDatabaseClient tests the factory function
func TestNewDatabaseClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *DatabaseConfig
		wantErr bool
	}{
		{
			name: "firestore config",
			config: &DatabaseConfig{
				Type: DatabaseTypeFirestore,
				FirestoreConfig: &FirestoreConfig{
					ProjectID: "test-project",
				},
			},
			wantErr: false, // May fail without credentials, but shouldn't panic
		},
		{
			name: "postgresql config",
			config: &DatabaseConfig{
				Type: DatabaseTypePostgreSQL,
				PostgreSQLConfig: &PostgreSQLConfig{
					Host:     "localhost",
					Port:     "5432",
					User:     "test",
					Password: "test",
					Database: "test",
					SSLMode:  "disable",
				},
			},
			wantErr: false, // May fail without server, but shouldn't panic
		},
		{
			name: "mysql config",
			config: &DatabaseConfig{
				Type: DatabaseTypeMySQL,
				MySQLConfig: &MySQLConfig{
					Host:      "localhost",
					Port:      "3306",
					User:      "test",
					Password:  "test",
					Database:  "test",
					ParseTime: true,
				},
			},
			wantErr: false, // May fail without server, but shouldn't panic
		},
		{
			name: "unsupported database type",
			config: &DatabaseConfig{
				Type: DatabaseType("unsupported"),
			},
			wantErr: true,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewDatabaseClient(context.Background(), tt.config)
			if tt.wantErr && err == nil {
				t.Errorf("NewDatabaseClient() expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Logf("NewDatabaseClient() got expected error (no server): %v", err)
			}
			// Safe to call Close - all database clients now check for nil internally
			if client != nil {
				client.Close()
			}
		})
	}
}

// TestParseTablePath tests the parseTablePath utility function
func TestParseTablePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "valid firestore path",
			path:    "sandbox/project123/users",
			want:    "users",
			wantErr: false,
		},
		{
			name:    "valid path with more parts",
			path:    "sandbox/project123/collections/users",
			want:    "users",
			wantErr: false,
		},
		{
			name:    "invalid path - too short",
			path:    "sandbox",
			want:    "",
			wantErr: true,
		},
		{
			name:    "invalid path - empty",
			path:    "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTablePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTablePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseTablePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
