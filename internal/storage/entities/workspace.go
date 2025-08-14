package entities

import (
	"time"

	"github.com/chrisabs/cadence/internal/models"
)

type Workspace struct {
    ID          models.WorkspaceID `json:"id"`
    Name        string             `json:"name"`
    Description string             `json:"description"`
    ProfileID   models.ProfileID   `json:"profileId"`
    FamilyID    models.FamilyID    `json:"familyId"`
    Containers  []Container        `json:"containers"`
    CreatedAt   time.Time          `json:"createdAt"`
    UpdatedAt   time.Time          `json:"updatedAt"`
}