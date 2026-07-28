// Package database provides utility functions for database operations

package database

import (
	"fmt"
	"strings"
)

// parseTablePath parses the table path to extract the actual table name
// Expected format: sandbox/{projectId}/{table}
func parseTablePath(path string) (string, error) {
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return "", fmt.Errorf("invalid table path: %s", path)
	}
	// Return the table name (last part)
	return parts[len(parts)-1], nil
}
