// Package database provides MySQL database adapter implementation

package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

// MySQLClient implements DatabaseClient interface for MySQL
type MySQLClient struct {
	db     *sql.DB
	config *MySQLConfig
}

// NewMySQLClient creates a new MySQL database client
func NewMySQLClient(ctx context.Context, config *MySQLConfig) (*MySQLClient, error) {
	if config == nil {
		return nil, fmt.Errorf("MySQL config cannot be nil")
	}

	// Build connection string
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=%t",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.ParseTime,
	)

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL connection: %v", err)
	}

	// Test connection
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %v", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(0) // No lifetime limit

	log.Printf("Successfully connected to MySQL at %s:%s", config.Host, config.Port)

	return &MySQLClient{
		db:     db,
		config: config,
	}, nil
}

// Type returns the database type
func (c *MySQLClient) Type() DatabaseType {
	return DatabaseTypeMySQL
}

// Get retrieves a single document by ID from the specified table
func (c *MySQLClient) Get(ctx context.Context, tablePath string, id string) (Document, error) {
	// Parse table path (format: sandbox/{projectId}/{table})
	tableName, err := parseTablePath(tablePath)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = ?", tableName)
	rows, err := c.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query MySQL: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("document not found")
	}

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %v", err)
	}

	// Create a slice of interface{} to hold the values
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	// Scan the row into the value pointers
	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf("failed to scan row: %v", err)
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
func (c *MySQLClient) GetAll(ctx context.Context, tablePath string, opts *GetAllOptions) ([]Document, error) {
	// Parse table path (format: sandbox/{projectId}/{table})
	tableName, err := parseTablePath(tablePath)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	args := []interface{}{}
	if opts != nil && opts.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, opts.Limit)
	}
	if opts != nil && opts.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, opts.Offset)
	}
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query MySQL: %v", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %v", err)
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
			return nil, fmt.Errorf("failed to scan row: %v", err)
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
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return results, nil
}

// Create creates a new document in the specified table
func (c *MySQLClient) Create(ctx context.Context, tablePath string, id string, data Document) (string, error) {
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

	for k, v := range data {
		columns = append(columns, k)
		placeholders = append(placeholders, "?")
		values = append(values, v)
	}

	// Add ID column
	columns = append(columns, "id")
	placeholders = append(placeholders, "?")
	values = append(values, id)

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", tableName, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	_, err = c.db.ExecContext(ctx, query, values...)
	if err != nil {
		return "", fmt.Errorf("failed to insert document: %v", err)
	}

	return id, nil
}

// Update updates an existing document in the specified table
func (c *MySQLClient) Update(ctx context.Context, tablePath string, id string, data Document) error {
	// Parse table path (format: sandbox/{projectId}/{table})
	tableName, err := parseTablePath(tablePath)
	if err != nil {
		return err
	}

	// Build UPDATE query
	setClauses := []string{}
	values := []interface{}{}

	for k, v := range data {
		if k != "id" { // Don't update ID
			setClauses = append(setClauses, fmt.Sprintf("%s = ?", k))
			values = append(values, v)
		}
	}

	// Add ID to WHERE clause
	values = append(values, id)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = ?", tableName, strings.Join(setClauses, ", "))

	_, err = c.db.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("failed to update document: %v", err)
	}

	return nil
}

// Delete deletes a document from the specified table
func (c *MySQLClient) Delete(ctx context.Context, tablePath string, id string) error {
	// Parse table path (format: sandbox/{projectId}/{table})
	tableName, err := parseTablePath(tablePath)
	if err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = ?", tableName)
	_, err = c.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete document: %v", err)
	}

	return nil
}

// Close closes the database connection
func (c *MySQLClient) Close() error {
	if c == nil {
		return nil
	}
	if c.db != nil {
		return c.db.Close()
	}
	return nil
}

// Ping checks if the database connection is alive
func (c *MySQLClient) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}
