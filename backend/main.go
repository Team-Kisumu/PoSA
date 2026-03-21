package main

import (
	"log"
	"net/http"

	"github.com/Murzuqisah/PoSA/handlers"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("POST /api/submit", handlers.Submit)
	mux.HandleFunc("GET /api/verify/{cid}", handlers.Verify)

	log.Println("PoSA backend listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
