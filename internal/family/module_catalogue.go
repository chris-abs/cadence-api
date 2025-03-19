package family

import "github.com/chrisabs/cadence/internal/models"

var SystemModules = map[models.ModuleID]models.ModuleDefinition{
    models.ModuleStorage: {
        ID:          models.ModuleStorage,
        Name:        "Storage",
        Description: "Organize containers, items, and manage storage spaces",
        IsAvailable: true,
    },
    models.ModuleChores: {
        ID:          models.ModuleChores,
        Name:        "Chores",
        Description: "Assign and track household chores and responsibilities",
        IsAvailable: true,
    },
    models.ModuleMeals: {
        ID:          models.ModuleMeals,
        Name:        "Meals",
        Description: "Plan meals, create shopping lists, and track ingredients",
        IsAvailable: true,
    },
    models.ModuleServices: {
        ID:          models.ModuleServices,
        Name:        "Services",
        Description: "Track subscriptions, bills, and recurring payments",
        IsAvailable: false, 
    },
}

func GetAvailableModules() []models.ModuleDefinition {
    var availableModules []models.ModuleDefinition
    
    for _, module := range SystemModules {
        if module.IsAvailable {
            availableModules = append(availableModules, module)
        }
    }
    
    return availableModules
}