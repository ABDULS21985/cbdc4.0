package models

import "time"

type Transaction struct {
	ID         string    `json:"id"`
	FromWallet string    `json:"from_wallet"`
	ToWallet   string    `json:"to_wallet"`
	Amount     int64     `json:"amount"`
	Status     string    `json:"status"`
	TxHash     string    `json:"tx_hash,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type PaymentRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int64  `json:"amount"`
}

type BatchTransferRequest struct {
	FromWalletID string `json:"from_wallet_id"`
	Transfers    []struct {
		ToWalletID string `json:"to_wallet_id"`
		Amount     int64  `json:"amount"`
	} `json:"transfers"`
}
