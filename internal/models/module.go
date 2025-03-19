package models

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

type ModuleAvailability string

const (
    ModuleAvailabilityPublic ModuleAvailability = "PUBLIC"
    ModuleAvailabilityBeta   ModuleAvailability = "BETA"
    ModuleAvailabilityHidden ModuleAvailability = "HIDDEN"
)

type ModuleDefinition struct {
    ID           ModuleID          `json:"id"`
    Name         string            `json:"name"`
    Description  string            `json:"description"`
    Availability ModuleAvailability `json:"availability"`
}