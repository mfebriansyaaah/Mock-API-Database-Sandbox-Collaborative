package sandbox

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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

// Pagination defaults and hard caps for the GET-collection endpoint.
// We cap limit to keep response bodies under Firestore's 1 MiB query limit
// (and to keep a single request cheap). Callers can ask for more by
// paginating with offset; we also surface a "nextOffset" hint so the UI
// can implement infinite scroll.
const (
	defaultPageSize = 25
	maxPageSize     = 100
)

// parseGetAllOptions extracts pagination parameters from the request.
// Returns a default options struct when no params are present.
func parseGetAllOptions(r *http.Request) *database.GetAllOptions {
	opts := &database.GetAllOptions{Limit: defaultPageSize}
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			opts.Limit = n
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil && n >= 0 {
			opts.Offset = n
		}
	}
	if opts.Limit > maxPageSize {
		opts.Limit = maxPageSize
	}
	return opts
}

// isQuotaError reports whether the underlying error from a database adapter
// is a Firestore "quota exceeded" / "resource exhausted" failure. We surface
// these as 503 so the UI can show a friendly retry message instead of a
// raw gRPC stack trace.
func isQuotaError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "ResourceExhausted") ||
		strings.Contains(msg, "Quota exceeded") ||
		strings.Contains(msg, "RESOURCE_EXHAUSTED") ||
		strings.Contains(msg, "quota_exceeded")
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
			if isQuotaError(err) {
				http.Error(w, "Upstream quota exceeded. Please retry shortly.", http.StatusServiceUnavailable)
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

	// Get documents in collection (with optional pagination)
	opts := parseGetAllOptions(r)
	docs, err := h.client.GetAll(ctx, collectionPath, opts)
	if err != nil {
		if isQuotaError(err) {
			http.Error(w,
				"Upstream quota exceeded. Try a smaller ?limit= or retry shortly.",
				http.StatusServiceUnavailable)
			return
		}
		http.Error(w, fmt.Sprintf("Failed to get documents: %v", err), http.StatusInternalServerError)
		return
	}

	if docs == nil {
		docs = []database.Document{}
	}

	// Wrap response with pagination metadata so the UI can request more.
	nextOffset := opts.Offset + len(docs)
	resp := map[string]interface{}{
		"data":       docs,
		"limit":      opts.Limit,
		"offset":     opts.Offset,
		"count":      len(docs),
		"nextOffset": nextOffset,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
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
