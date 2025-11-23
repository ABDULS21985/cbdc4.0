package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/centralbank/cbdc/backend/pkg/common"
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
		http.Error(w, "Invalid request", http.StatusBadRequest)
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
		http.Error(w, "Failed to create wallet", http.StatusInternalServerError)
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
		http.Error(w, "Failed to save wallet metadata", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"wallet_id": walletID, "status": "created"})
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
			http.Error(w, "Wallet not found", http.StatusNotFound)
		} else {
			log.Printf("DB Error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallet)
}

func (s *Service) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Call Fabric to get state
	result, err := s.fabric.EvaluateTransaction("GetWallet", id)
	if err != nil {
		http.Error(w, "Wallet not found", http.StatusNotFound)
		return
	}

	var wallet struct {
		Balance int64 `json:"balance"`
	}
	if err := json.Unmarshal(result, &wallet); err != nil {
		http.Error(w, "Failed to parse wallet data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models.WalletBalance{Balance: wallet.Balance, Currency: "NGN"})
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
	r.HandleFunc("/wallets/{id}", svc.GetWalletHandler).Methods("GET")
	r.HandleFunc("/wallets/{id}/balance", svc.GetBalanceHandler).Methods("GET")

	log.Printf("Wallet Service running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
