package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/centralbank/cbdc/backend/pkg/common"
	_ "github.com/lib/pq" // Postgres driver
)

// Connect establishes a connection to the database
func Connect(cfg common.DBConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open db connection: %v", err)
	}

	// Retry logic for waiting for DB to be ready
	for i := 0; i < 5; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		log.Printf("Waiting for DB... (%d/5): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to ping db: %v", err)
	}

	log.Println("Successfully connected to database")
	return db, nil
}
