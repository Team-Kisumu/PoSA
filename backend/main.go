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

	"github.com/Murzuqisah/PoSA/handlers"
	"github.com/Murzuqisah/PoSA/middleware"
)

func main() {
	mux := http.NewServeMux()

	// Register API routes using Go 1.22+ method-pattern syntax.
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("POST /api/submit", handlers.Submit)
	mux.HandleFunc("GET /api/verify/{cid}", handlers.Verify)

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
					middleware.CSRF(mux),
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
