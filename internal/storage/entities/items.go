package entities

import (
	"time"

	"github.com/chrisabs/cadence/internal/models"
)

type ItemImage struct {
    ID           models.ItemImageID `json:"id"`
    URL          string             `json:"url"`
    DisplayOrder int                `json:"displayOrder"`
    CreatedAt    time.Time          `json:"createdAt"`
    UpdatedAt    time.Time          `json:"updatedAt"`
}

type Item struct {
    ID          models.ItemID        `json:"id"`
    Name        string               `json:"name"`
    Description string               `json:"description"`
    Images      []ItemImage          `json:"images"`
    Quantity    int                  `json:"quantity"`
    ContainerID *models.ContainerID  `json:"containerId,omitempty"`
    Container   *Container           `json:"container,omitempty"`
    ProfileID   models.ProfileID     `json:"profileId"`
    FamilyID    models.FamilyID      `json:"familyId"`    
    Tags        []Tag                `json:"tags"`
    CreatedAt   time.Time            `json:"createdAt"`
    UpdatedAt   time.Time            `json:"updatedAt"`
}