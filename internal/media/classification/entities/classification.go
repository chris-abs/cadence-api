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
	CreatedBy   models.ProfileID        `json:"createdBy"`
	CreatedAt   time.Time               `json:"createdAt"`
	UpdatedAt   time.Time               `json:"updatedAt"`
	IsDeleted   bool                    `json:"isDeleted,omitempty"`
	DeletedAt   *time.Time              `json:"deletedAt,omitempty"`
	DeletedBy   *models.ProfileID       `json:"deletedBy,omitempty"`
}

type CreateClassificationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type UpdateClassificationRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Color       *string `json:"color,omitempty"`
}

type ClassificationSearchParams struct {
	FamilyID models.FamilyID `json:"familyId"`
	Limit    *int            `json:"limit,omitempty"`
	Offset   *int            `json:"offset,omitempty"`
}

type ClassificationSearchResponse struct {
	Classifications []Classification `json:"classifications"`
	Total          int              `json:"total"`
	Limit          int              `json:"limit"`
	Offset         int              `json:"offset"`
	HasMore        bool             `json:"hasMore"`
}
