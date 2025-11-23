package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/centralbank/cbdc/backend/pkg/common"
	"github.com/centralbank/cbdc/backend/pkg/common/api"
	"github.com/centralbank/cbdc/backend/pkg/common/db"
	"github.com/centralbank/cbdc/backend/pkg/common/migrations"
	"github.com/centralbank/cbdc/backend/services/auth-service/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var secretKey = []byte("super-secret-key-change-me") // In production, load from env/vault

type Service struct {
	db *sql.DB
}

func (s *Service) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to hash password", "")
		return
	}

	// Insert User
	userID := "user-" + req.Username

	_, err = s.db.Exec(`
		INSERT INTO wallet_db.users (
			id, username, password_hash, full_name, email, phone_number, bvn, nin, tier, role, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		userID, req.Username, string(hashedPassword), req.FullName, req.Email, req.PhoneNumber, req.BVN, req.NIN, "TIER_1", "CITIZEN", "ACTIVE")

	if err != nil {
		log.Printf("Failed to register user: %v", err)
		api.WriteError(w, http.StatusConflict, "user_exists", "Username, email, or phone already exists", "")
		return
	}

	api.WriteSuccess(w, http.StatusCreated, map[string]string{"user_id": userID, "status": "created"})
}

func (s *Service) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteError(w, http.StatusBadRequest, "invalid_request", "Invalid request body", "")
		return
	}

	// Query DB for user
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, password_hash, role, tier, status 
		FROM wallet_db.users WHERE username = $1`, req.Username).
		Scan(&user.ID, &user.PasswordHash, &user.Role, &user.Tier, &user.Status)

	if err == sql.ErrNoRows {
		api.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid username or password", "")
		return
	} else if err != nil {
		log.Printf("DB Error: %v", err)
		api.WriteError(w, http.StatusInternalServerError, "internal_error", "Database error", "")
		return
	}

	if user.Status != "ACTIVE" {
		api.WriteError(w, http.StatusForbidden, "account_inactive", "Account is not active", "")
		return
	}

	// Verify Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		api.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid username or password", "")
		return
	}

	// Update Last Login
	go func() {
		s.db.Exec("UPDATE wallet_db.users SET last_login_at = $1 WHERE id = $2", time.Now(), user.ID)
	}()

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &models.Claims{
		UserID:   user.ID,
		Username: req.Username,
		Role:     user.Role,
		Tier:     user.Tier,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			Issuer:    "cbdc-auth-service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to generate token", "")
		return
	}

	api.WriteSuccess(w, http.StatusOK, models.TokenResponse{Token: tokenString, ExpiresAt: expirationTime.Unix()})
}

func (s *Service) RefreshHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		api.WriteError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired token", "")
		return
	}

	// Issue new token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims.ExpiresAt = jwt.NewNumericDate(expirationTime)

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	newTokenString, err := newToken.SignedString(secretKey)
	if err != nil {
		api.WriteError(w, http.StatusInternalServerError, "internal_error", "Failed to refresh token", "")
		return
	}

	api.WriteSuccess(w, http.StatusOK, models.TokenResponse{Token: newTokenString, ExpiresAt: expirationTime.Unix()})
}

func (s *Service) VerifyHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		api.WriteError(w, http.StatusUnauthorized, "missing_token", "Missing Authorization header", "")
		return
	}
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	claims := &models.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		api.WriteError(w, http.StatusUnauthorized, "invalid_token", "Invalid or expired token", "")
		return
	}

	api.WriteSuccess(w, http.StatusOK, map[string]interface{}{
		"valid":    true,
		"user_id":  claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
		"tier":     claims.Tier,
	})
}

func main() {
	cfg := common.LoadConfig()

	// Connect to DB
	database, err := db.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer database.Close()

	// Run Migrations
	if err := migrations.RunMigrations(database, "backend/migrations/auth"); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	svc := &Service{db: database}

	r := mux.NewRouter()

	r.HandleFunc("/auth/register", svc.RegisterHandler).Methods("POST")
	r.HandleFunc("/auth/login", svc.LoginHandler).Methods("POST")
	r.HandleFunc("/auth/refresh", svc.RefreshHandler).Methods("POST")
	r.HandleFunc("/auth/verify", svc.VerifyHandler).Methods("GET")

	log.Printf("Auth Service running on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
