package models

import "time"

type ProfileRole string

const (
	RoleParent ProfileRole = "PARENT"
	RoleChild  ProfileRole = "CHILD"
)

type Profile struct {
	ID           int                `json:"id"`
	FamilyID     int                `json:"familyId"`
	Name         string             `json:"name"`
	Role         ProfileRole        `json:"role"`
	Pin          string             `json:"-"` 
	HasPin       bool               `json:"hasPin"`
	ImageURL     string             `json:"imageUrl"`
	Colour       string             `json:"colour"`
	TimezoneName string             `json:"timezoneName"`
	IsOwner      bool               `json:"isOwner"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
}