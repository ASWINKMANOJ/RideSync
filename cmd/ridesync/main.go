package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aswinkmanoj/RideSync/internal/cache"
	"github.com/aswinkmanoj/RideSync/internal/db"
	"github.com/aswinkmanoj/RideSync/internal/handlers"
	"github.com/aswinkmanoj/RideSync/internal/hub"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	redisCache, err := cache.NewRedisCache("localhost:6379")
	if err != nil {
		log.Fatalf("Critical startup error: %v", err)
	}

	pgConnStr := "postgres://admin:secretpassword@localhost:5432/ridesync?sslmode=disable"
	pgStore, err := db.NewPostgresStore(pgConnStr)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}

	driverHub := hub.NewHub()

	api := &handlers.API{
		Redis:    redisCache,
		Hub:      driverHub,
		Postgres: pgStore,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "service": "ridesync"}`))
	})

	r.Get("/api/v1/drivers/ws", api.HandleWebSocket)
	r.Post("/api/v1/rides/request", api.HandleRideRequest)
	r.Get("/api/v1/drivers/nearby", api.HandleGetNearby)

	port := ":8080"
	fmt.Printf("RideSync backend starting on port %s...\n", port)

	err = http.ListenAndServe(port, r)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}
