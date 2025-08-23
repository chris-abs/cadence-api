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

type CreateSourceRequest struct {
	Name     string `json:"name"`
	Color    string `json:"color"`
	Category string `json:"category"`
}

type UpdateSourceRequest struct {
	Name     *string `json:"name,omitempty"`
	Color    *string `json:"color,omitempty"`
	Category *string `json:"category,omitempty"`
}

type SourceSearchParams struct {
	Category *string `json:"category,omitempty"`
	Limit    *int    `json:"limit,omitempty"`
	Offset   *int    `json:"offset,omitempty"`
}

type SourceSearchResponse struct {
	Data   []Source `json:"data"`
	Total  int      `json:"total"`
	Limit  int      `json:"limit"`
	Offset int      `json:"offset"`
	HasMore bool    `json:"hasMore"`
}