package models

type UserContext struct {
	UserID   int    `json:"userId"`
	FamilyID *int   `json:"familyId"`
	Role     *UserRole `json:"role"`
}
