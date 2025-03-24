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
		IsModuleEnabled(familyID int, moduleID models.ModuleID) (bool, error)
		HasModulePermission(familyID int, role models.ProfileRole, moduleID models.ModuleID, permission models.Permission) (bool, error)
	}
	profileService interface {
		GetProfileByID(id int) (*models.Profile, error)
	}
}

func NewAuthMiddleware(
	jwtSecret string,
	db *sql.DB,
	familyService interface {
		IsModuleEnabled(familyID int, moduleID models.ModuleID) (bool, error)
		HasModulePermission(familyID int, role models.ProfileRole, moduleID models.ModuleID, permission models.Permission) (bool, error)
	},
	profileService interface {
		GetProfileByID(id int) (*models.Profile, error)
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
	familyID, ok := claims["familyId"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid token: missing family ID")
	}

	return &models.FamilyContext{
		FamilyID: int(familyID),
	}, nil
}

func (m *AuthMiddleware) buildProfileContext(claims jwt.MapClaims) (*models.ProfileContext, error) {
	familyID, ok := claims["familyId"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid token: missing family ID")
	}

	profileID, ok := claims["profileId"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid token: missing profile ID")
	}

	roleString, ok := claims["role"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid token: missing role")
	}

	isOwner, _ := claims["isOwner"].(bool)

	return &models.ProfileContext{
		FamilyID:  int(familyID),
		ProfileID: int(profileID),
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