package models

import "time"

type FamilyStatus string

const (
    FamilyStatusActive   FamilyStatus = "ACTIVE"
    FamilyStatusInactive FamilyStatus = "INACTIVE"
)

type Family struct {
    ID          int          `json:"id"`
    Name        string       `json:"name"`
    OwnerID     int          `json:"ownerId"`
    Status      FamilyStatus `json:"status"`
    CreatedAt   time.Time    `json:"createdAt"`
    UpdatedAt   time.Time    `json:"updatedAt"`
    Modules     []Module     `json:"modules"`
}

type ModuleID string

const (
    ModuleStorage  ModuleID = "storage"
    ModuleChores   ModuleID = "chores"
    ModuleMeals    ModuleID = "meals"
    ModuleServices ModuleID = "services"
)

type Module struct {
    ID        ModuleID `json:"id"`
    IsEnabled bool   `json:"isEnabled"`
}

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
