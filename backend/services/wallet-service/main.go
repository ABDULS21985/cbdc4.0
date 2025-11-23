package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/centralbank/cbdc/backend/pkg/common"
	"github.com/centralbank/cbdc/backend/pkg/fabricclient"
	"github.com/gorilla/mux"
)

type CreateWalletRequest struct {
	UserID string `json:"user_id"`
	Tier   string `json:"tier"`
}

type Service struct {
	fabric *fabricclient.Client
}

func (s *Service) CreateWalletHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// 1. Generate Wallet ID (e.g. UUID or Hash)
	walletID := "wallet-" + req.UserID // Simplified

	// 2. Call Fabric to register wallet on-chain
	// Args: ID, OwnerID, IntermediaryID, Tier
	_, err := s.fabric.SubmitTransaction("CreateWallet", walletID, req.UserID, "BankConsortiumMSP", req.Tier)
	if err != nil {
		log.Printf("Failed to create wallet on chain: %v", err)
		http.Error(w, "Failed to create wallet", http.StatusInternalServerError)
		return
	}

	// 3. Save metadata to local DB (TODO)

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

func main() {
	cfg := common.LoadConfig()

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

	svc := &Service{fabric: fabric}

	r := mux.NewRouter()
	r.HandleFunc("/wallets", svc.CreateWalletHandler).Methods("POST")
	r.HandleFunc("/wallets/{id}", svc.GetWalletHandler).Methods("GET")

	log.Printf("Wallet Service running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
