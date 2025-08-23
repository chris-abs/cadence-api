package entities

import (
	"time"

	"github.com/chrisabs/cadence/internal/models"
)

type Classification struct {
	ID          models.ClassificationID `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Color       string                  `json:"color"`
	ImageURL    string                  `json:"imageUrl"`
	FamilyID    models.FamilyID         `json:"familyId"`
	ProfileID   models.ProfileID        `json:"profileId"`
	CreatedBy   models.ProfileID        `json:"createdBy"`
	CreatedAt   time.Time               `json:"createdAt"`
	UpdatedAt   time.Time               `json:"updatedAt"`
	IsDeleted   bool                    `json:"isDeleted,omitempty"`
	DeletedAt   *time.Time              `json:"deletedAt,omitempty"`
	DeletedBy   *models.ProfileID       `json:"deletedBy,omitempty"`
}
