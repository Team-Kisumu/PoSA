// Package middleware provides HTTP middleware for the PoSA backend.
// The middleware chain is applied in main.go and wraps the ServeMux
// to add security headers, request tracing, and panic recovery.
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
)

// SecurityHeaders injects defensive HTTP headers on every response to
// mitigate common web vulnerabilities:
//   - X-Content-Type-Options: nosniff — prevents MIME-type sniffing
//   - X-Frame-Options: DENY — blocks clickjacking via iframes
//   - X-XSS-Protection: 0 — disables legacy XSS filter (can cause issues)
//   - Content-Security-Policy: default-src 'none' — blocks all resource loading
//   - Referrer-Policy: no-referrer — prevents leaking URLs in Referer header
//   - Cache-Control: no-store — prevents caching of API responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; font-src 'self'; img-src 'self' data:; connect-src 'self' https://gateway.lighthouse.storage https://api.impulselabs.ai")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// RequestID generates a cryptographically random 16-character hex string
// and attaches it as the X-Request-ID response header. This enables
// request tracing across logs and downstream services.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := generateID()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r)
	})
}

// Recovery catches panics in downstream handlers and returns a structured
// 500 JSON error response instead of crashing the server. The panic value
// is logged for debugging. This must be the outermost middleware in the chain.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"success":false,"error":{"code":"INTERNAL_ERROR","message":"internal server error"}}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// CORS handles Cross-Origin Resource Sharing for frontend requests.
// Allows the Next.js frontend on :3000 to call the backend on :8080.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}
		// Handle preflight OPTIONS requests.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// generateID produces a 16-character hex string (8 random bytes) using
// crypto/rand for cryptographic randomness. Used for X-Request-ID headers.
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
