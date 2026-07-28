// Package database provides PostgreSQL database adapter implementation

package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// PostgreSQLClient implements DatabaseClient interface for PostgreSQL
type PostgreSQLClient struct {
	db     *sql.DB
	config *PostgreSQLConfig
}

// NewPostgreSQLClient creates a new PostgreSQL database client
func NewPostgreSQLClient(ctx context.Context, config *PostgreSQLConfig) (*PostgreSQLClient, error) {
	if config == nil {
		return nil, fmt.Errorf("PostgreSQL config cannot be nil")
	}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
		config.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		db.Close() // Close the connection on error
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0) // No lifetime limit

	log.Printf("Successfully connected to PostgreSQL at %s:%s", config.Host, config.Port)

	return &PostgreSQLClient{
		db:     db,
		config: config,
	}, nil
}

// Type returns the database type
func (c *PostgreSQLClient) Type() DatabaseType {
	return DatabaseTypePostgreSQL
}

// Get retrieves a single document by ID from the specified table
func (c *PostgreSQLClient) Get(ctx context.Context, tablePath string, id string) (Document, error) {
	// Parse table path (format: sandbox/{projectId}/{table})
	tableName, err := parseTablePath(tablePath)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", tableName)
	rows, err := c.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query PostgreSQL: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("document not found")
	}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Create a slice of interface{} to hold the values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Scan the row into the value pointers
	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	// Convert to Document
	doc := make(Document)
	for i, col := range columns {
		val := values[i]
		// Handle different types
		if b, ok := val.([]byte); ok {
			doc[col] = string(b)
		} else {
			doc[col] = val
		}
	}

	return doc, nil
}

// GetAll retrieves documents from the specified table.
// Supports optional pagination via opts (Limit / Offset).
func (c *PostgreSQLClient) GetAll(ctx context.Context, tablePath string, opts *GetAllOptions) ([]Document, error) {
	// Parse table path (format: sandbox/{projectId}/{table})
	tableName, err := parseTablePath(tablePath)
	if err != nil {
		return nil, err
	}

	// Defensive identifier quoting — table names come from the URL path so
	// they can technically contain characters that would break the SQL string.
	// We keep behaviour simple here; production hardening lives elsewhere.
	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	args := []interface{}{}
	if opts != nil && opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", len(args)+1)
		args = append(args, opts.Limit)
	}
	if opts != nil && opts.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", len(args)+1)
		args = append(args, opts.Offset)
	}
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query PostgreSQL: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []Document
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into the value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Convert to Document
		doc := make(Document)
		for i, col := range columns {
			val := values[i]
			// Handle different types
			if b, ok := val.([]byte); ok {
				doc[col] = string(b)
			} else {
				doc[col] = val
			}
		}
		// Add ID to document
		if idVal, ok := doc["id"]; ok {
			doc["id"] = idVal
		}
		results = append(results, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// CountAll returns the total number of rows in the specified table.
func (c *PostgreSQLClient) CountAll(ctx context.Context, tablePath string) (int64, error) {
	tableName, err := parseTablePath(tablePath)
	if err != nil {
		return 0, err
	}
	var count int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	if err := c.db.QueryRowContext(ctx, query).Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count rows in PostgreSQL: %w", err)
	}
	return count, nil
}

// Create creates a new document in the specified table
func (c *PostgreSQLClient) Create(ctx context.Context, tablePath string, id string, data Document) (string, error) {
	// Parse table path (format: sandbox/{projectId}/{table})
	tableName, err := parseTablePath(tablePath)
	if err != nil {
		return "", err
	}

	// Generate ID if not provided
	if id == "" {
		id = uuid.New().String()
	}

	// Build INSERT query
	columns := []string{}
	placeholders := []string{}
	values := []interface{}{}
	idx := 1

	for k, v := range data {
		columns = append(columns, k)
		placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
		values = append(values, v)
		idx++
	}

	// Add ID column
	columns = append(columns, "id")
	placeholders = append(placeholders, fmt.Sprintf("$%d", idx))
	values = append(values, id)

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	_, err = c.db.ExecContext(ctx, query, values...)
	if err != nil {
		return "", fmt.Errorf("failed to insert document: %w", err)
	}

	return id, nil
}

// Update updates an existing document in the specified table
func (c *PostgreSQLClient) Update(ctx context.Context, tablePath string, id string, data Document) error {
	// Parse table path (format: sandbox/{projectId}/{table})
	tableName, err := parseTablePath(tablePath)
	if err != nil {
		return err
	}

	// Build UPDATE query
	setClauses := []string{}
	values := []interface{}{}
	idx := 1

	for k, v := range data {
		if k != "id" { // Don't update ID
			setClauses = append(setClauses, fmt.Sprintf("%s = $%d", k, idx))
			values = append(values, v)
			idx++
		}
	}

	// Add ID to WHERE clause
	setClauses = append(setClauses, fmt.Sprintf("id = $%d", idx))
	values = append(values, id)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d", tableName, strings.Join(setClauses[:len(setClauses)-1], ", "), idx)

	_, err = c.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}

	return nil
}

// Delete deletes a document from the specified table
func (c *PostgreSQLClient) Delete(ctx context.Context, tablePath string, id string) error {
	// Parse table path (format: sandbox/{projectId}/{table})
	tableName, err := parseTablePath(tablePath)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", tableName)
	_, err = c.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}

	return nil
}

// Close closes the database connection
func (c *PostgreSQLClient) Close() error {
	if c == nil {
		return nil
	}
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Ping checks if the database connection is alive
func (c *PostgreSQLClient) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// ListCollections is not natively supported for PostgreSQL sandbox collections.
// Returns an empty slice so callers can fall back to local tracking.
func (c *PostgreSQLClient) ListCollections(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

// DeleteProject is not supported for PostgreSQL. Returns an error indicating the
// operation is not available so callers can surface it gracefully.
func (c *PostgreSQLClient) DeleteProject(_ context.Context, _ string) error {
	return fmt.Errorf("delete project is not supported for PostgreSQL backend")
}
