package main

import (
	"log"
	"net/http"

	"github.com/Murzuqisah/PoSA/handlers"
	"github.com/Murzuqisah/PoSA/middleware"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("POST /api/submit", handlers.Submit)
	mux.HandleFunc("GET /api/verify/{cid}", handlers.Verify)

	chain := middleware.Recovery(
		middleware.RequestID(
			middleware.SecurityHeaders(mux),
		),
	)

	log.Println("PoSA backend listening on :8080")
	if err := http.ListenAndServe(":8080", chain); err != nil {
		log.Fatal(err)
	}
}
