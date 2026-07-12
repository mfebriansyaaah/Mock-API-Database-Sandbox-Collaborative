package sandbox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/database"
)

// SandboxHandler handles dynamic sandbox endpoints
type SandboxHandler struct {
	client    database.DatabaseClient
	projectID string
}

// NewSandboxHandler creates a new SandboxHandler
type NewSandboxHandlerFunc func(client database.DatabaseClient, projectID string) *SandboxHandler

func NewSandboxHandler(client database.DatabaseClient, projectID string) *SandboxHandler {
	return &SandboxHandler{
		client:    client,
		projectID: projectID,
	}
}

// HandleRequest handles dynamic sandbox requests
func (h *SandboxHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Extract projectId and table from path
	projectId := r.PathValue("projectId")
	table := r.PathValue("table")
	id := r.PathValue("id")

	if projectId == "" || table == "" {
		http.Error(w, "projectId and table are required", http.StatusBadRequest)
		return
	}

	// Use project-specific collection path
	collectionPath := fmt.Sprintf("sandbox/%s/%s", projectId, table)

	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r, collectionPath, id)
	case http.MethodPost:
		h.handlePost(w, r, collectionPath, id)
	case http.MethodDelete:
		h.handleDelete(w, r, collectionPath, id)
	case http.MethodPut, http.MethodPatch:
		h.handleUpdate(w, r, collectionPath, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *SandboxHandler) handleGet(w http.ResponseWriter, r *http.Request, collectionPath string, id string) {
	ctx := r.Context()

	if id != "" {
		// Get single document
		doc, err := h.client.Get(ctx, collectionPath, id)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				http.Error(w, "Document not found", http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to get document: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(doc); err != nil {
			http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
			return
		}
		return
	}

	// Get all documents in collection
	docs, err := h.client.GetAll(ctx, collectionPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get documents: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if docs == nil {
		docs = []database.Document{}
	}
	if err := json.NewEncoder(w).Encode(docs); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (h *SandboxHandler) handlePost(w http.ResponseWriter, r *http.Request, collectionPath string, id string) {
	ctx := r.Context()

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	// Parse JSON
	var data database.Document
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Add metadata
	if data == nil {
		data = make(database.Document)
	}
	data["_createdAt"] = "server_timestamp" // Will be handled by database adapter
	data["_createdBy"] = "anonymous"        // TODO: Add auth

	// Use provided ID or generate new one
	if id == "" {
		id = uuid.New().String()
	}

	// Create document
	createdID, err := h.client.Create(ctx, collectionPath, id, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create document: %v", err), http.StatusInternalServerError)
		return
	}

	// If ID was auto-generated, use the returned ID
	if id == "" {
		id = createdID
	}

	// Return created document
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := map[string]interface{}{
		"id":       id,
		"message":  "Document created successfully",
		"data":     data,
		"location": fmt.Sprintf("/sandbox/%s/%s", collectionPath, id),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}

func (h *SandboxHandler) handleDelete(w http.ResponseWriter, r *http.Request, collectionPath string, id string) {
	ctx := r.Context()

	if id == "" {
		http.Error(w, "Document ID is required for delete", http.StatusBadRequest)
		return
	}

	// Delete document
	if err := h.client.Delete(ctx, collectionPath, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to delete document: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Document deleted successfully"})
}

func (h *SandboxHandler) handleUpdate(w http.ResponseWriter, r *http.Request, collectionPath string, id string) {
	ctx := r.Context()

	if id == "" {
		http.Error(w, "Document ID is required for update", http.StatusBadRequest)
		return
	}

	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusBadRequest)
		return
	}

	// Parse JSON
	var data database.Document
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Add metadata
	if data == nil {
		data = make(database.Document)
	}
	data["_updatedAt"] = "server_timestamp" // Will be handled by database adapter
	data["_updatedBy"] = "anonymous"        // TODO: Add auth

	// Update document
	if err := h.client.Update(ctx, collectionPath, id, data); err != nil {
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to update document: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Document updated successfully",
		"id":      id,
	})
}
