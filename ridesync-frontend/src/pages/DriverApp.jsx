import { Car, Check, MapPin, Power } from "lucide-react";
import { useEffect, useRef, useState } from "react";

export default function DriverApp() {
  const [driverId, setDriverId] = useState("uber_001");
  const [vehicleType, setVehicleType] = useState("SUV");
  const [isOnline, setIsOnline] = useState(false);

  // Stores the target location when a ride is offered
  const [activeRide, setActiveRide] = useState(null);

  const driverLocRef = useRef([30.91, 75.852]);
  const wsRef = useRef(null);
  const telemetryIntervalRef = useRef(null);

  const toggleStatus = () => {
    if (isOnline) {
      // Go Offline
      wsRef.current?.close();
      setIsOnline(false);
      setActiveRide(null);
      clearInterval(telemetryIntervalRef.current);
    } else {
      // Go Online - Connect to Go Backend
      const socket = new WebSocket(
        `ws://localhost:8080/api/v1/drivers/ws?id=${driverId}`,
      );

      socket.onopen = () => {
        setIsOnline(true);

        // IMMEDIATE TELEMETRY: We must send location to Go so Redis can map us via H3
        telemetryIntervalRef.current = setInterval(() => {
          if (socket.readyState === WebSocket.OPEN) {
            // If active ride, simulate driving toward rider. Otherwise, idle.
            if (activeRide) {
              driverLocRef.current = [
                driverLocRef.current[0] - 0.0002,
                driverLocRef.current[1] - 0.0005,
              ];
            }

            // Exactly matches your Go `LocationUpdate` struct
            socket.send(
              JSON.stringify({
                lat: driverLocRef.current[0],
                lng: driverLocRef.current[1],
                vehicle_type: vehicleType,
              }),
            );
          }
        }, 3000);
      };

      socket.onmessage = (event) => {
        const data = JSON.parse(event.data);

        // Matches the "RIDE_OFFER" payload sent from HandleRideRequest
        if (data.type === "RIDE_OFFER") {
          setActiveRide({ lat: data.lat, lng: data.lng });
        }
      };

      socket.onclose = () => {
        setIsOnline(false);
        clearInterval(telemetryIntervalRef.current);
      };

      wsRef.current = socket;
    }
  };

  // Clean up on unmount
  useEffect(() => {
    return () => {
      wsRef.current?.close();
      clearInterval(telemetryIntervalRef.current);
    };
  }, []);

  return (
    <div
      style={{
        flex: 1,
        backgroundColor: "#0a0a0a",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        color: "white",
        padding: "20px",
        height: "100vh",
      }}
    >
      <div
        style={{
          backgroundColor: "#111",
          padding: "40px",
          borderRadius: "16px",
          boxShadow: "0 20px 40px rgba(0,0,0,0.4)",
          width: "100%",
          maxWidth: "450px",
          border: "1px solid #222",
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            justifyContent: "space-between",
            marginBottom: "30px",
          }}
        >
          <h2
            style={{
              margin: 0,
              fontWeight: "600",
              display: "flex",
              alignItems: "center",
              gap: "10px",
            }}
          >
            <Car color="#00ADD8" /> Driver Terminal
          </h2>
          <div
            style={{
              padding: "8px 16px",
              borderRadius: "20px",
              fontSize: "14px",
              fontWeight: "bold",
              backgroundColor: isOnline
                ? "rgba(40, 167, 69, 0.1)"
                : "rgba(255, 60, 60, 0.1)",
              color: isOnline ? "#28a745" : "#ff3c3c",
              border: `1px solid ${isOnline ? "#28a745" : "#ff3c3c"}`,
            }}
          >
            {isOnline ? "● ONLINE" : "○ OFFLINE"}
          </div>
        </div>

        {/* Configuration Inputs */}
        {!isOnline && (
          <div
            style={{
              display: "flex",
              flexDirection: "column",
              gap: "15px",
              marginBottom: "20px",
            }}
          >
            <div>
              <label
                htmlFor="driverId"
                style={{
                  display: "block",
                  fontSize: "12px",
                  color: "#888",
                  marginBottom: "8px",
                }}
              >
                DRIVER ID
              </label>
              <input
                type="text"
                value={driverId}
                onChange={(e) => setDriverId(e.target.value)}
                style={{
                  width: "100%",
                  padding: "12px",
                  borderRadius: "8px",
                  border: "1px solid #333",
                  backgroundColor: "#000",
                  color: "white",
                  outline: "none",
                  boxSizing: "border-box",
                }}
              />
            </div>
            <div>
              <label
                htmlFor="vehicleType"
                style={{
                  display: "block",
                  fontSize: "12px",
                  color: "#888",
                  marginBottom: "8px",
                }}
              >
                VEHICLE TYPE
              </label>
              <select
                id="vehicleType"
                value={vehicleType}
                onChange={(e) => setVehicleType(e.target.value)}
                style={{
                  width: "100%",
                  padding: "12px",
                  borderRadius: "8px",
                  border: "1px solid #333",
                  backgroundColor: "#000",
                  color: "white",
                  outline: "none",
                  boxSizing: "border-box",
                }}
              >
                <option value="SUV">SUV</option>
                <option value="SEDAN">SEDAN</option>
              </select>
            </div>
          </div>
        )}

        {/* Incoming Ride Alert */}
        {activeRide && (
          <div
            style={{
              backgroundColor: "rgba(0, 173, 216, 0.1)",
              border: "2px solid #00ADD8",
              borderRadius: "12px",
              padding: "20px",
              marginBottom: "20px",
              textAlign: "center",
            }}
          >
            <h3 style={{ margin: "0 0 10px 0", color: "#00ADD8" }}>
              Ride Dispatched!
            </h3>
            <p
              style={{ color: "#aaa", fontSize: "14px", marginBottom: "15px" }}
            >
              Proceed to rider at: <br /> {activeRide.lat.toFixed(4)},{" "}
              {activeRide.lng.toFixed(4)}
            </p>
            <div
              style={{
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                gap: "10px",
                color: "#28a745",
                fontWeight: "bold",
              }}
            >
              <Check size={20} /> Route Active
            </div>
          </div>
        )}

        {/* Status Tracker */}
        {isOnline && !activeRide && (
          <div
            style={{
              marginBottom: "20px",
              padding: "15px",
              backgroundColor: "#000",
              borderRadius: "8px",
              display: "flex",
              alignItems: "center",
              gap: "15px",
              border: "1px solid #222",
            }}
          >
            <MapPin color="#00ADD8" size={24} />
            <div>
              <p style={{ margin: 0, fontSize: "12px", color: "#888" }}>
                Broadcasting Telemetry via WebSocket
              </p>
              <p
                style={{
                  margin: "4px 0 0 0",
                  fontSize: "14px",
                  fontFamily: "monospace",
                  color: "#fff",
                }}
              >
                Redis H3 Grid Updated
              </p>
            </div>
          </div>
        )}

        <button
          type="button"
          onClick={toggleStatus}
          style={{
            width: "100%",
            padding: "20px",
            borderRadius: "12px",
            border: "none",
            backgroundColor: isOnline ? "#222" : "#00ADD8",
            color: isOnline ? "#fff" : "#000",
            fontSize: "18px",
            fontWeight: "bold",
            cursor: "pointer",
            display: "flex",
            alignItems: "center",
            justifyContent: "center",
            gap: "10px",
            transition: "all 0.2s ease",
          }}
        >
          <Power size={24} />
          {isOnline ? "GO OFFLINE" : "GO ONLINE"}
        </button>
      </div>
    </div>
  );
}
