package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/centralbank/cbdc/backend/pkg/common"
	"github.com/gorilla/mux"
)

type RegisterDeviceRequest struct {
	UserID    string `json:"user_id"`
	PublicKey string `json:"public_key"` // Hex encoded Ed25519
}

type OfflineTransaction struct {
	From      string `json:"from"` // Device Public Key
	To        string `json:"to"`   // Device Public Key
	Amount    int64  `json:"amount"`
	Counter   int64  `json:"counter"`
	Signature string `json:"signature"` // Hex encoded
}

type ReconcileRequest struct {
	Transactions []OfflineTransaction `json:"transactions"`
}

func RegisterDeviceHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// TODO: Verify UserID exists and authorize
	// TODO: Store DeviceID <-> UserID mapping in DB

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "registered"})
}

func ReconcileHandler(w http.ResponseWriter, r *http.Request) {
	var req ReconcileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	validCount := 0
	for _, tx := range req.Transactions {
		if verifySignature(tx) {
			// TODO: Check double spend (counter)
			// TODO: Aggregate net changes
			validCount++
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

func verifySignature(tx OfflineTransaction) bool {
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
	r := mux.NewRouter()

	r.HandleFunc("/offline/devices", RegisterDeviceHandler).Methods("POST")
	r.HandleFunc("/offline/reconcile", ReconcileHandler).Methods("POST")

	log.Printf("Offline Service running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
