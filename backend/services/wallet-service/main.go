package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/centralbank/cbdc/backend/pkg/common"
	"github.com/centralbank/cbdc/backend/pkg/common/api"
	"github.com/centralbank/cbdc/backend/pkg/common/db"
	"github.com/centralbank/cbdc/backend/pkg/common/migrations"
	"github.com/centralbank/cbdc/backend/pkg/fabricclient"
	"github.com/centralbank/cbdc/backend/services/wallet-service/models"
	"github.com/gorilla/mux"
)

type Service struct {
	fabric *fabricclient.Client
	db     *sql.DB
}

func (s *Service) CreateWalletHandler(w http.ResponseWriter, r *http.Request) {
	var req models.CreateWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	// Default values
	if req.Type == "" {
		req.Type = "RETAIL"
	}
	if req.Tier == "" {
		req.Tier = "TIER_1"
	}

	// 1. Generate Wallet ID
	walletID := "wallet-" + req.UserID

	// 2. Call Fabric to register wallet on-chain
	_, err := s.fabric.SubmitTransaction("CreateWallet", walletID, req.UserID, "BankConsortiumMSP", req.Tier)
	if err != nil {
		log.Printf("Failed to create wallet on chain: %v", err)
		api.WriteError(w, http.StatusInternalServerError, "chain_error", "Failed to create wallet on chain", "")
		return
	}

	// 3. Save metadata to local DB
	dailyLimit := int64(50000) // Default Tier 1
	if req.Tier == "TIER_2" {
		dailyLimit = 200000
	} else if req.Tier == "TIER_3" {
		dailyLimit = 5000000
	}

	_, err = s.db.Exec(`
		INSERT INTO wallet_db.wallets (
			id, user_id, address, type, status, currency, tier_level, daily_limit
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		walletID, req.UserID, walletID, req.Type, "ACTIVE", "NGN", req.Tier, dailyLimit)

	if err != nil {
		log.Printf("Failed to save wallet to DB: %v", err)
		api.WriteError(w, http.StatusInternalServerError, "db_error", "Failed to save wallet metadata", "")
		return
	}

	api.WriteSuccess(w, http.StatusCreated, map[string]string{"wallet_id": walletID, "status": "created"})
}

func (s *Service) GetWalletHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Get Metadata from DB
	var wallet models.Wallet
	err := s.db.QueryRow(`
		SELECT id, user_id, address, type, status, currency, tier_level, daily_limit, created_at 
		FROM wallet_db.wallets WHERE id = $1`, id).
		Scan(&wallet.ID, &wallet.UserID, &wallet.Address, &wallet.Type, &wallet.Status, &wallet.Currency, &wallet.TierLevel, &wallet.DailyLimit, &wallet.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			api.WriteError(w, http.StatusNotFound, "wallet_not_found", "Wallet not found", "")
		} else {
			log.Printf("DB Error: %v", err)
			api.WriteError(w, http.StatusInternalServerError, "internal_error", "Database error", "")
		}
		return
	}

	// Call Fabric to get Balance
	result, err := s.fabric.EvaluateTransaction("GetWallet", id)
	if err == nil {
		var chainWallet struct {
			Balance int64 `json:"balance"`
		}
		if err := json.Unmarshal(result, &chainWallet); err == nil {
			wallet.Balance = chainWallet.Balance
		}
	}

	api.WriteSuccess(w, http.StatusOK, wallet)
}

func (s *Service) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Call Fabric to get state
	result, err := s.fabric.EvaluateTransaction("GetWallet", id)
	if err != nil {
		api.WriteError(w, http.StatusNotFound, "wallet_not_found", "Wallet not found on chain", "")
		return
	}

	var wallet struct {
		Balance int64 `json:"balance"`
	}
	if err := json.Unmarshal(result, &wallet); err != nil {
		api.WriteError(w, http.StatusInternalServerError, "data_error", "Failed to parse chain data", "")
		return
	}

	api.WriteSuccess(w, http.StatusOK, models.WalletBalance{Balance: wallet.Balance, Currency: "NGN"})
}

func (s *Service) LockFundsHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		Amount int64  `json:"amount"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	// 1. Get Wallet ID
	walletID := "wallet-" + req.UserID

	// 2. Call Fabric to Transfer (Debit User, Credit "OfflineReserve" or Burn)
	// For this phase, we'll simulate locking by transferring to a "Reserve" wallet.
	// Assuming "ReserveWallet" exists or we just burn it.
	// Let's use a "Burn" or "Lock" chaincode method if available, or just Transfer to a known reserve.
	// We'll use "Transfer" to "offline-reserve-wallet" for now.
	reserveWallet := "offline-reserve-wallet"

	amountStr := fmt.Sprintf("%d", req.Amount)
	_, err := s.fabric.SubmitTransaction("Transfer", walletID, reserveWallet, amountStr)
	// Note: SubmitTransaction takes strings. string(req.Amount) is wrong, it converts rune. Need strconv.
	// But wait, my previous code used string(req.Amount) which is definitely a bug if Amount is int64.
	// I should fix that.

	if err != nil {
		log.Printf("Failed to lock funds: %v", err)
		api.WriteError(w, http.StatusInternalServerError, "chain_error", "Failed to lock funds on chain", "")
		return
	}

	api.WriteSuccess(w, http.StatusOK, map[string]string{"status": "locked", "tx_id": "simulated-tx-id"})
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
	if err := migrations.RunMigrations(database, "backend/migrations/wallet"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	fabric, err := fabricclient.NewClient(
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

	svc := &Service{fabric: fabric, db: database}

	r := mux.NewRouter()
	r.HandleFunc("/wallets", svc.CreateWalletHandler).Methods("POST")
	r.HandleFunc("/wallets/lock", svc.LockFundsHandler).Methods("POST") // New Endpoint
	r.HandleFunc("/wallets/{id}", svc.GetWalletHandler).Methods("GET")
	r.HandleFunc("/wallets/{id}/balance", svc.GetBalanceHandler).Methods("GET")

	log.Printf("Wallet Service running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
