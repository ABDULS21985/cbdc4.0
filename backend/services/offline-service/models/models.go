package models

import "time"

type Device struct {
	ID            string    `json:"id"`
	PublicKey     string    `json:"public_key"`
	Counter       int64     `json:"counter"`
	UserID        string    `json:"user_id"`
	HardwareID    string    `json:"hardware_id"` // Unique device hardware identifier
	OSVersion     string    `json:"os_version"`
	AppVersion    string    `json:"app_version"`
	TrustedStatus string    `json:"trusted_status"` // TRUSTED, COMPROMISED
	LastSyncAt    time.Time `json:"last_sync_at"`
	CreatedAt     time.Time `json:"created_at"`
}

type Voucher struct {
	ID            string    `json:"id"`
	DeviceID      string    `json:"device_id"`
	Amount        int64     `json:"amount"`
	Status        string    `json:"status"` // ACTIVE, REDEEMED, EXPIRED
	Signature     string    `json:"signature"`
	EncryptedData string    `json:"encrypted_data,omitempty"` // For secure storage
	ExpiresAt     time.Time `json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
}

type RegisterDeviceRequest struct {
	UserID     string `json:"user_id"`
	PublicKey  string `json:"public_key"`
	HardwareID string `json:"hardware_id"`
	OSVersion  string `json:"os_version"`
}

type OfflineTransaction struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int64  `json:"amount"`
	Counter   int64  `json:"counter"`
	Signature string `json:"signature"`
	Timestamp int64  `json:"timestamp"`
}

type ReconcileRequest struct {
	Transactions []OfflineTransaction `json:"transactions"`
}
