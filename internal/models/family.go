package models

type FamilyStatus string

const (
    FamilyStatusActive   FamilyStatus = "ACTIVE"
    FamilyStatusInactive FamilyStatus = "INACTIVE"
)

type ModuleID string

const (
    ModuleStorage  ModuleID = "storage"
    ModuleChores   ModuleID = "chores"
    ModuleMeals    ModuleID = "meals"
    ModuleServices ModuleID = "services"
)

type Module struct {
    ID        ModuleID `json:"id"`
    IsEnabled bool     `json:"isEnabled"`
}

