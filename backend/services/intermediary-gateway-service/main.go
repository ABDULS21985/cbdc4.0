package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/centralbank/cbdc/backend/pkg/common"
	"github.com/gorilla/mux"
)

type KYCRequest struct {
	NationalID string `json:"national_id"`
	Name       string `json:"name"`
}

type GatewayService struct {
	client *http.Client
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

// WebhookDeliveryStub simulates sending a webhook to a bank
func (s *GatewayService) WebhookDeliveryStub(w http.ResponseWriter, r *http.Request) {
	// This endpoint simulates the Gateway *receiving* a trigger to send a webhook
	
	targetURL := "http://localhost:9090/bank/webhook" // In real world, look up from DB based on bank ID
	payload := map[string]string{
		"event": "payment_received",
		"tx_id": "tx-999",
		"amount": "100",
	}
	jsonPayload, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bank-API-Key", "secret")

	resp, err := s.client.Do(req)
	if err != nil {
		log.Printf("Failed to deliver webhook: %v", err)
		http.Error(w, "Webhook delivery failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	
	log.Printf("Webhook delivered to %s, status: %s", targetURL, resp.Status)
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "webhook_delivered"})
}

func main() {
	cfg := common.LoadConfig()
	svc := &GatewayService{
		client: &http.Client{Timeout: 10 * time.Second},
	}

	r := mux.NewRouter()

	// Middleware for API Key check could go here

	r.HandleFunc("/api/v1/onboard", svc.OnboardCustomerHandler).Methods("POST")
	r.HandleFunc("/api/v1/transfer", svc.InitiateTransferHandler).Methods("POST")
	r.HandleFunc("/api/v1/kyc/validate", ValidateKYCHandler).Methods("POST")
	r.HandleFunc("/internal/webhook/trigger", svc.WebhookDeliveryStub).Methods("POST")

	log.Printf("Intermediary Gateway running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
```
