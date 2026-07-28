package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/internal/apikey"
	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/internal/config"
	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/internal/database"
	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/internal/logger"
	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/internal/middleware"
	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/internal/sandbox"
)

func main() {
	// Load configuration (env + optional .env).
	cfg := config.Load()

	// Initialize database client based on configuration
	dbConfig := createDatabaseConfig(&cfg)
	dbClient, err := database.NewDatabaseClient(context.Background(), dbConfig)
	if err != nil {
		log.Fatalf("Failed to initialize database client: %v", err)
	}
	defer dbClient.Close()

	log.Printf("Successfully connected to %s database", dbConfig.Type)

	// Initialize the asynchronous logger.
	ctx := context.Background()
	l, err := logger.NewLogger(ctx, &logger.LoggerConfig{
		ProjectID:         cfg.ProjectID,
		ChannelBuffer:     cfg.Logger.ChannelBuffer,
		NumWorkers:        cfg.Logger.NumWorkers,
		CleanupInterval:   cfg.Logger.CleanupInterval,
		MaxLogsPerProject: cfg.Logger.MaxLogsPerProject,
	})
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer l.Close()

	// Initialize sandbox handler.
	sandboxHandler := sandbox.NewSandboxHandler(dbClient, cfg.ProjectID)

	// Initialize API key manager.
	apiKeyManager := apikey.NewManager(dbClient)

	// HTTP routes.
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, Firebase World!"))
	})

	// ── API Key management endpoints ──────────────────────────────────
	// POST /__api/keys — generate a new key
	mux.HandleFunc("POST /__api/keys", handleCreateKey(apiKeyManager))
	// GET  /__api/keys?projectId=... — list keys for a project
	mux.HandleFunc("GET /__api/keys", handleListKeys(apiKeyManager))
	// DELETE /__api/keys/{projectId}/{keyId} — revoke (soft-delete) a key
	mux.HandleFunc("DELETE /__api/keys/{projectId}/{keyId}", handleRevokeKey(apiKeyManager))
	// DELETE /__api/keys/{projectId}/{keyId}/hard — permanently delete a key
	mux.HandleFunc("DELETE /__api/keys/{projectId}/{keyId}/hard", handleDeleteKey(apiKeyManager))
	// GET  /__api/tables?projectId=... — list tables (collections) for a project
	mux.HandleFunc("GET /__api/tables", handleListTables(dbClient))
	// DELETE /__api/projects/{projectId} — delete all backend data for a project
	mux.HandleFunc("DELETE /__api/projects/{projectId}", handleDeleteProject(dbClient))

	// ── Dynamic sandbox routes ────────────────────────────────────────
	mux.HandleFunc("GET /sandbox/{projectId}/{table}/_count", sandboxHandler.HandleCount)
	mux.HandleFunc("GET /sandbox/{projectId}/{table}", sandboxHandler.HandleRequest)
	mux.HandleFunc("GET /sandbox/{projectId}/{table}/", sandboxHandler.HandleRequest)
	mux.HandleFunc("GET /sandbox/{projectId}/{table}/{id}", sandboxHandler.HandleRequest)
	mux.HandleFunc("POST /sandbox/{projectId}/{table}", sandboxHandler.HandleRequest)
	mux.HandleFunc("POST /sandbox/{projectId}/{table}/{id}", sandboxHandler.HandleRequest)
	mux.HandleFunc("PUT /sandbox/{projectId}/{table}/{id}", sandboxHandler.HandleRequest)
	mux.HandleFunc("PATCH /sandbox/{projectId}/{table}/{id}", sandboxHandler.HandleRequest)
	mux.HandleFunc("DELETE /sandbox/{projectId}/{table}/{id}", sandboxHandler.HandleRequest)

	// Determine allowed CORS origins based on environment.
	// Development: allow all origins. Production: same-origin only.
	var allowedOrigins []string
	if cfg.IsDevelopment() {
		allowedOrigins = []string{"*"}
	} else {
		allowedOrigins = []string{""} // same-origin only
	}

	// Middleware chain: CORS → API Key auth → Access logging.
	handler := middleware.CORS(allowedOrigins)(mux)
	handler = middleware.APIKeyAuth(apiKeyManager)(handler)
	handler = middleware.Logging(l, cfg.ProjectID)(handler)

	// Start the server.
	log.Printf("Starting server on port %s with %s database backend [%s mode]",
		cfg.Port, dbConfig.Type, cfg.AppEnv)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// ── API Key management handlers ──────────────────────────────────────

func handleCreateKey(m *apikey.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ProjectID string `json:"projectId"`
			Name      string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON body"}`, http.StatusBadRequest)
			return
		}
		if req.ProjectID == "" {
			http.Error(w, `{"error":"projectId is required"}`, http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			req.Name = "unnamed"
		}

		key, err := m.Generate(r.Context(), req.ProjectID, req.Name)
		if err != nil {
			log.Printf("apikey: generate failed: %v", err)
			http.Error(w, `{"error":"failed to create API key"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(key)
	}
}

func handleListKeys(m *apikey.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.URL.Query().Get("projectId")
		if projectID == "" {
			http.Error(w, `{"error":"projectId query parameter is required"}`, http.StatusBadRequest)
			return
		}

		keys, err := m.List(r.Context(), projectID)
		if err != nil {
			http.Error(w, `{"error":"failed to list API keys"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(keys)
	}
}

func handleRevokeKey(m *apikey.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("projectId")
		keyID := r.PathValue("keyId")

		if projectID == "" || keyID == "" {
			http.Error(w, `{"error":"projectId and keyId are required"}`, http.StatusBadRequest)
			return
		}

		if err := m.Revoke(r.Context(), projectID, keyID); err != nil {
			http.Error(w, `{"error":"failed to revoke API key"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "API key revoked"})
	}
}

func handleDeleteKey(m *apikey.Manager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("projectId")
		keyID := r.PathValue("keyId")

		if projectID == "" || keyID == "" {
			http.Error(w, `{"error":"projectId and keyId are required"}`, http.StatusBadRequest)
			return
		}

		if err := m.Delete(r.Context(), projectID, keyID); err != nil {
			log.Printf("apikey: permanent delete failed: %v", err)
			http.Error(w, `{"error":"failed to permanently delete API key"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "API key permanently deleted"})
	}
}

func handleListTables(dbClient database.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.URL.Query().Get("projectId")
		if projectID == "" {
			http.Error(w, `{"error":"projectId query parameter is required"}`, http.StatusBadRequest)
			return
		}

		parentPath := fmt.Sprintf("sandbox/%s", projectID)
		tables, err := dbClient.ListCollections(r.Context(), parentPath)
		if err != nil {
			log.Printf("tables: list failed: %v", err)
			http.Error(w, `{"error":"failed to list tables"}`, http.StatusInternalServerError)
			return
		}
		if tables == nil {
			tables = []string{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"projectId": projectID,
			"tables":    tables,
		})
	}
}

func handleDeleteProject(dbClient database.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("projectId")
		if projectID == "" {
			http.Error(w, `{"error":"projectId is required"}`, http.StatusBadRequest)
			return
		}

		if err := dbClient.DeleteProject(r.Context(), projectID); err != nil {
			log.Printf("project: delete failed: %v", err)
			http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Project deleted successfully"})
	}
}

// createDatabaseConfig creates a database configuration from the main config
func createDatabaseConfig(cfg *config.Config) *database.DatabaseConfig {
	dbConfig := &database.DatabaseConfig{
		Type: database.DatabaseType(cfg.Database.Type),
	}

	switch cfg.Database.Type {
	case "firestore":
		dbConfig.FirestoreConfig = &database.FirestoreConfig{
			ProjectID:       cfg.Database.FirestoreProjectID,
			CredentialsFile: cfg.Database.FirestoreCredentialsFile,
		}
	case "postgresql":
		dbConfig.PostgreSQLConfig = &database.PostgreSQLConfig{
			Host:     cfg.Database.PostgreSQLHost,
			Port:     cfg.Database.PostgreSQLPort,
			User:     cfg.Database.PostgreSQLUser,
			Password: cfg.Database.PostgreSQLPassword,
			Database: cfg.Database.PostgreSQLDatabase,
			SSLMode:  cfg.Database.PostgreSQLSSLMode,
		}
	case "mysql":
		dbConfig.MySQLConfig = &database.MySQLConfig{
			Host:      cfg.Database.MySQLHost,
			Port:      cfg.Database.MySQLPort,
			User:      cfg.Database.MySQLUser,
			Password:  cfg.Database.MySQLPassword,
			Database:  cfg.Database.MySQLDatabase,
			ParseTime: true,
		}
	}

	return dbConfig
}
