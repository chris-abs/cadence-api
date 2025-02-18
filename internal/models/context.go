package models

type UserContext struct {
    UserID        int
    FamilyID      *int
    Role          UserRole
    ModuleAccess  map[ModuleID][]Permission
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