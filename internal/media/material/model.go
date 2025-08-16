package material

import (
	"github.com/chrisabs/cadence/internal/media/material/entities"
	"github.com/chrisabs/cadence/internal/models"
)

type CreateMaterialRequest struct {
	Name             string                        `json:"name"`
	Type             entities.MaterialType         `json:"type"`
	Genre            string                        `json:"genre"`
	ReleaseYear      int                           `json:"releaseYear"`
	Runtime          int                           `json:"runtime"`
	PosterURL        string                        `json:"posterUrl"`
	SourceIDs        []models.SourceID             `json:"sourceIds"`
	ClassificationID *models.ClassificationID      `json:"classificationId,omitempty"`
	WatchWith        entities.WatchWith            `json:"watchWith"`
	Status           entities.Status               `json:"status"`
	Priority         entities.Priority             `json:"priority"`
	Notes            string                        `json:"notes"`
}

type UpdateMaterialRequest struct {
	Name             string                        `json:"name"`
	Type             entities.MaterialType         `json:"type"`
	Genre            string                        `json:"genre"`
	ReleaseYear      int                           `json:"releaseYear"`
	Runtime          int                           `json:"runtime"`
	PosterURL        string                        `json:"posterUrl"`
	SourceIDs        []models.SourceID             `json:"sourceIds"`
	ClassificationID *models.ClassificationID      `json:"classificationId,omitempty"`
	WatchWith        entities.WatchWith            `json:"watchWith"`
	Status           entities.Status               `json:"status"`
	Priority         entities.Priority             `json:"priority"`
	Notes            string                        `json:"notes"`
}

type UpdateMaterialStatusRequest struct {
	Status entities.Status `json:"status"`
}

type MaterialSearchRequest struct {
	Query            string                   `json:"query"`
	ProfileID        *models.ProfileID        `json:"profileId,omitempty"` 
	Type             entities.MaterialType    `json:"type"`
	Genre            string                   `json:"genre"`
	SourceID         models.SourceID          `json:"sourceId"`
	ClassificationID *models.ClassificationID `json:"classificationId,omitempty"`
	WatchWith        entities.WatchWith       `json:"watchWith"`
	Status           entities.Status          `json:"status"`
	Priority         entities.Priority        `json:"priority"`
	Limit            int                      `json:"limit"`
	Offset           int                      `json:"offset"`
}

type MaterialSearchResponse struct {
	Material   []entities.Material `json:"material"`
	Total   int                    `json:"total"`
	Limit   int                    `json:"limit"`
	Offset  int                    `json:"offset"`
	HasMore bool                   `json:"hasMore"`
}

type MaterialEnumsResponse struct {
	Types      []entities.MaterialType `json:"types"`
	Genres     []string                `json:"genres"`
	WatchWith  []entities.WatchWith    `json:"watchWith"`
	Statuses   []entities.Status       `json:"statuses"`
	Priorities []entities.Priority     `json:"priorities"`
}

type MaterialStatusSummaryResponse struct {
	ToWatch         int `json:"toWatch"`
	InProgress      int `json:"inProgress"`
	Watching        int `json:"watching"`
	AwaitingRelease int `json:"awaitingRelease"`
	Watched         int `json:"watched"`
	Total           int `json:"total"`
}