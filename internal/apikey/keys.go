// Package apikey manages API keys for external sandbox access.
// Keys are stored as documents via the existing DatabaseClient interface
// in the collection path "__api_keys/{projectId}".
package apikey

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/internal/database"
)

const (
	KeyPrefix     = "msbx_"
	KeyLength     = 32
	collectionFmt = "_api_keys/%s/keys"
)

// APIKey represents a stored API key.
type APIKey struct {
	ID         string    `json:"id"`
	ProjectID  string    `json:"projectId"`
	Key        string    `json:"key"`
	Name       string    `json:"name"`
	Scopes     []string  `json:"scopes"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"createdAt"`
	LastUsedAt time.Time `json:"lastUsedAt,omitempty"`
}

// Manager handles CRUD operations for API keys.
type Manager struct {
	db database.DatabaseClient
}

// NewManager creates a new API key manager.
func NewManager(db database.DatabaseClient) *Manager {
	return &Manager{db: db}
}

// Generate creates a new API key for the given project.
func (m *Manager) Generate(ctx context.Context, projectID, name string) (*APIKey, error) {
	if projectID == "" {
		return nil, fmt.Errorf("projectId is required")
	}

	raw := make([]byte, KeyLength)
	if _, err := rand.Read(raw); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	keyValue := KeyPrefix + hex.EncodeToString(raw)

	key := &APIKey{
		ID:        hex.EncodeToString(raw[:8]),
		ProjectID: projectID,
		Key:       keyValue,
		Name:      name,
		Scopes:    []string{"read", "write"},
		Active:    true,
		CreatedAt: time.Now().UTC(),
	}

	doc := toDocument(key)
	collection := fmt.Sprintf(collectionFmt, projectID)
	if _, err := m.db.Create(ctx, collection, key.ID, doc); err != nil {
		return nil, fmt.Errorf("failed to store API key: %w", err)
	}

	return key, nil
}

// Validate checks whether the given key string is valid and active.
// Returns the associated APIKey if valid, nil otherwise.
func (m *Manager) Validate(ctx context.Context, key string) (*APIKey, error) {
	if !strings.HasPrefix(key, KeyPrefix) {
		return nil, fmt.Errorf("invalid key format")
	}

	// We need to search all projects for this key.
	// For efficiency, we extract the key ID from the key itself.
	// Key format: msbx_<32hex chars> — we use first 16 chars as the document ID.
	keyID := extractKeyID(key)
	if keyID == "" {
		return nil, fmt.Errorf("cannot extract key ID")
	}

	// We don't know which project this key belongs to, so we need to search.
	// This is a limitation of the current DatabaseClient interface.
	// For now, we accept a projectId hint.
	return nil, fmt.Errorf("use ValidateForProject instead")
}

// ValidateForProject checks whether the key belongs to the given project.
func (m *Manager) ValidateForProject(ctx context.Context, projectID, key string) (*APIKey, error) {
	if !strings.HasPrefix(key, KeyPrefix) {
		return nil, fmt.Errorf("invalid key format")
	}

	keyID := extractKeyID(key)
	if keyID == "" {
		return nil, fmt.Errorf("cannot extract key ID")
	}

	collection := fmt.Sprintf(collectionFmt, projectID)
	doc, err := m.db.Get(ctx, collection, keyID)
	if err != nil {
		return nil, fmt.Errorf("key not found")
	}

	apiKey := fromDocument(doc)
	if apiKey == nil || !apiKey.Active {
		return nil, fmt.Errorf("key is inactive")
	}

	if apiKey.Key != key {
		return nil, fmt.Errorf("key mismatch")
	}

	return apiKey, nil
}

// List returns all API keys for a project.
func (m *Manager) List(ctx context.Context, projectID string) ([]*APIKey, error) {
	collection := fmt.Sprintf(collectionFmt, projectID)
	docs, err := m.db.GetAll(ctx, collection, &database.GetAllOptions{Limit: 100, Offset: 0})
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	keys := make([]*APIKey, 0, len(docs))
	for _, doc := range docs {
		if k := fromDocument(doc); k != nil {
			keys = append(keys, k)
		}
	}
	return keys, nil
}

// Revoke deactivates an API key.
func (m *Manager) Revoke(ctx context.Context, projectID, keyID string) error {
	collection := fmt.Sprintf(collectionFmt, projectID)
	doc, err := m.db.Get(ctx, collection, keyID)
	if err != nil {
		return fmt.Errorf("key not found")
	}

	apiKey := fromDocument(doc)
	if apiKey == nil {
		return fmt.Errorf("invalid key document")
	}

	apiKey.Active = false
	updated := toDocument(apiKey)
	return m.db.Update(ctx, collection, keyID, updated)
}

// Delete permanently removes an API key.
func (m *Manager) Delete(ctx context.Context, projectID, keyID string) error {
	collection := fmt.Sprintf(collectionFmt, projectID)
	return m.db.Delete(ctx, collection, keyID)
}

// ListAllProjects returns all project IDs that have API keys.
func (m *Manager) ListAllProjects(ctx context.Context) ([]string, error) {
	// This is a limitation — we can't list collections in the DatabaseClient interface.
	// The frontend will track which projects have keys.
	return nil, nil
}

func extractKeyID(key string) string {
	// Key format: msbx_<keyID hex>
	withoutPrefix := strings.TrimPrefix(key, KeyPrefix)
	if len(withoutPrefix) < 16 {
		return ""
	}
	return withoutPrefix[:16]
}

func toDocument(k *APIKey) database.Document {
	doc := database.Document{
		"id":        k.ID,
		"projectId": k.ProjectID,
		"key":       k.Key,
		"name":      k.Name,
		"active":    k.Active,
		"createdAt": k.CreatedAt.Format(time.RFC3339),
	}
	if !k.LastUsedAt.IsZero() {
		doc["lastUsedAt"] = k.LastUsedAt.Format(time.RFC3339)
	}
	if k.Scopes != nil {
		doc["scopes"] = strings.Join(k.Scopes, ",")
	}
	return doc
}

func fromDocument(doc database.Document) *APIKey {
	if doc == nil {
		return nil
	}

	id, _ := doc["id"].(string)
	projectID, _ := doc["projectId"].(string)
	key, _ := doc["key"].(string)
	name, _ := doc["name"].(string)
	active, _ := doc["active"].(bool)

	k := &APIKey{
		ID:        id,
		ProjectID: projectID,
		Key:       key,
		Name:      name,
		Active:    active,
	}

	if createdAt, ok := doc["createdAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			k.CreatedAt = t
		}
	}
	if lastUsedAt, ok := doc["lastUsedAt"].(string); ok {
		if t, err := time.Parse(time.RFC3339, lastUsedAt); err == nil {
			k.LastUsedAt = t
		}
	}
	if scopesStr, ok := doc["scopes"].(string); ok && scopesStr != "" {
		k.Scopes = strings.Split(scopesStr, ",")
	}

	return k
}
