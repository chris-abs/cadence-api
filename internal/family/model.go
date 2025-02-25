package family

import "github.com/chrisabs/storage/internal/models"

type CreateFamilyRequest struct {
    Name    string          `json:"name"`
    Modules []models.ModuleID `json:"modules"`
}

type CreateInviteRequest struct {
    FamilyID int             `json:"familyId"`
    Email    string          `json:"email"`
    Role     models.UserRole `json:"role"`
}

type UpdateModuleRequest struct {
    ModuleID  models.ModuleID `json:"moduleId"`
    IsEnabled bool          `json:"isEnabled"`
}

type JoinFamilyRequest struct {
    Token    string `json:"token"`
}
