import { MapPin, Navigation } from "lucide-react";
import { Link } from "react-router-dom";

export default function Home() {
  return (
    <div
      style={{
        flex: 1,
        backgroundColor: "#f8f9fa",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        padding: "20px",
      }}
    >
      <div
        style={{ textAlign: "center", maxWidth: "800px", marginBottom: "50px" }}
      >
        <h1 style={{ fontSize: "3.5rem", color: "#111", marginBottom: "20px" }}>
          Next-Generation <span style={{ color: "#00ADD8" }}>Telemetry</span>
        </h1>
        <p style={{ fontSize: "1.2rem", color: "#666", lineHeight: "1.6" }}>
          Experience sub-millisecond ride matching powered by Golang, Redis, and
          Uber's H3 Hexagonal Grid mapping.
        </p>
      </div>

      <div
        style={{
          display: "flex",
          gap: "30px",
          flexWrap: "wrap",
          justifyContent: "center",
        }}
      >
        {/* Rider Card */}
        <Link to="/rider" style={{ textDecoration: "none" }}>
          <div
            style={{
              backgroundColor: "white",
              padding: "40px",
              borderRadius: "12px",
              boxShadow: "0 10px 30px rgba(0,0,0,0.05)",
              width: "300px",
              textAlign: "center",
              transition: "transform 0.2s",
            }}
          >
            <MapPin
              size={48}
              color="#00ADD8"
              style={{ marginBottom: "20px" }}
            />
            <h2 style={{ color: "#111", marginBottom: "15px" }}>
              I am a Rider
            </h2>
            <p style={{ color: "#666", marginBottom: "20px" }}>
              Open the interactive map to view live vehicles and dispatch a
              ride.
            </p>
            <span style={{ color: "#00ADD8", fontWeight: "bold" }}>
              Open Map &rarr;
            </span>
          </div>
        </Link>

        {/* Driver Card */}
        <Link to="/driver" style={{ textDecoration: "none" }}>
          <div
            style={{
              backgroundColor: "#111",
              padding: "40px",
              borderRadius: "12px",
              boxShadow: "0 10px 30px rgba(0,0,0,0.15)",
              width: "300px",
              textAlign: "center",
            }}
          >
            <Navigation
              size={48}
              color="#00ADD8"
              style={{ marginBottom: "20px" }}
            />
            <h2 style={{ color: "white", marginBottom: "15px" }}>
              I am a Driver
            </h2>
            <p style={{ color: "#999", marginBottom: "20px" }}>
              Connect to the WebSocket engine and stream your live GPS
              telemetry.
            </p>
            <span style={{ color: "#00ADD8", fontWeight: "bold" }}>
              Go Online &rarr;
            </span>
          </div>
        </Link>
      </div>
    </div>
  );
}
