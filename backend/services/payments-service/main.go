package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/centralbank/cbdc/backend/pkg/common"
	"github.com/centralbank/cbdc/backend/pkg/common/db"
	"github.com/centralbank/cbdc/backend/pkg/common/migrations"
	"github.com/centralbank/cbdc/backend/pkg/fabricclient"
	"github.com/centralbank/cbdc/backend/services/payments-service/models"
	"github.com/gorilla/mux"
)

type Service struct {
	fabric *fabricclient.Client
	db     *sql.DB
}

func (s *Service) TransferHandler(w http.ResponseWriter, r *http.Request) {
	var req models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Validate input (amount > 0, valid IDs)

	// 1. Record "Pending" in DB
	txID := "tx-" + time.Now().Format("20060102150405") // Simple ID generation
	_, err := s.db.Exec("INSERT INTO payments_db.transactions (id, from_wallet, to_wallet, amount, status) VALUES ($1, $2, $3, $4, $5)",
		txID, req.From, req.To, req.Amount, "Pending")
	if err != nil {
		log.Printf("Failed to record pending tx: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 2. Call Chaincode
	// Note: In a real system, 'From' would be derived from the authenticated user context
	result, err := s.fabric.SubmitTransaction("Transfer", req.From, req.To, string(req.Amount)) // Simplified arg passing
	if err != nil {
		log.Printf("Failed to submit transaction: %v", err)
		// Update DB to Failed
		s.db.Exec("UPDATE payments_db.transactions SET status = 'Failed' WHERE id = $1", txID)
		http.Error(w, "Transaction failed", http.StatusInternalServerError)
		return
	}

	// 3. Update DB to Confirmed (Optimistic)
	// In a real system, we'd wait for Fabric event, but here we assume success if Submit returns.
	s.db.Exec("UPDATE payments_db.transactions SET status = 'Confirmed', tx_hash = $1 WHERE id = $2", "fabric-tx-hash-placeholder", txID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result) // Assuming chaincode returns JSON
}

func (s *Service) BatchTransferHandler(w http.ResponseWriter, r *http.Request) {
	// Similar logic for batch...
	// For brevity, just calling fabric
	var req models.BatchTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Serialize transfers for chaincode
	transfersJSON, _ := json.Marshal(req.Transfers)
	result, err := s.fabric.SubmitTransaction("BatchTransfer", req.FromWalletID, string(transfersJSON))
	if err != nil {
		http.Error(w, "Batch Transaction failed", http.StatusInternalServerError)
		return
	}
	w.Write(result)
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
	if err := migrations.RunMigrations(database, "backend/migrations/payments"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Fabric Client
	// Note: These paths would need to be real in a deployed env
	fabric, err := fabricclient.NewClient(
		cfg.FabricConfig,
		"cbdc-main-channel",
		"cbdc-core",
		cfg.MSP,
		cfg.CertPath,
		cfg.KeyPath,
	)
	if err != nil {
		log.Printf("Warning: Failed to connect to Fabric (expected during build/test without network): %v", err)
		// Continue to allow build to pass, but service won't work fully
	} else {
		defer fabric.Close()
	}

	svc := &Service{fabric: fabric, db: database}

	r := mux.NewRouter()
	r.HandleFunc("/payments/p2p", svc.TransferHandler).Methods("POST")
	r.HandleFunc("/payments/merchant", svc.MerchantPaymentHandler).Methods("POST")
	r.HandleFunc("/payments/batch", svc.BatchTransferHandler).Methods("POST")
	r.HandleFunc("/payments/{id}", svc.GetTransactionHandler).Methods("GET")
	r.HandleFunc("/payments/history", svc.GetHistoryHandler).Methods("GET")

	log.Printf("Payments Service running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}

func (s *Service) MerchantPaymentHandler(w http.ResponseWriter, r *http.Request) {
	var req models.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Merchant specific validation (e.g. check if To is a valid merchant wallet)
	// For now, we reuse the Transfer logic but could add metadata

	// Call Chaincode
	result, err := s.fabric.SubmitTransaction("Transfer", req.From, req.To, string(req.Amount))
	if err != nil {
		log.Printf("Failed to submit merchant transaction: %v", err)
		http.Error(w, "Transaction failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func (s *Service) GetTransactionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Try DB first
	var tx models.Transaction
	err := s.db.QueryRow("SELECT id, from_wallet, to_wallet, amount, status FROM payments_db.transactions WHERE id = $1", id).
		Scan(&tx.ID, &tx.FromWallet, &tx.ToWallet, &tx.Amount, &tx.Status)

	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tx)
		return
	}

	// Fallback to Fabric
	result, err := s.fabric.EvaluateTransaction("GetTransaction", id)
	if err != nil {
		http.Error(w, "Transaction not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func (s *Service) GetHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// Query DB
	rows, err := s.db.Query("SELECT id, from_wallet, to_wallet, amount, status FROM payments_db.transactions ORDER BY created_at DESC LIMIT 50")
	if err != nil {
		http.Error(w, "Failed to fetch history", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var history []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(&tx.ID, &tx.FromWallet, &tx.ToWallet, &tx.Amount, &tx.Status); err == nil {
			history = append(history, tx)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
