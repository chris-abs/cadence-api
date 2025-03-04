package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/chrisabs/cadence/internal/models"
	"github.com/golang-jwt/jwt"
)

type AuthMiddleware struct {
	jwtSecret        string
	db               *sql.DB
	membershipService interface {
		GetActiveMembershipForUser(userID int) (*models.FamilyMembership, error)
	}
	familyService interface {
		IsModuleEnabled(familyID int, moduleID models.ModuleID) (bool, error)
		HasModulePermission(familyID int, userRole models.UserRole, moduleID models.ModuleID, permission models.Permission) (bool, error)
	}
}

func NewAuthMiddleware(
	jwtSecret string,
	db *sql.DB,
	membershipService interface {
		GetActiveMembershipForUser(userID int) (*models.FamilyMembership, error)
	},
	familyService interface {
		IsModuleEnabled(familyID int, moduleID models.ModuleID) (bool, error)
		HasModulePermission(familyID int, userRole models.UserRole, moduleID models.ModuleID, permission models.Permission) (bool, error)
	},
) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret:        jwtSecret,
		db:               db,
		membershipService: membershipService,
		familyService:    familyService,
	}
}

func (m *AuthMiddleware) buildUserContext(userID int) (*models.UserContext, error) {
	ctx := &models.UserContext{
		UserID: userID,
	}
	
	membership, err := m.membershipService.GetActiveMembershipForUser(userID)
	if err == nil && membership != nil {
		ctx.FamilyID = &membership.FamilyID
		ctx.Role = &membership.Role
	}
	
	return ctx, nil
}

func (m *AuthMiddleware) AuthHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		bearerToken := strings.Split(authHeader, " ")
		if len(bearerToken) != 2 || bearerToken[0] != "Bearer" {
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		tokenString := bearerToken[1]
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(m.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		userID, ok := claims["userId"].(float64)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		userCtx, err := m.buildUserContext(int(userID))
		if err != nil {
			userCtx = &models.UserContext{
				UserID: int(userID),
			}
		}

		ctx := context.WithValue(r.Context(), "user", userCtx)
		next(w, r.WithContext(ctx))
	}
}

func (m *AuthMiddleware) ModuleMiddleware(moduleID models.ModuleID, permission models.Permission) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return m.AuthHandler(func(w http.ResponseWriter, r *http.Request) {
			userCtx := r.Context().Value("user").(*models.UserContext)

			if userCtx.FamilyID == nil || userCtx.Role == nil {
				http.Error(w, "Access denied: Not a family member", http.StatusForbidden)
				return
			}

			hasPermission, err := m.familyService.HasModulePermission(*userCtx.FamilyID, *userCtx.Role, moduleID, permission)
			if err != nil {
				http.Error(w, "Error checking permissions", http.StatusInternalServerError)
				return
			}

			if !hasPermission {
				http.Error(w, "Access denied: Insufficient permissions", http.StatusForbidden)
				return
			}

			next(w, r)
		})
	}
}