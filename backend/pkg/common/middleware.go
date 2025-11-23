package common

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware verifies the JWT token and extracts claims
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		// Parse token (using the same secret as Auth Service - in prod use public key)
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("super-secret-key-change-me"), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// TODO: Inject claims into context
		next.ServeHTTP(w, r)
	})
}

// RequireRole enforces RBAC
func RequireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// In a real implementation, we would extract the role from the context
		// set by the AuthMiddleware.
		// For now, we assume the middleware passed and we re-parse or rely on context.

		// Mock check:
		// claims := r.Context().Value("claims").(Claims)
		// if claims.Role != role {
		//     http.Error(w, "Forbidden", http.StatusForbidden)
		//     return
		// }

		next(w, r)
	}
}
