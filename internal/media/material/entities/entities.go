package entities

import (
	"time"

	"github.com/chrisabs/cadence/internal/models"
)

type Media struct {
	ID          models.MediaID    `json:"id"`
	Name        string            `json:"name"`
	Type        MediaType         `json:"type"`
	Genre       string            `json:"genre"`
	ReleaseYear int               `json:"releaseYear"`
	Runtime     int               `json:"runtime"` 
	PosterURL   string            `json:"posterUrl"`
	SourceIDs   []models.SourceID `json:"sourceIds"`
	Sources     []Source          `json:"sources,omitempty"` 
	WatchWith   WatchWith         `json:"watchWith"`
	Status      Status            `json:"status"`
	Priority    Priority          `json:"priority"`
	Notes       string            `json:"notes"`
	ProfileID   models.ProfileID  `json:"profileId"`
	FamilyID    models.FamilyID   `json:"familyId"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type MediaType string

const (
	MediaTypeMovie MediaType = "movie"
	MediaTypeShow  MediaType = "show"
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

type Source struct {
    ID       models.SourceID `json:"id"`
    Name     string          `json:"name"`
    Color    string          `json:"color"`   
    Category string          `json:"category"`
}

var (
	ValidGenres = []string{
		"Action", "Adventure", "Animation", "Biography", "Comedy", "Crime",
		"Documentary", "Drama", "Family", "Fantasy", "History", "Horror",
		"Music", "Mystery", "Romance", "Sci-Fi", "Sport", "Thriller",
		"War", "Western",
	}
)