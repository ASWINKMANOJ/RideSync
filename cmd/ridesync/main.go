package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/aswinkmanoj/RideSync/internal/cache"
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

	driverHub := hub.NewHub()

	// 2. Inject dependencies into API Handlers
	api := &handlers.API{
		Redis: redisCache,
		Hub:   driverHub,
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
