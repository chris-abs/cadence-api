package models

import "time"

type Family struct {
    ID          int       `json:"id"`
    Name        string    `json:"name"`
    OwnerID     int      `json:"ownerId"`
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
    Modules     []Module  `json:"modules"`
}

type Module struct {
    ID        string         `json:"id"`
    IsEnabled bool          `json:"isEnabled"`
    Settings  ModuleSettings `json:"settings"`
}

type ModuleSettings struct {
    Permissions map[UserRole][]Permission `json:"permissions"`
}

type Permission string

const (
    PermissionRead   Permission = "READ"
    PermissionWrite  Permission = "WRITE"
    PermissionManage Permission = "MANAGE"
)

type FamilyInvite struct {
    ID        int       `json:"id"`
    FamilyID  int       `json:"familyId"`
    Email     string    `json:"email"`
    Role      UserRole  `json:"role"`
    Token     string    `json:"token"`
    ExpiresAt time.Time `json:"expiresAt"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}