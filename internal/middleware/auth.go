package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

type AuthMiddleware struct {
    jwtSecret string
    db        *sql.DB
}

type ModulePermission struct {
    ModuleID string   `json:"moduleId"`
    Enabled  bool     `json:"enabled"`
    Actions  []string `json:"actions"`
}

type UserModuleAccess struct {
    Role        string             `json:"role"`
    Permissions []ModulePermission `json:"permissions"`
}

func NewAuthMiddleware(jwtSecret string, db *sql.DB) *AuthMiddleware {
    return &AuthMiddleware{
        jwtSecret: jwtSecret,
        db:        db,
    }
}

func (m *AuthMiddleware) getUserAccess(userID string) (*UserModuleAccess, error) {
    query := `
        SELECT u.role, f.module_permissions
        FROM users u
        LEFT JOIN families f ON u.family_id = f.id
        WHERE u.id = $1`

    var (
        role        string
        permissions []byte
    )

    err := m.db.QueryRow(query, userID).Scan(&role, &permissions)
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found")
    }
    if err != nil {
        return nil, fmt.Errorf("error checking user access: %v", err)
    }

    access := &UserModuleAccess{Role: role}
    if permissions != nil {
        if err := json.Unmarshal(permissions, &access.Permissions); err != nil {
            return nil, fmt.Errorf("error parsing permissions: %v", err)
        }
    }

    return access, nil
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

        access, err := m.getUserAccess(userID)
        if err != nil {
            http.Error(w, "Error checking user access", http.StatusInternalServerError)
            return
        }

        r.Header.Set("UserId", userID)
        r.Header.Set("UserRole", access.Role)
        
        module := extractModuleFromPath(r.URL.Path)
        action := mapHTTPMethodToAction(r.Method)

        if !hasPermission(access.Permissions, module, action) {
            http.Error(w, "Insufficient permissions", http.StatusForbidden)
            return
        }

        next(w, r)
    }
}

func extractModuleFromPath(path string) string {
    parts := strings.Split(path, "/")
    if len(parts) > 1 {
        return parts[1] 
    }
    return ""
}

func mapHTTPMethodToAction(method string) string {
    switch method {
    case http.MethodGet:
        return "READ"
    case http.MethodPost, http.MethodPut, http.MethodPatch:
        return "WRITE"
    case http.MethodDelete:
        return "DELETE"
    default:
        return ""
    }
}

func hasPermission(permissions []ModulePermission, moduleID, action string) bool {
    for _, p := range permissions {
        if p.ModuleID == moduleID && p.Enabled {
            for _, a := range p.Actions {
                if a == action {
                    return true
                }
            }
        }
    }
    return false
}