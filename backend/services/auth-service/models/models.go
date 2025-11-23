```
package models

import (
	"encoding/json"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID           string          `json:"id"`
	Username     string          `json:"username"`
	PasswordHash string          `json:"-"`
	FullName     string          `json:"full_name"`
	Email        string          `json:"email"`
	PhoneNumber  string          `json:"phone_number"`
	BVN          string          `json:"bvn,omitempty"` // Bank Verification Number
	NIN          string          `json:"nin,omitempty"` // National Identity Number
	KYCData      json.RawMessage `json:"kyc_data"`      // Additional flexible KYC data
	Tier         string          `json:"tier"`
	Role         string          `json:"role"`   // CITIZEN, MERCHANT, ADMIN
	Status       string          `json:"status"` // ACTIVE, SUSPENDED, PENDING_KYC
	LastLoginAt  *time.Time      `json:"last_login_at,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	BVN         string `json:"bvn"`
	NIN         string `json:"nin"`
}

type TokenResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Tier     string `json:"tier"`
	jwt.RegisteredClaims
}
```
