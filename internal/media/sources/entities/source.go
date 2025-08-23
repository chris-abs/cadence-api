package entities

import (
	"time"

	"github.com/chrisabs/cadence/internal/models"
)

type Source struct {
	ID        models.SourceID `json:"id"`
	Name      string          `json:"name"`
	Color     string          `json:"color"`
	Category  string          `json:"category"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	IsDeleted bool            `json:"isDeleted,omitempty"`
	DeletedAt *time.Time      `json:"deletedAt,omitempty"`
	DeletedBy *models.ProfileID `json:"deletedBy,omitempty"`
}