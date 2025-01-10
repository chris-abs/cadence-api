package middleware

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

type AuthMiddleware struct {
    jwtSecret string
    db        *sql.DB
}

func NewAuthMiddleware(jwtSecret string, db *sql.DB) *AuthMiddleware {
    return &AuthMiddleware{
        jwtSecret: jwtSecret,
        db:        db,
    }
}

func (m *AuthMiddleware) userExists(userID string) (bool, error) {
    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
    err := m.db.QueryRow(query, userID).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("error checking user existence: %v", err)
    }
    return exists, nil
}

func (m *AuthMiddleware) AuthHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(bearerToken[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		userID := fmt.Sprintf("%.0f", claims["userId"])
        
        exists, err := m.userExists(userID) 
        if err != nil || !exists {
            http.Error(w, "User not found", http.StatusUnauthorized)
            return
        }

        r.Header.Set("UserId", userID)
        next(w, r)
	}
}
