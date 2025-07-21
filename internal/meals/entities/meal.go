package entities

import "time"

type Meal struct {
	ID                   int                `json:"id"`
	Name                 string             `json:"name"`
	Description          string             `json:"description"`
	ImageURL             string             `json:"imageUrl"`
	Cuisine              string             `json:"cuisine"`
	Subcategory          string             `json:"subcategory"`
	PrepTimeMinutes      int                `json:"prepTimeMinutes"`
	CookTimeMinutes      int                `json:"cookTimeMinutes"`
	Instructions         []string           `json:"instructions"`
	Macros               MealMacros         `json:"macros"`
	Options              []MealOption       `json:"options"`
	PrimaryIngredients   []MealIngredient   `json:"primaryIngredients"`
	SecondaryIngredients []MealIngredient   `json:"secondaryIngredients"`
	Sides                []MealIngredient   `json:"sides"`
	DietaryInfo          MealDietaryInfo    `json:"dietaryInfo"`
	CreatedAt            time.Time          `json:"createdAt"`
	UpdatedAt            time.Time          `json:"updatedAt"`
}

type MealMacros struct {
	Calories      int     `json:"calories"`
	Protein       float64 `json:"protein"`       
	Carbohydrates float64 `json:"carbohydrates"` 
	Fat           float64 `json:"fat"`           
	Fiber         float64 `json:"fiber"`         
	Sugar         float64 `json:"sugar"`         
	Sodium        float64 `json:"sodium"`        
}

type MealOption struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsDefault   bool   `json:"isDefault"`
}

type MealIngredient struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Quantity float64 `json:"quantity"` 
	Unit     string  `json:"unit"`     
	Category string  `json:"category"` 
}

type MealDietaryInfo struct {
	Allergens   []string `json:"allergens"`   
	DietaryTags []string `json:"dietaryTags"` 
}