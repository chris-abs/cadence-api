package family

import "github.com/chrisabs/storage/internal/models"

type CreateFamilyRequest struct {
    Name    string `json:"name"`
    OwnerID int    `json:"ownerId"`
}

type CreateInviteRequest struct {
    FamilyID int             `json:"familyId"`
    Email    string          `json:"email"`
    Role     models.UserRole `json:"role"`
}

type UpdateModuleRequest struct {
    ModuleID    models.ModuleID                          `json:"moduleId"`    
    IsEnabled   bool                                     `json:"isEnabled"`
    Permissions map[models.UserRole][]models.Permission  `json:"permissions"`
}