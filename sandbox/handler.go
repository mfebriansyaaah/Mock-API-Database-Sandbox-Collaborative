package sandbox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

// SandboxHandler handles dynamic sandbox endpoints
type SandboxHandler struct {
	client    *firestore.Client
	projectID string
}

// NewSandboxHandler creates a new SandboxHandler
type NewSandboxHandlerFunc func(client *firestore.Client, projectID string) *SandboxHandler

func NewSandboxHandler(client *firestore.Client, projectID string) *SandboxHandler {
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
		docRef := h.client.Doc(collectionPath + "/" + id)
		docSnap, err := docRef.Get(ctx)
		if err != nil {
			// Check if this is a "not found" error from Firestore
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "NotFound") {
				http.Error(w, "Document not found", http.StatusNotFound)
				return
			}
			http.Error(w, fmt.Sprintf("Failed to get document: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(docSnap.Data()); err != nil {
			http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
			return
		}
		return
	}

	// Get all documents in collection
	iter := h.client.Collection(collectionPath).Documents(ctx)
	var results []map[string]interface{}

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
			http.Error(w, fmt.Sprintf("Failed to iterate documents: %v", err), http.StatusInternalServerError)
			return
		}
		data := doc.Data()
		data["id"] = doc.Ref.ID
		results = append(results, data)
	}

	w.Header().Set("Content-Type", "application/json")
	if results == nil {
		results = []map[string]interface{}{}
	}
	if err := json.NewEncoder(w).Encode(results); err != nil {
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
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Add metadata
	if data == nil {
		data = make(map[string]interface{})
	}
	data["_createdAt"] = firestore.ServerTimestamp
	data["_createdBy"] = "anonymous" // TODO: Add auth

	// Use provided ID or generate new one
	if id == "" {
		id = uuid.New().String()
	}

	// Set document
	docRef := h.client.Doc(collectionPath + "/" + id)
	if _, err := docRef.Set(ctx, data); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create document: %v", err), http.StatusInternalServerError)
		return
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

	// First check if document exists
	docRef := h.client.Doc(collectionPath + "/" + id)
	_, err := docRef.Get(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "NotFound") {
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to check document: %v", err), http.StatusInternalServerError)
		return
	}

	// Document exists, delete it
	if _, err := docRef.Delete(ctx); err != nil {
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
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Add metadata
	if data == nil {
		data = make(map[string]interface{})
	}
	data["_updatedAt"] = firestore.ServerTimestamp
	data["_updatedBy"] = "anonymous" // TODO: Add auth

	// Update document
	docRef := h.client.Doc(collectionPath + "/" + id)
	if _, err := docRef.Set(ctx, data, firestore.MergeAll); err != nil {
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
