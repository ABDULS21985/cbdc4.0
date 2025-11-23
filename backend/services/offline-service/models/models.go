package models

import "time"

type Device struct {
	ID            string    `json:"id"`
	PublicKey     string    `json:"public_key"`
	UserID        string    `json:"user_id"`
	Counter       int64     `json:"counter"`
	HardwareID    string    `json:"hardware_id"`
	OSVersion     string    `json:"os_version"`
	TrustedStatus string    `json:"trusted_status"`
	LastSyncAt    time.Time `json:"last_sync_at"`
}

type Voucher struct {
	ID            string    `json:"id"`
	DeviceID      string    `json:"device_id"`
	Amount        int64     `json:"amount"`
	Status        string    `json:"status"`
	Signature     string    `json:"signature"`
	EncryptedData string    `json:"encrypted_data"`
	ExpiresAt     time.Time `json:"expires_at"`
}

type RegisterDeviceRequest struct {
	UserID     string `json:"user_id"`
	PublicKey  string `json:"public_key"`
	HardwareID string `json:"hardware_id"`
	OSVersion  string `json:"os_version"`
}

// OfflinePurse represents the shadow state of an offline device
type OfflinePurse struct {
	DeviceID     string    `json:"device_id"`
	UserID       string    `json:"user_id"`
	Balance      int64     `json:"balance"`
	Counter      int64     `json:"counter"`
	LastSyncHash string    `json:"last_sync_hash"`
	LastSyncAt   time.Time `json:"last_sync_at"`
	Status       string    `json:"status"`
}

// PaymentIntent is the data structure signed by the payer
type PaymentIntent struct {
	Amount  int64  `json:"amount"`
	PayeeID string `json:"payee_id"`
	Counter int64  `json:"counter"`
	Nonce   string `json:"nonce"` // Randomness to prevent identical hashes
}

// SignedPayment is the offline transaction blob
type SignedPayment struct {
	PayerID   string `json:"payer_id"`
	PayeeID   string `json:"payee_id"`
	Amount    int64  `json:"amount"`
	Counter   int64  `json:"counter"`
	Signature string `json:"signature"` // Ed25519 signature of PaymentIntent
	Intent    string `json:"intent"`    // JSON string of PaymentIntent
}

type ReconcileRequest struct {
	DeviceID     string          `json:"device_id"`
	Transactions []SignedPayment `json:"transactions"`
}

type FundPurseRequest struct {
	UserID   string `json:"user_id"`
	DeviceID string `json:"device_id"`
	Amount   int64  `json:"amount"`
}
