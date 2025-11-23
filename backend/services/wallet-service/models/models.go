package models

import "time"

type Wallet struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Address       string    `json:"address"`
	EncryptedKeys string    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateWalletRequest struct {
	UserID string `json:"user_id"`
	Tier   string `json:"tier"`
}

type WalletBalance struct {
	Balance int64 `json:"balance"`
}
