package entities

import "time"

type MealPlan struct {
	ID        int           `json:"id"`
	FamilyID  int           `json:"familyId"`
	StartDate time.Time     `json:"startDate"`
	EndDate   time.Time     `json:"endDate"`
	Days      []MealPlanDay `json:"days"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
	CreatedBy int           `json:"createdBy"`
}

type MealPlanDay struct {
	ID         int       `json:"id"`
	MealPlanID int       `json:"mealPlanId"`
	Date       time.Time `json:"date"`
	MealID     *int      `json:"mealId,omitempty"`
	Meal       *Meal     `json:"meal,omitempty"`
	AdultCount int       `json:"adultCount"`
	ChildCount int       `json:"childCount"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}