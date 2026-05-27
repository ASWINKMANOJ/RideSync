package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/aswinkmanoj/RideSync/internal/cache"
	"github.com/aswinkmanoj/RideSync/internal/db"
	"github.com/aswinkmanoj/RideSync/internal/hub"
	"github.com/gorilla/websocket"
)

type API struct {
	Redis    *cache.RedisCache
	Hub      *hub.Hub
	Postgres *db.PostgresStore
}

type LocationUpdate struct {
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"lng"`
	VehicleType string  `json:"vehicle_type"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func (api *API) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	driverID := r.URL.Query().Get("id")
	if driverID == "" {
		http.Error(w, "Driver ID is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	api.Hub.Lock()
	api.Hub.Connections[driverID] = conn
	api.Hub.Unlock()

	defer func() {
		api.Hub.Lock()
		delete(api.Hub.Connections, driverID)
		api.Hub.Unlock()
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
		err = api.Redis.SaveDriverLocation(r.Context(), driverID, update.Latitude, update.Longitude, update.VehicleType)
		if err != nil {
			log.Printf("Failed to save location for %s: %v", driverID, err)
		}

		api.Postgres.LocationChan <- db.DBLocation{
			DriverID:    driverID,
			Latitude:    update.Latitude,
			Longitude:   update.Longitude,
			VehicleType: update.VehicleType,
		}

		conn.WriteJSON(map[string]string{"status": "received"})
	}
}

func (api *API) HandleRideRequest(w http.ResponseWriter, r *http.Request) {
	riderLatStr := r.URL.Query().Get("rider_lat")
	riderLngStr := r.URL.Query().Get("rider_lng")
	riderVehicle := r.URL.Query().Get("vehicle_type")

	riderLat, errLat := strconv.ParseFloat(riderLatStr, 64)
	riderLng, errLng := strconv.ParseFloat(riderLngStr, 64)

	if errLat != nil || errLng != nil {
		http.Error(w, "Invalid Coordinates", http.StatusBadRequest)
		return
	}

	drivers, err := api.Redis.GetNearbyDrivers(r.Context(), riderLat, riderLng, riderVehicle)
	if err != nil {
		http.Error(w, "Failed to search for drivers", http.StatusInternalServerError)
		return
	}

	if len(drivers) == 0 {
		http.Error(w, "No drivers available", http.StatusNotFound)
		return
	}

	closestDriverID := drivers[0].DriverID

	api.Hub.RLock()
	driverConn, exists := api.Hub.Connections[closestDriverID]
	api.Hub.RUnlock()

	if exists {
		offerPayload := map[string]interface{}{
			"type": "RIDE_OFFER",
			"lat":  riderLat,
			"lng":  riderLng,
		}

		err := driverConn.WriteJSON(offerPayload)
		if err != nil {
			http.Error(w, "Failed to send offer", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status" : "ride requested", "driver_id": " ` + closestDriverID + `"}`))
		return
	}
}

func (api *API) HandleGetNearby(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	vehicleType := r.URL.Query().Get("vehicle_type")

	lat, errLat := strconv.ParseFloat(latStr, 64)
	lng, errLng := strconv.ParseFloat(lngStr, 64)

	if errLat != nil || errLng != nil {
		http.Error(w, "Invalid coordinates", http.StatusBadRequest)
		return
	}

	drivers, err := api.Redis.GetNearbyDrivers(r.Context(), lat, lng, vehicleType)
	if err != nil {
		http.Error(w, "Failed to search for drivers", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(drivers); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
