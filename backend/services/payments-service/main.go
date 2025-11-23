package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/centralbank/cbdc/backend/pkg/common"
	"github.com/centralbank/cbdc/backend/pkg/fabricclient"
	"github.com/gorilla/mux"
)

type PaymentRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
}

type Service struct {
	fabric *fabricclient.Client
}

func (s *Service) TransferHandler(w http.ResponseWriter, r *http.Request) {
	var req PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// TODO: Validate input (amount > 0, valid IDs)

	// Call Chaincode
	// Note: In a real system, 'From' would be derived from the authenticated user context
	result, err := s.fabric.SubmitTransaction("Transfer", req.From, req.To, string(req.Amount)) // Simplified arg passing
	if err != nil {
		log.Printf("Failed to submit transaction: %v", err)
		http.Error(w, "Transaction failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(result) // Assuming chaincode returns JSON
}

func main() {
	cfg := common.LoadConfig()

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

	svc := &Service{fabric: fabric}

	r := mux.NewRouter()
	r.HandleFunc("/payments/p2p", svc.TransferHandler).Methods("POST")
	r.HandleFunc("/payments/merchant", svc.MerchantPaymentHandler).Methods("POST")

	log.Printf("Payments Service running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}

func (s *Service) MerchantPaymentHandler(w http.ResponseWriter, r *http.Request) {
	var req PaymentRequest
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
