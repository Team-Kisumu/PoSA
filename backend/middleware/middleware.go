// Package middleware provides HTTP middleware for the PoSA backend.
// The middleware chain is applied in main.go and wraps the ServeMux
// to add security headers, CSRF protection, request tracing, and panic recovery.
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
)

// SecurityHeaders injects defensive HTTP headers on every response:
//   - X-Content-Type-Options: nosniff — prevents MIME-type sniffing
//   - X-Frame-Options: DENY — blocks clickjacking via iframes
//   - X-XSS-Protection: 0 — disables legacy XSS filter (can cause issues)
//   - Content-Security-Policy — controls resource loading origins
//   - Referrer-Policy: strict-origin-when-cross-origin — safe referrer behavior
//   - Strict-Transport-Security — enforces HTTPS for 1 year
//   - Permissions-Policy — restricts browser feature access
//   - Cross-Origin-Opener-Policy — isolates browsing context
//   - Cross-Origin-Resource-Policy — restricts cross-origin resource loading
//   - Cache-Control — prevents caching of API responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// MIME sniffing prevention.
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Clickjacking protection.
		w.Header().Set("X-Frame-Options", "DENY")

		// Disable legacy XSS auditor (causes more issues than it solves).
		w.Header().Set("X-XSS-Protection", "0")

		// Content Security Policy: allow same-origin resources + required externals.
		w.Header().Set("Content-Security-Policy", strings.Join([]string{
			"default-src 'self'",
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'",
			"style-src 'self' 'unsafe-inline'",
			"font-src 'self' data:",
			"img-src 'self' data: https:",
			"connect-src 'self' https://gateway.lighthouse.storage https://api.impulselabs.ai https://rest-mainnet.onflow.org https://rest-testnet.onflow.org",
			"frame-ancestors 'none'",
			"base-uri 'self'",
			"form-action 'self'",
		}, "; "))

		// Referrer policy: send origin on cross-origin, full URL on same-origin.
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// HSTS: enforce HTTPS for 1 year, include subdomains.
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Restrict browser features not needed by the application.
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")

		// Cross-origin isolation headers.
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")

		// Prevent caching of API responses (static assets are served by Next.js
		// with its own cache headers via the proxy).
		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/health" {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
			w.Header().Set("Pragma", "no-cache")
		}

		next.ServeHTTP(w, r)
	})
}

// CSRF validates that state-changing requests (POST, PUT, DELETE) include
// a valid CSRF token. The token is set as a cookie on GET requests and must
// be sent back in the X-CSRF-Token header on mutations.
//
// Flow:
//  1. GET request → server sets csrf_token cookie (HttpOnly=false so JS can read it)
//  2. POST request → client sends X-CSRF-Token header with the cookie value
//  3. Server compares header value to cookie value → rejects if mismatch
//
// This double-submit cookie pattern prevents CSRF without server-side state.
func CSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Issue a CSRF token cookie on safe methods if not already set.
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			if _, err := r.Cookie("csrf_token"); err != nil {
				token := generateToken()
				http.SetCookie(w, &http.Cookie{
					Name:     "csrf_token",
					Value:    token,
					Path:     "/",
					HttpOnly: false, // JS must read this to send in header.
					Secure:   true,
					SameSite: http.SameSiteStrictMode,
					MaxAge:   86400, // 24 hours.
				})
			}
			next.ServeHTTP(w, r)
			return
		}

		// Validate CSRF token on state-changing methods.
		cookie, err := r.Cookie("csrf_token")
		if err != nil || cookie.Value == "" {
			http.Error(w, `{"success":false,"error":{"code":"CSRF_MISSING","message":"CSRF token cookie missing"}}`, http.StatusForbidden)
			return
		}

		headerToken := r.Header.Get("X-CSRF-Token")
		if headerToken == "" || headerToken != cookie.Value {
			http.Error(w, `{"success":false,"error":{"code":"CSRF_INVALID","message":"CSRF token mismatch"}}`, http.StatusForbidden)
			return
		}

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
// In development, allows cross-origin from localhost:3000.
// In production (single-origin via proxy), CORS headers are not needed
// but are included for flexibility.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// generateID produces a 16-character hex string (8 random bytes).
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// generateToken produces a 32-character hex string (16 random bytes) for CSRF tokens.
func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
