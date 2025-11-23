package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/centralbank/cbdc/backend/pkg/common"
	"github.com/gorilla/mux"
)

// This service represents the integration with an external Instant Payment System (IPS)
// It allows CBDC -> Commercial Bank Money transfers via the IPS rail.

type IPSRequest struct {
	Amount          int64  `json:"amount"`
	DestinationIBAN string `json:"destination_iban"`
}

func InitiateOutboundTransfer(w http.ResponseWriter, r *http.Request) {
	var req IPSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Printf("Initiating IPS Transfer: %d to %s", req.Amount, req.DestinationIBAN)

	// Stub: Simulate successful settlement on external rail
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "settled", "ips_ref": "IPS-999888777"})
}

func main() {
	cfg := common.LoadConfig()
	r := mux.NewRouter()

	r.HandleFunc("/ips/outbound", InitiateOutboundTransfer).Methods("POST")

	log.Printf("IPS Adapter Service running on :%s", "8085")
	log.Fatal(http.ListenAndServe(":8085", r))
}
