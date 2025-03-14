package models

type UserContext struct {
	profileId   int    `json:"profileId"`
	FamilyID *int   `json:"familyId"`
	Role     *UserRole `json:"role"`
}
