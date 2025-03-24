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

type ModuleDefinition struct {
    ID          ModuleID `json:"id"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    IsAvailable bool     `json:"isAvailable"`
}