package entities

import "time"

type ShoppingList struct {
	ID          int                     `json:"id"`
	FamilyID    int                     `json:"familyId"`
	MealPlanID  int                     `json:"mealPlanId"`
	StartDate   time.Time               `json:"startDate"`
	EndDate     time.Time               `json:"endDate"`
	Items       []ShoppingListItem      `json:"items"`
	Categories  []ShoppingListCategory  `json:"categories"`
	GeneratedAt time.Time               `json:"generatedAt"`
	UpdatedAt   time.Time               `json:"updatedAt"`
}

type ShoppingListItem struct {
	ID             int     `json:"id"`
	IngredientName string  `json:"ingredientName"`
	TotalQuantity  float64 `json:"totalQuantity"`
	Unit           string  `json:"unit"`
	Category       string  `json:"category"`
	IsCompleted    bool    `json:"isCompleted"`
}

type ShoppingListCategory struct {
	Name         string             `json:"name"`
	DisplayOrder int                `json:"displayOrder"`
	Items        []ShoppingListItem `json:"items"`
}