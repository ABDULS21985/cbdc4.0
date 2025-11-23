package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/centralbank/cbdc/backend/pkg/common"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

var secretKey = []byte("super-secret-key-change-me")

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

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// TODO: Validate credentials against DB/LDAP
	// Mock:
	role := "CITIZEN"
	if req.Username == "admin" {
		role = "CBN_ADMIN"
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: req.Username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "cbdc-auth-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokenResponse{Token: tokenString})
}

func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	// Simplified: In a real app, we'd verify a refresh token.
	// Here we just issue a new token if the user is authenticated (mock).
	// For strict compliance, we'll assume the client sends the old token to be refreshed.

	// Mock implementation
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: "refreshed_user",
		Role:     "CITIZEN",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "cbdc-auth-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TokenResponse{Token: tokenString})
}

func VerifyHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}
	// Strip "Bearer " prefix if present
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":    true,
		"username": claims.Username,
		"role":     claims.Role,
	})
}

func main() {
	cfg := common.LoadConfig()
	r := mux.NewRouter()

	r.HandleFunc("/auth/login", LoginHandler).Methods("POST")
	r.HandleFunc("/auth/refresh", RefreshHandler).Methods("POST")
	r.HandleFunc("/auth/verify", VerifyHandler).Methods("GET")

	log.Printf("Auth Service running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
