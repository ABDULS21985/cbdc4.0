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

	// 1. Generate Wallet ID (e.g. UUID or Hash)
	walletID := "wallet-" + req.UserID // Simplified

	// 2. Call Fabric to register wallet on-chain
	_, err := s.fabric.SubmitTransaction("CreateWallet", walletID, req.UserID, "BankConsortiumMSP", req.Tier)
	if err != nil {
		log.Printf("Failed to create wallet on chain: %v", err)
		http.Error(w, "Failed to create wallet", http.StatusInternalServerError)
		return
	}

	// 3. Save metadata to local DB
	_, err = s.db.Exec("INSERT INTO wallet_db.wallets (id, user_id, address) VALUES ($1, $2, $3)", walletID, req.UserID, walletID)
	if err != nil {
		log.Printf("Failed to save wallet to DB: %v", err)
		http.Error(w, "Failed to save wallet metadata", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"wallet_id": walletID})
}

func (s *Service) GetWalletHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Call Fabric to get state
	result, err := s.fabric.EvaluateTransaction("GetWallet", id)
	if err != nil {
		http.Error(w, "Wallet not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
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
	json.NewEncoder(w).Encode(models.WalletBalance{Balance: wallet.Balance})
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
