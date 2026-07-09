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

	// Initialize Firebase.
	app, err := database.NewFirebaseApp()
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}
	log.Printf("Successfully connected to Firebase project: %s", cfg.ProjectID)

	// Initialize Firestore client.
	firestoreClient, err := database.GetFirestoreClient(app)
	if err != nil {
		log.Fatalf("Failed to initialize Firestore: %v", err)
	}
	log.Printf("Successfully connected to Firestore")

	// Initialize Auth client.
	if authClient, err := database.GetAuthClient(app); err != nil {
		log.Printf("Warning: Failed to initialize Auth client: %v", err)
	} else {
		log.Printf("Firebase Auth client initialized successfully")
		_ = authClient // reserved for future use
	}

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
	sandboxHandler := sandbox.NewSandboxHandler(firestoreClient, cfg.ProjectID)

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
	log.Printf("Starting server on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
