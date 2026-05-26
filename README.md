# RideSync

**High-Throughput Dispatch and Telemetry Engine**

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![Redis](https://img.shields.io/badge/Redis-7.0-DC382D?style=for-the-badge&logo=redis)
![Architecture](https://img.shields.io/badge/Architecture-Uber_H3-black?style=for-the-badge)

RideSync is a real-time geospatial telemetry and dispatch backend inspired by the core infrastructure of companies like Uber and Lyft. It is designed to ingest thousands of concurrent driver GPS coordinates and serve high-speed rider-matching queries with sub-millisecond latency.

---

## Architecture Overview

Instead of relying on heavy database-level geospatial math (e.g., PostGIS or Redis `GEOSEARCH`), RideSync brings compute into the highly concurrent Go application layer using spatial hashing.

**Uber H3 Hexagonal Grid**
Driver coordinates are algorithmically snapped into discrete Resolution 8 hexagons (~460 m wide) using the `uber/h3-go` library.

**O(1) Memory Lookups**
Instead of calculating Haversine distances against thousands of rows, the Go server translates locations into H3 Cell IDs and performs instant `SADD` and `SUNION` set operations in Redis.

**Persistent WebSockets**
Driver applications maintain a persistent TCP connection to the server via `gorilla/websocket`, streaming JSON telemetry with virtually zero network handshake overhead.

---

## Benchmarks

Load testing the read-path (`GET /api/v1/drivers/nearby`) on local hardware via `autocannon` with 100 concurrent connections:

| Metric      | Result                          |
|-------------|---------------------------------|
| Throughput  | ~39,500 Requests Per Second     |
| Avg Latency | 0.24 ms                         |
| Total Volume| 1.18 million requests / 30 sec  |

---

## Tech Stack

| Layer            | Technology                  |
|------------------|-----------------------------|
| Language         | Go 1.21+                    |
| API Routing      | `go-chi/chi/v5`             |
| Real-time        | `gorilla/websocket`         |
| Geospatial Engine| `uber/h3-go/v4`             |
| Caching Layer    | Redis 7 (`redis/go-redis/v9`)|
| Infrastructure   | Docker and Docker Compose   |

---

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose

### Installation

**1. Clone the repository:**

```bash
git clone https://github.com/aswinkmanoj/RideSync.git
cd RideSync
```

**2. Start the infrastructure (Redis and Postgres):**

```bash
docker-compose up -d
```

**3. Install dependencies and run the server:**

```bash
go mod tidy
go run cmd/ridesync/main.go
```

The server starts on port `:8080`.

---

## API Reference

### WebSocket Telemetry Ingestion (Driver)

Opens a persistent TCP tunnel to stream live GPS coordinates.

- **URL:** `ws://localhost:8080/api/v1/drivers/ws?id={driver_id}`

**Payload (JSON stream):**

```json
{
    "lat": 30.9080,
    "lng": 75.8500
}
```

---

### Find Nearby Drivers (Rider)

Calculates the H3 k-Ring disk (approximately 1.5 km radius) to instantly return nearby drivers and their exact distances.

- **Method:** `GET`
- **URL:** `/api/v1/drivers/nearby?lat=30.9050&lng=75.8500`

**Response:**

```json
[
  {
    "driver_id": "uber_001",
    "latitude": 30.908,
    "longitude": 75.85,
    "distance": 0.33
  }
]
```

---

### REST Telemetry Fallback (Driver)

Alternative HTTP endpoint for drivers unable to maintain a WebSocket connection.

- **Method:** `POST`
- **URL:** `/api/v1/drivers/update?id=uber_001&lat=30.9010&lng=75.8573`

---

## Roadmap

- [x] Phase 1: Setup high-performance HTTP router and Redis Docker environment
- [x] Phase 2: Implement Uber H3 spatial hashing and O(1) Redis caching
- [x] Phase 3: Build WebSocket ingestion engine for real-time telemetry
- [ ] Phase 4: Build Go Channel async workers for bulk PostgreSQL inserts (historical data)
- [ ] Phase 5: Build a React + Leaflet.js frontend dashboard to visualize live map traffic

---

## Contributing

Contributions, issues, and feature requests are welcome. Check the [issues page](https://github.com/aswinkmanoj/RideSync/issues) to get started.

---

## License

This project is licensed under the [MIT License](LICENSE).
