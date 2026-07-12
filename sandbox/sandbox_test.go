package sandbox

import (
	"context"
	"fmt"
	"testing"

	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/database"
)

// MockDatabaseClient is a mock implementation of DatabaseClient for testing
type MockDatabaseClient struct {
	data      map[string]map[string]database.Document
	lastID    int
	lastError error
}

func NewMockDatabaseClient() *MockDatabaseClient {
	return &MockDatabaseClient{
		data: make(map[string]map[string]database.Document),
	}
}

func (m *MockDatabaseClient) Type() database.DatabaseType {
	return database.DatabaseTypeMySQL
}

func (m *MockDatabaseClient) Get(ctx context.Context, collectionPath string, id string) (database.Document, error) {
	if m.lastError != nil {
		return nil, m.lastError
	}
	if collection, ok := m.data[collectionPath]; ok {
		if doc, ok := collection[id]; ok {
			return doc, nil
		}
	}
	return nil, fmt.Errorf("document not found")
}

func (m *MockDatabaseClient) GetAll(ctx context.Context, collectionPath string) ([]database.Document, error) {
	if m.lastError != nil {
		return nil, m.lastError
	}
	var results []database.Document
	if collection, ok := m.data[collectionPath]; ok {
		for _, doc := range collection {
			results = append(results, doc)
		}
	}
	return results, nil
}

func (m *MockDatabaseClient) Create(ctx context.Context, collectionPath string, id string, data database.Document) (string, error) {
	if m.lastError != nil {
		return "", m.lastError
	}
	if id == "" {
		m.lastID++
		id = fmt.Sprintf("mock-id-%d", m.lastID)
	}
	if _, ok := m.data[collectionPath]; !ok {
		m.data[collectionPath] = make(map[string]database.Document)
	}
	m.data[collectionPath][id] = data
	return id, nil
}

func (m *MockDatabaseClient) Update(ctx context.Context, collectionPath string, id string, data database.Document) error {
	if m.lastError != nil {
		return m.lastError
	}
	if collection, ok := m.data[collectionPath]; ok {
		if _, ok := collection[id]; ok {
			collection[id] = data
			return nil
		}
	}
	return fmt.Errorf("document not found")
}

func (m *MockDatabaseClient) Delete(ctx context.Context, collectionPath string, id string) error {
	if m.lastError != nil {
		return m.lastError
	}
	if collection, ok := m.data[collectionPath]; ok {
		if _, ok := collection[id]; ok {
			delete(collection, id)
			return nil
		}
	}
	return fmt.Errorf("document not found")
}

func (m *MockDatabaseClient) Close() error {
	return nil
}

func (m *MockDatabaseClient) Ping(ctx context.Context) error {
	return m.lastError
}

func (m *MockDatabaseClient) SetError(err error) {
	m.lastError = err
}

// TestMockDatabaseClient_CRUD tests all CRUD operations on the mock database client
func TestMockDatabaseClient_CRUD(t *testing.T) {
	ctx := context.Background()
	collectionPath := "sandbox/test-project/users"
	mockDB := NewMockDatabaseClient()

	// Test Create
	t.Run("Create", func(t *testing.T) {
		data := database.Document{"name": "Alice", "age": 30}
		id, err := mockDB.Create(ctx, collectionPath, "", data)
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		if id == "" {
			t.Error("Create returned empty ID")
		}
		if len(mockDB.data[collectionPath]) != 1 {
			t.Errorf("Expected 1 document, got %d", len(mockDB.data[collectionPath]))
		}
	})

	// Test Get
	t.Run("Get", func(t *testing.T) {
		doc, err := mockDB.Get(ctx, collectionPath, "mock-id-1")
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}
		if doc["name"] != "Alice" {
			t.Errorf("Get name = %v, want %v", doc["name"], "Alice")
		}
		if doc["age"] != 30 {
			t.Errorf("Get age = %v, want %v", doc["age"], 30)
		}
	})

	// Test GetAll
	t.Run("GetAll", func(t *testing.T) {
		// Add another document
		mockDB.Create(ctx, collectionPath, "mock-id-2", database.Document{"name": "Bob", "age": 25})

		docs, err := mockDB.GetAll(ctx, collectionPath)
		if err != nil {
			t.Fatalf("GetAll failed: %v", err)
		}
		if len(docs) != 2 {
			t.Errorf("GetAll returned %d documents, want 2", len(docs))
		}
	})

	// Test Update
	t.Run("Update", func(t *testing.T) {
		data := database.Document{"name": "Alice Updated", "age": 31}
		err := mockDB.Update(ctx, collectionPath, "mock-id-1", data)
		if err != nil {
			t.Fatalf("Update failed: %v", err)
		}
		// Verify update
		doc, err := mockDB.Get(ctx, collectionPath, "mock-id-1")
		if err != nil {
			t.Fatalf("Get after update failed: %v", err)
		}
		if doc["name"] != "Alice Updated" {
			t.Errorf("Update name = %v, want %v", doc["name"], "Alice Updated")
		}
		if doc["age"] != 31 {
			t.Errorf("Update age = %v, want %v", doc["age"], 31)
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		err := mockDB.Delete(ctx, collectionPath, "mock-id-1")
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}
		// Verify deletion
		_, err = mockDB.Get(ctx, collectionPath, "mock-id-1")
		if err == nil {
			t.Error("Get after delete should fail")
		}
		// Verify only 1 document remains
		docs, err := mockDB.GetAll(ctx, collectionPath)
		if err != nil {
			t.Fatalf("GetAll after delete failed: %v", err)
		}
		if len(docs) != 1 {
			t.Errorf("Expected 1 document after delete, got %d", len(docs))
		}
	})
}

// TestMockDatabaseClient_Errors tests error handling in the mock database client
func TestMockDatabaseClient_Errors(t *testing.T) {
	ctx := context.Background()
	collectionPath := "sandbox/test-project/users"
	mockDB := NewMockDatabaseClient()
	mockDB.SetError(fmt.Errorf("database error"))

	t.Run("Get with error", func(t *testing.T) {
		_, err := mockDB.Get(ctx, collectionPath, "123")
		if err == nil {
			t.Error("Get should return error")
		}
	})

	t.Run("GetAll with error", func(t *testing.T) {
		_, err := mockDB.GetAll(ctx, collectionPath)
		if err == nil {
			t.Error("GetAll should return error")
		}
	})

	t.Run("Create with error", func(t *testing.T) {
		_, err := mockDB.Create(ctx, collectionPath, "", database.Document{})
		if err == nil {
			t.Error("Create should return error")
		}
	})

	t.Run("Update with error", func(t *testing.T) {
		err := mockDB.Update(ctx, collectionPath, "123", database.Document{})
		if err == nil {
			t.Error("Update should return error")
		}
	})

	t.Run("Delete with error", func(t *testing.T) {
		err := mockDB.Delete(ctx, collectionPath, "123")
		if err == nil {
			t.Error("Delete should return error")
		}
	})
}

// TestMockDatabaseClient_NotFound tests not found scenarios
func TestMockDatabaseClient_NotFound(t *testing.T) {
	ctx := context.Background()
	collectionPath := "sandbox/test-project/users"
	mockDB := NewMockDatabaseClient()

	t.Run("Get non-existent document", func(t *testing.T) {
		_, err := mockDB.Get(ctx, collectionPath, "non-existent")
		if err == nil {
			t.Error("Get should return error for non-existent document")
		}
	})

	t.Run("GetAll from empty collection", func(t *testing.T) {
		docs, err := mockDB.GetAll(ctx, collectionPath)
		if err != nil {
			t.Fatalf("GetAll failed: %v", err)
		}
		if len(docs) != 0 {
			t.Errorf("Expected 0 documents, got %d", len(docs))
		}
	})

	t.Run("Update non-existent document", func(t *testing.T) {
		err := mockDB.Update(ctx, collectionPath, "non-existent", database.Document{})
		if err == nil {
			t.Error("Update should return error for non-existent document")
		}
	})

	t.Run("Delete non-existent document", func(t *testing.T) {
		err := mockDB.Delete(ctx, collectionPath, "non-existent")
		if err == nil {
			t.Error("Delete should return error for non-existent document")
		}
	})
}
