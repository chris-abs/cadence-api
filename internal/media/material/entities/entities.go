package entities

import (
	"time"

	"github.com/chrisabs/cadence/internal/models"
)

type Material struct {
	ID          models.MaterialID `json:"id"`
	Name        string            `json:"name"`
	Type        MaterialType      `json:"type"`
	Genre       string            `json:"genre"`
	ReleaseYear int               `json:"releaseYear"`
	Runtime     int               `json:"runtime"` 
	PosterURL   string            `json:"posterUrl"`
	SourceIDs   []models.SourceID `json:"sourceIds"`
	WatchWith   WatchWith         `json:"watchWith"`
	Status      Status            `json:"status"`
	Priority    Priority          `json:"priority"`
	Notes       string            `json:"notes"`
	ProfileID   models.ProfileID  `json:"profileId"`
	FamilyID    models.FamilyID   `json:"familyId"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type MaterialType string

const (
	MaterialTypeMovie MaterialType = "movie"
	MaterialTypeShow  MaterialType = "show"
)

type WatchWith string

const (
	WatchWithAlone   WatchWith = "alone"
	WatchWithPartner WatchWith = "partner"
	WatchWithFamily  WatchWith = "family"
)

type Status string

const (
	StatusToWatch        Status = "to_watch"
	StatusInProgress     Status = "in_progress"
	StatusWatching       Status = "watching"
	StatusAwaitingRelease Status = "awaiting_release"
	StatusWatched        Status = "watched"
)

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

var (
	ValidGenres = []string{
		"Action", "Adventure", "Animation", "Biography", "Comedy", "Crime",
		"Documentary", "Drama", "Family", "Fantasy", "History", "Horror",
		"Music", "Mystery", "Romance", "Sci-Fi", "Sport", "Thriller",
		"War", "Western",
	}
)

