package classification

import (
	"github.com/chrisabs/cadence/internal/media/classification/entities"
)

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

type ClassificationSearchRequest struct {
	Limit    *int            `json:"limit,omitempty"`
	Offset   *int            `json:"offset,omitempty"`
}

type ClassificationSearchResponse struct {
	Data   []entities.Classification `json:"data"`
	Total  int                       `json:"total"`
	Limit  int                       `json:"limit"`
	Offset int                       `json:"offset"`
	HasMore bool                     `json:"hasMore"`
}
