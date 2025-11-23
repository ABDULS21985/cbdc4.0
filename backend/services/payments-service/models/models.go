package models

import (
	"encoding/json"
	"time"
)

type Transaction struct {
	ID          string          `json:"id"`
	FromWallet  string          `json:"from_wallet"`
	ToWallet    string          `json:"to_wallet"`
	Amount      int64           `json:"amount"`
	Currency    string          `json:"currency"` // e.g., NGN
	Fee         int64           `json:"fee"`
	Status      string          `json:"status"`             // PENDING, CONFIRMED, FAILED
	Type        string          `json:"type"`               // P2P, P2B, G2P, B2B
	Channel     string          `json:"channel"`            // MOBILE, WEB, USSD, POS
	Metadata    json.RawMessage `json:"metadata,omitempty"` // Flexible metadata (e.g., invoice ref)
	TxHash      string          `json:"tx_hash,omitempty"`
	Description string          `json:"description,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type PaymentRequest struct {
	From        string          `json:"from"`
	To          string          `json:"to"`
	Amount      int64           `json:"amount"`
	Type        string          `json:"type"` // Optional, default P2P
	Description string          `json:"description"`
	Metadata    json.RawMessage `json:"metadata"`
}

type BatchTransferRequest struct {
	FromWalletID string `json:"from_wallet_id"`
	Transfers    []struct {
		ToWalletID string `json:"to_wallet_id"`
		Amount     int64  `json:"amount"`
	} `json:"transfers"`
	Description string `json:"description"`
}
