// Package main is the entrypoint for the PoSA backend API server.
// In production (Docker/Render), it serves as the single process on $PORT,
// proxying /ai/* to the Python AI engine and /* to the Next.js frontend.
// In development, it runs standalone on :8080.
package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/Murzuqisah/PoSA/auth"
	"github.com/Murzuqisah/PoSA/database"
	"github.com/Murzuqisah/PoSA/handlers"
	"github.com/Murzuqisah/PoSA/middleware"
)

func main() {
	// Initialize database.
	db, err := database.Open("")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Initialize auth handler.
	authHandler := auth.NewHandler(db)

	mux := http.NewServeMux()

	// API routes.
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("POST /api/submit", handlers.Submit)
	mux.HandleFunc("GET /api/verify/{cid}", handlers.Verify)

	// Proofs routes.
	proofsHandler := &handlers.ProofsHandler{DB: db}
	mux.HandleFunc("GET /api/proofs", proofsHandler.List)
	mux.HandleFunc("GET /api/proofs/{id}", proofsHandler.Get)

	// Submissions route (save evaluation results).
	submissionsHandler := &handlers.SubmissionsHandler{DB: db}
	mux.HandleFunc("POST /api/submissions", submissionsHandler.Create)

	// Admin routes (protected by RequireAdmin).
	adminHandler := &handlers.AdminHandler{DB: db}
	mux.Handle("GET /api/admin/users", auth.RequireAuth(auth.RequireAdmin(http.HandlerFunc(adminHandler.ListUsers))))
	mux.Handle("GET /api/admin/users/{id}", auth.RequireAuth(auth.RequireAdmin(http.HandlerFunc(adminHandler.GetUser))))
	mux.Handle("PATCH /api/admin/users/{id}", auth.RequireAuth(auth.RequireAdmin(http.HandlerFunc(adminHandler.UpdateUser))))
	mux.Handle("GET /api/admin/credentials", auth.RequireAuth(auth.RequireAdmin(http.HandlerFunc(adminHandler.ListCredentials))))
	mux.Handle("GET /api/admin/stats", auth.RequireAuth(auth.RequireAdmin(http.HandlerFunc(adminHandler.Stats))))

	// Auth routes.
	mux.HandleFunc("GET /auth/github", authHandler.GitHubLogin)
	mux.HandleFunc("GET /auth/github/callback", authHandler.GitHubCallback)
	mux.HandleFunc("GET /auth/me", authHandler.Me)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)

	// In production, proxy /ai/* to the Python AI engine and /* to Next.js.
	if os.Getenv("RENDER") == "true" || os.Getenv("DOCKER") == "true" {
		aiProxy := newProxy("http://127.0.0.1:8000", "/ai")
		mux.Handle("/ai/", aiProxy)

		frontendProxy := newProxy("http://127.0.0.1:3000", "")
		mux.Handle("/", frontendProxy)
	}

	chain := middleware.Recovery(
		middleware.RequestID(
			middleware.CORS(
				middleware.SecurityHeaders(
					authHandler.AuthMiddleware(
						middleware.CSRF(mux),
					),
				),
			),
		),
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("PoSA backend listening on :%s", port)
	if err := http.ListenAndServe(":"+port, chain); err != nil {
		log.Fatal(err)
	}
}

// newProxy creates a reverse proxy that forwards requests to the target URL.
// If stripPrefix is non-empty, it strips that prefix from the request path.
func newProxy(target, stripPrefix string) http.Handler {
	u, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(u)

	// Suppress proxy error logs for unavailable backends during startup.
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"success":false,"error":{"code":"PROXY_ERROR","message":"service unavailable"}}`))
	}

	if stripPrefix == "" {
		return proxy
	}
	return http.StripPrefix(stripPrefix, proxy)
}
