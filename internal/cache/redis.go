package cache

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/uber/h3-go/v4"
)

type DriverLocation struct {
	DriverID  string  `json:"driver_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Distance  float64 `json:"distance"`
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr string) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	fmt.Println("Connected to Redis(H3 Spatial Hashing Mode)")
	return &RedisCache{client: client}, nil
}

func (r *RedisCache) SaveDriverLocation(ctx context.Context, driverId string, lat, lng float64) error {
	latLng := h3.NewLatLng(lat, lng)
	cell, err := h3.LatLngToCell(latLng, 8)
	if err != nil {
		return fmt.Errorf("invalid coordinates for h3: %w", err)
	}

	tileKey := fmt.Sprintf("tile:%s", cell.String())

	driverData := fmt.Sprintf("%s|%f|%f", driverId, lat, lng)

	pipe := r.client.Pipeline()
	pipe.SAdd(ctx, tileKey, driverData)
	pipe.Expire(ctx, tileKey, 60*time.Second)
	_, execErr := pipe.Exec(ctx)

	return execErr
}

func (r *RedisCache) GetNearbyDrivers(ctx context.Context, lat, lng float64) ([]DriverLocation, error) {
	riderLatLng := h3.NewLatLng(lat, lng)
	originCell, err := h3.LatLngToCell(riderLatLng, 8)

	if err != nil {
		return nil, fmt.Errorf("invalid rider coordinates: %w", err)
	}

	cells, err := originCell.GridDisk(1)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate surrounding grid: %w", err)
	}

	var tileKeys []string
	for _, cell := range cells {
		tileKeys = append(tileKeys, fmt.Sprintf("tile:%s", cell.String()))
	}

	rawDrivers, err := r.client.SUnion(ctx, tileKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tiles: %w", err)
	}

	var drivers []DriverLocation
	for _, raw := range rawDrivers {
		parts := strings.Split(raw, "|")
		if len(parts) != 3 {
			continue
		}

		dLat, _ := strconv.ParseFloat(parts[1], 64)
		dLng, _ := strconv.ParseFloat(parts[2], 64)
		distance := haversine(lat, lng, dLat, dLng)

		drivers = append(drivers, DriverLocation{
			DriverID:  parts[0],
			Latitude:  dLat,
			Longitude: dLng,
			Distance:  math.Round(distance*100) / 100, // Round to 2 decimal places
		})
	}
	if drivers == nil {
		drivers = []DriverLocation{}
	}

	return drivers, nil
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0 // Earth radius in kilometers

	// Convert degrees to radians
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)
	lat1Rad := lat1 * (math.Pi / 180.0)
	lat2Rad := lat2 * (math.Pi / 180.0)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Sin(dLon/2)*math.Sin(dLon/2)*math.Cos(lat1Rad)*math.Cos(lat2Rad)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
