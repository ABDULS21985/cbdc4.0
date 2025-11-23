package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/centralbank/cbdc/backend/pkg/common"
	"github.com/centralbank/cbdc/backend/pkg/common/api"
	"github.com/centralbank/cbdc/backend/pkg/common/db"
	"github.com/centralbank/cbdc/backend/pkg/fabricclient"
	"github.com/gorilla/mux"
)

// Service represents the CBN Operations Service
// As per Phase 3 design: Backend for the Central Bank Operations Console
type Service struct {
	db     *sql.DB
	fabric *fabricclient.Client
}

// IssuanceRequest represents a request to mint new CBDC
type IssuanceRequest struct {
	Amount         int64  `json:"amount"`
	ToIntermediaryID string `json:"to_intermediary_id"`
	Reason         string `json:"reason"`
	ApprovedBy     string `json:"approved_by"`
}

// RedemptionRequest represents a request to burn CBDC
type RedemptionRequest struct {
	Amount           int64  `json:"amount"`
	FromIntermediaryID string `json:"from_intermediary_id"`
	Reason           string `json:"reason"`
	ApprovedBy       string `json:"approved_by"`
}

// FreezeRequest represents a request to freeze a wallet
type FreezeRequest struct {
	WalletID string `json:"wallet_id"`
	Reason   string `json:"reason"`
}

// IntermediaryStatus represents the status of an intermediary
type IntermediaryStatus struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Status        string    `json:"status"`
	CBDCBalance   int64     `json:"cbdc_balance"`
	CustomerCount int       `json:"customer_count"`
	LastActivity  time.Time `json:"last_activity"`
}

// DashboardStats represents the dashboard statistics
type DashboardStats struct {
	TotalSupply          int64              `json:"total_supply"`
	CirculatingSupply    int64              `json:"circulating_supply"`
	TotalTransactions24h int                `json:"total_transactions_24h"`
	TotalVolume24h       int64              `json:"total_volume_24h"`
	ActiveWallets        int                `json:"active_wallets"`
	Intermediaries       []IntermediaryStatus `json:"intermediaries"`
	LastUpdated          time.Time          `json:"last_updated"`
}

func main() {
	cfg := common.LoadConfig()

	// Connect to DB
	database, err := db.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer database.Close()

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

	// Dashboard & Analytics
	r.HandleFunc("/ops/dashboard", svc.DashboardHandler).Methods("GET")
	r.HandleFunc("/ops/supply", svc.GetTotalSupplyHandler).Methods("GET")

	// Issuance & Redemption (Mint/Burn)
	r.HandleFunc("/ops/issue", svc.IssueHandler).Methods("POST")
	r.HandleFunc("/ops/redeem", svc.RedeemHandler).Methods("POST")

	// Wallet Management
	r.HandleFunc("/ops/freeze", svc.FreezeWalletHandler).Methods("POST")
	r.HandleFunc("/ops/unfreeze", svc.UnfreezeWalletHandler).Methods("POST")

	// Intermediary Management
	r.HandleFunc("/ops/intermediaries", svc.ListIntermediariesHandler).Methods("GET")
	r.HandleFunc("/ops/intermediaries/{id}", svc.GetIntermediaryHandler).Methods("GET")

	// Governance
	r.HandleFunc("/ops/params", svc.GetGovernanceParamsHandler).Methods("GET")
	r.HandleFunc("/ops/params", svc.UpdateGovernanceParamsHandler).Methods("PUT")

	// Audit & Compliance
	r.HandleFunc("/ops/audit/transactions", svc.AuditTransactionsHandler).Methods("GET")
	r.HandleFunc("/ops/audit/wallets", svc.AuditWalletsHandler).Methods("GET")

	// Health
	r.HandleFunc("/health", svc.HealthHandler).Methods("GET")

	log.Printf("CBN Ops Service running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}

func (s *Service) HealthHandler(w http.ResponseWriter, r *http.Request) {
	api.WriteSuccess(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "cbn-ops-service",
	})
}

// DashboardHandler returns the main dashboard statistics
func (s *Service) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	stats := DashboardStats{
		LastUpdated: time.Now(),
	}

	// Get total supply from Fabric
	if s.fabric != nil {
		result, err := s.fabric.EvaluateTransaction("GetTotalSupply")
		if err == nil {
			json.Unmarshal(result, &stats.TotalSupply)
		}
	}

	// Get transaction stats from DB
	s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(amount), 0)
		FROM payments_db.transactions
		WHERE created_at > NOW() - INTERVAL '24 hours'
	`).Scan(&stats.TotalTransactions24h, &stats.TotalVolume24h)

	// Get active wallets count
	s.db.QueryRow(`
		SELECT COUNT(*) FROM wallet_db.wallets WHERE status = 'ACTIVE'
	`).Scan(&stats.ActiveWallets)

	// Get intermediary stats (mock for now)
	stats.Intermediaries = []IntermediaryStatus{
		{ID: "bank-a", Name: "Bank A", Status: "ACTIVE", CBDCBalance: 1000000, CustomerCount: 5000},
		{ID: "bank-b", Name: "Bank B", Status: "ACTIVE", CBDCBalance: 750000, CustomerCount: 3500},
		{ID: "fintech-x", Name: "FintechX", Status: "ACTIVE", CBDCBalance: 250000, CustomerCount: 10000},
	}

	stats.CirculatingSupply = stats.TotalSupply // Simplified

	api.WriteSuccess(w, http.StatusOK, stats)
}

// GetTotalSupplyHandler returns the total CBDC supply
func (s *Service) GetTotalSupplyHandler(w http.ResponseWriter, r *http.Request) {
	var totalSupply int64

	if s.fabric != nil {
		result, err := s.fabric.EvaluateTransaction("GetTotalSupply")
		if err != nil {
			api.WriteError(w, http.StatusInternalServerError, "fabric_error", err.Error(), "")
			return
		}
		json.Unmarshal(result, &totalSupply)
	}

	api.WriteSuccess(w, http.StatusOK, map[string]int64{"total_supply": totalSupply})
}

// IssueHandler mints new CBDC to an intermediary
func (s *Service) IssueHandler(w http.ResponseWriter, r *http.Request) {
	var req IssuanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if req.Amount <= 0 {
		api.WriteError(w, http.StatusBadRequest, "invalid_amount", "Amount must be positive", "")
		return
	}

	// Submit to Fabric
	if s.fabric == nil {
		api.WriteError(w, http.StatusServiceUnavailable, "fabric_unavailable", "Fabric not connected", "")
		return
	}

	walletID := "wallet-" + req.ToIntermediaryID
	_, err := s.fabric.SubmitTransaction("Issue", string(rune(req.Amount)), walletID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "fabric_error", err.Error(), "")
		return
	}

	// Log the issuance
	log.Printf("CBDC Issued: %d to %s by %s - Reason: %s", req.Amount, req.ToIntermediaryID, req.ApprovedBy, req.Reason)

	api.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"status":           "issued",
		"amount":           req.Amount,
		"to_intermediary":  req.ToIntermediaryID,
		"timestamp":        time.Now(),
	})
}

// RedeemHandler burns CBDC from an intermediary
func (s *Service) RedeemHandler(w http.ResponseWriter, r *http.Request) {
	var req RedemptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if req.Amount <= 0 {
		api.WriteError(w, http.StatusBadRequest, "invalid_amount", "Amount must be positive", "")
		return
	}

	if s.fabric == nil {
		api.WriteError(w, http.StatusServiceUnavailable, "fabric_unavailable", "Fabric not connected", "")
		return
	}

	walletID := "wallet-" + req.FromIntermediaryID
	_, err := s.fabric.SubmitTransaction("Redeem", string(rune(req.Amount)), walletID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "fabric_error", err.Error(), "")
		return
	}

	log.Printf("CBDC Redeemed: %d from %s by %s - Reason: %s", req.Amount, req.FromIntermediaryID, req.ApprovedBy, req.Reason)

	api.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"status":             "redeemed",
		"amount":             req.Amount,
		"from_intermediary":  req.FromIntermediaryID,
		"timestamp":          time.Now(),
	})
}

// FreezeWalletHandler freezes a wallet
func (s *Service) FreezeWalletHandler(w http.ResponseWriter, r *http.Request) {
	var req FreezeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if s.fabric == nil {
		api.WriteError(w, http.StatusServiceUnavailable, "fabric_unavailable", "Fabric not connected", "")
		return
	}

	_, err := s.fabric.SubmitTransaction("FreezeWallet", req.WalletID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "fabric_error", err.Error(), "")
		return
	}

	log.Printf("Wallet Frozen: %s - Reason: %s", req.WalletID, req.Reason)

	api.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"status":    "frozen",
		"wallet_id": req.WalletID,
		"timestamp": time.Now(),
	})
}

// UnfreezeWalletHandler unfreezes a wallet
func (s *Service) UnfreezeWalletHandler(w http.ResponseWriter, r *http.Request) {
	var req FreezeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if s.fabric == nil {
		api.WriteError(w, http.StatusServiceUnavailable, "fabric_unavailable", "Fabric not connected", "")
		return
	}

	_, err := s.fabric.SubmitTransaction("UnfreezeWallet", req.WalletID)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "fabric_error", err.Error(), "")
		return
	}

	log.Printf("Wallet Unfrozen: %s", req.WalletID)

	api.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"status":    "active",
		"wallet_id": req.WalletID,
		"timestamp": time.Now(),
	})
}

// ListIntermediariesHandler lists all registered intermediaries
func (s *Service) ListIntermediariesHandler(w http.ResponseWriter, r *http.Request) {
	// In production, fetch from DB or Fabric
	intermediaries := []IntermediaryStatus{
		{ID: "bank-a", Name: "Bank A", Status: "ACTIVE", CBDCBalance: 1000000, CustomerCount: 5000, LastActivity: time.Now()},
		{ID: "bank-b", Name: "Bank B", Status: "ACTIVE", CBDCBalance: 750000, CustomerCount: 3500, LastActivity: time.Now()},
		{ID: "fintech-x", Name: "FintechX PSP", Status: "ACTIVE", CBDCBalance: 250000, CustomerCount: 10000, LastActivity: time.Now()},
	}

	api.WriteSuccess(w, http.StatusOK, intermediaries)
}

// GetIntermediaryHandler returns details of a specific intermediary
func (s *Service) GetIntermediaryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Mock data
	intermediary := IntermediaryStatus{
		ID:            id,
		Name:          "Bank " + id,
		Status:        "ACTIVE",
		CBDCBalance:   500000,
		CustomerCount: 2500,
		LastActivity:  time.Now(),
	}

	api.WriteSuccess(w, http.StatusOK, intermediary)
}

// GetGovernanceParamsHandler returns the current governance parameters
func (s *Service) GetGovernanceParamsHandler(w http.ResponseWriter, r *http.Request) {
	if s.fabric == nil {
		api.WriteError(w, http.StatusServiceUnavailable, "fabric_unavailable", "Fabric not connected", "")
		return
	}

	result, err := s.fabric.EvaluateTransaction("GetParams")
	if err != nil {
		// Return defaults if not set
		api.WriteSuccess(w, http.StatusOK, map[string]interface{}{
			"max_transaction_limit": 1000000,
			"min_transaction_limit": 1,
			"fee_percentage":        0.001,
			"tier0_daily_limit":     10000,
			"tier1_daily_limit":     100000,
			"tier2_daily_limit":     1000000,
		})
		return
	}

	var params map[string]interface{}
	json.Unmarshal(result, &params)
	api.WriteSuccess(w, http.StatusOK, params)
}

// UpdateGovernanceParamsHandler updates governance parameters
func (s *Service) UpdateGovernanceParamsHandler(w http.ResponseWriter, r *http.Request) {
	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	if s.fabric == nil {
		api.WriteError(w, http.StatusServiceUnavailable, "fabric_unavailable", "Fabric not connected", "")
		return
	}

	paramsJSON, _ := json.Marshal(params)
	_, err := s.fabric.SubmitTransaction("UpdateParams", string(paramsJSON))
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "fabric_error", err.Error(), "")
		return
	}

	log.Printf("Governance params updated: %v", params)
	api.WriteSuccess(w, http.StatusOK, map[string]string{"status": "updated"})
}

// AuditTransactionsHandler returns transactions for audit purposes
func (s *Service) AuditTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	// Query parameters
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "100"
	}

	rows, err := s.db.Query(`
		SELECT id, from_wallet, to_wallet, amount, type, status, created_at
		FROM payments_db.transactions
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "db_error", err.Error(), "")
		return
	}
	defer rows.Close()

	var transactions []map[string]interface{}
	for rows.Next() {
		var id, from, to, txType, status string
		var amount int64
		var createdAt time.Time
		rows.Scan(&id, &from, &to, &amount, &txType, &status, &createdAt)
		transactions = append(transactions, map[string]interface{}{
			"id":         id,
			"from":       from,
			"to":         to,
			"amount":     amount,
			"type":       txType,
			"status":     status,
			"created_at": createdAt,
		})
	}

	api.WriteSuccess(w, http.StatusOK, transactions)
}

// AuditWalletsHandler returns wallet information for audit purposes
func (s *Service) AuditWalletsHandler(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "ACTIVE"
	}

	rows, err := s.db.Query(`
		SELECT id, user_id, type, status, tier_level, daily_limit, balance, created_at
		FROM wallet_db.wallets
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT 100`, status)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "db_error", err.Error(), "")
		return
	}
	defer rows.Close()

	var wallets []map[string]interface{}
	for rows.Next() {
		var id, userID, walletType, walletStatus, tier string
		var dailyLimit, balance int64
		var createdAt time.Time
		rows.Scan(&id, &userID, &walletType, &walletStatus, &tier, &dailyLimit, &balance, &createdAt)
		wallets = append(wallets, map[string]interface{}{
			"id":          id,
			"user_id":     userID,
			"type":        walletType,
			"status":      walletStatus,
			"tier":        tier,
			"daily_limit": dailyLimit,
			"balance":     balance,
			"created_at":  createdAt,
		})
	}

	api.WriteSuccess(w, http.StatusOK, wallets)
}
