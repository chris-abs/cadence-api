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
	jwtSecret       string
	db              *sql.DB
	familyService   interface {
		IsModuleEnabled(familyID models.FamilyID, moduleID models.ModuleID) (bool, error)
		HasModulePermission(familyID models.FamilyID, role models.ProfileRole, moduleID models.ModuleID, permission models.Permission) (bool, error)
	}
	profileService interface {
		GetProfileByID(id models.ProfileID) (*models.Profile, error)
	}
}

func NewAuthMiddleware(
	jwtSecret string,
	db *sql.DB,
	familyService interface {
		IsModuleEnabled(familyID models.FamilyID, moduleID models.ModuleID) (bool, error)
		HasModulePermission(familyID models.FamilyID, role models.ProfileRole, moduleID models.ModuleID, permission models.Permission) (bool, error)
	},
	profileService interface {
		GetProfileByID(id models.ProfileID) (*models.Profile, error)
	},
) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret:      jwtSecret,
		db:             db,
		familyService:  familyService,
		profileService: profileService,
	}
}

func (m *AuthMiddleware) buildFamilyContext(claims jwt.MapClaims) (*models.FamilyContext, error) {
	familyIDStr, ok := claims["familyId"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token: missing family ID")
	}

	familyID := models.FamilyID(familyIDStr)
	if !familyID.IsValid() {
		return nil, fmt.Errorf("invalid token: invalid family ID format")
	}

	return &models.FamilyContext{
		FamilyID: familyID,
	}, nil
}

func (m *AuthMiddleware) buildProfileContext(claims jwt.MapClaims) (*models.ProfileContext, error) {
	familyIDStr, ok := claims["familyId"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token: missing family ID")
	}

	profileIDStr, ok := claims["profileId"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token: missing profile ID")
	}

	roleString, ok := claims["role"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token: missing role")
	}

	isOwner, _ := claims["isOwner"].(bool)

	familyID := models.FamilyID(familyIDStr)
	if !familyID.IsValid() {
		return nil, fmt.Errorf("invalid token: invalid family ID format")
	}

	profileID := models.ProfileID(profileIDStr)
	if !profileID.IsValid() {
		return nil, fmt.Errorf("invalid token: invalid profile ID format")
	}

	return &models.ProfileContext{
		FamilyID:  familyID,
		ProfileID: profileID,
		Role:      models.ProfileRole(roleString),
		IsOwner:   isOwner,
	}, nil
}

func (m *AuthMiddleware) FamilyAuthHandler(next http.HandlerFunc) http.HandlerFunc {
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

		familyCtx, err := m.buildFamilyContext(claims)
		if err != nil {
			http.Error(w, "Invalid family token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "family", familyCtx)
		next(w, r.WithContext(ctx))
	}
}

func (m *AuthMiddleware) ProfileAuthHandler(next http.HandlerFunc) http.HandlerFunc {
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

		profileCtx, err := m.buildProfileContext(claims)
		if err != nil {
			http.Error(w, "Invalid profile token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "profile", profileCtx)
		next(w, r.WithContext(ctx))
	}
}

func (m *AuthMiddleware) ModuleMiddleware(moduleID models.ModuleID, permission models.Permission) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return m.ProfileAuthHandler(func(w http.ResponseWriter, r *http.Request) {
			profileCtx := r.Context().Value("profile").(*models.ProfileContext)

			hasPermission, err := m.familyService.HasModulePermission(profileCtx.FamilyID, profileCtx.Role, moduleID, permission)
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