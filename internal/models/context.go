package models

type ProfileContext struct {
	FamilyID  int      `json:"familyId"`
	ProfileID int      `json:"profileId"`
	Role      UserRole `json:"role"`
	IsOwner   bool     `json:"isOwner"`
}