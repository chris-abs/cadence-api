package family

import (
	"time"

	"github.com/chrisabs/cadence/internal/models"
)

type FamilyAccount struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` 
	FamilyName string   `json:"familyName"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type FamilySettings struct {
	FamilyID  int                 `json:"familyId"`
	Modules   []models.Module     `json:"modules"`
	Status    models.FamilyStatus `json:"status"`
	CreatedAt time.Time           `json:"createdAt"`
	UpdatedAt time.Time           `json:"updatedAt"`
}

type RegisterRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	FamilyName string `json:"familyName"`
	OwnerName  string `json:"ownerName"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type FamilyAuthResponse struct {
	Token    string        `json:"token"`
	Family   FamilyAccount `json:"family"`
	Profiles []models.Profile `json:"profiles"`
}

type UpdateFamilyRequest struct {
	FamilyName string `json:"familyName"`
}

type UpdateModuleRequest struct {
	ModuleID  models.ModuleID `json:"moduleId"`
	IsEnabled bool            `json:"isEnabled"`
}