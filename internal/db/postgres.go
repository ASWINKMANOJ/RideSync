package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type DBLocation struct {
	DriverID    string
	Latitude    float64
	Longitude   float64
	VehicleType string
}

type PostgresStore struct {
	DB           *sql.DB
	LocationChan chan DBLocation
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	fmt.Println("Connected to PostgresSQL")

	store := &PostgresStore{
		DB:           db,
		LocationChan: make(chan DBLocation, 1000),
	}

	go store.startBatchWorker()

	return store, nil
}

func (p *PostgresStore) startBatchWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var batch []DBLocation

	for {
		select {
		case loc := <-p.LocationChan:
			batch = append(batch, loc)
			if len(batch) >= 500 {
				p.executeBulkInsert(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				p.executeBulkInsert(batch)
				batch = nil
			}
		}
	}
}

func (p *PostgresStore) executeBulkInsert(batch []DBLocation) {
	if len(batch) == 0 {
		return
	}

	valueStrings := make([]string, 0, len(batch))
	valueArgs := make([]interface{}, 0, len(batch)*4)

	i := 1
	for _, loc := range batch {
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d)", i, i+1, i+2, i+3))
		valueArgs = append(valueArgs, loc.DriverID, loc.Latitude, loc.Longitude, loc.VehicleType)
		i += 4
	}

	stmt := fmt.Sprintf("INSERT INTO driver_locations (driver_id, latitude, longitude, vehicle_type) VALUES %s", strings.Join(valueStrings, ","))
	_, err := p.DB.Exec(stmt, valueArgs...)
	if err != nil {
		log.Printf("PostgreSQL Bulk Insert Failed: %v", err)
		return
	}

	log.Printf("Successfully saved batch of %d locations to PostgreSQL", len(batch))
}
