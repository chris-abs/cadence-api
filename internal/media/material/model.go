package media

import (
	"github.com/chrisabs/cadence/internal/media/entities"
	"github.com/chrisabs/cadence/internal/models"
)

type CreateMediaRequest struct {
	Name        string                `json:"name"`
	Type        entities.MediaType    `json:"type"`
	Genre       string                `json:"genre"`
	ReleaseYear int                   `json:"releaseYear"`
	Runtime     int                   `json:"runtime"`
	PosterURL   string                `json:"posterUrl"`
	SourceIDs   []models.SourceID     `json:"sourceIds"`
	WatchWith   entities.WatchWith    `json:"watchWith"`
	Status      entities.Status       `json:"status"`
	Priority    entities.Priority     `json:"priority"`
	Notes       string                `json:"notes"`
}

type UpdateMediaRequest struct {
	Name        string                `json:"name"`
	Type        entities.MediaType    `json:"type"`
	Genre       string                `json:"genre"`
	ReleaseYear int                   `json:"releaseYear"`
	Runtime     int                   `json:"runtime"`
	PosterURL   string                `json:"posterUrl"`
	SourceIDs   []models.SourceID     `json:"sourceIds"`
	WatchWith   entities.WatchWith    `json:"watchWith"`
	Status      entities.Status       `json:"status"`
	Priority    entities.Priority     `json:"priority"`
	Notes       string                `json:"notes"`
}

type UpdateMediaStatusRequest struct {
	Status entities.Status `json:"status"`
}

type MediaSearchRequest struct {
	Query     string                `json:"query"`
	ProfileID *models.ProfileID     `json:"profileId,omitempty"` 
	Type      entities.MediaType    `json:"type"`
	Genre     string                `json:"genre"`
	SourceID  models.SourceID       `json:"sourceId"`
	WatchWith entities.WatchWith    `json:"watchWith"`
	Status    entities.Status       `json:"status"`
	Priority  entities.Priority     `json:"priority"`
	Limit     int                   `json:"limit"`
	Offset    int                   `json:"offset"`
}

type MediaSearchResponse struct {
	Media   []entities.Media `json:"media"`
	Total   int              `json:"total"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
	HasMore bool             `json:"hasMore"`
}

type MediaEnumsResponse struct {
	Types      []entities.MediaType `json:"types"`
	Genres     []string             `json:"genres"`
	Sources    []entities.Source    `json:"sources"`
	WatchWith  []entities.WatchWith `json:"watchWith"`
	Statuses   []entities.Status    `json:"statuses"`
	Priorities []entities.Priority  `json:"priorities"`
}

type MediaStatusSummaryResponse struct {
	ToWatch         int `json:"toWatch"`
	InProgress      int `json:"inProgress"`
	Watching        int `json:"watching"`
	AwaitingRelease int `json:"awaitingRelease"`
	Watched         int `json:"watched"`
	Total           int `json:"total"`
}