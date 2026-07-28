package middleware

import (
	"net/http"
	"strings"
)

// CORS returns an HTTP middleware that sets CORS headers based on the app
// environment. In development mode all origins are allowed. In production
// only the configured origins (or same-origin) are permitted.
func CORS(allowOrigins []string) func(http.Handler) http.Handler {
	allowOriginsStr := strings.Join(allowOrigins, ", ")
	if len(allowOrigins) == 1 && allowOrigins[0] == "*" {
		allowOriginsStr = "*"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Set Access-Control-Allow-Origin based on the request origin
			if allowOriginsStr == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else if origin != "" && containsOrigin(allowOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func containsOrigin(origins []string, origin string) bool {
	for _, o := range origins {
		if o == origin {
			return true
		}
	}
	return false
}
