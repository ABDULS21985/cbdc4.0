package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/centralbank/cbdc/backend/pkg/common"
	"github.com/centralbank/cbdc/backend/pkg/common/api"
	"github.com/centralbank/cbdc/backend/pkg/common/db"
	"github.com/centralbank/cbdc/backend/pkg/fabricclient"
	"github.com/gorilla/mux"
)

// Service represents the RTGS Adapter Service
// As per Phase 3 design: Simulates settlement with the Real-Time Gross Settlement system
type Service struct {
	db     *sql.DB
	fabric *fabricclient.Client
}

// SettlementRequest represents a request to settle with RTGS
type SettlementRequest struct {
	FromBankID string `json:"from_bank_id"`
	ToBankID   string `json:"to_bank_id"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
	Reference  string `json:"reference"`
}

// SettlementResponse represents the response from RTGS
type SettlementResponse struct {
	SettlementID string    `json:"settlement_id"`
	Status       string    `json:"status"`
	FromBankID   string    `json:"from_bank_id"`
	ToBankID     string    `json:"to_bank_id"`
	Amount       int64     `json:"amount"`
	Currency     string    `json:"currency"`
	SettledAt    time.Time `json:"settled_at"`
}

// LiquidityRequest represents a request for liquidity from RTGS
type LiquidityRequest struct {
	BankID   string `json:"bank_id"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

// Settlement represents a settlement record
type Settlement struct {
	ID          string    `json:"id"`
	FromBankID  string    `json:"from_bank_id"`
	ToBankID    string    `json:"to_bank_id"`
	Amount      int64     `json:"amount"`
	Currency    string    `json:"currency"`
	Status      string    `json:"status"`
	Reference   string    `json:"reference"`
	CreatedAt   time.Time `json:"created_at"`
	SettledAt   time.Time `json:"settled_at,omitempty"`
}

func main() {
	cfg := common.LoadConfig()

	// Connect to DB
	database, err := db.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer database.Close()

	// Create RTGS schema if not exists
	_, err = database.Exec(`
		CREATE SCHEMA IF NOT EXISTS rtgs_db;
		CREATE TABLE IF NOT EXISTS rtgs_db.settlements (
			id VARCHAR(255) PRIMARY KEY,
			from_bank_id VARCHAR(255) NOT NULL,
			to_bank_id VARCHAR(255) NOT NULL,
			amount BIGINT NOT NULL,
			currency VARCHAR(10) DEFAULT 'NGN',
			status VARCHAR(50) DEFAULT 'PENDING',
			reference VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			settled_at TIMESTAMP WITH TIME ZONE
		);
		CREATE TABLE IF NOT EXISTS rtgs_db.liquidity_positions (
			bank_id VARCHAR(255) PRIMARY KEY,
			balance BIGINT NOT NULL DEFAULT 0,
			currency VARCHAR(10) DEFAULT 'NGN',
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		log.Printf("Warning: Schema creation failed: %v", err)
	}

	// Initialize Fabric client
	var fabric *fabricclient.Client
	fabric, err = fabricclient.NewClient(
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

	svc := &Service{db: database, fabric: fabric}

	r := mux.NewRouter()

	// Settlement endpoints
	r.HandleFunc("/rtgs/settle", svc.SettleHandler).Methods("POST")
	r.HandleFunc("/rtgs/settlements", svc.ListSettlementsHandler).Methods("GET")
	r.HandleFunc("/rtgs/settlements/{id}", svc.GetSettlementHandler).Methods("GET")

	// Liquidity management
	r.HandleFunc("/rtgs/liquidity", svc.GetLiquidityHandler).Methods("GET")
	r.HandleFunc("/rtgs/liquidity/inject", svc.InjectLiquidityHandler).Methods("POST")
	r.HandleFunc("/rtgs/liquidity/withdraw", svc.WithdrawLiquidityHandler).Methods("POST")

	// Interbank transfers
	r.HandleFunc("/rtgs/interbank", svc.InterbankTransferHandler).Methods("POST")

	// Health
	r.HandleFunc("/health", svc.HealthHandler).Methods("GET")

	log.Printf("RTGS Adapter Service running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}

func (s *Service) HealthHandler(w http.ResponseWriter, r *http.Request) {
	api.WriteSuccess(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "rtgs-adapter",
	})
}

// SettleHandler processes a settlement request
func (s *Service) SettleHandler(w http.ResponseWriter, r *http.Request) {
	var req SettlementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if req.Amount <= 0 {
		api.WriteError(w, http.StatusBadRequest, "invalid_amount", "Amount must be positive", "")
		return
	}

	if req.Currency == "" {
		req.Currency = "NGN"
	}

	// Generate settlement ID
	settlementID := fmt.Sprintf("RTGS-%d", time.Now().UnixNano())

	// Simulate RTGS processing (in production, this would call the actual RTGS system)
	// For now, we simulate instant settlement

	// Record settlement
	_, err := s.db.Exec(`
		INSERT INTO rtgs_db.settlements (id, from_bank_id, to_bank_id, amount, currency, status, reference, settled_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		settlementID, req.FromBankID, req.ToBankID, req.Amount, req.Currency, "SETTLED", req.Reference, time.Now())

	if err != nil {
		log.Printf("Failed to record settlement: %v", err)
		api.WriteError(w, http.StatusInternalServerError, "db_error", "Failed to record settlement", "")
		return
	}

	// Update liquidity positions
	s.db.Exec(`
		INSERT INTO rtgs_db.liquidity_positions (bank_id, balance, updated_at)
		VALUES ($1, -$2, $3)
		ON CONFLICT (bank_id) DO UPDATE SET balance = liquidity_positions.balance - $2, updated_at = $3`,
		req.FromBankID, req.Amount, time.Now())

	s.db.Exec(`
		INSERT INTO rtgs_db.liquidity_positions (bank_id, balance, updated_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (bank_id) DO UPDATE SET balance = liquidity_positions.balance + $2, updated_at = $3`,
		req.ToBankID, req.Amount, time.Now())

	log.Printf("RTGS Settlement: %s - %s -> %s, Amount: %d %s", settlementID, req.FromBankID, req.ToBankID, req.Amount, req.Currency)

	response := SettlementResponse{
		SettlementID: settlementID,
		Status:       "SETTLED",
		FromBankID:   req.FromBankID,
		ToBankID:     req.ToBankID,
		Amount:       req.Amount,
		Currency:     req.Currency,
		SettledAt:    time.Now(),
	}

	api.WriteSuccess(w, http.StatusOK, response)
}

// ListSettlementsHandler lists all settlements
func (s *Service) ListSettlementsHandler(w http.ResponseWriter, r *http.Request) {
	bankID := r.URL.Query().Get("bank_id")
	status := r.URL.Query().Get("status")

	query := "SELECT id, from_bank_id, to_bank_id, amount, currency, status, reference, created_at, settled_at FROM rtgs_db.settlements WHERE 1=1"
	args := []interface{}{}
	argNum := 1

	if bankID != "" {
		query += fmt.Sprintf(" AND (from_bank_id = $%d OR to_bank_id = $%d)", argNum, argNum)
		args = append(args, bankID)
		argNum++
	}

	if status != "" {
		query += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, status)
	}

	query += " ORDER BY created_at DESC LIMIT 100"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "db_error", err.Error(), "")
		return
	}
	defer rows.Close()

	var settlements []Settlement
	for rows.Next() {
		var s Settlement
		var settledAt sql.NullTime
		rows.Scan(&s.ID, &s.FromBankID, &s.ToBankID, &s.Amount, &s.Currency, &s.Status, &s.Reference, &s.CreatedAt, &settledAt)
		if settledAt.Valid {
			s.SettledAt = settledAt.Time
		}
		settlements = append(settlements, s)
	}

	api.WriteSuccess(w, http.StatusOK, settlements)
}

// GetSettlementHandler returns a specific settlement
func (s *Service) GetSettlementHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var settlement Settlement
	var settledAt sql.NullTime
	err := s.db.QueryRow(`
		SELECT id, from_bank_id, to_bank_id, amount, currency, status, reference, created_at, settled_at
		FROM rtgs_db.settlements WHERE id = $1`, id).
		Scan(&settlement.ID, &settlement.FromBankID, &settlement.ToBankID, &settlement.Amount,
			&settlement.Currency, &settlement.Status, &settlement.Reference, &settlement.CreatedAt, &settledAt)

	if err == sql.ErrNoRows {
		api.WriteError(w, http.StatusNotFound, "not_found", "Settlement not found", "")
		return
	}
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "db_error", err.Error(), "")
		return
	}

	if settledAt.Valid {
		settlement.SettledAt = settledAt.Time
	}

	api.WriteSuccess(w, http.StatusOK, settlement)
}

// GetLiquidityHandler returns liquidity positions
func (s *Service) GetLiquidityHandler(w http.ResponseWriter, r *http.Request) {
	bankID := r.URL.Query().Get("bank_id")

	if bankID != "" {
		var balance int64
		var currency string
		var updatedAt time.Time
		err := s.db.QueryRow(`
			SELECT balance, currency, updated_at FROM rtgs_db.liquidity_positions WHERE bank_id = $1`, bankID).
			Scan(&balance, &currency, &updatedAt)

		if err == sql.ErrNoRows {
			api.WriteSuccess(w, http.StatusOK, map[string]interface{}{
				"bank_id":    bankID,
				"balance":    0,
				"currency":   "NGN",
				"updated_at": time.Now(),
			})
			return
		}
		if err != nil {
			api.WriteError(w, http.StatusInternalServerError, "db_error", err.Error(), "")
			return
		}

		api.WriteSuccess(w, http.StatusOK, map[string]interface{}{
			"bank_id":    bankID,
			"balance":    balance,
			"currency":   currency,
			"updated_at": updatedAt,
		})
		return
	}

	// Return all positions
	rows, err := s.db.Query(`SELECT bank_id, balance, currency, updated_at FROM rtgs_db.liquidity_positions`)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "db_error", err.Error(), "")
		return
	}
	defer rows.Close()

	var positions []map[string]interface{}
	for rows.Next() {
		var bankID, currency string
		var balance int64
		var updatedAt time.Time
		rows.Scan(&bankID, &balance, &currency, &updatedAt)
		positions = append(positions, map[string]interface{}{
			"bank_id":    bankID,
			"balance":    balance,
			"currency":   currency,
			"updated_at": updatedAt,
		})
	}

	api.WriteSuccess(w, http.StatusOK, positions)
}

// InjectLiquidityHandler adds liquidity to a bank's position
func (s *Service) InjectLiquidityHandler(w http.ResponseWriter, r *http.Request) {
	var req LiquidityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if req.Currency == "" {
		req.Currency = "NGN"
	}

	_, err := s.db.Exec(`
		INSERT INTO rtgs_db.liquidity_positions (bank_id, balance, currency, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (bank_id) DO UPDATE SET balance = liquidity_positions.balance + $2, updated_at = $4`,
		req.BankID, req.Amount, req.Currency, time.Now())

	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "db_error", err.Error(), "")
		return
	}

	log.Printf("Liquidity Injected: %s, Amount: %d %s", req.BankID, req.Amount, req.Currency)

	api.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"status":   "injected",
		"bank_id":  req.BankID,
		"amount":   req.Amount,
		"currency": req.Currency,
	})
}

// WithdrawLiquidityHandler withdraws liquidity from a bank's position
func (s *Service) WithdrawLiquidityHandler(w http.ResponseWriter, r *http.Request) {
	var req LiquidityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	// Check current balance
	var currentBalance int64
	err := s.db.QueryRow(`SELECT balance FROM rtgs_db.liquidity_positions WHERE bank_id = $1`, req.BankID).Scan(&currentBalance)
	if err == sql.ErrNoRows || currentBalance < req.Amount {
		api.WriteError(w, http.StatusBadRequest, "insufficient_liquidity", "Insufficient liquidity", "")
		return
	}

	_, err = s.db.Exec(`
		UPDATE rtgs_db.liquidity_positions SET balance = balance - $1, updated_at = $2 WHERE bank_id = $3`,
		req.Amount, time.Now(), req.BankID)

	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "db_error", err.Error(), "")
		return
	}

	log.Printf("Liquidity Withdrawn: %s, Amount: %d", req.BankID, req.Amount)

	api.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"status":  "withdrawn",
		"bank_id": req.BankID,
		"amount":  req.Amount,
	})
}

// InterbankTransferHandler handles interbank CBDC transfers via RTGS
func (s *Service) InterbankTransferHandler(w http.ResponseWriter, r *http.Request) {
	var req SettlementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	// 1. First, settle via RTGS
	settlementID := fmt.Sprintf("RTGS-IB-%d", time.Now().UnixNano())

	_, err := s.db.Exec(`
		INSERT INTO rtgs_db.settlements (id, from_bank_id, to_bank_id, amount, currency, status, reference, settled_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		settlementID, req.FromBankID, req.ToBankID, req.Amount, req.Currency, "SETTLED", req.Reference, time.Now())

	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "db_error", "Failed to record settlement", "")
		return
	}

	// 2. Update CBDC balances on Fabric
	if s.fabric != nil {
		fromWallet := "wallet-" + req.FromBankID
		toWallet := "wallet-" + req.ToBankID
		_, err := s.fabric.SubmitTransaction("Transfer", fromWallet, toWallet, fmt.Sprintf("%d", req.Amount))
		if err != nil {
			log.Printf("Fabric transfer failed: %v", err)
			// Mark settlement as pending Fabric confirmation
			s.db.Exec(`UPDATE rtgs_db.settlements SET status = 'PENDING_CBDC' WHERE id = $1`, settlementID)
		}
	}

	log.Printf("Interbank Transfer: %s - %s -> %s, Amount: %d", settlementID, req.FromBankID, req.ToBankID, req.Amount)

	api.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"settlement_id": settlementID,
		"status":        "SETTLED",
		"from_bank":     req.FromBankID,
		"to_bank":       req.ToBankID,
		"amount":        req.Amount,
	})
}
