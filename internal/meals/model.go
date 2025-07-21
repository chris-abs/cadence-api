package meals

import "time"

type CreateMealPlanRequest struct {
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

type UpdateMealPlanDayRequest struct {
	MealID     *int `json:"mealId"`
	AdultCount int  `json:"adultCount"`
	ChildCount int  `json:"childCount"`
}

type MealSearchRequest struct {
	Query            string   `json:"query"`
	Cuisine          string   `json:"cuisine"`
	Subcategory      string   `json:"subcategory"`
	MaxPrepTime      int      `json:"maxPrepTime"`
	MaxCookTime      int      `json:"maxCookTime"`
	ExcludeAllergens []string `json:"excludeAllergens"`
	DietaryTags      []string `json:"dietaryTags"`
	Limit            int      `json:"limit"`
	Offset           int      `json:"offset"`
}

type UpdateShoppingListItemRequest struct {
	IsCompleted bool `json:"isCompleted"`
}

type GenerateShoppingListRequest struct {
	MealPlanID int `json:"mealPlanId"`
}

type MealSearchResponse struct {
	Meals   []MealSummary `json:"meals"`
	Total   int           `json:"total"`
	Limit   int           `json:"limit"`
	Offset  int           `json:"offset"`
	HasMore bool          `json:"hasMore"`
}

type MealSummary struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	ImageURL        string   `json:"imageUrl"`
	Cuisine         string   `json:"cuisine"`
	Subcategory     string   `json:"subcategory"`
	PrepTimeMinutes int      `json:"prepTimeMinutes"`
	CookTimeMinutes int      `json:"cookTimeMinutes"`
	Allergens       []string `json:"allergens"`
	DietaryTags     []string `json:"dietaryTags"`
}

type CuisineListResponse struct {
	Cuisines []CuisineInfo `json:"cuisines"`
}

type CuisineInfo struct {
	Name          string   `json:"name"`
	Subcategories []string `json:"subcategories"`
	MealCount     int      `json:"mealCount"`
}

type ShoppingListGenerationResponse struct {
	ShoppingListID int       `json:"shoppingListId"`
	GeneratedAt    time.Time `json:"generatedAt"`
	TotalItems     int       `json:"totalItems"`
	Categories     []string  `json:"categories"`
}