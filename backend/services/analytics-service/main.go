package main

import (
	"log"
	"net/http"
	"time"

	"github.com/centralbank/cbdc/backend/pkg/common"
)

// This service represents Layer 5: Data, Risk & Analytics
// It would listen to Fabric block events and index them into a Data Warehouse (e.g., Snowflake, BigQuery)
// For this prototype, it logs events to stdout.

func main() {
	cfg := common.LoadConfig()

	// Mock Event Listener
	go func() {
		for {
			time.Sleep(10 * time.Second)
			log.Println("[Analytics] Ingesting block data... No anomalies detected.")
		}
	}()

	log.Printf("Analytics Service (Data Warehouse) running on :%s", "8084")
	// Simple health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Analytics OK"))
	})

	log.Fatal(http.ListenAndServe(":8084", nil))
}
