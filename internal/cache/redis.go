package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type DriverLocation struct {
	DriveID   string  `json:"driver_id"`
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

	fmt.Println("Connected to Redis Geospatial Index")
	return &RedisCache{client: client}, nil
}

func (r *RedisCache) GetNearbyDrivers(ctx context.Context, lat, lng, radiusKm float64) ([]DriverLocation, error) {
	query := &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude:  lng,
			Latitude:   lat,
			Radius:     radiusKm,
			RadiusUnit: "km",
			Sort:       "ASC",
		},
		WithCoord: true,
		WithDist:  true,
	}

	results, err := r.client.GeoSearchLocation(ctx, "drivers:locations", query).Result()
	if err != nil {
		return nil, fmt.Errorf("geosearch failder: %w", err)
	}

	drivers := make([]DriverLocation, len(results))
	for i, res := range results {
		drivers[i] = DriverLocation{
			DriveID:   res.Name,
			Latitude:  res.Latitude,
			Longitude: res.Longitude,
			Distance:  res.Dist,
		}
	}
	return drivers, nil
}
