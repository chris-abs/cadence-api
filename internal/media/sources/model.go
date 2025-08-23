package sources

import (
	"github.com/chrisabs/cadence/internal/media/sources/entities"
)

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
	Data   []entities.Source `json:"data"`
	Total  int               `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
	HasMore bool             `json:"hasMore"`
}
