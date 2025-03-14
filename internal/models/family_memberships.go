package models

import "time"

type FamilyMembership struct {
	ID        int       `json:"id"`
	profileId    int       `json:"profileId"`
	FamilyID  int       `json:"familyId"`
	Role      UserRole  `json:"role"`
	IsOwner   bool      `json:"isOwner"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}