package models

import "time"

type FamilyMembership struct {
	ID        int       `json:"id"`
	UserID    int       `json:"userId"`
	FamilyID  int       `json:"familyId"`
	Role      UserRole  `json:"role"`
	IsOwner   bool      `json:"isOwner"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}