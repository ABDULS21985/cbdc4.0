package main

import (
	"crypto/ed25519"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/centralbank/cbdc/backend/pkg/common"
	"github.com/centralbank/cbdc/backend/pkg/common/db"
	"github.com/centralbank/cbdc/backend/pkg/common/migrations"
	"github.com/centralbank/cbdc/backend/services/offline-service/models"
	"github.com/gorilla/mux"
)

type Service struct {
	db *sql.DB
}

func (s *Service) RegisterDeviceHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Store DeviceID <-> UserID mapping in DB
	deviceID := "dev-" + req.PublicKey[:8] // Simplified ID

	_, err := s.db.Exec(`
		INSERT INTO offline_db.devices (
			id, public_key, user_id, counter, hardware_id, os_version, trusted_status
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		deviceID, req.PublicKey, req.UserID, 0, req.HardwareID, req.OSVersion, "TRUSTED")

	if err != nil {
		log.Printf("Failed to register device: %v", err)
		http.Error(w, "Failed to register device", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "registered", "device_id": deviceID})
}

func (s *Service) RequestPurseHandler(w http.ResponseWriter, r *http.Request) {
	// Request a new offline purse (signed voucher)
	// In a real system, this would interact with the Central Bank's signing service

	// Create voucher record
	voucherID := "vouch-" + time.Now().Format("20060102150405")
	deviceID := "dev-mock" // Should come from auth context
	amount := int64(1000)
	signature := "mock-signature-from-cbn"
	expiresAt := time.Now().Add(24 * time.Hour) // 24h validity

	_, err := s.db.Exec(`
		INSERT INTO offline_db.vouchers (
			id, device_id, amount, status, signature, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		voucherID, deviceID, amount, "Active", signature, expiresAt)

	if err != nil {
		log.Printf("Failed to create voucher: %v", err)
		http.Error(w, "Failed to issue purse", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"purse_id":   voucherID,
		"limit":      amount,
		"signature":  signature,
		"expires_at": expiresAt,
	})
}

func (s *Service) ReconcileHandler(w http.ResponseWriter, r *http.Request) {
	var req models.ReconcileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	validCount := 0
	for _, tx := range req.Transactions {
		if verifySignature(tx) {
			// Check double spend (counter)
			var currentCounter int64
			err := s.db.QueryRow("SELECT counter FROM offline_db.devices WHERE public_key = $1", tx.From).Scan(&currentCounter)
			if err == nil && tx.Counter > currentCounter {
				// Valid sequence
				s.db.Exec("UPDATE offline_db.devices SET counter = $1, last_sync_at = $2 WHERE public_key = $3", tx.Counter, time.Now(), tx.From)
				validCount++
			} else {
				log.Printf("Replay detected or device not found: %s", tx.From)
			}
		} else {
			log.Printf("Invalid signature for tx from %s", tx.From)
		}
	}

	// TODO: Submit net settlement to Fabric via Payments Service or direct SDK

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      "processed",
		"valid_count": validCount,
	})
}

func verifySignature(tx models.OfflineTransaction) bool {
	pubKey, err := hex.DecodeString(tx.From)
	if err != nil || len(pubKey) != ed25519.PublicKeySize {
		return false
	}

	sig, err := hex.DecodeString(tx.Signature)
	if err != nil {
		return false
	}

	// Reconstruct message: From + To + Amount + Counter
	// Simplified serialization for demo
	msg := []byte(tx.From + tx.To + string(tx.Amount) + string(tx.Counter))

	return ed25519.Verify(pubKey, msg, sig)
}

func main() {
	cfg := common.LoadConfig()

	// Connect to DB
	database, err := db.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer database.Close()

	// Run Migrations
	if err := migrations.RunMigrations(database, "backend/migrations/offline"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	svc := &Service{db: database}

	r := mux.NewRouter()

	r.HandleFunc("/offline/devices", svc.RegisterDeviceHandler).Methods("POST")
	r.HandleFunc("/offline/purse", svc.RequestPurseHandler).Methods("POST")
	r.HandleFunc("/offline/reconcile", svc.ReconcileHandler).Methods("POST")

	log.Printf("Offline Service running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
