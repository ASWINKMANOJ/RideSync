package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/aswinkmanoj/RideSync/internal/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {

	redisCache, err := cache.NewRedisCache("localhost:6379")
	if err != nil {
		log.Fatalf("Critical startup error: %v", err)
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "service": "ridesync"}`))
	})

	r.Get("/api/v1/drivers/nearby", func(w http.ResponseWriter, r *http.Request) {
		latStr := r.URL.Query().Get("lat")
		lngStr := r.URL.Query().Get("lng")
		radiusStr := r.URL.Query().Get("radius")

		lat, errLat := strconv.ParseFloat(latStr, 64)
		lng, errLng := strconv.ParseFloat(lngStr, 64)
		radius, errRad := strconv.ParseFloat(radiusStr, 64)

		if errLat != nil || errLng != nil || errRad != nil {
			// http.Error is a handy helper to send a text response with a status code
			http.Error(w, "Invalid lat, lng, or radius. Must be valid numbers.", http.StatusBadRequest)
			return
		}

		drivers, err := redisCache.GetNearbyDrivers(r.Context(), lat, lng, radius)
		if err != nil {
			http.Error(w, "Failed to search for drivers", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(drivers); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	})

	port := ":8080"
	fmt.Printf("RideSync backend starting on port %s...\n", port)

	err = http.ListenAndServe(port, r)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}
