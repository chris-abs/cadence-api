package profile

import (
	"github.com/chrisabs/cadence/internal/models"
)

type CreateProfileRequest struct {
	Name     string             `json:"name"`
	Role     models.ProfileRole `json:"role"`
	Pin      string             `json:"pin,omitempty"`
	ImageURL string             `json:"imageUrl,omitempty"`
}

type UpdateProfileRequest struct {
    ID         int                `json:"id"`
    Name       string             `json:"name,omitempty"`
    Role       models.ProfileRole `json:"role,omitempty"`
    Pin        *string            `json:"pin,omitempty"` 
    CurrentPin string             `json:"currentPin,omitempty"`
	ImageURL   string             `json:"imageUrl,omitempty"`
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
	Token   string         `json:"token"`
	Profile models.Profile `json:"profile"`
}

type ProfilesList struct {
	Profiles []*models.Profile `json:"profiles"`
}