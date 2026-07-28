package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/mfebriansyaaah/Mock-API-Database-Sandbox-Collaborative/internal/apikey"
)

type contextKey string

const APIKeyContextKey contextKey = "apikey"

// APIKeyAuth optionally validates the X-API-Key header on sandbox requests.
// If the key is present and valid, the validated *apikey.APIKey is stored in the request context.
// If the key is missing, the request passes through (backward compatible for the web UI).
// If the key is present but invalid, the request is rejected with 401.
func APIKeyAuth(manager *apikey.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only intercept on sandbox paths
			if !strings.HasPrefix(r.URL.Path, "/sandbox/") {
				next.ServeHTTP(w, r)
				return
			}

			key := r.Header.Get("X-API-Key")
			if key == "" {
				key = r.URL.Query().Get("api_key")
			}

			// No key provided — pass through (web UI and unauthenticated clients).
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Extract projectId from URL path: /sandbox/{projectId}/...
			parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/sandbox/"), "/")
			if len(parts) == 0 || parts[0] == "" {
				http.Error(w, `{"error":"cannot determine projectId from path"}`, http.StatusBadRequest)
				return
			}
			projectId := parts[0]

			apiKey, err := manager.ValidateForProject(r.Context(), projectId, key)
			if err != nil {
				log.Printf("apikey: rejected key for project %s: %v", projectId, err)
				http.Error(w, `{"error":"invalid or inactive API key"}`, http.StatusUnauthorized)
				return
			}

			// Store the validated key in context for downstream handlers.
			ctx := context.WithValue(r.Context(), APIKeyContextKey, apiKey)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
