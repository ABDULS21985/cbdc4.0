package models

import "time"

type Wallet struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Address       string    `json:"address"`
	Type          string    `json:"type"`     // RETAIL, MERCHANT, IOT, GOV
	Status        string    `json:"status"`   // ACTIVE, FROZEN, CLOSED
	Currency      string    `json:"currency"` // e.g., NGN
	Balance       int64     `json:"balance"`  // Cached balance
	TierLevel     string    `json:"tier_level"`
	DailyLimit    int64     `json:"daily_limit"`
	EncryptedKeys string    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateWalletRequest struct {
	UserID string `json:"user_id"`
	Tier   string `json:"tier"`
	Type   string `json:"type"` // Optional, default RETAIL
}

type WalletBalance struct {
	Balance  int64  `json:"balance"`
	Currency string `json:"currency"`
}
