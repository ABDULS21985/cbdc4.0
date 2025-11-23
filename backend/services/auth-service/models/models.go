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
	KYCData      json.RawMessage `json:"kyc_data"`
	Tier         string          `json:"tier"`
	CreatedAt    time.Time       `json:"created_at"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}
