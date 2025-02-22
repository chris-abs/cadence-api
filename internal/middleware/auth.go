package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/chrisabs/storage/internal/models"
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

func (m *AuthMiddleware) buildUserContext(userID int) (*models.UserContext, error) {
    query := `
        SELECT 
            u.id,
            u.family_id,
            u.role
        FROM users u
        WHERE u.id = $1`

    var ctx models.UserContext
    var familyID sql.NullInt64
    var role sql.NullString

    err := m.db.QueryRow(query, userID).Scan(
        &ctx.UserID,
        &familyID,
        &role,
    )

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found")
    }
    if err != nil {
        return nil, fmt.Errorf("error fetching user context: %v", err)
    }

    ctx.ModuleAccess = make(map[models.ModuleID][]models.Permission)

    if familyID.Valid {
        fid := int(familyID.Int64)  
        ctx.FamilyID = &fid        

        if role.Valid {
            userRole := models.UserRole(role.String)
            ctx.Role = &userRole
        }

        var modulesJSON []byte
        err := m.db.QueryRow(`SELECT modules FROM family WHERE id = $1`, fid).Scan(&modulesJSON)
        if err == nil {
            var modules []models.Module
            if err := json.Unmarshal(modulesJSON, &modules); err == nil {
                for _, module := range modules {
                    if module.IsEnabled {
                        if perms, exists := module.Settings.Permissions[*ctx.Role]; exists {
                            ctx.ModuleAccess[module.ID] = perms
                        }
                    }
                }
            }
        }
    }

    return &ctx, nil
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

        userID := int(claims["userId"].(float64))
        
        userCtx, err := m.buildUserContext(userID)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        moduleID := extractModuleFromPath(r.URL.Path)
        if moduleID != "" {
            if userCtx.FamilyID == nil {
                http.Error(w, "Must be part of a family to access modules", http.StatusForbidden)
                return
            }

            permission := mapHTTPMethodToPermission(r.Method)
            if !userCtx.CanAccess(moduleID, permission) {
                http.Error(w, "Insufficient permissions", http.StatusForbidden)
                return
            }
        }

        ctx := context.WithValue(r.Context(), "user", userCtx)
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}

func extractModuleFromPath(path string) models.ModuleID {
    parts := strings.Split(path, "/")
    if len(parts) > 1 {
        switch parts[1] {
        case "workspaces", "containers", "items", "tags":
            return models.ModuleStorage
        case "meals":
            return models.ModuleMeals
        case "services":
            return models.ModuleServices
        case "chores":
            return models.ModuleChores
        }
    }
    return ""
}

func mapHTTPMethodToPermission(method string) models.Permission {
    switch method {
    case http.MethodGet:
        return models.PermissionRead
    case http.MethodPost, http.MethodPut, http.MethodPatch:
        return models.PermissionWrite
    case http.MethodDelete:
        return models.PermissionManage
    default:
        return models.PermissionRead 
    }
}