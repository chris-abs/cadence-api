package models

type UserContext struct {
    UserID       int                        `json:"userId"`
    FamilyID     int                        `json:"familyId"`    
    Role         UserRole                   `json:"role"`
    ModuleAccess map[ModuleID][]Permission  `json:"moduleAccess"`
}

func (ctx *UserContext) CanAccess(module ModuleID, permission Permission) bool {
    permissions, exists := ctx.ModuleAccess[module]
    if !exists {
        return false
    }
    
    for _, p := range permissions {
        if p == permission {
            return true
        }
    }
    return false
}