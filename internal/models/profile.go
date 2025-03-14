package models

import "time"

type ProfileRole string

const (
	RoleParent ProfileRole = "PARENT"
	RoleChild  ProfileRole = "CHILD"
)

type Profile struct {
	ID        int                `json:"id"`
	FamilyID  int                `json:"familyId"`
	Name      string             `json:"name"`
	Role      ProfileRole        `json:"role"`
	Pin       string             `json:"-"` 
	ImageURL  string             `json:"imageUrl"`
	IsOwner   bool               `json:"isOwner"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}