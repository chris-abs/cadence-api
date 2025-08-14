package entities

import (
	"time"

	"github.com/chrisabs/cadence/internal/models"
)

type Tag struct {
    ID          models.TagID     `json:"id"`
    Name        string           `json:"name"`
    Colour      string           `json:"colour"`
    Description string           `json:"description"`
    ProfileID   models.ProfileID `json:"profileId"`
    FamilyID    models.FamilyID  `json:"familyId"` 
    Items       []Item           `json:"items"`
    CreatedAt   time.Time        `json:"createdAt"`
    UpdatedAt   time.Time        `json:"updatedAt"`
}