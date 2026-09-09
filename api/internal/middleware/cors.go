package middleware

import (
	"fmt"
	"net/http"
	"os"
)

// CORS returns middleware that handles Cross-Origin Resource Sharing.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get allowed origins from environment or default to localhost:3000
		allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")

		fmt.Println("CORS_ALLOWED_ORIGINS:", allowedOrigins) // Debugging line
		if allowedOrigins == "" {
			allowedOrigins = "http://localhost:3000"
		}

		origin := r.Header.Get("Origin")
		if origin == allowedOrigins {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
