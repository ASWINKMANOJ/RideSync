import { Route, BrowserRouter as Router, Routes } from "react-router-dom";
import Navbar from "./components/Navbar";
import DriverApp from "./pages/DriverApp";
import Home from "./pages/Home";
import RiderMap from "./pages/RiderMap";

function App() {
  return (
    <Router>
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          height: "100vh",
          fontFamily: "sans-serif",
        }}
      >
        <Navbar />
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/rider" element={<RiderMap />} />
          <Route path="/driver" element={<DriverApp />} />
        </Routes>
      </div>
    </Router>
  );
}

export default App;
