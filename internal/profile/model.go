package profile

import (
	"time"

	"github.com/chrisabs/cadence/internal/models"
)

type Profile struct {
	ID        int               `json:"id"`
	FamilyID  int               `json:"familyId"`
	Name      string            `json:"name"`
	Role      models.ProfileRole `json:"role"`
	Pin       string            `json:"-"` 
	ImageURL  string            `json:"imageUrl"`
	IsOwner   bool              `json:"isOwner"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdatedAt time.Time         `json:"updatedAt"`
}

type CreateProfileRequest struct {
	Name     string            `json:"name"`
	Role     models.ProfileRole `json:"role"`
	Pin      string            `json:"pin,omitempty"`
	ImageURL string            `json:"imageUrl,omitempty"`
}

type UpdateProfileRequest struct {
	Name     string            `json:"name"`
	Role     models.ProfileRole `json:"role,omitempty"`
	Pin      string            `json:"pin,omitempty"`
	ImageURL string            `json:"imageUrl,omitempty"`
}

type SelectProfileRequest struct {
	ProfileID int    `json:"profileId"`
	Pin       string `json:"pin,omitempty"`
}

type VerifyPinRequest struct {
	ProfileID int    `json:"profileId"`
	Pin       string `json:"pin,omitempty"`
}

type ProfileResponse struct {
	Token   string  `json:"token"`
	Profile Profile `json:"profile"`
}

type ProfilesList struct {
	Profiles []Profile `json:"profiles"`
}