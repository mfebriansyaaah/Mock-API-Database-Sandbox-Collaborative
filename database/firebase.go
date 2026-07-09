package database

import (
	"context"
	"fmt"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

const (
	// EnvProjectID is the environment variable key for the GCP project ID.
	EnvProjectID = "GOOGLE_CLOUD_PROJECT"
	// EnvCredentialsFile is the environment variable key for the service account JSON path.
	EnvCredentialsFile = "GOOGLE_APPLICATION_CREDENTIALS"

	// fallbackProjectID is the last-resort project ID if none is provided via env.
	// Kept as a constant (not a hardcoded value per se) so the binary still works
	// out-of-the-box in development. In production this MUST be set via env.
	fallbackProjectID = "mockapi-sandbox-dev"
)

// GetProjectID returns the GCP project ID from env, or the fallback constant.
// Logs a warning when falling back so the misconfiguration is visible.
func GetProjectID() string {
	if pid := os.Getenv(EnvProjectID); pid != "" {
		return pid
	}
	log.Printf("Warning: %s is not set, using fallback project ID %q", EnvProjectID, fallbackProjectID)
	return fallbackProjectID
}

// GetCredentialsFile returns the path to the service account JSON file from env.
// Returns an empty string if not set (caller may then rely on Application Default Credentials).
func GetCredentialsFile() string {
	return os.Getenv(EnvCredentialsFile)
}

// NewFirebaseApp creates a new Firebase app instance.
// Reads GOOGLE_APPLICATION_CREDENTIALS from env; if not set, falls back to
// Application Default Credentials (suitable for Cloud Run / GCE).
func NewFirebaseApp() (*firebase.App, error) {
	ctx := context.Background()

	var app *firebase.App
	var err error

	if credPath := GetCredentialsFile(); credPath != "" {
		app, err = firebase.NewApp(ctx, nil, option.WithCredentialsFile(credPath))
	} else {
		app, err = firebase.NewApp(ctx, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase app: %v", err)
	}

	return app, nil
}

// GetAuthClient returns the Firebase Auth client.
func GetAuthClient(app *firebase.App) (*auth.Client, error) {
	client, err := app.Auth(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %v", err)
	}
	return client, nil
}
