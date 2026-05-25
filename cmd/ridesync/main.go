package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "service": "ridesync"}`))
	})

	port := ":8080"
	fmt.Printf("RideSync backend starting on port %s...\n", port)

	err := http.ListenAndServe(port, r)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}
