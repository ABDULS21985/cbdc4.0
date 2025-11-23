package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/centralbank/cbdc/backend/pkg/common"
	"github.com/gorilla/mux"
)

type KYCRequest struct {
	NationalID string `json:"national_id"`
	Name       string `json:"name"`
}

func ValidateKYCHandler(w http.ResponseWriter, r *http.Request) {
	var req KYCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Stub: Integrate with Legacy Bank System or National ID Database
	log.Printf("Validating KYC for ID: %s", req.NationalID)
	
	// Mock success
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "verified", "risk_level": "low"})
}

// GatewayService acts as a B2B API for banks
type GatewayService struct {
	// In a real app, this would hold clients to internal gRPC services
	// For now, we stub the logic or call the other services via HTTP
}

func (s *GatewayService) OnboardCustomerHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Validate Bank API Key (Middleware)
	// 2. Call Wallet Service to create wallet

	// Mock Response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "customer_onboarded",
		"wallet_id": "wallet-xyz-123",
	})
}

func (s *GatewayService) InitiateTransferHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Validate Bank API Key
	// 2. Call Payments Service

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "transfer_initiated",
		"tx_id":  "tx-999",
	})
}

func main() {
	cfg := common.LoadConfig()
	svc := &GatewayService{}

	r := mux.NewRouter()

	// Middleware for API Key check could go here

	r.HandleFunc("/api/v1/onboard", svc.OnboardCustomerHandler).Methods("POST")
	r.HandleFunc("/api/v1/transfer", svc.InitiateTransferHandler).Methods("POST")
	r.HandleFunc("/api/v1/kyc/validate", svc.ValidateKYCHandler).Methods("POST")

	log.Printf("Intermediary Gateway running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
```
