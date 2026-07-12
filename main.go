package main

import (
	"context"
	"log"
	"net/http"

	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/config"
	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/database"
	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/logger"
	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/middleware"
	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/sandbox"
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

	// HTTP routes.
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, Firebase World!"))
	})

	// Dynamic sandbox routes: /sandbox/{projectId}/{table}
	mux.HandleFunc("GET /sandbox/{projectId}/{table}", sandboxHandler.HandleRequest)
	mux.HandleFunc("GET /sandbox/{projectId}/{table}/", sandboxHandler.HandleRequest)
	mux.HandleFunc("GET /sandbox/{projectId}/{table}/{id}", sandboxHandler.HandleRequest)
	mux.HandleFunc("POST /sandbox/{projectId}/{table}", sandboxHandler.HandleRequest)
	mux.HandleFunc("POST /sandbox/{projectId}/{table}/{id}", sandboxHandler.HandleRequest)
	mux.HandleFunc("PUT /sandbox/{projectId}/{table}/{id}", sandboxHandler.HandleRequest)
	mux.HandleFunc("PATCH /sandbox/{projectId}/{table}/{id}", sandboxHandler.HandleRequest)
	mux.HandleFunc("DELETE /sandbox/{projectId}/{table}/{id}", sandboxHandler.HandleRequest)

	// Wrap with logging middleware.
	handler := middleware.Logging(l, cfg.ProjectID)(mux)

	// Start the server.
	log.Printf("Starting server on port %s with %s database backend", cfg.Port, dbConfig.Type)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
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
