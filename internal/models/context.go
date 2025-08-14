package models

type FamilyContext struct {
	FamilyID FamilyID `json:"familyId"`
}

type ProfileContext struct {
	FamilyID  FamilyID    `json:"familyId"`
	ProfileID ProfileID   `json:"profileId"`
	Role      ProfileRole `json:"role"`
	IsOwner   bool        `json:"isOwner"`
}