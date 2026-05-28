import { Car } from "lucide-react";
import { Link } from "react-router-dom";

export default function Navbar() {
  return (
    <nav
      style={{
        backgroundColor: "#111",
        padding: "15px 30px",
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
      }}
    >
      <Link
        to="/"
        style={{
          color: "white",
          textDecoration: "none",
          fontSize: "24px",
          fontWeight: "bold",
          display: "flex",
          alignItems: "center",
          gap: "10px",
        }}
      >
        <Car color="#00ADD8" />
        RideSync
      </Link>
      <div style={{ display: "flex", gap: "20px" }}>
        <Link
          to="/rider"
          style={{ color: "white", textDecoration: "none", fontWeight: "500" }}
        >
          Rider Portal
        </Link>
        <Link
          to="/driver"
          style={{ color: "white", textDecoration: "none", fontWeight: "500" }}
        >
          Driver Portal
        </Link>
      </div>
    </nav>
  );
}
