import L from "leaflet";
import { useEffect, useState } from "react";
import { MapContainer, Marker, Polyline, TileLayer } from "react-leaflet";
import "leaflet/dist/leaflet.css";
import { Car, Search } from "lucide-react";

// Sleek Custom Icons
const riderIcon = new L.divIcon({
  className: "rider-marker",
  html: `<div class="pulse-dot"></div>`,
  iconSize: [20, 20],
});
const driverIcon = new L.divIcon({
  className: "driver-marker",
  html: `<div class="car-badge">DRV</div>`,
  iconSize: [40, 30],
});
const dispatchedIcon = new L.divIcon({
  className: "driver-marker",
  html: `<div class="car-badge" style="border-color: #28a745; background-color: #222;">DISPATCHED</div>`,
  iconSize: [80, 30],
});

export default function RiderMap() {
  const [riderLoc] = useState([30.908, 75.85]);
  const [vehicleType, setVehicleType] = useState(""); // "" (All), "SUV", or "SEDAN"

  const [nearbyDrivers, setNearbyDrivers] = useState([]);
  const [rideStatus, setRideStatus] = useState("IDLE"); // IDLE, REQUESTING, DISPATCHED, FAILED
  const [dispatchedDriverId, setDispatchedDriverId] = useState(null);

  // 1. Fetch from Go `GET /nearby` continuously
  useEffect(() => {
    const fetchDrivers = async () => {
      try {
        const url = `http://localhost:8080/api/v1/drivers/nearby?lat=${riderLoc[0]}&lng=${riderLoc[1]}${vehicleType ? `&vehicle_type=${vehicleType}` : ""}`;
        const response = await fetch(url);

        if (response.ok) {
          const data = await response.json();
          // Go returns array of `DriverLocation` structs
          setNearbyDrivers(data || []);
        }
      } catch (error) {
        console.error("Failed to fetch drivers", error);
      }
    };

    fetchDrivers();
    const interval = setInterval(fetchDrivers, 3000); // Polling
    return () => clearInterval(interval);
  }, [vehicleType, riderLoc]);

  // 2. Dispatch via Go `POST /request`
  const requestRide = async () => {
    setRideStatus("REQUESTING");

    try {
      const url = `http://localhost:8080/api/v1/rides/request?rider_lat=${riderLoc[0]}&rider_lng=${riderLoc[1]}${vehicleType ? `&vehicle_type=${vehicleType}` : ""}`;
      const response = await fetch(url, { method: "POST" });

      if (response.ok) {
        const data = await response.json();

        // NOTE: Your Go backend code outputs `{"driver_id": " ` + closestDriverID + `"}`.
        // We MUST trim the leading space from the Go string concatenation.
        const cleanDriverId = data.driver_id.trim();

        setDispatchedDriverId(cleanDriverId);
        setRideStatus("DISPATCHED");
      } else {
        setRideStatus("FAILED");
        setTimeout(() => setRideStatus("IDLE"), 3000);
      }
    } catch (error) {
      console.error("Error connecting to server:", error);
      setRideStatus("FAILED");
      setTimeout(() => setRideStatus("IDLE"), 3000);
    }
  };

  // Find the exact object of our assigned driver to draw the route line
  const activeDriver = nearbyDrivers.find(
    (d) => d.driver_id === dispatchedDriverId,
  );

  return (
    <div
      style={{
        flex: 1,
        display: "flex",
        flexDirection: "column",
        position: "relative",
        height: "100vh",
      }}
    >
      {/* Sleek UI Overlay */}
      <div
        style={{
          position: "absolute",
          bottom: "40px",
          left: "50%",
          transform: "translateX(-50%)",
          zIndex: 1000,
          backgroundColor: "#111",
          padding: "20px",
          borderRadius: "16px",
          color: "white",
          boxShadow: "0 10px 40px rgba(0,0,0,0.5)",
          width: "380px",
          textAlign: "center",
          border: "1px solid #333",
        }}
      >
        {rideStatus === "IDLE" && (
          <>
            <h3 style={{ margin: "0 0 15px 0" }}>Where to?</h3>
            <div style={{ display: "flex", gap: "10px", marginBottom: "20px" }}>
              {["", "SUV", "SEDAN"].map((type) => (
                <button
                  type="button"
                  key={type}
                  onClick={() => setVehicleType(type)}
                  style={{
                    flex: 1,
                    padding: "10px",
                    borderRadius: "8px",
                    border: "1px solid #333",
                    cursor: "pointer",
                    fontWeight: "bold",
                    backgroundColor: vehicleType === type ? "#00ADD8" : "#222",
                    color: vehicleType === type ? "#000" : "#aaa",
                  }}
                >
                  {type === "" ? "ANY" : type}
                </button>
              ))}
            </div>

            <button
              type="button"
              onClick={requestRide}
              style={{
                width: "100%",
                padding: "16px",
                backgroundColor: "#00ADD8",
                color: "black",
                border: "none",
                borderRadius: "8px",
                fontSize: "16px",
                fontWeight: "bold",
                cursor: "pointer",
                display: "flex",
                justifyContent: "center",
                alignItems: "center",
                gap: "10px",
              }}
            >
              <Car size={20} /> Request Ride
            </button>
          </>
        )}

        {rideStatus === "REQUESTING" && (
          <h3
            style={{
              margin: 0,
              color: "#00ADD8",
              animation: "pulse 1.5s infinite",
              display: "flex",
              justifyContent: "center",
              alignItems: "center",
              gap: "10px",
            }}
          >
            <Search size={20} /> Finding nearest via H3...
          </h3>
        )}

        {rideStatus === "FAILED" && (
          <h3 style={{ margin: 0, color: "#ff3c3c" }}>
            No drivers available. Try again.
          </h3>
        )}

        {rideStatus === "DISPATCHED" && (
          <>
            <h3 style={{ margin: "0 0 5px 0", color: "#28a745" }}>
              {dispatchedDriverId} En Route!
            </h3>
            <p style={{ margin: 0, color: "#ccc", fontSize: "14px" }}>
              Tracking live via HTTP Poll...
            </p>
          </>
        )}
      </div>

      <MapContainer
        center={riderLoc}
        zoom={14}
        style={{ height: "100%", width: "100%" }}
        zoomControl={false}
      >
        <TileLayer
          url="https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png"
          attribution="&copy; CARTO"
        />

        {/* Rider */}
        <Marker position={riderLoc} icon={riderIcon} />

        {/* Render Drivers. If dispatched, hide all other drivers. */}
        {nearbyDrivers.map((driver) => {
          if (
            rideStatus === "DISPATCHED" &&
            driver.driver_id !== dispatchedDriverId
          )
            return null;

          return (
            <Marker
              key={driver.driver_id}
              // BINDING TO THE EXACT KEYS FROM YOUR GO STRUCT
              position={[driver.latitude, driver.longitude]}
              icon={
                driver.driver_id === dispatchedDriverId
                  ? dispatchedIcon
                  : driverIcon
              }
            />
          );
        })}

        {/* Active Route Polyline */}
        {rideStatus === "DISPATCHED" && activeDriver && (
          <Polyline
            positions={[
              [activeDriver.latitude, activeDriver.longitude],
              riderLoc,
            ]}
            pathOptions={{
              color: "#00ADD8",
              weight: 4,
              dashArray: "10, 10",
              opacity: 0.8,
            }}
          />
        )}
      </MapContainer>
    </div>
  );
}
