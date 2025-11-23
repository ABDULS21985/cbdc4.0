package main

import (
	"bytes"
	"crypto/ed25519"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/centralbank/cbdc/backend/pkg/common"
	"github.com/centralbank/cbdc/backend/pkg/common/api"
	"github.com/centralbank/cbdc/backend/pkg/common/db"
	"github.com/centralbank/cbdc/backend/pkg/common/migrations"
	"github.com/centralbank/cbdc/backend/pkg/fabricclient"
	"github.com/centralbank/cbdc/backend/services/offline-service/models"
	"github.com/gorilla/mux"
)

// Risk control constants from Phase 7 design
const (
	MaxOfflineBalance     = 500 // $500 max offline balance
	MaxTransactionAmount  = 50  // $50 max per transaction
	SyncTTLDays           = 7   // 7 days max before sync required
)

type Service struct {
	db               *sql.DB
	fabric           *fabricclient.Client
	ganache          *GanacheClient
	walletServiceURL string
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

	// Initialize Fabric client
	var fabric *fabricclient.Client
	fabric, err = fabricclient.NewClient(
		cfg.FabricConfig,
		"cbdc-main-channel",
		"cbdc-core",
		cfg.MSP,
		cfg.CertPath,
		cfg.KeyPath,
	)
	if err != nil {
		log.Printf("Warning: Fabric connection failed: %v", err)
	} else {
		defer fabric.Close()
	}

	// Initialize Ganache client for offline voucher verification
	ganacheURL := os.Getenv("GANACHE_URL")
	if ganacheURL == "" {
		ganacheURL = "http://localhost:8545"
	}
	ganache := NewGanacheClient(ganacheURL)

	// Wallet service URL for fund locking
	walletServiceURL := os.Getenv("WALLET_SERVICE_URL")
	if walletServiceURL == "" {
		walletServiceURL = "http://localhost:8082"
	}

	svc := &Service{
		db:               database,
		fabric:           fabric,
		ganache:          ganache,
		walletServiceURL: walletServiceURL,
	}

	r := mux.NewRouter()
	r.HandleFunc("/offline/device", svc.RegisterDeviceHandler).Methods("POST")
	r.HandleFunc("/offline/fund", svc.FundPurseHandler).Methods("POST")
	r.HandleFunc("/offline/reconcile", svc.ReconcileHandler).Methods("POST")
	r.HandleFunc("/offline/purse/{deviceId}", svc.GetPurseHandler).Methods("GET")
	r.HandleFunc("/health", svc.HealthHandler).Methods("GET")

	log.Printf("Offline Service running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}

func (s *Service) HealthHandler(w http.ResponseWriter, r *http.Request) {
	api.WriteSuccess(w, http.StatusOK, map[string]string{"status": "healthy", "service": "offline-service"})
}

func (s *Service) RegisterDeviceHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	// Validate public key format (should be hex-encoded Ed25519 public key)
	if len(req.PublicKey) != 64 { // 32 bytes = 64 hex chars
		api.WriteError(w, http.StatusBadRequest, "invalid_public_key", "Public key must be 64 hex characters", "")
		return
	}

	// Store DeviceID <-> UserID mapping in DB
	deviceID := "dev-" + req.PublicKey[:8] // Simplified ID

	_, err := s.db.Exec(`
		INSERT INTO offline_db.devices (
			id, public_key, user_id, counter, hardware_id, os_version, trusted_status, last_sync_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		deviceID, req.PublicKey, req.UserID, 0, req.HardwareID, req.OSVersion, "TRUSTED", time.Now())

	if err != nil {
		log.Printf("Failed to register device: %v", err)
		api.WriteError(w, http.StatusInternalServerError, "db_error", "Failed to register device", "")
		return
	}

	api.WriteSuccess(w, http.StatusCreated, map[string]string{"status": "registered", "device_id": deviceID})
}

func (s *Service) FundPurseHandler(w http.ResponseWriter, r *http.Request) {
	var req models.FundPurseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	// Validate amount against balance cap
	var currentBalance int64
	err := s.db.QueryRow("SELECT COALESCE(balance, 0) FROM offline_db.purses WHERE device_id = $1", req.DeviceID).Scan(&currentBalance)
	if err != nil && err != sql.ErrNoRows {
		api.WriteError(w, http.StatusInternalServerError, "db_error", "Failed to query purse", "")
		return
	}

	if currentBalance+req.Amount > MaxOfflineBalance {
		api.WriteError(w, http.StatusBadRequest, "balance_limit_exceeded",
			fmt.Sprintf("Would exceed max offline balance of %d", MaxOfflineBalance), "")
		return
	}

	// 1. Call Wallet Service to Lock Funds (Real HTTP Call)
	lockReq := map[string]interface{}{
		"user_id": req.UserID,
		"amount":  req.Amount,
		"reason":  "offline_funding",
	}
	lockBody, _ := json.Marshal(lockReq)

	resp, err := http.Post(s.walletServiceURL+"/wallets/lock", "application/json", bytes.NewBuffer(lockBody))
	if err != nil {
		log.Printf("Failed to call wallet service: %v", err)
		api.WriteError(w, http.StatusInternalServerError, "upstream_error", "Failed to contact wallet service", "")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Wallet service returned error: %d", resp.StatusCode)
		api.WriteError(w, http.StatusPaymentRequired, "insufficient_funds", "Failed to lock funds", "")
		return
	}

	// 2. Update Offline Purse Balance (Shadow)
	if err == sql.ErrNoRows {
		// Create purse
		_, err = s.db.Exec(`INSERT INTO offline_db.purses (device_id, user_id, balance, counter, last_sync_at, status)
			VALUES ($1, $2, $3, 0, $4, 'ACTIVE')`, req.DeviceID, req.UserID, req.Amount, time.Now())
	} else {
		// Update purse
		_, err = s.db.Exec(`UPDATE offline_db.purses SET balance = balance + $1, last_sync_at = $2 WHERE device_id = $3`,
			req.Amount, time.Now(), req.DeviceID)
	}

	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "db_error", "Failed to update purse", "")
		return
	}

	// 3. Generate Signed Certificate (PurseUpdate) using Real Crypto
	// In production, load private key from HSM/Vault. Using static seed for demo.
	seed := make([]byte, ed25519.SeedSize)
	privateKey := ed25519.NewKeyFromSeed(seed)

	// Message to sign: DeviceID + Amount + Timestamp
	timestamp := time.Now().Unix()
	msg := fmt.Sprintf("%s:%d:%d", req.DeviceID, req.Amount, timestamp)
	signature := hex.EncodeToString(ed25519.Sign(privateKey, []byte(msg)))

	// Also verify on Ganache if available (for prototyping)
	if s.ganache != nil {
		go func() {
			if err := s.ganache.VerifyOfflineFunding(req.DeviceID, req.Amount, signature); err != nil {
				log.Printf("Ganache verification failed (non-blocking): %v", err)
			}
		}()
	}

	api.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"status":    "funded",
		"device_id": req.DeviceID,
		"amount":    req.Amount,
		"signature": signature,
		"timestamp": timestamp,
	})
}

func (s *Service) GetPurseHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	deviceID := vars["deviceId"]

	var purse models.OfflinePurse
	err := s.db.QueryRow(`
		SELECT device_id, user_id, balance, counter, COALESCE(last_sync_hash, ''), last_sync_at, status
		FROM offline_db.purses WHERE device_id = $1`, deviceID).Scan(
		&purse.DeviceID, &purse.UserID, &purse.Balance, &purse.Counter,
		&purse.LastSyncHash, &purse.LastSyncAt, &purse.Status)

	if err == sql.ErrNoRows {
		api.WriteError(w, http.StatusNotFound, "not_found", "Purse not found", "")
		return
	}
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "db_error", "Failed to query purse", "")
		return
	}

	// Check if locked due to TTL
	if time.Since(purse.LastSyncAt) > time.Duration(SyncTTLDays)*24*time.Hour {
		purse.Status = "LOCKED_TTL"
	}

	api.WriteSuccess(w, http.StatusOK, purse)
}

func (s *Service) ReconcileHandler(w http.ResponseWriter, r *http.Request) {
	var req models.ReconcileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	validCount := 0
	failedCount := 0
	validProofs := []map[string]interface{}{}
	failedReasons := []string{}

	for _, tx := range req.Transactions {
		reason := s.processTransaction(tx, &validProofs)
		if reason != "" {
			failedCount++
			failedReasons = append(failedReasons, reason)
		} else {
			validCount++
		}
	}

	// Submit Batch to Fabric using BatchReconcile
	if len(validProofs) > 0 && s.fabric != nil {
		proofsJSON, _ := json.Marshal(validProofs)
		log.Printf("Submitting batch to Fabric: %s", proofsJSON)
		_, err := s.fabric.SubmitTransaction("BatchReconcile", string(proofsJSON))
		if err != nil {
			log.Printf("Failed to submit batch to Fabric: %v", err)
			// Queue for retry in production
		}
	}

	// Update last sync time for the submitting device
	s.db.Exec("UPDATE offline_db.purses SET last_sync_at = $1 WHERE device_id = $2", time.Now(), req.DeviceID)

	api.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"status":         "processed",
		"valid_count":    validCount,
		"failed_count":   failedCount,
		"batch_size":     len(validProofs),
		"failed_reasons": failedReasons,
	})
}

func (s *Service) processTransaction(tx models.SignedPayment, validProofs *[]map[string]interface{}) string {
	// 1. Verify Signature (Real Ed25519)
	if tx.Signature == "" {
		log.Printf("Missing signature for tx from %s", tx.PayerID)
		return fmt.Sprintf("missing_signature:%s", tx.PayerID)
	}

	// Fetch Public Key
	var publicKeyHex string
	err := s.db.QueryRow("SELECT public_key FROM offline_db.devices WHERE id = $1", tx.PayerID).Scan(&publicKeyHex)
	if err != nil {
		log.Printf("Device not found or DB error: %s", tx.PayerID)
		return fmt.Sprintf("device_not_found:%s", tx.PayerID)
	}

	// Verify signature
	if !verifySignature(tx, publicKeyHex) {
		log.Printf("Invalid signature for tx from %s", tx.PayerID)
		return fmt.Sprintf("invalid_signature:%s", tx.PayerID)
	}

	// Optional: Verify on Ganache (ecrecover equivalent)
	if s.ganache != nil {
		if err := s.ganache.VerifyOfflineTransaction(tx); err != nil {
			log.Printf("Ganache verification failed: %v", err)
			// Non-blocking for now, Ed25519 is primary
		}
	}

	// 2. Risk Controls
	// 2a. Transaction Limit
	if tx.Amount > MaxTransactionAmount {
		log.Printf("Transaction amount %d exceeds limit %d", tx.Amount, MaxTransactionAmount)
		return fmt.Sprintf("amount_exceeded:%d", tx.Amount)
	}

	// 2b. TTL Check (7 Days)
	var lastSyncAt time.Time
	var currentBalance int64
	err = s.db.QueryRow("SELECT balance, last_sync_at FROM offline_db.purses WHERE device_id = $1", tx.PayerID).
		Scan(&currentBalance, &lastSyncAt)
	if err != nil {
		log.Printf("Purse not found for %s", tx.PayerID)
		return fmt.Sprintf("purse_not_found:%s", tx.PayerID)
	}

	if time.Since(lastSyncAt) > time.Duration(SyncTTLDays)*24*time.Hour {
		log.Printf("Device %s has not synced in %d days. Transaction rejected.", tx.PayerID, SyncTTLDays)
		return fmt.Sprintf("ttl_expired:%s", tx.PayerID)
	}

	// 2c. Balance check
	if currentBalance < tx.Amount {
		log.Printf("Insufficient shadow balance for %s: has %d, needs %d", tx.PayerID, currentBalance, tx.Amount)
		return fmt.Sprintf("insufficient_balance:%s", tx.PayerID)
	}

	// 2d. Double Spend (Used Counters)
	var exists bool
	err = s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM offline_db.used_counters WHERE device_id = $1 AND counter = $2)",
		tx.PayerID, tx.Counter).Scan(&exists)
	if err != nil {
		log.Printf("DB Error checking counter: %v", err)
		return "db_error:counter_check"
	}
	if exists {
		log.Printf("FRAUD ALERT: Double spend detected! Device: %s, Counter: %d", tx.PayerID, tx.Counter)
		// TODO: Flag account for investigation
		return fmt.Sprintf("double_spend:%s:%d", tx.PayerID, tx.Counter)
	}

	// 3. Process Transaction
	// Insert into used_counters
	_, err = s.db.Exec("INSERT INTO offline_db.used_counters (device_id, counter, tx_hash, created_at) VALUES ($1, $2, $3, $4)",
		tx.PayerID, tx.Counter, "hash-"+tx.Signature[:16], time.Now())
	if err != nil {
		log.Printf("Failed to record counter: %v", err)
		return "db_error:record_counter"
	}

	// Debit Payer Shadow Balance
	_, err = s.db.Exec("UPDATE offline_db.purses SET balance = balance - $1 WHERE device_id = $2", tx.Amount, tx.PayerID)
	if err != nil {
		log.Printf("Failed to debit shadow balance: %v", err)
	}

	// Map DeviceID to WalletID
	var payerUserID string
	s.db.QueryRow("SELECT user_id FROM offline_db.devices WHERE id = $1", tx.PayerID).Scan(&payerUserID)
	payerWalletID := "wallet-" + payerUserID

	var payeeUserID string
	err = s.db.QueryRow("SELECT user_id FROM offline_db.devices WHERE id = $1", tx.PayeeID).Scan(&payeeUserID)
	if err != nil {
		payeeUserID = tx.PayeeID // Fallback if PayeeID is already a UserID
	}
	payeeWalletID := "wallet-" + payeeUserID

	proof := map[string]interface{}{
		"from":      payerWalletID,
		"to":        payeeWalletID,
		"amount":    tx.Amount,
		"nonce":     tx.Counter,
		"signature": tx.Signature,
	}
	*validProofs = append(*validProofs, proof)

	// Credit payee's wallet in local DB (optimistic)
	_, err = s.db.Exec("UPDATE wallet_db.wallets SET balance = balance + $1 WHERE id = $2", tx.Amount, payeeWalletID)
	if err != nil {
		log.Printf("Failed to update local wallet balance: %v", err)
	}

	return "" // Success
}

func verifySignature(tx models.SignedPayment, pubKeyHex string) bool {
	pubKey, err := hex.DecodeString(pubKeyHex)
	if err != nil || len(pubKey) != ed25519.PublicKeySize {
		return false
	}

	sig, err := hex.DecodeString(tx.Signature)
	if err != nil {
		return false
	}

	// The Intent JSON string is what was signed
	msg := []byte(tx.Intent)

	return ed25519.Verify(pubKey, msg, sig)
}
