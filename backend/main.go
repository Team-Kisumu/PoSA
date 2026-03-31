// Package main is the entrypoint for the PoSA backend API server.
// It registers HTTP routes and applies the middleware chain before
// starting the server on port 8080.
package main

import (
	"log"
	"net/http"

	"github.com/Murzuqisah/PoSA/handlers"
	"github.com/Murzuqisah/PoSA/middleware"
)

func main() {
	mux := http.NewServeMux()

	// Register API routes using Go 1.22+ method-pattern syntax.
	// Each route is restricted to a single HTTP method at the mux level.
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("POST /api/submit", handlers.Submit)
	mux.HandleFunc("GET /api/verify/{cid}", handlers.Verify)

	// Build the middleware chain (outermost runs first):
	//   Recovery  → catches panics, returns 500 JSON
	//   RequestID → attaches unique X-Request-ID header
	//   CORS      → allows cross-origin requests from frontend
	//   SecurityHeaders → sets defensive HTTP headers
	chain := middleware.Recovery(
		middleware.RequestID(
			middleware.CORS(
				middleware.SecurityHeaders(mux),
			),
		),
	)

	log.Println("PoSA backend listening on :8080")
	if err := http.ListenAndServe(":8080", chain); err != nil {
		log.Fatal(err)
	}
}
