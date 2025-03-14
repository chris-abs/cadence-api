package models

type ProfileContext struct {
	FamilyID  int         `json:"familyId"`
	ProfileID int         `json:"profileId"`
	Role      ProfileRole `json:"role"`
	IsOwner   bool        `json:"isOwner"`
}