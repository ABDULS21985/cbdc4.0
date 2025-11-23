package models

import "time"

type Device struct {
	ID        string    `json:"id"`
	PublicKey string    `json:"public_key"`
	Counter   int64     `json:"counter"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Voucher struct {
	ID        string    `json:"id"`
	DeviceID  string    `json:"device_id"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	Signature string    `json:"signature"`
	CreatedAt time.Time `json:"created_at"`
}

type RegisterDeviceRequest struct {
	UserID    string `json:"user_id"`
	PublicKey string `json:"public_key"`
}

type OfflineTransaction struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int64  `json:"amount"`
	Counter   int64  `json:"counter"`
	Signature string `json:"signature"`
}

type ReconcileRequest struct {
	Transactions []OfflineTransaction `json:"transactions"`
}
