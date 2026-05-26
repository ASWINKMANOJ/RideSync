package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/aswinkmanoj/RideSync/internal/cache"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
)

func main() {
	type LocationUpdate struct {
		Latitude  float64 `json:"lat"`
		Longitude float64 `json:"lng"`
	}

	type DriverHub struct {
		sync.RWMutex
		Connections map[string]*websocket.Conn
	}

	var hub = DriverHub{
		Connections: make(map[string]*websocket.Conn),
	}

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,

		CheckOrigin: func(r *http.Request) bool { return true },
	}

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

	r.Get("/api/v1/drivers/ws", func(w http.ResponseWriter, r *http.Request) {
		driverID := r.URL.Query().Get("id")
		if driverID == "" {
			http.Error(w, "Driver ID is required", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade the connection: %v", err)
			return
		}

		hub.Lock()
		hub.Connections[driverID] = conn
		hub.Unlock()

		defer func() {
			hub.Lock()
			delete(hub.Connections, driverID)
			hub.Unlock()
			conn.Close()
		}()

		fmt.Printf("Driver Connected: %s\n", driverID)

		for {
			var update LocationUpdate

			err := conn.ReadJSON(&update)
			if err != nil {
				fmt.Printf("Driver Disconnected: %s\n", driverID)
				break
			}

			err = redisCache.SaveDriverLocation(r.Context(), driverID, update.Latitude, update.Longitude)
			if err != nil {
				log.Printf("Failed to save location for %s: %v", driverID, err)
				continue
			}

			conn.WriteJSON(map[string]string{"status": "received"})
		}
	})

	r.Post("/api/v1/rides/request", func(w http.ResponseWriter, r *http.Request) {
		riderLatStr := r.URL.Query().Get("rider_lat")
		riderLngStr := r.URL.Query().Get("rider_lng")

		rider_lat, errRiderLat := strconv.ParseFloat(riderLatStr, 64)
		rider_lng, errRiderLng := strconv.ParseFloat(riderLngStr, 64)

		if errRiderLat != nil || errRiderLng != nil {
			http.Error(w, "Invalid rider_lat or rider_lng", http.StatusBadRequest)
			return
		}

		drivers, err := redisCache.GetNearbyDrivers(r.Context(), rider_lat, rider_lng)

		if err != nil {
			http.Error(w, "Failed to search for drivers", http.StatusInternalServerError)
			return
		}

		if len(drivers) == 0 {
			http.Error(w, "No drivers available", http.StatusNotFound)
			return
		}

		closestDriverID := drivers[0].DriverID

		hub.RLock()
		driverConn, exists := hub.Connections[closestDriverID]
		hub.RUnlock()

		if exists {
			offerPayload := map[string]interface{}{
				"type": "RIDE_OFFER",
				"lat":  rider_lat,
				"lng":  rider_lng,
			}
			err := driverConn.WriteJSON(offerPayload)
			if err != nil {
				http.Error(w, "Failed to send offer to driver", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ride requested", "driver_id": "` + closestDriverID + `"}`))
			return
		} else {
			http.Error(w, "Driver disconnected", http.StatusServiceUnavailable)
			return
		}
	})

	r.Get("/api/v1/drivers/nearby", func(w http.ResponseWriter, r *http.Request) {
		latStr := r.URL.Query().Get("lat")
		lngStr := r.URL.Query().Get("lng")

		lat, errLat := strconv.ParseFloat(latStr, 64)
		lng, errLng := strconv.ParseFloat(lngStr, 64)

		if errLat != nil || errLng != nil {
			http.Error(w, "Invalid lat or lng", http.StatusBadRequest)
			return
		}

		// Call the updated method (Radius removed)
		drivers, err := redisCache.GetNearbyDrivers(r.Context(), lat, lng)
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
